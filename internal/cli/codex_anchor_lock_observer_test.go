package cli

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/session"
)

func TestAnchorLockObservationUsesDirectoryIdentity(t *testing.T) {
	r := newAnchorRepo(t)
	tree := r.tree(t, "observed")
	other := r.tree(t, "foreign")
	reason := session.CodexAnchorLockReason("observed", os.Getpid(), "")
	lockAnchorTree(t, tree, reason)
	lockAnchorTree(t, other, "foreign lock without a pid")
	for _, path := range []string{tree, tree + string(os.PathSeparator) + "."} {
		// Git is the independent source of truth, including when it prints a
		// different spelling for the same directory supplied through -C.
		out := anchorGit(t, path, "worktree", "list", "--porcelain")
		found := false
		for _, lock := range session.ParseWorktreeLocks(out) {
			found = found || lock.Locked && lock.Reason == reason
		}
		if !found {
			t.Fatal("Git did not report the fixture's stored target lock")
		}
		got := observedLock(t, path)
		if !got.Locked || got.Reason != reason {
			t.Fatalf("observed target %q: %+v, want its actual lock %q", path, got, reason)
		}
	}
}

func TestAnchorLockObservationPreservesAbsentAndUnreadableStates(t *testing.T) {
	r := newAnchorRepo(t)
	tree := r.tree(t, "states")
	other := r.tree(t, "foreign")
	lockAnchorTree(t, other, session.CodexAnchorLockReason("foreign", os.Getpid(), ""))
	alias := tree + string(os.PathSeparator) + "."
	if got := observedLock(t, alias); got != (session.LockInfo{}) {
		t.Fatalf("unlocked target borrowed another directory's lock: %+v", got)
	}
	for _, reason := range []string{"", "manual lock without a pid"} {
		if reason == "" {
			anchorGit(t, tree, "worktree", "lock", tree)
		} else {
			lockAnchorTree(t, tree, reason)
		}
		got := observedLock(t, alias)
		if !got.Locked || got.Reason != reason {
			t.Fatalf("unreadable holder lock lost: got %+v, want locked reason %q", got, reason)
		}
		if pid, ok := session.LockReasonPID(got.Reason); ok {
			t.Fatalf("invented pid %d from unreadable reason %q", pid, got.Reason)
		}
		anchorGit(t, tree, "worktree", "unlock", tree)
	}
}
