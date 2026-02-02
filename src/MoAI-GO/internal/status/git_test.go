package status

import (
	"os"
	"os/exec"
	"testing"
	"time"
)

// --- FormatTimeAgo ---

func TestFormatTimeAgo(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "just now (0 seconds)",
			duration: 0,
			want:     "just now",
		},
		{
			name:     "just now (30 seconds)",
			duration: 30 * time.Second,
			want:     "just now",
		},
		{
			name:     "1 minute",
			duration: 1 * time.Minute,
			want:     "1m ago",
		},
		{
			name:     "5 minutes",
			duration: 5 * time.Minute,
			want:     "5m ago",
		},
		{
			name:     "59 minutes",
			duration: 59 * time.Minute,
			want:     "59m ago",
		},
		{
			name:     "1 hour",
			duration: 1 * time.Hour,
			want:     "1h ago",
		},
		{
			name:     "23 hours",
			duration: 23 * time.Hour,
			want:     "23h ago",
		},
		{
			name:     "1 day",
			duration: 24 * time.Hour,
			want:     "1d ago",
		},
		{
			name:     "7 days",
			duration: 7 * 24 * time.Hour,
			want:     "7d ago",
		},
		{
			name:     "30 days",
			duration: 30 * 24 * time.Hour,
			want:     "30d ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a time in the past
			pastTime := time.Now().Add(-tt.duration)
			got := FormatTimeAgo(pastTime)
			if got != tt.want {
				t.Errorf("FormatTimeAgo() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- GitInfo struct ---

func TestGitInfoStruct(t *testing.T) {
	info := &GitInfo{
		Branch:     "main",
		IsClean:    true,
		HasChanges: false,
	}

	if info.Branch != "main" {
		t.Errorf("Branch = %q, want %q", info.Branch, "main")
	}
	if !info.IsClean {
		t.Error("IsClean should be true")
	}
	if info.HasChanges {
		t.Error("HasChanges should be false")
	}
}

func TestGitInfoStruct_DirtyState(t *testing.T) {
	info := &GitInfo{
		Branch:     "feature/test",
		IsClean:    false,
		HasChanges: true,
	}

	if info.Branch != "feature/test" {
		t.Errorf("Branch = %q, want %q", info.Branch, "feature/test")
	}
	if info.IsClean {
		t.Error("IsClean should be false")
	}
	if !info.HasChanges {
		t.Error("HasChanges should be true")
	}
}

// --- CommitInfo struct ---

func TestCommitInfoStruct(t *testing.T) {
	commit := CommitInfo{
		Hash:    "abc1234",
		Message: "feat: add new feature",
		TimeAgo: "2h ago",
	}

	if commit.Hash != "abc1234" {
		t.Errorf("Hash = %q, want %q", commit.Hash, "abc1234")
	}
	if commit.Message != "feat: add new feature" {
		t.Errorf("Message = %q", commit.Message)
	}
	if commit.TimeAgo != "2h ago" {
		t.Errorf("TimeAgo = %q, want %q", commit.TimeAgo, "2h ago")
	}
}

// --- GetGitInfo ---

// initTempGitRepo creates a temp directory with a git repo and one commit.
func initTempGitRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "checkout", "-b", "main"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("cmd %v failed: %v\n%s", args, err, out)
		}
	}

	// Create a file and commit it
	if err := os.WriteFile(tmpDir+"/README.md", []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = tmpDir
	if out, err := addCmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}
	commitCmd := exec.Command("git", "commit", "-m", "feat: initial commit")
	commitCmd.Dir = tmpDir
	if out, err := commitCmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, out)
	}

	return tmpDir
}

func TestGetGitInfo_InGitRepo(t *testing.T) {
	tmpDir := initTempGitRepo(t)

	// GetGitInfo uses cwd, so we need to change to the temp dir
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()

	info, err := GetGitInfo()
	if err != nil {
		t.Fatalf("GetGitInfo() error = %v", err)
	}

	if info == nil {
		t.Fatal("GetGitInfo returned nil")
	}
	if info.Branch != "main" {
		t.Errorf("Branch = %q, want %q", info.Branch, "main")
	}
	if !info.IsClean {
		t.Error("expected IsClean = true for clean repo")
	}
	if info.HasChanges {
		t.Error("expected HasChanges = false for clean repo")
	}
}

func TestGetGitInfo_DirtyRepo(t *testing.T) {
	tmpDir := initTempGitRepo(t)

	// Make the repo dirty
	if err := os.WriteFile(tmpDir+"/dirty.txt", []byte("dirty"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()

	info, err := GetGitInfo()
	if err != nil {
		t.Fatalf("GetGitInfo() error = %v", err)
	}

	if info.IsClean {
		t.Error("expected IsClean = false for dirty repo")
	}
	if !info.HasChanges {
		t.Error("expected HasChanges = true for dirty repo")
	}
}

func TestGetGitInfo_NotAGitRepo(t *testing.T) {
	tmpDir := t.TempDir() // No git init

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()

	_, err = GetGitInfo()
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

// --- GetRecentCommits ---

func TestGetRecentCommits_InGitRepo(t *testing.T) {
	tmpDir := initTempGitRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()

	commits, err := GetRecentCommits(5)
	if err != nil {
		t.Fatalf("GetRecentCommits() error = %v", err)
	}

	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}

	if commits[0].Message != "feat: initial commit" {
		t.Errorf("commit message = %q, want %q", commits[0].Message, "feat: initial commit")
	}
	if commits[0].Hash == "" {
		t.Error("commit hash is empty")
	}
}

func TestGetRecentCommits_NotAGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()

	_, err = GetRecentCommits(5)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}
