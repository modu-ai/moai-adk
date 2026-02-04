package quality

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ToolRegistry manages code quality tool registration and execution.
// It provides type-safe tool registration, language-based tool lookup,
// file extension mapping, and priority-based tool selection.
type ToolRegistry struct {
	mu           sync.RWMutex
	tools        map[string][]ToolConfig // language -> tools
	extensionMap map[string]string       // extension -> language
	availCache   map[string]bool         // tool name -> available
	initialized  bool
}

// NewToolRegistry creates a new ToolRegistry with default tools registered.
func NewToolRegistry() *ToolRegistry {
	r := &ToolRegistry{
		tools:        make(map[string][]ToolConfig),
		extensionMap: make(map[string]string),
		availCache:   make(map[string]bool),
	}
	r.registerDefaultTools()
	return r
}

// registerDefaultTools registers all supported language tools.
func (r *ToolRegistry) registerDefaultTools() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.initialized {
		return
	}

	r.registerPythonTools()
	r.registerJavaScriptTools()
	r.registerTypeScriptTools()
	r.registerGoTools()
	r.registerRustTools()
	r.registerJavaTools()
	r.registerKotlinTools()
	r.registerSwiftTools()
	r.registerCppTools()
	r.registerRubyTools()
	r.registerPhpTools()
	r.registerElixirTools()
	r.registerScalaTools()
	r.registerRTools()
	r.registerDartTools()
	r.registerCSharpTools()
	r.registerMarkdownTools()

	r.initialized = true
}

// RegisterTool adds a tool configuration to the registry.
func (r *ToolRegistry) RegisterTool(tool ToolConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Build language from extensions or derive from tool name
	language := r.deriveLanguageFromTool(tool)

	r.tools[language] = append(r.tools[language], tool)

	// Update extension map
	for _, ext := range tool.Extensions {
		if _, exists := r.extensionMap[ext]; !exists {
			r.extensionMap[ext] = language
		}
	}
}

// deriveLanguageFromTool determines the language from tool configuration.
func (r *ToolRegistry) deriveLanguageFromTool(tool ToolConfig) string {
	// Check if extensions map to an existing language
	for _, ext := range tool.Extensions {
		if lang, ok := r.extensionMap[ext]; ok {
			return lang
		}
	}

	// Derive from tool name
	if strings.Contains(tool.Name, "python") || strings.Contains(tool.Command, "python") {
		return "python"
	}

	// Use first extension as hint
	if len(tool.Extensions) > 0 {
		ext := tool.Extensions[0]
		switch ext {
		case ".py", ".pyi":
			return "python"
		case ".go":
			return "go"
		case ".rs":
			return "rust"
		case ".js", ".jsx", ".mjs", ".cjs":
			return "javascript"
		case ".ts", ".tsx", ".mts", ".cts":
			return "typescript"
		case ".java":
			return "java"
		case ".kt", ".kts":
			return "kotlin"
		case ".swift":
			return "swift"
		case ".c", ".cpp", ".cc", ".cxx", ".h", ".hpp", ".hxx":
			return "cpp"
		case ".rb", ".rake", ".gemspec":
			return "ruby"
		case ".php":
			return "php"
		case ".ex", ".exs":
			return "elixir"
		case ".scala", ".sc":
			return "scala"
		case ".r", ".R", ".Rmd":
			return "r"
		case ".dart":
			return "dart"
		case ".cs":
			return "csharp"
		case ".md", ".mdx":
			return "markdown"
		}
	}

	return "unknown"
}

// GetToolsForLanguage returns tools for a language, filtered by type and sorted by priority.
func (r *ToolRegistry) GetToolsForLanguage(language string, toolType ToolType) []ToolConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools, ok := r.tools[language]
	if !ok {
		return nil
	}

	var filtered []ToolConfig
	for _, tool := range tools {
		if tool.ToolType == toolType {
			filtered = append(filtered, tool)
		}
	}

	// Sort by priority (lower = higher priority)
	sortByPriority(filtered)

	return filtered
}

