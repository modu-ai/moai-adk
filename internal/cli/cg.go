package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/spf13/cobra"
)

var cgCmd = &cobra.Command{
	Use:   "cg [-p profile]",
	Short: "Login with Claude + GLM hybrid mode and launch Claude Code",
	Long: `Login with hybrid mode where the lead uses Claude and teammates use GLM, then launch Claude Code.

CG stands for "Claude + GLM" - a cost-optimized team configuration:
  - Lead (current tmux pane): Uses Claude models (opus/sonnet)
  - Teammates (new tmux panes): Use GLM models via Z.AI proxy

This command:
  1. Validates tmux session (required for pane isolation)
  2. Removes GLM env from settings.local.json (lead = Claude)
  3. Injects GLM env into tmux session (teammates = GLM)
  4. Optionally sets a profile via -p flag (CLAUDE_CONFIG_DIR)
  5. Sets CLAUDE_CODE_TEAMMATE_DISPLAY=tmux
  6. Saves team_mode: cg
  7. Launches Claude Code via exec (replaces current process)

Flags:
  -p, --profile <name>   Use a named Claude profile (~/.claude-profiles/<name>/)

Prerequisites:
  1. A GLM API key configured via 'moai glm setup <api-key>'
  2. Running inside a tmux session for pane-level environment isolation

Examples:
  moai glm setup sk-xxx    # First: save API key (one-time)
  moai cg                  # Then: launch hybrid mode
  moai cg -p work          # Use 'work' profile with hybrid mode

Use 'moai cc' to switch back to Claude-only mode.
Use 'moai glm' for all-GLM mode.`,
	DisableFlagParsing: true,
	RunE:               runCG,
}

func init() {
	rootCmd.AddCommand(cgCmd)
}

// runCG enables Claude + GLM hybrid mode and launches Claude Code.
func runCG(cmd *cobra.Command, args []string) error {
	// Parse -p/--profile from args
	profileName, filteredArgs := parseProfileFlag(args)

	root, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
	}

	// Load GLM config for environment variable injection
	glmConfig, err := loadGLMConfig(root)
	if err != nil {
		return fmt.Errorf("load GLM config: %w", err)
	}

	// Get API key
	apiKey := getGLMAPIKey(glmConfig.EnvVar)
	if apiKey == "" {
		return fmt.Errorf("GLM API key not found\n\n"+
			"Set up your API key first, then enable CG mode:\n"+
			"  1. moai glm setup <api-key>   (saves key to ~/.moai/.env.glm)\n"+
			"  2. moai cg                     (enable hybrid mode)\n\n"+
			"Or set the %s environment variable", glmConfig.EnvVar)
	}

	settingsPath := fmt.Sprintf("%s/.claude/settings.local.json", root)

	// Check if we're in a tmux session
	inTmux := isInTmuxSession()

	// CG mode requires tmux for pane-level environment isolation.
	if !inTmux && os.Getenv("MOAI_TEST_MODE") != "1" {
		return fmt.Errorf("CG mode requires a tmux session.\n\n" +
			"tmux is required because:\n" +
			"  - This pane (lead): uses Claude API\n" +
			"  - New panes (teammates): inherit GLM env for Z.AI API\n\n" +
			"Start a tmux session first:\n" +
			"  tmux new -s moai\n" +
			"  moai cg\n\n" +
			"Or use 'moai glm' for all-GLM mode (no tmux required)")
	}

	// Inject GLM environment variables into tmux session
	if inTmux {
		if err := injectTmuxSessionEnv(glmConfig, apiKey); err != nil {
			return fmt.Errorf("failed to inject GLM env into tmux session: %w\n"+
				"CG mode relies on tmux session env for teammate isolation.\n"+
				"Try restarting your tmux session", err)
		}

		// Also inject CLAUDE_CONFIG_DIR into tmux if profile is set
		if profileName != "" && profileName != "default" && !isTestEnvironment() {
			profileDir := profile.GetProfileDir(profileName)
			if profileDir != "" {
				tmuxCmd := exec.Command("tmux", "set-environment", "CLAUDE_CONFIG_DIR", profileDir)
				_ = tmuxCmd.Run()
			}
		}
	}

	// Persist team mode
	if err := persistTeamMode(root, "cg"); err != nil {
		return fmt.Errorf("persist team mode: %w", err)
	}

	// Clean up existing GLM env vars from settings.local.json (leader = Claude)
	if err := removeGLMEnv(settingsPath); err != nil {
		return fmt.Errorf("clean up GLM env for CG mode: %w", err)
	}

	// Ensure CLAUDE_CODE_TEAMMATE_DISPLAY=tmux
	if err := ensureSettingsLocalJSON(settingsPath); err != nil {
		return fmt.Errorf("ensure settings.local.json: %w", err)
	}

	fmt.Fprintln(os.Stderr, "CG mode: Lead (Claude) + Teammates (GLM)")
	fmt.Fprintln(os.Stderr, "Launching Claude Code...")

	// Launch claude with profile and extra args
	return launchClaude(profileName, filteredArgs)
}
