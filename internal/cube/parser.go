package cube

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type yamlFile struct {
	Cubes []yamlCube `yaml:"cubes"`
}

type yamlCube struct {
	Name            string               `yaml:"name"`
	SQL             string               `yaml:"sql"`
	Title           string               `yaml:"title"`
	Description     string               `yaml:"description"`
	Public          any                  `yaml:"public"`
	Measures        []Measure            `yaml:"measures"`
	Dimensions      []Dimension          `yaml:"dimensions"`
	Segments        []Segment            `yaml:"segments"`
	Joins           []Join               `yaml:"joins"`
	PreAggregations []PreAggregation     `yaml:"pre_aggregations"`
	Extra           map[string]yaml.Node `yaml:",inline"`
}

func ParseDir(root string) (Model, error) {
	model := Model{Cubes: map[string]Cube{}}
	if _, err := os.Stat(root); err != nil {
		return model, err
	}

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".yml", ".yaml":
			return parseFile(path, &model)
		case ".js", ".ts":
			model.Unsupported = append(model.Unsupported, UnsupportedFile{
				Path:   path,
				Reason: "Cube JS/TS models are unsupported in SemCI v1",
			})
		}
		return nil
	})
	if err != nil {
		return model, err
	}

	slices.SortFunc(model.Unsupported, func(a, b UnsupportedFile) int {
		return strings.Compare(a.Path, b.Path)
	})
	return model, nil
}

func parseFile(path string, model *Model) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(data)) == "" {
		return nil
	}

	var doc yamlFile
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	for _, raw := range doc.Cubes {
		if raw.Name == "" {
			return fmt.Errorf("%s: cube is missing name", path)
		}
		if _, exists := model.Cubes[raw.Name]; exists {
			return fmt.Errorf("%s: duplicate cube %q", path, raw.Name)
		}
		c := Cube{
			Name:            raw.Name,
			SQL:             raw.SQL,
			Title:           raw.Title,
			Description:     raw.Description,
			Public:          raw.Public,
			Measures:        map[string]Measure{},
			Dimensions:      map[string]Dimension{},
			Segments:        map[string]Segment{},
			Joins:           map[string]Join{},
			PreAggregations: map[string]PreAggregation{},
		}
		for _, measure := range raw.Measures {
			if measure.Name == "" {
				return namedItemError(path, raw.Name, "measure")
			}
			normalizeFilters(&measure)
			c.Measures[measure.Name] = measure
		}
		for _, dimension := range raw.Dimensions {
			if dimension.Name == "" {
				return namedItemError(path, raw.Name, "dimension")
			}
			c.Dimensions[dimension.Name] = dimension
		}
		for _, segment := range raw.Segments {
			if segment.Name == "" {
				return namedItemError(path, raw.Name, "segment")
			}
			c.Segments[segment.Name] = segment
		}
		for _, join := range raw.Joins {
			if join.Name == "" {
				return namedItemError(path, raw.Name, "join")
			}
			c.Joins[join.Name] = join
		}
		for _, preAgg := range raw.PreAggregations {
			if preAgg.Name == "" {
				return namedItemError(path, raw.Name, "pre_aggregation")
			}
			c.PreAggregations[preAgg.Name] = preAgg
		}
		model.Cubes[c.Name] = c
	}
	return nil
}

func normalizeFilters(measure *Measure) {
	for i := range measure.Filters {
		slices.Sort(measure.Filters[i].Values)
	}
	slices.SortFunc(measure.Filters, func(a, b Filter) int {
		if a.Member != b.Member {
			return strings.Compare(a.Member, b.Member)
		}
		return strings.Compare(a.SQL, b.SQL)
	})
}

func namedItemError(path, cubeName, itemType string) error {
	return errors.New(path + ": cube " + cubeName + " has " + itemType + " without name")
}
