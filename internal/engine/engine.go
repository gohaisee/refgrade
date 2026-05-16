package engine

import (
	"context"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/project"
	"github.com/gohaisee/refgrade/internal/refgradeconfig"
)

// aggregated scan output
type Result struct {
	Module   *project.Module
	Findings []check.Finding
	Statuses []check.CheckStatus
}

// scan options
type Options struct {
	Config       *refgradeconfig.Config
	Stacks       []string
	WithSecurity bool
}

// runs all checkers and localizes finding metadata
func Scan(ctx context.Context, mod *project.Module, opts Options, translate func(string) string) (*Result, error) {
	if opts.Config == nil {
		opts.Config = &refgradeconfig.Config{Checks: make(map[string]refgradeconfig.CheckSetting)}
	}
	findings, statuses, err := check.RunAll(ctx, mod, opts.Config, check.ScanOptions{
		Stacks:       opts.Stacks,
		WithSecurity: opts.WithSecurity,
	}, translate)
	if err != nil {
		return nil, err
	}
	return &Result{Module: mod, Findings: findings, Statuses: statuses}, nil
}

// whether result has fail-level findings
func HasFail(r *Result) bool {
	return check.HasSeverity(r.Findings, check.SeverityFail)
}
