package codexadapter

import (
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// RequiredInputUserApproval names the input a needs_input decision waits for.
// Every Codex deny converted from needs_input carries it in its reason, so the
// model reading the deny learns what is missing rather than only that the call
// was refused (REQ-HPR-008).
const RequiredInputUserApproval = "user approval"

// needsInputDiscardKey is the discard record key for a needs_input conversion.
const needsInputDiscardKey = "needs_input"

// IsDecisionBearing reports whether ev is one of DecisionBearingEvents.
func IsDecisionBearing(ev hook.EventType) bool {
	for _, d := range DecisionBearingEvents() {
		if d == ev {
			return true
		}
	}
	return false
}

// @MX:ANCHOR: [AUTO] TranslateCodex is the single Codex rendering of a normalized decision — MapOutput (PreToolUse ask/defer) and the CLI fault path both route through it
// @MX:REASON: [AUTO] fan_in=3 (normalizePreToolUseDecision, cli writeCodexFailClosed, AC-HPR-006/022 tests); a change here changes what every Codex deny, needs_input, and fault looks like to the host
// @MX:WARN: [AUTO] fail-closed boundary — returning the no-opinion {} or an allow for deny / needs_input / fatal_error lets Codex proceed under a non-prompting approval policy (REQ-HPR-007)
// @MX:REASON: [AUTO] card t590 degraded ask to {} for parser compatibility; that loosening is exactly what this function exists to prevent (operator decision Q2)
// @MX:SPEC: SPEC-DUAL-HARNESS-HOOK-PARITY-001
// TranslateCodex renders normalized decision d on event ev for Codex through
// the translation table (Lookup + Render). It returns the hook stdout bytes and
// the discard records the conversion requires — one record for a needs_input,
// which Codex cannot express and receives as a fail-closed deny (design.md §D5).
//
// reason is the handler's own reason. For needs_input and fatal_error it is
// wrapped with what happened; for deny a blank reason gets the default, since
// Codex rejects a blank-reason deny and the deny would become a no-op.
func TranslateCodex(ev hook.EventType, d Decision, reason string) ([]byte, []Discard, error) {
	row, ok := Lookup(HarnessCodex, ev, d)
	if !ok {
		return nil, nil, fmt.Errorf("translate %s on %s: no Codex translation row", d, ev)
	}

	reason = strings.TrimSpace(reason)
	text := reason
	switch d {
	case DecisionNeedsInput:
		text = fmt.Sprintf("%s required: Codex hooks cannot ask for it, so MoAI denied this %s fail-closed", RequiredInputUserApproval, ev)
		if reason != "" {
			text += " — " + reason
		}
	case DecisionFatalError:
		text = fmt.Sprintf("MoAI %s hook failed, so the call was denied fail-closed", ev)
		if reason != "" {
			text += ": " + reason
		}
	case DecisionDeny:
		if text == "" {
			text = defaultDenyReason
		}
	}

	out, err := Render(ev, row.Outcome, text)
	if err != nil {
		return nil, nil, fmt.Errorf("translate %s on %s: %w", d, ev, err)
	}

	var discards []Discard
	if row.DiscardRecorded {
		discards = append(discards, Discard{
			Event:         ev,
			Key:           needsInputDiscardKey,
			ContentLength: len(reason),
			Reason:        fmt.Sprintf("%s converted to a fail-closed %s: Codex cannot ask for %s on this event (UNSUPPORTED-native)", d, row.Outcome, RequiredInputUserApproval),
		})
	}
	return out, discards, nil
}
