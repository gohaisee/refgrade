package check

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gohaisee/refgrade/internal/astutil"
	"github.com/gohaisee/refgrade/internal/project"
)

func TestCfg01_flagsInternalGetenv(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/service/svc.go"
	path := writeGoFile(t, dir, rel, `package service

import "os"

func Load() string {
	return os.Getenv("KEY")
}
`)
	mod := stubModule{
		root:  dir,
		files: []GoFile{{Path: path, RelPath: rel}},
	}

	findings, err := NewCfg01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
	if findings[0].ID != "cfg-01" || findings[0].Severity != SeverityFail {
		t.Fatalf("finding = %+v", findings[0])
	}
}

func TestCfg01_allowsCmd(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import "os"

func main() {
	_ = os.Getenv("PORT")
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewCfg01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestCfg01_aliasImport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/service/svc.go"
	path := writeGoFile(t, dir, rel, `package service

import o "os"

func Load() string {
	return o.Getenv("KEY")
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewCfg01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestCfg01_skipsTestFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/service/svc_test.go"
	path := writeGoFile(t, dir, rel, `package service

import "os"

func TestX(t *testing.T) {
	_ = os.Getenv("X")
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewCfg01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestCfg01_skipsIntegrationBuildTag(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/integration/hook.go"
	path := writeGoFile(t, dir, rel, `//go:build integration

package integration

import "os"

func Hook() string {
	return os.Getenv("X")
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewCfg01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestHasIntegration_legacyBuildTag(t *testing.T) {
	t.Parallel()

	src := "// +build e2e\npackage p\n"
	if !astutil.HasIntegrationOrE2EBuildTag(src) {
		t.Fatal("expected e2e tag match")
	}
}

func TestSetMeta(t *testing.T) {
	t.Parallel()

	findings := SetMeta([]Finding{{ID: "cfg-01", Severity: SeverityFail}}, func(k string) string {
		return "t:" + k
	})
	if !strings.HasPrefix(findings[0].When, "t:") {
		t.Fatalf("When = %q", findings[0].When)
	}
}

func TestFromProject_adapter(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	mod, err := project.Load(context.Background(), root, project.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	view := FromProject(mod, nil)
	if view.Root() != mod.Root {
		t.Fatal("root mismatch")
	}
	files := view.GoSourceFiles()
	if len(files) == 0 {
		t.Fatal("expected files")
	}
}

func TestHasSeverityAndCount(t *testing.T) {
	t.Parallel()

	fs := []Finding{{ID: "cfg-01", Severity: SeverityFail}, {ID: "cfg-01", Severity: SeverityFail}}
	if !HasSeverity(fs, SeverityFail) {
		t.Fatal("expected fail")
	}
	if CountByID(fs, "cfg-01") != 2 {
		t.Fatal("count mismatch")
	}
}

func TestErrUnsupported(t *testing.T) {
	t.Parallel()

	err := ErrUnsupported{Check: "x", Why: "y"}
	if err.Error() == "" {
		t.Fatal("expected message")
	}
}

func TestGatesMatch(t *testing.T) {
	t.Parallel()

	stacks := map[string]struct{}{"gin": {}}
	if !gatesMatch([]string{"gin", "echo"}, stacks) {
		t.Fatal("expected match")
	}
	if gatesMatch([]string{"pgx"}, stacks) {
		t.Fatal("expected no match")
	}
}
