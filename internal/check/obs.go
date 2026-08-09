package check

import (
	"context"
	"go/ast"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
)

// Obs01 — fmt.Print* в internal/
type Obs01 struct{ Base }

func NewObs01() *Obs01 {
	return &Obs01{Base: Base{meta: Meta{ID: "obs-01", Domain: "observability", DefaultSeverity: SeverityWarn}}}
}

func (c *Obs01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if !astutil.IsInternalFile(f.RelPath) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isFmtPrint(f, call.Fun) {
			findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(n)))
		}
		return true
	})
	return findings, nil
}

func isFmtPrint(f *astutil.File, fun ast.Expr) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok || f.ImportPathOf(ident.Name) != "fmt" {
		return false
	}
	return strings.HasPrefix(sel.Sel.Name, "Print")
}

// Obs02 — password/token/otp в логах
type Obs02 struct{ Base }

func NewObs02() *Obs02 {
	return &Obs02{Base: Base{meta: Meta{ID: "obs-02", Domain: "observability", DefaultSeverity: SeverityFail}}}
}

func (c *Obs02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if !isLogCall(f, call.Fun) {
			return true
		}
		for _, arg := range call.Args {
			if argHasPII(arg) {
				findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), f.RelPath, pool.Line(arg)))
				break
			}
		}
		return true
	})
	return findings, nil
}

func isLogCall(f *astutil.File, fun ast.Expr) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	name := sel.Sel.Name
	if name != "Print" && name != "Printf" && name != "Println" && !strings.HasPrefix(name, "Log") {
		return false
	}
	if ident, ok := sel.X.(*ast.Ident); ok {
		pkg := f.ImportPathOf(ident.Name)
		return pkg == "log" || pkg == "log/slog" || pkg == "fmt"
	}
	return false
}

func argHasPII(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return isPIIName(e.Name)
	case *ast.SelectorExpr:
		return isPIIName(e.Sel.Name)
	case *ast.KeyValueExpr:
		return argHasPII(e.Value) || isPIIName(exprString(e.Key))
	default:
		return false
	}
}

func isPIIName(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "password") ||
		strings.Contains(lower, "token") ||
		strings.Contains(lower, "otp") ||
		strings.Contains(lower, "secret")
}

// Con01 — send на chan в handler без буфера
type Con01 struct{ Base }

func NewCon01() *Con01 {
	return &Con01{Base: Base{meta: Meta{ID: "con-01", Domain: "concurrency", DefaultSeverity: SeverityInfo}}}
}

func (c *Con01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if !isHandlerPackage(f.RelPath) {
			return true
		}
		send, ok := n.(*ast.SendStmt)
		if !ok {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityInfo), f.RelPath, pool.Line(send)))
		return true
	})
	return findings, nil
}
