package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	DefaultLayer        = "cube"
	DefaultModelPath    = "model"
	DefaultFailOn       = "breaking"
	DefaultReportFormat = "markdown"
	DefaultReportOutput = "semci-report.md"
)

type Config struct {
	Layer     string       `yaml:"layer"`
	ModelPath string       `yaml:"model_path"`
	FailOn    string       `yaml:"fail_on"`
	Report    ReportConfig `yaml:"report"`
	GitHub    GitHubConfig `yaml:"github"`
}

type ReportConfig struct {
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type GitHubConfig struct {
	Comment bool `yaml:"comment"`
}

func Defaults() Config {
	return Config{
		Layer:     DefaultLayer,
		ModelPath: DefaultModelPath,
		FailOn:    DefaultFailOn,
		Report: ReportConfig{
			Format: DefaultReportFormat,
			Output: DefaultReportOutput,
		},
		GitHub: GitHubConfig{Comment: true},
	}
}

func Load(path string) (Config, error) {
	cfg := Defaults()
	if path == "" {
		if _, err := os.Stat("semci.yaml"); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return cfg, nil
			}
			return cfg, err
		}
		path = "semci.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return ApplyDefaults(cfg), nil
}

func ApplyDefaults(cfg Config) Config {
	defaults := Defaults()
	if cfg.Layer == "" {
		cfg.Layer = defaults.Layer
	}
	if cfg.ModelPath == "" {
		cfg.ModelPath = defaults.ModelPath
	}
	if cfg.FailOn == "" {
		cfg.FailOn = defaults.FailOn
	}
	if cfg.Report.Format == "" {
		cfg.Report.Format = defaults.Report.Format
	}
	if cfg.Report.Output == "" {
		cfg.Report.Output = defaults.Report.Output
	}
	return cfg
}
