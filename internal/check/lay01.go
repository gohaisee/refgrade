package check

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// Lay01 — нет cmd/<app>/main.go
type Lay01 struct{ Base }

func NewLay01() *Lay01 {
	return &Lay01{Base: Base{meta: Meta{ID: "lay-01", Domain: "layout", DefaultSeverity: SeverityWarn}}}
}

func (c *Lay01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	hasCmdMain := false
	for _, gf := range mod.GoSourceFiles() {
		rel := filepath.ToSlash(gf.RelPath)
		if strings.HasPrefix(rel, "cmd/") && strings.HasSuffix(rel, "/main.go") {
			hasCmdMain = true
			break
		}
	}
	if hasCmdMain {
		return nil, nil
	}
	// app with go files but no cmd entry
	if len(mod.GoSourceFiles()) == 0 {
		return nil, nil
	}
	for _, gf := range mod.GoSourceFiles() {
		if strings.HasSuffix(gf.RelPath, "main.go") {
			return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), gf.RelPath, 1)}, nil
		}
	}
	return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), "go.mod", 1)}, nil
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}
