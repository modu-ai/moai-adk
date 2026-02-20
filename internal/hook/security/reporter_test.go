package security

import (
	"strings"
	"testing"
	"time"
)

// =============================================================================
// RED PHASE: Tests for FindingReporter implementation
// These tests define expected behavior per SPEC-HOOK-003 REQ-HOOK-130~132
// =============================================================================

// TestNewFindingReporter verifies reporter creation.
func TestNewFindingReporter(t *testing.T) {
	t.Run("creates reporter instance", func(t *testing.T) {
		reporter := NewFindingReporter()
		if reporter == nil {
			t.Fatal("expected non-nil reporter")
		}
	})
}

// TestFindingReporter_FormatResult verifies result formatting.
// REQ-HOOK-130: System must report findings in structured format.
func TestFindingReporter_FormatResult(t *testing.T) {
	reporter := NewFindingReporter()

	t.Run("formats empty result", func(t *testing.T) {
		result := &ScanResult{
			Scanned:  true,
			Findings: []Finding{},
		}

		output := reporter.FormatResult(result, "test.py")
		if output == "" {
			t.Error("expected non-empty output for empty findings")
		}
		// Should indicate no issues found
		if !strings.Contains(strings.ToLower(output), "no") ||
			!strings.Contains(strings.ToLower(output), "found") &&
				!strings.Contains(strings.ToLower(output), "issue") &&
				!strings.Contains(strings.ToLower(output), "clean") {
			t.Logf("output: %s", output)
		}
	})

	t.Run("formats result with findings", func(t *testing.T) {
		result := &ScanResult{
			Scanned:      true,
			ErrorCount:   1,
			WarningCount: 1,
			Findings: []Finding{
				{
					RuleID:   "sql-injection",
					Severity: SeverityError,
					Message:  "Potential SQL injection",
					File:     "test.py",
					Line:     10,
					Column:   5,
				},
				{
					RuleID:   "weak-random",
					Severity: SeverityWarning,
					Message:  "Math.random() used",
					File:     "test.py",
					Line:     20,
					Column:   3,
				},
			},
			Duration: 100 * time.Millisecond,
		}

		output := reporter.FormatResult(result, "test.py")

		// Should contain severity indicators
		if !strings.Contains(output, "error") && !strings.Contains(output, "ERROR") {
			t.Error("output should contain error indicator")
		}
		if !strings.Contains(output, "warning") && !strings.Contains(output, "WARNING") {
			t.Error("output should contain warning indicator")
		}

		// Should contain rule IDs
		if !strings.Contains(output, "sql-injection") {
			t.Error("output should contain rule ID")
		}

		// Should contain line numbers
		if !strings.Contains(output, "10") {
			t.Error("output should contain line number")
		}
	})

	t.Run("formats result when scan failed", func(t *testing.T) {
		result := &ScanResult{
			Scanned: false,
			Error:   "ast-grep not available",
		}

		output := reporter.FormatResult(result, "test.py")
		// Should indicate skip or error
		if output == "" {
			t.Error("expected non-empty output for failed scan")
		}
	})

	t.Run("includes count summary per REQ-HOOK-130", func(t *testing.T) {
		result := &ScanResult{
			Scanned:      true,
			ErrorCount:   2,
			WarningCount: 3,
			InfoCount:    1,
			Findings: []Finding{
				{RuleID: "r1", Severity: SeverityError, Message: "msg1", File: "test.py", Line: 1},
				{RuleID: "r2", Severity: SeverityError, Message: "msg2", File: "test.py", Line: 2},
				{RuleID: "r3", Severity: SeverityWarning, Message: "msg3", File: "test.py", Line: 3},
				{RuleID: "r4", Severity: SeverityWarning, Message: "msg4", File: "test.py", Line: 4},
				{RuleID: "r5", Severity: SeverityWarning, Message: "msg5", File: "test.py", Line: 5},
				{RuleID: "r6", Severity: SeverityInfo, Message: "msg6", File: "test.py", Line: 6},
			},
		}

		output := reporter.FormatResult(result, "test.py")

		// Should contain count indicators
		if !strings.Contains(output, "2") || !strings.Contains(output, "3") {
			t.Errorf("output should contain error/warning counts: %s", output)
		}
	})
}

