package tools

import (
	"context"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// NewToolRegistry tests
// ============================================================================

func TestNewToolRegistry(t *testing.T) {
	r := NewToolRegistry()

	if r == nil {
		t.Fatal("NewToolRegistry returned nil")
	}
	if len(r.tools) == 0 {
		t.Error("Expected tools to be populated")
	}
	if len(r.extensionMap) == 0 {
		t.Error("Expected extension map to be populated")
	}
	if r.toolCache == nil {
		t.Error("Expected tool cache to be initialized")
	}
}

func TestNewToolRegistry_RegistersAllLanguages(t *testing.T) {
	r := NewToolRegistry()

	expectedLanguages := []string{
		"python", "javascript", "typescript", "go", "rust",
		"java", "kotlin", "swift", "cpp", "ruby",
		"php", "elixir", "scala", "r", "dart",
		"csharp", "markdown", "yaml", "json", "shell", "lua",
	}

	for _, lang := range expectedLanguages {
		t.Run(lang, func(t *testing.T) {
			tools, ok := r.tools[lang]
			if !ok {
				t.Errorf("Expected language '%s' to be registered", lang)
				return
			}
			if len(tools) == 0 {
				t.Errorf("Expected tools for language '%s', got none", lang)
			}
		})
	}
}

// ============================================================================
// GetGlobalRegistry tests
// ============================================================================

func TestGetGlobalRegistry(t *testing.T) {
	r1 := GetGlobalRegistry()
	r2 := GetGlobalRegistry()

	if r1 == nil {
		t.Fatal("GetGlobalRegistry returned nil")
	}
	if r1 != r2 {
		t.Error("Expected GetGlobalRegistry to return the same instance (singleton)")
	}
}

// ============================================================================
// ToolType constant tests
// ============================================================================

func TestToolTypeConstants(t *testing.T) {
	tests := []struct {
		toolType ToolType
		expected string
	}{
		{ToolTypeFormatter, "formatter"},
		{ToolTypeLinter, "linter"},
		{ToolTypeTypeChecker, "type_checker"},
		{ToolTypeSecurityScan, "security_scanner"},
		{ToolTypeASTAnalyzer, "ast_analyzer"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.toolType) != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, string(tt.toolType))
			}
		})
	}
}

// ============================================================================
// GetLanguageForFile tests
// ============================================================================

func TestGetLanguageForFile(t *testing.T) {
	r := NewToolRegistry()

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		// Python
		{"Python .py", "main.py", "python"},
		{"Python .pyi", "types.pyi", "python"},
		// JavaScript
		{"JavaScript .js", "app.js", "javascript"},
		{"JavaScript .jsx", "Component.jsx", "javascript"},
		{"JavaScript .mjs", "module.mjs", "javascript"},
		{"JavaScript .cjs", "common.cjs", "javascript"},
		// TypeScript
		{"TypeScript .ts", "index.ts", "typescript"},
		{"TypeScript .tsx", "App.tsx", "typescript"},
		{"TypeScript .mts", "module.mts", "typescript"},
		{"TypeScript .cts", "common.cts", "typescript"},
		// Go
		{"Go .go", "main.go", "go"},
		// Rust
		{"Rust .rs", "lib.rs", "rust"},
		// Java
		{"Java .java", "Main.java", "java"},
		// Kotlin
		{"Kotlin .kt", "Main.kt", "kotlin"},
		{"Kotlin .kts", "build.gradle.kts", "kotlin"},
		// Swift
		{"Swift .swift", "ViewController.swift", "swift"},
		// C/C++
		{"C .c", "main.c", "cpp"},
		{"C++ .cpp", "main.cpp", "cpp"},
		{"C++ .cc", "main.cc", "cpp"},
		{"C++ .cxx", "main.cxx", "cpp"},
		{"C header .h", "header.h", "cpp"},
		{"C++ header .hpp", "header.hpp", "cpp"},
		{"C++ header .hxx", "header.hxx", "cpp"},
		// Ruby
		{"Ruby .rb", "app.rb", "ruby"},
		{"Ruby .rake", "Rakefile.rake", "ruby"},
		{"Ruby .gemspec", "mygem.gemspec", "ruby"},
		// PHP
		{"PHP .php", "index.php", "php"},
		// Elixir
		{"Elixir .ex", "app.ex", "elixir"},
		{"Elixir .exs", "test_helper.exs", "elixir"},
		// Scala
		{"Scala .scala", "Main.scala", "scala"},
		{"Scala .sc", "worksheet.sc", "scala"},
		// R
		{"R .r", "analysis.r", "r"},
		{"R .R", "Analysis.R", "r"},
		{"R .Rmd", "report.Rmd", "r"},
		// Dart
		{"Dart .dart", "main.dart", "dart"},
		// C#
		{"CSharp .cs", "Program.cs", "csharp"},
		// Markdown
		{"Markdown .md", "README.md", "markdown"},
		{"Markdown .mdx", "page.mdx", "markdown"},
		{"Markdown .markdown", "doc.markdown", "markdown"},
		// YAML
		{"YAML .yaml", "config.yaml", "yaml"},
		{"YAML .yml", "docker-compose.yml", "yaml"},
		// JSON
		{"JSON .json", "package.json", "json"},
		// Shell
		{"Shell .sh", "install.sh", "shell"},
		{"Shell .bash", "setup.bash", "shell"},
		{"Shell .zsh", "config.zsh", "shell"},
		// Lua
		{"Lua .lua", "init.lua", "lua"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang := r.GetLanguageForFile(tt.filePath)
			if lang != tt.expected {
				t.Errorf("Expected '%s' for '%s', got '%s'", tt.expected, tt.filePath, lang)
			}
		})
	}
}

