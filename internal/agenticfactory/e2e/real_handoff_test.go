package e2e

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

// PR handoff, repair, backlog cap and closure on lane-18's real code.

const branch = "WT-handoff-branch"

// goodHandoff is a complete receipt under lane-18's current contract: the URL of
// the PR the receipt carries, the base, the adopted spec revision, what the task
// produced, the check state, and a worktree that exists (it is reserved by its
// directory identity, so a path that is not there is refused).
func goodHandoff(t *testing.T) runner.Handoff {
	t.Helper()
	return runner.Handoff{
		HandoffID: "handoff-1", LeaderAccepted: true,
		PR: 7, Head: headA, Branch: branch, Worktree: t.TempDir(), WriterStopped: true, BoundaryHead: headA,
		PRURL: "https://github.com/o/r/pull/7", Base: "main", AdoptedSpec: "rev-1",
		TaskResult: "implemented; tests added", CheckStatus: "passed",
	}
}

func landingFor(t *testing.T, r *runner.Runner, n int, holder string) runner.Ticket {
	t.Helper()
	ref, err := core.NewIssueRef(n)
	if err != nil {
		t.Fatal(err)
	}
	run, _ := core.ParseRunID("run-1")
	tk, err := r.Claim(ref, holder, "d-"+holder, run)
	if err != nil {
		t.Fatal(err)
	}
	upTo(t, r, tk, core.StateLanding)
	return tk
}

// PR published is not done. Under lane-18's contract the handoff records the
// receipt but does NOT release the Issue: the agent that published the PR stays
// its owner (the receipt names its session) and keeps the Issue in landing, so
// another session's claim is refused as held. Only after the owner releases does
// a new claimant meet the repair-only rule.
func TestRealHandoff_PRPublished_KeepsTheOwnerButIsNotDone(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	h := goodHandoff(t)
	rec, err := r.HandOff(tk, h)
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != core.StateLanding || rec.Receipt == nil || rec.Receipt.PR != 7 {
		t.Fatalf("after handoff: %+v", rec)
	}
	if rec.Holder != tk.Holder || rec.Receipt.Session != tk.Holder {
		t.Errorf("holder %q, receipt session %q; the publisher must stay the owner (%q)", rec.Holder, rec.Receipt.Session, tk.Holder)
	}
	run, _ := core.ParseRunID("run-1")
	if _, err := r.Store.Acquire(tk.Issue, "other", "d-other", run); !errors.Is(err, core.ErrHeld) {
		t.Errorf("claim on a handed-off Issue its publisher still owns: %v, want ErrHeld", err)
	}
	// A replay of the SAME receipt (same id, same content) converges on the one
	// stored acceptance: no error, nothing changes (S46 E03).
	stored, err := r.Store.Read(tk.Issue)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := r.HandOff(tk, h)
	if err != nil || !reflect.DeepEqual(replay, stored) {
		t.Errorf("replayed receipt: %+v, %v; want the stored record unchanged", replay, err)
	}
	// The same id with other content is a conflict.
	other := h
	other.PR = 8
	other.PRURL = "https://github.com/o/r/pull/8"
	if _, err := r.HandOff(tk, other); !errors.Is(err, core.ErrDuplicateHandoff) {
		t.Errorf("same id, other PR: %v, want ErrDuplicateHandoff", err)
	}
	// A new id for the head that already has a receipt is a duplicate, not a second receipt.
	fresh := h
	fresh.HandoffID = "handoff-2"
	if _, err := r.HandOff(tk, fresh); !errors.Is(err, core.ErrDuplicateHandoff) {
		t.Errorf("a second receipt for the same head: %v, want ErrDuplicateHandoff", err)
	}
	if after, _ := r.Store.Read(tk.Issue); !reflect.DeepEqual(after, stored) {
		t.Errorf("record changed by the replay and refusals: %+v -> %+v", stored, after)
	}
	// Once the owner releases, a new claimant is told it is repair-only (the
	// original rule, now reached through the release).
	if err := r.Store.Release(tk.Issue, tk.Holder); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Store.Acquire(tk.Issue, "other", "d-other", run); !errors.Is(err, core.ErrRepairOnly) {
		t.Errorf("fresh claim on a released handed-off Issue: %v, want ErrRepairOnly", err)
	}
}

// The claim is returned only after the leader accepted the receipt, and a
// receipt needs an id.
func TestRealHandoff_NeedsLeaderAcceptanceAndAnId(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	before := state(t, r, tk)
	for name, mod := range map[string]func(*runner.Handoff){
		"leader-not-accepted": func(h *runner.Handoff) { h.LeaderAccepted = false },
		"no-id":               func(h *runner.Handoff) { h.HandoffID = "" },
	} {
		h := goodHandoff(t)
		mod(&h)
		if _, err := r.HandOff(tk, h); !errors.Is(err, core.ErrEvidenceIncomplete) {
			t.Errorf("%s: %v, want ErrEvidenceIncomplete", name, err)
		}
	}
	if after := state(t, r, tk); !reflect.DeepEqual(after, before) {
		t.Errorf("a refused receipt changed the record: %+v -> %+v", before, after)
	}
}

