// Package config provides configuration management for MoAI-ADK Go Edition.
// It loads YAML section files, applies defaults, validates, and provides
// thread-safe access to configuration values.
package config

// @MX:NOTE: [AUTO] Environment variable key constants centralize all env var names to prevent typos and enable IDE navigation
//
// Environment variable key constants.
//
// Centralizes all environment variable names used across the codebase
// to prevent typos and enable IDE navigation.

// MoAI configuration environment variables.
const (
	// EnvHome overrides the ~/.moai root directory: when set to a non-empty
	// absolute path, the entire MoAI state tree (state/, cache/, releases/,
	// worktrees/, claude-profiles/, .env.glm, user-tier settings) resolves
	// under it instead of the user's home. An empty value equals unset and
	// relative values are disregarded (XDG semantics). Resolved exclusively
	// via internal/paths (SPEC-V3R6-MOAI-HOME-PATHS-001 REQ-MHP-001/004);
	// internal/paths mirrors this constant locally to stay stdlib-only.
	EnvHome = "MOAI_HOME"

	// EnvConfigDir overrides the MoAI configuration directory path.
	EnvConfigDir = "MOAI_CONFIG_DIR"

	// EnvConfigCacheDisabled disables the disk config cache layer
	// (LoadWithCache → full Load path, no cache read or write). Set to "1" or
	// "true" to opt out. This is a debug / test escape hatch: the cache is an
	// optimization whose write side effect (creating <configDir>/state/) can
	// pollute tests that assert on filesystem state (e.g. doctor golden
	// snapshots). Production config values are unaffected — only the cache
	// optimization is bypassed (SPEC-HOOK-PRETOOL-PERF-001).
	EnvConfigCacheDisabled = "MOAI_CONFIG_CACHE_DISABLED"

	// EnvDevelopmentMode overrides the development methodology (ddd or tdd).
	EnvDevelopmentMode = "MOAI_DEVELOPMENT_MODE"

	// EnvLogLevel overrides the log level (debug, info, warn, error).
	EnvLogLevel = "MOAI_LOG_LEVEL"

	// EnvLogFormat overrides the log format (text or json).
	EnvLogFormat = "MOAI_LOG_FORMAT"

	// EnvNoColor disables color output when set to "true" or "1".
	EnvNoColor = "MOAI_NO_COLOR"

	// EnvStatuslineMode selects the statusline display mode.
	EnvStatuslineMode = "MOAI_STATUSLINE_MODE"

	// EnvStatuslineContextSize overrides the context window size used by the
	// statusline gauge. Useful when the upstream provider reports a context
	// size that does not match the actual API limit (e.g., GLM models served
	// behind the Anthropic-compatible endpoint, see issue #653).
	// Value is parsed as int (tokens). Zero or invalid → fall back to stdin.
	EnvStatuslineContextSize = "MOAI_STATUSLINE_CONTEXT_SIZE"

	// EnvSkipBinaryUpdate skips binary self-update when set to "1".
	EnvSkipBinaryUpdate = "MOAI_SKIP_BINARY_UPDATE"

	// EnvGLMNoAutoTools skips automatic Z.AI MCP server enable on moai glm launch.
	EnvGLMNoAutoTools = "MOAI_GLM_NO_AUTO_TOOLS"

	// EnvGitConvention overrides the git commit convention.
	EnvGitConvention = "MOAI_GIT_CONVENTION"

	// EnvEnforceOnPush overrides the pre-push convention enforcement flag.
	EnvEnforceOnPush = "MOAI_ENFORCE_ON_PUSH"

	// EnvPrePushMode overrides the per-mode pre-push severity dial
	// (skip / warn / enforce). It sits BELOW the enforce_on_push gate: setting
	// this value does NOT enable enforcement; the gate remains governed solely
	// by EnvEnforceOnPush / git_convention.validation.enforce_on_push.
	EnvPrePushMode = "MOAI_PRE_PUSH"

	// EnvUpdateSource overrides the update source ("github" or "local").
	EnvUpdateSource = "MOAI_UPDATE_SOURCE"

	// EnvReleasesDir specifies a local directory for release archives.
	EnvReleasesDir = "MOAI_RELEASES_DIR"

	// EnvUpdateURL overrides the GitHub releases API URL.
	EnvUpdateURL = "MOAI_UPDATE_URL"

	// EnvNoProfileFallback disables the last-used-profile fallback for bare
	// `moai cc`/`glm`/`cg` invocations. When set to "1", a bare launch with no
	// -p flag uses the default profile even when the default is empty and a
	// named profile was previously launched. Users who want strict default
	// semantics set this to opt out of the fallback.
	EnvNoProfileFallback = "MOAI_NO_PROFILE_FALLBACK"

	// EnvSecurityCommitReview activates the in-session security guardian's
	// Layer-3 commit-time cross-file review (SPEC-SEC-GUARDIAN-001). Layer 3 is
	// dormant (a silent no-op) unless this flag is set to a truthy value. Aligned
	// with the MOAI_SYNC_GATE_BLOCKING opt-in precedent.
	EnvSecurityCommitReview = "MOAI_SECURITY_COMMIT_REVIEW"

	// EnvSecurityTurnReview opts the guardian's Layer-2 turn-diff review into a
	// model-backed / agentic escalation (SPEC-SEC-GUARDIAN-001). Unset, Layer 2
	// stays regex-only and advisory; the hook itself never invokes a sub-model —
	// the orchestrator translates the hook's structured signal into an Agent()
	// review.
	EnvSecurityTurnReview = "MOAI_SECURITY_TURN_REVIEW"

	// EnvSecurityBlocking promotes a guardian finding from advisory to a blocking
	// decision (SPEC-SEC-GUARDIAN-001). Unset, every guardian layer is advisory.
	// Aligned with the MOAI_SYNC_GATE_BLOCKING opt-in precedent.
	EnvSecurityBlocking = "MOAI_SECURITY_BLOCKING"

	// EnvSessionWorktree overrides the session-worktree auto-entry activation
	// (SPEC-SESSION-WORKTREE-001 REQ-SW-003). "1" forces the feature ON
	// regardless of the workflow.session_worktree.enabled config flag; "0"
	// forces it OFF. Any other value (including unset) falls through to the
	// config flag. Env wins over config.
	EnvSessionWorktree = "MOAI_SESSION_WORKTREE"

	// EnvAutonomyTier is the MOAI_AUTONOMY_TIER mode token
	// (SPEC-STOPCHAIN-TRIM-001 REQ-003). The SINGLE canonical source for the
	// autonomy tier — an env-key (NOT a workflow.yaml key) because pure shell
	// hooks (handle-stop-goal.sh, sync-phase-quality-gate.sh) MUST be able to
	// read it via "$MOAI_AUTONOMY_TIER" without invoking the moai binary or
	// parsing YAML. The Go side reads it via os.Getenv(EnvAutonomyTier); see
	// AutonomyTier() in autonomy.go. Values: semi-auto | automatic |
	// fully-autonomous (see defaults.go). Unset/empty/invalid → semi-auto
	// (REQ-003 backward compat).
	EnvAutonomyTier = "MOAI_AUTONOMY_TIER"

	// EnvSandboxProof carries a sandbox/container proof marker set by a
	// container/VM launcher (Docker, Firecracker, gVisor, E2B, devcontainer)
	// OR by the --sandbox-proof CLI flag. A non-empty value attests that the
	// session runs inside an OS-level sandbox, which is the precondition the
	// fully-autonomous tier (bypassPermissions) requires
	// (SPEC-AUTONOMY-TIERS-001 REQ-002 / OQ-1). The value names the isolation
	// tech (e.g. "docker") and is recorded in the downgrade audit log. A git
	// worktree is NOT a sandbox (it isolates the working tree, not the
	// process/OS).
	EnvSandboxProof = "MOAI_SANDBOX_PROOF"

	// EnvDisableBypassPermissionsMode is the env seam for the Claude Code
	// documented enterprise kill-switch disableBypassPermissionsMode
	// (SPEC-AUTONOMY-TIERS-001 REQ-005). When the managed/enterprise config
	// layer sets it to a truthy value, fully-autonomous is unselectable in
	// every surface and an existing bypassPermissions session downgrades to
	// automatic at the next resolution. MoAI does not ship a managed-config
	// file loader; the managed layer injects this env var instead.
	EnvDisableBypassPermissionsMode = "MOAI_DISABLE_BYPASS_PERMISSIONS_MODE"

	// EnvMoaiKanban carries the Kanban Mode signal from the launcher entry
	// point to the block-cap inject further down the launch chain. The launcher
	// sets it on the process environment before launching (restoring the prior
	// value and prior presence afterwards) rather than threading a parameter
	// through the chain, because the launch environment is already derived from
	// os.Environ() at the inject's call site — so the variable reaches both the
	// inject and the child session without a signature change. A non-empty
	// value means the session is a kanban session.
	//
	// This variable is load-bearing: removing or renaming it silently disables
	// the raised Stop-hook block cap, and the failure is quiet — the chain
	// simply stops after the default number of consecutive blocks.
	EnvMoaiKanban = "MOAI_KANBAN"

	// EnvMoaiKanbanSpec names the SPEC a kanban chain targets. It is set only
	// when the operator supplied an identifier; its absence means the chain
	// begins at plan-phase from the operator's first prompt.
	EnvMoaiKanbanSpec = "MOAI_KANBAN_SPEC"

	// EnvMoaiKanbanID carries the run identifier that distinguishes one kanban
	// run from another on the same machine. The lead session generates it once at
	// launch; the SessionStart hook reads it to name itself and the leader
	// socket. It is lead-owned state: a companion neither carries nor publishes
	// it (companion names are bare roles, so no companion surface holds a run id
	// that could disagree with the lead's).
	EnvMoaiKanbanID = "MOAI_KANBAN_ID"

	// EnvMoaiKanbanLabel marks a session as a COMPANION of a kanban run, and
	// carries its label — the bare role name, or the bumped `<role>-<n>` form a
	// collision produces. It is deliberately distinct from
	// EnvMoaiKanban: a companion needs the raised Stop-hook block cap (it arms
	// its own goal mid-session, exactly like the lead) but must NOT be seeded
	// with the plan -> run -> verify -> sync chain, which only the lead drives.
	// Setting EnvMoaiKanban on a companion would give every session the whole
	// chain to drive.
	EnvMoaiKanbanLabel = "MOAI_KANBAN_LABEL"

	// EnvMoaiKanbanSettingsInjected signals to the SessionStart hook that the
	// launcher wrote a transient settings file carrying
	// {"crossSessionInbound": "accept"} and passed it to the backend via
	// --settings. The hook reads it to decide which inbound-automation notice
	// line to print: when set to "1", cross-session messages are auto-accepted
	// (no operator action needed); when unset in a kanban session, the hook
	// prints the operator advisory instead (verify the field is present in the
	// operator's own --settings file, or it was not injected due to a fail-open
	// write failure).
	EnvMoaiKanbanSettingsInjected = "MOAI_KANBAN_SETTINGS_INJECTED"

	// EnvMoaiKanbanLeadAddr carries the leader socket path — the address on
	// the cross-session messaging substrate that companions send messages to.
	// Set by the launcher when enterKanbanMode classifies a lead, read by the
	// SessionStart hook to surface the address in the lead notice.
	EnvMoaiKanbanLeadAddr = "MOAI_KANBAN_LEAD_ADDR"

	// EnvMoaiFactoryWorkers carries the Factory Mode signal and the run's
	// worker count from the launcher entry point to the block-cap inject and
	// the SessionStart hook. It is set on BOTH the factory lead and every
	// worker — the count travels in the worker's own `-f <N>` token, which is
	// why the worker launch command carries it. A non-empty value is what
	// marks a session a factory session; the value is the fan-out size N.
	//
	// A factory run reuses EnvMoaiKanbanID and EnvMoaiKanbanLeadAddr on the
	// lead (run id, leader socket) and deliberately does NOT set
	// EnvMoaiKanban or EnvMoaiKanbanLabel: those seed the four-role kanban
	// chain, which a factory run never drives.
	EnvMoaiFactoryWorkers = "MOAI_FACTORY_WORKERS"

	// EnvMoaiFactoryWorker marks a session as a WORKER of a factory run and
	// carries its `lane-<n>` label. It is the factory counterpart of
	// EnvMoaiKanbanLabel: the lane needs the raised Stop-hook block cap
	// (it fields long dispatch-driven turns) but must not be seeded with any
	// chain, and the lead is signalled by EnvMoaiFactoryWorkers instead.
	EnvMoaiFactoryWorker = "MOAI_FACTORY_WORKER"

	// EnvMoaiSessionPID carries an explicit override for the PID recorded in
	// the multi-session coordination registry. A hook subprocess exits within
	// milliseconds of registering, so its own PID is worthless to the liveness
	// probe that reads the registry back; the registry needs the PID of the
	// long-lived session process instead. The resolver normally derives that
	// PID by walking the process ancestry, and reads this variable first when
	// a caller knows the session PID outright (the launcher path, or a test
	// injecting a known-live PID).
	EnvMoaiSessionPID = "MOAI_SESSION_PID"

	// EnvChainNodeID carries the origin-trail chain node ID from the spawning
	// context to the child process. Set by the spawner (moai cc -w,
	// EnterWorktree, Agent isolation:worktree) on the child environment before
	// exec'ing into Claude Code; read by the child's SessionStart handler to
	// re-inject after /clear env loss (REQ-CHAIN-013).
	// SPEC-CHAIN-CORE-001 REQ-CHAIN-006.
	EnvChainNodeID = "MOAI_CHAIN_NODE_ID"
)

