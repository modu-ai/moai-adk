package e2e

import (
	"errors"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

// PR handoff, repair, backlog cap and closure on lane-18's real code.

const branch = "WT-handoff-branch"

func goodHandoff() runner.Handoff {
	return runner.Handoff{PR: 7, Head: headA, Branch: branch, Worktree: "/tmp/wt", WriterStopped: true, BoundaryHead: headA}
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

// PR published is not done: the handoff releases the claim and keeps the Issue
// in landing; a new claimant is told it is repair-only.
func TestRealHandoff_PRPublished_ReleasesClaimButIsNotDone(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	rec, err := r.HandOff(tk, goodHandoff())
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != core.StateLanding || rec.Receipt == nil || rec.Receipt.PR != 7 {
		t.Fatalf("after handoff: %+v", rec)
	}
	run, _ := core.ParseRunID("run-1")
	if _, err := r.Store.Acquire(tk.Issue, "other", "d-other", run); !errors.Is(err, core.ErrRepairOnly) {
		t.Errorf("fresh claim on a handed-off Issue: %v, want ErrRepairOnly", err)
	}
	if _, err := r.HandOff(tk, goodHandoff()); err == nil {
		t.Error("second handoff by the released ticket was accepted")
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
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, tk := newRunner(t)
			upTo(t, r, tk, core.StateLanding)
			before := state(t, r, tk)
			h := goodHandoff()
			tc.mod(&h)
			if _, err := r.HandOff(tk, h); !errors.Is(err, tc.want) {
				t.Fatalf("handoff: %v, want %v", err, tc.want)
			}
			if after := state(t, r, tk); after != before {
				t.Errorf("record changed by a refused handoff: %+v -> %+v", before, after)
			}
		})
	}
}

// Repair contention: only the receipt's branch is repairable, and concurrent
// repair claimants yield exactly one holder.
func TestRealHandoff_RepairClaim_BranchBoundAndSingleHolder(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	if _, err := r.HandOff(tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	run, _ := core.ParseRunID("run-1")
	if _, err := r.ClaimRepair(tk.Issue, "fixer", "d-fix", run, "some-other-branch"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("repair on a different branch: %v, want ErrTransitionInvalid", err)
	}

	const n = 6
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			_, errs[i] = r.ClaimRepair(tk.Issue, "fixer-"+string(rune('a'+i)), "d-fix-"+string(rune('a'+i)), run, branch)
		})
	}
	wg.Wait()
	wins := 0
	for _, err := range errs {
		if err == nil {
			wins++
		}
	}
	if wins != 1 {
		t.Errorf("repair holders = %d, want 1 (%v)", wins, errs)
	}
}

// PR backlog cap: with the cap reached a new claim is held (ErrBacklogFull);
// recording the first PR's done frees the slot.
func TestRealHandoff_BacklogCap_HoldsNewClaimUntilDone(t *testing.T) {
	st, _, run := realStore(t)
	st.MaxOpenPRs = 1
	r := runner.New(st)
	first := landingFor(t, r, 1948, "w1")
	if _, err := r.HandOff(first, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	next, _ := core.NewIssueRef(1949)
	if _, err := r.Claim(next, "w2", "d-w2", run); !errors.Is(err, core.ErrBacklogFull) {
		t.Fatalf("claim at the cap: %v, want ErrBacklogFull", err)
	}
	land := runner.Landing{RemoteMerged: true, LocalSyncRecorded: true, Deploy: "N/A: no deploy surface"}
	if _, err := r.CompleteHandedOff(first.Issue, land); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Claim(next, "w2", "d-w2", run); err != nil {
		t.Errorf("claim after the PR is done: %v", err)
	}
}

// A handed-off Issue is completed from the merge observation, never by the
// agent that left. (The earlier "close only after done" assertion was removed
// on purpose: S46 reverses the order, see real_closure_red_test.go.)
func TestRealHandoff_CompletedFromMergeObservation(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	if _, err := r.HandOff(tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	if _, err := r.CompleteHandedOff(tk.Issue, runner.Landing{RemoteMerged: false, LocalSyncRecorded: true, Deploy: "N/A: x"}); err == nil {
		t.Error("handed-off Issue completed without an observed remote merge")
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
