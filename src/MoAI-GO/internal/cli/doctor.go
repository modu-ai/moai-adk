package cli

import (
	"fmt"
	"os"

	"github.com/anthropics/moai-adk-go/internal/doctor"
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
			return runDoctor()
		},
	}

	return cmd
}

// runDoctor executes the doctor command logic
func runDoctor() error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	// Create doctor and run checks
	doc := doctor.NewDoctor(cwd)
	return doc.RunAllChecks()
}
