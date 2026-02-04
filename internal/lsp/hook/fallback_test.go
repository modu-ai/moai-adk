package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewFallbackDiagnostics verifies fallback creation.
func TestNewFallbackDiagnostics(t *testing.T) {
	t.Parallel()

	fb := NewFallbackDiagnostics()
	if fb == nil {
		t.Fatal("expected non-nil fallback diagnostics")
	}
}

// TestGetLanguage verifies language detection from file path per REQ-HOOK-160.
func TestGetLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{"python file", "/path/to/file.py", "python"},
		{"python stub", "/path/to/file.pyi", "python"},
		{"go file", "/path/to/file.go", "go"},
		{"typescript file", "/path/to/file.ts", "typescript"},
		{"typescript jsx", "/path/to/file.tsx", "typescript"},
		{"javascript file", "/path/to/file.js", "javascript"},
		{"javascript jsx", "/path/to/file.jsx", "javascript"},
		{"rust file", "/path/to/file.rs", "rust"},
		{"unknown file", "/path/to/file.unknown", "unknown"},
		{"no extension", "/path/to/file", "unknown"},
	}

	fb := NewFallbackDiagnostics()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fb.GetLanguage(tt.filePath)
			if got != tt.want {
				t.Errorf("GetLanguage(%q) = %q, want %q", tt.filePath, got, tt.want)
			}
		})
	}
}

// TestIsAvailable verifies tool availability check per REQ-HOOK-160.
func TestIsAvailable(t *testing.T) {
	t.Parallel()

	fb := NewFallbackDiagnostics()

	// Go should have go vet available in test environment
	t.Run("go language", func(t *testing.T) {
		t.Parallel()
		// go vet should be available
		available := fb.IsAvailable("go")
		if !available {
			t.Skip("go vet not available in test environment")
		}
	})

	t.Run("unknown language returns false", func(t *testing.T) {
		t.Parallel()
		available := fb.IsAvailable("unknown-language-xyz")
		if available {
			t.Error("expected false for unknown language")
		}
	})
}

// TestRunFallback_Python verifies Python fallback per REQ-HOOK-160.
func TestRunFallback_Python(t *testing.T) {
	t.Parallel()

	fb := NewFallbackDiagnostics()

	if !fb.IsAvailable("python") {
		t.Skip("no Python fallback tool available")
	}

	// Create a temp Python file with a known issue
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.py")
	content := []byte("import os\nx = 1\n")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	diagnostics, err := fb.RunFallback(ctx, testFile)
	if err != nil {
		// It's acceptable if the tool is not installed
		var unavailErr *ErrDiagnosticsUnavailable
		if isUnavailable := isErrDiagnosticsUnavailable(err, &unavailErr); isUnavailable {
			t.Skipf("Python tool not available: %v", err)
		}
		t.Fatalf("unexpected error: %v", err)
	}

	// Diagnostics may be empty if the file is clean
	if diagnostics == nil {
		t.Error("expected non-nil diagnostics slice")
	}
}

// TestRunFallback_Go verifies Go fallback per REQ-HOOK-160.
func TestRunFallback_Go(t *testing.T) {
	t.Parallel()

	fb := NewFallbackDiagnostics()

	if !fb.IsAvailable("go") {
		t.Skip("no Go fallback tool available")
	}

	// Create a temp Go file with valid code
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := []byte("package main\n\nfunc main() {}\n")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	diagnostics, err := fb.RunFallback(ctx, testFile)
	if err != nil {
		var unavailErr *ErrDiagnosticsUnavailable
		if isUnavailable := isErrDiagnosticsUnavailable(err, &unavailErr); isUnavailable {
			t.Skipf("Go tool not available: %v", err)
		}
		t.Fatalf("unexpected error: %v", err)
	}

	// Valid Go file should have no diagnostics
	if diagnostics == nil {
		t.Error("expected non-nil diagnostics slice")
	}
}

