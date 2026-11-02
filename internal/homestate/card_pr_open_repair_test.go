package homestate

import (
	"context"
	"errors"
	"testing"
)

// Card t2013 (R200): a card whose pull request is open (`pr-open`) can be
// taken back by its own registered owner to repair it. The edge is not a row of
// the requested-transition table (AC-005 pins that table); like the expired
// owner re-lease it is one record transaction: the card returns to the repair
// stage and the owner takes a lease again.

func prOpenRepairCard(t *testing.T, db *FactoryDB, owner string) Card {
	t.Helper()
	repo := frNewRepo(t, true)
	c := Card{
		RunID: frRun, CardID: "pr-card", State: CardPROpen, Version: 4, OwnerLabel: owner,
		WorktreePath: repo.Dir, Stage: CardMergeReady, SpecID: frSpecID, EvidenceSHA: repo.Commit,
	}
	frPlace(t, db, c)
	return c
}

func TestResumePROpenRepairLeasesTheOwnersCard(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	frRegisterWorker(t, db, "worker-1")
	before := prOpenRepairCard(t, db, "worker-1")

	got, ok, err := db.ResumePROpenRepair(ctx, frRun, "pr-card", "worker-1", frNow, nil)
	if err != nil || !ok {
		t.Fatalf("resume = (%v, %v), want the owner re-leased", ok, err)
	}
	after, err := db.LoadCard(ctx, frRun, "pr-card")
	if err != nil {
		t.Fatal(err)
	}
	if after.Version != got.Version || after.Version != before.Version+2 {
		t.Errorf("version = %d (returned %d), want %d", after.Version, got.Version, before.Version+2)
	}
	if after.State != CardLeased || after.LeaseHolder != "worker-1" || after.OwnerLabel != "worker-1" {
		t.Errorf("row = %s holder=%q owner=%q, want leased to worker-1", after.State, after.LeaseHolder, after.OwnerLabel)
	}
	if after.Stage != CardRun {
		t.Errorf("stage = %q, want %q (repair resumes in run)", after.Stage, CardRun)
	}
	if after.WorktreePath != before.WorktreePath || after.EvidenceSHA != before.EvidenceSHA {
		t.Errorf("worktree/evidence not kept: %q %q → %q %q", before.WorktreePath, before.EvidenceSHA, after.WorktreePath, after.EvidenceSHA)
	}
	kinds := reacqEvents(t, db)
	want := []string{"pr.repair_resumed", "card.transition"}
	if len(kinds) != len(want) || kinds[0] != want[0] || kinds[1] != want[1] {
		t.Fatalf("events = %v, want %v", kinds, want)
	}
}

func TestResumePROpenRepairRollsBackWholly(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	prOpenRepairCard(t, db, "worker-1") // worker-1 is NOT registered
	rowBefore := frRowDump(t, db, frRun, "pr-card")

	_, ok, err := db.ResumePROpenRepair(ctx, frRun, "pr-card", "worker-1", frNow, nil)
	if !errors.Is(err, ErrLeaseHolder) || ok {
		t.Fatalf("resume by an unregistered owner = (%v, %v), want ErrLeaseHolder", ok, err)
	}
	if after := frRowDump(t, db, frRun, "pr-card"); after != rowBefore {
		t.Errorf("a failed resume left a partial write:\nbefore:\n%s\nafter:\n%s", rowBefore, after)
	}
}

func TestResumePROpenRepairIgnoresEveryOtherShape(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name  string
		label string
		state string
	}{
		{"another lane", "worker-2", CardPROpen},
		{"empty label", "", CardPROpen},
		{"merged by pull request", "worker-1", CardMergedPR},
		{"already assigned", "worker-1", CardAssigned},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := frOpen(t)
			frRegisterWorker(t, db, "worker-1")
			frRegisterWorker(t, db, "worker-2")
			prOpenRepairCard(t, db, "worker-1")
			if _, err := db.DB.Exec(`UPDATE cards SET state=? WHERE card_id=?`, tc.state, "pr-card"); err != nil {
				t.Fatal(err)
			}
			before := frRowDump(t, db, frRun, "pr-card")
			_, ok, err := db.ResumePROpenRepair(ctx, frRun, "pr-card", tc.label, frNow, nil)
			if ok || err != nil {
				t.Fatalf("resume = (%v, %v), want (false, nil)", ok, err)
			}
			if after := frRowDump(t, db, frRun, "pr-card"); after != before {
				t.Errorf("a refused resume changed the record:\nbefore:\n%s\nafter:\n%s", before, after)
			}
		})
	}
}
