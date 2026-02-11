package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active worktrees",
		Long:  "Display all active Git worktrees including the main worktree.",
		RunE:  runList,
	}
	cmd.Flags().BoolP("verbose", "v", false, "Show detailed information for each worktree")
	return cmd
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

	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		_, _ = fmt.Fprintf(out, "Active Worktrees (%d):\n\n", len(worktrees))
		for _, wt := range worktrees {
			branchDisplay := wt.Branch
			if branchDisplay == "" {
				branchDisplay = "(detached)"
			}
			_, _ = fmt.Fprintf(out, "  Branch: %s\n", branchDisplay)
			_, _ = fmt.Fprintf(out, "  Path:   %s\n", wt.Path)
			_, _ = fmt.Fprintf(out, "  HEAD:   %s\n\n", wt.HEAD)
		}
	} else {
		_, _ = fmt.Fprintln(out, "Active Worktrees:")
		for _, wt := range worktrees {
			_, _ = fmt.Fprintf(out, "  %s  [%s]  %s\n", wt.Path, wt.Branch, wt.HEAD[:minLen(len(wt.HEAD), 8)])
		}
	}
	return nil
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}
