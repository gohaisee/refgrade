package check

import (
	"context"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
)

var (
	jwtImports = []string{
		"github.com/golang-jwt/jwt",
		"github.com/golang-jwt/jwt/v5",
		"github.com/dgrijalva/jwt-go",
	}

	rateLimitHints = []string{
		"RateLimit", "rate.Limit", "Limiter", "Throttle", "limiter.",
		"RateLimiter", "LimitHandler", "tollbooth",
	}

	roleHints = []string{
		"RequireRole", "HasRole", "RBAC", "IsAdmin", "Authorize",
		"middleware.Auth", "RequireAdmin", "AdminOnly", "CheckRole",
	}

	stepUpHints = []string{
		"StepUp", "MFA", "2FA", "TwoFactor", "Reauth", "VerifyOTP", "OtpVerify",
	}

	hmacHints = []string{
		"VerifySignature", "HMAC", "CheckSignature", "ValidateSignature",
		"WebhookSecret", "VerifyWebhook", "stripe.Verify",
	}

	aliasLimitHints = []string{
		"MaxAliases", "AliasLimit", "maxAliases", "LimitAliases",
	}

	cookieAuthHints = []string{
		"CookieAuth", "SessionCookie", "Set-Cookie", "CookieSession",
	}

	csrfHints = []string{
		"CSRF", "csrf", "CsrfToken", "csrfToken", "X-CSRF",
	}

	dynamicTableNameRe = regexp.MustCompile(`(?i)(table|column|tbl|field)`)
)

// weak md5/sha1 near password hashing
type Sec01 struct{ Base }

func NewSec01() *Sec01 {
	return &Sec01{Base: Base{meta: Meta{ID: "sec-01", Domain: "security", DefaultSeverity: SeverityFail}}}
}

func (c *Sec01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if !fileImportsWeakHash(f) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isWeakHashCall(f, call) {
			return true
		}
		if !fileHasPasswordContext(f) && !funcHasPasswordContext(pool, f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// tls.Config InsecureSkipVerify
type Sec02 struct{ Base }

func NewSec02() *Sec02 {
	return &Sec02{Base: Base{meta: Meta{ID: "sec-02", Domain: "security", DefaultSeverity: SeverityFail}}}
}

func (c *Sec02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		lit, ok := n.(*ast.CompositeLit)
		if !ok || !isTLSConfigLiteral(f, lit) {
			return true
		}
		if tlsConfigSkipsVerify(lit) {
			findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(lit)))
		}
		return true
	})
	return findings, nil
}

// math/rand for tokens or session ids
type Sec03 struct{ Base }

func NewSec03() *Sec03 {
	return &Sec03{Base: Base{meta: Meta{ID: "sec-03", Domain: "security", DefaultSeverity: SeverityFail}}}
}

func (c *Sec03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if !fileImportsMathRand(f) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isMathRandCall(f, call) {
			return true
		}
		if !fileHasTokenContext(f) && !funcHasTokenContext(pool, f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// jwt parse without algorithm pin in callback
type Sec04 struct{ Base }

func NewSec04() *Sec04 {
	return &Sec04{Base: Base{meta: Meta{ID: "sec-04", Domain: "security", DefaultSeverity: SeverityFail}}}
}

func (c *Sec04) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if !fileImportsJWT(f) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isJWTParseCall(f, call) {
			return true
		}
		if jwtParsePinsAlgorithm(pool, f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// jwt SigningMethodNone
type Sec05 struct{ Base }

func NewSec05() *Sec05 {
	return &Sec05{Base: Base{meta: Meta{ID: "sec-05", Domain: "security", DefaultSeverity: SeverityFail}}}
}

func (c *Sec05) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if !fileImportsJWT(f) {
			return true
		}
		sel, ok := n.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "SigningMethodNone" {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(sel)))
		return true
	})
	return findings, nil
}

// os/exec with user-controlled args
type Sec07 struct{ Base }

func NewSec07() *Sec07 {
	return &Sec07{Base: Base{meta: Meta{ID: "sec-07", Domain: "security", DefaultSeverity: SeverityFail}}}
}

