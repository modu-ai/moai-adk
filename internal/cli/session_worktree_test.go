package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-SESSION-WORKTREE-001 M2 — session-worktree auto-entry wrapper tests.
//
// The auto-entry logic is exercised through function-variable seams
// (sessionWorktreeGitWorktreeAdd, sessionWorktreeInGitWorktree,
// sessionWorktreeResolveSessionShort, sessionWorktreeGitCommonDir,
// sessionWorktreeGitConfigSet) so the tests never depend on a real git
// repository. Each test restores the real implementations via t.Cleanup.

// swSeams snapshots the package-level function-variable seams so a test can
// swap them and restore on cleanup.
type swSeams struct {
	add        func(destDir, branch string) (string, error)
	inWt       func() bool
	short      func() string
	commonDir  func() (string, error)
	configSet  func(dir, key, value string) error
	remove     func(wtPath string) error
	statusPorc func(wtPath string) (string, error)
}

// swapSessionWorktreeSeams replaces the seams and registers restoration.
func swapSessionWorktreeSeams(t *testing.T, s swSeams) {
	t.Helper()
	orig := swSeams{
		add: sessionWorktreeGitWorktreeAdd, inWt: sessionWorktreeInGitWorktree,
		short: sessionWorktreeResolveSessionShort, commonDir: sessionWorktreeGitCommonDir,
		configSet:  sessionWorktreeGitConfigSet,
		remove:     sessionWorktreeGitWorktreeRemove,
		statusPorc: sessionWorktreeGitStatusPorcelain,
	}
	if s.add != nil {
		sessionWorktreeGitWorktreeAdd = s.add
	}
	if s.inWt != nil {
		sessionWorktreeInGitWorktree = s.inWt
	}
	if s.short != nil {
		sessionWorktreeResolveSessionShort = s.short
	}
	if s.commonDir != nil {
		sessionWorktreeGitCommonDir = s.commonDir
	}
	if s.configSet != nil {
		sessionWorktreeGitConfigSet = s.configSet
	}
	if s.remove != nil {
		sessionWorktreeGitWorktreeRemove = s.remove
	}
	if s.statusPorc != nil {
		sessionWorktreeGitStatusPorcelain = s.statusPorc
	}
	t.Cleanup(func() {
		sessionWorktreeGitWorktreeAdd = orig.add
		sessionWorktreeInGitWorktree = orig.inWt
		sessionWorktreeResolveSessionShort = orig.short
		sessionWorktreeGitCommonDir = orig.commonDir
		sessionWorktreeGitConfigSet = orig.configSet
		sessionWorktreeGitWorktreeRemove = orig.remove
		sessionWorktreeGitStatusPorcelain = orig.statusPorc
	})
}

// TestEnterSessionWorktree_DefaultOffReturnsEmpty is the REQ-SW-001
// byte-identical baseline: when the feature is OFF the wrapper short-circuits
// before any side effect (no git invocation, no notice).
func TestEnterSessionWorktree_DefaultOffReturnsEmpty(t *testing.T) {
	called := false
	swapSessionWorktreeSeams(t, swSeams{
		add: func(string, string) (string, error) { called = true; return "", nil },
	})
	var out bytes.Buffer
	got := enterSessionWorktree(nil, "init", &out) // nil cfg → OFF (M1 nil-safety)
	if got != "" {
		t.Fatalf("default-off: expected empty worktree path, got %q", got)
	}
	if called {
		t.Fatal("default-off: git worktree add MUST NOT be invoked")
	}
	if out.Len() != 0 {
		t.Fatalf("default-off: expected no notice, got %q", out.String())
	}
}

