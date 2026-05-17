package check

import (
	"context"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
)

// Err01 — _ = err и пустой if err != nil {} без обработки
type Err01 struct{ Base }

func NewErr01() *Err01 {
	return &Err01{Base: Base{meta: Meta{ID: "err-01", Domain: "errors", DefaultSeverity: SeverityWarn}}}
}

func (c *Err01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		switch node := n.(type) {
		case *ast.AssignStmt:
			if isIgnoredErrAssign(node) {
				findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(n)))
			}
		case *ast.IfStmt:
			if isEmptyErrNilCheck(node) {
				findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(n)))
			}
		}
		return true
	})
	return findings, nil
}

func isIgnoredErrAssign(stmt *ast.AssignStmt) bool {
	if stmt.Tok != token.ASSIGN && stmt.Tok != token.DEFINE {
		return false
	}
	if len(stmt.Lhs) != 1 || len(stmt.Rhs) != 1 {
		return false
	}
	lhs, ok := stmt.Lhs[0].(*ast.Ident)
	if !ok || lhs.Name != "_" {
		return false
	}
	rhs, ok := stmt.Rhs[0].(*ast.Ident)
	return ok && rhs.Name == "err"
}

func isEmptyErrNilCheck(stmt *ast.IfStmt) bool {
	if stmt.Body == nil || len(stmt.Body.List) != 0 {
		return false
	}
	return isErrNotNilExpr(stmt.Cond)
}

func isErrNotNilExpr(expr ast.Expr) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op.String() != "!=" {
		return false
	}
	return isErrIdent(bin.X) && isNilIdent(bin.Y) || isErrIdent(bin.Y) && isNilIdent(bin.X)
}

func isErrIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "err"
}

func isNilIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

// Err02 — panic в internal/
type Err02 struct{ Base }

func NewErr02() *Err02 {
	return &Err02{Base: Base{meta: Meta{ID: "err-02", Domain: "errors", DefaultSeverity: SeverityWarn}}}
}

func (c *Err02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if !astutil.IsInternalFile(f.RelPath) || astutil.IsTestFile(f.RelPath) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isBuiltinCall(call.Fun, "panic") {
			findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(n)))
		}
		return true
	})
	return findings, nil
}

func isBuiltinCall(fun ast.Expr, name string) bool {
	ident, ok := fun.(*ast.Ident)
	return ok && ident.Name == name
}

// Err03 — http.DefaultClient и пакетные http.Get/http.Post
type Err03 struct{ Base }

func NewErr03() *Err03 {
	return &Err03{Base: Base{meta: Meta{ID: "err-03", Domain: "errors", DefaultSeverity: SeverityFail}}}
}

func (c *Err03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		switch node := n.(type) {
		case *ast.CallExpr:
			if isHTTPCall(f, node.Fun, "Get", "Post") {
				findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), f.RelPath, pool.Line(n)))
			}
		case *ast.SelectorExpr:
			if isHTTPSelector(f, node, "DefaultClient") {
				findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), f.RelPath, pool.Line(n)))
			}
		}
		return true
	})
	return findings, nil
}

func isHTTPCall(f *astutil.File, fun ast.Expr, names ...string) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok || f.ImportPathOf(ident.Name) != "net/http" {
		return false
	}
	for _, n := range names {
		if sel.Sel.Name == n {
			return true
		}
	}
	return false
}

func isHTTPSelector(f *astutil.File, sel *ast.SelectorExpr, name string) bool {
	ident, ok := sel.X.(*ast.Ident)
	if !ok || f.ImportPathOf(ident.Name) != "net/http" {
		return false
	}
	return sel.Sel.Name == name
}

// Err04 — http.Client без timeout
type Err04 struct{ Base }

func NewErr04() *Err04 {
	return &Err04{Base: Base{meta: Meta{ID: "err-04", Domain: "errors", DefaultSeverity: SeverityFail}}}
}

func (c *Err04) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		if !isHTTPClientLiteral(f, lit) {
			return true
		}
		if clientHasZeroTimeout(lit) {
			findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), f.RelPath, pool.Line(n)))
		}
		return true
	})
	return findings, nil
}

func isHTTPClientLiteral(f *astutil.File, lit *ast.CompositeLit) bool {
	sel, ok := lit.Type.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return f.ImportPathOf(ident.Name) == "net/http" && sel.Sel.Name == "Client"
}

func clientHasZeroTimeout(lit *ast.CompositeLit) bool {
	if len(lit.Elts) == 0 {
		return true
	}
	hasTimeout := false
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Timeout" {
			continue
		}
		hasTimeout = true
		if isZeroValue(kv.Value) {
			return true
		}
	}
	return !hasTimeout
}

func isZeroValue(expr ast.Expr) bool {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name == "0"
	case *ast.BasicLit:
		return v.Value == "0"
	default:
		return false
	}
}

// Err05 — context.Background в handler
type Err05 struct{ Base }

func NewErr05() *Err05 {
	return &Err05{Base: Base{meta: Meta{ID: "err-05", Domain: "errors", DefaultSeverity: SeverityWarn}}}
}

func (c *Err05) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if !isHandlerPackage(f.RelPath) && !strings.Contains(filepath.ToSlash(f.RelPath), "/handler") {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isContextBackground(f, call.Fun) {
			findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(n)))
		}
		return true
	})
	return findings, nil
}

func isContextBackground(f *astutil.File, fun ast.Expr) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Background" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && f.ImportPathOf(ident.Name) == "context"
}

// Err06 — go func в for без лимита
type Err06 struct{ Base }

func NewErr06() *Err06 {
	return &Err06{Base: Base{meta: Meta{ID: "err-06", Domain: "errors", DefaultSeverity: SeverityWarn}}}
}

func (c *Err06) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		forStmt, ok := n.(*ast.ForStmt)
		if !ok {
			return true
		}
		ast.Inspect(forStmt.Body, func(n ast.Node) bool {
			goStmt, ok := n.(*ast.GoStmt)
			if !ok {
				return true
			}
			if _, ok := goStmt.Call.Fun.(*ast.FuncLit); ok {
				findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(goStmt)))
			}
			return true
		})
		return true
	})
	return findings, nil
}
