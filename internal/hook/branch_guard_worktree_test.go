package hook

// branch_guard_worktree_test.go — SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001 M1.
//
// Exercises the discriminant directory correction (Seam A): the git-context
// query MUST resolve input.CWD (the command's actual cwd), NOT the audit-log
// project dir (the primary checkout). The headline worktree-classification
// tests use a REAL git worktree fixture (t.TempDir + git init + git worktree
// add); the gitcore.ExecCommand indirection (internal/core/git, extracted by
// SPEC-KANBAN-BOARD-001 REQ-KB-005) is used ONLY for the fail-open fallback
// path (AC-WBG-D-005), per AP-D-006 (no mocked discriminant for the headline
// AC).
//
// @MX:NOTE: [AUTO] SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001 M1 worktree discriminant regression suite.

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
)

// failingRevParseMock is a gitcore.ExecCommand replacement that makes
// every `git ... rev-parse ...` invocation exit non-zero, forcing the
// isPrimaryCheckout fail-open path. Used ONLY by AC-WBG-D-004 (audit-log
// placement) — the headline worktree-classification tests use real git
// (AP-D-006).
var failingRevParseMock = func(name string, args ...string) *exec.Cmd {
	return exec.Command("false")
}

// containsCwdToken reports whether the audit-log body records the resolved
// cwd. AP-D-003 requires the resolved cwd to be observable in the advisory so
// a silent $CLAUDE_PROJECT_DIR fallback cannot re-introduce the bug. The check
// matches either a `cwd="<path>"` structured token or the bare path substring
// (robust to the exact format string).
func containsCwdToken(body, cwd string) bool {
	return strings.Contains(body, "cwd=") && strings.Contains(body, cwd)
}

// worktreeFixture builds a primary git repo plus a real git worktree under
// t.TempDir() and returns (primaryDir, worktreeDir). Both directories are
// real git contexts — isPrimaryCheckout classifies them differently. Skips
// when the git binary is absent (AC §B test-infrastructure note).
func worktreeFixture(t *testing.T) (primaryDir, worktreeDir string) {
	t.Helper()
	requireGit(t)
	primaryDir = t.TempDir()
	gitInitRepo(t, primaryDir)
	parent := t.TempDir()
	worktreeDir = filepath.Join(parent, "wt")
	mustRunGit(t, primaryDir, "worktree", "add", worktreeDir, "-b", "wt-branch")
	return primaryDir, worktreeDir
}

// TestIsPrimaryCheckout_Worktree is the AC-WBG-D-001 headline assertion: a
// real git worktree cwd classifies as NOT primary. The existing
// TestIsPrimaryCheckout/Worktree subtest covers the same contract; this
// dedicated top-level test matches the AC verification command's -run pattern
// exactly and pins the headline classification independently of the suite's
// other subtests.
func TestIsPrimaryCheckout_Worktree(t *testing.T) {
	t.Parallel()
	_, worktreeDir := worktreeFixture(t)
	isPrimary, err := isPrimaryCheckout(worktreeDir)
	if err != nil {
		t.Fatalf("isPrimaryCheckout(worktree) err = %v, want nil", err)
	}
	if isPrimary {
		t.Fatalf("isPrimaryCheckout(worktree) = true, want false")
	}
}

// TestBranchGuard_Worktree_Allow is the AC-WBG-D-002 end-to-end regression:
// a branch-state git command (git rebase) running with input.CWD inside a
// worktree MUST be allowed (not denied). Before the Seam A fix,
// checkBranchState fed the primary checkout path into isPrimaryCheckout,
// misclassifying the worktree context as primary and denying the command.
// This is the RED driver for the fix (it fails against the pre-fix code).
func TestBranchGuard_Worktree_Allow(t *testing.T) {
	t.Parallel()
	primaryDir, worktreeDir := worktreeFixture(t)

	input := &HookInput{
		ToolName:  "Bash",
		CWD:       worktreeDir,
		ToolInput: json.RawMessage(`{"command": "git rebase origin/main"}`),
	}
	// The audit-log project dir argument stays pinned to the primary checkout
	// (REQ-WBG-D-004) — only the git-context query follows input.CWD.
	decision, reason := checkBranchState(input, primaryDir)
	if decision != "" {
		t.Fatalf("checkBranchState(worktree cwd) decision = %q, want \"\" (allow); reason=%q", decision, reason)
	}
	if reason != "" {
		t.Fatalf("checkBranchState(worktree cwd) reason = %q, want \"\"", reason)
	}
}

// TestBranchGuard_Primary_Still_Denies is the AC-WBG-D-003 regression anchor:
// the primary-checkout deny path is preserved post-fix. A branch-state
// command with input.CWD = primary MUST still be denied. Failure here means
// the fix over-corrected and broke the guard's core contract.
func TestBranchGuard_Primary_Still_Denies(t *testing.T) {
	t.Parallel()
	primaryDir, _ := worktreeFixture(t)

	input := &HookInput{
		ToolName:  "Bash",
		CWD:       primaryDir,
		ToolInput: json.RawMessage(`{"command": "git switch feature-x"}`),
	}
	decision, reason := checkBranchState(input, primaryDir)
	if decision != DecisionDeny {
		t.Fatalf("checkBranchState(primary cwd) decision = %q, want %q", decision, DecisionDeny)
	}
	const wantPrefix = "BRANCH_GUARD_VIOLATION: git switch"
	if len(reason) < len(wantPrefix) || reason[:len(wantPrefix)] != wantPrefix {
		t.Fatalf("checkBranchState(primary cwd) reason = %q, want prefix %q", reason, wantPrefix)
	}
}

