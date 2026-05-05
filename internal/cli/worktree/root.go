// Package worktree provides Git worktree management subcommands.
// @MX:NOTE: [AUTO] Worktree management for parallel SPEC development with isolated working directories
// @MX:NOTE: [AUTO] Dependency injection pattern: WorktreeProvider set from parent CLI package
// @MX:NOTE: [AUTO] Supports create, list, switch, sync, remove, clean, recover, config, status

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
	Long:    "Manage Git worktrees for parallel SPEC development. Supports creating, listing, switching, syncing, removing, and cleaning worktrees.",
}

func init() {
	WorktreeCmd.AddCommand(
		newNewCmd(),
		newListCmd(),
		newSwitchCmd(),
		newGoCmd(),
		newSyncCmd(),
		newRemoveCmd(),
		newCleanCmd(),
		newRecoverCmd(),
		newDoneCmd(),
		newConfigCmd(),
		newStatusCmd(),
	)
}
