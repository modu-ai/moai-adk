package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// Audit a1ec370195 N2, third item (issue-1948): a card whose candidate landed and which was
// then moved back to BUILDING/IMPLEMENT records no evidence at all, and used to finish through
// the no-handoff-finished composition on the strength of the landed row alone. Nothing in the
// record says the rework was done, so the landed row cannot stand in for it: a card back in
// the build stage is not done until it leaves that stage by the receipt path.
func TestMove_DoneRefusedForALandedCardBackInBuilding(t *testing.T) {
	t.Run("BUILDING/IMPLEMENT with no new evidence", func(t *testing.T) {
		s, e := landedThenWithdrawn(t)
		if err := s.Move(bg, "a", e, core.Position{Stage: core.StageBuilding, Sub: core.SubImplement}, "lane", "rework"); err != nil {
			t.Fatal(err)
		}
		err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "lane", "done")
		if !errors.Is(err, core.ErrGuardMissing) {
			t.Fatalf("DONE from BUILDING/IMPLEMENT on a landed row alone: %v", err)
		}
		if got := cardOf(t, s, "a").Position.Stage; got == core.StageDone {
			t.Error("the card is DONE")
		}
	})
	t.Run("a card that never left REVIEWING still finishes", func(t *testing.T) {
		s, e := landedThenWithdrawn(t)
		if err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "lane", "done"); err != nil {
			t.Fatalf("control: DONE from REVIEWING on the landed candidate: %v", err)
		}
	})
}
