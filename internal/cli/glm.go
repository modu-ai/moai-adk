package cli

// @MX:NOTE: [AUTO] GLM command launches Claude Code with GLM backend via Z.AI proxy
// @MX:NOTE: [AUTO] Requires 'moai glm setup <key>' to save API key to ~/.moai/.env.glm
// @MX:NOTE: [AUTO] Main session uses GLM: 128K/200K/204K context windows per model tier
// @MX:NOTE: [AUTO] DISABLE_PROMPT_CACHING=1 disables prompt caching for Z.AI compatibility

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/statusline"
	"github.com/modu-ai/moai-adk/internal/tmux"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var glmCmd = &cobra.Command{
	Use:   "glm [-p profile] [-- claude-args...]",
	Short: "Launch Claude Code with GLM backend",
	Long: `Launch Claude Code with GLM backend.

All agents use GLM models via Z.AI proxy.

This command:
  1. Loads GLM credentials from ~/.moai/.env.glm
  2. Injects GLM environment variables (ANTHROPIC_AUTH_TOKEN, ANTHROPIC_BASE_URL, etc.)
  3. Optionally sets a profile via -p flag (CLAUDE_CONFIG_DIR)
  4. Launches Claude Code via exec (replaces current process)

Use 'moai glm setup <key>' to save your API key first.

Flags:
  -p, --profile <name>          Use a named Claude profile (~/.moai/claude-profiles/<name>/)
  --permission-mode <mode>      Set permission mode (default, acceptEdits, plan, bypassPermissions, dontAsk)
  -b, --bypass                  Shorthand for --permission-mode bypassPermissions

Note: Auto mode is not available with GLM (third-party provider).
Use 'moai cc --permission-mode auto' or 'moai cg --permission-mode auto' instead.

Note: Z.AI enforces low concurrency limits (paid tiers observe 1-3 in-flight
requests). Multi-agent workflows that exceed this limit can surface as opaque
errors (sometimes misreported by clients as "context window limit"). The GLM
models themselves have ample context (glm-5.1 ~204K, glm-4.7 ~202K). For more
stable parallel execution with MoAI Agent Teams, prefer 'moai cg' (hybrid mode).

Examples:
  moai glm setup sk-xxx    # Save API key (one-time)
  moai glm                 # Launch with GLM backend
  moai glm -p work         # Use 'work' profile with GLM

For hybrid mode (Claude lead + GLM teammates), use 'moai cg' instead.
Use 'moai cc' to switch back to Claude backend.`,
	GroupID:            "launch",
	DisableFlagParsing: true,
	RunE:               runGLM,
}

var glmSetupCmd = &cobra.Command{
	Use:   "setup [api-key]",
	Short: "Store a GLM API key",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runGLMSetup,
}

var glmStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current GLM credential status",
	RunE: func(cmd *cobra.Command, _ []string) error {
		key := loadGLMKey()
		if key == "" {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no GLM credentials configured")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Run 'moai glm setup <api-key>' to save your key")
			return nil
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "GLM API key: %s\n", maskAPIKey(key))
		return nil
	},
}

func init() {
	// Note: glm has DisableFlagParsing=true, so subcommand routing is manual.
	// We register setup and status as subcommands for discoverability (help output).
	glmCmd.AddCommand(glmSetupCmd, glmStatusCmd)
	rootCmd.AddCommand(glmCmd)
}

// SettingsLocal represents .claude/settings.local.json structure.
type SettingsLocal struct {
	Meta                  map[string]any    `json:"_meta,omitempty"`
	EnabledMcpjsonServers []string          `json:"enabledMcpjsonServers,omitempty"`
	CompanyAnnouncements  []string          `json:"companyAnnouncements,omitempty"`
	Env                   map[string]string `json:"env,omitempty"`
	Permissions           map[string]any    `json:"permissions,omitempty"`
	// TeammateMode controls how Claude Code displays Agent Teams teammates.
	// This native settings key takes precedence over the CLAUDE_CODE_TEAMMATE_DISPLAY
	// env var. CG/GLM modes set this to "tmux" to ensure teammates spawn in tmux
	// panes and inherit GLM session env vars. CC mode clears it so the project
	// default from settings.json ("auto") applies.
	TeammateMode string `json:"teammateMode,omitempty"`
}

