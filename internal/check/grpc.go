package check

import (
	"context"
	"go/ast"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
)

const grpcFatRPCLines = 50

var (
	grpcGates = []string{"grpc"}

	connectGates = []string{"connect"}

	gwGates = []string{"grpc-gateway"}

	grpcInterceptorHints = []string{
		"UnaryInterceptor",
		"StreamInterceptor",
		"ChainUnaryInterceptor",
		"ChainStreamInterceptor",
	}

	connectInterceptorHints = []string{
		"WithInterceptors",
		"UnaryInterceptorFunc",
		"Interceptor",
	}

	connectTimeoutHints = []string{
		"WithTimeout",
		"WithConnectTimeout",
		"connect.WithTimeout",
	}

	gwAuthHints = []string{
		"AuthMiddleware",
		"Authenticator",
		"middleware.Auth",
		"JWTAuth",
		"grpc_auth",
		"runtime.WithMiddleware",
	}

	gwRegisterHints = []string{
		"RegisterGatewayHandler",
		"RegisterMux",
		"runtime.NewServeMux",
		"grpc-gateway",
	}

	grpcMetadataTrustKeys = []string{
		"x-user-id",
		"x-is-admin",
	}
)

// insecure or cleartext grpc client dial
type Grpc01 struct{ Base }

func NewGrpc01() *Grpc01 {
	return &Grpc01{Base: Base{meta: Meta{ID: "grpc-01", Gates: grpcGates, Domain: "grpc", DefaultSeverity: SeverityFail}}}
}

func (c *Grpc01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if hasDevBuildTag(string(f.Src)) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isGrpcDialCall(f, call) {
			return true
		}
		sev := effectiveSeverity(mod, c.ID(), SeverityFail)
		if callHasInsecureTransport(f, call) {
			findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		} else if dialWithoutSecureTransport(f, call) {
			findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		}
		return true
	})
	return findings, nil
}

// grpc.NewServer without interceptors
type Grpc02 struct{ Base }

func NewGrpc02() *Grpc02 {
	return &Grpc02{Base: Base{meta: Meta{ID: "grpc-02", Gates: grpcGates, Domain: "grpc", DefaultSeverity: SeverityWarn}}}
}

func (c *Grpc02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	if moduleHasHint(pool, grpcInterceptorHints) {
		return nil, nil
	}
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isGrpcNewServerCall(f, call) {
			return true
		}
		if grpcServerCallHasInterceptors(f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// fat rpc handler with inline sql
type Grpc03 struct{ Base }

func NewGrpc03() *Grpc03 {
	return &Grpc03{Base: Base{meta: Meta{ID: "grpc-03", Gates: grpcGates, Domain: "grpc", DefaultSeverity: SeverityWarn}}}
}

func (c *Grpc03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}
		if funcLineCount(pool, fn) <= grpcFatRPCLines {
			return true
		}
		if !funcBodyHasDbQuery(f, fn.Body) {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(fn)))
		return true
	})
	return findings, nil
}

// trusted metadata identity
type Grpc04 struct{ Base }

func NewGrpc04() *Grpc04 {
	return &Grpc04{Base: Base{meta: Meta{ID: "grpc-04", Gates: grpcGates, Domain: "grpc", DefaultSeverity: SeverityFail}}}
}

func (c *Grpc04) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind.String() != "STRING" {
			return true
		}
		val := strings.Trim(lit.Value, `"`)
		if !isGrpcMetadataTrustKey(val) {
			return true
		}
		if !metadataTrustLiteralUsed(f, lit) {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), f.RelPath, pool.Line(lit)))
		return true
	})
	return findings, nil
}

// reflection on public server
type Grpc05 struct{ Base }

func NewGrpc05() *Grpc05 {
	return &Grpc05{Base: Base{meta: Meta{ID: "grpc-05", Gates: grpcGates, Domain: "grpc", DefaultSeverity: SeverityWarn}}}
}

