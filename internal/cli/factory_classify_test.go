// factory_classify_test.go — SPEC-TODO-CLASSIFY-DISPATCH-001 M3 acceptance
// tests: mode-aware factory dispatch on top of the t1240 self-dispatch
// surface. Priority-order auto-promotion with the blocked filter (AC-TCD-007),
// serial-card mutual exclusivity served in priority order while
// parallelizable selection stays unaffected (AC-TCD-008), concurrent
// multi-lane parallelizable leases (AC-TCD-009), the duplicate-dispatch
// regression guard over the version-checked claim edges (AC-TCD-010), and
// the status surface's mode/priority cells (AC-TCD-011).
package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// fcClassify sets one card's classification (the add path's write shape:
// classify AND re-sort inside one locked write).
func fcClassify(t *testing.T, store *factory.BacklogStore, cardID, prio string, blocked bool, mode string) {
	t.Helper()
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID != cardID {
				continue
			}
			rec.Items[i].Classification = &factory.CardClassification{
				Priority: prio, Blocked: blocked, Mode: mode, Decider: factory.DeciderIdentityLLM,
			}
			rec.SortByClassification()
			return nil
		}
		return fmt.Errorf("no card %s", cardID)
	}); err != nil {
		t.Fatalf("classify %s: %v", cardID, err)
	}
}

// fcSetCardState places one recorded card directly at the named state (a
// fixture placement in the fcPlace mold — the subject of these tests is the
// SELECTION reading the record's state value, not the transition machinery
// between stages, whose evidence gates are t1240's own surface).
func fcSetCardState(t *testing.T, root, cardID, state string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`UPDATE cards SET state = ? WHERE run_id = ? AND card_id = ?`, state, fcRun, cardID); err != nil {
		t.Fatalf("place %s at %s: %v", cardID, state, err)
	}
}

// TestFactoryNextSerialMutualExclusivity — AC-TCD-008: (a) a parallelizable
// card is leasable while a serial card is in flight; (b) no lane leases the
// second serial card before the first reaches a positively-enumerated
// terminal state, and the only-serial-candidates case ends on the existing
// no-card exit; (c) after the terminal state the low serial is served. A
// late NON-terminal state (merged-local) must keep the slot held — that is
// the boundary a terminal-enumeration mutation would break.
func TestFactoryNextSerialMutualExclusivity(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
	fcClassify(t, store, "t1", factory.ClassPriorityHigh, false, factory.ClassModeSerial)
	fcClassify(t, store, "t2", factory.ClassPriorityLow, false, factory.ClassModeSerial)
	fcClassify(t, store, "t3", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	// Lane A leases the high serial card.
	if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
		t.Fatalf("lane-1 next: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
		t.Fatalf("t1 = %s holder=%q, want lane-1's high serial lease", c.State, c.LeaseHolder)
	}

	// (a) lane B's next leases the parallelizable card — serial in flight
	// does not touch parallelizable selection.
	sdLaneEnv(t, "lane-2", "")
	if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
		t.Fatalf("lane-2 next (parallelizable): %v", err)
	}
	if c := fcCard(t, root, "t3"); c.State != homestate.CardLeased || c.LeaseHolder != "lane-2" {
		t.Fatalf("t3 = %s holder=%q, want lane-2's parallelizable lease while t1 is in flight", c.State, c.LeaseHolder)
	}

	// (b) no parallelizable candidate remains: the second serial card is not
	// leased — the existing no-card exit (exit 3), never a serial lease.
	for i := 0; i < 2; i++ {
		out, _, err := runFactory(t, "next", "--run", fcRun)
		sdExit3(t, fmt.Sprintf("only-serial-candidates (attempt %d)", i+1), err)
		if !strings.Contains(out, "no card") {
			t.Errorf("attempt %d stdout = %q, want a no-card line", i+1, out)
		}
	}
	if fcHasCard(t, root, "t2") {
		if c := fcCard(t, root, "t2"); c.State == homestate.CardLeased {
			t.Fatalf("t2 leased while high serial t1 is in flight — serial mutual exclusivity violated")
		}
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range rec.Items {
		if it.ID == "t2" && it.State != factory.BacklogStateQueued {
			t.Errorf("t2 queue state = %s while the high serial is in flight, want still queued", it.State)
		}
	}

	// The boundary arm: t1 at sync-audit — still INSIDE the implementation
	// pipeline — holds the serial slot.
	fcSetCardState(t, root, "t1", homestate.CardSyncAudit)
	if _, _, err := runFactory(t, "next", "--run", fcRun); err == nil {
		t.Fatalf("next succeeded with t1 at sync-audit; an implementation-pipeline state must keep the serial slot held")
	} else if !strings.Contains(outOf(err), "no card") {
		t.Fatalf("expected the no-card exit at sync-audit, got: %v", err)
	}
	// t1 reaching merge-ready — the enumerated boundary, where the card's
	// implementation is finished and the integration window takes over the
	// ordering — releases the slot for the low serial card.
	fcSetCardState(t, root, "t1", homestate.CardMergeReady)
	if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
		t.Fatalf("next after merge-ready: %v", err)
	}
	if c := fcCard(t, root, "t2"); c.State != homestate.CardLeased {
		t.Fatalf("t2 = %s, want the low serial leased once the high serial's implementation finished (merge-ready)", c.State)
	}

	// (c) a fully terminal state (abandoned — an irreversible exit) also
	// re-admits selection; here the low card is already served, so a further
	// next finds no card and exits 3.
	fcSetCardState(t, root, "t1", homestate.CardAbandoned)
	out, _, err := runFactory(t, "next", "--run", fcRun)
	sdExit3(t, "exhausted queue", err)
	if !strings.Contains(out, "no card") {
		t.Errorf("stdout = %q, want a no-card line", out)
	}
}

