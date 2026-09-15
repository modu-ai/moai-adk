package mission

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type memorySupervisorStore struct {
	state     SupervisorState
	saves     int
	loadErr   error
	saveErrAt int
}

func (s *memorySupervisorStore) Load(context.Context, string) (SupervisorState, error) {
	return s.state, s.loadErr
}
func (s *memorySupervisorStore) Save(_ context.Context, state SupervisorState) error {
	s.state = state
	s.saves++
	if s.saveErrAt == s.saves {
		return errors.New("save failed")
	}
	return nil
}

type recordingSupervisorEngine struct {
	calls      []string
	failAt     int
	completion bool
}

func (e *recordingSupervisorEngine) Snapshot(_ context.Context, step SupervisionStep) (SupervisionSnapshot, error) {
	e.calls = append(e.calls, "snapshot:"+string(step.Action))
	return SupervisionSnapshot{Hash: "snap-" + string(step.Action)}, nil
}
func (e *recordingSupervisorEngine) Governance(_ context.Context, step SupervisionStep, snapshot SupervisionSnapshot) error {
	e.calls = append(e.calls, "governance:"+string(step.Action))
	return nil
}
func (e *recordingSupervisorEngine) Validate(_ context.Context, step SupervisionStep, snapshot SupervisionSnapshot) error {
	e.calls = append(e.calls, "validate:"+string(step.Action))
	return nil
}
func (e *recordingSupervisorEngine) Execute(_ context.Context, step SupervisionStep, snapshot SupervisionSnapshot) (OperationReceipt, error) {
	e.calls = append(e.calls, "owner:"+string(step.Action))
	if e.failAt == len(e.calls) {
		return OperationReceipt{}, errors.New("owner blocked")
	}
	return OperationReceipt{OperationID: "op-" + string(step.Action), Action: step.Action, State: ReceiptReconciled}, nil
}
func (e *recordingSupervisorEngine) Readback(_ context.Context, step SupervisionStep, _ OperationReceipt) error {
	e.calls = append(e.calls, "readback:"+string(step.Action))
	return nil
}
func (e *recordingSupervisorEngine) Finalize(_ context.Context, plan SupervisionPlan, _ SupervisorState) error {
	e.calls = append(e.calls, "completion")
	if !e.completion || len(plan.CompletionEvidence) == 0 {
		return errors.New("completion evidence missing")
	}
	return nil
}

func TestSuperviseAutoMissionPersistsBoundedOrderedLoopAndFinalizes(t *testing.T) {
	steps := []SupervisionStep{{Action: ActionPublish, Target: "gtd:x"}, {Action: ActionPick, Target: "gtd:x"}, {Action: ActionDispatch, Target: "gtd:x"}}
	store := &memorySupervisorStore{state: SupervisorState{MissionID: "m", ContractHash: "h", State: StateApproved}}
	engine := &recordingSupervisorEngine{completion: true}
	got, err := SuperviseAutoMission(context.Background(), SupervisionPlan{MissionID: "m", ContractHash: "h", MaxOperations: 3, Steps: steps, CompletionEvidence: []string{"done"}}, store, engine)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != StateCompleted || got.NextStep != 3 || len(got.OperationIDs) != 3 || store.saves < 7 {
		t.Fatalf("state=%+v saves=%d", got, store.saves)
	}
	want := []string{"snapshot:publish", "governance:publish", "validate:publish", "owner:publish", "readback:publish", "snapshot:pick", "governance:pick", "validate:pick", "owner:pick", "readback:pick", "snapshot:dispatch", "governance:dispatch", "validate:dispatch", "owner:dispatch", "readback:dispatch", "completion"}
	if !reflect.DeepEqual(engine.calls, want) {
		t.Fatalf("calls=%v want=%v", engine.calls, want)
	}
}

func TestSuperviseAutoMissionCompletedReplayHasNoEffects(t *testing.T) {
	steps := []SupervisionStep{{Action: ActionPublish, Target: "gtd:x"}}
	store := &memorySupervisorStore{state: SupervisorState{
		MissionID:    "m",
		ContractHash: "h",
		State:        StateCompleted,
		NextStep:     1,
		OperationIDs: []string{"op-publish"},
	}}
	engine := &recordingSupervisorEngine{completion: true}

	got, err := SuperviseAutoMission(context.Background(), SupervisionPlan{
		MissionID: "m", ContractHash: "h", MaxOperations: 1, Steps: steps, CompletionEvidence: []string{"done"},
	}, store, engine)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != StateCompleted || !reflect.DeepEqual(engine.calls, []string{"completion"}) || store.saves != 0 {
		t.Fatalf("completed replay changed state: state=%+v calls=%v saves=%d", got, engine.calls, store.saves)
	}
}

func TestSuperviseAutoMissionPersistsBlockedWithoutQuestionOrFurtherEffect(t *testing.T) {
	store := &memorySupervisorStore{state: SupervisorState{MissionID: "m", ContractHash: "h", State: StateApproved}}
	engine := &recordingSupervisorEngine{failAt: 4}
	got, err := SuperviseAutoMission(context.Background(), SupervisionPlan{MissionID: "m", ContractHash: "h", MaxOperations: 2, Steps: []SupervisionStep{{Action: ActionPublish, Target: "gtd:x"}, {Action: ActionPick, Target: "gtd:x"}}, CompletionEvidence: []string{"done"}}, store, engine)
	if err == nil || got.State != StateBlocked || got.Blocker == "" || got.NextStep != 0 {
		t.Fatalf("state=%+v err=%v", got, err)
	}
	if got.AskUserQuestion {
		t.Fatal("auto supervisor requested a question")
	}
	if len(engine.calls) != 4 {
		t.Fatalf("effects continued after block: %v", engine.calls)
	}
}

