package factorylane

import (
	"strings"
	"sync"
	"testing"
)

// classifiedFixture classifies t1 (the candidate) as sequential group alpha
// and t2 the same, mirroring what t1332's future metadata will say.
func classifiedFixture() StaticClassifier {
	return StaticClassifier{
		"t1": {Axis: AxisSequential, Group: "alpha", Known: true},
		"t2": {Axis: AxisSequential, Group: "alpha", Known: true},
		"t3": {Axis: AxisSequential, Group: "beta", Known: true},
		"tp": {Axis: AxisParallel, Known: true},
	}
}

// AC-FLA-006 (sequential, free group): a sequential-classified card with no
// other lane holding its group is pickable — exclusively (no multi-pick).
func TestPickupSequentialFreeGroupAllowedExclusive(t *testing.T) {
	d, err := PlanPickup("lane-1", "t1", nil, classifiedFixture())
	if err != nil {
		t.Fatalf("PlanPickup on a sequential card: %v", err)
	}
	if !d.Allowed || d.Fallback || !d.Classification.Known {
		t.Fatalf("decision = %+v, want allowed under the known sequential classification", d)
	}
	if d.MultiPick {
		t.Errorf("decision = %+v, want MultiPick=false for sequential work", d)
	}
}

// AC-FLA-006 (sequential, contended group): while another lane holds a card
// of the same sequential group, the candidate is not pickable — the waiting
// lane is named.
func TestPickupSequentialContendedGroupDenied(t *testing.T) {
	holds := []Hold{{Lane: "lane-2", Card: "t2"}}
	d, err := PlanPickup("lane-1", "t1", holds, classifiedFixture())
	if err != nil {
		t.Fatalf("PlanPickup on a contended sequential card: %v", err)
	}
	if d.Allowed {
		t.Fatalf("decision = %+v, want the sequential pickup denied while lane-2 holds group alpha", d)
	}
	if d.WaitOn != "lane-2" {
		t.Errorf("WaitOn = %q, want lane-2 (the holder)", d.WaitOn)
	}
}

// A same-axis hold in a DIFFERENT group never blocks: exclusivity is scoped
// to the sequential group, not to the axis as a whole.
func TestPickupSequentialOtherGroupHoldDoesNotBlock(t *testing.T) {
	holds := []Hold{{Lane: "lane-2", Card: "t3"}} // group beta
	d, err := PlanPickup("lane-1", "t1", holds, classifiedFixture())
	if err != nil {
		t.Fatalf("PlanPickup: %v", err)
	}
	if !d.Allowed {
		t.Fatalf("decision = %+v, want allowed — group beta is not group alpha", d)
	}
}

// AC-FLA-006 (parallel): a parallel-classified card stays pickable while
// another lane already holds it.
func TestPickupParallelHoldDoesNotBlock(t *testing.T) {
	holds := []Hold{{Lane: "lane-2", Card: "tp"}}
	d, err := PlanPickup("lane-1", "tp", holds, classifiedFixture())
	if err != nil {
		t.Fatalf("PlanPickup on a parallel card: %v", err)
	}
	if !d.Allowed || !d.MultiPick {
		t.Fatalf("decision = %+v, want allowed with MultiPick=true for parallel work", d)
	}
}

// contendPickup races two simulated lanes for one card over a shared hold
// set — test-local concurrency in place of a real cross-session lease. The
// take re-checks the decision under the lock, mirroring the F1 lease's
// version-checked claim; the function returns how many lanes ended holding.
func contendPickup(t *testing.T, reader Classifier, card string) int {
	t.Helper()
	var mu sync.Mutex
	var holds []Hold
	var wg sync.WaitGroup
	winners := 0
	for _, lane := range []string{"lane-1", "lane-2"} {
		wg.Add(1)
		go func(lane string) {
			defer wg.Done()
			for attempt := 0; attempt < 8; attempt++ {
				mu.Lock()
				snapshot := append([]Hold(nil), holds...)
				mu.Unlock()
				d, err := PlanPickup(lane, card, snapshot, reader)
				if err != nil {
					t.Errorf("lane %s: PlanPickup: %v", lane, err)
					return
				}
				if !d.Allowed {
					return // this lane waits / proceeds to a different pick
				}
				mu.Lock()
				recheck, err := PlanPickup(lane, card, holds, reader)
				if err == nil && recheck.Allowed {
					holds = append(holds, Hold{Lane: lane, Card: card})
					winners++
				}
				mu.Unlock()
				if err == nil && recheck.Allowed {
					return
				}
			}
		}(lane)
	}
	wg.Wait()
	return winners
}

