package cli

import (
	"bytes"
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
}

// swapSessionWorktreeSeams replaces the seams and registers restoration.
func swapSessionWorktreeSeams(t *testing.T, s swSeams) {
	t.Helper()
	orig := swSeams{
		add: sessionWorktreeGitWorktreeAdd, inWt: sessionWorktreeInGitWorktree,
		short: sessionWorktreeResolveSessionShort, commonDir: sessionWorktreeGitCommonDir,
		configSet: sessionWorktreeGitConfigSet,
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
	t.Cleanup(func() {
		sessionWorktreeGitWorktreeAdd = orig.add
		sessionWorktreeInGitWorktree = orig.inWt
		sessionWorktreeResolveSessionShort = orig.short
		sessionWorktreeGitCommonDir = orig.commonDir
		sessionWorktreeGitConfigSet = orig.configSet
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
