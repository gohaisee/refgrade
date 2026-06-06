package check

import (
	"context"
	"go/ast"
	"go/token"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
)

const (
	mongoDriverImport = "go.mongodb.org/mongo-driver/mongo"
	mongoBsonImport   = "go.mongodb.org/mongo-driver/bson"
)

var (
	mongoConnectCalls = map[string]struct{}{
		"Connect": {}, "NewClient": {},
	}

	mongoDataMethods = map[string]struct{}{
		"InsertOne": {}, "InsertMany": {}, "Find": {}, "FindOne": {},
		"UpdateOne": {}, "UpdateMany": {}, "DeleteOne": {}, "DeleteMany": {},
		"CountDocuments": {}, "Aggregate": {}, "Distinct": {},
		"FindOneAndUpdate": {}, "FindOneAndDelete": {}, "ReplaceOne": {},
		"BulkWrite": {},
	}

	mongoPoolSetters = map[string]struct{}{
		"SetMaxPoolSize": {}, "SetMinPoolSize": {},
	}
)

func mongoGates() []string {
	return []string{"mongo-driver"}
}

func mongoInspect(mod ModuleView) (*astutil.Pool, error) {
	return poolFor(mod, astutil.Filter{SkipTestFiles: true})
}

func isMongoDriverImport(imp string) bool {
	return imp == mongoDriverImport || strings.HasPrefix(imp, mongoDriverImport+"/")
}

func fileImportsMongo(f *astutil.File) bool {
	return fileImportsPath(f, mongoDriverImport)
}

func fileImportsBson(f *astutil.File) bool {
	return fileImportsPath(f, mongoBsonImport) ||
		fileImportsPath(f, "go.mongodb.org/mongo-driver/bson/primitive")
}

func isMongoConnectCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	imp, name, ok := selectorImport(f, sel)
	if !ok || !isMongoDriverImport(imp) {
		return false
	}
	_, ok = mongoConnectCalls[name]
	return ok
}

func isMongoDataCall(f *astutil.File, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	_, name, ok := selectorImport(f, sel)
	if !ok {
		name = sel.Sel.Name
	}
	_, ok = mongoDataMethods[name]
	if !ok {
		return false
	}
	if fileImportsMongo(f) {
		return true
	}
	return looksLikeDBReceiver(sel.X)
}

func isContextBackgroundExpr(f *astutil.File, expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	return isContextBackground(f, call.Fun)
}

func funcHasSetLimit(body *ast.BlockStmt) bool {
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
		if !ok || sel.Sel.Name != "SetLimit" {
			return true
		}
		found = true
		return false
	})
	return found
}

func bsonKeyValue(n ast.Node) (key string, value ast.Expr, ok bool) {
	switch e := n.(type) {
	case *ast.KeyValueExpr:
		lit, ok := e.Key.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return "", nil, false
		}
		return unquoteBasicLit(lit), e.Value, true
	case *ast.CallExpr:
		sel, ok := e.Fun.(*ast.SelectorExpr)
		if !ok {
			return "", nil, false
		}
		if sel.Sel.Name != "E" {
			return "", nil, false
		}
		if len(e.Args) < 2 {
			return "", nil, false
		}
		lit, ok := e.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return "", nil, false
		}
		return unquoteBasicLit(lit), e.Args[1], true
	default:
		return "", nil, false
	}
}

func isNonLiteralExpr(expr ast.Expr) bool {
	switch v := expr.(type) {
	case *ast.BasicLit:
		return false
	case *ast.Ident:
		return v.Name != "_"
	default:
		return true
	}
}

func funcUsesQuoteMeta(body *ast.BlockStmt) bool {
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
		if !ok || sel.Sel.Name != "QuoteMeta" {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if ok && ident.Name == "regexp" {
			found = true
			return false
		}
		return true
	})
	return found
}

func scanBsonFilters(body ast.Node, fn func(key string, value ast.Expr) bool) {
	if body == nil {
		return
	}
	ast.Inspect(body, func(n ast.Node) bool {
		comp, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		for _, elt := range comp.Elts {
			key, value, ok := bsonKeyValue(elt)
			if !ok {
				continue
			}
			if !fn(key, value) {
				return false
			}
		}
		return true
	})
}

// mongo.Connect inside handler package
type Mongo01 struct{ Base }

func NewMongo01() *Mongo01 {
	return &Mongo01{Base: Base{meta: Meta{ID: "mongo-01", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityFail}}}
}

