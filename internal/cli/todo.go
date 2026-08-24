// todo.go — SPEC-KANBAN-TODO-CLI-001 M2: the `moai todo` command surface.
//
// Thin cobra wiring over internal/kanban.BacklogStore: every mutation
// delegates to the store's locked Mutate path, reads go through the
// lock-free Load. The verbs serve the kanban dispatch protocol's entry rule
// (`/moai todo` is the operator's act — the lead never picks for the
// operator): `add` and `done` mutate, `list` and bare `next` observe, and
// `next <n> [--spec]` records the operator's pick as one locked write.
//
// Queue residence (t106): the backlog file hangs from the PRIMARY checkout
// of the repository, never from the working tree the command happens to
// run in — see resolveTodoQueueRoot.
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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// todoBacklogPath returns the backlog file location under root — the same
// .moai/state/kanban/backlog.json the todo skill and the dispatch protocol
// name (REQ-TODO-001). The root itself is resolved by resolveTodoQueueRoot.
func todoBacklogPath(root string) string {
	return filepath.Join(root, ".moai", "state", "kanban", "backlog.json")
}

// resolveTodoQueueRoot returns the directory the backlog queue hangs from:
// the PRIMARY checkout of the repository the launch context sits in, or a
// home-based fallback when git cannot answer.
//
// The queue is the delegation channel between sessions — the lead, the
// foreman loop, and the operator's picks all read and write one queue. A
// linked worktree holding its own .moai/state would fork that channel: an
// `add` from inside a card worktree lands in a queue file the primary
// checkout never sees (measured 2026-08-17: 30 queued cards on the
// primary, "queue is empty" from a linked worktree). The primary is
// therefore resolved through the repository itself: git's common directory
// is shared by every checkout and its parent IS the primary checkout's
// root, from any worktree and from the primary alike (internal/core/git
// checkout.go) — the same discrimination branch_guard.go applies, with no
// file moving.
//
// Fail-open direction: an unresolvable git context (no git binary, not a
// repository) keeps the queue usable via the home-based fallback rather
// than erroring — a project without git metadata still gets exactly one
// queue, keyed under ~/.moai/todo/ so two such projects cannot collide.
func resolveTodoQueueRoot() string {
	base := resolveProjectDir()
	if dirs, err := gitcore.ResolveGitDirs(base); err == nil && dirs.CommonDir != "" {
		return filepath.Dir(dirs.CommonDir)
	}
	return fallbackTodoQueueRoot(base)
}

// fallbackTodoQueueRoot returns the home-based queue root for a launch
// context git could not resolve to a primary checkout. userHomeDirFn is the
// package's existing test-injection seam, so tests override the home
// lookup instead of mutating the real HOME.
//
// The directory is named for the command that owns the queue (`moai todo` —
// no `moai kanban` command exists), deliberately NOT "kanban": the
// .moai/state/kanban/ directory also holds per-session kanban records
// (<uuid>.json) owned by internal/kanban, and a "kanban" fallback name would
// read as moving those too — a scope this queue never touches.
//
// Before the fallback's queue is created, adoptLocalTodoQueue carries an
// existing project-local queue over (adopt-not-shadow): the fallback's first
// run must present the cards the project already has, never an empty queue
// shadowing them.
func fallbackTodoQueueRoot(base string) string {
	if base == "" {
		base = "."
	}
	home, err := userHomeDirFn()
	if err != nil {
		// No home resolvable either: keep the queue inside the project
		// directory rather than failing the command outright.
		return filepath.Join(base, ".moai", "state", "kanban")
	}
	root := filepath.Join(home, ".moai", "todo", todoQueueProjectKey(base))
	adoptLocalTodoQueue(base, root)
	return root
}

// adoptLocalTodoQueue moves a pre-existing project-local queue into a fresh
// fallback root, so the fallback's first run adopts the project's cards
// instead of shadowing them behind an empty queue (the lossless-migration
// requirement: item count and states must survive the cutover).
//
// Best-effort throughout — a failure leaves the local queue exactly where it
// was and the fallback simply starts empty THIS run; the data is never
// destroyed, so a later run can adopt it again once the obstruction clears.
// When the fallback already has a queue file the local one is left untouched
// (adopted on an earlier run; a local file reappearing after that is a
// downgrade-era snapshot the populated fallback deliberately ignores).
func adoptLocalTodoQueue(base, fallbackRoot string) {
	local := todoBacklogPath(base)
	target := filepath.Join(fallbackRoot, "backlog.json")
	if _, err := os.Stat(target); err == nil {
		return
	}
	if _, err := os.Stat(local); err != nil {
		return
	}
	if err := os.MkdirAll(fallbackRoot, 0o755); err != nil {
		return
	}
	// Same-volume rename is atomic and leaves no duplicate behind.
	if err := os.Rename(local, target); err == nil {
		return
	}
	// Cross-volume (EXDEV) or rename-refusing filesystem: copy the bytes and
	// KEEP the original — deletion is the one outcome the lossless
	// requirement forbids, and a leftover original is inert (the populated
	// fallback wins on every later run).
	data, err := os.ReadFile(local)
	if err != nil {
		return
	}
	_ = os.WriteFile(target, data, 0o600)
}

