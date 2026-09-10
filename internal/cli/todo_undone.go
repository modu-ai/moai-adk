// todo_undone.go — `moai todo undone <n>` (SPEC-TODO-DESTRUCTIVE-GUARD-001
// REQ-TDG-001/002/013/016): the inverse `done` never had.
//
// `drop` has `undrop` and `next` has `unpick`; `done` was the one destructive
// verb with no way back, and the row it removed took every finding naming it
// along silently. This verb restores both, at the positions they held, and
// empties the archive entry so each row has exactly one home.
//
// SUBAGENT BOUNDARY (C-HRA-008 / REQ-TODO-014): nothing here prompts. Every
// path — success, refusal, error — is positional arguments in, one structured
// stdout line out, human-readable errors on stderr.
package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// todoWriteLine writes one line to w and folds a stream-write failure into
// cause rather than discarding it.
//
// The rest of this package discards the Fprintf result outright, which is
// almost always harmless — the mutation has already landed by the time the
// confirmation is printed. Joining is cheap and strictly better: a caller
// whose stdout is a closed pipe learns that its confirmation never arrived
// instead of reading a silent success.
func todoWriteLine(w io.Writer, cause error, format string, args ...any) error {
	if _, werr := fmt.Fprintf(w, format, args...); werr != nil {
		return errors.Join(cause, fmt.Errorf("todo: writing output: %w", werr))
	}
	return cause
}

// newTodoUndoneCmd — `moai todo undone <n>`.
func newTodoUndoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "undone <n>",
		Short: "Restore a card archived by done, with its findings",
		Long: `Return an archived card to the live queue as one locked write, at the
position it held, together with every finding that named it.

Together with ` + "`done`" + ` this is an exact reversal — the queue record returns to
the same bytes. The findings matter as much as the card: they are the operator's
recorded judgment about it, and a restore that recovered the row alone would lose
them without saying so.

The restore is REFUSED when the id has since been reissued to a different live
card. Overwriting the live row would destroy a card nobody asked to lose, and
splitting one id across two cards is worse than the mistake being undone.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := normalizeTodoRef(args[0])
			store := newTodoStore()
			var restored string
			if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
				// Refused mutations: Mutate writes nothing, so the record
				// stays byte-identical on every one of them.
				at := rec.ArchivedIndex(id)
				if at < 0 {
					return fmt.Errorf("no archived backlog item %s", id)
				}
				restored = rec.Archived[at].Item.Text
				return rec.RestoreCard(id)
			}); err != nil {
				return todoWriteLine(cmd.ErrOrStderr(), err, "Error: %v\n", err)
			}
			return todoWriteLine(cmd.OutOrStdout(), nil, "undone %s %s\n", id, todoTextPrefix(restored))
		},
	}
}
