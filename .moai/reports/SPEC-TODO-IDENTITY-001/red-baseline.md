# SPEC-TODO-IDENTITY-001 — RED baseline

## Claim

현재 생산 코드에 identity 기능이 없는 상태에서 7개 AC 모두를 컴파일·fixture·timeout·zero-test가 아닌 의미적 RED로 재현했다. AC별 selector는 총 15개 top-level test를 실행한다. 기존에 지원하던 missing-card 거절과 retired writer 거절은 신규 RED와 분리해 2 PASS로 보존했다. 이 문서는 RED 증거이며 구현 완료나 PASS를 주장하지 않는다.

## Evidence

공통 환경 scrub은 모든 selector에서 한 invocation 안에 적용했다.

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '<selector>' -count=1 -v -timeout=120s
```

### AC-TID-001 — Git·비Git, 같은 프로젝트 2카드, tNN/last_seq, UUIDv7

Selector: `^TestTodoIdentityAC001`; top-level test 1, subtest 2, fixture 2, card 4. Exit 1.

```text
=== RUN   TestTodoIdentityAC001GitAndNonGitUUIDv7
=== RUN   TestTodoIdentityAC001GitAndNonGitUUIDv7/git
    todo_identity_red_test.go:178: git project: missing project_uuid after successful writer
    todo_identity_red_test.go:179: git first Add: missing card_uuid after successful writer
    todo_identity_red_test.go:180: git second Add: missing card_uuid after successful writer
    todo_identity_red_test.go:181: git first LoadPure: missing card_uuid after successful writer
    todo_identity_red_test.go:182: git second LoadPure: missing card_uuid after successful writer
=== RUN   TestTodoIdentityAC001GitAndNonGitUUIDv7/non-git
    todo_identity_red_test.go:178: non-git project: missing project_uuid after successful writer
    todo_identity_red_test.go:179: non-git first Add: missing card_uuid after successful writer
    todo_identity_red_test.go:180: non-git second Add: missing card_uuid after successful writer
    todo_identity_red_test.go:181: non-git first LoadPure: missing card_uuid after successful writer
    todo_identity_red_test.go:182: non-git second LoadPure: missing card_uuid after successful writer
=== NAME  TestTodoIdentityAC001GitAndNonGitUUIDv7
    todo_identity_red_test.go:188: AC-TID-001 executed fixtures=2 git=1 non_git=1 cards=4
--- FAIL: TestTodoIdentityAC001GitAndNonGitUUIDv7
FAIL
```

RED reason: 기존 Add와 LoadPure는 성공했고 `t1`, `t2`, `last_seq=2` 검사는 통과했지만 `project_uuid`/`card_uuid`를 노출하지 않는다. helper는 nonzero, canonical lowercase, RFC UUID version 7을 모두 요구한다.

### AC-TID-002 — legacy JSON·SQLite·active WAL, live/archive, 명시 null, 순수읽기

Selector: `^TestTodoIdentityAC002`; top-level test 3, JSON 1, SQLite 1, active-WAL committed snapshot 1, 각 live/archive 1. Exit 1.

```text
=== RUN   TestTodoIdentityAC002LegacyJSONExplicitNullNoMutation
    todo_identity_red_test.go:201: JSON project: project_uuid key absent, want key-present literal null
    todo_identity_red_test.go:202: JSON live: card_uuid key absent, want key-present literal null
    todo_identity_red_test.go:203: JSON archive: card_uuid key absent, want key-present literal null
    todo_identity_red_test.go:210: AC-TID-002 executed JSON=1 live=1 archive=1
--- FAIL: TestTodoIdentityAC002LegacyJSONExplicitNullNoMutation
=== RUN   TestTodoIdentityAC002LegacySQLiteExplicitNullNoMutation
    todo_identity_red_test.go:221: SQLite project: project_uuid key absent, want key-present literal null
    todo_identity_red_test.go:222: SQLite live: card_uuid key absent, want key-present literal null
    todo_identity_red_test.go:223: SQLite archive: card_uuid key absent, want key-present literal null
    todo_identity_red_test.go:230: AC-TID-002 executed SQLite=1 live=1 archive=1