// runGLM launches Claude Code with GLM backend, or routes to subcommands.
func runGLM(cmd *cobra.Command, args []string) error {
	// Handle --help/-h manually since DisableFlagParsing: true
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return cmd.Help()
		}
		if arg == "--" {
			break
		}
	}

	// Manual subcommand routing (DisableFlagParsing prevents automatic routing)
	if len(args) > 0 {
		switch args[0] {
		case "setup":
			return runGLMSetup(cmd, args[1:])
		case "status":
			return glmStatusCmd.RunE(cmd, nil)
		}
	}

	profileName, filteredArgs, err := parseProfileFlag(args)
	if err != nil {
		return err
	}

	// Auto mode is not available with third-party providers (GLM/Z.AI).
	// Validate before launch to give a clear error instead of a cryptic Claude Code rejection.
	if containsPermissionMode(filteredArgs, "auto") {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "auto mode requires Claude Sonnet 4.6 or Opus 4.6 running on Anthropic's API")
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "use 'moai cc --permission-mode auto' or 'moai cg --permission-mode auto' instead")
		return fmt.Errorf("auto mode is not available with GLM (third-party provider)")
	}

	// Warn about main-session GLM limitations before launch.
	// DISABLE_PROMPT_CACHING=1 forces full system prompt re-send per request (~30-40K tokens),
	// which hits GLM context limits faster than expected. Z.AI concurrency limits (1-3 in-flight
	// requests per paid tier) are sometimes misreported by Claude Code as "context window limit".
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "WARNING: moai glm uses GLM models for the MAIN SESSION. Known limitations:")
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "  - Main session context window: 128K (glm-4.5-air), 202K (glm-4.7), 204K (glm-5.1)")
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "  - DISABLE_PROMPT_CACHING=1 causes full system prompt re-send per request")
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "  - Z.AI concurrency is limited (1-3 in-flight requests per paid tier)")
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "If you want Claude as leader and GLM for teammates, use 'moai cg' instead.")

	return unifiedLaunch(profileName, "glm", filteredArgs)
}

// containsPermissionMode checks if args contain --permission-mode with the given value.
func containsPermissionMode(args []string, mode string) bool {
	for i, arg := range args {
		if arg == "--permission-mode" && i+1 < len(args) && args[i+1] == mode {
			return true
		}
		if arg == "--permission-mode="+mode {
			return true
		}
	}
	return false
}

// setGLMEnv sets GLM environment variables in the current process.
// @MX:WARN: [AUTO] Global environment variable mutation without rollback mechanism
// @MX:REASON: Process-level state mutation affects all subsequent goroutines; no cleanup on error
func setGLMEnv(glmConfig *GLMConfigFromYAML, apiKey string) {
	_ = os.Setenv("ANTHROPIC_AUTH_TOKEN", apiKey)                            //nolint:errcheck
	_ = os.Setenv("ANTHROPIC_BASE_URL", glmConfig.BaseURL)                   //nolint:errcheck
	_ = os.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", glmConfig.Models.High)     //nolint:errcheck
	_ = os.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", glmConfig.Models.Medium) //nolint:errcheck
	_ = os.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", glmConfig.Models.Low)     //nolint:errcheck
	// Z.AI proxy compatibility: strip Anthropic beta headers and prompt caching
	_ = os.Setenv("CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS", "1")    //nolint:errcheck
	_ = os.Setenv("DISABLE_PROMPT_CACHING", "1")                    //nolint:errcheck
	_ = os.Setenv("API_TIMEOUT_MS", "3000000")                      //nolint:errcheck
	_ = os.Setenv("CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC", "1") //nolint:errcheck
}

