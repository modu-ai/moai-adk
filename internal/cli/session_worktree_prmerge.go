package cli

// session_worktree_prmerge.go — SPEC-SESSION-WORKTREE-001 M8 PR-merge
// auto-cleanup.
//
// prMergeCleanup is the on-touch cleanup invoked at `moai session register`
// and `moai session list` BEFORE the subcommand's main work. It enumerates
// WT-* worktrees, detects whether each one's branch has been merged via PR
// (gh primary, git-branch-merged fallback), and removes the merged + clean
// ones — unsetting their safe.directory entry (R5 mitigation, reused from M7).
// Uncommitted changes are NEVER deleted (REQ-SW-024 dirty guard, reuses M4's
// worktreeIsDirty with an immediate-before-removal re-check for EC-11).
//
// Activation reuses the same `workflow.worktree.auto_cleanup` toggle as the
// M4 session-exit cleanup (REQ-SW-022) — turning it ON enables BOTH paths.
// OFF (the distributed default) → no PR-merge cleanup, byte-identical to the
// baseline.
//
// The PR-merge notice (PRMergeCleanupNoticePrefix) is DISTINCT from the
// session-exit notice (SessionExitCleanupNoticePrefix) so the two are
// attributable in combined output (EC-13).

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/session"
)

// PRMergeCleanupNoticePrefix is the literal prefix of the PR-merge removal
// notice. It MUST NOT collide with SessionExitCleanupNoticePrefix
// ("removed by session-exit cleanup:") so the two notices are distinguishable
// in combined output (REQ-SW-022 / AC-SW-022 / EC-13).
const PRMergeCleanupNoticePrefix = "removed by PR-merge cleanup:"

// Function-variable seams for M8 test injection. Each has a Real counterpart
// below; tests swap these via swapPRMergeSeams and restore on cleanup.
var (
	// sessionWorktreeGitWorktreeList runs `git worktree list --porcelain` and
	// returns its raw stdout. The porcelain format is parsed by
	// parseWorktreeList into (path, branch) pairs.
	sessionWorktreeGitWorktreeList = gitWorktreeListReal

	// sessionWorktreeGhLookPath reports whether `gh` is on PATH (BI-8). Absence
	// routes PR-merge detection to the git-branch-merged fallback
	// (REQ-SW-023), which is squash-merge blind.
	sessionWorktreeGhLookPath = ghLookPathReal

	// sessionWorktreeGhPRViewState runs `gh pr view <branch> --json state` and
	// returns the parsed `state` field (e.g. "MERGED", "OPEN", "CLOSED"). An
	// empty string means gh could not resolve the branch to a PR (no PR, or
	// gh error) — the worktree is NOT a cleanup candidate in that case. The
	// primary path sees squash merges (state == "MERGED") where the fallback
	// does not (AC-SW-023 primary-path correctness).
	sessionWorktreeGhPRViewState = ghPRViewStateReal

	// sessionWorktreeGitBranchMerged runs `git branch --merged origin/main` and
	// returns the list of branch names fully merged into origin/main. This is
	// the squash-merge-blind fallback (REQ-SW-023): squash-merged branches do
	// NOT appear in this list because their commits are not ancestors of
	// origin/main.
	sessionWorktreeGitBranchMerged = gitBranchMergedReal

	// sessionWorktreeGitStatusIgnored runs `git -C <path> status --porcelain
	// --ignored` and returns its raw stdout. It is DISTINCT from the M4
	// dirty-check seam (sessionWorktreeGitStatusPorcelain) on purpose: that one
	// is shared with the session-exit path, and widening it to --ignored would
	// silently impose this SPEC's ignored-content policy on that path too
	// (design.md §B.6a, retained rejection).
	sessionWorktreeGitStatusIgnored = gitStatusIgnoredReal
)

// wtEntry is a parsed worktree record from `git worktree list --porcelain`.
// A detached-HEAD worktree carries an empty branch.
//
// lock carries the git worktree lock state (SPEC-WORKTREE-REAPER-001
// REQ-WR-006). It is the AUTHORITATIVE anchor source: the porcelain already
// reports it, so capturing it here costs one parser case and makes both the
// anchor decision and the refusal-class pre-detection free of extra git calls.
type wtEntry struct {
	path   string
	branch string
	lock   session.LockInfo
}

