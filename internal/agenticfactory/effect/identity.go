package effect

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// The Issue-author policy. Only the author of an Issue may edit its body; on someone else's Issue the
// factory may create comments and edit only comments it wrote. Closing, relabeling, reassigning and taking
// over someone else's Issue are refused (reason not-issue-author), and when the authenticated user or the
// author cannot be established the write is refused and parked (reason identity-unknown).
//
// This file is the enforcement plumbing in the effect layer: it classifies a request by method and path
// segments (and, for an Issue PATCH, by the JSON keys of the body), establishes both numeric IDs through
// read-only lookups, and asks a pure judging function for the verdict. The judging function belongs to
// lane 18's core package; until it is wired, Judge is nil and every conditional write is refused, so a
// missing wire fails closed. Identities are compared as numeric IDs, never as names.

// IssueOp is the kind of write an Issue-author policy is asked about.
type IssueOp string

const (
	OpCommentCreate  IssueOp = "comment-create"  // any Issue
	OpCommentEdit    IssueOp = "comment-edit"    // only a comment we wrote
	OpIssueBodyEdit  IssueOp = "issue-body-edit" // only our own Issue
	OpIssueClose     IssueOp = "issue-close"     // state=closed
	OpIssueLabels    IssueOp = "issue-labels"    // labels on an Issue
	OpIssueAssignees IssueOp = "issue-assignees" // assignee changes (also what a takeover looks like)
	OpIssueOther     IssueOp = "issue-other"     // any other field of an Issue PATCH
)

// ForeignAllows says whether the collaboration settings allow this operation on someone else's Issue in
// this repository: the option is enabled, the repository is on the allowed list, and the permission for
// exactly this operation is on. It is a pure lookup supplied by the settings reader. A nil function, and
// every operation it does not name, means "no": the default is off, and a collaboration option alone
// allows no write. (A comment edit is never widened by it: only a comment we wrote may be edited.)
type ForeignAllows func(repo string, op IssueOp) bool

// Subject carries the two numeric IDs the verdict depends on. An ID is only meaningful when its Known
// flag is true; the judge must treat an unknown one as identity-unknown.
type Subject struct {
	ActorID     int64
	ActorKnown  bool
	AuthorID    int64 // author of the Issue (or of the comment, for OpCommentEdit)
	AuthorKnown bool
	// ForeignAllowed is the answer of the settings lookup for this repository and operation. It is data
	// for the judge, false by default.
	ForeignAllowed bool
}

// Decision is the judge's verdict.
type Decision struct {
	Allow  bool
	Reason RefusalReason // set when Allow is false: ReasonNotIssueAuthor or ReasonIdentityUnknown
}

// Judge is the pure policy function. It must not do I/O.
type Judge func(op IssueOp, s Subject) Decision

// Identity establishes numeric IDs by read-only GETs. Any error means "unknown".
type Identity interface {
	CurrentUserID(ctx context.Context) (int64, error)
	IssueAuthorID(ctx context.Context, repo string, issue int) (int64, error)
	CommentAuthorID(ctx context.Context, repo string, comment int64) (int64, error)
}

// Enforcer applies the static guard and then the Issue-author policy.
type Enforcer struct {
	Identity Identity
	Judge    Judge
	Foreign  ForeignAllows // nil: nothing is allowed on someone else's Issue
}

type routeKind int

const (
	routeCommentEdit routeKind = iota + 1
	routeIssuePatch
)

type conditionalRoute struct {
	kind  routeKind
	repo  string
	issue int
	cid   int64
}

// classifyConditional recognises the routes that are allowed only after an identity check. segs are
// lower-cased path segments.
func classifyConditional(method string, segs []string) (conditionalRoute, bool) {
	if method != "PATCH" || len(segs) < 4 || segs[0] != "repos" {
		return conditionalRoute{}, false
	}
	repo := segs[1] + "/" + segs[2]
	switch {
	case len(segs) == 6 && segs[3] == "issues" && segs[4] == "comments":
		id, err := strconv.ParseInt(segs[5], 10, 64)
		if err != nil || id <= 0 {
			return conditionalRoute{}, false
		}
		return conditionalRoute{kind: routeCommentEdit, repo: repo, cid: id}, true
	case len(segs) == 5 && segs[3] == "issues":
		n, err := strconv.Atoi(segs[4])
		if err != nil || n <= 0 {
			return conditionalRoute{}, false
		}
		return conditionalRoute{kind: routeIssuePatch, repo: repo, issue: n}, true
	}
	return conditionalRoute{}, false
}

