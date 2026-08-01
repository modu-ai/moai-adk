package git

import (
	"path/filepath"
	"testing"
)

// TestTrimBranchListMarker covers both markers `git branch` can prefix a
// listing line with, plus the unmarked case.
func TestTrimBranchListMarker(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"current branch marker", "* main", "main"},
		{"worktree branch marker", "+ feature/in-worktree", "feature/in-worktree"},
		{"no marker", "  feature/plain", "  feature/plain"},
		{"marker-like name is not stripped", "*feature", "*feature"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := trimBranchListMarker(tc.in); got != tc.want {
				t.Errorf("trimBranchListMarker(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestWorktreeIsBranchMerged_BranchCheckedOutInWorktree is the regression guard
// for the `+ ` marker. A branch checked out in another worktree is listed by
// `git branch --merged` with a "+ " prefix rather than "* ", and every
// worktree-scoped caller hits exactly that case — a worktree's branch is by
// definition checked out somewhere. Stripping only "* " made this report false
// for every worktree branch, silently disabling worktree cleanup.
func TestWorktreeIsBranchMerged_BranchCheckedOutInWorktree(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	// Create a worktree on a new branch with no commits of its own, so it is
	// genuinely merged into main and only the marker can change the verdict.
	wtPath := filepath.Join(t.TempDir(), "merged-wt")
	if err := wm.Add(wtPath, "feature/in-worktree"); err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	merged, err := wm.IsBranchMerged("feature/in-worktree", "main")
	if err != nil {
		t.Fatalf("IsBranchMerged() error: %v", err)
	}
	if !merged {
		t.Error("a worktree branch with no unique commits must report as merged into main")
	}
}
