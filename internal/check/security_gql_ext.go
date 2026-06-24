package check

import (
	"context"
	"go/ast"
	"go/token"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
)

// admin graphql mutation resolver without role middleware
type SecG07 struct{ Base }

func NewSecG07() *SecG07 {
	return &SecG07{Base: Base{meta: Meta{
		ID: "sec-g07", Gates: gqlGates, Domain: "security", DefaultSeverity: SeverityWarn,
	}}}
}

func (c *SecG07) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	if moduleHasHint(pool, roleHints) {
		return nil, nil
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if !isHandWrittenResolverRel(f.RelPath) {
			return true
		}
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil || fn.Name == nil {
			return true
		}
		if !isAdminResolverName(fn.Name.Name) {
			return true
		}
		if funcHasAuthHint(fn) || fileHasHint(string(f.Src), roleHints) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(fn)))
		return true
	})
	return findings, nil
}

// graphql subscription resolver without auth check
type SecG08 struct{ Base }

func NewSecG08() *SecG08 {
	return &SecG08{Base: Base{meta: Meta{
		ID: "sec-g08", Gates: gqlGates, Domain: "security", DefaultSeverity: SeverityWarn,
	}}}
}

func (c *SecG08) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if !isHandWrittenResolverRel(f.RelPath) {
			return true
		}
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil || fn.Name == nil {
			return true
		}
		if !isSubscriptionResolverName(fn.Name.Name) {
			return true
		}
		if funcHasAuthHint(fn) || fileHasHint(string(f.Src), authHints) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(fn)))
		return true
	})
	return findings, nil
}

// graphql route via http.Handle outside cmd setup
type SecG09 struct{ Base }

func NewSecG09() *SecG09 {
	return &SecG09{Base: Base{meta: Meta{
		ID: "sec-g09", Gates: gqlGates, Domain: "security", DefaultSeverity: SeverityInfo,
	}}}
}

func (c *SecG09) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityInfo)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if astutil.IsCmdFile(f.RelPath) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isGlobalHTTPHandleCall(f, call) {
			return true
		}
		if !callRegistersGraphQLRoute(call) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// graphql webhook resolver without signature verify
type SecG10 struct{ Base }

func NewSecG10() *SecG10 {
	return &SecG10{Base: Base{meta: Meta{
		ID: "sec-g10", Gates: gqlGates, Domain: "security", DefaultSeverity: SeverityFail,
	}}}
}

func (c *SecG10) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityFail)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if !isHandWrittenResolverRel(f.RelPath) {
			return true
		}
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil || fn.Name == nil {
			return true
		}
		if !isWebhookResolverName(fn.Name.Name) && !funcBodyHasWebhookRoute(fn) {
			return true
		}
		if funcHasHmacHint(fn) || fileHasHint(string(f.Src), hmacHints) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(fn)))
		return true
	})
	return findings, nil
}

func isAdminResolverName(name string) bool {
	return strings.Contains(strings.ToLower(name), "admin")
}

func isSubscriptionResolverName(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "subscribe") || strings.HasPrefix(lower, "subscription")
}

func isWebhookResolverName(name string) bool {
	return strings.Contains(strings.ToLower(name), "webhook")
}

func callRegistersGraphQLRoute(call *ast.CallExpr) bool {
	for _, arg := range call.Args {
		lit, ok := arg.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			continue
		}
		val := strings.ToLower(strings.Trim(lit.Value, `"`))
		if strings.Contains(val, "graphql") || strings.Contains(val, "query") {
			return true
		}
	}
	return false
}
