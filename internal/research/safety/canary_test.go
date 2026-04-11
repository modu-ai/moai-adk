package safety

import (
	"testing"
	"time"
)

// TestCanaryChecker_Check verifies the canary regression check logic.
func TestCanaryChecker_Check(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		baselines []Baseline
		proposed  float64
		threshold float64
		want      bool
		wantErr   bool
	}{
		{
			name:      "no baselines → true (nothing to compare against)",
			baselines: nil,
			proposed:  0.50,
			threshold: 0.10,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "empty baseline slice → true",
			baselines: []Baseline{},
			proposed:  0.50,
			threshold: 0.10,
			want:      true,
			wantErr:   false,
		},
		{
			name: "within threshold: 0.85 → 0.80, threshold=0.10 → true",
			baselines: []Baseline{
				{Target: "test", Score: 0.85, Timestamp: now},
			},
			proposed:  0.80,
			threshold: 0.10,
			want:      true,
			wantErr:   false,
		},
		{
			name: "exceeds threshold: 0.85 → 0.70, threshold=0.10 → false",
			baselines: []Baseline{
				{Target: "test", Score: 0.85, Timestamp: now},
			},
			proposed:  0.70,
			threshold: 0.10,
			want:      false,
			wantErr:   false,
		},
		{
			name: "exactly at threshold: 0.85 → 0.75, threshold=0.10 → true (boundary value)",
			baselines: []Baseline{
				{Target: "test", Score: 0.85, Timestamp: now},
			},
			proposed:  0.75,
			threshold: 0.10,
			want:      true,
			wantErr:   false,
		},
		{
			name: "multiple baselines, one triggers regression → false",
			baselines: []Baseline{
				{Target: "a", Score: 0.80, Timestamp: now},
				{Target: "b", Score: 0.90, Timestamp: now},
			},
			proposed:  0.75,
			threshold: 0.10,
			want:      false, // b: 0.90 - 0.75 = 0.15 > 0.10
			wantErr:   false,
		},
		{
			name: "multiple baselines, all within threshold → true",
			baselines: []Baseline{
				{Target: "a", Score: 0.80, Timestamp: now},
				{Target: "b", Score: 0.82, Timestamp: now},
			},
			proposed:  0.75,
			threshold: 0.10,
			want:      true, // a: 0.05, b: 0.07 — both <= 0.10
			wantErr:   false,
		},
		{
			name: "proposed score higher than baseline → true (improvement)",
			baselines: []Baseline{
				{Target: "test", Score: 0.70, Timestamp: now},
			},
			proposed:  0.85,
			threshold: 0.10,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "threshold 0 → error",
			baselines: []Baseline{{Target: "t", Score: 0.80, Timestamp: now}},
			proposed:  0.75,
			threshold: 0.0,
			want:      false,
			wantErr:   true,
		},
		{
			name:      "negative threshold → error",
			baselines: []Baseline{{Target: "t", Score: 0.80, Timestamp: now}},
			proposed:  0.75,
			threshold: -0.05,
			want:      false,
			wantErr:   true,
		},
	}

	checker := NewCanaryChecker()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.Check(tt.baselines, tt.proposed, tt.threshold)
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Check() = %v, want %v", got, tt.want)
			}
		})
	}
}
