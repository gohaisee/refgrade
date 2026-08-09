package check

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"go/ast"
	"go/token"

	"github.com/gohaisee/refgrade/internal/astutil"
	"github.com/gohaisee/refgrade/internal/subprocess"
)

const (
	deadcodeTimeout    = 2 * time.Minute
	staticcheckTimeout = 3 * time.Minute
	dead07MinLines     = 6
	dead07MinCodeLike  = 3
)

type deadcodeHit struct {
	file string
	line int
	name string
}

// unreachable func via deadcode tool
type Dead01 struct{ Base }

func NewDead01() *Dead01 {
	return &Dead01{Base: Base{meta: Meta{ID: "dead-01", Domain: "dead-code", DefaultSeverity: SeverityWarn}}}
}

func (c *Dead01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	hits, err := runDeadcode(ctx, mod)
	if err != nil {
		return nil, err
	}
	var findings []Finding
	for _, h := range hits {
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), h.file, h.line))
	}
	return findings, nil
}

// Dead02 — exported unreachable symbol in internal/
type Dead02 struct{ Base }

func NewDead02() *Dead02 {
	return &Dead02{Base: Base{meta: Meta{ID: "dead-02", Domain: "dead-code", DefaultSeverity: SeverityInfo}}}
}

func (c *Dead02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	hits, err := runDeadcode(ctx, mod)
	if err != nil {
		return nil, err
	}
	var findings []Finding
	for _, h := range hits {
		if !pathUnder(h.file, "internal") {
			continue
		}
		if h.name == "" || !isExportedName(h.name) {
			continue
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityInfo), h.file, h.line))
	}
	return findings, nil
}

// Dead03 — staticcheck U1000 unused func/type/const
type Dead03 struct{ Base }

func NewDead03() *Dead03 {
	return &Dead03{Base: Base{meta: Meta{ID: "dead-03", Domain: "dead-code", DefaultSeverity: SeverityWarn}}}
}

func (c *Dead03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	res, err := subprocess.RunInDir(ctx, mod.Root(), "staticcheck", []string{"-f", "json", "-checks", "U1000", "./..."}, staticcheckTimeout)
	if err != nil {
		if errors.Is(err, subprocess.ErrNotFound) || strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		return nil, err
	}
	var findings []Finding
	for _, line := range splitLines(string(res.Stdout)) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry struct {
			Location struct {
				File string `json:"file"`
				Line int    `json:"line"`
			} `json:"location"`
			Code     string `json:"code"`
			Category string `json:"category"`
		}
		if json.Unmarshal([]byte(line), &entry) != nil {
			continue
		}
		checkID := entry.Code
		if checkID == "" {
			checkID = entry.Category
		}
		if checkID != "U1000" {
			continue
		}
		rel := entry.Location.File
		if r, err := mod.RelPath(entry.Location.File); err == nil {
			rel = r
		}
		if mod.Excluded(rel) {
			continue
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), rel, entry.Location.Line))
	}
	return findings, nil
}

// Dead04 — .go file on disk but not in go list package build
type Dead04 struct{ Base }

func NewDead04() *Dead04 {
	return &Dead04{Base: Base{meta: Meta{ID: "dead-04", Domain: "dead-code", DefaultSeverity: SeverityWarn}}}
}

func (c *Dead04) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	built := builtGoFiles(mod)
	var findings []Finding
	err := walkModuleGoFiles(mod.Root(), func(abs, rel string) error {
		if mod.Excluded(rel) {
			return nil
		}
		if built[abs] {
			return nil
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), rel, 1))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return findings, nil
}

// Dead05 — package with only package clause, no declarations
type Dead05 struct{ Base }

func NewDead05() *Dead05 {
	return &Dead05{Base: Base{meta: Meta{ID: "dead-05", Domain: "dead-code", DefaultSeverity: SeverityWarn}}}
}

func (c *Dead05) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mod.ASTPool(astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	byPkg := map[string][]*astutil.File{}
	for _, f := range pool.Files {
		dir := filepath.ToSlash(filepath.Dir(f.RelPath))
		byPkg[dir] = append(byPkg[dir], f)
	}
	var findings []Finding
	for _, files := range byPkg {
		if len(files) == 0 || !packageIsEmpty(files) {
			continue
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), files[0].RelPath, 1))
	}
	return findings, nil
}

