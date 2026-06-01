package config

import (
	"slices"
	"time"

	"github.com/modu-ai/moai-adk/pkg/models"
)

// Config is the root configuration aggregate containing all sections.
// It imports types from pkg/models for shared types (UserConfig, LanguageConfig,
// QualityConfig, ProjectConfig) and defines internal types for the rest.
// @MX:WARN: [AUTO] Large mutable struct with 20+ fields — global configuration state
// @MX:REASON: Any field change can affect entire system; concurrent access requires synchronization
type Config struct {
	User          models.UserConfig          `yaml:"user"`
	Language      models.LanguageConfig      `yaml:"language"`
	Quality       models.QualityConfig       `yaml:"quality"`
	Project       models.ProjectConfig       `yaml:"project"`
	GitStrategy   GitStrategyConfig          `yaml:"git_strategy"`
	GitConvention models.GitConventionConfig `yaml:"git_convention"`
	System        SystemConfig               `yaml:"system"`
	LLM           LLMConfig                  `yaml:"llm"`
	Pricing       PricingConfig              `yaml:"pricing"`
	Ralph         RalphConfig                `yaml:"ralph"`
	Workflow      WorkflowConfig             `yaml:"workflow"`
	State         StateConfig                `yaml:"state"`
	Statusline    models.StatuslineConfig    `yaml:"statusline"`
	Gate          GateConfig                 `yaml:"gate"`
	Sunset        SunsetConfig               `yaml:"sunset"`
	Research      ResearchConfig             `yaml:"research"`
	Session       SessionConfig              `yaml:"session"` // SPEC-V3R2-RT-004 REQ-022: STALE_SECONDS
	// MIG-003: 4 new sections loaded by Loader.Load() (REQ-MIG003-001)
	Constitution  ConstitutionConfig `yaml:"constitution"`
	ContextSearch ContextConfig      `yaml:"context_search"`
	Interview     InterviewConfig    `yaml:"interview"`
	Design        DesignConfig       `yaml:"design"`
}

// GitLabConfig holds the nested gitlab.* settings of the git strategy section.
//
// Per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 REQ-GSS-001: the top-level
// gitlab.instance_url key migrates here from the deprecated FLAT
// GitStrategyConfig.GitLabInstanceURL field (Option (c) accessor-compat).
type GitLabConfig struct {
	InstanceURL string `yaml:"instance_url"`
}

// AutomationConfig holds the nested {mode}.automation.* settings.
//
// Per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 REQ-GSS-002 (forward-compat scaffold):
// these keys are currently consumed by skill bodies via direct yaml reads, not
// by Go code. The struct fields preserve yaml↔struct symmetry for future Go
// consumers without altering the existing skill-body workflow.
type AutomationConfig struct {
	AutoBranch bool `yaml:"auto_branch"`
	AutoCommit bool `yaml:"auto_commit"`
	AutoPR     bool `yaml:"auto_pr"`
	AutoPush   bool `yaml:"auto_push"`
}

// BranchCreationConfig holds the nested {mode}.branch_creation.* settings
// (the Late-Branch opt-in default pair).
//
// Per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 REQ-GSS-002 (forward-compat scaffold).
type BranchCreationConfig struct {
	// AutoEnabled controls whether /moai plan creates a feat/SPEC-* branch.
	// Forward-compat scaffold per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001.
	// Currently consumed by .claude/skills/moai/workflows/plan/spec-assembly.md
	// Phase 3 (yaml-direct read), NOT by Go code. Future Go consumer may migrate
	// to read this struct field via Config.GitStrategy.ActiveModeProfile().
	// See SPEC-V3R5-LATE-BRANCH-001 REQ-LB-004.
	AutoEnabled bool `yaml:"auto_enabled"`
	// PromptAlways controls whether the workflow always prompts before creating
	// a branch. Forward-compat scaffold; see AutoEnabled.
	PromptAlways bool `yaml:"prompt_always"`
}

// CommitStyleConfig holds the nested {mode}.commit_style.* settings.
//
// Per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 REQ-GSS-002 (forward-compat scaffold).
type CommitStyleConfig struct {
	Format        string `yaml:"format"`
	ScopeRequired bool   `yaml:"scope_required"`
}

// HooksConfig holds the nested {mode}.hooks.* settings.
//
// Per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 REQ-GSS-002 (forward-compat scaffold).
type HooksConfig struct {
	PreCommit string `yaml:"pre_commit"`
	PrePush   string `yaml:"pre_push"`
	CommitMsg string `yaml:"commit_msg"`
}

// ModeProfile holds the per-mode (manual/personal/team) settings of the git
// strategy section.
//
// Per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 REQ-GSS-002: scalar fields plus four
// sub-structs (Automation, BranchCreation, CommitStyle, Hooks) plus
// mode-conditional optional fields (AutoCheckpoint is manual-only; BranchPrefix
// and MainBranch are personal/team-only; DraftPR / RequiredReviews /
// BranchProtection are team-only). Missing keys unmarshal to Go zero values.
type ModeProfile struct {
	Workflow          string `yaml:"workflow"`
	Environment       string `yaml:"environment"`
	GitHubIntegration bool   `yaml:"github_integration"`
	PushToRemote      bool   `yaml:"push_to_remote"`

	// Mode-conditional optional fields (zero value when the mode lacks the key).
	AutoCheckpoint   string `yaml:"auto_checkpoint"`   // manual mode only
	BranchPrefix     string `yaml:"branch_prefix"`     // personal/team modes only
	MainBranch       string `yaml:"main_branch"`       // personal/team modes only
	DraftPR          bool   `yaml:"draft_pr"`          // team mode only
	RequiredReviews  int    `yaml:"required_reviews"`  // team mode only
	BranchProtection bool   `yaml:"branch_protection"` // team mode only

	// Nested sub-structs.
	Automation     AutomationConfig     `yaml:"automation"`
	BranchCreation BranchCreationConfig `yaml:"branch_creation"`
	CommitStyle    CommitStyleConfig    `yaml:"commit_style"`
	Hooks          HooksConfig          `yaml:"hooks"`
}

