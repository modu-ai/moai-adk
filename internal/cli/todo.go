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
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// init wires internal/cli's existing userHomeDirFn test-injection seam
// through to kanban.HomeDirFn, the equivalent seam the relocated queue-root
// resolution owns (SPEC-WEB-TODO-QUEUE-001 M1). The closure closes over the
// package-level var by reference, so a test that reassigns userHomeDirFn at
// runtime is still observed by the resolution — the same pattern glm.go uses
// for glmcred.HomeDirFn.
func init() {
	kanban.HomeDirFn = func() (string, error) { return userHomeDirFn() }
}

// todoBacklogPath returns the backlog file location under root — the same
// path the todo skill and the dispatch protocol name (REQ-TODO-001). The root
// itself is resolved by resolveTodoQueueRoot.
//
// This is the ADOPTING form: the `moai todo` command path is where the
// one-time relocation of the legacy state directory belongs, because it is
// where the queue lock is already in play (REQ-TOSQ-015). The read-only
// surfaces — the console, the statusline — use the pure form and move nothing.
func todoBacklogPath(root string) string {
	return kanban.BacklogPathForRootAdopting(root)
}

// resolveTodoQueueRoot returns the directory the backlog queue hangs from
// for the `moai todo` command path: the PRIMARY checkout of the repository
// the launch context sits in, or a home-based fallback when git cannot
// answer — adopting a pre-existing project-local queue on that fallback
// branch, exactly as before.
//
// The resolution itself lives in internal/kanban since
// SPEC-WEB-TODO-QUEUE-001 M1, so the command layer and the web console share
// ONE resolution (a second implementation is a second chance to fork the
// queue). The command path takes the ADOPTING entry point; the console takes
// the pure one, which never writes.
func resolveTodoQueueRoot() string {
	return kanban.ResolveTodoQueueRootAdopting(resolveProjectDir())
}

// newTodoStore is the single constructor every todo verb goes through, so
// every verb resolves — and sees — the same queue file.
func newTodoStore() *kanban.BacklogStore {
	return kanban.NewBacklogStore(todoBacklogPath(resolveTodoQueueRoot()))
}

// todoLandedRef is the single place the todo surface resolves the ref the
// landing question is asked about, so the help text, the flag description, the
// refusal, and the query itself can never name different refs.
//
// It resolves against the SAME root the queue does — the primary checkout —
// because the queue and the integration branch are properties of one
// repository, not of whichever worktree the command happens to run in.
func todoLandedRef() string {
	return kanban.LandedRefFor(resolveTodoQueueRoot())
}

// todoLandedRefOnce resolves the landed ref at most once per process, and only
// when something actually asks for it.
//
// Resolving it eagerly is what made the resolution expensive out of all
// proportion to its use: cobra builds the WHOLE command tree at process start,
// so every `moai <anything>` invocation — `moai statusline`, once per render,
// included — paid a `git rev-parse` for each command that named the ref in its
// help text, for help it was never going to print. Worse, the root resolution
// takes the ADOPTING entry point, which can write to the filesystem; that has
// no business firing on a pure render path.
var todoLandedRefOnce = sync.OnceValue(todoLandedRef)

// withResolvedLandedRef defers the parts of cmd's help surface that name the
// landed ref until help is actually rendered, and applies them at most once.
//
// The help function is the hook because `--help` returns before RunE, so a
// PreRun hook would silently drop the resolved ref from the printed text. The
// usage function is hooked for the same reason on the error path, which prints
// flag usage without printing help.
//
// Both delegate up to the PARENT's function rather than naming a renderer,
// because the effective one differs between the fang-wrapped binary and the
// in-process test path; resolving it at call time keeps both intact.
func withResolvedLandedRef(cmd *cobra.Command, apply func(ref string)) {
	var once sync.Once
	resolve := func() { once.Do(func() { apply(todoLandedRefOnce()) }) }

	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		resolve()
		if p := c.Parent(); p != nil {
			p.HelpFunc()(c, args)
			return
		}
		_ = c.Usage()
	})
	cmd.SetUsageFunc(func(c *cobra.Command) error {
		resolve()
		if p := c.Parent(); p != nil {
			return p.UsageFunc()(c)
		}
		return nil
	})
}

