package report

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"prom-verifier/internal/config"

	"github.com/prometheus/common/model"
)

var testRule = config.Rule{Alert: "HighCPU"}
var testMetric = model.Metric{"instance": "host1"}

func TestNew_ReturnsCorrectType(t *testing.T) {
	if _, ok := New("json").(*JSONReporter); !ok {
		t.Error("New(json) should return *JSONReporter")
	}
	if _, ok := New("yaml").(*YAMLReporter); !ok {
		t.Error("New(yaml) should return *YAMLReporter")
	}
	if _, ok := New("text").(*TextReporter); !ok {
		t.Error("New(text) should return *TextReporter")
	}
	if _, ok := New("").(*TextReporter); !ok {
		t.Error("New(unknown) should default to *TextReporter")
	}
}

func TestTextReporter_NoPanic(t *testing.T) {
	r := newTextReporter()
	r.StartGroup("infra")
	r.AddResult(testRule, testMetric, 10*time.Minute, "FIRING", map[string]string{"summary": "CPU high"})
	r.AddResult(testRule, model.Metric{}, 3*time.Minute, "PENDING", nil)
	r.AddResult(testRule, model.Metric{}, 0, "SILENT", nil)
	if err := r.Flush(); err != nil {
		t.Errorf("unexpected flush error: %v", err)
	}
}

func TestTextReporter_CountsSummary(t *testing.T) {
	r := newTextReporter()
	r.StartGroup("g")
	r.AddResult(testRule, testMetric, 10*time.Minute, "FIRING", nil)
	r.AddResult(testRule, testMetric, 2*time.Minute, "PENDING", nil)
	r.AddResult(testRule, model.Metric{}, 0, "SILENT", nil)

	if r.counts["FIRING"] != 1 {
		t.Errorf("expected 1 FIRING, got %d", r.counts["FIRING"])
	}
	if r.counts["PENDING"] != 1 {
		t.Errorf("expected 1 PENDING, got %d", r.counts["PENDING"])
	}
	if r.counts["SILENT"] != 1 {
		t.Errorf("expected 1 SILENT, got %d", r.counts["SILENT"])
	}
}

func TestJSONReporter_GroupAndFields(t *testing.T) {
	jr := &JSONReporter{}
	jr.StartGroup("infra")
	jr.AddResult(testRule, testMetric, 5*time.Minute, "FIRING", map[string]string{"summary": "high"})
	jr.StartGroup("app")
	jr.AddResult(testRule, testMetric, 1*time.Minute, "PENDING", nil)

	if len(jr.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(jr.Results))
	}
	if jr.Results[0].Group != "infra" {
		t.Errorf("expected group 'infra', got %q", jr.Results[0].Group)
	}
	if jr.Results[1].Group != "app" {
		t.Errorf("expected group 'app', got %q", jr.Results[1].Group)
	}
	if jr.Results[0].Alert != "HighCPU" {
		t.Errorf("expected alert HighCPU, got %s", jr.Results[0].Alert)
	}
}

func TestJSONReporter_ValidJSON(t *testing.T) {
	jr := &JSONReporter{}
	jr.StartGroup("infra")
	jr.AddResult(testRule, testMetric, 5*time.Minute, "FIRING", nil)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(jr.Results); err != nil {
		t.Fatalf("failed to encode JSON: %v", err)
	}
	if !json.Valid(buf.Bytes()) {
		t.Error("output is not valid JSON")
	}
}
