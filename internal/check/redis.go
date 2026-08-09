package check

import (
	"context"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
)

var (
	redisImportPaths = []string{
		"github.com/redis/go-redis/v9",
		"github.com/redis/go-redis/v8",
		"github.com/go-redis/redis/v8",
	}

	redisClientConstructors = map[string]struct{}{
		"NewClient": {}, "NewClusterClient": {}, "NewFailoverClient": {},
	}

	redisDataMethods = map[string]struct{}{
		"Get": {}, "Set": {}, "Del": {}, "HGet": {}, "HSet": {}, "HMGet": {},
		"Keys": {}, "FlushAll": {}, "FlushDB": {}, "MGet": {}, "Pipelined": {},
		"SetNX": {}, "GetSet": {},
	}

	redisDangerousMethods = map[string]struct{}{
		"Keys": {}, "FlushAll": {}, "FlushDB": {},
	}

	redisAuthNameHints = []string{
		"auth", "session", "permission", "authorize", "isadmin", "role",
	}

	redisWriteHints = []string{
		"update", "insert", "create", "save", "write", "upsert",
	}
)

func redisGates() []string {
	return []string{"go-redis"}
}

func redisInspect(mod ModuleView) (*astutil.Pool, error) {
	return poolFor(mod, astutil.Filter{SkipTestFiles: true})
}

func isRedisImport(imp string) bool {
	for _, p := range redisImportPaths {
		if imp == p || strings.HasPrefix(imp, p+"/") {
			return true
		}
	}
	return false
}

func fileImportsRedis(f *astutil.File) bool {
	for _, imp := range f.Imports {
		if isRedisImport(imp) {
			return true
		}
	}
	return false
}

func isRedisClientCtor(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	imp, name, ok := selectorImport(f, sel)
	if !ok || !isRedisImport(imp) {
		return false
	}
	_, ok = redisClientConstructors[name]
	return ok
}

func isRedisDataCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	name := sel.Sel.Name
	if _, ok := redisDataMethods[name]; !ok {
		return false
	}
	if imp, resolved, ok := selectorImport(f, sel); ok && isRedisImport(imp) {
		_ = resolved
		return true
	}
	if fileImportsRedis(f) {
		return looksLikeDBReceiver(sel.X) || isRedisLikeReceiver(sel.X)
	}
	return false
}

func isRedisLikeReceiver(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	lower := strings.ToLower(ident.Name)
	return lower == "rdb" || lower == "redis" || lower == "cache" || strings.HasSuffix(lower, "redis")
}

func isRedisSetWithoutTTL(call *ast.CallExpr) bool {
	if len(call.Args) < 4 {
		return false
	}
	return isZeroDuration(call.Args[3])
}

func isZeroDuration(expr ast.Expr) bool {
	switch v := expr.(type) {
	case *ast.BasicLit:
		return v.Value == "0"
	case *ast.Ident:
		return v.Name == "0"
	case *ast.CallExpr:
		sel, ok := v.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		ident, ok := sel.X.(*ast.Ident)
		return ok && ident.Name == "time" && sel.Sel.Name == "Duration"
	default:
		return false
	}
}

func redisKeyArg(call *ast.CallExpr) (ast.Expr, bool) {
	if len(call.Args) < 2 {
		return nil, false
	}
	return call.Args[1], true
}

func stringLitValue(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	return unquoteBasicLit(lit), true
}

func keyMissingNamespace(key string) bool {
	return key != "" && !strings.Contains(key, ":")
}

func funcNameHintsAuth(name string) bool {
	lower := strings.ToLower(name)
	for _, hint := range redisAuthNameHints {
		if strings.Contains(lower, hint) {
			return true
		}
	}
	return false
}

func funcNameHintsWrite(name string) bool {
	lower := strings.ToLower(name)
	for _, hint := range redisWriteHints {
		if strings.Contains(lower, hint) {
			return true
		}
	}
	return false
}

