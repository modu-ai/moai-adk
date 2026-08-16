package worktree

// @MX:NOTE: [AUTO] Clean stale worktree references and merged branch worktrees
// @MX:NOTE: [AUTO] --merged-only flag removes only fully merged branches
// @MX:NOTE: [AUTO] --stale flag sweeps abandoned worktrees that hold nothing to lose

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/session"
)

func newCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean stale worktree references",
		Long: `Remove stale worktree references and optionally remove worktrees
whose branches have been merged into the base branch.

--stale sweeps abandoned worktrees that hold nothing to lose: a clean working
tree (no uncommitted or untracked files) AND no commits of its own beyond the
base branch. Worktrees that would lose work are reported and kept, and so are
worktrees anchoring a live session (tree-local registry check). Branches are
never deleted. --stale previews by default; pass --yes to actually remove.`,
		RunE: runClean,
	}
	cmd.Flags().Bool("merged-only", false, "Only remove worktrees whose branches are merged into base")
	cmd.Flags().Bool("stale", false, "Remove abandoned worktrees that are clean and hold no unique commits (preview unless --yes)")
	cmd.Flags().Bool("yes", false, "Actually perform the --stale removals instead of previewing them")
	cmd.Flags().String("base", "main", "Base branch for --merged-only and --stale checks")
	return cmd
}

func runClean(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	mergedOnly, _ := cmd.Flags().GetBool("merged-only")
	stale, _ := cmd.Flags().GetBool("stale")

	if stale && mergedOnly {
		return fmt.Errorf("--stale and --merged-only are mutually exclusive")
	}

	if stale {
		base, _ := cmd.Flags().GetString("base")
		apply, _ := cmd.Flags().GetBool("yes")
		return cleanStaleWorktrees(cmd, base, apply)
	}

	if mergedOnly {
		base, _ := cmd.Flags().GetString("base")
		return cleanMergedWorktrees(cmd, base)
	}

	if err := WorktreeProvider.Prune(); err != nil {
		return fmt.Errorf("prune worktrees: %w", err)
	}

	_, _ = fmt.Fprintln(out, wtSuccessCard("Cleaned stale worktree references"))
	return nil
}

// cleanMergedWorktrees removes worktrees whose branches are fully merged.
func cleanMergedWorktrees(cmd *cobra.Command, base string) error {
	out := cmd.OutOrStdout()

	worktrees, err := WorktreeProvider.List()
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}

	var removed int
	for _, wt := range worktrees {
		if wt.Branch == "" || wt.Branch == base {
			continue
		}
		merged, err := WorktreeProvider.IsBranchMerged(wt.Branch, base)
		if err != nil {
			_, _ = fmt.Fprintf(out, "  Warning: could not check %s: %v\n", wt.Branch, err)
			continue
		}
		if merged {
			// Anchor guard (t73): --merged-only has no dirty guard of its
			// own, so this is the only protection between the sweep and a
			// live lane's tree.
			if anchored := session.LiveAnchoredSessions(wt.Path, time.Now()); len(anchored) > 0 {
				_, _ = fmt.Fprintf(out, "  Keeping %s [%s]: live session(s) anchored\n", wt.Path, wt.Branch)
				continue
			}
			_, _ = fmt.Fprintf(out, "  Removing merged worktree: %s [%s]\n", wt.Path, wt.Branch)
			if err := WorktreeProvider.Remove(wt.Path, false); err != nil {
				_, _ = fmt.Fprintf(out, "  Warning: could not remove %s: %v\n", wt.Path, err)
				continue
			}
			removed++
		}
	}

	if removed == 0 {
		_, _ = fmt.Fprintln(out, "No merged worktrees to clean.")
	} else {
		_, _ = fmt.Fprintf(out, "Removed %d merged worktree(s).\n", removed)
	}
	return nil
}

// staleCandidate is one worktree the --stale sweep classified.
type staleCandidate struct {
	path   string
	branch string
	// keepReason is empty when the worktree is safe to remove; otherwise it
	// states why the sweep is keeping it.
	keepReason string
}