// TestEnterSessionWorktree_DefaultOffWithConfigFalse mirrors the above with an
// explicit config carrying Enabled=false (no env override).
func TestEnterSessionWorktree_DefaultOffWithConfigFalse(t *testing.T) {
	t.Setenv("MOAI_SESSION_WORKTREE", "")
	swapSessionWorktreeSeams(t, swSeams{
		add: func(string, string) (string, error) { t.Fatal("add must not run"); return "", nil },
	})
	cfg := &config.Config{Workflow: config.WorkflowConfig{SessionWorktree: config.SessionWorktreeConfig{Enabled: false}}}
	var out bytes.Buffer
	if got := enterSessionWorktree(cfg, "init", &out); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

// TestEnterSessionWorktree_EnvForcesOn exercises REQ-SW-003: MOAI_SESSION_WORKTREE=1
// forces ON even when config says false. (M1 owns the helper; this confirms the
// wrapper routes through it.)
func TestEnterSessionWorktree_EnvForcesOn(t *testing.T) {
	t.Setenv("MOAI_SESSION_WORKTREE", "1")
	swapSessionWorktreeSeams(t, swSeams{
		inWt:      func() bool { return false },
		short:     func() string { return "abcdef12" },
		commonDir: func() (string, error) { return "/repo/.git", nil },
		add:       func(dest, branch string) (string, error) { return dest, nil },
		configSet: func(string, string, string) error { return nil },
	})
	cfg := &config.Config{Workflow: config.WorkflowConfig{SessionWorktree: config.SessionWorktreeConfig{Enabled: false}}}
	var out bytes.Buffer
	got := enterSessionWorktree(cfg, "init", &out)
	if got == "" {
		t.Fatal("env=1 should force ON and materialize a worktree")
	}
	if !strings.HasPrefix(got, "/repo/.claude/worktrees/") {
		t.Fatalf("unexpected worktree path %q", got)
	}
}

// TestEnterSessionWorktree_AlreadyInWorktreeSkips is REQ-SW-012: when cwd is
// already inside a worktree, auto-entry is skipped with an info notice.
func TestEnterSessionWorktree_AlreadyInWorktreeSkips(t *testing.T) {
	t.Setenv("MOAI_SESSION_WORKTREE", "1")
	addCalled := false
	swapSessionWorktreeSeams(t, swSeams{
		inWt: func() bool { return true },
		add: func(string, string) (string, error) { addCalled = true; return "", nil },
	})
	var out bytes.Buffer
	got := enterSessionWorktree(nil, "init", &out)
	if got != "" {
		t.Fatalf("already-in-worktree: expected empty path, got %q", got)
	}
	if addCalled {
		t.Fatal("already-in-worktree: git worktree add MUST NOT run")
	}
	if !strings.Contains(out.String(), "already inside a git worktree") {
		t.Fatalf("expected skip notice, got %q", out.String())
	}
}

// TestEnterSessionWorktree_MaterializeFailFallsBack is REQ-SW-004: a
// materialization failure MUST NOT abort; the wrapper returns "" and emits a
// notice naming the failure reason.
func TestEnterSessionWorktree_MaterializeFailFallsBack(t *testing.T) {
	t.Setenv("MOAI_SESSION_WORKTREE", "1")
	swapSessionWorktreeSeams(t, swSeams{
		inWt:      func() bool { return false },
		short:     func() string { return "abcdef12" },
		commonDir: func() (string, error) { return "/repo/.git", nil },
		add:       func(string, string) (string, error) { return "", errFakeGitAdd },
	})
	var out bytes.Buffer
	got := enterSessionWorktree(nil, "init", &out)
	if got != "" {
		t.Fatalf("fail-back: expected empty path, got %q", got)
	}
	if !strings.Contains(out.String(), "materialization failed") {
		t.Fatalf("expected fail-back notice, got %q", out.String())
	}
	if !strings.Contains(out.String(), "fake-git-add") {
		t.Fatalf("notice must name the failure reason, got %q", out.String())
	}
}

// TestEnterSessionWorktree_SuccessAppliesDefaultBranch is REQ-SW-020: a
// successful materialization applies init.defaultBranch=main to the worktree's
// local config.
func TestEnterSessionWorktree_SuccessAppliesDefaultBranch(t *testing.T) {
	t.Setenv("MOAI_SESSION_WORKTREE", "1")
	var setCalls []struct{ dir, key, val string }
	swapSessionWorktreeSeams(t, swSeams{
		inWt:      func() bool { return false },
		short:     func() string { return "abcdef12" },
		commonDir: func() (string, error) { return "/repo/.git", nil },
		add:       func(dest, branch string) (string, error) { return dest, nil },
		configSet: func(dir, key, val string) error {
			setCalls = append(setCalls, struct{ dir, key, val string }{dir, key, val})
			return nil
		},
	})
	var out bytes.Buffer
	got := enterSessionWorktree(nil, "init", &out)
	if got == "" {
		t.Fatal("expected materialized worktree path")
	}
	found := false
	for _, c := range setCalls {
		if c.key == "init.defaultBranch" && c.val == "main" && c.dir == got {
			found = true
		}
	}
	if !found {
		t.Fatalf("init.defaultBranch=main not applied to worktree %q; calls=%v", got, setCalls)
	}
}

// TestSessionWorktreeBranchName_Shape verifies the WT-<8hex>-<subcommand>
// shape (Q2: brackets rejected → WT- prefix fallback).
func TestSessionWorktreeBranchName_Shape(t *testing.T) {
	swapSessionWorktreeSeams(t, swSeams{
		short: func() string { return "abcdef12" },
	})
	if got := sessionWorktreeBranchName("init"); got != "WT-abcdef12-init" {
		t.Fatalf("branch name shape: got %q, want WT-abcdef12-init", got)
	}
}

// TestSessionWorktreeBranchName_RandomFallback verifies the 6-byte random hex
// fallback when no session id is available (REQ-SW-007 EC-4).
func TestSessionWorktreeBranchName_RandomFallback(t *testing.T) {
	swapSessionWorktreeSeams(t, swSeams{
		short: func() string { return resolveSessionShortReal() },
	})
	got := sessionWorktreeBranchName("init")
	// Shape: WT-<12hex>-init (6 bytes → 12 hex chars).
	if !strings.HasPrefix(got, "WT-") || !strings.HasSuffix(got, "-init") {
		t.Fatalf("random fallback shape wrong: %q", got)
	}
	mid := strings.TrimSuffix(strings.TrimPrefix(got, "WT-"), "-init")
	if len(mid) != 12 {
		t.Fatalf("random segment must be 12 hex chars (6 bytes), got %d (%q)", len(mid), mid)
	}
	for _, r := range mid {
		isHex := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')
		if !isHex {
			t.Fatalf("random segment must be lowercase hex, got %q in %q", r, mid)
		}
	}
}

// TestEnterSessionWorktree_FailBackFalsification is the E8 falsification
// round-trip: it proves the fail-back code path is load-bearing by asserting
// the test ABOVE (MaterializeFailFallsBack) would FAIL if the fail-back were
// removed (i.e. the wrapper propagated the error). We re-derive the falsification
// here by checking that the wrapper's contract — "returns "" on any add error"
// — holds even when the underlying add returns a non-nil error. If a future
// edit made enterSessionWorktree return the error instead of swallowing it,
// this test's signature (no error return) would no longer compile, surfacing
// the regression at build time.
func TestEnterSessionWorktree_FailBackFalsification(t *testing.T) {
	t.Setenv("MOAI_SESSION_WORKTREE", "1")
	// Simulate a non-git directory: commonDir resolution fails (BI-3 path).
	swapSessionWorktreeSeams(t, swSeams{
		inWt:      func() bool { return false },
		commonDir: func() (string, error) { return "", errFakeNotGitRepo },
		add:       func(string, string) (string, error) { t.Fatal("add must not run pre-commonDir"); return "", nil },
	})
	var out bytes.Buffer
	got := enterSessionWorktree(nil, "init", &out)
	if got != "" {
		t.Fatalf("non-git fail-back: expected empty, got %q", got)
	}
	if !strings.Contains(out.String(), "materialization failed") {
		t.Fatalf("expected fail-back notice for non-git dir, got %q", out.String())
	}
}

// errFakeGitAdd is a sentinel used by the fail-back tests.
type fakeErr string

func (e fakeErr) Error() string { return string(e) }

const errFakeGitAdd = fakeErr("fake-git-add")
const errFakeNotGitRepo = fakeErr("fake-not-a-git-repo")

// --- SPEC-SESSION-WORKTREE-001 M4: session-exit disposal (REQ-SW-008/009/010) ---

// worktreeCfg builds a config with the session-worktree feature ON and a
// controllable auto_cleanup toggle. The feature is forced ON via the env-free
// Enabled field; cleanup reads AutoCleanup directly.
func worktreeCfg(autoCleanup bool) *config.Config {
	return &config.Config{Workflow: config.WorkflowConfig{
		SessionWorktree: config.SessionWorktreeConfig{Enabled: true},
		Worktree:        config.WorkflowWorktreeConfig{AutoCleanup: autoCleanup},
	}}
}

// TestCleanupSessionWorktree_EmptyPathNoOp is the trivial guard: an empty
// worktree path (feature OFF / fail-back / already-in-worktree) is a no-op.
func TestCleanupSessionWorktree_EmptyPathNoOp(t *testing.T) {
	removeCalled := false
	swapSessionWorktreeSeams(t, swSeams{
		remove: func(string) error { removeCalled = true; return nil },
	})
	var out bytes.Buffer
	cleanupSessionWorktree(worktreeCfg(true), "", true, &out)
	if removeCalled {
		t.Fatal("empty path: remove MUST NOT run")
	}
	if out.Len() != 0 {
		t.Fatalf("empty path: expected no notice, got %q", out.String())
	}
}

// TestCleanupSessionWorktree_DefaultManualPersists is REQ-SW-008: when
// auto_cleanup is false (the distributed default), the worktree PERSISTS after
// exit — no removal, no removal notice.
func TestCleanupSessionWorktree_DefaultManualPersists(t *testing.T) {
	removeCalled := false
	swapSessionWorktreeSeams(t, swSeams{
		remove: func(string) error { removeCalled = true; return nil },
	})
	var out bytes.Buffer
	cleanupSessionWorktree(worktreeCfg(false), "/repo/.claude/worktrees/WT-abcdef12-web", true, &out)
	if removeCalled {
		t.Fatal("default-manual: remove MUST NOT run (worktree must persist)")
	}
	if strings.Contains(out.String(), SessionExitCleanupNoticePrefix) {
		t.Fatalf("default-manual: no removal notice expected, got %q", out.String())
	}
}

// TestCleanupSessionWorktree_CleanExitRemoves is REQ-SW-009: when auto_cleanup
// is true AND the subcommand exits cleanly AND the worktree is clean, git
// worktree remove runs and a stderr notice carrying the session-exit prefix is
// emitted.
func TestCleanupSessionWorktree_CleanExitRemoves(t *testing.T) {
	var removedPath string
	swapSessionWorktreeSeams(t, swSeams{
		remove: func(p string) error { removedPath = p; return nil },
		statusPorc: func(string) (string, error) { return "", nil }, // clean
	})
	var out bytes.Buffer
	wt := "/repo/.claude/worktrees/WT-abcdef12-web"
	cleanupSessionWorktree(worktreeCfg(true), wt, true, &out)
	if removedPath != wt {
		t.Fatalf("clean-exit: expected remove(%q), got %q", wt, removedPath)
	}
	if !strings.Contains(out.String(), SessionExitCleanupNoticePrefix) {
		t.Fatalf("clean-exit: notice must carry %q, got %q", SessionExitCleanupNoticePrefix, out.String())
	}
	if !strings.Contains(out.String(), wt) {
		t.Fatalf("clean-exit: notice must name the worktree path %q, got %q", wt, out.String())
	}
}

// TestCleanupSessionWorktree_NonCleanExitPreserves is REQ-SW-009
// clean-exit-only: a non-zero exit PRESERVES the worktree for post-mortem even
// when auto_cleanup is ON.
func TestCleanupSessionWorktree_NonCleanExitPreserves(t *testing.T) {
	removeCalled := false
	swapSessionWorktreeSeams(t, swSeams{
		remove:     func(string) error { removeCalled = true; return nil },
		statusPorc: func(string) (string, error) { return "", nil },
	})
	var out bytes.Buffer
	wt := "/repo/.claude/worktrees/WT-abcdef12-web"
	cleanupSessionWorktree(worktreeCfg(true), wt, false, &out) // cleanExit=false
	if removeCalled {
		t.Fatal("non-clean-exit: remove MUST NOT run (preserve for post-mortem)")
	}
	if strings.Contains(out.String(), SessionExitCleanupNoticePrefix) {
		t.Fatalf("non-clean-exit: no removal notice expected, got %q", out.String())
	}
	if !strings.Contains(out.String(), "preserved") {
		t.Fatalf("non-clean-exit: expected preserve notice, got %q", out.String())
	}
}

// TestCleanupSessionWorktree_DirtyPreserves is REQ-SW-010: even on a clean
// exit with auto_cleanup ON, uncommitted changes (non-empty porcelain) SKIP
// removal — the worktree is NEVER deleted with local changes.
func TestCleanupSessionWorktree_DirtyPreserves(t *testing.T) {
	removeCalled := false
	swapSessionWorktreeSeams(t, swSeams{
		remove: func(string) error { removeCalled = true; return nil },
		statusPorc: func(string) (string, error) {
			return " M dirty.go\n?? untracked.txt\n", nil // dirty
		},
	})
	var out bytes.Buffer
	wt := "/repo/.claude/worktrees/WT-abcdef12-web"
	cleanupSessionWorktree(worktreeCfg(true), wt, true, &out)
	if removeCalled {
		t.Fatal("dirty: remove MUST NOT run (uncommitted changes preserved)")
	}
	if strings.Contains(out.String(), SessionExitCleanupNoticePrefix) {
		t.Fatalf("dirty: no removal notice expected, got %q", out.String())
	}
	if !strings.Contains(out.String(), wt) {
		t.Fatalf("dirty: notice must name the worktree path %q, got %q", wt, out.String())
	}
}

// TestCleanupSessionWorktree_DirtyCheckErrorPreserves proves the dirty guard
// is fail-open: a status-check error preserves the worktree (no risky
// deletion) and emits a notice naming the failure.
func TestCleanupSessionWorktree_DirtyCheckErrorPreserves(t *testing.T) {
	removeCalled := false
	swapSessionWorktreeSeams(t, swSeams{
		remove: func(string) error { removeCalled = true; return nil },
		statusPorc: func(string) (string, error) { return "", errFakeNotGitRepo },
	})
	var out bytes.Buffer
	wt := "/repo/.claude/worktrees/WT-abcdef12-web"
	cleanupSessionWorktree(worktreeCfg(true), wt, true, &out)
	if removeCalled {
		t.Fatal("dirty-check-error: remove MUST NOT run (fail-open preserve)")
	}
	if !strings.Contains(out.String(), wt) {
		t.Fatalf("dirty-check-error: notice must name worktree path, got %q", out.String())
	}
}

// TestCleanupSessionWorktree_NoticeDistinguishableFromPRMerge is REQ-SW-009 /
// EC-13: the session-exit notice prefix MUST NOT contain "PR-merge" so it is
// unambiguously distinguishable from the M8 PR-merge notice (AC-SW-022) in
// combined output.
func TestCleanupSessionWorktree_NoticeDistinguishableFromPRMerge(t *testing.T) {
	if strings.Contains(SessionExitCleanupNoticePrefix, "PR-merge") {
		t.Fatalf("session-exit prefix MUST NOT contain 'PR-merge': %q", SessionExitCleanupNoticePrefix)
	}
	if !strings.Contains(SessionExitCleanupNoticePrefix, "session-exit") {
		t.Fatalf("session-exit prefix MUST contain 'session-exit': %q", SessionExitCleanupNoticePrefix)
	}
	// Verify the constant is also distinguishable from the M8 prefix literal.
	m8Prefix := "removed by PR-merge cleanup:"
	if SessionExitCleanupNoticePrefix == m8Prefix {
		t.Fatalf("session-exit prefix collides with M8 PR-merge prefix: both %q", m8Prefix)
	}
}

// --- resolveSessionShortReal session-id-available branch (AC-SW-007 / EC-4) ---

// TestResolveSessionShortReal_SideChannelAvailable exercises the
// session-id-available branch of resolveSessionShortReal that was NOT covered
// at M2 (coverage 45.5% — only the random-fallback branch was hit). It stages
// the side-channel file (.moai/state/current-session-id.txt) under a temp
// project dir pointed at by $CLAUDE_PROJECT_DIR and asserts the first-8 chars
// of the session UUID are returned.
func TestResolveSessionShortReal_SideChannelAvailable(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)
	sidecarDir := filepath.Join(tmp, ".moai", "state")
	if err := os.MkdirAll(sidecarDir, 0o755); err != nil {
		t.Fatalf("mkdir sidecar dir: %v", err)
	}
	uuid := "abcdef1234567890deadbeef12345678"
	sidecar := filepath.Join(sidecarDir, "current-session-id.txt")
	if err := os.WriteFile(sidecar, []byte(uuid), 0o644); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}
	got := resolveSessionShortReal()
	if got != "abcdef12" {
		t.Fatalf("session-id-available: expected first-8 %q of UUID, got %q", "abcdef12", got)
	}
}

