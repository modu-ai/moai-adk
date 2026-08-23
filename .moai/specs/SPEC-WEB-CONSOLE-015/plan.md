# SPEC-WEB-CONSOLE-015 — Implementation Plan

## §A Context

Consumer: `internal/web`. Producers touched: `internal/kanban`, `internal/cli`,
`internal/statusline`. One doctrine rule and its template mirror.

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
| Files | > 10 — 3 Go producers + web viewmodel/templ/handler/i18n + 2 rule copies + ≥ 5 test files asserting the old context path |
| Cross-cutting schema change | 2 — `kanban.Record` field additions, and a state-file path relocation with a documented external consumer |
| Consumers outside this repo's code | 1 — the doctrine rule read by agents and the orchestrator |

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

**Carve-out — M3 may be promoted to a sibling.** The context-usage relocation is the one piece
with consumers beyond this console: the doctrine rule, the statusline itself, and twelve
docs-site pages. If the doctrine edit turns out to require re-deriving the Detection
Heuristics read procedure rather than re-pointing it, M3 is a SPEC of its own and this one
depends on it. That call is an open decision (§G-2), not a silent one.

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

Ships alone: additive fields with no writer are inert.

### M2 — Launcher threads model, effort, lane, and card id (producer)

Widen `recordKanbanSession` (`internal/cli/kanban.go:472`) past its current
`(specID, backend, role)` and update its eight callers (`cc.go` 161/175/192/208, `glm.go`
224/237/250/264). Model and effort resolve from the existing `internal/config/profile.go`
`ModelEffort` / `EffectiveProfile` surface at launch.

Card id has no producer today — see §G-1; M2 cannot close until that decision lands.

Ships alone: records gain fields no reader yet displays.

### M3 — Context-usage per-session split (producer, largest blast radius)

`.moai/state/context-usage.json` → `.moai/state/context-usage/<session-id>.json`, plus one
**exported** reader in `internal/statusline` (REQ-WC15-021) so no second copy of the schema
appears (REQ-WC15-022). Treat §C.3 of spec.md as the checklist, not as a rename: writer, reader,
struct, call site, five-plus test assertions, the doctrine rule, its template mirror plus
`make build`, twelve docs-site pages, and `.moai/README.md`.

The single-slot validation (`isFreshForSession`, the `writer_pid` discriminator, the
same-payload check) exists because one file served N sessions. With the session id in the path
most of it becomes unreachable; delete what the split makes dead rather than leaving it as
decoration.

Ships alone; it is also the milestone most likely to become a sibling SPEC (§G-2).

### M4 — Relocate queue-root resolution (producer refactor, no behaviour change)

Move `resolveTodoQueueRoot` (`internal/cli/todo.go:66`) and its `fallbackTodoQueueRoot` /
`adoptLocalTodoQueue` support into `internal/kanban`, exported. `internal/cli/todo.go`
delegates. `internal/kanban` is the natural home — `BacklogPathForRoot` and
`QueuedBacklogCountForRoot` already live there and already take a root.

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

### M6 — Console consumer: todo section

Read-only backlog section: `NewBacklogStore(BacklogPathForRoot(root)).Load()` with the M4 root,
wired to the existing `kanban` refresh area. Four-locale i18n for every new string.

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

## §G Open decisions — handed back, not resolved

**G-1. Where does the card id come from?** No per-lane card id exists on disk (`MOAI_KANBAN_ID`
is the *run* id, `envkeys.go:167-173`). Three candidates:

- (a) a new env var set by the lead at dispatch, read at launch — explicit, but a new producer
  surface and a new way to be absent;
- (b) the lead writes it into the record after launch — accurate, dependent on lead discipline,
  and the record goes stale silently when the lead forgets;
- (c) derive from the worktree directory name, since the dispatch format already fixes
  `wt: .claude/worktrees/<card-id>` and `git rev-parse --show-toplevel` is already available —
  zero new producer surface, but it silently yields nothing for a lane that is not in a card
  worktree.

  My lean is (c) with (a) as an override. Operator call.
  [NEEDS CLARIFICATION: card-id producer — env var, lead-written, or worktree-derived]

**G-2. Is M3 a sibling SPEC?** §C carve-out. Cheap to decide before run-phase, expensive after.
[NEEDS CLARIFICATION: promote the context-usage split to a sibling SPEC]

**G-3. Compatibility window for the context-usage path.** Dual-write both the old single slot
and the new per-session file for one release, or cut hard? Dual-write keeps any unmigrated
reader working and keeps the last-writer-wins slot alive as a trap; a hard cut is cleaner and
breaks a reader nobody has enumerated outside §C.3.
[NEEDS CLARIFICATION: dual-write window vs hard cut for context-usage]

**G-4. Where does the todo section live?** Its own `/todo` route and nav entry (shareable,
one more nav item), or a panel on `/kanban` (fewer top-level items, the queue seen beside the
board it feeds). Nav currently carries five entries (`i18n.js` `nav.overview` … `nav.settings`).
[NEEDS CLARIFICATION: /todo route vs panel on /kanban]

**G-5. Does the todo section show dropped cards?** `BacklogState` has three values. Showing only
`queued` is the operator's working view; showing all three is the audit view.
[NEEDS CLARIFICATION: todo section state filter]

**G-6. Lane number field type.** `Lane int` with 0 meaning "not a lane", or `*int` so
"not recorded" stays distinct from lane 0. `VerifyRung` already chose the pointer for exactly
this reason (`record.go:76-88`). Lane numbering starts at 1, so `int` may be sufficient —
but the precedent argues the other way.
[NEEDS CLARIFICATION: Lane int vs *int]

## §H Deferred Tier-L artifacts

`research.md` is present. `design.md` is **not** written: its content — the schema shape, the
join topology, the split's blast radius — is carried by spec.md §A / §C and this plan's
§D / §G, and a separate design document would restate them. If the plan-auditor requires the
Tier-L artifact set literally, that is a gap to close before run-phase entry, not an omission
by oversight.

## §I Cross-references

- `.moai/reports/webredesign/moai-web-menu-spec.md` §4.6, §5, §6.1, §7, §8 — the prior
  investigation. §1-§6 largely shipped; check claims against code.
- `.moai/reports/webredesign/moai-web-redesign-brief.md` — visual constraints.
- `SPEC-KANBAN-TODO-CLI-001` — owner of the backlog store.
- `.claude/rules/moai/workflow/context-window-management.md` § Detection Heuristics — the
  consumer contract M3 must move.
