package engine

import (
	"context"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/project"
	"github.com/gohaisee/refgrade/internal/refgradeconfig"
)

type Result struct {
	Module   *project.Module
	Findings []check.Finding
	Statuses []check.CheckStatus
}

type Options struct {
	Config       *refgradeconfig.Config
	Stacks       []string
	WithSecurity bool
	BuildTags    []string
	IncludeTests bool
}

func Scan(ctx context.Context, mod *project.Module, opts Options, translate func(string) string) (*Result, error) {
	if opts.Config == nil {
		opts.Config = &refgradeconfig.Config{Checks: make(map[string]refgradeconfig.CheckSetting)}
	}
	findings, statuses, err := check.RunAll(ctx, mod, opts.Config, check.ScanOptions{
		Stacks:       opts.Stacks,
		WithSecurity: opts.WithSecurity,
		BuildTags:    opts.BuildTags,
		IncludeTests: opts.IncludeTests,
	}, translate)
	if err != nil {
		return nil, err
	}
	return &Result{Module: mod, Findings: findings, Statuses: statuses}, nil
}

func ScanDeadcode(ctx context.Context, mod *project.Module, opts Options, translate func(string) string) (*Result, error) {
	if opts.Config == nil {
		opts.Config = &refgradeconfig.Config{Checks: make(map[string]refgradeconfig.CheckSetting)}
	}
	findings, statuses, err := check.RunDeadcode(ctx, mod, opts.Config, check.ScanOptions{
		BuildTags:    opts.BuildTags,
		IncludeTests: opts.IncludeTests,
	}, translate)
	if err != nil {
		return nil, err
	}
	return &Result{Module: mod, Findings: findings, Statuses: statuses}, nil
}

func HasFail(r *Result) bool {
	return check.HasSeverity(r.Findings, check.SeverityFail)
}

func HasWarnOrFail(r *Result) bool {
	return check.HasSeverity(r.Findings, check.SeverityFail) || check.HasSeverity(r.Findings, check.SeverityWarn)
}
