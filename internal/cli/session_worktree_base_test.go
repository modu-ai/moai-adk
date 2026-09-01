package cli

// session_worktree_base_test.go — SPEC-WORKTREE-BASEREF-001 M3 (card t313).
//
// Consumer 2: `moai cc -w` passes the configured base branch as the
// `git worktree add` base operand. Covers AC-WBR-007 (the created tree is cut
// from the configured base) and AC-WBR-008 (empty / unresolvable degrade to
// today's byte-identical invocation, and the unresolvability decision comes
// from the SHARED predicate rather than a second rule).

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// swapHookResolvableForTest replaces the SHARED resolvability helper that
// consumer 1 owns, and returns a restore func. Used to prove the cli seam
// delegates to that helper rather than carrying a second rule.
func swapHookResolvableForTest(fn func(string) bool) func() {
	orig := hook.WorktreeBaseBranchResolvable
	hook.WorktreeBaseBranchResolvable = fn
	return func() { hook.WorktreeBaseBranchResolvable = orig }
}

// writeWorktreeBaseConfig writes a git-strategy.yaml carrying the given base
// value under projectRoot, so materializeSessionWorktree reads a real config
// file through the real reader rather than a stubbed one.
func writeWorktreeBaseConfig(t *testing.T, projectRoot, value string) {
	t.Helper()
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "git_strategy:\n    mode: manual\n    worktree_base_branch: \"" + value + "\"\n"
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// swapWorktreeBaseResolvable installs a fake shared-resolvability predicate and
// records its arguments. This is the SAME seam mechanism the package already
// uses for sessionWorktreeGitWorktreeAdd.
func swapWorktreeBaseResolvable(t *testing.T, result bool, args *[]string) {
	t.Helper()
	orig := sessionWorktreeBaseResolvable
	sessionWorktreeBaseResolvable = func(branch string) bool {
		*args = append(*args, branch)
		return result
	}
	t.Cleanup(func() { sessionWorktreeBaseResolvable = orig })
}

// --- AC-WBR-008: the no-base invocation is byte-identical to today ---

// TestSessionWorktreeNoBaseOperandArgvIsUnchanged pins the literal argv. An
// extra trailing empty-string operand FAILS: `git worktree add -b b dest ""`
// is not the same command as `git worktree add -b b dest`.
func TestSessionWorktreeNoBaseOperandArgvIsUnchanged(t *testing.T) {
	got := gitWorktreeAddArgs("/tmp/dest", "WT-x", "")
	want := []string{"git", "worktree", "add", "-b", "WT-x", "/tmp/dest"}
	if len(got) != len(want) {
		t.Fatalf("argv = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("argv = %q, want %q", got, want)
		}
	}
}

// TestSessionWorktreeBaseOperandAppendedWhenSet is the positive half: a
// non-empty base is appended as the FINAL operand.
func TestSessionWorktreeBaseOperandAppendedWhenSet(t *testing.T) {
	got := gitWorktreeAddArgs("/tmp/dest", "WT-x", "develop")
	want := []string{"git", "worktree", "add", "-b", "WT-x", "/tmp/dest", "develop"}
	for i := range want {
		if i >= len(got) || got[i] != want[i] {
			t.Fatalf("argv = %q, want %q", got, want)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("argv = %q, want %q", got, want)
	}
}

// TestSessionWorktreeNoBaseWhenConfigEmpty: an empty configured value reaches
// the add seam as an empty base.
func TestSessionWorktreeNoBaseWhenConfigEmpty(t *testing.T) {
	root := t.TempDir()
	writeWorktreeBaseConfig(t, root, "")

	var gotBase string
	var resolvableArgs []string
	swapWorktreeBaseResolvable(t, true, &resolvableArgs)
	swapSessionWorktreeSeams(t, swSeams{
		add:       func(dest, branch, base string) (string, error) { gotBase = base; return dest, nil },
		commonDir: func() (string, error) { return filepath.Join(root, ".git"), nil },
		configSet: func(string, string, string) error { return nil },
	})

	if _, err := materializeSessionWorktree("WT-x", &bytes.Buffer{}); err != nil {
		t.Fatalf("materialize: %v", err)
	}
	if gotBase != "" {
		t.Errorf("empty setting: base operand = %q, want \"\"", gotBase)
	}
	if len(resolvableArgs) != 0 {
		t.Errorf("empty setting must not consult the resolvability predicate, got %v", resolvableArgs)
	}
}

// TestSessionWorktreeUnresolvableBaseFallsBackToNoOperand: an unresolvable
// value degrades to today's invocation rather than to an error. A worktree on
// the wrong base is recoverable; a worktree that failed to materialize blocks
// the lane.
func TestSessionWorktreeUnresolvableBaseFallsBackToNoOperand(t *testing.T) {
	root := t.TempDir()
	writeWorktreeBaseConfig(t, root, "no-such-branch")

	var gotBase string
	var resolvableArgs []string
	swapWorktreeBaseResolvable(t, false, &resolvableArgs)
	swapSessionWorktreeSeams(t, swSeams{
		add:       func(dest, branch, base string) (string, error) { gotBase = base; return dest, nil },
		commonDir: func() (string, error) { return filepath.Join(root, ".git"), nil },
		configSet: func(string, string, string) error { return nil },
	})

	if _, err := materializeSessionWorktree("WT-x", &bytes.Buffer{}); err != nil {
		t.Fatalf("materialize must succeed on an unresolvable base: %v", err)
	}
	if gotBase != "" {
		t.Errorf("unresolvable setting: base operand = %q, want \"\"", gotBase)
	}
}

// TestSessionWorktreeSharedPredicateResolverInvokedOnce is AC-WBR-008's third
// assertion, stated STRUCTURALLY rather than behaviourally: the unresolvability
// determination must come from the same helper consumer 1 uses, not from a
// second rule that happens to agree. A behavioural-equivalence check would pass
// for a divergent second rule, which is the defect this asserts against.
func TestSessionWorktreeSharedPredicateResolverInvokedOnce(t *testing.T) {
	root := t.TempDir()
	writeWorktreeBaseConfig(t, root, "develop")

	var resolvableArgs []string
	swapWorktreeBaseResolvable(t, true, &resolvableArgs)
	swapSessionWorktreeSeams(t, swSeams{
		add:       func(dest, branch, base string) (string, error) { return dest, nil },
		commonDir: func() (string, error) { return filepath.Join(root, ".git"), nil },
		configSet: func(string, string, string) error { return nil },
	})

	if _, err := materializeSessionWorktree("WT-x", &bytes.Buffer{}); err != nil {
		t.Fatalf("materialize: %v", err)
	}
	if len(resolvableArgs) != 1 {
		t.Fatalf("shared resolvability seam invoked %d times, want exactly 1: %v", len(resolvableArgs), resolvableArgs)
	}
	if resolvableArgs[0] != "develop" {
		t.Errorf("shared resolvability seam got %q, want \"develop\"", resolvableArgs[0])
	}
}

// TestSessionWorktreeSharedPredicateDefaultsToConsumerOnesHelper pins that the
// cli seam's DEFAULT is the internal/hook helper — the structural half of
// REQ-WBR-011. A second, independent rule in this package would make this
// delegation unnecessary, which is exactly what must not happen.
func TestSessionWorktreeSharedPredicateDefaultsToConsumerOnesHelper(t *testing.T) {
	// The default seam delegates to hook.WorktreeBaseBranchResolvable at call
	// time; swapping the hook-side variable must therefore be observable here.
	var seen []string
	restore := swapHookResolvableForTest(func(branch string) bool {
		seen = append(seen, branch)
		return true
	})
	t.Cleanup(restore)

	if !sessionWorktreeBaseResolvable("develop") {
		t.Error("cli seam did not return the shared helper's verdict")
	}
	if len(seen) != 1 || seen[0] != "develop" {
		t.Errorf("cli seam did not delegate to hook.WorktreeBaseBranchResolvable; saw %v", seen)
	}
}

// --- AC-WBR-007: the created tree is cut from the configured base ---

// TestSessionWorktreeConfiguredBaseCutsFromThatBranch drives the REAL
// `git worktree add` against a throwaway repository, so the assertion is on git's
// own reflog rather than on a fake.
func TestSessionWorktreeConfiguredBaseCutsFromThatBranch(t *testing.T) {
	repo := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	git("init", "-q", "-b", "main")
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(repo, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("a.txt", "one\n")
	git("add", "a.txt")
	git("commit", "-q", "-m", "first")

	// A second branch carrying its own commit, then back to main. The base we
	// configure is deliberately NOT the checked-out branch: with no base
	// operand, `git worktree add` would cut from main's HEAD.
	git("checkout", "-q", "-b", "develop")
	write("b.txt", "two\n")
	git("add", "b.txt")
	git("commit", "-q", "-m", "second")
	developSHA := git("rev-parse", "HEAD")
	git("checkout", "-q", "main")
	git("update-ref", "refs/remotes/origin/develop", developSHA)

	writeWorktreeBaseConfig(t, repo, "develop")

	// materializeSessionWorktree and the real gitWorktreeAddReal both shell out
	// without an explicit Dir, so the process cwd must be the repository. The
	// test is not parallel for exactly this reason.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	swapSessionWorktreeSeams(t, swSeams{
		commonDir: func() (string, error) { return filepath.Join(repo, ".git"), nil },
		configSet: func(string, string, string) error { return nil },
	})

	wtPath, err := materializeSessionWorktree("WT-base-probe", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("materialize: %v", err)
	}

	// The substantive check: the new tree's HEAD is develop's commit, not main's.
	head := exec.Command("git", "-C", wtPath, "rev-parse", "HEAD")
	headOut, err := head.CombinedOutput()
	if err != nil {
		t.Fatalf("rev-parse in new worktree: %v\n%s", err, headOut)
	}
	if got := strings.TrimSpace(string(headOut)); got != developSHA {
		t.Errorf("new worktree HEAD = %s, want develop's %s (the base operand was not honoured)", got, developSHA)
	}

	// The reflog form the acceptance criterion states by hand.
	reflog := exec.Command("git", "-C", wtPath, "reflog", "show", "WT-base-probe")
	reflogOut, _ := reflog.CombinedOutput()
	if !strings.Contains(string(reflogOut), "Created from") || !strings.Contains(string(reflogOut), "develop") {
		t.Errorf("reflog does not name develop as the base:\n%s", reflogOut)
	}
}
