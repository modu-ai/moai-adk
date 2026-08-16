// todo.go — SPEC-KANBAN-TODO-CLI-001 M2: the `moai todo` command surface.
//
// Thin cobra wiring over internal/kanban.BacklogStore: every mutation
// delegates to the store's locked Mutate path, reads go through the
// lock-free Load. The verbs serve the kanban dispatch protocol's entry rule
// (`/moai todo` is the operator's act — the lead never picks for the
// operator): `add` and `done` mutate, `list` and bare `next` observe, and
// `next <n> [--spec]` records the operator's pick as one locked write.
//
// Pick-marking race hardening (t71): `add --pick` folds the add and the
// pick into ONE locked write (no queued window for a guessed id to slip
// into), `unpick <n>` is the picked→queued recovery verb, and every pick
// confirmation carries the card text prefix so a mis-pick is immediately
// observable (`--expect <prefix>` additionally refuses a text mismatch).
//
// SUBAGENT BOUNDARY (C-HRA-008 / REQ-TODO-014): this command never prompts.
// It is headless-safe: positional arguments + flags in, one structured
// stdout line out, human-readable errors on stderr, exit 0/1.
package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// todoBacklogPath returns the backlog file location under root — the same
// .moai/state/kanban/backlog.json the todo skill and the dispatch protocol
// name (REQ-TODO-001).
func todoBacklogPath(root string) string {
	return filepath.Join(root, ".moai", "state", "kanban", "backlog.json")
}

