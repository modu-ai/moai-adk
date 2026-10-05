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
		hook.EventPreToolUse:       {"pre-tool", true},
		hook.EventPostToolUse:      {"post-tool", true},
		hook.EventSessionStart:     {"session-start", true},
		hook.EventSessionEnd:       {"session-end", true},
		hook.EventStop:             {"stop", true},
		hook.EventUserPromptSubmit: {"user-prompt-submit", true},
		// SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2e: the four rows held back by
		// SPEC-CODEX-EVENT-COVERAGE-001 are adapted, and Interrupt gains the
		// Codex-only `interrupt` dispatcher arg (intentional amendment of
		// REQ-CEV-001).
		hook.EventPreCompact:        {"compact", true},
		hook.EventPostCompact:       {"post-compact", true},
		hook.EventPermissionRequest: {"permission-request", true},
		hook.EventSubagentStart:     {"subagent-start", true},
		hook.EventSubagentStop:      {"subagent-stop", true},
		hook.EventType("Interrupt"): {"interrupt", true},
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

// TestAdaptedRowCount pins the adapted subset at twelve — every row. The six
// events adapted on the 0.147.0 measurement basis, SubagentStart/SubagentStop
// (measured FIRING by the 0.153.4 campaign, t496), and the four rows
// SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2e adapts: PreCompact, PostCompact,
// PermissionRequest, and Interrupt (intentional amendment, 8 → 12).
func TestAdaptedRowCount(t *testing.T) {
	t.Parallel()

	const wantAdapted = 12
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

// TestResolveFormerlyUnadaptedNowResolve asserts the four events
// SPEC-CODEX-EVENT-COVERAGE-001 held back now resolve to their dispatcher
// argument (SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2e — intentional amendment of
// TestResolveRecognizedButUnadapted, REQ-CEV-003). The unknown-vs-unadapted
// distinction that test guarded stays covered by TestResolveUnknownEvent and
// by TestResolveUnadaptedRowIsRefused.
func TestResolveFormerlyUnadaptedNowResolve(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		event string
		arg   string
	}{
		{"PreCompact", "compact"},
		{"PostCompact", "post-compact"},
		{"PermissionRequest", "permission-request"},
		{"Interrupt", "interrupt"},
	} {
		arg, err := Resolve(tc.event)
		if err != nil {
			t.Errorf("Resolve(%s) error = %v, want %q", tc.event, err, tc.arg)
			continue
		}
		if arg != tc.arg {
			t.Errorf("Resolve(%s) = %q, want %q", tc.event, arg, tc.arg)
		}
	}
}

// TestResolveInterruptResolvesToInterrupt asserts Interrupt resolves to the
// Codex-only `interrupt` dispatcher arg handled in internal/cli
// (SPEC-DUAL-HARNESS-HOOK-PARITY-001 design.md §D7 — intentional amendment of
// TestResolveInterruptNoCounterpart, REQ-CEV-003). The event constant still
// lives in this package, not internal/hook (REQ-CEV-002 kept).
func TestResolveInterruptResolvesToInterrupt(t *testing.T) {
	t.Parallel()

	arg, err := Resolve(string(CodexEventInterrupt))
	if err != nil {
		t.Fatalf("Resolve(Interrupt) error = %v, want \"interrupt\"", err)
	}
	if arg != "interrupt" {
		t.Fatalf("Resolve(Interrupt) = %q, want \"interrupt\"", arg)
	}
}

// TestResolveUnadaptedRowIsRefused keeps the recognized-but-unadapted refusal
// path covered now that no shipped row is unadapted: a row with Adapted false
// is refused with the unadapted class, never classified unknown, and — when it
// has no dispatcher arg — without claiming one exists.
func TestResolveUnadaptedRowIsRefused(t *testing.T) {
	t.Parallel()

	rows := []EventRow{
		{hook.EventPreCompact, "compact", false},
		{CodexEventInterrupt, "", false},
	}
	for _, name := range []string{"PreCompact", "Interrupt"} {
		_, err := resolveIn(rows, name)
		if err == nil {
			t.Errorf("Resolve(%s) error = nil, want refusal", name)
			continue
		}
		if !IsUnadapted(err) || IsUnknownEvent(err) {
			t.Errorf("Resolve(%s) error = %v, want an unadapted (not unknown) refusal", name, err)
		}
	}
	if _, err := resolveIn(rows, "Interrupt"); err != nil && contains(err.Error(), "dispatcher arg") {
		t.Errorf("an arg-less unadapted row claims a dispatcher arg exists: %v", err)
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
