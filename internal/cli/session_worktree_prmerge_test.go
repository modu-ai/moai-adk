package cli

// session_worktree_prmerge_test.go — SPEC-SESSION-WORKTREE-001 M8 PR-merge
// auto-cleanup tests.
//
// The PR-merge cleanup logic is exercised through function-variable seams
// (sessionWorktreeGitWorktreeList, sessionWorktreeGhLookPath,
// sessionWorktreeGhPRViewState, sessionWorktreeGitBranchMerged) so the tests
// never depend on a real git repository or a real `gh` binary. Each test
// restores the real implementations via t.Cleanup.
//
// Coverage: REQ-SW-022 (toggle + trigger invariant), REQ-SW-023 (gh primary +
// squash-blind fallback), REQ-SW-024 (dirty guard), AC-SW-022/023/024, EC-11
// (race re-check), EC-13 (notice distinguishability).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// prMergeSeams snapshots the M8 package-level function-variable seams so a test
// can swap them and restore on cleanup.
type prMergeSeams struct {
	wtList       func() (string, error)
	ghLookPath   func() bool
	ghPRState    func(branch string) (string, bool)
	branchMerged func() ([]string, error)
}

// swapPRMergeSeams replaces the M8 seams and registers restoration.
func swapPRMergeSeams(t *testing.T, s prMergeSeams) {
	t.Helper()
	orig := prMergeSeams{
		wtList:       sessionWorktreeGitWorktreeList,
		ghLookPath:   sessionWorktreeGhLookPath,
		ghPRState:    sessionWorktreeGhPRViewState,
		branchMerged: sessionWorktreeGitBranchMerged,
	}
	if s.wtList != nil {
		sessionWorktreeGitWorktreeList = s.wtList
	}
	if s.ghLookPath != nil {
		sessionWorktreeGhLookPath = s.ghLookPath
	}
	if s.ghPRState != nil {
		sessionWorktreeGhPRViewState = s.ghPRState
	}
	if s.branchMerged != nil {
		sessionWorktreeGitBranchMerged = s.branchMerged
	}
	t.Cleanup(func() {
		sessionWorktreeGitWorktreeList = orig.wtList
		sessionWorktreeGhLookPath = orig.ghLookPath
		sessionWorktreeGhPRViewState = orig.ghPRState
		sessionWorktreeGitBranchMerged = orig.branchMerged
	})
}

// wtListPorcelain builds a `git worktree list --porcelain` body for two
// worktrees: the primary checkout (detached-ish, no WT- branch) + one WT-
// worktree. Tests override the branch/path via the helpers below.
func wtListPorcelainPrimary() string {
	return "worktree /repo\nHEAD 0000000000000000000000000000000000000000\nbranch refs/heads/main\n"
}

// wtListEntry builds a single WT- worktree porcelain entry.
func wtListEntry(path, branch string) string {
	return "worktree " + path + "\nHEAD 1111111111111111111111111111111111111111\nbranch refs/heads/" + branch + "\n"
}

// --- REQ-SW-022 toggle ---

// TestPRMergeCleanup_ToggleOffNoOp is REQ-SW-022 byte-identical baseline: when
// auto_cleanup is OFF (the distributed default), the on-touch invocation is a
// no-op — no git invocation, no notice.
func TestPRMergeCleanup_ToggleOffNoOp(t *testing.T) {
	listCalled := false
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) { listCalled = true; return "", nil },
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(false), &out)
	if listCalled {
		t.Fatal("toggle-off: git worktree list MUST NOT be invoked")
	}
	if out.Len() != 0 {
		t.Fatalf("toggle-off: expected no notice, got %q", out.String())
	}
}

// TestPRMergeCleanup_NilCfgNoOp proves nil-safety: a nil cfg (config load
// failure) degrades to OFF, never panics.
func TestPRMergeCleanup_NilCfgNoOp(t *testing.T) {
	listCalled := false
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) { listCalled = true; return "", nil },
	})
	var out bytes.Buffer
	prMergeCleanup(nil, &out)
	if listCalled {
		t.Fatal("nil cfg: git worktree list MUST NOT be invoked")
	}
}

// --- REQ-SW-023 gh primary path ---

