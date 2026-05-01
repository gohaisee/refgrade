package engine_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/engine"
	"github.com/gohaisee/refgrade/internal/i18n"
	"github.com/gohaisee/refgrade/internal/project"
)

func TestScan_badGetenv(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "bad-getenv")
	mod, err := project.Load(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	res, err := engine.Scan(context.Background(), mod, check.Catalog(), b.T)
	if err != nil {
		t.Fatal(err)
	}
	if !engine.HasFail(res) {
		t.Fatal("expected fail result")
	}
}

func TestScan_goodMinimal(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	mod, err := project.Load(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	res, err := engine.Scan(context.Background(), mod, check.Catalog(), b.T)
	if err != nil {
		t.Fatal(err)
	}
	if engine.HasFail(res) {
		t.Fatalf("unexpected fail: %+v", res.Findings)
	}
}