func funcHasRedisGet(f *astutil.File, body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}
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
		switch sel.Sel.Name {
		case "Get", "HGet":
			if f == nil || isRedisDataCall(f, call) {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func funcHasRedisSet(body *ast.BlockStmt) bool {
	return funcHasRedisMethod(body, "Set", "HSet")
}

func funcHasRedisDel(body *ast.BlockStmt) bool {
	return funcHasRedisMethod(body, "Del")
}

func funcHasRedisMethod(body *ast.BlockStmt, names ...string) bool {
	if body == nil {
		return false
	}
	want := make(map[string]struct{}, len(names))
	for _, n := range names {
		want[n] = struct{}{}
	}
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
		if _, ok := want[sel.Sel.Name]; ok {
			found = true
			return false
		}
		return true
	})
	return found
}

func funcHasDBWrite(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}
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
		switch sel.Sel.Name {
		case "Exec", "ExecContext", "Create", "Save", "Update", "InsertOne", "UpdateOne", "ReplaceOne":
			found = true
			return false
		}
		return true
	})
	return found
}

func funcHasDBRead(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}
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
		switch sel.Sel.Name {
		case "Query", "QueryRow", "QueryRowContext", "Select", "Find", "FindOne", "First":
			found = true
			return false
		case "Get":
			if isRedisLikeReceiver(sel.X) {
				return true
			}
			if looksLikeDBReceiver(sel.X) {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func funcUsesSingleflight(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}
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
		if sel.Sel.Name == "Do" {
			if ident, ok := sel.X.(*ast.Ident); ok && strings.Contains(strings.ToLower(ident.Name), "flight") {
				found = true
				return false
			}
		}
		if sel.Sel.Name == "SetNX" {
			found = true
			return false
		}
		return true
	})
	return found
}

func inspectRedisLoopBody(pool *astutil.Pool, f *astutil.File, body ast.Node, id, sev string, findings *[]Finding) {
	if body == nil {
		return
	}
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Get" {
			return true
		}
		if !isRedisDataCall(f, call) {
			return true
		}
		*findings = append(*findings, finding(id, sev, f.RelPath, pool.Line(call)))
		return true
	})
}

func redisOptionsZeroPool(opts *ast.CompositeLit) bool {
	if opts == nil {
		return false
	}
	hasPoolField := false
	zeroPool := false
	for _, elt := range opts.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		ident, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		switch ident.Name {
		case "PoolSize", "MinIdleConns":
			hasPoolField = true
			if isZeroDuration(kv.Value) || isZeroInt(kv.Value) {
				zeroPool = true
			}
		}
	}
	return hasPoolField && zeroPool
}

func isZeroInt(expr ast.Expr) bool {
	switch v := expr.(type) {
	case *ast.BasicLit:
		return v.Value == "0"
	case *ast.Ident:
		return v.Name == "0"
	default:
		return false
	}
}

// redis.NewClient inside handler package
type Redis01 struct{ Base }

func NewRedis01() *Redis01 {
	return &Redis01{Base: Base{meta: Meta{ID: "redis-01", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityFail}}}
}

