---
id: SPEC-HOME-STATE-ROLLOUT-001
title: "홈 상태 안전 전환, handoff 재청구, 전역 프로필 lease"
version: "0.5.0"
status: draft
created: 2026-09-09
updated: 2026-09-09
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/homestate, internal/cli, internal/hook, internal/session, internal/kanban"
lifecycle: spec-anchored
tags: "home-state, migration, sqlite, handoff, profile-lease, concurrency, recovery"
tier: L
depends_on: [SPEC-TODO-SQLITE-001, SPEC-V3R6-MOAI-HOME-PATHS-001, SPEC-V3R6-MOAI-CLEAN-HOME-001, SPEC-PROFILE-MEMORY-001, SPEC-V3R6-SESSION-HANDOFF-AUTO-001]
---

# SPEC-HOME-STATE-ROLLOUT-001 — 홈 상태 안전 전환

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-09 | manager-spec | 카드 `t592`의 Tier L 계획 초안. 명시적 홈 상태 이전, 시작 장벽, handoff v2 재청구, 전역 프로필 lease 및 실제 프로젝트 전환 검증을 정의했다. |
| 0.2.0 | 2026-09-09 | manager-spec | Plan audit D1–D6 반영. legacy claim 승격, SessionStart 중단 프로토콜, crash recovery/rollback, 실행형 live gate, non-exec lease 인계와 AC 증거 장부를 명확히 했다. |
| 0.3.0 | 2026-09-09 | manager-spec | Plan audit iteration 2의 D1/D2/D5 반영. 공유 문자열 probe를 제거하고 AC별 named validator 계약, legacy claim operator recovery, in-process verified live gate를 고정했다. |
| 0.4.0 | 2026-09-09 | manager-spec | Plan audit iteration 3의 D2/D5 반영. 기존 handoff status만 쓰는 legacy recovery와 in-memory authorization CAS, pre/post verifier 분리를 도입했다. |
| 0.5.0 | 2026-09-09 | manager-spec | Final plan gate D5a/D5b 반영. nonce CAS 뒤 admission marker를 먼저 설치하고 backup/data apply로 진행하며, AC-022 입력과 자체 결과 append/closure를 비순환 단계로 분리했다. |

## §A 배경과 목표

선행 작업은 Todo와 Factory/handoff의 새 SQLite 경로를 마련했지만 실제 사용자 홈을
자동으로 옮기지 않았다. 이는 의도된 안전 정지점이다. 사용자 상태를 옮기는 동안 새
세션이나 Factory/MCP 런타임이 시작되면, 최초 census가 깨끗하더라도 쓰기가 다시
발생해 복사본이 즉시 오래된 상태가 될 수 있다. 또한 현재 resume handoff의 claim은
소비자가 중단되면 회수할 만료 정보가 없고, 프로필 정리기는 다른 프로세스가 사용 중인
프로필을 전역적으로 판별하지 못한다.

이 SPEC의 목표는 다음 세 계약을 하나의 안전한 rollout으로 완성하는 것이다.

1. `moai migrate home-state`를 통해 명시적으로만 사용자 상태를 이전한다.
2. 이전 중에는 프로젝트의 신규 런타임 진입을 동일한 장벽으로 차단한다.
3. resume handoff와 Claude 프로필 사용 상태를 프로세스 중단 뒤에도 보수적으로 복구한다.

## §B 요구사항 (GEARS)

### §B.1 명시적 홈 상태 이전

- **REQ-HSR-001** (Event-driven): **When** an operator invokes `moai migrate home-state`, the command **shall** 기본적으로 읽기 전용 dry-run을 수행하고, 명시적인 `--apply`가 있을 때만 대상 상태를 변경한다.
- **REQ-HSR-002** (Event-driven): **When** dry-run or apply begins, the migration command **shall** source, target, canonical project root, project key, active-runtime census, logical row counts, and SQLite integrity results를 보고한다. Search 상태는 런타임 producer가 없으므로 정확히 `not-applicable (no runtime producer)`로 보고한다.
- **REQ-HSR-003** (State-driven): **While** apply is preparing to write, the migration command **shall** 프로젝트 전용 private 장벽을 획득하고, 장벽 전 census와 대상 쓰기 직전 census를 모두 수행한다. 둘 중 하나라도 읽을 수 없거나 활성 여부가 불확정이면 fail-closed로 중단한다.
- **REQ-HSR-004** (Event-driven): **When** apply is authorized to proceed, the migration command **shall** 첫 대상 변경 전에 timestamped private backup과 manifest를 만들고, canonical root, project key, source 경로, 논리 개수, integrity 결과 및 SHA-256을 기록한다.
- **REQ-HSR-005** (Unwanted): The migration command **shall not** overwrite a non-empty target whose logical content differs from the source. Source와 target이 논리적으로 같고 양쪽 `PRAGMA integrity_check`가 `ok`인 경우에만 이미 이전된 상태로 인정한다.
- **REQ-HSR-006** (Event-driven): **When** an apply copy completes, the migration command **shall** source/target의 Todo 논리 개수와 전체 논리 표현을 비교하고 양쪽 integrity를 확인한 뒤 성공을 보고한다. 검증 실패 시 불완전 target을 성공 상태로 남기지 않는다.
- **REQ-HSR-007** (Ubiquitous): The migration command **shall** 원본 상태와 backup을 복구 가능한 형태로 보존하며, 이 SPEC은 원본 삭제·강제 정리·전역 cleanup을 수행하지 않는다.
- **REQ-HSR-008** (Event-driven): **When** the same completed migration is invoked again, the command **shall** source와 target의 동등성을 다시 확인한 뒤 mutation 없는 idempotent no-op으로 종료한다.
- **REQ-HSR-009** (Ubiquitous): Linked worktrees and the primary checkout **shall** canonicalize to one project key, one target Todo DB, and one migration barrier; worktree별로 홈 상태를 분기하지 않는다.

