// todo_edit_move.go — t119: the queue's two correction verbs, `edit` (card
// text) and `move` (queue order).
//
// Both are OPERATOR acts, exactly like `add` and `next <n>`: the CLI records
// a correction the operator decided on. Nothing here infers what a card
// should say or where it belongs — no analysis, no absorption, no silent
// promotion. Those would collide head-on with the [HARD] clauses in
// workflows/todo.md and kanban-dispatch.md (the pick is the operator's; the
// queue is never auto-populated or reordered by inferred priority), and a
// doctrine change would have to come first.
//
// Recoverability is the property that makes a mis-correction survivable, and
// each verb carries it differently:
//   - edit prints the PRIOR text alongside the new one, so a wrong edit is
//     reversed by editing back — the identity fields (id, added_at, state,
//     spec_id) never move, so nothing else has to be reconstructed.
//   - move only permutes the item slice; no item is dropped, duplicated, or
//     otherwise altered, so a wrong move is reversed by another move.
//
// Every refusal returns from inside Mutate's callback, which writes nothing:
// the queue file stays byte-identical on a miss, a mismatch, or a malformed
// invocation.
package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// newTodoEditCmd — `moai todo edit <n> <text> [--expect <prefix>]`: rewrite
// an existing card's text under the lock, leaving every other field alone.
//
// The verb exists because correcting a card previously meant done + re-add,
// which churns the id and loses added_at — the same cost `unpick` removed
// for a mis-pick. `--expect` mirrors `next --expect`: it refuses the edit
// unless the addressed card's text starts with the prefix, so an id typed
// from a stale listing cannot silently rewrite a different card.
func newTodoEditCmd() *cobra.Command {
	var expect string
	cmd := &cobra.Command{
		Use:   "edit <n> <text>",
		Short: "Rewrite a card's text by id, preserving its identity",
		Long: `Rewrite the addressed card's text as one locked write. The id,
added_at, state, and spec_id are preserved — only the text changes, so a
correction never churns the card's identity the way done + re-add does.

The confirmation carries BOTH the new text and the prior text, so a wrong
edit is reversed by editing back. ` + "`--expect <prefix>`" + ` refuses the edit
unless the addressed card's text starts with the prefix, leaving the file
untouched.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := normalizeTodoRef(args[0])
			text := args[1]
			if strings.TrimSpace(text) == "" {
				err := fmt.Errorf("todo edit: text must be non-empty")
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			store := newTodoStore()
			var prior string
			if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
				for i := range rec.Items {
					if rec.Items[i].ID != id {
						continue
					}
					if expect != "" && !strings.HasPrefix(rec.Items[i].Text, expect) {
						// Refused mutation: Mutate writes nothing, so the
						// file stays byte-identical on a mismatch.
						return fmt.Errorf("backlog item %s is %q, not matching --expect %q",
							id, todoTextPrefix(rec.Items[i].Text), expect)
					}
					prior = rec.Items[i].Text
					rec.Items[i].Text = text
					return nil
				}
				return fmt.Errorf("no backlog item %s", id)
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "edited %s %s (was %s)\n",
				id, todoTextPrefix(text), todoTextPrefix(prior))
			return nil
		},
	}
	cmd.Flags().StringVar(&expect, "expect", "",
		"Refuse the edit unless the card text starts with this prefix")
	return cmd
}

// newTodoMoveCmd — `moai todo move <n> (--top|--bottom|--before <m>|--after <m>)`:
// reposition a card within the queue file's order under the lock.
//
// Order is the only thing the queue records about priority — there are no
// priority fields (workflows/todo.md § Boundaries: not a task tracker) — so
// an operator who wants a card considered sooner previously had to hand-edit
// the file, the one thing the doctrine tells them not to do.
//
// Exactly one position flag is required: a move with no destination, or with
// two, is a malformed invocation rather than a guess the CLI resolves.
func newTodoMoveCmd() *cobra.Command {
	var top, bottom bool
	var before, after string
	cmd := &cobra.Command{
		Use:   "move <n> (--top | --bottom | --before <m> | --after <m>)",
		Short: "Reposition a card within the queue order by id",
		Long: `Move the addressed card within the queue file's order as one
locked write. Exactly one destination flag is required — ` + "`--top`" + `,
` + "`--bottom`" + `, ` + "`--before <m>`" + `, or ` + "`--after <m>`" + ` — because a move with no
destination, or with two, is a malformed invocation and not a guess.

The move permutes the item slice and nothing else: no card is dropped,
duplicated, or altered, so a wrong move is reversed by another move. The
confirmation prints the card's new 1-based position.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := normalizeTodoRef(args[0])
			anchor, err := todoMoveDestination(top, bottom, before, after)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			store := newTodoStore()
			var pos int
			var text string
			if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
				idx, err := applyTodoMove(rec, id, anchor)
				if err != nil {
					// Refused mutation: Mutate writes nothing, so the file
					// stays byte-identical on any refusal below.
					return err
				}
				pos = idx + 1
				text = rec.Items[idx].Text
				return nil
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "moved %s %d %s\n", id, pos, todoTextPrefix(text))
			return nil
		},
	}
	cmd.Flags().BoolVar(&top, "top", false, "Move the card to the front of the queue")
	cmd.Flags().BoolVar(&bottom, "bottom", false, "Move the card to the back of the queue")
	cmd.Flags().StringVar(&before, "before", "", "Move the card immediately before this id")
	cmd.Flags().StringVar(&after, "after", "", "Move the card immediately after this id")
	return cmd
}

