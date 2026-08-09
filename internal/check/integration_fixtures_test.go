package check_test

import (
	"context"
	"testing"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/i18n"
	"github.com/gohaisee/refgrade/internal/project"
)

func TestIntegration_fixtures(t *testing.T) {
	cases := []struct {
		fixture string
		checkID string
		min     int
	}{
		{"bad-layers", "lyr-01", 1},
		{"bad-layers", "lyr-02", 1},
		{"bad-hardcoded-secret", "cfg-03", 1},
		{"bad-zero-timeout-client", "err-04", 1},
		{"bad-panic-internal", "err-02", 1},
		{"bad-fmt-print", "obs-01", 1},
		{"bad-missing-tests", "tst-01", 1},
		{"bad-dead-code", "dead-01", 1},
		{"bad-dead-code", "dead-02", 1},
		{"bad-dead-code", "dead-03", 1},
		{"bad-dead-code", "dead-07", 1},
		{"bad-dead-code", "dead-08", 1},
		{"bad-orphan-go", "dead-04", 1},
		{"bad-empty-pkg", "dead-05", 1},
		{"bad-unused-mod", "dead-06", 1},
	}
	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.fixture+"/"+tc.checkID, func(t *testing.T) {
			root := fixturePath(t, tc.fixture)
			mod, err := project.Load(context.Background(), root)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			findings, _, err := check.RunAll(context.Background(), mod, nil, check.ScanOptions{}, b.T)
			if err != nil {
				t.Fatal(err)
			}
			if check.CountByID(findings, tc.checkID) < tc.min {
				t.Fatalf("expected %s findings, got %+v", tc.checkID, findings)
			}
		})
	}
}
