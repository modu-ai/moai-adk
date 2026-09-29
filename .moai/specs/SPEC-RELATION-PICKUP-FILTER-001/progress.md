# SPEC-RELATION-PICKUP-FILTER-001 — Progress

Tier S (dispatcher-mandated acceptance layer added) · card t1343 · plan-phase
artifact set authored 2026-09-29 at HEAD `113082295` (worktree
`.moai/worktrees/t1343`, branch `WT-relation-pickup-filter`). Status: `draft`
(plan-phase creation per ownership; run phase not entered).

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifact set complete: `spec.md` (7 REQ, GEARS — Ubiquitous 3 /
  When 3 / Unwanted 1), `plan.md` (§A-§H, 4 milestones, cycle_type=tdd),
  `acceptance.md` (AC-RPF-001..007, Given-When-Then, RED-now/green 채택 셀,
  품질 게이트 + DoD), this `progress.md`. The acceptance layer is
  dispatcher-mandated (card t1343 gate: release-blocking criteria name command
  + expected output); Tier S scope unchanged (< 300 LOC, 2 production files).
- SPEC ID regex pre-write check: run as Bash, verbatim output `PASS`
  (`SPEC-RELATION-PICKUP-FILTER-001`); ID unique in `.moai/specs/` (no
  `SPEC-RELATION-*` / `SPEC-PICKUP-*` directory existed).
- Frontmatter validated against the canonical 12-field schema SSOT; `status:
  draft` set at creation; no snake_case aliases.
- Measured input verification (all in this tree, HEAD `113082295`):
  (b) record-only claim verified at `internal/kanban/backlog_store.go:139-148`
  (the card text's `:135-138` citation predates a comment growth — quoted in
  spec.md §A.2 F-1); (c) pickup-path finding verified —
  `autoPickTargets` (`internal/cli/todo_auto.go:142`), queued arm appends every
  queued card (`:162-166`), zero relation consumers outside definitions
  (non-test grep). Cycle-guard gap verified at `runTodoRelate`
  (`internal/cli/todo_relate.go:60-94`).
- Merge-vs-separate decision recorded: **t1343 stands alone** — t1338 has no
  landed SPEC (no `.moai/specs/` hit; report hits are coincidental test
  temp-dir substrings in t60 artifacts; the proposal names no card for P3) and
  its queued card text enumerates four factory-autonomy pieces that exclude
  the relation filter. Full rationale: spec.md §A.4.
- Out of Scope section carries four `### Out of Scope — <topic>` H3
  sub-headings with `-` bullets (factory-lease t1240 boundary / t1338
  absorption / display surfaces / relation vocabulary policy).
- Post-authoring lint: `go run ./cmd/moai spec lint SPEC-RELATION-PICKUP-FILTER-001`
  (tool built from this tree, HEAD `113082295`) → `✓ No findings — all SPEC
  documents are valid` (0 errors, 0 warnings). One authoring fix during
  bring-up: the first lint pass emitted 7 `CoverageIncomplete` warnings; the
  AC sections carried REQ references in prose but lacked the parser's
  `**Covers**: maps REQ-…` mapping clause — added to all seven ACs (matches
  the SPEC-TODO-CLASSIFY-DISPATCH-001 convention), lint re-run clean.
- Known bounded gaps: none at plan-phase. §G risks R-1..R-3 are design
  boundaries, not gaps.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
