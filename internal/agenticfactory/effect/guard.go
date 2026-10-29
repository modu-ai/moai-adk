package effect

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// The weakening-operation guard (REQ-AF-019): the single place where a request that effect code is about
// to send is checked. Everything that writes to a code host passes its method and path through
// CheckRequest first; a write that is not on the allowlist is refused before it is sent, and the intent
// is parked with a reason. The decision uses the HTTP method and the path segments only — never the
// wording of a command, a body or an error message.

// ErrRefused is matched (errors.Is) by every refusal.
var ErrRefused = errors.New("effect: request refused by the protection guard")

// RefusalReason says why a request was refused. These are local codes; if core wants a shared type for
// them, lane 18 adds it.
type RefusalReason string

const (
	// ReasonWeakensProtection: a write to branch protection, rulesets, required checks, permissions,
	// webhooks, keys or similar settings.
	ReasonWeakensProtection RefusalReason = "weakens-protection"
	// ReasonNotAllowlisted: any other write that is not one of the few the factory is meant to send.
	ReasonNotAllowlisted RefusalReason = "not-allowlisted"
	// ReasonNotIssueAuthor: a write that only the author of the Issue (or of the comment) may make, aimed
	// at someone else's. Used by the identity policy (see identity.go).
	ReasonNotIssueAuthor RefusalReason = "not-issue-author"
	// ReasonRepoNotAllowed, ReasonPermissionNotGranted and ReasonOwnerUnconfirmed are the other denials of
	// core's authorization function, carried through under the same strings core records.
	ReasonRepoNotAllowed       RefusalReason = "repository-not-allowed"
	ReasonPermissionNotGranted RefusalReason = "permission-not-granted"
	ReasonOwnerUnconfirmed     RefusalReason = "owner-unconfirmed"
	// ReasonIdentityUnknown: the authenticated user or the author could not be established (lookup failed,
	// field missing, or no policy wired), so the write is refused rather than guessed.
	ReasonIdentityUnknown RefusalReason = "identity-unknown"
	// ReasonMalformed: the method or path is not in a clean form (traversal, double slash, encoded
	// separators, an absolute URL, control characters).
	ReasonMalformed RefusalReason = "malformed-request"
)

// RefusedError is a refusal with its reason.
type RefusedError struct {
	Reason RefusalReason
	Method string
	Path   string
}

func (e *RefusedError) Error() string {
	return fmt.Sprintf("effect: refused %s %s (%s)", e.Method, e.Path, e.Reason)
}

func (e *RefusedError) Is(target error) bool { return target == ErrRefused }

// rule is one permitted write: a method and a path pattern. "*" matches exactly one segment.
type rule struct {
	method string
	segs   []string
}

func r(method, pattern string) rule {
	return rule{method: method, segs: strings.Split(strings.Trim(pattern, "/"), "/")}
}

// allowedWrites are the only writes the factory sends: create an Issue, post and edit an Issue comment,
// create and edit a label. Editing a comment or an Issue body is NOT here: it is permitted only through
// Enforcer, after the author is checked (identity.go). Nothing else is permitted, and nothing is added
// without a decision.
//
// Path meanings were compared with the GitHub REST documentation on 2026-10-29 (page summaries read with
// WebFetch, no remote call to GitHub itself):
//
//	POST  /repos/{owner}/{repo}/issues                          https://docs.github.com/en/rest/issues/issues
//	POST  /repos/{owner}/{repo}/issues/{issue_number}/comments  https://docs.github.com/en/rest/issues/comments
//	PATCH /repos/{owner}/{repo}/issues/comments/{comment_id}    https://docs.github.com/en/rest/issues/comments
//	POST  /repos/{owner}/{repo}/labels                          https://docs.github.com/en/rest/issues/labels
//	PATCH /repos/{owner}/{repo}/labels/{name}                   https://docs.github.com/en/rest/issues/labels
//
// The same label page also lists DELETE /repos/{owner}/{repo}/labels/{name}, which is deliberately absent
// here. Not compared: the read route the comment finder uses (GET .../issues/{n}/comments), which this
// guard lets through as a read anyway.
var allowedWrites = []rule{
	r("POST", "/repos/*/*/issues"),
	r("POST", "/repos/*/*/labels"),
	r("PATCH", "/repos/*/*/labels/*"),
}

