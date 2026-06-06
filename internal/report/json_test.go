package report

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mattheworford/semci/internal/cube"
	"github.com/mattheworford/semci/internal/diff"
)

func TestJSONIncludesSummaryFindingsAndLocations(t *testing.T) {
	output, err := JSON(sampleResult())
	if err != nil {
		t.Fatalf("JSON returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, output)
	}
	summary := payload["summary"].(map[string]any)
	if summary["breaking"].(float64) != 1 {
		t.Fatalf("expected 1 breaking finding, got %#v", summary)
	}
	if !strings.Contains(output, `"location"`) {
		t.Fatalf("expected JSON location, got:\n%s", output)
	}
}

func TestGitHubAnnotationsEscapesAndUsesLocations(t *testing.T) {
	output := GitHubAnnotations(sampleResult())

	if !strings.Contains(output, "::error ") {
		t.Fatalf("expected error annotation, got:\n%s", output)
	}
	if !strings.Contains(output, "file=model/orders.yml,line=12") {
		t.Fatalf("expected file and line, got:\n%s", output)
	}
	if !strings.Contains(output, "amount %25 tax") {
		t.Fatalf("expected escaped percent in message, got:\n%s", output)
	}
}

func sampleResult() diff.Result {
	return diff.Result{
		Findings: []diff.Finding{
			{
				Severity:   diff.SeverityBreaking,
				ObjectType: "measure",
				ID:         "orders.total_revenue",
				Change:     "changed measure SQL",
				Before:     "amount",
				After:      "amount % tax",
				Location:   cube.SourceLocation{Path: "model/orders.yml", Line: 12},
			},
		},
	}
}
