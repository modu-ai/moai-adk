package e2e

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mirror"
)

// Rate-limit and refusal handling on the real mirror. A refusal reaches mirror
// the way `gh api -i` delivers it: a status line, headers, a blank line and the
// body on stdout, exit 1. Header values and bodies are TEST INPUTS only; they
// say nothing about a real host's limits.

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

// failingHTTP answers every call with an HTTP failure in the gh -i shape.
func failingHTTP(status int, header map[string]string, body string) *scriptExec {
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var hdr strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&hdr, "%s: %s\r\n", k, header[k])
	}
	out := fmt.Sprintf("HTTP/2.0 %d status\r\n%s\r\n%s", status, hdr.String(), body)
	errb := fmt.Sprintf("gh: request failed (HTTP %d)", status)
	return &scriptExec{reply: func([]string) ([]byte, []byte, int) { return []byte(out), []byte(errb), 1 }}
}

// failingWording answers with wording on stderr only: no status line reached stdout.
func failingWording(stderr string) *scriptExec {
	return &scriptExec{reply: func([]string) ([]byte, []byte, int) { return nil, []byte(stderr), 1 }}
}

func blockedReason(t *testing.T, err error) core.BlockReason {
	t.Helper()
	return blockedErr(t, err).Reason
}

func blockedErr(t *testing.T, err error) *mirror.BlockedError {
	t.Helper()
	var be *mirror.BlockedError
	if !errors.As(err, &be) {
		t.Fatalf("error %v is not a BlockedError", err)
	}
	return be
}

func writeOnce(x *scriptExec) error {
	m := &mirror.Mirror{
		Repo: "o/r", Exec: x, AllowWrite: true,
		Now: func() time.Time { return t0 }, Backoff: testBackoff,
	}
	return m.WriteLedger(context.Background(), ledgerAt(1, core.StateClaimed))
}

// 429 maps to rate-limited, carries the time the host stated, and is not retried.
func TestRateLimit_429_BlockedWithRetryTime_NotRetried(t *testing.T) {
	x := failingHTTP(429, map[string]string{"retry-after": "30"}, "")
	be := blockedErr(t, writeOnce(x))
	if be.Reason != core.ReasonRateLimited || !be.RetryAt.Equal(t0.Add(30*time.Second)) {
		t.Errorf("429 -> %s retryAt=%v, want rate-limited at %v", be.Reason, be.RetryAt, t0.Add(30*time.Second))
	}
	if len(x.calls) != 1 {
		t.Errorf("gh calls after a 429 = %d, want 1 (no retry loop)", len(x.calls))
	}
}

// A limit-style 403 (Retry-After, or exhausted quota with a reset, or the limit
// wording in the body) is rate-limited, not a permission denial.
func TestRateLimit_LimitStyle403_IsRateLimited(t *testing.T) {
	reset := t0.Add(7 * time.Minute)
	cases := map[string]*scriptExec{
		"retry-after":       failingHTTP(403, map[string]string{"retry-after": "60"}, ""),
		"exhausted+reset":   failingHTTP(403, map[string]string{"x-ratelimit-remaining": "0", "x-ratelimit-reset": fmt.Sprint(reset.Unix())}, ""),
		"secondary-wording": failingHTTP(403, nil, `{"message":"You have exceeded a secondary rate limit."}`),
	}
	for name, x := range cases {
		t.Run(name, func(t *testing.T) {
			if got := blockedReason(t, writeOnce(x)); got != core.ReasonRateLimited {
				t.Errorf("%s -> %s, want rate-limited", name, got)
			}
			if len(x.calls) != 1 {
				t.Errorf("gh calls = %d, want 1", len(x.calls))
			}
		})
	}
}

// A permission 403 and a 401 are blocks that are never retried.
func TestRateLimit_PermissionDenied403_And401_NotRetried(t *testing.T) {
	cases := []struct {
		name   string
		x      *scriptExec
		reason core.BlockReason
	}{
		{"403 permission", failingHTTP(403, nil, `{"message":"Resource not accessible by integration"}`), core.ReasonPermissionDenied},
		{"401", failingHTTP(401, nil, `{"message":"Bad credentials"}`), core.ReasonAuthFailed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := blockedReason(t, writeOnce(tc.x)); got != tc.reason {
				t.Errorf("%s -> %s, want %s", tc.name, got, tc.reason)
			}
			if len(tc.x.calls) != 1 {
				t.Errorf("gh calls = %d, want 1", len(tc.x.calls))
			}
		})
	}
}

// Wording alone never decides a block: stderr that shouts "rate limit" without a
// status line on stdout is a plain error, not a BlockedError.
func TestRateLimit_WordingWithoutStatus_IsAPlainError(t *testing.T) {
	err := writeOnce(failingWording("HTTP 429: API rate limit exceeded"))
	var be *mirror.BlockedError
	if err == nil || errors.As(err, &be) {
		t.Errorf("wording without a status decided a block: %v", err)
	}
}

// No refusal ever makes the code switch credentials or accounts: none of the gh
// invocations carries an auth, token, host or login argument.
func TestRateLimit_NoCredentialOrAccountSwitching(t *testing.T) {
	forbidden := []string{"auth", "login", "--with-token", "--hostname", "GH_TOKEN", "GITHUB_TOKEN", "token"}
	for _, x := range []*scriptExec{
		failingHTTP(429, map[string]string{"retry-after": "1"}, ""),
		failingHTTP(403, nil, `{"message":"Resource not accessible by integration"}`),
		failingHTTP(401, nil, ""),
	} {
		_ = writeOnce(x)
		for _, call := range x.calls {
			for _, a := range call {
				for _, f := range forbidden {
					if strings.EqualFold(a, f) {
						t.Errorf("gh call carries %q: %v", f, call)
					}
				}
			}
		}
	}
}
