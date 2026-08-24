package worktree

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// runStaleClean builds a fresh clean command (so cobra flag state cannot leak
// between tests), applies the given flags, and returns its combined output.
func runStaleClean(t *testing.T, flags map[string]string) (string, error) {
	t.Helper()

	cmd := newCleanCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	for name, value := range flags {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set flag %s=%s: %v", name, value, err)
		}
	}
	err := cmd.RunE(cmd, nil)
	return buf.String(), err
}

// staleTestEnv installs a mock provider plus stubbed git/porcelain and repo-root
// helpers, and restores every global it touched when the test ends.
//
// porcelain maps a worktree path to the `git status --porcelain` output the
// stub should return; an absent entry means a clean working tree.
func staleTestEnv(t *testing.T, worktrees []git.Worktree, porcelain map[string]string, merged map[string]bool) *[]string {
	t.Helper()

	origProvider := WorktreeProvider
	origGitCmd := gitWorktreeCmd
	origRepoRoot := gitRepoRootFunc
	origMerged := mockIsBranchMergedFunc
	t.Cleanup(func() {
		WorktreeProvider = origProvider
		gitWorktreeCmd = origGitCmd
		gitRepoRootFunc = origRepoRoot
		mockIsBranchMergedFunc = origMerged
	})

	removedPaths := []string{}

	WorktreeProvider = &mockWorktreeManager{
		rootPath: "/repo",
		listFunc: func() ([]git.Worktree, error) { return worktrees, nil },
		removeFunc: func(path string, force bool) error {
			if force {
				t.Errorf("--stale must never force-remove; got force=true for %s", path)
			}
			removedPaths = append(removedPaths, path)
			return nil
		},
	}
	mockIsBranchMergedFunc = func(branch, _ string) (bool, error) {
		return merged[branch], nil
	}
	gitWorktreeCmd = func(args ...string) (string, error) {
		// Expected shape: -C <path> status --porcelain
		if len(args) >= 2 && args[0] == "-C" {
			return porcelain[args[1]], nil
		}
		return "", nil
	}
	gitRepoRootFunc = func() (string, error) { return "/repo", nil }

	return &removedPaths
}

// TestCleanStale_KeepsDirtyWorktree is the load-bearing safety test: a worktree
// with uncommitted work must be kept and reported, never removed — even when its
// branch is fully merged.
func TestCleanStale_KeepsDirtyWorktree(t *testing.T) {
	removed := staleTestEnv(t,
		[]git.Worktree{{Path: "/wt/dirty", Branch: "feat/dirty"}},
		map[string]string{"/wt/dirty": " M internal/foo.go\n?? scratch.txt\n"},
		map[string]bool{"feat/dirty": true},
	)

	out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("dirty worktree was removed: %v", *removed)
	}
	if !strings.Contains(out, "uncommitted or untracked changes") {
		t.Errorf("expected a keep reason naming uncommitted changes, got:\n%s", out)
	}
}

// TestCleanStale_KeepsUnmergedWorktree verifies the second half of the
// predicate: a clean worktree whose branch still carries its own commits is kept.
func TestCleanStale_KeepsUnmergedWorktree(t *testing.T) {
	removed := staleTestEnv(t,
		[]git.Worktree{{Path: "/wt/unmerged", Branch: "feat/unmerged"}},
		map[string]string{},
		map[string]bool{"feat/unmerged": false},
	)

	out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("unmerged worktree was removed: %v", *removed)
	}
	// The base ref moved from the local `main` to `origin/main`
	// (SPEC-WORKTREE-REAPER-001 REQ-WR-022), so the keep reason names the ref
	// actually compared against.
	if !strings.Contains(out, "commits not in origin/main") {
		t.Errorf("expected a keep reason naming unmerged commits, got:\n%s", out)
	}
}

// TestCleanStale_PreviewsByDefault verifies --stale without --yes removes nothing.
func TestCleanStale_PreviewsByDefault(t *testing.T) {
	removed := staleTestEnv(t,
		[]git.Worktree{{Path: "/wt/spent", Branch: "feat/spent"}},
		map[string]string{},
		map[string]bool{"feat/spent": true},
	)

	out, err := runStaleClean(t, map[string]string{"stale": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("preview run removed worktrees: %v", *removed)
	}
	if !strings.Contains(out, "Would remove 1 stale worktree(s)") {
		t.Errorf("expected a preview listing, got:\n%s", out)
	}
	if !strings.Contains(out, "--yes") {
		t.Errorf("expected the preview to point at --yes, got:\n%s", out)
	}
}

// TestCleanStale_RemovesWithYes verifies the qualifying worktree is removed
// (non-force) once --yes is given.
func TestCleanStale_RemovesWithYes(t *testing.T) {
	removed := staleTestEnv(t,
		[]git.Worktree{{Path: "/wt/spent", Branch: "feat/spent"}},
		map[string]string{},
		map[string]bool{"feat/spent": true},
	)

	out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 1 || (*removed)[0] != "/wt/spent" {
		t.Fatalf("expected /wt/spent to be removed, got: %v", *removed)
	}
	if !strings.Contains(out, "Branches were left intact") {
		t.Errorf("expected the branch-preservation notice, got:\n%s", out)
	}
}

// TestCleanStale_SkipsProtectedAndDetached verifies the sweep never touches the
// repository root, the worktree it runs in, the base branch, or a detached HEAD.
func TestCleanStale_SkipsProtectedAndDetached(t *testing.T) {
	removed := staleTestEnv(t,
		[]git.Worktree{
			{Path: "/repo", Branch: "main"},
			{Path: "/wt/detached", Branch: ""},
			{Path: "/wt/on-base", Branch: "main"},
		},
		map[string]string{},
		map[string]bool{"main": true},
	)

	out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("protected/detached/base worktrees were removed: %v", *removed)
	}
	if strings.Contains(out, "/repo [main]") {
		t.Errorf("repository root should be skipped silently, got:\n%s", out)
	}
	if !strings.Contains(out, "detached HEAD") {
		t.Errorf("expected the detached-HEAD keep reason, got:\n%s", out)
	}
	if !strings.Contains(out, "checked out on the base branch") {
		t.Errorf("expected the base-branch keep reason, got:\n%s", out)
	}
}

// TestCleanStale_RejectsMergedOnlyCombination verifies the two sweep modes are
// mutually exclusive rather than silently picking one.
func TestCleanStale_RejectsMergedOnlyCombination(t *testing.T) {
	staleTestEnv(t, nil, map[string]string{}, map[string]bool{})

	_, err := runStaleClean(t, map[string]string{"stale": "true", "merged-only": "true"})
	if err == nil {
		t.Fatal("expected --stale with --merged-only to be rejected")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected a mutual-exclusion error, got: %v", err)
	}
}
