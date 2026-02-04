package security

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// =============================================================================
// RED PHASE: Tests for SecurityScanner (main entry point)
// These tests define expected behavior per SPEC-HOOK-003 REQ-HOOK-120~123
// =============================================================================

// TestNewSecurityScanner verifies scanner creation.
func TestNewSecurityScanner(t *testing.T) {
	t.Run("creates scanner with default components", func(t *testing.T) {
		scanner := NewSecurityScanner()
		if scanner == nil {
			t.Fatal("expected non-nil scanner")
		}
	})

	t.Run("creates scanner with custom config", func(t *testing.T) {
		config := &ScannerConfig{
			Timeout:    60 * time.Second,
			ConfigPath: "/custom/sgconfig.yml",
		}
		scanner := NewSecurityScannerWithConfig(config)
		if scanner == nil {
			t.Fatal("expected non-nil scanner")
		}
	})
}

// TestSecurityScanner_ScanFile verifies single file scanning.
// REQ-HOOK-120: System must complete scans within timeout.
// REQ-HOOK-121: System must execute sg scan --json.
func TestSecurityScanner_ScanFile(t *testing.T) {
	scanner := NewSecurityScanner()

	t.Run("scans supported file type", func(t *testing.T) {
		if !scanner.IsAvailable() {
			t.Skip("ast-grep (sg) not installed")
		}

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.py")
		content := `
import os
password = "secret123"  # potential issue
`
		_ = os.WriteFile(testFile, []byte(content), 0644)

		ctx := context.Background()
		result, err := scanner.ScanFile(ctx, testFile, tmpDir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if !result.Scanned {
			t.Error("expected scanned=true")
		}
	})

	t.Run("skips unsupported file type", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		_ = os.WriteFile(testFile, []byte("hello world"), 0644)

		ctx := context.Background()
		result, err := scanner.ScanFile(ctx, testFile, tmpDir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Scanned {
			t.Error("expected scanned=false for unsupported file type")
		}
	})

	t.Run("respects timeout per REQ-HOOK-120", func(t *testing.T) {
		if !scanner.IsAvailable() {
			t.Skip("ast-grep (sg) not installed")
		}

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.py")
		_ = os.WriteFile(testFile, []byte("print('hello')"), 0644)

		// Very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(1 * time.Millisecond)

		_, err := scanner.ScanFile(ctx, testFile, tmpDir)
		if err == nil {
			t.Error("expected timeout error")
		}
	})
}

// TestSecurityScanner_ScanFiles verifies multiple file scanning.
// REQ-HOOK-123: Optional parallel scanning for multiple files.
func TestSecurityScanner_ScanFiles(t *testing.T) {
	scanner := NewSecurityScanner()

	t.Run("scans multiple files in parallel", func(t *testing.T) {
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
		results, err := scanner.ScanFiles(ctx, files, tmpDir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}
	})

	t.Run("handles mixed file types", func(t *testing.T) {
		tmpDir := t.TempDir()
		files := []string{
			filepath.Join(tmpDir, "test.py"),
			filepath.Join(tmpDir, "test.txt"), // unsupported
			filepath.Join(tmpDir, "test.go"),
		}

		for _, f := range files {
			_ = os.WriteFile(f, []byte("content"), 0644)
		}

		ctx := context.Background()
		results, err := scanner.ScanFiles(ctx, files, tmpDir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}

		// txt file should be skipped
		if results[1].Scanned {
			t.Error("expected txt file to be skipped")
		}
	})
}

// TestSecurityScanner_IsAvailable verifies availability check.
func TestSecurityScanner_IsAvailable(t *testing.T) {
	scanner := NewSecurityScanner()

	t.Run("returns boolean without panic", func(t *testing.T) {
		available := scanner.IsAvailable()
		_ = available // Just verify it doesn't panic
	})
}

// TestSecurityScanner_GetReport verifies report generation.
func TestSecurityScanner_GetReport(t *testing.T) {
	scanner := NewSecurityScanner()

	t.Run("generates report for scan result", func(t *testing.T) {
		result := &ScanResult{
			Scanned:    true,
			ErrorCount: 1,
			Findings: []Finding{
				{
					RuleID:   "sql-injection",
					Severity: SeverityError,
					Message:  "Potential SQL injection",
					File:     "test.py",
					Line:     10,
				},
			},
		}

		report := scanner.GetReport(result, "test.py")
		if report == "" {
			t.Error("expected non-empty report")
		}
	})
}

// TestSecurityScanner_Integration verifies full integration flow.
func TestSecurityScanner_Integration(t *testing.T) {
	scanner := NewSecurityScanner()

	if !scanner.IsAvailable() {
		t.Skip("ast-grep (sg) not installed - skipping integration test")
	}

	t.Run("full scan workflow", func(t *testing.T) {
		// Create project with config and test file
		tmpDir := t.TempDir()

		// Create test Python file with potential issues
		testFile := filepath.Join(tmpDir, "app.py")
		content := `
import os
import subprocess

def run_query(user_input):
    # Potential SQL injection
    query = "SELECT * FROM users WHERE name = '" + user_input + "'"
    return query

def run_command(cmd):
    # Potential command injection
    os.system(cmd)

# Hardcoded secret
API_KEY = "sk-12345678901234567890"
`
		_ = os.WriteFile(testFile, []byte(content), 0644)

		// Scan file
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := scanner.ScanFile(ctx, testFile, tmpDir)
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}

		if !result.Scanned {
			t.Error("expected file to be scanned")
		}

		// Generate report
		report := scanner.GetReport(result, testFile)
		t.Logf("Integration test report:\n%s", report)
	})
}

