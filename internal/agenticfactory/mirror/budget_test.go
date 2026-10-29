package mirror

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// fakeBudget stands in for the shared call budget that lane-20 owns.
type fakeBudget struct {
	mu    sync.Mutex
	allow map[core.Priority]bool
	asked []core.Priority
}

func (b *fakeBudget) Reserve(p core.Priority) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.asked = append(b.asked, p)
	ok, set := b.allow[p]
	return !set || ok
}

func TestReportWriteIsDeferredWhenTheBudgetIsShort(t *testing.T) {
	f := &fakeGH{results: []result{{out: ""}, {out: `{"id":5}`}}}
	b := &fakeBudget{allow: map[core.Priority]bool{core.PriorityReport: false}}
	m := writable(f)
	m.Budget = b
	err := m.WriteLedger(context.Background(), ledger())
	if !errors.Is(err, core.ErrDeferred) {
		t.Fatalf("want ErrDeferred, got %v", err)
	}
	if len(f.calls) != 0 {
		t.Errorf("a deferred report must make no gh call: %v", calls(f))
	}
	if len(b.asked) != 1 || b.asked[0] != core.PriorityReport {
		t.Errorf("budget asked %v", b.asked)
	}
	// The deferral remembered nothing: once the budget allows it, the write happens.
	b.allow[core.PriorityReport] = true
	if err := m.WriteLedger(context.Background(), ledger()); err != nil || len(f.calls) != 2 {
		t.Errorf("write after the budget recovered: %v calls=%v", err, calls(f))
	}
}

func TestUnchangedBodyCostsNoBudget(t *testing.T) {
	f := &fakeGH{results: []result{{out: ""}, {out: `{"id":5}`}}}
	b := &fakeBudget{}
	m := writable(f)
	m.Budget = b
	_ = m.WriteLedger(context.Background(), ledger())
	_ = m.WriteLedger(context.Background(), ledger())
	if len(b.asked) != 1 {
		t.Errorf("an unchanged body must not spend budget, asked %v", b.asked)
	}
}

func TestPublishMarksMirrorStaleWhenDeferredAndLeavesStateAlone(t *testing.T) {
	st, _ := claim.OpenWithLimit(t.TempDir(), 5)
	ref, _ := core.NewIssueRef(1948)
	rec, _ := st.Acquire(ref, "sess-a", "d1", "run-1")
	m := writable(&fakeGH{})
	m.Budget = &fakeBudget{allow: map[core.Priority]bool{core.PriorityReport: false}}
	out, err := Publish(context.Background(), m, st, ref, "sess-a", rec.Epoch)
	if !errors.Is(err, core.ErrDeferred) {
		t.Fatalf("want ErrDeferred, got %v", err)
	}
	if out.Mirror != "stale" || out.State != core.StateClaimed {
		t.Errorf("record after deferral: %+v", out)
	}
}

func TestEnqueueCoalescesAndFlushWritesEachIssueOnce(t *testing.T) {
	f := &fakeGH{results: []result{{out: ""}, {out: `{"id":1}`}, {out: ""}, {out: `{"id":2}`}}}
	m := writable(f)
	for _, next := range []string{"a", "b", "c", "d", "latest"} {
		m.Enqueue(ledgerWith(next))
	}
	other := ledgerWith("x")
	other.Issue = core.IssueRef{Number: 2000}
	m.Enqueue(other)
	if got := m.Pending(); len(got) != 2 || got[0] != 1948 || got[1] != 2000 {
		t.Fatalf("pending = %v", got)
	}
	n, err := m.Flush(context.Background())
	if err != nil || n != 2 {
		t.Fatalf("flush: %d %v", n, err)
	}
	if len(f.calls) != 4 { // one list + one create per issue, not one per enqueue
		t.Errorf("calls = %v", calls(f))
	}
	if len(m.Pending()) != 0 {
		t.Errorf("flushed entries must leave the queue: %v", m.Pending())
	}
	// Only the latest body was written.
	if body := f.calls[1][len(f.calls[1])-1]; !contains(body, "latest") || contains(body, `"a"`) {
		t.Errorf("coalesced body: %s", body)
	}
}

func contains(s, sub string) bool {
	return len(sub) > 0 && (len(s) >= len(sub)) && indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestFlushStopsAtAShortBudgetAndKeepsEverythingQueued(t *testing.T) {
	f := &fakeGH{}
	m := writable(f)
	m.Budget = &fakeBudget{allow: map[core.Priority]bool{core.PriorityReport: false}}
	m.Enqueue(ledgerWith("a"))
	other := ledgerWith("x")
	other.Issue = core.IssueRef{Number: 2000}
	m.Enqueue(other)
	n, err := m.Flush(context.Background())
	if !errors.Is(err, core.ErrDeferred) || n != 0 {
		t.Fatalf("flush: %d %v", n, err)
	}
	if len(f.calls) != 0 || len(m.Pending()) != 2 {
		t.Errorf("calls=%v pending=%v", calls(f), m.Pending())
	}
}

func TestFlushStopsAtABlockAndDoesNotHammer(t *testing.T) {
	f := &fakeGH{results: []result{{errOut: "HTTP 429", exit: 1}}}
	m := writable(f)
	m.Enqueue(ledgerWith("a"))
	other := ledgerWith("x")
	other.Issue = core.IssueRef{Number: 2000}
	m.Enqueue(other)
	n, err := m.Flush(context.Background())
	var be *BlockedError
	if !errors.As(err, &be) || n != 0 {
		t.Fatalf("flush: %d %v", n, err)
	}
	if len(f.calls) != 1 || len(m.Pending()) != 2 {
		t.Errorf("a block must stop the flush after one call: calls=%v pending=%v", calls(f), m.Pending())
	}
}

func TestEnqueueDuringFlushIsNotLost(t *testing.T) {
	f := &fakeGH{results: []result{{out: ""}, {out: `{"id":1}`}}}
	m := writable(f)
	m.Enqueue(ledgerWith("first"))
	snapshotTaken := make(chan struct{})
	release := make(chan struct{})
	f2 := &blockingGH{inner: f, entered: snapshotTaken, release: release}
	m.Exec = f2
	done := make(chan struct{})
	go func() { _, _ = m.Flush(context.Background()); close(done) }()
	<-snapshotTaken
	m.Enqueue(ledgerWith("newer"))
	close(release)
	<-done
	if got := m.Pending(); len(got) != 1 {
		t.Errorf("an entry queued while flushing must stay queued, pending=%v", got)
	}
}

type blockingGH struct {
	inner   Exec
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (b *blockingGH) Run(ctx context.Context, args ...string) ([]byte, []byte, int, error) {
	b.once.Do(func() { close(b.entered); <-b.release })
	return b.inner.Run(ctx, args...)
}