func TestSuperviseAutoMissionDoesNotCompleteWithoutContractEvidence(t *testing.T) {
	store := &memorySupervisorStore{state: SupervisorState{MissionID: "m", ContractHash: "h", State: StateApproved}}
	state, err := SuperviseAutoMission(context.Background(), SupervisionPlan{
		MissionID: "m", ContractHash: "h", MaxOperations: 1,
		Steps: []SupervisionStep{{Action: ActionPublish, Target: "gtd:x"}},
	}, store, &recordingSupervisorEngine{})
	if err == nil || state.State == StateCompleted {
		t.Fatalf("step exhaustion completed without contract evidence: state=%+v err=%v", state, err)
	}
}

type failingSupervisorEngine struct{ stage string }

func (e failingSupervisorEngine) Snapshot(context.Context, SupervisionStep) (SupervisionSnapshot, error) {
	if e.stage == "snapshot" {
		return SupervisionSnapshot{}, errors.New("snapshot failed")
	}
	if e.stage == "empty_snapshot" {
		return SupervisionSnapshot{}, nil
	}
	return SupervisionSnapshot{Hash: "snapshot"}, nil
}
func (e failingSupervisorEngine) Governance(context.Context, SupervisionStep, SupervisionSnapshot) error {
	if e.stage == "governance" {
		return errors.New("governance failed")
	}
	return nil
}
func (e failingSupervisorEngine) Validate(context.Context, SupervisionStep, SupervisionSnapshot) error {
	if e.stage == "validate" {
		return errors.New("validate failed")
	}
	return nil
}
func (e failingSupervisorEngine) Execute(context.Context, SupervisionStep, SupervisionSnapshot) (OperationReceipt, error) {
	if e.stage == "owner" {
		return OperationReceipt{}, errors.New("owner failed")
	}
	if e.stage == "lineage" {
		return OperationReceipt{OperationID: "wrong", Action: ActionPick}, nil
	}
	return OperationReceipt{OperationID: "op", Action: ActionPublish}, nil
}
func (e failingSupervisorEngine) Readback(context.Context, SupervisionStep, OperationReceipt) error {
	if e.stage == "readback" {
		return errors.New("readback failed")
	}
	return nil
}
func (e failingSupervisorEngine) Finalize(context.Context, SupervisionPlan, SupervisorState) error {
	if e.stage == "completion" {
		return errors.New("completion failed")
	}
	return nil
}

func TestSuperviseAutoMissionFailClosedBranches(t *testing.T) {
	plan := SupervisionPlan{MissionID: "m", ContractHash: "h", MaxOperations: 1, Steps: []SupervisionStep{{Action: ActionPublish, Target: "gtd:x"}}, CompletionEvidence: []string{"done"}}
	for _, stage := range []string{"snapshot", "empty_snapshot", "governance", "validate", "owner", "lineage", "readback", "completion"} {
		t.Run(stage, func(t *testing.T) {
			store := &memorySupervisorStore{state: SupervisorState{MissionID: "m", ContractHash: "h", State: StateApproved}}
			state, err := SuperviseAutoMission(context.Background(), plan, store, failingSupervisorEngine{stage: stage})
			if err == nil || state.State != StateBlocked || state.AskUserQuestion {
				t.Fatalf("state=%+v err=%v", state, err)
			}
		})
	}
	if _, err := SuperviseAutoMission(context.Background(), SupervisionPlan{}, nil, nil); err == nil {
		t.Fatal("invalid plan accepted")
	}
	loadStore := &memorySupervisorStore{loadErr: errors.New("load failed")}
	if _, err := SuperviseAutoMission(context.Background(), plan, loadStore, failingSupervisorEngine{}); err == nil {
		t.Fatal("load error ignored")
	}
	badLineage := &memorySupervisorStore{state: SupervisorState{MissionID: "other", ContractHash: "h", State: StateApproved}}
	if state, err := SuperviseAutoMission(context.Background(), plan, badLineage, failingSupervisorEngine{}); err == nil || state.State != StateBlocked {
		t.Fatalf("lineage state=%+v err=%v", state, err)
	}
	initialSave := &memorySupervisorStore{state: SupervisorState{MissionID: "m", ContractHash: "h", State: StateApproved}, saveErrAt: 1}
	if _, err := SuperviseAutoMission(context.Background(), plan, initialSave, failingSupervisorEngine{}); err == nil {
		t.Fatal("initial save error ignored")
	}
	blockSave := &memorySupervisorStore{state: SupervisorState{MissionID: "m", ContractHash: "h", State: StateApproved}, saveErrAt: 2}
	if state, err := SuperviseAutoMission(context.Background(), plan, blockSave, failingSupervisorEngine{stage: "snapshot"}); err == nil || state.State != StateBlocked {
		t.Fatalf("block save state=%+v err=%v", state, err)
	}
}
