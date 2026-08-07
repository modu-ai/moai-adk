package config

import (
	"time"

	"github.com/modu-ai/moai-adk/pkg/models"
)

// Default value constants to avoid magic numbers and strings.
const (
	DefaultConversationLanguage     = "en"
	DefaultConversationLanguageName = "English"
	DefaultAgentPromptLanguage      = "en"
	DefaultGitCommitMessages        = "en"
	DefaultCodeComments             = "en"
	DefaultDocumentation            = "en"
	DefaultErrorMessages            = "en"

	DefaultTestCoverageTarget    = 85
	DefaultMaxTransformationSize = "small"
	DefaultMinCoveragePerCommit  = 80
	DefaultMaxExemptPercentage   = 5

	DefaultLogLevel  = "info"
	DefaultLogFormat = "text"

	DefaultModel      = "sonnet"
	DefaultQualModel  = "opus"
	DefaultSpeedModel = "haiku"

	DefaultTokenBudget = 250000

	DefaultMaxIterations = 5

	// DefaultAgenticLoopMaxIterations is the default for the pipeline-level
	// completion-loop iteration ceiling (workflow.agentic_loop.max_iterations).
	// DISTINCT from loop_prevention.max_iterations (default 100, per-operation
	// diagnostic fix-loop bound) — SPEC-V3R6-AGENTIC-LOOP-CONFIG-001 §A.4.
	// This is the single source of truth for the literal 10; all other
	// references MUST use this const (CLAUDE.local.md §14 — no hardcoding).
	DefaultAgenticLoopMaxIterations = 10

	DefaultPlanTokens = 30000
	DefaultRunTokens  = 180000
	DefaultSyncTokens = 40000

	// DefaultSecurityMaxScanBytes bounds the input the in-session security
	// guardian (SPEC-SEC-GUARDIAN-001) scans in a single pass. It mirrors the
	// handle-pre-tool.sh 1MB stdin cap precedent so a very large Write/Edit
	// payload is not scanned unboundedly within the 5s hook budget. This is the
	// single source of truth for the guardian scan cap (CLAUDE.local.md §14 —
	// no hardcoding).
	DefaultSecurityMaxScanBytes = 1048576

	// DefaultHookDispatcherTimeout is the per-dispatch timeout for the moai hook
	// subcommand's HookRegistry.Dispatch calls (SPEC-CLIFIX-HYGIENE-001
	// REQ-HYG-001-004). Formerly duplicated as an inline `30 * time.Second`
	// literal at two dispatcher call sites in internal/cli/hook.go; this is the
	// single source of truth.
	DefaultHookDispatcherTimeout = 30 * time.Second

	// DefaultTraceFlushTimeout bounds how long a hook process waits at teardown
	// for the async trace writer to drain to disk before abandoning the wait
	// (SPEC-HOOK-TRACE-FLUSH-001 REQ-HTF-006). It is the single source of truth
	// for the flush budget — call sites pass it as an argument rather than
	// inlining a literal (CLAUDE.local.md §14).
	//
	// This is a ceiling, not a cost: the drain barrier is signal-confirmed
	// (CloseWithTimeout blocks on the writer's done channel, not on the timer),
	// so the normal path returns as soon as the background goroutine finishes
	// draining. The timer only ever fires on a pathologically slow drain or a
	// genuine hang (REQ-HTF-004 forbids unbounded waits).
	//
	// The budget is provisional and corrected by measurement (SPEC §3.1
	// REQ-HTF-013). The prior 200ms value was calibrated against a local-SSD
	// measurement (p99 140µs for the 3-entry dispatch case, p99 3.4ms / max
	// 5.3ms at the 100-entry channel capacity) and held ~40x headroom there.
	// It proved insufficient under ubuntu CI's `-race` + coverage
	// instrumentation, where per-syscall overhead inflates 5-20x and the
	// writeEntry path (MkdirAll + Stat + OpenFile + Write + Close per entry)
	// pushed a 3-handler dispatch's drain past 200ms deterministically —
	// costing the trailing handlers' trace entries (and, in the worst case, all
	// entries when the background goroutine was starved off the scheduler). 2s
	// restores ~10x headroom over that CI failure threshold while remaining a
	// bounded, no-stall cap: only the one-shot hook CLI entrypoints defer
	// Shutdown (the sole caller), so this only binds how long a `moai hook
	// <event>` process lingers at exit, and only when the drain is actually
	// slow — the signal path returns in microseconds on a healthy filesystem.
	DefaultTraceFlushTimeout = 2 * time.Second

	DefaultBranchPrefix = "moai/"
	DefaultCommitStyle  = "conventional"

	// MOAI_AUTONOMY_TIER 3-value enum (SPEC-STOPCHAIN-TRIM-001 REQ-003 / §3.1).
	// The canonical values for the autonomy-tier mode token. Shell hooks read
	// "$MOAI_AUTONOMY_TIER" verbatim and string-compare against these literals;
	// the Go reader (AutonomyTier() in autonomy.go) normalizes case + whitespace
	// and falls back to AutonomyTierSemiAuto on unset/empty/invalid (backward
	// compat). The mode-aware hooks branch on these values:
	//   - semi-auto        : everything ON (today's behavior — the default)
	//   - automatic        : commit gate OFF, sync-gate build-only-block,
	//                        lifecycle hooks still active
	//   - fully-autonomous : commit gate OFF, sync-gate advisory-only,
	//                        lifecycle hooks dormant (observe-only)
	// The deny/ask denylist is tier-INVARIANT (REQ-007) — no value here weakens
	// a destructive-pattern deny.
	AutonomyTierSemiAuto        = "semi-auto"
	AutonomyTierAutomatic       = "automatic"
	AutonomyTierFullyAutonomous = "fully-autonomous"

	DefaultGLMEnvVar  = "GLM_API_KEY"
	DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"
	// GLM model tiers
	// High maps to glm-5.2 — the model id z.ai accepts. The [1m] suffix was
	// previously appended to activate Claude Code's 1M context mode, but Claude
	// Code forwards the suffix verbatim to the z.ai endpoint, which rejects it as
	// an unknown model — so the suffix is NOT used. The 1M auto-compact window is
	// now driven by the High slot model's resolved context window (see
	// glmAutoCompactWindow), not by a model-id suffix.
	DefaultGLMHigh   = "glm-5.2"
	DefaultGLMMedium = "glm-4.7"
	DefaultGLMLow    = "glm-4.5-air"
	DefaultGLMFable  = "glm-5.2"
	// Additional GLM models (available but not default-mapped)
	DefaultGLM45     = "glm-4.5"
	DefaultGLM46     = "glm-4.6"
	DefaultGLM51     = "glm-5.1"
	DefaultGLM5Turbo = "glm-5-turbo"
	// Legacy GLM model names (map to tiers)
	DefaultGLMHaiku  = "glm-4.5-air"
	DefaultGLMSonnet = "glm-4.7"
	DefaultGLMOpus   = "glm-5.2"
	// Default1MContextTokens is the token count for Claude Code's 1M context
	// mode. Used to populate CLAUDE_CODE_AUTO_COMPACT_WINDOW when the High slot
	// model resolves to the 1M context tier.
	Default1MContextTokens = 1_000_000
	// Default performance tier
	DefaultPerformanceTier = "medium"

	DefaultCacheTTLSeconds = 5
	DefaultTimeoutSeconds  = 3
	DefaultMaxWarnings     = 10

	DefaultGitConvention                    = "auto"
	DefaultGitConventionSampleSize          = 100
	DefaultGitConventionConfidenceThreshold = 0.5
	DefaultGitConventionFallback            = "conventional-commits"
	DefaultGitConventionMaxLength           = 100

	DefaultStateDir = ".moai/state"

	// DefaultTraceRetentionDays is the age threshold (in days) past which
	// non-empty trace-*.jsonl files under .moai/logs/ are pruned at SessionEnd
	// (SPEC-OBSERVE-HYGIENE-001 REQ-OBH-002). Zero-byte traces are pruned
	// unconditionally regardless of age; the current session's active trace is
	// always preserved (EC-3).
	DefaultTraceRetentionDays = 30

	// Memory taxonomy defaults (SPEC-V3R2-EXT-001)
	// @MX:NOTE: [AUTO] 메모리 감사 서브시스템의 실제 배선(wiring)은 아래 패키지 레벨 상수 +
	// MOAI_MEMORY_AUDIT 환경변수 경로다. 과거 workflow.memory.* YAML 블록을 미러링하던
	// 타입드 memory 설정 구조체는 프로덕션 소비자가 전무한 inert config(설정 극장)였기에 제거되었다.
	// SessionStart 훅과 감사 엔진은 이 상수들을 직접 읽고, PostToolUse 훅은 환경변수를 읽는다.
	DefaultMemoryStalenessHours          = 24  // files older than this are wrapped in staleness caveat
	DefaultMemoryIndexLineCap            = 200 // MEMORY.md lines beyond this trigger MEMORY_INDEX_OVERFLOW
	DefaultMemoryStaleAggregateThreshold = 10  // stale files >= this count emit one aggregated warning

	// DefaultFeedbackRepository is the default target repository for the /moai
	// feedback workflow (SPEC-INVOCATION-MODEL-001). Feedback targets the remote
	// MoAI-ADK tool repository (bug reports about the tool itself), NOT the user's
	// own repo; fork maintainers override via .moai/config/sections/feedback.yaml.
	DefaultFeedbackRepository = "modu-ai/moai-adk"

	// DefaultHandoffMode is the compiled default for HandoffConfig.Mode.
	// SPEC-HANDOFF-AUTORESUME-001: auto-resume is opt-in — the default is
	// "manual" (pure no-op), preserving the unchanged baseline UX.
	DefaultHandoffMode = "manual"

	// DefaultArchiveGraceDays is the compiled default for ArchiveConfig.GraceDays
	// — how long a terminal SPEC stays under .moai/specs/ after its last activity
	// before `moai spec archive` considers it eligible.
	//
	// SPEC-SESSIONSTART-PERF-001 REQ-SSP-009 / REQ-SSP-018: this is the single
	// source of truth for the literal 90; it mirrors the template-shipped
	// archive.yaml and backs the `--grace-days` flag default. It is the fallback
	// when archive.yaml is absent, and the zero-value guard in
	// Config.ArchiveGraceDays() — an unset grace window must never degrade into
	// "archive every terminal SPEC immediately".
	DefaultArchiveGraceDays = 90

	// SPEC-SESSIONSTART-PERF-001 M3 (REQ-SSP-014 / REQ-SSP-015 / REQ-SSP-018):
	// the three M3 regression-guard thresholds, extracted here so no literal
	// 4s / 500 lives inline in business logic or tests (CLAUDE.local.md §14).

	// DefaultDriftPerfBudget is the wall-clock budget the drift-detection
	// perf-regression test asserts against (REQ-SSP-014). The single-pass design
	// completes far under this at the current corpus; a budget breach signals the
	// O(n)-subprocess pattern has regressed.
	DefaultDriftPerfBudget = 4 * time.Second

	// DefaultSessionStartDriftTimeout time-boxes the session-start drift advisory
	// check (REQ-SSP-015). On deadline exceed the handler skips the (abandoned)
	// computation and emits the advisory instead of blocking session start
	// unboundedly — an advisory computation on the critical path must never block.
	DefaultSessionStartDriftTimeout = 2 * time.Second

	// DefaultCodexReviewGateTimeout is the per-invocation timeout override for
	// the `moai hook codex-review-gate` Stop hook (SPEC-MOAI-MCP-SERVER-001
	// REQ-MCP-008 / AC-MCP-010). The moai-default 5s hook budget does NOT apply
	// to this hook: a codex review legitimately runs for minutes, so the
	// manifest pins 900s (15 min) for this hook only. Other hooks keep 5s.
	// Centralized here per CLAUDE.local.md §14 (thresholds in defaults.go).
	DefaultCodexReviewGateTimeout = 900 * time.Second

	// DefaultMultiReviewGateTimeout is the per-invocation timeout override for
	// the `moai hook multi-review-gate` Stop hook (SPEC-AUDIT-MULTI-MODEL-001
	// M5 REQ-AMM-013 / AC-AMM-018). The moai-default 5s hook budget does NOT
	// apply to this hook: the gate itself only reads a state file, but the
	// generous 900s budget (sibling to DefaultCodexReviewGateTimeout) keeps
	// Stop-hook-composition uniform across both review gates so a future
	// evolution that adds I/O does not silently regress the budget.
	// Centralized here per CLAUDE.local.md §14 (thresholds in defaults.go).
	DefaultMultiReviewGateTimeout = 900 * time.Second

	// DefaultDriftPerfFixtureSpecs is the synthetic SPEC-directory count the
	// perf-regression fixture builds (REQ-SSP-014, N=500). It is the SSOT for the
	// literal 500 so the fixture size is not an inline magic number.
	DefaultDriftPerfFixtureSpecs = 500

	// SPEC-HANDOFF-THRESHOLD-001 (Handoff-v2 M4): handoff-guide band boundaries.
	// Named constants so renderer.go carries no inline band literals (§14
	// hardcoding-prevention). These boundaries are compiled-in and NOT
	// config-overridable — the "M3 lands / M4 consumes" contract keeps
	// HandoffConfig limited to {Mode, Guide}, so band boundaries never become
	// user-tunable config fields.
	//
	// Soft-stage thresholds (raw context-usage %): the large-window class uses
	// 50%, the standard/medium class uses 90%. HandoffLargeWindowCutoff (tokens)
	// separates the two classes. The hard (stage-2) ceiling is auto-compact-aware:
	// min(HandoffHardCeilingCapPct, autoCompactThreshold + HandoffHardCeilingMarginPct),
	// clamped up to the band's soft threshold when a degenerate override would put
	// it below soft.
	HandoffSoftLargePct         = 50      // ≥ HandoffLargeWindowCutoff → soft at 50% raw usage
	HandoffSoftStandardPct      = 90      // < HandoffLargeWindowCutoff → soft at 90% raw usage
	HandoffLargeWindowCutoff    = 500_000 // token cutoff separating large from standard/medium window
	HandoffHardCeilingCapPct    = 95      // absolute cap for the hard (stage-2) ceiling
	HandoffHardCeilingMarginPct = 10      // margin above auto-compact threshold for the hard ceiling
)

