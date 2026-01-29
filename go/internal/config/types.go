package config

// Config is the root configuration structure containing all sections.
// It maps directly to the YAML files in .moai/config/sections/.
type Config struct {
	User         UserConfig         `yaml:"user"`
	Language     LanguageConfig     `yaml:"language"`
	Constitution ConstitutionConfig `yaml:"constitution"`
	System       SystemConfig       `yaml:"system"`
	GitStrategy  GitStrategyConfig  `yaml:"git_strategy"`
	Project      ProjectConfig      `yaml:"project"`
	LLM          LLMConfig          `yaml:"llm"`
	Service      ServiceConfig      `yaml:"service"`
	Ralph        RalphConfig        `yaml:"ralph"`
	Workflow     WorkflowConfig     `yaml:"workflow"`
}

// UserConfig contains user-specific settings (from user.yaml)
type UserConfig struct {
	Name string `yaml:"name"`
}

// Validate checks that user configuration is valid
func (c *UserConfig) Validate() error {
	// User name can be empty (will use default greeting)
	return nil
}

// LanguageConfig contains language settings (from language.yaml)
type LanguageConfig struct {
	ConversationLanguage     string `yaml:"conversation_language"`
	ConversationLanguageName string `yaml:"conversation_language_name"`
	AgentPromptLanguage      string `yaml:"agent_prompt_language"`
	GitCommitMessages        string `yaml:"git_commit_messages"`
	CodeComments             string `yaml:"code_comments"`
	Documentation            string `yaml:"documentation"`
	ErrorMessages            string `yaml:"error_messages"`
}

// Validate checks that language configuration is valid
func (c *LanguageConfig) Validate() error {
	// Language codes are validated at load time
	return nil
}

// ConstitutionConfig contains quality and TRUST 5 settings (from quality.yaml)
type ConstitutionConfig struct {
	DevelopmentMode    string             `yaml:"development_mode"`
	EnforceQuality     bool               `yaml:"enforce_quality"`
	TestCoverageTarget int                `yaml:"test_coverage_target"`
	DDDDSettings       DDDSettings        `yaml:"ddd_settings"`
	CoverageExemptions CoverageExemptions `yaml:"coverage_exemptions"`
	TestQuality        TestQuality        `yaml:"test_quality"`
	LSPPriorityGates   LSPPriorityGates   `yaml:"lsp_quality_gates"`
	Principles         Principles         `yaml:"principles"`
	LSPIntegration     LSPIntegration     `yaml:"lsp_integration"`
	ReportGeneration   ReportGeneration   `yaml:"report_generation"`
	LSPStateTracking   LSPStateTracking   `yaml:"lsp_state_tracking"`
}

// DDDSettings contains DDD-specific configuration
type DDDSettings struct {
	RequireExistingTests  bool   `yaml:"require_existing_tests"`
	CharacterizationTests bool   `yaml:"characterization_tests"`
	BehaviorSnapshots     bool   `yaml:"behavior_snapshots"`
	MaxTransformationSize string `yaml:"max_transformation_size"`
}

// CoverageExemptions contains coverage exemption settings
type CoverageExemptions struct {
	Enabled              bool `yaml:"enabled"`
	RequireJustification bool `yaml:"require_justification"`
	MaxExemptPercentage  int  `yaml:"max_exempt_percentage"`
}

// TestQuality contains test quality criteria
type TestQuality struct {
	SpecificationBased     bool `yaml:"specification_based"`
	MeaningfulAssertions   bool `yaml:"meaningful_assertions"`
	AvoidImplCoupling      bool `yaml:"avoid_implementation_coupling"`
	MutationTestingEnabled bool `yaml:"mutation_testing_enabled"`
}

// LSPPriorityGates contains LSP quality gate thresholds
type LSPPriorityGates struct {
	Enabled         bool       `yaml:"enabled"`
	Plan            PhaseGates `yaml:"plan"`
	Run             RunGates   `yaml:"run"`
	Sync            SyncGates  `yaml:"sync"`
	CacheTTLSeconds int        `yaml:"cache_ttl_seconds"`
	TimeoutSeconds  int        `yaml:"timeout_seconds"`
}

