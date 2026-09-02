package cli

// @MX:NOTE: [AUTO] GLM command launches Claude Code with GLM backend via Z.AI proxy
// @MX:NOTE: [AUTO] Requires 'moai glm setup <key>' to save API key to ~/.moai/.env.glm
// @MX:NOTE: [AUTO] Main session uses GLM: 128K/200K/1M context windows per model tier (high=glm-5.2)
// @MX:NOTE: [AUTO] WARNING block on stderr is plain fmt by design (non-TTY safe)

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/glmcred"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/statusline"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/tmux"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// init wires internal/cli's existing userHomeDirFn test-injection seam (declared
// in glm_tools.go) through to glmcred.HomeDirFn. The closure closes over the
// package-level var by reference, so a test that reassigns userHomeDirFn at
// runtime is observed by glmcred.Path/Save/Load without each call site having
// to be repointed. This preserves every existing internal/cli credential test
// after the M1 extraction (SPEC-GLM-KEY-INPUT-001 §D constraint).
func init() {
	glmcred.HomeDirFn = func() (string, error) { return userHomeDirFn() }
}

var glmCmd = &cobra.Command{
	Use:   "glm [-p profile] [-k [SPEC-ID] | -k --name <role> | -f [N] | -f lane-<n>] [-- claude-args...]",
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
  -w, --worktree [name]         Launch in an isolated git worktree (.claude/worktrees/<name>/);
                                name omitted = auto-generated (same as claude --worktree)
      --spawn                   Run this command in a new tmux window instead of
                                replacing the current session (requires tmux)

Kanban Mode:
  -k, --kanban [SPEC-ID]       Enter as the LEAD of a kanban run. Seeds a
                                plan -> run -> sync chain in this
                                session. The optional SPEC-ID ties the run to a
                                SPEC. The lead drives the whole chain; three
                                companion sessions are launched by hand.
  -k --name <role>             Enter as a COMPANION of an existing kanban run.
                                Joins the run without seeding a chain. The three
                                roles are: plan, run, sync. A role name
                                held by a live session is bumped to the next
                                free number (plan-1, plan-2, ...).

Factory Mode (dedicated -f entry):
  -f, --factory [N]            Enter as the LEAD of a factory run with N
                                numbered lanes; N omitted = one lane
                                (lane-1), grown afterwards with the
                                incremental form below. The lead routes
                                operator-picked cards to free lanes over
                                cross-session messages — each card goes
                                WHOLE to one lane, which carries it through
                                plan -> run -> sync in-session.
  -f lane-<n>                  Launch exactly one additional lane — lane
                                n — and connect it to the lead socket of the
                                running factory. A number whose label is held by a
                                live session is bumped to the next free number.
  -k <N> / -k <N> --name lane-<i>
                                The v1.2.0 unified -k factory shapes, still
                                valid: -k N is the lead of an N-lane run,
                                -k N --name lane-<i> is lane i of it (a
                                bare -k --name lane-<i> defaults to 8).
                                One entry token per launch: -k and -f
                                together is an error.

  Genealogy: the pre-3.1 "factory" flag (-f/--factory) was RENAMED to
  -k/--kanban in #1513 (7f61332ef) and now drives the three-role kanban chain
  above. -f briefly returned as the factory fan-out flag and was RETIRED
  (v1.2.0) in favor of '-k <N>'; t118 (v3.1.1) revived it as the dedicated
  factory entry — the kanban chain keeps -k, the factory gets -f.

Note: Auto mode is not available with GLM (third-party provider).
Use 'moai cc --permission-mode auto' or 'moai cg --permission-mode auto' instead.

Note: Z.AI enforces low concurrency limits (paid tiers observe 1-3 in-flight
requests). Multi-agent workflows that exceed this limit can surface as opaque
errors (sometimes misreported by clients as "context window limit"). The GLM
models themselves have ample context (glm-5.2 1M, glm-4.7 ~202K). For more
stable parallel execution with MoAI Agent Teams, prefer 'moai cg' (hybrid mode).

Examples:
  moai glm setup sk-xxx    # Save API key (one-time)
  moai glm                 # Launch with GLM backend
  moai glm -p work         # Use 'work' profile with GLM
  moai glm -k              # Kanban lead on GLM: seeds the chain
  moai glm -k --name run           # Kanban companion on GLM (the GLM-recommended role)
  moai glm -f              # Factory lead on GLM: one lane (lane-1)
  moai glm -f 4            # Factory lead on GLM: announces lane-1..lane-4
  moai glm -f lane-2       # Add lane 2 to the running factory (GLM backend)

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
		case "tools":
			// SPEC-GLM-MCP-001: Z.AI MCP server enable/disable routing.
			// glmToolsCmd uses standard cobra parsing (--scope, --force), so delegate to Execute().
			glmToolsCmd.SetArgs(args[1:])
			return glmToolsCmd.Execute()
		}
	}

	// --spawn: open a GLM session in a new tmux window and keep this session.
	// Placed after subcommand routing so `moai glm setup` is never intercepted.
	// See cc.go for the ordering rationale.
	if spawnArgs, spawn := stripSpawnFlag(args); spawn {
		return spawnLaunch(cmd.OutOrStdout(), "glm", spawnArgs)
	}

	profileName, filteredArgs, err := parseProfileFlag(args)
	if err != nil {
		return err
	}
	// t118 launcher axis: the unified entry parse — parseLauncherEntry covers
	// the -k shapes (SPEC-FACTORY-BOOTSTRAP-001 + the v1.2.0 factory shapes)
	// and the revived dedicated -f surface. See cc.go for the truth table;
	// glm mirrors cc exactly except for the backend constant.
	entry, err := parseLauncherEntry(filteredArgs)
	if err != nil {
		return err
	}
	filteredArgs = entry.Rest
	label, isCompanion := parseCompanionLabel(filteredArgs)
	factoryLabel, isFactoryLane := parseFactoryLaneLabel(filteredArgs)
	switch resolveFactoryBranch(entry.FactoryEnabled, isFactoryLane) {
	case factoryBranchLead:
		leadLabel, _ := parseLeadLabel(filteredArgs)
		defer enterFactoryLeadMode(entry.FactoryWorkers, leadLabel)()
		defer exportKanbanLaunchFacts(entry.Spec, kanban.BackendGLM)()
		var leadName string
		filteredArgs, leadName = appendLeadName(filteredArgs, launchProjectRoot(), cmd.ErrOrStderr())
		defer exportLeadSessionName(leadName)()
		settingsFlag, settingsCleanup := prepareKanbanSettings(filteredArgs)
		if len(settingsFlag) > 0 {
			filteredArgs = append(filteredArgs, settingsFlag...)
		}
		defer settingsCleanup()
	case factoryBranchWorker:
		// See cc.go: a live-held lane number is bumped, and the bumped value
		// must reach the backend argv.
		finalLabel := resolveFactoryWorkerName(launchProjectRoot(), factoryLabel, cmd.ErrOrStderr())
		filteredArgs = replaceNamedLabel(filteredArgs, factoryLabel, finalLabel)
		defer enterFactoryWorkerMode(finalLabel, entry.FactoryWorkers)()
		defer exportKanbanLaunchFacts(entry.Spec, kanban.BackendGLM)()
		settingsFlag, settingsCleanup := prepareKanbanSettings(filteredArgs)
		if len(settingsFlag) > 0 {
			filteredArgs = append(filteredArgs, settingsFlag...)
		}
		defer settingsCleanup()
	}
	if !entry.FactoryEnabled {
		switch resolveKanbanBranch(entry.KanbanEnabled, isCompanion) {
		case kanbanBranchLead:
			// See cc.go: the operator's lead run id is adopted rather than replaced.
			leadLabel, _ := parseLeadLabel(filteredArgs)
			defer enterKanbanMode(entry.Spec, leadLabel)()
			defer exportKanbanLaunchFacts(entry.Spec, kanban.BackendGLM)()
			// See cc.go: glm mirrors the lead branch exactly.
			var leadName string
			filteredArgs, leadName = appendLeadName(filteredArgs, launchProjectRoot(), cmd.ErrOrStderr())
			defer exportLeadSessionName(leadName)()
			settingsFlag, settingsCleanup := prepareKanbanSettings(filteredArgs)
			if len(settingsFlag) > 0 {
				filteredArgs = append(filteredArgs, settingsFlag...)
			}
			defer settingsCleanup()
		case kanbanBranchCompanion:
			// See cc.go: a live-held label is bumped, and the bumped value must
			// reach the backend argv.
			finalLabel := resolveCompanionName(launchProjectRoot(), label, cmd.ErrOrStderr())
			filteredArgs = replaceNamedLabel(filteredArgs, label, finalLabel)
			defer enterKanbanCompanionMode(finalLabel)()
			defer exportKanbanLaunchFacts(entry.Spec, kanban.BackendGLM)()
			settingsFlag, settingsCleanup := prepareKanbanSettings(filteredArgs)
			if len(settingsFlag) > 0 {
				filteredArgs = append(filteredArgs, settingsFlag...)
			}
			defer settingsCleanup()
		}
	}
	// SPEC-WORKTREE-ENTRY-STRATEGY-001 M3a: see cc.go for the rationale.
	if err := resolveWorktreeL2Path(filteredArgs); err != nil {
		return err
	}
	filteredArgs = normalizeWorktreeFlag(filteredArgs)

	// Auto mode is not available with third-party providers (GLM/Z.AI).
	// Validate before launch to give a clear error instead of a cryptic Claude Code rejection.
	if containsPermissionMode(filteredArgs, "auto") {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "auto mode requires Claude Sonnet 4.6 or Opus 4.6 running on Anthropic's API")
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "use 'moai cc --permission-mode auto' or 'moai cg --permission-mode auto' instead")
		return fmt.Errorf("auto mode is not available with GLM (third-party provider)")
	}

	// Warn about main-session GLM limitations before launch.
	// Z.AI concurrency limits (1-3 in-flight requests per paid tier) are sometimes
	// misreported by Claude Code as "context window limit".
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "WARNING: moai glm uses GLM models for the MAIN SESSION. Known limitations:")
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "  - Main session context window: 1M (glm-5.3, glm-5.3-flash)")
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

