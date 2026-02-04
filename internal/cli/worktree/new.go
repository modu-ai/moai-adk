package worktree

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [branch-name]",
		Short: "Create a new worktree",
		Long:  "Create a new Git worktree for the given branch name. If the branch does not exist, it is created automatically.",
		Args:  cobra.ExactArgs(1),
		RunE:  runNew,
	}
	cmd.Flags().String("path", "", "Custom path for the worktree (default: ../<branch-name>)")
	return cmd
}

func runNew(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	branchName := args[0]

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	wtPath, err := cmd.Flags().GetString("path")
	if err != nil || wtPath == "" {
		wtPath = filepath.Join("..", branchName)
	}

	if err := WorktreeProvider.Add(wtPath, branchName); err != nil {
		return fmt.Errorf("create worktree: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Created worktree at %s for branch %s\n", wtPath, branchName)
	return nil
}
