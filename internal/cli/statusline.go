package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statuslineCmd = &cobra.Command{
	Use:    "statusline",
	Short:  "Render statusline for Claude Code",
	Long:   "Generate a compact statusline string for display in Claude Code's status bar.",
	Hidden: true,
	RunE:   runStatusline,
}

func init() {
	rootCmd.AddCommand(statuslineCmd)
}

// runStatusline renders a statusline string suitable for Claude Code's status bar.
func runStatusline(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	// Render a compact status using available information
	fmt.Fprint(out, "moai")

	if deps != nil && deps.Config != nil {
		cfg := deps.Config.Get()
		if cfg != nil && cfg.Quality.DevelopmentMode != "" {
			fmt.Fprintf(out, " [%s]", cfg.Quality.DevelopmentMode)
		}
	}

	fmt.Fprintln(out)
	return nil
}
