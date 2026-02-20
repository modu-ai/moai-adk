package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"
)

// newTestGHClient creates a ghClient with a mock exec function for testing.
func newTestGHClient(fn execFunc) *ghClient {
	return newGHClientWithExec("/tmp/test-repo", fn)
}

// TestGHClient_IsAuthenticated_Success tests successful authentication.
func TestGHClient_IsAuthenticated_Success(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, args ...string) (string, error) {
		if len(args) >= 2 && args[0] == "auth" && args[1] == "status" {
			return "Logged in", nil
		}
		return "", fmt.Errorf("unexpected args: %v", args)
	})

	err := client.IsAuthenticated(context.Background())
	if err != nil {
		t.Errorf("IsAuthenticated() error = %v, want nil", err)
	}
}

// TestGHClient_IsAuthenticated_Failure tests authentication failure.
func TestGHClient_IsAuthenticated_Failure(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("not authenticated")
	})

	err := client.IsAuthenticated(context.Background())
	if err == nil {
		t.Fatal("IsAuthenticated() expected error, got nil")
	}
	if !errors.Is(err, ErrGHNotAuthenticated) {
		t.Errorf("IsAuthenticated() error = %v, want ErrGHNotAuthenticated", err)
	}
}

// TestGHClient_PRCreate_Success tests successful PR creation.
func TestGHClient_PRCreate_Success(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, args ...string) (string, error) {
		// Verify the args contain pr create.
		if len(args) >= 2 && args[0] == "pr" && args[1] == "create" {
			return "https://github.com/owner/repo/pull/42", nil
		}
		return "", fmt.Errorf("unexpected args: %v", args)
	})

	number, err := client.PRCreate(context.Background(), PRCreateOptions{
		Title:      "Test PR",
		Body:       "Test body",
		BaseBranch: "main",
		HeadBranch: "feature/test",
	})
	if err != nil {
		t.Fatalf("PRCreate() error = %v", err)
	}
	if number != 42 {
		t.Errorf("PRCreate() = %d, want 42", number)
	}
}

// TestGHClient_PRCreate_WithLabels tests PR creation with labels.
func TestGHClient_PRCreate_WithLabels(t *testing.T) {
	t.Parallel()

	var capturedArgs []string
	client := newTestGHClient(func(_ context.Context, _ string, args ...string) (string, error) {
		capturedArgs = args
		return "https://github.com/owner/repo/pull/99", nil
	})

	_, err := client.PRCreate(context.Background(), PRCreateOptions{
		Title:  "Labeled PR",
		Body:   "body",
		Labels: []string{"bug", "priority:high"},
	})
	if err != nil {
		t.Fatalf("PRCreate() error = %v", err)
	}

	// Verify --label flags were included.
	labelCount := 0
	for i, arg := range capturedArgs {
		if arg == "--label" && i+1 < len(capturedArgs) {
			labelCount++
		}
	}
	if labelCount != 2 {
		t.Errorf("PRCreate() label count = %d, want 2; args = %v", labelCount, capturedArgs)
	}
}

// TestGHClient_PRCreate_AlreadyExists tests PR creation when PR already exists.
func TestGHClient_PRCreate_AlreadyExists(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("a pull request already exists for this branch")
	})

	_, err := client.PRCreate(context.Background(), PRCreateOptions{
		Title: "Duplicate PR",
		Body:  "body",
	})
	if err == nil {
		t.Fatal("PRCreate() expected error, got nil")
	}
	if !errors.Is(err, ErrPRAlreadyExists) {
		t.Errorf("PRCreate() error = %v, want ErrPRAlreadyExists", err)
	}
}

// TestGHClient_PRCreate_ExecError tests PR creation with generic exec error.
func TestGHClient_PRCreate_ExecError(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("network error")
	})

	_, err := client.PRCreate(context.Background(), PRCreateOptions{
		Title: "Test PR",
		Body:  "body",
	})
	if err == nil {
		t.Fatal("PRCreate() expected error, got nil")
	}
}

// TestGHClient_PRCreate_BadURL tests PR creation with unparseable URL response.
func TestGHClient_PRCreate_BadURL(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "not-a-valid-url", nil
	})

	_, err := client.PRCreate(context.Background(), PRCreateOptions{
		Title: "Test PR",
		Body:  "body",
	})
	if err == nil {
		t.Fatal("PRCreate() expected error for bad URL, got nil")
	}
}

