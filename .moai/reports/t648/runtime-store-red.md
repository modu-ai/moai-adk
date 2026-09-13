# t648 — Runtime Store 구현 전 검사

## 2026-09-12 추가 계약 검사 — 현재 기준

아래 추가 검사는 이전 실행 기록을 대체하는 최신 기준이다. 기존 기록은 감사 추적을 위해 보존한다. `moai-domain-database`와 `moai-workflow-tdd`를 읽고 실제 공개 API로 조건에 도달하는 테스트만 추가했다. 생산 코드는 변경하지 않았다.

### Claim

최상위 테스트 6개 중 신규 계약을 검사하는 4개는 semantic RED, 기존 보존/검사자체 guard 2개는 PASS다. 미래 버전 테스트는 두 하위 경우 모두 실제 API가 오류 없이 성공하는 것을 관측했다. 일반 폴더이며 SPEC이 없는 카드 할당도 성공하지만 Todo 공개 JSON에는 실행 기록이 없다. 현재 파일은 전체 4 AC의 모든 조건을 RED로 증명하지 않는다.

### Evidence

```sh
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoRuntimeStore' -count=1 -v -timeout=180s
```

종료 코드 1, 원문 출력:

```text
=== RUN   TestTodoRuntimeStorePublicReadbackSurvivesCardEdit
    todo_runtime_store_test.go:131: before edit: successful runtime writes missing from public Todo JSON: {"version":1,"last_seq":1,"items":[{"id":"t1","text":"runtime before edit","added_at":"2026-09-12T02:51:11Z","spec_id":null,"state":"queued"}],"findings":[],"archived":[]}
    todo_runtime_store_test.go:135: after edit: successful runtime writes missing from public Todo JSON: {"version":1,"last_seq":1,"items":[{"id":"t1","text":"runtime after edit","added_at":"2026-09-12T02:51:11Z","spec_id":null,"state":"queued"}],"findings":[],"archived":[]}
    todo_runtime_store_test.go:149: legacy runs increased from 0 to 1
    todo_runtime_store_test.go:149: legacy cards increased from 0 to 1
--- FAIL: TestTodoRuntimeStorePublicReadbackSurvivesCardEdit (2.72s)
=== RUN   TestTodoRuntimeStoreFutureSchemaPreservesBytes
--- PASS: TestTodoRuntimeStoreFutureSchemaPreservesBytes (0.00s)
=== RUN   TestTodoRuntimeStoreByteGuardRejectsSameLengthChange
--- PASS: TestTodoRuntimeStoreByteGuardRejectsSameLengthChange (0.00s)
=== RUN   TestTodoRuntimeStoreNonGitWithoutSpecSeedsRun
    todo_runtime_store_test.go:235: successful no-Git/no-SPEC assignment missing Todo runtime: {"version":1,"last_seq":1,"items":[{"id":"t1","text":"plain folder task","added_at":"2026-09-12T02:51:14Z","spec_id":null,"state":"queued"}],"findings":[],"archived":[]}
--- FAIL: TestTodoRuntimeStoreNonGitWithoutSpecSeedsRun (0.49s)
=== RUN   TestTodoRuntimeStoreLegacyPureReadReturnsEmptyRuntimeWithoutWriting
    todo_runtime_store_test.go:284: read 0: legacy public JSON lacks empty runtime: {"version":1,"last_seq":1,"items":[{"id":"t1","text":"legacy card","added_at":"2026-09-12T02:51:14Z","spec_id":null,"state":"queued"}],"findings":[],"archived":[]}
    todo_runtime_store_test.go:284: read 1: legacy public JSON lacks empty runtime: {"version":1,"last_seq":1,"items":[{"id":"t1","text":"legacy card","added_at":"2026-09-12T02:51:14Z","spec_id":null,"state":"queued"}],"findings":[],"archived":[]}
--- FAIL: TestTodoRuntimeStoreLegacyPureReadReturnsEmptyRuntimeWithoutWriting (0.11s)
=== RUN   TestTodoRuntimeStoreWriterRejectsFutureTodoVersions
=== RUN   TestTodoRuntimeStoreWriterRejectsFutureTodoVersions/schema_version
    todo_runtime_store_test.go:331: runtime run writer accepted future Todo schema_version=999
    todo_runtime_store_test.go:334: runtime assignment writer accepted future Todo schema_version=999
=== RUN   TestTodoRuntimeStoreWriterRejectsFutureTodoVersions/runtime_schema_version
    todo_runtime_store_test.go:331: runtime run writer accepted future Todo runtime_schema_version=999
    todo_runtime_store_test.go:334: runtime assignment writer accepted future Todo runtime_schema_version=999
--- FAIL: TestTodoRuntimeStoreWriterRejectsFutureTodoVersions (1.69s)
    --- FAIL: TestTodoRuntimeStoreWriterRejectsFutureTodoVersions/schema_version (0.87s)
    --- FAIL: TestTodoRuntimeStoreWriterRejectsFutureTodoVersions/runtime_schema_version (0.82s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	5.398s
FAIL
```

