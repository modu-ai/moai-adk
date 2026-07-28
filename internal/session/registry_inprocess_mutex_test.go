package session

// Structural in-process serialization probe for AC-CFS3-002
// (SPEC-V3R6-CI-FLAKY-STABILIZE-003).
//
// This test verifies the in-process mutex added by the SPEC serializes
// withLock critical sections, AND that the AC is falsifiable: when the mutex
// is disabled (MOAI_CFS3_FALSIFIABILITY=1) the probe must observe concurrent
// residency >= 2. The probe (withLockProbe) brackets the ENTIRE withLock body
// (including the NB-flock retry loop), so it measures "goroutines inside
// withLock" rather than "flock holders" — the OS serializes flock, so a
// flock-scoped probe would be <= 1 regardless of the mutex and the AC would
// be vacuous.

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// runContention drives N goroutines × M Register calls against a single
// registry and returns the maximum concurrent withLock residency observed by
// the probe during the run. The probe is reset at entry so the returned max
// reflects only this run.
func runContention(t *testing.T, r *Registry, workers, perWorker int) int64 {
	t.Helper()
	withLockProbe.reset()

	done := make(chan struct{})
	errCh := make(chan error, workers*perWorker)
	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer func() { done <- struct{}{} }()
			for i := 0; i < perWorker; i++ {
				id := fmt.Sprintf("uuid-probe-%d-%d", workerID, i)
				if err := r.Register(id, "SPEC-PROBE", "run"); err != nil {
					errCh <- err
					return
				}
			}
		}(w)
	}
	for w := 0; w < workers; w++ {
		<-done
	}
	close(errCh)
	for err := range errCh {
		t.Fatalf("Register under contention: %v", err)
	}
	return withLockProbe.max()
}

// TestRegistryWithLockInProcessSerialization is the AC-CFS3-002 structural
// test. The in-process mutex MUST keep concurrent withLock-body residency at
// <= 1 under heavy in-process contention (10 goroutines x 1000 iterations =
// 10000 acquisitions).
func TestRegistryWithLockInProcessSerialization(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in-process serialization probe in -short mode")
	}
	r, _ := newTestRegistry(t)
	r = r.WithLockTimeout(60 * time.Second) // operational headroom; mutex is the correctness mechanism

	// AC-CFS3-002 recommends 10000 total acquisitions (10 x 1000). Reduced to
	// 1000 (10 x 100) here because under -race the mutex serializes every
	// acquisition, and 10000 sequential atomic file writes (temp + rename)
	// exceed 8 minutes under -race — impractical for CI. The serialization
	// invariant (max <= 1) is count-independent: the mutex guarantees it
	// whether the count is 10 or 10000, so 1000 acquisitions across 10
	// concurrent goroutines is more than sufficient to demonstrate it. This
	// matches the AC-CFS3-001 TestRegisterSessionConcurrent scale.
	const (
		workers   = 10
		perWorker = 100
	)

	maxObserved := runContention(t, r, workers, perWorker)
	t.Logf("withLock concurrent residency (mutex enabled): max=%d (workers=%d perWorker=%d total=%d)",
		maxObserved, workers, perWorker, workers*perWorker)

	// AC-CFS3-002: with the in-process mutex, at most one goroutine is
	// resident inside withLock at any instant.
	if maxObserved > 1 {
		t.Errorf("AC-CFS3-002 FAILED: withLock saw %d concurrent residents; want <= 1 (in-process mutex not serializing)",
			maxObserved)
	}
}

// TestRegistryWithLockInProcessSerializationFalsifiability is the
// falsifiability sub-case (AC-CFS3-002 both-directions requirement). Under
// MOAI_CFS3_FALSIFIABILITY=1 the in-process mutex is disabled, so the probe
// MUST observe >= 2 concurrent residents under sufficient contention — proving
// the AC is not vacuous. This sub-case runs ONLY under the env flag (local
// verification; CI never sets it). Per verification-claim-integrity §1.1
// surface 3, if >= 2 is not observed the AC is UNMET.
func TestRegistryWithLockInProcessSerializationFalsifiability(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping falsifiability probe in -short mode")
	}
	if os.Getenv("MOAI_CFS3_FALSIFIABILITY") != "1" {
		t.Skip("MOAI_CFS3_FALSIFIABILITY!=1; skipping falsifiability sub-case (local verification only)")
	}

	r, _ := newTestRegistry(t)
	r = r.WithLockTimeout(60 * time.Second)

	// The mutex is DISABLED here (env=1), so a large iteration count would
	// re-trigger the pathological NB-flock starvation the SPEC fixes. 500
	// acquisitions (10 x 50) is more than enough to reliably observe max >= 2
	// (concurrent residency appears within the first few entries) while
	// completing promptly under -race even without the mutex.
	const (
		workers   = 10
		perWorker = 50
	)

	maxObserved := runContention(t, r, workers, perWorker)
	t.Logf("withLock concurrent residency (mutex DISABLED via MOAI_CFS3_FALSIFIABILITY=1): max=%d (workers=%d perWorker=%d total=%d)",
		maxObserved, workers, perWorker, workers*perWorker)

	// AC-CFS3-002 falsifiability direction: with the mutex disabled,
	// concurrent residency >= 2 MUST be observable under sufficient contention.
	if maxObserved < 2 {
		t.Errorf("AC-CFS3-002 FALSIFIABILITY UNMET: with mutex disabled, withLock saw max=%d concurrent residents; want >= 2 under %d acquisitions (AC may be vacuous)",
			maxObserved, workers*perWorker)
	}
}
