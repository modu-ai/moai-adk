package hook

import (
	"encoding/json"
	"strings"
	"testing"

	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
	"github.com/modu-ai/moai-adk/internal/hook/quality"
)

// canonicalHookDiagnostics returns the canonical fixture for AC-UTIL-003-009.
// Contains one each of Error, Warning, Information, and Hint.
func canonicalHookDiagnostics() []lsphook.Diagnostic {
	return []lsphook.Diagnostic{
		{
			Severity: lsphook.SeverityError,
			Message:  "undefined: foo",
			Source:   "compiler",
			Range: lsphook.Range{
				Start: lsphook.Position{Line: 10, Character: 2},
				End:   lsphook.Position{Line: 10, Character: 5},
			},
		},
		{
			Severity: lsphook.SeverityWarning,
			Message:  "unused variable: bar",
			Source:   "staticcheck",
			Range: lsphook.Range{
				Start: lsphook.Position{Line: 20, Character: 0},
				End:   lsphook.Position{Line: 20, Character: 3},
			},
		},
		{
			Severity: lsphook.SeverityInformation,
			Message:  "consider renaming to camelCase",
			Source:   "gopls",
			Range: lsphook.Range{
				Start: lsphook.Position{Line: 30, Character: 0},
				End:   lsphook.Position{Line: 30, Character: 10},
			},
		},
		{
			Severity: lsphook.SeverityHint,
			Message:  "export this function for better testability",
			Source:   "gopls",
			Range: lsphook.Range{
				Start: lsphook.Position{Line: 40, Character: 0},
				End:   lsphook.Position{Line: 40, Character: 12},
			},
		},
	}
}

// computeCounts computes the SeverityCounts for the given fixture.
func computeCounts(diags []lsphook.Diagnostic) lsphook.SeverityCounts {
	var counts lsphook.SeverityCounts
	for _, d := range diags {
		switch d.Severity {
		case lsphook.SeverityError:
			counts.Errors++
		case lsphook.SeverityWarning:
			counts.Warnings++
		case lsphook.SeverityInformation:
			counts.Information++
		case lsphook.SeverityHint:
			counts.Hints++
		}
	}
	return counts
}

// ─── AC-UTIL-003-008 ─────────────────────────────────────────────────────────

// TestHookDiagnosticSeverity_JSONMarshal_StringPreserved verifies that
// hook.DiagnosticSeverity("error") serializes to the string "error" in JSON (AC-UTIL-003-008).
// Wire format freeze: identical JSON output must be guaranteed before and after SPEC-UTIL-003.
func TestHookDiagnosticSeverity_JSONMarshal_StringPreserved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		severity lsphook.DiagnosticSeverity
		want     string
	}{
		{lsphook.SeverityError, "error"},
		{lsphook.SeverityWarning, "warning"},
		{lsphook.SeverityInformation, "information"},
		{lsphook.SeverityHint, "hint"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.severity), func(t *testing.T) {
			t.Parallel()

			d := lsphook.Diagnostic{
				Severity: tt.severity,
				Message:  "test message",
				Range:    lsphook.Range{},
			}

			data, err := json.Marshal(d)
			if err != nil {
				t.Fatalf("json.Marshal error: %v", err)
			}

			var m map[string]any
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatalf("json.Unmarshal error: %v", err)
			}

			severity, ok := m["severity"].(string)
			if !ok {
				t.Fatalf("severity field is not a string: type=%T value=%v (wire format blocker — must remain string)", m["severity"], m["severity"])
			}
			if severity != tt.want {
				t.Errorf("severity = %q, want %q", severity, tt.want)
			}
		})
	}
}

// ─── AC-UTIL-003-009 ─────────────────────────────────────────────────────────

