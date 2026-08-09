package check

import (
	"context"

	"github.com/gohaisee/refgrade/internal/astutil"
	"go/ast"
)

// Cfg01 — os.Getenv вне cmd/, тестов и integration/e2e сборок
type Cfg01 struct{ Base }

func NewCfg01() *Cfg01 {
	return &Cfg01{Base: Base{meta: Meta{ID: "cfg-01", Domain: "config", DefaultSeverity: SeverityFail}}}
}

func (c *Cfg01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipIntegrationE2E: true})
	if err != nil {
		return nil, err
	}
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if astutil.IsCmdFile(f.RelPath) || astutil.IsTestFile(f.RelPath) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if !isGetenvCall(f, call.Fun) {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), f.RelPath, pool.Line(n)))
		return true
	})
	return findings, nil
}

func isGetenvCall(f *astutil.File, fun ast.Expr) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Getenv" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return f.ImportPathOf(ident.Name) == "os"
}