// SandboxProofKinds is the allowlist of recognized sandbox/container isolation
// kinds for the MOAI_SANDBOX_PROOF env marker (SPEC-AUTONOMY-TIERS-001 REQ-002
// S3 hardening — sandbox-proof spoofing). A proof whose kind is not in this
// list is rejected. Each kind names one of Claude Code's official full-process
// isolation options (container, VM, or the sandbox-runtime); the marker is an
// attestation and the allowlist is the proof — moai-adk does not re-verify the
// boundary from inside (see autonomy_tiers.go SandboxProofKind). The slice is a
// var (composite literals are not const expressions) so it remains
// config-extendable without code change; callers treat it as read-only.
var SandboxProofKinds = []string{
	"docker", "podman", "gvisor", "firecracker", "e2b", "devcontainer", "kata", "sandbox-runtime",
}

// DefaultHandoffStaleTTL is the age past which a handoff/pending.json is
// considered stale and silently removed by the SessionStart handler — auto-mode
// ONLY (SPEC-HANDOFF-AUTORESUME-001 REQ-019). Manual mode never removes a stale
// pending record (REQ-009 pure no-op). Single source of truth consumed by the
// M3 handoffInjectHandler; not a compile-time const because time.Duration
// multiplication is not a constant expression.
var DefaultHandoffStaleTTL = 7 * 24 * time.Hour

