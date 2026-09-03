package statusline

import (
	"testing"
)

// TestAC004_ArmedGoalSuppressesClearDirective asserts that when a goal is armed
// (data.GoalArmed == true) the statusline CW bar does NOT carry the soft
// (⚠️/clear) or hard (🛑/clear!) directive marker, even at a usage level that
// would normally trigger it. Unarmed (GoalArmed == false) retains the marker.
// SPEC-INFINITE-GOAL-001 REQ-3.
func TestAC004_ArmedGoalSuppressesClearDirective(t *testing.T) {
	// Build StatusData at a usage level that would normally trigger the soft
	// handoff marker (raw usage above the soft threshold for a standard window).
	soft := softAtStandard()
	data := StatusData{
		Memory: MemoryData{
			Available:         true,
			TokenBudget:       200000,
			TokensUsed:        int(float64(200000) * soft / 100 * 1.05), // 5% above soft
			ContextWindowSize: 200000,
		},
	}

	// Unarmed: marker present (backward compat).
	r := newDefaultRendererForTest()
	bar := r.renderBarsInline(&data, 10)
	if !contains(bar, "/clear") {
		t.Errorf("AC-004 unarmed: expected /clear marker at soft-stage usage, got %q", bar)
	}

	// Armed: marker suppressed.
	data.GoalArmed = true
	r2 := newDefaultRendererForTest()
	bar2 := r2.renderBarsInline(&data, 10)
	if contains(bar2, "/clear") {
		t.Errorf("AC-004 armed: /clear directive marker must be suppressed when goal armed, got %q", bar2)
	}
}

// softAtStandard returns the soft-stage threshold (%) for a standard window,
// reusing the package's own band constant so the test tracks the real band.
func softAtStandard() float64 {
	return softThresholdPct(200000) // 200000 < HandoffLargeWindowCutoff → standard band
}

// contains is a tiny strings.Contains local to avoid an extra import in this
// focused test.
func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// newDefaultRendererForTest constructs a Renderer with default theme for the
// suppression test (no color asserted; only the marker text matters).
func newDefaultRendererForTest() *Renderer {
	return NewRenderer("", false, nil)
}
