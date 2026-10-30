package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

var (
	posBuilding = core.Position{Stage: core.StageBuilding, Sub: core.SubImplement}
	posReview   = core.Position{Stage: core.StageReviewing}
	posDone     = core.Position{Stage: core.StageDone}
)

func TestMoveNeedsTheLeaseEpochAndTheFactsTheLedgerHolds(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	if err := s.Move(bg, "a", c.Epoch+1, posReview, "lane", "r"); !errors.Is(err, ErrStaleEpoch) {
		t.Fatalf("stale epoch: %v", err)
	}
	// BUILDING to REVIEWING (E03) stands on a receipt and its acceptance: here, green
	// evidence and gate evidence for the same tree
	if err := s.Move(bg, "a", c.Epoch, posReview, "lane", "r"); !errors.Is(err, core.ErrGuardMissing) {
		t.Fatalf("no evidence: %v", err)
	}
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGreen, TreeSHA: "t1", SubjectSHA: "t1", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "a", c.Epoch, posReview, "lane", "r"); !errors.Is(err, core.ErrGuardMissing) {
		t.Fatalf("green but no gate: %v", err)
	}
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, TreeSHA: "other", SubjectSHA: "other", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "a", c.Epoch, posReview, "lane", "r"); !errors.Is(err, core.ErrGuardMissing) {
		t.Fatalf("a gate for another tree: %v", err)
	}
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, TreeSHA: "t1", SubjectSHA: "t1", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "a", c.Epoch, posReview, "lane-1", "reviewable"); err != nil {
		t.Fatal(err)
	}
	if got := cardOf(t, s, "a").Position; got.Stage != core.StageReviewing {
		t.Errorf("position: %+v", got)
	}
	var actor, reason string
	if err := s.db.QueryRowContext(bg, `SELECT actor,reason FROM transitions WHERE card_id='a' AND to_stage='REVIEWING'`).Scan(&actor, &reason); err != nil || actor != "lane-1" || reason != "reviewable" {
		t.Errorf("the transition: %q %q %v", actor, reason, err)
	}
	// a position no edge reaches is refused and stores nothing
	if err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageReady}, "lane", "back"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("backwards: %v", err)
	}
}

func TestMergingIsEnteredByEnqueueAlone(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	if err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageMerging}, "lane", "r"); !errors.Is(err, ErrInvalid) {
		t.Errorf("%v", err)
	}
}

func TestGuardsCarryGuardKeysAndNeverFacts(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	for _, g := range []string{"merged", "closed", "settled", "receipt", "receipt-accepted", "queued", ""} {
		err := s.Move(bg, "a", c.Epoch, posBuilding, "lane", "r", Evidence{Kind: KindGate, TreeSHA: "t", SubjectSHA: "t", Result: ResultPass, Guards: []string{g}})
		if !errors.Is(err, ErrInvalid) {
			t.Errorf("guard %q: %v", g, err)
		}
	}
	if got, _ := s.EvidenceFor(bg, "a"); len(got) != 0 {
		t.Errorf("a refused move stored evidence: %+v", got)
	}
	if err := s.Move(bg, "a", c.Epoch, posBuilding, "lane", "r", Evidence{Kind: KindGate, TreeSHA: "t", SubjectSHA: "t", Result: ResultPass, Guards: []string{"resume-guard:x"}}); err != nil {
		t.Errorf("a resume guard key: %v", err)
	}
}

func TestMoveToDoneOrCancelledReleasesTheLeaseAndKeysInTheSameTransaction(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "gone", "internal/g/x.go")
	if err := s.Move(bg, "gone", c.Epoch, core.Position{Stage: core.StageCancelled}, "lane", "dropped"); err != nil {
		t.Fatal(err)
	}
	if ls, _ := s.Leases(bg); len(ls) != 0 {
		t.Errorf("a cancelled card holds no lease: %+v", ls)
	}
	var keys int
	if err := s.db.QueryRowContext(bg, `SELECT count(*) FROM lease_keys`).Scan(&keys); err != nil || keys != 0 {
		t.Errorf("keys %d %v", keys, err)
	}
	// and the same for a card that lands
	d := gated(t, s, "done", "c9")
	if err := s.Enqueue(bg, "done", d.Epoch, "c9"); err != nil {
		t.Fatal(err)
	}
	landOne(t, s, "done", d.Epoch)
	if err := s.Move(bg, "done", d.Epoch, core.Position{Stage: core.StageMerging, RemoteIntegrated: true}, "mergeq", "merge seen"); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "done", d.Epoch, posDone, "mergeq", "landed"); err != nil {
		t.Fatal(err)
	}
	if ls, _ := s.Leases(bg); len(ls) != 0 {
		t.Errorf("a done card holds no lease: %+v", ls)
	}
	if got := cardOf(t, s, "done").Position; got.Stage != core.StageDone {
		t.Errorf("%+v", got)
	}
}

// landOne takes the card's queue row to landed the way mergeq does.
func landOne(t *testing.T, s *Store, id string, cardEpoch int64) {
	t.Helper()
	lock, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq", PID: 1, PIDStart: "t"}, LandingLockTTL)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []struct {
		epoch int64
		state QueueState
	}{{cardEpoch, QueueVerifying}, {lock, QueueLanding}, {lock, QueueLanded}} {
		if err := s.QueueUpdate(bg, id, step.epoch, step.state, "trunk", ""); err != nil {
			t.Fatalf("%s: %v", step.state, err)
		}
	}
	if err := s.ReleaseLock(bg, LockMerge, lock); err != nil {
		t.Fatal(err)
	}
}