// glmAutoCompactWindow returns the CLAUDE_CODE_AUTO_COMPACT_WINDOW value and a
// flag indicating whether it should be injected. The window is only set when
// the High slot model resolves to the 1M context tier (e.g. glm-5.2), which
// activates Claude Code's 1M context mode — auto-compact must then scale to the
// enlarged window. The trigger is the model's resolved context window (via
// statusline.ResolveGLMContextWindow), not a model-id suffix: the suffix was
// retired after z.ai rejected a suffixed id such as glm-5.2[1m] (current
// Claude Code strips "[1m]" before sending the ID to the provider, but the
// resolved-window trigger also covers non-1M tiers). Note the value only takes
// effect together with glmMaxContextTokens: Claude Code caps the auto-compact
// window at the model's assumed window — 200K for unrecognized custom IDs
// unless CLAUDE_CODE_MAX_CONTEXT_TOKENS declares otherwise.
func glmAutoCompactWindow(highModel string) (string, bool) {
	if statusline.ResolveGLMContextWindow(highModel) >= config.Default1MContextTokens {
		return strconv.Itoa(config.Default1MContextTokens), true
	}
	return "", false
}

// glmMaxContextTokens returns the CLAUDE_CODE_MAX_CONTEXT_TOKENS value and a
// flag indicating whether it should be injected. Claude Code assumes a 200K
// window for any model ID it cannot resolve (no "claude-" prefix, no "[1m]"),
// which both mislabels the context telemetry and caps
// CLAUDE_CODE_AUTO_COMPACT_WINDOW at 200K. Declaring the resolved window —
// for every GLM tier, not just 1M — lifts both.
// Ref: code.claude.com/docs/en/model-config ("Correct the window for a gateway
// or custom model ID"); Issue #653.
func glmMaxContextTokens(highModel string) (string, bool) {
	if size := statusline.ResolveGLMContextWindow(highModel); size > 0 {
		return strconv.Itoa(size), true
	}
	return "", false
}