func TestGetLanguageForFile_CaseInsensitive(t *testing.T) {
	r := NewToolRegistry()

	// Extension detection uses strings.ToLower so uppercase should work
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{"Uppercase .PY", "main.PY", "python"},
		{"Uppercase .GO", "main.GO", "go"},
		{"Uppercase .JS", "app.JS", "javascript"},
		{"Uppercase .TS", "index.TS", "typescript"},
		{"Mixed case .Py", "script.Py", "python"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang := r.GetLanguageForFile(tt.filePath)
			if lang != tt.expected {
				t.Errorf("Expected '%s' for '%s', got '%s'", tt.expected, tt.filePath, lang)
			}
		})
	}
}

func TestGetLanguageForFile_UnknownExtension(t *testing.T) {
	r := NewToolRegistry()

	tests := []struct {
		name     string
		filePath string
	}{
		{"No extension", "Makefile"},
		{"Unknown extension", "data.xyz"},
		{"Binary", "program.bin"},
		{"Image", "photo.jpg"},
		{"Empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang := r.GetLanguageForFile(tt.filePath)
			if lang != "" {
				t.Errorf("Expected empty string for unknown extension '%s', got '%s'", tt.filePath, lang)
			}
		})
	}
}

func TestGetLanguageForFile_NestedPaths(t *testing.T) {
	r := NewToolRegistry()

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{"Nested Python", "src/app/views.py", "python"},
		{"Nested TypeScript", "packages/core/src/index.ts", "typescript"},
		{"Nested Go", "internal/hooks/security/guards.go", "go"},
		{"Absolute path", "/home/user/project/main.rs", "rust"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang := r.GetLanguageForFile(tt.filePath)
			if lang != tt.expected {
				t.Errorf("Expected '%s' for '%s', got '%s'", tt.expected, tt.filePath, lang)
			}
		})
	}
}

// ============================================================================
// GetToolsForLanguage tests
// ============================================================================

func TestGetToolsForLanguage_UnknownLanguage(t *testing.T) {
	r := NewToolRegistry()

	tools := r.GetToolsForLanguage("brainfuck", "")
	if tools != nil {
		t.Errorf("Expected nil for unknown language, got %d tools", len(tools))
	}
}

func TestGetToolsForLanguage_WithPreCachedTools(t *testing.T) {
	r := NewToolRegistry()

	// Pre-populate tool cache to avoid deadlock from nested lock acquisition.
	// When GetToolsForLanguage holds a RLock and calls IsToolAvailable,
	// a cache miss would try to acquire a write Lock, causing deadlock.
	func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		for _, langTools := range r.tools {
			for _, tool := range langTools {
				r.toolCache[tool.Name] = true
			}
		}
	}()

	// Test: all tools for Python (no type filter)
	pythonTools := r.GetToolsForLanguage("python", "")
	if len(pythonTools) == 0 {
		t.Error("Expected Python tools when all are cached as available")
	}

	// Test: only formatters for Python
	formatters := r.GetToolsForLanguage("python", ToolTypeFormatter)
	for _, tool := range formatters {
		if tool.ToolType != ToolTypeFormatter {
			t.Errorf("Expected formatter tool type, got '%s' for tool '%s'", tool.ToolType, tool.Name)
		}
	}

	// Test: only linters for Go
	linters := r.GetToolsForLanguage("go", ToolTypeLinter)
	for _, tool := range linters {
		if tool.ToolType != ToolTypeLinter {
			t.Errorf("Expected linter tool type, got '%s' for tool '%s'", tool.ToolType, tool.Name)
		}
	}
}

