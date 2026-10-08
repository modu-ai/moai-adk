// todo_auto_lane_test.go — card t1554: the unified `--auto` engine's lane
// surface. A lane session running `moai todo --auto` executes the serial
// cycle's foreman contract with the queue writes routed through the lease
// edges (the nominated lease `moai factory next --card <id>` takes), the
// selection driven by the shared ranking stage (Jev when the capability
// answers, else the priority/readiness fallback), and quota steering applied
// (a hold admits only the lane's own in-flight card).
//
// Every fixture is built under t.TempDir() with the lane environment scrubbed
// first (nmBase), then stamped per test (nmLaneEnv); the run resolution, the
// ranking seams, and the worktree creator are stubbed so no test contacts a
// real run, `gh`, Jev, or git.
package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/jev"
)

// laneAutoSeams installs the hermetic seam set the lane cycle shares with the
// serial cycle: the ranking seams (the live landed seam runs `gh`, the live
// Jev ranker contacts the endpoint) are stubbed inert, and the run resolution
// returns the fixture run id.
func laneAutoSeams(t *testing.T) {
	t.Helper()
	origLanded, origJev := todoAutoLandedLookup, todoAutoJevRanker
	origRun := autoLaneRunResolveFn
	t.Cleanup(func() {
		todoAutoLandedLookup, todoAutoJevRanker, autoLaneRunResolveFn = origLanded, origJev, origRun
	})
	todoAutoLandedLookup = func(*factory.BacklogRecord) (map[string]factory.PRLinkKind, error) {
		return nil, nil
	}
	todoAutoJevRanker = func(jev.Request) jev.Result { return jev.Result{Availability: jev.Disabled} }
	autoLaneRunResolveFn = func(context.Context, string) (string, error) { return fcRun, nil }
}

// laneAutoJevScores replaces the ranker with one that answers every question
// the ranking stage asks with the given score (confidence 0.9), so a test
// names the processing order it wants.
func laneAutoJevScores(t *testing.T, scores map[string]float64) {
	t.Helper()
	orig := todoAutoJevRanker
	t.Cleanup(func() { todoAutoJevRanker = orig })
	todoAutoJevRanker = func(req jev.Request) jev.Result {
		answers := make([]jev.Answer, 0, len(req.Questions))
		for _, q := range req.Questions {
			score, ok := scores[q.ID]
			if !ok {
				t.Errorf("jev ranker asked about unplanned card %s", q.ID)
				continue
			}
			answers = append(answers, jev.Answer{
				QuestionID: q.ID, Kind: jev.KindScore, Score: score, Probability: 0.9,
			})
		}
		return jev.Result{Availability: jev.Available, Answers: answers}
	}
}

