package core

import (
	"errors"
	"testing"
)

func collab() Collaboration {
	return Collaboration{Provider: "github", Enabled: true, AllowedRepositories: []string{"acme/widgets"},
		EditBody: true, Close: true, Labels: true, Assignees: true}
}

func req(op Op) AuthzRequest {
	return AuthzRequest{Op: op, RequestedRepo: "acme/widgets", ActualRepo: "acme/widgets",
		UserID: 7, IssueAuthorID: 99, CommentAuthorID: 7, Owner: OwnerNone}
}

func TestParseID(t *testing.T) {
	for in, want := range map[string]int64{"7": 7, "123456789012": 123456789012} {
		if got, ok := ParseID(in); !ok || got != want {
			t.Errorf("%q -> %d %v", in, got, ok)
		}
	}
	for _, bad := range []string{"", "0", "-5", "abc", "7x", " 7", "1.5", "99999999999999999999"} {
		if _, ok := ParseID(bad); ok {
			t.Errorf("%q must be unknown", bad)
		}
	}
}

func TestUnknownIdentityBlocksEverything(t *testing.T) {
	for _, op := range allOps {
		for name, mut := range map[string]func(*AuthzRequest){
			"user zero":     func(r *AuthzRequest) { r.UserID = 0 },
			"user negative": func(r *AuthzRequest) { r.UserID = -1 },
			"author zero":   func(r *AuthzRequest) { r.IssueAuthorID = 0 },
		} {
			r := req(op)
			mut(&r)
			if v := collab().Authorize(r); v != DenyIdentityUnknown {
				t.Errorf("%s / %s: %s", op, name, v)
			}
			if v := (Collaboration{}).Authorize(r); v != DenyIdentityUnknown {
				t.Errorf("%s / %s with defaults: %s", op, name, v)
			}
		}
	}
	r := req(OpCommentEdit)
	r.CommentAuthorID = 0
	if v := collab().Authorize(r); v != DenyIdentityUnknown {
		t.Errorf("a comment edit with an unknown comment author: %s", v)
	}
}

func TestOwnIssueIsAllowedWithoutAnyOption(t *testing.T) {
	// Closing one's own Issue is allowed: the factory is required to close the
	// Issues it owns (E13/E14), so it is not a new permission.
	for _, op := range []Op{OpEditBody, OpCommentCreate, OpClose, OpTakeOver} {
		r := req(op)
		r.IssueAuthorID = r.UserID
		if v := (Collaboration{}).Authorize(r); v != Allow {
			t.Errorf("%s on own issue: %s", op, v)
		}
	}
	// Even on one's own issue a comment edit is only for one's own comment.
	r := req(OpCommentEdit)
	r.IssueAuthorID = r.UserID
	r.CommentAuthorID = 5
	if v := (Collaboration{}).Authorize(r); v != DenyNotIssueAuthor {
		t.Errorf("editing someone else's comment: %s", v)
	}
	r.CommentAuthorID = r.UserID
	if v := (Collaboration{}).Authorize(r); v != Allow {
		t.Errorf("editing one's own comment: %s", v)
	}
}

// By default another author's Issue accepts a comment and nothing else.
func TestForeignIssueByDefaultAcceptsOnlyComments(t *testing.T) {
	def := Collaboration{}
	if v := def.Authorize(req(OpCommentCreate)); v != Allow {
		t.Errorf("comment: %s", v)
	}
	if v := def.Authorize(req(OpCommentEdit)); v != Allow { // CommentAuthorID == UserID
		t.Errorf("own comment on a foreign issue: %s", v)
	}
	r := req(OpCommentEdit)
	r.CommentAuthorID = 99
	if v := def.Authorize(r); v != DenyNotIssueAuthor {
		t.Errorf("someone else's comment: %s", v)
	}
	for _, op := range []Op{OpEditBody, OpLabel, OpAssignee, OpClose, OpTakeOver} {
		if v := def.Authorize(req(op)); v != DenyNotIssueAuthor {
			t.Errorf("%s by default: %s", op, v)
		}
	}
}

