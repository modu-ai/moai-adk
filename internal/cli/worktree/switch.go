package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch [branch-name]",
		Short: "Switch to a worktree",
		Long:  "Change the working directory to the worktree associated with the given branch.",
		Args:  cobra.ExactArgs(1),
		RunE:  runSwitch,
	}
}

func runSwitch(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	branchName := resolveSpecBranch(args[0])

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	worktrees, err := WorktreeProvider.List()
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}

	for _, wt := range worktrees {
		if wt.Branch == branchName {
			_, _ = fmt.Fprintf(out, "Worktree for branch %s is at: %s\n", branchName, wt.Path)
			return nil
		}
	}

	return fmt.Errorf("no worktree found for branch %q", branchName)
}
