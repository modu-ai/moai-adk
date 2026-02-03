package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var glmCmd = &cobra.Command{
	Use:   "glm",
	Short: "Switch to GLM backend",
	Long: `Switch the active LLM backend to GLM by injecting env variables into .claude/settings.local.json.

This command reads GLM configuration from .moai/config/sections/llm.yaml and
injects the appropriate environment variables into Claude Code's settings.

Use 'moai cc' to switch back to Claude backend.`,
	Args: cobra.NoArgs,
	RunE: runGLM,
}

func init() {
	rootCmd.AddCommand(glmCmd)
}

// SettingsLocal represents .claude/settings.local.json structure.
type SettingsLocal struct {
	Meta                  map[string]any    `json:"_meta,omitempty"`
	EnabledMcpjsonServers []string          `json:"enabledMcpjsonServers,omitempty"`
	CompanyAnnouncements  []string          `json:"companyAnnouncements,omitempty"`
	Env                   map[string]string `json:"env,omitempty"`
	Permissions           map[string]any    `json:"permissions,omitempty"`
}

// runGLM switches the LLM backend to GLM by modifying settings.local.json.
func runGLM(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	// Get project root
	root, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
	}

	// Load GLM config from llm.yaml
	glmConfig, err := loadGLMConfig(root)
	if err != nil {
		return fmt.Errorf("load GLM config: %w", err)
	}

	// Inject env into settings.local.json
	settingsPath := filepath.Join(root, ".claude", "settings.local.json")
	if err := injectGLMEnv(settingsPath, glmConfig); err != nil {
		return fmt.Errorf("inject GLM env: %w", err)
	}

	fmt.Fprintln(out, "✅ Switched to GLM backend.")
	fmt.Fprintln(out, "   Environment variables injected into .claude/settings.local.json")
	fmt.Fprintln(out, "   Run 'moai cc' to switch back to Claude.")
	return nil
}

// GLMConfigFromYAML represents the GLM settings from llm.yaml.
type GLMConfigFromYAML struct {
	BaseURL string
	Models  struct {
		Haiku  string
		Sonnet string
		Opus   string
	}
	EnvVar string
}

// loadGLMConfig reads GLM configuration from llm.yaml.
func loadGLMConfig(root string) (*GLMConfigFromYAML, error) {
	// If config is available via deps, use it
	if deps != nil && deps.Config != nil {
		cfg := deps.Config.Get()
		if cfg != nil && cfg.LLM.GLM.BaseURL != "" {
			return &GLMConfigFromYAML{
				BaseURL: cfg.LLM.GLM.BaseURL,
				Models: struct {
					Haiku  string
					Sonnet string
					Opus   string
				}{
					Haiku:  cfg.LLM.GLM.Models.Haiku,
					Sonnet: cfg.LLM.GLM.Models.Sonnet,
					Opus:   cfg.LLM.GLM.Models.Opus,
				},
				EnvVar: cfg.LLM.GLMEnvVar,
			}, nil
		}
	}

	// Fallback to default values
	return &GLMConfigFromYAML{
		BaseURL: "https://api.z.ai/api/anthropic",
		Models: struct {
			Haiku  string
			Sonnet string
			Opus   string
		}{
			Haiku:  "glm-4.7-flashx",
			Sonnet: "glm-4.7",
			Opus:   "glm-4.7",
		},
		EnvVar: "GLM_API_KEY",
	}, nil
}

// injectGLMEnv adds GLM environment variables to settings.local.json.
func injectGLMEnv(settingsPath string, glmConfig *GLMConfigFromYAML) error {
	var settings SettingsLocal

	// Read existing settings if file exists
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parse settings.local.json: %w", err)
		}
	}

	// Initialize env map if nil
	if settings.Env == nil {
		settings.Env = make(map[string]string)
	}

	// Inject GLM environment variables
	settings.Env["ANTHROPIC_AUTH_TOKEN"] = fmt.Sprintf("${%s}", glmConfig.EnvVar)
	settings.Env["ANTHROPIC_BASE_URL"] = glmConfig.BaseURL
	settings.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = glmConfig.Models.Haiku
	settings.Env["ANTHROPIC_DEFAULT_SONNET_MODEL"] = glmConfig.Models.Sonnet
	settings.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"] = glmConfig.Models.Opus

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Write back
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		return fmt.Errorf("write settings.local.json: %w", err)
	}

	return nil
}

// findProjectRoot finds the project root by looking for .moai directory.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".moai")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not in a MoAI project (no .moai directory found)")
		}
		dir = parent
	}
}
