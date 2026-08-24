package cli

// session_worktree_prmerge_lock_test.go — SPEC-WORKTREE-REAPER-001 M2:
// the lock-aware anchor guard, the refusal-class pre-detection and its
// fall-through, and the P2 ignored-content guard, as seen from the sweep.
//
// AC-WR-008, AC-WR-014, AC-WR-016, AC-WR-024, AC-WR-024b, AC-WR-025b.

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// deadPID is above every platform's pid_max, so `kill(pid, 0)` reports ESRCH
// — a POSITIVELY dead process rather than an undetermined probe. The unit-level
// semantics of "confirmed dead" are pinned in internal/session with a swapped
// probe; here it only has to not be alive.
const deadPID = 2147483646

// wtListLockedEntry builds a porcelain entry carrying a `locked <reason>` line.
func wtListLockedEntry(path, branch, reason string) string {
	return wtListEntry(path, branch) + "locked " + reason + "\n"
}

// TestParseWorktreeList_CapturesLockReason is AC-WR-008: git's own `locked `
// prefix is stripped, and a bare `locked` line is a lock with an empty reason.
func TestParseWorktreeList_CapturesLockReason(t *testing.T) {
	reason := "claude session t207 (pid 36912 start Sun Aug 23 07:26:09 2026)"
	porcelain := wtListPorcelainPrimary() + "\n" +
		wtListLockedEntry("/repo/.claude/worktrees/t207", "WT-web-live-todo", reason) + "\n" +
		wtListEntry("/repo/.claude/worktrees/bare", "WT-bare") + "locked\n"

	entries := parseWorktreeList(porcelain)

	byPath := map[string]wtEntry{}
	for _, e := range entries {
		byPath[e.path] = e
	}
	got := byPath["/repo/.claude/worktrees/t207"]
	if !got.lock.Locked || got.lock.Reason != reason {
		t.Fatalf("stored reason: got %+v, want Locked=true Reason=%q", got.lock, reason)
	}
	bare := byPath["/repo/.claude/worktrees/bare"]
	if !bare.lock.Locked || bare.lock.Reason != "" {
		t.Fatalf("bare locked line: got %+v, want Locked=true Reason=\"\"", bare.lock)
	}
	if primary := byPath["/repo"]; primary.lock.Locked {
		t.Fatal("an unlocked worktree must carry Locked=false (the lock source has no opinion)")
	}
}

// TestPRMergeCleanup_T207SamplePreservedByLock is AC-WR-014, the mandatory
// sample criterion. The registry is EMPTY — t207 is the one tree it can see,
// so letting it answer would leave the guard 4-of-5 blind — and the lock alone
// must preserve the tree, naming itself as the source.
func TestPRMergeCleanup_T207SamplePreservedByLock(t *testing.T) {
	tree := t.TempDir() // no registry written: the registry source sees nothing
	branch := "WT-web-live-todo"
	reason := "claude session t207 (pid " + itoaTest(os.Getpid()) + " start Sun Aug 23 07:26:09 2026)"

	removeCalled := false
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + "\n" + wtListLockedEntry(tree, branch, reason), nil
		},
		ghLookPath: func() bool { return true },
		// gh gives NO answer and git does not list the branch as merged, so the
		// merge decision comes from neither — but the lock guard is what this
		// criterion pins, so the branch is made a candidate via the git fallback.
		ghPRState:    func(string) (string, bool) { return "MERGED", true },
		branchMerged: func() ([]string, error) { return []string{"main"}, nil },
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(string) error { removeCalled = true; return nil },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})

	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)

	if removeCalled {
		t.Fatal("t207 sample: a locked tree with a live pid must never be removed")
	}
	if !strings.Contains(out.String(), tree) {
		t.Fatalf("preserve notice must name the worktree path %q, got %q", tree, out.String())
	}
	if !strings.Contains(out.String(), "lock") {
		t.Fatalf("preserve notice must name the LOCK as the anchor source, got %q", out.String())
	}
}

// TestPRMergeCleanup_PorcelainFailureRemovesNothing is AC-WR-016: an
// unreadable worktree listing removes nothing and says so.
func TestPRMergeCleanup_PorcelainFailureRemovesNothing(t *testing.T) {
	removeCalled := false
	swapPRMergeSeams(t, prMergeSeams{
		wtList:     func() (string, error) { return "", errFakeNotGitRepo },
		ghLookPath: func() bool { return true },
		ghPRState:  func(string) (string, bool) { return "MERGED", true },
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(string) error { removeCalled = true; return nil },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})

	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)

	if removeCalled {
		t.Fatal("porcelain failure: nothing may be removed on an unreadable listing")
	}
	if !strings.Contains(out.String(), "PR-merge cleanup skipped") {
		t.Fatalf("expected a skip notice, got %q", out.String())
	}
}

