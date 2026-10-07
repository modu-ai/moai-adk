// todo.go — SPEC-KANBAN-TODO-CLI-001 M2: the `moai todo` command surface.
//
// Thin cobra wiring over internal/factory.BacklogStore: every mutation
// delegates to the store's locked Mutate path, reads go through the
// lock-free Load. The verbs serve the factory dispatch protocol's entry rule
// (`/moai todo` is the operator's act — the leader never picks for the
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
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// init wires internal/cli's existing userHomeDirFn test-injection seam
// through to factory.HomeDirFn, the equivalent seam the relocated queue-root
// resolution owns (SPEC-WEB-TODO-QUEUE-001 M1). The closure closes over the
// package-level var by reference, so a test that reassigns userHomeDirFn at
// runtime is still observed by the resolution — the same pattern glm.go uses
// for glmcred.HomeDirFn.
func init() {
	factory.HomeDirFn = func() (string, error) { return userHomeDirFn() }
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
	return factory.BacklogPathForRootAdopting(root)
}

// resolveTodoQueueRoot returns the directory the backlog queue hangs from
// for the `moai todo` command path: the PRIMARY checkout of the repository
// the launch context sits in, or a home-based fallback when git cannot
// answer — adopting a pre-existing project-local queue on that fallback
// branch, exactly as before.
//
// The resolution itself lives in internal/factory since
// SPEC-WEB-TODO-QUEUE-001 M1, so the command layer and the web console share
// ONE resolution (a second implementation is a second chance to fork the
// queue). The command path takes the ADOPTING entry point; the console takes
// the pure one, which never writes.
func resolveTodoQueueRoot() string {
	return factory.ResolveTodoQueueRootAdopting(resolveProjectDir())
}

// warnTempOriginQueueRefusal surfaces the temporary-origin refusal on the
// COMMAND path (SPEC-TODO-HOME-TEMP-GUARD-001, REQ-THG-006): the queue-root
// resolution declined to create a home queue under ~/.moai/db because the
// launch directory lives inside a temporary root, and the run continues
// against the project-local queue instead.
//
// Three things the guidance must carry, because a refusal that reads as a
// silent success is indistinguishable from the bug it replaced: WHICH temp
// root matched, WHICH root the run continues against, and that the run is in
// fact continuing. The exit code is unchanged — refusing the home queue is
// already the whole of the protection, so failing the command would withdraw
// working behaviour from every script that runs `moai todo` inside a temp
// directory without preventing anything further.
//
// Silent on every other path, the console included: this is called only from
// the command's PersistentPreRun, and the pure resolver the web console
// imports neither writes nor speaks.
func warnTempOriginQueueRefusal(cmd *cobra.Command) {
	substitute, matched, refused := factory.TempOriginRefusal(resolveProjectDir())
	if !refused {
		return
	}
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
		"moai todo: the launch directory is inside the temporary root %s, so no home queue was created under ~/.moai/db/<project-key>/todo; continuing against the project-local queue at %s\n",
		matched, substitute)
}

// newTodoStore is the single constructor every todo verb goes through, so
// every verb resolves — and sees — the same queue file.
func newTodoStore() *factory.BacklogStore {
	return factory.NewBacklogStore(todoBacklogPath(resolveTodoQueueRoot()))
}

// Observational commands must not relocate a legacy queue while constructing
// their store, before LoadPure even gets a chance to preserve it.
func newTodoReadStore() *factory.BacklogStore {
	root := factory.ResolveTodoQueueRoot(resolveProjectDir())
	return factory.NewBacklogStore(factory.BacklogPathForRoot(root))
}

// todoStoreAt and todoReadStoreAt anchor the queue at an explicit root — the
// MCP tool path, whose caller names its tree with the project_root argument
// (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-024). The CLI verbs keep resolving
// through resolveTodoQueueRoot; the two shapes share the same path builders,
// so a root both surfaces agree on sees the same queue file.
func todoStoreAt(root string) *factory.BacklogStore {
	return factory.NewBacklogStore(todoBacklogPath(root))
}

func todoReadStoreAt(root string) *factory.BacklogStore {
	return factory.NewBacklogStore(factory.BacklogPathForRoot(root))
}

// todoLandedRef is the single place the todo surface resolves the ref the
// landing question is asked about, so the help text, the flag description, the
// refusal, and the query itself can never name different refs.
//
// It resolves against the SAME root the queue does — the primary checkout —
// because the queue and the integration branch are properties of one
// repository, not of whichever worktree the command happens to run in.
func todoLandedRef() string {
	return factory.LandedRefFor(factory.ResolveTodoQueueRoot(resolveProjectDir()))
}

// todoLandedRefResolved is todoLandedRef with its provenance: which chain
// level answered. The `todo done` verdict discloses levels below the
// configured key (REQ-TLA-011) — a ref the repository supplied through its
// own recorded default rather than through configuration is the exceptional
// path, and a silent fallback is exactly how the wrong-ref answer hid.
func todoLandedRefResolved() (string, factory.LandedRefLevel) {
	return factory.LandedRefForWithLevel(factory.ResolveTodoQueueRoot(resolveProjectDir()))
}