// runGLMSetup saves a GLM API key.
func runGLMSetup(cmd *cobra.Command, args []string) error {
	apiKey := ""
	if len(args) >= 1 {
		apiKey = strings.TrimSpace(args[0])
	} else {
		_, _ = fmt.Fprint(cmd.OutOrStdout(), "GLM API key: ")
		scanner := bufio.NewScanner(cmd.InOrStdin())
		if !scanner.Scan() {
			return nil
		}
		apiKey = strings.TrimSpace(scanner.Text())
	}

	if apiKey == "" {
		return fmt.Errorf("empty API key")
	}

	if err := saveGLMKey(apiKey); err != nil {
		return fmt.Errorf("save GLM API key: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "GLM API key stored (%s)\n", maskAPIKey(apiKey))
	return nil
}

// maskAPIKey masks an API key for display, showing only prefix and suffix.
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// enableTeamMode enables GLM Team mode with settings.local.json configuration.
// isHybrid: false = all agents use GLM, true = lead uses Claude, agents use GLM
// Note: teammateMode is forced to "tmux" to ensure GLM env var inheritance (#468)
func enableTeamMode(cmd *cobra.Command, isHybrid bool) error {
	out := cmd.OutOrStdout()

	root, err := findProjectRootFn()
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
		if isHybrid {
			return fmt.Errorf("GLM API key not found\n\n"+
				"Set up your API key first, then enable CG mode:\n"+
				"  1. moai glm setup <api-key>   (saves key to ~/.moai/.env.glm)\n"+
				"  2. moai cg                     (enable hybrid mode)\n\n"+
				"Or set the %s environment variable", glmConfig.EnvVar)
		}
		return fmt.Errorf("GLM API key not found. Run 'moai glm setup <api-key>' to save your key, or set %s environment variable", glmConfig.EnvVar)
	}

	settingsPath := filepath.Join(root, defs.ClaudeDir, defs.SettingsLocalJSON)

	// Check whether we are inside a tmux session
	inTmux := tmux.NewDetector().InTmuxSession()

	// CG mode requires tmux for pane-level environment isolation.
	if isHybrid && !inTmux && os.Getenv(config.EnvTestMode) != "1" {
		return fmt.Errorf("CG mode requires a tmux session: tmux is required for Claude + GLM hybrid mode because: Leader (this pane) uses Claude API, Teammates (new panes) inherit GLM env to use Z.AI API. Start a tmux session first: tmux new -s moai; moai cg. Or use 'moai glm' for all-GLM mode (no tmux required)")
	}

	// Inject GLM environment variables into tmux session (if available)
	if inTmux {
		if err := injectTmuxSessionEnv(glmConfig, apiKey); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to inject tmux session env: %v\n", err)
			if isHybrid {
				return fmt.Errorf("failed to inject GLM env into tmux session: %w (CG mode relies on tmux session env for teammate isolation, try restarting your tmux session)", err)
			}
		}
	}

	if isHybrid {
		if err := persistTeamMode(root, "cg"); err != nil {
			return fmt.Errorf("persist team mode: %w", err)
		}

		if err := removeGLMEnv(settingsPath); err != nil {
			return fmt.Errorf("clean up GLM env for CG mode: %w", err)
		}

		if err := ensureSettingsLocalJSON(settingsPath); err != nil {
			return fmt.Errorf("ensure settings.local.json: %w", err)
		}

		_, _ = fmt.Fprintln(out, renderSuccessCard(
			"CG mode enabled (Claude + GLM)",
			"",
			"Architecture: Lead (Claude) + Teammates (GLM)",
			"Isolation: tmux pane-level environment variables",
			"tmux session: active (GLM env vars injected for new panes)",
			"Config saved to: .moai/config/sections/llm.yaml",
			"",
			"How it works:",
			"  - This pane: No Z.AI env -> Claude models (lead)",
			"  - New panes: Inherit Z.AI env -> GLM models (teammates)",
			"",
			"IMPORTANT: Start Claude Code in THIS pane (not a new one).",
			"Opening a new tmux pane for the lead will cause it to use GLM.",
			"",
			"Next steps:",
			"  1. Start Claude Code in this pane: claude",
			"  2. Run workflow: /moai --team \"your task\"",
			"",
			"Run 'moai cc' to disable CG/GLM team mode.",
		))
	} else {
		if err := persistTeamMode(root, "glm"); err != nil {
			return fmt.Errorf("persist team mode: %w", err)
		}

		if err := injectGLMEnvForTeam(settingsPath, glmConfig, apiKey); err != nil {
			return fmt.Errorf("inject GLM env for team: %w", err)
		}

		tmuxStatus := "tmux session: active (env vars injected)"
		if !inTmux {
			tmuxStatus = "Warning: tmux session NOT DETECTED. GLM teammates require tmux for env propagation.\n" +
				"  Recommended steps:\n" +
				"    1. tmux new -s moai\n" +
				"    2. moai glm                # (re-run inside tmux to configure session env)\n" +
				"    3. claude\n" +
				"    4. /moai --team \"task\""
		}

		_, _ = fmt.Fprintln(out, renderSuccessCard(
			"GLM Team mode enabled",
			"",
			"Architecture: All agents use GLM models",
			"Display mode: tmux (split panes)",
			tmuxStatus,
			"Config saved to: .moai/config/sections/llm.yaml",
			"",
			"Role model mapping (dynamic teams):",
			"  - lead: glm-5.1",
				"Role model mapping (dynamic teams):",
				"  - lead: glm-5.1",
				"  - researcher/reviewer: glm-4.7-flash (read-only roles)",
				"  - analyst/architect: glm-5.1 (planning roles)",
				"  - implementer/tester/designer: glm-4.7 (write roles)",
				"",
				"Available models: glm-5.1, glm-4.7, glm-4.6, glm-4.5, glm-4.5-air",
				"",
				"Next steps:",
			"",
			"Next steps:",
			"  1. Ensure you're in a tmux session (tmux new -s moai)",
			"  2. Start Claude Code: claude",
			"  3. Run workflow: /moai --team \"your task\"",
			"",
			"Run 'moai cc' to disable GLM team mode.",
		))
	}

	return nil
}