// newTodoCmd creates the `moai todo` parent command.
func newTodoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "todo",
		Short: "Operate the kanban backlog queue",
		Long: `Operate the kanban backlog queue at .moai/state/todo/backlog.db.

The queue resolves against the PRIMARY checkout even when this command runs
inside a linked worktree — one repository, one queue; a card worktree adds
to and reads the same store the lead and the foreman loop see. A project
without git metadata keeps its queue at ~/.moai/todo/<project-key>/backlog.db
instead. A backlog.json sitting beside the database is NOT the queue — it is
an export or a legacy leftover, and its contents can be arbitrarily stale.

The backlog is the operator's queue: entry into the board is the operator's
act (add), and picking the next card is the operator's act too (next <n>).
Mutations serialize on a sibling cross-process lock; reads are lock-free.

A bare invocation renders the queue, which is the form the skill surface and
workflows/todo.md both document; ` + "`moai todo list`" + ` remains valid and prints the
same thing. A single unknown token stays an error (a mistyped verb must not
become a card), while a phrase of two or more words falls through to add:
` + "`moai todo fix the flaky gate`" + ` adds that card. A one-word card therefore
needs the explicit add verb — the price of keeping typos loud.

One fallthrough shape is refused outright: a verb-shaped first token followed
by a card id (` + "`moai todo pick t151`" + `) is a mistyped verb, not a card, and
becomes an error naming the known verbs. A card text that merely mentions an
id later in the sentence still falls through, and ` + "`moai todo add \"<text>\"`" + `
adds any text verbatim.`,
		Args: func(cmd *cobra.Command, args []string) error {
			// t69 fallthrough: two or more words are natural language → add.
			// Deliberate failure modes: a single token (the mistyped verb
			// "lst", or a one-word card like "docs") stays an error — a
			// mistyped verb must not silently become a card — so a one-word
			// card needs the explicit add verb; conversely a mistyped verb
			// followed by more words ("lst the queue") DOES become a card,
			// the accepted cost of the fallthrough.
			//
			// t203 narrows that accepted cost where the cost is highest:
			// a verb-shaped first token addressing a card id is a mistyped
			// verb, not a card — see todoMistypedVerbGuard.
			if err := todoMistypedVerbGuard(cmd, args); err != nil {
				return err
			}
			if len(args) > 1 {
				return nil
			}
			return cobra.NoArgs(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runTodoList(cmd, false, false, todoListDefaultLimit)
			}
			return runTodoAddAppend(cmd, strings.Join(args, " "), false)
		},
		GroupID: "tools",
	}
	cmd.AddCommand(newTodoAddCmd(), newTodoListCmd(), newTodoDoneCmd(), newTodoUndoneCmd(), newTodoNextCmd(),
		newTodoUnpickCmd(), newTodoEditCmd(), newTodoMoveCmd(),
		newTodoDropCmd(), newTodoUndropCmd(),
		newTodoAnalyzeCmd(), newTodoRelateCmd(), newTodoUnrelateCmd(), newTodoWhyCmd(),
		newTodoPRCmd(), newTodoExportJSONCmd(), newTodoHistoryCmd())
	return cmd
}

// todoVerbShaped matches a first token that reads as a command verb: one
// lowercase ASCII word, optionally hyphenated. Bounded in length so a long
// word in a card text cannot pass for a verb.
var todoVerbShaped = regexp.MustCompile(`^[a-z][a-z-]{1,15}$`)

// todoCardIDShaped matches the id form the queue issues (`t<decimal>`, see
// kanban.BacklogStore). Deliberately NOT the looser reference form the verbs
// accept (`done 151` normalizes a bare number): a bare number is ordinary
// card text ("fix 3 flaky tests"), while an explicit `t151` in second
// position is an address.
var todoCardIDShaped = regexp.MustCompile(`^t\d+$`)

// todoMistypedVerbGuard refuses the one fallthrough shape that is almost
// never a card: a verb-shaped first token followed by a card id
// (`moai todo pick t151`). Cobra routes a REGISTERED verb to its subcommand,
// so anything reaching the parent is an unregistered word — and a word
// addressing a card id is a mistyped verb whose silent conversion into a
// card is the data pollution #1597 reports.
//
// The t69 usability trade-off is preserved everywhere else: a natural-language
// card still falls through to add, including one that merely mentions a card
// id later in the sentence ("fix the drift found in t151"). Only the exact
// verb-then-id shape is refused, and `moai todo add "<text>"` remains the
// escape hatch for a card that genuinely reads that way.
func todoMistypedVerbGuard(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return nil
	}
	if !todoVerbShaped.MatchString(args[0]) || !todoCardIDShaped.MatchString(args[1]) {
		return nil
	}
	phrase := strings.Join(args, " ")
	return fmt.Errorf(
		"todo: %q is not a todo verb and %q is a card id — refusing to create a card named %q.\n"+
			"Known verbs: %s\nTo add this text as a card anyway: moai todo add %q",
		args[0], args[1], phrase, strings.Join(todoVerbNames(cmd), ", "), phrase)
}

