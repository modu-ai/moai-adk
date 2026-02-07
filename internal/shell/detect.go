package shell

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Detector detects shell type and configuration.
type Detector interface {
	// DetectShell returns the user's default shell.
	DetectShell() ShellType

	// IsWSL returns true if running in Windows Subsystem for Linux.
	IsWSL() bool

	// GetShellConfig returns the full shell configuration.
	GetShellConfig() *ShellConfig
}

// detector is the default implementation of Detector.
type detector struct {
	// envGetter allows injecting custom env for testing.
	// If nil, os.Getenv is used.
	envGetter func(string) string
}

// NewDetector creates a new Detector instance.
func NewDetector() Detector {
	return &detector{}
}

// newDetectorWithEnv creates a Detector with a custom environment getter for testing.
func newDetectorWithEnv(getter func(string) string) Detector {
	return &detector{envGetter: getter}
}

// getEnv returns the environment variable value.
func (d *detector) getEnv(key string) string {
	if d.envGetter != nil {
		return d.envGetter(key)
	}
	return os.Getenv(key)
}

// DetectShell checks environment to determine the shell type.
// On Unix systems, it checks $SHELL. On Windows, it checks for PowerShell.
func (d *detector) DetectShell() ShellType {
	// Check for Windows PowerShell first
	if runtime.GOOS == "windows" || d.getEnv("OS") == "Windows_NT" {
		// PowerShell sets PSModulePath environment variable
		if d.getEnv("PSModulePath") != "" {
			return ShellPowerShell
		}
		// Also check for pwsh (PowerShell Core) on any platform
		if d.getEnv("POWERSHELL_DISTRIBUTION_CHANNEL") != "" {
			return ShellPowerShell
		}
	}

	// Check for Git Bash / MSYS2 on Windows (sets MSYSTEM env var)
	if msystem := d.getEnv("MSYSTEM"); msystem != "" {
		return ShellBash
	}

	shellPath := d.getEnv("SHELL")
	if shellPath == "" {
		// On Windows without $SHELL, default to PowerShell if PSModulePath exists
		if d.getEnv("PSModulePath") != "" {
			return ShellPowerShell
		}
		return ShellUnknown
	}

	shellPath = strings.ToLower(shellPath)

	switch {
	case strings.Contains(shellPath, "zsh"):
		return ShellZsh
	case strings.Contains(shellPath, "bash"):
		return ShellBash
	case strings.Contains(shellPath, "fish"):
		return ShellFish
	case strings.Contains(shellPath, "pwsh") || strings.Contains(shellPath, "powershell"):
		return ShellPowerShell
	default:
		return ShellUnknown
	}
}

// IsWSL checks for WSL environment variables.
// Works for both WSL 1 and WSL 2.
func (d *detector) IsWSL() bool {
	// WSL sets these environment variables
	return d.getEnv("WSL_DISTRO_NAME") != "" ||
		d.getEnv("WSLENV") != "" ||
		d.getEnv("WSL_INTEROP") != ""
}

// GetShellConfig returns the complete shell configuration.
func (d *detector) GetShellConfig() *ShellConfig {
	shell := d.DetectShell()
	isWSL := d.IsWSL()

	// Determine platform
	platform := runtime.GOOS
	if isWSL {
		platform = "wsl"
	}

	// Select config file based on shell and platform
	// Prefer login shell configs for WSL/Linux/macOS for GUI app compatibility
	preferLogin := isWSL || platform == "linux" || platform == "darwin"
	configFile := selectConfigFile(shell, preferLogin)

	return &ShellConfig{
		Shell:      shell,
		ConfigFile: configFile,
		IsWSL:      isWSL,
		Platform:   platform,
	}
}

// selectConfigFile returns the appropriate config file for the shell.
func selectConfigFile(shell ShellType, preferLogin bool) string {
	home := os.Getenv("HOME")
	if home == "" {
		// Fallback for Windows
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		home = "~"
	}

	switch shell {
	case ShellZsh:
		// .zshenv is loaded for ALL shells (interactive and non-interactive)
		// This is critical for IDE/non-interactive contexts
		if preferLogin {
			return filepath.Join(home, ".zshenv")
		}
		return filepath.Join(home, ".zshrc")

	case ShellBash:
		// For non-interactive (login) shell support, prefer .profile
		// WSL uses login shell, so .profile is more reliable
		if preferLogin {
			// Check if .profile exists
			profilePath := filepath.Join(home, ".profile")
			if _, err := os.Stat(profilePath); err == nil {
				return profilePath
			}
			// Check .bash_profile
			bashProfilePath := filepath.Join(home, ".bash_profile")
			if _, err := os.Stat(bashProfilePath); err == nil {
				return bashProfilePath
			}
			// Default to .profile even if it doesn't exist
			return profilePath
		}
		return filepath.Join(home, ".bashrc")

	case ShellFish:
		return filepath.Join(home, ".config", "fish", "config.fish")

	case ShellPowerShell:
		// PowerShell profile path
		// Windows PowerShell: Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1
		// PowerShell Core: Documents\PowerShell\Microsoft.PowerShell_profile.ps1
		return getPowerShellProfilePath(home)

	default:
		// On Windows, default to PowerShell profile
		if runtime.GOOS == "windows" {
			return getPowerShellProfilePath(home)
		}
		// Default to .profile for unknown shells on Unix
		return filepath.Join(home, ".profile")
	}
}

// getPowerShellProfilePath returns the PowerShell profile path.
// It checks for PowerShell Core first, then Windows PowerShell.
func getPowerShellProfilePath(home string) string {
	// PowerShell Core profile (cross-platform)
	psCorePath := filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	if _, err := os.Stat(psCorePath); err == nil {
		return psCorePath
	}

	// Check if PowerShell Core directory exists
	psCoreDir := filepath.Join(home, "Documents", "PowerShell")
	if _, err := os.Stat(psCoreDir); err == nil {
		return psCorePath
	}

	// Windows PowerShell profile
	winPSPath := filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	if _, err := os.Stat(winPSPath); err == nil {
		return winPSPath
	}

	// Check if Windows PowerShell directory exists
	winPSDir := filepath.Join(home, "Documents", "WindowsPowerShell")
	if _, err := os.Stat(winPSDir); err == nil {
		return winPSPath
	}

	// Default to PowerShell Core path (modern default)
	return psCorePath
}

// IsCmdPrompt checks if running in Windows Command Prompt (not PowerShell).
// Command Prompt is not supported by MoAI-ADK.
func IsCmdPrompt() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// PowerShell sets PSModulePath environment variable
	hasPSModulePath := os.Getenv("PSModulePath") != ""

	// In CMD, PROMPT is usually set to $P$G
	prompt := os.Getenv("PROMPT")
	isCmdPrompt := prompt == "$P$G"

	// If PSModulePath exists, it's PowerShell
	// If it doesn't exist and PROMPT is $P$G, it's CMD
	return !hasPSModulePath && isCmdPrompt
}
