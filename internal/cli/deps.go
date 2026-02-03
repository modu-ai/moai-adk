// Package cli provides the Cobra command tree and dependency injection
// wiring for the MoAI-ADK CLI. This file defines the Dependencies struct
// (Composition Root) that wires all domain modules together.
package cli

import (
	"log/slog"

	"github.com/modu-ai/moai-adk-go/internal/config"
	"github.com/modu-ai/moai-adk-go/internal/core/git"
	"github.com/modu-ai/moai-adk-go/internal/hook"
	"github.com/modu-ai/moai-adk-go/internal/rank"
	"github.com/modu-ai/moai-adk-go/internal/update"
)

// Dependencies holds all domain-level services used by CLI commands.
// This is the Composition Root: the only place where concrete types
// are instantiated and wired together. All CLI commands access
// dependencies through interfaces only.
type Dependencies struct {
	Config        *config.ConfigManager
	Git           git.Repository
	GitBranch     git.BranchManager
	GitWorktree   git.WorktreeManager
	HookRegistry  hook.Registry
	HookProtocol  hook.Protocol
	UpdateChecker update.Checker
	UpdateOrch    update.Orchestrator
	RankClient    rank.Client
	RankCredStore rank.CredentialStore
	Logger        *slog.Logger
}

// deps is the global dependencies instance, initialized by InitDependencies.
// CLI commands access this through the package-level variable.
var deps *Dependencies

// InitDependencies creates and wires all domain dependencies.
// It should be called once during application startup.
// Dependencies that require a project root (Config, Git) are
// initialized lazily on first use or when the project root is available.
func InitDependencies() {
	logger := slog.Default()

	deps = &Dependencies{
		Config:        config.NewConfigManager(),
		HookProtocol:  hook.NewProtocol(),
		RankCredStore: rank.NewFileCredentialStore(""),
		Logger:        logger,
	}

	// Hook registry requires a ConfigProvider; use ConfigManager
	deps.HookRegistry = hook.NewRegistry(deps.Config)
}

// GetDeps returns the current Dependencies instance.
// Returns nil if InitDependencies has not been called.
func GetDeps() *Dependencies {
	return deps
}

// SetDeps replaces the global dependencies (used for testing).
func SetDeps(d *Dependencies) {
	deps = d
}