// todoRefLevelSource names where a chain level's answer came from, for the
// disclosure line.
func todoRefLevelSource(level factory.LandedRefLevel) string {
	switch level {
	case factory.LandedRefOriginHEAD:
		return "refs/remotes/origin/HEAD"
	default:
		return "the compiled-in default"
	}
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
//
// Construction is serialized behind a mutex: the flag registrations below
// bind package-level variables (`todoAutoFlag` · `todoAutoWait`), so pflag
// writes those globals at REGISTRATION time — and concurrent constructors
// (the landed-verb concurrency test builds one command per goroutine) race
// on them. The per-call command instances themselves share nothing.
var newTodoCmdMu sync.Mutex

func newTodoCmd() *cobra.Command {
	newTodoCmdMu.Lock()
	defer newTodoCmdMu.Unlock()
	cmd := &cobra.Command{
		Use:   "todo",
		Short: "Operate the backlog queue",
		Long: `Operate the backlog queue at ~/.moai/db/<project-key>/todo/backlog.db.

The queue resolves against the PRIMARY checkout even when this command runs
inside a linked worktree — one repository, one queue; a card worktree adds
to and reads the same store the leader and the foreman loop see. A project
without git metadata uses the same project-keyed home layout. A backlog.json
left at the former project-local path is NOT the queue — it is
an export or a legacy leftover, and its contents can be arbitrarily stale.

The backlog is the operator's queue: entry into the board is the operator's
act (add), and picking the next card is the operator's act too (next <n>).
Mutations serialize on a sibling cross-process lock; reads are lock-free.

A bare invocation renders the queue, which is the form the skill surface and
workflows/gtd.md both document; ` + "`moai todo list`" + ` remains valid and prints the
same thing. A single unknown token stays an error (a mistyped verb must not
become a card), while a phrase of two or more words falls through to add:
` + "`moai todo fix the flaky gate`" + ` adds that card. A one-word card therefore
needs the explicit add verb — the price of keeping typos loud.

One fallthrough shape is refused outright: a verb-shaped first token followed
by a card address (` + "`moai todo pick t151`" + `, ` + "`moai todo pick 151`" + `) is a mistyped
verb, not a card, and becomes an error naming the known verbs. The address
forms are the ones the verbs themselves accept — an explicit ` + "`t<n>`" + ` anywhere,
or a bare ` + "`<n>`" + ` when it is the whole remainder or the first token is a
near-miss of a real verb (` + "`moai todo drp 401 stale`" + `). A number after an
ordinary word is still card text (` + "`moai todo fix 3 flaky tests`" + `). A card text that merely
mentions an id later in the sentence still falls through, and
` + "`moai todo add \"<text>\"`" + ` adds any text verbatim.`,
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
			if todoAutoFlag {
				if len(args) > 0 {
					return fmt.Errorf("--auto takes no card arguments; the invocation is the operator's batch approval of the queue, never an admission")
				}
				opts := autoOptions{
					wait:     todoAutoWait,
					liveness: newAutoLiveness(),
					landed:   todoAutoLandedLookup,
					jevRank:  todoAutoJevRanker,
					quota:    todoAutoQuotaLine,
				}
				if todoLaneSession() {
					// Card t1554: a lane session's --auto invocation IS the
					// cycle — the lease-based lane form. Its only queue write
					// is the nominated lease edge, so the REQ-SD-015 queue
					// guarantee holds by construction.
					return runAutoLaneCycle(cmd.Context(), resolveTodoQueueRoot(), cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
				}
				return runAutoCycle(cmd.OutOrStdout(), newTodoStore(), resolveTodoQueueRoot(), opts)
			}
			if len(args) == 0 {
				return runTodoList(cmd, false, false, todoListDefaultLimit)
			}
			return runTodoAddAppend(cmd, strings.Join(args, " "), false, todoCardDecider)
		},
		// PersistentPreRun fires once per `moai todo ...` invocation, for the
		// parent and every subcommand alike, which is why the guidance lives
		// here rather than inside resolveTodoQueueRoot: that helper is called
		// several times per run (store, landed ref, ...) and would repeat the
		// notice once per call.
		//
		// The lane queue guard (REQ-SD-015) rides the same hook, ahead of
		// every subcommand's RunE, so a refused lane session never reaches a
		// mutation and the queue file stays byte-identical.
		PersistentPreRunE: func(run *cobra.Command, args []string) error {
			if err := todoRefuseLaneMutation(todoTreeRoot(run), run, args); err != nil {
				return err
			}
			warnTempOriginQueueRefusal(run)
			return nil
		},
		GroupID: "tools",
	}
	cmd.AddCommand(newTodoAddCmd(), newTodoListCmd(), newTodoDoneCmd(), newTodoUndoneCmd(), newTodoNextCmd(),
		newTodoClaimCmd(),
		newTodoUnpickCmd(), newTodoEditCmd(), newTodoMoveCmd(),
		newTodoDropCmd(), newTodoUndropCmd(), newTodoMergeCmd(),
		newTodoHoldCmd(), newTodoUnholdCmd(),
		newTodoAnalyzeCmd(), newTodoRelateCmd(), newTodoUnrelateCmd(), newTodoWhyCmd(), newTodoTraceCmd(),
		newTodoPRCmd(), newTodoLandedCmd(), newTodoAutoDoneCmd(), newTodoExportJSONCmd(), newTodoHistoryCmd(),
		newTodoShowCmd(), newTodoTriageCmd())
	cmd.Flags().BoolVar(&todoAutoFlag, "auto", false,
		"process the queue serially: pick one card, dispatch one isolated worker, judge completion on disk evidence, then accept the next; the invocation is the operator's batch approval of the queue and nothing else; the queued candidates are ranked first (a Jev signal when available, else recorded priority and readiness), and only a card whose text begins with the [보류 marker is demoted — a hold stated in prose without the marker is not (the structural hold is moai todo hold); a lane session runs the same cycle through the lease edges — its only queue write is the nominated lease, and completion stays with the existing completion path")
	cmd.Flags().DurationVar(&todoAutoWait, "auto-wait", 30*time.Minute,
		"per-card deadline for the worker evidence file before the card is unpicked with a labelled non-finding")
	return cmd
}

// todoAutoFlag / todoAutoWait back the `--auto` serial-processing cycle. They
// live on the parent command so the gtd compatibility spelling (the same verb
// tree, NewGTDCommand) carries them identically — one implementation, both
// entry points.
var (
	todoAutoFlag bool
	todoAutoWait time.Duration
)

// The live seams of the `--auto` ranking stage (SPEC-TODO-AUTO-PRIORITY-001
// plan D-5). runAutoCycle leaves a nil seam inert; this is where production
// installs the live ones. They are variables so a test that reaches the real
// command path replaces both and stays hermetic: the landed seam runs `gh`.
var (
	todoAutoLandedLookup autoLandedLookup = liveAutoLandedLookup
	todoAutoJevRanker    autoJevRanker    = liveAutoJevRanker
	// todoAutoQuotaLine is the quota steering seam's live value
	// (SPEC-QUOTA-AWARE-SCHEDULING-001): the shared pressure evaluation plus the
	// lane inventory, read-only. A variable for the same reason as the two above.
	todoAutoQuotaLine autoQuotaLine = factoryQuotaSteering
)

// todoLaneReadOnlyVerbs is the REQ-SD-015 read-only allowlist: the only
// `moai todo` forms a lane session may run. Everything else — including
// `next <n>`, which the operator's pick path shares — is refused; the one
// queue write a lane performs is the promotion inside `moai factory next`
// (OD-1, operator-authorized for the self-dispatch lane mode).
var todoLaneReadOnlyVerbs = map[string]bool{
	"list":    true,
	"history": true,
	"show":    true,
	"why":     true,
	"pr":      true,
	"triage":  true,
	"trace":   true,
}

// todoTreeRoot returns the todo tree's own root for run: the nearest
// ancestor (run included) that defines this PersistentPreRunE. Inside the
// todo tree that is the `moai todo` command itself, so the lane guard can
// tell "the parent invoked bare" from "a subcommand" without closing over
// the not-yet-defined root variable.
func todoTreeRoot(run *cobra.Command) *cobra.Command {
	for p := run; p != nil; p = p.Parent() {
		if p.PersistentPreRunE != nil {
			return p
		}
	}
	return run
}

// todoRefuseLaneMutation guards the todo surface against a lane session
// (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-015): when lane refusal holds, only
// the read-only allowlist — and a bare parent render with no arguments — may
// proceed. The parent-with-args form is refused because it falls through to
// `add`, a mutation. root is the tree the hook was defined on; run is the
// command actually executing.
//
// Card t1554 replaced the dedicated `--auto` lane refusal this guard once
// carried (SPEC-TODO-AUTO-PICK-001 REQ-TAU-008): a lane session's `--auto`
// invocation now runs the lease-based lane cycle (runAutoLaneCycle, dispatched
// from RunE on todoLaneSession), whose only queue write is the nominated
// lease edge — the queue guarantee that refusal carried holds by construction,
// and the REQ-SD-015 mutation refusal below is unchanged for every other
// form. The bare-parent allowance below is what lets the `--auto` invocation
// reach RunE.
//
// @MX:NOTE: [AUTO] SPEC-TODO-CLAIM-LEASE-001 C6 flag-form guard extension:
// `todo claim` is deliberately NOT exempted from this guard in either form.
// A bare claim from a lane session and a `--lane <label>` claim while
// factoryLaneRefusal() holds BOTH refuse with the same text below — the
// flag form grants nothing, because the refusal predicate assumes nothing
// about caller identity (REQ-TCL-013 arms 1/3). REQ-SD-015 bare-refusal
// semantics are unchanged by the claim verb's arrival, and lane self-claim
// governance remains t1338's decision; this guard is where that decision
// would land if it ever widens the allowlist.
func todoRefuseLaneMutation(root, run *cobra.Command, args []string) error {
	if !factoryLaneRefusal() {
		return nil
	}
	if run == root {
		if len(args) == 0 {
			return nil
		}
	} else if todoLaneReadOnlyVerbs[run.Name()] {
		return nil
	}
	return fmt.Errorf("%s", todoLaneMutationRefusalText(todoSurfaceName(run)))
}

// todoLaneSession reports whether this process is a lane session for the
// `--auto` dispatch: the lane role marker equals the role value, or the lane
// label variable is non-empty. It is deliberately narrower than
// factoryLaneRefusal, whose Codex-backend clause would also capture a non-lane
// Codex-backend session — one with no lease alternative (`moai factory next`
// refuses outside a lane) and whose batch approval must stay usable.
func todoLaneSession() bool {
	return factoryLaneAdmission() || os.Getenv(config.EnvMoaiFactoryWorker) != ""
}

// todoLaneMutationRefusalText is the one wording source for the REQ-SD-015
// queue-mutation refusal, shared by the CLI guard and the MCP todo_add tool
// so the two surfaces cannot drift (AC-SD-014 refusal equality).
func todoLaneMutationRefusalText(surface string) string {
	return fmt.Sprintf("moai %s: refused — %s: a lane session cannot mutate the queue (read-only here: bare todo, list, history, show, why, pr, triage); a lane takes its next card through moai factory next",
		surface, factoryLaneBoundarySentinel)
}

// todoVerbShaped matches a first token that reads as a command verb: one
// ASCII word, optionally carrying digits, hyphens, or underscores. Bounded in
// length so a long word in a card text cannot pass for a verb.
//
// t555 (#1654) widened this from `^[a-z][a-z-]{1,15}$`, which admitted only
// the lowercase spelling. A mistyped verb is mistyped in more ways than that:
// the measured leaks were `Show t401` / `SHOW t401` (case), `show2 t401`
// (adjacent-key typo), `show_it t401` (separator), and `s t401` (an
// abbreviation the single-character floor excluded). Each one addressed a
// real card id and each one silently became a card.
//
// Staying ASCII is the deliberate boundary, not an oversight: a non-ASCII
// first token is prose in the operator's own language, never a mistyped
// English verb, so a Korean/Japanese/Chinese card text still falls through
// to add.
var todoVerbShaped = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,23}$`)

// todoCardIDShaped matches the id form the queue issues (`t<decimal>`, see
// factory.BacklogStore). An explicit id in second position is an address at
// any arity.
var todoCardIDShaped = regexp.MustCompile(`^t\d+$`)

// todoBareRefShaped matches the looser reference form the verbs themselves
// accept: normalizeTodoRef maps a bare `<n>` to the id `t<n>` (REQ-TODO-004),
// so `done 151` and `done t151` address the same card.
//
// t555: the guard used to exclude this form outright, on the reasoning that a
// bare number is ordinary card text ("fix 3 flaky tests"). That reasoning
// holds for a number sitting inside a phrase, and only there — `show 401` is
// the very address grammar the verbs publish, and it leaked.
//
// A leading zero is excluded because the queue never issues one
// (`fmt.Sprintf("t%d", …)`), so `0401` addresses no card the store can hold.
var todoBareRefShaped = regexp.MustCompile(`^[1-9][0-9]*$`)

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
	if !todoVerbShaped.MatchString(args[0]) {
		return nil
	}
	kind := todoCardAddressKind(cmd, args)
	if kind == "" {
		return nil
	}
	phrase := strings.Join(args, " ")
	surface := todoSurfaceName(cmd)
	return fmt.Errorf(
		"%s: %q is not a %s verb and %q is %s — refusing to create a card named %q.\n"+
			"Known verbs: %s\nTo add this text as a card anyway: moai %s add %q",
		surface, args[0], surface, args[1], kind, phrase,
		strings.Join(todoVerbNames(cmd), ", "), surface, phrase)
}

// todoSurfaceName reports the command name the operator actually invoked, so
// the refusal above speaks in that surface's own voice.
//
// The verb tree is built once by newTodoCmd and mounted under two names:
// `gtd`, the canonical surface, and `todo`, the thin compatibility spelling
// (REQ-GTD-003). One guard therefore serves both, and a hard-coded "todo"
// answered an operator on the canonical surface with the compatibility name —
// and pointed the recovery line at a command they had not called.
//
// No per-surface branch is needed, because the invoked command already carries
// its own name: this derives the surface exactly as todoVerbNames derives the
// verb list, and for the same reason — a derived name cannot drift from the
// tree, and a third mounting would be named correctly without another edit.
func todoSurfaceName(cmd *cobra.Command) string {
	if cmd == nil || cmd.Name() == "" {
		return "todo"
	}
	return cmd.Name()
}

// todoCardAddressKind reports how args[1] addresses a card, or "" when it does
// not address one at all. The two forms are the two the verbs themselves
// accept (normalizeTodoRef): an explicit `t<n>`, and a bare `<n>`.
//
// The two are NOT symmetric, and the asymmetry is the whole difficulty. A
// `t<n>` is unambiguous wherever it appears, so it needs no further test. A
// bare number is genuinely ambiguous: `drp 401 stale` (a mistyped `drop`) and
// `fix 3 flaky tests` (a card) have the SAME shape — word, number, word — so
// no shape test can separate them. What separates them is the first token:
// `drp` is one edit from a registered verb and `fix` is not.
//
// So the bare form is read as an address in two situations, and only these:
//   - it is the whole remainder (`show 401`), or
//   - the first token is a near-miss of a registered verb (`drp 401 stale`).
//
// The near-miss arm is what the arity test alone could not reach: three verbs
// take TWO positional arguments — `relate <a> <b>`, `drop <n> <reason>`,
// `edit <n> <text>` — so a mistyped call of any of them is three tokens or
// more and slips straight past an arity-only condition.
//
// The returned string is the message fragment naming the form, so the refusal
// tells the operator which grammar it recognized.
func todoCardAddressKind(cmd *cobra.Command, args []string) string {
	switch {
	case todoCardIDShaped.MatchString(args[1]):
		return "a card id"
	case !todoBareRefShaped.MatchString(args[1]):
		return ""
	case len(args) == 2 || todoNearMissVerb(cmd, args[0]):
		return "a card reference"
	default:
		return ""
	}
}

// todoNearMissVerb reports whether tok reads as a mistyping of one of the
// registered verbs: equal ignoring case, within a single edit, or a prefix of
// one from three characters up (`relat` for `relate`).
//
// Derived from the command tree, like the verb list in the refusal message, so
// a verb added later is covered without a second edit here.
func todoNearMissVerb(cmd *cobra.Command, tok string) bool {
	lowered := strings.ToLower(tok)
	for _, verb := range todoVerbNames(cmd) {
		if withinOneEdit(lowered, verb) {
			return true
		}
		if len(lowered) >= 3 && strings.HasPrefix(verb, lowered) {
			return true
		}
	}
	return false
}

// withinOneEdit reports whether a and b are at most one insertion, deletion,
// or substitution apart.
//
// Bounded at one, so it is a short scan rather than the usual edit-distance
// matrix: at that bound the strings differ in length by at most one, which
// leaves exactly two cases — equal lengths differ in at most one position, and
// unequal lengths mean the longer is the shorter with one character inserted.
// (internal/harness carries its own unexported copy for prefix-conflict
// detection; exporting across that package boundary for this one call would be
// a wider change than the twenty lines it saves.)
func withinOneEdit(a, b string) bool {
	if len(b) < len(a) {
		a, b = b, a
	}
	if len(b)-len(a) > 1 {
		return false
	}
	if len(a) == len(b) {
		diffs := 0
		for i := range a {
			if a[i] != b[i] {
				diffs++
				if diffs > 1 {
					return false
				}
			}
		}
		return true
	}
	// b is one longer: b must be a with one character inserted.
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return a[i:] == b[i+1:]
		}
	}
	return true
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

// todoAddScan is the result of the add command's own argument scan: the
// known flags the surface registers, plus the one card text.
type todoAddScan struct {
	pick          bool
	force         bool
	dryRun        bool
	classFile     string
	haveClassFile bool
	help          bool
	text          string
	// SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-013 (add clause): the issuance
	// flags. The have* markers turn a repeated flag into a named refusal —
	// a silent overwrite would bury the first judgment.
	origin        string
	haveOrigin    bool
	parent        string
	haveParent    bool
	sizeLines     string
	haveSizeLines bool
	files         string
	haveFiles     bool
}

// scanTodoAddArgs separates the known flags from the card text in the raw
// argument vector the add command receives under DisableFlagParsing
// (SPEC-TODO-SURFACE-POLISH-001 REQ-TSP-010).
//
// Why the add path parses its own flags: pflag's interspersed parser
// consumes every `-`-prefixed token as flags, and this surface registers no
// shorthands — so a quoted body starting with `-f` (`moai todo add "-f fix
// the flaky tests"`) died as `unknown shorthand flag: 'f'`. A body is TEXT;
// the scanner treats a single-dash token as text and reserves flag meaning
// for the three known long forms plus the `--` separator, so the known-flag
// semantics (REQ-TSP-011) survive the change verbatim.
//
// The scan runs in Args — before PersistentPreRunE, where pflag's errors
// used to fire — so an errored add still skips the run-phase hooks exactly
// as it did under the old parser.
func scanTodoAddArgs(raw []string) (*todoAddScan, error) {
	scan := &todoAddScan{}
	var positionals []string
	separator := false // after `--`, every remaining token is the text
	for i := 0; i < len(raw); i++ {
		tok := raw[i]
		switch {
		case separator:
			positionals = append(positionals, tok)
		case tok == "--":
			separator = true
		case strings.HasPrefix(tok, "--"):
			name, value, hasValue := strings.Cut(tok, "=")
			switch name {
			case "--pick":
				scan.pick = true
			case "--dry-run":
				scan.dryRun = true
			case "--force":
				scan.force = true
			case "--origin":
				if scan.haveOrigin {
					return nil, fmt.Errorf("flag repeated: %s", name)
				}
				scan.haveOrigin = true
				if hasValue {
					scan.origin = value
				} else {
					i++
					if i >= len(raw) {
						return nil, fmt.Errorf("flag needs an argument: %s", name)
					}
					scan.origin = raw[i]
				}
			case "--parent":
				if scan.haveParent {
					return nil, fmt.Errorf("flag repeated: %s", name)
				}
				scan.haveParent = true
				if hasValue {
					scan.parent = value
				} else {
					i++
					if i >= len(raw) {
						return nil, fmt.Errorf("flag needs an argument: %s", name)
					}
					scan.parent = raw[i]
				}
			case "--size-lines":
				if scan.haveSizeLines {
					return nil, fmt.Errorf("flag repeated: %s", name)
				}
				scan.haveSizeLines = true
				if hasValue {
					scan.sizeLines = value
				} else {
					i++
					if i >= len(raw) {
						return nil, fmt.Errorf("flag needs an argument: %s", name)
					}
					scan.sizeLines = raw[i]
				}
			case "--files":
				if scan.haveFiles {
					return nil, fmt.Errorf("flag repeated: %s", name)
				}
				scan.haveFiles = true
				if hasValue {
					scan.files = value
				} else {
					i++
					if i >= len(raw) {
						return nil, fmt.Errorf("flag needs an argument: %s", name)
					}
					scan.files = raw[i]
				}
			case "--classification-file":
				scan.haveClassFile = true
				if hasValue {
					scan.classFile = value
				} else {
					i++
					if i >= len(raw) {
						return nil, fmt.Errorf("flag needs an argument: %s", name)
					}
					scan.classFile = raw[i]
				}
			case "--help":
				scan.help = true
			default:
				return nil, fmt.Errorf("unknown flag: %s", name)
			}
		case tok == "-h":
			scan.help = true
		default:
			// Includes every single-dash token (`-f`, `-x`, prose starting
			// with one): the surface registers no shorthands, so a leading
			// dash reaches the store verbatim instead of dying as an
			// unknown shorthand flag. A body that genuinely reads like a
			// known long flag keeps the `--` escape hatch.
			positionals = append(positionals, tok)
		}
	}
	if scan.help {
		return scan, nil
	}
	if len(positionals) != 1 {
		return nil, fmt.Errorf("accepts 1 arg(s), received %d", len(positionals))
	}
	scan.text = positionals[0]
	return scan, nil
}

