package shell

// ShellType represents supported shell types.
type ShellType string

const (
	// ShellZsh is the Z shell.
	ShellZsh ShellType = "zsh"
	// ShellBash is the Bourne Again shell.
	ShellBash ShellType = "bash"
	// ShellFish is the Friendly Interactive Shell.
	ShellFish ShellType = "fish"
	// ShellPowerShell is Windows PowerShell or PowerShell Core.
	ShellPowerShell ShellType = "powershell"
	// ShellUnknown is an unrecognized shell.
	ShellUnknown ShellType = "unknown"
)

// String returns the string representation of the shell type.
func (s ShellType) String() string {
	return string(s)
}

// ShellConfig holds shell configuration information.
type ShellConfig struct {
	Shell      ShellType // Detected shell type
	ConfigFile string    // Path to the config file (e.g., ~/.zshenv)
	IsWSL      bool      // Whether running in WSL
	Platform   string    // darwin, linux, windows
}

// ConfigOptions configures the shell environment setup.
type ConfigOptions struct {
	AddClaudeWarningDisable bool // Add CLAUDE_DISABLE_PATH_WARNING=1
	AddLocalBinPath         bool // Add ~/.local/bin to PATH
	AddGoBinPath            bool // Add ~/go/bin to PATH (Go install location)
	DryRun                  bool // Don't actually modify files
	BackupFirst             bool // Create backup before modification
	PreferLoginShell        bool // Prefer login shell config files (.zshenv, .profile)
}

// ConfigResult holds the result of shell configuration.
type ConfigResult struct {
	Success    bool     // Whether configuration succeeded
	Message    string   // Human-readable result message
	ConfigFile string   // Which config file was modified
	LinesAdded []string // Lines that were added
	BackupPath string   // Path to backup file if created
	Skipped    bool     // Whether configuration was skipped (already configured)
}

// Recommendation holds the recommended configuration.
type Recommendation struct {
	Shell       ShellType // Detected shell
	ConfigFile  string    // Recommended config file
	Changes     []string  // Changes to add
	Explanation string    // Why this recommendation
}
