package check

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gohaisee/refgrade/internal/project"
)

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

func writeGoFile(t *testing.T, dir, rel, content string) string {
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
	if findings[0].ID != cfg01ID || findings[0].Severity != SeverityFail {
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

func TestHasIntegrationE2EBuildTag(t *testing.T) {
	t.Parallel()

	cases := []struct {
		src  string
		want bool
	}{
		{"//go:build integration\npackage p\n", true},
		{"//go:build e2e\npackage p\n", true},
		{"//go:build unit\npackage p\n", false},
		{"package p\n", false},
	}
	for _, tc := range cases {
		if got := hasIntegrationE2EBuildTag(tc.src); got != tc.want {
			t.Fatalf("hasIntegrationE2EBuildTag() = %v, want %v", got, tc.want)
		}
	}
}

func TestSkipCfg01File(t *testing.T) {
	t.Parallel()

	if skipCfg01File("internal/x.go") {
		t.Fatal("internal should be scanned")
	}
	if !skipCfg01File("cmd/main.go") {
		t.Fatal("cmd should skip")
	}
	if !skipCfg01File("pkg/x_test.go") {
		t.Fatal("test should skip")
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

func TestCfg01_ID(t *testing.T) {
	t.Parallel()

	if NewCfg01().ID() != cfg01ID {
		t.Fatal("unexpected id")
	}
}

func TestErrUnsupported(t *testing.T) {
	t.Parallel()

	err := ErrUnsupported{Check: "x", Why: "y"}
	if err.Error() == "" {
		t.Fatal("expected message")
	}
}

func TestHasIntegration_legacyBuildTag(t *testing.T) {
	t.Parallel()

	src := "// +build e2e\npackage p\n"
	if !hasIntegrationE2EBuildTag(src) {
		t.Fatal("expected e2e tag match")
	}
}

func TestIsOSGetenv_negative(t *testing.T) {
	t.Parallel()

	if isOSGetenv(nil) {
		t.Fatal("nil should be false")
	}
}

func TestFromProject_adapter(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	mod, err := project.Load(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	view := FromProject(mod)
	if view.Root() != mod.Root {
		t.Fatal("root mismatch")
	}
	files := view.GoSourceFiles()
	if len(files) == 0 {
		t.Fatal("expected files")
	}
	rel, err := view.RelPath(files[0].Path)
	if err != nil || rel == "" {
		t.Fatalf("RelPath = %q %v", rel, err)
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
	if HasSeverity(fs, SeverityWarn) {
		t.Fatal("unexpected warn")
	}
}
