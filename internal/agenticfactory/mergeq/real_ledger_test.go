package mergeq

// Defect answer: the fake queue accepts any state write, so it cannot tell a Land that writes a state the real
// ledger refuses. These tests run Land against the real ledger: a lock held by another lander must end as
// "waiting" without an error and without moving the queue row out of verifying, and the same card must land once
// the lock is free.

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// queuedCard opens a real ledger, puts a claimed card with passing gate evidence for sha on the merge queue and
// returns the store.
func queuedCard(t *testing.T, id, sha string) *ledger.Store {
	t.Helper()
	ctx := context.Background()
	s, err := ledger.Open(ctx, filepath.Join(t.TempDir(), "db", "proj", "af", "ledger.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if err := s.AddCard(ctx, ledger.Card{ID: id, Title: id, Position: core.Position{Stage: core.StageReady}, Priority: 1,
		Spec: ledger.Spec{Class: "A", Paths: []string{"a.txt"}}}); err != nil {
		t.Fatal(err)
	}
	c, err := s.Claim(ctx, ledger.ClaimRequest{SessionID: "s-" + id, Worktree: "/wt/" + id, PID: os.Getpid(), PIDStart: "start-" + id})
	if err != nil || c.CardID != id {
		t.Fatalf("claim: %+v %v", c, err)
	}
	for _, e := range []ledger.Evidence{
		{Kind: ledger.KindGreen, TreeSHA: "tree-" + sha, SubjectSHA: sha, Result: ledger.ResultPass},
		{Kind: ledger.KindGate, TreeSHA: "tree-" + sha, SubjectSHA: sha, Command: "go test ./...", Result: ledger.ResultPass},
	} {
		if err := s.AddEvidence(ctx, id, c.Epoch, e); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Enqueue(ctx, id, c.Epoch, sha); err != nil {
		t.Fatal(err)
	}
	return s
}

func rowState(t *testing.T, s *ledger.Store, id string) ledger.QueueState {
	t.Helper()
	items, err := s.Queue(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		if it.CardID == id {
			return it.State
		}
	}
	t.Fatalf("%s is not in the queue", id)
	return ""
}

func TestLand_RealLedger_ALockHeldByAnotherLanderIsWaitingThenTheCardLands(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	before := git(t, dir, "rev-parse", "refs/heads/trunk")
	s := queuedCard(t, "t1", cand)

	other, err := s.AcquireLock(ctx, ledger.LockMerge, ledger.LockHolder{SessionID: "other", PID: os.Getpid(), PIDStart: "start-other"}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	g := GitCLI{Repo: dir, ScratchD: t.TempDir(), Who: Identity{Name: "mergeq", Email: "mergeq@example.test"}}
	m := &Merger{
		Git: g, Queue: NewLedgerQueue(s), Trunk: "refs/heads/trunk", Verify: passing,
		Holder:  Holder{Session: "s1", PID: os.Getpid(), PIDStart: "start-s1"},
		LockTTL: DefaultLockTTL, MaxRetries: DefaultMaxRetries,
		Message: func(card string) string { return "merge(" + card + "): land" },
	}

	res, err := m.Land(ctx, "t1", cand)
	if err != nil {
		t.Fatalf("a held landing lock must end as waiting, not an error: %v", err)
	}
	if res.Outcome != OutcomeWaiting {
		t.Fatalf("result %+v, want waiting", res)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueVerifying {
		t.Errorf("the row is %s, want verifying (the ledger lets only a lock holder send it back to waiting)", got)
	}
	if git(t, dir, "rev-parse", "refs/heads/trunk") != before {
		t.Error("the trunk moved without the lock")
	}

	if err := s.ReleaseLock(ctx, ledger.LockMerge, other); err != nil {
		t.Fatal(err)
	}
	res, err = m.Land(ctx, "t1", cand)
	if err != nil {
		t.Fatalf("the second Land: %v", err)
	}
	if res.Outcome != OutcomeLanded {
		t.Fatalf("result %+v, want landed once the lock is free", res)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueLanded {
		t.Errorf("the row is %s, want landed", got)
	}
	if git(t, dir, "rev-parse", "refs/heads/trunk") == before {
		t.Error("the trunk did not move")
	}
}