// injectTmuxSessionEnv sets GLM environment variables at the tmux session level.
//
// Issue #742: Pre-computes MOAI_STATUSLINE_CONTEXT_SIZE from the High slot
// (Opus equivalent) so the Claude Code statusline reflects the real GLM model
// context window (128K/200K/etc.) instead of the Claude slot's nominal size
// (1M for the Opus slot).
func injectTmuxSessionEnv(glmConfig *GLMConfigFromYAML, apiKey string) error {
	if isTestEnvironment() {
		return nil
	}
	if !tmux.NewDetector().InTmuxSession() {
		return nil
	}

	vars := map[string]string{
		"ANTHROPIC_AUTH_TOKEN":           apiKey,
		"ANTHROPIC_BASE_URL":             glmConfig.BaseURL,
		"ANTHROPIC_DEFAULT_OPUS_MODEL":   glmConfig.Models.High,
		"ANTHROPIC_DEFAULT_SONNET_MODEL": glmConfig.Models.Medium,
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":  glmConfig.Models.Low,
		// Z.AI proxy compatibility: strip Anthropic beta headers and prompt caching
		"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS":    "1",
		"DISABLE_PROMPT_CACHING":                    "1",
		"API_TIMEOUT_MS":                            "3000000",
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
	}

	// Issue #742: Map the High slot model to its real context window so
	// statusline gauge reflects GLM limits, not Claude's Opus slot 1M nominal.
	if size := statusline.ResolveGLMContextWindow(glmConfig.Models.High); size > 0 {
		vars[config.EnvStatuslineContextSize] = strconv.Itoa(size)
	}

	mgr := tmux.NewSessionManager()
	return mgr.InjectEnv(context.Background(), vars)
}

// clearTmuxSessionEnv removes GLM environment variables from the tmux session.
// Called when switching back to Claude mode (moai cc).
// ANTHROPIC_AUTH_TOKEN is intentionally excluded: it may be an OAuth token
// that must survive mode switches. ANTHROPIC_BASE_URL serves as the GLM
// activation indicator — removing it is sufficient to deactivate GLM mode.
func clearTmuxSessionEnv() error {
	if isTestEnvironment() {
		return nil
	}
	if !tmux.NewDetector().InTmuxSession() {
		return nil
	}

	// ANTHROPIC_AUTH_TOKEN intentionally excluded from this list:
	// it may be an OAuth token that must survive mode switches.
	// ANTHROPIC_BASE_URL serves as the GLM activation indicator.
	vars := []string{
		"ANTHROPIC_BASE_URL",
		"ANTHROPIC_DEFAULT_OPUS_MODEL",
		"ANTHROPIC_DEFAULT_SONNET_MODEL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL",
		"CLAUDE_CONFIG_DIR",
		// Z.AI proxy compatibility flags
		"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS",
		"DISABLE_PROMPT_CACHING",
		"API_TIMEOUT_MS",
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC",
		// Issue #742: clear GLM context-size hint when leaving GLM mode
		config.EnvStatuslineContextSize,
	}

	mgr := tmux.NewSessionManager()
	_ = mgr.ClearEnv(context.Background(), vars) //nolint:errcheck // best-effort cleanup
	return nil
}

