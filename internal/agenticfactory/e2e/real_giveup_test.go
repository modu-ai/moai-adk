package e2e

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// MaxAttempts is enforced (lane-20 8bae802cdd): after that many consecutive
// rate-limited or failing answers for one path, Client.Get gives up with
// ErrGiveUp instead of deferring forever, and the collector parks. The counts
// and waits are test inputs.

var threeAttempts = forge.Backoff{MinSecondaryWait: time.Second, Base: time.Second, Max: 10 * time.Second, MaxAttempts: 3}

func giveUpClient(tr forge.Transport, clk *clock) *forge.Client {
	return &forge.Client{T: tr, B: &forge.Budget{}, Backoff: threeAttempts, Now: clk.Now}
}

// Transport failures defer until the attempts run out, then the client stops.
func TestGiveUp_TransportErrors_StopAtMaxAttempts(t *testing.T) {
	tr := &scriptTransport{err: errors.New("connection reset")}
	clk := &clock{now: t0}
	c := giveUpClient(tr, clk)
	for i := 1; i < threeAttempts.MaxAttempts; i++ {
		r, err := c.Get(context.Background(), "/p")
		if err != nil || !r.Deferred {
			t.Fatalf("attempt %d: %+v, %v; want a deferral", i, r, err)
		}
	}
	_, err := c.Get(context.Background(), "/p")
	if !errors.Is(err, forge.ErrGiveUp) {
		t.Fatalf("attempt %d: %v, want ErrGiveUp", threeAttempts.MaxAttempts, err)
	}
	t.Logf("MEASURED give-up: %d transport calls for MaxAttempts=%d", tr.n(), threeAttempts.MaxAttempts)
	if tr.n() != threeAttempts.MaxAttempts {
		t.Errorf("transport calls = %d, want exactly MaxAttempts (%d)", tr.n(), threeAttempts.MaxAttempts)
	}
}

// Consecutive limits (each answered once the previous window has passed) also end in ErrGiveUp.
func TestGiveUp_ConsecutiveLimits_StopAtMaxAttempts(t *testing.T) {
	tr := &scriptTransport{replies: []forge.Response{{Status: 429, Header: map[string]string{"retry-after": "5"}}}}
	clk := &clock{now: t0}
	c := giveUpClient(tr, clk)
	var last error
	for i := 0; i < threeAttempts.MaxAttempts; i++ {
		_, last = c.Get(context.Background(), "/p")
		clk.now = clk.now.Add(10 * time.Second) // past the stated window
	}
	if !errors.Is(last, forge.ErrGiveUp) {
		t.Fatalf("after %d consecutive limits: %v, want ErrGiveUp", threeAttempts.MaxAttempts, last)
	}
	if tr.n() != threeAttempts.MaxAttempts {
		t.Errorf("transport calls = %d, want %d", tr.n(), threeAttempts.MaxAttempts)
	}
}

// A success in between resets the count: only CONSECUTIVE failures give up.
func TestGiveUp_SuccessResetsTheCount(t *testing.T) {
	tr := &scriptTransport{replies: []forge.Response{
		{Status: 502}, {Status: 502}, {Status: 200, Body: []byte(`[]`)}, {Status: 502}, {Status: 502}, {Status: 200, Body: []byte(`[]`)},
	}}
	clk := &clock{now: t0}
	c := giveUpClient(tr, clk)
	for i := range 6 {
		if _, err := c.Get(context.Background(), "/p"); err != nil {
			t.Fatalf("call %d: %v; failures separated by a success must not give up", i+1, err)
		}
		clk.now = clk.now.Add(time.Minute)
	}
}

// Another path has its own count.
func TestGiveUp_CountIsPerPath(t *testing.T) {
	tr := &scriptTransport{err: errors.New("down")}
	clk := &clock{now: t0}
	c := giveUpClient(tr, clk)
	for i := 1; i < threeAttempts.MaxAttempts; i++ {
		_, _ = c.Get(context.Background(), "/a")
	}
	if _, err := c.Get(context.Background(), "/b"); err != nil {
		t.Errorf("a fresh path inherited another path's attempts: %v", err)
	}
}

// The collector parks on ErrGiveUp as it does on a denial: later ticks spend no call.
func TestGiveUp_Collector_ParksAndSpendsNoMoreCalls(t *testing.T) {
	tr := &scriptTransport{err: errors.New("down")}
	clk := &clock{now: t0}
	c := &forge.Collector{
		Client: giveUpClient(tr, clk), Path: "/issues", Interval: testInterval, Differ: forge.DiffByKey("id"),
	}
	var rep forge.Report
	for i := 0; i < threeAttempts.MaxAttempts; i++ {
		rep = c.Tick(context.Background(), clk.now)
		clk.now = rep.Next
	}
	if !rep.Parked {
		t.Fatalf("collector not parked after %d failures: %+v", threeAttempts.MaxAttempts, rep)
	}
	calls := tr.n()
	for i := 0; i < 5; i++ {
		c.Tick(context.Background(), clk.now.Add(time.Duration(i+1)*time.Hour))
	}
	if tr.n() != calls {
		t.Errorf("a parked collector spent %d more calls", tr.n()-calls)
	}
}
