// integration_merge_step.go — the ONE in-window merge step (card t1479,
// SPEC-MERGE-WINDOW-QUEUE-001 REQ-MWQ-017/018): `moai integration merge
// --card <id>` calls it, and `moai factory complete` merges ONLY by calling
// it (REQ-MWQ-019 step 4) — one merge path, no second implementation.
//
// The step holds the window for seconds: it renews the holder's lease,
// applies the queue's liveness drops, walks the pre-merge gates IN THE
// SPEC'S ORDER (each failure a distinct exit code, each pre-merge failure
// releasing the window so the next live ticket is promoted), runs the
// collision check, merges --no-ff of the PINNED SHA (never the branch
// name), verifies the merge commit's tree against the record, and releases.
// It runs no test suite — the re-measure happened outside the window
// (REQ-MWQ-014), keyed to the very tree this step verifies.
//
// Causes 7 and 8 leave the window held for nobody: the step writes the
// hold policy FIRST (a system write naming the cause, and for 8 the merge
// commit SHA it leaves in place for the leader), and only then releases —
// so no later holder is ever promoted onto a half-merged tree (O3's seam
// injects the residue that reaches cause 8b).
package factory

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/factorylane"
)

// Merge exit codes (REQ-MWQ-018's thirteen causes — each row of
// AC-MWQ-018 carries its own code; rows 11a-11c share 11 and 13a-13d
// share 13). 14 and 15 are the holder-refusal codes AC-MWQ-017 scenarios
// 4/5 add: they RELEASE NOTHING (the window is not the caller's to
// release), which is why they are outside the thirteen.
const (
	MergeExitRecordInvalid  = 1  // (1) record invalid under REQ-MWQ-014/015
	MergeExitBaseMoved      = 2  // (2) develop advanced past the record's base — re-measure and re-acquire at the tail
	MergeExitNotDescendant  = 3  // (3) pinned SHA does not descend from the base
	MergeExitTreeMismatch   = 4  // (4) pinned tree differs from the record's tree
	MergeExitLandingRefused = 5  // (5) the shared landing check refused
	MergeExitMergeFailed    = 6  // (6) merge failed, worktree clean after abort
	MergeExitMergeDirty     = 7  // (7) merge failed, worktree still dirty after abort — hold first
	MergeExitPostMerge      = 8  // (8) failure after the merge commit exists — commit left, hold naming the SHA
	MergeExitOther          = 9  // (9) any other error before the merge
	MergeExitNothingToMerge = 10 // (10) pinned SHA equals the record's base
	MergeExitCardGate       = 11 // (11) card gate refused (11a foreign lease, 11b not merge-ready, 11c card mismatch)
	MergeExitWorktreeDirty  = 12 // (12) integration worktree not clean before the merge
	MergeExitCollision      = 13 // (13) added-path collision — refused before any merge call
	MergeExitNotHolder      = 14 // AC-MWQ-017 scenario 4 — refuses with the record unchanged
	MergeExitExpiredLease   = 15 // AC-MWQ-017 scenario 5 — same, holder but lapsed lease
)

// CompleteExitPostMergeTransitionConflict is REQ-MWQ-019's own code: a
// state transition failing after the merge commit exists (complete's step
// 4). Distinct from the thirteen and from the holder refusals.
const CompleteExitPostMergeTransitionConflict = 20

// MergeStepError carries a cause code the CLI promotes to the process exit
// (the ExitCoder chain cmd/moai/main.go already resolves) and a message
// naming what failed. Pre-merge causes release the window before this
// error escapes; the holder refusals release nothing.
type MergeStepError struct {
	Code int
	Msg  string
}

func (e *MergeStepError) Error() string  { return e.Msg }
func (e *MergeStepError) ExitCode() int  { return e.Code }

// MergeExitCode reports the step's cause code carried by err.
func MergeExitCode(err error) (int, bool) {
	var step *MergeStepError
	if errors.As(err, &step) {
		return step.Code, true
	}
	return 0, false
}

