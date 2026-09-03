package cli

// integration_target_test.go — card t449.
//
// The acquire record existed but named the wrong tree: it recorded the
// CALLER's cwd and the CALLER's checked-out branch, while the window it
// serializes is the integration worktree — the tree holding the configured
// integration branch. Observed four times by the lead as `branch: main` /
// `worktree: <primary checkout>` on a window whose whole purpose was
// serializing `.claude/worktrees/develop`. These tests pin the corrected
// record and the two seams behind it: the branch resolution order (explicit
// flag → configured git-flow develop branch → caller fallback) and the
// `git worktree list --porcelain` lookup.
//
// Every scratch repository uses a distinctive fixture branch name
// ("fixture-integration") so a hardcoded "develop" default cannot pass.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// gitRun runs one git command inside dir, with a fixed commit identity, and
// fails the test on any error. Output is returned for the rare caller that
// reads it.
func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// scratchRepo builds a two-worktree git repository in temp dirs: repo on its
// default branch (main), plus a second worktree with "fixture-integration"
// checked out. The caller's tree therefore differs from the integration tree,
// which is exactly the shape under test.
func scratchRepo(t *testing.T) (repo, integrationWT string) {
	t.Helper()
	repo = t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(repo, "seed.txt"), []byte("seed"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "add", "seed.txt")
	gitRun(t, repo, "commit", "-q", "-m", "seed")
	gitRun(t, repo, "branch", "fixture-integration")
	integrationWT = filepath.Join(t.TempDir(), "integration-tree")
	gitRun(t, repo, "worktree", "add", integrationWT, "fixture-integration")
	return repo, integrationWT
}

// chdirRepo moves the test process into dir for the duration of the test. The
// resolution seams run git against the process cwd, so the "caller's tree" of
// a scenario is whatever this helper points at.
func chdirRepo(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

// realPath normalizes a path for comparison: macOS temp dirs are reached
// through /var/... symlinks that git resolves to /private/var/... in its own
// output.
func realPath(t *testing.T, p string) string {
	t.Helper()
	rp, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("resolve %s: %v", p, err)
	}
	return rp
}