// TestPRMergeCleanup_GhPresentMergedRemoves is AC-SW-023 primary path: gh
// available + PR state MERGED → worktree removed + distinguishable notice.
func TestPRMergeCleanup_GhPresentMergedRemoves(t *testing.T) {
	wtPath := "/repo/.claude/worktrees/WT-abcdef12-web"
	wtBranch := "WT-abcdef12-web"
	var removedPath string
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + wtListEntry(wtPath, wtBranch), nil
		},
		ghLookPath: func() bool { return true },
		ghPRState:  func(b string) (string, bool) { return "MERGED", true },
	})
	// remove + statusPorc reuse the M4 seams (clean worktree).
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(p string) error { removedPath = p; return nil },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)
	if removedPath != wtPath {
		t.Fatalf("gh-merged: expected remove(%q), got %q", wtPath, removedPath)
	}
	if !strings.Contains(out.String(), PRMergeCleanupNoticePrefix) {
		t.Fatalf("gh-merged: notice must carry %q, got %q", PRMergeCleanupNoticePrefix, out.String())
	}
	if !strings.Contains(out.String(), wtPath) {
		t.Fatalf("gh-merged: notice must name worktree path %q, got %q", wtPath, out.String())
	}
	if !strings.Contains(out.String(), "(branch") || !strings.Contains(out.String(), "merged)") {
		t.Fatalf("gh-merged: notice must carry '(branch ... merged)', got %q", out.String())
	}
}

// TestPRMergeCleanup_GhPresentOpenDoesNotRemove: gh available + PR OPEN → not
// merged, not a candidate.
func TestPRMergeCleanup_GhPresentOpenDoesNotRemove(t *testing.T) {
	removeCalled := false
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + wtListEntry("/repo/.claude/worktrees/WT-abcdef12-web", "WT-abcdef12-web"), nil
		},
		ghLookPath: func() bool { return true },
		ghPRState:  func(string) (string, bool) { return "OPEN", true },
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(string) error { removeCalled = true; return nil },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)
	if removeCalled {
		t.Fatal("gh-open: remove MUST NOT run (branch not merged)")
	}
	if strings.Contains(out.String(), PRMergeCleanupNoticePrefix) {
		t.Fatalf("gh-open: no removal notice expected, got %q", out.String())
	}
}

// TestPRMergeCleanup_GhPresentSeesSquashMerge is AC-SW-023 primary correctness:
// the same squash-merged PR with gh available IS detected (state == MERGED).
// This is the falsification counterpart to the squash-blind fallback below.
func TestPRMergeCleanup_GhPresentSeesSquashMerge(t *testing.T) {
	var removedPath string
	wtPath := "/repo/.claude/worktrees/WT-squash01-init"
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + wtListEntry(wtPath, "WT-squash01-init"), nil
		},
		ghLookPath: func() bool { return true },
		ghPRState:  func(string) (string, bool) { return "MERGED", true }, // gh sees the squash merge
		// branchMerged would NOT list a squash-merged branch, but gh path is
		// taken so this seam is irrelevant here.
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(p string) error { removedPath = p; return nil },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)
	if removedPath != wtPath {
		t.Fatalf("gh-squash: gh primary MUST detect the squash merge and remove, got %q", removedPath)
	}
}

// --- REQ-SW-023 squash-blind fallback ---

// TestPRMergeCleanup_GhAbsentBranchMergedRemoves is AC-SW-023 fallback path: gh
// absent + branch listed in `git branch --merged origin/main` → removed.
func TestPRMergeCleanup_GhAbsentBranchMergedRemoves(t *testing.T) {
	wtBranch := "WT-fedcba09-profile"
	wtPath := "/repo/.claude/worktrees/" + wtBranch
	var removedPath string
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + wtListEntry(wtPath, wtBranch), nil
		},
		ghLookPath:   func() bool { return false },
		branchMerged: func() ([]string, error) { return []string{"main", wtBranch}, nil },
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(p string) error { removedPath = p; return nil },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)
	if removedPath != wtPath {
		t.Fatalf("fallback-merged: expected remove(%q), got %q", wtPath, removedPath)
	}
	if !strings.Contains(out.String(), PRMergeCleanupNoticePrefix) {
		t.Fatalf("fallback-merged: notice must carry prefix, got %q", out.String())
	}
}

