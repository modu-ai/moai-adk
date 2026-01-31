package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root CLI command
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "moai",
		Short: "MoAI Strategic Orchestrator for Claude Code",
		Long: `MoAI is the Strategic Orchestrator for Claude Code. It provides project initialization,
template management, configuration handling, and CLI command infrastructure for AI-powered
application development.`,
		Version: "dev",
	}

	// Add subcommands
	rootCmd.AddCommand(
		NewInitCommand(),
		NewDoctorCommand(),
		NewStatusCommand(),
		NewUpdateCommand(),
		NewStatuslineCommand(),
		NewHookCommand(),
		NewVersionCommand(),
		NewSelfUpdateCommand(),
		NewMigrateCommand(),
		NewClaudeCommand(),
		NewGLMCommand(),
		NewWorktreeCommand(),
		NewRankCommand(),
	)

	return rootCmd
}
