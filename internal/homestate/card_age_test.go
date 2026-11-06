package homestate

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-CARD-AGE-ALARM-001 (card t1838), milestone M2: the read-only age
// measurement — the gate-wait anchor of REQ-CAA-003, the three separate
// signals of REQ-CAA-004 (stage checkpoint counting stage changes only, the
// injected commit source, the heartbeat), the newest-of no-progress basis of
// REQ-CAA-005 (creation event included, most-recent-transition fallback
// withdrawn), the timestamp-validity rule of plan-audit iter2 finding B3, and
// the class (a)/(b) coverage set of REQ-CAA-008. Every fixture seeds through
// the real record API on a temp-dir factory DB; only the future-dated
// timestamps of TestAgeAnchorsFutureTimestamps are stamped directly, because
// the API cannot produce a store row dated ahead of the reading clock.

// caaStep drives one transition at at, returning the card after it.
func caaStep(t *testing.T, db *FactoryDB, req TransitionRequest, at time.Time) Card {
	t.Helper()
	req.Now = at
	c, err := db.Transition(context.Background(), req)
	if err != nil {
		t.Fatalf("transition %s -> %s at %s: %v", req.CardID, req.To, at.Format(time.RFC3339), err)
	}
	return c
}

// caaExpire triggers the lazy lease expiry of the card at at (any request
// finds the expired lease first and commits the expiry return).
func caaExpire(t *testing.T, db *FactoryDB, runID, cardID string, version int64, at time.Time) Card {
	t.Helper()
	_, err := db.Transition(context.Background(), TransitionRequest{
		RunID: runID, CardID: cardID, To: CardLeased, ExpectedVersion: version, Actor: "worker-1", Now: at,
	})
	if !errors.Is(err, ErrLeaseExpired) {
		t.Fatalf("lazy expiry at %s: err = %v, want ErrLeaseExpired (the expiry itself commits)", at.Format(time.RFC3339), err)
	}
	c, err := db.LoadCard(context.Background(), runID, cardID)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// caaEventsOf filters the loader's flat slice down to one card's rows.
func caaEventsOf(t *testing.T, all []CardEvent, runID, cardID string) []CardEvent {
	t.Helper()
	var out []CardEvent
	for _, e := range all {
		if e.RunID == runID && e.CardID == cardID {
			out = append(out, e)
		}
	}
	return out
}

// caaNoCommit is the commit source of a card with no commit signal.
func caaNoCommit(string) (time.Time, bool) { return time.Time{}, false }

// caaAssertNoNegativeAge fails when any known age of a is negative — the
// rolled-back-clock invariant of REQ-CAA-004 (iter2 B3) on every path.
func caaAssertNoNegativeAge(t *testing.T, a CardAges) {
	t.Helper()
	for name, s := range map[string]AgeSignal{
		"gate-wait": a.GateWait, "stage": a.Stage, "commit": a.Commit,
		"activity": a.Activity, "no-progress": a.NoProgress,
	} {
		if s.Known && s.Age < 0 {
			t.Fatalf("%s age = %s, never negative on any path", name, s.Age)
		}
	}
}

// TestGateWaitAge covers REQ-CAA-003: the anchor is the created_at of the
// most recent card.transition or lease.expired event whose payload `to` equals
// the card's current state, with cards.updated_at the fallback only when no
// such event exists.
func TestGateWaitAge(t *testing.T) {
	ctx := context.Background()
	now := frNow.Add(30 * time.Hour)

	t.Run("kickoff-entry-event-not-updated-at", func(t *testing.T) {
		db := frOpen(t)
		repo := frNewRepo(t, true)
		frRegisterWorker(t, db, "worker-1")
		c := frFixtureCard(repo, "gw-kickoff", CardPlanAudit, CardKickoff)
		frPlace(t, db, c)
		c.Version = 1
		frWriteVerdictsFor(t, repo, "gw-kickoff", CardPlanAudit)
		entry := frNow.Add(30 * time.Minute) // inside the fixture's 1h lease
		caaStep(t, db, frFullRequest(repo, c, CardKickoff), entry)
		// A later row write stamps updated_at without appending a transition
		// (the card.fields write of card_picked.go:197 is the motivating
		// shape; the record API refuses it past `picked`, so the stamp is
		// seeded directly). The entering event, not the stamp, is the anchor.
		later := entry.Add(2 * time.Hour)
		if _, err := db.DB.Exec(`UPDATE cards SET updated_at=? WHERE run_id=? AND card_id=?`,
			later.Format(time.RFC3339Nano), frRun, "gw-kickoff"); err != nil {
			t.Fatal(err)
		}
		card, err := db.LoadCard(ctx, frRun, "gw-kickoff")
		if err != nil {
			t.Fatal(err)
		}
		events, err := db.LoadCardEvents(ctx, "")
		if err != nil {
			t.Fatal(err)
		}
		got := ComputeCardAges(card, caaEventsOf(t, events, frRun, "gw-kickoff"), caaNoCommit, now)
		if !got.GateWait.Known {
			t.Fatal("gate-wait age of a kickoff card with an entering event is unknown")
		}
		if want := now.Sub(entry); got.GateWait.Age != want {
			t.Fatalf("gate-wait age = %s, want %s (the entering event at %s, not updated_at at %s)",
				got.GateWait.Age, want, entry.Format(time.RFC3339), later.Format(time.RFC3339))
		}
	})

	t.Run("mid-merge-expiry-anchors-on-lease-expired-event", func(t *testing.T) {
		db := frOpen(t)
		repo := frNewRepo(t, true)
		frRegisterWorker(t, db, "worker-1")
		expired := frNow.Add(-time.Minute)
		frPlace(t, db, Card{
			RunID: frRun, CardID: "gw-blocked", State: CardMerging, Version: 1,
			OwnerLabel: "worker-1", WorktreePath: repo.Dir,
			LeaseHolder: "worker-1", LeaseExpiresAt: expired.Format(time.RFC3339Nano),
		})
		entry := frNow.Add(time.Hour)
		after := caaExpire(t, db, frRun, "gw-blocked", 1, entry)
		if after.State != CardBlocked {
			t.Fatalf("mid-merge expiry landed %s, want blocked", after.State)
		}
		events, err := db.LoadCardEvents(ctx, "")
		if err != nil {
			t.Fatal(err)
		}
		got := ComputeCardAges(after, caaEventsOf(t, events, frRun, "gw-blocked"), caaNoCommit, now)
		if !got.GateWait.Known {
			t.Fatal("gate-wait age of a mid-merge expired card is unknown")
		}
		if want := now.Sub(entry); got.GateWait.Age != want {
			t.Fatalf("gate-wait age = %s, want %s (the lease.expired event, not updated_at)", got.GateWait.Age, want)
		}
	})

	t.Run("no-events-row-falls-back-to-updated-at", func(t *testing.T) {
		db := frOpen(t)
		stamp := frNow.Add(-3 * time.Hour)
		frPlace(t, db, Card{RunID: frRun, CardID: "gw-legacy", State: CardKickoff, Version: 1, UpdatedAt: stamp.Format(time.RFC3339Nano)})
		events, err := db.LoadCardEvents(ctx, "")
		if err != nil {
			t.Fatal(err)
		}
		card, err := db.LoadCard(ctx, frRun, "gw-legacy")
		if err != nil {
			t.Fatal(err)
		}
		got := ComputeCardAges(card, caaEventsOf(t, events, frRun, "gw-legacy"), caaNoCommit, now)
		if !got.GateWait.Known {
			t.Fatal("gate-wait age of a no-events row is unknown (updated_at fallback)")
		}
		if want := now.Sub(stamp); got.GateWait.Age != want {
			t.Fatalf("gate-wait age = %s, want %s (updated_at fallback)", got.GateWait.Age, want)
		}
	})

	t.Run("needs-decision-and-blocked-are-gate-states", func(t *testing.T) {
		db := frOpen(t)
		repo := frNewRepo(t, true)
		frRegisterWorker(t, db, "worker-1")
		// T21 into needs-decision, then T23 into blocked — both real edges.
		c := frFixtureCard(repo, "gw-nd", CardPicked, CardNeedsDecision)
		frPlace(t, db, c)
		c.Version = 1
		entry := frNow.Add(30 * time.Minute)
		after := caaStep(t, db, frFullRequest(repo, c, CardNeedsDecision), entry)
		events, _ := db.LoadCardEvents(ctx, "")
		got := ComputeCardAges(after, caaEventsOf(t, events, frRun, "gw-nd"), caaNoCommit, now)
		if !got.GateWait.Known || got.GateWait.Age != now.Sub(entry) {
			t.Fatalf("needs-decision gate-wait = %+v, want %s from the entering event", got.GateWait, now.Sub(entry))
		}
		blocked := caaStep(t, db, frFullRequest(repo, after, CardBlocked), entry.Add(time.Minute))
		events, _ = db.LoadCardEvents(ctx, "")
		got = ComputeCardAges(blocked, caaEventsOf(t, events, frRun, "gw-nd"), caaNoCommit, now)
		if !got.GateWait.Known || got.GateWait.Age != now.Sub(entry.Add(time.Minute)) {
			t.Fatalf("blocked gate-wait = %+v, want the T23 entering event's age", got.GateWait)
		}
	})
}

// TestCardAgeSignals covers REQ-CAA-004: the three signals compute from three
// distinct sources (events / the injected commit source / heartbeat_at),
// render unknown for an empty worktree path and an empty heartbeat, and a
// fresh heartbeat renewal does NOT advance the no-progress age (activity is
// not progress — design principle 2).
func TestCardAgeSignals(t *testing.T) {
	ctx := context.Background()
	db := frOpen(t)
	repo := frNewRepo(t, true)
	frRegisterWorker(t, db, "worker-1")
	wp := repo.Dir
	spec := frSpecID
	pick, assign, lease, stage := frNow, frNow.Add(10*time.Minute), frNow.Add(20*time.Minute), frNow.Add(30*time.Minute)
	if _, err := db.RecordPicked(ctx, frRun, "sig", CardFields{WorktreePath: &wp, SpecID: &spec}, "op", pick); err != nil {
		t.Fatal(err)
	}
	c1 := caaStep(t, db, TransitionRequest{RunID: frRun, CardID: "sig", To: CardAssigned, ExpectedVersion: 1, Actor: "op", Owner: "worker-1"}, assign)
	c2 := caaStep(t, db, TransitionRequest{RunID: frRun, CardID: "sig", To: CardLeased, ExpectedVersion: c1.Version, Actor: "worker-1"}, lease)
	c3 := caaStep(t, db, TransitionRequest{RunID: frRun, CardID: "sig", To: CardRun, ExpectedVersion: c2.Version, Actor: "worker-1"}, stage)
	if c3.State != CardRun {
		t.Fatalf("walk landed %s, want run", c3.State)
	}

	commitAt := stage.Add(5 * time.Minute)
	now := frNow.Add(10 * time.Hour)
	commitSrc := func(path string) (time.Time, bool) {
		if path != repo.Dir {
			return time.Time{}, false
		}
		return commitAt, true
	}
	events, err := db.LoadCardEvents(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	ev := caaEventsOf(t, events, frRun, "sig")
	got := ComputeCardAges(c3, ev, commitSrc, now)

	if !got.Stage.Known || got.Stage.Age != now.Sub(stage) {
		t.Fatalf("stage signal = %+v, want %s from the T4b stage-change event", got.Stage, now.Sub(stage))
	}
	if !got.Commit.Known || got.Commit.Age != now.Sub(commitAt) {
		t.Fatalf("commit signal = %+v, want %s from the injected head-commit source", got.Commit, now.Sub(commitAt))
	}
	if !got.Activity.Known || got.Activity.Age != now.Sub(lease) {
		t.Fatalf("activity signal = %+v, want %s from heartbeat_at (the T3 lease write)", got.Activity, now.Sub(lease))
	}
	// The newest-of basis is the injected commit (newest of stage, commit,
	// creation — the commit lands after the stage entry here); the heartbeat
	// never enters it.
	if !got.NoProgress.Known || got.NoProgress.Age != now.Sub(commitAt) {
		t.Fatalf("no-progress age = %+v, want %s from the newest basis member (the commit), heartbeat excluded", got.NoProgress, now.Sub(commitAt))
	}

	// A fresh heartbeat renewal advances only the activity signal. The lease
	// is extended first so the renewal at +9h is a live-lease renewal, not an
	// expired-lease refusal.
	if _, err := db.DB.Exec(`UPDATE cards SET lease_expires_at=? WHERE run_id=? AND card_id=?`,
		frNow.Add(20*time.Hour).Format(time.RFC3339Nano), frRun, "sig"); err != nil {
		t.Fatal(err)
	}
	renewed := frNow.Add(9 * time.Hour)
	if _, err := db.RenewLease(ctx, frRun, "sig", "worker-1", renewed); err != nil {
		t.Fatal(err)
	}
	card, err := db.LoadCard(ctx, frRun, "sig")
	if err != nil {
		t.Fatal(err)
	}
	events, _ = db.LoadCardEvents(ctx, "")
	got = ComputeCardAges(card, caaEventsOf(t, events, frRun, "sig"), commitSrc, now)
	if !got.Activity.Known || got.Activity.Age != now.Sub(renewed) {
		t.Fatalf("activity after renewal = %+v, want %s", got.Activity, now.Sub(renewed))
	}
	if !got.NoProgress.Known || got.NoProgress.Age != now.Sub(commitAt) {
		t.Fatalf("no-progress after renewal = %+v, want still %s — activity is not progress", got.NoProgress, now.Sub(commitAt))
	}

	// Unknown renderings: an empty worktree path and an empty heartbeat.
	noPath := c3
	noPath.WorktreePath = ""
	noHeart := c3
	noHeart.HeartbeatAt = ""
	emptyCommit := ComputeCardAges(noPath, ev, commitSrc, now)
	if emptyCommit.Commit.Known {
		t.Fatal("commit signal of an empty worktree path is known, want unknown")
	}
	emptyBeat := ComputeCardAges(noHeart, ev, caaNoCommit, now)
	if emptyBeat.Activity.Known {
		t.Fatal("activity signal of an empty heartbeat is known, want unknown")
	}
	// The loader's bounded query carries transitions and expiries only — the
	// card.fields rows the anchors never read are not in the set.
	for _, e := range ev {
		if e.Kind == "card.fields" {
			t.Fatalf("LoadCardEvents returned a card.fields row: %+v", e)
		}
	}
}

// TestNoProgressAge covers REQ-CAA-005's both-unknown path: with the stage
// checkpoint and the commit both unknown, the age anchors on the card's
// creation event — never on the most recent transition (the iter2 B1 reset
// vector, withdrawn).
func TestNoProgressAge(t *testing.T) {
	ctx := context.Background()
	db := frOpen(t)
	frRegisterWorker(t, db, "worker-1")
	create, assign, lease := frNow, frNow.Add(time.Hour), frNow.Add(2*time.Hour)
	wp := frNewRepo(t, true).Dir
	if _, err := db.RecordPicked(ctx, frRun, "np", CardFields{WorktreePath: &wp}, "op", create); err != nil {
		t.Fatal(err)
	}
	c1 := caaStep(t, db, TransitionRequest{RunID: frRun, CardID: "np", To: CardAssigned, ExpectedVersion: 1, Actor: "op", Owner: "worker-1"}, assign)
	after := caaStep(t, db, TransitionRequest{RunID: frRun, CardID: "np", To: CardLeased, ExpectedVersion: c1.Version, Actor: "worker-1"}, lease)
	events, err := db.LoadCardEvents(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	ev := caaEventsOf(t, events, frRun, "np")
	now := frNow.Add(3 * time.Hour)
	got := ComputeCardAges(after, ev, caaNoCommit, now)
	if got.Stage.Known {
		t.Fatal("stage checkpoint of a card that never entered a pipeline stage is known")
	}
	if !got.NoProgress.Known || got.NoProgress.Age != now.Sub(create) {
		t.Fatalf("no-progress age = %+v, want %s from the creation event (not %s from the most recent transition)",
			got.NoProgress, now.Sub(create), now.Sub(lease))
	}
}

// TestNoProgressAgeCreationAnchor pins the iter2 rule-C cases (plan-audit
// findings B1/B2): a pre-stage crash loop preserves the stuck time from
// creation, a fresh card with an old inherited base commit measures ~0 from
// creation, and a genuinely stuck pre-stage card still reads its full age.
func TestNoProgressAgeCreationAnchor(t *testing.T) {
	ctx := context.Background()

	t.Run("pre-stage-crash-loop-preserves-age", func(t *testing.T) {
		db := frOpen(t)
		frRegisterWorker(t, db, "worker-1")
		wp := frNewRepo(t, true).Dir
		create := frNow
		if _, err := db.RecordPicked(ctx, frRun, "loop", CardFields{WorktreePath: &wp}, "op", create); err != nil {
			t.Fatal(err)
		}
		c := caaStep(t, db, TransitionRequest{RunID: frRun, CardID: "loop", To: CardAssigned, ExpectedVersion: 1, Actor: "op", Owner: "worker-1"}, create.Add(time.Minute))
		c = caaStep(t, db, TransitionRequest{RunID: frRun, CardID: "loop", To: CardLeased, ExpectedVersion: c.Version, Actor: "worker-1"}, create.Add(2*time.Minute))
		// >= 10 expiry -> re-acquire cycles (lease 15m default): each cycle is
		// activity, not progress, and none of them enters a pipeline stage.
		for i := 0; i < 10; i++ {
			at := create.Add(time.Duration(i+2) * 20 * time.Minute) // past the 15m lease
			c = caaExpire(t, db, frRun, "loop", c.Version, at)
			if c.State != CardAssigned {
				t.Fatalf("cycle %d: expiry landed %s, want assigned", i, c.State)
			}
			c = caaStep(t, db, TransitionRequest{RunID: frRun, CardID: "loop", To: CardLeased, ExpectedVersion: c.Version, Actor: "worker-1"}, at.Add(30*time.Second))
		}
		events, err := db.LoadCardEvents(ctx, "")
		if err != nil {
			t.Fatal(err)
		}
		now := create.Add(48 * time.Hour)
		got := ComputeCardAges(c, caaEventsOf(t, events, frRun, "loop"), caaNoCommit, now)
		if !got.NoProgress.Known {
			t.Fatal("no-progress age of the crash-looping card is unknown")
		}
		if want := now.Sub(create); got.NoProgress.Age != want {
			t.Fatalf("no-progress age after 10 cycles = %s, want %s from the creation event (pool-state cycling is activity, not progress)",
				got.NoProgress.Age, want)
		}
		if got.NoProgress.Age < config.DefaultCardAgeAlarmStallAlertAfter {
			t.Fatalf("no-progress age %s stays below the stall alert band %s — the crash loop this SPEC exists to surface is masked",
				got.NoProgress.Age, config.DefaultCardAgeAlarmStallAlertAfter)
		}
		caaAssertNoNegativeAge(t, got)
	})

	t.Run("fresh-card-old-inherited-base-measures-from-creation", func(t *testing.T) {
		db := frOpen(t)
		create := frNow
		wp := frNewRepo(t, true).Dir
		if _, err := db.RecordPicked(ctx, frRun, "fresh", CardFields{WorktreePath: &wp}, "op", create); err != nil {
			t.Fatal(err)
		}
		c, err := db.LoadCard(ctx, frRun, "fresh")
		if err != nil {
			t.Fatal(err)
		}
		events, _ := db.LoadCardEvents(ctx, "")
		inheritedBase := create.Add(-48 * time.Hour) // the worktree's head predates the card
		now := create.Add(time.Minute)
		got := ComputeCardAges(c, caaEventsOf(t, events, frRun, "fresh"), func(string) (time.Time, bool) {
			return inheritedBase, true
		}, now)
		if !got.NoProgress.Known {
			t.Fatal("no-progress age of the fresh card is unknown")
		}
		if got.NoProgress.Age > time.Hour {
			t.Fatalf("no-progress age of a fresh card with a 48h-old inherited base = %s, want ~0 measured from creation", got.NoProgress.Age)
		}
		if got.NoProgress.Age >= config.DefaultCardAgeAlarmStallWarnAfter {
			t.Fatalf("no-progress age %s crosses the stall warn band %s — the inherited-base false alert the creation member exists to prevent",
				got.NoProgress.Age, config.DefaultCardAgeAlarmStallWarnAfter)
		}
		caaAssertNoNegativeAge(t, got)
	})

	t.Run("stuck-pre-stage-card-still-alerts", func(t *testing.T) {
		db := frOpen(t)
		wp := frNewRepo(t, true).Dir
		create := frNow
		if _, err := db.RecordPicked(ctx, frRun, "stuck", CardFields{WorktreePath: &wp}, "op", create); err != nil {
			t.Fatal(err)
		}
		c, err := db.LoadCard(ctx, frRun, "stuck")
		if err != nil {
			t.Fatal(err)
		}
		events, _ := db.LoadCardEvents(ctx, "")
		now := create.Add(9 * time.Hour) // past the 8h stall alert band
		got := ComputeCardAges(c, caaEventsOf(t, events, frRun, "stuck"), caaNoCommit, now)
		if !got.NoProgress.Known {
			t.Fatal("no-progress age of the stuck card is unknown")
		}
		if want := now.Sub(create); got.NoProgress.Age != want {
			t.Fatalf("no-progress age = %s, want %s from creation (the true positive the clamp must keep)", got.NoProgress.Age, want)
		}
		if got.NoProgress.Age < config.DefaultCardAgeAlarmStallAlertAfter {
			t.Fatalf("no-progress age %s stays below the stall alert band %s — the stuck card is masked", got.NoProgress.Age, config.DefaultCardAgeAlarmStallAlertAfter)
		}
		caaAssertNoNegativeAge(t, got)
	})
}

// TestAgeAnchorsFutureTimestamps covers the iter2 B3 validity reach: with
// store sources dated ahead of the reading clock beyond
// CardAgeAlarmClockSkewTolerance (a rolled-back clock, or a DB read from
// another machine), every derived age renders unknown and none is negative —
// the gate-wait anchor event, its updated_at fallback, the creation-event
// anchor, and the legacy updated_at fallback.
func TestAgeAnchorsFutureTimestamps(t *testing.T) {
	ctx := context.Background()
	now := frNow
	future := now.Add(24 * time.Hour)
	if future.Sub(now) <= config.CardAgeAlarmClockSkewTolerance {
		t.Fatal("fixture future stamp does not exceed the tolerance")
	}
	caaStampEvents := func(t *testing.T, db *FactoryDB, to, created string) {
		t.Helper()
		if _, err := db.DB.Exec(`UPDATE events SET created_at=? WHERE kind='card.transition' AND payload_json LIKE ?`,
			created, fmt.Sprintf(`%%"to":%q%%`, to)); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("gate-anchor-event-future", func(t *testing.T) {
		db := frOpen(t)
		repo := frNewRepo(t, true)
		frRegisterWorker(t, db, "worker-1")
		c := frFixtureCard(repo, "fut-gate", CardPicked, CardNeedsDecision)
		frPlace(t, db, c)
		c.Version = 1
		caaStep(t, db, frFullRequest(repo, c, CardNeedsDecision), frNow)
		caaStampEvents(t, db, CardNeedsDecision, future.Format(time.RFC3339Nano))
		card, _ := db.LoadCard(ctx, frRun, "fut-gate")
		events, _ := db.LoadCardEvents(ctx, "")
		ev := caaEventsOf(t, events, frRun, "fut-gate")
		// The future-dated anchor renders that source unknown; the updated_at
		// fallback (valid, stamped by the same write) keeps the age answering.
		got := ComputeCardAges(card, ev, caaNoCommit, now)
		if !got.GateWait.Known || got.GateWait.Age != 0 {
			t.Fatalf("gate-wait age with a future anchor and a valid updated_at = %+v, want a known zero from the fallback", got.GateWait)
		}
		// With every basis member future-dated, no valid basis remains.
		if _, err := db.DB.Exec(`UPDATE cards SET updated_at=? WHERE run_id=? AND card_id=?`,
			future.Format(time.RFC3339Nano), frRun, "fut-gate"); err != nil {
			t.Fatal(err)
		}
		card, _ = db.LoadCard(ctx, frRun, "fut-gate")
		got = ComputeCardAges(card, ev, caaNoCommit, now)
		if got.GateWait.Known {
			t.Fatalf("gate-wait age with every basis member future-dated = %+v, want unknown", got.GateWait)
		}
		caaAssertNoNegativeAge(t, got)
	})

	t.Run("gate-updated-at-fallback-future", func(t *testing.T) {
		db := frOpen(t)
		frPlace(t, db, Card{RunID: frRun, CardID: "fut-fb", State: CardBlocked, Version: 1, UpdatedAt: future.Format(time.RFC3339Nano)})
		card, _ := db.LoadCard(ctx, frRun, "fut-fb")
		got := ComputeCardAges(card, nil, caaNoCommit, now)
		if got.GateWait.Known {
			t.Fatalf("gate-wait age from a future-dated updated_at fallback = %+v, want unknown", got.GateWait)
		}
		if got.NoProgress.Known {
			t.Fatalf("no-progress age from a future-dated updated_at fallback = %+v, want unknown", got.NoProgress)
		}
		caaAssertNoNegativeAge(t, got)
	})

	t.Run("creation-anchor-future", func(t *testing.T) {
		db := frOpen(t)
		wp := frNewRepo(t, true).Dir
		if _, err := db.RecordPicked(ctx, frRun, "fut-create", CardFields{WorktreePath: &wp}, "op", frNow); err != nil {
			t.Fatal(err)
		}
		caaStampEvents(t, db, CardPicked, future.Format(time.RFC3339Nano))
		card, _ := db.LoadCard(ctx, frRun, "fut-create")
		events, _ := db.LoadCardEvents(ctx, "")
		got := ComputeCardAges(card, caaEventsOf(t, events, frRun, "fut-create"), caaNoCommit, now)
		if got.NoProgress.Known {
			t.Fatalf("no-progress age with every basis member future-dated = %+v, want unknown (never negative)", got.NoProgress)
		}
		caaAssertNoNegativeAge(t, got)
	})

	t.Run("within-tolerance-skew-clamps-not-negates", func(t *testing.T) {
		// A source barely ahead of the reader (within the tolerance) is a
		// small clock skew, not a foreign record: it reads age zero, never
		// negative.
		db := frOpen(t)
		slight := frNow.Add(config.CardAgeAlarmClockSkewTolerance - time.Minute)
		frPlace(t, db, Card{RunID: frRun, CardID: "fut-slight", State: CardBlocked, Version: 1, UpdatedAt: slight.Format(time.RFC3339Nano)})
		card, _ := db.LoadCard(ctx, frRun, "fut-slight")
		got := ComputeCardAges(card, nil, caaNoCommit, now)
		if !got.GateWait.Known || got.GateWait.Age != 0 {
			t.Fatalf("gate-wait age from a within-tolerance skew = %+v, want a known zero", got.GateWait)
		}
		caaAssertNoNegativeAge(t, got)
	})
}

// TestStageCheckpointSameStageReEntry pins the iter1 D1 stage-change rule: an
// expiry -> re-acquire -> T4 same-stage re-entry does not refresh the stage
// checkpoint or the no-progress age, while a forward advance and the backward
// T6 plan-audit -> plan after an audit FAIL both do.
func TestStageCheckpointSameStageReEntry(t *testing.T) {
	ctx := context.Background()
	db := frOpen(t)
	repo := frNewRepo(t, true)
	frRegisterWorker(t, db, "worker-1")
	wp, spec := repo.Dir, frSpecID
	t0 := frNow
	if _, err := db.RecordPicked(ctx, frRun, "reenter", CardFields{WorktreePath: &wp, SpecID: &spec}, "op", t0); err != nil {
		t.Fatal(err)
	}
	t1 := t0.Add(10 * time.Minute)
	c := caaStep(t, db, TransitionRequest{RunID: frRun, CardID: "reenter", To: CardAssigned, ExpectedVersion: 1, Actor: "op", Owner: "worker-1"}, t1)
	t2 := t0.Add(20 * time.Minute)
	c = caaStep(t, db, TransitionRequest{RunID: frRun, CardID: "reenter", To: CardLeased, ExpectedVersion: c.Version, Actor: "worker-1"}, t2)

	ages := func(card Card) CardAges {
		events, err := db.LoadCardEvents(ctx, "")
		if err != nil {
			t.Fatal(err)
		}
		return ComputeCardAges(card, caaEventsOf(t, events, frRun, "reenter"), caaNoCommit, t0.Add(40*time.Hour))
	}
	now := t0.Add(40 * time.Hour)

	// First pipeline entry (T4a into plan) is a stage change. The walk times
	// stay inside the 15m lease T3 took at t2.
	t3 := t0.Add(22 * time.Minute)
	c = caaStep(t, db, TransitionRequest{RunID: frRun, CardID: "reenter", To: CardPlan, ExpectedVersion: c.Version, Actor: "worker-1"}, t3)
	if got := ages(c); !got.Stage.Known || got.Stage.Age != now.Sub(t3) {
		t.Fatalf("stage checkpoint after the first T4a = %+v, want %s (a first entry is a stage change)", got.Stage, now.Sub(t3))
	}

	// A forward advance (T5 plan -> plan-audit) refreshes the checkpoint.
	t4 := t0.Add(24 * time.Minute)
	req := frFullRequest(repo, c, CardPlanAudit)
	req.ExpectedVersion = c.Version
	c = caaStep(t, db, req, t4)
	if got := ages(c); !got.Stage.Known || got.Stage.Age != now.Sub(t4) {
		t.Fatalf("stage checkpoint after the forward T5 = %+v, want %s", got.Stage, now.Sub(t4))
	}

	// The backward T6 (plan-audit -> plan after an audit FAIL) refreshes it too.
	frWriteVerdict(t, repo.Dir, "reenter", "plan-audit.md", "FAIL", repo.Commit)
	t5 := t0.Add(26 * time.Minute)
	req = frFullRequest(repo, c, CardPlan)
	req.ExpectedVersion = c.Version
	c = caaStep(t, db, req, t5)
	if got := ages(c); !got.Stage.Known || got.Stage.Age != now.Sub(t5) {
		t.Fatalf("stage checkpoint after the backward T6 = %+v, want %s (backward moves refresh)", got.Stage, now.Sub(t5))
	}

	// The expiry -> re-acquire -> T4 same-stage re-entry cycle does NOT. The
	// lease taken at T3 (t2) already reads expired by t6; the stamp below
	// makes that explicit rather than relying on the arithmetic.
	if _, err := db.DB.Exec(`UPDATE cards SET lease_expires_at=? WHERE run_id=? AND card_id=?`,
		t0.Add(51*time.Minute).Format(time.RFC3339Nano), frRun, "reenter"); err != nil {
		t.Fatal(err)
	}
	t6 := t0.Add(time.Hour)
	c = caaExpire(t, db, frRun, "reenter", c.Version, t6)
	if c.State != CardAssigned || c.Stage != CardPlan {
		t.Fatalf("after expiry: state=%s stage=%s, want assigned with the stage kept", c.State, c.Stage)
	}
	t7 := t6.Add(time.Minute)
	c = caaStep(t, db, TransitionRequest{RunID: frRun, CardID: "reenter", To: CardLeased, ExpectedVersion: c.Version, Actor: "worker-1"}, t7)
	t8 := t6.Add(2 * time.Minute)
	c = caaStep(t, db, TransitionRequest{RunID: frRun, CardID: "reenter", To: CardPlan, ExpectedVersion: c.Version, Actor: "worker-1"}, t8)
	if c.State != CardPlan {
		t.Fatalf("re-entry landed %s, want plan", c.State)
	}
	got := ages(c)
	if !got.Stage.Known || got.Stage.Age != now.Sub(t5) {
		t.Fatalf("stage checkpoint after the same-stage re-entry = %+v, want still %s — re-entry is not a stage change", got.Stage, now.Sub(t5))
	}
	if !got.NoProgress.Known || got.NoProgress.Age != now.Sub(t5) {
		t.Fatalf("no-progress age after the same-stage re-entry = %+v, want still %s — the crash-loop pattern must stay visible", got.NoProgress, now.Sub(t5))
	}
	caaAssertNoNegativeAge(t, got)
}

// TestAgeAlarmCoverage covers REQ-CAA-008: the class (a)/(b) partition, with
// terminal states and legacy rows excluded from both.
func TestAgeAlarmCoverage(t *testing.T) {
	gate := map[string]bool{CardKickoff: true, CardNeedsDecision: true, CardBlocked: true}
	noProgress := map[string]bool{
		CardPicked: true, CardAssigned: true, CardLeased: true,
		CardPlan: true, CardPlanAudit: true, CardRun: true, CardSync: true, CardSyncAudit: true,
		CardMergeReady: true, CardMerging: true,
		CardMergedLocal: true, CardPushed: true, CardCIGreen: true, CardPROpen: true, CardMergedPR: true,
	}
	terminal := map[string]bool{CardDone: true, CardFailed: true, CardAbandoned: true}
	covered := 0
	for _, s := range CardStates() {
		card := Card{RunID: frRun, CardID: "cov", State: s}
		got := AgeClassOf(card)
		covered++
		switch {
		case gate[s]:
			if got != AgeClassGateWait {
				t.Fatalf("state %s classifies %q, want gate-wait", s, got)
			}
		case noProgress[s]:
			if got != AgeClassNoProgress {
				t.Fatalf("state %s classifies %q, want no-progress", s, got)
			}
		case terminal[s]:
			if got != "" {
				t.Fatalf("terminal state %s classifies %q, want no class", s, got)
			}
		default:
			t.Fatalf("fixture does not place state %s in a partition", s)
		}
	}
	if covered != len(CardStates()) || len(gate)+len(noProgress)+len(terminal) != len(CardStates()) {
		t.Fatal("the fixture partition does not cover the 21 states exactly once")
	}
	if got := AgeClassOf(Card{RunID: frRun, CardID: "cov", State: "pre-f1"}); got != "" {
		t.Fatalf("legacy row classifies %q, want no class", got)
	}
}
