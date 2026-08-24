---
id: SPEC-WEB-TODO-QUEUE-001
title: "Backlog queue-root resolution split and a read-only /todo console route"
version: "0.1.0"
status: in-progress
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/web
lifecycle: spec-anchored
tags: web-console, kanban, todo, backlog, worktree
era: V3R6
tier: M
related_specs: [SPEC-KANBAN-TODO-CLI-001, SPEC-WEB-CONSOLE-015, SPEC-WEB-CONSOLE-REDESIGN-001]
---

# SPEC-WEB-TODO-QUEUE-001 — backlog queue root and the `/todo` route

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-08-24 | Initial draft. Carved out of SPEC-WEB-CONSOLE-015 as its own connected component — that SPEC's dependency graph already held `M4 → M6` disjoint from everything else. Inherits its resolved decisions G-4 and G-5; adds D-2 (read-through) and the conditional wording of the refresh requirement. |

## §A Background

Two halves of one change: a producer refactor that gives the backlog queue root a single
resolution both `internal/cli` and `internal/web` can import, and a read-only console route that
consumes it.

### A.1 The console has no backlog consumer, and the store is ready to be read

Measured in this tree at `dfbf828a6`:

| Claim | Evidence |
|---|---|
| `internal/web` does not reference the backlog today | `grep -rn "Backlog" internal/web` returns exactly one hit, and it is a translation string: `assets/i18n.js:483` `"f.workflow.todo.enabled.title": "Backlog queue (todo)"`. No Go file, no template. |
| The store is read-ready | `internal/kanban/backlog_store.go` exports `NewBacklogStore` (`:240`), `BacklogPathForRoot` (`:249`), `(*BacklogStore).Load` (`:299`), and `QueuedBacklogCountForRoot` (`:277`). `Load` acquires no lock; `Mutate` (`:341`) is the lock-guarded write path and is not needed here. |
| The item contract is frozen and five-field | `BacklogItem` (`:66-72`) is `{ID, Text, AddedAt, SpecID *string, State}`; `BacklogState` (`:51-61`) is `queued \| picked \| dropped`. `SpecID` is a pointer so an absent SPEC id round-trips as JSON null rather than an omitted key. |
| The store belongs to another SPEC | `SPEC-KANBAN-TODO-CLI-001` owns lock-guarded writes and id issuance. This SPEC is a read-only consumer and takes no ownership of it. |

### A.2 The queue root resolves to the primary checkout, and one branch of that resolution writes

`resolveTodoQueueRoot` (`internal/cli/todo.go:66`) is unexported and deliberately resolves the
**primary checkout** through git's common directory rather than the worktree the process runs in.
Its own comment records the measured incident that motivated it: 30 queued cards on the primary,
"queue is empty" from a linked worktree (2026-08-17). `moai web` can be launched from inside a
worktree, so the console must use the same resolution — and a second implementation of it is a
second chance to fork the queue.

The resolution has three branches, and they do not behave alike. Measured:

| Branch | Behaviour | Evidence |
|---|---|---|
| git resolvable | returns `filepath.Dir(dirs.CommonDir)`; **no side effect** | `todo.go:68-70` |
| git unresolvable, home resolvable | returns `~/.moai/todo/<key>` **after calling `adoptLocalTodoQueue`** | `todo.go:99-101` |
| git unresolvable, home unresolvable | returns `<base>/.moai/state/kanban`; no side effect | `todo.go:94-98` |

`adoptLocalTodoQueue` (`:115-139`) performs `os.MkdirAll` (`:124`), `os.Rename` (`:128`), and, on
a rename-refusing filesystem, `os.WriteFile` (`:139`). A console that performed the literal
resolution would migrate the operator's backlog as a side effect of rendering a page. The
resolution must therefore be **split**: resolving is what the console imports, adopting is what
`moai todo` keeps.

### A.3 Splitting resolution from adoption opens a read divergence, and this SPEC closes it

A **pure** resolver returns the fallback root **without** adopting. In a non-git launch context
where the operator's cards still sit at the project-local path, that produces two answers on one
tree: the console resolves to the fallback root, finds no file, and renders an empty queue, while
`moai todo` adopts first and reports N cards. Whichever ran first decides what the operator sees,
and the console keeps showing empty until `moai todo` runs.

