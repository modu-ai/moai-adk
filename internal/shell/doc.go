// Package shell provides shell detection, environment configuration,
// and PATH diagnostics for MoAI-ADK.
//
// It supports zsh, bash, and fish shells on macOS, Linux, and WSL.
// The package automatically selects the appropriate configuration file
// based on shell type and environment:
//
//   - zsh: ~/.zshenv (loaded for all shells, including non-interactive)
//   - bash: ~/.profile or ~/.bash_profile (for login shell support)
//   - fish: ~/.config/fish/config.fish
//
// This ensures that environment variables like CLAUDE_DISABLE_PATH_WARNING
// are available in IDE contexts (VS Code, Cursor) where non-interactive
// shells are used.
package shell
