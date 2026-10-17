package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Resolver seams: the production values are the very resolvers the owning verbs
// use. Tests replace them to simulate an unresolvable moai home.
var (
	pathsProjectDir    = homestate.ProjectDir
	pathsFactoryDBPath = homestate.FactoryDBPath
)

// newPathsCmd builds `moai paths` (SPEC-MOAI-DB-PATHS-001, card t1827).
func newPathsCmd() *cobra.Command {
	var jsonOutput bool
	var field, run string
	cmd := &cobra.Command{
		Use:     "paths",
		Short:   "Print the confirmed database paths of this project",
		GroupID: "tools",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "{}")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "write one JSON object")
	cmd.Flags().StringVar(&field, "field", "", "write only one field's bare value")
	cmd.Flags().StringVar(&run, "run", "", "report the broker path of this factory run id")
	return cmd
}