// newTodoCmd creates the `moai todo` parent command.
func newTodoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "todo",
		Short: "Operate the kanban backlog queue",
		Long: `Operate the kanban backlog queue at .moai/state/kanban/backlog.json.

The backlog is the operator's queue: entry into the board is the operator's
act (add), and picking the next card is the operator's act too (next <n>).
Mutations serialize on a sibling cross-process lock; reads are lock-free.

A bare invocation renders the queue, which is the form the skill surface and
workflows/todo.md both document; ` + "`moai todo list`" + ` remains valid and prints the
same thing. NoArgs keeps a mistyped verb an error rather than letting it fall
through to the listing.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTodoList(cmd, false)
		},
		GroupID: "tools",
	}
	cmd.AddCommand(newTodoAddCmd(), newTodoListCmd(), newTodoDoneCmd(), newTodoNextCmd(), newTodoUnpickCmd())
	return cmd
}

// newTodoAddCmd — `moai todo add "<text>"` (REQ-TODO-002): append under the
// lock, print the issued id and its 1-based queue position. `--pick` (t71)
// folds the pick into the same locked write.
func newTodoAddCmd() *cobra.Command {
	var pick bool
	cmd := &cobra.Command{
		Use:   "add <text>",
		Short: "Append a card to the backlog queue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[0]
			if strings.TrimSpace(text) == "" {
				return fmt.Errorf("todo add: text must be non-empty")
			}
			store := kanban.NewBacklogStore(todoBacklogPath(resolveProjectDir()))
			if pick {
				return runTodoAddPick(cmd, store, text)
			}
			item, pos, err := store.Add(text)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %d\n", item.ID, pos)
			return nil
		},
	}
	cmd.Flags().BoolVar(&pick, "pick", false,
		"Append AND mark picked as one locked write, printing the issued id")
	return cmd
}

// runTodoAddPick — the `add --pick` body (t71): append AND mark picked as
// ONE cross-process-locked write. The id is issued from the high-water mark
// inside the same Mutate that appends (mirroring store.Add's callback — the
// store deliberately keeps no picked-add variant, so the CLI layer owns this
// composition), so there is no queued window between the add and the pick
// where a guessed id could address a concurrent session's card — the exact
// race that mis-picked t67 on 2026-08-16. The confirmation prints the
// issued id and the card text prefix; the caller never has to guess what
// `--pick` just picked.
func runTodoAddPick(cmd *cobra.Command, store *kanban.BacklogStore, text string) error {
	var item kanban.BacklogItem
	err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		rec.LastSeq++
		item = kanban.BacklogItem{
			ID:      fmt.Sprintf("t%d", rec.LastSeq),
			Text:    text,
			AddedAt: time.Now().UTC().Format(time.RFC3339),
			State:   kanban.BacklogStatePicked,
		}
		rec.Items = append(rec.Items, item)
		return nil
	})
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "picked %s %s\n", item.ID, todoTextPrefix(item.Text))
	return nil
}

// runTodoList renders the backlog lock-free. It backs both entry points —
// the bare `moai todo` and the explicit `moai todo list` — so the two cannot
// drift apart in output.
func runTodoList(cmd *cobra.Command, jsonOutput bool) error {
	store := kanban.NewBacklogStore(todoBacklogPath(resolveProjectDir()))
	rec, err := store.Load()
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}
	out := cmd.OutOrStdout()
	if jsonOutput {
		data, err := json.Marshal(rec)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(out, string(data))
		return nil
	}
	if len(rec.Items) == 0 {
		_, _ = fmt.Fprintln(out, "queue is empty")
		return nil
	}
	for _, it := range rec.Items {
		_, _ = fmt.Fprintf(out, "%s\t%s\t%s\n", it.ID, it.State, it.Text)
	}
	return nil
}

// newTodoListCmd — `moai todo list [--json]` (REQ-TODO-003): render the
// queue lock-free; --json emits the structured records.
func newTodoListCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Render the backlog queue (lock-free)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTodoList(cmd, jsonOutput)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false,
		"Emit the backlog records as JSON on stdout")
	return cmd
}

// newTodoDoneCmd — `moai todo done <n>` (REQ-TODO-004): remove the
// addressed row under the lock. A bare <n> is normalized to the item id
// t<n>; the explicit id is the preferred form because queue positions move
// under concurrent adds.
func newTodoDoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "done <n>",
		Short: "Remove a card from the backlog queue by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := normalizeTodoRef(args[0])
			store := kanban.NewBacklogStore(todoBacklogPath(resolveProjectDir()))
			if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
				for i, it := range rec.Items {
					if it.ID == id {
						rec.Items = append(rec.Items[:i], rec.Items[i+1:]...)
						return nil
					}
				}
				// Refused mutation: Mutate writes nothing, so the file stays
				// byte-identical on a miss.
				return fmt.Errorf("no backlog item %s", id)
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "done %s\n", id)
			return nil
		},
	}
}

// newTodoNextCmd — `moai todo next [<n>] [--spec <SPEC-ID>]` (REQ-TODO-005).
// Bare: print the queued items oldest-first as read-only candidates — the
// pick stays the operator's act. With <n>: mark the addressed item picked
// (attaching spec_id when given) as a single locked write. The confirmation
// carries the card text prefix (t71: a mis-pick must be observable), and
// `--expect <prefix>` refuses the pick when the addressed card's text does
// not match, leaving the file untouched.
func newTodoNextCmd() *cobra.Command {
	var specID string
	var expect string
	cmd := &cobra.Command{
		Use:   "next [<n>]",
		Short: "List queued cards (bare) or pick one (<n>)",
		Long: `Bare ` + "`moai todo next`" + ` prints the queued items oldest-first as