// protectedSegments name the settings areas a write must never touch. They are compared as whole path
// segments, so a label or an Issue that merely contains the word is not affected.
var protectedSegments = map[string]bool{
	"protection": true, "branch_protection": true, "protected_branches": true, "protected_tags": true,
	"rulesets": true, "rules": true, "required_status_checks": true, "required_pull_request_reviews": true,
	"enforce_admins": true, "restrictions": true, "approval_rules": true, "push_rule": true, "push_rules": true,
	"hooks": true, "webhooks": true, "collaborators": true, "invitations": true, "members": true,
	"permissions": true, "teams": true, "keys": true, "deploy_keys": true, "secrets": true, "variables": true,
	"environments": true, "installations": true, "codeowners": true, "visibility": true, "transfer": true,
}

var overrideHeaders = map[string]bool{
	"x-http-method-override": true, "x-method-override": true, "x-http-method": true,
}

func (r rule) matches(method string, segs []string) bool {
	if r.method != method || len(r.segs) != len(segs) {
		return false
	}
	for i, want := range r.segs {
		if want != "*" && want != segs[i] {
			return false
		}
	}
	return true
}

// CheckRequest returns nil when the request may be sent and a *RefusedError (matching ErrRefused)
// otherwise. Reads (GET, HEAD) are always allowed; every other method needs an allowlist match.
func CheckRequest(method, rawPath string, header map[string]string) error {
	refuse := func(reason RefusalReason) error { return &RefusedError{Reason: reason, Method: method, Path: rawPath} }

	if method != strings.ToUpper(method) || method == "" {
		return refuse(ReasonMalformed)
	}
	for k := range header {
		if overrideHeaders[strings.ToLower(k)] {
			return refuse(ReasonMalformed)
		}
	}
	if method == "GET" || method == "HEAD" {
		return nil
	}

	segs, ok := cleanSegments(rawPath)
	if !ok {
		return refuse(ReasonMalformed)
	}
	lowered := make([]string, len(segs))
	for i, s := range segs {
		lowered[i] = strings.ToLower(s)
	}
	for _, ru := range allowedWrites {
		if ru.matches(method, lowered) {
			return nil
		}
	}
	if _, conditional := classifyConditional(method, lowered, segs); conditional {
		return refuse(ReasonIdentityUnknown) // allowed only through Enforcer.Check, which establishes who is writing
	}
	for _, s := range lowered {
		for _, part := range strings.Split(s, "/") {
			if protectedSegments[part] {
				return refuse(ReasonWeakensProtection)
			}
		}
	}
	return refuse(ReasonNotAllowlisted)
}

// cleanSegments parses a request path into decoded segments, or reports that it is not a clean, rooted
// path: no scheme or host, no control characters, no dot segments, no empty segments, no trailing
// slash. Each segment is decoded on its own, so an encoded slash stays inside its segment (label names
// such as "kind/bug" are one segment).
func cleanSegments(rawPath string) ([]string, bool) {
	if rawPath == "" || !strings.HasPrefix(rawPath, "/") || strings.ContainsAny(rawPath, "\x00\r\n\t\\") {
		return nil, false
	}
	u, err := url.Parse(rawPath)
	if err != nil || u.Scheme != "" || u.Host != "" {
		return nil, false
	}
	var segs []string
	for _, raw := range strings.Split(strings.TrimPrefix(u.EscapedPath(), "/"), "/") {
		seg, err := url.PathUnescape(raw)
		if err != nil || seg == "" || strings.ContainsAny(seg, "\x00\r\n\t\\") {
			return nil, false
		}
		for _, part := range strings.Split(seg, "/") {
			if part == "." || part == ".." {
				return nil, false
			}
		}
		segs = append(segs, seg)
	}
	return segs, true
}
