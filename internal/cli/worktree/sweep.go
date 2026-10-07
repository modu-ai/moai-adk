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
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/factory"
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
	causeHoistPartial      = "hoist-partial"
	causeDirtyCheckFailed  = "dirty-check-failed"
	causeDirtyTree         = "dirty-tree"
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
	if err == nil {
		return 0, nil // a successful command exits 0 — landed
	}
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

// sweepHoistEvidence is the sweep's evidence-retrieval seam (REQ-WS-009).
// Unlike the plain done-path routine — whose contract reports only hard
// failures — this returns whether the retrieval was COMPLETE: the tree is
// removable only when every evidence file landed at the project root.
var sweepHoistEvidence = sweepHoistEvidenceImpl

// sweepHoistEvidenceImpl runs the shared hoist routine and judges its
// completeness. Two fail-open shapes of the plain routine land here as
// complete=false (nil error): an unresolved project root — no retrieval was
// attempted at all — and a partial copy where destination conflicts were
// skipped per REQ-RLC-006's never-overwrite policy. Either way the tree
// still holds the only copy of something, so it must not be removed.
func sweepHoistEvidenceImpl(w io.Writer, treePath string) (complete bool, err error) {
	mainRoot, rootErr := hoistTargetMainRoot(treePath)
	if rootErr != nil {
		_, _ = fmt.Fprintf(w, "hoist skipped (project root unresolved from %s)\n", treePath)
		return false, nil
	}
	res, hoistErr := hoistWorktreeReports(treePath, mainRoot)
	if hoistErr != nil {
		return false, hoistErr
	}
	printHoistResult(w, res, treePath)
	return len(res.Skipped) == 0, nil
}

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

The default base is the landed ref, resolved through the SAME chain the todo
surface's landing questions answer from (factory.LandedRefForWithLevel): the
configured git_strategy.worktree_base_branch first, then the integration
branch the repository itself records (refs/remotes/origin/HEAD), then the
compiled-in default (origin/main). When the resolved ref is absent on its
remote (an integration-branch cutover deleted it), the derived base falls
back to the remote's own default branch, with a notice on stderr (REQ-CR-001).
This diverges from clean --stale,
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
	cmd.Flags().String("base", "", "Remote integration base the landing check compares against (default: the chain-resolved landed ref — git_strategy.worktree_base_branch, else refs/remotes/origin/HEAD, else origin/main)")
	return cmd
}

// sweepRemoteRefExists probes whether refs/heads/<ref> exists on the remote
// (REQ-CR-001): `git ls-remote --exit-code` answers structurally — exit 0
// present, exit 2 absent, anything else undeterminable — instead of parsing
// fetch stderr text, which is not stable across git versions.
var sweepRemoteRefExists = func(repoRoot, remote, ref string) (bool, error) {
	_, err := gitWorktreeCmd("-C", repoRoot, "ls-remote", "--exit-code", remote, "refs/heads/"+ref)
	if err == nil {
		return true, nil
	}
	if sweepProcessExitCode(err) == 2 {
		return false, nil
	}
	return false, err
}

// sweepRemoteHead resolves the remote's default branch name from the remote
// itself (REQ-CR-003): the first `ref: refs/heads/<name>` line of
// `git ls-remote --symref <remote> HEAD` — the same answer `git clone` uses
// to pick its initial branch, so no branch name is hardcoded here.
var sweepRemoteHead = func(repoRoot, remote string) (string, error) {
	out, err := gitWorktreeCmd("-C", repoRoot, "ls-remote", "--symref", remote, "HEAD")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(out, "\n") {
		const prefix = "ref: refs/heads/"
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		rest := strings.TrimPrefix(line, prefix)
		if i := strings.IndexAny(rest, "\t "); i >= 0 {
			rest = rest[:i]
		}
		if rest != "" {
			return rest, nil
		}
	}
	return "", fmt.Errorf("ls-remote --symref %s HEAD: no symbolic ref line", remote)
}

