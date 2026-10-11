package hook

// branch_guard_integration_verb_test.go — SPEC-LOCAL-MAIN-FLOW-001 (card t1616),
// audit finding F8 (AC-LMF-003, clause V10), route A.
//
// The PreToolUse branch guard's verdict on the designed landing verb had never
// been observed. This test calls the real in-process hook handler
// (preToolHandler.Handle) with the branch guard ON and
// workflow.local_main_integration.enabled ON, against a PRIMARY checkout
// fixture (newBranchGuardRepoFixture: git-dir == git-common-dir, not a
// worktree), and observes two verdicts:
//
//   - ALLOWED: the Bash command `moai integration merge --card t1616` is not
//     denied and carries no BRANCH_GUARD_VIOLATION sentinel.
//   - DENIED: the Bash command `git merge --no-ff WT-10-10-class` is denied with
//     a reason starting with BRANCH_GUARD_VIOLATION:.
//
// Limits (not covered here): the handler is called in-process. The installed
// moai binary's argv path and Claude Code's runtime hook wiring are not
// exercised. Route B (an operator-terminal end-to-end probe) is not done.
//
// The local_main_integration flag is set in the fixture config because the
// route-A design requires it. The hook handler does not read that flag; the
// CLI merge-target resolver (internal/cli/integration_merge.go) reads it. This
// test therefore cannot show that the flag changes any hook verdict.

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// TestBranchGuardRouteAIntegrationVerb observes the ALLOWED and DENIED verdicts
// described in the file header. Not parallel: the parent calls t.Setenv, which
// mutates process-global environment.
func TestBranchGuardRouteAIntegrationVerb(t *testing.T) {
	t.Setenv(branchGuardExemptEnv, "")
	repo := newBranchGuardRepoFixture(t)

	// Non-vacuity: the fixture must classify as a primary checkout, not a
	// worktree, or the DENIED observation would not exercise the primary path.
	primary, err := isPrimaryCheckout(repo)
	if err != nil || !primary {
		t.Fatalf("fixture is not a primary checkout: isPrimaryCheckout = %t, err = %v", primary, err)
	}
	t.Logf("fixture: repo=%q primary=%t", repo, primary)

	cfg := cfgWithBranchGuard(true)
	cfg.Workflow.LocalMainIntegration.Enabled = true
	handler := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: cfg},
		policy:     DefaultSecurityPolicy(),
		projectDir: repo,
	}

	// observe sends one Bash command through the handler as a main-thread
	// payload: no AgentType, so the identity exemption does not apply, and CWD
	// at the primary checkout.
	observe := func(t *testing.T, command string) (decision, reason string) {
		t.Helper()
		toolInput, err := json.Marshal(map[string]string{"command": command})
		if err != nil {
			t.Fatalf("marshal tool input: %v", err)
		}
		input := &HookInput{
			SessionID:     "sess-t1616-route-a",
			HookEventName: "PreToolUse",
			ToolName:      "Bash",
			CWD:           repo,
			ToolInput:     json.RawMessage(toolInput),
		}
		out, err := handler.Handle(context.Background(), input)
		if err != nil {
			t.Fatalf("Handle err = %v", err)
		}
		return decisionOf(out), reasonOf(out)
	}

	t.Run("ALLOWED_moai_integration_merge", func(t *testing.T) {
		const command = "moai integration merge --card t1616"
		decision, reason := observe(t, command)
		t.Logf("verdict case=ALLOWED command=%q decision=%q reason=%q", command, decision, reason)
		if decision == DecisionDeny {
			t.Fatalf("decision = %q, want not deny (reason=%q)", decision, reason)
		}
		if strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
			t.Fatalf("reason = %q carries the %s: sentinel, want none", reason, branchGuardViolationPrefix)
		}
	})

	t.Run("DENIED_git_merge_no_ff", func(t *testing.T) {
		const command = "git merge --no-ff WT-10-10-class"
		decision, reason := observe(t, command)
		t.Logf("verdict case=DENIED command=%q decision=%q reason=%q", command, decision, reason)
		if decision != DecisionDeny {
			t.Fatalf("decision = %q, want %q (reason=%q)", decision, DecisionDeny, reason)
		}
		if !strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
			t.Fatalf("reason = %q, want prefix %q", reason, branchGuardViolationPrefix+":")
		}
	})
}