// TestFactoryNextSkipsClassificationBlocked — AC-TCD-007's blocked-exclusion
// pin (sync-audit finding F3): the `cls.Blocked → continue` skip in the
// auto-promotion arm (factoryNextSelectAndLease, factory_card.go) is the only
// thing standing between a high-priority blocked card and a lane's lease.
// Removing that skip turns this test red. The pin: repeated `next` leases
// only eligible cards in priority order, the blocked card is never leased
// while it stays blocked (and gains no record row), the only-blocked-candidate
// queue ends on the existing no-card exit 3 (the REQ-TCD-008-shaped
// ineligible-candidates boundary — a no-card answer, never a spin or a
// blocked lease), and the lease arrives only after the block lifts.
func TestFactoryNextSkipsClassificationBlocked(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
	// The blocked card carries the HIGHEST priority: without the skip, the
	// promotion loop would take it first and every assertion below fails.
	fcClassify(t, store, "t1", factory.ClassPriorityHigh, true, factory.ClassModeParallelizable)
	fcClassify(t, store, "t2", factory.ClassPriorityLow, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t3", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	sdRegisterLane(t, root, "lane-1")
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	// Repeated next leases only the eligible cards, in priority order —
	// never the blocked high card ahead of them.
	for _, want := range []string{"t3", "t2"} {
		if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
			t.Fatalf("next (%s expected): %v", want, err)
		}
		if c := fcCard(t, root, want); c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
			t.Fatalf("%s = %s holder=%q, want lane-1's lease", want, c.State, c.LeaseHolder)
		}
	}
	if fcHasCard(t, root, "t1") {
		t.Fatalf("t1 gained a record row while blocked — a blocked card must never be auto-selected")
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range rec.Items {
		if it.ID == "t1" && it.State != factory.BacklogStateQueued {
			t.Fatalf("t1 queue state = %s while blocked, want still queued", it.State)
		}
	}

	// Only the blocked candidate remains: the no-card exit (exit 3) — the
	// ineligible-candidates answer, never a blocked lease.
	out, _, err := runFactory(t, "next", "--run", fcRun)
	sdExit3(t, "only-blocked-candidate", err)
	if !strings.Contains(out, "no card") {
		t.Errorf("stdout = %q, want a no-card line", out)
	}
	if fcHasCard(t, root, "t1") {
		t.Errorf("t1 gained a record row on the only-blocked-candidate attempt")
	}

	// The block lifts (the operator unblock path): the promotion arm now
	// leases the formerly blocked card.
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
		t.Fatalf("next after unblock: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
		t.Fatalf("t1 = %s holder=%q after unblock, want lane-1's lease", c.State, c.LeaseHolder)
	}
}

// outOf renders an error's message (test helper).
func outOf(err error) string { return fmt.Sprint(err) }

