package check

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gohaisee/refgrade/internal/astutil"
)

func TestDead04_flagsOrphanGoFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	root := writeGoFile(t, dir, "internal/foo/foo.go", `package foo

func Live() int { return 1 }
`)
	writeGoFile(t, dir, "internal/foo/orphan.go", `//go:build never

package foo

func Orphan() int { return 2 }
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: root, RelPath: "internal/foo/foo.go"}}}
	findings, err := NewDead04().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "dead-04" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestDead05_flagsEmptyPackage(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/empty/empty.go", `package empty
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/empty/empty.go"}}}
	findings, err := NewDead05().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "dead-05" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestDead07_flagsLargeCommentBlock(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/svc/legacy.go", `package svc

// func old() {
// 	if x := 1; x > 0 {
// 		return
// 	}
// 	for i := 0; i < 3; i++ {
// 		_ = i
// 	}
// }
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/svc/legacy.go"}}}
	findings, err := NewDead07().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "dead-07" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestParseDeadcodeLine(t *testing.T) {
	t.Parallel()
	file, line, name := parseDeadcodeLine("internal/svc/svc.go:7:6: unreachable func: DeadExport")
	if file != "internal/svc/svc.go" || line != 7 || name != "DeadExport" {
		t.Fatalf("got %s %d %s", file, line, name)
	}
}

func TestHelperUsedInFileCount(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := writeGoFile(t, dir, "pkg/a_test.go", `package pkg

import "testing"

func HelperExport() {}

func TestA(t *testing.T) { HelperExport() }
`)
	b := writeGoFile(t, dir, "pkg/b_test.go", `package pkg

import "testing"

func TestB(t *testing.T) {}
`)
	pool, err := astutil.NewPool([]astutil.SourceFile{
		{AbsPath: a, RelPath: filepath.ToSlash("pkg/a_test.go")},
		{AbsPath: b, RelPath: filepath.ToSlash("pkg/b_test.go")},
	}, nil, astutil.DefaultFilter())
	if err != nil {
		t.Fatal(err)
	}
	if helperUsedInFileCount(pool.Files, "HelperExport") != 1 {
		t.Fatal("expected helper in one file only")
	}
}
