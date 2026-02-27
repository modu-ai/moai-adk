package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/spf13/cobra"
)

var ccCmd = &cobra.Command{
	Use:   "cc [-p profile] [-- claude-args...]",
	Short: "Login to Claude backend and launch Claude Code",
	Long: `Login to Claude by removing GLM env variables, then launch Claude Code.

This command:
  1. Removes GLM-specific environment variables from .claude/settings.local.json
  2. Resets team mode if it was enabled (glm or cg)
  3. Optionally sets a profile via -p flag (CLAUDE_CONFIG_DIR)
  4. Reads DO_CLAUDE_* settings and converts them to CLI flags
  5. Launches Claude Code via exec (replaces current process)

Flags:
  -p, --profile <name>   Use a named Claude profile (~/.claude-profiles/<name>/)
  -b, --bypass           Enable --dangerously-skip-permissions
  -c, --continue         Continue previous session
  -m, --model <model>    Override model selection
  --chrome / --no-chrome Toggle Chrome MCP

Examples:
  moai cc                     # Default profile, launch Claude
  moai cc -p work             # Use 'work' profile
  moai cc -p work -- --print  # Profile + pass-through args to Claude`,
	DisableFlagParsing: true,
	RunE:               runCC,
}

func init() {
	rootCmd.AddCommand(ccCmd)
}

// runCC switches the LLM backend to Claude, then launches Claude Code.
func runCC(cmd *cobra.Command, args []string) error {
	// Handle --help/-h manually since DisableFlagParsing: true
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return cmd.Help()
		}
		if arg == "--" {
			break
		}
	}

	// Parse -p/--profile from args
	profileName, filteredArgs := parseProfileFlag(args)

	// Get project root
	root, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
	}

	// Clear tmux session environment variables (for Agent Teams)
	if err := clearTmuxSessionEnv(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to clear tmux session env: %v\n", err)
	}

	// Remove env from settings.local.json
	settingsPath := filepath.Join(root, ".claude", "settings.local.json")
	if err := removeGLMEnv(settingsPath); err != nil {
		return fmt.Errorf("remove GLM env: %w", err)
	}

	// Handle team_mode: disable and cleanup worktrees
	teamModeMsg := resetTeamModeForCC(root)
	if teamModeMsg != "" {
		fmt.Fprintln(os.Stderr, teamModeMsg)
	}

	// Cleanup moai worktrees if any exist
	worktreeMsg := cleanupMoaiWorktrees(root)
	if worktreeMsg != "" {
		fmt.Fprintln(os.Stderr, worktreeMsg)
	}

	fmt.Fprintln(os.Stderr, "Launching Claude Code...")

	// Launch claude with profile and extra args
	return launchClaude(profileName, filteredArgs)
}

// resetTeamModeForCC disables team_mode when switching to CC.
// Returns a message string describing what was changed, or empty if unchanged.
func resetTeamModeForCC(projectRoot string) string {
	mgr := config.NewConfigManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		return ""
	}

	cfg := mgr.Get()
	if cfg == nil || cfg.LLM.TeamMode == "" {
		return ""
	}

	prev := cfg.LLM.TeamMode
	if err := disableTeamMode(projectRoot); err != nil {
		return fmt.Sprintf("Warning: failed to disable team mode: %v", err)
	}
	return fmt.Sprintf("Team mode disabled (was: %s)", prev)
}

// removeGLMEnv removes GLM environment variables from settings.local.json.
func removeGLMEnv(settingsPath string) error {
	// Read existing settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Nothing to remove
			return nil
		}
		return fmt.Errorf("read settings.local.json: %w", err)
	}

	var settings SettingsLocal
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("parse settings.local.json: %w", err)
	}

	// Remove GLM-specific env variables
	if settings.Env != nil {
		delete(settings.Env, "ANTHROPIC_AUTH_TOKEN")
		delete(settings.Env, "ANTHROPIC_BASE_URL")
		delete(settings.Env, "ANTHROPIC_DEFAULT_HAIKU_MODEL")
		delete(settings.Env, "ANTHROPIC_DEFAULT_SONNET_MODEL")
		delete(settings.Env, "ANTHROPIC_DEFAULT_OPUS_MODEL")

		// Remove env key entirely if empty
		if len(settings.Env) == 0 {
			settings.Env = nil
		}
	}

	// Write back
	data, err = json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		return fmt.Errorf("write settings.local.json: %w", err)
	}

	return nil
}

// cleanupMoaiWorktrees removes moai-related git worktrees.
// These are worktrees created by /moai --team with names like worker-SPEC-XXX.
func cleanupMoaiWorktrees(projectRoot string) string {
	// Skip git call if no worktrees directory exists
	worktreeBase := filepath.Join(projectRoot, ".claude", "worktrees")
	if _, err := os.Stat(worktreeBase); os.IsNotExist(err) {
		return ""
	}

	// Check if we're in a git repository
	gitDir := filepath.Join(projectRoot, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return "" // Not a git repo, nothing to clean up
	}

	// List worktrees and find moai-related ones
	output, err := runGitCommand(projectRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return "" // Silently ignore errors
	}

	// Parse worktree list to find moai worker worktrees
	var cleanedWorktrees []string

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			worktreePath := strings.TrimPrefix(line, "worktree ")
			if strings.HasPrefix(worktreePath, worktreeBase) {
				workerName := filepath.Base(worktreePath)
				if strings.HasPrefix(workerName, "worker-") {
					if err := removeWorktree(projectRoot, workerName); err == nil {
						cleanedWorktrees = append(cleanedWorktrees, workerName)
					}
				}
			}
		}
	}

	if len(cleanedWorktrees) > 0 {
		return fmt.Sprintf("Cleaned up %d worktree(s): %s", len(cleanedWorktrees), strings.Join(cleanedWorktrees, ", "))
	}
	return ""
}

// removeWorktree removes a single git worktree.
func removeWorktree(projectRoot, worktreeName string) error {
	_, err := runGitCommand(projectRoot, "worktree", "remove", "--force", worktreeName)
	return err
}

// runGitCommand executes a git command in the given directory.
func runGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