// setGLMEnv sets GLM environment variables in the current process.
// @MX:WARN: [AUTO] Global environment variable mutation without rollback mechanism
// @MX:REASON: Process-level state mutation affects all subsequent goroutines; no cleanup on error
func setGLMEnv(glmConfig *GLMConfigFromYAML, apiKey string) {
	_ = os.Setenv(config.EnvAnthropicAuthToken, apiKey)                           //nolint:errcheck
	_ = os.Setenv(config.EnvAnthropicBaseURL, glmConfig.BaseURL)                  //nolint:errcheck
	_ = os.Setenv(config.EnvAnthropicDefaultOpusModel, glmConfig.Models.High)     //nolint:errcheck
	_ = os.Setenv(config.EnvAnthropicDefaultSonnetModel, glmConfig.Models.Medium) //nolint:errcheck
	_ = os.Setenv(config.EnvAnthropicDefaultHaikuModel, glmConfig.Models.Low)     //nolint:errcheck
	_ = os.Setenv(config.EnvAnthropicDefaultFableModel, glmConfig.Models.Fable)   //nolint:errcheck
	// 1M context activation: when the High slot model resolves to the 1M context
	// tier, scale the auto-compact window to the full 1M context.
	if window, ok := glmAutoCompactWindow(glmConfig.Models.High); ok {
		_ = os.Setenv(config.EnvClaudeCodeAutoCompactWindow, window) //nolint:errcheck
	}
	// Declare the model's real context window (Issue #653): Claude Code
	// assumes 200K for the unrecognized custom ID and caps the auto-compact
	// window above at that assumption.
	if tokens, ok := glmMaxContextTokens(glmConfig.Models.High); ok {
		_ = os.Setenv(config.EnvClaudeCodeMaxContextTokens, tokens) //nolint:errcheck
	}
	// Z.AI proxy compatibility: strip Anthropic beta headers
	_ = os.Setenv(config.EnvClaudeCodeDisableExperimentalBetas, "1") //nolint:errcheck
	_ = os.Setenv("API_TIMEOUT_MS", "3000000")                       //nolint:errcheck
	// Z.AI MCP server (zai-mcp-server) reads this env for authentication.
	_ = os.Setenv("Z_AI_API_KEY", apiKey) //nolint:errcheck
	// GLM effort overlay (SPEC-MODEL-TIER-PLANTYPE-001 M5, REQ-MTP-030): inject the
	// session-global GLM reasoning-control derived from the effort overlay.
	for k, v := range glmReasoningEnvVars() {
		_ = os.Setenv(k, v) //nolint:errcheck
	}
}

