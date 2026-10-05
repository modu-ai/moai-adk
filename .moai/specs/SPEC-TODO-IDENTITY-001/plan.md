# Plan — SPEC-TODO-IDENTITY-001

## A. 승인·현재 상태

DELTA-ID-03: iteration-2 D10–D13 계약 수정본. 최신 RED 증거 반영 완료. Tier M plan audit ceiling=2가 소진되었다. iteration-2 FAIL 0.75 이후 D10–D13 post-audit repair complete이며 independent re-audit unavailable이다. Explicit informed user override 및 formal Implementation Kickoff Approval approved. audit_verdict는 BYPASSED이며 감사 FAIL 원문은 보존한다. 구현/PASS/in-progress가 아니다. 사용자는 Tier M/UUIDv7/명시 null을 선택하고 구현을 요청했다. Formal post-audit Implementation Kickoff Approval approved (user override; BYPASSED). 독립 감사 FAIL은 그대로이며 명시적 informed override와 정식 kickoff가 승인되었다. run 준비 후 manager-develop만 in-progress 전이를 수행한다. 첫 storage의 검증 순서 예외는 이 child에 적용하지 않는다.

## B. 확정 저장 계약과 경계

기존 github.com/google/uuid v1.6.0의 UUIDv7 생성기를 재사용한다. canonical lowercase/nonzero/version7을 검증하고 오류는 반환한다. ULID 도입 금지. tNN은 표시·기존 local ID이며 UUID timestamp는 정렬/index locality 힌트일 뿐 어떤 업무 권위도 아니다.

현재 RED fixture와 동일한 최소 DDL을 사용한다. CHECK는 해당 fixture에 없으므로 추가 전제로 삼지 않는다. entity_kind 허용값 project/card는 writer 및 schema/row 검증에서 검사한다.

```sql
CREATE TABLE todo_identities(
  entity_kind TEXT NOT NULL,
  local_id TEXT NOT NULL,
  uuid TEXT NOT NULL UNIQUE,
  PRIMARY KEY(entity_kind,local_id)
);
INSERT INTO meta(key,value) VALUES('identity_schema_version','1');
```

프로젝트 행은 entity_kind='project', local_id='project'라는 단일 상수다. 카드 행은 entity_kind='card', local_id=기존 card.ID(tNN)다. 경로 column/path hash를 발급 ID로 사용하지 않는다. core meta.schema_version='1'과 meta.runtime_schema_version='1'은 변경하지 않는다. identity 표/표식 모두 없는 상태만 legacy다. 한쪽만 존재하거나 version이 unknown/future이거나 필수 열/제약이 누락된 partial schema는 reader/writer 모두 오류로 거절한다. malformed UUID, 중복 live/archive local ID와 모호한 연결은 추정/재발급하지 않는다.

DDL·기존 live/archive backfill·identity_schema_version stamp·카드/기존 runtime 연결 쓰기는 승인 writer의 같은 *sql.Tx 안에서 실행한다. 실패 시 전부 rollback한다. 모든 writer는 기존 BoardLock seam을 재사용한다. MaxOpenConns(1)을 유지하며 transaction 중 e.db.Query/Exec로 재진입하지 않고 tx.Query/Exec만 쓴다. schema 판정도 같은 writer transaction의 snapshot에서 수행한다. reader는 schema/row를 검사만 하며 발급·DDL·stamp를 수행하지 않는다.

LoadPure 공개 JSON은 project_uuid, live item.card_uuid, archived.item.card_uuid를 미발급이면 key-present null로 직렬화한다. runtime runs.project_uuid와 assignments.project_uuid/card_uuid도 실제 동일 snapshot을 따른다. Add 반환 item은 commit 성공 시 저장된 UUID와 같아야 한다.

기존 RecordFactoryRunStart/RecordFactoryCardAssignment/RecordFactoryCardState를 유지한다. live 또는 archive에서 단 하나의 카드만 연결하며 missing/ambiguous 거절 시 run/card/identity 추가 0이다. run_uuid alias 전환과 owner-token은 제외한다. whole-record writer의 필수 순서는 Add→runtime 기록→Mutate(text/state/spec)→ArchiveCard→runtime state 기록→LoadPure→RestoreCard→재개방/LoadPure이며 각 단계에서 같은 project/card UUID와 기존 runtime provenance를 비교한다. 현재 AC004/005가 각각 수행하는 부분과 이 결합 sequence 전체의 추가 검증을 구분한다.

