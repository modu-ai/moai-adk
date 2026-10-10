// local_main_resync.go — `moai integration resync` (SPEC-LOCAL-MAIN-FLOW-001,
// card t1616, REQ-LMF-014, plan §B3a): the fast-forward re-sync of the primary
// checkout's local main from origin/main, run inside the integration window.
//
// The re-sync is tool-owned and fast-forward-only, and it runs on the primary
// surface alone (plan §B2 case 4): the gate workflow.local_main_integration.enabled
// is on, and the primary checkout's HEAD names the integration branch. After the
// window is taken, the sequence is the one §B3a fixes: confirm HEAD and a fully
// clean primary, fetch origin/main and take BASELINE_SHA from it, classify HEAD
// against BASELINE_SHA with read-only ancestry tests, then refuse, report a
// no-op, or fast-forward. The window is released on every path that took it.
package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/spf13/cobra"
)

// localMainResyncBranch is the branch the re-sync follows: local main tracks
// origin/main (plan §B3a trigger). The configured integration branch must carry
// this name, because the fetch and BASELINE_SHA are origin/main.
const localMainResyncBranch = "main"

// localMainResyncBaselineRef is the remote-tracking ref BASELINE_SHA is read from
// (plan §B3a step 4). The fetch names it as its destination, so the ref moves to
// the fetched commit whatever the remote's configured fetch refspecs map.
const localMainResyncBaselineRef = "refs/remotes/origin/" + localMainResyncBranch

// localMainResyncFetch is the fetch seam (plan §B3a step 4). Its production
// default runs `git fetch origin main:refs/remotes/origin/main` in the primary
// checkout and returns its error; the explicit destination is the ref the baseline
// is read from (leader ruling (A), 2026-10-10). The tests replace it with a no-op,
// except TestLocalMainResyncFetchPinsFetchedCommit, which runs this default against
// a local bare repository, so no test contacts a remote.
var localMainResyncFetch = func(repoRoot string) error {
	_, err := factorylane.ExecGitRunner{Dir: repoRoot}.Git("fetch", "origin", localMainResyncBranch+":"+localMainResyncBaselineRef)
	return err
}

// localMainResyncAfterFastForward is the post-fast-forward test seam (card t1616,
// F3). It runs once the fast-forward has returned and before step 6 reads HEAD
// back, so a test can produce the post-merge anomaly that step 6 classifies. The
// production default does nothing.
var localMainResyncAfterFastForward = func(repoRoot string) {}

// newIntegrationResyncCmd registers `moai integration resync`. It takes no
// flags: the session is the one the window verbs resolve from the environment.
func newIntegrationResyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resync",
		Short: "Fast-forward the primary checkout's local main to origin/main, inside the integration window",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := integrationLockRoot()
			// The settings-drift precondition runs before the window is taken, as
			// it does for acquire. A refusal here takes no window.
			if _, err := acquireSettingsDriftPrecondition(cmd, root, "", false); err != nil {
				return err
			}
			report, err := runLocalMainResync(root)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "integration resync: "+report)
			return nil
		},
	}
}

