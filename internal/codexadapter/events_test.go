package codexadapter

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// TestEventTableRowCount pins the table size. The Codex hook event set
// enumerates twelve events (SPEC-CODEX-EVENT-COVERAGE-001 REQ-CEV-001);
// asserting the count means a dropped row fails here rather than silently
// shrinking coverage.
func TestEventTableRowCount(t *testing.T) {
	t.Parallel()

	const wantRows = 12
	if got := len(EventTable); got != wantRows {
		t.Fatalf("EventTable rows = %d, want %d", got, wantRows)
	}
}

// TestEventTableMapping asserts every row maps its Codex event name to the
// dispatcher argument REQ-1's table names (AC-REQ-1a).
func TestEventTableMapping(t *testing.T) {
	t.Parallel()

	want := map[hook.EventType]struct {
		arg     string
		adapted bool
	}{
		hook.EventPreToolUse:        {"pre-tool", true},
		hook.EventPostToolUse:       {"post-tool", true},
		hook.EventSessionStart:      {"session-start", true},
		hook.EventSessionEnd:        {"session-end", true},
		hook.EventStop:              {"stop", true},
		hook.EventUserPromptSubmit:  {"user-prompt-submit", true},
		hook.EventPreCompact:        {"compact", false},
		hook.EventPostCompact:       {"post-compact", false},
		hook.EventPermissionRequest: {"permission-request", false},
		hook.EventSubagentStart:     {"subagent-start", true},
		hook.EventSubagentStop:      {"subagent-stop", true},
		hook.EventType("Interrupt"): {"", false},
	}

	if len(want) != len(EventTable) {
		t.Fatalf("expectation set size %d != EventTable size %d", len(want), len(EventTable))
	}

	for _, row := range EventTable {
		exp, ok := want[row.CodexEvent]
		if !ok {
			t.Errorf("EventTable carries unexpected event %q", row.CodexEvent)
			continue
		}
		if row.DispatcherArg != exp.arg {
			t.Errorf("%s: dispatcher arg = %q, want %q", row.CodexEvent, row.DispatcherArg, exp.arg)
		}
		if row.Adapted != exp.adapted {
			t.Errorf("%s: adapted = %v, want %v", row.CodexEvent, row.Adapted, exp.adapted)
		}
	}
}

// TestAdaptedRowCount pins the adapted subset at eight — the six events
// adapted on the 0.147.0 measurement basis plus SubagentStart/SubagentStop,
// which the 0.153.4 campaign (t496) measured FIRING.
func TestAdaptedRowCount(t *testing.T) {
	t.Parallel()

	const wantAdapted = 8
	got := 0
	for _, row := range EventTable {
		if row.Adapted {
			got++
		}
	}
	if got != wantAdapted {
		t.Fatalf("adapted rows = %d, want %d", got, wantAdapted)
	}
}

// TestResolveAdapted routes the adapted events to their dispatcher argument
// (AC-REQ-1a). SubagentStart/SubagentStop joined the adapted set after the
// 0.153.4 campaign measured them firing (SPEC-CODEX-EVENT-COVERAGE-001 M3).
func TestResolveAdapted(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		event string
		arg   string
	}{
		{"PreToolUse", "pre-tool"},
		{"SubagentStart", "subagent-start"},
		{"SubagentStop", "subagent-stop"},
	} {
		arg, err := Resolve(tc.event)
		if err != nil {
			t.Fatalf("Resolve(%s) error = %v, want nil", tc.event, err)
		}
		if arg != tc.arg {
			t.Fatalf("Resolve(%s) = %q, want %q", tc.event, arg, tc.arg)
		}
	}
}

// TestResolveRecognizedButUnadapted asserts the unadapted events are refused
// as recognized — distinguishable from an unknown name (AC-REQ-1a). After the
// 0.153.4 campaign the set is the two compaction events (trigger not achieved
// in a non-interactive run), PermissionRequest (approval request never raised
// non-interactively), and Interrupt (no MoAI dispatcher counterpart).
func TestResolveRecognizedButUnadapted(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"PreCompact", "PostCompact", "PermissionRequest", "Interrupt"} {
		_, err := Resolve(name)
		if err == nil {
			t.Errorf("Resolve(%s) error = nil, want refusal", name)
			continue
		}
		if !IsUnadapted(err) {
			t.Errorf("Resolve(%s) error = %v, want an unadapted refusal", name, err)
		}
		if IsUnknownEvent(err) {
			t.Errorf("Resolve(%s) classified as unknown; it is recognized-but-unadapted", name)
		}
	}
}

// TestResolveInterruptNoCounterpart asserts the Interrupt refusal carries the
// unadapted class AND a truthful message: Interrupt has no MoAI dispatcher
// counterpart, so the message must not assert that a dispatcher argument
// exists (SPEC-CODEX-EVENT-COVERAGE-001 REQ-CEV-003) — the generic unadapted
// format's "dispatcher arg %q exists" clause is false for an empty arg.
func TestResolveInterruptNoCounterpart(t *testing.T) {
	t.Parallel()

	_, err := Resolve("Interrupt")
	if err == nil {
		t.Fatal("Resolve(Interrupt) error = nil, want refusal")
	}
	if !IsUnadapted(err) {
		t.Fatalf("Resolve(Interrupt) error = %v, want an unadapted refusal", err)
	}
	if IsUnknownEvent(err) {
		t.Fatal("Resolve(Interrupt) classified as unknown; it is recognized-but-unadapted")
	}
	if msg := err.Error(); contains(msg, "dispatcher arg") {
		t.Fatalf("Resolve(Interrupt) message asserts a dispatcher arg exists, but Interrupt has none: %q", msg)
	}
}

// TestResolveUnknownEvent asserts an unrecognized name is rejected and is
// distinguishable from a recognized-but-unadapted one (AC-REQ-1b).
//
// Codex silently ignores unknown event names in its own config, so an adapter
// that defaulted quietly would leave a hook that appears installed and never
// fires.
func TestResolveUnknownEvent(t *testing.T) {
	t.Parallel()

	_, err := Resolve("PreToolUze")
	if err == nil {
		t.Fatal("Resolve of an unknown name returned nil error; want refusal")
	}
	if !IsUnknownEvent(err) {
		t.Fatalf("error = %v, want an unknown-event refusal", err)
	}
	if IsUnadapted(err) {
		t.Fatal("unknown name misclassified as recognized-but-unadapted")
	}
}

// TestResolveErrorNamesReceivedValue asserts the diagnostic names what it got,
// so a typo is identifiable from the message alone (AC-REQ-1b).
func TestResolveErrorNamesReceivedValue(t *testing.T) {
	t.Parallel()

	_, err := Resolve("Bogus")
	if err == nil {
		t.Fatal("want error")
	}
	if got := err.Error(); got == "" || !contains(got, "Bogus") {
		t.Fatalf("error %q does not name the received value", got)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || indexOf(haystack, needle) >= 0)
}

func indexOf(h, n string) int {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return i
		}
	}
	return -1
}
