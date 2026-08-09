package check

import (
	"context"
	"go/parser"
	"testing"
)

func TestErr01_flagsIgnoredErrAssign(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/service/svc.go"
	path := writeGoFile(t, dir, rel, `package service

import "os"

func Save(path string, data []byte) error {
	err := os.WriteFile(path, data, 0o644)
	_ = err
	return nil
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewErr01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != err01ID || findings[0].Severity != SeverityWarn {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestErr01_flagsEmptyErrNilIf(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/service/svc.go"
	path := writeGoFile(t, dir, rel, `package service

import "os"

func Load(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
	}
	return data, nil
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewErr01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestErr01_flagsEmptyErrNilIfWithInit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/service/svc.go"
	path := writeGoFile(t, dir, rel, `package service

import "os"

func Open(path string) (*os.File, error) {
	if f, err := os.Open(path); err != nil {
	}
	return nil, nil
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewErr01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestErr01_allowsHandledErr(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/service/svc.go"
	path := writeGoFile(t, dir, rel, `package service

import "os"

func Load(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewErr01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestErr01_allowsOtherBlankAssign(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/service/svc.go"
	path := writeGoFile(t, dir, rel, `package service

func Drop(x int) {
	_ = x
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewErr01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestErr01_skipsTestFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/service/svc_test.go"
	path := writeGoFile(t, dir, rel, `package service

func TestX(t *testing.T) {
	var err error
	_ = err
	if err != nil {
	}
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewErr01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestErr01_ID(t *testing.T) {
	t.Parallel()

	if NewErr01().ID() != err01ID {
		t.Fatal("unexpected id")
	}
}

func TestIsErrNotNilExpr(t *testing.T) {
	t.Parallel()

	cases := []struct {
		src  string
		want bool
	}{
		{"err != nil", true},
		{"nil != err", true},
		{"someErr != nil", false},
		{"err == nil", false},
	}
	for _, tc := range cases {
		file, err := parser.ParseExpr(tc.src)
		if err != nil {
			t.Fatalf("parse %q: %v", tc.src, err)
		}
		if got := isErrNotNilExpr(file); got != tc.want {
			t.Fatalf("isErrNotNilExpr(%q) = %v, want %v", tc.src, got, tc.want)
		}
	}
}
