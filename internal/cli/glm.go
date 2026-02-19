package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var glmTeamFlag bool

var glmCmd = &cobra.Command{
	Use:   "glm [api-key]",
	Short: "Switch to GLM backend",
	Long: `Switch the active LLM backend to GLM by injecting env variables into .claude/settings.local.json.

If an API key is provided as an argument, it will be saved to ~/.moai/.env.glm
for future use. The key is stored securely with owner-only permissions (600).

This command reads GLM configuration from .moai/config/sections/llm.yaml and
injects the appropriate environment variables into Claude Code's settings.

Use --team flag for hybrid mode: leader stays on Anthropic Opus while
Agent Teams teammates use GLM. Requires tmux.

Examples:
  moai glm                    # Use saved or environment API key
  moai glm sk-xxx-your-key    # Save API key and switch to GLM
  moai glm --team             # Hybrid: leader=Opus, teammates=GLM (tmux)

Use 'moai cc' to switch back to Claude backend.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGLM,
}

func init() {
	rootCmd.AddCommand(glmCmd)
	glmCmd.Flags().BoolVar(&glmTeamFlag, "team", false, "Hybrid mode: teammates use GLM, leader stays on Opus (requires tmux)")
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

	// Branch based on --team flag
	if glmTeamFlag {
		return runGLMTeam(cmd)
	}

	// Get project root
	root, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
	}

	// Prevent modifying real project settings during tests
	if isTestEnvironment() {
		_, _ = fmt.Fprintln(out, "⚠️  Test environment detected - skipping settings.local.json modification")
		return nil
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

	// Create project-level .env.glm for status_line.sh sourcing (Agent Teams tmux mode)
	if err := createProjectEnvGLM(glmConfig, root); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to create project .env.glm: %v\n", err)
	}

	_, _ = fmt.Fprintln(out, renderSuccessCard(
		"Switched to GLM backend",
		"Environment variables configured:",
		"  - .claude/settings.local.json (project)",
		"  - .moai/.env.glm (for Agent Teams tmux mode)",
		"",
		"Run 'moai cc' to switch back to Claude.",
	))
	return nil
}

// runGLMTeam starts a tmux session for hybrid model mode.
// Leader pane uses Anthropic Opus, teammate panes inherit GLM env vars.
func runGLMTeam(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()

	// Verify tmux is available
	if err := checkTmuxAvailable(); err != nil {
		return err
	}

	// Get project root
	root, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
	}

	// Load GLM config
	glmConfig, err := loadGLMConfig(root)
	if err != nil {
		return fmt.Errorf("load GLM config: %w", err)
	}

	// Get API key
	apiKey := getGLMAPIKey(glmConfig.EnvVar)
	if apiKey == "" {
		return fmt.Errorf("GLM API key not found. Run 'moai glm <api-key>' to save your key, or set %s environment variable", glmConfig.EnvVar)
	}

	sessionName := "moai-hybrid-glm"

	// Check if session already exists
	if tmuxSessionExists(sessionName) {
		_, _ = fmt.Fprintln(out, renderSuccessCard(
			"Hybrid session already exists",
			"",
			"Attaching to existing session: "+sessionName,
			"Run 'tmux kill-session -t "+sessionName+"' to stop it.",
		))
		return execTmuxAttach(sessionName)
	}

	// Build env vars map
	envVars := buildGLMEnvVars(glmConfig, apiKey)

	// Create tmux session and inject env vars
	if err := createHybridTmuxSession(sessionName, envVars, root); err != nil {
		return fmt.Errorf("create hybrid tmux session: %w", err)
	}

	_, _ = fmt.Fprintln(out, renderSuccessCard(
		"Hybrid mode started",
		"",
		"Session: "+sessionName,
		"Leader: Anthropic Opus (direct API)",
		"Teammates: GLM ("+glmConfig.Models.Opus+", "+glmConfig.Models.Sonnet+", "+glmConfig.Models.Haiku+")",
		"",
		"Claude is starting in the leader pane...",
	))

	// Attach to the tmux session
	return execTmuxAttach(sessionName)
}

// checkTmuxAvailable verifies that tmux is installed and accessible.
func checkTmuxAvailable() error {
	_, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux is required for --team mode but was not found in PATH. Install tmux first")
	}
	return nil
}

// tmuxSessionExists checks if a tmux session with the given name exists.
func tmuxSessionExists(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	return cmd.Run() == nil
}

// buildGLMEnvVars constructs the environment variable map for GLM hybrid mode.
func buildGLMEnvVars(glmConfig *GLMConfigFromYAML, apiKey string) map[string]string {
	return map[string]string{
		"ANTHROPIC_AUTH_TOKEN":           apiKey,
		"ANTHROPIC_BASE_URL":             glmConfig.BaseURL,
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":  glmConfig.Models.Haiku,
		"ANTHROPIC_DEFAULT_SONNET_MODEL": glmConfig.Models.Sonnet,
		"ANTHROPIC_DEFAULT_OPUS_MODEL":   glmConfig.Models.Opus,
	}
}

// createHybridTmuxSession creates a tmux session with GLM env vars at session level,
// then starts claude in the leader pane with those env vars unset.
func createHybridTmuxSession(sessionName string, envVars map[string]string, projectRoot string) error {
	// Create detached tmux session starting in the project root
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", projectRoot)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("create tmux session: %s: %w", string(out), err)
	}

	// Set session-level environment variables
	for key, value := range envVars {
		cmd := exec.Command("tmux", "set-environment", "-t", sessionName, key, value)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("set tmux env %s: %s: %w", key, string(out), err)
		}
	}

	// Build the leader pane command:
	// 1. Unset ANTHROPIC_* env vars (so leader uses Anthropic API directly = Opus)
	// 2. Start claude
	leaderCmd := "unset ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL " +
		"ANTHROPIC_DEFAULT_HAIKU_MODEL ANTHROPIC_DEFAULT_SONNET_MODEL " +
		"ANTHROPIC_DEFAULT_OPUS_MODEL && claude"

	// Send the command to the leader pane
	cmd = exec.Command("tmux", "send-keys", "-t", sessionName, leaderCmd, "Enter")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("send leader command: %s: %w", string(out), err)
	}

	return nil
}

// execTmuxAttach attaches to the given tmux session, replacing the current process.
func execTmuxAttach(sessionName string) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("find tmux: %w", err)
	}
	return syscall.Exec(tmuxPath, []string{"tmux", "attach", "-t", sessionName}, os.Environ())
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
			Opus:   "glm-5",
		},
		EnvVar: "GLM_API_KEY",
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
		if strings.HasPrefix(line, "GLM_API_KEY=") {
			value := strings.TrimPrefix(line, "GLM_API_KEY=")
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

// createProjectEnvGLM creates a project-level .moai/.env.glm with all ANTHROPIC_*
// variables as shell export statements. This file is sourced by status_line.sh
// so that Agent Teams teammates in tmux mode inherit GLM configuration.
func createProjectEnvGLM(glmConfig *GLMConfigFromYAML, projectRoot string) error {
	apiKey := getGLMAPIKey(glmConfig.EnvVar)
	if apiKey == "" {
		return fmt.Errorf("GLM API key not found")
	}

	envPath := filepath.Join(projectRoot, ".moai", ".env.glm")

	content := fmt.Sprintf(`# GLM environment variables for Agent Teams
# Generated by moai glm
# This file is sourced by .moai/status_line.sh in tmux mode
export ANTHROPIC_AUTH_TOKEN="%s"
export ANTHROPIC_BASE_URL="%s"
export ANTHROPIC_DEFAULT_HAIKU_MODEL="%s"
export ANTHROPIC_DEFAULT_SONNET_MODEL="%s"
export ANTHROPIC_DEFAULT_OPUS_MODEL="%s"
`,
		escapeDotenvValue(apiKey),
		glmConfig.BaseURL,
		glmConfig.Models.Haiku,
		glmConfig.Models.Sonnet,
		glmConfig.Models.Opus,
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
