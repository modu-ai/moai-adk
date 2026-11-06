package homestate

import (
	"context"
	"testing"
	"time"
)

// Card t2116: the owner's re-lease of an expired lease acts on every
// lease-holding state a lane parks a card in, not only on `leased`. An expiry
// return (T27) sends the card to `assigned` with its stage, worktree and
// evidence kept; the existing lease edge (T3) then returns it to `leased`, and
// the lane resumes the kept stage. `merging` stays out: its expiry means an
// interrupted merge (blocked), which no lane may quietly re-lease.

// stagedReacquirableStates are the lease-holding states other than `leased` and
// `merging`.
func stagedReacquirableStates() []string {
	var out []string
	for _, s := range leaseHoldingStates {
		if s != CardLeased && s != CardMerging {
			out = append(out, s)
		}
	}
	return out
}

// stagedCard places a card in state with its stage set to the same name (the
// resumable stages are named like their states) and a lease that lapsed
// expiredFor ago.
func stagedCard(t *testing.T, db *FactoryDB, state, label string, expiredFor time.Duration) Card {
	t.Helper()
	c := reacqCard(t, db, label, expiredFor)
	if _, err := db.DB.Exec(`UPDATE cards SET state=?, stage=? WHERE card_id=?`, state, state, c.CardID); err != nil {
		t.Fatal(err)
	}
	c.State, c.Stage = state, state
	return c
}

// The pure predicate: exactly the owner-and-holder of an expired lease in a
// lease-holding state other than `merging`.
func TestIsReacquirableExpiredLeaseCoversStagedStates_T2116(t *testing.T) {
	now := frNow
	expired := now.Add(-2 * time.Minute).Format(time.RFC3339Nano)
	valid := now.Add(5 * time.Minute).Format(time.RFC3339Nano)
	mk := func(state, owner, holder, expires string) Card {
		return Card{CardID: "c", State: state, OwnerLabel: owner, LeaseHolder: holder, LeaseExpiresAt: expires}
	}
	for _, state := range append([]string{CardLeased}, stagedReacquirableStates()...) {
		if !IsReacquirableExpiredLease(mk(state, "lane-1", "lane-1", expired), "lane-1", now) {
			t.Errorf("state %s: owner with an expired lease is not reacquirable, want reacquirable", state)
		}
		if IsReacquirableExpiredLease(mk(state, "lane-1", "lane-1", valid), "lane-1", now) {
			t.Errorf("state %s: an unexpired lease is reacquirable, want not", state)
		}
		if IsReacquirableExpiredLease(mk(state, "lane-1", "lane-1", expired), "lane-2", now) {
			t.Errorf("state %s: another lane may reacquire, want owner only", state)
		}
		if IsReacquirableExpiredLease(mk(state, "lane-1", "lane-2", expired), "lane-1", now) {
			t.Errorf("state %s: a card held by another lane is reacquirable by the owner, want not", state)
		}
	}
	for _, state := range []string{CardMerging, CardPicked, CardAssigned, CardBlocked, CardDone, CardMergedLocal, CardPROpen, CardNeedsDecision, CardFailed, CardAbandoned} {
		if IsReacquirableExpiredLease(mk(state, "lane-1", "lane-1", expired), "lane-1", now) {
			t.Errorf("state %s is reacquirable, want not (not a state a lane parks in, or an interrupted merge)", state)
		}
	}
}

// The transaction: a card parked in a staged state comes back `leased` to its
// owner with stage, worktree and evidence kept, and the kept stage resumes.
func TestReacquireExpiredLeaseStagedStates_T2116(t *testing.T) {
	ctx := context.Background()
	for _, state := range stagedReacquirableStates() {
		t.Run(state, func(t *testing.T) {
			db := frOpen(t)
			frRegisterWorker(t, db, "worker-1")
			before := stagedCard(t, db, state, "worker-1", 2*time.Minute)

			got, ok, err := db.ReacquireExpiredLease(ctx, frRun, "stale", "worker-1", frNow, nil)
			if err != nil || !ok {
				t.Fatalf("reacquire of an expired %s card = (%v, %v), want a re-lease", state, ok, err)
			}
			after, err := db.LoadCard(ctx, frRun, "stale")
			if err != nil {
				t.Fatal(err)
			}
			if after.State != CardLeased || after.LeaseHolder != "worker-1" || after.OwnerLabel != "worker-1" || got.Version != after.Version {
				t.Fatalf("row = %s holder=%q owner=%q v%d (returned v%d), want leased to worker-1", after.State, after.LeaseHolder, after.OwnerLabel, after.Version, got.Version)
			}
			if after.Stage != state || after.WorktreePath != before.WorktreePath || after.EvidenceSHA != before.EvidenceSHA {
				t.Errorf("stage/worktree/evidence not kept: %q %q %q → %q %q %q",
					before.Stage, before.WorktreePath, before.EvidenceSHA, after.Stage, after.WorktreePath, after.EvidenceSHA)
			}
			kinds := reacqEvents(t, db)
			want := []string{"lease.expired", "card.transition", "lease.reacquired"}
			if len(kinds) != len(want) {
				t.Fatalf("events = %v, want %v", kinds, want)
			}
			for i := range want {
				if kinds[i] != want[i] {
					t.Fatalf("events = %v, want %v", kinds, want)
				}
			}
			// The kept stage resumes: the lane steps back into the stage it was in.
			resumed, err := db.Transition(ctx, TransitionRequest{
				RunID: frRun, CardID: "stale", To: state, ExpectedVersion: after.Version, Actor: "worker-1", Now: frNow,
			})
			if err != nil {
				t.Fatalf("resuming stage %s after the re-lease: %v", state, err)
			}
			if resumed.State != state {
				t.Errorf("after resume state = %s, want %s", resumed.State, state)
			}
		})
	}
}

// Not the owner, an unexpired lease, or an interrupted merge: nothing changes,
// no error, reported false so the caller keeps its refusal.
func TestReacquireExpiredLeaseStagedIgnoresOtherShapes_T2116(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name  string
		label string
		build func(t *testing.T, db *FactoryDB)
	}{
		{"another lane, run", "worker-2", func(t *testing.T, db *FactoryDB) { stagedCard(t, db, CardRun, "worker-1", 2*time.Minute) }},
		{"another lane, merge-ready", "worker-2", func(t *testing.T, db *FactoryDB) { stagedCard(t, db, CardMergeReady, "worker-1", 2*time.Minute) }},
		{"lease still valid, sync", "worker-1", func(t *testing.T, db *FactoryDB) { stagedCard(t, db, CardSync, "worker-1", -time.Minute) }},
		{"interrupted merge", "worker-1", func(t *testing.T, db *FactoryDB) { stagedCard(t, db, CardMerging, "worker-1", 2*time.Minute) }},
		{"blocked", "worker-1", func(t *testing.T, db *FactoryDB) { stagedCard(t, db, CardBlocked, "worker-1", 2*time.Minute) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := frOpen(t)
			frRegisterWorker(t, db, "worker-1")
			frRegisterWorker(t, db, "worker-2")
			tc.build(t, db)
			before := frRowDump(t, db, frRun, "stale")
			_, ok, err := db.ReacquireExpiredLease(ctx, frRun, "stale", tc.label, frNow, nil)
			if ok || err != nil {
				t.Fatalf("reacquire = (%v, %v), want (false, nil)", ok, err)
			}
			if after := frRowDump(t, db, frRun, "stale"); after != before {
				t.Errorf("an ignored shape changed the record:\nbefore:\n%s\nafter:\n%s", before, after)
			}
		})
	}
}
