package cli

import (
	"strings"
	"testing"
)

// SPEC-AUTONOMY-TIERS-001 M4 — moai init --autonomy-tier flag (AC-001).
// The flag validates the 3-value closed set fail-loud (REQ-001 / AP-5),
// distinct from the fail-safe reader. AC-006: the flag help offers the 3 tiers
// and does not pre-pick fully-autonomous.

func TestInitCmd_HasAutonomyTierFlag(t *testing.T) {
	flag := initCmd.Flags().Lookup("autonomy-tier")
	if flag == nil {
		t.Fatal("--autonomy-tier flag not registered on initCmd")
	}
	// AC-006: the flag help MUST name all 3 tiers so the selector OFFERS them.
	help := flag.Usage
	for _, v := range []string{"semi-auto", "automatic", "fully-autonomous"} {
		if !strings.Contains(help, v) {
			t.Errorf("--autonomy-tier help %q does not name tier %q", help, v)
		}
	}
}

func TestValidateInitFlags_ValidAutonomyTier(t *testing.T) {
	for _, tier := range []string{"semi-auto", "automatic", "fully-autonomous", "SEMI-AUTO", "  automatic  "} {
		t.Run(tier, func(t *testing.T) {
			if err := initCmd.Flags().Set("autonomy-tier", tier); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = initCmd.Flags().Set("autonomy-tier", "") }()
			if err := validateInitFlags(initCmd, []string{}); err != nil {
				t.Errorf("validateInitFlags with autonomy-tier=%q should pass, got: %v", tier, err)
			}
		})
	}
}

func TestValidateInitFlags_InvalidAutonomyTier(t *testing.T) {
	for _, tier := range []string{"bogus", "full-autonomous", "auto", ""} {
		if tier == "" {
			continue // empty is valid (defaults to semi-auto downstream)
		}
		t.Run(tier, func(t *testing.T) {
			if err := initCmd.Flags().Set("autonomy-tier", tier); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = initCmd.Flags().Set("autonomy-tier", "") }()
			err := validateInitFlags(initCmd, []string{})
			if err == nil {
				t.Fatalf("validateInitFlags with autonomy-tier=%q should error, got nil", tier)
			}
			if !strings.Contains(err.Error(), "invalid --autonomy-tier") {
				t.Errorf("error should mention 'invalid --autonomy-tier', got: %v", err)
			}
			// AP-5: the error MUST name the valid values so a typo is visible.
			for _, v := range []string{"semi-auto", "automatic", "fully-autonomous"} {
				if !strings.Contains(err.Error(), v) {
					t.Errorf("error %q does not name valid tier %q", err.Error(), v)
				}
			}
		})
	}
}

func TestValidateInitFlags_AutonomyTierEmptyIsValid(t *testing.T) {
	// AC-007: empty/unset → semi-auto (zero behavior delta). Validation MUST
	// NOT error on empty (the default).
	if err := initCmd.Flags().Set("autonomy-tier", ""); err != nil {
		t.Fatal(err)
	}
	if err := validateInitFlags(initCmd, []string{}); err != nil {
		t.Errorf("validateInitFlags with empty autonomy-tier should pass (defaults to semi-auto), got: %v", err)
	}
}
