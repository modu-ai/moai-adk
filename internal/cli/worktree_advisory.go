package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-WORKTREE-BRANCH-GUARD-001 (REQ-WBG-009). The primary project checkout is
// shared mutable state — a branch switch/reset/rebase issued by one session
// relocates every other concurrent session's tree. The CLI surfaces an advisory
// on `init`, `update`, and `web` naming the hazard and recommending worktree
// isolation for branch-changing work. The advisory consumes
// `workflow.worktree.auto_create` (`workflow.yaml`) to honor the project's
// auto-create setting: when auto_create is `true` the advisory is phrased as an
// auto-creation notice; when `false` (the current default — L1 worktree is a
// Claude Code runtime-autonomous step, MoAI does not auto-create) it stays a
// recommendation. Either phrasing MUST contain a substring matching
// `worktree.*isolation|use a worktree|moai (cc|cg) -w|claude --worktree` so the
// advisory is mechanically observable (AC-WBG-009).

// emitWorktreeAdvisory prints the shared-checkout worktree advisory to out,
// phrased according to the project's `workflow.worktree.auto_create` setting.
// Failures to read the config degrade silently to the recommendation wording
// (the default `auto_create: false` policy) — the advisory is non-blocking.
func emitWorktreeAdvisory(out io.Writer, projectRoot string) {
	autoCreate := readWorktreeAutoCreate(projectRoot)
	if autoCreate {
		// Auto-create is enabled: surface the auto-creation intent rather than
		// a recommendation. Still names the shared-checkout hazard and the -w
		// flag so AC-WBG-009's regex matches.
		_, _ = fmt.Fprintln(out,
			"Note: this checkout is shared across concurrent sessions; "+
				"branch-changing work (switch/reset/rebase) is auto-creating a worktree for isolation "+
				"(use `moai cc -w` / `moai cg -w`, or `claude --worktree`). "+
				"See .claude/rules/moai/workflow/main-checkout-branch-guard.md.")
		return
	}
	_, _ = fmt.Fprintln(out,
		"Tip: this checkout is shared across concurrent sessions; "+
			"for branch-changing work (switch/reset/rebase), use a worktree for isolation — "+
			"`moai cc -w` / `moai cg -w`, or `claude --worktree`. "+
			"See .claude/rules/moai/workflow/main-checkout-branch-guard.md.")
}

// readWorktreeAutoCreate reads workflow.worktree.auto_create from the project's
// workflow.yaml. Returns false (the documented default) on any error, missing
// file, or absent key — fail-safe to the recommendation wording.
func readWorktreeAutoCreate(projectRoot string) bool {
	if projectRoot == "" {
		return false
	}
	loader := config.NewLoader()
	cfg, err := loader.Load(filepath.Join(projectRoot, ".moai"))
	if err != nil || cfg == nil {
		return false
	}
	return cfg.Workflow.Worktree.AutoCreate
}