// DefaultTierThresholds is the canonical 4-tier harness-learning cutoff vector
// per V3R4-HARNESS-003 (count >= 1 → observation, >= 3 → heuristic,
// >= 5 → rule, >= 10 → auto_update). SPEC-CLIFIX-HYGIENE-001 REQ-HYG-001-004:
// formerly duplicated as an inline `[]int{1, 3, 5, 10}` literal at THREE sites
// (internal/cli/harness.go runHarnessStatus fallback, harness.go
// defaultLearningConfig struct initializer, and internal/cli/hook.go
// defaultTierThresholds). This var is the single source of truth; all three
// sites reference it. A `var` (not `const`) because composite literals are not
// constant expressions. Callers treat the value as read-only.
//
// @MX:ANCHOR: [AUTO] harness tier-threshold SSOT — single source for the 4-tier vector
// @MX:REASON: REQ-HYG-001-004; diverging defaults across the 3 former sites silently changed tier classification
var DefaultTierThresholds = []int{1, 3, 5, 10}

// NewDefaultConfig returns a Config with all fields set to compiled defaults.
func NewDefaultConfig() *Config {
	return &Config{
		User:          NewDefaultUserConfig(),
		Language:      NewDefaultLanguageConfig(),
		Quality:       NewDefaultQualityConfig(),
		Project:       NewDefaultProjectConfig(),
		GitStrategy:   NewDefaultGitStrategyConfig(),
		GitConvention: NewDefaultGitConventionConfig(),
		System:        NewDefaultSystemConfig(),
		LLM:           NewDefaultLLMConfig(),
		Pricing:       NewDefaultPricingConfig(),
		Ralph:         NewDefaultRalphConfig(),
		Workflow:      NewDefaultWorkflowConfig(),
		State:         NewDefaultStateConfig(),
		Gate:          NewDefaultGateConfig(),
		Sunset:        NewDefaultSunsetConfig(),
		Research:      NewDefaultResearchConfig(),
		Feedback:      NewDefaultFeedbackConfig(),
		Handoff:       NewDefaultHandoffConfig(),
		Archive:       NewDefaultArchiveConfig(),
		Session:       NewDefaultSessionConfig(),
		// MIG-003: 4 new section defaults (REQ-MIG003-004)
		Constitution:  defaultConstitutionConfig(),
		ContextSearch: defaultContextConfig(),
		Interview:     defaultInterviewConfig(),
		Design:        defaultDesignConfig(),
	}
}