func (c *Sec07) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		call, ok := n.(*ast.CallExpr)
		if !ok || !isExecCommandCall(f, call) {
			return true
		}
		if !execHasUserControlledArg(f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// template.HTML on user input
type Sec08 struct{ Base }

func NewSec08() *Sec08 {
	return &Sec08{Base: Base{meta: Meta{ID: "sec-08", Domain: "security", DefaultSeverity: SeverityFail}}}
}

func (c *Sec08) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		call, ok := n.(*ast.CallExpr)
		if !ok || !isTemplateHTMLCall(f, call) {
			return true
		}
		if len(call.Args) == 0 || isStringLiteral(call.Args[0]) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// http client to url from user input
type Sec09 struct{ Base }

func NewSec09() *Sec09 {
	return &Sec09{Base: Base{meta: Meta{ID: "sec-09", Domain: "security", DefaultSeverity: SeverityWarn}}}
}

func (c *Sec09) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		call, ok := n.(*ast.CallExpr)
		if !ok || !isOutboundHTTPWithURLArg(call) {
			return true
		}
		urlArg := outboundHTTPURLArg(call)
		if urlArg == nil || isStringLiteral(urlArg) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// .env present but not gitignored
type Sec10 struct{ Base }

func NewSec10() *Sec10 {
	return &Sec10{Base: Base{meta: Meta{ID: "sec-10", Domain: "security", DefaultSeverity: SeverityFail}}}
}

func (c *Sec10) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	envPath := filepath.Join(mod.Root(), ".env")
	if _, err := os.Stat(envPath); err != nil {
		return nil, nil
	}
	if gitignoreCoversEnv(mod.Root()) {
		return nil, nil
	}
	rel, err := mod.RelPath(envPath)
	if err != nil {
		rel = ".env"
	}
	return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), rel, 1)}, nil
}

// net/http/pprof import without dev build tag
type Sec15 struct{ Base }

func NewSec15() *Sec15 {
	return &Sec15{Base: Base{meta: Meta{ID: "sec-15", Domain: "security", DefaultSeverity: SeverityWarn}}}
}

func (c *Sec15) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
	for _, f := range pool.Files {
		if astutil.IsTestFile(f.RelPath) || hasDevBuildTag(string(f.Src)) {
			continue
		}
		if fileImportsPath(f, "net/http/pprof") || strings.Contains(string(f.Src), "net/http/pprof") {
			findings = append(findings, finding(c.ID(), sev, f.RelPath, 1))
		}
	}
	return findings, nil
}

// handler uses path id without authz check
type SecR01 struct{ Base }

func NewSecR01() *SecR01 {
	return &SecR01{Base: Base{meta: Meta{
		ID: "sec-r01", Gates: restGates(), Domain: "security", DefaultSeverity: SeverityWarn,
	}}}
}

func (c *SecR01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if n == nil || !isHandlerPackage(f.RelPath) {
			return true
		}
		fn := enclosingFunc(pool, f, n)
		if fn == nil || fn.Body == nil {
			return true
		}
		if funcHasAuthHint(fn) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isIDParamRead(f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// json.Unmarshal into db model from request
type SecR03 struct{ Base }

func NewSecR03() *SecR03 {
	return &SecR03{Base: Base{meta: Meta{
		ID: "sec-r03", Gates: restGates(), Domain: "security", DefaultSeverity: SeverityFail,
	}}}
}

func (c *SecR03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
	for _, f := range pool.Files {
		if !isHandlerPackage(f.RelPath) || !fileImportsModelPackage(f) {
			continue
		}
		if !strings.Contains(string(f.Src), "json.Unmarshal") {
			continue
		}
		line := 1
		if idx := strings.Index(string(f.Src), "json.Unmarshal"); idx >= 0 {
			line = strings.Count(string(f.Src)[:idx], "\n") + 1
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, line))
	}
	return findings, nil
}

// login route without rate limiter
type SecR04 struct{ Base }

func NewSecR04() *SecR04 {
	return &SecR04{Base: Base{meta: Meta{
		ID: "sec-r04", Gates: restGates(), Domain: "security", DefaultSeverity: SeverityWarn,
	}}}
}

func (c *SecR04) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	if moduleHasHint(pool, rateLimitHints) {
		return nil, nil
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		path := strings.Trim(lit.Value, `"`)
		if !isLoginRoutePath(path) {
			return true
		}
		if !routeLiteralRegistered(f, lit) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(lit)))
		return true
	})
	return findings, nil
}

// admin route without role middleware
type SecR05 struct{ Base }

