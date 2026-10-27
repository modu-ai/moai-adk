package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// An open PR is pending integration, not an active implementation lease.
func TestFactoryPROpenReleasesLaneForIndependentCard(t *testing.T) {
	f := ghfNew(t, ghfOpts{syncStatus: "complete"})
	sdClearLaneEnv(t)
	if _, _, err := runTodo(t, "add", "independent follow-up"); err != nil {
		t.Fatal(err)
	}
	store := factory.NewBacklogStore(todoBacklogPath(f.root))
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	sdRegisterLane(t, f.root, ghfLane)
	secondTree := filepath.Join(t.TempDir(), "t2")
	fcGit(t, f.root, "worktree", "add", "-q", "-b", "WT-second-card", secondTree)
	fcPlace(t, f.root, homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: ghfLane, WorktreePath: secondTree})
	sdLaneEnv(t, ghfLane, "")
	newGHDouble(t, f)
	if _, err := ghfComplete(t); err != nil {
		t.Fatalf("deliver first card: %v", err)
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardPROpen || c.LeaseHolder != "" {
		t.Fatalf("first card: state=%s lease=%q", c.State, c.LeaseHolder)
	}
	db := fcOpen(t, f.root)
	beforeDone := fcCard(t, f.root, "t1")
	_, err := db.Transition(context.Background(), homestate.TransitionRequest{RunID: fcRun, CardID: "t1", To: homestate.CardDone, ExpectedVersion: beforeDone.Version, Actor: "operator", Decider: homestate.DeciderHuman, Now: fcNow})
	if !errors.Is(err, homestate.ErrIllegalTransition) {
		t.Fatalf("done before landing: %v", err)
	}
	if after := fcCard(t, f.root, "t1"); after != beforeDone {
		t.Fatal("refused done changed card")
	}
	if got := fbLeasedCard(t, f.root, ghfLane); got != "t2" {
		t.Fatalf("next card = %q, want independent t2", got)
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardPROpen {
		t.Fatalf("open PR became %s before landing", c.State)
	}
}

func TestFactoryExpiredLeaseReacquireRefusalsPreserveState(t *testing.T) {
	for _, tc := range []struct {
		name, owner, holder string
		queue               factory.BacklogState
		expires             time.Time
	}{
		{"other owner", "lane-2", "lane-2", factory.BacklogStatePicked, fcNow.Add(-time.Minute)},
		{"hold", "lane-1", "lane-1", factory.BacklogStateHold, fcNow.Add(-time.Minute)},
		{"live lease", "lane-1", "lane-1", factory.BacklogStatePicked, fcNow.Add(time.Hour)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, store := nmBase(t, tc.queue)
			fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun, OwnerLabel: tc.owner, LeaseHolder: tc.holder, LeaseExpiresAt: tc.expires.Format(time.RFC3339Nano)})
			nmLaneEnv(t, "lane-1", "")
			before := nmSnapshot(t, root, store)
			_, _, err := runFactory(t, "next", "--run", fcRun, "--card", "t1")
			if err == nil {
				t.Fatal("unexpected lease")
			}
			if after := nmSnapshot(t, root, store); after != before {
				t.Fatal("refused request changed queue or card record")
			}
		})
	}
}

func TestFactoryExpiredLeaseReacquireKeepsDirtyTreeAndRejectsDuplicate(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStatePicked)
	setReacquireRegistration(t, root)
	tree := filepath.Join(t.TempDir(), "card")
	fcGit(t, root, "worktree", "add", "-q", "-b", "WT-reacquire", tree)
	file := filepath.Join(tree, "uncommitted.txt")
	if err := os.WriteFile(file, []byte("keep me"), 0600); err != nil {
		t.Fatal(err)
	}
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun, OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano), HeartbeatAt: fcNow.Add(-2 * time.Minute).Format(time.RFC3339Nano), WorktreePath: tree})
	nmLaneEnv(t, "lane-1", "")
	if _, stderr, err := runFactory(t, "next", "--run", fcRun, "--card", "t1"); err != nil {
		t.Fatalf("reacquire: %v %s", err, stderr)
	}
	before := nmSnapshot(t, root, store)
	if _, _, err := runFactory(t, "next", "--run", fcRun, "--card", "t1"); err == nil {
		t.Fatal("duplicate request leased again")
	}
	if after := nmSnapshot(t, root, store); after != before {
		t.Fatal("duplicate changed record")
	}
	data, err := os.ReadFile(file)
	if err != nil || strings.TrimSpace(string(data)) != "keep me" {
		t.Fatalf("dirty file lost: %q %v", data, err)
	}
}

func TestFactorySameOwnerReacquiresExpiredLease(t *testing.T) {
	root, _ := nmBase(t, factory.BacklogStatePicked)
	nmIsolatedWorktrees(t, "t1")
	setReacquireRegistration(t, root)
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano),
		HeartbeatAt:    fcNow.Add(-2 * time.Minute).Format(time.RFC3339Nano),
	})
	nmLaneEnv(t, "lane-1", "")
	_, stderr, err := runFactory(t, "next", "--run", fcRun, "--card", "t1")
	if err != nil {
		t.Fatalf("same owner could not reacquire expired lease: %v; stderr=%q", err, stderr)
	}
	c := fcCard(t, root, "t1")
	if c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
		t.Fatalf("reacquired card = %s holder=%q", c.State, c.LeaseHolder)
	}
}