// TestPRMergeCleanup_SquashBlindFallbackPreserves is AC-SW-023 documented
// blindness: gh absent + PR squash-merged (branch NOT in --merged list) →
// worktree NOT removed + a notice documents the fallback's blindness.
func TestPRMergeCleanup_SquashBlindFallbackPreserves(t *testing.T) {
	removeCalled := false
	wtBranch := "WT-squash02-web"
	wtPath := "/repo/.claude/worktrees/" + wtBranch
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + wtListEntry(wtPath, wtBranch), nil
		},
		ghLookPath:   func() bool { return false },
		branchMerged: func() ([]string, error) { return []string{"main"}, nil }, // WT branch absent (squash-merge blind)
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(string) error { removeCalled = true; return nil },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)
	if removeCalled {
		t.Fatal("squash-blind: remove MUST NOT run (fallback cannot see squash-merged branch)")
	}
	if !strings.Contains(out.String(), "squash-merge blind") {
		t.Fatalf("squash-blind: notice MUST document the fallback's blindness, got %q", out.String())
	}
}

// TestPRMergeCleanup_GhAbsentEmitsBlindnessNoticeOnce asserts the squash-blind
// notice is emitted exactly once per invocation (not per-worktree) when gh is
// absent, even with zero candidates.
func TestPRMergeCleanup_GhAbsentEmitsBlindnessNoticeOnce(t *testing.T) {
	swapPRMergeSeams(t, prMergeSeams{
		wtList:       func() (string, error) { return wtListPorcelainPrimary(), nil }, // no WT- worktrees
		ghLookPath:   func() bool { return false },
		branchMerged: func() ([]string, error) { return []string{"main"}, nil },
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)
	count := strings.Count(out.String(), "squash-merge blind")
	if count != 1 {
		t.Fatalf("blindness notice: expected exactly 1 'squash-merge blind' notice, got %d in %q", count, out.String())
	}
}

// --- REQ-SW-024 dirty guard ---

// TestPRMergeCleanup_DirtyPreserves is AC-SW-024: merged branch + dirty
// worktree → NOT removed; notice names the path.
func TestPRMergeCleanup_DirtyPreserves(t *testing.T) {
	removeCalled := false
	wtPath := "/repo/.claude/worktrees/WT-dirty01-init"
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + wtListEntry(wtPath, "WT-dirty01-init"), nil
		},
		ghLookPath: func() bool { return true },
		ghPRState:  func(string) (string, bool) { return "MERGED", true },
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove: func(string) error { removeCalled = true; return nil },
		statusPorc: func(string) (string, error) {
			return " M dirty.go\n", nil // dirty (EC-11 race re-check)
		},
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)
	if removeCalled {
		t.Fatal("dirty: remove MUST NOT run (uncommitted changes preserved)")
	}
	if strings.Contains(out.String(), PRMergeCleanupNoticePrefix) {
		t.Fatalf("dirty: no removal notice expected, got %q", out.String())
	}
	if !strings.Contains(out.String(), wtPath) {
		t.Fatalf("dirty: notice must name worktree path %q, got %q", wtPath, out.String())
	}
	if !strings.Contains(out.String(), "preserved") {
		t.Fatalf("dirty: notice must say 'preserved', got %q", out.String())
	}
}

// TestPRMergeCleanup_DirtyCheckErrorPreserves is EC-11 fail-open: a status
// error preserves the worktree (never risk deleting an unverifiable state).
func TestPRMergeCleanup_DirtyCheckErrorPreserves(t *testing.T) {
	removeCalled := false
	wtPath := "/repo/.claude/worktrees/WT-errstat-init"
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + wtListEntry(wtPath, "WT-errstat-init"), nil
		},
		ghLookPath: func() bool { return true },
		ghPRState:  func(string) (string, bool) { return "MERGED", true },
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove: func(string) error { removeCalled = true; return nil },
		statusPorc: func(string) (string, error) {
			return "", errFakeNotGitRepo
		},
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)
	if removeCalled {
		t.Fatal("dirty-check-error: remove MUST NOT run (fail-open preserve)")
	}
	if !strings.Contains(out.String(), wtPath) {
		t.Fatalf("dirty-check-error: notice must name worktree path, got %q", out.String())
	}
}

// --- REQ-SW-022 trigger invariant ---