### §B.2 이전 TOCTOU 차단

- **REQ-HSR-010** (State-driven): **While** a home-state apply marker is active or unreadable, SessionStart **shall** serialized `{"continue":false,"stopReason":"..."}`를 반환하고 host는 prompt·tool 후속 처리를 시작하지 않는다. Factory worker registration과 MCP server start도 신규 등록 또는 serving 전에 명시적으로 중단한다.
- **REQ-HSR-011** (Ubiquitous): Each guarded runtime start **shall** 동일한 프로젝트 장벽 아래에서 marker 부재 확인과 자체 runtime 등록을 하나의 직렬화 구간으로 수행하여 check-then-register 경쟁을 만들지 않는다.
- **REQ-HSR-012** (Event-driven): **When** an apply process exits before clearing its marker, later starts **shall** 해당 marker를 활성 또는 불확정 상태로 취급해 fail-closed하고, operator-visible 복구 정보 없이 자동 삭제하지 않는다.

### §B.3 Resume handoff schema v2

- **REQ-HSR-013** (Event-driven): **When** a v1 Factory DB is opened, the handoff store **shall** 기존 row를 보존하면서 claim expiry, owner PID/session, legacy-recovery flag/reason 및 만료 조회 index를 갖는 schema v2로 원자적으로 승격한다. 기존 `claimed` row는 parse 가능한 `claimed_at`을 기준으로 expiry를 backfill하고, timestamp가 없거나 해석 불가능하면 기존 `claimed` status를 유지한 채 legacy-recovery flag로 식별해 정상 claim에서 제외한다.
- **REQ-HSR-014** (Event-driven): **When** consumers race to claim resume work, the handoff store **shall** 최신 eligible handoff를 하나의 transaction에서 선택한다. 최신 pending이 우선하며, pending이 없을 때 만료된 claimed row를 원자적으로 재청구한다.
- **REQ-HSR-015** (Event-driven): **When** a consumer finishes or fails a claimed handoff, the state transition **shall** handoff identity, `claimed` 상태, and `claim_token`을 모두 비교하는 CAS여야 하며 오래된 consumer가 재청구된 row를 끝내는 ABA를 허용하지 않는다.
- **REQ-HSR-016** (Ubiquitous): Resume delivery **shall** at-least-once semantics를 명시한다. 외부 context 주입 성공 뒤 consumed CAS 전에 중단되면 동일 payload가 다시 주입될 수 있으며, receiver-side exactly-once dedupe는 이 SPEC에서 구현하지 않는다.

### §B.4 전역 Claude 프로필 lease

- **REQ-HSR-017** (Ubiquitous): Claude profile leases **shall** 전역 private SQLite registry `~/.moai/run/profile-leases.db`에 저장하고 project Factory DB에는 저장하지 않는다. Registry DB, WAL, SHM은 사용자 전용 권한을 유지한다.
- **REQ-HSR-018** (Event-driven): **When** a MoAI launcher is about to replace or start the Claude process, it **shall** exec 전에 provisional lease를 기록한다. SessionStart는 session identity로 이를 enrich하고, MoAI launcher를 거치지 않은 direct Claude start도 `CLAUDE_CONFIG_DIR`에서 named profile을 식별할 수 있으면 lease를 등록한다. Child-process 플랫폼에서는 parent token을 조건으로 child PID/fingerprint를 CAS 인계하며, 인계가 완료되거나 명시적으로 취소될 때까지 provisional lease를 live 또는 indeterminate로 보존한다.
- **REQ-HSR-019** (Event-driven): **When** SessionEnd runs, the lease registry **shall** 일치하는 session lease를 해제한다. 비정상 종료 뒤 stale 판정은 PID와 process-start fingerprint가 함께 일치하지 않을 때만 확정하며, PID가 재사용됐거나 판별이 불가능한 경우를 서로 구분한다.
- **REQ-HSR-020** (Capability gate): **Where** `moai clean --home` evaluates profile content, the cleaner **shall** live lease가 있거나 lease 판정이 불확정인 프로필 전체를 건너뛰고 이유를 보고한다. 확정적으로 stale인 lease만 정리 후보 판정 전에 회수할 수 있다.