// NewDefaultResearchConfig returns a ResearchConfig with safe defaults.
func NewDefaultResearchConfig() ResearchConfig {
	return ResearchConfig{
		Enabled: false,
		Passive: ResearchPassiveConfig{
			Enabled:                 true,
			CorrectionWindowSeconds: 60,
			PatternThresholds: ResearchPatternThresholds{
				Heuristic:      3,
				Rule:           5,
				HighConfidence: 10,
			},
		},
		Active: ResearchActiveConfig{
			RunsPerExperiment: 3,
			MaxExperiments:    20,
			PassThreshold:     0.80,
			TargetScore:       0.95,
			BudgetCapTokens:   500000,
		},
		Safety: ResearchSafetyConfig{
			WorktreeIsolation:         true,
			CanaryRegressionThreshold: 0.10,
			RateLimits: ResearchRateLimitConfig{
				MaxExperimentsPerSession: 20,
				MaxAcceptedPerSession:    5,
				MaxAutoResearchPerWeek:   3,
			},
		},
		Dashboard: ResearchDashboardConfig{
			DefaultMode:     "terminal",
			HTMLOpenBrowser: true,
		},
	}
}

// NewDefaultFeedbackConfig returns a FeedbackConfig whose target repository is
// the default tool feedback channel (DefaultFeedbackRepository). An absent
// feedback.yaml therefore still resolves to the tool channel.
func NewDefaultFeedbackConfig() FeedbackConfig {
	return FeedbackConfig{
		Repository: DefaultFeedbackRepository,
	}
}

