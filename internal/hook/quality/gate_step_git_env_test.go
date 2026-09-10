package quality

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"context"
)

// Card t516 (GH #1691) — a gate step's child process inherited the git
// environment of whatever invoked `moai gate`.
//
// The invoker that matters is the git pre-commit hook: git exports GIT_DIR
// (and, on the commit path, GIT_INDEX_FILE) into the hook's environment, the
// hook shells out to `moai gate` (see internal/cli/hook_install_precommit.go),
// and runStep built its exec.Cmd without ever setting cmd.Env — so os/exec
// handed the whole parent environment to the step's child.
//
// GIT_DIR outranks the working directory. cmd.Dir therefore did not confine
// the child at all: a project test suite whose fixtures create throwaway
// commits had those commits land in the HOST repository, not in the temp
// directory the fixture built. The reported incident lost an uncommitted
// reviewed diff and injected a junk commit chain into a live branch.
//
// The three tests below are one reproduction at three depths: the leak itself,
// the write it lets through, and whether the same leak also explains the
// core.bare flip reported alongside it.

const (
	// helperGitEnvReportEnv carries the path the env-reporting helper writes
	// to. It is a MOAI_ variable on purpose: the fix scrubs git repo-scoping
	// variables only, so this channel survives the fix and the test can still
	// observe what the child saw.
	helperGitEnvReportEnv = "MOAI_TEST_HELPER_GIT_ENV_REPORT"

	// helperGitFixtureEnv switches the test binary into "behave like a project
	// test fixture that makes a throwaway commit" mode.
	helperGitFixtureEnv = "MOAI_TEST_HELPER_GIT_FIXTURE"

	// helperGitInitEnv switches the test binary into "behave like a fixture
	// that initialises a scratch repository" mode.
	helperGitInitEnv = "MOAI_TEST_HELPER_GIT_INIT"
)

// gitEnvNamesUnderTest is the test's OWN list of the variables a leak would
// carry. It is deliberately not the production list: a test that asserted
// against the production variable would pass by construction whatever that
// variable contained.
var gitEnvNamesUnderTest = []string{
	"GIT_DIR",
	"GIT_WORK_TREE",
	"GIT_INDEX_FILE",
	"GIT_OBJECT_DIRECTORY",
}

// ---------------------------------------------------------------------------
// helpers re-executed as gate-step children
// ---------------------------------------------------------------------------

// TestHelperReportGitEnv is not a test. Re-executed as a gate step, it records
// which git repo-scoping variables reached the child.
func TestHelperReportGitEnv(t *testing.T) {
	path := os.Getenv(helperGitEnvReportEnv)
	if path == "" {
		t.Skip("helper process only")
	}
	var b strings.Builder
	for _, name := range gitEnvNamesUnderTest {
		fmt.Fprintf(&b, "%s=%s\n", name, os.Getenv(name))
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("helper: write report: %v", err)
	}
}

// TestHelperGitFixtureCommit is not a test. Re-executed as a gate step, it does
// what an ancestry-fixture in a project's own suite does: write a file and
// commit it, from the step's working directory, using a plain `git` that
// inherits whatever environment it was given.
func TestHelperGitFixtureCommit(t *testing.T) {
	if os.Getenv(helperGitFixtureEnv) == "" {
		t.Skip("helper process only")
	}
	if err := os.WriteFile("t516_fixture.txt", []byte("throwaway\n"), 0o644); err != nil {
		t.Fatalf("helper: write fixture: %v", err)
	}
	// Errors are not fatal here: the parent judges the outcome by inspecting
	// both repositories, not by this process's exit code.
	_ = exec.Command("git", "add", "t516_fixture.txt").Run()
	_ = exec.Command("git",
		"-c", "user.name=t516 fixture",
		"-c", "user.email=fixture@t516.invalid",
		"commit", "-q", "-m", "t516 throwaway fixture commit",
	).Run()
}

