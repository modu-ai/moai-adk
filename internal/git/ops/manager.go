package ops

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Compile-time interface compliance check.
var _ GitOperationsManager = (*GitManager)(nil)

// GitManager implements GitOperationsManager with parallel execution,
// result caching, and connection pooling.
type GitManager struct {
	config  ManagerConfig
	cache   *Cache
	pool    *WorkerPool
	stats   *StatsTracker
	workDir string
}

// NewGitManager creates a new GitManager with the specified configuration.
func NewGitManager(config ManagerConfig) *GitManager {
	// Apply defaults
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = 4
	}
	if config.CacheSizeLimit <= 0 {
		config.CacheSizeLimit = 100
	}
	if config.DefaultTTLSeconds <= 0 {
		config.DefaultTTLSeconds = 60
	}
	if config.DefaultTimeoutSeconds <= 0 {
		config.DefaultTimeoutSeconds = 10
	}
	if config.DefaultRetryCount < 0 {
		config.DefaultRetryCount = 2
	}

	workDir := config.WorkDir
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	return &GitManager{
		config:  config,
		cache:   NewCache(config.CacheSizeLimit, time.Duration(config.DefaultTTLSeconds)*time.Second),
		pool:    NewWorkerPool(config.MaxWorkers),
		stats:   NewStatsTracker(),
		workDir: workDir,
	}
}

// ExecuteCommand executes a single Git command.
func (m *GitManager) ExecuteCommand(cmd GitCommand) GitResult {
	start := time.Now()

	// Apply defaults
	if cmd.TimeoutSeconds <= 0 {
		cmd.TimeoutSeconds = m.config.DefaultTimeoutSeconds
	}
	if cmd.RetryCount < 0 {
		cmd.RetryCount = m.config.DefaultRetryCount
	}
	workDir := cmd.WorkDir
	if workDir == "" {
		workDir = m.workDir
	}

	// Build command arguments
	args := m.buildArgs(cmd)
	fullCommand := append([]string{"git"}, args...)

	// Check cache if caching is enabled
	if cmd.CacheTTLSeconds >= 0 {
		branch := m.getCurrentBranch(workDir)
		cacheKey := GenerateCacheKey(cmd.OperationType, args, workDir, branch)

		if result, hit := m.cache.Get(cacheKey); hit {
			result.ExecutionTime = time.Since(start)
			result.Command = fullCommand
			m.stats.RecordOperation(result.ExecutionTime, true, false)
			m.updateCacheStats()
			return result
		}
	}

	// Execute the command with retry
	var result GitResult
	var lastErr error
	maxRetries := max(cmd.RetryCount, 0)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result = m.executeGit(workDir, args, cmd.TimeoutSeconds)
		if result.Success {
			break
		}
		lastErr = result.Error
	}

	result.ExecutionTime = time.Since(start)
	result.OperationType = cmd.OperationType
	result.Command = fullCommand

	// Cache the result if successful and caching is enabled
	if result.Success && cmd.CacheTTLSeconds >= 0 {
		ttl := time.Duration(cmd.CacheTTLSeconds) * time.Second
		if ttl <= 0 {
			ttl = time.Duration(m.config.DefaultTTLSeconds) * time.Second
		}
		branch := m.getCurrentBranch(workDir)
		cacheKey := GenerateCacheKey(cmd.OperationType, args, workDir, branch)
		result.Cached = true
		m.cache.Set(cacheKey, result, ttl)
	}

	m.stats.RecordOperation(result.ExecutionTime, false, lastErr != nil)
	m.updateCacheStats()

	return result
}

// ExecuteParallel executes multiple Git commands in parallel.
func (m *GitManager) ExecuteParallel(cmds []GitCommand) []GitResult {
	if len(cmds) == 0 {
		return nil
	}

	results := make([]GitResult, len(cmds))
	var wg sync.WaitGroup
	sem := make(chan struct{}, m.config.MaxWorkers)

	for i, cmd := range cmds {
		wg.Add(1)
		go func(idx int, c GitCommand) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release
			results[idx] = m.ExecuteCommand(c)
		}(i, cmd)
	}

	wg.Wait()
	return results
}

// GetProjectInfo returns comprehensive project information.
func (m *GitManager) GetProjectInfo() ProjectInfo {
	// Execute 4 commands in parallel
	cmds := []GitCommand{
		{OperationType: OpBranch, Args: []string{"--show-current"}, CacheTTLSeconds: 5},
		{OperationType: OpLog, Args: []string{"-1", "--format=%H"}, CacheTTLSeconds: 5},
		{OperationType: OpLog, Args: []string{"-1", "--format=%ci"}, CacheTTLSeconds: 5},
		{OperationType: OpStatus, Args: []string{"--porcelain"}, CacheTTLSeconds: 1},
	}

	results := m.ExecuteParallel(cmds)

	info := ProjectInfo{
		FetchTime: time.Now(),
	}

	if len(results) >= 1 && results[0].Success {
		info.Branch = strings.TrimSpace(results[0].Stdout)
	}
	if len(results) >= 2 && results[1].Success {
		info.LastCommit = strings.TrimSpace(results[1].Stdout)
	}
	if len(results) >= 3 && results[2].Success {
		info.CommitTime = strings.TrimSpace(results[2].Stdout)
	}
	if len(results) >= 4 && results[3].Success {
		// Count changed files
		lines := strings.Split(strings.TrimSpace(results[3].Stdout), "\n")
		if results[3].Stdout != "" {
			info.Changes = len(lines)
		}
	}

	return info
}

