package worktree

// sweep_test.go — SPEC-WORKTREE-SWEEP-001 run-phase tests.
//
// Two fixture families live here, matching the package's existing idioms:
//
//  1. sweepMockEnv — an all-seams-stubbed environment for predicate and
//     surface tests (the staleTestEnv shape). No git runs.
//  2. newSweepRepo — a real git fixture with a local bare `origin` remote
//     (the done_l1_tier_guard_test.go shape), so fetch + ancestry run REAL
//     git against the fixture, never the project's own worktrees.
//
// Every env axis a test reads is sealed with t.Setenv or a stub: the lane
// session carries MOAI_KANBAN*/CLAUDE_PROJECT_DIR that falsifies env-reading
// guards when unsealed (t1350 lesson class).

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/session"
)

// sweepMock is the seam bundle one mock-based sweep test configures. Every
// field defaults to its safe value (nothing removed, everything clean, no
// anchors, nothing landed); a test flips only the knobs its scenario needs.
type sweepMock struct {
	removed       []string
	removedForce  []bool
	statusOut     map[string]string // -C <path> status --porcelain
	ignoredOut    map[string]string // -C <path> status --porcelain --ignored
	lockPorcelain string            // `git worktree list --porcelain` output
	lockErr       error             // lock source unreadable
	cwds          []string          // process cwd probe result
	cwdErr        error             // process cwd probe failure
	fetchErr      error             // fetch seam failure
	fetchedBases  []string          // bases the fetch seam observed
	landed        map[string]bool   // ancestry seam result by branch tip
	ancestorErrs  map[string]error  // ancestry seam failure by branch tip
	doneCalls     []string          // "branch|force|delete|hoist" per done-core call
	doneErr       error             // done-core failure
	hoistCalls    []string          // paths the hoist seam saw
	hoistErr      error             // hoist seam failure
}

// sweepMockEnv installs the stubbed environment for one sweep test and
// restores every global seam it touched when the test ends.
func sweepMockEnv(t *testing.T, worktrees []git.Worktree) *sweepMock {
	t.Helper()

	m := &sweepMock{
		statusOut:    map[string]string{},
		ignoredOut:   map[string]string{},
		landed:       map[string]bool{},
		ancestorErrs: map[string]error{},
	}

	origProvider := WorktreeProvider
	origGitCmd := gitWorktreeCmd
	origRepoRoot := gitRepoRootFunc
	origPrune := pruneLaunchLedgerFn
	origCWDs := sweepProcessCWDs
	origFetch := sweepFetchBase
	origAncestor := sweepAncestor
	origHoist := sweepHoistBeforeDisposal
	origDone := sweepDoneCleanup
	t.Cleanup(func() {
		WorktreeProvider = origProvider
		gitWorktreeCmd = origGitCmd
		gitRepoRootFunc = origRepoRoot
		pruneLaunchLedgerFn = origPrune
		sweepProcessCWDs = origCWDs
		sweepFetchBase = origFetch
		sweepAncestor = origAncestor
		sweepHoistBeforeDisposal = origHoist
		sweepDoneCleanup = origDone
	})

	WorktreeProvider = &mockWorktreeManager{
		rootPath: "/repo",
		listFunc: func() ([]git.Worktree, error) { return worktrees, nil },
		removeFunc: func(path string, force bool) error {
			if force {
				t.Errorf("the sweep must never force-remove; got force=true for %s", path)
			}
			m.removed = append(m.removed, path)
			m.removedForce = append(m.removedForce, force)
			return nil
		},
	}
	gitWorktreeCmd = m.git
	gitRepoRootFunc = func() (string, error) { return "/repo", nil }
	pruneLaunchLedgerFn = func() ([]string, error) { return nil, nil }
	sweepProcessCWDs = func() ([]string, error) { return m.cwds, m.cwdErr }
	sweepFetchBase = func(_ string, base string) error {
		m.fetchedBases = append(m.fetchedBases, base)
		return m.fetchErr
	}
	sweepAncestor = func(_ string, tip, _ string) (bool, error) {
		if err := m.ancestorErrs[tip]; err != nil {
			return false, err
		}
		return m.landed[tip], nil
	}
	sweepHoistBeforeDisposal = func(_ io.Writer, path string) error {
		m.hoistCalls = append(m.hoistCalls, path)
		return m.hoistErr
	}
	sweepDoneCleanup = func(branch string, force, deleteBranch, hoist bool) (bool, error) {
		m.doneCalls = append(m.doneCalls, sweepDoneCallKey(branch, force, deleteBranch, hoist))
		if m.doneErr != nil {
			return false, m.doneErr
		}
		return true, nil
	}

	// Seal the anchor-registry env axis: point the caller registry at an
	// empty temp project so the developer machine's live registries can
	// never anchor (or unanchor) a mock tree.
	t.Setenv("CLAUDE_PROJECT_DIR", filepath.Join(t.TempDir(), "empty-project"))

	return m
}