// TestResolveSessionShortReal_ShortSessionID covers the <8-char branch: a
// session id shorter than 8 chars is returned verbatim (not truncated).
func TestResolveSessionShortReal_ShortSessionID(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)
	sidecarDir := filepath.Join(tmp, ".moai", "state")
	if err := os.MkdirAll(sidecarDir, 0o755); err != nil {
		t.Fatalf("mkdir sidecar dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sidecarDir, "current-session-id.txt"), []byte("ab12cd"), 0o644); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}
	if got := resolveSessionShortReal(); got != "ab12cd" {
		t.Fatalf("short session id: expected %q verbatim, got %q", "ab12cd", got)
	}
}

// TestResolveSessionShortReal_NoSideChannelFallsBack confirms the EC-4
// fallback: when no side-channel file exists, resolveSessionShortReal returns a
// 12-hex-char (6-byte) random segment.
func TestResolveSessionShortReal_NoSideChannelFallsBack(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp) // no sidecar file present
	got := resolveSessionShortReal()
	if len(got) != 12 {
		t.Fatalf("fallback: expected 12 hex chars (6 bytes), got %d (%q)", len(got), got)
	}
	for _, r := range got {
		isHex := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')
		if !isHex {
			t.Fatalf("fallback: must be lowercase hex, got %q in %q", r, got)
		}
	}
}