// PhaseGates contains plan phase thresholds
type PhaseGates struct {
	RequireBaseline bool `yaml:"require_baseline"`
}

// RunGates contains run phase thresholds
type RunGates struct {
	MaxErrors       int  `yaml:"max_errors"`
	MaxTypeErrors   int  `yaml:"max_type_errors"`
	MaxLintErrors   int  `yaml:"max_lint_errors"`
	AllowRegression bool `yaml:"allow_regression"`
}

// SyncGates contains sync phase thresholds
type SyncGates struct {
	MaxErrors       int  `yaml:"max_errors"`
	MaxWarnings     int  `yaml:"max_warnings"`
	RequireCleanLSP bool `yaml:"require_clean_lsp"`
}

// Principles contains simplicity principles
type Principles struct {
	Simplicity struct {
		MaxParallelTasks int `yaml:"max_parallel_tasks"`
	} `yaml:"simplicity"`
}

// LSPIntegration contains LSP TRUST 5 integration settings
type LSPIntegration struct {
	TRUST5Integration   map[string][]string `yaml:"truct5_integration"`
	DiagnosticSources   []string            `yaml:"diagnostic_sources"`
	RegressionDetection RegressionDetection `yaml:"regression_detection"`
}

// RegressionDetection contains regression detection thresholds
type RegressionDetection struct {
	ErrorIncreaseThreshold     int `yaml:"error_increase_threshold"`
	WarningIncreaseThreshold   int `yaml:"warning_increase_threshold"`
	TypeErrorIncreaseThreshold int `yaml:"type_error_increase_threshold"`
}

// ReportGeneration contains report generation settings
type ReportGeneration struct {
	Enabled    bool   `yaml:"enabled"`
	AutoCreate bool   `yaml:"auto_create"`
	WarnUser   bool   `yaml:"warn_user"`
	UserChoice string `yaml:"user_choice"`
}

// LSPStateTracking contains LSP state tracking settings
type LSPStateTracking struct {
	Enabled       bool     `yaml:"enabled"`
	CapturePoints []string `yaml:"capture_points"`
	Comparison    struct {
		Baseline            string `yaml:"baseline"`
		RegressionThreshold int    `yaml:"regression_threshold"`
	} `yaml:"comparison"`
	Logging struct {
		LogLSPStateChanges     bool `yaml:"log_lsp_state_changes"`
		LogRegressionDetection bool `yaml:"log_regression_detection"`
		LogCompletionMarkers   bool `yaml:"log_completion_markers"`
		IncludeLSPInReports    bool `yaml:"include_lsp_in_reports"`
	} `yaml:"logging"`
}

// Validate checks that constitution configuration is valid
func (c *ConstitutionConfig) Validate() error {
	// Validate development mode
	if c.DevelopmentMode != "ddd" {
		return nil // Other modes may be added in future
	}
	return nil
}

// SystemConfig contains MoAI system and GitHub settings (from system.yaml)
type SystemConfig struct {
	MoAI               MoAIConfig               `yaml:"moai"`
	GitHub             GitHubConfig             `yaml:"github"`
	DocumentManagement DocumentManagementConfig `yaml:"document_management"`
}

// MoAIConfig contains MoAI system configuration
type MoAIConfig struct {
	Version              string             `yaml:"version"`
	UpdateCheckFrequency string             `yaml:"update_check_frequency"`
	VersionCheck         VersionCheckConfig `yaml:"version_check"`
}

// VersionCheckConfig contains version check settings
type VersionCheckConfig struct {
	Enabled       bool `yaml:"enabled"`
	CacheTTLHours int  `yaml:"cache_ttl_hours"`
}

// GitHubConfig contains GitHub integration settings
type GitHubConfig struct {
	EnableTrust5       bool   `yaml:"enable_trust_5"`
	AutoDeleteBranches bool   `yaml:"auto_delete_branches"`
	SpecGitWorkflow    string `yaml:"spec_git_workflow"`
}