// runLocalMainResync fast-forwards local main to origin/main inside the
// integration window and returns the report line. A refusal before the window
// is taken changes no record. Every refusal or success after the acquisition
// releases the window it took (plan §B3a step 8); a window the caller already
// held stays held.
func runLocalMainResync(repoRoot string) (string, error) {
	sessionID := integrationSessionID("")
	if sessionID == "" {
		return "", fmt.Errorf("integration resync: cannot resolve this session's id; the window needs an address the holder decision can recognize")
	}
	if !localMainIntegrationEnabled(repoRoot) {
		return "", fmt.Errorf("integration resync: refused — workflow.local_main_integration.enabled is off, so the primary checkout is not an integration surface and its local main is left as it is")
	}
	branch := config.LoadGitFlowIntegrationConfig(repoRoot).IntegrationTarget
	if branch != localMainResyncBranch {
		return "", fmt.Errorf("integration resync: refused — the integration branch is %q, but the re-sync fast-forwards local main from origin/%s and runs only when the integration branch is %s", branch, localMainResyncBranch, localMainResyncBranch)
	}

	// Step 1: the existing acquisition, with the record the acquire verb writes.
	// heldBefore is read first. A window this session already holds is re-acquired
	// (the acquire refreshes the holder's record and lease), and the release at the
	// end skips it. The read is outside the acquire's mutation, so a takeover in
	// between shows up as replaced.
	initWindowLeaseOverride(repoRoot)
	prior := localMainResyncPriorHold(repoRoot, sessionID)
	heldBefore := prior != nil
	ownerPID, _ := session.ResolveOwnerPID()
	want := factory.IntegrationLock{
		SessionID:    sessionID,
		PID:          ownerPID,
		PIDSource:    factory.PIDSourceSessionOwner,
		Branch:       localMainResyncBranch,
		BranchSource: factory.BranchSourceConfig,
		Worktree:     repoRoot,
	}
	if prior != nil {
		localMainResyncCarryHolder(&want, prior)
	}
	replaced, err := factory.AcquireIntegrationWindow(repoRoot, want, false, &factory.AcquireWindowOptions{LeaseDuration: integrationLeaseDuration(repoRoot)})
	if err != nil {
		return "", err
	}
	// Release only a window this call took. The window is kept when the session
	// held it before the call and the acquire displaced nobody.
	keep := heldBefore && replaced == nil

	report, err := localMainResyncInWindow(repoRoot, sessionID)
	if err != nil {
		// A retained refusal (F3) keeps the window held: its policy hold could not be
		// written, and a release would promote the next queued lane onto the anomaly.
		var retained *localMainResyncRetained
		if keep || errors.As(err, &retained) {
			return "", err
		}
		return "", localMainResyncRelease(repoRoot, sessionID, err)
	}
	if !keep {
		if _, err := factory.ReleaseIntegrationLock(repoRoot, sessionID, 0, false); err != nil {
			return "", fmt.Errorf("integration resync: %s, but releasing the window failed: %v (moai integration status reads it)", report, err)
		}
	}
	if replaced != nil {
		report += fmt.Sprintf(" (displaced %s, pid %d, held since %s)", replaced.SessionID, replaced.PID, replaced.AcquiredAt)
	}
	return report, nil
}