// TestScannerConfig verifies configuration options.
func TestScannerConfig(t *testing.T) {
	t.Run("default config values", func(t *testing.T) {
		config := DefaultScannerConfig()

		if config.Timeout != DefaultScanTimeout {
			t.Errorf("expected default timeout %v, got %v", DefaultScanTimeout, config.Timeout)
		}
	})

	t.Run("custom timeout", func(t *testing.T) {
		config := &ScannerConfig{
			Timeout: 60 * time.Second,
		}
		scanner := NewSecurityScannerWithConfig(config)

		if scanner == nil {
			t.Fatal("expected non-nil scanner")
		}
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		scanner := NewSecurityScannerWithConfig(nil)
		if scanner == nil {
			t.Fatal("expected non-nil scanner")
		}
	})

	t.Run("zero timeout uses default", func(t *testing.T) {
		config := &ScannerConfig{
			Timeout: 0, // Should use default
		}
		scanner := NewSecurityScannerWithConfig(config)
		if scanner == nil {
			t.Fatal("expected non-nil scanner")
		}
	})
}

// TestSecurityScanner_GetMultiReport verifies multi-file report generation.
func TestSecurityScanner_GetMultiReport(t *testing.T) {
	scanner := NewSecurityScanner()

	t.Run("generates report for multiple results", func(t *testing.T) {
		results := []*ScanResult{
			{
				Scanned:    true,
				ErrorCount: 1,
				Findings: []Finding{
					{RuleID: "r1", Severity: SeverityError, Message: "msg1", File: "f1.py", Line: 1},
				},
			},
			{
				Scanned:      true,
				WarningCount: 1,
				Findings: []Finding{
					{RuleID: "r2", Severity: SeverityWarning, Message: "msg2", File: "f2.py", Line: 1},
				},
			},
		}

		report := scanner.GetMultiReport(results)
		if report == "" {
			t.Error("expected non-empty report")
		}
	})

	t.Run("handles empty results", func(t *testing.T) {
		report := scanner.GetMultiReport([]*ScanResult{})
		// Should return empty string for empty results
		_ = report
	})
}

// TestSecurityScanner_ShouldAlert verifies alert decision logic.
func TestSecurityScanner_ShouldAlert(t *testing.T) {
	scanner := NewSecurityScanner()

	t.Run("returns false for no errors", func(t *testing.T) {
		result := &ScanResult{Scanned: true, ErrorCount: 0}
		if scanner.ShouldAlert(result) {
			t.Error("should not alert for no errors")
		}
	})

	t.Run("returns true for errors", func(t *testing.T) {
		result := &ScanResult{
			Scanned:    true,
			ErrorCount: 1,
			Findings:   []Finding{{Severity: SeverityError}},
		}
		if !scanner.ShouldAlert(result) {
			t.Error("should alert for errors")
		}
	})
}

// TestSecurityScanner_GetExitCode verifies exit code logic.
// REQ-HOOK-131: exit code 2 for error-severity findings.
func TestSecurityScanner_GetExitCode(t *testing.T) {
	scanner := NewSecurityScanner()

	t.Run("returns 0 for no errors", func(t *testing.T) {
		result := &ScanResult{Scanned: true, ErrorCount: 0}
		exitCode := scanner.GetExitCode(result)
		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}
	})

	t.Run("returns 2 for errors per REQ-HOOK-131", func(t *testing.T) {
		result := &ScanResult{
			Scanned:    true,
			ErrorCount: 1,
			Findings:   []Finding{{Severity: SeverityError}},
		}
		exitCode := scanner.GetExitCode(result)
		if exitCode != 2 {
			t.Errorf("expected exit code 2, got %d", exitCode)
		}
	})

	t.Run("returns 0 for warnings only", func(t *testing.T) {
		result := &ScanResult{
			Scanned:      true,
			WarningCount: 5,
			Findings:     []Finding{{Severity: SeverityWarning}},
		}
		exitCode := scanner.GetExitCode(result)
		if exitCode != 0 {
			t.Errorf("expected exit code 0 for warnings, got %d", exitCode)
		}
	})
}

// TestSecurityScanner_ScanFiles_EmptyList verifies empty file list handling.
func TestSecurityScanner_ScanFiles_EmptyList(t *testing.T) {
	scanner := NewSecurityScanner()

	t.Run("handles empty file list", func(t *testing.T) {
		ctx := context.Background()
		results, err := scanner.ScanFiles(ctx, []string{}, "")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})
}
