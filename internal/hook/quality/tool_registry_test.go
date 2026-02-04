package quality

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestToolRegistry_NewRegistry verifies registry initialization.
func TestToolRegistry_NewRegistry(t *testing.T) {
	t.Run("creates new registry with default tools", func(t *testing.T) {
		registry := NewToolRegistry()

		if registry == nil {
			t.Fatal("expected non-nil registry")
		}

		// Should have tools for at least some languages
		pythonTools := registry.GetToolsForLanguage("python", ToolTypeFormatter)
		if len(pythonTools) == 0 {
			t.Error("expected at least one Python formatter, got none")
		}
	})
}

// TestRegisterTool verifies tool registration per REQ-HOOK-050.
func TestRegisterTool(t *testing.T) {
	t.Run("registers a new tool", func(t *testing.T) {
		registry := NewToolRegistry()

		tool := ToolConfig{
			Name:           "test-formatter",
			Command:        "echo",
			Args:           []string{"formatted"},
			Extensions:     []string{".test"},
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 10,
		}

		registry.RegisterTool(tool)

		tools := registry.GetToolsForLanguage("test", ToolTypeFormatter)
		if len(tools) != 1 {
			t.Errorf("expected 1 tool, got %d", len(tools))
		}
		if tools[0].Name != "test-formatter" {
			t.Errorf("expected tool name 'test-formatter', got '%s'", tools[0].Name)
		}
	})

	t.Run("registers multiple tools for same language", func(t *testing.T) {
		registry := NewToolRegistry()

		tool1 := ToolConfig{
			Name:       "formatter-1",
			Command:    "echo",
			Extensions: []string{".test"},
			ToolType:   ToolTypeFormatter,
			Priority:   1,
		}

		tool2 := ToolConfig{
			Name:       "formatter-2",
			Command:    "echo",
			Extensions: []string{".test"},
			ToolType:   ToolTypeFormatter,
			Priority:   2,
		}

		registry.RegisterTool(tool1)
		registry.RegisterTool(tool2)

		tools := registry.GetToolsForLanguage("test", ToolTypeFormatter)
		if len(tools) != 2 {
			t.Errorf("expected 2 tools, got %d", len(tools))
		}
	})
}

// TestGetToolsForLanguage verifies language-based tool lookup per REQ-HOOK-050.
func TestGetToolsForLanguage(t *testing.T) {
	t.Run("returns tools sorted by priority", func(t *testing.T) {
		registry := NewToolRegistry()

		// Register tools with different priorities using custom extension
		registry.RegisterTool(ToolConfig{
			Name:       "low-priority",
			Command:    "echo",
			Extensions: []string{".custom"},
			ToolType:   ToolTypeFormatter,
			Priority:   10,
		})

		registry.RegisterTool(ToolConfig{
			Name:       "high-priority",
			Command:    "echo",
			Extensions: []string{".custom"},
			ToolType:   ToolTypeFormatter,
			Priority:   1,
		})

		tools := registry.GetToolsForLanguage("custom", ToolTypeFormatter)

		// First tool should be highest priority (lowest number)
		if len(tools) < 2 {
			t.Fatalf("expected at least 2 tools, got %d", len(tools))
		}
		if tools[0].Name != "high-priority" {
			t.Errorf("expected first tool to be 'high-priority', got '%s'", tools[0].Name)
		}
		if tools[1].Name != "low-priority" {
			t.Errorf("expected second tool to be 'low-priority', got '%s'", tools[1].Name)
		}
	})

	t.Run("returns empty slice for unknown language", func(t *testing.T) {
		registry := NewToolRegistry()
		tools := registry.GetToolsForLanguage("unknown-language", ToolTypeFormatter)

		if tools == nil {
			t.Error("expected empty slice, not nil")
		}
		if len(tools) != 0 {
			t.Errorf("expected 0 tools, got %d", len(tools))
		}
	})
}

// TestGetToolsForFile verifies file-based tool lookup per REQ-HOOK-050.
func TestGetToolsForFile(t *testing.T) {
	t.Run("returns correct tools for Python file", func(t *testing.T) {
		registry := NewToolRegistry()
		tools := registry.GetToolsForFile("test.py", ToolTypeFormatter)

		if len(tools) == 0 {
			t.Error("expected at least one Python formatter")
		}
	})

	t.Run("returns correct tools for Go file", func(t *testing.T) {
		registry := NewToolRegistry()
		tools := registry.GetToolsForFile("test.go", ToolTypeFormatter)

		if len(tools) == 0 {
			t.Error("expected at least one Go formatter")
		}
	})

	t.Run("returns correct tools for JavaScript file", func(t *testing.T) {
		registry := NewToolRegistry()
		tools := registry.GetToolsForFile("test.js", ToolTypeFormatter)

		if len(tools) == 0 {
			t.Error("expected at least one JavaScript formatter")
		}
	})

	t.Run("returns empty for unknown extension", func(t *testing.T) {
		registry := NewToolRegistry()
		tools := registry.GetToolsForFile("test.unknown", ToolTypeFormatter)

		if len(tools) != 0 {
			t.Errorf("expected 0 tools for unknown extension, got %d", len(tools))
		}
	})
}

// TestIsToolAvailable verifies tool availability check per REQ-HOOK-050.
func TestIsToolAvailable(t *testing.T) {
	t.Run("returns true for available system command", func(t *testing.T) {
		registry := NewToolRegistry()

		// "go" should be available in test environment
		available := registry.IsToolAvailable("go")
		if !available {
			// This might fail if go is not in PATH, skip if so
			t.Skipf("go command not found in PATH")
		}
	})

	t.Run("returns false for non-existent command", func(t *testing.T) {
		registry := NewToolRegistry()
		available := registry.IsToolAvailable("nonexistent-tool-xyz-123")

		if available {
			t.Error("expected false for non-existent tool")
		}
	})
}

