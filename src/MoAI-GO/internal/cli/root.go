package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root CLI command
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "moai-adk",
		Short: "MoAI Application Development Kit",
		Long: `MoAI-ADK is a comprehensive development toolkit for creating
AI-powered applications with Claude Code. It provides project initialization,
template management, configuration handling, and CLI command infrastructure.`,
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
	)

	return rootCmd
}
