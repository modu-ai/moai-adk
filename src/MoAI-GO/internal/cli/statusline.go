package cli

import (
	"github.com/spf13/cobra"
)

// NewStatuslineCommand creates the statusline command
func NewStatuslineCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "statusline",
		Short: "Display statusline",
		Long: `Display the MoAI statusline with project information,
AI session state, and development context. Useful for monitoring
your development workflow.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement statusline command
			cmd.Println("statusline command: not yet implemented")
			return nil
		},
	}

	return cmd
}
