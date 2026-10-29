package e2e

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

// The write boundary: effect.Enforcer applies the static guard and then the
// Issue-author policy (core.Collaboration) using read-only identity lookups. The
// Identity here is a fake host: numeric ids and the repository it "answered for".
// A refusal is a decision made before any write is built; no write exists here.

type fakeIdentity struct {
	mu       sync.Mutex
	user     int64
	userErr  error
	issues   map[int]effect.IssueFacts
	comments map[int64]effect.CommentFacts
	lookups  int
}

func (f *fakeIdentity) CurrentUserID(context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lookups++
	return f.user, f.userErr
}

func (f *fakeIdentity) IssueFacts(_ context.Context, _ string, issue int) (effect.IssueFacts, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lookups++
	if v, ok := f.issues[issue]; ok {
		return v, nil
	}
	return effect.IssueFacts{}, errors.New("issue lookup failed")
}

func (f *fakeIdentity) CommentFacts(_ context.Context, _ string, id int64) (effect.CommentFacts, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lookups++
	if v, ok := f.comments[id]; ok {
		return v, nil
	}
	return effect.CommentFacts{}, errors.New("comment lookup failed")
}

const (
	pathCommentCreateOwn     = "/repos/Org/Repo/issues/7/comments" // issue 7: mine
	pathCommentCreateForeign = "/repos/Org/Repo/issues/8/comments" // issue 8: theirs
	pathIssueOwn             = "/repos/Org/Repo/issues/7"
	pathIssueForeign         = "/repos/Org/Repo/issues/8"
	pathCommentMine          = "/repos/Org/Repo/issues/comments/501" // my comment on their issue
	pathCommentTheirs        = "/repos/Org/Repo/issues/comments/502"
)

func host() *fakeIdentity {
	return &fakeIdentity{
		user: me,
		issues: map[int]effect.IssueFacts{
			7: {AuthorID: me, ActualRepo: repoA},
			8: {AuthorID: them, ActualRepo: repoA},
		},
		comments: map[int64]effect.CommentFacts{
			501: {AuthorID: me, Issue: 8, ActualRepo: repoA},
			502: {AuthorID: them, Issue: 8, ActualRepo: repoA},
		},
	}
}

func refusedWith(t *testing.T, err error, want effect.RefusalReason) {
	t.Helper()
	if !errors.Is(err, effect.ErrRefused) {
		t.Fatalf("want a refusal (%s), got %v", want, err)
	}
	var re *effect.RefusedError
	if !errors.As(err, &re) || re.Reason != want {
		t.Fatalf("refusal = %v, want reason %s", err, want)
	}
}

func check(e *effect.Enforcer, method, path string, body string) error {
	return e.Check(context.Background(), method, path, nil, []byte(body))
}

// Writes on the plain allowlist need no identity: no lookup is made.
func TestEnforcer_PlainAllowlist_NeedsNoLookup(t *testing.T) {
	id := host()
	e := &effect.Enforcer{Identity: id}
	for _, tc := range []struct{ method, path string }{
		{"POST", "/repos/Org/Repo/issues"}, {"POST", "/repos/Org/Repo/labels"}, {"PATCH", "/repos/Org/Repo/labels/p0"}, {"GET", "/repos/Org/Repo/issues/8"},
	} {
		if err := check(e, tc.method, tc.path, ""); err != nil {
			t.Errorf("%s %s refused: %v", tc.method, tc.path, err)
		}
	}
	if id.lookups != 0 {
		t.Errorf("identity lookups = %d, want 0 for unconditional writes", id.lookups)
	}
}

// Comments: allowed on your own Issue and on another author's; your own comment
// may be edited, another person's may not.
func TestEnforcer_Comments_OwnAndOthersIssues_OnlyOwnCommentsEdited(t *testing.T) {
	e := &effect.Enforcer{Identity: host()}
	if err := check(e, "POST", pathCommentCreateOwn, `{"body":"x"}`); err != nil {
		t.Errorf("comment on my Issue: %v", err)
	}
	if err := check(e, "POST", pathCommentCreateForeign, `{"body":"x"}`); err != nil {
		t.Errorf("comment on another author's Issue: %v", err)
	}
	if err := check(e, "PATCH", pathCommentMine, `{"body":"x"}`); err != nil {
		t.Errorf("edit my own comment: %v", err)
	}
	refusedWith(t, check(e, "PATCH", pathCommentTheirs, `{"body":"x"}`), effect.ReasonNotIssueAuthor)
}

