---
id: SPEC-TODO-HOLD-STATE-001
title: "A held card the machine cannot pick — a fourth backlog state with operator-only hold/unhold verbs, a rebuilt CHECK constraint behind a schema_version bump, and positive state enumeration on every lease surface"
version: "0.1.0"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec (card t1308)
priority: P3
phase: "v3.2.0 target"
module: "internal/kanban, internal/cli, internal/statusline, internal/web, .claude/skills/moai/workflows/gtd.md, .claude/commands/moai/todo.md, internal/template/templates/.claude/skills/moai/workflows/gtd.md"
lifecycle: spec-anchored
tags: "kanban, todo, hold, state-machine, sqlite, table-rebuild, schema-version, positive-enumeration, machine-lease"
tier: M
related_specs:
  - SPEC-TODO-SQLITE-001
  - SPEC-TODO-DESTRUCTIVE-GUARD-001
  - SPEC-TODO-LANDING-EVIDENCE-001
  - SPEC-KANBAN-TODO-CLI-001
---

# SPEC: A held card the machine cannot pick

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-29 | Initial plan-phase authoring (card t1308, todo 로직 개선 P3-①), measured in worktree `.moai/worktrees/t1308`, branch `WT-todo-hold-state`, at HEAD `8a969dfc0`. The full measured basis (state vocabulary, schema, selection-predicate inventory, migration mechanics, doc surfaces) is exported at `.moai/reports/t1308/predicate-sweep.md`. One card premise is falsified in this tree: the claim "internal/kanban 이력상 ALTER TABLE 0건" does not hold — `ensureLandingColumn` (`backlog_sqlite.go:439-455`, t359, commit `3bcb0c33a`) is one guarded `ALTER TABLE ... ADD COLUMN`. The correction does not change the migration decision (an `ADD COLUMN` cannot widen a CHECK; the rebuild path stands — §B.2). Two lead memos are folded in as requirements, not prose: the negative-predicate full-enumeration sweep (REQ-THS-011/012) and the rebuild safety contract named as REQs (REQ-THS-002/003), inheriting the JSON→SQLite migration discipline of SPEC-TODO-SQLITE-001. |

> **Provenance discipline.** Every `file:line` citation below was measured at HEAD `8a969dfc0` in
> this worktree. Nothing is carried from another tree or time. The selection-predicate inventory in
> `.moai/reports/t1308/predicate-sweep.md` §5 is a starting inventory measured at plan time; the
> run-phase sweep (plan.md M3) re-runs it at run-phase HEAD and owns the exhaustive enumeration.

## §A Context

### A.1 The defect: a hold that lives in prose is invisible to the machine

Today the only way to hold a card out of the queue is prose in the card text — the t1158
「[보류 — …]」 pattern. The queue's storage layer knows three states (`queued`, `picked`,
`dropped` — `backlog_store.go:59-66`), and every machine selector keys on those states alone. A
prose hold reads, to every machine consumer, as a queued card: it remains a pick candidate for
`next`, for the factory lease, and — measured, not inferred — for the auto-done scan
(`todo_autodone.go:283` admits `Queued||Picked` candidates by a NEGATIVE filter, so a prose-held
card is an auto-done close candidate today). The defect becomes acute once machine leasers land:
after t1240 (factory next) and t1294 (Codex `-f` lane claim), a machine reads the queue and picks
without a human reading the prose first.

### A.2 Why a fourth state, not a marker convention

The drop/undrop pair already established the correct shape for a parked-but-alive card: the STATE
is the authority, not a text marker (`todo_drop.go:141-148`). A `[HOLD]` marker would repeat the
exact defect — text is what machines do not read. `hold` becomes a fourth value of the existing
`BacklogState` enum, gaining for free everything the enum already carries: the CHECK constraint's
DB-level guard, `list --json`'s machine-filterable `state` field, and the refusal discipline of
`Mutate` (byte-identical on refusal).

### A.3 What the schema already anticipated

The store's own DDL comment (`backlog_sqlite.go:100-102`) names this SPEC's central cost before
this card existed: "SQLite cannot ALTER a CHECK constraint, so admitting a fourth state would need
a table rebuild on every operator queue in the field." The rebuild is therefore not an accident of
this design — it is the price the schema accepted when it chose a constrained live enum over an
open text column, and the `schema_version` stamp (`backlog_sqlite.go:47-50`) plus its
refuse-unknown-version open path (`backlog_sqlite.go:288-291, 393-396`) is the machinery the
schema already carries for exactly this change.

### A.4 Coordination frame (lead-fixed, not re-opened)

- t1306 (`--auto` admission) is QUEUED-ONLY by its own contract; hold is auto-skip compatible and
  t1306 needs no change.
- t1240 (factory next) / t1294 (Codex `-f` lane claim): the machine-lease predicate is pinned as a
  POSITIVE `state='queued'` filter (REQ-THS-013), so hold is skipped by construction, not by
  convention.
- hold is an operator/lead act, drop-isomorphic. Lanes and machine leasers never set or clear it;
  the lead self-promotion ban is unchanged.

