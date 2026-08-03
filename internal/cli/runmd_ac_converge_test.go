package cli

import (
	"os"
	"strings"
	"testing"
)

// TestAC009_AcConvergeDocCorrectedToActual30Turns asserts the run.md ac_converge
// block carries a corrected statement stating the ACTUAL 30-turn execution AND
// referencing the unparseable "Max 20 turns" clause. OQ-1 RESOLVED to option (b)
// doc-only: parseCondition matches ONLY `exits <N>`, so "Max 20 turns" is NOT
// parsed and the goal runs at DefaultMaxTurns=30. SPEC-INFINITE-GOAL-001 REQ-7.
//
// A bare "30" without the explanatory reference is insufficient (the AC proves
// the doc explains WHY the actual execution is 30, not merely asserts it).
func TestAC009_AcConvergeDocCorrectedToActual30Turns(t *testing.T) {
	path := "../../.claude/skills/moai/workflows/run.md"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	content := string(data)

	// The corrected statement must carry BOTH "30" AND a reference to the
	// unparseable "Max 20 turns" clause. Look in the ac_converge region.
	acRegion := extractAcConvergeRegion(content)
	if acRegion == "" {
		t.Fatalf("AC-009: could not locate ac_converge region in run.md")
	}
	has30 := strings.Contains(acRegion, "30")
	hasUnparseableRef := strings.Contains(acRegion, "Max 20 turns") &&
		(strings.Contains(strings.ToLower(acRegion), "not parsed") ||
			strings.Contains(strings.ToLower(acRegion), "unparseable") ||
			strings.Contains(strings.ToLower(acRegion), "not parseable") ||
			strings.Contains(strings.ToLower(acRegion), "does not parse"))
	if !has30 {
		t.Errorf("AC-009: ac_converge region must state the actual 30-turn execution; region=%q", acRegion)
	}
	if !hasUnparseableRef {
		t.Errorf("AC-009: ac_converge region must reference the unparseable 'Max 20 turns' clause (explaining WHY the actual execution is 30); region=%q", acRegion)
	}
}

// extractAcConvergeRegion returns the run.md text around the ac_converge goal
// condition block (the fenced ```text block that contains "Max ... turns").
func extractAcConvergeRegion(content string) string {
	lines := strings.Split(content, "\n")
	start := -1
	for i, l := range lines {
		if strings.Contains(l, "Max") && strings.Contains(l, "turns") {
			start = i
			break // first match is the ac_converge block region
		}
	}
	if start == -1 {
		return ""
	}
	// Grab a window around the first match (the ac_converge block).
	lo := start - 20
	if lo < 0 {
		lo = 0
	}
	hi := start + 20
	if hi > len(lines) {
		hi = len(lines)
	}
	return strings.Join(lines[lo:hi], "\n")
}
