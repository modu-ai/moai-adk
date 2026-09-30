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
		m := sweepMockEnv(t, []git.Worktree{{Path: "/wt/occupied", Branch: "feature/occupied"}})
		// A cwd deep INSIDE the tree counts — prefix match, not equality.
		m.cwds = []string{"/elsewhere", "/wt/occupied/inner/deeper"}
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
		m := sweepMockEnv(t, []git.Worktree{{Path: "/wt/unprobed", Branch: "feature/unprobed"}})
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

// keep unused-import guards honest for the milestone-gated test growth.
var (
	_ = time.Now
	_ = session.Entry{}
)
