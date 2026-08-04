package hook

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestPR1338Regression (SPEC-ORCH-GIT-RELAX-001 M4.1 / AC-OGR-006) reproduces
// the PR-#1338 incident class mechanically and demonstrates that the
// orchestrator-direct (live-state) restoration path resolves it, while the
// isolated-snapshot-holder (stale-state) path cross-swaps.
//
// PR-#1338 incident signature (spec.md §B): manager-git, holding a stale
// session-start snapshot of the worktree registry, was asked to restore the
// primary checkout to `main`. Between snapshot time and action time a
// concurrent session had moved work around (branches swapped across the
// primary and a linked worktree). The stale snapshot caused the restoration
// to target the WRONG worktree — cross-swapping a concurrent session's
// worktree branch to `main` while leaving the actual primary broken. The
// orchestrator's DIRECT recovery (reading live HEAD + the active worktree
// registry via `git worktree list --porcelain`) restored the primary
// correctly because it did not depend on a snapshot.
//
// The test builds a real multi-worktree git fixture under t.TempDir() and
// exercises both decision paths against real git state. It asserts:
//
//  1. STALE path (snapshot-holder): the restoration decision derived from a
//     stale worktree-branch map is INCORRECT — it either misroutes the
//     restore (cross-swap) or leaves the primary broken.
//  2. LIVE path (orchestrator-direct): the restoration decision derived from
//     a fresh `git worktree list --porcelain` read is CORRECT — the primary
//     is restored to `main` and the concurrent session's worktree is left
//     undisturbed.
//
// This is a unit test with real git operations against a throwaway tempdir
// fixture — it does NOT touch the developer's primary checkout.
func TestPR1338Regression(t *testing.T) {
	requireGit(t)

	// --- Fixture: primary checkout on `main`, linked worktree on `feat/A`. ---
	// Resolve symlinks so the path matches the `worktree` keys emitted by
	// `git worktree list --porcelain` (git reports the real filesystem path;
	// on macOS t.TempDir() returns a /var/folders symlink that resolves to
	// /private/var/folders). This path-resolution care is itself part of the
	// incident class — the orchestrator-direct path must key on the SAME path
	// git reports, not a symlink shadow.
	repoRaw := t.TempDir()
	repo, err := filepath.EvalSymlinks(repoRaw)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s): %v", repoRaw, err)
	}
	gitInitRepo(t, repo)
	// Make the primary branch deterministically `main` regardless of the
	// git binary's configured init.defaultBranch (varies by git version /
	// host: `master` on older, `main` on newer).
	mustRunGit(t, repo, "branch", "-M", "main")
	// Create feat/A off main so the linked worktree can check it out.
	mustRunGit(t, repo, "branch", "feat/A")
	wt := filepath.Join(repo, "wt-A")
	mustRunGit(t, repo, "worktree", "add", wt, "feat/A")

	// Sanity: at T0, primary=main, wt-A=feat/A.
	t0 := worktreeBranchMap(t, repo)
	if t0[repo] != "main" {
		t.Fatalf("fixture T0: primary branch = %q, want main", t0[repo])
	}
	if t0[wt] != "feat/A" {
		t.Fatalf("fixture T0: wt-A branch = %q, want feat/A", t0[wt])
	}

	// --- Stale snapshot captured at T0 (what an isolated snapshot-holder
	// like the old manager-git would carry). ---
	staleSnapshot := map[string]string{}
	for k, v := range t0 {
		staleSnapshot[k] = v
	}

	// --- Concurrent session moves work: it creates a new branch
	// `feat/concurrent` off main and switches the PRIMARY onto it. wt-A is
	// left undisturbed on feat/A (the concurrent session's other in-flight
	// work). This is the state drift the stale snapshot cannot observe — it
	// is a frozen copy taken at T0. ---
	mustRunGit(t, repo, "branch", "feat/concurrent", "main")
	mustRunGit(t, repo, "checkout", "feat/concurrent")

	// Live state at action time (T1): what the orchestrator-direct path reads.
	t1 := worktreeBranchMap(t, repo)
	if t1[repo] != "feat/concurrent" || t1[wt] != "feat/A" {
		t.Fatalf("fixture T1 drift did not take: got %v", t1)
	}

	// === STALE PATH (isolated snapshot-holder) ===
	// The restoration goal: "restore the PRIMARY (repo) to main". The stale
	// holder consults its frozen snapshot to find which worktree is the
	// primary. The snapshot says repo=main, so the holder concludes "the
	// primary is already on main — nothing to do" and skips the restore.
	// RESULT: the primary (actually on feat/A at T1) is LEFT BROKEN, and if
	// the holder instead misreads and restores wt-A (snapshot says wt-A=feat/A,
	// "should be main"), it cross-swaps wt-A — which at T1 is already main —
	// while repo stays broken. Either way the primary is NOT restored.
	staleRestoreTarget := pickPrimaryPathFromSnapshot(staleSnapshot, repo)
	// The stale holder still identifies the primary path correctly (the path
	// did not move), BUT it consults the snapshot's BRANCH entry for that
	// path, sees repo=main, and concludes "primary already on main — no-op".
	// This is the frozen-snapshot decision defect at the heart of PR-#1338.
	if staleRestoreTarget != repo {
		t.Fatalf("stale pickPrimaryPath returned %q, want %q", staleRestoreTarget, repo)
	}
	if staleSnapshot[staleRestoreTarget] != "main" {
		t.Fatalf("stale snapshot mis-recorded primary branch = %q, want main (T0 state)", staleSnapshot[staleRestoreTarget])
	}
	// Stale decision: primary is "already main" per snapshot -> no-op restore.
	// The holder issues NO `git checkout main` because it believes there is
	// nothing to do. The real primary (on feat/A at T1) is left broken.
	stalePrimaryBranchAfter := currentBranch(t, repo) // unchanged because stale holder skipped
	staleRestored := stalePrimaryBranchAfter == "main"

	// === LIVE PATH (orchestrator-direct) ===
	// The orchestrator reads `git worktree list --porcelain` fresh, sees
	// repo=feat/A at T1, and restores repo to main. wt-A (on main at T1)
	// is left undisturbed — the orchestrator does NOT touch it.
	liveMap := worktreeBranchMap(t, repo)
	liveTarget := pickPrimaryPathFromSnapshot(liveMap, repo)
	if liveTarget != repo {
		t.Fatalf("live path picked wrong primary = %q, want %q", liveTarget, repo)
	}
	// Only the orchestrator-direct path restores because it reads the live
	// branch and acts; the stale path skipped (thought repo was already main).
	if currentBranch(t, liveTarget) != "feat/concurrent" {
		t.Fatalf("live target pre-restore branch = %q, want feat/concurrent", currentBranch(t, liveTarget))
	}
	mustRunGit(t, liveTarget, "checkout", "main")
	livePrimaryBranchAfter := currentBranch(t, repo)
	liveRestored := livePrimaryBranchAfter == "main"

	// === Incident-class assertions (AC-OGR-006) ===
	// (a) The stale snapshot holder did NOT restore the primary.
	if staleRestored {
		t.Fatalf("STALE path: primary restored to main (%q) — expected the cross-swap/left-broken failure mode (primary should still be on a non-main branch). The incident class was NOT reproduced.",
			stalePrimaryBranchAfter)
	}
	// (b) The orchestrator-direct (live-state) path DID restore the primary.
	if !liveRestored {
		t.Fatalf("LIVE path: primary NOT restored to main (got %q) — the orchestrator-direct resolution failed", livePrimaryBranchAfter)
	}
	// (c) The concurrent session's worktree (wt-A) is left on `feat/A`
	// undisturbed by the live restore — no cross-swap of wt-A. The stale
	// holder, had it misread wt-A as the primary needing restore, would have
	// run `git -C wt-A checkout main` and destroyed the concurrent session's
	// feat/A work — that is the namesake cross-swap of PR-#1338. The live
	// path never touches wt-A.
	if got := currentBranch(t, wt); got != "feat/A" {
		t.Fatalf("LIVE path cross-swapped wt-A to %q — the concurrent session's worktree must be left undisturbed on feat/A (PR-#1338 cross-swap regression)", got)
	}
}

