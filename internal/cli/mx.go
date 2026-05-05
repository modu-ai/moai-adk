package cli

import (
	"github.com/spf13/cobra"
)

// newMxCmd creates the 'moai mx' parent command.
// Includes @MX TAG related subcommands (query, etc.).
func newMxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mx",
		Short:   "@MX TAG 관리 도구",
		Long:    `@MX TAG 사이드카 인덱스를 관리하고 조회하는 도구입니다.`,
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