// TestHelperGitInit is not a test. Re-executed as a gate step, it initialises a
// repository in the step's working directory — the other common fixture shape.
func TestHelperGitInit(t *testing.T) {
	if os.Getenv(helperGitInitEnv) == "" {
		t.Skip("helper process only")
	}
	// Not -q: the message distinguishes "Initialized empty" (a fresh repo here)
	// from "Reinitialized existing" (the caller's repo, reached through the
	// leak). The parent needs that distinction to say what git actually did,
	// rather than only that core.bare happened not to move.
	out, err := exec.Command("git", "init").CombinedOutput()
	if path := os.Getenv(helperGitEnvReportEnv); path != "" {
		_ = os.WriteFile(path, append([]byte(fmt.Sprintf("err=%v\n", err)), out...), 0o644)
	}
}

// ---------------------------------------------------------------------------
// test-local git plumbing (never inherits the leaked environment)
// ---------------------------------------------------------------------------

// cleanGitEnv returns the current environment with every variable the test sets
// removed, so a parent-side assertion reads the repository it names rather than
// the one GIT_DIR points at.
func cleanGitEnv() []string {
	out := make([]string, 0, len(os.Environ()))
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(name, "GIT_") {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// gitIn runs git in dir with a scrubbed environment and returns trimmed stdout.
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = cleanGitEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// gitInAllowFail is gitIn without the fatal: used where a non-zero exit is a
// legitimate observation (a config key that is simply absent).
func gitInAllowFail(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = cleanGitEnv()
	out, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(out))
}

// newRepo creates an initialised repository with one commit and returns its path.
func newRepo(t *testing.T, dir string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	gitIn(t, dir, "init", "-q")
	gitIn(t, dir, "config", "user.name", "t516")
	gitIn(t, dir, "config", "user.email", "t516@test.invalid")
	if err := os.WriteFile(filepath.Join(dir, "seed.txt"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	gitIn(t, dir, "add", "seed.txt")
	gitIn(t, dir, "commit", "-q", "-m", "seed")
	return dir
}

// requireGit skips when git is unavailable.
func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
}

// leakPreCommitGitEnv reproduces the environment git hands a pre-commit hook:
// GIT_DIR names the repository being committed to, GIT_INDEX_FILE names the
// index the commit is being assembled in. Faithfulness matters here — an
// earlier draft of this file also set GIT_WORK_TREE, which git does NOT export
// to hooks, and the extra variable made the child's `git add` fail outright
// ("outside repository") instead of writing into the host. A leak model
// stronger than the real one hides the very write it is meant to catch.
func leakPreCommitGitEnv(t *testing.T, hostRepo string) {
	t.Helper()
	gitDir := filepath.Join(hostRepo, ".git")
	t.Setenv("GIT_DIR", gitDir)
	t.Setenv("GIT_INDEX_FILE", filepath.Join(gitDir, "index"))
}

// stepGate builds a gate whose steps run in stepDir.
func stepGate(stepDir string) *QualityGate {
	cfg := DefaultGateConfig()
	cfg.ProjectDir = stepDir
	return NewQualityGate(cfg)
}

// helperStep returns the argv that re-executes this binary as the named helper.
func helperStep(name string) (string, []string) {
	return os.Args[0], []string{"-test.run=^" + name + "$", "-test.timeout=60s"}
}

// ---------------------------------------------------------------------------
// the scrub itself
// ---------------------------------------------------------------------------

// An empty input must still produce a non-nil slice: os/exec reads a nil Env as
// "inherit the parent's environment", so returning nil here would restore the
// exact defect on any caller that passes nothing.
func TestScrubGitRepoScoping_EmptyInputIsNonNil(t *testing.T) {
	if got := scrubGitRepoScoping(nil); got == nil {
		t.Fatal("scrubGitRepoScoping(nil) returned nil; os/exec reads a nil Env as full inheritance")
	}
	if got := scrubGitRepoScoping([]string{}); got == nil {
		t.Fatal("scrubGitRepoScoping(empty) returned nil; os/exec reads a nil Env as full inheritance")
	}
}

// Only repository location is removed. Identity and behaviour variables reach
// the step, and a variable that merely starts with GIT_ is not removed for that
// reason alone.
func TestScrubGitRepoScoping_KeepsIdentityAndBehaviour(t *testing.T) {
	in := []string{
		"GIT_DIR=/outer/.git",
		"GIT_INDEX_FILE=/outer/.git/index",
		"GIT_AUTHOR_NAME=Someone",
		"GIT_COMMITTER_EMAIL=someone@example.invalid",
		"GIT_EDITOR=true",
		"GIT_CONFIG_GLOBAL=/ci/gitconfig",
		"PATH=/usr/bin",
		"NOT_A_PAIR",
	}
	want := []string{
		"GIT_AUTHOR_NAME=Someone",
		"GIT_COMMITTER_EMAIL=someone@example.invalid",
		"GIT_EDITOR=true",
		"GIT_CONFIG_GLOBAL=/ci/gitconfig",
		"PATH=/usr/bin",
		"NOT_A_PAIR",
	}
	got := scrubGitRepoScoping(in)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("scrub changed the wrong entries\n got: %v\nwant: %v", got, want)
	}
}

// ---------------------------------------------------------------------------
// A — the leak itself
// ---------------------------------------------------------------------------

// A gate step's child must not be told which repository the caller was working
// in. cmd.Dir is the only repository scoping a step may carry.
func TestRunStep_DoesNotLeakOuterGitEnvToChild(t *testing.T) {
	requireGit(t)

	base := t.TempDir()
	host := newRepo(t, filepath.Join(base, "host"))
	stepDir := filepath.Join(base, "step")
	if err := os.MkdirAll(stepDir, 0o755); err != nil {
		t.Fatalf("mkdir step: %v", err)
	}
	report := filepath.Join(base, "child-env.txt")

	leakPreCommitGitEnv(t, host)
	// Widened beyond the pre-commit set on purpose: the scrub is asserted over
	// every repo-scoping variable a caller might hold, not only the two git
	// itself exports to a hook.
	t.Setenv("GIT_WORK_TREE", host)
	t.Setenv("GIT_OBJECT_DIRECTORY", filepath.Join(host, ".git", "objects"))
	t.Setenv(helperGitEnvReportEnv, report)

	name, args := helperStep("TestHelperReportGitEnv")
	if ok, msg := stepGate(stepDir).runStep(context.Background(), "probe", 60*time.Second, name, args...); !ok {
		t.Fatalf("probe step failed: %s", msg)
	}

	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatalf("read child env report: %v", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		key, value, _ := strings.Cut(line, "=")
		if value != "" {
			t.Errorf("gate step child inherited %s=%q from the caller; a step's only repository scoping may be its working directory", key, value)
		}
	}
}

// ---------------------------------------------------------------------------
// B — the write the leak lets through (the reported incident)
// ---------------------------------------------------------------------------

// A fixture commit made by a gate step's child must land in the step's own
// repository, never in the repository the caller was committing to.
func TestRunStep_ChildFixtureCommitDoesNotLandInOuterRepo(t *testing.T) {
	requireGit(t)

	base := t.TempDir()
	host := newRepo(t, filepath.Join(base, "host"))
	stepDir := newRepo(t, filepath.Join(base, "step"))

	hostBefore := gitIn(t, host, "rev-list", "--count", "HEAD")
	stepBefore := gitIn(t, stepDir, "rev-list", "--count", "HEAD")

	leakPreCommitGitEnv(t, host)
	t.Setenv(helperGitFixtureEnv, "1")

	name, args := helperStep("TestHelperGitFixtureCommit")
	if ok, msg := stepGate(stepDir).runStep(context.Background(), "fixture", 60*time.Second, name, args...); !ok {
		t.Fatalf("fixture step failed: %s", msg)
	}

	hostAfter := gitIn(t, host, "rev-list", "--count", "HEAD")
	stepAfter := gitIn(t, stepDir, "rev-list", "--count", "HEAD")

	if hostAfter != hostBefore {
		t.Errorf("the gate step's child wrote a commit into the OUTER repository: %s commits before, %s after.\nlatest outer commit: %s",
			hostBefore, hostAfter, gitIn(t, host, "log", "-1", "--oneline"))
	}
	if stepAfter == stepBefore {
		t.Errorf("the fixture commit did not land in the step's own repository either (%s commits before and after) — the child wrote somewhere else entirely",
			stepBefore)
	}
	if _, err := os.Stat(filepath.Join(host, "t516_fixture.txt")); err == nil {
		t.Error("the fixture file was checked out into the OUTER working tree")
	}
}

// ---------------------------------------------------------------------------
// C — does the same leak explain the reported core.bare flip?
// ---------------------------------------------------------------------------

// GH #1691 also reports that a failed gate run leaves core.bare=true in the
// repository's config, root-caused there "by elimination". No MoAI code writes
// that key — a grep for `core.bare` over the whole tree (no extension filter,
// .git excluded) returns nothing, while the same grep for `core.hooksPath`
// returns matches, so the absence is the tool reporting rather than the tool
// failing. If the flip is real, its author is git itself acting on a leaked
// GIT_DIR.
//
// What this test measured, on git 2.50.1 (Apple Git-155):
//
//   - The leak DOES reach the outer repository through `git init`: the child
//     printed "Reinitialized existing Git repository in <host>/.git/" while its
//     working directory was the step's own, untracked directory. This is a
//     second write path into the caller's repo, independent of the commit path
//     the sibling test covers.
//   - core.bare did NOT move. So the reporter's attribution is UNVERIFIED, not
//     disproven: the mechanism reaches the repository but was not observed to
//     set that key on this git version, and the reporter's environment (Linux /
//     WSL2, moai-adk 3.1.2) was not reproduced here.
//
// The assertion is kept as a regression guard on what was actually established:
// a gate step's child must not re-initialise the caller's repository.
func TestRunStep_ChildGitInitDoesNotFlipOuterRepoCoreBare(t *testing.T) {
	requireGit(t)

	base := t.TempDir()
	host := newRepo(t, filepath.Join(base, "host"))
	stepDir := filepath.Join(base, "step")
	if err := os.MkdirAll(stepDir, 0o755); err != nil {
		t.Fatalf("mkdir step: %v", err)
	}

	bareBefore := gitInAllowFail(host, "config", "--get", "core.bare")
	report := filepath.Join(base, "init-output.txt")

	leakPreCommitGitEnv(t, host)
	t.Setenv(helperGitInitEnv, "1")
	t.Setenv(helperGitEnvReportEnv, report)

	name, args := helperStep("TestHelperGitInit")
	if ok, msg := stepGate(stepDir).runStep(context.Background(), "init", 60*time.Second, name, args...); !ok {
		t.Fatalf("init step failed: %s", msg)
	}

	// What git did is asserted, not merely logged: core.bare not moving is the
	// weaker fact, and on its own it would let a pass stand while the child was
	// still re-initialising the caller's repository.
	out, err := os.ReadFile(report)
	if err != nil {
		t.Fatalf("read child init output: %v", err)
	}
	t.Logf("child `git init` reported:\n%s", out)
	if strings.Contains(string(out), host) {
		t.Errorf("the gate step's child re-initialised the OUTER repository (%s); its git init reported:\n%s", host, out)
	}

	bareAfter := gitInAllowFail(host, "config", "--get", "core.bare")
	if bareAfter != bareBefore {
		t.Errorf("a gate step's child re-initialised the OUTER repository: core.bare %q before, %q after", bareBefore, bareAfter)
	}
}
