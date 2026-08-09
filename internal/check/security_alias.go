package check

import (
	"context"

	"github.com/gohaisee/refgrade/internal/refgradeconfig"
)

// alias checker runs another checker and rewrites finding ids
type aliasChecker struct {
	Base
	delegate Checker
	targetID string
}

func newAliasChecker(id string, meta Meta, delegate Checker) *aliasChecker {
	return &aliasChecker{
		Base:     Base{meta: Meta{ID: id, Gates: meta.Gates, Domain: "security", DefaultSeverity: meta.DefaultSeverity}},
		delegate: delegate,
		targetID: meta.ID,
	}
}

func (c *aliasChecker) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	cfg := mod.Config()
	if cfg != nil && cfg.CheckEnabled(c.targetID) {
		return nil, nil
	}
	findings, err := c.delegate.Run(ctx, aliasConfigView{ModuleView: mod, targetID: c.targetID})
	if err != nil {
		return nil, err
	}
	sev := effectiveSeverity(mod, c.ID(), c.meta.DefaultSeverity)
	out := make([]Finding, 0, len(findings))
	for _, f := range findings {
		if f.ID == c.targetID {
			f.ID = c.ID()
		}
		if f.Severity == "" {
			f.Severity = sev
		} else {
			f.Severity = effectiveSeverity(mod, c.ID(), f.Severity)
		}
		out = append(out, f)
	}
	return out, nil
}

// enables target id for delegate run when user disabled the base check
type aliasConfigView struct {
	ModuleView
	targetID string
}

func (v aliasConfigView) Config() *refgradeconfig.Config {
	base := v.ModuleView.Config()
	if base == nil {
		return &refgradeconfig.Config{Checks: map[string]refgradeconfig.CheckSetting{
			v.targetID: {Enabled: true},
		}}
	}
	checks := make(map[string]refgradeconfig.CheckSetting, len(base.Checks)+1)
	for k, s := range base.Checks {
		checks[k] = s
	}
	set := checks[v.targetID]
	set.Enabled = true
	checks[v.targetID] = set
	cloned := *base
	cloned.Checks = checks
	return &cloned
}

func NewSec13() *aliasChecker {
	meta, _ := MetaForID("obs-02")
	return newAliasChecker("sec-13", meta, NewObs02())
}

func NewSec14() *aliasChecker {
	meta, _ := MetaForID("rest-05")
	return newAliasChecker("sec-14", meta, NewRest05())
}

func NewSecG01() *aliasChecker {
	meta, _ := MetaForID("gql-05")
	return newAliasChecker("sec-g01", meta, NewGql05())
}

func NewSecG02() *aliasChecker {
	meta, _ := MetaForID("gql-04")
	return newAliasChecker("sec-g02", meta, NewGql04())
}

func NewSecG03() *aliasChecker {
	meta, _ := MetaForID("gql-04")
	return newAliasChecker("sec-g03", meta, NewGql04())
}

func NewSecDB01() *aliasChecker {
	meta, _ := MetaForID("sql-02")
	return newAliasChecker("sec-db01", meta, NewSql02())
}

func NewSecDB02() *aliasChecker {
	meta, _ := MetaForID("cfg-03")
	return newAliasChecker("sec-db02", meta, NewCfg03())
}

func NewSecDB03() *aliasChecker {
	meta, _ := MetaForID("pgx-03")
	return newAliasChecker("sec-db03", meta, NewPgx03())
}