func mergeStepErr(code int, format string, args ...any) *MergeStepError {
	return &MergeStepError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

// MergeCardState is what the card gate reads (O4): the SAME read
// REQ-MWQ-019 step 1 performs — one read, shared by the verb's gate and
// complete's gate, so "version as read" is one read in both.
type MergeCardState struct {
	Stage          string // homestate.CardMergeReady or later is required
	LeaseUnexpired bool   // the caller holds the card's unexpired lease
	Version        int    // the card version AS READ — the gate never bumps it
	WorktreePath   string // where the card's tree lives (branch resolution)
}

// MergeStepSeams carries the step's test seams. Every nil field falls back
// to the production behavior, so production callers pass the struct
// zero-valued except for ReadCard.
type MergeStepSeams struct {
	// Git overrides the git runner — every call runs in the integration
	// worktree, so the seam carries the args only; nil is ExecGitRunner.
	Git func(args ...string) (string, error)
	// LandingCheck is SPEC-CANDIDATE-CI-001's REQ-CCI-011 shared check;
	// nil (or the key disabled) is the absent no-op seam.
	LandingCheck func(cardID, sha string) error
	// AfterPrecheck runs after every pre-merge gate has passed and before
	// `git merge` — the O3 injection seam a test uses to present the
	// cause-8b dirty state a real autostash failure leaves.
	AfterPrecheck func()
	// ReadCard performs the card-gate read (O4). Production callers pass
	// the factory-db read complete also uses.
	ReadCard func(cardID string) (MergeCardState, error)
	// Now pins the clock; nil is WindowClock.
	Now func() time.Time
	// Probe overrides the liveness seam; nil is the default probe.
	Probe WindowProcProbe
	// LeaseDuration overrides the lease stamp; 0 is the default.
	LeaseDuration time.Duration
}

func (s *MergeStepSeams) gitRunner(worktree string) func(args ...string) (string, error) {
	if s.Git != nil {
		return s.Git
	}
	runner := factorylane.ExecGitRunner{Dir: worktree}
	return runner.Git
}

func (s *MergeStepSeams) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return WindowClock()
}

func (s *MergeStepSeams) probe() WindowProcProbe {
	if s.Probe.OwnerAlive != nil {
		return s.Probe
	}
	return DefaultWindowProcProbe()
}

// MergeStepInput names the step's operands. Root is where the window
// record and the re-measure store live (the primary checkout);
// IntegrationWorktree is the tree holding the integration branch
// checked out; CardWorktree is the card's tree (its WT- branch is resolved
// there, per the REQ-CCI-004 contract the resolver behind).
type MergeStepInput struct {
	Root                string
	IntegrationWorktree string
	IntegrationBranch   string
	CardID              string
	CallerSessionID     string
}

