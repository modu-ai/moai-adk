package timing

import (
	"strings"
	"testing"
	"time"
)

// TestCalibratedRatioSurvivesOffsetLoadStep pins the property that made the
// calibrated arm report a code regression against code that had not changed
// (card t222): a load STEP that the reference series and the measured series
// cross ONE ROUND APART.
//
// The alternating ref-first/fn-first order produces exactly that offset at the
// crossing round. When the crossing sits at the median, the two order
// statistics land on opposite sides of the step and median(fn)/median(ref)
// reports the step's magnitude as if it were code cost — 1.89x here, against
// series whose true per-round ratio is 1.00x at every single round.
//
// This is not a hypothetical: CI observed 1.82x on TestBranchGuard_Latency
// (run 32687843472 attempt 1, 2026-08-24) and 2.32x / 2.72x / 4.64x on this
// package's own byte-identical self-test before #1591. The per-round paired
// ratio cannot express the artifact — both of its terms are measured
// microseconds apart, so the step cancels inside each term.
func TestCalibratedRatioSurvivesOffsetLoadStep(t *testing.T) {
	t.Parallel()

	const (
		n    = 100
		low  = 1900 * time.Microsecond
		high = 3600 * time.Microsecond // ~1.9x load level after the step
	)

	// crossAt is the round at which fn crosses the step; ref crosses one round
	// later. Sweeping across the median exercises the knife edge.
	for _, crossAt := range []int{40, 48, 49, 50, 51, 60} {
		fnS := make([]time.Duration, 0, n)
		ratios := make([]float64, 0, n)
		for i := range n {
			r, f := low, low
			if i >= crossAt {
				f = high
			}
			if i >= crossAt+1 {
				r = high
			}
			fnS = append(fnS, f)
			ratios = append(ratios, float64(f)/float64(r))
		}

		st := summarize(fnS)
		b := Bound{Name: "offset-step", Budget: time.Hour, SteadyCeiling: time.Hour, MaxUnits: 1.5}

		// The paired estimator — what AssertPaired now enforces — must not
		// trip: every round's true ratio is 1.00x.
		if errs := CheckRatio(b, medianFloat(ratios), st); len(errs) != 0 {
			t.Errorf("crossAt=%d: paired ratio tripped a bound on a pure load step: %v", crossAt, errs)
		}
	}

	// And the discarded-pairing estimator does trip at the knife edge — the
	// falsifier: without this arm the test above would pass for any estimator,
	// including one that ignores the data entirely.
	const crossAt = 49
	refS := make([]time.Duration, 0, n)
	fnS := make([]time.Duration, 0, n)
	for i := range n {
		r, f := low, low
		if i >= crossAt {
			f = high
		}
		if i >= crossAt+1 {
			r = high
		}
		refS = append(refS, r)
		fnS = append(fnS, f)
	}
	b := Bound{Name: "offset-step-unpaired", Budget: time.Hour, SteadyCeiling: time.Hour, MaxUnits: 1.5}
	errs := Check(b, summarize(refS).Median, summarize(fnS))
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "calibrated bound") {
		t.Fatalf("ratio-of-medians did NOT trip at the knife edge — the artifact this test pins is gone or mis-simulated: %v", errs)
	}
}

// TestPairedRatioStillCatchesRealCostGrowth is the held-out arm: making the
// estimator step-robust must not make it blind. A measured side that costs 4x
// the reference on EVERY round still trips the bound.
func TestPairedRatioStillCatchesRealCostGrowth(t *testing.T) {
	t.Parallel()

	ratios := make([]float64, 0, 50)
	samples := make([]time.Duration, 0, 50)
	for i := range 50 {
		// Load varies wildly round to round; the per-round ratio stays 4x.
		load := time.Duration(1+i%7) * time.Millisecond
		ratios = append(ratios, 4.0)
		samples = append(samples, 4*load)
	}
	b := Bound{Name: "cpu-4x", Budget: time.Hour, SteadyCeiling: time.Hour, MaxUnits: 1.5}
	errs := CheckRatio(b, medianFloat(ratios), summarize(samples))
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "calibrated bound") {
		t.Fatalf("4x per-round cost growth not caught by the paired arm: %v", errs)
	}
}
