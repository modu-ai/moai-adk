package cli

import (
	"github.com/spf13/cobra"
)

// NewVersionCommand creates the version command
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display the version of MoAI-ADK currently installed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			version := "dev"
			cmd.Printf("moai-adk version %s\n", version)
			return nil
		},
	}

	return cmd
}