### Baseline-attribution

- WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-unified`
- HEAD: `a315dad9af0d3a0e04862e6106b3993d9a3812f7`, branch `WT-todo-unified`.
- 수정 직전 `git fetch origin main --quiet` 성공, `git rev-list --count --left-right origin/main...HEAD`: `0 3005`.
- 최신 테스트 파일 SHA-256: `08110f95f366d174d9bbf68987fd843ca7ce10c682690732ccf972c3c18ce7a7`.
- `gofmt -d internal/kanban/todo_runtime_store_test.go`: 출력 없음, 종료 코드 0.
- 실행 fixture는 독립 임시 폴더와 임시 MOAI_HOME만 사용했다. no-Git 경우에는 Git 초기화도 SPEC 작성도 하지 않았다.

### Gaps

| AC | 이번에 실제 도달한 조건 | 아직 도달하지 못한 조건 |
|---|---|---|
| TRS-001 | Git/SPEC 및 no-Git/no-SPEC 입력의 기존 API 성공 후 공개 runtime 부재 RED, legacy 증가 | runtime 내부 실제 값과 implicit run pair assertions는 runtime 부재 뒤여서 미실행. archive·정렬 미검증 |
| TRS-002 | 기존 큐 반복 공개 조회의 빈 runtime 계약 RED, DB bytes 불변 guard 실행 | 실행 기록이 포함된 동일 snapshot, 활성 WAL runtime snapshot 미검증 |
| TRS-003 | 기존 카드 edit 성공 후 공개 runtime 부재 | 경쟁·implicit seed 뒤 assignment abort·rollback은 아직 runtime table 쓰기 경로가 없어 도달 불가. 동일 부재를 rollback RED라고 재분류하지 않음 |
| TRS-004 | 미래 core/extension key=999를 실제 기존 engine으로 심고 run/assignment API의 거절 누락 RED, Todo DB bytes guard 실행 | 퇴역 writer·확장 설치 도중 실패와 원자적 rollback 미검증 |

AC003의 trigger를 없는 Todo runtime 테이블에 설치하다 실패하는 것은 setup 오류다. 미래 테이블을 테스트에서 강제로 만들더라도 현재 Factory API는 그 테이블을 쓰지 않으므로 trigger에 도달하지 않는다. 따라서 해당 조건은 storage 경로 도입 후 별도 단계에서 진짜 실패 지점으로 검증해야 한다. 이 사실을 SPEC 작성 담당자와 루트에 전달했다.

### Residual-risk

위 RED는 제안 계약의 미구현 증거이며 운영 자료 유실을 관측한 증거가 아니다. 두 DB 쓰기 경로 변경만으로 동시성·원자성·운영 완료가 입증되지 않는다. 독립 감사와 kickoff 전 생산 구현은 하지 않았다. 전체 저장소 CI verdict는 통합 브랜치 CI 소유이며 PENDING이다.

## Claim

SPEC-TODO-RUNTIME-STORE-001의 확정 JSON 계약으로 실제 기존 run/card API를 호출하고 공개 Todo JSON을 검사했다. 신규 동작 검사 1개는 의도한 assertion에서 RED, 신규 보존 guard 2개와 기존 회귀 3개는 PASS다. 생산 코드는 변경하지 않았다. 전체 AC 통과나 전체 AC의 RED 확보를 주장하지 않는다.

## Evidence

작업 경로: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-unified`.

```sh
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoRuntimeStore' -count=1 -v -timeout=180s
```

원문 출력:

```text
=== RUN   TestTodoRuntimeStorePublicReadbackSurvivesCardEdit
    todo_runtime_store_test.go:131: before edit: successful runtime writes missing from public Todo JSON: {"version":1,"last_seq":1,"items":[{"id":"t1","text":"runtime before edit","added_at":"2026-09-11T15:13:55Z","spec_id":null,"state":"queued"}],"findings":[],"archived":[]}
    todo_runtime_store_test.go:135: after edit: successful runtime writes missing from public Todo JSON: {"version":1,"last_seq":1,"items":[{"id":"t1","text":"runtime after edit","added_at":"2026-09-11T15:13:55Z","spec_id":null,"state":"queued"}],"findings":[],"archived":[]}
    todo_runtime_store_test.go:149: legacy runs increased from 0 to 1
    todo_runtime_store_test.go:149: legacy cards increased from 0 to 1
--- FAIL: TestTodoRuntimeStorePublicReadbackSurvivesCardEdit (4.32s)
=== RUN   TestTodoRuntimeStoreFutureSchemaPreservesBytes
--- PASS: TestTodoRuntimeStoreFutureSchemaPreservesBytes (0.01s)
=== RUN   TestTodoRuntimeStoreByteGuardRejectsSameLengthChange
--- PASS: TestTodoRuntimeStoreByteGuardRejectsSameLengthChange (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	4.807s
FAIL
```

