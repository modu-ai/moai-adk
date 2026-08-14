// checkout_test.go — git-directory resolution extracted from
// internal/hook/branch_guard.go (SPEC-KANBAN-BOARD-001 REQ-KB-005, taking the
// SPEC-KANBAN-WORKTREE-001 REQ-KW-018 extraction disposition).
//
// The resolution is judged on the path it returns, not on the boolean the
// original caller derived from it: isPrimaryCheckout compares two values for
// equality, and equality is insensitive to an offset shared by both sides,
// while the board takes the PARENT of one of them as its root — a wrong path
// there is silently fatal (AC-KB-002's rationale).
package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// initTestRepoWithWorktree creates a primary repository plus one linked
// worktree, returning (primary, worktree). Both paths are symlink-resolved so
// path comparisons against git output hold on macOS.
func initTestRepoWithWorktree(t *testing.T) (string, string) {
	t.Helper()
	primary := initTestRepo(t)
	worktreeRaw := filepath.Join(primary, "..", "wt-"+filepath.Base(primary))
	runGit(t, primary, "worktree", "add", worktreeRaw, "-b", "topic")
	// Resolve only after creation — EvalSymlinks needs the path to exist.
	return primary, resolveSymlinks(t, worktreeRaw)
}

// TestResolveGitDirs_PrimaryCheckout — from the primary checkout both paths are
// absolute and equal, and the parent of the common dir is the checkout root
// (the board-root rule).
func TestResolveGitDirs_PrimaryCheckout(t *testing.T) {
	t.Parallel()
	primary := initTestRepo(t)

	dirs, err := ResolveGitDirs(primary)
	if err != nil {
		t.Fatalf("ResolveGitDirs(primary) error = %v", err)
	}
	wantGit := filepath.Join(primary, ".git")
	if dirs.GitDir != wantGit {
		t.Errorf("GitDir = %q, want %q", dirs.GitDir, wantGit)
	}
	if dirs.CommonDir != wantGit {
		t.Errorf("CommonDir = %q, want %q (equal in primary)", dirs.CommonDir, wantGit)
	}
	if !filepath.IsAbs(dirs.CommonDir) {
		t.Errorf("CommonDir = %q is not absolute", dirs.CommonDir)
	}
	if got := filepath.Dir(dirs.CommonDir); got != primary {
		t.Errorf("parent of CommonDir = %q, want primary root %q", got, primary)
	}
}

// TestResolveGitDirs_Worktree — from a linked worktree the git dir lives under
// the primary's .git/worktrees/, the common dir is the primary's .git, and the
// parent of the common dir is still the PRIMARY root: the resolution is a
// function of the repository, not of the caller's tree.
func TestResolveGitDirs_Worktree(t *testing.T) {
	t.Parallel()
	primary, worktree := initTestRepoWithWorktree(t)

	dirs, err := ResolveGitDirs(worktree)
	if err != nil {
		t.Fatalf("ResolveGitDirs(worktree) error = %v", err)
	}
	wantCommon := filepath.Join(primary, ".git")
	if dirs.CommonDir != wantCommon {
		t.Errorf("CommonDir = %q, want %q", dirs.CommonDir, wantCommon)
	}
	if dirs.GitDir == dirs.CommonDir {
		t.Errorf("GitDir = CommonDir = %q; in a worktree they must differ", dirs.GitDir)
	}
	if !strings.HasPrefix(dirs.GitDir, wantCommon+string(filepath.Separator)) {
		t.Errorf("GitDir = %q, want under %q", dirs.GitDir, wantCommon)
	}
	if got := filepath.Dir(dirs.CommonDir); got != primary {
		t.Errorf("parent of CommonDir = %q, want primary root %q", got, primary)
	}
}

