package worktree

// sweep.go — SPEC-WORKTREE-SWEEP-001: post-landing disposal of merged
// worktrees across both tiers.
//
// The sweep disposes a worktree only when its branch is confirmed landed on
// the REMOTE integration base: a fresh `git fetch` of the base followed by a
// three-way `git merge-base --is-ancestor` contract (exit 0 = landed, exit 1
// = not landed, anything else = unanswerable). Every safety predicate is
// AND-ed, and every predicate that cannot be answered PRESERVES the tree —
// absence of a failure signal is never a dispose signal (REQ-WS-005).
//
// The predicates themselves are NOT re-implemented here: this file composes
// the shared internals (worktreeLockStates, ignoredContentVerdict,
// protectedWorktreePaths, isBaseBranch, worktreeHasLocalChanges,
// isL1WorktreePath, hoistBeforeDisposal, runDoneWorktreeCleanupWithOptions,
// session.AnchorDecision) behind two NEW seams — the fetch+ancestry
// observation and the process-cwd probe — plus the verdict record and the
// orchestration loop. Stale-reference sweeping (no unique commits, no
// landing evidence) remains in `clean --stale`; this command disposes only
// remote-landing-confirmed trees.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/session"
)

// The sweep's verdict and tier vocabulary.
const (
	sweepDispose  = "DISPOSE"
	sweepPreserve = "PRESERVE"
	sweepTierL1   = "L1"
	sweepTierL2   = "L2"
)

// Cause tokens for the sweep's verdict reasons, extending the `cause=`
// vocabulary the PR-merge sweep and the clean sweeps share (REQ-WR-023), so
// one grep finds every sweep's notices.
const (
	causeFetchFailed       = "fetch-failed"
	causeLandedCheckFailed = "landed-check-failed"
	causeNotLanded         = "not-landed"
	causeAnchored          = "anchored"
	causeCWDOccupied       = "cwd-occupied"
	causeCWDProbeFailed    = "cwd-probe-failed"
	causeOnBaseBranch      = "on-base-branch"
	causeDetachedHead      = "detached-head"
	causeHoistFailed       = "hoist-failed"
	causeDirtyCheckFailed  = "dirty-check-failed"
	causeDirtyTree         = "dirty-tree"
	causeNotEvaluated      = "not-evaluated"
)

// sweepVerdict is one worktree's audited evaluation — the record
// `sweep --json` emits and the text report renders (REQ-WS-011). The
// predicate fields use clean --stale's four-valued vocabulary
// (staleStateYes / staleStateNo / staleStateUndetermined /
// staleStateNotChecked); an unobserved predicate is never reported as a
// negative.
type sweepVerdict struct {
	Path    string `json:"path"`
	Branch  string `json:"branch"`
	Tier    string `json:"tier"`
	Verdict string `json:"verdict"`
	// Reason is empty when the verdict is DISPOSE; otherwise it names the
	// preserving cause with a `cause=` token.
	Reason string `json:"reason"`
	Landed string `json:"landed"`
	Dirty  string `json:"dirty"`
	// Ignored is the shared ignored-content predicate (REQ-WR-024): "yes"
	// when the tree holds irreplaceable gitignored content.
	Ignored string `json:"ignored"`
	// Anchored follows the clean --stale convention: "no" when no source
	// claimed an anchor, otherwise the source name ("lock" / "registry"),
	// or "undetermined" when the authoritative lock source was unreadable.
	Anchored string `json:"anchored"`
	// CWDOccupied is the process-cwd predicate (REQ-WS-006): "yes" when a
	// live process has its working directory inside the tree,
	// "undetermined" when the probe could not answer.
	CWDOccupied string `json:"cwd_occupied"`
	// OnBase is "yes" when the tree checks out the base branch itself —
	// a member of the never-dispose list (REQ-WS-007).
	OnBase string `json:"on_base"`
}

