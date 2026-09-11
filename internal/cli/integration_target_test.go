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
	"encoding/json"
	"fmt"
	"io"
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

// --- card t637: branch provenance and the git-flow fallback warning ---

// integrationWarningPrefix is the line prefix of the acquire warning. The
// tests count warning lines by this prefix, never by total stderr lines, so an
// unrelated advisory cannot flip a result.
const integrationWarningPrefix = "[moai:integration-lock] warning:"

// warningLines returns every stderr line that carries the warning prefix.
func warningLines(stderr string) []string {
	var lines []string
	for _, line := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(line, integrationWarningPrefix) {
			lines = append(lines, line)
		}
	}
	return lines
}

// writeGitStrategyBody writes a git-strategy.yaml with an arbitrary body, for
// the shapes writeGitStrategyFixture cannot express (it hard-codes manual
// mode).
func writeGitStrategyBody(t *testing.T, projectRoot, body string) {
	t.Helper()
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// readLock reads the record a scenario wrote, failing the test on error.
func readLock(t *testing.T, root string) *kanban.IntegrationLock {
	t.Helper()
	lock, err := kanban.ReadIntegrationLock(root)
	if err != nil {
		t.Fatal(err)
	}
	return lock
}

// The record names which tier decided its branch. The source is decided from
// the same trimmed value that decides the branch, so the two cannot disagree.
func TestIntegrationAcquire_RecordsBranchSource(t *testing.T) {
	t.Run("flag", func(t *testing.T) {
		repo, _ := scratchRepo(t)
		// A git-flow project whose develop branch is empty: without the flag
		// this would be the warning scenario. A non-blank flag decides the
		// branch, so there is no fallback to warn about.
		writeGitStrategyFixture(t, repo, "git-flow", `""`)
		chdirRepo(t, repo)

		_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12", "--branch", "fixture-integration")
		if err != nil {
			t.Fatalf("acquire: %v", err)
		}
		lock := readLock(t, repo)
		if lock.BranchSource != kanban.BranchSourceFlag {
			t.Errorf("branch_source = %q, want %q", lock.BranchSource, kanban.BranchSourceFlag)
		}
		if lock.Branch != "fixture-integration" {
			t.Errorf("branch = %q, want the flag value", lock.Branch)
		}
		if got := warningLines(stderr); len(got) != 0 {
			t.Errorf("a flag resolution warned: %q", got)
		}
	})

	t.Run("blank_flag", func(t *testing.T) {
		repo, _ := scratchRepo(t)
		writeGitStrategyFixture(t, repo, "git-flow", "fixture-integration")
		chdirRepo(t, repo)

		if _, _, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12", "--branch", "   "); err != nil {
			t.Fatalf("acquire: %v", err)
		}
		lock := readLock(t, repo)
		if lock.BranchSource != kanban.BranchSourceConfig {
			t.Errorf("branch_source = %q, want %q — a blank --branch decided nothing, so it must never be recorded as the source", lock.BranchSource, kanban.BranchSourceConfig)
		}
		if lock.Branch != "fixture-integration" {
			t.Errorf("branch = %q, want the configured develop branch", lock.Branch)
		}
	})

	t.Run("config", func(t *testing.T) {
		repo, _ := scratchRepo(t)
		writeGitStrategyFixture(t, repo, "git-flow", "fixture-integration")
		chdirRepo(t, repo)

		_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12")
		if err != nil {
			t.Fatalf("acquire: %v", err)
		}
		lock := readLock(t, repo)
		if lock.BranchSource != kanban.BranchSourceConfig {
			t.Errorf("branch_source = %q, want %q", lock.BranchSource, kanban.BranchSourceConfig)
		}
		if lock.Branch != "fixture-integration" {
			t.Errorf("branch = %q, want fixture-integration", lock.Branch)
		}
		if got := warningLines(stderr); len(got) != 0 {
			t.Errorf("a configured resolution warned: %q", got)
		}

		text, err := runIntegration(t, repo, "status")
		if err != nil {
			t.Fatalf("status: %v", err)
		}
		if !strings.Contains(text, "\n  branch:   fixture-integration (source: config)\n") {
			t.Errorf("status text does not show the provenance on the branch line:\n%s", text)
		}
		jsonOut, err := runIntegration(t, repo, "status", "--json")
		if err != nil {
			t.Fatalf("status --json: %v", err)
		}
		if !strings.Contains(jsonOut, `"branch_source":"config"`) {
			t.Errorf("status --json lock object does not carry branch_source: %s", jsonOut)
		}
	})

	t.Run("caller", func(t *testing.T) {
		repo, _ := scratchRepo(t)
		chdirRepo(t, repo)

		if _, _, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12"); err != nil {
			t.Fatalf("acquire: %v", err)
		}
		lock := readLock(t, repo)
		if lock.BranchSource != kanban.BranchSourceCaller {
			t.Errorf("branch_source = %q, want %q", lock.BranchSource, kanban.BranchSourceCaller)
		}
	})
}

