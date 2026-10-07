package atomicfile

import (
	"context"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// TestBreakerExcludesRivalsForTheWholeVerdictToDisposalSpan pins the P1
// review-gate finding on the rename-based break: RENAME is not a
// compare-and-swap. A rival that re-acquires AFTER the breaker's stale
// judgment but BEFORE its critical step has its fresh live lock MOVED by
// that step; the path is then free while the rival still holds its
// section, a third entrant acquires, and mutual exclusion is broken. The
// post-hoc byte-compare + restore cannot close an already-open window.
//
// The repro uses REAL actors: B and C are ordinary ClaimSection callers.
// The breaker's critical step (the disposal its implementation chooses) is
// interposed once; B's full reclaim attempt runs immediately BEFORE it and
// C's plain acquire immediately AFTER it. The gate's shape — "B still
// holds section; C acquired=true" — must be impossible: a breaker step may
// never free the path out from under a rival that acquired during the
// breaker's verdict-to-disposal span.
func TestBreakerExcludesRivalsForTheWholeVerdictToDisposalSpan(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "queue.lock")
	previousBootFixture(t, lockPath)

	var bAcquired, cAcquired atomic.Bool
	var bRelease func() error
	bAttempt := func() {
		// B: a full reclaim attempt (verify + break + claim) by a rival.
		release, err := ClaimSection(context.Background(), lockPath, 0o600, 4, 2*time.Millisecond)
		if err == nil {
			bAcquired.Store(true)
			bRelease = release
		}
	}
	cAttempt := func() {
		// C: a plain acquire while B holds the section. It must fail for as
		// long as B holds it — no breaker step may free the path under B.
		release, err := ClaimSection(context.Background(), lockPath, 0o600, 1, time.Millisecond)
		if err == nil {
			cAcquired.Store(true)
			_ = release()
		}
	}

	t.Cleanup(func() {
		if bRelease != nil {
			_ = bRelease()
		}
	})

	// Interpose the rivals around the breaker's critical disposal step.
	interposed := false
	prevRemove := sectionRemoveFn
	t.Cleanup(func() { sectionRemoveFn = prevRemove })
	sectionRemoveFn = func(path string) error {
		var err error
		if !interposed {
			interposed = true
			bAttempt()
			err = prevRemove(path)
			cAttempt()
		} else {
			err = prevRemove(path)
		}
		return err
	}

	BreakStaleLock(lockPath)

	if bAcquired.Load() && cAcquired.Load() {
		t.Fatal("C acquired while B held the section — the breaker's critical step freed the path under a live holder")
	}
}
