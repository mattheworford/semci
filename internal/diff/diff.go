package diff

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/mattheworford/semci/internal/cube"
)

type Severity string

const (
	SeveritySafe     Severity = "safe"
	SeverityRisky    Severity = "risky"
	SeverityBreaking Severity = "breaking"
)

type Finding struct {
	Severity   Severity            `json:"severity"`
	ObjectType string              `json:"object_type"`
	ID         string              `json:"id"`
	Change     string              `json:"change"`
	Before     string              `json:"before,omitempty"`
	After      string              `json:"after,omitempty"`
	Suggestion string              `json:"suggestion,omitempty"`
	Location   cube.SourceLocation `json:"location,omitempty"`
}

type Result struct {
	Findings    []Finding              `json:"findings"`
	Unsupported []cube.UnsupportedFile `json:"unsupported"`
}

func Compare(base, head cube.Model) Result {
	result := Result{Unsupported: append(base.Unsupported, head.Unsupported...)}
	compareCubes(&result, base.Cubes, head.Cubes)
	slices.SortFunc(result.Findings, func(a, b Finding) int {
		if severityRank(a.Severity) != severityRank(b.Severity) {
			return severityRank(b.Severity) - severityRank(a.Severity)
		}
		return strings.Compare(a.ID, b.ID)
	})
	return result
}

func compareCubes(result *Result, base, head map[string]cube.Cube) {
	for name, oldCube := range base {
		newCube, ok := head[name]
		if !ok {
			result.add(SeverityBreaking, "cube", name, "removed cube", oldCube.SQL, "", migrationHint(name), oldCube.Source)
			continue
		}
		if oldCube.SQL != newCube.SQL {
			result.add(SeverityRisky, "cube", name, "changed cube SQL", oldCube.SQL, newCube.SQL, "", newCube.Source)
		}
		if oldCube.Title != newCube.Title {
			result.add(SeverityRisky, "cube", name, "changed public title", oldCube.Title, newCube.Title, "", newCube.Source)
		}
		if oldCube.Description != newCube.Description {
			result.add(SeverityRisky, "cube", name, "changed public description", oldCube.Description, newCube.Description, "", newCube.Source)
		}
		compareMeasures(result, name, oldCube.Measures, newCube.Measures)
		compareDimensions(result, name, oldCube.Dimensions, newCube.Dimensions)
		compareSegments(result, name, oldCube.Segments, newCube.Segments)
		compareJoins(result, name, oldCube.Joins, newCube.Joins)
		comparePreAggregations(result, name, oldCube.PreAggregations, newCube.PreAggregations)
	}
	for name := range head {
		if _, ok := base[name]; !ok {
			result.add(SeveritySafe, "cube", name, "added cube", "", head[name].SQL, "", head[name].Source)
		}
	}
}

func compareMeasures(result *Result, cubeName string, base, head map[string]cube.Measure) {
	for name, oldMeasure := range base {
		id := cubeName + "." + name
		newMeasure, ok := head[name]
		if !ok {
			result.add(SeverityBreaking, "measure", id, "removed measure", oldMeasure.SQL, "", migrationHint(id), oldMeasure.Source)
			continue
		}
		if oldMeasure.SQL != newMeasure.SQL {
			result.add(SeverityBreaking, "measure", id, "changed measure SQL", oldMeasure.SQL, newMeasure.SQL, migrationHint(id), newMeasure.Source)
		}
		if oldMeasure.Type != newMeasure.Type {
			result.add(SeverityBreaking, "measure", id, "changed measure type", oldMeasure.Type, newMeasure.Type, migrationHint(id), newMeasure.Source)
		}
		if !reflect.DeepEqual(oldMeasure.Filters, newMeasure.Filters) {
			severity := SeverityBreaking
			change := "changed measure filters"
			if len(newMeasure.Filters) > len(oldMeasure.Filters) && len(oldMeasure.Filters) == 0 {
				severity = SeverityRisky
				change = "added measure filters"
			}
			result.add(severity, "measure", id, change, fmt.Sprint(oldMeasure.Filters), fmt.Sprint(newMeasure.Filters), migrationHint(id), newMeasure.Source)
		}
		if oldMeasure.Title != newMeasure.Title {
			result.add(SeverityRisky, "measure", id, "changed public title", oldMeasure.Title, newMeasure.Title, "", newMeasure.Source)
		}
		if oldMeasure.Description != newMeasure.Description {
			result.add(SeverityRisky, "measure", id, "changed public description", oldMeasure.Description, newMeasure.Description, "", newMeasure.Source)
		}
	}
	for name, measure := range head {
		if _, ok := base[name]; !ok {
			result.add(SeveritySafe, "measure", cubeName+"."+name, "added measure", "", measure.SQL, "", measure.Source)
		}
	}
}

