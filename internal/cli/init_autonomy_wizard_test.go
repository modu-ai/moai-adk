package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/core/project"
)

// SPEC-AUTONOMY-TIERS-001 M7 — wizard→opts apply wiring with gating (AC-001).
// applyAutonomyTierFromWizard (defined in init.go) is the unit under test: it
// applies the interactive page's selection to opts, honouring (1) flag-over-
// wizard precedence and (2) the two gates that bind fully-autonomous (sandbox
// proof + kill-switch) via the EXISTING config.EffectiveTierWithGates core —
// gating is REUSED, not duplicated.

func TestApplyAutonomyTierFromWizard_FlagWinsOverWizard(t *testing.T) {
	opts := &project.InitOptions{}
	res := &wizard.WizardResult{AutonomyTier: config.AutonomyTierAutomatic}
	applyAutonomyTierFromWizard(true, config.AutonomyTierFullyAutonomous, res, opts)
	if opts.AutonomyTier != config.AutonomyTierFullyAutonomous {
		t.Errorf("flag should win: want %s, got %s", config.AutonomyTierFullyAutonomous, opts.AutonomyTier)
	}
}

func TestApplyAutonomyTierFromWizard_WizardFillsWhenFlagAbsent(t *testing.T) {
	// No proof, no kill-switch: automatic passes through unchanged.
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	opts := &project.InitOptions{}
	res := &wizard.WizardResult{AutonomyTier: config.AutonomyTierAutomatic}
	applyAutonomyTierFromWizard(false, "", res, opts)
	if opts.AutonomyTier != config.AutonomyTierAutomatic {
		t.Errorf("wizard automatic should pass through: got %s", opts.AutonomyTier)
	}
}

func TestApplyAutonomyTierFromWizard_FullyAutonomousDowngradedWithoutProof(t *testing.T) {
	// AC-002/AC-005 gating: fully-autonomous selected but no sandbox proof →
	// downgraded to automatic. REUSES EffectiveTierWithGates.
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	opts := &project.InitOptions{}
	res := &wizard.WizardResult{AutonomyTier: config.AutonomyTierFullyAutonomous}
	applyAutonomyTierFromWizard(false, "", res, opts)
	if opts.AutonomyTier != config.AutonomyTierAutomatic {
		t.Errorf("fully-autonomous without proof should downgrade to automatic: got %s", opts.AutonomyTier)
	}
}

func TestApplyAutonomyTierFromWizard_KillSwitchDowngradesEvenWithProof(t *testing.T) {
	// AC-005 trumps AC-002: kill-switch wins over proof.
	t.Setenv("MOAI_SANDBOX_PROOF", "docker")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "true")
	opts := &project.InitOptions{}
	res := &wizard.WizardResult{AutonomyTier: config.AutonomyTierFullyAutonomous}
	applyAutonomyTierFromWizard(false, "", res, opts)
	if opts.AutonomyTier != config.AutonomyTierAutomatic {
		t.Errorf("kill-switch should downgrade fully-autonomous even with proof: got %s", opts.AutonomyTier)
	}
}

func TestApplyAutonomyTierFromWizard_FullyAutonomousPassesWithProof(t *testing.T) {
	t.Setenv("MOAI_SANDBOX_PROOF", "docker")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	opts := &project.InitOptions{}
	res := &wizard.WizardResult{AutonomyTier: config.AutonomyTierFullyAutonomous}
	applyAutonomyTierFromWizard(false, "", res, opts)
	if opts.AutonomyTier != config.AutonomyTierFullyAutonomous {
		t.Errorf("fully-autonomous with proof + no kill-switch should pass: got %s", opts.AutonomyTier)
	}
}

func TestApplyAutonomyTierFromWizard_EmptyWizardLeavesOptsEmpty(t *testing.T) {
	// AC-007: unset → semi-auto (zero behavior delta). opts.AutonomyTier empty
	// means the downstream reader resolves to semi-auto.
	opts := &project.InitOptions{}
	res := &wizard.WizardResult{}
	applyAutonomyTierFromWizard(false, "", res, opts)
	if opts.AutonomyTier != "" {
		t.Errorf("empty wizard selection should leave opts empty (semi-auto downstream): got %q", opts.AutonomyTier)
	}
	// Sanity: the closed-set values are exactly the 3 canonical tiers.
	for _, tier := range []string{config.AutonomyTierSemiAuto, config.AutonomyTierAutomatic, config.AutonomyTierFullyAutonomous} {
		if strings.TrimSpace(tier) == "" {
			t.Errorf("canonical tier must be non-empty: %q", tier)
		}
	}
}
