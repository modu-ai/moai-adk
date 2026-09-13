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

### M2 — backup/restore + M4/M5 fixture rehearsals — 2026-09-13

- Files: `internal/kanban/todo_merge_backup.go`, `internal/kanban/todo_merge_procedure.go` (+ 2 test files), `cmd/t657-merge/main.go` (one-off shell, plan §H shape — no CLI verb).
- RED-first evidence: (a) identity-UUID-collision test → `undefined: MergeBacklogRecords … FAIL [build failed]` before any implementation; (b) ordering-evidence comparator → `undefined: SnapshotStoreDirState/StoreDirMutations … FAIL [build failed]`; corrupted-fixture probes (added/removed/mtime-moved artifact) each observed failing before the comparator existed and passing after.
- GREEN: backup hash-verified (db + WAL/SHM + legacy json + .migrated; absent artifacts recorded absent); tampered-copy test proves the hash gate aborts (REQ-TQM-002); restore round-trip byte-identical (AC-TQM-007); full M4/M5 rehearsal on t.TempDir() fixtures — dry-run byte identity, ONE-Mutate apply (seam-counted = 1, AC-TQM-009), zero-loss + mapping + reference-rewrite verification, rollback restore, M5 retirement (fence marker + rename, never delete) and un-retire.
- Package suite `go test ./internal/kanban/ -count=1` → `ok 166.487s`; `go vet` + `go build ./...` clean.

### M3 — freshness re-derivation (READ-ONLY against real stores) — 2026-09-13

Commands: `/tmp/t657-merge -observe …` and `-dry-run` (backup copies to /tmp; stores read via LoadPure only). Post-dry-run hash check: project `backlog.db` sha256 prefix `9802d3f6d1c4e7c6` and home `backlog.db` prefix `d0b7da4cb8c9d98b` both EQUAL their dry-run-time backup records — the dry-run mutated nothing.

| Store | Live (q/p/d) | Archived | last_seq | db mtime |
|---|---|---|---|---|
| home `~/.moai/db/moai-adk-go-1bd3d038/todo` | 59 (6/24/29) | 349 | 697 | 2026-09-13 20:11 — MOVED during observation (19:43→20:11, other writers) |
| project `<primary>/.moai/state/todo` | 105 (72/6/27) | 267 | 661 | 2026-09-11 03:20:10 (stable) |

**Freshness vs plan baseline (REQ-TQM-014): MOVED — M4 MUST re-plan.** Plan-time "105 project / 95 home" was LIVE-only; the fresh (live+archived) derivation shows the real collision taxonomy is: **290 duplicates / 82 renumbered / 0 pure migrations, union 780 cards, high-water 779** — home's ARCHIVED population already occupies nearly the whole project id-space, so project-only LIVE ids (e.g. t658-t661) collide with home ARCHIVED copies of the same completed work. The merge core handles archived collisions (tested), and the dry-run verified zero-loss 780/780 with 82 mapping rows — but the taxonomy decision (a project live card whose number matches a home ARCHIVED copy resolves as duplicate, i.e. stays archived) is an operator-visible outcome that differs from the plan's "32 migrate" expectation. Recorded as the M4 re-plan input; no store touched.

**Landed-ref verification**: `worktree_base_branch: ""` (level 1 absent) + `git symbolic-ref refs/remotes/origin/HEAD` → `refs/remotes/origin/develop` ⇒ `LandedRefFor(primary)` = **origin/develop** at level 2 (LandedRefOriginHEAD) — NOT the origin/main default. Reconciliation (REQ-TQM-010) must ask the landed question of origin/develop.

**Additional finding for M4**: identity-UUID collisions = 0 in real data (reconciliation.tsv header-only at dry-run).

**SPEC wording tension recorded (no blocker raised — REQ is authoritative and behavior unchanged)**: AC-TQM-003's "every old_id does NOT [exist]" cannot hold literally for renumbered pairs, because REQ-TQM-007 keeps the HOME card under the shared number; the verifier therefore checks the renumbered PROJECT variant does not survive under its old id (mapping target carries its text), and that every PROJECT-provenance reference is rewritten.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
