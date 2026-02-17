package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newGoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "go [branch-name]",
		Short: "Print worktree path for shell navigation",
		Long: `Print the worktree path for the given branch for shell navigation.

Use with shell command substitution to change directory:
  cd $(moai worktree go my-branch)
  cd $(moai wt go my-branch)`,
		Args: cobra.ExactArgs(1),
		RunE: runGo,
	}
}

func runGo(cmd *cobra.Command, args []string) error {
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
			// Output only the path for shell eval: cd $(moai wt go branch)
			_, _ = fmt.Fprintln(out, wt.Path)
			return nil
		}
	}

	return fmt.Errorf("no worktree found for branch %q", branchName)
}
