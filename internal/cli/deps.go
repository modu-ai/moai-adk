// Package cli provides the Cobra command tree and dependency injection
// wiring for the MoAI-ADK CLI. This file defines the Dependencies struct
// (Composition Root) that wires all domain modules together.
package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/hook/security"
	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
	"github.com/modu-ai/moai-adk/internal/rank"
	"github.com/modu-ai/moai-adk/internal/update"
	"github.com/modu-ai/moai-adk/pkg/version"
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
	RankBrowser   rank.BrowserOpener
	Logger        *slog.Logger
}

// deps is the global dependencies instance, initialized by InitDependencies.
// CLI commands access this through the package-level variable.
var deps *Dependencies

// @MX:ANCHOR: [AUTO] InitDependencies is the Composition Root that wires all domain modules
// @MX:REASON: [AUTO] fan_in=5, called from root.go, deps_test.go, integration_test.go, hook_e2e_test.go, deps.go
// InitDependencies creates and wires all domain dependencies.
// It should be called once during application startup.
// Dependencies that require a project root (Config, Git) are
// initialized lazily on first use or when the project root is available.
func InitDependencies() {
	// Disable JSON logging for CLI commands by using a no-op logger
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	deps = &Dependencies{
		Config:        config.NewConfigManager(),
		HookProtocol:  hook.NewProtocol(),
		RankCredStore: rank.NewFileCredentialStore(""),
		Logger:        logger,
	}

	// Hook registry requires a ConfigProvider; use ConfigManager
	deps.HookRegistry = hook.NewRegistry(deps.Config)

	// Create security scanner for AST-based scanning
	securityScanner := security.NewSecurityScanner()

	// Create LSP diagnostics collector with fallback tools
	// LSP client is nil (not yet integrated), but fallback CLI tools will work
	fallbackDiags := lsphook.NewFallbackDiagnostics()
	diagnosticsCollector := lsphook.NewDiagnosticsCollector(nil, fallbackDiags)

	// Register default hook handlers
	deps.HookRegistry.Register(hook.NewSessionStartHandler(deps.Config))
	deps.HookRegistry.Register(hook.NewSessionEndHandler())

	// Register rank session handler if credentials exist
	rankHandler, err := hook.EnsureRankSessionHandler()
	if err != nil {
		logger.Warn("failed to initialize rank session handler", "error", err)
	} else if rankHandler != nil {
		deps.HookRegistry.Register(rankHandler)
		logger.Info("rank session handler registered")
	}

	// Register auto-update handler for SessionStart
	deps.HookRegistry.Register(hook.NewAutoUpdateHandler(buildAutoUpdateFunc()))

	deps.HookRegistry.Register(hook.NewStopHandler())
	deps.HookRegistry.Register(hook.NewPreToolHandlerWithScanner(deps.Config, hook.DefaultSecurityPolicy(), securityScanner))
	deps.HookRegistry.Register(hook.NewPostToolHandlerWithDiagnostics(diagnosticsCollector))
	deps.HookRegistry.Register(hook.NewCompactHandler())
	deps.HookRegistry.Register(hook.NewPostToolUseFailureHandler())
	deps.HookRegistry.Register(hook.NewNotificationHandler())
	deps.HookRegistry.Register(hook.NewSubagentStartHandler())
	deps.HookRegistry.Register(hook.NewUserPromptSubmitHandler())
	deps.HookRegistry.Register(hook.NewPermissionRequestHandler())
	deps.HookRegistry.Register(hook.NewTeammateIdleHandler())
	deps.HookRegistry.Register(hook.NewTaskCompletedHandler())
	deps.HookRegistry.Register(hook.NewWorktreeCreateHandler())
	deps.HookRegistry.Register(hook.NewWorktreeRemoveHandler())
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

	// Determine the appropriate update source based on environment variable
	// - MOAI_UPDATE_SOURCE=local: use local file-based releases
	// - MOAI_UPDATE_URL: custom GitHub API URL
	// - Default: GitHub releases based on version
	currentVersion := version.GetVersion()
	updateSource := os.Getenv("MOAI_UPDATE_SOURCE")

	// Get current binary path for updater and rollback
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	if updateSource == "local" {
		// Local file-based updates
		localConfig := update.LocalConfig{
			ReleasesDir:    os.Getenv("MOAI_RELEASES_DIR"),
			CurrentVersion: currentVersion,
		}
		d.UpdateChecker = update.NewLocalChecker(localConfig)
		d.UpdateOrch = update.NewOrchestrator(
			currentVersion,
			d.UpdateChecker,
			update.NewLocalUpdater(localConfig.ReleasesDir, binaryPath),
			update.NewRollback(binaryPath),
		)
		return nil
	}

	// Remote GitHub updates
	apiURL := os.Getenv("MOAI_UPDATE_URL")
	if apiURL == "" {
		// Check if this is a development or pre-release version
		isDevVersion := currentVersion == "dev" ||
			strings.Contains(currentVersion, "rc") ||
			strings.Contains(currentVersion, "alpha") ||
			strings.Contains(currentVersion, "beta") ||
			strings.HasPrefix(currentVersion, "go-v")

		if isDevVersion {
			// Dev/RC version: use moai-go-v2 branch releases (tagged with go-v prefix)
			apiURL = "https://api.github.com/repos/modu-ai/moai-adk/releases"
		} else {
			// Production version: use main branch releases
			apiURL = "https://api.github.com/repos/modu-ai/moai-adk/releases/latest"
		}
	}

	d.UpdateChecker = update.NewChecker(apiURL, nil)
	updater := update.NewUpdater(binaryPath, nil)
	rollback := update.NewRollback(binaryPath)
	d.UpdateOrch = update.NewOrchestrator(currentVersion, d.UpdateChecker, updater, rollback)

	return nil
}

