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

const cfg01ID = "cfg-01"

// Cfg01 — os.Getenv вне cmd/, тестов и integration/e2e сборок
type Cfg01 struct{}

func NewCfg01() *Cfg01 {
	return &Cfg01{}
}

func (c *Cfg01) ID() string {
	return cfg01ID
}

func (c *Cfg01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	var findings []Finding
	fset := token.NewFileSet()

	for _, gf := range mod.GoSourceFiles() {
		if skipCfg01File(gf.RelPath) {
			continue
		}

		src, err := os.ReadFile(gf.Path)
		if err != nil {
			return nil, err
		}
		if hasIntegrationE2EBuildTag(string(src)) {
			continue
		}

		file, err := parser.ParseFile(fset, gf.Path, src, 0)
		if err != nil {
			return nil, err
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if !isOSGetenv(call.Fun) {
				return true
			}
			pos := fset.Position(call.Pos())
			findings = append(findings, Finding{
				ID:       cfg01ID,
				Severity: SeverityFail,
				File:     gf.RelPath,
				Line:     pos.Line,
			})
			return true
		})
	}
	return findings, nil
}

func skipCfg01File(rel string) bool {
	rel = filepath.ToSlash(rel)
	base := filepath.Base(rel)
	if strings.HasSuffix(base, "_test.go") {
		return true
	}
	if strings.HasPrefix(rel, "cmd/") {
		return true
	}
	return false
}

func hasIntegrationE2EBuildTag(src string) bool {
	// go:build в шапке файла — иначе тег не учитываем
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
		if strings.HasPrefix(body, "go:build ") {
			expr := strings.TrimPrefix(body, "go:build ")
			if buildTagMatchesIntegrationOrE2E(expr) {
				return true
			}
		}
		if strings.HasPrefix(body, "+build ") {
			expr := strings.TrimPrefix(body, "+build ")
			if buildTagMatchesIntegrationOrE2E(expr) {
				return true
			}
		}
	}
	return false
}

func buildTagMatchesIntegrationOrE2E(expr string) bool {
	expr = strings.TrimSpace(expr)
	for _, part := range strings.FieldsFunc(expr, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '|' || r == '&' || r == '!' || r == '(' || r == ')'
	}) {
		if part == "integration" || part == "e2e" {
			return true
		}
	}
	return false
}

func isOSGetenv(fun ast.Expr) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "os" && sel.Sel.Name == "Getenv"
}
