package runner

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func fullTakeover() Takeover {
	return Takeover{OwnerGone: true, WriterStopped: true, UncommittedPreserved: true, PreservedRef: "refs/heads/WT-agentic-x@preserve-1"}
}

// A dead owner's Issue is reassigned only after three facts are in hand: the
// owner's session is gone, its writer has stopped, and what it left uncommitted
// is preserved. Short of any one, nothing moves.
func TestTakeoverNeedsTheOwnerGoneTheWriterStoppedAndTheWorkPreserved(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	for name, mut := range map[string]func(*Takeover){
		"owner not confirmed gone":  func(x *Takeover) { x.OwnerGone = false },
		"writer not stopped":        func(x *Takeover) { x.WriterStopped = false },
		"uncommitted not preserved": func(x *Takeover) { x.UncommittedPreserved = false },
		"no preserved reference":    func(x *Takeover) { x.PreservedRef = " " },
	} {
		ev := fullTakeover()
		mut(&ev)
		if _, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", ev); !errors.Is(err, core.ErrEvidenceIncomplete) {
			t.Errorf("%s: %v", name, err)
		}
	}
	cur, _ := r.Store.Read(ref)
	if cur.Holder != "sess-a" || cur.Epoch != tk.Epoch {
		t.Fatalf("a refused takeover changed the record: %+v", cur)
	}
}

func TestTakeoverHandsTheIssueToTheNewOwnerAndKeepsTheWork(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	if _, err := r.SetWait(tk, wait1()); err != nil {
		t.Fatal(err)
	}
	if _, err := r.BindWorktree(tk, wtDir("bound")); err != nil {
		t.Fatal(err)
	}
	rec, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", fullTakeover())
	if err != nil {
		t.Fatal(err)
	}
	if rec.Holder != "sess-new" || rec.Epoch != tk.Epoch+1 || rec.Receipt == nil || rec.State != core.StateLanding || rec.Worktree != wtDir("bound") {
		t.Errorf("new owner, next epoch, PR and worktree kept: %+v", rec)
	}
	if rec.Wait != nil || rec.Repair != nil {
		t.Errorf("the old owner's wait and repair claim do not follow: %+v %+v", rec.Wait, rec.Repair)
	}
	last := rec.History[len(rec.History)-1]
	if last.Kind != "takeover" || !strings.Contains(last.Note, "preserve-1") {
		t.Errorf("history must name the preserved reference: %+v", last)
	}
	if _, err := r.SetWait(tk, wait1()); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("the dead owner's ticket still wrote: %v", err)
	}
	if _, err := r.Complete(Ticket{Issue: ref, Holder: "sess-new", Epoch: rec.Epoch}, Landing{RemoteMerged: true, LocalSyncRecorded: true, IssueClosed: true, Deploy: "N/A: x"}); err != nil {
		t.Errorf("the new owner finishes the PR: %v", err)
	}
}

func TestTakeoverRefusesAStaleDecision(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	stale := tk
	stale.Epoch++ // the decision was taken against an epoch the owner has not reached
	if _, err := r.TakeOver(ref, stale, "sess-new", "disp-t", "run-2", fullTakeover()); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("wrong epoch: %v", err)
	}
	other := Ticket{Issue: ref, Holder: "someone", Epoch: tk.Epoch}
	if _, err := r.TakeOver(ref, other, "sess-new", "disp-t", "run-2", fullTakeover()); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("wrong holder: %v", err)
	}
	if _, err := r.TakeOver(ref, tk, "", "disp-t", "run-2", fullTakeover()); !errors.Is(err, core.ErrIDInvalid) {
		t.Errorf("no new holder: %v", err)
	}
}