func TestGetToolsForLanguage_NoneAvailable(t *testing.T) {
	r := NewToolRegistry()

	// Pre-populate all tool caches as unavailable
	r.mu.Lock()
	for _, langTools := range r.tools {
		for _, tool := range langTools {
			r.toolCache[tool.Name] = false
		}
	}
	r.mu.Unlock()

	tools := r.GetToolsForLanguage("python", "")
	if len(tools) != 0 {
		t.Errorf("Expected no tools when all are cached as unavailable, got %d", len(tools))
	}
}

// ============================================================================
// GetToolsForFile tests
// ============================================================================

func TestGetToolsForFile_UnknownExtension(t *testing.T) {
	r := NewToolRegistry()

	tools := r.GetToolsForFile("data.xyz", "")
	if tools != nil {
		t.Errorf("Expected nil for unknown extension, got %d tools", len(tools))
	}
}

func TestGetToolsForFile_KnownExtension(t *testing.T) {
	r := NewToolRegistry()

	// Pre-populate cache so GetToolsForLanguage works safely
	r.mu.Lock()
	for _, langTools := range r.tools {
		for _, tool := range langTools {
			r.toolCache[tool.Name] = true
		}
	}
	r.mu.Unlock()

	tools := r.GetToolsForFile("main.py", ToolTypeFormatter)
	if len(tools) == 0 {
		t.Error("Expected formatter tools for .py file")
	}
	for _, tool := range tools {
		if tool.ToolType != ToolTypeFormatter {
			t.Errorf("Expected formatter type, got '%s'", tool.ToolType)
		}
	}
}

// ============================================================================
// IsToolAvailable tests
// ============================================================================

func TestIsToolAvailable_NotInRegistry(t *testing.T) {
	r := NewToolRegistry()

	available := r.IsToolAvailable("nonexistent-tool-xyz-123")
	if available {
		t.Error("Expected false for tool not in registry")
	}
}

func TestIsToolAvailable_CacheHit(t *testing.T) {
	r := NewToolRegistry()

	// Manually populate cache
	r.mu.Lock()
	r.toolCache["gofmt"] = true
	r.mu.Unlock()

	available := r.IsToolAvailable("gofmt")
	if !available {
		t.Error("Expected true for cached available tool")
	}

	// Test cached as unavailable
	r.mu.Lock()
	r.toolCache["gofmt"] = false
	r.mu.Unlock()

	available = r.IsToolAvailable("gofmt")
	if available {
		t.Error("Expected false for cached unavailable tool")
	}
}

func TestIsToolAvailable_CacheMiss_RealTool(t *testing.T) {
	r := NewToolRegistry()

	// "gofmt" should be available since we are running Go tests
	available := r.IsToolAvailable("gofmt")
	if !available {
		t.Log("gofmt not found on system, skipping availability check")
		return
	}

	// Verify it was cached
	r.mu.RLock()
	cached, ok := r.toolCache["gofmt"]
	r.mu.RUnlock()

	if !ok {
		t.Error("Expected gofmt to be cached after availability check")
	}
	if !cached {
		t.Error("Expected gofmt to be cached as available")
	}
}

func TestIsToolAvailable_CacheMiss_NonexistentCommand(t *testing.T) {
	r := NewToolRegistry()

	// Add a tool with a command that does not exist on the system
	r.mu.Lock()
	r.tools["test-lang"] = []*ToolConfig{
		{
			Name:    "fake-tool-abc123",
			Command: "nonexistent-command-abc123-xyz789",
		},
	}
	r.mu.Unlock()

	available := r.IsToolAvailable("fake-tool-abc123")
	if available {
		t.Error("Expected false for tool with nonexistent command")
	}

	// Verify it was cached as unavailable
	r.mu.RLock()
	cached, ok := r.toolCache["fake-tool-abc123"]
	r.mu.RUnlock()

	if !ok {
		t.Error("Expected tool to be cached after availability check")
	}
	if cached {
		t.Error("Expected tool to be cached as unavailable")
	}
}

// ============================================================================
// RunTool tests
// ============================================================================