// todoVerbNames lists the registered verb names, so the guard's message is
// derived from the command tree rather than from a hand-maintained list that
// drifts as verbs are added.
func todoVerbNames(cmd *cobra.Command) []string {
	if cmd == nil {
		return nil
	}
	names := make([]string, 0, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		if sub.Name() == "help" || sub.Name() == "completion" || sub.Hidden {
			continue
		}
		names = append(names, sub.Name())
	}
	sort.Strings(names)
	return names
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

// todoListDefaultLimit is the list render's default bound — the same shape
// (and value) as the history verb's REQ-TAQ-007 contract: a bounded read is
// the default, --limit raises or lowers it, --limit 0 lifts it entirely,
// and a truncated listing states the withheld count on stderr because a
// truncated read must never be mistaken for a complete one.
const todoListDefaultLimit = 20

// runTodoList renders the backlog lock-free. It backs both entry points —
// the bare `moai todo` and the explicit `moai todo list` — so the two cannot
// drift apart in output.
//
// The default view renders live cards only (queued + picked) and collapses
// the dropped set into one count line naming the recovery path: a dropped
// card never leaves rec.Items, so rendering it forever made the list length
// diverge from the queue's actual load. `--dropped` renders the discarded
// set instead — the surface `undrop` needs to find a card and its reason
// (t153's exact-reversal contract is untouched; only the render filters).
//
// The render is bounded at limit rows (t403): an unbounded render pushed the
// truncation downstream to the reading harness, where rows vanished from the
// visible output with no withheld count. `--json` ignores the limit — the
// structured record is the full read, and a bounded JSON would be the same
// silent truncation.
func runTodoList(cmd *cobra.Command, jsonOutput bool, droppedOnly bool, limit int) error {
	store := newTodoStore()
	// REQ-BJD-002 — probed before the read, because Load adopts (see
	// todo_disclosure.go). stderr only: stdout is what the foreman reads.
	_ = discloseQueueLayout(cmd, "todo")
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
	if limit < 0 {
		return fmt.Errorf("todo list: --limit must be >= 0 (got %d)", limit)
	}
	var visible []kanban.BacklogItem
	dropped := 0
	for _, it := range rec.Items {
		isDropped := it.State == kanban.BacklogStateDropped
		if isDropped {
			dropped++
		}
		if isDropped != droppedOnly {
			continue
		}
		visible = append(visible, it)
	}
	shown := len(visible)
	if limit > 0 && limit < shown {
		shown = limit
	}
	for _, it := range visible[:shown] {
		_, _ = fmt.Fprintf(out, "%s\t%s\t%s\n", it.ID, it.State, it.Text)
		for _, f := range rec.Findings {
			if !f.Names(it.ID) {
				continue
			}
			_, _ = fmt.Fprintln(out, todoFindingLine(rec, it.ID, f))
		}
	}
	if droppedOnly && shown == 0 {
		_, _ = fmt.Fprintln(out, "no dropped cards")
		return nil
	}
	if !droppedOnly && dropped > 0 {
		_, _ = fmt.Fprintf(out, "%d dropped (hidden — see: moai todo list --dropped)\n", dropped)
	}
	if withheld := len(visible) - shown; withheld > 0 {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"list: %d rows withheld — showing %d of %d (--limit 0 lists all)\n",
			withheld, shown, len(visible))
	}
	return nil
}

// newTodoListCmd — `moai todo list [--json] [--dropped] [--limit <n>]`
// (REQ-TODO-003): render the queue lock-free; --json emits the structured
// records, --dropped renders the discarded set the default view hides behind
// a count line, and --limit bounds the row render (0 = unbounded; ignored
// with --json).
func newTodoListCmd() *cobra.Command {
	var jsonOutput bool
	var droppedOnly bool
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Render the backlog queue (lock-free)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTodoList(cmd, jsonOutput, droppedOnly, limit)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false,
		"Emit the backlog records as JSON on stdout")
	cmd.Flags().BoolVar(&droppedOnly, "dropped", false,
		"Render only the dropped cards (the default view hides them behind a count line)")
	cmd.Flags().IntVar(&limit, "limit", todoListDefaultLimit,
		"Maximum rows to render (0 = unbounded; ignored with --json)")
	return cmd
}

