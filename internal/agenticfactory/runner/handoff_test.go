package runner

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func goodHandoff() Handoff {
	return Handoff{PR: 321, Head: headA, Branch: "WT-agentic-x", Worktree: "/wt/x", WriterStopped: true, BoundaryHead: headA}
}

// landed drives an Issue to landing (PR published) and returns its ticket.
func landed(t *testing.T, r *Runner, ref core.IssueRef) Ticket {
	t.Helper()
	tk := toVerifying(t, r, ref)
	if _, err := r.Verify(tk, passAll(headA)); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Land(tk, headA); err != nil {
		t.Fatal(err)
	}
	return tk
}

func TestAttemptSeqCountsAttempts(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	if rec, _ := r.Store.Load(ref); rec.AttemptSeq != 1 {
		t.Errorf("first attempt seq = %d", rec.AttemptSeq)
	}
	v := passAll(headA)
	v.Results[1].Verdict = VerdictFail
	_, _ = r.Verify(tk, v)
	if _, err := r.StartAttempt(tk, "attempt-2", core.DeliveryExecuted); err == nil {
		t.Error("implementing cannot start another attempt")
	}
}

func TestHandoffReleasesClaimKeepsReceipt(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	rec, err := r.HandOff(tk, goodHandoff())
	if err != nil {
		t.Fatal(err)
	}
	if rec.Holder != "" || rec.Receipt == nil || rec.Receipt.PR != 321 || rec.State != core.StateLanding {
		t.Fatalf("after handoff: %+v", rec)
	}
	if _, err := r.Store.Load(ref); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("worker must no longer hold the claim: %v", err)
	}
	// The worker's old ticket can no longer write.
	if _, err := r.Complete(tk, Landing{RemoteMerged: true, LocalSyncRecorded: true, Deploy: "N/A: x"}); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("released worker wrote: %v", err)
	}
	// A different Issue can now be claimed by the same worker.
	other, _ := core.NewIssueRef(2000)
	if _, err := r.Claim(other, "sess-a", "disp-2", "run-1"); err != nil {
		t.Errorf("worker could not take a new issue: %v", err)
	}
}

func TestHandoffRefusedWithoutStopOrBoundaryProof(t *testing.T) {
	bad := map[string]func(*Handoff){
		"writer not stopped":      func(h *Handoff) { h.WriterStopped = false },
		"no pr":                   func(h *Handoff) { h.PR = 0 },
		"no worktree":             func(h *Handoff) { h.Worktree = "" },
		"no branch":               func(h *Handoff) { h.Branch = "" },
		"pr head not candidate":   func(h *Handoff) { h.Head = headB },
		"commit after publish":    func(h *Handoff) { h.BoundaryHead = headB },
		"boundary head malformed": func(h *Handoff) { h.BoundaryHead = "abc" },
	}
	for name, mut := range bad {
		t.Run(name, func(t *testing.T) {
			r, ref := newRunner(t)
			tk := landed(t, r, ref)
			h := goodHandoff()
			mut(&h)
			if _, err := r.HandOff(tk, h); err == nil {
				t.Fatal("want refusal")
			}
			cur, err := r.Store.Load(ref)
			if err != nil || cur.Receipt != nil || cur.Holder != "sess-a" {
				t.Errorf("claim must stay with the worker: %+v %v", cur, err)
			}
		})
	}
}

func TestHandoffOnlyFromLanding(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	if _, err := r.HandOff(tk, goodHandoff()); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("handoff before a PR exists: %v", err)
	}
}

func TestDuplicateHandoffRejected(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	if _, err := r.HandOff(tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	// Same worker retries with its old ticket.
	if _, err := r.HandOff(tk, goodHandoff()); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("retry by released worker: %v", err)
	}
	// A repair holder re-publishing the same head is a duplicate receipt.
	rep, err := r.ClaimRepair(ref, "sess-b", "disp-r", "run-2", "WT-agentic-x")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.HandOff(rep, goodHandoff()); !errors.Is(err, core.ErrTransitionInvalid) && !errors.Is(err, core.ErrDuplicateHandoff) {
		t.Errorf("repair handoff on same head / wrong stage: %v", err)
	}
}