func TestRunTool_EndPosition(t *testing.T) {
	r := NewToolRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tool := &ToolConfig{
		Name:             "test-echo-end",
		Command:          "echo",
		Args:             []string{"hello"},
		FileArgsPosition: "end",
	}

	result, err := r.RunTool(ctx, tool, "test.txt")
	if err != nil {
		t.Fatalf("RunTool failed: %v", err)
	}
	if !result.Success {
		t.Error("Expected success")
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
	if result.ToolName != "test-echo-end" {
		t.Errorf("Expected tool name 'test-echo-end', got '%s'", result.ToolName)
	}
	// Output should contain both "hello" and "test.txt"
	if !strings.Contains(result.Output, "hello") || !strings.Contains(result.Output, "test.txt") {
		t.Errorf("Expected output to contain 'hello test.txt', got '%s'", result.Output)
	}
}

func TestRunTool_StartPosition_WithArgs(t *testing.T) {
	r := NewToolRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tool := &ToolConfig{
		Name:             "test-echo-start",
		Command:          "echo",
		Args:             []string{"suffix"},
		FileArgsPosition: "start",
	}

	result, err := r.RunTool(ctx, tool, "myfile.go")
	if err != nil {
		t.Fatalf("RunTool failed: %v", err)
	}
	if !result.Success {
		t.Error("Expected success")
	}
	// File should appear before other args: "myfile.go suffix"
	output := strings.TrimSpace(result.Output)
	if output != "myfile.go suffix" {
		t.Errorf("Expected 'myfile.go suffix', got '%s'", output)
	}
}

func TestRunTool_StartPosition_NoArgs(t *testing.T) {
	r := NewToolRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tool := &ToolConfig{
		Name:             "test-echo-start-noargs",
		Command:          "echo",
		Args:             []string{},
		FileArgsPosition: "start",
	}

	result, err := r.RunTool(ctx, tool, "myfile.go")
	if err != nil {
		t.Fatalf("RunTool failed: %v", err)
	}
	if !result.Success {
		t.Error("Expected success")
	}
	output := strings.TrimSpace(result.Output)
	if output != "myfile.go" {
		t.Errorf("Expected 'myfile.go', got '%s'", output)
	}
}

func TestRunTool_ReplacePosition(t *testing.T) {
	r := NewToolRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tool := &ToolConfig{
		Name:             "test-echo-replace",
		Command:          "echo",
		Args:             []string{"before", "style_file", "after"},
		FileArgsPosition: "replace:style_file",
	}

	result, err := r.RunTool(ctx, tool, "/path/to/file.R")
	if err != nil {
		t.Fatalf("RunTool failed: %v", err)
	}
	if !result.Success {
		t.Error("Expected success")
	}
	// The placeholder should be replaced with style_file('/path/to/file.R')
	if !strings.Contains(result.Output, "style_file('/path/to/file.R')") {
		t.Errorf("Expected replaced placeholder in output, got '%s'", result.Output)
	}
	if !strings.Contains(result.Output, "before") || !strings.Contains(result.Output, "after") {
		t.Errorf("Expected surrounding args preserved, got '%s'", result.Output)
	}
}

func TestRunTool_ReplacePosition_EscapesSingleQuotes(t *testing.T) {
	r := NewToolRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tool := &ToolConfig{
		Name:             "test-echo-replace-escape",
		Command:          "echo",
		Args:             []string{"style_file"},
		FileArgsPosition: "replace:style_file",
	}

	result, err := r.RunTool(ctx, tool, "/path/to/file's.R")
	if err != nil {
		t.Fatalf("RunTool failed: %v", err)
	}
	// Single quotes in path should be escaped
	if !strings.Contains(result.Output, "style_file('/path/to/file\\'s.R')") {
		t.Errorf("Expected escaped single quote in output, got '%s'", result.Output)
	}
}

func TestRunTool_ReplacePosition_PlaceholderNotFound(t *testing.T) {
	r := NewToolRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tool := &ToolConfig{
		Name:             "test-echo-replace-missing",
		Command:          "echo",
		Args:             []string{"no_match_here"},
		FileArgsPosition: "replace:missing_placeholder",
	}

	result, err := r.RunTool(ctx, tool, "file.txt")
	if err != nil {
		t.Fatalf("RunTool failed: %v", err)
	}
	// Placeholder not found, so args remain unchanged (file not added)
	output := strings.TrimSpace(result.Output)
	if output != "no_match_here" {
		t.Errorf("Expected 'no_match_here' (unchanged), got '%s'", output)
	}
}

func TestRunTool_DefaultPosition(t *testing.T) {
	r := NewToolRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tool := &ToolConfig{
		Name:             "test-echo-default",
		Command:          "echo",
		Args:             []string{"hello"},
		FileArgsPosition: "unknown-position",
	}

	result, err := r.RunTool(ctx, tool, "test.txt")
	if err != nil {
		t.Fatalf("RunTool failed: %v", err)
	}
	// Default case: append file at end (same as "end")
	output := strings.TrimSpace(result.Output)
	if output != "hello test.txt" {
		t.Errorf("Expected 'hello test.txt', got '%s'", output)
	}
}

func TestRunTool_CommandError(t *testing.T) {
	r := NewToolRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tool := &ToolConfig{
		Name:             "test-nonexistent",
		Command:          "nonexistent-command-abc123",
		Args:             []string{},
		FileArgsPosition: "end",
	}

	result, err := r.RunTool(ctx, tool, "test.txt")
	if err == nil {
		t.Fatal("Expected error for nonexistent command")
	}
	if result.Success {
		t.Error("Expected failure")
	}
	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", result.ExitCode)
	}
	if result.Error == "" {
		t.Error("Expected error message")
	}
	if result.ToolName != "test-nonexistent" {
		t.Errorf("Expected tool name 'test-nonexistent', got '%s'", result.ToolName)
	}
}

