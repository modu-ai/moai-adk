# SPEC-WEB-CONSOLE-015 — Design

Deliberately thin, and deliberately narrow. This document does **not** restate the schema shape
or the factory-lane join topology — spec.md §A and §C already carry those, and the plan-auditor
accepted that rationale. It exists for the content that had no home in iteration 1, which is
exactly the content whose absence let two acceptance criteria silently hard-code one branch of a
decision that was still open.

Three sections: the resolved cross-cutting decisions (§1), the M3 migration sequence (§2), and
the M4 resolution/adoption split (§3).

## §1 Resolved cross-cutting decisions

Four of the six decisions have consequences that outlive the decision itself. G-2 and G-5 are
recorded in plan.md §G and need no design treatment — G-2 is a corollary of G-3, and G-5 is a
product choice with a one-line consequence already captured in AC-WC15-030.

### 1.1 G-1 — the card identifier is derived, not produced

**Decision.** The launcher derives the card id from the basename of
`git rev-parse --show-toplevel`, preferring an explicit environment-variable override when one
is set, and writes an empty field when neither yields a value.

**Why this shape.** The card id already exists on disk — the dispatch protocol requires the
worktree directory to keep it (`wt: .claude/worktrees/<card-id>`), so every card-carrying lane is
standing inside a directory named for its card. Deriving costs no new file, no new registry, and
no new lifecycle. The override exists because a derived value is only as good as the convention
it reads, and an operator who needs to override should not have to rename a directory.

**Consequences.**

- A lane not inside a card worktree yields nothing. This is correct rather than lossy: such a
  lane is already outside the dispatch protocol, and REQ-WC15-042's empty-field rule routes it
  through the same "not recorded" honesty path the model and effort cells use (REQ-WC15-012).
  The console never guesses a card id.
- The derivation runs at launch, in `internal/cli`, alongside the existing model/effort
  resolution — one `git rev-parse` per session launch, not per render.
- The override is read-first, so an explicit value always wins over a directory name. A future
  producer (a lead writing the value at dispatch) can populate the override with no schema change.

**Rejected.** *(a) A launch-time env var as the sole source.* Factory routes a card whole to a
free lane *after* the lane launched, so process environment cannot carry a per-card value without
relaunching a lane per card — it would be both a new producer surface and a new way to be absent.
It survives only as the override. *(b) The lead writes it into the record after launch.* Accurate
while the lead is disciplined, silently stale the first time it is not, and a stale card label on
an audit board is worse than an absent one.

### 1.2 G-3 — hard cut, and what that forecloses

**Decision.** `.moai/state/context-usage.json` → `.moai/state/context-usage/<session-id>.json`
with no dual-write window.

**Why this shape.** The file is render-ephemeral telemetry, regenerated on every statusline
render, so there is no durable data to migrate — a cut loses at most one render's snapshot. Every
reader is in-tree and enumerated (spec.md §C.3). And a dual-write window would keep the
last-writer-wins single slot alive, which is the exact trap spec.md §A.3 documents with a live
observation: session `368a2bd9…` at 260,000 tokens overwritten by session `e463a3c9…` at 0. A
compatibility window whose function is to preserve a known defect is not a compatibility window.

**Consequences.**

- Two acceptance criteria assert the cut directly — AC-WC15-020 (`context-usage.json` absent
  after a render) and AC-WC15-024 (the bare path absent from all four doctrine files). Both are
  now correct **by decision**; iteration 1 had them correct by accident, which is the defect D3
  named.
- The single-slot validation becomes dead rather than merely redundant: `isFreshForSession`, the
  `writer_pid` discriminator, and the same-payload check exist because one file served N
  sessions. M3 deletes what the split makes unreachable rather than leaving it as decoration —
  and REQ-WC15-024 requires the doctrine to drop the read procedure that describes it, so leaving
  the code would put doctrine and implementation into contradiction.
- There is no migration step for existing files. A stale `.moai/state/context-usage.json` on an
  operator's disk is inert after the cut; it is gitignored and regenerated state, so no cleanup
  is specified.

### 1.3 G-4 — the `/todo` route, and the cost it accepted

**Decision.** The todo section gets its own top-level `/todo` route and a sixth nav entry, not a
panel on `/kanban`. This goes **against** the plan-auditor's recommendation.

