package hook

// branch_guard_protected_commit_test.go — the protected-branch commit deny
// suite (SPEC-MAIN-COMMIT-BAN-001 M2, RED→GREEN core).
//
// Structure mirrors the sibling suites: the HEADLINE deny test (AC-4) runs the
// fully-real path — real git repo on branch main, real Seam A discriminant,
// REAL ResolveHeadBranch (`git branch --show-current`) — per the AP-D-006
// discipline that a headline AC is never carried by a mock. The matrix arms
// that need branch names the fixture does not have (develop, detached, error)
// use the package-var swap idiom over resolveHeadBranch (the M6 deny-origin
// precedent), and those tests are NON-PARALLEL while the swap is live.
//
// Documented under-match residual (spec §E E-1, accepted fail-open): a
// `git -C <primary-path> commit` typed from a WORKTREE cwd classifies as
// worktree by the cwd-based discriminant and is NOT caught — the same
// under-match direction as the documented `git -C <path> branch` residual in
// the flagclass classifier. Obfuscated shell-wrapped forms
// (`bash -c "git commit …"`) are equally out of scope, by the same
// quoted-span collapse that guards the branch-state class.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// mainRepoFixture builds a real primary git repo whose checked-out branch is
// main: gitInitRepo creates the repo (whatever init.defaultBranch yields) and
// the branch is renamed in place. Skips when git is absent.
func mainRepoFixture(t *testing.T) string {
	t.Helper()
	requireGit(t)
	repo := t.TempDir()
	gitInitRepo(t, repo)
	mustRunGit(t, repo, "branch", "-m", "main")
	return repo
}

// headBranchStub is the counting stub swapped over the resolveHeadBranch
// package var. calls counts invocations — the AC-7 zero-invocation assertion
// reads it.
type headBranchStub struct {
	calls  int
	branch string
	err    error
}

func (s *headBranchStub) resolve(string) (string, error) {
	s.calls++
	return s.branch, s.err
}

// swapHeadBranchStub installs stub over resolveHeadBranch for the test's
// lifetime and restores it on cleanup. Non-parallel contract: the caller test
// MUST NOT call t.Parallel (package-global swap, TestBranchStatePatterns_
// Blankable precedent).
func swapHeadBranchStub(t *testing.T, stub *headBranchStub) {
	t.Helper()
	orig := resolveHeadBranch
	t.Cleanup(func() { resolveHeadBranch = orig })
	resolveHeadBranch = stub.resolve
}

// protectedCommitInput builds a Bash HookInput for command running at cwd.
// The command is marshaled (not string-concatenated) so commands carrying
// double quotes — the quoted-data matrix arms — arrive as valid JSON, exactly
// as Claude Code emits them.
func protectedCommitInput(cwd, command string) *HookInput {
	payload, err := json.Marshal(map[string]string{"command": command})
	if err != nil {
		// Marshal of a plain map[string]string cannot fail; kept for shape.
		panic(err)
	}
	return &HookInput{
		ToolName:  "Bash",
		CWD:       cwd,
		ToolInput: payload,
	}
}

