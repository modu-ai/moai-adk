package ctxledger

import (
	"context"
	"errors"
	"testing"
	"time"
)

// The ledger's external writes are bounded by effect.Executor, which applies Timeouts.For around Send (that
// is why this package has no Timeouts.For call of its own). Its external READS - the check whether the
// comment was edited, a conditional Fetch or a marker Find - run outside the Executor, so they need the
// same bound here: a remote that never answers must not hold Flush for ever.

type hangFetch struct {
	*fetchRemote
	hang bool
}

func (h *hangFetch) Fetch(ctx context.Context, issue int, id int64, etag string) (Comment, string, bool, error) {
	if h.hang {
		<-ctx.Done()
		return Comment{}, "", false, ctx.Err()
	}
	return h.fetchRemote.Fetch(ctx, issue, id, etag)
}

type hangFind struct {
	*fakeRemote
	hang bool
}

func (h *hangFind) Find(ctx context.Context, issue int, marker string) (Comment, bool, error) {
	if h.hang {
		<-ctx.Done()
		return Comment{}, false, ctx.Err()
	}
	return h.fakeRemote.Find(ctx, issue, marker)
}

// flushWithin runs Flush on a context with no deadline and fails the test if it has not returned in time.
func flushWithin(t *testing.T, l *Ledger, within time.Duration) error {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel() // releases the goroutine if the test fails
	done := make(chan error, 1)
	go func() { done <- l.Flush(ctx) }()
	select {
	case err := <-done:
		return err
	case <-time.After(within):
		t.Fatalf("Flush did not return within %v: an external read has no timeout", within)
		return nil
	}
}

func TestACheckReadThatNeverAnswersIsBoundedByTheLedgerTimeoutOnTheConditionalPath(t *testing.T) {
	r := &hangFetch{fetchRemote: &fetchRemote{fakeRemote: newRemote()}}
	l, _ := newLedger(r, 4096)
	l.Timeouts.LedgerComment = 50 * time.Millisecond
	submitFlush(t, l, "text")
	writes := r.creates + r.updates
	r.hang = true
	_, _ = l.Submit(t.Context(), body("text")) // the same text again: the check runs
	err := flushWithin(t, l, 3*time.Second)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want the read's deadline", err)
	}
	if r.creates+r.updates != writes {
		t.Error("wrote although the remote state could not be read")
	}
	r.hang = false // the check stays marked: the next Flush looks again and finds nothing to do
	if err := flushWithin(t, l, 3*time.Second); err != nil {
		t.Errorf("after the remote answers again: %v", err)
	}
}

func TestACheckReadThatNeverAnswersIsBoundedByTheLedgerTimeoutOnTheMarkerLookupPath(t *testing.T) {
	r := &hangFind{fakeRemote: newRemote()}
	l, _ := newLedger(r, 4096)
	l.Timeouts.LedgerComment = 50 * time.Millisecond
	submitFlush(t, l, "text")
	writes := r.creates + r.updates
	r.hang = true
	_, _ = l.Submit(t.Context(), body("text"))
	if err := flushWithin(t, l, 3*time.Second); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want the read's deadline", err)
	}
	if r.creates+r.updates != writes {
		t.Error("wrote although the remote state could not be read")
	}
}
