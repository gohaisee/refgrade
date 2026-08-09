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
	FormatSARIF    Format = "sarif"
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
	case "sarif":
		return FormatSARIF, nil
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
		return RenderJSON(result, b)
	case FormatSARIF:
		return RenderSARIF(result, b)
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
	} else {
		out.WriteString(b.T("report.findings"))
		out.WriteString(":\n")
		for _, f := range result.Findings {
			writeTextFinding(&out, b, f)
		}
	}

	writeStatusSummary(&out, b, result.Statuses)
	writeSummary(&out, b, result.Findings, result.Statuses)
	writeFooter(&out, b)
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

func writeStatusSummary(out *strings.Builder, b *i18n.Bundle, statuses []check.CheckStatus) {
	na := countStatus(statuses, check.SeverityNA)
	if na == 0 {
		return
	}
	out.WriteString(b.T("report.check_status"))
	out.WriteString(":\n")
	for _, st := range statuses {
		if st.Severity != check.SeverityNA {
			continue
		}
		out.WriteString("  ")
		out.WriteString(st.ID)
		out.WriteString(" [")
		out.WriteString(severityLabel(b, check.SeverityNA))
		out.WriteString("]\n")
	}
}

func writeSummary(out *strings.Builder, b *i18n.Bundle, findings []check.Finding, statuses []check.CheckStatus) {
	out.WriteString(b.T("report.summary"))
	out.WriteString(": ")
	out.WriteString(summarize(findings, statuses, b))
	out.WriteByte('\n')
}

func writeFooter(out *strings.Builder, b *i18n.Bundle) {
	lead := b.T("report.gaps.lead")
	if lead == "" || lead == "report.gaps.lead" {
		return
	}
	out.WriteByte('\n')
	out.WriteString(lead)
	out.WriteByte('\n')
	for _, key := range []string{"report.gap.idor", "report.gap.dast", "report.gap.k8s"} {
		if line := b.T(key); line != "" && line != key {
			out.WriteString("  - ")
			out.WriteString(line)
			out.WriteByte('\n')
		}
	}
	if see := b.T("report.gaps.see"); see != "" && see != "report.gaps.see" {
		out.WriteString(see)
		out.WriteByte('\n')
	}
}

func summarize(findings []check.Finding, statuses []check.CheckStatus, b *i18n.Bundle) string {
	fail, warn, na := countFindings(findings)
	na += countStatus(statuses, check.SeverityNA)
	parts := []string{}
	if fail > 0 {
		parts = append(parts, fmt.Sprintf("%s=%d", b.T("report.fail"), fail))
	}
	if warn > 0 {
		parts = append(parts, fmt.Sprintf("%s=%d", b.T("report.warn"), warn))
	}
	if na > 0 {
		parts = append(parts, fmt.Sprintf("%s=%d", b.T("report.na"), na))
	}
	if len(parts) == 0 {
		return b.T("report.ok")
	}
	return strings.Join(parts, ", ")
}

func countFindings(findings []check.Finding) (fail, warn, na int) {
	for _, f := range findings {
		switch f.Severity {
		case check.SeverityFail:
			fail++
		case check.SeverityWarn:
			warn++
		case check.SeverityNA:
			na++
		}
	}
	return fail, warn, na
}

func countStatus(statuses []check.CheckStatus, sev string) int {
	n := 0
	for _, st := range statuses {
		if st.Severity == sev {
			n++
		}
	}
	return n
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
	} else {
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
	}
	fmt.Fprintf(&out, "**%s:** %s\n", b.T("report.summary"), summarize(result.Findings, result.Statuses, b))
	writeFooter(&out, b)
	return out.String()
}

type jsonReport struct {
	Module   string              `json:"module"`
	Findings []check.Finding     `json:"findings"`
	Statuses []check.CheckStatus `json:"statuses"`
	Summary  string              `json:"summary"`
}

// encodes scan result as JSON
func RenderJSON(result *engine.Result, b *i18n.Bundle) ([]byte, error) {
	payload := jsonReport{
		Module:   result.Module.ModPath,
		Findings: result.Findings,
		Statuses: result.Statuses,
		Summary:  summarize(result.Findings, result.Statuses, b),
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SARIF 2.1.0 minimal mapping
func RenderSARIF(result *engine.Result, b *i18n.Bundle) ([]byte, error) {
	type sarifResult struct {
		RuleID    string `json:"ruleId"`
		Level     string `json:"level"`
		Message   struct {
			Text string `json:"text"`
		} `json:"message"`
		Locations []struct {
			PhysicalLocation struct {
				ArtifactLocation struct {
					URI string `json:"uri"`
				} `json:"artifactLocation"`
				Region struct {
					StartLine int `json:"startLine"`
				} `json:"region"`
			} `json:"physicalLocation"`
		} `json:"locations"`
	}
	type sarifRule struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	type sarifDoc struct {
		Version string `json:"version"`
		Schema  string `json:"$schema"`
		Runs    []struct {
			Tool struct {
				Driver struct {
					Name  string      `json:"name"`
					Rules []sarifRule `json:"rules"`
				} `json:"driver"`
			} `json:"tool"`
			Results []sarifResult `json:"results"`
		} `json:"runs"`
	}

	rules := map[string]struct{}{}
	var results []sarifResult
	for _, f := range result.Findings {
		rules[f.ID] = struct{}{}
		var r sarifResult
		r.RuleID = f.ID
		r.Level = sarifLevel(f.Severity)
		r.Message.Text = f.Why
		r.Locations = append(r.Locations, struct {
			PhysicalLocation struct {
				ArtifactLocation struct {
					URI string `json:"uri"`
				} `json:"artifactLocation"`
				Region struct {
					StartLine int `json:"startLine"`
				} `json:"region"`
			} `json:"physicalLocation"`
		}{})
		r.Locations[0].PhysicalLocation.ArtifactLocation.URI = f.File
		r.Locations[0].PhysicalLocation.Region.StartLine = f.Line
		results = append(results, r)
	}
	var ruleList []sarifRule
	for id := range rules {
		ruleList = append(ruleList, sarifRule{ID: id, Name: id})
	}

	doc := sarifDoc{
		Version: "2.1.0",
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
	}
	doc.Runs = append(doc.Runs, struct {
		Tool struct {
			Driver struct {
				Name  string      `json:"name"`
				Rules []sarifRule `json:"rules"`
			} `json:"driver"`
		} `json:"tool"`
		Results []sarifResult `json:"results"`
	}{})
	doc.Runs[0].Tool.Driver.Name = "refgrade"
	doc.Runs[0].Tool.Driver.Rules = ruleList
	doc.Runs[0].Results = results

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	_ = b
	return buf.Bytes(), nil
}

func sarifLevel(sev string) string {
	switch sev {
	case check.SeverityFail:
		return "error"
	case check.SeverityWarn:
		return "warning"
	default:
		return "note"
	}
}
