package check_test

import (
	"context"
	"testing"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/i18n"
	"github.com/gohaisee/refgrade/internal/project"
)

func grpcScanOpts() check.ScanOptions {
	return check.ScanOptions{Stacks: []string{"grpc", "connect", "grpc-gateway"}}
}

func TestIntegration_grpcFixtures(t *testing.T) {
	cases := []struct {
		fixture string
		checkID string
		min     int
	}{
		{"bad-grpc-insecure", "grpc-01", 1},
		{"bad-grpc-no-interceptor", "grpc-02", 1},
		{"bad-grpc-metadata-auth", "grpc-04", 1},
		{"bad-grpc-reflection", "grpc-05", 1},
		{"bad-connect-timeout", "conn-01", 1},
	}
	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.fixture+"/"+tc.checkID, func(t *testing.T) {
			root := fixturePath(t, tc.fixture)
			mod, err := project.Load(context.Background(), root, project.LoadOptions{})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			findings, _, err := check.RunAll(context.Background(), mod, nil, grpcScanOpts(), b.T)
			if err != nil {
				t.Fatal(err)
			}
			if check.CountByID(findings, tc.checkID) < tc.min {
				t.Fatalf("expected %s findings, got %+v", tc.checkID, findings)
			}
		})
	}
}
