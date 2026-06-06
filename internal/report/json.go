package report

import (
	"bytes"
	"encoding/json"

	"github.com/mattheworford/semci/internal/cube"
	"github.com/mattheworford/semci/internal/diff"
)

type jsonReport struct {
	Summary     jsonSummary            `json:"summary"`
	Findings    []diff.Finding         `json:"findings"`
	Unsupported []cube.UnsupportedFile `json:"unsupported,omitempty"`
}

type jsonSummary struct {
	Breaking int `json:"breaking"`
	Risky    int `json:"risky"`
	Safe     int `json:"safe"`
}

func JSON(result diff.Result) (string, error) {
	payload := jsonReport{
		Summary: jsonSummary{
			Breaking: result.Count(diff.SeverityBreaking),
			Risky:    result.Count(diff.SeverityRisky),
			Safe:     result.Count(diff.SeveritySafe),
		},
		Findings:    result.Findings,
		Unsupported: result.Unsupported,
	}
	var b bytes.Buffer
	encoder := json.NewEncoder(&b)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		return "", err
	}
	return b.String(), nil
}
