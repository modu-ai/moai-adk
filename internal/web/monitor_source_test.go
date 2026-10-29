package web

// monitor_source_test.go — the real round trip between the monitor package and
// the web panels (Issue #1948): a monitor.Collector builds the View,
// monitor.MarshalView encodes it, and the web decodes those very bytes. No
// hand-written JSON, so a renamed key on either side fails here.
//
// Defect answer: these tests catch a web decoder that drifts from the monitor's
// real output, a source that fails open to an empty-but-fine panel, and a
// failing reader being shown as zero.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

type rtClaims struct{}

func (rtClaims) ReadClaim(_ context.Context, n int) (monitor.ClaimRow, error) {
	if n == 31 {
		return monitor.ClaimRow{}, monitor.ErrNotFound
	}
	return monitor.ClaimRow{
		Project: "p", Issue: n, Owner: "agent-3", AgentName: "agent-3", SessionID: "sess-0001-aaaa",
		State: "verifying", Stage: "verifying", HeadSHA: "0123456789abcdef0123",
	}, nil
}

type rtRemote struct{}

func (rtRemote) RequiredChecks(context.Context, string) ([]monitor.CheckRun, error) {
	return []monitor.CheckRun{{Name: "lint", Conclusion: "success", HeadSHA: "0123456789abcdef0123"}}, nil
}

type rtLedger struct{}

func (rtLedger) NextAction(context.Context, int) (string, error) { return "wait for review", nil }

type rtPRs struct{}

func (rtPRs) ReadPR(_ context.Context, n int) (monitor.PRInfo, bool, error) {
	return monitor.PRInfo{Number: 900 + n, Open: true, HeadSHA: "0123456789abcdef0123",
		Checks: []monitor.CheckRun{{Name: "lint", Conclusion: "success", HeadSHA: "0123456789abcdef0123"}}, ReviewVerdict: "FAIL"}, true, nil
}
func (rtPRs) CountOpenPRs(context.Context) (int, error) { return 4, nil }

type rtIntake struct{}

func (rtIntake) ReadIntake(context.Context) ([]monitor.IntakeItem, error) {
	return []monitor.IntakeItem{{Issue: 30, Stage: monitor.StageWTReady, State: "confirmed"}}, nil
}

type rtContext struct{}

func (rtContext) ReadContext(context.Context, int) ([]monitor.ContextItem, error) {
	return []monitor.ContextItem{{
		Section: "spec", Link: "https://example.test/spec", State: monitor.CtxCurrent,
		Stage: monitor.StageRemoteVerified, RemoteVerified: true, At: time.Unix(1_700_000_000, 0).UTC(),
	}}, nil
}

type rtQueue struct{ err error }

func (q rtQueue) ReadQueue(context.Context) (monitor.QueueStats, error) {
	return monitor.QueueStats{AwaitingHandoff: 2, RepairBacklog: 1, OldestWaitSeconds: 600}, q.err
}

func rtCollector(q monitor.QueueReader) *monitor.Collector {
	return &monitor.Collector{
		Claims: rtClaims{}, Remote: rtRemote{}, Ledger: rtLedger{}, PRs: rtPRs{},
		Intake: rtIntake{}, Context: rtContext{}, Queue: q, BacklogLimit: 5,
		Now: func() time.Time { return time.Unix(1_700_000_100, 0).UTC() },
	}
}

func TestMonitorRowsRoundTripThroughRealMarshalView(t *testing.T) {
	src := MonitorRowsFrom(rtCollector(rtQueue{}), func(context.Context) []int { return []int{30, 31} })
	a := newApp(Config{ProjectRoot: t.TempDir(), MonitorRows: src})
	vm := a.buildPRTracking(context.Background())

	if !vm.Available || len(vm.Rows) != 2 {
		t.Fatalf("vm = %+v", vm)
	}
	r := vm.Rows[0]
	if r.Issue != 30 || r.Agent != "agent-3" || r.SessionID != "sess-0001-aaaa" || r.PR != "#930" || r.Phase != "repair-needed" || r.Track != "repair" || r.Head != "01234567" {
		t.Errorf("row 30 lost data across the real JSON: %+v", r)
	}
	if !r.Context.Known || len(r.Context.Items) != 1 || !r.Context.Items[0].LinkOpen || r.Context.Items[0].Stage != "remote-verified" {
		t.Errorf("context lost across the real JSON: %+v", r.Context)
	}
	if g := vm.Rows[1]; g.Agent != "—" || g.SessionID != "—" || g.State != "—" {
		t.Errorf("an issue with no claim must render gaps, not values: %+v", g)
	}
	b := vm.Backlog
	if !b.Known || b.Open != 4 || b.Limit != 5 || !b.QueueKnown || b.AwaitingHandoff != 2 || b.RepairBacklog != 1 || b.OldestWaitSeconds != 600 {
		t.Errorf("backlog = %+v", b)
	}
	if !vm.Intake.Known || len(vm.Intake.Rows) != 1 || vm.Intake.Rows[0].Stage != "wt-ready" {
		t.Errorf("intake = %+v", vm.Intake)
	}
}

func TestMonitorRows_QueueReadFailureIsNotZero(t *testing.T) {
	src := MonitorRowsFrom(rtCollector(rtQueue{err: errors.New("down")}), func(context.Context) []int { return []int{30} })
	vm := newApp(Config{ProjectRoot: t.TempDir(), MonitorRows: src}).buildPRTracking(context.Background())
	if vm.Backlog.QueueKnown {
		t.Fatalf("a failed queue read must not read as known zeros: %+v", vm.Backlog)
	}
	if got := queueFigure(vm.Backlog, vm.Backlog.RepairBacklog); got != "not provided" {
		t.Errorf("figure = %q, want not provided", got)
	}
}

func TestMonitorRows_NoSourceOrNoCollectorIsNotProvided(t *testing.T) {
	for name, cfg := range map[string]Config{
		"unwired":       {ProjectRoot: t.TempDir()},
		"nil collector": {ProjectRoot: t.TempDir(), MonitorRows: MonitorRowsFrom(nil, nil)},
	} {
		vm := newApp(cfg).buildPRTracking(context.Background())
		if vm.Available || vm.Note == "" || len(vm.Rows) != 0 {
			t.Errorf("%s: want not provided, got %+v", name, vm)
		}
	}
}
