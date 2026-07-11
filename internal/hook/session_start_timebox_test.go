package hook

import (
	"context"
	"strings"
	"testing"
	"time"
)

// Session-start drift time-box tests (SPEC-SESSIONSTART-PERF-001 M3 —
// REQ-SSP-015 / AC-SSP-015).
//
// detectStatusDrift runs an ADVISORY check on the session-start critical path.
// It MUST be time-boxed: on deadline exceed it skips the (abandoned) computation
// and emits the advisory string instead of blocking session start unboundedly.
//
// These tests mutate the package-level driftCountFn / sessionStartDriftTimeout
// seams, so they MUST NOT run in parallel.

// TestDetectStatusDrift_TimeBoxEmitsAdvisory injects a slow drift computation and
// a short deadline, and proves detectStatusDrift returns the advisory PROMPTLY
// (does not block) once the time-box fires.
func TestDetectStatusDrift_TimeBoxEmitsAdvisory(t *testing.T) {
	origFn := driftCountFn
	origTimeout := sessionStartDriftTimeout
	t.Cleanup(func() {
		driftCountFn = origFn
		sessionStartDriftTimeout = origTimeout
	})

	sessionStartDriftTimeout = 50 * time.Millisecond
	// A drift fn that honors the deadline (like the real DriftCountCtx would):
	// it returns ctx.Err() once the box fires. Blocks for 10s otherwise, which
	// would hang the test if the time-box were absent.
	driftCountFn = func(ctx context.Context, _ string) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(10 * time.Second):
			return 99, nil
		}
	}

	start := time.Now()
	msg := detectStatusDrift("/some/project")
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Fatalf("detectStatusDrift blocked %v, want a prompt time-boxed return (non-blocking)", elapsed)
	}
	if !strings.Contains(msg, "moai spec drift") {
		t.Errorf("time-box advisory = %q, want it to preserve the 'moai spec drift' advisory", msg)
	}
}

// TestDetectStatusDrift_CountBranch confirms the ordinary path is unchanged: a
// drift count at or above the warning threshold surfaces the count advisory.
func TestDetectStatusDrift_CountBranch(t *testing.T) {
	origFn := driftCountFn
	t.Cleanup(func() { driftCountFn = origFn })

	driftCountFn = func(context.Context, string) (int, error) {
		return driftWarningThreshold, nil
	}

	msg := detectStatusDrift("/some/project")
	if !strings.Contains(msg, "status drift") || !strings.Contains(msg, "moai spec drift") {
		t.Errorf("count-branch advisory = %q, want the drift-count advisory", msg)
	}
}

// TestDetectStatusDrift_BelowThresholdSilent confirms a sub-threshold count emits
// nothing (the check is best-effort and quiet below the threshold).
func TestDetectStatusDrift_BelowThresholdSilent(t *testing.T) {
	origFn := driftCountFn
	t.Cleanup(func() { driftCountFn = origFn })

	driftCountFn = func(context.Context, string) (int, error) {
		return driftWarningThreshold - 1, nil
	}

	if msg := detectStatusDrift("/some/project"); msg != "" {
		t.Errorf("below-threshold advisory = %q, want empty", msg)
	}
}

// TestDetectStatusDrift_NonTimeoutErrorSilent confirms a genuine (non-timeout)
// error — git absent, no specs directory — is silently ignored, preserving the
// pre-existing best-effort, non-blocking behavior.
func TestDetectStatusDrift_NonTimeoutErrorSilent(t *testing.T) {
	origFn := driftCountFn
	t.Cleanup(func() { driftCountFn = origFn })

	driftCountFn = func(context.Context, string) (int, error) {
		return 0, context.Canceled // any non-DeadlineExceeded error stands in for "git absent"
	}

	if msg := detectStatusDrift("/some/project"); msg != "" {
		t.Errorf("non-timeout-error advisory = %q, want empty (silent)", msg)
	}
}
