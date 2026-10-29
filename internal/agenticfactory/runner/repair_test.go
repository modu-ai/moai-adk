package runner

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
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

// A repair is a new attempt (spec, design E06): the owner's reopen takes the next
// attempt number and a fresh attempt id, so the old attempt's receipt and the
// repair's can be told apart.
func TestReopenOwnStartsANewAttempt(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	before, _ := r.Store.Read(ref)
	if before.AttemptSeq != 1 || before.Attempt != "attempt-1" {
		t.Fatalf("setup: %q seq %d", before.Attempt, before.AttemptSeq)
	}
	rec, err := r.ReopenOwn(tk, "ci failed on latest head", RepairEvidence{ScopeReserved: true, OldWriterEnded: true})
	if err != nil {
		t.Fatal(err)
	}
	if rec.AttemptSeq != before.AttemptSeq+1 || rec.Attempt == "" || rec.Attempt == before.Attempt {
		t.Errorf("a repair is a new attempt: %q seq %d (was %q seq %d)", rec.Attempt, rec.AttemptSeq, before.Attempt, before.AttemptSeq)
	}
	if _, err := core.ParseAttemptID(string(rec.Attempt)); err != nil {
		t.Errorf("the attempt id must be a valid id: %v", err)
	}
	if rec.Receipt == nil || rec.Receipt.Attempt != before.Attempt || rec.Receipt.AttemptSeq != before.AttemptSeq {
		t.Errorf("the earlier receipt keeps its own attempt: %+v", rec.Receipt)
	}
	// a second reopen is a third attempt
	if _, err := r.Store.Update(ref, tk.Holder, tk.Epoch, func(x *claim.Record) error { x.State = core.StateLanding; return nil }); err != nil {
		t.Fatal(err)
	}
	again, err := r.ReopenOwn(tk, "review asked for changes", RepairEvidence{ScopeReserved: true, OldWriterEnded: true})
	if err != nil || again.AttemptSeq != before.AttemptSeq+2 || again.Attempt == rec.Attempt {
		t.Errorf("third attempt: %q seq %d %v", again.Attempt, again.AttemptSeq, err)
	}
}
