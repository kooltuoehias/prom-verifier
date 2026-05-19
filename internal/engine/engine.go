package engine

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"
	"time"

	"prom-verifier/internal/config"
	"prom-verifier/internal/report"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// AlertState represents the state of an alert (Pending or Firing).
type AlertState string

const (
	StateFiring  AlertState = "FIRING"
	StatePending AlertState = "PENDING"
	StateSilent  AlertState = "SILENT"
)

// AlertResult holds the evaluation result for a single time series.
type AlertResult struct {
	Metric      model.Metric
	Duration    time.Duration
	State       AlertState
	Annotations map[string]string // Added
}

// Run executes the evaluation of rules against the Prometheus API.
func Run(ctx context.Context, v1api v1.API, cfg *config.Config, rep report.Reporter) {
	for _, group := range cfg.RuleFile.Groups {
		rep.StartGroup(group.Name)
		for _, rule := range group.Rules {
			evaluateRule(ctx, v1api, rule, cfg.Start, cfg.End, rep)
		}
	}
	rep.Flush()
}

func evaluateRule(ctx context.Context, v1api v1.API, rule config.Rule, start, end time.Time, rep report.Reporter) {
	r := v1.Range{Start: start, End: end, Step: 1 * time.Minute}
	result, _, err := v1api.QueryRange(ctx, rule.Expr, r)

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Query Error: %v\n", err)
		return
	}

	matrix, ok := result.(model.Matrix)
	if !ok {
		return
	}

	if len(matrix) == 0 {
		rep.AddResult(rule, model.Metric{}, 0, string(StateSilent), nil)
		return
	}

	results := calculateAlerts(matrix, rule)

	for _, res := range results {
		rep.AddResult(rule, res.Metric, res.Duration, string(res.State), res.Annotations)
	}
}

func calculateAlerts(matrix model.Matrix, rule config.Rule) []AlertResult {
	var forDuration time.Duration
	if rule.For != "" {
		var err error
		forDuration, err = time.ParseDuration(rule.For)
		if err != nil {
			forDuration = 0
		}
	}

	var results []AlertResult

	for _, stream := range matrix {
		if len(stream.Values) < 2 {
			state := StatePending
			if forDuration == 0 {
				state = StateFiring
			}

			// If firing (instant), render annotations
			var annotations map[string]string
			if state == StateFiring {
				annotations = renderAnnotations(rule.Annotations, stream.Metric, stream.Values[len(stream.Values)-1].Value)
			}

			results = append(results, AlertResult{
				Metric:      stream.Metric,
				Duration:    0,
				State:       state,
				Annotations: annotations,
			})
			continue
		}

		firstTime := stream.Values[0].Timestamp.Time()
		lastTime := stream.Values[len(stream.Values)-1].Timestamp.Time()
		duration := lastTime.Sub(firstTime)

		state := StatePending
		if duration >= forDuration {
			state = StateFiring
		}

		var annotations map[string]string
		if state == StateFiring {
			annotations = renderAnnotations(rule.Annotations, stream.Metric, stream.Values[len(stream.Values)-1].Value)
		}

		results = append(results, AlertResult{
			Metric:      stream.Metric,
			Duration:    duration,
			State:       state,
			Annotations: annotations,
		})
	}
	return results
}

// renderAnnotations templates annotation values using Prometheus-style $labels/$value variables.
func renderAnnotations(rawAnnotations map[string]string, labels model.Metric, value model.SampleValue) map[string]string {
	rendered := make(map[string]string)

	data := struct {
		Labels map[string]string
		Value  float64
	}{
		Labels: make(map[string]string),
		Value:  float64(value),
	}
	for k, v := range labels {
		data.Labels[string(k)] = string(v)
	}

	// Prepend variable bindings so templates can use {{ $labels.foo }} and {{ $value }}.
	helper := "{{ $labels := .Labels }}{{ $value := .Value }}"

	for k, v := range rawAnnotations {
		tmpl, err := template.New("validity").Parse(helper + v)
		if err != nil {
			rendered[k] = fmt.Sprintf("<template_error: %v>", err)
			continue
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			rendered[k] = fmt.Sprintf("<render_error: %v>", err)
			continue
		}
		rendered[k] = buf.String()
	}
	return rendered
}