// merged, closed and settled come from the queue row of the card's current epoch and
// from nowhere else: not from a note, not from a row of an older epoch.
func TestTheFinishingFactsComeFromTheCurrentEpochsQueueRowOnly(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddNote(bg, Note{Kind: NoteLanded, CardID: "a", Text: "landed at c1"}); err != nil {
		t.Fatal(err)
	}
	merged := core.Position{Stage: core.StageMerging, RemoteIntegrated: true}
	if err := s.Move(bg, "a", c.Epoch, merged, "mergeq", "r"); !errors.Is(err, core.ErrGuardMissing) {
		t.Fatalf("a landed note is not a landed row: %v", err)
	}
	landOne(t, s, "a", c.Epoch)
	// the card was claimed again since: the old row does not speak for the new epoch
	if _, err := s.db.ExecContext(bg, `UPDATE cards SET epoch=epoch+1 WHERE id='a'`); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "a", c.Epoch, merged, "mergeq", "r"); !errors.Is(err, core.ErrGuardMissing) {
		t.Fatalf("a row of an older epoch: %v", err)
	}
	if _, err := s.db.ExecContext(bg, `UPDATE cards SET epoch=epoch-1 WHERE id='a'`); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "a", c.Epoch, merged, "mergeq", "r"); err != nil {
		t.Fatalf("the current epoch's landed row: %v", err)
	}
}

// A park the ledger made for the operator holds no lease. The operator's resume is a
// Move at the card's own epoch, in place, with the resume guard, and it clears the
// reclaim count.
func TestAnOperatorResumeOfAReclaimCapParkNeedsNoLeaseAndClearsTheCount(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 9, "A", nil, "internal/a/x.go")
	for i := 0; i < ReclaimCap; i++ {
		r := req("h", 100+i)
		if i > 0 {
			r.Alive = alwaysDead
		}
		_, _ = s.Claim(bg, r)
	}
	r := req("last", 999)
	r.Alive = alwaysDead
	_, _ = s.Claim(bg, r)
	parked := cardOf(t, s, "a")
	if parked.Position.Park == nil || parked.Position.Park.Reason != ParkReasonReclaimCap || parked.ReclaimCount != ReclaimCap {
		t.Fatalf("setup: %+v", parked)
	}
	free := core.Position{Stage: parked.Position.Stage, Sub: parked.Position.Sub}
	resume := func(epoch int64, to core.Position, g ...string) error {
		return s.Move(bg, "a", epoch, to, "operator", "resume", Evidence{Kind: KindGate, TreeSHA: "n/a", SubjectSHA: "n/a", Result: ResultPass, Guards: g})
	}
	if err := resume(parked.Epoch, free); !errors.Is(err, core.ErrGuardMissing) {
		t.Errorf("without the resume guard: %v", err)
	}
	if err := resume(parked.Epoch+1, free, "resume-guard:reclaim-cap"); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("a wrong epoch: %v", err)
	}
	if err := resume(parked.Epoch, core.Position{Stage: core.StageReviewing}, "resume-guard:reclaim-cap"); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("without a lease only the resume in place is allowed: %v", err)
	}
	if err := resume(parked.Epoch, free, "resume-guard:reclaim-cap"); err != nil {
		t.Fatal(err)
	}
	got := cardOf(t, s, "a")
	if got.Position.Park != nil || got.ReclaimCount != 0 {
		t.Errorf("resumed and counted from zero: %+v", got)
	}
	if c, err := s.Claim(bg, req("again", 5)); err != nil || c.CardID != "a" {
		t.Errorf("the card can be claimed again: %+v %v", c, err)
	}
}

func TestAnInPlaceParkNeedsItsOwnGuard(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	to := core.Position{Stage: core.StageBuilding, Sub: core.SubImplement, Park: &core.Park{Reason: ParkReasonFixLimit, Evidence: "3 fix rounds", Resume: "the operator decides"}}
	if err := s.Move(bg, "a", c.Epoch, to, "lane", "fix-limit"); !errors.Is(err, core.ErrGuardMissing) {
		t.Fatalf("no guard: %v", err)
	}
	if err := s.Move(bg, "a", c.Epoch, to, "lane", "fix-limit", Evidence{Kind: KindGate, TreeSHA: "t", SubjectSHA: "t", Result: ResultFail, Guards: []string{core.GuardFixLimit}}); err != nil {
		t.Fatal(err)
	}
	if got := cardOf(t, s, "a").Position.Park; got == nil || got.Reason != ParkReasonFixLimit {
		t.Errorf("parked: %+v", got)
	}
}

// A failing green run is not a receipt.
func TestAFailingGreenRunIsNoReceipt(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	for _, e := range []Evidence{
		{Kind: KindGreen, TreeSHA: "t1", SubjectSHA: "t1", Result: ResultFail},
		{Kind: KindGate, TreeSHA: "t1", SubjectSHA: "t1", Result: ResultPass},
	} {
		if err := s.AddEvidence(bg, "a", c.Epoch, e); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Move(bg, "a", c.Epoch, posReview, "lane", "r"); !errors.Is(err, core.ErrGuardMissing) {
		t.Errorf("%v", err)
	}
}

func TestMoveStoresNoEvidenceItWouldRefuseToAdd(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	for name, e := range map[string]Evidence{
		"kind":   {Kind: "vibes", TreeSHA: "t", SubjectSHA: "t", Result: ResultPass},
		"result": {Kind: KindGate, TreeSHA: "t", SubjectSHA: "t", Result: "maybe"},
		"tree":   {Kind: KindGate, Result: ResultPass},
	} {
		if err := s.Move(bg, "a", c.Epoch, posBuilding, "lane", "r", e); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if got, _ := s.EvidenceFor(bg, "a"); len(got) != 0 {
		t.Errorf("stored: %+v", got)
	}
}
