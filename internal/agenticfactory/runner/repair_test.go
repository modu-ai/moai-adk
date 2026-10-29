package runner

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// The repair of an open PR belongs to the agent that published it. A released
// record is not up for grabs: only the responsible agent (its session) may claim the repair, and anyone else is
// refused. The way for another holder to get the Issue is TakeOver, with its
// evidence.
func TestOnlyTheResponsibleAgentMayClaimTheRepair(t *testing.T) {
	r, ref := newRunner(t)
	a1, _ := r.Store.RegisterAgent("aaaaaaaa-0000-0000-0000-000000000001")
	a2, _ := r.Store.RegisterAgent("aaaaaaaa-0000-0000-0000-000000000002")
	tk, err := r.ClaimAs(ref, a1, "disp-1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.StartAttempt(tk, "attempt-1", core.DeliveryExecuted); err != nil {
		t.Fatal(err)
	}
	_, _ = r.Submit(tk, headA, false)
	if _, err := r.Verify(tk, passAllFor(tk, headA)); err != nil {
		t.Fatal(err)
	}
	_, _ = landOK(r, tk, headA)
	if _, err := handOffAndLeave(r, tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	mustReopen(t, r, ref) // the leader path still reopens a released record
	if _, err := r.ClaimRepairAs(ref, a2, "disp-r", "run-2", "WT-agentic-x"); !errors.Is(err, core.ErrNotResponsibleAgent) {
		t.Fatalf("another agent claimed the repair: %v", err)
	}
	if _, err := r.ClaimRepair(ref, "sess-b", "disp-r", "run-2", "WT-agentic-x"); !errors.Is(err, core.ErrNotResponsibleAgent) {
		t.Fatalf("another session claimed the repair: %v", err)
	}
	if cur, _ := r.Store.Read(ref); cur.Holder != "" {
		t.Fatalf("a refused repair claim changed the record: %+v", cur)
	}
	// the session that published the PR may
	if _, err := r.ClaimRepair(ref, a1.SessionID, "disp-r", "run-2", "WT-agentic-x"); err != nil {
		t.Errorf("the original session: %v", err)
	}
}

// After a takeover the new holder is the responsible one: it may repair when the
// record is released, and the old owner may not.
func TestAfterATakeoverTheNewHolderIsTheResponsibleAgent(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	rec, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", fullTakeover())
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Store.Release(ref, "sess-new"); err != nil {
		t.Fatal(err)
	}
	mustReopen(t, r, ref)
	if _, err := r.ClaimRepair(ref, "sess-a", "disp-x", "run-3", goodHandoff().Branch); !errors.Is(err, core.ErrNotResponsibleAgent) {
		t.Errorf("the old owner after a takeover: %v", err)
	}
	if _, err := r.ClaimRepair(ref, "sess-new", "disp-r", "run-3", goodHandoff().Branch); err != nil {
		t.Errorf("the new holder: %v (epoch was %d)", err, rec.Epoch)
	}
}
