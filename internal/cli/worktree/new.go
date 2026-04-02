package worktree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// userHomeDirFunc resolves the user's home directory.
// Overridable in tests.
var userHomeDirFunc = os.UserHomeDir

// getProjectNameFunc resolves the project name for worktree path construction.
// Overridable in tests.
var getProjectNameFunc = func() string { return detectProjectName(".") }

// isTmuxAvailableFunc checks if tmux is available.
// Overridable in tests.
var isTmuxAvailableFunc = IsTmuxAvailable

// legacyWorktreeDir is the project-local worktrees path checked for migration warnings.
// Overridable in tests.
var legacyWorktreeDir = filepath.Join(".", ".moai", "worktrees")

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
	cmd.Flags().Bool("tmux", false, "Create a tmux session after worktree creation")
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
		homeDir, err := userHomeDirFunc()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		projectName := getProjectNameFunc()
		wtPath = filepath.Join(homeDir, ".moai", "worktrees", projectName, specID)
		if err := os.MkdirAll(filepath.Dir(wtPath), 0o755); err != nil {
			return fmt.Errorf("create worktree directory: %w", err)
		}
	}

	// R5: warn when legacy project-local worktrees exist.
	if dirHasEntries(legacyWorktreeDir) {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Warning: Legacy worktrees detected in .moai/worktrees/. Consider moving to ~/.moai/worktrees/{Project}/.")
	}

	if err := WorktreeProvider.Add(wtPath, branchName); err != nil {
		return fmt.Errorf("create worktree: %w", err)
	}

	_, _ = fmt.Fprintln(out, wtSuccessCard(
		fmt.Sprintf("Created worktree for branch %s", branchName),
		fmt.Sprintf("Path: %s", wtPath),
	))

	// R5: tmux session creation after worktree creation
	// Check if --tmux flag is set
	tmuxFlag, _ := cmd.Flags().GetBool("tmux")
	if tmuxFlag || isTmuxPreferred() {
		if isTmuxAvailableFunc() {
			// Create tmux session with environment isolation
			projectName := getProjectNameFunc()
			_, err := BuildTmuxSessionConfig(projectName, specID, wtPath, ".")
			if err != nil {
				return fmt.Errorf("build tmux config: %w", err)
			}

			// Note: Using a nil tmux manager for now - will need proper initialization
			// This is a simplified implementation for the TDD cycle
			_, _ = fmt.Fprintln(out, "Tmux session creation requested but tmux manager not yet initialized.")
			_, _ = fmt.Fprintf(out, "To manually create session: tmux new-session -s %s -c %s\n", GenerateTmuxSessionName(projectName, specID), wtPath)
		} else {
			// Graceful degradation: print manual instructions
			err := NewTmuxNotAvailableError(specID, wtPath)
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
		}
	}

	return nil
}

// dirHasEntries returns true when dir exists and contains at least one entry.
func dirHasEntries(dir string) bool {
	entries, err := os.ReadDir(dir)
	return err == nil && len(entries) > 0
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

// GenerateTmuxSessionName implements R5.1: tmux session name generation pattern.
// SPEC-WORKTREE-002 requirement: moai-{ProjectName}-{SPEC-ID}
//
// @MX:NOTE: SPEC-WORKTREE-002 R5.1 implementation - tmux session name pattern
// @MX:SPEC: SPEC-WORKTREE-002
//
// Args:
//   - projectName: the project name (e.g., "moai-adk-go")
//   - specID: the SPEC identifier (e.g., "SPEC-WORKTREE-002")
//
// Returns:
//   - the tmux session name (e.g., "moai-moai-adk-go-SPEC-WORKTREE-002")
//
// Example:
//
//	sessionName := GenerateTmuxSessionName("moai-adk-go", "SPEC-WORKTREE-002")
//	fmt.Println(sessionName) // Output: moai-moai-adk-go-SPEC-WORKTREE-002
func GenerateTmuxSessionName(projectName, specID string) string {
	return fmt.Sprintf("moai-%s-%s", projectName, specID)
}

// ShouldAutoMerge implements R3: default auto-merge behavior.
// SPEC-WORKTREE-002 requirement: auto-merge is the default in the worktree flow.
//
// @MX:NOTE: SPEC-WORKTREE-002 R3 implementation - auto-merge default
// @MX:SPEC: SPEC-WORKTREE-002
//
// Args:
//   - noMergeFlag: whether the --no-merge flag is set
//
// Returns:
//   - true if auto-merge should be performed, false otherwise
//
// Behavior:
//   - Default (true): performs auto-merge when the flag is absent
//   - --no-merge flag: returns false, skipping auto-merge
//
// Example:
//
//	shouldMerge := ShouldAutoMerge(false) // true
//	shouldMerge := ShouldAutoMerge(true)  // false
func ShouldAutoMerge(noMergeFlag bool) bool {
	// R3: auto-merge is the default (true)
	// Can only be disabled via the --no-merge flag
	return !noMergeFlag
}

// isTmuxPreferred checks if tmux session creation is preferred in workflow config.
// This is a placeholder for future workflow.yaml integration.
//
// @MX:NOTE: SPEC-WORKTREE-002 workflow config integration point
// @MX:SPEC: SPEC-WORKTREE-002
func isTmuxPreferred() bool {
	// TODO: Read from .moai/config/sections/workflow.yaml
	// For now, return false to require explicit --tmux flag
	return false
}
