package flow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

var t0 = time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

type fakeSource struct {
	cards  []ledger.Card
	leases []ledger.Lease
	err    error
}

func (f fakeSource) Cards(context.Context) ([]ledger.Card, error)   { return f.cards, f.err }
func (f fakeSource) Leases(context.Context) ([]ledger.Lease, error) { return f.leases, f.err }
func (f fakeSource) Queue(context.Context) ([]ledger.QueueItem, error) {
	return nil, f.err
}
func (f fakeSource) Notes(context.Context, int64) ([]ledger.Note, error) { return nil, f.err }

func probeOf(l ledger.Liveness) Probe { return func(int, string) ledger.Liveness { return l } }

// card builds a ledger card; park non-nil marks it parked.
func card(id string, stage core.Stage, sub core.SubStage, updated time.Time, park *core.Park) ledger.Card {
	return ledger.Card{ID: id, Position: core.Position{Stage: stage, Sub: sub, Park: park}, UpdatedAt: updated}
}

func ptr(t time.Time) *time.Time { return &t }

func TestSnapshot_WIPAndStageAge(t *testing.T) {
	src := fakeSource{cards: []ledger.Card{
		card("a", core.StageReady, "", t0.Add(-10*time.Minute), nil),
		card("b", core.StageBuilding, core.SubImplement, t0.Add(-30*time.Minute), nil),
		card("c", core.StageBuilding, core.SubImplement, t0.Add(-5*time.Minute), nil),
		card("d", core.StageMerging, "", t0.Add(-1*time.Minute), nil),
		card("e", core.StageDone, "", t0.Add(-99*time.Hour), nil),
		card("f", core.StageBuilding, core.SubFix, t0.Add(-2*time.Hour), &core.Park{}),
	}}
	s, err := Measure(context.Background(), src, probeOf(ledger.LivenessAlive), t0, Defaults())
	if err != nil {
		t.Fatal(err)
	}
	if s.WIP != 3 {
		t.Fatalf("WIP = %d, want 3 (b, c, d; parked and terminal excluded)", s.WIP)
	}
	g := s.Stages["BUILDING/IMPLEMENT"]
	if g.Count != 2 || g.Oldest != 30*time.Minute {
		t.Fatalf("BUILDING/IMPLEMENT = %+v", g)
	}
	if s.Stages["READY"].Oldest != 10*time.Minute {
		t.Fatalf("READY = %+v", s.Stages["READY"])
	}
	if s.Waiting["parked"].Count != 1 || s.Waiting["parked"].Oldest != 2*time.Hour {
		t.Fatalf("parked = %+v", s.Waiting["parked"])
	}
	if s.Waiting["ready"].Count != 1 {
		t.Fatalf("ready wait = %+v", s.Waiting["ready"])
	}
	if _, ok := s.Stages["DONE"]; ok {
		t.Fatal("terminal stages are not a flow stage")
	}
}

func lease(hb time.Duration, wake *time.Time) ledger.Lease {
	return ledger.Lease{CardID: "b", PID: 42, PIDStart: "S", HeartbeatAt: t0.Add(-hb), NextWakeAt: wake}
}

func TestStalled_Rule(t *testing.T) {
	stale := Defaults().HeartbeatStale + time.Minute
	fresh := time.Minute
	cases := []struct {
		name  string
		l     ledger.Lease
		live  ledger.Liveness
		stall bool
	}{
		{"all three: stale, wake past, dead", lease(stale, ptr(t0.Add(-time.Minute))), ledger.LivenessDead, true},
		{"all three: stale, wake past, unknown", lease(stale, ptr(t0.Add(-time.Minute))), ledger.LivenessUnknown, true},
		{"no promised wake counts as past", lease(stale, nil), ledger.LivenessDead, true},
		{"process alive is never stalled", lease(stale, ptr(t0.Add(-time.Minute))), ledger.LivenessAlive, false},
		{"fresh heartbeat", lease(fresh, ptr(t0.Add(-time.Minute))), ledger.LivenessDead, false},
		{"wake still ahead", lease(stale, ptr(t0.Add(time.Minute))), ledger.LivenessDead, false},
		{"wake exactly now is past", lease(stale, ptr(t0)), ledger.LivenessDead, true},
	}
	for _, c := range cases {
		src := fakeSource{cards: []ledger.Card{card("b", core.StageBuilding, "", t0, nil)}, leases: []ledger.Lease{c.l}}
		s, err := Measure(context.Background(), src, probeOf(c.live), t0, Defaults())
		if err != nil {
			t.Fatal(err)
		}
		if got := len(s.Stalled) == 1; got != c.stall {
			t.Errorf("%s: stalled=%v, want %v (%+v)", c.name, got, c.stall, s.Stalled)
		}
	}
}

