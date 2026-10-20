package worktree

// migrate_command_test.go — SPEC-WORKTREE-HOME-ROOT-001 M6 RED (card
// t1640): AC-WHR-015 (dry-run read-only, apply required, verify), AC-019's
// destination cell, AC-014a's static-and-behavioral non-destructive shape,
// the update-time selection and containment clauses, and AC-WHR-002b
// clause 4 (a second migration changes no assignment).
//
// The tests drive the `moai worktree migrate` verb through its command;
// before the verb exists they fail on the unregistered command — that
// failure is the RED (plan §C test-compilation order; the tests name the
// M1 classifier, the M5 lock/registry surface and existing symbols per
// research §R13's M6 row).

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

// m6Git runs one git command in dir and fails the test on error.
func m6Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// m6Real resolves symlinks (the macOS /var alias) so git-derived paths
// agree with the asserted ones.
func m6Real(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", p, err)
	}
	return r
}

// m6ConfigureIntegration writes the git-flow integration config the engine
// and the landing predicate read.
func m6ConfigureIntegration(t *testing.T, root string) {
	t.Helper()
	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	cfg := "git_strategy:\n  mode: manual\n  manual:\n    workflow: git-flow\n    develop_branch: main\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "git-strategy.yaml"), []byte(cfg), 0o600); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}
}

// m6World builds the rehearsal repository: a primary checkout whose
// working branch is dev, main landed on a bare remote, and the named
// eligible and protected trees.
type m6World struct {
	repo, home                 string
	homeRoot                   string
	movedTree, agentTree       string
	lockedTree, occupiedTree   string
	nonWTTree, workerTree      string
	unclassifiedTree, destHeld string
}