// newTodoAddCmd — `moai todo add "<text>"` (REQ-TODO-002): append under the
// lock, print the issued id and its 1-based queue position. `--pick` (t71)
// folds the pick into the same locked write. `--classification-file`
// (SPEC-TODO-CLASSIFY-DISPATCH-001 REQ-TCD-004) supplies a validated
// classification judgement — `<path>` or `-` for standard input; it is the
// ONLY classification injection seam, and the product computes no judgment
// of its own beyond the deterministic default (plan D.3).
//
// The command parses its own arguments (DisableFlagParsing +
// scanTodoAddArgs, REQ-TSP-010); the flag declarations below exist for the
// usage text and are not consulted by a parser anymore.
func newTodoAddCmd() *cobra.Command {
	var scan *todoAddScan
	cmd := &cobra.Command{
		Use:                "add <text>",
		Short:              "Append a card to the backlog queue",
		DisableFlagParsing: true,
		Args: func(cmd *cobra.Command, args []string) error {
			parsed, err := scanTodoAddArgs(args)
			if err != nil {
				return err
			}
			scan = parsed
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if scan.help {
				return cmd.Help()
			}
			text := scan.text
			if strings.TrimSpace(text) == "" {
				return fmt.Errorf("todo add: text must be non-empty")
			}
			// SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-004: --dry-run prints the
			// same presentation and writes nothing — the queue file stays
			// byte-identical, no id is consumed, and an exact duplicate is
			// reported as a would-be refusal instead of refusing.
			if scan.dryRun {
				return runTodoAddDryRun(cmd, text)
			}
			// REQ-TCD-004: the supplied classification is validated BEFORE
			// the locked write — an out-of-set value or the jev identity is
			// a usage refusal with nothing written.
			// SPEC-TCD-LLM-DECIDER-001 REQ-TLD-002: --classification-file
			// outranks the standing selection; without a supplied file the
			// standing decider resolves from MOAI_TODO_DECIDER.
			var dec factory.CardDecider
			if scan.haveClassFile {
				resolved, err := todoDeciderFromClassificationFile(scan.classFile)
				if err != nil {
					return err
				}
				dec = resolved
			} else {
				selected, err := todoDeciderFromEnv()
				if err != nil {
					return err
				}
				dec = selected
			}
			// SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-013: the issuance flags
			// resolve to the attribute record the locked write validates and
			// attaches; a flag-less surface passes nil. The resolution sits
			// BEFORE the pick branch (card t1454 card-review r2 finding 7):
			// `--pick` carries the issuance flags too — an early return here
			// left them unvalidated and unsaved.
			var iss *factory.BacklogIssuance
			if scan.haveOrigin || scan.haveParent || scan.haveSizeLines || scan.haveFiles {
				iss = &factory.BacklogIssuance{Origin: scan.origin, SpawnedBy: scan.parent}
				if scan.haveSizeLines {
					n, parseErr := strconv.Atoi(strings.TrimSpace(scan.sizeLines))
					if parseErr != nil {
						return fmt.Errorf("todo add: --size-lines must be an integer (got %q)", scan.sizeLines)
					}
					iss.SizeLines = &n
				}
				if scan.haveFiles {
					for _, f := range strings.Split(scan.files, ",") {
						if f = strings.TrimSpace(f); f != "" {
							iss.Files = append(iss.Files, f)
						}
					}
				}
			}
			if scan.pick {
				return runTodoAddPick(cmd, newTodoStore(), text, scan.force, dec, iss)
			}
			return runTodoAddAppendIss(cmd, text, scan.force, dec, iss)
		},
	}
	cmd.Flags().BoolVar(new(bool), "pick", false,
		"Append AND mark picked as one locked write, printing the issued id")
	cmd.Flags().BoolVar(new(bool), "dry-run", false,
		"Print the issuance presentation and write nothing")
	cmd.Flags().StringVar(new(string), "origin", "",
		"Issuance origin (closed set: "+strings.Join(factory.IssuanceOrigins, ", ")+")")
	cmd.Flags().StringVar(new(string), "parent", "",
		"Card id this card was spawned from (live, dropped or archived)")
	cmd.Flags().StringVar(new(string), "size-lines", "",
		"Estimated product line count")
	cmd.Flags().StringVar(new(string), "files", "",
		"Comma-separated expected files")
	cmd.Flags().BoolVar(new(bool), "force", false,
		"Admit a card the analyser reads as an exact duplicate, recording that it was forced")
	cmd.Flags().StringVar(new(string), "classification-file", "",
		"Classification judgement JSON (<path> or - for stdin); validated against the closed value sets before the write")
	return cmd
}

