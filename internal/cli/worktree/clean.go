package worktree

// @MX:NOTE: [AUTO] Clean stale worktree references and merged branch worktrees
// @MX:NOTE: [AUTO] --merged-only flag removes only fully merged branches
// @MX:NOTE: [AUTO] --stale flag sweeps abandoned worktrees that hold nothing to lose

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/core/git"
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
	cmd.Flags().Bool("json", false, "Report every non-protected worktree and its state as JSON; removes nothing")
	// The base default is origin/main, not the local `main`: a local main that
	// is behind the remote reports fewer branches as merged, so the two sweeps
	// in this repository would otherwise disagree about the same worktree.
	// prMergeCleanup compares against origin/main (REQ-WR-022).
	cmd.Flags().String("base", "origin/main", "Base branch for --merged-only and --stale checks")
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
		asJSON, _ := cmd.Flags().GetBool("json")
		if asJSON {
			// REQ-WR-013: the reporting path removes nothing. --json is a
			// report, so it overrides --yes rather than combining with it —
			// an inventory that could delete on a stray flag is not an
			// inventory.
			return reportStaleWorktrees(cmd, base)
		}
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

	locks := worktreeLockStates()

	var removed int
	for _, wt := range worktrees {
		if wt.Branch == "" || isBaseBranch(wt.Branch, base) {
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
			// live lane's tree. It consumes the SHARED lock-∪-registry
			// decision (REQ-WR-019): the registry alone was measured to name
			// 1 of 5 live anchors.
			if v := session.AnchorDecision(wt.Path, locks[filepath.Clean(wt.Path)], time.Now()); v.Anchored {
				_, _ = fmt.Fprintf(out, "  Keeping %s [%s]: live session(s) anchored — %s (source: %s)\n", wt.Path, wt.Branch, v.Detail, v.Source)
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

// staleCandidate is one worktree the --stale sweep classified. It is also the
// JSON record `clean --stale --json` emits (REQ-WR-012), so the inventory and
// the sweep can never disagree: they are the same evaluation, rendered twice.
type staleCandidate struct {
	Path   string `json:"path"`
	Branch string `json:"branch"`
	// KeepReason is empty when the worktree is safe to remove; otherwise it
	// states why the sweep is keeping it.
	KeepReason string `json:"keep_reason"`
	// Dirty, Merged, and Anchored carry the three predicates behind
	// KeepReason. Each is one of the staleState* values below — including
	// "not-checked", which is deliberately distinct from "undetermined":
	// the first means the sweep short-circuited before asking, the second
	// means it asked and could not tell.
	Dirty  string `json:"dirty"`
	Merged string `json:"merged"`
	// Anchored is "no" when no source claimed an anchor, otherwise the name
	// of the source that did ("lock" or "registry").
	Anchored string `json:"anchored"`
}

// The tri-state (plus not-checked) vocabulary the JSON record uses. An
// unobserved predicate is never reported as a negative.
const (
	staleStateYes          = "yes"
	staleStateNo           = "no"
	staleStateUndetermined = "undetermined"
	staleStateNotChecked   = "not-checked"
)

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

	candidates := classifyStaleWorktrees(worktrees, base)

	var removable []staleCandidate
	var kept []staleCandidate
	for _, c := range candidates {
		if c.KeepReason == "" {
			removable = append(removable, c)
		} else {
			kept = append(kept, c)
		}
	}

	for _, c := range kept {
		_, _ = fmt.Fprintf(out, "  Keeping %s [%s]: %s\n", c.Path, c.Branch, c.KeepReason)
	}

	if len(removable) == 0 {
		_, _ = fmt.Fprintln(out, "No stale worktrees to clean.")
		return nil
	}

	if !apply {
		_, _ = fmt.Fprintf(out, "\nWould remove %d stale worktree(s):\n", len(removable))
		for _, c := range removable {
			_, _ = fmt.Fprintf(out, "  %s [%s]\n", c.Path, c.Branch)
		}
		_, _ = fmt.Fprintln(out, "\nThis was a preview. Re-run with --yes to remove them.")
		return nil
	}

	var removed int
	for _, c := range removable {
		_, _ = fmt.Fprintf(out, "  Removing stale worktree: %s [%s]\n", c.Path, c.Branch)
		if err := WorktreeProvider.Remove(c.Path, false); err != nil {
			_, _ = fmt.Fprintf(out, "  Warning: could not remove %s: %v\n", c.Path, err)
			continue
		}
		removed++
	}
	_, _ = fmt.Fprintf(out, "Removed %d stale worktree(s). Branches were left intact.\n", removed)
	return nil
}

// classifyStaleWorktrees evaluates every non-protected worktree once. It is
// the SINGLE evaluation behind both the human sweep and the `--json`
// inventory (REQ-WR-012), so the inventory can never describe a tree the sweep
// would treat differently.
//
// Protected trees — the main checkout and the worktree this command runs in —
// are absent from the result entirely, not reported as kept: they are outside
// the sweep's universe rather than candidates it declined.
func classifyStaleWorktrees(worktrees []git.Worktree, base string) []staleCandidate {
	protected := protectedWorktreePaths()
	locks := worktreeLockStates()

	var candidates []staleCandidate
	for _, wt := range worktrees {
		if protected[filepath.Clean(wt.Path)] {
			continue
		}
		c := staleCandidate{
			Path: wt.Path, Branch: wt.Branch,
			Dirty: staleStateNotChecked, Merged: staleStateNotChecked, Anchored: staleStateNo,
		}
		anchor := session.AnchorDecision(wt.Path, locks[filepath.Clean(wt.Path)], time.Now())
		if anchor.Anchored {
			c.Anchored = string(anchor.Source)
		}

		switch {
		case anchor.Anchored:
			// Anchor guard (t73): a live session's shell dies with its tree.
			// The decision is the SHARED lock-∪-registry one (REQ-WR-019).
			c.KeepReason = fmt.Sprintf("live session anchored in this worktree — %s (source: %s)", anchor.Detail, anchor.Source)
		case wt.Branch == "":
			c.KeepReason = "detached HEAD (no branch to compare against base)"
		case isBaseBranch(wt.Branch, base):
			c.KeepReason = "checked out on the base branch"
		default:
			c.KeepReason = staleKeepReason(&c, wt.Path, wt.Branch, base)
		}

		candidates = append(candidates, c)
	}
	return candidates
}

// isBaseBranch reports whether a worktree's LOCAL branch is the base branch.
//
// The base is a remote-tracking ref by default (`origin/main`, REQ-WR-022)
// while a worktree checks out a LOCAL branch (`main`), so a literal comparison
// stops recognising the base checkout the moment the default changed — and a
// second worktree sitting on `main` would then be evaluated by the merge
// predicate, which reports `main` as merged into `origin/main` and makes it
// removable. Comparing the trailing segment as well keeps that guard standing.
//
// The comparison errs toward KEEPING: a local branch that merely shares its
// name with the base's trailing segment is treated as the base and preserved.
func isBaseBranch(branch, base string) bool {
	if branch == "" {
		return false
	}
	if branch == base {
		return true
	}
	if i := strings.LastIndex(base, "/"); i >= 0 {
		return branch == base[i+1:]
	}
	return false
}

// reportStaleWorktrees emits the machine-readable inventory (REQ-WR-012) and
// removes nothing (REQ-WR-013). It covers EVERY non-protected registered
// worktree — worktree-ness is a checkout property, not a branch-name one — so
// the report reaches trees no branch-name glob would find.
func reportStaleWorktrees(cmd *cobra.Command, base string) error {
	if err := WorktreeProvider.Prune(); err != nil {
		return fmt.Errorf("prune worktrees: %w", err)
	}
	worktrees, err := WorktreeProvider.List()
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}
	candidates := classifyStaleWorktrees(worktrees, base)
	if candidates == nil {
		candidates = []staleCandidate{}
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(candidates)
}

// staleKeepReason returns the reason a worktree must be kept, or "" when both
// safety conditions hold and it is safe to remove. It records each predicate's
// observed value on c as it goes, so an unobserved predicate is reported as
// undetermined rather than silently as a negative.
func staleKeepReason(c *staleCandidate, path, branch, base string) string {
	dirty, err := worktreeHasLocalChanges(path)
	if err != nil {
		c.Dirty = staleStateUndetermined
		return fmt.Sprintf("could not read working tree state: %v", err)
	}
	c.Dirty = staleStateNo
	if dirty {
		c.Dirty = staleStateYes
		return "uncommitted or untracked changes"
	}

	merged, err := WorktreeProvider.IsBranchMerged(branch, base)
	if err != nil {
		c.Merged = staleStateUndetermined
		return fmt.Sprintf("could not compare against %s: %v", base, err)
	}
	c.Merged = staleStateNo
	if !merged {
		return fmt.Sprintf("branch has commits not in %s", base)
	}
	c.Merged = staleStateYes
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

// worktreeLockStates reads the git worktree lock state for every registered
// worktree, keyed by cleaned path. The lock is the authoritative anchor source
// (SPEC-WORKTREE-REAPER-001 REQ-WR-006); the git.WorktreeManager listing does
// not carry it, so it is read straight from the porcelain.
//
// Fail-open on the READ, fail-closed on the DECISION: an unreadable porcelain
// yields an empty map, so every tree falls back to the registry source alone —
// the pre-repair behaviour — rather than wedging the sweep.
func worktreeLockStates() map[string]session.LockInfo {
	porcelain, err := gitWorktreeCmd("worktree", "list", "--porcelain")
	if err != nil {
		return nil
	}
	locks := make(map[string]session.LockInfo)
	for path, info := range session.ParseWorktreeLocks(porcelain) {
		locks[filepath.Clean(path)] = info
	}
	return locks
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
