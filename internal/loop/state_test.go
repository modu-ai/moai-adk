package loop

import (
	"encoding/json"
	"testing"
	"time"
)

func TestValidTransition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current LoopPhase
		next    LoopPhase
		want    bool
	}{
		{name: "analyze to implement", current: PhaseAnalyze, next: PhaseImplement, want: true},
		{name: "implement to test", current: PhaseImplement, next: PhaseTest, want: true},
		{name: "test to review", current: PhaseTest, next: PhaseReview, want: true},
		{name: "review to analyze", current: PhaseReview, next: PhaseAnalyze, want: true},
		{name: "analyze to review (skip)", current: PhaseAnalyze, next: PhaseReview, want: false},
		{name: "analyze to test (skip)", current: PhaseAnalyze, next: PhaseTest, want: false},
		{name: "implement to review (skip)", current: PhaseImplement, next: PhaseReview, want: false},
		{name: "test to analyze (skip)", current: PhaseTest, next: PhaseAnalyze, want: false},
		{name: "review to implement (skip)", current: PhaseReview, next: PhaseImplement, want: false},
		{name: "same phase analyze", current: PhaseAnalyze, next: PhaseAnalyze, want: false},
		{name: "same phase review", current: PhaseReview, next: PhaseReview, want: false},
		{name: "unknown phase", current: LoopPhase("unknown"), next: PhaseAnalyze, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ValidTransition(tt.current, tt.next)
			if got != tt.want {
				t.Errorf("ValidTransition(%q, %q) = %v, want %v", tt.current, tt.next, got, tt.want)
			}
		})
	}
}

func TestNextPhase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current LoopPhase
		want    LoopPhase
	}{
		{name: "after analyze", current: PhaseAnalyze, want: PhaseImplement},
		{name: "after implement", current: PhaseImplement, want: PhaseTest},
		{name: "after test", current: PhaseTest, want: PhaseReview},
		{name: "after review (wrap)", current: PhaseReview, want: PhaseAnalyze},
		{name: "unknown defaults to analyze", current: LoopPhase("unknown"), want: PhaseAnalyze},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NextPhase(tt.current)
			if got != tt.want {
				t.Errorf("NextPhase(%q) = %q, want %q", tt.current, got, tt.want)
			}
		})
	}
}

func TestIsValidPhase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		phase LoopPhase
		want  bool
	}{
		{name: "analyze", phase: PhaseAnalyze, want: true},
		{name: "implement", phase: PhaseImplement, want: true},
		{name: "test", phase: PhaseTest, want: true},
		{name: "review", phase: PhaseReview, want: true},
		{name: "empty", phase: LoopPhase(""), want: false},
		{name: "unknown", phase: LoopPhase("deploy"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsValidPhase(tt.phase)
			if got != tt.want {
				t.Errorf("IsValidPhase(%q) = %v, want %v", tt.phase, got, tt.want)
			}
		})
	}
}

func TestLoopState_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(time.Second)
	state := &LoopState{
		SpecID:    "SPEC-TEST-001",
		Phase:     PhaseTest,
		Iteration: 3,
		MaxIter:   5,
		Feedback: []Feedback{
			{
				Phase:        PhaseAnalyze,
				Iteration:    1,
				TestsPassed:  10,
				TestsFailed:  2,
				LintErrors:   1,
				BuildSuccess: true,
				Coverage:     78.5,
				Duration:     5 * time.Second,
				Notes:        "first iteration",
			},
		},
		StartedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Verify it is valid JSON
	if !json.Valid(data) {
		t.Fatal("marshaled data is not valid JSON")
	}

	var restored LoopState
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if restored.SpecID != state.SpecID {
		t.Errorf("SpecID = %q, want %q", restored.SpecID, state.SpecID)
	}
	if restored.Phase != state.Phase {
		t.Errorf("Phase = %q, want %q", restored.Phase, state.Phase)
	}
	if restored.Iteration != state.Iteration {
		t.Errorf("Iteration = %d, want %d", restored.Iteration, state.Iteration)
	}
	if restored.MaxIter != state.MaxIter {
		t.Errorf("MaxIter = %d, want %d", restored.MaxIter, state.MaxIter)
	}
	if len(restored.Feedback) != len(state.Feedback) {
		t.Fatalf("Feedback length = %d, want %d", len(restored.Feedback), len(state.Feedback))
	}
	if restored.Feedback[0].TestsPassed != 10 {
		t.Errorf("Feedback[0].TestsPassed = %d, want 10", restored.Feedback[0].TestsPassed)
	}
	if restored.Feedback[0].Coverage != 78.5 {
		t.Errorf("Feedback[0].Coverage = %f, want 78.5", restored.Feedback[0].Coverage)
	}
}