// runTodoAddAppend is the plain-add body shared by `todo add <text>` and the
// parent's natural-language fallthrough (t69): non-empty guard, locked
// append, "<id> <position>" stdout line. `--pick` stays add-only — the
// fallthrough path has no flags. The presentation renders to stderr here;
// the MCP surface takes the returned text instead.
func runTodoAddAppend(cmd *cobra.Command, text string, force bool, dec factory.CardDecider) error {
	return runTodoAddAppendIss(cmd, text, force, dec, nil)
}

// runTodoAddAppendIss is the issuance-carrying form: the add surface with
// flags passes the resolved attributes; the fallthrough and MCP surfaces
// run the nil form.
func runTodoAddAppendIss(cmd *cobra.Command, text string, force bool, dec factory.CardDecider, iss *factory.BacklogIssuance) error {
	presentation, err := runTodoAddAppendRoot(resolveTodoQueueRoot(), cmd, text, force, dec, iss)
	if err == nil && presentation != "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), presentation)
	}
	return err
}

// runTodoAddAppendRoot is runTodoAddAppend anchored at an explicit root —
// the shape the MCP todo_add tool calls (REQ-SD-024), so both surfaces run
// one implementation. The decider argument is the classification seam this
// invocation resolves; the MCP surface passes the package default. It
// returns the rendered issuance presentation (SPEC-TODO-CARD-ISSUANCE-001
// REQ-TCI-002/005): the CLI prints it to stderr, the MCP tool appends it
// after the result's first line; "" means nothing fired.
func runTodoAddAppendRoot(root string, cmd *cobra.Command, text string, force bool, dec factory.CardDecider, iss *factory.BacklogIssuance) (string, error) {
	if dec == nil {
		// SPEC-TCD-LLM-DECIDER-001 REQ-TLD-002: the MCP todo_add surface
		// resolves the same standing decider the CLI path resolves, so the
		// MOAI_TODO_DECIDER selection applies to both surfaces through this
		// one nil branch.
		selected, err := todoDeciderFromEnv()
		if err != nil {
			return "", err
		}
		dec = selected
	}
	// SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-003: the presentation is computed
	// BEFORE the queue lock is acquired (queue snapshot, completed-SPEC
	// directory read, lane probes — all outside the lock) and rendered after
	// the admission is decided; a refusal still carries it.
	presentation := todoIssuancePresentation(root, text)
	// SPEC-TCD-LLM-DECIDER-001 REQ-TLD-005: the LLM judgment is computed
	// BEFORE the queue lock is acquired and attached inside the same locked
	// write as a static carrier — only the computation moved out of the
	// lock, never the attachment.
	dec = todoPreClassifyLLM(dec, text, cmd.ErrOrStderr())
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("todo add: text must be non-empty")
	}
	// Card t1313 (GitHub #1732): the WRITE verb discloses the store
	// DIVERGENCE the read verbs disclose (SPEC-TODO-STALE-STORE-001
	// REQ-TSS-001 family) — the response's issued id is a receipt for the
	// store that ANSWERED, and a divergent project-local store beside it is
	// exactly the silent split the incident reported (ids issued by one
	// store, queue read from another). Scoped to the stale-store line only:
	// the t395 backlog.json note stays read-surface by its operator decision
	// (TestTodoWriteVerbs_CarryNoDisclosure). Runs before the Mutate, on
	// stderr; stdout stays the bare "id position" machine line.
	if err := discloseStaleLocalStores(cmd.ErrOrStderr(), "add",
		factory.InspectStaleLocalStores(todoQueueRootForDisclosure())); err != nil {
		return "", err
	}
	var item factory.BacklogItem
	var pos int
	err := todoStoreAt(root).Mutate(func(rec *factory.BacklogRecord) error {
		// SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-013: the issuance validations
		// run BEFORE the append — nothing is written and no id is consumed
		// on a refusal. Parent existence reads the record the same locked
		// write will append into: live, dropped AND archived all count
		// (a follow-up of a closed card is the normal flow).
		if err := validateIssuanceInLock(rec, iss); err != nil {
			return err
		}
		var mutErr error
		item, pos, mutErr = appendAnalyzedCard(rec, text, factory.BacklogStateQueued, force)
		if mutErr != nil {
			return mutErr
		}
		attachIssuance(rec, item.ID, iss)
		// REQ-TCD-001: the classification is resolved INSIDE the same locked
		// write — no card becomes visible to a machine selector unclassified,
		// and no two-step window exists (plan G1).
		todoApplyClassification(rec, item.ID, todoClassifyInLock(dec, text, cmd.ErrOrStderr()))
		// REQ-TCD-005/-006: the sort is re-established inside the same locked
		// write, and the printed position is the sorted 1-based position.
		rec.SortByClassification()
		pos = rec.QueuedPosition(item.ID)
		return nil
	})
	if err != nil {
		if presText := renderIssuanceText(presentation); presText != "" {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), presText)
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return "", err
	}
	presText := renderIssuanceText(presentation)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %d\n", item.ID, pos)
	return presText, nil
}