**Why this shape.** The backlog queue is an operator surface in its own right — the thing the
lead reads to pick the next card — and a route makes it addressable and shareable as a URL. A
panel makes it a subordinate of the board, which inverts the actual relationship: the queue feeds
the board.

**The cost is accepted, not avoided.** The auditor's objection was concrete: a separate route
"requires an event-vocabulary change that §D Out of Scope explicitly forbids". Choosing the route
anyway means either the vocabulary changes and the exclusion moves, or the objection does not
hold. It was measured rather than argued, and it does not hold:

1. `watchMap["kanban"]` watches `.moai/state/kanban` (`events.go:30`), and the backlog file is
   `.moai/state/kanban/backlog.json` — inside it. The event already fires on backlog changes.
2. `refresh(area)` gates on `document.querySelector('[data-live="' + area + '"]')` and then
   re-fetches `window.location.href`. The marker binds a DOM region, not a route, so a `/todo`
   page carrying `data-live="kanban"` is refreshed by the existing event with no producer change.
3. A seventh event would be **worse than unnecessary**. It would watch the same directory, and
   `eventFor` (`events.go:171-183`) resolves a change to the *longest* registered watch path — two
   entries of equal length tie, and the winner falls back to map iteration order. That is
   precisely the nondeterminism `events.go:6-9` records as a fixed bug.

So the §D exclusion stands on evidence, and AC-WC15-034 pins it with an unchanged-diff assertion
on `watchMap` and `EVENTS` rather than leaving it to assumption.

**Provenance of this exclusion.** Iteration 2 was dispatched with an instruction to do the
opposite — to move the vocabulary change *into* scope, on the reading that the route decision had
made the auditor's cost unavoidable. The three measurements above were produced in answer to that
instruction and withdrew its premise, and the withdrawal was reviewed and approved rather than
taken unilaterally. The operator's decision (a separate `/todo` route) is unchanged; only the
consequential scope the dispatch had inferred from it is retracted. Recorded here because a later
reader comparing the audit's warning against this SPEC's exclusion would otherwise find the
exclusion looking like the warning was ignored.

**What did enter scope** is the route surface itself, enumerated in spec.md §C.6 and asserted by
AC-WC15-035: `app.go` `routes()`, a sixth `navRow` in `shell.templ` `rail()`, a `todo` case in
`icons.templ`'s `iconAt` switch (a missing case renders a blank glyph, since `navRow` passes the
nav id straight to `@iconAt`), the `Area` value in `screens.go`, and `nav.todo` in all four
locale maps. Two generated files follow: `shell_templ.go` and `icons_templ.go`.

**One residual limitation, recorded not fixed.** `Hub.Watch(root, …)` watches the *served*
project root, while REQ-WC15-031 resolves the backlog to the *primary checkout*. A console served
from a linked worktree therefore renders the primary's queue correctly but receives no live event
when that queue changes, and the 30s fallback poll does not engage because SSE is healthy. The
section is correct on load and on any other `kanban` event, stale in between. Widening the
watched paths is a producer change this SPEC declines (spec.md §D).

### 1.4 G-6 — `Lane int`, and why the pointer precedent does not transfer

**Decision.** `Lane int`, with `0` meaning "not a lane". Not `*int`.

**Why this shape.** Factory lanes number from 1 (`lane-1..N`), so 0 is unreachable by legitimate
data. A pointer would add a nil check at every reader to defend a state that cannot occur.

**Why `VerifyRung` does not settle it the other way.** `VerifyRung` is `*Rung` because a
*recorded-empty* rung is a reachable state distinct from an absent one (`record.go:76-88`) — the
pointer earns its indirection by carrying a distinction the value type cannot. Lane has no such
distinction: there is no meaningful "recorded lane 0". Absent → 0 → the "not recorded" marker
REQ-WC15-051 already requires, so the honest rendering survives without the pointer.

**Consequence.** Readers test `Lane == 0` for absence. AC-WC15-040's baseline half (the
pre-change tree produces role `lane` with no recoverable lane number) remains the evidence that a
passing assertion is new information.

## §2 M3 migration sequence

The order matters because two steps are silent when performed alone. Execute in this sequence:

