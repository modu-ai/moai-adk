package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active worktrees",
		Long:  "Display all active Git worktrees including the main worktree.",
		RunE:  runList,
	}
}

func runList(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	worktrees, err := WorktreeProvider.List()
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}

	if len(worktrees) == 0 {
		_, _ = fmt.Fprintln(out, "No worktrees found.")
		return nil
	}

	_, _ = fmt.Fprintln(out, "Active Worktrees:")
	for _, wt := range worktrees {
		_, _ = fmt.Fprintf(out, "  %s  [%s]  %s\n", wt.Path, wt.Branch, wt.HEAD[:minLen(len(wt.HEAD), 8)])
	}
	return nil
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}