// TestFactoryNextParallelizableConcurrentLeases — AC-TCD-009: two lanes
// racing `next` over two distinct parallelizable cards both succeed and hold
// DIFFERENT cards, exactly one each (goroutine contention over the
// version-checked store, per plan E4 — the env cannot hold two lane labels,
// so the arms call the selection directly, the same shape the t1240 suite
// uses).
func TestFactoryNextParallelizableConcurrentLeases(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")

	type result struct {
		lane  string
		card  homestate.Card
		owned bool
		err   error
	}
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for _, lane := range []string{"lane-1", "lane-2"} {
		wg.Add(1)
		go func(lane string) {
			defer wg.Done()
			card, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, lane)
			results <- result{lane: lane, card: card, owned: owned, err: err}
		}(lane)
	}
	wg.Wait()
	close(results)

	held := map[string]string{} // card -> lane
	for r := range results {
		if r.err != nil {
			t.Errorf("lane %s errored: %v", r.lane, r.err)
			continue
		}
		if !r.owned {
			t.Errorf("lane %s leased nothing; distinct parallelizable cards must both lease under contention", r.lane)
			continue
		}
		if prev, dup := held[r.card.CardID]; dup {
			t.Errorf("card %s held by both %s and %s — a card is never leased twice", r.card.CardID, prev, r.lane)
		}
		held[r.card.CardID] = r.lane
	}
	if len(held) != 2 {
		t.Errorf("lanes hold %d distinct cards, want 2 (t1, t2)", len(held))
	}
}

// TestFactoryNextRecordAndClaimRaceOnLeasedRow — the CI interleaving from
// run 36577159420, pinned by construction instead of goroutines: lane-1
// leases t1 behind lane-2's back (between the ListCards snapshot and the
// queue-picked read), then lane-2's b2 arm calls RecordAndClaim on the
// already-leased card. RecordPicked returns the existing row unchanged when
// fields are empty — whatever state it is in — and the claimer must read
// that as "another lane took it", a race to re-select, not a hard error.
func TestFactoryNextRecordAndClaimRaceOnLeasedRow(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStateQueued)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")

	// The other lane's progress, exactly as the CI interleaving left it:
	// t1 recorded and driven picked → assigned → leased (v3) behind
	// lane-2's back, while the queue item still reads picked.
	db := fcOpen(t, root)
	ctx := context.Background()
	picked, err := db.RecordPicked(ctx, fcRun, "t1", homestate.CardFields{}, "factory-next", factoryCardNow())
	if err != nil {
		t.Fatalf("record t1: %v", err)
	}
	assigned, err := db.Transition(ctx, homestate.TransitionRequest{
		RunID: fcRun, CardID: "t1", To: homestate.CardAssigned,
		ExpectedVersion: picked.Version, Actor: "factory-next", Owner: "lane-1", Now: factoryCardNow(),
	})
	if err != nil {
		t.Fatalf("assign t1: %v", err)
	}
	if _, err := db.Transition(ctx, homestate.TransitionRequest{
		RunID: fcRun, CardID: "t1", To: homestate.CardLeased,
		ExpectedVersion: assigned.Version, Actor: "lane-1", Now: factoryCardNow(),
	}); err != nil {
		t.Fatalf("lease t1: %v", err)
	}

	// lane-2 claims t1 through the b2 arm: no error, a race signal.
	card, owned, raced, err := factoryNextRecordAndClaim(ctx, db, root, fcRun, "t1", "lane-2")
	if err != nil {
		t.Fatalf("RecordAndClaim on an already-leased card: %v", err)
	}
	if owned || raced == false {
		t.Fatalf("RecordAndClaim = (%s, owned=%v, raced=%v), want no ownership and raced=true", card.CardID, owned, raced)
	}

	// The re-select leases the OTHER card, not the taken one.
	leased, owned2, err := factoryNextLeaseOnce(ctx, root, fcRun, "lane-2")
	if err != nil {
		t.Fatalf("re-select after race: %v", err)
	}
	if !owned2 || leased.CardID != "t2" || leased.LeaseHolder != "lane-2" {
		t.Fatalf("lane-2 holds %s holder=%s owned=%v after re-select, want t2/lane-2", leased.CardID, leased.LeaseHolder, owned2)
	}
}