// todoQueueProjectKey derives the fallback queue's directory name from the
// launch directory: a readable base name plus a short digest of the
// absolute path. Two distinct projects sharing a base name still occupy
// two keys, and the mapping is deterministic across runs.
func todoQueueProjectKey(base string) string {
	abs, err := filepath.Abs(base)
	if err != nil {
		abs = filepath.Clean(base)
	}
	sum := sha256.Sum256([]byte(abs))
	return fmt.Sprintf("%s-%x", filepath.Base(abs), sum[:4])
}

// newTodoStore is the single constructor every todo verb goes through, so
// every verb resolves — and sees — the same queue file.
func newTodoStore() *kanban.BacklogStore {
	return kanban.NewBacklogStore(todoBacklogPath(resolveTodoQueueRoot()))
}

// newTodoCmd creates the `moai todo` parent command.
func newTodoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "todo",
		Short: "Operate the kanban backlog queue",
		Long: `Operate the kanban backlog queue at .moai/state/kanban/backlog.json.

The queue resolves against the PRIMARY checkout even when this command runs
inside a linked worktree — one repository, one queue; a card worktree adds
to and reads the same file the lead and the foreman loop see. A project
without git metadata keeps its queue at ~/.moai/todo/<project-key>/backlog.json
instead.

The backlog is the operator's queue: entry into the board is the operator's
act (add), and picking the next card is the operator's act too (next <n>).
Mutations serialize on a sibling cross-process lock; reads are lock-free.

A bare invocation renders the queue, which is the form the skill surface and
workflows/todo.md both document; ` + "`moai todo list`" + ` remains valid and prints the
same thing. A single unknown token stays an error (a mistyped verb must not
become a card), while a phrase of two or more words falls through to add:
` + "`moai todo fix the flaky gate`" + ` adds that card. A one-word card therefore
needs the explicit add verb — the price of keeping typos loud.`,
		Args: func(cmd *cobra.Command, args []string) error {
			// t69 fallthrough: two or more words are natural language → add.
			// Deliberate failure modes: a single token (the mistyped verb
			// "lst", or a one-word card like "docs") stays an error — a
			// mistyped verb must not silently become a card — so a one-word
			// card needs the explicit add verb; conversely a mistyped verb
			// followed by more words ("lst the queue") DOES become a card,
			// the accepted cost of the fallthrough.
			if len(args) > 1 {
				return nil
			}
			return cobra.NoArgs(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runTodoList(cmd, false)
			}
			return runTodoAddAppend(cmd, strings.Join(args, " "), false)
		},
		GroupID: "tools",
	}
	cmd.AddCommand(newTodoAddCmd(), newTodoListCmd(), newTodoDoneCmd(), newTodoNextCmd(),
		newTodoUnpickCmd(), newTodoEditCmd(), newTodoMoveCmd(),
		newTodoDropCmd(), newTodoUndropCmd(),
		newTodoAnalyzeCmd(), newTodoRelateCmd(), newTodoUnrelateCmd(), newTodoWhyCmd(),
		newTodoPRCmd())
	return cmd
}

// newTodoAddCmd — `moai todo add "<text>"` (REQ-TODO-002): append under the
// lock, print the issued id and its 1-based queue position. `--pick` (t71)
// folds the pick into the same locked write.
func newTodoAddCmd() *cobra.Command {
	var pick bool
	var force bool
	cmd := &cobra.Command{
		Use:   "add <text>",
		Short: "Append a card to the backlog queue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[0]
			if strings.TrimSpace(text) == "" {
				return fmt.Errorf("todo add: text must be non-empty")
			}
			if pick {
				return runTodoAddPick(cmd, newTodoStore(), text, force)
			}
			return runTodoAddAppend(cmd, text, force)
		},
	}
	cmd.Flags().BoolVar(&pick, "pick", false,
		"Append AND mark picked as one locked write, printing the issued id")
	cmd.Flags().BoolVar(&force, "force", false,
		"Admit a card the analyser reads as an exact duplicate, recording that it was forced")
	return cmd
}

// runTodoAddAppend is the plain-add body shared by `todo add <text>` and the
// parent's natural-language fallthrough (t69): non-empty guard, locked
// append, "<id> <position>" stdout line. `--pick` stays add-only — the
// fallthrough path has no flags.
func runTodoAddAppend(cmd *cobra.Command, text string, force bool) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("todo add: text must be non-empty")
	}
	var item kanban.BacklogItem
	var pos int
	err := newTodoStore().Mutate(func(rec *kanban.BacklogRecord) error {
		var mutErr error
		item, pos, mutErr = appendAnalyzedCard(rec, text, kanban.BacklogStateQueued, force)
		return mutErr
	})
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %d\n", item.ID, pos)
	return nil
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
func runTodoAddPick(cmd *cobra.Command, store *kanban.BacklogStore, text string, force bool) error {
	var item kanban.BacklogItem
	err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		var mutErr error
		item, _, mutErr = appendAnalyzedCard(rec, text, kanban.BacklogStatePicked, force)
		return mutErr
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
	store := newTodoStore()
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
		for _, f := range rec.Findings {
			if !f.Names(it.ID) {
				continue
			}
			_, _ = fmt.Fprintln(out, todoFindingLine(rec, it.ID, f))
		}
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
			store := newTodoStore()
			if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
				for i, it := range rec.Items {
					if it.ID == id {
						rec.Items = append(rec.Items[:i], rec.Items[i+1:]...)
						// A finding outliving its subject points at a card
						// the operator can no longer see, so the card and
						// every finding naming it leave together.
						rec.RemoveFindingsNaming(id)
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
			store := newTodoStore()
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
			store := newTodoStore()
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