// todoMoveAnchor is a validated destination: `kind` is one of top, bottom,
// before, after, and `ref` carries the anchor id for the relative kinds.
type todoMoveAnchor struct {
	kind string
	ref  string
}

// todoMoveDestination validates that exactly one destination flag was given
// and returns it. Zero flags and two-or-more flags are both errors — the CLI
// never picks a destination the operator did not name.
func todoMoveDestination(top, bottom bool, before, after string) (todoMoveAnchor, error) {
	var chosen []todoMoveAnchor
	if top {
		chosen = append(chosen, todoMoveAnchor{kind: "top"})
	}
	if bottom {
		chosen = append(chosen, todoMoveAnchor{kind: "bottom"})
	}
	if before != "" {
		chosen = append(chosen, todoMoveAnchor{kind: "before", ref: normalizeTodoRef(before)})
	}
	if after != "" {
		chosen = append(chosen, todoMoveAnchor{kind: "after", ref: normalizeTodoRef(after)})
	}
	if len(chosen) != 1 {
		return todoMoveAnchor{}, fmt.Errorf(
			"todo move: give exactly one of --top, --bottom, --before <m>, --after <m> (got %d)",
			len(chosen))
	}
	return chosen[0], nil
}

// applyTodoMove repositions id within rec.Items and returns its new index.
// It is a pure permutation of the slice: the item value is carried over
// untouched, and no other item is created or removed.
//
// Relative destinations resolve against the list WITHOUT the moved card, so
// "--before <m>" lands immediately ahead of m regardless of which side of m
// the card started on.
func applyTodoMove(rec *kanban.BacklogRecord, id string, anchor todoMoveAnchor) (int, error) {
	from := todoItemIndex(rec.Items, id)
	if from < 0 {
		return 0, fmt.Errorf("no backlog item %s", id)
	}
	if (anchor.kind == "before" || anchor.kind == "after") && anchor.ref == id {
		return 0, fmt.Errorf("cannot move backlog item %s relative to itself", id)
	}

	item := rec.Items[from]
	rest := make([]kanban.BacklogItem, 0, len(rec.Items)-1)
	rest = append(rest, rec.Items[:from]...)
	rest = append(rest, rec.Items[from+1:]...)

	var to int
	switch anchor.kind {
	case "top":
		to = 0
	case "bottom":
		to = len(rest)
	default:
		ref := todoItemIndex(rest, anchor.ref)
		if ref < 0 {
			return 0, fmt.Errorf("no backlog item %s", anchor.ref)
		}
		to = ref
		if anchor.kind == "after" {
			to = ref + 1
		}
	}

	moved := make([]kanban.BacklogItem, 0, len(rec.Items))
	moved = append(moved, rest[:to]...)
	moved = append(moved, item)
	moved = append(moved, rest[to:]...)
	rec.Items = moved
	return to, nil
}

// todoItemIndex returns the index of id in items, or -1.
func todoItemIndex(items []kanban.BacklogItem, id string) int {
	for i := range items {
		if items[i].ID == id {
			return i
		}
	}
	return -1
}
