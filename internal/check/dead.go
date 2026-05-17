package check

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gohaisee/refgrade/internal/subprocess"
)

// Dead01 — unreachable func via deadcode tool
type Dead01 struct{ Base }

func NewDead01() *Dead01 {
	return &Dead01{Base: Base{meta: Meta{ID: "dead-01", Domain: "dead-code", DefaultSeverity: SeverityWarn}}}
}

func (c *Dead01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	res, err := subprocess.RunInDir(ctx, mod.Root(), "deadcode", []string{"-test", "./..."}, 2*time.Minute)
	if err != nil {
		if errors.Is(err, subprocess.ErrNotFound) || strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		return nil, err
	}
	if res.ExitCode != 0 && len(res.Stdout) == 0 {
		return nil, nil
	}
	var findings []Finding
	for _, line := range splitLines(string(res.Stdout)) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		file, lineNum := parseDeadcodeLine(line)
		if file == "" {
			continue
		}
		rel := file
		if r, err := mod.RelPath(file); err == nil {
			rel = r
		}
		findings = append(findings, finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), rel, lineNum))
	}
	return findings, nil
}

func parseDeadcodeLine(line string) (string, int) {
	// deadcode: path/to/file.go:42: unreachable func foo
	parts := strings.SplitN(line, ":", 3)
	if len(parts) < 2 {
		return "", 0
	}
	var lineNum int
	_, _ = fmtSscanf(parts[1], &lineNum)
	return parts[0], lineNum
}

func fmtSscanf(s string, n *int) (int, error) {
	var x int
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		x = x*10 + int(c-'0')
	}
	*n = x
	return 1, nil
}

// Dead06 — go mod tidy would change go.mod
type Dead06 struct{ Base }

func NewDead06() *Dead06 {
	return &Dead06{Base: Base{meta: Meta{ID: "dead-06", Domain: "dead-code", DefaultSeverity: SeverityWarn}}}
}

func (c *Dead06) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	res, err := subprocess.RunInDir(ctx, mod.Root(), "go", []string{"mod", "tidy", "-diff"}, 2*time.Minute)
	if err != nil {
		return nil, err
	}
	out := strings.TrimSpace(string(res.Stdout) + string(res.Stderr))
	if out == "" {
		return nil, nil
	}
	return []Finding{finding(c.ID(), effectiveSeverity(mod, c.ID(), SeverityWarn), "go.mod", 1)}, nil
}

// govulncheck via --with-security (sec-16)
func runGovulncheck(ctx context.Context, mod ModuleView) ([]Finding, error) {
	res, err := subprocess.RunInDir(ctx, mod.Root(), "govulncheck", []string{"-json", "./..."}, 3*time.Minute)
	if err != nil {
		if errors.Is(err, subprocess.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var findings []Finding
	for _, line := range splitLines(string(res.Stdout)) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry struct {
			OSV string `json:"osv"`
		}
		if json.Unmarshal([]byte(line), &entry) != nil || entry.OSV == "" {
			continue
		}
		findings = append(findings, finding("sec-16", SeverityWarn, "go.mod", 1))
		break
	}
	return findings, nil
}