// GitStrategyConfig represents the git strategy configuration section.
//
// Schema reorganized per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 (Option (c)):
//   - Top-level fields (Mode, Provider, GitHubUsername, GitLab) are wire-through.
//   - Mode profiles (Manual, Personal, Team) are forward-compat scaffolds for
//     future Go consumers of Late-Branch keys (see SPEC-V3R5-LATE-BRANCH-001).
//   - FLAT fields (AutoBranch, BranchPrefix, CommitStyle, WorktreeRoot,
//     GitLabInstanceURL) are deprecated per backward-compat Option (c) but
//     retained per SPEC-CONFIG-001 historical contract.
type GitStrategyConfig struct {
	// Top-level wire-through fields.
	Mode           string       `yaml:"mode"`     // "manual", "personal", "team"
	Provider       string       `yaml:"provider"` // "github", "gitlab"
	GitHubUsername string       `yaml:"github_username"`
	GitLab         GitLabConfig `yaml:"gitlab"`

	// Mode profile forward-compat scaffolds.
	Manual   ModeProfile `yaml:"manual"`
	Personal ModeProfile `yaml:"personal"`
	Team     ModeProfile `yaml:"team"`

	// Deprecated FLAT fields — preserved for backward-compat per Option (c).
	// A future SPEC may sunset these with a SemVer major-bump.
	// Deprecated: use ActiveModeProfile().Automation.AutoBranch instead.
	AutoBranch bool `yaml:"auto_branch"`
	// Deprecated: use ActiveModeProfile().BranchPrefix instead.
	BranchPrefix string `yaml:"branch_prefix"`
	// Deprecated: use ActiveModeProfile().CommitStyle.Format instead.
	CommitStyle string `yaml:"commit_style"`
	// Deprecated: no production consumer; preserved per SPEC-CONFIG-001.
	WorktreeRoot string `yaml:"worktree_root"`
	// Deprecated: use GitLab.InstanceURL instead.
	GitLabInstanceURL string `yaml:"gitlab_instance_url"`
}

// ActiveModeProfile returns a pointer to the currently selected ModeProfile
// based on the Mode field. Returns (nil, false) when Mode is empty or invalid.
//
// Per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 REQ-GSS-004. The accessor does not
// modify any field; callers MUST check the boolean before dereferencing.
func (c *GitStrategyConfig) ActiveModeProfile() (*ModeProfile, bool) {
	switch c.Mode {
	case "manual":
		return &c.Manual, true
	case "personal":
		return &c.Personal, true
	case "team":
		return &c.Team, true
	default:
		return nil, false
	}
}

// SystemHookConfig holds hook observability settings (SPEC-V3R2-RT-006 REQ-004).
// It controls which retired events are re-enabled as observability taps and
// whether strict mode behavior applies to retired events.
//
// COHABITATION NOTE (SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 §A.3): the OptIn field
// is an INDEPENDENT master toggle for 3 hook series (TaskCreated, Notification,
// handle-harness-observe-*). ObservabilityEvents (SPEC-V3R2-RT-006 REQ-040
// per-event whitelist) and Observability.Enabled (REQ-OBS-005 trace-logging
// master, separate file) are NOT collapsed with OptIn — do NOT unify without
// a fresh SPEC. See internal/hook/observability.go file-top note.
type SystemHookConfig struct {
	// ObservabilityEvents is the list of retired event names that are re-enabled
	// as observability taps. Empty list (default) means silent no-op for all.
	// Valid names: notification, elicitation, elicitationResult, taskCreated.
	ObservabilityEvents []string `yaml:"observability_events" validate:"omitempty"`
	// StrictMode: when true, retired events in strict mode still succeed silently.
	StrictMode bool `yaml:"strict_mode"`
	// OptIn is the master toggle for SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 hook series
	// (TaskCreated, Notification, handle-harness-observe-*). When absent in YAML,
	// Go zero-value (false) is used; this matches plan.md R3 mitigation.
	OptIn HookOptInConfig `yaml:"opt_in"`
}

// HookOptInConfig holds the master opt-in toggle for SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001.
// Currently a single boolean; designed as a sub-struct so future fields (per-series
// granularity, scheduling, etc.) can be added without breaking existing YAML.
type HookOptInConfig struct {
	// Enabled is the master toggle for 3 observability hook series. Default false.
	Enabled bool `yaml:"enabled"`
}

// SystemConfig represents the system configuration section.
type SystemConfig struct {
	Version        string           `yaml:"version"`
	LogLevel       string           `yaml:"log_level"`
	LogFormat      string           `yaml:"log_format"`
	NoColor        bool             `yaml:"no_color"`
	NonInteractive bool             `yaml:"non_interactive"`
	Migrations     MigrationsConfig `yaml:"migrations"`
	Hook           SystemHookConfig `yaml:"hook"`
}

// MigrationsConfig represents the migrations configuration section.
// REQ-V3R2-RT-007-032: session-start migration can be disabled via migrations.disabled.
type MigrationsConfig struct {
	Disabled bool `yaml:"disabled"`
}

// LLMConfig represents the LLM configuration section.
type LLMConfig struct {
	// Mode selection: "", "glm"
	Mode string `yaml:"mode"`
	// TeamMode selection: "", "claude", "glm", "hybrid"
	TeamMode string `yaml:"team_mode"`
	// Environment variable name for GLM API key
	GLMEnvVar string `yaml:"glm_env_var"`
	// Performance tier: "high", "medium", "low"
	// Controls model selection for all sub-agents and team agents
	PerformanceTier string `yaml:"performance_tier" validate:"omitempty,oneof=high medium low"`
	// Claude model mapping by tier
	ClaudeModels ClaudeTierModels `yaml:"claude_models"`
	// GLM API configuration
	GLM GLMSettings `yaml:"glm"`
	// Legacy fields (kept for backward compatibility, mapped from tiers)
	DefaultModel string `yaml:"default_model"`
	QualityModel string `yaml:"quality_model"`
	SpeedModel   string `yaml:"speed_model"`
}

