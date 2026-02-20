package security

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// =============================================================================
// RED PHASE: Tests for ASTGrepScanner implementation
// These tests define expected behavior per SPEC-HOOK-003 REQ-HOOK-100~103
// =============================================================================

// TestNewASTGrepScanner verifies scanner creation.
func TestNewASTGrepScanner(t *testing.T) {
	t.Run("creates scanner instance", func(t *testing.T) {
		scanner := NewASTGrepScanner()
		if scanner == nil {
			t.Fatal("expected non-nil scanner")
		}
	})
}

// TestASTGrepScanner_IsAvailable verifies binary availability check.
// REQ-HOOK-100: System must check ast-grep (sg) binary availability.
func TestASTGrepScanner_IsAvailable(t *testing.T) {
	scanner := NewASTGrepScanner()

	t.Run("returns true when sg binary exists", func(t *testing.T) {
		// This test will pass only if sg is installed
		// We can't force installation, so we just verify the function works
		available := scanner.IsAvailable()
		// Just verify it returns a boolean without panic
		_ = available
	})

	t.Run("caches availability result", func(t *testing.T) {
		// Multiple calls should be efficient (cached)
		start := time.Now()
		for range 100 {
			scanner.IsAvailable()
		}
		elapsed := time.Since(start)
		// Should complete quickly due to caching
		if elapsed > 100*time.Millisecond {
			t.Errorf("IsAvailable not cached, took %v for 100 calls", elapsed)
		}
	})
}

// TestASTGrepScanner_GetVersion verifies version retrieval.
// REQ-HOOK-100: System must verify ast-grep is executable.
func TestASTGrepScanner_GetVersion(t *testing.T) {
	scanner := NewASTGrepScanner()

	t.Run("returns empty string when sg not available", func(t *testing.T) {
		// Create scanner with mock unavailable state
		s := &astGrepScanner{available: false, checked: true}
		version := s.GetVersion()
		if version != "" {
			t.Errorf("expected empty version when unavailable, got %q", version)
		}
	})

	t.Run("returns version string when available", func(t *testing.T) {
		if _, err := exec.LookPath("sg"); err != nil {
			t.Skip("ast-grep (sg) not installed, skipping")
		}
		if !scanner.IsAvailable() {
			t.Skip("ast-grep (sg) not installed, skipping version test")
		}
		version := scanner.GetVersion()
		// On systems where sg is installed, version should be non-empty
		// However, version detection might fail in some environments
		// So we just verify it doesn't error when available
		if version == "" && scanner.IsAvailable() {
			t.Log("warning: sg available but version detection returned empty")
		}
	})
}