// The collaboration option alone enables nothing; each write needs its own permission.
func TestCollaborationAloneEnablesNoWrite(t *testing.T) {
	c := Collaboration{Provider: "github", Enabled: true, AllowedRepositories: []string{"acme/widgets"}}
	for _, op := range []Op{OpEditBody, OpLabel, OpAssignee, OpClose} {
		if v := c.Authorize(req(op)); v != DenyPermissionNotGranted {
			t.Errorf("%s with the option but no permission: %s", op, v)
		}
	}
	for op, flag := range map[Op]func(*Collaboration){
		OpEditBody: func(c *Collaboration) { c.EditBody = true },
		OpLabel:    func(c *Collaboration) { c.Labels = true },
		OpAssignee: func(c *Collaboration) { c.Assignees = true },
		OpClose:    func(c *Collaboration) { c.Close = true },
	} {
		only := c
		flag(&only)
		if v := only.Authorize(req(op)); v != Allow {
			t.Errorf("%s with its own permission: %s", op, v)
		}
		for other := range map[Op]bool{OpEditBody: true, OpLabel: true, OpAssignee: true, OpClose: true} {
			if other != op {
				if v := only.Authorize(req(other)); v != DenyPermissionNotGranted {
					t.Errorf("permission for %s must not enable %s: %s", op, other, v)
				}
			}
		}
	}
}

func TestPermissionsWithoutTheOptionOrTheRepositoryAreRefused(t *testing.T) {
	off := collab()
	off.Enabled = false
	for _, op := range []Op{OpEditBody, OpLabel, OpAssignee, OpClose, OpTakeOver} {
		if v := off.Authorize(req(op)); v != DenyNotIssueAuthor {
			t.Errorf("%s with the option off: %s", op, v)
		}
	}
	empty := collab()
	empty.AllowedRepositories = nil
	other := collab()
	other.AllowedRepositories = []string{"acme/other"}
	for name, c := range map[string]Collaboration{"empty list": empty, "other repository": other} {
		for _, op := range []Op{OpEditBody, OpLabel, OpAssignee, OpClose, OpTakeOver} {
			if v := c.Authorize(req(op)); v != DenyRepoNotAllowed {
				t.Errorf("%s / %s: %s", name, op, v)
			}
		}
	}
}

func TestTakeOverNeedsAConfirmedOwnerState(t *testing.T) {
	c := Collaboration{Provider: "github", Enabled: true, AllowedRepositories: []string{"acme/widgets"}}
	for state, want := range map[OwnerState]AuthzVerdict{
		OwnerNone: Allow, OwnerHandedOver: Allow, OwnerActive: DenyOwnerUnconfirmed, OwnerUnknown: DenyOwnerUnconfirmed, "": DenyOwnerUnconfirmed, "bogus": DenyOwnerUnconfirmed,
	} {
		r := req(OpTakeOver)
		r.Owner = state
		if v := c.Authorize(r); v != want {
			t.Errorf("owner %q: %s, want %s", state, v, want)
		}
	}
	// Taking over needs no write permission: the four writes stay separate inputs.
	if v := c.Authorize(req(OpTakeOver)); v != Allow {
		t.Errorf("take-over without any write permission: %s", v)
	}
}

func TestRepositoryNamesFollowTheProvider(t *testing.T) {
	if NormalizeRepo("github", "Acme/Widgets") != "acme/widgets" {
		t.Error("github names are case-insensitive")
	}
	for _, p := range []string{"gitlab", "", "bitbucket"} {
		if NormalizeRepo(p, "Acme/Widgets") != "Acme/Widgets" {
			t.Errorf("provider %q must compare exactly", p)
		}
	}
	gh := collab()
	r := req(OpClose)
	r.RequestedRepo, r.ActualRepo = "ACME/Widgets", "ACME/Widgets"
	if v := gh.Authorize(r); v != Allow {
		t.Errorf("github, differing case: %s", v)
	}
	gl := collab()
	gl.Provider = "gitlab"
	if v := gl.Authorize(r); v != DenyRepoNotAllowed {
		t.Errorf("gitlab compares exactly: %s", v)
	}
	gl.AllowedRepositories = []string{"ACME/Widgets"}
	if v := gl.Authorize(r); v != Allow {
		t.Errorf("gitlab, exact match: %s", v)
	}
}

