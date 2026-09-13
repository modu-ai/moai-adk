# progress.md — SPEC-TODO-QUEUE-HOME-MERGE-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-13
plan_version: 0.4.0 (0.3.0 re-plan against the measured M3 baseline + §F0 runtime decision resolved; M4 live execution FORBIDDEN until re-audit)
artifacts: spec.md, plan.md, acceptance.md, progress.md (Tier M, 4 files)
baseline: worktree .claude/worktrees/t657-queue-merge, branch WT-todo-queue-merge, base origin/develop 5e0f71175
notes: destructive steps (M4/M5) gated behind lead window + operator approval via lead + freshness bracket; gates are verdict-evidenced preconditions per acceptance.md §AC-TQM-008. v0.3.0 additions: duplicate discriminator (REQ-TQM-006 v2, AC-TQM-010), the §F0 decision gate (AC-TQM-011), measured baseline supersedes plan-time figures. v0.4.0: the §F0 decision gate SATISFIED — option (b) merge-scope exclusion, decider = operator, via the lead's question round 2026-09-13; falsifier and loss ceiling recorded in plan.md §F0.
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

### M4 — live merge execution (2026-09-13, window 21:50–22:0x) — EXECUTED, post-verify FAIL 1건, 운영자 판정으로 유지

**게이트 이력**: (1) 창 획득 `moai integration acquire --name lane-4 --card t657` (2) 운영자 승인 — 리드 질문 채널 2026-09-13 (3) §F0 결정 기록 option (b) (4) 흡수 병합 `dae7e6219` (develop 7a7a08f20 흡수, 충돌 없음) (5) 병합 트리 재측정: kanban `AutoDone|Landed` 37 PASS/0 FAIL (11.8s), cli `TestTodoAutoDone` 22 PASS/0 FAIL (45.4s) (6) 1차 dry-run — `backlog.db-shm` 복사 중 변경으로 해시 검증 FAIL, **기록 전 안전 중단**(REQ-TQM-002 게이트 정상 작동), 재시도 성공 (7) M3-delta 새 기준선(dry-run, 판별식 적용 후): **duplicates 246 / renumbered 126 / migrated 0, high-water 834, TotalPre=TotalPost=791, ZeroLoss=true**, identity-UUID 충돌 0 (8) 신선도 브래킷 P1 — dry-run 기준선과 완전 일치(홈 59/360/seq 708, 프로젝트 105/267/seq 661) 후 실행.

**실행 결과**: 병합 기록 완료 — 홈 live 164(=59+105)/archived 381(=360+21 비중복 archived 이관)/seq 834. 실행 후 검증: ZeroLoss=true, 791/791, 126=126 일치 — 단 **StaleReferences 1건** `project archived finding t204->t538 not rewritten`으로 도구가 exit 1 (REQ-TQM-016 복구 지시문 출력, 자동 복구 없음).

**운영자 판정(리드 채널 2026-09-13): (A) 병합 유지 + 결함 추적** — 결정자=운영자, REQ-TQM-016 복구 절차를 운영자 결정으로 예외 처리, 레인 쓰기 정지 불필요. 판정 근거: 데이터 무손실(791/791), 결함은 카드 내 참조 텍스트 1건, 매핑 대조상 `t538`은 매핑 부재(재번호 대상 아님 → 해당 참조는 재작성 불요)로 **검증기 참조 장부 결함 가능성**이 지배적. 후속 카드(검증기 결함 규명 + t538 참조 수리)는 리드 발행.

**M5(프로젝트 스토어 정리): 보류 유지** — 이번 지명 범위 밖, 프로젝트 스토어는 원태 유지(105/267/661 실측).

**리드 관측 판정(t204/t718 쌍)**: 홈에 t204(원본, queued)와 t718(프로젝트 t204의 재번호 이관본, queued)이 동일 본문으로 공존하는 것은 **의도된 결과** — 실행 전 백업에 홈 원본 t204이 존재함을 실측(store-1 백업 조회), 프로젝트 t204은 live라 REQ-TQM-006 v2 판별식(생존 카드는 내용 동일 여부와 무관하게 재번호)의 정상 작동. 내용 동일 생존 쌍의 수동 중복 정리는 후속 판단 사항.

**증거**: /tmp/l4m_execute.json(오류 원문), .moai/reports/t657/{id-mapping.tsv 127행, reconciliation.tsv, backup-hashes.json}, 백업 `/Users/goos/MoAI/moai-adk-go/.moai/state/todo-merge-backup-20260913/{store-1,store-2}` (해시 검증본, 삭제 금지 — option (b)의 영구 복구 경로).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
