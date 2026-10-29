package effect

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// fakeIdentity answers with numeric ids and the repository the "host" answered for, and records what it was
// asked. It has no write method at all.
type fakeIdentity struct {
	actor    int64
	actorErr error
	issues   map[int]IssueFacts
	comments map[int64]CommentFacts
	asked    []string
}

func (f *fakeIdentity) CurrentUserID(context.Context) (int64, error) {
	f.asked = append(f.asked, "user")
	return f.actor, f.actorErr
}

func (f *fakeIdentity) IssueFacts(_ context.Context, repo string, n int) (IssueFacts, error) {
	f.asked = append(f.asked, fmt.Sprintf("issue %s#%d", repo, n))
	v, ok := f.issues[n]
	if !ok {
		return IssueFacts{}, errors.New("lookup failed")
	}
	return v, nil
}

func (f *fakeIdentity) CommentFacts(_ context.Context, repo string, c int64) (CommentFacts, error) {
	f.asked = append(f.asked, fmt.Sprintf("comment %s/%d", repo, c))
	v, ok := f.comments[c]
	if !ok {
		return CommentFacts{}, errors.New("lookup failed")
	}
	return v, nil
}

const me, other = int64(1001), int64(2002)

// world: issue 7 is ours, issue 8 belongs to someone else; comment 501 is ours (on issue 8), 502 is theirs.
func world() *fakeIdentity {
	return &fakeIdentity{
		actor: me,
		issues: map[int]IssueFacts{
			7: {AuthorID: me, ActualRepo: "o/r"},
			8: {AuthorID: other, ActualRepo: "o/r"},
		},
		comments: map[int64]CommentFacts{
			501: {AuthorID: me, Issue: 8, ActualRepo: "o/r"},
			502: {AuthorID: other, Issue: 8, ActualRepo: "o/r"},
		},
	}
}

func enforcer(f *fakeIdentity) *Enforcer { return &Enforcer{Identity: f} }

func reasonOf(t *testing.T, err error) RefusalReason {
	t.Helper()
	var re *RefusedError
	if !errors.As(err, &re) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	return re.Reason
}

func patch(e *Enforcer, path, body string) error {
	return e.Check(context.Background(), "PATCH", path, nil, []byte(body))
}

func TestOwnIssueBodyEditAllowedSomeoneElsesRefusedByDefault(t *testing.T) {
	e := enforcer(world())
	if err := patch(e, "/repos/o/r/issues/7", `{"body":"new"}`); err != nil {
		t.Errorf("own issue refused: %v", err)
	}
	if r := reasonOf(t, patch(e, "/repos/o/r/issues/8", `{"body":"new"}`)); r != ReasonNotIssueAuthor {
		t.Errorf("someone else's issue: reason = %s", r)
	}
}

func TestOwnIssueCloseIsAllowedOwnLabelsAndAssigneesAreNot(t *testing.T) {
	e := enforcer(world())
	if err := patch(e, "/repos/o/r/issues/7", `{"state":"closed"}`); err != nil {
		t.Errorf("closing our own issue refused: %v", err)
	}
	for _, body := range []string{`{"labels":["x"]}`, `{"assignees":["a"]}`, `{"assignee":"a"}`} {
		if r := reasonOf(t, patch(e, "/repos/o/r/issues/7", body)); r != ReasonPermissionNotGranted {
			t.Errorf("%s on our own issue: reason = %s", body, r)
		}
	}
}

