package homestate

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Card t2013: ReacquireExpiredLease applies the expiry return and the new lease
// in one transaction, for the registered owner of an expired `leased` card only.

func reacqCard(t *testing.T, db *FactoryDB, label string, expiredFor time.Duration) Card {
	t.Helper()
	repo := frNewRepo(t, true)
	c := frLeasedCard(repo, "stale", CardLeased)
	c.Version = 3
	c.Stage = CardRun
	c.EvidenceSHA = repo.Commit
	c.OwnerLabel, c.LeaseHolder = label, label
	c.LeaseExpiresAt = frNow.Add(-expiredFor).Format(time.RFC3339Nano)
	frPlace(t, db, c)
	return c
}

func reacqEvents(t *testing.T, db *FactoryDB) []string {
	t.Helper()
	rows, err := db.DB.Query(`SELECT kind FROM events ORDER BY seq`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var kinds []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			t.Fatal(err)
		}
		kinds = append(kinds, k)
	}
	return kinds
}

func TestReacquireExpiredLeaseIsOneTransaction(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	frRegisterWorker(t, db, "worker-1")
	before := reacqCard(t, db, "worker-1", 2*time.Minute)

	hooked := false
	got, ok, err := db.ReacquireExpiredLease(ctx, frRun, "stale", "worker-1", frNow, func(c Card) error {
		hooked = c.State == CardLeased && c.Version == before.Version+2
		return nil
	})
	if err != nil || !ok {
		t.Fatalf("reacquire = (%v, %v), want a re-lease", ok, err)
	}
	if !hooked {
		t.Errorf("beforeCommit did not see the re-leased card at version %d", before.Version+2)
	}
	after, err := db.LoadCard(ctx, frRun, "stale")
	if err != nil {
		t.Fatal(err)
	}
	if after.Version != got.Version || after.Version != before.Version+2 {
		t.Errorf("version = %d (returned %d), want %d (expiry bump + lease bump)", after.Version, got.Version, before.Version+2)
	}
	if after.State != CardLeased || after.LeaseHolder != "worker-1" || after.OwnerLabel != "worker-1" {
		t.Errorf("row = %s holder=%q owner=%q, want leased to worker-1", after.State, after.LeaseHolder, after.OwnerLabel)
	}
	if want := frNow.Add(db.CardLeaseDuration()).Format(time.RFC3339Nano); after.LeaseExpiresAt != want {
		t.Errorf("lease expiry = %q, want %q", after.LeaseExpiresAt, want)
	}
	if after.Stage != before.Stage || after.WorktreePath != before.WorktreePath || after.EvidenceSHA != before.EvidenceSHA {
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
}

// A failure of the lease step rolls the expiry back with it: an owner that is
// not a registered worker leaves the card exactly as it was.
func TestReacquireExpiredLeaseRollsBackWholly(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	reacqCard(t, db, "worker-1", 2*time.Minute) // worker-1 is NOT registered
	rowBefore := frRowDump(t, db, frRun, "stale")

	_, ok, err := db.ReacquireExpiredLease(ctx, frRun, "stale", "worker-1", frNow, nil)
	if !errors.Is(err, ErrLeaseHolder) || ok {
		t.Fatalf("reacquire by an unregistered owner = (%v, %v), want ErrLeaseHolder", ok, err)
	}
	if after := frRowDump(t, db, frRun, "stale"); after != rowBefore {
		t.Errorf("a failed re-lease left a partial write:\nbefore:\n%s\nafter:\n%s", rowBefore, after)
	}
}

// Not the owner, an unexpired lease, or a card that is not `leased`: no change,
// no error, reported false so the caller keeps its refusal.
func TestReacquireExpiredLeaseIgnoresEveryOtherShape(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name  string
		label string
		build func(t *testing.T, db *FactoryDB)
	}{
		{"another lane", "worker-2", func(t *testing.T, db *FactoryDB) { reacqCard(t, db, "worker-1", 2*time.Minute) }},
		{"holder but not owner", "worker-1", func(t *testing.T, db *FactoryDB) {
			c := reacqCard(t, db, "worker-1", 2*time.Minute)
			if _, err := db.DB.Exec(`UPDATE cards SET owner_label='worker-2' WHERE card_id=?`, c.CardID); err != nil {
				t.Fatal(err)
			}
		}},
		{"lease still valid", "worker-1", func(t *testing.T, db *FactoryDB) { reacqCard(t, db, "worker-1", -time.Minute) }},
		{"empty label", "", func(t *testing.T, db *FactoryDB) { reacqCard(t, db, "worker-1", 2*time.Minute) }},
		// Card t2116: a staged state (run included) is now reacquirable by its
		// owner; the shape that stays ignored is an interrupted merge.
		{"interrupted merge", "worker-1", func(t *testing.T, db *FactoryDB) {
			c := reacqCard(t, db, "worker-1", 2*time.Minute)
			if _, err := db.DB.Exec(`UPDATE cards SET state=? WHERE card_id=?`, CardMerging, c.CardID); err != nil {
				t.Fatal(err)
			}
		}},
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
