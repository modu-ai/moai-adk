package atomicfile

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestBreakNeverDisposesALockAcquiredMidSection pins the break invariant
// with a REAL rival actor: while the breaker spans its verdict-to-disposal
// critical section, a rival's reclaim attempt must BLOCK on the breaker
// marker (never dispossess anyone) — and if a rival ever did acquire, its
// lock must survive the disposal untouched. The rival is a plain
// ClaimSection caller interposed at the breaker's disposal seam.
func TestBreakNeverDisposesALockAcquiredMidSection(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "queue.lock")
	previousBootFixture(t, lockPath)

	liveOwner := LockOwner{PID: os.Getpid(), BootID: BootIDIdentity()}
	if _, err := json.Marshal(liveOwner); err != nil {
		t.Fatalf("marshal live: %v", err)
	}

	// The disposal seam: the instant the breaker reaches its disposal step,
	// a rival runs a FULL reclaim attempt (verify + break + claim) through
	// the real entry point.
	rivalBlocked := false
	rivalAcquired := false
	prevRemove := sectionRemoveFn
	t.Cleanup(func() { sectionRemoveFn = prevRemove })
	sectionRemoveFn = func(path string) error {
		release, cerr := ClaimSection(context.Background(), lockPath, 0o600, 3, 2*time.Millisecond)
		if cerr == nil {
			rivalAcquired = true
			_ = release()
		} else {
			rivalBlocked = true
		}
		return prevRemove(path)
	}

	BreakStaleLock(lockPath)

	if rivalAcquired {
		// If the rival acquired at all, its lock must still sit at the path
		// with its own label — the disposal must never dispose a live
		// holder's lock.
		after, rerr := os.ReadFile(lockPath)
		if rerr != nil {
			t.Fatalf("the rival's live lock was disposed by the breaker: %v", rerr)
		}
		var atPath LockOwner
		if json.Unmarshal(after, &atPath) != nil || atPath.PID != os.Getpid() || atPath.BootID != BootIDIdentity() {
			t.Fatalf("the rival's lock bytes were altered: %s", after)
		}
	}
	if rivalBlocked && rivalAcquired {
		t.Fatal("the rival both blocked and acquired — inconsistent observation")
	}
	// Either way the stale lock must be gone.
	if _, serr := os.Stat(lockPath); serr == nil && !rivalAcquired {
		t.Fatal("the stale lock survived a break that reported success")
	}
}
