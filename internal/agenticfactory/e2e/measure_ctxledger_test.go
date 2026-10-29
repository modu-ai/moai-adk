package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ctxledger"
)

// M2: context-ledger write cost for a run of changing updates, plus the safety
// invariants that must hold for any implementation of it. The scenario is a
// function of a flusher so the same measurement can run against two builds.

type ctxFlusher interface {
	Submit(ctx context.Context, b ctxledger.Body) (effect.Intent, error)
	Flush(ctx context.Context) error
}

type ctxCounts struct{ Finds, Creates, Updates, Comments int }

// changingUpdates submits and flushes n different texts, then resubmits the last
// one twice (an unchanged update). It returns what the remote saw.
func changingUpdates(t *testing.T, mk func(r *ctxRemote) ctxFlusher, n int) (ctxCounts, ctxCounts) {
	t.Helper()
	r := &ctxRemote{}
	l := mk(r)
	ctx := context.Background()
	last := ""
	for i := range n {
		last = "entry " + strings.Repeat("x", i+1) + " head aaaa"
		if _, err := l.Submit(ctx, ctxBody(last)); err != nil {
			t.Fatal(err)
		}
		if err := l.Flush(ctx); err != nil {
			t.Fatal(err)
		}
	}
	changing := ctxCounts{r.Finds, r.Creates, r.Updates, len(r.comments)}
	for range 2 {
		if _, err := l.Submit(ctx, ctxBody(last)); err != nil {
			t.Fatal(err)
		}
		if err := l.Flush(ctx); err != nil {
			t.Fatal(err)
		}
	}
	total := ctxCounts{r.Finds, r.Creates, r.Updates, len(r.comments)}
	// Safety invariants, whatever the call counts are.
	if len(r.comments) != 1 {
		t.Errorf("comments = %d, want exactly 1", len(r.comments))
	}
	if len(r.comments) == 1 && !strings.Contains(r.comments[0].Body, last) {
		t.Errorf("the one comment is not at the newest text")
	}
	return changing, total
}

// Observation: a person edits the context comment, then the SAME text is
// submitted again. ctxledger keys an intent by the digest of the exact text and
// never resends confirmed text, so the edit may survive; this logs what happens
// (for the lane that owns the sender) rather than asserting an unspecified rule.
func TestMeasure_CtxLedger_ExternalEdit_Observed(t *testing.T) {
	r := &ctxRemote{}
	l := &ctxledger.Ledger{Remote: r, Store: effect.NewMemIntentStore(), Timeouts: effect.Timeouts{LedgerComment: time.Second}}
	ctx := context.Background()
	_, _ = l.Submit(ctx, ctxBody("authoritative text"))
	_ = l.Flush(ctx)
	authoritative := r.comments[0].Body
	r.mu.Lock()
	r.comments[0].Body = ctxledger.Marker(1948) + "\nedited by a person"
	r.mu.Unlock()
	before := r.calls()
	_, _ = l.Submit(ctx, ctxBody("authoritative text"))
	err := l.Flush(ctx)
	t.Logf("OBSERVED ctxledger external edit then same text: err=%v restored=%v extra-calls=%d", err, r.comments[0].Body == authoritative, r.calls()-before)
}

// Observation only: after the update that hit a deleted comment, does another
// Flush recover, and what do the stored intents say?
func TestMeasure_CtxLedger_DeletedRemotely_RetryObserved(t *testing.T) {
	r := &ctxRemote{}
	st := effect.NewMemIntentStore()
	l := &ctxledger.Ledger{Remote: r, Store: st, Timeouts: effect.Timeouts{LedgerComment: time.Second}}
	ctx := context.Background()
	_, _ = l.Submit(ctx, ctxBody("first text"))
	_ = l.Flush(ctx)
	r.mu.Lock()
	r.comments = nil
	r.mu.Unlock()
	_, _ = l.Submit(ctx, ctxBody("second text"))
	err1 := l.Flush(ctx)
	after1 := len(r.comments)
	err2 := l.Flush(ctx)
	all, _ := st.List()
	var states []string
	for _, in := range all {
		states = append(states, string(in.Status))
	}
	t.Logf("OBSERVED flush1 err=%v comments=%d | flush2 err=%v comments=%d | intent statuses=%v | finds=%d creates=%d updates=%d",
		err1, after1, err2, len(r.comments), states, r.Finds, r.Creates, r.Updates)
}

func TestMeasure_CtxLedger_ChangingUpdates_Current(t *testing.T) {
	changing, total := changingUpdates(t, func(r *ctxRemote) ctxFlusher {
		return &ctxledger.Ledger{Remote: r, Store: effect.NewMemIntentStore(), Timeouts: effect.Timeouts{LedgerComment: time.Second}}
	}, 5)
	t.Logf("MEASURED current: 5 changing updates finds=%d creates=%d updates=%d | + 2 unchanged resubmits finds=%d creates=%d updates=%d",
		changing.Finds, changing.Creates, changing.Updates, total.Finds, total.Creates, total.Updates)
}
