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

// localMainResyncFetch is the fetch seam (plan §B3a step 4). Its production
// default runs `git fetch origin main` in the primary checkout and returns its
// error. The tests replace it with a no-op, so no test contacts a remote.
var localMainResyncFetch = func(repoRoot string) error {
	_, err := factorylane.ExecGitRunner{Dir: repoRoot}.Git("fetch", "origin", localMainResyncBranch)
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
	heldBefore := localMainResyncHeldBy(repoRoot, sessionID)
	ownerPID, _ := session.ResolveOwnerPID()
	replaced, err := factory.AcquireIntegrationWindow(repoRoot, factory.IntegrationLock{
		SessionID:    sessionID,
		PID:          ownerPID,
		PIDSource:    factory.PIDSourceSessionOwner,
		Branch:       localMainResyncBranch,
		BranchSource: factory.BranchSourceConfig,
		Worktree:     repoRoot,
	}, false, &factory.AcquireWindowOptions{LeaseDuration: integrationLeaseDuration(repoRoot)})
	if err != nil {
		return "", err
	}
	// Release only a window this call took. The window is kept when the session
	// held it before the call and the acquire displaced nobody.
	keep := heldBefore && replaced == nil

	report, err := localMainResyncInWindow(repoRoot)
	if err != nil {
		if keep {
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
func localMainResyncInWindow(repoRoot string) (string, error) {
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

	// Step 4: fetch, observe the exit status, then take BASELINE_SHA from the
	// fetched origin ref. Every later origin-facing comparison uses it.
	if err := localMainResyncFetch(repoRoot); err != nil {
		return "", localMainResyncRefusal(factory.MergeExitOther, "integration resync: git fetch origin %s failed, so BASELINE_SHA was not taken: %v", localMainResyncBranch, err)
	}
	baselineOut, err := git("rev-parse", "--verify", "--quiet", "refs/remotes/origin/"+localMainResyncBranch+"^{commit}")
	if err != nil {
		return "", localMainResyncRefusal(factory.MergeExitOther, "integration resync: origin/%s does not resolve after the fetch: %v", localMainResyncBranch, err)
	}
	baseline := strings.TrimSpace(baselineOut)

	headOut, err := git("rev-parse", "HEAD")
	if err != nil {
		return "", localMainResyncRefusal(factory.MergeExitOther, "integration resync: read HEAD: %v", err)
	}
	head := strings.TrimSpace(headOut)

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
	// The --no-overwrite-ignore flag is the second guard (plan §B4). Autostash is
	// disabled so the verb never stashes another session's changes (plan §B5).
	if _, err := git("-c", "merge.autoStash=false", "merge", "--ff-only", "--no-overwrite-ignore", "-q", baseline); err != nil {
		return "", localMainResyncFailed(git, head, err)
	}
	// F3 test seam (card t1616): see localMainResyncAfterFastForward.
	localMainResyncAfterFastForward(repoRoot)

	// Step 6: HEAD equals BASELINE_SHA, and the symbolic HEAD still names the
	// branch. SHA equality alone does not prove that the branch did not change.
	afterOut, err := git("rev-parse", "HEAD")
	if err != nil || strings.TrimSpace(afterOut) != baseline {
		return "", localMainResyncRefusal(factory.MergeExitPostMerge, "integration resync: after the fast-forward HEAD is not origin/%s %s. Inspect the primary from your own terminal before any further merge.", localMainResyncBranch, localMainResyncShort(baseline))
	}
	if symAfter, err := git("symbolic-ref", "-q", "HEAD"); err != nil || strings.TrimSpace(symAfter) != ref {
		return "", localMainResyncRefusal(factory.MergeExitPostMerge, "integration resync: after the fast-forward HEAD no longer names %s. Inspect the primary from your own terminal before any further merge.", localMainResyncBranch)
	}

	// Step 7: report the old and new SHAs.
	return fmt.Sprintf("local main fast-forwarded %s -> %s to origin/%s", localMainResyncShort(head), localMainResyncShort(baseline), localMainResyncBranch), nil
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
// is class 8.
func localMainResyncFailed(git func(args ...string) (string, error), head string, cause error) error {
	if now, err := git("rev-parse", "HEAD"); err == nil && strings.TrimSpace(now) != head {
		return localMainResyncRefusal(factory.MergeExitPostMerge, "integration resync: the fast-forward failed (%v) and HEAD moved from %s. Inspect the primary from your own terminal before any further merge.", cause, localMainResyncShort(head))
	}
	if status, err := git("status", "--porcelain=v1", "-z", "--untracked-files=all"); err != nil || status != "" {
		return localMainResyncRefusal(factory.MergeExitMergeDirty, "integration resync: the fast-forward failed (%v), and the primary is not clean or could not be read. Hold first.", cause)
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
// release-on-refusal pattern of the merge step.
func localMainResyncRelease(repoRoot, sessionID string, cause error) error {
	if _, err := factory.ReleaseIntegrationLock(repoRoot, sessionID, 0, false); err != nil {
		return fmt.Errorf("%w (releasing the window also failed: %v — moai integration status reads it)", cause, err)
	}
	return cause
}

// localMainResyncHeldBy reports whether the recorded window already belongs to
// sessionID. An unreadable record reads as not held. The acquire that follows
// refuses an unreadable record, so no release is reached on that path.
func localMainResyncHeldBy(repoRoot, sessionID string) bool {
	prior, err := factory.ReadIntegrationLock(repoRoot)
	return err == nil && prior.Held() && prior.SessionID == sessionID
}

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