func NewSecR05() *SecR05 {
	return &SecR05{Base: Base{meta: Meta{
		ID: "sec-r05", Gates: restGates(), Domain: "security", DefaultSeverity: SeverityFail,
	}}}
}

func (c *SecR05) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		path := strings.Trim(lit.Value, `"`)
		if !isAdminRoutePath(path) {
			return true
		}
		if !routeLiteralRegistered(f, lit) {
			return true
		}
		if fileHasHint(string(f.Src), roleHints) || moduleHasHint(pool, roleHints) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(lit)))
		return true
	})
	return findings, nil
}

// otp/payment flow without step-up auth
type SecR06 struct{ Base }

func NewSecR06() *SecR06 {
	return &SecR06{Base: Base{meta: Meta{
		ID: "sec-r06", Gates: restGates(), Domain: "security", DefaultSeverity: SeverityWarn,
	}}}
}

func (c *SecR06) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil || fn.Name == nil {
			return true
		}
		if !isHandlerPackage(f.RelPath) && !astutil.IsCmdFile(f.RelPath) {
			return true
		}
		name := strings.ToLower(fn.Name.Name)
		if !isSensitiveFlowName(name) && !funcBodyHasSensitiveRoute(fn) {
			return true
		}
		if funcHasStepUpHint(fn) || fileHasHint(string(f.Src), stepUpHints) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(fn)))
		return true
	})
	return findings, nil
}

// http.Handle outside cmd router setup
type SecR09 struct{ Base }

func NewSecR09() *SecR09 {
	return &SecR09{Base: Base{meta: Meta{
		ID: "sec-r09", Gates: restGates(), Domain: "security", DefaultSeverity: SeverityInfo,
	}}}
}

func (c *SecR09) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// webhook handler without hmac verify
type SecR10 struct{ Base }

func NewSecR10() *SecR10 {
	return &SecR10{Base: Base{meta: Meta{
		ID: "sec-r10", Gates: restGates(), Domain: "security", DefaultSeverity: SeverityFail,
	}}}
}

func (c *SecR10) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil || fn.Name == nil {
			return true
		}
		if !isHandlerPackage(f.RelPath) && !astutil.IsCmdFile(f.RelPath) {
			return true
		}
		name := strings.ToLower(fn.Name.Name)
		if !strings.Contains(name, "webhook") && !funcBodyHasWebhookRoute(fn) {
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

// node(id) resolver without ownership check
type SecG04 struct{ Base }

func NewSecG04() *SecG04 {
	return &SecG04{Base: Base{meta: Meta{
		ID: "sec-g04", Gates: gqlGates, Domain: "security", DefaultSeverity: SeverityWarn,
	}}}
}

func (c *SecG04) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if !isNodeByIDResolver(fn.Name.Name) {
			return true
		}
		if funcHasAuthHint(fn) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(fn)))
		return true
	})
	return findings, nil
}

// graphql server without alias batch limit
type SecG05 struct{ Base }

func NewSecG05() *SecG05 {
	return &SecG05{Base: Base{meta: Meta{
		ID: "sec-g05", Gates: gqlGates, Domain: "security", DefaultSeverity: SeverityWarn,
	}}}
}

func (c *SecG05) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	if !moduleHasGraphQLServer(mod) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	if moduleHasHint(pool, aliasLimitHints) {
		return nil, nil
	}
	file, line, ok := firstGraphQLServerSetup(pool)
	if !ok {
		return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), "go.mod", 1)}, nil
	}
	return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), file, line)}, nil
}

// cookie auth with get-based graphql queries
type SecG06 struct{ Base }

func NewSecG06() *SecG06 {
	return &SecG06{Base: Base{meta: Meta{
		ID: "sec-g06", Gates: gqlGates, Domain: "security", DefaultSeverity: SeverityWarn,
	}}}
}

func (c *SecG06) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	if !moduleHasHint(pool, cookieAuthHints) {
		return nil, nil
	}
	if moduleHasHint(pool, csrfHints) {
		return nil, nil
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	for _, f := range pool.Files {
		if astutil.IsTestFile(f.RelPath) {
			continue
		}
		if !fileHasGetGraphQLQuery(f) {
			continue
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, 1))
	}
	return findings, nil
}

// dynamic table or column name from user input in sql
type SecDB05 struct{ Base }

func NewSecDB05() *SecDB05 {
	return &SecDB05{Base: Base{meta: Meta{
		ID: "sec-db05", Gates: sqlGates(), Domain: "security", DefaultSeverity: SeverityFail,
	}}}
}