// Unknown identity always refuses: a failed or zero user lookup, a missing Issue
// fact, a missing comment fact, or no Identity wired at all.
func TestEnforcer_UnknownIdentity_RefusesEveryConditionalWrite(t *testing.T) {
	cases := map[string]*effect.Enforcer{
		"user lookup fails": {Identity: func() *fakeIdentity { h := host(); h.userErr = errors.New("auth lookup failed"); return h }()},
		"user id 0":         {Identity: func() *fakeIdentity { h := host(); h.user = 0; return h }()},
		"no issue facts":    {Identity: func() *fakeIdentity { h := host(); h.issues = nil; return h }()},
		"issue author 0": {Identity: func() *fakeIdentity {
			h := host()
			h.issues[7] = effect.IssueFacts{AuthorID: 0, ActualRepo: repoA}
			h.issues[8] = effect.IssueFacts{AuthorID: 0, ActualRepo: repoA}
			return h
		}()},
		"no identity wired": {},
	}
	writes := []struct{ method, path, body string }{
		{"POST", pathCommentCreateOwn, `{"body":"x"}`},
		{"POST", pathCommentCreateForeign, `{"body":"x"}`},
		{"PATCH", pathCommentMine, `{"body":"x"}`},
		{"PATCH", pathIssueOwn, `{"body":"x"}`},
	}
	for name, e := range cases {
		for _, w := range writes {
			refusedWith(t, check(e, w.method, w.path, w.body), effect.ReasonIdentityUnknown)
		}
		t.Logf("%s: all %d conditional writes refused as identity-unknown", name, len(writes))
	}
	var nilEnforcer *effect.Enforcer
	refusedWith(t, check(nilEnforcer, "POST", pathCommentCreateOwn, `{"body":"x"}`), effect.ReasonIdentityUnknown)
}

// A comment edit whose comment facts cannot be read is refused: the comment's
// author is not guessed.
func TestEnforcer_CommentFactsMissing_IsRefused(t *testing.T) {
	h := host()
	h.comments = nil
	refusedWith(t, check(&effect.Enforcer{Identity: h}, "PATCH", pathCommentMine, `{"body":"x"}`), effect.ReasonIdentityUnknown)
}

// Issue PATCH: the JSON keys decide the operation. Your own Issue's body and
// closing are yours; another author's are refused by default.
func TestEnforcer_IssuePatch_OwnVersusForeign_ByTheBodyKeys(t *testing.T) {
	e := &effect.Enforcer{Identity: host()}
	if err := check(e, "PATCH", pathIssueOwn, `{"body":"new text"}`); err != nil {
		t.Errorf("edit my Issue's body: %v", err)
	}
	if err := check(e, "PATCH", pathIssueOwn, `{"state":"closed"}`); err != nil {
		t.Errorf("close my Issue: %v", err)
	}
	for _, body := range []string{`{"body":"x"}`, `{"state":"closed"}`, `{"labels":["p0"]}`, `{"assignees":["someone"]}`} {
		refusedWith(t, check(e, "PATCH", pathIssueForeign, body), effect.ReasonNotIssueAuthor)
	}
	// On your own Issue, labels and assignees need their own permission (off by default).
	refusedWith(t, check(e, "PATCH", pathIssueOwn, `{"labels":["p0"]}`), effect.ReasonPermissionNotGranted)
	refusedWith(t, check(e, "PATCH", pathIssueOwn, `{"assignees":["someone"]}`), effect.ReasonPermissionNotGranted)
}

// A PATCH carrying several fields is allowed only if EVERY operation is.
func TestEnforcer_IssuePatch_AllOperationsMustBeAllowed(t *testing.T) {
	e := &effect.Enforcer{Identity: host()}
	refusedWith(t, check(e, "PATCH", pathIssueOwn, `{"body":"ok","labels":["p0"]}`), effect.ReasonPermissionNotGranted)
}

// A PATCH body that is not a recognizable edit (empty, non-object, not JSON) is
// refused as malformed, not guessed.
func TestEnforcer_IssuePatch_UnreadableBodies_AreMalformed(t *testing.T) {
	e := &effect.Enforcer{Identity: host()}
	for _, body := range []string{`{}`, `[]`, `not json`, ``} {
		refusedWith(t, check(e, "PATCH", pathIssueOwn, body), effect.ReasonMalformed)
	}
}

// An operation the Issue permission policy does not define (reopening, title,
// milestone) gets its own reason, on my Issue and on another author's alike.
func TestEnforcer_IssuePatch_UndefinedOperations_AreOpNotDefined(t *testing.T) {
	e := &effect.Enforcer{Identity: host()}
	for _, path := range []string{pathIssueOwn, pathIssueForeign} {
		for _, body := range []string{`{"state":"open"}`, `{"title":"x"}`, `{"milestone":3}`} {
			refusedWith(t, check(e, "PATCH", path, body), effect.ReasonOpNotDefined)
		}
	}
}

