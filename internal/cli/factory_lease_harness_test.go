// factory_lease_harness_test.go — shared harnesses of the SPEC-FACTORY-ATOMIC-LEASE-001
// (card t1458) acceptance tests: the tolerant seam gate, the operator-write
// probe, the claim-write counter, and the small fixture readers the criteria
// share. Every fixture lives under t.TempDir() (nmBase); nothing here touches
// the real queue or the real factory record.
package cli

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// flGrace is the tolerant gate's grace window (plan WM1): a heuristic, not a
// derivation. It must exceed the time the first lane needs to leave the seam,
// which was milliseconds on the unfixed tree.
const flGrace = 750 * time.Millisecond

// flMargin is the 500 ms margin of AC-FAL-007 (a stated heuristic).
const flMargin = 500 * time.Millisecond

// flBoardBudgetSizing is the queue lock's wait budget as the tree derived it at
// plan time (ledger L6: 3.3 s). It is a fallback used only while
// kanban.LockWaitBudget is the WM1 stub returning 0, so a bound test still
// measures the behavior; the stub value itself is reported as a failure.
const flBoardBudgetSizing = 3300 * time.Millisecond

// flBudget returns kanban.LockWaitBudget(), failing the test (not stopping it)
// when the accessor is the WM1 stub's zero, and then falling back to the plan
// sizing figure so the behavioral half of the test still runs.
func flBudget(t *testing.T) time.Duration {
	t.Helper()
	b := kanban.LockWaitBudget()
	if b <= 0 {
		t.Errorf("kanban.LockWaitBudget() = %s; the bound cannot be derived from the accessor (WM1 stub, WM2 replaces it); using the plan sizing figure %s for the behavioral half", b, flBoardBudgetSizing)
		return flBoardBudgetSizing
	}
	return b
}

// flGoroutineID returns the calling goroutine's id (test-only device: a hook
// holds each lane once, and a lane is one goroutine of the test process).
func flGoroutineID() int {
	buf := make([]byte, 64)
	buf = buf[:runtime.Stack(buf, false)]
	fields := strings.Fields(string(buf))
	if len(fields) < 2 {
		return -1
	}
	id, _ := strconv.Atoi(fields[1])
	return id
}

// flStackHas reports whether any frame of the calling stack is one of the named
// functions of this package (matched by suffix, e.g. ".factoryNextClaim").
func flStackHas(names ...string) bool {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		f, more := frames.Next()
		for _, name := range names {
			if strings.HasSuffix(f.Function, "."+name) {
				return true
			}
		}
		if !more {
			return false
		}
	}
}

// flAtClaimWrite reports whether factoryCardNow was called by a claim record
// write: RecordPicked in factoryNextRecordAndClaim or an edge in
// factoryNextClaim (the claim calls the clock once per record write, as an
// argument of the write).
func flAtClaimWrite() bool {
	return flStackHas("factoryNextRecordAndClaim", "factoryNextClaim")
}

// flGate is the tolerant seam gate (plan WM1, "nmRaceAtSeamTolerant"): each lane
// that reaches the seam waits until all n lanes have arrived OR flGrace has
// passed since its own arrival, then proceeds. On the unfixed tree every lane
// arrives within milliseconds and all are released together; on a tree whose
// seam sits inside an exclusive section the second lane cannot arrive while
// the first holds the section, so the first waits out the grace and proceeds
// alone, and the second proceeds at once when it arrives.
type flGate struct {
	n       int
	mu      sync.Mutex
	count   int
	all     chan struct{}
	arrived chan struct{}
	seen    map[string]bool
}

func newFLGate(n int) *flGate {
	return &flGate{n: n, all: make(chan struct{}), arrived: make(chan struct{}, 16), seen: map[string]bool{}}
}

// hold is the seam body. key identifies a lane (a goroutine id or a card id);
// a key already held passes straight through, so a re-selection attempt or a
// second claim write is not held twice.
func (g *flGate) hold(key string) {
	g.mu.Lock()
	if g.seen[key] {
		g.mu.Unlock()
		return
	}
	g.seen[key] = true
	g.count++
	if g.count == g.n {
		close(g.all)
	}
	g.mu.Unlock()
	g.arrived <- struct{}{}
	select {
	case <-g.all:
	case <-time.After(flGrace):
	}
}

// flLaneResult is what one raced invocation of `factory next` produced.
type flLaneResult struct {
	lane, card  string
	out, stderr string
	err         error
}

