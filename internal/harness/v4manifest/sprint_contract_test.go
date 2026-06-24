package v4manifest

import (
	"strings"
	"testing"
)

// TestDecideEvaluator_SkipsForSoloRangeTask verifies the evaluator is SKIPPED
// when the task is within the model's solo reliable range (AC-HV4-008b).
func TestDecideEvaluator_SkipsForSoloRangeTask(t *testing.T) {
	profile := TaskProfile{WithinSoloRange: true, ComplexitySignals: "single-skill generation"}
	contract := SprintContract{
		Dimensions: []string{"neutrality"},
		Thresholds: map[string]interface{}{"neutrality": 0},
	}
	d := DecideEvaluator(profile, contract)
	if d.Invoked {
		t.Fatal("DecideEvaluator invoked evaluator for solo-range task (expected skip)")
	}
	if !strings.Contains(d.Rationale, "skipped") {
		t.Fatalf("Decision rationale %q does not record the skip", d.Rationale)
	}
	if !strings.Contains(d.Rationale, "single-skill generation") {
		t.Fatalf("Decision rationale %q does not carry the complexity signals", d.Rationale)
	}
	// When skipped, Dimensions is NOT echoed (evaluator did not grade).
	if len(d.Dimensions) != 0 {
		t.Fatalf("Skipped decision echoed dimensions %v (expected empty)", d.Dimensions)
	}
}

// TestDecideEvaluator_InvokesForComplexTask verifies the evaluator is INVOKED
// when the task exceeds the model's solo range (AC-HV4-008b).
func TestDecideEvaluator_InvokesForComplexTask(t *testing.T) {
	profile := TaskProfile{WithinSoloRange: false}
	contract := SprintContract{
		Dimensions: []string{"neutrality", "coverage"},
		Thresholds: map[string]interface{}{"neutrality": 0, "coverage": 0.85},
	}
	d := DecideEvaluator(profile, contract)
	if !d.Invoked {
		t.Fatal("DecideEvaluator skipped evaluator for complex task (expected invoke)")
	}
	if !strings.Contains(d.Rationale, "invoked") {
		t.Fatalf("Decision rationale %q does not record the invocation", d.Rationale)
	}
	if !strings.Contains(d.Rationale, "skeptical adversarial reviewer") {
		t.Fatalf("Decision rationale %q does not describe skeptical-adversarial tuning", d.Rationale)
	}
	// When invoked, Dimensions ARE echoed (evaluator will grade them).
	if len(d.Dimensions) != 2 {
		t.Fatalf("Invoked decision echoed %d dimensions (expected 2)", len(d.Dimensions))
	}
}

// TestDecideEvaluator_RationaleAlwaysNonEmpty verifies the rationale is always
// populated (NFR-HV4-002 observability — a third party can audit why the
// evaluator ran or was skipped).
func TestDecideEvaluator_RationaleAlwaysNonEmpty(t *testing.T) {
	cases := []TaskProfile{
		{WithinSoloRange: true},
		{WithinSoloRange: false},
		{WithinSoloRange: true, ComplexitySignals: "bounded scope"},
		{WithinSoloRange: false, ComplexitySignals: "multi-domain"},
	}
	contract := SprintContract{Dimensions: []string{"d"}, Thresholds: map[string]interface{}{"d": 0}}
	for i, profile := range cases {
		d := DecideEvaluator(profile, contract)
		if strings.TrimSpace(d.Rationale) == "" {
			t.Fatalf("case %d: rationale empty for profile %+v", i, profile)
		}
	}
}
