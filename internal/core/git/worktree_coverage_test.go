package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Tests targeting uncovered functions in worktree.go and manager.go
// to push core/git coverage above 85%.

func TestWorktreeRepair(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	// Repair should succeed on a healthy repo
	if err := wm.Repair(); err != nil {
		t.Fatalf("Repair() error: %v", err)
	}
}

func TestWorktreeRoot(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	root := wm.Root()
	if root != dir {
		t.Errorf("Root() = %q, want %q", root, dir)
	}
}

func TestWorktreeSync_Merge(t *testing.T) {
	localDir, _ := initTestRepoWithRemote(t)
	wm := NewWorktreeManager(localDir)

	// Create a worktree
	wtPath := filepath.Join(resolveSymlinks(t, t.TempDir()), "wt-sync")
	if err := wm.Add(wtPath, "feature/sync-test"); err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	// Sync with merge strategy (default)
	if err := wm.Sync(wtPath, "main", "merge"); err != nil {
		t.Fatalf("Sync(merge) error: %v", err)
	}
}

func TestWorktreeSync_Rebase(t *testing.T) {
	localDir, _ := initTestRepoWithRemote(t)
	wm := NewWorktreeManager(localDir)

	// Create a worktree
	wtPath := filepath.Join(resolveSymlinks(t, t.TempDir()), "wt-rebase")
	if err := wm.Add(wtPath, "feature/rebase-test"); err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	// Sync with rebase strategy
	if err := wm.Sync(wtPath, "main", "rebase"); err != nil {
		t.Fatalf("Sync(rebase) error: %v", err)
	}
}

func TestWorktreeDeleteBranch(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	// Create a branch
	bm := NewBranchManager(dir)
	if err := bm.Create("feature/to-delete"); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Delete the branch
	if err := wm.DeleteBranch("feature/to-delete"); err != nil {
		t.Fatalf("DeleteBranch() error: %v", err)
	}
}

func TestWorktreeIsBranchMerged(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)
	bm := NewBranchManager(dir)

	// Create and switch to a new branch
	if err := bm.Create("feature/test-merged"); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Switch back to main
	if err := bm.Switch("main"); err != nil {
		t.Fatalf("Switch() error: %v", err)
	}

	// Branch should be merged (it was created from main and has no new commits)
	merged, err := wm.IsBranchMerged("feature/test-merged", "main")
	if err != nil {
		t.Fatalf("IsBranchMerged() error: %v", err)
	}
	if !merged {
		t.Error("expected branch to be merged into main")
	}
}

func TestWorktreeIsBranchMerged_NotMerged(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)
	bm := NewBranchManager(dir)

	// Create and switch to a new branch
	if err := bm.Create("feature/not-merged"); err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if err := bm.Switch("feature/not-merged"); err != nil {
		t.Fatalf("Switch() error: %v", err)
	}

	// Add a commit on the feature branch
	writeTestFile(t, filepath.Join(dir, "new-file.txt"), "new content")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "new commit on feature")

	// Switch back to main
	if err := bm.Switch("main"); err != nil {
		t.Fatalf("Switch() error: %v", err)
	}

	// Branch should NOT be merged (it has a commit that main doesn't)
	merged, err := wm.IsBranchMerged("feature/not-merged", "main")
	if err != nil {
		t.Fatalf("IsBranchMerged() error: %v", err)
	}
	if merged {
		t.Error("expected branch to NOT be merged into main")
	}
}

func TestManagerLog(t *testing.T) {
	dir := initTestRepo(t)

	// Add some more commits
	writeTestFile(t, filepath.Join(dir, "file1.txt"), "content1")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "second commit")

	writeTestFile(t, filepath.Join(dir, "file2.txt"), "content2")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "third commit")

	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatalf("NewRepository() error: %v", err)
	}

	commits, err := repo.Log(3)
	if err != nil {
		t.Fatalf("Log() error: %v", err)
	}

	if len(commits) != 3 {
		t.Errorf("Log(3) returned %d commits, want 3", len(commits))
	}

	// Most recent commit should be first
	if commits[0].Message != "third commit" {
		t.Errorf("commits[0].Message = %q, want %q", commits[0].Message, "third commit")
	}
	if commits[0].Author != "Test User" {
		t.Errorf("commits[0].Author = %q, want %q", commits[0].Author, "Test User")
	}
	if commits[0].Hash == "" {
		t.Error("commits[0].Hash should not be empty")
	}
	if commits[0].Date.IsZero() {
		t.Error("commits[0].Date should not be zero")
	}
}