// RunMergeStep executes the merge step and returns the merge commit SHA.
// Every pre-merge failure releases the window (so the next live ticket is
// promoted) and escapes with its own exit code; the holder refusals leave
// the record untouched and release nothing.
func RunMergeStep(in MergeStepInput, seams MergeStepSeams) (string, error) {
	if in.CallerSessionID == "" {
		return "", mergeStepErr(MergeExitNotHolder, "integration merge: no session id — the holder decision needs an address")
	}
	git := seams.gitRunner(in.IntegrationWorktree)
	now := seams.now()

	// The holder decision (REQ-MWQ-017's first sentence): read the record
	// and refuse a non-holder or an expired-lease holder BEFORE applying
	// any queue mutation — the record's bytes are unchanged, and nothing is
	// released (the window is not ours to release; AC-MWQ-017 scenarios
	// 4 and 5).
	lock, err := ReadIntegrationLock(in.Root)
	if err != nil {
		return "", mergeStepErr(MergeExitOther, "integration merge: read the window record: %v", err)
	}
	if !lock.Held() || lock.SessionID != in.CallerSessionID {
		holder := "nobody"
		if lock.Held() {
			holder = lock.SessionID
		}
		return "", mergeStepErr(MergeExitNotHolder, "integration merge: refused — %s holds the window, not %s (moai integration status reads it)", holder, in.CallerSessionID)
	}
	if lock.LeaseExpired(now) {
		return "", mergeStepErr(MergeExitExpiredLease, "integration merge: refused — your lease expired at %s; re-acquire with --wait", lock.LeaseExpiresAt)
	}

	// The card gate BEFORE any queue mutation continues (REQ-MWQ-018 cause
	// 11 fires with card and branch untouched — but the gate read itself is
	// the one read O4 shares with complete's step 1).
	card, err := readCardState(seams, in.CardID)
	if err != nil {
		return "", mergeStepErr(MergeExitOther, "integration merge: read card %s: %v", in.CardID, err)
	}
	if cardErr := validateCardGate(card, in.CardID, lock.Card); cardErr != nil {
		return "", releaseWindow(in, seams, cardErr)
	}

	// The caller holds the window with an unexpired lease: renew it and
	// apply the queue's liveness drops — the same serialized mutation
	// (REQ-MWQ-017's second sentence).
	lease := seams.LeaseDuration
	if lease == 0 {
		lease = IntegrationLeaseDefault
	}
	if err := UpdateIntegrationWindow(in.Root, func(w *IntegrationLock) error {
		if w.Held() && w.SessionID == in.CallerSessionID {
			StampLease(w, now, lease)
		}
		policy, policyErr := ReadIntegrationWindowPolicy(in.Root)
		if policyErr != nil {
			return policyErr
		}
		RefreshWindow(w, policy, seams.probe(), now, lease)
		return nil
	}); err != nil {
		return "", mergeStepErr(MergeExitOther, "integration merge: refresh the window: %v", err)
	}

	// (12) the integration worktree must be clean before the merge —
	// --untracked-files=all so a hidden untracked byte cannot hide behind
	// the config (O2).
	clean, status, err := gitIntegrationWorktreeClean(in.IntegrationWorktree)
	if err != nil {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitOther, "integration merge: status the integration worktree: %v", err))
	}
	if !clean {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitWorktreeDirty, "integration merge: the integration worktree is not clean (%d status lines); clean it and re-measure, then re-acquire", strings.Count(status, "\n")+1))
	}

	// Resolve the card's WT- branch at the card's tree and pin ONE SHA
	// (REQ-MWQ-017; the resolver behind is REQ-CCI-004's contract).
	cardBranch, err := resolveCardBranch(card.WorktreePath, in.CardID)
	if err != nil {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitOther, "integration merge: %v", err))
	}
	pinned, err := git("rev-parse", "refs/heads/"+cardBranch)
	if err != nil {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitOther, "integration merge: pin %s: %v", cardBranch, err))
	}
	pinned = strings.TrimSpace(pinned)
	pinnedTree, err := git("rev-parse", pinned+"^{tree}")
	if err != nil {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitOther, "integration merge: read the pinned tree: %v", err))
	}
	pinnedTree = strings.TrimSpace(pinnedTree)
	tip, err := git("rev-parse", "refs/heads/"+in.IntegrationBranch)
	if err != nil {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitOther, "integration merge: read the integration tip: %v", err))
	}
	tip = strings.TrimSpace(tip)

	// (1) the re-measure record must be valid for THIS tree (REQ-MWQ-014/015
	// — one verifier, the local form while candidate_ci is absent).
	record, err := ReadRemeasureRecord(in.Root, pinnedTree)
	if err != nil {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitRecordInvalid, "integration merge: %v", err))
	}
	if err := ValidateRemeasureRecord(record); err != nil {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitRecordInvalid, "integration merge: the re-measure for tree %s is invalid: %v", pinnedTree[:12], err))
	}

	// (2) the record's base must still be the integration tip — the
	// re-measure-and-re-acquire code names both SHAs; the lane re-absorbs,
	// re-measures, and re-acquires at the tail.
	if record.Base != tip {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitBaseMoved, "integration merge: the integration branch moved past the record's base %s to %s — re-measure against the new tip, then re-acquire --wait", record.Base[:12], tip[:12]))
	}

	// (10) nothing to merge — the pinned SHA IS the base.
	if pinned == record.Base {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitNothingToMerge, "integration merge: %s equals the record's base — nothing to merge", pinned[:12]))
	}

	// (3) ancestry, (4) tree identity.
	if err := gitFail(git, "merge-base", "--is-ancestor", record.Base, pinned); err != nil {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitNotDescendant, "integration merge: pinned %s does not descend from the record's base %s", pinned[:12], record.Base[:12]))
	}
	if pinnedTree != record.Tree {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitTreeMismatch, "integration merge: pinned tree %s differs from the record's tree %s", pinnedTree[:12], record.Tree[:12]))
	}

	// (5) the shared landing check — absent no-op while candidate_ci is
	// disabled (spec.md §F).
	if seams.LandingCheck != nil {
		if err := seams.LandingCheck(in.CardID, pinned); err != nil {
			return "", releaseWindow(in, seams, mergeStepErr(MergeExitLandingRefused, "integration merge: the landing check refused %s: %v", pinned[:12], err))
		}
	}

	// (13) the collision check — refused before ANY git merge call, every
	// colliding byte untouched.
	colliding, err := FindAddedPathCollisions(in.IntegrationWorktree, tip, pinned)
	if err != nil {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitOther, "integration merge: the collision check errored: %v", err))
	}
	if len(colliding) > 0 {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitCollision, "integration merge: refused — the candidate would overwrite ignored/untracked bytes at %s; remove or commit them, then re-measure", strings.Join(colliding, ", ")))
	}

	// The O3 seam: everything pre-merge passed; a test presents the
	// dirty/autostash state here so the post-merge checks have something to
	// find without a fixture that cause 12 would refuse first.
	if seams.AfterPrecheck != nil {
		seams.AfterPrecheck()
	}

	// The merge, of the PINNED SHA never the branch name (REQ-MWQ-017).
	mergeMsg := fmt.Sprintf("Merge %s into %s (card %s, integration merge)", cardBranch, in.IntegrationBranch, in.CardID)
	if _, err := git("merge", "--no-ff", "-q", "-m", mergeMsg, pinned); err != nil {
		// (6)/(7): abort, then decide by the worktree the abort left.
		_, _ = git("merge", "--abort")
		clean, _, cleanErr := gitIntegrationWorktreeClean(in.IntegrationWorktree)
		if cleanErr != nil || !clean {
			// (7): still dirty after the abort — hold FIRST, then release
			// (REQ-MWQ-018: no later holder is promoted onto this state).
			holdErr := writeMergeHold(in, seams, fmt.Sprintf("merge failed and the worktree is still dirty after abort (card %s)", in.CardID))
			if holdErr != nil {
				return "", mergeStepErr(MergeExitMergeDirty, "integration merge: merge failed, the abort left the worktree dirty, and writing the hold failed: %v (merge error: %v)", holdErr, err)
			}
			return "", releaseWindow(in, seams, mergeStepErr(MergeExitMergeDirty, "integration merge: merge failed and the worktree is still dirty after abort; the window policy is held for the leader: %v", err))
		}
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitMergeFailed, "integration merge: merge failed: %v", err))
	}
	mergeSHA, err := git("rev-parse", "HEAD")
	if err != nil {
		return "", postMergeHold(in, seams, mergeStepErr(MergeExitPostMerge, "integration merge: read the merge HEAD: %v", err), "")
	}
	mergeSHA = strings.TrimSpace(mergeSHA)

	// The post-merge checks (cause 8): the merge commit's tree must equal
	// the record's tree and the worktree must be clean again — including
	// autostash residue, which the porcelain check sees as conflict state.
	mergedTree, err := git("rev-parse", "HEAD^{tree}")
	if err != nil || strings.TrimSpace(mergedTree) != record.Tree {
		got := "unreadable"
		if err == nil {
			got = strings.TrimSpace(mergedTree)[:12]
		}
		return "", postMergeHold(in, seams, mergeStepErr(MergeExitPostMerge, "integration merge: the merge commit's tree %s differs from the record's tree %s", got, record.Tree[:12]), mergeSHA)
	}
	if clean, _, err := gitIntegrationWorktreeClean(in.IntegrationWorktree); err != nil || !clean {
		return "", postMergeHold(in, seams, mergeStepErr(MergeExitPostMerge, "integration merge: the worktree is not clean after the merge (autostash residue included)"), mergeSHA)
	}

	// Success: release the window so the next live ticket is promoted.
	if err := releaseHeldWindow(in, seams); err != nil {
		// The merge commit exists — a release failure is the post-merge
		// class, not a pre-merge one.
		return "", postMergeHold(in, seams, mergeStepErr(MergeExitPostMerge, "integration merge: releasing the window failed: %v", err), mergeSHA)
	}
	return mergeSHA, nil
}

