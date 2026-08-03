package hook

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestAC005_CommitGateOffAtHigherTiers (A11 / AC-STOPCHAIN-TRIM-005):
// at MOAI_AUTONOMY_TIER ∈ {automatic, fully-autonomous}, the synchronous
// vet+lint+test commit gate in the IsGitCommit branch is OFF — a git commit
// PreToolUse returns allow WITHOUT invoking the quality-gate toolchain. At
// semi-auto (and unset) the gate retains its current behavior.
//
// We assert the OFF behavior by setting cfg.Gate.Enabled=true (so the ONLY
// thing that can skip the gate is the tier branch) and confirming the commit
// is allowed at automatic/fully-autonomous. The tier branch is the
// K-6-ordered deny-before-tier structural guarantee: the deny path
// (checkBashCommand) runs unconditionally AFTER the commit gate regardless of
// tier, so a tier that turns the gate OFF never weakens a deny.
func TestAC005_CommitGateOffAtHigherTiers(t *testing.T) {
	cases := []struct {
		name    string
		tier    string
		setTier bool
	}{
		{"automatic → gate OFF", config.AutonomyTierAutomatic, true},
		{"fully-autonomous → gate OFF", config.AutonomyTierFullyAutonomous, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setTier {
				t.Setenv(config.EnvAutonomyTier, tc.tier)
			}
			cfg := config.NewDefaultConfig()
			cfg.Gate.Enabled = true // gate would run if the tier branch is missing
			handler := NewPreToolHandler(&mockConfigProvider{cfg: cfg}, DefaultSecurityPolicy())

			toolInput, _ := json.Marshal(map[string]string{
				"command": `git commit -m "feat: x"`,
			})
			input := &HookInput{
				SessionID:     "sess-ac005",
				HookEventName: "PreToolUse",
				ToolName:      "Bash",
				ToolInput:     json.RawMessage(toolInput),
			}
			got, err := handler.Handle(context.Background(), input)
			if err != nil {
				t.Fatalf("Handle error: %v", err)
			}
			if got == nil || got.HookSpecificOutput == nil {
				t.Fatal("expected non-nil output")
			}
			if got.HookSpecificOutput.PermissionDecision == DecisionDeny {
				t.Fatalf("AC-005: tier=%s denied a plain git commit — the commit gate ran when it should be OFF (tier branch missing?). reason: %s",
					tc.tier, got.HookSpecificOutput.PermissionDecisionReason)
			}
		})
	}
}

// TestAC006_DenyInvariantAtEveryTier (A11 cross-cutting / AC-STOPCHAIN-TRIM-006):
// the deny/ask denylist (destructive-pattern denylist — main push, force-push,
// reset --hard) SHALL bind at EVERY tier. This is the load-bearing K-6 safety
// regression guard: no tier value weakens or skips a deny/ask rule. The
// invariant is TIER-INVARIANCE — the decision at every tier MUST equal the
// decision at the baseline (semi-auto). A tier flipping deny→ask or ask→allow
// is a hard P0 FAIL (would let an unattended fully-autonomous session push to
// main / force-push / wipe a directory).
//
// Note: not every fixture below is a DENY in current policy — some are ASK
// (force-push to non-main, reset --hard). The invariant binds regardless: the
// decision is identical across all tier values. We baseline at semi-auto and
// assert no tier diverges.
func TestAC006_DenyInvariantAtEveryTier(t *testing.T) {
	fixtures := []struct {
		name    string
		command string
	}{
		{"git push --force origin main (DENY baseline)", `git push --force origin main`},
		{"git push --force origin feature/x (ASK baseline)", `git push --force origin feature/x`},
		{"git reset --hard (ASK baseline)", `git reset --hard origin/main`},
	}
	tiers := []struct {
		name string
		tier string
		set  bool
	}{
		{"unset (backward compat)", "", false},
		{"automatic", config.AutonomyTierAutomatic, true},
		{"fully-autonomous", config.AutonomyTierFullyAutonomous, true},
	}

	// withTier runs one fixture under one tier value and returns the decision.
	// Uses os.Setenv directly (NOT t.Setenv) so it can be called repeatedly
	// across tier values within the same parent test; restores the prior value.
	withTier := func(tier string, set bool, cmd string) string {
		prev, hadPrev := os.LookupEnv(config.EnvAutonomyTier)
		if set {
			_ = os.Setenv(config.EnvAutonomyTier, tier)
		} else {
			_ = os.Unsetenv(config.EnvAutonomyTier)
		}
		defer func() {
			if hadPrev {
				_ = os.Setenv(config.EnvAutonomyTier, prev)
			} else {
				_ = os.Unsetenv(config.EnvAutonomyTier)
			}
		}()
		cfg := config.NewDefaultConfig()
		cfg.Gate.Enabled = false
		handler := NewPreToolHandler(&mockConfigProvider{cfg: cfg}, DefaultSecurityPolicy())
		toolInput, _ := json.Marshal(map[string]string{"command": cmd})
		input := &HookInput{
			SessionID:     "sess-ac006",
			HookEventName: "PreToolUse",
			ToolName:      "Bash",
			ToolInput:     json.RawMessage(toolInput),
		}
		got, err := handler.Handle(context.Background(), input)
		if err != nil {
			t.Fatalf("Handle error: %v", err)
		}
		if got == nil || got.HookSpecificOutput == nil {
			t.Fatal("expected non-nil output")
		}
		return got.HookSpecificOutput.PermissionDecision
	}

	// Capture baseline at semi-auto (MUST itself be restrictive — deny/ask, never allow).
	baseline := map[string]string{}
	for _, fx := range fixtures {
		baseline[fx.name] = withTier(config.AutonomyTierSemiAuto, true, fx.command)
		if baseline[fx.name] == DecisionAllow {
			t.Fatalf("AC-006 baseline setup: %s returned allow at semi-auto — the denylist itself is broken (not a tier issue, but the invariant test is meaningless). Fix the denylist first.", fx.name)
		}
	}

	// Every other tier MUST match the baseline decision (tier-invariance).
	for _, tr := range tiers {
		for _, fx := range fixtures {
			got := withTier(tr.tier, tr.set, fx.command)
			if got != baseline[fx.name] {
				t.Errorf("AC-006 REGRESSION (P0): tier=%s command=%q → decision=%q, baseline(semi-auto)=%q. A tier value changed a destructive-pattern decision — deny/ask rules must be tier-INVARIANT (REQ-007).",
					tr.name, fx.command, got, baseline[fx.name])
			}
		}
	}
}
