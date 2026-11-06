package monitor

import (
	"context"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

type fakePRs struct {
	info  map[int]PRInfo
	open  int
	err   error
	cnErr error
}

func (f fakePRs) ReadPR(_ context.Context, issue int) (PRInfo, bool, error) {
	if f.err != nil {
		return PRInfo{}, false, f.err
	}
	p, ok := f.info[issue]
	return p, ok, nil
}

func (f fakePRs) CountOpenPRs(context.Context) (int, error) { return f.open, f.cnErr }

func withPR(prs PRReader, limit int) *Collector {
	c := collector(
		fakeClaims{rows: map[int]ClaimRow{7: {Issue: 7, State: "verifying", Stage: string(core.StageBuilding), HeadSHA: head}, 8: {Issue: 8, State: "done", Stage: string(core.StageDone), HeadSHA: head}}},
		fakeRemote{}, fakeLedger{},
	)
	c.PRs, c.BacklogLimit = prs, limit
	return c
}

func TestOpenPRIsNeverComplete(t *testing.T) {
	green := []CheckRun{{Name: "build", Conclusion: "success", HeadSHA: head}}
	c := withPR(fakePRs{info: map[int]PRInfo{7: {Number: 55, Open: true, HeadSHA: head, Checks: green, ReviewVerdict: "PASS"}}}, 5)
	r := c.Collect(context.Background(), []int{7})[0]
	if r.PR.Value.Number != 55 || r.PR.Source != SrcRemote {
		t.Fatalf("pr = %+v", r.PR)
	}
	if r.Completed {
		t.Error("an open PR must not show as completed")
	}
	if r.PR.Value.Phase != PhaseReadyToMerge {
		t.Errorf("phase = %s", r.PR.Value.Phase)
	}
}

func TestPRPhases(t *testing.T) {
	ok := []CheckRun{{Name: "b", Conclusion: "success", HeadSHA: head}}
	bad := []CheckRun{{Name: "b", Conclusion: "failure", HeadSHA: head}}
	stale := []CheckRun{{Name: "b", Conclusion: "success", HeadSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}
	cases := []struct {
		name string
		in   PRInfo
		want PRPhase
	}{
		{"no checks yet", PRInfo{Open: true, HeadSHA: head}, PhaseChecksPending},
		{"checks failed", PRInfo{Open: true, HeadSHA: head, Checks: bad}, PhaseChecksFailed},
		{"stale checks", PRInfo{Open: true, HeadSHA: head, Checks: stale}, PhaseChecksPending},
		{"review pending", PRInfo{Open: true, HeadSHA: head, Checks: ok}, PhaseReviewPending},
		{"review failed repair dispatched", PRInfo{Open: true, HeadSHA: head, Checks: ok, ReviewVerdict: "FAIL", RepairDispatched: true}, PhaseRepairDispatched},
		{"review failed no repair", PRInfo{Open: true, HeadSHA: head, Checks: ok, ReviewVerdict: "FAIL"}, PhaseRepairNeeded},
		{"merged issue open", PRInfo{Merged: true, HeadSHA: head}, PhaseMergedIssueOpen},
		{"merged issue closed", PRInfo{Merged: true, IssueClosed: true, HeadSHA: head}, PhaseIssueClosed},
	}
	for _, c := range cases {
		if got := Phase(c.in); got != c.want {
			t.Errorf("%s: phase = %s, want %s", c.name, got, c.want)
		}
	}
}

func TestCompletedNeedsMergeClosedIssueAndDoneState(t *testing.T) {
	merged := PRInfo{Number: 55, Merged: true, IssueClosed: true, HeadSHA: head}
	c := withPR(fakePRs{info: map[int]PRInfo{7: merged, 8: merged}}, 5)
	rows := c.Collect(context.Background(), []int{7, 8})
	if rows[0].Completed {
		t.Error("claim state verifying: merge + closed issue alone must not complete")
	}
	if !rows[1].Completed {
		t.Error("done + merged + closed should be completed")
	}
}

func TestPRUnreadableIsUnconfirmed(t *testing.T) {
	c := withPR(fakePRs{err: errors.New("gh 502")}, 5)
	r := c.Collect(context.Background(), []int{7})[0]
	if r.PR.Confirmed || r.Completed {
		t.Errorf("pr = %+v completed=%v", r.PR, r.Completed)
	}
}

func TestBacklogStatus(t *testing.T) {
	for _, c := range []struct {
		open, limit int
		want        BacklogState
	}{{3, 5, BacklogWithin}, {5, 5, BacklogAtLimit}, {6, 5, BacklogOver}, {4, 0, BacklogNoLimit}} {
		if got := backlogState(c.open, c.limit); got != c.want {
			t.Errorf("open %d limit %d = %s, want %s", c.open, c.limit, got, c.want)
		}
	}
	v := withPR(fakePRs{open: 6}, 5).View(context.Background(), nil)
	if v.Backlog.Value.Open != 6 || v.Backlog.Value.Limit != 5 || v.Backlog.Value.State != BacklogOver || !v.Backlog.Confirmed {
		t.Errorf("backlog = %+v", v.Backlog)
	}
	bad := withPR(fakePRs{cnErr: errors.New("x")}, 5).View(context.Background(), nil)
	if bad.Backlog.Confirmed {
		t.Error("unreadable count must be unconfirmed")
	}
}

func TestNilPRReaderLeavesPRUnconfirmed(t *testing.T) {
	c := withPR(nil, 5)
	r := c.Collect(context.Background(), []int{7})[0]
	if r.PR.Confirmed {
		t.Error("no reader: pr must be unconfirmed")
	}
}
