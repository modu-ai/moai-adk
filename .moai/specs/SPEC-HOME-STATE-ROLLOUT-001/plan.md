# plan.md — SPEC-HOME-STATE-ROLLOUT-001

## §A Context

카드 `t592`는 선행 홈 SQLite 기반을 실제 사용자 상태에 안전하게 적용하는 후속 작업이다.
구현은 TDD로 진행하며, 데이터 모델과 동시성 장벽처럼 되돌리기 어려운 결정을 먼저
고정하고 실제 live apply는 모든 자동 검증 뒤 마지막 단계에서만 수행한다.

### 확정된 가정

- `spec.md`가 요구사항 원본이며 별도 `spec.db`는 만들지 않는다.
- Todo 원본은 이전 뒤에도 보존한다. 이 작업에는 cleanup 승인이 포함되지 않는다.
- Search에는 현재 runtime producer가 없으므로 이전 대상이 아니다.
- 프로필은 프로젝트를 넘나드는 전역 자원이므로 project Factory DB를 lease 저장소로 쓰지 않는다.
- Resume handoff는 receiver dedupe가 없는 at-least-once 계약이다.
- plan-time read-only census는 `integrity_check=ok`, items `75`; `74`는 직전 보고 시점의 관측값이다.

## §B Known Issues

1. 현재 Factory schema version은 1이고 claimed resume row에는 만료와 owner identity가 없다.
2. 현재 claim 완료 경로는 token을 최종 CAS에 포함하지 않아 재청구 도입 시 ABA 방지가 필요하다.
3. 현재 `clean --home`은 현재 프로세스의 `CLAUDE_CONFIG_DIR`만 보므로 다른 Claude 프로세스를 알 수 없다.
4. MCP runtime stamp와 session registry는 시작 후 기록되므로 migration census와 시작 사이에 TOCTOU가 있다.
5. 실제 source queue는 plan 도중 74→75로 변했다. 고정된 수를 성공 기준으로 삼으면 정상 추가를 손실로 오판한다.

## §C Pre-flight

- worktree branch와 HEAD, `origin/develop...HEAD`, 변경 전 status를 기록한다.
- source/target/legacy 경로와 canonical project key를 읽기 전용으로 재계산한다.
- source 및 존재하는 target에 immutable/read-only SQLite census를 수행한다.
- 현재 active session, Factory worker, MCP runtime을 각각 조회하고 unreadable 결과를 별도 상태로 유지한다.
- 관련 package의 test, race, vet baseline을 변경 전에 기록한다.
- 각 milestone의 첫 구현 전에 대응 test를 작성하고 의도한 RED 출력을 verbatim으로 보존한다.

## §D Constraints and Preserve List

- PRESERVE: `moai migrate agency`와 `moai migration`의 기존 command tree.
- PRESERVE: Todo item/findings/archive/last-sequence 논리 표현과 순서.
- PRESERVE: linked worktree가 primary checkout과 같은 project key를 쓰는 계약.
- PRESERVE: v1 Factory DB의 모든 기존 table과 row.
- PRESERVE: `clean --home`의 dry-run 기본값과 allowlist/carve-out 삭제 방어.
- PRESERVE: SessionEnd hook의 non-fatal output contract. Lease release 실패는 보고 가능하되 기존 hook JSON을 깨지 않는다.
- PRESERVE: MCP stdio server의 blocking serve 및 startup stamp 정리.
- NEVER: 실제 사용자 홈을 테스트 fixture로 사용하지 않는다.
- NEVER: live apply와 global forced clean을 한 invocation에서 결합하지 않는다.

## §E Self-Verification Strategy

| Area | Required proof |
|------|----------------|
| CLI contract | 기본 dry-run 무변경, `--apply`만 mutation, 중복 실행 no-op |
| Data safety | backup manifest/hash, restore parity, divergent target refusal, integrity check |
| Concurrency | start-vs-migrate 세 경로, 두 census, marker crash recovery, race detector |
| Handoff | v1→v2, crash/reclaim, latest wins, token ABA refusal, at-least-once boundary |
| Profiles | provisional→enriched, direct Claude, SessionEnd release, PID reuse/fingerprint, clean skip |
| Topology | primary/linked worktree 동일 key/lock/DB |
| Platform | POSIX runtime tests; Windows compile and platform helper tests where feasible |
| Rollout | fresh live census → backup verification → apply → source/target count and integrity parity |

Changed-package coverage floor is 85%. 동시성 package는 `go test -race -count=1`을
통과해야 하며, 새 test skip은 허용하지 않는다.

## §F Milestones

### M1 — Migration barrier와 명시적 CLI (High)

1. 프로젝트 marker lifecycle과 SessionStart/Factory/MCP admission 경쟁 RED를 작성한다.
   SessionStart RED는 serialized `continue:false`와 host의 prompt/tool 처리 0건을 함께 판정한다.
