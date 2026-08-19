// todo_drop.go — t153: the queue's discard verb `drop` and its exact
// inverse `undrop`.
//
// The store has carried BacklogStateDropped since the queue existed, with no
// CLI verb reaching it — which is why every dropped card on a live queue was
// written into the file by hand, the one thing the doctrine tells the
// operator not to do. These two verbs close that, under the contract the
// edit/move pair established: refusals return from inside Mutate's callback
// so the file stays byte-identical, `--expect` guards against an id typed
// from a stale listing, and nothing is inferred.
//
// [HARD] The drop decision is the OPERATOR's. Nothing here judges whether a
// card is still worth keeping — no staleness heuristic, no duplicate
// detection, no absorption of one card into another. An agent that drops
// cards on its own initiative collides head-on with the [HARD] clauses in
// workflows/todo.md and kanban-dispatch.md, and a doctrine change would have
// to come first.
//
// EXACT REVERSAL is the property that makes a wrong drop survivable, and it
// is the reason for three refusals that would otherwise look overcautious:
//
//   - Only a QUEUED card may be dropped. Dropping a picked card would lose
//     the picked state (and its spec_id) with nowhere to record it — the
//     per-item field set is frozen at five fields — so undrop could not
//     restore what drop took. `unpick` then `drop` is the two-step path, and
//     both steps are reversible.
//   - The reason may not contain `]`. The marker is parsed back off the text
//     at the first `] `, so a reason carrying one would strip the wrong span
//     and silently corrupt the restored card.
//   - An already-dropped card is refused rather than re-marked. A second
//     marker would leave undrop restoring the card to a still-marked state.
//
// With those held, drop followed by undrop returns the queue file to the
// same bytes — asserted directly by the round-trip test.
package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// The drop marker follows the convention the hand-written dropped cards
// already established on the live queue: `[DROPPED — <reason>] <text>`.
// Keeping the CLI's output identical to what operators have been writing by
// hand means the existing cards read the same as the new ones, and undrop
// recovers both.
const (
	todoDropMarkerOpen  = "[DROPPED — "
	todoDropMarkerClose = "] "
)

