package git

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWorktreeAdd_NewBranch(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	wtPath := filepath.Join(resolveSymlinks(t, t.TempDir()), "wt-parallel")

	if err := wm.Add(wtPath, "feature/parallel"); err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	// Verify directory exists.
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Error("worktree directory was not created")
	}

	// Verify it appears in List.
	worktrees, err := wm.List()
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, wt := range worktrees {
		// Normalize paths for cross-platform comparison (Windows uses \ instead of /)
		normalizedWtPath := filepath.ToSlash(wt.Path)
		normalizedExpectedPath := filepath.ToSlash(wtPath)
		if normalizedWtPath == normalizedExpectedPath {
			found = true
			if wt.Branch != "feature/parallel" {
				t.Errorf("worktree.Branch = %q, want %q", wt.Branch, "feature/parallel")
			}
			break
		}
	}
	if !found {
		t.Errorf("added worktree not found in List(); paths: %v", worktreePaths(worktrees))
	}
}

func TestWorktreeAdd_ExistingBranch(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)
	wm := NewWorktreeManager(dir)

	// Pre-create the branch.
	if err := bm.Create("feature/existing"); err != nil {
		t.Fatal(err)
	}

	wtPath := filepath.Join(resolveSymlinks(t, t.TempDir()), "wt-existing")

	if err := wm.Add(wtPath, "feature/existing"); err != nil {
		t.Fatalf("Add() with existing branch error: %v", err)
	}

	// Verify the worktree is linked to the existing branch.
	worktrees, err := wm.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, wt := range worktrees {
		// Normalize paths for cross-platform comparison
		normalizedWtPath := filepath.ToSlash(wt.Path)
		normalizedExpectedPath := filepath.ToSlash(wtPath)
		if normalizedWtPath == normalizedExpectedPath {
			if wt.Branch != "feature/existing" {
				t.Errorf("worktree.Branch = %q, want %q", wt.Branch, "feature/existing")
			}
			return
		}
	}
	t.Errorf("worktree with existing branch not found in List(); paths: %v", worktreePaths(worktrees))
}

func TestWorktreeAdd_PathExists(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	// Create the path first.
	existingPath := filepath.Join(t.TempDir(), "existing-dir")
	if err := os.MkdirAll(existingPath, 0o755); err != nil {
		t.Fatal(err)
	}

	err := wm.Add(existingPath, "feature/new")
	if err == nil {
		t.Fatal("Add() at existing path should return error")
	}
	if !errors.Is(err, ErrWorktreePathExists) {
		t.Errorf("error = %v, want ErrWorktreePathExists", err)
	}
}

func TestWorktreeList(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	// Create two additional worktrees.
	wt1 := filepath.Join(resolveSymlinks(t, t.TempDir()), "wt-1")
	wt2 := filepath.Join(resolveSymlinks(t, t.TempDir()), "wt-2")

	if err := wm.Add(wt1, "feat-1"); err != nil {
		t.Fatal(err)
	}
	if err := wm.Add(wt2, "feat-2"); err != nil {
		t.Fatal(err)
	}

	worktrees, err := wm.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	// Expect 3: main worktree + 2 added.
	if len(worktrees) != 3 {
		t.Fatalf("List() returned %d worktrees, want 3", len(worktrees))
	}

	for i, wt := range worktrees {
		if wt.Path == "" {
			t.Errorf("worktrees[%d].Path is empty", i)
		}
		if wt.HEAD == "" {
			t.Errorf("worktrees[%d].HEAD is empty", i)
		}
	}
}

func TestWorktreeList_PorcelainParsing(t *testing.T) {
	input := "worktree /Users/goos/project\nHEAD abc123def456\nbranch refs/heads/main\n\nworktree /tmp/wt-feature\nHEAD def789abc012\nbranch refs/heads/feature\n"

	worktrees := parsePorcelainWorktreeList(input)

	if len(worktrees) != 2 {
		t.Fatalf("parsePorcelainWorktreeList() returned %d entries, want 2", len(worktrees))
	}

	// Use filepath.FromSlash for expected paths since parsePorcelainWorktreeList
	// normalizes paths with filepath.Clean (converts / to \ on Windows)
	tests := []struct {
		idx    int
		path   string
		head   string
		branch string
	}{
		{0, filepath.FromSlash("/Users/goos/project"), "abc123def456", "main"},
		{1, filepath.FromSlash("/tmp/wt-feature"), "def789abc012", "feature"},
	}

	for _, tt := range tests {
		wt := worktrees[tt.idx]
		if wt.Path != tt.path {
			t.Errorf("worktrees[%d].Path = %q, want %q", tt.idx, wt.Path, tt.path)
		}
		if wt.HEAD != tt.head {
			t.Errorf("worktrees[%d].HEAD = %q, want %q", tt.idx, wt.HEAD, tt.head)
		}
		if wt.Branch != tt.branch {
			t.Errorf("worktrees[%d].Branch = %q, want %q", tt.idx, wt.Branch, tt.branch)
		}
	}
}

func TestWorktreeRemove_Success(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	wtPath := filepath.Join(resolveSymlinks(t, t.TempDir()), "wt-remove")
	if err := wm.Add(wtPath, "feat-remove"); err != nil {
		t.Fatal(err)
	}

	if err := wm.Remove(wtPath, false); err != nil {
		t.Fatalf("Remove() error: %v", err)
	}

	// Verify the worktree is gone from List.
	worktrees, err := wm.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, wt := range worktrees {
		if wt.Path == wtPath {
			t.Error("removed worktree still appears in List()")
		}
	}
}

func TestWorktreeRemove_NotFound(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	err := wm.Remove("/tmp/nonexistent-worktree-path-"+t.Name(), false)
	if err == nil {
		t.Fatal("Remove() on nonexistent path should return error")
	}
	if !errors.Is(err, ErrWorktreeNotFound) {
		t.Errorf("error = %v, want ErrWorktreeNotFound", err)
	}
}

func TestWorktreePrune(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	// Create a worktree then manually delete its directory.
	wtPath := filepath.Join(resolveSymlinks(t, t.TempDir()), "wt-stale")
	if err := wm.Add(wtPath, "feat-stale"); err != nil {
		t.Fatal(err)
	}

	// Manually remove the directory to simulate a stale reference.
	if err := os.RemoveAll(wtPath); err != nil {
		t.Fatal(err)
	}

	// Prune should clean up the stale reference.
	if err := wm.Prune(); err != nil {
		t.Fatalf("Prune() error: %v", err)
	}

	// After pruning, the stale worktree should not appear in list.
	worktrees, err := wm.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, wt := range worktrees {
		if wt.Path == wtPath {
			t.Error("stale worktree still appears in List() after Prune()")
		}
	}
}

// worktreePaths extracts paths from a slice of Worktrees for debug output.
func worktreePaths(wts []Worktree) []string {
	paths := make([]string, len(wts))
	for i, wt := range wts {
		paths[i] = wt.Path
	}
	return paths
}
