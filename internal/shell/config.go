package shell

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Configurator modifies shell configuration files.
type Configurator interface {
	// GetConfigFile returns the appropriate config file for the shell.
	// For zsh: ~/.zshenv (not .zshrc) for non-interactive shell support.
	// For bash: ~/.profile or ~/.bash_profile
	// For fish: ~/.config/fish/config.fish
	GetConfigFile(shell ShellType, preferLogin bool) string

	// AddEnvVar adds an environment variable export to the config file.
	// Returns ErrAlreadyConfigured if the variable is already set.
	AddEnvVar(configFile, varName, varValue string) (*ConfigResult, error)

	// AddPathEntry adds a PATH entry to the config file.
	// Idempotent: returns ErrAlreadyConfigured if already present.
	AddPathEntry(configFile, pathEntry string) (*ConfigResult, error)

	// HasEntry checks if a string exists in the config file.
	HasEntry(configFile, search string) (bool, error)
}

// configurator is the default implementation of Configurator.
type configurator struct {
	homeDir string
}

// NewConfigurator creates a new Configurator instance.
func NewConfigurator() Configurator {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to environment variable if UserHomeDir fails
		home = os.Getenv("HOME")
		if home == "" {
			home = os.Getenv("USERPROFILE") // Windows fallback
		}
	}
	return &configurator{homeDir: home}
}

// GetConfigFile returns the appropriate config file for the shell.
func (c *configurator) GetConfigFile(shell ShellType, preferLogin bool) string {
	return selectConfigFile(shell, preferLogin)
}

// HasEntry checks if a string exists in the config file.
func (c *configurator) HasEntry(configFile, search string) (bool, error) {
	expandedPath := expandPath(configFile)

	file, err := os.Open(expandedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("open config file: %w", err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), search) {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("read config file: %w", err)
	}

	return false, nil
}

// AddEnvVar adds an environment variable export to the config file.
func (c *configurator) AddEnvVar(configFile, varName, varValue string) (*ConfigResult, error) {
	expandedPath := expandPath(configFile)

	// Check idempotency
	has, err := c.HasEntry(expandedPath, varName)
	if err != nil {
		return nil, err
	}
	if has {
		return &ConfigResult{
			Success:    true,
			Skipped:    true,
			Message:    fmt.Sprintf("%s already configured in %s", varName, configFile),
			ConfigFile: configFile,
		}, ErrAlreadyConfigured
	}

	// Determine syntax based on shell type
	var exportLine string
	switch {
	case isPowerShellProfile(configFile):
		// PowerShell syntax: $env:VARNAME = "value"
		exportLine = fmt.Sprintf("$env:%s = \"%s\"", varName, varValue)
	case strings.Contains(configFile, "fish") || strings.HasSuffix(configFile, "config.fish"):
		// Fish syntax
		exportLine = fmt.Sprintf("set -gx %s %s", varName, varValue)
	default:
		// Bash/Zsh syntax
		exportLine = fmt.Sprintf("export %s=%s", varName, varValue)
	}

	// Build content block
	content := fmt.Sprintf("\n# Added by MoAI-ADK\n%s\n", exportLine)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(expandedPath), 0o755); err != nil {
		return nil, fmt.Errorf("create config directory: %w", err)
	}

	// Append to config file
	f, err := os.OpenFile(expandedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		if os.IsPermission(err) {
			return nil, ErrPermissionDenied
		}
		return nil, fmt.Errorf("open config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(content); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}

	return &ConfigResult{
		Success:    true,
		Message:    fmt.Sprintf("Added %s to %s", varName, configFile),
		ConfigFile: configFile,
		LinesAdded: []string{exportLine},
	}, nil
}

// AddPathEntry adds a PATH entry to the config file.
func (c *configurator) AddPathEntry(configFile, pathEntry string) (*ConfigResult, error) {
	expandedPath := expandPath(configFile)

	// Determine search pattern based on the path entry
	var searchPattern string
	switch {
	case strings.Contains(pathEntry, "go/bin"):
		searchPattern = "go/bin"
	case strings.Contains(pathEntry, ".local/bin"):
		searchPattern = ".local/bin"
	default:
		// Extract the last component of the path for generic check
		searchPattern = pathEntry
	}

	// Check idempotency - look for the specific path entry
	has, err := c.HasEntry(expandedPath, searchPattern)
	if err != nil {
		return nil, err
	}
	if has {
		return &ConfigResult{
			Success:    true,
			Skipped:    true,
			Message:    fmt.Sprintf("PATH (%s) already configured in %s", searchPattern, configFile),
			ConfigFile: configFile,
		}, ErrAlreadyConfigured
	}

	// Determine syntax based on shell type
	var exportLine string
	switch {
	case isPowerShellProfile(configFile):
		// PowerShell syntax: prepend to PATH with Windows path separator
		// Convert Unix path to Windows path for PowerShell
		winPath := strings.ReplaceAll(pathEntry, "$HOME", "$env:USERPROFILE")
		winPath = strings.ReplaceAll(winPath, "/", "\\")
		exportLine = fmt.Sprintf("$env:PATH = \"%s;$env:PATH\"", winPath)
	case strings.Contains(configFile, "fish") || strings.HasSuffix(configFile, "config.fish"):
		// Fish syntax
		exportLine = fmt.Sprintf("set -gx PATH %s $PATH", pathEntry)
	default:
		// Bash/Zsh syntax
		exportLine = fmt.Sprintf("export PATH=\"%s:$PATH\"", pathEntry)
	}

	// Build content block
	content := fmt.Sprintf("\n# Added by MoAI-ADK\n%s\n", exportLine)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(expandedPath), 0o755); err != nil {
		return nil, fmt.Errorf("create config directory: %w", err)
	}

	// Append to config file
	f, err := os.OpenFile(expandedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		if os.IsPermission(err) {
			return nil, ErrPermissionDenied
		}
		return nil, fmt.Errorf("open config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(content); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}

	return &ConfigResult{
		Success:    true,
		Message:    fmt.Sprintf("Added PATH entry to %s", configFile),
		ConfigFile: configFile,
		LinesAdded: []string{exportLine},
	}, nil
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// isPowerShellProfile checks if the config file is a PowerShell profile.
func isPowerShellProfile(configFile string) bool {
	return strings.HasSuffix(configFile, ".ps1") ||
		strings.Contains(configFile, "PowerShell") ||
		strings.Contains(configFile, "WindowsPowerShell")
}
