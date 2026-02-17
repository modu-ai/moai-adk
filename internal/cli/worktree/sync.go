package worktree

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync [branch-name]",
		Short: "Sync worktree with base branch",
		Long: `Synchronize a worktree with changes from the base branch.

If branch-name is provided, syncs that worktree.
If omitted, syncs the worktree at the current directory.

Strategies:
  merge   - Merge base branch into worktree (default)
  rebase  - Rebase worktree onto base branch`,
		Args: cobra.MaximumNArgs(1),
		RunE: runSync,
	}
	cmd.Flags().String("base", "main", "Base branch to sync from")
	cmd.Flags().String("strategy", "merge", "Sync strategy: merge or rebase")
	return cmd
}

func runSync(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	base, _ := cmd.Flags().GetString("base")
	strategy, _ := cmd.Flags().GetString("strategy")

	// Determine the worktree path.
	var wtPath string
	if len(args) > 0 {
		// Branch name provided — look up worktree path.
		branchName := resolveSpecBranch(args[0])
		worktrees, err := WorktreeProvider.List()
		if err != nil {
			return fmt.Errorf("list worktrees: %w", err)
		}
		for _, wt := range worktrees {
			if wt.Branch == branchName {
				wtPath = wt.Path
				break
			}
		}
		if wtPath == "" {
			return fmt.Errorf("no worktree found for branch %q", branchName)
		}
	} else {
		// No argument — use current working directory.
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		wtPath = cwd
	}

	_, _ = fmt.Fprintf(out, "Syncing worktree at %s with %s (%s)...\n", wtPath, base, strategy)

	if err := WorktreeProvider.Sync(wtPath, base, strategy); err != nil {
		return fmt.Errorf("sync worktree: %w", err)
	}

	_, _ = fmt.Fprintln(out, wtSuccessCard("Sync complete",
		fmt.Sprintf("Path: %s", wtPath),
		fmt.Sprintf("Base: %s (%s)", base, strategy),
	))
	return nil
}