func (c *SecDB05) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := sqlInspect(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityFail)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		switch v := n.(type) {
		case *ast.CallExpr:
			if isDynamicTableSprintf(v) {
				findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(v)))
			}
		case *ast.BinaryExpr:
			if isDynamicTableConcat(v) {
				findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(v)))
			}
		}
		return true
	})
	return findings, nil
}

func restGates() []string {
	return []string{"gin", "echo", "chi", "net/http"}
}

func fileImportsWeakHash(f *astutil.File) bool {
	return fileImportsPath(f, "crypto/md5", "crypto/sh1")
}

func isWeakHashCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel.Name {
	case "Sum", "New", "Sum256", "Sum512":
		imp, _, ok := f.Selector(sel)
		return ok && (imp == "crypto/md5" || imp == "crypto/sh1")
	default:
		return false
	}
}

func fileHasPasswordContext(f *astutil.File) bool {
	return strings.Contains(strings.ToLower(string(f.Src)), "password")
}

func funcHasPasswordContext(pool *astutil.Pool, f *astutil.File, call *ast.CallExpr) bool {
	fn := enclosingFunc(pool, f, call)
	if fn == nil || fn.Name == nil {
		return false
	}
	name := strings.ToLower(fn.Name.Name)
	return strings.Contains(name, "password") || strings.Contains(name, "passwd")
}

func isTLSConfigLiteral(f *astutil.File, lit *ast.CompositeLit) bool {
	sel, ok := lit.Type.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return f.ImportPathOf(ident.Name) == "crypto/tls" && sel.Sel.Name == "Config"
}

func tlsConfigSkipsVerify(lit *ast.CompositeLit) bool {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "InsecureSkipVerify" {
			continue
		}
		if ident, ok := kv.Value.(*ast.Ident); ok && ident.Name == "true" {
			return true
		}
	}
	return false
}

func fileImportsMathRand(f *astutil.File) bool {
	return fileImportsPath(f, "math/rand")
}

func isMathRandCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok || f.ImportPathOf(ident.Name) != "math/rand" {
		return false
	}
	switch sel.Sel.Name {
	case "Int", "Intn", "Int63", "Int63n", "Uint32", "Read":
		return true
	default:
		return false
	}
}

func fileHasTokenContext(f *astutil.File) bool {
	src := strings.ToLower(string(f.Src))
	for _, hint := range []string{"token", "session", "nonce", "secret", "otp"} {
		if strings.Contains(src, hint) {
			return true
		}
	}
	return false
}

func funcHasTokenContext(pool *astutil.Pool, f *astutil.File, call *ast.CallExpr) bool {
	fn := enclosingFunc(pool, f, call)
	if fn == nil || fn.Name == nil {
		return false
	}
	name := strings.ToLower(fn.Name.Name)
	for _, hint := range []string{"token", "session", "nonce", "secret", "otp"} {
		if strings.Contains(name, hint) {
			return true
		}
	}
	return false
}

func fileImportsJWT(f *astutil.File) bool {
	for _, imp := range jwtImports {
		if fileImportsPath(f, imp) {
			return true
		}
	}
	return false
}

func isJWTParseCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != "Parse" && sel.Sel.Name != "ParseWithClaims" {
		return false
	}
	imp, _, ok := f.Selector(sel)
	if !ok {
		return false
	}
	for _, jwtImp := range jwtImports {
		if imp == jwtImp {
			return true
		}
	}
	return strings.Contains(imp, "jwt")
}

func jwtParsePinsAlgorithm(pool *astutil.Pool, f *astutil.File, call *ast.CallExpr) bool {
	fn := enclosingFunc(pool, f, call)
	if fn == nil || fn.Body == nil {
		return false
	}
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.Ident:
			if strings.Contains(v.Name, "SigningMethod") || strings.Contains(v.Name, "Algorithm") {
				found = true
				return false
			}
		case *ast.SelectorExpr:
			if strings.Contains(v.Sel.Name, "Method") || strings.Contains(v.Sel.Name, "Alg") {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func isExecCommandCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != "Command" && sel.Sel.Name != "CommandContext" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return f.ImportPathOf(ident.Name) == "os/exec" || ident.Name == "exec"
}

func execHasUserControlledArg(f *astutil.File, call *ast.CallExpr) bool {
	for i, arg := range call.Args {
		if i == 0 && call.Fun.(*ast.SelectorExpr).Sel.Name == "CommandContext" {
			continue
		}
		if isStringLiteral(arg) {
			continue
		}
		if exprLooksLikeUserInput(f, arg) {
			return true
		}
	}
	return false
}

func isTemplateHTMLCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "HTML" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return f.ImportPathOf(ident.Name) == "html/template" || ident.Name == "template"
}

