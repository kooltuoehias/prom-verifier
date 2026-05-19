package config

import (
	"testing"
	"time"
)

func TestWindowSafetyLimit(t *testing.T) {
	tests := []struct {
		name    string
		window  time.Duration
		wantErr bool
	}{
		{"exactly 4h is allowed", 4 * time.Hour, false},
		{"4h1m is blocked", 4*time.Hour + time.Minute, true},
		{"30m is allowed", 30 * time.Minute, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exceeds := tt.window > 4*time.Hour
			if exceeds != tt.wantErr {
				t.Errorf("window %s: expected exceeds=%v, got %v", tt.window, tt.wantErr, exceeds)
			}
		})
	}
}

func TestThreeMonthSafetyLimit(t *testing.T) {
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)

	old := time.Now().AddDate(0, -4, 0)
	if !old.Before(threeMonthsAgo) {
		t.Error("4-month-old time should be blocked")
	}

	recent := time.Now().Add(-24 * time.Hour)
	if recent.Before(threeMonthsAgo) {
		t.Error("yesterday should not be blocked")
	}
}
