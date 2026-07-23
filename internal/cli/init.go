package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/foundation"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// @MX:NOTE: [AUTO] CATALOG-002 REQ-021 4-substring informational notice — extracted as a helper so the contract is unit-testable without running the full runInit pipeline.
//
// emitSlimModeNotice prints the slim-mode informational notice to the given
// writer. The notice MUST contain the four substrings "slim mode", "--all",
// "MOAI_DISTRIBUTE_ALL=1", and "SPEC-V3R4-CATALOG-005" so downstream tooling
// (moai doctor) can pattern-match on them. See SPEC-V3R4-CATALOG-002 REQ-021
// and acceptance scenario S1.
func emitSlimModeNotice(out io.Writer) {
	_, _ = fmt.Fprintln(out,
		"Deploying core templates only (slim mode). "+
			"Use --all or MOAI_DISTRIBUTE_ALL=1 for full deploy. "+
			"Note: builder-harness agent is omitted (see SPEC-V3R4-CATALOG-005 for bootstrap).")
}

var initCmd = &cobra.Command{
	Use:     "init [project-name]",
	Short:   "Initialize a new MoAI project",
	GroupID: "project",
	Long: `Initialize a new MoAI project with Claude Code integration.

Usage patterns:
  moai init <project-name>   Create a new folder and initialize inside it
  moai init .                Initialize in current directory
  moai init                  Initialize in current directory (same as "moai init .")

Examples:
  moai init my-app           Creates ./my-app/ and initializes MoAI inside
  moai init .                Initializes MoAI in the current directory
  moai init --mode tdd       Initialize with specific development mode (default: tdd)
  moai init --all            Deploy all catalog entries (default is core-only slim mode; SPEC-V3R4-CATALOG-002)`,
	Args:    cobra.MaximumNArgs(1),
	PreRunE: validateInitFlags,
	RunE:    runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().String("root", "", "Project root directory (default: current directory)")
	initCmd.Flags().String("name", "", "Project name (default: directory name)")
	initCmd.Flags().String("language", "", "Primary programming language")
	initCmd.Flags().String("framework", "", "Framework name (default: auto-detect or \"none\")")
	initCmd.Flags().String("mode", "", "Development mode: ddd or tdd (default: tdd, auto-configured by /moai project)")
	initCmd.Flags().String("git-mode", "", "Git workflow mode: manual, personal, or team (default: manual)")
	initCmd.Flags().String("git-provider", "", "Git provider (github, gitlab)")
	initCmd.Flags().String("github-username", "", "GitHub username (required for personal/team modes)")
	initCmd.Flags().String("gitlab-instance-url", "", "GitLab instance URL for self-hosted")
	initCmd.Flags().Bool("non-interactive", false, "Skip interactive wizard; use flags and defaults")
	initCmd.Flags().Bool("force", false, "Reinitialize an existing project (backs up current .moai/)")
	initCmd.Flags().Bool("no-hooks", false, "Skip git hook installation (REQ-CIAUT-002)")
	initCmd.Flags().Bool("all", false, "Deploy all catalog entries (core + optional packs + harness-generated). Bypasses slim mode (SPEC-V3R4-CATALOG-002).")

	// Phase 1 mode flags (REQ-IWE-006, REQ-IWE-007)
	initCmd.Flags().Bool("standard", false, "Present Phase 1 questions (project mode, harness profile, LSP, quality gates, design)")
	initCmd.Flags().Bool("advanced", false, "Present Phase 1 + Phase 2 questions (implies --standard; Phase 2 skipped when prerequisites absent)")

	// Phase 1 non-interactive override flags (REQ-IWE-008)
	initCmd.Flags().String("project-mode", "", "Project mode: personal or team (default: personal)")
	initCmd.Flags().String("harness-profile", "", "Default harness evaluator profile: default, strict, lenient, frontend")
	initCmd.Flags().Bool("enable-lsp", false, "Enable LSP integration (default: false)")
	initCmd.Flags().Bool("enforce-quality", true, "Enforce quality gates (default: true)")
	initCmd.Flags().Bool("enable-design", true, "Enable design workflow (default: true)")

	// SPEC-AGENT-ARCH-V2-001 M3c (REQ-AA2-010): No-Haiku 3-tier performance
	// tier flag. New canonical name --model-policy max|medium|low; legacy
	// --high/--medium/--low accepted as deprecated aliases (one-cycle, plan.md D4).
	initCmd.Flags().String("model-policy", "", "Performance tier: max, medium, or low (persists to llm.yaml performance_tier)")
	initCmd.Flags().Bool("high", false, "Deprecated alias for --model-policy max (one-cycle backward compat)")
	initCmd.Flags().Bool("medium-alias", false, "Deprecated alias for --model-policy medium (one-cycle backward compat)")
	initCmd.Flags().Bool("low", false, "Deprecated alias for --model-policy low (one-cycle backward compat)")

	// SPEC-MODEL-PROFILE-MATRIX-001 (REQ-MPM-015): per-agent model+effort profile
	// selection. Persists to llm.profile; closed-set validated {max, medium, low}.
	// Takes precedence over the wizard answer. Supersedes the retired --plan-type.
	initCmd.Flags().String("profile", "", "Model+effort profile: max, medium, or low (persists to llm.yaml profile)")
}

