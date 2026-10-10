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
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

func (e *MergeStepError) Error() string { return e.Msg }
func (e *MergeStepError) ExitCode() int { return e.Code }

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
	State          string // the card's CURRENT state — a moved-on card keeps its stage fields (t1576 review round 5)
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
	// DeferRelease (REQ-MWQ-019 step 4): complete calls the step with the
	// release DEFERRED — its state transitions finish first, and complete
	// releases the window itself (or, on a post-merge transition conflict,
	// holds first and then releases). The verb leaves it false and the step
	// releases as its final act.
	DeferRelease bool
}

// RunMergeStep executes the merge step and returns the merge commit SHA.
// Every pre-merge failure releases the window (so the next live ticket is
// promoted) and escapes with its own exit code; the holder refusals leave
// the record untouched and release nothing.
//
// @MX:WARN: [AUTO] an ordered 13-cause gate table read in one body — decision points well over the complexity-15 warn bar (measured 39 at the t1479 landing; the t1576 in-section card re-gate adds an if chain of its own, and the next mx scan re-measures the [AUTO] count)
// @MX:REASON: the complexity is inherent to REQ-MWQ-017/018's prescribed gate ORDER (holder → card gate → clean check → collision → pinned SHA → merge → post-merge); reordering or rewriting it is the data-loss and double-merge hazard, so changes go gate-by-gate with the cause table in view.
// @MX:SPEC: SPEC-MERGE-WINDOW-QUEUE-001
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
		// t1576 review round 1: a store read failure is a pre-merge cause
		// like any other — the window releases so the next live ticket is
		// promoted, instead of the hold parking until the lease lapses.
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitOther, "integration merge: read card %s: %v", in.CardID, err))
	}
	if cardErr := validateCardGate(card, in.CardID, lock.Card); cardErr != nil {
		return "", releaseWindow(in, seams, cardErr)
	}

	// The caller holds the window with an unexpired lease: renew it and
	// apply the queue's liveness drops — the same serialized mutation
	// (REQ-MWQ-017's second sentence).
	//
	// F7 (card-review r3): an unset seam falls back to WindowLeaseDuration —
	// the CONFIGURED duration the CLI verbs initialize (the release-path
	// promotion and the status refresh stamp with it too) — never straight
	// to the shipped default. The former fallback re-enabled a lease the
	// project had configured AWAY (lease_minutes: 0, the disabled lease the
	// acquire verb honors) the moment the merge renewed it.
	lease := seams.LeaseDuration
	if lease == 0 {
		lease = WindowLeaseDuration
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

	// (12) the dirty precondition depends on the surface (plan §B4, §B5). The
	// primary checkout keeps the fully-clean rule: its uncommitted changes may
	// belong to other sessions, and an overlap test cannot judge a path whose
	// owner it cannot identify. A separate integration worktree has no foreign
	// owners, so the overlap decision applies there; it needs the pinned SHA and
	// runs after the pin below (REQ-LMF-005). The surface is the integration
	// worktree being the primary checkout that Root names.
	primary := sameIntegrationTree(in.IntegrationWorktree, in.Root)
	if primary {
		// --untracked-files=all so a hidden untracked byte cannot hide behind
		// the config (O2).
		clean, status, err := gitIntegrationWorktreeClean(in.IntegrationWorktree)
		if err != nil {
			return "", releaseWindow(in, seams, mergeStepErr(MergeExitOther, "integration merge: status the integration worktree: %v", err))
		}
		if !clean {
			return "", releaseWindow(in, seams, mergeStepErr(MergeExitWorktreeDirty, "integration merge: the integration worktree is not clean (%d status lines); clean it and re-measure, then re-acquire. Move the uncommitted or untracked changes into a card worktree and commit them there, or ask the session that owns them to commit or discard them. Do not stash: the stash is repository-wide.", strings.Count(status, "\n")+1))
		}
	}

	// F1 (sync audit, card t1616): an unfinished merge in the integration worktree
	// belongs to another actor, so refuse it here on both surfaces, before the
	// overlap decision and before any merge call. The clean check above already
	// refuses unmerged entries; a MERGE_HEAD with a clean tree passes it, and the
	// abort after a failed merge would discard that merge.
	inFlight, probeErr := unfinishedMergeReason(git, in.IntegrationWorktree)
	if probeErr != nil {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitOther, "integration merge: the unfinished-merge check errored: %v", probeErr))
	}
	if inFlight != "" {
		return "", releaseWindow(in, seams, unfinishedMergeRefusal(inFlight))
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

	// (12, separate worktree) the overlap decision of plan §B5 (REQ-LMF-005):
	// refuse when a dirty path equals or nests with a target path, or when an
	// ignored file sits at a target path. The pinned SHA is needed for the
	// target set, so this runs after the pin.
	if !primary {
		reason, err := integrationOverlapRefusal(git, pinned)
		if err != nil {
			return "", releaseWindow(in, seams, mergeStepErr(MergeExitOther, "integration merge: the overlap decision errored: %v", err))
		}
		if reason != "" {
			return "", releaseWindow(in, seams, mergeStepErr(MergeExitWorktreeDirty, "integration merge: refused — %s. Move or commit those changes, re-measure, then re-acquire. Do not stash: the stash is repository-wide.", reason))
		}
	}

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

	// Class C (card-review r2) + r3 F4/F5/F9: the OWNERSHIP RE-VERIFICATION,
	// the late collision re-probe, and the `git merge` subprocess run inside
	// ONE mutation critical section. The former shape re-verified ownership
	// in a mutation that RELEASED the lock before the merge ran as an
	// unserialized subprocess — a --force takeover landing in between still
	// produced the previous holder's merge commit (F4, reproduced: "merge
	// commit ... left in place for the leader"). Nothing re-examined the
	// worktree between the collision check above and the merge, so an
	// ignored file at an added path created in that window was silently
	// overwritten (F5; git refuses untracked clobber, overwrites ignored
	// bytes). The section holds the mutation lock across the merge: a
	// takeover waits for it and then refuses on holdership, and the
	// re-probe closes the check→merge gap. The clock is read INSIDE the
	// section (F9) — the recheck's own purpose is "expired mid-step".
	mergeMsg := fmt.Sprintf("Merge %s into %s (card %s, integration merge)", cardBranch, in.IntegrationBranch, in.CardID)
	var recheckErr *MergeStepError
	var mergeErr error
	var headReadErr error
	var mergeDirtyAfterAbort bool
	// mergeForeignErr is set when the failed merge left another actor's MERGE_HEAD in
	// place: the step refuses and does not abort (F1 follow-up).
	var mergeForeignErr *MergeStepError
	var mergeSHA string
	// REQ-LMF-006 (plan §B5): S-before, set inside the section, and whether the
	// status set changed across the merge. The primary surface's baseline stays nil,
	// the empty set its fully-clean precondition established.
	var statusBefore []statusRecord
	var statusChanged bool
	sectionErr := withIntegrationLockMutation(in.Root, func() error {
		w, readErr := ReadIntegrationLock(in.Root)
		if readErr != nil {
			recheckErr = mergeStepErr(MergeExitOther, "integration merge: re-read the window record: %v", readErr)
			return nil
		}
		if !w.Held() || w.SessionID != in.CallerSessionID {
			recheckErr = mergeStepErr(MergeExitNotHolder, "integration merge: refused — the window was taken mid-step; the merge will not proceed on ownership it does not hold")
			return nil
		}
		// t1576 review round 1: the session id alone is not the acquisition —
		// the same session re-acquiring the window for another branch or
		// worktree mid-check passed a session comparison while the merge ran
		// on targets the record no longer names. The merge inputs must still
		// be the acquisition's.
		// t1576 review round 5: the worktree comparison reads DIRECTORIES,
		// not strings — macOS presents /var/... and /private/var/... for one
		// directory, and the string form refused a window the caller
		// legitimately acquired.
		if w.Branch != in.IntegrationBranch || !sameIntegrationTree(w.Worktree, in.IntegrationWorktree) {
			recheckErr = mergeStepErr(MergeExitNotHolder, "integration merge: refused — the window now names branch %s at %s, not the %s at %s the merge was acquired for; re-acquire", w.Branch, w.Worktree, in.IntegrationBranch, in.IntegrationWorktree)
			return nil
		}
		// t1576 review round 2: the record and the inputs are both
		// bookkeeping — the merge lands on whatever branch the integration
		// worktree has CHECKED OUT, and on whatever its branch tip has moved
		// to. A checkout switched after the pre-checks merged onto the wrong
		// branch while the step returned success. Both are re-read inside
		// the section and either drift aborts before the merge (the tip
		// drift takes cause 2 — the same re-measure-and-re-acquire remedy).
		checkedOut, checkoutErr := git("symbolic-ref", "HEAD")
		if checkoutErr != nil {
			recheckErr = mergeStepErr(MergeExitOther, "integration merge: read the integration worktree's checkout (a detached checkout is its own refusal): %v", checkoutErr)
			return nil
		}
		// t1576 review round 14: the FULL ref is what compares — with a tag
		// named like the integration branch, the abbreviated name
		// disambiguates to heads/<branch> and the string comparison refused
		// a correct checkout.
		if checkedOut = strings.TrimSpace(checkedOut); checkedOut != "refs/heads/"+in.IntegrationBranch {
			recheckErr = mergeStepErr(MergeExitOther, "integration merge: refused — the integration worktree is on %q, not %s; restore the checkout, re-measure, and re-acquire", checkedOut, in.IntegrationBranch)
			return nil
		}
		tipNow, tipErr := git("rev-parse", "refs/heads/"+in.IntegrationBranch)
		if tipErr != nil {
			recheckErr = mergeStepErr(MergeExitOther, "integration merge: re-read the integration tip: %v", tipErr)
			return nil
		}
		if tipNow = strings.TrimSpace(tipNow); tipNow != tip {
			recheckErr = mergeStepErr(MergeExitBaseMoved, "integration merge: refused — the integration branch moved mid-step (tip %s as gated, %s now); re-measure against the new tip, then re-acquire", tip[:12], tipNow[:12])
			return nil
		}
		recheckNow := seams.now()
		if w.LeaseExpired(recheckNow) {
			recheckErr = mergeStepErr(MergeExitExpiredLease, "integration merge: refused — your lease expired mid-step; re-acquire with --wait")
			return nil
		}
		StampLease(w, recheckNow, lease)
		if writeErr := writeIntegrationLock(integrationLockPath(in.Root), w); writeErr != nil {
			return writeErr
		}
		// F5: the collision probe re-runs immediately before the merge, in
		// the same serialized section — an ignored byte at an added path
		// created after the first check is caught here, with every colliding
		// byte still untouched.
		colliding, err := FindAddedPathCollisions(in.IntegrationWorktree, tip, pinned)
		if err != nil {
			recheckErr = mergeStepErr(MergeExitOther, "integration merge: the collision check errored: %v", err)
			return nil
		}
		if len(colliding) > 0 {
			recheckErr = mergeStepErr(MergeExitCollision, "integration merge: refused — the candidate would overwrite ignored/untracked bytes at %s; remove or commit them, then re-measure", strings.Join(colliding, ", "))
			return nil
		}
		// t1576: the card gate re-runs INSIDE the section, immediately
		// before the merge — the gate at the top read the card once, and a
		// card-lease invalidation (or a record mutation) landing between
		// that read and the merge still produced the merge commit (the
		// lane-9 turn-end gate overlay reproduction; t1572's tree base
		// 067fdced2). The section already closes the holder and collision
		// gaps (F4/F5); the card gate is the deliver-half chain's remaining
		// member (t1542 r5-7's principle: re-check what you act on at the
		// point of effect). The version rides in the state as read and the
		// step never bumps it, so a drift between the two reads is the
		// record changing mid-step — the same cause-11 refusal class.
		recheckCard, err := readCardState(seams, in.CardID)
		if err != nil {
			recheckErr = mergeStepErr(MergeExitOther, "integration merge: re-read card %s: %v", in.CardID, err)
			return nil
		}
		if cardErr := validateCardGate(recheckCard, in.CardID, w.Card); cardErr != nil {
			var gateErr *MergeStepError
			if !errors.As(cardErr, &gateErr) {
				gateErr = mergeStepErr(MergeExitCardGate, "integration merge: the card gate re-run refused %s: %v", in.CardID, cardErr)
			}
			recheckErr = gateErr
			return nil
		}
		if recheckCard.Version != card.Version {
			recheckErr = mergeStepErr(MergeExitCardGate, "integration merge: refused — card %s's record changed mid-step (version %d as gated, %d as re-read); re-acquire", in.CardID, card.Version, recheckCard.Version)
			return nil
		}
		// S-before (REQ-LMF-006, plan §B5): the separate surface takes it here, in
		// the section, immediately before the merge. The primary surface keeps the
		// nil baseline set above.
		if !primary {
			before, statusErr := integrationStatusSet(git)
			if statusErr != nil {
				recheckErr = mergeStepErr(MergeExitOther, "integration merge: read the status set before the merge: %v", statusErr)
				return nil
			}
			statusBefore = before
		}

		// F1 (sync audit, card t1616): the unfinished-merge probe is the last read
		// before the merge subprocess, inside the same section. A merge already in
		// progress belongs to another actor: git refuses to start one over it. A merge
		// that begins after this probe is caught from MERGE_HEAD after the failed merge,
		// before any abort (the failure branch below).
		pendingMerge, pendingErr := unfinishedMergeReason(git, in.IntegrationWorktree)
		if pendingErr != nil {
			recheckErr = mergeStepErr(MergeExitOther, "integration merge: the unfinished-merge probe errored: %v", pendingErr)
			return nil
		}
		if pendingMerge != "" {
			recheckErr = unfinishedMergeRefusal(pendingMerge)
			return nil
		}

		// The merge, of the PINNED SHA never the branch name (REQ-MWQ-017),
		// inside the section (F4).
		//
		// t1576 review round 1 asked for --no-overwrite-ignore at the effect
		// point. It is passed — and probed INEFFECTIVE on the path this step
		// always takes: current git enforces it on the fast-forward update
		// only, while --no-ff (REQ-MWQ-017's merge-commit contract) forces
		// the three-way (ort) path, which overwrites an ignored byte at an
		// added path whatever the flag says (probe: ff rc=1 refused;
		// --no-ff and true-3way both rc=0, byte overwritten). The in-section
		// re-probe above catches every byte present before the merge
		// subprocess starts; the residual below is what remains.
		//
		// @MX:DEBT: an ignored byte at an added path created inside the re-probe→merge span is overwritten by the three-way merge — the flag does not reach the ort path on current git
		// @MX:CEILING: the window is one git-subprocess spawn inside a flock-serialized section in a policy-protected integration worktree; the post-merge tree check cannot see it (a worktree-only loss, the commit's tree is unchanged)
		// @MX:UPGRADE: a git whose ort path honors --no-overwrite-ignore (then this flag starts refusing), or an in-process merge that checks ignored paths atomically with the write
		// --no-autostash is git's switch for merge.autoStash=false on this call
		// (plan §B5): with autostash on, a staged change is stashed and re-applied
		// unstaged, which changes the status set by itself.
		if _, err := git("merge", "--no-ff", "--no-autostash", "--no-overwrite-ignore", "-q", "-m", mergeMsg, pinned); err != nil {
			mergeErr = err
			// F1 follow-up (sync audit, card t1616): abort only a merge this call
			// started. The in-section probe ran before the merge call, so a merge another
			// actor began after it shows up here, in MERGE_HEAD, naming a commit other than
			// the pin. Its abort would discard that merge, so the step refuses and changes
			// nothing instead.
			foreignName, foreign, ownerErr := foreignMergeHead(git, in.IntegrationWorktree, pinned)
			if ownerErr != nil {
				mergeForeignErr = mergeStepErr(MergeExitOther, "integration merge: read MERGE_HEAD after the failed merge: %v", ownerErr)
				return nil
			}
			if foreign {
				mergeForeignErr = unfinishedMergeRefusal(fmt.Sprintf("MERGE_HEAD names another commit %s", foreignName[:12]))
				return nil
			}
			// F1 same-SHA (sync audit, card t1616): foreignMergeHead above passes a MERGE_HEAD
			// that names the pin, so a same-SHA merge another actor began after the probe
			// would reach the abort below. git exits 1 only when this call began the merge and
			// stopped with it in progress (a conflict, or a pre-merge-commit hook). Any other
			// exit gives this call no claim on a MERGE_HEAD it finds (git refuses to start
			// before it writes one, observed on git 2.54.0; a call killed by a signal is not
			// credited either), so the step refuses and does not abort it.
			if code, didNotStop := mergeCallDidNotStop(err); didNotStop {
				present, presentErr := mergeHeadPresent(git, in.IntegrationWorktree)
				if presentErr != nil {
					mergeForeignErr = mergeStepErr(MergeExitOther, "integration merge: read MERGE_HEAD after the failed merge: %v", presentErr)
					return nil
				}
				if present {
					mergeForeignErr = unfinishedMergeRefusal(fmt.Sprintf("MERGE_HEAD exists, and this merge call exited %d rather than stopping with exit 1", code))
					return nil
				}
			}
			// (6)/(7): abort, then decide by the status set the abort left against S-before —
			// inside the section, so no acquisition can interleave between
			// our failed merge and its cleanup. The abort reaches this point only when
			// MERGE_HEAD is absent (git's abort then refuses and changes nothing), names
			// the pinned SHA only when this call began that merge (exit 1, checked above),
			// or names no commit at all.
			_, _ = git("merge", "--abort")
			after, statusErr := integrationStatusSet(git)
			mergeDirtyAfterAbort = statusErr != nil || !integrationStatusEqual(after, statusBefore)
			return nil
		}
		sha, shaErr := git("rev-parse", "HEAD")
		if shaErr != nil {
			// The commit exists — only the in-section read failed. The
			// cause-8 class (post-merge) names no SHA it cannot read.
			headReadErr = shaErr
			return nil
		}
		mergeSHA = strings.TrimSpace(sha)
		// S-after (REQ-LMF-006, plan §B5): taken after the merge commit, in the section.
		after, statusErr := integrationStatusSet(git)
		statusChanged = statusErr != nil || !integrationStatusEqual(after, statusBefore)
		return nil
	})
	if sectionErr != nil {
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitOther, "integration merge: the pre-merge serialized section failed: %v", sectionErr))
	}
	if recheckErr != nil {
		switch recheckErr.Code {
		case MergeExitNotHolder, MergeExitExpiredLease:
			// Our hold is gone (or lapsed) — release OUR nothing and refuse.
			// The takeover's window record is untouched (the release refuses
			// on a foreign holder and the refusal path surfaces it).
			if err := releaseHeldWindow(in, seams); err != nil && !IsIntegrationLockNotHeld(err) && !IsIntegrationLockForeign(err) {
				return "", fmt.Errorf("%w (the post-takeover release also failed: %v)", recheckErr, err)
			}
			return "", recheckErr
		default:
			// The re-probe's cause-13 refusal (and any record-read failure)
			// is a pre-merge cause: release so the next live ticket is
			// promoted (REQ-MWQ-018).
			return "", releaseWindow(in, seams, recheckErr)
		}
	}
	if headReadErr != nil {
		return "", postMergeHold(in, seams, mergeStepErr(MergeExitPostMerge, "integration merge: read the merge HEAD: %v", headReadErr), "")
	}
	if mergeErr != nil {
		if mergeForeignErr != nil {
			// F1 follow-up: another actor's merge stays in place and was not aborted;
			// the refusal releases the window through the normal pre-merge path.
			return "", releaseWindow(in, seams, mergeForeignErr)
		}
		if mergeDirtyAfterAbort {
			// (7): still dirty after the abort — hold FIRST, then release
			// (REQ-MWQ-018: no later holder is promoted onto this state).
			holdErr := writeMergeHold(in, seams, fmt.Sprintf("merge failed and the worktree is still dirty after abort (card %s)", in.CardID))
			if holdErr != nil {
				return "", mergeStepErr(MergeExitMergeDirty, "integration merge: merge failed, the abort left the worktree dirty, and writing the hold failed: %v (merge error: %v)", holdErr, mergeErr)
			}
			return "", releaseWindow(in, seams, mergeStepErr(MergeExitMergeDirty, "integration merge: merge failed and the worktree is still dirty after abort; the window policy is held for the leader: %v", mergeErr))
		}
		return "", releaseWindow(in, seams, mergeStepErr(MergeExitMergeFailed, "integration merge: merge failed: %v", mergeErr))
	}

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
	// B3 step 6 (SPEC-LOCAL-MAIN-FLOW-001): SHA equality alone does not prove
	// the branch did not change, because a post-merge hook can move HEAD to
	// another branch at the merge commit. The hold names that merge; the tool
	// never switches HEAD back.
	if head, headErr := git("symbolic-ref", "HEAD"); headErr != nil || strings.TrimSpace(head) != "refs/heads/"+in.IntegrationBranch {
		return "", postMergeHold(in, seams, mergeStepErr(MergeExitPostMerge, "integration merge: HEAD no longer names %s after the merge (it moved off the branch during the merge)", in.IntegrationBranch), mergeSHA)
	}
	if statusChanged {
		msg := "integration merge: the worktree is not clean after the merge (autostash residue included)"
		if !primary {
			msg = "integration merge: the status set changed across the merge (autostash residue included)"
		}
		return "", postMergeHold(in, seams, mergeStepErr(MergeExitPostMerge, "%s", msg), mergeSHA)
	}

	// t1576 review round 1 (F1): the section's card re-gate reads BEFORE the
	// merge; a card transition landing between that read and the merge
	// commit still lands — the integration lock does not serialize the card
	// store. What the step refuses to do is let it land silently: the
	// post-merge re-read surfaces the drift as the post-merge class — the
	// commit stays, the hold names it, the leader decides.
	//
	// @MX:DEBT: the card gate is a point-of-effect re-check, not a cross-store serialization — a transition landing inside the read→merge span is detected and held post-merge (cause 8), not prevented
	// @MX:CEILING: the drift window is one in-section git merge subprocess; anything landing outside it is refused pre-merge by the re-gate
	// @MX:UPGRADE: expose the card-store lock (or an atomic merge reservation) and hold it across the merge — a lock-order design (queue→record, see factory_step_lock.go's @MX:REASON) audited before wiring
	mergedCard, cardErr := readCardState(seams, in.CardID)
	if cardErr != nil {
		return "", postMergeHold(in, seams, mergeStepErr(MergeExitPostMerge, "integration merge: re-read card %s after the merge: %v", in.CardID, cardErr), mergeSHA)
	}
	if mergedCard.Version != card.Version {
		return "", postMergeHold(in, seams, mergeStepErr(MergeExitPostMerge, "integration merge: card %s's record changed mid-merge (version %d as gated, %d as merged) — the commit stays for the leader", in.CardID, card.Version, mergedCard.Version), mergeSHA)
	}

	// Success. The verb releases here; complete (DeferRelease) takes the
	// window's fate with it — its transitions run first, and ITS failure
	// path holds with cause post-merge-transition-conflict naming this
	// merge SHA (REQ-MWQ-019).
	if in.DeferRelease {
		return mergeSHA, nil
	}
	if err := releaseHeldWindow(in, seams); err != nil {
		// The merge commit exists — a release failure is the post-merge
		// class, not a pre-merge one.
		return "", postMergeHold(in, seams, mergeStepErr(MergeExitPostMerge, "integration merge: releasing the window failed: %v", err), mergeSHA)
	}
	return mergeSHA, nil
}

