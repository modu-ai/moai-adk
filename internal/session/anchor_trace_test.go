package session

// SPEC-SESSION-ANCHOR-ATTR-001 W3 — the MOAI_ANCHOR_TRACE switch
// (REQ-SAA-007..009). Off by default: no trace output and no per-decision
// overhead beyond the single environment lookup. On: one verbose JSONL row
// per anchor decision carrying session_id, pid, cwd, and a timestamp.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// readTraceRows parses the anchor trace log under root.
func readTraceRows(t *testing.T, root string) []AnchorTraceEvent {
	t.Helper()
	data, err := os.ReadFile(AnchorTracePath(root))
	if err != nil {
		t.Fatalf("read anchor trace log: %v", err)
	}
	var rows []AnchorTraceEvent
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var row AnchorTraceEvent
		if err := json.Unmarshal([]byte(line), &row); err == nil {
			rows = append(rows, row)
		}
	}
	return rows
}

// TestAnchorTraceEnabledTruthy covers the gate's truthiness: "1" and "true"
// enable (case-insensitive); unset, empty, "0", and "false" do not.
func TestAnchorTraceEnabledTruthy(t *testing.T) {
	truthy := []string{"1", "true", "TRUE", "True"}
	for _, v := range truthy {
		t.Setenv(config.EnvAnchorTrace, v)
		if !AnchorTraceEnabled() {
			t.Errorf("AnchorTraceEnabled() with %s=%q = false, want true", config.EnvAnchorTrace, v)
		}
	}
	falsy := []string{"", "0", "false", "no", "off"}
	for _, v := range falsy {
		t.Setenv(config.EnvAnchorTrace, v)
		if AnchorTraceEnabled() {
			t.Errorf("AnchorTraceEnabled() with %s=%q = true, want false", config.EnvAnchorTrace, v)
		}
	}
	// The unset case reads as falsy too — asserted via the empty-string arm
	// above plus an explicit unset through os to keep t.Setenv's restore
	// semantics (the value is restored at test end either way).
	_ = os.Unsetenv(config.EnvAnchorTrace)
	if AnchorTraceEnabled() {
		t.Errorf("AnchorTraceEnabled() with the variable unset = true, want false")
	}
}

// TestAnchorTraceRelocateSessionRow covers REQ-SAA-007 on the relocation
// decision point: with the switch on, every relocation appends one row
// carrying session_id, pid, cwd, and a timestamp.
func TestAnchorTraceRelocateSessionRow(t *testing.T) {
	t.Setenv(config.EnvAnchorTrace, "1")
	host, _ := os.Hostname()
	root := t.TempDir()
	writeAnchorRegistry(t, root, []Entry{
		{SessionID: "sess-trace", SpecID: "SPEC-X", Phase: "run", LastHeartbeat: time.Now().UTC().Add(time.Hour), PID: 4243, Host: host, CWD: "/tree/old"},
	})
	reg := NewRegistry(filepath.Join(root, DefaultRegistryPath), nil)

	if err := reg.RelocateSession("sess-trace", "/tree/new"); err != nil {
		t.Fatalf("relocate: %v", err)
	}

	rows := readTraceRows(t, root)
	var relocate []AnchorTraceEvent
	for _, r := range rows {
		if r.Decision == "relocate" {
			relocate = append(relocate, r)
		}
	}
	if len(relocate) != 1 {
		t.Fatalf("want exactly 1 relocate trace row, got %d (all: %+v)", len(relocate), rows)
	}
	row := relocate[0]
	if row.SessionID != "sess-trace" {
		t.Errorf("session_id = %q, want sess-trace", row.SessionID)
	}
	if row.PID <= 0 {
		t.Errorf("pid = %d, want a positive pid", row.PID)
	}
	if row.Cwd != "/tree/new" {
		t.Errorf("cwd = %q, want /tree/new", row.Cwd)
	}
	if _, err := time.Parse(time.RFC3339, row.Timestamp); err != nil {
		t.Errorf("timestamp %q is not RFC3339: %v", row.Timestamp, err)
	}
}

// TestAnchorTraceDisabledEmitsNothing covers REQ-SAA-008's negative path WITH
// its positive control (acceptance AC-008): the same relocation fixture emits
// nothing with the switch unset, and the control run with the switch on lands
// a row — the zero output is evidence of the gate, not of a broken writer.
func TestAnchorTraceDisabledEmitsNothing(t *testing.T) {
	host, _ := os.Hostname()

	// Negative arm: switch unset — no trace file at all.
	_ = os.Unsetenv(config.EnvAnchorTrace)
	root := t.TempDir()
	writeAnchorRegistry(t, root, []Entry{
		{SessionID: "sess-off", SpecID: "SPEC-X", Phase: "run", LastHeartbeat: time.Now().UTC().Add(time.Hour), PID: os.Getpid(), Host: host, CWD: "/tree/old"},
	})
	reg := NewRegistry(filepath.Join(root, DefaultRegistryPath), nil)
	if err := reg.RelocateSession("sess-off", "/tree/new"); err != nil {
		t.Fatalf("relocate: %v", err)
	}
	if _, err := os.Stat(AnchorTracePath(root)); !os.IsNotExist(err) {
		t.Fatalf("trace file must not exist with the switch unset, stat err: %v", err)
	}

	// Positive control: the identical fixture with the switch on lands a row.
	t.Setenv(config.EnvAnchorTrace, "1")
	if err := reg.RelocateSession("sess-off", "/tree/new2"); err != nil {
		t.Fatalf("relocate: %v", err)
	}
	rows := readTraceRows(t, root)
	if len(rows) == 0 {
		t.Fatal("positive control produced no trace rows — the zero above proves nothing")
	}
}

// TestAnchorTraceAnchorDecisionRow covers REQ-SAA-007 on the disposal-side
// anchor decision point: the AnchorDecision outcome is traced with the
// deciding process's pid, the judged tree as cwd, and the verdict source.
func TestAnchorTraceAnchorDecisionRow(t *testing.T) {
	t.Setenv(config.EnvAnchorTrace, "1")
	root := t.TempDir()
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	lock := LockInfo{Locked: true, Reason: "claude session t-other (pid " + itoaForTest(os.Getpid()) + ")"}
	verdict := AnchorDecision(root, lock, time.Now().UTC())
	if !verdict.Anchored {
		t.Fatalf("fixture expected an anchored verdict, got %+v", verdict)
	}

	rows := readTraceRows(t, root)
	var decisions []AnchorTraceEvent
	for _, r := range rows {
		if r.Decision == "anchor_decision" {
			decisions = append(decisions, r)
		}
	}
	if len(decisions) != 1 {
		t.Fatalf("want exactly 1 anchor_decision trace row, got %d (all: %+v)", len(decisions), rows)
	}
	row := decisions[0]
	if row.PID <= 0 {
		t.Errorf("pid = %d, want the deciding process pid", row.PID)
	}
	if row.Cwd != root {
		t.Errorf("cwd = %q, want the judged tree %q", row.Cwd, root)
	}
	if !strings.Contains(row.Detail, string(AnchorSourceLock)) {
		t.Errorf("detail = %q, want the verdict source %q recorded", row.Detail, AnchorSourceLock)
	}
}

func itoaForTest(n int) string {
	return strconv.Itoa(n)
}