// flRaceLanes starts each lane's `factory next` (with `--card` when card is
// set) in its own goroutine, stamping the lane's environment before the launch.
// A lane label is read once at the start of an invocation, so the next label is
// stamped only after the previous lane has reached the gate or finished. It
// returns every lane's result.
func flRaceLanes(t *testing.T, g *flGate, lanes []nmLaneRun) []flLaneResult {
	t.Helper()
	done := make(chan flLaneResult, len(lanes))
	for _, l := range lanes {
		nmLaneEnv(t, l.label, "")
		go func(l nmLaneRun) {
			args := []string{"--run", fcRun}
			if l.card != "" {
				args = append(args, "--card", l.card)
			}
			out, stderr, err := qasRunNext(t, args...)
			done <- flLaneResult{lane: l.label, card: l.card, out: out, stderr: stderr, err: err}
		}(l)
		select {
		case <-g.arrived:
		case r := <-done:
			// The lane finished without reaching the gate (refused, or no card);
			// keep its result and go on.
			done <- r
			time.Sleep(10 * time.Millisecond)
		case <-time.After(30 * time.Second):
			t.Fatalf("lane %s never reached the gate nor finished", l.label)
		}
	}
	var results []flLaneResult
	for range lanes {
		select {
		case r := <-done:
			results = append(results, r)
		case <-time.After(90 * time.Second):
			t.Fatal("a raced invocation never finished")
		}
	}
	return results
}

// flLeasedRows counts the cards of ids whose record row is `leased`, and lists
// every row it found.
func flLeasedRows(t *testing.T, root string, ids ...string) (int, []string) {
	t.Helper()
	n := 0
	var rows []string
	for _, id := range ids {
		if !fcHasCard(t, root, id) {
			continue
		}
		c := fcCard(t, root, id)
		rows = append(rows, fmt.Sprintf("%s=%s/%s", id, c.State, c.LeaseHolder))
		if c.State == homestate.CardLeased {
			n++
		}
	}
	return n, rows
}

// flRow returns a card's record state and lease holder ("" when no row).
func flRow(t *testing.T, root, cardID string) (string, string) {
	t.Helper()
	if !fcHasCard(t, root, cardID) {
		return "", ""
	}
	c := fcCard(t, root, cardID)
	return c.State, c.LeaseHolder
}

// flOp is an operator queue write running in its own goroutine (the probe of
// ledger L11): a write that can block on the queue lock without deadlocking the
// lease that holds it.
type flOp struct {
	ch        chan error
	completed bool
	err       error
}

// flStartWrite starts mutate under the public Mutate of store in a goroutine.
func flStartWrite(store *kanban.BacklogStore, mutate func(*kanban.BacklogRecord) error) *flOp {
	op := &flOp{ch: make(chan error, 1)}
	go func() { op.ch <- store.Mutate(mutate) }()
	return op
}

// flStartState starts an operator write that sets cardID's queue state.
func flStartState(store *kanban.BacklogStore, cardID string, states ...kanban.BacklogState) *flOp {
	return flStartWrite(store, func(r *kanban.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == cardID {
				r.Items[i].State = states[len(states)-1]
				return nil
			}
		}
		return fmt.Errorf("no card %s", cardID)
	})
}

// flStartNoop starts an operator write that changes nothing (a lock probe).
func flStartNoop(store *kanban.BacklogStore) *flOp {
	return flStartWrite(store, func(*kanban.BacklogRecord) error { return nil })
}

// within reports whether the write has completed, waiting at most d for it.
func (o *flOp) within(d time.Duration) bool {
	if o.completed {
		return true
	}
	select {
	case o.err = <-o.ch:
		o.completed = true
	case <-time.After(d):
	}
	return o.completed
}

// flJoin waits for every started write to finish (nothing outlives the test).
func flJoin(t *testing.T, ops []*flOp) {
	t.Helper()
	for _, op := range ops {
		if op != nil && !op.within(15*time.Second) {
			t.Errorf("an operator write never completed after the verb returned")
		}
	}
}

// flAssertOneRefusalLine asserts err carries exit status code, stdout is empty
// and stderr is exactly one line beginning with prefix.
func flAssertOneRefusalLine(t *testing.T, r flLaneResult, code int, prefix string) {
	t.Helper()
	if got := nmExit(r.err); got != code {
		t.Errorf("lane %s exited %d (err=%v stderr=%q), want %d", r.lane, got, r.err, r.stderr, code)
	}
	lines := nmNonEmptyLines(r.stderr)
	if len(lines) != 1 || !strings.HasPrefix(lines[0], prefix) {
		t.Errorf("lane %s stderr = %q, want exactly one line beginning %q", r.lane, r.stderr, prefix)
	}
}
