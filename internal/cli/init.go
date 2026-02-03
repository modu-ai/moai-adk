package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk-go/internal/cli/wizard"
	"github.com/modu-ai/moai-adk-go/internal/core/project"
	"github.com/modu-ai/moai-adk-go/internal/foundation"
	"github.com/modu-ai/moai-adk-go/internal/manifest"
	"github.com/modu-ai/moai-adk-go/internal/template"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new MoAI project",
	Long:  "Initialize a new MoAI project with Claude Code integration, including agents, skills, commands, hooks, and rules.",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().String("root", "", "Project root directory (default: current directory)")
	initCmd.Flags().String("name", "", "Project name (default: directory name)")
	initCmd.Flags().String("language", "", "Primary programming language")
	initCmd.Flags().String("framework", "", "Framework name (default: auto-detect or \"none\")")
	initCmd.Flags().String("username", "", "User display name")
	initCmd.Flags().String("conv-lang", "", "Conversation language code (e.g., \"en\", \"ko\")")
	initCmd.Flags().String("mode", "", "Development mode: ddd, tdd, or hybrid (default: auto-detect)")
	initCmd.Flags().String("git-mode", "", "Git workflow mode: manual, personal, or team (default: manual)")
	initCmd.Flags().String("github-username", "", "GitHub username (required for personal/team modes)")
	initCmd.Flags().String("git-commit-lang", "", "Git commit message language (default: en)")
	initCmd.Flags().String("code-comment-lang", "", "Code comment language (default: en)")
	initCmd.Flags().String("doc-lang", "", "Documentation language (default: en)")
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

// runInit executes the project initialization workflow.
func runInit(cmd *cobra.Command, _ []string) error {
	rootFlag := getStringFlag(cmd, "root")
	if rootFlag == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		rootFlag = cwd
	}

	nonInteractive := getBoolFlag(cmd, "non-interactive")

	opts := project.InitOptions{
		ProjectRoot:     rootFlag,
		ProjectName:     getStringFlag(cmd, "name"),
		Language:        getStringFlag(cmd, "language"),
		Framework:       getStringFlag(cmd, "framework"),
		UserName:        getStringFlag(cmd, "username"),
		ConvLang:        getStringFlag(cmd, "conv-lang"),
		DevelopmentMode: getStringFlag(cmd, "mode"),
		GitMode:         getStringFlag(cmd, "git-mode"),
		GitHubUsername:  getStringFlag(cmd, "github-username"),
		GitCommitLang:   getStringFlag(cmd, "git-commit-lang"),
		CodeCommentLang: getStringFlag(cmd, "code-comment-lang"),
		DocLang:         getStringFlag(cmd, "doc-lang"),
		NonInteractive:  nonInteractive,
		Force:           getBoolFlag(cmd, "force"),
	}

	// Run interactive wizard if not in non-interactive mode and running in a TTY
	if !nonInteractive && isatty.IsTerminal(os.Stdin.Fd()) {
		result, err := wizard.RunWithDefaults(rootFlag)
		if err != nil {
			if errors.Is(err, wizard.ErrCancelled) {
				fmt.Fprintln(cmd.OutOrStderr(), "Initialization cancelled.")
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
		if opts.DevelopmentMode == "" {
			opts.DevelopmentMode = result.DevelopmentMode
		}
		if opts.GitMode == "" {
			opts.GitMode = result.GitMode
		}
		if opts.GitHubUsername == "" {
			opts.GitHubUsername = result.GitHubUsername
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
	deployer := template.NewDeployer(embeddedFS)

	initializer := project.NewInitializer(deployer, mgr, nil)
	executor := project.NewPhaseExecutor(detector, methDetector, validator, initializer, nil)

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	result, err := executor.Execute(ctx, opts)
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "MoAI project initialized successfully.\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Development mode: %s\n", result.DevelopmentMode)
	fmt.Fprintf(cmd.OutOrStdout(), "  Created %d directories and %d files.\n", len(result.CreatedDirs), len(result.CreatedFiles))

	for _, w := range result.Warnings {
		fmt.Fprintf(cmd.OutOrStdout(), "  Warning: %s\n", w)
	}

	return nil
}
