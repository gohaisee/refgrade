package check

import (
	"context"
	"regexp"
)

// Cfg03 — секреты в исходниках
type Cfg03 struct{ Base }

func NewCfg03() *Cfg03 {
	return &Cfg03{Base: Base{meta: Meta{ID: "cfg-03", Domain: "config", DefaultSeverity: SeverityFail}}}
}

func (c *Cfg03) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	var findings []Finding
	for _, gf := range mod.GoSourceFiles() {
		if mod.Excluded(gf.RelPath) {
			continue
		}
		for _, re := range secretPatterns {
			if loc := lineOfMatch(gf.Path, re); loc > 0 {
				findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityFail), gf.RelPath, loc))
				break
			}
		}
	}
	return findings, nil
}

func lineOfMatch(path string, re *regexp.Regexp) int {
	data, err := readFileLines(path)
	if err != nil {
		return 0
	}
	for i, line := range data {
		if re.MatchString(line) {
			return i + 1
		}
	}
	return 0
}

func readFileLines(path string) ([]string, error) {
	b, err := readFile(path)
	if err != nil {
		return nil, err
	}
	return splitLines(string(b)), nil
}
