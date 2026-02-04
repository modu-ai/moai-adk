package quality

import (
	"context"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

// ToolType represents the type of code quality tool.
type ToolType string

const (
	ToolTypeFormatter   ToolType = "formatter"
	ToolTypeLinter      ToolType = "linter"
	ToolTypeTypeChecker ToolType = "type_checker"
)

// ToolConfig represents configuration for a single tool.
type ToolConfig struct {
	Name           string
	Command        string
	Args           []string
	Extensions     []string
	ToolType       ToolType
	Priority       int
	TimeoutSeconds int
	CheckArgs      []string
	FixArgs        []string
}

// ToolResult represents the result of tool execution.
type ToolResult struct {
	Success       bool
	ToolName      string
	Output        string
	Error         string
	ExitCode      int
	FileModified  bool
	IssuesFound   int
	IssuesFixed   int
	ExecutionTime time.Duration
}

// toolRegistry manages code quality tool registration and execution per REQ-HOOK-050.
type toolRegistry struct {
	mu        sync.RWMutex
	tools     map[ToolType][]ToolConfig
	available map[string]bool
}

// NewToolRegistry creates a new ToolRegistry with default tools per REQ-HOOK-051.
func NewToolRegistry() *toolRegistry {
	r := &toolRegistry{
		tools:     make(map[ToolType][]ToolConfig),
		available: make(map[string]bool),
	}
	r.registerDefaultTools()
	return r
}

// RegisterTool adds a tool to the registry per REQ-HOOK-050.
func (r *toolRegistry) RegisterTool(tool ToolConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tool.TimeoutSeconds == 0 {
		tool.TimeoutSeconds = 30
	}

	r.tools[tool.ToolType] = append(r.tools[tool.ToolType], tool)
}

// GetToolsForLanguage returns tools for a language sorted by priority per REQ-HOOK-052.
func (r *toolRegistry) GetToolsForLanguage(language string, toolType ToolType) []ToolConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := []ToolConfig{}
	for _, tool := range r.tools[toolType] {
		for _, ext := range tool.Extensions {
			if languageFromExtension(ext) == language {
				result = append(result, tool)
				break
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Priority < result[j].Priority
	})

	return result
}

// GetToolsForFile returns tools for a specific file per REQ-HOOK-050.
func (r *toolRegistry) GetToolsForFile(filePath string, toolType ToolType) []ToolConfig {
	ext := extensionFromPath(filePath)
	language := languageFromExtension(ext)
	return r.GetToolsForLanguage(language, toolType)
}

// IsToolAvailable checks if a tool binary exists per REQ-HOOK-050.
func (r *toolRegistry) IsToolAvailable(toolName string) bool {
	r.mu.RLock()
	cached, ok := r.available[toolName]
	r.mu.RUnlock()

	if ok {
		return cached
	}

	_, err := exec.LookPath(toolName)

	r.mu.Lock()
	defer r.mu.Unlock()
	r.available[toolName] = err == nil

	return err == nil
}

// RunTool executes a tool per REQ-HOOK-050 using exec.Command (not shell) per REQ-HOOK-053.
func (r *toolRegistry) RunTool(tool ToolConfig, filePath string, cwd string) ToolResult {
	start := time.Now()
	result := ToolResult{
		ToolName: tool.Name,
	}

	args := make([]string, len(tool.Args))
	for i, arg := range tool.Args {
		if arg == "{file}" || arg == "$FILE" {
			args[i] = filePath
		} else {
			args[i] = arg
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tool.TimeoutSeconds)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, tool.Command, args...)
	cmd.Dir = cwd

	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	result.ExecutionTime = time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		result.Success = false
		result.Error = "command timed out"
		result.ExitCode = -1
		return result
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		return result
	}

	result.Success = true
	result.ExitCode = 0
	return result
}

