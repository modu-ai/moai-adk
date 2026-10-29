package runner

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func mustReopen(t *testing.T, r *Runner, ref core.IssueRef) {
	t.Helper()
	if _, err := r.Reopen(ref, "ci failed on latest head"); err != nil {
		t.Fatal(err)
	}
}

func goodHandoff() Handoff {
	return Handoff{HandoffID: "ho-1", LeaderAccepted: true, PR: 321, Head: headA, Branch: "WT-agentic-x", Worktree: "/wt/x", WriterStopped: true, BoundaryHead: headA}
}

// landed drives an Issue to landing (PR published) and returns its ticket.
func landed(t *testing.T, r *Runner, ref core.IssueRef) Ticket {
	t.Helper()
	tk := toVerifying(t, r, ref)
	if _, err := r.Verify(tk, passAllFor(tk, headA)); err != nil {
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
		t.Errorf("agent must no longer hold the claim: %v", err)
	}
	// The agent's old ticket can no longer write.
	if _, err := r.Complete(tk, Landing{RemoteMerged: true, LocalSyncRecorded: true, Deploy: "N/A: x"}); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("released agent wrote: %v", err)
	}
	// A different Issue can now be claimed by the same agent.
	other, _ := core.NewIssueRef(2000)
	if _, err := r.Claim(other, "sess-a", "disp-2", "run-1"); err != nil {
		t.Errorf("agent could not take a new issue: %v", err)
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
				t.Errorf("claim must stay with the agent: %+v %v", cur, err)
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

// The same receipt delivered twice (or late, after the claim was returned)
// converges on one acceptance and one claim return.
func TestSameReceiptTwiceIsAcceptedOnce(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	first, err := r.HandOff(tk, goodHandoff())
	if err != nil {
		t.Fatal(err)
	}
	again, err := r.HandOff(tk, goodHandoff())
	if err != nil {
		t.Fatalf("the same receipt again must be accepted idempotently: %v", err)
	}
	if again.Seq != first.Seq || again.Receipt == nil || again.Receipt.HandoffID != "ho-1" || len(again.History) != len(first.History) {
		t.Errorf("the replay changed the record: first=%+v again=%+v", first, again)
	}
	// The same id with different content is a conflict, not a replay.
	other := goodHandoff()
	other.PR = 999
	if _, err := r.HandOff(tk, other); !errors.Is(err, core.ErrDuplicateHandoff) {
		t.Errorf("same id, other PR: %v", err)
	}
	// Another receipt for the same HEAD from the released ticket is not accepted:
	// the claim was already returned once.
	second := goodHandoff()
	second.HandoffID = "ho-other"
	if _, err := r.HandOff(tk, second); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("a different receipt after the claim was returned: %v", err)
	}
}

func TestReceiptNeedsAnIdAndTheLeadersAcceptance(t *testing.T) {
	for name, mut := range map[string]func(*Handoff){
		"no handoff id":       func(h *Handoff) { h.HandoffID = "" },
		"bad handoff id":      func(h *Handoff) { h.HandoffID = "Not Valid/1" },
		"leader not accepted": func(h *Handoff) { h.LeaderAccepted = false },
	} {
		r, ref := newRunner(t)
		tk := landed(t, r, ref)
		h := goodHandoff()
		mut(&h)
		if _, err := r.HandOff(tk, h); err == nil {
			t.Errorf("%s: accepted", name)
		}
		if cur, err := r.Store.Load(ref); err != nil || cur.Holder != "sess-a" || cur.Receipt != nil {
			t.Errorf("%s: the claim must stay with the agent until the leader accepted: %+v %v", name, cur, err)
		}
	}
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	rec, err := r.HandOff(tk, goodHandoff())
	if err != nil || !rec.Receipt.LeaderAccepted || rec.Receipt.HandoffID != "ho-1" {
		t.Errorf("accepted receipt: %+v %v", rec.Receipt, err)
	}
	if last := rec.History[len(rec.History)-1]; last.Kind != "handoff" || !strings.Contains(last.Note, "ho-1") {
		t.Errorf("the history must name the receipt: %+v", last)
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
	rec, err := r.CompleteHandedOff(ref, Landing{RemoteMerged: true, LocalSyncRecorded: true, IssueClosed: true, Deploy: "N/A: x"})
	if err != nil || rec.State != core.StateDone || !rec.Closed {
		t.Fatalf("complete after merge: %+v %v", rec, err)
	}
}

// Done includes closing the Issue (S46 E13). A failed close keeps the merge
// fact, names the open obligation and leaves done at zero (E14).
func TestDoneNeedsIssueCloseAndPendingEffectKeepsMergeFact(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	_, _ = r.HandOff(tk, goodHandoff())
	noClose := Landing{RemoteMerged: true, LocalSyncRecorded: true, Deploy: "N/A: x"}
	if _, err := r.CompleteHandedOff(ref, noClose); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("close missing: %v", err)
	}
	cur, _ := r.Store.Read(ref)
	if cur.State != core.StateLanding || !cur.Merged || cur.Pending != "issue-close" || cur.Closed {
		t.Errorf("pending effect not recorded: %+v", cur)
	}
	noClose.IssueClosed = true
	rec, err := r.CompleteHandedOff(ref, noClose)
	if err != nil || rec.State != core.StateDone || rec.Pending != "" || !rec.Closed {
		t.Errorf("retry after close: %+v %v", rec, err)
	}
}

