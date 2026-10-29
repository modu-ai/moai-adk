package ctxledger

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

type recRun struct{ calls [][]string }

func (r *recRun) run(_ context.Context, args ...string) ([]byte, error) {
	r.calls = append(r.calls, args)
	return []byte(`{"id":1}`), nil
}

func TestGHRemoteRefusesAPathThatIsNotACommentRoute(t *testing.T) {
	rr := &recRun{}
	g := GHRemote{Repo: "o/r/hooks/x", Run: rr.run, Enforcer: ownEnforcer()}
	if _, err := g.Create(t.Context(), 7, "b"); !errors.Is(err, effect.ErrRefused) {
		t.Errorf("create err = %v", err)
	}
	if err := g.Update(t.Context(), 7, 5, "b"); !errors.Is(err, effect.ErrRefused) {
		t.Errorf("update err = %v", err)
	}
	if len(rr.calls) != 0 {
		t.Errorf("gh ran: %v", rr.calls)
	}
}

// weOwnEverything is an Identity for which the authenticated user wrote every comment and opened every Issue.
type weOwnEverything struct{}

func (weOwnEverything) CurrentUserID(context.Context) (int64, error) { return 1001, nil }
func (weOwnEverything) IssueFacts(_ context.Context, repo string, _ int) (effect.IssueFacts, error) {
	return effect.IssueFacts{AuthorID: 1001, ActualRepo: repo}, nil
}
func (weOwnEverything) CommentFacts(_ context.Context, repo string, _ int64) (effect.CommentFacts, error) {
	return effect.CommentFacts{AuthorID: 1001, Issue: 7, ActualRepo: repo}, nil
}

func ownEnforcer() *effect.Enforcer { return &effect.Enforcer{Identity: weOwnEverything{}} }

func TestEveryGHCallOfTheRemoteIsAnAllowlistedRoute(t *testing.T) {
	rr := &recRun{}
	g := GHRemote{Repo: "o/r", Run: rr.run, Enforcer: ownEnforcer()}
	_, _, _ = g.Find(t.Context(), 7, "m")
	_, _ = g.Create(t.Context(), 7, "b")
	_ = g.Update(t.Context(), 7, 5, "b")
	if len(rr.calls) != 3 {
		t.Fatalf("calls = %d", len(rr.calls))
	}
	for _, c := range rr.calls {
		if strings.Contains(strings.Join(c, " "), "protection") {
			t.Errorf("unexpected call %v", c)
		}
	}
}

// Without an identity check wired, a comment edit is refused and gh never runs.
func TestCommentWritesWithoutAnEnforcerAreRefusedAsIdentityUnknown(t *testing.T) {
	rr := &recRun{}
	g := GHRemote{Repo: "o/r", Run: rr.run}
	for name, err := range map[string]error{
		"edit":   g.Update(t.Context(), 7, 5, "b"),
		"create": func() error { _, e := g.Create(t.Context(), 7, "b"); return e }(),
	} {
		var re *effect.RefusedError
		if !errors.As(err, &re) || re.Reason != effect.ReasonIdentityUnknown {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	if len(rr.calls) != 0 {
		t.Errorf("gh ran: %v", rr.calls)
	}
}
