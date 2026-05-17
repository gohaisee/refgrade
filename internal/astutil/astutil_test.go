package astutil

import (
	"go/ast"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, rel, content string) string {
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

func TestNewPool_parsesAliasImport(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rel := "internal/svc/svc.go"
	abs := writeFile(t, dir, rel, `package svc

import o "os"

func X() string {
	return o.Getenv("K")
}
`)
	pool, err := NewPool([]SourceFile{{AbsPath: abs, RelPath: rel}}, nil, DefaultFilter())
	if err != nil {
		t.Fatal(err)
	}
	if len(pool.Files) != 1 {
		t.Fatalf("files = %d", len(pool.Files))
	}
	f := pool.Files[0]
	if f.ImportPathOf("o") != "os" {
		t.Fatalf("import = %q", f.ImportPathOf("o"))
	}
}

func TestHasIntegrationOrE2EBuildTag(t *testing.T) {
	t.Parallel()
	if !HasIntegrationOrE2EBuildTag("//go:build integration\npackage p\n") {
		t.Fatal("expected integration")
	}
	if HasIntegrationOrE2EBuildTag("package p\n") {
		t.Fatal("expected false")
	}
}

func TestPoolInspect(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rel := "x.go"
	abs := writeFile(t, dir, rel, "package p\nfunc F() {}\n")
	pool, err := NewPool([]SourceFile{{AbsPath: abs, RelPath: rel}}, nil, DefaultFilter())
	if err != nil {
		t.Fatal(err)
	}
	var calls int
	pool.Inspect(func(f *File, n ast.Node) bool {
		if _, ok := n.(*ast.FuncDecl); ok {
			calls++
		}
		return true
	})
	if calls != 1 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestExclude(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	abs := writeFile(t, dir, "skip.go", "package p\n")
	pool, err := NewPool([]SourceFile{{AbsPath: abs, RelPath: "skip.go"}}, func(rel string) bool {
		return rel == "skip.go"
	}, DefaultFilter())
	if err != nil {
		t.Fatal(err)
	}
	if len(pool.Files) != 0 {
		t.Fatalf("expected exclude, got %d", len(pool.Files))
	}
}
