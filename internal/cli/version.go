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
		title := "moai-adk " + version.GetVersion()
		detail := fmt.Sprintf("commit: %s  built: %s", version.GetCommit(), version.GetDate())
		_, _ = fmt.Fprintln(out, renderCard(title, detail))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
