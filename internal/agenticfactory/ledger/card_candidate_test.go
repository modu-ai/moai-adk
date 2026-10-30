//go:build afred

package ledger

import (
	"errors"
	"testing"
)

// Card-kind writes name the candidate too: an old verifier of c1 cannot move a row that
// failed and was restarted with c2 inside the same card epoch.
func TestQueueUpdate_CardKindWriteNamesTheVerifiedCandidate(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", ""); err != nil { // old verifier A
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueFailed, "", "red"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, TreeSHA: "t2", SubjectSHA: "c2", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "a", c.Epoch, "c2"); err != nil { // restart inside the same epoch
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c2"), QueueVerifying, "", ""); err != nil { // new verifier B
		t.Fatal(err)
	}
	for name, state := range map[string]QueueState{"failed": QueueFailed, "verifying": QueueVerifying} {
		if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), state, "", "late"); !errors.Is(err, ErrStaleEpoch) {
			t.Errorf("%s by the old verifier: %v", name, err)
		}
		if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, ""), state, "", "late"); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s naming no candidate: %v", name, err)
		}
	}
	if it := rowOf(t, s); it.State != QueueVerifying || it.CandidateSHA != "c2" {
		t.Fatalf("a refused write changed the row: %+v", it)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c2"), QueueVerifying, "", ""); err != nil {
		t.Errorf("the right candidate re-enters verifying: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c2"), QueueFailed, "", "red"); err != nil {
		t.Errorf("the right candidate fails the row: %v", err)
	}
}
