package core

import (
	"fmt"
	"strconv"
	"strings"
)

// This file is the one place that decides who may do what to which Issue
// (Issue #1948). It is a pure function of its inputs: no network call, no
// configuration read. Callers fill the inputs from the host (the authenticated
// user, the Issue's author, the repository the host answered for) and from the
// collaboration settings; a value they cannot establish is passed as unknown,
// and unknown always blocks.
//
// The default is: work on your own Issues only (its body, comments, closing and
// taking it over; never its labels or assignees). On another author's Issue you
// may add comments and edit your own comments, nothing else. The collaboration
// option widens that only for a repository on the allow list, and it opens no
// write by itself: editing the body, closing, labels and assignees are four
// separate permissions, and taking over (claiming) another author's Issue is a
// fifth thing that also needs a confirmed owner state.

// Op is an operation on an Issue.
type Op string

const (
	OpEditBody      Op = "edit-body"
	OpCommentCreate Op = "comment-create"
	OpCommentEdit   Op = "comment-edit"
	OpLabel         Op = "label"
	OpAssignee      Op = "assignee"
	OpClose         Op = "close"
	OpTakeOver      Op = "take-over" // claim an Issue as the executing agent
)

var allOps = []Op{OpEditBody, OpCommentCreate, OpCommentEdit, OpLabel, OpAssignee, OpClose, OpTakeOver}

// OwnerState is what is known about who currently owns the work on an Issue.
// Taking another author's Issue over needs OwnerNone or OwnerHandedOver; an
// active owner, or an owner that could not be established, blocks.
type OwnerState string

const (
	OwnerNone       OwnerState = "none"
	OwnerHandedOver OwnerState = "handed-over"
	OwnerActive     OwnerState = "active"
	OwnerUnknown    OwnerState = "unknown"
)

// Collaboration holds the collaboration settings. The keys that fill it
// (workflow.agentic_factory.collaboration.*) are provisional; the evaluation
// below is the only copy of the rule. Every field defaults to off or empty.
type Collaboration struct {
	// Provider selects how repository names compare: "github" is
	// case-insensitive, anything else (gitlab, unknown, empty) compares exactly.
	Provider            string
	Enabled             bool
	AllowedRepositories []string
	EditBody            bool
	Close               bool
	Labels              bool
	Assignees           bool
}

// AuthzRequest is one operation to decide.
type AuthzRequest struct {
	Op Op
	// RequestedRepo is the repository the caller asked about. ActualRepo is the
	// repository the host answered for (full_name of the response). Redirected is
	// true when the lookup was answered with a redirect (GitHub's 301 for a
	// renamed or moved repository).
	RequestedRepo string
	ActualRepo    string
	Redirected    bool
	// UserID is the authenticated user, IssueAuthorID the Issue's author,
	// CommentAuthorID the comment's author (comment edits only). All are numeric
	// ids; zero or negative means unknown.
	UserID          int64
	IssueAuthorID   int64
	CommentAuthorID int64
	// Owner is the current owner state, used by OpTakeOver on another author's Issue.
	Owner OwnerState
}

// AuthzVerdict is the answer.
type AuthzVerdict string

const (
	Allow                    AuthzVerdict = "allow"
	DenyNotIssueAuthor       AuthzVerdict = "not-issue-author"
	DenyIdentityUnknown      AuthzVerdict = "identity-unknown"
	DenyRepoNotAllowed       AuthzVerdict = "repository-not-allowed"
	DenyPermissionNotGranted AuthzVerdict = "permission-not-granted"
	DenyOwnerUnconfirmed     AuthzVerdict = "owner-unconfirmed"
	DenyUnknownOp            AuthzVerdict = "unknown-operation"
)

// ParseID reads a numeric host id. Empty, non-numeric, zero and negative values
// are unknown.
func ParseID(s string) (int64, bool) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// NormalizeRepo puts a repository path in the form it is compared in. GitHub
// treats owner and repository names as case-insensitive, so only that provider
// is lower-cased; GitLab's path case rules are not established, so it and any
// unknown provider compare exactly (the narrower reading).
func NormalizeRepo(provider, path string) string {
	if provider == "github" {
		return strings.ToLower(path)
	}
	return path
}