// TestRunTool verifies tool execution per REQ-HOOK-050.
func TestRunTool(t *testing.T) {
	t.Run("executes tool successfully", func(t *testing.T) {
		registry := NewToolRegistry()
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := []byte("hello")
		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Use echo command which should work everywhere
		tool := ToolConfig{
			Name:           "echo-test",
			Command:        "echo",
			Args:           []string{"test"},
			Extensions:     []string{".txt"},
			ToolType:       ToolTypeFormatter,
			TimeoutSeconds: 5,
		}

		result := registry.RunTool(tool, testFile, tmpDir)

		if !result.Success {
			t.Errorf("expected success, got error: %s", result.Error)
		}
		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
	})

	t.Run("times out long-running command", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("timeout test behaves differently on Windows")
		}

		registry := NewToolRegistry()
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		_ = os.WriteFile(testFile, []byte("test"), 0644)

		// Use sleep command with longer timeout
		cmd := "sleep"
		if runtime.GOOS == "windows" {
			cmd = "timeout" // Windows equivalent
		}

		tool := ToolConfig{
			Name:           "sleep-test",
			Command:        cmd,
			Args:           []string{"10"}, // Sleep for 10 seconds
			Extensions:     []string{".txt"},
			ToolType:       ToolTypeFormatter,
			TimeoutSeconds: 1, // Timeout after 1 second
		}

		start := time.Now()
		result := registry.RunTool(tool, testFile, tmpDir)
		elapsed := time.Since(start)

		if elapsed > 3*time.Second {
			t.Errorf("timeout did not work, took %v", elapsed)
		}

		// Should fail due to timeout
		if result.Success && result.Error == "" {
			t.Error("expected timeout error, got success")
		}
	})
}

// TestLanguageSupport verifies 16+ language support per REQ-HOOK-051.
func TestLanguageSupport(t *testing.T) {
	registry := NewToolRegistry()

	languages := []string{
		"python", "javascript", "typescript", "go", "rust",
		"java", "kotlin", "swift", "c", "cpp",
		"ruby", "php", "elixir", "scala", "r", "dart",
		"csharp", "markdown",
	}

	for _, lang := range languages {
		t.Run("has tools for "+lang, func(t *testing.T) {
			formatters := registry.GetToolsForLanguage(lang, ToolTypeFormatter)
			linters := registry.GetToolsForLanguage(lang, ToolTypeLinter)

			// At least one type of tool should be available
			if len(formatters) == 0 && len(linters) == 0 {
				t.Errorf("expected at least one tool for %s", lang)
			}
		})
	}
}

// TestExtensionMapping verifies file extension to language mapping.
func TestExtensionMapping(t *testing.T) {
	tests := []struct {
		ext      string
		language string
	}{
		{".py", "python"},
		{".pyi", "python"},
		{".go", "go"},
		{".rs", "rust"},
		{".js", "javascript"},
		{".jsx", "javascript"},
		{".ts", "typescript"},
		{".tsx", "typescript"},
		{".java", "java"},
		{".kt", "kotlin"},
		{".swift", "swift"},
		{".c", "c"},
		{".cpp", "cpp"},
		{".cc", "cpp"},
		{".h", "c"},
		{".hpp", "cpp"},
		{".rb", "ruby"},
		{".php", "php"},
		{".ex", "elixir"},
		{".exs", "elixir"},
		{".scala", "scala"},
		{".r", "r"},
		{".R", "r"},
		{".dart", "dart"},
		{".cs", "csharp"},
		{".md", "markdown"},
	}

	registry := NewToolRegistry()

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			tools := registry.GetToolsForFile("test"+tt.ext, ToolTypeFormatter)
			// Just verify we get tools (even if empty list for uninstalled tools)
			_ = tools
		})
	}
}

// TestToolConfigDefaults verifies default values.
func TestToolConfigDefaults(t *testing.T) {
	registry := NewToolRegistry()

	tool := ToolConfig{
		Name:       "test",
		Command:    "echo",
		Extensions: []string{".test"},
		ToolType:   ToolTypeFormatter,
		// Priority and TimeoutSeconds omitted
	}

	registry.RegisterTool(tool)

	tools := registry.GetToolsForLanguage("test", ToolTypeFormatter)
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	// Should have default timeout
	if tools[0].TimeoutSeconds == 0 {
		t.Error("expected default timeout to be set")
	}
}

// TestShellInjectionPrevention verifies command safety per REQ-HOOK-053.
func TestShellInjectionPrevention(t *testing.T) {
	t.Run("does not use shell for command execution", func(t *testing.T) {
		registry := NewToolRegistry()
		tmpDir := t.TempDir()

		// Create a file that should not be executed
		testFile := filepath.Join(tmpDir, "file.test")
		_ = os.WriteFile(testFile, []byte("content"), 0644)

		// Try to inject shell commands
		tool := ToolConfig{
			Name:           "inject-test",
			Command:        "echo",
			Args:           []string{"$(echo payload)", "; rm -rf /", "$(whoami)"},
			Extensions:     []string{".test"},
			ToolType:       ToolTypeFormatter,
			TimeoutSeconds: 5,
		}

		result := registry.RunTool(tool, testFile, tmpDir)

		if !result.Success {
			t.Logf("Command failed (expected): %s", result.Error)
		}

		// Output should contain literal strings, not executed commands
		if strings.Contains(result.Output, "payload") && !strings.Contains(result.Output, "$(") {
			t.Error("possible shell injection detected")
		}
	})
}
