package ledger

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// P2-1: a lane that dies between Enqueue and DONE must not strand the card. The queue
// row is the authority once it is settled; the card's own epoch is the key.
func TestACardStrandedInMergingByItsHoldersDeathIsSettledThroughTheQueue(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if got, err := s.Reap(bg, alwaysDead); err != nil || len(got) != 1 {
		t.Fatalf("the holder is reclaimed: %+v %v", got, err)
	}
	merged := core.Position{Stage: core.StageMerging, RemoteIntegrated: true}
	// not settled yet: nothing moves it without a lease
	if err := s.Move(bg, "a", c.Epoch, merged, "mergeq", "r"); !errors.Is(err, ErrStaleEpoch) {
		t.Fatalf("a waiting row settles nothing: %v", err)
	}
	landOne(t, s, "a", c.Epoch)
	if err := s.Move(bg, "a", c.Epoch+1, merged, "mergeq", "r"); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("only at the card's own epoch: %v", err)
	}
	if err := s.Move(bg, "a", c.Epoch, merged, "mergeq", "merge seen"); err != nil {
		t.Fatalf("landed, no lease: %v", err)
	}
	if err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageDone}, "mergeq", "landed"); err != nil {
		t.Fatalf("done, no lease: %v", err)
	}
	if got := cardOf(t, s, "a").Position; got.Stage != core.StageDone {
		t.Errorf("%+v", got)
	}
}

func TestAFailedRowOfAStrandedCardSendsItBackToReviewingWhereItCanBeClaimed(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "b", "c1")
	if err := s.Enqueue(bg, "b", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Reap(bg, alwaysDead); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "b", CardEpoch(c.Epoch), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "b", CardEpoch(c.Epoch), QueueFailed, "", "gate red on the merge result"); err != nil {
		t.Fatal(err)
	}
	// a Move that is no edge is still refused, with the table's error, not as a missing lease
	if err := s.Move(bg, "b", c.Epoch, core.Position{Stage: core.StageDone}, "mergeq", "r"); errors.Is(err, ErrStaleEpoch) || err == nil {
		t.Errorf("the table refuses it: %v", err)
	}
	if err := s.Move(bg, "b", c.Epoch, core.Position{Stage: core.StageReviewing}, "mergeq", "queue failed"); err != nil {
		t.Fatalf("back to REVIEWING without a lease: %v", err)
	}
	got, err := s.Claim(bg, req("next", 77))
	if err != nil || got.CardID != "b" || got.Epoch != c.Epoch+1 {
		t.Errorf("the card is claimable again: %+v %v", got, err)
	}
}

// P2-2: mergeq reads the card epoch from the row it is writing.
func TestAQueueItemCarriesTheCardEpochItWasQueuedAt(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	it, ok := queueRow(t, s, "a")
	if !ok || it.CardEpoch != c.Epoch {
		t.Fatalf("%+v", it)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(it.CardEpoch), QueueVerifying, "", ""); err != nil {
		t.Errorf("a write with the epoch the item carries: %v", err)
	}
}

// P2-3: a card whose design was never audited does not enter MERGING on a queued fact.
func TestAClassCCardStillInSpecCannotEnterMerging(t *testing.T) {
	s := openTemp(t)
	if err := s.AddCard(bg, Card{ID: "c", Title: "c", Position: core.Position{Stage: core.StageReady}, Spec: Spec{Class: "C", Paths: []string{"internal/c/x.go"}}}); err != nil {
		t.Fatal(err)
	}
	cl, err := s.Claim(bg, req("s", 1))
	if err != nil || cl.Position.Sub != core.SubSpec {
		t.Fatalf("%+v %v", cl, err)
	}
	if err := s.AddEvidence(bg, "c", cl.Epoch, Evidence{Kind: KindGate, TreeSHA: "t", SubjectSHA: "c1", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "c", cl.Epoch, "c1"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Fatalf("BUILDING/SPEC does not enter MERGING: %v", err)
	}
	if _, ok := queueRow(t, s, "c"); ok {
		t.Error("a refused enqueue left a queue row")
	}
	if err := s.Move(bg, "c", cl.Epoch, core.Position{Stage: core.StageBuilding, Sub: core.SubImplement}, "lane", "spec audited"); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "c", cl.Epoch, "c1"); err != nil {
		t.Errorf("after the spec step: %v", err)
	}
}

// P3-4: evidence is judged by the commit that was measured, nothing else.
func TestEvidenceIsJudgedBySubjectSHAAlone(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, TreeSHA: "only-a-tree", Result: ResultPass}); !errors.Is(err, ErrInvalid) {
		t.Errorf("evidence with no subject_sha: %v", err)
	}
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGreen, SubjectSHA: "s1", Result: ResultPass}); err != nil {
		t.Errorf("a subject alone is enough: %v", err)
	}
	// green and gate on the same commit: reviewable. A gate for another commit is not.
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, SubjectSHA: "s2", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "a", c.Epoch, posReview, "lane", "r"); !errors.Is(err, core.ErrGuardMissing) {
		t.Fatalf("a gate for another commit: %v", err)
	}
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, SubjectSHA: "s1", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	// the newest gate row for the commit decides, as it does at Enqueue
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, SubjectSHA: "s1", Result: ResultFail}); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "a", c.Epoch, posReview, "lane", "r"); !errors.Is(err, core.ErrGuardMissing) {
		t.Fatalf("the newest gate for the commit failed: %v", err)
	}
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, SubjectSHA: "s1", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "a", c.Epoch, posReview, "lane", "r"); err != nil {
		t.Errorf("green and the newest gate agree on s1: %v", err)
	}
}

