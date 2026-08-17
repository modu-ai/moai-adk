package timing

import (
	"slices"
	"strings"
	"testing"
	"time"
)

// cpuUnit is a deterministic CPU-bound unit: its duration scales linearly
// with n and is independent of the filesystem or subprocess machinery, so
// ratios between two cpuUnit calls are stable under machine load.
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
func TestMeasureCalibratedRatioHealthy(t *testing.T) {
	t.Parallel()
	ref := Median(func() { cpuUnit(5_000_000) }, 2, 10)
	st := measure(func() { cpuUnit(5_000_000) }, 30, 3)
	b := Bound{Budget: 30 * time.Second, SteadyCeiling: 10 * time.Second, MaxUnits: 2.0, Name: "cpu-1x"}
	if errs := Check(b, ref, st); len(errs) != 0 {
		t.Errorf("healthy 1x ratio tripped a bound (ref=%v median=%v): %v", ref, st.Median, errs)
	}
}

// TestMeasureCalibratedRatioTripsAt4x verifies the calibrated arm catches a
// genuine 4x cost growth through a real measurement — the property the
// budget-fraction arms cannot provide when absolute figures stay generous.
func TestMeasureCalibratedRatioTripsAt4x(t *testing.T) {
	t.Parallel()
	ref := Median(func() { cpuUnit(2_000_000) }, 2, 10)
	st := measure(func() { cpuUnit(8_000_000) }, 30, 3)
	b := Bound{Budget: time.Hour, SteadyCeiling: time.Hour, MaxUnits: 1.5, Name: "cpu-4x"}
	errs := Check(b, ref, st)
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "calibrated bound") {
		t.Fatalf("4x cost growth not caught by the calibrated arm (ref=%v median=%v): %v", ref, st.Median, errs)
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
	// A real (if small) unit: an empty closure can complete below the
	// monotonic clock's resolution and legitimately measure as 0.
	got := Median(func() { calls++; cpuUnit(200_000) }, 3, 7)
	if calls != 10 {
		t.Errorf("ref ran %d times, want 10 (3 warmup + 7 measured)", calls)
	}
	if got <= 0 {
		t.Errorf("Median of a real call = %v, want > 0", got)
	}
}

func TestAssertHealthyEndToEnd(t *testing.T) {
	t.Parallel()
	ref := Median(func() { cpuUnit(5_000_000) }, 2, 10)
	// Generous bounds: this exercises the Assert wiring (measure + log +
	// Check) on healthy code and must not fail.
	Assert(t, Bound{
		Name:          "cpu-e2e",
		Budget:        30 * time.Second,
		SteadyCeiling: 10 * time.Second,
		MaxUnits:      2.0,
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