// sweepProcessCWDs lists the working directories of the host's processes
// (the REQ-WS-006 probe). The platform split lives in sweep_cwd_posix.go /
// sweep_cwd_windows.go; a probe that cannot answer returns an error, which
// the sweep renders as an unanswerable predicate — never a negative.
var sweepProcessCWDs = platformProcessCWDs

// sweepFetchBase refreshes the remote-tracking ref the landing predicate
// compares against (REQ-WS-001: a stale remote-tracking ref never satisfies
// the predicate on its own — SPEC-WORKTREE-GC-001 REQ-WGC-004). Runs ONCE
// per invocation.
var sweepFetchBase = func(repoRoot, base string) error {
	remote, ref := "origin", base
	if i := strings.Index(base, "/"); i >= 0 {
		remote, ref = base[:i], base[i+1:]
	}
	if _, err := gitWorktreeCmd("-C", repoRoot, "fetch", remote, ref); err != nil {
		return fmt.Errorf("fetch %s %s: %w", remote, ref, err)
	}
	return nil
}

// sweepMergeBaseExit runs the ancestry check for one tree and returns the
// subprocess exit code: 0 = the tip is an ancestor of base, 1 = it is not,
// anything else (or a non-exit error) = unanswerable.
var sweepMergeBaseExit = func(treePath, branchTip, base string) (int, error) {
	_, err := gitWorktreeCmd("-C", treePath, "merge-base", "--is-ancestor", branchTip, base)
	return sweepProcessExitCode(err), err
}

// sweepAncestryOutcome maps the merge-base exit code onto the three-way
// landing contract (REQ-WS-001/002/003). It is a pure function so the exit
// semantics are testable without spawning git.
func sweepAncestryOutcome(exit int) (landed bool, determined bool) {
	switch exit {
	case 0:
		return true, true
	case 1:
		return false, true
	default:
		return false, false
	}
}

// sweepAncestor is the per-tree landing observation: the ancestry half of
// REQ-WS-001. A determined result reports landed/not-landed; an undetermined
// one returns the error the verdict renders as cause=landed-check-failed.
var sweepAncestor = func(treePath, branchTip, base string) (bool, error) {
	exit, err := sweepMergeBaseExit(treePath, branchTip, base)
	landed, determined := sweepAncestryOutcome(exit)
	if !determined {
		return false, fmt.Errorf("merge-base --is-ancestor %s %s exited %d: %w", branchTip, base, exit, err)
	}
	return landed, nil
}

// sweepProcessExitCode extracts a subprocess exit code; -1 when err is not
// an exit error (a failed exec, a killed process).
var sweepProcessExitCode = func(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

// sweepHoistBeforeDisposal is the L1 evidence-retrieval seam (REQ-WS-009):
// the tree's .moai/reports/ is copied into the project root BEFORE removal,
// and a hoist failure preserves the tree.
var sweepHoistBeforeDisposal = hoistBeforeDisposal

// sweepDoneCleanup is the L2 disposal seam (REQ-WS-008): the done removal
// core with built-in evidence hoist, anchor refusal, non-forced removal, and
// no branch deletion.
var sweepDoneCleanup = runDoneWorktreeCleanupWithOptions

func newSweepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sweep",
		Short: "Dispose worktrees whose branches are confirmed landed on the remote base",
		Long: `Dispose worktrees whose branches are confirmed landed on the remote
integration base, across both L1 (.claude/worktrees/ tier) and L2 (registry
tier) trees, hoisting evidence before every disposal.

The landing observation is deliberately stricter than the other sweeps': a
fresh git fetch of the base remote followed by merge-base --is-ancestor,
with a THREE-WAY contract — ancestry exit 0 means landed, exit 1 means not
landed, and ANY other exit (fetch failure, unresolvable base) means the
sweep cannot answer, which PRESERVES the tree. A stale remote-tracking ref
never satisfies the predicate on its own.

The default base is origin/develop — this diverges from clean --stale,
whose default is origin/main (that flag sweeps stale references, not
landings). Override with --base.

Every safety predicate is AND-ed and fail-safe: not protected, not locked,
not anchored by a live session, no live process cwd inside, clean working
tree, no irreplaceable ignored content, and remotely landed. Any predicate
that cannot be answered preserves the tree. Branches are never deleted.
Previews by default; pass --yes to remove.`,
		RunE: runSweep,
	}
	cmd.Flags().Bool("yes", false, "Actually perform the disposals instead of previewing them")
	cmd.Flags().Bool("json", false, "Report every non-protected worktree's evaluation as JSON; removes nothing")
	cmd.Flags().String("base", "origin/develop", "Remote integration base the landing check compares against (default diverges from clean --stale's origin/main)")
	return cmd
}

