package ledger

import (
	"errors"
	"testing"
)

// issue-1948, audit db0a69698a N28, accepted as a risk by the leader's decision E8 (the shapes were
// the same at the baseline commit). These tests MEASURE what the ledger does today in the three
// shapes, so the report states the outcome of a run and not an inference, and so a change that
// closes one of them is seen as a test to turn round, not as an accident. Each one passes by
// recording the present outcome; none of them says the outcome is wanted.
//
// What the Enqueue rules of N25 and N27 do for them: nothing, because in all three shapes the
// commit that is left out of the trunk is never offered, and both rules sit at the offer.

// Withdrawn to REVIEWING (never through BUILDING), a gate-only new commit that is never offered,
// then DONE on the landing of c1: the card finishes with c2 not in the trunk.
func TestAcceptedRisk_WithdrawnToReviewingThenAGateOnlyNewCommitNeverOffered(t *testing.T) {
	s, e := landedThenWithdrawn(t)
	gateOn(t, s, e, "c2", ResultPass)
	if err := s.Move(bg, "a", e, posDone, "lane", "done"); err != nil {
		t.Fatalf("measured outcome changed (DONE was accepted before): %v", err)
	}
	t.Log("measured: DONE accepted with c2 gated and never offered; the offer rules are not reached")
}

// In that shape the gate-only commit has no route through the queue either: the landed row stays
// until the card has been taken back into BUILDING, so Enqueue refuses c2 and the only way to
// finish from REVIEWING is the DONE above.
func TestAcceptedRisk_WithdrawnToReviewingTheGateOnlyCommitCannotBeOffered(t *testing.T) {
	s, e := landedThenWithdrawn(t)
	gateOn(t, s, e, "c2", ResultPass)
	if err := s.Enqueue(bg, "a", e, "c2"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("measured outcome changed (the offer was refused as landed before): %v", err)
	}
	t.Log("measured: Enqueue refuses c2 (the card has landed and was not taken back into BUILDING)")
}

// A failure recorded after the landing, the card still in MERGING and never taken back: the
// newest results are read only for a card that went back, so the card finishes.
func TestAcceptedRisk_AFailureRecordedAfterTheLandingWithoutAnyRework(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	landOne(t, s, "a", c.Epoch)
	greenOn(t, s, c.Epoch, "c1", ResultFail)
	gateOn(t, s, c.Epoch, "c1", ResultFail)
	if err := finish(s, c.Epoch); err != nil {
		t.Fatalf("measured outcome changed (the card finished before): %v", err)
	}
	t.Log("measured: the card finishes with the newest green and gate of c1 failing, no move back")
}

// Taken back to BUILDING, a gate-only fix c2 that is never offered, and c1 measured again: the
// card finishes on the landing of c1 with the fix out of the trunk. Without the gate on c1 the
// card is refused, which is the part of the shape that is closed.
func TestAcceptedRisk_ReworkProducesAGateOnlyFixThatIsNeverOffered(t *testing.T) {
	s, e := landedThenWithdrawn(t)
	mustMove(t, s, e, posImplement, "lane")
	gateOn(t, s, e, "c2", ResultPass)
	if err := s.Move(bg, "a", e, posDone, "lane", "done"); err == nil {
		t.Fatal("DONE accepted with only c2 recorded since the move back (this part is closed)")
	}
	gateOn(t, s, e, "c1", ResultPass)
	if err := s.Move(bg, "a", e, posDone, "lane", "done"); err != nil {
		t.Fatalf("measured outcome changed (DONE was accepted before): %v", err)
	}
	t.Log("measured: DONE accepted after a gate on c1, with the fix c2 gated and never offered")
}
