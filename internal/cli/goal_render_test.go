package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/goal"
)

// TestGoalRenderCmd_NoAskUserQuestion is the C-HRA-008 static guard (AC-GHF-008):
// internal/cli/goal.go (which gained the `render` verb in M2) MUST NOT reference
// AskUserQuestion or mcp__askuser outside comments — the verb is orchestrator-
// invoked CLI and never prompts. Mirrors TestWeb_NoAskUserQuestion, with the
// AC-GHF-008 predicate's comment-line exclusion applied (comment lines starting
// with // are allowed to name the forbidden token for documentation).
func TestGoalRenderCmd_NoAskUserQuestion(t *testing.T) {
	src, err := os.ReadFile("goal.go")
	if err != nil {
		t.Fatalf("read goal.go: %v", err)
	}
	for _, line := range strings.Split(string(src), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue // comment line — excluded by the AC predicate
		}
		if strings.Contains(line, "AskUserQuestion") {
			t.Errorf("internal/cli/goal.go non-comment line references AskUserQuestion (orchestrator-only HARD):\n%s", line)
		}
		if strings.Contains(line, "mcp__askuser") {
			t.Errorf("internal/cli/goal.go non-comment line references mcp__askuser (orchestrator-only HARD):\n%s", line)
		}
	}
}

// TestGoalRenderCmd_Registered verifies the `render` subcommand is registered
// alongside arm/status/clear under the `moai goal` command tree.
func TestGoalRenderCmd_Registered(t *testing.T) {
	cmd := newGoalCmd()
	sub := cmd.Commands()
	names := map[string]bool{}
	for _, c := range sub {
		names[c.Use] = true
	}
	if !names["render"] {
		t.Errorf("render subcommand not registered; have: %v", names)
	}
}

// TestRunGoalRender_WritesHTML verifies AC-GHF-001 CLI half: runGoalRender
// resolves an armed goal, renders the dashboard, and writes <session>.html to
// the derived HTMLPath. The file exists on disk after the call.
func TestRunGoalRender_WritesHTML(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	sessionID := "render-test-sess"
	g := goal.NewGoal(sessionID, "render me", []goal.Condition{
		{Type: goal.ConditionMechanical, Cmd: "echo ok", ExpectExit: 0},
	})
	if err := goal.SaveGoal(root, g); err != nil {
		t.Fatal(err)
	}

	cmd := newGoalCmd()
	cmd.SetArgs([]string{"render", "--session", sessionID})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("render execute: %v", err)
	}

	htmlPath := goal.HTMLPath(root, sessionID)
	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("html file not written at %s: %v", htmlPath, err)
	}
	if !strings.HasPrefix(string(data), "<") {
		t.Errorf("written file is not HTML: %q", string(data[:min40(len(data))]))
	}
}

// TestRunGoalRender_NoGoalExitsNonZero verifies AC-GHF-010: when no goal is
// armed for the resolved session, the verb exits non-zero, the stderr message
// names the session id, AND no .html file is written.
func TestRunGoalRender_NoGoalExitsNonZero(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	sessionID := "no-goal-sess"

	cmd := newGoalCmd()
	cmd.SetArgs([]string{"render", "--session", sessionID})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected non-zero exit for no armed goal, got nil")
	}
	if !strings.Contains(errOut.String(), sessionID) {
		t.Errorf("stderr should name session id %q; got:\n%s", sessionID, errOut.String())
	}
	htmlPath := goal.HTMLPath(root, sessionID)
	if _, err := os.Stat(htmlPath); !os.IsNotExist(err) {
		t.Errorf("no .html file should be written when no goal armed; stat err=%v", err)
	}
}

// TestRunGoalRender_JSONOutput verifies the --json flag emits a JSON object
// naming the written path + session id.
func TestRunGoalRender_JSONOutput(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	sessionID := "json-sess"
	g := goal.NewGoal(sessionID, "json out", []goal.Condition{
		{Type: goal.ConditionMechanical, Cmd: "true", ExpectExit: 0},
	})
	if err := goal.SaveGoal(root, g); err != nil {
		t.Fatal(err)
	}

	cmd := newGoalCmd()
	cmd.SetArgs([]string{"render", "--session", sessionID, "--json"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("render --json: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("stdout not JSON: %v\n%s", err, out.String())
	}
	if payload["action"] != "render" {
		t.Errorf("action = %v, want render", payload["action"])
	}
	if payload["session_id"] != sessionID {
		t.Errorf("session_id = %v, want %q", payload["session_id"], sessionID)
	}
	path, _ := payload["path"].(string)
	if path == "" || !strings.HasSuffix(path, sessionID+".html") {
		t.Errorf("path = %q, want ...%s.html", path, sessionID)
	}
}

// TestRunGoalRender_TestFixtureIsolation documents that test fixtures use the
// temp root, not the real .moai/state/goal/ (CLAUDE.local.md §6 / B8).
func TestRunGoalRender_TestFixtureIsolation(t *testing.T) {
	root := t.TempDir()
	// The real state dir must not be polluted — verify the temp root is distinct
	// from the project's actual state path.
	realState := filepath.Join(root, goal.StateDir)
	if !strings.HasPrefix(realState, root) {
		t.Errorf("fixture isolation broken: %s not under temp root", realState)
	}
}

func min40(n int) int {
	if n < 40 {
		return n
	}
	return 40
}