// ClaudeTierModels represents Claude model mappings by performance tier.
type ClaudeTierModels struct {
	High   string `yaml:"high"`   // Complex reasoning, architecture, security
	Medium string `yaml:"medium"` // Balanced performance for most tasks
	Low    string `yaml:"low"`    // Fast exploration, simple tasks
}

// GLMSettings represents GLM API configuration.
type GLMSettings struct {
	BaseURL string    `yaml:"base_url"`
	Models  GLMModels `yaml:"models"`
	// ContextWindows maps GLM model names to their actual context window sizes
	// (in tokens), overriding the built-in statusline table. Issue #653.
	// Example entries: {"glm-5.1": 230000, "glm-4.5-air": 128000}.
	// Takes precedence over the built-in glmContextWindows table in
	// internal/statusline/memory.go but yields to MOAI_STATUSLINE_CONTEXT_SIZE.
	ContextWindows map[string]int `yaml:"context_windows,omitempty"`
}

// GLMModels represents GLM model mappings by performance tier.
type GLMModels struct {
	High   string `yaml:"high"`   // Complex reasoning
	Medium string `yaml:"medium"` // Balanced performance
	Low    string `yaml:"low"`    // Fast exploration
	// Legacy fields for backward compatibility
	Opus   string `yaml:"opus"`   // Maps to High
	Sonnet string `yaml:"sonnet"` // Maps to Medium
	Haiku  string `yaml:"haiku"`  // Maps to Low
}

// PricingConfig represents the pricing configuration section.
type PricingConfig struct {
	TokenBudget  int  `yaml:"token_budget"`
	CostTracking bool `yaml:"cost_tracking"`
}

// RalphConfig represents the Ralph engine configuration section.
type RalphConfig struct {
	MaxIterations int  `yaml:"max_iterations"`
	AutoConverge  bool `yaml:"auto_converge"`
	HumanReview   bool `yaml:"human_review"`

	// LintAsInstruction enables injecting LSP diagnostics as systemMessage
	// so the AI receives errors as its next prompt (REQ-LAI-003).
	// Default: true.
	LintAsInstruction bool `yaml:"lint_as_instruction"`

	// WarnAsInstruction includes warnings in the systemMessage when there are
	// no errors and this flag is true (REQ-LAI-006). Default: false.
	WarnAsInstruction bool `yaml:"warn_as_instruction"`
}

// WorkflowConfig represents the workflow configuration section.
type WorkflowConfig struct {
	AutoClear     bool                    `yaml:"auto_clear"`
	PlanTokens    int                     `yaml:"plan_tokens"`
	RunTokens     int                     `yaml:"run_tokens"`
	SyncTokens    int                     `yaml:"sync_tokens"`
	AutoSelection TeamAutoSelectionConfig `yaml:"auto_selection"`
}

// RoleProfile represents an agent role profile configuration.
// It extends the base role profile from workflow.yaml with sandbox settings.
// @MX:SPEC: SPEC-V3R2-RT-003 REQ-003
type RoleProfile struct {
	// Sandbox is the default sandbox backend for this role.
	// Values: "none", "bubblewrap", "seatbelt", "docker".
	// Defaults: implementer/tester/designer → OS-resolved (seatbelt|bubblewrap);
	//           researcher/analyst/reviewer/architect → "none".
	Sandbox string `yaml:"sandbox"`
}

// SecuritySandbox holds sandbox-specific security configuration.
// Extended from security.yaml sandbox.* keys per REQ-V3R2-RT-003-008/030.
//
// @MX:ANCHOR: [AUTO] SecuritySandbox is the config schema for all sandbox knobs
// @MX:REASON: Fan_in >= 3: loaded by config/loader.go, consumed by sandbox/launcher.go,
//             displayed by doctor_sandbox.go, tested by config/types_test.go
// @MX:SPEC: SPEC-V3R2-RT-003 REQ-008/020/030/031
type SecuritySandbox struct {
	// Required: when true, agents with sandbox: none fail to spawn unless they provide
	// sandbox.justification frontmatter. Default: false.
	Required bool `yaml:"required"`

	// NetworkAllowlist lists additional allowed outbound hosts, appended to the built-in
	// default 8-host list in sandbox.DefaultNetworkAllowlist.
	NetworkAllowlist []string `yaml:"network_allowlist"`

	// EnvScrubExtra lists additional environment variable names to scrub beyond the
	// built-in denylist. These are additive (never replace the built-in list).
	EnvScrubExtra []string `yaml:"env_scrub_extra"`

	// DockerImage is the default Docker image for the docker backend.
	// Default: "alpine:latest" (production image pending SPEC-V3R2-EXT-004).
	DockerImage string `yaml:"docker_image"`
}

// TeamAutoSelectionConfig holds thresholds for automatic team vs solo mode selection.
// These values are evaluated by the orchestrator to determine execution mode
// when no explicit --team or --solo flag is provided.
type TeamAutoSelectionConfig struct {
	// MinDomainsForTeam is the minimum number of distinct domains to trigger team mode.
	MinDomainsForTeam int `yaml:"min_domains_for_team"`
	// MinFilesForTeam is the minimum number of affected files to trigger team mode.
	MinFilesForTeam int `yaml:"min_files_for_team"`
	// MinComplexityScore is the minimum complexity score (1-10) to trigger team mode.
	MinComplexityScore int `yaml:"min_complexity_score"`
}

// StateConfig represents the project state storage configuration.
// It controls the directory where structured state data (checkpoints,
// coverage, diagnostics) is stored.
type StateConfig struct {
	StateDir      string `yaml:"state_dir"`
	RetentionDays int    `yaml:"retention_days"` // SPEC-V3R2-RT-004 REQ-031: retention days for the runs/ directory
}

// SessionConfig holds session state management configuration.
// SPEC-V3R2-RT-004 REQ-022: STALE_SECONDS setting.
type SessionConfig struct {
	// StaleSeconds is the threshold (in seconds) at which a checkpoint is considered stale.
	// Default: 3600 (1 hour). Configured via the stale_seconds key in ralph.yaml.
	StaleSeconds int `yaml:"stale_seconds"`
}