// getStringFlag retrieves a string flag value from the command.
func getStringFlag(cmd *cobra.Command, name string) string {
	val, err := cmd.Flags().GetString(name)
	if err != nil {
		return ""
	}
	return val
}

// getBoolFlag retrieves a bool flag value from the command.
func getBoolFlag(cmd *cobra.Command, name string) bool {
	val, err := cmd.Flags().GetBool(name)
	if err != nil {
		return false
	}
	return val
}

// getBoolFlagWithDefault retrieves a bool flag value, returning defaultVal when
// the flag is not set or an error occurs.
func getBoolFlagWithDefault(cmd *cobra.Command, name string, defaultVal bool) bool {
	if !cmd.Flags().Changed(name) {
		return defaultVal
	}
	val, err := cmd.Flags().GetBool(name)
	if err != nil {
		return defaultVal
	}
	return val
}

// validateInitFlags validates flag values before execution.
func validateInitFlags(cmd *cobra.Command, _ []string) error {
	// Validate development mode
	mode := getStringFlag(cmd, "mode")
	if mode != "" {
		validModes := []string{"ddd", "tdd"}
		valid := slices.Contains(validModes, mode)
		if !valid {
			return fmt.Errorf("invalid --mode value %q: must be one of: ddd, tdd", mode)
		}
	}

	// Validate git workflow mode
	gitMode := getStringFlag(cmd, "git-mode")
	if gitMode != "" {
		validGitModes := []string{"manual", "personal", "team"}
		valid := slices.Contains(validGitModes, gitMode)
		if !valid {
			return fmt.Errorf("invalid --git-mode value %q: must be one of: manual, personal, team", gitMode)
		}
	}

	// Validate git provider
	gitProvider := getStringFlag(cmd, "git-provider")
	if gitProvider != "" {
		validProviders := []string{"github", "gitlab"}
		valid := slices.Contains(validProviders, gitProvider)
		if !valid {
			return fmt.Errorf("invalid --git-provider value %q: must be one of: github, gitlab", gitProvider)
		}
	}

	// SPEC-AGENT-ARCH-V2-001 M3c (REQ-AA2-010): validate --model-policy enum.
	// Invalid value exits non-zero with a stderr usage error naming the 3-enum.
	modelPolicy := getStringFlag(cmd, "model-policy")
	if modelPolicy != "" {
		validTiers := []string{"max", "medium", "low"}
		if !slices.Contains(validTiers, modelPolicy) {
			return fmt.Errorf("invalid --model-policy value %q: must be one of: max, medium, low", modelPolicy)
		}
	}

	// SPEC-MODEL-PROFILE-MATRIX-001 (REQ-MPM-015): validate --profile enum.
	// Invalid value exits non-zero with a usage error naming the closed set.
	profileFlag := getStringFlag(cmd, "profile")
	if profileFlag != "" && !config.IsValidProfile(profileFlag) {
		return fmt.Errorf("invalid --profile value %q: must be one of: max, medium, low", profileFlag)
	}

	// F3 git-provider identity validation (init-path parity with the
	// reconfigure path's validateWizardInput). Reuses the in-package helpers
	// from wizard_validate.go so a malformed username or a plaintext http URL
	// never lands in a persisted config. Empty values are permitted: the field
	// is simply not set.
	githubUsername := getStringFlag(cmd, "github-username")
	if githubUsername != "" && !isValidGitHubUsername(githubUsername) {
		return fmt.Errorf("invalid --github-username value %q: must be 1-%d characters, alphanumeric or single hyphens, with no leading or trailing hyphen",
			githubUsername, githubUsernameMaxLen)
	}

	gitlabInstanceURL := getStringFlag(cmd, "gitlab-instance-url")
	if gitlabInstanceURL != "" {
		if err := validateHTTPSURL(gitlabInstanceURL); err != nil {
			return fmt.Errorf("invalid --gitlab-instance-url value %q: %w", gitlabInstanceURL, err)
		}
	}

	return nil
}

