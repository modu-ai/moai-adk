package template

import (
	"reflect"
	"testing"
)

// TestRetainedAgents_IsTheModelFreeRosterSSOT pins the retained-agent roster
// that outlives the per-agent profile matrix: it names the same agents in the
// same order as the matrix display order (the literal it replaces as the
// canonical source), and the accessor hands out a defensive copy.
func TestRetainedAgents_IsTheModelFreeRosterSSOT(t *testing.T) {
	got := RetainedAgents()
	if !reflect.DeepEqual(got, ProfileMatrixAgents()) {
		t.Fatalf("RetainedAgents() = %v, want the matrix display order %v", got, ProfileMatrixAgents())
	}
	seen := map[string]bool{}
	for _, n := range got {
		if n == "" || seen[n] {
			t.Errorf("roster entry %q is empty or duplicated", n)
		}
		seen[n] = true
	}
	if !seen["Explore"] || !seen["manager-lead"] {
		t.Errorf("roster %v lacks the built-in Explore or manager-lead", got)
	}

	got[0] = "mutated"
	if RetainedAgents()[0] == "mutated" {
		t.Error("RetainedAgents() returned the backing slice, not a copy")
	}
}
