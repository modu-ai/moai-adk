package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var ccCmd = &cobra.Command{
	Use:   "cc",
	Short: "Switch to Claude backend",
	Long: `Switch the active LLM backend to Claude by removing GLM env variables from .claude/settings.local.json.

This command removes the GLM-specific environment variables that were injected
by 'moai glm', restoring Claude Code to use the default Claude API.

Use 'moai glm' to switch to GLM backend.`,
	Args: cobra.NoArgs,
	RunE: runCC,
}

func init() {
	rootCmd.AddCommand(ccCmd)
}

// runCC switches the LLM backend to Claude by removing GLM env from settings.local.json.
func runCC(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	// Get project root
	root, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
	}

	// Remove env from settings.local.json
	settingsPath := filepath.Join(root, ".claude", "settings.local.json")
	if err := removeGLMEnv(settingsPath); err != nil {
		return fmt.Errorf("remove GLM env: %w", err)
	}

	// Remove project-level .env.glm
	if err := removeProjectEnvGLM(root); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to remove project .env.glm: %v\n", err)
	}

	_, _ = fmt.Fprintln(out, renderSuccessCard(
		"Switched to Claude backend",
		"GLM configuration removed from:",
		"  - .claude/settings.local.json",
		"  - .moai/.env.glm",
		"",
		"Run 'moai glm' to switch to GLM.",
	))
	return nil
}

// removeProjectEnvGLM removes the project-level .moai/.env.glm file.
func removeProjectEnvGLM(projectRoot string) error {
	envPath := filepath.Join(projectRoot, ".moai", ".env.glm")

	if err := os.Remove(envPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already removed, no-op
		}
		return fmt.Errorf("remove project .env.glm: %w", err)
	}

	return nil
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