// TestBranchGuard_AuditLog_Placement is the AC-WBG-D-004 invariant: when the
// command cwd is a worktree but rev-parse fails (simulated via the
// gitcore.ExecCommand indirection), the fail-open advisory MUST be appended to
// <primary>/.moai/logs/branch-guard-audit.log and NO file created at
// <worktree>/.moai/logs/. AP-D-003: the resolved cwd MUST be observable in
// the audit entry.
func TestBranchGuard_AuditLog_Placement(t *testing.T) {
	primaryDir, worktreeDir := worktreeFixture(t)

	// Simulate rev-parse failure at the worktree cwd so the fail-open path
	// fires. The mock rejects every git invocation.
	orig := gitcore.ExecCommand
	t.Cleanup(func() { gitcore.ExecCommand = orig })
	gitcore.ExecCommand = failingRevParseMock

	input := &HookInput{
		ToolName:  "Bash",
		CWD:       worktreeDir,
		ToolInput: json.RawMessage(`{"command": "git rebase origin/main"}`),
	}
	decision, reason := checkBranchState(input, primaryDir)
	if decision != "" {
		t.Fatalf("checkBranchState(simulated rev-parse fail) decision = %q, want \"\" (fail-open allow); reason=%q", decision, reason)
	}

	primaryLog := filepath.Join(primaryDir, branchGuardAuditRelPath)
	if _, err := os.Stat(primaryLog); err != nil {
		t.Fatalf("audit log NOT created at primary %s: %v", primaryLog, err)
	}
	worktreeLog := filepath.Join(worktreeDir, branchGuardAuditRelPath)
	if _, err := os.Stat(worktreeLog); err == nil {
		t.Fatalf("audit log MUST NOT be created at worktree %s — it must stay on the primary", worktreeLog)
	}

	// AP-D-003: the resolved cwd (the worktree) MUST be observable in the
	// audit entry so a silent $CLAUDE_PROJECT_DIR fallback can't re-introduce
	// the bug. Read the log and assert it records the worktree cwd.
	data, err := os.ReadFile(primaryLog)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	body := string(data)
	if !containsCwdToken(body, worktreeDir) {
		t.Fatalf("audit log entry does not record the resolved worktree cwd %q; AP-D-003 observability violated.\nlog:\n%s", worktreeDir, body)
	}
}

// TestBranchGuard_FailOpen_NonGitCwd is the AC-WBG-D-005 contract: when
// input.CWD is a non-git directory (real git invoked, rev-parse exits
// non-zero), the guard MUST fail open (allow) AND write the audit entry to
// the primary. No mock is used here — the real `git rev-parse` fails against
// a non-git t.TempDir() (AP-D-006 headline-test discipline).
func TestBranchGuard_FailOpen_NonGitCwd(t *testing.T) {
	t.Parallel()
	requireGit(t)
	primaryDir := t.TempDir()
	// Ensure primaryDir is itself non-git so the audit-log placement assertion
	// is meaningful (the audit dir is primaryDir; rev-parse fails there too,
	// but appendBranchGuardAdvisory creates the .moai/logs tree regardless).
	nonGitCwd := t.TempDir()

	input := &HookInput{
		ToolName:  "Bash",
		CWD:       nonGitCwd,
		ToolInput: json.RawMessage(`{"command": "git rebase origin/main"}`),
	}
	decision, reason := checkBranchState(input, primaryDir)
	if decision != "" {
		t.Fatalf("checkBranchState(non-git cwd) decision = %q, want \"\" (fail-open allow); reason=%q", decision, reason)
	}
	primaryLog := filepath.Join(primaryDir, branchGuardAuditRelPath)
	if _, err := os.Stat(primaryLog); err != nil {
		t.Fatalf("audit log NOT created at primary %s: %v", primaryLog, err)
	}
	// AP-D-003: the resolved cwd (nonGitCwd) MUST be observable.
	data, err := os.ReadFile(primaryLog)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	if !containsCwdToken(string(data), nonGitCwd) {
		t.Fatalf("audit log does not record the non-git cwd %q (AP-D-003)", nonGitCwd)
	}
}

// TestBranchGuard_Exempt is the AC-WBG-D-006 regression anchor: the
// MOAI_BRANCH_GUARD_EXEMPT=1 exemption fires before the discriminant, so a
// branch-state command in the primary checkout is allowed. Non-parallel:
// t.Setenv mutates process-global env.
func TestBranchGuard_Exempt(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "1")
	t.Cleanup(func() { t.Setenv(branchGuardExemptEnv, "") })
	primaryDir, _ := worktreeFixture(t)

	input := &HookInput{
		ToolName:  "Bash",
		CWD:       primaryDir,
		AgentType: "manager-develop",
		ToolInput: json.RawMessage(`{"command": "git switch feature-x"}`),
	}
	decision, reason := checkBranchState(input, primaryDir)
	if decision != "" {
		t.Fatalf("checkBranchState(exempt) decision = %q, want \"\" (allow); reason=%q", decision, reason)
	}
}
