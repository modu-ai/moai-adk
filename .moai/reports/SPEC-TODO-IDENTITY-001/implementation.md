# SPEC-TODO-IDENTITY-001 — run-phase implementation evidence

## Claim

UUIDv7 project/card identity side table, public nullable JSON projection, writer issuance/backfill/rollback, lifecycle stability, runtime linkage, future/partial/malformed fail-closed, concurrent-handle conservation, timestamp 비권위 규칙을 구현했다. 관측 명령인 `todo list`도 `newTodoReadStore().LoadPure()`를 사용해 legacy JSON을 migration하거나 UUID 발급하지 않는다.

첫 sync 감사의 F1 High는 `PRAGMA index_list`의 partial UNIQUE를 전역 UUID UNIQUE로 오인하는 결함이었다. `identityUUIDUnique`가 `unique == 1 && partial == 0`인 인덱스만 후보로 인정하도록 수정했고, 같은 UUID를 가진 card 2행을 실제 허용하는 partial index fixture에서 reader/writer 오류와 mutation 0을 고정했다. F2/F3 선택 후보는 결함으로 확정하거나 범위를 확장하지 않았다.

이 문서는 change-scoped run-phase 결과다. 저장소 전체 테스트 판정의 소유자는 integration branch CI이며 현재 **PENDING**이다. commit, push, PR, merge, 운영 migration은 수행하지 않았다. plan audit iteration-2의 FAIL 0.75는 사용자 informed override로 BYPASSED했으며 PASS로 바꾸지 않는다.

### 구현 계약

- `todo_identities(entity_kind,local_id,uuid)`와 `identity_schema_version=1`을 별도 side table/meta로 추가했다.
- core `schema_version=1`, runtime `runtime_schema_version=1`을 유지했다.
- `project_uuid`와 `card_uuid`는 non-omitempty `*string`이며 미발급 pure read는 literal `null`이다.
- UUID는 canonical lowercase nonzero UUIDv7만 허용한다.
- writer transaction 안에서만 schema/UUID를 만들고, 모든 helper는 `*sql.Tx`를 사용한다.
- 기존/preferred UUID는 검증 후 보존하며 불일치는 fail-closed한다.
- storage duplicate local ID는 `ErrBacklogIDConflict`, runtime live/archive 후보 수 불일치는 별도 ambiguity 오류다.
- UUID timestamp는 AddedAt, State, ReportedState, OwnerLabel, ProvenanceJSON의 권위가 아니다.

## Evidence

### RED baseline

pre-GREEN 원문은 `red-baseline.md`에 보존됐다. 기준 HEAD는 `a315dad9af0d3a0e04862e6106b3993d9a3812f7`, 당시 test SHA256은 `3c0e6cd4e581c02a6cbeb5d2991d1256738863ef98133cfe1b55595abc13699f`다. 7개 selector의 top-level 수는 `1/3/3/1/3/3/1=15`였고 모두 semantic RED exit 1이었다.

후속 CLI pure-read RED:

```text
$ env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/cli -run '^TestTodoLegacyRecordRoundTrips$' -count=1 -v -timeout=60s
todo_analysis_test.go:124: legacy item 0 card_uuid = "01a094f4-c9ec-7013-946d-23ae258ad838", present=true; want key-present literal null
todo_analysis_test.go:124: legacy item 1 card_uuid = "01a094f4-c9ec-7061-9256-9e9a491cedec", present=true; want key-present literal null
--- FAIL: TestTodoLegacyRecordRoundTrips (1.29s)
FAIL
exit=1
```

후속 sync-audit F1 RED:

```text
$ env -u ... go test ./internal/kanban -run '^TestTodoIdentityAC006FutureIdentityVersionFailClosed$/partial-uuid-unique-' -count=1 -v -timeout=120s
=== RUN   TestTodoIdentityAC006FutureIdentityVersionFailClosed/partial-uuid-unique-LoadPure
todo_identity_red_test.go:883: LoadPure accepted/mutated partial UNIQUE(uuid): err=<nil> mutation_zero=true duplicates=2
=== RUN   TestTodoIdentityAC006FutureIdentityVersionFailClosed/partial-uuid-unique-writer
todo_identity_red_test.go:883: writer accepted/mutated partial UNIQUE(uuid): err=<nil> mutation_zero=false duplicates=2
--- FAIL: TestTodoIdentityAC006FutureIdentityVersionFailClosed (0.21s)
FAIL
exit=1
```

