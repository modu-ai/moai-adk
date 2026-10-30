package runner

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

var (
	headA = strings.Repeat("a", 40)
	headB = strings.Repeat("b", 40)
)

func newRunner(t *testing.T) (*Runner, core.IssueRef) {
	t.Helper()
	st, err := claim.OpenWithLimit(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	ref, _ := core.NewIssueRef(1948)
	return withRoles(New(st)), ref // identity lookup and host, both faked
}

// landOK publishes the PR the way the flow does: the preflight on this head
// first, then the landing.
func landOK(r *Runner, tk Ticket, head string) (claimRecord, error) {
	if _, err := r.PreflightPR(tk, head, "acme/widgets", "Refs #1948"); err != nil {
		return claimRecord{}, err
	}
	return r.Land(tk, head)
}

// passAllFor is a complete passing evidence set for the ticket's holder and epoch.
func passAllFor(tk Ticket, head string) Verification {
	return Verification{Head: head, Results: []Result{
		{Kind: KindCheck, Name: "build", Head: head, Verdict: VerdictPass},
		{Kind: KindReview, Name: "code-review", Head: head, Verdict: VerdictPass},
		{Kind: KindAudit, Name: "independent-audit", Head: head, Verdict: VerdictPass},
		{Kind: KindHandoffWriter, Name: "writer", Head: head, Verdict: VerdictPass, Session: tk.Holder, Epoch: tk.Epoch},
		{Kind: KindHandoffCoordinator, Name: "coordinator", Head: head, Verdict: VerdictPass},
	}}
}

func passAll(head string) Verification { return passAllFor(Ticket{Holder: "sess-a", Epoch: 1}, head) }

type claimRecord = claim.Record

// withRoles gives a Runner built by hand the faked attribution lookup.
func withRoles(r *Runner) *Runner {
	r.Attributor = goodRoles()
	r.Facts = ownIssues{}
	return r
}

// ownIssues is a host where every Issue was opened by the acting user.
type ownIssues struct{}

func (ownIssues) IssueFacts(context.Context, core.IssueRef) (IssueFacts, error) {
	return IssueFacts{RequestedRepo: "acme/widgets", ActualRepo: "acme/widgets", UserID: 7, AuthorID: 7, Owner: core.OwnerNone}, nil
}

func toVerifying(t *testing.T, r *Runner, ref core.IssueRef) Ticket {
	t.Helper()
	tk, err := r.Claim(ref, "sess-a", "disp-1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.StartAttempt(tk, "attempt-1", core.DeliveryExecuted); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Submit(tk, headA, false); err != nil {
		t.Fatal(err)
	}
	return tk
}

func TestHappyPathQueuedToDone(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	rec, err := r.Verify(tk, passAllFor(tk, headA))
	if err != nil || rec.State != core.StateReady {
		t.Fatalf("verify: %+v %v", rec, err)
	}
	if rec, err = landOK(r, tk, headA); err != nil || rec.State != core.StateLanding {
		t.Fatalf("land: %+v %v", rec, err)
	}
	rec, err = r.Complete(tk, Landing{RemoteMerged: true, LocalSyncRecorded: true, IssueClosed: true, Deploy: "N/A: no deploy target"})
	if err != nil || rec.State != core.StateDone {
		t.Fatalf("complete: %+v %v", rec, err)
	}
}

func TestStartAttemptNeedsExecutedDelivery(t *testing.T) {
	r, ref := newRunner(t)
	tk, _ := r.Claim(ref, "sess-a", "disp-1", "run-1")
	for _, d := range []core.Delivery{core.DeliverySent, core.DeliveryReceived, ""} {
		if _, err := r.StartAttempt(tk, "attempt-1", d); !errors.Is(err, core.ErrTransitionInvalid) {
			t.Errorf("delivery %q must not start an attempt, got %v", d, err)
		}
	}
}

func TestSubmitRefusesDirtyTreeAndShortHead(t *testing.T) {
	r, ref := newRunner(t)
	tk, _ := r.Claim(ref, "sess-a", "disp-1", "run-1")
	if _, err := r.StartAttempt(tk, "attempt-1", core.DeliveryExecuted); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Submit(tk, headA, true); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("dirty tree: %v", err)
	}
	if _, err := r.Submit(tk, "cb30e05b3e", false); !errors.Is(err, core.ErrHeadInvalid) {
		t.Errorf("short head: %v", err)
	}
}

