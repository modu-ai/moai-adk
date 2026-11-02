package review

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

type scripted struct {
	err   error
	calls int
}

func (s *scripted) Review(context.Context, Request) (Result, error) {
	s.calls++
	return Result{HeadSHA: headA, Verdict: VerdictPass}, s.err
}

// REQ-AF-014: auth failure, permission denial, and 429 are recorded as a
// block on the failed call and never retried or bypassed.
func TestCallOnce_BlocksWithoutRetry(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want BlockReason
	}{
		{"401", &StatusError{Code: 401}, BlockAuthFailed},
		{"403", &StatusError{Code: 403}, BlockPermissionDenied},
		{"429", &StatusError{Code: 429}, BlockRateLimited},
		{"wrapped 429", fmt.Errorf("codex: %w", &StatusError{Code: 429}), BlockRateLimited},
		{"not configured", ErrNotConfigured, BlockNotConfigured},
		{"binary absent", ErrBinaryAbsent, BlockNotConfigured},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &scripted{err: tc.err}
			_, reason, err := CallOnce(context.Background(), p, Request{HeadSHA: headA})
			if reason != tc.want || err == nil {
				t.Fatalf("reason=%q err=%v, want %q with error", reason, err, tc.want)
			}
			if p.calls != 1 {
				t.Fatalf("provider called %d times, want exactly 1", p.calls)
			}
		})
	}
}

func TestCallOnce_OtherErrorsAndSuccess(t *testing.T) {
	p := &scripted{err: errors.New("boom")}
	if _, reason, err := CallOnce(context.Background(), p, Request{}); reason != BlockNone || err == nil {
		t.Fatalf("plain error: reason=%q err=%v", reason, err)
	}
	ok := &scripted{}
	res, reason, err := CallOnce(context.Background(), ok, Request{HeadSHA: headA})
	if err != nil || reason != BlockNone || res.Verdict != VerdictPass {
		t.Fatalf("success: res=%+v reason=%q err=%v", res, reason, err)
	}
}

// The adapter layer must not depend on the retired Codex surfaces or any
// GLM package, directly or transitively.
func TestImportGraph_NoCodexSurfaceNoGLM(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}", ".").Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	for pkg := range strings.FieldsSeq(string(out)) {
		low := strings.ToLower(pkg)
		if strings.Contains(low, "codexadapter") || strings.Contains(low, "codexwiring") ||
			strings.Contains(low, "codextools") || strings.Contains(low, "glm") {
			t.Errorf("forbidden dependency: %s", pkg)
		}
	}
}