func (c *Redis01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := redisInspect(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityFail)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if !isHandlerPackage(f.RelPath) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isRedisClientCtor(f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// Set without expiration ttl
type Redis02 struct{ Base }

func NewRedis02() *Redis02 {
	return &Redis02{Base: Base{meta: Meta{ID: "redis-02", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityWarn}}}
}

func (c *Redis02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := redisInspect(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isRedisDataCall(f, call) {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Set" {
			return true
		}
		if !isRedisSetWithoutTTL(call) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// Keys or FlushAll outside cmd
type Redis03 struct{ Base }

func NewRedis03() *Redis03 {
	return &Redis03{Base: Base{meta: Meta{ID: "redis-03", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityFail}}}
}

func (c *Redis03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := redisInspect(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityFail)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if astutil.IsCmdFile(f.RelPath) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isRedisDataCall(f, call) {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if _, ok := redisDangerousMethods[sel.Sel.Name]; !ok {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// write path updates cache with Set instead of Del
type Redis04 struct{ Base }

func NewRedis04() *Redis04 {
	return &Redis04{Base: Base{meta: Meta{ID: "redis-04", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityWarn}}}
}

func (c *Redis04) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := redisInspect(mod)
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
		if !funcNameHintsWrite(fn.Name.Name) {
			return true
		}
		if !funcHasRedisSet(fn.Body) || !funcHasDBWrite(fn.Body) {
			return true
		}
		if funcHasRedisDel(fn.Body) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(fn)))
		return true
	})
	return findings, nil
}

// auth decision from redis Get without db read
type Redis05 struct{ Base }

func NewRedis05() *Redis05 {
	return &Redis05{Base: Base{meta: Meta{ID: "redis-05", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityFail}}}
}

func (c *Redis05) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := redisInspect(mod)
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
		if !isHandlerPackage(f.RelPath) && !strings.Contains(filepath.ToSlash(f.RelPath), "/auth") {
			return true
		}
		if !funcNameHintsAuth(fn.Name.Name) {
			return true
		}
		if !funcHasRedisGet(f, fn.Body) {
			return true
		}
		if funcHasDBRead(fn.Body) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(fn)))
		return true
	})
	return findings, nil
}

// redis Get inside for loop
type Redis06 struct{ Base }

func NewRedis06() *Redis06 {
	return &Redis06{Base: Base{meta: Meta{ID: "redis-06", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityInfo}}}
}

func (c *Redis06) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := redisInspect(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityInfo)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		switch loop := n.(type) {
		case *ast.ForStmt:
			inspectRedisLoopBody(pool, f, loop.Body, c.ID(), sev, &findings)
		case *ast.RangeStmt:
			inspectRedisLoopBody(pool, f, loop.Body, c.ID(), sev, &findings)
		}
		return true
	})
	return findings, nil
}

// cache key literal without namespace separator
type Redis07 struct{ Base }

func NewRedis07() *Redis07 {
	return &Redis07{Base: Base{meta: Meta{ID: "redis-07", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityInfo}}}
}

func (c *Redis07) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := redisInspect(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityInfo)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isRedisDataCall(f, call) {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch sel.Sel.Name {
		case "Get", "Set", "HGet", "HSet", "Del":
		default:
			return true
		}
		keyExpr, ok := redisKeyArg(call)
		if !ok {
			return true
		}
		key, ok := stringLitValue(keyExpr)
		if !ok || !keyMissingNamespace(key) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// cache miss loads db without singleflight or lock
type Redis08 struct{ Base }

func NewRedis08() *Redis08 {
	return &Redis08{Base: Base{meta: Meta{ID: "redis-08", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityInfo}}}
}

func (c *Redis08) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := redisInspect(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityInfo)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}
		if !funcHasRedisGet(f, fn.Body) || !funcHasDBRead(fn.Body) {
			return true
		}
		if funcUsesSingleflight(fn.Body) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(fn)))
		return true
	})
	return findings, nil
}

// redis.Options with zero pool fields
type Redis09 struct{ Base }

func NewRedis09() *Redis09 {
	return &Redis09{Base: Base{meta: Meta{ID: "redis-09", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityInfo}}}
}

func (c *Redis09) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := redisInspect(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityInfo)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if astutil.IsTestFile(f.RelPath) || astutil.IsCmdFile(f.RelPath) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isRedisClientCtor(f, call) {
			return true
		}
		if len(call.Args) == 0 {
			return true
		}
		if unary, ok := call.Args[0].(*ast.UnaryExpr); ok && unary.Op == token.AND {
			if comp, ok := unary.X.(*ast.CompositeLit); ok && redisOptionsZeroPool(comp) {
				findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
			}
		}
		return true
	})
	return findings, nil
}
