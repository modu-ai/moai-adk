package spec

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Time-box tests for the context-aware drift entry point
// (SPEC-SESSIONSTART-PERF-001 M3 — REQ-SSP-015 / AC-SSP-015).
//
// These drive DriftCountCtx through the driftWorkFn seam so a deliberately-slow,
// context-IGNORING worker can prove the deadline is enforced by DriftCountCtx's
// own select — not by cooperation from the worker. That is the guarantee the
// session-start critical path relies on: the drift computation itself is a
// bounded synchronous git+in-memory pass with no ctx awareness, so the ONLY
// thing that can abandon a pathological run is the time-box wrapper.
//
// The tests mutate the package-level driftWorkFn seam, so they MUST NOT run in
// parallel.

// TestDriftCountCtx_TimeoutAbandonsSlowWorker proves the time-box abandons a
// worker that ignores cancellation entirely: DriftCountCtx returns
// context.DeadlineExceeded promptly (well before the worker would finish) rather
// than blocking on it.
func TestDriftCountCtx_TimeoutAbandonsSlowWorker(t *testing.T) {
	orig := driftWorkFn
	t.Cleanup(func() { driftWorkFn = orig })

	// A worker that IGNORES the deadline and just sleeps. If DriftCountCtx waited
	// on it, this test would block for 5s.
	driftWorkFn = func(string) (int, error) {
		time.Sleep(5 * time.Second)
		return 42, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	count, err := DriftCountCtx(ctx, t.TempDir())
	elapsed := time.Since(start)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("DriftCountCtx err = %v, want context.DeadlineExceeded", err)
	}
	if count != 0 {
		t.Errorf("DriftCountCtx count = %d, want 0 on timeout", count)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("DriftCountCtx blocked %v on a slow worker, want prompt abandon (< 2s)", elapsed)
	}
}

// TestDriftCountCtx_CompletesWithinDeadline is the happy path: a worker that
// returns before the deadline yields its result unchanged, with no error.
func TestDriftCountCtx_CompletesWithinDeadline(t *testing.T) {
	orig := driftWorkFn
	t.Cleanup(func() { driftWorkFn = orig })

	driftWorkFn = func(string) (int, error) { return 7, nil }

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	count, err := DriftCountCtx(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("DriftCountCtx err = %v, want nil", err)
	}
	if count != 7 {
		t.Errorf("DriftCountCtx count = %d, want 7", count)
	}
}

// TestDriftCountCtx_PropagatesWorkerError confirms a genuine worker error (git
// absent, unreadable specs dir) is surfaced unchanged — the time-box does not
// swallow real errors, only the timeout path returns ctx.Err().
func TestDriftCountCtx_PropagatesWorkerError(t *testing.T) {
	orig := driftWorkFn
	t.Cleanup(func() { driftWorkFn = orig })

	sentinel := errors.New("git unavailable")
	driftWorkFn = func(string) (int, error) { return 0, sentinel }

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := DriftCountCtx(ctx, t.TempDir())
	if !errors.Is(err, sentinel) {
		t.Fatalf("DriftCountCtx err = %v, want the worker's sentinel error", err)
	}
}
