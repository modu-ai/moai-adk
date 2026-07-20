package spec

import (
	"os/exec"
	"sync"
)

// gitQueryCache is a per-Lint() run cache that memoizes git environment
// checks (git rev-parse) to eliminate redundant subprocess spawns across
// N SPECs (REQ-PERF-001-A). The cache is scoped to a single Lint() execution:
// startGitQueryCache() creates it at Lint() entry, stopGitQueryCache()
// discards it at exit (per-run invalidation — REQ-PERF-001-B).
//
// Keyed by query-kind per plan §H R1 (cross-rule isolation):
//   - "git-dir": git rev-parse --git-dir (environment availability)
//   - "main-branch": git rev-parse --verify main (branch existence → "main" or "master")
//
// @MX:ANCHOR: [AUTO] gitQueryCache — REQ-PERF-001-A per-run git env cache
// @MX:REASON: eliminates ~2×N redundant git rev-parse spawns in moai spec lint (441 SPECs × 2 env checks → 2 total)
var (
	gitQueryCacheMu sync.RWMutex
	gitQueryCacheV  *gitEnvCache
)

// gitEnvCache stores cached git environment check results for one Lint() run.
type gitEnvCache struct {
	// gitDirResult: result of git rev-parse --git-dir (true = git available).
	// Keyed by "git-dir" (query-kind per R1).
	gitDirResult bool
	gitDirCached bool

	// mainBranch: "main" if git rev-parse --verify main succeeds, "master" otherwise.
	// Keyed by "main-branch" (query-kind per R1).
	mainBranch    string
	mainBranchSet bool
}

// startGitQueryCache initializes a fresh per-run cache. Called at Lint() entry.
func startGitQueryCache() {
	gitQueryCacheMu.Lock()
	gitQueryCacheV = &gitEnvCache{}
	gitQueryCacheMu.Unlock()
}

// stopGitQueryCache discards the per-run cache. Called at Lint() exit
// (deferred). REQ-PERF-001-B: per-run invalidation prevents stale results
// from leaking into the next Lint() invocation.
func stopGitQueryCache() {
	gitQueryCacheMu.Lock()
	gitQueryCacheV = nil
	gitQueryCacheMu.Unlock()
}

// cachedGitDirAvailable checks if git is available in the current directory,
// using the per-run cache if available. Equivalent to:
//   exec.Command("git", "rev-parse", "--git-dir").Output() → err == nil
//
// When no cache is active (outside Lint() context, e.g., direct calls from
// DetectDrift), performs the check directly (preserving original behavior).
func cachedGitDirAvailable() bool {
	gitQueryCacheMu.RLock()
	c := gitQueryCacheV
	gitQueryCacheMu.RUnlock()

	if c != nil {
		// Cache active — use memoized result or compute and cache
		gitQueryCacheMu.Lock()
		if c.gitDirCached {
			result := c.gitDirResult
			gitQueryCacheMu.Unlock()
			return result
		}
		_, err := exec.Command("git", "rev-parse", "--git-dir").Output()
		c.gitDirResult = err == nil
		c.gitDirCached = true
		result := c.gitDirResult
		gitQueryCacheMu.Unlock()
		return result
	}

	// No cache — direct check (original behavior for non-Lint callers)
	_, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	return err == nil
}

// cachedMainBranch determines the default branch name ("main" or "master"),
// using the per-run cache if available. Equivalent to checking
// git rev-parse --verify main, falling back to "master".
func cachedMainBranch() string {
	gitQueryCacheMu.RLock()
	c := gitQueryCacheV
	gitQueryCacheMu.RUnlock()

	if c != nil {
		gitQueryCacheMu.Lock()
		if c.mainBranchSet {
			result := c.mainBranch
			gitQueryCacheMu.Unlock()
			return result
		}
		branch := "main"
		if _, err := exec.Command("git", "rev-parse", "--verify", "main").Output(); err != nil {
			branch = "master"
		}
		c.mainBranch = branch
		c.mainBranchSet = true
		gitQueryCacheMu.Unlock()
		return branch
	}

	// No cache — direct check (original behavior)
	branch := "main"
	if _, err := exec.Command("git", "rev-parse", "--verify", "main").Output(); err != nil {
		branch = "master"
	}
	return branch
}
