package quality

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/astgrep"
	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
)

// makeError creates an error-severity Diagnostic for use in tests.
func makeError(line int, message, source string) lsphook.Diagnostic {
	return lsphook.Diagnostic{
		Range: lsphook.Range{
			Start: lsphook.Position{Line: line - 1}, // caller passes 1-based line
			End:   lsphook.Position{Line: line - 1},
		},
		Severity: lsphook.SeverityError,
		Message:  message,
		Source:   source,
	}
}

// makeWarning creates a warning-severity Diagnostic for use in tests.
func makeWarning(line int, message, source string) lsphook.Diagnostic {
	return lsphook.Diagnostic{
		Range: lsphook.Range{
			Start: lsphook.Position{Line: line - 1},
			End:   lsphook.Position{Line: line - 1},
		},
		Severity: lsphook.SeverityWarning,
		Message:  message,
		Source:   source,
	}
}

func TestFormatDiagnosticsAsInstructionWithFile_WithErrors(t *testing.T) {
	t.Parallel()

	diagnostics := []lsphook.Diagnostic{
		makeError(42, "undefined: foo", "gopls"),
		makeError(55, "unused import", "gopls"),
	}
	counts := lsphook.SeverityCounts{Errors: 2, Warnings: 0}

	got := FormatDiagnosticsAsInstructionWithFile("main.go", diagnostics, counts, false)

	// AC-LAI-001: both errors must appear.
	if !strings.Contains(got, "[Quality Gate] 2 error(s) detected in main.go:") {
		t.Errorf("header missing in output: %q", got)
	}
	// AC-LAI-002: specific file:line: message format.
	if !strings.Contains(got, "main.go:42: undefined: foo (gopls)") {
		t.Errorf("expected error entry not found in output: %q", got)
	}
	if !strings.Contains(got, "main.go:55: unused import (gopls)") {
		t.Errorf("expected second error entry not found in output: %q", got)
	}
	if !strings.Contains(got, "Fix these errors before proceeding.") {
		t.Errorf("closing instruction missing: %q", got)
	}
}

func TestFormatDiagnosticsAsInstructionWithFile_MessageFormat(t *testing.T) {
	t.Parallel()

	// AC-LAI-002: exact format check.
	diagnostics := []lsphook.Diagnostic{
		makeError(42, "undefined: foo", "gopls"),
	}
	counts := lsphook.SeverityCounts{Errors: 1}

	got := FormatDiagnosticsAsInstructionWithFile("main.go", diagnostics, counts, false)

	if !strings.Contains(got, "[Quality Gate] 1 error(s) detected in main.go:") {
		t.Errorf("header format incorrect: %q", got)
	}
	if !strings.Contains(got, "main.go:42: undefined: foo") {
		t.Errorf("entry format incorrect: %q", got)
	}
}

func TestFormatDiagnosticsAsInstructionWithFile_WithWarningsOnly_Disabled(t *testing.T) {
	t.Parallel()

	// AC-LAI-006: warnings with warn_as_instruction=false → empty.
	diagnostics := []lsphook.Diagnostic{
		makeWarning(10, "exported function missing comment", "golint"),
		makeWarning(20, "consider using early return", "staticcheck"),
		makeWarning(30, "unused variable", "gopls"),
	}
	counts := lsphook.SeverityCounts{Errors: 0, Warnings: 3}

	got := FormatDiagnosticsAsInstructionWithFile("main.go", diagnostics, counts, false)

	if got != "" {
		t.Errorf("expected empty string with warn_as_instruction=false, got: %q", got)
	}
}

func TestFormatDiagnosticsAsInstructionWithFile_WithWarningsOnly_Enabled(t *testing.T) {
	t.Parallel()

	// AC-LAI-007: warnings with warn_as_instruction=true → generate message.
	diagnostics := []lsphook.Diagnostic{
		makeWarning(10, "exported function missing comment", "golint"),
		makeWarning(20, "unused variable", "gopls"),
	}
	counts := lsphook.SeverityCounts{Errors: 0, Warnings: 2}

	got := FormatDiagnosticsAsInstructionWithFile("main.go", diagnostics, counts, true)

	if got == "" {
		t.Fatal("expected non-empty string with warn_as_instruction=true and warnings")
	}
	if !strings.Contains(got, "[Quality Gate]") {
		t.Errorf("header missing: %q", got)
	}
	if !strings.Contains(got, "warning(s)") {
		t.Errorf("warning kind missing: %q", got)
	}
	if !strings.Contains(got, "Fix these errors before proceeding.") {
		t.Errorf("closing instruction missing: %q", got)
	}
}

func TestFormatDiagnosticsAsInstructionWithFile_CleanFile(t *testing.T) {
	t.Parallel()

	// AC-LAI-008: clean file → empty string.
	diagnostics := []lsphook.Diagnostic{}
	counts := lsphook.SeverityCounts{Errors: 0, Warnings: 0}

	got := FormatDiagnosticsAsInstructionWithFile("main.go", diagnostics, counts, false)
	if got != "" {
		t.Errorf("expected empty string for clean file, got: %q", got)
	}

	got2 := FormatDiagnosticsAsInstructionWithFile("main.go", diagnostics, counts, true)
	if got2 != "" {
		t.Errorf("expected empty string for clean file (warn enabled), got: %q", got2)
	}
}

