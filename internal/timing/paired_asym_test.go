package timing

import (
	"strings"
	"testing"
	"time"
)

// TestCalibratedGateSurvivesAlternationLockedAsymmetry pins the post-#1591
// residual failure mode of the per-round estimator: noise PHASE-LOCKED to the
// ref-first/fn-first alternation inflates the per-round ratios of half the
// rounds while leaving both medians equal (CI run 32779472351 attempt 1,
// 2026-08-24: per-round median 2.47x against medians 1.09x on byte-identical
// ref/fn — the false positive SPEC-CI-FLAKE-SERIES-001 exists to close).
//
// The synthetic distribution reproduces the observed shape with the true
// aggregate cost ratio 1.00x: even rounds favor fn (ratio 2.50x), odd rounds
// favor ref (ratio 0.40x). The per-round median reads 2.50x — past the bound —
// while both order statistics sit at the same level, so the ratio-of-medians
// reads 1.00x.
func TestCalibratedGateSurvivesAlternationLockedAsymmetry(t *testing.T) {
	t.Parallel()

	const (
		n        = 20
		unit     = 400 * time.Microsecond
		favored  = 1000 * time.Microsecond // 2.5x the unit
		maxUnits = 2.0
	)
	fnS := make([]time.Duration, 0, n)
	refS := make([]time.Duration, 0, n)
	ratios := make([]float64, 0, n)
	for i := range n {
		fn, ref := unit, unit
		if i%2 == 0 {
			fn = favored // fn-favored round
		} else {
			ref = favored // ref-favored round
		}
		fnS = append(fnS, fn)
		refS = append(refS, ref)
		ratios = append(ratios, float64(fn)/float64(ref))
	}

	perRound := medianFloat(ratios)
	medians := ratioOfMedians(summarize(refS).Median, summarize(fnS))
	st := summarize(fnS)
	b := Bound{Name: "alternation-asymmetry", Budget: time.Hour, SteadyCeiling: time.Hour, MaxUnits: maxUnits}

	// Fixture sanity: the distribution actually reproduces the observed form
	// — per-round inflated past the bound, medians healthy. If this arm
	// fails, the fixture drifted, not the estimator.
	if perRound <= maxUnits {
		t.Fatalf("fixture drifted: per-round median %.2fx does not exceed the %.2fx bound (the observed form requires it past the bound)", perRound, maxUnits)
	}
	if medians > maxUnits {
		t.Fatalf("fixture drifted: ratio-of-medians %.2fx exceeds the %.2fx bound (the observed form requires healthy medians)", medians, maxUnits)
	}

	// The AND-gate — what AssertPaired now enforces — must not trip: the true
	// aggregate cost ratio is 1.00x.
	if errs := CheckRatioAnd(b, perRound, medians, st); len(errs) != 0 {
		t.Errorf("AND-gate tripped on alternation-locked asymmetry: %v", errs)
	}

	// Falsifier: the per-round-ALONE estimator (the pre-fix rule) DOES trip on
	// this distribution — without this arm the test above would pass for any
	// estimator, including one that ignores the data entirely.
	errs := CheckRatio(b, perRound, st)
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "calibrated bound") {
		t.Fatalf("per-round-alone estimator did NOT trip on alternation-locked asymmetry — the pinned defect is gone or mis-simulated: %v", errs)
	}
}

// TestPairedAndGateStillCatchesHomogeneousRegression is the held-out arm:
// opening the gate for alternation asymmetry must not blind it to a real
// regression. fn costing a uniform 3x the reference on EVERY round moves BOTH
// calibrated figures past the bound, and the gate still fires (REQ-CFS-004).
func TestPairedAndGateStillCatchesHomogeneousRegression(t *testing.T) {
	t.Parallel()

	const n = 20
	fnS := make([]time.Duration, 0, n)
	refS := make([]time.Duration, 0, n)
	ratios := make([]float64, 0, n)
	for i := range n {
		// Load varies round to round; the cost ratio stays a uniform 3x.
		load := time.Duration(1+i%5) * time.Millisecond
		refS = append(refS, load)
		fnS = append(fnS, 3*load)
		ratios = append(ratios, 3.0)
	}

	b := Bound{Name: "homogeneous-3x", Budget: time.Hour, SteadyCeiling: time.Hour, MaxUnits: 2.0}
	errs := CheckRatioAnd(b, medianFloat(ratios), ratioOfMedians(summarize(refS).Median, summarize(fnS)), summarize(fnS))
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "calibrated bound") {
		t.Fatalf("uniform 3x cost growth not caught by the AND-gate: %v", errs)
	}
}
