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
   {"section":"spec","link":"https://example.test/a","target_head":"abc12345","hash":"h1","owner":"agent-1","state":"current","stage":"remote-verified","remote_verified":true,"availability":"available","at":"2026-10-29T03:00:00Z"},
   {"section":"audit","link":"https://example.test/b","target_head":"old00000","hash":"h2","owner":"agent-2","state":"stale","stage":"pushed","remote_verified":false,"availability":"unverified","at":"2026-10-28T03:00:00Z"},
   {"section":"qa","link":"https://example.test/c","state":"inaccessible","stage":"pushed","remote_verified":false,"availability":"inaccessible","at":"2026-10-29T04:00:00Z"},
   {"section":"claim","link":"https://example.test/d","state":"current","stage":"remote-verified","remote_verified":false,"availability":"available","at":"2026-10-29T05:00:00Z"},
   {"section":"plain","link":"http://example.test/e","state":"current","stage":"remote-verified","remote_verified":true,"availability":"available","at":"2026-10-29T05:00:00Z"},
   {"section":"limited","link":"https://example.test/r","state":"current","stage":"restricted","remote_verified":true,"availability":"restricted","at":"2026-10-29T05:00:00Z"},
   {"section":"early","link":"https://example.test/p","state":"current","stage":"committed","remote_verified":false,"availability":"unverified","at":"2026-10-29T05:00:00Z"},
   {"section":"weird","link":"https://example.test/w","state":"current","stage":"unknown","remote_verified":true,"availability":"available","at":"2026-10-29T05:00:00Z"},
   {"section":"xss","link":"javascript:alert(1)","state":"current","stage":"remote-verified","remote_verified":true,"availability":"available","at":"2026-10-29T05:00:00Z"}
  ],"confirmed":true}},
 {"Project":"p","Issue":22,"State":{"value":"queued","confirmed":true},"context":{"value":null,"confirmed":false}}
],
"backlog":{"value":{"open":1,"limit":5,"state":"within-limit","awaiting_handoff":0,"repair_backlog":0,"oldest_wait_seconds":0,"queue_known":false},"confirmed":true},
"intake":{"value":[
 {"key":"k21","issue":21,"stage":"WT_PREPARED","state":"WT_PREPARED","agent":"sess-aaaa","slot":"slot-1","started":false},
 {"key":"k23","issue":23,"stage":"ACCEPTED","state":"BLOCKED","block":{"step":"ACCEPTED","detail":"token=****","retryable":true,"unknown":false,"resume":"wait for the rate limit"}},
 {"key":"k24","issue":24,"stage":"ACCEPTED","state":"EFFECT_UNKNOWN","block":{"step":"ACCEPTED","detail":"send reply lost","retryable":false,"unknown":true,"resume":"FindIssue decides"}},
 {"key":"k25","stage":"ACCEPTED","state":"PARKED","parked":"backpressure"},
 {"key":"k26","issue":26,"stage":"READY_ACKED","state":"READY_ACKED","started":true,"agent":"sess-bbbb"}
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

func TestIntakeFromMonitor_S51States(t *testing.T) {
	vm := prApp(t, intakeCtxJSON, nil).buildPRTracking(context.Background()).Intake
	if !vm.Known || vm.Source != "monitor" || len(vm.Rows) != 5 {
		t.Fatalf("intake = %+v", vm)
	}
	by := map[string]IntakeRequestVM{}
	for _, r := range vm.Rows {
		by[r.Key] = r
	}
	if r := by["k23"]; r.Status != "BLOCKED" || r.Resume != "wait for the rate limit" || r.Retryable != "yes" {
		t.Errorf("blocked = %+v", r)
	}
	if r := by["k24"]; r.Status != "EFFECT_UNKNOWN" || r.Retryable != "no" {
		t.Errorf("effect unknown = %+v", r)
	}
	if r := by["k25"]; r.Status != "PARKED" || r.Parked != "backpressure" || r.Issue != "—" {
		t.Errorf("parked is a wait, not a failure: %+v", r)
	}
	if r := by["k26"]; !r.Started || r.Stage != "READY_ACKED" {
		t.Errorf("started is a flag next to the stage: %+v", r)
	}
	if by["k21"].Agent != "sess-aaaa" || by["k21"].Slot != "slot-1" || by["k21"].AgentName != "—" {
		t.Errorf("monitor intake carries a session id only: %+v", by["k21"])
	}
	for _, r := range vm.Rows {
		switch r.Status {
		case "ack", "issue-created", "wt-ready", "started", "failed", "parked":
			t.Errorf("retired lowercase stage shown: %+v", r)
		}
	}
	if vm.ParkedCount != 1 || vm.BlockedCount != 2 {
		t.Errorf("counts parked=%d blocked=%d", vm.ParkedCount, vm.BlockedCount)
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
	if by["spec"].Stage != "remote-verified" || by["audit"].Stage != "pushed" || by["early"].Stage != "committed" {
		t.Errorf("evidence stage must be shown as reported: %+v", by)
	}
	if by["limited"].LinkStatus != "restricted" || by["limited"].LinkOpen {
		t.Errorf("restricted evidence must never open: %+v", by["limited"])
	}
	for _, sec := range []string{"audit", "qa", "claim", "plain", "xss", "limited", "early", "weird"} {
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
		`id="intake-requests"`, "PARKED", "BLOCKED", "EFFECT_UNKNOWN", "do not re-issue", "resume wait for the rate limit", "token=****",
		`id="context-ledger"`, `href="https://example.test/a"`, "link unverified", "link inaccessible",
		"stale", "inaccessible", "owner agent-1", "stage remote-verified", "stage pushed", "link restricted", "handoff waiting not provided",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("monitor missing %q", want)
		}
	}
	for _, bad := range []string{`href="javascript:`, `href="http://example.test/e"`, `href="https://example.test/d"`, `href="https://example.test/b"`, `href="https://example.test/r"`, `href="https://example.test/p"`, `href="https://example.test/w"`} {
		if strings.Contains(html, bad) {
			t.Errorf("monitor opened a link it must not: %s", bad)
		}
	}
}
