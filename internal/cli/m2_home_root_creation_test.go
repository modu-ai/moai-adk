package cli

// m2_home_root_creation_test.go — SPEC-WORKTREE-HOME-ROOT-001 M2 RED (card
// t1640), branch Y of D-5 having been selected by the recorded PM-3 FAIL.
//
// These tests are written before the destination swap: the shared
// materializer (materializeSessionWorktree) and the launcher's --branch path
// (launcherWorktreeMaterializeReal) still build their destinations under the
// in-repo roots, so they fail — that failure is the RED (plan §C,
// test compilation order; the tests name M1 and existing symbols only, so
// the red is behavioral, per research §R13's M2 row).
//
// Acceptance mapping: AC-WHR-005a (shared materializer registers trees at
// the home root with the WT- convention and no legacy destination; the
// count-first static scan), AC-WHR-005b (launcher -w <n> --branch at the
// home root, branch and HEAD preserved, no branch created or renamed,
// registered through the state file, idempotent re-entry, missing branch
// refused).
//
// Fixtures are scratch repositories under t.TempDir() with their own
// MOAI_HOME; no test touches this repository's trees. The global
// safe.directory mutation seam is stubbed to a recorder so the suite never
// writes global git config.

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// ---------------------------------------------------------------- fixtures

// m2Real resolves symlinks (the macOS /var alias) so paths derived from git
// agree with the paths asserted here.
func m2Real(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", p, err)
	}
	return r
}

