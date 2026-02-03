package hook

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
)

// executionContract implements the Contract interface per ADR-012.
// It validates that the hook execution environment meets the required
// guarantees before dispatching events.
type executionContract struct {
	workDir string
}

// NewContract creates a new Contract for the given working directory.
func NewContract(workDir string) Contract {
	return &executionContract{workDir: workDir}
}

// Validate checks that the execution environment meets contract requirements.
// It verifies:
//   - The context is not already cancelled or expired
//   - The working directory is specified and accessible
//
// Returns ErrHookContractFail if any check fails.
func (c *executionContract) Validate(ctx context.Context) error {
	slog.Debug("validating hook execution contract",
		"work_dir", c.workDir,
		"os", runtime.GOOS,
		"arch", runtime.GOARCH,
	)

	// Check context is still valid
	if ctx.Err() != nil {
		slog.Error("contract validation failed: context already done",
			"error", ctx.Err().Error(),
		)
		return fmt.Errorf("%w: context already done: %v", ErrHookContractFail, ctx.Err())
	}

	// Check working directory is specified
	if c.workDir == "" {
		slog.Error("contract validation failed: empty working directory")
		return fmt.Errorf("%w: working directory not specified", ErrHookContractFail)
	}

	// Check working directory exists and is accessible
	info, err := os.Stat(c.workDir)
	if err != nil {
		slog.Error("contract validation failed: working directory not accessible",
			"work_dir", c.workDir,
			"error", err.Error(),
		)
		return fmt.Errorf("%w: working directory not accessible: %v", ErrHookContractFail, err)
	}

	if !info.IsDir() {
		slog.Error("contract validation failed: path is not a directory",
			"work_dir", c.workDir,
		)
		return fmt.Errorf("%w: path is not a directory: %s", ErrHookContractFail, c.workDir)
	}

	slog.Debug("hook execution contract validated successfully")
	return nil
}

// Guarantees returns the list of guaranteed execution conditions per ADR-012.
func (c *executionContract) Guarantees() []string {
	return []string{
		"stdin: valid JSON conforming to Claude Code hook protocol",
		"exit code: 0 (allow/success), 2 (block), other (non-blocking error)",
		"timeout: configurable via context.WithTimeout (default 30s)",
		"config access: same binary, same config loader as CLI commands",
		"working directory: project root ($CLAUDE_PROJECT_DIR)",
	}
}

// NonGuarantees returns the list of non-guaranteed execution conditions per ADR-012.
func (c *executionContract) NonGuarantees() []string {
	return []string{
		"User PATH: moai binary must be in system PATH",
		"shell environment variables: .bashrc/.zshrc are not loaded",
		"shell functions or alias definitions",
		"Python/Node.js/uv runtime availability",
	}
}
