package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, audit 22a0afd247 N14 (decision D5): a candidate that carries only a gate is accepted by
// Enqueue (REQ-AFL-008), so one that lands must be able to finish the card without a green being
// added afterwards. The newest-work check compared the landed candidate with the card's newest
// GREEN evidence, which for a gate-only c2 offered after a green on c1 is still c1, and the card was
// stuck in MERGING (E12 needs merged) until the holder recorded a green that no one asked for. The
// check stands for what it was written for: a green that is newer than the offer of the landed
// candidate and names another commit is newer work the landing does not hold.

func greenOn(t *testing.T, s *Store, e int64, sha, result string) {
	t.Helper()
	if err := s.AddEvidence(bg, "a", e, Evidence{Kind: KindGreen, TreeSHA: "tree-" + sha, SubjectSHA: sha, Result: result}); err != nil {
		t.Fatal(err)
	}
}

func gateOn(t *testing.T, s *Store, e int64, sha, result string) {
	t.Helper()
	if err := s.AddEvidence(bg, "a", e, Evidence{Kind: KindGate, TreeSHA: "tree-" + sha, SubjectSHA: sha, Result: result}); err != nil {
		t.Fatal(err)
	}
}

func countGreen(t *testing.T, s *Store, sha string) int {
	t.Helper()
	var n int
	if err := s.db.QueryRowContext(bg, `SELECT count(*) FROM evidence WHERE card_id='a' AND kind=? AND subject_sha=?`, KindGreen, sha).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// c1 has its green and gate and is offered; the holder withdraws it before it lands and offers a
// gate-only c2 in its place, which lands.
func gateOnlySecondCandidate(t *testing.T) (*Store, int64) {
	t.Helper()
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	e := c.Epoch
	if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
		t.Fatal(err)
	}
	mustMove(t, s, e, posReviewing, "lane")
	gateOn(t, s, e, "c2", ResultPass)
	if err := s.Enqueue(bg, "a", e, "c2"); err != nil {
		t.Fatalf("a gate-only candidate is refused at Enqueue (REQ-AFL-008 accepts it): %v", err)
	}
	landOne(t, s, "a", e)
	return s, e
}

func TestDone_AGateOnlyCandidateThatLandedFinishesTheCardWithoutAnotherGreen(t *testing.T) {
	t.Run("c2 offered after c1 was withdrawn", func(t *testing.T) {
		s, e := gateOnlySecondCandidate(t)
		if err := finish(s, e); err != nil {
			t.Fatalf("c2 landed with a passing gate and the card cannot finish: %v", err)
		}
		if got := cardOf(t, s, "a").Position.Stage; got != core.StageDone {
			t.Errorf("stage %s", got)
		}
		if n := countGreen(t, s, "c2"); n != 0 {
			t.Errorf("the test must not have added a green for c2: %d", n)
		}
	})
	t.Run("the holder is gone after the landing", func(t *testing.T) {
		s, e := gateOnlySecondCandidate(t)
		if err := s.Release(bg, "a", e); err != nil {
			t.Fatal(err)
		}
		// The row of the card's own epoch is landed: it, not a lease, is the authority for settling.
		if err := finish(s, e); err != nil {
			t.Fatalf("no lease, c2 landed with a passing gate: %v", err)
		}
		if got := cardOf(t, s, "a").Position.Stage; got != core.StageDone {
			t.Errorf("stage %s", got)
		}
	})
	t.Run("c2 offered over the landed c1 after the card was taken back", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		e := c.Epoch
		if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
			t.Fatal(err)
		}
		landOne(t, s, "a", e)
		mustMove(t, s, e, posReviewing, "lane")
		mustMove(t, s, e, posImplement, "lane")
		gateOn(t, s, e, "c2", ResultPass)
		if err := s.Enqueue(bg, "a", e, "c2"); err != nil {
			t.Fatal(err)
		}
		landOne(t, s, "a", e)
		if err := finish(s, e); err != nil {
			t.Fatalf("the rework produced a gate-only c2 and it landed: %v", err)
		}
	})
}

// What stays refused: a green newer than the offer of the landed candidate that names another
// commit is newer work (N2). The gate-only rule does not weaken it.
func TestDone_NewerWorkThanTheLandedCandidateStaysRefused(t *testing.T) {
	cases := []struct {
		name string
		do   func(t *testing.T, s *Store, e int64)
	}{
		{"a green on a third commit after the gate-only candidate landed", func(t *testing.T, s *Store, e int64) {
			greenOn(t, s, e, "c3", ResultPass)
		}},
		{"a green and a gate on a third commit after it landed", func(t *testing.T, s *Store, e int64) {
			evidenceOn(t, s, e, "c3")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, e := gateOnlySecondCandidate(t)
			tc.do(t, s, e)
			if err := finish(s, e); !errors.Is(err, core.ErrGuardMissing) {
				t.Fatalf("the card finished with newer work than the landed candidate: %v", err)
			}
		})
	}
	t.Run("c1 landed, then a new commit measured and never offered", func(t *testing.T) {
		s, e := landedThenWithdrawn(t)
		evidenceOn(t, s, e, "c2")
		if err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
			t.Fatalf("DONE with c2 as the newest work while c1 landed: %v", err)
		}
	})
	t.Run("through BUILDING, a new commit measured and never offered", func(t *testing.T) {
		s, e := landedThenWithdrawn(t)
		mustMove(t, s, e, posImplement, "lane")
		evidenceOn(t, s, e, "c2")
		if err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
			t.Fatalf("DONE from BUILDING/IMPLEMENT with c2 as the newest work: %v", err)
		}
	})
}
