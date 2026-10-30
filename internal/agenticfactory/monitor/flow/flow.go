// Package flow is the read model of how cards move through the local ledger:
// work in progress, the age of each stage, how long cards wait in a state, and
// which holders have gone quiet. It only reads.
package flow

import (
	"context"
	"sort"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// DefaultHeartbeatStale is how long a heartbeat may be silent before it counts
// as past (twice the 5-minute renew period of the ledger contract).
const DefaultHeartbeatStale = 10 * time.Minute

// Thresholds are the limits the stalled judgment uses.
type Thresholds struct{ HeartbeatStale time.Duration }

// Defaults returns the contract defaults.
func Defaults() Thresholds { return Thresholds{HeartbeatStale: DefaultHeartbeatStale} }

// Card and Lease carry the ledger columns this package reads. They mirror
// ledger.Card / ledger.Lease and are filled by the Source adapter.
type Card struct {
	ID        string
	Stage     core.Stage
	Sub       core.SubStage
	Parked    bool
	UpdatedAt time.Time
}

type Lease struct {
	CardID      string
	PID         int
	PIDStart    string
	HeartbeatAt time.Time
	NextWakeAt  *time.Time
}

// Source is the slice of ledger.Reader the flow read model needs.
type Source interface {
	Cards(ctx context.Context) ([]Card, error)
	Leases(ctx context.Context) ([]Lease, error)
}

// Liveness is what a probe says about a lease holder's process.
type Liveness int

const (
	// Unknown is the zero value: no knowledge never reads as alive.
	Unknown Liveness = iota
	Alive
	Dead
)

func (l Liveness) String() string {
	switch l {
	case Alive:
		return "alive"
	case Dead:
		return "dead"
	}
	return "unknown"
}

// Probe classifies the process a lease recorded (pid plus start fingerprint).
type Probe func(pid int, start string) Liveness

// ProcessProbe adapts a platform process probe to Probe through the same
// classification the claim presence uses, so pid reuse reads as Dead and every
// absence of knowledge reads as Unknown.
func ProcessProbe(probe func(int) (string, homestate.ProcessIdentityState)) Probe {
	return func(pid int, start string) Liveness {
		switch homestate.ClassifyOwnerWith(probe, pid, start) {
		case homestate.OwnerLive:
			return Alive
		case homestate.OwnerDead:
			return Dead
		}
		return Unknown
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
	Liveness     Liveness
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

func stageKey(c Card) string {
	if c.Sub != "" {
		return string(c.Stage) + "/" + string(c.Sub)
	}
	return string(c.Stage)
}

// Measure reads the ledger once and computes the snapshot at now. A nil probe
// reads every holder as Unknown.
func Measure(ctx context.Context, src Source, probe Probe, now time.Time, th Thresholds) (Snapshot, error) {
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
		if c.Stage == core.StageDone || c.Stage == core.StageCancelled {
			continue
		}
		age := now.Sub(c.UpdatedAt)
		g := s.Stages[stageKey(c)]
		g.add(age)
		s.Stages[stageKey(c)] = g
		switch {
		case c.Parked:
			w := s.Waiting["parked"]
			w.add(age)
			s.Waiting["parked"] = w
		case c.Stage == core.StageReady:
			w := s.Waiting["ready"]
			w.add(age)
			s.Waiting["ready"] = w
		default:
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
		live := Unknown
		if probe != nil {
			live = probe(l.PID, l.PIDStart)
		}
		if live == Alive {
			continue
		}
		s.Stalled = append(s.Stalled, Stalled{CardID: l.CardID, HeartbeatAge: hb, Liveness: live})
	}
	sort.Slice(s.Stalled, func(i, j int) bool { return s.Stalled[i].CardID < s.Stalled[j].CardID })
	return s, nil
}