// A handoff needs a stopped writer and an unmoved branch; refusals change nothing.
func TestRealHandoff_Refusals_ChangeNothing(t *testing.T) {
	cases := map[string]struct {
		mod  func(*runner.Handoff)
		want error
	}{
		"writer-running": {func(h *runner.Handoff) { h.WriterStopped = false }, core.ErrEvidenceIncomplete},
		"branch-moved":   {func(h *runner.Handoff) { h.BoundaryHead = headB }, core.ErrStaleHead},
		"no-pr":          {func(h *runner.Handoff) { h.PR = 0 }, core.ErrEvidenceIncomplete},
		"no-worktree":    {func(h *runner.Handoff) { h.Worktree = "" }, core.ErrEvidenceIncomplete},
		// The rest of the receipt (spec 148) is required of every new handoff.
		"no-pr-url":          {func(h *runner.Handoff) { h.PRURL = "" }, core.ErrEvidenceIncomplete},
		"not-an-http-url":    {func(h *runner.Handoff) { h.PRURL = "ftp://github.com/o/r/pull/7" }, core.ErrEvidenceIncomplete},
		"url-of-another-pr":  {func(h *runner.Handoff) { h.PRURL = "https://github.com/o/r/pull/8" }, core.ErrEvidenceIncomplete},
		"no-base":            {func(h *runner.Handoff) { h.Base = "" }, core.ErrEvidenceIncomplete},
		"no-adopted-spec":    {func(h *runner.Handoff) { h.AdoptedSpec = "" }, core.ErrEvidenceIncomplete},
		"no-task-result":     {func(h *runner.Handoff) { h.TaskResult = "" }, core.ErrEvidenceIncomplete},
		"no-check-status":    {func(h *runner.Handoff) { h.CheckStatus = "" }, core.ErrEvidenceIncomplete},
		"worktree-not-there": {func(h *runner.Handoff) { h.Worktree = h.Worktree + "/does-not-exist" }, core.ErrEvidenceIncomplete},
		"relative-worktree":  {func(h *runner.Handoff) { h.Worktree = "wt" }, core.ErrEvidenceIncomplete},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, tk := newRunner(t)
			upTo(t, r, tk, core.StateLanding)
			before := state(t, r, tk)
			h := goodHandoff(t)
			tc.mod(&h)
			if _, err := r.HandOff(tk, h); !errors.Is(err, tc.want) {
				t.Fatalf("handoff: %v, want %v", err, tc.want)
			}
			if after := state(t, r, tk); !reflect.DeepEqual(after, before) {
				t.Errorf("record changed by a refused handoff: %+v -> %+v", before, after)
			}
		})
	}
}

// Repair contention on a RELEASED handed-off Issue (the leader/monitor path): the
// PR belongs to the session that published it, only the receipt's branch is
// repairable, and concurrent repair claims by that session yield exactly one
// holder. While the publisher still owns the Issue nobody else can claim it.
func TestRealHandoff_RepairClaim_BranchBoundAndSingleHolder(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	if _, err := r.HandOff(tk, goodHandoff(t)); err != nil {
		t.Fatal(err)
	}
	run, _ := core.ParseRunID("run-1")
	// The publisher still owns the Issue: another session's repair claim is held off.
	if _, err := r.ClaimRepair(tk.Issue, "fixer", "d-fix", run, branch); !errors.Is(err, core.ErrHeld) {
		t.Errorf("repair claim while the publisher owns the Issue: %v, want ErrHeld", err)
	}
	if err := r.Store.Release(tk.Issue, tk.Holder); err != nil {
		t.Fatal(err)
	}
	// Released: the PR still belongs to the session that published it.
	if _, err := r.ClaimRepair(tk.Issue, "fixer", "d-fix", run, branch); !errors.Is(err, core.ErrNotResponsibleAgent) {
		t.Errorf("repair claim by another session: %v, want ErrNotResponsibleAgent", err)
	}
	// A repair needs the PR reopened first (E10); without it the claim is refused.
	if _, err := r.ClaimRepair(tk.Issue, tk.Holder, "d-fix", run, branch); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("repair before reopen: %v, want ErrTransitionInvalid", err)
	}
	if _, err := r.Reopen(tk.Issue, ""); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("reopen without a reason: %v, want ErrEvidenceIncomplete", err)
	}
	if _, err := r.Reopen(tk.Issue, "required check failed on the PR head"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ClaimRepair(tk.Issue, tk.Holder, "d-fix", run, "some-other-branch"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("repair on a different branch: %v, want ErrTransitionInvalid", err)
	}

	const n = 6
	contend := func(holder func(int) string) []error {
		errs := make([]error, n)
		var wg sync.WaitGroup
		for i := range n {
			wg.Go(func() {
				_, errs[i] = r.ClaimRepair(tk.Issue, holder(i), "d-fix-"+string(rune('a'+i)), run, branch)
			})
		}
		wg.Wait()
		return errs
	}
	wins := func(errs []error) int {
		w := 0
		for _, err := range errs {
			if err == nil {
				w++
			}
		}
		return w
	}
	// Six other sessions: none of them is the publisher, so none wins.
	foreign := contend(func(i int) string { return "fixer-" + string(rune('a'+i)) })
	if w := wins(foreign); w != 0 {
		t.Errorf("repair holders among sessions that did not publish the PR = %d, want 0 (%v)", w, foreign)
	}
	// Six attempts by the publisher's session (distinct dispatches): exactly one holder.
	own := contend(func(int) string { return tk.Holder })
	if w := wins(own); w != 1 {
		t.Errorf("repair holders among the publisher's attempts = %d, want 1 (%v)", w, own)
	}
}

