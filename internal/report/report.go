package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/engine"
	"github.com/gohaisee/refgrade/internal/i18n"
)

// output encoding for scan results
type Format string

const (
	FormatText     Format = "text"
	FormatMarkdown Format = "markdown"
	FormatJSON     Format = "json"
)

// normalizes cli --format value
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "text", "":
		return FormatText, nil
	case "markdown", "md":
		return FormatMarkdown, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unknown format %q", s)
	}
}

// formats scan result using bundle strings
func Render(result *engine.Result, format Format, b *i18n.Bundle) ([]byte, error) {
	switch format {
	case FormatText:
		return []byte(RenderText(result, b)), nil
	case FormatMarkdown:
		return []byte(RenderMarkdown(result, b)), nil
	case FormatJSON:
		return RenderJSON(result)
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}
}

// plain text report
func RenderText(result *engine.Result, b *i18n.Bundle) string {
	var out strings.Builder
	out.WriteString(b.T("report.title"))
	out.WriteByte('\n')
	out.WriteString(b.T("report.module"))
	out.WriteString(": ")
	out.WriteString(result.Module.ModPath)
	out.WriteByte('\n')

	if len(result.Findings) == 0 {
		out.WriteString(b.T("report.no_findings"))
		out.WriteByte('\n')
		return out.String()
	}

	out.WriteString(b.T("report.findings"))
	out.WriteString(":\n")
	for _, f := range result.Findings {
		writeTextFinding(&out, b, f)
	}
	writeSummary(&out, b, result.Findings)
	return out.String()
}

func writeTextFinding(out *strings.Builder, b *i18n.Bundle, f check.Finding) {
	out.WriteString("  ")
	out.WriteString(f.ID)
	out.WriteString(" [")
	out.WriteString(severityLabel(b, f.Severity))
	out.WriteString("] ")
	out.WriteString(f.File)
	fmt.Fprintf(out, ":%d\n", f.Line)
}

func writeSummary(out *strings.Builder, b *i18n.Bundle, findings []check.Finding) {
	out.WriteString(b.T("report.summary"))
	out.WriteString(": ")
	out.WriteString(summarize(findings, b))
	out.WriteByte('\n')
}

func summarize(findings []check.Finding, b *i18n.Bundle) string {
	fail := 0
	warn := 0
	for _, f := range findings {
		switch f.Severity {
		case check.SeverityFail:
			fail++
		case check.SeverityWarn:
			warn++
		}
	}
	if fail > 0 {
		return fmt.Sprintf("%s=%d", b.T("report.fail"), fail)
	}
	if warn > 0 {
		return fmt.Sprintf("%s=%d", b.T("report.warn"), warn)
	}
	return b.T("report.ok")
}

func severityLabel(b *i18n.Bundle, sev string) string {
	switch sev {
	case check.SeverityFail:
		return b.T("report.fail")
	case check.SeverityWarn:
		return b.T("report.warn")
	case check.SeverityOK:
		return b.T("report.ok")
	case check.SeverityNA:
		return b.T("report.na")
	default:
		return sev
	}
}

// markdown report
func RenderMarkdown(result *engine.Result, b *i18n.Bundle) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# %s\n\n", b.T("report.title"))
	fmt.Fprintf(&out, "**%s:** %s\n\n", b.T("report.module"), result.Module.ModPath)

	if len(result.Findings) == 0 {
		out.WriteString(b.T("report.no_findings"))
		out.WriteByte('\n')
		return out.String()
	}

	out.WriteString("| ")
	out.WriteString(b.T("report.check"))
	out.WriteString(" | ")
	out.WriteString(b.T("report.severity"))
	out.WriteString(" | ")
	out.WriteString(b.T("report.file"))
	out.WriteString(" | ")
	out.WriteString(b.T("report.line"))
	out.WriteString(" |\n")

	out.WriteString("|---|---|---|---|\n")
	for _, f := range result.Findings {
		fmt.Fprintf(&out, "| %s | %s | %s | %d |\n",
			f.ID, severityLabel(b, f.Severity), f.File, f.Line)
	}
	out.WriteByte('\n')
	for _, f := range result.Findings {
		fmt.Fprintf(&out, "### %s\n\n", f.ID)
		fmt.Fprintf(&out, "- **%s:** %s\n", b.T("report.when"), f.When)
		fmt.Fprintf(&out, "- **%s:** %s\n", b.T("report.why"), f.Why)
		fmt.Fprintf(&out, "- **%s:** %s\n", b.T("report.fix"), f.Fix)
		fmt.Fprintf(&out, "- **%s:** %s:%d\n\n", b.T("report.file"), f.File, f.Line)
	}
	fmt.Fprintf(&out, "**%s:** %s\n", b.T("report.summary"), summarize(result.Findings, b))
	return out.String()
}

type jsonReport struct {
	Module   string          `json:"module"`
	Findings []check.Finding `json:"findings"`
	Summary  string          `json:"summary"`
}

// encodes scan result as JSON
func RenderJSON(result *engine.Result) ([]byte, error) {
	payload := jsonReport{
		Module:   result.Module.ModPath,
		Findings: result.Findings,
		Summary:  summarizeJSON(result.Findings),
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func summarizeJSON(findings []check.Finding) string {
	if check.HasSeverity(findings, check.SeverityFail) {
		return check.SeverityFail
	}
	if check.HasSeverity(findings, check.SeverityWarn) {
		return check.SeverityWarn
	}
	if len(findings) == 0 {
		return check.SeverityOK
	}
	return check.SeverityOK
}
