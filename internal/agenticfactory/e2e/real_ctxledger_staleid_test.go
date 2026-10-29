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

// A 404 that does NOT mean "this comment is gone" (the repository or Issue is
// not found, or the comment exists but the actor may not edit it) must never be
// handled as a deletion: no second ledger comment is created, and the failure
// is reported instead of looking like success.
func TestCtxLedger_Update404ThatIsNotGone_NeverRecreates(t *testing.T) {
	for name, uerr := range map[string]error{
		"repository-not-found": &statusErr{Status: 404, Msg: "Not Found (repository)"},
		"permission-404":       &statusErr{Status: 404, Msg: "Not Found (no access to this resource)"},
	} {
		t.Run(name, func(t *testing.T) {
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
			r.updateErr = uerr // the comment still exists, but every update is refused
			r.mu.Unlock()
			if _, err := l.Submit(ctx, ctxBody("second text")); err != nil {
				t.Fatal(err)
			}
			var flushErr error
			for range 3 { // a few flushes: the failure must not turn into a re-create at any point
				flushErr = l.Flush(ctx)
			}
			t.Logf("MEASURED %s: err=%v finds=%d creates=%d updates=%d comments=%d", name, flushErr, r.Finds, r.Creates, r.Updates, len(r.comments))
			if r.Creates != 1 || len(r.comments) != 1 {
				t.Errorf("creates=%d comments=%d: a non-gone 404 led to a second ledger comment", r.Creates, len(r.comments))
			}
			// No silent failure: with every update refused, a Flush that reports
			// no error would have to show the newest text, which it cannot.
			if flushErr == nil && len(r.comments) == 1 && !strings.Contains(r.comments[0].Body, "second text") {
				t.Errorf("Flush reported success while the comment still shows the old text")
			}
			if errors.Is(flushErr, ctxledger.ErrGone) {
				t.Errorf("a non-gone 404 was reported as ErrGone")
			}
		})
	}
}
