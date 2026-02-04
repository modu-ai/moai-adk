package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Clean stale worktree references",
		Long:  "Remove stale worktree references for directories that have been deleted.",
		RunE:  runClean,
	}
}

func runClean(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	if err := WorktreeProvider.Prune(); err != nil {
		return fmt.Errorf("prune worktrees: %w", err)
	}

	_, _ = fmt.Fprintln(out, "Cleaned stale worktree references.")
	return nil
}
