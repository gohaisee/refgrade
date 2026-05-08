package check

import (
	"context"
	"strings"
)

// Lay03 — placeholder module path в go.mod
type Lay03 struct{ Base }

func NewLay03() *Lay03 {
	return &Lay03{Base: Base{meta: Meta{ID: "lay-03", Domain: "layout", DefaultSeverity: SeverityInfo}}}
}

func (c *Lay03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	modPath := parseModulePath(mod.GoModContent())
	if modPath == "" {
		return nil, nil
	}
	for _, re := range placeholderModulePaths {
		if re.MatchString(modPath) {
			return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityInfo), "go.mod", 1)}, nil
		}
	}
	return nil, nil
}

func parseModulePath(data []byte) string {
	for _, line := range splitLines(string(data)) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}
