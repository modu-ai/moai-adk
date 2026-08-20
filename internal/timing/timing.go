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
//   - Price the reference in the SAME WINDOW as the measured operation, not
//     in a burst before it. Load is not constant across a multi-second
//     measured loop, so a reference median taken once at t=0 stops
//     representing the load the measured samples actually saw (measured
//     2026-08-20, GitHub windows-latest: worst=305ms against a 24ms median —
//     a spike inside the measured window that the pre-priced reference never
//     saw, inflating only the numerator). AssertPaired interleaves the two
//     and is the preferred entry point; Assert's precomputed refUnit is the
//     legacy form.
//   - Give both sides the SAME sample count. A 10-sample reference median
//     under a 100-sample measured median makes the denominator the noisiest
//     term in the ratio.
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
	"os"
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
	report(t, b, refUnit, measure(fn, b.Iterations, b.Warmup))
}

// AssertPaired prices the reference and the measured operation TOGETHER —
// alternating one reference run with one measured run for b.Iterations rounds
// — and then enforces the same three bounds as Assert. Prefer it over Assert
// wherever the reference can be expressed as a func: a precomputed refUnit
// forces the reference to be priced in a burst BEFORE the measured loop, which
// breaks the premise the calibrated arm rests on (both sides must see the same
// load).
//
// Three differences from Assert, each closing one way the ratio moved on
// healthy code:
//
//   - Same window: every measured sample is bracketed by a reference sample,
//     so a load excursion inside the measured loop lands on both sides of the
//     ratio instead of only the numerator.
//   - Equal N: both medians rest on b.Iterations samples, so the denominator
//     is no noisier than the numerator.
//   - Alternating order: odd rounds run fn first, even rounds run ref first,
//     so neither side systematically inherits the other's warm caches.
//
// b.Warmup rounds of BOTH functions run first and are discarded.
func AssertPaired(t *testing.T, b Bound, ref, fn func()) {
	refSt, st := measurePaired(ref, fn, b.Iterations, b.Warmup)
	report(t, b, refSt.Median, st)
}

// report logs the measured distribution beside the reference unit and raises
// one t.Errorf per violated bound.
func report(t *testing.T, b Bound, refUnit time.Duration, st Stats) {
	ratio := math.Inf(1)
	if refUnit > 0 {
		ratio = float64(st.Median) / float64(refUnit)
	}
	line := fmt.Sprintf("%s: n=%d median=%v p95=%v worst=%v avg=%v | refUnit=%v ratio=%.2fx (maxUnits=%.2fx, steadyCeiling=%v, budget=%v)",
		b.Name, st.N, st.Median.Round(time.Microsecond), st.P95.Round(time.Microsecond),
		st.Worst.Round(time.Microsecond), st.Avg.Round(time.Microsecond),
		refUnit.Round(time.Microsecond), ratio, b.MaxUnits, b.SteadyCeiling, b.Budget)
	t.Log(line)
	publish(line)

	for _, err := range Check(b, refUnit, st) {
		t.Errorf("%s: %v", b.Name, err)
	}
}

// publish appends line to the GitHub Actions job summary when running there,
// and does nothing otherwise.
//
// t.Log alone is not enough to keep the ratio observable: `go test` discards a
// PASSING package's output unless -v is set, and the CI suite does not run
// verbose (it would flood the log). So the healthy ratio — the only baseline
// that tells a genuine regression apart from a bound set too tight — was
// visible ONLY on the run that failed, leaving one sample per platform in
// existence and no way to judge whether it was high. Appending one line to the
// job summary records the healthy figure on green runs too.
//
// Failures are ignored on purpose: an unwritable summary file must never turn
// a passing latency assertion red.
func publish(line string) {
	path := os.Getenv("GITHUB_STEP_SUMMARY")
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = fmt.Fprintf(f, "`%s`\n\n", line)
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
	for range iterations {
		samples = append(samples, timeOne(fn))
	}
	return summarize(samples)
}

// measurePaired runs ref and fn alternately for iterations rounds (after
// warmup rounds of both, discarded) and summarizes each side separately. Odd
// rounds run fn first so neither side always follows the other.
func measurePaired(ref, fn func(), iterations, warmup int) (refSt, st Stats) {
	for range warmup {
		ref()
		fn()
	}
	refSamples := make([]time.Duration, 0, iterations)
	samples := make([]time.Duration, 0, iterations)
	for i := range iterations {
		if i%2 == 0 {
			refSamples = append(refSamples, timeOne(ref))
			samples = append(samples, timeOne(fn))
			continue
		}
		samples = append(samples, timeOne(fn))
		refSamples = append(refSamples, timeOne(ref))
	}
	return summarize(refSamples), summarize(samples)
}

// timeOne runs fn once on the calling goroutine and returns its duration.
func timeOne(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

// summarize reduces per-invocation durations to the reported distribution.
// Returns the zero Stats on empty input.
func summarize(samples []time.Duration) Stats {
	if len(samples) == 0 {
		return Stats{}
	}
	var total time.Duration
	for _, s := range samples {
		total += s
	}
	return Stats{
		N:      len(samples),
		Median: median(samples),
		P95:    percentile(samples, 0.95),
		Worst:  slices.Max(samples),
		Avg:    total / time.Duration(len(samples)),
	}
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
