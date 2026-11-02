package ledger

import (
	"testing"
)

// issue-1948: a commit that was offered to the queue and failed or was withdrawn, after which the card went back into
// the build stage.

// What Enqueue accepts: a candidate the card never offered before (REQ-AFL-008, decision D5), a card not
// taken back (a flaky retry of the same commit), and a commit the card was taken back for and has
// since measured.
func TestEnqueue_NoReworkMarkNoNewDemand(t *testing.T) {
	t.Run("a gate-only candidate never offered before, after a withdrawn c1", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		e := c.Epoch
		if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
			t.Fatal(err)
		}
		mustMove(t, s, e, posReviewing, "lane")
		mustMove(t, s, e, posImplement, "lane")
		gateOn(t, s, e, "c2", ResultPass)
		mustMove(t, s, e, posReviewing, "lane")
		if err := s.Enqueue(bg, "a", e, "c2"); err != nil {
			t.Fatalf("a gate-only candidate is refused (D5): %v", err)
		}
	})
	t.Run("the same commit again after a failure, the card never taken back", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		e := c.Epoch
		if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
			t.Fatal(err)
		}
		if err := s.QueueUpdate(bg, "a", CardEpoch(e, "c1"), QueueFailed, "", "flaky gate"); err != nil {
			t.Fatal(err)
		}
		mustMove(t, s, e, posReviewing, "lane")
		if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
			t.Fatalf("a retry of the same commit with no rework is refused: %v", err)
		}
	})
	t.Run("withdrawn to REVIEWING and offered again, never back in BUILDING", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		e := c.Epoch
		if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
			t.Fatal(err)
		}
		mustMove(t, s, e, posReviewing, "lane")
		if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
			t.Fatalf("c1 withdrawn and offered again is refused: %v", err)
		}
	})
}
