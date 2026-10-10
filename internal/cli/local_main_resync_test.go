package cli

// local_main_resync_test.go — SPEC-LOCAL-MAIN-FLOW-001 (card t1616), M1 step 1
// RED tests for the fast-forward re-sync of the primary checkout's local main
// (REQ-LMF-014, plan §B3a). Each test runs runLocalMainResync against a real
// git fixture. The origin value is set with git update-ref in the scratch
// repository, and the fetch seam is replaced with a no-op, so no test contacts
// a remote. The stub in local_main_resync.go returns "not implemented", so each
// assertion below fails at its own check.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// lmfResyncFixture builds a primary checkout on main (lmfRepo), stamps the lane
// session and label, and replaces the origin fetch with a no-op for the test.
func lmfResyncFixture(t *testing.T, gitignore string) string {
	t.Helper()
	root, _ := lmfRepo(t, true, gitignore)
	t.Setenv(config.EnvClaudeCodeSessionID, lmfSession)
	sdLaneEnv(t, "lane-1", "")
	prev := localMainResyncFetch
	localMainResyncFetch = func(string) error { return nil }
	t.Cleanup(func() { localMainResyncFetch = prev })
	return root
}

// lmfLocalCommit commits path with content on the primary's local main and
// returns the new HEAD.
func lmfLocalCommit(t *testing.T, root, path, content string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, path), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	fcGit(t, root, "add", path)
	fcGit(t, root, "commit", "-q", "-m", "local "+path)
	return lmfHead(t, root)
}

// lmfOriginCommit commits top-level files on a scratch branch cut from base and
// points refs/remotes/origin/main at the new commit with git update-ref. Paths
// the primary ignores are added with force.
func lmfOriginCommit(t *testing.T, root, base, branch string, files map[string]string, force bool) string {
	t.Helper()
	wt := filepath.Join(t.TempDir(), branch)
	fcGit(t, root, "worktree", "add", "-q", "-b", branch, wt, base)
	add := []string{"add"}
	if force {
		add = append(add, "-f")
	}
	for p, content := range files {
		if err := os.WriteFile(filepath.Join(wt, p), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		add = append(add, p)
	}
	fcGit(t, wt, add...)
	fcGit(t, wt, "commit", "-q", "-m", branch)
	sha := fcGit(t, wt, "rev-parse", "HEAD")
	fcGit(t, root, "update-ref", "refs/remotes/origin/main", sha)
	return sha
}

// --- Re-sync of local main (REQ-LMF-014, plan §B3a) ---

func TestLocalMainResyncFastForwards(t *testing.T) {
	// Local main is one commit behind origin/main, which descends from it. The
	// re-sync moves local main to origin/main by fast-forward and reports both
	// SHAs.
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	target := lmfOriginCommit(t, root, base, "origin-ahead", map[string]string{"ahead.txt": "ahead\n"}, false)
	report, err := runLocalMainResync(root)
	if err != nil {
		t.Fatalf("local main behind origin/main must fast-forward: %v", err)
	}
	if head := lmfHead(t, root); head != target {
		t.Fatalf("local main must move from %s to origin/main %s, got HEAD %s", base, target, head)
	}
	if sym := fcGit(t, root, "symbolic-ref", "HEAD"); sym != "refs/heads/"+lmfBranch {
		t.Fatalf("the re-sync must not switch branches: HEAD names %q", sym)
	}
	if !strings.Contains(report, base[:12]) || !strings.Contains(report, target[:12]) {
		t.Fatalf("the report must name the old and new SHAs %s and %s: %q", base[:12], target[:12], report)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the re-sync must release the window: %+v", lock)
	}
}

func TestLocalMainResyncRefusesDiverged(t *testing.T) {
	// Local main holds a commit that origin/main lacks, and origin/main holds a
	// commit that local main lacks. The re-sync must refuse without reconciling.
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	local := lmfLocalCommit(t, root, "local.txt", "local\n")
	lmfOriginCommit(t, root, base, "origin-diverged", map[string]string{"remote.txt": "remote\n"}, false)
	_, err := runLocalMainResync(root)
	if err == nil {
		t.Fatalf("a diverged local main must refuse the re-sync")
	}
	// Plan §B3a names no class for this refusal. The nearest existing class is
	// MergeExitNotDescendant: the fast-forward target does not descend from HEAD.
	if code, ok := factory.MergeExitCode(err); !ok || code != factory.MergeExitNotDescendant {
		t.Fatalf("the diverged refusal must carry MergeExitNotDescendant (%d), got code %d (ok=%v): %v", factory.MergeExitNotDescendant, code, ok, err)
	}
	if head := lmfHead(t, root); head != local {
		t.Fatalf("a diverged refusal must not move HEAD: %s -> %s", local, head)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("a diverged refusal must release the window: %+v", lock)
	}
}

func TestLocalMainResyncAheadIsNoop(t *testing.T) {
	// Local main already contains origin/main. The re-sync reports that no
	// fast-forward is needed and leaves HEAD where it is.
	root := lmfResyncFixture(t, "")
	base := lmfHead(t, root)
	local := lmfLocalCommit(t, root, "local.txt", "local\n")
	fcGit(t, root, "update-ref", "refs/remotes/origin/main", base)
	report, err := runLocalMainResync(root)
	if err != nil {
		t.Fatalf("a local main that contains origin/main must be a no-op, not an error: %v", err)
	}
	if head := lmfHead(t, root); head != local {
		t.Fatalf("the no-op must leave HEAD at %s, got %s", local, head)
	}
	if !strings.Contains(strings.ToLower(report), "no fast-forward") {
		t.Fatalf("the no-op must report that no fast-forward is needed: %q", report)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the no-op must release the window: %+v", lock)
	}
}

func TestLocalMainResyncPreservesIgnoredFile(t *testing.T) {
	// The primary's ignored data.txt sits at a path that origin/main adds. The
	// fast-forward would overwrite it, so the re-sync refuses with guidance, HEAD
	// stays put, and the file keeps its content (plan §B3a step 5a).
	root := lmfResyncFixture(t, "data.txt")
	base := lmfHead(t, root)
	if err := os.WriteFile(filepath.Join(root, "data.txt"), []byte("local-keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lmfOriginCommit(t, root, base, "origin-ignored", map[string]string{"data.txt": "card:data.txt\n"}, true)
	_, err := runLocalMainResync(root)
	if err == nil {
		t.Fatalf("an ignored file at a path the fast-forward would write must refuse the re-sync")
	}
	// Plan §B3a names no class for this refusal. The nearest existing class is
	// MergeExitCollision, the merge step's refusal for an ignored-byte overwrite.
	if code, ok := factory.MergeExitCode(err); !ok || code != factory.MergeExitCollision {
		t.Fatalf("the ignored-file refusal must carry MergeExitCollision (%d), got code %d (ok=%v): %v", factory.MergeExitCollision, code, ok, err)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "data.txt") {
		t.Fatalf("the refusal must name the ignored path data.txt: %v", err)
	}
	if head := lmfHead(t, root); head != base {
		t.Fatalf("an ignored-file refusal must not move HEAD: %s -> %s", base, head)
	}
	lmfAssertFile(t, filepath.Join(root, "data.txt"), "local-keep\n")
	if lock := sdWindow(t, root); lock.Held() {
		t.Fatalf("the refusal must release the window: %+v", lock)
	}
}