// AC-FLA-006 (contention): two lanes reaching pickup simultaneously over one
// sequential group end with exactly one lane picking it.
func TestPickupSequentialGroupConcurrentContentionSingleWinner(t *testing.T) {
	if got := contendPickup(t, classifiedFixture(), "t1"); got != 1 {
		t.Fatalf("sequential contention winners = %d, want exactly 1", got)
	}
}

// AC-FLA-006 (contention, parallel): two lanes picking a parallel-classified
// card concurrently both succeed — two observable holds.
func TestPickupParallelConcurrentContentionTwoWinners(t *testing.T) {
	if got := contendPickup(t, classifiedFixture(), "tp"); got != 2 {
		t.Fatalf("parallel contention winners = %d, want 2 (both lanes hold)", got)
	}
}

// AC-FLA-007: a queued card with NO classification metadata is evaluated
// under the fallback classification — the lane does not autonomously
// multi-pick it and the decision carries the observable single-dispatch
// behavior marker. The evaluation exits without error (REQ-FLA-008 absent
// arm): absent metadata is a tolerated value, never a failure.
func TestPickupUnclassifiedCardFallsBackToSingleDispatch(t *testing.T) {
	d, err := PlanPickup("lane-1", "t1", nil, StaticClassifier{})
	if err != nil {
		t.Fatalf("PlanPickup on an unclassified card: %v", err)
	}
	if !d.Allowed {
		t.Fatalf("decision = %+v, want the fallback pickup allowed", d)
	}
	if !d.Fallback {
		t.Errorf("decision = %+v, want Fallback=true (the observable single-dispatch behavior marker)", d)
	}
	if d.MultiPick {
		t.Errorf("decision = %+v, want MultiPick=false — an unclassified card is never autonomously multi-picked", d)
	}
	if !d.ToleratedUnknown {
		t.Errorf("decision = %+v, want ToleratedUnknown=true (absent metadata logged, not errored)", d)
	}
	if !strings.Contains(d.Reason, "single-dispatch") {
		t.Errorf("reason = %q, want it to name the single-dispatch fallback behavior", d.Reason)
	}
}

// The consumption seam's absent-metadata contract: a card the reader knows
// nothing about reads back as an unknown classification with a nil error —
// the pickup then follows the fallback classification (REQ-FLA-007/008).
func TestStaticClassifierAbsentMetadataTolerated(t *testing.T) {
	cls, err := StaticClassifier{}.Classify("t-missing")
	if err != nil {
		t.Fatalf("Classify of a card with no metadata: %v", err)
	}
	if cls.Known {
		t.Fatalf("classification = %+v, want Known=false for absent metadata", cls)
	}
	if cls.Axis != AxisFallback {
		t.Errorf("axis = %q, want %q for absent metadata", cls.Axis, AxisFallback)
	}
}

// AC-FLA-008: the normalization point maps the two recognized axis tokens
// onto known classifications and everything else — an empty axis field, an
// unrecognized token — onto the unknown fallback classification.
func TestNormalizeClassificationMapsRecognizedAxes(t *testing.T) {
	seq := NormalizeClassification("sequential", "alpha")
	if !seq.Known || seq.Axis != AxisSequential || seq.Group != "alpha" {
		t.Fatalf("sequential normalization = %+v, want known sequential group alpha", seq)
	}
	par := NormalizeClassification("parallel", "")
	if !par.Known || par.Axis != AxisParallel {
		t.Fatalf("parallel normalization = %+v, want known parallel", par)
	}
	for _, raw := range []string{"", "banana", "SEQUENTIAL", "priority-only"} {
		got := NormalizeClassification(raw, "alpha")
		if got.Known || got.Axis != AxisFallback {
			t.Errorf("normalization of %q = %+v, want Known=false fallback", raw, got)
		}
	}
}

// AC-FLA-008: a metadata record carrying an axis value the consumer does not
// recognize is tolerated — the decision exits without error under the
// fallback classification and logs the tolerated-unknown condition.
func TestPickupUnknownAxisValueTolerated(t *testing.T) {
	reader := StaticClassifier{"t1": NormalizeClassification("banana", "alpha")}
	d, err := PlanPickup("lane-1", "t1", nil, reader)
	if err != nil {
		t.Fatalf("unknown axis value must not error (exit 0): %v", err)
	}
	if !d.Fallback || !d.ToleratedUnknown {
		t.Fatalf("decision = %+v, want the fallback classification with the tolerated-unknown marker", d)
	}
	if !strings.Contains(d.Reason, "tolerated") {
		t.Errorf("reason = %q, want it to log the tolerated-unknown condition", d.Reason)
	}
}
