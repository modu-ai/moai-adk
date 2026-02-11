// Package worktree provides Git worktree management subcommands.
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