// validateIssuanceInLock runs the issuance attribute validations inside the
// caller's locked write (SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-013): an
// out-of-set origin, or a parent that names no card, refuses with nothing
// written. Parent existence reads the record the same locked write appends
// into: live, dropped AND archived all count — a follow-up of a closed card
// is the normal flow. Both add surfaces (append and --pick) share it.
func validateIssuanceInLock(rec *factory.BacklogRecord, iss *factory.BacklogIssuance) error {
	if iss == nil {
		return nil
	}
	if iss.Origin != "" && !factory.IssuanceOriginValid(iss.Origin) {
		return fmt.Errorf("todo add: --origin must be one of %s (got %q)",
			strings.Join(factory.IssuanceOrigins, ", "), iss.Origin)
	}
	if iss.SpawnedBy != "" {
		parentExists := false
		for i := range rec.Items {
			if rec.Items[i].ID == iss.SpawnedBy {
				parentExists = true
				break
			}
		}
		for i := range rec.Archived {
			if rec.Archived[i].Item.ID == iss.SpawnedBy {
				parentExists = true
				break
			}
		}
		if !parentExists {
			return fmt.Errorf("todo add: --parent names no card: %s", iss.SpawnedBy)
		}
	}
	return nil
}

// attachIssuance records the issuance attributes on the freshly appended
// item; a nil record attaches nothing.
func attachIssuance(rec *factory.BacklogRecord, cardID string, iss *factory.BacklogIssuance) {
	if iss == nil {
		return
	}
	for i := range rec.Items {
		if rec.Items[i].ID == cardID {
			rec.Items[i].Issuance = iss
			return
		}
	}
}

