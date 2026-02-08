package shell

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime"
)

// EnvConfigurator orchestrates shell environment configuration.
type EnvConfigurator interface {
	// Configure sets up shell environment for MoAI/Claude Code.
	// This adds CLAUDE_DISABLE_PATH_WARNING and optionally PATH setup.
	Configure(opts ConfigOptions) (*ConfigResult, error)

	// GetRecommendation returns the recommended configuration.
	GetRecommendation() *Recommendation

	// GetShellConfig returns the detected shell configuration.
	GetShellConfig() *ShellConfig
}

// envConfigurator is the default implementation of EnvConfigurator.
type envConfigurator struct {
	detector     Detector
	configurator Configurator
	logger       *slog.Logger
}

// NewEnvConfigurator creates a new EnvConfigurator instance.
func NewEnvConfigurator(logger *slog.Logger) EnvConfigurator {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &envConfigurator{
		detector:     NewDetector(),
		configurator: NewConfigurator(),
		logger:       logger,
	}
}

// GetShellConfig returns the detected shell configuration.
func (e *envConfigurator) GetShellConfig() *ShellConfig {
	return e.detector.GetShellConfig()
}

// GetRecommendation returns the recommended configuration.
func (e *envConfigurator) GetRecommendation() *Recommendation {
	config := e.detector.GetShellConfig()

	var changes []string
	var explanation string

	switch config.Shell {
	case ShellZsh:
		changes = append(changes, "export CLAUDE_DISABLE_PATH_WARNING=1")
		changes = append(changes, "export PATH=\"$HOME/.local/bin:$PATH\"")
		changes = append(changes, "export PATH=\"$HOME/go/bin:$PATH\"")
		explanation = ".zshenv is loaded for ALL shells (interactive and non-interactive), ensuring IDE compatibility"
	case ShellBash:
		changes = append(changes, "export CLAUDE_DISABLE_PATH_WARNING=1")
		changes = append(changes, "export PATH=\"$HOME/.local/bin:$PATH\"")
		changes = append(changes, "export PATH=\"$HOME/go/bin:$PATH\"")
		if config.IsWSL {
			explanation = ".profile is used because WSL runs bash as a login shell"
		} else {
			explanation = ".profile or .bash_profile is preferred for login shell compatibility"
		}
	case ShellFish:
		changes = append(changes, "set -gx CLAUDE_DISABLE_PATH_WARNING 1")
		changes = append(changes, "set -gx PATH $HOME/.local/bin $PATH")
		changes = append(changes, "set -gx PATH $HOME/go/bin $PATH")
		explanation = "config.fish is the standard configuration file for fish shell"
	case ShellPowerShell:
		changes = append(changes, "$env:CLAUDE_DISABLE_PATH_WARNING = \"1\"")
		changes = append(changes, "$env:PATH = \"$env:USERPROFILE\\.local\\bin;$env:PATH\"")
		changes = append(changes, "$env:PATH = \"$env:USERPROFILE\\go\\bin;$env:PATH\"")
		explanation = "PowerShell profile is loaded when PowerShell starts, ensuring IDE compatibility"
	default:
		changes = append(changes, "export CLAUDE_DISABLE_PATH_WARNING=1")
		changes = append(changes, "export PATH=\"$HOME/.local/bin:$PATH\"")
		changes = append(changes, "export PATH=\"$HOME/go/bin:$PATH\"")
		explanation = ".profile is used as the fallback for unknown shells"
	}

	return &Recommendation{
		Shell:       config.Shell,
		ConfigFile:  config.ConfigFile,
		Changes:     changes,
		Explanation: explanation,
	}
}

