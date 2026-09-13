# Tasks — SPEC-TODO-IDENTITY-001

DELTA-ID-04. Mode serial / scale Standard. 사용자 informed override 및 kickoff 승인에 따라 run 구현과 change-scoped 검증을 완료했다. 감사 판정은 BYPASSED이며 이전 FAIL을 PASS로 바꾸지 않는다. 저장소 전체 CI 판정과 sync 상태 전이는 아직 수행하지 않았다.

## 실행 규칙

각 작업은 먼저 유효 fixture에서 해당 assertion의 RED를 실행하고 원문·exit·HEAD·테스트 hash를 기록한 뒤 최소 GREEN과 보존 회귀를 수행한다. test-list 존재만으로 실행 성공을 주장하지 않는다. 후속 조건이 선행 구현에 의존하면 그 baseline을 명시하고 해당 변경 직전 RED를 실행한다. 기존 plan/AC를 축소하거나 첫 storage의 예외를 재사용하지 않는다.

계획 파일 집합은 internal/kanban 아래 backlog_store.go, backlog_sqlite.go, backlog_migrate.go, todo_runtime.go, factory_runtime.go, todo_identity_red_test.go 및 필요한 최소 identity helper/안전 테스트의 합계 6–8개다. 실제 편집 전 소유권과 변경 범위를 확인하며 새 helper 파일은 기존 분할 구조를 확인한 뒤 결정한다. 다른 작업의 변경을 되돌리지 않는다.

| ID / atomic task | REQ / AC | Dependency | Planned files | Status |
|---|---|---|---|---|
| T01 schema/version validation | REQ006 / AC006 | 없음 | backlog_sqlite.go, todo_identity_red_test.go, 최소 identity helper | completed |
| T02 pure projection / explicit null | REQ002 / AC002, AC007 | T01 | backlog_store.go, backlog_sqlite.go, todo_runtime.go, todo_identity_red_test.go | completed |
| T03 writer issuance/backfill/rollback | REQ001,003 / AC001,003 | T01,T02 | backlog_store.go, backlog_sqlite.go, backlog_migrate.go, 최소 identity helper, todo_identity_red_test.go | completed |
| T04 lifecycle / return identity | REQ004 / AC004 | T03 | backlog_store.go, todo_identity_red_test.go | completed |
| T05 runtime links / ambiguous / state authority | REQ005,007 / AC005,007 | T03,T04 | todo_runtime.go, factory_runtime.go, todo_identity_red_test.go | completed |
| T06 concurrent handles / future schema conservation | REQ006 / AC006 | T03,T05 | backlog_store.go, backlog_sqlite.go, todo_runtime.go, todo_identity_red_test.go | completed |
| T07 scoped regression / quality / MX | REQ001–007 / AC001–007 | T01–T06 | 위 변경 파일 및 필요한 최소 안전 테스트 | completed |

## 작업별 RED → GREEN 종료 기준

### T01

지원 version1과 identity 미설치 상태를 구분한다. future/partial schema 거절 assertion을 실행한 뒤 reader 비변경과 writer fail-closed를 구현한다. core/runtime version1은 유지한다. schema DDL/stamp는 approved writer transaction에서만 생성한다.

### T02

AC002 JSON/SQLite/active WAL 세 fixture의 key-present null RED를 실행한다. LoadPure는 identity 발급·DDL·stamp·migration을 하지 않고 snapshot projection만 한다. 발급된 UUID를 null로 낮추지 않는다. main/WAL/JSON과 SHM의 관측 경계는 plan을 따른다.

### T03

AC001 및 AC003을 실행한다. UUIDv7 generator를 재사용하고 project/card global cardinality 2/4/6, tNN 유지, 기존 live/archive backfill을 구현한다. entropy/insert/stamp fault 각각 실제 도달+오류+record/row/schema 변경0, fault 제거 재시도·반복 보존을 확인한다. BoardLock 및 *sql.Tx를 공유하고 MaxOpenConns(1)에서 tx 밖 DB 재진입을 금지한다. legacy JSON은 staging 성공 전 그대로 유지한다.

### T04

AC004에서 Add 반환=LoadPure와 Mutate→archive→restore→reopen의 UUID 불변을 검증한다. 기존 text/state/spec 및 archive 내용의 whole-record 보존을 검사한다.

### T05

AC005/007을 실행한다. 실제 run/assignment의 identity equality, missing/candidates2 오류·변경0을 구현한다. completed 보고는 assignment.ReportedState에만 반영한다. state API 이후 snapshot에서 UUID-only 변경 시 AddedAt/State 및 assignment OwnerLabel/ProvenanceJSON/ReportedState를 보존한다. whole-record 수정 뒤 runtime 연결도 readback한다.

### T06

서로 다른 store/SQLite handles2와 barrier의 AC006을 실행하고 successes2/cards2/last_seq2/card UUID2/project UUID1 보존을 확인한다. future identity reader/writer, 기존 future core/runtime·retired 회귀를 함께 실행한다. race는 별도 process-local 보조 검사이며 semantic 실패를 race PASS로 보지 않는다.

### T07

plan D의 정확 selector 7개와 기존 회귀 3개, race 보조 검사를 실행한다. 변경 로직 coverage85% 이상과 범위 lint를 검사하고 failure injection 도달을 기록한다. plan F의 exact MX marker5개를 bounded rg/readback으로 확인한다. 각 AC의 실제 증거가 확보되어야 독립 sync 감사로 넘긴다. umbrella23AC 완료·운영 migration·push/PR/deploy는 이 작업표의 종료 선언 대상이 아니다.
