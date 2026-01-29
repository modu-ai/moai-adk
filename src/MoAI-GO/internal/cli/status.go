package cli

import (
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the status command
func NewStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show project status",
		Long: `Display the current status of your MoAI project, including
active SPEC documents, configuration state, and development progress.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement status command
			cmd.Println("status command: not yet implemented")
			return nil
		},
	}

	return cmd
}
