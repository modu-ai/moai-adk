package codexadapter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// DiscardBranchCount is the number of output-transform branches this adapter
// carries that drop or repair output: the three inert keys, plus the PreToolUse
// permissionDecision branch (card t590). It is shared with the test that
// enumerates them, so adding a branch without a matching test case fails
// rather than passing silently (SPEC-CODEX-HOOK-ADAPTER-001 AC-REQ-3b).
const DiscardBranchCount = 4

// defaultBlockReason fills in when a hook blocks without saying why. Codex
// rejects a decision:block carrying an empty reason, so substituting a reason
// is what keeps the block a block instead of a silent no-op.
const defaultBlockReason = "blocked by a MoAI hook (no reason supplied)"

// defaultDenyReason fills in when a PreToolUse deny arrives without a usable
// reason. Codex rejects permissionDecision:deny carrying an empty
// permissionDecisionReason (output_parser.rs: deny requires a non-empty
// reason), so substituting one keeps the deny a deny (card t590).
const defaultDenyReason = "denied by a MoAI hook (no reason supplied)"

// preToolUseDropDecisions are the permissionDecision values Codex refuses on
// PreToolUse, per codex-rs hooks/src/engine/output_parser.rs
// (unsupported_pre_tool_use_hook_specific_output): allow without updatedInput
// is "unsupported permissionDecision:allow", ask is always rejected, and defer
// fails deserialization. Dropping the decision degrades the output to a
// no-opinion `{}`, which hands the choice to Codex's own approval flow — the
// harness-native reading of "no opinion". deny is deliberately absent: with a
// non-empty permissionDecisionReason it is valid and passes through
// byte-identical (card t590).
var preToolUseDropDecisions = map[string]bool{
	"allow": true,
	"ask":   true,
	"defer": true,
}

// inertKeys are the output keys Codex 0.147.0 declares but does not act on.
//
// Measured on PreToolUse and PostToolUse: a hook returning continue:false let
// the turn run to completion in both cases, and neither stopReason nor
// systemMessage appeared anywhere in the event stream or on stderr. These are
// exactly the keys MoAI's own hooks emit — team-ac-verify.sh rejects a task
// with continue:false plus stopReason, and the sync gate emits systemMessage —
// so without this mapping both would do nothing under Codex.
//
// Everything measured working is deliberately absent here and passes through
// untouched: every unnecessary translation is a drift point between the two
// harnesses.
var inertKeys = map[string]bool{
	"systemMessage": true,
	"continue":      true,
	"stopReason":    true,
}

func isDiscardableKey(key string) bool { return inertKeys[key] }

// Discard records one undeliverable message.
//
// It carries the content's LENGTH, never the content: a diagnostic that echoed
// what a hook was reporting would become an exfiltration path for it.
type Discard struct {
	Event         hook.EventType `json:"event"`
	Key           string         `json:"key"`
	ContentLength int            `json:"content_length"`
	Reason        string         `json:"reason"`
}

// additionalContextEvents are the events with a working additionalContext
// channel. Only UserPromptSubmit was measured delivering it.
var additionalContextEvents = map[hook.EventType]bool{
	hook.EventUserPromptSubmit: true,
}

// MapOutput rewrites a MoAI hook's output for Codex.
//
// Two rewrites exist. The inert-key rewrite (continue:false → decision:block
// with a filled reason, systemMessage → additionalContext where a channel
// exists) is event-blind. The PreToolUse decision rewrite (card t590) drops
// the decision shapes Codex's PreToolUse parser refuses and repairs the deny
// shape it accepts only with a reason — see normalizePreToolUseDecision.
//
// It returns the mapped payload and the list of messages that could not be
// delivered on this event. A caller that ignores the discards violates REQ-3:
// the whole point of this package is that Codex fails quietly, so the adapter
// must not answer silence with silence.
func MapOutput(event hook.EventType, raw []byte) ([]byte, []Discard, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, nil, fmt.Errorf("parse hook output: %w", err)
	}

	// Nothing to translate: hand back the original bytes rather than a
	// re-marshalled equivalent, so pass-through is byte-identical.
	if !needsMapping(payload, event) {
		return raw, nil, nil
	}

	var discards []Discard
	out := make(map[string]any, len(payload))
	for k, v := range payload {
		if !inertKeys[k] {
			var decoded any
			if err := json.Unmarshal(v, &decoded); err != nil {
				return nil, nil, fmt.Errorf("parse %q: %w", k, err)
			}
			out[k] = decoded
		}
	}

	// PreToolUse decisions need Codex-specific normalization beyond the inert
	// keys: its parser rejects allow-without-updatedInput and ask outright
	// (card t590 — the reported "invalid ... JSON output" failures).
	if event == hook.EventPreToolUse {
		drops, err := normalizePreToolUseDecision(out)
		if err != nil {
			return nil, nil, err
		}
		discards = append(discards, drops...)
	}

	if blocking, reason := blockingContinue(payload); blocking {
		out["decision"] = "block"
		out["reason"] = reason
	} else if rawContinue, ok := payload["continue"]; ok {
		// continue:true carries no blocking intent; keep it verbatim.
		var decoded any
		if err := json.Unmarshal(rawContinue, &decoded); err != nil {
			return nil, nil, fmt.Errorf("parse %q: %w", "continue", err)
		}
		out["continue"] = decoded
	}

	if rawMsg, ok := payload["systemMessage"]; ok {
		var msg string
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			return nil, nil, fmt.Errorf("parse %q: %w", "systemMessage", err)
		}
		if additionalContextEvents[event] {
			out["hookSpecificOutput"] = map[string]any{
				"hookEventName":     string(event),
				"additionalContext": msg,
			}
		} else {
			discards = append(discards, Discard{
				Event:         event,
				Key:           "systemMessage",
				ContentLength: len(msg),
				Reason:        "no delivery channel on this event",
			})
		}
	}

	mapped, err := json.Marshal(out)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal mapped output: %w", err)
	}
	return mapped, discards, nil
}