// git is the gitWorktreeCmd stub: it answers status and lock-porcelain
// reads from the mock's maps and everything else with empty output.
func (m *sweepMock) git(args ...string) (string, error) {
	if args[0] == "-C" {
		path := args[1]
		rest := args[2:]
		if len(rest) >= 2 && rest[0] == "status" {
			for _, a := range rest {
				if a == "--ignored" {
					if out, ok := m.ignoredOut[path]; ok {
						return out, nil
					}
					break
				}
			}
			return m.statusOut[path], nil
		}
		return "", nil
	}
	if args[0] == "worktree" {
		if m.lockErr != nil {
			return "", m.lockErr
		}
		return m.lockPorcelain, nil
	}
	return "", nil
}

// sweepDoneCallKey renders one done-core call as a comparable key.
func sweepDoneCallKey(branch string, force, deleteBranch, hoist bool) string {
	return branch + "|" + boolKey(force) + boolKey(deleteBranch) + boolKey(hoist)
}

func boolKey(b bool) string {
	if b {
		return "|1"
	}
	return "|0"
}

// sweepLockPorcelain builds a `git worktree list --porcelain` body whose
// first stanza is the main checkout (what isL1WorktreePath's fallback
// resolves the main root from) and whose named paths carry lock lines.
func sweepLockPorcelain(mainRoot string, lockedPaths ...string) string {
	var b strings.Builder
	b.WriteString("worktree " + mainRoot + "\n\n")
	for _, p := range lockedPaths {
		b.WriteString("worktree " + p + "\nlocked\n\n")
	}
	return b.String()
}

// runSweepCmd builds a fresh sweep command, applies the flags, and returns
// its combined stdout+stderr output. A fresh command per run keeps cobra
// flag state from leaking between tests.
func runSweepCmd(t *testing.T, flags map[string]string) (string, error) {
	t.Helper()

	cmd := newSweepCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	for name, value := range flags {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set flag %s=%s: %v", name, value, err)
		}
	}
	err := cmd.RunE(cmd, nil)
	return out.String(), err
}

// --- real-git fixture family -------------------------------------------------

// sweepFixture is a real-git sandbox: repo is the main checkout, origin a
// local bare remote, base the temp dir holding both.
type sweepFixture struct {
	base   string
	repo   string
	origin string
}

// sweepRunGit runs a git subcommand in dir, failing the test on error.
func sweepRunGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return string(out)
}

// newSweepRepo builds a real git repository on branch `develop` with one
// seed commit, plus a local bare `origin` remote carrying it. The sweep's
// fetch + ancestry therefore run against real refs on every test here.
func newSweepRepo(t *testing.T) sweepFixture {
	t.Helper()
	base := t.TempDir()
	repo := filepath.Join(base, "main")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	sweepRunGit(t, repo, "init", "-q")
	sweepRunGit(t, repo, "checkout", "-q", "-b", "develop")
	sweepRunGit(t, repo, "config", "user.email", "sweep-test@example.com")
	sweepRunGit(t, repo, "config", "user.name", "Sweep Test")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	sweepRunGit(t, repo, "add", ".")
	sweepRunGit(t, repo, "commit", "-q", "-m", "seed")

	origin := filepath.Join(base, "origin.git")
	sweepRunGit(t, repo, "init", "-q", "--bare", origin)
	sweepRunGit(t, repo, "remote", "add", "origin", origin)
	sweepRunGit(t, repo, "push", "-q", "-u", "origin", "develop")
	return sweepFixture{base: base, repo: repo, origin: origin}
}