// m2Git runs one git command in dir and fails the test on error.
func m2Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// m2Repo creates a scratch repository with one commit and returns its
// symlink-resolved root.
func m2Repo(t *testing.T) string {
	t.Helper()
	root := filepath.Join(m2Real(t, t.TempDir()), "repo")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	m2Git(t, root, "init", "-q", "--template=", "-b", "main")
	m2Git(t, root, "config", "user.email", "m2-test@example.com")
	m2Git(t, root, "config", "user.name", "M2 Test")
	m2Git(t, root, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m2Git(t, root, "add", "README.md")
	m2Git(t, root, "commit", "-q", "-m", "seed")
	return root
}

// m2InRepo runs the body with the process cwd inside repo (the materializer
// resolves the repository from cwd through gitCommonDirReal).
func m2InRepo(t *testing.T, repo string) {
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

// m2NoGlobalGitConfig stubs the one permitted global git-config mutation so
// the suite never writes the user's global config.
func m2NoGlobalGitConfig(t *testing.T) *[]string {
	t.Helper()
	var recorded []string
	orig := sessionWorktreeGitSafeDirAdd
	sessionWorktreeGitSafeDirAdd = func(path string) error {
		recorded = append(recorded, path)
		return nil
	}
	t.Cleanup(func() { sessionWorktreeGitSafeDirAdd = orig })
	return &recorded
}

// ------------------------------------------------- AC-WHR-005a materializer

// TestMaterializeSessionWorktree_CreatesAtHomeRoot pins the shared
// materializer's destination: <MoaiHome>/worktrees/projects/<key>/<name>,
// registered by git, on the requested WT- branch, with nothing under the
// in-repo .moai/worktrees (REQ-WHR-005a).
func TestMaterializeSessionWorktree_CreatesAtHomeRoot(t *testing.T) {
	repo := m2Repo(t)
	home := m2Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	m2InRepo(t, repo)
	recordedSafeDirs := m2NoGlobalGitConfig(t)

	var out bytes.Buffer
	got, err := materializeSessionWorktree("WT-m2a", &out)
	if err != nil {
		t.Fatalf("materializeSessionWorktree: %v\n%s", err, out.String())
	}
	want := filepath.Join(home, "worktrees", "projects", filepath.Base(repo), "WT-m2a")
	if got != want {
		t.Errorf("destination = %q, want the home root %q", got, want)
	}
	if _, lerr := os.Lstat(filepath.Join(repo, ".moai", "worktrees", "WT-m2a")); !os.IsNotExist(lerr) {
		t.Errorf("a tree was created under the in-repo .moai/worktrees (lstat err=%v)", lerr)
	}
	if listing := m2Git(t, repo, "worktree", "list", "--porcelain"); !strings.Contains(listing, want) {
		t.Errorf("the home-root tree is not registered by git:\n%s", listing)
	}
	if branch := m2Git(t, got, "branch", "--show-current"); branch != "WT-m2a" {
		t.Errorf("branch = %q, want WT-m2a", branch)
	}
	if len(*recordedSafeDirs) != 1 || (*recordedSafeDirs)[0] != want {
		t.Errorf("safe.directory registrations = %v, want exactly the home-root tree", *recordedSafeDirs)
	}
}

// TestWorktreeNew_PathReachesTheSharedMaterializer pins the moai worktree
// new wiring: the command's creator is the shared materializer, so the
// destination clauses above hold for it without a second creation path.
func TestWorktreeNew_PathReachesTheSharedMaterializer(t *testing.T) {
	repo := m2Repo(t)
	home := m2Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	m2InRepo(t, repo)
	m2NoGlobalGitConfig(t)

	if worktree.WorktreeCreator == nil {
		t.Fatal("worktree.WorktreeCreator is not wired; moai worktree new has no creation path")
	}
	got, err := worktree.WorktreeCreator("WT-m2new", io.Discard)
	if err != nil {
		t.Fatalf("WorktreeCreator: %v", err)
	}
	want := filepath.Join(home, "worktrees", "projects", filepath.Base(repo), "WT-m2new")
	if got != want {
		t.Errorf("moai worktree new created %q, want the home root %q", got, want)
	}
}

// -------------------------------------------------- AC-WHR-005b --branch path

// TestLauncherBranchMaterialize_HomeRootBranchAndHead pins the launcher's
// existing-branch path: the tree lands at the home root with the existing
// branch checked out at the HEAD it had, no branch is created or renamed,
// the state file registers the home-root path, a second run reuses the tree,
// and a missing branch is refused (REQ-WHR-005b). The shared materializer's
// new-branch guarantees are deliberately not asserted here.
func TestLauncherBranchMaterialize_HomeRootBranchAndHead(t *testing.T) {
	repo := m2Repo(t)
	home := m2Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)

	m2Git(t, repo, "branch", "existing-b")
	head := m2Git(t, repo, "rev-parse", "existing-b")
	branchesBefore := m2Git(t, repo, "branch", "--list")

	var notices bytes.Buffer
	if err := launcherWorktreeMaterializeReal(repo, "m2b", "existing-b", &notices); err != nil {
		t.Fatalf("launcherWorktreeMaterializeReal: %v", err)
	}
	want := filepath.Join(home, "worktrees", "projects", filepath.Base(repo), "m2b")
	if _, lerr := os.Lstat(want); lerr != nil {
		t.Fatalf("tree missing at the home root %s: %v", want, lerr)
	}
	if _, lerr := os.Lstat(filepath.Join(repo, ".claude", "worktrees", "m2b")); !os.IsNotExist(lerr) {
		t.Errorf("a tree was created under the legacy .claude/worktrees (lstat err=%v)", lerr)
	}
	if got := m2Git(t, want, "branch", "--show-current"); got != "existing-b" {
		t.Errorf("checked-out branch = %q, want existing-b", got)
	}
	if got := m2Git(t, want, "rev-parse", "HEAD"); got != head {
		t.Errorf("HEAD = %q, want the branch's pre-materialization %q", got, head)
	}
	if after := m2Git(t, repo, "branch", "--list"); after != branchesBefore {
		t.Errorf("branch list changed:\nbefore:\n%s\nafter:\n%s", branchesBefore, after)
	}
	state, err := os.ReadFile(filepath.Join(repo, ".moai", "state", "worktrees.json"))
	if err != nil {
		t.Fatalf("worktree state file: %v", err)
	}
	if !strings.Contains(string(state), want) {
		t.Errorf("state file does not register the home-root path %q:\n%s", want, state)
	}

	// Idempotent re-entry: the second run reports and reuses.
	var reuse bytes.Buffer
	if err := launcherWorktreeMaterializeReal(repo, "m2b", "existing-b", &reuse); err != nil {
		t.Fatalf("second launcherWorktreeMaterializeReal: %v", err)
	}
	if !strings.Contains(reuse.String(), "already exists") {
		t.Errorf("second run notice = %q, want the reuse report", reuse.String())
	}

	// A missing branch is still refused, and creates nothing.
	var refuse bytes.Buffer
	if err := launcherWorktreeMaterializeReal(repo, "m2c", "no-such-branch", &refuse); err == nil {
		t.Error("a missing branch was accepted; --branch checks out an existing branch and never creates one")
	}
	if _, lerr := os.Lstat(filepath.Join(home, "worktrees", "projects", filepath.Base(repo), "m2c")); !os.IsNotExist(lerr) {
		t.Errorf("the refused --branch created a tree (lstat err=%v)", lerr)
	}
}

// ------------------------------------------------ AC-WHR-005a static scan

// TestM2_NoLegacyDestinationConstructions is AC-WHR-005a's count-first
// static scan over the production creation code: no call site in the
// creation paths constructs a .claude/worktrees or .moai/worktrees
// destination, and the two destination functions take their paths from the
// resolver. The branch-Y pass-through creates nothing, so it contributes no
// site (REQ-WHR-006 branch Y, REQ-WHR-020).
func TestM2_NoLegacyDestinationConstructions(t *testing.T) {
	files := []string{"session_worktree.go", "worktree_branch_flag.go", filepath.Join("worktree", "new.go")}
	t.Logf("count-first expected-site list: zero legacy destination sites in %v (branch Y creates nothing)", files)

	pairs := []string{`".claude", "worktrees"`, `".moai", "worktrees"`}
	for _, f := range files {
		body, err := os.ReadFile(filepath.Join(f))
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		src := string(body)
		for _, pair := range pairs {
			if n := strings.Count(src, pair); n != 0 {
				t.Errorf("%s constructs %s %d time(s); creation paths name no legacy root", f, pair, n)
			}
		}
	}

	// The two destination functions read their paths from the resolver.
	for _, check := range []struct{ file, fn, mustContain, mustNotContain string }{
		{"session_worktree.go", "materializeSessionWorktree", "worktreeroot.PathFor", "sessionWorktreeSubdir"},
		{"worktree_branch_flag.go", "launcherWorktreeMaterializeReal", "worktreeroot.PathFor", `".claude", "worktrees"`},
	} {
		src, err := os.ReadFile(filepath.Join(check.file))
		if err != nil {
			t.Fatalf("read %s: %v", check.file, err)
		}
		fnBody := m2FuncBody(t, string(src), check.fn)
		if !strings.Contains(fnBody, check.mustContain) {
			t.Errorf("%s.%s does not take its destination from worktreeroot.PathFor", check.file, check.fn)
		}
		if check.mustNotContain != "" && strings.Contains(fnBody, check.mustNotContain) {
			t.Errorf("%s.%s still names the legacy destination %s", check.file, check.fn, check.mustNotContain)
		}
	}
}

// m2FuncBody extracts the source text of one top-level function.
func m2FuncBody(t *testing.T, src, name string) string {
	t.Helper()
	start := strings.Index(src, "func "+name+"(")
	if start < 0 {
		t.Fatalf("function %s not found", name)
	}
	end := strings.Index(src[start:], "\nfunc ")
	if end < 0 {
		end = len(src) - start
	}
	return src[start : start+end]
}
