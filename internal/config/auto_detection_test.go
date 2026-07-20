// Package config — auto_detection value-range validation tests (SPEC-HARNESS-EVOLVE-003 M2).
// REQ-HEV3-002: validate proposed auto_detection threshold values against [lower, upper] bounds.
package config

import (
	"testing"
)

// TestValidateAutoDetectionConditions_InRange verifies in-range thresholds pass.
func TestValidateAutoDetectionConditions_InRange(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		conditions []string
	}{
		{"typical minimal", []string{"file_count <= 3", "single_domain"}},
		{"typical thorough", []string{"file_count > 10", "multi_domain"}},
		{"no thresholds", []string{"single_domain", "security_keywords"}},
		{"empty", []string{}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := ValidateAutoDetectionConditions(tc.conditions); err != nil {
				t.Errorf("ValidateAutoDetectionConditions(%v) = %v, want nil", tc.conditions, err)
			}
		})
	}
}

// TestValidateAutoDetectionConditions_OutOfRange verifies AC-HEV3-002a/002b:
// out-of-range thresholds are rejected with ErrAutoDetectionOutOfRange.
func TestValidateAutoDetectionConditions_OutOfRange(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		conditions []string
	}{
		{"AC-002a below lower bound", []string{"file_count <= 0"}},
		{"AC-002b above upper bound", []string{"file_count <= 99999"}},
		{"negative", []string{"file_count <= -5"}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateAutoDetectionConditions(tc.conditions)
			if err == nil {
				t.Fatalf("ValidateAutoDetectionConditions(%v) = nil, want ErrAutoDetectionOutOfRange", tc.conditions)
			}
			if !IsErrAutoDetectionOutOfRange(err) {
				t.Errorf("ValidateAutoDetectionConditions(%v) = %v, want ErrAutoDetectionOutOfRange", tc.conditions, err)
			}
		})
	}
}

// TestAutoDetectionBounds_Hardcoded verifies H-2 resolved: bounds are hardcoded
// in the Go struct, not read from a schema file.
func TestAutoDetectionBounds_Hardcoded(t *testing.T) {
	t.Parallel()

	bounds, ok := AutoDetectionBounds["file_count"]
	if !ok {
		t.Fatal("AutoDetectionBounds missing file_count entry (H-2: hardcoded bounds)")
	}
	if bounds[0] != 1 || bounds[1] != 1000 {
		t.Errorf("file_count bounds = [%d, %d], want [1, 1000]", bounds[0], bounds[1])
	}
}
