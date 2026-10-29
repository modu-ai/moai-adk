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
	g := GHRemote{Repo: "o/r/hooks/x", Run: rr.run}
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

// weOwnEverything is an Identity for which every comment was written by the authenticated user.
type weOwnEverything struct{}

func (weOwnEverything) CurrentUserID(context.Context) (int64, error)              { return 1001, nil }
func (weOwnEverything) IssueAuthorID(context.Context, string, int) (int64, error) { return 1001, nil }
func (weOwnEverything) CommentAuthorID(context.Context, string, int64) (int64, error) {
	return 1001, nil
}

func ownEnforcer() *effect.Enforcer {
	return &effect.Enforcer{Identity: weOwnEverything{}, Judge: func(op effect.IssueOp, s effect.Subject) effect.Decision {
		return effect.Decision{Allow: s.ActorKnown && s.AuthorKnown && s.ActorID == s.AuthorID && op == effect.OpCommentEdit}
	}}
}

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
func TestCommentEditWithoutAnEnforcerIsRefusedAsIdentityUnknown(t *testing.T) {
	rr := &recRun{}
	err := (GHRemote{Repo: "o/r", Run: rr.run}).Update(t.Context(), 7, 5, "b")
	var re *effect.RefusedError
	if !errors.As(err, &re) || re.Reason != effect.ReasonIdentityUnknown || len(rr.calls) != 0 {
		t.Errorf("err=%v calls=%v", err, rr.calls)
	}
}
