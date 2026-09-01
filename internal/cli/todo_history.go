// todo_history.go — SPEC-TODO-ARCHIVE-QUERY-001: `moai todo history [id]`,
// the read surface over the backlog archive.
//
// The archive already exists and rides in memory on every read
// (readRecord → readArchive); what the queue never had is a reader. This
// verb answers, in one call, which of the four fates an id holds — the
// three live states (`queued` | `picked` | `dropped`, all of them live rows)
// or `archived` — and lists the archive most-recently-archived first when
// asked with no id.
//
// The line shape is the contract an operator scripts against: tab-separated,
// card text LAST, so a consumer reading the tail is unaffected if a column
// is ever added — the convention `pr` already holds:
//
//	<id>\tlive\tqueued|picked|dropped\t<text>
//	<id>\tarchived\t<state-at-archive>\t<text>
//	<id>\tabsent
//
// READ-ONLY (REQ-TAQ-010): the verb reads through LoadPure — the read that
// never adopts, never migrates a legacy queue, and never takes the lock —
// and calls no Mutate. A store that cannot vouch for an archive is
// disclosed on stderr rather than answered silently (REQ-TAQ-013).
//
// SUBAGENT BOUNDARY (REQ-TAQ-014): nothing here prompts.
package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// todoHistoryEmptyArchive is the explicit empty-archive line (REQ-TAQ-009):
// silence is indistinguishable from a crash, so an empty listing says so.
const todoHistoryEmptyArchive = "archive is empty"

// newTodoHistoryCmd — `moai todo history [<id|n>]`. The constructor name is
// a fixed contract: AC-TAQ-011 clause 1 and AC-TAQ-014 locate the verb by
// grepping this exact symbol.
func newTodoHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "history [<id|n>]",
		Short: "Look up a card's fate, or list the archive",
		Long: `Answer what became of a card. 'moai todo history <id>' prints one
line naming the id and its fate — 'live' with the card's current state
('queued', 'picked', or 'dropped'), 'archived' with the state it held when
it was closed, or 'absent' when the queue holds no record of it. A bare
ordinal is normalized to the id form, the same rule done, undone, why and
next accept.

'moai todo history' with no id lists the archive most-recently-archived
first, bounded at 20 entries ('--limit 0' lifts the bound).

The verb is read-only: it takes no lock, writes nothing, and archived rows
stay invisible to every other reader (list, next, why, analyze and the
counts unchanged).`,
		Args: cobra.MaximumNArgs(1),
		RunE: runTodoHistory,
	}
}

// runTodoHistory renders the fate answer or the archive listing.
func runTodoHistory(cmd *cobra.Command, args []string) error {
	rec, err := newTodoStore().LoadPure()
	if err != nil {
		if _, werr := fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err); werr != nil {
			return werr
		}
		return err
	}
	out := cmd.OutOrStdout()
	if len(args) == 0 {
		return renderTodoHistoryListing(out, rec)
	}
	return renderTodoHistoryLookup(out, rec, normalizeTodoRef(args[0]))
}

// renderTodoHistoryLookup prints the one fate line for id.
func renderTodoHistoryLookup(out io.Writer, rec *kanban.BacklogRecord, id string) error {
	for _, it := range rec.Items {
		if it.ID == id {
			_, err := fmt.Fprintf(out, "%s\tlive\t%s\t%s\n", it.ID, it.State, it.Text)
			return err
		}
	}
	if at := rec.ArchivedIndex(id); at >= 0 {
		entry := rec.Archived[at]
		_, err := fmt.Fprintf(out, "%s\tarchived\t%s\t%s\n", entry.Item.ID, entry.Item.State, entry.Item.Text)
		return err
	}
	_, err := fmt.Fprintf(out, "%s\tabsent\n", id)
	return err
}

// renderTodoHistoryListing prints the archive newest-first (the record
// stores archive order oldest-first, so the listing walks it backwards).
func renderTodoHistoryListing(out io.Writer, rec *kanban.BacklogRecord) error {
	if len(rec.Archived) == 0 {
		_, err := fmt.Fprintln(out, todoHistoryEmptyArchive)
		return err
	}
	for i := len(rec.Archived) - 1; i >= 0; i-- {
		entry := rec.Archived[i]
		if _, err := fmt.Fprintf(out, "%s\tarchived\t%s\t%s\n", entry.Item.ID, entry.Item.State, entry.Item.Text); err != nil {
			return err
		}
	}
	return nil
}