// LSPQualityGates represents LSP quality gate configuration.
type LSPQualityGates struct {
	Enabled         bool     `yaml:"enabled"`
	Plan            PlanGate `yaml:"plan"`
	Run             RunGate  `yaml:"run"`
	Sync            SyncGate `yaml:"sync"`
	CacheTTLSeconds int      `yaml:"cache_ttl_seconds"`
	TimeoutSeconds  int      `yaml:"timeout_seconds"`
}

// PlanGate represents the plan phase quality gate.
type PlanGate struct {
	RequireBaseline bool `yaml:"require_baseline"`
}

// RunGate represents the run phase quality gate.
type RunGate struct {
	MaxErrors       int  `yaml:"max_errors"`
	MaxTypeErrors   int  `yaml:"max_type_errors"`
	MaxLintErrors   int  `yaml:"max_lint_errors"`
	AllowRegression bool `yaml:"allow_regression"`
}

// SyncGate represents the sync phase quality gate.
type SyncGate struct {
	MaxErrors       int  `yaml:"max_errors"`
	MaxWarnings     int  `yaml:"max_warnings"`
	RequireCleanLSP bool `yaml:"require_clean_lsp"`
}

// GateConfig represents configuration for the deterministic quality gate
// that runs before git commit (SPEC-GATE-001).
type GateConfig struct {
	// Enabled controls whether the quality gate runs.
	Enabled bool `yaml:"enabled"`
	// SkipTests skips the go test step when true.
	SkipTests bool `yaml:"skip_tests"`
	// Timeouts holds per-step timeout values in seconds.
	Timeouts GateTimeouts `yaml:"timeouts"`
	// AstGrepGate configures the ast-grep domain rule scan step (SPEC-SLQG-001).
	AstGrepGate AstGrepGateConfig `yaml:"ast_grep_gate"`
}

// AstGrepGateConfig holds configuration for ast-grep quality gate scanning.
type AstGrepGateConfig struct {
	// Enabled controls whether ast-grep scanning is performed.
	Enabled bool `yaml:"enabled"`
	// RulesDir is the directory containing domain-specific ast-grep rule files.
	RulesDir string `yaml:"rules_dir"`
	// BlockOnError causes the gate to block a commit when error-severity matches are found.
	BlockOnError bool `yaml:"block_on_error"`
	// WarnOnlyMode prevents blocking even when error-severity matches are found.
	WarnOnlyMode bool `yaml:"warn_only_mode"`
}

// GateTimeouts holds per-step timeout configuration in seconds.
type GateTimeouts struct {
	Vet  int `yaml:"vet"`
	Lint int `yaml:"lint"`
	Test int `yaml:"test"`
}

// VetTimeoutDuration converts the Vet timeout to time.Duration.
// Returns 30s when the value is zero or negative.
func (g *GateConfig) VetTimeoutDuration() time.Duration {
	if g.Timeouts.Vet <= 0 {
		return 30 * time.Second
	}
	return time.Duration(g.Timeouts.Vet) * time.Second
}

// LintTimeoutDuration converts the Lint timeout to time.Duration.
// Returns 60s when the value is zero or negative.
func (g *GateConfig) LintTimeoutDuration() time.Duration {
	if g.Timeouts.Lint <= 0 {
		return 60 * time.Second
	}
	return time.Duration(g.Timeouts.Lint) * time.Second
}

// TestTimeoutDuration converts the Test timeout to time.Duration.
// Returns 120s when the value is zero or negative.
func (g *GateConfig) TestTimeoutDuration() time.Duration {
	if g.Timeouts.Test <= 0 {
		return 120 * time.Second
	}
	return time.Duration(g.Timeouts.Test) * time.Second
}

// SunsetConfig defines the Build-to-Delete framework configuration.
// Quality gates that consistently pass can be relaxed over time.
//
// @MX:NOTE: [AUTO] DORMANT — struct schema defined but no runtime hot path
// enforces sunset conditions. Activation deferred to a future SPEC.
// Do NOT add LoadSunsetConfig until activation SPEC is filed.
// REQ-MIG003-006 ↔ REQ-MIG003-015 reference.
type SunsetConfig struct {
	// Enabled controls whether sunset tracking is active.
	Enabled    bool              `yaml:"enabled"`
	Conditions []SunsetCondition `yaml:"conditions"`
}

// SunsetCondition defines when a quality gate can be relaxed.
type SunsetCondition struct {
	Gate        string `yaml:"gate"`
	Metric      string `yaml:"metric"`
	Threshold   int    `yaml:"threshold"`
	Action      string `yaml:"action"`
	Description string `yaml:"description"`
}

// ResearchConfig represents the Self-Research System configuration section.
type ResearchConfig struct {
	Enabled   bool                    `yaml:"enabled"`
	Passive   ResearchPassiveConfig   `yaml:"passive"`
	Active    ResearchActiveConfig    `yaml:"active"`
	Safety    ResearchSafetyConfig    `yaml:"safety"`
	Dashboard ResearchDashboardConfig `yaml:"dashboard"`
}

// ResearchPassiveConfig represents passive observation settings.
type ResearchPassiveConfig struct {
	Enabled                 bool                      `yaml:"enabled"`
	CorrectionWindowSeconds int                       `yaml:"correction_window_seconds"`
	PatternThresholds       ResearchPatternThresholds `yaml:"pattern_thresholds"`
}

// ResearchPatternThresholds defines observation count thresholds for pattern classification.
type ResearchPatternThresholds struct {
	Heuristic      int `yaml:"heuristic"`
	Rule           int `yaml:"rule"`
	HighConfidence int `yaml:"high_confidence"`
}

