package git

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestWorktreeRemove_LockedTreeMapsToSentinel (t41 b): a session-held lock
// (`git worktree lock`) makes `git worktree remove` fail even with a single
// --force. The failure must map to ErrWorktreeLocked and keep the lock
// reason visible, so the CLI can answer "why did nothing happen" instead of
// failing silently (measured 2026-08-15: rc=128, 0 lines of stderr, three
// wasted retries by a blind caller).
func TestWorktreeRemove_LockedTreeMapsToSentinel(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	wtPath := filepath.Join(resolveSymlinks(t, t.TempDir()), "wt-locked")
	if err := wm.Add(wtPath, "feat-locked"); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "worktree", "lock", "--reason", "claude session (pid 4242)", wtPath)

	err := wm.Remove(wtPath, false)
	if !errors.Is(err, ErrWorktreeLocked) {
		t.Fatalf("Remove(force=false) on a locked tree = %v, want ErrWorktreeLocked", err)
	}
	if !strings.Contains(err.Error(), "claude session (pid 4242)") {
		t.Errorf("lock reason (git stderr) must stay visible, got: %v", err)
	}

	// One --force is NOT enough for a locked tree (git semantics), so the
	// mapping must hold on the forced path too — that is what lets the CLI
	// offer the unlock / double-force guidance instead of a silent retry.
	err = wm.Remove(wtPath, true)
	if !errors.Is(err, ErrWorktreeLocked) {
		t.Fatalf("Remove(force=true) on a locked tree = %v, want ErrWorktreeLocked", err)
	}

	// Fixture cleanup: unlock and remove so nothing leaks between tests.
	runGit(t, dir, "worktree", "unlock", wtPath)
	runGit(t, dir, "worktree", "remove", "--force", wtPath)
}

// TestWorktreeRemoveTimeoutBudget (t41 b2): removing a large worktree can
// take far longer than 10s — while the raw `git worktree remove` the user
// runs afterwards has no deadline and succeeds. That cap is the structural
// divergence that made `done` fail on trees the raw command handles.
func TestWorktreeRemoveTimeoutBudget(t *testing.T) {
	if removeTimeout < 2*time.Minute {
		t.Errorf("removeTimeout = %v, want >= 2m (2026-08-15 incident: 10s killed git mid-removal)", removeTimeout)
	}
}
