package cli

import "github.com/spf13/cobra"

// newAgenticFactoryCmd is the Issue #1948 canonical-command skeleton.
// The name and subcommands are provisional until SPEC-AGENTIC-FACTORY-001
// fixes them (design.md §8), so the command is hidden and changes no state.
// The legacy `moai cc -f/-a` and `moai factory` entry points are untouched.
func newAgenticFactoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "agentic-factory",
		Short:  "Agentic factory (provisional skeleton, Issue #1948)",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Println("agentic-factory: provisional skeleton; command names are not fixed yet")
			return nil
		},
	}
}
