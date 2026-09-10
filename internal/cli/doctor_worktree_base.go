package cli

// doctor_worktree_base.go — SPEC-WORKTREE-BASEREF-001 M4 (card t313).
//
// The read-only counterpart of the SessionStart alignment step: it reports the
// same comparison without writing anything. This is also the stated fallback
// for plan.md §B G2 — EnterWorktree's read of refs/remotes/origin/HEAD is
// inferred from behaviour rather than from source, so if a future runtime stops
// reading that symref, this item is what surfaces the resulting mismatch.

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// Seams for test injection, matching the idiom used elsewhere in this package.
// Both delegate to internal/hook so the diagnostic reads exactly what consumer 1
// reads — a second implementation here could report a state the alignment step
// does not act on.
var (
	doctorWorktreeBaseOriginHead = hook.WorktreeBaseBranchOriginHead
	doctorWorktreeBaseResolvable = func(branch string) bool {
		return hook.WorktreeBaseBranchResolvable(branch)
	}
)

// checkWorktreeBaseBranch compares the configured card-worktree base branch
// against refs/remotes/origin/HEAD.
//
// Four distinct states (REQ-WBR-012). The two non-OK states are deliberately
// NOT collapsed: a mismatch is repaired by running the alignment, an
// unresolvable value is repaired by correcting the setting, so one shared
// "base branch problem" message would leave the user with no next step.
//
// The item reports metadata state only. Worktrees already cut from the wrong
// base are not rebased or re-created by this check — see spec.md §C.
func checkWorktreeBaseBranch(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Worktree Base Branch"}

	configured := config.LoadWorktreeBaseBranch(projectRoot)
	if configured == "" {
		check.Status = uikit.CheckOK
		check.Message = "git_strategy.worktree_base_branch unset — card worktrees follow the repository default"
		if verbose {
			check.Detail = "set it in .moai/config/sections/git-strategy.yaml to pin the base branch"
		}
		return check
	}

	current, err := doctorWorktreeBaseOriginHead()
	if err != nil {
		// No origin remote, or no origin/HEAD symref. Consumer 1 is a silent
		// no-op in this state, so the item is informational, not a failure.
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("git_strategy.worktree_base_branch = %q; refs/remotes/origin/HEAD is unset — nothing to compare", configured)
		if verbose {
			check.Detail = err.Error()
		}
		return check
	}

	if current == configured {
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("refs/remotes/origin/HEAD names %s, matching git_strategy.worktree_base_branch", current)
		return check
	}

	if !doctorWorktreeBaseResolvable(configured) {
		check.Status = uikit.CheckFail
		check.Message = fmt.Sprintf(
			"git_strategy.worktree_base_branch names %q, which has no remote-tracking branch — correct the setting in .moai/config/sections/git-strategy.yaml",
			configured)
		if verbose {
			check.Detail = fmt.Sprintf("probe: git show-ref --verify refs/remotes/origin/%s", configured)
		}
		return check
	}

	check.Status = uikit.CheckWarn
	check.Message = fmt.Sprintf(
		"refs/remotes/origin/HEAD names %s but git_strategy.worktree_base_branch is %s — repair with: git remote set-head origin %s",
		current, configured, configured)
	if verbose {
		check.Detail = "existing worktrees are not re-cut by this repair; an empty one can be removed and re-entered"
	}
	return check
}