func (c *Grpc05) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if hasDevBuildTag(string(f.Src)) || hasDevOrEnvGuard(string(f.Src)) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isReflectionRegisterCall(f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// shutdown without GracefulStop
type Grpc06 struct{ Base }

func NewGrpc06() *Grpc06 {
	return &Grpc06{Base: Base{meta: Meta{ID: "grpc-06", Gates: grpcGates, Domain: "grpc", DefaultSeverity: SeverityInfo}}}
}

func (c *Grpc06) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	if moduleHasHint(pool, []string{"GracefulStop"}) {
		return nil, nil
	}
	file, line, ok := firstGrpcServerSetup(pool)
	if !ok {
		return nil, nil
	}
	return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityInfo), file, line)}, nil
}

// connect client without timeout
type Conn01 struct{ Base }

func NewConn01() *Conn01 {
	return &Conn01{Base: Base{meta: Meta{ID: "conn-01", Gates: connectGates, Domain: "grpc", DefaultSeverity: SeverityWarn}}}
}

func (c *Conn01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if !ok || !isConnectNewClient(f, call) {
			return true
		}
		if connectClientHasTimeout(f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// connect handler without interceptors
type Conn02 struct{ Base }

func NewConn02() *Conn02 {
	return &Conn02{Base: Base{meta: Meta{ID: "conn-02", Gates: connectGates, Domain: "grpc", DefaultSeverity: SeverityWarn}}}
}

func (c *Conn02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	if moduleHasHint(pool, connectInterceptorHints) {
		return nil, nil
	}
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isConnectNewHandler(f, call) {
			return true
		}
		if connectHandlerHasInterceptors(f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// grpc-gateway without auth middleware
type Gw01 struct{ Base }

func NewGw01() *Gw01 {
	return &Gw01{Base: Base{meta: Meta{ID: "gw-01", Gates: gwGates, Domain: "grpc", DefaultSeverity: SeverityWarn}}}
}

func (c *Gw01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	if !moduleHasGatewaySetup(pool) {
		return nil, nil
	}
	if moduleHasHint(pool, gwAuthHints) {
		return nil, nil
	}
	file, line, ok := firstGatewaySetup(pool)
	if !ok {
		return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), "go.mod", 1)}, nil
	}
	return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), file, line)}, nil
}

// gateway registered before auth middleware
type Gw02 struct{ Base }

func NewGw02() *Gw02 {
	return &Gw02{Base: Base{meta: Meta{ID: "gw-02", Gates: gwGates, Domain: "grpc", DefaultSeverity: SeverityWarn}}}
}

func (c *Gw02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	var findings []Finding
	for _, f := range pool.Files {
		if !astutil.IsCmdFile(f.RelPath) && !strings.Contains(f.RelPath, "gateway") {
			continue
		}
		gwLine, authLine, ok := gatewayAuthOrder(pool, f)
		if !ok || gwLine >= authLine {
			continue
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, gwLine))
	}
	return findings, nil
}

func isGrpcDialCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	imp, name, ok := f.Selector(sel)
	if !ok || imp != "google.golang.org/grpc" {
		return false
	}
	switch name {
	case "Dial", "DialContext", "NewClient":
		return true
	default:
		return false
	}
}

func isGrpcNewServerCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	imp, name, ok := f.Selector(sel)
	return ok && imp == "google.golang.org/grpc" && name == "NewServer"
}

func callHasInsecureTransport(f *astutil.File, call *ast.CallExpr) bool {
	for _, arg := range call.Args {
		if exprUsesInsecureCredentials(f, arg) {
			return true
		}
	}
	return false
}

func dialWithoutSecureTransport(f *astutil.File, call *ast.CallExpr) bool {
	return !callHasSecureTransport(f, call)
}