// TestPRMergeCleanup_RefusalClassNamesCause is AC-WR-024: a locked tree is
// pre-detected as refusal-class and never reaches `git worktree remove`, and
// its preserve notice carries the cause-specific token `locked`.
//
// The assertion is a POSITIVE token match, not string inequality: two distinct
// literals can differ while naming nothing (REQ-WR-023).
//
// Platform note: on unix the dead pid routes through the refusal-class branch;
// on Windows the liveness probe cannot assert death, so the same tree is
// preserved one branch earlier, as anchored-by-lock. Both preserve, and both
// notices name `locked` — which is what the criterion asserts.
func TestPRMergeCleanup_RefusalClassNamesCause(t *testing.T) {
	tree := t.TempDir()
	branch := "WT-dead-lock"
	reason := "claude session t212 (pid " + itoaTest(deadPID) + " start Sun Aug 23 07:26:09 2026)"

	removeCalled := false
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + "\n" + wtListLockedEntry(tree, branch, reason), nil
		},
		ghLookPath: func() bool { return true },
		ghPRState:  func(string) (string, bool) { return "MERGED", true },
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(string) error { removeCalled = true; return nil },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})

	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)

	if removeCalled {
		t.Fatal("refusal class: a locked tree must never reach `git worktree remove`")
	}
	if !strings.Contains(out.String(), "locked") {
		t.Fatalf("preserve notice must carry the cause token %q, got %q", "locked", out.String())
	}
	if strings.Contains(out.String(), "PR-merge cleanup failed") {
		t.Fatalf("a pre-detected refusal must not produce a failure-shaped notice, got %q", out.String())
	}
}

// TestPRMergeCleanup_RefusalFallThroughNamesCauseAndContinues is AC-WR-024b —
// REQ-WR-021's DEFINING limb. The pre-detection set is deliberately
// non-exhaustive (a populated submodule refuses with a clean porcelain,
// EC-12), so every unlisted member takes this path: git refuses, the sweep
// names the cause from git's own message, and the sweep CONTINUES.
func TestPRMergeCleanup_RefusalFallThroughNamesCauseAndContinues(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	const gitRefusal = "working trees containing submodules cannot be moved or removed"

	var removed []string
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + "\n" +
				wtListEntry(first, "WT-submodule") + "\n" +
				wtListEntry(second, "WT-plain"), nil
		},
		ghLookPath: func() bool { return true },
		ghPRState:  func(string) (string, bool) { return "MERGED", true },
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove: func(p string) error {
			if p == first {
				return fakeErr(gitRefusal)
			}
			removed = append(removed, p)
			return nil
		},
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(string) (string, error) { return "", nil },
	})

	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)

	if !strings.Contains(out.String(), gitRefusal) {
		t.Fatalf("the fall-through notice must carry git's own refusal cause, got %q", out.String())
	}
	if !strings.Contains(out.String(), first) {
		t.Fatalf("the fall-through notice must name the preserved worktree %q, got %q", first, out.String())
	}
	if len(removed) != 1 || removed[0] != second {
		t.Fatalf("the sweep must continue past a refusal and process the next candidate; removed = %v", removed)
	}
}

// TestPRMergeCleanup_IgnoredContentPolicyP2 is AC-WR-025b. Three candidates:
// (a) regenerable ignored content only → removed; (b) `.claude/agent-memory/`
// → preserved; (c) an ignored path in NEITHER category → preserved, because
// unclassifiable means irreplaceable (fail-closed).
//
// It also asserts the in-code stopgap record required by REQ-WR-025: the
// preserve branch must carry an @MX:NOTE naming drain-then-dispose, so the
// next implementer reads it where the decision lives rather than only in the
// SPEC.
func TestPRMergeCleanup_IgnoredContentPolicyP2(t *testing.T) {
	regenerable := t.TempDir()
	agentMemory := t.TempDir()
	unclassifiable := t.TempDir()

	ignored := map[string]string{
		regenerable:    "!! .moai/state/\n!! .claude/settings.local.json\n!! bin/\n",
		agentMemory:    "!! .moai/state/\n!! .claude/agent-memory/\n",
		unclassifiable: "!! vendor/some-unclassified-cache/\n",
	}

	var removed []string
	swapPRMergeSeams(t, prMergeSeams{
		wtList: func() (string, error) {
			return wtListPorcelainPrimary() + "\n" +
				wtListEntry(regenerable, "WT-regenerable") + "\n" +
				wtListEntry(agentMemory, "WT-agent-memory") + "\n" +
				wtListEntry(unclassifiable, "WT-unclassified"), nil
		},
		ghLookPath: func() bool { return true },
		ghPRState:  func(string) (string, bool) { return "MERGED", true },
	})
	swapSessionWorktreeSeams(t, swSeams{
		remove:      func(p string) error { removed = append(removed, p); return nil },
		statusPorc:  func(string) (string, error) { return "", nil },
		ignoredPorc: func(p string) (string, error) { return ignored[p], nil },
	})

	var out bytes.Buffer
	prMergeCleanup(worktreeCfg(true), &out)

	if len(removed) != 1 || removed[0] != regenerable {
		t.Fatalf("only the all-regenerable tree may be removed; removed = %v", removed)
	}
	for _, preserved := range []string{agentMemory, unclassifiable} {
		if !strings.Contains(out.String(), preserved) {
			t.Fatalf("preserve notice must name %q, got %q", preserved, out.String())
		}
	}
	if !strings.Contains(out.String(), "ignored") {
		t.Fatalf("preserve notice must name ignored content as the cause, got %q", out.String())
	}

	// REQ-WR-025: the stopgap is recorded where the decision lives.
	src, err := os.ReadFile("session_worktree_prmerge.go")
	if err != nil {
		t.Fatalf("read implementation source: %v", err)
	}
	if !strings.Contains(string(src), "@MX:NOTE") || !strings.Contains(string(src), "drain-then-dispose") {
		t.Fatal("the ignored-content preserve branch must carry an @MX:NOTE naming drain-then-dispose (REQ-WR-025)")
	}
}

// itoaTest renders a pid without pulling strconv into the fixture helpers.
func itoaTest(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