func TestManagerDiff(t *testing.T) {
	dir := initTestRepo(t)

	// Create a second commit
	writeTestFile(t, filepath.Join(dir, "diff-file.txt"), "initial content")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "add diff file")

	// Modify the file
	writeTestFile(t, filepath.Join(dir, "diff-file.txt"), "modified content")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "modify diff file")

	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatalf("NewRepository() error: %v", err)
	}

	diff, err := repo.Diff("HEAD~1", "HEAD")
	if err != nil {
		t.Fatalf("Diff() error: %v", err)
	}

	if diff == "" {
		t.Error("expected non-empty diff")
	}
}

func TestManagerIsClean(t *testing.T) {
	dir := initTestRepo(t)

	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatalf("NewRepository() error: %v", err)
	}

	// Clean repo should be clean
	clean, err := repo.IsClean()
	if err != nil {
		t.Fatalf("IsClean() error: %v", err)
	}
	if !clean {
		t.Error("expected clean repo")
	}

	// Add an untracked file
	writeTestFile(t, filepath.Join(dir, "untracked.txt"), "untracked")

	clean, err = repo.IsClean()
	if err != nil {
		t.Fatalf("IsClean() error: %v", err)
	}
	if clean {
		t.Error("expected dirty repo with untracked file")
	}
}

func TestManagerRoot(t *testing.T) {
	dir := initTestRepo(t)

	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatalf("NewRepository() error: %v", err)
	}

	root := repo.Root()
	if root != dir {
		t.Errorf("Root() = %q, want %q", root, dir)
	}
}

func TestManagerStatus_WithRenamedFile(t *testing.T) {
	dir := initTestRepo(t)

	// Create a file, commit, then rename it
	writeTestFile(t, filepath.Join(dir, "original.txt"), "content")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "add original")

	runGit(t, dir, "mv", "original.txt", "renamed.txt")

	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatalf("NewRepository() error: %v", err)
	}

	status, err := repo.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}

	// Renamed file should appear in staged
	if len(status.Staged) == 0 {
		t.Error("expected staged files for rename")
	}
}

func TestManagerCurrentBranch(t *testing.T) {
	dir := initTestRepo(t)

	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatalf("NewRepository() error: %v", err)
	}

	branch, err := repo.CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch() error: %v", err)
	}

	if branch != "main" {
		t.Errorf("CurrentBranch() = %q, want %q", branch, "main")
	}
}

func TestNewRepository_NotARepo(t *testing.T) {
	dir := t.TempDir()

	_, err := NewRepository(dir)
	if err == nil {
		t.Fatal("expected error for non-repo directory")
	}
}

func TestEventDetector_Poll(t *testing.T) {
	dir := initTestRepo(t)

	ed := NewEventDetector(dir, WithPollInterval(50*time.Millisecond))

	ch := make(chan GitEvent, 10)

	// Create a context that we cancel after a short duration
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := ed.Poll(ctx, ch)
	// Poll should return ctx.Err() when context is cancelled
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

func TestParsePorcelainWorktreeList_DetachedHEAD(t *testing.T) {
	input := "worktree /path/to/repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /tmp/detached\nHEAD def456\ndetached\n"

	worktrees := parsePorcelainWorktreeList(input)

	if len(worktrees) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(worktrees))
	}

	if worktrees[1].Branch != "" {
		t.Errorf("detached worktree branch = %q, want empty", worktrees[1].Branch)
	}
}

func TestParsePorcelainWorktreeList_EmptyOutput(t *testing.T) {
	worktrees := parsePorcelainWorktreeList("")
	if len(worktrees) != 0 {
		t.Errorf("expected 0 worktrees from empty output, got %d", len(worktrees))
	}
}

func TestWorktreeRemove_Force(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	wtPath := filepath.Join(resolveSymlinks(t, t.TempDir()), "wt-dirty")
	if err := wm.Add(wtPath, "feature/dirty"); err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	// Create uncommitted changes in the worktree
	writeTestFile(t, filepath.Join(wtPath, "uncommitted.txt"), "dirty content")

	// Force remove should succeed
	if err := wm.Remove(wtPath, true); err != nil {
		t.Fatalf("Remove(force=true) error: %v", err)
	}

	// Verify worktree is gone
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Error("worktree directory should be removed")
	}
}

func TestWorktreeSync_FetchError(t *testing.T) {
	// Use a repo without a remote to trigger fetch error
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	wtPath := filepath.Join(resolveSymlinks(t, t.TempDir()), "wt-no-remote")
	if err := wm.Add(wtPath, "feature/no-remote"); err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	// Sync should fail because there's no remote
	err := wm.Sync(wtPath, "main", "merge")
	if err == nil {
		t.Error("expected error when syncing without remote")
	}
}

func TestWorktreeDeleteBranch_NonExistent(t *testing.T) {
	dir := initTestRepo(t)
	wm := NewWorktreeManager(dir)

	err := wm.DeleteBranch("nonexistent-branch")
	if err == nil {
		t.Error("expected error when deleting non-existent branch")
	}
}