func (c *Mongo01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mongoInspect(mod)
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
		if !ok || !isMongoConnectCall(f, call) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// mongo call with context.Background on request path
type Mongo02 struct{ Base }

func NewMongo02() *Mongo02 {
	return &Mongo02{Base: Base{meta: Meta{ID: "mongo-02", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityWarn}}}
}

func (c *Mongo02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mongoInspect(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if !isHandlerPackage(f.RelPath) && !isServicePackage(f.RelPath) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isMongoDataCall(f, call) || len(call.Args) == 0 {
			return true
		}
		if !isContextBackgroundExpr(f, call.Args[0]) {
			return true
		}
		findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
		return true
	})
	return findings, nil
}

// List*/FindAll* with Find and no SetLimit
type Mongo03 struct{ Base }

func NewMongo03() *Mongo03 {
	return &Mongo03{Base: Base{meta: Meta{ID: "mongo-03", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityWarn}}}
}

func (c *Mongo03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mongoInspect(mod)
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
		if !isListFuncName(fn.Name.Name) {
			return true
		}
		if funcHasSetLimit(fn.Body) {
			return true
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Find" {
				return true
			}
			if !fileImportsMongo(f) && !isMongoDataCall(f, call) {
				return true
			}
			findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(call)))
			return true
		})
		return true
	})
	return findings, nil
}

// no Ping after mongo connect in cmd main
type Mongo04 struct{ Base }

func NewMongo04() *Mongo04 {
	return &Mongo04{Base: Base{meta: Meta{ID: "mongo-04", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityInfo}}}
}

func (c *Mongo04) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mongoInspect(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityInfo)
	var findings []Finding
	cmdMongoConnect := make(map[string]int)
	cmdMongoPing := make(map[string]bool)
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if !astutil.IsCmdFile(f.RelPath) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isMongoConnectCall(f, call) {
			cmdMongoConnect[f.RelPath] = pool.Line(call)
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Ping" {
			cmdMongoPing[f.RelPath] = true
		}
		return true
	})
	for rel, line := range cmdMongoConnect {
		if !cmdMongoPing[rel] {
			findings = append(findings, finding(c.ID(), sev, rel, line))
		}
	}
	return findings, nil
}

// mongo connect without pool size options
type Mongo05 struct{ Base }

func NewMongo05() *Mongo05 {
	return &Mongo05{Base: Base{meta: Meta{ID: "mongo-05", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityInfo}}}
}

func (c *Mongo05) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mongoInspect(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityInfo)
	var findings []Finding
	fileConnect := make(map[string]int)
	filePoolOpts := make(map[string]bool)
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		if astutil.IsTestFile(f.RelPath) {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isMongoConnectCall(f, call) {
			fileConnect[f.RelPath] = pool.Line(call)
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if _, ok := mongoPoolSetters[sel.Sel.Name]; ok {
				filePoolOpts[f.RelPath] = true
			}
		}
		return true
	})
	for rel, line := range fileConnect {
		if !filePoolOpts[rel] {
			findings = append(findings, finding(c.ID(), sev, rel, line))
		}
	}
	return findings, nil
}

// $where with non-literal value in bson filter
type Mongo06 struct{ Base }

func NewMongo06() *Mongo06 {
	return &Mongo06{Base: Base{meta: Meta{ID: "mongo-06", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityFail}}}
}

func (c *Mongo06) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mongoInspect(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityFail)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}
		scanBsonFilters(fn.Body, func(key string, value ast.Expr) bool {
			if key != "$where" {
				return true
			}
			if !isNonLiteralExpr(value) {
				return true
			}
			findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(value)))
			return true
		})
		return true
	})
	return findings, nil
}

// user $regex without regexp.QuoteMeta in same function
type Mongo07 struct{ Base }

func NewMongo07() *Mongo07 {
	return &Mongo07{Base: Base{meta: Meta{ID: "mongo-07", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityWarn}}}
}

func (c *Mongo07) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	pool, err := mongoInspect(mod)
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), SeverityWarn)
	var findings []Finding
	pool.Inspect(func(f *astutil.File, n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}
		if funcUsesQuoteMeta(fn.Body) {
			return true
		}
		scanBsonFilters(fn.Body, func(key string, value ast.Expr) bool {
			if key != "$regex" {
				return true
			}
			if !isNonLiteralExpr(value) {
				return true
			}
			findings = append(findings, finding(c.ID(), sev, f.RelPath, pool.Line(value)))
			return true
		})
		return true
	})
	return findings, nil
}

// bson import in handler package
type Mongo08 struct{ Base }

func NewMongo08() *Mongo08 {
	return &Mongo08{Base: Base{meta: Meta{ID: "mongo-08", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityWarn}}}
}

func (c *Mongo08) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
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
			if fileImportsBson(f) {
				findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), f.RelPath, 1))
				break
			}
		}
	}
	return findings, nil
}
