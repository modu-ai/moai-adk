---
id: SPEC-GTD-AUTONOMY-001
created: 2026-09-15
updated: 2026-09-15
---

# SPEC-GTD-AUTONOMY-001 구현 계획

## §A. 배경과 결정

기존 todo는 승인된 개발 카드의 순서·선택·보관에 적합하지만, 실행 전 메모 수집과 명료화, 대기·참고, 관계 최신성, 정기 검토를 표현하지 못한다. 본 변경은 큐를 교체하지 않고 관리 계층과 제한된 자율 실행을 덧붙인다.

먼저 되돌리기 어려운 결정을 고정한다.

1. 기존 SQLite 파일과 카드 정체성은 유지하고 GTD 데이터만 추가한다.
2. GTD 관계의 원본은 SQLite에 두고 비공개 그래프는 결정적 파생물로 만든다.
3. LLM은 구조화된 결정만 제안하며 일반 코드가 정책과 최신성을 검사한다.
4. `/moai:goal --auto`는 기존 goal 조건 반복과 별도인 전체 임무 상태기계다.
5. 자율 delivery는 카드 WT→단일 local develop 통합→lead batch push→검증된 develop에서 release 분기→main release PR의 저장소 로컬 순서를 따른다.
6. 지원되는 세션 runtime이 없으면 durable 실행을 주장하지 않고 `active-session-only`로 제한한다.

## §B. 알려진 위험

- 구버전 바이너리가 확장 SQLite를 다시 저장할 때 신규 정보가 손실될 수 있다.
- 정식 이름과 호환 이름을 별도 구현하면 help, completion, template가 쉽게 갈라진다.
- 외부 자료나 임무 자연어가 정책 입력으로 승격되면 prompt injection이 권한 상승으로 이어진다.
- WAL·vacuum에 민감한 파일 바이트나 mtime을 최신성 기준으로 쓰면 거짓 stale가 발생한다.
- agent idle 또는 lease 만료를 완료 신호로 사용하면 중복 배차와 중복 merge가 생긴다.
- 카드 head의 PASS를 local develop, origin/develop CI, release PR head에 재사용하면 검증되지 않은 조합이 병합될 수 있다.
- 질문 금지를 무조건 완료로 해석하면 범위 확대나 게이트 우회가 발생한다.

## §C. 실행 전 기준선

구현자는 M1 착수 직전에 다음을 같은 worktree에서 다시 관측한다.

- branch, HEAD, dirty files, 원격 기준과의 divergence
- `internal/cli/todo.go`, `internal/kanban`, `internal/graph`, goal workflow, foreman, dispatch 규칙의 현재 계약
- 실제 schema와 migration 경로, archive/undone, backup/restore, lock와 transaction 경계
- command emitter와 template mirror 목록, 4개 언어 문서 표면
- 지원되는 session start/reconnect API와 비대화식 실행 가능 여부
- 실제 운영 큐와 격리 검증 큐를 구분하는 환경 경계

기준선이 이 SPEC의 research.md와 다르면 현재 코드를 우선하고, 요구사항 충돌은 구현 전에 manager-spec으로 되돌린다.

## §D. 구현 원칙

- `constitution.development_mode: tdd`에 따라 각 변경 단위는 RED → GREEN → REFACTOR로 진행한다.
- 동작을 공유하는 기존 helper와 transaction 경계를 재사용하며 사용자 표면만 얇게 분기한다.
- migration은 추가형이고 반복 실행 가능해야 한다.
- 모든 외부 부작용은 stable `operation_id`를 먼저 기록하고 readback으로 조정한다.
- LLM의 출력은 권한이 아니라 검증 대상 데이터다.
- 일반 goal, Kanban Mode, Factory Mode의 기존 승인·질문 계약은 `mission_mode=auto`가 아닐 때 그대로 둔다.
- 실제 사용자 큐에는 마이그레이션·동시성·고장 주입 테스트를 실행하지 않는다.

## §E. 자체 검증 전략

각 milestone은 단위·통합·반례 테스트와 해당 패키지 정적 검사를 통과해야 한다. M2, M5, M7, M8은 중간 종료 위치별 fault injection을 포함한다. 모든 AC는 acceptance.md의 단일 ID로 추적한다.

## §F. Milestones

### §F.0 계획 파일 지도와 단일 writer 순서

구현은 manager-develop 한 명이 milestone 순서대로 작성한다. 병렬 write는 하지 않는다. 파일은 현재 탐색 결과에 따라 최소화하되 아래 새 경계와 기존 소유 파일을 우선 사용한다.