// persistTeamMode saves the team_mode value to .moai/config/sections/llm.yaml.
func persistTeamMode(projectRoot, mode string) error {
	sectionsDir := filepath.Join(filepath.Clean(projectRoot), defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	llmCfg, err := loadLLMSectionOnly(sectionsDir)
	if err != nil {
		return fmt.Errorf("load LLM section: %w", err)
	}

	llmCfg.TeamMode = mode

	return saveLLMSection(sectionsDir, llmCfg)
}

// ensureSettingsLocalJSON ensures settings.local.json exists with teammateMode=tmux.
// CG mode requires tmux, so we force tmux display to prevent inline fallback
// which would cause teammates to lose GLM env var inheritance (see #468).
func ensureSettingsLocalJSON(settingsPath string) error {
	var settings SettingsLocal

	// Read existing settings if file exists (skip empty files)
	if data, err := os.ReadFile(settingsPath); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parse settings.local.json: %w", err)
		}
	}

	if settings.Env == nil {
		settings.Env = make(map[string]string)
	}

	// Force tmux display mode: CG mode requires tmux for pane-level env isolation.
	// "auto" can fall back to inline mode, causing teammates to lose GLM env vars (#468).
	// Single source of truth: teammateMode native settings key in settings.local.json.
	settings.TeammateMode = "tmux"
	// Clean up legacy env var if present (superseded by native key).
	delete(settings.Env, "CLAUDE_CODE_TEAMMATE_DISPLAY")

	if len(settings.Env) == 0 {
		settings.Env = nil
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		return fmt.Errorf("write settings.local.json: %w", err)
	}

	return nil
}

// loadLLMSectionOnly loads only the LLM section from llm.yaml.
func loadLLMSectionOnly(sectionsDir string) (config.LLMConfig, error) {
	llmPath := filepath.Join(sectionsDir, "llm.yaml")

	if _, err := os.Stat(llmPath); os.IsNotExist(err) {
		return config.NewDefaultLLMConfig(), nil
	}

	data, err := os.ReadFile(llmPath)
	if err != nil {
		return config.LLMConfig{}, fmt.Errorf("read llm.yaml: %w", err)
	}

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
//
// API key preservation: if a non-GLM ANTHROPIC_AUTH_TOKEN already exists
// in settings.local.json (e.g. a user's Anthropic API key), it is saved as
// MOAI_BACKUP_AUTH_TOKEN before being overwritten. removeGLMEnv restores it.
// Note: Claude OAuth tokens live in ~/.claude/, not here, so OAuth is unaffected.
func injectGLMEnvForTeam(settingsPath string, glmConfig *GLMConfigFromYAML, apiKey string) error {
	var settings SettingsLocal

	// Read existing settings if file exists (skip empty files)
	if data, err := os.ReadFile(settingsPath); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parse settings.local.json: %w", err)
		}
	}

	if settings.Env == nil {
		settings.Env = make(map[string]string)
	}

	// Back up any existing ANTHROPIC_AUTH_TOKEN that is not the GLM key itself.
	// This preserves a Claude OAuth token so that removeGLMEnv can restore it.
	if existing := settings.Env["ANTHROPIC_AUTH_TOKEN"]; existing != "" && existing != apiKey {
		settings.Env["MOAI_BACKUP_AUTH_TOKEN"] = existing
	}

	// Inject GLM environment variables for teammates
	settings.Env["ANTHROPIC_AUTH_TOKEN"] = apiKey
	settings.Env["ANTHROPIC_BASE_URL"] = glmConfig.BaseURL
	settings.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"] = glmConfig.Models.High
	settings.Env["ANTHROPIC_DEFAULT_SONNET_MODEL"] = glmConfig.Models.Medium
	settings.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = glmConfig.Models.Low
	// Z.AI proxy compatibility: strip Anthropic beta headers and prompt caching
	settings.Env["CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"] = "1"
	settings.Env["DISABLE_PROMPT_CACHING"] = "1"
	settings.Env["API_TIMEOUT_MS"] = "3000000"
	settings.Env["CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC"] = "1"
	// Issue #742: pre-compute statusline context size from the High slot
	// (Opus equivalent) so SessionStart hook propagates it via tmux env.
	if size := statusline.ResolveGLMContextWindow(glmConfig.Models.High); size > 0 {
		settings.Env[config.EnvStatuslineContextSize] = strconv.Itoa(size)
	} else {
		// Clean up stale value from a prior session that resolved a known model.
		delete(settings.Env, config.EnvStatuslineContextSize)
	}

	// Force tmux display mode: GLM team mode uses tmux for env var inheritance.
	// "auto" can fall back to inline mode, causing teammates to lose GLM env vars (#468).
	// Single source of truth: teammateMode native settings key in settings.local.json.
	settings.TeammateMode = "tmux"
	// Clean up legacy env var if present (superseded by native key).
	delete(settings.Env, "CLAUDE_CODE_TEAMMATE_DISPLAY")

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		return fmt.Errorf("write settings.local.json: %w", err)
	}

	return nil
}

