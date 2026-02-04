package quality

import (
	"context"
	"strings"
	"testing"
)

func TestNewToolRegistry(t *testing.T) {
	t.Parallel()

	registry := NewToolRegistry()
	if registry == nil {
		t.Fatal("NewToolRegistry returned nil")
	}

	if !registry.initialized {
		t.Error("registry not initialized")
	}
}

func TestToolRegistry_GetLanguageForFile(t *testing.T) {
	t.Parallel()

	registry := NewToolRegistry()

	tests := []struct {
		name     string
		filePath string
		wantLang string
	}{
		{"Python file", "test.py", "python"},
		{"Python stub", "test.pyi", "python"},
		{"Go file", "main.go", "go"},
		{"JavaScript", "app.js", "javascript"},
		{"JSX", "App.jsx", "javascript"},
		{"TypeScript", "app.ts", "typescript"},
		{"TSX", "App.tsx", "typescript"},
		{"Rust", "main.rs", "rust"},
		{"Java", "Main.java", "java"},
		{"Kotlin", "Main.kt", "kotlin"},
		{"Swift", "main.swift", "swift"},
		{"C file", "main.c", "cpp"},
		{"C++ header", "header.hpp", "cpp"},
		{"Ruby", "script.rb", "ruby"},
		{"PHP", "index.php", "php"},
		{"Elixir", "main.ex", "elixir"},
		{"Scala", "Main.scala", "scala"},
		{"R script", "script.R", "r"},
		{"Dart", "main.dart", "dart"},
		{"C#", "Program.cs", "csharp"},
		{"Markdown", "README.md", "markdown"},
		{"Unknown", "unknown.xyz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.GetLanguageForFile(tt.filePath)
			if got != tt.wantLang {
				t.Errorf("GetLanguageForFile() = %q, want %q", got, tt.wantLang)
			}
		})
	}
}

func TestToolRegistry_GetToolsForLanguage(t *testing.T) {
	t.Parallel()

	registry := NewToolRegistry()

	// Test Python tools
	pythonFormatters := registry.GetToolsForLanguage("python", ToolTypeFormatter)
	if len(pythonFormatters) == 0 {
		t.Error("no Python formatters registered")
	}

	// Check priority ordering (ruff-format should be first)
	if len(pythonFormatters) > 0 && pythonFormatters[0].Name != "ruff-format" {
		t.Errorf("first Python formatter = %q, want ruff-format", pythonFormatters[0].Name)
	}

	// Test Go tools
	goFormatters := registry.GetToolsForLanguage("go", ToolTypeFormatter)
	if len(goFormatters) == 0 {
		t.Error("no Go formatters registered")
	}

	// Test TypeScript linters
	tsLinters := registry.GetToolsForLanguage("typescript", ToolTypeLinter)
	if len(tsLinters) == 0 {
		t.Error("no TypeScript linters registered")
	}
}

func TestToolRegistry_GetToolsForFile(t *testing.T) {
	t.Parallel()

	registry := NewToolRegistry()

	tests := []struct {
		name      string
		filePath  string
		toolType  ToolType
		wantCount int // Expected at least this many tools
	}{
		{"Python formatter", "test.py", ToolTypeFormatter, 1},
		{"Go formatter", "main.go", ToolTypeFormatter, 1},
		{"JS linter", "app.js", ToolTypeLinter, 1},
		{"Unknown file", "unknown.xyz", ToolTypeFormatter, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := registry.GetToolsForFile(tt.filePath, tt.toolType)
			if len(tools) < tt.wantCount {
				t.Errorf("GetToolsForFile() returned %d tools, want at least %d", len(tools), tt.wantCount)
			}
		})
	}
}

func TestToolRegistry_RegisterTool(t *testing.T) {
	t.Parallel()

	registry := NewToolRegistry()

	// Register a custom tool
	customTool := ToolConfig{
		Name:       "custom-formatter",
		Command:    "custom",
		Args:       []string{"--format"},
		Extensions: []string{".custom"},
		ToolType:   ToolTypeFormatter,
		Priority:   1,
	}

	registry.RegisterTool(customTool)

	// Verify it's registered
	tools := registry.GetToolsForFile("test.custom", ToolTypeFormatter)
	if len(tools) != 1 {
		t.Errorf("got %d tools, want 1", len(tools))
	}

	if len(tools) > 0 && tools[0].Name != "custom-formatter" {
		t.Errorf("tool name = %q, want custom-formatter", tools[0].Name)
	}
}

func TestToolRegistry_PriorityOrdering(t *testing.T) {
	t.Parallel()

	registry := NewToolRegistry()

	// Python formatters should be ordered by priority
	formatters := registry.GetToolsForLanguage("python", ToolTypeFormatter)

	// Verify priority ordering (lower number = higher priority)
	for i := 1; i < len(formatters); i++ {
		if formatters[i-1].Priority > formatters[i].Priority {
			t.Errorf("tools not ordered by priority: [%d].Priority=%d > [%d].Priority=%d",
				i-1, formatters[i-1].Priority, i, formatters[i].Priority)
		}
	}
}

