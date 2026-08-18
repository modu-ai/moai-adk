// Package timing provides calibrated latency bounds for tests: assertions
// that measure the CODE under test rather than the machine it runs on.
//
// An absolute wall-clock bound ("this must finish in 500ms") measures the
// machine as much as the code: process spawns and file writes track system
// load, so a loaded CI runner or developer box breaches the bound on healthy
// code (measured 2026-08-15: a spawn-bound hook check averaged 135-256ms and
// breached a 500ms ceiling 3 of 5 runs on unmodified code).
//
// The calibrated bound closes the gap the budget-fraction bounds leave open.
// It compares the measured operation against a REFERENCE operation of the
// same machine-cost class — one subprocess spawn, one file append — measured
// in the same process, at the same moment, under the same load. Their ratio
// is a property of the code (how many machine-cost units the operation
// consumes), so it survives machine differences, CI ephemeral runners, and
// full-suite load: load inflates both sides roughly equally and the ratio
// stays put. A code regression that doubles the underlying work (an added
// subprocess, an fsync per write) moves the ratio and trips the bound even
// when every absolute figure still sits inside a generous ceiling.
//
// That inflation-cancellation holds only while the reference and the
// measured operation carry the same cost MIX — the same weights of CPU work
// and syscalls — not merely the same scale. On VM CI runners a µs-scale
// mismatch decouples the two classes: the write-cycle reference stayed flat
// while a marshal+mkdir+append operation inflated to 2.56x-3.61x on healthy
// code (Release Verify ubuntu job 95500006280, 2026-08-17). A reference
// that mirrors the measured operation's healthy mix cancels class-specific
// inflation on both sides; an additive regression (a new fsync or spawn)
// still lands far above the bound.
//
// # Why not a persisted baseline
//
// A recorded-previous-run baseline (as opposed to this in-run reference)
// cannot honestly serve, for three reasons:
//
//   - The verify snapshot store keys every record to the exact tree state
//     (HEAD SHA + porcelain digest + `git diff HEAD` content hash — see
//     internal/verify/key.go Key). Any new commit mints a new key, so the
//     previous run's baseline is unreachable by construction.
//   - A baseline recorded at another TIME measures another LOAD state.
//     Comparing across time reintroduces exactly the machine-load
//     sensitivity this package exists to remove: an idle-machine baseline
//     makes a loaded healthy run look like a 2x regression.
//   - CI runners are ephemeral: the first-run fallback path would dominate
//     there anyway, so the calibrated arm would never fire where the tests
//     run most.
//
// The in-run reference is the only baseline measured under conditions
// identical to the measured operation. It is the honest form of
// "diff against a recorded baseline".
//
// # Reference rules
//
//   - The reference must be the SAME cost class as the measured operation
//     (the same syscalls, the same spawn, the same filesystem): a spawn-bound
//     operation is calibrated against one spawn, a write-bound operation
//     against one append. A mismatched cost class decouples the two under
//     load and the ratio stops being stable. For µs-scale operations the
//     reference must further mirror the measured operation's MIX of CPU work
//     and syscalls (timestamping, marshaling, stat, the write itself): VM
//     runners inflate the CPU class and the syscall class by different
//     factors, so a mix-mismatched reference moves the ratio on healthy
//     code (observed 2.56x-3.61x, ubuntu job 95500006280).
//   - Measure the reference immediately before the measured operation, in
//     the same process.
//   - Use the median of several reference runs; use the median of the
//     measured samples for the ratio (medians move together under load;
//     quantile mixes like p95-vs-median do not).
//   - Keep the reference at or above one clock tick: on coarse monotonic
//     clocks (GitHub Windows runners read interrupt time) a sub-tick
//     reference legitimately measures as 0, which silently disables the
//     calibrated arm. See TestMedianRunsWarmupPlusSamples for the
//     tick-guaranteed unit pattern.
package timing

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"testing"
	"time"
)

// Bound declares the three latency assertions Assert enforces.
type Bound struct {
	// Name labels the operation in log and failure output.
	Name string

	// Budget is the hard contract ceiling: no single invocation may consume
	// it (worst < Budget). Typically a hook's execution budget.
	Budget time.Duration

	// SteadyCeiling is the regression bound as a fraction of Budget
	// (p95 <= SteadyCeiling). Typically Budget/5 (20%).
	SteadyCeiling time.Duration

	// MaxUnits is the calibrated bound: median <= MaxUnits x refUnit, where
	// refUnit is the median cost of one reference operation of the same
	// machine-cost class, measured in the same run (see Median). 0 disables
	// the calibrated arm.
	//
	// Choose MaxUnits against the healthy unit count: an operation that
	// performs exactly one reference-class unit sits at ~1.0x (sub-ms
	// bookkeeping is invisible next to a spawn or a disk write), so a
	// threshold of 1.5x-2.0x has wide headroom over healthy code while
	// tripping on any change that roughly doubles the underlying work.
	MaxUnits float64

	// Iterations is the number of measured invocations of fn.
	Iterations int

	// Warmup is the number of leading invocations measured but discarded
	// (cold caches, first-touch filesystem effects).
	Warmup int
}

