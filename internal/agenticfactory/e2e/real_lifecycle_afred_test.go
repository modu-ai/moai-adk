//go:build afred

package e2e

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// RED, not counted as a pass (Issue #1948). A retry past the limit is refused with
// ErrRetryLimit, and runner.RetryWait's own comment says "nothing changes" for an
// exhausted wait. It does change one thing: the closure guards that case with
// `&& !w.Escalated`, so a wait that is ALREADY escalated falls through and the
// record is written again with the same content and Seq+1.
//
// Observed: Seq 10 -> 11 on a refused RetryWait. Harmless to the notify service
// (an escalated wait is handled before its progress logic), but the revision number
// is the signal other readers take as "someone wrote the record", and a refusal
// should write nothing. Owner: lane-18 (internal/agenticfactory/runner/runner.go,
// RetryWait). Fix: return the error before writing when the wait is already
// escalated.
//
// Run: go test -tags afred -run TestAFRED_RetryWaitPastTheLimitWritesNothing ./internal/agenticfactory/e2e/
func TestAFRED_RetryWaitPastTheLimitWritesNothing(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	if _, err := r.HandOff(tk, goodHandoff(t)); err != nil {
		t.Fatal(err)
	}
	w := waitOf(tk)
	w.MaxRetries = 1
	if _, err := r.SetWait(tk, w); err != nil {
		t.Fatal(err)
	}
	// The first retry reaches the limit: recorded as escalated, reported as ErrRetryLimit.
	rec, err := r.RetryWait(tk, "no reply", "tell the escalation target", t0.Add(2*time.Hour))
	if !errors.Is(err, core.ErrRetryLimit) || rec.Wait == nil || !rec.Wait.Escalated {
		t.Fatalf("setup: the retry that reaches the limit = %+v, %v", rec.Wait, err)
	}
	// Positive control: the record was written by the call above (Seq moves on a real write).
	if rec.Seq == 0 {
		t.Fatal("control: the record has no revision number")
	}
	before := state(t, r, tk)
	if _, err := r.RetryWait(tk, "again", "again", t0.Add(3*time.Hour)); !errors.Is(err, core.ErrRetryLimit) {
		t.Fatalf("a retry past the limit: %v, want ErrRetryLimit", err)
	}
	after := state(t, r, tk)
	if after.Seq != before.Seq {
		t.Errorf("RED: a refused retry wrote the record again: Seq %d -> %d", before.Seq, after.Seq)
	}
	if !reflect.DeepEqual(after.Wait, before.Wait) {
		t.Errorf("a refused retry changed the wait: %+v -> %+v", before.Wait, after.Wait)
	}
}