func isOutboundHTTPWithURLArg(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel.Name {
	case "Get", "Head", "Post", "PostForm":
		return true
	case "NewRequest", "NewRequestWithContext":
		return len(call.Args) >= 2
	default:
		return false
	}
}

func outboundHTTPURLArg(call *ast.CallExpr) ast.Expr {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	switch sel.Sel.Name {
	case "Get", "Head", "Post", "PostForm":
		if len(call.Args) > 0 {
			return call.Args[0]
		}
	case "NewRequest", "NewRequestWithContext":
		if len(call.Args) > 1 {
			return call.Args[1]
		}
	}
	return nil
}

func exprLooksLikeUserInput(f *astutil.File, expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isUserInputReadCall(f, call) {
			found = true
			return false
		}
		return true
	})
	return found
}

func isUserInputReadCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel.Name {
	case "Param", "URLParam", "Query", "FormValue", "PostFormValue", "Bind", "BindJSON", "ShouldBind":
		return true
	case "Get":
		if inner, ok := sel.X.(*ast.SelectorExpr); ok && inner.Sel.Name == "Query" {
			return true
		}
		return isHeaderReadCall(f, call)
	default:
		return false
	}
}

func isStringLiteral(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	return ok && lit.Kind == token.STRING
}

func gitignoreCoversEnv(root string) bool {
	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		return false
	}
	for _, line := range splitLines(string(data)) {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == ".env" || line == ".env*" || strings.HasSuffix(line, "/.env") {
			return true
		}
	}
	return false
}

func enclosingFunc(pool *astutil.Pool, f *astutil.File, n ast.Node) *ast.FuncDecl {
	if n == nil || f == nil || f.AST == nil {
		return nil
	}
	pos := n.Pos()
	for _, decl := range f.AST.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if fn.Pos() <= pos && pos <= fn.End() {
			return fn
		}
	}
	return nil
}

func isIDParamRead(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel.Name {
	case "Param":
		if len(call.Args) == 1 {
			if lit, ok := call.Args[0].(*ast.BasicLit); ok {
				val := strings.Trim(lit.Value, `"`)
				return val == "id" || strings.HasSuffix(val, "_id") || strings.HasSuffix(val, "Id")
			}
		}
	case "URLParam":
		if len(call.Args) == 2 {
			if lit, ok := call.Args[1].(*ast.BasicLit); ok {
				val := strings.Trim(lit.Value, `"`)
				return val == "id" || strings.HasSuffix(val, "_id")
			}
		}
	}
	return false
}

func isJSONUnmarshalCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Unmarshal" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return f.ImportPathOf(ident.Name) == "encoding/json" || ident.Name == "json"
}

func unmarshalTargetLooksLikeModel(f *astutil.File, expr ast.Expr) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	switch v := star.X.(type) {
	case *ast.CompositeLit:
		sel, ok := v.Type.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		return selectorLooksLikeModel(f, sel)
	case *ast.Ident:
		if fileImportsModelPackage(f) {
			return true
		}
		return looksLikeModelTypeName(v.Name)
	case *ast.SelectorExpr:
		return selectorLooksLikeModel(f, v)
	default:
		return false
	}
}

func selectorLooksLikeModel(f *astutil.File, sel *ast.SelectorExpr) bool {
	imp, _, ok := f.Selector(sel)
	if ok && (strings.Contains(imp, "/model") || strings.Contains(imp, "/models") || strings.Contains(imp, "/entity")) {
		return true
	}
	return looksLikeModelTypeName(sel.Sel.Name)
}

func fileImportsModelPackage(f *astutil.File) bool {
	for _, imp := range f.Imports {
		if strings.Contains(imp, "/model") || strings.Contains(imp, "/models") || strings.Contains(imp, "/entity") {
			return true
		}
	}
	return false
}

func looksLikeModelTypeName(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, "model") ||
		strings.HasSuffix(lower, "entity") ||
		lower == "user" || lower == "account" || lower == "order"
}