// glmReasoningEnvVars returns the Branch-B explicit reasoning-control env
// injection for a GLM backend (SPEC-MODEL-TIER-PLANTYPE-001 M5, REQ-MTP-030).
// It is the overlay wire point invoked by setGLMEnv (process env). It derives
// the session-global GLM reasoning state from the effort overlay
// (template.SessionGLMReasoningState) and maps it to the ANTHROPIC_REASONING_EFFORT
// env var (thinking-enabled states) or the thinking toggle (thinking-off state).
// (The settings.local.json twin of this wire point, injectGLMEnvForTeam, was
// removed with its dead caller enableTeamMode in #1531.)
//
// Delivery status MEASURED, direction reversed post-close (SPEC-V3R6-AUDIT-MODEL-PIN-001
// acceptance.md AC-AMP-006 amendment, 2026-08-24, lead-approved; closes the
// AC-MTP-032b residual of SPEC-MODEL-TIER-PLANTYPE-001): the null-controlled
// live differential proved the top-level `reasoning_effort` request field is
// the effective delivery channel (ratios 1.34/1.85/1.48 against the 1.25 bound;
// thinking-budget null 1.02). The earlier t175 probe
// (.moai/reports/t175/measurements.md §3) reported the opposite direction and
// is superseded by that record. Delivered spend per level remains unquantified.
func glmReasoningEnvVars() map[string]string {
	state := template.SessionGLMReasoningState()
	out := make(map[string]string, 1)
	if state.ThinkingEnabled {
		out[config.EnvAnthropicReasoningEffort] = state.ReasoningEffort
	}
	return out
}

// glmReasoningEnvVarsForModel returns the MAIN-SESSION reasoning-control env
// injection derived from the web-set effort preference under the resolved
// session model. It is the prefs-driven counterpart to glmReasoningEnvVars()
// (which derives the hardcoded max session default — SPEC-GLM-EFFORT-MAX-001 —
// used for sub-agents and the empty-effort fallback). When
// effort is non-empty it collapses the Claude effort onto z.ai's reasoning
// control via SessionGLMReasoningStateForModel (which pins every effort to max
// under glm-5.3-flash — flash accepts reasoning_effort: max only — and
// collapses as before under any other model); when empty it falls back to the
// session default (glmReasoningEnvVars). Thinking-off states emit no
// ANTHROPIC_REASONING_EFFORT entry (reasoning_effort is moot when thinking is
// off). glmReasoningEnvVars() stays intact — setGLMEnv still calls it as the
// session-default wire point (and this helper falls back to it when effort is
// empty); this helper is ADDITIVE and is consumed only by the main-session
// launch path.
//
// Delivery status MEASURED, direction reversed post-close (SPEC-V3R6-AUDIT-MODEL-PIN-001
// acceptance.md AC-AMP-006 amendment, 2026-08-24 — supersedes the t175 probe):
// the top-level `reasoning_effort` request field is the effective delivery
// channel, so this env reaches z.ai through that field, not a thinking-budget
// mapping.
func glmReasoningEnvVarsForModel(model, effort string) map[string]string {
	state := template.SessionGLMReasoningStateForModel(model, effort)
	out := make(map[string]string, 1)
	if state.ThinkingEnabled {
		out[config.EnvAnthropicReasoningEffort] = state.ReasoningEffort
	}
	return out
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
	return injectTmuxSessionEnvVia(tmux.NewSessionManager(), glmConfig, apiKey)
}

