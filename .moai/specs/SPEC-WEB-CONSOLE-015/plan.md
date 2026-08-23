# SPEC-WEB-CONSOLE-015 — Implementation Plan

## §A Context

Consumer: `internal/web`. Producers touched: `internal/kanban`, `internal/cli`,
`internal/statusline`. Two doctrine mirror pairs (four files: the context-window-management
rule and its `-detail.md` companion, each with its template mirror).

Tier **L**, justified in §B.

Milestones below are ordered by **decision reversibility** — the schema and path decisions
that are expensive to unwind lead, and the mechanical view assembly trails. Read M1-M3 with
care; M5-M6 are largely follow-the-shape work.

## §B Tier justification — L

The operator read the scope as L, and the measurement agrees. Against the Tier envelope:

| Signal | Measurement |
|---|---|
| Packages touched | 4 (`internal/web`, `internal/kanban`, `internal/cli`, `internal/statusline`) |
| Milestones | 6, each independently shippable |
| Files | > 10 — 3 Go producers + web viewmodel/templ/handler/route/nav/icon/i18n + 4 rule copies + ≥ 5 test files asserting the old context path |
| Cross-cutting schema change | 2 — `kanban.Record` field additions, and a state-file path relocation with a documented external consumer |
| Consumers outside this repo's code | 2 — the doctrine rule and its detail companion, read by agents and the orchestrator |

Any one of the last two alone would push past M. The context-usage relocation in particular
carries a documented read procedure in an always-loaded rule; that is not an M-sized change.

## §C Producer / consumer split — recommendation

The prior investigation (`.moai/reports/webredesign/moai-web-menu-spec.md` §6.1) recommended
lifting the producer work into a **sibling SPEC**, so that a block is attributable to one
side. I recommend against that here, with one carve-out.

**Keep producer and consumer in this SPEC, split by milestone.** Three reasons:

1. **A producer SPEC here would have no testable acceptance.** Every field M1 adds has exactly
   one consumer — this console. Split out, its acceptance criteria degrade to "the field
   exists and round-trips", which is the weak-AC shape: it passes without observing the
   behaviour anyone wanted.
2. **Attribution is already available at the milestone boundary.** M1-M4 land as producer
   commits and M5-M6 as consumer commits. A blocked run names the milestone; nothing about a
   second SPEC id sharpens that.
3. **Two SPEC bodies would fork the join contract.** The `workers.json` PID →
   active-sessions → `Record` chain (REQ-WC15-043) is simultaneously a producer obligation
   (record the lane number and card id) and a consumer obligation (perform the join). Written
   in two documents it drifts; §A.5 keeps it in one.

**Carve-out considered and closed — M3 stays here.** The context-usage relocation is the one
piece with consumers beyond this console: two doctrine files and their mirrors, the statusline
itself, and twelve docs-site pages. The carve-out was conditional on a compat window making it
release-spanning lifecycle work; resolved decision G-3 chose a hard cut, so it is a single
in-tree change with every consumer enumerated (spec.md §C.3) and testable here. Resolved
decision G-2 (§G) — no longer open.

## §D Milestones

Each milestone is independently shippable: it leaves the tree green and adds no half-state.

### M1 — `kanban.Record` schema extension (producer, highest reversibility cost)

Add four `omitempty` fields to `internal/kanban/record.go`: model, effort, lane number, card
id. Nothing renamed, nothing removed — `record.go:45` `@MX:ANCHOR` binds readers this package
cannot see.

The lane number is added as its **own field** rather than by widening `WithRole` to accept
`lane-3`. `WithRole` (`record.go:116-130`) deliberately drops unrecognised role strings so
consumers never defend against arbitrary launch-label text; widening it to a pattern match
reopens exactly that. A separate `Lane int` keeps the drop-unknown guard intact and keeps role
and lane number observably distinct — the same reasoning `RoleDeclaration` already applies to
`Role` versus `Label` (`role.go:47-52`).

The type is `Lane int` with `0` meaning "not a lane", not `*int` — resolved decision G-6 (§G).

Ships alone: additive fields with no writer are inert.

### M2 — Launcher threads model, effort, lane, and card id (producer)

