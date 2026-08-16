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
)

// wtEntry is a parsed (path, branch) pair from `git worktree list --porcelain`.
// A detached-HEAD worktree carries an empty branch.
type wtEntry struct {
	path   string
	branch string
}

// parseWorktreeList parses the `git worktree list --porcelain` output into
// (path, branch) pairs. Each worktree entry begins with a `worktree <path>`
// line and carries at most one `branch refs/heads/<name>` line; detached
// entries have `detached` instead and yield an empty branch. Entries are
// separated by blank lines.
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
			_, _ = fmt.Fprintf(out, "moai: PR-merge cleanup skipped (dirty-check failed: %v): worktree %s preserved\n", derr, e.path)
			continue
		}
		if dirty {
			_, _ = fmt.Fprintf(out, "moai: PR-merge cleanup skipped (uncommitted changes): worktree %s preserved (dispose manually via 'moai worktree remove' or 'git worktree remove')\n", e.path)
			continue
		}
		// t73 anchor guard: never remove a worktree a live session is
		// anchored in — the anchored session's shell dies with the tree
		// (Claude Code blocks all its Bash once the tree is gone). Same
		// immediately-before-removal position as the dirty guard (EC-11).
		if anchored := session.LiveAnchoredSessions(e.path, time.Now()); len(anchored) > 0 {
			_, _ = fmt.Fprintf(out, "moai: PR-merge cleanup skipped (live session anchored, %d): worktree %s preserved\n", len(anchored), e.path)
			continue
		}
		// Remove + unset safe.directory (R5, reused from M7). A removal failure
		// is non-blocking: the worktree is left on disk and the sweep
		// continues with the next candidate.
		if rerr := sessionWorktreeGitWorktreeRemove(e.path); rerr != nil {
			_, _ = fmt.Fprintf(out, "moai: PR-merge cleanup failed (%v): worktree %s left on disk\n", rerr, e.path)
			continue
		}
		_ = sessionWorktreeGitSafeDirUnset(e.path)
		_, _ = fmt.Fprintf(out, "moai: %s [WT] worktree %s (branch %s merged)\n", PRMergeCleanupNoticePrefix, e.path, e.branch)
	}
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