func TestToolRegistry_RunTool_InvalidPath(t *testing.T) {
	t.Parallel()

	registry := NewToolRegistry()

	tool := ToolConfig{
		Name:       "test",
		Command:    "echo",
		Args:       []string{"hello"},
		Extensions: []string{".txt"},
		ToolType:   ToolTypeFormatter,
	}

	// Test null byte rejection
	result := registry.RunTool(context.Background(), tool, "test\x00.txt", "")
	if result.Success {
		t.Error("expected failure for null byte in path")
	}

	// Test shell metacharacter rejection
	result = registry.RunTool(context.Background(), tool, "test'.txt", "")
	if result.Success {
		t.Error("expected failure for shell metacharacter in path")
	}
}

func TestToolRegistry_IsToolAvailable(t *testing.T) {
	t.Parallel()

	registry := NewToolRegistry()

	// Test a tool that should exist (sh or echo should be available on most systems)
	available := registry.IsToolAvailable("sh")
	if !available {
		// sh might not be in PATH on all systems
		t.Skip("sh not available in PATH")
	}

	// Test a non-existent tool
	available = registry.IsToolAvailable("definitely-not-a-real-tool-xyz123")
	if available {
		t.Error("non-existent tool reported as available")
	}
}

func TestToolRegistry_Caching(t *testing.T) {
	t.Parallel()

	registry := NewToolRegistry()

	// First call checks availability
	_ = registry.IsToolAvailable("sh")

	// Second call should use cache
	_ = registry.IsToolAvailable("sh")

	// If cache wasn't working, this would have issues
	// Just verify no panics occur
}

func TestToolConfigDefaults(t *testing.T) {
	t.Parallel()

	tool := ToolConfig{
		Name:       "test",
		Command:    "test",
		Extensions: []string{".test"},
		ToolType:   ToolTypeFormatter,
	}

	// Verify default values
	if tool.Priority != 0 {
		t.Errorf("default Priority = %d, want 0", tool.Priority)
	}

	if tool.TimeoutSeconds != 0 {
		t.Errorf("default TimeoutSeconds = %d, want 0", tool.TimeoutSeconds)
	}
}

func TestToolRegistry_AllLanguagesRegistered(t *testing.T) {
	t.Parallel()

	registry := NewToolRegistry()

	expectedLanguages := []string{
		"python", "javascript", "typescript", "go", "rust",
		"java", "kotlin", "swift", "cpp", "ruby",
		"php", "elixir", "scala", "r", "dart", "csharp", "markdown",
	}

	for _, lang := range expectedLanguages {
		t.Run(lang, func(t *testing.T) {
			formatters := registry.GetToolsForLanguage(lang, ToolTypeFormatter)
			linters := registry.GetToolsForLanguage(lang, ToolTypeLinter)

			if len(formatters) == 0 && len(linters) == 0 {
				t.Errorf("language %q has no tools registered", lang)
			}
		})
	}
}

func TestEscapePathForCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{`simple`, `simple`},
		// ' is escaped to \'
		{`path'with'quotes`, `path\'with\'quotes`},
		// " is escaped to \"
		{`path"with"quotes`, `path\"with\"quotes`},
		// \ is escaped to \\
		{`path\with\backslashes`, `path\\with\\backslashes`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapePathForCode(tt.input)
			if got != tt.want {
				t.Errorf("escapePathForCode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSortByPriority(t *testing.T) {
	t.Parallel()

	tools := []ToolConfig{
		{Name: "tool3", Priority: 3},
		{Name: "tool1", Priority: 1},
		{Name: "tool2", Priority: 2},
		{Name: "tool0", Priority: 0},
	}

	sortByPriority(tools)

	if tools[0].Priority != 0 {
		t.Errorf("first tool priority = %d, want 0", tools[0].Priority)
	}
	if tools[1].Priority != 1 {
		t.Errorf("second tool priority = %d, want 1", tools[1].Priority)
	}
	if tools[2].Priority != 2 {
		t.Errorf("third tool priority = %d, want 2", tools[2].Priority)
	}
	if tools[3].Priority != 3 {
		t.Errorf("fourth tool priority = %d, want 3", tools[3].Priority)
	}
}

func TestToolRegistry_RustTools(t *testing.T) {
	t.Parallel()

	registry := NewToolRegistry()

	// Rust should have rustfmt and clippy
	formatters := registry.GetToolsForLanguage("rust", ToolTypeFormatter)
	if len(formatters) == 0 {
		t.Error("no Rust formatters")
	} else if formatters[0].Name != "rustfmt" {
		t.Errorf("first Rust formatter = %q, want rustfmt", formatters[0].Name)
	}

	linters := registry.GetToolsForLanguage("rust", ToolTypeLinter)
	if len(linters) == 0 {
		t.Error("no Rust linters")
	} else if !strings.Contains(linters[0].Name, "clippy") {
		t.Errorf("Rust linter name = %q, want clippy", linters[0].Name)
	}
}

func TestToolRegistry_TypeScriptTypeChecker(t *testing.T) {
	t.Parallel()

	registry := NewToolRegistry()

	typeCheckers := registry.GetToolsForLanguage("typescript", ToolTypeTypeChecker)
	if len(typeCheckers) == 0 {
		t.Error("no TypeScript type checkers")
	} else if typeCheckers[0].Name != "tsc" {
		t.Errorf("TypeScript type checker = %q, want tsc", typeCheckers[0].Name)
	}
}
