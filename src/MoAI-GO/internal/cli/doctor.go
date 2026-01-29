package cli

import (
	"github.com/spf13/cobra"
)

// NewDoctorCommand creates the doctor command
func NewDoctorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check project health",
		Long: `Diagnose issues with your MoAI project configuration,
dependencies, and setup. This command checks for common problems and
provides actionable recommendations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement doctor command
			cmd.Println("doctor command: not yet implemented")
			return nil
		},
	}

	return cmd
}