// Configure sets up shell environment for MoAI/Claude Code.
func (e *envConfigurator) Configure(opts ConfigOptions) (*ConfigResult, error) {
	config := e.detector.GetShellConfig()

	e.logger.Info("configuring shell environment",
		"shell", config.Shell,
		"configFile", config.ConfigFile,
		"isWSL", config.IsWSL,
		"platform", config.Platform,
	)

	if config.Shell == ShellUnknown {
		e.logger.Warn("unknown shell type, configuration may not work correctly")
	}

	if opts.DryRun {
		rec := e.GetRecommendation()
		return &ConfigResult{
			Success:    true,
			Message:    fmt.Sprintf("Would modify %s", config.ConfigFile),
			ConfigFile: config.ConfigFile,
			LinesAdded: rec.Changes,
		}, nil
	}

	var allResults []string
	var anySuccess bool

	// Add CLAUDE_DISABLE_PATH_WARNING
	if opts.AddClaudeWarningDisable {
		res, err := e.configurator.AddEnvVar(config.ConfigFile, "CLAUDE_DISABLE_PATH_WARNING", "1")
		if err != nil {
			if !errors.Is(err, ErrAlreadyConfigured) {
				e.logger.Error("failed to add CLAUDE_DISABLE_PATH_WARNING",
					"error", err,
					"configFile", config.ConfigFile,
				)
				return nil, fmt.Errorf("add CLAUDE_DISABLE_PATH_WARNING: %w", err)
			}
			e.logger.Debug("CLAUDE_DISABLE_PATH_WARNING already configured")
		} else {
			allResults = append(allResults, res.LinesAdded...)
			anySuccess = true
		}
	}

	// Add PATH entry for ~/.local/bin
	if opts.AddLocalBinPath {
		res, err := e.configurator.AddPathEntry(config.ConfigFile, "$HOME/.local/bin")
		if err != nil {
			if !errors.Is(err, ErrAlreadyConfigured) {
				e.logger.Error("failed to add PATH entry",
					"error", err,
					"configFile", config.ConfigFile,
				)
				return nil, fmt.Errorf("add PATH entry: %w", err)
			}
			e.logger.Debug("PATH entry for .local/bin already configured")
		} else {
			allResults = append(allResults, res.LinesAdded...)
			anySuccess = true
		}
	}

	// Add PATH entry for ~/go/bin (Go install location)
	if opts.AddGoBinPath {
		res, err := e.configurator.AddPathEntry(config.ConfigFile, "$HOME/go/bin")
		if err != nil {
			if !errors.Is(err, ErrAlreadyConfigured) {
				e.logger.Error("failed to add Go bin PATH entry",
					"error", err,
					"configFile", config.ConfigFile,
				)
				return nil, fmt.Errorf("add Go bin PATH entry: %w", err)
			}
			e.logger.Debug("PATH entry for go/bin already configured")
		} else {
			allResults = append(allResults, res.LinesAdded...)
			anySuccess = true
		}
	}

	// Determine result message
	var message string
	if anySuccess {
		message = fmt.Sprintf("Shell environment configured in %s", config.ConfigFile)
	} else {
		message = fmt.Sprintf("Shell environment already configured in %s", config.ConfigFile)
	}

	return &ConfigResult{
		Success:    true,
		Skipped:    !anySuccess,
		Message:    message,
		ConfigFile: config.ConfigFile,
		LinesAdded: allResults,
	}, nil
}

// ConfigureDefault configures the shell environment with default options.
// This is a convenience function for common use cases.
func ConfigureDefault(logger *slog.Logger) (*ConfigResult, error) {
	configurator := NewEnvConfigurator(logger)
	return configurator.Configure(ConfigOptions{
		AddClaudeWarningDisable: true,
		AddLocalBinPath:         true,
		AddGoBinPath:            true,
		PreferLoginShell:        true,
	})
}

// GetPlatformInfo returns information about the current platform.
func GetPlatformInfo() string {
	detector := NewDetector()
	config := detector.GetShellConfig()

	platform := runtime.GOOS
	if config.IsWSL {
		platform = "wsl"
	}

	return fmt.Sprintf("Platform: %s | Shell: %s | Config: %s",
		platform, config.Shell, config.ConfigFile)
}
