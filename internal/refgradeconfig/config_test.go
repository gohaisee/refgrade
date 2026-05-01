package refgradeconfig

import (
	"os"
	"path/filepath"
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

func TestLoad_invalidYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeYAML(t, dir, "not yaml line\n")
	if _, err := Load(dir); err == nil {
		t.Fatal("expected parse error")
	}
}
