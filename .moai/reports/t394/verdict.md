# Verdict — card t394 · SPEC-TODO-ARCHIVE-QUERY-001

Lane: lane-6 (factory) · Branch: `WT-todo-done-history` · Base: `origin/develop@2c18091d1` (absorbed to `e8ae9798a` at run entry) · Status: **integration-ready**

## What the card became

The dispatch's premise — "`moai todo done` deletes records, so closure history is lost" — was re-measured and **falsified** during plan entry: the archive has existed since SPEC-TODO-DESTRUCTIVE-GUARD-001, and the binary installed 2026-09-01 06:37 KST (develop build `v3.1.2-1033-g64bba61aa`) archives on `done` (`archived_items` observed with live rows). What survived the re-measurement is the gap the SPEC delivers: **the archive had no reader** — of the sixteen verbs at the time, only `export-json` (a whole-record downgrade dump) touched it. Both founding incidents were reclassified on measurement: the harness-selection card was never lost (t393, live at the 16:20 KST snapshot; closed and archived later that day — now itself an instance of the gap), while t81/t83/t88/t89 are pre-archive-era destruction, unrecoverable by any surface (~289 cards, a frozen pre-archive-era estimate that survived two independent re-measures).

## Gate trajectory

| Stage | Verdict |
|---|---|
| Plan phase (iter-1 → iter-2) | FAIL 0.775 → PASS-WITH-DEBT 0.875 (Tier M) |
| Correction round 3 (lead-mandated premise repair, v0.2.2) | 4 items landed, orchestrator-verified 7/7 |
| Run-entry Phase 1 gate | **PASS 0.91** (`plan-audit-run-gate-2026-09-01.md`) |
| Run phase M0–M5 | 15/15 ACs PASS; orchestrator re-verification 7/7 |
| Sync phase (3-phase close) | sync `973832f94` + backfill `ffab66fba`; B12 self-tests clean |
| Sync audit | **PASS, integration-ready, weighted 0.95** (`sync-audit.md`) — 4 mutation guards confirmed firing |

## What shipped

`moai todo history` — read-only reader over the archive (`internal/cli/todo_history.go`, `newTodoHistoryCmd` at `internal/cli/todo.go:152`): four fates in one call (live `queued`/`picked`/`dropped`, `archived` with state-at-archive, `absent` at exit 0), newest-first listing with `--limit` (default 20, `0` unbounded, negative refused), and three stderr-only honesty disclosures: the absent-qualifier keyed on the durable `meta.last_seq` (fires after the first `done`, unlike an emptiness test), the withheld count on truncation, and the degraded-store naming via `internal/kanban/backlog_archive_vouch.go` (schema probed on its own connection before any engine open, so `IF NOT EXISTS` DDL cannot mask a missing archive). Invariant guards: six live-reader goldens captured pre-verb at C=`e7eec6122` with clause-1 provenance/integrity assertions (branch-scoped), storage byte-identity (SHA-256 over the state dir), schema freeze, never-prompts (symbol-located regex). Documented on both surfaces (mirror neutral). C is a proper ancestor of HEAD; clause 2 regeneration from C compared byte-identical.

## Incidents

1. **Primary-queue pollution (repaired, reported)**: the first M1 RED run executed `runTodo` without `todoFixture(t)`; queue-root resolution walked worktree→primary and seeded fixture cards t414–t420 onto the operator's live queue (`last_seq` 413→420). Repaired via the public CLI (`done` → archived, individually restorable via `undone`); no real-card mutation (all mutation verbs in that run were refused). Lead ledgered the consumed ids; a fail-loud guard proposal is with the operator.
2. **429 interruption**: the run-phase agent hit the account's 5-hour usage limit mid-E3; state was pinned from disk (7 commits) and the agent resumed from task 8, completing the §E report at the final tree.

## Known debts (non-blocking)

- 3 MINOR optional defects from sync-audit: F1 degraded-DB DDL regeneration (literal limit of REQ-TAQ-010), F2 `--limit` ignored when combined with an id argument, F3 error double-print (matches the existing `todo_why.go` pattern).
- 1 spec-lint WARNING: `MovingRefUnpinned` at `plan.md:94` (body-ownership boundary; report-only).
- Post-integration residual: an unrelated default-read change landing on develop after this card will turn clause 3 RED by design (acceptance.md §D.3) — regenerate goldens only after reading the golden↔develop diff.
- CI verdict on `origin/develop` is the lead's read (HARD rule 2026-09-01); this lane performed no pushes of the WT branch and no direct CI requests.

## Evidence index

- Premise re-verification: `premise-check.md` · audit responses: `audit-response-iter1.md`, `audit-response-iter2.md` · run-entry gate: `plan-audit-run-gate-2026-09-01.md` · sync audit: `sync-audit.md` · SPEC record: `.moai/specs/SPEC-TODO-ARCHIVE-QUERY-001/` (progress.md §E.1–§E.4)
