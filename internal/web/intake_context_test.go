package web

// intake_context_test.go — background intake stages, queue figures and the
// per-Issue context ledger (Issue #1948). Keys follow lane-20's monitor JSON.
//
// Defect answer: these tests catch an unknown queue shown as zero, a stale
// item and an inaccessible item rendered alike, a link opened without the
// monitor's own remote verification (or with a non-https scheme), an
// unconfirmed ledger or intake record shown as empty-but-fine, and a missing
// "started" stage being invented.

import (
	"context"
	"strings"
	"testing"
	"time"
)

const intakeCtxJSON = `{"rows":[
 {"Project":"p","Issue":21,"State":{"value":"implementing","confirmed":true},
  "context":{"value":[
   {"section":"spec","link":"https://example.test/a","target_head":"abc12345","hash":"h1","owner":"agent-1","state":"current","remote_verified":true,"availability":"available","at":"2026-10-29T03:00:00Z"},
   {"section":"audit","link":"https://example.test/b","target_head":"old00000","hash":"h2","owner":"agent-2","state":"stale","remote_verified":false,"availability":"unverified","at":"2026-10-28T03:00:00Z"},
   {"section":"qa","link":"https://example.test/c","state":"inaccessible","remote_verified":false,"availability":"inaccessible","at":"2026-10-29T04:00:00Z"},
   {"section":"claim","link":"https://example.test/d","state":"current","remote_verified":false,"availability":"available","at":"2026-10-29T05:00:00Z"},
   {"section":"plain","link":"http://example.test/e","state":"current","remote_verified":true,"availability":"available","at":"2026-10-29T05:00:00Z"},
   {"section":"xss","link":"javascript:alert(1)","state":"current","remote_verified":true,"availability":"available","at":"2026-10-29T05:00:00Z"}
  ],"confirmed":true}},
 {"Project":"p","Issue":22,"State":{"value":"queued","confirmed":true},"context":{"value":null,"confirmed":false}}
],
"backlog":{"value":{"open":1,"limit":5,"state":"within-limit","awaiting_handoff":0,"repair_backlog":0,"oldest_wait_seconds":0,"queue_known":false},"confirmed":true},
"intake":{"value":[
 {"issue":21,"stage":"wt-ready","state":"queued","at":"2026-10-29T02:00:00Z"},
 {"issue":23,"stage":"failed","state":"blocked","reason":"token=****","at":"2026-10-29T02:30:00Z"},
 {"issue":24,"stage":"parked","state":"on-hold","reason":"rate-limited","at":"2026-10-29T02:40:00Z"}
],"confirmed":true}}`

func TestQueueUnknownIsNotZero(t *testing.T) {
	vm := prApp(t, intakeCtxJSON, nil).buildPRTracking(context.Background())
	if vm.Backlog.QueueKnown {
		t.Fatalf("queue_known=false must stay false: %+v", vm.Backlog)
	}
	if got := queueFigure(vm.Backlog, vm.Backlog.RepairBacklog); got != "not provided" {
		t.Errorf("unknown queue figure = %q, want not provided", got)
	}
	if got := oldestWait(vm.Backlog); got != "not provided" {
		t.Errorf("unknown oldest wait = %q, want not provided", got)
	}
}

func TestIntakeStages(t *testing.T) {
	vm := prApp(t, intakeCtxJSON, nil).buildPRTracking(context.Background())
	if !vm.Intake.Known || len(vm.Intake.Rows) != 3 {
		t.Fatalf("intake = %+v", vm.Intake)
	}
	stages := map[int]string{}
	for _, r := range vm.Intake.Rows {
		stages[r.Issue] = r.Stage
	}
	if stages[21] != "wt-ready" || stages[23] != "failed" || stages[24] != "parked" {
		t.Errorf("stages = %v", stages)
	}
	for _, r := range vm.Intake.Rows {
		if r.Stage == "started" {
			t.Errorf("a started stage the monitor did not report was invented: %+v", r)
		}
	}
	none := prApp(t, `{"rows":[],"backlog":{"value":{},"confirmed":false},"intake":{"value":null,"confirmed":false}}`, nil).buildPRTracking(context.Background())
	if none.Intake.Known {
		t.Errorf("unconfirmed intake must read as not provided: %+v", none.Intake)
	}
}

func TestContextLedger_StaleVersusInaccessibleAndLinks(t *testing.T) {
	vm := prApp(t, intakeCtxJSON, nil).buildPRTracking(context.Background())
	var row PRTrackRowVM
	for _, r := range vm.Rows {
		if r.Issue == 21 {
			row = r
		}
	}
	by := map[string]ContextItemVM{}
	for _, it := range row.Context.Items {
		by[it.Section] = it
	}
	if by["audit"].State != "stale" || by["qa"].State != "inaccessible" || by["spec"].State != "current" {
		t.Errorf("states = %+v", by)
	}
	if by["audit"].State == by["qa"].State {
		t.Errorf("stale and inaccessible must stay distinct")
	}
	if !by["spec"].LinkOpen {
		t.Errorf("verified https link should open: %+v", by["spec"])
	}
	for _, sec := range []string{"audit", "qa", "claim", "plain", "xss"} {
		if by[sec].LinkOpen {
			t.Errorf("%s link must not open: %+v", sec, by[sec])
		}
	}
	if by["claim"].LinkStatus != "unverified" {
		t.Errorf("available without remote verification must read unverified: %+v", by["claim"])
	}
	if by["qa"].LinkStatus != "inaccessible" || by["qa"].TargetHead != "—" || by["qa"].Hash != "—" || by["qa"].Owner != "—" {
		t.Errorf("missing fields must be gaps: %+v", by["qa"])
	}
	if by["spec"].At != "2026-10-29 03:00Z" {
		t.Errorf("time = %q", by["spec"].At)
	}
	for _, r := range vm.Rows {
		if r.Issue == 22 && r.Context.Known {
			t.Errorf("unconfirmed ledger must be not provided: %+v", r.Context)
		}
	}
}

func TestMonitorRendersIntakeAndContext(t *testing.T) {
	a := prApp(t, intakeCtxJSON, nil)
	m, _ := a.buildMonitor(time.Now())
	html := renderTempl(t, Monitor(tg1ShellVM(), m))
	for _, want := range []string{
		`id="intake-stages"`, "stage wt-ready", "stage failed", "token=****",
		`id="context-ledger"`, `href="https://example.test/a"`, "link unverified", "link inaccessible",
		"stale", "inaccessible", "owner agent-1", "handoff waiting not provided",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("monitor missing %q", want)
		}
	}
	for _, bad := range []string{`href="javascript:`, `href="http://example.test/e"`, `href="https://example.test/d"`, `href="https://example.test/b"`} {
		if strings.Contains(html, bad) {
			t.Errorf("monitor opened a link it must not: %s", bad)
		}
	}
}
