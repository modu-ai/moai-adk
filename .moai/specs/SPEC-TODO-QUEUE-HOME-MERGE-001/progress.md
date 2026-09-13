# progress.md — SPEC-TODO-QUEUE-HOME-MERGE-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-13
plan_version: 0.2.0 (plan-audit delta D1-D8 applied)
artifacts: spec.md, plan.md, acceptance.md, progress.md (Tier M, 4 files)
baseline: worktree .claude/worktrees/t657-queue-merge, branch WT-todo-queue-merge, base origin/develop 5e0f71175
notes: destructive steps (M4/M5) gated behind lead window + operator approval via lead + freshness re-check; gates are verdict-evidenced preconditions per acceptance.md §AC-TQM-008.
accepted_debt: 18 REQs vs Tier M ceiling 16 — knowingly over budget, accepted on coordinator authority (dangerous-operation granularity defensible; SPEC NOT split).

## §E.2 Run-phase Evidence

### M1 — merge-core (pure, tested) — 2026-09-13

- Files: `internal/kanban/todo_queue_merge.go` (merge core: collision taxonomy, token-boundary rewrite, high-water, identity-collision resolution), `internal/kanban/todo_queue_merge_test.go` (11 tests).
- RED evidence (verbatim, pre-implementation): `go test ./internal/kanban/ -run 'TestMergeBacklogRecords' -count=1` → `undefined: MergeBacklogRecords` / `undefined: MergeOptions` … `FAIL github.com/modu-ai/moai-adk/internal/kanban [build failed]` — tests ran before any implementation existed.
- GREEN: same selector → 11/11 PASS (`ok github.com/modu-ai/moai-adk/internal/kanban`); full package suite `go test ./internal/kanban/ -count=1` → `ok … 170.004s`; `go vet` + `go build ./...` clean.
- Identity-UUID collision (acceptance §D.5): fresh UUIDv7 + reconciliation row, RED-first observed. Taxonomy: identical-content shared number → resolved-duplicate; differing → home wins + project renumbered above `max(last_seq, max id)` across BOTH stores including archived; token-boundary rewrite verified against the t642/t6420 hazard.
- KNOWN GAP (recorded for the M4 design decision): `writeRecordArchive` (backlog_migrate.go) persists items/findings/archive/meta/identities but NOT the `todo_runtime_*` tables — the whole-record write drops a merged `Runtime` field on the floor. The merge core rewrites runtime assignments in memory per REQ-TQM-005; persisting them through the ONE-Mutate path needs either a runtime write extension in the transaction or a separate decision. Flagged to the lead; NOT silently extended in M1.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
