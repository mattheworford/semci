package report

import (
	"fmt"
	"strings"

	"github.com/mattheworford/semci/internal/cube"
	"github.com/mattheworford/semci/internal/diff"
)

func GitHubAnnotations(result diff.Result) string {
	var b strings.Builder
	for _, unsupported := range result.Unsupported {
		writeAnnotation(&b, "warning", "SemCI unsupported file", unsupported.Path, unsupported.Line, unsupported.Reason)
	}
	for _, finding := range result.Findings {
		level := annotationLevel(finding.Severity)
		title := fmt.Sprintf("SemCI %s: %s", finding.Severity, finding.Change)
		location := annotationLocation(finding)
		writeAnnotation(&b, level, title, location.Path, location.Line, annotationMessage(finding))
	}
	return b.String()
}

func annotationLocation(finding diff.Finding) cube.SourceLocation {
	if strings.HasPrefix(finding.Change, "removed ") {
		return cube.SourceLocation{}
	}
	return finding.Location
}

func annotationLevel(severity diff.Severity) string {
	switch severity {
	case diff.SeverityBreaking:
		return "error"
	case diff.SeverityRisky:
		return "warning"
	default:
		return "notice"
	}
}

func annotationMessage(finding diff.Finding) string {
	parts := []string{fmt.Sprintf("%s %s", finding.ID, finding.Change)}
	if finding.Before != "" {
		parts = append(parts, "Before: "+oneLine(finding.Before))
	}
	if finding.After != "" {
		parts = append(parts, "After: "+oneLine(finding.After))
	}
	if finding.Suggestion != "" {
		parts = append(parts, "Suggestion: "+finding.Suggestion)
	}
	return strings.Join(parts, ". ")
}

func writeAnnotation(b *strings.Builder, level, title, path string, line int, message string) {
	properties := []string{"title=" + escapeProperty(title)}
	location := cube.SourceLocation{Path: path, Line: line}
	if location.Path != "" {
		properties = append(properties, "file="+escapeProperty(location.Path))
	}
	if location.Line > 0 {
		properties = append(properties, "line="+fmt.Sprint(location.Line))
	}
	b.WriteString("::")
	b.WriteString(level)
	b.WriteString(" ")
	b.WriteString(strings.Join(properties, ","))
	b.WriteString("::")
	b.WriteString(escapeMessage(message))
	b.WriteString("\n")
}

func escapeMessage(value string) string {
	value = strings.ReplaceAll(value, "%", "%25")
	value = strings.ReplaceAll(value, "\r", "%0D")
	value = strings.ReplaceAll(value, "\n", "%0A")
	return value
}

func escapeProperty(value string) string {
	value = escapeMessage(value)
	value = strings.ReplaceAll(value, ":", "%3A")
	value = strings.ReplaceAll(value, ",", "%2C")
	return value
}
