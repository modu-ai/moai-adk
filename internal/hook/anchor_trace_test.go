package hook

// SPEC-SESSION-ANCHOR-ATTR-001 W3 — the branch-guard Seam A anchor read is
// one of the traced anchor decision points (REQ-SAA-007). The trace row
// carries session_id, the hook process pid, the resolved git-context cwd
// (Seam A preserved: input.CWD chain — REQ-SAA-012), and the read outcome.

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/session"
)

// branchGuardTraceInput builds a HookInput carrying a branch-state command
// the guard's matcher recognizes.
func branchGuardTraceInput(sessionID, cwd string) *HookInput {
	command := map[string]string{"command": "git switch main"}
	raw, _ := json.Marshal(command)
	return &HookInput{
		SessionID:     sessionID,
		ToolName:      "Bash",
		ToolInput:     raw,
		CWD:           cwd,
		HookEventName: "PreToolUse",
	}
}

// readAnchorTraceRows parses the anchor trace log under root.
func readAnchorTraceRows(t *testing.T, root string) []session.AnchorTraceEvent {
	t.Helper()
	data, err := os.ReadFile(session.AnchorTracePath(root))
	if err != nil {
		t.Fatalf("read anchor trace log: %v", err)
	}
	var rows []session.AnchorTraceEvent
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var row session.AnchorTraceEvent
		if err := json.Unmarshal([]byte(line), &row); err == nil {
			rows = append(rows, row)
		}
	}
	return rows
}

// TestAnchorTraceBranchGuardAnchorRead covers the Seam A anchor-read trace:
// with the switch on, a judged branch-state command appends one row naming
// the session, the hook pid, the git-context cwd, and the read outcome —
// even on the fail-open uncertainty path (a non-git cwd).
func TestAnchorTraceBranchGuardAnchorRead(t *testing.T) {
	t.Setenv(config.EnvAnchorTrace, "1")
	projectDir := t.TempDir()
	// A non-git cwd forces the fail-open uncertainty path — the trace row
	// must still land, with the error recorded in the detail.
	nonGitCwd := t.TempDir()

	input := branchGuardTraceInput("sess-bg-trace", nonGitCwd)
	decision, reason := checkBranchState(input, projectDir)
	if decision != "" || reason != "" {
		t.Fatalf("fixture expected the fail-open allow path, got %q/%q", decision, reason)
	}

	rows := readAnchorTraceRows(t, projectDir)
	var reads []session.AnchorTraceEvent
	for _, r := range rows {
		if r.Decision == "branch_guard.anchor_read" {
			reads = append(reads, r)
		}
	}
	if len(reads) != 1 {
		t.Fatalf("want exactly 1 anchor_read trace row, got %d (all: %+v)", len(reads), rows)
	}
	row := reads[0]
	if row.SessionID != "sess-bg-trace" {
		t.Errorf("session_id = %q, want sess-bg-trace", row.SessionID)
	}
	if row.PID <= 0 {
		t.Errorf("pid = %d, want the hook process pid", row.PID)
	}
	if row.Cwd == "" {
		t.Error("cwd must carry the resolved git-context directory")
	}
	if row.Detail == "" {
		t.Error("detail must record the read outcome")
	}
	if _, err := time.Parse(time.RFC3339, row.Timestamp); err != nil {
		t.Errorf("timestamp %q is not RFC3339: %v", row.Timestamp, err)
	}
}

// TestAnchorTraceBranchGuardDisabledEmitsNothing is the negative arm with its
// positive control (acceptance AC-008): the switch off emits nothing on the
// identical fixture, the control run on emits the row.
func TestAnchorTraceBranchGuardDisabledEmitsNothing(t *testing.T) {
	_ = os.Unsetenv(config.EnvAnchorTrace)
	projectDir := t.TempDir()
	nonGitCwd := t.TempDir()

	checkBranchState(branchGuardTraceInput("sess-bg-off", nonGitCwd), projectDir)
	if _, err := os.Stat(session.AnchorTracePath(projectDir)); !os.IsNotExist(err) {
		t.Fatalf("trace file must not exist with the switch unset, stat err: %v", err)
	}

	// Positive control: switch on, same fixture.
	t.Setenv(config.EnvAnchorTrace, "1")
	checkBranchState(branchGuardTraceInput("sess-bg-off", nonGitCwd), projectDir)
	if rows := readAnchorTraceRows(t, projectDir); len(rows) == 0 {
		t.Fatal("positive control produced no trace rows — the zero above proves nothing")
	}
}
