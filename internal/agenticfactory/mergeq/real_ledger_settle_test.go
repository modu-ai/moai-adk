package mergeq

// Defect answer: a fake queue accepts any write order, so it cannot tell a settlement the real ledger refuses.
// These tests settle rows against the real ledger: a candidate already in the trunk, a row left landing by a
// dead lander, an older landing row whose card was claimed again, and a lander whose row was restarted with
// another candidate.

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func realMerger(t *testing.T, dir string, s *ledger.Store) *Merger {
	t.Helper()
	g := GitCLI{Repo: dir, ScratchD: t.TempDir(), Who: Identity{Name: "mergeq", Email: "mergeq@example.test"}}
	return &Merger{
		Git: g, Queue: NewLedgerQueue(s), Trunk: "refs/heads/trunk", Verify: passing,
		Holder:  Holder{Session: "s1", PID: os.Getpid(), PIDStart: "start-s1"},
		LockTTL: DefaultLockTTL, MaxRetries: DefaultMaxRetries,
		Message: func(card string) string { return "merge(" + card + "): land" },
	}
}

func itemsOf(t *testing.T, s *ledger.Store) []Item {
	t.Helper()
	rows, err := s.Queue(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var out []Item
	for _, r := range rows {
		out = append(out, Item{CardID: r.CardID, CandidateSHA: r.CandidateSHA, State: State(r.State)})
	}
	return out
}

// leaveLanding takes the row to landing the way a lander that then died would, and lets the lock go.
func leaveLanding(t *testing.T, s *ledger.Store, cand string) (cardEpoch int64) {
	t.Helper()
	ctx := context.Background()
	rows, err := s.Queue(ctx)
	if err != nil || len(rows) != 1 {
		t.Fatalf("queue: %+v %v", rows, err)
	}
	cardEpoch = rows[0].CardEpoch
	lock, err := s.AcquireLock(ctx, ledger.LockMerge, ledger.LockHolder{SessionID: "dead", PID: os.Getpid(), PIDStart: "start-dead"}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(ctx, "t1", ledger.CardEpoch(cardEpoch, cand), ledger.QueueVerifying, "t0", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(ctx, "t1", ledger.LockEpoch(lock, cand), ledger.QueueLanding, "t0", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseLock(ctx, ledger.LockMerge, lock); err != nil {
		t.Fatal(err)
	}
	return cardEpoch
}

// A candidate that is already in the trunk settles the whole way: the ledger moves a row to landed only from
// landing, so the card passes verifying and landing under the lock first.
func TestLand_RealLedger_ACandidateAlreadyInTheTrunkSettlesAsLanded(t *testing.T) {
	dir := repo(t)
	cand := advanceTrunk(t, dir, "a.txt", "a\n") // the candidate commit is now the trunk tip
	s := queuedCard(t, "t1", cand)
	res, err := realMerger(t, dir, s).Land(context.Background(), "t1", cand)
	if err != nil || res.Outcome != OutcomeLanded {
		t.Fatalf("result %+v err %v, want landed", res, err)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueLanded {
		t.Errorf("the row is %s, want landed", got)
	}
}

// The whole path through the mergeq call: Land moves the trunk and settles the row, after which the card's holder
// moves the card to DONE (the ledger shows it the merged, closed and settled facts) and cannot queue a second
// candidate for it.
func TestLand_RealLedger_ALandedCardGoesToDoneAndTakesNoSecondCandidate(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	rows, err := s.Queue(ctx)
	if err != nil || len(rows) != 1 {
		t.Fatalf("queue: %+v %v", rows, err)
	}
	cardEpoch := rows[0].CardEpoch

	res, err := realMerger(t, dir, s).Land(ctx, "t1", cand)
	if err != nil || res.Outcome != OutcomeLanded {
		t.Fatalf("result %+v err %v, want landed", res, err)
	}
	// While the holder still has its lease and a passing gate for a second candidate, the landed row refuses it
	// for the row's own reason (ErrInvalid: already queued, landed), not for a missing lease or missing evidence.
	if err := s.AddEvidence(ctx, "t1", cardEpoch, ledger.Evidence{Kind: ledger.KindGate, TreeSHA: "tree-c2", SubjectSHA: "c2", Command: "go test ./...", Result: ledger.ResultPass}); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(ctx, "t1", cardEpoch, "c2"); !errors.Is(err, ledger.ErrInvalid) {
		t.Fatalf("a second candidate for a landed card: %v, want ErrInvalid from the landed row", err)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueLanded {
		t.Fatalf("the row is %s after the refused second candidate, want landed", got)
	}
	if err := s.Move(ctx, "t1", cardEpoch, core.Position{Stage: core.StageMerging, RemoteIntegrated: true}, "mergeq", "merge seen"); err != nil {
		t.Fatalf("recording the integration: %v", err)
	}
	if err := s.Move(ctx, "t1", cardEpoch, core.Position{Stage: core.StageDone}, "mergeq", "landed"); err != nil {
		t.Fatalf("moving the landed card to DONE: %v", err)
	}
	cards, err := s.Cards(ctx)
	if err != nil || len(cards) != 1 || cards[0].Position.Stage != core.StageDone {
		t.Fatalf("cards %+v %v, want the card DONE", cards, err)
	}
}

// A retry of Land for a row that is already landed answers landed instead of failing on a landed -> landing
// write the table refuses (it can follow a landed result that came back with a lock-release error).
func TestLand_RealLedger_ARowAlreadyLandedAnswersLanded(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	m := realMerger(t, dir, s)
	if res, err := m.Land(ctx, "t1", cand); err != nil || res.Outcome != OutcomeLanded {
		t.Fatalf("first Land: %+v %v", res, err)
	}
	res, err := m.Land(ctx, "t1", cand)
	if err != nil || res.Outcome != OutcomeLanded {
		t.Fatalf("second Land: %+v %v, want landed", res, err)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueLanded {
		t.Errorf("the row is %s, want landed", got)
	}
}

// After a restart a row left in landing is judged by git ancestry: a candidate in the trunk is landed, one that
// is not goes back to waiting.
func TestReconcile_RealLedger_ALandingRowIsSettledByAncestry(t *testing.T) {
	for _, inTrunk := range []bool{true, false} {
		dir := repo(t)
		var cand string
		if inTrunk {
			cand = advanceTrunk(t, dir, "a.txt", "a\n")
		} else {
			cand = candidate(t, dir, "c1", "a.txt", "a\n")
		}
		s := queuedCard(t, "t1", cand)
		leaveLanding(t, s, cand)
		acts, err := realMerger(t, dir, s).Reconcile(context.Background(), itemsOf(t, s))
		if err != nil || len(acts) != 1 {
			t.Fatalf("inTrunk=%v: actions %+v err %v", inTrunk, acts, err)
		}
		want := ledger.QueueWaiting
		if inTrunk {
			want = ledger.QueueLanded
		}
		if got := rowState(t, s, "t1"); got != want {
			t.Errorf("inTrunk=%v: the row is %s, want %s", inTrunk, got, want)
		}
	}
}

// A landing row of an older card epoch (the card was claimed again) is not sent back to waiting: the ledger lets
// only the lock holder settle it. A candidate not in the trunk settles as failed; one in the trunk as landed.
func TestReconcile_RealLedger_AnOlderLandingRowIsSettledNotRestarted(t *testing.T) {
	ctx := context.Background()
	for _, inTrunk := range []bool{false, true} {
		dir := repo(t)
		var cand string
		if inTrunk {
			cand = advanceTrunk(t, dir, "a.txt", "a\n")
		} else {
			cand = candidate(t, dir, "c1", "a.txt", "a\n")
		}
		s := queuedCard(t, "t1", cand)
		oldEpoch := leaveLanding(t, s, cand)
		// the card's holder backs out and the card is claimed again at the next epoch
		if err := s.Move(ctx, "t1", oldEpoch, core.Position{Stage: core.StageReviewing}, "lane", "back out of the queue"); err != nil {
			t.Fatal(err)
		}
		if err := s.Release(ctx, "t1", oldEpoch); err != nil {
			t.Fatal(err)
		}
		c2, err := s.Claim(ctx, ledger.ClaimRequest{SessionID: "next", Worktree: "/wt/next", PID: os.Getpid(), PIDStart: "start-next"})
		if err != nil || c2.Epoch == oldEpoch {
			t.Fatalf("claimed again: %+v %v", c2, err)
		}

		acts, err := realMerger(t, dir, s).Reconcile(ctx, itemsOf(t, s))
		if err != nil || len(acts) != 1 {
			t.Fatalf("inTrunk=%v: actions %+v err %v", inTrunk, acts, err)
		}
		want := ledger.QueueFailed
		if inTrunk {
			want = ledger.QueueLanded
		}
		if got := rowState(t, s, "t1"); got != want {
			t.Errorf("inTrunk=%v: the row is %s, want %s", inTrunk, got, want)
		}
	}
}

// A lander whose row was restarted with another candidate cannot move it: the write names the candidate the
// lander verified, and the ledger refuses a different one.
func TestLedgerQueue_RealLedger_AnOldLanderCannotMoveARestartedRow(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	err := NewLedgerQueue(s).SetState(context.Background(), "t1", "an-older-candidate", 0, StateVerifying, "", "")
	if !errors.Is(err, ErrStaleRow) {
		t.Fatalf("a write naming another candidate: %v, want ErrStaleRow", err)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueWaiting {
		t.Errorf("the refused write moved the row to %s", got)
	}
}
