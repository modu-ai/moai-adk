package cli

import (
	"github.com/spf13/cobra"
)

// newMxCmd creates the 'moai mx' parent command.
// Includes @MX TAG related subcommands (query, etc.).
func newMxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mx",
		Short:   "@MX TAG management tool",
		Long:    `Tool for managing and querying the @MX TAG sidecar index.`,
		GroupID: "tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// SPEC-V3R2-SPC-004: Register query subcommand
	cmd.AddCommand(newMxQueryCmd())

	return cmd
}

func init() {
	rootCmd.AddCommand(newMxCmd())
}