| Milestone | 계획 파일·심볼 | TDD task |
|---|---|---|
| M1 | `internal/cli/gtd.go` `NewGTDCommand`; `internal/cli/todo.go`; `internal/cli/gtd_compat_test.go`; command/workflow/skill template와 4개 언어 문서 | 19 verb×flag parity test를 RED로 만들고 하나의 command constructor/handler로 GREEN |
| M2 | `internal/kanban/backlog_gtd_schema.go` `MigrateGTDSchema`; `internal/kanban/backlog_gtd_compat_test.go` `TestGTDLegacyRoundTrip` | schema/migration/old-new round-trip matrix를 RED→GREEN |
| M3 | `internal/kanban/gtd_capture.go`, `gtd_clarify.go`, `gtd_organize.go`, `gtd_reflect.go`, `gtd_engage.go`; 각 표면의 table test | 절차별 허용·보류 fixture를 먼저 작성하고 최소 service 경계 구현 |
| M4 | `internal/kanban/gtd_relation.go` `ValidateGTDRelation`; `internal/graph/gtd_private.go` `BuildPrivateGTDProjection`; permission/privacy tests | 기존 관계 round-trip, cycle, deterministic projection, 0700/0600, 노출면 negative fixture |
| M5 | `internal/mission/contract.go` `SealMissionContract`; `internal/mission/policy.go` `ValidateMissionDecision`; `internal/mission/receipt.go` | incomplete contract, stale/scope/permission mutant를 RED로 만든 뒤 deterministic validator 구현 |
| M6 | `internal/cli/goal.go` `NewAutoMissionCommand`; `internal/mission/governor.go`; `.claude/agents/moai/mission-governor.md`; goal/dispatch/question 규칙과 template mirror | parser isolation, state-axis separation, approval/no-question/blocked, role-permission tests |
| M7 | `internal/mission/runtime.go` `ProbeMissionRuntimeCapability`; `internal/mission/operation.go` `ReconcileMissionOperation`; runtime/lease tests | full/partial/unsupported와 reconnect/expiry/replace/takeover/snapshot matrix를 RED→GREEN |
| M8 | `internal/mission/dispatch.go`; `internal/mission/integration.go` `IntegrateCardIntoLocalDevelop`, `ValidateDevelopBatchMerge`; `internal/mission/gtd_autonomy_e2e_test.go` | E2E, crash-cut, explicit-stage, local `--no-ff`, no lane push/card PR, SHA별 gate tests |
| M9 | 기존 문서 navigation, help/completion golden, template emission tests | source/emitted/4개 언어와 일반 mode 회귀를 nonempty sweep으로 검증 |

세부 task, 선행 조건, RED probe와 green 명령은 `tasks.md`와 `acceptance.md` evidence ledger가 소유한다.

### M1 — 명칭과 호환 계약 (High)

- 하나의 명령 구현에서 `moai gtd` 정식 표면과 `moai todo` 호환 표면을 제공한다.
- slash command, workflow skill, agent skill, template source, emitted artifact, help와 completion의 canonical/compatibility 경계를 맞춘다.
- 같은 격리 DB에 대해 두 표면의 출력, ID, 순서, archive/undone 결과가 같음을 먼저 테스트한다.
- 대상: REQ-GTD-001~004, AC-GTD-001~004.

### M2 — 추가형 GTD 저장과 왕복 보존 (High)

- 기존 schema-version 의미를 유지하는 별도 GTD schema metadata와 migration을 정의한다.
- inbox, 분류, 프로젝트·행동, 검토 시점, 정책, relation assertion, event, decision, operation receipt, mission 데이터를 기존 DB에 추가한다.
- old→new, new→old→new, backup/restore, export/import, WAL checkpoint, vacuum, archive/reopen 왕복 fixture를 구축한다.
- 대상: REQ-GTD-010~011, AC-GTD-010~011.

### M3 — GTD 다섯 절차 (High)

- Capture, Clarify, Organize, Reflect, Engage의 명령·서비스 계약을 추가한다.
- 각 절차의 보류 이유가 기계 판독 가능하고 queue write 이전에 판정되게 한다.
- 실제 development column과 queue state에는 신규 GTD 단계 값을 추가하지 않는다.
- 대상: REQ-GTD-005~009, AC-GTD-005~009.

### M4 — 관계와 비공개 GTD 그래프 (High)

- blocking/non-blocking 관계, 출처, assertion status, source revision, policy version을 저장한다.
- `depends_on` 방향·완료 의미·cycle을 검사하고 취소를 완료로 간주하지 않는다.
- SQLite logical revision으로 결정적인 비공개 projection과 sidecar metadata를 만들고 privacy fail-closed를 적용한다.
- 대상: REQ-GTD-012~013, AC-GTD-012~013.

### M5 — 사전 위임 정책 실행기 (High)

