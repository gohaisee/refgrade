package check

import (
	"context"
	"fmt"

	"github.com/gohaisee/refgrade/internal/astutil"
	"github.com/gohaisee/refgrade/internal/refgradeconfig"
)

const (
	SeverityOK   = "ok"
	SeverityInfo = "info"
	SeverityWarn = "warn"
	SeverityFail = "fail"
	SeverityNA   = "n/a"
)

// one check result at a source location
type Finding struct {
	ID       string `json:"id"`
	When     string `json:"when"`
	Why      string `json:"why"`
	Fix      string `json:"fix"`
	Severity string `json:"severity"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// one rule against a loaded module
type Checker interface {
	ID() string
	Run(ctx context.Context, mod ModuleView) ([]Finding, error)
}

// slice of project.Module needed by checkers
type ModuleView interface {
	Root() string
	RelPath(file string) (string, error)
	GoSourceFiles() []GoFile
	Packages() []PackageInfo
	GoModContent() []byte
	Excluded(relPath string) bool
	Config() *refgradeconfig.Config
	BuildTags() []string
	IncludeTests() bool
	ASTPool(filter astutil.Filter) (*astutil.Pool, error)
}

// one non-test Go source with absolute path
type GoFile struct {
	Path    string
	RelPath string
}

// package info for layout and import checks
type PackageInfo struct {
	ImportPath string
	Dir        string
	RelDir     string
	GoFiles    []string
}

// Known reports whether id is a built-in checker
func Known(id string) bool {
	_, ok := MetaForID(id)
	return ok
}

// Explain returns localized when/why/fix for a check id
func Explain(id string, translate func(string) string) (when, why, fix string, ok bool) {
	if !Known(id) {
		return "", "", "", false
	}
	meta := LocalizedMeta{Translate: translate}
	return meta.when(id), meta.why(id), meta.fix(id), true
}

// fills When/Why/Fix from i18n keys check.<id>.*
type LocalizedMeta struct {
	Translate func(key string) string
}

func (m LocalizedMeta) when(id string) string {
	return m.Translate("check." + id + ".when")
}

func (m LocalizedMeta) why(id string) string {
	return m.Translate("check." + id + ".why")
}

func (m LocalizedMeta) fix(id string) string {
	return m.Translate("check." + id + ".fix")
}

func (m LocalizedMeta) finding(id, severity, file string, line int) Finding {
	return Finding{
		ID:       id,
		When:     m.when(id),
		Why:      m.why(id),
		Fix:      m.fix(id),
		Severity: severity,
		File:     file,
		Line:     line,
	}
}

// applies translate to findings missing metadata
func SetMeta(findings []Finding, translate func(string) string) []Finding {
	meta := LocalizedMeta{Translate: translate}
	out := make([]Finding, len(findings))
	for i, f := range findings {
		out[i] = f
		if f.When == "" {
			out[i].When = meta.when(f.ID)
		}
		if f.Why == "" {
			out[i].Why = meta.why(f.ID)
		}
		if f.Fix == "" {
			out[i].Fix = meta.fix(f.ID)
		}
	}
	return out
}

// whether any finding matches severity
func HasSeverity(findings []Finding, severity string) bool {
	for _, f := range findings {
		if f.Severity == severity {
			return true
		}
	}
	return false
}

// findings count per check id
func CountByID(findings []Finding, id string) int {
	n := 0
	for _, f := range findings {
		if f.ID == id {
			n++
		}
	}
	return n
}

// checker cannot run on this module
type ErrUnsupported struct {
	Check string
	Why   string
}

func (e ErrUnsupported) Error() string {
	return fmt.Sprintf("check %s unsupported: %s", e.Check, e.Why)
}