// registerDefaultTools registers default tools for 16+ languages per REQ-HOOK-051.
func (r *toolRegistry) registerDefaultTools() {
	// Python tools
	r.RegisterTool(ToolConfig{Name: "ruff-format", Command: "ruff", Args: []string{"format", "{file}"}, Extensions: []string{".py", ".pyi"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "black", Command: "black", Args: []string{"--quiet", "{file}"}, Extensions: []string{".py", ".pyi"}, ToolType: ToolTypeFormatter, Priority: 2})
	r.RegisterTool(ToolConfig{Name: "ruff", Command: "ruff", Args: []string{"check", "{file}"}, Extensions: []string{".py", ".pyi"}, ToolType: ToolTypeLinter, Priority: 1, FixArgs: []string{"--fix"}})
	r.RegisterTool(ToolConfig{Name: "mypy", Command: "mypy", Args: []string{"{file}"}, Extensions: []string{".py", ".pyi"}, ToolType: ToolTypeTypeChecker, Priority: 1})

	// JavaScript tools
	r.RegisterTool(ToolConfig{Name: "biome-format", Command: "biome", Args: []string{"format", "--write", "{file}"}, Extensions: []string{".js", ".jsx", ".mjs", ".cjs"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "prettier-format", Command: "prettier", Args: []string{"--write", "{file}"}, Extensions: []string{".js", ".jsx", ".mjs", ".cjs"}, ToolType: ToolTypeFormatter, Priority: 2})
	r.RegisterTool(ToolConfig{Name: "eslint", Command: "eslint", Args: []string{"{file}"}, Extensions: []string{".js", ".jsx", ".mjs", ".cjs"}, ToolType: ToolTypeLinter, Priority: 1, FixArgs: []string{"--fix"}})

	// TypeScript tools
	r.RegisterTool(ToolConfig{Name: "biome-format-ts", Command: "biome", Args: []string{"format", "--write", "{file}"}, Extensions: []string{".ts", ".tsx", ".mts", ".cts"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "prettier-format-ts", Command: "prettier", Args: []string{"--write", "{file}"}, Extensions: []string{".ts", ".tsx", ".mts", ".cts"}, ToolType: ToolTypeFormatter, Priority: 2})
	r.RegisterTool(ToolConfig{Name: "eslint-ts", Command: "eslint", Args: []string{"{file}"}, Extensions: []string{".ts", ".tsx", ".mts", ".cts"}, ToolType: ToolTypeLinter, Priority: 1, FixArgs: []string{"--fix"}})
	r.RegisterTool(ToolConfig{Name: "tsc", Command: "tsc", Args: []string{"--noEmit", "{file}"}, Extensions: []string{".ts", ".tsx"}, ToolType: ToolTypeTypeChecker, Priority: 1})

	// Go tools
	r.RegisterTool(ToolConfig{Name: "gofmt", Command: "gofmt", Args: []string{"-w", "{file}"}, Extensions: []string{".go"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "goimports", Command: "goimports", Args: []string{"-w", "{file}"}, Extensions: []string{".go"}, ToolType: ToolTypeFormatter, Priority: 2})
	r.RegisterTool(ToolConfig{Name: "golangci-lint", Command: "golangci-lint", Args: []string{"run", "--no-config", "{file}"}, Extensions: []string{".go"}, ToolType: ToolTypeLinter, Priority: 1})

	// Rust tools
	r.RegisterTool(ToolConfig{Name: "rustfmt", Command: "rustfmt", Args: []string{"{file}"}, Extensions: []string{".rs"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "clippy", Command: "cargo", Args: []string{"clippy", "--message-format=short"}, Extensions: []string{".rs"}, ToolType: ToolTypeLinter, Priority: 1})

	// Java tools
	r.RegisterTool(ToolConfig{Name: "google-java-format", Command: "google-java-format", Args: []string{"--replace", "{file}"}, Extensions: []string{".java"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "checkstyle", Command: "checkstyle", Args: []string{"-c", "/google_checks.xml", "{file}"}, Extensions: []string{".java"}, ToolType: ToolTypeLinter, Priority: 1})

	// Kotlin tools
	r.RegisterTool(ToolConfig{Name: "ktlint", Command: "ktlint", Args: []string{"-F", "{file}"}, Extensions: []string{".kt", ".kts"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "detekt", Command: "detekt", Args: []string{"{file}"}, Extensions: []string{".kt", ".kts"}, ToolType: ToolTypeLinter, Priority: 1})

	// Swift tools
	r.RegisterTool(ToolConfig{Name: "swift-format", Command: "swift-format", Args: []string{"-i", "{file}"}, Extensions: []string{".swift"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "swiftlint", Command: "swiftlint", Args: []string{"--path", "{file}"}, Extensions: []string{".swift"}, ToolType: ToolTypeLinter, Priority: 1})

	// C/C++ tools
	r.RegisterTool(ToolConfig{Name: "clang-format", Command: "clang-format", Args: []string{"-i", "{file}"}, Extensions: []string{".c", ".cpp", ".cc", ".h", ".hpp"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "clang-tidy", Command: "clang-tidy", Args: []string{"{file}"}, Extensions: []string{".c", ".cpp", ".cc"}, ToolType: ToolTypeLinter, Priority: 1})

	// Ruby tools
	r.RegisterTool(ToolConfig{Name: "rubocop-format", Command: "rubocop", Args: []string{"-a", "{file}"}, Extensions: []string{".rb", ".rake", ".gemspec"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "rubocop", Command: "rubocop", Args: []string{"{file}"}, Extensions: []string{".rb", ".rake", ".gemspec"}, ToolType: ToolTypeLinter, Priority: 1})

	// PHP tools
	r.RegisterTool(ToolConfig{Name: "php-cs-fixer", Command: "php-cs-fixer", Args: []string{"fix", "{file}"}, Extensions: []string{".php"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "phpstan", Command: "phpstan", Args: []string{"analyse", "{file}"}, Extensions: []string{".php"}, ToolType: ToolTypeLinter, Priority: 1})

	// Elixir tools
	r.RegisterTool(ToolConfig{Name: "mix-format", Command: "mix", Args: []string{"format", "{file}"}, Extensions: []string{".ex", ".exs"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "credo", Command: "mix", Args: []string{"credo", "{file}"}, Extensions: []string{".ex", ".exs"}, ToolType: ToolTypeLinter, Priority: 1})

	// Scala tools
	r.RegisterTool(ToolConfig{Name: "scalafmt", Command: "scalafmt", Args: []string{"{file}"}, Extensions: []string{".scala", ".sc"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "scalafix", Command: "scalafix", Args: []string{"{file}"}, Extensions: []string{".scala", ".sc"}, ToolType: ToolTypeLinter, Priority: 1})

	// R tools
	r.RegisterTool(ToolConfig{Name: "styler", Command: "Rscript", Args: []string{"-e", "library(styler); style_file('{file}')"}, Extensions: []string{".r", ".R", ".Rmd"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "lintr", Command: "Rscript", Args: []string{"-e", "library(lintr); lint('{file}')"}, Extensions: []string{".r", ".R", ".Rmd"}, ToolType: ToolTypeLinter, Priority: 1})

	// Dart tools
	r.RegisterTool(ToolConfig{Name: "dart-format", Command: "dart", Args: []string{"format", "{file}"}, Extensions: []string{".dart"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "dart-analyze", Command: "dart", Args: []string{"analyze", "{file}"}, Extensions: []string{".dart"}, ToolType: ToolTypeLinter, Priority: 1})

	// C# tools
	r.RegisterTool(ToolConfig{Name: "dotnet-format", Command: "dotnet", Args: []string{"format", "{file}"}, Extensions: []string{".cs"}, ToolType: ToolTypeFormatter, Priority: 1})

	// Markdown tools
	r.RegisterTool(ToolConfig{Name: "prettier-md", Command: "prettier", Args: []string{"--write", "{file}"}, Extensions: []string{".md", ".mdx"}, ToolType: ToolTypeFormatter, Priority: 1})
	r.RegisterTool(ToolConfig{Name: "markdownlint", Command: "markdownlint", Args: []string{"{file}"}, Extensions: []string{".md", ".mdx"}, ToolType: ToolTypeLinter, Priority: 1})
}

// languageFromExtension maps file extension to programming language.
func languageFromExtension(ext string) string {
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
	case ".c":
		return "c"
	case ".cpp", ".cc", ".cxx":
		return "cpp"
	case ".h", ".hpp":
		return "c"
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
	default:
		// For unknown extensions, return extension without dot (e.g., ".test" -> "test")
		if len(ext) > 1 && ext[0] == '.' {
			return strings.ToLower(ext[1:])
		}
		return ""
	}
}

// extensionFromPath extracts the file extension from a path.
func extensionFromPath(path string) string {
	idx := strings.LastIndex(path, ".")
	if idx == -1 {
		return ""
	}
	return path[idx:]
}