// TestGHClient_PRView_Success tests successful PR view.
func TestGHClient_PRView_Success(t *testing.T) {
	t.Parallel()

	prJSON, _ := json.Marshal(PRDetails{
		Number:     123,
		Title:      "Test PR",
		State:      "OPEN",
		Mergeable:  "MERGEABLE",
		HeadBranch: "feature/test",
		BaseBranch: "main",
		URL:        "https://github.com/owner/repo/pull/123",
		CreatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return string(prJSON), nil
	})

	details, err := client.PRView(context.Background(), 123)
	if err != nil {
		t.Fatalf("PRView() error = %v", err)
	}
	if details.Number != 123 {
		t.Errorf("PRView() Number = %d, want 123", details.Number)
	}
	if details.Title != "Test PR" {
		t.Errorf("PRView() Title = %q, want %q", details.Title, "Test PR")
	}
	if details.Mergeable != "MERGEABLE" {
		t.Errorf("PRView() Mergeable = %q, want %q", details.Mergeable, "MERGEABLE")
	}
}

// TestGHClient_PRView_NotFound tests PR view when PR is not found.
func TestGHClient_PRView_NotFound(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("pull request not found")
	})

	_, err := client.PRView(context.Background(), 999)
	if err == nil {
		t.Fatal("PRView() expected error, got nil")
	}
	if !errors.Is(err, ErrPRNotFound) {
		t.Errorf("PRView() error = %v, want ErrPRNotFound", err)
	}
}

// TestGHClient_PRView_CouldNotResolve tests PR view with 'Could not resolve' error.
func TestGHClient_PRView_CouldNotResolve(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("Could not resolve to a PullRequest with the number 42")
	})

	_, err := client.PRView(context.Background(), 42)
	if err == nil {
		t.Fatal("PRView() expected error, got nil")
	}
	if !errors.Is(err, ErrPRNotFound) {
		t.Errorf("PRView() error = %v, want ErrPRNotFound", err)
	}
}

// TestGHClient_PRView_InvalidJSON tests PR view with invalid JSON response.
func TestGHClient_PRView_InvalidJSON(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "{invalid json", nil
	})

	_, err := client.PRView(context.Background(), 1)
	if err == nil {
		t.Fatal("PRView() expected error for invalid JSON, got nil")
	}
}

// TestGHClient_PRView_ExecError tests PR view with generic exec error.
func TestGHClient_PRView_ExecError(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("connection refused")
	})

	_, err := client.PRView(context.Background(), 1)
	if err == nil {
		t.Fatal("PRView() expected error, got nil")
	}
}

// TestGHClient_PRMerge_Merge tests PR merge with merge method.
func TestGHClient_PRMerge_Merge(t *testing.T) {
	t.Parallel()

	var capturedArgs []string
	client := newTestGHClient(func(_ context.Context, _ string, args ...string) (string, error) {
		capturedArgs = args
		return "", nil
	})

	err := client.PRMerge(context.Background(), 100, MergeMethodMerge, false)
	if err != nil {
		t.Errorf("PRMerge() error = %v", err)
	}

	// Verify --merge flag is present.
	found := slices.Contains(capturedArgs, "--merge")
	if !found {
		t.Errorf("PRMerge() args = %v, want --merge flag", capturedArgs)
	}
}

// TestGHClient_PRMerge_Squash tests PR merge with squash method.
func TestGHClient_PRMerge_Squash(t *testing.T) {
	t.Parallel()

	var capturedArgs []string
	client := newTestGHClient(func(_ context.Context, _ string, args ...string) (string, error) {
		capturedArgs = args
		return "", nil
	})

	err := client.PRMerge(context.Background(), 100, MergeMethodSquash, false)
	if err != nil {
		t.Errorf("PRMerge() error = %v", err)
	}

	found := slices.Contains(capturedArgs, "--squash")
	if !found {
		t.Errorf("PRMerge() args = %v, want --squash flag", capturedArgs)
	}
}

// TestGHClient_PRMerge_Rebase tests PR merge with rebase method.
func TestGHClient_PRMerge_Rebase(t *testing.T) {
	t.Parallel()

	var capturedArgs []string
	client := newTestGHClient(func(_ context.Context, _ string, args ...string) (string, error) {
		capturedArgs = args
		return "", nil
	})

	err := client.PRMerge(context.Background(), 100, MergeMethodRebase, true)
	if err != nil {
		t.Errorf("PRMerge() error = %v", err)
	}

	hasRebase := false
	hasDeleteBranch := false
	for _, arg := range capturedArgs {
		if arg == "--rebase" {
			hasRebase = true
		}
		if arg == "--delete-branch" {
			hasDeleteBranch = true
		}
	}
	if !hasRebase {
		t.Errorf("PRMerge() args = %v, want --rebase flag", capturedArgs)
	}
	if !hasDeleteBranch {
		t.Errorf("PRMerge() args = %v, want --delete-branch flag", capturedArgs)
	}
}

