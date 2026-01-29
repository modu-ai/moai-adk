package tools

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ToolType represents the category of tool
type ToolType string

const (
	ToolTypeFormatter    ToolType = "formatter"
	ToolTypeLinter       ToolType = "linter"
	ToolTypeTypeChecker  ToolType = "type_checker"
	ToolTypeSecurityScan ToolType = "security_scanner"
	ToolTypeASTAnalyzer  ToolType = "ast_analyzer"
)

// ToolConfig represents a single tool configuration
type ToolConfig struct {
	Name             string
	Command          string
	Args             []string
	FileArgsPosition string // "end", "start", or "replace:placeholder"
	CheckArgs        []string
	FixArgs          []string
	Extensions       []string
	ToolType         ToolType
	Priority         int
	TimeoutSeconds   int
	RequiresConfig   bool
	ConfigFiles      []string
}

// ToolResult represents the result of tool execution
type ToolResult struct {
	Success      bool
	ToolName     string
	Output       string
	Error        string
	ExitCode     int
	FileModified bool
	IssuesFound  int
	IssuesFixed  int
}

// ToolRegistry manages language tools (formatters, linters, etc.)
type ToolRegistry struct {
	mu           sync.RWMutex
	tools        map[string][]*ToolConfig
	extensionMap map[string]string
	toolCache    map[string]bool
}

// NewToolRegistry creates a new tool registry with default tools
func NewToolRegistry() *ToolRegistry {
	r := &ToolRegistry{
		tools:        make(map[string][]*ToolConfig),
		extensionMap: make(map[string]string),
		toolCache:    make(map[string]bool),
	}
	r.registerDefaultTools()
	return r
}

// GetGlobalRegistry returns the singleton registry instance
var globalRegistry *ToolRegistry
var registryOnce sync.Once

func GetGlobalRegistry() *ToolRegistry {
	registryOnce.Do(func() {
		globalRegistry = NewToolRegistry()
	})
	return globalRegistry
}

// registerDefaultTools registers all default tools for supported languages
func (r *ToolRegistry) registerDefaultTools() {
	r.registerPythonTools()
	r.registerJSTSTools()
	r.registerGoTools()
	r.registerRustTools()
	r.registerJavaTools()
	r.registerKotlinTools()
	r.registerSwiftTools()
	r.registerCppTools()
	r.registerRubyTools()
	r.registerPHPTools()
	r.registerElixirTools()
	r.registerScalaTools()
	r.registerRTools()
	r.registerDartTools()
	r.registerCSharpTools()
	r.registerMarkdownTools()
	r.registerYAMLTools()
	r.registerJSONTools()
	r.registerShellTools()
	r.registerLuaTools()
}