func TestExpiredReacquirePreservesClaimDeadlineAndPRSnapshot(t *testing.T) {
	root, _ := nmBase(t, factory.BacklogStatePicked)
	nmIsolatedWorktrees(t, "t1")
	setReacquireRegistration(t, root)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano),
		HeartbeatAt:    fcNow.Add(-2 * time.Minute).Format(time.RFC3339Nano)})
	nmLaneEnv(t, "lane-1", "")
	before := fcCard(t, root, "t1")
	db := fcOpen(t, root)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	_, leased, _, err := factoryNextNominatedClaim(ctx, db, root, fcRun, "lane-1", factoryNominee{item: factory.BacklogItem{ID: "t1"}, row: &before}, homestate.CardFields{})
	if leased || !errors.Is(err, errFactoryRecordBusy) {
		t.Errorf("expired claim: leased=%v err=%v, want bounded record-busy refusal", leased, err)
	}
	if after := fcCard(t, root, "t1"); after != before {
		t.Fatal("deadline refusal changed the card")
	}
	gh := flCountGhCalls(t, func() (string, error) { return "[]", nil })
	q := factoryNextPRSnapshot(context.Background(), root, fcRun, "lane-1", "t1", false)
	if q.err != nil || gh() != 1 {
		t.Errorf("eligible expired claim PR snapshot: calls=%d err=%v, want one query", gh(), q.err)
	}
	q = factoryNextPRSnapshot(context.Background(), root, fcRun, "lane-2", "t1", false)
	if !errors.Is(q.err, errFactoryNextPRNotTaken) || gh() != 1 {
		t.Errorf("other lane PR snapshot: calls=%d err=%v, want no extra query", gh(), q.err)
	}
}

func TestExpiredReacquireBusyRecordIsBounded(t *testing.T) {
	root, _ := nmBase(t, factory.BacklogStatePicked)
	nmIsolatedWorktrees(t, "t1")
	setReacquireRegistration(t, root)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano),
		HeartbeatAt:    fcNow.Add(-2 * time.Minute).Format(time.RFC3339Nano)})
	before := fcCard(t, root, "t1")
	db, err := homestate.OpenFactoryBounded(root, factoryLeaseClaimBusyTimeout)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	release := flHoldRecord(t, root)
	old := factoryLeaseClaimDeadlineOverride
	factoryLeaseClaimDeadlineOverride = 200 * time.Millisecond
	t.Cleanup(func() { factoryLeaseClaimDeadlineOverride = old })
	start := time.Now()
	_, leased, _, err := factoryNextNominatedClaim(context.Background(), db, root, fcRun, "lane-1", factoryNominee{item: factory.BacklogItem{ID: "t1"}, row: &before}, homestate.CardFields{})
	elapsed := time.Since(start)
	release()
	if leased || !errors.Is(err, errFactoryRecordBusy) {
		t.Fatalf("busy record: leased=%v err=%v", leased, err)
	}
	if elapsed > 200*time.Millisecond+factoryLeaseClaimBusyTimeout+flOverheadCeil {
		t.Fatalf("claim took %s, beyond its deadline and busy allowance", elapsed)
	}
	if after := fcCard(t, root, "t1"); after != before {
		t.Fatal("busy refusal changed the card")
	}
}

func setReacquireRegistration(t *testing.T, root string) {
	t.Helper()
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE workers SET registered_at=? WHERE label=?`, fcNow.Add(-time.Hour).Format(time.RFC3339Nano), "lane-1"); err != nil {
		t.Fatal(err)
	}
}

func TestCollectedMergeReadyDoesNotHoldImplementationSlot(t *testing.T) {
	card := homestate.Card{State: homestate.CardAssigned, Stage: homestate.CardMergeReady, OwnerLabel: "lane-1", HeartbeatAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano)}
	if factorySerialSlotHeld(card, fcNow, func(homestate.Card) bool { return true }) {
		t.Fatal("collected merge-ready card still holds the implementation slot")
	}
}

func TestFactoryNextAfterMergeReadyExpiryCollection(t *testing.T) {
	root, _ := expiredLeaseFixture(t)
	serialSlotWorkerPID(t, root, "lane-2", os.Getpid(), serialSlotIncarnationEarly)
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE cards SET stage=? WHERE run_id=? AND card_id=?`, homestate.CardMergeReady, fcRun, "t1"); err != nil {
		t.Fatal(err)
	}
	card := fcCard(t, root, "t1")
	_, err := db.Transition(context.Background(), homestate.TransitionRequest{RunID: fcRun, CardID: "t1", To: homestate.CardMergeReady, ExpectedVersion: card.Version, Actor: "lane-2", Now: fcNow})
	if !errors.Is(err, homestate.ErrLeaseExpired) {
		t.Fatalf("collect expired card: %v", err)
	}
	if got := fcCard(t, root, "t1"); got.State != homestate.CardAssigned || got.Stage != homestate.CardMergeReady {
		t.Fatalf("collected card: %+v", got)
	}
	if got := fbLeasedCard(t, root, "lane-1"); got != "t2" {
		t.Fatalf("next card = %q, want t2", got)
	}
	if got := fcCard(t, root, "t1"); got.State != homestate.CardAssigned {
		t.Fatalf("waiting card changed to %s", got.State)
	}
}