// laneAutoSeedEvidence writes the evidence file the directive names for each
// id, so the cycle's wait collects at the first poll.
func laneAutoSeedEvidence(t *testing.T, root string, ids ...string) {
	t.Helper()
	for _, id := range ids {
		path := autoEvidencePath(root, id)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("# evidence\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// laneAutoTickClock returns a clock/sleep pair where every sleep advances the
// clock past a five-minute deadline, so a miss path exits after one poll.
func laneAutoTickClock(tick *int) (func(time.Duration), func() time.Time) {
	return func(time.Duration) { *tick++ },
		func() time.Time { return fcNow.Add(time.Duration(*tick) * time.Minute) }
}

// laneAutoOpts builds the cycle options with the (already stubbed) package
// ranking seams wired in — a nil seam is INERT by contract, so a direct call
// must carry them explicitly, exactly as the production wiring does.
func laneAutoOpts(t *testing.T, wait time.Duration, sleep func(time.Duration), now func() time.Time) autoOptions {
	t.Helper()
	return autoOptions{
		wait:     wait,
		sleep:    sleep,
		now:      now,
		landed:   todoAutoLandedLookup,
		jevRank:  todoAutoJevRanker,
		liveness: newAutoLiveness(),
	}
}

// TestAutoLaneCycleJevRankBeatsQueueOrderAndContinues — the two selection
// cells of the unified engine in one run: the Jev rank (t2 = 4, t1 = 0)
// overrides the queue order (t1 was added first), and after the first card's
// evidence is collected the loop CONTINUES to the next candidate and stops
// only when the queue is exhausted (REQ-SD-003). The cycle writes no done:
// the cards leave the queue only through the existing completion path.
func TestAutoLaneCycleJevRankBeatsQueueOrderAndContinues(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
	nmLaneEnv(t, "lane-1", "")
	nmIsolatedWorktrees(t, "t1", "t2")
	laneAutoSeams(t)
	laneAutoJevScores(t, map[string]float64{"t1": 0, "t2": 4})
	laneAutoSeedEvidence(t, root, "t1", "t2")

	tick := 0
	sleep, now := laneAutoTickClock(&tick)
	var out, errOut bytes.Buffer
	if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, 5*time.Minute, sleep, now)); err != nil {
		t.Fatalf("lane cycle: %v (stderr %q)", err, errOut.String())
	}
	got := out.String()
	t2At := strings.Index(got, "t2 stage=")
	t1At := strings.Index(got, "t1 stage=")
	if t2At < 0 || t1At < 0 {
		t.Fatalf("lease head lines missing from the cycle output:\n%s", got)
	}
	if t2At > t1At {
		t.Errorf("queue order won the selection (t1 leased before t2); the jev rank must pick first:\n%s", got)
	}
	if !strings.Contains(got, "selection: source=jev") {
		t.Errorf("the shared ranking stage's selection record is missing:\n%s", got)
	}
	if strings.Contains(got, "\ndone t") {
		t.Errorf("the lane cycle recorded a done; a lane's only queue write is the lease:\n%s", got)
	}
	if !strings.Contains(got, "complete t2") || !strings.Contains(got, "complete t1") {
		t.Errorf("completion lines missing:\n%s", got)
	}
	if !strings.Contains(got, "evidence collected") {
		t.Errorf("evidence lines missing:\n%s", got)
	}
	if s := nmQueueState(t, store, "t1"); s != factory.BacklogStatePicked {
		t.Errorf("t1 queue state = %s, want picked (leased, closed later by the existing completion path)", s)
	}
	if s := nmQueueState(t, store, "t2"); s != factory.BacklogStatePicked {
		t.Errorf("t2 queue state = %s, want picked", s)
	}
	nmAssertLeased(t, root, "t1", "lane-1")
	nmAssertLeased(t, root, "t2", "lane-1")
	// REQ-SD-011 (ported from the removed boot loops): the lease records the
	// card's worktree — materialized through the shared worktree creator and
	// bound on the card row.
	for _, id := range []string{"t1", "t2"} {
		if c := fcCard(t, root, id); strings.TrimSpace(c.WorktreePath) == "" {
			t.Errorf("%s recorded no worktree; the cycle must run the REQ-SD-011 record step", id)
		}
	}
}

// TestAutoLaneCycleKeepSetRefusalFallsThrough — the keep-set gates live at the
// lease: the ranked top candidate whose text opens with the hold marker is
// refused by the nominated lease (factoryRefuseHoldMarker), the cycle notes
// the refusal and falls through to the next-ranked candidate. The parked card
// stays queued.
func TestAutoLaneCycleKeepSetRefusalFallsThrough(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
	nmSetText(t, store, "t1", nmHoldMarker+" parked for the operator's decision")
	nmLaneEnv(t, "lane-1", "")
	nmIsolatedWorktrees(t, "t2")
	laneAutoSeams(t)
	laneAutoJevScores(t, map[string]float64{"t1": 4, "t2": 3})
	laneAutoSeedEvidence(t, root, "t2")

	tick := 0
	sleep, now := laneAutoTickClock(&tick)
	var out, errOut bytes.Buffer
	if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, 5*time.Minute, sleep, now)); err != nil {
		t.Fatalf("lane cycle: %v (stderr %q)", err, errOut.String())
	}
	if !strings.Contains(errOut.String(), "hold-marker") {
		t.Errorf("the hold-marker refusal is not narrated: %q", errOut.String())
	}
	if !strings.Contains(out.String(), "t2 stage=") {
		t.Errorf("the next-ranked candidate was not leased:\n%s", out.String())
	}
	if s := nmQueueState(t, store, "t1"); s != factory.BacklogStateQueued {
		t.Errorf("t1 queue state = %s, want queued (the parked card is never taken)", s)
	}
	if fcHasCard(t, root, "t1") {
		t.Errorf("t1 gained a factory record row; the keep-set refusal must leave it untouched")
	}
	nmAssertLeased(t, root, "t2", "lane-1")
}

