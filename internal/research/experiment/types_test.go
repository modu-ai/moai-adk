package experiment

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// TestExperimentState_Constants verifies that state constants hold their expected values.
func TestExperimentState_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state ExperimentState
		want  string
	}{
		{"idle", StateIdle, "idle"},
		{"baseline", StateBaseline, "baseline"},
		{"mutating", StateMutating, "mutating"},
		{"evaluating", StateEvaluating, "evaluating"},
		{"scoring", StateScoring, "scoring"},
		{"complete", StateComplete, "complete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.state) != tt.want {
				t.Errorf("ExperimentState %s = %q, want %q", tt.name, tt.state, tt.want)
			}
		})
	}
}

// TestDecision_Constants verifies that decision constants hold their expected values.
func TestDecision_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		decision Decision
		want     string
	}{
		{"keep", DecisionKeep, "keep"},
		{"discard", DecisionDiscard, "discard"},
		{"pending", DecisionPending, "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.decision) != tt.want {
				t.Errorf("Decision %s = %q, want %q", tt.name, tt.decision, tt.want)
			}
		})
	}
}

// TestChangeRecord_JSONRoundTrip verifies JSON marshal/unmarshal round-trip for ChangeRecord.
func TestChangeRecord_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := ChangeRecord{
		Type:    "modification",
		Section: "system_prompt",
		Diff:    "- old line\n+ new line",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ChangeRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded != original {
		t.Errorf("round-trip mismatch:\n  original: %+v\n  decoded:  %+v", original, decoded)
	}
}

// TestExperiment_JSONRoundTrip verifies JSON marshal/unmarshal round-trip for Experiment.
func TestExperiment_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	original := Experiment{
		ID:         "exp-001",
		Target:     "skills/moai-lang-go",
		Hypothesis: "adding Go code blocks improves accuracy",
		Change: ChangeRecord{
			Type:    "addition",
			Section: "examples",
			Diff:    "+ code block here",
		},
		Result: &eval.EvalResult{
			Overall: 0.85,
			PerCriterion: map[string]eval.CriterionResult{
				"accuracy": {Name: "accuracy", Passed: true, Weight: eval.MustPass},
			},
			MustPassOK: true,
			Timestamp:  now,
		},
		Decision:  DecisionKeep,
		Timestamp: now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Experiment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Compare key fields (time.Time may lose nanoseconds after JSON round-trip)
	if decoded.ID != original.ID {
		t.Errorf("ID: got %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Target != original.Target {
		t.Errorf("Target: got %q, want %q", decoded.Target, original.Target)
	}
	if decoded.Hypothesis != original.Hypothesis {
		t.Errorf("Hypothesis: got %q, want %q", decoded.Hypothesis, original.Hypothesis)
	}
	if decoded.Change != original.Change {
		t.Errorf("Change: got %+v, want %+v", decoded.Change, original.Change)
	}
	if decoded.Decision != original.Decision {
		t.Errorf("Decision: got %q, want %q", decoded.Decision, original.Decision)
	}
	if decoded.Result == nil {
		t.Fatal("Result is nil")
	}
	if decoded.Result.Overall != original.Result.Overall {
		t.Errorf("Result.Overall: got %f, want %f", decoded.Result.Overall, original.Result.Overall)
	}
	if decoded.Result.MustPassOK != original.Result.MustPassOK {
		t.Errorf("Result.MustPassOK: got %v, want %v", decoded.Result.MustPassOK, original.Result.MustPassOK)
	}
}

// TestChangelogEntry_JSONRoundTrip verifies JSON marshal/unmarshal round-trip for ChangelogEntry.
func TestChangelogEntry_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := ChangelogEntry{
		ExperimentID: "exp-003",
		Score:        0.92,
		Change:       "added examples to system prompt",
		Reasoning:    "prompts with examples had higher accuracy",
		Decision:     DecisionKeep,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ChangelogEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded != original {
		t.Errorf("round-trip mismatch:\n  original: %+v\n  decoded:  %+v", original, decoded)
	}
}
