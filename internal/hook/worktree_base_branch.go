package hook

// worktree_base_branch.go — SPEC-WORKTREE-BASEREF-001 M2 (card t313).
//
// Consumer 1 of git_strategy.worktree_base_branch: the SessionStart step that
// aligns refs/remotes/origin/HEAD with the configured branch.
//
// Why this exists. The Claude Code EnterWorktree tool's `fresh` mode branches
// from refs/remotes/origin/HEAD, and worktree.baseRef cannot name a branch
// ("accepts only \"fresh\" or \"head\", not arbitrary refs"). So the only handle
// on that base is the remote-HEAD symref itself — local repository metadata that
// does not survive a fresh clone. This step derives that metadata from a stored
// setting instead of leaving it hand-applied.
//
// Why it fires only from the primary checkout (REQ-WBR-004, second clause).
// git-strategy.yaml is a TRACKED file, so every card worktree carries its own
// copy whose content follows its own branch, while refs/remotes/origin/HEAD is
// a single repository-global handle. Unnarrowed, every active lane would be a
// writer of one shared handle and two lanes on different branches would each
// reverse the other's write forever. One writer suffices, because the handle is
// repository-global — so the step gates on the primary checkout BEFORE reading
// anything. Consumer 2 (internal/cli worktree creation) is deliberately NOT
// narrowed: it passes an operand to its own `git worktree add` rather than
// mutating shared metadata.

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Function-variable seams for test injection. Each has a Real counterpart
// below; tests swap these via swapWorktreeBaseBranchSeams and restore on
// cleanup — the idiom already used by internal/cli/session_worktree.go.
//
// Three of them are EXPORTED, and only because a guard in another package must
// be able to observe them: the anti-dead-key regression guard asserts that a
// stored key reaches the git write seam, which is not something a grep or an
// in-package test can discharge on its behalf.
var (
	// WorktreeBaseBranchInPrimaryCheckout reports whether cwd is the primary
	// checkout (git-dir == git-common-dir), the REQ-WBR-004 discriminant.
	WorktreeBaseBranchInPrimaryCheckout = worktreeBaseBranchInPrimaryCheckoutReal

	// worktreeBaseBranchReadConfig is the ALIGNMENT-ENTRY seam: the read of the
	// configured git_strategy.worktree_base_branch value. AC-WBR-016 counts THIS
	// seam — exactly 1 invocation per Handle from the primary checkout (for every
	// configured value, empty included), 0 from a linked worktree. It is
	// deliberately separate from the origin/HEAD read seam below, which the empty
	// path never reaches.
	worktreeBaseBranchReadConfig = worktreeBaseBranchReadConfigReal

	// WorktreeBaseBranchOriginHead returns the branch name currently named
	// by refs/remotes/origin/HEAD. Also read by the `moai doctor` diagnostic,
	// which reports this same comparison without writing anything.
	WorktreeBaseBranchOriginHead = worktreeBaseBranchReadOriginHeadReal

	// WorktreeBaseBranchSetHead runs `git remote set-head origin <branch>`.
	WorktreeBaseBranchSetHead = worktreeBaseBranchSetHeadReal

	// worktreeBaseBranchStderr is the notice channel, matching the existing
	// SessionStart precedent (the empty-session_id warning).
	worktreeBaseBranchStderr io.Writer = os.Stderr
)

// WorktreeBaseBranchResolvable reports whether a configured base-branch value
// resolves to an existing remote-tracking branch.
//
// This is the SOLE resolvability authority for BOTH consumers (REQ-WBR-011).
// internal/cli's worktree creation path calls it through its own seam rather
// than carrying a second rule: two consumers that disagree about whether a
// configured value is usable is the defect this single-helper contract exists
// to prevent — and a divergent rule whose runtime behaviour happens to agree
// today is still a violation, because it can drift tomorrow.
//
// Exported as a variable so callers can swap it in tests through the same seam
// mechanism the rest of this file uses.
var WorktreeBaseBranchResolvable = worktreeBaseBranchResolvableReal