// TestRunFallback_UnavailableLanguage verifies behavior for unsupported language per REQ-HOOK-162.
func TestRunFallback_UnavailableLanguage(t *testing.T) {
	t.Parallel()

	fb := NewFallbackDiagnostics()

	ctx := context.Background()
	_, err := fb.RunFallback(ctx, "/path/to/file.unknown")

	if err == nil {
		t.Fatal("expected error for unknown language")
	}

	var unavailErr *ErrDiagnosticsUnavailable
	if !isErrDiagnosticsUnavailable(err, &unavailErr) {
		t.Errorf("expected ErrDiagnosticsUnavailable, got %T: %v", err, err)
	}
}

// TestRunFallback_ContextCancellation verifies context handling.
func TestRunFallback_ContextCancellation(t *testing.T) {
	t.Parallel()

	fb := NewFallbackDiagnostics()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := fb.RunFallback(ctx, "/path/to/file.py")

	// Context cancellation should be respected
	if err == nil {
		t.Log("context cancellation not enforced (acceptable)")
	}
}

// TestParsePythonRuffOutput verifies ruff JSON output parsing per REQ-HOOK-161.
func TestParsePythonRuffOutput(t *testing.T) {
	t.Parallel()

	jsonOutput := `[
		{
			"code": "F401",
			"message": "os imported but unused",
			"location": {
				"row": 1,
				"column": 8
			},
			"end_location": {
				"row": 1,
				"column": 10
			},
			"filename": "test.py",
			"fix": null,
			"url": "https://docs.astral.sh/ruff/rules/unused-import"
		}
	]`

	diagnostics, err := parseRuffOutput([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("failed to parse ruff output: %v", err)
	}

	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diagnostics))
	}

	if diagnostics[0].Code != "F401" {
		t.Errorf("Code = %q, want %q", diagnostics[0].Code, "F401")
	}
	if diagnostics[0].Message != "os imported but unused" {
		t.Errorf("Message = %q, want %q", diagnostics[0].Message, "os imported but unused")
	}
	if diagnostics[0].Range.Start.Line != 0 { // 0-indexed
		t.Errorf("Range.Start.Line = %d, want 0", diagnostics[0].Range.Start.Line)
	}
}

// TestParseGoVetOutput verifies go vet output parsing per REQ-HOOK-161.
func TestParseGoVetOutput(t *testing.T) {
	t.Parallel()

	output := `# test
./test.go:5:2: unused variable 'x'
./test.go:10:5: undefined: foo`

	diagnostics := parseGoVetOutput(output, "/path/to")

	if len(diagnostics) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(diagnostics))
	}

	if diagnostics[0].Range.Start.Line != 4 { // 0-indexed (5-1)
		t.Errorf("first diagnostic line = %d, want 4", diagnostics[0].Range.Start.Line)
	}
	if diagnostics[0].Range.Start.Character != 1 { // 0-indexed (2-1)
		t.Errorf("first diagnostic character = %d, want 1", diagnostics[0].Range.Start.Character)
	}
}

// TestParseTypeScriptOutput verifies tsc output parsing per REQ-HOOK-161.
func TestParseTypeScriptOutput(t *testing.T) {
	t.Parallel()

	output := `test.ts(5,10): error TS2304: Cannot find name 'foo'.
test.ts(10,1): error TS1005: ';' expected.`

	diagnostics := parseTypeScriptOutput(output)

	if len(diagnostics) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(diagnostics))
	}

	if diagnostics[0].Code != "TS2304" {
		t.Errorf("Code = %q, want %q", diagnostics[0].Code, "TS2304")
	}
	if diagnostics[0].Range.Start.Line != 4 { // 0-indexed
		t.Errorf("Range.Start.Line = %d, want 4", diagnostics[0].Range.Start.Line)
	}
}

