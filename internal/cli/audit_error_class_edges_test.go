package cli

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M4 S1 — the classifier's precedence and the
// per-origin mappings behind AC-AEC-002a (REQ-AEC-006; card t2017).

import (
	"net/http"
	"testing"
)

func TestClassifyUnrun_Precedence(t *testing.T) {
	t.Run("a leg with a verdict has no class", func(t *testing.T) {
		for _, v := range []string{"pass", "fail"} {
			if got := classifyUnrun(ReviewOutput{Verdict: v, Summary: "z.ai returned HTTP 429", ErrorClass: errClassQuota}); got != "" {
				t.Errorf("a %s leg classified as %q", v, got)
			}
		}
	})
	t.Run("the origin's class beats the message", func(t *testing.T) {
		out := ReviewOutput{Verdict: VerdictInconclusive, Summary: "request failed: connection refused", ErrorClass: errClassTimeout}
		if got := classifyUnrun(out); got != errClassTimeout {
			t.Errorf("got %q, want the origin's %q", got, errClassTimeout)
		}
	})
	t.Run("an origin class outside the closed set is ignored", func(t *testing.T) {
		out := ReviewOutput{Verdict: VerdictInconclusive, Summary: "z.ai returned HTTP 429", ErrorClass: "lunch"}
		if got := classifyUnrun(out); got != errClassQuota {
			t.Errorf("got %q, want the message's %q", got, errClassQuota)
		}
	})
	t.Run("a Claude error code beats the message", func(t *testing.T) {
		out := claudeInconclusive(claudeErrAuditTimeout, "Claude audit timed out during authentication", "", "")
		out.ErrorClass = "" // as if the origin had set none
		if got := classifyUnrun(out); got != errClassTimeout {
			t.Errorf("got %q, want timeout from the code, not auth from the word authentication", got)
		}
	})
}

func TestClaudeErrorClass_EveryCodeMapsToTheClosedSet(t *testing.T) {
	want := map[string]string{
		claudeErrAuditTimeout:            errClassTimeout,
		claudeErrAuthStatusFailed:        errClassAuth,
		claudeErrSubscriptionUnavailable: errClassAuth,
		claudeErrProviderMismatch:        errClassAuth,
		claudeErrCapacityUnavailable:     errClassQuota,
		claudeErrRateLimited:             errClassQuota,
		claudeErrBinaryMissing:           errClassUnavailable,
		claudeErrModelUnavailable:        errClassUnavailable,
		claudeErrProtocol:                errClassOther,
		claudeErrOutputMalformed:         errClassOther,
		claudeErrOutputTruncated:         errClassOther,
		claudeErrAuditCancelled:          errClassOther,
		claudeErrDiffUnavailable:         errClassOther,
		"CLAUDE_SOMETHING_NEW":           errClassOther,
	}
	for code, class := range want {
		if got := claudeErrorClass(code); got != class {
			t.Errorf("claudeErrorClass(%s) = %q, want %q", code, got, class)
		}
		if !validErrorClass(claudeErrorClass(code)) {
			t.Errorf("claudeErrorClass(%s) is outside the closed set", code)
		}
	}
}

func TestGLMStatusClass(t *testing.T) {
	for status, want := range map[int]string{
		http.StatusTooManyRequests:     errClassQuota,
		http.StatusUnauthorized:        errClassAuth,
		http.StatusForbidden:           errClassAuth,
		http.StatusInternalServerError: errClassUnavailable,
		http.StatusServiceUnavailable:  errClassUnavailable,
		http.StatusBadRequest:          errClassOther,
		http.StatusNotFound:            errClassOther,
	} {
		if got := glmStatusClass(status); got != want {
			t.Errorf("glmStatusClass(%d) = %q, want %q", status, got, want)
		}
	}
}

func TestClassifyMessage_KeywordFamilies(t *testing.T) {
	for msg, want := range map[string]string{
		"codex leg timed out after 5m0s":               errClassTimeout,
		"Claude audit timed out during authentication": errClassTimeout, // a deadline outranks the word authentication
		"rate limit reached":                           errClassQuota,
		"Too Many Requests":                            errClassQuota,
		"401 Unauthorized":                             errClassAuth,
		"not logged in":                                errClassAuth,
		"GLM API key not configured":                   errClassUnavailable,
		"dial tcp: connection refused":                 errClassUnavailable,
		"z.ai returned HTTP 503":                       errClassUnavailable,
		"":                                             errClassOther,
		"something nobody anticipated":                 errClassOther,
	} {
		if got := classifyMessage(msg); got != want {
			t.Errorf("classifyMessage(%q) = %q, want %q", msg, got, want)
		}
	}
}
