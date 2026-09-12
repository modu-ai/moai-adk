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
	"errors"
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

// TestClean_JSONModesReportAndMutateNothing walks the three ways --json can
// be reached. They share one body on purpose: the property under test is the
// same in all three, and a per-mode copy is how the --stale branch came to be
// the only one that honoured the flag.
func TestClean_JSONModesReportAndMutateNothing(t *testing.T) {
	cases := []struct {
		name      string
		flags     map[string]string
		worktrees []git.Worktree
		merged    bool
		wantCount int
	}{
		{
			// The reported case: `moai worktree clean --json`, no other flag.
			name:  "json alone",
			flags: map[string]string{"json": "true"},
			worktrees: []git.Worktree{
				{Path: "/wt/slug", Branch: "worktree-reaper"},
				{Path: "/wt/other", Branch: "docs-refresh"},
			},
			wantCount: 2,
		},
		{
			// --merged-only removes merged worktrees, and --json was ignored
			// there too, so the inventory flag used to delete.
			name:      "json with merged-only",
			flags:     map[string]string{"json": "true", "merged-only": "true"},
			worktrees: []git.Worktree{{Path: "/wt/merged", Branch: "WT-merged-card"}},
			merged:    true,
			wantCount: 1,
		},
		{
			// The property the --stale branch already had: a report cannot
			// delete because another flag asked it to.
			name:      "json overrides yes",
			flags:     map[string]string{"json": "true", "stale": "true", "yes": "true"},
			worktrees: []git.Worktree{{Path: "/wt/removable", Branch: "WT-removable-card"}},
			merged:    true,
			wantCount: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			removed, pruned := jsonModeEnv(t, tc.worktrees)
			mockIsBranchMergedFunc = func(string, string) (bool, error) { return tc.merged, nil }

			out, err := runStaleClean(t, tc.flags)
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
			if got := decodeReport(t, out); len(got) != tc.wantCount {
				t.Errorf("expected %d object(s) in the inventory, got %d:\n%s", tc.wantCount, len(got), out)
			}
		})
	}
}

// TestClean_JSONReportsAnUnavailableWorktree is what dropping the prune buys:
// a worktree whose directory is gone stays IN the inventory, with its state
// undetermined and a keep_reason naming what could not be read. The prune used
// to drop that entry before the report was built, so the operator saw nothing
// where the interesting case was.
func TestClean_JSONReportsAnUnavailableWorktree(t *testing.T) {
	removed, pruned := jsonModeEnv(t, []git.Worktree{
		{Path: "/wt/gone", Branch: "WT-vanished-card"},
		{Path: "/wt/live", Branch: "docs-refresh"},
	})
	mockIsBranchMergedFunc = func(string, string) (bool, error) { return true, nil }
	// The vanished tree's status call fails the way git does when the
	// directory is no longer there; the live one answers normally.
	gitWorktreeCmd = func(args ...string) (string, error) {
		if len(args) >= 2 && args[0] == "-C" && args[1] == "/wt/gone" {
			return "", errors.New("cannot chdir to '/wt/gone': No such file or directory")
		}
		return "", nil
	}

	out, err := runStaleClean(t, map[string]string{"json": "true"})
	if err != nil {
		t.Fatalf("runClean error: %v", err)
	}
	if *pruned != 0 || len(*removed) != 0 {
		t.Errorf("the report pruned %d time(s) and removed %v; it must do neither", *pruned, *removed)
	}

	byPath := map[string]staleCandidate{}
	for _, c := range decodeReport(t, out) {
		byPath[c.Path] = c
	}

	gone, ok := byPath["/wt/gone"]
	if !ok {
		t.Fatalf("the unavailable worktree is missing from the inventory:\n%s", out)
	}
	if gone.Dirty != staleStateUndetermined {
		t.Errorf("unavailable worktree reported dirty=%q, want %q", gone.Dirty, staleStateUndetermined)
	}
	if gone.KeepReason == "" {
		t.Error("unavailable worktree reported an empty keep_reason; the report must name what it could not read")
	}
	if _, ok := byPath["/wt/live"]; !ok {
		t.Errorf("the reachable worktree vanished from the inventory:\n%s", out)
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
