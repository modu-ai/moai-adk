package hook

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/guardliveness"
)

// runGit executes one git command in dir, failing the test on error. Test-side
// only: the deliverable itself carries no such call, which is what
// TestAdvisoryPathIssuesNoMutatingCall counts.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

// fixtureRepo builds a throwaway repository whose porcelain output is
// deliberately NON-empty: one modified tracked file and one untracked file. An
// empty baseline compares equal to an empty result even when a write happened
// and was somehow ignored, so the fixture carries state a write would disturb.
func fixtureRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	runGit(t, repo, "init", "--quiet")
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatalf("seed tracked file: %v", err)
	}
	runGit(t, repo, "add", "tracked.txt")
	runGit(t, repo, "commit", "--quiet", "-m", "seed")
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("two\n"), 0o644); err != nil {
		t.Fatalf("modify tracked file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "untracked.txt"), []byte("three\n"), 0o644); err != nil {
		t.Fatalf("seed untracked file: %v", err)
	}
	return repo
}

// AC-GDL-008 (b) — the evaluated working tree is byte-identical before and
// after a render, verified by `git status --porcelain` across the run.
//
// The fixture carries a non-clean entry on purpose: it is the case most likely
// to tempt an action. The persistence the change-leading diff requires lives
// outside the tree, which is an output of this criterion rather than a
// preference — a cache written into the tree passes the mutating-call count of
// clause (a) and still leaves drift for the next reader.
func TestGuardLivenessRenderLeavesTheWorkingTreeByteIdentical(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH")
	}

	storeDir := t.TempDir()
	store := guardliveness.NewStore(storeDir)
	orig := newGuardLivenessStore
	t.Cleanup(func() { newGuardLivenessStore = orig })
	newGuardLivenessStore = func() (*guardliveness.Store, error) { return store, nil }

	repo := fixtureRepo(t)
	if strings.HasPrefix(storeDir, repo) {
		t.Fatalf("the persistence directory %q sits inside the evaluated tree %q", storeDir, repo)
	}
	if err := store.Save(repo, nonCleanResult(), time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("seed the persisted result: %v", err)
	}

	before := runGit(t, repo, "status", "--porcelain")
	if strings.TrimSpace(before) == "" {
		t.Fatal("the fixture's porcelain baseline is empty, so the comparison would hold vacuously")
	}

	text := guardLivenessAdvisory(repo, deferredScansAsyncEnabled())
	if text == "" {
		t.Fatal("the render said nothing on a result carrying a non-clean entry, so nothing was measured")
	}

	after := runGit(t, repo, "status", "--porcelain")
	if before != after {
		t.Fatalf("the render changed the working tree.\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// The render's own memory of what it last announced must live outside the
// evaluated tree, and it must actually be written — a diff against nothing
// re-announces every standing entry every session, which is the noise profile
// REQ-GDL-007 exists to remove.
func TestGuardLivenessRenderPersistsItsRecordOutsideTheTree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH")
	}

	storeDir := t.TempDir()
	store := guardliveness.NewStore(storeDir)
	orig := newGuardLivenessStore
	t.Cleanup(func() { newGuardLivenessStore = orig })
	newGuardLivenessStore = func() (*guardliveness.Store, error) { return store, nil }

	repo := fixtureRepo(t)
	if err := store.Save(repo, nonCleanResult(), time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("seed the persisted result: %v", err)
	}

	first := guardLivenessAdvisory(repo, deferredScansAsyncEnabled())
	if !strings.Contains(first, "subject-fires") {
		t.Fatalf("the first render did not announce the non-clean subject:\n%s", first)
	}

	second := guardLivenessAdvisory(repo, deferredScansAsyncEnabled())
	if strings.Contains(second, "subject-fires") {
		t.Fatalf("the second render re-announced a standing subject individually:\n%s", second)
	}
	if second == "" {
		t.Fatal("the second render went silent while a subject was still not reporting clean")
	}

	rec, err := store.LoadRendered(repo)
	if err != nil {
		t.Fatalf("LoadRendered: %v", err)
	}
	if len(rec.Classifications) == 0 {
		t.Fatal("the render recorded nothing, so nothing carries across sessions")
	}
	if strings.HasPrefix(storeDir, repo) {
		t.Fatalf("the render record %q sits inside the evaluated tree %q", storeDir, repo)
	}
}
