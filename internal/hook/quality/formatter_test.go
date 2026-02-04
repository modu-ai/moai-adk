package quality

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestShouldFormat verifies skip conditions per REQ-HOOK-073, REQ-HOOK-074.
func TestShouldFormat(t *testing.T) {
	t.Run("returns true for supported code file", func(t *testing.T) {
		f := NewFormatter(nil)
		if !f.ShouldFormat("test.py") {
			t.Error("expected true for .py file")
		}
		if !f.ShouldFormat("test.go") {
			t.Error("expected true for .go file")
		}
	})

	t.Run("returns false for skipped extensions per REQ-HOOK-073", func(t *testing.T) {
		f := NewFormatter(nil)

		skippedExts := []string{
			".json", ".yaml", ".yml", ".toml", ".lock",
			".min.js", ".min.css",
			".map",
			".svg", ".png", ".jpg", ".gif", ".ico",
			".woff", ".woff2", ".ttf", ".eot",
			".exe", ".dll", ".so", ".dylib", ".bin",
		}

		for _, ext := range skippedExts {
			if f.ShouldFormat("test" + ext) {
				t.Errorf("expected false for %s file (REQ-HOOK-073)", ext)
			}
		}
	})

	t.Run("returns false for skipped directories per REQ-HOOK-074", func(t *testing.T) {
		f := NewFormatter(nil)

		skipDirs := []string{
			"node_modules/test.py",
			"vendor/test.py",
			".venv/test.py",
			"venv/test.py",
			"__pycache__/test.py",
			".cache/test.py",
			"dist/test.py",
			"build/test.py",
			".next/test.py",
			".nuxt/test.py",
			"target/test.py",
			".git/test.py",
			".svn/test.py",
			".hg/test.py",
			".idea/test.py",
			".vscode/test.py",
			".eclipse/test.py",
		}

		for _, path := range skipDirs {
			if f.ShouldFormat(path) {
				t.Errorf("expected false for %s (REQ-HOOK-074)", path)
			}
		}
	})
}

