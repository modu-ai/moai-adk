package cli

// session_worktree_automerge_test.go — SPEC-WORKTREE-KEY-WIRING-001 M1/M2
// tests. One test per acceptance criterion (AC-WKW-001..010, 013, 014), each
// named after its AC and executing its acceptance.md judging selector.
//
// Two verification styles, deliberately mixed:
//   - seam-injected (the family's house style): every git invocation and both
//     lock calls go through function-variable seams, which lets the tests
//     record the exact subcommands the path executed — the mechanism of the
//     AC-WKW-006 zero-push assertion.
//   - real-git fixtures (t.TempDir + git init): the happy path and the
//     conflict path assert what a fake merge cannot — that develop's HEAD is a
//     two-parent merge commit naming the session branch (AC-WKW-002), and
//     that a conflicted merge leaves no MERGE_HEAD behind (AC-WKW-007).

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// amSeams snapshots the auto-merge seams so a test can swap them and restore
// on cleanup (same pattern as swSeams / prMergeSeams).
type amSeams struct {
	lockRoot    func() string
	sessionID   func() string
	gitFlow     func(root string) config.GitFlowIntegrationConfig
	branchOf    func(wtPath string) (string, error)
	aheadCount  func(wtPath, develop, branch string) (int, error)
	wtForBranch func(branch string) string
	ownerPID    func() (int, bool)
	readLock    func(root string) (*kanban.IntegrationLock, error)
	acquire     func(root string, lock kanban.IntegrationLock) (*kanban.IntegrationLock, error)
	release     func(root, sessionID string) (*kanban.IntegrationLock, error)
	merge       func(dir, branch string) (string, error)
	inProgress  func(dir string) bool
	abort       func(dir string) error
	headShort   func(dir string) string
}

// swapAutoMergeSeams replaces the auto-merge seams and registers restoration.
// Nil fields leave the current seam in place.
func swapAutoMergeSeams(t *testing.T, s amSeams) {
	t.Helper()
	orig := amSeams{
		lockRoot: autoMergeLockRoot, sessionID: autoMergeSessionID,
		gitFlow: autoMergeLoadGitFlow, branchOf: autoMergeBranchOf,
		aheadCount: autoMergeAheadCount, wtForBranch: autoMergeWorktreeForBranch,
		ownerPID: autoMergeResolveOwnerPID, readLock: autoMergeReadLock,
		acquire: autoMergeAcquireLock, release: autoMergeReleaseLock,
		merge: autoMergeGitMerge, inProgress: autoMergeMergeInProgress,
		abort: autoMergeGitMergeAbort, headShort: autoMergeHeadShort,
	}
	if s.lockRoot != nil {
		autoMergeLockRoot = s.lockRoot
	}
	if s.sessionID != nil {
		autoMergeSessionID = s.sessionID
	}
	if s.gitFlow != nil {
		autoMergeLoadGitFlow = s.gitFlow
	}
	if s.branchOf != nil {
		autoMergeBranchOf = s.branchOf
	}
	if s.aheadCount != nil {
		autoMergeAheadCount = s.aheadCount
	}
	if s.wtForBranch != nil {
		autoMergeWorktreeForBranch = s.wtForBranch
	}
	if s.ownerPID != nil {
		autoMergeResolveOwnerPID = s.ownerPID
	}
	if s.readLock != nil {
		autoMergeReadLock = s.readLock
	}
	if s.acquire != nil {
		autoMergeAcquireLock = s.acquire
	}
	if s.release != nil {
		autoMergeReleaseLock = s.release
	}
	if s.merge != nil {
		autoMergeGitMerge = s.merge
	}
	if s.inProgress != nil {
		autoMergeMergeInProgress = s.inProgress
	}
	if s.abort != nil {
		autoMergeGitMergeAbort = s.abort
	}
	if s.headShort != nil {
		autoMergeHeadShort = s.headShort
	}
	t.Cleanup(func() {
		autoMergeLockRoot = orig.lockRoot
		autoMergeSessionID = orig.sessionID
		autoMergeLoadGitFlow = orig.gitFlow
		autoMergeBranchOf = orig.branchOf
		autoMergeAheadCount = orig.aheadCount
		autoMergeWorktreeForBranch = orig.wtForBranch
		autoMergeResolveOwnerPID = orig.ownerPID
		autoMergeReadLock = orig.readLock
		autoMergeAcquireLock = orig.acquire
		autoMergeReleaseLock = orig.release
		autoMergeGitMerge = orig.merge
		autoMergeMergeInProgress = orig.inProgress
		autoMergeGitMergeAbort = orig.abort
		autoMergeHeadShort = orig.headShort
	})
}

// amRecorder collects every observable side effect the engine produces, in
// order, so the tests can assert the whole ceremony from one struct.
type amRecorder struct {
	// gitLog is the seam log: every git subcommand the path invoked, in
	// order. The AC-WKW-006 zero-push assertion enumerates this log.
	gitLog []string
	// order records merge/remove side effects in a single timeline (the
	// merge-before-dispose ordering assertion).
	order []string
	// lockLog records "acquire" / "release" in order.
	lockLog []string
	// written is the last lock record acquire wrote; nil after release.
	written *kanban.IntegrationLock
	// held is what ReadIntegrationLock returns (nil → free window).
	held *kanban.IntegrationLock
	// mergeDirs / mergeBranches record each merge invocation's operands.
	mergeDirs     []string
	mergeBranches []string
	abortDirs     []string
	removed       []string
}

