package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// gated is a claimed class-A card with passing gate evidence for commit sha.
func gated(t *testing.T, s *Store, id, sha string, risk ...string) Claimed {
	t.Helper()
	if err := s.AddCard(bg, Card{ID: id, Title: id, Position: core.Position{Stage: core.StageReady}, Priority: 1,
		Spec: Spec{Class: "A", Paths: []string{"internal/" + id + "/x.go"}, Risk: risk}}); err != nil {
		t.Fatal(err)
	}
	c, err := s.Claim(bg, req("s-"+id, 10))
	if err != nil || c.CardID != id {
		t.Fatalf("claim %s: %+v %v", id, c, err)
	}
	for _, e := range []Evidence{
		{Kind: KindGreen, TreeSHA: "tree-" + sha, SubjectSHA: sha, Result: ResultPass},
		{Kind: KindGate, TreeSHA: "tree-" + sha, SubjectSHA: sha, Command: "go test ./...", Result: ResultPass},
	} {
		if err := s.AddEvidence(bg, id, c.Epoch, e); err != nil {
			t.Fatal(err)
		}
	}
	return c
}

func queueRow(t *testing.T, s *Store, id string) (QueueItem, bool) {
	t.Helper()
	q, err := s.Queue(bg)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range q {
		if it.CardID == id {
			return it, true
		}
	}
	return QueueItem{}, false
}

func TestEnqueueBindsTheCandidateToPassingGateEvidenceAndMovesTheCardIntoMerging(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	it, ok := queueRow(t, s, "a")
	if !ok || it.State != QueueWaiting || it.CandidateSHA != "c1" || it.EnqueuedAt.IsZero() {
		t.Fatalf("queue row: %+v %v", it, ok)
	}
	var cardEpoch int64
	if err := s.db.QueryRowContext(bg, `SELECT card_epoch FROM merge_queue WHERE card_id='a'`).Scan(&cardEpoch); err != nil || cardEpoch != c.Epoch {
		t.Errorf("card_epoch %d %v", cardEpoch, err)
	}
	if got := cardOf(t, s, "a").Position; got.Stage != core.StageMerging || got.Park != nil {
		t.Errorf("the card is in MERGING: %+v", got)
	}
	var n int
	if err := s.db.QueryRowContext(bg, `SELECT count(*) FROM transitions WHERE card_id='a' AND to_stage='MERGING' AND actor='enqueue'`).Scan(&n); err != nil || n != 1 {
		t.Errorf("the move is recorded: %d %v", n, err)
	}
}

func TestEnqueueWithoutGateEvidenceForThatCommitIsStale(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, TreeSHA: "tree-c3", SubjectSHA: "c3", Result: ResultFail}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindRed, TreeSHA: "tree-c4", SubjectSHA: "c4", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	for name, sha := range map[string]string{"another commit": "c2", "a failing gate": "c3", "not a gate": "c4"} {
		if err := s.Enqueue(bg, "a", c.Epoch, sha); !errors.Is(err, ErrEvidenceStale) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, ok := queueRow(t, s, "a"); ok {
		t.Error("a refused enqueue left a queue row")
	}
	if got := cardOf(t, s, "a").Position; got.Stage != core.StageBuilding {
		t.Errorf("the card did not move: %+v", got)
	}
	if err := s.Enqueue(bg, "a", c.Epoch, ""); !errors.Is(err, ErrInvalid) {
		t.Errorf("an empty candidate: %v", err)
	}
	if err := s.Enqueue(bg, "a", c.Epoch+1, "c1"); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("a stale epoch: %v", err)
	}
}

func TestACardWithRiskNeedsReviewingAndPassingReviewEvidence(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1", "security")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); !errors.Is(err, ErrEvidenceStale) {
		t.Fatalf("a card with risk, still BUILDING: %v", err)
	}
	if err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageReviewing}, "lane", "ready for review"); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); !errors.Is(err, ErrEvidenceStale) {
		t.Fatalf("REVIEWING but no review evidence: %v", err)
	}
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindReview, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultFail}); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); !errors.Is(err, ErrEvidenceStale) {
		t.Fatalf("a failing review: %v", err)
	}
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindReview, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatalf("with a passing review: %v", err)
	}
}

func TestACardInTheQueueIsNotEnqueuedAgainUntilItFailed(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); !errors.Is(err, ErrInvalid) {
		t.Errorf("twice: %v", err)
	}
}