--- FAIL: TestTodoIdentityAC002LegacySQLiteExplicitNullNoMutation
=== RUN   TestTodoIdentityAC002LegacyActiveWALExplicitNullNoMutation
    todo_identity_red_test.go:248: WAL project: project_uuid key absent, want key-present literal null
    todo_identity_red_test.go:249: WAL live: card_uuid key absent, want key-present literal null
    todo_identity_red_test.go:250: WAL archive: card_uuid key absent, want key-present literal null
    todo_identity_red_test.go:254: AC-TID-002 executed active_WAL=1 committed_snapshot=1 live=1 archive=1
--- FAIL: TestTodoIdentityAC002LegacyActiveWALExplicitNullNoMutation
FAIL
```

RED reason: 모든 매체에서 key가 생략된다. JSON bytes/DB 미생성, SQLite main bytes/schema, 활성 WAL의 main DB/WAL bytes 불변 검사는 실패하지 않아 pure-read 보존 전제는 실행됐다.

### AC-TID-003 — UUID entropy 실패와 side-table backfill fault rollback

Selector: `^TestTodoIdentityAC003`; top-level test 2, entropy fault 1, backfill trigger 1. Exit 1.

```text
=== RUN   TestTodoIdentityAC003UUIDEntropyFailureRollsBack
    todo_identity_red_test.go:277: AC-TID-003 entropy fault reached=0 writer_error=<nil>
    todo_identity_red_test.go:279: UUIDv7 entropy seam reached=0 want=1
    todo_identity_red_test.go:282: UUIDv7 entropy failure ignored
    todo_identity_red_test.go:285: entropy failure left partial change
--- FAIL: TestTodoIdentityAC003UUIDEntropyFailureRollsBack
=== RUN   TestTodoIdentityAC003BackfillFaultRollsBack
    todo_identity_red_test.go:314: AC-TID-003 backfill fault reached=0 writer_error=<nil> identity_rows=0
    todo_identity_red_test.go:316: writer missed backfill fault: reached=0 err=<nil>
    todo_identity_red_test.go:319: rollback failed: record_equal=false identity_rows=0
--- FAIL: TestTodoIdentityAC003BackfillFaultRollsBack
FAIL
```

RED reason: writer가 UUIDv7 generator와 precreated `todo_identities` hook을 호출하지 않아 fault reached가 각각 0이며 카드 쓰기를 커밋한다. entropy test는 selector 내부에 parallel test가 없고 `uuid.SetRand(nil)`을 defer로 복구한다. trigger는 `identity-fault-reached` marker로 실제 도달을 판별한다.

### AC-TID-004 — Add 반환 equality와 lifecycle 안정성

Selector: `^TestTodoIdentityAC004`; top-level test 1, mutate/archive/restore/reopen 각 1. Exit 1.

```text
=== RUN   TestTodoIdentityAC004LifecycleStability
    todo_identity_red_test.go:334: initial project: missing project_uuid after successful writer
    todo_identity_red_test.go:335: Add return: missing card_uuid after successful writer
    todo_identity_red_test.go:352: archive: missing card_uuid after successful writer
    todo_identity_red_test.go:360: restore: missing card_uuid after successful writer
    todo_identity_red_test.go:361: reopen project: missing project_uuid after successful writer
    todo_identity_red_test.go:368: AC-TID-004 executed mutate=1 archive=1 restore=1 reopen=1
--- FAIL: TestTodoIdentityAC004LifecycleStability
FAIL
```

RED reason: text/state/spec lifecycle 보존은 실행됐지만 비교할 UUID가 발급·projection되지 않는다.

### AC-TID-005 — live/archive runtime link와 missing-card 변경 0

Selector: `^TestTodoIdentityAC005`; top-level test 2, semantic RED 1, regression PASS 1. Exit 1.

```text
=== RUN   TestTodoIdentityAC005RuntimeLinksLiveAndArchived
    todo_identity_red_test.go:398: project: missing project_uuid after successful writer
    todo_identity_red_test.go:399: live: missing card_uuid after successful writer
    todo_identity_red_test.go:400: archive: missing card_uuid after successful writer
    todo_identity_red_test.go:412: run: missing project_uuid after successful writer
    todo_identity_red_test.go:419: assignment t1: missing project_uuid after successful writer
    todo_identity_red_test.go:420: assignment t1: missing card_uuid after successful writer
    todo_identity_red_test.go:419: assignment t2: missing project_uuid after successful writer
    todo_identity_red_test.go:420: assignment t2: missing card_uuid after successful writer
    todo_identity_red_test.go:428: AC-TID-005 executed runs=1 assignments=2 live=1 archive=1
