package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/session"
)

// newGoalTestRoot builds a fresh cobra root carrying the goal command. It
// dispatches the SAME RunE the real rootCmd registers via init() (newGoalCmd is
// the single constructor), so a test driving this root exercises the registered
// `goal arm` code path — NOT a bypassed engine helper (AC-GLE-036 D1 hardening).
func newGoalTestRoot() (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	root := &cobra.Command{Use: "moai", SilenceUsage: true, SilenceErrors: true}
	root.AddCommand(newGoalCmd())
	root.SetOut(buf)
	root.SetErr(buf)
	return root, buf
}

// driveStopGoal runs the real `moai hook stop-goal` verb (runStopGoalHook) with
// the given stdin JSON, capturing its stdout. It proves the hook loads the SAME
// per-session goal file the arm CLI wrote.
func driveStopGoal(t *testing.T, stdinJSON string) string {
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

	buf := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := runStopGoalHook(cmd, nil); err != nil {
		t.Fatalf("runStopGoalHook: %v", err)
	}
	return buf.String()
}

// TestGoalArmEvalLinkage is the make-or-break reachability pin (AC-GLE-036).
// It drives `goal arm` THROUGH the registered command, asserts the arm RunE
// wrote the exact per-session file, then drives `stop-goal` with the SAME
// session id to prove arm and eval share that file.
func TestGoalArmEvalLinkage(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	rc, buf := newGoalTestRoot()
	// "false exits 0" → mechanical cmd="false" expect_exit=0. `false` exits 1 != 0
	// so a live stop-goal evaluation MUST block.
	rc.SetArgs([]string{"goal", "arm", "false exits 0", "--session", "X"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("goal arm execute: %v (out=%s)", err, buf.String())
	}

	// The registered arm RunE MUST have written the state file at the EXACT
	// per-session path (NOT pid-*.json).
	want := filepath.Join(root, goal.StateDir, "X.json")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("arm did not write %s: %v", want, err)
	}
	entries, _ := os.ReadDir(filepath.Join(root, goal.StateDir))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "pid-") {
			t.Fatalf("arm silently pid-fell-back: %s", e.Name())
		}
	}

	// LoadGoal(root, "X") returns the armed goal — parsed condition + ceiling 30.
	g, err := goal.LoadGoal(root, "X")
	if err != nil || g == nil {
		t.Fatalf("LoadGoal(root,\"X\") after arm: g=%+v err=%v", g, err)
	}
	if len(g.Conditions) != 1 || g.Conditions[0].Type != goal.ConditionMechanical {
		t.Fatalf("parsed conditions: %+v", g.Conditions)
	}
	if g.Conditions[0].Cmd != "false" || g.Conditions[0].ExpectExit != 0 {
		t.Fatalf("mechanical parse: cmd=%q expect=%d", g.Conditions[0].Cmd, g.Conditions[0].ExpectExit)
	}
	if g.Ceiling.MaxTurns != goal.DefaultMaxTurns {
		t.Fatalf("ceiling: want %d got %d", goal.DefaultMaxTurns, g.Ceiling.MaxTurns)
	}

	// Drive stop-goal with the SAME session id X → it loads X.json, runs `false`
	// (exit 1 != 0) → block. This closes the arm→eval linkage.
	stdout := driveStopGoal(t, `{"session_id":"X"}`)
	if !strings.Contains(stdout, `"decision":"block"`) {
		t.Fatalf("stop-goal did not block on the armed goal; stdout=%q", stdout)
	}
}

// TestGoalArmResolvesSessionId pins the session-id consistency property
// (AC-GLE-037): given a resolvable real session id, arm writes <id>.json and
// does NOT silently fall back to a pid-<n>.json file.
func TestGoalArmResolvesSessionId(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	// Make a real session id resolvable via the side-channel file that
	// resolveCurrentSessionID reads (written by the SessionStart hook in prod).
	sidecar := filepath.Join(root, session.CurrentSideChannelFile)
	if err := os.MkdirAll(filepath.Dir(sidecar), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sidecar, []byte("real-sess-77"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Arm WITHOUT --session: the arm path MUST resolve the real id, not pid.
	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "arm", "go test ./... exits 0"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("arm: %v out=%s", err, buf.String())
	}

	if _, err := os.Stat(filepath.Join(root, goal.StateDir, "real-sess-77.json")); err != nil {
		t.Fatalf("arm did not key on the resolved session id: %v", err)
	}
	entries, _ := os.ReadDir(filepath.Join(root, goal.StateDir))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "pid-") {
			t.Fatalf("silent pid fallback despite a resolvable session id: %s", e.Name())
		}
	}
}

// TestGoalCmdListsDeliveredVerbs pins AC-GLE-035 (arm/status/clear present, each
// an independent check) + AC-GLE-039a (resume NOT a delivered subcommand).
func TestGoalCmdListsDeliveredVerbs(t *testing.T) {
	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "--help"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("goal --help: %v", err)
	}
	help := buf.String()
	// Three INDEPENDENT checks (not one OR-count) per the reachability lesson.
	if !strings.Contains(help, "arm") {
		t.Error("goal --help missing verb: arm")
	}
	if !strings.Contains(help, "status") {
		t.Error("goal --help missing verb: status")
	}
	if !strings.Contains(help, "clear") {
		t.Error("goal --help missing verb: clear")
	}
	// AC-GLE-039a: resume is deliberately NOT registered (§D.6).
	if strings.Contains(strings.ToLower(help), "resume") {
		t.Error("goal --help must NOT list resume (out of scope §D.6)")
	}
}