// opsOf maps the JSON keys of an Issue PATCH body to the operations they perform. A body that is not a
// JSON object, or an empty one, has no operations and is refused as malformed.
func opsOf(body []byte) ([]IssueOp, bool) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil || len(m) == 0 {
		return nil, false
	}
	seen := map[IssueOp]bool{}
	var ops []IssueOp
	add := func(op IssueOp) {
		if !seen[op] {
			seen[op] = true
			ops = append(ops, op)
		}
	}
	for k, v := range m {
		switch strings.ToLower(k) {
		case "body":
			add(OpIssueBodyEdit)
		case "state":
			if strings.Contains(strings.ToLower(string(v)), "closed") {
				add(OpIssueClose)
			} else {
				add(OpIssueOther) // reopening and other state values are not part of this policy
			}
		case "labels":
			add(OpIssueLabels)
		case "assignee", "assignees":
			add(OpIssueAssignees)
		default:
			add(OpIssueOther)
		}
	}
	return ops, true
}

// Check returns nil when the write may be sent. Routes that need no identity go through CheckRequest
// unchanged; the conditional routes (comment edit, Issue PATCH) are decided here.
func (e *Enforcer) Check(ctx context.Context, method, rawPath string, header map[string]string, body []byte) error {
	refuse := func(reason RefusalReason) error { return &RefusedError{Reason: reason, Method: method, Path: rawPath} }

	if err := CheckRequest(method, rawPath, header); err != nil {
		var re *RefusedError
		if !errors.As(err, &re) || re.Reason != ReasonIdentityUnknown {
			return err // refused for a reason other than "needs an identity check"
		}
	} else {
		return nil // a read, or a write on the plain allowlist
	}

	segs, ok := cleanSegments(rawPath)
	if !ok {
		return refuse(ReasonMalformed)
	}
	for i := range segs {
		segs[i] = strings.ToLower(segs[i])
	}
	route, ok := classifyConditional(method, segs)
	if !ok {
		return refuse(ReasonNotAllowlisted)
	}

	var ops []IssueOp
	switch route.kind {
	case routeCommentEdit:
		ops = []IssueOp{OpCommentEdit}
	case routeIssuePatch:
		var okBody bool
		if ops, okBody = opsOf(body); !okBody {
			return refuse(ReasonMalformed)
		}
	}
	if e == nil || e.Identity == nil || e.Judge == nil {
		return refuse(ReasonIdentityUnknown) // nothing wired: fail closed
	}

	var sub Subject
	if id, err := e.Identity.CurrentUserID(ctx); err == nil && id > 0 {
		sub.ActorID, sub.ActorKnown = id, true
	}
	var aerr error
	var aid int64
	switch route.kind {
	case routeCommentEdit:
		aid, aerr = e.Identity.CommentAuthorID(ctx, route.repo, route.cid)
	case routeIssuePatch:
		aid, aerr = e.Identity.IssueAuthorID(ctx, route.repo, route.issue)
	}
	if aerr == nil && aid > 0 {
		sub.AuthorID, sub.AuthorKnown = aid, true
	}

	for _, op := range ops {
		sub.ForeignAllowed = route.kind == routeIssuePatch && e.Foreign != nil && e.Foreign(route.repo, op)
		d := e.Judge(op, sub)
		if d.Allow {
			// A judge that says "allow" while an ID is unknown is a bug in the judge, not a permission.
			if !sub.ActorKnown || !sub.AuthorKnown {
				return refuse(ReasonIdentityUnknown)
			}
			continue
		}
		reason := d.Reason
		if reason != ReasonNotIssueAuthor && reason != ReasonIdentityUnknown {
			reason = ReasonIdentityUnknown // an unrecognised reason still refuses, and parks as unknown
		}
		return refuse(reason)
	}
	return nil
}
