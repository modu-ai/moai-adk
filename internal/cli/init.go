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

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/foundation"
	"github.com/modu-ai/moai-adk/internal/homestate"
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
//
// Common skills and agents install into user folders; optional bundles require
// an explicit selection and do not follow the project template --all switch.
func emitSlimModeNotice(out io.Writer) {
	_, _ = fmt.Fprintln(out,
		"Deploying core templates only (slim mode). "+
			"Use --all or MOAI_DISTRIBUTE_ALL=1 for all project harness templates. "+
			"Note: builder-harness agent is omitted (see SPEC-V3R4-CATALOG-005 for bootstrap).")
	_, _ = fmt.Fprintln(out,
		"Common skills and agents install into your user folders. "+
			"Use --bundles to opt into optional user assets.")
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
  moai init --all            Select all project harness templates; optional user assets require --bundles

Note: moai init / moai update do NOT auto-enter a worktree. To work inside a
worktree, enter one with the launcher flag (moai cc -w <name>).`,
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
	initCmd.Flags().Bool("no-plugin", false, "Deprecated compatibility flag; init always deploys locally and installs common assets into user folders")
	initCmd.Flags().Bool("all", false, "Select all project harness templates; optional user assets still require --bundles")
	// SPEC-USER-ASSET-INSTALL-001 (REQ-004, iter2 D18): the initial opt-in
	// bundle selection, recorded in the per-user manifest. L0 installs
	// regardless; each named bundle adds its catalog entries.
	initCmd.Flags().String("bundles", "", "Comma-separated opt-in bundle names to install into your user folders on first init (e.g. devops,frontend)")

	// The two wizard mode flags are retired (REQ-WIZ-018): the wizard presents
	// the same three pages to every user, so there is no mode to select.

	// Page-3 non-interactive override flags (REQ-IWE-008).
	// (--project-mode was removed by SPEC-INIT-UPDATE-CONSISTENCY-001 REQ-ICU-001:
	// project.mode had no Go reader.)
	// Registered with the same true default the help text states and the wizard's
	// lsp_enabled seed uses; explicitness is probed with Changed() (see
	// getBoolFlagWithDefault), so --enable-lsp=false still wins.
	initCmd.Flags().Bool("enable-lsp", true, "Enable LSP integration (default: true)")
	initCmd.Flags().Bool("enforce-quality", true, "Enforce quality gates (default: true)")
	initCmd.Flags().Bool("enable-design", true, "Enable design workflow (default: true)")

	// Retired per-agent model flags (SPEC-AGENT-MODEL-INHERIT-001 D10/D13).
	// Subagents inherit the main session's model and effort, so these flags
	// are accepted for script compatibility, print a deprecation warning, and
	// have no effect. The main-session policy lives in `moai profile setup`.
	initCmd.Flags().String("model-policy", "", deprecatedAgentModelFlagUsage)
	initCmd.Flags().Bool("high", false, deprecatedAgentModelFlagUsage)
	initCmd.Flags().Bool("medium-alias", false, deprecatedAgentModelFlagUsage)
	initCmd.Flags().Bool("low", false, deprecatedAgentModelFlagUsage)
	initCmd.Flags().String("profile", "", deprecatedAgentModelFlagUsage)

	// SPEC-WT-DOC-001 workflow toggle flags. Each ships default-off and uses an
	// opt-in tracker (see applyWorkflowBranchGuardFlags) so a flag-absent init
	// leaves the deployed template default untouched.
	initCmd.Flags().Bool("branch-guard", false, "Enable the main-checkout branch-state guard (workflow.branch_guard.enabled)")
	initCmd.Flags().Bool("worktree-auto-create", false, "Enable worktree auto-creation (workflow.worktree.auto_create)")
	initCmd.Flags().Bool("worktree-auto-merge", false, "Enable worktree auto-merge (workflow.worktree.auto_merge)")
	initCmd.Flags().Bool("worktree-auto-cleanup", false, "Enable worktree auto-cleanup (workflow.worktree.auto_cleanup)")

	// SPEC-AUTONOMY-TIERS-001 M4 (AC-001 / AC-006): the autonomy-tier selector.
	// Validates the 3-value closed set fail-loud in validateInitFlags; help
	// names all 3 tiers so the selector OFFERS them (does not pre-pick
	// fully-autonomous).
	initCmd.Flags().String("autonomy-tier", "", "Session permission mode: accept edits on (semi-auto, default), auto mode (automatic), or bypass permissions (fully-autonomous; requires sandbox proof). Writes user-scope defaultMode: acceptEdits for the default")

	// The LLM harness selector is claude-only since the general codex
	// execution surface removal (card t1909): init deploys and wires the
	// claude harness alone. The flag stays for script compatibility and is
	// validated fail-loud in validateInitFlags.
	initCmd.Flags().String("llm", "", "LLM harness to deploy and wire: claude (default and only value — the general codex execution surface was removed)")
}

// agentWiring is the harness selection label. The closed deployment set is
// claude-only (card t1909 removed the general codex execution and init
// surface); gpt and both survive solely as LEGACY RECORD values an interrupted
// pre-removal init may have left under .moai/state/resume-wiring, which
// readResumeWiring still recognizes so the resume can refuse a harness
// mismatch instead of mis-wiring the tree.
type agentWiring string

const (
	agentWiringClaude agentWiring = "claude"
	agentWiringGPT    agentWiring = "gpt"
	agentWiringBoth   agentWiring = "both"
)

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

// provisionMCPEntryUnlessDeclined writes the single neutral `moai` entry into
// the project's .mcp.json unless the caller declined it (default-on per
// SPEC-MCP-DEFAULT-ON-001). On `moai init` the decline comes from the harness
// selection (codex) or from the non-interactive path, never from a wizard
// question.
//
// It is the call site that makes the MCP server reachable at runtime: without
// it the provisioning default is collected into opts and then dropped, leaving
// the stdio server registered in code but absent from every host's config.
//
// The write goes through the shared atomic-config seam
// (provisionMoaiMCPServerEntryAt -> mutateClaudeJSONAtomic), so it inherits the
// same lock + backup + idempotent-skip behaviour the other entry writers use.
// Provisioning is best-effort: a failure is RETURNED, not printed — the caller
// records it into the warning collector and the exit summary panel renders it
// exactly once (card t1527 D5 + repair round: one surface per failure), so a
// broken or unwritable config can never fail an init. The user's explicit
// decline is honored absolutely (C-A-5): default-on is a default, not a
// mandate.
func provisionMCPEntryUnlessDeclined(out io.Writer, projectRoot string, declined bool) error {
	if declined {
		return nil
	}
	configPath := filepath.Join(projectRoot, ".mcp.json")
	if err := provisionMoaiMCPServerEntryAt(configPath); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(out, "Provisioned the moai MCP server entry in .mcp.json (default-on).")
	return nil
}

// applyWizardPage3ToOpts applies the wizard result's fixed Page-3 seeds to
// opts, honouring flag-over-wizard precedence (REQ-WIZ-020).
//
// The seeds reach opts on every interactive run, so without this rule the
// wizard result would unconditionally overwrite every flag-seeded value —
// inverting the documented `--profile` precedence above ("the wizard fills
// opts.Profile only when the flag is absent"). Each of the three overlapping
// settings therefore yields to the wizard ONLY when its flag was not
// explicitly supplied.
//
// The page-3 questions removed by SPEC-INIT-QUIET-WIZARD-001 (project mode,
// worktree auto-create, todo, feedback, continuation, audit, MCP) carry no
// result field, so they are not mapped here: their keys resolve to shipped
// defaults. --worktree-auto-create still reaches opts through the flag path;
// --project-mode was removed outright by SPEC-INIT-UPDATE-CONSISTENCY-001
// REQ-ICU-001 (project.mode had no Go reader).
//
// Explicitness is probed with cmd.Flags().Changed(name), never by value:
// getBoolFlag / getBoolFlagWithDefault cannot distinguish "flag absent" from
// "flag explicitly set to the same value as the default", so a value-only
// check would silently drop `--enable-lsp=false` and `--enforce-quality=false`.
func applyWizardPage3ToOpts(cmd *cobra.Command, result *wizard.WizardResult, opts *project.InitOptions) {
	if !cmd.Flags().Changed("enable-lsp") {
		opts.LSPEnabled = result.LSPEnabled
	}
	if !cmd.Flags().Changed("enforce-quality") {
		opts.EnforceQuality = result.EnforceQuality
	}
	if !cmd.Flags().Changed("enable-design") {
		opts.DesignEnabled = result.DesignEnabled
	}

	// coverage_exemptions_enabled is not asked (REQ-WIZ-013) and has no CLI
	// flag: opts.CoverageExemptionsEnabled keeps the false the runInit literal
	// seeds. Card t1966: the inert result→opts copy is removed; a restored
	// question must reintroduce the mapping with its flag-precedence rule.

	// claude_design_enabled is wizard-only (no CLI flag), so it always applies.
	opts.ClaudeDesignEnabled = result.ClaudeDesignEnabled
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

// alreadyInitializedError is the shared refusal for a re-run of init on an
// initialized project: the pre-wizard probe and the executor's validation
// report it identically.
func alreadyInitializedError(cause error) error {
	return fmt.Errorf("initialization failed: %w\n  Hint: this directory already contains a MoAI project — did you mean 'moai update' (refresh templates in place)? Re-run with --force only to reinitialize from scratch", cause)
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

	// Retired per-agent model flags: any value is accepted and only warned
	// about (SPEC-AGENT-MODEL-INHERIT-001 D10/D13).
	warnDeprecatedAgentModelFlags(cmd, "profile", "model-policy", "high", "medium-alias", "low")

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

	// SPEC-AUTONOMY-TIERS-001 M4: validate the autonomy-tier closed set
	// fail-loud (REQ-001 / AP-5). Empty is valid (defaults to semi-auto
	// downstream, AC-007). Reuses config.ValidateAutonomyTierSelection so the
	// selector and the reader share one normalization + error vocabulary.
	autonomyTier := getStringFlag(cmd, "autonomy-tier")
	if autonomyTier != "" {
		if _, err := config.ValidateAutonomyTierSelection(autonomyTier); err != nil {
			return fmt.Errorf("invalid --autonomy-tier value: %w", err)
		}
	}

	// The --llm closed set is claude-only (card t1909): the gpt/both/codex
	// deployment paths are gone, so any other value fails loud here rather
	// than silently deploying claude.
	llm := getStringFlag(cmd, "llm")
	if llm != "" && llm != string(agentWiringClaude) {
		return fmt.Errorf("invalid --llm value %q: must be claude — moai init deploys the claude harness only", llm)
	}

	// B7: validate --bundles names EARLY — a typo must fail before the deploy,
	// not after the project is already initialized (where the re-run refuses
	// with 'already initialized' and the missing install has no path forward).
	if raw := getStringFlag(cmd, "bundles"); strings.TrimSpace(raw) != "" {
		cat, catErr := template.LoadEmbeddedCatalog()
		if catErr != nil {
			return fmt.Errorf("load catalog for bundle validation: %w", catErr)
		}
		var unknown []string
		for _, name := range parseBundleSelection(raw) {
			if _, ok := cat.Catalog.OptionalPacks[name]; !ok {
				unknown = append(unknown, name)
			}
		}
		if len(unknown) > 0 {
			return fmt.Errorf("invalid --bundles value: unknown bundle(s) %s — valid bundles: run 'moai bundle add --help' or check catalog.yaml", strings.Join(unknown, ", "))
		}
	}

	return nil
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

// resolveInitDeployMode resolves the run's deploy mode. SPEC-USER-ASSET-
// INSTALL-001 (REQ-017, D4): the plugin carrier is retired — init NEVER
// invokes a plugin marketplace or install command, and the deploy is the
// local (slim) payload in every case; the common skills and agents install
// into the user folders (REQ-005). M7 removes the mode split entirely.
func resolveInitDeployMode(cmd *cobra.Command) template.DeployMode {
	return template.DeployModeLocal
}

// ensureUserAssetsLockedFn is ensureUserAssetsLocked behind a package-level
// variable so a test can fail or interrupt the user-asset ensure, which runs before
// executor.Execute (card t1587). It follows the applyAutonomyTierBundleFn pattern
// (update_settings_snapshot.go). Tests that replace it must not call t.Parallel().
var ensureUserAssetsLockedFn = ensureUserAssetsLocked

// initWindowFn is the post-executor interruption seam (card t1619): it runs once the
// resume record and the init marker are on disk, before the post-steps. Nil in
// production. Tests that replace it must not call t.Parallel().
var initWindowFn func(projectRoot string) error

// @MX:ANCHOR: [AUTO] runInit is the main entry point for project initialization
// @MX:REASON: [AUTO] fan_in=3, called from init.go init(), coverage_test.go, init_coverage_test.go
// runInit executes the project initialization workflow.
//
// The binary self-update check is DEFERRED: it starts only after the wizard
// has completed (and after the first phase output in non-interactive mode),
// is check-only (no install, no re-exec), and surfaces as a non-blocking
// stderr notice with the `moai update` hint at exit
// (SPEC-CLI-TUX-V3-002 REQ-TUX2-001..004; see init_update_notice.go).
func runInit(cmd *cobra.Command, args []string) (err error) {
	// Unified output gateway: warnings and progress go to stderr, data to
	// stdout (SPEC-CLI-TUX-V3-001 REQ-CTX-012/016). The warning collector
	// wraps the printer so every Warn is re-emitted exactly once as a
	// consolidated stderr summary panel when init terminates — success or
	// failure (REQ-TUX2-013).
	p := newWarnCollector(printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr())))
	defer p.emitSummary(cmd.ErrOrStderr())

	// SPEC-SESSION-WORKTREE-001 M2/M4: session-worktree auto-entry + exit
	// disposal. The auto-entry runs BEFORE any shared-state mutation
	// (REQ-SW-002). When the feature is OFF (default) the wrapper returns ""
	// and the rest of runInit is byte-identical to the baseline (REQ-SW-001).
	// On success the process chdir's into the worktree (BI-2 / R1 mitigation).
	// On fail-back (REQ-SW-004) or already-in-worktree skip (REQ-SW-012) the
	// wrapper returns "" and init continues unchanged in the shared checkout.
	//
	// M4 (REQ-SW-008/009/010): the deferred cleanup disposes the materialized
	// worktree at subcommand exit. It honors three cases (default-manual
	// persist / clean-exit remove / non-clean preserve) + the dirty guard.
	// cleanExit is derived from this function's named error return (err ==
	// nil means exit 0). wtPath == "" makes cleanup a no-op.
	swCfg := loadSessionWorktreeConfig(cmd)
	wtPath := enterSessionWorktree(swCfg, "init", cmd.ErrOrStderr())
	if wtPath != "" {
		if cerr := os.Chdir(wtPath); cerr != nil {
			// Chdir failure is a fail-back (REQ-SW-004): continue in the
			// shared checkout. The worktree was materialized but unusable
			// from this process; the user can enter it manually later.
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
				"moai: session-worktree chdir into %s failed (%v); continuing in shared checkout for %q\n",
				wtPath, cerr, "init")
		}
	}
	defer func() {
		// SPEC-WORKTREE-KEY-WIRING-001 M2: auto-merge runs BEFORE disposal —
		// the merge consumes only committed state, but disposal deletes the
		// tree, so merge-then-dispose is the only safe order. Independent of
		// auto_cleanup (REQ-WKW-013): each toggle gates only its own behavior.
		sessionExitAutoMerge(swCfg, wtPath, err == nil, cmd.ErrOrStderr())
		cleanupSessionWorktree(swCfg, wtPath, err == nil, cmd.ErrOrStderr())
	}()

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
		// Page-3 non-interactive overrides — defaults match wizard defaults (REQ-IWE-008).
		// The InitOptions mode field is gone (C33): the Page-3 writes are
		// unconditional now, so there is no mode to carry into the initializer.
		LSPEnabled:                getBoolFlagWithDefault(cmd, "enable-lsp", true),
		EnforceQuality:            getBoolFlagWithDefault(cmd, "enforce-quality", true),
		CoverageExemptionsEnabled: false, // no CLI flag; wizard/default only
		DesignEnabled:             getBoolFlagWithDefault(cmd, "enable-design", true),
		ClaudeDesignEnabled:       true, // wizard-only; default true
	}

	// Chain ② flag→opts link (SPEC-INIT-WIZARD-REPAIR-001 REQ-004): the four
	// workflow toggle flags, applied BEFORE the wizard block so the
	// flag-over-wizard precedence in applyWizardPage3ToOpts sees the explicit
	// selection. Tracker-gated: an absent flag leaves opts untouched
	// (distributed-default preservation).
	// @MX:SPEC: SPEC-INIT-WIZARD-REPAIR-001
	applyWorkflowBranchGuardFlags(cmd, &opts)

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

	// REQ-ITI-001: `moai init` carries NO profile entry — no confirmation, no
	// profile wizard, whatever stdin and the flags are. A missing profile just
	// leaves the preference values empty, and the init wizard asks the
	// conversation language once as its first question (REQ-ITI-002). The
	// profile wizard starts only from `moai profile setup` / `--setup`.
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

	// wizardResult carries the wizard's answers out of the interactive block
	// so the autonomy-tier apply below covers BOTH entry paths with a single
	// call: the flag branch for non-interactive runs and the wizard branch for
	// interactive ones (it stays empty when the wizard did not run).
	wizardResult := &wizard.WizardResult{}
	// wizardRan distinguishes "the user declined" from "nobody was asked".
	// Both leave the boolean answers at their zero value, and one of them must
	// not be written over a persisted choice (see applyJevFromWizard).
	wizardRan := false

	// Card t1619 (codex review P2): the resume route is decided before the ensure moves any
	// state. The ensure reconciles and clears the pending-install journal and fails on a
	// corrupt user manifest, so a checkpoint read after it would miss the interruption it
	// exists to detect, and the resume's quarantine would never run. A run with a checkpoint
	// skips the ensure: the resume runs it under the user lock, after the quarantine.
	// Card t1966: read before the wizard so the already-initialized probe below can use it.
	resuming := !getBoolFlag(cmd, "force") && initResumeCheckpoint(userHomeDirOrEmpty(), opts.ProjectRoot)

	// Card t1966: the already-initialized probe runs BEFORE the wizard. An initialized
	// project that is neither forced nor resumable gets the `moai update` guidance
	// before any question is asked, instead of after the answers are collected. The
	// executor's own validation stays the authority for every other case (the probe
	// only fires on positive evidence, so a validator error falls through to it).
	if !nonInteractive && isInteractiveStdin() && !getBoolFlag(cmd, "force") && !resuming {
		if vr, vErr := project.NewValidator(nil).Validate(opts.ProjectRoot); vErr == nil && !vr.Valid &&
			slices.ContainsFunc(vr.Errors, func(e string) bool { return strings.Contains(e, "already initialized") }) {
			return alreadyInitializedError(project.ErrProjectExists)
		}
	}

	if !nonInteractive && isInteractiveStdin() {
		// Print banner and welcome message
		uikit.PrintBanner(version.GetVersion())
		uikit.PrintWelcomeMessage()

		// One wizard entry point for every user: the mode flags that used to
		// select between two question sets are retired (REQ-WIZ-018).
		// runWizardFn is the injectable wizard seam (REQ-TUX2-001 order contract).
		// The profile locale + user name pre-fill the conversation_language and
		// user_name question defaults (empty when no profile exists).
		result, wizErr := runWizardFn(rootFlag, opts.ConvLang, opts.UserName)
		if wizErr != nil {
			if errors.Is(wizErr, wizard.ErrCancelled) {
				_, _ = fmt.Fprintln(cmd.OutOrStderr(), "Initialization cancelled.")
				return nil
			}
			return fmt.Errorf("wizard failed: %w", wizErr)
		}
		wizardResult = result
		wizardRan = true

		// Conversation language + user name: the wizard answer wins over the
		// profile value. Update opts (drives template deployment of
		// language.yaml/user.yaml). The profile sync after deployment takes the same
		// answers from wizardResult (syncInitProfile), so the healthy tail and the
		// resume dispatch cannot disagree about them. Empty answers leave the profile
		// fallback intact.
		if result.ConversationLang != "" {
			opts.ConvLang = result.ConversationLang
		}
		if result.UserName != "" {
			opts.UserName = result.UserName
		} else if result.UserNameCleared {
			// Card t1966: an explicit clear drops the profile-supplied name.
			opts.UserName = ""
		}

		// Project name, model policy, report format, and development mode are
		// not asked on the init path (SPEC-INIT-QUIET-WIZARD-001): the project
		// name falls to the directory-name default, the model profile to medium,
		// the report format to html+md, and the development mode to tdd (the
		// --mode flag is its only override; see below) — the same values a user
		// accepting every former default received. Card t1966: the former
		// result.DevelopmentMode mapping was dead (never asked) and is removed.
		// Git mode, provider, and credentials are NOT wizard-supplied on the
		// init path: mode/provider come from remote detection above, and the
		// remaining Git values stay flag-fed (--github-username,
		// --gitlab-instance-url). The reconfigure wizard still asks them.

		// Apply the fixed Page-3 seeds. The former mode gate on the wizard
		// result is removed (REQ-WIZ-001/002), so they always reach opts.
		applyWizardPage3ToOpts(cmd, result, &opts)

		// The interactive path provisions the .mcp.json moai entry by default:
		// the mcp_provision question is gone, so no wizard answer can decline
		// it. The non-interactive path leaves the zero value false and skips
		// the ensure-entry call — an intended asymmetry that preserves the
		// pinned non-interactive behaviour (REQ-IQW-006). The harness switch
		// below still overrides this default (codex declines, both forces on).
		// @MX:NOTE: [AUTO] Interactive default-on MCP provisioning with no
		// wizard question; non-interactive stays false by design.
		// @MX:SPEC: SPEC-INIT-QUIET-WIZARD-001
		opts.MCPProvision = true
	}

	// Chain ① wizard+flag → opts link (SPEC-INIT-WIZARD-REPAIR-001 REQ-001 /
	// REQ-002): applyAutonomyTierFromWizard covers BOTH entry paths — the
	// --autonomy-tier flag branch for non-interactive runs (closing the
	// validated-then-discarded gap: the value used to be checked in
	// validateInitFlags and then dropped) and the wizard branch for interactive
	// ones (wizardResult stays empty when the wizard did not run). Flag wins
	// over the wizard answer; the fully-autonomous gates are applied on the
	// wizard path and re-applied downstream on the flag path.
	// @MX:SPEC: SPEC-INIT-WIZARD-REPAIR-001
	applyAutonomyTierFromWizard(
		cmd.Flags().Changed("autonomy-tier"), getStringFlag(cmd, "autonomy-tier"),
		wizardResult, &opts,
	)

	// The harness selection is claude-only (card t1909): the flag is validated
	// to claude above and the wizard no longer asks, so every consumer below
	// reads this constant.
	agentWiringSelection := agentWiringClaude

	// SPEC-INIT-SHRINK-001 (REQ-001/REQ-003/REQ-007, OD-5/OD-7): resolve the
	// deploy mode ONCE, here — the initializer's scaffold (below), the
	// deployer option, and the mode record all read this local.
	deployMode := resolveInitDeployMode(cmd)

	// SPEC-INIT-HARNESS-001 (REQ-IH-005) named the codex suppression here; the
	// codex selection is gone, so the initializer always writes the claude
	// surface.
	opts.Harness = string(agentWiringSelection)
	// SPEC-INIT-SHRINK-001 (REQ-001): the initializer skips the dropped-root
	// scaffold shells on the plugin path, matching the deployer's file set.
	opts.DeployMode = string(deployMode)

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
	// The claude harness is the only init deployment (card t1909 removed the
	// gpt/both arms). The harnessFS reroute of SPEC-INIT-HARNESS-001 M2 and
	// the dual-harness deployers are gone with them; the update path still
	// reads llm.harness for projects initialized before the removal.
	//
	// SPEC-INIT-SHRINK-001 M2 (REQ-001/REQ-003/REQ-006/REQ-007): the deploy
	// mode rides through as an option — local (opt-out or --all) is today's
	// payload. SPEC-USER-ASSET-INSTALL-001 (M7): no deploy-mode options — the
	// split is retired with the plugin carrier.
	if shouldDistributeAll(cmd) {
		deployer, err = template.NewClaudeHarnessDeployerWithRenderer(cat, renderer)
	} else {
		deployer, err = template.NewClaudeHarnessSlimDeployerWithRenderer(cat, renderer)
		emitSlimModeNotice(cmd.OutOrStdout())
	}
	if err != nil {
		return fmt.Errorf("construct %s harness deployer: %w", agentWiringSelection, err)
	}

	initializer := project.NewInitializer(deployer, mgr, nil)
	executor := newInitPhaseExecutorFn(detector, methDetector, validator, initializer)

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

	// SPEC-USER-ASSET-INSTALL-001 (REQ-024): the first-install trigger. Init
	// ensures every L0 and opted-in-bundle asset is present user-side before
	// the run reports success — the judgment is PER-ASSET-STATE applying the
	// REQ-023 truth table exactly as update does (a user-edited file's bytes
	// are never clobbered; an untracked target is a REQ-010 collision; a
	// partial install's manifest does not suppress the run). Systemic
	// failures fail the init; per-file failures continue and surface in the
	// summary (REQ-013).
	//
	// Card t1587: this phase runs BEFORE executor.Execute records the project
	// as initialized. It reads only the embedded catalog and writes only under
	// the user home — no project dependency (`moai bundle add` runs the same
	// ensure standalone) — so sequencing it first is safe. Placed after the
	// deploy it made the failure advice a dead end: a systemic ensure failure
	// (~/.codex/agents as a regular file — observed twice on the t1561 lane)
	// aborted a run whose project was ALREADY recorded initialized, so the
	// error's own "re-run 'moai init'" advice hit the 'project already
	// initialized' refusal. Failing before the marker keeps the promise
	// honest: the re-run starts from an uninitialized project and completes
	// cleanly. Card t1591: the resume record and the init marker are NOT written here.
	// The marker would create .moai/ before executor.Execute, and validator.go:130 reads
	// that directory as "already initialized", so the retry of a failed ensure would be
	// refused. They open after the executor instead (the block below the layout step).
	// A run with a checkpoint (resuming, read before the wizard) skips this ensure.
	if !resuming {
		if homeDir, homeErr := userHomeDirFn(); homeErr == nil {
			selection := parseBundleSelection(getStringFlag(cmd, "bundles"))
			if err := ensureUserAssetsLockedFn(homeDir, selection, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("user-asset install failed: %w\n  Fix the cause and re-run 'moai init' — the ensure is idempotent and completes the shortfall", err)
			}
		} else {
			p.Warn("Could not resolve the user home; the user-folder asset install was skipped: %v", homeErr)
		}
	}

	// Deferred binary self-update check (REQ-TUX2-001/004): starts strictly
	// after wizard completion and after the first phase output; never blocks
	// phase execution; flushed as a stderr notice at exit. A wizard cancel
	// returns before this point, so the cancel path has zero network side
	// effects (acceptance.md §C).
	flushUpdateNotice := startDeferredUpdateNotice(cmd)

	// SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001 (REQ-USB-005): settle a
	// .claude/settings.json staging copy an interrupted earlier flow left behind
	// before executor.Execute — init's first step that can rewrite the file.
	//
	// @MX:WARN: [AUTO] leftover judgement placement — keep above executor.Execute
	// @MX:REASON: after the deploy the live file is the new render, so an abort with no revert
	// would be discarded instead of promoted (plan.md B8)
	// Finding 3 (card t1591): a resumed init keeps the staging copy its interrupted first
	// run took. The judgement discards a copy whose live file has changed since, and the
	// interrupted run's own autonomy answer changes it, so a resume skips the judgement
	// and the copy is promoted by the resume's settle once every post-step completed.
	// --force reinitializes from scratch and is judged as before.
	if !resuming {
		backup.JudgeLeftoverSettingsSnapshot(opts.ProjectRoot, cmd.ErrOrStderr())
	}
	// Card t1029: same judgement for the .mcp.json staging copy.
	backup.JudgeLeftoverMCPSnapshot(opts.ProjectRoot, cmd.ErrOrStderr())

	// SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001 (REQ-TBS-001) as corrected by card
	// t1139: capture the template snapshot — the NEXT update's merge BASE — at
	// the one moment the section files hold the pure template render: right
	// after the deploy, before the initializer's section patches (lsp.enabled
	// and the other Page-3 answers) and before the post-Execute writers below
	// (profile sync, Jev, performance tier, profile, harness). Taken later, the
	// snapshot records those answers as BASE and the first update resets them
	// to the template default. Best-effort non-blocking (REQ-TBS-014).
	errOut := cmd.ErrOrStderr()
	opts.AfterTemplateDeploy = func(root string) {
		writeTemplateSnapshotBestEffort(root, errOut)
	}
	result, err := executor.Execute(ctx, opts)
	if err != nil {
		// SPEC-INIT-DEPLOY-EXIT-001 (REQ-IDE-004): warnings the executor
		// recorded before failing (the skill-mirror notice) still belong in the
		// exit summary panel — the failure path is not a reason to drop them.
		// The executor returns a result alongside the error where it has one.
		if result != nil {
			for _, w := range result.Warnings {
				p.Collect(w)
			}
		}
		// REQ-TUX2-015: re-running init on an initialized project without
		// --force is usually a template-refresh intent — redirect to
		// `moai update` alongside the existing --force guidance.
		// M6 (SPEC: init resume, REQ-SRF-007 + audit DEBT R4): when the
		// PRIOR attempt died at the user-asset ensure (initialized but
		// incomplete), the re-run RESUMES instead of refusing: the ensure
		// shortfall completes, and the setup steps the failed attempt never
		// reached (ApplyHarness, the MCP entry, the Codex wiring) run now.
		// DEBT R4: the resume is judged by CONTENT — the harness config the
		// resume writes is asserted, not merely the directory's existence.
		// Gate round 35-3: the resume is GATED on an explicit interruption
		// checkpoint — a pending-install journal (interrupted mid-install)
		// or a corrupt user manifest (failed before/at the ensure). A
		// healthy project's ordinary re-run keeps the update redirect.
		if !getBoolFlag(cmd, "force") && strings.Contains(err.Error(), "already initialized") {
			// M6: the resume fires only on an interruption checkpoint
			// (gate round 35-3) — a healthy project's ordinary re-run
			// keeps the original update redirect below.
			if resuming {
				if resumeErr := resumeInitializedProject(cmd, &opts, wizardRan, wizardResult, prefs); resumeErr != nil {
					return fmt.Errorf("initialization resume failed: %w\n  Hint: this directory already contains a MoAI project — 'moai update' refreshes templates in place; --force reinitializes from scratch", resumeErr)
				}
				p.Info("Initialized MoAI project (resumed: user-asset shortfall completed, setup steps finished).")
				flushUpdateNotice(p)
				return nil
			}
			return alreadyInitializedError(err)
		}
		// SPEC-INIT-DEPLOY-EXIT-001 (REQ-IDE-003): a deployment failure aborts
		// the template walk partway, so files after the failing one were never
		// written. The surface states that fact and does NOT quantify it: the
		// only number available here is how many files were written before the
		// abort, while the expected total exists nowhere in the code — a bare
		// count would read as progress rather than as damage, and a denominator
		// would have to be invented. The failing template path is already
		// carried by the wrapped error, so it is preserved rather than restated.
		if strings.Contains(err.Error(), "template deployment") {
			return fmt.Errorf("initialization failed: %w\n  The project tree is INCOMPLETE — template deployment stopped at the failing template, so the files after it were never written. Do not use this directory as-is: fix the template error, then re-run 'moai init --force' to reinitialize from scratch", err)
		}
		return fmt.Errorf("initialization failed: %w", err)
	}
	// SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001 (REQ-USB-001/003/016): stage the
	// settings.json render the deploy inside executor.Execute wrote — only when
	// the manifest proves it wrote it (a skipped existing file records nothing)
	// — before the autonomy tier bundle below can rewrite the file.
	backup.StageDeployedSettingsSnapshot(opts.ProjectRoot, mgr, cmd.ErrOrStderr())
	// Card t1029: the same staging for .mcp.json, before any later provisioning
	// write (moai mcp add, the mcp-server opt-in) can rewrite the file — a
	// post-deploy addition belongs to the user and must read as theirs on the
	// next update, not as template content.
	backup.StageDeployedMCPSnapshot(opts.ProjectRoot, mgr, cmd.ErrOrStderr())

	if err := homestate.EnsureProjectLayout(opts.ProjectRoot); err != nil {
		return fmt.Errorf("initialize private MoAI home layout: %w", err)
	}

	// Finding 1 (card t1591): record the harness this run selected once the executor has
	// initialized the project, before the post-steps below. The resume reads the record only
	// under a checkpoint, so it need not precede the executor (see the ensure above).
	// Best-effort: a record that cannot be written is named, never dropped silently.
	if recErr := persistResumeWiring(opts.ProjectRoot, agentWiringSelection); recErr != nil {
		emitSeverityLine(cmd.ErrOrStderr(), sevWarn, resolveTheme(), "Could not record the harness selection for the resume: %v", recErr)
	}
	// Finding B (card t1591): open the post-executor window with the init marker. An
	// interruption in the post-steps below leaves this marker, and the rerun reads it as the
	// interruption checkpoint. Cleared at the end of runInit. Best-effort, like the record.
	if pendErr := persistInitPending(opts.ProjectRoot, parseBundleSelection(getStringFlag(cmd, "bundles"))); pendErr != nil {
		emitSeverityLine(cmd.ErrOrStderr(), sevWarn, resolveTheme(), "Could not record the init resume marker: %v", pendErr)
	}

	// Card t1619 (re-adjusted placement): the post-executor interruption seam. Nil in
	// production; the resume tests inject a failure here, which leaves the state a real
	// interruption after the executor leaves (project initialized, record and marker on disk).
	if initWindowFn != nil {
		if err := initWindowFn(opts.ProjectRoot); err != nil {
			return fmt.Errorf("initialization interrupted after the executor: %w", err)
		}
	}

	// Chain ① consumer link (SPEC-INIT-WIZARD-REPAIR-001 REQ-003): wire the
	// persisted tier selection into the deployed settings immediately after
	// the initializer returns. Paths are resolved here and passed in (no new
	// global state); the USER-scope write inside the bundle is a key-scoped
	// splice limited to the permissions block (spec.md §4 leader ruling). The
	// call is best-effort: semi-auto/unset produces the bounded delta (the
	// USER-scope acceptEdits record only, SPEC-AUT-PERMMODES-001 REQ-004) and
	// a failure warns without failing the init.
	// @MX:SPEC: SPEC-INIT-WIZARD-REPAIR-001
	if homeDir, homeErr := userHomeDirFn(); homeErr == nil {
		if tierErr := applyAutonomyTierBundleFn(
			opts.ProjectRoot,
			filepath.Join(homeDir, ".claude", "settings.json"),
			filepath.Join(opts.ProjectRoot, ".claude", "settings.json"),
			opts.AutonomyTier,
		); tierErr != nil {
			p.Warn("Failed to apply autonomy tier bundle: %v", tierErr)
		}
	} else {
		p.Warn("Failed to resolve home directory for autonomy tier bundle: %v", homeErr)
	}

	// SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001 (REQ-USB-005): init has no merge,
	// so it never takes a preserve path — the staged render (if any) becomes the
	// canonical base the first update merges against.
	backup.SettleSettingsSnapshot(opts.ProjectRoot, false, cmd.ErrOrStderr())
	backup.SettleMCPSnapshot(opts.ProjectRoot, false, cmd.ErrOrStderr())

	// Route executor result warnings into the collector (they surface once,
	// in the exit summary panel — REQ-TUX2-013). Human-facing status belongs
	// on stderr (internal/cli/CLAUDE.md Output streams; REQ-CTX-012).
	// Card t1527 D5: the completion card itself MOVED below the tail — the
	// profile/Jev/harness/hooks/MCP tail steps used to print their lines after
	// the card, so the card did not read as the end of the run.
	for _, w := range result.Warnings {
		p.Collect(w)
	}
	// Sync profile preferences to project config (after template deployment). The
	// helper is shared with the resume dispatch, so the two paths cannot drift (t1591 Item D).
	if err := syncInitProfile(opts.ProjectRoot, prefs, wizardResult); err != nil {
		p.Warn("Failed to sync profile to project config: %v", err)
	}

	// SPEC-JEV-OPTIN-MEASURE-001 (REQ-JEVO-002): route the wizard's Jev answer
	// into the SAME internal/settings seam the `moai web` settings screen
	// drives. Placed after template deployment because the seam patches
	// .moai/config/sections/workflow.yaml — the file the deploy has just
	// written. Best-effort, like its autonomy-tier and performance-tier
	// neighbours: a failure warns rather than failing the init, because the
	// capability is off by default and a project that could not record an
	// opt-in is a project with the capability off.
	if err := applyJevFromWizard(wizardRan, wizardResult, opts.ProjectRoot); err != nil {
		p.Warn("Failed to persist the Jev opt-in: %v", err)
	}

	// SPEC-FEEDBACK-PARTICIPATION-001 (REQ-ANON-003): route the wizard's
	// participation answer into the user-scoped consent file the update ask
	// and the console also drive. The bridge gates on wizardRan and an empty
	// CI environment; its own failure warns rather than failing the init, for
	// the same reason the Jev bridge's does — the capability is off by
	// default, and a consent that could not be recorded is no consent.
	if err := applyParticipationFromWizard(wizardRan, wizardResult, opts.ProjectRoot); err != nil {
		p.Warn("Failed to persist the participation opt-in: %v", err)
	}

	// SPEC-INIT-HARNESS-001 (REQ-IH-002): persist the RESOLVED harness value to
	// llm.harness on every init run — all three closed-set values INCLUDING the
	// claude default. agentWiringSelection is already the single resolution
	// (SPEC-INIT-HARNESS-PROMPT-001 REQ-IHP-004: flag > wizard > claude), so the
	// persisted value can never disagree with what the deployment below did.
	// Explicit record over implicit absence: doctor (REQ-IH-011) and update
	// re-deployment (REQ-IH-010) read the key instead of inferring claude.
	if err := template.ApplyHarness(opts.ProjectRoot, string(agentWiringSelection)); err != nil {
		p.Warn("Failed to apply harness: %v", err)
	}

	// SPEC-INIT-SHRINK-001 (REQ-009, OD-5 settled (a)): persist the resolved
	// deploy mode beside the harness value, on every init run including
	// re-init — plugin when the default deploy ran, local under the opt-out
	// or --all. Update reads this key (REQ-016) instead of inferring the
	// mode; the update flow's restore step re-asserts it from the pre-update
	// backup so the .moai/config Clean wipe cannot cost it.
	if err := template.ApplyDeployMode(opts.ProjectRoot, string(deployMode)); err != nil {
		p.Warn("Failed to apply deploy mode: %v", err)
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
	// Status/warning lines are human-facing -> stderr (REQ-CTX-016). Card
	// t1527 D5: an install failure reaches the warning collector, so the
	// terminal summary panel carries it.
	if pushErr := installPrePushHookOptional(opts.ProjectRoot, getBoolFlag(cmd, "no-hooks"), cmd.ErrOrStderr(), cmd.ErrOrStderr()); pushErr != nil {
		p.Warn("Pre-push hook installation failed: %v", pushErr)
	}

	// Install pre-commit hook (REQ-PC-001). Fast-subset commit tier; --no-hooks opts out.
	if commitErr := installPreCommitHookOptional(opts.ProjectRoot, getBoolFlag(cmd, "no-hooks"), cmd.ErrOrStderr(), cmd.ErrOrStderr()); commitErr != nil {
		p.Warn("Pre-commit hook installation failed: %v", commitErr)
	}

	// SPEC-WORKTREE-BRANCH-GUARD-001 (REQ-WBG-009): surface the shared-checkout
	// worktree advisory. Phrased per workflow.worktree.auto_create; rides stdout
	// alongside the slim-mode notice (informational, not a gate).
	emitWorktreeAdvisory(cmd.OutOrStdout(), opts.ProjectRoot)

	// SPEC-MCP-DEFAULT-ON-001 (default-on, REQ-A-3): turn opts.MCPProvision
	// into the single neutral .mcp.json entry. The interactive path sets it
	// true with no wizard question (SPEC-INIT-QUIET-WIZARD-001 REQ-IQW-005);
	// the non-interactive path leaves it false and skips the ensure-entry call
	// (REQ-IQW-006). The codex/both MCP-provisioning overrides of
	// SPEC-CODEX-WIRING-001 D3 are gone with the harness selection (card
	// t1909), so only the wizard answer decides mcpDeclined.
	// @MX:SPEC: SPEC-INIT-QUIET-WIZARD-001
	// The returned failure joins the warning collector exactly once.
	mcpDeclined := !opts.MCPProvision
	if mcpErr := provisionMCPEntryUnlessDeclined(cmd.OutOrStdout(), opts.ProjectRoot, mcpDeclined); mcpErr != nil {
		p.Collect("MCP server entry provisioning failed: " + mcpErr.Error())
	}

	// Deferred self-update notice (REQ-TUX2-002): non-blocking stderr notice
	// with the `moai update` hint; a failed or in-flight check never affects
	// the init result. Card t1527 repair round: this runs BEFORE the
	// completion card, so the card is followed only by the deferred warning
	// summary panel — the terminal surface by design (REQ-TUX2-013: the
	// collector re-emits every warning exactly once when the run terminates;
	// the card's own pointer text reads "see the warning summary on stderr
	// below", which is only true with the panel last).
	flushUpdateNotice(p)

	// card t1277: every post-deploy rewrite above (WritePhase1Configs patching
	// lsp/quality/design, ApplyHarness rewriting llm.yaml) happens AFTER the
	// deploy tracked the rendered sections, so the manifest saves the
	// pre-answer hashes and the next init --force reads the drifted files as
	// user edits. Re-record the section hashes here — the LAST writer wins, so
	// one retrack at the tail covers the whole family. template_managed-only
	// filtering keeps user-owned entries untouched (two-way invariant).
	// Card t1527 repair round 3: a retrack FAILURE routes into the warning
	// collector (per-file skip notes still stream to stderr — bookkeeping
	// noise, not the failure verdict). Card t1527 repair round 4: the whole
	// block runs BEFORE the completion card, so the card's warning count and
	// its "see the warning summary" hint include the retrack failure, and
	// nothing but the deferred summary panel ever follows the card.
	if retrackErr := retrackSectionFiles(opts.ProjectRoot, cmd.ErrOrStderr()); retrackErr != nil {
		p.Collect("manifest retrack (config sections) failed: " + retrackErr.Error())
	}

	// Card t1527 D5: the completion card prints after every tail step above —
	// its "initialized" verdict is the last thing the operator reads before
	// the collected warning summary panel (see the comment above). The card's
	// warning count is taken HERE, after every collector source above.
	cardName := opts.ProjectName
	if cardName == "" {
		cardName = filepath.Base(opts.ProjectRoot)
	}
	_, _ = fmt.Fprintln(cmd.ErrOrStderr())
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(),
		buildInitSuccessCard(cardName, len(result.CreatedDirs), len(result.CreatedFiles), p.Count(), string(deployMode)))

	// The template snapshot is written by opts.AfterTemplateDeploy (set before
	// executor.Execute), not here: by this point the section files carry the
	// wizard answers, and a snapshot of them is not a template render (t1139).

	// Finding B (card t1591): every post-ensure step above is warn-only, so reaching
	// this point means the ensure window closed with its steps complete. The init
	// marker retires here.
	clearInitPending(opts.ProjectRoot)
	// Finding 1 (card t1591): the selection record retires on normal completion,
	// with the checkpoint marker and only once the marker is gone. A marker that
	// still stands needs its selection for the resume it names.
	if _, statErr := os.Stat(filepath.Join(opts.ProjectRoot, ".moai", "resume-pending")); os.IsNotExist(statErr) {
		clearResumeWiring(opts.ProjectRoot)
	}

	return nil
}