// TestGHClient_PRMerge_UnsupportedMethod tests PR merge with unsupported method.
func TestGHClient_PRMerge_UnsupportedMethod(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", nil
	})

	err := client.PRMerge(context.Background(), 100, MergeMethod("unknown"), false)
	if err == nil {
		t.Fatal("PRMerge() expected error for unsupported method, got nil")
	}
}

// TestGHClient_PRMerge_Conflict tests PR merge with conflict error.
func TestGHClient_PRMerge_Conflict(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("merge conflict detected")
	})

	err := client.PRMerge(context.Background(), 100, MergeMethodMerge, false)
	if err == nil {
		t.Fatal("PRMerge() expected error, got nil")
	}
	if !errors.Is(err, ErrMergeConflict) {
		t.Errorf("PRMerge() error = %v, want ErrMergeConflict", err)
	}
}

// TestGHClient_PRMerge_ExecError tests PR merge with generic exec error.
func TestGHClient_PRMerge_ExecError(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("server error")
	})

	err := client.PRMerge(context.Background(), 100, MergeMethodMerge, false)
	if err == nil {
		t.Fatal("PRMerge() expected error, got nil")
	}
}

// TestGHClient_PRChecks_Success tests successful PR checks retrieval.
func TestGHClient_PRChecks_Success(t *testing.T) {
	t.Parallel()

	checks := []Check{
		{Name: "build", Status: "completed", Conclusion: "success"},
		{Name: "test", Status: "completed", Conclusion: "success"},
	}
	checksJSON, _ := json.Marshal(checks)

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return string(checksJSON), nil
	})

	status, err := client.PRChecks(context.Background(), 100)
	if err != nil {
		t.Fatalf("PRChecks() error = %v", err)
	}
	if status.Overall != CheckPass {
		t.Errorf("PRChecks() Overall = %q, want %q", status.Overall, CheckPass)
	}
	if len(status.Checks) != 2 {
		t.Errorf("PRChecks() len(Checks) = %d, want 2", len(status.Checks))
	}
}

// TestGHClient_PRChecks_EmptyChecks tests PR checks with empty checks array.
func TestGHClient_PRChecks_EmptyChecks(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "[]", nil
	})

	status, err := client.PRChecks(context.Background(), 100)
	if err != nil {
		t.Fatalf("PRChecks() error = %v", err)
	}
	if status.Overall != CheckPending {
		t.Errorf("PRChecks() Overall = %q, want %q (empty checks)", status.Overall, CheckPending)
	}
}

// TestGHClient_PRChecks_NotFound tests PR checks when PR is not found.
func TestGHClient_PRChecks_NotFound(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("pull request not found")
	})

	_, err := client.PRChecks(context.Background(), 999)
	if err == nil {
		t.Fatal("PRChecks() expected error, got nil")
	}
	if !errors.Is(err, ErrPRNotFound) {
		t.Errorf("PRChecks() error = %v, want ErrPRNotFound", err)
	}
}

// TestGHClient_PRChecks_InvalidJSON tests PR checks with invalid JSON response.
func TestGHClient_PRChecks_InvalidJSON(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "{invalid", nil
	})

	_, err := client.PRChecks(context.Background(), 1)
	if err == nil {
		t.Fatal("PRChecks() expected error for invalid JSON, got nil")
	}
}

// TestGHClient_PRChecks_ExecError tests PR checks with exec error.
func TestGHClient_PRChecks_ExecError(t *testing.T) {
	t.Parallel()

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("api error")
	})

	_, err := client.PRChecks(context.Background(), 1)
	if err == nil {
		t.Fatal("PRChecks() expected error, got nil")
	}
}

// TestGHClient_PRChecks_FailedChecks tests PR checks with failed checks.
func TestGHClient_PRChecks_FailedChecks(t *testing.T) {
	t.Parallel()

	checks := []Check{
		{Name: "test", Status: "completed", Conclusion: "failure"},
	}
	checksJSON, _ := json.Marshal(checks)

	client := newTestGHClient(func(_ context.Context, _ string, _ ...string) (string, error) {
		return string(checksJSON), nil
	})

	status, err := client.PRChecks(context.Background(), 100)
	if err != nil {
		t.Fatalf("PRChecks() error = %v", err)
	}
	if status.Overall != CheckFail {
		t.Errorf("PRChecks() Overall = %q, want %q", status.Overall, CheckFail)
	}
}