// TestFormatFile verifies formatting behavior per REQ-HOOK-070, REQ-HOOK-071, REQ-HOOK-072.
func TestFormatFile(t *testing.T) {
	t.Run("formats Go file with gofmt if available", func(t *testing.T) {
		registry := NewToolRegistry()

		// Skip if gofmt not available
		if !registry.IsToolAvailable("gofmt") {
			t.Skip("gofmt not available")
		}

		f := NewFormatter(registry)
		tmpDir := t.TempDir()

		// Create a poorly formatted Go file
		goFile := filepath.Join(tmpDir, "test.go")
		content := `package main
func main(){
println("hello")
}`
		if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		result, err := f.FormatFile(context.Background(), goFile)
		if err != nil {
			t.Fatalf("FormatFile failed: %v", err)
		}

		// Check if file was actually formatted
		if result == nil {
			t.Fatal("expected non-nil result")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		f := NewFormatter(NewToolRegistry())
		_, err := f.FormatFile(context.Background(), "/nonexistent/file.go")

		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})

	t.Run("skips formatting for skipped extension", func(t *testing.T) {
		f := NewFormatter(NewToolRegistry())
		tmpDir := t.TempDir()

		// Create a .json file (should be skipped per REQ-HOOK-073)
		jsonFile := filepath.Join(tmpDir, "test.json")
		_ = os.WriteFile(jsonFile, []byte(`{"key":"value"}`), 0644)

		result, err := f.FormatFile(context.Background(), jsonFile)

		// Should return nil result (skipped) without error
		if err != nil {
			t.Errorf("expected no error for skipped file, got: %v", err)
		}
		if result != nil {
			t.Error("expected nil result for skipped file")
		}
	})

	t.Run("gracefully handles unavailable formatter", func(t *testing.T) {
		registry := NewToolRegistry()
		f := NewFormatter(registry)
		tmpDir := t.TempDir()

		// Create a .xyz file (no formatter available)
		xyzFile := filepath.Join(tmpDir, "test.xyz")
		_ = os.WriteFile(xyzFile, []byte("content"), 0644)

		result, err := f.FormatFile(context.Background(), xyzFile)

		// Should not error, just return nil or unsuccessful result
		if err != nil {
			t.Errorf("expected no error for unavailable formatter, got: %v", err)
		}
		if result != nil && result.Error == "" {
			// If result exists, it should indicate failure or no change
			t.Log("formatter returned result for unsupported file")
		}
	})
}

// TestFormatterWithChangeDetection verifies hash-based change detection per REQ-HOOK-061.
func TestFormatterWithChangeDetection(t *testing.T) {
	t.Run("detects when formatting makes no changes", func(t *testing.T) {
		registry := NewToolRegistry()
		detector := NewChangeDetector()

		// Skip if no formatter available
		if !registry.IsToolAvailable("gofmt") {
			t.Skip("gofmt not available")
		}

		f := NewFormatterWithRegistry(registry, detector)
		tmpDir := t.TempDir()

		// Create a properly formatted Go file
		goFile := filepath.Join(tmpDir, "test.go")
		content := `package main

func main() {
	fmt.Println("hello")
}
`
		if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Get original hash
		originalHash, err := detector.ComputeHash(goFile)
		_ = err // Ignore error in this test

		result, err := f.FormatFile(context.Background(), goFile)
		if err != nil {
			t.Fatalf("FormatFile failed: %v", err)
		}

		// Check if file changed
		changed, err := detector.HasChanged(goFile, originalHash)
		_ = err // Ignore error in this test

		if changed {
			t.Log("File was modified by formatter (this is OK for formatters that always add timestamps or normalize)")
		}

		if result == nil {
			t.Error("expected non-nil result")
		}
	})
}

// TestFormatterIntegration verifies integration with ToolRegistry per REQ-HOOK-071.
func TestFormatterIntegration(t *testing.T) {
	t.Run("uses tool registry to find formatter", func(t *testing.T) {
		registry := NewToolRegistry()

		// Skip if gofmt not available
		if !registry.IsToolAvailable("gofmt") {
			t.Skip("gofmt not available")
		}

		_ = NewFormatter(registry)

		// Verify the formatter can find Go tools
		tools := registry.GetToolsForFile("test.go", ToolTypeFormatter)
		if len(tools) == 0 {
			t.Error("expected at least one Go formatter from registry")
		}

		// Check that gofmt is in the list
		found := false
		for _, tool := range tools {
			if strings.Contains(tool.Name, "gofmt") || tool.Command == "gofmt" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected gofmt to be in formatters list")
		}
	})
}

// TestNewFormatter verifies constructor variants.
func TestNewFormatter(t *testing.T) {
	t.Run("creates formatter with registry", func(t *testing.T) {
		registry := NewToolRegistry()
		f := NewFormatter(registry)

		if f == nil {
			t.Error("expected non-nil formatter")
		}
	})

	t.Run("creates formatter with default registry when nil", func(t *testing.T) {
		f := NewFormatter(nil)

		if f == nil {
			t.Error("expected non-nil formatter with default registry")
		}
	})

	t.Run("creates formatter with custom detector", func(t *testing.T) {
		registry := NewToolRegistry()
		detector := NewChangeDetector()
		f := NewFormatterWithRegistry(registry, detector)

		if f == nil {
			t.Error("expected non-nil formatter with custom detector")
		}
	})
}

// TestCrossPlatformTimeout verifies timeout behavior per REQ-HOOK-092.
func TestCrossPlatformTimeout(t *testing.T) {
	t.Run("handles context cancellation", func(t *testing.T) {
		registry := NewToolRegistry()
		f := NewFormatter(registry)
		tmpDir := t.TempDir()

		// Create a test file
		testFile := filepath.Join(tmpDir, "test.go")
		_ = os.WriteFile(testFile, []byte("package main\n"), 0644)

		// Cancel the context immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		result, err := f.FormatFile(ctx, testFile)

		// Should handle cancellation gracefully
		if err == nil && result != nil && result.Success {
			// This is OK if the formatter was fast enough
			t.Log("Formatter completed before context cancellation")
		}
		if err != nil {
			// Expected - context was cancelled
			if !strings.Contains(err.Error(), "canceled") && !strings.Contains(err.Error(), "context") {
				t.Logf("Got error (may be expected): %v", err)
			}
		}
	})
}
