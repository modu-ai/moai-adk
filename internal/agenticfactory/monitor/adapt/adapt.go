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
	// Session resolves the immutable session id of a holder. Nil or an error leaves it empty
	// (unconfirmed); the holder string is only a display name.
	Session func(issue int, holder string) (string, error)
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
	sid := ""
	if c.Session != nil {
		if v, serr := c.Session(issue, r.Holder); serr == nil {
			sid = v
		}
	}
	return rowFrom(r, c.Project, sid), nil
}

// rowFrom maps a record to the display row. Stage is the S46 stage (derived from the legacy
// state when the record carries none); State stays the legacy name for display only.
func rowFrom(r claim.Record, project, sid string) monitor.ClaimRow {
	row := monitor.ClaimRow{
		Project:         project,
		AgentName:       r.Holder,
		SessionID:       sid,
		Issue:           r.Issue,
		Owner:           r.Holder,
		State:           string(r.State),
		Stage:           string(StageOf(r)),
		SubStage:        string(r.Sub),
		Attempt:         attemptNo(r.Attempt),
		Epoch:           int(r.Epoch),
		HeadSHA:         string(r.Head),
		StoreSeq:        int(r.Seq),
		BlockReason:     string(r.Reason),
		ResumeCondition: r.Resume,
	}
	if p := r.Park; p != nil {
		row.Park = &monitor.ParkInfo{Reason: string(p.Reason), Evidence: p.Evidence, Resume: p.Resume}
	}
	if g := r.Gate; g != nil {
		row.Gate = &monitor.GateInfo{Reason: string(g.Reason), Detail: g.Detail, Resume: g.Resume, Head: string(g.Head)}
	}
	if f := r.Refusal; f != nil {
		row.Refusal = &monitor.RefusalInfo{Reason: string(f.Reason), Detail: f.Detail}
	}
	if w := r.Wait; w != nil {
		row.Wait = &monitor.WaitInfo{
			Owner: w.Owner, NextAction: w.NextAction, RecheckAt: w.RecheckAt, Diagnosis: w.Diagnosis,
			Retries: w.Retries, MaxRetries: w.MaxRetries, EscalateTo: w.Escalate, Escalated: w.Escalated,
		}
	}
	return row
}

// StageOf is the record's S46 stage: the stored one, else the position the legacy fields map to.
// A record that cannot be mapped has no stage; callers treat that as unconfirmed, never as a guess.
func StageOf(r claim.Record) core.Stage {
	if r.Stage != "" {
		return r.Stage
	}
	pos, err := core.FromLegacy(core.Legacy{
		State: r.State, Return: r.Return, Reason: r.Reason, Resume: r.Resume,
		HasReceipt: r.Receipt != nil, LeaderAccepted: r.Receipt != nil && r.Receipt.LeaderAccepted, Merged: r.Merged,
	})
	if err != nil {
		return ""
	}
	return pos.Stage
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
		Stage:           StageOf(r),
	}
}

// LeaseExpired reports expiry from the last renewal time. An unknown renewal time never claims expiry.
func LeaseExpired(lastRenewed time.Time, duration time.Duration, now time.Time) bool {
	if lastRenewed.IsZero() || duration <= 0 {
		return false
	}
	return now.Sub(lastRenewed) > duration
}