// TestAutoLaneCycleQueuedCandidatesOnly — the lane cycle's candidate list is
// the queued state alone: a held card and a dropped card are never selected
// (positive enumeration), and the clean queued card leases.
func TestAutoLaneCycleQueuedCandidatesOnly(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateHold, factory.BacklogStateDropped, factory.BacklogStateQueued)
	nmLaneEnv(t, "lane-1", "")
	nmIsolatedWorktrees(t, "t3")
	laneAutoSeams(t)
	laneAutoSeedEvidence(t, root, "t3")

	tick := 0
	sleep, now := laneAutoTickClock(&tick)
	var out, errOut bytes.Buffer
	if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, 5*time.Minute, sleep, now)); err != nil {
		t.Fatalf("lane cycle: %v (stderr %q)", err, errOut.String())
	}
	if !strings.Contains(out.String(), "t3 stage=") {
		t.Errorf("the clean queued card was not leased:\n%s", out.String())
	}
	if fcHasCard(t, root, "t1") || fcHasCard(t, root, "t2") {
		t.Errorf("a held or dropped card gained a factory record row")
	}
	if s := nmQueueState(t, store, "t1"); s != factory.BacklogStateHold {
		t.Errorf("t1 queue state = %s, want hold", s)
	}
	if s := nmQueueState(t, store, "t2"); s != factory.BacklogStateDropped {
		t.Errorf("t2 queue state = %s, want dropped", s)
	}
	nmAssertLeased(t, root, "t3", "lane-1")
}

// TestAutoLaneCycleQuotaHoldTakesNoNewCard — quota steering: with pressure
// held, the cycle prints the hold line and takes no NEW card, even with
// evidence ready for one.
func TestAutoLaneCycleQuotaHoldTakesNoNewCard(t *testing.T) {
	root, store := qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5)})
	nmIsolatedWorktrees(t, "t1", "t2", "t3", "t4")
	laneAutoSeams(t)
	laneAutoSeedEvidence(t, root, "t1")

	tick := 0
	sleep, now := laneAutoTickClock(&tick)
	var out, errOut bytes.Buffer
	if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, 5*time.Minute, sleep, now)); err != nil {
		t.Fatalf("lane cycle: %v (stderr %q)", err, errOut.String())
	}
	if !strings.Contains(errOut.String(), "quota hold: ") {
		t.Errorf("the hold line is missing: %q", errOut.String())
	}
	if strings.Contains(out.String(), "stage=leased") {
		t.Errorf("a new card was leased under quota hold:\n%s", out.String())
	}
	if s := nmQueueState(t, store, "t1"); s != factory.BacklogStateQueued {
		t.Errorf("t1 queue state = %s, want queued (no new card under hold)", s)
	}
	if fcHasCard(t, root, "t1") {
		t.Errorf("t1 gained a factory record row under quota hold")
	}
}

