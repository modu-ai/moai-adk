package cli

// session_worktree_automerge.go — SPEC-WORKTREE-KEY-WIRING-001 M1/M2.
//
// sessionExitAutoMerge is the session-exit auto-merge trigger (M2): when
// workflow.worktree.auto_merge is true, a clean exit of a subcommand that
// materialized a session worktree merges that worktree's branch into the
// project's configured git-flow develop branch with a LOCAL `git merge
// --no-ff` — never a push, fetch, or any other remote-mutating command
// (REQ-WKW-005). Pushing remains the lead/operator's explicit act.
//
// The integration-window ceremony is BINDING (REQ-WKW-004): the path records
// the release-integration window through the same kanban lock API the
// `moai integration acquire` verb uses, BEFORE the merge, and releases it
// AFTER the merge terminates (success or failure), via defer. The force flag
// is pinned to literal false at the seam signatures below — the auto path
// never displaces any hold, live or stale; a held window skips the merge with
// a notice. The integration target is the SAME key the acquire verb resolves
// (`git_strategy.<mode>.develop_branch` via config.LoadGitFlowIntegrationConfig)
// so the window record and the serialized merge can never name different
// branches (design.md §2).
//
// Every failure path is a non-blocking stderr notice carrying
// AutoMergeNoticePrefix — distinct from both SessionExitCleanupNoticePrefix
// and PRMergeCleanupNoticePrefix (REQ-WKW-009 / EC-13) — and this function
// never aborts the caller (REQ-SW-004 fail-open spirit).

import (
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/execerr"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/session"
)

// AutoMergeNoticePrefix is the literal prefix of every auto-merge notice. It
// MUST NOT collide with SessionExitCleanupNoticePrefix
// ("removed by session-exit cleanup:") or PRMergeCleanupNoticePrefix
// ("removed by PR-merge cleanup:") so the three notice families are
// distinguishable in combined output (REQ-WKW-009 / AC-WKW-013 / EC-13).
const AutoMergeNoticePrefix = "auto-merge:"

// autoMergeNoticef writes one non-blocking auto-merge notice line to out. All
// auto-merge notices route through here so the AC-WKW-013 prefix invariant
// holds on every failure path by construction.
func autoMergeNoticef(out io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(out, "moai: "+AutoMergeNoticePrefix+" "+format+"\n", args...)
}

// Function-variable seams for M1 test injection, following the
// swapSessionWorktreeSeams / swapPRMergeSeams pattern: every git invocation on
// the auto-merge path is observable to tests, which is also the mechanism of
// the AC-WKW-006 zero-push assertion (the seam log enumerates every executed
// git subcommand; none may be push/fetch/remote-mutating). Tests swap these
// via swapAutoMergeSeams and restore on cleanup.
var (
	// autoMergeLockRoot resolves the directory whose .moai/state holds the
	// release-integration window record — the primary checkout, shared by
	// every linked worktree.
	autoMergeLockRoot = integrationLockRoot

	// autoMergeSessionID resolves this session's id (the
	// integrationSessionID("") semantics). An empty result skips the merge:
	// a hold with an invented holder could never be released (design.md §5).
	autoMergeSessionID = func() string { return integrationSessionID("") }

	// autoMergeLoadGitFlow reads the git-flow integration config — the same
	// loader `moai integration acquire` uses for its window record (design.md
	// §2: one fact, one key).
	autoMergeLoadGitFlow = config.LoadGitFlowIntegrationConfig

	// autoMergeBranchOf reports the branch checked out in the session worktree.
	autoMergeBranchOf = gitBranchOfWorktreeReal

	// autoMergeAheadCount counts commits reachable from the session branch but
	// not from the develop branch (the REQ-WKW-010 no-op gate).
	autoMergeAheadCount = gitRevListCountReal

	// autoMergeWorktreeForBranch resolves the tree holding the integration
	// branch; empty → REQ-WKW-008 skip.
	autoMergeWorktreeForBranch = worktreeForBranch

	// autoMergeResolveOwnerPID resolves the OWNING SESSION's pid for the lock
	// record (the same resolution the acquire verb uses).
	autoMergeResolveOwnerPID = session.ResolveOwnerPID

	// autoMergeReadLock reads the window record (the busy-window pre-check).
	autoMergeReadLock = kanban.ReadIntegrationLock

	// autoMergeAcquireLock records the window hold. The force argument is
	// deliberately NOT a parameter of this seam — it is pinned to literal
	// false inside (REQ-WKW-004: the auto path never takes over any hold), so
	// no caller on this path can displace a window even by accident.
	autoMergeAcquireLock = func(root string, lock kanban.IntegrationLock) (*kanban.IntegrationLock, error) {
		return kanban.AcquireIntegrationLock(root, lock, false)
	}

	// autoMergeReleaseLock releases THIS session's own hold. force is pinned
	// to literal false for the same reason: a non-forced release can only
	// remove a record the auto path's own session holds.
	autoMergeReleaseLock = func(root, sessionID string) (*kanban.IntegrationLock, error) {
		return kanban.ReleaseIntegrationLock(root, sessionID, false)
	}

	// autoMergeGitMerge runs the one permitted branch-mutating invocation:
	// `git merge --no-ff <session-branch>` inside the integration worktree
	// (REQ-WKW-005).
	autoMergeGitMerge = gitMergeNoFFReal

	// autoMergeMergeInProgress reports whether a merge is in progress in the
	// integration worktree (MERGE_HEAD exists) — the conflict detector.
	autoMergeMergeInProgress = gitMergeInProgressReal

	// autoMergeGitMergeAbort runs `git merge --abort` to restore the
	// integration worktree to its pre-merge state (REQ-WKW-006).
	autoMergeGitMergeAbort = gitMergeAbortReal

	// autoMergeHeadShort reports the integration worktree's short HEAD sha
	// after a successful merge — the merge commit named in the success notice
	// so the lead's batch-push flow can read what landed.
	autoMergeHeadShort = gitHeadShortReal
)

