package cli

import (
	"github.com/spf13/cobra"
)

// NewInitCommand creates the init command
func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize MoAI project",
		Long: `Initialize a new MoAI project with templates, configuration,
and project structure. This command sets up the necessary files and directories
for AI-powered development with Claude Code.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement init command
			cmd.Println("init command: not yet implemented")
			return nil
		},
	}

	return cmd
}
