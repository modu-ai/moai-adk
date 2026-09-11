package hook

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Card t560 — the exception in the sweep, pinned so it survives the sweep.
//
// The card scrubbed the git repository-location variables from every child that
// is handed an explicit directory, because GIT_DIR outranks `git -C` and
// cmd.Dir and would otherwise silently redirect the child to the caller's
// repository (GH #1691).
//
// worktreeBaseBranchInPrimaryCheckoutReal is the opposite case, and applying the
// same scrub to it would BE the defect rather than fix one. The function has no
// directory to be confined to: its whole purpose is to ask the ambient git
// context "am I in the primary checkout or inside a linked worktree?", using the
// discriminant that --git-dir and --git-common-dir are equal in the primary
// checkout and differ inside a worktree. Remove GIT_DIR from its environment and
// it stops answering that question and starts answering about whatever directory
// the process happens to be running in.
//
// The sibling functions in that file (symbolic-ref, remote set-head, show-ref)
// take no directory either and resolve through the same ambient context.
//
// This test therefore asserts the ABSENCE of a change, which is a shape that
// passes for free unless it is shown to fail. It was verified by mutation:
// adding a scrub to that function turns this test red (recorded in
// .moai/reports/t560/).

// ambientRepoWithWorktree builds a repository with one linked worktree and
// returns (repoPath, sharedGitDir, worktreeGitDir).
func ambientRepoWithWorktree(t *testing.T) (string, string, string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}

	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		// Scrubbed on purpose: this plumbing must build the fixture the test
		// names, not the repository an ambient GIT_DIR points at.
		cmd.Env = scrubbedEnvForFixture()
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
		}
	}

	run(repo, "init", "-q")
	run(repo, "config", "user.name", "t560")
	run(repo, "config", "user.email", "t560@test.invalid")
	if err := os.WriteFile(filepath.Join(repo, "seed.txt"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	run(repo, "add", "seed.txt")
	run(repo, "commit", "-q", "-m", "seed")
	run(repo, "worktree", "add", "-q", filepath.Join(base, "wt"), "-b", "wtbr")

	shared := filepath.Join(repo, ".git")
	worktreeGitDir := filepath.Join(shared, "worktrees", "wt")
	if _, err := os.Stat(worktreeGitDir); err != nil {
		t.Fatalf("fixture broken: per-worktree gitdir %s: %v", worktreeGitDir, err)
	}
	return repo, shared, worktreeGitDir
}

// scrubbedEnvForFixture drops every GIT_* variable so fixture plumbing is not
// itself steered by the variables the tests set.
func scrubbedEnvForFixture() []string {
	out := make([]string, 0, len(os.Environ()))
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(name, "GIT_") {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// The primary-checkout discriminant must keep reading the ambient git context.
//
// Both cases are asserted together and they must DIFFER. That is what makes the
// test able to fail: a scrubbed implementation resolves from the process's
// working directory in both cases and therefore returns the same answer twice,
// whichever answer that is. An assertion on only one case would keep passing
// whenever the scrubbed answer happened to match it.
func TestWorktreeBaseBranchInPrimaryCheckout_ReadsAmbientGitDir(t *testing.T) {
	_, sharedGitDir, worktreeGitDir := ambientRepoWithWorktree(t)

	t.Setenv("GIT_DIR", sharedGitDir)
	fromShared := worktreeBaseBranchInPrimaryCheckoutReal()

	t.Setenv("GIT_DIR", worktreeGitDir)
	fromWorktree := worktreeBaseBranchInPrimaryCheckoutReal()

	if fromShared == fromWorktree {
		t.Fatalf("the discriminant stopped reading the ambient git context: "+
			"GIT_DIR=%s and GIT_DIR=%s both reported primary=%v.\n"+
			"This function must NOT have its git environment scrubbed — it has no directory "+
			"to be confined to, and the ambient GIT_DIR is the fact it exists to read.",
			sharedGitDir, worktreeGitDir, fromShared)
	}
	if !fromShared {
		t.Errorf("GIT_DIR=%s (shared gitdir) should report the primary checkout, got false", sharedGitDir)
	}
	if fromWorktree {
		t.Errorf("GIT_DIR=%s (per-worktree gitdir) should report NOT primary, got true", worktreeGitDir)
	}
}
