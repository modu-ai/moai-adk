package config

import (
	"os"
	"testing"
)

// TestAutonomyTierReader covers REQ-STOPCHAIN-TRIM-003 / AC-007:
// the MOAI_AUTONOMY_TIER env-key reader returns semi-auto on unset/empty
// (backward compat), round-trips the 3 canonical values, and falls back to
// semi-auto on any unrecognized value (fail-safe to the backward-compat default).
func TestAutonomyTierReader(t *testing.T) {
	cases := []struct {
		name string
		env  string
		set  bool
		want string
	}{
		{"unset → semi-auto (backward compat)", "", false, AutonomyTierSemiAuto},
		{"empty → semi-auto", "", true, AutonomyTierSemiAuto},
		{"semi-auto round-trips", AutonomyTierSemiAuto, true, AutonomyTierSemiAuto},
		{"automatic round-trips", AutonomyTierAutomatic, true, AutonomyTierAutomatic},
		{"fully-autonomous round-trips", AutonomyTierFullyAutonomous, true, AutonomyTierFullyAutonomous},
		{"whitespace-only → semi-auto", "   ", true, AutonomyTierSemiAuto},
		{"leading/trailing whitespace trimmed", "  fully-autonomous  ", true, AutonomyTierFullyAutonomous},
		{"invalid value → semi-auto (fail-safe)", "yolo-mode", true, AutonomyTierSemiAuto},
		{"case-insensitive upper", "FULLY-AUTONOMOUS", true, AutonomyTierFullyAutonomous},
		{"case-insensitive mixed", "AutoMatic", true, AutonomyTierAutomatic},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.set {
				t.Setenv(EnvAutonomyTier, tc.env)
			} else {
				// The kanban launcher exports MOAI_AUTONOMY_TIER into the
				// sessions it starts, so on a Kanban Mode developer machine
				// the ambient env would otherwise make the "unset" case read
				// a real tier. t.Setenv registers the restore; the unset
				// makes "not set" distinct from "set to empty".
				t.Setenv(EnvAutonomyTier, "")
				_ = os.Unsetenv(EnvAutonomyTier)
			}
			got := AutonomyTier()
			if got != tc.want {
				t.Fatalf("AutonomyTier() = %q, want %q (env=%q)", got, tc.want, tc.env)
			}
		})
	}
}

// TestAutonomyTierConstants asserts the canonical 3-value enum is recorded
// exactly once in defaults.go and matches the SPEC contract (semi-auto /
// automatic / fully-autonomous). Shell hooks read $MOAI_AUTONOMY_TIER verbatim,
// so any drift here silently breaks the shell↔Go contract.
func TestAutonomyTierConstants(t *testing.T) {
	if AutonomyTierSemiAuto != "semi-auto" {
		t.Errorf("AutonomyTierSemiAuto = %q, want %q", AutonomyTierSemiAuto, "semi-auto")
	}
	if AutonomyTierAutomatic != "automatic" {
		t.Errorf("AutonomyTierAutomatic = %q, want %q", AutonomyTierAutomatic, "automatic")
	}
	if AutonomyTierFullyAutonomous != "fully-autonomous" {
		t.Errorf("AutonomyTierFullyAutonomous = %q, want %q", AutonomyTierFullyAutonomous, "fully-autonomous")
	}
	if EnvAutonomyTier != "MOAI_AUTONOMY_TIER" {
		t.Errorf("EnvAutonomyTier = %q, want %q", EnvAutonomyTier, "MOAI_AUTONOMY_TIER")
	}
}

// TestIsHigherAutonomyTier encodes the tier ordering used by the mode-aware
// hooks (M4): the commit gate turns OFF at automatic AND fully-autonomous; the
// subagent lifecycle goes dormant ONLY at fully-autonomous. The predicate
// helpers make those branch points grep-able and self-documenting.
func TestIsHigherAutonomyTier(t *testing.T) {
	if IsAutonomyTierCommitGateOff(AutonomyTierSemiAuto) {
		t.Error("semi-auto must keep the commit gate ON")
	}
	if !IsAutonomyTierCommitGateOff(AutonomyTierAutomatic) {
		t.Error("automatic must turn the commit gate OFF")
	}
	if !IsAutonomyTierCommitGateOff(AutonomyTierFullyAutonomous) {
		t.Error("fully-autonomous must turn the commit gate OFF")
	}
	if IsAutonomyTierLifecycleDormant(AutonomyTierSemiAuto) {
		t.Error("semi-auto must keep lifecycle hooks active")
	}
	if IsAutonomyTierLifecycleDormant(AutonomyTierAutomatic) {
		t.Error("automatic must keep lifecycle hooks active")
	}
	if !IsAutonomyTierLifecycleDormant(AutonomyTierFullyAutonomous) {
		t.Error("fully-autonomous must make lifecycle hooks dormant")
	}
}
