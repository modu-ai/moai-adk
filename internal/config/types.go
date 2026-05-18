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
}

// GitStrategyConfig represents the git strategy configuration section.
type GitStrategyConfig struct {
	AutoBranch        bool   `yaml:"auto_branch"`
	BranchPrefix      string `yaml:"branch_prefix"`
	CommitStyle       string `yaml:"commit_style"`
	WorktreeRoot      string `yaml:"worktree_root"`
	Provider          string `yaml:"provider"`            // "github", "gitlab"
	GitLabInstanceURL string `yaml:"gitlab_instance_url"` // GitLab instance URL
}

// SystemHookConfig holds hook observability settings (SPEC-V3R2-RT-006 REQ-004).
// It controls which retired events are re-enabled as observability taps and
// whether strict mode behavior applies to retired events.
type SystemHookConfig struct {
	// ObservabilityEvents is the list of retired event names that are re-enabled
	// as observability taps. Empty list (default) means silent no-op for all.
	// Valid names: notification, elicitation, elicitationResult, taskCreated.
	ObservabilityEvents []string `yaml:"observability_events" validate:"omitempty"`
	// StrictMode: when true, retired events in strict mode still succeed silently.
	StrictMode bool `yaml:"strict_mode"`
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
// REQ-V3R2-RT-007-032: migrations.disabled로 session-start migration을 비활성화할 수 있습니다.
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
	RetentionDays int    `yaml:"retention_days"` // SPEC-V3R2-RT-004 REQ-031: runs/ 디렉토리 보존 일수
}