// P3-5: a restart of a failed row must not lose why the earlier attempt failed.
func TestRequeueingAFailedCardKeepsTheEarlierReasonInANote(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueFailed, "trunk9", "merge result red: TestX"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, SubjectSHA: "c2", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "a", c.Epoch, "c2"); err != nil {
		t.Fatal(err)
	}
	notes, _ := s.NotesSince(bg, 0)
	found := false
	for _, n := range notes {
		if n.Kind == NoteDecision && n.CardID == "a" && strings.Contains(n.Text, "c1") && strings.Contains(n.Text, "merge result red: TestX") && strings.Contains(n.Text, "c2") {
			found = true
		}
	}
	if !found {
		t.Errorf("no note keeps the earlier candidate and its reason: %+v", notes)
	}
}

// P3-8: a lander slower than its ttl can still release; one can renew.
func TestALockCanBeRenewedAndAnExpiredOneStillReleasesAtItsOwnEpoch(t *testing.T) {
	s := openTemp(t)
	e, err := s.AcquireLock(bg, "x", holder("a"), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RenewLock(bg, "x", e+1, time.Hour); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("renew at another epoch: %v", err)
	}
	if err := s.RenewLock(bg, "x", e, 0); !errors.Is(err, ErrInvalid) {
		t.Errorf("renew with no ttl: %v", err)
	}
	if err := s.RenewLock(bg, "x", e, time.Hour); err != nil {
		t.Fatal(err)
	}
	var secs float64
	if err := s.db.QueryRowContext(bg, `SELECT (julianday(expires_at)-julianday('now'))*86400 FROM locks WHERE name='x'`).Scan(&secs); err != nil || secs < 3500 {
		t.Errorf("the renewed lock runs about an hour: %v %v", secs, err)
	}
	// it ran out, and nobody took it
	if _, err := s.db.ExecContext(bg, `UPDATE locks SET expires_at='2000-01-01T00:00:00.000Z'`); err != nil {
		t.Fatal(err)
	}
	if err := s.RenewLock(bg, "x", e, time.Hour); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("a lock that ran out is taken again, not renewed: %v", err)
	}
	if err := s.ReleaseLock(bg, "x", e); err != nil {
		t.Errorf("the slow lander releases its own epoch: %v", err)
	}
	if err := s.ReleaseLock(bg, "x", e); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("released twice: %v", err)
	}
	// and once someone took it over, the old holder is fenced
	e1, _ := s.AcquireLock(bg, "y", holder("a"), time.Minute)
	if _, err := s.db.ExecContext(bg, `UPDATE locks SET expires_at='2000-01-01T00:00:00.000Z' WHERE name='y'`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AcquireLock(bg, "y", holder("b"), time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseLock(bg, "y", e1); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("after a takeover: %v", err)
	}
}

// P3-9: the snapshot's file and its directory are made durable.
func TestSnapshotSyncsTheFileThenTheDirectory(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
	var synced []string
	old := syncPath
	syncPath = func(p string) error { synced = append(synced, p); return nil }
	defer func() { syncPath = old }()
	dst := filepath.Join(t.TempDir(), "snap.db")
	if err := s.Snapshot(bg, dst); err != nil {
		t.Fatal(err)
	}
	if len(synced) != 2 || !strings.Contains(filepath.Base(synced[0]), ".tmp-") || synced[1] != filepath.Dir(dst) {
		t.Errorf("synced %v: want the temporary file, then %s", synced, filepath.Dir(dst))
	}
	// a failed sync fails the snapshot and leaves dst as it was
	if err := os.WriteFile(dst, []byte("previous"), 0o600); err != nil {
		t.Fatal(err)
	}
	syncPath = func(string) error { return errors.New("disk said no") }
	if err := s.Snapshot(bg, dst); err == nil {
		t.Error("a failed sync must fail the snapshot")
	}
	if b, _ := os.ReadFile(dst); string(b) != "previous" {
		t.Errorf("dst was replaced although the sync failed: %q", b)
	}
}

// N2: a receipt is a run of the card's current epoch.
func TestAReceiptFromAnEarlierEpochIsNoReceipt(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	for _, k := range []string{KindGreen, KindGate} {
		if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: k, SubjectSHA: "s1", Result: ResultPass}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Reap(bg, alwaysDead); err != nil {
		t.Fatal(err)
	}
	c2, err := s.Claim(bg, req("next", 78))
	if err != nil || c2.Epoch != c.Epoch+1 {
		t.Fatalf("%+v %v", c2, err)
	}
	if err := s.Move(bg, "a", c2.Epoch, posReview, "lane", "r"); !errors.Is(err, core.ErrGuardMissing) {
		t.Errorf("the earlier holder's green/gate pair must not count: %v", err)
	}
}

// N4: a row an older build wrote without a subject_sha must not break Move.
func TestAGreenRowWithNoSubjectDoesNotBreakMove(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	if _, err := s.db.ExecContext(bg, `INSERT INTO evidence(card_id,kind,tree_sha,result,epoch,at) VALUES('a','green','t','pass',?,strftime('%Y-%m-%dT%H:%M:%fZ','now'))`, c.Epoch); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "a", c.Epoch, posReview, "lane", "r"); !errors.Is(err, core.ErrGuardMissing) {
		t.Errorf("a NULL subject is a receipt with no accepted gate, not a scan error: %v", err)
	}
}