// GLM inject/clear env-var names (set onto the process env when entering
// GLM / CG mode, and deleted when leaving it). These three keys are the
// canonical GLM env-var set per SPEC-CLIFIX-HYGIENE-001 REQ-HYG-001-003 —
// every inject site AND every clear site references these constants (or the
// GLMEnvVarSet helper below) so the inject↔clear sets cannot drift apart.
//
// @MX:ANCHOR: [AUTO] GLM env-var name SSOT — single source for the 3 keys
// @MX:REASON: REQ-HYG-001-003 inject↔clear parity invariant; renaming one site without the others silently breaks GLM mode activation/deactivation
const (
	// EnvClaudeCodeDisableExperimentalBetas strips Anthropic beta headers for
	// Z.AI proxy compatibility. Injected at every GLM env build; cleared on exit.
	EnvClaudeCodeDisableExperimentalBetas = "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"

	// EnvClaudeCodeDisableNonessentialTraffic disables Anthropic non-essential
	// (telemetry) traffic while in GLM mode. Cleared on exit.
	EnvClaudeCodeDisableNonessentialTraffic = "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC"

	// EnvClaudeCodeTeammateDisplay is the legacy GLM activation indicator env
	// var (superseded by the native settings.local.json teammateMode key but
	// still cleared to deactivate legacy GLM mode).
	EnvClaudeCodeTeammateDisplay = "CLAUDE_CODE_TEAMMATE_DISPLAY"
)