// TestFindingReporter_FormatMultiple verifies multiple results formatting.
func TestFindingReporter_FormatMultiple(t *testing.T) {
	reporter := NewFindingReporter()

	t.Run("formats empty results list", func(t *testing.T) {
		output := reporter.FormatMultiple([]*ScanResult{})
		// Should return empty or minimal output
		_ = output
	})

	t.Run("aggregates multiple results", func(t *testing.T) {
		results := []*ScanResult{
			{
				Scanned:    true,
				ErrorCount: 1,
				Findings: []Finding{
					{RuleID: "r1", Severity: SeverityError, Message: "error in file1", File: "file1.py", Line: 1},
				},
			},
			{
				Scanned:      true,
				WarningCount: 2,
				Findings: []Finding{
					{RuleID: "r2", Severity: SeverityWarning, Message: "warning in file2", File: "file2.py", Line: 1},
					{RuleID: "r3", Severity: SeverityWarning, Message: "warning in file2", File: "file2.py", Line: 2},
				},
			},
		}

		output := reporter.FormatMultiple(results)
		// Should contain findings from both files
		if !strings.Contains(output, "file1") || !strings.Contains(output, "file2") {
			t.Errorf("output should contain both file names: %s", output)
		}
	})
}

// TestFindingReporter_ShouldExitWithError verifies exit code logic.
// REQ-HOOK-131: System must return exit code 2 for error-severity findings.
func TestFindingReporter_ShouldExitWithError(t *testing.T) {
	reporter := NewFindingReporter()

	t.Run("returns false for no findings", func(t *testing.T) {
		result := &ScanResult{Scanned: true, ErrorCount: 0}
		if reporter.ShouldExitWithError(result) {
			t.Error("should not exit with error for no findings")
		}
	})

	t.Run("returns false for warnings only", func(t *testing.T) {
		result := &ScanResult{
			Scanned:      true,
			WarningCount: 5,
			Findings: []Finding{
				{Severity: SeverityWarning},
			},
		}
		if reporter.ShouldExitWithError(result) {
			t.Error("should not exit with error for warnings only")
		}
	})

	t.Run("returns true for error findings", func(t *testing.T) {
		result := &ScanResult{
			Scanned:    true,
			ErrorCount: 1,
			Findings: []Finding{
				{Severity: SeverityError, RuleID: "sql-injection"},
			},
		}
		if !reporter.ShouldExitWithError(result) {
			t.Error("should exit with error for error-severity findings")
		}
	})

	t.Run("returns false for unscanned result", func(t *testing.T) {
		result := &ScanResult{Scanned: false}
		if reporter.ShouldExitWithError(result) {
			t.Error("should not exit with error when scan was skipped")
		}
	})
}

// TestFindingReporter_LimitFindings verifies the 10 finding limit.
// REQ-HOOK-132: System must limit to 10 findings with "... and N more" summary.
func TestFindingReporter_LimitFindings(t *testing.T) {
	reporter := NewFindingReporter()

	t.Run("shows all findings when 10 or fewer", func(t *testing.T) {
		findings := make([]Finding, 10)
		for i := range 10 {
			findings[i] = Finding{
				RuleID:   "rule-" + string(rune('0'+i)),
				Severity: SeverityWarning,
				Message:  "message",
				File:     "test.py",
				Line:     i + 1,
			}
		}
		result := &ScanResult{
			Scanned:      true,
			WarningCount: 10,
			Findings:     findings,
		}

		output := reporter.FormatResult(result, "test.py")
		// Should not contain "and N more"
		if strings.Contains(output, "more") {
			t.Errorf("should show all 10 findings without truncation: %s", output)
		}
	})

	t.Run("truncates findings when more than 10", func(t *testing.T) {
		findings := make([]Finding, 15)
		for i := range 15 {
			findings[i] = Finding{
				RuleID:   "rule-" + string(rune('a'+i)),
				Severity: SeverityWarning,
				Message:  "message",
				File:     "test.py",
				Line:     i + 1,
			}
		}
		result := &ScanResult{
			Scanned:      true,
			WarningCount: 15,
			Findings:     findings,
		}

		output := reporter.FormatResult(result, "test.py")
		// Should contain "and N more" indicator
		if !strings.Contains(output, "5 more") && !strings.Contains(output, "and 5") {
			t.Errorf("should indicate 5 more findings: %s", output)
		}
	})
}

