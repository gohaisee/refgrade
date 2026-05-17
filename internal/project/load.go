package project

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// loaded Go module: root, mod path, packages
type Module struct {
	Root     string
	ModPath  string
	Packages []*Package
}

// one package directory with non-test Go files
type Package struct {
	ImportPath string
	Dir        string
	GoFiles    []string
	GoTestFiles []string
}

// walks up from path to directory with go.mod
func ModuleRoot(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	root, _, err := findModuleRoot(abs)
	return root, err
}

// resolves module root at path and lists packages via go list
func Load(ctx context.Context, path string) (*Module, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	root, modPath, err := findModuleRoot(abs)
	if err != nil {
		return nil, err
	}

	pkgs, err := listPackages(ctx, root)
	if err != nil {
		return nil, err
	}

	return &Module{
		Root:     root,
		ModPath:  modPath,
		Packages: pkgs,
	}, nil
}

func findModuleRoot(start string) (string, string, error) {
	dir := start
	for {
		modFile := filepath.Join(dir, "go.mod")
		data, err := osReadFile(modFile)
		if err == nil {
			modPath := parseModulePath(data)
			if modPath == "" {
				return "", "", fmt.Errorf("parse module path in %s", modFile)
			}
			return dir, modPath, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", fmt.Errorf("no go.mod found from %s", start)
		}
		dir = parent
	}
}

func parseModulePath(data []byte) string {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

type listPackageJSON struct {
	ImportPath  string   `json:"ImportPath"`
	Dir         string   `json:"Dir"`
	GoFiles     []string `json:"GoFiles"`
	GoTestFiles []string `json:"GoTestFiles"`
	Error       *struct {
		Err string `json:"Err"`
	} `json:"Error"`
}

func listPackages(ctx context.Context, root string) ([]*Package, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-json", "./...")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("go list: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("go list: %w", err)
	}

	var pkgs []*Package
	dec := json.NewDecoder(strings.NewReader(string(out)))
	for dec.More() {
		var raw listPackageJSON
		if err := dec.Decode(&raw); err != nil {
			return nil, fmt.Errorf("decode go list json: %w", err)
		}
		if raw.Error != nil {
			return nil, fmt.Errorf("go list package: %s", raw.Error.Err)
		}
		if raw.ImportPath == "" || raw.Dir == "" {
			continue
		}
		pkgs = append(pkgs, &Package{
			ImportPath:  raw.ImportPath,
			Dir:         raw.Dir,
			GoFiles:     append([]string(nil), raw.GoFiles...),
			GoTestFiles: append([]string(nil), raw.GoTestFiles...),
		})
	}
	return pkgs, nil
}

// path relative to module root, slash separators
func (m *Module) RelPath(file string) (string, error) {
	rel, err := filepath.Rel(m.Root, file)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

func osReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
