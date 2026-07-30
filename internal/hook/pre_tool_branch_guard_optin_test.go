package hook

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// ---------------------------------------------------------------------------
// SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 — M3 config-gate tests:
//   * TestPreTool_BranchGuard_ConfigGate_DefaultOff   — AC-REQ-1a/1b: default
//     (and explicit false) config leaves a mutating command NOT denied.
//   * TestPreTool_BranchGuard_ConfigGate_FlagFlips    — AC-REQ-3: the ONLY
//     variable that flips the decision is the BranchGuard.Enabled flag —
//     proving the handler reads the ConfigProvider (non-vacuity).
//   * TestPreTool_BranchGuard_ConfigGate_Exemptions   — AC-REQ-6a/6b: when
//     enabled, MOAI_BRANCH_GUARD_EXEMPT=1 and AgentType=="manager-git" still
//     bypass (backward-compat, exemption additive above the gate... actually
//     consulted only on the enabled path).
//   * TestPreTool_BranchGuard_ConfigGate_FailOpen     — AC-REQ-6c: when
//     enabled but projectDir is a non-git dir (uncertainty), the handler
//     returns allow (fail-open preserved).
//
// Every test varies ONLY the config flag (or the exemption input), proving the
// gate is the sole decision variable. The repo fixture is a real primary git
// checkout so the deny path is genuinely reachable when enabled.
// ---------------------------------------------------------------------------

// cfgWithBranchGuard returns a *config.Config identical to the default except
// for the BranchGuard.Enabled flag. This is the non-vacuity primitive: every
// other field is held constant so the flag is provably the only variable.
func cfgWithBranchGuard(enabled bool) *config.Config {
	cfg := config.NewDefaultConfig()
	cfg.Workflow.BranchGuard.Enabled = enabled
	return cfg
}

// TestPreTool_BranchGuard_ConfigGate_DefaultOff covers AC-REQ-1a (default off)
// and AC-REQ-1b (explicit false): a mutating `git reset --hard` command in a
// real primary checkout MUST NOT be denied when BranchGuard.Enabled is false.
// The precondition sanity arm (enabled=true denies) is shared with the
// FlagFlips test below.
func TestPreTool_BranchGuard_ConfigGate_DefaultOff(t *testing.T) {
	requireGit(t)
	repo := newBranchGuardRepoFixture(t)
	t.Setenv(branchGuardExemptEnv, "")

	cases := []struct {
		name    string
		enabled bool
		command string
	}{
		{"default_off_reset_hard", false, "git reset --hard HEAD~1"},
		{"explicit_false_switch", false, "git switch feature/x"},
		{"default_off_switch", false, "git switch -c feat/y"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			handler := &preToolHandler{
				cfg:        &mockConfigProvider{cfg: cfgWithBranchGuard(tc.enabled)},
				policy:     DefaultSecurityPolicy(),
				projectDir: repo,
			}
			input := &HookInput{
				SessionID:     "sess-bg-default-off",
				HookEventName: "PreToolUse",
				ToolName:      "Bash",
				AgentType:     "manager-develop",
				ToolInput:     json.RawMessage(`{"command": "` + tc.command + `"}`),
			}
			out, err := handler.Handle(context.Background(), input)
			if err != nil {
				t.Fatalf("Handle err = %v", err)
			}
			if decision := decisionOf(out); decision == DecisionDeny {
				t.Fatalf("Handle(enabled=%v, %q) = deny; want non-deny (REQ-1 default-off)", tc.enabled, tc.command)
			}
			if reason := reasonOf(out); strings.Contains(reason, branchGuardViolationPrefix) {
				t.Fatalf("Handle(enabled=%v, %q) reason contains %q: %q (REQ-1)", tc.enabled, tc.command, branchGuardViolationPrefix, reason)
			}
		})
	}
}

// TestPreTool_BranchGuard_ConfigGate_FlagFlips covers AC-REQ-3: with everything
// else identical (same provider shape, same repo, same input), flipping ONLY
// BranchGuard.Enabled false→true flips the decision allow→deny. This is the
// non-vacuity proof that the handler reads the ConfigProvider.
func TestPreTool_BranchGuard_ConfigGate_FlagFlips(t *testing.T) {
	requireGit(t)
	repo := newBranchGuardRepoFixture(t)
	t.Setenv(branchGuardExemptEnv, "")

	command := "git switch -c feat/flagflips"
	input := &HookInput{
		SessionID:     "sess-bg-flagflips",
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		AgentType:     "manager-develop",
		ToolInput:     json.RawMessage(`{"command": "` + command + `"}`),
	}

	// Disabled → allow (not denied).
	hOff := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: cfgWithBranchGuard(false)},
		policy:     DefaultSecurityPolicy(),
		projectDir: repo,
	}
	outOff, err := hOff.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle(disabled) err = %v", err)
	}
	if decision := decisionOf(outOff); decision == DecisionDeny {
		t.Fatalf("Handle(disabled) = deny; want non-deny — the flag is the only variable and it is false here")
	}

	// Enabled → deny (same repo, same input, ONLY the flag differs).
	hOn := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: cfgWithBranchGuard(true)},
		policy:     DefaultSecurityPolicy(),
		projectDir: repo,
	}
	outOn, err := hOn.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle(enabled) err = %v", err)
	}
	if decision := decisionOf(outOn); decision != DecisionDeny {
		t.Fatalf("Handle(enabled) = %q, want %q — flipping ONLY the flag MUST flip the decision (AC-REQ-3 non-vacuity)", decision, DecisionDeny)
	}
	if reason := reasonOf(outOn); !strings.HasPrefix(reason, branchGuardViolationPrefix+":") {
		t.Fatalf("Handle(enabled) reason = %q, want prefix %q:", reason, branchGuardViolationPrefix)
	}
}

