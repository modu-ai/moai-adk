// Package flow is the read model of how cards move through the local ledger:
// work in progress, the age of each stage, how long cards wait in a state, and
// which holders have gone quiet. It only reads.
package flow

import (
	"context"
	"sort"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// DefaultHeartbeatStale is how long a heartbeat may be silent before it counts
// as past (twice the 5-minute renew period of the ledger contract).
const DefaultHeartbeatStale = 10 * time.Minute

// Thresholds are the limits the stalled judgment uses.
type Thresholds struct{ HeartbeatStale time.Duration }

// Defaults returns the contract defaults.
func Defaults() Thresholds { return Thresholds{HeartbeatStale: DefaultHeartbeatStale} }

// Probe classifies the process a lease recorded (pid plus start fingerprint).
type Probe func(pid int, start string) ledger.Liveness

// ProcessProbe adapts a platform process probe to Probe through the same
// classification the claim presence uses, so pid reuse reads as Dead and every
// absence of knowledge reads as Unknown.
func ProcessProbe(probe func(int) (string, homestate.ProcessIdentityState)) Probe {
	return func(pid int, start string) ledger.Liveness {
		switch homestate.ClassifyOwnerWith(probe, pid, start) {
		case homestate.OwnerLive:
			return ledger.LivenessAlive
		case homestate.OwnerDead:
			return ledger.LivenessDead
		}
		return ledger.LivenessUnknown
	}
}

// Group is a count of cards and the longest they have been in it.
type Group struct {
	Count  int
	Oldest time.Duration
}

func (g *Group) add(age time.Duration) {
	g.Count++
	if age > g.Oldest {
		g.Oldest = age
	}
}

// Stalled is one lease whose holder shows no sign of life.
type Stalled struct {
	CardID       string
	HeartbeatAge time.Duration
	Liveness     ledger.Liveness
}

// Snapshot is one measurement.
type Snapshot struct {
	At time.Time
	// WIP counts unparked cards in BUILDING, REVIEWING or MERGING.
	WIP int
	// Stages groups non-terminal cards by "STAGE" or "STAGE/SUB".
	Stages map[string]Group
	// Waiting groups cards by what they wait in: "ready" and "parked".
	Waiting map[string]Group
	Stalled []Stalled
}

// inProgress reports whether a stage counts as work in progress: BUILDING, REVIEWING and MERGING.
// A stage this model does not know is listed under its own key but is not counted.
func inProgress(s core.Stage) bool {
	return s == core.StageBuilding || s == core.StageReviewing || s == core.StageMerging
}

func stageKey(c ledger.Card) string {
	p := c.Position
	if p.Sub != "" {
		return string(p.Stage) + "/" + string(p.Sub)
	}
	return string(p.Stage)
}

// Measure reads the ledger once and computes the snapshot at now. A nil probe
// reads every holder as Unknown.
func Measure(ctx context.Context, src ledger.Reader, probe Probe, now time.Time, th Thresholds) (Snapshot, error) {
	cards, err := src.Cards(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	leases, err := src.Leases(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	s := Snapshot{At: now, Stages: map[string]Group{}, Waiting: map[string]Group{}}
	for _, c := range cards {
		if c.Position.Stage == core.StageDone || c.Position.Stage == core.StageCancelled {
			continue
		}
		age := now.Sub(c.UpdatedAt)
		g := s.Stages[stageKey(c)]
		g.add(age)
		s.Stages[stageKey(c)] = g
		switch {
		case c.Position.Park != nil:
			w := s.Waiting["parked"]
			w.add(age)
			s.Waiting["parked"] = w
		case c.Position.Stage == core.StageReady:
			w := s.Waiting["ready"]
			w.add(age)
			s.Waiting["ready"] = w
		case inProgress(c.Position.Stage):
			s.WIP++
		}
	}
	for _, l := range leases {
		hb := now.Sub(l.HeartbeatAt)
		if hb <= th.HeartbeatStale {
			continue
		}
		if l.NextWakeAt != nil && now.Before(*l.NextWakeAt) {
			continue
		}
		live := ledger.LivenessUnknown
		if probe != nil {
			live = probe(l.PID, l.PIDStart)
		}
		if live == ledger.LivenessAlive {
			continue
		}
		s.Stalled = append(s.Stalled, Stalled{CardID: l.CardID, HeartbeatAge: hb, Liveness: live})
	}
	sort.Slice(s.Stalled, func(i, j int) bool { return s.Stalled[i].CardID < s.Stalled[j].CardID })
	return s, nil
}