// GLMEnvVarSet returns the canonical GLM inject/clear env-var name set. It is
// the single enumeration consumed by both inject and clear paths, so the two
// are structurally identical by construction (SPEC-CLIFIX-HYGIENE-001
// REQ-HYG-001-003). Callers that clear the whole set may iterate this slice;
// callers that clear or inject a subset reference the individual constants
// above. The returned slice MUST NOT be mutated by callers.
func GLMEnvVarSet() []string {
	return []string{
		EnvClaudeCodeDisableExperimentalBetas,
		EnvClaudeCodeDisableNonessentialTraffic,
		EnvClaudeCodeTeammateDisplay,
	}
}

// MoAI test-only environment variables.
const (
	// EnvTestMode enables test mode behavior when set to "1".
	EnvTestMode = "MOAI_TEST_MODE"

	// EnvTestGLMKey provides a test GLM API key for integration tests.
	EnvTestGLMKey = "MOAI_TEST_GLM_KEY"
)

// Claude Code environment variables (set by Claude Code runtime).
const (
	// EnvClaudeProjectDir is the project root directory set by Claude Code.
	EnvClaudeProjectDir = "CLAUDE_PROJECT_DIR"

	// EnvClaudeConfigDir is the Claude Code configuration directory.
	EnvClaudeConfigDir = "CLAUDE_CONFIG_DIR"

	// EnvClaudeEnvFile is the path to Claude Code's environment file.
	EnvClaudeEnvFile = "CLAUDE_ENV_FILE"

	// EnvClaudeAutoCompactPct overrides the auto-compact percentage threshold.
	EnvClaudeAutoCompactPct = "CLAUDE_AUTOCOMPACT_PCT_OVERRIDE"

	// EnvClaudeCodeAutoCompactWindow sets the auto-compact threshold window in
	// tokens (integer). Set to the full 1M context size when a [1m]-suffixed
	// model activates Claude Code's 1M context mode, so auto-compact scales to
	// the enlarged window rather than the default ceiling.
	EnvClaudeCodeAutoCompactWindow = "CLAUDE_CODE_AUTO_COMPACT_WINDOW"

	// EnvClaudeCodeMaxContextTokens declares the context window Claude Code
	// should assume for a custom (non-Claude) model ID routed through
	// ANTHROPIC_BASE_URL. Claude Code assumes 200K for unrecognized IDs and
	// caps CLAUDE_CODE_AUTO_COMPACT_WINDOW at that assumed window, so GLM
	// tiers (128K/200K/1M) must be declared explicitly.
	// Ref: https://code.claude.com/docs/en/model-config ("Correct the window
	// for a gateway or custom model ID"); Issue #653.
	EnvClaudeCodeMaxContextTokens = "CLAUDE_CODE_MAX_CONTEXT_TOKENS"

	// EnvClaudeCodeEffortLevel sets the session effort level for Claude Code.
	// Valid values: "low", "medium", "high", "xhigh", "max".
	// "xhigh" and "max" are supported on Opus 4.7+.
	// When empty, the runtime default applies.
	EnvClaudeCodeEffortLevel = "CLAUDE_CODE_EFFORT_LEVEL"

	// EnvClaudeCodeStopHookBlockCap is the runtime consecutive-Stop-hook-block
	// cap (default 8). It is the silent terminator that pre-empts the goal loop
	// before MaxTurns fires (effective bound min(MaxTurns, cap) today).
	// SPEC-INFINITE-GOAL-001 REQ-2: when a goal is armed at --max-turns 0
	// (infinite), the launcher injects a raised value so the infinite loop
	// actually persists. MoAI does not own this env (it is a Claude Code runtime
	// env); the const centralizes the name per CLAUDE.local.md §14.
	EnvClaudeCodeStopHookBlockCap = "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP"

	// EnvClaudeCodeMaxConcurrentSubagents is the runtime per-session cap on
	// concurrently running subagents (default 20 per turn). t118 (v3.1.1
	// launcher axis): the launcher seeds it on every kanban companion and
	// factory worker lane so one lane's fan-out cannot crowd out the others —
	// the operator-confirmed architecture is DefaultLaneMaxConcurrentSubagents
	// (10) agents in parallel per lane. MoAI does not own this env (it is a
	// Claude Code runtime env); the const centralizes the name per
	// CLAUDE.local.md §14.
	EnvClaudeCodeMaxConcurrentSubagents = "CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS"
)