## C. 매체별 migration·rollback 관측

- 기존 SQLite: 실제 Todo DB 하나의 transaction이 원자적 대상이다. identity와 기존 live/archive/runtime/schema stamp가 모두 같은 commit에 속한다. 오류 후 공개 record deep equality, identity 행 증감 0, schema/stamp 전후 equality를 검사한다.
- legacy JSON: 외부 원본 JSON은 SQL transaction의 일부가 아니다. 기존 atomic publication/quarantine 흐름을 보존하고 staging SQLite 안에서 identity를 함께 구성·commit한 뒤 검증 성공시에만 기존 흐름으로 게시한다. staging 실패는 게시하지 않고 원본 JSON을 그대로 둔다. 원본 이동/격리 시점은 기존 게시 계약 이후이며 identity 작업이 선행 삭제를 추가하지 않는다.
- pure read/future 거절 fixture의 byte 집합: JSON 원본, main SQLite 파일, 실행 전 존재한 WAL 파일의 bytes와 존재 여부를 각각 기록한다. 활성 WAL의 committed snapshot을 읽되 검사를 위해 checkpoint하지 않는다. SHM은 SQLite의 공유 메모리 조정 파일이므로 byte hash 동일성 대상이 아니며 row/schema의 권위로 쓰지 않는다.
- 정상 writer가 시작한 뒤 rollback의 기준은 transaction record/row/schema 불변이다. SQLite header/WAL의 물리 표현 변경 가능성과 구분해 임의 byte equality를 논리 rollback 증명으로 사용하지 않는다. future/partial 거절은 쓰기 전 판정하여 영속 payload 변경 0을 요구한다.
- 테스트 manifest는 대상 경로(main/WAL/JSON), 존재 여부, hash, schema/rows snapshot을 보존한다. backup-file restore 실행 테스트는 현재 없다. 이는 child transaction rollback AC 밖의 비차단 후속 작업이며 운영 migration 전에 별도 검증해야 한다. 운영 home/설치/DB 전환은 여기서 하지 않는다.

## D. AC별 실행 계약

각 명령은 아래 공통 prefix의 <selector>를 표의 literal로 치환한 단일 invocation이다. 실행수 0/compile/setup/timeout 오류는 RED가 아니다.

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '<selector>' -count=1 -v -timeout=120s
```

| AC | 정확 selector | top-level 최소 수 | GREEN에서 필요한 증거 |
|---|---|---:|---|
|001|^TestTodoIdentityAC001|1|Git/비Git 2fixture, 카드4개, UUIDv7/nonzero/lowercase/불충돌, Add=LoadPure, tNN/last_seq|
|002|^TestTodoIdentityAC002|3|JSON/SQLite/active WAL 각각 live/archive null key 존재, 비변경|
|003|^TestTodoIdentityAC003|3|entropy/trigger 각각 fault reached>=1, error, changed=0, identity rows 증가0, schema stamp 부분 반영0; fault 제거 재시도/반복 보존|
|004|^TestTodoIdentityAC004|1|Mutate/archive/restore/reopen identity와 내용 보존|
|005|^TestTodoIdentityAC005|3|run1/assignment2/live1/archive1 실제 연결 equality, missing 변경0; 모호함/결합 sequence 보강|
|006|^TestTodoIdentityAC006|3|별도 store/DB handles2 barrier1 successes2 cards2 last_seq2 UUID2, future reader/writer2 거절, retired writers2 변경0|
|007|^TestTodoIdentityAC007|1|역순 UUIDv7 projection과 AddedAt/State/ReportedState 보존, 기존 OwnerLabel/provenance 비교 보강|

AC003 generator는 uuid.SetRand 실패 reader를 테스트 종료 시 반드시 복원한다. 해당 global seam 테스트를 parallel로 실행하지 않는다. trigger는 실제 fixture todo_identities의 BEFORE INSERT RAISE(ABORT)를 사용한다. 현재 reached=0 RED는 writer가 요구된 발급/쓰기 지점에 진입하지 않음을 증명할 뿐, rollback 성공을 입증하지 않는다. GREEN에는 도달 횟수/오류/record 변경0/row count를 함께 요구한다. DDL 없는 fixture의 meta stamp trigger 오류·표/stamp/record rollback assertion과 오류 제거 재시도도 보강되었다. 현재 reached=0이므로 rollback 성공은 아직 아니다.

경쟁 최소 경계는 동일 process의 서로 다른 BacklogStore/SQLite handles 두 개다. barrier에서 동시 시작한다. goroutine race detector만으로 SQLite lost update를 판정하지 않는다. 별도 subprocess 경쟁/취소/잠금 timeout을 이 AC의 현재 실측이라고 주장하지 않는다. process-local 보조 명령:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test -race ./internal/kanban -run '^TestTodoIdentityAC006ConcurrentDistinctHandles$' -count=1 -v -timeout=120s
```