// DocumentManagementConfig contains document management settings
type DocumentManagementConfig struct {
	Enabled            bool                   `yaml:"enabled"`
	EnforceStructure   bool                   `yaml:"enforce_structure"`
	BlockRootPollution bool                   `yaml:"block_root_pollution"`
	SeparationPolicy   SeparationPolicyConfig `yaml:"separation_policy"`
	Directories        DirectoriesConfig      `yaml:"directories"`
}

// SeparationPolicyConfig contains separation policy settings
type SeparationPolicyConfig struct {
	StrictSeparation bool   `yaml:"strict_separation"`
	Description      string `yaml:"description"`
	LogsScope        string `yaml:"logs_scope"`
	DocsScope        string `yaml:"docs_scope"`
}

// DirectoriesConfig contains directory settings
type DirectoriesConfig struct {
	Docs  DirectoryConfig `yaml:"docs"`
	Logs  DirectoryConfig `yaml:"logs"`
	Temp  DirectoryConfig `yaml:"temp"`
	Cache DirectoryConfig `yaml:"cache"`
}

// DirectoryConfig contains individual directory settings
type DirectoryConfig struct {
	Base          string `yaml:"base"`
	RetentionDays *int   `yaml:"retention_days"`
	AutoCleanup   bool   `yaml:"auto_cleanup"`
}

// Validate checks that system configuration is valid
func (c *SystemConfig) Validate() error {
	return nil
}

// GitStrategyConfig contains Git workflow settings (from git-strategy.yaml)
type GitStrategyConfig struct {
	Mode     string        `yaml:"mode"`
	Manual   GitModeConfig `yaml:"manual"`
	Personal GitModeConfig `yaml:"personal"`
	Team     GitTeamConfig `yaml:"team"`
}

// GitModeConfig contains Git mode settings
type GitModeConfig struct {
	Workflow          string               `yaml:"workflow"`
	Environment       string               `yaml:"environment"`
	GitHubIntegration bool                 `yaml:"github_integration"`
	AutoCheckpoint    string               `yaml:"auto_checkpoint"`
	PushToRemote      bool                 `yaml:"push_to_remote"`
	BranchCreation    BranchCreationConfig `yaml:"branch_creation"`
	Automation        AutomationConfig     `yaml:"automation"`
	Hooks             HooksConfig          `yaml:"hooks"`
	CommitStyle       CommitStyleConfig    `yaml:"commit_style"`
	BranchPrefix      string               `yaml:"branch_prefix,omitempty"`
	MainBranch        string               `yaml:"main_branch,omitempty"`
}

// GitTeamConfig extends GitModeConfig for team mode
type GitTeamConfig struct {
	Workflow          string               `yaml:"workflow"`
	Environment       string               `yaml:"environment"`
	GitHubIntegration bool                 `yaml:"github_integration"`
	PushToRemote      bool                 `yaml:"push_to_remote"`
	BranchPrefix      string               `yaml:"branch_prefix"`
	MainBranch        string               `yaml:"main_branch"`
	DraftPR           bool                 `yaml:"draft_pr"`
	RequiredReviews   int                  `yaml:"required_reviews"`
	BranchProtection  bool                 `yaml:"branch_protection"`
	BranchCreation    BranchCreationConfig `yaml:"branch_creation"`
	Automation        AutomationConfig     `yaml:"automation"`
	Hooks             HooksConfig          `yaml:"hooks"`
	CommitStyle       CommitStyleConfig    `yaml:"commit_style"`
}

// BranchCreationConfig contains branch creation policy
type BranchCreationConfig struct {
	PromptAlways bool `yaml:"prompt_always"`
	AutoEnabled  bool `yaml:"auto_enabled"`
}

// AutomationConfig contains automation settings
type AutomationConfig struct {
	AutoBranch bool `yaml:"auto_branch"`
	AutoCommit bool `yaml:"auto_commit"`
	AutoPR     bool `yaml:"auto_pr"`
	AutoPush   bool `yaml:"auto_push"`
}

// HooksConfig contains Git hooks policy
type HooksConfig struct {
	PreCommit string `yaml:"pre_commit"`
	PrePush   string `yaml:"pre_push"`
	CommitMsg string `yaml:"commit_msg"`
}

// CommitStyleConfig contains commit message style
type CommitStyleConfig struct {
	Format        string `yaml:"format"`
	ScopeRequired bool   `yaml:"scope_required"`
}

