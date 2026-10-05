package spec

import (
	"os/exec"
	"strings"
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
//   - "main-branch": the ordered mainBranchCandidates resolution walk
//     (local main → origin/main → local master → origin/master → unresolvable)
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

	// mainBranch: the first resolvable candidate of the mainBranchCandidates
	// chain (mainBranchUnresolvable when none resolves). Keyed by
	// "main-branch" (query-kind per R1); the unresolvable outcome is cached
	// exactly like a hit (REQ-SLGB-007).
	mainBranch    string
	mainBranchSet bool

	// shallowResult: result of git rev-parse --is-shallow-repository
	// (true = shallow clone). Keyed by "shallow" (query-kind per R1,
	// SPEC-SPECLINT-GITBLIND-001 M1): repository-level predicate, computed
	// once per run so the per-SPEC shape ②/③ decision never spawns a
	// per-SPEC subprocess.
	shallowResult bool
	shallowCached bool

	// unreachableEmitted records whether the StatusGitUnreachable finding
	// has already been emitted in this run (SPEC-SPECLINT-GITBLIND-001
	// §2.2): the emission cap is one per Lint() run, because the cause is
	// repository-wide and a per-SPEC flood would bury the signal in the very
	// CI state the rule exists to expose.
	unreachableEmitted bool

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
//
//	exec.Command("git", "rev-parse", "--git-dir").Output() → err == nil
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

// mainBranchCandidates is the ordered base-ref resolution chain
// (REQ-SLGB-006, SPEC-SPECLINT-GITBLIND-001 M2): local main → origin/main →
// local master → origin/master. Single source for the resolution walk AND
// the StatusGitUnreachable message (REQ-SLGB-002) — no scattered literals.
var mainBranchCandidates = []string{"main", "origin/main", "master", "origin/master"}

// mainBranchUnresolvable is the value cachedMainBranch returns when no
// candidate in mainBranchCandidates resolves. The empty string names no ref,
// so consumers that hand it to git fail visibly instead of silently querying
// a ref that does not exist — the pre-M2 "master" literal did exactly that.
const mainBranchUnresolvable = ""

// cachedMainBranch resolves the base branch through the ordered
// mainBranchCandidates chain, using the per-run cache if available. When no
// candidate resolves it returns mainBranchUnresolvable — never a literal
// naming a nonexistent ref (REQ-SLGB-006).
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
		branch := resolveMainBranch()
		c.mainBranch = branch
		c.mainBranchSet = true
		gitQueryCacheMu.Unlock()
		return branch
	}

	// No cache — direct resolution (original behavior for non-Lint callers)
	return resolveMainBranch()
}

// resolveMainBranch walks mainBranchCandidates in order and returns the
// first ref `git rev-parse --verify` confirms. REQ-SLGB-007: the
// unresolvable outcome is cached by the caller exactly like a hit — one walk
// per Lint() run, never one per SPEC (4 rev-parse spawns × N SPECs).
func resolveMainBranch() string {
	for _, candidate := range mainBranchCandidates {
		if _, err := exec.Command("git", "rev-parse", "--verify", candidate).Output(); err == nil {
			return candidate
		}
	}
	return mainBranchUnresolvable
}

// cachedIsShallowRepository reports whether the repository at the process
// working directory is a shallow clone, using the per-run cache if active.
// Equivalent to:
//
//	exec.Command("git", "rev-parse", "--is-shallow-repository").Output()
//	  → err == nil && strings.TrimSpace(stdout) == "true"
//
// A git failure (not a repository, git unavailable) reports false: that state
// surfaces through the shape ① error path (git log failed) instead, which
// fires unconditionally, so a false shallow answer cannot mask blindness.
//
// SPEC-SPECLINT-GITBLIND-001 M1 (REQ-SLGB-003/004): repository-level
// predicate deciding whether shapes ②/③ count as observation failures.
func cachedIsShallowRepository() bool {
	gitQueryCacheMu.RLock()
	c := gitQueryCacheV
	gitQueryCacheMu.RUnlock()

	if c != nil {
		gitQueryCacheMu.Lock()
		if c.shallowCached {
			result := c.shallowResult
			gitQueryCacheMu.Unlock()
			return result
		}
		result := queryIsShallowRepository()
		c.shallowResult = result
		c.shallowCached = true
		gitQueryCacheMu.Unlock()
		return result
	}

	// No cache — direct check (non-Lint callers).
	return queryIsShallowRepository()
}

// queryIsShallowRepository performs the raw shallow check.
func queryIsShallowRepository() bool {
	out, err := exec.Command("git", "rev-parse", "--is-shallow-repository").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

// takeUnreachableEmission records that StatusGitUnreachable is being emitted
// for this run and reports whether this call holds the run's single emission
// slot (true = emit; false = already emitted, suppress). With no per-run
// cache active (any caller outside Lint()) there is no suppression state, so
// it always reports true — the deliberate no-cache-path behavior of
// SPEC-SPECLINT-GITBLIND-001 §2.2.
func takeUnreachableEmission() bool {
	gitQueryCacheMu.RLock()
	c := gitQueryCacheV
	gitQueryCacheMu.RUnlock()

	if c == nil {
		return true
	}

	gitQueryCacheMu.Lock()
	defer gitQueryCacheMu.Unlock()
	if c.unreachableEmitted {
		return false
	}
	c.unreachableEmitted = true
	return true
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