// ResearchActiveConfig represents active experiment settings.
type ResearchActiveConfig struct {
	RunsPerExperiment int     `yaml:"runs_per_experiment"`
	MaxExperiments    int     `yaml:"max_experiments"`
	PassThreshold     float64 `yaml:"pass_threshold"`
	TargetScore       float64 `yaml:"target_score"`
	BudgetCapTokens   int     `yaml:"budget_cap_tokens"`
}

// ResearchSafetyConfig represents safety layer settings.
type ResearchSafetyConfig struct {
	WorktreeIsolation         bool                    `yaml:"worktree_isolation"`
	CanaryRegressionThreshold float64                 `yaml:"canary_regression_threshold"`
	RateLimits                ResearchRateLimitConfig `yaml:"rate_limits"`
}

// ResearchRateLimitConfig represents rate limiting settings.
type ResearchRateLimitConfig struct {
	MaxExperimentsPerSession int `yaml:"max_experiments_per_session"`
	MaxAcceptedPerSession    int `yaml:"max_accepted_per_session"`
	MaxAutoResearchPerWeek   int `yaml:"max_auto_research_per_week"`
}

// ResearchDashboardConfig represents dashboard display settings.
type ResearchDashboardConfig struct {
	DefaultMode     string `yaml:"default_mode"`
	HTMLOpenBrowser bool   `yaml:"html_open_browser"`
}

// sectionNames lists all valid configuration section names.
var sectionNames = []string{
	"user", "language", "quality", "project",
	"git_strategy", "git_convention", "system", "llm",
	"pricing", "ralph", "workflow", "state", "statusline", "gate", "sunset",
	"research",
}

// IsValidSectionName checks if the given name is a valid section name.
func IsValidSectionName(name string) bool {
	return slices.Contains(sectionNames, name)
}

// ValidSectionNames returns all valid section names.
func ValidSectionNames() []string {
	result := make([]string, len(sectionNames))
	copy(result, sectionNames)
	return result
}

// HarnessConfig is the top-level configuration struct for harness.yaml.
// @MX:ANCHOR: [AUTO] Go representation of the full harness.yaml schema — return type of LoadHarnessConfig()
// @MX:REASON: fan_in >= 3 (consumed by LoadHarnessConfig, router.Route, CLI validate, and others)
// REQ-HRN-001-001: struct covering the full harness.yaml schema (HRN-001 run-phase extension).
type HarnessConfig struct {
	// DefaultProfile is the default evaluator profile name.
	DefaultProfile string `yaml:"default_profile"`
	// ModeDefaults is the default harness level map per execution mode (solo/team/cg).
	// REQ-HRN-001-014: mode_defaults.cg = "thorough" (FROZEN).
	ModeDefaults map[string]string `yaml:"mode_defaults,omitempty"`
	// AutoDetection holds auto-detection rule settings.
	// REQ-HRN-001-007: priority order is minimal → standard → thorough.
	AutoDetection AutoDetectionConfig `yaml:"auto_detection,omitempty"`
	// Escalation holds escalation trigger settings.
	// REQ-HRN-001-004/009/013: max_escalations + triggers.
	Escalation EscalationConfig `yaml:"escalation,omitempty"`
	// EffortMapping is the level → effort tier map.
	// REQ-HRN-001-005: minimal→medium, standard→high, thorough→xhigh.
	EffortMapping map[string]string `yaml:"effort_mapping,omitempty"`
	// Levels is the per-level configuration map.
	// REQ-HRN-001-001: {minimal, standard, thorough} FROZEN enum.
	Levels map[string]LevelConfig `yaml:"levels,omitempty"`
	// ModelUpgradeReview holds model upgrade review settings.
	// REQ-HRN-001-016.
	ModelUpgradeReview ModelUpgradeReviewConfig `yaml:"model_upgrade_review,omitempty"`
	// PlanAuditGlobal holds the global plan audit settings.
	PlanAuditGlobal PlanAuditGlobalConfig `yaml:"plan_audit_global,omitempty"`
	// Evaluator is the HRN-002 substrate — used for memory_scope FROZEN validation.
	Evaluator EvaluatorConfig `yaml:"evaluator"`
}

// AutoDetectionConfig is the configuration struct for the auto_detection block.
// REQ-HRN-001-007: the rules map priority is minimal → standard → thorough.
type AutoDetectionConfig struct {
	// Enabled toggles auto-detection.
	Enabled bool `yaml:"enabled"`
	// Rules is the per-level detection condition map.
	Rules map[string]AutoDetectionRule `yaml:"rules,omitempty"`
}

// AutoDetectionRule lists auto-detection conditions for a single level.
type AutoDetectionRule struct {
	// Conditions is the list of condition strings required to route to this level.
	Conditions []string `yaml:"conditions,omitempty"`
}

// EscalationConfig is the configuration struct for the escalation block.
// REQ-HRN-001-004/009/013: max_escalations ceiling + trigger list.
type EscalationConfig struct {
	// Enabled toggles escalation.
	Enabled bool `yaml:"enabled"`
	// MaxEscalations is the maximum escalation count per stage (default 2, ceiling 3).
	// REQ-HRN-001-013: hard ceiling = 3.
	MaxEscalations int `yaml:"max_escalations"`
	// Triggers lists events that trigger escalation.
	// (e.g., quality_gate_fail, review_critical, test_coverage_low)
	Triggers []string `yaml:"triggers,omitempty"`
}

// LevelConfig is the configuration struct for a single harness level.
// REQ-HRN-001-001: settings for each of levels.{minimal,standard,thorough}.
type LevelConfig struct {
	// Description describes this level.
	Description string `yaml:"description,omitempty"`
	// Evaluator toggles the evaluator.
	Evaluator bool `yaml:"evaluator"`
	// EvaluatorMode is the evaluator mode (final-pass or per-sprint).
	EvaluatorMode string `yaml:"evaluator_mode,omitempty"`
	// EvaluatorProfile is the name of the evaluator profile to use.
	// When set, it resolves to the .moai/config/evaluator-profiles/{name}.md file.
	EvaluatorProfile string `yaml:"evaluator_profile,omitempty"`
	// SprintContract toggles sprint contract behavior.
	SprintContract bool `yaml:"sprint_contract"`
	// PlaywrightTesting toggles Playwright testing.
	PlaywrightTesting bool `yaml:"playwright_testing"`
	// SkipPhases lists workflow phases to skip.
	SkipPhases []any `yaml:"skip_phases,omitempty"`
	// PlanAudit holds the plan audit settings.
	PlanAudit PlanAuditConfig `yaml:"plan_audit,omitempty"`
}

