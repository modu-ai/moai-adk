package worktree

// list_test.go — SPEC-WORKTREE-HOME-ROOT-001 M4 RED (card t1640):
// AC-WHR-010, `moai worktree list [--json]`.
//
// Written before the verb exists: today `moai worktree list` prints the
// parent help text (the observed M0 cell), so every assertion below fails
// on the unregistered command — that failure is the RED (behavioral, per
// research §R13's M4 row: the tests name the M1 classifier's outputs and
// existing symbols only). The fixture carries one tree of each class, a
// controlled unpushed-commit count, a dirty tree, and the AC's
// two-spelling guarantee is asserted as row uniqueness over the whole
// output (no two rows denote one directory).

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

// m4Git runs one git command in dir and fails the test on error.
func m4Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// m4Real resolves symlinks (the macOS /var alias) so git-derived paths
// agree with the asserted ones.
func m4Real(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", p, err)
	}
	return r
}

// m4Row is one JSON row of the list output (REQ-WHR-010's nine fields).
type m4Row struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Root     string `json:"root"`
	Class    string `json:"class"`
	Branch   string `json:"branch"`
	Head     string `json:"head"`
	Dirty    bool   `json:"dirty"`
	Unpushed int    `json:"unpushed"`
	Anchored bool   `json:"anchored"`
}

