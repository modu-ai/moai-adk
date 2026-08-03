package config

import (
	"strings"
	"testing"
)

// SPEC-AUTONOMY-TIERS-001 M1 — tier selection core.
// These tests exercise the user-facing SELECTION surface: validation,
// effective-tier resolution, and the tier → defaultMode mapping that the
// renderer (M3) consumes. They are distinct from the reader
// (AutonomyTier() in autonomy.go, fail-safe → semi-auto) because the
// selector MUST fail-LOUD on invalid input (REQ-001 / AP-5).

func TestValidateAutonomyTierSelection_ClosedSet(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"semi-auto", AutonomyTierSemiAuto},
		{"automatic", AutonomyTierAutomatic},
		{"fully-autonomous", AutonomyTierFullyAutonomous},
		// case-insensitive + whitespace-trimmed (REQ-001: same normalization as
		// config.AutonomyTier(), applied at the selector surface).
		{"SEMI-AUTO", AutonomyTierSemiAuto},
		{"  automatic  ", AutonomyTierAutomatic},
		{"Fully-Autonomous", AutonomyTierFullyAutonomous},
	}
	for _, c := range cases {
		got, err := ValidateAutonomyTierSelection(c.in)
		if err != nil {
			t.Errorf("ValidateAutonomyTierSelection(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ValidateAutonomyTierSelection(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestValidateAutonomyTierSelection_RejectsInvalid_FailLoud(t *testing.T) {
	// REQ-001 / AP-5: the selector is the user-facing surface, so an invalid
	// value MUST produce an error (NOT silent semi-auto fallback, which is the
	// reader's contract). The error MUST name the valid values so a typo is
	// visible.
	invalid := []string{"bogus", "full-autonomous", "auto", "semiauto", "", "none"}
	for _, in := range invalid {
		got, err := ValidateAutonomyTierSelection(in)
		if err == nil {
			t.Errorf("ValidateAutonomyTierSelection(%q): expected error, got nil (returned %q)", in, got)
			continue
		}
		// The error message MUST name all three canonical values so the user
		// can correct the typo.
		for _, v := range []string{AutonomyTierSemiAuto, AutonomyTierAutomatic, AutonomyTierFullyAutonomous} {
			if !strings.Contains(err.Error(), v) {
				t.Errorf("ValidateAutonomyTierSelection(%q) error %q does not name valid value %q", in, err.Error(), v)
			}
		}
	}
}

func TestResolveEffectiveTier_UnsetSemiAuto(t *testing.T) {
	// AC-007: a session that does not opt in pays zero behavior delta —
	// persisted unset/empty resolves to semi-auto.
	cases := []string{"", "   "}
	for _, in := range cases {
		if got := ResolveEffectiveTier(in); got != AutonomyTierSemiAuto {
			t.Errorf("ResolveEffectiveTier(%q) = %q, want %q", in, got, AutonomyTierSemiAuto)
		}
	}
}

func TestResolveEffectiveTier_Preserved(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{AutonomyTierAutomatic, AutonomyTierAutomatic},
		{AutonomyTierFullyAutonomous, AutonomyTierFullyAutonomous},
		{AutonomyTierSemiAuto, AutonomyTierSemiAuto},
	}
	for _, c := range cases {
		if got := ResolveEffectiveTier(c.in); got != c.want {
			t.Errorf("ResolveEffectiveTier(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTierDefaultMode_Mapping(t *testing.T) {
	// The mode-token → knob mapping (spec.md §C). Each tier maps to a single
	// canonical defaultMode value; deny/ask are tier-INVARIANT (REQ-004) and
	// are NOT part of this mapping.
	cases := []struct {
		tier string
		want string
	}{
		{AutonomyTierSemiAuto, "default"},
		{AutonomyTierAutomatic, "auto"},
		{AutonomyTierFullyAutonomous, "bypassPermissions"},
	}
	for _, c := range cases {
		if got := TierDefaultMode(c.tier); got != c.want {
			t.Errorf("TierDefaultMode(%q) = %q, want %q", c.tier, got, c.want)
		}
	}
}

func TestTierDefaultMode_InvalidTierSemiAutoDefault(t *testing.T) {
	// Defensive: an unknown tier maps to the semi-auto defaultMode, not
	// bypassPermissions — a never-silently-enable guarantee.
	if got := TierDefaultMode("bogus"); got != "default" {
		t.Errorf("TierDefaultMode(%q) = %q, want %q (fail-safe)", "bogus", got, "default")
	}
}