// amTargetWt is the fake integration worktree path the base seams resolve.
const amTargetWt = "/proj/.claude/worktrees/develop"

// amSessionWt is the fake session worktree path the tests trigger with.
const amSessionWt = "/proj/.claude/worktrees/WT-test-session"

// baseAMSeams returns fully-stubbed happy-path seams recording into rec: a
// git-flow project with develop_branch "develop", the session branch one
// commit ahead, a free window, and a successful merge. Tests override
// individual fields to reach the failure paths.
func baseAMSeams(rec *amRecorder) amSeams {
	return amSeams{
		lockRoot:  func() string { return "/proj" },
		sessionID: func() string { return "sess-am-1" },
		gitFlow: func(string) config.GitFlowIntegrationConfig {
			return config.GitFlowIntegrationConfig{Manual: true, GitFlowWorkflow: true, DevelopBranch: "develop"}
		},
		branchOf: func(wt string) (string, error) {
			rec.gitLog = append(rec.gitLog, "git -C "+wt+" rev-parse --abbrev-ref HEAD")
			return "WT-test-session", nil
		},
		aheadCount: func(wt, develop, branch string) (int, error) {
			rec.gitLog = append(rec.gitLog, "git -C "+wt+" rev-list --count "+develop+".."+branch)
			return 1, nil
		},
		wtForBranch: func(branch string) string {
			rec.gitLog = append(rec.gitLog, "git worktree list --porcelain")
			return amTargetWt
		},
		ownerPID: func() (int, bool) { return 4242, true },
		readLock: func(_ string) (*kanban.IntegrationLock, error) {
			if rec.held != nil {
				return rec.held, nil
			}
			return &kanban.IntegrationLock{}, nil
		},
		acquire: func(_ string, lock kanban.IntegrationLock) (*kanban.IntegrationLock, error) {
			rec.lockLog = append(rec.lockLog, "acquire")
			rec.written = &lock
			return nil, nil
		},
		release: func(_, _ string) (*kanban.IntegrationLock, error) {
			rec.lockLog = append(rec.lockLog, "release")
			rec.written = nil
			return nil, nil
		},
		merge: func(dir, branch string) (string, error) {
			rec.gitLog = append(rec.gitLog, "git -C "+dir+" merge --no-ff "+branch)
			rec.mergeDirs = append(rec.mergeDirs, dir)
			rec.mergeBranches = append(rec.mergeBranches, branch)
			rec.order = append(rec.order, "merge")
			return "Merge made by the 'ort' strategy.", nil
		},
		inProgress: func(string) bool { return false },
		abort: func(dir string) error {
			rec.abortDirs = append(rec.abortDirs, dir)
			rec.gitLog = append(rec.gitLog, "git -C "+dir+" merge --abort")
			return nil
		},
		headShort: func(string) string { return "abc1234" },
	}
}

// amCfg builds a Config with the two worktree toggles set.
func amCfg(autoMerge, autoCleanup bool) *config.Config {
	return &config.Config{
		Workflow: config.WorkflowConfig{
			Worktree: config.WorkflowWorktreeConfig{
				AutoMerge:   autoMerge,
				AutoCleanup: autoCleanup,
			},
		},
	}
}

// requireNoticePrefixed asserts every non-empty output line carries the
// auto-merge prefix (AC-WKW-013's per-line invariant).
func requireNoticePrefixed(t *testing.T, out string) {
	t.Helper()
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "moai: "+AutoMergeNoticePrefix) {
			t.Errorf("notice line lacks the %q prefix: %q", AutoMergeNoticePrefix, line)
		}
	}
}

// swapCleanStatusSeams makes the shared M4 status seam report every worktree
// clean — the base seams' fixture paths do not exist on disk, and the real
// `git status` probe against them would fail the dirty check and skip before
// the merge. Tests that must REACH the merge (or a specific dirty state) call
// this; tests that swap statusPorc themselves are unaffected.
func swapCleanStatusSeams(t *testing.T) {
	t.Helper()
	swapSessionWorktreeSeams(t, swSeams{
		statusPorc: func(string) (string, error) { return "", nil },
	})
}

// ---------------------------------------------------------------------------
// AC-WKW-001 — OFF baseline (characterization)
// ---------------------------------------------------------------------------

// TestAutoMergeOffBaseline is the REQ-WKW-001 byte-identical baseline: with
// auto_merge false (the distributed default) the session-exit path performs no
// merge invocation, no window record write, and emits no notice — observable
// output is exactly what the pre-SPEC flow produced (nothing merge-related).
func TestAutoMergeOffBaseline(t *testing.T) {
	rec := &amRecorder{}
	swapAutoMergeSeams(t, baseAMSeams(rec))
	out := &bytes.Buffer{}

	sessionExitAutoMerge(amCfg(false, false), amSessionWt, true, out)

	if out.Len() != 0 {
		t.Errorf("OFF baseline must be silent, got output:\n%s", out.String())
	}
	if len(rec.gitLog) != 0 {
		t.Errorf("OFF baseline must invoke no git command, got: %v", rec.gitLog)
	}
	if len(rec.lockLog) != 0 {
		t.Errorf("OFF baseline must write no window record, got: %v", rec.lockLog)
	}

	// Same for the no-session-worktree case: even with auto_merge ON, no
	// materialized worktree means nothing to merge and nothing observable.
	out2 := &bytes.Buffer{}
	sessionExitAutoMerge(amCfg(true, false), "", true, out2)
	if out2.Len() != 0 || len(rec.gitLog) != 0 || len(rec.lockLog) != 0 {
		t.Errorf("empty wtPath must be a full no-op: out=%q gitLog=%v lockLog=%v", out2.String(), rec.gitLog, rec.lockLog)
	}
}

