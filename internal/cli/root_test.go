package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffCommandWritesReportAndFailsOnBreaking(t *testing.T) {
	reportPath := filepath.Join(t.TempDir(), "report.md")
	cmd := NewRootCommand()
	cmd.SetArgs([]string{
		"diff",
		"--base", "../../fixtures/cube/old",
		"--head", "../../fixtures/cube/new",
		"--report-output", reportPath,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected breaking changes to fail by default")
	}
	report, readErr := os.ReadFile(reportPath)
	if readErr != nil {
		t.Fatalf("expected report to be written: %v", readErr)
	}
	if !strings.Contains(string(report), "orders.total_revenue") {
		t.Fatalf("expected report to mention changed metric, got:\n%s", report)
	}
}

func TestDiffCommandAllowsNeverFailPolicy(t *testing.T) {
	reportPath := filepath.Join(t.TempDir(), "report.md")
	cmd := NewRootCommand()
	cmd.SetArgs([]string{
		"diff",
		"--base", "../../fixtures/cube/old",
		"--head", "../../fixtures/cube/new",
		"--fail-on", "never",
		"--report-output", reportPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected never fail policy to allow changes: %v", err)
	}
}

func TestDiffCommandLoadsConfigAndAllowsFlagOverrides(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "semci.yaml")
	reportPath := filepath.Join(tmp, "configured.md")
	if err := os.WriteFile(configPath, []byte(`
layer: cube
fail_on: never
report:
  output: ignored.md
`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cmd := NewRootCommand()
	cmd.SetArgs([]string{
		"diff",
		"--config", configPath,
		"--base", "../../fixtures/cube/old",
		"--head", "../../fixtures/cube/new",
		"--report-output", reportPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected configured never fail policy: %v", err)
	}
	if _, err := os.Stat(reportPath); err != nil {
		t.Fatalf("expected overridden report path to exist: %v", err)
	}
}
