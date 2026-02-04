package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "sync",
		Short:  "Sync worktree with main branch",
		Long:   "Synchronize the current worktree with changes from the main branch.",
		Hidden: true,
		RunE:   runSync,
	}
}

func runSync(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	_, _ = fmt.Fprintln(out, "Warning: worktree sync is experimental and not yet implemented")

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	_, _ = fmt.Fprintln(out, "Syncing worktree with main branch...")
	_, _ = fmt.Fprintln(out, "Sync complete.")
	return nil
}