// localMainResyncInWindow runs plan §B3a steps 2 through 7 with the window held.
// Each refusal carries the exit class the merge step uses for the same cause.
// Between the fetch and the move, the re-validation (card t1616, F2 and F5)
// reads the holder, HEAD, and the clean state again.
func localMainResyncInWindow(repoRoot, sessionID string) (string, error) {
	git := factorylane.ExecGitRunner{Dir: repoRoot}.Git
	ref := "refs/heads/" + localMainResyncBranch

	// Step 2: HEAD names the integration branch. A detached HEAD fails the same
	// check, because symbolic-ref has no answer for it.
	if headRef, err := git("symbolic-ref", "-q", "HEAD"); err != nil || strings.TrimSpace(headRef) != ref {
		return "", localMainResyncNoHolder(repoRoot)
	}

	// Step 3: a fully clean primary (plan §B4). Status does not list ignored
	// files, so they do not trip this check; the ignored-file check is step 5a.
	status, err := git("status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return "", localMainResyncRefusal(factory.MergeExitOther, "integration resync: status the primary checkout: %v", err)
	}
	if status != "" {
		return "", localMainResyncRefusal(factory.MergeExitWorktreeDirty, "integration resync: refused — the primary checkout has %d uncommitted or untracked change(s). The local-main re-sync needs a fully clean primary, because changes owned by other sessions cannot be told apart from yours. Move your changes into a card worktree and commit them there, or ask the session that owns them to commit or discard them. Do not stash: the stash is repository-wide. Then re-run.", strings.Count(status, "\x00"))
	}

	// HEAD is read before the fetch, so the re-validation before the move can see
	// whether HEAD moved while the fetch ran (card t1616, F2).
	headOut, err := git("rev-parse", "HEAD")
	if err != nil {
		return "", localMainResyncRefusal(factory.MergeExitOther, "integration resync: read HEAD: %v", err)
	}
	head := strings.TrimSpace(headOut)

	// Step 4: fetch origin/main under an explicit destination, observe the exit
	// status, then take BASELINE_SHA from refs/remotes/origin/main. The destination
	// is named on the command line, so the fetch writes that ref whatever the
	// remote's configured fetch refspecs map. Only a fetch of origin/main moves that
	// ref, so a fetch of another ref cannot move the baseline (leader ruling (A),
	// 2026-10-10). The refspec has no leading `+`: a non-fast-forward move of
	// origin/main fails the fetch, and the re-sync refuses rather than forcing the
	// ref. Every later origin-facing comparison uses BASELINE_SHA.
	if err := localMainResyncFetch(repoRoot); err != nil {
		return "", localMainResyncRefusal(factory.MergeExitOther, "integration resync: git fetch origin %s:%s failed, so BASELINE_SHA was not taken: %v", localMainResyncBranch, localMainResyncBaselineRef, err)
	}
	baselineOut, err := git("rev-parse", "--verify", "--quiet", localMainResyncBaselineRef+"^{commit}")
	if err != nil {
		return "", localMainResyncRefusal(factory.MergeExitOther, "integration resync: %s does not name a commit after the fetch of origin/%s, so BASELINE_SHA was not taken: %v", localMainResyncBaselineRef, localMainResyncBranch, err)
	}
	baseline := strings.TrimSpace(baselineOut)

	// Step 5, case (b): BASELINE_SHA is an ancestor of HEAD, equality included.
	containsBaseline, err := localMainResyncAncestor(git, baseline, head)
	if err != nil {
		return "", localMainResyncRefusal(factory.MergeExitOther, "integration resync: ancestry of origin/%s: %v", localMainResyncBranch, err)
	}
	if containsBaseline {
		return fmt.Sprintf("local main %s already contains origin/%s %s: no fast-forward needed", localMainResyncShort(head), localMainResyncBranch, localMainResyncShort(baseline)), nil
	}

	// Step 5, case (c): neither commit is an ancestor of the other. The tool
	// does not reconcile the two histories, and HEAD does not move.
	fastForwards, err := localMainResyncAncestor(git, head, baseline)
	if err != nil {
		return "", localMainResyncRefusal(factory.MergeExitOther, "integration resync: ancestry of origin/%s: %v", localMainResyncBranch, err)
	}
	if !fastForwards {
		return "", localMainResyncRefusal(factory.MergeExitNotDescendant, "integration resync: refused — local main (%s) has diverged from origin/%s (%s): local main holds commits that origin/%s lacks, and the tool does not reconcile them. HEAD does not move.", localMainResyncShort(head), localMainResyncBranch, localMainResyncShort(baseline), localMainResyncBranch)
	}

	// Step 5, case (a): HEAD is an ancestor of BASELINE_SHA. Ignored bytes at a
	// path the fast-forward writes refuse before any merge runs (step 5a).
	changedOut, err := git("diff", "--name-only", "--no-renames", "-z", head, baseline)
	if err != nil {
		return "", localMainResyncRefusal(factory.MergeExitOther, "integration resync: list the paths the fast-forward changes: %v", err)
	}
	ignoredOut, err := git("ls-files", "--others", "--ignored", "--exclude-standard", "-z")
	if err != nil {
		return "", localMainResyncRefusal(factory.MergeExitOther, "integration resync: list the ignored files: %v", err)
	}
	if overlap := localMainResyncOverlap(changedOut, ignoredOut); len(overlap) > 0 {
		return "", localMainResyncRefusal(factory.MergeExitCollision, "integration resync: refused — the primary checkout holds ignored bytes at paths this fast-forward would write: %s. The fast-forward would overwrite them. Move or rename them from your own terminal, then re-run. The tool does not remove or stash files.", strings.Join(overlap, ", "))
	}
	// The re-validation (card t1616, F2 and F5), immediately before the move. The
	// holder check and the clean check ran earlier, and the fetch ran in between.
	if err := localMainResyncRevalidate(git, repoRoot, sessionID, ref, head); err != nil {
		return "", err
	}
	// The --no-overwrite-ignore flag is the second guard (plan §B4). Autostash is
	// disabled so the verb never stashes another session's changes (plan §B5).
	if _, err := git("-c", "merge.autoStash=false", "merge", "--ff-only", "--no-overwrite-ignore", "-q", baseline); err != nil {
		return "", localMainResyncFailed(git, repoRoot, head, baseline, err)
	}
	// F3 test seam (card t1616): see localMainResyncAfterFastForward.
	localMainResyncAfterFastForward(repoRoot)

	// Step 6: HEAD equals BASELINE_SHA, and the symbolic HEAD still names the
	// branch. SHA equality alone does not prove that the branch did not change.
	afterOut, err := git("rev-parse", "HEAD")
	if err != nil || strings.TrimSpace(afterOut) != baseline {
		return "", localMainResyncAnomaly(repoRoot, baseline, factory.MergeExitPostMerge, "integration resync: after the fast-forward HEAD is not origin/%s %s. Inspect the primary from your own terminal before any further merge.", localMainResyncBranch, localMainResyncShort(baseline))
	}
	if symAfter, err := git("symbolic-ref", "-q", "HEAD"); err != nil || strings.TrimSpace(symAfter) != ref {
		return "", localMainResyncAnomaly(repoRoot, baseline, factory.MergeExitPostMerge, "integration resync: after the fast-forward HEAD no longer names %s. Inspect the primary from your own terminal before any further merge.", localMainResyncBranch)
	}

	// Step 6, status (card t1616, F5): the clean check ran before the move, and the
	// fast-forward writes tracked paths only, so a file that appears in the primary
	// after the move goes unseen unless the status is read here. An unreadable status
	// is not proof of a clean primary, so it takes the same class, as
	// localMainResyncFailed does. The porcelain output is NUL-separated; the hold names
	// it with the NULs turned into "; ".
	postStatus, err := git("status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return "", localMainResyncAnomaly(repoRoot, baseline, factory.MergeExitMergeDirty, "integration resync: the primary could not be read after the fast-forward (%v). Inspect the primary from your own terminal before any further merge.", err)
	}
	if postStatus != "" {
		dirty := strings.ReplaceAll(strings.TrimSuffix(postStatus, "\x00"), "\x00", "; ")
		return "", localMainResyncAnomaly(repoRoot, baseline, factory.MergeExitMergeDirty, "integration resync: the primary is not clean after the fast-forward: %s. Inspect the primary from your own terminal before any further merge.", dirty)
	}

	// Step 7: report the old and new SHAs.
	return fmt.Sprintf("local main fast-forwarded %s -> %s to origin/%s", localMainResyncShort(head), localMainResyncShort(baseline), localMainResyncBranch), nil
}

