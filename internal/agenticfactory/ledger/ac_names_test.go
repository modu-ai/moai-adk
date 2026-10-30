package ledger

import (
	"errors"
	"reflect"
	"testing"
)

const longAgo = `2000-01-01T00:00:00.000Z`

func stalledNoteCount(t *testing.T, s *Store, cardID string) int {
	t.Helper()
	var n int
	if err := s.db.QueryRowContext(bg, `SELECT count(*) FROM notes WHERE kind='stalled' AND card_id=?`, cardID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func countOf(t *testing.T, s *Store, query string, args ...any) int {
	t.Helper()
	var n int
	if err := s.db.QueryRowContext(bg, query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// AC-AFL-005: three expired leases, one verdict each; only the dead holder's is reclaimed.
func TestReap_LivenessCases_ReclaimOnlyDead(t *testing.T) {
	s := openTemp(t)
	for i, id := range []string{"live", "unknown", "dead"} {
		addCard(t, s, id, 3-i, "A", nil, "internal/"+id+"/x.go")
		if c, err := s.Claim(bg, req(id, 100+i)); err != nil || c.CardID != id {
			t.Fatalf("claim %s: %+v %v", id, c, err)
		}
	}
	if _, err := s.db.ExecContext(bg, `UPDATE leases SET expires_at=?`, longAgo); err != nil {
		t.Fatal(err)
	}
	got, err := s.Reap(bg, func(pid int, _ string) Liveness {
		switch pid {
		case 100:
			return LivenessAlive
		case 101:
			return LivenessUnknown
		}
		return LivenessDead
	})
	if err != nil || len(got) != 1 || got[0].CardID != "dead" {
		t.Fatalf("reclaimed %+v %v", got, err)
	}
	for _, id := range []string{"live", "unknown"} {
		if countOf(t, s, `SELECT count(*) FROM leases WHERE card_id=?`, id) != 1 || countOf(t, s, `SELECT count(*) FROM lease_keys WHERE card_id=?`, id) != 1 {
			t.Errorf("%s: lease or key gone", id)
		}
		if c := cardOf(t, s, id); c.ReclaimCount != 0 {
			t.Errorf("%s: reclaim_count %d", id, c.ReclaimCount)
		}
		if n := stalledNoteCount(t, s, id); n != 1 {
			t.Errorf("%s: %d stalled notes, want 1", id, n)
		}
	}
	if countOf(t, s, `SELECT count(*) FROM leases WHERE card_id='dead'`) != 0 || countOf(t, s, `SELECT count(*) FROM lease_keys WHERE card_id='dead'`) != 0 {
		t.Error("dead: lease or key stayed")
	}
	if c := cardOf(t, s, "dead"); c.ReclaimCount != 1 {
		t.Errorf("dead: reclaim_count %d", c.ReclaimCount)
	}
}

// N1: the lazy reap inside Claim leaves a live holder's lease alone, notes it once, and
// still hands out other work.
func TestClaim_LazyReap_DoesNotReclaimLiveOwner(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "held", 9, "A", nil, "internal/held/x.go")
	addCard(t, s, "other", 1, "A", nil, "internal/other/x.go")
	if c, err := s.Claim(bg, req("owner", 111)); err != nil || c.CardID != "held" {
		t.Fatalf("%+v %v", c, err)
	}
	if _, err := s.db.ExecContext(bg, `UPDATE leases SET expires_at=?`, longAgo); err != nil {
		t.Fatal(err)
	}
	r := req("s2", 222)
	r.Alive = func(pid int, _ string) Liveness {
		if pid == 111 || pid == 222 { // the second Claim's own lease is alive too
			return LivenessAlive
		}
		return LivenessDead
	}
	c, err := s.Claim(bg, r)
	if err != nil || c.CardID != "other" {
		t.Fatalf("the other card is still claimable: %+v %v", c, err)
	}
	if _, err := s.Claim(bg, r); !errors.Is(err, ErrNoWork) {
		t.Fatalf("nothing left: %v", err)
	}
	ls, _ := s.Leases(bg)
	held := 0
	for _, l := range ls {
		if l.CardID == "held" && l.SessionID == "owner" && l.Epoch == 1 {
			held++
		}
	}
	if held != 1 || countOf(t, s, `SELECT count(*) FROM lease_keys WHERE card_id='held'`) != 1 {
		t.Errorf("the live owner's lease or key was taken: %+v", ls)
	}
	if c := cardOf(t, s, "held"); c.ReclaimCount != 0 || c.Epoch != 1 {
		t.Errorf("held: %+v", c)
	}
	if n := stalledNoteCount(t, s, "held"); n != 1 {
		t.Errorf("stalled notes %d after two Claims, want exactly one", n)
	}
}

// AC-AFL-002: epoch 1 writes after the card was claimed again at epoch 2 change nothing.
func TestStore_StaleEpochWrite_ReturnsErrStaleEpoch(t *testing.T) {
	s := openTemp(t)
	c1 := claimOne(t, s, "a", "internal/a/x.go")
	if _, err := s.Reap(bg, alwaysDead); err != nil {
		t.Fatal(err)
	}
	c2, err := s.Claim(bg, req("s2", 2))
	if err != nil || c2.Epoch != c1.Epoch+1 {
		t.Fatalf("%+v %v", c2, err)
	}
	before := cardOf(t, s, "a")
	transitions := countOf(t, s, `SELECT count(*) FROM transitions WHERE card_id='a'`)
	for name, err := range map[string]error{
		"move":      s.Move(bg, "a", c1.Epoch, posReview, "lane", "stale", Evidence{Kind: KindGreen, TreeSHA: "t", SubjectSHA: "t", Result: ResultPass}),
		"evidence":  s.AddEvidence(bg, "a", c1.Epoch, Evidence{Kind: KindRed, TreeSHA: "t", SubjectSHA: "t", Result: ResultFail}),
		"heartbeat": s.Heartbeat(bg, "a", c1.Epoch, nil),
	} {
		if !errors.Is(err, ErrStaleEpoch) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if after := cardOf(t, s, "a"); !reflect.DeepEqual(before.Position, after.Position) || before.Epoch != after.Epoch || before.ReclaimCount != after.ReclaimCount {
		t.Errorf("the card changed: %+v -> %+v", before, after)
	}
	if countOf(t, s, `SELECT count(*) FROM evidence WHERE card_id='a'`) != 0 || countOf(t, s, `SELECT count(*) FROM transitions WHERE card_id='a'`) != transitions {
		t.Error("a stale write left a row behind")
	}
}

// N2 (fencing, I2): the old holder's Release cannot delete the new holder's lease or keys.
func TestRelease_StaleEpoch_KeepsNewLeaseAndKeys(t *testing.T) {
	s := openTemp(t)
	c1 := claimOne(t, s, "a", "internal/a/x.go")
	if _, err := s.Reap(bg, alwaysDead); err != nil {
		t.Fatal(err)
	}
	c2, err := s.Claim(bg, req("s2", 2))
	if err != nil || c2.Epoch != c1.Epoch+1 {
		t.Fatalf("%+v %v", c2, err)
	}
	if err := s.Release(bg, "a", c1.Epoch); !errors.Is(err, ErrStaleEpoch) {
		t.Fatalf("the old holder's release: %v", err)
	}
	ls, _ := s.Leases(bg)
	if len(ls) != 1 || ls[0].SessionID != "s2" || ls[0].Epoch != c2.Epoch || len(ls[0].Keys) != 1 {
		t.Errorf("the new holder's lease changed: %+v", ls)
	}
	if countOf(t, s, `SELECT count(*) FROM lease_keys WHERE card_id='a'`) != 1 {
		t.Error("the new holder's key was deleted")
	}
	if err := s.Release(bg, "a", c2.Epoch); err != nil {
		t.Errorf("the new holder releases its own: %v", err)
	}
}

// N14 (REQ-AFL-014): waiting, verifying and failed writes carry the card epoch the row was
// queued at; landing and landed carry the landing lock's epoch.
func TestQueueUpdate_EpochSource_FencesByKind(t *testing.T) {
	s := openTemp(t)
	// warm-up: the merge lock's epoch is not the card's epoch below
	if e, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "warm"}, LandingLockTTL); err != nil {
		t.Fatal(err)
	} else if err := s.ReleaseLock(bg, LockMerge, e); err != nil {
		t.Fatal(err)
	}
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	lock, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL)
	if err != nil || lock == c.Epoch {
		t.Fatalf("lock epoch %d must differ from card epoch %d: %v", lock, c.Epoch, err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch+1), QueueVerifying, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("verifying with another card epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(lock), QueueVerifying, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("verifying with the lock epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueVerifying, "", ""); err != nil {
		t.Fatalf("verifying with the row's card epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(c.Epoch), QueueLanding, "t0", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("landing with the card epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock+1), QueueLanding, "t0", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("landing with another lock epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock), QueueLanding, "t0", ""); err != nil {
		t.Fatalf("landing with the lock epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(c.Epoch), QueueLanded, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("landed with the card epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock), QueueLanded, "", ""); err != nil {
		t.Errorf("landed with the lock epoch: %v", err)
	}
	// a row of an older card epoch belongs to a fenced holder, whatever epoch it brings
	if _, err := s.db.ExecContext(bg, `UPDATE cards SET epoch=epoch+1 WHERE id='a'`); err != nil {
		t.Fatal(err)
	}
	for name, ep := range map[string]int64{"card": c.Epoch, "lock": lock} {
		if err := s.QueueUpdate(bg, "a", CardEpoch(ep), QueueFailed, "", "late"); !errors.Is(err, ErrStaleEpoch) {
			t.Errorf("a write for the old card epoch with the %s epoch: %v", name, err)
		}
	}
}