Widen `recordKanbanSession` (`internal/cli/kanban.go:472`) past its current
`(specID, backend, role)` and update its eight callers (`cc.go` 161/175/192/208, `glm.go`
224/237/250/264). Model and effort resolve from the existing `internal/config/profile.go`
`ModelEffort` / `EffectiveProfile` surface at launch.

Card id is derived at launch from the basename of `git rev-parse --show-toplevel`, with an
explicit environment-variable override consulted first, and left empty when neither yields a
value — resolved decision G-1 (§G). No new on-disk producer.

Ships alone: records gain fields no reader yet displays.

### M3 — Context-usage per-session split (producer, largest blast radius)

`.moai/state/context-usage.json` → `.moai/state/context-usage/<session-id>.json`, plus one
**exported** reader in `internal/statusline` (REQ-WC15-021) so no second copy of the schema
appears (REQ-WC15-022). Treat §C.3 of spec.md as the checklist, not as a rename: writer, reader,
struct, call site, five-plus test assertions, the `internal/cli/tokens.go` duplicate reader and
its `tokens_test.go` fixture, **two** doctrine mirror pairs (the main rule and its
`-detail.md` companion, four files) plus `make build`, twelve docs-site pages, and the
`drift_cache.go:24` comment. `.moai/README.md` needs no change (its mention is generic).

**Hard cut, no dual-write window** — resolved decision G-3 (§G). **M3 stays inside this SPEC**
rather than being promoted to a sibling — resolved decision G-2 (§G), which follows from G-3.

The single-slot validation (`isFreshForSession`, the `writer_pid` discriminator, the
same-payload check) exists because one file served N sessions. With the session id in the path
most of it becomes unreachable; delete what the split makes dead rather than leaving it as
decoration.

Ships alone.

### M4 — Relocate queue-root resolution, splitting resolution from adoption (producer refactor)

Move `resolveTodoQueueRoot` (`internal/cli/todo.go:66`) and its `fallbackTodoQueueRoot` /
`adoptLocalTodoQueue` support into `internal/kanban`, exported. `internal/cli/todo.go`
delegates. `internal/kanban` is the natural home — `BacklogPathForRoot` and
`QueuedBacklogCountForRoot` already live there and already take a root.

**This is not a literal move.** The relocation must split the resolution from the adoption side
effect, and a run that performs the literal move fails AC-WC15-031c and AC-WC15-032. Verified
in this tree: `resolveTodoQueueRoot` → on a git-unresolvable context → `fallbackTodoQueueRoot`
(`:89-102`) → `adoptLocalTodoQueue` (`:115-139`), which performs `os.MkdirAll` (`:124`),
`os.Rename` (`:128`), and `os.WriteFile` (`:139`). The git-resolvable path is pure; the write
lives exactly on the fail-open branch a console cannot exclude, so a console launched in a
non-git directory would migrate the operator's backlog while rendering a page — breaking
REQ-WC15-002.

The shape M4 must produce: **two** exported entry points, not one.

- a **pure resolver** — resolves the primary checkout, falls back to the home-based root, and
  performs no `MkdirAll`, `Rename`, or `WriteFile` on any branch. This is what `internal/web`
  imports.
- an **adopt-then-resolve** entry point that calls the pure resolver and then performs the
  adoption. This is what `internal/cli/todo.go` calls, so `moai todo`'s behaviour is unchanged.

The adoption logic itself moves verbatim; only its call site is narrowed.

The point is that one resolution exists, not that the console gets a copy: the measured
failure (30 queued cards on the primary, "queue is empty" from a worktree, 2026-08-17) is what
a second implementation reproduces.

Ships alone: a pure move with identical behaviour, verifiable by the existing `moai todo` tests.

### M5 — Console consumer: session telemetry and factory lanes

Fill the three `RoleVM` placeholders at `internal/web/viewmodel_ops.go:250-256` from M1-M3.
Add the factory lane section: `LoadFactoryRegistry` → PID → `session.Entry` → `Record`, beside
the `ChainRoles` iteration rather than inside it (lanes are not chain roles). Keep
`StageEstimated` and surface it.

Depends on M1, M2, M3.

### M6 — Console consumer: `/todo` route

Read-only backlog section at its **own top-level route** `/todo` with a sixth nav entry —
resolved decision G-4 (§G). `NewBacklogStore(BacklogPathForRoot(root)).Load()` with M4's **pure**
resolver, all three `BacklogState` values listed with a state badge (resolved decision G-5).