// GitHub answers a renamed repository with a redirect. If the repository the
// host answered for is not the one that was asked about, it is not allowed.
func TestARenamedRepositoryIsNotTreatedAsTheAllowedOne(t *testing.T) {
	c := collab()
	r := req(OpClose)
	r.ActualRepo = "acme/renamed"
	if v := c.Authorize(r); v != DenyRepoNotAllowed {
		t.Errorf("redirected to another name: %s", v)
	}
	r = req(OpClose)
	r.ActualRepo = ""
	if v := c.Authorize(r); v != DenyRepoNotAllowed {
		t.Errorf("no answered repository to compare: %s", v)
	}
	r = req(OpClose)
	r.RequestedRepo = ""
	if v := c.Authorize(r); v != DenyRepoNotAllowed {
		t.Errorf("no requested repository: %s", v)
	}
	// The allowed list is matched against the repository the host answered for.
	r = req(OpClose)
	r.RequestedRepo, r.ActualRepo = "acme/other", "acme/other"
	if v := c.Authorize(r); v != DenyRepoNotAllowed {
		t.Errorf("a repository outside the list: %s", v)
	}
}

func TestUnknownOperationIsRefused(t *testing.T) {
	for _, op := range []Op{"", "delete-everything"} {
		r := req(op)
		r.IssueAuthorID = r.UserID
		if v := collab().Authorize(r); v != DenyUnknownOp {
			t.Errorf("%q: %s", op, v)
		}
	}
}

func TestVerdictMapsToReasonAndError(t *testing.T) {
	cases := map[AuthzVerdict]struct {
		reason BlockReason
		err    error
	}{
		DenyNotIssueAuthor:       {ReasonNotIssueAuthor, ErrNotIssueAuthor},
		DenyIdentityUnknown:      {ReasonIdentityUnknown, ErrIdentityUnknown},
		DenyRepoNotAllowed:       {ReasonRepoNotAllowed, ErrRepoNotAllowed},
		DenyPermissionNotGranted: {ReasonPermissionNotGranted, ErrPermissionNotGranted},
		DenyOwnerUnconfirmed:     {ReasonOwnerUnconfirmed, ErrOwnerUnconfirmed},
		DenyUnknownOp:            {ReasonOpNotDefined, ErrOpNotDefined},
	}
	for v, want := range cases {
		reason, ok := v.Reason()
		if !ok || reason != want.reason || !reason.Valid() {
			t.Errorf("%s -> %q %v", v, reason, ok)
		}
		if err := v.Err(); !errors.Is(err, want.err) {
			t.Errorf("%s -> %v", v, err)
		}
	}
	if _, ok := Allow.Reason(); ok {
		t.Error("allow has no reason")
	}
	if Allow.Err() != nil {
		t.Error("allow is not an error")
	}
}

func TestNewBlockReasonsAreInTheClosedSet(t *testing.T) {
	for _, r := range []BlockReason{ReasonNotIssueAuthor, ReasonIdentityUnknown, ReasonRepoNotAllowed, ReasonPermissionNotGranted, ReasonOwnerUnconfirmed, ReasonForeignClosing, ReasonOpNotDefined} {
		if !r.Valid() {
			t.Errorf("%q is not a valid reason", r)
		}
		if err := ValidatePause(StateBlocked, r); err != nil {
			t.Errorf("%q cannot be recorded on blocked: %v", r, err)
		}
	}
}