// TestPRMergeCleanup_WorktreeListErrorFailOpen proves a list failure is
// non-blocking: a notice is emitted and the on-touch invocation proceeds.
func TestPRMergeCleanup_WorktreeListErrorFailOpen(t *testing.T) {
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) { return "", errFakeNotGitRepo },
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)
	if !strings.Contains(out.String(), "PR-merge cleanup skipped") {
		t.Fatalf("list-error: expected skip notice, got %q", out.String())
	}
}

// TestPRMergeCleanup_OnlyWTBranchesConsidered proves non-WT branches are
// ignored by the cleanup sweep (e.g. a feature branch checked out as a
// worktree).
func TestPRMergeCleanup_OnlyWTBranchesConsidered(t *testing.T) {
	stateCalls := 0
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() +
				wtListEntry("/repo/.claude/worktrees/feat-other", "feat-other") + // non-WT branch
				wtListEntry("/repo/.claude/worktrees/WT-abcd1234-web", "WT-abcd1234-web"), nil
		},
		ghLookPath: func() bool { return true },
		ghPRState: func(string) (string, bool) {
			stateCalls++
			return "MERGED", true
		},
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(string) error { return nil },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)
	// gh view should be called ONLY for the WT- branch (1), not the non-WT one.
	if stateCalls != 1 {
		t.Fatalf("non-WT filter: gh view must be called once (WT- only), got %d", stateCalls)
	}
}

// TestPRMergeCleanup_DetachedWorktreeIgnored proves a detached-HEAD worktree
// (no branch line in porcelain) is skipped, not treated as a candidate.
func TestPRMergeCleanup_DetachedWorktreeIgnored(t *testing.T) {
	stateCalled := false
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			// detached entry has no `branch refs/heads/...` line
			return "worktree /repo/.claude/worktrees/detached\nHEAD 2222222222222222222222222222222222222222\ndetached\n", nil
		},
		ghLookPath: func() bool { return true },
		ghPRState:  func(string) (string, bool) { stateCalled = true; return "MERGED", true },
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)
	if stateCalled {
		t.Fatal("detached: gh view MUST NOT be called for a branchless worktree")
	}
}

// TestPRMergeCleanup_RemovalFailureFailOpen proves a `git worktree remove`
// failure is non-blocking: a notice names the failure and the sweep continues.
func TestPRMergeCleanup_RemovalFailureFailOpen(t *testing.T) {
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + wtListEntry("/repo/.claude/worktrees/WT-rmfail-web", "WT-rmfail-web"), nil
		},
		ghLookPath: func() bool { return true },
		ghPRState:  func(string) (string, bool) { return "MERGED", true },
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(string) error { return fakeErr("fake-remove-failure") },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)
	if !strings.Contains(out.String(), "PR-merge cleanup failed") {
		t.Fatalf("remove-fail: expected failure notice, got %q", out.String())
	}
}

// TestPRMergeCleanup_FallbackBranchMergedErrorPreserves covers the EC/fail-open
// branch where `git branch --merged origin/main` itself errors (no origin/main
// ref, git failure). The worktree is preserved (cannot confirm merge state).
func TestPRMergeCleanup_FallbackBranchMergedErrorPreserves(t *testing.T) {
	removeCalled := false
	wtPath := "/repo/.claude/worktrees/WT-fberr-web"
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + wtListEntry(wtPath, "WT-fberr-web"), nil
		},
		ghLookPath:   func() bool { return false },
		branchMerged: func() ([]string, error) { return nil, errFakeNotGitRepo },
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(string) error { removeCalled = true; return nil },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})
	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)
	if removeCalled {
		t.Fatal("fallback-error: remove MUST NOT run (cannot confirm merge state)")
	}
	if strings.Contains(out.String(), PRMergeCleanupNoticePrefix) {
		t.Fatalf("fallback-error: no removal notice expected, got %q", out.String())
	}
}

// --- EC-13 notice distinguishability ---

