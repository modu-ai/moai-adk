package adapt

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

// RequestSource implements monitor.IntakeReader over lane 18's durable intake request records. It only
// calls ViewRequests, which returns the stored fields plus Started derived from the claim and writes
// nothing. The stage names are the canonical S51 ones and are passed through unchanged.
type RequestSource struct{ Store *claim.Store }

func (s RequestSource) ReadIntake(context.Context) ([]monitor.IntakeItem, error) {
	reqs, err := s.Store.ViewRequests()
	if err != nil {
		return nil, err
	}
	out := make([]monitor.IntakeItem, 0, len(reqs))
	for _, r := range reqs {
		it := monitor.IntakeItem{
			Key: r.Key, Issue: r.Issue, Stage: monitor.IntakeStage(r.Stage),
			Parked: r.Parked, Started: r.Started, Agent: r.Agent, Slot: r.Slot,
		}
		if r.Blocked != nil {
			it.Block = &monitor.IntakeBlock{
				Step: monitor.IntakeStage(r.Blocked.Step), Detail: r.Blocked.Detail,
				Retryable: r.Blocked.Retryable, Unknown: r.Blocked.Unknown, Resume: r.Blocked.Resume,
			}
		}
		it.State = monitor.DeriveIntakeState(it)
		out = append(out, it)
	}
	return out, nil
}
