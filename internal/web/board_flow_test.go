package web

// board_flow_test.go — the board's flow panel (Issue #1948, local-only redesign,
// contract §4 lane-21 row). The panel shows monitor/flow's measurement of the same
// ledger the board reads: work in progress, the age of each stage, what waits and
// for how long, and the holders that have gone quiet. It measures nothing itself.
//
// Defect answer: these tests catch a flow panel that stays "not provided" over a
// readable ledger, a WIP that counts parked or waiting cards, a stalled holder
// reported alive or one that is alive reported stalled, a stage age shown for a
// card whose update time was never recorded, a group without its oldest age, and a
// flow figure presented when no ledger is wired.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/flow"
)

func flowFixture() *fakeBoardReader {
	f := boardFixture()
	f.cards = f.cards[:6] // c1..c6: the unknown-stage card is not a flow case
	for i := range f.cards {
		f.cards[i].UpdatedAt = boardNow.Add(-time.Duration(i+1) * 30 * time.Minute)
	}
	return f
}

func groupBy(gs []BoardGroupVM) map[string]BoardGroupVM {
	m := map[string]BoardGroupVM{}
	for _, g := range gs {
		m[g.Key] = g
	}
	return m
}

func TestBoardFlow_MeasuredFromTheSameLedger(t *testing.T) {
	vm := boardApp(t, flowFixture()).buildBoard(context.Background(), boardNow)
	if !vm.Flow.Known {
		t.Fatalf("a readable ledger gives a measured flow, got %+v", vm.Flow)
	}
	if vm.Flow.WIP != "3" {
		t.Errorf("WIP = %q, want 3 (c2, c3 building and c5 merging; the parked and the ready card are not in progress)", vm.Flow.WIP)
	}
	st := groupBy(vm.Flow.Stages)
	if g := st["BUILDING/SPEC"]; g.Count != "1" || g.Oldest != "1h00m" {
		t.Errorf("BUILDING/SPEC = %+v, want 1 card, oldest 1h00m", g)
	}
	if g := st["MERGING"]; g.Count != "1" || g.Oldest != "2h30m" {
		t.Errorf("MERGING = %+v", g)
	}
	var order []string
	for _, g := range vm.Flow.Stages {
		order = append(order, g.Key)
	}
	if want := "READY,BUILDING/FIX,BUILDING/SPEC,REVIEWING,MERGING"; strings.Join(order, ",") != want {
		t.Errorf("stage groups = %v, want the flow order %s", order, want)
	}
	w := groupBy(vm.Flow.Waiting)
	if w["ready"].Count != "1" || w["ready"].Oldest != "30m" || w["parked"].Count != "1" || w["parked"].Oldest != "2h00m" {
		t.Errorf("waiting = %+v", vm.Flow.Waiting)
	}
}

func TestBoardFlow_StalledHolderIsJudgedByTheProbe(t *testing.T) {
	a := boardApp(t, flowFixture())
	vm := a.buildBoard(context.Background(), boardNow)
	if len(vm.Flow.Stalled) != 1 || vm.Flow.Stalled[0].CardID != "c3" || vm.Flow.Stalled[0].HeartbeatAge != "20m" || vm.Flow.Stalled[0].Liveness != "unknown" {
		t.Errorf("without a probe a quiet holder is stalled and unknown, never dead: %+v", vm.Flow.Stalled)
	}
	a.flowProbe = func(int, string) ledger.Liveness { return ledger.LivenessAlive }
	if vm := a.buildBoard(context.Background(), boardNow); len(vm.Flow.Stalled) != 0 {
		t.Errorf("a holder whose process is alive is not stalled: %+v", vm.Flow.Stalled)
	}
	a.flowProbe = func(int, string) ledger.Liveness { return ledger.LivenessDead }
	if vm := a.buildBoard(context.Background(), boardNow); len(vm.Flow.Stalled) != 1 || vm.Flow.Stalled[0].Liveness != "dead" {
		t.Errorf("a dead holder says dead: %+v", vm.Flow.Stalled)
	}
}

func TestBoardFlow_UnrecordedUpdateTimeIsAGapNotAnAge(t *testing.T) {
	f := boardFixture() // cards carry no UpdatedAt
	vm := boardApp(t, f).buildBoard(context.Background(), boardNow)
	if len(vm.Flow.Stages) == 0 {
		t.Fatalf("the sweep below would be empty: no stage groups were measured (%+v)", vm.Flow)
	}
	for _, g := range vm.Flow.Stages {
		if g.Oldest != boardGap {
			t.Errorf("%s oldest = %q, want the gap for a card with no recorded update time", g.Key, g.Oldest)
		}
	}
}

func TestBoardFlow_PanelShowsTheMeasurementOrSaysItIsNotProvided(t *testing.T) {
	a := boardApp(t, flowFixture())
	a.nowFunc = func() time.Time { return boardNow }
	body := serveGet(t, a.routes(), "/board").Body.String()
	i := strings.Index(body, `id="board-flow"`)
	if i < 0 {
		t.Fatalf("no flow panel")
	}
	panel := body[i:]
	if end := strings.Index(panel, `class="pop"`); end > 0 {
		panel = panel[:end]
	}
	for _, want := range []string{"BUILDING/SPEC", "1h00m", "c3", "unknown", "parked", "WIP"} {
		if !strings.Contains(panel, want) {
			t.Errorf("flow panel missing %q", want)
		}
	}
	if strings.Contains(panel, "not provided") {
		t.Errorf("a measured flow must not say it is not provided")
	}
	none := boardApp(t, nil)
	noBody := serveGet(t, none.routes(), "/board").Body.String()
	if strings.Contains(noBody, `id="board-flow"`) && strings.Contains(noBody, "WIP") {
		t.Errorf("with no ledger wired no flow figure may appear")
	}
}

func TestBoardFlow_AgeText(t *testing.T) {
	for _, c := range []struct {
		d    time.Duration
		want string
	}{
		{-time.Minute, boardGap}, // a time in the future is not an age
		{30 * time.Second, "<1m"},
		{20 * time.Minute, "20m"},
		{3*time.Hour + 5*time.Minute, "3h05m"},
		{51 * time.Hour, "2d03h"},
		{200 * 365 * 24 * time.Hour, boardGap}, // never recorded
	} {
		if got := boardAge(c.d); got != c.want {
			t.Errorf("boardAge(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}

// A stage the flow order does not know is kept, after the known ones and in name
// order, never dropped or folded into a known stage.
func TestBoardFlow_UnknownStageSortsAfterTheKnownOnes(t *testing.T) {
	g := flow.Group{Count: 1, Oldest: time.Hour}
	got := boardFlow(flow.Snapshot{Stages: map[string]flow.Group{
		"ZETA": g, "MERGING": g, "QUARANTINE": g, "READY": g, "BUILDING/FIX": g,
	}})
	var order []string
	for _, s := range got.Stages {
		order = append(order, s.Key)
	}
	if want := "READY,BUILDING/FIX,MERGING,QUARANTINE,ZETA"; strings.Join(order, ",") != want {
		t.Errorf("stage groups = %v, want %s", order, want)
	}
}