// NewDefaultHandoffConfig returns a HandoffConfig with safe defaults.
// SPEC-HANDOFF-AUTORESUME-001 REQ-001: Mode defaults to "manual" (auto-resume
// is opt-in) and Guide defaults to false.
func NewDefaultHandoffConfig() HandoffConfig {
	return HandoffConfig{
		Mode:  DefaultHandoffMode,
		Guide: false,
	}
}

// NewDefaultArchiveConfig returns an ArchiveConfig with safe defaults.
// SPEC-SESSIONSTART-PERF-001 REQ-SSP-012: an absent archive.yaml still resolves
// to the 90-day grace window.
func NewDefaultArchiveConfig() ArchiveConfig {
	return ArchiveConfig{
		GraceDays: DefaultArchiveGraceDays,
	}
}

// NewDefaultGateConfig returns a GateConfig with production-safe defaults.
func NewDefaultGateConfig() GateConfig {
	return GateConfig{
		Enabled:   true,
		SkipTests: false,
		Timeouts: GateTimeouts{
			Vet:  30,
			Lint: 60,
			Test: 120,
		},
		// The ast-grep sub-gate is ON by default in advisory mode (findings
		// reported, commits never blocked); blocking is opt-in via gate.yaml.
		AstGrepGate: AstGrepGateConfig{
			Enabled:      true,
			BlockOnError: false,
			WarnOnlyMode: true,
		},
	}
}

// NewDefaultUserConfig returns a UserConfig with default values.
// Note: Name is intentionally empty; it is populated from user.yaml.
func NewDefaultUserConfig() models.UserConfig {
	return models.UserConfig{}
}

// NewDefaultLanguageConfig returns a LanguageConfig with default values.
func NewDefaultLanguageConfig() models.LanguageConfig {
	return models.LanguageConfig{
		ConversationLanguage:     DefaultConversationLanguage,
		ConversationLanguageName: DefaultConversationLanguageName,
		AgentPromptLanguage:      DefaultAgentPromptLanguage,
		GitCommitMessages:        DefaultGitCommitMessages,
		CodeComments:             DefaultCodeComments,
		Documentation:            DefaultDocumentation,
		ErrorMessages:            DefaultErrorMessages,
	}
}

// NewDefaultQualityConfig returns a QualityConfig with default values.
func NewDefaultQualityConfig() models.QualityConfig {
	return models.QualityConfig{
		DevelopmentMode:    models.ModeTDD,
		EnforceQuality:     true,
		TestCoverageTarget: DefaultTestCoverageTarget,
		DDDSettings:        NewDefaultDDDSettings(),
		TDDSettings:        NewDefaultTDDSettings(),
		CoverageExemptions: NewDefaultCoverageExemptions(),
	}
}

// NewDefaultDDDSettings returns DDDSettings with default values.
func NewDefaultDDDSettings() models.DDDSettings {
	return models.DDDSettings{
		RequireExistingTests:  true,
		CharacterizationTests: true,
		BehaviorSnapshots:     true,
		MaxTransformationSize: DefaultMaxTransformationSize,
		PreserveBeforeImprove: true,
	}
}

// NewDefaultTDDSettings returns TDDSettings with default values.
func NewDefaultTDDSettings() models.TDDSettings {
	return models.TDDSettings{
		RedGreenRefactor:       true,
		TestFirstRequired:      true,
		MinCoveragePerCommit:   DefaultMinCoveragePerCommit,
		MutationTestingEnabled: false,
	}
}

// NewDefaultCoverageExemptions returns CoverageExemptions with default values.
func NewDefaultCoverageExemptions() models.CoverageExemptions {
	return models.CoverageExemptions{
		Enabled:              false,
		RequireJustification: true,
		MaxExemptPercentage:  DefaultMaxExemptPercentage,
	}
}

// NewDefaultProjectConfig returns a ProjectConfig with default values.
func NewDefaultProjectConfig() models.ProjectConfig {
	return models.ProjectConfig{}
}

