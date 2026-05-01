package engine

import (
	"context"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/project"
)

// aggregated scan output
type Result struct {
	Module   *project.Module
	Findings []check.Finding
}

// runs all checkers and localizes finding metadata
func Scan(ctx context.Context, mod *project.Module, checkers []check.Checker, translate func(string) string) (*Result, error) {
	findings, err := check.RunAll(ctx, mod, checkers, translate)
	if err != nil {
		return nil, err
	}
	return &Result{Module: mod, Findings: findings}, nil
}

// whether result has fail-level findings
func HasFail(r *Result) bool {
	return check.HasSeverity(r.Findings, check.SeverityFail)
}