--- FAIL: TestTodoIdentityAC005RuntimeLinksLiveAndArchived
=== RUN   TestTodoIdentityAC005MissingRuntimeCardMutationZero
    todo_identity_red_test.go:449: AC-TID-005 regression guard executed missing=1 mutation_zero=1
--- PASS: TestTodoIdentityAC005MissingRuntimeCardMutationZero
FAIL
```

RED reason: 기존 runtime run/assignment와 archived-card lookup은 성공하지만 UUID linkage projection이 없다. missing card는 오류를 반환하고 공개 record가 동일하게 남는 기존 보존 guard가 PASS했다.

### AC-TID-006 — distinct handles, future identity version, retired

Selector: `^TestTodoIdentityAC006`; top-level test 3, subtest 2, distinct store handle 2, semantic RED 2, regression PASS 1. Exit 1.

```text
=== RUN   TestTodoIdentityAC006ConcurrentDistinctHandles
    todo_identity_red_test.go:483: concurrent Add: missing card_uuid after successful writer
    todo_identity_red_test.go:483: concurrent Add: missing card_uuid after successful writer
    todo_identity_red_test.go:495: concurrent project: missing project_uuid after successful writer
    todo_identity_red_test.go:497: stored t1: missing card_uuid after successful writer
    todo_identity_red_test.go:497: stored t2: missing card_uuid after successful writer
    todo_identity_red_test.go:500: conservation success=2 items=2 last_seq=2 project="" uuids=0
    todo_identity_red_test.go:502: AC-TID-006 executed handles=2 barrier=1 successes=2 cards=2 UUIDs=0
--- FAIL: TestTodoIdentityAC006ConcurrentDistinctHandles
=== RUN   TestTodoIdentityAC006FutureIdentityVersionFailClosed
=== RUN   TestTodoIdentityAC006FutureIdentityVersionFailClosed/LoadPure
    todo_identity_red_test.go:513: LoadPure accepted future identity version
=== RUN   TestTodoIdentityAC006FutureIdentityVersionFailClosed/writer
    todo_identity_red_test.go:540: writer accepted/mutated future identity version err=<nil> unchanged=false before=1/1/0 after=2/2/0
=== NAME  TestTodoIdentityAC006FutureIdentityVersionFailClosed
    todo_identity_red_test.go:544: AC-TID-006 executed future_version operations=2
--- FAIL: TestTodoIdentityAC006FutureIdentityVersionFailClosed
=== RUN   TestTodoIdentityAC006RetiredRegression
    todo_identity_red_test.go:569: AC-TID-006 retired regression executed writers=2 mutation_zero=1
--- PASS: TestTodoIdentityAC006RetiredRegression
FAIL
```

RED reason: 두 독립 BacklogStore handle의 동시 Add는 카드/tNN을 보존하지만 UUID가 0개다. reader와 writer 모두 `identity_schema_version=999`를 거절하지 않고 writer는 item/last_seq를 변경한다. retired Add/runtime writer의 오류와 변경 0은 PASS했다.

별도 race invocation:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test -race ./internal/kanban -run '^TestTodoIdentityAC006ConcurrentDistinctHandles$' -count=1 -v -timeout=120s
```

Exit 1. `WARNING: DATA RACE`는 출력되지 않았지만 UUID semantic assertion이 실패했으므로 race PASS로 분류하지 않는다.

```text
=== RUN   TestTodoIdentityAC006ConcurrentDistinctHandles
    todo_identity_red_test.go:500: conservation success=2 items=2 last_seq=2 project="" uuids=0
    todo_identity_red_test.go:502: AC-TID-006 executed handles=2 barrier=1 successes=2 cards=2 UUIDs=0
--- FAIL: TestTodoIdentityAC006ConcurrentDistinctHandles
FAIL
```

### AC-TID-007 — UUID timestamp 비권위

Selector: `^TestTodoIdentityAC007`; top-level test 1, 실제 UUIDv7 2, 기존 authority field 5. Exit 1.

