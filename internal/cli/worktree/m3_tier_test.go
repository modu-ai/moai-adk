package worktree

// m3_tier_test.go — SPEC-WORKTREE-HOME-ROOT-001 M3 RED (card t1640):
// AC-WHR-009, tier by creation class at the home root.
//
// Written before isL1WorktreePath learns the home root through the M1
// classifier: today it answers by the two in-repo prefixes only, so a
// home-root moai-made tree — WT- or a migrated non-WT- shape — reads as
// L2. That failure is the RED. The legacy in-repo shapes are pinned: they
// stay L1, and the existing done_l1_tier_guard_test.go expectations are
// unchanged by the M3 work.

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

// m3TierReal resolves symlinks (the macOS /var alias) so git-derived paths
// agree with the asserted ones.
func m3TierReal(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", p, err)
	}
	return r
}

// m3TierFixture creates a scratch repository with two home-root trees —
// one on a WT- branch, one on a migrated release-shaped branch — both
// owned by the repository through the resolver's marker, and returns the
// three paths the tier predicates read.
func m3TierFixture(t *testing.T) (repo, wtTree, releaseTree string) {
	t.Helper()
	repo = filepath.Join(m3TierReal(t, t.TempDir()), "tier-repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	tf := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
		}
	}
	git := tf
	git(repo, "init", "-q", "--template=", "-b", "main")
	git(repo, "config", "user.email", "tier@example.com")
	git(repo, "config", "user.name", "Tier Test")
	git(repo, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	git(repo, "add", "README.md")
	git(repo, "commit", "-q", "-m", "seed")

	home := m3TierReal(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)

	// The resolver assigns the key and writes the owner marker naming this
	// repository — the marker witness the tier classifier reads.
	if _, err := worktreeroot.Resolve(repo); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	root, err := worktreeroot.Root(repo)
	if err != nil {
		t.Fatalf("root: %v", err)
	}
	wtTree = filepath.Join(root, "wt-tree")
	releaseTree = filepath.Join(root, "release-v314")
	git(repo, "worktree", "add", "-q", "-b", "WT-tier", "--", wtTree)
	git(repo, "worktree", "add", "-q", "-b", "release/v3.1.4", "--", releaseTree)
	return repo, wtTree, releaseTree
}

// TestIsL1WorktreePath_HomeRootIsL1 pins AC-WHR-009: every moai-made tree
// at the home root is an L1 session tree — including one whose branch does
// not begin with WT-.
func TestIsL1WorktreePath_HomeRootIsL1(t *testing.T) {
	_, wtTree, releaseTree := m3TierFixture(t)
	for name, tree := range map[string]string{"WT- home tree": wtTree, "non-WT- home tree": releaseTree} {
		if !isL1WorktreePath(tree) {
			t.Errorf("%s (%s) is not L1; a home-root moai-made tree is an L1 session tree (REQ-WHR-009)", name, tree)
		}
		if got := sweepTierOf(tree); got != sweepTierL1 {
			t.Errorf("%s: sweepTierOf = %q, want %q", name, got, sweepTierL1)
		}
	}
}

// TestIsL1WorktreePath_LegacyRootsPinned pins the legacy tiers the M3
// change must preserve: both in-repo roots stay L1.
func TestIsL1WorktreePath_LegacyRootsPinned(t *testing.T) {
	repo, _, _ := m3TierFixture(t)
	for name, tree := range map[string]string{
		"legacy-moai":   filepath.Join(repo, ".moai", "worktrees", "t9"),
		"legacy-claude": filepath.Join(repo, ".claude", "worktrees", "t9"),
	} {
		if err := os.MkdirAll(tree, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", tree, err)
		}
		if !isL1WorktreePath(tree) {
			t.Errorf("%s tree (%s) is not L1; the legacy in-repo roots remain L1", name, tree)
		}
	}
}
