package session

// Lock-contention characterization (SPEC-CI-FLAKY-STABILIZE-001, REQ-CFS-015/016).
//
// HONEST SCOPE — this is a CHARACTERIZATION test, not a reproduction. The CI
// failure it relates to (TestRegisterSessionConcurrent exceeding a 60s
// per-acquisition budget with ErrLockTimeout) does not reproduce locally. This
// test does not attempt to force starvation; it measures the per-acquisition
// wait distribution under contention so the fixed-interval and jittered-backoff
// retry loops can be compared on the same machine.
//
// What is measured is the wall time of a full Register call — lock acquisition
// plus read, mutate, and atomic write. Acquisition wait is the dominant and the
// only term this SPEC changes, but the number is a proxy, not an isolated
// acquisition timer.
//
// Only ONE assertion is made (a generous ceiling on the single worst
// acquisition). Asserting on p50/p95 would make this test itself a flaky CI
// contributor, which is precisely the class of defect this SPEC removes.

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"
)

// TestRegisterStarvationCharacterization reports p50/p95/max per-acquisition
// wait under a multi-goroutine pile-up on the registry's advisory lock.
func TestRegisterStarvationCharacterization(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping lock-contention characterization in -short mode")
	}

	const (
		workers   = 8
		perWorker = 25
		// Generous single-acquisition ceiling. The point is to catch a
		// starved goroutine (seconds-scale outlier), not to police normal
		// scheduling noise.
		maxWaitCeiling = 15 * time.Second
		// Well above maxWaitCeiling so a starved acquisition surfaces as a
		// slow sample rather than as ErrLockTimeout.
		lockTimeout = 60 * time.Second
	)

	r, _ := newTestRegistry(t)
	r = r.WithLockTimeout(lockTimeout)

	var (
		mu      sync.Mutex
		samples []time.Duration
		wg      sync.WaitGroup
	)
	errCh := make(chan error, workers)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			local := make([]time.Duration, 0, perWorker)
			for i := 0; i < perWorker; i++ {
				start := time.Now()
				err := r.Register(fmt.Sprintf("uuid-starve-%d-%d", workerID, i), "SPEC-STARVE", "run")
				local = append(local, time.Since(start))
				if err != nil {
					errCh <- err
					break
				}
			}
			mu.Lock()
			samples = append(samples, local...)
			mu.Unlock()
		}(w)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("Register under contention: %v", err)
	}

	if len(samples) == 0 {
		t.Fatal("no acquisition samples collected")
	}
	sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })

	quantile := func(q float64) time.Duration {
		idx := int(q * float64(len(samples)-1))
		return samples[idx]
	}
	p50, p95, max := quantile(0.50), quantile(0.95), samples[len(samples)-1]

	t.Logf("per-acquisition wait under contention (workers=%d perWorker=%d n=%d): p50=%v p95=%v max=%v",
		workers, perWorker, len(samples), p50, p95, max)

	if max > maxWaitCeiling {
		t.Errorf("worst single acquisition %v exceeds ceiling %v — a goroutine was starved", max, maxWaitCeiling)
	}
}

// TestWithLockTimeoutContract pins the REQ-CFS-014 error contract: when the
// budget elapses with the lock held elsewhere, withLock still returns an error
// wrapping ErrLockTimeout, and errors.Is still recognizes it. Swapping the
// fixed retry sleep for jittered backoff must not change either fact.
func TestWithLockTimeoutContract(t *testing.T) {
	r, _ := newTestRegistry(t)

	// Hold the lock via a separate open file description so the registry's
	// non-blocking acquisition genuinely fails on every retry.
	blocker := newRegistryLock()
	if err := blocker.acquire(r.path + ".lock"); err != nil {
		t.Fatalf("test blocker could not take the lock: %v", err)
	}
	defer func() { _ = blocker.release() }()

	const budget = 80 * time.Millisecond
	start := time.Now()
	err := r.WithLockTimeout(budget).Register("uuid-timeout", "SPEC-TIMEOUT", "run")
	elapsed := time.Since(start)

	if !errors.Is(err, ErrLockTimeout) {
		t.Fatalf("Register while the lock is held: got %v, want an error wrapping ErrLockTimeout", err)
	}
	// A sleep that overshot the deadline would stretch this far past budget
	// (REQ-CFS-011 / EC-1).
	if elapsed > time.Second {
		t.Errorf("timeout path took %v against an %v budget — backoff is not clamped to the deadline",
			elapsed, budget)
	}
	t.Logf("timeout path returned %q after %v", err, elapsed)
}

// TestLockRetryDelayBoundsAndJitter covers REQ-CFS-010 (jitter is real) and
// REQ-CFS-011 (the ceiling and the deadline clamp both hold).
func TestLockRetryDelayBoundsAndJitter(t *testing.T) {
	// The exponential ceiling never exceeds lockBackoffCap, at any attempt.
	for attempt := 0; attempt < 20; attempt++ {
		d := lockRetryDelay(attempt, time.Hour)
		if d <= 0 || d > lockBackoffCap {
			t.Fatalf("attempt %d: delay %v outside (0, %v]", attempt, d, lockBackoffCap)
		}
	}

	// A sleep never exceeds the remaining budget (EC-1).
	for _, remaining := range []time.Duration{time.Nanosecond, time.Microsecond, time.Millisecond} {
		for attempt := 0; attempt < 20; attempt++ {
			if d := lockRetryDelay(attempt, remaining); d > remaining {
				t.Fatalf("attempt %d: delay %v exceeds remaining budget %v", attempt, d, remaining)
			}
		}
	}

	// An exhausted budget yields an immediate retry, so the caller's deadline
	// check runs instead of sleeping past it.
	if d := lockRetryDelay(3, 0); d != 0 {
		t.Errorf("zero budget: got %v, want 0", d)
	}
	if d := lockRetryDelay(3, -time.Second); d != 0 {
		t.Errorf("negative budget: got %v, want 0", d)
	}

	// Jitter must actually vary: the fixed-interval implementation this replaces
	// would collapse to a single value here.
	seen := map[time.Duration]bool{}
	for i := 0; i < 200; i++ {
		seen[lockRetryDelay(4, time.Hour)] = true
	}
	if len(seen) < 2 {
		t.Errorf("delay is constant across 200 samples (%v) — jitter is absent", seen)
	}
}
