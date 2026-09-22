package worktree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/core/git"
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
	cmd.Flags().Bool("auto", false, "Auto mode: no success output for automation (e.g., after PR merge); failures still exit non-zero")
	return cmd
}

// AutoCleanupFlag is the flag name for auto-cleanup mode.
// Used by sync workflow to trigger cleanup after PR merge.
const AutoCleanupFlag = "auto"

// runDoneWorktreeCleanup is the shared removal core for the done command.
// Auto mode (--auto) suppresses the SUCCESS output only: a failed removal
// returns an error so automation sees a non-zero exit instead of a silent
// rc=0 (t41 c, 2026-08-15). The two intentional non-error exits are the
// no-worktree case (nothing left to clean is a completed cleanup,
// SPEC-WORKTREE-002 R2) and the anchored-session skip (t46).
//
// @MX:NOTE: SPEC-WORKTREE-002 R2 implementation - auto-cleanup for PR merge workflow
// @MX:SPEC: SPEC-WORKTREE-002
func runDoneWorktreeCleanup(branchName string, force, deleteBranch bool) (success bool, err error) {
	if WorktreeProvider == nil {
		return false, fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	// Find the worktree for the given branch.
	worktrees, err := WorktreeProvider.List()
	if err != nil {
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

	// L1 tier guard (SPEC-WORKTREE-DONE-TIER-001): session worktrees under
	// <mainRoot>/.claude/worktrees/ are never done disposal targets, with
	// or without --force. Ordered BEFORE the anchor guard; neither guard
	// replaces the other.
	if isL1WorktreePath(targetPath) {
		return false, refuseL1SessionWorktree(targetPath)
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
		return false, doneRemoveError(err, targetPath)
	}

	// Reclaim the launch-ledger row(s) that died with this tree (card t297).
	// Quiet form: --auto suppresses success output, never the reclamation.
	pruneLaunchLedgerAfterDisposal(nil, os.Stderr)

	// Optionally delete the branch
	if deleteBranch {
		if err := WorktreeProvider.DeleteBranch(branchName); err != nil {
			return false, fmt.Errorf("delete branch: %w", err)
		}
	}

	return true, nil
}

// doneRemoveError wraps a worktree-removal failure for the done command.
// When the tree was locked it appends the actionable remedy; done never
// escalates its own --force to git's double force, because a lock usually
// means a live session still uses the tree (t41 b).
func doneRemoveError(err error, path string) error {
	if errors.Is(err, git.ErrWorktreeLocked) {
		return fmt.Errorf("remove worktree: %w%s", err, lockGuidance(path))
	}
	return fmt.Errorf("remove worktree: %w", err)
}

// lockGuidance renders the two exits a locked worktree leaves the user:
// unlock and retry, or remove with git's double force. Named after the tree
// so the command lines are copy-pasteable.
func lockGuidance(path string) string {
	return fmt.Sprintf("\n\nThe worktree is locked — usually a live session is still using it:\n"+
		"  unlock and retry:  git worktree unlock %s\n"+
		"  remove anyway:     git worktree remove -f -f %s\n"+
		"moai does not force a locked tree on its own",
		path, path)
}

// gitMainRootFromTargetFunc resolves the MAIN repository root FROM THE
// TARGET worktree path — never from the process CWD. Inside a linked
// worktree the process CWD names the worktree itself, so a CWD-anchored
// resolution would compute a wrong L1 prefix and the tier guard would fail
// open in the primary operating mode (plan-audit iter-1 D1).
//
// Primary mechanism (git >= 2.31): `git -C <targetPath> rev-parse
// --path-format=absolute --git-common-dir`; the main root is the parent of
// the returned common dir. Fallback for older git: the FIRST `worktree`
// stanza of `git -C <targetPath> worktree list --porcelain` — the main
// worktree, present in every porcelain output.
//
// Overridable in tests; the default above is the tested mechanism
// (SPEC-WORKTREE-DONE-TIER-001 M2f mandate: the CWD-independence cell runs
// this default against real git).
var gitMainRootFromTargetFunc = func(targetPath string) (string, error) {
	out, err := gitWorktreeCmd("-C", targetPath, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err == nil {
		if commonDir := strings.TrimSpace(out); commonDir != "" {
			return filepath.Dir(commonDir), nil
		}
	}
	// Fallback for git < 2.31 (--path-format unsupported): parse the first
	// `worktree` stanza of the porcelain listing instead.
	list, listErr := gitWorktreeCmd("-C", targetPath, "worktree", "list", "--porcelain")
	if listErr != nil {
		return "", fmt.Errorf("resolve main repo root from %s: %w", targetPath, err)
	}
	for line := range strings.SplitSeq(list, "\n") {
		if mainPath, ok := strings.CutPrefix(line, "worktree "); ok && mainPath != "" {
			return mainPath, nil
		}
	}
	return "", fmt.Errorf("resolve main repo root from %s: no worktree stanza in porcelain output", targetPath)
}

// canonicalTierPath best-effort canonicalizes a path via EvalSymlinks and
// falls back to the raw path on error (macOS /tmp vs /private/tmp; a path
// that does not exist yet must not fail the comparison).
//
// @MX:NOTE: [AUTO] symlink fallback — EvalSymlinks failure returns the raw path, keeping the predicate total over missing/unresolvable targets
func canonicalTierPath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return path
}

// isL1WorktreePath reports whether targetPath is a session-scoped L1 tree
// under <mainRoot>/.claude/worktrees/. The main root is resolved FROM THE
// TARGET PATH (see gitMainRootFromTargetFunc); both sides are canonicalized
// before the separator-safe prefix comparison. A resolver failure is not
// evidence of tier: the predicate reports false and the command proceeds
// exactly as before the guard (REQ-005 — the L2 flow is unchanged).
//
// @MX:ANCHOR: [AUTO] done L1 tier guard — the single removal-refusal boundary shared by both done paths
// @MX:REASON: runDone and runDoneWorktreeCleanup both route their WorktreeProvider.Remove calls through this predicate; weakening it re-opens the silent L1 data-loss shapes (SPEC-WORKTREE-DONE-TIER-001 §D A1/A3/C), so its target-derived resolution and --force immunity are invariants
// @MX:SPEC: SPEC-WORKTREE-DONE-TIER-001
func isL1WorktreePath(targetPath string) bool {
	mainRoot, err := gitMainRootFromTargetFunc(targetPath)
	if err != nil {
		return false
	}
	target := canonicalTierPath(targetPath)
	l1Root := canonicalTierPath(filepath.Join(mainRoot, ".claude", "worktrees"))
	// Separator-safe boundary: a path equal to the L1 root itself, or a
	// sibling whose name merely shares the prefix as a substring, is not an
	// L1 tree.
	return strings.HasPrefix(target, l1Root+string(os.PathSeparator))
}

// l1Guidance renders the two exits an L1 session worktree leaves the user:
// the session-end keep/remove prompt, or manual git unlock + remove. Named
// after the tree so the command lines are copy-pasteable, matching the
// lockGuidance style.
func l1Guidance(path string) string {
	return fmt.Sprintf("\n\nThe worktree is an L1 session worktree under .claude/worktrees/ — session-scoped, not owned by moai worktree verbs:\n"+
		"  session-end keep/remove prompt:  choose when the owning session exits\n"+
		"  remove manually:                 git worktree unlock %s\n"+
		"                                   git worktree remove %s",
		path, path)
}

// refuseL1SessionWorktree is the ONE shared refusal site for the done
// command: both the interactive path and the --auto core print the REQ-003
// message to stderr through it and return the wrapped error that surfaces
// as a non-zero exit. --force never reaches it as a parameter — the L1
// refusal is not bypassable (REQ-002).
func refuseL1SessionWorktree(path string) error {
	fmt.Fprintf(os.Stderr, "moai: worktree %s kept: L1 session worktree%s\n", path, l1Guidance(path))
	return fmt.Errorf("L1_SESSION_WORKTREE: %s is an L1 session worktree%s", path, l1Guidance(path))
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

	// Handle auto mode: success stays output-silent; failures propagate so
	// the process exits non-zero instead of swallowing the error (t41 c).
	if autoMode {
		_, err := runDoneWorktreeCleanup(branchName, force, deleteBranch)
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

	// L1 tier guard (SPEC-WORKTREE-DONE-TIER-001): same shared refusal site
	// as the --auto core above, so the two paths cannot diverge.
	if isL1WorktreePath(targetPath) {
		return refuseL1SessionWorktree(targetPath)
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
		return doneRemoveError(err, targetPath)
	}

	// Reclaim the launch-ledger row(s) that died with this tree (card t297).
	pruneLaunchLedgerAfterDisposal(out, cmd.ErrOrStderr())

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