// runTodoAddDryRun is the `--dry-run` body (REQ-TCI-004): the same
// presentation with the floor lifted (design §3.2 — top-3 regardless of
// score), nothing written, and an exact duplicate reported as a would-be
// refusal instead of refusing.
func runTodoAddDryRun(cmd *cobra.Command, text string) error {
	root := resolveTodoQueueRoot()
	rec, err := todoStoreAt(root).LoadPure()
	if err != nil {
		return err
	}
	presentation := todoIssuancePresentationFloor(root, text, 0)
	if presText := renderIssuanceText(presentation); presText != "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), presText)
	}
	if match := factory.ClassifyCardText(text, rec.Items); match.Kind == factory.BacklogMatchExact {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "a real add would refuse: %s already holds this card\n", match.ID)
	}
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "dry-run: nothing was written")
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
// `--pick` just picked. The issuance attributes ride the same locked write
// and the same validations as the append path (card t1454 card-review r2
// finding 7).
func runTodoAddPick(cmd *cobra.Command, store *factory.BacklogStore, text string, force bool, dec factory.CardDecider, iss *factory.BacklogIssuance) error {
	if dec == nil {
		dec = todoCardDecider
	}
	// SPEC-TCD-LLM-DECIDER-001 REQ-TLD-005: the same outside-the-lock
	// computation the append path carries — the pick's locked write is a
	// queue lock too, and an in-lock LLM call would stall it identically.
	dec = todoPreClassifyLLM(dec, text, cmd.ErrOrStderr())
	// Card t1313: the same stale-store disclosure the append path carries —
	// the issued id is a receipt for the store that answered. Scoped to the
	// t1307 divergence line only (see the append-path comment).
	if err := discloseStaleLocalStores(cmd.ErrOrStderr(), "add --pick",
		factory.InspectStaleLocalStores(todoQueueRootForDisclosure())); err != nil {
		return err
	}
	var item factory.BacklogItem
	err := store.Mutate(func(rec *factory.BacklogRecord) error {
		// REQ-TCI-013: the issuance validations run BEFORE the append —
		// nothing is written and no id is consumed on a refusal.
		if err := validateIssuanceInLock(rec, iss); err != nil {
			return err
		}
		var mutErr error
		item, _, mutErr = appendAnalyzedCard(rec, text, factory.BacklogStatePicked, force)
		if mutErr != nil {
			return mutErr
		}
		attachIssuance(rec, item.ID, iss)
		// REQ-TCD-001: same locked write, same seam, same sort duty.
		todoApplyClassification(rec, item.ID, todoClassifyInLock(dec, text, cmd.ErrOrStderr()))
		rec.SortByClassification()
		return nil
	})
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "picked %s %s\n", item.ID, todoTextPrefix(item.Text))
	return nil
}

// todoListDefaultLimit is the list render's default bound. A bounded read
// stays the default, --limit raises or lowers it, --limit 0 lifts it
// entirely, and a truncated listing states the withheld count on stderr
// because a truncated read must never be mistaken for a complete one.
//
// The value is 100 (SPEC-TODO-SURFACE-POLISH-001 REQ-TSP-020, operator
// decision at the 2026-09-30 kickoff): the measured live queue (55 rows,
// 2026-09-29) was cut in half every single day by the old 20, while the
// withheld line made the cut VISIBLE — the bound stopped matching the
// queue's actual scale. 100 renders today's queue whole; when a queue
// outgrows it again the same withheld line reports the fact.
const todoListDefaultLimit = 100

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
	return runTodoListRoot(resolveTodoQueueRoot(), cmd, jsonOutput, droppedOnly, limit)
}