func TestLoopState_ToStatus(t *testing.T) {
	t.Parallel()

	t.Run("normal state", func(t *testing.T) {
		t.Parallel()
		state := &LoopState{
			SpecID:    "SPEC-TEST-001",
			Phase:     PhaseTest,
			Iteration: 2,
			MaxIter:   5,
		}

		status := state.ToStatus(true, false)

		if status.SpecID != "SPEC-TEST-001" {
			t.Errorf("SpecID = %q, want %q", status.SpecID, "SPEC-TEST-001")
		}
		if status.Phase != PhaseTest {
			t.Errorf("Phase = %q, want %q", status.Phase, PhaseTest)
		}
		if status.Iteration != 2 {
			t.Errorf("Iteration = %d, want 2", status.Iteration)
		}
		if status.MaxIter != 5 {
			t.Errorf("MaxIter = %d, want 5", status.MaxIter)
		}
		if !status.Running {
			t.Error("Running = false, want true")
		}
		if status.Converged {
			t.Error("Converged = true, want false")
		}
	})

	t.Run("converged state", func(t *testing.T) {
		t.Parallel()
		state := &LoopState{
			SpecID:    "SPEC-TEST-002",
			Phase:     PhaseReview,
			Iteration: 3,
			MaxIter:   5,
		}

		status := state.ToStatus(false, true)

		if status.Running {
			t.Error("Running = true, want false")
		}
		if !status.Converged {
			t.Error("Converged = false, want true")
		}
	})

	t.Run("nil state", func(t *testing.T) {
		t.Parallel()
		var state *LoopState
		status := state.ToStatus(false, false)

		if status.SpecID != "" {
			t.Errorf("SpecID = %q, want empty", status.SpecID)
		}
		if status.Running {
			t.Error("Running = true, want false")
		}
	})
}

func TestFeedback_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	fb := Feedback{
		Phase:        PhaseTest,
		Iteration:    2,
		TestsPassed:  42,
		TestsFailed:  3,
		LintErrors:   1,
		BuildSuccess: true,
		Coverage:     81.5,
		Duration:     10 * time.Second,
		Notes:        "test notes",
	}

	data, err := json.Marshal(fb)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var restored Feedback
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if restored.Phase != fb.Phase {
		t.Errorf("Phase = %q, want %q", restored.Phase, fb.Phase)
	}
	if restored.TestsPassed != fb.TestsPassed {
		t.Errorf("TestsPassed = %d, want %d", restored.TestsPassed, fb.TestsPassed)
	}
	if restored.Coverage != fb.Coverage {
		t.Errorf("Coverage = %f, want %f", restored.Coverage, fb.Coverage)
	}
}

func TestDecision_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	d := Decision{
		Action:    ActionConverge,
		NextPhase: PhaseAnalyze,
		Converged: true,
		Reason:    "quality gate satisfied",
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var restored Decision
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if restored.Action != d.Action {
		t.Errorf("Action = %q, want %q", restored.Action, d.Action)
	}
	if restored.Converged != d.Converged {
		t.Errorf("Converged = %v, want %v", restored.Converged, d.Converged)
	}
}

func TestPhaseOrderCycle(t *testing.T) {
	t.Parallel()

	// Verify that the full cycle works: analyze -> implement -> test -> review -> analyze
	phases := []LoopPhase{PhaseAnalyze, PhaseImplement, PhaseTest, PhaseReview}
	for i, phase := range phases {
		next := NextPhase(phase)
		expected := phases[(i+1)%len(phases)]
		if next != expected {
			t.Errorf("NextPhase(%q) = %q, want %q", phase, next, expected)
		}
	}
}
