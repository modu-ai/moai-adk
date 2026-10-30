package e2e

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// Issue authorization (core.Collaboration.Authorize) over a fake host: the
// inputs are what a host answer would fill in (numeric ids, the repository the
// host answered for, whether it redirected). Nothing here touches a real Issue.
// Provenance note: the policy was relayed to the leader session without a stated
// source; these tests pin lane-18's implementation of it, they do not decide it.

const (
	me    int64 = 10 // the authenticated user
	them  int64 = 20 // another author
	repoA       = "Org/Repo"
	repoB       = "Org/Other"
)

func ownIssue(op core.Op) core.AuthzRequest {
	return core.AuthzRequest{Op: op, RequestedRepo: repoA, ActualRepo: repoA, UserID: me, IssueAuthorID: me, CommentAuthorID: me}
}

func foreignIssue(op core.Op) core.AuthzRequest {
	return core.AuthzRequest{Op: op, RequestedRepo: repoA, ActualRepo: repoA, UserID: me, IssueAuthorID: them, CommentAuthorID: me, Owner: core.OwnerNone}
}

func collab(mod func(*core.Collaboration)) core.Collaboration {
	c := core.Collaboration{Provider: "github"}
	if mod != nil {
		mod(&c)
	}
	return c
}

// Default (everything off): your own Issues are yours except labels and
// assignees; on another author's Issue only comments are allowed.
func TestAuthz_Defaults_OwnIssueVersusForeignIssue(t *testing.T) {
	c := collab(nil)
	for _, tc := range []struct {
		op   core.Op
		want core.AuthzVerdict
	}{
		{core.OpEditBody, core.Allow}, {core.OpCommentCreate, core.Allow}, {core.OpCommentEdit, core.Allow},
		{core.OpClose, core.Allow}, {core.OpTakeOver, core.Allow},
		{core.OpLabel, core.DenyPermissionNotGranted}, {core.OpAssignee, core.DenyPermissionNotGranted},
	} {
		if got := c.Authorize(ownIssue(tc.op)); got != tc.want {
			t.Errorf("own issue %s -> %s, want %s", tc.op, got, tc.want)
		}
	}
	for _, tc := range []struct {
		op   core.Op
		want core.AuthzVerdict
	}{
		{core.OpCommentCreate, core.Allow}, {core.OpCommentEdit, core.Allow},
		{core.OpEditBody, core.DenyNotIssueAuthor}, {core.OpClose, core.DenyNotIssueAuthor},
		{core.OpLabel, core.DenyNotIssueAuthor}, {core.OpAssignee, core.DenyNotIssueAuthor},
		{core.OpTakeOver, core.DenyNotIssueAuthor},
	} {
		if got := c.Authorize(foreignIssue(tc.op)); got != tc.want {
			t.Errorf("foreign issue %s -> %s, want %s", tc.op, got, tc.want)
		}
	}
}

// Another person's comment is never edited, on any Issue, whoever owns the Issue.
func TestAuthz_OtherPersonsComment_NeverEdited(t *testing.T) {
	all := collab(func(c *core.Collaboration) {
		c.Enabled, c.AllowedRepositories = true, []string{repoA}
		c.EditBody, c.Close, c.Labels, c.Assignees = true, true, true, true
	})
	for name, r := range map[string]core.AuthzRequest{"own issue": ownIssue(core.OpCommentEdit), "foreign issue": foreignIssue(core.OpCommentEdit)} {
		r.CommentAuthorID = them
		if got := all.Authorize(r); got != core.DenyNotIssueAuthor {
			t.Errorf("%s: editing another person's comment -> %s, want not-issue-author", name, got)
		}
	}
}

