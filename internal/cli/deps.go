// Package cli provides the Cobra command tree and dependency injection
// wiring for the MoAI-ADK CLI. This file defines the Dependencies struct
// (Composition Root) that wires all domain modules together.
package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/modu-ai/moai-adk-go/internal/config"
	"github.com/modu-ai/moai-adk-go/internal/core/git"
	"github.com/modu-ai/moai-adk-go/internal/hook"
	"github.com/modu-ai/moai-adk-go/internal/rank"
	"github.com/modu-ai/moai-adk-go/internal/update"
	"github.com/modu-ai/moai-adk-go/pkg/version"
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

	// Register default hook handlers
	deps.HookRegistry.Register(hook.NewSessionStartHandler(deps.Config))
	deps.HookRegistry.Register(hook.NewPreToolHandler(deps.Config, hook.DefaultSecurityPolicy()))
	deps.HookRegistry.Register(hook.NewPostToolHandler())
	deps.HookRegistry.Register(hook.NewCompactHandler())
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

// EnsureGit lazily initializes Git-related dependencies.
// It should be called before using Git, GitBranch, or GitWorktree.
// Thread-safe: subsequent calls are no-ops if Git is already initialized.
func (d *Dependencies) EnsureGit(projectRoot string) error {
	if d.Git != nil {
		return nil
	}
	repo, err := git.NewRepository(projectRoot)
	if err != nil {
		return fmt.Errorf("open git repository: %w", err)
	}
	d.Git = repo
	d.GitBranch = git.NewBranchManager(repo.Root())
	d.GitWorktree = git.NewWorktreeManager(repo.Root())
	return nil
}

// EnsureUpdate lazily initializes Update-related dependencies.
// It should be called before using UpdateChecker or UpdateOrch.
// Thread-safe: subsequent calls are no-ops if UpdateChecker is already initialized.
func (d *Dependencies) EnsureUpdate() error {
	if d.UpdateChecker != nil {
		return nil
	}

	// Default GitHub releases URL for moai-adk
	// Can be overridden via MOAI_UPDATE_URL environment variable
	const defaultAPIURL = "https://api.github.com/repos/modu-ai/moai-adk/releases/latest"
	apiURL := os.Getenv("MOAI_UPDATE_URL")
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	d.UpdateChecker = update.NewChecker(apiURL, nil)

	// Get current binary path for updater and rollback
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	currentVersion := version.GetVersion()
	updater := update.NewUpdater(binaryPath, nil)
	rollback := update.NewRollback(binaryPath)
	d.UpdateOrch = update.NewOrchestrator(currentVersion, d.UpdateChecker, updater, rollback)

	return nil
}

// EnsureRank lazily initializes the Rank client.
// It should be called before using RankClient.
// Thread-safe: subsequent calls are no-ops if RankClient is already initialized.
// Returns an error if RankCredStore is not initialized or has no API key.
func (d *Dependencies) EnsureRank() error {
	if d.RankClient != nil {
		return nil
	}
	if d.RankCredStore == nil {
		return fmt.Errorf("RankCredStore not initialized")
	}
	apiKey, err := d.RankCredStore.GetAPIKey()
	if err != nil {
		return fmt.Errorf("get API key: %w", err)
	}
	if apiKey == "" {
		return fmt.Errorf("no API key configured")
	}
	d.RankClient = rank.NewClient(apiKey)
	return nil
}
