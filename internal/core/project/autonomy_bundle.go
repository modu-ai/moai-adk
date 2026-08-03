package project

// SPEC-AUTONOMY-TIERS-001 M9 — init applies the tier bundle (end-to-end wiring).
//
// Gap 1 closure: applyAutonomyTierFromWizard (init.go) captures the effective
// tier into opts.AutonomyTier, but the init deployment path never CONSUMED it —
// no permission bundle (defaultMode + deny/ask) was written, so selecting a tier
// had NO effect on the deployed permissions. ApplyAutonomyTierBundle wires
// opts.AutonomyTier into the deployed settings at init time, REUSING the existing
// core (no gating/rendering logic is duplicated):
//
//   - config.ResolveEffectiveTier + config.EffectiveTierWithGates (M1/M2 gating)
//   - config.TierDefaultMode (M1 mode-token → knob mapping)
//   - config.SandboxProofKind + config.IsBypassDisabled (M2 env seams)
//   - config.AppendDowngradeAdvisory (M5 AC-005 sink)
//   - toolpolicy.RenderTierPermissions (M3 full-bundle renderer) when a
//     tool-policy.yaml is available, else toolpolicy.WriteUserDefaultMode (M9
//     USER-scope fallback that reuses the same RenderSettingsJSON codegen).
//
// REQ-007 invariant: semi-auto / unset → ZERO behavior delta (no file written).
// REQ-006 invariant: init MUST NOT default to fully-autonomous (it never passes
// fully-autonomous in unless the user explicitly selected it, and even then the
// gates downgrade it without a sandbox proof).

import (
	"fmt"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/config/toolpolicy"
)

// @MX:NOTE: [AUTO] SPEC-AUTONOMY-TIERS-001 M9 — init wiring consumes opts.AutonomyTier into the deployed settings
// @MX:REASON: gap-1 closure; the wizard captured the tier but the deployer never wrote the bundle, so tier selection was a no-op end-to-end

// ApplyAutonomyTierBundle wires a persisted autonomy-tier selection into the
// deployed settings at init time. It is the consumer that the init path was
// missing: the wizard/flag captured opts.AutonomyTier, but nothing translated
// that selection into a written permission bundle.
//
// projectRoot is the project being initialized. userSettingsPath is the
// USER-scope settings.json (typically ~/.claude/settings.json). projectSettingsPath
// is the PROJECT-scope settings.json (<project>/.claude/settings.json) that the
// template deployer just wrote. persistedTier is opts.AutonomyTier (already
// gate-applied on the wizard path; re-gated here so the flag path is covered).
//
// Behavior:
//
//   - semi-auto / unset → ZERO delta (REQ-007): no file is written or modified.
//   - automatic → defaultMode="auto" written to USER scope; deny/ask regenerated
//     in PROJECT scope when a tool-policy.yaml is available (otherwise the
//     template-shipped deny/ask stay untouched and only the USER defaultMode is
//     written).
//   - fully-autonomous → gated by sandbox proof + kill-switch (AC-002/AC-005);
//     a downgrade to automatic is advisory-logged and then behaves as automatic.
func ApplyAutonomyTierBundle(projectRoot, userSettingsPath, projectSettingsPath, persistedTier string) error {
	effective := config.ResolveEffectiveTier(persistedTier)
	if effective == config.AutonomyTierSemiAuto {
		// REQ-007: zero behavior delta. Do NOT touch either file.
		return nil
	}

	// Re-apply the two gates (AC-002 sandbox proof + AC-005 kill-switch). The
	// wizard path already applied them, but the --autonomy-tier flag path does
	// not, so gating here covers both entry points idempotently.
	_, proofOK := config.SandboxProofKind()
	gated, downgraded := config.EffectiveTierWithGates(
		effective, proofOK, config.IsBypassDisabled(),
	)
	if downgraded {
		// AC-005 sink: log the downgrade so it is observable post-hoc.
		_ = config.AppendDowngradeAdvisory(
			filepath.Join(projectRoot, ".moai", "logs", "autonomy-downgrade.log"),
			effective, gated, "init: no sandbox proof or kill-switch engaged",
		)
	}
	if gated == config.AutonomyTierSemiAuto {
		// A gate forced the selection all the way back to semi-auto → zero delta.
		return nil
	}

	defaultMode := config.TierDefaultMode(gated)

	// Full-bundle path: when the project ships a tool-policy.yaml (the
	// maintainer surface), reuse RenderTierPermissions so deny/ask are
	// regenerated from the doc across both scopes (AC-003/AC-004).
	if doc, err := toolpolicy.LoadFromProjectDir(projectRoot); err == nil && doc != nil {
		if _, rerr := toolpolicy.RenderTierPermissions(projectSettingsPath, userSettingsPath, defaultMode, doc); rerr != nil {
			return fmt.Errorf("render tier permissions: %w", rerr)
		}
		return nil
	}

	// Distributed default (no tool-policy.yaml): deny/ask already ship in the
	// PROJECT template, so only the USER-scope defaultMode is missing. Reuse
	// the same RenderSettingsJSON codegen via WriteUserDefaultMode.
	if err := toolpolicy.WriteUserDefaultMode(userSettingsPath, defaultMode); err != nil {
		return fmt.Errorf("write user defaultMode: %w", err)
	}
	return nil
}
