package codexadapter

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// TestIsDecisionBearingMatchesTheDeclaredList pins the predicate to
// DecisionBearingEvents, so the CLI fault path and the table cannot disagree.
func TestIsDecisionBearingMatchesTheDeclaredList(t *testing.T) {
	t.Parallel()

	for _, ev := range DecisionBearingEvents() {
		if !IsDecisionBearing(ev) {
			t.Errorf("%s is declared decision-bearing but IsDecisionBearing says no", ev)
		}
	}
	for _, ev := range []hook.EventType{hook.EventPostToolUse, hook.EventSessionStart, hook.EventSubagentStop} {
		if IsDecisionBearing(ev) {
			t.Errorf("%s is not decision-bearing but IsDecisionBearing says yes", ev)
		}
	}
}

// TestTranslateCodexFatalErrorIsFailClosed: a handler fault on every
// decision-bearing event renders a deny that names the cause, with or without
// a cause text, and never the empty object.
func TestTranslateCodexFatalErrorIsFailClosed(t *testing.T) {
	t.Parallel()

	for _, ev := range DecisionBearingEvents() {
		for _, cause := range []string{"handler 0: boom", ""} {
			out, discards, err := TranslateCodex(ev, DecisionFatalError, cause)
			if err != nil {
				t.Fatalf("%s/%q: %v", ev, cause, err)
			}
			s := string(out)
			if strings.TrimSpace(s) == "{}" || strings.Contains(s, `"allow"`) {
				t.Errorf("%s/%q: fault rendered %s, want a deny", ev, cause, s)
			}
			if !strings.Contains(s, "fail-closed") || !strings.Contains(s, cause) {
				t.Errorf("%s/%q: deny %s does not say it failed closed on this cause", ev, cause, s)
			}
			if len(discards) != 0 {
				t.Errorf("%s: fatal_error carries no table discard, got %v", ev, discards)
			}
		}
	}
}

// TestTranslateCodexDenyAndAdvisoryShapes covers the remaining rows the real
// path can reach: a blank-reason deny keeps a reason; a retryable error is the
// advisory no-decision object; an event without a row is refused.
func TestTranslateCodexDenyAndAdvisoryShapes(t *testing.T) {
	t.Parallel()

	out, _, err := TranslateCodex(hook.EventPreToolUse, DecisionDeny, "  ")
	if err != nil {
		t.Fatalf("blank-reason deny: %v", err)
	}
	if !strings.Contains(string(out), defaultDenyReason) {
		t.Errorf("blank-reason deny = %s, want the default reason", out)
	}

	out, _, err = TranslateCodex(hook.EventStop, DecisionRetryableError, "transient")
	if err != nil {
		t.Fatalf("retryable_error: %v", err)
	}
	if string(out) != "{}" {
		t.Errorf("retryable_error = %s, want the advisory {}", out)
	}

	if _, _, err := TranslateCodex(hook.EventPostToolUse, DecisionDeny, "x"); err == nil {
		t.Error("an event without a translation row must be refused, not rendered")
	}
}
