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

// A parked Issue handed to a new holder keeps a wait record, and it is the new
// holder's: the old owner's does not follow, and an Issue that is not parked
// needs none.
func TestTakingOverAParkedIssueNeedsTheNewHoldersWait(t *testing.T) {
	r, tk := parkedForContractChange(t)
	ev := fullTakeover()
	if _, err := r.TakeOver(tk.Issue, tk, "sess-new", "disp-t", "run-2", ev); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("a parked Issue without the new holder's wait: %v", err)
	}
	ev.Wait = ptr(wait1()) // owned by sess-a: the old owner's
	if _, err := r.TakeOver(tk.Issue, tk, "sess-new", "disp-t", "run-2", ev); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Fatalf("the old owner's wait handed to the new holder: %v", err)
	}
	if cur, _ := r.Store.Read(tk.Issue); cur.Holder != "sess-a" || cur.Epoch != tk.Epoch {
		t.Fatalf("a refused takeover moved the Issue: %+v", cur)
	}
	w := wait1()
	w.Owner = "sess-new"
	ev.Wait = &w
	rec, err := r.TakeOver(tk.Issue, tk, "sess-new", "disp-t", "run-2", ev)
	if err != nil || rec.Holder != "sess-new" || rec.Park == nil || rec.Wait == nil || rec.Wait.Owner != "sess-new" {
		t.Fatalf("takeover of a parked Issue: %+v %v", rec, err)
	}
}

func ptr[T any](v T) *T { return &v }

// Design E05: a takeover is a new attempt too. The epoch goes up, the attempt
// counter goes up with a rule-made id, and the attempt that was running stays in
// the history.
func TestTakeoverStartsANewAttemptAndKeepsTheOldOneInTheHistory(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	before, _ := r.Store.Read(ref)
	rec, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", fullTakeover())
	if err != nil {
		t.Fatal(err)
	}
	if rec.AttemptSeq != before.AttemptSeq+1 || rec.Attempt == before.Attempt || !strings.HasPrefix(string(rec.Attempt), "takeover-") {
		t.Errorf("a takeover is a new attempt: %q seq %d (was %q seq %d)", rec.Attempt, rec.AttemptSeq, before.Attempt, before.AttemptSeq)
	}
	if _, err := core.ParseAttemptID(string(rec.Attempt)); err != nil {
		t.Errorf("the attempt id must be valid: %v", err)
	}
	last := rec.History[len(rec.History)-1]
	if last.Kind != "takeover" || last.Attempt != string(before.Attempt) {
		t.Errorf("the attempt that was running must be kept in the history: %+v", last)
	}
	if rec.Receipt == nil || rec.Receipt.Attempt != before.Attempt {
		t.Errorf("the PR keeps the attempt that published it: %+v", rec.Receipt)
	}
}

// REQ-AF-009: the owner's departure is established by the session being gone OR by
// a stop record the old owner left (a safe stop); one of the two is required.
func TestTakeoverAcceptsASafeStopInsteadOfOwnerGone(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	stop, err := r.RecordSafeStop(tk, "owner stopped itself")
	if err != nil {
		t.Fatal(err)
	}
	none := fullTakeover()
	none.OwnerGone, none.SafeStop = false, ""
	if _, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", none); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("neither the owner gone nor a safe stop: %v", err)
	}
	blank := none
	blank.SafeStop = "  "
	if _, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", blank); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("a blank safe stop reference: %v", err)
	}
	// the other three facts stay required
	for name, mut := range map[string]func(*Takeover){
		"writer not stopped": func(x *Takeover) { x.WriterStopped = false },
		"not preserved":      func(x *Takeover) { x.UncommittedPreserved = false },
	} {
		ev := none
		ev.SafeStop = stop
		mut(&ev)
		if _, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", ev); !errors.Is(err, core.ErrEvidenceIncomplete) {
			t.Errorf("with a safe stop but %s: %v", name, err)
		}
	}
	ok := none
	ok.SafeStop = stop
	rec, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", ok)
	if err != nil || rec.Holder != "sess-new" {
		t.Fatalf("a safe stop in place of owner-gone: %+v %v", rec, err)
	}
	if last := rec.History[len(rec.History)-1]; !strings.Contains(last.Note, stop) {
		t.Errorf("the history must name the safe stop: %+v", last)
	}
}

