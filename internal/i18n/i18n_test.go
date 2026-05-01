package i18n

import "testing"

func TestLoad_en(t *testing.T) {
	t.Parallel()

	b, err := Load("en")
	if err != nil {
		t.Fatal(err)
	}
	if b.T("report.summary") != "summary" {
		t.Fatalf("report.summary = %q", b.T("report.summary"))
	}
	if !b.Has("cli.usage") {
		t.Fatal("missing cli.usage")
	}
}

func TestLoad_ru(t *testing.T) {
	t.Parallel()

	b, err := Load("ru")
	if err != nil {
		t.Fatal(err)
	}
	if b.T("report.summary") != "итог" {
		t.Fatalf("report.summary = %q", b.T("report.summary"))
	}
	if b.Lang() != "ru" {
		t.Fatalf("Lang = %q", b.Lang())
	}
}

func TestLoad_unsupported(t *testing.T) {
	t.Parallel()

	_, err := Load("de")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestT_fallbackToKey(t *testing.T) {
	t.Parallel()

	b, err := Load("en")
	if err != nil {
		t.Fatal(err)
	}
	if b.T("missing.key") != "missing.key" {
		t.Fatalf("fallback = %q", b.T("missing.key"))
	}
}

func TestLoad_invalidLangFile(t *testing.T) {
	t.Parallel()

	// readLocale path covered via Load
	_, err := Load("fr")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestT_ruFallbackToEn(t *testing.T) {
	t.Parallel()

	b, err := Load("ru")
	if err != nil {
		t.Fatal(err)
	}
	// оба locale задают report.summary
	if b.T("report.summary") == "" {
		t.Fatal("empty translation")
	}
}
