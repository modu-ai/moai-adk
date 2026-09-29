# SPEC-TODO-CLASSIFY-DISPATCH-001 — Progress

Tier M · card t1332 · plan-phase artifact set authored 2026-09-29 at HEAD `145c3d98c`
(worktree `.moai/worktrees/t1332`, branch `WT-card-autodispatch`). Status: `in-progress`
(run phase entered 2026-09-29, cycle_type=tdd; run-phase basis HEAD `51f3e9878` = plan
commits `875a4b33b` + clean develop absorb `8fc5a7407` carrying the t1240 surface —
REQ-TCD-013 entry precondition verified, see §E.2).

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifact set complete: `spec.md` (14 REQ, GEARS), `plan.md` (§A-§H, 5 milestones,
  decision records OD-1..3 — OD-1/OD-3 folded as leader rulings 2026-09-29, OD-2 open — plus one
  flagged lead follow-up on the failure-default vs read-default tension), `acceptance.md`
  (AC-TCD-001..014, Given-When-Then, RED-now/green adoption table), this `progress.md`.
- Revision 0.2.0 folded the leader rulings of 2026-09-29: OD-1 serial-card mutual exclusivity
  (pipeline exclusivity rejected), OD-3 serial failure default; provenance recorded in plan.md
  §C (noul OD-1 0.31 / OD-3 0.36). D1 citation-prefix repair applied. Post-repair lint:
  0 findings.
- Revision 0.3.0 resolved the flagged tension by leader ruling (2026-09-29, OD-3 extension):
  REQ-TCD-014's absent-field READ mode default flips parallelizable → serial (priority/blocked
  defaults untouched); plan.md §C records the resolution. No read-default AC added; no existing
  AC asserted the old default. Post-revision lint: 0 findings; REQ/AC 14/14 unchanged.
- Measured surface basis exported: `.moai/reports/t1332/surface-notes.md` (all card premises
  verified; measured corrections recorded — t1240 branch 28 commits ahead of develop, tip
  `d43e50bb3`, `9866ca25e` an ancestor).
- SPEC ID regex pre-write check: PASS (`SPEC-TODO-CLASSIFY-DISPATCH-001`); ID unique in
  `.moai/specs/`.
- Frontmatter validated against the canonical 12-field schema SSOT; `status: draft` set at
  creation per plan-phase ownership.
- Out of Scope section carries five `### Out of Scope — <topic>` H3 sub-headings with `-` bullets
  (t1240 / t1306 / t1261 / Jev / scheduling-intelligence boundaries).
- Known bounded gaps: none. The absorption-order precondition (REQ-TCD-013) is a run-phase entry
  gate by design, not a plan-phase gap.

## §E.2 Run-phase Evidence

Run phase entered 2026-09-29, cycle_type=tdd, lane worker-62 (card t1332).

- **Entry precondition (REQ-TCD-014 / AC-TCD-014)**: verified at entry — this branch's HEAD
  `51f3e9878` is the plan commits plus the clean absorb of develop `8fc5a7407`, which carries the
  t1240 surface (`newFactoryNextCommand` :459 / `newFactoryStageCommand` :547 /
  `newFactoryCompleteCommand` :640 in `internal/cli/factory_card.go`). AC-TCD-014 is now a pinned
  regression test: `internal/cli/factory_entry_precondition_test.go`
  (`TestFactorySelfDispatchSurfacePresent`, `TestAbsorptionPreconditionDocumented`) — both GREEN.
- **M1 — classification data model (GREEN)**: `internal/kanban/classification.go` (closed value
  sets, validation, JSON parse, read-default derivation, SQLite value codec);
  `BacklogItem.Classification *CardClassification json:"classification,omitempty"`
  (backlog_store.go); `classification` TEXT column appended last on items + archived_items via the
  pragma_table_info-gated ADD COLUMN path (backlog_sqlite.go); read/write wiring in
  backlog_migrate.go following the landing/columnExpr contract. RED evidence captured before GREEN
  (compile-failure output on the missing API, recorded in the M1 commit message trail); GREEN:
  - `go test ./internal/kanban/ -run 'TestCardClassification|TestParseCardClassification|TestEffectiveCardClassification|TestBacklogClassification|TestSchemaFreezeCarriesClassification' -count=1` → `ok github.com/modu-ai/moai-adk/internal/kanban 0.418s` (this run, this tree)
  - `go test -timeout 30m ./internal/kanban/ -count=1` → 2 failures BEFORE the freeze-tuple update
    (`TestSchemaFreezeRecordsTransitionStamps`, `TestTransitionStampColumns_FreshUpgradedConverge`
    — both stale pins, updated in the same milestone); full suite re-run pending after M2/M3 land.
  - AC-TCD-013 (jev boundary + positive control): `go test ./internal/cli/ -run
    'TestProductPathsCarryNoJevReference|TestJevBoundaryScanPositiveControl' -count=1` →
    `ok ... 0.779s`. The positive control is a planted fixture (scripts/jev is untracked, so the
    control must be clone-independent). Product-path hits: 0; `_test.go` files excluded per the
    B3 boundary-grep precedent.
- Files touched (M1): internal/kanban/{classification.go,classification_test.go,backlog_store.go,
  backlog_sqlite.go,backlog_migrate.go,backlog_schema_freeze_test.go,backlog_transition_stamps_test.go},
  internal/cli/{product_jev_boundary_test.go,factory_entry_precondition_test.go}.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
