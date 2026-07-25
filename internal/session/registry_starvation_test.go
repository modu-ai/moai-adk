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
