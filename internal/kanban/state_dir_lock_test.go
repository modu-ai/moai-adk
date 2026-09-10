// state_dir_lock_test.go — relocating a state directory while the queue lock
// inside it is held (SPEC-TODO-SQLITE-001 M6, REQ-TOSQ-015, design.md R6).
//
// POSIX and Windows disagree here, and the disagreement is the reason this
// test states a platform-NEUTRAL invariant rather than a specific outcome.
// Renaming a directory containing an open, locked file is permitted on POSIX
// and refused on Windows (LockFileEx holds a sharing violation on the open
// handle). A test asserting "the relocation succeeds" would be a Windows
// regression waiting to be written down as correct; one asserting "it is
// refused" would fail on every developer machine in this repository.
//
// What must hold on BOTH is narrower and is what the operator actually cares
// about: whichever branch runs, the queue is still readable, every card is
// still there, and no half-moved debris is left behind. That is asserted here.
// The Windows BEHAVIORAL verdict comes from CI (acceptance.md D.3); local
// evidence stops at compile plus this invariant.
package kanban

import (
	"os"
	"path/filepath"
	"testing"
)

// TestStateDirRelocationUnderHeldLock — the relocation runs while a foreign
// holder owns the queue lock inside the legacy directory.
func TestStateDirRelocationUnderHeldLock(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLegacyStateDir(t, root, 2)

	// A foreign holder takes the queue lock where it currently lives.
	legacyQueue := filepath.Join(LegacyStateDirForRoot(root), backlogFileName)
	legacyStore := NewBacklogStore(legacyQueue)
	held, err := acquireBoardLockImpl(legacyStore.LockPath())
	if err != nil {
		t.Fatalf("take the foreign lock: %v", err)
	}
	released := false
	t.Cleanup(func() {
		if !released {
			_ = held.release()
		}
	})

	path := BacklogPathForRootAdopting(root)
	dir := filepath.Dir(path)

	// Branch A (POSIX) or branch B (Windows) — either is a correct outcome for
	// its platform, and the invariants below bind both.
	relocated := dir == StateDirForRoot(root)
	if !relocated && dir != LegacyStateDirForRoot(root) {
		t.Fatalf("resolution landed at %q, which is neither directory", dir)
	}
	t.Logf("relocation under a held lock: relocated=%v (dir=%s)", relocated, filepath.Base(dir))

	// INVARIANT 1 — the queue is readable and complete, whichever branch ran.
	rec, err := NewBacklogStore(path).LoadPure()
	if err != nil {
		t.Fatalf("queue unreadable after the attempt: %v", err)
	}
	if len(rec.Items) != 4 {
		t.Fatalf("items = %d, want 4 — no card may be lost on either branch", len(rec.Items))
	}

	// INVARIANT 2 — no half-moved debris. Exactly ONE of the two directories
	// holds the queue; a state where both do is the partial rename this
	// invariant exists to catch.
	currentHasQueue := fileExists(filepath.Join(StateDirForRoot(root), backlogFileName))
	legacyHasQueue := fileExists(legacyQueue)
	if currentHasQueue == legacyHasQueue {
		t.Fatalf("queue present in current=%v legacy=%v — exactly one must hold it",
			currentHasQueue, legacyHasQueue)
	}

	// INVARIANT 3 — the session records went wherever the queue went. The
	// registry and the queue are one channel; splitting them across two
	// directories is the failure mode the directory-level rename prevents.
	records, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read resolved dir: %v", err)
	}
	jsonRecords := 0
	for _, e := range records {
		if filepath.Ext(e.Name()) == ".json" && e.Name() != backlogFileName {
			jsonRecords++
		}
	}
	if jsonRecords != 2 {
		t.Errorf("session records beside the queue = %d, want 2 — the registry must travel with it", jsonRecords)
	}

	// Release, then retry: on a platform that refused, the next open completes
	// the move; on one that already moved, the retry is a no-op. Both end with
	// the queue readable at whatever the resolver now returns.
	if err := held.release(); err != nil {
		t.Fatalf("release the foreign lock: %v", err)
	}
	released = true

	retryPath := BacklogPathForRootAdopting(root)
	retryRec, err := NewBacklogStore(retryPath).LoadPure()
	if err != nil {
		t.Fatalf("queue unreadable after the retry: %v", err)
	}
	if len(retryRec.Items) != 4 {
		t.Fatalf("items after retry = %d, want 4", len(retryRec.Items))
	}
	if filepath.Dir(retryPath) != StateDirForRoot(root) {
		t.Errorf("after the lock was released the retry resolved to %q, want the current directory",
			filepath.Dir(retryPath))
	}
}