// parseWorktreeList parses the `git worktree list --porcelain` output into
// worktree records. Each worktree entry begins with a `worktree <path>` line
// and carries at most one `branch refs/heads/<name>` line; detached entries
// have `detached` instead and yield an empty branch. A locked worktree carries
// a `locked <reason>` line, or a bare `locked` line when it was locked with no
// reason — `locked ` is git's own prefix and is NOT part of the stored reason
// (design.md §B.3). Entries are separated by blank lines.
func parseWorktreeList(porcelain string) []wtEntry {
	var entries []wtEntry
	var cur wtEntry
	haveCur := false
	flush := func() {
		if haveCur {
			entries = append(entries, cur)
			cur = wtEntry{}
			haveCur = false
		}
	}
	for _, line := range strings.Split(porcelain, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
			flush()
		case strings.HasPrefix(line, "worktree "):
			flush()
			cur.path = strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
			haveCur = true
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimSpace(strings.TrimPrefix(line, "branch "))
			// Strip the refs/heads/ prefix when present; bare branch names are
			// returned verbatim.
			cur.branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "detached":
			cur.branch = ""
		case line == "locked":
			cur.lock = session.LockInfo{Locked: true}
		case strings.HasPrefix(line, "locked "):
			cur.lock = session.LockInfo{Locked: true, Reason: strings.TrimPrefix(line, "locked ")}
		}
	}
	flush()
	return entries
}

// prMergeCleanup is the on-touch PR-merge auto-cleanup (M8,
// REQ-SW-022/023/024). Invoked at `moai session register` and `moai session
// list` RunE, gated by the AutoCleanup toggle. NEVER returns an error and
// NEVER aborts the caller — every failure path is a non-blocking notice
// (REQ-SW-004 fail-open spirit).
//
// @MX:ANCHOR: [AUTO] PR-merge auto-cleanup sweep (M8 on-touch at session register/list)
// @MX:REASON: REQ-SW-022/023/024 — same toggle as session-exit cleanup; gh primary + squash-blind fallback; dirty guard re-checked immediately before removal (EC-11); notice prefix MUST stay distinct from session-exit
func prMergeCleanup(cfg *config.Config, out io.Writer) {
	if cfg == nil || !cfg.Workflow.Worktree.AutoCleanup {
		// REQ-SW-022: reuses the session-exit AutoCleanup toggle. OFF (the
		// distributed default) → no PR-merge cleanup. Byte-identical to the
		// baseline: no notice, no git invocation, no observable side effect.
		return
	}
	// Enumerate all worktrees.
	porcelain, err := sessionWorktreeGitWorktreeList()
	if err != nil {
		// Fail-open: a worktree-list failure MUST NOT abort the on-touch
		// invocation.
		_, _ = fmt.Fprintf(out, "moai: PR-merge cleanup skipped (git worktree list failed: %v)\n", err)
		return
	}
	ghAvailable := sessionWorktreeGhLookPath()
	if !ghAvailable {
		// Document the squash-merge blindness once per invocation (REQ-SW-023).
		_, _ = fmt.Fprintln(out, "moai: PR-merge detection: gh not found; using git branch --merged fallback (squash-merge blind — squash-merged branches will NOT be detected)")
	}
	for _, e := range parseWorktreeList(porcelain) {
		// Only WT-* branches are session-worktree candidates
		// (SessionWorktreeBranchPrefix). Non-WT branches and detached entries
		// are ignored.
		if !strings.HasPrefix(e.branch, SessionWorktreeBranchPrefix) {
			continue
		}
		if !branchMergedForCleanup(e.branch, ghAvailable) {
			continue
		}
		// REQ-SW-024 / EC-11 dirty guard: re-check immediately before removal.
		// Reuses M4's worktreeIsDirty so one helper backs both call sites.
		dirty, derr := worktreeIsDirty(e.path)
		if derr != nil {
			_, _ = fmt.Fprintf(out, "moai: PR-merge cleanup skipped (cause=dirty-check-failed; dirty-check failed: %v): worktree %s preserved\n", derr, e.path)
			continue
		}
		if dirty {
			_, _ = fmt.Fprintf(out, "moai: PR-merge cleanup skipped (cause=dirty; uncommitted changes): worktree %s preserved (dispose manually via 'moai worktree remove' or 'git worktree remove')\n", e.path)
			continue
		}
		// t73 anchor guard: never remove a worktree a live session is
		// anchored in — the anchored session's shell dies with the tree
		// (Claude Code blocks all its Bash once the tree is gone). Same
		// immediately-before-removal position as the dirty guard (EC-11).
		//
		// The decision is the SHARED lock-∪-registry one (REQ-WR-009/019); the
		// registry alone was measured to name 1 of 5 live anchors.
		if v := session.AnchorDecision(e.path, e.lock, time.Now()); v.Anchored {
			_, _ = fmt.Fprintf(out, "moai: PR-merge cleanup skipped (cause=anchored-by-%s; live session anchored — %s): worktree %s preserved\n", v.Source, v.Detail, e.path)
			continue
		}
		// REQ-WR-021 pre-detection: a locked tree is refused by non-forced
		// `git worktree remove` whatever the lock's pid liveness, and the
		// condition never clears — so attempting it would emit an error-shaped
		// notice forever for a correctly-behaving tree. The sweep never
		// unlocks and never passes --force.
		if session.LockRefusesRemoval(e.lock) {
			_, _ = fmt.Fprintf(out, "moai: PR-merge cleanup skipped (cause=refusal-class; locked worktree — git refuses non-forced removal): worktree %s preserved\n", e.path)
			continue
		}
		if !ignoredContentAllowsRemoval(e.path, out) {
			continue
		}
		// Remove + unset safe.directory (R5, reused from M7). A removal failure
		// is non-blocking: the worktree is left on disk and the sweep
		// continues with the next candidate. This is REQ-WR-021's DEFINING
		// limb — every refusal cause outside the pre-detection set (a populated
		// submodule, EC-12) lands here, and git's own message is the cause token.
		if rerr := sessionWorktreeGitWorktreeRemove(e.path); rerr != nil {
			_, _ = fmt.Fprintf(out, "moai: PR-merge cleanup failed (cause=refusal-class; %v): worktree %s left on disk\n", rerr, e.path)
			continue
		}
		_ = sessionWorktreeGitSafeDirUnset(e.path)
		_, _ = fmt.Fprintf(out, "moai: %s [WT] worktree %s (branch %s merged)\n", PRMergeCleanupNoticePrefix, e.path, e.branch)
	}
}

