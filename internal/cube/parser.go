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
	Source          SourceLocation       `yaml:"-"`
}

func ParseDir(root string) (Model, error) {
	return ParseDirWithSourcePrefix(root, root)
}

func ParseDirWithSourcePrefix(root, sourcePrefix string) (Model, error) {
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
			return parseFile(root, sourcePrefix, path, &model)
		case ".js", ".ts":
			model.Unsupported = append(model.Unsupported, UnsupportedFile{
				Path:   sourcePath(root, sourcePrefix, path),
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

func parseFile(root, sourcePrefix, path string, model *Model) error {
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
			Source:          withPath(raw.Source, root, sourcePrefix, path),
		}
		for _, measure := range raw.Measures {
			if measure.Name == "" {
				return namedItemError(path, raw.Name, "measure")
			}
			normalizeFilters(&measure)
			measure.Source = withPath(measure.Source, root, sourcePrefix, path)
			c.Measures[measure.Name] = measure
		}
		for _, dimension := range raw.Dimensions {
			if dimension.Name == "" {
				return namedItemError(path, raw.Name, "dimension")
			}
			dimension.Source = withPath(dimension.Source, root, sourcePrefix, path)
			c.Dimensions[dimension.Name] = dimension
		}
		for _, segment := range raw.Segments {
			if segment.Name == "" {
				return namedItemError(path, raw.Name, "segment")
			}
			segment.Source = withPath(segment.Source, root, sourcePrefix, path)
			c.Segments[segment.Name] = segment
		}
		for _, join := range raw.Joins {
			if join.Name == "" {
				return namedItemError(path, raw.Name, "join")
			}
			join.Source = withPath(join.Source, root, sourcePrefix, path)
			c.Joins[join.Name] = join
		}
		for _, preAgg := range raw.PreAggregations {
			if preAgg.Name == "" {
				return namedItemError(path, raw.Name, "pre_aggregation")
			}
			preAgg.Source = withPath(preAgg.Source, root, sourcePrefix, path)
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

func (c *yamlCube) UnmarshalYAML(value *yaml.Node) error {
	type raw yamlCube
	var decoded raw
	if err := value.Decode(&decoded); err != nil {
		return err
	}
	*c = yamlCube(decoded)
	c.Source.Line = value.Line
	return nil
}

func (m *Measure) UnmarshalYAML(value *yaml.Node) error {
	type raw Measure
	var decoded raw
	if err := value.Decode(&decoded); err != nil {
		return err
	}
	*m = Measure(decoded)
	m.Source.Line = value.Line
	return nil
}

func (d *Dimension) UnmarshalYAML(value *yaml.Node) error {
	type raw Dimension
	var decoded raw
	if err := value.Decode(&decoded); err != nil {
		return err
	}
	*d = Dimension(decoded)
	d.Source.Line = value.Line
	return nil
}

func (s *Segment) UnmarshalYAML(value *yaml.Node) error {
	type raw Segment
	var decoded raw
	if err := value.Decode(&decoded); err != nil {
		return err
	}
	*s = Segment(decoded)
	s.Source.Line = value.Line
	return nil
}

func (j *Join) UnmarshalYAML(value *yaml.Node) error {
	type raw Join
	var decoded raw
	if err := value.Decode(&decoded); err != nil {
		return err
	}
	*j = Join(decoded)
	j.Source.Line = value.Line
	return nil
}

func (p *PreAggregation) UnmarshalYAML(value *yaml.Node) error {
	type raw PreAggregation
	var decoded raw
	if err := value.Decode(&decoded); err != nil {
		return err
	}
	*p = PreAggregation(decoded)
	p.Source.Line = value.Line
	return nil
}

func withPath(location SourceLocation, root, sourcePrefix, path string) SourceLocation {
	location.Path = sourcePath(root, sourcePrefix, path)
	return location
}

func sourcePath(root, sourcePrefix, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = filepath.Base(path)
	}
	if sourcePrefix == "" || sourcePrefix == "." {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(filepath.Join(sourcePrefix, rel))
}