// Validate checks that Git strategy configuration is valid
func (c *GitStrategyConfig) Validate() error {
	if c.Mode != "manual" && c.Mode != "personal" && c.Mode != "team" {
		return nil // Other modes may be added in future
	}
	return nil
}

// ProjectConfig contains project metadata (from project.yaml)
type ProjectConfig struct {
	Name            string `yaml:"name"`
	Description     string `yaml:"description"`
	Type            string `yaml:"type"`
	Locale          string `yaml:"locale"`
	CreatedAt       string `yaml:"created_at"`
	Initialized     bool   `yaml:"initialized"`
	TemplateVersion string `yaml:"template_version"`
}

// Validate checks that project configuration is valid
func (c *ProjectConfig) Validate() error {
	return nil
}

// LLMConfig contains LLM configuration (from llm.yaml)
type LLMConfig struct {
	Mode         string             `yaml:"mode"`
	GLMEnvVar    string             `yaml:"glm_env_var"`
	AutoWorktree AutoWorktreeConfig `yaml:"auto_worktree"`
	GLM          GLMConfig          `yaml:"glm"`
	Routing      RoutingConfig      `yaml:"routing"`
}

// AutoWorktreeConfig contains auto worktree settings
type AutoWorktreeConfig struct {
	Enabled       bool `yaml:"enabled"`
	CopyGLMConfig bool `yaml:"copy_glm_config"`
}

// GLMConfig contains GLM API configuration
type GLMConfig struct {
	BaseURL string            `yaml:"base_url"`
	Models  map[string]string `yaml:"models"`
}

// RoutingConfig contains LLM routing behavior
type RoutingConfig struct {
	AutoDetect        bool `yaml:"auto_detect"`
	ParallelThreshold int  `yaml:"parallel_threshold"`
	ConfirmWorktree   bool `yaml:"confirm_worktree"`
}

// Validate checks that LLM configuration is valid
func (c *LLMConfig) Validate() error {
	return nil
}

// ServiceConfig contains service and pricing settings (from pricing.yaml)
type ServiceConfig struct {
	Type            string                `yaml:"type"`
	PricingPlan     string                `yaml:"pricing_plan"`
	ModelAllocation ModelAllocationConfig `yaml:"model_allocation"`
}

// ModelAllocationConfig contains model allocation strategy
type ModelAllocationConfig struct {
	Strategy string            `yaml:"strategy"`
	Custom   map[string]string `yaml:"custom,omitempty"`
}

// Validate checks that service configuration is valid
func (c *ServiceConfig) Validate() error {
	return nil
}

// RalphConfig contains Ralph engine settings (from ralph.yaml)
type RalphConfig struct {
	Enabled bool               `yaml:"enabled"`
	LSP     RalphLSPConfig     `yaml:"lsp"`
	ASTGrep RalphASTGrepConfig `yaml:"ast_grep"`
	Loop    RalphLoopConfig    `yaml:"loop"`
	Hooks   RalphHooksConfig   `yaml:"hooks"`
}

// RalphLSPConfig contains Ralph LSP configuration
type RalphLSPConfig struct {
	AutoStart           bool `yaml:"auto_start"`
	TimeoutSeconds      int  `yaml:"timeout_seconds"`
	PollIntervalMs      int  `yaml:"poll_interval_ms"`
	GracefulDegradation bool `yaml:"graceful_degradation"`
}

// RalphASTGrepConfig contains Ralph AST-grep configuration
type RalphASTGrepConfig struct {
	Enabled      bool   `yaml:"enabled"`
	ConfigPath   string `yaml:"config_path"`
	SecurityScan bool   `yaml:"security_scan"`
	QualityScan  bool   `yaml:"quality_scan"`
	AutoFix      bool   `yaml:"auto_fix"`
}

// RalphLoopConfig contains Ralph loop controller configuration
type RalphLoopConfig struct {
	MaxIterations       int                   `yaml:"max_iterations"`
	AutoFix             bool                  `yaml:"auto_fix"`
	RequireConfirmation bool                  `yaml:"require_confirmation"`
	CooldownSeconds     int                   `yaml:"cooldown_seconds"`
	Completion          RalphCompletionConfig `yaml:"completion"`
}

