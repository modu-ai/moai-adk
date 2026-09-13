# progress.md — SPEC-TODO-QUEUE-HOME-MERGE-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-13
plan_version: 0.3.0 (re-planned against the measured M3 baseline per operator decision via lead; M4 live execution FORBIDDEN until re-audit)
artifacts: spec.md, plan.md, acceptance.md, progress.md (Tier M, 4 files)
baseline: worktree .claude/worktrees/t657-queue-merge, branch WT-todo-queue-merge, base origin/develop 5e0f71175
notes: destructive steps (M4/M5) gated behind lead window + operator approval via lead + freshness bracket; gates are verdict-evidenced preconditions per acceptance.md §AC-TQM-008. v0.3.0 additions: duplicate discriminator (REQ-TQM-006 v2, AC-TQM-010), runtime-persistence decision gate ([NEEDS CLARIFICATION] plan.md §F0, AC-TQM-011), measured baseline supersedes plan-time figures.
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

**Landed-ref observation (DATED, not a baseline)**: on 2026-09-13, `worktree_base_branch` read `""` (level 1 absent) and `git symbolic-ref refs/remotes/origin/HEAD` answered `refs/remotes/origin/develop` — so `LandedRefFor` resolved to **origin/develop** (level 2, LandedRefOriginHEAD), not the origin/main default. `origin/HEAD` is a MOVING ref: per plan.md §C P4 this reading is a dated reference only; the in-window re-derivation inside M4 (P4, recorded fresh in the verdict) is the evidence reconciliation must rely on.

**Additional finding for M4**: identity-UUID collisions = 0 in real data (reconciliation.tsv header-only at dry-run).

**SPEC wording tension recorded (no blocker raised — REQ is authoritative and behavior unchanged)**: AC-TQM-003's "every old_id does NOT [exist]" cannot hold literally for renumbered pairs, because REQ-TQM-007 keeps the HOME card under the shared number; the verifier therefore checks the renumbered PROJECT variant does not survive under its old id (mapping target carries its text), and that every PROJECT-provenance reference is rewritten.

### M1-delta — REQ-TQM-006 v2 discriminator (live-card absorption closed) — 2026-09-13

- Change: `classify` (internal/kanban/todo_queue_merge.go) now takes the project population as input. Duplicate ⟺ project card from the ARCHIVED population AND byte-exact content equality. ANY project LIVE card (queued/picked/dropped) colliding with a home card — live or archived — renumbers above the high-water with a mapping row, regardless of content similarity. Duplicate rows carry `Origin: project-archived` (the only admitted origin).
- AC-TQM-010 additions: `CensusRecords` pre-merge population census on every procedure outcome (surfaced in the tool's JSON), and a verifier population check — a duplicate row whose id is absent from the project ARCHIVED population is a stale-reference failure (live-work absorption = zero-loss violation).
- RED (verbatim, pre-fix, `go test ./internal/kanban/ -run 'TestMergeBacklogRecordsLiveArchivedPairRenumbers' -count=1 -v`): all three live states observed absorbed — `LIVE card absorbed as duplicate (the operator-named hazard): [{ID:t5 Origin:}]` for picked/queued/dropped; positive control failed on `duplicate row Origin = ""` (field not yet set). `FAIL github.com/modu-ai/moai-adk/internal/kanban`.
- GREEN: same selector → PASS; full new-suite selector (32 tests) → `ok github.com/modu-ai/moai-adk/internal/kanban 0.632s`. Rehearsal fixture updated: live identical t10 pair now renumbers (t31), project-archived t30 twin resolves duplicate; census assertions added.
- Real-store dry-run figures (290/82) are superseded by the discriminator and by plan v0.3.0 §A.1: post-discriminator counts are M3-delta's to re-derive. NO real-store re-run performed in M1-delta (delegation forbids it).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
