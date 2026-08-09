package check

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/gohaisee/refgrade/internal/project"
)

// exposes project.Module to checkers
type ModuleAdapter struct {
	Mod *project.Module
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
			if strings.HasSuffix(name, ".go") {
				files = append(files, GoFile{Path: abs, RelPath: rel})
			}
		}
	}
	return files
}

// wraps loaded module for check package
func FromProject(mod *project.Module) ModuleView {
	return ModuleAdapter{Mod: mod}
}

// runs checkers and localizes metadata
func RunAll(ctx context.Context, mod *project.Module, checkers []Checker, translate func(string) string) ([]Finding, error) {
	view := FromProject(mod)
	var all []Finding
	for _, ch := range checkers {
		findings, err := ch.Run(ctx, view)
		if err != nil {
			return nil, err
		}
		all = append(all, findings...)
	}
	return SetMeta(all, translate), nil
}
