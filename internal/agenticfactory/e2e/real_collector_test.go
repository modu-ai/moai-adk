package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// Collector behavior around lost reports, restarts and duplicate events.
// Intervals and bodies are test inputs. Contract (lane-20 8558f07fa8): a report
// is redelivered until the consumer calls Ack; the baseline is the last
// ACKNOWLEDGED body; after a restart the first report is flagged Baseline. A
// persistent cursor is not implemented, which these tests do not claim.

const bodyTwoItems = `[{"id":"1","s":"open"},{"id":"2","s":"open"}]`

var testInterval = forge.Adaptive{Min: time.Minute, Max: time.Hour, Safety: 10 * time.Minute}

func newCollector(tr forge.Transport, clk *clock, feed forge.WebhookFeed) *forge.Collector {
	return &forge.Collector{
		Client:   newClient(tr, clk, 0),
		Path:     "/issues",
		Interval: testInterval,
		Differ:   forge.DiffByKey("id"),
		Feed:     feed,
	}
}

func kinds2(ch []forge.Change) []string {
	var out []string
	for _, c := range ch {
		out = append(out, c.Key+":"+string(c.Kind))
	}
	return out
}

// Only changes are handed on, in key order, and an unchanged poll hands on
// nothing, when the consumer acknowledges each report.
func TestCollector_HandsOnlyChanges_WhenAcknowledged(t *testing.T) {
	tr := &scriptTransport{replies: []forge.Response{
		{Status: 200, Header: map[string]string{"etag": `"a"`}, Body: []byte(bodyTwoItems)},
		{Status: 200, Header: map[string]string{"etag": `"b"`}, Body: []byte(`[{"id":"1","s":"closed"},{"id":"2","s":"open"},{"id":"3","s":"open"}]`)},
		{Status: 304},
	}}
	clk := &clock{now: t0}
	c := newCollector(tr, clk, nil)
	r1 := c.Tick(context.Background(), clk.now)
	c.Ack(r1)
	clk.now = r1.Next
	r2 := c.Tick(context.Background(), clk.now)
	c.Ack(r2)
	clk.now = r2.Next
	r3 := c.Tick(context.Background(), clk.now)
	c.Ack(r3)
	if got := kinds2(r1.Changes); len(got) != 2 {
		t.Errorf("first poll changes = %v, want 2 added", got)
	}
	if got := kinds2(r2.Changes); len(got) != 2 || got[0] != "1:modified" || got[1] != "3:added" {
		t.Errorf("second poll changes = %v, want [1:modified 3:added]", got)
	}
	if len(r3.Changes) != 0 {
		t.Errorf("unchanged poll handed on %v", kinds2(r3.Changes))
	}
	t.Logf("MEASURED collector: 3 acknowledged ticks -> %d transport calls, changes handed on: %d / %d / %d", tr.n(), len(r1.Changes), len(r2.Changes), len(r3.Changes))
}

// Before its own next time the collector spends no call.
func TestCollector_BeforeNextTime_SpendsNoCall(t *testing.T) {
	tr := &scriptTransport{replies: []forge.Response{{Status: 200, Body: []byte(bodyTwoItems)}}}
	clk := &clock{now: t0}
	c := newCollector(tr, clk, nil)
	r := c.Tick(context.Background(), clk.now)
	for range 5 {
		c.Tick(context.Background(), clk.now.Add(time.Second))
	}
	if tr.n() != 1 || !r.Next.After(clk.now) {
		t.Errorf("calls = %d, next = %v: ticks before the next time must not call", tr.n(), r.Next)
	}
}

// A permission refusal parks the collector: later ticks spend no call.
func TestCollector_Denied_ParksAndSpendsNoMoreCalls(t *testing.T) {
	tr := &scriptTransport{replies: []forge.Response{{Status: 403, Body: []byte("Resource not accessible")}}}
	clk := &clock{now: t0}
	c := newCollector(tr, clk, nil)
	first := c.Tick(context.Background(), clk.now)
	for i := range 5 {
		c.Tick(context.Background(), clk.now.Add(time.Duration(i+1)*time.Hour))
	}
	if !first.Parked || tr.n() != 1 {
		t.Errorf("parked=%v calls=%d, want parked after exactly one call", first.Parked, tr.n())
	}
}