// newTodoDoneCmd — `moai todo done <n>` (REQ-TODO-004): take the addressed
// row out of the live queue under the lock. A bare <n> is normalized to the
// item id t<n>; the explicit id is the preferred form because queue positions
// move under concurrent adds.
//
// The row is ARCHIVED rather than discarded (SPEC-TODO-DESTRUCTIVE-GUARD-001
// REQ-TDG-003), so `undone` can put it back. `done` keeps its plain meaning —
// the card leaves the queue and no live reader sees it again — but it is no
// longer the one destructive verb with no way back.
//
// Two opt-in guards refuse INSIDE the mutation callback, so each inherits
// Mutate's byte-identity-on-refusal contract: `--expect <prefix>` follows the
// convention `next`, `edit`, `drop` and `undrop` already carry, and
// `--require-landed` asks the landing question with its limit stated.
func newTodoDoneCmd() *cobra.Command {
	var expect string
	var requireLanded bool
	cmd := &cobra.Command{
		Use:   "done <n>",
		Short: "Archive a card out of the backlog queue by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := normalizeTodoRef(args[0])
			store := newTodoStore()
			// Unknown until a query answers otherwise. Absent the flag no
			// query runs at all, and `unknown` is the honest report of that.
			verdict := kanban.LandingUnknown
			if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
				// Refused mutations below: Mutate writes nothing, so the
				// record stays byte-identical on every one of them.
				at := -1
				for i := range rec.Items {
					if rec.Items[i].ID == id {
						at = i
						break
					}
				}
				if at < 0 {
					return fmt.Errorf("no backlog item %s", id)
				}
				if expect != "" && !strings.HasPrefix(rec.Items[at].Text, expect) {
					return fmt.Errorf("backlog item %s is %q, not matching --expect %q",
						id, todoTextPrefix(rec.Items[at].Text), expect)
				}
				if requireLanded {
					answer, err := todoRequireLanded(cmd, id)
					if err != nil {
						return err
					}
					verdict = answer
				}
				return rec.ArchiveCard(id)
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			// One line per act, carrying exactly one landing verdict — a second
			// line would give an operator script two records for one event.
			// The suffix preserves the `done <id>` prefix every existing
			// reader keys off.
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "done %s landing=%s\n", id, verdict)
			return nil
		},
	}
	cmd.Flags().StringVar(&expect, "expect", "",
		"Refuse unless the card text starts with this prefix")
	cmd.Flags().BoolVar(&requireLanded, "require-landed", false, "")
	withResolvedLandedRef(cmd, func(landedRef string) {
		cmd.Long = todoDoneLong(landedRef)
		cmd.Flags().Lookup("require-landed").Usage =
			"Refuse unless a commit on " + landedRef + " names the card (opt-in; see --help for its limit)"
	})
	return cmd
}

// todoDoneLong renders `todo done`'s help body against the ref the landing
// guard will actually ask about. Resolved lazily — see withResolvedLandedRef.
func todoDoneLong(landedRef string) string {
	return `Move the addressed card out of the live queue as one locked write.

The card and every finding naming it move into the archive rather than being
discarded, so ` + "`moai todo undone <n>`" + ` restores both. Archived rows are invisible
to every live-queue reader (` + "`list`, `next`, `why`, `analyze`" + `, the counts).

` + "`--expect <prefix>`" + ` refuses unless the addressed card's text starts with the
prefix — the guard against closing the wrong card.

` + "`--require-landed`" + ` refuses unless a commit on ` + landedRef + ` names the card.
It is OPT-IN and honestly limited: it answers "has anything naming this card
landed on that ref", NOT "has this card's last step landed", so it cannot tell a
run commit from a sync commit. Absent the flag no landing query runs at all.

Every successful invocation prints one landing verdict on stdout —
` + "`done <id> landing=landed|not-landed|unknown`" + `. Without the flag the verdict is
` + "`unknown`" + `, because no query ran: "the guard passed" and "the guard did not
run" are different facts and no longer the same bytes.`
}

// todoRequireLanded answers the opt-in landing question for id.
//
// It refuses ONLY on positive evidence of not-landed, and PROCEEDS when the
// answer is inconclusive — no git, no such ref, a query error. That asymmetry
// is the point: the guard exists to catch a card closed before its work
// shipped, and a guard that also blocked every machine without the ref would
// be refusing on the absence of evidence rather than on evidence.
//
// The limit is stated rather than hidden. The predicate asks whether ANY
// commit on the ref names the card, so a card whose run commit landed reads as
// landed even though its sync commit has not — the exact case that motivated
// this SPEC. Making it answer the right question needs a persisted
// landing-state field, which is a separate card's scope; this ships the seam
// and says plainly what it can and cannot answer (spec.md §A.4).
func todoRequireLanded(cmd *cobra.Command, id string) (kanban.LandingAnswer, error) {
	q := kanban.GitLandedQuerier{Run: todoRunCommand, Ref: todoLandedRef()}
	answer, err := q.Landed(id)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"note: --require-landed could not answer for %s against %s (%v) — proceeding, because an unanswerable query is not evidence of not-landed\n",
			id, q.LandedRef(), err)
		return kanban.LandingUnknown, nil
	}
	if answer == kanban.LandingNotLanded {
		return answer, fmt.Errorf("backlog item %s is named by no commit on %s — --require-landed refuses "+
			"(the check asks whether anything naming the card has landed on that ref, not whether the card's last step has)",
			id, q.LandedRef())
	}
	return answer, nil
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
