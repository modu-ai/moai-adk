package claim

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// t2072 (RC40 carried-over review finding, claim/next.go:160): BindWorktree
// reserved the path before Update verified the caller's ownership, so a caller
// that did not hold the Issue left a reservation behind and the real holder was
// then refused its own path.
func TestBindWorktreeRefusedForAForeignCallerLeavesNoReservation(t *testing.T) {
	wt := realWT(t, "owned")
	sa, sb := twoStores(t)
	ra, err := sa.Acquire(ref(t, 1), "w1", "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	rb, err := sb.Acquire(ref(t, 2), "w2", "d2", "run-1")
	if err != nil {
		t.Fatal(err)
	}

	// a caller that is not the holder of issue 1 asks for the path
	if _, err := sb.BindWorktree(ref(t, 1), "intruder", ra.Epoch, wt); err == nil {
		t.Fatal("a foreign caller bound the worktree")
	}
	// a stale epoch of the real holder is refused as well
	if _, err := sa.BindWorktree(ref(t, 1), "w1", ra.Epoch+99, wt); err == nil {
		t.Fatal("a wrong epoch bound the worktree")
	}

	if _, err := sa.GetDoc("worktree", WorktreeKey(wt)); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("a refused bind left a reservation behind: %v", err)
	}
	// the refused attempts must not take the path away from other Issues
	if _, err := sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, wt); err != nil {
		t.Errorf("a refused bind blocked an honest Issue: %v", err)
	}
}