read-only candidates — the selection remains the operator's act performed
through the lead session's question channel. ` + "`moai todo next <n> [--spec <SPEC-ID>]`" + `
marks the addressed item picked (attaching spec_id when given) as one
locked write, confirming with the card text prefix. ` + "`--expect <prefix>`" + `
refuses the pick unless the addressed card's text starts with the prefix.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := kanban.NewBacklogStore(todoBacklogPath(resolveProjectDir()))
			out := cmd.OutOrStdout()

			if len(args) == 0 {
				rec, err := store.Load()
				if err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
					return err
				}
				queued := 0
				for _, it := range rec.Items {
					if it.State != kanban.BacklogStateQueued {
						continue
					}
					queued++
					_, _ = fmt.Fprintf(out, "%s\t%s\n", it.ID, it.Text)
				}
				if queued == 0 {
					_, _ = fmt.Fprintln(out, "queue is empty")
				}
				return nil
			}

			id := normalizeTodoRef(args[0])
			var pickedText string
			if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
				for i := range rec.Items {
					if rec.Items[i].ID == id {
						if expect != "" && !strings.HasPrefix(rec.Items[i].Text, expect) {
							// Refused mutation: Mutate writes nothing, so the
							// file stays byte-identical on a mismatch.
							return fmt.Errorf("backlog item %s is %q, not matching --expect %q",
								id, todoTextPrefix(rec.Items[i].Text), expect)
						}
						rec.Items[i].State = kanban.BacklogStatePicked
						pickedText = rec.Items[i].Text
						if specID != "" {
							// Recorded as-is: the store is not a SPEC registry;
							// normalization is out of scope (acceptance.md §C).
							spec := specID
							rec.Items[i].SpecID = &spec
						}
						return nil
					}
				}
				return fmt.Errorf("no backlog item %s", id)
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			_, _ = fmt.Fprintf(out, "picked %s %s\n", id, todoTextPrefix(pickedText))
			return nil
		},
	}
	cmd.Flags().StringVar(&specID, "spec", "",
		"SPEC-ID to attach when picking (recorded verbatim)")
	cmd.Flags().StringVar(&expect, "expect", "",
		"Refuse the pick unless the card text starts with this prefix")
	return cmd
}

// newTodoUnpickCmd — `moai todo unpick <n>` (t71): revert a picked card to
// queued as one locked write. The recovery verb the pick-marking race
// incident lacked — recovery used to require done+re-add, churning the id
// and losing added_at. Unpick preserves both and clears the spec_id
// attached at pick time, restoring the card to the shape `add` issued.
func newTodoUnpickCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unpick <n>",
		Short: "Revert a picked card to queued by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := normalizeTodoRef(args[0])
			store := kanban.NewBacklogStore(todoBacklogPath(resolveProjectDir()))
			var text string
			if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
				for i := range rec.Items {
					if rec.Items[i].ID != id {
						continue
					}
					if rec.Items[i].State != kanban.BacklogStatePicked {
						// Refused mutation: Mutate writes nothing, so the
						// file stays byte-identical on a refusal.
						return fmt.Errorf("backlog item %s is %s, not picked", id, rec.Items[i].State)
					}
					rec.Items[i].State = kanban.BacklogStateQueued
					rec.Items[i].SpecID = nil
					text = rec.Items[i].Text
					return nil
				}
				return fmt.Errorf("no backlog item %s", id)
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "unpicked %s %s\n", id, todoTextPrefix(text))
			return nil
		},
	}
}

// normalizeTodoRef maps a bare <n> argument to the item id t<n>; an explicit
// id (t<n>) passes through unchanged (REQ-TODO-004's normalization rule).
func normalizeTodoRef(arg string) string {
	if !strings.HasPrefix(arg, "t") {
		return "t" + arg
	}
	return arg
}

// todoTextPrefixMax bounds the card text carried in a pick confirmation —
// long enough to spot a mis-pick at a glance, short enough to keep the
// one-line output contract (t71 names ~40 chars).
const todoTextPrefixMax = 40

// todoTextPrefix renders the first todoTextPrefixMax runes of a card text
// for pick confirmations. Rune-safe by construction: the backlog carries
// multi-byte card texts (ko/ja/zh), and a byte slice could cut a character
// mid-sequence. Truncated text is marked with a trailing ellipsis.
func todoTextPrefix(text string) string {
	runes := []rune(text)
	if len(runes) <= todoTextPrefixMax {
		return text
	}
	return string(runes[:todoTextPrefixMax]) + "..."
}

func init() {
	rootCmd.AddCommand(newTodoCmd())
}
