package ops

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGitManager_GetBranch(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	branch := mgr.GetBranch()
	if branch != "main" {
		t.Errorf("GetBranch() = %q, want %q", branch, "main")
	}
}

func TestGitManager_GetLastCommit(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	commit := mgr.GetLastCommit()
	if len(commit) != 40 {
		t.Errorf("GetLastCommit() = %q, expected 40-char hash", commit)
	}
}

func TestGitManager_GetStatus(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// Clean repo
	status := mgr.GetStatus()
	if status != "" {
		t.Errorf("GetStatus() = %q, want empty for clean repo", status)
	}

	// Add untracked file
	testFile := filepath.Join(dir, "untracked.txt")
	if err := os.WriteFile(testFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	// Clear cache to see the change
	mgr.ClearCache(OpStatus)

	status = mgr.GetStatus()
	if status == "" {
		t.Error("GetStatus() should return non-empty with untracked file")
	}
}

func TestGitManager_GetChangeCount(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// Clean repo
	count := mgr.GetChangeCount()
	if count != 0 {
		t.Errorf("GetChangeCount() = %d, want 0 for clean repo", count)
	}

	// Add files
	for i := range 3 {
		testFile := filepath.Join(dir, "file"+string(rune('a'+i))+".txt")
		if err := os.WriteFile(testFile, []byte("content\n"), 0644); err != nil {
			t.Fatalf("write test file: %v", err)
		}
	}

	// Clear cache to see the change
	mgr.ClearCache(OpStatus)

	count = mgr.GetChangeCount()
	if count != 3 {
		t.Errorf("GetChangeCount() = %d, want 3", count)
	}
}

func TestGitManager_GetRemotes(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// No remotes initially
	remotes := mgr.GetRemotes()
	if len(remotes) != 0 {
		t.Errorf("GetRemotes() = %v, want empty", remotes)
	}

	// Add a remote
	runGit(t, dir, "remote", "add", "origin", "https://example.com/repo.git")

	// Clear cache to see the change
	mgr.ClearCache(OpRemote)

	remotes = mgr.GetRemotes()
	if len(remotes) != 1 || remotes[0] != "origin" {
		t.Errorf("GetRemotes() = %v, want [origin]", remotes)
	}
}

func TestGitManager_GetConfig(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	name := mgr.GetConfig("user.name")
	if name != "Test User" {
		t.Errorf("GetConfig(user.name) = %q, want %q", name, "Test User")
	}

	email := mgr.GetConfig("user.email")
	if email != "test@example.com" {
		t.Errorf("GetConfig(user.email) = %q, want %q", email, "test@example.com")
	}

	// Non-existent config
	nonexistent := mgr.GetConfig("nonexistent.key")
	if nonexistent != "" {
		t.Errorf("GetConfig(nonexistent.key) = %q, want empty", nonexistent)
	}
}

func TestGitManager_HasUncommittedChanges(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// Clean repo
	if mgr.HasUncommittedChanges() {
		t.Error("HasUncommittedChanges() = true, want false for clean repo")
	}

	// Add untracked file
	testFile := filepath.Join(dir, "untracked.txt")
	if err := os.WriteFile(testFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	// Clear cache to see the change
	mgr.ClearCache(OpStatus)

	if !mgr.HasUncommittedChanges() {
		t.Error("HasUncommittedChanges() = false, want true with untracked file")
	}
}

func TestGitManager_GetCommitCount(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	count := mgr.GetCommitCount()
	if count < 1 {
		t.Errorf("GetCommitCount() = %d, want at least 1", count)
	}

	// Add another commit
	testFile := filepath.Join(dir, "new.txt")
	if err := os.WriteFile(testFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Second commit")

	// Clear cache to see the new commit
	mgr.ClearCache(OpLog)

	count2 := mgr.GetCommitCount()
	if count2 != count+1 {
		t.Errorf("GetCommitCount() = %d after second commit, want %d", count2, count+1)
	}
}

func TestGitManager_GetDiff(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// Modify a file
	testFile := filepath.Join(dir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Modified\n"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	diff := mgr.GetDiff()
	if diff == "" {
		t.Error("GetDiff() should return non-empty with modified file")
	}
}

func TestGitManager_String(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	str := mgr.String()
	if str == "" {
		t.Error("String() should return non-empty")
	}
}

func TestGitManager_ExecuteRaw(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	result := mgr.ExecuteRaw([]string{"rev-parse", "HEAD"}, 5)
	if !result.Success {
		t.Errorf("ExecuteRaw failed: %s", result.Stderr)
	}
	if len(result.Stdout) != 40 {
		t.Errorf("ExecuteRaw returned unexpected output: %q", result.Stdout)
	}
}

func TestGitManager_IsInsideWorkTree(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	if !mgr.IsInsideWorkTree() {
		t.Error("IsInsideWorkTree() = false, want true inside git repo")
	}

	// Test outside git repo
	tmpDir := t.TempDir()
	config2 := DefaultConfig()
	config2.WorkDir = tmpDir
	mgr2 := NewGitManager(config2)
	defer mgr2.Shutdown()

	if mgr2.IsInsideWorkTree() {
		t.Error("IsInsideWorkTree() = true, want false outside git repo")
	}
}

func TestGitManager_GetRepoRoot(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	root := mgr.GetRepoRoot()
	if root == "" {
		t.Error("GetRepoRoot() should return non-empty")
	}
}

func TestGitManager_IsClean(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// Clean repo
	if !mgr.IsClean() {
		t.Error("IsClean() = false, want true for clean repo")
	}

	// Add untracked file
	testFile := filepath.Join(dir, "untracked.txt")
	if err := os.WriteFile(testFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	// Clear cache to see the change
	mgr.ClearCache(OpStatus)

	if mgr.IsClean() {
		t.Error("IsClean() = true, want false with untracked file")
	}
}

func TestGitManager_GetAheadBehind(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// No upstream, should return 0, 0
	ahead, behind := mgr.GetAheadBehind()
	if ahead != 0 || behind != 0 {
		t.Errorf("GetAheadBehind() = (%d, %d), want (0, 0) with no upstream", ahead, behind)
	}
}

func TestGitManager_BatchExecute(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	results := mgr.BatchExecute(
		GitCommand{OperationType: OpBranch, Args: []string{"--show-current"}},
		GitCommand{OperationType: OpStatus, Args: []string{"--porcelain"}},
	)

	if len(results) != 2 {
		t.Fatalf("BatchExecute returned %d results, want 2", len(results))
	}

	for i, r := range results {
		if !r.Success {
			t.Errorf("results[%d] failed: %s", i, r.Stderr)
		}
	}
}

func TestGitManager_GetCommitsSince(t *testing.T) {
	dir := initTestRepo(t)

	// Create a second commit
	testFile := filepath.Join(dir, "new.txt")
	if err := os.WriteFile(testFile, []byte("content\n"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Second commit")

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// Get initial commit hash
	initialCommit := runGit(t, dir, "rev-list", "--max-parents=0", "HEAD")

	commits := mgr.GetCommitsSince(initialCommit)
	if len(commits) != 1 {
		t.Errorf("GetCommitsSince() returned %d commits, want 1", len(commits))
	}
}

func TestGitManager_GetFileAtCommit(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// Get content of README.md at HEAD
	content := mgr.GetFileAtCommit("HEAD", "README.md")
	if content != "# Test" {
		t.Errorf("GetFileAtCommit() = %q, want %q", content, "# Test")
	}

	// Non-existent file
	content = mgr.GetFileAtCommit("HEAD", "nonexistent.txt")
	if content != "" {
		t.Errorf("GetFileAtCommit(nonexistent) = %q, want empty", content)
	}
}

func TestGitManager_MustExecute_Success(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	result := mgr.MustExecute(GitCommand{
		OperationType: OpBranch,
		Args:          []string{"--show-current"},
	})

	if !result.Success {
		t.Error("MustExecute should succeed")
	}
}

func TestGitManager_MustExecute_Panic(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustExecute should panic on failure")
		}
	}()

	mgr.MustExecute(GitCommand{
		OperationType: OpBranch,
		Args:          []string{"--invalid-flag"},
	})
}

func TestParseProjectInfo(t *testing.T) {
	results := []GitResult{
		{Success: true, OperationType: OpBranch, Stdout: "main"},
		{Success: true, OperationType: OpLog, Stdout: "abc123def456789012345678901234567890abcd"},
		{Success: true, OperationType: OpLog, Stdout: "2024-01-15 10:30:00 +0900"},
		{Success: true, OperationType: OpStatus, Stdout: "M file1.txt\n?? file2.txt"},
	}

	info := ParseProjectInfo(results)

	if info.Branch != "main" {
		t.Errorf("Branch = %q, want %q", info.Branch, "main")
	}
	if info.LastCommit != "abc123def456789012345678901234567890abcd" {
		t.Errorf("LastCommit = %q, unexpected", info.LastCommit)
	}
	if info.CommitTime != "2024-01-15 10:30:00 +0900" {
		t.Errorf("CommitTime = %q, unexpected", info.CommitTime)
	}
	if info.Changes != 2 {
		t.Errorf("Changes = %d, want 2", info.Changes)
	}
}

func TestParseProjectInfo_FailedResults(t *testing.T) {
	results := []GitResult{
		{Success: false, OperationType: OpBranch},
		{Success: false, OperationType: OpLog},
	}

	info := ParseProjectInfo(results)

	if info.Branch != "" {
		t.Errorf("Branch should be empty for failed result")
	}
}

func TestExecuteParallelWithSemaphore(t *testing.T) {
	tasks := make([]func() int, 5)
	for i := range 5 {
		val := i
		tasks[i] = func() int {
			return val * 2
		}
	}

	results := ExecuteParallelWithSemaphore(tasks, 2)

	if len(results) != 5 {
		t.Fatalf("len(results) = %d, want 5", len(results))
	}

	for i, result := range results {
		if result != i*2 {
			t.Errorf("results[%d] = %d, want %d", i, result, i*2)
		}
	}
}

func TestExecuteParallelWithSemaphore_Empty(t *testing.T) {
	var tasks []func() int
	results := ExecuteParallelWithSemaphore(tasks, 2)
	if results != nil {
		t.Error("expected nil for empty tasks")
	}
}

func TestExecuteParallelWithSemaphore_ZeroConcurrency(t *testing.T) {
	tasks := []func() int{
		func() int { return 1 },
	}

	results := ExecuteParallelWithSemaphore(tasks, 0)
	if len(results) != 1 || results[0] != 1 {
		t.Errorf("unexpected results: %v", results)
	}
}

func TestNewGitManager_Defaults(t *testing.T) {
	// Test with zero/negative config values
	config := ManagerConfig{
		MaxWorkers:            0,
		CacheSizeLimit:        0,
		DefaultTTLSeconds:     0,
		DefaultTimeoutSeconds: 0,
		DefaultRetryCount:     -1,
	}

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	if mgr.config.MaxWorkers != 4 {
		t.Errorf("MaxWorkers = %d, want 4 (default)", mgr.config.MaxWorkers)
	}
	if mgr.config.CacheSizeLimit != 100 {
		t.Errorf("CacheSizeLimit = %d, want 100 (default)", mgr.config.CacheSizeLimit)
	}
}

func TestCache_CleanExpired(t *testing.T) {
	c := NewCache(100, 50*time.Millisecond)

	c.Set("key1", GitResult{Stdout: "1"}, 50*time.Millisecond)
	c.Set("key2", GitResult{Stdout: "2"}, 50*time.Millisecond)
	c.Set("key3", GitResult{Stdout: "3"}, 1*time.Hour)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	cleaned := c.CleanExpired()
	if cleaned != 2 {
		t.Errorf("CleanExpired() = %d, want 2", cleaned)
	}

	// key3 should still exist
	if _, hit := c.Get("key3"); !hit {
		t.Error("key3 should still be in cache")
	}
}
