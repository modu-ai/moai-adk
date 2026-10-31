package cli

// lane-19 35th review (4) and the settlement of a withdrawn card's landing row: reclaiming a dead holder's landing
// lock must not touch any card lease or write any note about one, and a landing row whose card left MERGING after
// the lander wrote landing is settled landed, not failed. Both run on a real ledger and a real git repository.

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestAFLand_ReclaimingALockLeavesOtherCardsLeasesAndNotesAlone(t *testing.T) {
	afSkipWithoutSh(t)
	ctx := context.Background()
	db, repo, _, st, _ := afQueuedCard(t)
	// another card whose lease has run out and whose holder is unknown (no pid): the full Reap writes a stalled
	// note for it, and the landing must not
	if err := st.AddCard(ctx, ledger.Card{ID: "t2", Title: "t2", Position: core.Position{Stage: core.StageReady}, Priority: 1,
		Spec: ledger.Spec{Class: "A", Paths: []string{"other/pkg/b.go"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Claim(ctx, ledger.ClaimRequest{SessionID: "s-t2", Worktree: "/wt/t2", Lease: time.Millisecond}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	before, _ := st.NotesSince(ctx, 0)

	if outcome, reason := afLandJSON(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/trunk", "--gate", "true"); outcome != "landed" {
		t.Fatalf("outcome %q reason %q", outcome, reason)
	}
	after, _ := st.NotesSince(ctx, 0)
	for _, n := range after[len(before):] {
		if n.Kind == ledger.NoteStalled || n.CardID == "t2" {
			t.Errorf("the landing wrote a note about another card's lease: %+v", n)
		}
	}
	leases, _ := st.Leases(ctx)
	found := false
	for _, l := range leases {
		found = found || l.CardID == "t2"
	}
	if !found {
		t.Error("the landing took the other card's lease back")
	}
}

func TestAFReconcile_ALandingRowOfAWithdrawnCardIsSettledLanded(t *testing.T) {
	ctx := context.Background()
	db, repo, cand, st, epoch := afQueuedCard(t)
	lock, err := st.AcquireLock(ctx, ledger.LockMerge, ledger.LockHolder{SessionID: "dead", PID: os.Getpid(), PIDStart: "start-dead"}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.QueueUpdate(ctx, "t1", ledger.CardEpoch(epoch, cand), ledger.QueueVerifying, "t0", ""); err != nil {
		t.Fatal(err)
	}
	if err := st.QueueUpdate(ctx, "t1", ledger.LockEpoch(lock, cand), ledger.QueueLanding, "t0", ""); err != nil {
		t.Fatal(err)
	}
	if err := st.ReleaseLock(ctx, ledger.LockMerge, lock); err != nil {
		t.Fatal(err)
	}
	afGit(t, repo, "update-ref", "refs/heads/trunk", cand) // the dead lander had moved the trunk
	// and the holder withdrew the card after the lander wrote landing
	if err := st.Move(ctx, "t1", epoch, core.Position{Stage: core.StageReviewing}, "lane", "withdrawn"); err != nil {
		t.Fatal(err)
	}

	out := runAF(t, "reconcile", "--db", db, "--repo", repo, "--trunk", "refs/heads/trunk")
	if got := afRowState(t, st); got != ledger.QueueLanded {
		t.Errorf("the row is %s after reconcile (output %q), want landed: the candidate is in the trunk", got, out)
	}
}
