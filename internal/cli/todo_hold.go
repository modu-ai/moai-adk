// todo_hold.go — SPEC-TODO-HOLD-STATE-001 M2: the operator's park verbs
// `hold` and `unhold`.
//
// A held card is parked out of the queue WITHOUT its text being touched —
// the state is the whole feature, and no `[HOLD]` marker ever enters the
// text (the contrast that defines the design: prose holds are exactly what
// machines do not read). The verbs are drop-isomorphic under the contract
// the edit/move pair established: refusals return from inside Mutate's
// callback so the file stays byte-identical, `--expect` guards against an id
// typed from a stale listing, and nothing is inferred.
//
// [HARD] The hold decision is the OPERATOR's (or the leader speaking for the
// operator). No lease path, lane session, or machine consumer gains a verb
// that sets or clears the state — the actor boundary is surface absence, and
// the queue's machine selectors exclude held cards by positive enumeration
// of the states they accept.
package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// holdRecoveryAdvice names the verb that returns a card to queued from the
// state it is in, so a refusal tells the operator the two-step path instead
// of only the fact of the refusal.
func holdRecoveryAdvice(state kanban.BacklogState) string {
	switch state {
	case kanban.BacklogStatePicked:
		return "unpick it first"
	case kanban.BacklogStateDropped:
		return "undrop it first"
	case kanban.BacklogStateHold:
		return "it is already held — unhold re-queues it"
	default:
		return "no recovery verb is defined for that state"
	}
}

// newTodoHoldCmd — `moai todo hold <n> [--expect <prefix>]`: move a queued
// card to the hold state as one locked write. The text is never touched: a
// held card carries no marker, no reason field, and no timestamp — an
// operator who wants a reason keeps it in the card text, as today.
func newTodoHoldCmd() *cobra.Command {
	var expect string
	cmd := &cobra.Command{
		Use:   "hold <n>",
		Short: "Park a queued card out of the queue by id (operator only)",
		Long: `Move the addressed queued card to the hold state as one locked
write. The text is NEVER touched: hold is a state, not a marker — every
machine selector enumerates the states it accepts positively, so a held card
is invisible to ` + "`next`" + `, the auto-done scan, and every lease path by
construction.

` + "`moai todo unhold <n>`" + ` returns the card to queued. Only an operator (or
the leader speaking for one) holds a card; no lane or machine leaser can reach
this verb's effect. ` + "`--expect <prefix>`" + ` refuses the hold unless the
addressed card's text starts with the prefix, leaving the file untouched.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := normalizeTodoRef(args[0])
			store := newTodoStore()
			var held string
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
					case kanban.BacklogStateQueued:
						// the only holdable state
					default:
						return fmt.Errorf("backlog item %s is %s, not queued — %s",
							id, rec.Items[i].State, holdRecoveryAdvice(rec.Items[i].State))
					}
					held = rec.Items[i].Text
					rec.Items[i].State = kanban.BacklogStateHold
					return nil
				}
				return fmt.Errorf("no backlog item %s", id)
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "held %s %s\n", id, todoTextPrefix(held))
			return nil
		},
	}
	cmd.Flags().StringVar(&expect, "expect", "",
		"Refuse the hold unless the card text starts with this prefix")
	return cmd
}

// newTodoUnholdCmd — `moai todo unhold <n> [--expect <prefix>]`: return a
// held card to queued. The STATE is the authority, not any marker: the text
// a hold never touched is the text an unhold never touches, and the pair is
// an exact reversal.
func newTodoUnholdCmd() *cobra.Command {
	var expect string
	cmd := &cobra.Command{
		Use:   "unhold <n>",
		Short: "Return a held card to queued by id",
		Long: `Revert the addressed held card to queued as one locked write.
The state is the authority: hold never touched the text, so unhold has
nothing to strip and the pair is an exact reversal.

` + "`--expect <prefix>`" + ` matches the card's text and refuses the unhold on
a mismatch, writing nothing.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := normalizeTodoRef(args[0])
			store := newTodoStore()
			var released string
			if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
				for i := range rec.Items {
					if rec.Items[i].ID != id {
						continue
					}
					if expect != "" && !strings.HasPrefix(rec.Items[i].Text, expect) {
						return fmt.Errorf("backlog item %s is %q, not matching --expect %q",
							id, todoTextPrefix(rec.Items[i].Text), expect)
					}
					switch rec.Items[i].State {
					case kanban.BacklogStateHold:
						// the only unholdable state
					default:
						return fmt.Errorf("backlog item %s is %s, not held", id, rec.Items[i].State)
					}
					released = rec.Items[i].Text
					rec.Items[i].State = kanban.BacklogStateQueued
					return nil
				}
				return fmt.Errorf("no backlog item %s", id)
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "unheld %s %s\n", id, todoTextPrefix(released))
			return nil
		},
	}
	cmd.Flags().StringVar(&expect, "expect", "",
		"Refuse the unhold unless the card text starts with this prefix")
	return cmd
}