// CompletePostMergeHold is the hold complete writes when a state
// transition fails after the merge commit exists (REQ-MWQ-019's
// post-merge-transition-conflict): a system write naming the cause and the
// merge SHA, released after — so no queued ticket is promoted onto the
// conflict and the leader reads the SHA from the policy record.
func CompletePostMergeHold(projectRoot, cardID, mergeSHA string) error {
	return WriteIntegrationWindowPolicy(projectRoot, IntegrationWindowPolicy{
		Policy: PolicyHold,
		Reason: fmt.Sprintf("post-merge-transition-conflict: the card %s state transition failed after the merge commit %s — the commit stays on the integration branch for the leader", cardID, mergeSHA[:minStrLen(mergeSHA, 12)]),
		SetBy:  fmt.Sprintf("factory-complete (card %s)", cardID),
		SetAt:  WindowClock().Format(time.RFC3339),
	})
}

// minStrLen returns the shorter of the string's length and max — the small
// guard the SHA prefix renders use.
func minStrLen(s string, max int) int {
	if len(s) < max {
		return len(s)
	}
	return max
}

// sameIntegrationTree compares two worktree paths as DIRECTORIES, not
// strings — macOS presents /var/... and /private/var/... for one directory,
// and the string form refused a window the caller legitimately acquired
// (t1576 review round 5). The cli package's factorySameTree carries the
// same semantics across its own boundary; the dependency direction keeps
// this package from importing it, so the shape is mirrored here.
func sameIntegrationTree(a, b string) bool {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	sa, ea := os.Stat(a)
	sb, eb := os.Stat(b)
	if ea == nil && eb == nil {
		if os.SameFile(sa, sb) {
			return true
		}
	}
	// t1576 review round 16: an acquire from a subdirectory records the
	// subdirectory, while the merge passes the worktree root — the same git
	// worktree under two paths. Both sides resolve to their git worktree
	// root before the final comparison.
	if ra := gitToplevelOf(a); ra != "" {
		a = ra
	}
	if rb := gitToplevelOf(b); rb != "" {
		b = rb
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

// gitToplevelOf resolves the git worktree root containing path, or "" when
// path is not inside a git worktree (the caller keeps its original value).
func gitToplevelOf(path string) string {
	out, err := execGitIn(path, "rev-parse", "--show-toplevel")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
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
	// t1576 review round 5: the stage lingers on a card the operator moved
	// on from — an abandoned card kept stage=merge-ready with its lease
	// fields and merged. The CURRENT state is the authority; empty reads as
	// a reader that does not populate it (the scripted tests) and stays
	// admitted.
	if card.State != "" && card.State != "merge-ready" {
		return mergeStepErr(MergeExitCardGate, "integration merge: refused — card %s's current state is %q, not merge-ready (11b; a moved-on card keeps its stage fields)", requestedCard, card.State)
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

// statusRecord is one record of the status set S (plan §B5): its exact bytes
// and every path it names. A rename or copy names its source as well.
type statusRecord struct {
	raw   string
	paths []string
}

// integrationStatusSet reads the status set S of plan §B5: the records of
// `git status --porcelain=v1 -z --untracked-files=all`, sorted bytewise by path.
func integrationStatusSet(git func(args ...string) (string, error)) ([]statusRecord, error) {
	out, err := git("status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return nil, err
	}
	fields := strings.Split(out, "\x00")
	var records []statusRecord
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		if field == "" {
			continue
		}
		if len(field) < 4 || field[2] != ' ' {
			return nil, fmt.Errorf("unrecognized status record %q", field)
		}
		rec := statusRecord{raw: field, paths: []string{field[3:]}}
		if x, y := field[0], field[1]; x == 'R' || x == 'C' || y == 'R' || y == 'C' {
			// A rename or copy carries its source in the next NUL-terminated field.
			if i+1 >= len(fields) {
				return nil, fmt.Errorf("status record %q carries no source path", field)
			}
			i++
			rec.raw = field + "\x00" + fields[i]
			rec.paths = append(rec.paths, fields[i])
		}
		records = append(records, rec)
	}
	sort.SliceStable(records, func(a, b int) bool { return records[a].paths[0] < records[b].paths[0] })
	return records, nil
}

// integrationStatusEqual reports whether two status sets hold the same records,
// as exact record bytes in the same order (plan §B5).
func integrationStatusEqual(a, b []statusRecord) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].raw != b[i].raw {
			return false
		}
	}
	return true
}

