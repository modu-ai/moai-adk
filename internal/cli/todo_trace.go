// todo_trace.go — SPEC-TODO-CARD-ISSUANCE-001 M3 (REQ-TCI-015): the
// read-only transitive walk over the unified relation vocabulary. Determined
// by the resolver's mapped edges; terminates on legacy cycles via the
// visit set; writes nothing; available to lane sessions like the other
// read-only verbs.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// newTodoTraceCmd — `moai todo trace <id> [--kind <k>]… [--depth <n>]`:
// print every node reachable from the card over the named kinds up to the
// depth bound, deterministically ordered.
func newTodoTraceCmd() *cobra.Command {
	var kinds []string
	var depth int
	cmd := &cobra.Command{
		Use:   "trace <id> [--kind <k>]... [--depth <n>]",
		Short: "Trace relations reachable from a card (read-only, lane-safe)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := normalizeTodoRef(args[0])
			rec, err := newTodoStore().LoadPure()
			if err != nil {
				return err
			}
			lines := factory.TraceCardRelations(rec, id, kinds, depth)
			if len(lines) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no relations reachable from "+id)
				return nil
			}
			for _, line := range lines {
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&kinds, "kind", nil,
		"Relation kinds to walk (repeatable; default: all)")
	cmd.Flags().IntVar(&depth, "depth", 0,
		"Maximum walk depth (0 = unbounded)")
	return cmd
}

