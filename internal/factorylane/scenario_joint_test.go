package factorylane

// scenario_joint_test.go — the AC-FLA-017 joint three-lane scenario (the M5
// cross-fragment test). ONE test process simulates the mixed workload the
// acceptance criterion describes: one lane declaring messaging fallback (its
// transition logged), two lanes contending over classified cards (one
// sequential group that ends with exactly one winner, one parallel pair that
// ends with both holding), and a merge that proceeds ONLY inside an acquired
// integration window. Every step appends to one joint log, printed at the end
// as the single observable scenario output.
//
// The whole simulation rides the existing packages' APIs against a throwaway
// store in t.TempDir with the FakeClock — no real cross-session state, no
// real window, no real origin, no real merge anywhere. The GitRunner is the
// merge_test.go scripted fake; the window is the forged WindowSnapshot the
// package documents as its test seam.

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// scenarioContend races the named lanes for one card over a shared hold set —
// the contendPickup pattern (pickup_test.go) with scenario logging and
// caller-chosen lanes: a take is decided on a snapshot, re-checked under the
// lock, and only a recheck that still allows records the hold. Returns how
// many lanes ended holding the card.
func scenarioContend(t *testing.T, lanes []string, card string, reader Classifier, holds *[]Hold, step func(string, ...any)) int {
	t.Helper()
	var mu sync.Mutex
	var wg sync.WaitGroup
	winners := 0
	for _, lane := range lanes {
		wg.Add(1)
		go func(lane string) {
			defer wg.Done()
			for attempt := 0; attempt < 8; attempt++ {
				mu.Lock()
				snapshot := append([]Hold(nil), *holds...)
				mu.Unlock()
				d, err := PlanPickup(lane, card, snapshot, reader)
				if err != nil {
					t.Errorf("lane %s: PlanPickup: %v", lane, err)
					return
				}
				if !d.Allowed {
					mu.Lock()
					step("%s waits — %s", lane, d.Reason)
					mu.Unlock()
					return
				}
				mu.Lock()
				recheck, err := PlanPickup(lane, card, *holds, reader)
				if err == nil && recheck.Allowed {
					*holds = append(*holds, Hold{Lane: lane, Card: card})
					winners++
					now := winners // local copy under the lock; the log reads it after unlock
					mu.Unlock()
					step("%s holds %s (holders now: %d)", lane, card, now)
					return
				}
				mu.Unlock()
				if err == nil {
					mu.Lock()
					step("%s re-check denied — %s", lane, recheck.Reason)
					mu.Unlock()
					return
				}
			}
		}(lane)
	}
	wg.Wait()
	return winners
}