// PlanAuditConfig is the configuration struct for plan audit settings.
type PlanAuditConfig struct {
	// Enabled toggles plan audit.
	Enabled bool `yaml:"enabled"`
	// MaxIterations is the maximum iteration count.
	MaxIterations int `yaml:"max_iterations"`
	// RequireMustPass toggles must-pass enforcement.
	RequireMustPass bool `yaml:"require_must_pass"`
	// CrossValidateWithEvaluatorActive toggles cross-validation with sync-auditor.
	CrossValidateWithEvaluatorActive bool `yaml:"cross_validate_with_evaluator_active"`
}

// ModelUpgradeReviewConfig is the configuration struct for the model_upgrade_review block.
// REQ-HRN-001-016: checklist notification on model upgrade.
type ModelUpgradeReviewConfig struct {
	// Enabled toggles model upgrade review.
	Enabled bool `yaml:"enabled"`
	// Checklist is the list of review items.
	Checklist []ReviewChecklistItem `yaml:"checklist,omitempty"`
	// Trigger holds the review trigger settings.
	Trigger ModelUpgradeTrigger `yaml:"trigger,omitempty"`
	// Output holds the review output settings.
	Output ModelUpgradeOutput `yaml:"output,omitempty"`
}

// ReviewChecklistItem is a single item in the model upgrade review.
type ReviewChecklistItem struct {
	ID       string `yaml:"id"`
	Question string `yaml:"question"`
	Action   string `yaml:"action"`
	Affects  string `yaml:"affects"`
}

// ModelUpgradeTrigger holds the model upgrade review trigger settings.
type ModelUpgradeTrigger struct {
	OnModelChange     bool   `yaml:"on_model_change"`
	ManualCommand     string `yaml:"manual_command,omitempty"`
	ReviewIntervalDays int   `yaml:"review_interval_days"`
}

// ModelUpgradeOutput holds the model upgrade review output settings.
type ModelUpgradeOutput struct {
	ReportPath      string `yaml:"report_path,omitempty"`
	RequireApproval bool   `yaml:"require_approval"`
}

// PlanAuditGlobalConfig is the configuration struct for the plan_audit_global block.
type PlanAuditGlobalConfig struct {
	// AlwaysEnabled toggles permanent plan audit activation.
	AlwaysEnabled bool `yaml:"always_enabled"`
	// EnforceGateOnSpecCreation toggles gate enforcement at SPEC creation time.
	EnforceGateOnSpecCreation bool `yaml:"enforce_gate_on_spec_creation"`
	// Rationale describes the reason for these settings.
	Rationale string `yaml:"rationale,omitempty"`
}

// EvaluatorConfig is the sub-configuration struct for the evaluator.
// @MX:NOTE: FROZEN at per_iteration per design-constitution §11.4.1 (SPEC-V3R2-HRN-002)
// @MX:NOTE: [AUTO] HRN-003 M4: added Profiles + Aggregation + MustPassDimensions fields (SPEC-V3R2-HRN-003)
type EvaluatorConfig struct {
	// MemoryScope is the evaluator memory scope setting.
	// FROZEN to the value per_iteration by design-constitution §11.4.1.
	// Other values (e.g., cumulative) return an HRN_EVAL_MEMORY_FROZEN error.
	MemoryScope string `yaml:"memory_scope"`
	// Profiles is the map from evaluator profile name to .md file path.
	// REQ-HRN-003-005, AC-HRN-003-07.c.
	Profiles map[string]string `yaml:"profiles,omitempty"`
	// Aggregation is the default aggregation method ("min" or "mean").
	// REQ-HRN-003-007: default is "min".
	Aggregation string `yaml:"aggregation,omitempty"`
	// MustPassDimensions lists dimension names treated as must-pass.
	// REQ-HRN-003-018: default is [Functionality, Security].
	MustPassDimensions []string `yaml:"must_pass_dimensions,omitempty"`
}

// ConstitutionConfig is the top-level configuration struct for constitution.yaml.
// @MX:ANCHOR: [AUTO] Go representation of the full constitution.yaml schema
// @MX:REASON: fan_in >= 3 (LoadConstitutionConfig, Loader.Load, SPEC-V3R2-EXT-004 hook consumer)
//
// Hot path: SPEC-V3R2-EXT-004 framework optional hook for forbidden-library policy enforcement.
// ForbiddenPatterns is exposed as the "ForbiddenLibraries" list per REQ-MIG003-009.
type ConstitutionConfig struct {
	ApprovedFrameworks []string                   `yaml:"approved_frameworks"`
	ApprovedLanguages  []string                   `yaml:"approved_languages"`
	Architecture       ConstitutionArchitecture   `yaml:"architecture"`
	ForbiddenPatterns  []string                   `yaml:"forbidden_patterns"`
	NamingConventions  ConstitutionNaming         `yaml:"naming_conventions"`
	Performance        ConstitutionPerformance    `yaml:"performance"`
	Security           ConstitutionSecurity       `yaml:"security"`
}

// ConstitutionArchitecture holds architecture constraint fields.
type ConstitutionArchitecture struct {
	ForbiddenDependencies []string `yaml:"forbidden_dependencies"`
	Patterns              []string `yaml:"patterns"`
}

// ConstitutionNaming holds naming convention fields.
type ConstitutionNaming struct {
	Exported string `yaml:"exported"`
	Files    string `yaml:"files"`
	Packages string `yaml:"packages"`
}

// ConstitutionPerformance holds performance threshold fields.
type ConstitutionPerformance struct {
	MaxMemoryMB        *int `yaml:"max_memory_mb"`
	MaxResponseTimeMS  *int `yaml:"max_response_time_ms"`
}

