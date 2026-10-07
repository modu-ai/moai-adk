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
	statusErr     error             // -C <path> status failure (dirty-check-failed)
	ignoredOut    map[string]string // -C <path> status --porcelain --ignored
	ignoredCalls  int               // --ignored invocations observed
	ignoredAfter  int               // when > 0: --ignored returns content only after this many calls
	lockPorcelain string            // `git worktree list --porcelain` output
	lockErr       error             // lock source unreadable
	cwds          []string          // process cwd probe result
	cwdErr        error             // process cwd probe failure
	fetchCmdErr   error             // git fetch subprocess failure (default seam)
	fetchArgs     [][]string        // git fetch subprocess invocations
	removeErr     map[string]error  // provider Remove failure by path
	fetchErr      error             // fetch seam failure
	fetchedBases  []string          // bases the fetch seam observed
	landed        map[string]bool   // ancestry seam result by branch tip
	ancestorErrs  map[string]error  // ancestry seam failure by branch tip
	doneCalls     []string          // "branch|force|delete|hoist" per done-core call
	doneErr       error             // done-core failure
	doneSuccess   bool              // done-core success return (initialized true; a cell flips it)
	hoistCalls    []string          // paths the hoist seam saw
	hoistErr      error             // hoist seam hard failure
	hoistComplete bool              // hoist seam completeness (initialized true; a cell flips it)
}

