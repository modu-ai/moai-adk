package homestate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"
)

func frWorkerHeartbeat(t *testing.T, db *FactoryDB, label string) string {
	t.Helper()
	var hb string
	if err := db.DB.QueryRow(`SELECT heartbeat_at FROM workers WHERE label=?`, label).Scan(&hb); err != nil {
		t.Fatalf("read heartbeat %s: %v", label, err)
	}
	return hb
}

func frWorkersDump(t *testing.T, db *FactoryDB) string {
	t.Helper()
	return strings.Join(frDump(t, db.DB, "workers", []string{"label", "pid", "registered_at", "heartbeat_at"}), "\n")
}

// frWorktreeHash digests HEAD, the branch list, and the porcelain status of a
// fixture worktree — the witness that expiry touched no file or ref.
func frWorktreeHash(t *testing.T, dir string) string {
	t.Helper()
	h := sha256.New()
	for _, args := range [][]string{{"rev-parse", "HEAD"}, {"branch", "--list"}, {"status", "--porcelain"}} {
		h.Write([]byte(frGit(t, dir, args...)))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// AC-010 — the lease goes only to the registered owner; renewal extends it
// and refreshes the roster heartbeat; anyone else's renewal changes nothing.
func TestFR_AC010_LeaseAcquireAndRenew(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	frRegisterWorker(t, db, "worker-1")
	frRegisterWorker(t, db, "worker-2")
	c := Card{RunID: frRun, CardID: "lease", State: CardAssigned, Version: 1, OwnerLabel: "worker-1", WorktreePath: repo.Dir}
	frPlace(t, db, c)
	before := frRowDump(t, db, frRun, "lease")
	for _, label := range []string{"ghost", "worker-2"} {
		_, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "lease", To: CardLeased, ExpectedVersion: 1, Actor: label, Now: frNow})
		if !errors.Is(err, ErrLeaseHolder) {
			t.Fatalf("lease by %s: err = %v, want ErrLeaseHolder", label, err)
		}
		if after := frRowDump(t, db, frRun, "lease"); after != before {
			t.Fatalf("refused lease by %s changed the record", label)
		}
	}
	got, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "lease", To: CardLeased, ExpectedVersion: 1, Actor: "worker-1", Now: frNow})
	if err != nil {
		t.Fatalf("lease by owner: %v", err)
	}
	if got.State != CardLeased || got.LeaseHolder != "worker-1" || got.LeaseExpiresAt != frNow.Add(FactoryLeaseDuration).Format(time.RFC3339Nano) {
		t.Fatalf("leased card = %s holder=%q expiry=%q", got.State, got.LeaseHolder, got.LeaseExpiresAt)
	}
	renewAt := frNow.Add(5 * time.Minute)
	renewed, err := db.RenewLease(ctx, frRun, "lease", "worker-1", renewAt)
	if err != nil {
		t.Fatalf("renew by holder: %v", err)
	}
	if renewed.LeaseExpiresAt != renewAt.Add(FactoryLeaseDuration).Format(time.RFC3339Nano) {
		t.Fatalf("renewed expiry = %q, want %q", renewed.LeaseExpiresAt, renewAt.Add(FactoryLeaseDuration).Format(time.RFC3339Nano))
	}
	if hb := frWorkerHeartbeat(t, db, "worker-1"); hb != renewAt.Format(time.RFC3339Nano) {
		t.Fatalf("workers.heartbeat_at(worker-1) = %q, want %q", hb, renewAt.Format(time.RFC3339Nano))
	}
	cardBefore, workersBefore := frRowDump(t, db, frRun, "lease"), frWorkersDump(t, db)
	if _, err := db.RenewLease(ctx, frRun, "lease", "worker-2", renewAt.Add(time.Minute)); !errors.Is(err, ErrLeaseHolder) {
		t.Fatalf("renew by worker-2: err = %v, want ErrLeaseHolder", err)
	}
	if frRowDump(t, db, frRun, "lease") != cardBefore || frWorkersDump(t, db) != workersBefore {
		t.Fatal("refused renewal changed the card or the roster")
	}
}

