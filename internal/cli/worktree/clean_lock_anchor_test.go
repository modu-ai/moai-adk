package worktree

// clean_lock_anchor_test.go — SPEC-WORKTREE-REAPER-001 M2, AC-WR-015 and
// AC-WR-026: the two `moai worktree clean` sweeps consume the SHARED lock-∪-
// registry anchor decision, so a worktree anchored by a LOCK ALONE — with no
// registry entry at all — survives both of them.
//
// The registry is deliberately empty in both fixtures. It is the source
// measured to name 1 of 5 live anchors, so a fixture that let it answer would
// prove nothing about the repair.

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// lockedStaleEnv installs the staleTestEnv fixtures for a clean, merged
// worktree at tree, plus a `git worktree list --porcelain` stub reporting it
// as locked by a LIVE pid. No registry entry is written.
func lockedStaleEnv(t *testing.T, tree, branch string) *[]string {
	t.Helper()
	removed := staleTestEnv(t,
		[]git.Worktree{{Path: tree, Branch: branch}},
		map[string]string{}, // clean working tree
		map[string]bool{branch: true},
	)

	// staleTestEnv stubs gitWorktreeCmd for the `-C <path> status --porcelain`
	// shape only; extend it with the porcelain listing the lock lives in.
	origGitCmd := gitWorktreeCmd
	gitWorktreeCmd = func(args ...string) (string, error) {
		if len(args) >= 3 && args[0] == "worktree" && args[1] == "list" {
			return "worktree " + tree + "\nHEAD 1111\nbranch refs/heads/" + branch +
				"\nlocked claude session t213 (pid " + strconv.Itoa(os.Getpid()) + " start Sun Aug 23 07:26:09 2026)\n", nil
		}
		return origGitCmd(args...)
	}
	t.Cleanup(func() { gitWorktreeCmd = origGitCmd })

	return removed
}

// TestCleanStale_LockAnchoredWorktreeKept is AC-WR-015: the --stale sweep —
// the consumer with the widest blast radius, since it sweeps every registered
// tree rather than only WT-* — keeps a lock-anchored worktree even with --yes.
func TestCleanStale_LockAnchoredWorktreeKept(t *testing.T) {
	tree := t.TempDir()
	removed := lockedStaleEnv(t, tree, "feat/lock-anchored")

	out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("lock-anchored worktree was removed by --stale: %v", *removed)
	}
	if !strings.Contains(out, "live session anchored") {
		t.Errorf("expected a keep reason naming the anchored session, got:\n%s", out)
	}
	if !strings.Contains(out, "lock") {
		t.Errorf("expected the keep reason to name the LOCK as the source, got:\n%s", out)
	}
}

// TestCleanMergedOnly_LockAnchoredWorktreeKept is AC-WR-026: the third
// consumer, and the most exposed — `--merged-only` has no dirty guard of its
// own, so the anchor decision is its ONLY protection.
func TestCleanMergedOnly_LockAnchoredWorktreeKept(t *testing.T) {
	tree := t.TempDir()
	removed := lockedStaleEnv(t, tree, "feat/lock-anchored-merged")

	out, err := runStaleClean(t, map[string]string{"merged-only": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("lock-anchored worktree was removed by --merged-only: %v", *removed)
	}
	if !strings.Contains(out, tree) || !strings.Contains(out, "lock") {
		t.Errorf("expected a keeping notice naming the worktree and the lock source, got:\n%s", out)
	}
}
