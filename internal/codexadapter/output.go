package codexadapter

import (
	"encoding/json"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// DiscardBranchCount is the number of output keys this adapter may have to
// discard. It is shared with the test that enumerates them, so adding a branch
// without a matching test case fails rather than passing silently
// (SPEC-CODEX-HOOK-ADAPTER-001 AC-REQ-3b).
const DiscardBranchCount = 3

// defaultBlockReason fills in when a hook blocks without saying why. Codex
// rejects a decision:block carrying an empty reason, so substituting a reason
// is what keeps the block a block instead of a silent no-op.
const defaultBlockReason = "blocked by a MoAI hook (no reason supplied)"

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
	if !needsMapping(payload) {
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

// needsMapping reports whether any inert key is present.
func needsMapping(payload map[string]json.RawMessage) bool {
	for k := range payload {
		if inertKeys[k] {
			return true
		}
	}
	return false
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
