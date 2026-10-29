package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// Collector behavior around lost reports, restarts and duplicate events.
// Intervals and bodies are test inputs. The tests that follow log what the code
// does today; none asserts a rule the requirements have not fixed (cursor/ACK
// semantics are not specified in code yet), and they say so in their names.

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

// Only changes are handed on, in key order, and an unchanged poll hands on nothing.
func TestCollector_HandsOnlyChanges(t *testing.T) {
	tr := &scriptTransport{replies: []forge.Response{
		{Status: 200, Header: map[string]string{"etag": `"a"`}, Body: []byte(bodyTwoItems)},
		{Status: 200, Header: map[string]string{"etag": `"b"`}, Body: []byte(`[{"id":"1","s":"closed"},{"id":"2","s":"open"},{"id":"3","s":"open"}]`)},
		{Status: 304},
	}}
	clk := &clock{now: t0}
	c := newCollector(tr, clk, nil)
	r1 := c.Tick(context.Background(), clk.now)
	clk.now = r1.Next
	r2 := c.Tick(context.Background(), clk.now)
	clk.now = r2.Next
	r3 := c.Tick(context.Background(), clk.now)
	if got := kinds2(r1.Changes); len(got) != 2 {
		t.Errorf("first poll changes = %v, want 2 added", got)
	}
	if got := kinds2(r2.Changes); len(got) != 2 || got[0] != "1:modified" || got[1] != "3:added" {
		t.Errorf("second poll changes = %v, want [1:modified 3:added]", got)
	}
	if len(r3.Changes) != 0 {
		t.Errorf("unchanged poll handed on %v", kinds2(r3.Changes))
	}
	t.Logf("MEASURED collector: 3 ticks -> %d transport calls, changes handed on: %d / %d / %d", tr.n(), len(r1.Changes), len(r2.Changes), len(r3.Changes))
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

// OBSERVED, not asserted: the collector replaces its stored body when the poll
// returns, before anything confirms the caller received the report. If the
// report is lost, the changes are not handed on again.
func TestCollector_LostReport_Observed_NoRedeliveryWithoutAck(t *testing.T) {
	tr := &scriptTransport{replies: []forge.Response{
		{Status: 200, Header: map[string]string{"etag": `"a"`}, Body: []byte(bodyTwoItems)},
		{Status: 304},
	}}
	clk := &clock{now: t0}
	c := newCollector(tr, clk, nil)
	lost := c.Tick(context.Background(), clk.now) // the caller never sees this report
	clk.now = lost.Next
	again := c.Tick(context.Background(), clk.now)
	t.Logf("OBSERVED lost report: first tick handed on %d change(s) (dropped by the caller); the next tick hands on %d", len(lost.Changes), len(again.Changes))
}

// OBSERVED, not asserted: a restarted collector has no stored body, so the first
// poll hands every record on again as "added".
func TestCollector_Restart_Observed_ReemitsEverything(t *testing.T) {
	mk := func() (*forge.Collector, *scriptTransport) {
		tr := &scriptTransport{replies: []forge.Response{{Status: 200, Header: map[string]string{"etag": `"a"`}, Body: []byte(bodyTwoItems)}}}
		return newCollector(tr, &clock{now: t0}, nil), tr
	}
	before, _ := mk()
	first := before.Tick(context.Background(), t0)
	after, _ := mk() // process restart: a new collector over the same remote state
	second := after.Tick(context.Background(), t0)
	t.Logf("OBSERVED restart: before=%v after-restart=%v", kinds2(first.Changes), kinds2(second.Changes))
}

type fakeFeed struct{ changes []forge.Change }

func (f *fakeFeed) Healthy() bool           { return true }
func (f *fakeFeed) Drain() []forge.Change { out := f.changes; f.changes = nil; return out }

// OBSERVED, not asserted: events from a healthy push feed (duplicates, reversed
// order) are passed on exactly as drained.
func TestCollector_WebhookDuplicatesAndOrder_Observed_PassedThrough(t *testing.T) {
	feed := &fakeFeed{changes: []forge.Change{
		{Key: "2", Kind: forge.ChangeModified},
		{Key: "1", Kind: forge.ChangeAdded},
		{Key: "1", Kind: forge.ChangeAdded},
	}}
	tr := &scriptTransport{}
	clk := &clock{now: t0}
	c := newCollector(tr, clk, feed)
	r := c.Tick(context.Background(), clk.now)
	t.Logf("OBSERVED webhook feed: drained %v -> handed on %v (transport calls %d)", []string{"2:modified", "1:added", "1:added"}, kinds2(r.Changes), tr.n())
}