// SessionConfig holds session state management configuration.
// SPEC-V3R2-RT-004 REQ-022: STALE_SECONDS 설정.
type SessionConfig struct {
	// StaleSeconds는 checkpoint가 stale로 판정되는 기준 시간 (초).
	// 기본값: 3600 (1시간). ralph.yaml의 stale_seconds 키로 설정.
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

// HarnessConfig는 harness.yaml 최상위 설정 구조체입니다.
// @MX:ANCHOR: [AUTO] harness.yaml 전체 스키마의 Go 표현 — LoadHarnessConfig()의 반환 타입
// @MX:REASON: fan_in >= 3 (LoadHarnessConfig, router.Route, CLI validate 등 다수에서 소비)
// REQ-HRN-001-001: 전체 harness.yaml 스키마를 포괄하는 구조체 (HRN-001 run-phase 확장).
type HarnessConfig struct {
	// DefaultProfile은 기본 evaluator 프로필 이름입니다.
	DefaultProfile string `yaml:"default_profile"`
	// ModeDefaults는 실행 모드(solo/team/cg)별 기본 harness 레벨 맵입니다.
	// REQ-HRN-001-014: mode_defaults.cg = "thorough" (FROZEN).
	ModeDefaults map[string]string `yaml:"mode_defaults,omitempty"`
	// AutoDetection은 자동 감지 규칙 설정입니다.
	// REQ-HRN-001-007: minimal → standard → thorough 우선순위 순서.
	AutoDetection AutoDetectionConfig `yaml:"auto_detection,omitempty"`
	// Escalation은 에스컬레이션 트리거 설정입니다.
	// REQ-HRN-001-004/009/013: max_escalations + triggers.
	Escalation EscalationConfig `yaml:"escalation,omitempty"`
	// EffortMapping은 레벨 → 노력 수준 맵입니다.
	// REQ-HRN-001-005: minimal→medium, standard→high, thorough→xhigh.
	EffortMapping map[string]string `yaml:"effort_mapping,omitempty"`
	// Levels는 레벨별 설정 맵입니다.
	// REQ-HRN-001-001: {minimal, standard, thorough} FROZEN enum.
	Levels map[string]LevelConfig `yaml:"levels,omitempty"`
	// ModelUpgradeReview는 모델 업그레이드 검토 설정입니다.
	// REQ-HRN-001-016.
	ModelUpgradeReview ModelUpgradeReviewConfig `yaml:"model_upgrade_review,omitempty"`
	// PlanAuditGlobal은 전역 plan audit 설정입니다.
	PlanAuditGlobal PlanAuditGlobalConfig `yaml:"plan_audit_global,omitempty"`
	// Evaluator는 HRN-002 substrate — memory_scope FROZEN 검증용.
	Evaluator EvaluatorConfig `yaml:"evaluator"`
}

// AutoDetectionConfig는 auto_detection 블록의 설정 구조체입니다.
// REQ-HRN-001-007: rules 맵의 우선순위는 minimal → standard → thorough입니다.
type AutoDetectionConfig struct {
	// Enabled는 자동 감지 활성화 여부입니다.
	Enabled bool `yaml:"enabled"`
	// Rules는 레벨별 감지 조건 맵입니다.
	Rules map[string]AutoDetectionRule `yaml:"rules,omitempty"`
}

// AutoDetectionRule은 단일 레벨의 자동 감지 조건 목록입니다.
type AutoDetectionRule struct {
	// Conditions는 이 레벨로 라우팅되기 위한 조건 문자열 목록입니다.
	Conditions []string `yaml:"conditions,omitempty"`
}

// EscalationConfig는 escalation 블록의 설정 구조체입니다.
// REQ-HRN-001-004/009/013: max_escalations 상한 + 트리거 목록.
type EscalationConfig struct {
	// Enabled는 에스컬레이션 활성화 여부입니다.
	Enabled bool `yaml:"enabled"`
	// MaxEscalations는 단계당 최대 에스컬레이션 횟수입니다 (기본값 2, 상한 3).
	// REQ-HRN-001-013: hard ceiling = 3.
	MaxEscalations int `yaml:"max_escalations"`
	// Triggers는 에스컬레이션을 발동하는 이벤트 목록입니다.
	// (예: quality_gate_fail, review_critical, test_coverage_low)
	Triggers []string `yaml:"triggers,omitempty"`
}

// LevelConfig는 단일 harness 레벨의 설정 구조체입니다.
// REQ-HRN-001-001: levels.{minimal,standard,thorough} 각각의 설정.
type LevelConfig struct {
	// Description은 이 레벨에 대한 설명입니다.
	Description string `yaml:"description,omitempty"`
	// Evaluator는 evaluator 활성화 여부입니다.
	Evaluator bool `yaml:"evaluator"`
	// EvaluatorMode는 evaluator 모드입니다 (final-pass 또는 per-sprint).
	EvaluatorMode string `yaml:"evaluator_mode,omitempty"`
	// EvaluatorProfile은 사용할 evaluator 프로필 이름입니다.
	// 값이 있으면 .moai/config/evaluator-profiles/{name}.md 파일로 해석합니다.
	EvaluatorProfile string `yaml:"evaluator_profile,omitempty"`
	// SprintContract는 sprint contract 활성화 여부입니다.
	SprintContract bool `yaml:"sprint_contract"`
	// PlaywrightTesting은 playwright 테스트 활성화 여부입니다.
	PlaywrightTesting bool `yaml:"playwright_testing"`
	// SkipPhases는 건너뛸 워크플로우 단계 목록입니다.
	SkipPhases []any `yaml:"skip_phases,omitempty"`
	// PlanAudit은 plan audit 설정입니다.
	PlanAudit PlanAuditConfig `yaml:"plan_audit,omitempty"`
}

// PlanAuditConfig는 plan audit 설정 구조체입니다.
type PlanAuditConfig struct {
	// Enabled는 plan audit 활성화 여부입니다.
	Enabled bool `yaml:"enabled"`
	// MaxIterations는 최대 반복 횟수입니다.
	MaxIterations int `yaml:"max_iterations"`
	// RequireMustPass는 must-pass 요구 여부입니다.
	RequireMustPass bool `yaml:"require_must_pass"`
	// CrossValidateWithEvaluatorActive는 evaluator-active 교차 검증 여부입니다.
	CrossValidateWithEvaluatorActive bool `yaml:"cross_validate_with_evaluator_active"`
}

// ModelUpgradeReviewConfig는 model_upgrade_review 블록의 설정 구조체입니다.
// REQ-HRN-001-016: 모델 업그레이드 시 체크리스트 알림.
type ModelUpgradeReviewConfig struct {
	// Enabled는 모델 업그레이드 검토 활성화 여부입니다.
	Enabled bool `yaml:"enabled"`
	// Checklist는 검토 항목 목록입니다.
	Checklist []ReviewChecklistItem `yaml:"checklist,omitempty"`
	// Trigger는 검토 트리거 설정입니다.
	Trigger ModelUpgradeTrigger `yaml:"trigger,omitempty"`
	// Output은 검토 출력 설정입니다.
	Output ModelUpgradeOutput `yaml:"output,omitempty"`
}

// ReviewChecklistItem은 모델 업그레이드 검토 항목입니다.
type ReviewChecklistItem struct {
	ID       string `yaml:"id"`
	Question string `yaml:"question"`
	Action   string `yaml:"action"`
	Affects  string `yaml:"affects"`
}

// ModelUpgradeTrigger는 모델 업그레이드 검토 트리거 설정입니다.
type ModelUpgradeTrigger struct {
	OnModelChange     bool   `yaml:"on_model_change"`
	ManualCommand     string `yaml:"manual_command,omitempty"`
	ReviewIntervalDays int   `yaml:"review_interval_days"`
}

// ModelUpgradeOutput은 모델 업그레이드 검토 출력 설정입니다.
type ModelUpgradeOutput struct {
	ReportPath      string `yaml:"report_path,omitempty"`
	RequireApproval bool   `yaml:"require_approval"`
}

// PlanAuditGlobalConfig는 plan_audit_global 블록의 설정 구조체입니다.
type PlanAuditGlobalConfig struct {
	// AlwaysEnabled는 항상 plan audit 활성화 여부입니다.
	AlwaysEnabled bool `yaml:"always_enabled"`
	// EnforceGateOnSpecCreation는 SPEC 생성 시 gate 강제 여부입니다.
	EnforceGateOnSpecCreation bool `yaml:"enforce_gate_on_spec_creation"`
	// Rationale은 설정 이유 설명입니다.
	Rationale string `yaml:"rationale,omitempty"`
}

// EvaluatorConfig는 evaluator 하위 설정 구조체입니다.
// @MX:NOTE: FROZEN at per_iteration per design-constitution §11.4.1 (SPEC-V3R2-HRN-002)
// @MX:NOTE: [AUTO] HRN-003 M4: Profiles + Aggregation + MustPassDimensions 필드 추가 (SPEC-V3R2-HRN-003)
type EvaluatorConfig struct {
	// MemoryScope는 evaluator 메모리 범위 설정입니다.
	// design-constitution §11.4.1에 의해 per_iteration 값으로 FROZEN됩니다.
	// 다른 값(e.g., cumulative)은 HRN_EVAL_MEMORY_FROZEN 오류를 반환합니다.
	MemoryScope string `yaml:"memory_scope"`
	// Profiles는 evaluator 프로필 이름 → .md 파일 경로 맵입니다.
	// REQ-HRN-003-005, AC-HRN-003-07.c.
	Profiles map[string]string `yaml:"profiles,omitempty"`
	// Aggregation은 기본 집계 방식입니다 ("min" 또는 "mean").
	// REQ-HRN-003-007: 기본값은 "min"입니다.
	Aggregation string `yaml:"aggregation,omitempty"`
	// MustPassDimensions는 must-pass 차원 이름 목록입니다.
	// REQ-HRN-003-018: 기본값은 [Functionality, Security]입니다.
	MustPassDimensions []string `yaml:"must_pass_dimensions,omitempty"`
}

// harnessFileWrapper는 harness.yaml 파일 언마샬링용 래퍼입니다.
type harnessFileWrapper struct {
	Harness HarnessConfig `yaml:"harness"`
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
// stale_seconds는 ralph.yaml의 ralph: 키 하위에 위치하며 Config.Session.StaleSeconds에 주입됩니다.
// SPEC-V3R2-RT-004 REQ-022: STALE_SECONDS 설정 소스.
type ralphFileWrapper struct {
	Ralph struct {
		RalphConfig  `yaml:",inline"`
		StaleSeconds int `yaml:"stale_seconds"` // → Config.Session.StaleSeconds
	} `yaml:"ralph"`
}