// TestAutoLaneCycleQuotaHoldOwnAssignedCardProceeds — the other half of the
// steering rule: a hold leaves only the card already assigned to this lane
// leasable (arm (a) of the gated lease), and the cycle works it.
func TestAutoLaneCycleQuotaHoldOwnAssignedCardProceeds(t *testing.T) {
	root, store := qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5), assigned: true})
	nmIsolatedWorktrees(t, "t4")
	laneAutoSeams(t)
	laneAutoSeedEvidence(t, root, "t4")

	tick := 0
	sleep, now := laneAutoTickClock(&tick)
	var out, errOut bytes.Buffer
	if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, 5*time.Minute, sleep, now)); err != nil {
		t.Fatalf("lane cycle: %v (stderr %q)", err, errOut.String())
	}
	if !strings.Contains(errOut.String(), "quota hold: ") {
		t.Errorf("the hold line is missing: %q", errOut.String())
	}
	if !strings.Contains(out.String(), "t4 stage=") {
		t.Errorf("the lane's own assigned card was not leased:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "card: t4") {
		t.Errorf("the dispatch directive does not name the leased card:\n%s", out.String())
	}
	nmAssertLeased(t, root, "t4", "lane-1")
	if s := nmQueueState(t, store, "t1"); s != factory.BacklogStateQueued {
		t.Errorf("t1 queue state = %s, want queued (a hold admits no new card)", s)
	}
}

// TestAutoLaneCycleEvidenceMissKeepsLease — at the deadline with no evidence
// the cycle records the labelled non-finding and moves on WITHOUT unpicking:
// reverting the pick is a direct queue write, so the lease stands until the
// F1 expiry machinery returns the card (the same semantic the lane boot loops
// carry for a failed card session).
func TestAutoLaneCycleEvidenceMissKeepsLease(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateQueued)
	nmLaneEnv(t, "lane-1", "")
	nmIsolatedWorktrees(t, "t1")
	laneAutoSeams(t)

	tick := 0
	sleep, now := laneAutoTickClock(&tick)
	var out, errOut bytes.Buffer
	if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, time.Millisecond, sleep, now)); err != nil {
		t.Fatalf("lane cycle: %v (stderr %q)", err, errOut.String())
	}
	got := out.String()
	if !strings.Contains(got, "non-finding: t1") {
		t.Errorf("the deadline miss is not narrated:\n%s", got)
	}
	if !strings.Contains(got, "lease") {
		t.Errorf("the miss line does not name the lease-expiry path:\n%s", got)
	}
	if s := nmQueueState(t, store, "t1"); s != factory.BacklogStatePicked {
		t.Errorf("t1 queue state = %s, want picked (the cycle must not unpick)", s)
	}
	nmAssertLeased(t, root, "t1", "lane-1")
}

// TestAutoLaneCycleSerialSlotRefusalStops — a waitable refusal (another
// serial card holds the slot) ends the pass: retrying belongs to the caller's
// next invocation, not to re-ranking past the slot.
func TestAutoLaneCycleSerialSlotRefusalStops(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2", Stage: homestate.CardRun})
	nmLaneEnv(t, "lane-1", "")
	laneAutoSeams(t)

	tick := 0
	sleep, now := laneAutoTickClock(&tick)
	var out, errOut bytes.Buffer
	if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, 5*time.Minute, sleep, now)); err != nil {
		t.Fatalf("lane cycle: %v (stderr %q)", err, errOut.String())
	}
	if !strings.Contains(errOut.String(), "serial-slot") {
		t.Errorf("the serial-slot refusal is not narrated: %q", errOut.String())
	}
	if strings.Contains(out.String(), "stage=leased") {
		t.Errorf("a card was leased while the serial slot was held:\n%s", out.String())
	}
	if s := nmQueueState(t, store, "t2"); s != factory.BacklogStateQueued {
		t.Errorf("t2 queue state = %s, want queued", s)
	}
	if fcHasCard(t, root, "t2") {
		t.Errorf("t2 gained a factory record row past a waitable refusal")
	}
}

// TestAutoLaneCycleRunsThroughCommandSurface — the gate replacement, through
// the real flag surface: a lane session's `--auto` invocation runs the cycle
// where REQ-TAU-008 refused it. One variant exercises the command path because
// the lane predicate (todoLaneSession) is a single switch; the identity
// variants are covered by the direct tests and the nomination table.
func TestAutoLaneCycleRunsThroughCommandSurface(t *testing.T) {
	_, _ = nmBase(t, factory.BacklogStateQueued)
	nmIsolatedWorktrees(t, "t1")
	laneAutoSeams(t)
	t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")

	out, _, err := runTodo(t, "--auto", "--auto-wait", "1ms")
	if err != nil {
		t.Fatalf("a lane session's `moai todo --auto` failed: %v", err)
	}
	if !strings.Contains(out, "t1 stage=") {
		t.Errorf("command-path output shows no lease:\n%s", out)
	}
}