// The option alone, or the list alone, opens nothing on a foreign Issue.
func TestAuthz_OptionAndListMustBothBeSet(t *testing.T) {
	onlyOption := collab(func(c *core.Collaboration) { c.Enabled = true })
	onlyList := collab(func(c *core.Collaboration) { c.AllowedRepositories = []string{repoA} })
	permsOnly := collab(func(c *core.Collaboration) { c.EditBody, c.Close, c.Labels, c.Assignees = true, true, true, true })
	for _, op := range []core.Op{core.OpEditBody, core.OpClose, core.OpLabel, core.OpAssignee, core.OpTakeOver} {
		if got := onlyOption.Authorize(foreignIssue(op)); got != core.DenyRepoNotAllowed {
			t.Errorf("option without a list, %s -> %s, want repository-not-allowed", op, got)
		}
		if got := onlyList.Authorize(foreignIssue(op)); got != core.DenyNotIssueAuthor {
			t.Errorf("list without the option, %s -> %s, want not-issue-author", op, got)
		}
		if got := permsOnly.Authorize(foreignIssue(op)); got != core.DenyNotIssueAuthor {
			t.Errorf("permissions without the option, %s -> %s, want not-issue-author", op, got)
		}
	}
}

// Option + list opens taking over only; the four writes are four separate
// permissions and each opens exactly its own operation.
func TestAuthz_EnabledWithList_OpensOnlyTakeOver_EachPermissionOpensOnlyItsOp(t *testing.T) {
	base := func(mod func(*core.Collaboration)) core.Collaboration {
		return collab(func(c *core.Collaboration) {
			c.Enabled, c.AllowedRepositories = true, []string{repoA}
			if mod != nil {
				mod(c)
			}
		})
	}
	writes := []core.Op{core.OpEditBody, core.OpClose, core.OpLabel, core.OpAssignee}
	noPerms := base(nil)
	if got := noPerms.Authorize(foreignIssue(core.OpTakeOver)); got != core.Allow {
		t.Errorf("take-over with option+list and no owner -> %s, want allow", got)
	}
	for _, op := range writes {
		if got := noPerms.Authorize(foreignIssue(op)); got != core.DenyPermissionNotGranted {
			t.Errorf("option+list, no permission, %s -> %s, want permission-not-granted", op, got)
		}
	}
	for name, mod := range map[string]struct {
		set   func(*core.Collaboration)
		opens core.Op
	}{
		"edit_body": {func(c *core.Collaboration) { c.EditBody = true }, core.OpEditBody},
		"close":     {func(c *core.Collaboration) { c.Close = true }, core.OpClose},
		"labels":    {func(c *core.Collaboration) { c.Labels = true }, core.OpLabel},
		"assignees": {func(c *core.Collaboration) { c.Assignees = true }, core.OpAssignee},
	} {
		c := base(mod.set)
		for _, op := range writes {
			want := core.DenyPermissionNotGranted
			if op == mod.opens {
				want = core.Allow
			}
			if got := c.Authorize(foreignIssue(op)); got != want {
				t.Errorf("only %s on, %s -> %s, want %s", name, op, got, want)
			}
		}
	}
}

// Taking over needs a confirmed owner state: none or handed-over only.
func TestAuthz_TakeOver_NeedsAConfirmedOwnerState(t *testing.T) {
	c := collab(func(c *core.Collaboration) { c.Enabled, c.AllowedRepositories = true, []string{repoA} })
	for owner, want := range map[core.OwnerState]core.AuthzVerdict{
		core.OwnerNone: core.Allow, core.OwnerHandedOver: core.Allow,
		core.OwnerActive: core.DenyOwnerUnconfirmed, core.OwnerUnknown: core.DenyOwnerUnconfirmed, "": core.DenyOwnerUnconfirmed,
	} {
		r := foreignIssue(core.OpTakeOver)
		r.Owner = owner
		if got := c.Authorize(r); got != want {
			t.Errorf("owner %q -> %s, want %s", owner, got, want)
		}
	}
}

