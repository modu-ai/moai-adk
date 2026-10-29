package monitor

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeExt struct {
	intake    []IntakeItem
	intakeErr error
	ctx       map[int][]ContextItem
	ctxErr    error
	queue     QueueStats
	queueErr  error
}

func (f fakeExt) ReadIntake(context.Context) ([]IntakeItem, error) { return f.intake, f.intakeErr }
func (f fakeExt) ReadContext(_ context.Context, issue int) ([]ContextItem, error) {
	return f.ctx[issue], f.ctxErr
}
func (f fakeExt) ReadQueue(context.Context) (QueueStats, error) { return f.queue, f.queueErr }

func extCollector(e fakeExt) *Collector {
	c := withPR(fakePRs{open: 6}, 5)
	c.Intake, c.Context, c.Queue = e, e, e
	return c
}

func TestIntakeStagesAreShownWithReasonAndAreMasked(t *testing.T) {
	v := extCollector(fakeExt{intake: []IntakeItem{
		{Issue: 7, Stage: StageIssueCreated, State: "confirmed"},
		{Issue: 8, Stage: StageParked, State: "parked", Reason: "lookup unsure token=abc123def456ghi789"},
	}}).View(context.Background(), nil)
	if !v.Intake.Confirmed || len(v.Intake.Value) != 2 || v.Intake.Source != SrcClaimStore {
		t.Fatalf("intake = %+v", v.Intake)
	}
	if strings.Contains(v.Intake.Value[1].Reason, "abc123def456ghi789") {
		t.Errorf("reason leaks secret: %q", v.Intake.Value[1].Reason)
	}
}

func TestContextItemsDistinguishStaleFromInaccessible(t *testing.T) {
	c := extCollector(fakeExt{ctx: map[int][]ContextItem{7: {
		{Section: "spec", Link: "https://x/1", TargetHead: head, Hash: "h1", Owner: "agent-1", State: CtxCurrent},
		{Section: "audit", Link: "https://x/2", TargetHead: head, Hash: "h2", Owner: "agent-2", State: CtxStale},
		{Section: "qa", Link: "https://x/3", State: CtxInaccessible},
	}}})
	r := c.Collect(context.Background(), []int{7})[0]
	if !r.Context.Confirmed || len(r.Context.Value) != 3 {
		t.Fatalf("context = %+v", r.Context)
	}
	got := map[ContextState]bool{}
	for _, it := range r.Context.Value {
		got[it.State] = true
	}
	if !got[CtxCurrent] || !got[CtxStale] || !got[CtxInaccessible] {
		t.Errorf("states = %v", got)
	}
}

func TestQueueStatsJoinTheBacklog(t *testing.T) {
	v := extCollector(fakeExt{queue: QueueStats{AwaitingHandoff: 2, RepairBacklog: 1, OldestWaitSeconds: 900}}).View(context.Background(), nil)
	b := v.Backlog.Value
	if b.Open != 6 || b.AwaitingHandoff != 2 || b.RepairBacklog != 1 || b.OldestWaitSeconds != 900 {
		t.Errorf("backlog = %+v", b)
	}
}

func TestMissingOrFailingReadersAreUnconfirmedNotEmpty(t *testing.T) {
	v := extCollector(fakeExt{intakeErr: errors.New("x"), ctxErr: errors.New("y"), queueErr: errors.New("z")}).View(context.Background(), nil)
	if v.Intake.Confirmed {
		t.Error("intake must be unconfirmed")
	}
	if !v.Backlog.Confirmed || v.Backlog.Value.AwaitingHandoff != 0 || v.Backlog.Value.QueueKnown {
		t.Errorf("queue unreadable must be flagged: %+v", v.Backlog.Value)
	}
	r := extCollector(fakeExt{ctxErr: errors.New("y")}).Collect(context.Background(), []int{7})[0]
	if r.Context.Confirmed {
		t.Error("context must be unconfirmed")
	}
	bare := withPR(fakePRs{open: 1}, 5).View(context.Background(), nil)
	if bare.Intake.Confirmed {
		t.Error("no intake reader: unconfirmed")
	}
}
