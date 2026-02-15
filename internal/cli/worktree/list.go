package worktree

import (
	"fmt"
	"strings"

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

	title := fmt.Sprintf("Active Worktrees (%d)", len(worktrees))
	var lines []string
	for _, wt := range worktrees {
		branchDisplay := wt.Branch
		if branchDisplay == "" {
			branchDisplay = "(detached)"
		}
		if verbose {
			lines = append(lines,
				fmt.Sprintf("Branch: %s", branchDisplay),
				fmt.Sprintf("Path:   %s", wt.Path),
				fmt.Sprintf("HEAD:   %s", wt.HEAD),
				"",
			)
		} else {
			head := wt.HEAD[:minLen(len(wt.HEAD), 8)]
			lines = append(lines, fmt.Sprintf("%-14s  %s  %s", branchDisplay, wt.Path, head))
		}
	}
	content := strings.Join(lines, "\n")
	_, _ = fmt.Fprintln(out, wtCard(title, content))
	return nil
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}
