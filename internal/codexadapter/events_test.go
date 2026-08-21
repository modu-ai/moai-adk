package codexadapter

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// TestEventTableRowCount pins the table size. REQ-1 enumerates eleven Codex
// events; asserting the count means a dropped row fails here rather than
// silently shrinking coverage (AC-REQ-1a).
func TestEventTableRowCount(t *testing.T) {
	t.Parallel()

	const wantRows = 11
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
		hook.EventSubagentStart:     {"subagent-start", false},
		hook.EventSubagentStop:      {"subagent-stop", false},
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

// TestAdaptedRowCount pins the adapted subset at six — the events with both a
// payload capture and observed behavior (SPEC §B).
func TestAdaptedRowCount(t *testing.T) {
	t.Parallel()

	const wantAdapted = 6
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

// TestResolveAdapted routes the six adapted events to their dispatcher
// argument (AC-REQ-1a).
func TestResolveAdapted(t *testing.T) {
	t.Parallel()

	arg, err := Resolve("PreToolUse")
	if err != nil {
		t.Fatalf("Resolve(PreToolUse) error = %v, want nil", err)
	}
	if arg != "pre-tool" {
		t.Fatalf("Resolve(PreToolUse) = %q, want %q", arg, "pre-tool")
	}
}

// TestResolveRecognizedButUnadapted asserts the five unadapted events are
// refused as recognized — distinguishable from an unknown name (AC-REQ-1a).
func TestResolveRecognizedButUnadapted(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"PreCompact", "PostCompact", "PermissionRequest", "SubagentStart", "SubagentStop"} {
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
