package cli

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// stubGhClient is a test double for GhClient.
type stubGhClient struct {
	// args captures the arguments passed to Run.
	args []string
	// stdin captures the stdin payload passed as a variadic arg.
	stdin string
	// output is returned from Run.
	output []byte
	// err is returned from Run.
	err error
	// onRun is called with the full args slice if set.
	onRun func(args []string, stdin string) ([]byte, error)
}

func (s *stubGhClient) Run(_ context.Context, args ...string) ([]byte, error) {
	s.args = args
	if s.onRun != nil {
		return s.onRun(args, s.stdin)
	}
	return s.output, s.err
}

func (s *stubGhClient) RunWithStdin(_ context.Context, stdin string, args ...string) ([]byte, error) {
	s.args = args
	s.stdin = stdin
	if s.onRun != nil {
		return s.onRun(args, stdin)
	}
	return s.output, s.err
}

// makeTestChecks returns a minimal RequiredChecks for tests.
func makeTestChecks() *config.RequiredChecks {
	return &config.RequiredChecks{
		Version: 1,
		Branches: map[string]config.BranchChecks{
			"main": {
				Contexts: []string{"Lint", "Test (ubuntu-latest)", "CodeQL"},
			},
		},
	}
}

// makeTestChecksWithRelease returns a RequiredChecks including release/* branch.
// Used to test URL-encoding of branch names containing /.
func makeTestChecksWithRelease() *config.RequiredChecks {
	return &config.RequiredChecks{
		Version: 1,
		Branches: map[string]config.BranchChecks{
			"main":      {Contexts: []string{"Lint"}},
			"release/*": {Contexts: []string{"Lint"}},
		},
	}
}

// TestApplyBranchProtection_URLEncodesBranchSlash verifies that branch names
// containing `/` (e.g. release/*) are URL-encoded when constructing the
// gh api path. Without encoding, GitHub interprets `/` as path separator
// and returns 404.
//
// REQ-CIAUT-025/026: branch protection must support release/* pattern.
func TestApplyBranchProtection_URLEncodesBranchSlash(t *testing.T) {
	t.Parallel()

	checks := makeTestChecksWithRelease()
	stub := &stubGhClient{output: []byte("ok")}

	err := ApplyBranchProtection(context.Background(), stub, "owner", "repo", "release/*", checks)
	if err != nil {
		t.Fatalf("ApplyBranchProtection: %v", err)
	}

	// Captured args should contain the encoded path with %2F instead of literal /.
	gotPath := ""
	for _, a := range stub.args {
		if strings.Contains(a, "/branches/") && strings.Contains(a, "/protection") {
			gotPath = a
			break
		}
	}
	if gotPath == "" {
		t.Fatalf("expected gh api args to contain branch protection path, got args=%v", stub.args)
	}
	if !strings.Contains(gotPath, "release%2F*") {
		t.Errorf("path %q must contain encoded branch 'release%%2F*' (URL-encoding REQ-CIAUT-025)", gotPath)
	}
	if strings.Contains(gotPath, "branches/release/*/protection") {
		t.Errorf("path %q must NOT contain raw 'release/*' segment (regression risk)", gotPath)
	}
}

// TestRenderBranchProtectionJSON verifies that the rendered JSON contains
// the expected contexts array and boolean configuration fields.
func TestRenderBranchProtectionJSON(t *testing.T) {
	checks := makeTestChecks()

	got, err := RenderBranchProtectionJSON(checks, "main")
	if err != nil {
		t.Fatalf("RenderBranchProtectionJSON() error: %v", err)
	}

	wantSubstrings := []string{
		`"Lint"`,
		`"Test (ubuntu-latest)"`,
		`"CodeQL"`,
		`"strict": true`,
		`"allow_force_pushes": false`,
		`"allow_deletions": false`,
		`"required_conversation_resolution": true`,
	}
	for _, want := range wantSubstrings {
		if !strings.Contains(got, want) {
			t.Errorf("RenderBranchProtectionJSON() missing %q in output:\n%s", want, got)
		}
	}
}

