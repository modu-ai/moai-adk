package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ctxledger"
)

// Issue #1948: secret, email or home-path input publishes nothing. This was a
// RED test (the two values were published as-is, observed at ed364ebdcd) until
// lane-20 b1fed052c7; promoted to a regular regression test once it passed.
// The fixtures are assembled at run time and are not real personal data.
func TestCtxLedger_EmailAndHomePath_PublishNothing(t *testing.T) {
	email := "person" + "@" + "example.org"
	home := "/Us" + "ers/someone/project"
	r := &ctxRemote{}
	l := newCtxLedger(r, effect.NewMemIntentStore())
	ctx := context.Background()
	if _, err := l.Submit(ctx, ctxledger.Body{Issue: 1948, Text: "owner " + email + " worktree " + home, MaskPassed: true}); err == nil {
		if err := l.Flush(ctx); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range r.comments {
		if strings.Contains(c.Body, email) {
			t.Error("an email address reached the published comment")
		}
		if strings.Contains(c.Body, home) {
			t.Error("a home path reached the published comment")
		}
	}
}
