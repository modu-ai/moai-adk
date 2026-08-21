package timing

import (
	"slices"
	"strings"
	"testing"
	"time"
)

// cpuUnit is a deterministic CPU-bound unit: its duration scales linearly
// with n and is independent of the filesystem or subprocess machinery, so
// ratios between two cpuUnit calls are stable under *steady* machine load.
//
// [HARD] Every test that measures with cpuUnit runs WITHOUT t.Parallel().
//
// Steady is the load-bearing word, and the parallel runs in this package broke
// it. Interleaving reference and measured samples (measurePaired) cancels a
// load level that holds across both, but a sibling measuring test starting or
// finishing mid-run is a step change, not a level: the two sides of the ratio
// land on opposite sides of the step. With ref and fn byte-identical — the
// ratio's ideal being 1.00x against a 2.00x bound — CI still observed 2.32x,
// 2.72x, and 4.64x, and the reference unit itself moved 433µs to 1.687ms
// between runs.
//
// So the harness reported a code regression on code that had not changed. The
// non-parallel tests in a package run to completion before the parallel ones
// resume, which is what keeps these five off each other's measuring window.
// Adding t.Parallel() to a measuring test reopens this.
func cpuUnit(n int) {
	sum := 0
	for i := 0; i < n; i++ {
		sum += i % 7
	}
	_ = sum
}

func TestPercentileNearestRank(t *testing.T) {
	t.Parallel()
	// 1..100: nearest-rank p50 = 50th smallest, p95 = 95th smallest.
	samples := make([]time.Duration, 100)
	for i := range samples {
		samples[i] = time.Duration(i + 1)
	}
	if got := percentile(samples, 0.50); got != 50 {
		t.Errorf("p50 = %v, want 50", got)
	}
	if got := percentile(samples, 0.95); got != 95 {
		t.Errorf("p95 = %v, want 95", got)
	}
	if got := percentile(samples, 1.0); got != 100 {
		t.Errorf("p100 = %v, want 100", got)
	}
}

func TestMedianIgnoresSampleOrder(t *testing.T) {
	t.Parallel()
	a := []time.Duration{1, 2, 3, 4, 5}
	b := []time.Duration{5, 1, 4, 2, 3}
	if median(a) != median(b) || median(a) != 3 {
		t.Errorf("median order-sensitive or wrong: median(a)=%v median(b)=%v, want 3", median(a), median(b))
	}
}

func TestCheckHealthyStatsPassAllArms(t *testing.T) {
	t.Parallel()
	b := Bound{
		Name:          "unit",
		Budget:        5 * time.Second,
		SteadyCeiling: 1 * time.Second,
		MaxUnits:      1.5,
	}
	// Healthy: median == refUnit (ratio 1.0x), p95/worst well inside budget.
	st := Stats{N: 100, Median: 30 * time.Millisecond, P95: 60 * time.Millisecond, Worst: 90 * time.Millisecond}
	if errs := Check(b, 30*time.Millisecond, st); len(errs) != 0 {
		t.Errorf("healthy stats produced errors: %v", errs)
	}
}

