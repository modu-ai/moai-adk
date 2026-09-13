package project

// SPEC-AUTONOMY-TIERS-001 M9 — init applies the tier bundle (end-to-end wiring).
//
// Gap 1 closure: applyAutonomyTierFromWizard (init_autonomy_wizard.go, called
// from runInit in init.go — wired by SPEC-INIT-WIZARD-REPAIR-001) captures the
// effective tier into opts.AutonomyTier. The init deployment path originally
// never CONSUMED it — no permission bundle (defaultMode + deny/ask) was
// written, so selecting a tier had NO effect on the deployed permissions.
// ApplyAutonomyTierBundle wires opts.AutonomyTier into the deployed settings at
// init time (runInit invokes it right after the initializer returns), REUSING
// the existing core (no gating/rendering logic is duplicated):
//
//   - config.ResolveEffectiveTier + config.EffectiveTierWithGates (M1/M2 gating)
//   - config.TierDefaultMode (M1 mode-token → knob mapping)
//   - config.SandboxProofKind + config.IsBypassDisabled (M2 env seams)
//   - config.AppendDowngradeAdvisory (M5 AC-005 sink)
//   - toolpolicy.RenderTierPermissions (M3 full-bundle renderer) when a
//     tool-policy.yaml is available, else toolpolicy.WriteUserDefaultMode (M9
//     USER-scope fallback that reuses the same RenderSettingsJSON codegen).
//
// REQ-004 invariant (SPEC-AUT-PERMMODES-001, re-scoped from SPEC-AUTONOMY-TIERS-001
// REQ-007): semi-auto / unset writes ONLY the USER-scope defaultMode="acceptEdits"
// record — the PROJECT file and every other deployed file stay byte-identical.
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
// Behavior (SPEC-AUT-PERMMODES-001):
//
//   - semi-auto / unset → BOUNDED delta (REQ-004, re-scoped from
//     SPEC-AUTONOMY-TIERS-001 REQ-007): ONLY the USER-scope
//     defaultMode="acceptEdits" record is written. The PROJECT-scope
//     allow/ask/deny arrays, the deployed template files, and every other
//     file stay byte-identical — the sanctioned delta is exactly one JSON key
//     in exactly one file.
//   - automatic → defaultMode="auto" written to USER scope; deny/ask regenerated
//     in PROJECT scope when a tool-policy.yaml is available (otherwise the
//     template-shipped deny/ask stay untouched and only the USER defaultMode is
//     written).
//   - fully-autonomous → gated by sandbox proof + kill-switch (AC-002/AC-005);
//     a downgrade to automatic is advisory-logged and then behaves as automatic.
//
// The USER-scope write is what makes all three values effective: per the
// official Claude Code permission-modes docs, "auto" and "bypassPermissions"
// set in PROJECT-scope settings files are silently ignored — only
// acceptEdits applies from every scope, and auto/bypass take effect from USER
// scope (M1 scope finding, progress.md §E.2).
func ApplyAutonomyTierBundle(projectRoot, userSettingsPath, projectSettingsPath, persistedTier string) error {
	effective := config.ResolveEffectiveTier(persistedTier)
	if effective == config.AutonomyTierSemiAuto {
		// REQ-004 (re-scoped): bounded delta — the USER-scope acceptEdits
		// record is the ONLY sanctioned write. The gates never bind semi-auto
		// (EffectiveTierWithGates passes lower tiers through) and the
		// PROJECT-scope deny/ask are tier-invariant, so neither is touched.
		if err := toolpolicy.WriteUserDefaultMode(userSettingsPath, config.TierDefaultMode(effective)); err != nil {
			return fmt.Errorf("write user defaultMode: %w", err)
		}
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
