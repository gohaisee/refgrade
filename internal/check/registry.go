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
	}
}

func entry(meta Meta, factory func() Checker) RegistryEntry {
	return RegistryEntry{Meta: meta, Factory: factory}
}

// Catalog returns all built-in checkers (legacy)
func Catalog() []Checker {
	return CatalogForStack(nil)
}

// CatalogForStack returns checkers applicable to detected stacks
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