// TestProtectedCommit_DenyOnProtectedBranch is the AC-4 headline: commit /
// revert / cherry-pick on a protected branch at the primary checkout, guard
// configured, agent non-exempt → DecisionDeny with the BRANCH_GUARD_VIOLATION
// prefix, the protected branch named, "primary checkout" named, and NO
// manager-git delegation suggested as remediation (the t43 rule). Real
// resolver end to end.
func TestProtectedCommit_DenyOnProtectedBranch(t *testing.T) {
	t.Parallel()
	repo := mainRepoFixture(t)
	denyList := []string{"main"}

	cases := []string{
		"git commit -m x",
		"git commit --amend -m y",
		"git revert HEAD",
		"git cherry-pick abc123",
		"git commit -a -m z",
	}
	for _, command := range cases {
		command := command
		t.Run(command, func(t *testing.T) {
			t.Parallel()
			decision, reason := checkProtectedCommit(protectedCommitInput(repo, command), repo, denyList)
			if decision != DecisionDeny {
				t.Fatalf("checkProtectedCommit(%q) decision = %q, want %q", command, decision, DecisionDeny)
			}
			const prefix = "BRANCH_GUARD_VIOLATION:"
			if !strings.HasPrefix(reason, prefix) {
				t.Fatalf("reason %q lacks the %q sentinel prefix (REQ-2.2)", reason, prefix)
			}
			if !strings.Contains(reason, "main") {
				t.Fatalf("reason %q does not name the protected branch", reason)
			}
			if !strings.Contains(reason, "primary checkout") {
				t.Fatalf("reason %q does not name the primary checkout", reason)
			}
			// The t43 rule: the remediation must NOT suggest delegating to a
			// manager-git agent. The old wording "invoke via manager-git" and
			// any "delegate to manager-git" phrasing are the failure shapes.
			for _, banned := range []string{"invoke via manager-git", "delegate to manager-git", "delegating to manager-git"} {
				if strings.Contains(reason, banned) {
					t.Fatalf("reason %q suggests the banned %q remediation (t43)", reason, banned)
				}
			}
		})
	}
}

// TestProtectedCommit_AllowMatrix is the allow arms of the plan M2 matrix.
// Arms that need a branch name the fixture cannot have (develop) or an
// injected failure swap the resolver stub; arms about command shape
// (quoted data, foreign carrier, PowerShell payload, case) run real.
func TestProtectedCommit_AllowMatrix(t *testing.T) {
	requireGit(t)

	t.Run("NonProtectedBranch", func(t *testing.T) {
		// AC-5: commit on develop @ primary + configured → allow.
		repo := mainRepoFixture(t)
		stub := &headBranchStub{branch: "develop"}
		swapHeadBranchStub(t, stub)
		decision, reason := checkProtectedCommit(protectedCommitInput(repo, "git commit -m x"), repo, []string{"main"})
		if decision != "" {
			t.Fatalf("decision = %q reason = %q, want allow on non-protected branch", decision, reason)
		}
	})

	t.Run("WorktreeCwd", func(t *testing.T) {
		// AC-6: commit on a protected branch @ worktree cwd → allow (the
		// discriminant holds). The check must short-circuit at !isPrimary
		// BEFORE the resolver runs — asserted via stub.calls == 0.
		primaryDir, worktreeDir := worktreeFixture(t)
		stub := &headBranchStub{branch: "main"}
		swapHeadBranchStub(t, stub)
		decision, reason := checkProtectedCommit(protectedCommitInput(worktreeDir, "git commit -m x"), primaryDir, []string{"main"})
		if decision != "" {
			t.Fatalf("decision = %q reason = %q, want allow at worktree cwd", decision, reason)
		}
		if stub.calls != 0 {
			t.Fatalf("resolveHeadBranch invoked %d times at a worktree cwd, want 0 (short-circuit at the discriminant)", stub.calls)
		}
	})

	t.Run("EmptyList", func(t *testing.T) {
		// AC-7: empty deny list → allow AND zero resolver invocations
		// (REQ-2.5), for nil and for the empty-slice spellings.
		repo := mainRepoFixture(t)
		for name, list := range map[string][]string{"nil": nil, "empty": {}} {
			stub := &headBranchStub{branch: "main"}
			swapHeadBranchStub(t, stub)
			decision, _ := checkProtectedCommit(protectedCommitInput(repo, "git commit -m x"), repo, list)
			if decision != "" {
				t.Fatalf("%s: decision = %q, want allow", name, decision)
			}
			if stub.calls != 0 {
				t.Fatalf("%s: resolver invoked %d times with an empty list, want 0 (REQ-2.5)", name, stub.calls)
			}
		}
	})

	t.Run("HeadResolutionError", func(t *testing.T) {
		// AC-8: resolver error → allow + advisory appended to the audit log
		// at the audit-log project dir, with the resolved cwd recorded.
		repo := mainRepoFixture(t)
		auditDir := t.TempDir() // the audit-log project dir argument
		stub := &headBranchStub{err: os.ErrPermission}
		swapHeadBranchStub(t, stub)
		decision, reason := checkProtectedCommit(protectedCommitInput(repo, "git commit -m x"), auditDir, []string{"main"})
		if decision != "" {
			t.Fatalf("decision = %q reason = %q, want fail-open allow", decision, reason)
		}
		logPath := filepath.Join(auditDir, branchGuardAuditRelPath)
		data, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("audit log not appended at %s: %v", logPath, err)
		}
		body := string(data)
		if !containsCwdToken(body, repo) {
			t.Fatalf("audit entry does not record the resolved cwd %q (AP-D-003):\n%s", repo, body)
		}
	})

	t.Run("DetachedHead", func(t *testing.T) {
		// AC-8 twin: resolver returns ("", nil) → allow, deliberately.
		repo := mainRepoFixture(t)
		stub := &headBranchStub{branch: ""}
		swapHeadBranchStub(t, stub)
		decision, reason := checkProtectedCommit(protectedCommitInput(repo, "git commit -m x"), repo, []string{"main"})
		if decision != "" {
			t.Fatalf("decision = %q reason = %q, want allow on detached HEAD", decision, reason)
		}
	})

	t.Run("ExemptEnv", func(t *testing.T) {
		// AC-9 axis 1: MOAI_BRANCH_GUARD_EXEMPT=1 suppresses the deny.
		// Non-parallel: t.Setenv mutates process-global env.
		repo := mainRepoFixture(t)
		t.Setenv(branchGuardExemptEnv, "1")
		t.Cleanup(func() { t.Setenv(branchGuardExemptEnv, "") })
		decision, reason := checkProtectedCommit(protectedCommitInput(repo, "git commit -m x"), repo, []string{"main"})
		if decision != "" {
			t.Fatalf("decision = %q reason = %q, want allow under the env exemption", decision, reason)
		}
	})

	t.Run("ExemptAgentIdentity", func(t *testing.T) {
		// AC-9 axis 2: AgentType == "manager-git" suppresses the deny.
		repo := mainRepoFixture(t)
		input := protectedCommitInput(repo, "git commit -m x")
		input.AgentType = "manager-git"
		decision, reason := checkProtectedCommit(input, repo, []string{"main"})
		if decision != "" {
			t.Fatalf("decision = %q reason = %q, want allow under the identity exemption", decision, reason)
		}
	})
}

