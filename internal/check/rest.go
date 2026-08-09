package check

import (
	"context"
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
)

const restFatHandlerLines = 80

var (
	trustHeaderNames = []string{
		"X-User-Id",
		"X-User-ID",
		"X-Is-Admin",
	}

	bodyLimitHints = []string{
		"MaxBytesReader",
		"MaxBytesHandler",
		"BodyLimit",
		"LimitReader",
	}

	postRouteHints = []string{
		".POST(",
		".Post(",
		"http.MethodPost",
		`"POST"`,
	}

	recoverMiddlewareHints = []string{
		".Recovery(",
		"middleware.Recover",
		"Recoverer(",
		"gin.Default(",
	}

	serverSetupHints = []string{
		"ListenAndServe",
		".Run(",
		".Start(",
		"NewRouter(",
		"gin.New(",
		"echo.New(",
	}
)

// handler >80 lines without service calls
type Rest01 struct{ Base }

func NewRest01() *Rest01 {
	return &Rest01{Base: Base{meta: Meta{ID: "rest-01", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityWarn}}}
}

func (c *Rest01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if !ok || fn.Body == nil || fn.Name == nil {
			return true
		}
		if funcLineCount(pool, fn) <= restFatHandlerLines {
			return true
		}
		if funcCallsService(f, fn.Body) {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(fn)))
		return true
	})
	return findings, nil
}

// sql/db import in handler package
type Rest02 struct{ Base }

func NewRest02() *Rest02 {
	return &Rest02{Base: Base{meta: Meta{ID: "rest-02", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityFail}}}
}

func (c *Rest02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
			if !pathUnder(f.RelPath, p.RelDir) || astutil.IsTestFile(f.RelPath) {
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
	return findings, nil
}

// http.Server without read timeouts
type Rest03 struct{ Base }

func NewRest03() *Rest03 {
	return &Rest03{Base: Base{meta: Meta{ID: "rest-03", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityWarn}}}
}

func (c *Rest03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if !ok || !isHTTPServerLiteral(f, lit) {
			return true
		}
		if serverHasReadTimeouts(lit) {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(lit)))
		return true
	})
	return findings, nil
}

// post routes without body size limit
type Rest04 struct{ Base }

func NewRest04() *Rest04 {
	return &Rest04{Base: Base{meta: Meta{ID: "rest-04", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityWarn}}}
}

func (c *Rest04) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	if moduleHasBodyLimit(pool) {
		return nil, nil
	}
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if !fileHasPostRoute(f) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !callLooksLikePostRoute(f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(call)))
		return true
	})
	if len(findings) == 0 {
		pool.Inspect(func(f *astutil.File, n ast.Node) bool {
			if fileHasPostRoute(f) {
				findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, 1))
			}
			return true
		})
	}
	return findings, nil
}

// cors wildcard with credentials
type Rest05 struct{ Base }

func NewRest05() *Rest05 {
	return &Rest05{Base: Base{meta: Meta{ID: "rest-05", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityFail}}}
}

func (c *Rest05) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if corsWildcardWithCredentials(lit) {
			findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), f.RelPath, pool.Line(lit)))
		}
		return true
	})
	return findings, nil
}

// trust header auth
type Rest06 struct{ Base }

func NewRest06() *Rest06 {
	return &Rest06{Base: Base{meta: Meta{ID: "rest-06", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityFail}}}
}

func (c *Rest06) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if !isHandlerPackage(f.RelPath) && !astutil.IsCmdFile(f.RelPath) {
			return true
		}
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind.String() != "STRING" {
			return true
		}
		val := strings.Trim(lit.Value, `"`)
		if !isTrustHeaderName(val) {
			return true
		}
		if !trustHeaderUsedInCall(f, lit) {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), f.RelPath, pool.Line(lit)))
		return true
	})
	return findings, nil
}

// err.Error in response body
type Rest07 struct{ Base }

func NewRest07() *Rest07 {
	return &Rest07{Base: Base{meta: Meta{ID: "rest-07", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityWarn}}}
}

