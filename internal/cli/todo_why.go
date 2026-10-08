// todo_why.go — SPEC-TODO-ANALYSIS-001 M4: `moai todo why <n>`.
//
// The verb answers one question — "what does the queue know about this
// card?" — and it always answers. A card with no findings prints an explicit
// line saying so, because silence is indistinguishable from a crash, and an
// operator who cannot tell those apart learns to distrust the whole surface.
//
// SUBAGENT BOUNDARY (REQ-TA-015): nothing here prompts.
package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// newTodoWhyCmd — `moai todo why <n>` (REQ-TA-012): print every finding
// naming the card, lock-free. SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-014: the
// output additionally shows the card's issuance attribute projections and
// the GTD relations, all through the common relation resolver (card t1454
// card-review r2 findings 11/14) — the follow-up projection reads child →
// origin, the parent projection parent → child.
func newTodoWhyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "why <n>",
		Short: "Print every finding naming a card",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := normalizeTodoRef(args[0])
			// REQ-BJD-002 — probed before the read (todo_disclosure.go).
			_ = discloseQueueLayout(cmd, "why")
			rec, err := newTodoStore().Load()
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			out := cmd.OutOrStdout()
			// The common relation resolver (card t1454 card-review r2
			// findings 11/14): findings keep their finding-line rendering;
			// the resolver's other edges — the spawned_by projections
			// (follow-up child → origin, parent parent → child) and the GTD
			// relations — render one line each, before the findings so the
			// operator sees the card's provenance first.
			edges := factory.ResolveCardEdges(rec, factory.ListGTDCardRelations(cmd.Context(), newTodoStore()), id)
			hasFinding := false
			for _, e := range edges {
				if e.Index > 0 {
					hasFinding = true
					continue
				}
				_, _ = fmt.Fprintf(out, "  %s %s → %s  source=%s\n", e.Kind, e.From, e.To, e.Source)
			}
			if !hasFinding {
				_, _ = fmt.Fprintf(out, "%s: no findings\n", id)
				return nil
			}
			for _, e := range edges {
				if e.Index <= 0 {
					continue
				}
				// The index is the address `unrelate` takes, so it leads
				// the line rather than trailing it.
				_, _ = fmt.Fprintf(out, "%d %s\n", e.Index,
					strings.TrimPrefix(todoFindingLine(rec, id, rec.Findings[e.Index-1]), "\t"))
			}
			return nil
		},
	}
}
