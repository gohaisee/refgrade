package astutil

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// one parsed Go source file with import map
type File struct {
	AbsPath string
	RelPath string
	Src     []byte
	AST     *ast.File
	Imports map[string]string // local name → import path
}

// shared parse cache for one module scan
type Pool struct {
	Fset  *token.FileSet
	Files []*File
	byAbs map[string]*File
}

// file filter for pool construction
type Filter struct {
	SkipTestFiles      bool
	SkipIntegrationE2E bool
	Match              func(relPath string) bool
}

// DefaultFilter skips nothing except caller exclusions
func DefaultFilter() Filter {
	return Filter{}
}

// builds pool: one read + one parse per file
func NewPool(files []SourceFile, exclude func(rel string) bool, filter Filter) (*Pool, error) {
	fset := token.NewFileSet()
	pool := &Pool{Fset: fset, byAbs: make(map[string]*File)}

	for _, sf := range files {
		rel := filepath.ToSlash(sf.RelPath)
		if exclude != nil && exclude(rel) {
			continue
		}
		if filter.Match != nil && !filter.Match(rel) {
			continue
		}
		if filter.SkipTestFiles && IsTestFile(rel) {
			continue
		}

		src, err := os.ReadFile(sf.AbsPath)
		if err != nil {
			return nil, err
		}
		if filter.SkipIntegrationE2E && HasIntegrationOrE2EBuildTag(string(src)) {
			continue
		}

		astFile, err := parser.ParseFile(fset, sf.AbsPath, src, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		f := &File{
			AbsPath: sf.AbsPath,
			RelPath: rel,
			Src:     src,
			AST:     astFile,
			Imports: importMap(astFile),
		}
		pool.Files = append(pool.Files, f)
		pool.byAbs[sf.AbsPath] = f
	}
	return pool, nil
}

// abs path input for pool build
type SourceFile struct {
	AbsPath string
	RelPath string
}

func importMap(file *ast.File) map[string]string {
	m := make(map[string]string)
	for _, spec := range file.Imports {
		path := strings.Trim(spec.Path.Value, `"`)
		local := path
		if spec.Name != nil {
			if spec.Name.Name == "_" || spec.Name.Name == "." {
				continue
			}
			local = spec.Name.Name
		} else {
			local = pathBase(path)
		}
		m[local] = path
	}
	return m
}

func pathBase(importPath string) string {
	if i := strings.LastIndex(importPath, "/"); i >= 0 {
		return importPath[i+1:]
	}
	return importPath
}

// local ident from selector resolves to import path
func (f *File) ImportPathOf(ident string) string {
	return f.Imports[ident]
}

// selector package import path and field/method name
func (f *File) Selector(sel *ast.SelectorExpr) (importPath, name string, ok bool) {
	switch x := sel.X.(type) {
	case *ast.Ident:
		imp := f.Imports[x.Name]
		if imp == "" {
			return "", sel.Sel.Name, false
		}
		return imp, sel.Sel.Name, true
	default:
		return "", sel.Sel.Name, false
	}
}

// line number for node position
func (p *Pool) Line(node ast.Node) int {
	return p.Fset.Position(node.Pos()).Line
}

func IsTestFile(rel string) bool {
	return strings.HasSuffix(filepath.Base(rel), "_test.go")
}

func IsCmdFile(rel string) bool {
	return strings.HasPrefix(filepath.ToSlash(rel), "cmd/")
}

func IsInternalFile(rel string) bool {
	rel = filepath.ToSlash(rel)
	return strings.HasPrefix(rel, "internal/") && !IsTestFile(rel)
}

// go:build integration|e2e in file header
func HasIntegrationOrE2EBuildTag(src string) bool {
	lines := strings.Split(src, "\n")
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		if !strings.HasPrefix(trim, "//") {
			break
		}
		body := strings.TrimSpace(strings.TrimPrefix(trim, "//"))
		var expr string
		switch {
		case strings.HasPrefix(body, "go:build "):
			expr = strings.TrimPrefix(body, "go:build ")
		case strings.HasPrefix(body, "+build "):
			expr = strings.TrimPrefix(body, "+build ")
		default:
			continue
		}
		if buildTagHasIntegrationOrE2E(expr) {
			return true
		}
	}
	return false
}

func buildTagHasIntegrationOrE2E(expr string) bool {
	for _, part := range strings.FieldsFunc(strings.TrimSpace(expr), func(r rune) bool {
		return r == ' ' || r == '\t' || r == '|' || r == '&' || r == '!' || r == '(' || r == ')'
	}) {
		if part == "integration" || part == "e2e" {
			return true
		}
	}
	return false
}

// walks all files with ast.Inspect
func (p *Pool) Inspect(fn func(f *File, n ast.Node) bool) {
	for _, file := range p.Files {
		ast.Inspect(file.AST, func(n ast.Node) bool {
			return fn(file, n)
		})
	}
}

// source contains any pattern (case-sensitive)
func (f *File) ContainsAny(patterns ...string) bool {
	s := string(f.Src)
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
