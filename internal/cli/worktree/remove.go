package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [path]",
		Short: "Remove a worktree",
		Long:  "Remove a Git worktree at the specified path.",
		Args:  cobra.ExactArgs(1),
		RunE:  runRemove,
	}
	cmd.Flags().Bool("force", false, "Force removal even with uncommitted changes")
	return cmd
}

func runRemove(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	wtPath := args[0]

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("get force flag: %w", err)
	}

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	if err := WorktreeProvider.Remove(wtPath, force); err != nil {
		return fmt.Errorf("remove worktree: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Removed worktree at %s\n", wtPath)
	return nil
}