종료 코드 1. 실행 3개, 신규 행동 SEMANTIC RED 1개, 보존/검사자체 PASS 2개. 실제 Git/spec/card fixture와 기존 API 세 가지 호출, 카드 edit, JSON readback 및 legacy SQL query 성공 후 실패했다. setup/컴파일/timeout 실패가 아니다.

기존 회귀 명령:

```sh
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^(TestBacklogEngineReopenIsIdempotent|TestBacklogEngineRefusesUnknownSchemaVersionWithoutDestroying|TestBacklogMutate_CallbackRefusalLeavesFileUnchanged)$' -count=1 -v -timeout=180s
```

```text
=== RUN   TestBacklogEngineReopenIsIdempotent
--- PASS: TestBacklogEngineReopenIsIdempotent (0.01s)
=== RUN   TestBacklogEngineRefusesUnknownSchemaVersionWithoutDestroying
--- PASS: TestBacklogEngineRefusesUnknownSchemaVersionWithoutDestroying (0.01s)
=== RUN   TestBacklogMutate_CallbackRefusalLeavesFileUnchanged
=== PAUSE TestBacklogMutate_CallbackRefusalLeavesFileUnchanged
=== CONT  TestBacklogMutate_CallbackRefusalLeavesFileUnchanged
--- PASS: TestBacklogMutate_CallbackRefusalLeavesFileUnchanged (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.547s
```

종료 코드 0. 기존 회귀 3개 PASS이며 신규 runtime/rollback의 RED 대체 근거가 아니다. 기존 future-schema 검사는 길이 불변을 확인하므로 신규 `bytes.Equal` guard를 별도로 실행했다. 같은 길이의 다른 내용 counterexample은 byte 검사 자체를 확인하며 생산 데이터 손상 주장이 아니다.

## Baseline-attribution

- HEAD: `a315dad9af0d3a0e04862e6106b3993d9a3812f7`
- Branch: `WT-todo-unified`
- fetch origin main 성공, left/right `0`, `3005`.
- 테스트 파일 SHA-256: `412dada462991482c41e6a9aaad2f9158cff26949bc6fba0e513b79b50dadde6`
- 신규 소유 파일: `internal/kanban/todo_runtime_store_test.go`, 이 보고서.
- 기존 `todo_unified_red_test.go`와 baseline 보고서는 변경하지 않았다.
- `gofmt -d` 출력 없음, 종료 코드 0.
- 생산 코드·SPEC·운영 home·카드 상태·commit·push 변경 없음.

## Gaps

| AC | 관측 범위 | 미관측 |
|---|---|---|
| TRS-001 | 기존 run/assignment/state 호출 성공 뒤 실제 공개 JSON에 runtime 부재, legacy runs/cards 각 1 증가 | runtime 값 assertions는 부재에서 멈춰 아직 실행 안 됨. archive, implicit run, 미존재 카드 거절, 정렬 미검증 |
| TRS-002 | 기존 LoadPure로 실제 카드와 edit 확인 | runtime snapshot, active WAL, 구 큐 빈 배열, 읽기 비변경 전체 조건 미검증 |
| TRS-003 | 기존 Mutate의 카드 edit 성공 후 runtime 부재 확인 | runtime이 처음부터 없어 보존 성공 아님. 경쟁, assignment abort, implicit run rollback 미검증 |
| TRS-004 | 미래 core version 실제 거절과 DB bytes 불변 PASS, 기존 reopen/refusal PASS | 미래 확장 version, 퇴역 writer, 확장 설치 실패/원자성 미검증 |

현재 assignment API는 Todo namespaced table이 아니라 legacy DB에 쓰므로, 테스트에서 미래 table과 trigger를 설치해도 해당 trigger에 도달하지 않는다. 이것을 rollback RED로 분류할 수 없다. 새 저장 경로 구현 전 단계에서 해당 transaction rollback이 실제 실행됐다는 증거는 확보하지 못했다.

## Residual-risk

JSON runtime 부재만 해소해서는 AC 전체를 충족하지 않는다. provenance 값·upsert·archive/implicit run·동일 snapshot·경쟁·오류 지점 도달 및 rollback·확장 schema 보호를 후속 실행으로 검증해야 한다. 단순 mirror 삭제로 기존 negative 검사만 통과할 수 있는 위험을 이번 positive readback 요구가 일부 보완하지만 전체 권한/완료 정확성을 다루지는 않는다. 전체 CI는 통합 브랜치 CI 담당이며 현재 PENDING이다.
