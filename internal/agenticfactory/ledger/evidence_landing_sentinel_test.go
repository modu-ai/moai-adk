package ledger

import (
	"errors"
	"testing"
)

// The refusal of a failing run while the row is landing is a named, retryable error: callers tell it from a stale epoch
// or an invalid argument with errors.Is.
func TestLandingWindow_TheRefusalIsTheNamedErrLanding(t *testing.T) {
	s, e, _ := rowInLanding(t)
	err := s.AddEvidence(bg, "a", e, Evidence{Kind: KindGate, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultFail})
	if !errors.Is(err, ErrLanding) {
		t.Fatalf("AddEvidence = %v, want ErrLanding", err)
	}
	if errors.Is(err, ErrStaleEpoch) || errors.Is(err, ErrInvalid) {
		t.Errorf("the refusal is also taken for a stale epoch or an invalid argument: %v", err)
	}
}
