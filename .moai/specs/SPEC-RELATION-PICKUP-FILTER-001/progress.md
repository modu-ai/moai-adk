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

Run entered 2026-09-30 on branch `WT-relation-pickup-filter` at plan baseline
HEAD `28e672b39`, tree clean. cycle_type=tdd, milestones M1-M4 in plan.md §F
order. RED-before-GREEN observed per milestone; verbatim outputs below.

### M1 — the pickup filter (REQ-RPF-001/002/003, AC-RPF-001..003)

**RED (observed on the pre-implementation tree, HEAD `28e672b39`, tests added
first):**

```
$ go test -timeout 30m -count=1 ./internal/cli/ -run 'TestAutoPickTargets'
--- FAIL: TestAutoPickTargetsRelationBlocked (0.30s)
    --- FAIL: TestAutoPickTargetsRelationBlocked/depends_blocks_its_subject (0.17s)
        todo_relation_filter_test.go:57: pickup = [t1,t2], want [t2] — t1 must be excluded while t2 stays a candidate
    --- FAIL: TestAutoPickTargetsRelationBlocked/blocks_blocks_its_related (0.14s)
        todo_relation_filter_test.go:57: pickup = [t1,t2], want [t1] — t2 must be excluded while t1 stays a candidate
--- FAIL: TestAutoPickTargetsReturnsAfterDone (0.14s)
    todo_relation_filter_test.go:89: blocked card t2 was a candidate before the predecessor's done
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.230s
```

RED for the right stated reason: both direction arms failed because the
blocked-side card was still a pickup candidate (no relation consumer in the
queued arm), and the done-resolution arm failed before any archive move. (A
test-side Given fix happened between this RED and GREEN: the done-resolution
test originally recorded `depends {t1,t2}` — whose blocked side is the
subject t1, not the successor t2 — and was corrected to `t2 depends t1`
before GREEN; the direction arms above were unchanged.)

**GREEN (same command, after implementation):**

```
$ go test -timeout 30m -count=1 ./internal/cli/ -run 'TestAutoPickTargets'
ok  	github.com/modu-ai/moai-adk/internal/cli	1.348s
```

Implementation: `kanban.WaitsOnOf` (direction normalization SSOT — `depends`
{S,R} → S waits on R; `blocks` {S,R} → R waits on S) + `BacklogRecord.FindingsBlocking`
(finding-existence predicate per spec.md B.1) in
`internal/kanban/backlog_store.go`; the queued arm of `autoPickTargets`
(`internal/cli/todo_auto.go`) excludes cards with a non-empty blocking set.
Dead-owner rescue arm untouched (REQ-RPF-006).

**Regression check (M1):**

```
$ go test -timeout 30m -count=1 ./internal/cli/ -run 'TestTodo|TestAuto|TestRelate|TestSemantic|TestMachine|TestFinding'
ok  	github.com/modu-ai/moai-adk/internal/cli	460.801s
$ go vet ./internal/cli/... ./internal/kanban/...
(no output) exit 0
```

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