// AC-FLA-017: three lanes run one mixed workload; the sequential group never
// shows two simultaneous holds, the parallel pair does, the fallback lane's
// transition is logged, and the run's merge sits inside an acquired window.
func TestJointThreeLaneScenario(t *testing.T) {
	store, clock := newTestStore(t)

	var logMu sync.Mutex
	var log []string
	step := func(format string, args ...any) {
		logMu.Lock()
		defer logMu.Unlock()
		log = append(log, fmt.Sprintf(format, args...))
	}

	reader := StaticClassifier{
		"card-seq": {Axis: AxisSequential, Group: "gamma", Known: true},
		"card-par": {Axis: AxisParallel, Known: true},
	}

	// Act I — lane-fb's channel dies; the self-service switch is a logged,
	// countable event (REQ-FLA-003).
	clock.Current = base.Add(1 * time.Minute)
	ev, err := store.DeclareFallback("lane-fb", TriggerChannelUnavailable, "card-fb")
	if err != nil {
		t.Fatalf("lane-fb DeclareFallback: %v", err)
	}
	fbEvents, err := store.Transitions("lane-fb")
	if err != nil {
		t.Fatalf("lane-fb Transitions: %v", err)
	}
	if len(fbEvents) != 1 || fbEvents[0].Trigger != TriggerChannelUnavailable {
		t.Fatalf("lane-fb transition log = %+v, want exactly one channel-unavailable event", fbEvents)
	}
	step("act I   lane-fb declares fallback (trigger=%s card=%s at %s) — transition logged, %d event(s)",
		ev.Trigger, ev.Card, ev.At.Format(time.RFC3339), len(fbEvents))

	// Act II — lane-a and lane-b contend over the classified workload. The
	// sequential group ends with exactly one holder; the parallel pair with
	// two.
	holds := []Hold{}
	seqWinners := scenarioContend(t, []string{"lane-a", "lane-b"}, "card-seq", reader, &holds, step)
	if seqWinners != 1 {
		t.Fatalf("sequential group gamma holders = %d, want exactly 1 (never two simultaneous holds)", seqWinners)
	}
	parWinners := scenarioContend(t, []string{"lane-a", "lane-b"}, "card-par", reader, &holds, step)
	if parWinners != 2 {
		t.Fatalf("parallel card holders = %d, want 2 (the parallel pair both hold)", parWinners)
	}
	step("act II  sequential group gamma: %d holder; parallel card: %d holders (joint hold set: %d)",
		seqWinners, parWinners, len(holds))

	// Act III — lane-a's merge proceeds only inside an acquired window. The
	// triple runs first (REQ-FLA-009), the record lands, the window is taken
	// AFTER the checks, and the merge moment is covered (REQ-FLA-010).
	in, fake := tripleFixture(t, "complete")
	in.Lane = "lane-a"
	in.Card = "card-seq"
	run, err := EvaluateMergeTriple(in, fake)
	if err != nil {
		t.Fatalf("EvaluateMergeTriple: %v", err)
	}
	if !run.AllPassed {
		t.Fatalf("merge triple failed in the scenario: %+v", run)
	}
	recorded, err := store.RecordMergeCheckRun(run)
	if err != nil {
		t.Fatalf("RecordMergeCheckRun: %v", err)
	}
	step("act III lane-a merge triple all-passed (checks: %s), recorded at %s",
		checkNames(run), recorded.CheckedAt.Format(time.RFC3339))

	mergeAt := base.Add(5 * time.Minute)
	// Outside any window the merge is refused, and the refusal says why.
	ok, why := WindowCoversMerge(WindowSnapshot{}, "lane-a", mergeAt)
	if ok {
		t.Fatal("merge proceeded with no window held — REQ-FLA-010 violated")
	}
	step("act III merge refused outside the window: %s", why)

	// The window is acquired after the checks — the AC-FLA-009 order.
	acquireAt := base.Add(4 * time.Minute)
	cleared, why := VerifyRunBeforeAcquire(&recorded, acquireAt)
	if !cleared {
		t.Fatalf("recorded run does not clear the pre-acquire proof: %s", why)
	}
	window := WindowSnapshot{Held: true, Live: true, HolderName: "lane-a", AcquiredAt: acquireAt, KnownAt: true}
	ok, why = WindowCoversMerge(window, "lane-a", mergeAt)
	if !ok {
		t.Fatalf("merge refused inside lane-a's own live window: %s", why)
	}
	step("act III merge proceeds inside the window: %s", why)

	// The same window does not cover another lane's merge moment.
	ok, why = WindowCoversMerge(window, "lane-b", mergeAt)
	if ok {
		t.Fatal("lane-b merged inside lane-a's window — the window is per-lane")
	}
	step("act III lane-b refused inside lane-a's window: %s", why)

	// The joint observable — one count-by-lane read over the whole run.
	counts, err := store.TransitionCountsByLane()
	if err != nil {
		t.Fatalf("TransitionCountsByLane: %v", err)
	}
	var fbRow string
	for _, c := range counts {
		if c.Lane == "lane-fb" {
			fbRow = fmt.Sprintf("%s=%d", c.Lane, c.Count)
		}
	}
	if fbRow != "lane-fb=1" {
		t.Fatalf("count-by-lane = %+v, want lane-fb=1 in the joint summary", counts)
	}
	step("epilog  count-by-lane: %s (fallback transition logged exactly once)", fbRow)

	// The ONE joint scenario output — observable at the end.
	logMu.Lock()
	defer logMu.Unlock()
	t.Log("=== joint three-lane scenario ===\n" + strings.Join(log, "\n"))
}

// checkNames renders a run's condition names in recorded order, for the log.
func checkNames(run MergeCheckRun) string {
	names := make([]string, 0, len(run.Checks))
	for _, c := range run.Checks {
		if c.Passed {
			names = append(names, c.Name)
		}
	}
	return strings.Join(names, ",")
}