// TestAutoLaneCycleAssignedCardBeforeQueuedRanking — card t1577: a quota-free
// pass still owes the lane's own assigned card before any new-candidate
// ranking. The leader-dispatched card (t1: picked in the queue, leased to
// lane-1 in the factory record) is invisible to the queued-only candidate
// list, so the cycle skipped it entirely and worked only the fresh candidate.
func TestAutoLaneCycleAssignedCardBeforeQueuedRanking(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStatePicked, factory.BacklogStateQueued)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun})
	nmLaneEnv(t, "lane-1", "")
	nmIsolatedWorktrees(t, "t1", "t2")
	laneAutoSeams(t)
	laneAutoSeedEvidence(t, root, "t1", "t2")

	tick := 0
	sleep, now := laneAutoTickClock(&tick)
	var out, errOut bytes.Buffer
	if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, 5*time.Minute, sleep, now)); err != nil {
		t.Fatalf("lane cycle: %v (stderr %q)", err, errOut.String())
	}
	got := out.String()
	t1At := strings.Index(got, "t1 stage=")
	t2At := strings.Index(got, "t2 stage=")
	if t1At < 0 {
		t.Fatalf("the lane's own assigned card was never worked (the candidate list is queued-only):\n%s", got)
	}
	if t2At >= 0 && t2At < t1At {
		t.Errorf("a fresh queued candidate was ranked ahead of the lane's assigned card:\n%s", got)
	}
	if !strings.Contains(got, "card: t1") {
		t.Errorf("the dispatch directive does not name the assigned card:\n%s", got)
	}
	nmAssertLeased(t, root, "t1", "lane-1")
	if s := nmQueueState(t, store, "t2"); s != factory.BacklogStatePicked {
		t.Errorf("t2 queue state = %s, want picked (the cycle continues past the assigned card)", s)
	}
}

// TestAutoLaneCycleAssignedCardWithNoQueuedCandidates — card t1588: the exact
// reported instance, a quota-free pass whose only remaining card is the one
// assigned to this lane (zero queued candidates). The pre-#1804 cycle ranked
// the queued-only candidate list, found nothing, and returned "no work" with
// the assigned card still owed; the assigned arm precedes the ranking, so the
// pass leases the card instead of terminating.
func TestAutoLaneCycleAssignedCardWithNoQueuedCandidates(t *testing.T) {
	root, _ := nmBase(t, factory.BacklogStatePicked)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun})
	nmLaneEnv(t, "lane-1", "")
	nmIsolatedWorktrees(t, "t1")
	laneAutoSeams(t)
	laneAutoSeedEvidence(t, root, "t1")

	tick := 0
	sleep, now := laneAutoTickClock(&tick)
	var out, errOut bytes.Buffer
	if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, 5*time.Minute, sleep, now)); err != nil {
		t.Fatalf("lane cycle: %v (stderr %q)", err, errOut.String())
	}
	got := out.String()
	if !strings.Contains(got, "t1 stage=") {
		t.Fatalf("the cycle terminated with the lane's assigned card still owed (the t1588 starvation):\n%s", got)
	}
	if !strings.Contains(got, "card: t1") {
		t.Errorf("the dispatch directive does not name the assigned card:\n%s", got)
	}
	nmAssertLeased(t, root, "t1", "lane-1")
}

