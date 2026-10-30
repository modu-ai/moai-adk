package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// The context a pass holds reaches the host lookup: cancelling the pass cancels the
// lookup instead of leaving it to run to its own time limit.
func TestHostFactsLookupFollowsTheCallersContext(t *testing.T) {
	ref, _ := core.NewIssueRef(1948)
	h := hostWithAuthor(99)
	started := make(chan struct{})
	h.User = func(ctx context.Context) (int64, error) {
		close(started)
		<-ctx.Done()
		return 0, ctx.Err()
	}
	h.Timeout = time.Hour // far beyond the test; only the caller's cancel can end it
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := h.IssueFacts(ctx, ref); done <- err }()
	<-started
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("a cancelled pass must end the lookup with the cancel: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the lookup did not follow the caller's cancel")
	}
}

// ctxFacts records the context the runner hands to the host lookup.
type ctxFacts struct{ got context.Context }

func (c *ctxFacts) IssueFacts(ctx context.Context, _ core.IssueRef) (IssueFacts, error) {
	c.got = ctx
	return IssueFacts{RequestedRepo: "acme/widgets", ActualRepo: "acme/widgets", UserID: 7, AuthorID: 7, Owner: core.OwnerNone}, nil
}

type ctxKey struct{}

func TestRunnerHandsItsContextToTheHostLookup(t *testing.T) {
	r, ref := newRunner(t)
	f := &ctxFacts{}
	r.Facts = f
	pass := *r // a pass takes its own copy and gives it its own context
	pass.Ctx = context.WithValue(context.Background(), ctxKey{}, "pass-1")
	if _, err := pass.Claim(ref, "sess-a", "disp-1", "run-1"); err != nil {
		t.Fatal(err)
	}
	if f.got == nil || f.got.Value(ctxKey{}) != "pass-1" {
		t.Errorf("the lookup did not get the pass's context: %v", f.got)
	}
	// without one, a lookup still gets a usable context
	f.got = nil
	other, _ := core.NewIssueRef(2001)
	if _, err := r.Claim(other, "sess-b", "disp-2", "run-1"); err != nil {
		t.Fatal(err)
	}
	if f.got == nil {
		t.Error("a runner without a context must hand the lookup a background one")
	}
}