// TestRenderBranchProtectionJSON_UnknownBranch verifies that requesting a branch
// not present in the checks returns an error.
func TestRenderBranchProtectionJSON_UnknownBranch(t *testing.T) {
	checks := makeTestChecks()

	_, err := RenderBranchProtectionJSON(checks, "nonexistent")
	if err == nil {
		t.Fatal("RenderBranchProtectionJSON() expected error for unknown branch, got nil")
	}
}

// TestPreflightGh_Authed verifies that a functioning gh CLI returns nil.
func TestPreflightGh_Authed(t *testing.T) {
	stub := &stubGhClient{output: []byte("Logged in to github.com\n")}
	err := PreflightGh(context.Background(), stub)
	if err != nil {
		t.Errorf("PreflightGh() expected nil for authenticated gh, got: %v", err)
	}
}

// TestPreflightGh_Missing verifies that a failing gh CLI returns ErrGhUnavailable.
func TestPreflightGh_Missing(t *testing.T) {
	stub := &stubGhClient{err: errors.New("exec: not found")}
	err := PreflightGh(context.Background(), stub)
	if !errors.Is(err, ErrGhUnavailable) {
		t.Errorf("PreflightGh() expected ErrGhUnavailable, got: %v", err)
	}
}

// TestApplyBranchProtection_Success verifies the gh api call shape.
func TestApplyBranchProtection_Success(t *testing.T) {
	checks := makeTestChecks()
	var capturedArgs []string
	var capturedStdin string

	stub := &stubGhClient{
		onRun: func(args []string, stdin string) ([]byte, error) {
			// First call is preflight (auth status), return success.
			if len(args) > 0 && args[0] == "auth" {
				return []byte("ok"), nil
			}
			capturedArgs = args
			capturedStdin = stdin
			return []byte("ok"), nil
		},
	}

	err := ApplyBranchProtection(context.Background(), stub, "owner", "repo", "main", checks)
	if err != nil {
		t.Fatalf("ApplyBranchProtection() unexpected error: %v", err)
	}

	// Verify call shape.
	wantArgs := []string{"api", "-X", "PUT",
		"/repos/owner/repo/branches/main/protection",
		"--input", "-"}
	if len(capturedArgs) != len(wantArgs) {
		t.Errorf("args length: want %d, got %d\nwant: %v\ngot:  %v",
			len(wantArgs), len(capturedArgs), wantArgs, capturedArgs)
	} else {
		for i, want := range wantArgs {
			if capturedArgs[i] != want {
				t.Errorf("args[%d]: want %q, got %q", i, want, capturedArgs[i])
			}
		}
	}

	// Verify stdin contains rendered JSON.
	if !strings.Contains(capturedStdin, `"strict": true`) {
		t.Errorf("stdin should contain rendered JSON, got: %s", capturedStdin)
	}
}

// TestApplyBranchProtection_AuthFailure verifies that gh auth failure returns
// an error containing the manual gh api command for copy-paste.
func TestApplyBranchProtection_AuthFailure(t *testing.T) {
	checks := makeTestChecks()

	stub := &stubGhClient{
		onRun: func(args []string, stdin string) ([]byte, error) {
			return nil, errors.New("gh: not authenticated")
		},
	}

	err := ApplyBranchProtection(context.Background(), stub, "myowner", "myrepo", "main", checks)
	if err == nil {
		t.Fatal("ApplyBranchProtection() expected error on auth failure, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "gh api -X PUT /repos/") {
		t.Errorf("error should contain manual gh command, got: %v", err)
	}
}

// TestYesConfirmer_AlwaysTrue verifies that yesConfirmer returns true without prompting.
func TestYesConfirmer_AlwaysTrue(t *testing.T) {
	c := &yesConfirmer{}
	got, err := c.Confirm("apply branch protection?")
	if err != nil {
		t.Fatalf("yesConfirmer.Confirm() error: %v", err)
	}
	if !got {
		t.Error("yesConfirmer.Confirm() should always return true")
	}
}