func TestFormatDiagnosticsAsInstructionWithFile_MaxErrors(t *testing.T) {
	t.Parallel()

	// AC-LAI-004: 15 errors → show 10, truncate 5.
	diagnostics := make([]lsphook.Diagnostic, 15)
	for i := range diagnostics {
		diagnostics[i] = makeError(i+1, "error message", "gopls")
	}
	counts := lsphook.SeverityCounts{Errors: 15}

	got := FormatDiagnosticsAsInstructionWithFile("main.go", diagnostics, counts, false)

	if !strings.Contains(got, "... and 5 more errors") {
		t.Errorf("truncation notice missing or wrong: %q", got)
	}

	// Count the number of "- main.go:" entries to verify exactly 10 are shown.
	lines := strings.Split(got, "\n")
	entryCount := 0
	for _, l := range lines {
		if strings.HasPrefix(l, "- main.go:") {
			entryCount++
		}
	}
	if entryCount != 10 {
		t.Errorf("expected 10 entries, got %d", entryCount)
	}
}

func TestFormatDiagnosticsAsInstructionWithFile_MixedErrorsWarnings(t *testing.T) {
	t.Parallel()

	// Mixed: errors take priority; warnings are not included when errors exist.
	diagnostics := []lsphook.Diagnostic{
		makeError(5, "undefined: bar", "gopls"),
		makeWarning(10, "unused import", "gopls"),
		makeError(20, "syntax error", "gopls"),
	}
	counts := lsphook.SeverityCounts{Errors: 2, Warnings: 1}

	got := FormatDiagnosticsAsInstructionWithFile("handler.go", diagnostics, counts, true)

	// Should show only errors (2), not warnings.
	if !strings.Contains(got, "2 error(s)") {
		t.Errorf("expected 2 errors in header: %q", got)
	}
	if strings.Contains(got, "warning") && strings.Contains(got, "unused import") {
		t.Errorf("warnings should not appear when errors exist: %q", got)
	}
}

func TestFormatDiagnosticsAsInstructionWithFile_EmptyFilePath(t *testing.T) {
	t.Parallel()

	diagnostics := []lsphook.Diagnostic{
		makeError(1, "undefined: x", ""),
	}
	counts := lsphook.SeverityCounts{Errors: 1}

	got := FormatDiagnosticsAsInstructionWithFile("", diagnostics, counts, false)

	// Should still produce output, just without a filename.
	if got == "" {
		t.Fatal("expected non-empty output even with empty filePath")
	}
	if !strings.Contains(got, "[Quality Gate] 1 error(s) detected") {
		t.Errorf("header missing: %q", got)
	}
}

func TestFormatDiagnosticsAsInstructionWithFile_NoSourceField(t *testing.T) {
	t.Parallel()

	// Diagnostic without source should omit the "(source)" parenthetical.
	diagnostics := []lsphook.Diagnostic{
		{
			Range:    lsphook.Range{Start: lsphook.Position{Line: 9}},
			Severity: lsphook.SeverityError,
			Message:  "undefined: something",
			Source:   "",
		},
	}
	counts := lsphook.SeverityCounts{Errors: 1}

	got := FormatDiagnosticsAsInstructionWithFile("file.go", diagnostics, counts, false)

	// Should NOT contain "()" when source is empty.
	if strings.Contains(got, "()") {
		t.Errorf("unexpected empty parenthetical in output: %q", got)
	}
	if !strings.Contains(got, "file.go:10: undefined: something") {
		t.Errorf("entry format incorrect without source: %q", got)
	}
}

// --- FormatDiagnosticsAsInstruction (no file path) tests ---

func TestFormatDiagnosticsAsInstruction_WithErrors(t *testing.T) {
	t.Parallel()

	diagnostics := []lsphook.Diagnostic{
		makeError(5, "undefined: Foo", "gopls"),
	}
	counts := lsphook.SeverityCounts{Errors: 1}

	got := FormatDiagnosticsAsInstruction(diagnostics, counts, false)
	if got == "" {
		t.Fatal("expected non-empty output")
	}
	if !strings.Contains(got, "[Quality Gate]") {
		t.Errorf("header missing: %q", got)
	}
}

func TestFormatDiagnosticsAsInstruction_Clean(t *testing.T) {
	t.Parallel()

	got := FormatDiagnosticsAsInstruction(nil, lsphook.SeverityCounts{}, false)
	if got != "" {
		t.Errorf("expected empty string, got: %q", got)
	}
}

