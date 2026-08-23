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

// TestFindRegistryUpwardStopsAtHome pins the t168 boundary: the walk must
// stop at the user's home directory rather than climb past it into
// ~/.moai/state/active-sessions.json. Without the guard, a session whose
// working directories sit under $HOME but outside any checkout relocates its
// entry inside the GLOBAL registry — a write to shared state on behalf of a
// project that was never found.
//
// The layout is home-shaped on purpose; every assertion compares paths inside
// a directory the test owns, so it holds on every platform without reading
// the machine's real $HOME.
func TestFindRegistryUpwardStopsAtHome(t *testing.T) {
	t.Run("a registry at the home directory is not claimed", func(t *testing.T) {
		home := t.TempDir()
		writeRelocateRegistry(t, home, "sess-t168-home", home)
		start := mustMkdirAllHook(t, filepath.Join(home, "scratch", "nested"))

		if got, ok := findRegistryUpwardFrom(start, home); ok {
			t.Fatalf("walk claimed %q; it must stop at the home boundary %q", got, home)
		}
	})

	t.Run("a registry above the home directory is not reached", func(t *testing.T) {
		above := t.TempDir()
		writeRelocateRegistry(t, above, "sess-t168-above", above)
		home := mustMkdirAllHook(t, filepath.Join(above, "home", "user"))
		start := mustMkdirAllHook(t, filepath.Join(home, "scratch"))

		if got, ok := findRegistryUpwardFrom(start, home); ok {
			t.Fatalf("walk climbed past home to %q", got)
		}
	})

	t.Run("a registry below the home directory is still found", func(t *testing.T) {
		home := t.TempDir()
		checkout := mustMkdirAllHook(t, filepath.Join(home, "projects", "app"))
		writeRelocateRegistry(t, checkout, "sess-t168-project", checkout)
		start := mustMkdirAllHook(t, filepath.Join(checkout, "internal", "pkg"))

		got, ok := findRegistryUpwardFrom(start, home)
		if !ok {
			t.Fatal("walk found no registry; the guard must not block descendants of home")
		}
		want := filepath.Join(resolveSymlinks(checkout), session.DefaultRegistryPath)
		if got != want {
			t.Fatalf("got %q, want the project registry %q", got, want)
		}
	})

	t.Run("an unresolvable home boundary leaves the walk unbounded", func(t *testing.T) {
		root := t.TempDir()
		writeRelocateRegistry(t, root, "sess-t168-nohome", root)
		start := mustMkdirAllHook(t, filepath.Join(root, "nested"))

		if _, ok := findRegistryUpwardFrom(start, ""); !ok {
			t.Fatal("walk found nothing with an empty home; the guard must be inert, not blocking")
		}
	})
}

// mustMkdirAllHook is a t.Fatal-on-error mkdir helper for the cases above.
func mustMkdirAllHook(t *testing.T, dir string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	return dir
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