// localMainResyncRevalidate re-checks, immediately before the fast-forward, what
// the move relies on (card t1616, F2 and F5). This session must still hold the
// window with an unexpired lease (i), the symbolic HEAD must still name the
// integration branch (ii), HEAD must be where it was before the fetch (iii), and
// the primary must still be fully clean (iv). A refusal carries the exit class the
// merge step uses for the same cause, and it moves nothing.
func localMainResyncRevalidate(git func(args ...string) (string, error), repoRoot, sessionID, ref, headBeforeFetch string) error {
	lock, err := factory.ReadIntegrationLock(repoRoot)
	if err != nil {
		return localMainResyncRefusal(factory.MergeExitOther, "integration resync: read the window record before the move: %v", err)
	}
	if !lock.Held() || lock.SessionID != sessionID {
		holder := "nobody"
		if lock.Held() {
			holder = lock.SessionID
		}
		return localMainResyncRefusal(factory.MergeExitNotHolder, "integration resync: refused — %s holds the window, not %s. The window was taken while the fetch ran. The primary is unchanged; re-run the re-sync to take the window again (moai integration status reads it).", holder, sessionID)
	}
	if lock.LeaseExpired(factory.WindowClock()) {
		return localMainResyncRefusal(factory.MergeExitExpiredLease, "integration resync: refused — your lease expired at %s while the fetch ran. The primary is unchanged; re-run the re-sync.", lock.LeaseExpiresAt)
	}
	if headRef, err := git("symbolic-ref", "-q", "HEAD"); err != nil || strings.TrimSpace(headRef) != ref {
		return localMainResyncRefusal(factory.MergeExitOther, "integration resync: refused — HEAD no longer names %s after the fetch. The primary is unchanged. Inspect it from your own terminal before re-running.", localMainResyncBranch)
	}
	headNow, err := git("rev-parse", "HEAD")
	if err != nil {
		return localMainResyncRefusal(factory.MergeExitOther, "integration resync: read HEAD before the move: %v", err)
	}
	if now := strings.TrimSpace(headNow); now != headBeforeFetch {
		return localMainResyncRefusal(factory.MergeExitBaseMoved, "integration resync: refused — HEAD moved from %s to %s while the fetch ran. The primary is unchanged; re-run the re-sync against the new HEAD.", localMainResyncShort(headBeforeFetch), localMainResyncShort(now))
	}
	status, err := git("status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return localMainResyncRefusal(factory.MergeExitOther, "integration resync: status the primary checkout before the move: %v", err)
	}
	if status != "" {
		return localMainResyncRefusal(factory.MergeExitWorktreeDirty, "integration resync: refused — the primary checkout became dirty while the fetch ran (%d uncommitted or untracked change(s)). The primary is unchanged. Move or commit those changes from your own terminal, then re-run. Do not stash: the stash is repository-wide.", strings.Count(status, "\x00"))
	}
	return nil
}

