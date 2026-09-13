# SPEC-TODO-IDENTITY-001 독립 동기화 감사

- 감사 종류: post-implementation sync audit, iteration 1
- 감사일: 2026-09-12
- 활성 프로필: `default` (flat weighted-percentage)
- 최종 verdict: **FAIL**
- 가중 점수: **84/100** (`83.5` 반올림)
- 강제 실패 사유: must-pass 차원인 **Functionality 75/100**이 “7개 AC 전부 PASS” 임계값을 충족하지 못했다. AC-TID-006의 partial-schema fail-closed와 AC-TID-001의 전역 UUID 유일성에 차단 결함 F1이 있다.
- plan audit 상태: iteration-2 `FAIL 0.75` 뒤 사용자 informed override로 `BYPASSED`; 이 감사는 이를 PASS로 재해석하지 않는다.

## Claim

현재 dirty worktree에서 15개 identity AC 테스트, 6개 고정 회귀, 관련 CLI 범위, 전체 `internal/kanban`, AC006 race, 변경 identity 파일 커버리지, vet/build/Windows build/lint/SPEC lint/diff/MX/security 검사는 모두 exit 0이었다. 그러나 성공한 기존 테스트가 검사하지 않는 partial UNIQUE schema를 독립 SQLite fixture로 재현한 결과, production validator가 부분 인덱스를 전역 UUID UNIQUE 제약으로 오인한다. 그 schema는 카드 UUID 중복을 실제로 허용하므로 AC-TID-006과 AC-TID-001을 모두 만족한다고 주장할 수 없다.

따라서 구현은 **FAIL**이다. 이 보고서는 결함을 고치지 않았고 production/test/SPEC/DB/card/Git index/history를 변경하지 않았다.

### AC 판정표

| AC | 판정 | 독립 관측과 판단 |
|---|---|---|
| AC-TID-001 | **FAIL** | 정상 Git/비Git fixture의 2 project/4 card/union 6, UUIDv7, Add=LoadPure, t1/t2/last_seq=2는 PASS했다. 그러나 F1의 승인 가능한 partial schema에서는 서로 다른 카드가 같은 UUID를 가질 수 있어 전역 유일성 계약이 보장되지 않는다. |
| AC-TID-002 | PASS | legacy JSON/SQLite/active-WAL의 key-present `null`, main/WAL bytes 및 schema 불변이 PASS했다. 별도 동일-store 2회 pure-read/CLI idempotence 회귀도 PASS했다. |
| AC-TID-003 | PASS | entropy/backfill/stamp fault가 각각 실제 도달했고 rollback, 같은 fixture 재시도, DDL/meta/index readback이 PASS했다. identity DDL/stamp/card write가 한 `*sql.Tx` 안에 있고 runtime writer도 한 transaction 안에서 identity/runtime을 기록함을 소스와 fault suite로 확인했다. |
| AC-TID-004 | PASS | Add→Mutate(text/state/spec)→Archive→Restore→reopen 동안 project/card UUID와 내용·상태가 유지됐다. |
| AC-TID-005 | PASS | live/archive 실제 UUID 연결, missing/후보 2개 mutation-zero, `ReportedState` 비권위, whole-record edit 후 runtime 보존이 PASS했다. storage duplicate는 `ErrBacklogIDConflict`, runtime 후보 중복은 별도 ambiguity error로 유지된다. |
| AC-TID-006 | **FAIL** | 두 handle 동시 Add와 race, future identity/core/runtime 및 retired guard는 PASS했다. 하지만 `partial=1`인 uuid 단일열 UNIQUE 인덱스를 validator가 승인할 수 있고 fixture가 동일 card UUID 2행을 허용했다. partial/malformed schema fail-closed 계약 위반이다. |
| AC-TID-007 | PASS | 역순 UUIDv7 projection 후 AddedAt/State/ReportedState/OwnerLabel/ProvenanceJSON이 유지되고 UUID timestamp가 업무 권위로 승격되지 않았다. |

## Evaluation Report