2. 세 시작 경로가 동일 lock 아래 marker 확인과 runtime 등록을 끝내도록 연결한다.
3. 기존 `migrate` parent 아래 `home-state` child를 추가하고 기본 dry-run census를 구현한다.
4. `--apply`의 immediate census, private backup, divergent-target refusal, Todo logical copy,
   parity/integrity 검증, source 보존 및 idempotent no-op을 구현한다.
5. Active-owner/indeterminate refusal, backup restore probe, pre-existing target 보존,
   atomic restore, marker-last removal을 갖는 crash recovery와 성공 apply rollback을 구현한다.

Exit: start-vs-migrate, dry-run 무변경, backup restore, linked-worktree parity가 GREEN.

### M2 — Resume handoff lease와 ABA 방지 (High)

1. Factory schema v1의 pending row, parse 가능한 `claimed_at` row, NULL/invalid `claimed_at` row를
   함께 담은 fixture로 v2 additive migration과 `claimed`+legacy flag 식별 RED를 작성한다.
2. claim expiry/owner/index와 latest-eligible atomic claim을 구현한다.
3. finish CAS에 claim token을 포함하고 crash/reclaim/ABA RED를 GREEN으로 만든다.
4. `recover-resume`의 project barrier/zero-active 또는 dead-owner 확인과
   `claimed`+legacy flag→pending/failed CAS, 정상 lease claim, old-token finish 거부를 검증한다.
5. at-least-once duplicate injection boundary를 verdict와 operator 문서에 남긴다.

Exit: v1 row 보존, 최신 pending 우선, 만료 재청구, stale token 완료 거부가 GREEN.

### M3 — 전역 profile lease, 통합 검증, live rollout (High)

1. private global SQLite registry와 process-start fingerprint의 OS별 판정 seam을 정의한다.
2. launcher provisional lease, SessionStart enrichment/direct-start, SessionEnd release를 연결한다.
3. non-exec child PID 보정은 parent-token CAS handoff로 구현하고, 보정 완료 전 및 실패 후
   live/indeterminate 보호를 유지한다. Stale/PID-reuse/indeterminate 분류를 clean skip에 연결한다.
4. 모든 RED/GREEN, race, coverage, vet, native/Windows build evidence를 수집하고
   `.moai/reports/t592/verdict.md`를 작성한다.
5. `moai migrate home-state --apply --verified-live`의 pre-apply 단계가 같은 process에서
   AC-022를 제외한 named validators, current HEAD, 실행 수, lint/race/coverage/vet/build와
   fresh zero-active census를 검증하고 in-memory nonce CAS를 소비할 때까지 mutation 0건을
   보장한다. 그 다음 exclusive admission marker를 첫 mutation으로 설치하고 backup을 첫 data
   mutation으로 생성·검증한 뒤 apply한다. 적용 뒤 AC-001..021/023..025 evidence와 readback만
   입력받는 AC-022 verifier를 실행하고, harness가 AC-022 결과를 append한 뒤 sync audit가
   completeness를 닫는다. Bare apply와 stale/tampered/replayed authorization은 marker/backup/target
   생성 전에 거부한다.

Exit: live/indeterminate profile 보호, audit-ready report, current-project rollout parity가 모두 PASS.

## §G Risks and Mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| 최초 census 후 새 런타임 시작 | Critical | marker + 동일 lock 아래 start admission/register + immediate census |
| PID 재사용으로 stale lease 오판 | High | PID와 process-start fingerprint 동시 비교; 불확정은 live-safe |
| backup이 WAL 일관성을 잃음 | Critical | SQLite-consistent snapshot 후 manifest hash/restore probe |
| target의 기존 사용자 데이터 덮어쓰기 | Critical | nonempty divergent target 무조건 거부 |
| claim 재청구 뒤 오래된 consumer 완료 | High | id/status/token CAS |
| 외부 주입 후 crash로 중복 전달 | Medium | at-least-once 문서화; receiver dedupe는 후속 SPEC |
| marker가 crash 뒤 영구 차단 | Medium | operator-visible owner/started metadata와 명시적 recovery; 자동 삭제 금지 |
| Windows fingerprint 차이 | Medium | OS-specific helper tests + cross-build; 불확정 fail-safe |
| non-exec child PID 보정 중 clean 경합 | Critical | parent-token CAS handoff; 완료 전 live/indeterminate 유지; concurrent cleaner test |
| 오래된 GREEN으로 live apply | Critical | current-HEAD pre-apply set → nonce CAS → admission marker → backup → data apply 순서 고정 |

## §H Cross-References

- `SPEC-TODO-SQLITE-001` — source logical model and SQLite migration behavior
- `SPEC-V3R6-MOAI-HOME-PATHS-001` — `MOAI_HOME` path SSOT
- `SPEC-V3R6-MOAI-CLEAN-HOME-001` — cleanup allowlist and dry-run/force contract
- `SPEC-PROFILE-MEMORY-001` — profile launcher ordering and path validation
- `SPEC-V3R6-SESSION-HANDOFF-AUTO-001` — resume handoff origin contract
- Report target: `.moai/reports/t592/verdict.md`