// ConstitutionSecurity holds security policy fields.
type ConstitutionSecurity struct {
	ForbiddenPractices []string `yaml:"forbidden_practices"`
	RequiredChecks     []string `yaml:"required_checks"`
}

// ContextConfig is the top-level configuration struct for context.yaml (context_search:).
// @MX:ANCHOR: [AUTO] Go representation of the context_search section
// @MX:REASON: fan_in >= 3 (LoadContextConfig, Loader.Load, CLAUDE.md §16 Context Search consumer)
//
// Hot path: CLAUDE.md §16 Context Search Protocol — token_budget.max_injection_tokens
// and token_budget.skip_if_usage_above gate context injection.
// search.date_range_days is the "staleness_window_days" alias per REQ-MIG003-010.
type ContextConfig struct {
	AutoDetect        ContextAutoDetect        `yaml:"auto_detect"`
	Enabled           bool                     `yaml:"enabled"`
	MemoryIntegration ContextMemoryIntegration `yaml:"memory_integration"`
	Performance       ContextPerformance       `yaml:"performance"`
	Search            ContextSearch            `yaml:"search"`
	TokenBudget       ContextTokenBudget       `yaml:"token_budget"`
}

// ContextAutoDetect holds auto-detection settings.
type ContextAutoDetect struct {
	Enabled bool `yaml:"enabled"`
}

// ContextMemoryIntegration holds memory integration settings.
type ContextMemoryIntegration struct {
	Enabled          bool `yaml:"enabled"`
	IncludeInContext bool `yaml:"include_in_context"`
	PriorityOverSearch bool `yaml:"priority_over_search"`
}

// ContextPerformance holds performance settings.
type ContextPerformance struct {
	CacheTTLSeconds int `yaml:"cache_ttl_seconds"`
	TimeoutSeconds  int `yaml:"timeout_seconds"`
}

// ContextSearch holds search configuration fields.
// DateRangeDays is aliased as "staleness_window_days" in spec.md (REQ-MIG003-010).
type ContextSearch struct {
	DateRangeDays      int  `yaml:"date_range_days"`
	MaxResults         int  `yaml:"max_results"`
	MaxTokensPerResult int  `yaml:"max_tokens_per_result"`
	ProjectScopeOnly   bool `yaml:"project_scope_only"`
}

// ContextTokenBudget holds token budget fields consumed by CLAUDE.md §16.
type ContextTokenBudget struct {
	MaxInjectionTokens int `yaml:"max_injection_tokens"`
	SkipIfUsageAbove   int `yaml:"skip_if_usage_above"`
}

// InterviewConfig is the top-level configuration struct for interview.yaml.
// @MX:ANCHOR: [AUTO] Go representation of the interview section
// @MX:REASON: fan_in >= 3 (LoadInterviewConfig, Loader.Load, SPEC-V3R2-WF-003 discovery mode)
//
// Hot path: SPEC-V3R2-WF-003 discovery mode consumes clarity_threshold, plan.max_rounds,
// plan.questions_per_round, and skip_conditions to control Socratic interview behavior.
type InterviewConfig struct {
	ClarityThreshold int            `yaml:"clarity_threshold"`
	Enabled          bool           `yaml:"enabled"`
	Plan             InterviewMode  `yaml:"plan"`
	Project          InterviewMode  `yaml:"project"`
	SkipConditions   []string       `yaml:"skip_conditions"`
}

// InterviewMode holds per-mode interview settings.
type InterviewMode struct {
	MaxRounds        int `yaml:"max_rounds"`
	QuestionsPerRound int `yaml:"questions_per_round"`
}

// DesignConfig is the top-level configuration struct for design.yaml.
// @MX:ANCHOR: [AUTO] Go representation of the design section
// @MX:REASON: fan_in >= 3 (LoadDesignConfig, Loader.Load, GAN loop runtime sprint contract consumer)
//
// Hot path: GAN loop runtime consumes gan_loop.pass_threshold (FROZEN floor 0.60),
// gan_loop.sprint_contract.enabled, and adaptation.iteration_limits.
// Note: adaptation.iteration_limits is the Go field for spec.md's conceptual "phase_weights" (REQ-MIG003-014 OQ4).
type DesignConfig struct {
	Adaptation      DesignAdaptation   `yaml:"adaptation"`
	BrandContext    DesignBrandContext  `yaml:"brand_context"`
	ClaudeDesign    DesignClaudeDesign  `yaml:"claude_design"`
	DefaultFramework string             `yaml:"default_framework"`
	DesignDocs      DesignDocs         `yaml:"design_docs"`
	Enabled         bool               `yaml:"enabled"`
	Evaluator       DesignEvaluator    `yaml:"evaluator"`
	Evolution       DesignEvolution    `yaml:"evolution"`
	Figma           DesignFigma        `yaml:"figma"`
	GanLoop         DesignGanLoop      `yaml:"gan_loop"`
	Version         string             `yaml:"version"`
}

// DesignAdaptation holds pipeline adaptation settings.
type DesignAdaptation struct {
	ConfidenceThreshold       float64                   `yaml:"confidence_threshold"`
	Enabled                   bool                      `yaml:"enabled"`
	IterationLimits            DesignIterationLimits    `yaml:"iteration_limits"`
	MinProjectsForAdaptation  int                       `yaml:"min_projects_for_adaptation"`
}

// DesignIterationLimits holds per-role iteration limit settings.
type DesignIterationLimits struct {
	Builder    int `yaml:"builder"`
	Copywriter int `yaml:"copywriter"`
	Designer   int `yaml:"designer"`
}

// DesignBrandContext holds brand context directory settings.
type DesignBrandContext struct {
	Dir                string `yaml:"dir"`
	InterviewOnFirstRun bool   `yaml:"interview_on_first_run"`
}

// DesignClaudeDesign holds Claude Design import settings.
type DesignClaudeDesign struct {
	Enabled                 bool     `yaml:"enabled"`
	FallbackPath            string   `yaml:"fallback_path"`
	SupportedBundleVersions []string `yaml:"supported_bundle_versions"`
}

