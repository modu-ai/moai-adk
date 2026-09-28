# SPEC-TODO-HOLD-STATE-001 — Implementation Plan

Tier M · card t1308 · plan-phase authored 2026-09-29 at HEAD `8a969dfc0` (worktree
`.moai/worktrees/t1308`, branch `WT-todo-hold-state`). Milestones are ordered by
decision-reversibility: the data-model change (M1) is the least reversible and leads; mechanical
doc edits (M5) close. Priority labels only — no time estimates.

## §A Context

The queue carries three states (`backlog_store.go:59-66`) and a CHECK constraint pinning them
(`backlog_sqlite.go:118`). Holds live in prose today, so every machine selector — including the
auto-done scan (`todo_autodone.go:283`, NEGATIVE filter) and, after t1240/t1294, the factory/Codex
machine leasers — treats a held card as pickable. This SPEC adds a fourth state `hold` with
operator-only verbs, rebuilds the `items` CHECK behind a `schema_version` bump, and converts every
actionable selection predicate to positive state enumeration.

## §B Known facts (measured in this tree; full detail + sweep table: `.moai/reports/t1308/predicate-sweep.md`)

1. **Schema**: `items.state TEXT NOT NULL CHECK (state IN ('queued','picked','dropped'))`
   (`backlog_sqlite.go:118`); `archived_items` carries NO state CHECK by design
   (`backlog_sqlite.go:104-106`) — the rebuild touches `items` only.
2. **Rebuild is pre-acknowledged**: the DDL comment (`backlog_sqlite.go:100-102`) states a fourth
   state "would need a table rebuild on every operator queue in the field."
3. **Version machinery exists**: `backlogSchemaVersion = "1"` (`backlog_sqlite.go:47-50`); unknown
   stamped versions refuse at open with `ErrBacklogCorrupt` semantics (`backlog_sqlite.go:288-291`
   reader path, `393-396` writer path). The old-binary compat half needs NO new mechanism.
4. **Premise correction (card falsified)**: "internal/kanban 이력상 ALTER TABLE 0건" is false —
   `ensureLandingColumn` (`backlog_sqlite.go:439-455`, t359, `3bcb0c33a`) is one guarded
   `ALTER TABLE ... ADD COLUMN`. Does not change the decision: ADD COLUMN cannot widen a CHECK.
5. **Known predicate gaps in the CURRENT tree** (both fixed by M3, both RED-now today):
   a held card is (i) PICKABLE via `next <n>` (`todo.go:920` refuses only `Dropped`) and (ii) an
   AUTO-DONE candidate (`todo_autodone.go:283`, `!= Queued && != Picked`).
6. **Actor precedent**: drop/undrop define the operator-verb shape — queued-only admission with
   recovery guidance (`todo_drop.go:111-120`), state-is-authority reversal with untouched text
   (`todo_drop.go:141-148`), byte-identical-on-refusal via `Mutate`.
7. **Doc surfaces**: `.claude/skills/moai/workflows/gtd.md` (canonical; verb table lines 48-64,
   JSON example line 214), `.claude/commands/moai/todo.md` stub, plus both template mirrors.
8. **t1306** (`--auto`) is QUEUED-ONLY — hold is auto-skip compatible, no t1306-side change.

## §C Pre-flight (run-phase entry conditions)

- **P1 [HARD, lead-fixed]**: re-read t1240 (factory next) and t1294 (Codex `-f` lane claim) card
  and branch state immediately before run entry. If either landed new lease-selection code, M3's
  sweep inventory absorbs it before any predicate edit.
- **P2**: re-run the selection-predicate sweep at run-phase HEAD (dispatch S1-S13 list is a
  plan-time starting inventory, not the contract).
- **P3**: confirm `.moai/reports/todo-logic-review-20260929.md` §P3 premise (prose holds remain
  pick candidates) still holds on the live queue shape.

## §D Constraints

- C-1: ALTER TABLE count stays at the documented-exception level — the ONLY new DDL beyond
  `backlogDDL` is the rebuild inside the migration transaction; no schema change outside
  `items` (+ the `meta.schema_version` stamp). `archived_items` is untouched.
- C-2: No new read path may open with a stamp it does not recognize (preserve the
  refuse-to-operate contract, REQ-THS-005).
- C-3: Every verb mutation goes through `store.Mutate` (byte-identical on refusal) — no direct
  record writes.
- C-4: Template-First — any doc surface edit lands in `internal/template/templates/**` in the
  same change as the local copy.
- C-5: No docs-site edits. No prose conversion of existing cards. No `hold` metadata (no reason
  column, no marker text).

## §E Self-verification scope (lane-local; NO full-suite runs)

```bash
go test ./internal/kanban/...
go test ./internal/cli -run '^(TestTodo|TestBacklog)'
go test ./internal/statusline -run '^(TestBacklog|TestLanded)'
go test ./internal/web -run '^TestTodo'
go vet ./internal/kanban/... ./internal/cli/...
```

