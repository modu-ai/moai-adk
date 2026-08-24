package codexwiring

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// adaptedEvents derives the expected event-key set from the adapter's
// EventTable — the single data source the renderer itself consumes. The test
// deliberately does not re-enumerate the events: if the adapter adapts a new
// event, the renderer must follow automatically (REQ-CW-002).
func adaptedEvents(t *testing.T) []string {
	t.Helper()
	var events []string
	for _, row := range codexadapter.EventTable {
		if row.Adapted {
			events = append(events, string(row.CodexEvent))
		}
	}
	if len(events) == 0 {
		t.Fatal("codexadapter.EventTable carries no adapted rows — renderer derivation impossible")
	}
	return events
}

// TestRenderHooks_FreshProjectCoversAdaptedEvents verifies the rendered event
// key set equals the EventTable's adapted set exactly (AC-CW-002a).
func TestRenderHooks_FreshProjectCoversAdaptedEvents(t *testing.T) {
	rendered, err := RenderHooks(nil)
	if err != nil {
		t.Fatalf("RenderHooks(nil): %v", err)
	}
	var doc struct {
		Hooks map[string]json.RawMessage `json:"hooks"`
	}
	if err := json.Unmarshal(rendered, &doc); err != nil {
		t.Fatalf("parse rendered hooks: %v\n%s", err, rendered)
	}

	want := adaptedEvents(t)
	if len(doc.Hooks) != len(want) {
		t.Errorf("rendered event count = %d, want %d (adapted EventTable rows): %v",
			len(doc.Hooks), len(want), doc.Hooks)
	}
	for _, ev := range want {
		if _, ok := doc.Hooks[ev]; !ok {
			t.Errorf("adapted event %q missing from rendered hooks", ev)
		}
	}
	for ev := range doc.Hooks {
		known := false
		for _, w := range want {
			if ev == w {
				known = true
				break
			}
		}
		if !known {
			t.Errorf("rendered event %q is not an adapted EventTable row", ev)
		}
	}
}

// TestRenderHooks_CommandsResolveToDispatcherArgs verifies every emitted
// handler command is `moai hook <arg> --harness codex` with <arg> exactly the
// EventTable's dispatcher arg for that event (AC-CW-002b, REQ-CW-002).
func TestRenderHooks_CommandsResolveToDispatcherArgs(t *testing.T) {
	rendered, err := RenderHooks(nil)
	if err != nil {
		t.Fatalf("RenderHooks(nil): %v", err)
	}
	var doc struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Type    string `json:"type"`
				Command string `json:"command"`
				Timeout int    `json:"timeout"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(rendered, &doc); err != nil {
		t.Fatalf("parse rendered hooks: %v", err)
	}
	if len(doc.Hooks) == 0 {
		t.Fatal("no hooks rendered")
	}
	for ev, entries := range doc.Hooks {
		wantArg, rerr := codexadapter.Resolve(ev)
		if rerr != nil {
			t.Errorf("event %q does not Resolve (renderer emitted an unadapted event?): %v", ev, rerr)
			continue
		}
		handlers := 0
		for _, entry := range entries {
			for _, h := range entry.Hooks {
				handlers++
				if !strings.HasPrefix(h.Command, "moai hook ") {
					t.Errorf("%s handler command %q does not start with 'moai hook '", ev, h.Command)
				}
				if !strings.Contains(h.Command, " --harness codex") {
					t.Errorf("%s handler command %q lacks ' --harness codex'", ev, h.Command)
				}
				if h.Command != "moai hook "+wantArg+" --harness codex" {
					t.Errorf("%s handler command = %q, want 'moai hook %s --harness codex' (EventTable dispatcher arg)",
						ev, h.Command, wantArg)
				}
				if h.Type != "command" {
					t.Errorf("%s handler type = %q, want \"command\"", ev, h.Type)
				}
			}
		}
		if handlers == 0 {
			t.Errorf("event %q rendered with zero handlers", ev)
		}
	}
}

// TestRenderHooks_SessionEndTimeoutCap verifies the SessionEnd handler timeout
// stays within Codex's documented ceiling of 3 seconds (REQ-CW-002).
func TestRenderHooks_SessionEndTimeoutCap(t *testing.T) {
	rendered, err := RenderHooks(nil)
	if err != nil {
		t.Fatalf("RenderHooks(nil): %v", err)
	}
	var doc struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Command string `json:"command"`
				Timeout int    `json:"timeout"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(rendered, &doc); err != nil {
		t.Fatalf("parse rendered hooks: %v", err)
	}
	for _, entry := range doc.Hooks[string(hook.EventSessionEnd)] {
		for _, h := range entry.Hooks {
			if h.Timeout > sessionEndTimeoutCeiling {
				t.Errorf("SessionEnd handler timeout = %d, exceeds ceiling %d", h.Timeout, sessionEndTimeoutCeiling)
			}
		}
	}
}

