package cli

// session_worktree_prmerge_merge_test.go — SPEC-WORKTREE-REAPER-001 M1:
// merge detection gains a third outcome, and the git fallback becomes
// reachable while gh is installed.
//
// AC-WR-002, AC-WR-003, AC-WR-004, AC-WR-007.
//
// The defect these pin: the seam returned a bare string, so gh-errored /
// no-PR / malformed-JSON all arrived at the caller as "", indistinguishable
// from a determinate negative. The git fallback was therefore dead code
// whenever `gh` was on PATH — which is always, here.

import (
	"bytes"
	"strings"
	"testing"
)

// ghNoAnswer is the seam stub for "gh produced nothing usable".
func ghNoAnswer(string) (string, bool) { return "", false }

// TestPRMergeCleanup_GhNoAnswerConsultsGitFallback is AC-WR-002: a gh
// no-answer routes the decision to `git branch --merged`, which removes the
// tree. Before M1 this path was unreachable — "" ended the decision.
func TestPRMergeCleanup_GhNoAnswerConsultsGitFallback(t *testing.T) {
	branch := "WT-forge-counts"
	tree := "/repo/.claude/worktrees/" + branch

	var removedPath string
	fallbackCalls := 0
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + "\n" + wtListEntry(tree, branch), nil
		},
		ghLookPath: func() bool { return true },
		ghPRState:  ghNoAnswer,
		branchMerged: func() ([]string, error) {
			fallbackCalls++
			return []string{"main", branch}, nil
		},
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(p string) error { removedPath = p; return nil },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})

	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)

	if fallbackCalls == 0 {
		t.Fatal("gh no-answer must consult the git fallback; it was never called")
	}
	if removedPath != tree {
		t.Fatalf("expected remove(%q), got %q — output: %s", tree, removedPath, out.String())
	}
}

// TestPRMergeCleanup_GhOpenSkipsGitFallback is AC-WR-003 / REQ-WR-004: a
// DETERMINATE gh negative ends the decision. git is squash-blind in this
// fallback, so consulting it after a determinate answer could only contradict
// a better source — git fills a hole, it never overrides.
func TestPRMergeCleanup_GhOpenSkipsGitFallback(t *testing.T) {
	branch := "WT-open-pr"
	tree := "/repo/.claude/worktrees/" + branch

	removeCalled := false
	fallbackCalls := 0
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + "\n" + wtListEntry(tree, branch), nil
		},
		ghLookPath: func() bool { return true },
		ghPRState:  func(string) (string, bool) { return "OPEN", true },
		branchMerged: func() ([]string, error) {
			fallbackCalls++
			return []string{"main", branch}, nil // would say "merged" if consulted
		},
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(string) error { removeCalled = true; return nil },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})

	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)

	if fallbackCalls != 0 {
		t.Fatalf("a determinate gh negative must NOT consult the git fallback; it was called %d time(s)", fallbackCalls)
	}
	if removeCalled {
		t.Fatalf("an OPEN PR's worktree must be preserved — output: %s", out.String())
	}
}

// TestPRMergeCleanup_UndeterminedMergePreserves is AC-WR-004 / REQ-WR-003:
// when NEITHER source answers, the sweep preserves and says so. Absence of a
// merge signal is not evidence either way, and the notice carries the
// cause-specific token (REQ-WR-023) so this tree is distinguishable from one
// preserved for any other reason.
func TestPRMergeCleanup_UndeterminedMergePreserves(t *testing.T) {
	branch := "WT-undetermined"
	tree := "/repo/.claude/worktrees/" + branch

	removeCalled := false
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + "\n" + wtListEntry(tree, branch), nil
		},
		ghLookPath:   func() bool { return true },
		ghPRState:    ghNoAnswer,
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
		t.Fatal("an undetermined merge state must never remove a worktree")
	}
	if !strings.Contains(out.String(), tree) {
		t.Fatalf("the preserve notice must name the worktree path %q, got %q", tree, out.String())
	}
	if !strings.Contains(out.String(), "undetermined-merge") {
		t.Fatalf("the preserve notice must name the undetermined merge state, got %q", out.String())
	}
}

// TestPRMergeCleanup_ZeroUniqueCommitPreserved is AC-WR-007 / REQ-WR-018: it
// pins the removal class M1 NEWLY REACHES and the guard that bounds it.
//
// `git branch --merged <base>` lists every branch whose tip is an ancestor of
// the base — which includes a branch with zero commits of its own. Such a
// branch holds no committed work the base lacks, so removing its worktree
// destroys no commit; the boundary is the DIRTY GUARD, and it covers untracked
// files as well as tracked ones. This SPEC's own branch is a live member of
// the class.
//
// Variant (b) additionally asserts the dirty cause token — the one of the five
// cause tokens no other criterion exercises.
func TestPRMergeCleanup_ZeroUniqueCommitPreserved(t *testing.T) {
	t.Run("clean tree is removed", func(t *testing.T) {
		branch := "WT-zero-unique-clean"
		tree := "/repo/.claude/worktrees/" + branch

		var removedPath string
		swapPRMergeSeams(t, prMergeSeams{
			wtList: func() (string, error) {
				return wtListPorcelainPrimary() + "\n" + wtListEntry(tree, branch), nil
			},
			ghLookPath:   func() bool { return true },
			ghPRState:    ghNoAnswer, // no PR was ever opened for this branch
			branchMerged: func() ([]string, error) { return []string{"main", branch}, nil },
		})
		swapSessionWorktreeSeams(t, swSeams{
			remove:      func(p string) error { removedPath = p; return nil },
			statusPorc:  func(string) (string, error) { return "", nil },
			ignoredPorc: func(string) (string, error) { return "", nil },
		})

		var out bytes.Buffer
		prMergeCleanup(worktreeCfg(true), &out)

		if removedPath != tree {
			t.Fatalf("a zero-unique-commit branch on a clean tree is removable; got %q — output: %s", removedPath, out.String())
		}
	})

	t.Run("untracked file preserves", func(t *testing.T) {
		branch := "WT-zero-unique-dirty"
		tree := "/repo/.claude/worktrees/" + branch

		removeCalled := false
		swapPRMergeSeams(t, prMergeSeams{
			wtList: func() (string, error) {
				return wtListPorcelainPrimary() + "\n" + wtListEntry(tree, branch), nil
			},
			ghLookPath:   func() bool { return true },
			ghPRState:    ghNoAnswer,
			branchMerged: func() ([]string, error) { return []string{"main", branch}, nil },
		})
		swapSessionWorktreeSeams(t, swSeams{
			remove: func(string) error { removeCalled = true; return nil },
			// `??` is exactly the class git itself refuses to discard without
			// --force, and the class the dirty guard exists for.
			statusPorc:  func(string) (string, error) { return "?? scratch.txt\n", nil },
			ignoredPorc: func(string) (string, error) { return "", nil },
		})

		var out bytes.Buffer
		prMergeCleanup(worktreeCfg(true), &out)

		if removeCalled {
			t.Fatal("the dirty guard must bound the zero-unique-commit class: an untracked file preserves")
		}
		if !strings.Contains(out.String(), "cause=dirty") {
			t.Fatalf("the preserve notice must carry the dirty cause token, got %q", out.String())
		}
	})
}