// TestAutoLaneCycleHonorsLauncherRunID — card t1577: the launcher names the
// factory run in the environment (config.EnvFactoryRunID) and the lane cycle
// must honor that selection, falling back to auto-discovery only when the
// environment names none. With two active runs the old auto-discovery-only
// resolution failed AMBIGUOUS_FACTORY even though the launcher had chosen.
func TestAutoLaneCycleHonorsLauncherRunID(t *testing.T) {
	t.Run("env selection wins over auto-discovery", func(t *testing.T) {
		root, _ := nmBase(t, factory.BacklogStateQueued)
		nmLaneEnv(t, "lane-1", "")
		nmIsolatedWorktrees(t, "t1")
		// The ranking seams are stubbed inert, but the RUN RESOLUTION stays
		// real: this cell is about the resolution's own precedence.
		origLanded, origJev := todoAutoLandedLookup, todoAutoJevRanker
		t.Cleanup(func() { todoAutoLandedLookup, todoAutoJevRanker = origLanded, origJev })
		todoAutoLandedLookup = func(*factory.BacklogRecord) (map[string]factory.PRLinkKind, error) { return nil, nil }
		todoAutoJevRanker = func(jev.Request) jev.Result { return jev.Result{Availability: jev.Disabled} }

		// Two active runs: without an explicit selection the resolution is
		// ambiguous; the launcher's env names one of them.
		db := fcOpen(t, root)
		for _, run := range []string{"run-a", "run-launcher"} {
			if _, err := db.DB.Exec(`INSERT INTO runs(run_id,status,created_at,updated_at) VALUES(?,'active','t','t')`, run); err != nil {
				t.Fatalf("seed run %s: %v", run, err)
			}
		}
		_ = db.Close()
		t.Setenv(config.EnvFactoryRunID, "run-launcher")
		laneAutoSeedEvidence(t, root, "t1")

		tick := 0
		sleep, now := laneAutoTickClock(&tick)
		var out, errOut bytes.Buffer
		if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, 5*time.Minute, sleep, now)); err != nil {
			t.Fatalf("lane cycle with a launcher-selected run: %v (stderr %q)", err, errOut.String())
		}
		db = fcOpen(t, root)
		c, err := db.LoadCard(context.Background(), "run-launcher", "t1")
		if err != nil {
			t.Fatalf("load t1 under run-launcher: %v", err)
		}
		_ = db.Close()
		if c.RunID != "run-launcher" {
			t.Errorf("t1 leased under run %q, want the launcher-selected run-launcher", c.RunID)
		}
	})
	t.Run("no env selection still fails closed on ambiguity", func(t *testing.T) {
		root, _ := nmBase(t, factory.BacklogStateQueued)
		nmLaneEnv(t, "lane-1", "")
		origLanded, origJev := todoAutoLandedLookup, todoAutoJevRanker
		t.Cleanup(func() { todoAutoLandedLookup, todoAutoJevRanker = origLanded, origJev })
		todoAutoLandedLookup = func(*factory.BacklogRecord) (map[string]factory.PRLinkKind, error) { return nil, nil }
		todoAutoJevRanker = func(jev.Request) jev.Result { return jev.Result{Availability: jev.Disabled} }

		db := fcOpen(t, root)
		for _, run := range []string{"run-a", "run-launcher"} {
			if _, err := db.DB.Exec(`INSERT INTO runs(run_id,status,created_at,updated_at) VALUES(?,'active','t','t')`, run); err != nil {
				t.Fatalf("seed run %s: %v", run, err)
			}
		}
		_ = db.Close()
		t.Setenv(config.EnvFactoryRunID, "")

		var out, errOut bytes.Buffer
		if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, time.Millisecond, func(time.Duration) {}, func() time.Time { return fcNow })); err == nil {
			t.Errorf("two active runs with no env selection resolved without an error; discovery must fail closed:\n%s", out.String())
		}
	})
	t.Run("retired env run id fails closed", func(t *testing.T) {
		root, store := nmBase(t, factory.BacklogStateQueued)
		nmLaneEnv(t, "lane-1", "")
		origLanded, origJev := todoAutoLandedLookup, todoAutoJevRanker
		t.Cleanup(func() { todoAutoLandedLookup, todoAutoJevRanker = origLanded, origJev })
		todoAutoLandedLookup = func(*factory.BacklogRecord) (map[string]factory.PRLinkKind, error) { return nil, nil }
		todoAutoJevRanker = func(jev.Request) jev.Result { return jev.Result{Availability: jev.Disabled} }

		db := fcOpen(t, root)
		if _, err := db.DB.Exec(`INSERT INTO runs(run_id,status,created_at,updated_at) VALUES('run-old','retired','t','t')`); err != nil {
			t.Fatalf("seed retired run: %v", err)
		}
		_ = db.Close()
		t.Setenv(config.EnvFactoryRunID, "run-old")

		var out, errOut bytes.Buffer
		if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, time.Millisecond, func(time.Duration) {}, func() time.Time { return fcNow })); err == nil {
			t.Errorf("a retired env run id leased anyway (t1 state %s); the selection must fail closed:\n%s", nmQueueState(t, store, "t1"), out.String())
		}
		if s := nmQueueState(t, store, "t1"); s != factory.BacklogStateQueued {
			t.Errorf("t1 queue state = %s, want queued (nothing leases against a retired run)", s)
		}
	})
}

