package adapt

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/reconcile"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Holder identifies the process and worktree that own an attempt.
type Holder struct {
	PID          int
	ProcessStart string // fingerprint recorded at claim time
	Worktree     string
}

// Sources are the read-only lookups behind a reconcile.Prober. The caller wires them to the claim
// store, git and the remote; any error makes the whole probe fail, which reconcile reports as
// "unconfirmed" rather than guessing.
type Sources struct {
	Holder           func(issue int) (Holder, error)
	Classify         func(pid int, processStart string) reconcile.Liveness
	WorktreeHead     func(path string) (string, error)
	RemoteHead       func(ctx context.Context, issue int) (string, error)
	Delivery         func(issue int) (effect.Stage, error)
	ActiveDispatches func(issue int) ([]string, error)
	PR               func(ctx context.Context, issue int) (number int, merged bool, err error)
}

// Probe implements reconcile.Prober.
func (s Sources) Probe(ctx context.Context, rec reconcile.Recorded) (reconcile.Actual, error) {
	h, err := s.Holder(rec.Issue)
	if err != nil {
		return reconcile.Actual{}, err
	}
	var a reconcile.Actual
	a.Liveness = s.Classify(h.PID, h.ProcessStart)
	if a.WorktreeHead, err = s.WorktreeHead(h.Worktree); err != nil {
		return reconcile.Actual{}, err
	}
	if a.RemoteHead, err = s.RemoteHead(ctx, rec.Issue); err != nil {
		return reconcile.Actual{}, err
	}
	if a.Delivery, err = s.Delivery(rec.Issue); err != nil {
		return reconcile.Actual{}, err
	}
	if a.ActiveDispatchIDs, err = s.ActiveDispatches(rec.Issue); err != nil {
		return reconcile.Actual{}, err
	}
	if a.PRNumber, a.PRMerged, err = s.PR(ctx, rec.Issue); err != nil {
		return reconcile.Actual{}, err
	}
	return a, nil
}

// ClassifierFromHomestate wraps the existing owner classifier (pid plus process-start fingerprint).
// It calls homestate.ClassifyOwnerWith and does not reimplement liveness.
func ClassifierFromHomestate(probe func(pid int) (string, homestate.ProcessIdentityState)) func(int, string) reconcile.Liveness {
	return func(pid int, start string) reconcile.Liveness {
		switch homestate.ClassifyOwnerWith(probe, pid, start) {
		case homestate.OwnerLive:
			return reconcile.Live
		case homestate.OwnerDead:
			return reconcile.Dead
		default:
			return reconcile.Indeterminate
		}
	}
}
