package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRecoverCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "recover",
		Short: "Repair worktree registry",
		Long:  "Repair worktree administrative files by scanning disk and running 'git worktree repair'.",
		RunE:  runRecover,
	}
}

func runRecover(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	_, _ = fmt.Fprintf(out, "Scanning for worktrees in %s...\n", WorktreeProvider.Root())

	if err := WorktreeProvider.Repair(); err != nil {
		return fmt.Errorf("repair worktrees: %w", err)
	}

	// Prune stale references after repair.
	if err := WorktreeProvider.Prune(); err != nil {
		return fmt.Errorf("prune worktrees: %w", err)
	}

	// List recovered worktrees.
	worktrees, err := WorktreeProvider.List()
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}

	if len(worktrees) == 0 {
		_, _ = fmt.Fprintln(out, "No worktrees found.")
	} else {
		_, _ = fmt.Fprintf(out, "Recovered %d worktree(s):\n", len(worktrees))
		for _, wt := range worktrees {
			_, _ = fmt.Fprintf(out, "  %s  [%s]\n", wt.Path, wt.Branch)
		}
	}

	return nil
}
