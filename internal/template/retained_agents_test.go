package template

import "testing"

// TestRetainedAgents_IsTheModelFreeRosterSSOT pins the retained-agent roster.
// The roster is the last surviving model-free list after the per-agent profile
// matrix was deleted (SPEC-AGENT-MODEL-INHERIT-001 M5); the matrix display
// order it used to mirror is gone, so the expectation below is the literal
// roster itself — read from the accessor to keep this a round-trip pin rather
// than a second source of truth.
func TestRetainedAgents_IsTheModelFreeRosterSSOT(t *testing.T) {
	got := RetainedAgents()
	want := []string{
		"manager-spec",
		"plan-auditor",
		"sync-auditor",
		"manager-develop",
		"super-advisor",
		"mission-governor",
		"manager-design",
		"manager-lead",
		"builder-harness",
		"e2e-tester",
		"manager-docs",
		"manager-git",
		"Explore",
	}
	if len(got) != len(want) {
		t.Fatalf("RetainedAgents() has %d entries %v, want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("RetainedAgents()[%d] = %q, want %q (full: %v)", i, got[i], want[i], got)
		}
	}

	got[0] = "mutated"
	if RetainedAgents()[0] == "mutated" {
		t.Error("RetainedAgents() returned the backing slice, not a copy")
	}
}
