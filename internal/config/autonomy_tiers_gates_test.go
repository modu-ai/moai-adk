package config

import (
	"testing"
)

// SPEC-AUTONOMY-TIERS-001 M2 — sandbox-proof gate + manager kill-switch.
// These tests exercise REQ-002 + REQ-005 gating: fully-autonomous is only
// allowed when a sandbox proof is present AND the kill-switch is off.

func TestSandboxProof_EnvMarker(t *testing.T) {
	// Two proof paths (OQ-1 resolved): env marker MOAI_SANDBOX_PROOF=<kind>
	// set by a container/VM launcher.
	t.Setenv(EnvSandboxProof, "docker")
	kind, ok := SandboxProofKind()
	if !ok {
		t.Fatalf("SandboxProofKind() ok=false, want true (env marker set)")
	}
	if kind != "docker" {
		t.Errorf("SandboxProofKind() kind=%q, want %q", kind, "docker")
	}
}

func TestSandboxProof_FlagExplicit(t *testing.T) {
	// The --sandbox-proof CLI flag surfaces as the same env marker (the CLI
	// helper sets the env). So an explicit non-empty marker is proof.
	t.Setenv(EnvSandboxProof, "gvisor")
	if _, ok := SandboxProofKind(); !ok {
		t.Errorf("SandboxProofKind() ok=false with explicit gvisor marker")
	}
}

func TestSandboxProof_Absent(t *testing.T) {
	// No env marker → no proof. fully-autonomous MUST be gated off (AC-002).
	t.Setenv(EnvSandboxProof, "")
	if _, ok := SandboxProofKind(); ok {
		t.Errorf("SandboxProofKind() ok=true with empty marker, want false")
	}
}

func TestIsBypassDisabled_ManagedConfig(t *testing.T) {
	// The manager kill-switch (disableBypassPermissionsMode) is injected by the
	// managed/enterprise config layer via an env seam.
	t.Setenv(EnvDisableBypassPermissionsMode, "1")
	if !IsBypassDisabled() {
		t.Errorf("IsBypassDisabled() = false with MOAI_DISABLE_BYPASS_PERMISSIONS_MODE=1, want true")
	}
}

func TestIsBypassDisabled_DefaultOff(t *testing.T) {
	t.Setenv(EnvDisableBypassPermissionsMode, "")
	if IsBypassDisabled() {
		t.Errorf("IsBypassDisabled() = true with empty env, want false (kill-switch off by default)")
	}
}

func TestEffectiveTierWithGates_FullyAutonomousAllowedWithProof(t *testing.T) {
	// fully-autonomous + sandbox proof present + kill-switch off → allowed.
	got, downgraded := EffectiveTierWithGates(AutonomyTierFullyAutonomous, true, false)
	if got != AutonomyTierFullyAutonomous || downgraded {
		t.Errorf("EffectiveTierWithGates(fully-autonomous, proof=true, kill=false) = (%q, %v), want (%q, false)",
			got, downgraded, AutonomyTierFullyAutonomous)
	}
}

func TestEffectiveTierWithGates_FullyAutonomousDowngradedNoProof(t *testing.T) {
	// fully-autonomous + NO proof + kill-switch off → downgrade to automatic.
	// AC-002: fully-autonomous disabled without sandbox proof.
	got, downgraded := EffectiveTierWithGates(AutonomyTierFullyAutonomous, false, false)
	if got != AutonomyTierAutomatic || !downgraded {
		t.Errorf("EffectiveTierWithGates(fully-autonomous, proof=false, kill=false) = (%q, %v), want (%q, true)",
			got, downgraded, AutonomyTierAutomatic)
	}
}

func TestEffectiveTierWithGates_KillSwitchWinsOverProof(t *testing.T) {
	// AC-005 trumps AC-002: kill-switch engaged → fully-autonomous downgraded
	// EVEN WHEN sandbox proof is present.
	got, downgraded := EffectiveTierWithGates(AutonomyTierFullyAutonomous, true, true)
	if got != AutonomyTierAutomatic || !downgraded {
		t.Errorf("EffectiveTierWithGates(fully-autonomous, proof=true, kill=true) = (%q, %v), want (%q, true) — kill-switch wins",
			got, downgraded, AutonomyTierAutomatic)
	}
}

func TestEffectiveTierWithGates_LowerTiersUnaffected(t *testing.T) {
	// The kill-switch does NOT affect semi-auto or automatic (REQ-005).
	for _, tier := range []string{AutonomyTierSemiAuto, AutonomyTierAutomatic} {
		got, downgraded := EffectiveTierWithGates(tier, false, true)
		if got != tier || downgraded {
			t.Errorf("EffectiveTierWithGates(%q, proof=false, kill=true) = (%q, %v), want (%q, false) — lower tiers unaffected",
				tier, got, downgraded, tier)
		}
	}
}

func TestEffectiveTierWithGates_DefaultModeRespectsGates(t *testing.T) {
	// The downgraded effective tier MUST map to the automatic defaultMode
	// ("auto"), NOT bypassPermissions — the renderer reads the EFFECTIVE tier.
	got, _ := EffectiveTierWithGates(AutonomyTierFullyAutonomous, false, false)
	if mode := TierDefaultMode(got); mode != "auto" {
		t.Errorf("downgraded tier defaultMode = %q, want %q", mode, "auto")
	}
}