### §B.5 Rollout 및 검증 계약

- **REQ-HSR-021** (Event-driven): **When** the current project is rolled out, the sole mutation entry `moai migrate home-state --apply --verified-live` **shall** pre-apply validator set에서 AC-022를 제외하고 current HEAD, executed-test counts, strict lint, race/coverage/vet/native, Windows disposition 및 fresh determinate zero-active census를 검증한다. It **shall** 외부 입력으로 받지 않은 nonce를 HEAD와 census fingerprint에 결합해 in-memory CAS로 한 번 소비한 다음 exclusive admission marker를 첫 mutation으로 설치하여 신규 SessionStart/Factory/MCP admission을 차단하고, 그 뒤 backup을 첫 data mutation으로 생성한 후 data apply를 수행한다. Bare apply와 stale/tampered/replayed authorization은 marker/backup/target 생성 전에 거부한다. Post-apply readback과 AC-022 verdict validator는 authorization input이 아니라 적용 뒤 독립 단계다.
- **REQ-HSR-022** (Ubiquitous): Run-phase completion **shall** 각 신규 동작의 verbatim RED 실패 출력, GREEN 결과, race-sensitive concurrency checks, backup restore proof, linked-worktree/project-key proof, Windows compile/test evidence where feasible, 그리고 `.moai/reports/t592/verdict.md`를 포함한다.
- **REQ-HSR-023** (Event-driven): **When** apply가 marker를 남긴 채 중단되면, an operator recovery entry point **shall** active owner를 거부하고 owner fingerprint가 불확정이면 fail-closed하며, source/target/backup manifest와 hash를 검증하고 안전 상태를 복구한 뒤 marker를 마지막에 제거한다. 같은 recovery 반복 실행은 mutation 없는 no-op이어야 한다.
- **REQ-HSR-024** (Event-driven): **When** a completed migration must be rolled back, an operator rollback entry point **shall** 선택한 verified backup ID를 이용해 기존 target을 보존 가능한 형태로 격리하고 원자적으로 복원하며, logical parity와 integrity를 확인한 뒤 marker를 마지막에 제거한다.
- **REQ-HSR-025** (Event-driven): **When** schema v2 identifies a legacy `claimed` handoff with NULL/invalid expiry, `moai factory handoff recover-resume --id <id> --expected-token <token> --decision <requeue|fail>` **shall** row ID, current `claimed` status/token, legacy-recovery flag, owner PID/start fingerprint와 operator decision을 한 transaction의 CAS 조건으로 검증한다. Live 또는 indeterminate owner는 거부한다. Dead owner 또는 project barrier 아래 determinate zero-active census로 확인된 unknown legacy owner는 기존 허용 상태인 `pending` 또는 `failed`로만 전이한다. Requeue 뒤 정상 claim이 새 token/expiry를 발급하며 이전 token finish는 새 상태를 바꾸지 못한다.

## §C 성공 조건

- 상태를 바꾸지 않는 dry-run만으로 이전 가능 여부와 차단 원인을 판정할 수 있다.
- apply 성공은 backup, logical parity, integrity, and source preservation으로 재현 가능하다.
- SessionStart, Factory, MCP의 세 시작 경로가 동일한 장벽 계약을 공유한다.
- 중단된 handoff는 만료 후 재청구할 수 있고 오래된 token은 최종 상태를 바꿀 수 없다.
- 다른 프로세스가 사용하는 프로필은 `clean --home --force`에서도 삭제 후보가 되지 않는다.

## §D 제약

- Go 1.26과 기존 pure-Go SQLite driver를 유지한다.
- 프로젝트 home-state 장벽과 전역 profile lease registry를 혼합하지 않는다.
- unreadable 또는 indeterminate 상태는 안전한 상태로 추정하지 않는다.
- 기존 Todo 논리 모델과 CLI 출력·exit-code 계약을 보존한다.
- 테스트는 실제 사용자 홈 대신 격리된 `MOAI_HOME`을 기본으로 사용한다. 실제 rollout AC만 승인된 live project를 읽고 쓴다.

## §E 제외 범위

### Out of Scope — Search producer

- Search DB producer, rebuild, invalidation, FTS schema 및 자동 SessionEnd indexing은 구현하지 않는다.
- 이 SPEC의 이전 보고서는 Search를 `not-applicable (no runtime producer)`로만 표시한다.

### Out of Scope — Exactly-once 수신자

- 주입 대상이 payload ID를 기억하는 receiver-side exactly-once dedupe는 구현하지 않는다.
- Resume handoff는 at-least-once이며 중단 경계의 중복 가능성을 문서화한다.

### Out of Scope — 전역 강제 정리

- `~/.moai` 전체 삭제, 원본 Todo 삭제, backup 자동 삭제, 모든 프로필 일괄 강제 정리는 수행하지 않는다.
- 기존 Factory 메시징 topology나 Claude↔Codex transport를 변경하지 않는다.