// buildTmuxInjectVars returns the GLM environment variable set that is injected
// into the tmux session env for teammate isolation. Extracted as a pure helper
// (no tmux side effects) so the credential-routing invariant and the
// inject↔clear key-parity invariant (REQ-CGH-009) can be asserted directly.
// ANTHROPIC_AUTH_TOKEN is included here as the teammate-facing GLM credential set;
// the sensitive-channel routing (which deletes it from the bulk map) happens in
// injectTmuxSessionEnvVia.
func buildTmuxInjectVars(glmConfig *GLMConfigFromYAML, apiKey string) map[string]string {
	vars := map[string]string{
		config.EnvAnthropicAuthToken:          apiKey,
		config.EnvAnthropicBaseURL:            glmConfig.BaseURL,
		config.EnvAnthropicDefaultOpusModel:   glmConfig.Models.High,
		config.EnvAnthropicDefaultSonnetModel: glmConfig.Models.Medium,
		config.EnvAnthropicDefaultHaikuModel:  glmConfig.Models.Low,
		config.EnvAnthropicDefaultFableModel:  glmConfig.Models.Fable,
		// Z.AI proxy compatibility: strip Anthropic beta headers
		config.EnvClaudeCodeDisableExperimentalBetas: "1",
		"API_TIMEOUT_MS": "3000000",
	}

	// Issue #742: Map the High slot model to its real context window so
	// statusline gauge reflects GLM limits, not Claude's Opus slot 1M nominal.
	if size := statusline.ResolveGLMContextWindow(glmConfig.Models.High); size > 0 {
		vars[config.EnvStatuslineContextSize] = strconv.Itoa(size)
	}

	// 1M context activation: scale auto-compact window when the High slot model
	// resolves to the 1M context tier.
	if window, ok := glmAutoCompactWindow(glmConfig.Models.High); ok {
		vars[config.EnvClaudeCodeAutoCompactWindow] = window
	}

	// Issue #653: declare the same window to Claude Code itself — without the
	// declaration, CC assumes 200K for the unrecognized custom ID and caps the
	// auto-compact window above at 200K regardless of its configured value.
	if tokens, ok := glmMaxContextTokens(glmConfig.Models.High); ok {
		vars[config.EnvClaudeCodeMaxContextTokens] = tokens
	}

	return vars
}

// injectTmuxSessionEnvVia performs the actual credential routing through the given
// SessionManager. Split out from injectTmuxSessionEnv (which holds the
// test-environment / in-tmux guards) so tests can exercise the production routing
// path with a recording-fake SessionManager (REQ-CGH-009 Scenario 9a).
func injectTmuxSessionEnvVia(mgr tmux.SessionManager, glmConfig *GLMConfigFromYAML, apiKey string) error {
	vars := buildTmuxInjectVars(glmConfig, apiKey)

	// SPEC-V3R5-SECURITY-CRIT-001 P0-2 (CWE-214): route ANTHROPIC_AUTH_TOKEN
	// through the argv-safe sensitive injection channel. The remaining
	// non-sensitive vars stay on the fast bulk path. On sensitive-injection
	// failure we MUST NOT fall back to argv (would re-leak the token).
	ctx := context.Background()

	const sensitiveKey = config.EnvAnthropicAuthToken
	if token := vars[sensitiveKey]; token != "" {
		if err := mgr.InjectSensitiveEnv(ctx, sensitiveKey, token); err != nil {
			return fmt.Errorf("inject sensitive tmux env: %w", err)
		}
		delete(vars, sensitiveKey)
	}

	if len(vars) == 0 {
		return nil
	}
	return mgr.InjectEnv(ctx, vars)
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

	mgr := tmux.NewSessionManager()
	_ = mgr.ClearEnv(context.Background(), buildTmuxClearVars()) //nolint:errcheck // best-effort cleanup
	return nil
}