// GetStatistics returns performance and cache statistics.
func (m *GitManager) GetStatistics() Statistics {
	m.updateCacheStats()
	return m.stats.GetStats()
}

// ClearCache clears cache entries for a specific operation type.
func (m *GitManager) ClearCache(opType GitOperationType) int {
	return m.cache.Clear(opType)
}

// Shutdown gracefully shuts down the manager.
func (m *GitManager) Shutdown() {
	m.pool.Shutdown()
}

// buildArgs builds the git command arguments from a GitCommand.
func (m *GitManager) buildArgs(cmd GitCommand) []string {
	args := []string{string(cmd.OperationType)}
	args = append(args, cmd.Args...)
	return args
}

// executeGit executes a git command and returns the result.
func (m *GitManager) executeGit(workDir string, args []string, timeoutSeconds int) GitResult {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := GitResult{
		Stdout: strings.TrimSpace(stdout.String()),
		Stderr: strings.TrimSpace(stderr.String()),
	}

	if err != nil {
		result.Success = false
		result.Error = err
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ReturnCode = exitErr.ExitCode()
		} else {
			result.ReturnCode = 1
		}
	} else {
		result.Success = true
		result.ReturnCode = 0
	}

	return result
}

// getCurrentBranch returns the current branch name.
func (m *GitManager) getCurrentBranch(workDir string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = workDir

	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// updateCacheStats updates the statistics with current cache info.
func (m *GitManager) updateCacheStats() {
	m.stats.SetCacheStats(m.cache.Stats())
	m.stats.SetPending(m.pool.Pending())
}

// ExecuteRaw executes a raw git command string.
// This is useful for complex commands that don't fit the GitCommand structure.
func (m *GitManager) ExecuteRaw(args []string, timeoutSeconds int) GitResult {
	if timeoutSeconds <= 0 {
		timeoutSeconds = m.config.DefaultTimeoutSeconds
	}

	start := time.Now()
	result := m.executeGit(m.workDir, args, timeoutSeconds)
	result.ExecutionTime = time.Since(start)
	result.Command = append([]string{"git"}, args...)

	m.stats.RecordOperation(result.ExecutionTime, false, !result.Success)
	return result
}

// GetBranch returns the current branch name.
func (m *GitManager) GetBranch() string {
	result := m.ExecuteCommand(GitCommand{
		OperationType:   OpBranch,
		Args:            []string{"--show-current"},
		CacheTTLSeconds: 5,
	})
	if result.Success {
		return strings.TrimSpace(result.Stdout)
	}
	return ""
}

// GetLastCommit returns the last commit hash.
func (m *GitManager) GetLastCommit() string {
	result := m.ExecuteCommand(GitCommand{
		OperationType:   OpLog,
		Args:            []string{"-1", "--format=%H"},
		CacheTTLSeconds: 5,
	})
	if result.Success {
		return strings.TrimSpace(result.Stdout)
	}
	return ""
}

// GetStatus returns the current status.
func (m *GitManager) GetStatus() string {
	result := m.ExecuteCommand(GitCommand{
		OperationType:   OpStatus,
		Args:            []string{"--porcelain"},
		CacheTTLSeconds: 1,
	})
	if result.Success {
		return result.Stdout
	}
	return ""
}

// GetChangeCount returns the number of changed files.
func (m *GitManager) GetChangeCount() int {
	status := m.GetStatus()
	if status == "" {
		return 0
	}
	lines := strings.Split(status, "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

// GetRemotes returns the list of remote names.
func (m *GitManager) GetRemotes() []string {
	result := m.ExecuteCommand(GitCommand{
		OperationType:   OpRemote,
		Args:            []string{},
		CacheTTLSeconds: 30,
	})
	if !result.Success || result.Stdout == "" {
		return nil
	}
	return strings.Split(result.Stdout, "\n")
}

// GetConfig returns a git config value.
func (m *GitManager) GetConfig(key string) string {
	result := m.ExecuteCommand(GitCommand{
		OperationType:   OpConfig,
		Args:            []string{key},
		CacheTTLSeconds: 60,
	})
	if result.Success {
		return strings.TrimSpace(result.Stdout)
	}
	return ""
}

// HasUncommittedChanges returns true if there are uncommitted changes.
func (m *GitManager) HasUncommittedChanges() bool {
	return m.GetChangeCount() > 0
}

// GetCommitCount returns the total number of commits.
func (m *GitManager) GetCommitCount() int {
	result := m.ExecuteCommand(GitCommand{
		OperationType:   OpLog,
		Args:            []string{"--oneline"},
		CacheTTLSeconds: 30,
	})
	if !result.Success || result.Stdout == "" {
		return 0
	}
	lines := strings.Split(result.Stdout, "\n")
	return len(lines)
}

// GetDiff returns the diff output.
func (m *GitManager) GetDiff(args ...string) string {
	result := m.ExecuteCommand(GitCommand{
		OperationType:   OpDiff,
		Args:            args,
		CacheTTLSeconds: 1,
	})
	if result.Success {
		return result.Stdout
	}
	return ""
}

// String returns a string representation of the manager status.
func (m *GitManager) String() string {
	stats := m.GetStatistics()
	return fmt.Sprintf("GitManager(workers=%d, cache=%d/%d, ops=%d, hitRate=%.1f%%)",
		m.config.MaxWorkers,
		stats.Cache.Size,
		stats.Cache.SizeLimit,
		stats.Operations.Total,
		stats.Operations.CacheHitRate*100)
}

// formatDuration formats a duration for display.
// Currently unused but kept for future debugging/logging use.
func formatDuration(ns int64) string { //nolint:unused
	d := time.Duration(ns)
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// ParseProjectInfo parses project info from parallel command results.
func ParseProjectInfo(results []GitResult) ProjectInfo {
	info := ProjectInfo{
		FetchTime: time.Now(),
	}

	for _, r := range results {
		if !r.Success {
			continue
		}
		switch r.OperationType {
		case OpBranch:
			info.Branch = strings.TrimSpace(r.Stdout)
		case OpLog:
			// Check if it's a commit hash or time
			out := strings.TrimSpace(r.Stdout)
			if len(out) == 40 { // SHA-1 hash
				info.LastCommit = out
			} else if strings.Contains(out, "-") { // Likely a date
				info.CommitTime = out
			}
		case OpStatus:
			lines := strings.Split(strings.TrimSpace(r.Stdout), "\n")
			if r.Stdout != "" {
				info.Changes = len(lines)
			}
		}
	}

	return info
}

// BatchExecute is a convenience function for executing multiple commands.
func (m *GitManager) BatchExecute(commands ...GitCommand) []GitResult {
	return m.ExecuteParallel(commands)
}

// MustExecute executes a command and panics on failure.
// Use only in situations where failure is unrecoverable.
func (m *GitManager) MustExecute(cmd GitCommand) GitResult {
	result := m.ExecuteCommand(cmd)
	if !result.Success {
		panic(fmt.Sprintf("git command failed: %s", result.Stderr))
	}
	return result
}

// GetCommitsSince returns commits since a given reference.
func (m *GitManager) GetCommitsSince(ref string) []string {
	result := m.ExecuteCommand(GitCommand{
		OperationType:   OpLog,
		Args:            []string{"--oneline", ref + "..HEAD"},
		CacheTTLSeconds: 5,
	})
	if !result.Success || result.Stdout == "" {
		return nil
	}
	return strings.Split(result.Stdout, "\n")
}

// GetFileAtCommit returns the content of a file at a specific commit.
func (m *GitManager) GetFileAtCommit(commit, filePath string) string {
	result := m.ExecuteRaw([]string{"show", commit + ":" + filePath}, 5)
	if result.Success {
		return result.Stdout
	}
	return ""
}

// IsInsideWorkTree returns true if the current directory is inside a git work tree.
func (m *GitManager) IsInsideWorkTree() bool {
	result := m.ExecuteRaw([]string{"rev-parse", "--is-inside-work-tree"}, 2)
	return result.Success && strings.TrimSpace(result.Stdout) == "true"
}

// GetRepoRoot returns the root directory of the repository.
func (m *GitManager) GetRepoRoot() string {
	result := m.ExecuteRaw([]string{"rev-parse", "--show-toplevel"}, 2)
	if result.Success {
		return strings.TrimSpace(result.Stdout)
	}
	return ""
}

// IsClean returns true if the working tree is clean.
func (m *GitManager) IsClean() bool {
	return m.GetChangeCount() == 0
}

// GetAheadBehind returns the number of commits ahead and behind the upstream.
func (m *GitManager) GetAheadBehind() (ahead, behind int) {
	result := m.ExecuteRaw([]string{"rev-list", "--left-right", "--count", "@{u}...HEAD"}, 5)
	if !result.Success {
		return 0, 0
	}
	parts := strings.Fields(result.Stdout)
	if len(parts) != 2 {
		return 0, 0
	}
	behind, _ = strconv.Atoi(parts[0])
	ahead, _ = strconv.Atoi(parts[1])
	return ahead, behind
}