F1 최소 GREEN:

```text
$ env -u ... go test ./internal/kanban -run '^TestTodoIdentityAC006FutureIdentityVersionFailClosed$/partial-uuid-unique-' -count=1 -v -timeout=120s
--- PASS: TestTodoIdentityAC006FutureIdentityVersionFailClosed/partial-uuid-unique-LoadPure (0.10s)
--- PASS: TestTodoIdentityAC006FutureIdentityVersionFailClosed/partial-uuid-unique-writer (0.12s)
PASS
ok  github.com/modu-ai/moai-adk/internal/kanban  0.657s
exit=0
```

최종 F1 검증:

```text
AC006: PASS, ok github.com/modu-ai/moai-adk/internal/kanban 2.677s
combined identity: PASS, ok github.com/modu-ai/moai-adk/internal/kanban 4.478s
fixed regressions 6: PASS, ok github.com/modu-ai/moai-adk/internal/kanban 0.468s
race: PASS, ok github.com/modu-ai/moai-adk/internal/kanban 1.629s, DATA RACE not observed
coverage: todo_identity.go 180/211 = 85.3%; package 30.8%
full package: ok github.com/modu-ai/moai-adk/internal/kanban 167.114s
go vet: exit=0
host build: exit=0
Windows amd64 CGO-disabled build: exit=0
golangci-lint: 0 issues, exit=0
strict SPEC lint: [], exit=0
git diff --check: exit=0
final identity test SHA256: 7f161190b64d36e89bd32d7d67b1e4d87f632d1d3404170593fb9c4ec3274ae7
```

### GREEN — six diagnosed regressions

```text
$ go test ./internal/kanban -run '^(TestBacklogArchive_PerItemContractFrozen|TestTodoHistoryAddsNoSchemaChange|TestTodoRuntimeSafetyRelocationPreservesRuntime|TestBacklogAdd_CreatesVersion1File|TestMigrationPartialFailureRemovesArtifacts|TestDuplicateIDRejectedByStorage)$' -count=1 -v -timeout=90s
--- PASS: TestBacklogArchive_PerItemContractFrozen (0.00s)
--- PASS: TestTodoHistoryAddsNoSchemaChange (0.01s)
--- PASS: TestTodoRuntimeSafetyRelocationPreservesRuntime (0.20s)
--- PASS: TestMigrationPartialFailureRemovesArtifacts (0.01s)
--- PASS: TestDuplicateIDRejectedByStorage (0.01s)
--- PASS: TestBacklogAdd_CreatesVersion1File (0.01s)
PASS
ok  github.com/modu-ai/moai-adk/internal/kanban  0.681s
exit=0
```

### GREEN — AC matrix

공통 환경 scrub:

```text
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR go test ./internal/kanban -run '<selector>' -count=1 -v -timeout=120s
```

관측 결과:

```text
^TestTodoIdentityAC001  exit=0  fixtures=2 git=1 non_git=1 cards=4
^TestTodoIdentityAC002  exit=0  JSON=1 SQLite=1 active_WAL=1 live/archive
^TestTodoIdentityAC003  exit=0  entropy/backfill/stamp reached=1, retry=1
^TestTodoIdentityAC004  exit=0  mutate=1 archive=1 restore=1 reopen=1
^TestTodoIdentityAC005  exit=0  assignments=2, missing mutation_zero=1, candidates=2 mutation_zero=true
^TestTodoIdentityAC006  exit=0  successes=2 cards=2 UUIDs=2, future/retired guards
^TestTodoIdentityAC007  exit=0  authority_fields=added_at,state,reported_state,owner_label,provenance_json
```

Combined readback:

```text
$ env -u ... go test ./internal/kanban -run '^TestTodoIdentityAC' -count=1 -v -timeout=120s
PASS
ok  github.com/modu-ai/moai-adk/internal/kanban  12.385s
exit=0
```

최종 top-level test count와 hash:

```text
identity_top_level_tests=15
747506b9c31e037967f611e26e9e03d8fb21f119912f92813ab7c4eb86c7c3e5  internal/kanban/todo_identity_red_test.go
```

### 전체 kanban·CLI·race

```text
$ env -u ... go test ./internal/kanban -count=1 -timeout=240s
ok  github.com/modu-ai/moai-adk/internal/kanban  184.579s
exit=0
```

```text
$ env -u ... go test -race ./internal/kanban -run '^TestTodoIdentityAC006ConcurrentDistinctHandles$' -count=1 -v -timeout=120s
AC-TID-006 executed handles=2 barrier=1 successes=2 cards=2 UUIDs=2
PASS
ok  github.com/modu-ai/moai-adk/internal/kanban  2.199s
exit=0
DATA RACE: not observed
```

```text
$ env -u ... go test ./internal/cli -run '^(TestTodoLegacyRecordRoundTrips|TestTodoList_JSONKeepsDroppedCards|TestTodoList_JSONIgnoresLimit|TestTodoList_JSONStructured|TestTodoList_EmptyQueueIsNotAnError|TestTodoList_DroppedHiddenByDefault|TestTodoList_DefaultBoundedWithWithheldNotice|TestTodoList_LockFreeWhileForeignProcessHoldsLock|TestTodoListJSONIsIdempotent|TestTodoReadSurface_.*|TestTodoDisclosure_LeavesBacklogJSONUntouched|TestResolveTodoQueueRoot_.*|TestTodoQueue_FallbackAdoptsExistingLocalQueue|TestTodoQueue_WorktreeSeesPrimaryQueue)$' -count=1 -v -timeout=180s
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  85.351s
exit=0
```

### Coverage

```text
$ env -u ... go test ./internal/kanban -run '^TestTodoIdentityAC' -count=1 -coverprofile=/tmp/todo-identity-final-01a09337.cover -covermode=atomic -timeout=120s
ok  github.com/modu-ai/moai-adk/internal/kanban  4.507s  coverage: 30.8% of statements
todo_identity.go aggregate covered=180 total=211 coverage=85.3%
```

핵심 함수 관측치는 `identitySchemaPresent=89.5%`, `readIdentitySnapshot=92.0%`, `ensureRecordIdentities=91.3%`다. 방어적 DB driver 오류 분기를 포함한 일부 함수는 `validateIdentityTable=81.2%`, `identityUUIDUnique=76.5%`, `ensureIdentity=83.3%`, `ensureStoredIdentities=66.7%`이며 파일 전체 변경 로직 기준은 85.3%다. package 전체 30.8%와 혼동하지 않는다.

### DB schema/meta/index readback

```text
$ env -u ... go test ./internal/kanban -run '^TestTodoIdentityAC003SchemaCreationFaultRollsBack$' -count=1 -v -timeout=120s
AC-TID-003 schema fault reached=1 writer_error=stamp identity schema: constraint failed: identity-stamp-fault-reached (1811) table=0 stamp=0 rows=0
AC-TID-003 schema fault retry=1 ddl="CREATE TABLE todo_identities(\n  entity_kind TEXT NOT NULL,\n  local_id TEXT NOT NULL,\n  uuid TEXT NOT NULL UNIQUE,\n  PRIMARY KEY(entity_kind,local_id)\n)" meta=schema:1/runtime:absent/identity:1 indexes=[sqlite_autoindex_todo_identities_2:pk:unique=1 sqlite_autoindex_todo_identities_1:u:unique=1]
PASS
ok  github.com/modu-ai/moai-adk/internal/kanban  0.606s
exit=0
```

runtime-only stamp는 identity-only Add fixture에서 생성되지 않아 `absent`가 정상이다. 기존 runtime fixture와 future-version 회귀는 `runtime_schema_version=1` 계약을 별도로 통과했다.