func TestStepsCannotBeSkipped(t *testing.T) {
	r, ref := newRunner(t)
	tk, _ := r.Claim(ref, "sess-a", "disp-1", "run-1")
	if _, err := r.Submit(tk, headA, false); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("claimed->verifying must be refused: %v", err)
	}
	if _, err := landOK(r, tk, headA); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("claimed->landing must be refused: %v", err)
	}
	if _, err := r.Complete(tk, Landing{RemoteMerged: true, LocalSyncRecorded: true, Deploy: "N/A: x"}); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("claimed->done must be refused: %v", err)
	}
}

func TestWriteRefusedWithoutOwnership(t *testing.T) {
	r, ref := newRunner(t)
	tk, _ := r.Claim(ref, "sess-a", "disp-1", "run-1")
	bad := Ticket{Issue: ref, Holder: "intruder", Epoch: tk.Epoch}
	if _, err := r.StartAttempt(bad, "attempt-1", core.DeliveryExecuted); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("want ownership refusal, got %v", err)
	}
	if rec, _ := r.Store.Load(ref); rec.State != core.StateClaimed {
		t.Errorf("state moved by refused write: %v", rec.State)
	}
}

func TestDuplicateClaimDispatchRefused(t *testing.T) {
	r, ref := newRunner(t)
	if _, err := r.Claim(ref, "sess-a", "disp-1", "run-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Claim(ref, "sess-a", "disp-1", "run-1"); !errors.Is(err, core.ErrDuplicateDispatch) {
		t.Errorf("want duplicate, got %v", err)
	}
}

func TestVerifyStaleHeadRejected(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	rec, err := r.Verify(tk, passAllFor(tk, headB))
	if !errors.Is(err, core.ErrStaleHead) {
		t.Fatalf("want ErrStaleHead, got %v", err)
	}
	if rec.State != core.StateVerifying && rec.State != "" {
		t.Errorf("state must not advance on stale head: %v", rec.State)
	}
	cur, _ := r.Store.Load(ref)
	if cur.State != core.StateVerifying {
		t.Errorf("stored state = %v", cur.State)
	}
	// A single mismatched result inside an otherwise good set is stale too.
	v := passAll(headA)
	v.Results[1].Head = headB
	if _, err := r.Verify(tk, v); !errors.Is(err, core.ErrStaleHead) {
		t.Errorf("mixed heads: %v", err)
	}
}

func TestVerifyIncompleteEvidenceStaysVerifying(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	v := passAll(headA)
	v.Results = v.Results[:3] // no handoffs
	if _, err := r.Verify(tk, v); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("want ErrEvidenceIncomplete, got %v", err)
	}
	if cur, _ := r.Store.Load(ref); cur.State != core.StateVerifying {
		t.Errorf("state = %v", cur.State)
	}
}

func TestVerifyVerdictMapping(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*Verification)
		wantSt  core.State
		wantWhy core.BlockReason
	}{
		{"fail goes back to implementing", func(v *Verification) { v.Results[1].Verdict = VerdictFail }, core.StateImplementing, ""},
		{"debt records a gate", func(v *Verification) { v.Results[2].Verdict = VerdictPassWithDebt }, core.StateVerifying, core.ReasonDebtPresent},
		{"skipped check records a gate", func(v *Verification) { v.Results[0].Verdict = "skipped" }, core.StateVerifying, core.ReasonCheckNotPassed},
		{"unknown audit verdict records a gate", func(v *Verification) { v.Results[2].Verdict = "" }, core.StateVerifying, core.ReasonVerdictNotPass},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, ref := newRunner(t)
			tk := toVerifying(t, r, ref)
			v := passAll(headA)
			c.mutate(&v)
			rec, err := r.Verify(tk, v)
			if err != nil {
				t.Fatal(err)
			}
			if rec.State != c.wantSt || gateReason(rec) != c.wantWhy {
				t.Errorf("got %v/%q want %v/%q", rec.State, gateReason(rec), c.wantSt, c.wantWhy)
			}
			if c.wantWhy != "" && (rec.Park != nil || rec.Return != "" || rec.Reason != "" || rec.Stage != core.StageBuilding) {
				t.Errorf("a gate is a record on the stage, not a pause: %+v", rec)
			}
		})
	}
}

func TestLandRechecksHead(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	if _, err := r.Verify(tk, passAllFor(tk, headA)); err != nil {
		t.Fatal(err)
	}
	if _, err := landOK(r, tk, headB); !errors.Is(err, core.ErrStaleHead) {
		t.Errorf("want stale head, got %v", err)
	}
}