func (c *Rest07) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		call, ok := n.(*ast.CallExpr)
		if !ok || !restErrErrorCall(call) {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// gin debug mode outside dev build
type Rest08 struct{ Base }

func NewRest08() *Rest08 {
	return &Rest08{Base: Base{meta: Meta{ID: "rest-08", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityWarn}}}
}

func (c *Rest08) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
		if !ok || !isGinSetDebugMode(f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// gorilla/mux without migration note
type Rest09 struct{ Base }

func NewRest09() *Rest09 {
	return &Rest09{Base: Base{meta: Meta{ID: "rest-09", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityInfo}}}
}

func (c *Rest09) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	goMod := string(mod.GoModContent())
	if !strings.Contains(goMod, "github.com/gorilla/mux") {
		return nil, nil
	}
	if hasMigrationNote(mod, goMod) {
		return nil, nil
	}
	return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityInfo), "go.mod", 1)}, nil
}

// missing recover middleware at server edge
type Rest10 struct{ Base }

func NewRest10() *Rest10 {
	return &Rest10{Base: Base{meta: Meta{ID: "rest-10", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityWarn}}}
}

func (c *Rest10) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := poolFor(mod, astutil.Filter{SkipTestFiles: true})
	if err != nil {
		return nil, err
	}
	if moduleHasRecoverMiddleware(pool) {
		return nil, nil
	}
	file, line, ok := firstHTTPServerSetup(pool)
	if !ok {
		return nil, nil
	}
	return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), file, line)}, nil
}

func funcLineCount(pool *astutil.Pool, fn *ast.FuncDecl) int {
	start := pool.Fset.Position(fn.Pos()).Line
	end := pool.Fset.Position(fn.End()).Line
	if end < start {
		return 0
	}
	return end - start + 1
}

func funcCallsService(f *astutil.File, body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		imp, _, ok := f.Selector(sel)
		if !ok {
			return true
		}
		if strings.Contains(imp, "/service") || strings.Contains(imp, "/services") {
			found = true
			return false
		}
		return true
	})
	return found
}

func isHTTPServerLiteral(f *astutil.File, lit *ast.CompositeLit) bool {
	sel, ok := lit.Type.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return f.ImportPathOf(ident.Name) == "net/http" && sel.Sel.Name == "Server"
}

func serverHasReadTimeouts(lit *ast.CompositeLit) bool {
	hasHeader := false
	hasRead := false
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		switch key.Name {
		case "ReadHeaderTimeout":
			if !isZeroOrMissingTimeout(kv.Value) {
				hasHeader = true
			}
		case "ReadTimeout":
			if !isZeroOrMissingTimeout(kv.Value) {
				hasRead = true
			}
		}
	}
	return hasHeader || hasRead
}

func isZeroOrMissingTimeout(expr ast.Expr) bool {
	return isZeroValue(expr)
}

func moduleHasBodyLimit(pool *astutil.Pool) bool {
	for _, f := range pool.Files {
		if astutil.IsTestFile(f.RelPath) {
			continue
		}
		for _, hint := range bodyLimitHints {
			if strings.Contains(string(f.Src), hint) {
				return true
			}
		}
	}
	return false
}

func fileHasPostRoute(f *astutil.File) bool {
	src := string(f.Src)
	for _, hint := range postRouteHints {
		if strings.Contains(src, hint) {
			return true
		}
	}
	return false
}

func callLooksLikePostRoute(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel.Name {
	case "POST", "Post":
		return true
	default:
		return false
	}
}

func corsWildcardWithCredentials(lit *ast.CompositeLit) bool {
	allowCreds := false
	allowWildcard := false
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		switch key.Name {
		case "AllowCredentials":
			allowCreds = isTrueLiteral(kv.Value)
		case "AllowOrigins", "AllowOriginFunc":
			allowWildcard = exprHasWildcardOrigin(kv.Value)
		}
	}
	return allowCreds && allowWildcard
}

func isTrueLiteral(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "true"
}

