package worktree

// clean_lock_unreadable_test.go — SPEC-WORKTREE-REAPER-001 post-audit repair
// F2: an unreadable `git worktree list --porcelain` must not silently drop the
// authoritative anchor source.
//
// AC-WR-016 ("a porcelain parse failure removes nothing") was written for
// prMergeCleanup, which aborts its sweep and preserves everything. Both `clean`
// sweeps took the opposite branch: worktreeLockStates returned nil, every tree
// fell back to the registry alone — the source measured to name 1 of 5 live
// anchors (REQ-WR-010) — and no notice told the operator the run was degraded.
// With --yes that removes a lock-anchored tree and kills a live session's
// shell, which is the exact harm M2 exists to prevent.

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// unreadableLockEnv installs the stale fixtures for a clean, merged worktree
// and then makes `git worktree list --porcelain` fail. The tree carries no
// registry entry, so the lock is the only source that could have spoken.
func unreadableLockEnv(t *testing.T, tree, branch string) *[]string {
	t.Helper()
	removed := staleTestEnv(t,
		[]git.Worktree{{Path: tree, Branch: branch}},
		map[string]string{}, // clean working tree
		map[string]bool{branch: true},
	)

	origGitCmd := gitWorktreeCmd
	gitWorktreeCmd = func(args ...string) (string, error) {
		if len(args) >= 2 && args[0] == "worktree" && args[1] == "list" {
			return "", errors.New("fatal: not a git repository")
		}
		return origGitCmd(args...)
	}
	t.Cleanup(func() { gitWorktreeCmd = origGitCmd })

	return removed
}

// TestCleanStale_UnreadableLockSourceRemovesNothing is the --stale limb of the
// fail-closed contract: the sweep removes nothing and says why.
func TestCleanStale_UnreadableLockSourceRemovesNothing(t *testing.T) {
	removed := unreadableLockEnv(t, "/wt/lock-unreadable", "feat/lock-unreadable")

	out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runClean must stay non-blocking (REQ-WR-016), got error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("an unreadable lock source must remove nothing; removed %v", *removed)
	}
	if !strings.Contains(out, "cause=lock-source-unreadable") {
		t.Errorf("expected a notice naming the cause, got:\n%s", out)
	}
}

// TestCleanMergedOnly_UnreadableLockSourceRemovesNothing is the --merged-only
// limb — the most exposed of the three consumers, since that path has no dirty
// guard and the anchor decision is its sole protection.
func TestCleanMergedOnly_UnreadableLockSourceRemovesNothing(t *testing.T) {
	removed := unreadableLockEnv(t, "/wt/merged-lock-unreadable", "feat/merged-lock-unreadable")

	out, err := runStaleClean(t, map[string]string{"merged-only": "true"})
	if err != nil {
		t.Fatalf("runClean must stay non-blocking (REQ-WR-016), got error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("an unreadable lock source must remove nothing; removed %v", *removed)
	}
	if !strings.Contains(out, "cause=lock-source-unreadable") {
		t.Errorf("expected a notice naming the cause, got:\n%s", out)
	}
}

// TestCleanStale_JSONReportsUndeterminedAnchor keeps the inventory and the
// sweep the same evaluation (REQ-WR-012): a degraded run reports the anchor as
// undetermined rather than as a negative, and carries the keep reason.
func TestCleanStale_JSONReportsUndeterminedAnchor(t *testing.T) {
	removed := unreadableLockEnv(t, "/wt/json-lock-unreadable", "feat/json-lock-unreadable")

	out, err := runStaleClean(t, map[string]string{"stale": "true", "json": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("the reporting path must remove nothing; removed %v", *removed)
	}

	var got []staleCandidate
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("stdout must be valid JSON: %v\n%s", err, out)
	}
	if len(got) != 1 {
		t.Fatalf("expected one record, got %d:\n%s", len(got), out)
	}
	if got[0].Anchored != staleStateUndetermined {
		t.Errorf("an unreadable lock source must report anchored=%q, got %q", staleStateUndetermined, got[0].Anchored)
	}
	if !strings.Contains(got[0].KeepReason, "cause=lock-source-unreadable") {
		t.Errorf("keep_reason must name the cause, got %q", got[0].KeepReason)
	}
}