func compareDimensions(result *Result, cubeName string, base, head map[string]cube.Dimension) {
	for name, oldDimension := range base {
		id := cubeName + "." + name
		newDimension, ok := head[name]
		if !ok {
			result.add(SeverityBreaking, "dimension", id, "removed dimension", oldDimension.SQL, "", migrationHint(id), oldDimension.Source)
			continue
		}
		if oldDimension.SQL != newDimension.SQL {
			result.add(SeverityBreaking, "dimension", id, "changed dimension SQL", oldDimension.SQL, newDimension.SQL, migrationHint(id), newDimension.Source)
		}
		if oldDimension.Type != newDimension.Type {
			result.add(SeverityBreaking, "dimension", id, "changed dimension type", oldDimension.Type, newDimension.Type, migrationHint(id), newDimension.Source)
		}
		if oldDimension.Title != newDimension.Title {
			result.add(SeverityRisky, "dimension", id, "changed public title", oldDimension.Title, newDimension.Title, "", newDimension.Source)
		}
		if oldDimension.Description != newDimension.Description {
			result.add(SeverityRisky, "dimension", id, "changed public description", oldDimension.Description, newDimension.Description, "", newDimension.Source)
		}
	}
	for name, dimension := range head {
		if _, ok := base[name]; !ok {
			result.add(SeveritySafe, "dimension", cubeName+"."+name, "added dimension", "", dimension.SQL, "", dimension.Source)
		}
	}
}

func compareSegments(result *Result, cubeName string, base, head map[string]cube.Segment) {
	for name, oldSegment := range base {
		id := cubeName + "." + name
		newSegment, ok := head[name]
		if !ok {
			result.add(SeverityBreaking, "segment", id, "removed segment", oldSegment.SQL, "", migrationHint(id), oldSegment.Source)
			continue
		}
		if oldSegment.SQL != newSegment.SQL {
			result.add(SeverityRisky, "segment", id, "changed segment SQL", oldSegment.SQL, newSegment.SQL, "", newSegment.Source)
		}
		if oldSegment.Title != newSegment.Title {
			result.add(SeverityRisky, "segment", id, "changed public title", oldSegment.Title, newSegment.Title, "", newSegment.Source)
		}
		if oldSegment.Description != newSegment.Description {
			result.add(SeverityRisky, "segment", id, "changed public description", oldSegment.Description, newSegment.Description, "", newSegment.Source)
		}
	}
	for name, segment := range head {
		if _, ok := base[name]; !ok {
			result.add(SeveritySafe, "segment", cubeName+"."+name, "added segment", "", segment.SQL, "", segment.Source)
		}
	}
}

func compareJoins(result *Result, cubeName string, base, head map[string]cube.Join) {
	for name, oldJoin := range base {
		id := cubeName + " -> " + name
		newJoin, ok := head[name]
		if !ok {
			result.add(SeverityBreaking, "join", id, "removed join", oldJoin.SQL, "", "Keep the old join until dependent queries are migrated.", oldJoin.Source)
			continue
		}
		if oldJoin.Relationship != newJoin.Relationship {
			result.add(SeverityBreaking, "join", id, "changed join relationship", oldJoin.Relationship, newJoin.Relationship, "Review for fanout risk before merging.", newJoin.Source)
		}
		if oldJoin.SQL != newJoin.SQL {
			result.add(SeverityRisky, "join", id, "changed join SQL", oldJoin.SQL, newJoin.SQL, "Review affected queries for row multiplication or changed filters.", newJoin.Source)
		}
		if oldJoin.Type != newJoin.Type {
			result.add(SeverityRisky, "join", id, "changed join type", oldJoin.Type, newJoin.Type, "", newJoin.Source)
		}
	}
	for name, join := range head {
		if _, ok := base[name]; !ok {
			result.add(SeverityRisky, "join", cubeName+" -> "+name, "added join", "", join.SQL, "Review for fanout risk before relying on this join.", join.Source)
		}
	}
}

func comparePreAggregations(result *Result, cubeName string, base, head map[string]cube.PreAggregation) {
	for name, oldPreAgg := range base {
		id := cubeName + "." + name
		if _, ok := head[name]; !ok {
			result.add(SeverityRisky, "pre_aggregation", id, "removed pre-aggregation", oldPreAgg.Type, "", "", oldPreAgg.Source)
		}
	}
	for name, preAgg := range head {
		if _, ok := base[name]; !ok {
			result.add(SeveritySafe, "pre_aggregation", cubeName+"."+name, "added pre-aggregation", "", preAgg.Type, "", preAgg.Source)
		}
	}
}

func (r *Result) add(severity Severity, objectType, id, change, before, after, suggestion string, location cube.SourceLocation) {
	r.Findings = append(r.Findings, Finding{
		Severity:   severity,
		ObjectType: objectType,
		ID:         id,
		Change:     change,
		Before:     strings.TrimSpace(before),
		After:      strings.TrimSpace(after),
		Suggestion: suggestion,
		Location:   location,
	})
}

func (r Result) Count(severity Severity) int {
	count := 0
	for _, finding := range r.Findings {
		if finding.Severity == severity {
			count++
		}
	}
	return count
}

func (r Result) ShouldFail(failOn string) bool {
	switch strings.ToLower(failOn) {
	case "never", "none":
		return false
	case "risky":
		return r.Count(SeverityBreaking) > 0 || r.Count(SeverityRisky) > 0
	default:
		return r.Count(SeverityBreaking) > 0
	}
}

func severityRank(severity Severity) int {
	switch severity {
	case SeverityBreaking:
		return 3
	case SeverityRisky:
		return 2
	default:
		return 1
	}
}

func migrationHint(id string) string {
	return "Keep `" + id + "` available, add a versioned replacement, and migrate dependents before removal."
}