func exprHasWildcardOrigin(expr ast.Expr) bool {
	switch v := expr.(type) {
	case *ast.BasicLit:
		return strings.Trim(v.Value, `"`) == "*"
	case *ast.CompositeLit:
		for _, elt := range v.Elts {
			if lit, ok := elt.(*ast.BasicLit); ok && strings.Trim(lit.Value, `"`) == "*" {
				return true
			}
		}
	}
	return false
}

func isTrustHeaderName(name string) bool {
	for _, h := range trustHeaderNames {
		if strings.EqualFold(name, h) {
			return true
		}
	}
	return false
}

func trustHeaderUsedInCall(f *astutil.File, lit *ast.BasicLit) bool {
	parent := false
	ast.Inspect(f.AST, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		for _, arg := range call.Args {
			if arg == lit {
				if isHeaderReadCall(f, call) {
					parent = true
					return false
				}
			}
		}
		return true
	})
	return parent
}

func isHeaderReadCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel.Name {
	case "Get", "GetHeader", "Header":
		return true
	default:
		return false
	}
}

func restErrErrorCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Error" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == "err"
}

func hasDevBuildTag(src string) bool {
	lines := strings.Split(src, "\n")
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		if !strings.HasPrefix(trim, "//") {
			break
		}
		body := strings.TrimSpace(strings.TrimPrefix(trim, "//"))
		var expr string
		switch {
		case strings.HasPrefix(body, "go:build "):
			expr = strings.TrimPrefix(body, "go:build ")
		case strings.HasPrefix(body, "+build "):
			expr = strings.TrimPrefix(body, "+build ")
		default:
			continue
		}
		if buildTagHasDev(expr) {
			return true
		}
	}
	return false
}

func buildTagHasDev(expr string) bool {
	for _, part := range strings.FieldsFunc(expr, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '|' || r == '&' || r == '!' || r == '(' || r == ')'
	}) {
		if part == "dev" || part == "development" {
			return true
		}
	}
	return false
}

func isGinSetDebugMode(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "SetMode" {
		return false
	}
	imp, _, ok := f.Selector(sel)
	if !ok || imp != "github.com/gin-gonic/gin" {
		return false
	}
	if len(call.Args) != 1 {
		return false
	}
	switch arg := call.Args[0].(type) {
	case *ast.SelectorExpr:
		return arg.Sel.Name == "DebugMode"
	case *ast.BasicLit:
		return strings.Trim(arg.Value, `"`) == "debug"
	default:
		return false
	}
}

func hasMigrationNote(mod ModuleView, goMod string) bool {
	lower := strings.ToLower(goMod)
	if strings.Contains(lower, "migration") || strings.Contains(lower, "migrate") {
		return true
	}
	if _, err := os.Stat(filepath.Join(mod.Root(), "MIGRATION.md")); err == nil {
		return true
	}
	return false
}

func moduleHasRecoverMiddleware(pool *astutil.Pool) bool {
	for _, f := range pool.Files {
		if astutil.IsTestFile(f.RelPath) {
			continue
		}
		src := string(f.Src)
		for _, hint := range recoverMiddlewareHints {
			if strings.Contains(src, hint) {
				return true
			}
		}
	}
	return false
}

func firstHTTPServerSetup(pool *astutil.Pool) (file string, line int, ok bool) {
	for _, f := range pool.Files {
		if astutil.IsTestFile(f.RelPath) {
			continue
		}
		if !astutil.IsCmdFile(f.RelPath) && !strings.Contains(f.RelPath, "server") {
			continue
		}
		found := false
		pool.Inspect(func(ff *astutil.File, n ast.Node) bool {
			if ff.RelPath != f.RelPath {
				return true
			}
			call, isCall := n.(*ast.CallExpr)
			if !isCall || !callLooksLikeServerSetup(call) {
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
		if f.ContainsAny(serverSetupHints...) {
			return f.RelPath, 1, true
		}
	}
	return "", 0, false
}

func callLooksLikeServerSetup(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name == "ListenAndServe"
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel.Name {
	case "ListenAndServe", "Run", "Start", "NewRouter":
		return true
	default:
		return false
	}
}
