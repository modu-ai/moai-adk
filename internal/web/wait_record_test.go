package web

// wait_record_test.go — the agent's wait record on the PR tracking row
// (Issue #1948): the monitor Row's `wait` key, value keys named as core.Wait.
//
// Defect answer: these tests catch a value key renamed on the monitor side, a
// missing or unconfirmed `wait` shown as "none" (or as a value), a confirmed
// null shown as a gap, a retry figure invented when no limit was reported, and
// a wait record that moves the stage, state, park or gate shown.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

// waitRowsJSON is a bare fixture: key names copied from monitor.go (c0f9ce77af)
// WaitInfo's JSON tags.
const waitRowsJSON = `{"rows":[
 {"Issue":51,"State":{"value":"verifying","confirmed":true},"Stage":{"value":"REVIEWING","confirmed":true},
  "park":{"value":{"reason":"backpressure","evidence":"6/5","resume_condition":"below limit"},"confirmed":true},
  "wait":{"value":{"owner":"agent-3","next_action":"re-run checks","recheck_at":"2026-10-30T01:02:00Z","diagnosis":"ci queue full","retries":2,"max_retries":3,"escalate_to":"leader","escalated":false},"confirmed":true}},
 {"Issue":52,"wait":{"value":{"owner":"agent-4","next_action":"ask host","recheck_at":"2026-10-30T02:00:00Z","diagnosis":"no answer","retries":3,"max_retries":3,"escalate_to":"operator","escalated":true},"confirmed":true}},
 {"Issue":53,"wait":{"value":null,"confirmed":true}},
 {"Issue":54,"wait":{"value":null,"confirmed":false}},
 {"Issue":55},
 {"Issue":56,"wait":{"value":{"owner":"agent-5"},"confirmed":true}}
]}`

func waitRows(t *testing.T, data string) map[int]PRTrackRowVM {
	t.Helper()
	vm := prApp(t, data, nil).buildPRTracking(context.Background())
	by := map[int]PRTrackRowVM{}
	for _, r := range vm.Rows {
		by[r.Issue] = r
	}
	return by
}

func TestWaitRecord_Decode(t *testing.T) {
	by := waitRows(t, waitRowsJSON)
	w := by[51].Wait
	if w.State != "recorded" || w.Owner != "agent-3" || w.NextAction != "re-run checks" || w.RecheckAt != "2026-10-30 01:02Z" ||
		w.Diagnosis != "ci queue full" || w.Retries != "2 of 3" || w.EscalateTo != "leader" || w.Escalated != "no" {
		t.Errorf("row 51 wait = %+v", w)
	}
	if by[52].Wait.Escalated != "yes" || by[52].Wait.Retries != "3 of 3" {
		t.Errorf("row 52 wait = %+v", by[52].Wait)
	}
	if by[53].Wait.State != "none" {
		t.Errorf("confirmed null must be none: %+v", by[53].Wait)
	}
	for _, issue := range []int{54, 55} {
		if got := by[issue].Wait; got != (WaitVM{State: prGap}) {
			t.Errorf("row %d: unconfirmed or absent wait must be a gap, got %+v", issue, got)
		}
	}
	sparse := by[56].Wait
	if sparse.State != "recorded" || sparse.Owner != "agent-5" || sparse.NextAction != prGap || sparse.RecheckAt != prGap || sparse.Retries != prGap {
		t.Errorf("fields the monitor did not send must be gaps, no invented figures: %+v", sparse)
	}
	// The wait sits beside the position and moves nothing.
	if r := by[51]; r.State != "verifying" || r.Stage != "REVIEWING" || !strings.HasPrefix(r.Park, "backpressure") {
		t.Errorf("wait changed the position: %+v", r)
	}
}

func TestWaitRecord_RoundTrip(t *testing.T) {
	at := time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC)
	recheck := time.Date(2026, 10, 30, 3, 4, 0, 0, time.UTC)
	row := monitor.Row{
		Issue: 61,
		Wait: monitor.Field[*monitor.WaitInfo]{Value: &monitor.WaitInfo{
			Owner: "agent-9", NextAction: "retry merge", RecheckAt: recheck, Diagnosis: "conflict", Retries: 1, MaxRetries: 4,
			EscalateTo: "leader", Escalated: false,
		}, Source: monitor.SrcClaimStore, ReadAt: at, Confirmed: true},
	}
	raw, err := monitor.MarshalView(monitor.View{Rows: []monitor.Row{row}})
	if err != nil {
		t.Fatal(err)
	}
	vm := prApp(t, string(raw), nil).buildPRTracking(context.Background())
	if len(vm.Rows) != 1 {
		t.Fatalf("rows = %+v", vm.Rows)
	}
	w := vm.Rows[0].Wait
	want := WaitVM{State: "recorded", Owner: "agent-9", NextAction: "retry merge", RecheckAt: "2026-10-30 03:04Z",
		Diagnosis: "conflict", Retries: "1 of 4", EscalateTo: "leader", Escalated: "no"}
	if w != want {
		t.Errorf("wait = %+v, want %+v", w, want)
	}
}

func TestMonitorRendersWaitRecord(t *testing.T) {
	m, _ := prApp(t, waitRowsJSON, nil).buildMonitor(time.Now())
	html := renderTempl(t, Monitor(tg1ShellVM(), m))
	for _, want := range []string{
		"wait recorded", "wait owner agent-3 · next re-run checks · recheck 2026-10-30 01:02Z",
		"wait diagnosis ci queue full · retries 2 of 3 · escalate to leader · escalated no",
		"escalated yes", "wait none", "wait —",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("monitor missing %q", want)
		}
	}
}