func runSweep(cmd *cobra.Command, _ []string) error {
	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	base, _ := cmd.Flags().GetString("base")
	apply, _ := cmd.Flags().GetBool("yes")
	asJSON, _ := cmd.Flags().GetBool("json")

	worktrees, err := WorktreeProvider.List()
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}

	records, lockErr := classifySweepVerdicts(worktrees, base)

	if asJSON {
		if err := renderSweepJSON(cmd, records); err != nil {
			return err
		}
		return cleanDegradedExit(lockErr)
	}
	renderSweepReport(cmd, records)
	if !apply {
		return cleanDegradedExit(lockErr)
	}
	applySweepVerdicts(cmd, records)
	return cleanDegradedExit(lockErr)
}

// sweepEvalInputs carries the per-run observation results one evaluation
// consumes. locks is nil exactly when lockErr != nil (the lock source could
// not be read — the degraded-run shape).
type sweepEvalInputs struct {
	base     string
	fetchErr error
	cwdDirs  []string
	cwdErr   error
	locks    map[string]session.LockInfo
	lockErr  error
	now      time.Time
}

// classifySweepVerdicts evaluates every non-protected worktree once. It is
// the SINGLE evaluation behind the text report, the --json inventory, and
// the --yes apply path (REQ-WS-012: one evaluation, rendered twice — the
// inventory and the sweep can never disagree).
//
// Protected trees — the main checkout and the worktree this command runs in
// — are absent from the result entirely, not reported as kept: they are
// outside the sweep's universe rather than candidates it declined.
func classifySweepVerdicts(worktrees []git.Worktree, base string) ([]sweepVerdict, error) {
	protected := protectedWorktreePaths()
	locks, lockErr := worktreeLockStates()
	cwdDirs, cwdErr := sweepProcessCWDs()
	fetchErr := sweepFetchBase(WorktreeProvider.Root(), base)

	inputs := sweepEvalInputs{
		base:     base,
		fetchErr: fetchErr,
		cwdDirs:  cwdDirs,
		cwdErr:   cwdErr,
		locks:    locks,
		lockErr:  lockErr,
		now:      time.Now(),
	}

	records := []sweepVerdict{}
	for _, wt := range worktrees {
		if protected[filepath.Clean(wt.Path)] {
			continue
		}
		records = append(records, sweepEvaluate(wt, inputs))
	}
	return records, lockErr
}