func TestUnmergedAttemptRecordsNoMergeFact(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	_, _ = r.HandOff(tk, goodHandoff())
	_, _ = r.CompleteHandedOff(ref, Landing{LocalSyncRecorded: true, IssueClosed: true, Deploy: "N/A: x"})
	if cur, _ := r.Store.Read(ref); cur.Merged || cur.Pending != "" {
		t.Errorf("merge fact invented: %+v", cur)
	}
}

// E10 then E06: a PR that is not merged goes back to verifying, and only from
// there can a single repair claim be taken; the merge fact is never reverted.
func TestReopenThenSingleRepairClaim(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	_, _ = r.HandOff(tk, goodHandoff())
	if _, err := r.ClaimRepair(ref, "sess-b", "disp-r", "run-2", "WT-agentic-x"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("repair straight from landing: %v", err)
	}
	if _, err := r.Reopen(ref, ""); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("reopen needs a reason: %v", err)
	}
	rec, err := r.Reopen(ref, "ci failed on latest head")
	if err != nil || rec.State != core.StateVerifying || rec.Holder != "" || rec.Receipt == nil {
		t.Fatalf("reopen: %+v %v", rec, err)
	}
	if _, err := r.ClaimRepair(ref, "sess-b", "disp-r", "run-2", "WT-agentic-x"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ClaimRepair(ref, "sess-c", "disp-r2", "run-3", "WT-agentic-x"); !errors.Is(err, core.ErrHeld) {
		t.Errorf("second repair claim: %v", err)
	}
}

func TestReopenRefusedAfterMergeOrWithoutHandoff(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	if _, err := r.Reopen(ref, "x"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("reopen while still held: %v", err)
	}
	_, _ = r.HandOff(tk, goodHandoff())
	_, _ = r.CompleteHandedOff(ref, Landing{RemoteMerged: true, LocalSyncRecorded: true, Deploy: "N/A: x"}) // pending: close
	if _, err := r.Reopen(ref, "x"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("a merged PR must not be reverted by reopen: %v", err)
	}
}

// History stays under the one Issue key; at most one active claim exists.
func TestHistoryKeptUnderOneIssueAcrossRepair(t *testing.T) {
	r, ref := newRunner(t)
	a1, _ := r.Store.RegisterAgent("aaaaaaaa-0000-0000-0000-000000000001")
	a2, _ := r.Store.RegisterAgent("aaaaaaaa-0000-0000-0000-000000000002")
	tk, _ := r.ClaimAs(ref, a1, "disp-1", "run-1")
	_, _ = r.StartAttempt(tk, "attempt-1", core.DeliveryExecuted)
	_, _ = r.Submit(tk, headA, false)
	_, _ = r.Verify(tk, passAllFor(tk, headA))
	_, _ = r.Land(tk, headA)
	if _, err := r.HandOff(tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	_, _ = r.Reopen(ref, "ci failed")
	rep, err := r.ClaimRepairAs(ref, a2, "disp-r", "run-2", "WT-agentic-x")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = r.StartAttempt(rep, "attempt-2", core.DeliveryExecuted)
	_, _ = r.Submit(rep, headB, false)
	_, _ = r.Verify(rep, passAllFor(rep, headB))
	_, _ = r.Land(rep, headB)
	h := goodHandoff()
	h.HandoffID, h.Head, h.BoundaryHead = "ho-2", headB, headB
	if _, err := r.HandOff(rep, h); err != nil {
		t.Fatal(err)
	}
	recs, _ := r.Store.List()
	if len(recs) != 1 {
		t.Fatalf("one record per issue key, got %d", len(recs))
	}
	var kinds []string
	for _, e := range recs[0].History {
		kinds = append(kinds, e.Kind)
	}
	want := []string{"claim", "handoff", "reopen", "repair-claim", "handoff"}
	if len(kinds) != len(want) {
		t.Fatalf("history = %v", kinds)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("history = %v, want %v", kinds, want)
		}
	}
	if recs[0].History[0].Agent != "agent-1" || recs[0].History[3].Agent != "agent-2" {
		t.Errorf("provenance lost: %+v", recs[0].History)
	}
	if recs[0].Holder != "" {
		t.Errorf("no active claim expected after the second handoff: %+v", recs[0])
	}
}

