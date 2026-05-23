// Package testutil provides test-only utilities for the hook package, in
// particular for deterministically awaiting completion of side-effect
// goroutines launched by hook handlers per SPEC-V3R6-HOOK-ASYNC-EXPAND-001.
//
// Production code MUST NOT import this package. The build will not enforce
// that, but reviewers are expected to flag any non-_test.go import.
//
// REQ-HAE-006: hook test files MUST use WaitForAsync (not time.Sleep) to
// assert side-effect completion. AC-HAE-008 verifies the helper is used by
// ≥ 4 hook test files.
package testutil

import (
	"sync"
	"testing"
	"time"
)

// WaitForAsync blocks until wg.Done() has been called for every wg.Add(),
// or until deadline elapses. On deadline expiry it calls tb.Fatalf so the
// test fails deterministically. This is the canonical replacement for
// time.Sleep in tests that assert hook side-effects.
//
// Accepts testing.TB so that both *testing.T and *testing.B (benchmarks)
// can use the same helper.
//
// Typical usage:
//
//	wg := h.WaitGroup()      // package-internal accessor
//	out, err := h.Handle(ctx, input)
//	testutil.WaitForAsync(t, wg, 2*time.Second)
//	// now safe to read side-effect state (files, metrics, etc.)
//
// REQ-HAE-006 mandate: the helper uses select { case <-done: case <-After }
// (NOT time.Sleep) so the test wakes up as soon as the side-effect completes,
// which keeps test wall-time low without sacrificing determinism.
//
// @MX:ANCHOR: [AUTO] WaitForAsync is the canonical async-completion barrier for hook tests
// @MX:REASON: fan_in expected >= 4 (REQ-HAE-006 mandate, AC-HAE-008 verifies ≥4 test file usage)
func WaitForAsync(tb testing.TB, wg *sync.WaitGroup, deadline time.Duration) {
	tb.Helper()
	if wg == nil {
		tb.Fatalf("WaitForAsync: wg is nil")
		return
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return
	case <-time.After(deadline):
		tb.Fatalf("WaitForAsync: deadline %v exceeded before goroutines completed", deadline)
	}
}
