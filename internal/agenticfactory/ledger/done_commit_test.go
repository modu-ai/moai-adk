package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func evidenceOn(t *testing.T, s *Store, epoch int64, sha string) {
	t.Helper()
	for _, k := range []string{KindGreen, KindGate} {
		if err := s.AddEvidence(bg, "a", epoch, Evidence{Kind: k, TreeSHA: "tree-" + sha, SubjectSHA: sha, Result: ResultPass}); err != nil {
			t.Fatal(err)
		}
	}
}

// landedThenWithdrawn: c1 lands, then the holder moves the card back out of MERGING.
func landedThenWithdrawn(t *testing.T) (*Store, int64) {
	t.Helper()
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	landOne(t, s, "a", c.Epoch)
	backOut(t, s, c.Epoch)
	return s, c.Epoch
}

// Audit a1ec370195 N2: the merged, closed and settled facts come from the queue row, the
// receipt from the newest green; they must describe the same commit. A card whose newest work is a
// commit that never landed is not done because an older commit did.
func TestMove_DoneNeedsTheLandedCandidateToBeTheNewestWork(t *testing.T) {
	t.Run("a new commit after the landing", func(t *testing.T) {
		s, e := landedThenWithdrawn(t)
		evidenceOn(t, s, e, "c2")
		err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "lane", "done")
		if !errors.Is(err, core.ErrGuardMissing) {
			t.Fatalf("DONE with c2 as the newest work while c1 landed: %v", err)
		}
		if got := cardOf(t, s, "a").Position.Stage; got == core.StageDone {
			t.Error("the card is DONE")
		}
	})
	t.Run("through BUILDING/IMPLEMENT", func(t *testing.T) {
		s, e := landedThenWithdrawn(t)
		if err := s.Move(bg, "a", e, core.Position{Stage: core.StageBuilding, Sub: core.SubImplement}, "lane", "rework"); err != nil {
			t.Fatal(err)
		}
		evidenceOn(t, s, e, "c2")
		err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "lane", "done")
		if !errors.Is(err, core.ErrGuardMissing) {
			t.Fatalf("DONE from BUILDING/IMPLEMENT with c2 as the newest work: %v", err)
		}
	})
	t.Run("a gate on another commit alone does not decide", func(t *testing.T) {
		s, e := landedThenWithdrawn(t)
		if err := s.AddEvidence(bg, "a", e, Evidence{Kind: KindGate, TreeSHA: "tree-c2", SubjectSHA: "c2", Result: ResultPass}); err != nil {
			t.Fatal(err)
		}
		if err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "lane", "done"); err != nil {
			t.Fatalf("a gated but never built second candidate must not block finishing on the landed one: %v", err)
		}
	})
	t.Run("the same candidate still finishes", func(t *testing.T) {
		s, e := landedThenWithdrawn(t)
		if err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "lane", "done"); err != nil {
			t.Fatalf("DONE with c1 still the newest work: %v", err)
		}
	})
	t.Run("more evidence on the landed candidate still finishes", func(t *testing.T) {
		s, e := landedThenWithdrawn(t)
		evidenceOn(t, s, e, "c1")
		if err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "lane", "done"); err != nil {
			t.Fatalf("DONE after fresh evidence on c1: %v", err)
		}
	})
}

// A card built without a green receipt (the no-handoff path: a gate on the candidate is all
// Enqueue asks for) has no newer work to contradict its landed row, so it still finishes.
func TestMove_DoneStillWorksForACardWithOnlyAGateOnItsCandidate(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	landOne(t, s, "a", c.Epoch)
	if err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageMerging, RemoteIntegrated: true}, "mergeq", "merge seen"); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageDone}, "mergeq", "landed"); err != nil {
		t.Fatalf("a card with no green must still finish on its landed candidate: %v", err)
	}
}