// sweepEffectiveBase resolves the DERIVED default base against the remote
// (SPEC-CUTOVER-RESIDUE-001 REQ-CR-002/003): a derived ref the remote no
// longer carries (a workflow cutover deleting the old integration branch)
// makes every tree's landing predicate fail with cause=fetch-failed, so the
// absent-ref case falls back to the remote's own default branch. Every case
// the probe cannot answer affirmatively keeps the derived base — the
// existing three-way landing contract then renders it honestly (fetch
// failure → PRESERVE). The second return reports that a fallback engaged;
// the caller surfaces it so the switch is never silent. An explicit --base
// never reaches this function: the operator's word is the base (REQ-CR-004).
func sweepEffectiveBase(repoRoot, base string) (string, bool) {
	i := strings.Index(base, "/")
	if i <= 0 || i == len(base)-1 {
		return base, false
	}
	remote, ref := base[:i], base[i+1:]
	exists, err := sweepRemoteRefExists(repoRoot, remote, ref)
	if err != nil || exists {
		return base, false
	}
	head, err := sweepRemoteHead(repoRoot, remote)
	if err != nil || head == "" || head == ref {
		return base, false
	}
	return remote + "/" + head, true
}

// sweepConfigRoot names the project root the default --base is derived from.
// A seam in the file's style (sweepFetchBase, sweepAncestor) so tests whose
// provider root is a fake path can point it at a fixture root.
var sweepConfigRoot = func() string { return WorktreeProvider.Root() }

// sweepDefaultBase derives the default --base from the landed-ref chain the
// todo surface answers from (factory.LandedRefForWithLevel —
// SPEC-GITHUB-FLOW-CI-RESIDUE-001 REQ-GFC-001): the configured
// git_strategy.worktree_base_branch, then refs/remotes/origin/HEAD, then the
// compiled-in default (origin/main). The chain always answers — there is no
// "no configured target" error path anymore; --base keeps precedence over it.
func sweepDefaultBase(root string) string {
	ref, _ := factory.LandedRefForWithLevel(root)
	return ref
}

func runSweep(cmd *cobra.Command, _ []string) error {
	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	base, _ := cmd.Flags().GetString("base")
	apply, _ := cmd.Flags().GetBool("yes")
	asJSON, _ := cmd.Flags().GetBool("json")

	if strings.TrimSpace(base) == "" {
		derived := sweepDefaultBase(sweepConfigRoot())
		effective, fellBack := sweepEffectiveBase(WorktreeProvider.Root(), derived)
		if fellBack {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "base %s is absent on its remote; using the remote default branch %s as the sweep base\n", derived, effective)
		}
		base = effective
	}

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
	v.OnBase = staleStateNo

	// Landing predicate (REQ-WS-001/002/003). The fetch ran once for the
	// whole invocation; its failure makes every tree's landing unanswerable.
	// The ancestry check is the three-way contract: exit 0 = landed, exit 1
	// = not landed (the ONE negative this predicate may assert), any other
	// exit = unanswerable. A stale remote-tracking ref never satisfies the
	// predicate on its own — the fetch above is what makes it an observation.
	if in.fetchErr != nil {
		v.Landed = staleStateUndetermined
		v.Reason = fmt.Sprintf("cause=%s; could not fetch the %s remote: %v", causeFetchFailed, in.base, in.fetchErr)
		return v
	}
	landed, err := sweepAncestor(wt.Path, wt.Branch, in.base)
	if err != nil {
		v.Landed = staleStateUndetermined
		v.Reason = fmt.Sprintf("cause=%s; %v", causeLandedCheckFailed, err)
		return v
	}
	if !landed {
		// Ancestry said "no" — the one negative. A squash-merged card is
		// never an ancestor, so the shared predicate's layers 2-3 get the
		// last word (landing_predicate.go); nothing confirmed keeps the
		// verdict exactly as it was.
		landed, _ = landedBeyondAncestry(wt.Path, wt.Branch, in.base)
	}
	if !landed {
		v.Landed = staleStateNo
		v.Reason = fmt.Sprintf("cause=%s; branch has commits not in %s", causeNotLanded, in.base)
		return v
	}
	v.Landed = staleStateYes

	// Dirty predicate: uncommitted or untracked content (the work git itself
	// would refuse to discard without --force).
	dirty, err := worktreeHasLocalChanges(wt.Path)
	if err != nil {
		v.Dirty = staleStateUndetermined
		v.Reason = fmt.Sprintf("cause=%s; could not read working tree state: %v", causeDirtyCheckFailed, err)
		return v
	}
	if dirty {
		v.Dirty = staleStateYes
		v.Reason = fmt.Sprintf("cause=%s; uncommitted or untracked changes", causeDirtyTree)
		return v
	}
	v.Dirty = staleStateNo

	// Ignored-content predicate — the SHARED decision (REQ-WR-024), composed
	// with the sweep's hoist-aware filter (see sweepIgnoredReason).
	state, ignoreReason := sweepIgnoredReason(wt.Path)
	v.Ignored = state
	if ignoreReason != "" {
		v.Reason = ignoreReason
		return v
	}

	// Process-cwd predicate (REQ-WS-006): a live process sitting in the tree
	// dies with it. An unanswerable probe preserves; it is never a negative.
	if in.cwdErr != nil {
		v.CWDOccupied = staleStateUndetermined
		v.Reason = fmt.Sprintf("cause=%s; %v", causeCWDProbeFailed, in.cwdErr)
		return v
	}
	if sweepCWDOccupied(in.cwdDirs, wt.Path) {
		v.CWDOccupied = staleStateYes
		v.Reason = fmt.Sprintf("cause=%s; a live process has its working directory inside this tree (process-cwd probe)", causeCWDOccupied)
		return v
	}
	v.CWDOccupied = staleStateNo

	// Every predicate asked and answered affirmatively: the tree is a
	// disposal candidate.
	v.Verdict = sweepDispose
	v.Reason = ""
	return v
}

