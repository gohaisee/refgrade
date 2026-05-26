package check_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/detect"
	"github.com/gohaisee/refgrade/internal/i18n"
	"github.com/gohaisee/refgrade/internal/project"
)

func fixturePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "fixtures", name)
}

func runFixtureREST(t *testing.T, fixture, checkID string, minCount int) {
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

func TestIntegration_badFatHandler_rest01Warn(t *testing.T) {
	runFixtureREST(t, "bad-fat-handler", "rest-01", 1)
}

func TestIntegration_badHandlerSQL_rest02Fail(t *testing.T) {
	runFixtureREST(t, "bad-handler-sql", "rest-02", 1)
}

func TestIntegration_badServerTimeouts_rest03Warn(t *testing.T) {
	runFixtureREST(t, "bad-server-timeouts", "rest-03", 1)
}

func TestIntegration_badNoBodyLimit_rest04Warn(t *testing.T) {
	runFixtureREST(t, "bad-no-body-limit", "rest-04", 1)
}

func TestIntegration_badCorsWildcard_rest05Fail(t *testing.T) {
	runFixtureREST(t, "bad-cors-wildcard", "rest-05", 1)
}

func TestIntegration_badHeaderAuth_rest06Fail(t *testing.T) {
	runFixtureREST(t, "bad-header-auth", "rest-06", 1)
}

func TestIntegration_badErrorLeak_rest07Warn(t *testing.T) {
	runFixtureREST(t, "bad-error-leak", "rest-07", 1)
}

func TestIntegration_badGinDebug_rest08Warn(t *testing.T) {
	runFixtureREST(t, "bad-gin-debug", "rest-08", 1)
}

func TestIntegration_badNoRecover_rest10Warn(t *testing.T) {
	runFixtureREST(t, "bad-no-recover", "rest-10", 1)
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

	findings, _, err := check.RunAll(context.Background(), mod, nil, check.ScanOptions{}, b.T)
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

	findings, _, err := check.RunAll(context.Background(), mod, nil, check.ScanOptions{}, b.T)
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

	findings, _, err := check.RunAll(context.Background(), mod, nil, check.ScanOptions{}, b.T)
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

	findings, _, err := check.RunAll(context.Background(), mod, nil, check.ScanOptions{}, b.T)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range findings {
		if f.ID == "cfg-01" && f.Severity == check.SeverityFail {
			t.Fatalf("unexpected cfg-01 fail: %+v", f)
		}
	}
}

func runFixtureSQL(t *testing.T, fixture, checkID string, minCount int) {
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
	if len(stacks) == 0 {
		t.Fatalf("expected sql stack in %s", fixture)
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

func TestIntegration_badSqlLoop_sql01Warn(t *testing.T) {
	runFixtureSQL(t, "bad-sql-loop", "sql-01", 1)
}

func TestIntegration_badSqlSprintf_sql02Fail(t *testing.T) {
	runFixtureSQL(t, "bad-sql-sprintf", "sql-02", 1)
}

func TestIntegration_badPoolPerRequest_sql05Fail(t *testing.T) {
	runFixtureSQL(t, "bad-pool-per-request", "sql-05", 1)
}

func TestIntegration_badPgxConnect_pgx01Warn(t *testing.T) {
	runFixtureSQL(t, "bad-pgx-connect", "pgx-01", 1)
}

func TestIntegration_badSSLModeDisable_pgx03Fail(t *testing.T) {
	runFixtureSQL(t, "bad-sslmode-disable", "pgx-03", 1)
}

func TestIntegration_badGormRaw_gorm01Fail(t *testing.T) {
	runFixtureSQL(t, "bad-gorm-raw", "gorm-01", 1)
}

func TestIntegration_badAutomigrateMain_gorm03Warn(t *testing.T) {
	runFixtureSQL(t, "bad-automigrate-main", "gorm-03", 1)
}

func TestIntegration_badGormDebug_gorm04Warn(t *testing.T) {
	runFixtureSQL(t, "bad-gorm-debug", "gorm-04", 1)
}

func TestIntegration_badGormGlobal_gorm05Warn(t *testing.T) {
	runFixtureSQL(t, "bad-gorm-global", "gorm-05", 1)
}

func TestIntegration_badSqlxNoCtx_sqlx01Warn(t *testing.T) {
	runFixtureSQL(t, "bad-sqlx-no-ctx", "sqlx-01", 1)
}

func TestIntegration_badSqlc01_sqlc01Warn(t *testing.T) {
	runFixtureSQL(t, "bad-sqlc-01", "sqlc-01", 1)
}

func TestIntegration_badEntLoop_ent01Warn(t *testing.T) {
	runFixtureSQL(t, "bad-ent-loop", "ent-01", 1)
}