// m4ListWorld builds the AC-WHR-010 fixture: one tree of each class, a
// controlled unpushed count, a dirty tree, and the integration branch in a
// dedicated tree (the primary checkout holds the working branch dev).
func m4ListWorld(t *testing.T) (repo string, expect []m4Row) {
	t.Helper()
	home := m4Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	t.Setenv("HOME", home) // the codex pool derives from the user home

	repo = filepath.Join(m4Real(t, t.TempDir()), "list-repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	git := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
		}
	}
	git(repo, "init", "-q", "--template=", "-b", "dev")
	git(repo, "config", "user.email", "list@example.com")
	git(repo, "config", "user.name", "List Test")
	git(repo, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	git(repo, "add", "-A")
	git(repo, "commit", "-q", "-m", "seed")
	git(repo, "branch", "main")
	cfgDir := filepath.Join(repo, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	cfg := "git_strategy:\n  mode: manual\n  manual:\n    workflow: git-flow\n    develop_branch: main\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "git-strategy.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}
	remote := filepath.Join(filepath.Dir(repo), "list-remote.git")
	git(repo, "init", "-q", "--bare", remote)
	git(repo, "remote", "add", "origin", remote)
	git(repo, "push", "-q", "-u", "origin", "dev")

	// Trees of each class.
	homeRoot, err := worktreeroot.Root(repo)
	if err != nil {
		t.Fatalf("resolve home root: %v", err)
	}
	homeTree := filepath.Join(homeRoot, "home4")
	lmTree := filepath.Join(repo, ".moai", "worktrees", "lm4")
	relTree := filepath.Join(repo, ".moai", "worktrees", "release-v314")
	claudeTree := filepath.Join(repo, ".claude", "worktrees", "t1497m")
	agentTree := filepath.Join(repo, ".claude", "worktrees", "agent-a1b2c3d4")
	integTree := filepath.Join(repo, ".claude", "worktrees", "main")
	codexRoot := filepath.Join(home, ".codex", "worktrees", "ab12cd4")
	codexTree := filepath.Join(codexRoot, "repo")
	for _, d := range []string{filepath.Dir(homeTree), filepath.Dir(lmTree), filepath.Dir(relTree), filepath.Dir(claudeTree), filepath.Dir(agentTree), filepath.Dir(integTree), codexRoot} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	git(repo, "worktree", "add", "-q", "-b", "WT-home4", "--", homeTree)
	git(repo, "worktree", "add", "-q", "-b", "WT-lm4", "--", lmTree)
	git(repo, "worktree", "add", "-q", "-b", "release/v3.1.4", "--", relTree)
	git(repo, "worktree", "add", "-q", "-b", "WT-t1497m", "--", claudeTree)
	git(repo, "worktree", "add", "-q", "-b", "worktree-agent-a1b2c3d4", "--", agentTree)
	git(repo, "worktree", "add", "-q", integTree, "main")
	git(repo, "worktree", "add", "-q", "-b", "WT-cx4", "--", codexTree)

	// A controlled unpushed count and a dirty tree.
	m4Git(t, lmTree, "commit", "-q", "--allow-empty", "-m", "unpushed")
	if err := os.WriteFile(filepath.Join(homeTree, "dirty.txt"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatalf("write dirty: %v", err)
	}

	expect = []m4Row{
		// The primary row is dirty: the fixture writes its untracked
		// git-strategy.yaml into the checkout.
		{Name: "list-repo", Path: m4Real(t, repo), Root: "other", Class: "primary", Branch: "dev", Dirty: true},
		{Name: "home4", Path: homeTree, Root: "home", Class: "moai-made", Branch: "WT-home4", Dirty: true},
		{Name: "lm4", Path: lmTree, Root: "legacy-moai", Class: "moai-made", Branch: "WT-lm4", Unpushed: 1},
		{Name: "release-v314", Path: relTree, Root: "legacy-moai", Class: "moai-made", Branch: "release/v3.1.4"},
		{Name: "t1497m", Path: claudeTree, Root: "legacy-claude", Class: "moai-made", Branch: "WT-t1497m"},
		{Name: "agent-a1b2c3d4", Path: agentTree, Root: "legacy-claude", Class: "claude-native", Branch: "worktree-agent-a1b2c3d4"},
		{Name: "main", Path: integTree, Root: "legacy-claude", Class: "integration", Branch: "main"},
		{Name: "repo", Path: codexTree, Root: "codex-managed", Class: "codex-managed", Branch: "WT-cx4"},
	}
	for i := range expect {
		expect[i].Path = m4Real(t, expect[i].Path)
		if expect[i].Branch != "" {
			expect[i].Head = m4Git(t, expect[i].Path, "rev-parse", "HEAD")[:7]
		}
	}
	return repo, expect
}

// TestWorktreeList_IsRegistered pins the registration clause: the list
// verb is part of the worktree command tree.
func TestWorktreeList_IsRegistered(t *testing.T) {
	if _, _, err := WorktreeCmd.Find([]string{"list"}); err != nil {
		t.Errorf("the list verb is not registered: %v", err)
	}
}

// m4RunList invokes `moai worktree list` with args on a fresh command
// instance. Execute on a registered child delegates to the root and
// re-parses os.Args (the go-test flags), so the harness builds an
// unattached instance and scopes SetArgs to this invocation.
func m4RunList(t *testing.T, args ...string) (string, error) {
	t.Helper()
	listCmd := newListCmd()
	buf := &strings.Builder{}
	listCmd.SetOut(buf)
	listCmd.SetErr(buf)
	listCmd.SetArgs(args)
	if err := listCmd.Execute(); err != nil {
		return buf.String(), err
	}
	return buf.String(), nil
}

// TestWorktreeList_JSON_ReportsEachClassOnce pins AC-WHR-010: exit 0, a JSON
// array with exactly one row per tree — each of the nine fields present —
// the two spellings of one directory collapsing to one row (row uniqueness
// over the whole output), and the run mutating nothing.
func TestWorktreeList_JSON_ReportsEachClassOnce(t *testing.T) {
	repo, expect := m4ListWorld(t)
	m4InRepo(t, repo)

	out, err := m4RunList(t, "--json")
	if err != nil {
		t.Fatalf("moai worktree list --json: %v\n%s", err, out)
	}
	var rows []m4Row
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		t.Fatalf("stdout is not a JSON array of rows: %v\n%s", err, out)
	}
	if len(rows) != len(expect) {
		t.Fatalf("list reported %d rows, want %d:\n%s", len(rows), len(expect), out)
	}
	byPath := map[string]m4Row{}
	for _, r := range rows {
		if prev, dup := byPath[r.Path]; dup {
			t.Errorf("two rows denote one directory: %+v and %+v", prev, r)
		}
		byPath[r.Path] = r
	}
	for _, want := range expect {
		got, ok := byPath[want.Path]
		if !ok {
			t.Errorf("no row for %s in:\n%s", want.Path, out)
			continue
		}
		if got.Name != want.Name || got.Root != want.Root || got.Class != want.Class || got.Branch != want.Branch {
			t.Errorf("row %s = name %q root %q class %q branch %q, want name %q root %q class %q branch %q",
				want.Path, got.Name, got.Root, got.Class, got.Branch, want.Name, want.Root, want.Class, want.Branch)
		}
		if got.Head != want.Head {
			t.Errorf("row %s short HEAD = %q, want %q", want.Path, got.Head, want.Head)
		}
		if got.Dirty != want.Dirty {
			t.Errorf("row %s dirty = %v, want %v", want.Path, got.Dirty, want.Dirty)
		}
		if got.Unpushed != want.Unpushed {
			t.Errorf("row %s unpushed = %d, want %d", want.Path, got.Unpushed, want.Unpushed)
		}
		if got.Anchored {
			t.Errorf("row %s anchored = true; the fixture anchors no session", want.Path)
		}
	}
}

