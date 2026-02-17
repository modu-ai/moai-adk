package worktree

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done [branch-name]",
		Short: "Complete worktree and cleanup",
		Long: `Complete a worktree by removing it and optionally deleting the branch.

This command performs the completion workflow:
1. Remove the worktree at the specified branch
2. Optionally delete the feature branch (with --delete-branch)

Note: Merging to base branch should be done separately via git merge or PR.`,
		Args: cobra.ExactArgs(1),
		RunE: runDone,
	}
	cmd.Flags().Bool("force", false, "Force removal even with uncommitted changes")
	cmd.Flags().Bool("delete-branch", false, "Delete the branch after removing worktree")
	return cmd
}

func runDone(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	branchName := resolveSpecBranch(args[0])

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("get force flag: %w", err)
	}

	deleteBranch, err := cmd.Flags().GetBool("delete-branch")
	if err != nil {
		return fmt.Errorf("get delete-branch flag: %w", err)
	}

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	// Find the worktree for the given branch.
	worktrees, err := WorktreeProvider.List()
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}

	var targetPath string
	for _, wt := range worktrees {
		if wt.Branch == branchName {
			targetPath = wt.Path
			break
		}
	}

	if targetPath == "" {
		return fmt.Errorf("no worktree found for branch %q", branchName)
	}

	// Remove the worktree.
	if err := WorktreeProvider.Remove(targetPath, force); err != nil {
		return fmt.Errorf("remove worktree: %w", err)
	}

	details := []string{
		fmt.Sprintf("Path: %s", targetPath),
		"Worktree removed.",
	}

	if deleteBranch {
		if err := WorktreeProvider.DeleteBranch(branchName); err != nil {
			details = append(details,
				fmt.Sprintf("Warning: could not delete branch: %v", err),
				fmt.Sprintf("To delete manually: git branch -d %s", branchName),
			)
		} else {
			details = append(details, fmt.Sprintf("Branch %s deleted.", branchName))
		}
	}

	_, _ = fmt.Fprintln(out, wtSuccessCard(
		fmt.Sprintf("Done: worktree for branch %s", branchName),
		details...,
	))
	return nil
}