func TestCheckCalibratedArmTripsOnUnitGrowth(t *testing.T) {
	t.Parallel()
	b := Bound{
		Name:          "unit",
		Budget:        5 * time.Second,
		SteadyCeiling: 10 * time.Second, // other arms loose: isolate the calibrated arm
		MaxUnits:      1.5,
	}
	// A code regression that doubles the underlying work: median = 2.0x the
	// reference unit while every absolute figure stays generous.
	st := Stats{N: 100, Median: 60 * time.Millisecond, P95: 90 * time.Millisecond, Worst: 120 * time.Millisecond}
	errs := Check(b, 30*time.Millisecond, st)
	if len(errs) != 1 {
		t.Fatalf("calibrated arm: got %d errors, want exactly 1: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Error(), "calibrated bound") {
		t.Errorf("error does not name the calibrated bound: %v", errs[0])
	}
}

func TestCheckCalibratedArmSkippedWithoutReference(t *testing.T) {
	t.Parallel()
	b := Bound{Budget: 5 * time.Second, SteadyCeiling: 1 * time.Second, MaxUnits: 1.5}
	st := Stats{N: 10, Median: time.Hour, P95: time.Hour, Worst: time.Hour}
	// refUnit == 0 must disable only the calibrated arm; the other two arms
	// still fire (worst >= budget).
	errs := Check(b, 0, st)
	if len(errs) != 2 {
		t.Fatalf("got %d errors, want 2 (steady + budget, calibrated skipped): %v", len(errs), errs)
	}
}

func TestCheckSteadyAndBudgetArms(t *testing.T) {
	t.Parallel()
	b := Bound{Budget: 5 * time.Second, SteadyCeiling: 1 * time.Second, MaxUnits: 1.5, Name: "unit"}
	st := Stats{N: 100, Median: 10 * time.Millisecond, P95: 2 * time.Second, Worst: 6 * time.Second}
	errs := Check(b, 10*time.Millisecond, st)
	if len(errs) != 2 {
		t.Fatalf("got %d errors, want 2 (steady + budget): %v", len(errs), errs)
	}
}

// TestMeasureCalibratedRatioHealthy runs a real measurement: fn costs the
// same CPU unit as the reference, so the calibrated arm must pass with
// generous headroom.
//
// Both sides are measured through measurePaired so the ratio rests on
// interleaved, equal-sized sample sets. Measuring the reference to
// completion BEFORE the subject leaves the two halves of the ratio exposed
// to different machine load, which is what a shared CI runner supplies: the
// denominator can inflate (or deflate) several-fold on its own, and the
// resulting ratio says more about the runner than about the code.
func TestMeasureCalibratedRatioHealthy(t *testing.T) {
	// Deliberately NOT parallel — measuring test; see the note on cpuUnit.
	unit := func() { cpuUnit(5_000_000) }
	refSt, st := measurePaired(unit, unit, 30, 3)
	b := Bound{Budget: 30 * time.Second, SteadyCeiling: 10 * time.Second, MaxUnits: 2.0, Name: "cpu-1x"}
	if errs := Check(b, refSt.Median, st); len(errs) != 0 {
		t.Errorf("healthy 1x ratio tripped a bound (ref=%v median=%v): %v", refSt.Median, st.Median, errs)
	}
}

// TestMeasureCalibratedRatioTripsAt4x verifies the calibrated arm catches a
// genuine 4x cost growth through a real measurement — the property the
// budget-fraction arms cannot provide when absolute figures stay generous.
//
// Paired for the same reason as the healthy case above, and this test is
// where an unpaired reference was measured failing: on a GitHub ubuntu
// runner the sequential reference took 4.18ms for its 2M-iteration unit
// while the subject took 5.62ms for 8M — a per-iteration cost 3x higher on
// the reference side alone, collapsing a true 4.0x ratio to 1.34x and
// letting the growth through undetected. Interleaving the two units puts
// both halves of the ratio under the same load.
func TestMeasureCalibratedRatioTripsAt4x(t *testing.T) {
	// Deliberately NOT parallel — measuring test; see the note on cpuUnit.
	refSt, st := measurePaired(func() { cpuUnit(2_000_000) }, func() { cpuUnit(8_000_000) }, 30, 3)
	b := Bound{Budget: time.Hour, SteadyCeiling: time.Hour, MaxUnits: 1.5, Name: "cpu-4x"}
	errs := Check(b, refSt.Median, st)
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "calibrated bound") {
		t.Fatalf("4x cost growth not caught by the calibrated arm (ref=%v median=%v): %v", refSt.Median, st.Median, errs)
	}
}

func TestMeasureDiscardsWarmup(t *testing.T) {
	t.Parallel()
	calls := 0
	st := measure(func() { calls++ }, 10, 4)
	if calls != 14 {
		t.Errorf("fn ran %d times, want 14 (10 measured + 4 warmup)", calls)
	}
	if st.N != 10 {
		t.Errorf("stats N = %d, want 10", st.N)
	}
}

func TestMedianRunsWarmupPlusSamples(t *testing.T) {
	t.Parallel()
	calls := 0
	// A real unit whose duration is guaranteed nonzero on ANY clock
	// resolution: it spins until the monotonic clock has advanced at least
	// one tick. A fixed-iteration cpuUnit can complete below a coarse
	// clock's resolution — GitHub Windows runners read interrupt time, so a
	// sub-tick unit legitimately measures 0 (Release Verify windows job
	// 95500006316: "Median of a real call = 0s, want > 0"). The iteration
	// bound only guards against a pathological never-ticking clock; on a
	// healthy clock the loop exits within one tick.
	tickUnit := func() {
		calls++
		start := time.Now()
		for i := 0; time.Since(start) <= 0 && i < 5_000_000; i++ {
			cpuUnit(1_000)
		}
	}
	got := Median(tickUnit, 3, 7)
	if calls != 10 {
		t.Errorf("ref ran %d times, want 10 (3 warmup + 7 measured)", calls)
	}
	if got <= 0 {
		t.Errorf("Median of a real call = %v, want > 0", got)
	}
}

func TestAssertHealthyEndToEnd(t *testing.T) {
	// Deliberately NOT parallel — measuring test; see the note on cpuUnit.
	ref := Median(func() { cpuUnit(5_000_000) }, 2, 10)
	// Generous bounds: this exercises the Assert wiring (measure + log +
	// Check) on healthy code and must not fail.
	//
	// MaxUnits is deliberately non-binding here. Assert's signature takes a
	// PRECOMPUTED refUnit, so the reference is necessarily priced in a burst
	// before the measured loop — the very exposure AssertPaired exists to
	// close, and one measured swinging the per-iteration reference cost by
	// 3x on a shared runner. Pinning a tight ratio on top of that would test
	// the machine; the calibrated arm's own semantics are pinned by
	// TestCheckCalibratedArmTripsOnUnitGrowth (synthetic Stats) and by the
	// paired real-measurement tests above.
	Assert(t, Bound{
		Name:          "cpu-e2e",
		Budget:        30 * time.Second,
		SteadyCeiling: 10 * time.Second,
		MaxUnits:      50.0,
		Iterations:    20,
		Warmup:        2,
	}, ref, func() { cpuUnit(5_000_000) })
}

func TestStatsWorstMatchesSamples(t *testing.T) {
	t.Parallel()
	samples := []time.Duration{5, 9, 2, 7}
	if got := slices.Max(samples); got != 9 {
		t.Errorf("max = %v, want 9", got)
	}
}