// Control 1 — the silent fallback is no longer silent: a git-flow project
// whose develop branch is empty falls back to the caller's tree, the record
// says so, and exactly one warning names the recorded branch and both
// remedies.
func TestIntegrationAcquire_GitFlowEmptyDevelopWarnsOnCallerFallback(t *testing.T) {
	repo, _ := scratchRepo(t)
	writeGitStrategyFixture(t, repo, "git-flow", `""`)
	chdirRepo(t, repo)

	_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12", "--name", "lane-12")
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	lines := warningLines(stderr)
	if len(lines) != 1 {
		t.Fatalf("want exactly one warning line, got %d: %q (stderr %q)", len(lines), lines, stderr)
	}
	for _, token := range []string{"main", "develop_branch", "--branch"} {
		if !strings.Contains(lines[0], token) {
			t.Errorf("warning does not name %q: %q", token, lines[0])
		}
	}
	lock := readLock(t, repo)
	if lock.BranchSource != kanban.BranchSourceCaller {
		t.Errorf("branch_source = %q, want %q", lock.BranchSource, kanban.BranchSourceCaller)
	}
	if lock.Branch != "main" {
		t.Errorf("branch = %q, want the caller's branch main (resolution unchanged)", lock.Branch)
	}
}

// Control 2 — a lane standing in one tree while integrating another card:
// the tool cannot know the intended source branch, but text status alone now
// shows the card next to a caller-sourced branch, which is the mismatch.
func TestIntegrationAcquire_CallerFallbackMismatchIsVisible(t *testing.T) {
	repo, _ := scratchRepo(t)
	writeGitStrategyFixture(t, repo, "git-flow", `""`)
	chdirRepo(t, repo)

	_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12", "--card", "t-other")
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if got := warningLines(stderr); len(got) != 1 {
		t.Errorf("want exactly one warning line, got %d: %q", len(got), got)
	}
	text, err := runIntegration(t, repo, "status")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	for _, want := range []string{"\n  card:     t-other\n", "\n  branch:   main (source: caller)\n"} {
		if !strings.Contains(text, want) {
			t.Errorf("status text lacks %q:\n%s", want, text)
		}
	}
}

// No-warn negative — github-flow with an EMPTY develop branch. The workflow
// half of the git-flow predicate is the only difference from the warning
// cell.
func TestIntegrationAcquire_GitHubFlowEmptyDevelopDoesNotWarn(t *testing.T) {
	repo, _ := scratchRepo(t)
	writeGitStrategyFixture(t, repo, "github-flow", `""`)
	chdirRepo(t, repo)

	_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12")
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if got := warningLines(stderr); len(got) != 0 {
		t.Errorf("a github-flow caller fallback warned: %q", got)
	}
	lock := readLock(t, repo)
	if lock.BranchSource != kanban.BranchSourceCaller || lock.Branch != "main" {
		t.Errorf("record = (%q, source %q), want (main, source %q)", lock.Branch, lock.BranchSource, kanban.BranchSourceCaller)
	}
}

// No-warn negative — a git-flow workflow under a non-manual mode with an
// EMPTY develop branch. The mode half of the predicate is the only
// difference from the warning cell: develop_branch is a manual-mode key.
func TestIntegrationAcquire_NonManualModeGitFlowDoesNotWarn(t *testing.T) {
	repo, _ := scratchRepo(t)
	writeGitStrategyBody(t, repo, "git_strategy:\n    mode: personal\n    personal:\n        workflow: git-flow\n        develop_branch: \"\"\n")
	chdirRepo(t, repo)

	_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12")
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if got := warningLines(stderr); len(got) != 0 {
		t.Errorf("a personal-mode git-flow caller fallback warned: %q", got)
	}
	lock := readLock(t, repo)
	if lock.BranchSource != kanban.BranchSourceCaller || lock.Branch != "main" {
		t.Errorf("record = (%q, source %q), want (main, source %q)", lock.Branch, lock.BranchSource, kanban.BranchSourceCaller)
	}
}