// Anthropic API environment variables.
const (
	// EnvAnthropicBaseURL overrides the Anthropic API base URL.
	EnvAnthropicBaseURL = "ANTHROPIC_BASE_URL"

	// EnvAnthropicAuthToken provides the Anthropic API authentication token.
	EnvAnthropicAuthToken = "ANTHROPIC_AUTH_TOKEN"

	// EnvAnthropicAPIKey provides the Anthropic API key. It is the API-key
	// counterpart of EnvAnthropicAuthToken: both authenticate against the same
	// API, but they are distinct variables with distinct client precedence.
	EnvAnthropicAPIKey = "ANTHROPIC_API_KEY"

	// EnvAnthropicDefaultHaikuModel overrides the default Haiku model ID.
	EnvAnthropicDefaultHaikuModel = "ANTHROPIC_DEFAULT_HAIKU_MODEL"

	// EnvAnthropicDefaultFableModel overrides the default Fable model ID.
	EnvAnthropicDefaultFableModel = "ANTHROPIC_DEFAULT_FABLE_MODEL"

	// EnvAnthropicDefaultSonnetModel overrides the default Sonnet model ID.
	EnvAnthropicDefaultSonnetModel = "ANTHROPIC_DEFAULT_SONNET_MODEL"

	// EnvAnthropicDefaultOpusModel overrides the default Opus model ID.
	EnvAnthropicDefaultOpusModel = "ANTHROPIC_DEFAULT_OPUS_MODEL"

	// EnvAnthropicReasoningEffort carries the GLM effort-overlay's session-global
	// reasoning-control value (SPEC-MODEL-TIER-PLANTYPE-001 M5, REQ-MTP-030
	// Branch-B explicit write). It mirrors the ANTHROPIC_DEFAULT_* injection
	// namespace at the GLM launch path. UNVERIFIED delivery: whether z.ai consumes
	// this env var through the Anthropic-compat shim (Branch A passthrough) or
	// requires the reasoning_effort field in the request body (making this env
	// inert) is a run-phase empirical determination (AC-MTP-032b).
	EnvAnthropicReasoningEffort = "ANTHROPIC_REASONING_EFFORT"

	// EnvAnthropicPrefix is the namespace prefix shared by every ANTHROPIC_*
	// environment variable. It is a prefix, NOT a variable name: it is intended
	// for prefix matching (strings.HasPrefix) when filtering an environment map
	// down to the Anthropic namespace, and must never be read with os.Getenv.
	EnvAnthropicPrefix = "ANTHROPIC_"
)

// GitHub API environment variables.
const (
	// EnvGitHubToken provides a GitHub API token for authenticated requests
	// (the update checker's release lookups), lifting the shared-IP 60 req/h
	// anonymous budget to 5,000 req/h. Standard GitHub Actions / gh-tooling
	// variable — MoAI reads it but never sets it.
	EnvGitHubToken = "GITHUB_TOKEN"

	// EnvGHToken is the GitHub CLI's equivalent of GITHUB_TOKEN and takes
	// precedence over it, matching gh's own resolution order.
	EnvGHToken = "GH_TOKEN"
)
