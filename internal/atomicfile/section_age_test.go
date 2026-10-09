package atomicfile

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLiveOwnerOfAgeExceededLockStillBlocks (AC-018, D40): a lock whose
// recorded owner is OLD but verifiably ALIVE must keep blocking — the
// break fires on verified owner death and NOTHING else; no age-based break
// may exist (an age-only break could discard a live slow owner's committed
// mutation).
func TestLiveOwnerOfAgeExceededLockStillBlocks(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "queue.lock")

	// The lock: THIS process's live label, but CreatedAt far in the past —
	// the age-exceeded shape that must never break.
	owner := LockOwner{
		PID:       os.Getpid(),
		BootID:    BootIDIdentity(),
		CreatedAt: time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339),
	}
	if err := writeLiveFixture(lockPath, owner); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	if BreakStaleLock(lockPath) {
		t.Fatal("a live owner's age-exceeded lock was broken — the invariant is verified death, nothing else")
	}
	if _, serr := os.Stat(lockPath); serr != nil {
		t.Fatal("the live lock did not survive the break attempt")
	}

	// And a claimant still blocks through its budget: no silent takeover.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	if _, err := ClaimSection(ctx, lockPath, 0o600, 1, 2*time.Millisecond); err == nil {
		t.Fatal("a claim succeeded against a LIVE age-exceeded lock — the live owner must block")
	}
}