func TestCompleteNeedsAllThreeRecords(t *testing.T) {
	cases := map[string]Landing{
		"no remote merge": {LocalSyncRecorded: true, IssueClosed: true, Deploy: "N/A: x"},
		"no issue close":  {RemoteMerged: true, LocalSyncRecorded: true, Deploy: "N/A: x"},
		"no local sync":   {RemoteMerged: true, IssueClosed: true, Deploy: "N/A: x"},
		"no deploy":       {RemoteMerged: true, LocalSyncRecorded: true, IssueClosed: true},
		"bare N/A":        {RemoteMerged: true, LocalSyncRecorded: true, IssueClosed: true, Deploy: "N/A"},
	}
	for name, l := range cases {
		t.Run(name, func(t *testing.T) {
			r, ref := newRunner(t)
			tk := toVerifying(t, r, ref)
			_, _ = r.Verify(tk, passAllFor(tk, headA))
			_, _ = landOK(r, tk, headA)
			if _, err := r.Complete(tk, l); !errors.Is(err, core.ErrEvidenceIncomplete) {
				t.Errorf("want ErrEvidenceIncomplete, got %v", err)
			}
			cur, _ := r.Store.Load(ref)
			if cur.State != core.StateLanding || cur.Closed {
				t.Errorf("state = %v closed=%v", cur.State, cur.Closed)
			}
			// A merge that was observed is kept as a fact with the open obligation named.
			if l.RemoteMerged != cur.Merged || (l.RemoteMerged && cur.Pending == "") {
				t.Errorf("merged=%v pending=%q for %+v", cur.Merged, cur.Pending, l)
			}
		})
	}
}

func TestPauseAndResumeReturnOnlyToReturnStage(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	rec, err := r.Pause(tk, core.StateBlocked, core.ReasonRateLimited, "retry after 15:00", wait1())
	if err != nil || rec.State != core.StateBlocked || rec.Return != core.StateVerifying || rec.Resume == "" {
		t.Fatalf("pause: %+v %v", rec, err)
	}
	if _, err := r.Pause(tk, core.StateBlocked, core.ReasonOperatorHold, "x", wait1()); !errors.Is(err, core.ErrReasonNotAllowed) {
		t.Errorf("operator-hold on blocked: %v", err)
	}
	if _, err := r.Resume(tk, false); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("resume without condition met: %v", err)
	}
	rec, err = r.Resume(tk, true)
	if err != nil || rec.State != core.StateVerifying || rec.Reason != "" || rec.Return != "" {
		t.Errorf("resume: %+v %v", rec, err)
	}
}

func TestPauseRequiresResumeConditionAndCannotNest(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	if _, err := r.Pause(tk, core.StateOnHold, core.ReasonOperatorHold, "", wait1()); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("empty resume condition: %v", err)
	}
	if _, err := r.Pause(tk, core.StateOnHold, core.ReasonOperatorHold, "operator resumes", wait1()); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Pause(tk, core.StateBlocked, core.ReasonRateLimited, "later", wait1()); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("pause while paused: %v", err)
	}
}

func TestPausedIssueRefusesForwardSteps(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	_, _ = r.Pause(tk, core.StateOnHold, core.ReasonOperatorHold, "operator resumes", wait1())
	if _, err := r.Verify(tk, passAllFor(tk, headA)); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("verify while on-hold: %v", err)
	}
}

// Session drop: a fresh Runner over the same store sees the persisted state
// and continues with the original ticket; nothing is replayed.
func TestReconnectContinuesFromPersistedState(t *testing.T) {
	st, _ := claim.OpenWithLimit(t.TempDir(), 100)
	ref, _ := core.NewIssueRef(1948)
	first := withRoles(New(st))
	tk := toVerifying(t, first, ref)

	second := withRoles(New(st)) // new process, same store
	cur, err := second.Store.Load(ref)
	if err != nil || cur.State != core.StateVerifying || cur.Head != core.HeadSHA(headA) {
		t.Fatalf("persisted state lost: %+v %v", cur, err)
	}
	if rec, err := second.Verify(tk, passAllFor(tk, headA)); err != nil || rec.State != core.StateReady {
		t.Fatalf("continue after reconnect: %+v %v", rec, err)
	}
	// The re-sent dispatch after reconnect is not a second run.
	if _, err := second.Claim(ref, "sess-a", "disp-1", "run-1"); !errors.Is(err, core.ErrDuplicateDispatch) {
		t.Errorf("re-sent dispatch: %v", err)
	}
}

