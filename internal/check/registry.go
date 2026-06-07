package check

// checker metadata
type Meta struct {
	ID              string
	Gates           []string // empty = universal
	Domain          string
	DefaultSeverity string
}

// extended checker with gates and default severity
type GatedChecker interface {
	Checker
	Meta() Meta
}

// base embed for check implementations
type Base struct {
	meta Meta
}

func (b Base) Meta() Meta { return b.meta }

func (b Base) ID() string { return b.meta.ID }

// registry entry
type RegistryEntry struct {
	Meta    Meta
	Factory func() Checker
}

// all registered checkers
func Registry() []RegistryEntry {
	return []RegistryEntry{
		entry(Meta{ID: "cfg-01", Domain: "config", DefaultSeverity: SeverityFail}, func() Checker { return NewCfg01() }),
		entry(Meta{ID: "cfg-02", Domain: "config", DefaultSeverity: SeverityWarn}, func() Checker { return NewCfg02() }),
		entry(Meta{ID: "cfg-03", Domain: "config", DefaultSeverity: SeverityFail}, func() Checker { return NewCfg03() }),
		entry(Meta{ID: "cfg-04", Domain: "config", DefaultSeverity: SeverityFail}, func() Checker { return NewCfg04() }),
		entry(Meta{ID: "lay-01", Domain: "layout", DefaultSeverity: SeverityWarn}, func() Checker { return NewLay01() }),
		entry(Meta{ID: "lay-02", Domain: "layout", DefaultSeverity: SeverityWarn}, func() Checker { return NewLay02() }),
		entry(Meta{ID: "lay-03", Domain: "layout", DefaultSeverity: SeverityInfo}, func() Checker { return NewLay03() }),
		entry(Meta{ID: "lyr-01", Domain: "layering", DefaultSeverity: SeverityFail}, func() Checker { return NewLyr01() }),
		entry(Meta{ID: "lyr-02", Domain: "layering", DefaultSeverity: SeverityFail}, func() Checker { return NewLyr02() }),
		entry(Meta{ID: "lyr-03", Domain: "layering", DefaultSeverity: SeverityWarn}, func() Checker { return NewLyr03() }),
		entry(Meta{ID: "lyr-04", Domain: "layering", DefaultSeverity: SeverityWarn}, func() Checker { return NewLyr04() }),
		entry(Meta{ID: "err-01", Domain: "errors", DefaultSeverity: SeverityWarn}, func() Checker { return NewErr01() }),
		entry(Meta{ID: "err-02", Domain: "errors", DefaultSeverity: SeverityWarn}, func() Checker { return NewErr02() }),
		entry(Meta{ID: "err-03", Domain: "errors", DefaultSeverity: SeverityFail}, func() Checker { return NewErr03() }),
		entry(Meta{ID: "err-04", Domain: "errors", DefaultSeverity: SeverityFail}, func() Checker { return NewErr04() }),
		entry(Meta{ID: "err-05", Domain: "errors", DefaultSeverity: SeverityWarn}, func() Checker { return NewErr05() }),
		entry(Meta{ID: "err-06", Domain: "errors", DefaultSeverity: SeverityWarn}, func() Checker { return NewErr06() }),
		entry(Meta{ID: "tst-01", Domain: "tests", DefaultSeverity: SeverityWarn}, func() Checker { return NewTst01() }),
		entry(Meta{ID: "tst-02", Domain: "tests", DefaultSeverity: SeverityWarn}, func() Checker { return NewTst02() }),
		entry(Meta{ID: "tst-03", Domain: "tests", DefaultSeverity: SeverityWarn}, func() Checker { return NewTst03() }),
		entry(Meta{ID: "obs-01", Domain: "observability", DefaultSeverity: SeverityWarn}, func() Checker { return NewObs01() }),
		entry(Meta{ID: "obs-02", Domain: "observability", DefaultSeverity: SeverityFail}, func() Checker { return NewObs02() }),
		entry(Meta{ID: "con-01", Domain: "concurrency", DefaultSeverity: SeverityInfo}, func() Checker { return NewCon01() }),
		entry(Meta{ID: "dead-01", Domain: "dead-code", DefaultSeverity: SeverityWarn}, func() Checker { return NewDead01() }),
		entry(Meta{ID: "dead-06", Domain: "dead-code", DefaultSeverity: SeverityWarn}, func() Checker { return NewDead06() }),
		entry(Meta{ID: "rest-01", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityWarn}, func() Checker { return NewRest01() }),
		entry(Meta{ID: "rest-02", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityFail}, func() Checker { return NewRest02() }),
		entry(Meta{ID: "rest-03", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityWarn}, func() Checker { return NewRest03() }),
		entry(Meta{ID: "rest-04", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityWarn}, func() Checker { return NewRest04() }),
		entry(Meta{ID: "rest-05", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityFail}, func() Checker { return NewRest05() }),
		entry(Meta{ID: "rest-06", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityFail}, func() Checker { return NewRest06() }),
		entry(Meta{ID: "rest-07", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityWarn}, func() Checker { return NewRest07() }),
		entry(Meta{ID: "rest-08", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityWarn}, func() Checker { return NewRest08() }),
		entry(Meta{ID: "rest-09", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityInfo}, func() Checker { return NewRest09() }),
		entry(Meta{ID: "rest-10", Gates: []string{"gin", "echo", "chi", "net/http"}, Domain: "rest", DefaultSeverity: SeverityWarn}, func() Checker { return NewRest10() }),
		entry(Meta{ID: "grpc-01", Gates: []string{"grpc"}, Domain: "grpc", DefaultSeverity: SeverityFail}, func() Checker { return NewGrpc01() }),
		entry(Meta{ID: "grpc-02", Gates: []string{"grpc"}, Domain: "grpc", DefaultSeverity: SeverityWarn}, func() Checker { return NewGrpc02() }),
		entry(Meta{ID: "grpc-03", Gates: []string{"grpc"}, Domain: "grpc", DefaultSeverity: SeverityWarn}, func() Checker { return NewGrpc03() }),
		entry(Meta{ID: "grpc-04", Gates: []string{"grpc"}, Domain: "grpc", DefaultSeverity: SeverityFail}, func() Checker { return NewGrpc04() }),
		entry(Meta{ID: "grpc-05", Gates: []string{"grpc"}, Domain: "grpc", DefaultSeverity: SeverityWarn}, func() Checker { return NewGrpc05() }),
		entry(Meta{ID: "grpc-06", Gates: []string{"grpc"}, Domain: "grpc", DefaultSeverity: SeverityInfo}, func() Checker { return NewGrpc06() }),
		entry(Meta{ID: "conn-01", Gates: []string{"connect"}, Domain: "grpc", DefaultSeverity: SeverityWarn}, func() Checker { return NewConn01() }),
		entry(Meta{ID: "conn-02", Gates: []string{"connect"}, Domain: "grpc", DefaultSeverity: SeverityWarn}, func() Checker { return NewConn02() }),
		entry(Meta{ID: "gw-01", Gates: []string{"grpc-gateway"}, Domain: "grpc", DefaultSeverity: SeverityWarn}, func() Checker { return NewGw01() }),
		entry(Meta{ID: "gw-02", Gates: []string{"grpc-gateway"}, Domain: "grpc", DefaultSeverity: SeverityWarn}, func() Checker { return NewGw02() }),
		entry(Meta{ID: "gql-01", Gates: []string{"gqlgen", "graphql-go", "graphql"}, Domain: "graphql", DefaultSeverity: SeverityFail}, func() Checker { return NewGql01() }),
		entry(Meta{ID: "gql-02", Gates: []string{"gqlgen", "graphql-go", "graphql"}, Domain: "graphql", DefaultSeverity: SeverityWarn}, func() Checker { return NewGql02() }),
		entry(Meta{ID: "gql-03", Gates: []string{"gqlgen"}, Domain: "graphql", DefaultSeverity: SeverityFail}, func() Checker { return NewGql03() }),
		entry(Meta{ID: "gql-04", Gates: []string{"gqlgen", "graphql-go", "graphql"}, Domain: "graphql", DefaultSeverity: SeverityWarn}, func() Checker { return NewGql04() }),
		entry(Meta{ID: "gql-05", Gates: []string{"gqlgen", "graphql-go", "graphql"}, Domain: "graphql", DefaultSeverity: SeverityWarn}, func() Checker { return NewGql05() }),
		entry(Meta{ID: "gql-06", Gates: []string{"gqlgen", "graphql-go", "graphql"}, Domain: "graphql", DefaultSeverity: SeverityFail}, func() Checker { return NewGql06() }),
		entry(Meta{ID: "gql-07", Gates: []string{"gqlgen"}, Domain: "graphql", DefaultSeverity: SeverityWarn}, func() Checker { return NewGql07() }),
		entry(Meta{ID: "gql-08", Gates: []string{"gqlgen", "graphql-go", "graphql"}, Domain: "graphql", DefaultSeverity: SeverityInfo}, func() Checker { return NewGql08() }),
		entry(Meta{ID: "gql-09", Gates: []string{"gqlgen", "graphql-go", "graphql"}, Domain: "graphql", DefaultSeverity: SeverityInfo}, func() Checker { return NewGql09() }),
		entry(Meta{ID: "gqlgen-01", Gates: []string{"gqlgen"}, Domain: "graphql", DefaultSeverity: SeverityFail}, func() Checker { return NewGqlgen01() }),
		entry(Meta{ID: "gqlgen-02", Gates: []string{"gqlgen"}, Domain: "graphql", DefaultSeverity: SeverityWarn}, func() Checker { return NewGqlgen02() }),
		entry(Meta{ID: "ggl-01", Gates: []string{"graphql-go", "graphql"}, Domain: "graphql", DefaultSeverity: SeverityWarn}, func() Checker { return NewGgl01() }),
		entry(Meta{ID: "ggl-02", Gates: []string{"graphql-go", "graphql"}, Domain: "graphql", DefaultSeverity: SeverityInfo}, func() Checker { return NewGgl02() }),
		entry(Meta{ID: "sql-01", Gates: sqlGates(), Domain: "sql", DefaultSeverity: SeverityWarn}, func() Checker { return NewSql01() }),
		entry(Meta{ID: "sql-02", Gates: sqlGates(), Domain: "sql", DefaultSeverity: SeverityFail}, func() Checker { return NewSql02() }),
		entry(Meta{ID: "sql-03", Gates: sqlGates(), Domain: "sql", DefaultSeverity: SeverityWarn}, func() Checker { return NewSql03() }),
		entry(Meta{ID: "sql-04", Gates: sqlGates(), Domain: "sql", DefaultSeverity: SeverityInfo}, func() Checker { return NewSql04() }),
		entry(Meta{ID: "sql-05", Gates: sqlGates(), Domain: "sql", DefaultSeverity: SeverityFail}, func() Checker { return NewSql05() }),
		entry(Meta{ID: "sql-06", Gates: sqlGates(), Domain: "sql", DefaultSeverity: SeverityWarn}, func() Checker { return NewSql06() }),
		entry(Meta{ID: "sql-07", Gates: sqlGates(), Domain: "sql", DefaultSeverity: SeverityInfo}, func() Checker { return NewSql07() }),
		entry(Meta{ID: "pgx-01", Gates: []string{"pgx"}, Domain: "sql", DefaultSeverity: SeverityWarn}, func() Checker { return NewPgx01() }),
		entry(Meta{ID: "pgx-02", Gates: []string{"pgx"}, Domain: "sql", DefaultSeverity: SeverityInfo}, func() Checker { return NewPgx02() }),
		entry(Meta{ID: "pgx-03", Gates: []string{"pgx"}, Domain: "sql", DefaultSeverity: SeverityFail}, func() Checker { return NewPgx03() }),
		entry(Meta{ID: "gorm-01", Gates: []string{"gorm"}, Domain: "sql", DefaultSeverity: SeverityFail}, func() Checker { return NewGorm01() }),
		entry(Meta{ID: "gorm-02", Gates: []string{"gorm"}, Domain: "sql", DefaultSeverity: SeverityWarn}, func() Checker { return NewGorm02() }),
		entry(Meta{ID: "gorm-03", Gates: []string{"gorm"}, Domain: "sql", DefaultSeverity: SeverityWarn}, func() Checker { return NewGorm03() }),
		entry(Meta{ID: "gorm-04", Gates: []string{"gorm"}, Domain: "sql", DefaultSeverity: SeverityWarn}, func() Checker { return NewGorm04() }),
		entry(Meta{ID: "gorm-05", Gates: []string{"gorm"}, Domain: "sql", DefaultSeverity: SeverityWarn}, func() Checker { return NewGorm05() }),
		entry(Meta{ID: "sqlx-01", Gates: []string{"sqlx"}, Domain: "sql", DefaultSeverity: SeverityWarn}, func() Checker { return NewSqlx01() }),
		entry(Meta{ID: "sqlc-01", Gates: []string{"sqlc"}, Domain: "sql", DefaultSeverity: SeverityWarn}, func() Checker { return NewSqlc01() }),
		entry(Meta{ID: "ent-01", Gates: []string{"ent"}, Domain: "sql", DefaultSeverity: SeverityWarn}, func() Checker { return NewEnt01() }),
		entry(Meta{ID: "mongo-01", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityFail}, func() Checker { return NewMongo01() }),
		entry(Meta{ID: "mongo-02", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityWarn}, func() Checker { return NewMongo02() }),
		entry(Meta{ID: "mongo-03", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityWarn}, func() Checker { return NewMongo03() }),
		entry(Meta{ID: "mongo-04", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityInfo}, func() Checker { return NewMongo04() }),
		entry(Meta{ID: "mongo-05", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityInfo}, func() Checker { return NewMongo05() }),
		entry(Meta{ID: "mongo-06", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityFail}, func() Checker { return NewMongo06() }),
		entry(Meta{ID: "mongo-07", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityWarn}, func() Checker { return NewMongo07() }),
		entry(Meta{ID: "mongo-08", Gates: mongoGates(), Domain: "mongodb", DefaultSeverity: SeverityWarn}, func() Checker { return NewMongo08() }),
		entry(Meta{ID: "redis-01", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityFail}, func() Checker { return NewRedis01() }),
		entry(Meta{ID: "redis-02", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityWarn}, func() Checker { return NewRedis02() }),
		entry(Meta{ID: "redis-03", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityFail}, func() Checker { return NewRedis03() }),
		entry(Meta{ID: "redis-04", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityWarn}, func() Checker { return NewRedis04() }),
		entry(Meta{ID: "redis-05", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityFail}, func() Checker { return NewRedis05() }),
		entry(Meta{ID: "redis-06", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityInfo}, func() Checker { return NewRedis06() }),
		entry(Meta{ID: "redis-07", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityInfo}, func() Checker { return NewRedis07() }),
		entry(Meta{ID: "redis-08", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityInfo}, func() Checker { return NewRedis08() }),
		entry(Meta{ID: "redis-09", Gates: redisGates(), Domain: "redis", DefaultSeverity: SeverityInfo}, func() Checker { return NewRedis09() }),

	}
}

