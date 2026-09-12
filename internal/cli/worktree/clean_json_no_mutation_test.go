package worktree

// clean_json_no_mutation_test.go — `clean --json` is the whole command (#1704).
//
// The flag's help calls it a report that "removes nothing", but --json was read
// inside the --stale branch alone. `moai worktree clean --json` therefore fell
// through to the default prune and printed a box banner where an inventory was
// promised, and `--merged-only --json` removed merged worktrees while printing
// human text. REQ-WR-013 is the contract these cases pin: in --json mode the
// command reports and mutates nothing, whatever else is on the command line.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// jsonModeEnv installs the stale test environment and additionally counts
// Prune calls, which the default path makes and the report must not.
func jsonModeEnv(t *testing.T, worktrees []git.Worktree) (removed *[]string, pruned *int) {
	t.Helper()

	removedPaths := staleTestEnv(t, worktrees, nil, map[string]bool{})
	pruneCalls := 0
	provider, ok := WorktreeProvider.(*mockWorktreeManager)
	if !ok {
		t.Fatalf("WorktreeProvider is %T, want the mock installed by staleTestEnv", WorktreeProvider)
	}
	provider.pruneFunc = func() error {
		pruneCalls++
		return nil
	}
	return removedPaths, &pruneCalls
}

func decodeReport(t *testing.T, out string) []staleCandidate {
	t.Helper()
	var got []staleCandidate
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("stdout must be a JSON inventory, got %d bytes: %v\n%s", len(out), err, out)
	}
	return got
}

// TestClean_JSONWithoutStaleReportsAndMutatesNothing is the reported case:
// `moai worktree clean --json`, no other flag.
func TestClean_JSONWithoutStaleReportsAndMutatesNothing(t *testing.T) {
	removed, pruned := jsonModeEnv(t, []git.Worktree{
		{Path: "/wt/slug", Branch: "worktree-reaper"},
		{Path: "/wt/other", Branch: "docs-refresh"},
	})

	out, err := runStaleClean(t, map[string]string{"json": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}

	if *pruned != 0 {
		t.Errorf("--json pruned %d time(s); the help calls this path inert", *pruned)
	}
	if len(*removed) != 0 {
		t.Errorf("--json removed %v; it must remove nothing", *removed)
	}
	if strings.Contains(out, "Cleaned stale worktree references") {
		t.Errorf("--json printed the human banner instead of an inventory:\n%s", out)
	}
	if got := decodeReport(t, out); len(got) != 2 {
		t.Errorf("expected one object per registered worktree (2), got %d:\n%s", len(got), out)
	}
}

// TestClean_JSONWithMergedOnlyReportsAndRemovesNothing covers the other
// window: --merged-only removes merged worktrees, and --json was ignored there
// too, so the inventory flag used to delete.
func TestClean_JSONWithMergedOnlyReportsAndRemovesNothing(t *testing.T) {
	removed, pruned := jsonModeEnv(t, []git.Worktree{
		{Path: "/wt/merged", Branch: "WT-merged-card"},
	})
	mockIsBranchMergedFunc = func(string, string) (bool, error) { return true, nil }

	out, err := runStaleClean(t, map[string]string{"json": "true", "merged-only": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}

	if len(*removed) != 0 {
		t.Errorf("--merged-only --json removed %v; --json removes nothing", *removed)
	}
	if *pruned != 0 {
		t.Errorf("--merged-only --json pruned %d time(s)", *pruned)
	}
	if got := decodeReport(t, out); len(got) != 1 {
		t.Errorf("expected the merged tree in the inventory, got %d objects:\n%s", len(got), out)
	}
}

// TestClean_JSONOverridesYes keeps the property the --stale branch already
// had: a report cannot delete because another flag asked it to.
func TestClean_JSONOverridesYes(t *testing.T) {
	removed, pruned := jsonModeEnv(t, []git.Worktree{
		{Path: "/wt/removable", Branch: "WT-removable-card"},
	})
	mockIsBranchMergedFunc = func(string, string) (bool, error) { return true, nil }

	out, err := runStaleClean(t, map[string]string{"json": "true", "stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}

	if len(*removed) != 0 {
		t.Errorf("--stale --yes --json removed %v; --json removes nothing", *removed)
	}
	if *pruned != 0 {
		t.Errorf("--stale --yes --json pruned %d time(s)", *pruned)
	}
	if got := decodeReport(t, out); len(got) != 1 {
		t.Errorf("expected one object, got %d:\n%s", len(got), out)
	}
}

// TestClean_WithoutJSONStillPrunesAndReportsIt is the negative control: the
// default path is the one that mutates, and it must keep doing so.
func TestClean_WithoutJSONStillPrunesAndReportsIt(t *testing.T) {
	_, pruned := jsonModeEnv(t, nil)

	out, err := runStaleClean(t, map[string]string{})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}

	if *pruned != 1 {
		t.Errorf("the default path pruned %d time(s), want 1", *pruned)
	}
	if !strings.Contains(out, "Cleaned stale worktree references") {
		t.Errorf("the default path must still say what it did:\n%s", out)
	}
}