// A report the caller never acknowledged (lost, or the caller crashed) comes
// back: nothing is lost between Tick and delivery. Once acknowledged, it does not.
func TestCollector_LostReport_IsRedeliveredUntilAcked(t *testing.T) {
	tr := &scriptTransport{replies: []forge.Response{
		{Status: 200, Header: map[string]string{"etag": `"a"`}, Body: []byte(bodyTwoItems)},
		{Status: 304},
		{Status: 304},
	}}
	clk := &clock{now: t0}
	c := newCollector(tr, clk, nil)
	lost := c.Tick(context.Background(), clk.now) // the caller never sees this report
	clk.now = lost.Next
	again := c.Tick(context.Background(), clk.now)
	if got := kinds2(again.Changes); len(got) != len(lost.Changes) || len(got) == 0 {
		t.Fatalf("without an Ack the next tick handed on %v, want the lost changes %v again", got, kinds2(lost.Changes))
	}
	c.Ack(again)
	clk.now = again.Next
	after := c.Tick(context.Background(), clk.now)
	if len(after.Changes) != 0 {
		t.Errorf("after Ack the changes came back again: %v", kinds2(after.Changes))
	}
	t.Logf("MEASURED lost report: lost=%d, redelivered=%d, after Ack=%d", len(lost.Changes), len(again.Changes), len(after.Changes))
}

// After a restart the collector has no acknowledged baseline: its first report
// carries every record, and says so (Baseline), so a consumer can tell initial
// state from new events. A persistent cursor would remove this; it does not exist.
func TestCollector_Restart_FirstReportIsFlaggedBaseline(t *testing.T) {
	mk := func() (*forge.Collector, *clock) {
		clk := &clock{now: t0}
		tr := &scriptTransport{replies: []forge.Response{
			{Status: 200, Header: map[string]string{"etag": `"a"`}, Body: []byte(bodyTwoItems)},
			{Status: 304},
		}}
		return newCollector(tr, clk, nil), clk
	}
	c, clk := mk()
	first := c.Tick(context.Background(), clk.now)
	c.Ack(first)
	clk.now = first.Next
	second := c.Tick(context.Background(), clk.now)

	restarted, rclk := mk() // a new process: a new collector over the same remote state
	afterRestart := restarted.Tick(context.Background(), rclk.now)

	t.Logf("MEASURED baseline flags: first=%v second(after Ack)=%v after-restart=%v; after-restart changes=%v",
		first.Baseline, second.Baseline, afterRestart.Baseline, kinds2(afterRestart.Changes))
	if !afterRestart.Baseline {
		t.Error("the first report after a restart is not flagged Baseline")
	}
	if second.Baseline {
		t.Error("a report after an acknowledged one is flagged Baseline")
	}
}

type fakeFeed struct{ changes []forge.Change }

func (f *fakeFeed) Healthy() bool         { return true }
func (f *fakeFeed) Drain() []forge.Change { out := f.changes; f.changes = nil; return out }

// OBSERVED, not asserted: events from a healthy push feed (duplicates, reversed
// order) are logged as they come out; no rule about deduplicating or ordering
// push events has been fixed.
func TestCollector_WebhookDuplicatesAndOrder_Observed(t *testing.T) {
	feed := &fakeFeed{changes: []forge.Change{
		{Key: "2", Kind: forge.ChangeModified},
		{Key: "1", Kind: forge.ChangeAdded},
		{Key: "1", Kind: forge.ChangeAdded},
	}}
	tr := &scriptTransport{}
	clk := &clock{now: t0}
	c := newCollector(tr, clk, feed)
	r := c.Tick(context.Background(), clk.now)
	t.Logf("OBSERVED webhook feed: drained [2:modified 1:added 1:added] -> handed on %v (transport calls %d)", kinds2(r.Changes), tr.n())
}
