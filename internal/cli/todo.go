// todo.go — SPEC-KANBAN-TODO-CLI-001 M2: the `moai todo` command surface.
//
// Thin cobra wiring over internal/kanban.BacklogStore: every mutation
// delegates to the store's locked Mutate path, reads go through the
// lock-free Load. The verbs serve the kanban dispatch protocol's entry rule
// (`/moai todo` is the operator's act — the lead never picks for the
// operator): `add` and `done` mutate, `list` and bare `next` observe, and
// `next <n> [--spec]` records the operator's pick as one locked write.
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
	cmd.AddCommand(newTodoAddCmd(), newTodoListCmd(), newTodoDoneCmd(), newTodoNextCmd())
	return cmd
}

// newTodoAddCmd — `moai todo add "<text>"` (REQ-TODO-002): append under the
// lock, print the issued id and its 1-based queue position.
func newTodoAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <text>",
		Short: "Append a card to the backlog queue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[0]
			if strings.TrimSpace(text) == "" {
				return fmt.Errorf("todo add: text must be non-empty")
			}
			store := kanban.NewBacklogStore(todoBacklogPath(resolveProjectDir()))
			item, pos, err := store.Add(text)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %d\n", item.ID, pos)
			return nil
		},
	}
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
// (attaching spec_id when given) as a single locked write.
func newTodoNextCmd() *cobra.Command {
	var specID string
	cmd := &cobra.Command{
		Use:   "next [<n>]",
		Short: "List queued cards (bare) or pick one (<n>)",
		Long: `Bare ` + "`moai todo next`" + ` prints the queued items oldest-first as
read-only candidates — the selection remains the operator's act performed
through the lead session's question channel. ` + "`moai todo next <n> [--spec <SPEC-ID>]`" + `
marks the addressed item picked (attaching spec_id when given) as one
locked write.`,
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
			if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
				for i := range rec.Items {
					if rec.Items[i].ID == id {
						rec.Items[i].State = kanban.BacklogStatePicked
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
			_, _ = fmt.Fprintf(out, "picked %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&specID, "spec", "",
		"SPEC-ID to attach when picking (recorded verbatim)")
	return cmd
}

// normalizeTodoRef maps a bare <n> argument to the item id t<n>; an explicit
// id (t<n>) passes through unchanged (REQ-TODO-004's normalization rule).
func normalizeTodoRef(arg string) string {
	if !strings.HasPrefix(arg, "t") {
		return "t" + arg
	}
	return arg
}

func init() {
	rootCmd.AddCommand(newTodoCmd())
}
