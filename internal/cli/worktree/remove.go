package worktree

// @MX:NOTE: [AUTO] Remove worktree at specified path with optional force flag
// @MX:NOTE: [AUTO] Force flag bypasses uncommitted changes safety check

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/session"
)

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [path]",
		Short: "Remove a worktree",
		Long:  "Remove a Git worktree at the specified path.",
		Args:  cobra.ExactArgs(1),
		RunE:  runRemove,
	}
	cmd.Flags().Bool("force", false, "Force removal even with uncommitted changes")
	return cmd
}

func runRemove(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	wtPath := args[0]

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("get force flag: %w", err)
	}

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	// Resolve to absolute so the anchor guard's registry lookup and cwd
	// comparison work regardless of how the caller spelled the path.
	absPath, aerr := filepath.Abs(wtPath)
	if aerr != nil {
		return fmt.Errorf("resolve path: %w", aerr)
	}

	// Anchor guard (t73): refuse to remove the tree while a live session is
	// anchored in it. Once the tree is gone, Claude Code's native
	// worktree-isolation guard blocks every Bash call in that session.
	if anchored := session.LiveAnchoredSessions(absPath, time.Now()); len(anchored) > 0 {
		if !force {
			return fmt.Errorf("ANCHORED_SESSIONS_PRESENT: %d live session(s) anchored in %s:\n%s\n\nClose the session(s) first (their shells die with the tree), or rerun with --force to remove anyway",
				len(anchored), absPath, formatAnchored(anchored))
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: --force removing %s while %d live session(s) are anchored there:\n%s\n",
			absPath, len(anchored), formatAnchored(anchored))
	}

	if err := WorktreeProvider.Remove(wtPath, force); err != nil {
		return fmt.Errorf("remove worktree: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Removed worktree at %s\n", wtPath)
	return nil
}