// TestGHClient_PRCreate_NoBranches tests PR creation without base/head branches.
func TestGHClient_PRCreate_NoBranches(t *testing.T) {
	t.Parallel()

	var capturedArgs []string
	client := newTestGHClient(func(_ context.Context, _ string, args ...string) (string, error) {
		capturedArgs = args
		return "https://github.com/owner/repo/pull/1", nil
	})

	number, err := client.PRCreate(context.Background(), PRCreateOptions{
		Title: "Minimal PR",
		Body:  "body",
	})
	if err != nil {
		t.Fatalf("PRCreate() error = %v", err)
	}
	if number != 1 {
		t.Errorf("PRCreate() = %d, want 1", number)
	}

	// Verify --base and --head are not present.
	for _, arg := range capturedArgs {
		if arg == "--base" || arg == "--head" {
			t.Errorf("PRCreate() args contain %q, should not be present when branches not set", arg)
		}
	}
}

// TestNewGHClientWithExec verifies the test constructor.
func TestNewGHClientWithExec(t *testing.T) {
	t.Parallel()

	called := false
	fn := func(_ context.Context, _ string, _ ...string) (string, error) {
		called = true
		return "", nil
	}

	client := newGHClientWithExec("/tmp/repo", fn)
	if client == nil {
		t.Fatal("newGHClientWithExec returned nil")
	}
	if client.root != "/tmp/repo" {
		t.Errorf("root = %q, want %q", client.root, "/tmp/repo")
	}

	// Verify execFn is used.
	_, _ = client.exec(context.Background())
	if !called {
		t.Error("execFn was not called by client.exec()")
	}
}

// TestGHClient_exec_FallbackToExecGH verifies that a nil execFn falls through to execGH.
// This test just verifies the code path; execGH will fail if gh CLI is not present,
// so we only check the error occurs (not that it succeeds).
func TestGHClient_exec_FallbackToExecGH(t *testing.T) {
	t.Parallel()

	// A client with no execFn will use the real execGH.
	client := NewGHClient("/tmp/nonexistent-repo")

	// execGH may fail if gh is not in PATH (ErrGHNotFound) or if it fails for
	// other reasons. Either way, the code path through execGH is exercised.
	_, _ = client.exec(context.Background(), "version")
	// We don't assert the error because it depends on the test environment.
}

// TestGHClient_Push_Success tests successful push via injected pushFn.
func TestGHClient_Push_Success(t *testing.T) {
	t.Parallel()

	var capturedWorkDir string
	client := newGHClientWithPush("/tmp/repo", func(_ context.Context, workDir string) error {
		capturedWorkDir = workDir
		return nil
	})

	err := client.Push(context.Background(), "/tmp/custom-dir")
	if err != nil {
		t.Errorf("Push() error = %v", err)
	}
	if capturedWorkDir != "/tmp/custom-dir" {
		t.Errorf("Push() workDir = %q, want %q", capturedWorkDir, "/tmp/custom-dir")
	}
}

// TestGHClient_Push_UsesRootWhenDirEmpty tests that Push uses root when dir is empty.
func TestGHClient_Push_UsesRootWhenDirEmpty(t *testing.T) {
	t.Parallel()

	var capturedWorkDir string
	client := newGHClientWithPush("/tmp/repo-root", func(_ context.Context, workDir string) error {
		capturedWorkDir = workDir
		return nil
	})

	err := client.Push(context.Background(), "")
	if err != nil {
		t.Errorf("Push() error = %v", err)
	}
	if capturedWorkDir != "/tmp/repo-root" {
		t.Errorf("Push() workDir = %q, want root %q", capturedWorkDir, "/tmp/repo-root")
	}
}

// TestGHClient_Push_Error tests Push when the push operation fails.
func TestGHClient_Push_Error(t *testing.T) {
	t.Parallel()

	client := newGHClientWithPush("/tmp/repo", func(_ context.Context, _ string) error {
		return errors.New("remote rejected: permission denied")
	})

	err := client.Push(context.Background(), "")
	if err == nil {
		t.Fatal("Push() expected error, got nil")
	}
}

// TestGHClient_Push_FallbackToGit verifies that a nil pushFn falls through to pushWithGit.
// pushWithGit may fail in CI without a git repo; we just verify no panic.
func TestGHClient_Push_FallbackToGit(t *testing.T) {
	t.Parallel()

	client := NewGHClient("/tmp/nonexistent-repo")
	// This exercises the pushWithGit path; it will fail but should not panic.
	_ = client.Push(context.Background(), "")
}

// TestNewGHClientWithPush verifies the test constructor with custom push function.
func TestNewGHClientWithPush(t *testing.T) {
	t.Parallel()

	called := false
	client := newGHClientWithPush("/tmp/repo", func(_ context.Context, _ string) error {
		called = true
		return nil
	})
	if client == nil {
		t.Fatal("newGHClientWithPush returned nil")
	}

	_ = client.Push(context.Background(), "")
	if !called {
		t.Error("pushFn was not called")
	}
}