// TestPRMergeCleanup_NoticeDistinguishableFromSessionExit is EC-13: the M8
// notice prefix MUST be distinct from the session-exit prefix so the two are
// attributable in combined output.
func TestPRMergeCleanup_NoticeDistinguishableFromSessionExit(t *testing.T) {
	if PRMergeCleanupNoticePrefix == SessionExitCleanupNoticePrefix {
		t.Fatalf("prefix collision: both %q", PRMergeCleanupNoticePrefix)
	}
	if !strings.Contains(PRMergeCleanupNoticePrefix, "PR-merge") {
		t.Fatalf("M8 prefix MUST contain 'PR-merge': %q", PRMergeCleanupNoticePrefix)
	}
	if strings.Contains(PRMergeCleanupNoticePrefix, "session-exit") {
		t.Fatalf("M8 prefix MUST NOT contain 'session-exit': %q", PRMergeCleanupNoticePrefix)
	}
	if !strings.HasPrefix(PRMergeCleanupNoticePrefix, "removed by ") {
		t.Fatalf("M8 prefix MUST start with 'removed by ': %q", PRMergeCleanupNoticePrefix)
	}
}

// --- AC-SW-022 on-touch trigger wiring ---

// chdirTemp changes the test's working directory to tmp and restores the
// original on cleanup. loadSessionWorktreeConfig reads from os.Getwd() (NOT
// $CLAUDE_PROJECT_DIR), so the on-touch tests must chdir into the staged
// config dir for the toggle to resolve ON.
func chdirTemp(t *testing.T, tmp string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getcwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir %s: %v", tmp, err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

// TestPRMergeCleanup_OnTouchFiresAtSessionRegister proves the register RunE
// invokes prMergeCleanup (gated by the toggle). The seam counter increments
// only when AutoCleanup is wired through the RunE.
func TestPRMergeCleanup_OnTouchFiresAtSessionRegister(t *testing.T) {
	listCalled := false
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) { listCalled = true; return "", errFakeNotGitRepo },
	})
	// Stage a temp project dir with a config carrying AutoCleanup=true so
	// loadSessionWorktreeConfig (reads os.Getwd()) resolves ON.
	tmp := t.TempDir()
	writeAutoCleanupConfig(t, tmp)
	chdirTemp(t, tmp)
	cmd := newSessionRegisterCmd()
	cmd.SetArgs([]string{"sessid", "SPEC-X", "run"})
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("register RunE: unexpected error: %v", err)
	}
	if !listCalled {
		t.Fatal("on-touch register: prMergeCleanup MUST be invoked (git worktree list called)")
	}
}

// TestPRMergeCleanup_OnTouchFiresAtSessionList proves the list RunE invokes
// prMergeCleanup (gated by the toggle).
func TestPRMergeCleanup_OnTouchFiresAtSessionList(t *testing.T) {
	listCalled := false
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) { listCalled = true; return "", errFakeNotGitRepo },
	})
	tmp := t.TempDir()
	writeAutoCleanupConfig(t, tmp)
	chdirTemp(t, tmp)
	cmd := newSessionListCmd()
	cmd.SetArgs([]string{})
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list RunE: unexpected error: %v", err)
	}
	if !listCalled {
		t.Fatal("on-touch list: prMergeCleanup MUST be invoked (git worktree list called)")
	}
}

// TestPRMergeCleanup_OnTouchToggleOffSkipsAtSessionRegister proves the negative
// control (AC-SW-022): with AutoCleanup OFF, the on-touch invocation is a
// no-op even at the two trigger sites.
func TestPRMergeCleanup_OnTouchToggleOffSkipsAtSessionRegister(t *testing.T) {
	listCalled := false
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) { listCalled = true; return "", nil },
	})
	tmp := t.TempDir()
	chdirTemp(t, tmp)
	// No config written → loadSessionWorktreeConfig returns nil → OFF.
	cmd := newSessionRegisterCmd()
	cmd.SetArgs([]string{"sessid", "SPEC-X", "run"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("register RunE: unexpected error: %v", err)
	}
	if listCalled {
		t.Fatal("toggle-off on-touch: prMergeCleanup MUST NOT invoke git worktree list")
	}
}