func callHasSecureTransport(f *astutil.File, call *ast.CallExpr) bool {
	found := false
	for _, arg := range call.Args {
		ast.Inspect(arg, func(n ast.Node) bool {
			c, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := c.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			imp, name, ok := f.Selector(sel)
			if ok && imp == "google.golang.org/grpc" && name == "WithTransportCredentials" && len(c.Args) > 0 {
				if !exprUsesInsecureCredentials(f, c.Args[0]) {
					found = true
					return false
				}
			}
			if ok && strings.HasPrefix(imp, "google.golang.org/grpc/credentials/") && imp != "google.golang.org/grpc/credentials/insecure" {
				found = true
				return false
			}
			return true
		})
		if found {
			return true
		}
	}
	return false
}

func exprUsesInsecureCredentials(f *astutil.File, expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		imp, name, ok := f.Selector(sel)
		if ok && imp == "google.golang.org/grpc" && name == "WithInsecure" {
			found = true
			return false
		}
		if ok && imp == "google.golang.org/grpc/credentials/insecure" && name == "NewCredentials" {
			found = true
			return false
		}
		return true
	})
	return found
}

func grpcServerCallHasInterceptors(f *astutil.File, call *ast.CallExpr) bool {
	for _, arg := range call.Args {
		if callExprHasGrpcInterceptor(f, arg) {
			return true
		}
	}
	return false
}

func callExprHasGrpcInterceptor(f *astutil.File, expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		imp, name, ok := f.Selector(sel)
		if !ok || imp != "google.golang.org/grpc" {
			return true
		}
		for _, hint := range grpcInterceptorHints {
			if name == hint {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func funcBodyHasDbQuery(f *astutil.File, body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isDbQueryCall(f, call) {
			found = true
			return false
		}
		return true
	})
	return found
}

func isGrpcMetadataTrustKey(key string) bool {
	for _, k := range grpcMetadataTrustKeys {
		if strings.EqualFold(key, k) {
			return true
		}
	}
	return false
}

func metadataTrustLiteralUsed(f *astutil.File, lit *ast.BasicLit) bool {
	if !fileImportsPath(f, "google.golang.org/grpc/metadata") {
		return false
	}
	used := false
	ast.Inspect(f.AST, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isMetadataGetCall(call) {
			return true
		}
		for _, arg := range call.Args {
			if arg == lit {
				used = true
				return false
			}
		}
		return true
	})
	return used
}

func isMetadataGetCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "Get"
}

func isReflectionRegisterCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Register" {
		return false
	}
	imp, _, ok := f.Selector(sel)
	return ok && imp == "google.golang.org/grpc/reflection"
}

func firstGrpcServerSetup(pool *astutil.Pool) (file string, line int, ok bool) {
	for _, f := range pool.Files {
		if astutil.IsTestFile(f.RelPath) {
			continue
		}
		found := false
		pool.Inspect(func(ff *astutil.File, n ast.Node) bool {
			if ff.RelPath != f.RelPath {
				return true
			}
			call, isCall := n.(*ast.CallExpr)
			if !isCall || !isGrpcNewServerCall(ff, call) {
				return true
			}
			file = f.RelPath
			line = pool.Line(call)
			ok = true
			found = true
			return false
		})
		if found {
			return file, line, ok
		}
	}
	return "", 0, false
}

func isConnectNewClient(f *astutil.File, call *ast.CallExpr) bool {
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		return isConnectSelector(f, fun, "NewClient")
	case *ast.IndexExpr:
		return isConnectSelector(f, fun.X, "NewClient")
	case *ast.IndexListExpr:
		return isConnectSelector(f, fun.X, "NewClient")
	default:
		return false
	}
}

func isConnectNewHandler(f *astutil.File, call *ast.CallExpr) bool {
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		return isConnectSelector(f, fun, "NewUnaryHandler", "NewHandler")
	case *ast.IndexExpr:
		return isConnectSelector(f, fun.X, "NewUnaryHandler", "NewHandler")
	case *ast.IndexListExpr:
		return isConnectSelector(f, fun.X, "NewUnaryHandler", "NewHandler")
	default:
		return false
	}
}