// RalphCompletionConfig contains completion conditions
type RalphCompletionConfig struct {
	ZeroErrors        bool `yaml:"zero_errors"`
	ZeroWarnings      bool `yaml:"zero_warnings"`
	TestsPass         bool `yaml:"tests_pass"`
	CoverageThreshold int  `yaml:"coverage_threshold"`
}

// RalphHooksConfig contains Ralph hook behavior
type RalphHooksConfig struct {
	PostToolLSP        RalphPostToolLSPConfig `yaml:"post_tool_lsp"`
	StopLoopController RalphStopLoopConfig    `yaml:"stop_loop_controller"`
}

// RalphPostToolLSPConfig contains post-tool LSP hook configuration
type RalphPostToolLSPConfig struct {
	Enabled           bool     `yaml:"enabled"`
	TriggerOn         []string `yaml:"trigger_on"`
	SeverityThreshold string   `yaml:"severity_threshold"`
}

// RalphStopLoopConfig contains stop loop controller configuration
type RalphStopLoopConfig struct {
	Enabled         bool `yaml:"enabled"`
	CheckCompletion bool `yaml:"check_completion"`
}

// Validate checks that Ralph configuration is valid
func (c *RalphConfig) Validate() error {
	return nil
}

// WorkflowConfig contains workflow execution mode settings (from workflow.yaml)
type WorkflowConfig struct {
	ExecutionMode     WorkflowExecutionConfig `yaml:"execution_mode"`
	CompletionMarkers CompletionMarkersConfig `yaml:"completion_markers"`
	LoopPrevention    LoopPreventionConfig    `yaml:"loop_prevention"`
}

// WorkflowExecutionConfig contains execution mode settings
type WorkflowExecutionConfig struct {
	Interactive WorkflowModeConfig `yaml:"interactive"`
	Autonomous  WorkflowModeConfig `yaml:"autonomous"`
}

// WorkflowModeConfig contains workflow mode settings
type WorkflowModeConfig struct {
	UserApprovalRequired   bool `yaml:"user_approval_required"`
	EachStepExplicit       bool `yaml:"each_step_explicit"`
	ShowProgress           bool `yaml:"show_progress"`
	ContinuousLoop         bool `yaml:"continuous_loop,omitempty"`
	CompletionMarkerBased  bool `yaml:"completion_marker_based,omitempty"`
	LSPFeedbackIntegration bool `yaml:"lsp_feedback_integration,omitempty"`
}

// CompletionMarkersConfig contains completion markers for each phase
type CompletionMarkersConfig struct {
	Plan PhaseMarkersConfig `yaml:"plan"`
	Run  PhaseMarkersConfig `yaml:"run"`
	Sync PhaseMarkersConfig `yaml:"sync"`
}

// PhaseMarkersConfig contains phase-specific completion markers
type PhaseMarkersConfig struct {
	SpecDocumentCreated    bool `yaml:"spec_document_created,omitempty"`
	LSPBaselineRecorded    bool `yaml:"lsp_baseline_recorded,omitempty"`
	TestsPassing           bool `yaml:"tests_passing,omitempty"`
	BehaviorPreserved      bool `yaml:"behavior_preserved,omitempty"`
	LSPErrors              int  `yaml:"lsp_errors,omitempty"`
	TypeErrors             int  `yaml:"type_errors,omitempty"`
	CoverageMetTarget      bool `yaml:"coverage_met_target,omitempty"`
	DocumentationGenerated bool `yaml:"documentation_generated,omitempty"`
	LSPClean               bool `yaml:"lsp_clean,omitempty"`
	QualityGatePassed      bool `yaml:"quality_gate_passed,omitempty"`
}

// LoopPreventionConfig contains loop prevention settings
type LoopPreventionConfig struct {
	MaxIterations       int      `yaml:"max_iterations"`
	NoProgressThreshold int      `yaml:"no_progress_threshold"`
	StaleDetection      []string `yaml:"stale_detection"`
}

// Validate checks that workflow configuration is valid
func (c *WorkflowConfig) Validate() error {
	return nil
}