// sweepIgnoredReason is the ignored-content predicate as the SWEEP evaluates
// it: the shared decision (ignoredContentVerdict, REQ-WR-024) composed with
// one sweep-specific filter — the `.moai/reports/` class is NOT irreplaceable
// here, because the apply path hoists exactly that class into the project
// root BEFORE every removal (REQ-WS-009): the evidence survives the
// disposal, which is the concern the predicate protects. Everything else
// stays fail-closed — a card tree holding `.claude/agent-memory/` is
// preserved exactly as `clean --stale` preserves it.
func sweepIgnoredReason(path string) (state, reason string) {
	state, reason = ignoredContentVerdict(path)
	if reason == "" {
		return state, ""
	}
	porcelain, err := gitWorktreeCmd("-C", path, "status", "--porcelain", "--ignored")
	if err != nil {
		return state, reason // the shared verdict already renders the failure
	}
	stillIrreplaceable := make([]string, 0, 2)
	for _, entry := range session.IrreplaceableIgnoredEntries(porcelain) {
		if entry == ".moai/reports" || strings.HasPrefix(entry, ".moai/reports/") {
			continue // hoisted before removal — the concern is discharged
		}
		stillIrreplaceable = append(stillIrreplaceable, entry)
	}
	if len(stillIrreplaceable) == 0 {
		return staleStateNo, ""
	}
	return staleStateYes, fmt.Sprintf("cause=%s; irreplaceable gitignored content: %s", causeIgnoredContent, strings.Join(stillIrreplaceable, ", "))
}

