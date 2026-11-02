package worktree

// Hoist tracked-path tests — card t2013.
//
// A card tree's .moai/reports/ holds git-tracked files and untracked run-phase
// evidence. Hoisting skips a tracked path only when the primary checkout's HEAD
// holds the identical blob and the working file is unmodified; every other path
// (untracked, modified, staged differently, committed only on the card branch)
// is copied (sync-audit F6).

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// newTrackedReportsTree builds a primary checkout whose HEAD commits the given
// tracked files under .moai/reports/, plus a linked worktree "t2013" checked out
// from that commit (the card tree). The hoist compares the card tree's blobs
// with the primary checkout's HEAD, so the fixture needs a real second
// checkout. It returns (tree, primary).
func newTrackedReportsTree(t *testing.T, tracked map[string]string) (string, string) {
	t.Helper()
	base := t.TempDir()
	primary := filepath.Join(base, "primary")
	tree := filepath.Join(base, "t2013")
	if err := os.MkdirAll(primary, 0o755); err != nil {
		t.Fatalf("mkdir primary: %v", err)
	}
	runGitDir(t, primary, "init", "-q")
	for rel, content := range tracked {
		writeFileRel(t, primary, ".moai/reports/"+rel, content)
	}
	runGitDir(t, primary, "add", "-f", ".moai/reports")
	runGitDir(t, primary, "-c", "user.email=t@test", "-c", "user.name=t", "commit", "--allow-empty", "-q", "-m", "tracked reports")
	runGitDir(t, primary, "worktree", "add", "-q", "-b", "card", tree)
	return tree, primary
}

