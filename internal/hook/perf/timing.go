// Package perf provides a lightweight, env-gated timing collector for hook
// invocations, plus the concurrent-stress profiling harness used by
// SPEC-HOOK-PRETOOL-PERF-001 M0 (baseline) and M3 (post-change).
//
// When the environment variable MOAI_HOOK_PERF_TIMING is set to a non-empty
// value, the TimingCollector records per-phase wall-time and emits a JSON line
// to stderr at the end of the hook invocation. When the variable is unset, every
// method is a no-op, so the hot path incurs zero overhead in production.
//
// The emitted JSON line has the shape:
//
//	{"phase":"perf_timing","fork_exec_ms":1.2,"config_load_ms":15.3,"dispatch_ms":2.1,"total_ms":18.6}
//
// The profiling harness (harness_test.go) parses these lines to aggregate p50 /
// p99 / max-tail across batches of parallel subprocess invocations.
package perf

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// envPerfTiming is the environment variable that enables per-phase timing.
const envPerfTiming = "MOAI_HOOK_PERF_TIMING"

// Enabled reports whether per-phase timing collection is active.
func Enabled() bool {
	return os.Getenv(envPerfTiming) != ""
}

// TimingSnapshot captures wall-time for each phase of a single hook invocation.
// All durations are in milliseconds. The JSON tag names are short to keep the
// stderr line compact under high concurrency.
type TimingSnapshot struct {
	ForkExecMs   float64 `json:"fork_exec_ms"`
	ConfigLoadMs float64 `json:"config_load_ms"`
	DispatchMs   float64 `json:"dispatch_ms"`
	TotalMs      float64 `json:"total_ms"`
}

// TimingCollector records per-phase timestamps within a single hook invocation.
// It is safe for use by a single goroutine (the hook process itself is
// single-threaded for the phases measured). The zero value is a no-op
// collector; use NewTimingCollector to get an active one.
type TimingCollector struct {
	enabled  bool
	procStart time.Time // process start (captured by caller before InitDependencies)
	mu       sync.Mutex
	snapshot TimingSnapshot
}

// NewTimingCollector returns a collector that is active only when
// MOAI_HOOK_PERF_TIMING is set. The procStart parameter should be captured as
// early as possible in the process lifecycle (e.g. at the top of main or the
// earliest available init point).
func NewTimingCollector(procStart time.Time) *TimingCollector {
	return &TimingCollector{
		enabled:  Enabled(),
		procStart: procStart,
	}
}

// MarkForkExec records the fork+exec phase: the wall-time from procStart to the
// moment InitDependencies is about to be entered. This captures process
// creation + Go runtime init + cobra root setup.
func (tc *TimingCollector) MarkForkExec(at time.Time) {
	if tc == nil || !tc.enabled {
		return
	}
	tc.mu.Lock()
	tc.snapshot.ForkExecMs = float64(at.Sub(tc.procStart).Microseconds()) / 1000.0
	tc.mu.Unlock()
}

// MarkConfigLoad records the config-load phase duration. The caller records
// `start` before calling ConfigManager.Load and `end` after it returns.
func (tc *TimingCollector) MarkConfigLoad(start, end time.Time) {
	if tc == nil || !tc.enabled {
		return
	}
	tc.mu.Lock()
	tc.snapshot.ConfigLoadMs = float64(end.Sub(start).Microseconds()) / 1000.0
	tc.mu.Unlock()
}

// MarkDispatch records the handler-dispatch phase duration.
func (tc *TimingCollector) MarkDispatch(start, end time.Time) {
	if tc == nil || !tc.enabled {
		return
	}
	tc.mu.Lock()
	tc.snapshot.DispatchMs = float64(end.Sub(start).Microseconds()) / 1000.0
	tc.mu.Unlock()
}

// Emit writes the timing snapshot as a JSON line to stderr if timing is enabled.
// totalEnd should be the wall-clock time at process exit (or as close to it as
// the caller can capture).
func (tc *TimingCollector) Emit(totalEnd time.Time) {
	if tc == nil || !tc.enabled {
		return
	}
	tc.mu.Lock()
	tc.snapshot.TotalMs = float64(totalEnd.Sub(tc.procStart).Microseconds()) / 1000.0
	snap := tc.snapshot
	tc.mu.Unlock()

	line := map[string]any{
		"phase":           "perf_timing",
		"fork_exec_ms":    snap.ForkExecMs,
		"config_load_ms":  snap.ConfigLoadMs,
		"dispatch_ms":     snap.DispatchMs,
		"total_ms":        snap.TotalMs,
	}
	data, err := json.Marshal(line)
	if err != nil {
		return
	}
	fmt.Fprintln(os.Stderr, string(data))
}