// waiting, verifying and failed carry the card epoch; landing and landed carry the
// epoch of the landing lock; and so does a reconcile that sends a card back to waiting.
func TestQueueUpdateChecksTheEpochItsStateCallsFor(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	// the lock's epoch is not the card's, or the checks below could not tell them apart
	warm, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "warm"}, LandingLockTTL)
	if err != nil || s.ReleaseLock(bg, LockMerge, warm) != nil {
		t.Fatalf("warm-up lock: %d %v", warm, err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch+7), QueueVerifying, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("verifying with a wrong card epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	// landing is the lock's
	if err := s.QueueUpdate(bg, "a", LockEpoch(c.Epoch), QueueLanding, "trunk0", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("landing with the card epoch and no lock: %v", err)
	}
	lock, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq", PID: 1, PIDStart: "t"}, LandingLockTTL)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock+5), QueueLanding, "trunk0", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("landing with another lock epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock), QueueLanding, "trunk0", ""); err != nil {
		t.Fatal(err)
	}
	if it, _ := queueRow(t, s, "a"); it.State != QueueLanding || it.TrunkBefore != "trunk0" {
		t.Errorf("landing recorded: %+v", it)
	}
	// a reconcile sends a landing row back to waiting, under the lock
	if err := s.QueueUpdate(bg, "a", LockEpoch(c.Epoch), QueueWaiting, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("waiting from landing with the card epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock), QueueWaiting, "", ""); err != nil {
		t.Errorf("waiting from landing with the lock epoch: %v", err)
	}
	// and a normal run to the end
	for _, step := range []struct {
		epoch QueueEpoch
		state QueueState
	}{{CardEpoch(c.Epoch), QueueVerifying}, {LockEpoch(lock), QueueLanding}, {LockEpoch(lock), QueueLanded}} {
		if err := s.QueueUpdate(bg, "a", step.epoch, step.state, "", ""); err != nil {
			t.Fatalf("%s: %v", step.state, err)
		}
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(c.Epoch), QueueWaiting, "", ""); !errors.Is(err, ErrInvalid) {
		t.Errorf("landed is final: %v", err)
	}
}

func TestQueueUpdateRefusesStatesThatDoNotFollowAndAFailureWithoutAReason(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	lock, _ := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL)
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock), QueueLanded, "", ""); !errors.Is(err, ErrInvalid) {
		t.Errorf("waiting to landed: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueFailed, "", ""); !errors.Is(err, ErrInvalid) {
		t.Errorf("failed with no reason: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueFailed, "", "gate red on the merge result"); err != nil {
		t.Fatal(err)
	}
	if it, _ := queueRow(t, s, "a"); it.State != QueueFailed || it.FailReason != "gate red on the merge result" {
		t.Errorf("failed recorded: %+v", it)
	}
	if err := s.QueueUpdate(bg, "nobody", CardEpoch(1), QueueVerifying, "", ""); !errors.Is(err, ErrNotFound) {
		t.Errorf("no such queue row: %v", err)
	}
}

// A card that was reclaimed and claimed again has a higher cards.epoch: a lander that
// is still working from the older epoch is fenced, lock or no lock.
func TestAReclaimedCardFencesItsStaleLander(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	lock, _ := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL)
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(bg, `UPDATE cards SET epoch=epoch+1 WHERE id='a'`); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock), QueueLanding, "t", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("a lander of the old card epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueFailed, "", "stale"); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("a verifier of the old card epoch: %v", err)
	}
}

func TestAFailedCardCanBeEnqueuedAgainWithANewCandidate(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueFailed, "", "red"); err != nil {
		t.Fatal(err)
	}
	// the lane fixes it: back to BUILDING/FIX, new commit, new gate evidence
	if err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageBuilding, Sub: core.SubFix}, "lane", "queue failed"); err == nil {
		// not every path is an edge; the point here is the queue row, so go through REVIEWING
		t.Log("moved straight to FIX")
	}
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, TreeSHA: "tree-c2", SubjectSHA: "c2", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(bg, `UPDATE cards SET stage='BUILDING', sub_stage='IMPLEMENT' WHERE id='a'`); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "a", c.Epoch, "c2"); err != nil {
		t.Fatalf("a failed row may be queued again: %v", err)
	}
	if it, _ := queueRow(t, s, "a"); it.State != QueueWaiting || it.CandidateSHA != "c2" || it.FailReason != "" {
		t.Errorf("the row starts over: %+v", it)
	}
}

// Passing gate evidence for another commit does not stand for this one, even when it is
// the card's newest.
func TestGateEvidenceForAnotherCommitDoesNotStandForThisOne(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c2"); !errors.Is(err, ErrEvidenceStale) {
		t.Errorf("%v", err)
	}
}

func TestAQueueRowKeepsItsTrunkAndOnlyAFailureKeepsAReason(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	lock, _ := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL)
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueVerifying, "", "a reason on a state that is not a failure"); err != nil {
		t.Fatal(err)
	}
	if it, _ := queueRow(t, s, "a"); it.FailReason != "" {
		t.Errorf("only a failed row has a reason: %+v", it)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock), QueueLanding, "trunk0", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock), QueueLanded, "", ""); err != nil {
		t.Fatal(err)
	}
	if it, _ := queueRow(t, s, "a"); it.TrunkBefore != "trunk0" {
		t.Errorf("an empty trunk does not erase the one recorded: %+v", it)
	}
}
