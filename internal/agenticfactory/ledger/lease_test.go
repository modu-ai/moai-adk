package ledger

import (
	"errors"
	"testing"
	"time"
)

func claimOne(t *testing.T, s *Store, id string, paths ...string) Claimed {
	t.Helper()
	addCard(t, s, id, 1, "A", nil, paths...)
	c, err := s.Claim(bg, req("s-"+id, 10))
	if err != nil || c.CardID != id {
		t.Fatalf("claim %s: %+v %v", id, c, err)
	}
	return c
}

func TestHeartbeatRenewsFromTheDatabaseClockKeepingTheLeaseLength(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
	r := req("s", 1)
	r.Lease = time.Minute
	c, err := s.Claim(bg, r)
	if err != nil {
		t.Fatal(err)
	}
	// age the lease by a minute
	if _, err := s.db.ExecContext(bg, `UPDATE leases SET heartbeat_at=strftime('%Y-%m-%dT%H:%M:%fZ','now','-60 seconds'), expires_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')`); err != nil {
		t.Fatal(err)
	}
	wake := time.Now().Add(5 * time.Minute).UTC()
	if err := s.Heartbeat(bg, "a", c.Epoch, &wake); err != nil {
		t.Fatal(err)
	}
	ls, _ := s.Leases(bg)
	l := ls[0]
	if time.Since(l.HeartbeatAt) > 5*time.Second {
		t.Errorf("heartbeat %v is not now", l.HeartbeatAt)
	}
	if d := l.ExpiresAt.Sub(l.HeartbeatAt); d < 59*time.Second || d > 61*time.Second {
		t.Errorf("the lease length is kept: %v", d)
	}
	if l.NextWakeAt == nil || l.NextWakeAt.Sub(wake).Abs() > time.Second {
		t.Errorf("next wake %v, want %v", l.NextWakeAt, wake)
	}
	// nextWake is information only: it does not move the expiry
	far := time.Now().Add(240 * time.Hour)
	if err := s.Heartbeat(bg, "a", c.Epoch, &far); err != nil {
		t.Fatal(err)
	}
	ls, _ = s.Leases(bg)
	if d := ls[0].ExpiresAt.Sub(ls[0].HeartbeatAt); d > 61*time.Second {
		t.Errorf("a far next wake moved the expiry: %v", d)
	}
	if err := s.Heartbeat(bg, "a", c.Epoch, nil); err != nil {
		t.Fatal(err)
	}
	ls, _ = s.Leases(bg)
	if ls[0].NextWakeAt != nil {
		t.Errorf("nil clears the next wake: %v", ls[0].NextWakeAt)
	}
}

func TestEveryWriteAfterClaimNeedsTheCurrentEpoch(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	for name, err := range map[string]error{
		"heartbeat": s.Heartbeat(bg, "a", c.Epoch+1, nil),
		"release":   s.Release(bg, "a", c.Epoch+1),
		"expand":    s.ExpandKeys(bg, "a", c.Epoch+1, []string{"pkg:internal/z"}),
		"evidence":  s.AddEvidence(bg, "a", c.Epoch+1, Evidence{Kind: KindRed, TreeSHA: "t", SubjectSHA: "t", Result: ResultFail}),
		"no card":   s.Heartbeat(bg, "nobody", 1, nil),
	} {
		if !errors.Is(err, ErrStaleEpoch) {
			t.Errorf("%s with the wrong epoch: %v", name, err)
		}
	}
	if ls, _ := s.Leases(bg); len(ls) != 1 || len(ls[0].Keys) != 1 {
		t.Errorf("a stale write changed the lease: %+v", ls)
	}
	// a reclaimed and claimed-again card fences the old holder
	if _, err := s.Reap(bg, alwaysDead); err != nil {
		t.Fatal(err)
	}
	c2, err := s.Claim(bg, req("other", 99))
	if err != nil || c2.Epoch != c.Epoch+1 {
		t.Fatalf("%+v %v", c2, err)
	}
	if err := s.Heartbeat(bg, "a", c.Epoch, nil); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("the old holder after a reclaim: %v", err)
	}
	if err := s.Heartbeat(bg, "a", c2.Epoch, nil); err != nil {
		t.Errorf("the new holder: %v", err)
	}
}

func TestReleaseDropsTheLeaseAndKeysAndKeepsTheCard(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	addCard(t, s, "b", 1, "A", nil, "internal/a/y.go")
	if _, err := s.Claim(bg, req("s2", 2)); !errors.Is(err, ErrNoWork) {
		t.Fatalf("b clashes with a: %v", err)
	}
	if err := s.Release(bg, "a", c.Epoch); err != nil {
		t.Fatal(err)
	}
	if ls, _ := s.Leases(bg); len(ls) != 0 {
		t.Errorf("leases: %+v", ls)
	}
	var keys int
	if err := s.db.QueryRowContext(bg, `SELECT count(*) FROM lease_keys`).Scan(&keys); err != nil || keys != 0 {
		t.Errorf("keys %d %v", keys, err)
	}
	if got := cardOf(t, s, "a"); got.ID != "a" || got.Epoch != c.Epoch {
		t.Errorf("the card stays: %+v", got)
	}
	if err := s.Release(bg, "a", c.Epoch); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("a second release: %v", err)
	}
	if c2, err := s.Claim(bg, req("s2", 2)); err != nil || c2.CardID != "a" && c2.CardID != "b" {
		t.Errorf("the keys are free again: %+v %v", c2, err)
	}
}

func TestExpandKeysAddsAllOrNothingAndConflictsParkNothing(t *testing.T) {
	s := openTemp(t)
	a := claimOne(t, s, "a", "internal/a/x.go")
	b := claimOne(t, s, "b", "internal/b/x.go")
	if err := s.ExpandKeys(bg, "a", a.Epoch, []string{"pkg:internal/free", KeyGoMod}); err != nil {
		t.Fatal(err)
	}
	ls, _ := s.Leases(bg)
	if len(ls[0].Keys) != 3 {
		t.Errorf("a's keys: %v", ls[0].Keys)
	}
	if err := s.ExpandKeys(bg, "a", a.Epoch, []string{"pkg:internal/free"}); err != nil {
		t.Errorf("a key it already holds: %v", err)
	}
	// b wants one a holds and one that is free: nothing is added
	if err := s.ExpandKeys(bg, "b", b.Epoch, []string{"pkg:internal/also-free", KeyGoMod}); !errors.Is(err, ErrKeyConflict) {
		t.Fatalf("conflict: %v", err)
	}
	ls, _ = s.Leases(bg)
	if len(ls[1].Keys) != 1 {
		t.Errorf("a refused expansion added keys: %v", ls[1].Keys)
	}
	if got := cardOf(t, s, "b"); got.Position.Park != nil {
		t.Errorf("a conflict parks nothing: %+v", got.Position)
	}
	for _, bad := range [][]string{{""}, {"  "}, {"has space"}} {
		if err := s.ExpandKeys(bg, "b", b.Epoch, bad); !errors.Is(err, ErrInvalid) {
			t.Errorf("%q: %v", bad, err)
		}
	}
	// an expanded key is held against the next claim
	addCard(t, s, "c", 1, "A", nil, "go.sum")
	if _, err := s.Claim(bg, req("s3", 3)); !errors.Is(err, ErrNoWork) {
		t.Errorf("go.sum clashes with the go.mod key a took: %v", err)
	}
}
