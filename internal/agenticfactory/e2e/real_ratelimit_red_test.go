//go:build afred

package e2e

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// RED on purpose (Issue #1948: "limit-style 403 recovers; a permission 403 is
// never retried"). Observed today: a 403 whose text says the API (or a
// secondary) rate limit was exceeded is classified permission-denied, the same
// as a real permission refusal, so a recoverable limit is reported as a refusal
// that needs an operator. The texts are test inputs, not real GitHub values.
// Expected: rate-limited, which recovers after the stated wait.
// Run with: go test -tags afred ./internal/agenticfactory/e2e/ -run RED_
func TestRED_RateLimit_LimitStyle403_IsRateLimited(t *testing.T) {
	for _, text := range []string{
		"gh: HTTP 403: API rate limit exceeded for user (Retry-After: 60)",
		"gh: HTTP 403: You have exceeded a secondary rate limit",
	} {
		x := failing(text)
		if got := blockedReason(t, writeOnce(x)); got != core.ReasonRateLimited {
			t.Errorf("%q -> %s, want rate-limited", text, got)
		}
		if len(x.calls) != 1 {
			t.Errorf("%q: gh calls = %d, want 1 (no retry loop)", text, len(x.calls))
		}
	}
}