1. **Export the reader first.** `readContextUsage` (`context_usage.go:186`) becomes the single
   exported reader. Measured baseline: `internal/statusline` currently exports **zero** readers,
   so the "exactly one" assertion in AC-WC15-021 is genuinely new state.
2. **Move the writer path.** `writeContextUsage` (`context_usage.go:128`) writes to
   `.moai/state/context-usage/<session-id>.json`. Reject a session id carrying a path separator,
   a `..` traversal, or an absolute prefix — the id is externally shaped and becomes a path
   component here (AC-WC15-052). The write stays best-effort: a refused write must not fail the
   render.
3. **Migrate `internal/cli/tokens.go` in this same milestone.** This is the step that is silent
   if skipped: `readTokensContextSnapshot` (`tokens.go:393-397`) returns `nil` on any read error,
   so a reader that can no longer find the file produces no compile error, no runtime error, and
   no log line — the snapshot block just disappears from `moai tokens` output. Delete
   `tokensContextSnapshotFilename` (`:30`) and `tokensContextSnapshot` (`:79`), consume the
   exported reader, and move `tokens_test.go:283`'s fixture with them.
4. **Delete the single-slot validation the split makes unreachable** (§1.2 consequence).
5. **Update the test assertions** — five-plus literal-path assertions across
   `internal/statusline/{builder,context_usage}_test.go`.
6. **Update both doctrine mirror pairs, then `make build`.** Four files, not two: the main rule
   and its `-detail.md` companion, each with its `internal/template/templates/…` mirror. The
   detail companion carries the snapshot field list and the validity-guard read procedure the
   main rule defers to, so updating only the main pair leaves the procedure stale — the defect D6
   named. The `.moai/README.md` mention is generic (a `state/` category row with no filename) and
   needs no change.
7. **Sweep the comment** at `internal/spec/drift_cache.go:24` and the twelve docs-site pages.

Steps 1-3 must land together. Steps 1-2 without 3 produce a green build with a silently broken
`moai tokens`; that is the single most important ordering constraint in this SPEC.

## §3 M4 resolution/adoption split

**The problem.** `resolveTodoQueueRoot` (`todo.go:66`) is pure on the git-resolvable path — it
returns `filepath.Dir(dirs.CommonDir)` and nothing else. On the fail-open branch it is not: it
calls `fallbackTodoQueueRoot` (`:89-102`), which calls `adoptLocalTodoQueue` (`:115-139`), which
performs `os.MkdirAll` (`:124`), `os.Rename` (`:128`), and `os.WriteFile` (`:139`). A console
launched in a non-git directory that reuses this function would migrate the operator's backlog as
a side effect of rendering a page — breaking REQ-WC15-002 and failing AC-WC15-032.

The branch is not hypothetical and cannot be excluded by the console: it is the fail-open path
git-unresolvable launches take by design, and a console has no way to know it is on it before
calling.

**The shape.** Two exported entry points in `internal/kanban`, not one:

| Entry point | Behaviour | Caller |
|---|---|---|
| pure resolver | resolves the primary checkout via the git common dir; falls back to the home-based root; performs **no** `MkdirAll`, `Rename`, or `WriteFile` on any branch | `internal/web` |
| adopt-then-resolve | calls the pure resolver, then performs the adoption | `internal/cli/todo.go` |

`adoptLocalTodoQueue`'s logic moves verbatim; only its call site narrows. `moai todo`'s observable
behaviour is unchanged, which is what keeps M4 shippable alone against the existing `moai todo`
tests.

**Why the AC targets the fallback branch specifically.** An acceptance criterion that only
exercised the git-resolvable path would pass on a literal move — that path is already pure, so it
proves nothing about the property being asserted. AC-WC15-031c constructs the exact preconditions
under which `adoptLocalTodoQueue` fires (no git metadata, a project-local queue present, no
fallback queue yet) and asserts the console's resolver leaves the disk untouched, while the
`moai todo` path still adopts.

**Ordering consequence.** AC-WC15-032 (M6, read-only console) transitively depends on M4
performing this split. If M4 does the literal move, M6's AC fails on the fallback branch — an
otherwise-invisible coupling between two milestones the dependency graph shows as adjacent.
