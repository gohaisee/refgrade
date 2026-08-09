package check_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/i18n"
	"github.com/gohaisee/refgrade/internal/project"
)

func fixturePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "fixtures", name)
}

func TestIntegration_badIgnoredError_err01Warn(t *testing.T) {
	root := fixturePath(t, "bad-ignored-error")
	mod, err := project.Load(context.Background(), root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}

	findings, err := check.RunAll(context.Background(), mod, check.Catalog(), b.T)
	if err != nil {
		t.Fatal(err)
	}
	if check.CountByID(findings, "err-01") < 2 {
		t.Fatalf("expected at least 2 err-01 findings, got %+v", findings)
	}
	if !check.HasSeverity(findings, check.SeverityWarn) {
		t.Fatalf("expected warn severity, got %+v", findings)
	}
	if check.HasSeverity(findings, check.SeverityFail) {
		t.Fatalf("unexpected fail severity, got %+v", findings)
	}

	var inService int
	for _, f := range findings {
		if f.ID == "err-01" && f.File == "internal/service/service.go" {
			inService++
		}
	}
	if inService < 2 {
		t.Fatalf("expected err-01 in internal/service/service.go, got %+v", findings)
	}
}

func TestIntegration_badDefaultClient_err03Fail(t *testing.T) {
	root := fixturePath(t, "bad-default-client")
	mod, err := project.Load(context.Background(), root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}

	findings, err := check.RunAll(context.Background(), mod, check.Catalog(), b.T)
	if err != nil {
		t.Fatal(err)
	}
	if check.CountByID(findings, "err-03") == 0 {
		t.Fatalf("expected err-03 findings, got %+v", findings)
	}
}

func TestIntegration_badGetenv_cfg01Fail(t *testing.T) {
	root := fixturePath(t, "bad-getenv")
	mod, err := project.Load(context.Background(), root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}

	findings, err := check.RunAll(context.Background(), mod, check.Catalog(), b.T)
	if err != nil {
		t.Fatal(err)
	}
	if check.CountByID(findings, "cfg-01") == 0 {
		t.Fatalf("expected cfg-01 findings, got %+v", findings)
	}
	if !check.HasSeverity(findings, check.SeverityFail) {
		t.Fatalf("expected fail severity, got %+v", findings)
	}

	var inService bool
	for _, f := range findings {
		if f.ID == "cfg-01" && f.File == "internal/service/service.go" {
			inService = true
		}
	}
	if !inService {
		t.Fatalf("expected finding in internal/service/service.go, got %+v", findings)
	}
}

func TestIntegration_goodMinimal_noCfg01Fail(t *testing.T) {
	root := fixturePath(t, "good-minimal")
	mod, err := project.Load(context.Background(), root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}

	findings, err := check.RunAll(context.Background(), mod, check.Catalog(), b.T)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range findings {
		if f.ID == "cfg-01" && f.Severity == check.SeverityFail {
			t.Fatalf("unexpected cfg-01 fail: %+v", f)
		}
	}
}
