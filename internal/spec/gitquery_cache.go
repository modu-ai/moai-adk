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
//   - "transition": the most recent `status:` transition of one SPEC document,
//     keyed additionally by that document (see transitions below)
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

	// transitions memoizes the `git log --follow -p` history walk that finds a
	// SPEC's most recent `status:` transition, keyed per document. Two rules
	// ask the same question of the same document in one Lint() run —
	// OwnershipTransitionRule (*did the right agent sign it?*) and
	// StatusTransitionValidityRule (*is the pair itself legal?*) — and the walk
	// is by far the most expensive query in the package, so serving the second
	// asker from the first's result is what keeps corpus lint at one walk per
	// document rather than one per rule.
	transitions map[string]transitionMemo
}

// transitionMemo is one memoized history-walk outcome. Both return values are
// stored, so a lookup that failed is not silently retried per rule.
type transitionMemo struct {
	rec *ownershipTransitionRecord
	err error
}

// startGitQueryCache initializes a fresh per-run cache. Called at Lint() entry.
func startGitQueryCache() {
	gitQueryCacheMu.Lock()
	gitQueryCacheV = &gitEnvCache{transitions: make(map[string]transitionMemo)}
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

// cachedOwnershipTransition returns the most recent recorded `status:`
// transition for one SPEC document, memoized for the duration of a Lint() run.
//
// The lookup itself stays behind getOwnershipTransitionRunner, so the test
// injection hook keeps working and this function only decides whether the hook
// is called or its previous answer reused.
//
// When no cache is active (any caller outside Lint()), every call reaches the
// lookup — the pre-memoization behavior, preserved deliberately.
//
// @MX:ANCHOR: [AUTO] cachedOwnershipTransition — one history walk per document per Lint() run
// @MX:REASON: the `git log --follow -p` walk is the package's most expensive query, and two registered rules read the same document's transition; without this memo the corpus pays one walk per rule per document
func cachedOwnershipTransition(specPath, specID string) (*ownershipTransitionRecord, error) {
	gitQueryCacheMu.RLock()
	c := gitQueryCacheV
	gitQueryCacheMu.RUnlock()

	if c == nil {
		// No cache — direct lookup (original behavior for non-Lint callers).
		return getOwnershipTransitionRunner(specPath, specID)
	}

	// The document, not the rule, is the unit of the answer. NUL separates the
	// two components so no (path, id) pair can spell another pair's key.
	key := specPath + "\x00" + specID

	gitQueryCacheMu.RLock()
	memo, ok := c.transitions[key]
	gitQueryCacheMu.RUnlock()
	if ok {
		return memo.rec, memo.err
	}

	rec, err := getOwnershipTransitionRunner(specPath, specID)

	gitQueryCacheMu.Lock()
	if c.transitions != nil {
		c.transitions[key] = transitionMemo{rec: rec, err: err}
	}
	gitQueryCacheMu.Unlock()

	return rec, err
}
