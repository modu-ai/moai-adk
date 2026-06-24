package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestEmitAcceptEditsConfirmationAnchor covers AC-CCI-006: when the wizard
// normalizes permissionMode "acceptEdits" to empty string, an explicit
// confirmation line MUST be emitted to the output writer carrying a
// deterministic, grep-stable anchor. The anchor states the two facts
// REQ-CCI-006 requires:
//   (1) "acceptEdits" is the project default;
//   (2) settings.local.json will NOT receive a defaultMode override.
//
// This converts AC-CCI-006 from a weasel-phrase ("wizard 실행 출력 검사") into a
// binary-testable assertion (D4 pin — plan-auditor review-2 defect, resolved
// by manager-develop run-phase M3 of SPEC-V3R6-CLI-CONFIG-INTEGRITY-001).
func TestEmitAcceptEditsConfirmationAnchor(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	emitAcceptEditsConfirmation(&buf)
	out := buf.String()

	// Deterministic anchor tokens — grep-stable, unlikely to drift.
	for _, anchor := range []string{
		"acceptEdits",
		"project default",
		"settings.local.json",
	} {
		if !strings.Contains(out, anchor) {
			t.Errorf("acceptEdits confirmation line missing anchor %q; got:\n%s", anchor, out)
		}
	}
	// The line MUST state the "no override will be written" fact so the user
	// does not perceive the acceptEdits selection as a silent no-op.
	if !strings.Contains(strings.ToLower(out), "no") {
		t.Errorf("acceptEdits confirmation line must state that NO override will be written; got:\n%s", out)
	}
}

// TestAcceptEditsConfirmationEmittedOnce ensures the helper writes exactly one
// line (no duplicate prints, no trailing blank-line inflation).
func TestAcceptEditsConfirmationEmittedOnce(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	emitAcceptEditsConfirmation(&buf)
	out := buf.String()
	if out == "" {
		t.Fatal("emitAcceptEditsConfirmation wrote nothing")
	}
	// Exactly one trailing newline (single line emitted).
	if got := strings.Count(out, "\n"); got != 1 {
		t.Errorf("emitAcceptEditsConfirmation must emit exactly one line (1 trailing newline); got %d newlines:\n%s", got, out)
	}
}
