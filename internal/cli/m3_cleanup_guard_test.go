package cli

// m3_cleanup_guard_test.go — SPEC-WORKTREE-HOME-ROOT-001 M3 RED (card
// t1640): AC-WHR-014b, the launcher auto-cleanup guard.
//
// Nine registered worker-* trees in the measured states; after
// cleanupMoaiWorktrees runs, exactly worker-landed and worker-premove are
// removed (today's behavior, pinned) and the other seven are kept with a
// reason. Written before the guard exists: today every clean worker-*
// tree is removed, so the seven keep-assertions fail — that failure is the
// RED. The guard consumes the M1 MovedBy witness and the disposal verbs'
// landing predicate (LandedByPatchID / LandedByCommitPatchIDs).
//
// The remote check for worker-unchecked must FAIL, which needs a broken
// remote, while worker-landed needs a working one — the two states cannot
// share one repository's remote config, so the table spans two fixtures
// and the cleanup runs once per fixture.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// m3Gitf runs one git command in dir and fails the test on error.
func m3Gitf(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// m3SeedRepo creates a scratch repository on main with a bare remote and
// returns its root.
func m3SeedRepo(t *testing.T, name string) string {
	t.Helper()
	repo := filepath.Join(m2Real(t, t.TempDir()), name)
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	m3Gitf(t, repo, "init", "-q", "--template=", "-b", "main")
	m3Gitf(t, repo, "config", "user.email", "m3@example.com")
	m3Gitf(t, repo, "config", "user.name", "M3 Test")
	m3Gitf(t, repo, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m3Gitf(t, repo, "add", "README.md")
	m3Gitf(t, repo, "commit", "-q", "-m", "seed")
	// The landing predicate reads the configured integration target from
	// the repository's git-strategy.yaml; the fixture configures main.
	cfgDir := filepath.Join(repo, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	cfg := "git_strategy:\n  mode: manual\n  manual:\n    workflow: git-flow\n    develop_branch: main\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "git-strategy.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}
	remote := filepath.Join(filepath.Dir(repo), name+"-remote.git")
	m3Gitf(t, repo, "init", "-q", "--bare", remote)
	m3Gitf(t, repo, "remote", "add", "origin", remote)
	m3Gitf(t, repo, "push", "-q", "-u", "origin", "main")
	return repo
}

// m3Worker creates a registered worker tree on its own branch at path and
// returns the branch name.
func m3Worker(t *testing.T, repo, path string) string {
	t.Helper()
	branch := "worker-" + filepath.Base(path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir worker parent: %v", err)
	}
	m3Gitf(t, repo, "worktree", "add", "-q", "-b", branch, "--", path)
	return branch
}

// m3CommitInTree commits one file inside the tree and returns the new tip.
func m3CommitInTree(t *testing.T, tree, msg string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(tree, msg+".txt"), []byte(msg+"\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	m3Gitf(t, tree, "add", "-A")
	m3Gitf(t, tree, "commit", "-q", "-m", msg)
	return m3Gitf(t, tree, "rev-parse", "HEAD")
}

// m3AppendMoveLine appends one hand-written move-list record to repo's
// move list (the shape the M6 writer will emit; the M1 reader consumes it).
func m3AppendMoveLine(t *testing.T, repo, result, oldPath, newPath, ident string) {
	t.Helper()
	listPath := filepath.Join(repo, ".moai", "state", "worktree-migrate", "move-list.jsonl")
	if err := os.MkdirAll(filepath.Dir(listPath), 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	line := fmt.Sprintf(`{"schema":1,"ts":"2026-10-20T00:00:00Z","run":"r","trigger":"migrate","tree":"w","card":"w","old_path":%q,"new_path":%q,"result":%q,"reason":"","ident":%q}`+"\n", oldPath, newPath, result, ident)
	fh, err := os.OpenFile(listPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open move list: %v", err)
	}
	defer fh.Close()
	if _, err := fh.WriteString(line); err != nil {
		t.Fatalf("append move list: %v", err)
	}
}

// m3Registered reports whether git still registers path.
func m3Registered(t *testing.T, repo, path string) bool {
	t.Helper()
	return strings.Contains(m3Gitf(t, repo, "worktree", "list", "--porcelain"), path)
}

// TestM3CleanupGuard_RemovesOnlyIntegratedLandedUnmoved is AC-WHR-014b's
// nine-tree table: the launcher cleanup keeps every tree whose branch is
// not landed on a remote or which the migration moved, keeps it with a
// reason, and keeps today's removal only for an integrated, landed,
// unmoved tree.
func TestM3CleanupGuard_RemovesOnlyIntegratedLandedUnmoved(t *testing.T) {
	home := m2Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	repo := m3SeedRepo(t, "guarded")
	key := filepath.Base(repo)
	homeRoot := filepath.Join(home, "worktrees", "projects", key)

	// --- repo A: a working remote ---
	// worker-review: home root, one commit on no remote, clean.
	m3Worker(t, repo, filepath.Join(homeRoot, "worker-review"))
	m3CommitInTree(t, filepath.Join(homeRoot, "worker-review"), "review-note")
	// worker-moved: home root, integrated and landed, moved by the
	// migration (its own moved line), followed by a foreign skipped line
	// naming the same path.
	movedPath := filepath.Join(homeRoot, "worker-moved")
	movedBranch := m3Worker(t, repo, movedPath)
	m3CommitInTree(t, movedPath, "moved-work")
	m3Gitf(t, repo, "merge", "-q", "--no-ff", "-m", "integrate moved", movedBranch)
	m3Gitf(t, repo, "push", "-q", "origin", "main")
	id, err := worktree.EnsureInstanceToken(movedPath)
	if err != nil {
		t.Fatalf("token worker-moved: %v", err)
	}
	m3AppendMoveLine(t, repo, "moved", filepath.Join(homeRoot, "worker-moved-old"), movedPath, id.Ident())
	m3AppendMoveLine(t, repo, "skipped", filepath.Join(homeRoot, "worker-other"), movedPath, "")
	// worker-pending: home root, integrated and landed, its own latest
	// line a pending whose new_path is its path (kill after the move).
	pendingPath := filepath.Join(homeRoot, "worker-pending")
	pendingBranch := m3Worker(t, repo, pendingPath)
	m3CommitInTree(t, pendingPath, "pending-work")
	m3Gitf(t, repo, "merge", "-q", "--no-ff", "-m", "integrate pending", pendingBranch)
	m3Gitf(t, repo, "push", "-q", "origin", "main")
	pid, err := worktree.EnsureInstanceToken(pendingPath)
	if err != nil {
		t.Fatalf("token worker-pending: %v", err)
	}
	m3AppendMoveLine(t, repo, "pending", filepath.Join(homeRoot, "worker-pending-old"), pendingPath, pid.Ident())
	// worker-refused: home root, integrated and landed, its own latest
	// line a refused whose new_path is its path.
	refusedPath := filepath.Join(homeRoot, "worker-refused")
	refusedBranch := m3Worker(t, repo, refusedPath)
	m3CommitInTree(t, refusedPath, "refused-work")
	m3Gitf(t, repo, "merge", "-q", "--no-ff", "-m", "integrate refused", refusedBranch)
	m3Gitf(t, repo, "push", "-q", "origin", "main")
	rid, err := worktree.EnsureInstanceToken(refusedPath)
	if err != nil {
		t.Fatalf("token worker-refused: %v", err)
	}
	// D-11's refused orientation: the recorded source, then the destination
	// where the tree stands — new_path names the tree's current path.
	m3AppendMoveLine(t, repo, "refused", filepath.Join(homeRoot, "worker-refused-old"), refusedPath, rid.Ident())
	// worker-landed: legacy root, integrated AND landed, not moved —
	// today's removal is kept. Created BEFORE worker-integrated: its main
	// push must not carry worker-integrated's commit into origin/main.
	landedPath := filepath.Join(repo, ".claude", "worktrees", "worker-landed")
	landedBranch := m3Worker(t, repo, landedPath)
	m3CommitInTree(t, landedPath, "landed-work")
	m3Gitf(t, repo, "merge", "-q", "--no-ff", "-m", "integrate landed", landedBranch)
	m3Gitf(t, repo, "push", "-q", "origin", "main")
	// worker-pushed: legacy root, pushed to its own remote-tracking ref,
	// not integrated.
	pushedPath := filepath.Join(repo, ".claude", "worktrees", "worker-pushed")
	pushedBranch := m3Worker(t, repo, pushedPath)
	m3CommitInTree(t, pushedPath, "pushed-work")
	m3Gitf(t, repo, "push", "-q", "origin", pushedBranch)
	// worker-premove: legacy root, integrated and landed, its own latest
	// line a pending whose move was never made (the tree stands at the
	// old_path) — today's removal is kept. Its main push runs BEFORE
	// worker-integrated exists, so the later merge never lands.
	premovePath := filepath.Join(repo, ".claude", "worktrees", "worker-premove")
	premoveBranch := m3Worker(t, repo, premovePath)
	m3CommitInTree(t, premovePath, "premove-work")
	m3Gitf(t, repo, "merge", "-q", "--no-ff", "-m", "integrate premove", premoveBranch)
	m3Gitf(t, repo, "push", "-q", "origin", "main")
	premoveID, err := worktree.EnsureInstanceToken(premovePath)
	if err != nil {
		t.Fatalf("token worker-premove: %v", err)
	}
	m3AppendMoveLine(t, repo, "pending", premovePath, filepath.Join(homeRoot, "worker-premove-new"), premoveID.Ident())
	// worker-integrated: legacy root, integrated into main, present on no
	// remote — created and merged LAST, after every main push, so no push
	// lands its commit.
	integratedPath := filepath.Join(repo, ".claude", "worktrees", "worker-integrated")
	integratedBranch := m3Worker(t, repo, integratedPath)
	m3CommitInTree(t, integratedPath, "integrated-work")
	m3Gitf(t, repo, "merge", "-q", "--no-ff", "-m", "integrate", integratedBranch)

	report := cleanupMoaiWorktrees(repo)
	t.Logf("cleanup report: %s", report)

	// AC-WHR-014b's reporting clause: every kept tree is reported as
	// skipped, with its reason.
	for _, name := range []string{"worker-review", "worker-moved", "worker-pending", "worker-refused", "worker-pushed", "worker-integrated"} {
		if !strings.Contains(report, name) {
			t.Errorf("the cleanup report does not account for kept tree %s:\n%s", name, report)
		}
	}

	// The seven protected trees survive, registered, with their branches.
	for _, p := range []string{
		filepath.Join(homeRoot, "worker-review"),
		filepath.Join(homeRoot, "worker-moved"),
		filepath.Join(homeRoot, "worker-pending"),
		filepath.Join(homeRoot, "worker-refused"),
		pushedPath,
		integratedPath,
	} {
		if !m3Registered(t, repo, p) {
			t.Errorf("protected tree %s was removed; the cleanup guard must keep it", p)
		}
	}
	// The two removal clauses keep today's behavior.
	for _, p := range []string{landedPath, premovePath} {
		if m3Registered(t, repo, p) {
			t.Errorf("unprotected tree %s was kept; integrated+landed+unmoved worker trees are still removed", p)
		}
	}
	// Their branches survive in the branch list (removal never deletes a
	// branch).
	for _, b := range []string{landedBranch, premoveBranch, movedBranch, pendingBranch, refusedBranch, pushedBranch, integratedBranch} {
		if !strings.Contains(m3Gitf(t, repo, "branch", "--list"), b) {
			t.Errorf("branch %s vanished from the branch list", b)
		}
	}
}

// TestM3CleanupGuard_FailedRemoteCheckKeepsTree pins worker-unchecked in a
// fixture whose remote-tracking integration ref cannot be read: integrated,
// but the landing predicate's remote check fails, so the tree is kept.
func TestM3CleanupGuard_FailedRemoteCheckKeepsTree(t *testing.T) {
	home := m2Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	repo := m3SeedRepo(t, "unchecked")
	// Break the remote check: the integration ref cannot be read.
	m3Gitf(t, repo, "update-ref", "-d", "refs/remotes/origin/main")

	unchekedPath := filepath.Join(repo, ".claude", "worktrees", "worker-unchecked")
	branch := m3Worker(t, repo, unchekedPath)
	m3CommitInTree(t, unchekedPath, "unchecked-work")
	m3Gitf(t, repo, "merge", "-q", "--no-ff", "-m", "integrate unchecked", branch)

	cleanupMoaiWorktrees(repo)

	if !m3Registered(t, repo, unchekedPath) {
		t.Errorf("worker-unchecked was removed; a failed remote check is not a landing")
	}
}
