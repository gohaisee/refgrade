package check

import (
	"context"
	"path/filepath"
	"strings"
)

// Lay02 — бизнес-код вне internal/ без pkg/
type Lay02 struct{ Base }

func NewLay02() *Lay02 {
	return &Lay02{Base: Base{meta: Meta{ID: "lay-02", Domain: "layout", DefaultSeverity: SeverityWarn}}}
}

func (c *Lay02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	var findings []Finding
	for _, p := range mod.Packages() {
		rel := filepath.ToSlash(p.RelDir)
		if rel == "." || rel == "cmd" || strings.HasPrefix(rel, "cmd/") {
			continue
		}
		if strings.HasPrefix(rel, "internal/") || strings.HasPrefix(rel, "pkg/") {
			continue
		}
		if strings.HasPrefix(rel, "testdata/") || strings.HasPrefix(rel, "vendor/") {
			continue
		}
		if len(p.GoFiles) == 0 {
			continue
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), rel, 1))
	}
	return findings, nil
}