// No-warn negative — no git strategy file. The tool cannot assert a
// misconfiguration it could not read.
func TestIntegrationAcquire_NoConfigCallerFallbackDoesNotWarn(t *testing.T) {
	repo, _ := scratchRepo(t)
	chdirRepo(t, repo)

	_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12")
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if got := warningLines(stderr); len(got) != 0 {
		t.Errorf("a caller fallback with no git strategy file warned: %q", got)
	}
	lock := readLock(t, repo)
	if lock.BranchSource != kanban.BranchSourceCaller || lock.Branch != "main" {
		t.Errorf("record = (%q, source %q), want (main, source %q)", lock.Branch, lock.BranchSource, kanban.BranchSourceCaller)
	}
}

// Warn-only: the warning never touches the result channel. Text-mode stdout
// is exactly the acquired line, and --json stdout is exactly one JSON object;
// the warning lives on standard error alone.
func TestIntegrationAcquire_WarningIsOnStderrOnly(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		repo, _ := scratchRepo(t)
		writeGitStrategyFixture(t, repo, "git-flow", `""`)
		chdirRepo(t, repo)

		stdout, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12")
		if err != nil {
			t.Fatalf("acquire: %v", err)
		}
		if want := "release-integration window acquired by sess-lane12 on main\n"; stdout != want {
			t.Errorf("stdout = %q, want exactly %q", stdout, want)
		}
		if got := warningLines(stderr); len(got) != 1 {
			t.Errorf("want exactly one warning line on stderr, got %d: %q", len(got), got)
		}
		if lock := readLock(t, repo); !lock.Held() {
			t.Error("the warning scenario did not write the record")
		}
	})

	t.Run("json", func(t *testing.T) {
		repo, _ := scratchRepo(t)
		writeGitStrategyFixture(t, repo, "git-flow", `""`)
		chdirRepo(t, repo)

		stdout, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12", "--json")
		if err != nil {
			t.Fatalf("acquire --json: %v", err)
		}
		dec := json.NewDecoder(strings.NewReader(stdout))
		var obj map[string]any
		if err := dec.Decode(&obj); err != nil {
			t.Fatalf("acquire --json stdout is not JSON (%v): %q", err, stdout)
		}
		if obj["acquired"] != true {
			t.Errorf(`acquire --json object lacks "acquired": true: %q`, stdout)
		}
		var extra any
		if err := dec.Decode(&extra); err != io.EOF {
			t.Errorf("acquire --json stdout carries more than one JSON value (%v): %q", err, stdout)
		}
		if got := warningLines(stderr); len(got) != 1 {
			t.Errorf("want exactly one warning line on stderr, got %d: %q", len(got), got)
		}
	})
}

// A refused acquire takes no window, so there is no recorded fallback to
// warn about: it returns the held error and writes no warning.
func TestIntegrationAcquire_RefusedAcquireDoesNotWarn(t *testing.T) {
	repo, _ := scratchRepo(t)
	writeGitStrategyFixture(t, repo, "git-flow", `""`)
	chdirRepo(t, repo)

	// Held by another LIVE session: this test process's own pid is alive for
	// the whole test, so the record cannot read as reclaimable.
	if _, err := kanban.AcquireIntegrationLock(repo, kanban.IntegrationLock{
		SessionID:   "sess-holder",
		SessionName: "lane-holder",
		PID:         os.Getpid(),
		Branch:      "fixture-integration",
	}, false); err != nil {
		t.Fatalf("seed holder: %v", err)
	}

	_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12")
	if !kanban.IsIntegrationLockHeld(err) {
		t.Fatalf("acquire over a live holder returned %v, want the held error", err)
	}
	if got := warningLines(stderr); len(got) != 0 {
		t.Errorf("a refused acquire warned: %q", got)
	}
	if lock := readLock(t, repo); lock.SessionID != "sess-holder" {
		t.Errorf("a refused acquire rewrote the record: holder %q", lock.SessionID)
	}
}
