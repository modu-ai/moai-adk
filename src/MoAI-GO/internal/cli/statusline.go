package cli

import (
	"fmt"
	"os"

	"github.com/anthropics/moai-adk-go/internal/statusline"
	"github.com/anthropics/moai-adk-go/pkg/version"
	"github.com/spf13/cobra"
)

// NewStatuslineCommand creates the statusline command
func NewStatuslineCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "statusline",
		Short: "Display statusline",
		Long: `Display the MoAI statusline with project information,
AI session state, and development context. Useful for monitoring
your development workflow.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusline(format)
		},
	}

	cmd.Flags().StringVar(&format, "format", "", "Custom format string")

	return cmd
}

// runStatusline executes the statusline command logic
func runStatusline(format string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	// Get version from build info (set via ldflags)
	ver := version.GetVersion()

	// Create formatter
	formatter := statusline.NewFormatter(cwd, ver)

	// Generate statusline
	output, err := formatter.Format(format)
	if err != nil {
		return fmt.Errorf("error generating statusline: %w", err)
	}

	// Print statusline (no newline for integration purposes)
	fmt.Print(output)

	return nil
}
