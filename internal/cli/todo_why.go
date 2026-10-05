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
)

// newTodoWhyCmd — `moai todo why <n>` (REQ-TA-012): print every finding
// naming the card, lock-free. SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-014: the
// output additionally shows the card's issuance attribute projections
// (parent-of / follow-up-of) for cards that carry them — additive for
// attributed cards, absent for legacy records (REQ-TCI-017 as clarified by
// P2-3).
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
			// Issuance attribute projection (card t1454): show the parent
			// follow-up origin before the findings so the operator sees the
			// card's provenance first.
			for i := range rec.Items {
				it := &rec.Items[i]
				if it.ID != id || it.Issuance == nil {
					continue
				}
				if it.Issuance.SpawnedBy != "" {
					kind := "follow-up-of"
					if it.Issuance.Origin == "split" {
						kind = "parent-of"
					}
					_, _ = fmt.Fprintf(out, "  %s %s → %s\n", kind, it.Issuance.SpawnedBy, id)
				}
			}
			findings, indexes := rec.FindingsNaming(id)
			if len(findings) == 0 {
				_, _ = fmt.Fprintf(out, "%s: no findings\n", id)
				return nil
			}
			for i, f := range findings {
				// The index is the address `unrelate` takes, so it leads
				// the line rather than trailing it.
				_, _ = fmt.Fprintf(out, "%d %s\n", indexes[i],
					strings.TrimPrefix(todoFindingLine(rec, id, f), "\t"))
			}
			return nil
		},
	}
}