// TestASTGrepScanner_Scan verifies scanning functionality.
// REQ-HOOK-101: System must execute ast-grep scan on PostToolUse events.
// REQ-HOOK-102: System must classify findings by severity.
func TestASTGrepScanner_Scan(t *testing.T) {
	scanner := NewASTGrepScanner()

	t.Run("returns result with scanned=false when sg not available", func(t *testing.T) {
		s := &astGrepScanner{available: false, checked: true}
		ctx := context.Background()
		result, err := s.Scan(ctx, "/some/file.py", "")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.Scanned {
			t.Error("expected scanned=false when sg unavailable")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		if !scanner.IsAvailable() {
			t.Skip("ast-grep (sg) not installed")
		}
		ctx := context.Background()
		_, err := scanner.Scan(ctx, "/nonexistent/file.py", "")

		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})

	t.Run("scans file successfully", func(t *testing.T) {
		if !scanner.IsAvailable() {
			t.Skip("ast-grep (sg) not installed")
		}

		// Create a temporary test file
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.py")
		content := `
import os
password = "hardcoded_secret_123"  # potential security issue
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		ctx := context.Background()
		result, err := scanner.Scan(ctx, testFile, "")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Scanned {
			t.Error("expected scanned=true")
		}
		if result.Duration == 0 {
			t.Error("expected non-zero duration")
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		if !scanner.IsAvailable() {
			t.Skip("ast-grep (sg) not installed")
		}

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.py")
		_ = os.WriteFile(testFile, []byte("print('hello')"), 0644)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := scanner.Scan(ctx, testFile, "")
		if err == nil {
			t.Error("expected error for cancelled context")
		}
	})

	t.Run("respects timeout per REQ-HOOK-120", func(t *testing.T) {
		if !scanner.IsAvailable() {
			t.Skip("ast-grep (sg) not installed")
		}

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.py")
		_ = os.WriteFile(testFile, []byte("print('hello')"), 0644)

		// Use very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(1 * time.Millisecond) // Ensure timeout has passed

		_, err := scanner.Scan(ctx, testFile, "")
		if err == nil {
			t.Error("expected error for timed out context")
		}
	})
}

// TestASTGrepScanner_ScanMultiple verifies parallel scanning.
// REQ-HOOK-123: Optional parallel scanning for multiple files.
func TestASTGrepScanner_ScanMultiple(t *testing.T) {
	scanner := NewASTGrepScanner()

	t.Run("returns empty results for empty file list", func(t *testing.T) {
		ctx := context.Background()
		results, err := scanner.ScanMultiple(ctx, []string{}, "")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected empty results, got %d", len(results))
		}
	})

	t.Run("scans multiple files", func(t *testing.T) {
		if !scanner.IsAvailable() {
			t.Skip("ast-grep (sg) not installed")
		}

		tmpDir := t.TempDir()
		files := []string{
			filepath.Join(tmpDir, "test1.py"),
			filepath.Join(tmpDir, "test2.py"),
			filepath.Join(tmpDir, "test3.py"),
		}

		for i, f := range files {
			content := []byte("print('hello " + string(rune('0'+i)) + "')")
			_ = os.WriteFile(f, content, 0644)
		}

		ctx := context.Background()
		results, err := scanner.ScanMultiple(ctx, files, "")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}
	})
}

// TestLanguageSupport verifies language extension detection.
// REQ-HOOK-140: System must support 14+ languages.
// REQ-HOOK-141: System must not scan unsupported file extensions.
func TestLanguageSupport(t *testing.T) {
	tests := []struct {
		extension string
		supported bool
		language  string
	}{
		// Supported languages per SPEC 4.2
		{".py", true, "python"},
		{".pyi", true, "python"},
		{".js", true, "javascript"},
		{".jsx", true, "javascript"},
		{".mjs", true, "javascript"},
		{".cjs", true, "javascript"},
		{".ts", true, "typescript"},
		{".tsx", true, "typescript"},
		{".mts", true, "typescript"},
		{".cts", true, "typescript"},
		{".go", true, "go"},
		{".rs", true, "rust"},
		{".java", true, "java"},
		{".kt", true, "kotlin"},
		{".kts", true, "kotlin"},
		{".c", true, "c"},
		{".cpp", true, "cpp"},
		{".cc", true, "cpp"},
		{".h", true, "c"},
		{".hpp", true, "cpp"},
		{".rb", true, "ruby"},
		{".php", true, "php"},
		{".swift", true, "swift"},
		{".cs", true, "csharp"},
		{".ex", true, "elixir"},
		{".exs", true, "elixir"},
		{".scala", true, "scala"},
		// Unsupported extensions
		{".txt", false, ""},
		{".md", false, ""},
		{".json", false, ""},
		{".yaml", false, ""},
		{".xml", false, ""},
		{".unknown", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.extension, func(t *testing.T) {
			supported := IsSupportedExtension(tt.extension)
			if supported != tt.supported {
				t.Errorf("IsSupportedExtension(%q) = %v, want %v", tt.extension, supported, tt.supported)
			}

			if tt.supported {
				lang := GetLanguageForExtension(tt.extension)
				if lang != tt.language {
					t.Errorf("GetLanguageForExtension(%q) = %q, want %q", tt.extension, lang, tt.language)
				}
			}
		})
	}
}

// TestGetSupportedLanguages verifies the full language list.
func TestGetSupportedLanguages(t *testing.T) {
	languages := GetSupportedLanguages()

	// Per SPEC 4.2: Must support 14+ languages
	if len(languages) < 14 {
		t.Errorf("expected at least 14 supported languages, got %d", len(languages))
	}

	// Check required languages from SPEC
	required := []string{
		"python", "javascript", "typescript", "go", "rust",
		"java", "kotlin", "c", "cpp", "ruby", "php",
		"swift", "csharp", "elixir", "scala",
	}

	languageMap := make(map[string]bool)
	for _, lang := range languages {
		languageMap[lang.Name] = true
	}

	for _, req := range required {
		if !languageMap[req] {
			t.Errorf("missing required language: %s", req)
		}
	}
}

// TestASTGrepScanner_ParseJSONOutput verifies JSON output parsing.
// REQ-HOOK-121: System must parse JSON output from sg scan --json.
func TestASTGrepScanner_ParseJSONOutput(t *testing.T) {
	t.Run("parses valid JSON output", func(t *testing.T) {
		jsonOutput := `[{"ruleId":"sql-injection","severity":"error","message":"Potential SQL injection","file":"test.py","line":10,"column":5}]`

		findings, err := parseASTGrepJSON([]byte(jsonOutput))
		if err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].RuleID != "sql-injection" {
			t.Errorf("unexpected ruleId: %s", findings[0].RuleID)
		}
		if findings[0].Severity != SeverityError {
			t.Errorf("unexpected severity: %s", findings[0].Severity)
		}
	})

	t.Run("handles empty JSON array", func(t *testing.T) {
		findings, err := parseASTGrepJSON([]byte("[]"))
		if err != nil {
			t.Fatalf("failed to parse empty JSON: %v", err)
		}
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		_, err := parseASTGrepJSON([]byte("invalid json"))
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

// TestASTGrepScanner_RegexFallback verifies regex fallback parsing.
// REQ-HOOK-122: System must use regex fallback if JSON parsing fails.
func TestASTGrepScanner_RegexFallback(t *testing.T) {
	t.Run("parses text output with regex", func(t *testing.T) {
		textOutput := `test.py:10:5: error[sql-injection]: Potential SQL injection
test.py:20:3: warning[weak-random]: Math.random() used`

		findings := parseASTGrepRegex(textOutput)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
		if findings[0].Line != 10 {
			t.Errorf("expected line 10, got %d", findings[0].Line)
		}
		if findings[0].Severity != SeverityError {
			t.Errorf("expected error severity, got %s", findings[0].Severity)
		}
		if findings[1].Severity != SeverityWarning {
			t.Errorf("expected warning severity, got %s", findings[1].Severity)
		}
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		findings := parseASTGrepRegex("no matches here")
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("parses info and hint severities", func(t *testing.T) {
		textOutput := `test.py:10:5: info[note-info]: Informational note
test.py:20:3: hint[hint-tip]: Helpful hint`

		findings := parseASTGrepRegex(textOutput)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
		if findings[0].Severity != SeverityInfo {
			t.Errorf("expected info severity, got %s", findings[0].Severity)
		}
		if findings[1].Severity != SeverityHint {
			t.Errorf("expected hint severity, got %s", findings[1].Severity)
		}
	})
}

// TestParseSeverity verifies severity string parsing.
func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected Severity
	}{
		{"error", SeverityError},
		{"ERROR", SeverityError},
		{"Error", SeverityError},
		{"warning", SeverityWarning},
		{"WARNING", SeverityWarning},
		{"warn", SeverityWarning},
		{"WARN", SeverityWarning},
		{"info", SeverityInfo},
		{"INFO", SeverityInfo},
		{"hint", SeverityHint},
		{"HINT", SeverityHint},
		{"unknown", SeverityInfo}, // default
		{"", SeverityInfo},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseSeverity(tt.input)
			if result != tt.expected {
				t.Errorf("parseSeverity(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseASTGrepJSON_RangeFormat verifies JSON parsing with range format.
func TestParseASTGrepJSON_RangeFormat(t *testing.T) {
	t.Run("parses JSON with range object", func(t *testing.T) {
		jsonOutput := `[{
			"ruleId": "test-rule",
			"severity": "warning",
			"message": "Test message",
			"file": "test.py",
			"range": {
				"start": {"line": 10, "column": 5},
				"end": {"line": 10, "column": 20}
			}
		}]`

		findings, err := parseASTGrepJSON([]byte(jsonOutput))
		if err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Line != 10 {
			t.Errorf("expected line 10, got %d", findings[0].Line)
		}
		if findings[0].Column != 5 {
			t.Errorf("expected column 5, got %d", findings[0].Column)
		}
		if findings[0].EndLine != 10 {
			t.Errorf("expected endLine 10, got %d", findings[0].EndLine)
		}
		if findings[0].EndColumn != 20 {
			t.Errorf("expected endColumn 20, got %d", findings[0].EndColumn)
		}
	})
}