// regenerableIgnoredPaths is the allowlist of gitignored paths whose loss
// costs nothing, enumerated FROM the measurement rather than invented ahead of
// it (design.md §A.7.3): runtime state, runtime-managed config, build output,
// and test residue. Every ignored path outside this list — including one
// nobody has classified — preserves the worktree (REQ-WR-024, fail-closed).
//
// [HARD] Do not extend this list by omission. A path is added here only after
// it is observed AND shown to be reproduced by a build, a runtime write, or a
// re-run.
var regenerableIgnoredPaths = []string{
	".moai/state",
	".moai/logs",
	".claude/settings.local.json",
	".claude/settings.local.json.lock",
	"bin",
	"docs-site/public",
	".ruff_cache",
	"internal/cli/.moai",
}

// ignoredContentAllowsRemoval reports whether a candidate that has already
// passed the merge, dirty, anchor, and refusal guards may be removed, given
// what gitignored content it holds (REQ-WR-024, policy P2).
//
// Why this guard exists at all: `git status --porcelain` and `git worktree
// remove` AGREE in disregarding ignored files (design.md §A.6, measured), so
// the dirty guard has no backstop for this class — an ignored-only tree is
// removed, exit 0, and its content is destroyed. This is the only thing
// between the sweep and that loss.
//
// @MX:NOTE: [AUTO] P2 (preserve on non-allowlisted ignored content) is a
// STOPGAP, not the permanent answer — REQ-WR-025. drain-then-dispose is the fix.
// @MX:REASON: it preserves worktrees only because worktree-local agent memory
// has nowhere else to live. The correct fix is drain-then-dispose: move that
// memory into the primary checkout's store, then dispose of the tree freely.
// Held out of scope in spec.md §G. Do not read this branch as establishing
// preserve-forever.
func ignoredContentAllowsRemoval(wtPath string, out io.Writer) bool {
	porcelain, err := sessionWorktreeGitStatusIgnored(wtPath)
	if err != nil {
		// Fail-closed on the guard: an unreadable ignored-status is an
		// undetermined answer, and undetermined preserves.
		_, _ = fmt.Fprintf(out, "moai: PR-merge cleanup skipped (cause=ignored-check-failed; ignored-content check failed: %v): worktree %s preserved\n", err, wtPath)
		return false
	}
	if irreplaceable := irreplaceableIgnoredEntries(porcelain); len(irreplaceable) > 0 {
		_, _ = fmt.Fprintf(out, "moai: PR-merge cleanup skipped (cause=ignored-content; irreplaceable gitignored content: %s): worktree %s preserved\n", strings.Join(irreplaceable, ", "), wtPath)
		return false
	}
	return true
}

