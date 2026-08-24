package worktree

// clean_base_branch_test.go — SPEC-WORKTREE-REAPER-001 post-audit repair F4.
//
// The `--base` default moved from the local `main` to `origin/main`
// (REQ-WR-022). A worktree checks out a LOCAL branch, so a literal comparison
// would stop recognising the base checkout the moment that default changed,
// and a second worktree sitting on `main` would then be handed to the merge
// predicate — which reports `main` as merged into `origin/main` and makes it
// removable. The trailing-segment comparison is what keeps that guard
// standing, and until this file it was bound by no criterion at all:
// AC-WR-019 asserts only which base STRING reaches IsBranchMerged.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// TestIsBaseBranch pins the predicate in both directions. The over-breadth is
// deliberate and one-way: it errs toward KEEPING, never toward removing.
func TestIsBaseBranch(t *testing.T) {
	tests := []struct {
		name   string
		branch string
		base   string
		want   bool
	}{
		{"local main against the remote-tracking base", "main", "origin/main", true},
		{"plain equality still holds", "main", "main", true},
		{"a namespaced branch ending in the base segment is NOT the base", "feature/main", "origin/main", false},
		{"an empty branch is never the base", "", "origin/main", false},
		{"an unrelated branch is not the base", "WT-worktree-reaper", "origin/main", false},
		{"a different remote's main still matches on the trailing segment", "main", "upstream/main", true},
		{"the trailing segment must match exactly", "mainline", "origin/main", false},
		{"a base with no slash falls back to equality", "WT-x", "develop", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBaseBranch(tt.branch, tt.base); got != tt.want {
				t.Errorf("isBaseBranch(%q, %q) = %v, want %v", tt.branch, tt.base, got, tt.want)
			}
		})
	}
}

// TestCleanStale_KeepsWorktreeOnBaseBranch binds the predicate through the
// sweep rather than in isolation: a SECOND worktree checked out on the local
// `main` — clean, and reported merged into `origin/main` by the merge
// predicate — must be kept, and the keep reason must name the base branch.
func TestCleanStale_KeepsWorktreeOnBaseBranch(t *testing.T) {
	removed := staleTestEnv(t,
		[]git.Worktree{{Path: "/wt/second-main", Branch: "main"}},
		map[string]string{},
		map[string]bool{"main": true},
	)

	out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("a worktree on the base branch was removed: %v", *removed)
	}
	if !strings.Contains(out, "checked out on the base branch") {
		t.Errorf("expected the base-branch keep reason, got:\n%s", out)
	}
}
