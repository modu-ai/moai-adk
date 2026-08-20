package timing

import (
	"testing"
	"time"
)

// TestMeasurePairedEqualSampleCounts proves the property Assert's precomputed
// refUnit cannot offer: both sides of the ratio rest on the SAME number of
// samples, so the denominator is no noisier than the numerator.
func TestMeasurePairedEqualSampleCounts(t *testing.T) {
	t.Parallel()
	refSt, st := measurePaired(func() { cpuUnit(200_000) }, func() { cpuUnit(200_000) }, 20, 2)
	if refSt.N != 20 || st.N != 20 {
		t.Fatalf("sample counts = ref %d / measured %d; want 20 / 20", refSt.N, st.N)
	}
}

// TestMeasurePairedDiscardsWarmupOnBothSides verifies warmup rounds run BOTH
// functions and are excluded from the reported samples.
func TestMeasurePairedDiscardsWarmupOnBothSides(t *testing.T) {
	t.Parallel()
	refCalls, fnCalls := 0, 0
	refSt, st := measurePaired(func() { refCalls++ }, func() { fnCalls++ }, 10, 4)
	if refCalls != 14 || fnCalls != 14 {
		t.Fatalf("invocations = ref %d / fn %d; want 14 / 14 (10 measured + 4 warmup)", refCalls, fnCalls)
	}
	if refSt.N != 10 || st.N != 10 {
		t.Fatalf("reported samples = ref %d / measured %d; want 10 / 10 (warmup discarded)", refSt.N, st.N)
	}
}

// TestMeasurePairedAlternatesOrder proves neither side systematically runs
// first: over an even round count each side leads exactly half the rounds, so
// neither always inherits the other's warm caches.
func TestMeasurePairedAlternatesOrder(t *testing.T) {
	t.Parallel()
	var order []string
	measurePaired(
		func() { order = append(order, "ref") },
		func() { order = append(order, "fn") },
		4, 0,
	)
	want := []string{"ref", "fn", "fn", "ref", "ref", "fn", "fn", "ref"}
	if len(order) != len(want) {
		t.Fatalf("call order length = %d, want %d (%v)", len(order), len(want), order)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("call order = %v, want %v", order, want)
		}
	}
}

// TestAssertPairedHealthyEndToEnd runs the exported paired entry point on an
// operation that costs one reference unit: every bound must pass.
func TestAssertPairedHealthyEndToEnd(t *testing.T) {
	t.Parallel()
	AssertPaired(t, Bound{
		Name:          "paired-cpu-1x",
		Budget:        30 * time.Second,
		SteadyCeiling: 10 * time.Second,
		MaxUnits:      2.0,
		Iterations:    20,
		Warmup:        2,
	}, func() { cpuUnit(2_000_000) }, func() { cpuUnit(2_000_000) })
}

// TestSummarizeEmptyIsZero pins the empty-input contract summarize introduced
// (measure/measurePaired never pass an empty slice, but the helper is shared).
func TestSummarizeEmptyIsZero(t *testing.T) {
	t.Parallel()
	if got := summarize(nil); got != (Stats{}) {
		t.Fatalf("summarize(nil) = %+v, want zero Stats", got)
	}
}
