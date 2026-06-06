package cube

type Model struct {
	Cubes       map[string]Cube
	Unsupported []UnsupportedFile
}

type SourceLocation struct {
	Path string `json:"path,omitempty"`
	Line int    `json:"line,omitempty"`
}

type UnsupportedFile struct {
	Path   string `json:"path"`
	Line   int    `json:"line,omitempty"`
	Reason string `json:"reason"`
}

type Cube struct {
	Name            string
	SQL             string
	Title           string
	Description     string
	Public          any
	Measures        map[string]Measure
	Dimensions      map[string]Dimension
	Segments        map[string]Segment
	Joins           map[string]Join
	PreAggregations map[string]PreAggregation
	Source          SourceLocation `yaml:"-"`
}

type Measure struct {
	Name        string         `yaml:"name"`
	SQL         string         `yaml:"sql"`
	Type        string         `yaml:"type"`
	Title       string         `yaml:"title"`
	Description string         `yaml:"description"`
	Filters     []Filter       `yaml:"filters"`
	Public      any            `yaml:"public"`
	Source      SourceLocation `yaml:"-"`
}

type Dimension struct {
	Name        string         `yaml:"name"`
	SQL         string         `yaml:"sql"`
	Type        string         `yaml:"type"`
	Title       string         `yaml:"title"`
	Description string         `yaml:"description"`
	Public      any            `yaml:"public"`
	Source      SourceLocation `yaml:"-"`
}

type Segment struct {
	Name        string         `yaml:"name"`
	SQL         string         `yaml:"sql"`
	Title       string         `yaml:"title"`
	Description string         `yaml:"description"`
	Public      any            `yaml:"public"`
	Source      SourceLocation `yaml:"-"`
}

type Join struct {
	Name         string         `yaml:"name"`
	SQL          string         `yaml:"sql"`
	Relationship string         `yaml:"relationship"`
	Type         string         `yaml:"type"`
	Source       SourceLocation `yaml:"-"`
}

type PreAggregation struct {
	Name   string         `yaml:"name"`
	Type   string         `yaml:"type"`
	Source SourceLocation `yaml:"-"`
}

type Filter struct {
	SQL    string   `yaml:"sql"`
	Member string   `yaml:"member"`
	Values []string `yaml:"values"`
}