// TestRenderHooks_OtherEventTimeouts verifies non-SessionEnd handlers carry
// the table's default timeout constant (D1 data contract).
func TestRenderHooks_OtherEventTimeouts(t *testing.T) {
	rendered, err := RenderHooks(nil)
	if err != nil {
		t.Fatalf("RenderHooks(nil): %v", err)
	}
	var doc struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Timeout int `json:"timeout"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(rendered, &doc); err != nil {
		t.Fatalf("parse rendered hooks: %v", err)
	}
	if len(doc.Hooks) == 0 {
		t.Fatal("no hooks rendered")
	}
	for ev, entries := range doc.Hooks {
		if ev == string(hook.EventSessionEnd) {
			continue
		}
		for _, entry := range entries {
			for _, h := range entry.Hooks {
				if h.Timeout != defaultHandlerTimeout {
					t.Errorf("%s handler timeout = %d, want default %d", ev, h.Timeout, defaultHandlerTimeout)
				}
			}
		}
	}
}

// TestRenderHooks_TopLevelKeysWhitelisted verifies the rendered document's
// top-level keys are within the measured whitelist {description, hooks} — a
// stray key (e.g. "version") makes Codex silently disable the whole file
// (AC-CW-002c, REQ-CW-003).
func TestRenderHooks_TopLevelKeysWhitelisted(t *testing.T) {
	rendered, err := RenderHooks(nil)
	if err != nil {
		t.Fatalf("RenderHooks(nil): %v", err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(rendered, &top); err != nil {
		t.Fatalf("parse rendered hooks: %v", err)
	}
	for key := range top {
		switch key {
		case "description", "hooks":
		default:
			t.Errorf("top-level key %q outside the measured whitelist {description, hooks}", key)
		}
	}
	if strings.Contains(string(rendered), "\"version\"") {
		t.Errorf("rendered document contains a \"version\" key (Codex disables the whole file on it)")
	}
}

// TestRenderHooks_IdempotentBytes verifies two renders over unchanged input
// are byte-identical — fixed key order, no timestamps, no environment-derived
// values (REQ-CW-006).
func TestRenderHooks_IdempotentBytes(t *testing.T) {
	first, err := RenderHooks(nil)
	if err != nil {
		t.Fatalf("first RenderHooks(nil): %v", err)
	}
	second, err := RenderHooks(first)
	if err != nil {
		t.Fatalf("second RenderHooks(first): %v", err)
	}
	if string(first) != string(second) {
		t.Errorf("two renders over unchanged input differ:\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}
}

// TestRenderHooks_MergePreservesUserEntries verifies the merge model (D2):
// user entries survive a regeneration, stale MoAI handlers are removed, the
// current table is appended (AC-CW-007, REQ-CW-005).
func TestRenderHooks_MergePreservesUserEntries(t *testing.T) {
	existing := []byte(`{
  "description": "user's own description",
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [{"type": "command", "command": "my-own-hook", "timeout": 30}]}
    ],
    "SessionEnd": [
      {"hooks": [{"type": "command", "command": "moai hook session-end --harness codex", "timeout": 999}]}
    ]
  }
}`)
	merged, err := RenderHooks(existing)
	if err != nil {
		t.Fatalf("RenderHooks(existing): %v", err)
	}
	body := string(merged)

	if !strings.Contains(body, "my-own-hook") {
		t.Errorf("user handler 'my-own-hook' dropped by merge:\n%s", merged)
	}
	if !strings.Contains(body, "\"matcher\": \"Bash\"") {
		t.Errorf("user matcher 'Bash' dropped by merge:\n%s", merged)
	}
	if !strings.Contains(body, "user's own description") {
		t.Errorf("user description replaced by merge:\n%s", merged)
	}
	if strings.Contains(body, `"timeout": 999`) {
		t.Errorf("stale MoAI SessionEnd handler (timeout 999) survived the merge:\n%s", merged)
	}

	// The stale SessionEnd handler must be replaced by a capped one.
	var doc struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Command string `json:"command"`
				Timeout int    `json:"timeout"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(merged, &doc); err != nil {
		t.Fatalf("parse merged hooks: %v", err)
	}
	for _, entry := range doc.Hooks["SessionEnd"] {
		for _, h := range entry.Hooks {
			if strings.HasPrefix(h.Command, "moai hook ") && h.Timeout > sessionEndTimeoutCeiling {
				t.Errorf("merged MoAI SessionEnd handler timeout = %d, exceeds ceiling %d", h.Timeout, sessionEndTimeoutCeiling)
			}
		}
	}

	// Re-merging the merged output is idempotent (REQ-CW-006).
	again, err := RenderHooks(merged)
	if err != nil {
		t.Fatalf("re-render merged output: %v", err)
	}
	if string(again) != string(merged) {
		t.Errorf("merge is not idempotent:\n--- merged ---\n%s\n--- again ---\n%s", merged, again)
	}
}

