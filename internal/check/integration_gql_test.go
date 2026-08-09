package check_test

import (
	"context"
	"testing"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/i18n"
	"github.com/gohaisee/refgrade/internal/project"
)

func gqlScanOpts() check.ScanOptions {
	return check.ScanOptions{Stacks: []string{"gqlgen", "graphql-go", "graphql"}}
}

func TestIntegration_gqlFixtures(t *testing.T) {
	cases := []struct {
		fixture string
		checkID string
		min     int
	}{
		{"bad-gql-resolver", "gql-01", 1},
		{"bad-gql-n1", "gql-02", 1},
		{"bad-gql-no-limits", "gql-04", 1},
		{"bad-gql-introspection", "gql-05", 1},
		{"bad-gql-playground", "gql-06", 1},
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
			findings, _, err := check.RunAll(context.Background(), mod, nil, gqlScanOpts(), b.T)
			if err != nil {
				t.Fatal(err)
			}
			if check.CountByID(findings, tc.checkID) < tc.min {
				t.Fatalf("expected %s findings, got %+v", tc.checkID, findings)
			}
		})
	}
}
