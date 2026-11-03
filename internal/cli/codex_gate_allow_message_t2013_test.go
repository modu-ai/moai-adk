package cli

import (
	"strings"
	"testing"
)

// TestT2013_CodexGateAllowMessageIsNotAReviewPass pins the allow message: the
// turn is allowed, the message says this is not a review PASS, and the state
// text is still carried.
func TestT2013_CodexGateAllowMessageIsNotAReviewPass(t *testing.T) {
	out := codexGateAllowed(codexStateSwitchOff)
	if out == nil {
		t.Fatal("codexGateAllowed returned nil")
	}
	want := "codex review gate: no fresh pass — turn allowed, not a review PASS (state: " + codexStateSwitchOff + ")"
	if out.SystemMessage != want {
		t.Errorf("SystemMessage = %q, want %q", out.SystemMessage, want)
	}
	if !strings.Contains(out.SystemMessage, "not a review PASS") {
		t.Errorf("message lacks %q: %q", "not a review PASS", out.SystemMessage)
	}
}
