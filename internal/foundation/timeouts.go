package foundation

import "time"

// Standard timeout durations for different operations across MoAI-ADK.
const (
	// DefaultGitTimeout is the standard timeout for git operations.
	DefaultGitTimeout = 5 * time.Second

	// DefaultCLITimeout is the standard timeout for CLI tool operations (ast-grep, etc).
	DefaultCLITimeout = 60 * time.Second

	// DefaultSearchTimeout is the standard timeout for search operations.
	DefaultSearchTimeout = 120 * time.Second

	// DefaultLSPTimeout is the standard timeout for LSP operations.
	DefaultLSPTimeout = 3 * time.Second
)