// TestAutoLaneCycleSkipsExcludedAssignedCard — card t1577 card-review: an
// assigned card the operator later excluded (queue item dropped, or the text
// hold-marked) must not lease through the assigned arm — the same keep-set
// the nominated path applies. Assigned priority is preserved: the cycle skips
// the excluded card and works the fresh candidate.
func TestAutoLaneCycleSkipsExcludedAssignedCard(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateDropped, factory.BacklogStateQueued)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun})
	nmSetText(t, store, "t1", nmHoldMarker+" parked after dispatch")
	nmSetText(t, store, "t2", "the fresh candidate")
	nmLaneEnv(t, "lane-1", "")
	nmIsolatedWorktrees(t, "t1", "t2")
	laneAutoSeams(t)
	laneAutoSeedEvidence(t, root, "t2")

	tick := 0
	sleep, now := laneAutoTickClock(&tick)
	var out, errOut bytes.Buffer
	if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, 5*time.Minute, sleep, now)); err != nil {
		t.Fatalf("lane cycle: %v (stderr %q)", err, errOut.String())
	}
	got := out.String()
	if strings.Contains(got, "t1 stage=") {
		t.Errorf("an operator-excluded assigned card was leased anyway:\n%s", got)
	}
	if !strings.Contains(got, "t2 stage=") {
		t.Errorf("the fresh candidate was not worked after the excluded card was skipped:\n%s", got)
	}
	if s := nmQueueState(t, store, "t1"); s != factory.BacklogStateDropped {
		t.Errorf("t1 queue state = %s, want dropped (the exclusion stands)", s)
	}
	nmAssertLeased(t, root, "t2", "lane-1")
}

// TestAutoLaneCycleNoteNamesNextCandidate — a permanent refusal names the
// candidate it skipped, so the cycle's narration is checkable against the
// queue (one line per fall-through).
func TestAutoLaneCycleNoteNamesNextCandidate(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
	nmSetText(t, store, "t1", nmHoldMarker+" parked for the operator's decision")
	nmLaneEnv(t, "lane-1", "")
	nmIsolatedWorktrees(t, "t2")
	laneAutoSeams(t)
	laneAutoJevScores(t, map[string]float64{"t1": 4, "t2": 3})
	laneAutoSeedEvidence(t, root, "t2")

	tick := 0
	sleep, now := laneAutoTickClock(&tick)
	var out, errOut bytes.Buffer
	if err := runAutoLaneCycle(context.Background(), root, &out, &errOut, laneAutoOpts(t, 5*time.Minute, sleep, now)); err != nil {
		t.Fatalf("lane cycle: %v (stderr %q)", err, errOut.String())
	}
	if !strings.Contains(errOut.String(), "t1 refused (hold-marker)") {
		t.Errorf("the fall-through note does not name the card and token: %q", errOut.String())
	}
	if !strings.Contains(out.String(), "t2 stage=") {
		t.Errorf("the next-ranked candidate was not leased:\n%s", out.String())
	}
}
