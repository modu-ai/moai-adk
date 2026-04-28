package cli

import (
	"github.com/spf13/cobra"
)

// newSpecCmd creates the 'moai spec' parent command
func newSpecCmd() *cobra.Command {
	specCmd := &cobra.Command{
		Use:   "spec",
		Short: "Manage SPEC documents",
		Long:  `Manage SPEC documents in .moai/specs/ directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		GroupID: "tools",
	}

	// Add subcommands
	specCmd.AddCommand(newSpecStatusCmd())

	return specCmd
}

func init() {
	rootCmd.AddCommand(newSpecCmd())
}
