package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// launchClaudeFunc is the function used by launchClaude. Override in tests.
var launchClaudeFunc = launchClaudeDefault

// launchClaude delegates to launchClaudeFunc for testability.
func launchClaude(profileName string, extraArgs []string) error {
	return launchClaudeFunc(profileName, extraArgs)
}

// launchClaudeDefault finds the claude binary, reads DO_CLAUDE_* settings from
// settings.local.json, and replaces the current process with claude via
// syscall.Exec. profileName may be empty for the default profile. extraArgs
// are additional CLI args to pass through to claude.
func launchClaudeDefault(profileName string, extraArgs []string) error {
	// 1. Profile setup
	if profileName != "" && profileName != "default" {
		if err := profile.EnsureDir(profileName); err != nil {
			return fmt.Errorf("set profile: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Profile: %s\n", profileName)
	}

	// 2. Find claude binary
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude not found in PATH. Install Claude Code first")
	}

	// 3. Read settings.local.json for DO_CLAUDE_* flags
	settings := readSettingsLocalForLaunch()

	bypass := settings["DO_CLAUDE_BYPASS"] == "true"
	chrome := settings["DO_CLAUDE_CHROME"] == "true"
	cont := settings["DO_CLAUDE_CONTINUE"] == "true"
	model := settings["DO_CLAUDE_MODEL"]

	// 4. Parse extra args (overrides)
	var passThrough []string
	for i := 0; i < len(extraArgs); i++ {
		arg := extraArgs[i]
		switch arg {
		case "--chrome":
			chrome = true
		case "--no-chrome":
			chrome = false
		case "-b", "--bypass":
			bypass = true
		case "-c", "--continue":
			cont = true
		case "--model", "-m":
			if i+1 < len(extraArgs) {
				model = extraArgs[i+1]
				i++
			}
		default:
			passThrough = append(passThrough, arg)
		}
	}

	// 5. Build args
	buildArgs := func(withContinue bool) []string {
		a := []string{"claude"}
		if bypass {
			a = append(a, "--dangerously-skip-permissions")
		}
		if !chrome {
			a = append(a, "--no-chrome")
		}
		if withContinue {
			a = append(a, "--continue")
		}
		if model != "" {
			a = append(a, "--model", model)
		}
		a = append(a, passThrough...)
		return a
	}

	// 6. Execute with --continue fallback
	if cont {
		tryCmd := exec.Command(claudeBin, buildArgs(true)[1:]...)
		tryCmd.Stdin = os.Stdin
		tryCmd.Stdout = os.Stdout
		tryCmd.Stderr = os.Stderr
		if err := tryCmd.Run(); err == nil {
			return nil
		}
		fmt.Fprintln(os.Stderr, "No previous session found, starting new session...")
	}

	return syscall.Exec(claudeBin, buildArgs(false), os.Environ())
}

// parseProfileFlag extracts -p/--profile from args and returns the profile name
// and the remaining args with the flag removed.
func parseProfileFlag(args []string) (string, []string) {
	var profileName string
	filtered := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		if (args[i] == "--profile" || args[i] == "-p") && i+1 < len(args) {
			profileName = args[i+1]
			i++
			continue
		}
		// Handle --profile=value form
		if strings.HasPrefix(args[i], "--profile=") {
			profileName = strings.TrimPrefix(args[i], "--profile=")
			continue
		}
		if strings.HasPrefix(args[i], "-p=") {
			profileName = strings.TrimPrefix(args[i], "-p=")
			continue
		}
		filtered = append(filtered, args[i])
	}

	return profileName, filtered
}

// readSettingsLocalForLaunch reads the env map from .claude/settings.local.json
// in the current directory (or project root). Returns an empty map on error.
func readSettingsLocalForLaunch() map[string]string {
	result := make(map[string]string)

	// Try project root first, fall back to current directory
	settingsPath := filepath.Join(".claude", "settings.local.json")
	root, err := findProjectRoot()
	if err == nil {
		settingsPath = filepath.Join(root, ".claude", "settings.local.json")
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return result
	}

	var settings SettingsLocal
	if err := json.Unmarshal(data, &settings); err != nil {
		return result
	}

	for k, v := range settings.Env {
		result[k] = v
	}
	return result
}