// NewDefaultGitStrategyConfig returns a GitStrategyConfig with default values.
//
// Per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 REQ-GSS-006: the three ModeProfile
// instances mirror the template-canonical defaults in
// internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl
// (the schema SSOT). The deprecated FLAT fields retain their pre-existing
// default values for backward-compat (Option (c)).
func NewDefaultGitStrategyConfig() GitStrategyConfig {
	return GitStrategyConfig{
		// Top-level wire-through defaults.
		Mode:           "team",
		Provider:       "github",
		GitHubUsername: "",
		GitLab:         GitLabConfig{InstanceURL: ""},

		Manual: ModeProfile{
			Workflow:          "github-flow",
			Environment:       "local",
			GitHubIntegration: false,
			PushToRemote:      false,
			AutoCheckpoint:    "disabled",
			MergeMethod:       "squash",
			BranchCreation:    BranchCreationConfig{AutoEnabled: false, PromptAlways: true},
			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: false},
			CommitStyle:       CommitStyleConfig{Format: "conventional", ScopeRequired: false},
			Hooks:             HooksConfig{PreCommit: "enforce", PrePush: "warn", CommitMsg: "warn"},
		},
		Personal: ModeProfile{
			Workflow:          "github-flow",
			Environment:       "github",
			GitHubIntegration: true,
			PushToRemote:      true,
			BranchPrefix:      "feature/SPEC-",
			MainBranch:        "main",
			MergeMethod:       "squash",
			BranchCreation:    BranchCreationConfig{AutoEnabled: false, PromptAlways: true},
			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: false},
			CommitStyle:       CommitStyleConfig{Format: "conventional", ScopeRequired: false},
			Hooks:             HooksConfig{PreCommit: "enforce", PrePush: "warn", CommitMsg: "warn"},
		},
		Team: ModeProfile{
			Workflow:          "github-flow",
			Environment:       "github",
			GitHubIntegration: true,
			PushToRemote:      true,
			BranchPrefix:      "feature/SPEC-",
			MainBranch:        "main",
			DraftPR:           true,
			RequiredReviews:   1,
			BranchProtection:  true,
			MergeMethod:       "squash",
			BranchCreation:    BranchCreationConfig{AutoEnabled: false, PromptAlways: true},
			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: true},
			CommitStyle:       CommitStyleConfig{Format: "conventional", ScopeRequired: true},
			Hooks:             HooksConfig{PreCommit: "enforce", PrePush: "warn", CommitMsg: "warn"},
		},

		// Deprecated FLAT fields — preserve existing default values for backward-compat.
		AutoBranch:        false,
		BranchPrefix:      DefaultBranchPrefix,
		CommitStyle:       DefaultCommitStyle,
		WorktreeRoot:      "",
		GitLabInstanceURL: "",
	}
}

// NewDefaultSystemConfig returns a SystemConfig with default values.
func NewDefaultSystemConfig() SystemConfig {
	return SystemConfig{
		LogLevel:  DefaultLogLevel,
		LogFormat: DefaultLogFormat,
	}
}

// @MX:ANCHOR: [AUTO] LLM configuration defaults factory. Single entry point for the full LLM config including model tiers, GLM settings, and performance policy.
// @MX:REASON: fan_in=6, referenced by many callers including config loader, CLI initialization, and test fixtures
// NewDefaultLLMConfig returns a LLMConfig with default values.
func NewDefaultLLMConfig() LLMConfig {
	return LLMConfig{
		GLMEnvVar:       DefaultGLMEnvVar,
		PerformanceTier: DefaultPerformanceTier,
		ClaudeModels: ClaudeTierModels{
			High:   "opus",
			Medium: "sonnet",
			Low:    "haiku",
		},
		DefaultModel: DefaultModel,
		QualityModel: DefaultQualModel,
		SpeedModel:   DefaultSpeedModel,
		GLM: GLMSettings{
			BaseURL: DefaultGLMBaseURL,
			Models: GLMModels{
				High:   DefaultGLMHigh,
				Medium: DefaultGLMMedium,
				Low:    DefaultGLMLow,
				Fable:  DefaultGLMFable,
				// Legacy fields for backward compatibility
				Opus:   DefaultGLMOpus,
				Sonnet: DefaultGLMSonnet,
				Haiku:  DefaultGLMHaiku,
			},
		},
	}
}

// NewDefaultPricingConfig returns a PricingConfig with default values.
func NewDefaultPricingConfig() PricingConfig {
	return PricingConfig{
		TokenBudget: DefaultTokenBudget,
	}
}

// NewDefaultRalphConfig returns a RalphConfig with default values.
func NewDefaultRalphConfig() RalphConfig {
	return RalphConfig{
		MaxIterations:     DefaultMaxIterations,
		AutoConverge:      true,
		HumanReview:       true,
		LintAsInstruction: true,  // REQ-LAI-003: enabled by default
		WarnAsInstruction: false, // REQ-LAI-006: disabled by default
	}
}