// A reference that names no stop record of this owner at this epoch is not a safe
// stop, however it is spelled: a string is not evidence (REQ-AF-009, design 412).
func TestTakeoverRefusesASafeStopThatIsNotARecordOfThisOwner(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	other, _ := core.NewIssueRef(2000)
	otherTk, err := r.Claim(other, "sess-b", "disp-o", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := r.RecordSafeStop(otherTk, "another owner's stop")
	if err != nil {
		t.Fatal(err)
	}
	for name, stop := range map[string]string{
		"an invented reference":              "stop-record-7",
		"a well-formed name nobody wrote":    "issue-1948-e1",
		"another Issue's stop record":        foreign,
		"a name that is not a valid doc key": "../etc/passwd",
	} {
		ev := fullTakeover()
		ev.OwnerGone, ev.SafeStop = false, stop
		if _, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", ev); !errors.Is(err, core.ErrEvidenceIncomplete) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if cur, _ := r.Store.Read(ref); cur.Holder != "sess-a" {
		t.Errorf("a refused takeover moved the record: %+v", cur)
	}
}

// A stop record belongs to one owner and one epoch: after a takeover the new
// owner's epoch needs its own.
func TestASafeStopOfAnEarlierEpochDoesNotStandForTheNextOwner(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	stop, err := r.RecordSafeStop(tk, "first owner stopped")
	if err != nil {
		t.Fatal(err)
	}
	ev := fullTakeover()
	ev.OwnerGone, ev.SafeStop = false, stop
	rec, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", ev)
	if err != nil {
		t.Fatal(err)
	}
	newTk := Ticket{Issue: ref, Holder: "sess-new", Epoch: rec.Epoch}
	ev2 := fullTakeover()
	ev2.OwnerGone, ev2.SafeStop = false, stop
	if _, err := r.TakeOver(ref, newTk, "sess-third", "disp-u", "run-3", ev2); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("the earlier owner's stop record must not carry the next takeover: %v", err)
	}
}

// The next owner records its own stop under its own epoch: the reference differs
// from the first owner's and a takeover resting on it goes through.
func TestEachOwnersSafeStopIsItsOwnRecord(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	first, err := r.RecordSafeStop(tk, "first owner")
	if err != nil {
		t.Fatal(err)
	}
	ev := fullTakeover()
	ev.OwnerGone, ev.SafeStop = false, first
	rec, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", ev)
	if err != nil {
		t.Fatal(err)
	}
	newTk := Ticket{Issue: ref, Holder: "sess-new", Epoch: rec.Epoch}
	second, err := r.RecordSafeStop(newTk, "second owner")
	if err != nil {
		t.Fatal(err)
	}
	if second == first {
		t.Fatalf("two epochs, one record name %q", first)
	}
	ev2 := fullTakeover()
	ev2.OwnerGone, ev2.SafeStop = false, second
	if _, err := r.TakeOver(ref, newTk, "sess-third", "disp-u", "run-3", ev2); err != nil {
		t.Errorf("the second owner's own stop record: %v", err)
	}
}

// A stop the owner has acted after is not a stop: the takeover is refused even
// though a stop record of this owner and epoch exists.
func TestTakeoverRefusesAStopTheOwnerHasActedAfter(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	stop, err := r.RecordSafeStop(tk, "owner stopped")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.SetWait(tk, wait1()); err != nil { // the owner is working again
		t.Fatal(err)
	}
	ev := fullTakeover()
	ev.OwnerGone, ev.SafeStop = false, stop
	if _, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", ev); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("a stale stop: %v", err)
	}
	if cur, _ := r.Store.Read(ref); cur.Holder != "sess-a" {
		t.Errorf("the record moved: %+v", cur)
	}
	again, err := r.RecordSafeStop(tk, "owner stopped again")
	if err != nil {
		t.Fatal(err)
	}
	ev.SafeStop = again
	if _, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", ev); err != nil {
		t.Errorf("the owner stopped again: %v", err)
	}
}

// What the owner does after TakeOver has decided and before the claim moves is
// caught by the store's check under the lock.
func TestTakeoverRefusesAStopTheOwnerActsAfterBeforeTheClaimMoves(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	stop, err := r.RecordSafeStop(tk, "owner stopped")
	if err != nil {
		t.Fatal(err)
	}
	r.beforeClaimMoves = func() {
		if _, err := r.SetWait(tk, wait1()); err != nil { // the owner resumes right here
			t.Error(err)
		}
	}
	ev := fullTakeover()
	ev.OwnerGone, ev.SafeStop = false, stop
	if _, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", ev); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("an owner that acted between the checks: %v", err)
	}
	if cur, _ := r.Store.Read(ref); cur.Holder != "sess-a" {
		t.Errorf("the record moved: %+v", cur)
	}
}