// TestFactoryNextDuplicateDispatchGuard — AC-TCD-010 (regression guard over
// the t1240 version-check machinery): two lanes racing the SAME queued card
// end with exactly one holder; the loser converges elsewhere or no-card.
// The reason green is the version check is pinned separately by
// TestFactoryNextClaimRefusedMapsRace — the stale-version and holder
// mismatches map to a retry, and nothing else does.
func TestFactoryNextDuplicateDispatchGuard(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued)
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")

	var wg sync.WaitGroup
	holders := make(chan string, 2)
	for _, lane := range []string{"lane-1", "lane-2"} {
		wg.Add(1)
		go func(lane string) {
			defer wg.Done()
			_, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, lane)
			if err == nil && owned {
				holders <- lane
			}
		}(lane)
	}
	wg.Wait()
	close(holders)

	var winners []string
	for lane := range holders {
		winners = append(winners, lane)
	}
	if len(winners) != 1 {
		t.Errorf("t1 held by %v, want exactly one lane", winners)
	}
	c := fcCard(t, root, "t1")
	if c.State != homestate.CardLeased || len(winners) != 1 || c.LeaseHolder != winners[0] {
		t.Errorf("t1 = %s holder=%q winners=%v; the record must name exactly the winning lane's lease", c.State, c.LeaseHolder, winners)
	}
}

// TestFactoryNextClaimRefusedMapsRace — the AC-TCD-010 mechanism pin: a
// stale version or a lease-holder mismatch is a race to retry; any other
// error is a real error. This is why the duplicate-dispatch guard is green —
// the mapping, not luck.
func TestFactoryNextClaimRefusedMapsRace(t *testing.T) {
	card, raced, _, err := factoryNextClaimRefused(fmt.Errorf("wrapped: %w", homestate.ErrStaleVersion))
	if err != nil || raced || card != (homestate.Card{}) {
		t.Errorf("ErrStaleVersion: (card=%v raced=%v err=%v), want (zero false nil)", card, raced, err)
	}
	card, raced, _, err = factoryNextClaimRefused(fmt.Errorf("wrapped: %w", homestate.ErrLeaseHolder))
	if err != nil || raced || card != (homestate.Card{}) {
		t.Errorf("ErrLeaseHolder: (card=%v raced=%v err=%v), want (zero false nil)", card, raced, err)
	}
	real := errors.New("record closed")
	if _, raced, _, err := factoryNextClaimRefused(real); err == nil || raced {
		t.Errorf("unrelated error mapped to (raced=%v err=%v), want (false, the error)", raced, err)
	}
}

// TestFactoryStatusShowsHolderModePriority — AC-TCD-011: the factory record
// and `factory status` (text and --json) show which lane holds which card
// together with the card's mode and priority, and both output forms agree.
func TestFactoryStatusShowsHolderModePriority(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued)
	fcClassify(t, store, "t1", factory.ClassPriorityHigh, false, factory.ClassModeSerial)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardLeased, LeaseHolder: "lane-1", Stage: homestate.CardRun},
		homestate.Card{CardID: "t2", State: homestate.CardLeased, LeaseHolder: "lane-2", Stage: homestate.CardRun},
	)

	textOut, _, err := runFactory(t, "status", "--run", fcRun)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(textOut, "t1") || !strings.Contains(textOut, "mode=serial") || !strings.Contains(textOut, "priority=high") {
		t.Errorf("text status missing t1's mode/priority cells:\n%s", textOut)
	}
	if !strings.Contains(textOut, "mode=parallelizable") || !strings.Contains(textOut, "priority=normal") {
		t.Errorf("text status missing t2's mode/priority cells:\n%s", textOut)
	}
	if !strings.Contains(textOut, "lane-1") || !strings.Contains(textOut, "lane-2") {
		t.Errorf("text status missing the lease holders:\n%s", textOut)
	}

	jsonOut, _, err := runFactory(t, "status", "--run", fcRun, "--json")
	if err != nil {
		t.Fatalf("status --json: %v", err)
	}
	for _, want := range []string{`"card_id": "t1"`, `"mode": "serial"`, `"priority": "high"`, `"card_id": "t2"`, `"mode": "parallelizable"`, `"priority": "normal"`, `"lease_holder": "lane-1"`, `"lease_holder": "lane-2"`} {
		if !strings.Contains(jsonOut, want) {
			t.Errorf("json status missing %s", want)
		}
	}
}