// NewDefaultWorkflowConfig returns a WorkflowConfig with default values.
// The nested defaults mirror the template SSOT workflow.yaml exactly
// (internal/template/templates/.moai/config/sections/workflow.yaml).
func NewDefaultWorkflowConfig() WorkflowConfig {
	return WorkflowConfig{
		AutoClear: AutoClearConfig{
			Enabled:        true,
			AfterPlan:      true,
			AfterRun:       false,
			TokenThreshold: 150000,
		},
		DefaultMode:   "",
		ExecutionMode: "team",
		AgenticLoop: AgenticLoopConfig{
			MaxIterations: DefaultAgenticLoopMaxIterations,
		},
		LoopPrevention: LoopPreventionConfig{
			FailurePatternDetection: true,
			MaxIterations:           100,
			MaxRetriesPerOperation:  3,
		},
		TokenBudget: TokenBudgetConfig{
			Plan: DefaultPlanTokens,
			Run:  DefaultRunTokens,
			Sync: DefaultSyncTokens,
		},
		Worktree: WorkflowWorktreeConfig{
			// SPEC-WORKTREE-ENTRY-STRATEGY-001 M1: web auto-toggles default OFF.
			// AutoCleanup and AutoMerge mutated true→false (sprawl mitigation,
			// EnterWorktree-first policy). AutoCreate unchanged (already false).
			AutoCleanup:        false,
			AutoCreate:         false,
			AutoMerge:          false,
			SessionNamePattern: "moai-{ProjectName}-{SPEC-ID}",
			TmuxPreferred:      true,
		},
		// SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 REQ-1/REQ-4: the guard ships
		// default-OFF (opt-in). Distributed users get an inert guard; the
		// maintainer of a shared multi-session checkout opts in via local
		// config. Template neutrality (CLAUDE.local.md §25): no `enabled: true`
		// anywhere under internal/template/templates/.
		BranchGuard: BranchGuardConfig{
			Enabled: false,
		},
		// SPEC-MOAI-MCP-SERVER-001 M2 (REQ-MCP-008 / C6): the codex review gate
		// ships default-OFF. Distributed users get an inert Stop hook; a
		// maintainer opts in via local config. Template neutrality (§25): no
		// `enabled: true` under internal/template/templates/.
		Codex: CodexConfig{
			ReviewGate: CodexReviewGateConfig{
				Enabled: false,
			},
		},
		// SPEC-AUDIT-MULTI-MODEL-001 M5 (REQ-AMM-013 / AC-AMM-018 / C6): the
		// multi-model review gate ships default-OFF (BranchGuard pattern —
		// sibling to Codex.ReviewGate). Distributed users get an inert Stop
		// hook; a maintainer opts in via local config. Template neutrality
		// (§25): no `enabled: true` under internal/template/templates/.
		Multi: MultiConfig{
			ReviewGate: MultiReviewGateConfig{
				Enabled: false,
			},
		},
		// SPEC-MOAI-MCP-SERVER-001 M3 (REQ-MCP-010 / AC-MCP-012, progress.md
		// §G.3 locked default profile): claude + codex required, glm advisory.
		// glm ships advisory (NOT required) so a distributed user without a GLM
		// key is never hard-blocked — the fail-open C2 invariant. `multi` is a
		// declared token only; convergence logic is SPEC-AUDIT-MULTI-MODEL.
		Audit: AuditConfig{
			Model: AuditModelClaude,
			Gates: AuditGates{
				Claude: AuditGateRequired,
				Codex:  AuditGateRequired,
				GLM:    AuditGateAdvisory,
			},
		},
		// SPEC-SESSION-WORKTREE-001 REQ-SW-001: the session-worktree auto-entry
		// feature ships default-OFF. When unset, moai init / moai profile /
		// moai web behave byte-identically to the shared-checkout baseline.
		// Template neutrality (CLAUDE.local.md §25): the default lives in Go
		// code, NOT in internal/template/templates/.moai/config/sections/
		// workflow.yaml — the distributed template MUST NOT leak this sub-key.
		SessionWorktree: SessionWorktreeConfig{
			Enabled: false,
		},
	}
}

// NewDefaultStateConfig returns a StateConfig with default values.
func NewDefaultStateConfig() StateConfig {
	return StateConfig{
		StateDir: DefaultStateDir,
	}
}

// NewDefaultSessionConfig returns a SessionConfig with default values.
//
// The StaleSeconds default was removed in SPEC-RALPH-CONFIG-REDESIGN-001 M3
// (dead producer-side pipeline, zero runtime consumers). The constructor is
// retained because NewDefaultConfig wires Config.Session via it.
func NewDefaultSessionConfig() SessionConfig {
	return SessionConfig{}
}

// NewDefaultGitConventionConfig returns a GitConventionConfig with default values.
func NewDefaultGitConventionConfig() models.GitConventionConfig {
	return models.GitConventionConfig{
		Convention: DefaultGitConvention,
		AutoDetection: models.AutoDetectionConfig{
			Enabled:             true,
			SampleSize:          DefaultGitConventionSampleSize,
			ConfidenceThreshold: DefaultGitConventionConfidenceThreshold,
			Fallback:            DefaultGitConventionFallback,
		},
		Validation: models.ConventionValidationConfig{
			EnforceOnPush: false,
			MaxLength:     DefaultGitConventionMaxLength,
		},
	}
}

// NewDefaultSunsetConfig returns a SunsetConfig with default values.
func NewDefaultSunsetConfig() SunsetConfig {
	return SunsetConfig{
		Enabled:    false,
		Conditions: nil,
	}
}

// NewDefaultLSPQualityGates returns LSPQualityGates with default values.
func NewDefaultLSPQualityGates() LSPQualityGates {
	return LSPQualityGates{
		Enabled: true,
		Plan: PlanGate{
			RequireBaseline: true,
		},
		Run: RunGate{
			MaxErrors:       0,
			MaxTypeErrors:   0,
			MaxLintErrors:   0,
			AllowRegression: false,
		},
		Sync: SyncGate{
			MaxErrors:       0,
			MaxWarnings:     DefaultMaxWarnings,
			RequireCleanLSP: true,
		},
		CacheTTLSeconds: DefaultCacheTTLSeconds,
		TimeoutSeconds:  DefaultTimeoutSeconds,
	}
}

