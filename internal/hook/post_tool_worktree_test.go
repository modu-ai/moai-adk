package hook

// t236 / issue #1640 — EnterWorktree/ExitWorktree PostToolUse must re-stamp
// MOAI_PROJECT_DIR into CLAUDE_ENV_FILE and relocate the session registry cwd,
// because Claude Code emits no CwdChanged for tool-based tree moves.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPostToolWorktreeMove_StampsEnvFile (t236): an EnterWorktree PostToolUse
// must re-stamp MOAI_PROJECT_DIR into CLAUDE_ENV_FILE for the new tree even
// though no CwdChanged fires, must preserve pre-existing content, and must be
// idempotent on repeat. An ExitWorktree to a sibling tree appends the new line
// while keeping the old one. Both moves surface a non-empty systemMessage.
func TestPostToolWorktreeMove_StampsEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "claude-env")

	if err := os.WriteFile(envFile, []byte("export USER_SET_VAR=\"keep-me\"\n"), 0o644); err != nil {
		t.Fatalf("seed env file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai", "config"), 0o755); err != nil {
		t.Fatalf("create moai config dir: %v", err)
	}
	t.Setenv("CLAUDE_ENV_FILE", envFile)

	want := "export MOAI_PROJECT_DIR=\"" + tmpDir + "\""

	h := NewPostToolHandler()
	out, err := h.Handle(context.Background(), &HookInput{SessionID: "sess-wt", ToolName: "EnterWorktree", CWD: tmpDir})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if out == nil || out.SystemMessage == "" {
		t.Fatalf("expected a systemMessage about the spawn-frozen MCP server root, got %+v", out)
	}

	data, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("env file missing stamp %q; got:\n%s", want, data)
	}
	if !strings.Contains(string(data), "USER_SET_VAR") {
		t.Fatalf("env file lost pre-existing content:\n%s", data)
	}

	// Idempotent on repeat: the same move must not append a second copy.
	outRepeat, repeatErr := h.Handle(context.Background(), &HookInput{SessionID: "sess-wt", ToolName: "EnterWorktree", CWD: tmpDir})
	if repeatErr != nil {
		t.Fatalf("repeat Handle() error = %v", repeatErr)
	}
	if outRepeat == nil {
		t.Fatalf("repeat Handle() returned nil output")
	}
	dataRepeat, repeatReadErr := os.ReadFile(envFile)
	if repeatReadErr != nil {
		t.Fatalf("re-read env file: %v", repeatReadErr)
	}
	if n := strings.Count(string(dataRepeat), want); n != 1 {
		t.Fatalf("stamp not idempotent: %d occurrences of %q", n, want)
	}

	// ExitWorktree to a second tree appends that tree's line; both lines stay.
	other := t.TempDir()
	if err := os.MkdirAll(filepath.Join(other, ".moai", "config"), 0o755); err != nil {
		t.Fatalf("create other moai config dir: %v", err)
	}
	out2, err2 := h.Handle(context.Background(), &HookInput{SessionID: "sess-wt", ToolName: "ExitWorktree", CWD: other})
	if err2 != nil {
		t.Fatalf("second Handle() error = %v", err2)
	}
	if out2 == nil || out2.SystemMessage == "" {
		t.Fatalf("expected systemMessage on exit move too, got %+v", out2)
	}
	dataExit, exitReadErr := os.ReadFile(envFile)
	if exitReadErr != nil {
		t.Fatalf("re-read env file after exit: %v", exitReadErr)
	}
	if !strings.Contains(string(dataExit), "export MOAI_PROJECT_DIR=\""+other+"\"") {
		t.Fatalf("second tree not stamped:\n%s", dataExit)
	}
	if !strings.Contains(string(dataExit), want) {
		t.Fatalf("first tree's stamp lost after exit move:\n%s", dataExit)
	}
}

// TestPostToolWorktreeMove_NoConfigDirNoStamp: a tree without .moai/config
// must not be stamped (same rule as the CwdChanged handler), and no env file
// is created for it.
func TestPostToolWorktreeMove_NoConfigDirNoStamp(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "claude-env")
	t.Setenv("CLAUDE_ENV_FILE", envFile)

	out, err := NewPostToolHandler().Handle(context.Background(),
		&HookInput{SessionID: "s", ToolName: "EnterWorktree", CWD: tmpDir})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if out == nil {
		t.Fatalf("Handle() returned nil output")
	}
	if _, statErr := os.Stat(envFile); !os.IsNotExist(statErr) {
		t.Fatalf("env file must not be created for a non-moai dir, stat err = %v", statErr)
	}
}

// TestPostToolWorktreeMove_EmptyCwdFailsOpen: a hook input carrying neither
// CWD nor NewCwd must produce no error and no panic — the branch fails open.
func TestPostToolWorktreeMove_EmptyCwdFailsOpen(t *testing.T) {
	out, err := NewPostToolHandler().Handle(context.Background(),
		&HookInput{SessionID: "s", ToolName: "EnterWorktree"})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if out == nil {
		t.Fatalf("Handle() returned nil output")
	}
}
