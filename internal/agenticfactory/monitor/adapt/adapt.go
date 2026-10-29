// Package adapt connects lane 18's claim store to the monitor and reconcile views. It only reads:
// the claim store is touched through Load, and the existing run-liveness and card-status helpers
// are called by whoever wires a Prober, never reimplemented here.
package adapt

import (
	"context"
	"errors"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/reconcile"
)

// ClaimSource implements monitor.ClaimReader over the local claim store.
type ClaimSource struct {
	Store   *claim.Store
	Project string
}

func (c ClaimSource) ReadClaim(_ context.Context, issue int) (monitor.ClaimRow, error) {
	ref, err := core.NewIssueRef(issue)
	if err != nil {
		return monitor.ClaimRow{}, err
	}
	r, err := c.Store.Load(ref)
	if errors.Is(err, core.ErrNotHeld) {
		return monitor.ClaimRow{}, monitor.ErrNotFound
	}
	if err != nil {
		return monitor.ClaimRow{}, err
	}
	return monitor.ClaimRow{
		Project:         c.Project,
		Issue:           r.Issue,
		Owner:           r.Holder,
		State:           string(r.State),
		Stage:           string(r.State),
		Attempt:         attemptNo(r.Attempt),
		Epoch:           int(r.Epoch),
		HeadSHA:         string(r.Head),
		StoreSeq:        int(r.Seq),
		BlockReason:     string(r.Reason),
		ResumeCondition: r.Resume,
	}, nil
}

// attemptNo is 1 when an attempt id is recorded and 0 otherwise; the store keeps an id, not a counter.
func attemptNo(a core.AttemptID) int {
	if a == "" {
		return 0
	}
	return 1
}

// RecordedFrom builds the reconcile input from a claim record.
func RecordedFrom(r claim.Record, leaseExpired bool, pendingDispatch string) reconcile.Recorded {
	return reconcile.Recorded{
		Issue:           r.Issue,
		HeadSHA:         string(r.Head),
		LeaseExpired:    leaseExpired,
		Delivery:        effect.Stage(r.Delivery),
		PendingDispatch: pendingDispatch,
	}
}

// LeaseExpired reports expiry from the last renewal time. An unknown renewal time never claims expiry.
func LeaseExpired(lastRenewed time.Time, duration time.Duration, now time.Time) bool {
	if lastRenewed.IsZero() || duration <= 0 {
		return false
	}
	return now.Sub(lastRenewed) > duration
}