// runTodoListRoot is runTodoList anchored at an explicit root — the shape
// the MCP todo_list tool calls (REQ-SD-024), so both surfaces render one
// implementation.
func runTodoListRoot(root string, cmd *cobra.Command, jsonOutput bool, droppedOnly bool, limit int) error {
	if !jsonOutput && limit < 0 {
		return fmt.Errorf("todo list: --limit must be >= 0 (got %d)", limit)
	}
	store := todoReadStoreAt(root)
	// REQ-BJD-002 — probed before the read. stderr only: stdout is what the
	// foreman reads.
	_ = discloseQueueLayout(cmd, "todo")
	rec, err := store.LoadPure()
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}
	out := cmd.OutOrStdout()
	if jsonOutput {
		// REQ-TCL-014: the JSON face renders through the lease-free
		// projection — the new columns are excluded from this
		// serialization entirely, keeping the frozen golden gate
		// byte-identical and the machine contract lease-blind.
		data, err := json.Marshal(todoJSONProjection(rec))
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
	var visible []factory.BacklogItem
	dropped := 0
	for _, it := range rec.Items {
		isDropped := it.State == factory.BacklogStateDropped
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
		// REQ-TCL-010: a card carrying lease fields exposes them on the
		// human surface (by=/lease= cells before the text, which stays the
		// last field); a card without them keeps the historical line shape.
		_, _ = fmt.Fprintf(out, "%s\t%s\t%s%s\n", it.ID, it.State, todoLeaseCells(it), todoPRCell(it.Text))
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
			var specID string
			// Resolved once, up front, so the query, the verdict line, and the
			// disclosure all name the same ref (todoLandedRef's contract).
			ref, refLevel := todoLandedRefResolved()
			// Unknown until a query answers otherwise. Absent the flag no
			// query runs at all, and `unknown` is the honest report of that
			// — UNLESS the card already carries recorded landing evidence,
			// which is an answer that was obtained earlier and stored (card
			// t665). Reporting `unknown` over a validated record was the
			// reported defect: eight cards archived on 2026-09-12 carried a
			// recorded delivering SHA and every one of them closed as if
			// nothing were known.
			verdict := factory.LandingUnknown
			var landing *factory.LandingEvidence
			if err := store.Mutate(func(rec *factory.BacklogRecord) error {
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
				if rec.Items[at].SpecID != nil {
					specID = *rec.Items[at].SpecID
				}
				// Read BEFORE ArchiveCard moves the row: the archive copies
				// the item, so the record survives either way, but reading it
				// here keeps the verdict and the line derived from the same
				// row the mutation addressed.
				landing = rec.Items[at].Landing
				if expect != "" && !strings.HasPrefix(rec.Items[at].Text, expect) {
					return fmt.Errorf("backlog item %s is %q, not matching --expect %q",
						id, todoTextPrefix(rec.Items[at].Text), expect)
				}
				if requireLanded {
					answer, err := todoRequireLanded(cmd, id, ref, refLevel)
					if err != nil {
						return err
					}
					verdict = answer
				}
				if err := rec.ArchiveCard(id); err != nil {
					return err
				}
				if requireLanded {
					// REQ-TST-008: the answering path persists what the query
					// said — verdict, answering ref, verdict time — onto the
					// entry ArchiveCard just appended, alongside (never
					// instead of) any operator-recorded evidence the row
					// already carried (REQ-TST-009). Without the flag nothing
					// is persisted here: no query ran, so no invented answer
					// and no fabricated record. The record carries no SHA —
					// a query-derived SHA is outside the evidence store's
					// write authority, and the delivering SHA is re-derived
					// at re-adjudication by re-running the predicate against
					// the recorded ref (REQ-TST-013).
					rec.Archived[len(rec.Archived)-1].LandingVerdict = &factory.LandingVerdict{
						Verdict: verdict,
						Ref:     ref,
						At:      time.Now().UTC().Format(time.RFC3339),
					}
				}
				return nil
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			// One line per act, carrying exactly one landing verdict — a second
			// line would give an operator script two records for one event.
			// The suffix preserves the `done <id>` prefix every existing
			// reader keys off. The answering ref is appended (REQ-TLA-010)
			// only when a landing query actually ran: without the flag no ref
			// answered, and naming one would dress "the guard did not run" up
			// as "the guard answered against ref X".
			//
			// The recorded-evidence suffix (card t665) is appended on BOTH
			// paths and carries its provenance, so a stored operator
			// assertion is never mistaken for an answer this run's query
			// produced. Without the flag it also supplies the verdict — a
			// validated record IS the answer, obtained earlier.
			recorded := todoDoneLandingSuffix(landing)
			if recorded != "" && !requireLanded {
				verdict = factory.LandingLanded
			}
			line := fmt.Sprintf("done %s landing=%s", id, verdict)
			if requireLanded {
				line += " ref=" + ref
			}
			if recorded != "" {
				line += " " + recorded
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), line)
			recordFactoryCardState(id, specID, "completed", "card.completed")
			// The queue write is committed and the lock released: the
			// card-close memory fold runs outside the store's mutation
			// window (AC-MFB-008 (iv)), gated, bounded and fail-open
			// (REQ-MFB-007).
			foldClosedCardMemoryFn(id)
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
` + "`done <id> landing=landed|not-landed|unknown`" + `, with ` + "`ref=<answering ref>`" + `
appended when ` + "`--require-landed`" + ` ran — without the flag no query ran, so
no ref answered and none is named — and the ` + "`done <id> `" + ` prefix every
existing reader keys off is preserved. Without the flag the verdict is
` + "`unknown`" + `, because no query ran: "the guard passed" and "the guard did not
run" are different facts and no longer the same bytes. When the answering
ref came from BELOW the configured ` + "`git_strategy.worktree_base_branch`" + ` — the
repository's own recorded default or the compiled-in fallback — the chain
level that supplied it is disclosed on stderr; a configured project gets
no such notice.`
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
func todoRequireLanded(cmd *cobra.Command, id, ref string, refLevel factory.LandedRefLevel) (factory.LandingAnswer, error) {
	// The answering level is disclosed when it sits BELOW the configured key:
	// a ref the repository supplied through refs/remotes/origin/HEAD or the
	// compiled-in default is the exceptional path, and the operator sees the
	// source rather than inferring it (REQ-TLA-011). A configured project
	// (level 1) gets no notice — the exceptional path is what the notice
	// marks, not every path.
	if refLevel != factory.LandedRefConfigured {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"note: landed ref %s was supplied by chain level %d (%s) — this project does not configure git_strategy.worktree_base_branch\n",
			ref, refLevel, todoRefLevelSource(refLevel))
	}
	q := factory.GitLandedQuerier{Run: todoRunCommand, Ref: ref}
	answer, err := q.Landed(id)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"note: --require-landed could not answer for %s against %s (%v) — proceeding, because an unanswerable query is not evidence of not-landed\n",
			id, q.LandedRef(), err)
		return factory.LandingUnknown, nil
	}
	if answer == factory.LandingNotLanded {
		return answer, fmt.Errorf("backlog item %s is named by no commit on %s — --require-landed refuses "+
			"(the check asks whether anything naming the card has landed on that ref, not whether the card's last step has)",
			id, q.LandedRef())
	}
	return answer, nil
}

// todoDoneLandingSuffix renders the recorded-evidence suffix of the `done`
// line, or "" when the card carries nothing worth reporting (card t665).
//
// Only a record that carries a delivering SHA and passes its own validation
// produces a suffix, because only that record answers the question the
// archive lost: which commit delivered this card. A record holding just the
// observed ref position asserts no delivering commit, so reporting it here
// would dress a ref position up as a delivery — the confusion REQ-TLE-013
// exists to prevent.
//
// The provenance travels with the value for the same reason `landed` refuses
// to store one without the other: `landed` as a verdict and `operator` as its
// source are two different facts, and a reader who sees only the first cannot
// tell a stored assertion from a query this run made.
func todoDoneLandingSuffix(e *factory.LandingEvidence) string {
	if e == nil || strings.TrimSpace(e.SHA) == "" {
		return ""
	}
	if err := e.Validate(); err != nil {
		return ""
	}
	return fmt.Sprintf("sha=%s source=%s", e.SHA, e.SHASource)
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
read-only candidates. A queued card is promoted to picked either by the
operator's pick made through the leader session or by a lane's own
self-dispatch — a lane may pick a queued card itself to start on it.
` + "`moai todo next <n> [--spec <SPEC-ID>]`" + `
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
					// POSITIVE enumeration (SPEC-TODO-HOLD-STATE-001
					// REQ-THS-011): the candidate list selects by naming the
					// state it accepts, never by refusing the ones it knows —
					// a state added later must not fall through a negative's
					// default.
					if it.State == factory.BacklogStateQueued {
						queued++
						_, _ = fmt.Fprintf(out, "%s\t%s\n", it.ID, todoPRCell(it.Text))
					}
				}
				if queued == 0 {
					_, _ = fmt.Fprintln(out, "queue is empty")
				}
				return nil
			}

			id := normalizeTodoRef(args[0])
			var pickedText string
			if err := store.Mutate(func(rec *factory.BacklogRecord) error {
				for i := range rec.Items {
					if rec.Items[i].ID == id {
						// The pick gate enumerates POSITIVELY
						// (SPEC-TODO-HOLD-STATE-001 REQ-THS-011/012): only a
						// queued card is pickable, and every other state —
						// including any state added after this code was
						// written — is refused by the switch's default rather
						// than admitted by a negative's fall-through. This
						// gate was the SPEC's one behavioral red-now: it used
						// to refuse only `dropped`, so a held card (and any
						// future state) was pickable.
						switch rec.Items[i].State {
						case factory.BacklogStateQueued:
							// the only pickable state
						case factory.BacklogStateDropped:
							return fmt.Errorf("backlog item %s is dropped — use moai todo undrop %s before picking", id, id)
						case factory.BacklogStateHold:
							return fmt.Errorf("backlog item %s is held — use moai todo unhold %s before picking", id, id)
						case factory.BacklogStatePicked:
							return fmt.Errorf("backlog item %s is already picked — use moai todo unpick %s before picking it again", id, id)
						default:
							return fmt.Errorf("backlog item %s is %q — not a pickable state", id, rec.Items[i].State)
						}
						if expect != "" && !strings.HasPrefix(rec.Items[i].Text, expect) {
							// Refused mutation: Mutate writes nothing, so the
							// file stays byte-identical on a mismatch.
							return fmt.Errorf("backlog item %s is %q, not matching --expect %q",
								id, todoTextPrefix(rec.Items[i].Text), expect)
						}
						rec.Items[i].State = factory.BacklogStatePicked
						// REQ-TST-004: the current picked episode begins now;
						// any stamp from a previous episode is overwritten.
						rec.Items[i].PickedAt = todoStampNow()
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
			recordFactoryCardState(id, specID, "picked", "card.assigned")
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
			if err := store.Mutate(func(rec *factory.BacklogRecord) error {
				for i := range rec.Items {
					if rec.Items[i].ID != id {
						continue
					}
					// POSITIVE enumeration (SPEC-TODO-HOLD-STATE-001
					// REQ-THS-012): the gate names the state it reverts, and
					// every other state refuses — no negated comparison whose
					// default could swallow a state added later.
					switch rec.Items[i].State {
					case factory.BacklogStatePicked:
						// the only unpickable-into-queued state
					default:
						// Refused mutation: Mutate writes nothing, so the
						// file stays byte-identical on a refusal.
						return fmt.Errorf("backlog item %s is %s, not picked", id, rec.Items[i].State)
					}
					rec.Items[i].State = factory.BacklogStateQueued
					// REQ-TST-005: a queued card carries no picked stamp.
					rec.Items[i].PickedAt = nil
					rec.Items[i].SpecID = nil
					text = rec.Items[i].Text
					return nil
				}
				return fmt.Errorf("no backlog item %s", id)
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			recordFactoryCardState(id, "", "queued", "card.unpicked")
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "unpicked %s %s\n", id, todoTextPrefix(text))
			return nil
		},
	}
}

func recordFactoryCardState(cardID, specID, state, eventKind string) {
	runID := os.Getenv(config.EnvFactoryRunID)
	if runID == "" || os.Getenv(config.EnvMoaiFactoryWorkers) == "" {
		return
	}
	owner := os.Getenv(config.EnvMoaiFactoryWorker)
	if owner == "" {
		owner = factory.RoleLeader // the factory card owner vocabulary: `leader` (REQ-RNC-010)
	}
	// Queue mutations resolve through the primary checkout, but provenance must
	// describe the lane checkout that actually selected and executed the card.
	// OpenFactory canonicalizes only the DB routing after capture.
	_ = factory.RecordFactoryCardState(resolveProjectDir(), runID, cardID, owner, specID, state, eventKind)
}

// normalizeTodoRef maps a bare <n> argument to the item id t<n>; an explicit
// id (t<n>) passes through unchanged (REQ-TODO-004's normalization rule).
func normalizeTodoRef(arg string) string {
	if !strings.HasPrefix(arg, "t") {
		return "t" + arg
	}
	return arg
}

// todoStampNow returns a pointer to the current instant in the store's
// added_at TEXT format (RFC 3339 UTC) — the value every transition stamp
// carries (SPEC-TODO-TRANSITION-STAMPS-001 REQ-TST-004..007). It is set
// inside the Mutate callback at the moment the transition happens, so the
// stamp and the state change land in one locked write.
func todoStampNow() *string {
	v := time.Now().UTC().Format(time.RFC3339)
	return &v
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
	runes := []rune(todoPRCell(text))
	if len(runes) <= todoTextPrefixMax {
		return string(runes)
	}
	return string(runes[:todoTextPrefixMax]) + "..."
}

func init() {
	rootCmd.AddCommand(newTodoCmd())
}