func TestStalled_ReportsEvidence(t *testing.T) {
	l := lease(20*time.Minute, nil)
	src := fakeSource{cards: []ledger.Card{card("b", core.StageBuilding, "", t0, nil)}, leases: []ledger.Lease{l}}
	s, _ := Measure(context.Background(), src, probeOf(ledger.LivenessUnknown), t0, Defaults())
	got := s.Stalled[0]
	if got.CardID != "b" || got.HeartbeatAge != 20*time.Minute || got.Liveness != ledger.LivenessUnknown {
		t.Fatalf("stalled evidence = %+v", got)
	}
}

func TestMeasure_SourceErrorIsReturned(t *testing.T) {
	boom := errors.New("boom")
	if _, err := Measure(context.Background(), fakeSource{err: boom}, probeOf(ledger.LivenessAlive), t0, Defaults()); !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
}

func TestMeasure_NilProbeReadsUnknown(t *testing.T) {
	src := fakeSource{cards: []ledger.Card{card("b", core.StageBuilding, "", t0, nil)}, leases: []ledger.Lease{lease(time.Hour, nil)}}
	s, _ := Measure(context.Background(), src, nil, t0, Defaults())
	if len(s.Stalled) != 1 || s.Stalled[0].Liveness != ledger.LivenessUnknown {
		t.Fatalf("a missing probe must not read as alive: %+v", s.Stalled)
	}
}

func TestProcessProbe_Mapping(t *testing.T) {
	mk := func(start string, st homestate.ProcessIdentityState) func(int) (string, homestate.ProcessIdentityState) {
		return func(int) (string, homestate.ProcessIdentityState) { return start, st }
	}
	cases := []struct {
		name  string
		probe func(int) (string, homestate.ProcessIdentityState)
		pid   int
		start string
		want  ledger.Liveness
	}{
		{"same process", mk("S", homestate.ProcessIdentityLive), 7, "S", ledger.LivenessAlive},
		{"pid reused by another process", mk("OTHER", homestate.ProcessIdentityLive), 7, "S", ledger.LivenessDead},
		{"process gone", mk("", homestate.ProcessIdentityDead), 7, "S", ledger.LivenessDead},
		{"cannot tell", mk("", homestate.ProcessIdentityIndeterminate), 7, "S", ledger.LivenessUnknown},
		{"no fingerprint recorded", mk("S", homestate.ProcessIdentityLive), 7, "", ledger.LivenessUnknown},
		{"no pid recorded", mk("S", homestate.ProcessIdentityLive), 0, "S", ledger.LivenessUnknown},
	}
	for _, c := range cases {
		if got := ProcessProbe(c.probe)(c.pid, c.start); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

// WIP is BUILDING, REVIEWING and MERGING only: a stage the model does not know
// (a newer ledger, a hand edit) is listed under its own key but is not work in progress.
func TestSnapshot_UnknownStageIsNotWIP(t *testing.T) {
	src := fakeSource{cards: []ledger.Card{
		card("a", core.StageReviewing, "", t0, nil),
		card("q", core.Stage("QUARANTINE"), "", t0.Add(-time.Hour), nil),
	}}
	s, err := Measure(context.Background(), src, probeOf(ledger.LivenessAlive), t0, Defaults())
	if err != nil {
		t.Fatal(err)
	}
	if s.WIP != 1 {
		t.Fatalf("WIP = %d, want 1 (only the REVIEWING card)", s.WIP)
	}
	if g := s.Stages["QUARANTINE"]; g.Count != 1 {
		t.Fatalf("the unknown stage must still be visible: %+v", g)
	}
}