// cleanStaleWorktrees sweeps abandoned worktrees that hold nothing to lose.
//
// A worktree qualifies only when BOTH safety conditions hold: its working tree
// is clean (no uncommitted or untracked files) AND its branch carries no commits
// of its own beyond base. Anything else is kept and reported, so the sweep can
// never destroy work. Branches are never deleted — the commits stay reachable
// by branch name after the worktree directory is gone.
//
// The sweep previews by default; apply=true (from --yes) performs the removals.
//
// @MX:ANCHOR: [AUTO] --stale removal is gated on a two-part no-work-lost predicate
// @MX:REASON: this command deletes directories in bulk across every registered
// worktree; dropping either the cleanliness check or the unique-commit check turns
// a routine sweep into silent data loss.
func cleanStaleWorktrees(cmd *cobra.Command, base string, apply bool) error {
	out := cmd.OutOrStdout()

	// Prune first so worktrees whose directories are already gone drop out of
	// the listing instead of being reported as unreadable.
	if err := WorktreeProvider.Prune(); err != nil {
		return fmt.Errorf("prune worktrees: %w", err)
	}

	worktrees, err := WorktreeProvider.List()
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}

	protected := protectedWorktreePaths()

	var candidates []staleCandidate
	for _, wt := range worktrees {
		c := staleCandidate{path: wt.Path, branch: wt.Branch}

		switch {
		case protected[filepath.Clean(wt.Path)]:
			// Main checkout or the worktree this command is running in.
			continue
		case len(session.LiveAnchoredSessions(wt.Path, time.Now())) > 0:
			// Anchor guard (t73): a live session's shell dies with its tree.
			c.keepReason = "live session anchored in this worktree"
		case wt.Branch == "":
			c.keepReason = "detached HEAD (no branch to compare against base)"
		case wt.Branch == base:
			c.keepReason = "checked out on the base branch"
		default:
			c.keepReason = staleKeepReason(wt.Path, wt.Branch, base)
		}

		candidates = append(candidates, c)
	}

	var removable []staleCandidate
	var kept []staleCandidate
	for _, c := range candidates {
		if c.keepReason == "" {
			removable = append(removable, c)
		} else {
			kept = append(kept, c)
		}
	}

	for _, c := range kept {
		_, _ = fmt.Fprintf(out, "  Keeping %s [%s]: %s\n", c.path, c.branch, c.keepReason)
	}

	if len(removable) == 0 {
		_, _ = fmt.Fprintln(out, "No stale worktrees to clean.")
		return nil
	}

	if !apply {
		_, _ = fmt.Fprintf(out, "\nWould remove %d stale worktree(s):\n", len(removable))
		for _, c := range removable {
			_, _ = fmt.Fprintf(out, "  %s [%s]\n", c.path, c.branch)
		}
		_, _ = fmt.Fprintln(out, "\nThis was a preview. Re-run with --yes to remove them.")
		return nil
	}

	var removed int
	for _, c := range removable {
		_, _ = fmt.Fprintf(out, "  Removing stale worktree: %s [%s]\n", c.path, c.branch)
		if err := WorktreeProvider.Remove(c.path, false); err != nil {
			_, _ = fmt.Fprintf(out, "  Warning: could not remove %s: %v\n", c.path, err)
			continue
		}
		removed++
	}
	_, _ = fmt.Fprintf(out, "Removed %d stale worktree(s). Branches were left intact.\n", removed)
	return nil
}

// staleKeepReason returns the reason a worktree must be kept, or "" when both
// safety conditions hold and it is safe to remove.
func staleKeepReason(path, branch, base string) string {
	dirty, err := worktreeHasLocalChanges(path)
	if err != nil {
		return fmt.Sprintf("could not read working tree state: %v", err)
	}
	if dirty {
		return "uncommitted or untracked changes"
	}

	merged, err := WorktreeProvider.IsBranchMerged(branch, base)
	if err != nil {
		return fmt.Sprintf("could not compare against %s: %v", base, err)
	}
	if !merged {
		return fmt.Sprintf("branch has commits not in %s", base)
	}
	return ""
}

// worktreeHasLocalChanges reports whether the worktree at path has any
// uncommitted or untracked files. Untracked files count: they are the class of
// work git itself would refuse to discard without --force.
func worktreeHasLocalChanges(path string) (bool, error) {
	out, err := gitWorktreeCmd("-C", path, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// protectedWorktreePaths returns the worktree paths the --stale sweep must
// never remove: the repository root and the worktree this process runs in.
func protectedWorktreePaths() map[string]bool {
	protected := make(map[string]bool, 2)
	if root := WorktreeProvider.Root(); root != "" {
		protected[filepath.Clean(root)] = true
	}
	if cwdRoot, err := gitRepoRootFunc(); err == nil && cwdRoot != "" {
		protected[filepath.Clean(cwdRoot)] = true
	}
	return protected
}