func isLoginRoutePath(path string) bool {
	lower := strings.ToLower(path)
	return strings.Contains(lower, "login") || strings.Contains(lower, "/auth")
}

func isAdminRoutePath(path string) bool {
	return strings.Contains(strings.ToLower(path), "admin")
}

func routeLiteralRegistered(f *astutil.File, lit *ast.BasicLit) bool {
	src := string(f.Src)
	return strings.Contains(src, "HandleFunc") ||
		strings.Contains(src, ".GET(") ||
		strings.Contains(src, ".POST(") ||
		strings.Contains(src, ".Route(") ||
		strings.Contains(src, "Handle(")
}

func fileHasHint(src string, hints []string) bool {
	for _, hint := range hints {
		if strings.Contains(src, hint) {
			return true
		}
	}
	return false
}

func isSensitiveFlowName(name string) bool {
	for _, hint := range []string{"otp", "payment", "transfer", "withdraw", "payout"} {
		if strings.Contains(name, hint) {
			return true
		}
	}
	return false
}

func funcBodyHasSensitiveRoute(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		val := strings.ToLower(strings.Trim(lit.Value, `"`))
		if strings.Contains(val, "otp") || strings.Contains(val, "payment") ||
			strings.Contains(val, "transfer") || strings.Contains(val, "withdraw") {
			found = true
			return false
		}
		return true
	})
	return found
}

func funcHasStepUpHint(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.Ident:
			for _, hint := range stepUpHints {
				if strings.Contains(v.Name, hint) {
					found = true
					return false
				}
			}
		case *ast.SelectorExpr:
			for _, hint := range stepUpHints {
				if strings.Contains(v.Sel.Name, hint) {
					found = true
					return false
				}
			}
		}
		return true
	})
	return found
}

func funcBodyHasWebhookRoute(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		if strings.Contains(strings.ToLower(strings.Trim(lit.Value, `"`)), "webhook") {
			found = true
			return false
		}
		return true
	})
	return found
}

func isGlobalHTTPHandleCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != "Handle" && sel.Sel.Name != "HandleFunc" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return f.ImportPathOf(ident.Name) == "net/http" || ident.Name == "http"
}

func funcHasHmacHint(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.Ident:
			for _, hint := range hmacHints {
				if strings.Contains(v.Name, hint) {
					found = true
					return false
				}
			}
		case *ast.SelectorExpr:
			for _, hint := range hmacHints {
				if strings.Contains(v.Sel.Name, hint) {
					found = true
					return false
				}
			}
		}
		return true
	})
	return found
}

func isNodeByIDResolver(name string) bool {
	if strings.HasPrefix(name, "Find") && strings.Contains(name, "ID") {
		return true
	}
	lower := strings.ToLower(name)
	return strings.Contains(lower, "node") && strings.Contains(lower, "id")
}

func isDynamicTableSprintf(call *ast.CallExpr) bool {
	if !isFmtSprintfSQL(call) {
		return false
	}
	for _, arg := range call.Args[1:] {
		if identNameMatchesDynamicTable(arg) {
			return true
		}
	}
	return false
}

func isDynamicTableConcat(bin *ast.BinaryExpr) bool {
	if bin.Op != token.ADD {
		return false
	}
	leftLit, leftIsLit := bin.X.(*ast.BasicLit)
	rightLit, rightIsLit := bin.Y.(*ast.BasicLit)
	if leftIsLit && leftLit.Kind == token.STRING {
		s := strings.ToUpper(unquoteBasicLit(leftLit))
		if strings.Contains(s, "FROM") || strings.Contains(s, "INTO") || strings.Contains(s, "UPDATE") {
			return identNameMatchesDynamicTable(bin.Y)
		}
	}
	if rightIsLit && rightLit.Kind == token.STRING {
		s := strings.ToUpper(unquoteBasicLit(rightLit))
		if strings.Contains(s, "WHERE") || strings.Contains(s, "SET") {
			return identNameMatchesDynamicTable(bin.X)
		}
	}
	return false
}

func identNameMatchesDynamicTable(expr ast.Expr) bool {
	switch v := expr.(type) {
	case *ast.Ident:
		return dynamicTableNameRe.MatchString(v.Name)
	case *ast.SelectorExpr:
		return dynamicTableNameRe.MatchString(v.Sel.Name)
	default:
		return false
	}
}
