package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// t1589 — the lock-root spelling check reads the repository named by an
// inherited GIT_DIR instead of the directory it was given. Given a symlink to a
// repository, git answers the real path; under the inherited variable it
// answers the symlink back, and the spelling correction is silently skipped.
func TestCanonicalTreeSpellingIgnoresInheritedGitEnv(t *testing.T) {
	repo := t1589Repo(t, "card tip", "main")
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(repo, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	other := t1589Repo(t, "unrelated tip", "other")
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)

	got := canonicalTreeSpelling(link)
	if got == link || !factorySameTree(got, repo) {
		t.Errorf("canonicalTreeSpelling(link) = %q, want the repository's own spelling of %q (the inherited GIT_DIR left the link unresolved)", got, repo)
	}
}

// t1589 — the gh call of the github-flow delivery edge must not hand an
// inherited GIT_DIR to gh: gh resolves the repository through it.
func TestFactoryGHIgnoresInheritedGitEnv(t *testing.T) {
	t.Setenv("GIT_DIR", "/some/other/repository/.git")
	for _, kv := range factoryGHEnv() {
		if strings.HasPrefix(kv, "GIT_DIR=") {
			t.Errorf("gh environment carries %q; an inherited GIT_DIR would point gh at another repository", kv)
		}
	}
}

// t1589 — the delivery edge's git reads answer about the repository named by an
// inherited GIT_DIR instead of the card worktree passed as -C. factoryGitRead
// and factoryBranchOfWorktree must both read the worktree they were given, even
// when GIT_DIR and GIT_WORK_TREE name another repository.
func TestFactoryDeliveryGitReadsIgnoreInheritedGitEnv(t *testing.T) {
	wt := t1589Repo(t, "card worktree tip", "WT-card-branch")
	other := t1589Repo(t, "unrelated repository tip", "other-branch")
	wantTip := t1589Git(t, wt, "rev-parse", "HEAD")
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)

	if tip := factoryBranchTip(wt, "WT-card-branch"); tip != wantTip {
		t.Errorf("factoryBranchTip() = %q, want the card worktree's tip %q (the inherited GIT_DIR answered for %q)", tip, wantTip, other)
	}

	subject, err := factoryGitRead(wt, "log", "-1", "--format=%s")
	if err != nil {
		t.Fatalf("factoryGitRead(log) error = %v", err)
	}
	if subject != "card worktree tip" {
		t.Errorf("factoryGitRead(log) = %q, want the card worktree's tip %q (the inherited GIT_DIR answered for %q)", subject, "card worktree tip", other)
	}

	if branch := factoryBranchOfWorktree(wt); branch != "WT-card-branch" {
		t.Errorf("factoryBranchOfWorktree() = %q, want the card worktree's branch %q (the inherited GIT_DIR answered for %q)", branch, "WT-card-branch", other)
	}

	if got := factoryWorktreeForBranchIn(wt, "WT-card-branch"); !factorySameTree(got, wt) {
		t.Errorf("factoryWorktreeForBranchIn() = %q, want the card worktree %q (the inherited GIT_DIR listed %q)", got, wt, other)
	}
}

// t1589Repo makes a one-commit repository whose tip subject and checked-out
// branch are the given values, and returns its directory.
func t1589Repo(t *testing.T, subject, branch string) string {
	t.Helper()
	dir := t.TempDir()
	t1589Git(t, dir, "init", "-q", "-b", branch)
	t1589Git(t, dir, "commit", "-q", "--allow-empty", "--no-gpg-sign", "-m", subject)
	return dir
}

// t1589Git runs one git command in dir and returns its trimmed stdout.
func t1589Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir,
		"-c", "user.name=test", "-c", "user.email=test@example.invalid"}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}