```text
=== RUN   TestTodoIdentityAC007TimestampNonAuthority
    todo_identity_red_test.go:606: later UUID card: missing card_uuid after successful writer
    todo_identity_red_test.go:607: earlier UUID card: missing card_uuid after successful writer
    todo_identity_red_test.go:609: projection got="","" want="01890f3e-8b01-7000-8000-000000000002","01890f3e-8b00-7000-8000-000000000001"
    todo_identity_red_test.go:614: AC-TID-007 executed UUIDv7_timestamps=2 authority_fields=added_at,state,reported_state
--- FAIL: TestTodoIdentityAC007TimestampNonAuthority
FAIL
```

RED reason: 실제 기존 `AddedAt`, `State`, `ReportedState`, `LoadPure`, `RecordFactoryCardState`만 사용했다. UUID timestamp를 달리한 side-table rows가 공개 카드에 projection되지 않는다. receipt/approval/owner-token subsystem은 만들거나 흉내 내지 않았다.

### DELTA-ID-03 — iteration 2 D10–D13 보강

동일 HEAD에서 7개 AC selector를 다시 독립 실행했다. top-level 실행 수는 `1/3/3/1/3/3/1`, 합계 15이며 모두 exit 1이다. 추가된 assertion도 compile/setup/timeout/zero-test 실패가 아니라 현재 누락 동작에 도달했다.

#### D10 cross-project uniqueness/cardinality

Git·비Git subtest에서 얻은 project UUID와 모든 card UUID를 외부 collection에 모아 project 2, card 4, 전역 unique 6을 검사한다. 현재 출력:

```text
todo_identity_red_test.go:204: cross-project UUID cardinality: projects=0 cards=0 globally_unique=0, want 2/4/6
todo_identity_red_test.go:206: AC-TID-001 executed fixtures=2 git=1 non_git=1 cards=4
```

runtime start/assignment clause는 문서에서 AC-005로 이동하므로 AC-001에 중복 호출을 추가하지 않았다.

#### D11 fault 제거 후 retry·반복 writer·schema/stamp rollback

Entropy fault를 제거한 같은 fixture에서 Add 재시도 후 non-null UUIDv7을 검사하고, Mutate 반복 writer 뒤 같은 project/card identity를 비교한다. Backfill trigger도 제거한 같은 live/archive fixture에서 성공 재시도, 기존 전체 카드 backfill, 반복 writer identity 보존을 검사한다.

identity DDL이 전혀 없는 legacy DB에는 기존 `meta` table의 `identity_schema_version` INSERT만 차단하는 trigger를 설치했다. 구현이 identity table을 만든 뒤 stamp를 쓰다가 실패하면 같은 transaction이 table/stamp/rows/record를 전부 원복해야 한다. 이 방식은 production seam 추가 없이 실제 SQLite DDL/stamp 실패 지점에 도달 가능하다. 현재 writer는 identity schema를 시도하지 않으므로 다음과 같이 RED다.

```text
AC-TID-003 entropy fault reached=0 writer_error=<nil>
AC-TID-003 entropy fault retry=1 repeated_writer=1
AC-TID-003 backfill fault reached=0 writer_error=<nil> identity_rows=0
AC-TID-003 backfill fault retry=1 repeated_writer=1
AC-TID-003 schema fault reached=0 writer_error=<nil> table=0 stamp=0 rows=0
schema/stamp failure was not atomic: reached=0 err=<nil> table=0 stamp=0 rows=0 record_equal=false
successful retry did not create identity schema: table=0 stamp=0
AC-TID-003 schema fault retry=1
```

#### D12 실제 ambiguous 후보와 state authority

`items`와 `archived_items`는 각각 내부 UNIQUE만 가지므로 raw fixture에서 같은 local card ID를 두 table에 한 번씩 두어 실제 후보 count 2를 만들 수 있다. 현재 `recordRuntime`의 `SELECT EXISTS(... UNION ALL ...)`은 후보 수를 구분하지 않고 assignment를 기록하므로 정확한 RED가 발생한다.

```text
AC-TID-005 ambiguous candidates=2 writer_error=<nil> mutation_zero=false
ambiguous live/archive card must error/change-zero: err=<nil> unchanged=false assignments=0/1
```

동일 selector에서 실제 `RecordFactoryCardState(..., "completed", "card.completed")`를 호출하고 `ReportedState="completed"`만 반영되며 live `BacklogItem.State="queued"`가 유지되는 기존 권위 경계도 검사한다.