// buildAutoUpdateFunc creates the callback that performs binary self-update.
// It uses a closure to avoid circular dependencies between hook and update.
func buildAutoUpdateFunc() hook.AutoUpdateFunc {
	return func(ctx context.Context) (*hook.AutoUpdateResult, error) {
		currentVersion := version.GetVersion()

		// Skip dev builds
		isDevBuild := strings.Contains(currentVersion, "dirty") ||
			currentVersion == "dev" ||
			strings.Contains(currentVersion, "none")
		if isDevBuild {
			return &hook.AutoUpdateResult{Updated: false}, nil
		}

		// Check cache first
		cache := update.NewCache("", 0)
		if entry := cache.Get(currentVersion); entry != nil {
			if !entry.Available {
				return &hook.AutoUpdateResult{Updated: false}, nil
			}
			// Cache says update available, proceed to update
		}

		// Initialize update system
		if deps != nil {
			if err := deps.EnsureUpdate(); err != nil {
				if deps.Logger != nil {
					deps.Logger.Debug("auto-update: failed to initialize update system", "error", err)
				}
				return nil, err
			}
		}

		if deps == nil || deps.UpdateChecker == nil {
			return &hook.AutoUpdateResult{Updated: false}, nil
		}

		// Check for available update via GitHub API
		available, info, err := deps.UpdateChecker.IsUpdateAvailable(currentVersion)
		if err != nil {
			// Cache the failure so we don't retry on every session
			_ = cache.Set(&update.CacheEntry{
				CheckedAt:  time.Now(),
				Available:  false,
				CurrentVer: currentVersion,
			})
			if deps.Logger != nil {
				deps.Logger.Debug("auto-update: version check failed", "error", err)
			}
			return nil, err
		}

		// Cache the result
		cacheEntry := &update.CacheEntry{
			CheckedAt:  time.Now(),
			Available:  available,
			CurrentVer: currentVersion,
		}
		if info != nil {
			cacheEntry.LatestInfo = info
		}
		_ = cache.Set(cacheEntry)

		if !available {
			return &hook.AutoUpdateResult{Updated: false}, nil
		}

		// Perform the update
		if deps.UpdateOrch == nil {
			return &hook.AutoUpdateResult{Updated: false}, nil
		}

		result, err := deps.UpdateOrch.Update(ctx)
		if err != nil {
			if deps.Logger != nil {
				deps.Logger.Debug("auto-update: update failed", "error", err)
			}
			return nil, err
		}

		return &hook.AutoUpdateResult{
			Updated:         true,
			PreviousVersion: result.PreviousVersion,
			NewVersion:      result.NewVersion,
		}, nil
	}
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
