# SPEC-WEB-TODO-QUEUE-001 — Implementation Plan

## §A Context

Producer touched: `internal/cli` (delegation) and `internal/kanban` (new home for the resolution).
Consumer: `internal/web`. No doctrine file, no always-loaded rule, no template mirror (spec.md
§C.4).

Tier **M**, justified in §B.

Milestones are ordered by **decision reversibility**. M1 fixes the shape of a shared resolution
and the semantics of its fallback branch — the decisions that are expensive to unwind and that a
later reader will inherit. M2 fixes a user-facing surface. M3 is largely follow-the-shape render
work. Read M1 with care.

## §B Tier justification — M

| Signal | Measurement |
|---|---|
| Packages touched | 3 (`internal/cli`, `internal/kanban`, `internal/web`) |
| Files | 12-14 — enumerated in §C; an enumeration, not a measured diff |
| Schema change | none (spec.md §C.3) |
| Always-loaded doctrine touched | none |
| Consumers outside this repo's code | none |
| Milestones | 3, each independently shippable |

Files land in the 5-15 band, no schema moves, and nothing an agent reads at session start changes.
That is Tier M: three artifacts, threshold 0.80.

This SPEC exists because its parent exceeded a Tier L budget. Its own budget — 8 requirements and
11 criteria against Tier M ceilings of 16 and 16 — leaves room for a re-audit to add without
hitting a ceiling. If scope grows past that, the answer is to cut scope, not to raise the ceiling.

## §C File enumeration

| File | Milestone | Change |
|---|---|---|
| `internal/kanban/todo_root.go` (new) | M1 | the relocated resolution: pure resolver + adopt-then-resolve entry point |
| `internal/kanban/todo_root_test.go` (new) | M1 | branch coverage incl. the fallback and read-through branches |
| `internal/cli/todo.go` | M1 | delete the local resolution, delegate |
| `internal/cli/todo_test.go` | M1 | existing adoption tests re-pointed at the command-path entry point |
| `internal/web/app.go` | M2 | one page route |
| `internal/web/shell.templ` (+ `shell_templ.go`) | M2 | sixth `navRow` |
| `internal/web/icons.templ` (+ `icons_templ.go`) | M2 | `case "todo"` |
| `internal/web/screens.go` | M2 | the `/todo` handler and its `Area` value |
| `internal/web/assets/i18n.js` | M2 | `nav.todo` plus the section's strings, four locales |
| `internal/web/screens.templ` | M3 | the section markup, badges, empty state, `data-live="kanban"` |
| `internal/web/viewmodel_ops.go` (or a new sibling) | M3 | the read-only view model built from `Load` |
| `internal/web/*_test.go` (new or extended) | M3 | render, empty-state, and read-only assertions |

Twelve entries; the `_templ.go` regenerations bring it to fourteen. `templ generate` produces
those — they are committed, not hand-edited.

## §D Milestones

Each milestone leaves the tree green and adds no half-state.

### M1 — Relocate the queue-root resolution, splitting resolving from adopting (highest reversibility cost)

Move `resolveTodoQueueRoot` (`internal/cli/todo.go:66`) and its `fallbackTodoQueueRoot` /
`adoptLocalTodoQueue` support into `internal/kanban`, exported. `internal/cli/todo.go` delegates.
`internal/kanban` is the natural home: `BacklogPathForRoot` (`backlog_store.go:249`) and
`QueuedBacklogCountForRoot` (`:277`) already live there and already take a root.

**This is not a literal move.** A run that performs one fails AC-WTQ-006 and AC-WTQ-007. The shape
M1 must produce is **two** exported entry points, not one:

- a **pure resolver** — the one `internal/web` imports. It resolves the primary checkout, falls
  back to the home-based root, and performs no `MkdirAll`, `Rename`, or `WriteFile` on **any**
  branch. Note that the resolution has three branches, not two (spec.md §A.2); "any branch"
  includes the home-unresolvable one at `todo.go:94-98`.
- an **adopt-then-resolve** entry point — the one `internal/cli/todo.go` calls. It invokes the pure
  resolver and then performs the adoption, so `moai todo`'s behaviour is unchanged.

The adoption logic itself moves **verbatim**; only its call site narrows. AC-WTQ-008 is the pin
that the behaviour did not drift while the call site did.

**Read-through (decision D-2, §F).** The pure resolver additionally returns the **project-local**
root when the fallback root holds no queue file and a project-local one exists — still writing
nothing. The predicate is the mirror of `adoptLocalTodoQueue`'s own early returns
(`todo.go:118-120` target exists, `:121-123` no local file), which is what makes the console and
`moai todo` agree by construction rather than by luck. Without it, the split closes a write hazard
and opens a read divergence (spec.md §A.3).

The point is that **one** resolution exists, not that the console gets a copy: the measured failure
this SPEC opens with — 30 queued cards on the primary, "queue is empty" from a worktree — is what a
second implementation reproduces.

Ships alone: a behaviour-preserving relocation, verifiable by the existing `moai todo` tests plus
the new branch tests.

### M2 — The `/todo` route surface (user-facing)

The route, the sixth navigation row, the icon case, the `Area` value, and the four-locale strings —
spec.md §C.2 is the checklist. At the end of M2 the route serves a page; it may serve an empty
section, because M3 fills it.

Two details that fail silently if missed: `navRow` calls `@iconAt(id, 16)` with the nav id, so a
missing `case "todo"` renders a blank glyph rather than an error (measured: 0 such cases today);
and `templ generate` must run, or the `.templ` edit is invisible to the binary.

Ships alone. Depends on nothing — it can land before M1.