// go mod tidy would change go.mod
type Dead06 struct{ Base }

func NewDead06() *Dead06 {
	return &Dead06{Base: Base{meta: Meta{ID: "dead-06", Domain: "dead-code", DefaultSeverity: SeverityWarn}}}
}

func (c *Dead06) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	res, err := subprocess.RunInDir(ctx, mod.Root(), "go", []string{"mod", "tidy", "-diff"}, deadcodeTimeout)
	if err != nil {
		return nil, err
	}
	out := strings.TrimSpace(string(res.Stdout) + string(res.Stderr))
	if out == "" {
		return nil, nil
	}
	return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), "go.mod", 1)}, nil
}

// Dead07 — large commented-out code blocks
type Dead07 struct{ Base }

func NewDead07() *Dead07 {
	return &Dead07{Base: Base{meta: Meta{ID: "dead-07", Domain: "dead-code", DefaultSeverity: SeverityWarn}}}
}

func (c *Dead07) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	var findings []Finding
	for _, gf := range mod.GoSourceFiles() {
		if astutil.IsTestFile(gf.RelPath) {
			continue
		}
		data, err := os.ReadFile(gf.Path)
		if err != nil {
			return nil, err
		}
		line := firstLargeCommentBlockLine(string(data))
		if line == 0 {
			continue
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), gf.RelPath, line))
	}
	return findings, nil
}

// Dead08 — exported test helper referenced from one file only
type Dead08 struct{ Base }

func NewDead08() *Dead08 {
	return &Dead08{Base: Base{meta: Meta{ID: "dead-08", Domain: "dead-code", DefaultSeverity: SeverityInfo}}}
}

func (c *Dead08) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mod.ASTPool(astutil.DefaultFilter())
	if err != nil {
		return nil, err
	}
	byPkg := map[string][]*astutil.File{}
	for _, f := range pool.Files {
		dir := filepath.ToSlash(filepath.Dir(f.RelPath))
		byPkg[dir] = append(byPkg[dir], f)
	}
	var findings []Finding
	for _, files := range byPkg {
		helpers := exportedTestHelpers(pool, files)
		for name, decl := range helpers {
			if helperUsedInFileCount(files, name) != 1 {
				continue
			}
			findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityInfo), decl.file, decl.line))
		}
	}
	return findings, nil
}

func runDeadcode(ctx context.Context, mod ModuleView) ([]deadcodeHit, error) {
	res, err := subprocess.RunInDir(ctx, mod.Root(), "deadcode", []string{"-test", "./..."}, deadcodeTimeout)
	if err != nil {
		if errors.Is(err, subprocess.ErrNotFound) || strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		return nil, err
	}
	if res.ExitCode != 0 && len(res.Stdout) == 0 {
		return nil, nil
	}
	var hits []deadcodeHit
	for _, line := range splitLines(string(res.Stdout)) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		file, lineNum, name := parseDeadcodeLine(line)
		if file == "" {
			continue
		}
		rel := file
		if r, err := mod.RelPath(file); err == nil {
			rel = r
		}
		if mod.Excluded(rel) {
			continue
		}
		hits = append(hits, deadcodeHit{file: rel, line: lineNum, name: name})
	}
	return hits, nil
}

func parseDeadcodeLine(line string) (string, int, string) {
	line = strings.TrimPrefix(strings.TrimSpace(line), "deadcode: ")
	marker := ": unreachable func"
	idx := strings.Index(line, marker)
	if idx == -1 {
		marker = ": unreachable function"
		idx = strings.Index(line, marker)
	}
	if idx == -1 {
		return "", 0, ""
	}
	head := line[:idx]
	name := strings.TrimSpace(strings.TrimPrefix(line[idx+len(marker):], ":"))
	parts := strings.Split(head, ":")
	if len(parts) < 2 {
		return "", 0, ""
	}
	var lineNum int
	_, _ = fmtSscanf(parts[len(parts)-2], &lineNum)
	file := strings.Join(parts[:len(parts)-2], ":")
	return file, lineNum, name
}

func fmtSscanf(s string, n *int) (int, error) {
	var x int
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		x = x*10 + int(c-'0')
	}
	*n = x
	return 1, nil
}

