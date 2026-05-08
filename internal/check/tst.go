package check

import (
	"context"
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
)

// Tst01 — foo.go без foo_test.go
type Tst01 struct{ Base }

func NewTst01() *Tst01 {
	return &Tst01{Base: Base{meta: Meta{ID: "tst-01", Domain: "tests", DefaultSeverity: SeverityWarn}}}
}

func (c *Tst01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	testFiles := map[string]struct{}{}
	logicFiles := map[string]string{}
	for _, gf := range mod.GoSourceFiles() {
		rel := filepath.ToSlash(gf.RelPath)
		dir := filepath.Dir(rel)
		base := filepath.Base(rel)
		if strings.HasSuffix(base, "_test.go") {
			testFiles[dir+"/"+strings.TrimSuffix(base, "_test.go")] = struct{}{}
			continue
		}
		if skipTst01File(rel) {
			continue
		}
		stem := strings.TrimSuffix(base, ".go")
		logicFiles[dir+"/"+stem] = rel
	}
	var findings []Finding
	for key, rel := range logicFiles {
		if _, ok := testFiles[key]; !ok {
			findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), rel, 1))
		}
	}
	return findings, nil
}

func skipTst01File(rel string) bool {
	base := filepath.Base(rel)
	if base == "doc.go" {
		return true
	}
	if strings.HasSuffix(base, "_gen.go") || base == "generated.go" {
		return true
	}
	if strings.Contains(rel, "/mocks/") {
		return true
	}
	if strings.HasPrefix(rel, "cmd/") && base == "main.go" {
		return true
	}
	return false
}

// Tst02 — один all_test.go на весь пакет
type Tst02 struct{ Base }

func NewTst02() *Tst02 {
	return &Tst02{Base: Base{meta: Meta{ID: "tst-02", Domain: "tests", DefaultSeverity: SeverityWarn}}}
}

func (c *Tst02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	byDir := map[string][]string{}
	for _, gf := range mod.GoSourceFiles() {
		rel := filepath.ToSlash(gf.RelPath)
		if !strings.HasSuffix(rel, "_test.go") {
			continue
		}
		dir := filepath.Dir(rel)
		byDir[dir] = append(byDir[dir], rel)
	}
	var findings []Finding
	for dir, tests := range byDir {
		if len(tests) != 1 {
			continue
		}
		if filepath.Base(tests[0]) == "all_test.go" {
			logicCount := 0
			for _, gf := range mod.GoSourceFiles() {
				rel := filepath.ToSlash(gf.RelPath)
				if filepath.Dir(rel) == dir && !strings.HasSuffix(rel, "_test.go") && !skipTst01File(rel) {
					logicCount++
				}
			}
			if logicCount > 2 {
				findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), tests[0], 1))
			}
		}
	}
	return findings, nil
}

// Tst03 — t.Skip без build tag
type Tst03 struct{ Base }

func NewTst03() *Tst03 {
	return &Tst03{Base: Base{meta: Meta{ID: "tst-03", Domain: "tests", DefaultSeverity: SeverityWarn}}}
}

func (c *Tst03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{})
	if err != nil {
		return nil, err
	}
	var findings []Finding
	for _, f := range pool.Files {
		if !strings.HasSuffix(f.RelPath, "_test.go") {
			continue
		}
		if astutil.HasIntegrationOrE2EBuildTag(string(f.Src)) {
			continue
		}
		pool.Inspect(func(ff *astutil.File, n ast.Node) bool {
			if ff.RelPath != f.RelPath {
				return true
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Skip" {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if ok && ident.Name == "t" {
				findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(n)))
			}
			return true
		})
	}
	return findings, nil
}