// readCardState calls the gate read and wraps its absence.
func readCardState(seams MergeStepSeams, cardID string) (MergeCardState, error) {
	if seams.ReadCard == nil {
		return MergeCardState{}, errors.New("no card reader wired (production callers pass the factory-db read; tests script it)")
	}
	return seams.ReadCard(cardID)
}

// validateCardGate is REQ-MWQ-019 step 1's predicate (the same read O4
// pins): merge-ready, the caller's own unexpired card lease, and the
// requested card equal to the window record's card. The version rides in
// the state as read — the step never bumps it. The stage literal is
// homestate.CardMergeReady's value; the string comparison keeps this
// package from importing the card store.
func validateCardGate(card MergeCardState, requestedCard, windowCard string) error {
	if windowCard != "" && windowCard != requestedCard {
		return mergeStepErr(MergeExitCardGate, "integration merge: refused — the window records card %s, not %s (11c)", windowCard, requestedCard)
	}
	if card.Stage != "merge-ready" {
		return mergeStepErr(MergeExitCardGate, "integration merge: refused — card %s is %q, not merge-ready (11b)", requestedCard, card.Stage)
	}
	if !card.LeaseUnexpired {
		return mergeStepErr(MergeExitCardGate, "integration merge: refused — the caller does not hold card %s's unexpired lease (11a)", requestedCard)
	}
	return nil
}

