package worktree

// clean_json_test.go — SPEC-WORKTREE-REAPER-001 M3: `clean --stale` gains a
// machine-readable inventory and a base ref that matches the other sweep.
//
// AC-WR-017, AC-WR-019, AC-WR-020.
//
// M3 EXTENDS the shipped command rather than adding a parallel inventory
// surface (design.md §C, option O3-d): `clean --stale` already enumerates
// every registered worktree, previews by default, and gates removal behind
// --yes. A second command answering the same question would have to be kept
// in agreement with this one forever.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// TestCleanStale_JSONEmitsAllTreesWithStates is AC-WR-017. The fixture mixes a
// legacy `WT-*` tree with two bare-slug ones, because worktree-ness is a
// CHECKOUT property, not a branch-name property: a report that keyed on a
// `WT-*` glob would silently omit every card tree created under the current
// naming convention.
func TestCleanStale_JSONEmitsAllTreesWithStates(t *testing.T) {
	removed := staleTestEnv(t,
		[]git.Worktree{
			{Path: "/wt/legacy", Branch: "WT-legacy-card"},
			{Path: "/wt/slug", Branch: "worktree-reaper"},
			{Path: "/wt/dirty", Branch: "docs-refresh"},
		},
		map[string]string{"/wt/dirty": "?? scratch.txt\n"},
		map[string]bool{"WT-legacy-card": true, "worktree-reaper": false, "docs-refresh": true},
	)

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
	if len(got) != 3 {
		t.Fatalf("expected one object per non-protected tree (3), got %d:\n%s", len(got), out)
	}

	byPath := map[string]staleCandidate{}
	for _, c := range got {
		byPath[c.Path] = c
	}

	// The non-WT entries are present — the gap the inventory exists to close.
	for _, p := range []string{"/wt/slug", "/wt/dirty"} {
		if _, ok := byPath[p]; !ok {
			t.Fatalf("non-WT worktree %q missing from the report:\n%s", p, out)
		}
	}

	removable := byPath["/wt/legacy"]
	if removable.KeepReason != "" {
		t.Errorf("a clean merged tree must report an empty keep_reason, got %q", removable.KeepReason)
	}
	if removable.Dirty != staleStateNo || removable.Merged != staleStateYes || removable.Anchored != staleStateNo {
		t.Errorf("clean merged unanchored tree reported as %+v", removable)
	}

	dirty := byPath["/wt/dirty"]
	if dirty.Dirty != staleStateYes {
		t.Errorf("a tree with an untracked file must report dirty=%q, got %q", staleStateYes, dirty.Dirty)
	}
	if dirty.KeepReason == "" {
		t.Error("a dirty tree must carry a keep_reason")
	}
	// The merge check is short-circuited by the dirty guard, so it was never
	// asked — that is "not-checked", never "no".
	if dirty.Merged != staleStateNotChecked {
		t.Errorf("an unasked predicate must report %q, got %q", staleStateNotChecked, dirty.Merged)
	}

	unmerged := byPath["/wt/slug"]
	if unmerged.Merged != staleStateNo || unmerged.KeepReason == "" {
		t.Errorf("an unmerged tree must report merged=%q with a keep_reason, got %+v", staleStateNo, unmerged)
	}
}

// TestCleanStale_JSONIgnoresYes pins REQ-WR-013 at the flag boundary: the
// inventory is a report, so it must not become a removal on a stray --yes.
func TestCleanStale_JSONIgnoresYes(t *testing.T) {
	removed := staleTestEnv(t,
		[]git.Worktree{{Path: "/wt/removable", Branch: "feat/removable"}},
		map[string]string{},
		map[string]bool{"feat/removable": true},
	)

	if _, err := runStaleClean(t, map[string]string{"stale": "true", "json": "true", "yes": "true"}); err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if len(*removed) != 0 {
		t.Fatalf("--json must remove nothing even with --yes; removed %v", *removed)
	}
}

// TestCleanStale_BaseDefaultsToOriginMain is AC-WR-019 / REQ-WR-022. The old
// default compared against the LOCAL main, which lags the remote — so the two
// sweeps in this repository could reach opposite conclusions about the same
// worktree while both behaved as documented.
func TestCleanStale_BaseDefaultsToOriginMain(t *testing.T) {
	var observedBase string
	staleTestEnv(t,
		[]git.Worktree{{Path: "/wt/base-check", Branch: "feat/base-check"}},
		map[string]string{},
		map[string]bool{},
	)
	orig := mockIsBranchMergedFunc
	mockIsBranchMergedFunc = func(_, base string) (bool, error) {
		observedBase = base
		return true, nil
	}
	t.Cleanup(func() { mockIsBranchMergedFunc = orig })

	if _, err := runStaleClean(t, map[string]string{"stale": "true"}); err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if observedBase != "origin/main" {
		t.Fatalf("the default base must be origin/main (the ref prMergeCleanup compares against), got %q", observedBase)
	}
}

// TestCleanStale_NoAskUserQuestion is AC-WR-020: the changed surface carries no
// interactive prompt. The name is exact and anchored on purpose — an
// unanchored `-run NoAskUserQuestion` matches 31 existing guards and would
// bind nothing about this change.
func TestCleanStale_NoAskUserQuestion(t *testing.T) {
	sources, err := filepath.Glob("clean*.go")
	if err != nil {
		t.Fatalf("glob clean sources: %v", err)
	}
	scanned := 0
	for _, name := range sources {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		scanned++
		if cleanPromptGuard(string(data)) {
			t.Errorf("%s must stay prompt-free: it references an interactive prompt", name)
		}
	}
	// Positive control on the scan itself: a glob that matched nothing would
	// report every file clean without reading one.
	if scanned < 1 {
		t.Fatalf("guard scanned %d clean sources, want the whole surface", scanned)
	}
	// Negative control: the guard must flag a synthetic violation.
	if !cleanPromptGuard("x := AskUserQuestion()") {
		t.Error("guard must detect an AskUserQuestion reference (negative control)")
	}
}

// cleanPromptGuard reports whether src references the orchestrator-only
// user-question channel. The CLI runs in subagent context, where prompting is
// a dead channel; missing inputs become flags and structured errors instead.
func cleanPromptGuard(src string) bool {
	return strings.Contains(src, "AskUserQuestion") || strings.Contains(src, "mcp__askuser")
}