// Collaboration widens exactly what its keys say, on a listed repository.
func TestEnforcer_Collaboration_OpensOnlyTheKeyedOperations(t *testing.T) {
	mk := func(mod func(*core.Collaboration)) *effect.Enforcer {
		return &effect.Enforcer{Identity: host(), Collaboration: collab(func(c *core.Collaboration) {
			c.Enabled, c.AllowedRepositories = true, []string{repoA}
			mod(c)
		})}
	}
	e := mk(func(c *core.Collaboration) { c.EditBody = true })
	if err := check(e, "PATCH", pathIssueForeign, `{"body":"x"}`); err != nil {
		t.Errorf("body edit with edit_body on: %v", err)
	}
	refusedWith(t, check(e, "PATCH", pathIssueForeign, `{"state":"closed"}`), effect.ReasonPermissionNotGranted)
	refusedWith(t, check(e, "PATCH", pathIssueForeign, `{"labels":["p0"]}`), effect.ReasonPermissionNotGranted)

	closer := mk(func(c *core.Collaboration) { c.Close = true })
	if err := check(closer, "PATCH", pathIssueForeign, `{"state":"closed"}`); err != nil {
		t.Errorf("close with close on: %v", err)
	}
	refusedWith(t, check(closer, "PATCH", pathIssueForeign, `{"body":"x"}`), effect.ReasonPermissionNotGranted)
}

// The option on but the repository off the list: refused whatever the keys say.
func TestEnforcer_Collaboration_RepositoryOffTheList_IsRefused(t *testing.T) {
	e := &effect.Enforcer{Identity: host(), Collaboration: collab(func(c *core.Collaboration) {
		c.Enabled, c.AllowedRepositories = true, []string{repoB}
		c.EditBody, c.Close, c.Labels, c.Assignees = true, true, true, true
	})}
	refusedWith(t, check(e, "PATCH", pathIssueForeign, `{"body":"x"}`), effect.ReasonRepoNotAllowed)
}

// A redirect or a different repository in the host's answer is not trusted.
func TestEnforcer_Collaboration_RedirectOrRename_IsRefused(t *testing.T) {
	for name, facts := range map[string]effect.IssueFacts{
		"redirect": {AuthorID: them, ActualRepo: repoA, Redirected: true},
		"renamed":  {AuthorID: them, ActualRepo: repoB},
		"unknown":  {AuthorID: them, ActualRepo: ""},
	} {
		h := host()
		h.issues[8] = facts
		e := &effect.Enforcer{Identity: h, Collaboration: collab(func(c *core.Collaboration) {
			c.Enabled, c.AllowedRepositories, c.EditBody = true, []string{repoA}, true
		})}
		refusedWith(t, check(e, "PATCH", pathIssueForeign, `{"body":"x"}`), effect.ReasonRepoNotAllowed)
		t.Logf("%s: refused as repository-not-allowed", name)
	}
}

// Two people with the same display name but different ids are different people:
// the Identity interface exposes ids only, so a name can never stand in for one.
func TestEnforcer_SameNameDifferentID_NeverMistaken(t *testing.T) {
	h := host()
	h.issues[8] = effect.IssueFacts{AuthorID: me + 1, ActualRepo: repoA} // someone else, whatever they are called
	refusedWith(t, check(&effect.Enforcer{Identity: h}, "PATCH", pathIssueForeign, `{"body":"x"}`), effect.ReasonNotIssueAuthor)
}

// Static refusals still come first: a protection path is refused without a lookup.
func TestEnforcer_ProtectionPaths_AreRefusedBeforeAnyLookup(t *testing.T) {
	id := host()
	e := &effect.Enforcer{Identity: id, Collaboration: collab(func(c *core.Collaboration) {
		c.Enabled, c.AllowedRepositories = true, []string{repoA}
		c.EditBody, c.Close, c.Labels, c.Assignees = true, true, true, true
	})}
	refusedWith(t, check(e, "PUT", "/repos/Org/Repo/branches/main/protection", `{"enforce_admins":false}`), effect.ReasonWeakensProtection)
	refusedWith(t, check(e, "DELETE", "/repos/Org/Repo/branches/main/protection/enforce_admins", ``), effect.ReasonWeakensProtection)
	if id.lookups != 0 {
		t.Errorf("identity lookups = %d, want 0: a static refusal needs no facts", id.lookups)
	}
}
