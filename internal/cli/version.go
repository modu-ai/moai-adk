package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/pkg/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	RunE: func(cmd *cobra.Command, _ []string) error {
		out := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(out, "moai-adk %s (commit: %s, built: %s)\n",
			version.GetVersion(),
			version.GetCommit(),
			version.GetDate(),
		)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