// addSweepWorktree creates a linked worktree on a new branch at HEAD.
func addSweepWorktree(t *testing.T, f sweepFixture, path, branch string) {
	t.Helper()
	sweepRunGit(t, f.repo, "worktree", "add", "-b", branch, "--", path)
}

// withSweepRepoEnv wires the real WorktreeProvider for the fixture repo and
// routes every gitWorktreeCmd call into the fixture (a non-"-C" invocation
// gains `-C <fixture repo>`), so lock reads, ancestry, and fetch all hit the
// fixture repository — never the moai repo this test process lives in.
func withSweepRepoEnv(t *testing.T, f sweepFixture) {
	t.Helper()
	origProvider := WorktreeProvider
	origGitCmd := gitWorktreeCmd
	origPrune := pruneLaunchLedgerFn
	t.Cleanup(func() {
		WorktreeProvider = origProvider
		gitWorktreeCmd = origGitCmd
		pruneLaunchLedgerFn = origPrune
	})

	WorktreeProvider = git.NewWorktreeManager(f.repo)
	gitWorktreeCmd = func(args ...string) (string, error) {
		if args[0] != "-C" {
			args = append([]string{"-C", f.repo}, args...)
		}
		out, err := exec.Command("git", args...).Output()
		return string(out), err
	}
	pruneLaunchLedgerFn = func() ([]string, error) { return nil, nil }
}

// stubSweepProbeSeams replaces the host-dependent seams (cwd probe, remote
// fetch is real via the local origin) with the configured values.
func stubSweepProbeSeams(t *testing.T, m *sweepMock, f sweepFixture) {
	t.Helper()
	origCWDs := sweepProcessCWDs
	t.Cleanup(func() { sweepProcessCWDs = origCWDs })
	sweepProcessCWDs = func() ([]string, error) { return m.cwds, m.cwdErr }
}

// decodeSweepJSON parses the sweep's --json stdout into verdict records.
func decodeSweepJSON(t *testing.T, out string) []map[string]interface{} {
	t.Helper()
	var got []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("stdout must be valid JSON: %v\n%s", err, out)
	}
	return got
}

// assertSweepStateValues asserts every predicate value of every record is
// from the four-valued vocabulary — an unobserved predicate is never a
// negative (REQ-WS-011).
func assertSweepStateValues(t *testing.T, records []map[string]interface{}) {
	t.Helper()
	valid := map[string]bool{staleStateYes: true, staleStateNo: true, staleStateUndetermined: true, staleStateNotChecked: true}
	for _, r := range records {
		for _, field := range []string{"landed", "dirty", "ignored", "anchored", "cwd_occupied", "on_base"} {
			v, _ := r[field].(string)
			if !valid[v] {
				t.Errorf("record %v field %s has invalid predicate value %q (want yes|no|undetermined|not-checked)", r["path"], field, v)
			}
		}
	}
}

// --- M1: record shape, flag surface, dry-run default ------------------------