// localMainResyncNoHolder refuses when the primary checkout does not hold the
// integration branch. When another tree holds it, the refusal names that tree
// (plan §B2 case 2). Otherwise it gives the plan's case-5 guidance.
func localMainResyncNoHolder(repoRoot string) error {
	if other := factoryWorktreeForBranchIn(repoRoot, localMainResyncBranch); other != "" && !factorySameTree(other, repoRoot) {
		return fmt.Errorf("integration resync: refused — %s is checked out in %s, not in the primary checkout. The re-sync applies to the primary checkout only, and that worktree is not touched", localMainResyncBranch, other)
	}
	return fmt.Errorf("integration resync: refused — no tree holds the integration branch %s. Provision its worktree (the leader does this at batch start), or check %s out in the primary checkout from your own terminal. The tool does not switch branches", localMainResyncBranch, localMainResyncBranch)
}

// localMainResyncFailed classifies a refused fast-forward. An ff-only merge
// applies whole or not at all, so a clean tree with HEAD where it was is the
// merge-failed class. A dirty or unreadable tree is class 7, and a moved HEAD
// is class 8. Classes 7 and 8 leave the primary in a state no later holder may
// be promoted onto, so they write the policy hold first (card t1616, F3).
func localMainResyncFailed(git func(args ...string) (string, error), repoRoot, head, baseline string, cause error) error {
	if now, err := git("rev-parse", "HEAD"); err == nil && strings.TrimSpace(now) != head {
		return localMainResyncAnomaly(repoRoot, baseline, factory.MergeExitPostMerge, "integration resync: the fast-forward failed (%v) and HEAD moved from %s. Inspect the primary from your own terminal before any further merge.", cause, localMainResyncShort(head))
	}
	if status, err := git("status", "--porcelain=v1", "-z", "--untracked-files=all"); err != nil || status != "" {
		return localMainResyncAnomaly(repoRoot, baseline, factory.MergeExitMergeDirty, "integration resync: the fast-forward failed (%v), and the primary is not clean or could not be read. The window policy is held for the leader.", cause)
	}
	return localMainResyncRefusal(factory.MergeExitMergeFailed, "integration resync: the fast-forward was refused (%v). The primary is clean and HEAD is unchanged.", cause)
}