// sessionExitAutoMerge is the M2 session-exit auto-merge trigger, invoked at
// the init/web/profile exit paths BEFORE cleanupSessionWorktree. Ordering is
// merge-first: the merge consumes only committed state, but disposal deletes
// the tree, so merge-then-dispose is the only safe order (design.md §4).
//
// Gate order (each gate is silent unless a notice is stated):
//  1. no live session worktree → no-op (REQ-WKW-002's trigger clause)
//  2. auto_merge OFF → byte-identical baseline, no notice (REQ-WKW-001)
//  3. non-clean exit → no merge, no notice (REQ-WKW-002 clean-exit-only)
//  4. unconfigured / non-git-flow target → notice naming the config key
//     (REQ-WKW-003)
//  5. ahead-of == 0 → silent no-op, no window churn (REQ-WKW-010)
//  6. window held (live or stale) → skip + notice (REQ-WKW-004)
//  7. source/target dirty or target absent → skip + notice (REQ-WKW-007/008)
//  8. local `git merge --no-ff` inside the window; conflicts abort (REQ-WKW-005/006)
//
// The function NEVER returns an error and NEVER aborts the exit flow.
//
// @MX:ANCHOR: [AUTO] session-exit auto-merge entry (init/web/profile exit paths)
// @MX:REASON: REQ-WKW-001..005 — the OFF short-circuit and the clean-exit gate MUST precede every side effect (REQ-WKW-001 byte-identical baseline), and no path on this surface may execute a remote-mutating git command (REQ-WKW-005)
func sessionExitAutoMerge(cfg *config.Config, wtPath string, cleanExit bool, out io.Writer) {
	if wtPath == "" {
		// No session worktree was materialized (feature OFF / fail-back /
		// already-in-worktree). Nothing to merge.
		return
	}
	if cfg == nil || !cfg.Workflow.Worktree.AutoMerge {
		// REQ-WKW-001: the distributed default is OFF — byte-identical
		// baseline: no notice, no git invocation, no window record write.
		return
	}
	if !cleanExit {
		// REQ-WKW-002 clean-exit-only: a non-zero exit produces no merge. The
		// worktree is preserved for post-mortem by the cleanup path's own
		// notice; the merge path stays silent here.
		return
	}

	root := autoMergeLockRoot()
	gitFlow := autoMergeLoadGitFlow(root)
	if !gitFlow.IsGitFlow() || gitFlow.DevelopBranch == "" {
		// REQ-WKW-003: the integration target is inert when the project is
		// not manual git-flow, develop_branch is empty, or the file is
		// unreadable (the loader yields the zero value on every failure).
		autoMergeNoticef(out, "skipped (no integration target): the project is not manual git-flow or git_strategy develop_branch is unset; set git_strategy.<mode>.develop_branch to enable session-exit auto-merge")
		return
	}
	develop := gitFlow.DevelopBranch

	branch, err := autoMergeBranchOf(wtPath)
	if err != nil || strings.TrimSpace(branch) == "" {
		autoMergeNoticef(out, "skipped (cannot resolve the session branch of %s): %v", wtPath, err)
		return
	}

	ahead, err := autoMergeAheadCount(wtPath, develop, branch)
	if err != nil {
		autoMergeNoticef(out, "skipped (ahead-of check failed for %s..%s): %v", develop, branch, err)
		return
	}
	if ahead == 0 {
		// REQ-WKW-010: the branch carries no commit absent from develop —
		// skip the ENTIRE ceremony silently: no window record, no merge
		// invocation, no notice (byte-identical to the REQ-WKW-001 baseline).
		return
	}

	sessionID := autoMergeSessionID()
	if sessionID == "" {
		// A hold with an invented holder could never be released by its
		// holder; the auto path refuses the merge instead of writing one
		// (design.md §5).
		autoMergeNoticef(out, "skipped (cannot resolve this session's id): a window hold with an invented holder could never be released; merge manually")
		return
	}

	existing, err := autoMergeReadLock(root)
	if err != nil {
		autoMergeNoticef(out, "skipped (window record unreadable): %v", err)
		return
	}
	if existing.Held() && existing.SessionID != sessionID {
		// REQ-WKW-004: never displace any existing hold — live OR stale. A
		// stale hold is reclaimable by a human (`moai integration status`);
		// an automated reclaim would turn a dead session's unreleased hold
		// into silent machine displacement. The notice says what happened.
		holder := existing.SessionName
		if holder == "" {
			holder = existing.SessionID
		}
		autoMergeNoticef(out, "skipped (release-integration window held by %s since %s): a held window is never taken over automatically, live or stale; merge manually after the holder releases", holder, existing.AcquiredAt)
		return
	}

	targetWt := autoMergeWorktreeForBranch(develop)
	ownerPID, _ := autoMergeResolveOwnerPID()
	if _, aerr := autoMergeAcquireLock(root, kanban.IntegrationLock{
		SessionID: sessionID,
		PID:       ownerPID,
		// The pid recorded is the OWNING SESSION's, never this process's —
		// same discipline as the acquire verb (integration.go).
		PIDSource: kanban.PIDSourceSessionOwner,
		// The record names the integration TARGET, source config: the same
		// branch the merge below lands on.
		Branch:       develop,
		BranchSource: kanban.BranchSourceConfig,
		Worktree:     targetWt,
	}); aerr != nil {
		autoMergeNoticef(out, "skipped (window acquire refused): %v", aerr)
		return
	}
	// REQ-WKW-004: release AFTER the merge terminates, on every path below —
	// guards, conflicts, and success alike.
	defer func() {
		if _, rerr := autoMergeReleaseLock(root, sessionID); rerr != nil {
			autoMergeNoticef(out, "window release failed (%v): the hold is recorded under session id %s; release with 'moai integration release'", rerr, sessionID)
		}
	}()

	// REQ-WKW-008 case 1: no worktree holds the integration branch.
	if targetWt == "" {
		autoMergeNoticef(out, "skipped (no worktree holds %s): provision the integration worktree first; nothing was merged", develop)
		return
	}

	// REQ-WKW-007 source dirty guard: uncommitted work is never merged and
	// never destroyed. Disposal then proceeds under auto_cleanup's own dirty
	// guard. The dirty checks reuse the M4 shared helper so one status seam
	// backs both call sites.
	if dirty, derr := worktreeIsDirty(wtPath); derr != nil {
		autoMergeNoticef(out, "skipped (source dirty-check failed: %v): nothing was merged", derr)
		return
	} else if dirty {
		autoMergeNoticef(out, "skipped (uncommitted changes in %s): uncommitted work is never merged; disposal proceeds under auto_cleanup's own guard", wtPath)
		return
	}

	// REQ-WKW-008 case 2: the integration worktree is dirty.
	if dirty, derr := worktreeIsDirty(targetWt); derr != nil {
		autoMergeNoticef(out, "skipped (target dirty-check failed: %v): nothing was merged", derr)
		return
	} else if dirty {
		autoMergeNoticef(out, "skipped (integration worktree %s has uncommitted changes): nothing was merged", targetWt)
		return
	}

	// REQ-WKW-005: the only branch-mutating invocation on this path — a
	// local merge inside the integration worktree.
	if _, merr := autoMergeGitMerge(targetWt, branch); merr != nil {
		// REQ-WKW-006 conflict safety: while a merge is in progress,
		// `git merge --abort` restores the integration worktree to its
		// pre-merge state. The session worktree is never touched.
		if autoMergeMergeInProgress(targetWt) {
			if aerr := autoMergeGitMergeAbort(targetWt); aerr != nil {
				autoMergeNoticef(out, "merge failed (%v) AND 'git merge --abort' failed (%v): the integration worktree %s may still hold merge state; resolve manually", merr, aerr, targetWt)
				return
			}
			autoMergeNoticef(out, "merge failed with conflicts (%v): 'git merge --abort' restored %s to its pre-merge state; the session worktree is untouched", merr, targetWt)
			return
		}
		autoMergeNoticef(out, "merge failed (%v): the integration worktree %s is as git left it; nothing was pushed", merr, targetWt)
		return
	}

	// Success: name the resulting merge commit so the lead's batch-push flow
	// can read what landed (spec.md §C — local/remote divergence is accepted;
	// pushing remains the lead's explicit act).
	autoMergeNoticef(out, "merged %s into %s as %s (local merge only — push remains the lead's explicit act)", branch, develop, autoMergeHeadShort(targetWt))
}