// saveLLMSection saves only the LLM section to llm.yaml.
// Empty GLM model values are populated with defaults to avoid confusion.
func saveLLMSection(sectionsDir string, llm config.LLMConfig) error {
	// Populate empty GLM model values with defaults for clarity.
	// This prevents llm.yaml from containing empty model strings that
	// confuse users about whether GLM is configured.
	defaults := config.NewDefaultLLMConfig()
	if llm.GLM.BaseURL == "" {
		llm.GLM.BaseURL = defaults.GLM.BaseURL
	}
	if llm.GLM.Models.High == "" {
		llm.GLM.Models.High = defaults.GLM.Models.High
	}
	if llm.GLM.Models.Medium == "" {
		llm.GLM.Models.Medium = defaults.GLM.Models.Medium
	}
	if llm.GLM.Models.Low == "" {
		llm.GLM.Models.Low = defaults.GLM.Models.Low
	}
	// Also populate legacy model name fields for consistency
	if llm.GLM.Models.Opus == "" {
		llm.GLM.Models.Opus = defaults.GLM.Models.Opus
	}
	if llm.GLM.Models.Sonnet == "" {
		llm.GLM.Models.Sonnet = defaults.GLM.Models.Sonnet
	}
	if llm.GLM.Models.Haiku == "" {
		llm.GLM.Models.Haiku = defaults.GLM.Models.Haiku
	}
	// Populate GLM env var if empty
	if llm.GLMEnvVar == "" {
		llm.GLMEnvVar = defaults.GLMEnvVar
	}

	wrapper := struct {
		LLM config.LLMConfig `yaml:"llm"`
	}{LLM: llm}

	data, err := yaml.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("marshal llm config: %w", err)
	}

	path := filepath.Join(sectionsDir, "llm.yaml")

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

// resolveGLMModels resolves the effective high, medium, and low model names.
func resolveGLMModels(models config.GLMModels) (high, medium, low string) {
	defaults := config.NewDefaultLLMConfig()

	high = models.High
	if high == "" {
		high = models.Opus
	}
	if high == "" {
		high = defaults.GLM.Models.High
	}

	medium = models.Medium
	if medium == "" {
		medium = models.Sonnet
	}
	if medium == "" {
		medium = defaults.GLM.Models.Medium
	}

	low = models.Low
	if low == "" {
		low = models.Haiku
	}
	if low == "" {
		low = defaults.GLM.Models.Low
	}

	return high, medium, low
}