// resolveModelPolicy resolves the effective performance tier from the
// --model-policy flag and its legacy aliases (--high/--medium-alias/--low).
// The new canonical flag takes precedence; legacy aliases map high→max,
// medium→medium, low→low (plan.md D4, one-cycle backward compat). Returns
// "" when no model-policy flag was set.
func resolveModelPolicy(cmd *cobra.Command) string {
	if mp := getStringFlag(cmd, "model-policy"); mp != "" {
		return mp
	}
	if getBoolFlag(cmd, "high") {
		return "max"
	}
	if getBoolFlag(cmd, "medium-alias") {
		return "medium"
	}
	if getBoolFlag(cmd, "low") {
		return "low"
	}
	return ""
}

// @MX:NOTE: [AUTO] CATALOG-002 REQ-012/013/EC3 — single decision point for slim/full opt-out. Narrow env matching: only "1" exact or case-insensitive "true".
//
// shouldDistributeAll returns true when the user has opted out of the slim
// init default by passing --all on the command line or setting the
// MOAI_DISTRIBUTE_ALL environment variable to "1" (exact) or "true"
// (case-insensitive). Any other value — including "0", "yes", "", or unset —
// returns false (slim mode remains active).
//
// Source: SPEC-V3R4-CATALOG-002 REQ-012, REQ-013, EC3.
func shouldDistributeAll(cmd *cobra.Command) bool {
	if all, _ := cmd.Flags().GetBool("all"); all {
		return true
	}
	v := os.Getenv("MOAI_DISTRIBUTE_ALL")
	return v == "1" || strings.EqualFold(v, "true")
}