// TestGoalStatusClearRoundTrip covers status + clear against an armed goal.
func TestGoalStatusClearRoundTrip(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "arm", "go test ./... exits 0", "--session", "S1"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("arm: %v out=%s", err, buf.String())
	}

	rc, buf = newGoalTestRoot()
	rc.SetArgs([]string{"goal", "status", "--session", "S1"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(buf.String(), "armed") {
		t.Fatalf("status did not report armed: %s", buf.String())
	}

	rc, _ = newGoalTestRoot()
	rc.SetArgs([]string{"goal", "clear", "--session", "S1"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, goal.StateDir, "S1.json")); !os.IsNotExist(err) {
		t.Fatalf("clear did not remove state file: err=%v", err)
	}

	rc, buf = newGoalTestRoot()
	rc.SetArgs([]string{"goal", "status", "--session", "S1"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("status after clear: %v", err)
	}
	if !strings.Contains(strings.ToLower(buf.String()), "no armed goal") {
		t.Fatalf("status after clear: %s", buf.String())
	}
}

// TestGoalArmParsesModelCondition pins the REQ-GLE-032 model branch: a claim
// referencing the transcript becomes a model condition.
func TestGoalArmParsesModelCondition(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "arm", "all AC rows show PASS in the transcript", "--session", "M1"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("arm model: %v out=%s", err, buf.String())
	}
	g, err := goal.LoadGoal(root, "M1")
	if err != nil || g == nil {
		t.Fatalf("load: %v", err)
	}
	if len(g.Conditions) != 1 || g.Conditions[0].Type != goal.ConditionModel {
		t.Fatalf("want model condition, got %+v", g.Conditions)
	}
	if g.Conditions[0].Claim == "" {
		t.Error("model claim empty")
	}
}

// TestGoalBareArmAlias covers the bare `goal "<condition>"` → arm alias.
func TestGoalBareArmAlias(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "true exits 0", "--session", "B1"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("bare arm alias: %v out=%s", err, buf.String())
	}
	if _, err := os.Stat(filepath.Join(root, goal.StateDir, "B1.json")); err != nil {
		t.Fatalf("bare alias did not arm: %v", err)
	}
}

// TestGoalStatusAllAndJSON covers the --all and --json status/arm output paths.
func TestGoalStatusAllAndJSON(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	for _, s := range []string{"A1", "A2"} {
		rc, buf := newGoalTestRoot()
		rc.SetArgs([]string{"goal", "arm", "go test ./... exits 0", "--session", s})
		if err := rc.Execute(); err != nil {
			t.Fatalf("arm %s: %v out=%s", s, err, buf.String())
		}
	}

	// status --all lists both sessions.
	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "status", "--all"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("status --all: %v", err)
	}
	if !strings.Contains(buf.String(), "A1") || !strings.Contains(buf.String(), "A2") {
		t.Fatalf("status --all missing goals: %s", buf.String())
	}

	// status --all --json emits a JSON array.
	rc, buf = newGoalTestRoot()
	rc.SetArgs([]string{"goal", "status", "--all", "--json"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("status --all --json: %v", err)
	}
	if !strings.Contains(buf.String(), `"session_id"`) {
		t.Fatalf("status --all --json not JSON: %s", buf.String())
	}

	// status --json for one session emits the goal JSON.
	rc, buf = newGoalTestRoot()
	rc.SetArgs([]string{"goal", "status", "--session", "A1", "--json"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("status --json: %v", err)
	}
	if !strings.Contains(buf.String(), `"ceiling"`) {
		t.Fatalf("status --json missing ceiling: %s", buf.String())
	}

	// arm --json emits an action=arm payload; clear --json a cleared payload.
	rc, buf = newGoalTestRoot()
	rc.SetArgs([]string{"goal", "arm", "go build ./... exits 0", "--session", "J1", "--json"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("arm --json: %v", err)
	}
	if !strings.Contains(buf.String(), "arm") {
		t.Fatalf("arm --json missing action: %s", buf.String())
	}
	rc, buf = newGoalTestRoot()
	rc.SetArgs([]string{"goal", "clear", "--session", "J1", "--json"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("clear --json: %v", err)
	}
	if !strings.Contains(buf.String(), "clear") {
		t.Fatalf("clear --json missing action: %s", buf.String())
	}
}

// TestGoalStatusAllEmpty covers the no-goal-dir branch of --all.
func TestGoalStatusAllEmpty(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "status", "--all"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("status --all empty: %v", err)
	}
	if !strings.Contains(strings.ToLower(buf.String()), "no armed goals") {
		t.Fatalf("status --all empty: %s", buf.String())
	}
}

// TestGoalArmNoSessionIdWarns covers the documented degrade: no resolvable
// session id → WriterPidKey fallback WITH a surfaced warning (never silent).
func TestGoalArmNoSessionIdWarns(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	// No side-channel file → resolveCurrentSessionID returns ok=false.
	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "arm", "go test ./... exits 0"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("arm no-session: %v out=%s", err, buf.String())
	}
	if !strings.Contains(strings.ToLower(buf.String()), "session.id not available") {
		t.Fatalf("expected a surfaced warning on the no-session-id degrade: %s", buf.String())
	}
	entries, _ := os.ReadDir(filepath.Join(root, goal.StateDir))
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "pid-") {
			found = true
		}
	}
	if !found {
		t.Fatalf("no pid-*.json fallback file written; entries=%v", entries)
	}
}

// TestGoalBareEmptyShowsHelp covers the bare `goal` (no args) → help branch.
func TestGoalBareEmptyShowsHelp(t *testing.T) {
	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("bare goal: %v", err)
	}
	if !strings.Contains(buf.String(), "arm") {
		t.Fatalf("bare goal did not show help: %s", buf.String())
	}
}