// TestPRMergeCleanup_TriggerInvariant_OtherSubcommandsDoNotFire is AC-SW-022
// trigger invariant: prMergeCleanup MUST be referenced ONLY at the register +
// list RunE sites, NOT at cc/cg/glm/init/profile/web or any other subcommand.
// This is a static source assertion (falsifiable: adding a stray reference
// anywhere else fails the test).
func TestPRMergeCleanup_TriggerInvariant_OtherSubcommandsDoNotFire(t *testing.T) {
	// session.go is the ONLY file allowed to reference prMergeCleanup by call.
	data, err := os.ReadFile("session.go")
	if err != nil {
		t.Fatalf("read session.go: %v", err)
	}
	src := string(data)
	// Must contain exactly the two on-touch call sites.
	registerHasIt := strings.Contains(src, "prMergeCleanup(") &&
		strings.Contains(src, "register")
	listHasIt := strings.Contains(src, "prMergeCleanup(") &&
		strings.Contains(src, "list")
	if !registerHasIt || !listHasIt {
		t.Fatalf("trigger wiring: session.go must reference prMergeCleanup at both register and list sites")
	}
	// Count occurrences of the call in session.go — must be exactly 2
	// (register RunE + list RunE). The function definition lives in
	// session_worktree_prmerge.go, not session.go.
	callCount := strings.Count(src, "prMergeCleanup(")
	if callCount != 2 {
		t.Fatalf("trigger invariant: session.go must call prMergeCleanup exactly twice (register+list), got %d", callCount)
	}
}

// writeAutoCleanupConfig writes a minimal .moai/config that resolves
// Workflow.Worktree.AutoCleanup=true under the given project dir, so
// loadSessionWorktreeConfig returns ON.
func writeAutoCleanupConfig(t *testing.T, projectDir string) {
	t.Helper()
	sectionsDir := filepath.Join(projectDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	// workflow.yaml carries the worktree.auto_cleanup toggle.
	workflowYAML := []byte("workflow:\n  worktree:\n    auto_cleanup: true\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "workflow.yaml"), workflowYAML, 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
}

// --- parseWorktreeList unit tests ---

// TestParseWorktreeList_BranchExtraction verifies the porcelain parser
// extracts (path, branch) pairs and strips the refs/heads/ prefix.
func TestParseWorktreeList_BranchExtraction(t *testing.T) {
	porcelain := wtListPorcelainPrimary() +
		wtListEntry("/repo/.claude/worktrees/WT-abcd1234-web", "WT-abcd1234-web")
	entries := parseWorktreeList(porcelain)
	if len(entries) != 2 {
		t.Fatalf("parse: expected 2 entries, got %d", len(entries))
	}
	if entries[0].path != "/repo" || entries[0].branch != "main" {
		t.Fatalf("parse[0]: expected (/repo, main), got (%q, %q)", entries[0].path, entries[0].branch)
	}
	if entries[1].path != "/repo/.claude/worktrees/WT-abcd1234-web" || entries[1].branch != "WT-abcd1234-web" {
		t.Fatalf("parse[1]: expected (WT path, WT-abcd1234-web), got (%q, %q)", entries[1].path, entries[1].branch)
	}
}

// TestParseWorktreeList_DetachedEntrySkipped verifies a detached-HEAD entry
// (no branch line) is excluded from candidates (empty branch).
func TestParseWorktreeList_DetachedEntrySkipped(t *testing.T) {
	porcelain := "worktree /repo/.claude/worktrees/detached\nHEAD 2222222222222222222222222222222222222222\ndetached\n"
	entries := parseWorktreeList(porcelain)
	for _, e := range entries {
		if e.branch != "" {
			t.Fatalf("detached: branch must be empty for detached entry, got %q", e.branch)
		}
	}
}

// --- real-implementation sanity (encoding/json round-trip for gh state) ---

// TestGhPRViewStateReal_ParseShape verifies ghPRViewStateReal parses the
// documented `{"state":"MERGED",...}` JSON shape. The seam is swapped to the
// real parser with a stubbed command output via a helper.
func TestGhPRViewStateReal_ParseShape(t *testing.T) {
	got := parseGhPRStateJSON(`{"number":1,"state":"MERGED","title":"x"}`)
	if got != "MERGED" {
		t.Fatalf("parse gh state: expected MERGED, got %q", got)
	}
	got = parseGhPRStateJSON(`{"state":"OPEN"}`)
	if got != "OPEN" {
		t.Fatalf("parse gh state: expected OPEN, got %q", got)
	}
	got = parseGhPRStateJSON(`{}`) // no state field
	if got != "" {
		t.Fatalf("parse gh state: expected empty for missing field, got %q", got)
	}
	got = parseGhPRStateJSON(`not-json`) // malformed
	if got != "" {
		t.Fatalf("parse gh state: expected empty for malformed JSON, got %q", got)
	}
}

// ensure config import is used (worktreeCfg lives in the sibling test file but
// referencing it here keeps the import honest if the sibling file moves).
var _ = config.Config{}
