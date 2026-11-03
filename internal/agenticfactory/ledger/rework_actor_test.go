package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, audit 22a0afd247 N17: the actor of a move is the caller's free text (the CLI
// passes its --actor flag through). The rework check found the moment the candidate was offered by
// reading the actor column for the string the ledger writes on an Enqueue, so a holder that named
// "enqueue" as the actor of a later move made that move look like the offer, and the rework before
// it disappeared: the landed card finished with nothing recorded since it was taken back.
// Which transition is the offer is a fact of the ledger's own structure (only Enqueue enters
// MERGING), never of a string a caller chose.

// Where the caller's actor can land in the history of a card taken back for rework.
var spoofPlaces = []struct {
	name  string
	moves func(t *testing.T, s *Store, e int64, actor string)
}{
	{"on the move out of MERGING", func(t *testing.T, s *Store, e int64, actor string) {
		mustMove(t, s, e, posReviewing, actor)
		mustMove(t, s, e, posImplement, "lane")
	}},
	{"on the move back into BUILDING", func(t *testing.T, s *Store, e int64, actor string) {
		mustMove(t, s, e, posReviewing, "lane")
		mustMove(t, s, e, posImplement, actor)
	}},
	{"on a move inside BUILDING after the move back", func(t *testing.T, s *Store, e int64, actor string) {
		mustMove(t, s, e, posReviewing, "lane")
		mustMove(t, s, e, posImplement, "lane")
		mustMove(t, s, e, posFix, actor)
		mustMove(t, s, e, posReviewing, "lane")
	}},
	{"on the move that hands the card off again", func(t *testing.T, s *Store, e int64, actor string) {
		mustMove(t, s, e, posReviewing, "lane")
		mustMove(t, s, e, posImplement, "lane")
		mustMove(t, s, e, posReviewing, actor)
	}},
}

func mustMove(t *testing.T, s *Store, e int64, to core.Position, actor string) {
	t.Helper()
	if err := s.Move(bg, "a", e, to, actor, "rework"); err != nil {
		t.Fatalf("move to %+v as %q: %v", to, actor, err)
	}
}

func TestRework_TheActorOfAMoveDoesNotSwitchTheReworkCheckOff(t *testing.T) {
	done := core.Position{Stage: core.StageDone}
	for _, actor := range []string{actorEnqueue, actorClaim, actorReap, actorAdd} {
		for _, place := range spoofPlaces {
			t.Run(actor+" "+place.name, func(t *testing.T) {
				// The card is in MERGING with c1 landed (not backed out yet): the spoofed move may be
				// the very first one out of it.
				s := openTemp(t)
				c := gated(t, s, "a", "c1")
				if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
					t.Fatal(err)
				}
				landOne(t, s, "a", c.Epoch)
				place.moves(t, s, c.Epoch, actor)
				err := s.Move(bg, "a", c.Epoch, done, "lane", "finish")
				if !errors.Is(err, core.ErrGuardMissing) {
					t.Fatalf("DONE with nothing recorded since the move back, actor %q %s: %v", actor, place.name, err)
				}
				if got := cardOf(t, s, "a").Position.Stage; got == core.StageDone {
					t.Error("the card is DONE")
				}
				// The control: the same history finishes once c1 is measured again after the move back.
				evidenceOn(t, s, c.Epoch, "c1")
				if err := s.Move(bg, "a", c.Epoch, done, "lane", "finish"); err != nil {
					t.Fatalf("DONE after c1 was measured again (actor %q %s): %v", actor, place.name, err)
				}
			})
		}
	}
}