// --- real implementations (overridable in tests via the seams above) ---

// gitBranchOfWorktreeReal reports the branch checked out in the worktree at
// wtPath via `git -C <wtPath> rev-parse --abbrev-ref HEAD`.
func gitBranchOfWorktreeReal(wtPath string) (string, error) {
	out, err := exec.Command("git", "-C", wtPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// gitRevListCountReal counts commits reachable from branch but not from
// develop (`git -C <wtPath> rev-list --count <develop>..<branch>`), the
// REQ-WKW-010 ahead-of gate. A non-numeric answer is an error, never a
// silent zero — "0 ahead" and "git could not answer" are different facts.
func gitRevListCountReal(wtPath, develop, branch string) (int, error) {
	out, err := exec.Command("git", "-C", wtPath, "rev-list", "--count", develop+".."+branch).Output()
	if err != nil {
		return 0, err
	}
	n, perr := strconv.Atoi(strings.TrimSpace(string(out)))
	if perr != nil {
		return 0, fmt.Errorf("unexpected rev-list output %q", strings.TrimSpace(string(out)))
	}
	return n, nil
}

// gitMergeNoFFReal runs `git -C <dir> merge --no-ff <branch>` — the one
// permitted branch-mutating invocation on the auto-merge path (REQ-WKW-005).
// The -C flag binds the merge to the integration worktree, the same semantics
// as cmd.Dir at that tree.
func gitMergeNoFFReal(dir, branch string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "merge", "--no-ff", branch)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// gitMergeInProgressReal reports whether a merge is in progress in dir
// (MERGE_HEAD resolves). `rev-parse -q --verify` exits non-zero both when the
// ref is absent and on a git error, and both read as "not in progress" — the
// merge has already failed at that point, so the abort is best-effort either way.
func gitMergeInProgressReal(dir string) bool {
	return exec.Command("git", "-C", dir, "rev-parse", "-q", "--verify", "MERGE_HEAD").Run() == nil
}

// gitMergeAbortReal runs `git -C <dir> merge --abort` (REQ-WKW-006).
func gitMergeAbortReal(dir string) error {
	cmd := exec.Command("git", "-C", dir, "merge", "--abort")
	if out, err := cmd.CombinedOutput(); err != nil {
		// execerr.StatusDetail, not %w: a raw *exec.ExitError chain would be
		// mistaken for an intentional ExitCoder at the cmd/moai seam (t130).
		return fmt.Errorf("%s (%s)", strings.TrimSpace(string(out)), execerr.StatusDetail(err))
	}
	return nil
}

// gitHeadShortReal reports dir's short HEAD sha, or the literal "unknown" on
// failure — the notice still fires, just without a citable commit.
func gitHeadShortReal(dir string) string {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