// defaultConstitutionConfig returns a ConstitutionConfig with defaults matching
// internal/template/templates/.moai/config/sections/constitution.yaml.
// REQ-MIG003-004: sensible defaults on absent file.
func defaultConstitutionConfig() ConstitutionConfig {
	return ConstitutionConfig{
		ApprovedFrameworks: []string{"cobra", "viper"},
		ApprovedLanguages:  []string{"go"},
		Architecture: ConstitutionArchitecture{
			ForbiddenDependencies: []string{
				"circular imports",
				"direct template access from CLI handlers",
			},
			Patterns: []string{"clean-architecture", "repository-pattern"},
		},
		ForbiddenPatterns: []string{
			"global mutable state",
			"init() with side effects",
			"panic() in library code",
			"raw SQL without parameterized queries",
		},
		NamingConventions: ConstitutionNaming{
			Exported: "PascalCase",
			Files:    "snake_case.go",
			Packages: "lowercase, single word",
		},
		Performance: ConstitutionPerformance{},
		Security: ConstitutionSecurity{
			ForbiddenPractices: []string{
				"hardcoded credentials",
				"os.Exit in library code",
			},
			RequiredChecks: []string{"input-validation"},
		},
	}
}

// defaultContextConfig returns a ContextConfig with defaults matching
// internal/template/templates/.moai/config/sections/context.yaml.
// REQ-MIG003-004: sensible defaults on absent file.
func defaultContextConfig() ContextConfig {
	return ContextConfig{
		AutoDetect: ContextAutoDetect{Enabled: true},
		Enabled:    true,
		MemoryIntegration: ContextMemoryIntegration{
			Enabled:            true,
			IncludeInContext:   true,
			PriorityOverSearch: true,
		},
		Performance: ContextPerformance{
			CacheTTLSeconds: 300,
			TimeoutSeconds:  10,
		},
		Search: ContextSearch{
			DateRangeDays:      30,
			MaxResults:         5,
			MaxTokensPerResult: 1000,
			ProjectScopeOnly:   true,
		},
		TokenBudget: ContextTokenBudget{
			MaxInjectionTokens: 5000,
			SkipIfUsageAbove:   150000,
		},
	}
}

// defaultInterviewConfig returns an InterviewConfig with defaults matching
// internal/template/templates/.moai/config/sections/interview.yaml.
// REQ-MIG003-004: sensible defaults on absent file.
func defaultInterviewConfig() InterviewConfig {
	return InterviewConfig{
		ClarityThreshold: 4,
		Enabled:          true,
		Plan: InterviewMode{
			MaxRounds:         5,
			QuestionsPerRound: 3,
		},
		Project: InterviewMode{
			MaxRounds:         3,
			QuestionsPerRound: 3,
		},
		SkipConditions: []string{
			"resume_spec_id_present",
			"skip_interview_flag",
			"technical_keywords_gte_5",
		},
	}
}

// defaultDesignConfig returns a DesignConfig with defaults matching
// internal/template/templates/.moai/config/sections/design.yaml.
// REQ-MIG003-004: sensible defaults on absent file.
// Note: PassThreshold default 0.75 is above the FROZEN floor 0.60.
func defaultDesignConfig() DesignConfig {
	return DesignConfig{
		Adaptation: DesignAdaptation{
			ConfidenceThreshold: 0.7,
			Enabled:             true,
			IterationLimits: DesignIterationLimits{
				Builder:    3,
				Copywriter: 3,
				Designer:   2,
			},
			MinProjectsForAdaptation: 5,
		},
		BrandContext: DesignBrandContext{
			Dir:                 ".moai/project/brand",
			InterviewOnFirstRun: true,
		},
		ClaudeDesign: DesignClaudeDesign{
			Enabled:                 true,
			FallbackPath:            "code_based",
			SupportedBundleVersions: []string{"1.0"},
		},
		DefaultFramework: "next.js",
		Enabled:          true,
		Evaluator: DesignEvaluator{
			MemoryScope: "per_iteration",
		},
		Evolution: DesignEvolution{
			ArchiveAfterEvolve:  true,
			AutoEvolveThreshold: 3,
			CooldownHours:       24,
			GraduationCriteria: DesignGraduationCriteria{
				ConsistencyRatio:    0.8,
				MinimumConfidence:   0.8,
				MinimumObservations: 5,
				StalenessWindowDays: 30,
			},
			MaxActiveLearnings:      50,
			MaxEvolutionRatePerWeek: 3,
			RequireApproval:         true,
		},
		Figma: DesignFigma{Enabled: false},
		GanLoop: DesignGanLoop{
			EscalationAfter:      3,
			ImprovementThreshold: 0.05,
			MaxIterations:        5,
			PassThreshold:        0.75,
			SprintContract: DesignSprintContract{
				ArtifactDir:           ".moai/sprints",
				Enabled:               true,
				MaxNegotiationRounds:  2,
				OptionalHarnessLevels: []string{"standard"},
				RequiredHarnessLevels: []string{"thorough"},
			},
			StrictMode: false,
		},
		Version: "1.0.0",
	}
}