// resolveCardBranch resolves the card's WT- branch at the card's tree: the
// branch checked out there, which MUST carry the WT- prefix (kanban
// dispatch § Isolation). The card worktree path is the read the card gate
// handed over — the REQ-CCI-004 contract resolves the tree, this resolves
// its branch.
func resolveCardBranch(cardWorktree, cardID string) (string, error) {
	if strings.TrimSpace(cardWorktree) == "" {
		return "", fmt.Errorf("card %s records no worktree, so its WT- branch cannot be resolved", cardID)
	}
	out, err := execGitIn(cardWorktree, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("read the branch of %s: %v", cardWorktree, err)
	}
	branch := strings.TrimSpace(out)
	if !strings.HasPrefix(branch, "WT-") {
		return "", fmt.Errorf("card %s's tree %s is on %q, not a WT- branch", cardID, cardWorktree, branch)
	}
	return branch, nil
}

// releaseWindow releases the window the step held and returns the original
// error — every pre-merge cause passes through here, so the next live
// ticket is promoted no matter which gate refused (REQ-MWQ-018).
func releaseWindow(in MergeStepInput, seams MergeStepSeams, cause error) error {
	if err := releaseHeldWindow(in, seams); err != nil {
		return fmt.Errorf("%w (releasing the window also failed: %v — moai integration status reads it)", cause, err)
	}
	return cause
}

// releaseHeldWindow releases the caller's hold through the queue-aware
// release: the first live ticket is promoted in the same mutation
// (REQ-MWQ-018's "causes 1-6 and 9-13 ... promote the next live ticket on
// release").
func releaseHeldWindow(in MergeStepInput, seams MergeStepSeams) error {
	_, err := ReleaseIntegrationLock(in.Root, in.CallerSessionID, 0, false)
	return err
}

// writeMergeHold writes the cause-7 hold BEFORE the release: a system
// write (REQ-MWQ-018 — the step and card as setter, not subject to the
// lane-role refusal).
func writeMergeHold(in MergeStepInput, seams MergeStepSeams, reason string) error {
	return WriteIntegrationWindowPolicy(in.Root, IntegrationWindowPolicy{
		Policy: PolicyHold,
		Reason: reason,
		SetBy:  fmt.Sprintf("integration-merge-step (card %s)", in.CardID),
		SetAt:  seams.now().Format(time.RFC3339),
	})
}

// postMergeHold is the cause-8 outcome: the merge commit exists, so it
// stays in place; the hold names the cause AND the merge SHA the leader
// inspects; then the window releases. The cause's message carries the same
// extension the hold's reason does — one message, both surfaces.
func postMergeHold(in MergeStepInput, seams MergeStepSeams, cause *MergeStepError, mergeSHA string) error {
	if mergeSHA != "" {
		cause.Msg = fmt.Sprintf("%s (merge commit %s left in place for the leader)", cause.Msg, mergeSHA[:12])
	}
	if err := writeMergeHold(in, seams, cause.Msg); err != nil {
		return fmt.Errorf("%w (writing the hold also failed: %v)", cause, err)
	}
	if err := releaseHeldWindow(in, seams); err != nil {
		return fmt.Errorf("%w (releasing the window also failed: %v)", cause, err)
	}
	return cause
}

// execGitIn runs one git command in dir (the small seams use this instead
// of the runner type so the step's call sites stay uniform).
func execGitIn(dir string, args ...string) (string, error) {
	runner := factorylane.ExecGitRunner{Dir: dir}
	return runner.Git(args...)
}

// gitFail reports an error only when the git call FAILS — the ancestry
// probe answers through its exit code, and a zero exit is the pass.
func gitFail(git func(args ...string) (string, error), args ...string) error {
	_, err := git(args...)
	return err
}
