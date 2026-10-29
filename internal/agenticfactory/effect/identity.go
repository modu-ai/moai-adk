package effect

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// The Issue-author policy, enforced at the effect boundary. The decision itself is lane 18's: core's
// Collaboration.Authorize is the only copy of the rule (own Issues, comments on other authors' Issues,
// the collaboration option, the allowed repositories, the per-operation permissions, redirects and
// renames). This file does the part that belongs here: classify a request by method and path segments
// (and, for an Issue PATCH, by the JSON keys of the body), establish the facts the function needs through
// read-only lookups, call it, and turn a denial into a refusal that parks the intent.
//
// Facts are numeric ids and the repository the host really answered for; names are never compared. When
// a fact cannot be established it is passed as unknown (zero), which the function always blocks. When
// nothing is wired (no Enforcer, no Identity) every conditional write is refused as identity-unknown.

// IssueFacts are what a lookup of an Issue established.
type IssueFacts struct {
	AuthorID   int64  // numeric id of the user who opened the Issue; 0 = unknown
	ActualRepo string // repository the host answered for ("owner/name"); "" = unknown
	Redirected bool   // the answer came through a redirect
}

// CommentFacts are what a lookup of a comment established.
type CommentFacts struct {
	AuthorID   int64  // numeric id of the comment's author; 0 = unknown
	Issue      int    // the Issue the comment belongs to; 0 = unknown
	ActualRepo string // repository the host answered for
	Redirected bool
}

// Identity establishes those facts by read-only lookups. Any error means "unknown".
type Identity interface {
	CurrentUserID(ctx context.Context) (int64, error)
	IssueFacts(ctx context.Context, repo string, issue int) (IssueFacts, error)
	CommentFacts(ctx context.Context, repo string, comment int64) (CommentFacts, error)
}

// Enforcer applies the static guard and then the Issue-author policy.
type Enforcer struct {
	Identity Identity
	// Collaboration is the collaboration setting, filled by whoever reads the configuration. Its zero
	// value is the default: nothing beyond your own Issues and comments.
	Collaboration core.Collaboration
}

type routeKind int

const (
	routeCommentCreate routeKind = iota + 1
	routeCommentEdit
	routeIssuePatch
)

type conditionalRoute struct {
	kind  routeKind
	repo  string // as the path wrote it
	issue int
	cid   int64
}

// classifyConditional recognises the routes that are allowed only after an identity check. lower are the
// lower-cased path segments, raw the same segments as written.
func classifyConditional(method string, lower, raw []string) (conditionalRoute, bool) {
	if len(lower) < 5 || lower[0] != "repos" {
		return conditionalRoute{}, false
	}
	repo := raw[1] + "/" + raw[2]
	switch {
	case method == "POST" && len(lower) == 6 && lower[3] == "issues" && lower[5] == "comments":
		n := atoiPositive(lower[4])
		return conditionalRoute{kind: routeCommentCreate, repo: repo, issue: n}, n > 0
	case method == "PATCH" && len(lower) == 6 && lower[3] == "issues" && lower[4] == "comments":
		id := int64(atoiPositive(lower[5]))
		return conditionalRoute{kind: routeCommentEdit, repo: repo, cid: id}, id > 0
	case method == "PATCH" && len(lower) == 5 && lower[3] == "issues":
		n := atoiPositive(lower[4])
		return conditionalRoute{kind: routeIssuePatch, repo: repo, issue: n}, n > 0
	}
	return conditionalRoute{}, false
}

func atoiPositive(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' || n > 1<<40 {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// opsOf maps the JSON keys of an Issue PATCH body to core operations. ok is false when the body is not a
// non-empty JSON object or names a field this policy has no operation for (title, milestone, ...): such a
// PATCH is not something the factory sends, so it is refused rather than guessed.
func opsOf(body []byte) (ops []core.Op, ok bool) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil || len(m) == 0 {
		return nil, false
	}
	seen := map[core.Op]bool{}
	for k, v := range m {
		var op core.Op
		switch strings.ToLower(k) {
		case "body":
			op = core.OpEditBody
		case "state":
			if !strings.Contains(strings.ToLower(string(v)), "closed") {
				return nil, false // reopening and other states are not part of this policy
			}
			op = core.OpClose
		case "labels":
			op = core.OpLabel
		case "assignee", "assignees":
			op = core.OpAssignee
		default:
			return nil, false
		}
		if !seen[op] {
			seen[op] = true
			ops = append(ops, op)
		}
	}
	return ops, true
}

// Check returns nil when the write may be sent. Routes that need no identity go through CheckRequest
// unchanged; the conditional routes (comment create and edit, Issue PATCH) are decided here.
func (e *Enforcer) Check(ctx context.Context, method, rawPath string, header map[string]string, body []byte) error {
	refuse := func(reason RefusalReason) error { return &RefusedError{Reason: reason, Method: method, Path: rawPath} }

	err := CheckRequest(method, rawPath, header)
	if err == nil {
		return nil // a read, or a write on the plain allowlist
	}
	var re *RefusedError
	if !errors.As(err, &re) || re.Reason != ReasonIdentityUnknown {
		return err // refused for a reason other than "needs an identity check"
	}

	raw, ok := cleanSegments(rawPath)
	if !ok {
		return refuse(ReasonMalformed)
	}
	lower := make([]string, len(raw))
	for i, s := range raw {
		lower[i] = strings.ToLower(s)
	}
	route, ok := classifyConditional(method, lower, raw)
	if !ok {
		return refuse(ReasonNotAllowlisted)
	}

	var ops []core.Op
	switch route.kind {
	case routeCommentCreate:
		ops = []core.Op{core.OpCommentCreate}
	case routeCommentEdit:
		ops = []core.Op{core.OpCommentEdit}
	case routeIssuePatch:
		var okBody bool
		if ops, okBody = opsOf(body); !okBody {
			return refuse(ReasonMalformed)
		}
	}
	if e == nil || e.Identity == nil {
		return refuse(ReasonIdentityUnknown) // nothing wired: fail closed
	}

	req := core.AuthzRequest{RequestedRepo: route.repo, Owner: core.OwnerUnknown}
	if id, err := e.Identity.CurrentUserID(ctx); err == nil && id > 0 {
		req.UserID = id
	}
	issue := route.issue
	if route.kind == routeCommentEdit {
		if cf, err := e.Identity.CommentFacts(ctx, route.repo, route.cid); err == nil && cf.AuthorID > 0 && cf.Issue > 0 {
			req.CommentAuthorID, issue = cf.AuthorID, cf.Issue
			req.Redirected = cf.Redirected
			req.ActualRepo = cf.ActualRepo
		}
	}
	if issue > 0 {
		if f, err := e.Identity.IssueFacts(ctx, route.repo, issue); err == nil && f.AuthorID > 0 {
			req.IssueAuthorID = f.AuthorID
			req.Redirected = req.Redirected || f.Redirected
			if f.ActualRepo != "" {
				req.ActualRepo = f.ActualRepo
			}
		}
	}

	for _, op := range ops {
		req.Op = op
		v := e.Collaboration.Authorize(req)
		if v == core.Allow {
			// Belt: an allow with a missing id would be a bug in the function, not a permission.
			if req.UserID <= 0 || req.IssueAuthorID <= 0 || (op == core.OpCommentEdit && req.CommentAuthorID <= 0) {
				return refuse(ReasonIdentityUnknown)
			}
			continue
		}
		reason, has := v.Reason()
		if !has {
			return refuse(ReasonIdentityUnknown)
		}
		return refuse(RefusalReason(reason))
	}
	return nil
}