// buildTmuxClearVars returns the GLM environment variable keys removed from the
// tmux session env when switching back to Claude mode (moai cc). Extracted as a
// pure helper so the inject↔clear key-parity invariant (REQ-CGH-009) can be
// asserted directly.
//
// ANTHROPIC_AUTH_TOKEN is intentionally excluded from this list: it may be an
// OAuth token that must survive mode switches. ANTHROPIC_BASE_URL serves as the
// GLM activation indicator — removing it is sufficient to deactivate GLM mode.
func buildTmuxClearVars() []string {
	return []string{
		config.EnvAnthropicBaseURL,
		config.EnvAnthropicDefaultOpusModel,
		config.EnvAnthropicDefaultSonnetModel,
		config.EnvAnthropicDefaultHaikuModel,
		config.EnvAnthropicDefaultFableModel,
		// GLM effort overlay (SPEC-MODEL-TIER-PLANTYPE-001 M5, REQ-MTP-030 Branch-B
		// inject↔clear parity, REQ-CGH-009): clear the reasoning-control env when
		// leaving GLM mode so it does not leak into a subsequent `moai cc` session.
		config.EnvAnthropicReasoningEffort,
		"CLAUDE_CONFIG_DIR",
		// Z.AI proxy compatibility flags
		config.EnvClaudeCodeDisableExperimentalBetas,
		"API_TIMEOUT_MS",
		config.EnvClaudeCodeDisableNonessentialTraffic,
		// Legacy cleanup: DISABLE_PROMPT_CACHING was removed from GLM env injection.
		// Kept here only to clean up residual values from older sessions.
		"DISABLE_PROMPT_CACHING",
		// Issue #742: clear GLM context-size hint when leaving GLM mode
		config.EnvStatuslineContextSize,
		// Clear 1M auto-compact window when leaving GLM mode (Claude slot scales
		// auto-compact itself).
		config.EnvClaudeCodeAutoCompactWindow,
		// Issue #653: clear the declared GLM context window (inject↔clear parity).
		config.EnvClaudeCodeMaxContextTokens,
	}
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

// saveLLMSection saves only the LLM section to llm.yaml.
// Empty GLM model values are populated with defaults to avoid confusion.
//
// @MX:NOTE: [AUTO] RC1 change-detection gate (glm-settings-persist): the write
// is SKIPPED when the persisted llm.yaml already equals the desired section
// state semantically (llmSectionSemanticallyEqual below). The launcher calls
// this on EVERY moai glm/cg/cc launch; an unconditional typed re-marshal there
// destroyed hand-written comments, flipped the file mode 0644→0600
// (writeFileAtomic's perm), and touched mtime on every launch — reopening a
// lost-update window against concurrent writers.
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
	if llm.GLM.Models.Fable == "" {
		llm.GLM.Models.Fable = defaults.GLM.Models.Fable
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

	// RC1 change-detection gate: skip the write when the persisted file already
	// carries this exact section state. A read error leaves the gate undecided
	// and falls through to the write (the write itself then surfaces the error).
	if persisted, ok, err := readPersistedLLMSection(sectionsDir); err == nil && ok &&
		llmSectionSemanticallyEqual(persisted, wrapper.LLM) {
		return nil
	}

	data, err := yaml.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("marshal llm config: %w", err)
	}

	path := filepath.Join(sectionsDir, "llm.yaml")
	return writeFileAtomic(path, data, 0o600)
}

// readPersistedLLMSection reads the current llm.yaml into a config.LLMConfig.
// ok is false only when the file does not exist (a first write must proceed);
// an unreadable or unparseable file returns the error so the caller can fall
// through to the write path rather than silently skipping it.
func readPersistedLLMSection(sectionsDir string) (config.LLMConfig, bool, error) {
	data, err := os.ReadFile(filepath.Join(sectionsDir, "llm.yaml"))
	if err != nil {
		if os.IsNotExist(err) {
			return config.LLMConfig{}, false, nil
		}
		return config.LLMConfig{}, false, fmt.Errorf("read llm.yaml: %w", err)
	}
	wrapper := struct {
		LLM config.LLMConfig `yaml:"llm"`
	}{}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return config.LLMConfig{}, false, fmt.Errorf("parse llm.yaml: %w", err)
	}
	return wrapper.LLM, true, nil
}

// llmSectionSemanticallyEqual compares two LLM sections by VALUE, not by file
// bytes: both are config.LLMConfig structs, so formatting, key order, and
// comments are invisible to the comparison. This is what lets a hand-written
// llm.yaml (with comments) be recognized as already carrying the desired
// state, so the launcher skips the rewrite entirely.
func llmSectionSemanticallyEqual(a, b config.LLMConfig) bool {
	normalizeLLMSectionMaps(&a)
	normalizeLLMSectionMaps(&b)
	return reflect.DeepEqual(a, b)
}

// normalizeLLMSectionMaps replaces nil maps with empty maps so a section whose
// map fields were never set compares equal to one round-tripped through a
// yaml.Marshal that renders nil maps as `{}`. Deliberate zero-value
// equivalence: "key absent" and "key present but empty" carry no different
// intent for these mirrors.
func normalizeLLMSectionMaps(c *config.LLMConfig) {
	if c.Profiles == nil {
		c.Profiles = map[string]map[string]config.ModelEffort{}
	}
	if c.HarnessAgents == nil {
		c.HarnessAgents = map[string]map[string]config.ModelEffort{}
	}
	if c.AgentOverrides == nil {
		c.AgentOverrides = map[string]config.ModelEffort{}
	}
	if c.GLM.ContextWindows == nil {
		c.GLM.ContextWindows = map[string]int{}
	}
}

