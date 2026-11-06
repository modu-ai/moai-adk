package cli

// update_migration_route_test.go — card t1640 M8S, finding F1: the routine
// update's worktree consolidation is routed through the shared migration
// engine (internal/cli/worktree). This file pins the update-side WIRING:
// the adapter is reachable through runUpdateWorktreeMigration, the AC-WHR-019
// grep cell reads green, and a moved update-path fixture is registered at
// its home-root destination (acceptance.md's own second-cell requirement).
// The engine behaviors themselves are pinned in internal/cli/worktree's
// migrate_sync_fix_test.go.

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

// sfxCliWorld builds a minimal update fixture: a repo with one eligible
// legacy tree and one agent-* tree, an unrecorded home root.
func sfxCliWorld(t *testing.T) (repo, movedTree, agentTree, homeRoot string) {
	t.Helper()
	home := sfxReal(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	t.Setenv("HOME", home)
	repo = filepath.Join(sfxReal(t, t.TempDir()), "sfx-cli-repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	sfxGit(t, repo, "init", "-q", "--template=", "-b", "main")
	sfxGit(t, repo, "config", "user.email", "sfx-cli@example.com")
	sfxGit(t, repo, "config", "user.name", "SFX CLI")
	if werr := os.WriteFile(filepath.Join(repo, "README.md"), []byte("seed\n"), 0o644); werr != nil {
		t.Fatalf("seed: %v", werr)
	}
	sfxGit(t, repo, "add", "-A")
	sfxGit(t, repo, "commit", "-q", "-m", "seed")
	remote := filepath.Join(filepath.Dir(repo), "sfx-cli-remote.git")
	sfxGit(t, repo, "init", "-q", "--bare", remote)
	sfxGit(t, repo, "remote", "add", "origin", remote)
	sfxGit(t, repo, "push", "-q", "-u", "origin", "main")

	movedTree = filepath.Join(repo, ".moai", "worktrees", "sfx-legacy")
	if err := os.MkdirAll(filepath.Dir(movedTree), 0o755); err != nil {
		t.Fatalf("mkdir legacy parent: %v", err)
	}
	sfxGit(t, repo, "worktree", "add", "-q", "-b", "WT-sfx-legacy", "--", movedTree)
	agentTree = filepath.Join(repo, ".claude", "worktrees", "agent-sfx")
	if err := os.MkdirAll(filepath.Dir(agentTree), 0o755); err != nil {
		t.Fatalf("mkdir agent parent: %v", err)
	}
	sfxGit(t, repo, "worktree", "add", "-q", "-b", "worktree-agent-sfx", "--", agentTree)

	res, err := worktreeroot.Resolve(repo)
	if err != nil {
		t.Fatalf("resolve home root: %v", err)
	}
	return repo, movedTree, agentTree, res.Root
}

func sfxReal(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", p, err)
	}
	return r
}

func sfxGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

// TestUpdateMigrationRoute_LandsAtHomeRoot — F1 through the update side:
// runUpdateWorktreeMigration moves the eligible legacy tree to its
// home-root destination, never touches the agent-* tree, and the move list
// carries the update trigger.
func TestUpdateMigrationRoute_LandsAtHomeRoot(t *testing.T) {
	repo, movedTree, agentTree, homeRoot := sfxCliWorld(t)

	buf := &strings.Builder{}
	if err := runUpdateWorktreeMigration(repo, false, buf); err != nil {
		t.Fatalf("update consolidation: %v\n%s", err, buf.String())
	}
	dest := filepath.Join(homeRoot, "sfx-legacy")
	if fi, statErr := os.Lstat(dest); statErr != nil || !fi.IsDir() {
		t.Errorf("the update did not move sfx-legacy to the home root %s (stat err=%v)", dest, statErr)
	}
	if sfxRegistered(t, repo, movedTree) {
		t.Errorf("sfx-legacy is still registered at the legacy path %s", movedTree)
	}
	if !sfxRegistered(t, repo, agentTree) {
		t.Errorf("the update moved the agent-* tree %s", agentTree)
	}
	mlPath := filepath.Join(repo, ".moai", "state", "worktree-migrate", "move-list.jsonl")
	raw, rerr := os.ReadFile(mlPath)
	if rerr != nil {
		t.Fatalf("read move list: %v", rerr)
	}
	sawMoved := false
	for _, line := range strings.Split(string(raw), "\n") {
		var rec struct {
			Result  string `json:"result"`
			Trigger string `json:"trigger"`
		}
		if jerr := json.Unmarshal([]byte(line), &rec); jerr == nil && rec.Result == "moved" {
			if rec.Trigger != "update" {
				t.Errorf("moved line trigger = %q, want update:\n%s", rec.Trigger, line)
			}
			sawMoved = true
		}
	}
	if !sawMoved {
		t.Errorf("no moved line in the move list:\n%s", raw)
	}
}

func sfxRegistered(t *testing.T, repo, path string) bool {
	t.Helper()
	cmd := exec.Command("git", "-C", repo, "worktree", "list", "--porcelain")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("worktree list: %v\n%s", err, out)
	}
	return strings.Contains(string(out), path)
}

// TestUpdateGrepCell_Green — AC-WHR-019's recorded first cell: the legacy
// in-repo destination expression is gone from the update path (expected
// green: grep exit 1, stdout 0).
func TestUpdateGrepCell_Green(t *testing.T) {
	cmd := exec.Command("grep", "-c", "Join(root, sessionWorktreeSubdir)", "update_worktree_migration.go")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("the legacy destination expression is still present (grep exit 0): %s", out)
	}
	if strings.TrimSpace(string(out)) != "0" {
		t.Errorf("grep count = %q, want 0", strings.TrimSpace(string(out)))
	}
}