// sweepCWDOccupied reports whether any observed process cwd lies at or
// inside the tree. Both sides are symlink-canonicalized where resolvable
// (the probe reports resolved paths; a recorded cwd may not be).
func sweepCWDOccupied(cwdDirs []string, tree string) bool {
	for _, dir := range cwdDirs {
		if isInsideRoot(dir, tree) {
			return true
		}
	}
	return false
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

// applySweepVerdicts performs the --yes removals, routing by tier: L1 trees
// hoist evidence then remove (REQ-WS-009); L2 trees dispose through the done
// removal core (REQ-WS-008). A removal failure is a non-blocking notice —
// the remaining trees still process (REQ-WS-013).
//
// @MX:WARN: [AUTO] bulk disposal loop — every guard here is load-bearing
// @MX:REASON: this loop deletes directories across both tiers; dropping the
// removal-time ignored-content re-read, the full-retrieval hoist check, the
// hoist-before-remove order, or the non-forced removal turns a routine sweep
// into silent evidence or work loss.
func applySweepVerdicts(cmd *cobra.Command, records []sweepVerdict) {
	out := cmd.OutOrStdout()
	removed := 0
	for _, r := range records {
		if r.Verdict != sweepDispose {
			continue
		}
		// Removal-time re-read of the ignored-content predicate (design §E),
		// through the same hoist-aware filter the classification used:
		// classification happened for the whole population first, so a tree
		// cleared early can acquire irreplaceable ignored content — a session
		// writing .claude/agent-memory/ — before its turn comes. Ignored
		// files trigger no refusal from non-forced removal, so this is the
		// only guard consulted at this distance. The window is narrowed, not
		// closed.
		if _, reason := sweepIgnoredReason(r.Path); reason != "" {
			_, _ = fmt.Fprintf(out, "  Keeping %s [%s]: %s (observed at removal time)\n", r.Path, r.Branch, reason)
			continue
		}
		// FULL-RETRIEVAL GUARD (REQ-WS-009), both tiers: the tree is removed
		// only when its evidence retrieval was COMPLETE. The plain done-path
		// routine reports only hard failures, so a PARTIAL copy (destination
		// conflicts skipped per REQ-RLC-006's never-overwrite policy) or an
		// unattempted one (project root unresolved) would otherwise remove a
		// tree still holding the only copy of the skipped evidence. A hard
		// hoist failure preserves with cause=hoist-failed; an incomplete
		// retrieval preserves with cause=hoist-partial. Running the check
		// sweep-side ahead of the L2 done core is safe: the core's own hoist
		// then re-copies nothing (identical destinations are a no-op under
		// REQ-RLC-006).
		complete, hoistErr := sweepHoistEvidence(cmd.ErrOrStderr(), r.Path)
		if hoistErr != nil {
			_, _ = fmt.Fprintf(out, "  Keeping %s [%s]: cause=%s; %v\n", r.Path, r.Branch, causeHoistFailed, hoistErr)
			continue
		}
		if !complete {
			_, _ = fmt.Fprintf(out, "  Keeping %s [%s]: cause=%s; evidence retrieval was partial or unattempted — the tree still holds the only copy\n", r.Path, r.Branch, causeHoistPartial)
			continue
		}
		switch r.Tier {
		case sweepTierL1:
			// Non-forced, never a branch deletion (REQ-WS-007/010).
			if err := WorktreeProvider.Remove(r.Path, false); err != nil {
				_, _ = fmt.Fprintf(out, "  Warning: could not remove %s: %v\n", r.Path, err)
				continue
			}
		default:
			// REQ-WS-008: L2 trees dispose through the done removal core —
			// built-in evidence hoist, anchor refusal, non-forced removal,
			// and no branch deletion.
			success, err := sweepDoneCleanup(r.Branch, false, false, true)
			if err != nil {
				_, _ = fmt.Fprintf(out, "  Warning: could not remove %s [%s]: %v\n", r.Path, r.Branch, err)
				continue
			}
			if !success {
				_, _ = fmt.Fprintf(out, "  Keeping %s [%s]: the done core kept the tree\n", r.Path, r.Branch)
				continue
			}
		}
		_, _ = fmt.Fprintf(out, "  Removing swept worktree: %s [%s]\n", r.Path, r.Branch)
		removed++
	}
	if removed > 0 {
		// Reclaim launch-ledger rows left dead by the removals (card t297).
		pruneLaunchLedgerAfterDisposal(out, cmd.ErrOrStderr())
		_, _ = fmt.Fprintf(out, "Removed %d worktree(s). Branches were left intact.\n", removed)
	}
}