func TestPRPublishedIsNotDone(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	if _, err := r.HandOff(tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	cur, _ := r.Store.Read(ref)
	if cur.State == core.StateDone {
		t.Fatal("PR publication must not be done")
	}
	if _, err := r.CompleteHandedOff(ref, Landing{LocalSyncRecorded: true, Deploy: "N/A: x"}); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("done without remote merge: %v", err)
	}
	rec, err := r.CompleteHandedOff(ref, Landing{RemoteMerged: true, LocalSyncRecorded: true, Deploy: "N/A: x"})
	if err != nil || rec.State != core.StateDone {
		t.Fatalf("complete after merge: %+v %v", rec, err)
	}
	// Issue closure is recorded after done, never before.
	if _, err := r.MarkClosed(ref); err != nil {
		t.Errorf("closure after done: %v", err)
	}
}

func TestMarkClosedRefusedBeforeDone(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	_, _ = r.HandOff(tk, goodHandoff())
	if _, err := r.MarkClosed(ref); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("closure before done: %v", err)
	}
}

func TestCompleteHandedOffNeedsAHandoff(t *testing.T) {
	r, ref := newRunner(t)
	_ = landed(t, r, ref) // worker still holds the claim, no receipt
	if _, err := r.CompleteHandedOff(ref, Landing{RemoteMerged: true, LocalSyncRecorded: true, Deploy: "N/A: x"}); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("complete without receipt: %v", err)
	}
}

func TestRepairIsSingleOwner(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	if _, err := r.HandOff(tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	// A fresh claim on a PR-published Issue is not allowed.
	if _, err := r.Claim(ref, "sess-b", "disp-x", "run-2"); !errors.Is(err, core.ErrRepairOnly) {
		t.Errorf("plain claim on published issue: %v", err)
	}
	// Repair on the wrong branch is refused.
	if _, err := r.ClaimRepair(ref, "sess-b", "disp-r", "run-2", "WT-other"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("wrong branch: %v", err)
	}
	rep, err := r.ClaimRepair(ref, "sess-b", "disp-r", "run-2", "WT-agentic-x")
	if err != nil {
		t.Fatal(err)
	}
	if rep.Epoch != 2 {
		t.Errorf("repair epoch = %d", rep.Epoch)
	}
	// A second repairer is refused while the first holds it.
	if _, err := r.ClaimRepair(ref, "sess-c", "disp-r2", "run-3", "WT-agentic-x"); !errors.Is(err, core.ErrHeld) {
		t.Errorf("contended repair: %v", err)
	}
	// Same dispatch re-arriving is a duplicate, not a second repair.
	if _, err := r.ClaimRepair(ref, "sess-b", "disp-r", "run-2", "WT-agentic-x"); !errors.Is(err, core.ErrDuplicateDispatch) {
		t.Errorf("duplicate repair dispatch: %v", err)
	}
	cur, _ := r.Store.Load(ref)
	if cur.Holder != "sess-b" || cur.Receipt == nil || cur.State != core.StateClaimed {
		t.Errorf("repair record: %+v", cur)
	}
}

func TestRepairRequiresAReceipt(t *testing.T) {
	r, ref := newRunner(t)
	if _, err := r.ClaimRepair(ref, "sess-b", "d", "run-1", "WT-x"); !errors.Is(err, core.ErrNotHeld) && !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("repair with no PR: %v", err)
	}
}