func TestSomeoneElsesIssueNeedsTheOptionTheRepositoryAndTheExactPermission(t *testing.T) {
	cases := map[string]struct {
		body string
		set  func(*core.Collaboration)
	}{
		"edit_body": {`{"body":"x"}`, func(c *core.Collaboration) { c.EditBody = true }},
		"close":     {`{"state":"closed"}`, func(c *core.Collaboration) { c.Close = true }},
		"labels":    {`{"labels":["x"]}`, func(c *core.Collaboration) { c.Labels = true }},
		"assignees": {`{"assignees":["a"]}`, func(c *core.Collaboration) { c.Assignees = true }},
	}
	on := func(set func(*core.Collaboration)) core.Collaboration {
		c := core.Collaboration{Provider: "github", Enabled: true, AllowedRepositories: []string{"o/r"}}
		set(&c)
		return c
	}
	for have, h := range cases {
		for name, c := range cases {
			e := enforcer(world())
			e.Collaboration = on(h.set)
			err := patch(e, "/repos/o/r/issues/8", c.body)
			if (name == have) != (err == nil) {
				t.Errorf("permission %s, operation %s: err = %v", have, name, err)
			}
		}
	}
	// option off: a permission alone opens nothing
	e := enforcer(world())
	e.Collaboration = core.Collaboration{Provider: "github", AllowedRepositories: []string{"o/r"}, EditBody: true, Close: true}
	if r := reasonOf(t, patch(e, "/repos/o/r/issues/8", `{"body":"x"}`)); r != ReasonNotIssueAuthor {
		t.Errorf("option off: reason = %s", r)
	}
	// repository not on the list
	e.Collaboration = core.Collaboration{Provider: "github", Enabled: true, AllowedRepositories: []string{"elsewhere/x"}, EditBody: true}
	if r := reasonOf(t, patch(e, "/repos/o/r/issues/8", `{"body":"x"}`)); r != ReasonRepoNotAllowed {
		t.Errorf("repository not listed: reason = %s", r)
	}
}

func TestARedirectOrADifferentAnsweringRepositoryBlocksForeignWrites(t *testing.T) {
	collab := core.Collaboration{Provider: "github", Enabled: true, AllowedRepositories: []string{"o/r"}, EditBody: true}
	moved := world()
	moved.issues[8] = IssueFacts{AuthorID: other, ActualRepo: "o/r", Redirected: true}
	e := &Enforcer{Identity: moved, Collaboration: collab}
	if r := reasonOf(t, patch(e, "/repos/o/r/issues/8", `{"body":"x"}`)); r != ReasonRepoNotAllowed {
		t.Errorf("redirected: reason = %s", r)
	}
	renamed := world()
	renamed.issues[8] = IssueFacts{AuthorID: other, ActualRepo: "o/new-name"}
	e = &Enforcer{Identity: renamed, Collaboration: collab}
	if r := reasonOf(t, patch(e, "/repos/o/r/issues/8", `{"body":"x"}`)); r != ReasonRepoNotAllowed {
		t.Errorf("answered for another repository: reason = %s", r)
	}
	unknownRepo := world()
	unknownRepo.issues[8] = IssueFacts{AuthorID: other}
	e = &Enforcer{Identity: unknownRepo, Collaboration: collab}
	if r := reasonOf(t, patch(e, "/repos/o/r/issues/8", `{"body":"x"}`)); r != ReasonRepoNotAllowed {
		t.Errorf("answering repository unknown: reason = %s", r)
	}
}

func TestCommentCreateWorksOnAnyIssueButNeedsKnownIds(t *testing.T) {
	e := enforcer(world())
	for _, p := range []string{"/repos/o/r/issues/7/comments", "/repos/o/r/issues/8/comments"} {
		if err := e.Check(context.Background(), "POST", p, nil, []byte(`{"body":"x"}`)); err != nil {
			t.Errorf("%s refused: %v", p, err)
		}
	}
	blind := &fakeIdentity{actorErr: errors.New("x"), issues: map[int]IssueFacts{8: {AuthorID: other, ActualRepo: "o/r"}}}
	if r := reasonOf(t, enforcer(blind).Check(context.Background(), "POST", "/repos/o/r/issues/8/comments", nil, []byte(`{}`))); r != ReasonIdentityUnknown {
		t.Errorf("unknown user: reason = %s", r)
	}
}

