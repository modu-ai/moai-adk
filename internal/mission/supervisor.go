package mission

import (
	"context"
	"errors"
	"fmt"
)

type SupervisionStep struct {
	Action Action `json:"action"`
	Target string `json:"target"`
}

type SupervisionPlan struct {
	MissionID, ContractHash string
	MaxOperations           int
	Steps                   []SupervisionStep
	CompletionEvidence      []string
	RequireLandedAncestry   bool
}

type SupervisionSnapshot struct{ Hash string }

type SupervisorState struct {
	MissionID, ContractHash string
	State                   MissionState
	NextStep                int
	OperationIDs            []string
	Blocker                 string
	AskUserQuestion         bool
}

type SupervisorStateStore interface {
	Load(context.Context, string) (SupervisorState, error)
	Save(context.Context, SupervisorState) error
}

// SupervisorEngine splits every iteration into the required trust boundaries.
// Governance is evidence loading, Validate is deterministic policy, Execute is
// the owning role, and Readback is authoritative observation after the effect.
type SupervisorEngine interface {
	Snapshot(context.Context, SupervisionStep) (SupervisionSnapshot, error)
	Governance(context.Context, SupervisionStep, SupervisionSnapshot) error
	Validate(context.Context, SupervisionStep, SupervisionSnapshot) error
	Execute(context.Context, SupervisionStep, SupervisionSnapshot) (OperationReceipt, error)
	Readback(context.Context, SupervisionStep, OperationReceipt) error
	Finalize(context.Context, SupervisionPlan, SupervisorState) error
}

func blockSupervisor(ctx context.Context, store SupervisorStateStore, state SupervisorState, stage string, err error) (SupervisorState, error) {
	state.State = StateBlocked
	state.Blocker = "mission supervisor: " + stage
	state.AskUserQuestion = false
	if saveErr := store.Save(ctx, state); saveErr != nil {
		return state, errors.Join(err, saveErr)
	}
	return state, fmt.Errorf("%s: %w", state.Blocker, err)
}

// SuperviseAutoMission runs a bounded, persist-before-advance loop. It never
// prompts: a missing capability or authority is durably blocked for later
// explicit re-approval instead of being broadened in process.
func SuperviseAutoMission(ctx context.Context, plan SupervisionPlan, store SupervisorStateStore, engine SupervisorEngine) (SupervisorState, error) {
	if store == nil || engine == nil || plan.MissionID == "" || plan.ContractHash == "" || plan.MaxOperations <= 0 || len(plan.Steps) == 0 || len(plan.Steps) > plan.MaxOperations || len(plan.CompletionEvidence) == 0 {
		return SupervisorState{}, errors.New("mission supervisor: invalid_plan")
	}
	state, err := store.Load(ctx, plan.MissionID)
	if err != nil {
		return SupervisorState{}, err
	}
	if state.MissionID == plan.MissionID && state.ContractHash == plan.ContractHash && state.State == StateCompleted && state.NextStep == len(plan.Steps) && len(state.OperationIDs) == len(plan.Steps) {
		if err := engine.Finalize(ctx, plan, state); err != nil {
			return blockSupervisor(ctx, store, state, "completion", err)
		}
		return state, nil
	}
	if state.MissionID != plan.MissionID || state.ContractHash != plan.ContractHash || (state.State != StateApproved && state.State != StateRunning) || state.NextStep < 0 || state.NextStep > len(plan.Steps) {
		return blockSupervisor(ctx, store, state, "lineage_mismatch", errors.New("invalid persisted state"))
	}
	state.State, state.Blocker, state.AskUserQuestion = StateRunning, "", false
	if err := store.Save(ctx, state); err != nil {
		return state, err
	}
	for state.NextStep < len(plan.Steps) {
		if err := ctx.Err(); err != nil {
			return blockSupervisor(ctx, store, state, "context_cancelled", err)
		}
		step := plan.Steps[state.NextStep]
		snapshot, err := engine.Snapshot(ctx, step)
		if err != nil || snapshot.Hash == "" {
			if err == nil {
				err = errors.New("empty snapshot")
			}
			return blockSupervisor(ctx, store, state, "snapshot", err)
		}
		if err := store.Save(ctx, state); err != nil {
			return state, err
		}
		if err := engine.Governance(ctx, step, snapshot); err != nil {
			return blockSupervisor(ctx, store, state, "governance", err)
		}
		if err := store.Save(ctx, state); err != nil {
			return state, err
		}
		if err := engine.Validate(ctx, step, snapshot); err != nil {
			return blockSupervisor(ctx, store, state, "validate", err)
		}
		receipt, err := engine.Execute(ctx, step, snapshot)
		if err != nil {
			return blockSupervisor(ctx, store, state, "owner", err)
		}
		if receipt.OperationID == "" || receipt.Action != step.Action {
			return blockSupervisor(ctx, store, state, "operation_lineage", errors.New("invalid owner receipt"))
		}
		if err := engine.Readback(ctx, step, receipt); err != nil {
			return blockSupervisor(ctx, store, state, "readback", err)
		}
		state.OperationIDs = append(state.OperationIDs, receipt.OperationID)
		state.NextStep++
		if err := store.Save(ctx, state); err != nil {
			return state, err
		}
	}
	if err := engine.Finalize(ctx, plan, state); err != nil {
		return blockSupervisor(ctx, store, state, "completion", err)
	}
	state.State = StateCompleted
	if err := store.Save(ctx, state); err != nil {
		return state, err
	}
	return state, nil
}
