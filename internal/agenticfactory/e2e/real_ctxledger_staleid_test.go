package e2e

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ctxledger"
)

// Issue #1948 S52 M1 follow-up. After lane-20 8663240cd5 the ledger remembers
// the last comment id; when a person deletes that comment the next update must
// not look successful with the comment missing. This was a RED test (first
// Flush returned nil with 0 comments) until lane-20 b1fed052c7; promoted once it
// passed. The remote now answers with the typed ctxledger.ErrGone.
func TestCtxLedger_DeletedRemotely_FirstFlushNotSilent(t *testing.T) {
	r := &ctxRemote{}
	l := &ctxledger.Ledger{Remote: r, Store: effect.NewMemIntentStore(), Timeouts: effect.Timeouts{LedgerComment: time.Second}}
	ctx := context.Background()
	if _, err := l.Submit(ctx, ctxBody("first text")); err != nil {
		t.Fatal(err)
	}
	if err := l.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	r.mu.Lock()
	r.comments = nil // deleted on GitHub
	r.mu.Unlock()
	if _, err := l.Submit(ctx, ctxBody("second text after the comment was deleted")); err != nil {
		t.Fatal(err)
	}
	err := l.Flush(ctx)
	t.Logf("MEASURED deleted-remotely (typed ErrGone) err=%v finds=%d creates=%d updates=%d comments=%d", err, r.Finds, r.Creates, r.Updates, len(r.comments))
	if err == nil && len(r.comments) != 1 {
		t.Errorf("Flush returned nil but the Issue has %d ledger comments (want 1, or an error)", len(r.comments))
	}
}

// refusedUpdateScenario: one comment exists, then the next text is submitted
// while `inject` makes the remote refuse. Flush runs `flushes` times; the
// result is what the remote saw.
func refusedUpdateScenario(t *testing.T, inject func(r *ctxRemote), flushes int) (*ctxRemote, error, bool) {
	t.Helper()
	r := &ctxRemote{}
	l := &ctxledger.Ledger{Remote: r, Store: effect.NewMemIntentStore(), Timeouts: effect.Timeouts{LedgerComment: time.Second}}
	ctx := context.Background()
	if _, err := l.Submit(ctx, ctxBody("first text")); err != nil {
		t.Fatal(err)
	}
	if err := l.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	r.mu.Lock()
	inject(r)
	r.mu.Unlock()
	if _, err := l.Submit(ctx, ctxBody("second text")); err != nil {
		t.Fatal(err)
	}
	var last error
	for range flushes {
		last = l.Flush(ctx)
	}
	return r, last, errors.Is(last, ctxledger.ErrGone)
}

// A 404 that does NOT mean "this comment is gone" must never be handled as a
// deletion: no second ledger comment is created, the failure is reported rather
// than looking like success, and it is never reported as ErrGone. The errors are
// the typed forge.StatusError the production runner returns, not messages.
func TestCtxLedger_Update404ThatIsNotGone_NeverRecreates(t *testing.T) {
	cases := map[string]func(r *ctxRemote){
		// The comment is still there and findable, but this actor may not edit it.
		"edit-refused-comment-findable": func(r *ctxRemote) { r.updateErr = httpStatus(404, "Not Found") },
		// The whole repository is refused: the update AND the marker lookup answer 404.
		"repository-not-found": func(r *ctxRemote) {
			r.updateErr = httpStatus(404, "Not Found")
			r.findErr = httpStatus(404, "Not Found")
		},
		"permission-403-on-lookup": func(r *ctxRemote) {
			r.updateErr = httpStatus(404, "Not Found")
			r.findErr = httpStatus(403, "Resource not accessible by integration")
		},
	}
	for name, inject := range cases {
		t.Run(name, func(t *testing.T) {
			r, flushErr, gone := refusedUpdateScenario(t, inject, 3)
			t.Logf("MEASURED %s over 3 flushes: err=%v finds=%d creates=%d updates=%d comments=%d", name, flushErr, r.Finds, r.Creates, r.Updates, len(r.comments))
			if r.Creates != 1 || len(r.comments) != 1 {
				t.Errorf("creates=%d comments=%d: a non-gone refusal led to a second ledger comment", r.Creates, len(r.comments))
			}
			if gone {
				t.Errorf("a non-gone refusal was reported as ErrGone")
			}
			// No silent failure: a Flush that reports no error must show the newest text.
			if flushErr == nil && len(r.comments) == 1 && !strings.Contains(r.comments[0].Body, "second text") {
				t.Errorf("Flush reported success while the comment still shows the old text")
			}
		})
	}
}

// A refusal of the whole repository (permission / not found on the lookup too)
// parks the intent: after the first refusal later Flushes spend no further
// calls. The cost is measured and compared with the case where only the edit is
// refused and the comment stays findable.
func TestCtxLedger_RefusedRepository_ParksAndStopsSpendingCalls(t *testing.T) {
	inject := func(r *ctxRemote) {
		r.updateErr = httpStatus(404, "Not Found")
		r.findErr = httpStatus(404, "Not Found")
	}
	one, _, _ := refusedUpdateScenario(t, inject, 1)
	three, _, _ := refusedUpdateScenario(t, inject, 3)
	callsOne, callsThree := one.calls(), three.calls()
	t.Logf("MEASURED refused repository: calls after 1 flush=%d, after 3 flushes=%d (includes the 2 calls of the first text)", callsOne, callsThree)
	if callsThree != callsOne {
		t.Errorf("calls grew from %d to %d across extra Flushes: the intent is not parked", callsOne, callsThree)
	}
}
