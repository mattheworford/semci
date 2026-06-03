package report

import (
	"fmt"
	"strings"

	"github.com/mattheworford/semci/internal/diff"
)

const StickyCommentMarker = "<!-- semci-report -->"

func Markdown(result diff.Result) string {
	var b strings.Builder
	b.WriteString(StickyCommentMarker + "\n")
	b.WriteString("# SemCI Report\n\n")
	b.WriteString("CI for semantic layers. Catch broken metrics and risky joins before they ship.\n\n")
	b.WriteString(fmt.Sprintf(
		"**Summary:** %d breaking, %d risky, %d safe changes.\n\n",
		result.Count(diff.SeverityBreaking),
		result.Count(diff.SeverityRisky),
		result.Count(diff.SeveritySafe),
	))

	if len(result.Unsupported) > 0 {
		b.WriteString("## Unsupported Files\n\n")
		for _, unsupported := range result.Unsupported {
			b.WriteString(fmt.Sprintf("- `%s`: %s\n", unsupported.Path, unsupported.Reason))
		}
		b.WriteString("\n")
	}

	writeGroup(&b, "Breaking Changes", diff.SeverityBreaking, result.Findings)
	writeGroup(&b, "Risky Changes", diff.SeverityRisky, result.Findings)
	writeGroup(&b, "Safe Changes", diff.SeveritySafe, result.Findings)

	if len(result.Findings) == 0 {
		b.WriteString("No semantic changes detected.\n")
	}
	return b.String()
}

func writeGroup(b *strings.Builder, title string, severity diff.Severity, findings []diff.Finding) {
	matched := make([]diff.Finding, 0)
	for _, finding := range findings {
		if finding.Severity == severity {
			matched = append(matched, finding)
		}
	}
	if len(matched) == 0 {
		return
	}

	b.WriteString("## " + title + "\n\n")
	for _, finding := range matched {
		b.WriteString(fmt.Sprintf("### `%s`\n\n", finding.ID))
		b.WriteString(fmt.Sprintf("- **Type:** %s\n", finding.ObjectType))
		b.WriteString(fmt.Sprintf("- **Change:** %s\n", finding.Change))
		if finding.Before != "" {
			b.WriteString(fmt.Sprintf("- **Before:** `%s`\n", oneLine(finding.Before)))
		}
		if finding.After != "" {
			b.WriteString(fmt.Sprintf("- **After:** `%s`\n", oneLine(finding.After)))
		}
		if finding.Suggestion != "" {
			b.WriteString(fmt.Sprintf("- **Suggestion:** %s\n", finding.Suggestion))
		}
		b.WriteString("\n")
	}
}

func oneLine(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 180 {
		return value[:177] + "..."
	}
	return value
}