### Static/build/MX

```text
$ go vet ./internal/kanban ./internal/cli
exit=0

$ go build ./...
exit=0

$ CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...
exit=0

$ /opt/homebrew/bin/timeout 120s /opt/homebrew/bin/golangci-lint run ./internal/kanban/... ./internal/cli/... --output.text.colors=false --output.text.print-issued-lines=false
0 issues.
exit=0

$ git diff --check
exit=0
```

첫 lint 시도는 설치된 버전이 `--out-format`을 지원하지 않아 코드 분석 전 exit 3이었다. `run --help`에서 현재 `--output.text.*` 옵션을 확인하고 위 명령으로 재실행했다.

MX readback:

```text
internal/kanban/todo_runtime.go:159: @MX:NOTE: [TID:LINK]
internal/kanban/backlog_store.go:585: @MX:NOTE: [TID:PURE]
internal/kanban/backlog_store.go:674: @MX:WARN: [TID:TX]
internal/kanban/backlog_store.go:675: @MX:REASON: [TID:TX]
internal/kanban/backlog_store.go:742: @MX:NOTE: [TID:RETURN]
mx_tid_marker_count=5
```

## Baseline-attribution

```text
branch=WT-todo-unified
HEAD=a315dad9af0d3a0e04862e6106b3993d9a3812f7
origin/main...HEAD=0 3005
```

preflight와 최초 GREEN, sync-audit F1 delta는 모두 이 동일 HEAD/current dirty tree에서 측정했다. F1 production 한 줄 수정 뒤 partial-index 단독, AC006, combined identity, 고정 회귀 6개, race, coverage, 전체 `internal/kanban`, static/build/lint를 다시 실행했다.

## Gaps

- integration branch CI의 저장소 전체 verdict: PENDING.
- commit/stage/push/PR/merge: NOT RUN, 권한 범위 밖.
- 운영 Todo DB migration/install/card-state 전환: NOT RUN.
- 첫 독립 sync auditor verdict: FAIL(F1 High). 이 run에서 F1을 수정했으며 delta re-audit verdict는 PENDING.
- plan audit iteration-2 FAIL 0.75는 BYPASSED 상태로 유지된다.
- package 전체 coverage는 30.8%이며 identity helper change scope 85.3%와 별도다.

## Residual-risk

- SQLite driver가 반환하는 저수준 Query/Scan/Rows 오류의 모든 분기를 강제하지는 않았다. 대신 malformed schema 6종, partial UNIQUE reader/writer 2종, malformed row 4종, preferred/UNIQUE 3종의 실제 fail-closed와 변경 0을 확인했다.
- `todo next`의 bare read는 이번 `runTodoList` 수정 범위가 아니며 기존 writer/adoption 의미를 유지했다.
- project/card UUID는 identity를 제공하지만 후속 owner token, generation, completion receipt/recovery, Graph/UI를 구현하지 않는다.
- dirty worktree의 다른 SPEC 변경은 보존했으며 이 보고서는 그 변경의 완료를 주장하지 않는다.

### Planned vs actual divergence

계획의 production/test 범위는 6–8개였다. 실제 identity run에서 production/test 11개를 다뤘다: kanban identity/storage/runtime 5개, kanban regression test 4개, CLI pure-reader production/test 2개. 상한 8개 대비 +3개, file-count drift는 `3/8 = 37.5%`다. 원인은 전체 package에서 발견된 frozen contract 3건과 공개 `todo list` pure-read 회귀 1건을 사용자 승인된 “전체 회귀 해결” 범위에서 닫았기 때문이다.

- feature drift: 0 — Graph, visualization, owner-token, receipt, approval subsystem 미도입.
- dependency drift: 0 — 기존 `github.com/google/uuid`만 재사용.
- directory drift: 0 — 기존 `internal/kanban`, `internal/cli`, SPEC/report 디렉터리만 사용.
- schema drift: 0 — 승인된 identity side table 외 core/runtime version·물리 컬럼 변화 없음.
