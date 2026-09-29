package git

// head_branch_test.go — ResolveHeadBranch unit coverage
// (SPEC-MAIN-COMMIT-BAN-001 D3). Real-git fixtures per the package's
// AP-D-006-style discipline: the resolution runs the actual `git branch
// --show-current` against a real repository, never a mocked runner.

import (
	"strings"
	"testing"
)

// TestResolveHeadBranch_OnBranch is the named-branch headline: a real repo
// checked out on main resolves "main".
func TestResolveHeadBranch_OnBranch(t *testing.T) {
	dir := initTestRepo(t) // git init -b main + one commit

	got, err := ResolveHeadBranch(dir)
	if err != nil {
		t.Fatalf("ResolveHeadBranch(%s) err = %v, want nil", dir, err)
	}
	if got != "main" {
		t.Fatalf("ResolveHeadBranch(%s) = %q, want %q", dir, got, "main")
	}
}

// TestResolveHeadBranch_DetachedHead pins the contract its caller relies on:
// a detached HEAD prints empty output and resolves to ("", nil) — a distinct
// outcome from an error, so the protected-commit deny can ALLOW it
// deliberately (REQ-2.3) instead of reading it as uncertainty.
func TestResolveHeadBranch_DetachedHead(t *testing.T) {
	dir := initTestRepo(t)
	runGit(t, dir, "checkout", "--detach", "HEAD")

	got, err := ResolveHeadBranch(dir)
	if err != nil {
		t.Fatalf("ResolveHeadBranch(detached) err = %v, want nil", err)
	}
	if got != "" {
		t.Fatalf("ResolveHeadBranch(detached) = %q, want empty string", got)
	}
}

// TestResolveHeadBranch_NonGitDir is the uncertainty arm: a non-git directory
// returns an error (never a silent empty-string allow signal) so the hook
// caller can fail open with its advisory.
func TestResolveHeadBranch_NonGitDir(t *testing.T) {
	dir := resolveSymlinks(t, t.TempDir())

	got, err := ResolveHeadBranch(dir)
	if err == nil {
		t.Fatalf("ResolveHeadBranch(non-git) = %q, want error", got)
	}
	if !strings.Contains(err.Error(), "git branch --show-current") {
		t.Fatalf("ResolveHeadBranch(non-git) err = %v, want it to name the failing command", err)
	}
}

// TestResolveHeadBranch_EmptyDir rejects the empty query directory before any
// process spawn (the validateDirArg guard ResolveGitDirs shares).
func TestResolveHeadBranch_EmptyDir(t *testing.T) {
	if _, err := ResolveHeadBranch(""); err == nil {
		t.Fatal("ResolveHeadBranch(\"\") err = nil, want non-nil")
	}
}