// localMainResyncRefusal builds a refusal that carries the merge step's exit
// class for the same cause, so the CLI exit code names the cause.
func localMainResyncRefusal(code int, format string, args ...any) error {
	return &factory.MergeStepError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

// localMainResyncRelease releases the window the re-sync took and returns the
// cause. A failed release is appended, so the operator sees both. This is the
// release-on-refusal pattern of the merge step. A holder refusal (cause 14) names
// a window another session took, so its release is refused by design and nothing
// is appended; the merge step's NotHolder branch makes the same exception.
func localMainResyncRelease(repoRoot, sessionID string, cause error) error {
	if _, err := factory.ReleaseIntegrationLock(repoRoot, sessionID, 0, false); err != nil {
		if code, ok := factory.MergeExitCode(cause); ok && code == factory.MergeExitNotHolder && (factory.IsIntegrationLockForeign(err) || factory.IsIntegrationLockNotHeld(err)) {
			return cause
		}
		return fmt.Errorf("%w (releasing the window also failed: %v — moai integration status reads it)", cause, err)
	}
	return cause
}

// localMainResyncPriorHold returns the recorded window when it already belongs to
// sessionID, and nil otherwise. An unreadable record reads as not held. The acquire
// that follows refuses an unreadable record, so no release is reached on that path.
func localMainResyncPriorHold(repoRoot, sessionID string) *factory.IntegrationLock {
	prior, err := factory.ReadIntegrationLock(repoRoot)
	if err != nil || !prior.Held() || prior.SessionID != sessionID {
		return nil
	}
	return prior
}

// localMainResyncCarryHolder copies the holder metadata of a window this session
// already holds into the record a holder re-entry writes (card t1616, F4). The
// acquire refreshes the record and does not re-identify the holder, so the card,
// the session name, the first-acquired instant, the settings-drift record, and the
// displacement history carry over. PID and PIDSource are not carried: the acquire
// re-resolves them from this session, as the acquire verb does, which re-anchors a
// record written before the anchor existed.
func localMainResyncCarryHolder(want *factory.IntegrationLock, prior *factory.IntegrationLock) {
	want.SessionName = prior.SessionName
	want.Card = prior.Card
	want.AcquiredAt = prior.AcquiredAt
	want.SettingsDriftBypass = prior.SettingsDriftBypass
	want.SettingsDriftPreserved = prior.SettingsDriftPreserved
	want.Displaced = prior.Displaced
	want.DisplacedReason = prior.DisplacedReason
}

// localMainResyncAnomaly is the post-merge outcome of the re-sync (card t1616,
// F3). The fast-forward left the primary where no later holder may be promoted
// onto it, so the policy hold is written before the window is released. It mirrors
// the merge step's postMergeHold: the hold names the cause, the first 12 characters
// of BASELINE_SHA, and the fast-forward, and the refusal carries the same message.
// When the hold cannot be written, the refusal says so and is returned as a
// localMainResyncRetained, so the caller keeps the window held.
func localMainResyncAnomaly(repoRoot, baseline string, code int, format string, args ...any) error {
	reason := fmt.Sprintf(format, args...) + fmt.Sprintf(" (fast-forward target origin/%s %s left in place for the leader)", localMainResyncBranch, localMainResyncShort(baseline))
	holdErr := factory.WriteIntegrationWindowPolicy(repoRoot, factory.IntegrationWindowPolicy{
		Policy: factory.PolicyHold,
		Reason: reason,
		SetBy:  "integration-resync (local main)",
		SetAt:  factory.WindowClock().Format(time.RFC3339),
	})
	if holdErr != nil {
		return &localMainResyncRetained{err: localMainResyncRefusal(code, "%s; writing the hold also failed: %v; the window stays held, because a release would promote the next queued lane onto this state", reason, holdErr)}
	}
	return localMainResyncRefusal(code, "%s", reason)
}

// localMainResyncRetained marks a refusal whose policy hold could not be written.
// The window stays held: releasing it would promote the next queued lane onto the
// state the hold was meant to keep it off. Unwrap keeps the exit class visible.
type localMainResyncRetained struct{ err error }

func (r *localMainResyncRetained) Error() string { return r.err.Error() }

func (r *localMainResyncRetained) Unwrap() error { return r.err }

// localMainResyncAncestor reports whether a is an ancestor of b, equality
// included. `git merge-base --is-ancestor` exits 1 for "not an ancestor". Any
// other failure is an error, never a negative answer.
func localMainResyncAncestor(git func(args ...string) (string, error), a, b string) (bool, error) {
	_, err := git("merge-base", "--is-ancestor", a, b)
	if err == nil {
		return true, nil
	}
	var exit *factorylane.GitExitError
	if errors.As(err, &exit) && exit.ExitCode == 1 {
		return false, nil
	}
	return false, err
}

// localMainResyncOverlap returns the paths the fast-forward changes that the
// primary holds as ignored files (plan §B3a step 5a). Both inputs are
// NUL-separated, as `-z` prints them.
func localMainResyncOverlap(changed, ignored string) []string {
	ignoredSet := make(map[string]bool)
	for _, p := range strings.Split(ignored, "\x00") {
		ignoredSet[p] = true
	}
	var overlap []string
	for _, p := range strings.Split(changed, "\x00") {
		if p != "" && ignoredSet[p] {
			overlap = append(overlap, p)
		}
	}
	return overlap
}

// localMainResyncShort abbreviates a full SHA to the 12 characters the report uses.
func localMainResyncShort(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