// PR backlog cap: with the cap reached a new claim is held (ErrBacklogFull);
// recording the first PR's done frees the slot.
func TestRealHandoff_BacklogCap_HoldsNewClaimUntilDone(t *testing.T) {
	st, _, run := realStore(t)
	if err := st.SetMaxOpenPRs(1); err != nil {
		t.Fatal(err)
	}
	r := wired(st)
	first := landingFor(t, r, 1948, "w1")
	if _, err := r.HandOff(first, goodHandoff(t)); err != nil {
		t.Fatal(err)
	}
	next, _ := core.NewIssueRef(1949)
	if _, err := r.Claim(next, "w2", "d-w2", run); !errors.Is(err, core.ErrBacklogFull) {
		t.Fatalf("claim at the cap: %v, want ErrBacklogFull", err)
	}
	// The publisher still owns the Issue, so it records done itself (Complete);
	// CompleteHandedOff is the path for a record nobody holds.
	if _, err := r.Complete(first, fullLanding); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Claim(next, "w2", "d-w2", run); err != nil {
		t.Errorf("claim after the PR is done: %v", err)
	}
}

// A handed-off Issue is completed from the merge observation, never without one.
// (The earlier "close only after done" assertion was removed on purpose: S46
// reverses the order, see real_closure_red_test.go.) The leader/monitor path,
// CompleteHandedOff, applies to a record nobody holds, and refuses a held one so
// it cannot race the owner's own writes.
func TestRealHandoff_CompletedFromMergeObservation(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	if _, err := r.HandOff(tk, goodHandoff(t)); err != nil {
		t.Fatal(err)
	}
	unobserved := runner.Landing{RemoteMerged: false, LocalSyncRecorded: true, Deploy: "N/A: x"}
	// Held by its publisher: the monitor path is refused outright.
	before, _ := r.Store.Read(tk.Issue)
	if _, err := r.CompleteHandedOff(tk.Issue, fullLanding); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("CompleteHandedOff on a record its publisher still holds: %v, want ErrTransitionInvalid", err)
	}
	if after, _ := r.Store.Read(tk.Issue); !reflect.DeepEqual(after, before) {
		t.Errorf("a refused monitor completion changed the record: %+v -> %+v", before, after)
	}
	// The owner itself does not complete without an observed merge either.
	if _, err := r.Complete(tk, unobserved); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("owner completed without an observed remote merge: %v, want ErrEvidenceIncomplete", err)
	}
	// Released: the monitor path applies, and still needs the observed merge.
	if err := r.Store.Release(tk.Issue, tk.Holder); err != nil {
		t.Fatal(err)
	}
	if _, err := r.CompleteHandedOff(tk.Issue, unobserved); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("handed-off Issue completed without an observed remote merge: %v, want ErrEvidenceIncomplete", err)
	}
	if rec, _ := r.Store.Read(tk.Issue); rec.State == core.StateDone {
		t.Error("done reached without an observed remote merge")
	}
}

// Agent-bound claims: resume reports the claims the agent still owns.
func TestRealAgent_Resume_ReportsOwnedClaims(t *testing.T) {
	st, ref, run := realStore(t)
	a, err := st.RegisterAgent(sessionID(21))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.AcquireFor(ref, a, "d-1", run); err != nil {
		t.Fatal(err)
	}
	_, owned, err := st.ResumeAgent(a.SessionID, a.Number)
	if err != nil {
		t.Fatal(err)
	}
	if len(owned) != 1 || owned[0].Issue != ref.Number || owned[0].AgentName != a.DisplayName() {
		t.Errorf("owned on resume = %+v", owned)
	}
}