func isConnectSelector(f *astutil.File, expr ast.Expr, names ...string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	imp, name, ok := f.Selector(sel)
	if !ok || imp != "connectrpc.com/connect" {
		return false
	}
	for _, n := range names {
		if name == n {
			return true
		}
	}
	return false
}

func connectClientHasTimeout(f *astutil.File, call *ast.CallExpr) bool {
	for _, arg := range call.Args {
		if connectOptionsHaveTimeout(f, arg) {
			return true
		}
	}
	return false
}

func connectOptionsHaveTimeout(f *astutil.File, expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		imp, name, ok := f.Selector(sel)
		if ok && imp == "connectrpc.com/connect" {
			for _, hint := range connectTimeoutHints {
				if name == strings.TrimPrefix(hint, "connect.") {
					found = true
					return false
				}
			}
		}
		if name == "WithTimeout" || name == "WithConnectTimeout" {
			found = true
			return false
		}
		return true
	})
	return found
}

func connectHandlerHasInterceptors(f *astutil.File, call *ast.CallExpr) bool {
	for _, arg := range call.Args {
		if connectOptionsHaveInterceptors(f, arg) {
			return true
		}
	}
	return false
}

func connectOptionsHaveInterceptors(f *astutil.File, expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		imp, name, ok := f.Selector(sel)
		if ok && imp == "connectrpc.com/connect" && name == "WithInterceptors" {
			found = true
			return false
		}
		return true
	})
	return found
}

func moduleHasGatewaySetup(pool *astutil.Pool) bool {
	for _, f := range pool.Files {
		if f.ContainsAny(gwRegisterHints...) {
			return true
		}
	}
	return false
}

func firstGatewaySetup(pool *astutil.Pool) (file string, line int, ok bool) {
	for _, f := range pool.Files {
		if astutil.IsTestFile(f.RelPath) {
			continue
		}
		src := string(f.Src)
		for _, hint := range gwRegisterHints {
			if idx := strings.Index(src, hint); idx >= 0 {
				return f.RelPath, strings.Count(src[:idx], "\n") + 1, true
			}
		}
	}
	return "", 0, false
}

func gatewayAuthOrder(pool *astutil.Pool, f *astutil.File) (gwLine, authLine int, ok bool) {
	gwLine = 0
	authLine = 0
	ast.Inspect(f.AST, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		line := pool.Line(call)
		if callLooksLikeGatewayRegister(f, call) && (gwLine == 0 || line < gwLine) {
			gwLine = line
		}
		if callLooksLikeGatewayAuth(pool, f, call) && (authLine == 0 || line < authLine) {
			authLine = line
		}
		return true
	})
	if gwLine == 0 {
		return 0, 0, false
	}
	if authLine == 0 {
		return gwLine, 0, true
	}
	return gwLine, authLine, true
}

func callLooksLikeGatewayRegister(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	name := sel.Sel.Name
	if strings.HasPrefix(name, "Register") && strings.Contains(name, "Handler") {
		return true
	}
	imp, n, ok := f.Selector(sel)
	if ok && strings.Contains(imp, "grpc-gateway") && (n == "NewServeMux" || strings.HasPrefix(n, "Register")) {
		return true
	}
	return false
}

func callLooksLikeGatewayAuth(pool *astutil.Pool, f *astutil.File, call *ast.CallExpr) bool {
	src := string(f.Src)
	line := pool.Line(call)
	start := 0
	for i := 1; i < line; i++ {
		start = strings.Index(src[start:], "\n") + 1 + start
	}
	end := strings.Index(src[start:], "\n")
	if end < 0 {
		end = len(src) - start
	}
	chunk := src[start : start+end]
	for _, hint := range gwAuthHints {
		if strings.Contains(chunk, hint) {
			return true
		}
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	for _, hint := range gwAuthHints {
		if strings.Contains(sel.Sel.Name, hint) {
			return true
		}
	}
	return false
}