func TestCommentEditOnlyForACommentWeWroteOnAnyIssue(t *testing.T) {
	e := enforcer(world())
	if err := e.Check(context.Background(), "PATCH", "/repos/o/r/issues/comments/501", nil, []byte(`{"body":"x"}`)); err != nil {
		t.Errorf("own comment on someone else's issue refused: %v", err)
	}
	if r := reasonOf(t, e.Check(context.Background(), "PATCH", "/repos/o/r/issues/comments/502", nil, nil)); r != ReasonNotIssueAuthor {
		t.Errorf("someone else's comment: reason = %s", r)
	}
	// no setting widens it
	e.Collaboration = core.Collaboration{Provider: "github", Enabled: true, AllowedRepositories: []string{"o/r"}, EditBody: true, Close: true, Labels: true, Assignees: true}
	if r := reasonOf(t, e.Check(context.Background(), "PATCH", "/repos/o/r/issues/comments/502", nil, nil)); r != ReasonNotIssueAuthor {
		t.Errorf("with every permission on: reason = %s", r)
	}
}

func TestUnknownFactsRefuseAsIdentityUnknown(t *testing.T) {
	w := func(mut func(*fakeIdentity)) *fakeIdentity { f := world(); mut(f); return f }
	for name, f := range map[string]*fakeIdentity{
		"user lookup fails":  w(func(f *fakeIdentity) { f.actorErr = errors.New("boom") }),
		"user id zero":       w(func(f *fakeIdentity) { f.actor = 0 }),
		"user id negative":   w(func(f *fakeIdentity) { f.actor = -3 }),
		"issue lookup fails": w(func(f *fakeIdentity) { f.issues = map[int]IssueFacts{} }),
		"issue author zero": w(func(f *fakeIdentity) {
			f.issues[7] = IssueFacts{ActualRepo: "o/r"}
			f.issues[8] = IssueFacts{ActualRepo: "o/r"}
		}),
		"comment lookup fails":  w(func(f *fakeIdentity) { f.comments = map[int64]CommentFacts{} }),
		"comment author zero":   w(func(f *fakeIdentity) { f.comments[501] = CommentFacts{Issue: 8, ActualRepo: "o/r"} }),
		"comment issue unknown": w(func(f *fakeIdentity) { f.comments[501] = CommentFacts{AuthorID: me, ActualRepo: "o/r"} }),
	} {
		e := enforcer(f)
		// a broken comment lookup cannot affect an Issue PATCH, which never reads a comment
		if !strings.HasPrefix(name, "comment ") {
			if r := reasonOf(t, patch(e, "/repos/o/r/issues/7", `{"body":"x"}`)); r != ReasonIdentityUnknown {
				t.Errorf("%s (issue patch): reason = %s", name, r)
			}
		}
		if r := reasonOf(t, e.Check(context.Background(), "PATCH", "/repos/o/r/issues/comments/501", nil, nil)); r != ReasonIdentityUnknown {
			t.Errorf("%s (comment edit): reason = %s", name, r)
		}
	}
}

func TestNothingWiredFailsClosedButPlainWritesStillPass(t *testing.T) {
	for name, e := range map[string]*Enforcer{"nil enforcer": nil, "no identity": {}} {
		for _, c := range [][3]string{
			{"PATCH", "/repos/o/r/issues/comments/501", ``},
			{"POST", "/repos/o/r/issues/7/comments", `{"body":"x"}`},
			{"PATCH", "/repos/o/r/issues/7", `{"body":"x"}`},
		} {
			if r := reasonOf(t, e.Check(context.Background(), c[0], c[1], nil, []byte(c[2]))); r != ReasonIdentityUnknown {
				t.Errorf("%s: %s %s reason = %s", name, c[0], c[1], r)
			}
		}
		for _, ok := range [][2]string{{"POST", "/repos/o/r/issues"}, {"POST", "/repos/o/r/labels"}, {"PATCH", "/repos/o/r/labels/p0"}, {"GET", "/repos/o/r/issues/8"}} {
			if err := e.Check(context.Background(), ok[0], ok[1], nil, []byte(`{}`)); err != nil {
				t.Errorf("%s: %s %s refused: %v", name, ok[0], ok[1], err)
			}
		}
	}
}

