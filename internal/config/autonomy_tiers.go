package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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

// --- M2: sandbox-proof gate + manager kill-switch (REQ-002 / REQ-005) ---

// SandboxProofKind reads the MOAI_SANDBOX_PROOF env marker and returns the
// isolation-tech kind plus ok=true when a sandbox/container proof is present.
// An empty marker returns ("", false). The proof is the precondition the
// fully-autonomous tier requires: fully-autonomous MUST be gated off when no
// proof is present (AC-002). A git worktree is NOT a sandbox — only an
// OS-level container/VM proof qualifies.
//
// S3 hardening (SPEC-AUTONOMY-TIERS-001 amendment): the marker MUST match a
// recognized isolation-tech kind in SandboxProofKinds before ok=true is
// returned. An unknown kind (e.g. "1", "foo", "true") is rejected — this kills
// the trivial MOAI_SANDBOX_PROOF=<anything> spoof, which is S3's goal.
//
// The marker is an attestation: a launcher/user vouches that the session runs
// inside an isolation boundary of the named kind (one of Claude Code's official
// sandbox options — container, VM, or sandbox-runtime). moai-adk does NOT
// re-verify the boundary from inside: only Docker/Podman leave discoverable
// fingerprints, so a fingerprint cross-check would be a false gate for the
// other kinds (gvisor, firecracker, e2b, devcontainer, kata, sandbox-runtime)
// and would reject legitimate non-container hosts (e.g. CI runners). The
// allowlist is therefore the proof; the unverified-attestation path is
// surfaced as a stderr advisory on every platform so the acceptance is visible.
//
// Residual spoof surface (intentional, OQ-1/OQ-4 deferred): a user who knows
// the allowlist can self-attest any listed kind. The bar raised here is "known
// isolation-tech kind" over "any truthy string"; cryptographic attestation is
// the follow-up SPEC.
func SandboxProofKind() (kind string, ok bool) {
	raw := strings.TrimSpace(os.Getenv(EnvSandboxProof))
	if raw == "" {
		return "", false
	}
	if !isKnownSandboxKind(raw) {
		return "", false
	}
	// The allowlist above is the proof. moai-adk cannot reliably re-verify the
	// isolation boundary from inside (only Docker/Podman leave fingerprints),
	// so a fingerprint gate would false-reject most kinds and non-container
	// hosts. Surface the unverified attestation on every platform instead.
	fmt.Fprintf(os.Stderr, "sandbox proof kind %q accepted as attestation (host GOOS=%s); boundary not independently verified\n", raw, runtime.GOOS)
	return raw, true
}

// isKnownSandboxKind reports whether kind is in the SandboxProofKinds allowlist.
func isKnownSandboxKind(kind string) bool {
	for _, k := range SandboxProofKinds {
		if k == kind {
			return true
		}
	}
	return false
}

// IsBypassDisabled reads the MOAI_DISABLE_BYPASS_PERMISSIONS_MODE env seam
// for the Claude Code documented enterprise kill-switch
// (disableBypassPermissionsMode, SPEC-AUTONOMY-TIERS-001 REQ-005). Truthy
// values ("1", "true", "yes") return true. When the kill-switch is engaged,
// fully-autonomous is unselectable in every surface and an existing bypass
// session downgrades to automatic (AC-005).
func IsBypassDisabled() bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(EnvDisableBypassPermissionsMode)))
	switch raw {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// EffectiveTierWithGates resolves the effective tier after applying the two
// gates that bind the fully-autonomous tier:
//  1. sandbox proof — fully-autonomous requires a present proof (AC-002).
//  2. kill-switch — disableBypassPermissionsMode forbids fully-autonomous
//     even when a proof is present (AC-005 trumps AC-002).
//
// Returns (effectiveTier, downgraded). downgraded is true ONLY when the input
// was fully-autonomous and a gate forced a downgrade to automatic. Lower tiers
// (semi-auto, automatic) are NEVER affected by either gate (REQ-005).
func EffectiveTierWithGates(tier string, sandboxProofPresent, killSwitchActive bool) (effective string, downgraded bool) {
	if tier != AutonomyTierFullyAutonomous {
		return tier, false
	}
	// AC-005 trumps AC-002: kill-switch wins over proof.
	if killSwitchActive || !sandboxProofPresent {
		return AutonomyTierAutomatic, true
	}
	return AutonomyTierFullyAutonomous, false
}

// --- M5: web toggle availability (AC-002) + downgrade advisory (AC-005) ---

// TierToggleOption is one row of the moai web console's tier toggle: a tier
// value and whether the toggle renders it as selectable. Used by
// TierToggleOptions to describe the toggle's enablement state.
type TierToggleOption struct {
	Tier    string
	Enabled bool
}

// TierToggleOptions returns the 3-tier toggle rows for the moai web console,
// applying the two gates that bind fully-autonomous (AC-002 + AC-005):
//   - fully-autonomous is ENABLED only when a sandbox proof is present AND the
//     kill-switch is off. AC-005 trumps AC-002: kill-switch disables it even
//     with proof.
//   - semi-auto and automatic are ALWAYS enabled (REQ-005 — the kill-switch
//     does not affect lower tiers).
func TierToggleOptions(sandboxProofPresent, killSwitchActive bool) []TierToggleOption {
	fullyAutonomousEnabled := sandboxProofPresent && !killSwitchActive
	return []TierToggleOption{
		{Tier: AutonomyTierSemiAuto, Enabled: true},
		{Tier: AutonomyTierAutomatic, Enabled: true},
		{Tier: AutonomyTierFullyAutonomous, Enabled: fullyAutonomousEnabled},
	}
}

// AppendDowngradeAdvisory appends a downgrade advisory record to the
// autonomy-downgrade log (AC-005 sink). Called when a fully-autonomous
// selection (or an existing bypassPermissions session) is downgraded to
// automatic by a gate (no sandbox proof, or kill-switch engaged). The log is
// created (with parent dirs) if absent, and the record is APPENDED (not
// overwritten) so multiple downgrades accumulate. The record contains the
// "autonomy-downgrade" marker (the AC-005 grep target), the original + effective
// tiers, the reason, and an RFC3339 timestamp.
func AppendDowngradeAdvisory(logPath, originalTier, effectiveTier, reason string) error {
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return fmt.Errorf("create advisory log dir: %w", err)
	}
	line := fmt.Sprintf("autonomy-downgrade %s original=%s effective=%s reason=%q\n",
		time.Now().UTC().Format(time.RFC3339), originalTier, effectiveTier, reason)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open advisory log: %w", err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("write advisory log: %w", err)
	}
	return nil
}
