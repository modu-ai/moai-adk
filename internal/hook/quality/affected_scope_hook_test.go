//go:build !windows

package quality

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// affected_scope_hook_test.go — AC-PAS-009: the changed set the gate uses equals
// what the commit records, measured through a REAL pre-commit hook for each
// commit shape, in a plain repository and in a linked worktree.
//
// The hook re-executes this test binary as a helper (the pattern helperStep
// uses). TestMain scrubs the repository-location variables from every process
// of this package, the helper child included, so the hook hands the index
// location git gave it to the helper through a second variable, and the helper
// puts it back before calling the production adapter. That restores exactly the
// environment the production `moai gate` process sees under a hook.

const (
	scopeHelperOutEnv   = "MOAI_SCOPE_HELPER_OUT"
	scopeHelperIndexEnv = "MOAI_SCOPE_HELPER_INDEX"
)

// TestScopeHookHelper is the hook-side process: a no-op in a normal run.
func TestScopeHookHelper(t *testing.T) {
	out := os.Getenv(scopeHelperOutEnv)
	if out == "" {
		return
	}
	if idx := os.Getenv(scopeHelperIndexEnv); idx != "" {
		t.Setenv("GIT_INDEX_FILE", idx)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	body := ""
	ch, err := queryCommitChanges(scopeCtx(t), cwd)
	if err != nil {
		body = "ERROR: " + err.Error()
	} else {
		body = strings.Join(ch.Files, "\n")
	}
	if err := os.WriteFile(out, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", out, err)
	}
}

// installScopeHook installs a pre-commit hook that records the helper's view.
func installScopeHook(t *testing.T, repo, out string) {
	t.Helper()
	script := "#!/bin/sh\n" +
		scopeHelperOutEnv + "='" + out + "' " + scopeHelperIndexEnv + "=\"$GIT_INDEX_FILE\" " +
		"'" + os.Args[0] + "' -test.run='^TestScopeHookHelper$' >/dev/null 2>&1\n" +
		"exit 0\n"
	hook := filepath.Join(repo, ".git", "hooks", "pre-commit")
	if err := os.MkdirAll(filepath.Dir(hook), 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}
	if err := os.WriteFile(hook, []byte(script), 0o755); err != nil {
		t.Fatalf("write hook: %v", err)
	}
}

// recordedSet reads the helper's output as a sorted, space-joined set.
func recordedSet(t *testing.T, out string) string {
	t.Helper()
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("the hook recorded nothing: %v", err)
	}
	if err := os.Remove(out); err != nil {
		t.Fatalf("remove %s: %v", out, err)
	}
	if strings.HasPrefix(string(data), "ERROR:") {
		return string(data) // an adapter error, kept readable rather than sorted
	}
	lines := strings.Fields(strings.ReplaceAll(string(data), "\n", " "))
	slices.Sort(lines)
	return strings.Join(lines, " ")
}

// runCommitShapes drives the five commit shapes of AC-PAS-009 inside dir and
// asserts the set the adapter recorded under each.
func runCommitShapes(t *testing.T, dir, out string) {
	t.Helper()
	write := func(rel, body string) {
		t.Helper()
		writeScopeFiles(t, dir, map[string]string{rel: body})
	}
	expect := func(shape, want string) {
		t.Helper()
		if got := recordedSet(t, out); got != want {
			t.Errorf("%s: the gate saw %q, the commit records %q", shape, got, want)
		}
	}

	write("a/x.go", "2\n")
	write("d/z.go", "2\n")
	gitIn(t, dir, "add", "a/x.go")
	gitIn(t, dir, "commit", "-q", "-m", "i")
	expect("(i) git commit (only a/x.go staged)", "a/x.go")

	write("a/x.go", "3\n")
	write("d/z.go", "3\n")
	gitIn(t, dir, "add", "a/x.go")
	gitIn(t, dir, "commit", "-q", "-a", "-m", "ii")
	expect("(ii) git commit -a", "a/x.go d/z.go")

	write("a/x.go", "4\n")
	write("d/z.go", "4\n")
	gitIn(t, dir, "add", "a/x.go")
	gitIn(t, dir, "commit", "-q", "-m", "iii", "d/z.go")
	expect("(iii) git commit d/z.go", "d/z.go")

	gitIn(t, dir, "reset", "-q")
	if err := os.MkdirAll(filepath.Join(dir, "new"), 0o755); err != nil {
		t.Fatalf("mkdir new: %v", err)
	}
	gitIn(t, dir, "mv", "old/o.go", "new/o.go")
	gitIn(t, dir, "commit", "-q", "-m", "iv")
	expect("(iv) staged rename old/o.go -> new/o.go", "new/o.go old/o.go")

	write("f.go", "5\n")
	gitIn(t, dir, "add", "f.go")
	gitIn(t, dir, "commit", "-q", "--amend", "-m", "v")
	expect("(v) git commit --amend after staging f.go", "f.go")
}

// AC-PAS-009 (release-blocking): the only green path for the changed-set
// criterion; probe009.sh stays the immutable witness of the pre-change helper.
func TestAffectedScope_ChangedSetMatchesCommit(t *testing.T) {
	seed := map[string]string{
		"a/x.go":   "1\n",
		"d/z.go":   "1\n",
		"old/o.go": "1\n",
		"f.go":     "1\n",
	}

	t.Run("plain repository", func(t *testing.T) {
		repo := newScopeRepo(t, seed)
		out := filepath.Join(t.TempDir(), "recorded.txt")
		installScopeHook(t, repo, out)
		runCommitShapes(t, repo, out)
	})

	t.Run("linked worktree", func(t *testing.T) {
		repo := newScopeRepo(t, seed)
		out := filepath.Join(t.TempDir(), "recorded.txt")
		installScopeHook(t, repo, out)
		wt := filepath.Join(t.TempDir(), "linked")
		gitIn(t, repo, "worktree", "add", "-q", wt, "-b", "linked-br")
		runCommitShapes(t, wt, out)
	})
}