func TestStaticGuardStillAppliesAndCollaborationNeverOverridesIt(t *testing.T) {
	e := &Enforcer{Identity: world(), Collaboration: core.Collaboration{Provider: "github", Enabled: true, AllowedRepositories: []string{"o/r"}, EditBody: true, Close: true, Labels: true, Assignees: true}}
	for _, c := range [][2]string{
		{"PUT", "/repos/o/r/branches/main/protection"},
		{"POST", "/repos/o/r/hooks"},
		{"DELETE", "/repos/o/r/issues/7"},
		{"PATCH", "/repos/o/r/issues/7/labels"},
		{"PUT", "/repos/o/r/issues/7/lock"},
	} {
		if err := e.Check(context.Background(), c[0], c[1], nil, []byte(`{"body":"x"}`)); !errors.Is(err, ErrRefused) {
			t.Errorf("%s %s: err = %v", c[0], c[1], err)
		}
	}
}

func TestBadBodiesAndFieldsOutsideThePolicyAreRefused(t *testing.T) {
	e := enforcer(world())
	for _, body := range []string{"", "not json", "[]", "{}", "null", `{"title":"x"}`, `{"milestone":1}`, `{"state":"open"}`, `{"body":"x","title":"y"}`} {
		if r := reasonOf(t, patch(e, "/repos/o/r/issues/7", body)); r != ReasonMalformed {
			t.Errorf("body %q: reason = %s", body, r)
		}
	}
	for _, p := range []string{"/repos/o/r/issues/abc", "/repos/o/r/issues/0", "/repos/o/r/issues/comments/abc", "/repos/o/r/issues/7/comments/9"} {
		if err := patch(e, p, `{"body":"x"}`); err == nil {
			t.Errorf("%s was allowed", p)
		}
	}
}

func TestLookupsAskForTheNumbersTheRouteNames(t *testing.T) {
	f := world()
	e := enforcer(f)
	_ = patch(e, "/repos/o/r/issues/7", `{"body":"x"}`)
	_ = e.Check(context.Background(), "PATCH", "/repos/o/r/issues/comments/501", nil, nil)
	want := []string{"user", "issue o/r#7", "user", "comment o/r/501", "issue o/r#8"}
	if fmt.Sprint(f.asked) != fmt.Sprint(want) {
		t.Errorf("asked %v, want %v", f.asked, want)
	}
}

func TestEveryDenialParksTheIntentWithItsReason(t *testing.T) {
	for _, reason := range []RefusalReason{ReasonNotIssueAuthor, ReasonIdentityUnknown, ReasonRepoNotAllowed, ReasonPermissionNotGranted} {
		s := &fakeSender{err: &RefusedError{Reason: reason, Method: "PATCH", Path: "/repos/o/r/issues/8"}}
		lk := &fakeLookup{}
		ex := newExec(s, lk, NewMemIntentStore())
		in, _ := ex.Execute(context.Background(), req())
		if in.Status != StatusParked || in.ParkReason != string(reason) || lk.calls != 0 {
			t.Errorf("%s: in=%+v lookups=%d", reason, in, lk.calls)
		}
		if again, _ := ex.Execute(context.Background(), req()); again.Status != StatusParked || s.calls != 1 {
			t.Errorf("%s: sent again (%d)", reason, s.calls)
		}
	}
}

// The reason strings the guard records are the strings core records for the same denial.
func TestReasonStringsMatchCore(t *testing.T) {
	pairs := map[RefusalReason]core.BlockReason{
		ReasonNotIssueAuthor: core.ReasonNotIssueAuthor, ReasonIdentityUnknown: core.ReasonIdentityUnknown,
		ReasonRepoNotAllowed: core.ReasonRepoNotAllowed, ReasonPermissionNotGranted: core.ReasonPermissionNotGranted,
		ReasonOwnerUnconfirmed: core.ReasonOwnerUnconfirmed,
	}
	for ours, theirs := range pairs {
		if string(ours) != string(theirs) {
			t.Errorf("%q != core %q", ours, theirs)
		}
	}
}

func TestCheckRequestAloneNeverLetsAConditionalRoutePass(t *testing.T) {
	for _, c := range [][2]string{{"PATCH", "/repos/o/r/issues/comments/501"}, {"PATCH", "/repos/o/r/issues/7"}, {"POST", "/repos/o/r/issues/7/comments"}} {
		if r := reasonOf(t, CheckRequest(c[0], c[1], nil)); r != ReasonIdentityUnknown {
			t.Errorf("%s %s: reason = %s", c[0], c[1], r)
		}
	}
}
