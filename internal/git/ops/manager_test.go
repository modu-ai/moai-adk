package ops

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// initTestRepo creates a temporary Git repository for testing.
func initTestRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	// Initialize git repo
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")

	// Create initial commit
	testFile := filepath.Join(dir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Initial commit")

	// Rename default branch to main if needed
	branch := strings.TrimSpace(runGit(t, dir, "branch", "--show-current"))
	if branch != "main" {
		runGit(t, dir, "branch", "-M", "main")
	}

	return dir
}

// runGit runs a git command and returns the output.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestNewGitManager(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	if mgr == nil {
		t.Fatal("NewGitManager returned nil")
	}
	defer mgr.Shutdown()

	// Verify configuration was applied
	if mgr.config.MaxWorkers != 4 {
		t.Errorf("MaxWorkers = %d, want 4", mgr.config.MaxWorkers)
	}
}

func TestGitManager_ExecuteCommand_Branch(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	cmd := GitCommand{
		OperationType: OpBranch,
		Args:          []string{"--show-current"},
	}

	result := mgr.ExecuteCommand(cmd)

	if !result.Success {
		t.Errorf("ExecuteCommand failed: %s", result.Stderr)
	}
	if strings.TrimSpace(result.Stdout) != "main" {
		t.Errorf("branch = %q, want %q", result.Stdout, "main")
	}
	if result.OperationType != OpBranch {
		t.Errorf("OperationType = %q, want %q", result.OperationType, OpBranch)
	}
}

func TestGitManager_ExecuteCommand_Status(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	cmd := GitCommand{
		OperationType: OpStatus,
		Args:          []string{"--porcelain"},
	}

	result := mgr.ExecuteCommand(cmd)

	if !result.Success {
		t.Errorf("ExecuteCommand failed: %s", result.Stderr)
	}
	// Clean repo should have empty status
	if result.Stdout != "" {
		t.Errorf("status = %q, want empty", result.Stdout)
	}
}

func TestGitManager_ExecuteCommand_Log(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	cmd := GitCommand{
		OperationType: OpLog,
		Args:          []string{"-1", "--oneline"},
	}

	result := mgr.ExecuteCommand(cmd)

	if !result.Success {
		t.Errorf("ExecuteCommand failed: %s", result.Stderr)
	}
	if !strings.Contains(result.Stdout, "Initial commit") {
		t.Errorf("log output = %q, want to contain 'Initial commit'", result.Stdout)
	}
}

func TestGitManager_ExecuteCommand_Caching(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir
	config.DefaultTTLSeconds = 60

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	cmd := GitCommand{
		OperationType:   OpBranch,
		Args:            []string{"--show-current"},
		CacheTTLSeconds: 60,
	}

	// First call should not be a cache hit
	result1 := mgr.ExecuteCommand(cmd)
	if !result1.Success {
		t.Fatalf("first ExecuteCommand failed: %s", result1.Stderr)
	}
	if result1.CacheHit {
		t.Error("first call should not be a cache hit")
	}

	// Second call should be a cache hit
	result2 := mgr.ExecuteCommand(cmd)
	if !result2.Success {
		t.Fatalf("second ExecuteCommand failed: %s", result2.Stderr)
	}
	if !result2.CacheHit {
		t.Error("second call should be a cache hit")
	}

	// Results should be the same
	if result1.Stdout != result2.Stdout {
		t.Errorf("results differ: %q vs %q", result1.Stdout, result2.Stdout)
	}
}

func TestGitManager_ExecuteCommand_CachingDisabled(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	cmd := GitCommand{
		OperationType:   OpBranch,
		Args:            []string{"--show-current"},
		CacheTTLSeconds: -1, // Disable caching
	}

	result1 := mgr.ExecuteCommand(cmd)
	result2 := mgr.ExecuteCommand(cmd)

	if result1.CacheHit || result2.CacheHit {
		t.Error("caching should be disabled when CacheTTLSeconds is negative")
	}
}

func TestGitManager_ExecuteCommand_Timeout(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir
	config.DefaultTimeoutSeconds = 1

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// Normal command should complete within timeout
	cmd := GitCommand{
		OperationType:  OpBranch,
		Args:           []string{"--show-current"},
		TimeoutSeconds: 1,
	}

	result := mgr.ExecuteCommand(cmd)
	if !result.Success {
		t.Errorf("command should succeed: %s", result.Stderr)
	}
}

func TestGitManager_ExecuteCommand_Retry(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir
	config.DefaultRetryCount = 2

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// Command that will fail (invalid git command)
	cmd := GitCommand{
		OperationType: OpBranch,
		Args:          []string{"--invalid-flag-that-does-not-exist"},
		RetryCount:    0, // No retries for faster test
	}

	result := mgr.ExecuteCommand(cmd)
	if result.Success {
		t.Error("command with invalid flag should fail")
	}
	if result.ReturnCode == 0 {
		t.Error("return code should be non-zero for failed command")
	}
}

func TestGitManager_ExecuteParallel(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	cmds := []GitCommand{
		{OperationType: OpBranch, Args: []string{"--show-current"}},
		{OperationType: OpStatus, Args: []string{"--porcelain"}},
		{OperationType: OpLog, Args: []string{"-1", "--oneline"}},
	}

	results := mgr.ExecuteParallel(cmds)

	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}

	// Check branch result
	if !results[0].Success || strings.TrimSpace(results[0].Stdout) != "main" {
		t.Errorf("branch result: success=%v, stdout=%q", results[0].Success, results[0].Stdout)
	}

	// Check status result
	if !results[1].Success {
		t.Errorf("status result failed: %s", results[1].Stderr)
	}

	// Check log result
	if !results[2].Success || !strings.Contains(results[2].Stdout, "Initial commit") {
		t.Errorf("log result: success=%v, stdout=%q", results[2].Success, results[2].Stdout)
	}
}

