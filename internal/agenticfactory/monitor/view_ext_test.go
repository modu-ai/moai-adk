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

func TestIntakeItemsUseTheCanonicalStagesAndDeriveTheirState(t *testing.T) {
	v := extCollector(fakeExt{intake: []IntakeItem{
		{Key: "k1", Issue: 7, Stage: StageIssueConfirmed},
		{Key: "k2", Stage: StageAccepted, Parked: "backpressure"},
		{Key: "k3", Stage: StageBranchRegistered, Block: &IntakeBlock{Step: StageBranchRegistered, Detail: "push refused token=abc123def456ghi789", Retryable: true, Resume: "retry after the remote answers"}},
		{Key: "k4", Stage: StageAccepted, Block: &IntakeBlock{Step: StageAccepted, Detail: "lost response", Unknown: true, Resume: "FindIssue decides"}},
		{Key: "k5", Issue: 9, Stage: StageReadyAcked, Started: true},
	}}).View(context.Background(), nil)
	if !v.Intake.Confirmed || len(v.Intake.Value) != 5 || v.Intake.Source != SrcClaimStore {
		t.Fatalf("intake = %+v", v.Intake)
	}
	want := []string{"ISSUE_CONFIRMED", "PARKED", "BLOCKED", "EFFECT_UNKNOWN", "READY_ACKED"}
	for i, it := range v.Intake.Value {
		if it.State != want[i] {
			t.Errorf("item %d state = %s, want %s", i, it.State, want[i])
		}
	}
	if strings.Contains(v.Intake.Value[2].Block.Detail, "abc123def456ghi789") {
		t.Errorf("block detail leaks a secret: %q", v.Intake.Value[2].Block.Detail)
	}
	if !v.Intake.Value[4].Started || v.Intake.Value[4].Stage != StageReadyAcked {
		t.Error("started must be shown beside the stage, not as a stage")
	}
}

// The monitor has no stage vocabulary of its own: only the six canonical stage names exist.
func TestNoStageOutsideTheCanonicalSixExists(t *testing.T) {
	for _, st := range []IntakeStage{StageAccepted, StageIssueConfirmed, StageBranchRegistered, StageAgentReserved, StageWTPrepared, StageReadyAcked} {
		if string(st) != strings.ToUpper(string(st)) || strings.ContainsAny(string(st), "- ") {
			t.Errorf("stage %q is not in the canonical SHOUTING_SNAKE form", st)
		}
	}
	for _, old := range []string{"ack", "issue-created", "wt-ready", "started", "failed", "parked"} {
		if _, err := marshalStage(old); err == nil {
			t.Errorf("old stage name %q still accepted", old)
		}
	}
}

func marshalStage(s string) (IntakeStage, error) {
	for _, st := range []IntakeStage{StageAccepted, StageIssueConfirmed, StageBranchRegistered, StageAgentReserved, StageWTPrepared, StageReadyAcked} {
		if string(st) == s {
			return st, nil
		}
	}
	return "", errors.New("not a canonical stage")
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

func TestLinkIsAvailableOnlyAfterRemoteVerification(t *testing.T) {
	c := extCollector(fakeExt{ctx: map[int][]ContextItem{7: {
		{Section: "a", Link: ".moai/evidence/issue-7/a.md", State: CtxCurrent, Stage: StageRemoteVerified, RemoteVerified: true, Availability: LinkAvailable},
		{Section: "b", Link: ".moai/evidence/issue-7/b.md", State: CtxCurrent, RemoteVerified: false, Availability: LinkAvailable}, // reader claims available without a check
		{Section: "c", Link: ".moai/evidence/issue-7/c.md", State: CtxInaccessible, RemoteVerified: true},
	}}})
	items := c.Collect(context.Background(), []int{7})[0].Context.Value
	want := []LinkAvailability{LinkAvailable, LinkUnverified, LinkInaccessible}
	for i, it := range items {
		if it.Availability != want[i] {
			t.Errorf("item %s: availability = %s, want %s", it.Section, it.Availability, want[i])
		}
	}
}

func TestEvidenceStageIsShownAndOnlyRemoteVerifiedIsAvailable(t *testing.T) {
	mk := func(stage EvidenceStage, verified bool, st ContextState) ContextItem {
		return ContextItem{Section: string(stage), Link: "l", State: st, Stage: stage, RemoteVerified: verified}
	}
	c := extCollector(fakeExt{ctx: map[int][]ContextItem{7: {
		mk(StagePrepared, false, CtxCurrent),
		mk(StageMasked, false, CtxCurrent),
		mk(StageCommitted, false, CtxCurrent),
		mk(StagePushed, true, CtxCurrent),          // pushed is not remote-verified, whatever the flag says
		mk(StageRemoteVerified, true, CtxCurrent),  // the only available one
		mk(StageRemoteVerified, false, CtxCurrent), // claims verified stage without the lookup flag
		mk(StageRestricted, true, CtxCurrent),
		mk(EvidenceStage("bogus"), true, CtxCurrent),
		mk(StageRemoteVerified, true, CtxInaccessible),
	}}})
	items := c.Collect(context.Background(), []int{7})[0].Context.Value
	want := []LinkAvailability{LinkUnverified, LinkUnverified, LinkUnverified, LinkUnverified, LinkAvailable, LinkUnverified, LinkRestricted, LinkUnverified, LinkInaccessible}
	for i, it := range items {
		if it.Availability != want[i] {
			t.Errorf("#%d stage=%s verified=%v: availability = %s, want %s", i, it.Stage, it.RemoteVerified, it.Availability, want[i])
		}
	}
	if items[7].Stage != StageUnknown {
		t.Errorf("unknown stage kept as %q", items[7].Stage)
	}
}
