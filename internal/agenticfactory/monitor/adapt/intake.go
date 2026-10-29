package adapt

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

// IntakeSource implements monitor.IntakeReader over the effect intent store. Only the intake
// effects (Issue creation, branch and worktree preparation) appear; ledger comments and check
// observations are other views.
type IntakeSource struct{ Store effect.IntentStore }

func (s IntakeSource) ReadIntake(context.Context) ([]monitor.IntakeItem, error) {
	all, err := s.Store.List()
	if err != nil {
		return nil, err
	}
	var out []monitor.IntakeItem
	for _, in := range all {
		if in.Kind != effect.IssueCreate && in.Kind != effect.BranchPrepare && in.Kind != effect.WorktreePrepare {
			continue
		}
		stage := monitor.StageAck
		switch in.Status {
		case effect.StatusConfirmed:
			stage = monitor.StageIssueCreated
			if in.Kind == effect.WorktreePrepare {
				stage = monitor.StageWTReady
			}
		case effect.StatusFailed:
			stage = monitor.StageFailed
		case effect.StatusParked:
			stage = monitor.StageParked
		}
		out = append(out, monitor.IntakeItem{Issue: in.Issue, Stage: stage, State: string(in.Status), Reason: in.LastError})
	}
	return out, nil
}