// loadGLMConfig reads GLM configuration from llm.yaml.
func loadGLMConfig(root string) (*GLMConfigFromYAML, error) {
	if deps != nil && deps.Config != nil {
		cfg := deps.Config.Get()
		if cfg != nil && cfg.LLM.GLM.BaseURL != "" {
			high, medium, low := resolveGLMModels(cfg.LLM.GLM.Models)
			return &GLMConfigFromYAML{
				BaseURL: cfg.LLM.GLM.BaseURL,
				Models: struct {
					High   string
					Medium string
					Low    string
				}{
					High:   high,
					Medium: medium,
					Low:    low,
				},
				EnvVar: cfg.LLM.GLMEnvVar,
			}, nil
		}
	}

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
	home, err := userHomeDir()
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

	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	escapedKey := escapeDotenvValue(key)
	content := fmt.Sprintf("# GLM API Key for MoAI-ADK\n# Generated by moai glm\nGLM_API_KEY=\"%s\"\n", escapedKey)
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// loadGLMKey loads the GLM API key from ~/.moai/.env.glm.
func loadGLMKey() string {
	// Allow tests to simulate a specific GLM key without requiring a real
	// ~/.moai/.env.glm file. Only set this in test code via t.Setenv.
	if testKey := os.Getenv(config.EnvTestGLMKey); testKey != "" {
		return testKey
	}

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
func getGLMAPIKey(envVar string) string {
	if key := loadGLMKey(); key != "" {
		return key
	}
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
		// Z.AI proxy compatibility
		"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS":    "1",
		"DISABLE_PROMPT_CACHING":                    "1",
		"API_TIMEOUT_MS":                            "3000000",
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
	}
}

// injectGLMEnv adds GLM environment variables to settings.local.json.
//
// API key preservation: if a non-GLM ANTHROPIC_AUTH_TOKEN already exists
// in settings.local.json (e.g. a user's Anthropic API key), it is saved as
// MOAI_BACKUP_AUTH_TOKEN before being overwritten. removeGLMEnv restores it.
// Note: Claude OAuth tokens live in ~/.claude/, not here, so OAuth is unaffected.
func injectGLMEnv(settingsPath string, glmConfig *GLMConfigFromYAML) error {
	apiKey := getGLMAPIKey(glmConfig.EnvVar)
	if apiKey == "" {
		return fmt.Errorf("GLM API key not found. Run 'moai glm setup <api-key>' to save your key, or set %s environment variable", glmConfig.EnvVar)
	}

	var settings SettingsLocal

	// Read existing settings if file exists (skip empty files)
	if data, err := os.ReadFile(settingsPath); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parse settings.local.json: %w", err)
		}
	}

	if settings.Env == nil {
		settings.Env = make(map[string]string)
	}

	// Back up any existing ANTHROPIC_AUTH_TOKEN that is not the GLM key itself.
	// This preserves a Claude OAuth token so that removeGLMEnv can restore it.
	if existing := settings.Env["ANTHROPIC_AUTH_TOKEN"]; existing != "" && existing != apiKey {
		settings.Env["MOAI_BACKUP_AUTH_TOKEN"] = existing
	}

	// Inject GLM environment variables with actual API key value
	settings.Env["ANTHROPIC_AUTH_TOKEN"] = apiKey
	settings.Env["ANTHROPIC_BASE_URL"] = glmConfig.BaseURL
	settings.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"] = glmConfig.Models.High
	settings.Env["ANTHROPIC_DEFAULT_SONNET_MODEL"] = glmConfig.Models.Medium
	settings.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = glmConfig.Models.Low
	// Z.AI proxy compatibility: strip Anthropic beta headers and prompt caching
	settings.Env["CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"] = "1"
	settings.Env["DISABLE_PROMPT_CACHING"] = "1"
	settings.Env["API_TIMEOUT_MS"] = "3000000"
	settings.Env["CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC"] = "1"

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		return fmt.Errorf("write settings.local.json: %w", err)
	}

	return nil
}

// isTestEnvironment detects if we're running in a test environment.
func isTestEnvironment() bool {
	if flag := os.Getenv(config.EnvTestMode); flag == "1" {
		return true
	}
	// Check if running under go test by examining os.Args.
	// On Windows the test binary has a .test.exe suffix instead of .test.
	for _, arg := range os.Args {
		if strings.HasSuffix(arg, ".test") || strings.HasSuffix(arg, ".test.exe") || strings.Contains(arg, "go.test") {
			return true
		}
	}
	return false
}

// findProjectRoot finds the project root by looking for .moai directory.
// It skips the user's home directory to prevent treating ~/.moai/ (global cache)
// as a project root. The home directory's .moai/ is for global state only
// (credentials, cache, releases), not a project.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Normalize to resolve symlinks (macOS /private/var) and Windows 8.3 short paths.
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}

	homeDir, _ := userHomeDir()
	if homeDir != "" {
		if resolved, err := filepath.EvalSymlinks(homeDir); err == nil {
			homeDir = resolved
		}
	}

	for {
		// Skip home directory — ~/.moai/ is global state, not a project root
		if homeDir != "" && dir == homeDir {
			return "", fmt.Errorf("not in a MoAI project (reached home directory)")
		}
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
