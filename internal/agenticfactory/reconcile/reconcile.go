// Package reconcile compares what the claim store recorded with what is actually running
// (SPEC-AGENTIC-FACTORY-001 REQ-AF-009, REQ-AF-010, REQ-AF-011). The actual state is probed first;
// the result is a list of divergences. The package only reports: it never kills a process, never
// re-sends a command, and offers a new attempt only for a holder classified dead.
package reconcile

import (
	"context"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

// Liveness is the classification of an attempt holder (pid plus process-start fingerprint, produced
// by the existing homestate classifier through an adapter).
type Liveness string

const (
	Live          Liveness = "live"
	Dead          Liveness = "dead"
	Indeterminate Liveness = "indeterminate"
	Unavailable   Liveness = "unavailable" // the probe itself could not tell
)

// Recorded is the claim store's view of one Issue.
type Recorded struct {
	Issue           int
	HeadSHA         string
	LeaseExpired    bool
	Delivery        effect.Stage
	PendingDispatch string // a dispatch waiting to be sent for this Issue, if any
}

// Actual is what a read-only probe observed.
type Actual struct {
	Liveness          Liveness
	WorktreeHead      string
	RemoteHead        string
	Delivery          effect.Stage
	ActiveDispatchIDs []string // dispatches that already have a live attempt or are undelivered
}

// Prober observes actual execution. It must not change anything.
type Prober interface {
	Probe(ctx context.Context, rec Recorded) (Actual, error)
}

type Kind string

const (
	KindUnconfirmed       Kind = "unconfirmed"
	KindLeaseLapsed       Kind = "lease-lapsed"
	KindLeaseDead         Kind = "lease-expired-holder-dead"
	KindStaleHead         Kind = "stale-head"
	KindRecordBehind      Kind = "record-behind-actual"
	KindRecordAhead       Kind = "record-ahead-of-actual"
	KindDuplicateDispatch Kind = "duplicate-dispatch"
)

// Action is a recommendation for the control service. There is deliberately no action that kills a
// process or sends a command again.
type Action string

const (
	ActionNone       Action = "none"
	ActionMarkLapsed Action = "mark-lapsed"
	ActionNewAttempt Action = "new-attempt-eligible" // epoch+1, only when expired and the holder is dead
)

type Divergence struct {
	Issue  int
	Kind   Kind
	Action Action
	Detail string
}

type Report struct{ Divergences []Divergence }

// Reconcile probes actual execution first, then compares. If the probe fails or cannot tell, the
// only output is "unconfirmed": nothing is concluded from the record alone.
func Reconcile(ctx context.Context, p Prober, rec Recorded) Report {
	act, err := p.Probe(ctx, rec)
	if err != nil || act.Liveness == Unavailable {
		return Report{Divergences: []Divergence{div(rec.Issue, KindUnconfirmed, ActionNone, "actual execution could not be read; no conclusion drawn")}}
	}
	var out []Divergence
	add := func(k Kind, a Action, format string, args ...any) {
		out = append(out, div(rec.Issue, k, a, fmt.Sprintf(format, args...)))
	}

	if rec.LeaseExpired {
		if act.Liveness == Dead {
			add(KindLeaseDead, ActionNewAttempt, "lease expired and holder is dead; a new attempt with epoch+1 is eligible")
		} else {
			add(KindLeaseLapsed, ActionMarkLapsed, "lease expired but holder is %s; no takeover", act.Liveness)
		}
	}

	if head := firstDiff(rec.HeadSHA, act.WorktreeHead, act.RemoteHead); head != "" {
		add(KindStaleHead, ActionNone, "recorded HEAD %s differs from observed %s", rec.HeadSHA, head)
	}

	switch rr, ar := rec.Delivery.Rank(), act.Delivery.Rank(); {
	case ar > rr:
		add(KindRecordBehind, ActionNone, "record says %q, actual is %q", rec.Delivery, act.Delivery)
	case ar < rr:
		add(KindRecordAhead, ActionNone, "record says %q, actual is only %q; do not re-send before querying", rec.Delivery, act.Delivery)
	}

	if rec.PendingDispatch != "" && len(act.ActiveDispatchIDs) > 0 && !contains(act.ActiveDispatchIDs, rec.PendingDispatch) {
		add(KindDuplicateDispatch, ActionNone, "dispatch %s refused: collides with active dispatch %s", rec.PendingDispatch, act.ActiveDispatchIDs[0])
	}
	return Report{Divergences: out}
}

// firstDiff returns the first non-empty observed HEAD that differs from the recorded one.
func firstDiff(recorded string, observed ...string) string {
	if recorded == "" {
		return ""
	}
	for _, o := range observed {
		if o != "" && o != recorded {
			return o
		}
	}
	return ""
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

func div(issue int, k Kind, a Action, detail string) Divergence {
	masked, _ := monitor.Mask(detail)
	return Divergence{Issue: issue, Kind: k, Action: a, Detail: masked}
}