func TestRepairCanRepublishNewHead(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	_, _ = r.HandOff(tk, goodHandoff())
	rep, _ := r.ClaimRepair(ref, "sess-b", "disp-r", "run-2", "WT-agentic-x")
	if _, err := r.StartAttempt(rep, "attempt-2", core.DeliveryExecuted); err != nil {
		t.Fatal(err)
	}
	if rec, _ := r.Store.Load(ref); rec.AttemptSeq != 2 {
		t.Errorf("attempt seq after repair = %d", rec.AttemptSeq)
	}
	if _, err := r.Submit(rep, headB, false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Verify(rep, passAll(headB)); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Land(rep, headB); err != nil {
		t.Fatal(err)
	}
	h := goodHandoff()
	h.Head, h.BoundaryHead = headB, headB
	if rec, err := r.HandOff(rep, h); err != nil || rec.Receipt.Head != core.HeadSHA(headB) || rec.Holder != "" {
		t.Errorf("second handoff: %+v %v", rec, err)
	}
}

func TestBacklogCapHoldsNewClaimsNotRepairs(t *testing.T) {
	r, ref := newRunner(t)
	r.Store.MaxOpenPRs = 1
	tk := landed(t, r, ref)
	if _, err := r.HandOff(tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	other, _ := core.NewIssueRef(2000)
	_, err := r.Claim(other, "sess-a", "disp-2", "run-1")
	if !errors.Is(err, core.ErrBacklogFull) {
		t.Fatalf("want backlog hold, got %v", err)
	}
	if reason, ok := core.ReasonFor(err); !ok || reason != core.ReasonCapacityShort {
		t.Errorf("reason = %q ok=%v", reason, ok)
	}
	if _, err := r.Store.Load(other); !errors.Is(err, core.ErrNotHeld) {
		t.Error("held claim must not leave a record")
	}
	if _, err := r.ClaimRepair(ref, "sess-b", "disp-r", "run-2", "WT-agentic-x"); err != nil {
		t.Errorf("repair must not be blocked by the cap: %v", err)
	}
}

func TestBacklogCapFreedByDone(t *testing.T) {
	r, ref := newRunner(t)
	r.Store.MaxOpenPRs = 1
	tk := landed(t, r, ref)
	_, _ = r.HandOff(tk, goodHandoff())
	if _, err := r.CompleteHandedOff(ref, Landing{RemoteMerged: true, LocalSyncRecorded: true, Deploy: "N/A: x"}); err != nil {
		t.Fatal(err)
	}
	other, _ := core.NewIssueRef(2000)
	if _, err := r.Claim(other, "sess-a", "disp-2", "run-1"); err != nil {
		t.Errorf("slot should be free after done: %v", err)
	}
}

func TestBacklogCapZeroMeansUnlimited(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	_, _ = r.HandOff(tk, goodHandoff())
	other, _ := core.NewIssueRef(2000)
	if _, err := r.Claim(other, "sess-a", "disp-2", "run-1"); err != nil {
		t.Errorf("no cap configured: %v", err)
	}
}

// Leader restart: a new Store/Runner over the same directory restores every
// live and handed-off claim; nothing is re-dispatched.
func TestLeaderRestartRestoresClaims(t *testing.T) {
	dir := t.TempDir()
	st, _ := claim.Open(dir)
	r := New(st)
	ref1, _ := core.NewIssueRef(1948)
	ref2, _ := core.NewIssueRef(2000)
	tk := landed(t, r, ref1)
	if _, err := r.HandOff(tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Claim(ref2, "sess-b", "disp-9", "run-9"); err != nil {
		t.Fatal(err)
	}

	st2, err := claim.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	recs, err := st2.List()
	if err != nil || len(recs) != 2 || recs[0].Issue != 1948 || recs[1].Issue != 2000 {
		t.Fatalf("restored list: %+v %v", recs, err)
	}
	if recs[0].Receipt == nil || recs[0].Holder != "" || recs[1].Holder != "sess-b" {
		t.Errorf("restored ownership: %+v", recs)
	}
	r2 := New(st2)
	if _, err := r2.Claim(ref2, "sess-b", "disp-9", "run-9"); !errors.Is(err, core.ErrDuplicateDispatch) {
		t.Errorf("re-sent dispatch after restart: %v", err)
	}
	if _, err := r2.ClaimRepair(ref1, "sess-c", "disp-r", "run-3", "WT-agentic-x"); err != nil {
		t.Errorf("repair after restart: %v", err)
	}
}
