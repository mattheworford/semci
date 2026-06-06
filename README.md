# SemCI

CI for semantic layers. Catch broken metrics and risky joins before they ship.

SemCI compares two Cube semantic model versions, classifies semantic changes, and writes a PR-ready Markdown or JSON report.

```text
This PR changes total revenue.

Impact:
- 2 breaking changes
- 3 risky changes
- 2 safe changes

Breaking:
- orders.total_revenue changed measure SQL
- orders -> customers changed join relationship

Risky:
- orders -> accounts added join
- orders.completed changed segment SQL
```

## Install

```bash
go install github.com/mattheworford/semci/cmd/semci@latest
```

## Run Locally

Compare two directories:

```bash
semci diff --layer cube --base fixtures/cube/old --head fixtures/cube/new
```

Compare two git refs using a model path:

```bash
semci diff --layer cube --base-ref main --head-ref HEAD --model-path model
```

Write machine-readable JSON:

```bash
semci diff --layer cube --base fixtures/cube/old --head fixtures/cube/new --report-format json --report-output semci-report.json
```

Emit GitHub Actions annotations:

```bash
semci diff --layer cube --base-ref main --head-ref HEAD --model-path model --github-annotations
```

Use a config file:

```yaml
layer: cube
model_path: model
fail_on: breaking
report:
  format: markdown
  output: semci-report.md
github:
  comment: true
  annotations: true
```

```bash
semci diff --config semci.yaml --base-ref main --head-ref HEAD
```

## GitHub Action

```yaml
name: SemCI

on:
  pull_request:

jobs:
  semci:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: read
      issues: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: mattheworford/semci@v1
        with:
          layer: cube
          model-path: model
          base-ref: main
          head-ref: ${{ github.sha }}
          fail-on: breaking
          github-annotations: "true"
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

SemCI exits nonzero only when the configured policy is met. The default is `breaking`, so risky changes are reported but do not block CI.

The GitHub Action emits annotations by default. Breaking changes are errors, risky changes are warnings, and safe changes are notices.

## What SemCI V1 Supports

SemCI v1 supports **Cube YAML** models only. It parses:

- cubes
- measures
- dimensions
- segments
- joins
- pre-aggregation names

It does not execute JavaScript Cube models. If SemCI finds `.js` model files, it reports them as unsupported.

## Classification

Breaking changes:

- removed cube, measure, dimension, or segment
- changed measure SQL, type, or filters
- changed dimension SQL or type
- removed join
- changed join relationship

Risky changes:

- added join
- changed join SQL
- changed segment SQL
- changed public titles or descriptions

Safe changes:

- added cube, measure, dimension, or segment
- added pre-aggregation
- formatting and ordering-only changes

## Roadmap

V1 is structural semantic CI. Future versions should add:

- certified query regression testing
- SQL result deltas against old and new semantic models
- agent and natural-language analytics regression tests
- adapters for LookML, dbt Semantic Layer, MetricFlow, and other semantic layers
