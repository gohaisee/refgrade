package check

import (
	"context"
	"go/ast"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
)

// Lyr01 — service импортирует db driver
type Lyr01 struct{ Base }

func NewLyr01() *Lyr01 {
	return &Lyr01{Base: Base{meta: Meta{ID: "lyr-01", Domain: "layering", DefaultSeverity: SeverityFail}}}
}

func (c *Lyr01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	var findings []Finding
	for _, p := range mod.Packages() {
		if !isServicePackage(p.RelDir) {
			continue
		}
		pool, err := poolFor(mod, astutil.DefaultFilter())
		if err != nil {
			return nil, err
		}
		for _, f := range pool.Files {
			if !pathUnder(f.RelPath, p.RelDir) {
				continue
			}
			for _, imp := range f.Imports {
				if _, ok := dbDriverImports[imp]; ok {
					findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), f.RelPath, 1))
					break
				}
			}
		}
	}
	findings = append(findings, layerForbidFindings(mod, c.ID())...)
	return findings, nil
}

// Lyr02 — handler импортирует internal/repo
type Lyr02 struct{ Base }

func NewLyr02() *Lyr02 {
	return &Lyr02{Base: Base{meta: Meta{ID: "lyr-02", Domain: "layering", DefaultSeverity: SeverityFail}}}
}

func (c *Lyr02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	var findings []Finding
	for _, p := range mod.Packages() {
		if !isHandlerPackage(p.RelDir) {
			continue
		}
		pool, err := poolFor(mod, astutil.DefaultFilter())
		if err != nil {
			return nil, err
		}
		for _, f := range pool.Files {
			if !pathUnder(f.RelPath, p.RelDir) {
				continue
			}
			for _, imp := range f.Imports {
				if strings.Contains(imp, "/internal/repo") || strings.HasSuffix(imp, "/repo") && strings.Contains(imp, "/internal/") {
					findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), f.RelPath, 1))
					break
				}
			}
		}
	}
	return findings, nil
}

// Lyr03 — New*/Init открывает сеть или db
type Lyr03 struct{ Base }

func NewLyr03() *Lyr03 {
	return &Lyr03{Base: Base{meta: Meta{ID: "lyr-03", Domain: "layering", DefaultSeverity: SeverityWarn}}}
}

func (c *Lyr03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if astutil.IsCmdFile(f.RelPath) || astutil.IsTestFile(f.RelPath) {
			return true
		}
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}
		if !isCtorName(fn.Name.Name) {
			return true
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if isNetworkSideEffect(f, call) {
				findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(call)))
			}
			return true
		})
		return true
	})
	return findings, nil
}

func isCtorName(name string) bool {
	return strings.HasPrefix(name, "New") || name == "Init" || strings.HasPrefix(name, "Initialize")
}

func isNetworkSideEffect(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	imp, name, ok := f.Selector(sel)
	if !ok {
		return false
	}
	methods, ok := networkSideEffectCalls[imp]
	if !ok {
		return false
	}
	_, ok = methods[name]
	return ok
}

// Lyr04 — package-level var db
type Lyr04 struct{ Base }

func NewLyr04() *Lyr04 {
	return &Lyr04{Base: Base{meta: Meta{ID: "lyr-04", Domain: "layering", DefaultSeverity: SeverityWarn}}}
}

func (c *Lyr04) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if !isServicePackage(f.RelPath) && !isHandlerPackage(f.RelPath) {
			return true
		}
		gen, ok := n.(*ast.GenDecl)
		if !ok || gen.Tok.String() != "var" {
			return true
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				lower := strings.ToLower(name.Name)
				if lower == "db" || strings.HasPrefix(lower, "db") {
					findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(name)))
				}
			}
		}
		return true
	})
	return findings, nil
}
