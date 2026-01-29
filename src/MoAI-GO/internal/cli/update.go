package cli

import (
	"github.com/spf13/cobra"
)

// NewUpdateCommand creates the update command
func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update MoAI templates",
		Long: `Update MoAI templates and configuration to the latest version.
This command syncs your local project with the latest template changes from
the MoAI-ADK distribution.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement update command
			cmd.Println("update command: not yet implemented")
			return nil
		},
	}

	return cmd
}