// sweepEvaluate evaluates one worktree's predicate chain. The first
// predicate that preserves the tree ends the chain — later predicates stay
// "not-checked" (never a negative) — and DISPOSE requires every predicate to
// have been asked and answered affirmatively.
func sweepEvaluate(wt git.Worktree, in sweepEvalInputs) sweepVerdict {
	v := sweepVerdict{
		Path:        wt.Path,
		Branch:      wt.Branch,
		Tier:        sweepTierOf(wt.Path),
		Verdict:     sweepPreserve,
		Landed:      staleStateNotChecked,
		Dirty:       staleStateNotChecked,
		Ignored:     staleStateNotChecked,
		Anchored:    staleStateNo,
		CWDOccupied: staleStateNotChecked,
		OnBase:      staleStateNotChecked,
	}

	// The authoritative lock source could not be read at all: an unobserved
	// anchor is UNDETERMINED, never a negative — every tree is preserved.
	if in.lockErr != nil {
		v.Anchored = staleStateUndetermined
		v.Reason = fmt.Sprintf("cause=%s; could not read the worktree lock state: %v", causeLockSourceUnreadable, in.lockErr)
		return v
	}

	anchor := session.AnchorDecision(wt.Path, in.locks[filepath.Clean(wt.Path)], in.now)
	if anchor.Anchored {
		v.Anchored = string(anchor.Source)
		v.Reason = fmt.Sprintf("cause=%s; live session anchored in this worktree — %s (source: %s)", causeAnchored, anchor.Detail, anchor.Source)
		return v
	}

	if wt.Branch == "" {
		// No branch tip: the landing predicate is unanswerable, not negative.
		v.Landed = staleStateUndetermined
		v.Reason = fmt.Sprintf("cause=%s; detached HEAD (no branch tip to compare against %s)", causeDetachedHead, in.base)
		return v
	}
	if isBaseBranch(wt.Branch, in.base) {
		v.OnBase = staleStateYes
		v.Reason = fmt.Sprintf("cause=%s; checked out on the base branch", causeOnBaseBranch)
		return v
	}

	// M2 placeholder: the landing predicate is evaluated in the next
	// milestone; until then every tree is preserved as not-evaluated.
	v.Reason = fmt.Sprintf("cause=%s", causeNotEvaluated)
	return v
}

// sweepTierOf classifies a tree's tier from its path: L1 when it sits under
// the main root's .claude/worktrees/ or .moai/worktrees/ (the
// SPEC-WORKTREE-DONE-TIER-001 predicate), L2 otherwise. The classification
// is computed for every record — it is informational, never a disposal
// decision.
func sweepTierOf(path string) string {
	if isL1WorktreePath(path) {
		return sweepTierL1
	}
	return sweepTierL2
}

// renderSweepJSON emits the machine-readable inventory and removes nothing:
// --json is the SAME evaluation as the text report, rendered twice
// (REQ-WS-012).
func renderSweepJSON(cmd *cobra.Command, records []sweepVerdict) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}

// renderSweepReport renders the text verdict table: one line per evaluated
// tree, every predicate's observed value visible, so an operator can audit
// every decision without inspecting trees (REQ-WS-011).
func renderSweepReport(cmd *cobra.Command, records []sweepVerdict) {
	out := cmd.OutOrStdout()
	if len(records) == 0 {
		_, _ = fmt.Fprintln(out, "Nothing to sweep: 0 worktree(s) evaluated.")
		return
	}
	disposable := 0
	for _, r := range records {
		_, _ = fmt.Fprintf(out, "  %s %s [%s] (%s): landed=%s dirty=%s ignored=%s anchored=%s cwd=%s on_base=%s\n",
			r.Verdict, r.Path, r.Branch, r.Tier,
			r.Landed, r.Dirty, r.Ignored, r.Anchored, r.CWDOccupied, r.OnBase)
		if r.Reason != "" {
			_, _ = fmt.Fprintf(out, "    %s\n", r.Reason)
		}
		if r.Verdict == sweepDispose {
			disposable++
		}
	}
	if disposable == 0 {
		_, _ = fmt.Fprintf(out, "Nothing to sweep: %d worktree(s) evaluated, 0 disposable.\n", len(records))
		return
	}
	_, _ = fmt.Fprintf(out, "\nWould remove %d worktree(s):\n", disposable)
	for _, r := range records {
		if r.Verdict == sweepDispose {
			_, _ = fmt.Fprintf(out, "  %s [%s]\n", r.Path, r.Branch)
		}
	}
	_, _ = fmt.Fprintln(out, "\nThis was a preview. Re-run with --yes to remove them.")
}

// applySweepVerdicts performs the --yes removals. M4 wires the tier routing
// (L1 hoist+remove, L2 done core); the M1 skeleton preserves everything —
// no record can be DISPOSE yet.
func applySweepVerdicts(cmd *cobra.Command, records []sweepVerdict) {
	_ = cmd
	_ = records
}
