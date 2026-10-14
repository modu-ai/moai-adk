package cli

// local_main_resync_test.go — SPEC-LOCAL-MAIN-FLOW-001 (card t1616), M1 step 1
// RED tests for the fast-forward re-sync of the primary checkout's local main
// (REQ-LMF-014, plan §B3a). Each test runs runLocalMainResync against a real
// git fixture. The origin value is set with git update-ref in the scratch
// repository, and the fetch seam is replaced with a no-op, so no test contacts
// a remote. The stub in local_main_resync.go returns "not implemented", so each
// assertion below fails at its own check.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/session"
)

// lmfResyncFixture builds a primary checkout on main (lmfRepo), stamps the lane
// session and label, and replaces the origin fetch with a no-op for the test. The
// production fetch sets refs/remotes/origin/main, the ref the baseline is read from.
// The fixture sets that ref itself (lmfOriginCommit or update-ref), so the no-op
// leaves it as the test arranged it, and the re-sync takes BASELINE_SHA from it.
func lmfResyncFixture(t *testing.T, gitignore string) string {
	t.Helper()
	root, _ := lmfRepo(t, true, gitignore)
	t.Setenv(config.EnvClaudeCodeSessionID, lmfSession)
	sdLaneEnv(t, "lane-1", "")
	prev := localMainResyncFetch
	localMainResyncFetch = func(string) error { return nil }
	t.Cleanup(func() { localMainResyncFetch = prev })
	return root
}

// lmfLocalCommit commits path with content on the primary's local main and
// returns the new HEAD.
func lmfLocalCommit(t *testing.T, root, path, content string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, path), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	fcGit(t, root, "add", path)
	fcGit(t, root, "commit", "-q", "-m", "local "+path)
	return lmfHead(t, root)
}