// GetToolsForFile returns tools for a specific file, filtered by type and sorted by priority.
func (r *ToolRegistry) GetToolsForFile(filePath string, toolType ToolType) []ToolConfig {
	ext := strings.ToLower(filepath.Ext(filePath))
	r.mu.RLock()
	language, ok := r.extensionMap[ext]
	r.mu.RUnlock()

	if !ok {
		return nil
	}

	return r.GetToolsForLanguage(language, toolType)
}

// GetLanguageForFile returns the language identifier for a file path.
func (r *ToolRegistry) GetLanguageForFile(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.extensionMap[ext]
}

// IsToolAvailable checks if a tool command exists on the system.
// Results are cached for performance.
func (r *ToolRegistry) IsToolAvailable(toolName string) bool {
	r.mu.RLock()
	if cached, ok := r.availCache[toolName]; ok {
		r.mu.RUnlock()
		return cached
	}
	r.mu.RUnlock()

	// Find the tool config to get the command
	r.mu.RLock()
	var command string
	for _, tools := range r.tools {
		for _, tool := range tools {
			if tool.Name == toolName {
				command = tool.Command
				break
			}
		}
		if command != "" {
			break
		}
	}
	r.mu.RUnlock()

	if command == "" {
		return false
	}

	// Check if command exists
	_, err := exec.LookPath(command)
	available := err == nil

	r.mu.Lock()
	r.availCache[toolName] = available
	r.mu.Unlock()

	return available
}

// RunTool executes a tool on a file with proper timeout and error handling.
func (r *ToolRegistry) RunTool(ctx context.Context, tool ToolConfig, filePath string, cwd string) ToolResult {
	startTime := time.Now()

	// Validate file path
	if err := validateFilePath(filePath); err != nil {
		return ToolResult{
			Success:       false,
			ToolName:      tool.Name,
			Error:         fmt.Sprintf("path validation failed: %v", err),
			ExitCode:      -1,
			ExecutionTime: time.Since(startTime),
		}
	}

	// Get file hash before execution
	detector := NewChangeDetector()
	hashBefore, _ := detector.ComputeHash(filePath)

	// Build command
	cmd, err := r.buildCommand(tool, filePath, cwd)
	if err != nil {
		return ToolResult{
			Success:       false,
			ToolName:      tool.Name,
			Error:         fmt.Sprintf("command build failed: %v", err),
			ExitCode:      -1,
			ExecutionTime: time.Since(startTime),
		}
	}

	// Set timeout
	timeout := time.Duration(tool.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = time.Duration(DefaultTimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute command
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	// Get file hash after execution
	hashAfter, _ := detector.ComputeHash(filePath)
	fileModified := !bytes.Equal(hashBefore, hashAfter)

	// Build result
	result := ToolResult{
		ToolName:      tool.Name,
		Output:        stdout.String(),
		Error:         stderr.String(),
		FileModified:  fileModified,
		ExecutionTime: time.Since(startTime),
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Success = false
			result.ExitCode = -1
			result.Error = fmt.Sprintf("tool timed out after %v", timeout)
		} else {
			result.Success = false
			result.ExitCode = getExitCode(err)
			if result.Error == "" {
				result.Error = err.Error()
			}
		}
	} else {
		result.Success = true
		result.ExitCode = 0
	}

	return result
}

// buildCommand constructs an exec.Cmd for the tool.
func (r *ToolRegistry) buildCommand(tool ToolConfig, filePath string, cwd string) (*exec.Cmd, error) {
	// Start with base command and args
	args := make([]string, len(tool.Args))
	copy(args, tool.Args)

	// Add file path based on position
	switch tool.FileArgsPosition {
	case FileArgsPositionEnd, "":
		args = append(args, filePath)
	case FileArgsPositionStart:
		// Insert after command name
		args = append([]string{filePath}, args...)
	case FileArgsPositionReplace:
		// For code interpolation tools (like R's styler)
		// Format: "replace:placeholder"
		placeholder := strings.TrimPrefix(string(tool.FileArgsPosition), "replace:")
		for i, arg := range args {
			if strings.Contains(arg, placeholder) {
				// Escape the file path for safe code inclusion
				escaped := escapePathForCode(filePath)
				args[i] = strings.ReplaceAll(arg, placeholder, escaped)
			}
		}
	}

	cmd := exec.Command(tool.Command, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}

	return cmd, nil
}

// validateFilePath validates and sanitizes a file path to prevent injection attacks.
func validateFilePath(filePath string) error {
	// Check for null bytes
	if strings.Contains(filePath, "\x00") {
		return fmt.Errorf("path contains null byte")
	}

	// Check for dangerous shell metacharacters
	dangerous := []string{"'", "\"", "`", "$", ";", "&", "|", "\n", "\r"}
	for _, char := range dangerous {
		if strings.Contains(filePath, char) {
			return fmt.Errorf("path contains dangerous character: %q", char)
		}
	}

	// Resolve path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if file exists
	info, err := filepath.Abs(absPath)
	if err != nil {
		return fmt.Errorf("cannot access path: %w", err)
	}

	if _, err := filepath.Abs(info); err != nil {
		return fmt.Errorf("path error: %w", err)
	}

	return nil
}

// escapePathForCode escapes a file path for safe inclusion in code strings.
func escapePathForCode(path string) string {
	escaped := strings.ReplaceAll(path, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "'", "\\'")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	return escaped
}

// getExitCode extracts the exit code from an error.
func getExitCode(err error) int {
	if err == nil {
		return 0
	}

	// Try to get exit code from *exec.ExitError
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() >= 0 {
			return exitErr.ExitCode()
		}
	}

	return -1
}

