package ledger

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func alwaysDead(int, string) Liveness { return LivenessDead }

func TestReapNeedsAJudge(t *testing.T) {
	s := openTemp(t)
	if _, err := s.Reap(bg, nil); !errors.Is(err, ErrInvalid) {
		t.Errorf("Reap(nil): %v", err)
	}
}

func TestReapReclaimsOnlyWhatTheJudgeCallsDead(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "dead", 3, "A", nil, "internal/d/x.go")
	addCard(t, s, "alive", 2, "A", nil, "internal/a/x.go")
	addCard(t, s, "unknown", 1, "A", nil, "internal/u/x.go")
	for i, id := range []string{"dead", "alive", "unknown"} {
		c, err := s.Claim(bg, req(id, 100+i))
		if err != nil || c.CardID != id {
			t.Fatalf("claim %s: %+v %v", id, c, err)
		}
	}
	got, err := s.Reap(bg, func(pid int, _ string) Liveness {
		switch pid {
		case 100:
			return LivenessDead
		case 101:
			return LivenessAlive
		}
		return LivenessUnknown
	})
	if err != nil || len(got) != 1 || got[0].CardID != "dead" || got[0].ReclaimCount != 1 || got[0].Parked || got[0].Epoch != 1 {
		t.Fatalf("reclaimed %+v %v", got, err)
	}
	ls, _ := s.Leases(bg)
	if len(ls) != 2 {
		t.Errorf("only the dead holder's lease goes: %+v", ls)
	}
	var keys int
	if err := s.db.QueryRowContext(bg, `SELECT count(*) FROM lease_keys WHERE card_id='dead'`).Scan(&keys); err != nil || keys != 0 {
		t.Errorf("its keys go with it: %d %v", keys, err)
	}
	if c := cardOf(t, s, "dead"); c.Position.Stage == "" || c.ReclaimCount != 1 {
		t.Errorf("the card stays, counted: %+v", c)
	}
}

// An Unknown holder is never reclaimed, and an expired one gets one stalled note.
func TestAnUnknownHolderGetsOneStalledNoteOnceItsLeaseExpired(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "u", 1, "A", nil, "internal/u/x.go")
	if _, err := s.Claim(bg, req("u", 5)); err != nil {
		t.Fatal(err)
	}
	unknown := func(int, string) Liveness { return LivenessUnknown }
	if got, err := s.Reap(bg, unknown); err != nil || len(got) != 0 {
		t.Fatalf("a live lease: %v %v", got, err)
	}
	notes := func() int {
		var n int
		if err := s.db.QueryRowContext(bg, `SELECT count(*) FROM notes WHERE kind='stalled' AND card_id='u'`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if notes() != 0 {
		t.Error("a lease that has not expired is not stalled")
	}
	if _, err := s.db.ExecContext(bg, `UPDATE leases SET expires_at='2000-01-01T00:00:00.000Z'`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if got, err := s.Reap(bg, unknown); err != nil || len(got) != 0 {
			t.Fatalf("an unknown holder is never reclaimed: %v %v", got, err)
		}
	}
	if notes() != 1 {
		t.Errorf("stalled notes = %d, want one", notes())
	}
	if ls, _ := s.Leases(bg); len(ls) != 1 {
		t.Errorf("the lease stays: %+v", ls)
	}
}

// The judge runs before the write transaction opens (I3). Were a transaction open
// while it ran, this write from another connection would wait out the busy timeout and
// fail; and a holder that renews while it is being judged is not reclaimed.
func TestTheJudgeRunsOutsideTheTransactionAndARenewalSinceTheProbeIsRespected(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
	if _, err := s.Claim(bg, req("a", 7)); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	calls := 0
	judge := func(int, string) Liveness {
		// the holder renews while the probe is looking at it; each renewal is a
		// different instant, as real ones (30 seconds apart) are
		calls++
		if _, err := s.db.ExecContext(bg, `UPDATE leases SET heartbeat_at=strftime('%Y-%m-%dT%H:%M:%fZ','now',?)`, fmt.Sprintf("+%d seconds", calls)); err != nil {
			t.Errorf("a write during the judge failed, so a transaction was open: %v", err)
		}
		return LivenessDead
	}
	got, err := s.Reap(bg, judge)
	if err != nil || len(got) != 0 {
		t.Fatalf("a lease renewed since the probe is not reclaimed: %v %v", got, err)
	}
	if time.Since(start) > 3*time.Second {
		t.Errorf("the judge waited for a lock: %v", time.Since(start))
	}
	if ls, _ := s.Leases(bg); len(ls) != 1 {
		t.Errorf("the lease stays: %+v", ls)
	}
	// the same through Claim's lazy reap
	r := req("b", 8)
	r.Alive = judge
	if _, err := s.Claim(bg, r); !errors.Is(err, ErrNoWork) {
		t.Errorf("Claim's lazy reap respects the renewal too: %v", err)
	}
	if ls, _ := s.Leases(bg); len(ls) != 1 || ls[0].SessionID != "a" {
		t.Errorf("leases: %+v", ls)
	}
}