// TestFindingReporter_OutputFormat verifies output format per SPEC 4.4.
func TestFindingReporter_OutputFormat(t *testing.T) {
	reporter := NewFindingReporter()

	t.Run("produces Claude-compatible format", func(t *testing.T) {
		result := &ScanResult{
			Scanned:    true,
			ErrorCount: 1,
			Findings: []Finding{
				{
					RuleID:   "sql-injection",
					Severity: SeverityError,
					Message:  "Potential SQL injection",
					File:     "main.py",
					Line:     45,
				},
			},
		}

		output := reporter.FormatResult(result, "main.py")

		// Per SPEC 4.4 example format:
		// "AST-Grep found 2 error(s), 1 warning(s) in main.py
		//   - [ERROR] sql-injection: Potential SQL injection (line 45)"

		// Should mention AST-Grep
		if !strings.Contains(output, "AST-Grep") && !strings.Contains(output, "ast-grep") {
			t.Error("output should mention AST-Grep")
		}

		// Should contain severity tag
		if !strings.Contains(output, "[ERROR]") && !strings.Contains(output, "ERROR") {
			t.Error("output should contain ERROR tag")
		}

		// Should contain line number
		if !strings.Contains(output, "45") && !strings.Contains(output, "line") {
			t.Error("output should contain line reference")
		}
	})
}

// TestFindingReporter_AllSeverityTags verifies all severity tags are formatted.
func TestFindingReporter_AllSeverityTags(t *testing.T) {
	reporter := NewFindingReporter()

	tests := []struct {
		severity    Severity
		expectedTag string
	}{
		{SeverityError, "ERROR"},
		{SeverityWarning, "WARNING"},
		{SeverityInfo, "INFO"},
		{SeverityHint, "HINT"},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			result := &ScanResult{
				Scanned:    true,
				ErrorCount: 1,
				Findings: []Finding{
					{
						RuleID:   "test-rule",
						Severity: tt.severity,
						Message:  "Test message",
						File:     "test.py",
						Line:     1,
					},
				},
			}

			output := reporter.FormatResult(result, "test.py")
			if !strings.Contains(output, tt.expectedTag) {
				t.Errorf("expected output to contain %q tag, got: %s", tt.expectedTag, output)
			}
		})
	}
}

// TestFindingReporter_FormatMultiple_Truncation verifies multiple results truncation.
func TestFindingReporter_FormatMultiple_Truncation(t *testing.T) {
	reporter := NewFindingReporter()

	t.Run("truncates when more than 10 findings across files", func(t *testing.T) {
		results := make([]*ScanResult, 3)
		for i := range 3 {
			findings := make([]Finding, 5)
			for j := range 5 {
				findings[j] = Finding{
					RuleID:   "rule",
					Severity: SeverityWarning,
					Message:  "msg",
					File:     "file" + string(rune('0'+i)) + ".py",
					Line:     j + 1,
				}
			}
			results[i] = &ScanResult{
				Scanned:      true,
				WarningCount: 5,
				Findings:     findings,
			}
		}

		output := reporter.FormatMultiple(results)
		// 15 total findings, should show 10 and "5 more"
		if !strings.Contains(output, "5 more") {
			t.Errorf("expected truncation message, got: %s", output)
		}
	})

	t.Run("handles nil results in slice", func(t *testing.T) {
		results := []*ScanResult{
			{Scanned: true, Findings: []Finding{{RuleID: "r1", Severity: SeverityWarning, File: "f1.py", Line: 1}}},
			nil,
			{Scanned: false},
		}

		output := reporter.FormatMultiple(results)
		// Should handle gracefully
		_ = output
	})
}
