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
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/astgrep"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/hook/security"
	"github.com/modu-ai/moai-adk/internal/loop"
	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
	"github.com/modu-ai/moai-adk/internal/ralph"
	"github.com/modu-ai/moai-adk/internal/resilience"
	"github.com/modu-ai/moai-adk/internal/update"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// Dependencies holds all domain-level services used by CLI commands.
// This is the Composition Root: the only place where concrete types
// are instantiated and wired together. All CLI commands access
// dependencies through interfaces only.
type Dependencies struct {
	Config         *config.ConfigManager
	Git            git.Repository
	GitBranch      git.BranchManager
	GitWorktree    git.WorktreeManager
	HookRegistry   hook.Registry
	HookProtocol   hook.Protocol
	UpdateChecker  update.Checker
	UpdateOrch     update.Orchestrator
	LoopController *loop.LoopController
	Logger         *slog.Logger
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
	// Replace the default logger with discard to prevent slog output from hook handlers leaking to stderr.
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	slog.SetDefault(logger)

	// Initialize Ralph engine and loop controller.
	ralphCfg := config.NewDefaultRalphConfig()
	ralphEngine := ralph.NewRalphEngine(ralphCfg)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}
	loopStorage := loop.NewFileStorage(filepath.Join(homeDir, ".moai", "state", "loop"))
	loopCtrl := loop.NewLoopController(loopStorage, ralphEngine, &noopFeedbackGenerator{}, ralphCfg.MaxIterations)

	deps = &Dependencies{
		Config:         config.NewConfigManager(),
		HookProtocol:   hook.NewProtocol(),
		LoopController: loopCtrl,
		Logger:         logger,
	}

	// Hook registry requires a ConfigProvider; use ConfigManager
	reg := hook.NewRegistry(deps.Config)
	deps.HookRegistry = reg

	// Determine current working directory once.
	cwd, _ := os.Getwd()

	// Enable observability when configured (REQ-OBS-001).
	// Reads the project-relative observability config; gracefully skips on error.
	enableObservabilityIfConfigured(reg, cwd)

	// Create security scanner for AST-based scanning
	securityScanner := security.NewSecurityScanner()

	// Apply circuit breaker to LSP fallback tool execution.
	// Allows go vet/golangci-lint to skip quickly on repeated failures.
	lspCircuitBreaker := resilience.NewCircuitBreaker(resilience.CircuitBreakerConfig{
		Threshold: 3,
		Timeout:   30 * time.Second,
	})

	// Initialize LSP diagnostics collector and AST analyzer.
	// LSP client is nil (not yet integrated); fallback CLI tools are used.
	fallbackDiags := lsphook.NewFallbackDiagnosticsWithCircuitBreaker(lspCircuitBreaker)
	diagnosticsCollector := lsphook.NewDiagnosticsCollector(nil, fallbackDiags)

	// Initialize ast-grep analyzer (ScanFile returns empty results if sg CLI is absent)
	astAnalyzer := astgrep.NewAnalyzer(cwd)

	// Register default hook handlers
	deps.HookRegistry.Register(hook.NewSessionStartHandler(deps.Config))
	// SessionEnd handler: use observability-aware variant when configured.
	deps.HookRegistry.Register(buildSessionEndHandler(cwd))

	// Register auto-update handler for SessionStart
	deps.HookRegistry.Register(hook.NewAutoUpdateHandler(buildAutoUpdateFunc()))

	deps.HookRegistry.Register(hook.NewStopHandler())
	// Build security policy: defaults + extra patterns from security.yaml (REQ-SEC-003).
	secPolicy := hook.DefaultSecurityPolicy()
	secPolicy.MergeExtraPatterns(security.LoadExtraSecurityConfig(cwd))
	deps.HookRegistry.Register(hook.NewPreToolHandlerWithScanner(deps.Config, secPolicy, securityScanner))
	deps.HookRegistry.Register(hook.NewPostToolHandlerWithAstgrep(diagnosticsCollector, astAnalyzer))
	deps.HookRegistry.Register(hook.NewCompactHandler())
	deps.HookRegistry.Register(hook.NewPostToolUseFailureHandler())
	deps.HookRegistry.Register(hook.NewNotificationHandler())
	deps.HookRegistry.Register(hook.NewSubagentStartHandler())
	deps.HookRegistry.Register(hook.NewUserPromptSubmitHandler(deps.Config))
	deps.HookRegistry.Register(hook.NewPermissionRequestHandler())
	deps.HookRegistry.Register(hook.NewTeammateIdleHandler())
	deps.HookRegistry.Register(hook.NewTaskCompletedHandler())
	deps.HookRegistry.Register(hook.NewWorktreeCreateHandler())
	deps.HookRegistry.Register(hook.NewWorktreeRemoveHandler())
	deps.HookRegistry.Register(hook.NewPostCompactHandler())
	deps.HookRegistry.Register(hook.NewInstructionsLoadedHandler())
	deps.HookRegistry.Register(hook.NewStopFailureHandler())
	deps.HookRegistry.Register(hook.NewSubagentStopHandler())
	deps.HookRegistry.Register(hook.NewTaskCreatedHandler())
	deps.HookRegistry.Register(hook.NewPermissionDeniedHandler())
	deps.HookRegistry.Register(hook.NewConfigChangeHandler())
	deps.HookRegistry.Register(hook.NewCwdChangedHandler())
}

// enableObservabilityIfConfigured reads observability config and enables
// trace writing on the registry if configured. Gracefully skips on error.
func enableObservabilityIfConfigured(reg hook.Registry, cwd string) {
	cfgPath := filepath.Join(cwd, ".moai", "config", "sections", "observability.yaml")
	if _, err := os.Stat(cfgPath); err != nil {
		return // Config file not found — observability disabled.
	}
	// Enable observability via type assertion; concrete registry supports it.
	type observabilityEnabler interface {
		EnableObservability(logDir string)
	}
	if oe, ok := reg.(observabilityEnabler); ok {
		logDir := filepath.Join(cwd, ".moai", "logs")
		_ = os.MkdirAll(logDir, 0o755)
		oe.EnableObservability(logDir)
	}
}

// buildSessionEndHandler creates a SessionEnd handler with observability
// support when the reports directory is available.
func buildSessionEndHandler(cwd string) hook.Handler {
	reportDir := filepath.Join(cwd, ".moai", "reports")
	traceDir := filepath.Join(cwd, ".moai", "logs")
	// Only use observability variant if trace dir exists.
	if _, err := os.Stat(traceDir); err == nil {
		_ = os.MkdirAll(reportDir, 0o755)
		return hook.NewSessionEndHandlerWithObservability(traceDir, reportDir)
	}
	return hook.NewSessionEndHandler()
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

// noopFeedbackGenerator is a default implementation that returns an empty Feedback without
// any actual collection. Used to satisfy loop.FeedbackGenerator when no feedback source
// is available during CLI execution.
type noopFeedbackGenerator struct{}

func (n *noopFeedbackGenerator) Collect(_ context.Context) (*loop.Feedback, error) {
	return &loop.Feedback{}, nil
}