// TestProtectedCommit_CommandShapeMatrix covers the matcher arms the plan M2
// matrix names, all real-resolver against a repo on main:
//
//   - `git commit -m "git switch main"` DENIES — the quoted text is data, but
//     the command verb IS a commit on the protected branch (the plan's
//     corrected F3 cell);
//   - `moai todo add "git commit -m x"` allows — the quoted git prose rides a
//     foreign command with no commit verb at command position;
//   - a compound `git status && git commit -m x` DENIES;
//   - a PowerShell -Command payload carrying the commit DENIES;
//   - case-folded `GIT COMMIT -M X` DENIES.
func TestProtectedCommit_CommandShapeMatrix(t *testing.T) {
	t.Parallel()
	repo := mainRepoFixture(t)
	denyList := []string{"main"}

	cases := []struct {
		command string
		want    string
	}{
		{`git commit -m "git switch main"`, DecisionDeny},
		{`moai todo add "git commit -m x"`, ""},
		{"git status && git commit -m x", DecisionDeny},
		{`pwsh -Command "git commit -m x"`, DecisionDeny},
		{"GIT COMMIT -M X", DecisionDeny},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.command, func(t *testing.T) {
			t.Parallel()
			decision, reason := checkProtectedCommit(protectedCommitInput(repo, tc.command), repo, denyList)
			if decision != tc.want {
				t.Fatalf("checkProtectedCommit(%q) decision = %q reason = %q, want %q", tc.command, decision, reason, tc.want)
			}
		})
	}
}