## §B Decisions

### B.1 Decision 1 — hold is a state, hold/unhold are operator verbs, and the text is never touched

`moai todo hold <id>` (queued → hold) and `moai todo unhold <id>` (hold → queued), both one locked
write through `Mutate`, both refusing with guidance on a wrong-state card, both leaving
id/text/added_at/spec_id/landing byte-identical. No `[HOLD]` marker exists — t1158's prose stays
prose and is NOT mechanically converted (§D Out of Scope); the operator who wants a prose-held card
held mechanically runs `hold <id>` on it.

### B.2 Decision 2 — the migration is a transactional table rebuild behind a schema_version bump to "2"

Three options were compared (full comparison: plan.md §F):

- **(A) Rebuild + version bump — CHOSEN.** Create `items_new` with the four-value CHECK, `INSERT
  SELECT` every row, verify full-field parity, swap, stamp `schema_version = "2"` — all inside the
  migration transaction, with the original preserved untouched on parity failure (the
  JSON→SQLite discipline of SPEC-TODO-SQLITE-001, folded here as REQ-THS-002/003).
- (B) Drop the CHECK, validate in application code only — REJECTED: the constrained live enum is
  deliberate ("keeps the live three-value enum the single constrained surface",
  `backlog_sqlite.go:104-106`); dropping the DB guard invites silent state corruption for a
  one-time cost saving.
- (C) Prose hold / marker convention — REJECTED: that is the defect under repair (§A.1).

**Premise correction recorded**: the card's "ALTER TABLE 0건" premise is false in this tree —
`ensureLandingColumn` (`backlog_sqlite.go:439-455`) is a guarded `ADD COLUMN` precedent. The
correction strengthens rather than weakens Decision 2: the package precedent is "additive preferred,
guarded exception documented at the code that performs it," and the rebuild lands as exactly such a
documented exception. `archived_items` carries no CHECK and is untouched by the rebuild.

### B.3 Decision 3 — compat is refuse-to-operate, by machinery that already exists

The compat matrix is carried by the version stamp, not by new code paths:

| Pair | Behavior |
|------|----------|
| NEW binary, v1 database | migration runs at open (REQ-THS-002); original preserved on failure (REQ-THS-003) |
| OLD binary, v2 database | open refused: `unsupported schema_version` + `ErrBacklogCorrupt` (measured at `backlog_sqlite.go:288-291, 393-396`; preserved as REQ-THS-005) — never operates on unknown bytes |
| Unknown state value, any reader | actionable surfaces exclude it by positive enumeration (REQ-THS-011); display surfaces render it literally (REQ-THS-016/018) — fail-visible, never fail-silent |

The old-binary refusal is a REAL cost for downgraded installs; it is the store's own standing
doctrine ("refuse-to-operate, never repair-by-delete", `backlog_sqlite.go:367-370`) and is cheaper
than every alternative reading of a live enum.

### B.4 Decision 4 — positive enumeration on every actionable surface (lead memo 1)

The core measured finding: `todo.go:903`'s `State != Queued` and `todo_autodone.go:283`'s
`!= Queued && != Picked` are NEGATIVE filters — each silently swallows not only `hold` but every
state added after it. The SPEC requires the inverse discipline: every selection predicate that
picks cards for actionable work enumerates the states it accepts positively (REQ-THS-011/012), and
the run-phase sweep re-derives the full inventory at run-phase HEAD rather than trusting the
plan-time list in `.moai/reports/t1308/predicate-sweep.md` §5.

### B.5 Decision 5 — display surfaces are truthful, not converted

`list` renders a held card with its literal state (default view discloses; `--json` carries
`"state":"hold"` for machine filtering), the statusline counts held cards in neither picked nor
queued (the positive aggregates already give this by construction — pinned by AC), and the web
console's state passthrough renders `hold` unchanged. docs-site is out of scope (§D).

## §C Requirements

> GEARS notation; one pattern per requirement. REQ-THS-011 and REQ-THS-012 are deliberately a
> pattern PAIR (positive obligation + prohibited shape) so the drift/mutation AC can fail on either
> half independently.

### C.1 Storage and migration

- **REQ-THS-001** (Ubiquitous): The backlog store shall admit exactly four live card states —
  `queued`, `picked`, `dropped`, `hold` — and the `items.state` CHECK constraint shall enumerate
  all four literally.
- **REQ-THS-002** (Event-driven): When the engine opens a database whose stamped `schema_version`
  predates the current stamp, the engine shall rebuild the `items` table to the four-state CHECK
  inside one transaction, verifying full-field row parity (row count plus every column tuple) before
  switching to the rebuilt table.
- **REQ-THS-003** (Event-detected): When full-field parity verification fails before the switch,
  the engine shall abort the migration leaving the original database file untouched, and shall not
  delete or overwrite either the database or any legacy artifact.
- **REQ-THS-004** (Event-driven): When the rebuild transaction commits, the engine shall stamp the
  new `schema_version` value so a binary carrying the previous stamp refuses the database at open.