// repoMatches reports whether the host answered for the repository that was asked
// about: both names are known, the lookup was not redirected (a moved or renamed
// repository), and the two names are the same. It is the first thing every
// operation needs, on one's own Issue and on another author's alike; it says
// nothing about the allow list.
func (c Collaboration) repoMatches(r AuthzRequest) bool {
	if r.Redirected || r.RequestedRepo == "" || r.ActualRepo == "" {
		return false
	}
	return NormalizeRepo(c.Provider, r.RequestedRepo) == NormalizeRepo(c.Provider, r.ActualRepo)
}

// repoAllowed is repoMatches and, in addition, the repository is on the allow list.
// Only the collaboration operations on another author's Issue ask for the list.
func (c Collaboration) repoAllowed(r AuthzRequest) bool {
	if !c.repoMatches(r) {
		return false
	}
	got := NormalizeRepo(c.Provider, r.ActualRepo)
	for _, a := range c.AllowedRepositories {
		if NormalizeRepo(c.Provider, a) == got {
			return true
		}
	}
	return false
}

// Authorize decides one operation. See the file comment for the rule.
func (c Collaboration) Authorize(r AuthzRequest) AuthzVerdict {
	known := false
	for _, op := range allOps {
		if r.Op == op {
			known = true
		}
	}
	if !known {
		return DenyUnknownOp
	}
	if r.UserID <= 0 || r.IssueAuthorID <= 0 || (r.Op == OpCommentEdit && r.CommentAuthorID <= 0) {
		return DenyIdentityUnknown
	}
	// Before any allowance: the host must have answered for the repository that was
	// asked about. A repository that moved, was renamed or was reached through a
	// redirect gives no allowance on any branch below (audit F20).
	if !c.repoMatches(r) {
		return DenyRepoNotAllowed
	}
	// A comment is edited only by its own author, on any Issue.
	if r.Op == OpCommentEdit && r.CommentAuthorID != r.UserID {
		return DenyNotIssueAuthor
	}
	if r.UserID == r.IssueAuthorID {
		// Your own Issue: its body, comments, closing and taking it over are yours.
		// Labels and assignees have no basis here, and the permission keys do not
		// create one: they are for writes to another author's Issue.
		switch r.Op {
		case OpLabel, OpAssignee:
			return DenyPermissionNotGranted
		}
		return Allow
	}
	// Another author's Issue: comments only, unless collaboration allows more.
	switch r.Op {
	case OpCommentCreate, OpCommentEdit:
		return Allow
	}
	if !c.Enabled {
		return DenyNotIssueAuthor
	}
	if !c.repoAllowed(r) {
		return DenyRepoNotAllowed
	}
	switch r.Op {
	case OpTakeOver:
		if r.Owner == OwnerNone || r.Owner == OwnerHandedOver {
			return Allow
		}
		return DenyOwnerUnconfirmed
	case OpEditBody:
		return flag(c.EditBody)
	case OpClose:
		return flag(c.Close)
	case OpLabel:
		return flag(c.Labels)
	case OpAssignee:
		return flag(c.Assignees)
	}
	return DenyUnknownOp
}

func flag(on bool) AuthzVerdict {
	if on {
		return Allow
	}
	return DenyPermissionNotGranted
}

// Reason is the block reason a denial records. Allow has none.
func (v AuthzVerdict) Reason() (BlockReason, bool) {
	switch v {
	case DenyNotIssueAuthor:
		return ReasonNotIssueAuthor, true
	case DenyUnknownOp:
		return ReasonOpNotDefined, true
	case DenyIdentityUnknown:
		return ReasonIdentityUnknown, true
	case DenyRepoNotAllowed:
		return ReasonRepoNotAllowed, true
	case DenyPermissionNotGranted:
		return ReasonPermissionNotGranted, true
	case DenyOwnerUnconfirmed:
		return ReasonOwnerUnconfirmed, true
	}
	return "", false
}

// Err is nil for Allow and otherwise a sentinel the verdict wraps.
func (v AuthzVerdict) Err() error {
	var base error
	switch v {
	case Allow:
		return nil
	case DenyNotIssueAuthor:
		base = ErrNotIssueAuthor
	case DenyUnknownOp:
		base = ErrOpNotDefined
	case DenyIdentityUnknown:
		base = ErrIdentityUnknown
	case DenyRepoNotAllowed:
		base = ErrRepoNotAllowed
	case DenyPermissionNotGranted:
		base = ErrPermissionNotGranted
	case DenyOwnerUnconfirmed:
		base = ErrOwnerUnconfirmed
	default:
		base = ErrNotIssueAuthor
	}
	return fmt.Errorf("%w: %s", base, v)
}
