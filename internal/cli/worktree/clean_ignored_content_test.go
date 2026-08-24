package worktree

// clean_ignored_content_test.go — SPEC-WORKTREE-REAPER-001 post-audit repair
// F3: `clean --stale` gains the ignored-content guard REQ-WR-024 built for the
// PR-merge sweep, and the two sweeps share ONE decision.
//
// Why the guard is the only thing standing there: `git status --porcelain` and
// non-forced `git worktree remove` AGREE in disregarding ignored files
// (design.md §A.6, measured), so a merged, clean, unanchored tree holding only
// `.claude/agent-memory/` reports dirty=no and is destroyed by --yes, exit 0.
// REQ-WR-019 shared the ANCHOR decision across the three sweeps; this shares
// the IGNORED-CONTENT one, rather than letting a second copy of the allowlist
// drift away from the first.

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// errIgnoredStatusUnreadableFixture stands in for a `git status --porcelain
// --ignored` that could not be run at all.
var errIgnoredStatusUnreadableFixture = errors.New("fatal: could not read the working tree")

// ignoredStaleEnv installs a clean, merged, unanchored worktree whose
// `git status --porcelain --ignored` reports the given ignored entries. The
// tracked status stays clean, so the dirty guard cannot be what preserves it.
func ignoredStaleEnv(t *testing.T, tree, branch, ignoredPorcelain string) *[]string {
	t.Helper()
	removed := staleTestEnv(t,
		[]git.Worktree{{Path: tree, Branch: branch}},
		map[string]string{}, // clean tracked status
		map[string]bool{branch: true},
	)

	origGitCmd := gitWorktreeCmd
	gitWorktreeCmd = func(args ...string) (string, error) {
		for _, a := range args {
			if a == "--ignored" {
				return ignoredPorcelain, nil
			}
		}
		return origGitCmd(args...)
	}
	t.Cleanup(func() { gitWorktreeCmd = origGitCmd })

	return removed
}

// TestCleanStale_KeepsIrreplaceableIgnoredContent is the load-bearing case:
// agent memory has no regenerator, so its loss is permanent and the tree is
// kept even with --yes.
func TestCleanStale_KeepsIrreplaceableIgnoredContent(t *testing.T) {
	removed := ignoredStaleEnv(t, "/wt/memory", "feat/memory",
		"!! .claude/agent-memory/\n!! .moai/state/\n")

	out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("a tree holding irreplaceable ignored content was removed: %v", *removed)
	}
	if !strings.Contains(out, "cause=ignored-content") {
		t.Errorf("expected a keep reason naming the cause, got:\n%s", out)
	}
	if !strings.Contains(out, ".claude/agent-memory") {
		t.Errorf("expected the keep reason to name the irreplaceable entry, got:\n%s", out)
	}
}

// TestCleanStale_RemovesWhenIgnoredContentIsRegenerable pins the other
// direction: the guard must not make every tree immortal. 156 of 156 worktrees
// in this repository were measured to hold ignored content, so a blunt "holds
// ignored content" predicate would have zero discriminating power (REQ-WR-024,
// policy P2).
func TestCleanStale_RemovesWhenIgnoredContentIsRegenerable(t *testing.T) {
	removed := ignoredStaleEnv(t, "/wt/regenerable", "feat/regenerable",
		"!! .moai/state/\n!! bin/\n!! .claude/settings.local.json\n")

	if _, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"}); err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 1 || (*removed)[0] != "/wt/regenerable" {
		t.Fatalf("a tree holding only regenerable ignored content must be removed, got %v", *removed)
	}
}

// TestCleanStale_UnreadableIgnoredStatusPreserves keeps the guard fail-closed:
// an unreadable ignored status is an undetermined answer, and undetermined
// preserves (REQ-WR-017).
func TestCleanStale_UnreadableIgnoredStatusPreserves(t *testing.T) {
	removed := staleTestEnv(t,
		[]git.Worktree{{Path: "/wt/unreadable", Branch: "feat/unreadable"}},
		map[string]string{},
		map[string]bool{"feat/unreadable": true},
	)
	origGitCmd := gitWorktreeCmd
	gitWorktreeCmd = func(args ...string) (string, error) {
		for _, a := range args {
			if a == "--ignored" {
				return "", errIgnoredStatusUnreadableFixture
			}
		}
		return origGitCmd(args...)
	}
	t.Cleanup(func() { gitWorktreeCmd = origGitCmd })

	out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("an unreadable ignored status must preserve; removed %v", *removed)
	}
	if !strings.Contains(out, "cause=ignored-check-failed") {
		t.Errorf("expected a keep reason naming the check failure, got:\n%s", out)
	}
}

// TestCleanStale_JSONReportsIgnoredPredicate is REQ-WR-012: the inventory and
// the sweep are the SAME evaluation, so the fourth predicate appears in the
// record rather than only in the human output.
func TestCleanStale_JSONReportsIgnoredPredicate(t *testing.T) {
	ignoredStaleEnv(t, "/wt/memory-json", "feat/memory-json", "!! .claude/agent-memory/\n")

	out, err := runStaleClean(t, map[string]string{"stale": "true", "json": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	var got []staleCandidate
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("stdout must be valid JSON: %v\n%s", err, out)
	}
	if len(got) != 1 {
		t.Fatalf("expected one record, got %d:\n%s", len(got), out)
	}
	if got[0].Ignored != staleStateYes {
		t.Errorf("a tree holding irreplaceable ignored content must report ignored=%q, got %q", staleStateYes, got[0].Ignored)
	}
	if got[0].KeepReason == "" {
		t.Error("the record must carry a keep_reason")
	}
}
