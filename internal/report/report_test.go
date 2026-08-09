package report

import (
	"strings"
	"testing"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/engine"
	"github.com/gohaisee/refgrade/internal/i18n"
	"github.com/gohaisee/refgrade/internal/project"
)

func TestParseFormat(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want Format
	}{
		{"text", FormatText},
		{"markdown", FormatMarkdown},
		{"md", FormatMarkdown},
		{"json", FormatJSON},
	}
	for _, tc := range cases {
		got, err := ParseFormat(tc.in)
		if err != nil {
			t.Fatalf("ParseFormat(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("ParseFormat(%q) = %q", tc.in, got)
		}
	}
}

func TestParseFormat_unknown(t *testing.T) {
	t.Parallel()

	_, err := ParseFormat("xml")
	if err == nil {
		t.Fatal("expected error")
	}
}

func sampleResult(findings []check.Finding) *engine.Result {
	return &engine.Result{
		Module: &project.Module{
			ModPath: "example.com/app",
		},
		Findings: findings,
	}
}

func TestRenderText_noFindings(t *testing.T) {
	t.Parallel()

	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	out := RenderText(sampleResult(nil), b)
	if !strings.Contains(out, "no findings") {
		t.Fatalf("output = %q", out)
	}
	if !strings.Contains(out, "example.com/app") {
		t.Fatalf("output = %q", out)
	}
}

func TestRenderText_withFinding(t *testing.T) {
	t.Parallel()

	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	findings := []check.Finding{{
		ID: "cfg-01", Severity: check.SeverityFail,
		File: "internal/svc/foo.go", Line: 10,
		When: "w", Why: "y", Fix: "f",
	}}
	out := RenderText(sampleResult(findings), b)
	if !strings.Contains(out, "cfg-01") || !strings.Contains(out, "fail") {
		t.Fatalf("output = %q", out)
	}
}

func TestRenderMarkdown(t *testing.T) {
	t.Parallel()

	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	findings := []check.Finding{{
		ID: "cfg-01", Severity: check.SeverityFail,
		File: "internal/svc/foo.go", Line: 3,
		When: "w", Why: "y", Fix: "f",
	}}
	out := RenderMarkdown(sampleResult(findings), b)
	if !strings.Contains(out, "| cfg-01 |") {
		t.Fatalf("output = %q", out)
	}
}

func TestRenderJSON(t *testing.T) {
	t.Parallel()

	findings := []check.Finding{{
		ID: "cfg-01", Severity: check.SeverityFail,
		File: "a.go", Line: 1,
	}}
	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	data, err := RenderJSON(sampleResult(findings), b)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, `"module": "example.com/app"`) {
		t.Fatalf("json = %s", s)
	}
	if !strings.Contains(s, `"summary":`) {
		t.Fatalf("json = %s", s)
	}
}

func TestRender_ru(t *testing.T) {
	t.Parallel()

	b, err := i18n.Load("ru")
	if err != nil {
		t.Fatal(err)
	}
	out := RenderText(sampleResult(nil), b)
	if !strings.Contains(out, "замечаний нет") {
		t.Fatalf("output = %q", out)
	}
}

func TestRender_formats(t *testing.T) {
	t.Parallel()

	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	res := sampleResult([]check.Finding{{
		ID: "cfg-01", Severity: check.SeverityWarn,
		File: "a.go", Line: 1, When: "w", Why: "y", Fix: "f",
	}})

	for _, format := range []Format{FormatText, FormatMarkdown, FormatJSON} {
		data, err := Render(res, format, b)
		if err != nil {
			t.Fatalf("Render %s: %v", format, err)
		}
		if len(data) == 0 {
			t.Fatalf("empty %s output", format)
		}
	}
}

func TestSummarize_warnOnly(t *testing.T) {
	t.Parallel()

	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	fs := []check.Finding{{Severity: check.SeverityWarn}}
	out := summarize(fs, nil, b)
	if !strings.Contains(out, "warn") {
		t.Fatalf("summarize = %q", out)
	}
}

func TestSeverityLabel_all(t *testing.T) {
	t.Parallel()

	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	for _, sev := range []string{check.SeverityFail, check.SeverityWarn, check.SeverityOK, check.SeverityNA, "custom"} {
		if severityLabel(b, sev) == "" {
			t.Fatalf("empty label for %q", sev)
		}
	}
}

func TestRenderSARIF(t *testing.T) {
	t.Parallel()

	b, err := i18n.Load("en")
	if err != nil {
		t.Fatal(err)
	}
	findings := []check.Finding{{
		ID: "cfg-01", Severity: check.SeverityFail,
		File: "a.go", Line: 1, Why: "test",
	}}
	data, err := RenderSARIF(sampleResult(findings), b)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "cfg-01") {
		t.Fatalf("sarif = %s", data)
	}
}
