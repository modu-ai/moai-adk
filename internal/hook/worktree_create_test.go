package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorktreeCreateHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewWorktreeCreateHandler()

	if got := h.EventType(); got != EventWorktreeCreate {
		t.Errorf("EventType() = %q, want %q", got, EventWorktreeCreate)
	}
}

// initWorktreeTestRepo creates a temporary git repository with one commit on
// "main" (git worktree add requires an existing HEAD). The returned path has
// symlinks resolved so it matches `git rev-parse --show-toplevel` output on
// macOS (/var -> /private/var).
func initWorktreeTestRepo(t *testing.T) string {
	t.Helper()

	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve symlinks: %v", err)
	}
	runWorktreeTestGit(t, dir, "init", "-b", "main")
	runWorktreeTestGit(t, dir, "config", "user.email", "test@example.com")
	runWorktreeTestGit(t, dir, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runWorktreeTestGit(t, dir, "add", ".")
	runWorktreeTestGit(t, dir, "commit", "-m", "Initial commit")
	return dir
}

// runWorktreeTestGit executes git in dir and fails the test on error.
func runWorktreeTestGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %s: %v", args, dir, string(out), err)
	}
	return strings.TrimSpace(string(out))
}

// TestWorktreeCreateHandler_ActiveCreator (issue #1570) verifies the
// active-creator contract: given the official `name` payload field, the
// handler creates a real git worktree under .claude/worktrees/<name> and
// returns its absolute path on the output for the CLI dispatcher to echo.
// The passthrough-observer implementation this replaces always returned an
// empty path, so Claude Code aborted every isolation: worktree spawn.
func TestWorktreeCreateHandler_ActiveCreator(t *testing.T) {
	repo := initWorktreeTestRepo(t)

	h := NewWorktreeCreateHandler()
	got, err := h.Handle(context.Background(), &HookInput{
		SessionID:     "sess-wt-1",
		AgentName:     "team-backend-dev",
		WorktreeName:  "backend-impl",
		CWD:           repo,
		HookEventName: "WorktreeCreate",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	wantPath := filepath.Join(repo, ".claude", "worktrees", "backend-impl")
	if got.WorktreePath != wantPath {
		t.Errorf("output.WorktreePath = %q, want %q", got.WorktreePath, wantPath)
	}

	info, err := os.Stat(wantPath)
	if err != nil {
		t.Fatalf("worktree directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("worktree path exists but is not a directory")
	}

	// The created directory must be a registered git worktree on the
	// namespaced branch, not a plain directory. git worktree list emits
	// forward-slash paths on every platform (verified on the windows
	// runner: "worktree C:/Users/..."), so the join-produced path is
	// compared in ToSlash form — a raw Contains on filepath.Join output
	// never matches on windows, where Join produces backslashes.
	listOut := runWorktreeTestGit(t, repo, "worktree", "list", "--porcelain")
	if !strings.Contains(listOut, filepath.ToSlash(wantPath)) {
		t.Errorf("git worktree list does not contain %q:\n%s", wantPath, listOut)
	}
	branchOut := runWorktreeTestGit(t, repo, "branch", "--list", "worktree-backend-impl")
	if branchOut == "" {
		t.Error("expected branch worktree-backend-impl to exist")
	}

	// The registry entry must be persisted under the input CWD. The check
	// unmarshals the registry and compares the Path fields — a raw
	// strings.Contains against the file bytes cannot work on windows,
	// where JSON escaping doubles every backslash in the stored path.
	stateFile := filepath.Join(repo, ".moai", "state", "worktrees.json")
	if _, err := os.Stat(stateFile); err != nil {
		t.Fatalf("worktree registry not written: %v", err)
	}
	registryHasPath(t, stateFile, wantPath)
}

// TestWorktreeCreateHandler_ReusesExistingDirectory verifies idempotency: an
// existing directory at the target path is echoed back without a second git
// worktree add (git would refuse the duplicate path).
func TestWorktreeCreateHandler_ReusesExistingDirectory(t *testing.T) {
	repo := initWorktreeTestRepo(t)

	existing := filepath.Join(repo, ".claude", "worktrees", "leftover")
	if err := os.MkdirAll(existing, 0o755); err != nil {
		t.Fatal(err)
	}

	h := NewWorktreeCreateHandler()
	got, err := h.Handle(context.Background(), &HookInput{
		SessionID:    "sess-wt-2",
		WorktreeName: "leftover",
		CWD:          repo,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.WorktreePath != existing {
		t.Errorf("output.WorktreePath = %q, want %q", got.WorktreePath, existing)
	}

	// A second run re-registers idempotently: one registry entry per path,
	// not one per invocation. Counted on the unmarshaled Path fields —
	// JSON escaping makes a byte-level strings.Count platform-dependent
	// (windows backslashes double in the stored JSON).
	if _, err := h.Handle(context.Background(), &HookInput{
		SessionID:    "sess-wt-2b",
		WorktreeName: "leftover",
		CWD:          repo,
	}); err != nil {
		t.Fatalf("second Handle: %v", err)
	}
	stateFile := filepath.Join(repo, ".moai", "state", "worktrees.json")
	if _, err := os.Stat(stateFile); err != nil {
		t.Fatalf("worktree registry not written: %v", err)
	}
	if got := registryCountPath(stateFile, existing); got != 1 {
		t.Errorf("registry references %q %d times, want exactly 1:\n%s", existing, got, loadWorktreeEntries(stateFile))
	}
}

// registryHasPath asserts the worktree registry at stateFile contains an
// entry whose Path equals want exactly.
func registryHasPath(t *testing.T, stateFile, want string) {
	t.Helper()
	for _, e := range loadWorktreeEntries(stateFile) {
		if e.Path == want {
			return
		}
	}
	t.Errorf("worktree registry does not reference %q:\n%v", want, loadWorktreeEntries(stateFile))
}

// registryCountPath returns how many registry entries carry Path == want.
func registryCountPath(stateFile, want string) int {
	n := 0
	for _, e := range loadWorktreeEntries(stateFile) {
		if e.Path == want {
			n++
		}
	}
	return n
}

// TestWorktreeCreateHandler_MissingNameAborts verifies the contract-honest
// failure: the official payload carries the slug in `name`; without it the
// hook cannot create anything and must fail (non-zero exit) rather than
// return empty stdout + exit 0, which aborts the spawn with no diagnostics.
func TestWorktreeCreateHandler_MissingNameAborts(t *testing.T) {
	h := NewWorktreeCreateHandler()

	_, err := h.Handle(context.Background(), &HookInput{
		SessionID: "sess-wt-3",
	})
	if err == nil {
		t.Fatal("expected error for missing name field, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should name the missing field, got %q", err.Error())
	}
}

// TestWorktreeCreateHandler_NotARepositoryAborts verifies the failure path in
// a directory outside any git repository: creation is impossible, so the
// handler errors (the CLI exits non-zero per "Other exit codes - worktree
// creation failed") instead of printing an empty success.
func TestWorktreeCreateHandler_NotARepositoryAborts(t *testing.T) {
	dir := t.TempDir()

	h := NewWorktreeCreateHandler()
	_, err := h.Handle(context.Background(), &HookInput{
		SessionID:    "sess-wt-4",
		WorktreeName: "probe",
		CWD:          dir,
	})
	if err == nil {
		t.Fatal("expected error outside a git repository, got nil")
	}
}

// TestSanitizeWorktreeBranchSuffix verifies slug-to-branch conversion for the
// nested-segment names Claude Code's name validation permits.
func TestSanitizeWorktreeBranchSuffix(t *testing.T) {
	t.Parallel()

	tests := []struct{ in, want string }{
		{"backend-impl", "backend-impl"},
		{"feat/auth", "feat-auth"},
		{"a/b/c", "a-b-c"},
	}
	for _, tt := range tests {
		if got := sanitizeWorktreeBranchSuffix(tt.in); got != tt.want {
			t.Errorf("sanitizeWorktreeBranchSuffix(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