// A redirect on the repository lookup (GitHub's 301 for a renamed or moved
// repository) is never an allowance, even when the names happen to match.
func TestARedirectedLookupIsNeverAllowed(t *testing.T) {
	for _, op := range []Op{OpEditBody, OpLabel, OpAssignee, OpClose, OpTakeOver} {
		r := req(op)
		r.Redirected = true
		if v := collab().Authorize(r); v != DenyRepoNotAllowed {
			t.Errorf("%s through a redirect: %s", op, v)
		}
	}
	// Comments on another author's Issue need no repository check at all.
	r := req(OpCommentCreate)
	r.Redirected = true
	if v := (Collaboration{}).Authorize(r); v != Allow {
		t.Errorf("a comment does not depend on the allow list: %s", v)
	}
}

// After a rename the old name may stay in the list: the host then answers for the
// new name, which differs from the one asked about, and that is refused.
func TestOldNameStillListedAfterARename(t *testing.T) {
	c := collab()
	c.AllowedRepositories = []string{"acme/widgets"} // the old name
	r := req(OpClose)
	r.RequestedRepo, r.ActualRepo = "acme/widgets", "acme/gadgets"
	if v := c.Authorize(r); v != DenyRepoNotAllowed {
		t.Errorf("old name, new answer: %s", v)
	}
	// Asking for the new name and getting it, but only the old name is listed.
	r.RequestedRepo, r.ActualRepo = "acme/gadgets", "acme/gadgets"
	if v := c.Authorize(r); v != DenyRepoNotAllowed {
		t.Errorf("only the old name is listed: %s", v)
	}
}

// The permission keys (edit_body, close, labels, assignees) are for writes to
// another author's Issue. They open nothing on one's own Issue: labels and
// assignees there have no basis and are refused whatever is configured.
func TestPermissionKeysNeverOpenOwnIssueLabelsOrAssignees(t *testing.T) {
	own := func(op Op) AuthzRequest {
		r := req(op)
		r.IssueAuthorID = r.UserID
		return r
	}
	every := Collaboration{Provider: "github", Enabled: true, AllowedRepositories: []string{"acme/widgets"},
		EditBody: true, Close: true, Labels: true, Assignees: true}
	for name, c := range map[string]Collaboration{"defaults": {}, "option only": {Provider: "github", Enabled: true, AllowedRepositories: []string{"acme/widgets"}},
		"label key": {Labels: true}, "assignee key": {Assignees: true}, "everything on": every} {
		for _, op := range []Op{OpLabel, OpAssignee} {
			if v := c.Authorize(own(op)); v != DenyPermissionNotGranted {
				t.Errorf("%s / %s on own issue: %s", name, op, v)
			}
		}
		// What the Issue's author keeps doing stays allowed in every configuration.
		for _, op := range []Op{OpEditBody, OpCommentCreate, OpClose, OpTakeOver} {
			if v := c.Authorize(own(op)); v != Allow {
				t.Errorf("%s / %s on own issue: %s", name, op, v)
			}
		}
	}
	// An unknown identity still blocks first.
	r := own(OpLabel)
	r.UserID, r.IssueAuthorID = 0, 0
	if v := every.Authorize(r); v != DenyIdentityUnknown {
		t.Errorf("unknown identity: %s", v)
	}
	// The same keys still govern another author's Issue.
	if v := every.Authorize(req(OpLabel)); v != Allow {
		t.Errorf("label on a foreign issue with its key: %s", v)
	}
}

// An operation the policy does not define (title, milestone, reopen ...) has its
// own reason, so a caller can tell it from a refusal of a defined one.
func TestUndefinedOperationHasItsOwnReason(t *testing.T) {
	for _, op := range []Op{"edit-title", "milestone", "reopen", ""} {
		r := req(op)
		r.IssueAuthorID = r.UserID
		v := collab().Authorize(r)
		if v != DenyUnknownOp {
			t.Fatalf("%q: %s", op, v)
		}
		reason, ok := v.Reason()
		if !ok || reason != ReasonOpNotDefined || !reason.Valid() {
			t.Errorf("%q -> %q %v", op, reason, ok)
		}
		if err := v.Err(); !errors.Is(err, ErrOpNotDefined) || errors.Is(err, ErrNotIssueAuthor) {
			t.Errorf("%q err = %v", op, err)
		}
	}
}