// newTodoDropCmd — `moai todo drop <n> "<reason>" [--expect <prefix>]`:
// move a queued card to the dropped state, recording why in its text.
//
// The card STAYS in the file. `done` removes a finished card; `drop` keeps a
// discarded one visible with the reason attached, so a later reader can see
// what was decided and reverse it if the judgement was wrong. A dropped card
// is not a pick candidate — bare `next` lists queued cards only.
func newTodoDropCmd() *cobra.Command {
	var expect string
	cmd := &cobra.Command{
		Use:   "drop <n> <reason>",
		Short: "Discard a queued card by id, recording the reason",
		Long: `Move the addressed queued card to the dropped state as one locked
write, prefixing its text with ` + "`[DROPPED — <reason>] `" + ` — the convention
already in use on hand-written dropped cards.

The card stays in the file: ` + "`done`" + ` removes a finished card, ` + "`drop`" + ` keeps a
discarded one visible with its reason, and ` + "`undrop <n>`" + ` reverses it exactly.
A dropped card is not a pick candidate.

Only a queued card may be dropped — a picked card is unpicked first, so
nothing that undrop cannot restore is ever taken. ` + "`--expect <prefix>`" + `
refuses the drop unless the addressed card's text starts with the prefix,
leaving the file untouched.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := normalizeTodoRef(args[0])
			reason := strings.TrimSpace(args[1])
			if reason == "" {
				err := fmt.Errorf("todo drop: reason must be non-empty — a drop without a recorded reason is unreviewable")
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			if strings.Contains(reason, "]") {
				err := fmt.Errorf("todo drop: reason must not contain %q — the marker is parsed back off at the first %q, so undrop could not restore the card", "]", todoDropMarkerClose)
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}

			store := newTodoStore()
			var original string
			if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
				for i := range rec.Items {
					if rec.Items[i].ID != id {
						continue
					}
					// Refused mutations below: Mutate writes nothing, so the
					// file stays byte-identical on every one of them.
					if expect != "" && !strings.HasPrefix(rec.Items[i].Text, expect) {
						return fmt.Errorf("backlog item %s is %q, not matching --expect %q",
							id, todoTextPrefix(rec.Items[i].Text), expect)
					}
					switch rec.Items[i].State {
					case kanban.BacklogStateDropped:
						return fmt.Errorf("backlog item %s is already dropped", id)
					case kanban.BacklogStateQueued:
						// the only droppable state
					default:
						return fmt.Errorf(
							"backlog item %s is %s, not queued — unpick it first, so undrop can restore it exactly",
							id, rec.Items[i].State)
					}
					original = rec.Items[i].Text
					rec.Items[i].Text = todoDropMarkerOpen + reason + todoDropMarkerClose + original
					rec.Items[i].State = kanban.BacklogStateDropped
					return nil
				}
				return fmt.Errorf("no backlog item %s", id)
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "dropped %s %s (reason: %s)\n",
				id, todoTextPrefix(original), reason)
			return nil
		},
	}
	cmd.Flags().StringVar(&expect, "expect", "",
		"Refuse the drop unless the card text starts with this prefix")
	return cmd
}

// newTodoUndropCmd — `moai todo undrop <n> [--expect <prefix>]`: return a
// dropped card to queued, stripping the marker `drop` added.
//
// The STATE is the authority, not the marker: a card written into the file
// by hand with state dropped and no marker still undrops, and its text is
// left exactly as it was. That is what makes the four hand-decided cards on
// the live queue recoverable through the CLI rather than by another hand
// edit.
func newTodoUndropCmd() *cobra.Command {
	var expect string
	cmd := &cobra.Command{
		Use:   "undrop <n>",
		Short: "Return a dropped card to queued by id",
		Long: `Revert the addressed dropped card to queued as one locked write,
stripping the ` + "`[DROPPED — <reason>] `" + ` marker when one is present. The
state is the authority: a card marked dropped by hand, with no marker in its
text, undrops with its text untouched.

Together with ` + "`drop`" + ` this is an exact reversal — the queue file returns to
the same bytes. ` + "`--expect <prefix>`" + ` matches the card's CURRENT text (the
form a listing shows, marker included) and refuses the undrop on a mismatch.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := normalizeTodoRef(args[0])
			store := newTodoStore()
			var restored string
			if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
				for i := range rec.Items {
					if rec.Items[i].ID != id {
						continue
					}
					if expect != "" && !strings.HasPrefix(rec.Items[i].Text, expect) {
						return fmt.Errorf("backlog item %s is %q, not matching --expect %q",
							id, todoTextPrefix(rec.Items[i].Text), expect)
					}
					if rec.Items[i].State != kanban.BacklogStateDropped {
						return fmt.Errorf("backlog item %s is %s, not dropped", id, rec.Items[i].State)
					}
					restored = stripTodoDropMarker(rec.Items[i].Text)
					rec.Items[i].Text = restored
					rec.Items[i].State = kanban.BacklogStateQueued
					return nil
				}
				return fmt.Errorf("no backlog item %s", id)
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "undropped %s %s\n", id, todoTextPrefix(restored))
			return nil
		},
	}
	cmd.Flags().StringVar(&expect, "expect", "",
		"Refuse the undrop unless the card text starts with this prefix")
	return cmd
}

// stripTodoDropMarker removes a leading `[DROPPED — <reason>] ` marker,
// returning the text unchanged when there is none. The span ends at the
// FIRST closing delimiter, which is exactly why `drop` refuses a reason
// containing `]` — with that refusal in place the parse cannot land in the
// wrong place.
func stripTodoDropMarker(text string) string {
	if !strings.HasPrefix(text, todoDropMarkerOpen) {
		return text
	}
	rest := text[len(todoDropMarkerOpen):]
	end := strings.Index(rest, todoDropMarkerClose)
	if end < 0 {
		return text
	}
	return rest[end+len(todoDropMarkerClose):]
}