// sweepMockEnv installs the stubbed environment for one sweep test and
// restores every global seam it touched when the test ends.
func sweepMockEnv(t *testing.T, worktrees []git.Worktree) *sweepMock {
	t.Helper()

	m := &sweepMock{
		statusOut:     map[string]string{},
		ignoredOut:    map[string]string{},
		landed:        map[string]bool{},
		ancestorErrs:  map[string]error{},
		removeErr:     map[string]error{},
		doneSuccess:   true,
		hoistComplete: true,
	}

	origProvider := WorktreeProvider
	origGitCmd := gitWorktreeCmd
	origRepoRoot := gitRepoRootFunc
	origPrune := pruneLaunchLedgerFn
	origCWDs := sweepProcessCWDs
	origFetch := sweepFetchBase
	origAncestor := sweepAncestor
	origHoist := sweepHoistEvidence
	origDone := sweepDoneCleanup
	origConfigRoot := sweepConfigRoot
	t.Cleanup(func() {
		sweepConfigRoot = origConfigRoot
		WorktreeProvider = origProvider
		gitWorktreeCmd = origGitCmd
		gitRepoRootFunc = origRepoRoot
		pruneLaunchLedgerFn = origPrune
		sweepProcessCWDs = origCWDs
		sweepFetchBase = origFetch
		sweepAncestor = origAncestor
		sweepHoistEvidence = origHoist
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
			if err := m.removeErr[path]; err != nil {
				return err
			}
			return nil
		},
	}
	// The default --base derives from the landed-ref chain
	// (SPEC-GITHUB-FLOW-CI-RESIDUE-001 REQ-GFC-001): the fixture root is a
	// bare temp dir — no worktree_base_branch, no origin/HEAD — so the chain's
	// compiled-in default (origin/main) is the default these tests exercise.
	// Override sweepConfigRoot per test to point the chain at a fixture repo.
	gitFlowRoot := t.TempDir()
	installBaseRow(t, gitFlowRoot, gitFlowBaseRow())
	sweepConfigRoot = func() string { return gitFlowRoot }
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
	sweepHoistEvidence = func(_ io.Writer, path string) (bool, error) {
		m.hoistCalls = append(m.hoistCalls, path)
		if m.hoistErr != nil {
			return false, m.hoistErr
		}
		return m.hoistComplete, nil
	}
	sweepDoneCleanup = func(branch string, force, deleteBranch, hoist bool) (bool, error) {
		m.doneCalls = append(m.doneCalls, sweepDoneCallKey(branch, force, deleteBranch, hoist))
		if m.doneErr != nil {
			return false, m.doneErr
		}
		return m.doneSuccess, nil
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
		if len(rest) >= 1 && rest[0] == "fetch" {
			m.fetchArgs = append(m.fetchArgs, args)
			return "", m.fetchCmdErr
		}
		if len(rest) >= 1 && rest[0] == "worktree" {
			if m.lockErr != nil {
				return "", m.lockErr
			}
			return m.lockPorcelain, nil
		}
		if len(rest) >= 2 && rest[0] == "status" {
			if m.statusErr != nil {
				return "", m.statusErr
			}
			for _, a := range rest {
				if a == "--ignored" {
					m.ignoredCalls++
					if m.ignoredAfter == 0 || m.ignoredCalls > m.ignoredAfter {
						if out, ok := m.ignoredOut[path]; ok {
							return out, nil
						}
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
	// A tracked marker under .moai/ mirrors the real repo (whose .moai/
	// holds tracked content): without it git collapses a fully-ignored
	// .moai/ into one `!! .moai` entry instead of `!! .moai/reports/`.
	if err := os.MkdirAll(filepath.Join(repo, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".moai", ".keep"), []byte("\n"), 0o644); err != nil {
		t.Fatalf("write .moai/.keep: %v", err)
	}
	sweepRunGit(t, repo, "add", ".")
	sweepRunGit(t, repo, "commit", "-q", "-m", "seed")
	// .moai/reports/ is machine-local run-phase evidence; the real repo
	// ignores `.moai/reports/*`. Mirror that shape here so the fixture's
	// porcelain matches what a real card tree shows (the entry the sweep's
	// hoist-aware ignored filter discharges), and so evidence does not read
	// as untracked dirt.
	exclude := filepath.Join(repo, ".git", "info", "exclude")
	fh, err := os.OpenFile(exclude, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open exclude: %v", err)
	}
	if _, err := fh.WriteString(".moai/reports/\n"); err != nil {
		t.Fatalf("append exclude: %v", err)
	}
	_ = fh.Close()

	origin := filepath.Join(base, "origin.git")
	sweepRunGit(t, repo, "init", "-q", "--bare", origin)
	sweepRunGit(t, repo, "remote", "add", "origin", origin)
	sweepRunGit(t, repo, "push", "-q", "-u", "origin", "develop")

	// Canonicalize every fixture path (macOS /var/folders ->
	// /private/var/folders): git's porcelain reports RESOLVED paths, so an
	// unresolved root would miss the protected-tree filter and the main
	// checkout would enter the sweep's universe.
	resolve := func(p string) string {
		r, err := filepath.EvalSymlinks(p)
		if err != nil {
			t.Fatalf("evalsymlinks %s: %v", p, err)
		}
		return r
	}
	// The default --base derives from the landed-ref chain
	// (SPEC-GITHUB-FLOW-CI-RESIDUE-001 REQ-GFC-001): the git-flow row that
	// names its base through worktree_base_branch keeps origin/develop — the
	// base this develop-only fixture's fetch and ancestry exercise. Untracked,
	// so no card tree carries it.
	installBaseRow(t, repo, gitFlowCompatRow())
	return sweepFixture{base: resolve(base), repo: resolve(repo), origin: resolve(origin)}
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
// fetch is real via the local origin) with the configured values, and seals
// the caller-registry env axis: the lane session's CLAUDE_PROJECT_DIR points
// at the moai checkout, whose live registry must never anchor a fixture tree
// (t1350 lesson class).
func stubSweepProbeSeams(t *testing.T, m *sweepMock, f sweepFixture) {
	t.Helper()
	origCWDs := sweepProcessCWDs
	t.Cleanup(func() { sweepProcessCWDs = origCWDs })
	sweepProcessCWDs = func() ([]string, error) { return m.cwds, m.cwdErr }
	t.Setenv("CLAUDE_PROJECT_DIR", filepath.Join(t.TempDir(), "empty-project"))
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
		{Path: "/repo", Branch: "develop"}, // protected: provider root
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

// TestSweep_BaseDefaultsToTheResolvedLandedRef is REQ-WS-004's default clause
// under the landed-ref chain (SPEC-GITHUB-FLOW-CI-RESIDUE-001 REQ-GFC-001):
// in a root with no worktree_base_branch and no origin/HEAD the chain's
// compiled-in default (origin/main) is the default base, and --base
// overrides it.
func TestSweep_BaseDefaultsToTheResolvedLandedRef(t *testing.T) {
	m := sweepMockEnv(t, []git.Worktree{{Path: "/wt/base-check", Branch: "feature/base-check"}})

	if _, err := runSweepCmd(t, map[string]string{}); err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	if len(m.fetchedBases) != 1 || m.fetchedBases[0] != "origin/main" {
		t.Fatalf("the sweep must fetch the chain default origin/main exactly once, observed %v", m.fetchedBases)
	}
}

// TestSweep_HelpStatesDivergenceFromCleanStale is REQ-WS-004's help clause:
// the help names the chain the default comes from and keeps the divergence
// from clean --stale's origin/main stated.
func TestSweep_HelpStatesDivergenceFromCleanStale(t *testing.T) {
	cmd := newSweepCmd()
	for _, want := range []string{"worktree_base_branch", "refs/remotes/origin/HEAD", "origin/main"} {
		if !strings.Contains(cmd.Long, want) {
			t.Errorf("help must state the chain that resolves the default base: missing %q", want)
		}
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
	if !strings.Contains(out, "Nothing to sweep: 0 worktree(s) evaluated.") {
		t.Errorf("the empty-run report must state the swept count, got:\n%s", out)
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

// --- M2: the remote-landing predicate (three-way contract) ------------------

// TestSweepAncestryUnanswerable is AC-WS-003: the three-way contract. No
// answer is never a no — fetch failure, a non-0/1 ancestry exit, and an
// unresolvable base each PRESERVE with their own cause token, and none of
// them may read as cause=not-landed.
func TestSweepAncestryUnanswerable(t *testing.T) {
	t.Run("fetch-failed", func(t *testing.T) {
		m := sweepMockEnv(t, []git.Worktree{{Path: "/wt/unfetched", Branch: "feature/unfetched"}})
		m.fetchErr = errors.New("could not read from remote repository")

		out, err := runSweepCmd(t, map[string]string{"json": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		got := decodeSweepJSON(t, out)
		if len(got) != 1 {
			t.Fatalf("expected 1 record, got %d:\n%s", len(got), out)
		}
		if got[0]["verdict"] != sweepPreserve {
			t.Errorf("a fetch failure must preserve, got %v", got[0]["verdict"])
		}
		if reason, _ := got[0]["reason"].(string); !strings.Contains(reason, "cause="+causeFetchFailed) {
			t.Errorf("reason must carry cause=%s, got %q", causeFetchFailed, reason)
		}
		if got[0]["landed"] != staleStateUndetermined {
			t.Errorf("an unasked landing must read %q, got %v", staleStateUndetermined, got[0]["landed"])
		}
	})

	t.Run("ancestry-exit-2", func(t *testing.T) {
		m := sweepMockEnv(t, []git.Worktree{{Path: "/wt/broken", Branch: "feature/broken"}})
		m.ancestorErrs["feature/broken"] = errors.New("merge-base exited 2")

		out, err := runSweepCmd(t, map[string]string{"json": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		got := decodeSweepJSON(t, out)
		if got[0]["verdict"] != sweepPreserve {
			t.Errorf("an unanswerable ancestry must preserve, got %v", got[0]["verdict"])
		}
		reason, _ := got[0]["reason"].(string)
		if !strings.Contains(reason, "cause="+causeLandedCheckFailed) {
			t.Errorf("reason must carry cause=%s, got %q", causeLandedCheckFailed, reason)
		}
		if strings.Contains(reason, "cause="+causeNotLanded) {
			t.Errorf("no answer must never read as cause=%s, got %q", causeNotLanded, reason)
		}
	})

	t.Run("unresolvable-base-ref", func(t *testing.T) {
		// Real git: the fetch is stubbed successful, but origin/no-such-base
		// resolves to nothing, so merge-base --is-ancestor exits 128.
		f := newSweepRepo(t)
		wt := filepath.Join(f.base, "trees", "wt-baseless")
		addSweepWorktree(t, f, wt, "feature/baseless")
		withSweepRepoEnv(t, f)
		m := &sweepMock{}
		stubSweepProbeSeams(t, m, f)
		origFetch := sweepFetchBase
		sweepFetchBase = func(string, string) error { return nil } // fetch pretends success
		t.Cleanup(func() { sweepFetchBase = origFetch })

		out, err := runSweepCmd(t, map[string]string{"json": "true", "base": "origin/no-such-base"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		got := decodeSweepJSON(t, out)
		if len(got) != 1 {
			t.Fatalf("expected 1 record, got %d:\n%s", len(got), out)
		}
		if got[0]["verdict"] != sweepPreserve {
			t.Errorf("an unresolvable base must preserve, got %v", got[0]["verdict"])
		}
		reason, _ := got[0]["reason"].(string)
		if !strings.Contains(reason, "cause="+causeLandedCheckFailed) {
			t.Errorf("reason must carry cause=%s, got %q", causeLandedCheckFailed, reason)
		}
		if strings.Contains(reason, "cause="+causeNotLanded) {
			t.Errorf("no answer must never read as cause=%s, got %q", causeNotLanded, reason)
		}
		if got[0]["landed"] != staleStateUndetermined {
			t.Errorf("an unanswerable landing must read %q, got %v", staleStateUndetermined, got[0]["landed"])
		}
	})

	t.Run("exit-code-parsing", func(t *testing.T) {
		// The pure three-way mapping: 0 landed, 1 not landed, anything else
		// unanswerable — including a non-exit error (-1).
		for _, tc := range []struct {
			exit       int
			landed     bool
			determined bool
		}{
			{0, true, true},
			{1, false, true},
			{2, false, false},
			{128, false, false},
			{-1, false, false},
		} {
			landed, determined := sweepAncestryOutcome(tc.exit)
			if landed != tc.landed || determined != tc.determined {
				t.Errorf("exit %d: got (landed=%v, determined=%v), want (%v, %v)", tc.exit, landed, determined, tc.landed, tc.determined)
			}
		}
	})
}

// TestSweepNotLandedPreserves is AC-WS-002: a branch tip that is NOT an
// ancestor of the fetched base (real git, real ancestry) preserves with
// cause=not-landed — the one negative the landing predicate may assert.
func TestSweepNotLandedPreserves(t *testing.T) {
	f := newSweepRepo(t)
	wt := filepath.Join(f.base, "trees", "wt-unlanded")
	addSweepWorktree(t, f, wt, "feature/unlanded")
	// Commit work that exists nowhere on the remote.
	sweepRunGit(t, wt, "commit", "-q", "--allow-empty", "-m", "unpushed card work")
	withSweepRepoEnv(t, f)
	m := &sweepMock{}
	stubSweepProbeSeams(t, m, f)

	out, err := runSweepCmd(t, map[string]string{"json": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	got := decodeSweepJSON(t, out)
	if len(got) != 1 {
		t.Fatalf("expected 1 record, got %d:\n%s", len(got), out)
	}
	if got[0]["verdict"] != sweepPreserve {
		t.Errorf("an unlanded branch must preserve, got %v", got[0]["verdict"])
	}
	reason, _ := got[0]["reason"].(string)
	if !strings.Contains(reason, "cause="+causeNotLanded) {
		t.Errorf("reason must carry cause=%s, got %q", causeNotLanded, reason)
	}
	if got[0]["landed"] != staleStateNo {
		t.Errorf("a determined negative must read landed=%q, got %v", staleStateNo, got[0]["landed"])
	}
	if _, statErr := os.Stat(wt); statErr != nil {
		t.Errorf("the tree must survive, got: %v", statErr)
	}
}

// TestSweepLandedBranchClassification pins the affirmative half of the
// contract at classification level: a branch tip that IS an ancestor of the
// fetched origin/develop reads landed=yes with no cause reason. (The --yes
// disposal itself is AC-WS-001, asserted in TestSweepLandedBranchDisposes.)
func TestSweepLandedBranchClassification(t *testing.T) {
	f := newSweepRepo(t)
	wt := filepath.Join(f.base, "trees", "wt-landed")
	addSweepWorktree(t, f, wt, "feature/landed")
	withSweepRepoEnv(t, f)
	m := &sweepMock{}
	stubSweepProbeSeams(t, m, f)

	out, err := runSweepCmd(t, map[string]string{"json": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	got := decodeSweepJSON(t, out)
	if len(got) != 1 {
		t.Fatalf("expected 1 record, got %d:\n%s", len(got), out)
	}
	if got[0]["landed"] != staleStateYes {
		t.Errorf("a landed branch tip must read landed=%q, got %v (record: %v)", staleStateYes, got[0]["landed"], got[0])
	}
	reason, _ := got[0]["reason"].(string)
	if strings.Contains(reason, "cause="+causeNotLanded) || strings.Contains(reason, "cause="+causeFetchFailed) || strings.Contains(reason, "cause="+causeLandedCheckFailed) {
		t.Errorf("a landed branch must not carry a landing-failure cause, got %q", reason)
	}
}

// TestSweepBaseFlag is AC-WS-012: the base default is origin/develop, and
// --base overrides it. The fixture's feature tip is an ancestor of
// origin/main but NOT of origin/develop, so the two bases classify the same
// tree differently.
func TestSweepBaseFlag(t *testing.T) {
	f := newSweepRepo(t)
	wt := filepath.Join(f.base, "trees", "wt-bases")
	addSweepWorktree(t, f, wt, "feature/bases")
	// Land feature/bases into a local main and push it: origin/main contains
	// the feature tip; origin/develop (still at the seed) does not.
	sweepRunGit(t, wt, "commit", "-q", "--allow-empty", "-m", "base flag work")
	sweepRunGit(t, f.repo, "branch", "main", "develop")
	sweepRunGit(t, f.repo, "checkout", "-q", "main")
	sweepRunGit(t, f.repo, "merge", "-q", "--no-ff", "-m", "merge feature/bases", "feature/bases")
	sweepRunGit(t, f.repo, "push", "-q", "origin", "main")
	sweepRunGit(t, f.repo, "checkout", "-q", "develop")
	withSweepRepoEnv(t, f)
	m := &sweepMock{}
	stubSweepProbeSeams(t, m, f)

	out, err := runSweepCmd(t, map[string]string{"json": "true"})
	if err != nil {
		t.Fatalf("runSweep (bare, base default) error: %v", err)
	}
	got := decodeSweepJSON(t, out)
	if got[0]["landed"] != staleStateNo {
		t.Errorf("against the origin/develop default the tip must read not landed, got %v", got[0]["landed"])
	}
	reason, _ := got[0]["reason"].(string)
	if !strings.Contains(reason, "cause="+causeNotLanded) {
		t.Errorf("bare run must preserve with cause=%s, got %q", causeNotLanded, reason)
	}

	out, err = runSweepCmd(t, map[string]string{"json": "true", "base": "origin/main"})
	if err != nil {
		t.Fatalf("runSweep (--base origin/main) error: %v", err)
	}
	got = decodeSweepJSON(t, out)
	if got[0]["landed"] != staleStateYes {
		t.Errorf("against --base origin/main the tip must read landed, got %v (record: %v)", got[0]["landed"], got[0])
	}
}

// TestSweepUnpushedPreserves is AC-WS-005: a branch carrying commits
// unreachable from ANY remote preserves via the landing predicate's exit-1
// path. The behavior was RED'd by TestSweepNotLandedPreserves (M2); this is
// the AC-named pin adding the remote-unreachability observation.
func TestSweepUnpushedPreserves(t *testing.T) {
	f := newSweepRepo(t)
	wt := filepath.Join(f.base, "trees", "wt-unpushed")
	addSweepWorktree(t, f, wt, "feature/unpushed")
	sweepRunGit(t, wt, "commit", "-q", "--allow-empty", "-m", "work no remote has")
	// Positive control: the commit is genuinely unreachable from every
	// remote-tracking ref — exit 1 names it.
	if out := sweepRunGitAllowFail(t, wt, "branch", "-r", "--contains", "HEAD"); strings.TrimSpace(out) != "" {
		t.Fatalf("fixture broken: the commit is reachable from a remote: %s", out)
	}
	withSweepRepoEnv(t, f)
	m := &sweepMock{}
	stubSweepProbeSeams(t, m, f)

	out, err := runSweepCmd(t, map[string]string{"json": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	got := decodeSweepJSON(t, out)
	reason, _ := got[0]["reason"].(string)
	if !strings.Contains(reason, "cause="+causeNotLanded) {
		t.Errorf("an unpushed branch must preserve via cause=%s, got %q", causeNotLanded, reason)
	}
	if _, statErr := os.Stat(wt); statErr != nil {
		t.Errorf("the tree must survive, got: %v", statErr)
	}
}

// --- M3: the safety predicate composition -----------------------------------

// TestSweepDirtyPreserves is AC-WS-004: a landed, unanchored worktree with
// one untracked file preserves (dirty=yes), the file intact.
func TestSweepDirtyPreserves(t *testing.T) {
	f := newSweepRepo(t)
	wt := filepath.Join(f.base, "trees", "wt-dirty")
	addSweepWorktree(t, f, wt, "feature/dirty")
	untracked := filepath.Join(wt, "scratch.txt")
	if err := os.WriteFile(untracked, []byte("precious scratch\n"), 0o644); err != nil {
		t.Fatalf("write untracked file: %v", err)
	}
	withSweepRepoEnv(t, f)
	m := &sweepMock{}
	stubSweepProbeSeams(t, m, f)

	out, err := runSweepCmd(t, map[string]string{"json": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	got := decodeSweepJSON(t, out)
	if len(got) != 1 {
		t.Fatalf("expected 1 record, got %d:\n%s", len(got), out)
	}
	if got[0]["dirty"] != staleStateYes {
		t.Errorf("an untracked file must read dirty=%q, got %v", staleStateYes, got[0]["dirty"])
	}
	if got[0]["verdict"] != sweepPreserve {
		t.Errorf("a dirty tree must preserve, got %v", got[0]["verdict"])
	}
	if got[0]["landed"] != staleStateYes {
		t.Errorf("the branch is landed; landed must read %q (record: %v)", staleStateYes, got[0])
	}
	content, readErr := os.ReadFile(untracked)
	if readErr != nil || !strings.Contains(string(content), "precious") {
		t.Errorf("the untracked file must survive intact (read error: %v)", readErr)
	}
	if _, statErr := os.Stat(wt); statErr != nil {
		t.Errorf("the tree must survive, got: %v", statErr)
	}

	t.Run("ignored-content", func(t *testing.T) {
		f := newSweepRepo(t)
		wt := filepath.Join(f.base, "trees", "wt-ignored")
		addSweepWorktree(t, f, wt, "feature/ignored")
		// An irreplaceable gitignored file: the class both `git status
		// --porcelain` and non-forced removal disregard.
		exclude := filepath.Join(f.repo, ".git", "info", "exclude")
		fh, err := os.OpenFile(exclude, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatalf("open exclude: %v", err)
		}
		if _, err := fh.WriteString(".claude/agent-memory/\n"); err != nil {
			t.Fatalf("append exclude: %v", err)
		}
		_ = fh.Close()
		ignored := filepath.Join(wt, ".claude", "agent-memory")
		if err := os.MkdirAll(ignored, 0o755); err != nil {
			t.Fatalf("mkdir ignored dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(ignored, "topic.md"), []byte("memory\n"), 0o644); err != nil {
			t.Fatalf("write ignored file: %v", err)
		}
		withSweepRepoEnv(t, f)
		m := &sweepMock{}
		stubSweepProbeSeams(t, m, f)

		out, err := runSweepCmd(t, map[string]string{"json": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		got := decodeSweepJSON(t, out)
		if got[0]["ignored"] != staleStateYes {
			t.Errorf("irreplaceable ignored content must read ignored=%q, got %v", staleStateYes, got[0]["ignored"])
		}
		reason, _ := got[0]["reason"].(string)
		if !strings.Contains(reason, "cause="+causeIgnoredContent) {
			t.Errorf("reason must carry cause=%s, got %q", causeIgnoredContent, reason)
		}
	})
}

// TestSweepLockedAndAnchored is AC-WS-006: (a) a git-locked tree, (b) a tree
// named by the caller-registry anchor, (c) an unreadable lock source. (a)
// and (b) preserve naming the anchor source; (c) preserves EVERY tree with
// anchored=undetermined and ends with the distinguished exit-2 signal.
func TestSweepLockedAndAnchored(t *testing.T) {
	t.Run("git-lock", func(t *testing.T) {
		m := sweepMockEnv(t, []git.Worktree{{Path: "/wt/locked", Branch: "feature/locked"}})
		m.lockPorcelain = sweepLockPorcelain("/repo", "/wt/locked")

		out, err := runSweepCmd(t, map[string]string{"json": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		got := decodeSweepJSON(t, out)
		if got[0]["anchored"] != "lock" {
			t.Errorf("a git-locked tree must name the anchor source %q, got %v", "lock", got[0]["anchored"])
		}
		if got[0]["verdict"] != sweepPreserve {
			t.Errorf("a locked tree must preserve, got %v", got[0]["verdict"])
		}
		reason, _ := got[0]["reason"].(string)
		if !strings.Contains(reason, "cause="+causeAnchored) {
			t.Errorf("reason must carry cause=%s, got %q", causeAnchored, reason)
		}
	})

	t.Run("registry-anchor", func(t *testing.T) {
		launcher := t.TempDir() // the checkout a `moai cc -w` lane launched from
		sweepMockEnv(t, []git.Worktree{{Path: "/wt/anchored", Branch: "feature/anchored"}})
		// Re-point the caller registry at the launcher AFTER the env helper
		// sealed it: this cell needs the registry anchor to be observable.
		t.Setenv("CLAUDE_PROJECT_DIR", launcher)
		writeTreeRegistry(t, launcher, []session.Entry{anchoredEntry(t, "/wt/anchored", os.Getpid())})

		out, err := runSweepCmd(t, map[string]string{"json": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		got := decodeSweepJSON(t, out)
		if got[0]["anchored"] != "registry" {
			t.Errorf("a registry-anchored tree must name the anchor source %q, got %v", "registry", got[0]["anchored"])
		}
		if got[0]["verdict"] != sweepPreserve {
			t.Errorf("an anchored tree must preserve, got %v", got[0]["verdict"])
		}
	})

	t.Run("lock-source-unreadable", func(t *testing.T) {
		m := sweepMockEnv(t, []git.Worktree{
			{Path: "/wt/one", Branch: "feature/one"},
			{Path: "/wt/two", Branch: "feature/two"},
		})
		m.lockErr = errors.New("git worktree list failed")

		out, err := runSweepCmd(t, map[string]string{"json": "true"})
		var ec *ExitCodeError
		if err == nil || !errors.As(err, &ec) || ec.Code != 2 {
			t.Fatalf("a degraded run must end with the exit-2 preservation signal, got %v", err)
		}
		got := decodeSweepJSON(t, out)
		if len(got) != 2 {
			t.Fatalf("the report must complete in full (2 records), got %d:\n%s", len(got), out)
		}
		for _, r := range got {
			if r["anchored"] != staleStateUndetermined {
				t.Errorf("an unreadable lock source must leave anchored=%q (never a negative), got %v on %v", staleStateUndetermined, r["anchored"], r["path"])
			}
			if r["verdict"] != sweepPreserve {
				t.Errorf("every tree must preserve on a degraded run, got %v on %v", r["verdict"], r["path"])
			}
			reason, _ := r["reason"].(string)
			if !strings.Contains(reason, "cause="+causeLockSourceUnreadable) {
				t.Errorf("reason must carry cause=%s, got %q", causeLockSourceUnreadable, reason)
			}
		}
	})
}

// TestSweepProcessCWDPredicate is AC-WS-007: (a) a probe list containing the
// tree (or a path under it) occupies the tree; (b) a probe error is an
// unanswerable, never a negative. (c) the Windows stub is verified by the
// GOOS=windows build (sweep_cwd_windows_test.go runs on the Windows side).
func TestSweepProcessCWDPredicate(t *testing.T) {
	t.Run("cwd-inside-tree", func(t *testing.T) {
		tree := filepath.Join(t.TempDir(), "occupied")
		m := sweepMockEnv(t, []git.Worktree{{Path: tree, Branch: "feature/occupied"}})
		// A cwd deep INSIDE the tree counts — prefix match, not equality.
		m.cwds = []string{filepath.Join(t.TempDir(), "elsewhere"), filepath.Join(tree, "inner", "deeper")}
		m.landed["feature/occupied"] = true

		out, err := runSweepCmd(t, map[string]string{"json": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		got := decodeSweepJSON(t, out)
		if got[0]["cwd_occupied"] != staleStateYes {
			t.Errorf("a process cwd inside the tree must read cwd_occupied=%q, got %v", staleStateYes, got[0]["cwd_occupied"])
		}
		if got[0]["verdict"] != sweepPreserve {
			t.Errorf("an occupied tree must preserve, got %v", got[0]["verdict"])
		}
		reason, _ := got[0]["reason"].(string)
		if !strings.Contains(reason, "cause="+causeCWDOccupied) {
			t.Errorf("reason must carry cause=%s, got %q", causeCWDOccupied, reason)
		}
	})

	t.Run("probe-unanswerable", func(t *testing.T) {
		m := sweepMockEnv(t, []git.Worktree{{Path: filepath.Join(t.TempDir(), "unprobed"), Branch: "feature/unprobed"}})
		m.cwdErr = errors.New("lsof: command not found")
		m.landed["feature/unprobed"] = true

		out, err := runSweepCmd(t, map[string]string{"json": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		got := decodeSweepJSON(t, out)
		if got[0]["cwd_occupied"] != staleStateUndetermined {
			t.Errorf("an unanswerable probe must read cwd_occupied=%q (never a negative), got %v", staleStateUndetermined, got[0]["cwd_occupied"])
		}
		reason, _ := got[0]["reason"].(string)
		if !strings.Contains(reason, "cause="+causeCWDProbeFailed) {
			t.Errorf("reason must carry cause=%s, got %q", causeCWDProbeFailed, reason)
		}
		if got[0]["verdict"] != sweepPreserve {
			t.Errorf("an unanswerable probe must preserve, got %v", got[0]["verdict"])
		}
	})
}

// TestSweepVerdictRecord is AC-WS-013: on a fully-evaluated record every
// field is present and every predicate reads affirmatively — an unobserved
// predicate is never rendered as a negative.
func TestSweepVerdictRecord(t *testing.T) {
	f := newSweepRepo(t)
	wt := filepath.Join(f.base, "trees", "wt-record")
	addSweepWorktree(t, f, wt, "feature/record")
	withSweepRepoEnv(t, f)
	m := &sweepMock{}
	stubSweepProbeSeams(t, m, f)

	out, err := runSweepCmd(t, map[string]string{"json": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	got := decodeSweepJSON(t, out)
	if len(got) != 1 {
		t.Fatalf("expected 1 record, got %d:\n%s", len(got), out)
	}
	r := got[0]
	for _, field := range []string{"path", "branch", "tier", "verdict", "reason", "landed", "dirty", "ignored", "anchored", "cwd_occupied", "on_base"} {
		if _, ok := r[field]; !ok {
			t.Errorf("record is missing field %q (record: %v)", field, r)
		}
	}
	for _, field := range []string{"landed", "dirty", "ignored", "anchored", "cwd_occupied", "on_base"} {
		if r[field] == staleStateUndetermined || r[field] == staleStateNotChecked {
			t.Errorf("a fully-evaluated record must not carry %q on %s (record: %v)", r[field], field, r)
		}
	}
	if r["verdict"] != sweepDispose {
		t.Errorf("a fully-affirmative tree must classify as %s, got %v (record: %v)", sweepDispose, r["verdict"], r)
	}
	if reason, _ := r["reason"].(string); reason != "" {
		t.Errorf("a DISPOSE record carries an empty reason, got %q", reason)
	}
}

// --- M4: tier routing and the apply path ------------------------------------

// recordingProvider decorates the real git provider so a test can observe
// removal calls (and force failures) without reimplementing git.
type recordingProvider struct {
	git.WorktreeManager
	removed  []string
	failFor  map[string]error
	onRemove func() // observation hook, called at removal time
}

func (r *recordingProvider) Remove(path string, force bool) error {
	r.removed = append(r.removed, path)
	if r.onRemove != nil {
		r.onRemove()
	}
	if err := r.failFor[path]; err != nil {
		return err
	}
	return r.WorktreeManager.Remove(path, force)
}

// TestSweepLandedBranchDisposes is AC-WS-001: a landed, clean, unanchored,
// probe-negative worktree is disposed with --yes; the branch ref survives
// (REQ-WS-010); every predicate field reads affirmatively.
func TestSweepLandedBranchDisposes(t *testing.T) {
	f := newSweepRepo(t)
	wt := filepath.Join(f.base, "trees", "wt-dispose")
	addSweepWorktree(t, f, wt, "feature/dispose")
	withSweepRepoEnv(t, f)
	m := &sweepMock{}
	stubSweepProbeSeams(t, m, f)

	out, err := runSweepCmd(t, map[string]string{"yes": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	if _, statErr := os.Stat(wt); !os.IsNotExist(statErr) {
		t.Fatalf("the disposable tree must be gone, stat error: %v\noutput:\n%s", statErr, out)
	}
	// REQ-WS-010: the branch is never deleted.
	if branchList := sweepRunGitAllowFail(t, f.repo, "rev-parse", "--verify", "feature/dispose"); strings.TrimSpace(branchList) == "" {
		t.Error("the branch ref must still resolve after disposal")
	}
	if !strings.Contains(out, "Removed 1 worktree(s). Branches were left intact.") {
		t.Errorf("expected the removal summary, got:\n%s", out)
	}

	// The --json inventory of the same fixture carries the affirmative
	// record (one evaluation, two renderings).
	out, err = runSweepCmd(t, map[string]string{"json": "true"})
	if err != nil {
		t.Fatalf("runSweep --json error: %v", err)
	}
	got := decodeSweepJSON(t, out)
	if len(got) != 0 {
		t.Errorf("the disposed tree is gone; the inventory must be empty, got %v", got)
	}
}

// sweepRunGitAllowFail runs git, returning its output even on failure (the
// caller asserts on empty output instead of failing the fixture).
func sweepRunGitAllowFail(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return string(out)
}

// TestSweepDryRunDefaultAndYes is AC-WS-008: bare previews and removes
// nothing; --yes removes exactly the disposable tree; --json removes nothing
// and carries the same predicate values the text report showed.
func TestSweepDryRunDefaultAndYes(t *testing.T) {
	buildFixture := func(t *testing.T) (sweepFixture, string, string) {
		f := newSweepRepo(t)
		disposable := filepath.Join(f.base, "trees", "wt-dry-disposable")
		addSweepWorktree(t, f, disposable, "feature/dry-disposable")
		preserved := filepath.Join(f.base, "trees", "wt-dry-kept")
		addSweepWorktree(t, f, preserved, "feature/dry-kept")
		if err := os.WriteFile(filepath.Join(preserved, "uncommitted.txt"), []byte("keep me\n"), 0o644); err != nil {
			t.Fatalf("dirty the preserved tree: %v", err)
		}
		withSweepRepoEnv(t, f)
		m := &sweepMock{}
		stubSweepProbeSeams(t, m, f)
		return f, disposable, preserved
	}

	t.Run("bare-previews-and-removes-nothing", func(t *testing.T) {
		_, disposable, _ := buildFixture(t)
		out, err := runSweepCmd(t, map[string]string{})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		if _, statErr := os.Stat(disposable); statErr != nil {
			t.Fatalf("a dry run must remove nothing, got: %v", statErr)
		}
		if !strings.Contains(out, "Would remove 1 worktree(s)") {
			t.Errorf("expected the preview header, got:\n%s", out)
		}
		if !strings.Contains(out, "Re-run with --yes") {
			t.Errorf("expected the preview to point at --yes, got:\n%s", out)
		}
	})

	t.Run("yes-removes-exactly-the-disposable", func(t *testing.T) {
		_, disposable, preserved := buildFixture(t)
		out, err := runSweepCmd(t, map[string]string{"yes": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		if _, statErr := os.Stat(disposable); !os.IsNotExist(statErr) {
			t.Fatalf("the disposable tree must be gone, stat error: %v", statErr)
		}
		if _, statErr := os.Stat(preserved); statErr != nil {
			t.Fatalf("the preserved tree must survive, got: %v", statErr)
		}
		if !strings.Contains(out, "Removed 1 worktree(s)") {
			t.Errorf("expected the removal summary, got:\n%s", out)
		}
	})

	t.Run("json-parity", func(t *testing.T) {
		_, disposable, _ := buildFixture(t)
		out, err := runSweepCmd(t, map[string]string{"json": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		if _, statErr := os.Stat(disposable); statErr != nil {
			t.Fatalf("--json must remove nothing, got: %v", statErr)
		}
		got := decodeSweepJSON(t, out)
		if len(got) != 2 {
			t.Fatalf("expected records for both trees, got %d:\n%s", len(got), out)
		}
		for _, r := range got {
			if r["branch"] == "feature/dry-disposable" {
				if r["verdict"] != sweepDispose {
					t.Errorf("the disposable tree must classify as %s in the inventory, got %v", sweepDispose, r["verdict"])
				}
			}
			if r["branch"] == "feature/dry-kept" {
				if r["verdict"] != sweepPreserve || r["dirty"] != staleStateYes {
					t.Errorf("the dirty tree must read PRESERVE/dirty=yes in the inventory, got %v", r)
				}
			}
		}
		assertSweepStateValues(t, got)
	})
}

// TestSweepL1HoistThenRemove is AC-WS-009: an L1 tree is hoisted BEFORE
// removal (order asserted), its evidence lands in the project root, and a
// hoist failure preserves the tree with cause=hoist-failed.
func TestSweepL1HoistThenRemove(t *testing.T) {
	t.Run("hoist-before-remove", func(t *testing.T) {
		f := newSweepRepo(t)
		l1 := filepath.Join(f.repo, ".claude", "worktrees", "sweep-l1")
		addSweepWorktree(t, f, l1, "feature/sweep-l1")
		reports := filepath.Join(l1, ".moai", "reports")
		if err := os.MkdirAll(reports, 0o755); err != nil {
			t.Fatalf("mkdir reports: %v", err)
		}
		if err := os.WriteFile(filepath.Join(reports, "verdict.md"), []byte("evidence\n"), 0o644); err != nil {
			t.Fatalf("write evidence: %v", err)
		}
		withSweepRepoEnv(t, f)
		m := &sweepMock{}
		stubSweepProbeSeams(t, m, f)

		provider := &recordingProvider{WorktreeManager: WorktreeProvider}
		WorktreeProvider = provider

		// One shared sequence records both seams, so the ORDER is evidence:
		// hoist must appear before remove.
		var order []string
		origHoist := sweepHoistEvidence
		sweepHoistEvidence = func(w io.Writer, path string) (bool, error) {
			order = append(order, "hoist")
			return origHoist(w, path)
		}
		provider.onRemove = func() { order = append(order, "remove") }
		t.Cleanup(func() { sweepHoistEvidence = origHoist })

		out, err := runSweepCmd(t, map[string]string{"yes": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}

		if _, statErr := os.Stat(l1); !os.IsNotExist(statErr) {
			t.Fatalf("the L1 tree must be gone, stat error: %v\noutput:\n%s", statErr, out)
		}
		if len(order) != 2 || order[0] != "hoist" || order[1] != "remove" {
			t.Fatalf("the hoist seam must record BEFORE the removal seam, got order %v", order)
		}
		if len(provider.removed) != 1 {
			t.Fatalf("exactly one removal must happen, got %v", provider.removed)
		}
		// The hoisted evidence survives the removal at the project root.
		hoisted := filepath.Join(f.repo, ".moai", "reports", "worktrees", "sweep-l1", "verdict.md")
		content, readErr := os.ReadFile(hoisted)
		if readErr != nil || !strings.Contains(string(content), "evidence") {
			t.Errorf("the evidence must exist under the project root after disposal (read: %v)", readErr)
		}
		if !strings.Contains(out, "Removed 1 worktree(s)") {
			t.Errorf("expected the removal summary, got:\n%s", out)
		}
	})

	t.Run("hoist-failure-preserves", func(t *testing.T) {
		f := newSweepRepo(t)
		l1 := filepath.Join(f.repo, ".claude", "worktrees", "sweep-l1-fail")
		addSweepWorktree(t, f, l1, "feature/sweep-l1-fail")
		withSweepRepoEnv(t, f)
		m := &sweepMock{}
		stubSweepProbeSeams(t, m, f)

		origHoist := sweepHoistEvidence
		sweepHoistEvidence = func(_ io.Writer, _ string) (bool, error) {
			return false, errors.New("cannot copy evidence")
		}
		t.Cleanup(func() { sweepHoistEvidence = origHoist })

		out, err := runSweepCmd(t, map[string]string{"yes": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		if _, statErr := os.Stat(l1); statErr != nil {
			t.Fatalf("a hoist failure must preserve the tree, got: %v", statErr)
		}
		if !strings.Contains(out, "cause="+causeHoistFailed) {
			t.Errorf("the notice must carry cause=%s, got:\n%s", causeHoistFailed, out)
		}
	})

	// Repair round F1: a PARTIAL retrieval must preserve. A different-content
	// file pre-planted at the hoist destination is skipped by REQ-RLC-006's
	// never-overwrite policy — the skipped file would be the tree's sole
	// copy, so the sweep keeps the tree (cause=hoist-partial).
	t.Run("destination-conflict-preserves", func(t *testing.T) {
		f := newSweepRepo(t)
		l1 := filepath.Join(f.repo, ".claude", "worktrees", "sweep-l1-conflict")
		addSweepWorktree(t, f, l1, "feature/sweep-l1-conflict")
		reports := filepath.Join(l1, ".moai", "reports")
		if err := os.MkdirAll(reports, 0o755); err != nil {
			t.Fatalf("mkdir reports: %v", err)
		}
		if err := os.WriteFile(filepath.Join(reports, "verdict.md"), []byte("fresh evidence\n"), 0o644); err != nil {
			t.Fatalf("write evidence: %v", err)
		}
		dest := filepath.Join(f.repo, ".moai", "reports", "worktrees", "sweep-l1-conflict")
		if err := os.MkdirAll(dest, 0o755); err != nil {
			t.Fatalf("mkdir destination: %v", err)
		}
		destFile := filepath.Join(dest, "verdict.md")
		if err := os.WriteFile(destFile, []byte("stale different content\n"), 0o644); err != nil {
			t.Fatalf("pre-plant destination conflict: %v", err)
		}
		withSweepRepoEnv(t, f)
		m := &sweepMock{}
		stubSweepProbeSeams(t, m, f)

		out, err := runSweepCmd(t, map[string]string{"yes": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		if _, statErr := os.Stat(l1); statErr != nil {
			t.Fatalf("a partially-retrieved tree must be preserved, got: %v", statErr)
		}
		if !strings.Contains(out, "cause="+causeHoistPartial) {
			t.Errorf("the notice must carry cause=%s, got:\n%s", causeHoistPartial, out)
		}
		content, readErr := os.ReadFile(destFile)
		if readErr != nil || string(content) != "stale different content\n" {
			t.Errorf("the destination conflict must be left untouched (read: %v)", readErr)
		}
	})

	// Repair round F1: an UNATTEMPTED retrieval must preserve. The main-root
	// resolver succeeds exactly once — for the L1 tier classification — then
	// fails, so the hoist cannot even find a destination.
	t.Run("root-unresolved-preserves", func(t *testing.T) {
		f := newSweepRepo(t)
		l1 := filepath.Join(f.repo, ".claude", "worktrees", "sweep-l1-rootless")
		addSweepWorktree(t, f, l1, "feature/sweep-l1-rootless")
		reports := filepath.Join(l1, ".moai", "reports")
		if err := os.MkdirAll(reports, 0o755); err != nil {
			t.Fatalf("mkdir reports: %v", err)
		}
		if err := os.WriteFile(filepath.Join(reports, "verdict.md"), []byte("evidence\n"), 0o644); err != nil {
			t.Fatalf("write evidence: %v", err)
		}
		withSweepRepoEnv(t, f)
		m := &sweepMock{}
		stubSweepProbeSeams(t, m, f)

		calls := 0
		origResolver := gitMainRootFromTargetFunc
		gitMainRootFromTargetFunc = func(string) (string, error) {
			calls++
			if calls == 1 {
				return f.repo, nil
			}
			return "", errors.New("project root unresolved")
		}
		t.Cleanup(func() { gitMainRootFromTargetFunc = origResolver })

		out, err := runSweepCmd(t, map[string]string{"yes": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		if _, statErr := os.Stat(l1); statErr != nil {
			t.Fatalf("a tree whose evidence retrieval was unattempted must be preserved, got: %v", statErr)
		}
		if !strings.Contains(out, "cause="+causeHoistPartial) {
			t.Errorf("the notice must carry cause=%s, got:\n%s", causeHoistPartial, out)
		}
	})
}

// TestSweepL2DonePath is AC-WS-010: an L2 tree disposes through the done
// removal core (hoist inside, non-forced, no branch deletion); an L1 tree in
// the same run never reaches the done core.
func TestSweepL2DonePath(t *testing.T) {
	f := newSweepRepo(t)
	l2 := filepath.Join(f.base, "trees", "sweep-l2")
	addSweepWorktree(t, f, l2, "feature/sweep-l2")
	l1 := filepath.Join(f.repo, ".claude", "worktrees", "sweep-l1-mix")
	addSweepWorktree(t, f, l1, "feature/sweep-l1-mix")
	withSweepRepoEnv(t, f)
	m := &sweepMock{}
	stubSweepProbeSeams(t, m, f)

	var doneCalls []string
	origDone := sweepDoneCleanup
	sweepDoneCleanup = func(branch string, force, deleteBranch, hoist bool) (bool, error) {
		doneCalls = append(doneCalls, sweepDoneCallKey(branch, force, deleteBranch, hoist))
		return origDone(branch, force, deleteBranch, hoist)
	}
	t.Cleanup(func() { sweepDoneCleanup = origDone })

	out, err := runSweepCmd(t, map[string]string{"yes": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}

	if _, statErr := os.Stat(l2); !os.IsNotExist(statErr) {
		t.Fatalf("the L2 tree must be gone, stat error: %v\noutput:\n%s", statErr, out)
	}
	// The done core saw exactly the L2 branch, non-forced, no branch
	// deletion, hoist on.
	if len(doneCalls) != 1 || doneCalls[0] != sweepDoneCallKey("feature/sweep-l2", false, false, true) {
		t.Fatalf("the done core must be called once with (force=false, deleteBranch=false, hoist=true), got %v", doneCalls)
	}
	// The branch survives (no deletion inside the done core either).
	if branchList := sweepRunGitAllowFail(t, f.repo, "rev-parse", "--verify", "feature/sweep-l2"); strings.TrimSpace(branchList) == "" {
		t.Error("the L2 branch must survive the done-core disposal")
	}
}

// TestSweepL2PartialRetrievalPreserves is repair round F1's L2 equivalent:
// the done core reports only hard hoist failures, so the sweep verifies full
// retrieval sweep-side BEFORE delegating — a partial or unattempted
// retrieval preserves the tree and the done core is never reached.
func TestSweepL2PartialRetrievalPreserves(t *testing.T) {
	t.Run("partial-retrieval", func(t *testing.T) {
		m := sweepMockEnv(t, []git.Worktree{{Path: "/wt/l2partial", Branch: "feature/l2partial"}})
		m.landed["feature/l2partial"] = true
		m.hoistComplete = false

		out, err := runSweepCmd(t, map[string]string{"yes": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		if len(m.doneCalls) != 0 {
			t.Errorf("the done core must not be reached on an incomplete retrieval, got %v", m.doneCalls)
		}
		if len(m.removed) != 0 {
			t.Errorf("a partially-retrieved tree must not be removed, removed %v", m.removed)
		}
		if !strings.Contains(out, "cause="+causeHoistPartial) {
			t.Errorf("the notice must carry cause=%s, got:\n%s", causeHoistPartial, out)
		}
	})

	t.Run("root-unresolved", func(t *testing.T) {
		f := newSweepRepo(t)
		l2 := filepath.Join(f.base, "trees", "sweep-l2-rootless")
		addSweepWorktree(t, f, l2, "feature/sweep-l2-rootless")
		reports := filepath.Join(l2, ".moai", "reports")
		if err := os.MkdirAll(reports, 0o755); err != nil {
			t.Fatalf("mkdir reports: %v", err)
		}
		if err := os.WriteFile(filepath.Join(reports, "verdict.md"), []byte("evidence\n"), 0o644); err != nil {
			t.Fatalf("write evidence: %v", err)
		}
		withSweepRepoEnv(t, f)
		m := &sweepMock{}
		stubSweepProbeSeams(t, m, f)

		calls := 0
		origResolver := gitMainRootFromTargetFunc
		gitMainRootFromTargetFunc = func(string) (string, error) {
			calls++
			if calls == 1 {
				return f.repo, nil
			}
			return "", errors.New("project root unresolved")
		}
		t.Cleanup(func() { gitMainRootFromTargetFunc = origResolver })

		var doneCalls []string
		origDone := sweepDoneCleanup
		sweepDoneCleanup = func(branch string, force, deleteBranch, hoist bool) (bool, error) {
			doneCalls = append(doneCalls, branch)
			return origDone(branch, force, deleteBranch, hoist)
		}
		t.Cleanup(func() { sweepDoneCleanup = origDone })

		out, err := runSweepCmd(t, map[string]string{"yes": "true"})
		if err != nil {
			t.Fatalf("runSweep error: %v", err)
		}
		if _, statErr := os.Stat(l2); statErr != nil {
			t.Fatalf("a tree whose evidence retrieval was unattempted must be preserved, got: %v", statErr)
		}
		if len(doneCalls) != 0 {
			t.Errorf("the done core must not be reached on an unattempted retrieval, got %v", doneCalls)
		}
		if !strings.Contains(out, "cause="+causeHoistPartial) {
			t.Errorf("the notice must carry cause=%s, got:\n%s", causeHoistPartial, out)
		}
	})
}

// TestSweepNeverDispose is AC-WS-011: the main checkout and the process's
// own tree are absent from the records entirely; a base-branch checkout and
// a locked tree preserve; a healthy candidate still disposes, proving the
// list protects rather than paralyzes.
func TestSweepNeverDispose(t *testing.T) {
	f := newSweepRepo(t)
	// Move the main checkout OFF develop so a linked worktree can hold the
	// base branch (git allows a branch in exactly one worktree).
	sweepRunGit(t, f.repo, "checkout", "-q", "-b", "hold")

	baseWt := filepath.Join(f.base, "trees", "wt-on-base")
	sweepRunGit(t, f.repo, "worktree", "add", "--", baseWt, "develop")

	lockedWt := filepath.Join(f.base, "trees", "wt-locked")
	addSweepWorktree(t, f, lockedWt, "feature/locked-wt")
	sweepRunGit(t, f.repo, "worktree", "lock", lockedWt)

	ownWt := filepath.Join(f.base, "trees", "wt-own")
	addSweepWorktree(t, f, ownWt, "feature/own-wt")

	disposeWt := filepath.Join(f.base, "trees", "wt-fine")
	addSweepWorktree(t, f, disposeWt, "feature/fine")

	withSweepRepoEnv(t, f)
	m := &sweepMock{}
	stubSweepProbeSeams(t, m, f)

	// Move the process INTO one of the fixture's trees: that tree is the
	// worktree this command runs in — protected by definition.
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(ownWt); err != nil {
		t.Fatalf("chdir into the own tree: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	out, err := runSweepCmd(t, map[string]string{"json": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	got := decodeSweepJSON(t, out)
	// The apply run is a separate invocation: --json removes nothing by
	// contract, even with --yes.
	if _, err := runSweepCmd(t, map[string]string{"yes": "true"}); err != nil {
		t.Fatalf("apply run error: %v", err)
	}
	byBranch := map[string]map[string]interface{}{}
	for _, r := range got {
		branch, _ := r["branch"].(string)
		byBranch[branch] = r
	}
	// The main checkout (hold) and the process's own tree are OUTSIDE the
	// universe — absent, not reported as kept.
	for _, absent := range []string{"hold", "feature/own-wt"} {
		if _, ok := byBranch[absent]; ok {
			t.Errorf("protected tree %q must be absent from the records, got %v", absent, byBranch[absent])
		}
	}
	// The base-branch checkout preserves, naming the cause.
	if r, ok := byBranch["develop"]; !ok {
		t.Error("the base-branch checkout must appear in the records")
	} else {
		if r["on_base"] != staleStateYes || r["verdict"] != sweepPreserve {
			t.Errorf("the base-branch checkout must read on_base=yes and PRESERVE, got %v", r)
		}
		reason, _ := r["reason"].(string)
		if !strings.Contains(reason, "cause="+causeOnBaseBranch) {
			t.Errorf("reason must carry cause=%s, got %q", causeOnBaseBranch, reason)
		}
	}
	// The locked tree preserves with the lock source named.
	if r, ok := byBranch["feature/locked-wt"]; !ok {
		t.Error("the locked tree must appear in the records")
	} else if r["anchored"] != "lock" || r["verdict"] != sweepPreserve {
		t.Errorf("a locked tree must read anchored=lock and PRESERVE, got %v", r)
	}
	// The healthy candidate still flows through.
	if r, ok := byBranch["feature/fine"]; !ok {
		t.Error("the healthy candidate must appear in the records")
	} else if r["verdict"] != sweepDispose {
		t.Errorf("a clean landed tree must be disposable, got %v", r)
	}

	// Survivals: locked, base-branch, own tree, main checkout.
	for _, p := range []string{lockedWt, baseWt, ownWt, f.repo} {
		if _, statErr := os.Stat(p); statErr != nil {
			t.Errorf("protected tree %s must survive, got: %v", p, statErr)
		}
	}
	// And the healthy candidate is gone.
	if _, statErr := os.Stat(disposeWt); !os.IsNotExist(statErr) {
		t.Errorf("the healthy candidate must be disposed, stat error: %v", statErr)
	}
}

// TestSweepRemovalFailure is AC-WS-014 (a): a failing removal is a
// non-blocking notice; the remaining trees still process; the run exits 0
// (removal failure is not the degraded-anchor signal).
func TestSweepRemovalFailure(t *testing.T) {
	f := newSweepRepo(t)
	failing := filepath.Join(f.base, "trees", "wt-fail")
	addSweepWorktree(t, f, failing, "feature/removal-fails")
	second := filepath.Join(f.base, "trees", "wt-second")
	addSweepWorktree(t, f, second, "feature/removal-second")
	withSweepRepoEnv(t, f)
	m := &sweepMock{}
	stubSweepProbeSeams(t, m, f)

	provider := &recordingProvider{
		WorktreeManager: WorktreeProvider,
		failFor:         map[string]error{failing: errors.New("git worktree remove refused")},
	}
	WorktreeProvider = provider

	out, err := runSweepCmd(t, map[string]string{"yes": "true"})
	if err != nil {
		t.Fatalf("a removal failure is non-blocking; got: %v", err)
	}
	if !strings.Contains(out, "could not remove") {
		t.Errorf("the failure must surface as a non-blocking notice, got:\n%s", out)
	}
	if _, statErr := os.Stat(failing); statErr != nil {
		t.Errorf("the failed tree must survive, got: %v", statErr)
	}
	if _, statErr := os.Stat(second); !os.IsNotExist(statErr) {
		t.Errorf("the remaining tree must still be processed, stat error: %v", statErr)
	}
	if !strings.Contains(out, "Removed 1 worktree(s)") {
		t.Errorf("the count must report only the actual removal, got:\n%s", out)
	}
}

// --- REFACTOR-phase edge-path pins ------------------------------------------
//
// These pin failure branches and the acceptance §D.2 edge cases. Each
// behavior's RED evidence lives in its milestone (M1-M4); these cells harden
// the error paths and pin the remaining edge vocabulary.

// TestSweepDetachedHeadPreserves pins acceptance §D.2: a detached-HEAD tree
// has no branch tip, so the landing predicate is unanswerable — PRESERVE,
// landed=undetermined, never a negative.
func TestSweepDetachedHeadPreserves(t *testing.T) {
	sweepMockEnv(t, []git.Worktree{{Path: "/wt/detached", Branch: ""}})

	out, err := runSweepCmd(t, map[string]string{"json": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	got := decodeSweepJSON(t, out)
	if got[0]["landed"] != staleStateUndetermined {
		t.Errorf("a detached HEAD must read landed=undetermined, got %v", got[0]["landed"])
	}
	reason, _ := got[0]["reason"].(string)
	if !strings.Contains(reason, "cause="+causeDetachedHead) {
		t.Errorf("reason must carry cause=%s, got %q", causeDetachedHead, reason)
	}
	if got[0]["verdict"] != sweepPreserve {
		t.Errorf("a detached-HEAD tree must preserve, got %v", got[0]["verdict"])
	}
}

// TestSweepDirtyCheckFailurePreserves pins the dirty predicate's unanswerable
// branch: an unreadable working-tree state is undetermined, never a negative.
func TestSweepDirtyCheckFailurePreserves(t *testing.T) {
	m := sweepMockEnv(t, []git.Worktree{{Path: "/wt/unreadable", Branch: "feature/unreadable"}})
	m.landed["feature/unreadable"] = true
	m.statusErr = errors.New("git status died")

	out, err := runSweepCmd(t, map[string]string{"json": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	got := decodeSweepJSON(t, out)
	if got[0]["dirty"] != staleStateUndetermined {
		t.Errorf("an unreadable dirty check must read undetermined, got %v", got[0]["dirty"])
	}
	reason, _ := got[0]["reason"].(string)
	if !strings.Contains(reason, "cause="+causeDirtyCheckFailed) {
		t.Errorf("reason must carry cause=%s, got %q", causeDirtyCheckFailed, reason)
	}
}

// TestSweepFetchErrorDefaultImpl pins the DEFAULT fetch seam's failure path
// and its remote/ref derivation: the chain default (origin/main in the bare
// mock root) fetches `git -C <root> fetch origin main`; the failure makes
// every tree's landing unanswerable.
func TestSweepFetchErrorDefaultImpl(t *testing.T) {
	defaultFetch := sweepFetchBase // capture before sweepMockEnv replaces it
	m := sweepMockEnv(t, []git.Worktree{{Path: "/wt/unfetched2", Branch: "feature/unfetched2"}})
	sweepFetchBase = defaultFetch
	m.fetchCmdErr = errors.New("network down")

	out, err := runSweepCmd(t, map[string]string{"json": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	if len(m.fetchArgs) != 1 {
		t.Fatalf("the default fetch must run exactly once, got %v", m.fetchArgs)
	}
	want := []string{"-C", "/repo", "fetch", "origin", "main"}
	for i, a := range want {
		if m.fetchArgs[0][i] != a {
			t.Fatalf("fetch invocation = %v, want %v", m.fetchArgs[0], want)
		}
	}
	got := decodeSweepJSON(t, out)
	reason, _ := got[0]["reason"].(string)
	if !strings.Contains(reason, "cause="+causeFetchFailed) {
		t.Errorf("reason must carry cause=%s, got %q", causeFetchFailed, reason)
	}
}

// TestSweepProcessExitCode pins the exit-code extractor: a real subprocess
// exit error reports its code; a non-exit error reports -1 (unanswerable).
func TestSweepProcessExitCode(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "rev-parse", "--verify", "no-such-ref-for-exitcode-test")
	cmd.Dir = dir
	_, err := cmd.Output()
	if err == nil {
		t.Fatal("fixture broken: expected a non-zero git exit")
	}
	if code := sweepProcessExitCode(err); code != 128 {
		t.Errorf("a real git failure must report exit 128, got %d", code)
	}
	if code := sweepProcessExitCode(errors.New("not an exit error")); code != -1 {
		t.Errorf("a non-exit error must report -1, got %d", code)
	}
}

// TestSweepRemovalTimeIgnoredKeep pins the removal-time re-read: a tree that
// classified clean but holds irreplaceable ignored content by the time its
// removal turn comes is kept with the "observed at removal time" notice.
func TestSweepRemovalTimeIgnoredKeep(t *testing.T) {
	m := sweepMockEnv(t, []git.Worktree{{Path: "/wt/lateignored", Branch: "feature/lateignored"}})
	m.landed["feature/lateignored"] = true
	m.ignoredAfter = 1 // the classification read is clean; the removal re-read is not
	m.ignoredOut["/wt/lateignored"] = "!! .claude/agent-memory/\n"

	out, err := runSweepCmd(t, map[string]string{"yes": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	if len(m.removed) != 0 {
		t.Fatalf("a tree acquiring irreplaceable ignored content must be kept, removed %v", m.removed)
	}
	if !strings.Contains(out, "observed at removal time") {
		t.Errorf("expected the removal-time keep notice, got:\n%s", out)
	}
}

// TestSweepDoneCoreKeptNotice pins the done-core's success=false branch: the
// core can keep a tree (its own anchor refusal) — the sweep reports it as a
// keep notice, not a failure.
func TestSweepDoneCoreKeptNotice(t *testing.T) {
	m := sweepMockEnv(t, []git.Worktree{{Path: "/wt/donekept", Branch: "feature/donekept"}})
	m.landed["feature/donekept"] = true
	m.doneSuccess = false

	out, err := runSweepCmd(t, map[string]string{"yes": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	if len(m.removed) != 0 {
		t.Fatalf("a kept tree must not be counted removed, removed %v", m.removed)
	}
	if !strings.Contains(out, "the done core kept the tree") {
		t.Errorf("expected the done-core keep notice, got:\n%s", out)
	}
}

// TestSweepL1RemoveErrorWarning pins the L1 apply path's removal failure: a
// warning notice, the tree survives, hoist already ran (evidence safe).
func TestSweepL1RemoveErrorWarning(t *testing.T) {
	mainRoot := t.TempDir()
	l1Path := filepath.Join(mainRoot, ".claude", "worktrees", "swepl1fail")
	m := sweepMockEnv(t, []git.Worktree{{Path: l1Path, Branch: "feature/l1fail"}})
	m.landed["feature/l1fail"] = true
	m.lockPorcelain = sweepLockPorcelain(mainRoot) // the fallback main-root resolver reads this
	m.removeErr[l1Path] = errors.New("git worktree remove refused")

	out, err := runSweepCmd(t, map[string]string{"yes": "true"})
	if err != nil {
		t.Fatalf("runSweep error: %v", err)
	}
	if !strings.Contains(out, "could not remove") {
		t.Errorf("expected the removal-failure warning, got:\n%s", out)
	}
	if len(m.hoistCalls) != 1 {
		t.Errorf("the hoist seam must have run before the failed removal, got %v", m.hoistCalls)
	}
}

// TestSweepNilProviderErrors pins the uninitialized-provider guard.
func TestSweepNilProviderErrors(t *testing.T) {
	orig := WorktreeProvider
	WorktreeProvider = nil
	t.Cleanup(func() { WorktreeProvider = orig })

	cmd := newSweepCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.RunE(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("a nil provider must error, got: %v", err)
	}
}

// keep unused-import guards honest for the milestone-gated test growth.
var (
	_ = time.Now
	_ = session.Entry{}
)
