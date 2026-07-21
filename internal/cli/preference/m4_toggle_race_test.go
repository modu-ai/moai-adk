package preference

// M4 toggle TOCTOU repro test (SPEC-CLIFIX-CONCURRENCY-001 REQ-CONC-001-005 /
// AC-CONC-001-005).
//
// This file references TogglePersonalization, the M4 locked flip entry point.
// Against pre-M4 (c31db9e2b) the symbol is undefined → compile failure (RED).
// After M4 lands TogglePersonalization with serialization, concurrent flips
// preserve parity (GREEN).

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestToggleRace_ConcurrentFlipsPreserveParity drives N concurrent
// TogglePersonalization calls starting from the active state (no sentinel) and
// asserts the final state reflects the parity: an even number of flips returns
// to active, an odd number lands on disabled. A read-then-flip without
// serialization loses updates (two concurrent toggles both read "active" and
// both disable → net disabled instead of the correct "back to active"), so this
// test fails unless the flip is serialized.
func TestToggleRace_ConcurrentFlipsPreserveParity(t *testing.T) {
	t.Run("even flips return to active", func(t *testing.T) {
		root := t.TempDir()
		// Start active (no sentinel). NewFileStore-equivalent state is ensured
		// by t.TempDir being empty.
		if IsPersonalizationDisabled(root) {
			t.Fatalf("precondition: sentinel should be absent in a fresh tempdir")
		}

		const N = 10 // even
		var toggleErrs atomic.Value
		var wg sync.WaitGroup
		barrier := make(chan struct{})
		start := make(chan struct{})
		var ready atomic.Int32
		for i := 0; i < N; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ready.Add(1)
				<-start
				// A swallowed error here is indistinguishable from a lost
				// update: the flip never happens and parity breaks with no
				// explanation. Record it so the failure is attributable.
				if _, err := TogglePersonalization(root, time.Now()); err != nil {
					toggleErrs.Store(err)
				}
				barrier <- struct{}{} // signal done (drained below)
			}()
		}
		// Wait until all goroutines are parked at <-start, then release.
		for ready.Load() < int32(N) {
			time.Sleep(time.Millisecond)
		}
		close(start)
		// Drain the done signals.
		for i := 0; i < N; i++ {
			<-barrier
		}
		wg.Wait()

		if err, ok := toggleErrs.Load().(error); ok {
			t.Fatalf("TogglePersonalization returned an error; a failed flip is a lost flip: %v", err)
		}
		if IsPersonalizationDisabled(root) {
			t.Fatalf("after %d (even) concurrent flips from active, expected ACTIVE (sentinel absent); got DISABLED — a concurrent toggle was lost (read-then-flip TOCTOU)", N)
		}
	})

	t.Run("odd flips land on disabled", func(t *testing.T) {
		root := t.TempDir()
		if IsPersonalizationDisabled(root) {
			t.Fatalf("precondition: sentinel should be absent in a fresh tempdir")
		}

		const N = 7 // odd
		var toggleErrs atomic.Value
		var wg sync.WaitGroup
		barrier := make(chan struct{})
		start := make(chan struct{})
		var ready atomic.Int32
		for i := 0; i < N; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ready.Add(1)
				<-start
				if _, err := TogglePersonalization(root, time.Now()); err != nil {
					toggleErrs.Store(err)
				}
				barrier <- struct{}{}
			}()
		}
		for ready.Load() < int32(N) {
			time.Sleep(time.Millisecond)
		}
		close(start)
		for i := 0; i < N; i++ {
			<-barrier
		}
		wg.Wait()

		if err, ok := toggleErrs.Load().(error); ok {
			t.Fatalf("TogglePersonalization returned an error; a failed flip is a lost flip: %v", err)
		}
		if !IsPersonalizationDisabled(root) {
			t.Fatalf("after %d (odd) concurrent flips from active, expected DISABLED (sentinel present); got ACTIVE — a concurrent toggle was lost (read-then-flip TOCTOU)", N)
		}
	})
}
