package worktree

// Hoist tracked-path tests — card t2013.
//
// A card tree's .moai/reports/ holds both git-tracked files (already present in
// the primary checkout) and untracked run-phase evidence. Hoisting must copy
// only the untracked evidence; copying tracked paths duplicates what the
// primary checkout already has.

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// newTrackedReportsTree builds a real git repo whose .moai/reports/ holds the
// given tracked files (committed), and returns its path (base name "t2013").
func newTrackedReportsTree(t *testing.T, tracked map[string]string) string {
	t.Helper()
	tree := filepath.Join(t.TempDir(), "t2013")
	if err := os.MkdirAll(tree, 0o755); err != nil {
		t.Fatalf("mkdir tree: %v", err)
	}
	runGitDir(t, tree, "init", "-q")
	for rel, content := range tracked {
		writeFileRel(t, tree, ".moai/reports/"+rel, content)
	}
	runGitDir(t, tree, "add", "-f", ".moai/reports")
	runGitDir(t, tree, "-c", "user.email=t@test", "-c", "user.name=t", "commit", "--allow-empty", "-q", "-m", "tracked reports")
	return tree
}

func hoistedFiles(t *testing.T, res HoistResult) []string {
	t.Helper()
	var got []string
	if _, err := os.Stat(res.Dest); os.IsNotExist(err) {
		return got
	}
	err := filepath.WalkDir(res.Dest, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(res.Dest, path)
		got = append(got, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", res.Dest, err)
	}
	sort.Strings(got)
	return got
}

// (i) tracked files only -> nothing is copied.
func TestHoistTracked_OnlyTrackedCopiesNothing(t *testing.T) {
	tree := newTrackedReportsTree(t, map[string]string{
		"a.md":         "a\n",
		"sub/b.md":     "b\n",
		"t100/verdict": "v\n",
	})
	main := t.TempDir()

	res, err := hoistWorktreeReports(tree, main)
	if err != nil {
		t.Fatalf("hoistWorktreeReports: %v", err)
	}
	if res.Count != 0 || res.Bytes != 0 {
		t.Errorf("tracked-only hoist = (%d files, %d bytes), want (0, 0)", res.Count, res.Bytes)
	}
	if got := hoistedFiles(t, res); len(got) != 0 {
		t.Errorf("destination holds %v, want nothing", got)
	}
}

// (ii) tracked files plus exactly one untracked file -> only that one is copied.
func TestHoistTracked_CopiesOnlyUntrackedEvidence(t *testing.T) {
	tree := newTrackedReportsTree(t, map[string]string{
		"a.md":     "a\n",
		"sub/b.md": "b\n",
	})
	writeFileRel(t, tree, ".moai/reports/sub/new evidence.md", "fresh evidence\n")
	main := t.TempDir()

	res, err := hoistWorktreeReports(tree, main)
	if err != nil {
		t.Fatalf("hoistWorktreeReports: %v", err)
	}
	if res.Count != 1 {
		t.Errorf("Count = %d, want 1", res.Count)
	}
	got := hoistedFiles(t, res)
	if len(got) != 1 || got[0] != "sub/new evidence.md" {
		t.Fatalf("hoisted files = %v, want [sub/new evidence.md]", got)
	}
	if body := readFileStrict(t, filepath.Join(res.Dest, "sub", "new evidence.md")); body != "fresh evidence\n" {
		t.Errorf("hoisted content = %q", body)
	}
}

// (iii) A tracked file with uncommitted modifications still counts as tracked
// and is skipped: the primary checkout already holds the committed version, so
// the tracked set is decided by index membership, not by working-tree cleanliness.
func TestHoistTracked_ModifiedTrackedFileStillSkipped(t *testing.T) {
	tree := newTrackedReportsTree(t, map[string]string{"a.md": "committed\n"})
	writeFileRel(t, tree, ".moai/reports/a.md", "locally modified\n")
	main := t.TempDir()

	res, err := hoistWorktreeReports(tree, main)
	if err != nil {
		t.Fatalf("hoistWorktreeReports: %v", err)
	}
	if res.Count != 0 {
		t.Errorf("Count = %d, want 0 (modified tracked file is skipped)", res.Count)
	}
	if got := hoistedFiles(t, res); len(got) != 0 {
		t.Errorf("destination holds %v, want nothing", got)
	}
}

// (iv) git cannot list the index (tree is not a repository) -> fail open to
// copy-everything; untracked evidence is never lost.
func TestHoistTracked_GitFailureFailsOpenToCopyAll(t *testing.T) {
	tree := filepath.Join(t.TempDir(), "t2013")
	writeFileRel(t, tree, ".moai/reports/a.md", "a\n")
	writeFileRel(t, tree, ".moai/reports/sub/b.md", "b\n")
	main := t.TempDir()

	res, err := hoistWorktreeReports(tree, main)
	if err != nil {
		t.Fatalf("hoistWorktreeReports: %v", err)
	}
	if res.Count != 2 {
		t.Errorf("Count = %d, want 2 (fail-open copy-all)", res.Count)
	}
	got := hoistedFiles(t, res)
	if len(got) != 2 || got[0] != "a.md" || got[1] != "sub/b.md" {
		t.Errorf("hoisted files = %v, want [a.md sub/b.md]", got)
	}
}