### M3 — The read-only section (follow-the-shape)

`NewBacklogStore(BacklogPathForRoot(root)).Load()` against M1's **pure** resolver. All three
`BacklogState` values listed with a state badge (decision G-5, §F), each row carrying id, text,
badge, and SPEC id where attached. Absent, empty, and malformed files render the empty state at
200 rather than an error.

The section carries the **existing** `data-live="kanban"` marker and adds no event name. The
reasoning — and why a seventh name would be worse than merely unnecessary — is spec.md §A.4; the
pin is AC-WTQ-010's unchanged-diff half.

Depends on M1 and M2.

## §E Dependency graph

```
M1 ──┐
     ├──> M3
M2 ──┘
```

M1 and M2 are independent and may land in either order or in parallel — they share no file. M3 is
the only place their outputs meet.

## §F Resolved decisions

Recorded as answers, not as open questions. Two are inherited from SPEC-WEB-CONSOLE-015 and are
restated here so this SPEC is readable alone; one is new to this SPEC.

**G-4 (inherited). The todo section gets its own `/todo` route and navigation entry, not a panel
on `/kanban`.** The queue is an operator surface in its own right, addressable and shareable as a
URL, which a panel is not. The cost the plan-auditor named is **accepted rather than avoided**: it
is in scope, enumerated in spec.md §C.2, and pinned by AC-WTQ-002 and AC-WTQ-011 — one route, a
sixth nav row, an `iconAt` case, an `Area` value, and `nav.todo` in four locales.

The auditor's concrete objection — that a separate route forces an event-vocabulary change the
exclusions forbid — does not hold against the tree, and this was measured rather than argued. See
the next entry.

**The event vocabulary does not change, and this is a measurement, not a preference.** The backlog
file is `.moai/state/kanban/backlog.json`, inside the directory `watchMap["kanban"]` already
watches (`internal/web/events.go:30`), and the client's `refresh(area)` gates on a `data-live` DOM
marker (`app.js:644`) rather than on a route — so a `/todo` page carrying `data-live="kanban"` is
refreshed with zero producer-side change. A seventh event name would be **worse** than
unnecessary: it would have to watch the same directory, and `eventFor` (`events.go:171-183`)
attributes a change to the longest registered watch path, so two equal-length entries tie and fall
back to map iteration order — the nondeterminism that function's own header comment records as a
fixed bug. The exclusion therefore stands on evidence, and AC-WTQ-010 pins it mechanically.

**G-5 (inherited). The section shows all three states, each with a state badge.** The audit view
rather than the working view: the lead's card cross-check consumes `picked` and `dropped` as an
audit trail, and a `queued`-only list cannot answer "where did card X go". A filter is a later
addition over a complete list, never the reverse. The badge is part of the resolution, so
AC-WTQ-003 asserts it — leaving it unasserted would leave half the decision unobserved.

**D-2 (new). The pure resolver reads through to the project-local root rather than reporting an
empty queue.** Splitting resolving from adopting closes a write hazard and, unaddressed, opens a
read divergence: in a non-git launch context the console would render empty while `moai todo`
reports N, with whichever ran first deciding what the operator sees (spec.md §A.3). Read-through
returns the project-local root when the fallback root holds no queue file, writes nothing, and
leaves adoption to happen on the first `moai todo` run exactly as today.

*Rejected:* recording the divergence as an accepted limitation, the way the stale-event limitation
is recorded. The two are not alike. The stale-event limitation is staleness between two correct
renders; this one is a wrong render that persists until an unrelated command is run — and it is
the very failure this SPEC's opening evidence describes, so recording it would contradict the
SPEC's own premise.

*Rejected:* having the console call the adopt-then-resolve entry point. It renders a page; it must
not move the operator's files.

## §G Anti-patterns to avoid

- **Performing the literal move.** Relocating `resolveTodoQueueRoot` unchanged carries
  `adoptLocalTodoQueue` into the console's call path, and the console then migrates the operator's
  backlog while rendering a page. AC-WTQ-006 is the guard.
- **Splitting the branches but forgetting the third one.** The resolution has three branches
  (spec.md §A.2). "No mutation on any branch" includes the home-unresolvable one.
- **Reimplementing queue-root resolution inside `internal/web`.** Same failure mode as today,
  different package. One resolution is the point.
- **Shipping the split without read-through.** It trades a write bug for a read bug (D-2).
- **Changing `adoptLocalTodoQueue`'s behaviour while moving it.** Its call site narrows; its logic
  is verbatim. AC-WTQ-008 fails otherwise.
- **Adding a seventh event name for the todo section.** It reintroduces a fixed bug (§F).
- **Filtering the list to `queued`.** That is the working view; G-5 chose the audit view.
- **Editing a `.templ` file without running `templ generate`.** The change compiles and does
  nothing.
- **Running the full Go suite locally.** Target the affected packages and read CI for the
  full-suite verdict.

## §H Cross-references

- `SPEC-KANBAN-TODO-CLI-001` — owner of the backlog store; this SPEC is a read-only consumer.
- `SPEC-WEB-CONSOLE-015` — the parent this SPEC was carved out of; retains the session-telemetry
  and factory-lane axes and shares no requirement with this one.
- `.moai/reports/t207/spec-split-design.md` §4 — the ratified split design.
- `.moai/reports/t207/plan-audit-iter2-independent.md` F3, F6, F11, F13 — the findings this SPEC's
  text answers: the read divergence (D-2), the conditional refresh requirement (REQ-WTQ-007), the
  mechanical form of the delegation criterion (AC-WTQ-004), and GEARS form.
