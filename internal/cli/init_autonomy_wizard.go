// Package cli — wizard→opts autonomy-tier apply wiring
// (SPEC-AUTONOMY-TIERS-001 M7, AC-001).
//
// init_autonomy_wizard.go holds applyAutonomyTierFromWizard: it applies the
// interactive page's autonomy-tier selection to opts, honouring (1) flag-over-
// wizard precedence and (2) the two gates that bind fully-autonomous (sandbox
// proof + kill-switch) via the EXISTING config.EffectiveTierWithGates core —
// gating is REUSED, not duplicated.
//
// @MX:SPEC: SPEC-AUTONOMY-TIERS-001
package cli

import (
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/core/project"
)

// applyAutonomyTierFromWizard applies the wizard's autonomy-tier selection to
// opts, honouring flag-over-wizard precedence (AC-001) and the two
// fully-autonomous gates (sandbox proof + kill-switch, AC-002/AC-005).
//
// Precedence:
//  1. When the --autonomy-tier flag was explicitly set (flagChanged && flagValue
//     != ""), it wins; the wizard selection is discarded. The flag path defers
//     gate enforcement to the downstream resolver (AutonomyTier()), matching the
//     non-interactive CLI surface.
//  2. Otherwise the wizard selection fills opts. fully-autonomous is gated via
//     EffectiveTierWithGates (downgraded to automatic when the proof is absent
//     OR the kill-switch is engaged). semi-auto and automatic pass through
//     unchanged (REQ-005 — the kill-switch never affects lower tiers).
//  3. An empty wizard selection leaves opts.AutonomyTier empty so the downstream
//     reader resolves semi-auto (AC-007 — unset ⇒ zero behavior delta).
func applyAutonomyTierFromWizard(flagChanged bool, flagValue string, res *wizard.WizardResult, opts *project.InitOptions) {
	if flagChanged && flagValue != "" {
		opts.AutonomyTier = flagValue
		return
	}
	if res.AutonomyTier == "" {
		return
	}
	_, proofOK := config.SandboxProofKind()
	killSwitch := config.IsBypassDisabled()
	effective, _ := config.EffectiveTierWithGates(res.AutonomyTier, proofOK, killSwitch)
	opts.AutonomyTier = effective
}
