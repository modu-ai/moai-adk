//go:build afred

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ctxledger"
)

// RED on purpose (Issue #1948 S52 M1 follow-up). After lane-20 commit 8663240cd5
// the ledger remembers the last comment id. When a person deletes that comment
// on GitHub, the next update goes to the stale id and fails with 404; the
// following Flush returns nil with zero comments on the Issue, and only a second
// Flush recreates it (measured: flush1 err=nil comments=0, flush2 comments=1).
// Before the change the same scenario created the comment in the first Flush
// (finds=2 creates=2 updates=0 comments=1, taken from the pre-change copy).
// Expected: the first Flush either leaves exactly one comment at the newest text
// or reports an error; it must not look successful with the comment missing.
// Run with: go test -tags afred ./internal/agenticfactory/e2e/ -run RED_
func TestRED_CtxLedger_DeletedRemotely_FirstFlushNotSilent(t *testing.T) {
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
	if err == nil && len(r.comments) != 1 {
		t.Errorf("Flush returned nil but the Issue has %d ledger comments (want 1, or an error)", len(r.comments))
	}
}
