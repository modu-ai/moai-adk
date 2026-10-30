package flow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

var t0 = time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

type fakeSource struct {
	cards  []Card
	leases []Lease
	err    error
}

func (f fakeSource) Cards(context.Context) ([]Card, error)   { return f.cards, f.err }
func (f fakeSource) Leases(context.Context) ([]Lease, error) { return f.leases, f.err }

func probeOf(l Liveness) Probe { return func(int, string) Liveness { return l } }

func ptr(t time.Time) *time.Time { return &t }

func TestSnapshot_WIPAndStageAge(t *testing.T) {
	src := fakeSource{cards: []Card{
		{ID: "a", Stage: core.StageReady, UpdatedAt: t0.Add(-10 * time.Minute)},
		{ID: "b", Stage: core.StageBuilding, Sub: core.SubImplement, UpdatedAt: t0.Add(-30 * time.Minute)},
		{ID: "c", Stage: core.StageBuilding, Sub: core.SubImplement, UpdatedAt: t0.Add(-5 * time.Minute)},
		{ID: "d", Stage: core.StageMerging, UpdatedAt: t0.Add(-1 * time.Minute)},
		{ID: "e", Stage: core.StageDone, UpdatedAt: t0.Add(-99 * time.Hour)},
		{ID: "f", Stage: core.StageBuilding, Sub: core.SubFix, Parked: true, UpdatedAt: t0.Add(-2 * time.Hour)},
	}}
	s, err := Measure(context.Background(), src, probeOf(Alive), t0, Defaults())
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

func lease(hb time.Duration, wake *time.Time) Lease {
	return Lease{CardID: "b", PID: 42, PIDStart: "S", HeartbeatAt: t0.Add(-hb), NextWakeAt: wake}
}

func TestStalled_Rule(t *testing.T) {
	stale := Defaults().HeartbeatStale + time.Minute
	fresh := time.Minute
	cases := []struct {
		name  string
		l     Lease
		live  Liveness
		stall bool
	}{
		{"all three: stale, wake past, dead", lease(stale, ptr(t0.Add(-time.Minute))), Dead, true},
		{"all three: stale, wake past, unknown", lease(stale, ptr(t0.Add(-time.Minute))), Unknown, true},
		{"no promised wake counts as past", lease(stale, nil), Dead, true},
		{"process alive is never stalled", lease(stale, ptr(t0.Add(-time.Minute))), Alive, false},
		{"fresh heartbeat", lease(fresh, ptr(t0.Add(-time.Minute))), Dead, false},
		{"wake still ahead", lease(stale, ptr(t0.Add(time.Minute))), Dead, false},
		{"wake exactly now is past", lease(stale, ptr(t0)), Dead, true},
	}
	for _, c := range cases {
		src := fakeSource{cards: []Card{{ID: "b", Stage: core.StageBuilding, UpdatedAt: t0}}, leases: []Lease{c.l}}
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
	src := fakeSource{cards: []Card{{ID: "b", Stage: core.StageBuilding, UpdatedAt: t0}}, leases: []Lease{l}}
	s, _ := Measure(context.Background(), src, probeOf(Unknown), t0, Defaults())
	got := s.Stalled[0]
	if got.CardID != "b" || got.HeartbeatAge != 20*time.Minute || got.Liveness != Unknown {
		t.Fatalf("stalled evidence = %+v", got)
	}
}

func TestMeasure_SourceErrorIsReturned(t *testing.T) {
	boom := errors.New("boom")
	if _, err := Measure(context.Background(), fakeSource{err: boom}, probeOf(Alive), t0, Defaults()); !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
}

func TestMeasure_NilProbeReadsUnknown(t *testing.T) {
	src := fakeSource{cards: []Card{{ID: "b", Stage: core.StageBuilding, UpdatedAt: t0}}, leases: []Lease{lease(time.Hour, nil)}}
	s, _ := Measure(context.Background(), src, nil, t0, Defaults())
	if len(s.Stalled) != 1 || s.Stalled[0].Liveness != Unknown {
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
		want  Liveness
	}{
		{"same process", mk("S", homestate.ProcessIdentityLive), 7, "S", Alive},
		{"pid reused by another process", mk("OTHER", homestate.ProcessIdentityLive), 7, "S", Dead},
		{"process gone", mk("", homestate.ProcessIdentityDead), 7, "S", Dead},
		{"cannot tell", mk("", homestate.ProcessIdentityIndeterminate), 7, "S", Unknown},
		{"no fingerprint recorded", mk("S", homestate.ProcessIdentityLive), 7, "", Unknown},
		{"no pid recorded", mk("S", homestate.ProcessIdentityLive), 0, "S", Unknown},
	}
	for _, c := range cases {
		if got := ProcessProbe(c.probe)(c.pid, c.start); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
