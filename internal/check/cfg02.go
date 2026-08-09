package check

import (
	"context"
	"go/ast"

	"github.com/gohaisee/refgrade/internal/astutil"
)

// Cfg02 — много getenv без internal/config
type Cfg02 struct{ Base }

func NewCfg02() *Cfg02 {
	return &Cfg02{Base: Base{meta: Meta{ID: "cfg-02", Domain: "config", DefaultSeverity: SeverityWarn}}}
}

func (c *Cfg02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	if hasConfigPackage(mod) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true, SkipIntegrationE2E: true})
	if err != nil {
		return nil, err
	}
	count := 0
	var sampleFile string
	var sampleLine int
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if astutil.IsCmdFile(f.RelPath) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isGetenvCall(f, call.Fun) {
			return true
		}
		count++
		if sampleFile == "" {
			sampleFile = f.RelPath
			sampleLine = pool.Line(n)
		}
		return true
	})
	if count <= 3 {
		return nil, nil
	}
	return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), sampleFile, sampleLine)}, nil
}