// TestResolveGitDirs_InsideSubdirectory — resolving from a nested directory of
// either checkout yields the same dirs as from its root, so no caller depends
// on standing in a particular location.
func TestResolveGitDirs_InsideSubdirectory(t *testing.T) {
	t.Parallel()
	_, worktree := initTestRepoWithWorktree(t)

	nested := filepath.Join(worktree, "internal", "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	fromRoot, err := ResolveGitDirs(worktree)
	if err != nil {
		t.Fatalf("ResolveGitDirs(worktree) error = %v", err)
	}
	fromNested, err := ResolveGitDirs(nested)
	if err != nil {
		t.Fatalf("ResolveGitDirs(nested) error = %v", err)
	}
	if fromNested.GitDir != fromRoot.GitDir || fromNested.CommonDir != fromRoot.CommonDir {
		t.Errorf("nested resolution (%q, %q) != root resolution (%q, %q)",
			fromNested.GitDir, fromNested.CommonDir, fromRoot.GitDir, fromRoot.CommonDir)
	}
}

// TestResolveGitDirs_FallbackForcedThroughIndirection — AC-KB-002's second
// half. The primary probe is forced to fail through the package-level
// ExecCommand indirection (simulating an older-git host rejecting
// --path-format=absolute); the dispatcher inside ResolveGitDirs must take the
// --absolute-git-dir + cwd-normalized --git-common-dir fallback and return the
// SAME paths. Non-fallback calls are delegated to the real git binary, so the
// assertion is against the actual repository, not canned output.
//
// Direct invocation of the fallback would be a vacuous pass that bypasses the
// dispatcher — the branch_guard_test.go comment records the same rule.
func TestResolveGitDirs_FallbackForcedThroughIndirection(t *testing.T) {
	if isWindowsRuntime() {
		t.Skip("fallback mock uses sh -c; skip on windows")
	}
	primary := initTestRepo(t)

	orig := ExecCommand
	t.Cleanup(func() { ExecCommand = orig })
	ExecCommand = func(name string, args ...string) *exec.Cmd {
		joined := strings.Join(args, " ")
		if strings.Contains(joined, "--path-format=absolute") {
			// Simulate an older-git host rejecting the flag.
			return exec.Command("sh", "-c", "echo 'unknown option: path-format=absolute' >&2; exit 1")
		}
		return orig(name, args...)
	}

	got, err := ResolveGitDirs(primary)
	if err != nil {
		t.Fatalf("ResolveGitDirs fallback error = %v", err)
	}
	wantGit := filepath.Join(primary, ".git")
	if got.GitDir != wantGit {
		t.Errorf("fallback GitDir = %q, want %q", got.GitDir, wantGit)
	}
	if got.CommonDir != wantGit {
		t.Errorf("fallback CommonDir = %q, want %q", got.CommonDir, wantGit)
	}
	if root := filepath.Dir(got.CommonDir); root != primary {
		t.Errorf("fallback board root = %q, want %q", root, primary)
	}
}

// TestResolveGitDirs_FallbackNormalizesRelativeCommonDir — the fallback's bare
// --git-common-dir returns a repo-relative path in the primary checkout; the
// extracted resolver must normalize it against the queried directory. Mocked
// end to end (both probe shapes fail-and-succeed per the older-git contract).
func TestResolveGitDirs_FallbackNormalizesRelativeCommonDir(t *testing.T) {
	if isWindowsRuntime() {
		t.Skip("fallback mock uses sh -c; skip on windows")
	}
	primary := initTestRepo(t)

	orig := ExecCommand
	t.Cleanup(func() { ExecCommand = orig })
	ExecCommand = func(name string, args ...string) *exec.Cmd {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "--path-format=absolute"):
			return exec.Command("sh", "-c", "echo 'unknown option' >&2; exit 1")
		case strings.Contains(joined, "--absolute-git-dir"):
			return exec.Command("sh", "-c", "printf %s "+shellQuotePath(filepath.Join(primary, ".git")))
		case strings.Contains(joined, "--git-common-dir"):
			// Bare form: repo-relative, exactly as a primary checkout yields.
			return exec.Command("sh", "-c", "printf %s .git")
		}
		return orig(name, args...)
	}

	got, err := ResolveGitDirs(primary)
	if err != nil {
		t.Fatalf("ResolveGitDirs relative-common fallback error = %v", err)
	}
	if got.CommonDir != filepath.Join(primary, ".git") {
		t.Errorf("normalized CommonDir = %q, want %q", got.CommonDir, filepath.Join(primary, ".git"))
	}
}

// TestResolveGitDirs_NonGitDir — an unresolvable directory surfaces an error
// rather than a zero-value result, so callers can fail in their own safe
// direction.
func TestResolveGitDirs_NonGitDir(t *testing.T) {
	t.Parallel()
	nonGit := t.TempDir()
	if _, err := ResolveGitDirs(nonGit); err == nil {
		t.Fatal("ResolveGitDirs(non-git) err = nil, want non-nil")
	}
	if err := validateDirArg(""); err == nil {
		t.Fatal("empty dir arg: err = nil, want non-nil")
	}
}

// TestIsPrimaryCheckout_PrimaryAndWorktree — the boolean discriminant the
// original hook caller consumes, preserved through the extraction.
func TestIsPrimaryCheckout_PrimaryAndWorktree(t *testing.T) {
	t.Parallel()
	primary, worktree := initTestRepoWithWorktree(t)

	gotPrimary, err := IsPrimaryCheckout(primary)
	if err != nil {
		t.Fatalf("IsPrimaryCheckout(primary) error = %v", err)
	}
	if !gotPrimary {
		t.Error("IsPrimaryCheckout(primary) = false, want true")
	}

	gotWorktree, err := IsPrimaryCheckout(worktree)
	if err != nil {
		t.Fatalf("IsPrimaryCheckout(worktree) error = %v", err)
	}
	if gotWorktree {
		t.Error("IsPrimaryCheckout(worktree) = true, want false")
	}
}

// shellQuotePath single-quote-escapes s for interpolation into a sh -c arg.
func shellQuotePath(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// isWindowsRuntime reports the test's runtime OS.
func isWindowsRuntime() bool {
	return runtime.GOOS == "windows"
}