```text
AC-TID-005 executed runs=1 assignments=2 live=1 archive=1 state_updates=1
```

#### D13 OwnerLabel·ProvenanceJSON deep equality

`BacklogItem`에는 owner/provenance가 없고 실제 carrier는 `TodoRuntimeAssignment.OwnerLabel`과 `ProvenanceJSON`이다. 실제 SPEC 파일을 fixture에 생성해 `SpecID`, `SpecPath`, `SpecSHA256`, `CapturedAt`이 채워진 provenance를 `RecordFactoryCardState`로 기록했다. state 호출 후 snapshot을 고정하고 역순 UUIDv7 side-table rows만 삽입한 다음 전체 `TodoRuntime`을 `reflect.DeepEqual`로 비교한다. 공개 projection은 여전히 누락되어 RED지만 owner/provenance deep equality와 state 비권위 assertion은 실행됐다.

```text
projection got="","" want="01890f3e-8b01-7000-8000-000000000002","01890f3e-8b00-7000-8000-000000000001"
AC-TID-007 executed UUIDv7_timestamps=2 authority_fields=added_at,state,reported_state,owner_label,provenance_json
```

### 분리한 회귀 PASS

```text
=== RUN   TestTodoIdentityAC005MissingRuntimeCardMutationZero
    todo_identity_red_test.go:449: AC-TID-005 regression guard executed missing=1 mutation_zero=1
--- PASS: TestTodoIdentityAC005MissingRuntimeCardMutationZero
=== RUN   TestTodoIdentityAC006RetiredRegression
    todo_identity_red_test.go:569: AC-TID-006 retired regression executed writers=2 mutation_zero=1
--- PASS: TestTodoIdentityAC006RetiredRegression
PASS
ok  github.com/modu-ai/moai-adk/internal/kanban
```

## Baseline-attribution

현재 tree에서 측정한 값:

```text
HEAD=a315dad9af0d3a0e04862e6106b3993d9a3812f7
BRANCH=WT-todo-unified
TEST_SHA256=3c0e6cd4e581c02a6cbeb5d2991d1256738863ef98133cfe1b55595abc13699f
AC_TEST_COUNT=15
HISTORICAL_SELECTOR_COUNT=4
```

첫 감사에 기록된 이전 test SHA256은 `b667cf6c4025e895fdb8b020ec722dd21f71f2a935c5c342404165e3e50a64e9`였다. 기존 selector 이름 4개는 wrapper로 보존했고, 이번 AC-specific tests가 그 계약을 강화한다.

Formatting/boundary command:

```bash
gofmt -w internal/kanban/todo_identity_red_test.go
git diff --check -- internal/kanban/todo_identity_red_test.go
```

Observed exit: 0, stdout 없음.

## Gaps

- 이 작업은 RED-only다. 생산 코드, SPEC body, DB/card state, git history를 변경하지 않았다.
- backup 파일 복원 hash는 아직 별도 runnable test가 아니다. 논리 transaction rollback의 record/row 0만 RED로 고정했다.
- cross-process subprocess 경쟁은 실행하지 않았다. AC-006의 최소 경계인 서로 다른 BacklogStore/SQLite handle 2개와 barrier를 실행했다.
- race invocation은 semantic assertion이 먼저 실패하므로 race suite PASS가 아니다.
- side-table DDL은 RED fixture 계약이다. migration 구현과 downgrade/backup 절차 검증은 GREEN 단계 책임이다.

## Residual-risk

- `uuid.SetRand`는 package-global seam이다. 이 selector는 parallel test를 사용하지 않고 defer로 복구하지만, 향후 동일 프로세스에서 병렬 identity test를 추가하면 별도 process 격리로 바꿔야 한다.
- active WAL test는 main DB와 WAL bytes를 비교한다. SQLite의 lock coordination 파일인 SHM bytes는 read lock으로 달라질 수 있어 불변 hash 대상에서 제외했다.
- trigger 도달 count는 오류 marker 관측으로 0/1을 계산한다. production이 다른 저장 순서를 택하더라도 marker를 반환하고 전체 rollback해야 한다.
- 기존 전체 kanban suite와 CI는 이 RED 단계에서 PASS 대상이 아니다. GREEN 이후 change-scoped suite, race, coverage, lint, 독립 sync audit이 필요하다.
