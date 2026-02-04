package security

import (
	"testing"
)

// TestScanResult_CountBySeverity verifies severity counting.
func TestScanResult_CountBySeverity(t *testing.T) {
	t.Run("counts empty findings", func(t *testing.T) {
		result := &ScanResult{Findings: []Finding{}}
		errors, warnings, infos := result.CountBySeverity()

		if errors != 0 || warnings != 0 || infos != 0 {
			t.Errorf("expected 0, 0, 0, got %d, %d, %d", errors, warnings, infos)
		}
	})

	t.Run("counts mixed severities", func(t *testing.T) {
		result := &ScanResult{
			Findings: []Finding{
				{Severity: SeverityError},
				{Severity: SeverityError},
				{Severity: SeverityWarning},
				{Severity: SeverityWarning},
				{Severity: SeverityWarning},
				{Severity: SeverityInfo},
				{Severity: SeverityHint},
			},
		}
		errors, warnings, infos := result.CountBySeverity()

		if errors != 2 {
			t.Errorf("expected 2 errors, got %d", errors)
		}
		if warnings != 3 {
			t.Errorf("expected 3 warnings, got %d", warnings)
		}
		if infos != 2 { // info + hint
			t.Errorf("expected 2 infos, got %d", infos)
		}
	})
}

// TestScanResult_HasErrors verifies error detection.
func TestScanResult_HasErrors(t *testing.T) {
	t.Run("returns false for no errors", func(t *testing.T) {
		result := &ScanResult{ErrorCount: 0}
		if result.HasErrors() {
			t.Error("expected false for no errors")
		}
	})

	t.Run("returns true for errors", func(t *testing.T) {
		result := &ScanResult{ErrorCount: 1}
		if !result.HasErrors() {
			t.Error("expected true for errors")
		}
	})
}

// TestSeverityConstants verifies severity string values.
func TestSeverityConstants(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityError, "error"},
		{SeverityWarning, "warning"},
		{SeverityInfo, "info"},
		{SeverityHint, "hint"},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			if string(tt.severity) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.severity)
			}
		})
	}
}

// TestConstants verifies package constants.
func TestConstants(t *testing.T) {
	t.Run("DefaultScanTimeout is 30 seconds", func(t *testing.T) {
		if DefaultScanTimeout.Seconds() != 30 {
			t.Errorf("expected 30s, got %v", DefaultScanTimeout)
		}
	})

	t.Run("MaxFindingsToReport is 10", func(t *testing.T) {
		if MaxFindingsToReport != 10 {
			t.Errorf("expected 10, got %d", MaxFindingsToReport)
		}
	})
}
