package ledger

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// inject makes the database refuse one kind of write on one table, the way a full disk or a
// locked file would, so the transaction that issues it fails half-way.
func inject(t *testing.T, s *Store, event, table string) {
	t.Helper()
	q := fmt.Sprintf(`CREATE TRIGGER inject_%s_%s BEFORE %s ON %s BEGIN SELECT RAISE(ABORT, 'injected'); END`, event, table, event, table)
	if _, err := s.db.ExecContext(bg, q); err != nil {
		t.Fatal(err)
	}
}

func wantInjected(t *testing.T, what string, err error) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), "injected") {
		t.Errorf("%s: want the injected database failure, got %v", what, err)
	}
}

func rowCount(t *testing.T, s *Store, table string) int {
	t.Helper()
	return countOf(t, s, `SELECT count(*) FROM `+table)
}

// When one statement of an operation fails, the operation fails as a whole and leaves the
// ledger as it found it: no half-moved card, no row without its history.
func TestAWriteThatFailsHalfWayLeavesNothingBehind(t *testing.T) {
	t.Run("Move: the transition row", func(t *testing.T) {
		s := openTemp(t)
		c := claimOne(t, s, "a", "internal/a/x.go")
		for _, k := range []string{KindGreen, KindGate} {
			if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: k, TreeSHA: "t", SubjectSHA: "c1", Result: ResultPass}); err != nil {
				t.Fatal(err)
			}
		}
		before := cardOf(t, s, "a").Position
		inject(t, s, "INSERT", "transitions")
		wantInjected(t, "Move", s.Move(bg, "a", c.Epoch, posReview, "lane", "r"))
		if got := cardOf(t, s, "a").Position; got.Stage != before.Stage || got.Sub != before.Sub {
			t.Errorf("the card moved although its history could not be written: %+v -> %+v", before, got)
		}
	})
	t.Run("Move: the evidence row", func(t *testing.T) {
		s := openTemp(t)
		c := claimOne(t, s, "a", "internal/a/x.go")
		inject(t, s, "INSERT", "evidence")
		wantInjected(t, "Move", s.Move(bg, "a", c.Epoch, posBuilding, "lane", "r", Evidence{Kind: KindGate, TreeSHA: "t", SubjectSHA: "c1", Result: ResultPass}))
	})
	t.Run("Enqueue: the queue row", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		inject(t, s, "INSERT", "merge_queue")
		wantInjected(t, "Enqueue", s.Enqueue(bg, "a", c.Epoch, "c1"))
		if got := cardOf(t, s, "a").Position.Stage; got == "MERGING" {
			t.Error("the card entered MERGING without a queue row")
		}
	})
	t.Run("Enqueue: the transition row", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		inject(t, s, "INSERT", "transitions")
		wantInjected(t, "Enqueue", s.Enqueue(bg, "a", c.Epoch, "c1"))
		if n := rowCount(t, s, "merge_queue"); n != 0 {
			t.Errorf("a queue row stayed although the card could not be moved: %d", n)
		}
	})
	t.Run("Enqueue: the note about a replaced row", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueFailed, "", "red"); err != nil {
			t.Fatal(err)
		}
		if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, TreeSHA: "t2", SubjectSHA: "c2", Result: ResultPass}); err != nil {
			t.Fatal(err)
		}
		inject(t, s, "INSERT", "notes")
		wantInjected(t, "Enqueue", s.Enqueue(bg, "a", c.Epoch, "c2"))
		if it := rowOf(t, s); it.CandidateSHA != "c1" || it.State != QueueFailed {
			t.Errorf("the failed row was replaced without its note: %+v", it)
		}
	})
	t.Run("QueueUpdate: the row", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		inject(t, s, "UPDATE", "merge_queue")
		wantInjected(t, "QueueUpdate", s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", ""))
		if it := rowOf(t, s); it.State != QueueWaiting {
			t.Errorf("the row moved: %+v", it)
		}
	})
	t.Run("Claim: the lease", func(t *testing.T) {
		s := openTemp(t)
		addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
		inject(t, s, "INSERT", "leases")
		_, err := s.Claim(bg, req("s", 1))
		wantInjected(t, "Claim", err)
		if c := cardOf(t, s, "a"); c.Epoch != 0 {
			t.Errorf("the epoch rose without a lease: %d", c.Epoch)
		}
	})
	t.Run("Claim: the keys", func(t *testing.T) {
		s := openTemp(t)
		addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
		inject(t, s, "INSERT", "lease_keys")
		_, err := s.Claim(bg, req("s", 1))
		wantInjected(t, "Claim", err)
		if n := rowCount(t, s, "leases"); n != 0 {
			t.Errorf("a lease stayed without its keys: %d", n)
		}
	})
	t.Run("Claim: the first moves", func(t *testing.T) {
		s := openTemp(t)
		addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
		before := rowCount(t, s, "transitions")
		inject(t, s, "INSERT", "transitions")
		_, err := s.Claim(bg, req("s", 1))
		wantInjected(t, "Claim", err)
		if n := rowCount(t, s, "leases"); n != 0 || rowCount(t, s, "transitions") != before {
			t.Errorf("a half-made claim stayed: %d leases", n)
		}
	})
	t.Run("AddEvidence", func(t *testing.T) {
		s := openTemp(t)
		c := claimOne(t, s, "a", "internal/a/x.go")
		inject(t, s, "INSERT", "evidence")
		wantInjected(t, "AddEvidence", s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, TreeSHA: "t", SubjectSHA: "c1", Result: ResultPass}))
	})
	t.Run("AddNote", func(t *testing.T) {
		s := openTemp(t)
		inject(t, s, "INSERT", "notes")
		wantInjected(t, "AddNote", s.AddNote(bg, Note{Kind: NoteDecision, Text: "x"}))
	})
	t.Run("Heartbeat", func(t *testing.T) {
		s := openTemp(t)
		c := claimOne(t, s, "a", "internal/a/x.go")
		inject(t, s, "UPDATE", "leases")
		wantInjected(t, "Heartbeat", s.Heartbeat(bg, "a", c.Epoch, nil))
	})
	t.Run("ExpandKeys", func(t *testing.T) {
		s := openTemp(t)
		c := claimOne(t, s, "a", "internal/a/x.go")
		inject(t, s, "INSERT", "lease_keys")
		wantInjected(t, "ExpandKeys", s.ExpandKeys(bg, "a", c.Epoch, []string{"pkg:internal/z"}))
	})
	t.Run("Release: the keys", func(t *testing.T) {
		s := openTemp(t)
		c := claimOne(t, s, "a", "internal/a/x.go")
		inject(t, s, "DELETE", "lease_keys")
		wantInjected(t, "Release", s.Release(bg, "a", c.Epoch))
		if n := rowCount(t, s, "leases"); n != 1 {
			t.Errorf("the lease went although its keys could not: %d", n)
		}
	})
	t.Run("Reap: the reclaim count", func(t *testing.T) {
		s := openTemp(t)
		claimOne(t, s, "a", "internal/a/x.go")
		inject(t, s, "UPDATE", "cards")
		_, err := s.Reap(bg, alwaysDead)
		wantInjected(t, "Reap", err)
		if n := rowCount(t, s, "leases"); n != 1 {
			t.Errorf("the dead holder's lease went although the card could not be updated: %d", n)
		}
	})
	t.Run("AcquireLock", func(t *testing.T) {
		s := openTemp(t)
		inject(t, s, "INSERT", "locks")
		_, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "x"}, time.Minute)
		wantInjected(t, "AcquireLock", err)
	})
	t.Run("ReleaseLock and RenewLock", func(t *testing.T) {
		s := openTemp(t)
		e, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "x"}, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		inject(t, s, "UPDATE", "locks")
		wantInjected(t, "RenewLock", s.RenewLock(bg, LockMerge, e, time.Minute))
		wantInjected(t, "ReleaseLock", s.ReleaseLock(bg, LockMerge, e))
		_, err = s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "y"}, time.Minute)
		if err == nil {
			t.Error("the failed release must leave the lock held")
		}
	})
}
