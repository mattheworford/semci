package diff

import (
	"testing"

	"github.com/mattheworford/semci/internal/cube"
)

func TestCompareClassifiesRemovedMeasureAsBreaking(t *testing.T) {
	result := Compare(
		modelWithMeasure("orders", "total_revenue", "amount"),
		modelWithCube("orders"),
	)

	assertFinding(t, result, SeverityBreaking, "orders.total_revenue", "removed measure")
}

func TestCompareClassifiesChangedJoinRelationshipAsBreaking(t *testing.T) {
	base := modelWithJoin("orders", "customers", "many_to_one", "{CUBE}.customer_id = {customers}.id")
	head := modelWithJoin("orders", "customers", "one_to_many", "{CUBE}.customer_id = {customers}.id")

	result := Compare(base, head)

	assertFinding(t, result, SeverityBreaking, "orders -> customers", "changed join relationship")
}

func TestCompareClassifiesAddedMeasureAsSafe(t *testing.T) {
	result := Compare(
		modelWithCube("orders"),
		modelWithMeasure("orders", "total_revenue", "amount"),
	)

	assertFinding(t, result, SeveritySafe, "orders.total_revenue", "added measure")
}

func TestCompareClassifiesChangedJoinSQLAsRisky(t *testing.T) {
	base := modelWithJoin("orders", "customers", "many_to_one", "{CUBE}.customer_id = {customers}.id")
	head := modelWithJoin("orders", "customers", "many_to_one", "{CUBE}.buyer_id = {customers}.id")

	result := Compare(base, head)

	assertFinding(t, result, SeverityRisky, "orders -> customers", "changed join SQL")
}

func TestCompareUsesHeadLocationForChangedObjects(t *testing.T) {
	base := modelWithMeasure("orders", "total_revenue", "amount")
	head := modelWithMeasure("orders", "total_revenue", "amount * exchange_rate")
	c := head.Cubes["orders"]
	measure := c.Measures["total_revenue"]
	measure.Source = cube.SourceLocation{Path: "model/orders.yml", Line: 7}
	c.Measures["total_revenue"] = measure
	head.Cubes["orders"] = c

	result := Compare(base, head)

	for _, finding := range result.Findings {
		if finding.ID == "orders.total_revenue" && finding.Change == "changed measure SQL" {
			if finding.Location.Path != "model/orders.yml" || finding.Location.Line != 7 {
				t.Fatalf("expected head location, got %#v", finding.Location)
			}
			return
		}
	}
	t.Fatalf("missing changed measure finding in %#v", result.Findings)
}

func TestCompareFormattingOnlyProducesNoFindings(t *testing.T) {
	base := modelWithMeasure("orders", "total_revenue", "amount")
	head := modelWithMeasure("orders", "total_revenue", "amount")

	result := Compare(base, head)

	if len(result.Findings) != 0 {
		t.Fatalf("expected no findings, got %#v", result.Findings)
	}
}

func assertFinding(t *testing.T, result Result, severity Severity, id, change string) {
	t.Helper()
	for _, finding := range result.Findings {
		if finding.Severity == severity && finding.ID == id && finding.Change == change {
			return
		}
	}
	t.Fatalf("missing finding %s %s %s in %#v", severity, id, change, result.Findings)
}

func modelWithCube(name string) cube.Model {
	return cube.Model{
		Cubes: map[string]cube.Cube{
			name: {
				Name:            name,
				Measures:        map[string]cube.Measure{},
				Dimensions:      map[string]cube.Dimension{},
				Segments:        map[string]cube.Segment{},
				Joins:           map[string]cube.Join{},
				PreAggregations: map[string]cube.PreAggregation{},
			},
		},
	}
}

func modelWithMeasure(cubeName, measureName, sql string) cube.Model {
	model := modelWithCube(cubeName)
	c := model.Cubes[cubeName]
	c.Measures[measureName] = cube.Measure{Name: measureName, SQL: sql, Type: "sum"}
	model.Cubes[cubeName] = c
	return model
}

func modelWithJoin(cubeName, joinName, relationship, sql string) cube.Model {
	model := modelWithCube(cubeName)
	c := model.Cubes[cubeName]
	c.Joins[joinName] = cube.Join{Name: joinName, Relationship: relationship, SQL: sql}
	model.Cubes[cubeName] = c
	return model
}
