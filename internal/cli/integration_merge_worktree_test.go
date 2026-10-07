package cli

// integration_merge_worktree_test.go — the merge verb's integration-worktree
// resolution (card t1479, card-review r3 F3): the primary checkout holding
// the integration branch is REFUSED as a merge target, a linked worktree
// holding it resolves, and the source routes through the guarded helper.
// The fixture is a REAL git repository in every shape — the r2-D pin test's
// fixture was not a repo, so its primary case was structurally unreachable
// and never exercised the guard it described.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// f3GitInit builds a primary-shaped checkout in dir with branch checked out
// and one commit on it.
func f3GitInit(t *testing.T, dir, branch string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	git("init", "-q", "-b", branch)
	git("config", "user.email", "t@t.local")
	git("config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "base.txt"), []byte("base"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "base.txt")
	git("commit", "-q", "-m", "base")
}

func TestIntegrationMergeWorktreeRefusesPrimaryHoldingBranch(t *testing.T) {
	// F3: the primary checkout IS a tree holding develop — the resolver's
	// "any tree" answer — and the merge verb must refuse it, not merge into
	// the shared primary checkout.
	root := t.TempDir()
	f3GitInit(t, root, "develop")

	resolved, err := integrationMergeWorktree(root, "develop")
	if err == nil {
		t.Fatalf("the primary checkout holding %q must be refused as a merge target, got %q", "develop", resolved)
	}
	if !strings.Contains(err.Error(), "primary checkout") {
		t.Fatalf("the refusal must name the primary checkout: %v", err)
	}
}

func TestIntegrationMergeWorktreeResolvesLinkedWorktree(t *testing.T) {
	// The legitimate shape stands: the primary holds another branch and a
	// LINKED worktree holds develop — the resolution returns the worktree.
	root := t.TempDir()
	f3GitInit(t, root, "main")
	integ := filepath.Join(root, "integ")
	cmd := exec.Command("git", "worktree", "add", "-q", "-b", "develop", integ, "HEAD")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add: %v: %s", err, out)
	}

	resolved, err := integrationMergeWorktree(root, "develop")
	if err != nil {
		t.Fatalf("a linked worktree holding %q must resolve: %v", "develop", err)
	}
	if !factorySameTree(resolved, integ) {
		t.Fatalf("the resolution must name the linked worktree %s, got %s", integ, resolved)
	}
}

func TestIntegrationMergeWorktreeRefusesUnheldBranch(t *testing.T) {
	// No tree at all holds the branch: the provisioning refusal.
	root := t.TempDir()
	f3GitInit(t, root, "main")

	_, err := integrationMergeWorktree(root, "develop")
	if err == nil || !strings.Contains(err.Error(), "no worktree holds") {
		t.Fatalf("an unheld integration branch must refuse with the provisioning message: %v", err)
	}
}

func TestIntegrationMergeWorktreeResolutionRoutesThroughGuard(t *testing.T) {
	// Static guard: the merge verb resolves THROUGH integrationMergeWorktree
	// (the primary refusal), and the bare resolver is called only INSIDE the
	// guard — a verb-side call reinstates the F3 hazard this file's fixture
	// reproduces.
	source, err := os.ReadFile("integration_merge.go")
	if err != nil {
		t.Fatal(err)
	}
	body := string(source)
	if !strings.Contains(body, "integrationMergeWorktree(root, integBranch)") {
		t.Fatalf("the merge verb must resolve through integrationMergeWorktree (the primary refusal)")
	}
	if got := strings.Count(body, "factoryWorktreeForBranchIn("); got != 1 {
		t.Fatalf("the bare resolver must be called only inside integrationMergeWorktree, found %d call sites", got)
	}
}