- **REQ-THS-005** (Event-detected): When a binary opens a store stamped with a `schema_version` it
  does not recognize, the engine shall refuse the open with `ErrBacklogCorrupt` semantics —
  refuse-to-operate, never repair-by-delete.

### C.2 Operator verbs

- **REQ-THS-006** (Event-driven): When an operator runs `moai todo hold <id>` addressing a queued
  card, the CLI shall set that card's state to `hold` as one locked write, preserving id, text,
  added_at, spec_id, and landing unchanged, and print one confirmation line carrying the id and
  text prefix.
- **REQ-THS-007** (Event-detected): When a hold is attempted on a card whose state is not
  `queued`, the CLI shall refuse with guidance naming the card's current state and the recovery
  verb, writing nothing.
- **REQ-THS-008** (Event-driven): When an operator runs `moai todo unhold <id>` addressing a held
  card, the CLI shall return that card to `queued` as one locked write, leaving the card text
  byte-identical.
- **REQ-THS-009** (Event-detected): When an unhold is attempted on a card whose state is not
  `hold`, the CLI shall refuse, writing nothing.
- **REQ-THS-010** (Unwanted): A lane session, a machine leaser, or any other non-operator actor
  shall not be able to set or clear the hold state through any todo verb or lease path.

### C.3 Selection predicates

- **REQ-THS-011** (Ubiquitous): Every card-selection predicate on an actionable-work surface —
  `next` (bare and `<n>`), the machine-lease claim paths, `--auto` admission, and the auto-done
  candidate scan — shall select exclusively by positive enumeration of the states that surface
  accepts.
- **REQ-THS-012** (Unwanted): A selection predicate on an actionable-work surface shall not
  exclude states by a negated comparison whose default would silently include a state added later.
- **REQ-THS-013** (Where machine lease): Where a machine leaser selects a card (factory next, the
  Codex `-f` lane claim), the leaser shall select only `state == 'queued'` cards, so a held card is
  skipped by construction.
- **REQ-THS-014** (Event-detected): When a pick (`todo next <n>`) is attempted on a held card, the
  CLI shall refuse with unhold guidance, writing nothing.
- **REQ-THS-015** (State-driven): While a card is held, the auto-done candidate scan shall produce
  no outcome for it.

### C.4 Display surfaces

- **REQ-THS-016** (State-driven): While a card is held, `todo list` shall render it with its
  literal `hold` state in the text view, and the `--json` records shall carry `"state":"hold"`.
- **REQ-THS-017** (Unwanted): The statusline backlog render shall not count a held card in either
  the picked count or the queued count.
- **REQ-THS-018** (Ubiquitous): The web console queue view shall render a held card's state as
  `hold`.

### C.5 Documentation

- **REQ-THS-019** (Ubiquitous): The todo/GTD documentation surfaces (`.claude/skills/moai/workflows/gtd.md`,
  the `.claude/commands/moai/todo.md` stub, and their template mirrors) shall document the `hold`
  state and both verbs, with local and template copies carrying identical content.

## §D Out of Scope

### Out of Scope — docs-site

- docs-site (`adk.mo.ai.kr`) pages are a separate concern with its own 4-locale sync chain; a
  separate card owns them. Only the in-repo skill/command surfaces (REQ-THS-019) are in scope.

### Out of Scope — run-phase implementation

- This dispatch author SPEC artifacts only. No Go code, no template edits, no migration code is
  written at plan phase; milestones M1-M5 in plan.md are run-phase work.

### Out of Scope — prose conversion

- Existing prose-prefixed holds (t1158 「[보류 — …]」 cards) are NOT mechanically converted to the
  hold state; their text stays exactly as written. `hold` is a state, not a text transformation,
  and no scan converts card bodies.

### Out of Scope — hold metadata

- `hold` carries no reason field, no timestamp column, and no marker text. An operator who wants a
  reason keeps it in the card text (as today); the state records only the fact of holding.

## §G Gaps and Residual Risks

- **G1 (gap)**: the selection-predicate inventory (`.moai/reports/t1308/predicate-sweep.md` §5) is
  measured at plan-phase HEAD `8a969dfc0`; it is a starting inventory. The exhaustive enumeration is
  a run-phase M3 obligation, re-run at run-phase HEAD — t1240/t1294 may have landed new lease
  surfaces in the interim (their re-read is a plan.md §C precondition).
- **G2 (residual)**: REQ-THS-010 (actor boundary) is enforced by construction — no lease path gains
  a hold/unhold verb — but the CLI cannot distinguish WHO invoked it. The binary AC tests the
  surface set (no verb/flag exists on lease paths), not actor identity.
- **G3 (residual)**: an old binary refused at open (B.3) cannot even READ the queue to display it.
  Downgrade guidance belongs to release notes; this SPEC pins the refusal, not the doc surface for
  it.
- **G4 (residual)**: `graph.go:71`'s `Queued || Picked` "live set" is a display-adjacent predicate;
  whether the dependency graph should treat held cards as live is decided in run-phase M3 with the
  sweep (default: exclude, matching every other actionable surface).
