package worktree

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/session"
)

// writeTreeRegistry writes a session registry at
// <tree>/.moai/state/active-sessions.json — the tree-LOCAL registry a
// `moai cc -w` lane session registers itself into (hooks resolve the
// registry from the session's own project dir).
func writeTreeRegistry(t *testing.T, tree string, entries []session.Entry) {
	t.Helper()
	dir := filepath.Join(tree, ".moai", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal entries: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "active-sessions.json"), data, 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
}

// anchoredEntry builds a registry entry whose working directory is cwd.
// Heartbeat is deliberately hours old: no per-turn heartbeat driver exists,
// so a long-running LIVE lane carries a stale heartbeat — liveness must be
// judged from the PID probe, not the timestamp.
func anchoredEntry(t *testing.T, cwd string, pid int) session.Entry {
	t.Helper()
	host, _ := os.Hostname()
	old := time.Now().UTC().Add(-2 * time.Hour)
	return session.Entry{
		SessionID:     "11111111-2222-3333-4444-555555555555",
		SpecID:        "SPEC-ANCHOR-001",
		Phase:         "run",
		StartedAt:     old,
		LastHeartbeat: old,
		PID:           pid,
		Host:          host,
		CWD:           cwd,
	}
}

// TestRunDoneWithAutoMode_SkipsRemovalWhenAnchoredSessionLive reproduces the
// t46 incident shape: a lane session launched via `moai cc -w` is still
// alive and registered in the tree-local registry when cleanup runs. Auto
// cleanup must not remove the worktree out from under it.
func TestRunDoneWithAutoMode_SkipsRemovalWhenAnchoredSessionLive(t *testing.T) {
	tree := t.TempDir()
	writeTreeRegistry(t, tree, []session.Entry{anchoredEntry(t, tree, os.Getpid())})

	origProvider := WorktreeProvider
	mock := &mockWorktreeProvider{
		worktrees: []git.Worktree{{Branch: "feature/SPEC-ANCHOR-001", Path: tree}},
	}
	WorktreeProvider = mock
	defer func() { WorktreeProvider = origProvider }()

	success, err := runDoneWorktreeCleanup("feature/SPEC-ANCHOR-001", false, false)
	if err != nil {
		t.Fatalf("auto mode must degrade gracefully, got error: %v", err)
	}
	if success {
		t.Error("auto mode should report not-done while a live session is anchored in the worktree")
	}
	if mock.removeCalled {
		t.Error("Remove() must not be called while a live session is anchored in the worktree")
	}
}

// TestRunDone_RefusesWhenAnchoredSessionLive: interactive `moai worktree done`
// must exit non-zero with the ANCHORED_SESSIONS_PRESENT sentinel instead of
// silently killing the anchored lane.
func TestRunDone_RefusesWhenAnchoredSessionLive(t *testing.T) {
	tree := t.TempDir()
	writeTreeRegistry(t, tree, []session.Entry{anchoredEntry(t, tree, os.Getpid())})

	origProvider := WorktreeProvider
	mock := &mockWorktreeProvider{
		worktrees: []git.Worktree{{Branch: "feature/SPEC-ANCHOR-001", Path: tree}},
	}
	WorktreeProvider = mock
	defer func() { WorktreeProvider = origProvider }()

	cmd := newDoneCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"feature/SPEC-ANCHOR-001"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("runDone must refuse (non-zero exit) while a live session is anchored in the worktree")
	}
	if !strings.Contains(err.Error(), "ANCHORED_SESSIONS_PRESENT") {
		t.Errorf("refusal must carry the ANCHORED_SESSIONS_PRESENT sentinel, got: %v", err)
	}
	if mock.removeCalled {
		t.Error("Remove() must not be called on refusal")
	}
}

// TestRunDone_ForceOverridesAnchorWarning: --force removes the tree but
// still tells the operator which live sessions it is cutting down.
func TestRunDone_ForceOverridesAnchorWarning(t *testing.T) {
	tree := t.TempDir()
	writeTreeRegistry(t, tree, []session.Entry{anchoredEntry(t, tree, os.Getpid())})

	origProvider := WorktreeProvider
	mock := &mockWorktreeProvider{
		worktrees: []git.Worktree{{Branch: "feature/SPEC-ANCHOR-001", Path: tree}},
	}
	WorktreeProvider = mock
	defer func() { WorktreeProvider = origProvider }()

	cmd := newDoneCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"--force", "feature/SPEC-ANCHOR-001"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("--force must remove despite the anchor, got error: %v", err)
	}
	if !mock.removeCalled {
		t.Error("Remove() must be called under --force")
	}
	if !strings.Contains(errOut.String(), "Warning") {
		t.Errorf("--force must warn about the anchored session(s), stderr: %q", errOut.String())
	}
}

// TestRunDone_RemovesWhenNoAnchoredSession: a tree without a tree-local
// registry (ordinary feature branches, never a lane home) disposes as before.
func TestRunDone_RemovesWhenNoAnchoredSession(t *testing.T) {
	tree := t.TempDir()

	origProvider := WorktreeProvider
	mock := &mockWorktreeProvider{
		worktrees: []git.Worktree{{Branch: "feature/SPEC-ANCHOR-001", Path: tree}},
	}
	WorktreeProvider = mock
	defer func() { WorktreeProvider = origProvider }()

	cmd := newDoneCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"feature/SPEC-ANCHOR-001"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("no anchor present — removal must proceed, got error: %v", err)
	}
	if !mock.removeCalled {
		t.Error("Remove() must be called when no anchor is present")
	}
}
