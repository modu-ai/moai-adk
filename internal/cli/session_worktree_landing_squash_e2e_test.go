package cli

// session_worktree_landing_squash_e2e_test.go — SPEC-GITHUB-FLOW-DEFAULT-001
// M2-A (card t1453): the real exit-cleanup function end to end under github-flow.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCleanupSessionWorktree_GitHubFlowSquash drives cleanupSessionWorktree: a
// two-commit squash-merged card is disposed from layers 1-2 alone, and a squash
// that only the PR layer can confirm is preserved with a notice naming the
// CONFIGURED integration ref. `gh` is a recording shim; it must stay uncalled.
func TestCleanupSessionWorktree_GitHubFlowSquash(t *testing.T) {
	marker := withGHShim(t)

	t.Run("two_commit_squash_is_removed", func(t *testing.T) {
		f := gfdSessRepo(t, "main", gfdSessGitHubFlowYAML)
		f.cardCommit(t, 5, "card")
		f.cardCommit(t, 15, "card")
		gfdSessGit(t, f.tree, "push", "-q", "-u", "origin", "WT-sess-card")
		f.squash(t)
		f.push(t)
		chdirTemp(t, f.repo)
		var out strings.Builder
		cleanupSessionWorktree(worktreeCfg(true), f.tree, true, &out)
		if _, err := os.Stat(f.tree); err == nil {
			t.Fatalf("a squash-merged two-commit card must be removed by session-exit cleanup; notice: %s", out.String())
		}
	})

	t.Run("layer_3_only_squash_is_preserved_naming_the_configured_ref", func(t *testing.T) {
		f := gfdSessRepo(t, "main", gfdSessGitHubFlowYAML)
		f.cardCommit(t, 10, "card")
		gfdSessGit(t, f.tree, "push", "-q", "-u", "origin", "WT-sess-card")
		f.mainCommit(t, 13, "other")
		f.squash(t)
		f.mainCommit(t, 10, "later")
		f.push(t)
		chdirTemp(t, f.repo)
		var out strings.Builder
		cleanupSessionWorktree(worktreeCfg(true), f.tree, true, &out)
		if _, err := os.Stat(f.tree); err != nil {
			t.Fatalf("a squash only the PR layer can confirm must be preserved at session exit; stat: %v\nnotice: %s", err, out.String())
		}
		if !strings.Contains(out.String(), "unconfirmed") || !strings.Contains(out.String(), "refs/remotes/origin/main") {
			t.Errorf("the notice must say the landing is unconfirmed and name refs/remotes/origin/main, got %q", out.String())
		}
	})

	if ghWasCalled(t, marker) {
		t.Fatal("session-exit cleanup ran gh")
	}
}

func TestCleanupSessionWorktreePreservesWhitespaceDistinctContent(t *testing.T) {
	f := gfdSessRepo(t, "main", gfdSessGitHubFlowYAML)
	card := "package fixture\nconst value = \"a  b\"\n"
	remote := "package fixture\nconst value = \"a b\"\n"
	for _, change := range []struct{ dir, content string }{{f.tree, card}, {f.repo, remote}} {
		if err := os.WriteFile(filepath.Join(change.dir, "value.go"), []byte(change.content), 0o600); err != nil {
			t.Fatal(err)
		}
		gfdSessGit(t, change.dir, "add", "value.go")
		gfdSessGit(t, change.dir, "commit", "-q", "-m", "distinct literal")
	}
	gfdSessGit(t, f.tree, "push", "-q", "-u", "origin", "WT-sess-card")
	f.push(t)
	if cherry := gfdSessGit(t, f.tree, "cherry", "origin/main", "HEAD"); !strings.HasPrefix(cherry, "-") {
		t.Fatalf("control: cherry must collide: %q", cherry)
	}
	chdirTemp(t, f.repo)
	if landed, err := gitBranchLandedReal(f.tree); err != nil || landed {
		t.Errorf("distinct bytes must not be landed: %v, %v", landed, err)
	}
	var out strings.Builder
	cleanupSessionWorktree(worktreeCfg(true), f.tree, true, &out)
	if raw, err := os.ReadFile(filepath.Join(f.tree, "value.go")); err != nil || string(raw) != card {
		t.Errorf("unlanded bytes must survive cleanup: %q, %v; notice %s", raw, err, out.String())
	}
}
