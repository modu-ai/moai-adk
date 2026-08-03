package config

import (
	"fmt"
	"strings"
)

// SPEC-AUTONOMY-TIERS-001 — autonomy 3-mode selection + behavior wiring.
//
// This file delivers the user-facing SELECTION surface (REQ-001 selector,
// REQ-007 backward-compat). It is a CONSUMER of the mode token owned by
// SPEC-STOPCHAIN-TRIM-001 (autonomy.go AutonomyTier() reader, envkeys.go
// EnvAutonomyTier, defaults.go 3-value enum) — it does NOT redefine them
// (AP-1).
//
// The distinction from the reader:
//   - AutonomyTier() (autonomy.go) is the fail-SAFE reader — unset / invalid
//     → semi-auto. Correct for hooks that MUST keep working on a bad value.
//   - ValidateAutonomyTierSelection (below) is the fail-LOUD selector —
//     invalid → error. Correct for the user-facing `--autonomy-tier` flag and
//     the moai web toggle, where a silent fallback would hide a typo (AP-5).

// @MX:ANCHOR: [AUTO] autonomy-tier selection surface — validation + resolution + defaultMode mapping
// @MX:REASON: SPEC-AUTONOMY-TIERS-001 REQ-001/REQ-007; the selector MUST fail-loud (distinct from the fail-safe reader) so a typo never silently enables fully-autonomous mode

// ValidateAutonomyTierSelection normalizes and validates a user-supplied
// autonomy-tier selection. It accepts the 3 canonical values
// {semi-auto, automatic, fully-autonomous} case-insensitively and
// whitespace-trimmed (the same normalization config.AutonomyTier() applies,
// but at the selection surface). An invalid value returns an error naming the
// valid values — the selector is fail-LOUD (REQ-001 / AP-5), distinct from the
// reader's fail-SAFE fallback.
func ValidateAutonomyTierSelection(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case AutonomyTierSemiAuto:
		return AutonomyTierSemiAuto, nil
	case AutonomyTierAutomatic:
		return AutonomyTierAutomatic, nil
	case AutonomyTierFullyAutonomous:
		return AutonomyTierFullyAutonomous, nil
	default:
		return "", fmt.Errorf(
			"invalid autonomy tier %q: must be one of %s, %s, %s",
			value, AutonomyTierSemiAuto, AutonomyTierAutomatic, AutonomyTierFullyAutonomous,
		)
	}
}

// ResolveEffectiveTier resolves a PERSISTED tier selection to the effective
// canonical tier. A persisted unset or whitespace-only selection resolves to
// semi-auto (REQ-007 / AC-007 — a session that does not opt in pays zero
// behavior delta). Non-empty values are returned verbatim after normalization;
// they are NOT re-validated here (the selector validated at write time, and the
// env-key wins per STOPCHAIN-TRIM's canonical-source rule).
func ResolveEffectiveTier(persistedTier string) string {
	normalized := strings.ToLower(strings.TrimSpace(persistedTier))
	if normalized == "" {
		return AutonomyTierSemiAuto
	}
	return normalized
}

// TierDefaultMode maps an autonomy tier to its Claude Code permissions
// defaultMode value (the mode-token → knob mapping from spec.md §C).
//
//   - semi-auto        → "default"     (today's behavior; full per-tool prompt)
//   - automatic        → "auto"        (per-tool auto-approval)
//   - fully-autonomous → "bypassPermissions" (all prompts skipped; sandbox-gated)
//
// deny/ask arrays are tier-INVARIANT (REQ-004) and are NOT part of this
// mapping — the renderer loads them identically for every tier. An unknown
// tier maps to "default" (the fail-safe never-silently-enable guarantee).
func TierDefaultMode(tier string) string {
	switch tier {
	case AutonomyTierAutomatic:
		return "auto"
	case AutonomyTierFullyAutonomous:
		return "bypassPermissions"
	default:
		// semi-auto AND any unknown value → "default". A bad tier never
		// silently enables bypassPermissions.
		return "default"
	}
}