// DesignDocs holds design document auto-loading settings.
type DesignDocs struct {
	AutoLoadOnDesignCommand bool     `yaml:"auto_load_on_design_command"`
	Dir                     string   `yaml:"dir"`
	Priority                []string `yaml:"priority"`
	TokenBudget             int      `yaml:"token_budget"`
}

// DesignEvaluator holds evaluator configuration for design pipeline.
// memory_scope is treated as a free string here (not FROZEN-enforced;
// that enforcement stays in LoadHarnessConfig per plan.md §11.4).
type DesignEvaluator struct {
	MemoryScope string `yaml:"memory_scope"`
}

// DesignEvolution holds evolution and self-learning settings.
type DesignEvolution struct {
	ArchiveAfterEvolve      bool                       `yaml:"archive_after_evolve"`
	AutoEvolveThreshold     int                        `yaml:"auto_evolve_threshold"`
	CooldownHours           int                        `yaml:"cooldown_hours"`
	GraduationCriteria      DesignGraduationCriteria   `yaml:"graduation_criteria"`
	MaxActiveLearnings      int                        `yaml:"max_active_learnings"`
	MaxEvolutionRatePerWeek int                        `yaml:"max_evolution_rate_per_week"`
	RequireApproval         bool                       `yaml:"require_approval"`
}

// DesignGraduationCriteria holds graduation thresholds for learnings.
type DesignGraduationCriteria struct {
	ConsistencyRatio    float64 `yaml:"consistency_ratio"`
	MinimumConfidence   float64 `yaml:"minimum_confidence"`
	MinimumObservations int     `yaml:"minimum_observations"`
	StalenessWindowDays int     `yaml:"staleness_window_days"`
}

// DesignFigma holds Figma integration settings.
type DesignFigma struct {
	Enabled bool `yaml:"enabled"`
}

// DesignGanLoop holds GAN loop configuration.
// PassThreshold is validated against the FROZEN floor 0.60 by LoadDesignConfig.
// REQ-MIG003-014, OQ2 decision: enforce floor in loader.
type DesignGanLoop struct {
	EscalationAfter     int                  `yaml:"escalation_after"`
	ImprovementThreshold float64             `yaml:"improvement_threshold"`
	MaxIterations       int                  `yaml:"max_iterations"`
	PassThreshold       float64              `yaml:"pass_threshold"`
	SprintContract      DesignSprintContract `yaml:"sprint_contract"`
	StrictMode          bool                 `yaml:"strict_mode"`
}

// DesignSprintContract holds sprint contract configuration.
type DesignSprintContract struct {
	ArtifactDir             string   `yaml:"artifact_dir"`
	Enabled                 bool     `yaml:"enabled"`
	MaxNegotiationRounds    int      `yaml:"max_negotiation_rounds"`
	OptionalHarnessLevels   []string `yaml:"optional_harness_levels"`
	RequiredHarnessLevels   []string `yaml:"required_harness_levels"`
}

// harnessFileWrapper is the wrapper used for unmarshaling the harness.yaml file.
type harnessFileWrapper struct {
	Harness HarnessConfig `yaml:"harness"`
}

// MIG-003 wrapper types for the 4 new section files.

// constitutionFileWrapper is the wrapper used for unmarshaling the constitution.yaml file.
// Note: Both constitution.yaml and quality.yaml use top-level key "constitution:".
// They are disambiguated by filename in loadYAMLFile (REQ-MIG003 risk §11.1).
type constitutionFileWrapper struct {
	Constitution ConstitutionConfig `yaml:"constitution"`
}

// contextFileWrapper is the wrapper used for unmarshaling the context.yaml file.
// context.yaml uses top-level key "context_search:" (NOT "context:").
type contextFileWrapper struct {
	ContextSearch ContextConfig `yaml:"context_search"`
}

// interviewFileWrapper is the wrapper used for unmarshaling the interview.yaml file.
type interviewFileWrapper struct {
	Interview InterviewConfig `yaml:"interview"`
}

// designFileWrapper is the wrapper used for unmarshaling the design.yaml file.
type designFileWrapper struct {
	Design DesignConfig `yaml:"design"`
}

// YAML file wrapper types for proper unmarshaling with top-level keys.
// Each section file wraps its content under a top-level key.

type userFileWrapper struct {
	User models.UserConfig `yaml:"user"`
}

type languageFileWrapper struct {
	Language models.LanguageConfig `yaml:"language"`
}

// qualityFileWrapper handles the quality.yaml file which uses "constitution:"
// as the top-level key (Python MoAI-ADK backward compatibility).
type qualityFileWrapper struct {
	Constitution models.QualityConfig `yaml:"constitution"`
}

// gitConventionFileWrapper handles the git-convention.yaml section file.
type gitConventionFileWrapper struct {
	GitConvention models.GitConventionConfig `yaml:"git_convention"`
}

// llmFileWrapper handles the llm.yaml section file.
type llmFileWrapper struct {
	LLM LLMConfig `yaml:"llm"`
}

// stateFileWrapper handles the state.yaml section file.
type stateFileWrapper struct {
	State StateConfig `yaml:"state"`
}

// statuslineFileWrapper handles the statusline.yaml section file.
type statuslineFileWrapper struct {
	Statusline models.StatuslineConfig `yaml:"statusline"`
}

// researchFileWrapper handles the research.yaml section file.
type researchFileWrapper struct {
	Research ResearchConfig `yaml:"research"`
}

// ralphFileWrapper handles the ralph.yaml section file.
// stale_seconds lives under the ralph: key in ralph.yaml and is injected into Config.Session.StaleSeconds.
// SPEC-V3R2-RT-004 REQ-022: source of the STALE_SECONDS setting.
type ralphFileWrapper struct {
	Ralph struct {
		RalphConfig  `yaml:",inline"`
		StaleSeconds int `yaml:"stale_seconds"` // → Config.Session.StaleSeconds
	} `yaml:"ralph"`
}