func TestRunTool_CommandFailure(t *testing.T) {
	r := NewToolRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tool := &ToolConfig{
		Name:             "test-false",
		Command:          "false",
		Args:             []string{},
		FileArgsPosition: "end",
	}

	result, err := r.RunTool(ctx, tool, "test.txt")
	if err == nil {
		t.Fatal("Expected error for command that exits with non-zero")
	}
	if result.Success {
		t.Error("Expected failure")
	}
	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", result.ExitCode)
	}
}

// ============================================================================
// ToolConfig field tests
// ============================================================================

func TestToolConfig_Fields(t *testing.T) {
	config := &ToolConfig{
		Name:             "test-tool",
		Command:          "test-cmd",
		Args:             []string{"--flag"},
		FileArgsPosition: "end",
		CheckArgs:        []string{"--check"},
		FixArgs:          []string{"--fix"},
		Extensions:       []string{".test"},
		ToolType:         ToolTypeFormatter,
		Priority:         1,
		TimeoutSeconds:   30,
		RequiresConfig:   true,
		ConfigFiles:      []string{".testrc"},
	}

	if config.Name != "test-tool" {
		t.Errorf("Expected name 'test-tool', got '%s'", config.Name)
	}
	if config.Command != "test-cmd" {
		t.Errorf("Expected command 'test-cmd', got '%s'", config.Command)
	}
	if len(config.Args) != 1 || config.Args[0] != "--flag" {
		t.Errorf("Expected args ['--flag'], got %v", config.Args)
	}
	if config.ToolType != ToolTypeFormatter {
		t.Errorf("Expected type 'formatter', got '%s'", config.ToolType)
	}
	if config.Priority != 1 {
		t.Errorf("Expected priority 1, got %d", config.Priority)
	}
	if config.TimeoutSeconds != 30 {
		t.Errorf("Expected timeout 30, got %d", config.TimeoutSeconds)
	}
	if !config.RequiresConfig {
		t.Error("Expected RequiresConfig to be true")
	}
	if len(config.ConfigFiles) != 1 || config.ConfigFiles[0] != ".testrc" {
		t.Errorf("Expected config files ['.testrc'], got %v", config.ConfigFiles)
	}
}

// ============================================================================
// ToolResult field tests
// ============================================================================

