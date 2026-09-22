// Package worktree provides Git worktree management subcommands.
// @MX:NOTE: [AUTO] Worktree management for parallel SPEC development with isolated working directories
// @MX:NOTE: [AUTO] Dependency injection pattern: WorktreeProvider set from parent CLI package
// @MX:NOTE: [AUTO] Supports new, sync, remove, clean, recover, done, and the guard subcommands

package worktree

import (
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// WorktreeProvider supplies git worktree operations to subcommands.
// Set this from the parent CLI package during DI wiring.
var WorktreeProvider git.WorktreeManager

// WorktreeCmd is the parent "worktree" command with alias "wt".
var WorktreeCmd = &cobra.Command{
	Use:     "worktree",
	Aliases: []string{"wt"},
	Short:   "Git worktree management",
	GroupID: "tools",
	Long: `Manage Git worktrees for parallel SPEC development: new, sync, remove, clean, recover and done, plus the guard verbs snapshot, verify and restore.

Create a harness-neutral L1 worktree through MoAI's shared materializer:
  moai worktree new <name>     create .claude/worktrees/<name>

Entering an existing worktree remains the launchers' job:
  moai cc -w <name>            work inside the worktree
  moai codex -w <name>         start Codex inside the worktree
  moai cc -w <name> --spawn    open it in a new tmux window, keep this session

For inspection, use git directly: git worktree list`,
}

func init() {
	WorktreeCmd.AddCommand(
		newNewCmd(),
		newSyncCmd(),
		newRemoveCmd(),
		newCleanCmd(),
		newRecoverCmd(),
		newDoneCmd(),
		newGuardSnapshotCmd(),
		newGuardVerifyCmd(),
		newGuardRestoreCmd(),
	)
}
