//go:build afred

package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ctxledger"
)

// RED on purpose (Issue #1948: "secret, email or home-path input => publish 0").
// Observed today (run without the tag, earlier probe): a body carrying an email
// address and a home path with MaskPassed=true is stored and published as-is,
// because neither the ctxledger secret check nor the masking it relies on
// covers those two kinds. Expected: the text is refused or the two values never
// reach the comment. Run with: go test -tags afred ./internal/agenticfactory/e2e/ -run RED_
func TestRED_CtxLedger_EmailAndHomePath_PublishNothing(t *testing.T) {
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
