package cli

// session_worktree_unpushed_test.go — card t673: the auto-cleanup dirty guard
// read `git status --porcelain` only, so a worktree holding COMMITTED but
// UNPUSHED work on a clean tree read as not-dirty — exactly the state the
// "an unpushed branch's worktree is the only copy" discipline protects.
//
// These tests drive the REAL cleanup functions against REAL temporary git
// repositories (no git seams swapped, so every git call below actually
// executes), per the card's first-judgment rule: the guard must be OBSERVED
// on a committed-unpushed tree, not only read in the source.
//
// SUPERSEDED COVERAGE NOTE (SPEC-WEB-SETTINGS-SAVE-001 scope ③): the former
// TestCleanupSessionWorktree_PushedBranchStillRemovable control asserted the
// t673-era premise that pushed alone makes a clean tree removable. REQ-WSS-306
// reverses that premise — a push alone does not satisfy the landing
// confirmation (RED observed when the landing guard landed: the test's tree
// was preserved, its assertion expected removal). The scenario coverage moved
// to session_worktree_landing_test.go: the pushed-unmerged preservation is
// TestCleanupSessionWorktree_PushedUnmergedPreserved, the main-only-repo
// fail-open shape is TestCleanupSessionWorktree_MissingIntegrationRefPreserved,
// and the no-freeze control it carried lives on as
// TestCleanupSessionWorktree_MergedIntoDevelopRemoved.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// canonicalTempDir returns t.TempDir() with symlinks resolved. On macOS the
// temp root is reached through /var → /private/var, and git registers
// worktrees under the RESOLVED spelling while `git worktree remove` matches
// the argument literally — a non-canonical spelling makes every removal fail
// with "is not a working tree", which would fake both removals and
// preservations in these tests.
func canonicalTempDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("canonicalize temp dir: %v", err)
	}
	return dir
}

// realWTRepo stages a real git repository in a temp dir with one linked
// worktree on a WT- branch, and optionally a commit made IN the worktree
// (committed, unpushed — the repo has no remote either way). The test is
// left chdir'd INTO the repo: gitWorktreeRemoveReal resolves the repository
// from the process CWD, so a temp-repo worktree is only removable when the
// process cwd is the temp repo (chdirTemp restores the original dir).
func realWTRepo(t *testing.T, commitInWorktree bool, detached bool) (repoDir, wtPath string) {
	t.Helper()

	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return string(out)
	}

	tmp := canonicalTempDir(t)
	repoDir = filepath.Join(tmp, "repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	git("-C", repoDir, "init", "-q", "-b", "main")
	git("-C", repoDir, "config", "user.email", "t@example.com")
	git("-C", repoDir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(repoDir, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatalf("write base: %v", err)
	}
	git("-C", repoDir, "add", ".")
	git("-C", repoDir, "commit", "-qm", "base")

	wtPath = filepath.Join(tmp, "wt")
	if detached {
		git("-C", repoDir, "worktree", "add", "-q", "--detach", wtPath)
	} else {
		git("-C", repoDir, "worktree", "add", "-q", "-b", "WT-unpushed01-fix", wtPath)
	}
	if commitInWorktree {
		if err := os.WriteFile(filepath.Join(wtPath, "card.txt"), []byte("card work\n"), 0o644); err != nil {
			t.Fatalf("write card file: %v", err)
		}
		git("-C", wtPath, "add", ".")
		git("-C", wtPath, "commit", "-qm", "card work: committed, never pushed")
	}
	chdirTemp(t, repoDir)
	return repoDir, wtPath
}

// TestCleanupSessionWorktree_UnpushedCleanTreePreserved is the card's first
// judgment, observed: a session worktree holding ONLY committed unpushed work
// (clean porcelain, no remote) must be preserved on clean exit. Pre-fix this
// test FAILS — the tree is removed with exit-code silence — which is the
// defect.
func TestCleanupSessionWorktree_UnpushedCleanTreePreserved(t *testing.T) {
	_, wtPath := realWTRepo(t, true, false)

	var out bytes.Buffer
	cleanupSessionWorktree(worktreeCfg(true), wtPath, true, &out)

	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("committed-unpushed worktree was REMOVED by session-exit cleanup (the t673 defect, observed); stat: %v\nnotice: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "unpushed") {
		t.Fatalf("preserved notice must name the unpushed cause, got %q", out.String())
	}
}

// TestCleanupSessionWorktree_DetachedHeadPreserved is the fail-closed
// counterpart: a detached-HEAD worktree with a committed tip has NO branch
// name to survive removal — after `git worktree remove` the commit is
// reachable only through reflogs until GC. Preservation is the only safe
// answer.
func TestCleanupSessionWorktree_DetachedHeadPreserved(t *testing.T) {
	_, wtPath := realWTRepo(t, true, true)

	var out bytes.Buffer
	cleanupSessionWorktree(worktreeCfg(true), wtPath, true, &out)

	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("detached-HEAD worktree with committed tip was REMOVED; stat: %v\nnotice: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "unpushed") {
		t.Fatalf("preserved notice must name the unpushed cause (a bare stat is not preservation evidence), got %q", out.String())
	}
}

// TestCleanupSessionWorktree_PushedBranchStillRemovable — REMOVED with the
// SPEC-WEB-SETTINGS-SAVE-001 scope-③ landing guard: see the SUPERSEDED
// COVERAGE NOTE in the file header. The guard landed, its RED was observed
// (pushed-but-unmerged now preserves, per REQ-WSS-306), and the scenario is
// re-covered in session_worktree_landing_test.go under the new contract.

// TestPRMergeCleanup_GhMergedWithUnpushedCommitsPreserved closes the second
// hole the card names: gh answers MERGED, but the branch gained commits after
// the merge. Pre-fix the merged verdict + clean porcelain reached the removal;
// post-fix the unpushed predicate preserves.
func TestPRMergeCleanup_GhMergedWithUnpushedCommitsPreserved(t *testing.T) {
	// Real repo; the gh seams are the ONLY overrides (deterministic MERGED
	// verdict — a real `gh` cannot answer in a temp repo).
	repoDir, wtPath := realWTRepo(t, true, false)
	porcelain, err := exec.Command("git", "-C", repoDir, "worktree", "list", "--porcelain").Output()
	if err != nil {
		t.Fatalf("worktree list: %v", err)
	}
	swapPRMergeSeams(t, prMergeSeams{
		wtList:     func() (string, error) { return string(porcelain), nil },
		ghLookPath: func() bool { return true },
		ghPRState:  func(string) (string, bool) { return "MERGED", true },
	})

	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)

	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("gh-MERGED + unpushed commits: worktree was REMOVED (the t673 hole, observed); stat: %v\nnotice: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "unpushed") {
		t.Fatalf("preserved notice must name the unpushed cause, got %q", out.String())
	}
}
