package check

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const err03ID = "err-03"

// Err03 — http.DefaultClient и пакетные http.Get/http.Post без своего клиента
type Err03 struct{}

func NewErr03() *Err03 {
	return &Err03{}
}

func (c *Err03) ID() string {
	return err03ID
}

func (c *Err03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	var findings []Finding
	fset := token.NewFileSet()

	for _, gf := range mod.GoSourceFiles() {
		if skipErr03File(gf.RelPath) {
			continue
		}

		src, err := os.ReadFile(gf.Path)
		if err != nil {
			return nil, err
		}

		file, err := parser.ParseFile(fset, gf.Path, src, 0)
		if err != nil {
			return nil, err
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.CallExpr:
				if isHTTPPackageCall(node.Fun, "Get", "Post") {
					pos := fset.Position(node.Pos())
					findings = append(findings, Finding{
						ID:       err03ID,
						Severity: SeverityFail,
						File:     gf.RelPath,
						Line:     pos.Line,
					})
				}
			case *ast.SelectorExpr:
				if isHTTPPackageSel(node, "DefaultClient") {
					pos := fset.Position(node.Pos())
					findings = append(findings, Finding{
						ID:       err03ID,
						Severity: SeverityFail,
						File:     gf.RelPath,
						Line:     pos.Line,
					})
				}
			}
			return true
		})
	}
	return findings, nil
}

func skipErr03File(rel string) bool {
	return strings.HasSuffix(filepath.Base(rel), "_test.go")
}

func isHTTPPackageSel(sel *ast.SelectorExpr, names ...string) bool {
	ident, ok := sel.X.(*ast.Ident)
	if !ok || ident.Name != "http" {
		return false
	}
	for _, name := range names {
		if sel.Sel.Name == name {
			return true
		}
	}
	return false
}

func isHTTPPackageCall(fun ast.Expr, names ...string) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok || ident.Name != "http" {
		return false
	}
	for _, name := range names {
		if sel.Sel.Name == name {
			return true
		}
	}
	return false
}