func TestToolResult_Fields(t *testing.T) {
	result := &ToolResult{
		Success:      true,
		ToolName:     "gofmt",
		Output:       "formatted output",
		Error:        "",
		ExitCode:     0,
		FileModified: true,
		IssuesFound:  3,
		IssuesFixed:  2,
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.ToolName != "gofmt" {
		t.Errorf("Expected ToolName 'gofmt', got '%s'", result.ToolName)
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected ExitCode 0, got %d", result.ExitCode)
	}
	if !result.FileModified {
		t.Error("Expected FileModified to be true")
	}
	if result.IssuesFound != 3 {
		t.Errorf("Expected IssuesFound 3, got %d", result.IssuesFound)
	}
	if result.IssuesFixed != 2 {
		t.Errorf("Expected IssuesFixed 2, got %d", result.IssuesFixed)
	}
}

// ============================================================================
// Registered tool configuration tests
// ============================================================================

func TestRegisteredPythonTools(t *testing.T) {
	r := NewToolRegistry()

	tools := r.tools["python"]
	if len(tools) < 3 {
		t.Fatalf("Expected at least 3 Python tools, got %d", len(tools))
	}

	// Verify tool names and types
	toolNames := make(map[string]ToolType)
	for _, tool := range tools {
		toolNames[tool.Name] = tool.ToolType
	}

	expectedTools := map[string]ToolType{
		"ruff-format": ToolTypeFormatter,
		"ruff-check":  ToolTypeLinter,
		"black":       ToolTypeFormatter,
		"mypy":        ToolTypeTypeChecker,
	}

	for name, expectedType := range expectedTools {
		tt, ok := toolNames[name]
		if !ok {
			t.Errorf("Expected Python tool '%s' to be registered", name)
			continue
		}
		if tt != expectedType {
			t.Errorf("Expected tool '%s' to have type '%s', got '%s'", name, expectedType, tt)
		}
	}
}

func TestRegisteredGoTools(t *testing.T) {
	r := NewToolRegistry()

	tools := r.tools["go"]
	if len(tools) < 2 {
		t.Fatalf("Expected at least 2 Go tools, got %d", len(tools))
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	expected := []string{"gofmt", "goimports", "golangci-lint"}
	for _, name := range expected {
		if !toolNames[name] {
			t.Errorf("Expected Go tool '%s' to be registered", name)
		}
	}
}

func TestRegisteredJSTSTools(t *testing.T) {
	r := NewToolRegistry()

	// JavaScript and TypeScript should share tools
	jsTools := r.tools["javascript"]
	tsTools := r.tools["typescript"]

	if len(jsTools) == 0 {
		t.Error("Expected JavaScript tools to be registered")
	}
	if len(tsTools) == 0 {
		t.Error("Expected TypeScript tools to be registered")
	}

	// They should reference the same slice
	if len(jsTools) != len(tsTools) {
		t.Errorf("Expected JS and TS to have same number of tools, got %d and %d", len(jsTools), len(tsTools))
	}

	toolNames := make(map[string]bool)
	for _, tool := range jsTools {
		toolNames[tool.Name] = true
	}

	expected := []string{"biome-format", "biome-lint", "prettier", "eslint"}
	for _, name := range expected {
		if !toolNames[name] {
			t.Errorf("Expected JS/TS tool '%s' to be registered", name)
		}
	}
}

func TestToolPriority(t *testing.T) {
	r := NewToolRegistry()

	// Ruff should have higher priority (lower number) than Black
	tools := r.tools["python"]

	var ruffPriority, blackPriority int
	for _, tool := range tools {
		if tool.Name == "ruff-format" {
			ruffPriority = tool.Priority
		}
		if tool.Name == "black" {
			blackPriority = tool.Priority
		}
	}

	if ruffPriority >= blackPriority {
		t.Errorf("Expected ruff-format priority (%d) < black priority (%d)", ruffPriority, blackPriority)
	}
}

func TestToolFileArgsPosition(t *testing.T) {
	r := NewToolRegistry()

	// Verify golangci-lint uses "start" position
	goTools := r.tools["go"]
	for _, tool := range goTools {
		if tool.Name == "golangci-lint" {
			if tool.FileArgsPosition != "start" {
				t.Errorf("Expected golangci-lint to use 'start' position, got '%s'", tool.FileArgsPosition)
			}
		}
	}

	// Verify R styler uses replace position
	rTools := r.tools["r"]
	for _, tool := range rTools {
		if tool.Name == "styler" {
			if tool.FileArgsPosition != "replace:style_file" {
				t.Errorf("Expected styler to use 'replace:style_file' position, got '%s'", tool.FileArgsPosition)
			}
		}
	}
}

// ============================================================================
// Extension map completeness tests
// ============================================================================

func TestExtensionMapCompleteness(t *testing.T) {
	r := NewToolRegistry()

	// Every registered extension should map to a language that has tools
	r.mu.RLock()
	defer r.mu.RUnlock()

	for ext, lang := range r.extensionMap {
		tools, ok := r.tools[lang]
		if !ok {
			t.Errorf("Extension '%s' maps to language '%s' which has no tools", ext, lang)
			continue
		}
		if len(tools) == 0 {
			t.Errorf("Extension '%s' maps to language '%s' which has empty tools", ext, lang)
		}
	}
}
