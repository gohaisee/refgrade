package check

import (
	"os"
	"path/filepath"

	"github.com/gohaisee/refgrade/internal/astutil"
	"github.com/gohaisee/refgrade/internal/refgradeconfig"
)

// minimal ModuleView for unit tests
type stubModule struct {
	root  string
	files []GoFile
}

func (s stubModule) Root() string { return s.root }

func (s stubModule) RelPath(file string) (string, error) {
	rel, err := filepath.Rel(s.root, file)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

func (s stubModule) GoSourceFiles() []GoFile { return s.files }

func (s stubModule) Packages() []PackageInfo { return nil }

func (s stubModule) GoModContent() []byte { return nil }

func (s stubModule) Excluded(string) bool { return false }

func (s stubModule) Config() *refgradeconfig.Config {
	return &refgradeconfig.Config{Checks: make(map[string]refgradeconfig.CheckSetting)}
}

func (s stubModule) ASTPool(filter astutil.Filter) (*astutil.Pool, error) {
	var src []astutil.SourceFile
	for _, f := range s.files {
		src = append(src, astutil.SourceFile{AbsPath: f.Path, RelPath: f.RelPath})
	}
	return astutil.NewPool(src, nil, filter)
}

func writeGoFile(t testingT, dir, rel, content string) string {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

type testingT interface {
	Helper()
	Fatal(...any)
}