// writeGitStrategyFixture writes a git-strategy.yaml naming the given workflow
// and develop branch under projectRoot, so the acquire verb reads a real
// config file through the real reader rather than a stubbed one.
func writeGitStrategyFixture(t *testing.T, projectRoot, workflow, develop string) {
	t.Helper()
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf("git_strategy:\n    mode: manual\n    manual:\n        workflow: %s\n        develop_branch: %s\n", workflow, develop)
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// --- worktreeForBranch: the porcelain lookup ---

func TestWorktreeForBranch_FindsTheCheckedOutWorktree(t *testing.T) {
	repo, integrationWT := scratchRepo(t)
	chdirRepo(t, repo)

	if got := worktreeForBranch("fixture-integration"); got != realPath(t, integrationWT) {
		t.Errorf("worktreeForBranch(fixture-integration) = %q, want %q", got, realPath(t, integrationWT))
	}
}

func TestWorktreeForBranch_UnknownBranchIsEmpty(t *testing.T) {
	repo, _ := scratchRepo(t)
	chdirRepo(t, repo)

	if got := worktreeForBranch("no-such-branch"); got != "" {
		t.Errorf("worktreeForBranch(no-such-branch) = %q, want an empty unknown, not a wrong path", got)
	}
}

// --- resolveIntegrationTarget: the resolution order ---

func TestResolveIntegrationTarget_ExplicitFlagWins(t *testing.T) {
	repo, integrationWT := scratchRepo(t)
	chdirRepo(t, repo)

	branch, wt := resolveIntegrationTarget("fixture-integration", "some-other-configured-branch")
	if branch != "fixture-integration" {
		t.Errorf("branch = %q, want the explicit flag to win over the configured value", branch)
	}
	if wt != realPath(t, integrationWT) {
		t.Errorf("worktree = %q, want the explicit branch's worktree %q", wt, realPath(t, integrationWT))
	}
}

func TestResolveIntegrationTarget_ConfiguredBranchUsedWhenFlagAbsent(t *testing.T) {
	repo, integrationWT := scratchRepo(t)
	chdirRepo(t, repo)

	branch, wt := resolveIntegrationTarget("", "fixture-integration")
	if branch != "fixture-integration" {
		t.Errorf("branch = %q, want the configured git-flow develop branch", branch)
	}
	if wt != realPath(t, integrationWT) {
		t.Errorf("worktree = %q, want the configured branch's worktree %q", wt, realPath(t, integrationWT))
	}
}

func TestResolveIntegrationTarget_NoWorktreeForBranchRecordsEmpty(t *testing.T) {
	repo, _ := scratchRepo(t)
	chdirRepo(t, repo)

	branch, wt := resolveIntegrationTarget("ghost-branch", "")
	if branch != "ghost-branch" {
		t.Errorf("branch = %q, want ghost-branch", branch)
	}
	if wt != "" {
		t.Errorf("worktree = %q, want an honest empty over a confidently wrong path", wt)
	}
}

func TestResolveIntegrationTarget_NoConfigNoFlagFallsBackToCallerTree(t *testing.T) {
	repo, _ := scratchRepo(t)
	chdirRepo(t, repo)

	branch, wt := resolveIntegrationTarget("", "")
	if branch != currentBranch() {
		t.Errorf("branch = %q, want the caller's checked-out branch %q", branch, currentBranch())
	}
	if wt != realPath(t, repo) {
		t.Errorf("worktree = %q, want the caller's cwd %q — in the fallback the caller's tree IS the tree being integrated", wt, realPath(t, repo))
	}
}

// --- the acquire verb end to end: the record names the integration tree ---

func TestIntegrationAcquire_RecordsTheIntegrationWorktreeNotTheCaller(t *testing.T) {
	repo, integrationWT := scratchRepo(t)
	writeGitStrategyFixture(t, repo, "git-flow", "fixture-integration")
	chdirRepo(t, repo)

	// No --branch: the configured develop branch must drive the record, even
	// though the caller (this process) sits in the repo on main.
	if _, err := runIntegration(t, repo, "acquire", "--session", "sess-lane12", "--name", "lane-12"); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	lock, err := kanban.ReadIntegrationLock(repo)
	if err != nil {
		t.Fatal(err)
	}
	if lock.Branch != "fixture-integration" {
		t.Errorf("recorded branch = %q, want the configured git-flow develop branch (the caller's own branch is not the window)", lock.Branch)
	}
	if lock.Worktree != realPath(t, integrationWT) {
		t.Errorf("recorded worktree = %q, want the tree that actually holds the branch %q (the caller's cwd is not the window)", lock.Worktree, realPath(t, integrationWT))
	}
}

func TestIntegrationAcquire_ExplicitBranchResolvesItsWorktree(t *testing.T) {
	repo, integrationWT := scratchRepo(t)
	// The config names a DIFFERENT branch; the explicit flag must win anyway.
	writeGitStrategyFixture(t, repo, "git-flow", "fixture-other")
	chdirRepo(t, repo)

	if _, err := runIntegration(t, repo, "acquire", "--session", "sess-lane12", "--branch", "fixture-integration"); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	lock, err := kanban.ReadIntegrationLock(repo)
	if err != nil {
		t.Fatal(err)
	}
	if lock.Branch != "fixture-integration" {
		t.Errorf("recorded branch = %q, want the explicit flag", lock.Branch)
	}
	if lock.Worktree != realPath(t, integrationWT) {
		t.Errorf("recorded worktree = %q, want %q", lock.Worktree, realPath(t, integrationWT))
	}
}

func TestIntegrationAcquire_NoMatchingWorktreeRecordsEmptyWorktree(t *testing.T) {
	repo, _ := scratchRepo(t)
	writeGitStrategyFixture(t, repo, "git-flow", "ghost-branch")
	chdirRepo(t, repo)

	if _, err := runIntegration(t, repo, "acquire", "--session", "sess-lane12"); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	lock, err := kanban.ReadIntegrationLock(repo)
	if err != nil {
		t.Fatal(err)
	}
	if lock.Branch != "ghost-branch" {
		t.Errorf("recorded branch = %q, want ghost-branch", lock.Branch)
	}
	if lock.Worktree != "" {
		t.Errorf("recorded worktree = %q, want an empty unknown over the caller's cwd", lock.Worktree)
	}
}

func TestIntegrationAcquire_NonGitFlowFallsBackToCallerTree(t *testing.T) {
	repo, _ := scratchRepo(t)
	// develop_branch present but the workflow is github-flow: the key must be
	// ignored, and the caller's tree — here genuinely the tree being
	// integrated — recorded as before.
	writeGitStrategyFixture(t, repo, "github-flow", "fixture-integration")
	chdirRepo(t, repo)

	if _, err := runIntegration(t, repo, "acquire", "--session", "sess-lane12"); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	lock, err := kanban.ReadIntegrationLock(repo)
	if err != nil {
		t.Fatal(err)
	}
	if lock.Branch != "main" {
		t.Errorf("recorded branch = %q, want the caller's branch main", lock.Branch)
	}
	if lock.Worktree != realPath(t, repo) {
		t.Errorf("recorded worktree = %q, want the caller's cwd %q", lock.Worktree, realPath(t, repo))
	}
}

// --- status: the human reading it must know WHICH lane holds the window ---

func TestIntegrationStatus_ShowsHolderNameAlongsideID(t *testing.T) {
	root := t.TempDir()
	if _, err := kanban.AcquireIntegrationLock(root, kanban.IntegrationLock{
		SessionID:   "sess-abc123",
		SessionName: "lane-12",
		Branch:      "fixture-integration",
		Worktree:    "/tmp/integration-tree",
	}, false); err != nil {
		t.Fatalf("seed lock: %v", err)
	}

	out, err := runIntegration(t, root, "status")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(out, "holder:   lane-12 (sess-abc123, pid") {
		t.Errorf("status does not show the holder name alongside the id:\n%s", out)
	}
}

func TestIntegrationStatus_NoNameKeepsTodaysShape(t *testing.T) {
	root := t.TempDir()
	if _, err := kanban.AcquireIntegrationLock(root, kanban.IntegrationLock{
		SessionID: "sess-abc123",
		Branch:    "fixture-integration",
		Worktree:  "/tmp/integration-tree",
	}, false); err != nil {
		t.Fatalf("seed lock: %v", err)
	}

	out, err := runIntegration(t, root, "status")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(out, "holder:   sess-abc123 (pid") {
		t.Errorf("nameless holder does not keep today's shape:\n%s", out)
	}
}