// GLMConfigFromYAML represents the GLM settings from llm.yaml.
type GLMConfigFromYAML struct {
	BaseURL string
	Models  struct {
		High   string
		Medium string
		Low    string
		Fable  string
	}
	// Effort carries the per-tier reasoning-effort preference (llm.glm.effort.*)
	// through to the launcher (RC3, glm-settings-persist): resolveGLMMainSessionEffort
	// lets a non-empty slot value override the prefs/model_policy effort chain for
	// the main session. Empty slots mean "unset" — the prefs chain applies.
	Effort config.GLMTierEffort
	EnvVar string
}

// resolveGLMModels resolves the effective high, medium, low, and fable model names.
func resolveGLMModels(models config.GLMModels) (high, medium, low, fable string) {
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

	fable = models.Fable
	if fable == "" {
		fable = defaults.GLM.Models.Fable
	}

	return high, medium, low, fable
}

// loadGLMConfig reads GLM configuration from llm.yaml.
//
// The GLM command path (runGLM → unifiedLaunch → applyGLMMode) does not call
// deps.Config.Load(), so deps.Config.Get() returns nil at runtime. Reading the
// project's llm.yaml directly from disk (via loadLLMSectionOnly) ensures the
// user's configured models are honored regardless of whether deps.Config has
// been loaded — and independently of whether base_url is explicitly set. This
// is the fix for issue #1065 (user models silently ignored). When the in-memory
// config is already populated (e.g. deps.Config.Load was called elsewhere), it
// is preferred to avoid a redundant disk read.
func loadGLMConfig(root string) (*GLMConfigFromYAML, error) {
	defaults := config.NewDefaultLLMConfig()

	llm := defaults
	switch {
	case deps != nil && deps.Config != nil && deps.Config.Get() != nil:
		// In-memory config is populated; use it.
		llm = deps.Config.Get().LLM
	default:
		// Fall back to reading llm.yaml directly from the project root. This is
		// the live runtime path for the glm command, where deps.Config is unloaded.
		sectionsDir := filepath.Join(filepath.Clean(root), defs.MoAIDir, defs.SectionsSubdir)
		diskLLM, err := loadLLMSectionOnly(sectionsDir)
		if err != nil {
			return nil, fmt.Errorf("load GLM config from llm.yaml: %w", err)
		}
		llm = diskLLM
	}

	baseURL := llm.GLM.BaseURL
	if baseURL == "" {
		baseURL = defaults.GLM.BaseURL
	}

	envVar := llm.GLMEnvVar
	if envVar == "" {
		envVar = defaults.GLMEnvVar
	}

	high, medium, low, fable := resolveGLMModels(llm.GLM.Models)
	return &GLMConfigFromYAML{
		BaseURL: baseURL,
		Models: struct {
			High   string
			Medium string
			Low    string
			Fable  string
		}{
			High:   high,
			Medium: medium,
			Low:    low,
			Fable:  fable,
		},
		Effort: llm.GLM.Effort,
		EnvVar: envVar,
	}, nil
}

// glmSlotEffortForModel resolves the stored per-tier effort (llm.glm.effort.*)
// for the GLM tier slot serving the main-session model (RC3,
// glm-settings-persist).
//
// Thin delegation to template.GLMSlotEffortForModel, which owns the ONE
// alias/slot mapping in the tree (mirroring setGLMEnv's
// ANTHROPIC_DEFAULT_*_MODEL assignments: opus feeds the high slot, sonnet
// medium, haiku low, fable fable, with the [1m] suffix split and canonical
// claude-* ids reverse-mapped). Sharing that mapping is what keeps this effort
// half and the model half (template.GLMSlotModelOrHigh, which keys the wire
// collapse) from drifting apart — t360 repaired exactly such a drift.
//
// An empty or unknown model (including a raw GLM id, which is not an alias)
// resolves "": the caller falls back to the prefs effort chain unchanged,
// byte-identical to the pre-RC3 launch.
func glmSlotEffortForModel(model string, effort config.GLMTierEffort) string {
	return template.GLMSlotEffortForModel(effort, model)
}

// resolveGLMMainSessionEffort resolves the effort fed to the GLM launch
// overlay. Precedence: a non-empty stored glm.effort[slot] for the slot
// serving the main-session model WINS over the prefs/model_policy chain
// (fallback = resolveLaunchEffort's result); an empty stored value or a model
// with no slot claim keeps the fallback unchanged.
//
// The downstream collapse overlay stays governing for the final wire value
// (glmReasoningEnvVarsForModel → SessionGLMReasoningStateForModel,
// SPEC-GLM-EFFORT-MAX-001): a stored "high" and a stored "max" BOTH reach
// z.ai as reasoning_effort=max (every Claude effort above low collapses to
// max), a stored "low" wires as low, and under glm-5.3-flash every stored
// effort pins to max (flash accepts reasoning_effort: max only). Callers
// choose the stored values from that z.ai state vocabulary, so no additional
// translation happens here.
func resolveGLMMainSessionEffort(model string, tier config.GLMTierEffort, fallback string) string {
	if stored := glmSlotEffortForModel(model, tier); stored != "" {
		return stored
	}
	return fallback
}

