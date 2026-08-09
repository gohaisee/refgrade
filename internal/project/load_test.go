package project

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestModuleRoot_goodMinimal(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	got, err := ModuleRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(root)
	if got != abs {
		t.Fatalf("ModuleRoot = %q want %q", got, abs)
	}
}

func TestLoad_goodMinimal(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	mod, err := Load(context.Background(), root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if mod.ModPath != "github.com/gohaisee/refgrade/fixtures/good-minimal" {
		t.Fatalf("ModPath = %q", mod.ModPath)
	}
	if len(mod.Packages) == 0 {
		t.Fatal("expected packages")
	}
}

func TestRelPath(t *testing.T) {
	t.Parallel()

	mod := &Module{Root: "/repo"}
	rel, err := mod.RelPath(filepath.Join("/repo", "internal", "svc", "foo.go"))
	if err != nil {
		t.Fatal(err)
	}
	if rel != "internal/svc/foo.go" {
		t.Fatalf("RelPath = %q", rel)
	}
}

func TestFindModuleRoot_missing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, _, err := findModuleRoot(dir)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseModulePath(t *testing.T) {
	t.Parallel()

	got := parseModulePath([]byte("module example.com/foo\n\ngo 1.22\n"))
	if got != "example.com/foo" {
		t.Fatalf("parseModulePath = %q", got)
	}
}

func TestLoad_badGetenv(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "testdata", "fixtures", "bad-getenv")
	mod, err := Load(context.Background(), root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if mod.ModPath != "github.com/gohaisee/refgrade/fixtures/bad-getenv" {
		t.Fatalf("ModPath = %q", mod.ModPath)
	}
}

func TestListPackages_goListFailure(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// invalid module path — go list must fail
	content := "module\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := listPackages(context.Background(), dir)
	if err == nil {
		t.Fatal("expected go list error for invalid go.mod")
	}
}