// RunWorktreeBaseAlignment performs the origin/HEAD alignment and returns its
// own local data map, matching the shape of the other SessionStart errgroup
// tasks. Every failure path is fail-open: this step never blocks, fails, or
// aborts session start (REQ-WBR-008).
func RunWorktreeBaseAlignment(projectRoot string) map[string]any {
	data := make(map[string]any)

	// Gate on the primary checkout FIRST — before any read (REQ-WBR-004).
	if !WorktreeBaseBranchInPrimaryCheckout() {
		return data
	}

	configured := worktreeBaseBranchReadConfig(projectRoot)
	if configured == "" {
		// REQ-WBR-005: the neutral value performs no git-metadata read at all.
		return data
	}

	current, err := WorktreeBaseBranchOriginHead()
	if err != nil {
		// No origin remote, no origin/HEAD, or git unavailable — silent no-op.
		return data
	}
	if current == configured {
		// REQ-WBR-006: steady state stays write-free and silent, which is what
		// keeps the notice below a signal rather than noise.
		return data
	}

	if !WorktreeBaseBranchResolvable(configured) {
		// REQ-WBR-009: pointing refs/remotes/origin/HEAD at a ref that does not
		// exist is strictly worse than the defect this SPEC repairs, so a typo
		// costs one diagnostic line and nothing else.
		fmt.Fprintf(worktreeBaseBranchStderr,
			"warning: git_strategy.worktree_base_branch names %q, which has no remote-tracking branch "+
				"(refs/remotes/origin/%s); refs/remotes/origin/HEAD left unchanged — correct the setting\n",
			configured, configured)
		data["worktree_base_branch_unresolvable"] = configured
		return data
	}

	if err := WorktreeBaseBranchSetHead(configured); err != nil {
		return data
	}

	fmt.Fprintf(worktreeBaseBranchStderr,
		"notice: refs/remotes/origin/HEAD realigned from %s to %s per git_strategy.worktree_base_branch\n",
		current, configured)
	data["worktree_base_branch_aligned"] = configured
	return data
}

// --- real implementations (overridable in tests via the seams above) ---

// worktreeBaseBranchInPrimaryCheckoutReal reports whether cwd is the primary
// checkout. Inverse of internal/cli's inGitWorktreeReal, using the same
// discriminant: --git-dir and --git-common-dir resolve to the same path in the
// primary checkout and differ inside a linked worktree. Any git error degrades
// to false, so an unreadable repository is treated as "not the primary
// checkout" and the step no-ops.
func worktreeBaseBranchInPrimaryCheckoutReal() bool {
	gitDir, err1 := exec.Command("git", "rev-parse", "--git-dir").Output()
	commonDir, err2 := exec.Command("git", "rev-parse", "--git-common-dir").Output()
	if err1 != nil || err2 != nil {
		return false
	}
	return strings.TrimSpace(string(gitDir)) == strings.TrimSpace(string(commonDir))
}

// worktreeBaseBranchReadConfigReal reads the configured value from the
// project's git-strategy.yaml.
func worktreeBaseBranchReadConfigReal(projectRoot string) string {
	return config.LoadWorktreeBaseBranch(projectRoot)
}

// worktreeBaseBranchReadOriginHeadReal returns the short branch name that
// refs/remotes/origin/HEAD points at, or an error when no such symref exists.
func worktreeBaseBranchReadOriginHeadReal() (string, error) {
	out, err := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "refs/remotes/origin/"), nil
}

// worktreeBaseBranchSetHeadReal points refs/remotes/origin/HEAD at the branch.
func worktreeBaseBranchSetHeadReal(branch string) error {
	return exec.Command("git", "remote", "set-head", "origin", branch).Run()
}

// worktreeBaseBranchResolvableReal is the REQ-WBR-009 predicate.
//
// Two constraints bind the implementation shape, not just the outcome:
//
//   - ONLY rc == 0 counts as resolvable. A missing ref exits 128, not 1, so an
//     `rc == 1` test would classify every missing ref as an execution error.
//   - The plumbing form `git show-ref --verify` is required. The porcelain
//     `git branch --list` / `git branch -vv` forms are refused at the tool layer
//     by BranchGuard's `\bgit\s+branch\b` pattern, which does not distinguish a
//     read-only branch query from branch-state mutation — so they are not merely
//     slower, they are blocked.
func worktreeBaseBranchResolvableReal(branch string) bool {
	if branch == "" {
		return false
	}
	return exec.Command("git", "show-ref", "--verify", "refs/remotes/origin/"+branch).Run() == nil
}