// needsMapping reports whether the payload needs any Codex-specific rewrite:
// an inert key on any event, or — on PreToolUse only — a hookSpecificOutput
// decision shape the Codex parser refuses or accepts only after repair
// (card t590). Valid shapes (deny with a reason, allow with updatedInput,
// additionalContext-only) return false so they stay byte-identical.
func needsMapping(payload map[string]json.RawMessage, event hook.EventType) bool {
	for k := range payload {
		if inertKeys[k] {
			return true
		}
	}
	if event != hook.EventPreToolUse {
		return false
	}
	raw, ok := payload["hookSpecificOutput"]
	if !ok {
		return false
	}
	var hso map[string]any
	if err := json.Unmarshal(raw, &hso); err != nil {
		return false // not an object — pass through; Codex rejects it loudly
	}
	if _, hasUpdatedInput := hso["updatedInput"]; hasUpdatedInput {
		return false
	}
	decision, _ := hso["permissionDecision"].(string)
	if preToolUseDropDecisions[decision] {
		return true
	}
	if decision == "deny" {
		reason, _ := hso["permissionDecisionReason"].(string)
		return strings.TrimSpace(reason) == ""
	}
	if decision == "" {
		_, dangling := hso["permissionDecisionReason"]
		return dangling
	}
	return false
}

// normalizePreToolUseDecision rewrites out["hookSpecificOutput"] into a shape
// the Codex PreToolUse parser accepts (card t590, contract: codex-rs
// hooks/src/engine/output_parser.rs):
//
//   - allow without updatedInput, ask, and defer are dropped entirely — each
//     is refused by Codex, and the no-opinion `{}` hands the choice to Codex's
//     own approval flow. Dropping is announced through the returned discards.
//   - deny with a blank permissionDecisionReason gains the default reason (a
//     blank-reason deny is rejected outright; the fill keeps the deny a deny).
//   - a permissionDecisionReason without a permissionDecision is dropped (a
//     dangling reason is rejected on its own).
//
// Anything else — deny with a reason, allow with updatedInput — is left
// untouched: every unnecessary translation is a drift point.
func normalizePreToolUseDecision(out map[string]any) ([]Discard, error) {
	hso, ok := out["hookSpecificOutput"].(map[string]any)
	if !ok {
		return nil, nil // absent or not an object — leave untouched
	}
	if _, hasUpdatedInput := hso["updatedInput"]; hasUpdatedInput {
		return nil, nil
	}
	decision, _ := hso["permissionDecision"].(string)
	switch {
	case preToolUseDropDecisions[decision]:
		dropped, err := json.Marshal(hso)
		if err != nil {
			return nil, fmt.Errorf("measure dropped PreToolUse decision: %w", err)
		}
		delete(out, "hookSpecificOutput")
		return []Discard{{
			Event:         hook.EventPreToolUse,
			Key:           "hookSpecificOutput",
			ContentLength: len(dropped),
			Reason:        fmt.Sprintf("permissionDecision:%s is not accepted by Codex PreToolUse; degraded to no-opinion", decision),
		}}, nil
	case decision == "deny":
		reason, _ := hso["permissionDecisionReason"].(string)
		if strings.TrimSpace(reason) == "" {
			hso["permissionDecisionReason"] = defaultDenyReason
		}
		return nil, nil
	case decision == "":
		if _, dangling := hso["permissionDecisionReason"]; dangling {
			delete(hso, "permissionDecisionReason")
			if len(hso) == 0 {
				delete(out, "hookSpecificOutput")
			}
		}
		return nil, nil
	}
	return nil, nil
}

// blockingContinue reports whether the payload carries continue:false, and the
// reason to attach. stopReason supplies it when present; otherwise the default
// keeps the reason non-empty.
func blockingContinue(payload map[string]json.RawMessage) (bool, string) {
	rawContinue, ok := payload["continue"]
	if !ok {
		return false, ""
	}
	var cont bool
	if err := json.Unmarshal(rawContinue, &cont); err != nil || cont {
		return false, ""
	}

	reason := defaultBlockReason
	if rawReason, ok := payload["stopReason"]; ok {
		var s string
		if err := json.Unmarshal(rawReason, &s); err == nil && s != "" {
			reason = s
		}
	}
	return true, reason
}
