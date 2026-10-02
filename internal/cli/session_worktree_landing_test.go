package cli

// session_worktree_landing_test.go — SPEC-WEB-SETTINGS-SAVE-001 scope ③
// (REQ-WSS-301..306): the session-exit disposal guards read uncommitted state
// (REQ-SW-010) and pushed-ness (card t673) only, so a clean worktree on a
// branch that IS pushed but NOT merged into the remote integration branch
// read as disposable — the exact state AGENTS.md §3 protects ("dispose of no
// worktree until the branch is integrated and the remote merge has landed").
//
// The decided predicate (decision-index Q1) is fetch-less remote-tracking
// reachability: the branch tip is an ancestor of refs/remotes/origin/develop,
// OR `git cherry refs/remotes/origin/develop <branch>` is empty (patch-id
// equivalence — covers squash merges, the SPEC-WORKTREE-SQUASH-MERGE-001
// lesson). A stale remote-tracking ref can only misjudge toward "not landed"
// → preserve (fail-open, the safe direction).
//
// These tests drive the REAL cleanup function against REAL temporary git
// repositories (no git seams swapped), per the t673 first-judgment rule: the
// guard must be OBSERVED on a pushed-unmerged tree, not only read in the
// source.

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// realLandingRepo stages a real git repository with a bare `origin` remote
// carrying BOTH main and develop (the integration branch the decided
// predicate reads), plus a linked worktree on a WT- branch holding one
// commit PUSHED to origin — the clean, pushed, upstream-current state the
// pre-guard path removed. The test is left chdir'd INTO the repo, per the
// realWTRepo convention (gitWorktreeRemoveReal resolves the repository from
// the process CWD).
func realLandingRepo(t *testing.T, withDevelop bool) (repoDir, wtPath string) {
	t.Helper()

	git := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	tmp := canonicalTempDir(t)
	repoDir = filepath.Join(tmp, "repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	git("init", "-q", "-b", "main", repoDir)
	git("-C", repoDir, "config", "user.email", "t@example.com")
	git("-C", repoDir, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(repoDir, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatalf("write base: %v", err)
	}
	git("-C", repoDir, "add", ".")
	git("-C", repoDir, "commit", "-qm", "base")

	remoteDir := filepath.Join(tmp, "origin.git")
	git("init", "-q", "--bare", remoteDir)
	git("-C", repoDir, "remote", "add", "origin", remoteDir)
	git("-C", repoDir, "push", "-q", "origin", "main")
	if withDevelop {
		git("-C", repoDir, "branch", "-q", "develop", "main")
		git("-C", repoDir, "push", "-q", "origin", "develop")
	}

	wtPath = filepath.Join(tmp, "wt")
	git("-C", repoDir, "worktree", "add", "-q", "-b", "WT-landing03-fix", wtPath)
	if err := os.WriteFile(filepath.Join(wtPath, "card.txt"), []byte("card work\n"), 0o644); err != nil {
		t.Fatalf("write card file: %v", err)
	}
	git("-C", wtPath, "add", ".")
	git("-C", wtPath, "commit", "-qm", "card work: committed and pushed, never merged")
	git("-C", wtPath, "push", "-q", "-u", "origin", "WT-landing03-fix")
	chdirTemp(t, repoDir)
	return repoDir, wtPath
}

// TestCleanupSessionWorktree_PushedUnmergedPreserved is AC-WSS-010's first
// judgment, observed: a clean worktree whose branch is pushed but NOT merged
// into origin/develop must be preserved with a notice naming the unconfirmed
// landing. Pre-guard this test FAILS — the tree is removed — which is the
// defect (REQ-WSS-306: a push alone does not satisfy the landing
// confirmation).
func TestCleanupSessionWorktree_PushedUnmergedPreserved(t *testing.T) {
	_, wtPath := realLandingRepo(t, true)

	var out bytes.Buffer
	cleanupSessionWorktree(worktreeCfg(true), wtPath, true, &out)

	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("pushed-but-unmerged worktree was REMOVED by session-exit cleanup (the scope-③ defect, observed); stat: %v\nnotice: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "unconfirmed") {
		t.Fatalf("preserved notice must name the unconfirmed landing, got %q", out.String())
	}
}

// TestCleanupSessionWorktree_MergedIntoDevelopRemoved is AC-WSS-011's
// common-path control: once the branch tip IS reachable from
// refs/remotes/origin/develop, disposal proceeds exactly as before — the
// cheap shared path the guard must not freeze.
func TestCleanupSessionWorktree_MergedIntoDevelopRemoved(t *testing.T) {
	repoDir, wtPath := realLandingRepo(t, true)
	git := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	// Land the branch: merge it into develop and push develop. The
	// remote-tracking ref is shared through the common git dir, so the
	// worktree-side predicate observes the landing without a fetch.
	git("-C", repoDir, "switch", "-q", "develop")
	git("-C", repoDir, "merge", "-q", "--no-ff", "-m", "merge WT-landing03-fix", "WT-landing03-fix")
	git("-C", repoDir, "push", "-q", "origin", "develop")

	var out bytes.Buffer
	cleanupSessionWorktree(worktreeCfg(true), wtPath, true, &out)

	if _, err := os.Stat(wtPath); err == nil {
		t.Fatalf("merged worktree should have been removed; notice: %s", out.String())
	}
	if !strings.Contains(out.String(), SessionExitCleanupNoticePrefix) {
		t.Fatalf("expected removal notice with prefix %q, got %q", SessionExitCleanupNoticePrefix, out.String())
	}
}

// TestCleanupSessionWorktree_SquashMergedPatchIdRemoved pins the git-cherry
// arm of the decided predicate: a SQUASH merge leaves no commit ancestry
// (the branch tip is NOT an ancestor of origin/develop), but the patches are
// upstream — `git cherry` answers empty and the disposal proceeds.
// SPEC-WORKTREE-SQUASH-MERGE-001: reachability alone cannot see a squash.
func TestCleanupSessionWorktree_SquashMergedPatchIdRemoved(t *testing.T) {
	repoDir, wtPath := realLandingRepo(t, true)
	git := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("-C", repoDir, "switch", "-q", "develop")
	git("-C", repoDir, "merge", "--squash", "WT-landing03-fix")
	git("-C", repoDir, "commit", "-qm", "card work: squash-merged")
	git("-C", repoDir, "push", "-q", "origin", "develop")

	var out bytes.Buffer
	cleanupSessionWorktree(worktreeCfg(true), wtPath, true, &out)

	if _, err := os.Stat(wtPath); err == nil {
		t.Fatalf("squash-merged worktree should have been removed via patch-id equivalence; notice: %s", out.String())
	}
	if !strings.Contains(out.String(), SessionExitCleanupNoticePrefix) {
		t.Fatalf("expected removal notice with prefix %q, got %q", SessionExitCleanupNoticePrefix, out.String())
	}
}

// TestCleanupSessionWorktree_MissingIntegrationRefPreserved is the
// fail-open arm ② of REQ-WSS-302: a repository whose remote carries no
// develop branch (a main-based distributed user) has no integration ref to
// confirm against — preservation with a notice is the only safe answer, and
// the stale/absent ref can never misjudge toward removal.
func TestCleanupSessionWorktree_MissingIntegrationRefPreserved(t *testing.T) {
	_, wtPath := realLandingRepo(t, false)

	var out bytes.Buffer
	cleanupSessionWorktree(worktreeCfg(true), wtPath, true, &out)

	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("worktree was REMOVED without an integration ref to confirm against; stat: %v\nnotice: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "landing-check failed") {
		t.Fatalf("preserved notice must name the failed landing check, got %q", out.String())
	}
}

// TestCleanupSessionWorktree_LandingCheckErrorPreserved is the fail-open arm
// ①/③ of REQ-WSS-302: an unreadable landing answer preserves, exactly like
// the dirty and unpushed guards' error paths.
func TestCleanupSessionWorktree_LandingCheckErrorPreserved(t *testing.T) {
	_, wtPath := realLandingRepo(t, true)
	swapSessionWorktreeSeams(t, swSeams{
		landed: func(string) (bool, error) { return false, errors.New("git cherry exploded") },
	})

	var out bytes.Buffer
	cleanupSessionWorktree(worktreeCfg(true), wtPath, true, &out)

	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("worktree was REMOVED while the landing check errored; stat: %v\nnotice: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "landing-check failed") {
		t.Fatalf("preserved notice must name the failed landing check, got %q", out.String())
	}
}