Wired to the **existing** `kanban` refresh area — the section carries `data-live="kanban"` and
no event name is added. The reasoning, and why a seventh event would be worse rather than
merely unnecessary, is spec.md §C.6; the pin is AC-WC15-034's `watchMap`/`EVENTS`
unchanged-diff half.

The route's own surface (spec.md §C.6): `app.go` `routes()`, `shell.templ` `rail()` sixth
`navRow`, `icons.templ` `iconAt` `todo` case, `screens.go` `Area`, and `nav.todo` in four
locales. `templ generate` regenerates `shell_templ.go` and `icons_templ.go`.

Depends on M4 only — **not** on M1-M3. M6 can ship before M5 if the producer work stalls.

## §E Dependency graph

```
M1 ──> M2 ──┐
            ├──> M5
M3 ─────────┘

M4 ──> M6
```

## §F Anti-patterns to avoid

- **Re-deciding the transport.** The card asks for a choice that was already made and built
  (spec.md §A.1). Editing `events.go` or the `connect()` block is out of scope.
- **Duplicating the context-usage struct** in `internal/web` instead of exporting a reader.
  Two declarations of one on-disk schema is how a format forks.
- **Reimplementing queue-root resolution** in `internal/web`. Same failure mode, different
  package.
- **Widening `WithRole` to pattern-match `lane-<n>`.** It drops unknown roles on purpose.
- **Displaying an empty string where a value was never recorded.** `RoleVM.ContextPct = -1` is
  the existing "not recorded" sentinel; the model and effort cells need the same honesty
  (REQ-WC15-012).
- **Treating the context-usage move as a rename.** It has a documented external read procedure.
- **Running the full Go suite locally.** Target the affected packages and read CI for the
  full-suite verdict.

## §G Resolved decisions

All six decisions previously handed back here are **resolved**, and every clarification marker
this section once carried has been deleted rather than annotated. Each entry records the answer,
the reasoning, and what was rejected, so a later reader inherits the decision rather than the
question.

**G-1. Card-id producer → derive from the worktree path, with an environment-variable
override.** The dispatch format fixes `wt: .claude/worktrees/<card-id>` and the kanban-dispatch
protocol requires that the worktree directory keep the card id, so the value already exists at
the one place every card-carrying lane stands, and `git rev-parse --show-toplevel` already
reaches it. Zero new on-disk producer surface. A lane outside a card worktree yields nothing —
but such a lane is already a dispatch violation, and the empty field renders through the
existing "not recorded" honesty path (REQ-WC15-012 shape) rather than as a guess.
*Rejected:* a launch-time env var as the **sole** source — factory routes a card whole to a
free lane *after* launch, so process env cannot carry a per-card value without relaunching a
lane per card; it survives only as the override, which costs no schema change. *Rejected:* the
lead writing it into the record after launch — accurate while the lead is disciplined, silently
stale the first time it is not. Lands in REQ-WC15-042 and M2.

**G-2. M3 stays inside this SPEC.** This follows from G-3 rather than standing on its own: the
§C carve-out promoted M3 to a sibling only if the compat window made it release-spanning
lifecycle work. A hard cut is a single in-tree change whose every consumer is enumerated
(spec.md §C.3, after the D6 correction) and testable here. Splitting it out would leave a SPEC
whose acceptance is the weak "the path moved and round-trips" shape — the same failure mode §C
reason 1 already names, and the one D2 caught this SPEC committing once.

**G-3. Hard cut for the context-usage path — no dual-write window.** Four reasons, in the order
that decided it: (i) every reader is in-tree and now fully enumerated — no plugin, no external
API, no persisted consumer reads this path; (ii) the file is render-ephemeral session telemetry
regenerated on every statusline render, so a cut loses at most one render's snapshot and there
is no durable data to migrate; (iii) dual-write would keep the last-writer-wins single slot
alive, which is precisely the trap spec.md §A.3 documents with a live observation (session
`368a2bd9…` at 260,000 tokens replaced by session `e463a3c9…` at 0) — a compat window whose
purpose is to preserve a known defect; (iv) dual-write contradicts REQ-WC15-024's own
instruction to drop the single-slot validation steps, keeping `isFreshForSession` and the
`writer_pid` discriminator alive as decoration. Lands in M3, AC-WC15-020, AC-WC15-024.

