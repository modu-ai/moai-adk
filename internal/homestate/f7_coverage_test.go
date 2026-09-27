package homestate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

// A committed refusal is decapped at the transaction boundary: the caller
// sees the inner error, never the committedRefusal wrapper. The wrapper's
// own contract still matters — anything that re-wraps it with %w must keep
// errors.Is working — so both halves are asserted here.
func TestCommittedRefusalBoundary(t *testing.T) {
	inner := fmt.Errorf("card %s: %w", "c1", ErrLeaseExpired)
	wrapped := commitThen(inner)
	if !errors.Is(wrapped, ErrLeaseExpired) {
		t.Fatalf("errors.Is(commitThen(inner), ErrLeaseExpired) = false, want true (Unwrap must chain)")
	}
	if wrapped.Error() != inner.Error() {
		t.Fatalf("Error() = %q, want inner message %q", wrapped.Error(), inner.Error())
	}

	// The production path must surface the decapped error: a refusal on an
	// expired lease reports ErrLeaseExpired and leaks no wrapper.
	db := frOpen(t)
	repo := frNewRepo(t, true)
	frRegisterWorker(t, db, "worker-1")
	c := frLeasedCard(repo, "boundary", CardRun)
	c.EvidenceSHA = repo.Commit
	c.LeaseExpiresAt = frNow.Add(-time.Minute).Format(time.RFC3339Nano)
	frPlace(t, db, c)
	req := frHolderRequest(c, CardSync)
	req.SHA = repo.Commit
	_, err := db.Transition(context.Background(), req)
	if !errors.Is(err, ErrLeaseExpired) {
		t.Fatalf("transition on expired lease: err = %v, want ErrLeaseExpired", err)
	}
	var keep *committedRefusal
	if errors.As(err, &keep) {
		t.Fatalf("committedRefusal escaped the transaction boundary: %v", err)
	}
}

// DefaultOwnerClassifier is the production wiring — the real process probe,
// not an injected one. It must classify this test process (a live pid with
// its own fingerprint) as live, a mismatched fingerprint as dead, and
// degenerate identities as indeterminate.
func TestDefaultOwnerClassifierRealProbeWiring(t *testing.T) {
	if got := DefaultOwnerClassifier(0, "2026-01-01T00:00:00Z"); got != OwnerIndeterminate {
		t.Fatalf("pid 0 = %q, want indeterminate", got)
	}
	if got := DefaultOwnerClassifier(os.Getpid(), "  "); got != OwnerIndeterminate {
		t.Fatalf("blank start = %q, want indeterminate", got)
	}
	fp, state := ProbeProcessIdentity(os.Getpid())
	if state != ProcessIdentityLive {
		t.Skipf("probe state for own pid = %q, cannot assert the live arm here", state)
	}
	if got := DefaultOwnerClassifier(os.Getpid(), fp); got != OwnerLive {
		t.Fatalf("own pid with own fingerprint = %q, want live", got)
	}
	if got := DefaultOwnerClassifier(os.Getpid(), fp+"-other"); got != OwnerDead {
		t.Fatalf("own pid with foreign fingerprint = %q, want dead", got)
	}
}
