---
id: SPEC-TODO-TRANSITION-STAMPS-001
title: "A queue with a time axis: transition stamps (picked_at, dropped_at) on items, an archive-time stamp and a persisted done-verdict landing record on archived_items, exposed through history"
version: "0.1.0"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec (card t1310)
priority: P3
phase: "v3.2.0 target"
module: "internal/kanban, internal/cli"
lifecycle: spec-anchored
tags: "kanban, backlog-queue, sqlite, add-column, transition-timestamps, landing-verdict, history"
tier: M
depends_on:
  - SPEC-TODO-LANDING-EVIDENCE-001
related_specs:
  - SPEC-TODO-LANDING-ATTRIBUTION-001
  - SPEC-TODO-LANDING-STATE-001
  - SPEC-TODO-SQLITE-001
  - SPEC-TODO-ARCHIVE-QUERY-001
---

# SPEC: Transition stamps and a persisted done-verdict landing record

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-29 | Initial plan-phase authoring (card t1310). Schema ground truth measured on the live queue db `/Users/goos/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db` by the dispatching lane, 2026-09-29; code paths re-verified in this tree (`9cc3fdc4d`). |

## 1. Overview

The backlog queue records exactly one moment in a card's life: `added_at`. A card
can sit `picked` for days with no record of when the pick happened, and a card
closed with `todo done --require-landed` prints its landing verdict to stdout
and then throws it away — the archived row remembers nothing. This SPEC adds a
time axis to the queue: when a card was picked, when it was dropped, when it
entered the archive, and what the done-time landing query actually answered.

## 2. Background — measured facts

All schema facts below were measured by the dispatching lane on 2026-09-29
against the live db, then re-verified against this tree's DDL
(`internal/kanban/backlog_sqlite.go:107-150`):

- `items` carries `seq, id, text, added_at NOT NULL, spec_id, state CHECK
  (queued|picked|dropped), landing`. `picked_at` / `dropped_at` exist nowhere.
  `added_at` is the only timestamp. The card's premise is CONFIRMED.
- `archived_items` carries `seq, id, text, added_at NOT NULL, spec_id, state
  (no CHECK), position, landing`. `archived_items.state` holds the
  pre-archival state — sample rows archived while still `picked` are the
  landed-but-picked backlog t472 axis E named. No archive timestamp exists.
- Landing stats on archived_items: 915 rows, 805 NULL, 110 populated, 0 stored
  literal `"-"`. **Precision the card body flattened**: `landing=-` is the
  `history` rendering of a NULL cell (`todoHistoryLandingCell`,
  `internal/cli/todo_history.go:108-119`), not a stored value. The gap is real
  either way — 88% of archived rows carry no landing evidence.
- The `done` verb runs `--require-landed`'s query, prints
  `done <id> landing=<verdict> [ref=<ref>]` (`internal/cli/todo.go:750-757`),
  and persists NOTHING: `ArchiveCard` copies the item as-is
  (`internal/kanban/backlog_store.go:255-289`). A verdict that answered exists
  only in scrollback.
- The additive-migration convention is visible in the schema itself: the
  trailing `landing TEXT` columns on both tables were added by
  `ensureLandingColumn` (`internal/kanban/backlog_sqlite.go:421-455`) —
  idempotent `ALTER TABLE ADD COLUMN` gated on `pragma_table_info`, no table
  rebuild, no destructive ALTER in the package's history.
- The five original `BacklogItem` fields are the frozen per-item contract
  (REQ-TODO-013); `Landing` and `CardUUID` were added after it as additive,
  `omitempty` pointer fields. The transition stamps follow that precedent.
- Sibling card t1308 is reviewing the same migration convention for a
  **state-CHECK** change (a rebuild, since SQLite cannot ALTER a CHECK). This
  SPEC's changes are pure additive columns and do not overlap that scope.

## 3. Requirements (GEARS)

Requirement IDs use the domain token TST (transition stamps).

### 3.1 Schema

**REQ-TST-001** — The backlog store shall carry nullable `picked_at TEXT` and
`dropped_at TEXT` columns on `items`, added by an idempotent additive
migration, such that a database created fresh and one already in the field
converge on the same column set.

**REQ-TST-002** — The backlog store shall carry nullable `archived_at TEXT` on
`archived_items`, added by the same idempotent additive migration.

**REQ-TST-003** — The backlog store shall not rebuild either table, drop any
column, or alter any constraint while adding these columns; the frozen
five-field per-item contract (REQ-TODO-013) keeps its field names, types, and
JSON tags, and the stamps ride alongside it as `omitempty` additive fields the
way `Landing` and `CardUUID` already do.

### 3.2 Transition stamping

**REQ-TST-004** — **When** a card enters the picked state (via `todo add
--pick`, `todo next`, or any other transition that sets `state='picked'`), the
store shall stamp `picked_at` with the transition time, overwriting any stamp
left by a previous picked episode — the stamp answers "when did the CURRENT
picked episode begin", not "when was it ever picked".

**REQ-TST-005** — **When** a card leaves the picked state without being
archived (`unpick` back to queued), the store shall clear `picked_at` to NULL;
a queued card carries no picked stamp.

**REQ-TST-006** — **When** a card is dropped, the store shall stamp
`dropped_at`; **when** an `undrop` returns it to the queue, the store shall
clear `dropped_at`.

**REQ-TST-007** — **When** `done` archives a card, the archived row shall
preserve the item's `picked_at` / `dropped_at` stamps as they stood at archive
time and carry `archived_at` stamped with the archive time — the archive is
the row's final home and its stamps must be readable there without the live
row.

