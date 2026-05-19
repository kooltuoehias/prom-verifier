# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o prom-verifier ./cmd/prom-verifier

# Run tests
go test ./...

# Run a single test
go test ./internal/engine/ -run TestCalculateAlerts

# Run with coverage
go test ./... -coverprofile=coverage.out

# Run the tool (requires a live Prometheus instance)
./prom-verifier -url "http://localhost:9090" -file alert.yaml -at "2024-01-15 14:30" -window 15m
```

## Architecture

The tool follows a linear pipeline: **Config → Client → Engine → Report**.

```
cmd/prom-verifier/main.go   # Entry point; wires the four packages together
internal/config/            # Flag parsing, safety validation, YAML rule file loading
internal/client/            # Thin wrapper around prometheus/client_golang v1 API
internal/engine/            # Core alert evaluation logic (the only testable package)
internal/report/            # Output formatting (text/json/yaml); Reporter interface
```

### Key design points

**`internal/config`** enforces two hard safety limits before any network call: window size ≤ 4h, and query start time no older than 3 months. The `RuleFile` / `RuleGroup` / `Rule` structs in `rules.go` mirror the Prometheus alerting YAML schema directly.

**`internal/engine`** is the domain core. `calculateAlerts` compares the observed metric duration against the rule's `for` field to classify each time series as `FIRING`, `PENDING`, or `SILENT`. `renderAnnotations` uses Go's `text/template` with a `$labels`/`$value` preamble to replicate Prometheus annotation templating.

**`internal/report`** uses the `Reporter` interface to decouple engine from output format. `TextReporter` streams results immediately; `JSONReporter` and `YAMLReporter` buffer all results and flush at the end via `Flush()`.

**Dependency direction**: `engine` imports `report` (to call `AddResult`), so `report` must not import `engine`. Alert state strings (`"FIRING"`, `"PENDING"`, `"SILENT"`) are passed as plain strings across the boundary to avoid a circular import.

### Alert state logic

A query range is fetched at 1-minute step resolution. For each matched time series:
- If no data → `SILENT`
- If single data point and `for == 0` → `FIRING`
- If `lastTimestamp - firstTimestamp >= for` → `FIRING`
- Otherwise → `PENDING`

Annotations are only rendered for `FIRING` alerts.
