package eval

import "testing"

// TestComputeResult table-driven tests for the ComputeResult function.
func TestComputeResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		criteria     []EvalCriterion
		results      map[string]bool
		wantOverall  float64
		wantMustPass bool
	}{
		{
			name: "all pass → Overall=1.0, MustPassOK=true",
			criteria: []EvalCriterion{
				{Name: "a", Weight: MustPass},
				{Name: "b", Weight: NiceToHave},
			},
			results:      map[string]bool{"a": true, "b": true},
			wantOverall:  1.0,
			wantMustPass: true,
		},
		{
			name: "all fail → Overall=0.0, MustPassOK=false",
			criteria: []EvalCriterion{
				{Name: "a", Weight: MustPass},
				{Name: "b", Weight: NiceToHave},
			},
			results:      map[string]bool{"a": false, "b": false},
			wantOverall:  0.0,
			wantMustPass: false,
		},
		{
			name: "must_pass pass, nice_to_have fail → MustPassOK=true, Overall < 1.0",
			criteria: []EvalCriterion{
				{Name: "a", Weight: MustPass},
				{Name: "b", Weight: NiceToHave},
			},
			results:      map[string]bool{"a": true, "b": false},
			wantOverall:  0.5,
			wantMustPass: true,
		},
		{
			name: "must_pass fail → MustPassOK=false (regardless of overall)",
			criteria: []EvalCriterion{
				{Name: "a", Weight: MustPass},
				{Name: "b", Weight: NiceToHave},
				{Name: "c", Weight: NiceToHave},
			},
			results:      map[string]bool{"a": false, "b": true, "c": true},
			wantOverall:  2.0 / 3.0,
			wantMustPass: false,
		},
		{
			name:         "empty results → Overall=0.0",
			criteria:     []EvalCriterion{},
			results:      map[string]bool{},
			wantOverall:  0.0,
			wantMustPass: true, // true when no must_pass criteria exist
		},
		{
			name: "only must_pass criteria and all pass",
			criteria: []EvalCriterion{
				{Name: "a", Weight: MustPass},
				{Name: "b", Weight: MustPass},
			},
			results:      map[string]bool{"a": true, "b": true},
			wantOverall:  1.0,
			wantMustPass: true,
		},
		{
			name: "criteria missing from results are treated as failed",
			criteria: []EvalCriterion{
				{Name: "a", Weight: MustPass},
				{Name: "b", Weight: NiceToHave},
			},
			results:      map[string]bool{"a": true},
			wantOverall:  0.5,
			wantMustPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ComputeResult(tt.criteria, tt.results)

			if result == nil {
				t.Fatal("ComputeResult() returned nil")
				return // guard against staticcheck SA5011
			}

			// float64 comparison with tolerance 0.001
			if diff := result.Overall - tt.wantOverall; diff > 0.001 || diff < -0.001 {
				t.Errorf("Overall = %f, want %f", result.Overall, tt.wantOverall)
			}

			if result.MustPassOK != tt.wantMustPass {
				t.Errorf("MustPassOK = %v, want %v", result.MustPassOK, tt.wantMustPass)
			}

			// Verify PerCriterion map size
			if len(result.PerCriterion) != len(tt.criteria) {
				t.Errorf("PerCriterion size = %d, want %d", len(result.PerCriterion), len(tt.criteria))
			}

			// Verify Timestamp is non-zero
			if result.Timestamp.IsZero() {
				t.Error("Timestamp is zero value")
			}
		})
	}
}