func TestGitManager_ExecuteParallel_OrderPreserved(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// Commands with different execution times
	cmds := []GitCommand{
		{OperationType: OpLog, Args: []string{"-1", "--format=%H"}},
		{OperationType: OpBranch, Args: []string{"--show-current"}},
		{OperationType: OpConfig, Args: []string{"user.name"}},
	}

	results := mgr.ExecuteParallel(cmds)

	// Results should be in the same order as commands
	if results[0].OperationType != OpLog {
		t.Errorf("results[0].OperationType = %q, want OpLog", results[0].OperationType)
	}
	if results[1].OperationType != OpBranch {
		t.Errorf("results[1].OperationType = %q, want OpBranch", results[1].OperationType)
	}
	if results[2].OperationType != OpConfig {
		t.Errorf("results[2].OperationType = %q, want OpConfig", results[2].OperationType)
	}
}

func TestGitManager_GetProjectInfo(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	info := mgr.GetProjectInfo()

	if info.Branch != "main" {
		t.Errorf("Branch = %q, want %q", info.Branch, "main")
	}
	if info.LastCommit == "" {
		t.Error("LastCommit should not be empty")
	}
	if info.CommitTime == "" {
		t.Error("CommitTime should not be empty")
	}
	if info.FetchTime.IsZero() {
		t.Error("FetchTime should not be zero")
	}
}

func TestGitManager_GetProjectInfo_WithChanges(t *testing.T) {
	dir := initTestRepo(t)

	// Create an untracked file
	testFile := filepath.Join(dir, "new_file.txt")
	if err := os.WriteFile(testFile, []byte("new content\n"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	info := mgr.GetProjectInfo()

	if info.Changes < 1 {
		t.Errorf("Changes = %d, want at least 1", info.Changes)
	}
}

func TestGitManager_GetStatistics(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// Execute some commands
	cmd := GitCommand{
		OperationType:   OpBranch,
		Args:            []string{"--show-current"},
		CacheTTLSeconds: 60,
	}
	mgr.ExecuteCommand(cmd) // First call (cache miss)
	mgr.ExecuteCommand(cmd) // Second call (cache hit)
	mgr.ExecuteCommand(cmd) // Third call (cache hit)

	stats := mgr.GetStatistics()

	if stats.Operations.Total != 3 {
		t.Errorf("Total = %d, want 3", stats.Operations.Total)
	}
	if stats.Operations.CacheHits != 2 {
		t.Errorf("CacheHits = %d, want 2", stats.Operations.CacheHits)
	}
	if stats.Operations.CacheMisses != 1 {
		t.Errorf("CacheMisses = %d, want 1", stats.Operations.CacheMisses)
	}

	// Cache hit rate should be ~66.67%
	expectedRate := 2.0 / 3.0
	if stats.Operations.CacheHitRate < expectedRate-0.01 || stats.Operations.CacheHitRate > expectedRate+0.01 {
		t.Errorf("CacheHitRate = %f, want ~%f", stats.Operations.CacheHitRate, expectedRate)
	}
}

func TestGitManager_ClearCache(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	// Execute commands to populate cache
	mgr.ExecuteCommand(GitCommand{OperationType: OpBranch, Args: []string{"--show-current"}, CacheTTLSeconds: 60})
	mgr.ExecuteCommand(GitCommand{OperationType: OpStatus, Args: []string{"--porcelain"}, CacheTTLSeconds: 60})

	// Clear only branch cache
	cleared := mgr.ClearCache(OpBranch)
	if cleared != 1 {
		t.Errorf("ClearCache returned %d, want 1", cleared)
	}

	// Branch should be re-executed (cache miss)
	result := mgr.ExecuteCommand(GitCommand{OperationType: OpBranch, Args: []string{"--show-current"}, CacheTTLSeconds: 60})
	if result.CacheHit {
		t.Error("branch should not be cached after clear")
	}

	// Status should still be cached
	result = mgr.ExecuteCommand(GitCommand{OperationType: OpStatus, Args: []string{"--porcelain"}, CacheTTLSeconds: 60})
	if !result.CacheHit {
		t.Error("status should still be cached")
	}
}

func TestGitManager_Shutdown(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)

	// Shutdown should be safe to call multiple times
	mgr.Shutdown()
	mgr.Shutdown()
}

func TestGitManager_ExecutionTime(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	cmd := GitCommand{
		OperationType: OpBranch,
		Args:          []string{"--show-current"},
	}

	result := mgr.ExecuteCommand(cmd)

	if result.ExecutionTime <= 0 {
		t.Error("ExecutionTime should be positive")
	}
	if result.ExecutionTime > 5*time.Second {
		t.Errorf("ExecutionTime = %v, too slow", result.ExecutionTime)
	}
}

func TestGitManager_CommandInResult(t *testing.T) {
	dir := initTestRepo(t)

	config := DefaultConfig()
	config.WorkDir = dir

	mgr := NewGitManager(config)
	defer mgr.Shutdown()

	cmd := GitCommand{
		OperationType: OpBranch,
		Args:          []string{"--show-current"},
	}

	result := mgr.ExecuteCommand(cmd)

	if len(result.Command) < 2 {
		t.Fatalf("Command should have at least 2 elements, got %v", result.Command)
	}
	if result.Command[0] != "git" {
		t.Errorf("Command[0] = %q, want %q", result.Command[0], "git")
	}
	if result.Command[1] != "branch" {
		t.Errorf("Command[1] = %q, want %q", result.Command[1], "branch")
	}
}