That is the failure this SPEC opens with, reappearing — not as two implementations, but as one
resolution with two behaviours.

**Resolved as read-through (decision D-2, plan.md §F).** When the fallback root holds no queue
file and a project-local queue file exists, the pure resolver resolves to the **project-local
root**, still writing nothing. The console and `moai todo` then see the same cards, and adoption
still happens the first time `moai todo` runs. The read-through predicate mirrors the adoption
predicate already in the code — `adoptLocalTodoQueue` returns early when the fallback target
exists (`:118-120`) and when no local file exists (`:121-123`) — so the two agree by construction
rather than by coincidence.

Recording the divergence as an accepted limitation was rejected: this SPEC's opening evidence IS
that failure, so recording it would contradict the SPEC's own premise.

### A.4 The live-refresh event vocabulary does not change — a measurement, not a preference

| Claim | Evidence |
|---|---|
| The queue file already sits inside a watched directory | The queue file is `.moai/state/kanban/backlog.json`; `watchMap["kanban"]` is `.moai/state/kanban` (`internal/web/events.go:30`). |
| The client keys refresh on a DOM marker, not on a route | `app.js:644` gates on `document.querySelector('[data-live="' + area + '"]')`, then `refresh(area)` (`:648`) re-fetches. A `/todo` page carrying `data-live="kanban"` is therefore refreshed by the existing event with zero producer-side change. |
| A seventh event name would be worse than unnecessary | It would have to watch the same directory. `eventFor` (`events.go:171-183`) attributes a change to the **longest** registered watch path; two entries of equal length tie and fall back to map iteration order — the nondeterminism the function's own header comment (`:168-170`) records as a fixed bug. |
| Today's vocabulary is six names on both sides | `watchMap` holds 6 entries (`events.go:25-32`); `EVENTS` is `["spec","session","goal","verify","kanban","config"]` (`app.js:637`). |

### A.5 One limitation the refresh requirement must carry in its own body

`Hub.Watch` watches the **served** project root, while §B's resolution requirement resolves the
backlog to the **primary checkout**. A console served from a linked worktree therefore renders
the primary's queue correctly but receives no live event when that queue changes, and the fallback
poll does not engage because SSE is healthy — `POLL_MS` is 30000 (`app.js:638`) and
`startPolling()` is reachable only from the no-`EventSource` branch (`:717`) and the
three-consecutive-failure branch (`:743`).

Worktree launch is not a fringe case here — §A.2 makes it load-bearing, since it is the entire
justification for resolving to the primary checkout. A requirement that is false in a supported
configuration must say so in the requirement, not only in a constraints section. REQ-WTQ-007 is
therefore written conditionally. Widening the watched paths stays out of scope (§D) — that ruling
is accepted and unchanged; this is a wording obligation, not a scope change.

## §B Requirements (GEARS)

- **REQ-WTQ-001** — The console shall not perform any write, mutation, or lock acquisition
  against the backlog queue.
- **REQ-WTQ-002** — The console shall present the backlog queue at its own top-level `/todo`
  route, reachable from a navigation entry that is marked as the current location **while** that
  route is being served. (Resolved decision G-4, inherited; the cost is enumerated in §C.2 and is
  in scope rather than avoided.)
- **REQ-WTQ-003** — The console shall list every item the queue holds, in all three of its
  states, filtering none out, and shall render per item its identifier, its text, its state as an
  explicit state badge, and its SPEC identifier where one is attached. (Resolved decision G-5,
  inherited.)
- **REQ-WTQ-004** — The backlog queue root shall be resolved by a single resolution, relocated
  into a package that both the command layer and the console import, with the command layer
  delegating to it rather than retaining its own copy. The relocation shall separate resolving
  from adopting: the entry point the console consumes shall perform no filesystem mutation on any
  branch, and the queue-adoption side effect shall remain reachable only from the `moai todo`
  command path.
- **REQ-WTQ-005** — **When** the resolution reaches its home-based fallback root and that root
  holds no queue file while a project-local queue file exists, the entry point the console
  consumes shall resolve to the project-local root, so that the console and `moai todo` present
  the same items, and shall still write nothing.
- **REQ-WTQ-006** — **When** the backlog queue file is absent, empty, or unreadable, the console
  shall render an empty state for the section and shall not return an error response.
