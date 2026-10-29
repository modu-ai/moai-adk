package e2e

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mirror"
)

// Rate-limit and refusal handling on the real mirror classifier. The stderr
// texts are test inputs only; they are not claims about real GitHub limits.

type scriptExec struct {
	mu    sync.Mutex
	calls [][]string
	reply func(args []string) (stdout, stderr []byte, exit int)
}

func (s *scriptExec) Run(_ context.Context, args ...string) ([]byte, []byte, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, append([]string(nil), args...))
	out, errb, exit := s.reply(args)
	return out, errb, exit, nil
}

func failing(stderr string) *scriptExec {
	return &scriptExec{reply: func([]string) ([]byte, []byte, int) { return nil, []byte(stderr), 1 }}
}

func blockedReason(t *testing.T, err error) core.BlockReason {
	t.Helper()
	var be *mirror.BlockedError
	if !errors.As(err, &be) {
		t.Fatalf("error %v is not a BlockedError", err)
	}
	return be.Reason
}

func writeOnce(x *scriptExec) error {
	m := &mirror.Mirror{Repo: "o/r", Exec: x, AllowWrite: true}
	return m.WriteLedger(context.Background(), ledgerAt(1, core.StateClaimed))
}

// 429 maps to rate-limited and is not retried by the caller.
func TestRateLimit_429_BlockedAndNotRetried(t *testing.T) {
	x := failing("gh: HTTP 429: too many requests")
	err := writeOnce(x)
	if got := blockedReason(t, err); got != core.ReasonRateLimited {
		t.Errorf("429 -> %s, want rate-limited", got)
	}
	if len(x.calls) != 1 {
		t.Errorf("gh calls after a 429 = %d, want 1 (no retry loop)", len(x.calls))
	}
}

// A permission-denied 403 maps to permission-denied and is never retried.
func TestRateLimit_PermissionDenied403_BlockedAndNotRetried(t *testing.T) {
	x := failing("gh: HTTP 403: Resource not accessible by integration")
	err := writeOnce(x)
	if got := blockedReason(t, err); got != core.ReasonPermissionDenied {
		t.Errorf("403 permission -> %s, want permission-denied", got)
	}
	if len(x.calls) != 1 {
		t.Errorf("gh calls after a permission 403 = %d, want 1", len(x.calls))
	}
}

// 401 is an authentication failure, blocked and not retried.
func TestRateLimit_401_AuthFailedAndNotRetried(t *testing.T) {
	x := failing("gh: HTTP 401: Bad credentials")
	if got := blockedReason(t, writeOnce(x)); got != core.ReasonAuthFailed {
		t.Errorf("401 -> %s, want auth-failed", got)
	}
	if len(x.calls) != 1 {
		t.Errorf("gh calls = %d, want 1", len(x.calls))
	}
}

// No refusal ever makes the code switch credentials or accounts: none of the
// gh invocations carries an auth, token, host or login argument.
func TestRateLimit_NoCredentialOrAccountSwitching(t *testing.T) {
	forbidden := []string{"auth", "login", "--with-token", "--hostname", "GH_TOKEN", "GITHUB_TOKEN", "token"}
	for _, text := range []string{"gh: HTTP 429: x", "gh: HTTP 403: Resource not accessible", "gh: HTTP 401: Bad credentials"} {
		x := failing(text)
		_ = writeOnce(x)
		for _, call := range x.calls {
			for _, a := range call {
				for _, f := range forbidden {
					if strings.EqualFold(a, f) {
						t.Errorf("%q: gh call carries %q: %v", text, f, call)
					}
				}
			}
		}
	}
}

// Observation: a 403 whose text says the API rate limit was exceeded. The
// requirement separates it from a permission 403 (recoverable vs not). What the
// classifier does today is logged, not assumed.
func TestRateLimit_LimitStyle403_Observed(t *testing.T) {
	for _, text := range []string{
		"gh: HTTP 403: API rate limit exceeded for user (Retry-After: 60)",
		"gh: HTTP 403: You have exceeded a secondary rate limit",
	} {
		x := failing(text)
		reason := blockedReason(t, writeOnce(x))
		t.Logf("OBSERVED %q -> %s (calls=%d)", text, reason, len(x.calls))
	}
}