// TestPreTool_BranchGuard_ConfigGate_Exemptions covers AC-REQ-6a (env) and
// AC-REQ-6b (agent identity): when the guard IS enabled, the existing exemption
// paths still bypass it. Backward-compat — the gate is additive below the
// exemption logic in the deny decision order.
//
// NOTE: cannot t.Parallel — subtests mutate the process-global
// MOAI_BRANCH_GUARD_EXEMPT env var via t.Setenv.
func TestPreTool_BranchGuard_ConfigGate_Exemptions(t *testing.T) {
	requireGit(t)
	repo := newBranchGuardRepoFixture(t)

	t.Run("MOAI_BRANCH_GUARD_EXEMPT_env", func(t *testing.T) {
		t.Setenv(branchGuardExemptEnv, "1")
		handler := &preToolHandler{
			cfg:        &mockConfigProvider{cfg: cfgWithBranchGuard(true)},
			policy:     DefaultSecurityPolicy(),
			projectDir: repo,
		}
		input := &HookInput{
			SessionID:     "sess-bg-exempt-env",
			HookEventName: "PreToolUse",
			ToolName:      "Bash",
			AgentType:     "manager-develop",
			ToolInput:     json.RawMessage(`{"command": "git reset --hard HEAD~1"}`),
		}
		out, err := handler.Handle(context.Background(), input)
		if err != nil {
			t.Fatalf("Handle err = %v", err)
		}
		if decision := decisionOf(out); decision == DecisionDeny {
			t.Fatalf("Handle(enabled + EXEMPT=1) = deny; want non-deny (AC-REQ-6a exemption bypasses regardless of flag)")
		}
	})

	t.Run("manager-git_agent_identity", func(t *testing.T) {
		t.Setenv(branchGuardExemptEnv, "")
		handler := &preToolHandler{
			cfg:        &mockConfigProvider{cfg: cfgWithBranchGuard(true)},
			policy:     DefaultSecurityPolicy(),
			projectDir: repo,
		}
		input := &HookInput{
			SessionID:     "sess-bg-exempt-agent",
			HookEventName: "PreToolUse",
			ToolName:      "Bash",
			AgentType:     "manager-git",
			ToolInput:     json.RawMessage(`{"command": "git switch feature/x"}`),
		}
		out, err := handler.Handle(context.Background(), input)
		if err != nil {
			t.Fatalf("Handle err = %v", err)
		}
		if decision := decisionOf(out); decision == DecisionDeny {
			t.Fatalf("Handle(enabled + AgentType=manager-git) = deny; want non-deny (AC-REQ-6b exemption bypasses regardless of flag)")
		}
	})
}

// TestPreTool_BranchGuard_ConfigGate_FailOpen covers AC-REQ-6c: when the guard
// is enabled but projectDir is a NON-git directory (rev-parse exits non-zero),
// the handler returns allow (fail-open preserved). The deny requires positive
// evidence of a primary checkout; uncertainty never denies.
func TestPreTool_BranchGuard_ConfigGate_FailOpen(t *testing.T) {
	requireGit(t)
	nonGit := t.TempDir() // not a git repo → rev-parse exits non-zero
	t.Setenv(branchGuardExemptEnv, "")

	handler := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: cfgWithBranchGuard(true)},
		policy:     DefaultSecurityPolicy(),
		projectDir: nonGit,
	}
	input := &HookInput{
		SessionID:     "sess-bg-failopen",
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		AgentType:     "manager-develop",
		ToolInput:     json.RawMessage(`{"command": "git reset --hard HEAD~1"}`),
	}
	out, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle err = %v", err)
	}
	if decision := decisionOf(out); decision == DecisionDeny {
		t.Fatalf("Handle(enabled + non-git projectDir) = deny; want non-deny (AC-REQ-6c fail-open on git-context uncertainty)")
	}
}
