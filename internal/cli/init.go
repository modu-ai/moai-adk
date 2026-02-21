package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/foundation"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new MoAI project",
	Long: `Initialize a new MoAI project with Claude Code integration.

Usage patterns:
  moai init <project-name>   Create a new folder and initialize inside it
  moai init .                Initialize in current directory
  moai init                  Initialize in current directory (same as "moai init .")

Examples:
  moai init my-app           Creates ./my-app/ and initializes MoAI inside
  moai init .                Initializes MoAI in the current directory
  moai init --mode tdd       Initialize with specific development mode (default: tdd)`,
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
	initCmd.Flags().String("username", "", "User display name")
	initCmd.Flags().String("conv-lang", "", "Conversation language code (e.g., \"en\", \"ko\")")
	initCmd.Flags().String("mode", "", "Development mode: ddd or tdd (default: tdd, auto-configured by /moai project)")
	initCmd.Flags().String("git-mode", "", "Git workflow mode: manual, personal, or team (default: manual)")
	initCmd.Flags().String("git-provider", "", "Git provider (github, gitlab)")
	initCmd.Flags().String("github-username", "", "GitHub username (required for personal/team modes)")
	initCmd.Flags().String("gitlab-instance-url", "", "GitLab instance URL for self-hosted")
	initCmd.Flags().String("git-commit-lang", "", "Git commit message language (default: en)")
	initCmd.Flags().String("code-comment-lang", "", "Code comment language (default: en)")
	initCmd.Flags().String("doc-lang", "", "Documentation language (default: en)")
	initCmd.Flags().String("model-policy", "", "Agent model policy: high, medium, or low (default: high)")
	initCmd.Flags().Bool("non-interactive", false, "Skip interactive wizard; use flags and defaults")
	initCmd.Flags().Bool("force", false, "Reinitialize an existing project (backs up current .moai/)")
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

	// Validate model policy
	modelPolicy := getStringFlag(cmd, "model-policy")
	if modelPolicy != "" {
		validPolicies := []string{"high", "medium", "low"}
		valid := slices.Contains(validPolicies, modelPolicy)
		if !valid {
			return fmt.Errorf("invalid --model-policy value %q: must be one of: high, medium, low", modelPolicy)
		}
	}

	// Validate conversation language (ISO 639-1 codes)
	convLang := getStringFlag(cmd, "conv-lang")
	if convLang != "" {
		validLangs := []string{"en", "ko", "ja", "zh", "es", "fr", "de", "pt", "ru", "it"}
		valid := slices.Contains(validLangs, convLang)
		if !valid {
			return fmt.Errorf("invalid --conv-lang value %q: must be a valid ISO 639-1 language code (e.g., en, ko, ja, zh)", convLang)
		}
	}

	return nil
}

