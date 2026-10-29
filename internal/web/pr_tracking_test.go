package web

// pr_tracking_test.go — PR tracking and merge backlog on the monitor
// (Issue #1948 DELTA4). The data is lane-20's monitor JSON, read only.
//
// Defect answer: these tests catch a PR row inventing review or repair status
// the monitor never provided, a failing or stale check read as passed, an
// unconfirmed field rendered as a value, the backlog counting rows outside
// ready/landing, and a screen that writes or runs anything.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const prRowsJSON = `{"rows":[
 {"Project":"p","Issue":11,"State":{"value":"verifying","confirmed":true},"Stage":{"value":"checks","confirmed":true},
  "HeadSHA":{"value":"0123456789abcdef","confirmed":true},
  "Checks":{"value":[{"Name":"lint","Conclusion":"success","HeadSHA":"0123456789abcdef"},{"Name":"test","Conclusion":"failure","HeadSHA":"0123456789abcdef"}],"confirmed":true},
  "agent":{"value":"agent-3","confirmed":true},"session_id":{"value":"11111111-2222-3333-4444-555555555555","confirmed":true},
  "ChecksStale":false,"pr":{"value":{"number":101,"phase":"checks-failed"},"confirmed":true},"completed":false},
 {"Project":"p","Issue":12,"State":{"value":"verifying","confirmed":true},"agent":{"value":"agent-3","confirmed":true},"session_id":{"value":"","confirmed":false},
  "HeadSHA":{"value":"aaaaaaaaaaaa","confirmed":true},
  "Checks":{"value":[{"Name":"lint","Conclusion":"success","HeadSHA":"bbbbbbbbbbbb"}],"confirmed":true},"ChecksStale":true,
  "pr":{"value":{"number":102,"phase":"repair-needed"},"confirmed":true}},
 {"Project":"p","Issue":13,"State":{"value":"done","confirmed":true},"Checks":{"value":null,"confirmed":false},
  "pr":{"value":{"number":103,"phase":"issue-closed"},"confirmed":true},"completed":true},
 {"Project":"p","Issue":14,"State":{"value":"implementing","confirmed":true},"pr":{"value":{"number":0,"phase":""},"confirmed":false}},
 {"Project":"p","Issue":15,"State":{"value":"","confirmed":false},
  "pr":{"value":{"number":104,"phase":"brand-new-phase"},"confirmed":true}}
],"backlog":{"value":{"open":6,"limit":5,"state":"over-limit"},"confirmed":true}}`

func prApp(t *testing.T, data string, err error) *app {
	t.Helper()
	a := newApp(Config{ProjectRoot: t.TempDir()})
	a.monitorRows = func(context.Context) ([]byte, error) { return []byte(data), err }
	return a
}

func TestPRTracking_RowsAndBacklog(t *testing.T) {
	vm := prApp(t, prRowsJSON, nil).buildPRTracking(context.Background())
	if !vm.Available || len(vm.Rows) != 5 {
		t.Fatalf("vm = %+v", vm)
	}
	by := map[int]PRTrackRowVM{}
	for _, r := range vm.Rows {
		by[r.Issue] = r
	}
	if r := by[11]; r.Checks != "1 passed, 1 failed" || r.PR != "#101" || r.Track != "checks" || r.Completed || r.Head != "01234567" {
		t.Errorf("row 11 = %+v", r)
	}
	if by[11].Agent != "agent-3" || by[11].SessionID != "11111111-2222-3333-4444-555555555555" {
		t.Errorf("agent name and session id must be separate fields: %+v", by[11])
	}
	if by[12].Agent != "agent-3" || by[12].SessionID != "—" {
		t.Errorf("same display name must not borrow another row's session id: %+v", by[12])
	}
	if r := by[12]; !r.ChecksStale || r.Checks != "1 passed" || r.Track != "repair" {
		t.Errorf("row 12 = %+v", r)
	}
	if r := by[13]; r.Checks != "not read" || !r.Completed || r.Track != "closed" {
		t.Errorf("row 13 = %+v", r)
	}
	if r := by[14]; r.PR != "—" || r.Phase != "—" || r.Track != "gap" {
		t.Errorf("unconfirmed PR must render as a gap: %+v", r)
	}
	if r := by[15]; r.State != "—" || r.Track != "gap" || r.Phase != "brand-new-phase" {
		t.Errorf("unknown phase must stay verbatim and ungrouped: %+v", r)
	}
	if b := vm.Backlog; b.RepairBacklog != 1 || b.HandoffWaiting != "not provided" || b.OldestWait != "not provided" {
		t.Errorf("repair backlog counts repair phases only and unprovided PRQueue fields stay gaps: %+v", b)
	}
	if got := vm.Backlog; !got.Known || got.Open != 6 || got.Limit != 5 || got.State != "over-limit" {
		t.Errorf("backlog = %+v", got)
	}
}

func TestPRTracking_BacklogUnconfirmed(t *testing.T) {
	vm := prApp(t, `{"rows":[],"backlog":{"value":{"open":0,"limit":0,"state":""},"confirmed":false}}`, nil).buildPRTracking(context.Background())
	if !vm.Available || vm.Backlog.Known {
		t.Fatalf("unconfirmed backlog must not read as a count: %+v", vm.Backlog)
	}
}

func TestPRTracking_Unavailable(t *testing.T) {
	for name, a := range map[string]*app{
		"read error": prApp(t, "", errors.New("down")),
		"bad json":   prApp(t, "{", nil),
		"unwired":    newApp(Config{ProjectRoot: t.TempDir()}),
	} {
		vm := a.buildPRTracking(context.Background())
		if vm.Available || vm.Note == "" || len(vm.Rows) != 0 {
			t.Errorf("%s: vm = %+v", name, vm)
		}
	}
}

func TestPRTracking_MonitorAndAPI(t *testing.T) {
	a := prApp(t, prRowsJSON, nil)
	m, _ := a.buildMonitor(time.Now())
	html := renderTempl(t, Monitor(tg1ShellVM(), m))
	for _, want := range []string{`id="pr-tracking"`, "#101", "agent agent-3", "session 11111111-2222-3333-4444-555555555555", "1 passed, 1 failed", "checks-failed", "open PRs 6/5 · over-limit · repair 1", "handoff waiting not provided", "oldest wait not provided"} {
		if !strings.Contains(html, want) {
			t.Errorf("monitor missing %q", want)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/api/pr-tracking", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	var vm PRTrackingVM
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &vm) != nil || vm.Backlog.Open != 6 {
		t.Fatalf("api = %d %s", rec.Code, rec.Body.String())
	}
	post := httptest.NewRequest(http.MethodPost, "/api/pr-tracking", nil)
	post.Host = "127.0.0.1:8080"
	post.Header.Set("Sec-Fetch-Site", "same-origin")
	rec = httptest.NewRecorder()
	a.routes().ServeHTTP(rec, post)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST = %d, want 405", rec.Code)
	}
}
