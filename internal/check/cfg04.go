package check

import (
	"context"
	"go/ast"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
)

// Cfg04 — пустой jwt secret в литерале
type Cfg04 struct{ Base }

func NewCfg04() *Cfg04 {
	return &Cfg04{Base: Base{meta: Meta{ID: "cfg-04", Domain: "config", DefaultSeverity: SeverityFail}}}
}

func (c *Cfg04) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.DefaultFilter())
	if err != nil {
		return nil, err
	}
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			findings = append(findings, c.checkAssign(mod, f, pool, node)...)
		case *ast.CompositeLit:
			findings = append(findings, c.checkComposite(mod, f, pool, node)...)
		}
		return true
	})
	return findings, nil
}

func (c *Cfg04) checkAssign(mod ModuleView, f *astutil.File, pool *astutil.Pool, stmt *ast.AssignStmt) []Finding {
	var out []Finding
	for i, lhs := range stmt.Lhs {
		if !isJWTKeyName(lhs) || i >= len(stmt.Rhs) {
			continue
		}
		if lit, ok := stmt.Rhs[i].(*ast.BasicLit); ok && lit.Kind.String() == "STRING" && lit.Value == `""` {
			out = append(out, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), f.RelPath, pool.Line(lhs)))
		}
	}
	return out
}

func (c *Cfg04) checkComposite(mod ModuleView, f *astutil.File, pool *astutil.Pool, lit *ast.CompositeLit) []Finding {
	var out []Finding
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok || !isJWTKeyName(kv.Key) {
			continue
		}
		if bl, ok := kv.Value.(*ast.BasicLit); ok && bl.Kind.String() == "STRING" && bl.Value == `""` {
			out = append(out, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), f.RelPath, pool.Line(kv.Key)))
		}
	}
	return out
}

func isJWTKeyName(expr ast.Expr) bool {
	name := exprString(expr)
	lower := strings.ToLower(name)
	return strings.Contains(lower, "jwt") && strings.Contains(lower, "secret")
}

func exprString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return e.Sel.Name
	default:
		return ""
	}
}