// Stats summarizes the measured per-invocation durations.
type Stats struct {
	N      int
	Median time.Duration
	P95    time.Duration
	Worst  time.Duration
	Avg    time.Duration
}

// Median runs ref warmup times (discarded), then n times, and returns the
// median per-run duration. Use it to price one reference operation — the
// same-cost-class unit the measured operation is calibrated against. A ref
// that is a no-op returns ~0, which disables the calibrated arm in Assert.
func Median(ref func(), warmup, n int) time.Duration {
	for range warmup {
		ref()
	}
	samples := make([]time.Duration, 0, n)
	for range n {
		start := time.Now()
		ref()
		samples = append(samples, time.Since(start))
	}
	return median(samples)
}

// Assert runs fn per b (after discarding b.Warmup runs) and enforces all
// three bounds, logging the measured distribution alongside the reference
// unit so the ratio is mechanically observable in test output:
//
//  1. p95 <= b.SteadyCeiling — a shifted distribution means the operation
//     itself got more expensive (the budget-fraction regression arm).
//  2. worst < b.Budget — no single invocation may consume the whole budget
//     (the contract arm).
//  3. median <= b.MaxUnits x refUnit — the calibrated arm. Measures code,
//     not machine: how many reference-class units one invocation costs.
//     Enforced only when b.MaxUnits > 0 and refUnit > 0.
//
// fn runs on the calling goroutine and may call t.Fatalf directly to fail
// on functional errors (wrong decision, write error) inside the loop.
func Assert(t *testing.T, b Bound, refUnit time.Duration, fn func()) {
	st := measure(fn, b.Iterations, b.Warmup)
	ratio := math.Inf(1)
	if refUnit > 0 {
		ratio = float64(st.Median) / float64(refUnit)
	}
	t.Logf("%s: n=%d median=%v p95=%v worst=%v avg=%v | refUnit=%v ratio=%.2fx (maxUnits=%.2fx, steadyCeiling=%v, budget=%v)",
		b.Name, st.N, st.Median.Round(time.Microsecond), st.P95.Round(time.Microsecond),
		st.Worst.Round(time.Microsecond), st.Avg.Round(time.Microsecond),
		refUnit.Round(time.Microsecond), ratio, b.MaxUnits, b.SteadyCeiling, b.Budget)

	for _, err := range Check(b, refUnit, st) {
		t.Errorf("%s: %v", b.Name, err)
	}
}

// Check evaluates the three bounds against already-measured stats and
// returns one error per violated bound. Pure: callers can unit-test the
// bound logic without running anything.
func Check(b Bound, refUnit time.Duration, st Stats) []error {
	var errs []error
	if st.P95 > b.SteadyCeiling {
		errs = append(errs, fmt.Errorf("p95 latency %v exceeds steady-state ceiling %v — "+
			"a whole-distribution slowdown: the operation itself got more expensive, not one unlucky sample",
			st.P95, b.SteadyCeiling))
	}
	if st.Worst >= b.Budget {
		errs = append(errs, fmt.Errorf("worst latency %v reaches the %v budget — "+
			"a single invocation consuming the whole budget stalls the caller",
			st.Worst, b.Budget))
	}
	if b.MaxUnits > 0 && refUnit > 0 {
		ratio := float64(st.Median) / float64(refUnit)
		if ratio > b.MaxUnits {
			errs = append(errs, fmt.Errorf("median latency %v is %.2fx the reference unit %v, above the %.2fx calibrated bound — "+
				"the operation now costs more machine-cost units than its design (e.g. an added subprocess or per-write fsync); "+
				"this is a code regression, not machine load (load inflates the reference equally)",
				st.Median, ratio, refUnit, b.MaxUnits))
		}
	}
	return errs
}

// measure runs fn warmup+iterations times on the calling goroutine,
// discards the warmup durations, and summarizes the rest.
func measure(fn func(), iterations, warmup int) Stats {
	for range warmup {
		fn()
	}
	samples := make([]time.Duration, 0, iterations)
	var total time.Duration
	for range iterations {
		start := time.Now()
		fn()
		elapsed := time.Since(start)
		samples = append(samples, elapsed)
		total += elapsed
	}
	st := Stats{
		N:      len(samples),
		Median: median(samples),
		P95:    percentile(samples, 0.95),
		Worst:  slices.Max(samples),
	}
	if len(samples) > 0 {
		st.Avg = total / time.Duration(len(samples))
	}
	return st
}

// median returns the 50th percentile of a non-empty sample set.
// Panics on empty input — callers always supply at least one sample.
func median(samples []time.Duration) time.Duration {
	return percentile(samples, 0.50)
}

// percentile returns the p-th percentile of a non-empty sample set using
// the nearest-rank method (the ceil(p*n)-th smallest value). Panics on
// empty input.
func percentile(samples []time.Duration, p float64) time.Duration {
	sorted := slices.Clone(samples)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(math.Ceil(p * float64(len(sorted))))
	if idx < 1 {
		idx = 1
	}
	return sorted[idx-1]
}
