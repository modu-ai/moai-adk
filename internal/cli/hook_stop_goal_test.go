package cli

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// assertSingleJSONObject asserts the Stop-hook stdout contract: stdout is
// either EMPTY (a silent allow / fail-open path) or EXACTLY one JSON object —
// decoding must exhaust the input. Claude Code parses the whole Stop-hook
// stdout as a single JSON object, so any trailing data after the first value
// fails the parse ("Extra data: line 2 column 1") and a carried block decision
// is silently lost (ISSUE #1703). A strings.Contains assertion cannot see
// trailing pollution; only decoder exhaustion can.
func assertSingleJSONObject(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" {
		return
	}
	dec := json.NewDecoder(strings.NewReader(stdout))
	var v json.RawMessage
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("stdout is not valid JSON: %v; stdout=%q", err, stdout)
	}
	var obj map[string]any
	if err := json.Unmarshal(v, &obj); err != nil {
		t.Fatalf("stdout JSON value is not an object: %v; stdout=%q", err, stdout)
	}
	if err := dec.Decode(&json.RawMessage{}); err != io.EOF {
		t.Fatalf("stdout carries data after the first JSON object (the ISSUE #1703 pollution shape); trailing=%q", stdout)
	}
}

const goalStateDirForTest = ".moai/state/goal"

// driveStopGoalStreams runs runStopGoalHook with the given stdin JSON and
// SEPARATE stdout / stderr captures. The shared driveStopGoal harness merges
// both streams into one buffer, which would mask the very contract this test
// locks (diagnostics on stderr must never reach stdout).
func driveStopGoalStreams(t *testing.T, stdinJSON string) (stdout, stderr string) {
	t.Helper()
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	go func() {
		_, _ = w.Write([]byte(stdinJSON))
		_ = w.Close()
	}()
	defer func() { os.Stdin = oldStdin }()

	outBuf := &strings.Builder{}
	errBuf := &strings.Builder{}
	cmd := &cobra.Command{}
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	if err := runStopGoalHook(cmd, nil); err != nil {
		t.Fatalf("runStopGoalHook: %v", err)
	}
	return outBuf.String(), errBuf.String()
}

// TestStopGoalStdoutSingleJSONObjectOnEveryExitPath is the ISSUE #1703
// regression: every stop-goal exit path must leave stdout EMPTY or holding
// EXACTLY one JSON object. The paths:
//
//	block        — armed goal, failing mechanical condition → decision:block
//	ceiling_exit — turns exhausted on a failing condition → ceiling verdict JSON
//	allow        — all conditions satisfied → silent (empty stdout)
//	load error   — unreadable/corrupt state file → fail-open, silent stdout
//	no goal      — no state file at all → silent stdout
//
// The production defect was a WRAPPER defect (t220 double-fire, fixed in
// c4e90cd58) whose visible symptom — a binary without the stop-goal subcommand
// appending the parent `hook` help to the verdict JSON's stdout — this binary
// cannot reproduce in-process; what this test locks is the evaluator-side half
// of the contract the pollution violated: whatever the path, stdout never
// carries more than one JSON object.
func TestStopGoalStdoutSingleJSONObjectOnEveryExitPath(t *testing.T) {
	cases := []struct {
		name      string
		stateJSON string // written verbatim as <session>.json; empty = no state file
		// wantEmpty asserts the silent paths; non-silent paths additionally
		// name a field that must appear in the single JSON object.
		wantEmpty bool
		wantField string
	}{
		{
			name: "block",
			stateJSON: `{"session_id":"t680-block","goal":"fixture","conditions":[{"type":"mechanical","cmd":"false","expect_exit":0}],` +
				`"ceiling":{"max_turns":30},"turns_used":0,"progress":[],"progression_mode":"autonomous","created_at":"2026-09-13T00:00:00Z","status":"armed"}`,
			wantField: "decision",
		},
		{
			name: "ceiling_exit",
			stateJSON: `{"session_id":"t680-ceiling","goal":"fixture","conditions":[{"type":"mechanical","cmd":"false","expect_exit":0}],` +
				`"ceiling":{"max_turns":30},"turns_used":30,"progress":[],"progression_mode":"autonomous","created_at":"2026-09-13T00:00:00Z","status":"armed"}`,
			wantField: "ceiling_exit",
		},
		{
			name: "allow",
			stateJSON: `{"session_id":"t680-allow","goal":"fixture","conditions":[{"type":"mechanical","cmd":"true","expect_exit":0}],` +
				`"ceiling":{"max_turns":30},"turns_used":0,"progress":[],"progression_mode":"autonomous","created_at":"2026-09-13T00:00:00Z","status":"armed"}`,
			wantEmpty: true,
		},
		{
			name:      "load error",
			stateJSON: `{"session_id":"t680-corrupt", NOT VALID JSON`,
			wantEmpty: true,
		},
		{
			name:      "no goal",
			stateJSON: "",
			wantEmpty: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("CLAUDE_PROJECT_DIR", root)
			sessionID := "t680-" + strings.ReplaceAll(tc.name, " ", "-")
			if tc.stateJSON != "" {
				stateDir := filepath.Join(root, goalStateDirForTest)
				if err := os.MkdirAll(stateDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(stateDir, sessionID+".json"), []byte(tc.stateJSON), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			stdout, stderr := driveStopGoalStreams(t, `{"session_id":"`+sessionID+`"}`)
			if tc.wantEmpty && stdout != "" {
				t.Fatalf("want silent stdout, got %q", stdout)
			}
			assertSingleJSONObject(t, stdout)
			if tc.wantField != "" && !strings.Contains(stdout, `"`+tc.wantField+`"`) {
				t.Fatalf("stdout JSON missing %q: %s", tc.wantField, stdout)
			}
			if !tc.wantEmpty && stderr != "" {
				t.Fatalf("want clean stderr on the emitting path, got %q", stderr)
			}
		})
	}
}