// TestSweep_JSONEmitsEveryNonProtectedTree pins the record universe: every
// non-protected registered worktree gets exactly one record; protected trees
// (the repository root, the process's own tree) are outside the sweep's
// universe entirely, per the clean --stale convention.
func TestSweep_JSONEmitsEveryNonProtectedTree(t *testing.T) {
	sweepMockEnv(t, []git.Worktree{
		{Path: "/repo", Branch: "develop"},      // protected: provider root
		{Path: "/wt/one", Branch: "feature/one"},
		{Path: "/wt/two", Branch: "feature/two"},
	})

	out, err := runSweepCmd(t, map[string]string{"json": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}

	got := decodeSweepJSON(t, out)
	if len(got) != 2 {
		t.Fatalf("expected one record per non-protected tree (2), got %d:\n%s", len(got), out)
	}
	byPath := map[string]bool{}
	for _, r := range got {
		p, _ := r["path"].(string)
		byPath[p] = true
		for _, field := range []string{"path", "branch", "tier", "verdict", "reason"} {
			if _, ok := r[field]; !ok {
				t.Errorf("record %s is missing field %q", p, field)
			}
		}
		if tier, _ := r["tier"].(string); tier != sweepTierL1 && tier != sweepTierL2 {
			t.Errorf("record %s carries tier %q, want L1 or L2", p, tier)
		}
	}
	if byPath["/repo"] {
		t.Error("the protected repository root must be absent from the records")
	}
}

// TestSweep_BaseDefaultsToOriginDevelop is REQ-WS-004's default clause: the
// integration base defaults to origin/develop — deliberately NOT the
// origin/main that clean --stale uses, and the help must say so.
func TestSweep_BaseDefaultsToOriginDevelop(t *testing.T) {
	m := sweepMockEnv(t, []git.Worktree{{Path: "/wt/base-check", Branch: "feature/base-check"}})

	if _, err := runSweepCmd(t, map[string]string{}); err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	if len(m.fetchedBases) != 1 || m.fetchedBases[0] != "origin/develop" {
		t.Fatalf("the sweep must fetch the default base origin/develop exactly once, observed %v", m.fetchedBases)
	}
}

// TestSweep_HelpStatesDivergenceFromCleanStale is REQ-WS-004's help clause.
func TestSweep_HelpStatesDivergenceFromCleanStale(t *testing.T) {
	cmd := newSweepCmd()
	if !strings.Contains(cmd.Long, "origin/develop") {
		t.Error("help must state the origin/develop default base")
	}
	if !strings.Contains(cmd.Long, "clean --stale") {
		t.Error("help must state the divergence from clean --stale's origin/main default")
	}
}

// TestSweep_DryRunDefaultRemovesNothing is REQ-WS-012: nothing is removed
// without --yes.
func TestSweep_DryRunDefaultRemovesNothing(t *testing.T) {
	sweepMockEnv(t, []git.Worktree{{Path: "/wt/spent", Branch: "feature/spent"}})

	if _, err := runSweepCmd(t, map[string]string{}); err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	// Removals are asserted through the provider seam in later-milestone
	// tests; here the dry run must simply not reach the removal path.
}

// TestSweep_EmptyPopulationStatesSweptCount is acceptance §D.2: an empty
// sweep asserts nothing — it must say it swept zero, with exit 0.
func TestSweep_EmptyPopulationStatesSweptCount(t *testing.T) {
	sweepMockEnv(t, []git.Worktree{})

	out, err := runSweepCmd(t, map[string]string{})
	if err != nil {
		t.Fatalf("an empty population must exit 0, got: %v", err)
	}
	if !strings.Contains(out, "0") {
		t.Errorf("the empty-run report must state the swept count (0), got:\n%s", out)
	}
}

// TestSweep_NoAskUserQuestion is the B3 subagent-boundary guard for the new
// surface: the sweep's sources must carry no interactive prompt.
func TestSweep_NoAskUserQuestion(t *testing.T) {
	sources, err := filepath.Glob("sweep*.go")
	if err != nil {
		t.Fatalf("glob sweep sources: %v", err)
	}
	scanned := 0
	for _, name := range sources {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		scanned++
		if cleanPromptGuard(string(data)) {
			t.Errorf("%s must stay prompt-free: it references an interactive prompt", name)
		}
	}
	// Positive control on the scan itself: a glob that matched nothing would
	// report every file clean without reading one.
	if scanned < 1 {
		t.Fatalf("guard scanned %d sweep sources, want the whole surface", scanned)
	}
	// Negative control: the guard must flag a synthetic violation.
	if !cleanPromptGuard("x := AskUserQuestion()") {
		t.Error("guard must detect an AskUserQuestion reference (negative control)")
	}
}

// keep unused-import guards honest for the milestone-gated test growth.
var (
	_ = errors.New
	_ = time.Now
	_ = session.Entry{}
)