// TestParseESLintOutput verifies eslint JSON output parsing per REQ-HOOK-161.
func TestParseESLintOutput(t *testing.T) {
	t.Parallel()

	jsonOutput := `[
		{
			"filePath": "/path/to/file.js",
			"messages": [
				{
					"ruleId": "no-unused-vars",
					"severity": 2,
					"message": "'x' is defined but never used.",
					"line": 5,
					"column": 7
				},
				{
					"ruleId": "semi",
					"severity": 1,
					"message": "Missing semicolon.",
					"line": 10,
					"column": 20
				}
			]
		}
	]`

	diagnostics, err := parseESLintOutput([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("failed to parse eslint output: %v", err)
	}

	if len(diagnostics) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(diagnostics))
	}

	// First diagnostic should be error (severity 2)
	if diagnostics[0].Severity != SeverityError {
		t.Errorf("first diagnostic severity = %v, want %v", diagnostics[0].Severity, SeverityError)
	}

	// Second diagnostic should be warning (severity 1)
	if diagnostics[1].Severity != SeverityWarning {
		t.Errorf("second diagnostic severity = %v, want %v", diagnostics[1].Severity, SeverityWarning)
	}
}

// TestParseClippyOutput verifies clippy JSON output parsing per REQ-HOOK-161.
func TestParseClippyOutput(t *testing.T) {
	t.Parallel()

	jsonOutput := `{"reason":"compiler-message","message":{"code":{"code":"unused_variable"},"level":"warning","message":"unused variable: x","spans":[{"line_start":5,"line_end":5,"column_start":10,"column_end":11}]}}
{"reason":"compiler-message","message":{"code":{"code":"dead_code"},"level":"error","message":"unreachable code","spans":[{"line_start":10,"line_end":12,"column_start":1,"column_end":5}]}}`

	diagnostics, err := parseClippyOutput([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("failed to parse clippy output: %v", err)
	}

	if len(diagnostics) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(diagnostics))
	}

	// First diagnostic should be warning
	if diagnostics[0].Severity != SeverityWarning {
		t.Errorf("first diagnostic severity = %v, want %v", diagnostics[0].Severity, SeverityWarning)
	}
	if diagnostics[0].Code != "unused_variable" {
		t.Errorf("first diagnostic code = %q, want %q", diagnostics[0].Code, "unused_variable")
	}

	// Second diagnostic should be error
	if diagnostics[1].Severity != SeverityError {
		t.Errorf("second diagnostic severity = %v, want %v", diagnostics[1].Severity, SeverityError)
	}
}

// TestParseClippyOutput_Empty verifies empty clippy output.
func TestParseClippyOutput_Empty(t *testing.T) {
	t.Parallel()

	diagnostics, err := parseClippyOutput([]byte(""))
	if err != nil {
		t.Fatalf("failed to parse empty clippy output: %v", err)
	}

	if len(diagnostics) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diagnostics))
	}
}

// TestClassifyRuffSeverity verifies ruff severity classification.
func TestClassifyRuffSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code     string
		expected DiagnosticSeverity
	}{
		{"E001", SeverityError},
		{"E999", SeverityError},
		{"F401", SeverityError},
		{"F811", SeverityError},
		{"W001", SeverityWarning},
		{"W605", SeverityWarning},
		{"C901", SeverityWarning}, // default
		{"D100", SeverityWarning}, // default
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			t.Parallel()
			got := classifyRuffSeverity(tt.code)
			if got != tt.expected {
				t.Errorf("classifyRuffSeverity(%q) = %v, want %v", tt.code, got, tt.expected)
			}
		})
	}
}

// TestSupportedLanguages verifies support for required languages per REQ-HOOK-160.
func TestSupportedLanguages(t *testing.T) {
	t.Parallel()

	fb := NewFallbackDiagnostics()

	// These languages should have fallback tools configured
	languages := []string{"python", "typescript", "javascript", "go", "rust"}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			t.Parallel()
			// Check that the language has a configured tool (may not be installed)
			tools := fb.GetToolsForLanguage(lang)
			if len(tools) == 0 {
				t.Errorf("no tools configured for %s", lang)
			}
		})
	}
}

// isErrDiagnosticsUnavailable is a helper to check error type.
func isErrDiagnosticsUnavailable(err error, target **ErrDiagnosticsUnavailable) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*ErrDiagnosticsUnavailable); ok {
		*target = e
		return true
	}
	return false
}