// irreplaceableIgnoredEntries returns the `git status --porcelain --ignored`
// entries that are NOT covered by regenerableIgnoredPaths. An empty result
// means every ignored entry is regenerable and the tree is safe to remove.
func irreplaceableIgnoredEntries(porcelain string) []string {
	var irreplaceable []string
	for _, line := range strings.Split(porcelain, "\n") {
		if !strings.HasPrefix(line, "!! ") {
			continue
		}
		entry := strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "!! ")), "\"")
		entry = strings.TrimSuffix(entry, "/")
		if entry == "" || isRegenerableIgnoredPath(entry) {
			continue
		}
		irreplaceable = append(irreplaceable, entry)
	}
	return irreplaceable
}

// isRegenerableIgnoredPath reports whether entry is at or below an allowlisted
// regenerable path.
func isRegenerableIgnoredPath(entry string) bool {
	for _, p := range regenerableIgnoredPaths {
		if entry == p || strings.HasPrefix(entry, p+"/") {
			return true
		}
	}
	return false
}

// branchMergedForCleanup decides whether a branch is a cleanup candidate per
// REQ-SW-023. Primary path (gh available): state == "MERGED". Fallback path
// (gh absent): branch appears in `git branch --merged origin/main`. The
// fallback is squash-merge blind — squash-merged branches are NOT listed, so
// the worktree is preserved (documented via the on-entry blindness notice).
func branchMergedForCleanup(branch string, ghAvailable bool) bool {
	if ghAvailable {
		return sessionWorktreeGhPRViewState(branch) == "MERGED"
	}
	merged, err := sessionWorktreeGitBranchMerged()
	if err != nil {
		// Fail-open: a --merged query failure means we cannot confirm merge
		// state → preserve (do not risk removing an unmerged worktree).
		return false
	}
	for _, b := range merged {
		if b == branch {
			return true
		}
	}
	return false
}

// --- real implementations (overridable in tests via the seams above) ---

// gitWorktreeListReal runs `git worktree list --porcelain` and returns its
// stdout. Empty stdout is valid (single worktree, no WT- candidates).
func gitWorktreeListReal() (string, error) {
	out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ghLookPathReal reports whether `gh` is on PATH. Absence is the fallback
// trigger (BI-8); it is NOT an error.
func ghLookPathReal() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// ghPRViewStateReal runs `gh pr view <branch> --json state` and returns the
// parsed `state` field. An empty string is returned when gh errors, when the
// branch has no PR, or when the JSON is malformed — the caller treats empty as
// "not MERGED" (not a cleanup candidate).
func ghPRViewStateReal(branch string) string {
	out, err := exec.Command("gh", "pr", "view", branch, "--json", "state").Output()
	if err != nil {
		return ""
	}
	return parseGhPRStateJSON(string(out))
}

// parseGhPRStateJSON extracts the `state` field from a `gh pr view --json
// state` payload. Returns "" for a missing field or malformed JSON.
func parseGhPRStateJSON(payload string) string {
	var v struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal([]byte(payload), &v); err != nil {
		return ""
	}
	return v.State
}

// gitStatusIgnoredReal runs `git -C <wtPath> status --porcelain --ignored` and
// returns its stdout. Ignored entries are the `!! ` lines; tracked/untracked
// entries also appear but are the dirty guard's business, not this one's.
func gitStatusIgnoredReal(wtPath string) (string, error) {
	out, err := exec.Command("git", "-C", wtPath, "status", "--porcelain", "--ignored").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// gitBranchMergedReal runs `git branch --merged origin/main` and returns the
// list of branch names fully merged into origin/main. An error (e.g. no remote
// origin/main ref) yields an empty list + the error so the caller can fail-open
// preserve.
func gitBranchMergedReal() ([]string, error) {
	out, err := exec.Command("git", "branch", "--merged", "origin/main").Output()
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, line := range strings.Split(string(out), "\n") {
		// `git branch` prefixes the current branch with `*`; strip it. Also
		// strip leading/trailing whitespace.
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "*"))
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}