// TestMatchProtectedCommitCommand is the matcher-level table (the sibling of
// the flagclass matcher tests): deny-side and allow-side vocabulary without
// any git process.
func TestMatchProtectedCommitCommand(t *testing.T) {
	t.Parallel()
	cases := []struct {
		command string
		want    bool
	}{
		{"git commit -m x", true},
		{"git revert HEAD", true},
		{"git cherry-pick abc", true},
		{"git commit --amend", true},
		{"git status", false},
		{"git push origin main", false},
		{"git pull origin main", false},
		{"git tag v1.0", false},
		{"git notes add", false},
		{"git precommit lint", false},              // \b word anchor: no commit verb
		{`moai todo add "git commit -m x"`, false}, // quoted git prose is data
		{"echo committed", false},
	}
	for _, tc := range cases {
		if got := matchProtectedCommitCommand(tc.command); got != tc.want {
			t.Errorf("matchProtectedCommitCommand(%q) = %v, want %v", tc.command, got, tc.want)
		}
	}
}

// TestPreTool_ProtectedCommit_Handle is the AC-4 Given/When/Then at the
// handler surface: preToolHandler.Handle with BranchGuard.Enabled AND
// deny_commits_on == ["main"] denies a `git commit` in a real primary
// checkout on main.
func TestPreTool_ProtectedCommit_Handle(t *testing.T) {
	requireGit(t)
	repo := mainRepoFixture(t)

	cfg := config.NewDefaultConfig()
	cfg.Workflow.BranchGuard.Enabled = true
	cfg.Workflow.BranchGuard.DenyCommitsOn = []string{"main"}

	handler := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: cfg},
		policy:     DefaultSecurityPolicy(),
		projectDir: repo,
	}
	input := &HookInput{
		SessionID:     "sess-t1337-handle",
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		AgentType:     "manager-develop",
		CWD:           repo,
		ToolInput:     json.RawMessage(`{"command": "git commit -m x"}`),
	}
	out, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle err = %v", err)
	}
	if decisionOf(out) != DecisionDeny {
		t.Fatalf("Handle decision = %v, want %q", decisionOf(out), DecisionDeny)
	}
	reason := reasonOf(out)
	if !strings.HasPrefix(reason, "BRANCH_GUARD_VIOLATION:") {
		t.Fatalf("Handle reason %q lacks the sentinel prefix", reason)
	}
}

// TestPreTool_ProtectedCommit_EmptyListInert proves the call-site inertness:
// BranchGuard.Enabled true but deny_commits_on empty (the shipped default) →
// a git commit at a primary checkout passes through Handle without a deny and
// without resolving HEAD.
func TestPreTool_ProtectedCommit_EmptyListInert(t *testing.T) {
	requireGit(t)
	repo := mainRepoFixture(t)

	cfg := config.NewDefaultConfig()
	cfg.Workflow.BranchGuard.Enabled = true
	cfg.Workflow.BranchGuard.DenyCommitsOn = nil

	stub := &headBranchStub{branch: "main"}
	swapHeadBranchStub(t, stub)

	handler := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: cfg},
		policy:     DefaultSecurityPolicy(),
		projectDir: repo,
	}
	input := &HookInput{
		SessionID:     "sess-t1337-inert",
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		AgentType:     "manager-develop",
		CWD:           repo,
		ToolInput:     json.RawMessage(`{"command": "git commit -m x"}`),
	}
	out, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle err = %v", err)
	}
	if decisionOf(out) == DecisionDeny {
		t.Fatalf("Handle denied with an empty deny list — REQ-2.5 violated")
	}
	if stub.calls != 0 {
		t.Fatalf("resolver invoked %d times with an empty list, want 0", stub.calls)
	}
}