// @MX:ANCHOR: [AUTO] runInit is the main entry point for project initialization
// @MX:REASON: [AUTO] fan_in=3, called from init.go init(), coverage_test.go, init_coverage_test.go
// runInit executes the project initialization workflow.
// It first checks for a binary update so the latest templates are used.
func runInit(cmd *cobra.Command, args []string) error {
	// Binary update step (non-fatal)
	if !shouldSkipBinaryUpdate(cmd) {
		updated, err := runBinaryUpdateStep(cmd)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: binary update check failed: %v\n", err)
		}
		if updated {
			if err := reexecNewBinary(); err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: failed to re-exec new binary: %v\n", err)
			}
			// reexecNewBinary replaces the process on success; only
			// reach here if it failed.
		}
	}

	// Git availability check (non-fatal warning)
	if _, err := exec.LookPath("git"); err != nil {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(),
			"Warning: git is not installed. Some features (plan/run/sync workflows, branch management) will be limited.\n  %s\n",
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
		UserName:          getStringFlag(cmd, "username"),
		ConvLang:          getStringFlag(cmd, "conv-lang"),
		DevelopmentMode:   getStringFlag(cmd, "mode"),
		GitMode:           getStringFlag(cmd, "git-mode"),
		GitProvider:       getStringFlag(cmd, "git-provider"),
		GitHubUsername:    getStringFlag(cmd, "github-username"),
		GitLabInstanceURL: getStringFlag(cmd, "gitlab-instance-url"),
		GitCommitLang:     getStringFlag(cmd, "git-commit-lang"),
		CodeCommentLang:   getStringFlag(cmd, "code-comment-lang"),
		DocLang:           getStringFlag(cmd, "doc-lang"),
		ModelPolicy:       getStringFlag(cmd, "model-policy"),
		NonInteractive:    nonInteractive,
		Force:             getBoolFlag(cmd, "force"),
	}

	if !nonInteractive && isatty.IsTerminal(os.Stdin.Fd()) {
		// Print banner and welcome message
		PrintBanner(version.GetVersion())
		PrintWelcomeMessage()

		result, err := wizard.RunWithDefaults(rootFlag)
		if err != nil {
			if errors.Is(err, wizard.ErrCancelled) {
				_, _ = fmt.Fprintln(cmd.OutOrStderr(), "Initialization cancelled.")
				return nil
			}
			return fmt.Errorf("wizard failed: %w", err)
		}

		// Apply wizard results to opts (wizard values override empty flags)
		if opts.ProjectName == "" {
			opts.ProjectName = result.ProjectName
		}
		if opts.UserName == "" {
			opts.UserName = result.UserName
		}
		if opts.ConvLang == "" {
			opts.ConvLang = result.Locale
		}
		// DevelopmentMode defaults to "tdd" via template context;
		// auto-configured later by /moai project workflow based on project analysis.
		if opts.GitMode == "" {
			opts.GitMode = result.GitMode
		}
		if opts.GitProvider == "" {
			opts.GitProvider = result.GitProvider
		}
		if opts.GitHubUsername == "" {
			opts.GitHubUsername = result.GitHubUsername
		}
		if opts.GitLabInstanceURL == "" {
			opts.GitLabInstanceURL = result.GitLabInstanceURL
		}
		if opts.GitCommitLang == "" {
			opts.GitCommitLang = result.GitCommitLang
		}
		if opts.CodeCommentLang == "" {
			opts.CodeCommentLang = result.CodeCommentLang
		}
		if opts.DocLang == "" {
			opts.DocLang = result.DocLang
		}
		if opts.ModelPolicy == "" && result.ModelPolicy != "" {
			opts.ModelPolicy = result.ModelPolicy
		}
	}

	// Default git provider to "github" for backward compatibility
	if opts.GitProvider == "" {
		opts.GitProvider = "github"
	}

	// Build dependencies
	registry := foundation.DefaultRegistry
	detector := project.NewDetector(registry, nil)
	methDetector := project.NewMethodologyDetector(nil)
	validator := project.NewValidator(nil)
	mgr := manifest.NewManager()

	// Wire embedded template deployer (REQ-E-030)
	embeddedFS, err := template.EmbeddedTemplates()
	if err != nil {
		return fmt.Errorf("load embedded templates: %w", err)
	}

	// Create renderer for template processing
	renderer := template.NewRenderer(embeddedFS)
	deployer := template.NewDeployerWithRenderer(embeddedFS, renderer)

	initializer := project.NewInitializer(deployer, mgr, nil)
	executor := project.NewPhaseExecutor(detector, methDetector, validator, initializer, nil)

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Use simple console output for progress reporting
	consoleReporter := project.NewConsoleReporter()
	executor.SetReporter(consoleReporter)

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Initializing MoAI project...") //nolint:errcheck

	result, err := executor.Execute(ctx, opts)
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	// Display success message
	details := []string{
		renderKeyValueLines([]kvPair{
			{"Directories", fmt.Sprintf("%d created", len(result.CreatedDirs))},
			{"Files", fmt.Sprintf("%d created", len(result.CreatedFiles))},
		}),
	}
	for _, w := range result.Warnings {
		details = append(details, cliWarn.Render("Warning: "+w))
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout())
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderSuccessCard("MoAI project initialized", details...))

	// Ensure global settings.json has required env variables
	if err := ensureGlobalSettingsEnv(); err != nil {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Warning: Failed to update global settings env: %v\n", err)
	}

	return nil
}
