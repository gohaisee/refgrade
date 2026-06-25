package check_test

import (
	"context"
	"testing"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/detect"
	"github.com/gohaisee/refgrade/internal/i18n"
	"github.com/gohaisee/refgrade/internal/project"
)

func runFixtureSecurity(t *testing.T, fixture, checkID string, minCount int) {
	t.Helper()
	root := fixturePath(t, fixture)
	mod, err := project.Load(context.Background(), root, project.LoadOptions{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	findings, _, err := check.RunAll(context.Background(), mod, nil, check.ScanOptions{}, b.T)
	if err != nil {
		t.Fatal(err)
	}
	if check.CountByID(findings, checkID) < minCount {
		t.Fatalf("expected at least %d %s in %s, got %+v", minCount, checkID, fixture, findings)
	}
}

func runFixtureSecurityREST(t *testing.T, fixture, checkID string, minCount int) {
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
	if len(stacks) == 0 {
		t.Fatalf("expected http stack in %s", fixture)
	}
	stackNames := make([]string, len(stacks))
	for i, s := range stacks {
		stackNames[i] = s.Name
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

func TestIntegration_badWeakCrypto_sec01(t *testing.T) {
	runFixtureSecurity(t, "bad-weak-crypto", "sec-01", 1)
}

func TestIntegration_badInsecureTLS_sec02(t *testing.T) {
	runFixtureSecurity(t, "bad-insecure-tls", "sec-02", 1)
}

func TestIntegration_badMathRandToken_sec03(t *testing.T) {
	runFixtureSecurity(t, "bad-math-rand-token", "sec-03", 1)
}

func TestIntegration_badJWTNone_sec05(t *testing.T) {
	runFixtureSecurity(t, "bad-jwt-none", "sec-05", 1)
}

func TestIntegration_badDotenvTracked_sec10(t *testing.T) {
	runFixtureSecurity(t, "bad-dotenv-tracked", "sec-10", 1)
}

func TestIntegration_badPprofImport_sec15(t *testing.T) {
	runFixtureSecurity(t, "bad-pprof-import", "sec-15", 1)
}

func TestIntegration_badMassAssign_secR03(t *testing.T) {
	runFixtureSecurityREST(t, "bad-mass-assign", "sec-r03", 1)
}

func TestIntegration_badAdminNoRBAC_secR05(t *testing.T) {
	runFixtureSecurityREST(t, "bad-admin-no-rbac", "sec-r05", 1)
}

func TestIntegration_badWebhookNoHMAC_secR10(t *testing.T) {
	runFixtureSecurityREST(t, "bad-webhook-no-hmac", "sec-r10", 1)
}

func TestIntegration_badDynamicTable_secDB05(t *testing.T) {
	runFixtureSQL(t, "bad-dynamic-table", "sec-db05", 1)
}

func TestIntegration_badMongoURIPassword_secM01(t *testing.T) {
	runFixtureMongo(t, "bad-mongo-uri-password", "sec-m01", 1)
}

func TestIntegration_badMongoNoTLSRemote_secM02(t *testing.T) {
	runFixtureMongo(t, "bad-mongo-no-tls-remote", "sec-m02", 1)
}

func TestIntegration_badRedisURIPassword_secRD01(t *testing.T) {
	runFixtureRedis(t, "bad-redis-uri-password", "sec-rd01", 1)
}

func TestIntegration_badRedisNoTLSRemote_secRD02(t *testing.T) {
	runFixtureRedis(t, "bad-redis-no-tls-remote", "sec-rd02", 1)
}

func TestIntegration_badMQCredentials_secMQ01(t *testing.T) {
	runFixtureMQ(t, "bad-mq-credentials", "sec-mq01", 1)
}

func TestIntegration_badMQPlaintextRemote_secMQ02(t *testing.T) {
	runFixtureMQ(t, "bad-mq-plaintext-remote", "sec-mq02", 1)
}

func TestIntegration_sec16_warnWithoutWithSecurity(t *testing.T) {
	runFixtureSecurity(t, "good-minimal", "sec-16", 1)
}