// lmfOriginCommit commits top-level files on a scratch branch cut from base and
// points refs/remotes/origin/main at the new commit with git update-ref. Paths
// the primary ignores are added with force.
func lmfOriginCommit(t *testing.T, root, base, branch string, files map[string]string, force bool) string {
	t.Helper()
	wt := filepath.Join(t.TempDir(), branch)
	fcGit(t, root, "worktree", "add", "-q", "-b", branch, wt, base)
	add := []string{"add"}
	if force {
		add = append(add, "-f")
	}
	for p, content := range files {
		if err := os.WriteFile(filepath.Join(wt, p), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		add = append(add, p)
	}
	fcGit(t, wt, add...)
	fcGit(t, wt, "commit", "-q", "-m", branch)
	sha := fcGit(t, wt, "rev-parse", "HEAD")
	fcGit(t, root, "update-ref", "refs/remotes/origin/main", sha)
	return sha
}

// --- Re-sync of local main (REQ-LMF-014, plan §B3a) ---

func TestLocalMainResyncFastForwards(t *testing.T) {
	// Local main is one commit behind origin/main, which descends from it. The
	// re-sync moves local main to origin/main by fast-forward and reports both
	// SHAs.
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	target := lmfOriginCommit(t, root, base, "origin-ahead", map[string]string{"ahead.txt": "ahead\n"}, false)
	report, err := runLocalMainResync(root)
	if err != nil {
		t.Fatalf("local main behind origin/main must fast-forward: %v", err)
	}
	if head := lmfHead(t, root); head != target {
		t.Fatalf("local main must move from %s to origin/main %s, got HEAD %s", base, target, head)
	}
	if sym := fcGit(t, root, "symbolic-ref", "HEAD"); sym != "refs/heads/"+lmfBranch {
		t.Fatalf("the re-sync must not switch branches: HEAD names %q", sym)
	}
	if !strings.Contains(report, base[:12]) || !strings.Contains(report, target[:12]) {
		t.Fatalf("the report must name the old and new SHAs %s and %s: %q", base[:12], target[:12], report)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the re-sync must release the window: %+v", lock)
	}
}

func TestLocalMainResyncRefusesDiverged(t *testing.T) {
	// Local main holds a commit that origin/main lacks, and origin/main holds a
	// commit that local main lacks. The re-sync must refuse without reconciling.
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	local := lmfLocalCommit(t, root, "local.txt", "local\n")
	lmfOriginCommit(t, root, base, "origin-diverged", map[string]string{"remote.txt": "remote\n"}, false)
	_, err := runLocalMainResync(root)
	if err == nil {
		t.Fatalf("a diverged local main must refuse the re-sync")
	}
	// Plan §B3a names no class for this refusal. The nearest existing class is
	// MergeExitNotDescendant: the fast-forward target does not descend from HEAD.
	if code, ok := factory.MergeExitCode(err); !ok || code != factory.MergeExitNotDescendant {
		t.Fatalf("the diverged refusal must carry MergeExitNotDescendant (%d), got code %d (ok=%v): %v", factory.MergeExitNotDescendant, code, ok, err)
	}
	if head := lmfHead(t, root); head != local {
		t.Fatalf("a diverged refusal must not move HEAD: %s -> %s", local, head)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("a diverged refusal must release the window: %+v", lock)
	}
}

func TestLocalMainResyncAheadIsNoop(t *testing.T) {
	// Local main already contains origin/main. The re-sync reports that no
	// fast-forward is needed and leaves HEAD where it is.
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	local := lmfLocalCommit(t, root, "local.txt", "local\n")
	fcGit(t, root, "update-ref", "refs/remotes/origin/main", base)
	report, err := runLocalMainResync(root)
	if err != nil {
		t.Fatalf("a local main that contains origin/main must be a no-op, not an error: %v", err)
	}
	if head := lmfHead(t, root); head != local {
		t.Fatalf("the no-op must leave HEAD at %s, got %s", local, head)
	}
	if !strings.Contains(strings.ToLower(report), "no fast-forward") {
		t.Fatalf("the no-op must report that no fast-forward is needed: %q", report)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the no-op must release the window: %+v", lock)
	}
}

func TestLocalMainResyncPreservesIgnoredFile(t *testing.T) {
	// The primary's ignored data.txt sits at a path that origin/main adds. The
	// fast-forward would overwrite it, so the re-sync refuses with guidance, HEAD
	// stays put, and the file keeps its content (plan §B3a step 5a).
	root := lmfResyncFixture(t, "data.txt")
	base := lmfHead(t, root)
	if err := os.WriteFile(filepath.Join(root, "data.txt"), []byte("local-keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lmfOriginCommit(t, root, base, "origin-ignored", map[string]string{"data.txt": "card:data.txt\n"}, true)
	_, err := runLocalMainResync(root)
	if err == nil {
		t.Fatalf("an ignored file at a path the fast-forward would write must refuse the re-sync")
	}
	// Plan §B3a names no class for this refusal. The nearest existing class is
	// MergeExitCollision, the merge step's refusal for an ignored-byte overwrite.
	if code, ok := factory.MergeExitCode(err); !ok || code != factory.MergeExitCollision {
		t.Fatalf("the ignored-file refusal must carry MergeExitCollision (%d), got code %d (ok=%v): %v", factory.MergeExitCollision, code, ok, err)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "data.txt") {
		t.Fatalf("the refusal must name the ignored path data.txt: %v", err)
	}
	if head := lmfHead(t, root); head != base {
		t.Fatalf("an ignored-file refusal must not move HEAD: %s -> %s", base, head)
	}
	lmfAssertFile(t, filepath.Join(root, "data.txt"), "local-keep\n")
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the refusal must release the window: %+v", lock)
	}
}

// lmfResyncHold records the integration window for lmfSession as the re-sync's
// own acquisition does: the owner pid resolved from the session, the configured
// branch, and the primary checkout as the worktree. The session is the holder,
// so the re-sync that follows is the holder calling again.
func lmfResyncHold(t *testing.T, root string) *factory.IntegrationLock {
	t.Helper()
	t.Setenv(config.EnvClaudeCodeSessionID, lmfSession)
	ownerPID, _ := session.ResolveOwnerPID()
	if _, err := factory.AcquireIntegrationWindow(root, factory.IntegrationLock{
		SessionID:    lmfSession,
		PID:          ownerPID,
		PIDSource:    factory.PIDSourceSessionOwner,
		Branch:       localMainResyncBranch,
		BranchSource: factory.BranchSourceConfig,
		Worktree:     root,
	}, false, &factory.AcquireWindowOptions{LeaseDuration: integrationLeaseDuration(root)}); err != nil {
		t.Fatalf("hold the window for %s: %v", lmfSession, err)
	}
	held := sdWindow(t, root)
	if !held.Held() || held.SessionID != lmfSession {
		t.Fatalf("fixture: the window must be held by %s before the re-sync, got %+v", lmfSession, held)
	}
	t.Logf("held before the re-sync: session %s, pid %d (%s)", held.SessionID, held.PID, held.PIDSource)
	return held
}

// lmfAssertWindowKeptBy asserts that the window is still held by the session that
// held it before the re-sync, under the same holder identity: the session id and
// the owner pid with its source.
func lmfAssertWindowKeptBy(t *testing.T, root string, held *factory.IntegrationLock) {
	t.Helper()
	lock := sdWindow(t, root)
	if !lock.Held() || lock.SessionID != held.SessionID {
		t.Fatalf("the holder's re-sync must keep the window held by %s, got %+v", held.SessionID, lock)
	}
	if lock.PID != held.PID || lock.PIDSource != held.PIDSource {
		t.Fatalf("the holder identity must not change across the re-sync: pid %d (%s) became pid %d (%s)", held.PID, held.PIDSource, lock.PID, lock.PIDSource)
	}
}

// TestLocalMainResyncHolderCallKeepsWindow pins the holder-calls-resync rule
// (card t1616, leader ruling d-20261010T044429Z-b3cd). The re-sync re-acquires
// the window for the session that already holds it (plan §B3a step 1), and at
// step 8 it must release only a window this call took itself. A holder that
// calls the re-sync must still hold the window when the call returns. Both
// paths are pinned as subtests: the no-op path, where local main already
// contains origin/main, and the fast-forward path, where origin/main is ahead.
func TestLocalMainResyncHolderCallKeepsWindow(t *testing.T) {
	t.Run("no-op path", func(t *testing.T) {
		root := lmfResyncFixture(t, "")
		base := lmfHead(t, root)
		lmfLocalCommit(t, root, "local.txt", "local\n")
		fcGit(t, root, "update-ref", "refs/remotes/origin/main", base)
		held := lmfResyncHold(t, root)
		if _, err := runLocalMainResync(root); err != nil {
			t.Fatalf("the holder's no-op re-sync must not error: %v", err)
		}
		lmfAssertWindowKeptBy(t, root, held)
	})
	t.Run("fast-forward path", func(t *testing.T) {
		root := lmfResyncFixture(t, "")
		base := lmfHead(t, root)
		lmfOriginCommit(t, root, base, "origin-ahead", map[string]string{"ahead.txt": "ahead\n"}, false)
		held := lmfResyncHold(t, root)
		if _, err := runLocalMainResync(root); err != nil {
			t.Fatalf("the holder's fast-forward re-sync must not error: %v", err)
		}
		lmfAssertWindowKeptBy(t, root, held)
	})
}

// --- Sync-audit defects F2-F6 (card t1616): RED regression tests ---

const (
	// lmfRival is the session that takes the window while the re-sync's fetch runs (F2).
	lmfRival = "sess-rival"
	// lmfWaiterSession is the lane queued behind the holder (F3).
	lmfWaiterSession = "sess-lane-2"
)

// lmfResyncFixtureLiveFetch is lmfResyncFixture without the fetch swap: the
// production default `git fetch origin main` runs against the origin the test
// configures.
func lmfResyncFixtureLiveFetch(t *testing.T) string {
	t.Helper()
	root, _ := lmfRepo(t, true, "")
	t.Setenv(config.EnvClaudeCodeSessionID, lmfSession)
	sdLaneEnv(t, "lane-1", "")
	return root
}

// lmfAfterFastForward installs fn as the post-fast-forward seam for one test and
// restores the production no-op on cleanup.
func lmfAfterFastForward(t *testing.T, fn func(repoRoot string)) {
	t.Helper()
	prev := localMainResyncAfterFastForward
	localMainResyncAfterFastForward = fn
	t.Cleanup(func() { localMainResyncAfterFastForward = prev })
}

// lmfResyncHoldCarded records the window for lmfSession with a session name and a
// card, as the lane's own acquire does. The re-sync that follows is the holder
// calling again.
func lmfResyncHoldCarded(t *testing.T, root string) {
	t.Helper()
	ownerPID, _ := session.ResolveOwnerPID()
	if _, err := factory.AcquireIntegrationWindow(root, factory.IntegrationLock{
		SessionID:    lmfSession,
		SessionName:  "lane-1",
		Card:         lmfCard,
		PID:          ownerPID,
		PIDSource:    factory.PIDSourceSessionOwner,
		Branch:       localMainResyncBranch,
		BranchSource: factory.BranchSourceConfig,
		Worktree:     root,
	}, false, &factory.AcquireWindowOptions{LeaseDuration: integrationLeaseDuration(root)}); err != nil {
		t.Fatalf("hold the window for %s: %v", lmfSession, err)
	}
	held := sdWindow(t, root)
	if !held.Held() || held.SessionID != lmfSession || held.SessionName != "lane-1" || held.Card != lmfCard {
		t.Fatalf("fixture: the window must be held by %s as lane-1 on card %s before the re-sync, got %+v", lmfSession, lmfCard, held)
	}
}

// lmfQueueWaiter queues a live lane behind the held window. The ticket names this
// test process as both its owner and its waiter, so the production liveness probe
// keeps it: a release that promotes the queue promotes this lane.
func lmfQueueWaiter(t *testing.T, root string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339)
	if err := factory.UpdateIntegrationWindow(root, func(w *factory.IntegrationLock) error {
		w.Queue = append(w.Queue, factory.IntegrationTicket{
			SessionID:    lmfWaiterSession,
			SessionName:  "lane-2",
			Card:         "t2",
			OwnerPID:     os.Getpid(),
			PIDSource:    factory.PIDSourceSessionOwner,
			Branch:       localMainResyncBranch,
			BranchSource: factory.BranchSourceConfig,
			Worktree:     root,
			WaiterPID:    os.Getpid(),
			WaiterStart:  homestate.CurrentProcessFingerprint(),
			Heartbeat:    now,
			EnqueuedAt:   now,
		})
		return nil
	}); err != nil {
		t.Fatalf("queue a waiter behind the window: %v", err)
	}
}

func TestLocalMainResyncTakeoverDuringFetchRefuses(t *testing.T) {
	// F2 (sync audit, card t1616): the holder check runs when the window is taken,
	// and nothing re-validates it before the fast-forward. A rival session that
	// takes the window while the fetch runs must stop the re-sync: it refuses, HEAD
	// does not move, and the rival keeps the window. The rival takes the window
	// through the real --force acquire.
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	target := lmfOriginCommit(t, root, base, "origin-ahead", map[string]string{"ahead.txt": "ahead\n"}, false)
	localMainResyncFetch = func(repoRoot string) error {
		if _, err := factory.AcquireIntegrationLock(repoRoot, factory.IntegrationLock{
			SessionID:    lmfRival,
			SessionName:  "lane-rival",
			PID:          os.Getpid(),
			PIDSource:    factory.PIDSourceSessionOwner,
			Branch:       localMainResyncBranch,
			BranchSource: factory.BranchSourceConfig,
			Worktree:     repoRoot,
		}, true); err != nil {
			t.Fatalf("fixture: the rival takeover while the fetch runs: %v", err)
		}
		return nil
	}
	_, err := runLocalMainResync(root)
	if err == nil {
		t.Fatalf("a takeover while the fetch runs must refuse the re-sync")
	}
	if head := lmfHead(t, root); head != base {
		t.Fatalf("a takeover during the fetch must not fast-forward: HEAD moved %s -> %s (origin/main is %s)", base, head, target)
	}
	if lock := sdWindow(t, root); lock.SessionID != lmfRival {
		t.Fatalf("the rival's window must stay in place after the refused re-sync: holder %q", lock.SessionID)
	}
}

func TestLocalMainResyncPostMergeWritesHold(t *testing.T) {
	// F3 (sync audit, card t1616): when step 6 finds HEAD off BASELINE_SHA after
	// the fast-forward, the refusal is MergeExitPostMerge. The merge step writes a
	// post-merge hold before it releases the window (integration_merge_step.go:554-623).
	// The re-sync releases with no hold, so a lane queued behind the window is
	// promoted onto the anomalous state. The anomaly comes from the post-fast-forward
	// seam, which moves the branch back to its pre-fast-forward commit. The waiter is
	// queued during the fetch, while the re-sync holds the window.
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	target := lmfOriginCommit(t, root, base, "origin-ahead", map[string]string{"ahead.txt": "ahead\n"}, false)
	localMainResyncFetch = func(repoRoot string) error {
		lmfQueueWaiter(t, repoRoot)
		return nil
	}
	lmfAfterFastForward(t, func(repoRoot string) {
		fcGit(t, repoRoot, "update-ref", "refs/heads/"+lmfBranch, base)
	})
	_, err := runLocalMainResync(root)
	if code, ok := factory.MergeExitCode(err); !ok || code != factory.MergeExitPostMerge {
		t.Fatalf("HEAD off BASELINE_SHA after the fast-forward must refuse with MergeExitPostMerge (%d), got code %d (ok=%v): %v", factory.MergeExitPostMerge, code, ok, err)
	}
	policy, policyErr := factory.ReadIntegrationWindowPolicy(root)
	if policyErr != nil || policy.Policy != factory.PolicyHold {
		t.Fatalf("the post-merge refusal must write a hold before it releases the window: policy %+v (err %v)", policy, policyErr)
	}
	if !strings.Contains(strings.ToLower(policy.Reason), "fast-forward") || !strings.Contains(policy.Reason, target[:12]) {
		t.Fatalf("the hold must name the cause and the SHA %s: %q", target[:12], policy.Reason)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the window must be released after the hold, with no promotion onto the anomalous state: holder %q", lock.SessionID)
	}
}

func TestLocalMainResyncKeepsCardMetadata(t *testing.T) {
	// F4 (sync audit, card t1616): the re-sync's acquisition writes the window record
	// from a literal that names neither the card nor the session name, and the holder
	// re-entry keeps that literal. The holder's no-op re-sync must leave the card and
	// the session name the window was taken for.
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	lmfLocalCommit(t, root, "local.txt", "local\n")
	fcGit(t, root, "update-ref", "refs/remotes/origin/"+localMainResyncBranch, base)
	lmfResyncHoldCarded(t, root)
	if _, err := runLocalMainResync(root); err != nil {
		t.Fatalf("the holder's no-op re-sync must not error: %v", err)
	}
	lock := sdWindow(t, root)
	if lock.Card != lmfCard {
		t.Fatalf("the holder's re-sync must keep the card %s the window was taken for, got %q", lmfCard, lock.Card)
	}
	if lock.SessionName != "lane-1" {
		t.Fatalf("the holder's re-sync must keep the session name lane-1, got %q", lock.SessionName)
	}
}

func TestLocalMainResyncLateDirtyPrimaryRefuses(t *testing.T) {
	// F5 (sync audit, card t1616): the primary's clean check runs before the fetch.
	// A file that appears in the primary while the fetch runs goes unseen, and
	// `git merge --ff-only` refuses only collisions, so the fast-forward lands on a
	// dirty primary. The re-sync must re-check the clean state before it moves HEAD.
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	lmfOriginCommit(t, root, base, "origin-ahead", map[string]string{"ahead.txt": "ahead\n"}, false)
	localMainResyncFetch = func(repoRoot string) error {
		if err := os.WriteFile(filepath.Join(repoRoot, "late-untracked.txt"), []byte("late\n"), 0o644); err != nil {
			return err
		}
		return nil
	}
	_, err := runLocalMainResync(root)
	if err == nil {
		t.Fatalf("a primary that turns dirty during the fetch must refuse the fast-forward")
	}
	if head := lmfHead(t, root); head != base {
		t.Fatalf("a dirty-primary refusal must not move HEAD: %s -> %s", base, head)
	}
	got, readErr := os.ReadFile(filepath.Join(root, "late-untracked.txt"))
	if readErr != nil || string(got) != "late\n" {
		t.Fatalf("the refusal must leave the late file untouched: %q (err %v)", got, readErr)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the refusal must release the window: %+v", lock)
	}
}

func TestLocalMainResyncFetchPinsFetchedCommit(t *testing.T) {
	// F6 (sync audit, card t1616): BASELINE_SHA is read from refs/remotes/origin/main.
	// Under a fetch refspec that does not map refs/heads/main, a fetch that names no
	// destination leaves that ref at the commit the primary last fetched, so the
	// re-sync compares against a stale baseline. The production fetch names its
	// destination (leader ruling (A), 2026-10-10), so the ref moves to the fetched
	// commit even under that refspec. origin is a local bare repository in the test's
	// temporary directory, and the production fetch seam runs.
	root := lmfResyncFixtureLiveFetch(t)
	base := lmfHead(t, root)
	bare := filepath.Join(t.TempDir(), "origin.git")
	fcGit(t, t.TempDir(), "init", "-q", "--bare", bare)
	fcGit(t, bare, "symbolic-ref", "HEAD", "refs/heads/"+lmfBranch)
	fcGit(t, root, "remote", "add", "origin", bare)
	fcGit(t, root, "push", "-q", "origin", "refs/heads/"+lmfBranch+":refs/heads/"+lmfBranch)
	fcGit(t, root, "update-ref", "refs/remotes/origin/"+lmfBranch, base)
	fcGit(t, root, "config", "remote.origin.fetch", "+refs/heads/feature:refs/remotes/origin/feature")
	clone := filepath.Join(t.TempDir(), "clone")
	fcGit(t, t.TempDir(), "clone", "-q", bare, clone)
	if err := os.WriteFile(filepath.Join(clone, "remote.txt"), []byte("remote\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fcGit(t, clone, "add", "remote.txt")
	fcGit(t, clone, "commit", "-q", "-m", "origin main advances")
	fcGit(t, clone, "push", "-q", "origin", "HEAD:refs/heads/"+lmfBranch)
	fetched := fcGit(t, clone, "rev-parse", "HEAD")
	if stale := fcGit(t, root, "rev-parse", "refs/remotes/origin/"+lmfBranch); stale != base {
		t.Fatalf("fixture: refs/remotes/origin/main must stay at the last fetched %s, got %s", base, stale)
	}
	report, err := runLocalMainResync(root)
	if err != nil {
		t.Fatalf("local main must fast-forward to the fetched origin/main: %v", err)
	}
	if head := lmfHead(t, root); head != fetched {
		t.Fatalf("the re-sync must fast-forward to the fetched commit %s, not the stale remote-tracking ref %s: HEAD is %s", fetched, base, head)
	}
	if !strings.Contains(report, fetched[:12]) {
		t.Fatalf("the report must name the fetched commit %s: %q", fetched[:12], report)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the re-sync must release the window it took: %+v", lock)
	}
}

// TestLocalMainResyncForeignFetchDoesNotMoveBaseline pins the baseline source (card
// t1616, leader ruling (A), 2026-10-10). A fetch of another ref in this repository
// overwrites FETCH_HEAD between our fetch and our read, so a BASELINE_SHA taken from
// FETCH_HEAD can name a foreign commit. The seam runs that order: the origin/main
// fetch leaves refs/remotes/origin/main at X, then another session's fetch leaves
// FETCH_HEAD at F, a commit that also descends from the primary's HEAD. The re-sync
// must fast-forward to X, not to F.
func TestLocalMainResyncForeignFetchDoesNotMoveBaseline(t *testing.T) {
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	foreign := lmfOriginCommit(t, root, base, "origin-f", map[string]string{"foreign.txt": "foreign\n"}, false)
	target := lmfOriginCommit(t, root, base, "origin-x", map[string]string{"ahead.txt": "ahead\n"}, false)
	prev := localMainResyncFetch
	localMainResyncFetch = func(repoRoot string) error {
		// The origin/main fetch leaves refs/remotes/origin/main at X.
		fcGit(t, repoRoot, "update-ref", "refs/remotes/origin/"+lmfBranch, target)
		// Another session's fetch of refs/heads/origin-f overwrites FETCH_HEAD with F.
		fcGit(t, repoRoot, "fetch", "-q", ".", "refs/heads/origin-f")
		return nil
	}
	t.Cleanup(func() { localMainResyncFetch = prev })
	if _, err := runLocalMainResync(root); err != nil {
		t.Fatalf("local main behind origin/main must fast-forward: %v", err)
	}
	if head := lmfHead(t, root); head != target {
		t.Fatalf("BASELINE_SHA must come from origin/main %s, not from FETCH_HEAD, which names the foreign commit %s: HEAD is %s", target, foreign, head)
	}
}

// TestLocalMainResyncHoldWriteFailureKeepsWindow pins the retained-refusal path (card
// t1616, F3). After a post-merge anomaly the policy hold is written before the window
// is released. When that write fails, the re-sync returns both the anomaly and the
// hold failure, and it keeps the window held, so no queued lane is promoted onto the
// anomalous state. The write fails deterministically: a directory sits at the policy
// record's path, so the write's rename onto that path fails. The directory is placed
// inside the post-fast-forward seam, because the acquire reads the policy record
// first. A probe confirms the cause before the assertions rely on it.
func TestLocalMainResyncHoldWriteFailureKeepsWindow(t *testing.T) {
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	lmfOriginCommit(t, root, base, "origin-ahead", map[string]string{"ahead.txt": "ahead\n"}, false)
	policyPath := filepath.Join(root, ".moai", "state", factory.IntegrationWindowPolicyFileName)
	lmfAfterFastForward(t, func(repoRoot string) {
		if err := os.MkdirAll(policyPath, 0o755); err != nil {
			t.Fatalf("fixture: place a directory at the policy record's path: %v", err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(policyPath) })
		probe := factory.WriteIntegrationWindowPolicy(repoRoot, factory.IntegrationWindowPolicy{Policy: factory.PolicyHold, Reason: "probe", SetBy: "test", SetAt: "probe"})
		var linkErr *os.LinkError
		if !errors.As(probe, &linkErr) || linkErr.Op != "rename" || linkErr.New != policyPath {
			t.Fatalf("fixture: the hold write must fail renaming onto %s, got %v", policyPath, probe)
		}
		// Move HEAD back, so step 6 finds HEAD off BASELINE_SHA.
		fcGit(t, repoRoot, "update-ref", "refs/heads/"+lmfBranch, base)
	})
	_, err := runLocalMainResync(root)
	if err == nil {
		t.Fatalf("an anomaly whose hold cannot be written must still refuse the re-sync")
	}
	if code, ok := factory.MergeExitCode(err); !ok || code != factory.MergeExitPostMerge {
		t.Fatalf("the anomaly must refuse with MergeExitPostMerge (%d), got code %d (ok=%v): %v", factory.MergeExitPostMerge, code, ok, err)
	}
	if !strings.Contains(err.Error(), "after the fast-forward HEAD is not origin/"+lmfBranch) {
		t.Fatalf("the refusal must name the anomaly: %v", err)
	}
	if !strings.Contains(err.Error(), "writing the hold also failed") {
		t.Fatalf("the refusal must name the hold failure: %v", err)
	}
	if head := lmfHead(t, root); head != base {
		t.Fatalf("the anomaly must leave HEAD at its position %s, got %s", base, head)
	}
	if lock := sdWindow(t, root); !lock.Held() || lock.SessionID != lmfSession {
		t.Fatalf("a hold that cannot be written must keep the window held by %s, got %+v", lmfSession, lock)
	}
}

func TestLocalMainResyncPostFastForwardLateTxtRefuses(t *testing.T) {
	// F5 (sync audit iteration 2, card t1616): after the fast-forward the re-sync checks
	// only HEAD and the symbolic ref (step 6), then returns success (step 7) without
	// reading the primary's status. A file that appears in the primary after the move
	// goes unseen, so the re-sync reports success on a dirty primary. The re-sync must
	// read the status after the fast-forward. A non-empty status writes the policy hold,
	// which names the porcelain line and the fast-forward target SHA, and refuses with
	// the merge-dirty class before the window is released.
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	target := lmfOriginCommit(t, root, base, "origin-ahead", map[string]string{"ahead.txt": "ahead\n"}, false)
	lmfAfterFastForward(t, func(repoRoot string) {
		if err := os.WriteFile(filepath.Join(repoRoot, "late.txt"), []byte("late\n"), 0o644); err != nil {
			t.Fatalf("fixture: write the late file after the fast-forward: %v", err)
		}
	})
	_, err := runLocalMainResync(root)
	if err == nil {
		t.Fatalf("a primary that turns dirty after the fast-forward must refuse the re-sync, not return success")
	}
	if code, ok := factory.MergeExitCode(err); !ok || code != factory.MergeExitMergeDirty {
		t.Fatalf("the post-fast-forward dirty refusal must carry MergeExitMergeDirty (%d), got code %d (ok=%v): %v", factory.MergeExitMergeDirty, code, ok, err)
	}
	policy, policyErr := factory.ReadIntegrationWindowPolicy(root)
	if policyErr != nil || policy.Policy != factory.PolicyHold {
		t.Fatalf("the dirty refusal must write a hold before it releases the window: policy %+v (err %v)", policy, policyErr)
	}
	if !strings.Contains(policy.Reason, "?? late.txt") {
		t.Fatalf("the hold must name the porcelain line ?? late.txt: %q", policy.Reason)
	}
	if !strings.Contains(policy.Reason, target[:12]) {
		t.Fatalf("the hold must name the fast-forward target SHA %s: %q", target[:12], policy.Reason)
	}
	// The anomaly path leaves the fast-forward in place for the leader, so HEAD stays at
	// the target, and the refusal never removes the late file.
	if head := lmfHead(t, root); head != target {
		t.Fatalf("the dirty refusal must leave the fast-forward in place at %s, got HEAD %s", target, head)
	}
	got, readErr := os.ReadFile(filepath.Join(root, "late.txt"))
	if readErr != nil || string(got) != "late\n" {
		t.Fatalf("the refusal must leave the late file untouched: %q (err %v)", got, readErr)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the dirty refusal must release the window after the hold: %+v", lock)
	}
}