// unfinishedMergeReason reports why the integration worktree holds a merge
// another actor began: MERGE_HEAD exists, or the index holds unmerged entries
// (a merge stopped at a conflict, or a stash pop that conflicted). It returns
// "" when neither holds. `git rev-parse --git-path` names MERGE_HEAD in the
// worktree's own git directory, which is right for a linked worktree and for
// the main checkout alike. git answers relative to the directory the runner
// executes in, so a relative answer is joined onto worktree.
func unfinishedMergeReason(git func(args ...string) (string, error), worktree string) (string, error) {
	mergeHead, err := mergeHeadPath(git, worktree)
	if err != nil {
		return "", err
	}
	_, statErr := os.Stat(mergeHead)
	if statErr == nil {
		return "MERGE_HEAD exists", nil
	}
	if !os.IsNotExist(statErr) {
		return "", fmt.Errorf("stat MERGE_HEAD: %w", statErr)
	}
	unmerged, err := git("ls-files", "-u")
	if err != nil {
		return "", fmt.Errorf("list unmerged entries: %w", err)
	}
	if strings.TrimSpace(unmerged) != "" {
		return "unmerged index entries exist", nil
	}
	return "", nil
}

// mergeHeadPath locates the integration worktree's MERGE_HEAD. `git rev-parse
// --git-path` names it in the worktree's own git directory, which is right for a
// linked worktree and for the main checkout alike. git answers relative to the
// directory the runner executes in, so a relative answer is joined onto worktree.
func mergeHeadPath(git func(args ...string) (string, error), worktree string) (string, error) {
	out, err := git("rev-parse", "--git-path", "MERGE_HEAD")
	if err != nil {
		return "", fmt.Errorf("locate MERGE_HEAD: %w", err)
	}
	path := strings.TrimSpace(out)
	if path == "" {
		return "", errors.New("git named no MERGE_HEAD path")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(worktree, path)
	}
	return path, nil
}