// @MX:ANCHOR: [AUTO] runInit is the main entry point for project initialization
// @MX:REASON: [AUTO] fan_in=3, called from init.go init(), coverage_test.go, init_coverage_test.go
// runInit executes the project initialization workflow.
//
// The binary self-update check is DEFERRED: it starts only after the wizard
// has completed (and after the first phase output in non-interactive mode),
// is check-only (no install, no re-exec), and surfaces as a non-blocking
// stderr notice with the `moai update` hint at exit
// (SPEC-CLI-TUX-V3-002 REQ-TUX2-001..004; see init_update_notice.go).
func runInit(cmd *cobra.Command, args []string) error {
	// Unified output gateway: warnings and progress go to stderr, data to
	// stdout (SPEC-CLI-TUX-V3-001 REQ-CTX-012/016). The warning collector
	// wraps the printer so every Warn is re-emitted exactly once as a
	// consolidated stderr summary panel when init terminates — success or
	// failure (REQ-TUX2-013).
	p := newWarnCollector(printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr())))
	defer p.emitSummary(cmd.ErrOrStderr())

	// Git availability check (non-fatal warning)
	if _, err := exec.LookPath("git"); err != nil {
		p.Warn("git is not installed. Some features (plan/run/sync workflows, branch management) will be limited.\n  %s",
			GitInstallHint())
	}

	rootFlag := getStringFlag(cmd, "root")
	projectName := getStringFlag(cmd, "name")

	// Determine project root based on positional argument
	// - moai init <name>  → create ./name/ directory
	// - moai init .       → use current directory
	// - moai init         → use current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if rootFlag != "" {
		// --root flag takes precedence
		// Keep rootFlag as-is
	} else if len(args) > 0 && args[0] != "." {
		// Positional argument provided (not ".")
		// Create new folder with that name
		targetDir := args[0]
		// Use filepath.Abs to correctly handle both absolute and relative paths.
		// filepath.Join(cwd, absPath) incorrectly prepends cwd to absolute paths,
		// e.g. Join("/a/b", "/c/d") = "/a/b/c/d" instead of "/c/d".
		absTarget, err := filepath.Abs(targetDir)
		if err != nil {
			return fmt.Errorf("resolve project path %q: %w", targetDir, err)
		}
		rootFlag = absTarget

		// Create the directory if it doesn't exist
		if err := os.MkdirAll(rootFlag, 0755); err != nil {
			return fmt.Errorf("create project directory %q: %w", targetDir, err)
		}

		// Use the directory name as project name if not specified
		if projectName == "" {
			projectName = targetDir
		}
	} else {
		// No positional arg or "." - use current directory
		rootFlag = cwd
	}

	nonInteractive := getBoolFlag(cmd, "non-interactive")

	// Resolve mode flags: --advanced implies --standard (REQ-IWE-007, EC-3)
	advancedMode := getBoolFlag(cmd, "advanced")
	standardMode := getBoolFlag(cmd, "standard") || advancedMode

	opts := project.InitOptions{
		ProjectRoot:       rootFlag,
		ProjectName:       projectName,
		Language:          getStringFlag(cmd, "language"),
		Framework:         getStringFlag(cmd, "framework"),
		DevelopmentMode:   getStringFlag(cmd, "mode"),
		GitMode:           getStringFlag(cmd, "git-mode"),
		GitProvider:       getStringFlag(cmd, "git-provider"),
		GitHubUsername:    getStringFlag(cmd, "github-username"),
		GitLabInstanceURL: getStringFlag(cmd, "gitlab-instance-url"),
		NonInteractive:    nonInteractive,
		Force:             getBoolFlag(cmd, "force"),
		// SPEC-MODEL-PROFILE-MATRIX-001 (REQ-MPM-015/016): --profile flag value
		// (validated in validateInitFlags). The wizard fills opts.Profile only when
		// the flag is absent, so the flag takes precedence over the wizard answer.
		Profile: getStringFlag(cmd, "profile"),
		// Phase 1 mode flags
		StandardMode: standardMode,
		// Phase 1 non-interactive overrides — defaults match wizard defaults (REQ-IWE-008)
		ProjectMode:               getStringFlag(cmd, "project-mode"),
		HarnessProfile:            getStringFlag(cmd, "harness-profile"),
		LSPEnabled:                getBoolFlag(cmd, "enable-lsp"),
		EnforceQuality:            getBoolFlagWithDefault(cmd, "enforce-quality", true),
		CoverageExemptionsEnabled: false, // no CLI flag; wizard/default only
		DesignEnabled:             getBoolFlagWithDefault(cmd, "enable-design", true),
		ClaudeDesignEnabled:       true, // wizard-only; default true
	}

	// Git mode + provider are auto-detected from the repository's configured
	// remotes rather than asked in the wizard (no remote → manual, ≥1 remote →
	// personal; provider from the origin host). Explicit --git-mode /
	// --git-provider flags take precedence: detection only fills a value still
	// empty after flag parsing.
	if opts.GitMode == "" || opts.GitProvider == "" {
		detectedMode, detectedProvider := detectGitConfig(rootFlag)
		if opts.GitMode == "" {
			opts.GitMode = detectedMode
		}
		if opts.GitProvider == "" {
			opts.GitProvider = detectedProvider
		}
	}

	// Apply user-level defaults from profile preferences.
	// Profile preferences (identity, languages, model policy) are set via
	// "moai profile setup" and stored in ~/.moai/claude-profiles/<name>/preferences.yaml.
	profileName := profile.GetCurrentName()

	// Auto-prompt profile setup if no profile exists yet
	if !nonInteractive && isatty.IsTerminal(os.Stdin.Fd()) && !profile.IsSetup(profileName) {
		var wantSetup bool
		confirm := huh.NewConfirm().
			Title("No profile found. Set up profile preferences now?").
			Description("Configure your name, language, and model preferences.").
			Value(&wantSetup)
		if err := confirm.Run(); err == nil && wantSetup {
			if err := runProfileSetup(cmd, nil); err != nil {
				p.Warn("profile setup failed: %v", err)
			}
		}
	}

	prefs, err := profile.ReadPreferences(profileName)
	if err != nil {
		p.Warn("failed to read profile preferences: %v", err)
	} else {
		if prefs.UserName != "" {
			opts.UserName = prefs.UserName
		}
		if prefs.ConversationLang != "" {
			opts.ConvLang = prefs.ConversationLang
		}
		if prefs.GitCommitLang != "" {
			opts.GitCommitLang = prefs.GitCommitLang
		}
		if prefs.CodeCommentLang != "" {
			opts.CodeCommentLang = prefs.CodeCommentLang
		}
		if prefs.DocLang != "" {
			opts.DocLang = prefs.DocLang
		}
		// Model policy is now configured per-project via the init wizard.
		// Profile-level model policy is no longer applied here.
	}

	if !nonInteractive && isInteractiveStdin() {
		// Print banner and welcome message
		uikit.PrintBanner(version.GetVersion())
		uikit.PrintWelcomeMessage()

		// Use RunWithDefaultsModes when --standard or --advanced is set; otherwise
		// fall back to RunWithDefaults for Quick mode backward-compat (REQ-IWE-006).
		// runWizardFn is the injectable wizard seam (REQ-TUX2-001 order contract).
		// The profile locale + user name pre-fill the conversation_language and
		// user_name question defaults (empty when no profile exists).
		result, wizErr := runWizardFn(rootFlag, opts.ConvLang, opts.UserName, standardMode, advancedMode)
		if wizErr != nil {
			if errors.Is(wizErr, wizard.ErrCancelled) {
				_, _ = fmt.Fprintln(cmd.OutOrStderr(), "Initialization cancelled.")
				return nil
			}
			return fmt.Errorf("wizard failed: %w", wizErr)
		}

		// Conversation language + user name: the wizard answer wins over the
		// profile value. Update both opts (drives template deployment of
		// language.yaml/user.yaml) AND prefs (SyncToProjectConfig runs after
		// deployment and would otherwise re-apply the stale profile value over
		// the wizard choice). Empty answers leave the profile fallback intact.
		if result.ConversationLang != "" {
			opts.ConvLang = result.ConversationLang
			prefs.ConversationLang = result.ConversationLang
		}
		if result.UserName != "" {
			opts.UserName = result.UserName
			prefs.UserName = result.UserName
		}

		// Apply wizard results to opts (wizard values override empty flags)
		if opts.ProjectName == "" {
			opts.ProjectName = result.ProjectName
		}
		if opts.DevelopmentMode == "" {
			opts.DevelopmentMode = result.DevelopmentMode
		}
		// Git mode, provider, and credentials are NOT wizard-supplied on the
		// init path: mode/provider come from remote detection above, and the
		// remaining Git values stay flag-fed (--github-username,
		// --gitlab-instance-url). The reconfigure wizard still asks them.
		if result.ModelPolicy != "" {
			opts.ModelPolicy = result.ModelPolicy
		}
		// SPEC-MODEL-PROFILE-MATRIX-001 (REQ-MPM-014/016): the model-routing wizard
		// answer IS the profile selection; it flows through opts.ModelPolicy and is
		// normalized to {max, medium, low} at profile persistence (the --profile flag
		// takes precedence over it).
		// Report format is wizard-only (no CLI flag); empty resolves to the
		// html+md default at persistence time (initializer.writeReportConfig).
		if opts.ReportFormat == "" && result.ReportFormat != "" {
			opts.ReportFormat = result.ReportFormat
		}
		// Apply Phase 1 wizard results (only when StandardMode was active)
		if result.StandardMode {
			if result.ProjectMode != "" {
				opts.ProjectMode = result.ProjectMode
			}
			if result.HarnessProfile != "" {
				opts.HarnessProfile = result.HarnessProfile
			}
			opts.LSPEnabled = result.LSPEnabled
			opts.EnforceQuality = result.EnforceQuality
			opts.CoverageExemptionsEnabled = result.CoverageExemptionsEnabled
			opts.DesignEnabled = result.DesignEnabled
			opts.ClaudeDesignEnabled = result.ClaudeDesignEnabled
		}
	}

	// Default git provider to "github" for backward compatibility
	if opts.GitProvider == "" {
		opts.GitProvider = "github"
	}

	// development_mode is no longer an interactive wizard question; TDD is the
	// silent default. The --mode flag still overrides (opts.DevelopmentMode was
	// set from the flag above). Setting a concrete value here — rather than
	// leaving it empty — makes phase.go take the explicitly-set validated branch
	// and skip methodology auto-detection (which would otherwise fall back to
	// ddd on an empty value).
	if opts.DevelopmentMode == "" {
		opts.DevelopmentMode = "tdd"
	}

	// Build dependencies
	registry := foundation.DefaultRegistry
	detector := project.NewDetector(registry, nil)
	methDetector := project.NewMethodologyDetector(nil)
	validator := project.NewValidator(nil)
	mgr := manifest.NewManager()

	// Wire embedded template deployer (REQ-E-030)
	// Load catalog manifest (SPEC-V3R4-CATALOG-001) for slim/full deploy routing (CATALOG-002).
	cat, catErr := template.LoadEmbeddedCatalog()
	if catErr != nil {
		return fmt.Errorf("CATALOG_LOAD_FAILED: %w", catErr)
	}

	// Both paths share the same renderer baseline (raw embedded FS).
	embeddedFS, err := template.EmbeddedTemplates()
	if err != nil {
		return fmt.Errorf("load embedded templates: %w", err)
	}
	renderer := template.NewRenderer(embeddedFS)

	var deployer template.Deployer
	if shouldDistributeAll(cmd) {
		deployer = template.NewDeployerWithRenderer(embeddedFS, renderer)
	} else {
		var slimErr error
		deployer, slimErr = template.NewSlimDeployerWithRenderer(cat, renderer)
		if slimErr != nil {
			return fmt.Errorf("CATALOG_LOAD_FAILED: slim deployer: %w", slimErr)
		}
		// REQ-021 informational notice on slim mode (4 substring guarantee).
		emitSlimModeNotice(cmd.OutOrStdout())
	}

	initializer := project.NewInitializer(deployer, mgr, nil)
	executor := project.NewPhaseExecutor(detector, methDetector, validator, initializer, nil)

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Use printer-backed console output for progress reporting
	// (REQ-CTX-015: ProgressReporter events route through the Printer to
	// stderr; the former project.ConsoleReporter wrote to stdout). The
	// spinner-backed reporter renders template deployment as an animated
	// live line on a TTY and degrades to plain lines otherwise
	// (REQ-TUX2-010/011).
	executor.SetReporter(newSpinnerReporter(p))

	p.Info("Initializing MoAI project...")

	// Deferred binary self-update check (REQ-TUX2-001/004): starts strictly
	// after wizard completion and after the first phase output; never blocks
	// phase execution; flushed as a stderr notice at exit. A wizard cancel
	// returns before this point, so the cancel path has zero network side
	// effects (acceptance.md §C).
	flushUpdateNotice := startDeferredUpdateNotice(cmd)

	result, err := executor.Execute(ctx, opts)
	if err != nil {
		// REQ-TUX2-015: re-running init on an initialized project without
		// --force is usually a template-refresh intent — redirect to
		// `moai update` alongside the existing --force guidance.
		if !getBoolFlag(cmd, "force") && strings.Contains(err.Error(), "already initialized") {
			return fmt.Errorf("initialization failed: %w\n  Hint: this directory already contains a MoAI project — did you mean 'moai update' (refresh templates in place)? Re-run with --force only to reinitialize from scratch", err)
		}
		return fmt.Errorf("initialization failed: %w", err)
	}

	// Route executor result warnings into the collector (they surface once,
	// in the exit summary panel — REQ-TUX2-013) and display the completion
	// card with the next-action sequence (REQ-TUX2-016). Human-facing status
	// belongs on stderr (internal/cli/CLAUDE.md Output streams; REQ-CTX-012).
	for _, w := range result.Warnings {
		p.Collect(w)
	}
	cardName := opts.ProjectName
	if cardName == "" {
		cardName = filepath.Base(opts.ProjectRoot)
	}
	_, _ = fmt.Fprintln(cmd.ErrOrStderr())
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(),
		buildInitSuccessCard(cardName, len(result.CreatedDirs), len(result.CreatedFiles), p.Count()))

	// Sync profile preferences to project config (after template deployment)
	if err := profile.SyncToProjectConfig(opts.ProjectRoot, prefs); err != nil {
		p.Warn("Failed to sync profile to project config: %v", err)
	}

	// SPEC-AGENT-ARCH-V2-001 M3c (REQ-AA2-010): persist the resolved
	// performance tier to llm.yaml. CLI --model-policy takes precedence over
	// the wizard's ModelPolicy; both resolve to one of {max, medium, low}.
	perfTier := resolveModelPolicy(cmd)
	if perfTier == "" && opts.ModelPolicy != "" {
		perfTier = opts.ModelPolicy
	}
	if perfTier != "" && template.IsValidPerformanceTier(perfTier) {
		if err := template.ApplyPerformanceTier(opts.ProjectRoot, perfTier); err != nil {
			p.Warn("Failed to apply performance tier: %v", err)
		}
	}

	// SPEC-MODEL-PROFILE-MATRIX-001 (REQ-MPM-016): persist the resolved per-agent
	// profile to llm.profile. Precedence: the --profile flag (opts.Profile, already
	// validated to {max, medium, low}), else the resolved model-policy tier
	// (perfTier / opts.ModelPolicy). NormalizeToTier is total (high→max, ""→medium),
	// so the wizard's legacy {high, medium, low} answer maps correctly.
	{
		profile := opts.Profile
		if profile == "" {
			profile = perfTier
		}
		if profile == "" {
			profile = opts.ModelPolicy
		}
		if resolved := template.NormalizeToTier(profile); resolved != "" {
			if err := template.ApplyProfile(opts.ProjectRoot, resolved); err != nil {
				p.Warn("Failed to apply profile: %v", err)
			}
		}
	}

	// Scaffold .moai/evolution/ directory structure (R2: Directory Scaffolding).
	// This is also handled by template deployment, but scaffoldEvolutionDir ensures
	// all required subdirectories and placeholder files are present even when the
	// template fs doesn't track empty directories.
	if err := deploy.ScaffoldEvolutionDir(opts.ProjectRoot); err != nil {
		p.Warn("Failed to scaffold evolution directory: %v", err)
	}

	// Ensure global settings.json has required env variables
	if err := ensureGlobalSettingsEnv(); err != nil {
		p.Warn("Failed to update global settings env: %v", err)
	}

	// Install pre-push hook (REQ-CIAUT-002). Non-fatal; --no-hooks opts out.
	// Status/warning lines are human-facing -> stderr (REQ-CTX-016).
	installPrePushHookOptional(opts.ProjectRoot, getBoolFlag(cmd, "no-hooks"), cmd.ErrOrStderr())

	// Install pre-commit hook (REQ-PC-001). Fast-subset commit tier; --no-hooks opts out.
	installPreCommitHookOptional(opts.ProjectRoot, getBoolFlag(cmd, "no-hooks"), cmd.ErrOrStderr())

	// Deferred self-update notice (REQ-TUX2-002): non-blocking stderr notice
	// with the `moai update` hint; a failed or in-flight check never affects
	// the init result.
	flushUpdateNotice(p)

	return nil
}
