package engine_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gohaisee/refgrade/internal/engine"
	"github.com/gohaisee/refgrade/internal/i18n"
	"github.com/gohaisee/refgrade/internal/project"
	"github.com/gohaisee/refgrade/internal/refgradeconfig"
)

func TestScan_badGetenv(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "bad-getenv")
	mod, err := project.Load(context.Background(), root, project.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	res, err := engine.Scan(context.Background(), mod, engine.Options{
		Config: &refgradeconfig.Config{Checks: make(map[string]refgradeconfig.CheckSetting)},
	}, b.T)
	if err != nil {
		t.Fatal(err)
	}
	if !engine.HasFail(res) {
		t.Fatal("expected fail result")
	}
}

func TestScan_goodMinimal(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	mod, err := project.Load(context.Background(), root, project.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	res, err := engine.Scan(context.Background(), mod, engine.Options{
		Config: &refgradeconfig.Config{Checks: make(map[string]refgradeconfig.CheckSetting)},
	}, b.T)
	if err != nil {
		t.Fatal(err)
	}
	if engine.HasFail(res) {
		t.Fatalf("unexpected fail: %+v", res.Findings)
	}
}

func TestScan_naStatuses(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	mod, err := project.Load(context.Background(), root, project.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	res, err := engine.Scan(context.Background(), mod, engine.Options{
		Config: &refgradeconfig.Config{Checks: make(map[string]refgradeconfig.CheckSetting)},
		Stacks: nil,
	}, b.T)
	if err != nil {
		t.Fatal(err)
	}
	var hasNA bool
	for _, st := range res.Statuses {
		if st.Severity == "n/a" {
			hasNA = true
		}
	}
	if !hasNA {
		t.Fatal("expected n/a statuses for gated checks")
	}
}