- **REQ-WTQ-007** — **While** the console is served from the checkout that holds the resolved
  backlog, **when** the existing `kanban` live-refresh event fires, the `/todo` route's section
  shall be re-fetched through the existing refresh path, and no new event name shall be added to
  the event vocabulary. **While** the console is served from a checkout other than the one holding
  the resolved backlog, the section shall be correct on load and on any other `kanban` event, and
  no live event is guaranteed for a change to the resolved backlog (§A.5).
- **REQ-WTQ-008** — Every user-visible string this SPEC adds shall be present in all four locale
  maps the console ships.

## §C Constraints

### C.1 Ownership and read-only posture

`SPEC-KANBAN-TODO-CLI-001` owns the backlog store: lock-guarded writes and id issuance. This SPEC
consumes `Load` only. Read-only is a standing console rule; REQ-WTQ-001 states it for this surface
specifically, and its criterion is written as **preservation** with a measured baseline
(AC-WTQ-001) rather than as a change.

The measurement that makes that baseline honest also shows why the scan needs care: a naive
`grep -rnE 'Mutate\(|acquireLock|os\.WriteFile' internal/web` returns 23 hits in this tree, all of
them in `_test.go` files or in comments that name the anti-pattern (`server.go:9`,
`projectconfig.go:158`, `projectconfig.go:221`). Scoped to production Go and excluding comment
lines, the same scan returns **0**. A criterion must state the executable form it means, not the
qualifier it wishes it could express.

### C.2 What the `/todo` route decision brought into scope

Resolved decision G-4 chose a top-level route over a panel. Every place the route and the sixth
navigation entry are declared, verified in this tree:

| Surface | Change | Measured today |
|---|---|---|
| `internal/web/app.go` `routes()` | one page route beside the existing ones | page routes registered at `:157-160` plus `/specs` at `:171` |
| `internal/web/shell.templ` `rail()` | a sixth `navRow` after `settings` | 5 `navRow` calls (`:130-134`) |
| `internal/web/icons.templ` `iconAt` | a `todo` case — `navRow` calls `@iconAt(id, 16)` with the nav id, so a missing case renders a blank glyph | 0 occurrences of `case "todo"` |
| `internal/web/screens.go` | the `Area` value that drives `aria-current` | `Area: area` at `:23` |
| `internal/web/assets/i18n.js` | `nav.todo` in four locale maps | 0 occurrences of `"nav.todo"` |

`templ generate` regenerates the corresponding `_templ.go` files.

### C.3 Schema

No schema change. `BacklogItem` and `BacklogState` are consumed as they stand (§A.1); no field is
added, renamed, or removed anywhere.

### C.4 Template-First

Go source under `internal/web`, `internal/kanban`, and `internal/cli` has no mirror under
`internal/template/templates/`, so the Template-First rule does not apply to any file this SPEC
changes.

## §D Exclusions

Explicitly out of scope. Each may be taken up separately.

### Out of Scope — live-refresh transport and event vocabulary

- Adding, replacing, or re-selecting a live-refresh transport; changing the debounce, the
  keepalive, or the poll period.
- Adding a new event name for the todo section. §A.4 measures why reuse is correct and why a
  seventh name would reintroduce a fixed bug.
- Widening the watched paths so a worktree-served console receives live events for the primary
  checkout's queue (§A.5). Separate producer change.

### Out of Scope — write paths

- Any web-initiated queue mutation: adding, picking, dropping, reordering, or editing a card.
- Backlog identifier issuance and lock-guarded writes — owned by `SPEC-KANBAN-TODO-CLI-001`.
- Changing when or whether `moai todo` adopts a project-local queue. The adoption logic changes
  call sites; its behaviour is unchanged and is asserted as such.

### Out of Scope — session telemetry and factory lanes

- Recorded model, effort, and context-window usage, and the per-lane factory progress section.
  Those belong to `SPEC-SESSION-TELEMETRY-001` and `SPEC-WEB-CONSOLE-015`; this SPEC shares no
  requirement with them.

### Out of Scope — filtering and queue views

- A state filter, a search box, sorting, or pagination over the queue. G-5 chose the complete
  audit view; a filter is a later addition over a complete list, never the reverse.

### Out of Scope — multi-project and auth

- Switching between checkouts, authentication, multi-user access. The console stays a loopback
  single-checkout single-user surface.
