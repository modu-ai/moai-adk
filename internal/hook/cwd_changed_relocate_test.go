package hook

// cwd_changed_relocate_test.go — t74: the CwdChanged hook relocates the
// session's registry entry CWD so anchor detection covers sessions that
// entered a worktree mid-session (their entry otherwise keeps the
// launch-time cwd forever).

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/session"
)

// writeRelocateRegistry seeds <root>/.moai/state/active-sessions.json with
// one entry whose CWD is cwd.
func writeRelocateRegistry(t *testing.T, root, sessionID, cwd string) session.Entry {
	t.Helper()
	host, _ := os.Hostname()
	now := time.Now().UTC()
	entry := session.Entry{
		SessionID:     sessionID,
		SpecID:        "(none)",
		Phase:         "(none)",
		StartedAt:     now,
		LastHeartbeat: now,
		PID:           os.Getpid(),
		Host:          host,
		CWD:           cwd,
	}
	dir := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	data, err := json.Marshal([]session.Entry{entry})
	if err != nil {
		t.Fatalf("marshal entries: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "active-sessions.json"), data, 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	return entry
}

// readFirstCWD re-reads the registry at root and returns the entry's CWD.
func readFirstCWD(t *testing.T, root, sessionID string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".moai", "state", "active-sessions.json"))
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	var entries []session.Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("unmarshal registry: %v", err)
	}
	for _, e := range entries {
		if e.SessionID == sessionID {
			return e.CWD
		}
	}
	t.Fatalf("session %s not found in registry at %s", sessionID, root)
	return ""
}

// TestCwdChangedHandler_RelocatesRegistryCwd is the t74 core: entering a
// worktree mid-session must move the entry's CWD to the new directory, in
// the registry the entry actually lives in (here: the launch checkout's).
func TestCwdChangedHandler_RelocatesRegistryCwd(t *testing.T) {
	primary := t.TempDir()
	tree := t.TempDir()
	const sid = "sess-t74-relocate"
	writeRelocateRegistry(t, primary, sid, primary)

	h := NewCwdChangedHandler()
	input := &HookInput{SessionID: sid, CWD: tree, OldCwd: primary, NewCwd: tree}
	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if got := readFirstCWD(t, primary, sid); got != tree {
		t.Fatalf("registry CWD after entry = %q, want the new worktree %q", got, tree)
	}
}

// TestCwdChangedHandler_RelocateIsSymmetricOnExit: leaving the worktree
// moves the CWD back, which also un-blocks disposal of the tree the session
// left (its entry no longer points into that tree).
func TestCwdChangedHandler_RelocateIsSymmetricOnExit(t *testing.T) {
	primary := t.TempDir()
	tree := t.TempDir()
	const sid = "sess-t74-exit"
	writeRelocateRegistry(t, primary, sid, tree) // entry pointing into the tree

	h := NewCwdChangedHandler()
	input := &HookInput{SessionID: sid, CWD: primary, OldCwd: tree, NewCwd: primary}
	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if got := readFirstCWD(t, primary, sid); got != primary {
		t.Fatalf("registry CWD after exit = %q, want the original %q", got, primary)
	}
}

// TestCwdChangedHandler_RelocateNoRegistryNoOp: with no registry anywhere
// up the candidate roots, the hook stays a silent no-op (fail-open).
func TestCwdChangedHandler_RelocateNoRegistryNoOp(t *testing.T) {
	h := NewCwdChangedHandler()
	input := &HookInput{SessionID: "sess-t74-none", CWD: t.TempDir(), NewCwd: t.TempDir()}
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil output")
	}
}
