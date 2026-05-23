// Package testutil — wait_async_test.go
// Self-test for the WaitForAsync helper. SPEC-V3R6-HOOK-ASYNC-EXPAND-001 M1.
package testutil

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestWaitForAsync_CompletesBeforeDeadline verifies the happy path: the
// helper returns as soon as the WaitGroup is drained.
func TestWaitForAsync_CompletesBeforeDeadline(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
	}()

	start := time.Now()
	WaitForAsync(t, &wg, 2*time.Second)
	elapsed := time.Since(start)

	if elapsed > 500*time.Millisecond {
		t.Errorf("WaitForAsync took %v, expected < 500ms (helper should wake on Done)", elapsed)
	}
}

// TestWaitForAsync_MultipleGoroutines verifies the helper handles parallel
// Done() calls correctly.
func TestWaitForAsync_MultipleGoroutines(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	const n = 10
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			time.Sleep(5 * time.Millisecond)
		}()
	}

	WaitForAsync(t, &wg, 2*time.Second)
}

// TestWaitForAsync_NilWaitGroup_Fatals verifies the helper rejects a nil
// WaitGroup. We use a fake testing.TB to capture the Fatalf call without
// failing this test.
func TestWaitForAsync_NilWaitGroup_Fatals(t *testing.T) {
	t.Parallel()

	fake := &fakeTB{}
	WaitForAsync(fake, nil, 1*time.Second)
	if !fake.fatalCalled.Load() {
		t.Errorf("WaitForAsync did not call tb.Fatalf on nil wg")
	}
}

// TestWaitForAsync_DeadlineExceeded_Fatals verifies the helper fails the
// test when goroutines do not complete in time.
func TestWaitForAsync_DeadlineExceeded_Fatals(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	wg.Add(1)
	// Goroutine that never calls Done — exercised by the deadline path.
	stop := make(chan struct{})
	go func() {
		<-stop
		wg.Done()
	}()
	defer close(stop)

	fake := &fakeTB{}
	WaitForAsync(fake, &wg, 50*time.Millisecond)
	if !fake.fatalCalled.Load() {
		t.Errorf("WaitForAsync did not call tb.Fatalf on deadline exceeded")
	}
}

// TestWaitForAsync_ZeroWaitGroup_Immediate verifies the helper returns
// immediately when no goroutines were registered.
func TestWaitForAsync_ZeroWaitGroup_Immediate(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	start := time.Now()
	WaitForAsync(t, &wg, 1*time.Second)
	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Errorf("WaitForAsync on zero wg took %v, expected near-immediate", elapsed)
	}
}

// fakeTB is a minimal testing.TB implementation that records whether
// Fatalf was called. It satisfies the testing.TB interface so we can
// verify defensive paths in WaitForAsync without crashing the test binary.
type fakeTB struct {
	testing.TB
	fatalCalled atomic.Bool
}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Fatalf(format string, args ...any) {
	f.fatalCalled.Store(true)
	// Do NOT propagate — keep the wrapping test passing.
}
