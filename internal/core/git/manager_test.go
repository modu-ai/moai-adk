package git

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"
)

func TestNewRepository_Valid(t *testing.T) {
	dir := initTestRepo(t)

	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatalf("NewRepository(%q) error: %v", dir, err)
	}
	if repo == nil {
		t.Fatal("NewRepository returned nil")
	}

	// Root should be the absolute, cleaned path.
	got := repo.Root()
	want := filepath.Clean(dir)
	if got != want {
		t.Errorf("Root() = %q, want %q", got, want)
	}
}

func TestNewRepository_InvalidPath(t *testing.T) {
	dir := t.TempDir() // empty directory, not a git repo

	repo, err := NewRepository(dir)
	if err == nil {
		t.Fatal("NewRepository on non-repo should return error")
	}
	if !errors.Is(err, ErrNotRepository) {
		t.Errorf("error = %v, want ErrNotRepository", err)
	}
	if repo != nil {
		t.Error("expected nil repo on error")
	}
}

func TestNewRepository_Subdirectory(t *testing.T) {
	dir := initTestRepo(t)
	subdir := filepath.Join(dir, "sub")
	writeTestFile(t, filepath.Join(subdir, "file.txt"), "content\n")

	repo, err := NewRepository(subdir)
	if err != nil {
		t.Fatalf("NewRepository(%q) from subdirectory error: %v", subdir, err)
	}

	// Root should still be the repo root, not the subdirectory.
	got := repo.Root()
	want := filepath.Clean(dir)
	if got != want {
		t.Errorf("Root() = %q, want %q", got, want)
	}
}

func TestCurrentBranch_Normal(t *testing.T) {
	dir := initTestRepo(t)
	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	branch, err := repo.CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch() error: %v", err)
	}
	if branch != "main" {
		t.Errorf("CurrentBranch() = %q, want %q", branch, "main")
	}
}

func TestCurrentBranch_DetachedHEAD(t *testing.T) {
	dir := initTestRepo(t)

	// Detach HEAD by checking out the commit hash directly.
	hash := runGit(t, dir, "rev-parse", "HEAD")
	runGit(t, dir, "checkout", hash)

	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	branch, err := repo.CurrentBranch()
	if err == nil {
		t.Fatal("CurrentBranch() on detached HEAD should return error")
	}
	if !errors.Is(err, ErrDetachedHEAD) {
		t.Errorf("error = %v, want ErrDetachedHEAD", err)
	}
	if branch != "" {
		t.Errorf("branch = %q, want empty string", branch)
	}
}

func TestStatus_Clean(t *testing.T) {
	dir := initTestRepo(t)
	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	status, err := repo.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}

	if len(status.Staged) != 0 {
		t.Errorf("Staged = %v, want empty", status.Staged)
	}
	if len(status.Modified) != 0 {
		t.Errorf("Modified = %v, want empty", status.Modified)
	}
	if len(status.Untracked) != 0 {
		t.Errorf("Untracked = %v, want empty", status.Untracked)
	}
}

func TestStatus_Dirty(t *testing.T) {
	dir := initTestRepo(t)

	// Create staged, modified, and untracked files.
	writeTestFile(t, filepath.Join(dir, "staged.txt"), "staged content\n")
	runGit(t, dir, "add", "staged.txt")

	writeTestFile(t, filepath.Join(dir, "README.md"), "modified content\n")

	writeTestFile(t, filepath.Join(dir, "untracked.txt"), "untracked content\n")

	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	status, err := repo.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}

	if !containsFile(status.Staged, "staged.txt") {
		t.Errorf("Staged = %v, want to contain %q", status.Staged, "staged.txt")
	}
	if !containsFile(status.Modified, "README.md") {
		t.Errorf("Modified = %v, want to contain %q", status.Modified, "README.md")
	}
	if !containsFile(status.Untracked, "untracked.txt") {
		t.Errorf("Untracked = %v, want to contain %q", status.Untracked, "untracked.txt")
	}
}

func TestStatus_NoUpstream(t *testing.T) {
	dir := initTestRepo(t) // no remote configured
	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	status, err := repo.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}

	if status.Ahead != 0 {
		t.Errorf("Ahead = %d, want 0 (no upstream)", status.Ahead)
	}
	if status.Behind != 0 {
		t.Errorf("Behind = %d, want 0 (no upstream)", status.Behind)
	}
}

