package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var glmHybridFlag bool

var glmCmd = &cobra.Command{
	Use:   "glm [api-key]",
	Short: "Switch to GLM backend",
	Long: `Switch the active LLM backend to GLM by injecting env variables into .claude/settings.local.json.

If an API key is provided as an argument, it will be saved to ~/.moai/.env.glm
for future use. The key is stored securely with owner-only permissions (600).

This command reads GLM configuration from .moai/config/sections/llm.yaml and
injects the appropriate environment variables into Claude Code's settings.

Modes:
  moai glm          All agents use GLM models (requires tmux or iTerm2)
  moai glm --hybrid Lead uses Claude (opus/sonnet), agents use GLM (requires tmux)

Examples:
  moai glm                    # Use saved or environment API key, all agents use GLM
  moai glm sk-xxx-your-key    # Save API key and switch to GLM
  moai glm --hybrid           # Lead: Claude, Agents: GLM (cost optimization)

Use 'moai cc' to switch back to Claude backend and disable team mode.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGLM,
}

func init() {
	rootCmd.AddCommand(glmCmd)
	glmCmd.Flags().BoolVar(&glmHybridFlag, "hybrid", false, "Enable hybrid mode: lead uses Claude, agents use GLM. Requires tmux or iTerm2.")
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
func runGLM(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	// If API key provided as argument, save it first
	if len(args) > 0 {
		apiKey := args[0]
		if err := saveGLMKey(apiKey); err != nil {
			return fmt.Errorf("save GLM API key: %w", err)
		}
		_, _ = fmt.Fprintln(out, renderSuccessCard("GLM API key saved to ~/.moai/.env.glm"))
	}

	// Handle --hybrid flag
	// moai glm = all agents use GLM
	// moai glm --hybrid = lead uses Claude, agents use GLM
	return enableTeamMode(cmd, glmHybridFlag)
}

// enableTeamMode enables GLM Team mode with tmux environment configuration injection.
// isHybrid: false = all agents use GLM, true = lead uses Claude, agents use GLM
// Both modes require tmux or iTerm2 for split-pane display.
func enableTeamMode(cmd *cobra.Command, isHybrid bool) error {
	out := cmd.OutOrStdout()

	// Check if tmux is available (required for team modes)
	// Note: isTmuxAvailable returns true in test environment to allow testing
	if !isTmuxAvailable() {
		return fmt.Errorf("tmux is required for team mode but not found in PATH. Install tmux or use iTerm2")
	}

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
		return fmt.Errorf("GLM API key not found. Run 'moai glm <api-key>' to save your key, or set %s environment variable", glmConfig.EnvVar)
	}

	settingsPath := filepath.Join(root, defs.ClaudeDir, defs.SettingsLocalJSON)

	// Check if we're in a tmux session for proper env propagation
	inTmux := isInTmuxSession()

	// Inject GLM environment variables into tmux session
	// This is critical for Agent Teams - new panes (teammates) will inherit these
	if inTmux {
		if err := injectTmuxSessionEnv(glmConfig, apiKey); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to inject tmux session env: %v\n", err)
			// Continue anyway - settings.local.json will still work for lead
		}
	}

	if isHybrid {
		// Hybrid mode: Lead uses Claude, agents use GLM
		// Only inject GLM env for teammates (via tmux), not for lead
		if err := persistTeamMode(root, "hybrid"); err != nil {
			return fmt.Errorf("persist team mode: %w", err)
		}

		// Inject only tmux display mode for lead session (lead uses Claude)
		// GLM env vars will be injected by status_line.sh for teammates only
		if err := injectTmuxDisplayOnly(settingsPath); err != nil {
			return fmt.Errorf("inject tmux display: %w", err)
		}

		// Create project-level .env.glm for teammate sourcing (not for lead)
		if err := createProjectEnvGLM(glmConfig, root); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to create project .env.glm: %v\n", err)
		}

		// Build status message
		tmuxStatus := "tmux session: active (env vars injected)"
		if !inTmux {
			tmuxStatus = "tmux session: NOT DETECTED (start claude inside tmux for teammates)"
		}

		_, _ = fmt.Fprintln(out, renderSuccessCard(
			"GLM Hybrid Team mode enabled",
			"",
			"Architecture: Teamlead (Claude opus/sonnet) + Agents (GLM-5 to glm-4.7)",
			"Display mode: tmux (split panes)",
			tmuxStatus,
			"Config saved to: .moai/config/sections/llm.yaml",
			"",
			"Agent model mapping:",
			"  - teamlead: opus/sonnet (your selected Claude model)",
			"  - team-researcher: glm-4.7-flash (fastest exploration)",
			"  - team-analyst: glm-5 (requirements analysis)",
			"  - team-architect: glm-5 (technical design)",
			"  - team-backend-dev: glm-4.7 (implementation)",
			"  - team-frontend-dev: glm-4.7 (implementation)",
			"  - team-tester: glm-4.7 (test creation)",
			"  - team-quality: glm-4.7-flash (quality validation)",
			"",
			"Cost savings: 60-70% on implementation tasks",
			"",
			"Next steps:",
			"  1. Ensure you're in a tmux session (tmux new -s moai)",
			"  2. Start Claude Code: claude",
			"  3. Run workflow: /moai --team \"your task\"",
			"",
			"Run 'moai cc' to disable team mode.",
		))
	} else {
		// Regular team mode: All agents use GLM
		if err := persistTeamMode(root, "glm"); err != nil {
			return fmt.Errorf("persist team mode: %w", err)
		}

		// Inject GLM environment variables for all (lead and agents)
		if err := injectGLMEnvForTeam(settingsPath, glmConfig, apiKey); err != nil {
			return fmt.Errorf("inject GLM env for team: %w", err)
		}

		// Create project-level .env.glm for status_line.sh sourcing
		if err := createProjectEnvGLM(glmConfig, root); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to create project .env.glm: %v\n", err)
		}

		// Build status message
		tmuxStatus := "tmux session: active (env vars injected)"
		if !inTmux {
			tmuxStatus = "tmux session: NOT DETECTED (start claude inside tmux for teammates)"
		}

		_, _ = fmt.Fprintln(out, renderSuccessCard(
			"GLM Team mode enabled",
			"",
			"Architecture: All agents use GLM models",
			"Display mode: tmux (split panes)",
			tmuxStatus,
			"Config saved to: .moai/config/sections/llm.yaml",
			"",
			"Agent model mapping:",
			"  - teamlead: glm-5",
			"  - team-researcher: glm-4.7-flash (fastest exploration)",
			"  - team-analyst: glm-5 (requirements analysis)",
			"  - team-architect: glm-5 (technical design)",
			"  - team-backend-dev: glm-4.7 (implementation)",
			"  - team-frontend-dev: glm-4.7 (implementation)",
			"  - team-tester: glm-4.7 (test creation)",
			"  - team-quality: glm-4.7-flash (quality validation)",
			"",
			"Next steps:",
			"  1. Ensure you're in a tmux session (tmux new -s moai)",
			"  2. Start Claude Code: claude",
			"  3. Run workflow: /moai --team \"your task\"",
			"",
			"Run 'moai cc' to disable team mode.",
		))
	}

	return nil
}

// isTmuxAvailable checks if tmux is available in PATH.
// Required for Agent Teams split-pane display mode.
// Returns true in test environment to allow testing without tmux.
func isTmuxAvailable() bool {
	// Skip tmux check in test environment
	if isTestEnvironment() {
		return true
	}
	_, err := exec.LookPath("tmux")
	return err == nil
}

// isInTmuxSession checks if we're running inside a tmux session.
// tmux set-environment only works when called from within a tmux session.
func isInTmuxSession() bool {
	return os.Getenv("TMUX") != ""
}

// injectTmuxSessionEnv sets environment variables at the tmux session level
// so that new panes (teammates) inherit them automatically.
// This is required because settings.local.json env only applies to the current session,
// not to new tmux panes which start fresh shells.
func injectTmuxSessionEnv(glmConfig *GLMConfigFromYAML, apiKey string) error {
	// Skip in test environment
	if isTestEnvironment() {
		return nil
	}

	// Only works inside a tmux session
	if !isInTmuxSession() {
		return nil
	}

	envVars := map[string]string{
		"ANTHROPIC_AUTH_TOKEN":           apiKey,
		"ANTHROPIC_BASE_URL":             glmConfig.BaseURL,
		"ANTHROPIC_DEFAULT_OPUS_MODEL":   glmConfig.Models.High,
		"ANTHROPIC_DEFAULT_SONNET_MODEL": glmConfig.Models.Medium,
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":  glmConfig.Models.Low,
	}

	for name, value := range envVars {
		// tmux set-environment sets a session-level environment variable
		// New panes will inherit these variables
		cmd := exec.Command("tmux", "set-environment", name, value)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("tmux set-environment %s: %w", name, err)
		}
	}

	return nil
}

// clearTmuxSessionEnv removes GLM environment variables from tmux session.
// Called when switching back to Claude mode (moai cc).
func clearTmuxSessionEnv() error {
	// Skip in test environment
	if isTestEnvironment() {
		return nil
	}

	// Only works inside a tmux session
	if !isInTmuxSession() {
		return nil
	}

	envVars := []string{
		"ANTHROPIC_AUTH_TOKEN",
		"ANTHROPIC_BASE_URL",
		"ANTHROPIC_DEFAULT_OPUS_MODEL",
		"ANTHROPIC_DEFAULT_SONNET_MODEL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	}

	for _, name := range envVars {
		cmd := exec.Command("tmux", "set-environment", "-u", name)
		if err := cmd.Run(); err != nil {
			// Log warning but don't fail - variable might not exist
			fmt.Fprintf(os.Stderr, "Warning: failed to clear tmux env %s: %v\n", name, err)
		}
	}

	return nil
}

// injectTmuxDisplayOnly injects only the tmux display mode setting.
// Used for hybrid mode where lead uses Claude but agents use GLM.
func injectTmuxDisplayOnly(settingsPath string) error {
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

	// Set tmux display mode for Agent Teams (NO GLM env vars - lead uses Claude)
	settings.Env["CLAUDE_CODE_TEAMMATE_DISPLAY"] = "tmux"

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

// persistTeamMode saves the team_mode value to .moai/config/sections/llm.yaml.
// Only loads the LLM section to avoid validation errors from missing user.name.
func persistTeamMode(projectRoot, mode string) error {
	sectionsDir := filepath.Join(filepath.Clean(projectRoot), defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Load existing LLM config (or use defaults)
	llmCfg, err := loadLLMSectionOnly(sectionsDir)
	if err != nil {
		return fmt.Errorf("load LLM section: %w", err)
	}

	// Update team_mode
	llmCfg.TeamMode = mode

	// Save only the llm.yaml section atomically
	return saveLLMSection(sectionsDir, llmCfg)
}

// loadLLMSectionOnly loads only the LLM section from llm.yaml.
// Returns default LLM config if file doesn't exist.
func loadLLMSectionOnly(sectionsDir string) (config.LLMConfig, error) {
	llmPath := filepath.Join(sectionsDir, "llm.yaml")

	// Check if file exists
	if _, err := os.Stat(llmPath); os.IsNotExist(err) {
		// Return default config
		return config.NewDefaultLLMConfig(), nil
	}

	// Read existing file
	data, err := os.ReadFile(llmPath)
	if err != nil {
		return config.LLMConfig{}, fmt.Errorf("read llm.yaml: %w", err)
	}

	// Parse YAML (llm.yaml wrapper format)
	wrapper := struct {
		LLM config.LLMConfig `yaml:"llm"`
	}{}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return config.LLMConfig{}, fmt.Errorf("parse llm.yaml: %w", err)
	}

	return wrapper.LLM, nil
}

// disableTeamMode resets team_mode to empty in llm.yaml.
func disableTeamMode(projectRoot string) error {
	return persistTeamMode(projectRoot, "")
}

// injectGLMEnvForTeam injects GLM environment variables AND tmux display mode
// to settings.local.json for GLM Team mode.
// This enables teammates to use GLM models instead of Claude models.
func injectGLMEnvForTeam(settingsPath string, glmConfig *GLMConfigFromYAML, apiKey string) error {
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

	// Inject GLM environment variables for teammates
	settings.Env["ANTHROPIC_AUTH_TOKEN"] = apiKey
	settings.Env["ANTHROPIC_BASE_URL"] = glmConfig.BaseURL
	// Use tier-based model names (Claude Code uses OPUS/SONNET/HAIKU as env var names)
	settings.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"] = glmConfig.Models.High
	settings.Env["ANTHROPIC_DEFAULT_SONNET_MODEL"] = glmConfig.Models.Medium
	settings.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = glmConfig.Models.Low

	// Set tmux display mode for Agent Teams
	settings.Env["CLAUDE_CODE_TEAMMATE_DISPLAY"] = "tmux"

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

// saveLLMSection saves only the LLM section to llm.yaml.
func saveLLMSection(sectionsDir string, llm config.LLMConfig) error {
	wrapper := struct {
		LLM config.LLMConfig `yaml:"llm"`
	}{LLM: llm}

	data, err := yaml.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("marshal llm config: %w", err)
	}

	path := filepath.Join(sectionsDir, "llm.yaml")

	// Atomic write: temp file + rename
	tmp, err := os.CreateTemp(sectionsDir, ".llm-config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	return os.Rename(tmpName, path)
}


// GLMConfigFromYAML represents the GLM settings from llm.yaml.
type GLMConfigFromYAML struct {
	BaseURL string
	Models  struct {
		High   string
		Medium string
		Low    string
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
					High   string
					Medium string
					Low    string
				}{
					High:   cfg.LLM.GLM.Models.High,
					Medium: cfg.LLM.GLM.Models.Medium,
					Low:    cfg.LLM.GLM.Models.Low,
				},
				EnvVar: cfg.LLM.GLMEnvVar,
			}, nil
		}
	}

	// Fallback to default values
	defaults := config.NewDefaultLLMConfig()
	return &GLMConfigFromYAML{
		BaseURL: defaults.GLM.BaseURL,
		Models: struct {
			High   string
			Medium string
			Low    string
		}{
			High:   defaults.GLM.Models.High,
			Medium: defaults.GLM.Models.Medium,
			Low:    defaults.GLM.Models.Low,
		},
		EnvVar: defaults.GLMEnvVar,
	}, nil
}

// getGLMEnvPath returns the path to ~/.moai/.env.glm.
func getGLMEnvPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".moai", ".env.glm")
}

// saveGLMKey saves the GLM API key to ~/.moai/.env.glm.
func saveGLMKey(key string) error {
	envPath := getGLMEnvPath()
	if envPath == "" {
		return fmt.Errorf("cannot determine home directory")
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Escape special characters for dotenv format
	escapedKey := escapeDotenvValue(key)

	// Write in dotenv format
	content := fmt.Sprintf("# GLM API Key for MoAI-ADK\n# Generated by moai glm\nGLM_API_KEY=\"%s\"\n", escapedKey)
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// loadGLMKey loads the GLM API key from ~/.moai/.env.glm.
func loadGLMKey() string {
	envPath := getGLMEnvPath()
	if envPath == "" {
		return ""
	}

	file, err := os.Open(envPath)
	if err != nil {
		return ""
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if after, ok := strings.CutPrefix(line, "GLM_API_KEY="); ok {
			value := after
			// Remove quotes if present
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = unescapeDotenvValue(value[1 : len(value)-1])
			} else if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
				value = value[1 : len(value)-1]
			}
			return value
		}
	}
	return ""
}

// escapeDotenvValue escapes special characters for dotenv double-quoted value.
func escapeDotenvValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "$", "\\$")
	return value
}

// unescapeDotenvValue unescapes dotenv double-quoted value.
func unescapeDotenvValue(value string) string {
	value = strings.ReplaceAll(value, "\\$", "$")
	value = strings.ReplaceAll(value, "\\\"", "\"")
	value = strings.ReplaceAll(value, "\\\\", "\\")
	return value
}

// getGLMAPIKey returns the GLM API key from multiple sources.
// Priority: 1. ~/.moai/.env.glm, 2. Environment variable GLM_API_KEY
func getGLMAPIKey(envVar string) string {
	// Check saved key first
	if key := loadGLMKey(); key != "" {
		return key
	}
	// Fall back to environment variable
	return os.Getenv(envVar)
}

// buildGLMEnvVars constructs the environment variable map for GLM mode.
func buildGLMEnvVars(glmConfig *GLMConfigFromYAML, apiKey string) map[string]string {
	return map[string]string{
		"ANTHROPIC_AUTH_TOKEN":           apiKey,
		"ANTHROPIC_BASE_URL":             glmConfig.BaseURL,
		"ANTHROPIC_DEFAULT_OPUS_MODEL":   glmConfig.Models.High,
		"ANTHROPIC_DEFAULT_SONNET_MODEL": glmConfig.Models.Medium,
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":  glmConfig.Models.Low,
	}
}

// injectGLMEnv adds GLM environment variables to settings.local.json.
func injectGLMEnv(settingsPath string, glmConfig *GLMConfigFromYAML) error {
	// Get API key from saved file or environment
	apiKey := getGLMAPIKey(glmConfig.EnvVar)
	if apiKey == "" {
		return fmt.Errorf("GLM API key not found. Run 'moai glm <api-key>' to save your key, or set %s environment variable", glmConfig.EnvVar)
	}

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

	// Inject GLM environment variables with actual API key value
	settings.Env["ANTHROPIC_AUTH_TOKEN"] = apiKey
	settings.Env["ANTHROPIC_BASE_URL"] = glmConfig.BaseURL
	// Use tier-based model names (Claude Code uses OPUS/SONNET/HAIKU as env var names)
	settings.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"] = glmConfig.Models.High
	settings.Env["ANTHROPIC_DEFAULT_SONNET_MODEL"] = glmConfig.Models.Medium
	settings.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = glmConfig.Models.Low

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

// createProjectEnvGLM creates a project-level .moai/.env.glm with all ANTHROPIC_*
// variables as shell export statements. This file is sourced by status_line.sh
// so that Agent Teams teammates inherit GLM configuration.
func createProjectEnvGLM(glmConfig *GLMConfigFromYAML, projectRoot string) error {
	apiKey := getGLMAPIKey(glmConfig.EnvVar)
	if apiKey == "" {
		return fmt.Errorf("GLM API key not found")
	}

	envPath := filepath.Join(projectRoot, ".moai", ".env.glm")

	content := fmt.Sprintf(`# GLM environment variables for Agent Teams
# Generated by moai glm
# This file is sourced by .moai/status_line.sh
export ANTHROPIC_AUTH_TOKEN="%s"
export ANTHROPIC_BASE_URL="%s"
export ANTHROPIC_DEFAULT_OPUS_MODEL="%s"
export ANTHROPIC_DEFAULT_SONNET_MODEL="%s"
export ANTHROPIC_DEFAULT_HAIKU_MODEL="%s"
`,
		escapeDotenvValue(apiKey),
		glmConfig.BaseURL,
		glmConfig.Models.High,
		glmConfig.Models.Medium,
		glmConfig.Models.Low,
	)

	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		return fmt.Errorf("create .moai directory: %w", err)
	}

	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write project .env.glm: %w", err)
	}

	return nil
}

// isTestEnvironment detects if we're running in a test environment.
// This prevents tests from modifying the actual project's settings.local.json.
func isTestEnvironment() bool {
	// Check if tests have explicitly enabled test mode
	// This allows tests to opt-in to test mode without affecting all tests
	if flag := os.Getenv("MOAI_TEST_MODE"); flag == "1" {
		return true
	}
	// Check if running under go test by examining os.Args
	for _, arg := range os.Args {
		if strings.HasSuffix(arg, ".test") || strings.Contains(arg, "go.test") {
			return true
		}
	}
	return false
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