SPEC: SPEC-TODO-IDENTITY-001
Overall Verdict: **FAIL — Functionality must-pass firewall (75/100; threshold: all AC PASS)**

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 75/100 | **FAIL** | Combined 15 tests와 전체 package는 PASS했지만, SQLite `:memory:` probe가 `partial=1`인 UNIQUE index 아래 동일 card UUID 2행을 허용했다. `identityUUIDUnique`는 `partial`을 읽고도 검사하지 않는다. |
| Security (25%) | 100/100 | PASS | SQL 값은 placeholder를 사용하고 동적 conflict 절은 내부 상수 두 값뿐이다. secret probe 무출력, `go mod verify` PASS, `govulncheck`는 호출 가능한 코드 취약점 0건이었다. F1은 현재 UUID가 업무/소유권 권위가 아니라는 AC007 경계를 지키므로 Security High로 분류하지 않고 Functionality 데이터 무결성 결함으로 분류했다. |
| Craft (20%) | 90/100 | PASS | identity selector 기준 package denominator 30.8%, 변경 핵심 파일 `todo_identity.go` 180/211 = 85.3%; vet와 golangci-lint가 clean이다. package 전체 30.8%로 오해하지 않으며, 이는 선택된 identity tests가 전체 kanban package를 분모로 계산한 수치다. |
| Consistency (15%) | 90/100 | PASS | host/Windows build, strict SPEC lint, `git diff --check`, exact-five MX가 PASS했다. F1의 index 검증 누락과 선택적 concurrency/test-hardening 위험 때문에 만점은 부여하지 않았다. |

가중치 계산: `75×0.40 + 100×0.25 + 90×0.20 + 90×0.15 = 83.5`. 점수와 무관하게 Functionality must-pass 실패가 전체 FAIL을 강제한다.

### TRUST 5

| 원칙 | 판정 | 근거 |
|---|---|---|
| Tested | **FAIL** | 기존 suite는 모두 green이고 변경 핵심 파일 85.3%지만, partial UNIQUE counterexample이 테스트 manifest에 없고 실제 계약 위반을 허용한다. |
| Readable | PASS | identity schema/read/write helper가 분리돼 있고 명칭·오류 경로가 의도를 드러낸다. |
| Unified | PASS | `gofmt` 전제를 검증하는 build/vet/lint/diff check가 clean이며 기존 `BacklogStore`, `BoardLock`, `ErrBacklogIDConflict` 패턴을 재사용한다. |
| Secured | PASS | parameterized SQL, source secret 무검출, 호출 경로 취약점 0건. UUID timestamp는 권위로 사용하지 않는다. |
| Trackable | PASS (local artifact) | SPEC/AC/MX 연결은 추적 가능하다. commit/push/PR은 권한 밖이며 NOT RUN이다. plan audit `BYPASSED`는 그대로 보존됐다. |

### Findings

- **F1 [High] [blocking] [confidence: 0.99]** `internal/kanban/todo_identity.go:121` — `identityUUIDUnique`가 `PRAGMA index_list`의 `partial` 값을 읽지만 버린 뒤 `unique==1`인 모든 인덱스를 후보로 삼고, `internal/kanban/todo_identity.go:155`에서 열이 `uuid` 하나면 전역 UNIQUE로 승인한다. SQLite fixture의 `uuid_project_only ... partial=1` 인덱스는 이 조건을 만족하면서 card UUID 중복 2행을 허용했다. 영향: malformed/partial identity schema가 LoadPure/writer에 의해 거절되지 않을 수 있고, 공개 card UUID 전역 유일성이 깨져 AC-TID-006과 AC-TID-001이 FAIL한다. 기존 `internal/kanban/todo_identity_red_test.go:793`의 6개 schema guard에는 partial UNIQUE 인덱스가 없다. Required fix: `partial != 0`인 인덱스를 전역 UNIQUE 후보에서 제외하고, `uuid TEXT NOT NULL UNIQUE`와 동등한 비부분 단일열 제약만 승인하라. `CREATE UNIQUE INDEX ... ON todo_identities(uuid) WHERE ...` fixture에서 reader와 writer가 모두 fail-closed하고 data/schema/bytes가 변하지 않는 회귀 테스트 및 동일 UUID 2-card counterexample을 추가하라.
- **F2 [Medium] [optional] [confidence: 0.78]** `internal/kanban/backlog_store.go:765` — `Add`는 `Mutate`가 반환하면서 `BoardLock`을 해제한 뒤 별도 `LoadPure`로 방금 추가한 카드를 찾는다. 그 사이 다른 writer가 해당 카드를 archive/remove하면 DB commit은 성공했는데 `Add`는 `added backlog item ... missing` 오류를 반환할 수 있는 제어 흐름이다. 현재 AC006은 동시 Add 두 건만 실행하며 이 interleaving을 강제하지 않아 실제 실패는 재현하지 않았다. 영향: caller가 실패로 판단해 재시도하면 이미 반영된 쓰기와 중복 의도가 생길 수 있다. Required fix: persisted UUID/result를 같은 locked mutation 경계 안에서 반환할 수 있게 최소 seam을 두거나, 결정적 barrier test로 현재 구조가 해당 interleaving에서도 계약을 지킴을 증명하라. 재현 전에는 F1 수정 범위에 자동 포함하지 않는 선택적 hardening이다.
- **F3 [Low] [optional] [confidence: 1.00]** `internal/kanban/todo_identity_red_test.go:215` — AC002의 JSON/SQLite/active-WAL 각 직접 selector는 fixture당 `LoadPure`를 한 번만 호출한다(`:215`, `:235`, `:262`). 동일 DB 2회 pure-read와 CLI 2회 idempotence는 다른 회귀 테스트가 보완하고 production 소스도 read-only지만, AC 문구의 각 legacy 매체 동일-fixture 반복을 한 selector가 직접 고정하지는 않는다. 영향: 향후 두 번째 호출에만 생기는 sidecar/state 회귀의 진단력이 낮다. Required fix: 각 AC002 fixture에서 두 번 읽고 두 결과의 JSON equality와 main/WAL/schema/existence 불변을 모두 검사하라.