Storage roundtrip tests run under `-race` (they own a goroutine-free but transaction-heavy path;
cheap to assert once). Full-suite judgment is CI's (origin/develop).

> **On the `-run` patterns**: the linter (`VacuousTestAssertion`) suggests fully anchored patterns
> (`^(TestTodo|TestBacklog)$`). That anchor is deliberately NOT applied: `TestTodo*` /
> `TestBacklog*` / `TestLanded*` are name FAMILIES (`TestTodoNext`, `TestBacklogCounts`, …), and an
> exact anchor would select zero tests — the genuinely vacuous run. The prefix forms are the
> precise scope here; the lint warnings are accepted as known-intentional (3 warnings, 0 errors).

## §F Migration decision — rebuild vs convention (the comparison plan.md owns)

| Axis | (A) Rebuild + stamp "2" — CHOSEN | (B) Drop CHECK | (C) Prose hold / marker |
|------|----------------------------------|----------------|-------------------------|
| DB-level enum guard | kept, widened to 4 values | LOST (the "single constrained surface" doctrine, `backlog_sqlite.go:104-106`, is abandoned) | never existed for hold |
| Migration mechanics | one transaction: create `items_new` (4-value CHECK) → `INSERT SELECT` → parity check → drop+rename → stamp "2" | one rebuild (same mechanics, worse endpoint) | none |
| Old binary vs new db | refuses at open (mechanism already measured, `backlog_sqlite.go:288-291`) | opens, misreads silently | opens, ignores text |
| New binary vs old db | migrates at open, idempotent via stamp | migrates | nothing to migrate |
| Cost | one-time rebuild per field queue (small tables; queue sizes are tens of rows) | one-time rebuild, permanent guard loss | zero migration, permanent defect |
| Fits `schema_version` convention | yes — physical layout changed, stamp bumps; exactly what the stamp exists for | yes but wasted | no stamp needed, no state either |

**Decision: (A).** The stamp bump is what makes (A) safe rather than heroic: an old binary never
sees a `hold` row because it refuses the database before reading `items` (REQ-THS-005, already
implemented and AC-pinned). The rebuild inherits the JSON→SQLite migration discipline
(SPEC-TODO-SQLITE-001) as binding requirements, not prose: full-field parity before the switch
(REQ-THS-002), original preserved untouched on failure (REQ-THS-003), stamp inside the committing
transaction (REQ-THS-004, mirroring `backlog_migrate.go:403`).

**Unknown-enum read path (required plan item)**: `BacklogState` is a string type, so a reader can
always observe a value outside the current enum (a hand-edited db, a future state). The binding
rule: actionable surfaces exclude unenumerated states by construction (positive enumeration,
REQ-THS-011/012); display surfaces render the literal value (REQ-THS-016/018). Fail-visible, never
fail-silent.

## §G Milestones

- **M1 (High) — storage**: `BacklogStateHold` constant; rebuild migration (4-value CHECK,
  transactional, parity-verified, stamp "2"); roundtrip + parity-failure tests; compat tests for
  the refuse path. *Data-model first — everything downstream keys on this.*
- **M2 (High) — verbs**: `todo hold` / `todo unhold` with drop-isomorphic refusal semantics;
  pick-refusal on held cards (`todo.go:920` gap); `--expect` flag parity with drop/undrop.
- **M3 (High) — predicate sweep**: convert every actionable-surface predicate to positive
  enumeration at run-phase HEAD (re-swept inventory); autodone candidate fix; machine-lease
  queued-only pin (factory next, `-f` lane claim); drift guard test that fails when a new state
  lands without the sweep re-run (mutation-testable).
- **M4 (Medium) — display truthfulness**: `list` default + `--json` carry `hold`; statusline
  counts exclude held (pin by test); web console passthrough verified.
- **M5 (Low) — docs**: `gtd.md` verb table + JSON example, `commands/moai/todo.md` stub, both
  template mirrors, byte-parity check.

## §H Anti-patterns (named to prevent recurrence)

- Do NOT convert held cards' text or add a `[HOLD]` marker — the state is the whole feature.
- Do NOT let M3 trust the plan-time sweep table (§B.5 / reports §5) — it is a starting inventory;
  run-phase HEAD is the contract.
- Do NOT "temporarily" drop the CHECK to ship faster — (B) was evaluated and rejected (§F).
- Do NOT add hold/unhold flags to any lease path (REQ-THS-010's enforcement is surface absence).

## §I Cross-references

- spec.md §B.2/§B.3 — decision record and compat matrix.
- acceptance.md — AC-THS-001..019 with RED-now/green-path table.
- `.moai/reports/t1308/predicate-sweep.md` — measured basis + sweep inventory.
- SPEC-TODO-SQLITE-001 — the migration discipline REQ-THS-002/003 inherit.
- SPEC-TODO-DESTRUCTIVE-GUARD-001 — archive storage this migration must not disturb (AC-THS-002
  asserts archive rows survive the rebuild).