func TestCompleteHandedOffNeedsAHandoff(t *testing.T) {
	r, ref := newRunner(t)
	_ = landed(t, r, ref) // agent still holds the claim, no receipt
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
	mustReopen(t, r, ref)
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
	mustReopen(t, r, ref)
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
	if _, err := r.Verify(rep, passAllFor(rep, headB)); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Land(rep, headB); err != nil {
		t.Fatal(err)
	}
	h := goodHandoff()
	h.HandoffID, h.Head, h.BoundaryHead = "ho-2", headB, headB
	if rec, err := r.HandOff(rep, h); err != nil || rec.Receipt.Head != core.HeadSHA(headB) || rec.Holder != "" {
		t.Errorf("second handoff: %+v %v", rec, err)
	}
}

func TestBacklogCapHoldsNewClaimsNotRepairs(t *testing.T) {
	r, ref := newRunner(t)
	_ = r.Store.SetMaxOpenPRs(1)
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
	mustReopen(t, r, ref)
	if _, err := r.ClaimRepair(ref, "sess-b", "disp-r", "run-2", "WT-agentic-x"); err != nil {
		t.Errorf("repair must not be blocked by the cap: %v", err)
	}
}

func TestBacklogCapFreedByDone(t *testing.T) {
	r, ref := newRunner(t)
	_ = r.Store.SetMaxOpenPRs(1)
	tk := landed(t, r, ref)
	_, _ = r.HandOff(tk, goodHandoff())
	if _, err := r.CompleteHandedOff(ref, Landing{RemoteMerged: true, LocalSyncRecorded: true, IssueClosed: true, Deploy: "N/A: x"}); err != nil {
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
	st, _ := claim.OpenWithLimit(dir, 100)
	r := withRoles(New(st))
	ref1, _ := core.NewIssueRef(1948)
	ref2, _ := core.NewIssueRef(2000)
	tk := landed(t, r, ref1)
	if _, err := r.HandOff(tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Claim(ref2, "sess-b", "disp-9", "run-9"); err != nil {
		t.Fatal(err)
	}

	st2, err := claim.OpenWithLimit(dir, 100)
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
	r2 := withRoles(New(st2))
	mustReopen(t, r2, ref1)
	if _, err := r2.Claim(ref2, "sess-b", "disp-9", "run-9"); !errors.Is(err, core.ErrDuplicateDispatch) {
		t.Errorf("re-sent dispatch after restart: %v", err)
	}
	if _, err := r2.ClaimRepair(ref1, "sess-c", "disp-r", "run-3", "WT-agentic-x"); err != nil {
		t.Errorf("repair after restart: %v", err)
	}
}

func TestAgentIdentitySurvivesHandoffAndRepair(t *testing.T) {
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
	_, _ = r.Land(tk, headA)
	rec, err := r.HandOff(tk, goodHandoff())
	if err != nil || rec.Receipt.Agent != "agent-1" {
		t.Fatalf("receipt agent: %+v %v", rec, err)
	}
	mustReopen(t, r, ref)
	rep, err := r.ClaimRepairAs(ref, a2, "disp-r", "run-2", "WT-agentic-x")
	if err != nil {
		t.Fatal(err)
	}
	cur, _ := r.Store.Load(ref)
	if cur.AgentName != "agent-2" || cur.Receipt.Agent != "agent-1" || rep.Holder != a2.SessionID {
		t.Errorf("identities: %+v", cur)
	}
}

// After a handoff the agent no longer holds a claim, so its number can be
// returned and reused, yet the open PR stays attributed to the old session.
func TestReceiptStaysAttributedAfterNumberReuse(t *testing.T) {
	r, ref := newRunner(t)
	a1, _ := r.Store.RegisterAgent("aaaaaaaa-0000-0000-0000-000000000001")
	tk, _ := r.ClaimAs(ref, a1, "disp-1", "run-1")
	_, _ = r.StartAttempt(tk, "attempt-1", core.DeliveryExecuted)
	_, _ = r.Submit(tk, headA, false)
	_, _ = r.Verify(tk, passAllFor(tk, headA))
	_, _ = r.Land(tk, headA)
	rec, err := r.HandOff(tk, goodHandoff())
	if err != nil || rec.Receipt.Session != a1.SessionID {
		t.Fatalf("receipt session: %+v %v", rec, err)
	}
	if err := r.Store.ReleaseAgent(a1.SessionID, a1.Number); err != nil {
		t.Fatalf("release after handoff: %v", err)
	}
	a2, _ := r.Store.RegisterAgent("bbbbbbbb-0000-0000-0000-000000000002")
	if a2.Number != a1.Number {
		t.Fatalf("number not reused: %+v", a2)
	}
	cur, _ := r.Store.Read(ref)
	if cur.Receipt.Agent != "agent-1" || cur.Receipt.Session != a1.SessionID || cur.Receipt.Session == a2.SessionID {
		t.Errorf("attribution after reuse: %+v", cur.Receipt)
	}
}
