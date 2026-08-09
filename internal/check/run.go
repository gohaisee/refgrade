package check

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
	"github.com/gohaisee/refgrade/internal/project"
	"github.com/gohaisee/refgrade/internal/refgradeconfig"
)

// exposes project.Module to checkers
type ModuleAdapter struct {
	Mod     *project.Module
	Cfg     *refgradeconfig.Config
	Exclude *refgradeconfig.ExcludeMatcher
}

func (a ModuleAdapter) Root() string {
	return a.Mod.Root
}

func (a ModuleAdapter) RelPath(file string) (string, error) {
	return a.Mod.RelPath(file)
}

func (a ModuleAdapter) GoSourceFiles() []GoFile {
	var files []GoFile
	for _, pkg := range a.Mod.Packages {
		for _, name := range pkg.GoFiles {
			abs := filepath.Join(pkg.Dir, name)
			rel, err := a.Mod.RelPath(abs)
			if err != nil {
				continue
			}
			if a.Excluded(rel) {
				continue
			}
			if strings.HasSuffix(name, ".go") {
				files = append(files, GoFile{Path: abs, RelPath: rel})
			}
		}
	}
	return files
}

func (a ModuleAdapter) Packages() []PackageInfo {
	var out []PackageInfo
	for _, pkg := range a.Mod.Packages {
		rel, err := a.Mod.RelPath(pkg.Dir)
		if err != nil {
			continue
		}
		out = append(out, PackageInfo{
			ImportPath: pkg.ImportPath,
			Dir:        pkg.Dir,
			RelDir:     rel,
			GoFiles:    append([]string(nil), pkg.GoFiles...),
		})
	}
	return out
}

func (a ModuleAdapter) GoModContent() []byte {
	data, err := os.ReadFile(filepath.Join(a.Mod.Root, "go.mod"))
	if err != nil {
		return nil
	}
	return data
}

func (a ModuleAdapter) Excluded(relPath string) bool {
	if a.Exclude == nil {
		return false
	}
	return a.Exclude.Excluded(relPath)
}

func (a ModuleAdapter) Config() *refgradeconfig.Config {
	return a.Cfg
}

func (a ModuleAdapter) ASTPool(filter astutil.Filter) (*astutil.Pool, error) {
	var src []astutil.SourceFile
	for _, gf := range a.GoSourceFiles() {
		src = append(src, astutil.SourceFile{AbsPath: gf.Path, RelPath: gf.RelPath})
	}
	return astutil.NewPool(src, a.Excluded, filter)
}

// wraps loaded module for check package
func FromProject(mod *project.Module, cfg *refgradeconfig.Config) ModuleView {
	if cfg == nil {
		cfg = &refgradeconfig.Config{Checks: make(map[string]refgradeconfig.CheckSetting)}
	}
	return ModuleAdapter{
		Mod:     mod,
		Cfg:     cfg,
		Exclude: refgradeconfig.NewExcludeMatcher(cfg.Exclude),
	}
}

// scan options for engine
type ScanOptions struct {
	Stacks       []string
	WithSecurity bool
}

// runs checkers with gating, config, and status rows
func RunAll(ctx context.Context, mod *project.Module, cfg *refgradeconfig.Config, opts ScanOptions, translate func(string) string) ([]Finding, []CheckStatus, error) {
	if cfg == nil {
		cfg = &refgradeconfig.Config{Checks: make(map[string]refgradeconfig.CheckSetting), Profile: refgradeconfig.ProfileStandard}
	}
	view := FromProject(mod, cfg)
	stackSet := make(map[string]struct{}, len(opts.Stacks))
	for _, s := range opts.Stacks {
		stackSet[s] = struct{}{}
	}

	var all []Finding
	var statuses []CheckStatus

	for _, entry := range Registry() {
		meta := entry.Meta
		if !cfg.CheckEnabled(meta.ID) {
			continue
		}
		if len(meta.Gates) > 0 && !gatesMatch(meta.Gates, stackSet) {
			statuses = append(statuses, CheckStatus{ID: meta.ID, Severity: SeverityNA, Count: 0})
			continue
		}

		ch := entry.Factory()
		findings, err := ch.Run(ctx, view)
		if err != nil {
			return nil, nil, err
		}

		sev := cfg.EffectiveSeverity(meta.ID, meta.DefaultSeverity)
		for i := range findings {
			if findings[i].Severity == "" {
				findings[i].Severity = sev
			} else {
				findings[i].Severity = cfg.EffectiveSeverity(meta.ID, findings[i].Severity)
			}
		}

		count := len(findings)
		statusSev := SeverityOK
		if count > 0 {
			statusSev = worstSeverity(findings)
		}
		statuses = append(statuses, CheckStatus{ID: meta.ID, Severity: statusSev, Count: count})
		all = append(all, findings...)
	}

	if opts.WithSecurity {
		secFindings, err := runGovulncheck(ctx, view)
		if err == nil {
			all = append(all, secFindings...)
		}
	}

	return SetMeta(all, translate), statuses, nil
}

func worstSeverity(findings []Finding) string {
	w := SeverityInfo
	for _, f := range findings {
		switch f.Severity {
		case SeverityFail:
			return SeverityFail
		case SeverityWarn:
			w = SeverityWarn
		}
	}
	return w
}
