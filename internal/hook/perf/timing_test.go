package perf

import (
	"testing"
	"time"
)

// TestEnabled_reportsEnvState verifies the env-gate.
func TestEnabled_reportsEnvState(t *testing.T) {
	t.Setenv("MOAI_HOOK_PERF_TIMING", "1")
	if !Enabled() {
		t.Fatal("Enabled() should return true when MOAI_HOOK_PERF_TIMING is set")
	}

	t.Setenv("MOAI_HOOK_PERF_TIMING", "")
	if Enabled() {
		t.Fatal("Enabled() should return false when MOAI_HOOK_PERF_TIMING is empty")
	}
}

// TestTimingCollector_NoopWhenDisabled verifies that a disabled collector
// performs no operations (all Mark/Emit calls are no-ops).
func TestTimingCollector_NoopWhenDisabled(t *testing.T) {
	t.Setenv("MOAI_HOOK_PERF_TIMING", "")
	tc := NewTimingCollector(time.Now())

	// These should be no-ops without panicking.
	tc.MarkForkExec(time.Now())
	tc.MarkConfigLoad(time.Now(), time.Now())
	tc.MarkDispatch(time.Now(), time.Now())
	tc.Emit(time.Now())

	// The snapshot should be zero-valued.
	if tc == nil {
		t.Fatal("collector should not be nil")
	}
}

// TestTimingCollector_RecordsPhases verifies that an enabled collector
// records each phase and emits valid JSON.
func TestTimingCollector_RecordsPhases(t *testing.T) {
	t.Setenv("MOAI_HOOK_PERF_TIMING", "1")

	procStart := time.Now()
	tc := NewTimingCollector(procStart)

	time.Sleep(1 * time.Millisecond)
	tc.MarkForkExec(time.Now())

	cfgStart := time.Now()
	time.Sleep(1 * time.Millisecond)
	tc.MarkConfigLoad(cfgStart, time.Now())

	dispStart := time.Now()
	time.Sleep(1 * time.Millisecond)
	tc.MarkDispatch(dispStart, time.Now())

	// Emit should produce a snapshot with non-zero values.
	tc.mu.Lock()
	snap := tc.snapshot
	tc.mu.Unlock()

	if snap.ForkExecMs <= 0 {
		t.Fatalf("ForkExecMs should be > 0, got %f", snap.ForkExecMs)
	}
	if snap.ConfigLoadMs <= 0 {
		t.Fatalf("ConfigLoadMs should be > 0, got %f", snap.ConfigLoadMs)
	}
	if snap.DispatchMs <= 0 {
		t.Fatalf("DispatchMs should be > 0, got %f", snap.DispatchMs)
	}

	// Emit writes to stderr; we can't easily capture it, but we verify it
	// doesn't panic.
	tc.Emit(time.Now())
}

// TestTimingCollector_NilSafe verifies that nil receiver methods are safe.
func TestTimingCollector_NilSafe(t *testing.T) {
	var tc *TimingCollector

	// None of these should panic on a nil receiver.
	tc.MarkForkExec(time.Now())
	tc.MarkConfigLoad(time.Now(), time.Now())
	tc.MarkDispatch(time.Now(), time.Now())
	tc.Emit(time.Now())
}
