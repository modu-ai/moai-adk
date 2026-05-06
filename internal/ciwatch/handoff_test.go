package ciwatch_test

import (
	"encoding/json"
	"testing"

	"github.com/modu-ai/moai-adk/internal/ciwatch"
)

// TestNewHandoff_SingleRequiredFailure verifies a single required check failure
// is captured in the handoff struct.
func TestNewHandoff_SingleRequiredFailure(t *testing.T) {
	state := ciwatch.CIState{
		PRNumber: 785,
		Branch:   "feat/SPEC-V3R3-CI-AUTONOMY-001-wave-2",
		RequiredFailed: []ciwatch.CheckResult{
			{
				Name:              "Lint",
				Status:            "completed",
				Conclusion:        "failure",
				RunID:             "12345678",
				LogURL:            "https://github.com/modu-ai/moai-adk/actions/runs/12345678",
				ConclusionDetail:  "golangci-lint found 2 issues",
			},
		},
		AuxiliaryFailed: nil,
		RequiredPending: nil,
		RequiredPassed:  5,
	}

	h := ciwatch.NewHandoff(state)

	if h.PRNumber != 785 {
		t.Errorf("PRNumber = %d, want 785", h.PRNumber)
	}
	if len(h.FailedChecks) != 1 {
		t.Fatalf("FailedChecks len = %d, want 1", len(h.FailedChecks))
	}
	fc := h.FailedChecks[0]
	if fc.Name != "Lint" {
		t.Errorf("FailedChecks[0].Name = %q, want %q", fc.Name, "Lint")
	}
	if fc.RunID != "12345678" {
		t.Errorf("FailedChecks[0].RunID = %q, want %q", fc.RunID, "12345678")
	}
	if fc.LogURL == "" {
		t.Error("FailedChecks[0].LogURL is empty")
	}
}

// TestNewHandoff_MultipleRequiredFailures verifies all required failures are captured.
func TestNewHandoff_MultipleRequiredFailures(t *testing.T) {
	state := ciwatch.CIState{
		PRNumber: 786,
		Branch:   "main",
		RequiredFailed: []ciwatch.CheckResult{
			{Name: "Lint", Status: "completed", Conclusion: "failure", RunID: "111", LogURL: "http://a"},
			{Name: "Test (ubuntu-latest)", Status: "completed", Conclusion: "failure", RunID: "222", LogURL: "http://b"},
			{Name: "CodeQL", Status: "completed", Conclusion: "failure", RunID: "333", LogURL: "http://c"},
		},
	}

	h := ciwatch.NewHandoff(state)

	if len(h.FailedChecks) != 3 {
		t.Fatalf("FailedChecks len = %d, want 3", len(h.FailedChecks))
	}
	names := map[string]bool{}
	for _, fc := range h.FailedChecks {
		names[fc.Name] = true
	}
	for _, want := range []string{"Lint", "Test (ubuntu-latest)", "CodeQL"} {
		if !names[want] {
			t.Errorf("missing check %q in FailedChecks", want)
		}
	}
}

// TestNewHandoff_AuxiliaryFailuresIgnored verifies auxiliary failures are NOT
// included in the handoff (they are advisory-only per AC-CIAUT-005).
func TestNewHandoff_AuxiliaryFailuresIgnored(t *testing.T) {
	state := ciwatch.CIState{
		PRNumber: 787,
		Branch:   "main",
		RequiredFailed: []ciwatch.CheckResult{
			{Name: "Lint", Status: "completed", Conclusion: "failure", RunID: "100", LogURL: "http://lint"},
		},
		AuxiliaryFailed: []ciwatch.CheckResult{
			{Name: "claude-code-review", Status: "completed", Conclusion: "failure", RunID: "999", LogURL: "http://aux"},
			{Name: "llm-panel", Status: "completed", Conclusion: "failure", RunID: "998", LogURL: "http://aux2"},
		},
	}

	h := ciwatch.NewHandoff(state)

	// Only required failures should be in the handoff.
	if len(h.FailedChecks) != 1 {
		t.Fatalf("FailedChecks len = %d, want 1 (auxiliary excluded)", len(h.FailedChecks))
	}
	if h.FailedChecks[0].Name != "Lint" {
		t.Errorf("expected Lint, got %q", h.FailedChecks[0].Name)
	}
	// AuxiliaryFailed count should be preserved for informational reporting.
	if h.AuxiliaryFailCount != 2 {
		t.Errorf("AuxiliaryFailCount = %d, want 2", h.AuxiliaryFailCount)
	}
}

// TestNewHandoff_JSONStable verifies the handoff serializes to stable JSON
// for downstream Wave 3 consumption.
func TestNewHandoff_JSONStable(t *testing.T) {
	state := ciwatch.CIState{
		PRNumber: 789,
		Branch:   "main",
		RequiredFailed: []ciwatch.CheckResult{
			{
				Name:       "Lint",
				Conclusion: "failure",
				RunID:      "42",
				LogURL:     "https://example.com/runs/42",
			},
		},
	}

	h := ciwatch.NewHandoff(state)
	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	// Round-trip unmarshal to verify shape stability.
	var got ciwatch.Handoff
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if got.PRNumber != 789 {
		t.Errorf("round-trip PRNumber = %d, want 789", got.PRNumber)
	}
	if len(got.FailedChecks) != 1 || got.FailedChecks[0].Name != "Lint" {
		t.Errorf("round-trip FailedChecks mismatch: %+v", got.FailedChecks)
	}
}
