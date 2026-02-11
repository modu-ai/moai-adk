package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show worktree status",
		Long:  "Show worktree status, prune stale references, and display active worktrees.",
		RunE:  runStatus,
	}
	cmd.Flags().Bool("all", false, "Show all details including full commit hashes")
	return cmd
}

func runStatus(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	showAll, _ := cmd.Flags().GetBool("all")

	// Prune stale worktree references first.
	if err := WorktreeProvider.Prune(); err != nil {
		return fmt.Errorf("prune worktrees: %w", err)
	}

	worktrees, err := WorktreeProvider.List()
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Repository: %s\n", WorktreeProvider.Root())
	_, _ = fmt.Fprintf(out, "Total worktrees: %d\n\n", len(worktrees))

	if len(worktrees) == 0 {
		_, _ = fmt.Fprintln(out, "No worktrees found.")
		return nil
	}

	for _, wt := range worktrees {
		branchDisplay := wt.Branch
		if branchDisplay == "" {
			branchDisplay = "(detached)"
		}
		headDisplay := wt.HEAD
		if !showAll && len(headDisplay) > 8 {
			headDisplay = headDisplay[:8]
		}
		_, _ = fmt.Fprintf(out, "%s\n", branchDisplay)
		_, _ = fmt.Fprintf(out, "  Path: %s\n", wt.Path)
		_, _ = fmt.Fprintf(out, "  HEAD: %s\n", headDisplay)
		_, _ = fmt.Fprintln(out)
	}

	return nil
}
