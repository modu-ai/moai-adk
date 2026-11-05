package ledger

import (
	"testing"
)

// issue-1948: a risk card goes through Enqueue with a passing review of the candidate.

func reviewOn(t *testing.T, s *Store, e int64, sha, result string) {
	t.Helper()
	if err := s.AddEvidence(bg, "a", e, Evidence{Kind: KindReview, TreeSHA: "tree-" + sha, SubjectSHA: sha, Result: result}); err != nil {
		t.Fatal(err)
	}
}

// A risk card's own flow is untouched: green, gate, then the review of the same commit is the
// newest evidence, and it is taken.
func TestEnqueue_ARiskCardWhoseNewestEvidenceIsItsOwnReviewIsStillTaken(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1", "security")
	reviewOn(t, s, c.Epoch, "c1", ResultPass)
	mustMove(t, s, c.Epoch, posReviewing, "lane")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatalf("green, gate and review of c1: %v", err)
	}
}
