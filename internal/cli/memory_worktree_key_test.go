// memory_worktree_key_test.go — which stores a worktree session audits.
//
// The store is keyed on the session's own working directory, so a session in a
// linked worktree keys on the worktree. That per-cwd divergence is deliberate
// on the WRITE path (.moai/docs/memory-dir-resolution-doctrine.md refuses to
// normalize it), but a worktree store is usually absent, so `moai memory
// doctor` run from a lane session reported "not present" twice and said
// nothing about the repository store that actually holds every memory.
//
// The contract these tests pin: the worktree's own key is KEPT, and the
// primary checkout's key is ADDED beside it. Either half alone is a defect —
// dropping the worktree key contradicts the doctrine, dropping the primary key
// is the original complaint.
package cli

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initRepoWithWorktree builds a real repository and links a worktree to it.
// The worktree is created OUTSIDE the repository directory on purpose: a
// derivation that merely strips a trailing path segment would satisfy an
// in-tree layout while still being wrong, and this layout rejects it.
func initRepoWithWorktree(t *testing.T) (repo, worktree string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}

	base := t.TempDir()
	repo = filepath.Join(base, "primary")
	worktree = filepath.Join(base, "elsewhere", "wt-card")

	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}

	if err := exec.Command("git", "init", "-q", repo).Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	run(repo, "config", "user.email", "t@example.com")
	run(repo, "config", "user.name", "t")
	run(repo, "commit", "-q", "--allow-empty", "-m", "seed")
	run(repo, "worktree", "add", "-q", "-b", "WT-probe", worktree)
	return repo, worktree
}

// storeSegments returns the key segment of each store — the directory name
// directly above "memory", which is the slug.
//
// Compared as whole segments, never as substrings: a symlink-resolved slug
// CONTAINS the unresolved one (-private-var-… contains -var-…), so a substring
// check passes on a key that is wrong by a whole path prefix.
func storeSegments(stores []memoryStore) []string {
	out := make([]string, 0, len(stores))
	for _, s := range stores {
		out = append(out, filepath.Base(filepath.Dir(s.Dir)))
	}
	return out
}

func containsSegment(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

// TestMemoryCandidateStores_WorktreeAlsoAuditsPrimaryCheckout is the
// regression guard for the reported defect: from inside a worktree the
// repository's own store must be among the audited candidates.
//
// Falsifiability: delete the memoryPrimaryCheckout branch from
// memoryCandidateStores and this test fails — only the worktree's own key is
// then resolved, which is the "not present" report the card was filed about.
func TestMemoryCandidateStores_WorktreeAlsoAuditsPrimaryCheckout(t *testing.T) {
	// Cannot run in parallel: t.Setenv mutates process-wide state.
	repo, worktree := initRepoWithWorktree(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(home, "cfg"))

	fromWorktree, err := memoryCandidateStores(worktree)
	if err != nil {
		t.Fatalf("memoryCandidateStores(worktree): %v", err)
	}
	fromRepo, err := memoryCandidateStores(repo)
	if err != nil {
		t.Fatalf("memoryCandidateStores(repo): %v", err)
	}
	if len(fromWorktree) == 0 || len(fromRepo) == 0 {
		t.Fatal("no store resolved; every assertion below would be vacuous")
	}

	got := storeSegments(fromWorktree)

	// The repository's key is expected symlink-RESOLVED, because git resolves
	// the common directory itself (measured: under macOS's temp root it hands
	// back /private/var/… for a /var/… repository). Expecting the unresolved
	// path here makes the test red against correct code.
	wantPrimary := memoryProjectSlug(mustEvalSymlinks(t, repo))
	if !containsSegment(got, wantPrimary) {
		t.Errorf("primary checkout's store is not audited from the worktree:\n  want segment: %s\n  got         : %v",
			wantPrimary, got)
	}

	// And the worktree's own key is still there. Without this half, a fix that
	// simply replaced the key would pass — and replacing it is what the
	// per-cwd doctrine refuses.
	wantWorktree := memoryProjectSlug(worktree)
	if !containsSegment(got, wantWorktree) {
		t.Errorf("worktree's own store was dropped:\n  want segment: %s\n  got         : %v",
			wantWorktree, got)
	}

	// The worktree audits strictly more than the primary checkout does: its own
	// keys plus the repository's. A count that did not grow means the two keys
	// collapsed into one.
	if len(fromWorktree) <= len(fromRepo) {
		t.Errorf("worktree resolved %d stores, primary checkout %d — expected strictly more",
			len(fromWorktree), len(fromRepo))
	}
}

// TestMemoryCandidateStores_PrimaryCheckoutIsUnchanged pins the other
// direction: from the primary checkout there is no second key, so nothing is
// added and the resolved set is exactly what it always was. Without this, a
// derivation that appended some root unconditionally would satisfy the test
// above while doubling every ordinary run's output.
//
// The expectation is the UNRESOLVED path: outside the worktree branch the
// derivation does not resolve symlinks, and on macOS t.TempDir() hands back
// /var/… which resolves to /private/var/…, so a widening would fail here.
func TestMemoryCandidateStores_PrimaryCheckoutIsUnchanged(t *testing.T) {
	// Cannot run in parallel: t.Setenv mutates process-wide state.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(home, "cfg"))

	plain := t.TempDir()
	stores, err := memoryCandidateStores(plain)
	if err != nil {
		t.Fatalf("memoryCandidateStores: %v", err)
	}
	if len(stores) == 0 {
		t.Fatal("no store resolved; the assertions below would be vacuous")
	}

	want := memoryProjectSlug(plain)
	for _, got := range storeSegments(stores) {
		if got != want {
			t.Errorf("a non-worktree path gained a foreign key:\n  want segment: %s\n  got segment : %s", want, got)
		}
	}
	// Two config roots, one key: the profile store and the default store.
	if len(stores) != 2 {
		t.Errorf("expected exactly 2 stores for a single key, got %d: %v", len(stores), storeSegments(stores))
	}
}

// TestMemoryCandidateStores_IdenticalRootsCollapse pins the de-duplication the
// pre-refactor code carried explicitly: when CLAUDE_CONFIG_DIR points at the
// default root, the profile store and the default store ARE the same
// directory, and it must be listed once.
//
// Without this the refactor could have dropped the check silently — every
// other test here sets CLAUDE_CONFIG_DIR to a distinct path, so none of them
// exercises the collision.
func TestMemoryCandidateStores_IdenticalRootsCollapse(t *testing.T) {
	// Cannot run in parallel: t.Setenv mutates process-wide state.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(home, ".claude"))

	stores, err := memoryCandidateStores(t.TempDir())
	if err != nil {
		t.Fatalf("memoryCandidateStores: %v", err)
	}
	if len(stores) != 1 {
		dirs := make([]string, 0, len(stores))
		for _, s := range stores {
			dirs = append(dirs, s.Dir)
		}
		t.Errorf("expected the two roots to collapse into 1 store, got %d:\n  %s",
			len(stores), strings.Join(dirs, "\n  "))
	}
}

// mustEvalSymlinks resolves path, failing the test if it cannot — a silent
// fallback to the unresolved path would make the assertion above compare the
// wrong thing while still looking like it checked something.
func mustEvalSymlinks(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s): %v", path, err)
	}
	return resolved
}