- 목표·완료 증거·범위·행위·merge·자원·금지·중단·복구·철회를 포함한 versioned contract를 도입한다.
- Proposal/Decision을 정책, snapshot, queue revision, evidence, operation identity에 대조하는 deterministic validator를 구현한다.
- 승인 없음, 범위 밖, stale, 권한 밖, 반복 이벤트, 외부 명령문을 반례로 먼저 작성한다.
- 대상: REQ-GTD-014~015, REQ-GTD-024, AC-GTD-014~015, AC-GTD-024.

### M6 — `/moai:goal --auto`와 mission governor (High)

- 자연어 임무를 기존 condition parser와 분리하는 전용 mission 생성 경로를 추가한다.
- `mission_mode=auto`와 기존 `progression_mode=autonomous|semi-autonomous`를 독립 저장·해석한다.
- approval 전 read-only, approval 후 범위 내 무질문, 범위 밖 blocked 계약을 기존 hard gate 문서·template와 함께 맞춘다.
- `mission-governor`는 쓰기 도구 없이 구조화된 Decision만 반환하고 advisor/auditor/executor 책임을 분리한다.
- 대상: REQ-GTD-016~018, AC-GTD-016~018.

### M7 — 지속 관리자와 안전한 배차 (High)

- mission lease, heartbeat, append-only event/decision/receipt log, recovery state를 도입한다.
- session runtime adapter의 supported probe, start/reconnect/replace, no-API fallback을 구현한다.
- lead가 picked card만 대상으로 현재 lane 소유권·lease·revision·기존 dispatch를 재확인하게 한다.
- 대상: REQ-GTD-019~021, AC-GTD-019~021.

### M8 — 자율 commit과 두 단계 merge (High)

- `local_develop_base_sha`에서 launcher-entered `WT-gtd-autonomy`를 만들고 lane은 explicit-path commit까지만 수행한다. lane push와 card PR은 금지한다.
- `manager-git`만 단일 `.claude/worktrees/develop` integration worktree의 lease 아래 `WT-gtd-autonomy`를 local develop에 `--no-ff`로 병합하고 `card_head_sha`와 `local_develop_merge_sha`를 readback한다.
- lead/manager-git가 여러 local merge를 모아 한 번 batch-push하고 `origin_develop_sha`가 local head와 같은지 확인한 뒤 그 SHA의 CI를 기다린다.
- CI가 PASS한 `origin_develop_sha`에서만 `release_head_sha`를 만들고, release→main PR의 audit/CI/review 후 `main_landed_sha`와 ancestry를 확인한다.
- 각 단계의 owner, integration lease, base/card/local merge/origin CI/release/main landed SHA를 별도 receipt로 기록한다.
- side effect 성공 직후 receipt 전 종료를 commit, local develop merge, batch push, release branch, release PR, final merge마다 검증한다.
- 대상: REQ-GTD-022~023, REQ-GTD-025, AC-GTD-022~023, AC-GTD-025.

### M9 — 문서·생성물·최종 회귀 (Medium)

- 4개 언어 문서, navigation, help, template source, emitted command/skill을 canonical GTD 용어로 동기화한다.
- todo 호환 표면과 기존 일반 goal/Kanban/Factory 동작을 회귀 검증한다.
- 모든 AC evidence를 progress.md §E.2에 기록할 수 있는 형태로 정리하고 독립 sync audit와 현재 SHA CI를 수행한다.

## §G. 금지 패턴

- 기존 DB 경로, 카드 ID, queue state, archive semantics를 새 이름에 맞추려고 변경한다.
- `gtd`와 `todo`를 복제 구현한다.
- GTD 다섯 단계를 board column으로 추가한다.
- 자연어 mission을 `parseCondition`, shell, command string에 전달한다.
- LLM 응답이나 advisor 권고를 직접 부작용으로 실행한다.
- repo `edges.jsonl`에 개인 GTD 메모를 합친다.
- mtime, DB 전체 바이트, PID, idle, exit 0만으로 최신성·소유권·완료를 판정한다.
- 카드 head의 CI나 audit를 local develop, origin/develop, release PR의 다른 SHA 근거로 재사용한다.
- lane이 push하거나 card-level PR을 만들고, release branch를 origin/develop CI 전에 생성한다.
- `git add -A`, `git add .`, `git commit -a`, force push, gate override를 자동 실행한다.
- 질문하지 않는다는 이유로 새 권한을 추론하거나 blocked 상태를 성공으로 바꾼다.

## §H. 교차 참조

- 개발 계획 원본: `.moai/reports/todo-gtd-autonomy/plan.md`
- 저장·큐: `internal/cli/todo.go`, `internal/kanban/`
- repo graph: `internal/graph/`
- goal surface: `.claude/skills/moai/workflows/goal.md`
- dispatch contract: `.claude/rules/moai/workflow/kanban-dispatch.md`
- foreman: `.claude/skills/moai-kanban-foreman/SKILL.md`
- frontmatter SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- phase/tier SSOT: `.claude/rules/moai/workflow/spec-workflow.md`
