package quality

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// WorktreeValidator provides TRUST 5 validation scoped to a worktree workspace.
type WorktreeValidator interface {
	// Validate runs full TRUST 5 quality gates for a worktree directory.
	Validate(ctx context.Context, wtPath string) (*Report, error)

	// ValidateWithConfig runs quality gates using a specific configuration override.
	ValidateWithConfig(ctx context.Context, wtPath string, config QualityConfig) (*Report, error)
}

// GateFactory creates a quality Gate for a given configuration.
// This allows the worktree validator to be tested with mock gates.
type GateFactory func(config QualityConfig) Gate

// worktreeValidator implements WorktreeValidator using the existing TrustGate framework.
type worktreeValidator struct {
	gateFactory GateFactory
	config      QualityConfig
	logger      *slog.Logger
}

// Compile-time interface compliance check.
var _ WorktreeValidator = (*worktreeValidator)(nil)

// NewWorktreeValidator creates a validator scoped to worktree directories.
// The gateFactory creates a quality Gate for each validation run, enabling
// dependency injection for testing. Returns an error if gateFactory is nil.
func NewWorktreeValidator(gateFactory GateFactory, config QualityConfig, logger *slog.Logger) (*worktreeValidator, error) {
	if gateFactory == nil {
		return nil, fmt.Errorf("worktree validator: gate factory must not be nil")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &worktreeValidator{
		gateFactory: gateFactory,
		config:      config,
		logger:      logger.With("module", "worktree-validator"),
	}, nil
}

// DefaultGateFactory creates a GateFactory that builds TrustGate instances
// with the provided LSP client.
func DefaultGateFactory(lsp LSPClient) GateFactory {
	return func(config QualityConfig) Gate {
		validators := []Validator{
			NewTestedValidator(lsp, config.TestCoverageTarget, 0),
			NewReadableValidator(lsp),
		}
		return NewTrustGate(config, validators,
			WithPhase(PhaseRun),
			WithLSPClient(lsp),
		)
	}
}

// Validate runs TRUST 5 quality gates for the specified worktree directory.
func (v *worktreeValidator) Validate(ctx context.Context, wtPath string) (*Report, error) {
	return v.ValidateWithConfig(ctx, wtPath, v.config)
}

// ValidateWithConfig runs quality gates with a custom configuration.
func (v *worktreeValidator) ValidateWithConfig(ctx context.Context, wtPath string, config QualityConfig) (*Report, error) {
	if err := validateWorktreePath(wtPath); err != nil {
		return nil, fmt.Errorf("validate worktree path %q: %w", wtPath, err)
	}

	v.logger.Info("starting worktree quality validation",
		"path", wtPath,
		"development_mode", string(config.DevelopmentMode),
		"coverage_target", config.TestCoverageTarget,
	)

	gate := v.gateFactory(config)

	report, err := gate.Validate(ctx)
	if err != nil {
		return nil, fmt.Errorf("trust gate validation in %q: %w", wtPath, err)
	}

	v.logger.Info("worktree quality validation complete",
		"path", wtPath,
		"passed", report.Passed,
		"score", report.Score,
	)

	return report, nil
}

// validateWorktreePath checks that the directory exists and is accessible.
func validateWorktreePath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %w", err)
		}
		return fmt.Errorf("stat: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}
	return nil
}