func TestFormatDiagnosticsAsInstruction_MaxErrors(t *testing.T) {
	t.Parallel()

	diagnostics := make([]lsphook.Diagnostic, 12)
	for i := range diagnostics {
		diagnostics[i] = makeError(i+1, "some error", "gopls")
	}
	counts := lsphook.SeverityCounts{Errors: 12}

	got := FormatDiagnosticsAsInstruction(diagnostics, counts, false)
	if !strings.Contains(got, "... and 2 more errors") {
		t.Errorf("truncation notice missing: %q", got)
	}
}

// --- AppendAstSecurityFindings tests (REQ-LAI-008) ---

func TestAppendAstSecurityFindings_NilResult(t *testing.T) {
	t.Parallel()

	// nil result should return existing message unchanged.
	existing := "some existing message"
	got := AppendAstSecurityFindings(existing, "/tmp/main.go", nil)
	if got != existing {
		t.Errorf("nil result: expected unchanged message, got: %q", got)
	}
}

func TestAppendAstSecurityFindings_EmptyMatches(t *testing.T) {
	t.Parallel()

	result := &astgrep.ScanResult{Matches: []astgrep.Match{}}
	existing := "some message"
	got := AppendAstSecurityFindings(existing, "/tmp/main.go", result)
	if got != existing {
		t.Errorf("empty matches: expected unchanged message, got: %q", got)
	}
}

func TestAppendAstSecurityFindings_WithMatches_NoExisting(t *testing.T) {
	t.Parallel()

	result := &astgrep.ScanResult{
		Matches: []astgrep.Match{
			{File: "/tmp/main.go", Line: 10, Rule: "no-hardcoded-secrets", Message: "hardcoded secret"},
		},
	}

	got := AppendAstSecurityFindings("", "/tmp/main.go", result)

	if got == "" {
		t.Fatal("expected non-empty output")
	}
	if !strings.Contains(got, "[Security Gate] 1 security issue(s) detected in main.go:") {
		t.Errorf("security header missing: %q", got)
	}
	if !strings.Contains(got, "no-hardcoded-secrets") {
		t.Errorf("rule name missing: %q", got)
	}
	if !strings.Contains(got, "Fix these errors before proceeding.") {
		t.Errorf("closing instruction missing: %q", got)
	}
}

func TestAppendAstSecurityFindings_WithMatches_AppendedToExisting(t *testing.T) {
	t.Parallel()

	// Simulate an existing LSP message.
	existing := "[Quality Gate] 1 error(s) detected in main.go:\n- main.go:5: undefined: foo (gopls)\nFix these errors before proceeding."

	result := &astgrep.ScanResult{
		Matches: []astgrep.Match{
			{File: "/tmp/main.go", Line: 20, Rule: "sql-injection", Message: "potential SQL injection"},
		},
	}

	got := AppendAstSecurityFindings(existing, "/tmp/main.go", result)

	// Both LSP and security sections should be present.
	if !strings.Contains(got, "[Quality Gate]") {
		t.Errorf("LSP section missing in merged message: %q", got)
	}
	if !strings.Contains(got, "[Security Gate]") {
		t.Errorf("security section missing in merged message: %q", got)
	}
	if !strings.Contains(got, "sql-injection") {
		t.Errorf("rule name missing in merged message: %q", got)
	}
	// Only one "Fix these errors" at the end.
	if strings.Count(got, "Fix these errors before proceeding.") != 1 {
		t.Errorf("expected exactly one closing instruction, got: %q", got)
	}
}

func TestAppendAstSecurityFindings_MaxEntries(t *testing.T) {
	t.Parallel()

	matches := make([]astgrep.Match, 15)
	for i := range matches {
		matches[i] = astgrep.Match{File: "/tmp/main.go", Line: i + 1, Rule: "rule", Message: "security issue"}
	}
	result := &astgrep.ScanResult{Matches: matches}

	got := AppendAstSecurityFindings("", "/tmp/main.go", result)

	if !strings.Contains(got, "... and 5 more security issues") {
		t.Errorf("truncation notice missing: %q", got)
	}
}

func TestAppendAstSecurityFindings_EmptyFilePath(t *testing.T) {
	t.Parallel()

	result := &astgrep.ScanResult{
		Matches: []astgrep.Match{
			{File: "/tmp/main.go", Line: 5, Rule: "test-rule", Message: "test message"},
		},
	}

	got := AppendAstSecurityFindings("", "", result)

	if got == "" {
		t.Fatal("expected non-empty output even with empty filePath")
	}
	if !strings.Contains(got, "[Security Gate]") {
		t.Errorf("security header missing: %q", got)
	}
}

func TestAppendAstSecurityFindings_NoRuleField(t *testing.T) {
	t.Parallel()

	result := &astgrep.ScanResult{
		Matches: []astgrep.Match{
			{File: "/tmp/main.go", Line: 3, Rule: "", Message: "suspicious pattern"},
		},
	}

	got := AppendAstSecurityFindings("", "/tmp/main.go", result)

	if strings.Contains(got, "()") {
		t.Errorf("unexpected empty parenthetical in output: %q", got)
	}
	if !strings.Contains(got, "suspicious pattern") {
		t.Errorf("message missing: %q", got)
	}
}
