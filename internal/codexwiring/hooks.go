package codexwiring

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// moaiHooksDescription is the top-level description set ONLY on first creation;
// an existing (user) description is preserved verbatim (plan D1).
const moaiHooksDescription = "MoAI-managed hook layer. MoAI refreshes only its own handlers (command prefix 'moai hook '); every other entry is user-owned and preserved."

// handlerJSON is one emitted handler, restricted to the measured-whitelist
// hook-level keys {type, command, timeout} (spec §A.5 — the t83 whitelist's
// subset is what both Codex versions accept).
type handlerJSON struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
}

// entryJSON is one emitted event entry. matcher is omitted — a matcher-less
// entry was measured firing in t83 and MoAI hooks are event-scoped already.
type entryJSON struct {
	Matcher string        `json:"matcher,omitempty"`
	Hooks   []handlerJSON `json:"hooks"`
}

// moaiHandlerTimeout returns the table timeout for one event (plan D1):
// SessionEnd is capped at Codex's documented ceiling, everything else carries
// the table default.
func moaiHandlerTimeout(event hook.EventType) int {
	if event == hook.EventSessionEnd {
		return sessionEndTimeoutCeiling
	}
	return defaultHandlerTimeout
}

// RenderHooks renders the complete project hooks.json content, merging the
// current MoAI table (derived from codexadapter.EventTable's adapted rows)
// into an existing document.
//
// Merge model (plan D2 / REQ-CW-005): stale MoAI handlers (command prefix
// "moai hook ") are removed from every entry, the current table is appended
// per event, and all user bytes — entries, event keys, the top-level
// description — are carried through verbatim as json.RawMessage. User entries
// that shared an array with a removed MoAI handler are re-marshalled with only
// their surviving handlers; an entry left with no handlers is dropped, and an
// event key left with no entries is dropped with it.
//
// RenderHooks does NOT apply the whitelist gate — the gate binds the write,
// not the render (REQ-CW-003), so callers can inspect and refuse. Unparseable
// existing input is an error; the caller warns and leaves the file untouched.
func RenderHooks(existing []byte) ([]byte, error) {
	// The output document starts as the user's own top-level keys carried
	// verbatim (json.RawMessage) — unknown user keys are NEVER silently
	// dropped: preserving them means the whitelist gate at write time can
	// refuse the whole file (REQ-CW-003) instead of the merge quietly
	// deleting user data.
	doc := make(map[string]any)
	entriesByEvent := make(map[string][]json.RawMessage)

	if len(existing) > 0 {
		var top map[string]json.RawMessage
		if err := json.Unmarshal(existing, &top); err != nil {
			return nil, fmt.Errorf("parse existing hooks.json: %w", err)
		}
		for key, raw := range top {
			if key == "hooks" {
				continue // replaced by the merged table below
			}
			doc[key] = raw
		}

		if rawHooks, ok := top["hooks"]; ok {
			var byEvent map[string][]json.RawMessage
			if err := json.Unmarshal(rawHooks, &byEvent); err != nil {
				return nil, fmt.Errorf("parse existing hooks block: %w", err)
			}
			for event, entries := range byEvent {
				kept, err := stripMoAIHandlers(entries)
				if err != nil {
					return nil, fmt.Errorf("existing %s entry: %w", event, err)
				}
				if len(kept) > 0 {
					entriesByEvent[event] = kept
				}
			}
		}
	}

	// Append the current table (D2: append AFTER user entries so MoAI's own
	// handlers are last within each event).
	for _, row := range codexadapter.EventTable {
		if !row.Adapted {
			continue
		}
		entry := entryJSON{
			Hooks: []handlerJSON{{
				Type:    "command",
				Command: "moai hook " + row.DispatcherArg + harnessCodexSuffix,
				Timeout: moaiHandlerTimeout(row.CodexEvent),
			}},
		}
		raw, err := json.Marshal(entry)
		if err != nil {
			return nil, fmt.Errorf("marshal %s entry: %w", row.CodexEvent, err)
		}
		entriesByEvent[string(row.CodexEvent)] = append(entriesByEvent[string(row.CodexEvent)], raw)
	}

	if _, hasDescription := doc["description"]; !hasDescription {
		doc["description"] = json.RawMessage(mustQuote(moaiHooksDescription))
	}
	doc["hooks"] = entriesByEvent

	// Map-key marshalling is sorted by encoding/json, so the rendered bytes
	// are deterministic without an explicit ordering pass (REQ-CW-006).
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("render hooks.json: %w", err)
	}
	return append(out, '\n'), nil
}

// stripMoAIHandlers removes every handler whose command carries the MoAI
// prefix from a list of raw entries, returning the surviving entries. Entries
// with no MoAI handlers pass through byte-identical; entries that lost
// handlers (but kept user ones) are re-marshalled.
func stripMoAIHandlers(entries []json.RawMessage) ([]json.RawMessage, error) {
	var kept []json.RawMessage
	for _, raw := range entries {
		var entry struct {
			Matcher string            `json:"matcher"`
			Hooks   []json.RawMessage `json:"hooks"`
		}
		if err := json.Unmarshal(raw, &entry); err != nil {
			return nil, fmt.Errorf("unparseable entry: %w", err)
		}

		var userHandlers []json.RawMessage
		for _, h := range entry.Hooks {
			var cmd struct {
				Command string `json:"command"`
			}
			if err := json.Unmarshal(h, &cmd); err != nil {
				return nil, fmt.Errorf("unparseable handler: %w", err)
			}
			if strings.HasPrefix(cmd.Command, moaiHandlerPrefix) {
				continue // stale MoAI handler — replaced by the current table
			}
			userHandlers = append(userHandlers, h)
		}

		if len(entry.Hooks) == 0 {
			// An entry with no handlers at all (or a malformed non-array one
			// that unmarshalled to nil) carries nothing MoAI manages; keep it
			// verbatim so user oddities are never silently dropped.
			kept = append(kept, raw)
			continue
		}
		if len(userHandlers) == len(entry.Hooks) {
			kept = append(kept, raw) // untouched user entry — byte-preserved
			continue
		}
		if len(userHandlers) == 0 {
			continue // entry held only MoAI handlers — drop the whole entry
		}
		// Mixed entry: re-marshal with only the user handlers.
		remarshalled, err := marshalEntry(entry.Matcher, userHandlers)
		if err != nil {
			return nil, err
		}
		kept = append(kept, remarshalled)
	}
	return kept, nil
}

// marshalEntry rebuilds one entry from a matcher and raw surviving handlers.
func marshalEntry(matcher string, handlers []json.RawMessage) (json.RawMessage, error) {
	out := map[string]json.RawMessage{"hooks": json.RawMessage("[" + joinRaw(handlers) + "]")}
	if matcher != "" {
		m, err := json.Marshal(matcher)
		if err != nil {
			return nil, fmt.Errorf("marshal matcher: %w", err)
		}
		out["matcher"] = m
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("re-marshal mixed entry: %w", err)
	}
	return b, nil
}

// joinRaw joins raw JSON values with ", " — deterministic element separation.
func joinRaw(items []json.RawMessage) string {
	parts := make([]string, len(items))
	for i, it := range items {
		parts[i] = string(it)
	}
	return strings.Join(parts, ",")
}

// mustQuote renders s as a JSON string, panicking only on impossible input.
func mustQuote(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(fmt.Sprintf("codexwiring: quote %q: %v", s, err))
	}
	return string(b)
}
