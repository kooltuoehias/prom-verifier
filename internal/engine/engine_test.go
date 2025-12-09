package engine

import (
	"testing"
	"time"

	"prom-verifier/internal/config"

	"github.com/prometheus/common/model"
)

func TestCalculateAlerts(t *testing.T) {
	now := model.Now()

	tests := []struct {
		name     string
		rule     config.Rule
		matrix   model.Matrix
		expected []AlertResult
	}{
		{
			name: "Firing: Duration > For",
			rule: config.Rule{For: "5m"},
			matrix: model.Matrix{
				{
					Metric: model.Metric{"foo": "bar"},
					Values: []model.SamplePair{
						{Timestamp: now.Add(-10 * time.Minute), Value: 1},
						{Timestamp: now, Value: 1},
					},
				},
			},
			expected: []AlertResult{
				{State: StateFiring}, // We only check State mostly
			},
		},
		{
			name: "Pending: Duration < For",
			rule: config.Rule{For: "15m"},
			matrix: model.Matrix{
				{
					Metric: model.Metric{"foo": "bar"},
					Values: []model.SamplePair{
						{Timestamp: now.Add(-10 * time.Minute), Value: 1},
						{Timestamp: now, Value: 1},
					},
				},
			},
			expected: []AlertResult{
				{State: StatePending},
			},
		},
		{
			name: "Firing: For is 0 (Instant)",
			rule: config.Rule{For: "0s"},
			matrix: model.Matrix{
				{
					Metric: model.Metric{"foo": "bar"},
					Values: []model.SamplePair{
						{Timestamp: now, Value: 1}, // Single point
					},
				},
			},
			expected: []AlertResult{
				{State: StateFiring},
			},
		},
		{
			name: "Firing: No 'For' specified (Default 0)",
			rule: config.Rule{}, // Empty For
			matrix: model.Matrix{
				{
					Metric: model.Metric{"foo": "bar"},
					Values: []model.SamplePair{
						{Timestamp: now, Value: 1},
					},
				},
			},
			expected: []AlertResult{
				{State: StateFiring},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := calculateAlerts(tt.matrix, tt.rule)

			if len(results) != len(tt.expected) {
				t.Fatalf("expected %d results, got %d", len(tt.expected), len(results))
			}

			for i, res := range results {
				if res.State != tt.expected[i].State {
					t.Errorf("expected state %s, got %s", tt.expected[i].State, res.State)
				}
			}
		})
	}
}
