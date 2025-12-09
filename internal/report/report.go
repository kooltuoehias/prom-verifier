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

// AlertState matches engine definition to avoid circular import if possible,
// or we can move types to a 'types' package. For now, we will duplicate or import if we move engine types.
// Better approach: Let's assume engine imports report, or report imports engine types?
// Engine depends on report to call AddResult. So report cannot depend on engine.
// We should define types in report or a shared package.
// For simplicity, let's redefine necessary types here or make them generic interface{} but that's ugly.
// Let's creating a simple 'types' package is overkill.
// We will accept basic types here to decouple.

type Result struct {
	RuleName    string            `json:"rule_name" yaml:"rule_name"`
	Alert       string            `json:"alert" yaml:"alert"`
	Metric      string            `json:"metric" yaml:"metric"`
	Duration    time.Duration     `json:"duration" yaml:"duration"`
	State       string            `json:"state" yaml:"state"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"` // Added
}

type Reporter interface {
	AddResult(rule config.Rule, metric model.Metric, duration time.Duration, state string, annotations map[string]string)
	Flush() error
}

// TextReporter prints to stdout immediately (Legacy behavior)
type TextReporter struct{}

func (t *TextReporter) AddResult(rule config.Rule, metric model.Metric, duration time.Duration, state string, annotations map[string]string) {
	// We can try to mimic the old output grouping, but for now linear output is fine
	// actually the old one grouped by rule.
	// The engine loop drives this. If we print immediately, we get the same behavior.
	if state == "FIRING" {
		fmt.Printf("      🔥 FIRING! [%s] (Duration: %s)\n", metric, duration)
		// Print annotations
		for k, v := range annotations {
			fmt.Printf("          - %s: %s\n", k, v)
		}
	} else if state == "SILENT" {
		fmt.Println("      ✅ Status: SILENT")
	} else {
		fmt.Printf("      ⚠️ PENDING... [%s] (Duration: %s)\n", metric, duration)
	}
}

func (t *TextReporter) Flush() error {
	return nil
}

// JSONReporter collects results and prints JSON at the end
type JSONReporter struct {
	Results []Result
}

func (j *JSONReporter) AddResult(rule config.Rule, metric model.Metric, duration time.Duration, state string, annotations map[string]string) {
	j.Results = append(j.Results, Result{
		RuleName:    rule.Alert, // Or Group Name? Rule has Alert field.
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

// YAMLReporter collects results and prints YAML at the end
type YAMLReporter struct {
	Results []Result
}

func (y *YAMLReporter) AddResult(rule config.Rule, metric model.Metric, duration time.Duration, state string, annotations map[string]string) {
	y.Results = append(y.Results, Result{
		RuleName:    rule.Alert,
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

// Factory
func New(format string) Reporter {
	switch format {
	case "json":
		return &JSONReporter{Results: []Result{}}
	case "yaml":
		return &YAMLReporter{Results: []Result{}}
	default:
		return &TextReporter{}
	}
}
