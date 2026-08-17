package worktree

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/session"
)

// anchoredStaleEnv installs the staleTestEnv fixtures for one worktree at
// tree whose branch is fully merged and whose working tree is clean — the
// only remaining removal blocker is the live anchored session.
func anchoredStaleEnv(t *testing.T, tree string) *[]string {
	t.Helper()
	writeTreeRegistry(t, tree, []session.Entry{anchoredEntry(t, tree, os.Getpid())})
	return staleTestEnv(t,
		[]git.Worktree{{Path: tree, Branch: "feat/anchored"}},
		map[string]string{}, // clean working tree
		map[string]bool{"feat/anchored": true},
	)
}

// TestCleanStale_KeepsAnchoredWorktree (t73 surface 2): a clean, fully
// merged worktree that anchors a live session must be kept and reported,
// never removed by the --stale sweep.
func TestCleanStale_KeepsAnchoredWorktree(t *testing.T) {
	tree := t.TempDir()
	removed := anchoredStaleEnv(t, tree)

	out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("anchored worktree was removed: %v", *removed)
	}
	if !strings.Contains(out, "live session anchored") {
		t.Errorf("expected a keep reason naming the anchored session, got:\n%s", out)
	}
}

// TestCleanMergedOnly_KeepsAnchoredWorktree: --merged-only has no dirty
// guard of its own, so the anchor guard is the only protection between the
// sweep and a live lane's tree.
func TestCleanMergedOnly_KeepsAnchoredWorktree(t *testing.T) {
	tree := t.TempDir()
	removed := anchoredStaleEnv(t, tree)

	out, err := runStaleClean(t, map[string]string{"merged-only": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("anchored worktree was removed by --merged-only: %v", *removed)
	}
	if !strings.Contains(out, "live session") || !strings.Contains(out, tree) {
		t.Errorf("expected a keeping notice naming the anchored worktree, got:\n%s", out)
	}
}