// A result produced for the previous HEAD must not count after a repair
// attempt submitted a new HEAD.
func TestPreviousHeadResultRejectedAfterRepair(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	v := passAll(headA)
	v.Results[1].Verdict = VerdictFail
	if rec, err := r.Verify(tk, v); err != nil || rec.State != core.StateImplementing {
		t.Fatalf("fail path: %+v %v", rec, err)
	}
	if _, err := r.Submit(tk, headB, false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Verify(tk, passAllFor(tk, headA)); !errors.Is(err, core.ErrStaleHead) {
		t.Errorf("old-HEAD pass must be stale: %v", err)
	}
	if rec, err := r.Verify(tk, passAllFor(tk, headB)); err != nil || rec.State != core.StateReady {
		t.Errorf("new-HEAD pass: %+v %v", rec, err)
	}
}

// Partial failure: a refused ownership write mid-flow leaves the stage where it was.
func TestPartialFailureKeepsStage(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	stale := Ticket{Issue: ref, Holder: tk.Holder, Epoch: tk.Epoch + 1}
	if _, err := r.Verify(stale, passAllFor(stale, headA)); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Fatalf("stale epoch: %v", err)
	}
	if cur, _ := r.Store.Load(ref); cur.State != core.StateVerifying {
		t.Errorf("stage moved: %v", cur.State)
	}
}

// Merge-gate behaviour change: a non-pass no longer pauses the Issue. The gate is
// a record on the verifying stage; judging the same HEAD with the same evidence
// records the same gate and nothing else, and only new evidence passes.
func TestAGateIsRecordedOnceAndOnlyNewEvidenceOrAHeadPassesIt(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	bad := func() Verification {
		v := passAllFor(tk, headA)
		v.Results[2].Verdict = VerdictPassWithDebt
		return v
	}
	first, err := r.Verify(tk, bad())
	if err != nil || first.Gate == nil || first.State != core.StateVerifying {
		t.Fatalf("first: %+v %v", first, err)
	}
	for i := 0; i < 3; i++ {
		again, err := r.Verify(tk, bad())
		if err != nil {
			t.Fatalf("again %d: %v", i, err)
		}
		if again.State != first.State || again.Stage != first.Stage || *again.Gate != *first.Gate || again.Park != nil || again.Return != "" {
			t.Fatalf("the same evidence must record the same gate and change nothing else: %+v vs %+v", again, first)
		}
	}
	// the gate does not let the Issue land
	if _, err := landOK(r, tk, headA); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("land with an open gate: %v", err)
	}
	// new evidence for the same HEAD (a rerun without debt) passes and clears the gate
	passed, err := r.Verify(tk, passAllFor(tk, headA))
	if err != nil || passed.State != core.StateReady || passed.Gate != nil {
		t.Fatalf("new evidence: %+v %v", passed, err)
	}
}

func TestANewHeadAfterAFailStartsFromACleanGate(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	v := passAllFor(tk, headA)
	v.Results[2].Verdict = VerdictPassWithDebt
	if rec, err := r.Verify(tk, v); err != nil || rec.Gate == nil {
		t.Fatalf("gate: %+v %v", rec, err)
	}
	fail := passAllFor(tk, headA)
	fail.Results[1].Verdict = VerdictFail
	if rec, err := r.Verify(tk, fail); err != nil || rec.State != core.StateImplementing || rec.Gate != nil {
		t.Fatalf("fail clears the gate: %+v %v", rec, err)
	}
	if rec, err := r.Submit(tk, headB, false); err != nil || rec.Gate != nil || string(rec.Head) != headB {
		t.Fatalf("submit: %+v %v", rec, err)
	}
}

func TestPauseTakesOnlyOverlayReasons(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	for _, why := range append(core.GateReasons(), core.ReasonForeignClosing, core.ReasonNotIssueAuthor) {
		if _, err := r.Pause(tk, core.StateBlocked, why, "later", wait1()); !errors.Is(err, core.ErrReasonNotAllowed) {
			t.Errorf("pause with %s: %v", why, err)
		}
	}
	rec, err := r.Pause(tk, core.StateBlocked, core.ReasonRateLimited, "later", wait1())
	if err != nil || rec.Park == nil || rec.Park.Reason != core.ReasonRateLimited || rec.Stage != core.StageBuilding {
		t.Errorf("overlay pause: %+v %v", rec, err)
	}
}