**G-4. The todo section gets its own `/todo` route and nav entry, not a panel on `/kanban`.**
This decision goes **against** the plan-auditor's recommendation, and the cost the auditor named
is accepted rather than avoided — it is in scope, enumerated in spec.md §C.6, and pinned by
AC-WC15-035: one route, a sixth nav row, an `iconAt` case, an `Area` value, and `nav.todo` in
four locales. The queue is an operator surface in its own right and is addressable and
shareable as a URL, which a panel is not.

The auditor's concrete objection — that a separate route "requires an event-vocabulary change
that §D Out of Scope explicitly forbids" — does **not** hold against the tree, and this was
measured rather than argued. The backlog file lives at `.moai/state/kanban/backlog.json`, inside
the directory `watchMap["kanban"]` already watches (`events.go:30`), and `refresh(area)` gates on
a `data-live` DOM marker rather than on the route — so the existing event covers `/todo` with no
producer change. Adding a seventh event name would be worse than unnecessary: it would have to
watch the same directory, and `eventFor` (`events.go:171-183`) resolves ties by *longest* watch
path, so two equal-length entries fall back to map iteration order — the exact nondeterminism
the file's header comment records as a fixed bug. The §D exclusion therefore stands on evidence,
and no AC contradicts it; AC-WC15-034 asserts the unchanged `watchMap`/`EVENTS` diff as the pin.

One residual limitation is recorded rather than fixed (spec.md §C.6): a console served from a
linked worktree watches that worktree's state directory while resolving the backlog to the
primary checkout, so it will not receive a live event for a primary-checkout queue change.
Correct on load and on any other `kanban` event, stale in between. Widening the watched paths is
a producer change this SPEC declines (§D).

**G-5. The todo section shows all three states, each with a state badge.** The audit view rather
than the working view: the kanban lead's card cross-check consumes `picked` and `dropped` as an
audit trail, and a `queued`-only list cannot answer "where did card X go". A filter is a later
addition over a complete list, never the reverse. Lands in REQ-WC15-030 and AC-WC15-030, whose
assertion now includes the badge — the badge is part of the resolution, so leaving it
unasserted would have left half the decision unobserved.

**G-6. `Lane int`, with `0` meaning "not a lane" — not `*int`.** Factory lanes number from 1
(`lane-1..N`), so 0 is unreachable by legitimate data and the pointer would defend a state that
cannot occur while adding a nil check at every reader. The `VerifyRung` precedent does not
transfer: there a *recorded-empty* rung is a reachable state distinct from an absent one
(`record.go:76-88`), which is what earns the pointer. Here absent → 0 → the "not recorded"
marker REQ-WC15-051 already requires, so the honest rendering is preserved without the
indirection. Lands in REQ-WC15-040 and M1.

## §H Tier-L artifact set

Complete: `spec.md`, `plan.md`, `acceptance.md`, `research.md`, `design.md`.

`design.md` **is written**. An earlier draft of this plan argued it would only restate spec.md
§A/§C and this plan's §D/§G; the plan-auditor accepted that rationale for the schema and
join-topology content and rejected it for the cross-cutting decisions, and the evidence it was
right is that two acceptance criteria (AC-020/024 and AC-030) had silently hard-coded one branch
of a decision that had no home to be recorded in. `design.md` is scoped to exactly that gap —
the resolved G-1/G-3/G-4/G-6 decisions and their consequences, the M3 migration sequence, and
the M4 resolution/adoption split — and deliberately does **not** restate the schema shape or the
join topology, which spec.md §A/§C already carries.

## §I Cross-references

- `.moai/reports/webredesign/moai-web-menu-spec.md` §4.6, §5, §6.1, §7, §8 — the prior
  investigation. §1-§6 largely shipped; check claims against code.
- `.moai/reports/webredesign/moai-web-redesign-brief.md` — visual constraints.
- `SPEC-KANBAN-TODO-CLI-001` — owner of the backlog store.
- `.claude/rules/moai/workflow/context-window-management.md` § Detection Heuristics and
  `context-window-management-detail.md` — the consumer contract M3 must move, both mirrored
  under `internal/template/templates/`.
