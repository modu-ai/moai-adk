package worktree

import (
	"fmt"
	"strings"

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

	if len(worktrees) == 0 {
		_, _ = fmt.Fprintln(out, wtCard("Worktree Status",
			fmt.Sprintf("Repository: %s\nTotal worktrees: %d\n\nNo worktrees found.", WorktreeProvider.Root(), len(worktrees))))
		return nil
	}

	var lines []string
	lines = append(lines,
		fmt.Sprintf("Repository: %s", WorktreeProvider.Root()),
		fmt.Sprintf("Total worktrees: %d", len(worktrees)),
		"",
	)
	for _, wt := range worktrees {
		branchDisplay := wt.Branch
		if branchDisplay == "" {
			branchDisplay = "(detached)"
		}
		headDisplay := wt.HEAD
		if !showAll && len(headDisplay) > 8 {
			headDisplay = headDisplay[:8]
		}
		lines = append(lines,
			branchDisplay,
			fmt.Sprintf("  Path: %s", wt.Path),
			fmt.Sprintf("  HEAD: %s", headDisplay),
			"",
		)
	}

	_, _ = fmt.Fprintln(out, wtCard("Worktree Status", strings.Join(lines, "\n")))
	return nil
}