기존 core/runtime future·retired 회귀 연결(각 함수가 존재하며 실행 결과는 별도 기록해야 한다):

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^(TestTodoRuntimeStoreFutureSchemaPreservesBytes|TestTodoRuntimeStoreWriterRejectsFutureTodoVersions|TestTodoRuntimeSafetyRetiredAndMissingCardRefuse)$' -count=1 -v -timeout=120s
```

## D.1 DELTA-ID-03 — iteration-2 검증 계약

D10: AC001은 Add/LoadPure만 사용한다. 두 프로젝트의 project UUID와 카드 4개 UUID를 subtest 밖 공통 collection에 모아 project=2/card=4/union=6을 검증한다. runtime 호출은 AC005에서만 담당한다.

D11: 기존 entropy/insert fault fixture는 오류를 해제한 후 같은 store에서 재시도하고, 그 반환/LoadPure UUID와 후속 writer 이후 UUID를 비교한다. 별도 fixture는 todo_identities와 identity_schema_version이 모두 없는 지원 DB로 시작한다. meta에 BEFORE INSERT trigger를 설치하여 NEW.key='identity_schema_version'일 때만 RAISE(ABORT)하도록 한다. writer 실패 후 identity 표 없음·stamp 없음·record/기존 runtime 동일을 확인한다. trigger 제거 후 재시도에서 표/표식 생성과 non-null identity, 반복 보존을 검증한다. trigger 그 자체는 fixture의 baseline에 포함한다. 최신 AC003 selector는 entropy/backfill/schema-creation 3개를 실행했다. 정확 원문과 hash는 E에 귀속한다.

D12: 실제 DDL의 items.id/archived_items.id는 각 표 내부 UNIQUE이며 표 간 중복은 표현 가능하다. 정상 fixture의 카드 행을 다른 표에 같은 id로 구성해 후보 count=2로 만들고 실제 runtime API 오류와 전후 record/runtime equality를 요구한다. 정상 fixture에서는 completed state API 호출→LoadPure→whole-record Mutate→LoadPure 순으로 카드 State 불변·assignment ReportedState='completed'·identity 연결 equality를 확인한다.

D13: BacklogItem에는 AddedAt/State가 있고 TodoRuntimeAssignment에는 OwnerLabel/ProvenanceJSON/ReportedState가 있다. state API는 provenance.CapturedAt을 새로 캡처하므로 호출 전후 provenance를 무조건 같다고 요구하지 않는다. state API 성공 후 snapshot을 고정한 다음 UUID-only fixture 변경과 pure read 전후 해당 필드를 deep equality로 비교한다. nonexistent receipt/approval/owner-token은 만들지 않는다.

## E. DELTA-ID-03 최신 RED 증거

[RED baseline](../../reports/SPEC-TODO-IDENTITY-001/red-baseline.md)의 보강 실행 원문을 인용한다. HEAD=a315dad9af0d3a0e04862e6106b3993d9a3812f7, branch=WT-todo-unified. TEST_SHA256=3c0e6cd4e581c02a6cbeb5d2991d1256738863ef98133cfe1b55595abc13699f. REPORT_SHA256=e693d26e394eca8fdc18011a87d63e3bdb0c3f801a66d76b79ea2a41970328ce. 문서 작성자가 hash를 직접 재확인했으며 테스트 실행은 보고서 작성자의 실측이다.

D의 공통 command와 정확 AC001…007 selector가 각각 1/3/3/1/3/3/1개, 총 15개를 실행했고 모두 exit 1 semantic RED였다. AC005 missing과 AC006 retired는 별도 회귀 PASS다. race 명령도 exit 1 semantic RED이며 DATA RACE 미관측은 PASS가 아니다. 보고서 앞부분의 이전 line-number/2-test 발췌는 역사적 기록이며 현재 count/hash는 DELTA-ID-03과 Baseline-attribution이 기준이다.

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


## F. 변경 경계와 MX 계획

생산 편집 후보는 internal/kanban/backlog_store.go, backlog_sqlite.go 및 기존 read/write/migrate 분할파일의 identity 경계, todo_runtime.go/factory_runtime.go의 연결 부분이다. 새로운 광역 refactor나 후속 owner/completion 기능을 추가하지 않는다.

기존 MX를 제거하지 않고 아래 식별 marker를 해당 함수 직전 설명에 추가한다. 기대 추가 marker는 NOTE 3개/WARN 1개/REASON 1개, 합계 5개다. 이는 계획이며 현재 생산 코드에 있다는 주장이 아니다.

| 파일/함수 anchor | 정확 marker | 설명 |
|---|---|---|
|backlog_store.go / BacklogStore.LoadPure|@MX:NOTE: [TID:PURE]|미발급 null·비발급·schema 검사용 read|
|backlog_store.go / BacklogStore.Add|@MX:NOTE: [TID:RETURN]|commit identity와 반환값 일치|
|backlog_store.go / BacklogStore.Mutate|@MX:WARN: [TID:TX]|identity/card/schema 동일 tx|
|backlog_store.go / BacklogStore.Mutate|@MX:REASON: [TID:TX]|MaxOpenConns(1)에서 tx 외 재진입 금지·rollback|
|todo_runtime.go / BacklogStore.recordRuntime|@MX:NOTE: [TID:LINK]|실제 live/archive identity·provenance 보존|

구현 후 아래 bounded 검사는 정확 5줄을 반환해야 하며 함수 위치도 readback한다. tag 수는 동작 검증을 대신하지 않는다.

```bash
rg -n '@MX:(NOTE|WARN|REASON): \[TID:(PURE|RETURN|TX|LINK)\]' internal/kanban/backlog_store.go internal/kanban/todo_runtime.go
```

## G. 순서·완료 기준·잔여 위험

High: explicit informed user override 및 formal Implementation Kickoff Approval→각 변경 직전 해당 RED→구현→fault 도달 및 실제 rollback/재시도·결합 sequence·schema 거절 보강→전체 7AC GREEN/관련 회귀→변경 로직 coverage85% 이상/독립 sync 감사. 기존 umbrella23AC 완료 의무는 유지한다.

D10 cardinality, D11 retry/반복 writer/stamp failure, D12 ambiguous/state update, D13 OwnerLabel/ProvenanceJSON deep equality의 assertion 누락은 보강되었다. 현재 결과는 RED이며 발급 후 version/lifetime/link equality와 실제 fault 도달 뒤 rollback GREEN은 아직 없다. partial identity schema fixture와 결합 sequence 전체의 추가 회귀는 구현 검증 때 확인해야 하며 이번 보강을 그 실측으로 확대하지 않는다. 운영 backup-file restore와 별도 process 경쟁은 미실행이다. no performance claim.

Tier M plan audit ceiling=2가 소진되었다. iteration-2 FAIL 0.75 이후 D10–D13 post-audit repair complete이며 independent re-audit unavailable이다. Explicit informed user override 및 formal Implementation Kickoff Approval approved. audit_verdict는 BYPASSED이며 감사 FAIL 원문은 보존한다. 구현/PASS/in-progress가 아니다.