// A repository outside the list is refused for every foreign-Issue operation
// beyond comments, even with the option and every permission on.
func TestAuthz_RepositoryOutsideTheList_IsRefused(t *testing.T) {
	c := collab(func(c *core.Collaboration) {
		c.Enabled, c.AllowedRepositories = true, []string{repoB}
		c.EditBody, c.Close, c.Labels, c.Assignees = true, true, true, true
	})
	for _, op := range []core.Op{core.OpEditBody, core.OpClose, core.OpLabel, core.OpAssignee, core.OpTakeOver} {
		if got := c.Authorize(foreignIssue(op)); got != core.DenyRepoNotAllowed {
			t.Errorf("%s in a repository off the list -> %s, want repository-not-allowed", op, got)
		}
	}
}

// Unknown identity always blocks, whatever else is on: ids of zero or below.
func TestAuthz_UnknownIdentity_BlocksEveryOperation(t *testing.T) {
	all := collab(func(c *core.Collaboration) {
		c.Enabled, c.AllowedRepositories = true, []string{repoA}
		c.EditBody, c.Close, c.Labels, c.Assignees = true, true, true, true
	})
	ops := []core.Op{core.OpEditBody, core.OpCommentCreate, core.OpCommentEdit, core.OpLabel, core.OpAssignee, core.OpClose, core.OpTakeOver}
	breakIt := map[string]func(*core.AuthzRequest){
		"user id 0":            func(r *core.AuthzRequest) { r.UserID = 0 },
		"user id negative":     func(r *core.AuthzRequest) { r.UserID = -1 },
		"issue author id 0":    func(r *core.AuthzRequest) { r.IssueAuthorID = 0 },
		"issue author unknown": func(r *core.AuthzRequest) { r.IssueAuthorID = -5 },
	}
	for _, op := range ops {
		for name, mod := range breakIt {
			for kind, base := range map[string]func(core.Op) core.AuthzRequest{"own": ownIssue, "foreign": foreignIssue} {
				r := base(op)
				mod(&r)
				if got := all.Authorize(r); got != core.DenyIdentityUnknown {
					t.Errorf("%s %s, %s -> %s, want identity-unknown", kind, name, op, got)
				}
			}
		}
	}
	r := foreignIssue(core.OpCommentEdit)
	r.CommentAuthorID = 0
	if got := all.Authorize(r); got != core.DenyIdentityUnknown {
		t.Errorf("comment edit with an unknown comment author -> %s, want identity-unknown", got)
	}
}

// The decision reads numeric ids, never a display name or login: the request type
// has no name field at all, so two people sharing a name but not an id can never
// be taken for one another.
func TestAuthz_SameNameDifferentID_DecidedByIDOnly(t *testing.T) {
	typ := reflect.TypeOf(core.AuthzRequest{})
	for i := range typ.NumField() {
		f := typ.Field(i)
		low := strings.ToLower(f.Name)
		if strings.Contains(low, "name") && f.Name != "RequestedRepo" || strings.Contains(low, "login") {
			t.Errorf("AuthzRequest has a name-like field %q: the decision must not read names", f.Name)
		}
		if f.Name == "UserID" || f.Name == "IssueAuthorID" || f.Name == "CommentAuthorID" {
			if f.Type.Kind() != reflect.Int64 {
				t.Errorf("%s is %s, want a numeric id", f.Name, f.Type)
			}
		}
	}
	c := collab(nil)
	if got := c.Authorize(core.AuthzRequest{Op: core.OpEditBody, RequestedRepo: repoA, ActualRepo: repoA, UserID: me, IssueAuthorID: me + 1}); got != core.DenyNotIssueAuthor {
		t.Errorf("a different id (whatever its display name) -> %s, want not-issue-author", got)
	}
}

