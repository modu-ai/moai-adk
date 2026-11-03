package mergeq

// Defect answer: the final code audit (1438ac5552) reproduced three ways the merge queue can go wrong that the
// earlier tests could not see, because they ran against a fake queue or a plain branch. A trunk that is a symbolic
// ref to a checked-out branch passed every check and moved that branch (F1); a lost compare-and-swap left the
// row landing and the retry wrote verifying, which the real ledger refuses (F2); and Reconcile judged a verifying
// row by a lease read taken before it held the lock (F3). Each test runs on a real git repository, and F2 and F3
// on the real ledger.

import (
	"context"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// F1: refs/heads/symtrunk is a symbolic ref to refs/heads/work, which is checked out. The trunk must be refused
// and the checked-out branch must not move.
func TestLand_ASymbolicTrunkIsRefusedAndTheCheckedOutBranchStays(t *testing.T) {
	dir := repo(t) // HEAD is on "work"
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	workBefore := git(t, dir, "rev-parse", "refs/heads/work")
	git(t, dir, "symbolic-ref", "refs/heads/symtrunk", "refs/heads/work")

	m := newMerger(t, dir, &fakeQueue{}, passing)
	m.Trunk = "refs/heads/symtrunk"
	_, err := m.Land(context.Background(), "t1", cand)
	if !errors.Is(err, ErrBadTrunk) {
		t.Fatalf("a symbolic trunk: %v, want ErrBadTrunk", err)
	}
	if got := git(t, dir, "rev-parse", "refs/heads/work"); got != workBefore {
		t.Errorf("the checked-out branch moved from %s to %s through the symbolic trunk", workBefore, got)
	}
}

// F1, second layer: even if a symbolic ref reaches the swap, update-ref must not follow it.
func TestGitCLI_UpdateRefDoesNotFollowASymbolicRef(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	workBefore := git(t, dir, "rev-parse", "refs/heads/work")
	git(t, dir, "symbolic-ref", "refs/heads/symtrunk", "refs/heads/work")
	g := GitCLI{Repo: dir, ScratchD: t.TempDir(), Who: Identity{Name: "mergeq", Email: "mergeq@example.test"}}
	_ = g.UpdateRef(context.Background(), "refs/heads/symtrunk", cand, workBefore) // an error is fine; a move of work is not
	if got := git(t, dir, "rev-parse", "refs/heads/work"); got != workBefore {
		t.Errorf("update-ref followed the symbolic ref and moved work from %s to %s", workBefore, got)
	}
}

// F2: the compare-and-swap loses once. The row is landing under the lock; the retry must be able to verify again,
// so the loser sends the row back to waiting under the lock first. Run on the real ledger, which refuses
// landing -> verifying.
func TestLand_RealLedger_ALostCompareAndSwapIsRetriedAndLands(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	m := realMerger(t, dir, s)
	m.Git = &movingGit{Git: m.Git, t: t, dir: dir}

	res, err := m.Land(ctx, "t1", cand)
	if err != nil {
		t.Fatalf("a lost compare-and-swap ended as an error: %v", err)
	}
	if res.Outcome != OutcomeLanded {
		t.Fatalf("result %+v, want landed after the retry", res)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueLanded {
		t.Errorf("the row is %s, want landed", got)
	}
	if git(t, dir, "show", "refs/heads/trunk:late.txt") != "late" || git(t, dir, "show", "refs/heads/trunk:a.txt") != "a" {
		t.Error("the commit that landed in the window or the card's own is missing from the trunk")
	}
}

// F3: Reconcile must judge a verifying row by the lease it reads while it holds the landing lock. The items it was
// given carry no lease information (read before the lock); the card's lease is live, so the row stays.
func TestReconcile_RealLedger_TheLeaseIsReadAgainUnderTheLock(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	rows, _ := s.Queue(ctx)
	if err := s.QueueUpdate(ctx, "t1", ledger.CardEpoch(rows[0].CardEpoch, cand), ledger.QueueVerifying, "t0", ""); err != nil {
		t.Fatal(err)
	}
	stale := itemsOf(t, s) // zero lease fields: "no live lease", as a read taken earlier could have shown
	if stale[0].LeaseLive {
		t.Fatal("the fixture items must carry no live lease")
	}
	acts, err := realMerger(t, dir, s).Reconcile(ctx, stale)
	if err != nil || len(acts) != 0 {
		t.Fatalf("actions %+v err %v, want none: the lease is live under the lock", acts, err)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueVerifying {
		t.Errorf("the row is %s, want verifying", got)
	}
}