// commitInTree commits everything under .moai/reports in the card tree.
func commitInTree(t *testing.T, tree string) {
	t.Helper()
	runGitDir(t, tree, "add", "-f", ".moai/reports")
	runGitDir(t, tree, "-c", "user.email=t@test", "-c", "user.name=t", "commit", "-q", "-m", "card reports")
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
	tree, main := newTrackedReportsTree(t, map[string]string{
		"a.md":         "a\n",
		"sub/b.md":     "b\n",
		"t100/verdict": "v\n",
	})

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
	tree, main := newTrackedReportsTree(t, map[string]string{
		"a.md":     "a\n",
		"sub/b.md": "b\n",
	})
	writeFileRel(t, tree, ".moai/reports/sub/new evidence.md", "fresh evidence\n")

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

// (iii) A tracked file with uncommitted modifications is COPIED. This inverts
// the card's earlier expectation (modified tracked files were skipped as
// "already held by the primary checkout"); sync-audit F6 found that premise
// false: the primary checkout holds only the committed version, so skipping the
// modified working file loses the evidence on a `done --force` disposal.
func TestHoistTracked_ModifiedTrackedFileIsCopied(t *testing.T) {
	tree, main := newTrackedReportsTree(t, map[string]string{"a.md": "committed\n"})
	writeFileRel(t, tree, ".moai/reports/a.md", "locally modified\n")

	res, err := hoistWorktreeReports(tree, main)
	if err != nil {
		t.Fatalf("hoistWorktreeReports: %v", err)
	}
	if res.Count != 1 {
		t.Errorf("Count = %d, want 1 (a modified tracked file is not what the primary holds)", res.Count)
	}
	got := hoistedFiles(t, res)
	if len(got) != 1 || got[0] != "a.md" {
		t.Fatalf("hoisted files = %v, want [a.md]", got)
	}
	if body := readFileStrict(t, filepath.Join(res.Dest, "a.md")); body != "locally modified\n" {
		t.Errorf("hoisted content = %q, want the modified working content", body)
	}
}

// F6: a tracked report committed only on the card branch is absent from the
// primary checkout's HEAD, so the hoist must copy it.
func TestHoistTracked_CommittedOnlyOnCardBranchIsCopied(t *testing.T) {
	tree, main := newTrackedReportsTree(t, map[string]string{"a.md": "a\n"})
	writeFileRel(t, tree, ".moai/reports/t2013/verdict.md", "card-only verdict\n")
	commitInTree(t, tree)
	if out := runGitDir(t, tree, "status", "--porcelain"); out != "" {
		t.Fatalf("fixture: card tree not clean after the commit:\n%s", out)
	}

	res, err := hoistWorktreeReports(tree, main)
	if err != nil {
		t.Fatalf("hoistWorktreeReports: %v", err)
	}
	got := hoistedFiles(t, res)
	if res.Count != 1 || len(got) != 1 || got[0] != "t2013/verdict.md" {
		t.Fatalf("Count = %d, hoisted = %v, want exactly [t2013/verdict.md] (a.md is identical to the primary and stays skipped)", res.Count, got)
	}
}

// F6: a tracked report whose committed content on the card branch differs from
// the primary checkout's HEAD blob is copied, even with a clean working tree.
func TestHoistTracked_CommittedDifferentBlobIsCopied(t *testing.T) {
	tree, main := newTrackedReportsTree(t, map[string]string{"a.md": "primary version\n", "same.md": "same\n"})
	writeFileRel(t, tree, ".moai/reports/a.md", "card version\n")
	commitInTree(t, tree)

	res, err := hoistWorktreeReports(tree, main)
	if err != nil {
		t.Fatalf("hoistWorktreeReports: %v", err)
	}
	got := hoistedFiles(t, res)
	if res.Count != 1 || len(got) != 1 || got[0] != "a.md" {
		t.Fatalf("Count = %d, hoisted = %v, want exactly [a.md]", res.Count, got)
	}
	if body := readFileStrict(t, filepath.Join(res.Dest, "a.md")); body != "card version\n" {
		t.Errorf("hoisted content = %q, want the card branch version", body)
	}
}

// F6: a staged-only modification (index differs from HEAD, working file equals
// the index) is copied: the index blob is not what the primary checkout holds.
func TestHoistTracked_StagedModificationIsCopied(t *testing.T) {
	tree, main := newTrackedReportsTree(t, map[string]string{"a.md": "committed\n"})
	writeFileRel(t, tree, ".moai/reports/a.md", "staged change\n")
	runGitDir(t, tree, "add", "-f", ".moai/reports/a.md")

	res, err := hoistWorktreeReports(tree, main)
	if err != nil {
		t.Fatalf("hoistWorktreeReports: %v", err)
	}
	got := hoistedFiles(t, res)
	if res.Count != 1 || len(got) != 1 || got[0] != "a.md" {
		t.Fatalf("Count = %d, hoisted = %v, want exactly [a.md]", res.Count, got)
	}
}

// F6: tracked and byte-identical to the primary checkout's HEAD, in a tree that
// also carries other tracked-and-different and untracked files, stays skipped.
func TestHoistTracked_IdenticalToPrimaryIsSkipped(t *testing.T) {
	tree, main := newTrackedReportsTree(t, map[string]string{"same.md": "same\n", "sub/also same.md": "same too\n"})
	writeFileRel(t, tree, ".moai/reports/untracked.md", "u\n")

	res, err := hoistWorktreeReports(tree, main)
	if err != nil {
		t.Fatalf("hoistWorktreeReports: %v", err)
	}
	got := hoistedFiles(t, res)
	if res.Count != 1 || len(got) != 1 || got[0] != "untracked.md" {
		t.Fatalf("Count = %d, hoisted = %v, want exactly [untracked.md]", res.Count, got)
	}
}

// F6: when the primary checkout cannot be queried (not a repository), the hoist
// fails open: everything is copied and the fallback warning is set.
func TestHoistTracked_PrimaryUnreadableFailsOpen(t *testing.T) {
	tree, _ := newTrackedReportsTree(t, map[string]string{"a.md": "a\n"})
	notARepo := t.TempDir()

	res, err := hoistWorktreeReports(tree, notARepo)
	if err != nil {
		t.Fatalf("hoistWorktreeReports: %v", err)
	}
	if res.Count != 1 || res.TrackedFallback == "" {
		t.Errorf("Count = %d, TrackedFallback = %q, want 1 copied and a fallback warning", res.Count, res.TrackedFallback)
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
