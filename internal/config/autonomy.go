package config

import (
	"os"
	"strings"
)

// @MX:ANCHOR: [AUTO] AutonomyTier reader — single Go-side read point for MOAI_AUTONOMY_TIER
// @MX:REASON: SPEC-STOPCHAIN-TRIM-001 REQ-003 + REQ-007; shell hooks read $MOAI_AUTONOMY_TIER directly, Go hooks/pre_tool.go MUST read the same token via this helper to avoid drift in the tier-branch (K-6 deny-before-tier ordering)

// AutonomyTier reads the MOAI_AUTONOMY_TIER env-key (the single canonical
// source per SPEC-STOPCHAIN-TRIM-001 REQ-003) and returns the normalized tier
// value. Unset / empty / whitespace-only / unrecognized values all fall back to
// AutonomyTierSemiAuto — the backward-compat default (REQ-003 / AC-007), so a
// session that does not opt in pays zero behavior delta.
//
// The value is trimmed of surrounding whitespace and matched case-insensitively
// against the canonical literals (semi-auto / automatic / fully-autonomous);
// any other value resolves to semi-auto (fail-safe). Callers that need the
// branch-point predicates (commit gate on/off, lifecycle dormant) should use
// IsAutonomyTierCommitGateOff / IsAutonomyTierLifecycleDormant instead of
// string-comparing here.
func AutonomyTier() string {
	raw := strings.TrimSpace(os.Getenv(EnvAutonomyTier))
	switch strings.ToLower(raw) {
	case AutonomyTierSemiAuto:
		return AutonomyTierSemiAuto
	case AutonomyTierAutomatic:
		return AutonomyTierAutomatic
	case AutonomyTierFullyAutonomous:
		return AutonomyTierFullyAutonomous
	default:
		// Unset, empty, or unrecognized → semi-auto (REQ-003 backward compat;
		// fail-safe so a typo never silently enables fully-autonomous mode).
		return AutonomyTierSemiAuto
	}
}

// IsAutonomyTierCommitGateOff reports whether the synchronous vet+lint+test
// commit gate in internal/hook/pre_tool.go (the IsGitCommit branch) SHALL be
// OFF for the current tier. Per SPEC-STOPCHAIN-TRIM-001 REQ-005 / AC-005, the
// gate is OFF at automatic AND fully-autonomous; it stays ON at semi-auto.
// The destructive-pattern denylist is tier-INVARIANT (REQ-007) and is NOT
// governed by this predicate — see IsAutonomyTierLifecycleDormant for the
// symmetric lifecycle predicate.
func IsAutonomyTierCommitGateOff(tier string) bool {
	return tier == AutonomyTierAutomatic || tier == AutonomyTierFullyAutonomous
}

// IsAutonomyTierLifecycleDormant reports whether the SubagentStop /
// TeammateIdle / TaskCompleted lifecycle hooks SHALL be dormant (observe-only
// — audit-log written, no block / reject / AskUserQuestion translation) for
// the current tier. Per SPEC-STOPCHAIN-TRIM-001 REQ-006 / AC-006b, lifecycle
// hooks are dormant ONLY at fully-autonomous; they stay active at semi-auto
// and automatic.
func IsAutonomyTierLifecycleDormant(tier string) bool {
	return tier == AutonomyTierFullyAutonomous
}