// TestWorktreeList_TextFormNamesEveryTree pins the text form: exit 0 and
// every tree's path appears in the output.
func TestWorktreeList_TextFormNamesEveryTree(t *testing.T) {
	repo, expect := m4ListWorld(t)
	m4InRepo(t, repo)
	out, err := m4RunList(t)
	if err != nil {
		t.Fatalf("moai worktree list: %v\n%s", err, out)
	}
	for _, want := range expect {
		if !strings.Contains(out, want.Path) {
			t.Errorf("text output does not name %s:\n%s", want.Path, out)
		}
	}
}

// TestWorktreeList_MutatesNothing pins the read-only contract: the
// fixture's content digest is identical before and after the run.
func TestWorktreeList_MutatesNothing(t *testing.T) {
	repo, _ := m4ListWorld(t)
	m4InRepo(t, repo)
	before := m4Git(t, repo, "worktree", "list", "--porcelain")
	statusBefore := m4Git(t, repo, "status", "--porcelain")
	if _, err := m4RunList(t, "--json"); err != nil {
		t.Fatalf("list: %v", err)
	}
	if after := m4Git(t, repo, "worktree", "list", "--porcelain"); after != before {
		t.Errorf("the run changed the worktree registration:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if statusAfter := m4Git(t, repo, "status", "--porcelain"); statusAfter != statusBefore {
		t.Errorf("the run changed the working state:\nbefore:\n%s\nafter:\n%s", statusBefore, statusAfter)
	}
}

// m4InRepo runs the body's caller with cwd inside repo (the list resolves
// the repository from the working directory).
func m4InRepo(t *testing.T, repo string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("chdir %s: %v", repo, err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

// TestWorktreeList_OwnedHomeRootTreeWithoutWTPrefixIsMoaiMade — card t1640
// review P2d: witness (c) needs THIS repository's owner identity on the
// classify context; a home-root tree whose branch carries no WT- prefix is
// moai-made when the project-folder marker is ours (the post-migration
// shape of a moved non-WT- tree), not unclassified.
func TestWorktreeList_OwnedHomeRootTreeWithoutWTPrefixIsMoaiMade(t *testing.T) {
	repo, _ := m4ListWorld(t)
	m4InRepo(t, repo)

	homeRoot, err := worktreeroot.Root(repo)
	if err != nil {
		t.Fatalf("resolve home root: %v", err)
	}
	owned := filepath.Join(homeRoot, "release4")
	if err := os.MkdirAll(filepath.Dir(owned), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	m4Git(t, repo, "worktree", "add", "-q", "-b", "release/v3.0.9", "--", owned)

	out, err := m4RunList(t, "--json")
	if err != nil {
		t.Fatalf("list --json: %v\n%s", err, out)
	}
	var rows []m4Row
	if jerr := json.Unmarshal([]byte(out), &rows); jerr != nil {
		t.Fatalf("parse the list rows: %v\n%s", jerr, out)
	}
	for _, r := range rows {
		if r.Name == "release4" {
			if r.Class != "moai-made" {
				t.Errorf("the owned home-root tree classifies as %q, want moai-made (row: %+v)", r.Class, r)
			}
			return
		}
	}
	t.Fatalf("the list carries no row for the owned home-root tree:\n%s", out)
}
