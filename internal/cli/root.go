package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattheworford/semci/internal/config"
	"github.com/mattheworford/semci/internal/cube"
	"github.com/mattheworford/semci/internal/diff"
	"github.com/mattheworford/semci/internal/gitutil"
	"github.com/mattheworford/semci/internal/report"
	"github.com/spf13/cobra"
)

type diffOptions struct {
	configPath   string
	layer        string
	base         string
	head         string
	baseRef      string
	headRef      string
	modelPath    string
	failOn       string
	reportOutput string
	reportFormat string
}

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "semci",
		Short:         "CI for semantic layers",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newDiffCommand())
	return root
}

func newDiffCommand() *cobra.Command {
	var opts diffOptions
	cmd := &cobra.Command{
		Use:           "diff",
		Short:         "Compare two semantic layer versions",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiff(opts)
		},
	}
	cmd.Flags().StringVar(&opts.configPath, "config", "", "Path to semci.yaml")
	cmd.Flags().StringVar(&opts.layer, "layer", "", "Semantic layer adapter")
	cmd.Flags().StringVar(&opts.base, "base", "", "Base model directory")
	cmd.Flags().StringVar(&opts.head, "head", "", "Head model directory")
	cmd.Flags().StringVar(&opts.baseRef, "base-ref", "", "Base git ref")
	cmd.Flags().StringVar(&opts.headRef, "head-ref", "", "Head git ref")
	cmd.Flags().StringVar(&opts.modelPath, "model-path", "", "Model path inside a git ref checkout")
	cmd.Flags().StringVar(&opts.failOn, "fail-on", "", "Policy that fails CI: breaking, risky, or never")
	cmd.Flags().StringVar(&opts.reportOutput, "report-output", "", "Markdown report output path")
	cmd.Flags().StringVar(&opts.reportFormat, "report-format", "", "Report format")
	return cmd
}

func runDiff(opts diffOptions) error {
	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return err
	}
	applyOverrides(&cfg, opts)

	if cfg.Layer != "cube" {
		return fmt.Errorf("unsupported layer %q: SemCI v1 supports cube only", cfg.Layer)
	}
	if cfg.Report.Format != "markdown" {
		return fmt.Errorf("unsupported report format %q: SemCI v1 supports markdown only", cfg.Report.Format)
	}

	basePath, headPath, cleanup, err := resolveInputs(cfg, opts)
	defer cleanup()
	if err != nil {
		return err
	}

	baseModel, err := cube.ParseDir(basePath)
	if err != nil {
		return fmt.Errorf("parse base model: %w", err)
	}
	headModel, err := cube.ParseDir(headPath)
	if err != nil {
		return fmt.Errorf("parse head model: %w", err)
	}

	result := diff.Compare(baseModel, headModel)
	markdown := report.Markdown(result)
	if err := writeReport(cfg.Report.Output, markdown); err != nil {
		return err
	}
	fmt.Print(markdown)

	if result.ShouldFail(cfg.FailOn) {
		return fmt.Errorf("SemCI found %d breaking changes", result.Count(diff.SeverityBreaking))
	}
	return nil
}

func applyOverrides(cfg *config.Config, opts diffOptions) {
	if opts.layer != "" {
		cfg.Layer = opts.layer
	}
	if opts.modelPath != "" {
		cfg.ModelPath = opts.modelPath
	}
	if opts.failOn != "" {
		cfg.FailOn = opts.failOn
	}
	if opts.reportOutput != "" {
		cfg.Report.Output = opts.reportOutput
	}
	if opts.reportFormat != "" {
		cfg.Report.Format = opts.reportFormat
	}
	*cfg = config.ApplyDefaults(*cfg)
}

func resolveInputs(cfg config.Config, opts diffOptions) (string, string, func(), error) {
	cleanups := []func(){}
	cleanup := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}

	if opts.base != "" || opts.head != "" {
		if opts.base == "" || opts.head == "" {
			return "", "", cleanup, fmt.Errorf("--base and --head must be provided together")
		}
		return opts.base, opts.head, cleanup, nil
	}

	if opts.baseRef == "" || opts.headRef == "" {
		return "", "", cleanup, fmt.Errorf("provide either --base/--head directories or --base-ref/--head-ref git refs")
	}

	baseRoot, baseCleanup, err := gitutil.MaterializeRef(opts.baseRef)
	if err != nil {
		return "", "", cleanup, err
	}
	cleanups = append(cleanups, baseCleanup)

	headRoot, headCleanup, err := gitutil.MaterializeRef(opts.headRef)
	if err != nil {
		cleanup()
		return "", "", func() {}, err
	}
	cleanups = append(cleanups, headCleanup)

	return filepath.Join(baseRoot, cfg.ModelPath), filepath.Join(headRoot, cfg.ModelPath), cleanup, nil
}

func writeReport(path, markdown string) error {
	if path == "" || path == "-" {
		return nil
	}
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(markdown), 0644)
}