// AC-011 — an expired lease returns the card to assigned before anything
// else, preserving stage, worktree, and evidence, and touching no git state.
func TestFR_AC011_ExpiredLeaseReturnsToAssigned(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	frRegisterWorker(t, db, "worker-1")
	c := frLeasedCard(repo, "expired", CardRun)
	c.EvidenceSHA = repo.Commit
	c.LeaseExpiresAt = frNow.Add(-time.Minute).Format(time.RFC3339Nano)
	frPlace(t, db, c)
	treeBefore := frWorktreeHash(t, repo.Dir)
	req := frHolderRequest(c, CardSync)
	req.SHA = repo.Commit
	if _, err := db.Transition(ctx, req); !errors.Is(err, ErrLeaseExpired) {
		t.Fatalf("transition on expired lease: err = %v, want ErrLeaseExpired", err)
	}
	got, err := db.LoadCard(ctx, frRun, "expired")
	if err != nil {
		t.Fatal(err)
	}
	if got.State != CardAssigned || got.LeaseHolder != "" || got.Stage != CardRun || got.WorktreePath != repo.Dir || got.EvidenceSHA != repo.Commit {
		t.Fatalf("after expiry: state=%s holder=%q stage=%s worktree=%q sha=%q", got.State, got.LeaseHolder, got.Stage, got.WorktreePath, got.EvidenceSHA)
	}
	if n := len(frEvents(t, db, "lease.expired")); n != 1 {
		t.Fatalf("lease.expired events = %d, want 1", n)
	}
	if after := frWorktreeHash(t, repo.Dir); after != treeBefore {
		t.Fatal("lease expiry changed the fixture worktree")
	}
	leased, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "expired", To: CardLeased, ExpectedVersion: got.Version, Actor: "worker-1", Now: frNow})
	if err != nil {
		t.Fatalf("re-lease by owner: %v", err)
	}
	if _, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "expired", To: CardRun, ExpectedVersion: leased.Version, Actor: "worker-1", Now: frNow}); err != nil {
		t.Fatalf("leased → run after re-lease: %v", err)
	}
}

// AC-012 — an expired lease during merging blocks the card instead.
func TestFR_AC012_ExpiredMergingBlocks(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	c := frLeasedCard(repo, "merging", CardMerging)
	c.LeaseExpiresAt = frNow.Add(-time.Second).Format(time.RFC3339Nano)
	frPlace(t, db, c)
	if _, err := db.Transition(ctx, frHolderRequest(c, CardMergeReady)); !errors.Is(err, ErrLeaseExpired) {
		t.Fatalf("err = %v, want ErrLeaseExpired", err)
	}
	got, _ := db.LoadCard(ctx, frRun, "merging")
	if got.State != CardBlocked {
		t.Fatalf("state = %s, want blocked", got.State)
	}
	events := frEvents(t, db, "lease.expired")
	if len(events) != 1 || !strings.Contains(events[0], "merge") {
		t.Fatalf("lease.expired events = %v, want one naming the interrupted merge", events)
	}
}

