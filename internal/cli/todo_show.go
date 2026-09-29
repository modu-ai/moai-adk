// todo_show.go — SPEC-TODO-SURFACE-POLISH-001 M1 (card t1349): `moai todo
// show <id|n>`, the single-card read surface the queue never had.
//
// The one-card fate line already existed — `history <id>` prints it — but
// the verb name an operator reaches for (`show 401`, cited by the
// mistyped-verb guard's own guidance since t555) was not registered, so the
// exact address grammar the surface publishes was refused as a typo'd verb.
// show registers that grammar, answering from the SAME store through the
// SAME lookup machine `history` uses (REQ-TSP-001): the live/archived/absent
// fate line, tab-separated with the full untruncated card text LAST, and the
// at-or-below-the-issued-mark qualifier on stderr for an absent id
// (REQ-TSP-002).
//
// READ-ONLY (REQ-TSP-003): the verb enters through LoadPure — no adopt, no
// migration, no lock, no Mutate — and carries the same stderr layout
// disclosure the other read verbs carry, so a non-authoritative
// backlog.json or a divergent stale store says so on the stream stdout
// never touches.
//
// SUBAGENT BOUNDARY (C-HRA-008): nothing here prompts.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newTodoShowCmd — `moai todo show <id|n>`. The verb is the read twin of
// `history <id>`: one line, the fate, the full text.
func newTodoShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id|n>",
		Short: "Print one card's fate and full text on a single line",
		Long: `Print the one line ` + "'moai todo history <id>'" + ` answers for a single card:
the id, its fate ('live' with the current state, 'archived' with the state
it held when closed, or 'absent'), the landing cell, the transition stamps,
and the card text LAST — flattened to one line (tabs and newlines become
spaces) and never truncated. A bare ordinal is normalized to the id form
('show 401' addresses t401). An absent id at or below the queue's issued-id
mark is qualified on stderr, never on stdout.

The verb is read-only: it changes no card, takes no queue mutation lock,
and answers from the same read the other read verbs resolve through.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTodoShow(cmd, args[0])
		},
	}
	return cmd
}

// runTodoShow answers the one-card fate line for the addressed id.
func runTodoShow(cmd *cobra.Command, ref string) error {
	store := newTodoReadStore()
	// REQ-BJD-002 — the layout probe runs BEFORE the read, the same entry
	// point list uses; the disclosures ride stderr only.
	_ = discloseQueueLayout(cmd, "show")
	rec, err := store.LoadPure()
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}
	return renderTodoLookup(cmd.OutOrStdout(), cmd.ErrOrStderr(), rec, normalizeTodoRef(ref), "show")
}
