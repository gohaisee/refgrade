package check

import "context"

type scanContextKey struct{}

type scanContext struct {
	WithSecurity bool
}

func withScanContext(ctx context.Context, opts ScanOptions) context.Context {
	return context.WithValue(ctx, scanContextKey{}, scanContext{WithSecurity: opts.WithSecurity})
}

func scanContextFrom(ctx context.Context) scanContext {
	if v, ok := ctx.Value(scanContextKey{}).(scanContext); ok {
		return v
	}
	return scanContext{}
}

// govulncheck subprocess; warn when --with-security not set
type Sec16 struct{ Base }

func NewSec16() *Sec16 {
	return &Sec16{Base: Base{meta: Meta{
		ID: "sec-16", Domain: "security", DefaultSeverity: SeverityWarn,
	}}}
}

func (c *Sec16) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	if !scanContextFrom(ctx).WithSecurity {
		return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), "go.mod", 1)}, nil
	}
	return runGovulncheck(ctx, mod)
}
