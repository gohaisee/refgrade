package check_test

import (
	"context"
	"testing"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/detect"
	"github.com/gohaisee/refgrade/internal/i18n"
	"github.com/gohaisee/refgrade/internal/project"
)

func runFixtureMongo(t *testing.T, fixture, checkID string, minCount int) {
	t.Helper()
	root := fixturePath(t, fixture)
	mod, err := project.Load(context.Background(), root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	stacks, err := detect.Detect(context.Background(), mod)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	stackNames := make([]string, len(stacks))
	found := false
	for i, s := range stacks {
		stackNames[i] = s.Name
		if s.Name == "mongo-driver" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected mongo-driver stack in %s, got %v", fixture, stackNames)
	}
	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	findings, _, err := check.RunAll(context.Background(), mod, nil, check.ScanOptions{Stacks: stackNames}, b.T)
	if err != nil {
		t.Fatal(err)
	}
	if check.CountByID(findings, checkID) < minCount {
		t.Fatalf("expected at least %d %s in %s, got %+v", minCount, checkID, fixture, findings)
	}
}

func TestIntegration_badMongoPerRequest_mongo01Fail(t *testing.T) {
	runFixtureMongo(t, "bad-mongo-per-request", "mongo-01", 1)
}

func TestIntegration_badMongoBgCtx_mongo02Warn(t *testing.T) {
	runFixtureMongo(t, "bad-mongo-bg-ctx", "mongo-02", 1)
}

func TestIntegration_badMongoNoLimit_mongo03Warn(t *testing.T) {
	runFixtureMongo(t, "bad-mongo-no-limit", "mongo-03", 1)
}

func TestIntegration_badMongoWhere_mongo06Fail(t *testing.T) {
	runFixtureMongo(t, "bad-mongo-where", "mongo-06", 1)
}

func TestIntegration_badMongoHandler_mongo08Warn(t *testing.T) {
	runFixtureMongo(t, "bad-mongo-handler", "mongo-08", 1)
}
