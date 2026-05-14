package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/pkg/version"
)

var rootCmd = &cobra.Command{
	Use:   "moai",
	Short: "MoAI-ADK: Agentic Development Kit for Claude Code",
	Long: `MoAI-ADK (Go Edition) is a high-performance development toolkit
that serves as the runtime backbone for the MoAI framework within Claude Code.

It provides CLI tooling, configuration management, LSP integration,
Git operations, quality gates, and autonomous development loop capabilities.

Use 'moai cc', 'moai cg', or 'moai glm' to launch Claude Code.`,
	Version: version.GetVersion(),
	Run: func(cmd *cobra.Command, args []string) {
		PrintBanner(version.GetVersion())
		_ = cmd.Help()
	},
}

// @MX:ANCHOR: [AUTO] Execute is the main entry point for the moai CLI
// @MX:REASON: [AUTO] fan_in=3, called from cmd/moai/main.go, root_test.go, integration_test.go
// Execute initializes dependencies and runs the root command.
func Execute() error {
	initConsole()
	InitDependencies()
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("moai-adk %s\n", version.GetVersion()))

	// M6-S5: Install tui-based help renderer for rootCmd.
	// SetHelpFunc applies to "moai --help" and "moai help" (cobra's built-in help subcommand).
	// Subcommand-level help (e.g. "moai doctor --help") still uses cobra's default template.
	rootCmd.SetHelpFunc(renderRootHelp)

	// Command groups
	rootCmd.AddGroup(
		&cobra.Group{ID: "launch", Title: "Launch Commands:"},
		&cobra.Group{ID: "project", Title: "Project Commands:"},
		&cobra.Group{ID: "tools", Title: "Tools:"},
	)

	// Wire worktree subcommand with lazy Git initialization
	worktree.WorktreeCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if deps == nil {
			return fmt.Errorf("dependencies not initialized")
		}
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		if err := deps.EnsureGit(cwd); err != nil {
			return fmt.Errorf("initialize git: %w", err)
		}
		worktree.WorktreeProvider = deps.GitWorktree
		return nil
	}

	// Register worktree subcommand tree
	rootCmd.AddCommand(worktree.WorktreeCmd)

	// Register statusline command
	rootCmd.AddCommand(StatuslineCmd)

	// ASTG-UPGRADE-001: register astgrep command
	rootCmd.AddCommand(NewAstGrepCmd())

	// SPEC-TELEMETRY-001: register telemetry subcommand
	rootCmd.AddCommand(telemetryCmd)

	// SPEC-V3R2-CON-001: register constitution subcommand
	rootCmd.AddCommand(newConstitutionCmd())

	// SPEC-V3R2-RT-004: register state subcommand
	rootCmd.AddCommand(newStateCmd())

	// SPEC-V3R2-RT-004 REQ-031: register clean subcommand
	rootCmd.AddCommand(newCleanCmd())

	// SPEC-V3R2-RT-007: register migration subcommand group
	rootCmd.AddCommand(migrationCmd)

	// NOTE: newHarnessCmd is intentionally NOT registered per SPEC-V3R4-HARNESS-001
	// (BC-V3R4-HARNESS-001-CLI-RETIREMENT). The harness CLI verb path is retired;
	// all lifecycle verbs (status / apply / rollback / disable) are owned by the
	// /moai:harness slash command + skill workflow surface (no Go binary invocation).
	// See internal/cli/harness_retirement_test.go for the CI guard test that
	// asserts the absence of a "harness" subcommand registration.
}
