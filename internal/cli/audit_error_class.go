package cli

// audit_error_class.go — the error class of an unrun audit leg
// (SPEC-AUDIT-ENGINE-CONVERGENCE-001 REQ-AEC-006). A leg that returns inconclusive
// because the review could not be performed names WHY from a closed set, so a
// reader can tell a quota wall from a deadline from a refused login without
// parsing prose. The class never changes the leg's verdict: it stays
// inconclusive, and a leg that reported pass or fail carries no class.
//
// Where the backend knows the cause structurally (an HTTP status, a context
// deadline, a Claude error code) it sets ReviewOutput.ErrorClass at the origin.
// The rest is classified here from the message the backend produced; a message
// that matches nothing is "other", never a guess at a more specific class.

import "strings"

// The closed error-class set (REQ-AEC-006).
const (
	errClassQuota       = "quota"
	errClassTimeout     = "timeout"
	errClassAuth        = "auth"
	errClassUnavailable = "unavailable"
	errClassOther       = "other"
)

// validErrorClass reports whether c is a member of the closed set.
func validErrorClass(c string) bool {
	switch c {
	case errClassQuota, errClassTimeout, errClassAuth, errClassUnavailable, errClassOther:
		return true
	}
	return false
}

// claudeErrorClass maps a Claude audit error code to its class.
func claudeErrorClass(code string) string {
	switch code {
	case claudeErrAuditTimeout:
		return errClassTimeout
	case claudeErrAuthStatusFailed, claudeErrSubscriptionUnavailable, claudeErrProviderMismatch:
		return errClassAuth
	case claudeErrCapacityUnavailable, claudeErrRateLimited:
		return errClassQuota
	case claudeErrBinaryMissing, claudeErrModelUnavailable:
		return errClassUnavailable
	}
	return errClassOther
}

// classifyUnrun returns the error class of a leg's output: "" for a leg that
// produced a verdict, the origin's class when it set one, the Claude code's class
// when there is one, and otherwise the class its message reads as.
//
// @MX:NOTE: [AUTO] the text fallback is a hypothesis, not a measurement: it only orders keywords the backends are known to emit, and anything it cannot place is "other".
func classifyUnrun(out ReviewOutput) string {
	if out.Verdict != VerdictInconclusive {
		return ""
	}
	if validErrorClass(out.ErrorClass) {
		return out.ErrorClass
	}
	if out.Provenance != nil && out.Provenance.ErrorCode != "" {
		return claudeErrorClass(out.Provenance.ErrorCode)
	}
	return classifyMessage(out.Summary + " " + out.StderrTail)
}

// classifyMessage orders the keyword families: a deadline first (a timed-out
// authentication is a timeout), then quota, then auth, then unavailable.
func classifyMessage(msg string) string {
	m := strings.ToLower(msg)
	has := func(words ...string) bool {
		for _, w := range words {
			if strings.Contains(m, w) {
				return true
			}
		}
		return false
	}
	switch {
	case has("timed out", "deadline exceeded", "timeout"):
		return errClassTimeout
	case has("quota", "rate limit", "rate-limit", "usage limit", "too many requests", "http 429", "exhausted"):
		return errClassQuota
	case has("http 401", "http 403", "unauthorized", "forbidden", "not logged in", "login required", "authentication"):
		return errClassAuth
	case has("not found in path", "not configured", "connection refused", "no such host", "http 500", "http 502", "http 503", "http 504"):
		return errClassUnavailable
	}
	return errClassOther
}
