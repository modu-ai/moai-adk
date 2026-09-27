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
	//
	// The decision is the SHARED lock-and-registry one the clean sweeps use:
	// a tree anchored by its git worktree lock alone (a `moai codex -w`
	// session registers nowhere else) is refused here with its anchor source
	// named, rather than reaching git and failing with git's own message.
	now := time.Now()
	if anchored := session.LiveAnchoredSessions(absPath, now); len(anchored) > 0 {
		if !force {
			return fmt.Errorf("ANCHORED_SESSIONS_PRESENT: %d live session(s) anchored in %s:\n%s\n\nClose the session(s) first (their shells die with the tree), or rerun with --force to remove anyway",
				len(anchored), absPath, formatAnchored(anchored))
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: --force removing %s while %d live session(s) are anchored there:\n%s\n",
			absPath, len(anchored), formatAnchored(anchored))
	} else if lock, lockErr := removalTargetLock(absPath); lockErr != nil || lock.Locked {
		verdict := session.AnchorDecision(absPath, lock, now)
		detail := fmt.Sprintf("%s (source: %s, holder: %s)", verdict.Detail, verdict.Source, lock.Reason)
		if lockErr != nil {
			// Unread is undetermined, never unlocked.
			verdict.Anchored = true
			detail = fmt.Sprintf("cause=%s; %v (source: lock, holder: undetermined)", causeLockSourceUnreadable, lockErr)
		}
		if verdict.Anchored {
			if !force {
				return fmt.Errorf("ANCHORED_SESSIONS_PRESENT: live session anchored in %s - %s\n\nClose that session first (its shell dies with the tree), or rerun with --force to remove anyway",
					absPath, detail)
			}
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: --force removing %s while a live session is anchored there - %s\n", absPath, detail)
		}
	}

	if err := WorktreeProvider.Remove(wtPath, force); err != nil {
		return fmt.Errorf("remove worktree: %w", err)
	}

	// Reclaim the launch-ledger row(s) that died with this tree (card t297).
	// Best-effort: a prune failure warns, never fails the removal.
	pruneLaunchLedgerAfterDisposal(out, cmd.ErrOrStderr())

	_, _ = fmt.Fprintf(out, "Removed worktree at %s\n", wtPath)
	return nil
}

// removalTargetLock returns the git worktree lock of the removal target, read
// from the same porcelain listing the clean sweeps use. git reports resolved
// paths, so the target is matched both as spelled and symlink-resolved.
func removalTargetLock(absPath string) (session.LockInfo, error) {
	locks, err := worktreeLockStates()
	if err != nil {
		return session.LockInfo{}, err
	}
	if lock, ok := locks[filepath.Clean(absPath)]; ok {
		return lock, nil
	}
	if resolved, rerr := filepath.EvalSymlinks(absPath); rerr == nil {
		return locks[filepath.Clean(resolved)], nil
	}
	return session.LockInfo{}, nil
}
