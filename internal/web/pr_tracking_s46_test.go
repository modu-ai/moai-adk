package web

// pr_tracking_s46_test.go — the S46 position detail on the PR tracking rows
// (Issue #1948): sub_stage, park, gate and refusal from lane-20's monitor Row.
//
// Defect answer: these tests catch a key renamed on the monitor side (the
// park/gate resume key is `resume_condition`), an old Row without the keys
// rendered as a value, a confirmed "nothing recorded" shown as a gap or a
// missing key shown as "none", and the screen branching on a stage value.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

// s46RowsJSON is a bare fixture: key names copied from monitor.go (3416006156).
const s46RowsJSON = `{"rows":[
 {"Issue":31,"Stage":{"value":"REVIEWING","confirmed":true},
  "sub_stage":{"value":"awaiting-review","confirmed":true},
  "park":{"value":{"reason":"backpressure","evidence":"open PRs 6/5","resume_condition":"open PRs below limit"},"confirmed":true},
  "gate":{"value":{"reason":"checks-failed","detail":"lint red","resume_condition":"checks green","head":"0123456789abcdef"},"confirmed":true},
  "refusal":{"value":{"reason":"stale-head","detail":"head moved"},"confirmed":true}},
 {"Issue":32,"Stage":{"value":"BUILDING","confirmed":true},
  "sub_stage":{"value":"","confirmed":true},
  "park":{"value":null,"confirmed":true},"gate":{"value":null,"confirmed":true},"refusal":{"value":null,"confirmed":true}},
 {"Issue":33,"Stage":{"value":"READY","confirmed":true},
  "sub_stage":{"value":"x","confirmed":false},
  "park":{"value":null,"confirmed":false},"gate":{"value":null,"confirmed":false},"refusal":{"value":null,"confirmed":false}},
 {"Issue":34,"State":{"value":"implementing","confirmed":true},"Stage":{"value":"checks","confirmed":true}}
]}`

func TestPRTracking_S46Detail(t *testing.T) {
	vm := prApp(t, s46RowsJSON, nil).buildPRTracking(context.Background())
	by := map[int]PRTrackRowVM{}
	for _, r := range vm.Rows {
		by[r.Issue] = r
	}
	r := by[31]
	if r.Stage != "REVIEWING" || r.SubStage != "awaiting-review" {
		t.Errorf("stage/sub_stage verbatim: %+v", r)
	}
	if r.Park != "backpressure · open PRs 6/5 · resume: open PRs below limit" {
		t.Errorf("park = %q", r.Park)
	}
	if r.Gate != "checks-failed · lint red · resume: checks green · head 01234567" {
		t.Errorf("gate = %q", r.Gate)
	}
	if r.Refusal != "stale-head · head moved" {
		t.Errorf("refusal = %q", r.Refusal)
	}
	// Confirmed null: nothing recorded. Empty confirmed sub_stage is a gap.
	if r := by[32]; r.Park != "none" || r.Gate != "none" || r.Refusal != "none" || r.SubStage != prGap {
		t.Errorf("confirmed-null row = %+v", r)
	}
	// Unconfirmed keys never show their value or "none".
	if r := by[33]; r.SubStage != prGap || r.Park != prGap || r.Gate != prGap || r.Refusal != prGap {
		t.Errorf("unconfirmed row = %+v", r)
	}
	// An old Row without the new keys decodes and shows gaps.
	if r := by[34]; r.Stage != "checks" || r.SubStage != prGap || r.Park != prGap || r.Gate != prGap || r.Refusal != prGap {
		t.Errorf("old row without S46 keys = %+v", r)
	}
}

// TestPRTracking_S46RoundTrip marshals a real monitor.Row so a renamed key
// fails here instead of rendering as a silent gap.
func TestPRTracking_S46RoundTrip(t *testing.T) {
	at := time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)
	src := monitor.SrcClaimStore
	row := monitor.Row{
		Issue:    41,
		Stage:    monitor.Field[string]{Value: "MERGING", Source: src, ReadAt: at, Confirmed: true},
		SubStage: monitor.Field[string]{Value: "gate-wait", Source: src, ReadAt: at, Confirmed: true},
		Park: monitor.Field[*monitor.ParkInfo]{Value: &monitor.ParkInfo{Reason: "no-slot", Evidence: "0 free", Resume: "slot frees"},
			Source: src, ReadAt: at, Confirmed: true},
		Gate: monitor.Field[*monitor.GateInfo]{Value: &monitor.GateInfo{Reason: "review", Detail: "pending", Resume: "approved", Head: "abcdef0123456789"},
			Source: src, ReadAt: at, Confirmed: true},
		Refusal: monitor.Field[*monitor.RefusalInfo]{Source: src, ReadAt: at, Confirmed: true},
	}
	raw, err := monitor.MarshalView(monitor.View{Rows: []monitor.Row{row}})
	if err != nil {
		t.Fatal(err)
	}
	vm := prApp(t, string(raw), nil).buildPRTracking(context.Background())
	if len(vm.Rows) != 1 {
		t.Fatalf("rows = %+v", vm.Rows)
	}
	r := vm.Rows[0]
	for field, got := range map[string]string{
		"stage": r.Stage, "sub_stage": r.SubStage, "park": r.Park, "gate": r.Gate, "refusal": r.Refusal,
	} {
		want := map[string]string{
			"stage": "MERGING", "sub_stage": "gate-wait",
			"park":    "no-slot · 0 free · resume: slot frees",
			"gate":    "review · pending · resume: approved · head abcdef01",
			"refusal": "none",
		}[field]
		if got != want {
			t.Errorf("%s = %q, want %q", field, got, want)
		}
	}
}

func TestMonitorRendersS46Detail(t *testing.T) {
	m, _ := prApp(t, s46RowsJSON, nil).buildMonitor(time.Now())
	html := renderTempl(t, Monitor(tg1ShellVM(), m))
	for _, want := range []string{"sub-stage awaiting-review", "park backpressure", "gate checks-failed", "refusal stale-head", "refusal none", "refusal —"} {
		if !strings.Contains(html, want) {
			t.Errorf("monitor missing %q", want)
		}
	}
}
