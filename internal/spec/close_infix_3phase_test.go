// close_infix_3phase_test.go — SPEC-V3R6-LIFECYCLE-REDESIGN-001 AC-LR-012 coverage.
//
// Verifies the close-infix matcher accepts BOTH the new canonical "3-phase close"
// (REQ-LR-020) AND the legacy "4-phase close" (REQ-LR-021, retained for git-history
// close commits). A doc-only rename without the matcher update would silently break
// drift close-recognition for all future closes (D4 / §B.6 of design.md).
package spec

import "testing"

// TestCloseInfixMatch_DualInfix asserts closeInfixMatch recognizes both infixes.
func TestCloseInfixMatch_DualInfix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		title string
		want  bool
	}{
		{"3-phase close (new canonical, REQ-LR-020)", "chore(spec-example-001): sync-phase audit-ready signal + 3-phase close", true},
		{"3-phase close bare", "3-phase close", true},
		{"4-phase close (legacy retained, REQ-LR-021)", "chore(spec-example-001): mx-phase audit-ready signal + 4-phase close", true},
		{"4-phase close bare (legacy)", "4-phase close", true},
		{"mx-phase audit-ready infix", "chore(spec-example-001): mx-phase audit-ready signal", true},
		{"no close infix (generic chore)", "chore(spec-example-001): backfill §e.2 commit sha", false},
		{"empty title", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := closeInfixMatch(tt.title)
			if got != tt.want {
				t.Errorf("closeInfixMatch(%q) = %v, want %v", tt.title, got, tt.want)
			}
		})
	}
}