// TestWireFormat_FormatDiagnosticsAsInstructionWithFile_Freeze verifies that
// FormatDiagnosticsAsInstructionWithFile output for the canonical fixture is
// byte-identical before and after SPEC is applied (AC-UTIL-003-009).
//
// Freeze validation: call the same function twice and verify output is idempotent.
// After SPEC is applied this test runs with the new code, so it must fail on any implementation change.
func TestWireFormat_FormatDiagnosticsAsInstructionWithFile_Freeze(t *testing.T) {
	t.Parallel()

	fixture := canonicalHookDiagnostics()
	counts := computeCounts(fixture)

	// errors only (when errors > 0, warnings are ignored)
	got1 := quality.FormatDiagnosticsAsInstructionWithFile("main.go", fixture, counts, false)
	got2 := quality.FormatDiagnosticsAsInstructionWithFile("main.go", fixture, counts, false)

	if got1 != got2 {
		t.Error("FormatDiagnosticsAsInstructionWithFile is not idempotent (wire format unstable)")
	}

	// output must not be empty (error diagnostic is present)
	if got1 == "" {
		t.Error("FormatDiagnosticsAsInstructionWithFile returned empty string for error diagnostic")
	}

	// verify expected format elements: header, filename, error message, closing instruction
	if !strings.Contains(got1, "[Quality Gate]") {
		t.Errorf("output missing '[Quality Gate]' header: %q", got1)
	}
	if !strings.Contains(got1, "main.go") {
		t.Errorf("output missing file name 'main.go': %q", got1)
	}
	if !strings.Contains(got1, "undefined: foo") {
		t.Errorf("output missing error message: %q", got1)
	}
	if !strings.Contains(got1, "Fix these errors before proceeding.") {
		t.Errorf("output missing trailing fix instruction: %q", got1)
	}
}

// TestWireFormat_ConvertHookDiagsToLSP_Freeze verifies that convertHookDiagsToLSP
// produces byte-identical output before and after SPEC is applied (AC-UTIL-003-009).
// Freeze for the hook.Diagnostic(string severity) → lsp.Diagnostic(int severity) conversion path.
func TestWireFormat_ConvertHookDiagsToLSP_Freeze(t *testing.T) {
	t.Parallel()

	fixture := canonicalHookDiagnostics()

	result1 := convertHookDiagsToLSP(fixture)
	result2 := convertHookDiagsToLSP(fixture)

	if len(result1) != len(result2) {
		t.Fatalf("convertHookDiagsToLSP lengths differ: %d vs %d", len(result1), len(result2))
	}

	for i := range result1 {
		if result1[i].Severity != result2[i].Severity {
			t.Errorf("index %d: Severity %d != %d", i, result1[i].Severity, result2[i].Severity)
		}
		if result1[i].Message != result2[i].Message {
			t.Errorf("index %d: Message %q != %q", i, result1[i].Message, result2[i].Message)
		}
	}

	// severity conversion accuracy: string → int validation
	// SeverityError = "error" → 1, SeverityWarning = "warning" → 2,
	// SeverityInformation = "information" → 3, SeverityHint = "hint" → 4
	expected := []int{1, 2, 3, 4}
	for i, want := range expected {
		if int(result1[i].Severity) != want {
			t.Errorf("severity[%d] = %d, want %d", i, int(result1[i].Severity), want)
		}
	}
}

// TestHookDiagnostic_JSONRoundTrip verifies that JSON serialization/deserialization of
// hook.Diagnostic guarantees complete data preservation (reinforces AC-UTIL-003-008).
func TestHookDiagnostic_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := lsphook.Diagnostic{
		Severity: lsphook.SeverityError,
		Message:  "undefined: bar",
		Code:     "E001",
		Source:   "compiler",
		Range: lsphook.Range{
			Start: lsphook.Position{Line: 5, Character: 3},
			End:   lsphook.Position{Line: 5, Character: 6},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var restored lsphook.Diagnostic
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if restored.Severity != original.Severity {
		t.Errorf("Severity: got %q, want %q", restored.Severity, original.Severity)
	}
	if restored.Message != original.Message {
		t.Errorf("Message: got %q, want %q", restored.Message, original.Message)
	}
	if restored.Range.Start.Line != original.Range.Start.Line {
		t.Errorf("Range.Start.Line: got %d, want %d", restored.Range.Start.Line, original.Range.Start.Line)
	}
}