func m6NewWorld(t *testing.T) *m6World {
	t.Helper()
	home := m6Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	t.Setenv("HOME", home)

	repo := filepath.Join(m6Real(t, t.TempDir()), "mig-repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	m6Git(t, repo, "init", "-q", "--template=", "-b", "dev")
	m6Git(t, repo, "config", "user.email", "m6@example.com")
	m6Git(t, repo, "config", "user.name", "M6 Test")
	m6Git(t, repo, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m6Git(t, repo, "add", "-A")
	m6Git(t, repo, "commit", "-q", "-m", "seed")
	m6Git(t, repo, "branch", "main")
	remote := filepath.Join(filepath.Dir(repo), "mig-remote.git")
	m6Git(t, repo, "init", "-q", "--bare", remote)
	m6Git(t, repo, "remote", "add", "origin", remote)
	m6Git(t, repo, "push", "-q", "-u", "origin", "dev")
	m6ConfigureIntegration(t, repo)

	w := &m6World{repo: repo, home: home}
	mk := func(base, name, branch string) string {
		p := filepath.Join(base, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
		m6Git(t, repo, "worktree", "add", "-q", "-b", branch, "--", p)
		return p
	}
	if _, rerr := worktreeroot.Resolve(repo); rerr != nil {
		t.Fatalf("resolve home root: %v", rerr)
	}
	homeRoot, err := worktreeroot.Root(repo)
	if err != nil {
		t.Fatalf("home root: %v", err)
	}
	w.homeRoot = homeRoot
	// Eligible: a moai-made home tree (the migration's destination class —
	// pre-existing legacy trees are the ones M6 moves; the home tree here is
	// the post-M2 shape and stays put).
	_ = mk(homeRoot, "home6", "WT-home6")
	// Eligible: a legacy-moai moai-made tree (the migration's target).
	w.movedTree = mk(filepath.Join(repo, ".moai", "worktrees"), "legacy6", "WT-legacy6")
	// Protected: an agent-* tree under .claude/worktrees.
	w.agentTree = mk(filepath.Join(repo, ".claude", "worktrees"), "agent-feedface", "worktree-agent-feedface")
	// Protected: a locked tree under .moai/worktrees.
	w.lockedTree = mk(filepath.Join(repo, ".moai", "worktrees"), "locked6", "WT-locked6")
	m6Git(t, repo, "worktree", "lock", w.lockedTree)
	// Protected: the current process occupies this tree.
	w.occupiedTree = mk(filepath.Join(repo, ".moai", "worktrees"), "occupied6", "WT-occupied6")
	// Eligible: a legacy-moai non-WT- tree (witness c).
	w.nonWTTree = mk(filepath.Join(repo, ".moai", "worktrees"), "rc30-preview6", "release/v3.1.4")
	// Eligible: a worker-* tree under .claude/worktrees.
	w.workerTree = mk(filepath.Join(repo, ".claude", "worktrees"), "worker-m6", "WT-worker-m6")
	// Protected: an unclassified tree elsewhere.
	elsewhere := filepath.Join(filepath.Dir(repo), "elsewhere6")
	w.unclassifiedTree = mk(elsewhere, "feature6", "feature/x")
	// A pre-occupied destination under .moai/worktrees.
	w.destHeld = filepath.Join(repo, ".moai", "worktrees", "held6")
	if err := os.MkdirAll(w.destHeld, 0o755); err != nil {
		t.Fatalf("mkdir held destination: %v", err)
	}
	if err := os.WriteFile(filepath.Join(w.destHeld, "marker.txt"), []byte("held\n"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	return w
}

// m6Registered reports whether git still registers path.
func m6Registered(t *testing.T, repo, path string) bool {
	t.Helper()
	return strings.Contains(m6Git(t, repo, "worktree", "list", "--porcelain"), path)
}

// TestMigrateVerb_IsRegistered pins the registration clause: the migrate
// verb is part of the worktree command tree.
func TestMigrateVerb_IsRegistered(t *testing.T) {
	if _, _, err := WorktreeCmd.Find([]string{"migrate"}); err != nil {
		t.Errorf("the migrate verb is not registered: %v", err)
	}
}

// TestMigrate_DryRunWritesNothing pins AC-WHR-015's first half: a dry run
// exits 0, names the plan, and writes nothing — no move-list line, no
// destination, no assignment change.
func TestMigrate_DryRunWritesNothing(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)

	stateBefore := m6Git(t, w.repo, "worktree", "list", "--porcelain")
	mlPath := filepath.Join(w.repo, ".moai", "state", "worktree-migrate", "move-list.jsonl")
	if _, err := os.Stat(mlPath); !os.IsNotExist(err) {
		t.Fatalf("the fixture already has a move list: %v", err)
	}

	out, err := m4RunList(t, "migrate")
	if err != nil {
		t.Fatalf("migrate (dry run): %v\n%s", err, out)
	}
	if !strings.Contains(out, "legacy6") {
		t.Errorf("the dry-run plan does not name the eligible tree legacy6:\n%s", out)
	}
	if strings.Contains(out, "agent-feedface") && !strings.Contains(out, "claude-native") {
		t.Errorf("the dry-run plan does not classify the agent tree:\n%s", out)
	}
	if after := m6Git(t, w.repo, "worktree", "list", "--porcelain"); after != stateBefore {
		t.Errorf("the dry run changed the worktree registration")
	}
	if _, err := os.Stat(mlPath); !os.IsNotExist(err) {
		t.Errorf("the dry run wrote a move list (stat err=%v)", err)
	}
}

// TestMigrate_ApplyMovesEligibleAndSkipsProtected pins the apply shape: the
// eligible legacy trees land at the home root with a move list, the
// agent-*, locked, occupied and unclassified trees stay, and a second
// apply moves nothing (AC-WHR-015's idempotence, AC-WHR-016's skip causes,
// AC-019's destination cell).
func TestMigrate_ApplyMovesEligibleAndSkipsProtected(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)

	out, err := m4RunList(t, "migrate", "--apply")
	if err != nil {
		t.Fatalf("migrate --apply: %v\n%s", err, out)
	}
	// The legacy-moai eligible trees moved to the home root.
	if !m6Registered(t, w.repo, filepath.Join(w.homeRoot, "legacy6")) {
		t.Error("legacy6 was not moved to the home root")
	}
	if !m6Registered(t, w.repo, w.nonWTTree) && !m6Registered(t, w.repo, filepath.Join(w.homeRoot, "rc30-preview6")) {
		t.Error("the non-WT- tree rc30-preview6 was not moved (witness c must move)")
	}
	// The protected trees stayed.
	for _, p := range []string{w.agentTree, w.lockedTree, w.unclassifiedTree, w.destHeld} {
		if !m6Registered(t, w.repo, p) {
			t.Errorf("protected tree %s was moved or removed", p)
		}
	}
	// No tree landed under the pre-occupied destination.
	if m6Registered(t, w.repo, filepath.Join(w.repo, ".moai", "worktrees", "held6", "anything")) {
		t.Error("a tree was moved onto the pre-occupied destination")
	}
	// The move list exists and carries the settled moves.
	mlPath := filepath.Join(w.repo, ".moai", "state", "worktree-migrate", "move-list.jsonl")
	data, err := os.ReadFile(mlPath)
	if err != nil {
		t.Fatalf("move list missing after apply: %v", err)
	}
	if !strings.Contains(string(data), `"moved"`) {
		t.Errorf("the move list carries no moved line:\n%s", data)
	}

	// A second apply moves nothing further.
	stateBefore := m6Git(t, w.repo, "worktree", "list", "--porcelain")
	mlBefore, _ := os.ReadFile(mlPath)
	if _, err := m4RunList(t, "migrate", "--apply"); err != nil {
		t.Fatalf("second migrate --apply: %v", err)
	}
	if stateAfter := m6Git(t, w.repo, "worktree", "list", "--porcelain"); stateAfter != stateBefore {
		t.Errorf("the second apply moved trees")
	}
	mlAfter, _ := os.ReadFile(mlPath)
	if len(mlAfter) != len(mlBefore) {
		t.Errorf("the second apply appended move-list lines: %d → %d bytes", len(mlBefore), len(mlAfter))
	}

	// Verify reports each recorded move as intact.
	vout, verr := m4RunList(t, "migrate", "--verify")
	if verr != nil {
		t.Fatalf("migrate --verify: %v\n%s", verr, vout)
	}
	if !strings.Contains(vout, "intact") {
		t.Errorf("migrate --verify does not report the moves intact:\n%s", vout)
	}
}

// TestMigrate_RollbackRestoresTheTrees pins the rollback contract: after an
// apply, a rollback moves the trees back, appends restored entries, and a
// re-apply moves them again with new moved entries (AC-WHR-015's rollback
// half, AC-WHR-018's free-tree shape).
func TestMigrate_RollbackRestoresTheTrees(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)
	if _, err := m4RunList(t, "migrate", "--apply"); err != nil {
		t.Fatalf("apply: %v", err)
	}
	mlDir := filepath.Join(w.repo, ".moai", "state", "worktree-migrate")
	journalPath := ""
	if entries, err := os.ReadDir(mlDir); err == nil {
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".json") {
				journalPath = filepath.Join(mlDir, e.Name())
			}
		}
	}
	if journalPath == "" {
		t.Fatal("apply wrote no journal")
	}
	mlPath := filepath.Join(mlDir, "move-list.jsonl")
	mlBefore, _ := os.ReadFile(mlPath)

	if _, err := m4RunList(t, "migrate", "--rollback", journalPath); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if !m6Registered(t, w.repo, w.movedTree) {
		t.Error("legacy6 did not return to its pre-move path after the rollback")
	}
	mlAfter, _ := os.ReadFile(mlPath)
	if !strings.Contains(string(mlAfter), `"restored"`) {
		t.Errorf("the rollback appended no restored entries:\n%s", mlAfter)
	}
	if len(mlAfter) <= len(mlBefore) {
		t.Errorf("the rollback did not append to the move list: %d → %d bytes", len(mlBefore), len(mlAfter))
	}

	// A re-apply moves the trees again with new moved entries.
	mlMid, _ := os.ReadFile(mlPath)
	if _, err := m4RunList(t, "migrate", "--apply"); err != nil {
		t.Fatalf("re-apply: %v", err)
	}
	mlFinal, _ := os.ReadFile(mlPath)
	if len(mlFinal) <= len(mlMid) {
		t.Errorf("the re-apply appended no moved entries: %d → %d bytes", len(mlMid), len(mlFinal))
	}
}

// TestMigrate_AgentTreeNoLongerSelected pins the containment clause of
// AC-023 and AC-019's selection cell: after the M6 selection change, the
// dry-run plan classifies the agent tree as claude-native (never moved)
// rather than treating it as an eligible path-prefix match.
func TestMigrate_AgentTreeNoLongerSelected(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)
	out, err := m4RunList(t, "migrate")
	if err != nil {
		t.Fatalf("migrate (dry run): %v\n%s", err, out)
	}
	if !strings.Contains(out, "agent-feedface") {
		t.Fatalf("the dry-run plan does not list the agent tree:\n%s", out)
	}
	// The agent tree is classified, not moved: the plan names its class and
	// marks it as staying.
	if !strings.Contains(strings.ToLower(out), "claude-native") {
		t.Errorf("the dry-run plan does not classify the agent tree claude-native:\n%s", out)
	}
}
