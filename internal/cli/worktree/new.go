package worktree

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [branch-name]",
		Short: "Create a new worktree",
		Long: `Create a new Git worktree for the given branch name.
If the branch does not exist, it is created automatically.

SPEC-ID patterns (e.g., SPEC-AUTH-001) are automatically converted
to branch names using the feature/ prefix convention.`,
		Args: cobra.ExactArgs(1),
		RunE: runNew,
	}
	cmd.Flags().String("path", "", "Custom path for the worktree (default: .moai/worktrees/<SPEC-ID> for SPEC IDs, ../<branch-name> otherwise)")
	cmd.Flags().String("base", "main", "Base branch to create the worktree from")
	return cmd
}

func runNew(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	specID := args[0]
	branchName := resolveSpecBranch(specID)

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	wtPath, err := cmd.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("get path flag: %w", err)
	}
	if wtPath == "" {
		if isSpecID(specID) {
			wtPath = filepath.Join(".moai", "worktrees", specID)
		} else {
			wtPath = filepath.Join("..", branchName)
		}
	}

	if err := WorktreeProvider.Add(wtPath, branchName); err != nil {
		return fmt.Errorf("create worktree: %w", err)
	}

	_, _ = fmt.Fprintln(out, wtSuccessCard(
		fmt.Sprintf("Created worktree for branch %s", branchName),
		fmt.Sprintf("Path: %s", wtPath),
	))
	return nil
}

// resolveSpecBranch converts SPEC-ID patterns to branch names.
// e.g., "SPEC-AUTH-001" -> "feature/SPEC-AUTH-001"
// Regular branch names pass through unchanged.
func resolveSpecBranch(name string) string {
	if isSpecID(name) {
		return "feature/" + name
	}
	return name
}

// isSpecID checks if the given name matches the SPEC-ID pattern.
// Pattern: SPEC-<CATEGORY>-<NUMBER> (e.g., SPEC-AUTH-001, SPEC-UI-042)
func isSpecID(name string) bool {
	if !strings.HasPrefix(name, "SPEC-") {
		return false
	}
	parts := strings.SplitN(name, "-", 3)
	return len(parts) >= 3 && parts[2] != ""
}
