package worktree

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/session"
)

func newDoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done [branch-name]",
		Short: "Complete worktree and cleanup",
		Long: `Complete a worktree by removing it and optionally deleting the branch.

This command performs the completion workflow:
1. Refuse while a live session is anchored in the worktree (tree-local
   session registry check; --force overrides with a warning)
2. Remove the worktree at the specified branch
3. Optionally delete the feature branch (with --delete-branch)

Note: Merging to base branch should be done separately via git merge or PR.`,
		Args: cobra.ExactArgs(1),
		RunE: runDone,
	}
	cmd.Flags().Bool("force", false, "Force removal even with uncommitted changes")
	cmd.Flags().Bool("delete-branch", false, "Delete the branch after removing worktree")
	cmd.Flags().Bool("auto", false, "Auto mode: silent execution for automation (e.g., after PR merge)")
	return cmd
}

// AutoCleanupFlag is the flag name for auto-cleanup mode.
// Used by sync workflow to trigger cleanup after PR merge.
const AutoCleanupFlag = "auto"

// runDoneWithAutoMode is the internal function that supports auto-cleanup mode.
// When autoMode is true, it suppresses output and handles errors gracefully.
//
// @MX:NOTE: SPEC-WORKTREE-002 R2 implementation - auto-cleanup for PR merge workflow
// @MX:SPEC: SPEC-WORKTREE-002
func runDoneWithAutoMode(branchName string, force, deleteBranch, autoMode bool) (success bool, err error) {
	if WorktreeProvider == nil {
		if autoMode {
			// In auto mode, return gracefully instead of error
			return false, nil
		}
		return false, fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	// Find the worktree for the given branch.
	worktrees, err := WorktreeProvider.List()
	if err != nil {
		if autoMode {
			// Graceful degradation: log and continue
			return false, nil
		}
		return false, fmt.Errorf("list worktrees: %w", err)
	}

	var targetPath string
	for _, wt := range worktrees {
		if wt.Branch == branchName {
			targetPath = wt.Path
			break
		}
	}

	if targetPath == "" {
		// No worktree found - not an error in auto mode
		return true, nil
	}

	// Anchor guard (t46): automation skips removal while a live session is
	// anchored in the tree, rather than killing that session's shell.
	if anchored := session.LiveAnchoredSessions(targetPath, time.Now()); len(anchored) > 0 && !force {
		fmt.Fprintf(os.Stderr, "moai: worktree %s kept: %d live anchored session(s):\n%s\n",
			targetPath, len(anchored), formatAnchored(anchored))
		return false, nil
	}

	// Remove the worktree.
	if err := WorktreeProvider.Remove(targetPath, force); err != nil {
		if autoMode {
			// Graceful degradation: log and continue
			return false, nil
		}
		return false, fmt.Errorf("remove worktree: %w", err)
	}

	// Optionally delete the branch
	if deleteBranch {
		if err := WorktreeProvider.DeleteBranch(branchName); err != nil {
			if autoMode {
				// Graceful degradation: log and continue
				return false, nil
			}
			return false, fmt.Errorf("delete branch: %w", err)
		}
	}

	return true, nil
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

	autoMode, err := cmd.Flags().GetBool("auto")
	if err != nil {
		return fmt.Errorf("get auto flag: %w", err)
	}

	// Handle auto mode (silent execution for automation)
	if autoMode {
		_, err := runDoneWithAutoMode(branchName, force, deleteBranch, true)
		return err
	}

	// Normal interactive mode
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

	// Anchor guard (t46): refuse to remove the tree while a live session is
	// anchored in it. Once the tree is gone, Claude Code's native
	// worktree-isolation guard blocks every Bash call in that session.
	if anchored := session.LiveAnchoredSessions(targetPath, time.Now()); len(anchored) > 0 {
		if !force {
			return fmt.Errorf("ANCHORED_SESSIONS_PRESENT: %d live session(s) anchored in %s:\n%s\n\nClose the session(s) first (their shells die with the tree), or rerun with --force to remove anyway",
				len(anchored), targetPath, formatAnchored(anchored))
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: --force removing %s while %d live session(s) are anchored there:\n%s\n",
			targetPath, len(anchored), formatAnchored(anchored))
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

// formatAnchored renders one line per live anchored session for refusal and
// warning output.
func formatAnchored(entries []session.Entry) string {
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		id := e.SessionID
		if len(id) > 8 {
			id = id[:8]
		}
		lines = append(lines, fmt.Sprintf("  - session %s pid=%d spec=%s cwd=%s", id, e.PID, e.SpecID, e.CWD))
	}
	return strings.Join(lines, "\n")
}
