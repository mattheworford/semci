package cube

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDirParsesCubeYAML(t *testing.T) {
	model, err := ParseDir("../../fixtures/cube/old")
	if err != nil {
		t.Fatalf("ParseDir returned error: %v", err)
	}

	orders := model.Cubes["orders"]
	if orders.Name != "orders" {
		t.Fatalf("expected orders cube, got %#v", model.Cubes)
	}
	if len(orders.Measures) != 3 {
		t.Fatalf("expected 3 measures, got %d", len(orders.Measures))
	}
	if orders.Joins["customers"].Relationship != "many_to_one" {
		t.Fatalf("expected many_to_one relationship, got %q", orders.Joins["customers"].Relationship)
	}
}

func TestParseDirReportsUnsupportedJSTSFiles(t *testing.T) {
	model, err := ParseDir("../../fixtures/cube/unsupported")
	if err != nil {
		t.Fatalf("ParseDir returned error: %v", err)
	}

	if len(model.Unsupported) != 1 {
		t.Fatalf("expected 1 unsupported file, got %d", len(model.Unsupported))
	}
	if model.Unsupported[0].Reason == "" {
		t.Fatal("expected unsupported reason")
	}
}

func TestParseDirIgnoresFormattingOnlyOrderingChanges(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "base.yml"), `
cubes:
  - name: orders
    measures:
      - name: total_revenue
        sql: amount
        type: sum
        filters:
          - member: status
            values: [paid, settled]
`)
	model, err := ParseDir(root)
	if err != nil {
		t.Fatalf("ParseDir returned error: %v", err)
	}

	filters := model.Cubes["orders"].Measures["total_revenue"].Filters
	if got := filters[0].Values[0]; got != "paid" {
		t.Fatalf("expected sorted filter values, got %q", got)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}