// foreignMergeHead reports a full object name in the integration worktree's
// MERGE_HEAD other than sha: a merge another actor began. An absent MERGE_HEAD
// names nothing, and one that names only sha is this call's own merge. A value
// that is not a full object name is not a merge any session began (git writes
// full names), so it is not foreign and the abort proceeds as before.
func foreignMergeHead(git func(args ...string) (string, error), worktree, sha string) (string, bool, error) {
	path, err := mergeHeadPath(git, worktree)
	if err != nil {
		return "", false, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("read MERGE_HEAD: %w", err)
	}
	for _, name := range strings.Fields(string(data)) {
		if name != sha && isFullObjectName(name) {
			return name, true, nil
		}
	}
	return "", false, nil
}

// isFullObjectName reports whether s is a full SHA-1 or SHA-256 object name.
func isFullObjectName(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// unfinishedMergeRefusal is the cause-12 refusal for an integration worktree
// that holds an unfinished merge this step did not start. The step never
// aborts such a merge: finishing or aborting it belongs to the session that
// began it.
func unfinishedMergeRefusal(reason string) *MergeStepError {
	return mergeStepErr(MergeExitWorktreeDirty, "integration merge: refused — the integration worktree holds an unfinished merge (%s) that this step did not start. Finish or resolve it where it was begun, re-measure, then re-acquire. The step does not abort it.", reason)
}

// mergeCallDidNotStop reports the exit status of a failed `git merge` call when that
// call did not stop with the merge in progress. git exits 1 only after the merge has
// begun and stopped (a conflict, or a pre-merge-commit hook that left it unfinished);
// any other status leaves a MERGE_HEAD found afterwards unattributable to this call.
// Only the production runner's *factorylane.GitExitError carries a status, so an error
// of any other shape is not classified here and the caller keeps its abort path.
func mergeCallDidNotStop(err error) (int, bool) {
	var exit *factorylane.GitExitError
	if !errors.As(err, &exit) || exit.ExitCode == 1 {
		return 0, false
	}
	return exit.ExitCode, true
}

// mergeHeadPresent reports whether the integration worktree's MERGE_HEAD exists,
// whichever commit it names.
func mergeHeadPresent(git func(args ...string) (string, error), worktree string) (bool, error) {
	path, err := mergeHeadPath(git, worktree)
	if err != nil {
		return false, err
	}
	_, statErr := os.Stat(path)
	if statErr == nil {
		return true, nil
	}
	if os.IsNotExist(statErr) {
		return false, nil
	}
	return false, fmt.Errorf("stat MERGE_HEAD: %w", statErr)
}

// integrationOverlapRefusal decides the separate-worktree dirty precondition
// (plan §B5) against the pinned SHA. It returns the refusal reason when
// Overlap(D, T) holds or the ignored overlap I is non-empty, and "" when the
// merge may proceed.
func integrationOverlapRefusal(git func(args ...string) (string, error), pinned string) (string, error) {
	base, err := git("merge-base", "HEAD", pinned)
	if err != nil {
		var exit *factorylane.GitExitError
		if errors.As(err, &exit) && exit.ExitCode == 1 {
			// No common ancestor: the ancestry gate refuses this pin before any
			// merge call, so no overlap decision is owed here.
			return "", nil
		}
		return "", fmt.Errorf("merge-base of HEAD and %s: %w", pinned, err)
	}
	status, err := integrationStatusSet(git)
	if err != nil {
		return "", fmt.Errorf("read the dirty set: %w", err)
	}
	diff, err := git("diff", "--name-only", "-z", "--no-renames", strings.TrimSpace(base), pinned)
	if err != nil {
		return "", fmt.Errorf("read the target set: %w", err)
	}
	targets := nonEmptyNULFields(diff)
	ignoredOut, err := git("ls-files", "--others", "--ignored", "--exclude-standard", "-z")
	if err != nil {
		return "", fmt.Errorf("read the ignored files: %w", err)
	}
	ignored := make(map[string]bool)
	for _, path := range nonEmptyNULFields(ignoredOut) {
		ignored[path] = true
	}
	var ignoredTargets []string
	for _, target := range targets {
		if ignored[target] {
			ignoredTargets = append(ignoredTargets, target)
		}
	}
	var overlapping []string
	seen := make(map[string]bool)
	for _, rec := range status {
		for _, dirty := range rec.paths {
			if seen[dirty] {
				continue
			}
			for _, target := range targets {
				if pathsOverlap(dirty, target) {
					seen[dirty] = true
					overlapping = append(overlapping, dirty)
					break
				}
			}
		}
	}
	var reasons []string
	if len(overlapping) > 0 {
		reasons = append(reasons, "the integration worktree holds uncommitted or untracked changes at paths this card changes: "+strings.Join(overlapping, ", "))
	}
	if len(ignoredTargets) > 0 {
		reasons = append(reasons, "the integration worktree holds ignored files at paths this card changes: "+strings.Join(ignoredTargets, ", "))
	}
	return strings.Join(reasons, "; "), nil
}

// pathsOverlap reports whether two repository paths are equal, or one names a
// directory that contains the other (plan §B5, Overlap).
func pathsOverlap(a, b string) bool {
	return a == b || strings.HasPrefix(b, a+"/") || strings.HasPrefix(a, b+"/")
}

// nonEmptyNULFields splits a NUL-terminated git listing into its paths.
func nonEmptyNULFields(out string) []string {
	var paths []string
	for _, path := range strings.Split(out, "\x00") {
		if path != "" {
			paths = append(paths, path)
		}
	}
	return paths
}

// gitFail reports an error only when the git call FAILS — the ancestry
// probe answers through its exit code, and a zero exit is the pass.
func gitFail(git func(args ...string) (string, error), args ...string) error {
	_, err := git(args...)
	return err
}