// registerPythonTools registers Python formatting and linting tools
func (r *ToolRegistry) registerPythonTools() {
	extensions := []string{".py", ".pyi"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "python"
	}

	r.tools["python"] = append(r.tools["python"],
		// Ruff - fastest Python linter and formatter
		&ToolConfig{
			Name:             "ruff-format",
			Command:          "ruff",
			Args:             []string{"format"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "ruff-check",
			Command:          "ruff",
			Args:             []string{"check", "--fix"},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
		// Black - fallback formatter
		&ToolConfig{
			Name:             "black",
			Command:          "black",
			Args:             []string{},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         2,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
		// mypy - type checker
		&ToolConfig{
			Name:             "mypy",
			Command:          "mypy",
			Args:             []string{"--ignore-missing-imports"},
			Extensions:       extensions,
			ToolType:         ToolTypeTypeChecker,
			Priority:         1,
			TimeoutSeconds:   60,
			FileArgsPosition: "end",
		},
	)
}

// registerJSTSTools registers JavaScript/TypeScript tools
func (r *ToolRegistry) registerJSTSTools() {
	jsExtensions := []string{".js", ".jsx", ".mjs", ".cjs"}
	tsExtensions := []string{".ts", ".tsx", ".mts", ".cts"}

	for _, ext := range jsExtensions {
		r.extensionMap[ext] = "javascript"
	}
	for _, ext := range tsExtensions {
		r.extensionMap[ext] = "typescript"
	}

	sharedTools := []*ToolConfig{
		// Biome - fast all-in-one tool
		{
			Name:             "biome-format",
			Command:          "biome",
			Args:             []string{"format", "--write"},
			Extensions:       append(jsExtensions, tsExtensions...),
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
		{
			Name:             "biome-lint",
			Command:          "biome",
			Args:             []string{"lint", "--apply"},
			Extensions:       append(jsExtensions, tsExtensions...),
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
		// Prettier - universal formatter
		{
			Name:             "prettier",
			Command:          "prettier",
			Args:             []string{"--write"},
			Extensions:       append(jsExtensions, tsExtensions...),
			ToolType:         ToolTypeFormatter,
			Priority:         2,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
		// ESLint
		{
			Name:             "eslint",
			Command:          "eslint",
			Args:             []string{"--fix"},
			Extensions:       append(jsExtensions, tsExtensions...),
			ToolType:         ToolTypeLinter,
			Priority:         2,
			TimeoutSeconds:   60,
			FileArgsPosition: "end",
		},
	}

	r.tools["javascript"] = sharedTools
	r.tools["typescript"] = sharedTools
}

// registerGoTools registers Go tools
func (r *ToolRegistry) registerGoTools() {
	extensions := []string{".go"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "go"
	}

	r.tools["go"] = append(r.tools["go"],
		&ToolConfig{
			Name:             "gofmt",
			Command:          "gofmt",
			Args:             []string{"-w"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   15,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "goimports",
			Command:          "goimports",
			Args:             []string{"-w"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         2,
			TimeoutSeconds:   15,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "golangci-lint",
			Command:          "golangci-lint",
			Args:             []string{"run", "--fix"},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   120,
			FileArgsPosition: "start",
		},
	)
}

// registerRustTools registers Rust tools
func (r *ToolRegistry) registerRustTools() {
	extensions := []string{".rs"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "rust"
	}

	r.tools["rust"] = append(r.tools["rust"],
		&ToolConfig{
			Name:             "rustfmt",
			Command:          "rustfmt",
			Args:             []string{},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "clippy",
			Command:          "cargo",
			Args:             []string{"clippy", "--fix", "--allow-dirty", "--allow-staged"},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   180,
			FileArgsPosition: "end",
		},
	)
}

// registerJavaTools registers Java tools
func (r *ToolRegistry) registerJavaTools() {
	extensions := []string{".java"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "java"
	}

	r.tools["java"] = append(r.tools["java"],
		&ToolConfig{
			Name:             "google-java-format",
			Command:          "google-java-format",
			Args:             []string{"-i"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "checkstyle",
			Command:          "checkstyle",
			Args:             []string{"-c", "/google_checks.xml"},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   60,
			FileArgsPosition: "end",
		},
	)
}

// registerKotlinTools registers Kotlin tools
func (r *ToolRegistry) registerKotlinTools() {
	extensions := []string{".kt", ".kts"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "kotlin"
	}

	r.tools["kotlin"] = append(r.tools["kotlin"],
		&ToolConfig{
			Name:             "ktlint",
			Command:          "ktlint",
			Args:             []string{"-F"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "detekt",
			Command:          "detekt",
			Args:             []string{"-i"},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   60,
			FileArgsPosition: "end",
		},
	)
}

// registerSwiftTools registers Swift tools
func (r *ToolRegistry) registerSwiftTools() {
	extensions := []string{".swift"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "swift"
	}

	r.tools["swift"] = append(r.tools["swift"],
		&ToolConfig{
			Name:             "swift-format",
			Command:          "swift-format",
			Args:             []string{"format", "-i"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "swiftlint",
			Command:          "swiftlint",
			Args:             []string{"--fix"},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   60,
			FileArgsPosition: "end",
		},
	)
}

// registerCppTools registers C/C++ tools
func (r *ToolRegistry) registerCppTools() {
	extensions := []string{".c", ".cpp", ".cc", ".cxx", ".h", ".hpp", ".hxx"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "cpp"
	}

	r.tools["cpp"] = append(r.tools["cpp"],
		&ToolConfig{
			Name:             "clang-format",
			Command:          "clang-format",
			Args:             []string{"-i"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   15,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "clang-tidy",
			Command:          "clang-tidy",
			Args:             []string{"--fix"},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   120,
			FileArgsPosition: "end",
		},
	)
}

// registerRubyTools registers Ruby tools
func (r *ToolRegistry) registerRubyTools() {
	extensions := []string{".rb", ".rake", ".gemspec"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "ruby"
	}

	r.tools["ruby"] = append(r.tools["ruby"],
		&ToolConfig{
			Name:             "rubocop",
			Command:          "rubocop",
			Args:             []string{"-a"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   60,
			FileArgsPosition: "end",
		},
	)
}

// registerPHPTools registers PHP tools
func (r *ToolRegistry) registerPHPTools() {
	extensions := []string{".php"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "php"
	}

	r.tools["php"] = append(r.tools["php"],
		&ToolConfig{
			Name:             "php-cs-fixer",
			Command:          "php-cs-fixer",
			Args:             []string{"fix"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   60,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "phpstan",
			Command:          "phpstan",
			Args:             []string{"analyze"},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   120,
			FileArgsPosition: "end",
		},
	)
}

// registerElixirTools registers Elixir tools
func (r *ToolRegistry) registerElixirTools() {
	extensions := []string{".ex", ".exs"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "elixir"
	}

	r.tools["elixir"] = append(r.tools["elixir"],
		&ToolConfig{
			Name:             "mix-format",
			Command:          "mix",
			Args:             []string{"format"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "credo",
			Command:          "mix",
			Args:             []string{"credo", "--strict"},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   60,
			FileArgsPosition: "end",
		},
	)
}

// registerScalaTools registers Scala tools
func (r *ToolRegistry) registerScalaTools() {
	extensions := []string{".scala", ".sc"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "scala"
	}

	r.tools["scala"] = append(r.tools["scala"],
		&ToolConfig{
			Name:             "scalafmt",
			Command:          "scalafmt",
			Args:             []string{},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   60,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "scalafix",
			Command:          "scalafix",
			Args:             []string{},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   120,
			FileArgsPosition: "end",
		},
	)
}

// registerRTools registers R tools
func (r *ToolRegistry) registerRTools() {
	extensions := []string{".r", ".R", ".Rmd"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "r"
	}

	r.tools["r"] = append(r.tools["r"],
		&ToolConfig{
			Name:             "styler",
			Command:          "Rscript",
			Args:             []string{"-e", "styler::style_file"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   60,
			FileArgsPosition: "replace:style_file",
		},
		&ToolConfig{
			Name:             "lintr",
			Command:          "Rscript",
			Args:             []string{"-e", "lintr::lint"},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   60,
			FileArgsPosition: "end",
		},
	)
}

// registerDartTools registers Dart/Flutter tools
func (r *ToolRegistry) registerDartTools() {
	extensions := []string{".dart"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "dart"
	}

	r.tools["dart"] = append(r.tools["dart"],
		&ToolConfig{
			Name:             "dart-format",
			Command:          "dart",
			Args:             []string{"format"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "dart-analyze",
			Command:          "dart",
			Args:             []string{"analyze"},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   60,
			FileArgsPosition: "end",
		},
	)
}

// registerCSharpTools registers C# tools
func (r *ToolRegistry) registerCSharpTools() {
	extensions := []string{".cs"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "csharp"
	}

	r.tools["csharp"] = append(r.tools["csharp"],
		&ToolConfig{
			Name:             "dotnet-format",
			Command:          "dotnet",
			Args:             []string{"format"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   60,
			FileArgsPosition: "end",
		},
	)
}

// registerMarkdownTools registers Markdown tools
func (r *ToolRegistry) registerMarkdownTools() {
	extensions := []string{".md", ".mdx", ".markdown"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "markdown"
	}

	r.tools["markdown"] = append(r.tools["markdown"],
		&ToolConfig{
			Name:             "prettier-md",
			Command:          "prettier",
			Args:             []string{"--write", "--parser", "markdown"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   15,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "markdownlint",
			Command:          "markdownlint",
			Args:             []string{"--fix"},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
	)
}

// registerYAMLTools registers YAML tools
func (r *ToolRegistry) registerYAMLTools() {
	extensions := []string{".yaml", ".yml"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "yaml"
	}

	r.tools["yaml"] = append(r.tools["yaml"],
		&ToolConfig{
			Name:             "prettier-yaml",
			Command:          "prettier",
			Args:             []string{"--write"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   15,
			FileArgsPosition: "end",
		},
	)
}

// registerJSONTools registers JSON tools
func (r *ToolRegistry) registerJSONTools() {
	extensions := []string{".json"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "json"
	}

	r.tools["json"] = append(r.tools["json"],
		&ToolConfig{
			Name:             "prettier-json",
			Command:          "prettier",
			Args:             []string{"--write"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   15,
			FileArgsPosition: "end",
		},
	)
}

// registerShellTools registers Shell tools
func (r *ToolRegistry) registerShellTools() {
	extensions := []string{".sh", ".bash", ".zsh"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "shell"
	}

	r.tools["shell"] = append(r.tools["shell"],
		&ToolConfig{
			Name:             "shfmt",
			Command:          "shfmt",
			Args:             []string{"-w"},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   15,
			FileArgsPosition: "end",
		},
		&ToolConfig{
			Name:             "shellcheck",
			Command:          "shellcheck",
			Args:             []string{},
			Extensions:       extensions,
			ToolType:         ToolTypeLinter,
			Priority:         1,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
	)
}

// registerLuaTools registers Lua tools
func (r *ToolRegistry) registerLuaTools() {
	extensions := []string{".lua"}
	for _, ext := range extensions {
		r.extensionMap[ext] = "lua"
	}

	r.tools["lua"] = append(r.tools["lua"],
		&ToolConfig{
			Name:             "stylua",
			Command:          "stylua",
			Args:             []string{},
			Extensions:       extensions,
			ToolType:         ToolTypeFormatter,
			Priority:         1,
			TimeoutSeconds:   30,
			FileArgsPosition: "end",
		},
	)
}

// IsToolAvailable checks if a tool is available on the system
func (r *ToolRegistry) IsToolAvailable(toolName string) bool {
	// Check cache first with read lock
	r.mu.RLock()
	available, ok := r.toolCache[toolName]
	r.mu.RUnlock()

	if ok {
		return available
	}

	// Find the tool and check availability
	for _, langTools := range r.tools {
		for _, tool := range langTools {
			if tool.Name == toolName {
				_, err := exec.LookPath(tool.Command)
				// Cache the result with write lock
				r.mu.Lock()
				defer r.mu.Unlock()
				r.toolCache[toolName] = err == nil
				return err == nil
			}
		}
	}

	return false
}

// GetLanguageForFile detects the programming language for a file
func (r *ToolRegistry) GetLanguageForFile(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	r.mu.RLock()
	defer r.mu.RUnlock()

	if lang, ok := r.extensionMap[ext]; ok {
		return lang
	}
	return ""
}

// GetToolsForLanguage gets available tools for a language
func (r *ToolRegistry) GetToolsForLanguage(language string, toolType ToolType) []*ToolConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools, ok := r.tools[language]
	if !ok {
		return nil
	}

	// Filter by type and availability
	var available []*ToolConfig
	for _, tool := range tools {
		if toolType != "" && tool.ToolType != toolType {
			continue
		}
		if r.IsToolAvailable(tool.Name) {
			available = append(available, tool)
		}
	}

	// Sort by priority (already sorted in registration, but double-check)
	// For now, return as-is since we registered in priority order
	return available
}

// GetToolsForFile gets available tools for a specific file
func (r *ToolRegistry) GetToolsForFile(filePath string, toolType ToolType) []*ToolConfig {
	language := r.GetLanguageForFile(filePath)
	if language == "" {
		return nil
	}
	return r.GetToolsForLanguage(language, toolType)
}

// RunTool executes a tool on a file
func (r *ToolRegistry) RunTool(ctx context.Context, tool *ToolConfig, filePath string) (*ToolResult, error) {
	// Build command args
	args := append([]string{}, tool.Args...)

	// Add file path based on position
	switch tool.FileArgsPosition {
	case "end":
		args = append(args, filePath)
	case "start":
		// Insert after command
		if len(args) > 0 {
			args = append([]string{filePath}, args...)
		} else {
			args = []string{filePath}
		}
	default:
		// Handle "replace:placeholder" case
		if strings.HasPrefix(tool.FileArgsPosition, "replace:") {
			placeholder := strings.TrimPrefix(tool.FileArgsPosition, "replace:")
			for i, arg := range args {
				if arg == placeholder {
					// Escape path for code string inclusion (R, etc.)
					escapedPath := strings.ReplaceAll(filePath, "'", "\\'")
					args[i] = fmt.Sprintf("%s('%s')", placeholder, escapedPath)
					break
				}
			}
		} else {
			// Default to end
			args = append(args, filePath)
		}
	}

	// Create command
	cmd := exec.CommandContext(ctx, tool.Command, args...)

	// Run and capture output
	output, err := cmd.CombinedOutput()

	result := &ToolResult{
		ToolName: tool.Name,
		Output:   string(output),
	}

	if err != nil {
		result.Error = err.Error()
		result.ExitCode = 1
		return result, err
	}

	result.Success = true
	result.ExitCode = 0
	return result, nil
}