### 3.3 The done-time landing verdict record

**REQ-TST-008** — **When** `done` runs with `--require-landed` and the landing
query answers, the archived row shall persist the verdict — landed (with the
delivering SHA), not-landed, or unknown-as-answered — together with the
answering ref and the verdict time, so `history` can reproduce what the query
said without re-running it.

**REQ-TST-009** — **When** `done` runs WITHOUT `--require-landed`, the store
shall not fabricate a query verdict: the archived row's landing record is
whatever recorded evidence the item already carried (the t665 path, printed
today and persisted from now on) or NULL. No flag, no query, no invented
answer.

**REQ-TST-010** — The store shall not write a persisted verdict as attribution
evidence absent its answering ref. A stored verdict is a snapshot of what a
NAMED ref answered at a NAMED time; it never becomes an independent
attribution claim, and every write path funnels through the existing
`LandingEvidenceValue` discipline (typed NULL, never `{}` or `""`).

### 3.4 History exposure

**REQ-TST-011** — **When** `todo history` renders an archived row (lookup or
listing), it shall expose `picked_at`, `dropped_at`, `archived_at`, and the
landing record, with every absent value rendered `-` — consistent with the
existing `landing=-` cell convention — and the output shall remain
tab-separated and machine-parseable.

**REQ-TST-012** — **While** a card sits in the live queue, `todo history
<id>` shall expose its `picked_at` / `dropped_at` stamps for the live row, and
`todo list --json` shall carry them as `omitempty` fields so a WIP age is
computable from the queue surface itself (t472 axis E's detection need).

### 3.5 Axis-F consistency

**REQ-TST-013** — The persisted verdict record shall carry the answering ref
so the t472 axis-F attribution predicate (title-attribution position within
commit subjects, `internal/kanban/prlink_landed.go`) can be RE-RUN against
that ref later; the stored verdict supplements the discriminator and shall not
replace, bypass, or short-circuit it. **Conclusion recorded by this SPEC: the
new columns CONFIRM the axis-F discriminator rather than changing it** —
attribution is decided at query time by the subject-position predicate;
storage records the predicate's answer (and the ref that produced it), it does
not re-derive attribution. Re-adjudication always means re-running the
predicate against the recorded ref.

## 4. Out of Scope

### Out of Scope — the landing predicate itself

- The title-attribution-vs-body-mention discriminator (t472 axis F,
  SPEC-TODO-LANDING-ATTRIBUTION-001) is landed and frozen. This SPEC stores
  its answers; it does not touch the predicate, `LandedGrepArgs`, or the
  subject-form table.

### Out of Scope — hold-state CHECK semantics

- Sibling card t1308 owns the question of the `items.state` CHECK constraint
  (SQLite cannot ALTER a CHECK; any change there is a table rebuild). This
  SPEC adds nullable columns only and must not entangle with that rebuild.

### Out of Scope — retrofitting history

- The 805 existing NULL-landing archived rows stay as they are. Backfilling
  landing evidence for already-archived cards would be fabrication, not
  storage. The columns make FUTURE closures measurable; the past stays
  honestly empty.

### Out of Scope — new detection subcommands

- No new aging/staleness verb. t472 axis E explicitly rejected a detection
  subcommand; the stamps make detection computable by the surfaces that
  already exist (`history`, `list --json`, SQL over the queue db).

### Out of Scope — dropped_at semantics beyond the drop/undrop pair

- A dropped card that is later done'd directly (drop → done) carries both
  stamps; no re-definition of what "dropped then done" means is attempted.

## 5. Constraints

- Additive schema changes only, per the repo's zero-destructive-ALTER history
  (`CREATE TABLE IF NOT EXISTS` + `pragma_table_info`-gated `ADD COLUMN`).
- The existing `landing` column machinery (SPEC-TODO-LANDING-EVIDENCE-001) is
  the persistence precedent: typed NULL, single write path, `omitempty` JSON.
- Migration idempotence is decided by reading column metadata, never by
  catching SQLite's duplicate-column error text.
- All timestamps are TEXT in the store's existing format (`added_at`-style);
  no new clock abstraction is introduced.
- The `backlog_schema_freeze_test.go` freeze test exists precisely to force a
  deliberate, reviewed decision on schema change — the run phase updates it
  as the recorded decision, not around it.

## 6. Acceptance

Acceptance criteria live in `acceptance.md` (AC-TST-001 … AC-TST-012), each
mechanically verifiable against a `t.TempDir()` queue db: schema assertions
via `pragma_table_info`, CLI behavior assertions via `todo` verb runs, and
history-output assertions via tab-field greps.

## 7. References

- `internal/kanban/backlog_sqlite.go` — DDL, `ensureLandingColumn`,
  `pragma_table_info` migration pattern
- `internal/kanban/backlog_store.go` — `BacklogItem`, `ArchiveCard`,
  frozen-contract comment (REQ-TODO-013)
- `internal/cli/todo.go` — `done` verb, `--require-landed`, verdict line
- `internal/cli/todo_history.go` — `history` rendering, `landing=-` cell
- `internal/kanban/prlink_landed.go` — the axis-F attribution predicate
- `.moai/specs/SPEC-TODO-LANDING-ATTRIBUTION-001/spec.md` — the durable
  t472 axis record in this tree
- `.moai/specs/SPEC-TODO-LANDING-EVIDENCE-001/spec.md` — the landing column
  this SPEC's persistence follows
