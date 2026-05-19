package report

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"prom-verifier/internal/config"

	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v3"
)

const (
	colorReset      = "\033[0m"
	colorBold       = "\033[1m"
	colorDim        = "\033[2m"
	colorBoldRed    = "\033[1;31m"
	colorBoldYellow = "\033[1;33m"
	colorGreen      = "\033[32m"
)

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

type Result struct {
	Group       string            `json:"group" yaml:"group"`
	Alert       string            `json:"alert" yaml:"alert"`
	Metric      string            `json:"metric" yaml:"metric"`
	Duration    time.Duration     `json:"duration" yaml:"duration"`
	State       string            `json:"state" yaml:"state"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type Reporter interface {
	StartGroup(name string)
	AddResult(rule config.Rule, metric model.Metric, duration time.Duration, state string, annotations map[string]string)
	Flush() error
}

// TextReporter streams results to stdout with colors and a final summary.
type TextReporter struct {
	useColor bool
	counts   map[string]int
	started  bool
}

func newTextReporter() *TextReporter {
	return &TextReporter{
		useColor: isTerminal(),
		counts:   map[string]int{"FIRING": 0, "PENDING": 0, "SILENT": 0},
	}
}

func (t *TextReporter) c(s, code string) string {
	if t.useColor {
		return code + s + colorReset
	}
	return s
}

func (t *TextReporter) StartGroup(name string) {
	if !t.started {
		fmt.Println("\n─────────────────── Results ───────────────────")
		t.started = true
	}
	fmt.Printf("\n▶ %s\n", t.c(name, colorBold))
}

func (t *TextReporter) AddResult(rule config.Rule, metric model.Metric, duration time.Duration, state string, annotations map[string]string) {
	t.counts[state]++
	fmt.Printf("  %s\n", t.c(rule.Alert, colorBold))
	switch state {
	case "FIRING":
		fmt.Printf("    %s  %s  duration=%s\n",
			t.c("🔥 FIRING", colorBoldRed),
			t.c(metric.String(), colorDim),
			duration.Round(time.Second))
		for k, v := range annotations {
			fmt.Printf("    %s %s\n", t.c(k+":", colorDim), v)
		}
	case "SILENT":
		fmt.Printf("    %s\n", t.c("✅ SILENT", colorGreen))
	default:
		fmt.Printf("    %s  %s  duration=%s\n",
			t.c("⚠️  PENDING", colorBoldYellow),
			t.c(metric.String(), colorDim),
			duration.Round(time.Second))
	}
}

func (t *TextReporter) Flush() error {
	total := t.counts["FIRING"] + t.counts["PENDING"] + t.counts["SILENT"]
	fmt.Println("\n───────────────────────────────────────────────")
	fmt.Printf("Summary: %d evaluated  •  %s  •  %s  •  %s\n",
		total,
		t.c(fmt.Sprintf("🔥 %d firing", t.counts["FIRING"]), colorBoldRed),
		t.c(fmt.Sprintf("⚠️  %d pending", t.counts["PENDING"]), colorBoldYellow),
		t.c(fmt.Sprintf("✅ %d silent", t.counts["SILENT"]), colorGreen),
	)
	return nil
}

// JSONReporter collects results and writes JSON on Flush.
type JSONReporter struct {
	Results      []Result
	currentGroup string
}

func (j *JSONReporter) StartGroup(name string) { j.currentGroup = name }

func (j *JSONReporter) AddResult(rule config.Rule, metric model.Metric, duration time.Duration, state string, annotations map[string]string) {
	j.Results = append(j.Results, Result{
		Group:       j.currentGroup,
		Alert:       rule.Alert,
		Metric:      metric.String(),
		Duration:    duration,
		State:       state,
		Annotations: annotations,
	})
}

func (j *JSONReporter) Flush() error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(j.Results)
}

// YAMLReporter collects results and writes YAML on Flush.
type YAMLReporter struct {
	Results      []Result
	currentGroup string
}

func (y *YAMLReporter) StartGroup(name string) { y.currentGroup = name }

func (y *YAMLReporter) AddResult(rule config.Rule, metric model.Metric, duration time.Duration, state string, annotations map[string]string) {
	y.Results = append(y.Results, Result{
		Group:       y.currentGroup,
		Alert:       rule.Alert,
		Metric:      metric.String(),
		Duration:    duration,
		State:       state,
		Annotations: annotations,
	})
}

func (y *YAMLReporter) Flush() error {
	enc := yaml.NewEncoder(os.Stdout)
	return enc.Encode(y.Results)
}

func New(format string) Reporter {
	switch format {
	case "json":
		return &JSONReporter{Results: []Result{}}
	case "yaml":
		return &YAMLReporter{Results: []Result{}}
	default:
		return newTextReporter()
	}
}
