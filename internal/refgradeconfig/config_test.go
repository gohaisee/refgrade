package refgradeconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeYAML(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".refgrade.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_missing(t *testing.T) {
	t.Parallel()

	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Lang != "" {
		t.Fatalf("Lang = %q", cfg.Lang)
	}
}

func TestLoad_lang(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeYAML(t, dir, "lang: ru\n")
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Lang != "ru" {
		t.Fatalf("Lang = %q", cfg.Lang)
	}
}

func TestResolveLang_precedence(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "lang: ru\n")

	got, err := ResolveLang("en", dir)
	if err != nil || got != "en" {
		t.Fatalf("flag: got %q %v", got, err)
	}

	t.Setenv("REFGRADE_LANG", "en")
	got, err = ResolveLang("", dir)
	if err != nil || got != "en" {
		t.Fatalf("env over yaml: got %q %v", got, err)
	}

	t.Setenv("REFGRADE_LANG", "")
	got, err = ResolveLang("", dir)
	if err != nil || got != "ru" {
		t.Fatalf("yaml: got %q %v", got, err)
	}

	noYAML := t.TempDir()
	got, err = ResolveLang("", noYAML)
	if err != nil || got != "en" {
		t.Fatalf("default: got %q %v", got, err)
	}
}

func TestResolveLang_unsupported(t *testing.T) {
	t.Parallel()

	_, err := ResolveLang("de", t.TempDir())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_profileAndChecks(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeYAML(t, dir, `lang: en
profile: strict
checks.cfg-01: off
exclude.0: "**/gen/**"
layers.0.path: internal/service
layers.0.forbid.0: github.com/jackc/pgx/v5
`)
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Profile != ProfileStrict {
		t.Fatalf("profile = %q", cfg.Profile)
	}
	if cfg.CheckEnabled("cfg-01") {
		t.Fatal("cfg-01 should be off")
	}
	if len(cfg.Exclude) != 1 {
		t.Fatalf("exclude = %v", cfg.Exclude)
	}
	if len(cfg.Layers) != 1 || cfg.Layers[0].Path != "internal/service" {
		t.Fatalf("layers = %v", cfg.Layers)
	}
}

func TestExcludeMatcher(t *testing.T) {
	t.Parallel()

	m := NewExcludeMatcher([]string{"**/generated/**", "*_gen.go"})
	if !m.Excluded("internal/generated/foo.go") {
		t.Fatal("expected **/generated/** match")
	}
	if !m.Excluded("internal/foo_gen.go") {
		t.Fatal("expected *_gen.go match")
	}
}

func TestApplyProfile(t *testing.T) {
	t.Parallel()

	if ApplyProfile(ProfileStrict, "x", SeverityWarn) != SeverityFail {
		t.Fatal("strict should upgrade warn")
	}
	if ApplyProfile(ProfileMinimal, "x", SeverityWarn) != SeverityInfo {
		t.Fatal("minimal should downgrade warn")
	}
}

func TestInitTemplate(t *testing.T) {
	t.Parallel()

	if !strings.Contains(InitTemplate(), "lang: en") {
		t.Fatal("expected template")
	}
}

func TestLoad_invalidYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeYAML(t, dir, "not yaml line\n")
	if _, err := Load(dir); err == nil {
		t.Fatal("expected parse error")
	}
}