func builtGoFiles(mod ModuleView) map[string]bool {
	built := make(map[string]bool)
	for _, gf := range mod.GoSourceFiles() {
		built[gf.Path] = true
	}
	return built
}

func walkModuleGoFiles(root string, fn func(abs, rel string) error) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := d.Name()
			if base == "vendor" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		return fn(path, rel)
	})
}

func packageIsEmpty(files []*astutil.File) bool {
	for _, f := range files {
		if fileHasDeclarations(f.AST) {
			return false
		}
	}
	return true
}

func fileHasDeclarations(file *ast.File) bool {
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.IMPORT {
				continue
			}
			return true
		case *ast.FuncDecl:
			return true
		}
	}
	return false
}

func firstLargeCommentBlockLine(src string) int {
	lines := strings.Split(src, "\n")
	var runStart int
	var runLen int
	var codeLike int
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "//") {
			if runLen == 0 {
				runStart = i + 1
			}
			runLen++
			if commentLooksLikeCode(strings.TrimSpace(strings.TrimPrefix(trim, "//"))) {
				codeLike++
			}
			if runLen >= dead07MinLines && codeLike >= dead07MinCodeLike {
				return runStart
			}
			continue
		}
		runLen = 0
		codeLike = 0
	}
	return 0
}

func commentLooksLikeCode(body string) bool {
	if body == "" {
		return false
	}
	markers := []string{"{", "}", "func ", "if ", "for ", "return ", ":=", " = ", ");", "struct ", "switch ", "case "}
	for _, m := range markers {
		if strings.Contains(body, m) {
			return true
		}
	}
	return false
}

type helperDecl struct {
	file string
	line int
}

func exportedTestHelpers(pool *astutil.Pool, files []*astutil.File) map[string]helperDecl {
	out := make(map[string]helperDecl)
	for _, f := range files {
		if !astutil.IsTestFile(f.RelPath) {
			continue
		}
		for _, decl := range f.AST.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Recv != nil {
				continue
			}
			name := fn.Name.Name
			if !isExportedName(name) || isTestFrameworkName(name) {
				continue
			}
			out[name] = helperDecl{file: f.RelPath, line: pool.Line(fn.Name)}
		}
	}
	return out
}

func isTestFrameworkName(name string) bool {
	switch {
	case strings.HasPrefix(name, "Test"),
		strings.HasPrefix(name, "Benchmark"),
		strings.HasPrefix(name, "Example"),
		strings.HasPrefix(name, "Fuzz"):
		return true
	default:
		return false
	}
}

func isExportedName(name string) bool {
	if name == "" {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}

func helperUsedInFileCount(files []*astutil.File, name string) int {
	n := 0
	for _, f := range files {
		if fileUsesIdent(f, name) {
			n++
		}
	}
	return n
}

func fileUsesIdent(f *astutil.File, name string) bool {
	found := false
	ast.Inspect(f.AST, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || id.Name != name || !isUseIdent(id) {
			return true
		}
		found = true
		return false
	})
	return found
}

func isUseIdent(id *ast.Ident) bool {
	if id.Obj == nil {
		return true
	}
	switch decl := id.Obj.Decl.(type) {
	case *ast.FuncDecl:
		return decl.Name != id
	case *ast.TypeSpec:
		return decl.Name != id
	case *ast.ValueSpec:
		for _, n := range decl.Names {
			if n == id {
				return false
			}
		}
		return true
	default:
		return true
	}
}

// govulncheck via --with-security (sec-16)
func runGovulncheck(ctx context.Context, mod ModuleView) ([]Finding, error) {
	res, err := subprocess.RunInDir(ctx, mod.Root(), "govulncheck", []string{"-json", "./..."}, 3*time.Minute)
	if err != nil {
		if errors.Is(err, subprocess.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var findings []Finding
	for _, line := range splitLines(string(res.Stdout)) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry struct {
			OSV string `json:"osv"`
		}
		if json.Unmarshal([]byte(line), &entry) != nil || entry.OSV == "" {
			continue
		}
		findings = append(findings, finding("sec-16", SeverityWarn, "go.mod", 1))
		break
	}
	return findings, nil
}