### Recommendations

- 우선 F1만 차단 수정으로 처리하고 `identityUUIDUnique`의 `partial==0` 조건과 reader/writer mutation-zero 회귀를 추가하라.
- F1 수정 뒤 `^TestTodoIdentityAC006FutureIdentityVersionFailClosed$`, combined identity selector, AC006 race, 전체 `internal/kanban`, coverage/lint를 다시 실행하고 독립 delta re-audit를 요청하라.
- F2/F3은 선택적 보강이다. F1과 섞어 불필요한 광역 refactor를 만들지 말라.

## Evidence

### E1 — 기준선과 변경 범위

Command:

```bash
git fetch origin main 2>&1
git rev-parse HEAD
git branch --show-current
git rev-list --count --left-right origin/main...HEAD
git status --short -- .moai/reports/SPEC-TODO-IDENTITY-001 internal/kanban internal/cli/todo.go internal/cli/todo_analysis_test.go
```

Observed output:

```text
From https://github.com/modu-ai/moai-adk
 * branch                main       -> FETCH_HEAD
a315dad9af0d3a0e04862e6106b3993d9a3812f7
WT-todo-unified
0	3005
 M internal/cli/todo.go
 M internal/cli/todo_analysis_test.go
 M internal/kanban/backlog_archive_test.go
 M internal/kanban/backlog_integrity_audit_test.go
 M internal/kanban/backlog_migrate.go
 M internal/kanban/backlog_pure_reader_test.go
 M internal/kanban/backlog_schema_freeze_test.go
 M internal/kanban/backlog_sqlite.go
 M internal/kanban/backlog_store.go
 M internal/kanban/backlog_store_test.go
 M internal/kanban/factory_runtime.go
 M internal/kanban/factory_runtime_test.go
?? .moai/reports/SPEC-TODO-IDENTITY-001/
?? internal/kanban/todo_identity.go
?? internal/kanban/todo_identity_red_test.go
?? internal/kanban/todo_runtime.go
?? internal/kanban/todo_runtime_safety_test.go
?? internal/kanban/todo_runtime_store_test.go
?? internal/kanban/todo_unified_red_test.go
```

### E2 — identity selector inventory와 7 AC

