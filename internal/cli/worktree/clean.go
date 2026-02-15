package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean stale worktree references",
		Long: `Remove stale worktree references and optionally remove worktrees
whose branches have been merged into the base branch.`,
		RunE: runClean,
	}
	cmd.Flags().Bool("merged-only", false, "Only remove worktrees whose branches are merged into base")
	cmd.Flags().String("base", "main", "Base branch for --merged-only check")
	return cmd
}

func runClean(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	mergedOnly, _ := cmd.Flags().GetBool("merged-only")

	if mergedOnly {
		base, _ := cmd.Flags().GetString("base")
		return cleanMergedWorktrees(cmd, base)
	}

	if err := WorktreeProvider.Prune(); err != nil {
		return fmt.Errorf("prune worktrees: %w", err)
	}

	_, _ = fmt.Fprintln(out, wtSuccessCard("Cleaned stale worktree references"))
	return nil
}

// cleanMergedWorktrees removes worktrees whose branches are fully merged.
func cleanMergedWorktrees(cmd *cobra.Command, base string) error {
	out := cmd.OutOrStdout()

	worktrees, err := WorktreeProvider.List()
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}

	var removed int
	for _, wt := range worktrees {
		if wt.Branch == "" || wt.Branch == base {
			continue
		}
		merged, err := WorktreeProvider.IsBranchMerged(wt.Branch, base)
		if err != nil {
			_, _ = fmt.Fprintf(out, "  Warning: could not check %s: %v\n", wt.Branch, err)
			continue
		}
		if merged {
			_, _ = fmt.Fprintf(out, "  Removing merged worktree: %s [%s]\n", wt.Path, wt.Branch)
			if err := WorktreeProvider.Remove(wt.Path, false); err != nil {
				_, _ = fmt.Fprintf(out, "  Warning: could not remove %s: %v\n", wt.Path, err)
				continue
			}
			removed++
		}
	}

	if removed == 0 {
		_, _ = fmt.Fprintln(out, "No merged worktrees to clean.")
	} else {
		_, _ = fmt.Fprintf(out, "Removed %d merged worktree(s).\n", removed)
	}
	return nil
}