// AC-013 (store half) — kickoff and needs-decision hold no lease, so a human
// who takes longer than the lease duration loses nothing.
func TestFR_AC013_DecisionPendingHoldsNoLease(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	c := frLeasedCard(repo, "kick", CardPlanAudit)
	c.EvidenceSHA = repo.Commit
	frPlace(t, db, c)
	frWriteVerdict(t, repo.Dir, "kick", "plan-audit.md", "PASS", repo.Commit)
	k, err := db.Transition(ctx, frHolderRequest(c, CardKickoff))
	if err != nil {
		t.Fatalf("T7: %v", err)
	}
	if k.LeaseHolder != "" || k.LeaseExpiresAt != "" || k.DecisionGate != DecisionGateKickoff {
		t.Fatalf("kickoff card holder=%q expiry=%q gate=%q", k.LeaseHolder, k.LeaseExpiresAt, k.DecisionGate)
	}
	transitionsBefore := len(frEvents(t, db, "card.transition"))
	later := frNow.Add(2 * FactoryLeaseDuration)
	a, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "kick", To: CardAssigned, ExpectedVersion: k.Version, Actor: "operator", Decider: DeciderHuman, Now: later})
	if err != nil {
		t.Fatalf("approve after clock advance: %v", err)
	}
	if a.State != CardAssigned || a.Stage != CardRun || a.Decider != DeciderHuman {
		t.Fatalf("approved card = %s stage=%s decider=%q", a.State, a.Stage, a.Decider)
	}
	if n := len(frEvents(t, db, "card.transition")) - transitionsBefore; n != 1 {
		t.Fatalf("card.transition events for the decision = %d, want 1", n)
	}

	q := frLeasedCard(repo, "question", CardRun)
	frPlace(t, db, q)
	nd, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "question", To: CardNeedsDecision, ExpectedVersion: 1, Actor: "worker-1", Question: "which API?", Now: frNow})
	if err != nil {
		t.Fatalf("T21: %v", err)
	}
	if nd.LeaseHolder != "" || nd.DecisionResume != CardRun || nd.DecisionQuestion != "which API?" {
		t.Fatalf("needs-decision card holder=%q resume=%q question=%q", nd.LeaseHolder, nd.DecisionResume, nd.DecisionQuestion)
	}
	r, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "question", To: CardAssigned, ExpectedVersion: nd.Version, Actor: "operator", Decider: DeciderHuman, Now: later})
	if err != nil || r.State != CardAssigned || r.Stage != CardRun {
		t.Fatalf("resume after clock advance: state=%s stage=%s err=%v", r.State, r.Stage, err)
	}
	if n := len(frEvents(t, db, "lease.expired")); n != 0 {
		t.Fatalf("lease.expired events = %d, want 0", n)
	}
}

// AC-017 (store half) — unblock keeps the stage; failed needs a reason,
// clears the lease, and is terminal.
func TestFR_AC017_UnblockAndFailed(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	frPlace(t, db, Card{RunID: frRun, CardID: "blocked", State: CardBlocked, Version: 1, OwnerLabel: "worker-1", Stage: CardSync, WorktreePath: repo.Dir})
	u, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "blocked", To: CardAssigned, ExpectedVersion: 1, Actor: "operator", Decider: DeciderHuman, Now: frNow})
	if err != nil || u.State != CardAssigned || u.Stage != CardSync {
		t.Fatalf("unblock: state=%s stage=%s err=%v", u.State, u.Stage, err)
	}

	c := frLeasedCard(repo, "fail", CardRun)
	frPlace(t, db, c)
	before := frRowDump(t, db, frRun, "fail")
	req := frHolderRequest(c, CardFailed)
	if _, err := db.Transition(ctx, req); err == nil {
		t.Fatal("failed with an empty reason was accepted")
	}
	if frRowDump(t, db, frRun, "fail") != before {
		t.Fatal("refused failed changed the record")
	}
	req.Reason = "build broken"
	f, err := db.Transition(ctx, req)
	if err != nil {
		t.Fatalf("failed with reason: %v", err)
	}
	if f.State != CardFailed || f.FailureReason != "build broken" || f.LeaseHolder != "" || f.LeaseExpiresAt != "" {
		t.Fatalf("failed card = %s reason=%q holder=%q expiry=%q", f.State, f.FailureReason, f.LeaseHolder, f.LeaseExpiresAt)
	}
	for _, to := range cardStates {
		_, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "fail", To: to, ExpectedVersion: f.Version, Actor: "worker-1", Decider: DeciderHuman, Reason: "x", Question: "q", Owner: "worker-1", Now: frNow})
		if err == nil {
			t.Fatalf("failed → %s accepted, want refused (failed is terminal)", to)
		}
	}
}
