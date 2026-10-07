package atomicfile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestBreakNeverDisposesALockAcquiredMidSection pins review-gate finding
// (P1): verify-fresh + break must be ONE critical section — the window
// between the reclaimer's last fresh read and its remove must not be wide
// enough for a rival to acquire a NEW lock whose file the late remove then
// deletes. The deterministic repro drives the removal seam: the moment the
// reclaimer reaches its remove, a rival reclaimer has already completed a
// full reclaim cycle (removed the dead lock, claimed, labelled with a LIVE
// pid). Whatever the reclaimer does then must not dispose of that live
// lock: the path must still hold a labelled live lock after the call.
func TestBreakNeverDisposesALockAcquiredMidSection(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "queue.lock")
	previousBootFixture(t, lockPath)

	liveOwner := LockOwner{PID: os.Getpid(), BootID: BootIDIdentity()}
	liveBytes, err := json.Marshal(liveOwner)
	if err != nil {
		t.Fatalf("marshal live: %v", err)
	}

	// The rival's full reclaim cycle, run the instant the breaker reaches
	// its disposal step. The rival's own removal of the dead lock tolerates
	// absence: under the rename-based break the dead lock may already be
	// detached from the path when the rival cycles, and the rival then
	// simply claims the free path.
	rivalReclaims := func() {
		if rmErr := os.Remove(lockPath); rmErr != nil && !os.IsNotExist(rmErr) {
			t.Errorf("rival remove: %v", rmErr)
		}
		if err := Claim(lockPath, 0o600); err != nil {
			t.Errorf("rival claim: %v", err)
			return
		}
		if werr := os.WriteFile(lockPath, liveBytes, 0o600); werr != nil {
			t.Errorf("rival label: %v", werr)
		}
	}

	prevRemove := sectionRemoveFn
	t.Cleanup(func() { sectionRemoveFn = prevRemove })
	sectionRemoveFn = func(path string) error {
		rivalReclaims()
		return prevRemove(path) // the breaker's own removal, after the rival acquired
	}

	BreakStaleLock(lockPath)

	after, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("the live lock acquired mid-section was disposed: %v", err)
	}
	if string(after) != string(liveBytes) {
		t.Fatalf("the live lock's bytes were altered by the late disposal: %s", after)
	}
}