// TestRenderHooks_MixedEntryKeepsUserHandlersOnly verifies a user entry that
// shares its handler array with a stale MoAI handler survives the merge with
// its matcher and its own handlers intact — only the MoAI handler is removed.
func TestRenderHooks_MixedEntryKeepsUserHandlersOnly(t *testing.T) {
	existing := []byte(`{
  "hooks": {
    "Stop": [
      {"matcher": "Bash", "hooks": [
        {"type": "command", "command": "my-own-hook", "timeout": 5},
        {"type": "command", "command": "moai hook stop --harness codex", "timeout": 999}
      ]}
    ]
  }
}`)
	merged, err := RenderHooks(existing)
	if err != nil {
		t.Fatalf("RenderHooks(mixed): %v", err)
	}
	var doc struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(merged, &doc); err != nil {
		t.Fatalf("parse merged: %v", err)
	}
	entries := doc.Hooks["Stop"]
	if len(entries) != 2 {
		t.Fatalf("Stop entries = %d, want 2 (user entry + current MoAI entry):\n%s", len(entries), merged)
	}
	first := entries[0]
	if first.Matcher != "Bash" {
		t.Errorf("mixed-entry matcher lost: %+v", first)
	}
	for _, h := range first.Hooks {
		if strings.HasPrefix(h.Command, "moai hook ") {
			t.Errorf("stale MoAI handler survived inside the user entry: %q", h.Command)
		}
	}
	if len(first.Hooks) != 1 || first.Hooks[0].Command != "my-own-hook" {
		t.Errorf("user handler altered in mixed entry: %+v", first.Hooks)
	}
}

// TestEnsureStatusLineAppendWithoutTrailingNewline verifies appendSection's
// newline normalization on the no-trailing-newline edge.
func TestEnsureStatusLineAppendWithoutTrailingNewline(t *testing.T) {
	out := EnsureStatusLine([]byte("model = \"gpt-5\""))
	body := string(out)
	if !strings.Contains(body, "[tui]\n") || !strings.Contains(body, statusLineDefaultTOML) {
		t.Errorf("status_line not appended to newline-less content:\n%q", body)
	}
}

// TestRenderHooks_ValidationGateRejectsViolatingUserKeys verifies the
// REQ-CW-003 pre-write gate at the render level: an existing document carrying
// whitelist-violating keys cannot produce writable bytes (the caller must
// refuse to write, leaving the file untouched).
func TestRenderHooks_ValidationGateRejectsViolatingUserKeys(t *testing.T) {
	existing := []byte(`{
  "version": 1,
  "hooks": {
    "PreToolUse": [
      {"hooks": [{"type": "command", "command": "my-own-hook"}]}
    ]
  }
}`)
	if violations, err := codexadapter.ValidateConfig(existing); err != nil || len(violations) == 0 {
		// sanity anchor: the adapter itself must already flag this document.
		t.Fatalf("codexadapter.ValidateConfig did not flag a 'version' key — test premise broken (violations=%v err=%v)", violations, err)
	}
	rendered, err := RenderHooks(existing)
	if err != nil {
		t.Fatalf("RenderHooks must not fail on violating user input (gate is at write time): %v", err)
	}
	violations, verr := codexadapter.ValidateConfig(rendered)
	if verr != nil {
		t.Fatalf("validate rendered: %v", verr)
	}
	if len(violations) == 0 {
		t.Errorf("rendered bytes carrying user's 'version' key passed the whitelist gate — Wire would write a file Codex silently disables:\n%s", rendered)
	}
}
