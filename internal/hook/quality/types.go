package quality

import "time"

// ToolType represents the type of code quality tool.
type ToolType string

const (
	// ToolTypeFormatter is a code formatting tool.
	ToolTypeFormatter ToolType = "formatter"

	// ToolTypeLinter is a code linting tool.
	ToolTypeLinter ToolType = "linter"

	// ToolTypeTypeChecker is a static type checking tool.
	ToolTypeTypeChecker ToolType = "type_checker"

	// ToolTypeASTAnalyzer is an AST-based code analyzer.
	ToolTypeASTAnalyzer ToolType = "ast_analyzer"
)

// FileArgsPosition specifies where the file path should be placed in tool arguments.
type FileArgsPosition string

const (
	// FileArgsPositionEnd places file path at the end of arguments.
	FileArgsPositionEnd FileArgsPosition = "end"

	// FileArgsPositionStart places file path at the beginning of arguments.
	FileArgsPositionStart FileArgsPosition = "start"

	// FileArgsPositionReplace replaces a placeholder with the file path.
	// Format: "replace:{placeholder}"
	FileArgsPositionReplace FileArgsPosition = "replace"
)

// ToolConfig represents configuration for a single code quality tool.
type ToolConfig struct {
	// Name is the unique identifier for this tool (e.g., "ruff-format", "gofmt").
	Name string

	// Command is the executable name or path (e.g., "ruff", "gofmt").
	Command string

	// Args are the base arguments passed to the command.
	Args []string

	// FileArgsPosition specifies where to insert the file path in the command.
	FileArgsPosition FileArgsPosition

	// CheckArgs are arguments to verify the tool exists (optional).
	CheckArgs []string

	// FixArgs are arguments to auto-fix issues (optional).
	FixArgs []string

	// Extensions are the file extensions this tool handles.
	Extensions []string

	// ToolType is the category of tool (formatter, linter, etc.).
	ToolType ToolType

	// Priority determines tool selection order (lower = higher priority).
	Priority int

	// TimeoutSeconds is the maximum execution time.
	TimeoutSeconds int

	// RequiresConfig indicates if the tool needs a config file to work.
	RequiresConfig bool

	// ConfigFiles are the config file names to check for.
	ConfigFiles []string
}

// ToolResult represents the result of executing a code quality tool.
type ToolResult struct {
	// Success indicates if the tool executed successfully.
	Success bool

	// ToolName is the name of the tool that was executed.
	ToolName string

	// Output contains the stdout from the tool.
	Output string

	// Error contains the stderr from the tool.
	Error string

	// ExitCode is the process exit code.
	ExitCode int

	// FileModified indicates if the file was changed by the tool.
	FileModified bool

	// IssuesFound is the count of issues detected (for linters).
	IssuesFound int

	// IssuesFixed is the count of issues auto-fixed (for linters with --fix).
	IssuesFixed int

	// ExecutionTime is how long the tool took to run.
	ExecutionTime time.Duration
}

// LanguageInfo contains metadata about a programming language.
type LanguageInfo struct {
	// Name is the language identifier (e.g., "python", "go").
	Name string

	// Extensions are the file extensions for this language.
	Extensions []string

	// Formatters are the available formatting tools.
	Formatters []string

	// Linters are the available linting tools.
	Linters []string
}

// Skip patterns for code quality operations.

// SkipExtensions are file extensions that should skip formatting/linting.
var SkipExtensions = map[string]bool{
	".json":    true, // JSON formatting can break configs
	".yaml":    true,
	".yml":     true,
	".toml":    true,
	".lock":    true, // Lock files should not be modified
	".min.js":  true, // Minified files
	".min.css": true,
	".map":     true, // Source maps
	".svg":     true, // Binary images
	".png":     true,
	".jpg":     true,
	".gif":     true,
	".jpeg":    true,
	".ico":     true,
	".webp":    true,
	".woff":    true, // Fonts
	".woff2":   true,
	".ttf":     true,
	".eot":     true,
	".otf":     true,
	".exe":     true, // Binaries
	".dll":     true,
	".so":      true,
	".dylib":   true,
	".bin":     true,
}

// SkipDirectories are directory names that should skip formatting/linting.
var SkipDirectories = map[string]bool{
	"node_modules":  true, // JavaScript dependencies
	".git":          true, // VCS
	".svn":          true,
	".hg":           true,
	".venv":         true, // Python virtual environments
	"venv":          true,
	"__pycache__":   true, // Python cache
	".pytest_cache": true,
	".cache":        true, // General cache
	"dist":          true, // Build outputs
	"build":         true,
	"target":        true, // Rust/Java build
	".next":         true, // Next.js
	".nuxt":         true, // Nuxt.js
	"out":           true, // Build outputs
	"vendor":        true, // PHP/Go dependencies
	".idea":         true, // IDE files
	".vscode":       true,
	".eclipse":      true,
}

// Default timeout for tool execution.
const DefaultTimeoutSeconds = 30

// Maximum number of lint issues to report to Claude.
const MaxIssuesToReport = 5

// Hash cache TTL to prevent memory leaks.
const HashCacheTTL = 5 * time.Minute
