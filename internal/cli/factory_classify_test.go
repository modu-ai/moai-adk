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

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// fcClassify sets one card's classification (the add path's write shape:
// classify AND re-sort inside one locked write).
func fcClassify(t *testing.T, store *kanban.BacklogStore, cardID, prio string, blocked bool, mode string) {
	t.Helper()
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID != cardID {
				continue
			}
			rec.Items[i].Classification = &kanban.CardClassification{
				Priority: prio, Blocked: blocked, Mode: mode, Decider: kanban.DeciderIdentityLLM,
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
	fcQueue(t, store, kanban.BacklogStateQueued, kanban.BacklogStateQueued, kanban.BacklogStateQueued)
	fcClassify(t, store, "t1", kanban.ClassPriorityHigh, false, kanban.ClassModeSerial)
	fcClassify(t, store, "t2", kanban.ClassPriorityLow, false, kanban.ClassModeSerial)
	fcClassify(t, store, "t3", kanban.ClassPriorityNormal, false, kanban.ClassModeParallelizable)
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
		if it.ID == "t2" && it.State != kanban.BacklogStateQueued {
			t.Errorf("t2 queue state = %s while the high serial is in flight, want still queued", it.State)
		}
	}

	// The boundary arm: t1 at merged-local — a LATE but NON-terminal state —
	// still holds the serial slot.
	fcSetCardState(t, root, "t1", homestate.CardMergedLocal)
	if _, _, err := runFactory(t, "next", "--run", fcRun); err == nil {
		t.Fatalf("next succeeded with t1 at merged-local; a non-terminal state must keep the serial slot held")
	} else if !strings.Contains(outOf(err), "no card") {
		t.Fatalf("expected the no-card exit at merged-local, got: %v", err)
	}

	// (c) t1 reaches a positively-enumerated terminal state (abandoned — an
	// irreversible state): the low serial card is then served.
	fcSetCardState(t, root, "t1", homestate.CardAbandoned)
	if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
		t.Fatalf("next after terminal: %v", err)
	}
	if c := fcCard(t, root, "t2"); c.State != homestate.CardLeased {
		t.Fatalf("t2 = %s, want the low serial leased after the high serial reached a terminal state", c.State)
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
	fcQueue(t, store, kanban.BacklogStateQueued, kanban.BacklogStateQueued)
	fcClassify(t, store, "t1", kanban.ClassPriorityNormal, false, kanban.ClassModeParallelizable)
	fcClassify(t, store, "t2", kanban.ClassPriorityNormal, false, kanban.ClassModeParallelizable)
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

// TestFactoryNextDuplicateDispatchGuard — AC-TCD-010 (regression guard over
// the t1240 version-check machinery): two lanes racing the SAME queued card
// end with exactly one holder; the loser converges elsewhere or no-card.
// The reason green is the version check is pinned separately by
// TestFactoryNextClaimRefusedMapsRace — the stale-version and holder
// mismatches map to a retry, and nothing else does.
func TestFactoryNextDuplicateDispatchGuard(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStateQueued)
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
	fcQueue(t, store, kanban.BacklogStateQueued, kanban.BacklogStateQueued)
	fcClassify(t, store, "t1", kanban.ClassPriorityHigh, false, kanban.ClassModeSerial)
	fcClassify(t, store, "t2", kanban.ClassPriorityNormal, false, kanban.ClassModeParallelizable)
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
