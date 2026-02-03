package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show worktree status",
		Long:  "Show worktree status, prune stale references, and display active worktrees.",
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	// Prune stale worktree references first.
	if err := WorktreeProvider.Prune(); err != nil {
		return fmt.Errorf("prune worktrees: %w", err)
	}

	worktrees, err := WorktreeProvider.List()
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}

	fmt.Fprintf(out, "Repository: %s\n", WorktreeProvider.Root())
	fmt.Fprintf(out, "Total worktrees: %d\n\n", len(worktrees))

	if len(worktrees) == 0 {
		fmt.Fprintln(out, "No worktrees found.")
		return nil
	}

	for _, wt := range worktrees {
		branchDisplay := wt.Branch
		if branchDisplay == "" {
			branchDisplay = "(detached)"
		}
		headShort := wt.HEAD
		if len(headShort) > 8 {
			headShort = headShort[:8]
		}
		fmt.Fprintf(out, "%s\n", branchDisplay)
		fmt.Fprintf(out, "  Path: %s\n", wt.Path)
		fmt.Fprintf(out, "  HEAD: %s\n", headShort)
		fmt.Fprintln(out)
	}

	return nil
}
