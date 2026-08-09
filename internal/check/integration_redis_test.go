package check_test

import (
	"context"
	"testing"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/detect"
	"github.com/gohaisee/refgrade/internal/i18n"
	"github.com/gohaisee/refgrade/internal/project"
)

func runFixtureRedis(t *testing.T, fixture, checkID string, minCount int) {
	t.Helper()
	root := fixturePath(t, fixture)
	mod, err := project.Load(context.Background(), root, project.LoadOptions{})
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
		if s.Name == "go-redis" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected go-redis stack in %s, got %v", fixture, stackNames)
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

func TestIntegration_badRedisPerRequest_redis01Fail(t *testing.T) {
	runFixtureRedis(t, "bad-redis-per-request", "redis-01", 1)
}

func TestIntegration_badRedisNoTTL_redis02Warn(t *testing.T) {
	runFixtureRedis(t, "bad-redis-no-ttl", "redis-02", 1)
}

func TestIntegration_badRedisKeys_redis03Fail(t *testing.T) {
	runFixtureRedis(t, "bad-redis-keys", "redis-03", 1)
}

func TestIntegration_badRedisAuthCache_redis05Fail(t *testing.T) {
	runFixtureRedis(t, "bad-redis-auth-cache", "redis-05", 1)
}