// getGLMEnvPath returns the path to ~/.moai/.env.glm.
//
// Thin delegation to internal/glmcred (SPEC-GLM-KEY-INPUT-001 M1, D-1) so that
// exactly one credential-writer implementation survives in the tree. The home
// directory is resolved through glmcred.HomeDirFn, which is aliased to this
// package's userHomeDirFn in init() below to preserve the existing test-injection
// seam (internal/cli tests override userHomeDirFn).
func getGLMEnvPath() string {
	return glmcred.Path()
}

// saveGLMKey saves the GLM API key to ~/.moai/.env.glm at mode 0600.
//
// Delegates to glmcred.Save. Behaviour change in SPEC-GLM-KEY-INPUT-001 M1:
// glmcred.Save explicitly asserts mode 0600 after writing, so a pre-existing
// ~/.moai/.env.glm created at a wider mode by an older writer is tightened to
// 0600 here too (REQ-GKI-006-002 / D-3).
func saveGLMKey(key string) error {
	return glmcred.Save(key)
}

// loadGLMKey loads the GLM API key from ~/.moai/.env.glm, honouring the
// MOAI_TEST_GLM_KEY test seam first. Delegates to glmcred.Load.
func loadGLMKey() string {
	return glmcred.Load()
}

// escapeDotenvValue / unescapeDotenvValue delegate to glmcred. Retained as
// thin wrappers because the existing internal/cli test suite (glm_new_test,
// coverage_improvement_test, target_coverage_test, glm_test) calls them
// directly; repointing every call site would be a no-op change that risks
// breaking test compilation.
func escapeDotenvValue(value string) string   { return glmcred.EscapeValue(value) }
func unescapeDotenvValue(value string) string { return glmcred.UnescapeValue(value) }

// getGLMAPIKey returns the GLM API key from multiple sources.
func getGLMAPIKey(envVar string) string {
	if key := loadGLMKey(); key != "" {
		return key
	}
	return os.Getenv(envVar)
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

	// SPEC-CLIFIX-CRITICAL-001 REQ-CRIT-001-001: round-trip as map[string]any so
	// unknown top-level keys survive the write.
	// SPEC-CLIFIX-CONCURRENCY-001 REQ-CONC-001-001: route through the locked+atomic
	// mutateSettingsLocal seam so concurrent sessions cannot lose updates.
	return mutateSettingsLocal(settingsPath, func(m map[string]any) {
		env := settingsEnvMap(m)

		// Back up any existing ANTHROPIC_AUTH_TOKEN that is not the GLM key itself.
		// This preserves a Claude OAuth token so that removeGLMEnv can restore it.
		if existing, ok := env[config.EnvAnthropicAuthToken].(string); ok && existing != "" && existing != apiKey {
			env["MOAI_BACKUP_AUTH_TOKEN"] = existing
		}

		// Inject GLM environment variables with actual API key value
		env[config.EnvAnthropicAuthToken] = apiKey
		env[config.EnvAnthropicBaseURL] = glmConfig.BaseURL
		env[config.EnvAnthropicDefaultOpusModel] = glmConfig.Models.High
		env[config.EnvAnthropicDefaultSonnetModel] = glmConfig.Models.Medium
		env[config.EnvAnthropicDefaultHaikuModel] = glmConfig.Models.Low
		env[config.EnvAnthropicDefaultFableModel] = glmConfig.Models.Fable
		// Z.AI proxy compatibility: strip Anthropic beta headers
		env[config.EnvClaudeCodeDisableExperimentalBetas] = "1"
		env["API_TIMEOUT_MS"] = "3000000"
		// 1M context activation: scale auto-compact window when the High slot model
		// resolves to the 1M context tier; otherwise clean up any stale value.
		if window, ok := glmAutoCompactWindow(glmConfig.Models.High); ok {
			env[config.EnvClaudeCodeAutoCompactWindow] = window
		} else {
			delete(env, config.EnvClaudeCodeAutoCompactWindow)
		}
		// Issue #653: declare the model's real context window (Claude Code
		// assumes 200K for unrecognized custom IDs and caps the auto-compact
		// window at it); otherwise clean up any stale value.
		if tokens, ok := glmMaxContextTokens(glmConfig.Models.High); ok {
			env[config.EnvClaudeCodeMaxContextTokens] = tokens
		} else {
			delete(env, config.EnvClaudeCodeMaxContextTokens)
		}
		if len(env) == 0 {
			delete(m, "env")
		}
	})
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