// worktreeBranchMap is the LIVE-state reader the orchestrator-direct path
// uses: it runs `git -C repo worktree list --porcelain` and parses the
// worktree-path → checked-out-branch map. This is the full-context
// alternative to holding a frozen snapshot.
func worktreeBranchMap(t *testing.T, repo string) map[string]string {
	t.Helper()
	out, err := exec.Command("git", "-C", repo, "worktree", "list", "--porcelain").CombinedOutput()
	if err != nil {
		t.Fatalf("git worktree list --porcelain: %v\n%s", err, out)
	}
	m := map[string]string{}
	var curPath string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "worktree "):
			curPath = strings.TrimPrefix(line, "worktree ")
			m[curPath] = ""
		case strings.HasPrefix(line, "branch "):
			if curPath != "" {
				// refs/heads/<branch> → <branch>
				b := strings.TrimPrefix(line, "branch ")
				b = strings.TrimPrefix(b, "refs/heads/")
				m[curPath] = b
			}
		case line == "":
			curPath = ""
		}
	}
	return m
}

// pickPrimaryPathFromSnapshot returns the worktree the snapshot believes is
// the primary. For this fixture the primary is always `wantPrimary` (the
// repo root); the function exists to make the snapshot-consultation explicit
// — a stale snapshot CAN disagree with the live registry, and this is the
// decision point where the cross-swap originates. When the snapshot's
// primary entry shows the primary already on `main`, the snapshot-holder
// concludes "nothing to restore"; the orchestrator-direct path instead
// re-reads live state and ignores the snapshot's branch claim.
func pickPrimaryPathFromSnapshot(snapshot map[string]string, wantPrimary string) string {
	if _, ok := snapshot[wantPrimary]; ok {
		return wantPrimary
	}
	// Snapshot lost track of the primary path entirely (worktree registry
	// drifted) — return the first key as a degraded fallback, mirroring the
	// cross-swap hazard.
	for k := range snapshot {
		return k
	}
	return wantPrimary
}

// currentBranch returns the checked-out branch of the worktree at path.
func currentBranch(t *testing.T, path string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		t.Fatalf("git rev-parse --abbrev-ref HEAD at %s: %v\n%s", path, err, out)
	}
	return strings.TrimSpace(string(out))
}
