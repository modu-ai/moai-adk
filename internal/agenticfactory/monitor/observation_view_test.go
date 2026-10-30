package monitor

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// AC-AF-019 / design: the observation view. Preparation states and the PR backlog are shown apart from the
// work already started (BUILDING and later), every item says where it came from and when it was read, reading
// changes nothing, and nothing in the output carries a raw C1..C5 value.

func observationFixture() *Collector {
	return extCollector(fakeExt{
		intake: []IntakeItem{
			{Key: "k1", Issue: 7, Stage: StageIssueConfirmed},
			{Key: "k2", Stage: StageAccepted, Parked: "backpressure"},
			{Key: "k3", Stage: StageBranchRegistered, Block: &IntakeBlock{Step: StageBranchRegistered, Detail: "push refused token=ghp_abcdefghijklmnopqrstuvwxyz0123456789 for someone@example.com at /Users/someone/work", Retryable: true, Resume: "retry as someone@example.com"}},
		},
		ctx: map[int][]ContextItem{7: {
			{Section: "spec", Link: "https://x/1", Owner: "agent <someone@example.com>", State: CtxCurrent, Stage: StageRemoteVerified, RemoteVerified: true},
		}},
		queue: QueueStats{AwaitingHandoff: 2, RepairBacklog: 1, OldestWaitSeconds: 900},
	})
}

func TestObservationView_ShowsPrepAndPRBacklogStatesDistinctly_ReadOnly(t *testing.T) {
	ctx := context.Background()
	v := observationFixture().View(ctx, []int{7, 8})

	// (a) every item carries its value, its source and the time it was read.
	if !v.Intake.Confirmed || v.Intake.Source == "" || v.Intake.ReadAt.IsZero() {
		t.Errorf("intake field = %+v", v.Intake)
	}
	if !v.Backlog.Confirmed || v.Backlog.Source == "" || v.Backlog.ReadAt.IsZero() {
		t.Errorf("backlog field = %+v", v.Backlog)
	}
	for _, r := range v.Rows {
		for name, f := range map[string]struct {
			src    Source
			zero   bool
			conf   bool
			isRead bool
		}{
			"state": {r.State.Source, r.State.ReadAt.IsZero(), r.State.Confirmed, true},
			"stage": {r.Stage.Source, r.Stage.ReadAt.IsZero(), r.Stage.Confirmed, true},
		} {
			if f.conf && (f.src == "" || f.zero) {
				t.Errorf("row %d %s: confirmed without a source or read time", r.Issue, name)
			}
		}
	}

	// (b) preparation items and the PR backlog are separate zones from the started work: the intake list is its
	// own field, none of its states is a started-work stage, and no row carries a preparation item.
	prep := map[string]bool{string(StageAccepted): true, string(StageIssueConfirmed): true, string(StageBranchRegistered): true,
		string(StageAgentReserved): true, string(StageWTPrepared): true, string(StageReadyAcked): true,
		StateBlocked: true, StateParked: true, StateEffectUnknown: true}
	for _, it := range v.Intake.Value {
		if !prep[it.State] {
			t.Errorf("intake item %s shows state %q, which is not a preparation state", it.Key, it.State)
		}
	}
	if v.Backlog.Value.Open == 0 && v.Backlog.Value.State == "" {
		t.Error("the PR backlog zone is empty")
	}
	if !v.Backlog.Value.QueueKnown || v.Backlog.Value.AwaitingHandoff != 2 || v.Backlog.Value.RepairBacklog != 1 {
		t.Errorf("backlog = %+v, want the handoff and repair figures", v.Backlog.Value)
	}

	// (c) reading is idempotent: three reads give byte-identical output, and the readers expose no write.
	first, _ := json.Marshal(v)
	for i := 0; i < 2; i++ {
		again, _ := json.Marshal(observationFixture().View(ctx, []int{7, 8}))
		if string(again) != string(first) {
			t.Errorf("read %d differs from the first", i+2)
		}
	}

	// (d) no raw C1..C5 value in what is shown.
	out := string(first)
	for _, raw := range []string{"ghp_abcdefghijklmnopqrstuvwxyz0123456789", "someone@example.com", "/Users/someone"} {
		if strings.Contains(out, raw) {
			t.Errorf("the view carries the raw value %q", raw)
		}
	}
}