// A repository that moved or was renamed is never trusted: a redirect, or an
// answer for a different repository than the one asked about, is refused.
func TestAuthz_RedirectOrRename_IsRefused(t *testing.T) {
	c := collab(func(c *core.Collaboration) {
		c.Enabled, c.AllowedRepositories = true, []string{repoA}
		c.EditBody = true
	})
	redirected := foreignIssue(core.OpEditBody)
	redirected.Redirected = true
	renamed := foreignIssue(core.OpEditBody)
	renamed.ActualRepo = repoB
	missing := foreignIssue(core.OpEditBody)
	missing.ActualRepo = ""
	// Behavior change (lane-18 F20): the host not answering for the repository that was asked about is its own
	// verdict, repository-unverified, and it is decided before any allowance, on one's own Issue as well. An
	// answered repository that is simply not on the allow list stays repository-not-allowed (below).
	for name, r := range map[string]core.AuthzRequest{"redirect": redirected, "rename": renamed, "no answer": missing} {
		if got := c.Authorize(r); got != core.DenyRepoUnverified {
			t.Errorf("%s -> %s, want repository-unverified", name, got)
		}
	}
	unlisted := foreignIssue(core.OpEditBody)
	unlisted.RequestedRepo, unlisted.ActualRepo = repoB, repoB // answered correctly, but not on the list
	if got := c.Authorize(unlisted); got != core.DenyRepoNotAllowed {
		t.Errorf("an answered repository that is not allowed -> %s, want repository-not-allowed", got)
	}
}

// Repository names compare per provider: GitHub ignores case, GitLab and any
// unknown provider compare exactly (the narrower reading).
func TestAuthz_RepositoryComparison_PerProvider(t *testing.T) {
	cases := []struct {
		provider string
		listed   string
		answered string
		want     core.AuthzVerdict
	}{
		{"github", "org/repo", "ORG/REPO", core.Allow},
		{"gitlab", "org/repo", "ORG/REPO", core.DenyRepoNotAllowed},
		{"gitlab", "group/sub/repo", "group/sub/repo", core.Allow},
		{"", "org/repo", "ORG/REPO", core.DenyRepoNotAllowed},
		{"somewhere-else", "org/repo", "ORG/REPO", core.DenyRepoNotAllowed},
	}
	for _, tc := range cases {
		c := core.Collaboration{Provider: tc.provider, Enabled: true, AllowedRepositories: []string{tc.listed}, EditBody: true}
		r := foreignIssue(core.OpEditBody)
		r.RequestedRepo, r.ActualRepo = tc.answered, tc.answered
		if got := c.Authorize(r); got != tc.want {
			t.Errorf("provider %q, listed %q, answered %q -> %s, want %s", tc.provider, tc.listed, tc.answered, got, tc.want)
		}
	}
}

// An unknown operation is never allowed.
func TestAuthz_UnknownOperation_IsRefused(t *testing.T) {
	all := collab(func(c *core.Collaboration) { c.Enabled, c.AllowedRepositories = true, []string{repoA} })
	r := ownIssue("delete-repository")
	if got := all.Authorize(r); got == core.Allow {
		t.Error("an unknown operation was allowed")
	}
}

// Every denial maps to a recorded block reason and a sentinel error; Allow has neither.
func TestAuthz_DenialsCarryReasonAndSentinel(t *testing.T) {
	sentinels := map[core.AuthzVerdict]error{
		core.DenyNotIssueAuthor:       core.ErrNotIssueAuthor,
		core.DenyIdentityUnknown:      core.ErrIdentityUnknown,
		core.DenyRepoNotAllowed:       core.ErrRepoNotAllowed,
		core.DenyPermissionNotGranted: core.ErrPermissionNotGranted,
		core.DenyOwnerUnconfirmed:     core.ErrOwnerUnconfirmed,
	}
	for v, sentinel := range sentinels {
		if _, ok := v.Reason(); !ok {
			t.Errorf("%s has no block reason", v)
		}
		if err := v.Err(); !errors.Is(err, sentinel) {
			t.Errorf("%s.Err() = %v, want it to wrap %v", v, err, sentinel)
		}
	}
	if _, ok := core.Allow.Reason(); ok {
		t.Error("Allow has a block reason")
	}
	if core.Allow.Err() != nil {
		t.Error("Allow has an error")
	}
}