func entry(meta Meta, factory func() Checker) RegistryEntry {
	return RegistryEntry{Meta: meta, Factory: factory}
}

// legacy catalog of all built-in checkers
func Catalog() []Checker {
	return CatalogForStack(nil)
}

// checkers applicable to detected stacks
func CatalogForStack(stacks []string) []Checker {
	stackSet := make(map[string]struct{}, len(stacks))
	for _, s := range stacks {
		stackSet[s] = struct{}{}
	}
	var out []Checker
	for _, e := range Registry() {
		if !gatesMatch(e.Meta.Gates, stackSet) {
			continue
		}
		out = append(out, e.Factory())
	}
	return out
}

// all check ids including gated-off for status rows
func AllCheckIDs() []string {
	seen := make(map[string]struct{})
	var ids []string
	for _, e := range Registry() {
		if _, ok := seen[e.Meta.ID]; ok {
			continue
		}
		seen[e.Meta.ID] = struct{}{}
		ids = append(ids, e.Meta.ID)
	}
	return ids
}

func MetaForID(id string) (Meta, bool) {
	for _, e := range Registry() {
		if e.Meta.ID == id {
			return e.Meta, true
		}
	}
	return Meta{}, false
}

func gatesMatch(gates []string, stacks map[string]struct{}) bool {
	if len(gates) == 0 {
		return true
	}
	for _, g := range gates {
		if _, ok := stacks[g]; ok {
			return true
		}
	}
	return false
}

// whether check should be n/a for stacks
func IsNA(meta Meta, stacks []string) bool {
	if len(meta.Gates) == 0 {
		return false
	}
	stackSet := make(map[string]struct{}, len(stacks))
	for _, s := range stacks {
		stackSet[s] = struct{}{}
	}
	return !gatesMatch(meta.Gates, stackSet)
}

// per-check scan status
type CheckStatus struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Count    int    `json:"count"`
}
