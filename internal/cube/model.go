package cube

type Model struct {
	Cubes       map[string]Cube
	Unsupported []UnsupportedFile
}

type UnsupportedFile struct {
	Path   string
	Reason string
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
}

type Measure struct {
	Name        string   `yaml:"name"`
	SQL         string   `yaml:"sql"`
	Type        string   `yaml:"type"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Filters     []Filter `yaml:"filters"`
	Public      any      `yaml:"public"`
}

type Dimension struct {
	Name        string `yaml:"name"`
	SQL         string `yaml:"sql"`
	Type        string `yaml:"type"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Public      any    `yaml:"public"`
}

type Segment struct {
	Name        string `yaml:"name"`
	SQL         string `yaml:"sql"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Public      any    `yaml:"public"`
}

type Join struct {
	Name         string `yaml:"name"`
	SQL          string `yaml:"sql"`
	Relationship string `yaml:"relationship"`
	Type         string `yaml:"type"`
}

type PreAggregation struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Filter struct {
	SQL    string   `yaml:"sql"`
	Member string   `yaml:"member"`
	Values []string `yaml:"values"`
}