func TestStatus_AheadBehind(t *testing.T) {
	localDir, remoteDir := initTestRepoWithRemote(t)

	// Push a commit from another clone to create "behind".
	otherDir := t.TempDir()
	runGit(t, otherDir, "clone", remoteDir, ".")
	runGit(t, otherDir, "config", "user.email", "other@example.com")
	runGit(t, otherDir, "config", "user.name", "Other User")
	writeTestFile(t, filepath.Join(otherDir, "remote-file.txt"), "remote content\n")
	runGit(t, otherDir, "add", ".")
	runGit(t, otherDir, "commit", "-m", "Remote commit")
	runGit(t, otherDir, "push", "origin", "main")

	// Create local commits to be "ahead".
	writeTestFile(t, filepath.Join(localDir, "local1.txt"), "local 1\n")
	runGit(t, localDir, "add", ".")
	runGit(t, localDir, "commit", "-m", "Local commit 1")
	writeTestFile(t, filepath.Join(localDir, "local2.txt"), "local 2\n")
	runGit(t, localDir, "add", ".")
	runGit(t, localDir, "commit", "-m", "Local commit 2")

	// Fetch to update remote tracking refs.
	runGit(t, localDir, "fetch", "origin")

	repo, err := NewRepository(localDir)
	if err != nil {
		t.Fatal(err)
	}

	status, err := repo.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}

	if status.Ahead != 2 {
		t.Errorf("Ahead = %d, want 2", status.Ahead)
	}
	if status.Behind != 1 {
		t.Errorf("Behind = %d, want 1", status.Behind)
	}
}

func TestLog(t *testing.T) {
	dir := initTestRepo(t)

	// Create additional commits.
	for i := range 4 {
		writeTestFile(t, filepath.Join(dir, "file.txt"), "content "+string(rune('A'+i))+"\n")
		runGit(t, dir, "add", ".")
		runGit(t, dir, "commit", "-m", "Commit "+string(rune('A'+i)))
	}
	// Now have 5 commits total (initial + 4).

	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("request 3 of 5", func(t *testing.T) {
		commits, err := repo.Log(3)
		if err != nil {
			t.Fatalf("Log(3) error: %v", err)
		}
		if len(commits) != 3 {
			t.Fatalf("Log(3) returned %d commits, want 3", len(commits))
		}
		// Newest first.
		if commits[0].Message != "Commit D" {
			t.Errorf("commits[0].Message = %q, want %q", commits[0].Message, "Commit D")
		}
		// Each commit must have non-empty fields.
		for i, c := range commits {
			if c.Hash == "" {
				t.Errorf("commits[%d].Hash is empty", i)
			}
			if c.Author == "" {
				t.Errorf("commits[%d].Author is empty", i)
			}
			if c.Date.IsZero() {
				t.Errorf("commits[%d].Date is zero", i)
			}
			if c.Message == "" {
				t.Errorf("commits[%d].Message is empty", i)
			}
		}
	})

	t.Run("request more than available", func(t *testing.T) {
		commits, err := repo.Log(100)
		if err != nil {
			t.Fatalf("Log(100) error: %v", err)
		}
		if len(commits) != 5 {
			t.Errorf("Log(100) returned %d commits, want 5", len(commits))
		}
	})
}

func TestDiff(t *testing.T) {
	dir := initTestRepo(t)

	// Create a branch with a different file.
	runGit(t, dir, "branch", "feature")
	runGit(t, dir, "checkout", "feature")
	writeTestFile(t, filepath.Join(dir, "feature.txt"), "feature content\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Feature commit")
	runGit(t, dir, "checkout", "main")

	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	diff, err := repo.Diff("main", "feature")
	if err != nil {
		t.Fatalf("Diff() error: %v", err)
	}

	if diff == "" {
		t.Error("Diff() returned empty string, want non-empty diff")
	}
	if !containsString(diff, "feature.txt") {
		t.Error("Diff() does not contain the modified file path")
	}
}

func TestIsClean_True(t *testing.T) {
	dir := initTestRepo(t)
	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	clean, err := repo.IsClean()
	if err != nil {
		t.Fatalf("IsClean() error: %v", err)
	}
	if !clean {
		t.Error("IsClean() = false, want true")
	}
}

func TestIsClean_False(t *testing.T) {
	dir := initTestRepo(t)
	writeTestFile(t, filepath.Join(dir, "README.md"), "modified\n")

	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	clean, err := repo.IsClean()
	if err != nil {
		t.Fatalf("IsClean() error: %v", err)
	}
	if clean {
		t.Error("IsClean() = true, want false")
	}
}

func TestRoot(t *testing.T) {
	dir := initTestRepo(t)
	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatal(err)
	}

	got := repo.Root()
	want := filepath.Clean(dir)
	if got != want {
		t.Errorf("Root() = %q, want %q", got, want)
	}
}

// containsFile checks if a slice of filenames contains the given name.
func containsFile(files []string, name string) bool {
	return slices.Contains(files, name)
}

// containsString checks if a string contains a substring.
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && contains(s, substr)
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