Command:

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -list '^TestTodoIdentityAC' -count=1
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '^TestTodoIdentityAC' -count=1 -v -timeout=120s
```

Observed output (15 top-level tests, exit 0):

```text
TestTodoIdentityAC001GitAndNonGitUUIDv7
TestTodoIdentityAC002LegacyJSONExplicitNullNoMutation
TestTodoIdentityAC002LegacySQLiteExplicitNullNoMutation
TestTodoIdentityAC002LegacyActiveWALExplicitNullNoMutation
TestTodoIdentityAC003UUIDEntropyFailureRollsBack
TestTodoIdentityAC003BackfillFaultRollsBack
TestTodoIdentityAC003SchemaCreationFaultRollsBack
TestTodoIdentityAC004LifecycleStability
TestTodoIdentityAC005RuntimeLinksLiveAndArchived
TestTodoIdentityAC005MissingRuntimeCardMutationZero
TestTodoIdentityAC005AmbiguousLiveArchiveMutationZero
TestTodoIdentityAC006ConcurrentDistinctHandles
TestTodoIdentityAC006FutureIdentityVersionFailClosed
TestTodoIdentityAC006RetiredRegression
TestTodoIdentityAC007TimestampNonAuthority
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.606s
todo_identity_red_test.go:206: AC-TID-001 executed fixtures=2 git=1 non_git=1 cards=4
todo_identity_red_test.go:228: AC-TID-002 executed JSON=1 live=1 archive=1
todo_identity_red_test.go:248: AC-TID-002 executed SQLite=1 live=1 archive=1
todo_identity_red_test.go:272: AC-TID-002 executed active_WAL=1 committed_snapshot=1 live=1 archive=1
todo_identity_red_test.go:296: AC-TID-003 entropy fault reached=1 writer_error=issue project UUIDv7: unexpected EOF
todo_identity_red_test.go:328: AC-TID-003 entropy fault retry=1 repeated_writer=1
todo_identity_red_test.go:356: AC-TID-003 backfill fault reached=1 writer_error=insert identity: constraint failed: identity-fault-reached (1811) identity_rows=0
todo_identity_red_test.go:394: AC-TID-003 backfill fault retry=1 repeated_writer=1
todo_identity_red_test.go:429: AC-TID-003 schema fault reached=1 writer_error=stamp identity schema: constraint failed: identity-stamp-fault-reached (1811) table=0 stamp=0 rows=0
todo_identity_red_test.go:491: AC-TID-003 schema fault retry=1 ddl="CREATE TABLE todo_identities(\n  entity_kind TEXT NOT NULL,\n  local_id TEXT NOT NULL,\n  uuid TEXT NOT NULL UNIQUE,\n  PRIMARY KEY(entity_kind,local_id)\n)" meta=schema:1/runtime:absent/identity:1 indexes=[sqlite_autoindex_todo_identities_2:pk:unique=1 sqlite_autoindex_todo_identities_1:u:unique=1]
todo_identity_red_test.go:539: AC-TID-004 executed mutate=1 archive=1 restore=1 reopen=1
todo_identity_red_test.go:611: AC-TID-005 executed runs=1 assignments=2 live=1 archive=1 state_updates=1
todo_identity_red_test.go:632: AC-TID-005 regression guard executed missing=1 mutation_zero=1
todo_identity_red_test.go:692: AC-TID-005 ambiguous candidates=2 writer_error=runtime card "t1" is ambiguous (2 candidates) mutation_zero=true
todo_identity_red_test.go:751: AC-TID-006 executed handles=2 barrier=1 successes=2 cards=2 UUIDs=2
todo_identity_red_test.go:910: AC-TID-006 executed future_version operations=2 schema_guards=6 row_guards=4 preferred_guards=3
todo_identity_red_test.go:935: AC-TID-006 retired regression executed writers=2 mutation_zero=1
todo_identity_red_test.go:1027: AC-TID-007 executed UUIDv7_timestamps=2 authority_fields=added_at,state,reported_state,owner_label,provenance_json
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	10.231s
```

### E3 — F1 partial UNIQUE counterexample

Command:

```bash
sqlite3 ':memory:' "CREATE TABLE todo_identities(entity_kind TEXT NOT NULL,local_id TEXT NOT NULL,uuid TEXT NOT NULL,PRIMARY KEY(entity_kind,local_id)); CREATE UNIQUE INDEX uuid_project_only ON todo_identities(uuid) WHERE entity_kind='project'; PRAGMA index_list(todo_identities); PRAGMA index_info(uuid_project_only); INSERT INTO todo_identities VALUES('card','t1','01890f3e-8b00-7000-8000-000000000001'); INSERT INTO todo_identities VALUES('card','t2','01890f3e-8b00-7000-8000-000000000001'); SELECT entity_kind,local_id,uuid FROM todo_identities ORDER BY local_id;"
```

Observed output (exit 0):

```text
0|uuid_project_only|1|c|1
1|sqlite_autoindex_todo_identities_1|1|pk|0
0|2|uuid
card|t1|01890f3e-8b00-7000-8000-000000000001
card|t2|01890f3e-8b00-7000-8000-000000000001
```

Production source observation:

```text
121: var seq, unique, partial int
123: rows.Scan(&seq, &name, &unique, &origin, &partial)
127: if unique == 1 {
155: if len(names) == 1 && names[0] == "uuid" {
156:     return true, nil
```

`partial=1`을 거르는 조건이 없다. Test manifest의 schema cases는 `table-without-stamp`, `stamp-without-table`, `wrong-column-type`, `missing-column`, `wrong-column-order`, `uuid-without-unique` 여섯 개로, partial UNIQUE가 빠져 있다.

### E4 — 고정 회귀, runtime/core guard, CLI, 전체 package와 race

Commands:

```bash
env -u ... go test ./internal/kanban -run '^(TestBacklogArchive_PerItemContractFrozen|TestTodoHistoryAddsNoSchemaChange|TestTodoRuntimeSafetyRelocationPreservesRuntime|TestBacklogAdd_CreatesVersion1File|TestMigrationPartialFailureRemovesArtifacts|TestDuplicateIDRejectedByStorage)$' -count=1 -v -timeout=90s
env -u ... go test ./internal/kanban -run '^(TestTodoRuntimeStorePublicReadbackSurvivesCardEdit|TestTodoRuntimeStoreLegacyPureReadReturnsEmptyRuntimeWithoutWriting|TestTodoRuntimeStoreFutureSchemaPreservesBytes|TestTodoRuntimeStoreWriterRejectsFutureTodoVersions|TestTodoRuntimeStorePartialSchemaRefusedWithoutMutation|TestTodoRuntimeSafetyCardMutationAndAssignmentSerialize)$' -count=1 -v -timeout=120s
env -u ... go test ./internal/cli -run '^(TestTodoLegacyRecordRoundTrips|TestTodoList_JSONKeepsDroppedCards|TestTodoList_JSONIgnoresLimit|TestTodoList_JSONStructured|TestTodoList_EmptyQueueIsNotAnError|TestTodoList_DroppedHiddenByDefault|TestTodoList_DefaultBoundedWithWithheldNotice|TestTodoList_LockFreeWhileForeignProcessHoldsLock|TestTodoListJSONIsIdempotent|TestTodoReadSurface_.*|TestTodoDisclosure_LeavesBacklogJSONUntouched|TestResolveTodoQueueRoot_.*|TestTodoQueue_FallbackAdoptsExistingLocalQueue|TestTodoQueue_WorktreeSeesPrimaryQueue)$' -count=1 -v -timeout=180s
env -u ... go test ./internal/kanban -count=1 -timeout=300s
env -u ... go test -race ./internal/kanban -run '^TestTodoIdentityAC006ConcurrentDistinctHandles$' -count=1 -v -timeout=120s
```

Observed output:

```text
--- PASS: TestBacklogArchive_PerItemContractFrozen (0.00s)
--- PASS: TestTodoHistoryAddsNoSchemaChange (0.08s)
--- PASS: TestTodoRuntimeSafetyRelocationPreservesRuntime (0.53s)
--- PASS: TestBacklogAdd_CreatesVersion1File (0.04s)
--- PASS: TestDuplicateIDRejectedByStorage (0.05s)
--- PASS: TestMigrationPartialFailureRemovesArtifacts (0.07s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	1.363s

--- PASS: TestTodoRuntimeSafetyCardMutationAndAssignmentSerialize (0.16s)
--- PASS: TestTodoRuntimeStorePublicReadbackSurvivesCardEdit (1.41s)
--- PASS: TestTodoRuntimeStoreFutureSchemaPreservesBytes (0.00s)
--- PASS: TestTodoRuntimeStoreLegacyPureReadReturnsEmptyRuntimeWithoutWriting (0.09s)
--- PASS: TestTodoRuntimeStoreWriterRejectsFutureTodoVersions (0.43s)
--- PASS: TestTodoRuntimeStorePartialSchemaRefusedWithoutMutation (0.16s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	2.545s

PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	71.301s

ok  	github.com/modu-ai/moai-adk/internal/kanban	188.502s

=== RUN   TestTodoIdentityAC006ConcurrentDistinctHandles
    todo_identity_red_test.go:751: AC-TID-006 executed handles=2 barrier=1 successes=2 cards=2 UUIDs=2
--- PASS: TestTodoIdentityAC006ConcurrentDistinctHandles (0.41s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	2.494s
```

### E5 — coverage, craft, consistency, security

Commands:

```bash
env -u ... go test ./internal/kanban -run '^TestTodoIdentityAC' -count=1 -coverprofile=<ephemeral>/identity.cover -covermode=atomic -timeout=120s
go tool cover -func=<ephemeral>/identity.cover
go vet ./internal/kanban ./internal/cli
go build ./...
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...
timeout 120s golangci-lint run ./internal/kanban/... ./internal/cli/... --output.text.colors=false --output.text.print-issued-lines=false
moai spec lint .moai/specs/SPEC-TODO-IDENTITY-001 --strict --json
git diff --check
go mod verify
govulncheck ./internal/kanban ./internal/cli
```

Observed output:

```text
ok  	github.com/modu-ai/moai-adk/internal/kanban	9.455s	coverage: 30.8% of statements
todo_identity.go:41:  identitySchemaPresent    89.5%
todo_identity.go:69:  validateIdentityTable   84.4%
todo_identity.go:114: identityUUIDUnique       76.5%
todo_identity.go:162: validIdentityUUID        100.0%
todo_identity.go:167: readIdentitySnapshot     92.0%
todo_identity.go:205: apply                    100.0%
todo_identity.go:222: cloneIdentity            100.0%
todo_identity.go:230: identityCard              100.0%
todo_identity.go:239: ensureIdentitySchema      88.9%
todo_identity.go:254: ensureIdentity            87.5%
todo_identity.go:293: ensureRecordIdentities    91.3%
todo_identity.go:327: ensureStoredIdentities    66.7%
todo_identity.go aggregate covered=180 total=211 coverage=85.3%

go vet: [no stdout], exit 0
host go build: [no stdout], exit 0
CGO_ENABLED=0 Windows go build: [no stdout], exit 0
golangci-lint: 0 issues.
strict SPEC lint: []
git diff --check: [no stdout], exit 0
all modules verified
=== Symbol Results ===

No vulnerabilities found.

Your code is affected by 0 vulnerabilities.
This scan also found 0 vulnerabilities in packages you import and 3
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
Use '-show verbose' for more details.
```

Static contract readback:

```text
mx_tid_marker_count=5
physical_runtime_uuid_columns=0
todo list route: newTodoReadStore() -> LoadPure()
sql concat probe: internal/kanban/todo_runtime.go:212 only
secret probe: [no stdout]
```

line 212의 SQL concat은 외부 입력이 아니라 코드 내부에서 둘 중 하나로 선택되는 고정 `ON CONFLICT` clause이며, run/card/owner/provenance 값은 모두 `?` placeholder로 전달된다.

### 수정 파일 범위

감사한 identity/runtime production 범위:

- `internal/kanban/todo_identity.go`
- `internal/kanban/backlog_store.go`
- `internal/kanban/backlog_sqlite.go`
- `internal/kanban/backlog_migrate.go`
- `internal/kanban/todo_runtime.go`
- `internal/kanban/factory_runtime.go`
- `internal/cli/todo.go`

감사한 직접/관련 테스트 범위:

- `internal/kanban/todo_identity_red_test.go`
- `internal/kanban/todo_runtime_safety_test.go`
- `internal/kanban/todo_runtime_store_test.go`
- `internal/kanban/backlog_archive_test.go`
- `internal/kanban/backlog_integrity_audit_test.go`
- `internal/kanban/backlog_pure_reader_test.go`
- `internal/kanban/backlog_schema_freeze_test.go`
- `internal/kanban/backlog_store_test.go`
- `internal/kanban/factory_runtime_test.go`
- `internal/cli/todo_analysis_test.go` 및 선택된 CLI list/read/root 회귀

`internal/homestate/factory.go`, umbrella `todo_unified_red_test.go`, 다른 SPEC/report와 공유 dirty 파일은 이 child의 완료 주장에 포함하지 않았다.

## Baseline-attribution

- 실제 감사 트리: `.claude/worktrees/todo-unified`
- branch: `WT-todo-unified`
- HEAD: `a315dad9af0d3a0e04862e6106b3993d9a3812f7`
- fresh `origin/main...HEAD`: `0 3005`
- SPEC SHA-256: `63a1c141c5dc292436a22430cdbbc7196a9e0d3666adff604a74fe29c11b6284`
- plan SHA-256: `761b5568f38b2e9ea4b13f486977eda51a9b97fbf333e04065a29effd400ebbc`
- acceptance SHA-256: `36274298abdc8234fbb8c2db12339ced977cbc5551a3bd640a17b6b8add816f3`
- implementation report SHA-256: `fcfbffde4d50d3d83f66c4052de95c83153b1b33ebb1ef7858a667c9041da765`
- identity production SHA-256: `fa734b19416b6fd85c524c3bf86225c87dde0d19a51784b5c3e8abe2ceca58cd`
- identity test SHA-256: `747506b9c31e037967f611e26e9e03d8fb21f119912f92813ab7c4eb86c7c3e5`
- 모든 test/build/lint/probe는 위 HEAD와 이 시점 dirty tree에서 이 감사자가 실행했다. 과거 implementation 보고서의 장기 결과는 독립 PASS 근거로 대체 사용하지 않았다.
- SQLite counterexample은 운영 DB가 아닌 `:memory:` fixture였다. Todo DB, card state, Git index/history는 변경하지 않았다.

## Gaps

- cross-model `mcp__moai__audit_multi`는 동일 worktree `project_root`로 호출했으나 `timed out awaiting tools/call after 300s`로 종료됐다. 보조 backend verdict는 **INCONCLUSIVE**이며 local 기계 검증의 PASS나 FAIL로 계산하지 않았다.
- repository 전체 suite와 remote CI: NOT RUN. 요청된 변경 영향 범위, 전체 `internal/kanban`, 관련 CLI만 실행했다.
- commit/stage/push/PR/merge/deploy 및 운영 Todo DB migration/install/card-state 변경: NOT RUN, 권한 밖.
- full-package coverage profile은 실행하지 않았다. `30.8%`는 identity selector가 전체 kanban package를 분모로 계산한 수치이고, 변경 핵심 파일 aggregate는 `85.3%`다.
- F2의 interleaving은 기존 test seam으로 결정적 재현하지 않았다. 그러므로 확정 결함이나 FAIL 사유로 사용하지 않았다.
- `govulncheck`가 required modules에서 찾은 호출되지 않는 취약점 3건의 verbose 모듈/CVE 식별은 이번 범위에서 추가 실행하지 않았다. 현재 두 대상 package의 호출 가능한 취약점은 0건이다.

## Residual-risk

- F1 수정이 단순히 테스트 fixture만 거절하도록 하면서 다른 비부분 global UNIQUE 형태를 잘못 거절하거나, row-level duplicate 검증을 빠뜨릴 수 있다. schema guard와 duplicate row counterexample을 함께 유지해야 한다.
- `Add` post-commit readback 경쟁 창(F2)은 미재현 상태다. 별도 deterministic test가 없으면 향후 archive/remove와의 경쟁에서 부분 성공 오류 가능성이 남는다.
- legacy JSON/SQLite/active-WAL의 각 동일 fixture 반복-read는 F3처럼 직접 한 selector로 묶여 있지 않아, 관련 회귀들의 조합에 의존한다.
- package 전체 coverage와 remote CI는 미관측이다. local green은 통합 branch/운영 수용을 뜻하지 않는다.
- project/card UUID는 이 child에서 owner/completion/approval 권위가 아니다. 후속 owner token, generation, completion receipt/recovery, Graph/UI의 완료를 이 verdict로 주장할 수 없다.

## Iteration history

| Iteration | Date | Verdict | Score | Delta |
|---|---|---|---:|---|
| sync-audit 1 | 2026-09-12 | **FAIL** | 84/100 | 최초 독립 post-implementation 감사. F1 partial UNIQUE validator 결함을 운영 DB 밖 `:memory:` counterexample으로 확인. |