// ---------------------------------------------------------------------------
// AC-WKW-002 — happy path (real git)
// ---------------------------------------------------------------------------

// autoMergeGitFixture builds a real git repository in a temp dir: branch
// develop (with a committed .gitignore so the window record never dirties the
// tree, and a git-strategy config naming develop), plus a session worktree on
// branch WT-am-test one commit ahead.
func autoMergeGitFixture(t *testing.T, conflicted bool) (root, swt string) {
	t.Helper()
	root = t.TempDir()
	git := func(args ...string) string {
		out, err := exec.Command("git", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
		return string(out)
	}
	git("-C", root, "init")
	git("-C", root, "config", "user.name", "t655-test")
	git("-C", root, "config", "user.email", "t655@test.local")
	git("-C", root, "symbolic-ref", "HEAD", "refs/heads/develop")
	if err := os.MkdirAll(filepath.Join(root, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	// .gitignore keeps the integration-lock record (.moai/state, written
	// during the ceremony before the merge) and the nested session worktree
	// (.claude/worktrees) from dirtying the target tree — same disposition a
	// real moai project's gitignore has.
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(".moai/state/\n.claude/worktrees/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	strategy := "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: develop\n"
	if err := os.WriteFile(filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml"), []byte(strategy), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "card.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("-C", root, "add", ".")
	git("-C", root, "commit", "-m", "init")

	swt = filepath.Join(root, ".claude", "worktrees", "WT-am-test")
	git("-C", root, "worktree", "add", "-b", "WT-am-test", swt)
	if err := os.WriteFile(filepath.Join(swt, "card.txt"), []byte("card work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("-C", swt, "add", ".")
	git("-C", swt, "commit", "-m", "card work")

	if conflicted {
		// develop changes the same file differently — the merge must conflict.
		if err := os.WriteFile(filepath.Join(root, "card.txt"), []byte("develop moved on\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		git("-C", root, "add", ".")
		git("-C", root, "commit", "-m", "develop diverges")
	}
	return root, swt
}

// TestAutoMergeHappyPath is AC-WKW-002: with auto_merge true, a clean exit,
// and the session branch one commit ahead, the path runs `git merge --no-ff`
// against the develop worktree, develop's HEAD becomes a two-parent merge
// commit whose second parent is the session branch tip, the window is
// released, and the notice names the merge commit.
func TestAutoMergeHappyPath(t *testing.T) {
	root, swt := autoMergeGitFixture(t, false)
	gitOut := func(args ...string) string {
		out, err := exec.Command("git", args...).Output()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	swapAutoMergeSeams(t, amSeams{
		lockRoot:  func() string { return root },
		sessionID: func() string { return "sess-fixture" },
		// Resolve against the fixture repo rather than the process cwd.
		wtForBranch: func(branch string) string {
			out, err := exec.Command("git", "-C", root, "worktree", "list", "--porcelain").Output()
			if err != nil {
				return ""
			}
			return worktreeForBranchFromList(string(out), branch)
		},
	})
	out := &bytes.Buffer{}

	sessionExitAutoMerge(amCfg(true, false), swt, true, out)

	// develop's HEAD is a merge commit with exactly two parents; the second
	// parent is the session branch tip.
	parents := strings.Fields(gitOut("-C", root, "rev-list", "--parents", "-n", "1", "HEAD"))
	if len(parents) != 3 {
		t.Fatalf("develop HEAD is not a two-parent merge commit: %v (notice: %q)", parents, out.String())
	}
	swtTip := gitOut("-C", swt, "rev-parse", "HEAD")
	if parents[2] != swtTip {
		t.Errorf("merge second parent %s != session branch tip %s", parents[2], swtTip)
	}

	// The window was acquired and released for this session (released at end).
	lock, err := kanban.ReadIntegrationLock(root)
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	if lock.Held() {
		t.Errorf("window must be released after the merge, still held by %s", lock.SessionID)
	}

	// The notice names the prefix, both branches, and the merge commit.
	notice := out.String()
	if !strings.HasPrefix(notice, "moai: "+AutoMergeNoticePrefix) {
		t.Errorf("success notice lacks the prefix: %q", notice)
	}
	if !strings.Contains(notice, "WT-am-test") || !strings.Contains(notice, "develop") {
		t.Errorf("success notice must name both branches: %q", notice)
	}
	if !strings.Contains(notice, parents[0][:7]) && !strings.Contains(notice, gitOut("-C", root, "rev-parse", "--short", "HEAD")) {
		t.Errorf("success notice must name the merge commit: %q (HEAD %s)", notice, parents[0])
	}
	requireNoticePrefixed(t, notice)
}

// ---------------------------------------------------------------------------
// AC-WKW-003 — non-clean exit
// ---------------------------------------------------------------------------

// TestAutoMergeNonCleanExit is AC-WKW-003: with auto_merge true and a
// non-zero session exit, no merge and no window mutation occur
// (clean-exit-only, mirroring REQ-SW-009).
func TestAutoMergeNonCleanExit(t *testing.T) {
	rec := &amRecorder{}
	swapAutoMergeSeams(t, baseAMSeams(rec))
	out := &bytes.Buffer{}

	sessionExitAutoMerge(amCfg(true, false), amSessionWt, false, out)

	if out.Len() != 0 {
		t.Errorf("non-clean exit must be silent on the merge path, got:\n%s", out.String())
	}
	if len(rec.gitLog) != 0 || len(rec.lockLog) != 0 {
		t.Errorf("non-clean exit must invoke nothing: gitLog=%v lockLog=%v", rec.gitLog, rec.lockLog)
	}
}

// ---------------------------------------------------------------------------
// AC-WKW-004 — inert when unconfigured
// ---------------------------------------------------------------------------

// TestAutoMergeUnconfiguredTarget is AC-WKW-004: with auto_merge true but the
// integration target unresolvable — develop_branch empty, github-flow mode,
// or an unreadable config — no merge runs and a non-blocking notice naming
// the config key is emitted.
func TestAutoMergeUnconfiguredTarget(t *testing.T) {
	cases := []struct {
		name string
		gfc  config.GitFlowIntegrationConfig
	}{
		{"develop_branch_empty", config.GitFlowIntegrationConfig{Manual: true, GitFlowWorkflow: true}},
		{"github_flow_mode", config.GitFlowIntegrationConfig{Manual: true, GitFlowWorkflow: false, DevelopBranch: "develop"}},
		{"config_unreadable", config.GitFlowIntegrationConfig{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := &amRecorder{}
			seams := baseAMSeams(rec)
			seams.gitFlow = func(string) config.GitFlowIntegrationConfig { return tc.gfc }
			swapAutoMergeSeams(t, seams)
			out := &bytes.Buffer{}

			sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, out)

			notice := out.String()
			if !strings.Contains(notice, "develop_branch") || !strings.Contains(notice, "git_strategy") {
				t.Errorf("notice must name the config key (git_strategy develop_branch), got: %q", notice)
			}
			requireNoticePrefixed(t, notice)
			// The target gate fires before any git invocation: no branch probe,
			// no ahead-of check, nothing.
			if len(rec.gitLog) != 0 {
				t.Errorf("no git invocation should precede the target gate, got: %v", rec.gitLog)
			}
			if len(rec.lockLog) != 0 {
				t.Errorf("no window record may be written on the unconfigured path, got: %v", rec.lockLog)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AC-WKW-005 — busy window (live AND stale hold)
// ---------------------------------------------------------------------------

// TestAutoMergeBusyWindow is AC-WKW-005: when the window is held — by a live
// session or by a stale record — the auto path writes no lock record,
// performs no merge, emits the skip notice, and leaves the existing hold
// untouched. (The engine skips on ANY foreign hold without consulting
// staleness: REQ-WKW-004 forbids displacing a stale hold as much as a live
// one.)
func TestAutoMergeBusyWindow(t *testing.T) {
	cases := []struct {
		name   string
		held   kanban.IntegrationLock
		holder string // expected to appear in the skip notice
	}{
		{
			name:   "live_holder",
			held:   kanban.IntegrationLock{SessionID: "other-live", SessionName: "lane-2", PID: os.Getpid(), Branch: "develop", AcquiredAt: "2026-09-12T00:00:00Z"},
			holder: "lane-2",
		},
		{
			name:   "stale_record",
			held:   kanban.IntegrationLock{SessionID: "dead-lane", PID: 999999999, Branch: "develop", AcquiredAt: "2026-09-12T00:00:00Z"},
			holder: "dead-lane",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := &amRecorder{held: &tc.held}
			swapAutoMergeSeams(t, baseAMSeams(rec))
			out := &bytes.Buffer{}

			sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, out)

			notice := out.String()
			if !strings.Contains(notice, "held by") || !strings.Contains(notice, tc.holder) {
				t.Errorf("skip notice must name the holder %q, got: %q", tc.holder, notice)
			}
			requireNoticePrefixed(t, notice)
			if len(rec.lockLog) != 0 {
				t.Errorf("the auto path must write no lock record on a held window, got: %v", rec.lockLog)
			}
			if len(rec.mergeDirs) != 0 {
				t.Errorf("no merge may run on a held window, got merges at %v", rec.mergeDirs)
			}
			if rec.written != nil {
				t.Errorf("the existing hold must be untouched, but a record was written: %+v", rec.written)
			}
		})
	}

	// Own-session hold: the lock API re-acquires (refresh) rather than
	// refusing, so the ceremony proceeds and the release frees our own hold.
	t.Run("own_session_hold_proceeds", func(t *testing.T) {
		rec := &amRecorder{held: &kanban.IntegrationLock{SessionID: "sess-am-1", PID: os.Getpid(), Branch: "develop"}}
		swapAutoMergeSeams(t, baseAMSeams(rec))
		swapCleanStatusSeams(t)
		out := &bytes.Buffer{}

		sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, out)

		if len(rec.mergeDirs) != 1 {
			t.Errorf("own-session hold proceeds to merge, got %d merges", len(rec.mergeDirs))
		}
		if len(rec.lockLog) != 2 || rec.lockLog[0] != "acquire" || rec.lockLog[1] != "release" {
			t.Errorf("ceremony must acquire then release, got: %v", rec.lockLog)
		}
		requireNoticePrefixed(t, out.String())
	})
}

// ---------------------------------------------------------------------------
// AC-WKW-006 — ceremony + zero-push (seam log enumeration)
// ---------------------------------------------------------------------------

// TestAutoMergeZeroPush is AC-WKW-006: on a successful auto-merge the lock
// transitioned acquire→release for this session, and the seam log of every
// git invocation on the path contains no push, fetch, pull, or
// remote-mutating subcommand — exactly one merge, and it carries --no-ff.
func TestAutoMergeZeroPush(t *testing.T) {
	rec := &amRecorder{}
	swapAutoMergeSeams(t, baseAMSeams(rec))
	swapCleanStatusSeams(t)
	out := &bytes.Buffer{}

	sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, out)

	// Exactly one merge, with --no-ff, against the develop worktree.
	merges := 0
	for _, entry := range rec.gitLog {
		if strings.Contains(entry, " merge ") || strings.Contains(entry, "merge --no-ff") {
			merges++
			if !strings.Contains(entry, "--no-ff") {
				t.Errorf("the merge invocation must carry --no-ff: %q", entry)
			}
		}
	}
	if merges != 1 {
		t.Errorf("exactly one merge expected, seam log shows %d: %v", merges, rec.gitLog)
	}
	if len(rec.mergeDirs) != 1 || rec.mergeDirs[0] != amTargetWt {
		t.Errorf("merge must run in the develop worktree %s, got %v", amTargetWt, rec.mergeDirs)
	}
	if len(rec.mergeBranches) != 1 || rec.mergeBranches[0] != "WT-test-session" {
		t.Errorf("merge must target the session branch, got %v", rec.mergeBranches)
	}

	// No remote-mutating subcommand anywhere in the log.
	for _, entry := range rec.gitLog {
		for _, banned := range []string{"push", "fetch", "pull", "remote", "origin"} {
			if strings.Contains(entry, banned) {
				t.Errorf("remote-mutating subcommand %q found in seam log: %q", banned, entry)
			}
		}
	}

	// Ceremony: acquire then release, released at end.
	if len(rec.lockLog) != 2 || rec.lockLog[0] != "acquire" || rec.lockLog[1] != "release" {
		t.Errorf("lock ceremony must be acquire→release, got: %v", rec.lockLog)
	}
	if rec.written != nil {
		t.Errorf("window must be released at end, still written: %+v", rec.written)
	}
	requireNoticePrefixed(t, out.String())
}

// ---------------------------------------------------------------------------
// AC-WKW-007 — conflict safety
// ---------------------------------------------------------------------------

// TestAutoMergeConflict is AC-WKW-007, in two parts: the seam-injected
// conflict path (abort called in the target worktree, window released,
// notice), and a real-git conflicted merge proving MERGE_HEAD does not remain
// and the session worktree is untouched.
func TestAutoMergeConflict(t *testing.T) {
	t.Run("seam_abort_path", func(t *testing.T) {
		rec := &amRecorder{}
		seams := baseAMSeams(rec)
		seams.merge = func(dir, branch string) (string, error) {
			rec.mergeDirs = append(rec.mergeDirs, dir)
			rec.order = append(rec.order, "merge")
			return "CONFLICT (content): Merge conflict in card.txt", &execExitErr{}
		}
		seams.inProgress = func(string) bool { return true }
		swapAutoMergeSeams(t, seams)
		swapCleanStatusSeams(t)
		out := &bytes.Buffer{}

		sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, out)

		if len(rec.abortDirs) != 1 || rec.abortDirs[0] != amTargetWt {
			t.Errorf("git merge --abort must run once in the target worktree, got %v", rec.abortDirs)
		}
		if len(rec.lockLog) != 2 || rec.lockLog[1] != "release" {
			t.Errorf("window must be released after the conflict, got: %v", rec.lockLog)
		}
		notice := out.String()
		if !strings.Contains(notice, "conflict") {
			t.Errorf("conflict notice must say so: %q", notice)
		}
		requireNoticePrefixed(t, notice)
	})

	t.Run("real_git_conflict_leaves_no_merge_head", func(t *testing.T) {
		root, swt := autoMergeGitFixture(t, true)
		swapAutoMergeSeams(t, amSeams{
			lockRoot:  func() string { return root },
			sessionID: func() string { return "sess-fixture" },
			wtForBranch: func(branch string) string {
				out, err := exec.Command("git", "-C", root, "worktree", "list", "--porcelain").Output()
				if err != nil {
					return ""
				}
				return worktreeForBranchFromList(string(out), branch)
			},
		})
		out := &bytes.Buffer{}

		sessionExitAutoMerge(amCfg(true, false), swt, true, out)

		// No MERGE_HEAD remains in the integration worktree.
		if err := exec.Command("git", "-C", root, "rev-parse", "-q", "--verify", "MERGE_HEAD").Run(); err == nil {
			t.Errorf("MERGE_HEAD still resolves in %s — the abort did not restore the tree", root)
		}
		// The session worktree is byte-identical to pre-trigger (its file
		// still carries the session's committed content).
		kept, rerr := os.ReadFile(filepath.Join(swt, "card.txt"))
		if rerr != nil || string(kept) != "card work\n" {
			t.Errorf("session worktree content changed: %q (err %v)", kept, rerr)
		}
		// Window released.
		lock, lerr := kanban.ReadIntegrationLock(root)
		if lerr != nil || lock.Held() {
			t.Errorf("window must be released after the conflict (err %v, lock %+v)", lerr, lock)
		}
		requireNoticePrefixed(t, out.String())
	})
}

// execExitErr is a minimal error standing in for a failed merge exit.
type execExitErr struct{}

func (*execExitErr) Error() string { return "exit status 1" }

// ---------------------------------------------------------------------------
// AC-WKW-008 — source dirty guard
// ---------------------------------------------------------------------------

// TestAutoMergeSourceDirty is AC-WKW-008: uncommitted changes in the session
// worktree skip the merge with the notice; the worktree is untouched; and
// with auto_cleanup also on, disposal follows its own dirty guard.
func TestAutoMergeSourceDirty(t *testing.T) {
	rec := &amRecorder{}
	swapAutoMergeSeams(t, baseAMSeams(rec))
	// The session worktree reports uncommitted changes; the target is clean.
	// The remove seam records into the same recorder so disposal is observable.
	swapSessionWorktreeSeams(t, swSeams{
		statusPorc: func(p string) (string, error) {
			if p == amSessionWt {
				return " M unfinished.txt\n", nil
			}
			return "", nil
		},
		remove: func(p string) error {
			rec.order = append(rec.order, "remove")
			rec.removed = append(rec.removed, p)
			return nil
		},
	})
	out := &bytes.Buffer{}
	cleanupOut := &bytes.Buffer{}

	sessionExitAutoMerge(amCfg(true, true), amSessionWt, true, out)
	// With auto_cleanup also on, disposal follows its own dirty guard: the
	// merge skipped, then cleanup runs and preserves the dirty tree.
	cleanupSessionWorktree(amCfg(true, true), amSessionWt, true, cleanupOut)

	notice := out.String()
	if !strings.Contains(notice, "uncommitted") {
		t.Errorf("source-dirty notice must say so: %q", notice)
	}
	// The prefix invariant binds the auto-merge family's own notices; the
	// cleanup path's notice carries ITS family's prefix (distinctness, not
	// uniformity — EC-13).
	requireNoticePrefixed(t, notice)
	if len(rec.mergeDirs) != 0 {
		t.Errorf("no merge may run on a dirty source, got merges at %v", rec.mergeDirs)
	}
	// Ceremony per plan.md M1 ordering: acquire → guards → (skip) → release.
	if len(rec.lockLog) != 2 || rec.lockLog[0] != "acquire" || rec.lockLog[1] != "release" {
		t.Errorf("guards run inside the window; expected acquire+release, got: %v", rec.lockLog)
	}
	if len(rec.removed) != 0 {
		t.Errorf("auto_cleanup's own dirty guard must preserve the dirty worktree, got removals: %v", rec.removed)
	}
	// Both families' notices appear in the same session-exit output, each
	// attributable to its own path (EC-13). The cleanup family's dirty-skip
	// notice — not its removal notice — is the one this scenario produces.
	if !strings.Contains(cleanupOut.String(), "session-exit cleanup skipped") {
		t.Errorf("cleanup's own skip notice expected alongside the merge notice:\nmerge: %q\ncleanup: %q", notice, cleanupOut.String())
	}
}

// ---------------------------------------------------------------------------
// AC-WKW-009 — target guards
// ---------------------------------------------------------------------------

// TestAutoMergeTargetGuards is AC-WKW-009: with develop configured, a missing
// develop worktree and a dirty develop worktree each skip the merge with
// their own notice.
func TestAutoMergeTargetGuards(t *testing.T) {
	t.Run("no_worktree_holds_develop", func(t *testing.T) {
		rec := &amRecorder{}
		seams := baseAMSeams(rec)
		seams.wtForBranch = func(string) string { return "" }
		swapAutoMergeSeams(t, seams)
		out := &bytes.Buffer{}

		sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, out)

		if !strings.Contains(out.String(), "no worktree holds") {
			t.Errorf("target-absent notice expected, got: %q", out.String())
		}
		requireNoticePrefixed(t, out.String())
		if len(rec.mergeDirs) != 0 {
			t.Errorf("no merge may run without a target worktree, got %v", rec.mergeDirs)
		}
		if len(rec.lockLog) != 2 {
			t.Errorf("guards run inside the window, got: %v", rec.lockLog)
		}
	})

	t.Run("target_worktree_dirty", func(t *testing.T) {
		rec := &amRecorder{}
		swapAutoMergeSeams(t, baseAMSeams(rec))
		swapSessionWorktreeSeams(t, swSeams{
			statusPorc: func(p string) (string, error) {
				if p == amTargetWt {
					return " M staged.txt\n", nil
				}
				return "", nil
			},
		})
		out := &bytes.Buffer{}

		sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, out)

		if !strings.Contains(out.String(), "uncommitted changes") {
			t.Errorf("target-dirty notice expected, got: %q", out.String())
		}
		requireNoticePrefixed(t, out.String())
		if len(rec.mergeDirs) != 0 {
			t.Errorf("no merge may run into a dirty target, got %v", rec.mergeDirs)
		}
	})
}

// ---------------------------------------------------------------------------
// AC-WKW-010 — no-op silent skip
// ---------------------------------------------------------------------------

// TestAutoMergeNoOpSilent is AC-WKW-010: when the session branch is fully
// contained in develop, the entire ceremony is skipped — no window record, no
// merge invocation, no notice (byte-identical to the OFF baseline).
func TestAutoMergeNoOpSilent(t *testing.T) {
	rec := &amRecorder{}
	seams := baseAMSeams(rec)
	seams.aheadCount = func(_, _, _ string) (int, error) { return 0, nil }
	swapAutoMergeSeams(t, seams)
	out := &bytes.Buffer{}

	sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, out)

	if out.Len() != 0 {
		t.Errorf("no-op skip must be silent, got:\n%s", out.String())
	}
	if len(rec.lockLog) != 0 {
		t.Errorf("no window churn on a no-op branch, got: %v", rec.lockLog)
	}
	for _, entry := range rec.gitLog {
		if strings.Contains(entry, "merge") {
			t.Errorf("no merge invocation on a no-op branch, got: %q", entry)
		}
	}
}

// ---------------------------------------------------------------------------
// AC-WKW-013 — prefix distinctness + non-blocking
// ---------------------------------------------------------------------------

// TestAutoMergeNoticePrefixDistinct is AC-WKW-013: the auto-merge prefix
// differs from both removal prefixes at the constant level, every notice line
// every failure path emits carries the prefix, and a forced merge failure
// leaves the caller unaffected (no error return to propagate).
func TestAutoMergeNoticePrefixDistinct(t *testing.T) {
	if AutoMergeNoticePrefix == SessionExitCleanupNoticePrefix {
		t.Errorf("AutoMergeNoticePrefix collides with SessionExitCleanupNoticePrefix: %q", AutoMergeNoticePrefix)
	}
	if AutoMergeNoticePrefix == PRMergeCleanupNoticePrefix {
		t.Errorf("AutoMergeNoticePrefix collides with PRMergeCleanupNoticePrefix: %q", AutoMergeNoticePrefix)
	}
	if SessionExitCleanupNoticePrefix == PRMergeCleanupNoticePrefix {
		t.Errorf("pre-existing invariant broken: session-exit and PR-merge prefixes collide")
	}

	// Every failure path's notice lines carry the prefix. A merge failure is
	// "forced" by the erroring merge seam; the call completes without
	// panicking and without an error to propagate, so the session-exit
	// flow's exit status is unaffected by construction (no error return).
	failure := &bytes.Buffer{}
	rec := &amRecorder{}
	seams := baseAMSeams(rec)
	seams.merge = func(string, string) (string, error) { return "boom", &execExitErr{} }
	swapAutoMergeSeams(t, seams)
	swapCleanStatusSeams(t)
	sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, failure)
	requireNoticePrefixed(t, failure.String())

	// Unconfigured-target and held-window notices too.
	unconfigured := &bytes.Buffer{}
	rec2 := &amRecorder{}
	seams2 := baseAMSeams(rec2)
	seams2.gitFlow = func(string) config.GitFlowIntegrationConfig { return config.GitFlowIntegrationConfig{} }
	swapAutoMergeSeams(t, seams2)
	sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, unconfigured)
	requireNoticePrefixed(t, unconfigured.String())

	busy := &bytes.Buffer{}
	rec3 := &amRecorder{held: &kanban.IntegrationLock{SessionID: "other", PID: 1}}
	swapAutoMergeSeams(t, baseAMSeams(rec3))
	sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, busy)
	requireNoticePrefixed(t, busy.String())
}

// ---------------------------------------------------------------------------
// AC-WKW-014 — toggle independence (4-combination matrix)
// ---------------------------------------------------------------------------

// TestAutoMergeToggleIndependence is AC-WKW-014: across all four
// auto_merge × auto_cleanup combinations on a clean exit with merge
// preconditions satisfied, merge occurs iff auto_merge is on, removal occurs
// iff auto_cleanup is on, and — in the both-on case — the merge is observed
// BEFORE the removal (merge-then-dispose).
func TestAutoMergeToggleIndependence(t *testing.T) {
	for _, tc := range []struct {
		merge, cleanup bool
	}{
		{false, false}, {true, false}, {false, true}, {true, true},
	} {
		name := "merge=" + boolWord(tc.merge) + "/cleanup=" + boolWord(tc.cleanup)
		t.Run(name, func(t *testing.T) {
			rec := &amRecorder{}
			swapAutoMergeSeams(t, baseAMSeams(rec))
			swapSessionWorktreeSeams(t, swSeams{
				statusPorc: func(string) (string, error) { return "", nil },
				remove: func(p string) error {
					rec.order = append(rec.order, "remove")
					rec.removed = append(rec.removed, p)
					return nil
				},
			})
			out := &bytes.Buffer{}

			sessionExitAutoMerge(amCfg(tc.merge, tc.cleanup), amSessionWt, true, out)
			cleanupSessionWorktree(amCfg(tc.merge, tc.cleanup), amSessionWt, true, out)

			mergeCount, removeCount := 0, 0
			for _, step := range rec.order {
				switch step {
				case "merge":
					mergeCount++
				case "remove":
					removeCount++
				}
			}
			if tc.merge {
				if mergeCount != 1 {
					t.Errorf("merge must fire exactly once when auto_merge is on, got %d (order %v)", mergeCount, rec.order)
				}
				if !strings.Contains(out.String(), AutoMergeNoticePrefix) {
					t.Errorf("merge notice expected in combined output:\n%s", out.String())
				}
			} else if mergeCount != 0 {
				t.Errorf("merge must not fire when auto_merge is off, got %d (order %v)", mergeCount, rec.order)
			}
			if tc.cleanup {
				if removeCount != 1 {
					t.Errorf("removal must fire exactly once when auto_cleanup is on, got %d (order %v)", removeCount, rec.order)
				}
				if !strings.Contains(out.String(), SessionExitCleanupNoticePrefix) {
					t.Errorf("removal notice expected in combined output:\n%s", out.String())
				}
			} else if removeCount != 0 {
				t.Errorf("removal must not fire when auto_cleanup is off, got %d (order %v)", removeCount, rec.order)
			}
			if tc.merge && tc.cleanup {
				mergeIdx, removeIdx := -1, -1
				for i, step := range rec.order {
					if step == "merge" && mergeIdx < 0 {
						mergeIdx = i
					}
					if step == "remove" && removeIdx < 0 {
						removeIdx = i
					}
				}
				if mergeIdx > removeIdx {
					t.Errorf("merge must precede removal (order %v)", rec.order)
				}
			}
		})
	}
}

func boolWord(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

// ---------------------------------------------------------------------------
// Failure-path notices (coverage + REQ-WKW-009 non-blocking family)
// ---------------------------------------------------------------------------

// TestAutoMergeFailurePaths drives every remaining guard/failure branch of
// sessionExitAutoMerge through a table: each case swaps exactly one seam into
// a failure state and asserts the path emits its notice, never merges, and
// never aborts the caller. Together with the AC tests above this covers the
// engine's decision tree end to end.
func TestAutoMergeFailurePaths(t *testing.T) {
	cases := []struct {
		name       string
		mutate     func(seams *amSeams) // seam overrides for the failure state
		status     func(p string) (string, error)
		notice     string // a substring the notice must carry
		noMerge    bool   // assert no merge ran (true for all but release_fails)
		noLock     bool   // assert no acquire happened
		wantMerges int    // expected merge count
	}{
		{
			name: "branch_probe_fails",
			mutate: func(s *amSeams) {
				s.branchOf = func(string) (string, error) { return "", &execExitErr{} }
			},
			notice: "cannot resolve the session branch", noMerge: true, noLock: true,
		},
		{
			name: "ahead_check_fails",
			mutate: func(s *amSeams) {
				s.aheadCount = func(string, string, string) (int, error) { return 0, &execExitErr{} }
			},
			notice: "ahead-of check failed", noMerge: true, noLock: true,
		},
		{
			name: "session_id_unresolvable",
			mutate: func(s *amSeams) {
				s.sessionID = func() string { return "" }
			},
			notice: "cannot resolve this session's id", noMerge: true, noLock: true,
		},
		{
			name: "window_record_unreadable",
			mutate: func(s *amSeams) {
				s.readLock = func(string) (*kanban.IntegrationLock, error) { return nil, &execExitErr{} }
			},
			notice: "window record unreadable", noMerge: true, noLock: true,
		},
		{
			name: "acquire_refused",
			mutate: func(s *amSeams) {
				s.acquire = func(string, kanban.IntegrationLock) (*kanban.IntegrationLock, error) {
					return nil, &execExitErr{}
				}
			},
			notice: "window acquire refused", noMerge: true,
		},
		{
			name: "release_fails",
			mutate: func(s *amSeams) {
				s.release = func(string, string) (*kanban.IntegrationLock, error) { return nil, &execExitErr{} }
			},
			notice: "window release failed", wantMerges: 1,
		},
		{
			name: "merge_error_without_conflict",
			mutate: func(s *amSeams) {
				s.merge = func(string, string) (string, error) { return "", &execExitErr{} }
				s.inProgress = func(string) bool { return false }
			},
			notice: "merge failed", wantMerges: 0, // the erroring override replaces the recording base seam; the attempt is evidenced by the failure notice
		},
		{
			name: "abort_also_fails",
			mutate: func(s *amSeams) {
				s.merge = func(string, string) (string, error) { return "", &execExitErr{} }
				s.inProgress = func(string) bool { return true }
				s.abort = func(string) error { return &execExitErr{} }
			},
			notice: "'git merge --abort' failed", wantMerges: 0, // same: the override seam is not the recording one
		},
		{
			name:   "source_dirty_check_fails",
			status: func(p string) (string, error) { return "", &execExitErr{} },
			notice: "source dirty-check failed", noMerge: true,
		},
		{
			name: "target_dirty_check_fails",
			status: func(p string) (string, error) {
				if p == amTargetWt {
					return "", &execExitErr{}
				}
				return "", nil
			},
			notice: "target dirty-check failed", noMerge: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := &amRecorder{}
			seams := baseAMSeams(rec)
			if tc.mutate != nil {
				tc.mutate(&seams)
			}
			swapAutoMergeSeams(t, seams)
			if tc.status != nil {
				sp := tc.status
				swapSessionWorktreeSeams(t, swSeams{statusPorc: sp})
			} else {
				swapCleanStatusSeams(t)
			}
			out := &bytes.Buffer{}

			sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, out)

			notice := out.String()
			if !strings.Contains(notice, tc.notice) {
				t.Errorf("notice must name the failure (%q), got: %q", tc.notice, notice)
			}
			requireNoticePrefixed(t, notice)
			if tc.noMerge && len(rec.mergeDirs) != 0 {
				t.Errorf("no merge may run on this path, got %v", rec.mergeDirs)
			}
			if tc.noLock && len(rec.lockLog) != 0 {
				t.Errorf("no window record may be written on this path, got: %v", rec.lockLog)
			}
			if len(rec.mergeDirs) != tc.wantMerges {
				t.Errorf("merge count = %d, want %d", len(rec.mergeDirs), tc.wantMerges)
			}
		})
	}
}

// TestAutoMergeRealImplErrorPaths covers the real git helpers' error returns
// directly (the seam-based tests above exercise the decision tree, not these
// wrappers' failure arms).
func TestAutoMergeRealImplErrorPaths(t *testing.T) {
	if _, err := gitBranchOfWorktreeReal(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Errorf("branch probe of a nonexistent tree must error")
	}
	if _, err := gitRevListCountReal(filepath.Join(t.TempDir(), "nope"), "develop", "x"); err == nil {
		t.Errorf("rev-list on a nonexistent tree must error")
	}
	if err := gitMergeAbortReal(t.TempDir()); err == nil {
		t.Errorf("merge --abort outside a merge must error")
	}
	if got := gitHeadShortReal(filepath.Join(t.TempDir(), "nope")); got != "unknown" {
		t.Errorf("head probe of a nonexistent tree must degrade to \"unknown\", got %q", got)
	}
}
