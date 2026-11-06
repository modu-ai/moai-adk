package cli

// update_worktree_migration_test.go — the update-side consolidation
// (SPEC-WORKTREE-HOME-ROOT-001 REQ-WHR-019; rewritten by card t1640 M8S,
// review finding F1: the consolidation is the shared migration engine's).
// The engine's own behaviors (skip causes, the manifest comparison, the
// reverse order) are pinned in internal/cli/worktree's suites; this file
// pins the update-side contract: the route is reachable through
// runUpdateWorktreeMigration, eligible legacy trees land at the home-root
// destination (never the legacy in-repo root), agent-* trees are never
// selected, the dry run moves nothing, a retry moves nothing, and the
// output carries the per-tree lines the update UI prints.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/lockfile"
	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

func migrationGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// migrationFixture builds the update rehearsal: a repo with one eligible
// legacy tree (.moai/worktrees/card on a WT- branch) and one protected
// agent-* tree (.claude/worktrees/agent-fix).
func migrationFixture(t *testing.T) (root, movedTree, agentTree, homeRoot string) {
	t.Helper()
	home := migrationReal(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	t.Setenv("HOME", home)
	root = filepath.Join(migrationReal(t, t.TempDir()), "project")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	migrationGit(t, root, "init", "-q", "-b", "main")
	migrationGit(t, root, "config", "user.name", "Migration Test")
	migrationGit(t, root, "config", "user.email", "migration-test@example.com")
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("ignored.txt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tracked.txt"), []byte("initial\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	migrationGit(t, root, "add", ".gitignore", "tracked.txt")
	migrationGit(t, root, "commit", "-q", "-m", "seed")
	remote := filepath.Join(filepath.Dir(root), "fixture-remote.git")
	migrationGit(t, root, "init", "-q", "--bare", remote)
	migrationGit(t, root, "remote", "add", "origin", remote)
	migrationGit(t, root, "push", "-q", "-u", "origin", "main")
	movedTree = filepath.Join(root, ".moai", "worktrees", "card")
	if err := os.MkdirAll(filepath.Dir(movedTree), 0o755); err != nil {
		t.Fatal(err)
	}
	migrationGit(t, root, "worktree", "add", "-q", "-b", "WT-migrate", movedTree)
	agentTree = filepath.Join(root, ".claude", "worktrees", "agent-fix")
	if err := os.MkdirAll(filepath.Dir(agentTree), 0o755); err != nil {
		t.Fatal(err)
	}
	migrationGit(t, root, "worktree", "add", "-q", "-b", "worktree-agent-fix", agentTree)
	res, err := worktreeroot.Resolve(root)
	if err != nil {
		t.Fatalf("resolve home root: %v", err)
	}
	return root, movedTree, agentTree, res.Root
}

func migrationReal(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", p, err)
	}
	return r
}

// TestUpdateConsolidation_MovesToHomeRoot — the update moves the eligible
// legacy tree to its home-root destination, never creates the legacy
// in-repo destination, and never selects the agent-* tree.
func TestUpdateConsolidation_MovesToHomeRoot(t *testing.T) {
	root, movedTree, agentTree, homeRoot := migrationFixture(t)
	var output bytes.Buffer
	if err := runUpdateWorktreeMigration(root, false, &output); err != nil {
		t.Fatalf("update consolidation: %v\n%s", err, output.String())
	}
	dest := filepath.Join(homeRoot, "card")
	if fi, statErr := os.Lstat(dest); statErr != nil || !fi.IsDir() {
		t.Errorf("the update did not move card to the home root %s (stat err=%v)", dest, statErr)
	}
	if _, statErr := os.Lstat(filepath.Join(root, ".moai", "worktrees", "card")); !os.IsNotExist(statErr) {
		t.Errorf("a tree exists at the legacy in-repo destination (stat err=%v)", statErr)
	}
	if !migrationRegistered(t, root, agentTree) {
		t.Errorf("the update moved the agent-* tree %s; the classifier never selects it", agentTree)
	}
	if !migrationRegistered(t, root, movedTree) {
		// The old path must be gone from the registration after the move.
	} else {
		t.Errorf("card is still registered at its legacy path %s", movedTree)
	}
	if got := migrationGit(t, dest, "branch", "--show-current"); got != "WT-migrate" {
		t.Errorf("moved branch = %q, want WT-migrate", got)
	}
	if !strings.Contains(output.String(), "Worktree migration: 1 moved, 1 skipped, 0 failed") {
		t.Errorf("the consolidation summary is missing:\n%s", output.String())
	}
}

func migrationRegistered(t *testing.T, root, path string) bool {
	t.Helper()
	out := migrationGit(t, root, "worktree", "list", "--porcelain")
	return strings.Contains(out, path)
}

// TestUpdateConsolidation_DryRunMovesNothing — AC-WHR-019(h): the preview
// prints the plan and moves nothing.
func TestUpdateConsolidation_DryRunMovesNothing(t *testing.T) {
	root, movedTree, _, homeRoot := migrationFixture(t)
	var output bytes.Buffer
	if err := runUpdateWorktreeMigration(root, true, &output); err != nil {
		t.Fatalf("update dry run: %v\n%s", err, output.String())
	}
	if !strings.Contains(output.String(), "Worktree migration planned") {
		t.Fatalf("dry run did not report the plan:\n%s", output.String())
	}
	if _, err := os.Stat(movedTree); err != nil {
		t.Fatalf("dry run moved the source: %v", err)
	}
	if _, err := os.Stat(filepath.Join(homeRoot, "card")); !os.IsNotExist(err) {
		t.Fatalf("dry run created a destination (stat err=%v)", err)
	}
}

// TestUpdateConsolidation_RetryMovesNothing — AC-WHR-019(i): the second
// run moves nothing further and reports no rows.
func TestUpdateConsolidation_RetryMovesNothing(t *testing.T) {
	root, _, _, _ := migrationFixture(t)
	var output bytes.Buffer
	if err := runUpdateWorktreeMigration(root, false, &output); err != nil {
		t.Fatalf("first run: %v\n%s", err, output.String())
	}
	output.Reset()
	if err := runUpdateWorktreeMigration(root, false, &output); err != nil {
		t.Fatalf("retry: %v\n%s", err, output.String())
	}
	if strings.Contains(output.String(), "Worktree migration moved") {
		t.Errorf("the retry moved something:\n%s", output.String())
	}
	if !strings.Contains(output.String(), "0 moved") {
		t.Errorf("the retry summary does not read 0 moved:\n%s", output.String())
	}
}

// TestUpdateConsolidation_SkipsAnchoredSession — the live-anchored-session
// skip cause rides the update path (AC-WHR-019(d)); the anchored tree
// stays at its old path.
func TestUpdateConsolidation_SkipsAnchoredSession(t *testing.T) {
	root, movedTree, _, _ := migrationFixture(t)
	registry := session.NewRegistry(filepath.Join(root, session.DefaultRegistryPath), nil)
	if err := registry.Register("active-session", session.SpecIDNone, session.PhaseNone); err != nil {
		t.Fatal(err)
	}
	if err := registry.RelocateSession("active-session", movedTree); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := runUpdateWorktreeMigration(root, false, &output); err != nil {
		t.Fatalf("update: %v\n%s", err, output.String())
	}
	if !strings.Contains(output.String(), "active session is anchored in this worktree") || !strings.Contains(output.String(), "2 skipped") {
		t.Errorf("the anchored tree's skip is not reported:\n%s", output.String())
	}
	if _, err := os.Stat(movedTree); err != nil {
		t.Fatalf("the anchored source was moved: %v", err)
	}
}

// TestUpdateConsolidation_HeldLockDegradesToError — AC-WHR-019(j): a held
// migration lock fails the consolidation (the caller degrades it to a
// warning), moves nothing, and the error names the lock.
func TestUpdateConsolidation_HeldLockDegradesToError(t *testing.T) {
	root, movedTree, _, _ := migrationFixture(t)
	lockPath := filepath.Join(root, ".moai", "state", "worktree-migrate", "lock")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatal(err)
	}
	holder, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = holder.Close() }()
	// The test actually TAKES the lock (an open descriptor alone holds
	// nothing): the same primitive the engine's withMigrationLock uses.
	if lerr := lockfile.TryLock(holder); lerr != nil {
		t.Fatalf("take the migration lock: %v", lerr)
	}
	defer func() { _ = lockfile.Unlock(holder) }()
	if err := runUpdateWorktreeMigration(root, false, &bytes.Buffer{}); err == nil {
		t.Errorf("the consolidation ran while the migration lock was held")
	}
	if _, statErr := os.Stat(movedTree); statErr != nil {
		t.Errorf("the tree moved although the lock was held (stat err=%v)", statErr)
	}
}
