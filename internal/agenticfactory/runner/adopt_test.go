package runner

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// parkedForContractChange is a PR under repair (BUILDING/FIX) that stopped
// because the contract changed (canon E24).
func parkedForContractChange(t *testing.T) (*Runner, Ticket) {
	t.Helper()
	r, _, tk := handedOffOwner(t)
	if _, err := r.ReopenOwn(tk, "ci failed on latest head", RepairEvidence{ScopeReserved: true, OldWriterEnded: true}); err != nil {
		t.Fatal(err)
	}
	rec, err := r.Pause(tk, core.StateBlocked, core.ReasonContractChange, "approved revision")
	if err != nil || rec.Stage != core.StageBuilding || rec.Sub != core.SubFix || rec.Park == nil || rec.Park.Reason != core.ReasonContractChange {
		t.Fatalf("setup: %+v %v", rec, err)
	}
	return r, tk
}

// Canon E18 after E24: the approved revision is applied to a NEW attempt, which
// takes the PR back to BUILDING/SPEC and releases the park. Resuming alone (E20)
// leaves the stage where it was.
func TestAdoptingTheApprovedRevisionStartsANewAttemptInSpec(t *testing.T) {
	r, tk := parkedForContractChange(t)
	before, _ := r.Store.Read(tk.Issue)
	if _, err := r.AdoptRevision(tk, ContractRevision{Diff: "", Approved: "rev-2"}); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("the contract diff is required: %v", err)
	}
	if _, err := r.AdoptRevision(tk, ContractRevision{Diff: "spec.md §3 changed", Approved: ""}); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("an unapproved revision is not adopted: %v", err)
	}
	if cur, _ := r.Store.Read(tk.Issue); cur.Seq != before.Seq || cur.Park == nil {
		t.Fatalf("a refused adoption changed the record: %+v", cur)
	}
	rec, err := r.AdoptRevision(tk, ContractRevision{Diff: "spec.md §3 changed", Approved: "rev-2"})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Stage != core.StageBuilding || rec.Sub != core.SubSpec || rec.Park != nil || rec.State != core.StateClaimed {
		t.Errorf("adoption lands on BUILDING/SPEC, unparked: %+v", rec)
	}
	if rec.Attempt != "" || rec.Receipt == nil || rec.Revision != "rev-2" || rec.Repair != nil || rec.Holder != tk.Holder {
		t.Errorf("a fresh attempt on the same PR and owner: %+v", rec)
	}
	next, err := r.StartAttempt(tk, "attempt-2", core.DeliveryExecuted)
	if err != nil || next.State != core.StateImplementing || next.Sub != core.SubImplement {
		t.Fatalf("the new attempt then implements the revised spec: %+v %v", next, err)
	}
	if next.AttemptSeq <= before.AttemptSeq {
		t.Errorf("attempt sequence must advance: %d -> %d", before.AttemptSeq, next.AttemptSeq)
	}
}

func TestResumeAloneKeepsTheFixStage(t *testing.T) {
	r, tk := parkedForContractChange(t)
	rec, err := r.Resume(tk, true)
	if err != nil || rec.Park != nil || rec.Stage != core.StageBuilding || rec.Sub != core.SubFix {
		t.Errorf("E20 releases the park and keeps FIX: %+v %v", rec, err)
	}
}

func TestAdoptionIsRefusedWithoutAContractChangePark(t *testing.T) {
	r, _, tk := handedOffOwner(t)
	if _, err := r.AdoptRevision(tk, ContractRevision{Diff: "d", Approved: "rev-2"}); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("an unparked PR: %v", err)
	}
	if _, err := r.ReopenOwn(tk, "x", RepairEvidence{ScopeReserved: true, OldWriterEnded: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Pause(tk, core.StateBlocked, core.ReasonRateLimited, "later"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.AdoptRevision(tk, ContractRevision{Diff: "d", Approved: "rev-2"}); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("parked on another cause: %v", err)
	}
}