// sortByPriority sorts tools by priority (lower = higher priority).
func sortByPriority(tools []ToolConfig) {
	// Simple insertion sort for small slices
	for i := 1; i < len(tools); i++ {
		for j := i; j > 0 && tools[j].Priority < tools[j-1].Priority; j-- {
			tools[j], tools[j-1] = tools[j-1], tools[j]
		}
	}
}

// Language tool registration methods

func (r *ToolRegistry) registerPythonTools() {
	exts := []string{".py", ".pyi"}
	for _, ext := range exts {
		r.extensionMap[ext] = "python"
	}

	r.tools["python"] = []ToolConfig{
		{
			Name:           "ruff-format",
			Command:        "ruff",
			Args:           []string{"format"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 30,
		},
		{
			Name:           "ruff-check",
			Command:        "ruff",
			Args:           []string{"check", "--fix"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 30,
		},
		{
			Name:           "black",
			Command:        "black",
			Args:           []string{},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       2,
			TimeoutSeconds: 30,
		},
		{
			Name:           "isort",
			Command:        "isort",
			Args:           []string{},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       3,
			TimeoutSeconds: 15,
		},
	}
}

func (r *ToolRegistry) registerJavaScriptTools() {
	exts := []string{".js", ".jsx", ".mjs", ".cjs"}
	for _, ext := range exts {
		r.extensionMap[ext] = "javascript"
	}

	r.tools["javascript"] = []ToolConfig{
		{
			Name:           "biome-format",
			Command:        "biome",
			Args:           []string{"format", "--write"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 30,
		},
		{
			Name:           "biome-lint",
			Command:        "biome",
			Args:           []string{"lint", "--apply"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 30,
		},
		{
			Name:           "prettier",
			Command:        "prettier",
			Args:           []string{"--write"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       2,
			TimeoutSeconds: 30,
		},
		{
			Name:           "eslint",
			Command:        "eslint",
			Args:           []string{"--fix"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       2,
			TimeoutSeconds: 60,
		},
	}
}

func (r *ToolRegistry) registerTypeScriptTools() {
	exts := []string{".ts", ".tsx", ".mts", ".cts"}
	for _, ext := range exts {
		r.extensionMap[ext] = "typescript"
	}

	r.tools["typescript"] = []ToolConfig{
		{
			Name:           "biome-format",
			Command:        "biome",
			Args:           []string{"format", "--write"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 30,
		},
		{
			Name:           "biome-lint",
			Command:        "biome",
			Args:           []string{"lint", "--apply"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 30,
		},
		{
			Name:           "prettier",
			Command:        "prettier",
			Args:           []string{"--write"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       2,
			TimeoutSeconds: 30,
		},
		{
			Name:           "eslint",
			Command:        "eslint",
			Args:           []string{"--fix"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       2,
			TimeoutSeconds: 60,
		},
		{
			Name:           "tsc",
			Command:        "tsc",
			Args:           []string{"--noEmit"},
			Extensions:     exts,
			ToolType:       ToolTypeTypeChecker,
			Priority:       1,
			TimeoutSeconds: 60,
		},
	}
}

func (r *ToolRegistry) registerGoTools() {
	exts := []string{".go"}
	for _, ext := range exts {
		r.extensionMap[ext] = "go"
	}

	r.tools["go"] = []ToolConfig{
		{
			Name:           "gofmt",
			Command:        "gofmt",
			Args:           []string{"-w"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 15,
		},
		{
			Name:           "goimports",
			Command:        "goimports",
			Args:           []string{"-w"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       2,
			TimeoutSeconds: 15,
		},
		{
			Name:           "golangci-lint",
			Command:        "golangci-lint",
			Args:           []string{"run", "--fix"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 120,
		},
	}
}

func (r *ToolRegistry) registerRustTools() {
	exts := []string{".rs"}
	for _, ext := range exts {
		r.extensionMap[ext] = "rust"
	}

	r.tools["rust"] = []ToolConfig{
		{
			Name:           "rustfmt",
			Command:        "rustfmt",
			Args:           []string{},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 30,
		},
		{
			Name:           "clippy",
			Command:        "cargo",
			Args:           []string{"clippy", "--fix", "--allow-dirty", "--allow-staged"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 180,
		},
	}
}

func (r *ToolRegistry) registerJavaTools() {
	exts := []string{".java"}
	for _, ext := range exts {
		r.extensionMap[ext] = "java"
	}

	r.tools["java"] = []ToolConfig{
		{
			Name:           "google-java-format",
			Command:        "google-java-format",
			Args:           []string{"-i"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 30,
		},
		{
			Name:           "checkstyle",
			Command:        "checkstyle",
			Args:           []string{"-c", "/google_checks.xml"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 60,
		},
	}
}

func (r *ToolRegistry) registerKotlinTools() {
	exts := []string{".kt", ".kts"}
	for _, ext := range exts {
		r.extensionMap[ext] = "kotlin"
	}

	r.tools["kotlin"] = []ToolConfig{
		{
			Name:           "ktlint",
			Command:        "ktlint",
			Args:           []string{"-F"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 30,
		},
		{
			Name:           "detekt",
			Command:        "detekt",
			Args:           []string{"-i"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 60,
		},
	}
}

func (r *ToolRegistry) registerSwiftTools() {
	exts := []string{".swift"}
	for _, ext := range exts {
		r.extensionMap[ext] = "swift"
	}

	r.tools["swift"] = []ToolConfig{
		{
			Name:           "swift-format",
			Command:        "swift-format",
			Args:           []string{"format", "-i"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 30,
		},
		{
			Name:           "swiftlint",
			Command:        "swiftlint",
			Args:           []string{"--fix"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 60,
		},
	}
}

func (r *ToolRegistry) registerCppTools() {
	exts := []string{".c", ".cpp", ".cc", ".cxx", ".h", ".hpp", ".hxx"}
	for _, ext := range exts {
		r.extensionMap[ext] = "cpp"
	}

	r.tools["cpp"] = []ToolConfig{
		{
			Name:           "clang-format",
			Command:        "clang-format",
			Args:           []string{"-i"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 15,
		},
		{
			Name:           "clang-tidy",
			Command:        "clang-tidy",
			Args:           []string{"--fix"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 120,
		},
	}
}

func (r *ToolRegistry) registerRubyTools() {
	exts := []string{".rb", ".rake", ".gemspec"}
	for _, ext := range exts {
		r.extensionMap[ext] = "ruby"
	}

	r.tools["ruby"] = []ToolConfig{
		{
			Name:           "rubocop",
			Command:        "rubocop",
			Args:           []string{"-a"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 60,
		},
	}
}

func (r *ToolRegistry) registerPhpTools() {
	exts := []string{".php"}
	for _, ext := range exts {
		r.extensionMap[ext] = "php"
	}

	r.tools["php"] = []ToolConfig{
		{
			Name:           "php-cs-fixer",
			Command:        "php-cs-fixer",
			Args:           []string{"fix"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 60,
		},
		{
			Name:           "phpstan",
			Command:        "phpstan",
			Args:           []string{"analyze"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 120,
		},
	}
}

func (r *ToolRegistry) registerElixirTools() {
	exts := []string{".ex", ".exs"}
	for _, ext := range exts {
		r.extensionMap[ext] = "elixir"
	}

	r.tools["elixir"] = []ToolConfig{
		{
			Name:           "mix-format",
			Command:        "mix",
			Args:           []string{"format"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 30,
		},
		{
			Name:           "credo",
			Command:        "mix",
			Args:           []string{"credo", "--strict"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 60,
		},
	}
}

func (r *ToolRegistry) registerScalaTools() {
	exts := []string{".scala", ".sc"}
	for _, ext := range exts {
		r.extensionMap[ext] = "scala"
	}

	r.tools["scala"] = []ToolConfig{
		{
			Name:           "scalafmt",
			Command:        "scalafmt",
			Args:           []string{},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 60,
		},
		{
			Name:           "scalafix",
			Command:        "scalafix",
			Args:           []string{},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 120,
		},
	}
}

func (r *ToolRegistry) registerRTools() {
	exts := []string{".r", ".R", ".Rmd"}
	for _, ext := range exts {
		r.extensionMap[ext] = "r"
	}

	r.tools["r"] = []ToolConfig{
		{
			Name:             "styler",
			Command:          "Rscript",
			Args:             []string{"-e", "styler::style_file"},
			FileArgsPosition: FileArgsPositionReplace + ":style_file",
			Extensions:       exts,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   60,
		},
		{
			Name:           "lintr",
			Command:        "Rscript",
			Args:           []string{"-e", "lintr::lint"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 60,
		},
	}
}

func (r *ToolRegistry) registerDartTools() {
	exts := []string{".dart"}
	for _, ext := range exts {
		r.extensionMap[ext] = "dart"
	}

	r.tools["dart"] = []ToolConfig{
		{
			Name:           "dart-format",
			Command:        "dart",
			Args:           []string{"format"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 30,
		},
		{
			Name:           "dart-analyze",
			Command:        "dart",
			Args:           []string{"analyze"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 60,
		},
	}
}

func (r *ToolRegistry) registerCSharpTools() {
	exts := []string{".cs"}
	for _, ext := range exts {
		r.extensionMap[ext] = "csharp"
	}

	r.tools["csharp"] = []ToolConfig{
		{
			Name:           "dotnet-format",
			Command:        "dotnet",
			Args:           []string{"format"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 60,
		},
	}
}

func (r *ToolRegistry) registerMarkdownTools() {
	exts := []string{".md", ".mdx", ".markdown"}
	for _, ext := range exts {
		r.extensionMap[ext] = "markdown"
	}

	r.tools["markdown"] = []ToolConfig{
		{
			Name:           "prettier-md",
			Command:        "prettier",
			Args:           []string{"--write", "--parser", "markdown"},
			Extensions:     exts,
			ToolType:       ToolTypeFormatter,
			Priority:       1,
			TimeoutSeconds: 15,
		},
		{
			Name:           "markdownlint",
			Command:        "markdownlint",
			Args:           []string{"--fix"},
			Extensions:     exts,
			ToolType:       ToolTypeLinter,
			Priority:       1,
			TimeoutSeconds: 30,
		},
	}
}
