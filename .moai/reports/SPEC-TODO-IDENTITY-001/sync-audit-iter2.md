# SPEC-TODO-IDENTITY-001 — sync audit iteration 2

## Claim

**Overall Verdict: PASS — 98/100**

첫 sync 감사의 차단 결함 F1은 현재 dirty tree에서 닫혔다. `identityUUIDUnique`는 이제 SQLite `PRAGMA index_list`의 `unique == 1 && partial == 0`인 인덱스만 전역 UUID UNIQUE 후보로 인정한다. 신규 회귀는 partial UNIQUE가 동일 card UUID 2행을 실제 허용하는 counterexample을 먼저 증명한 뒤 `LoadPure`와 writer 각각이 오류를 반환하고 schema/meta/items/identity 행을 바꾸지 않는지 검사한다. 두 하위 테스트는 독립 selector에서 모두 PASS했다.

7개 AC의 combined selector, AC006 selector와 race, 6개 고정 회귀, 전체 `internal/kanban`, 변경 핵심 파일 coverage, vet, host/Windows build, golangci-lint, strict SPEC lint, diff check, MX exact-five가 모두 exit 0이다. 차단 finding은 남지 않았다. 기존 F2/F3은 재현되지 않은 선택적 위험으로만 유지하며 Functionality/Security must-pass를 왜곡하지 않는다.

이 감사가 변경한 파일은 이 보고서 하나뿐이다. production, tests, SPEC, DB/card 상태, git index/history는 변경하지 않았다.

### AC 판정표

| AC | 판정 | 현재 실행 근거 |
|---|---|---|
| AC-TID-001 | PASS | combined selector: Git/비Git fixture 2개, project UUID 2개, card UUID 4개, 합집합 6개를 검사하는 테스트 PASS. 신규 partial UNIQUE counterexample도 잘못된 전역 유일성 승인을 거절함. |
| AC-TID-002 | PASS | combined selector: legacy JSON, SQLite, active WAL의 명시적 null 및 무변경 테스트 3개 PASS. 동일-fixture 반복 강화는 F3 선택 위험으로만 남김. |
| AC-TID-003 | PASS | entropy/backfill/schema-stamp fault 실제 도달, 같은 transaction rollback, fault 제거 후 재시도 테스트 3개 PASS. |
| AC-TID-004 | PASS | Add→Mutate→Archive→Restore→reopen identity 안정성 테스트 PASS. |
| AC-TID-005 | PASS | live/archive runtime projection, missing 및 live/archive ambiguity mutation-zero, state/owner/provenance 권위 테스트 3개 PASS. |
| AC-TID-006 | PASS | distinct SQLite handle 동시 Add, future/partial/malformed/preferred fail-closed, retired writer 보존 테스트 PASS. partial UNIQUE reader/writer 단독과 `-race`도 PASS. |
| AC-TID-007 | PASS | 역순 UUIDv7 timestamp가 AddedAt/State/ReportedState/OwnerLabel/ProvenanceJSON의 권위가 아님을 검사하는 테스트 PASS. |

### Dimension Scores

활성 SPEC에 `evaluator_profile`이 없고 `.moai/config/sections/harness.yaml`의 `default_profile`은 `default`이므로 `.moai/config/evaluator-profiles/default.md`의 40/25/20/15 flat profile을 적용했다.

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 100/100 | PASS (must-pass) | `go test ./internal/kanban -run '^TestTodoIdentityAC' ...`: `PASS`, `ok ... 5.761s`; partial 단독, AC006, fixed6 및 전체 package도 exit 0. |
| Security (25%) | 100/100 | PASS (must-pass) | fail-closed partial schema 검증 PASS; `govulncheck`: `No vulnerabilities found.` / 호출 코드 영향 0건; `go mod verify`: `all modules verified`. Critical/High finding 없음. |
| Craft (20%) | 95/100 | PASS | `todo_identity.go aggregate: 180/211 = 85.3%`; `go vet` 무출력 exit 0; golangci-lint `0 issues.`; 신규 회귀가 counterexample 도달과 reader/writer mutation-zero를 함께 검사함. |
| Consistency (15%) | 95/100 | PASS | host/Windows build 무출력 exit 0, gofmt diff 무출력, strict SPEC `[]`, `git diff --check` 무출력, MX count 정확히 5, physical runtime UUID declaration 0. |

가중치 계산: `100×0.40 + 100×0.25 + 95×0.20 + 95×0.15 = 98.25`, 표시 점수 98/100. Functionality와 Security must-pass 모두 PASS다.

### TRUST 5

| Dimension | 판정 | 근거 |
|---|---|---|
| Tested | PASS | F1 counterexample 단독, 15개 identity top-level, fixed6, race, 전체 package, coverage를 현재 tree에서 실행함. |
| Readable | PASS | 최소 조건 변경은 `unique == 1 && partial == 0` 한 분기로 의도가 직접 드러나며 신규 테스트 이름도 실패 모드를 명시함. |
| Unified | PASS | 기존 SQLite schema validator와 `BacklogStore` test fixture 안에서 수정했고 별도 추상화를 추가하지 않음. |
| Secured | PASS | malformed/partial schema가 reader/writer 양쪽에서 fail-closed하고, 호출 가능한 취약점 및 secret-shaped 신규 값은 관측되지 않음. |
| Trackable | PASS (local artifact) | F1 RED→GREEN, AC/REQ, 현재 hash 및 baseline을 이 보고서에 연결함. commit/push/PR/운영 migration은 권한 밖이며 NOT RUN. plan audit FAIL 0.75의 사용자-informed BYPASSED 이력은 PASS로 재해석하지 않음. |

### Findings

- **F1 [High] [blocking] [CLOSED in iteration 2] [confidence: 1.00]** `internal/kanban/todo_identity.go:121-128`, `internal/kanban/todo_identity_red_test.go:856-885` — production이 `partial == 0`을 요구하고, 실제 중복 2행 counterexample에서 reader/writer가 fail-closed하며 영속 상태를 바꾸지 않는 회귀가 독립 PASS했다. Required fix: 없음.
- **F2 [Medium] [optional] [not reproduced] [confidence: 0.78]** `internal/kanban/backlog_store.go:765-775` — `Add`의 commit 뒤 별도 `LoadPure` 사이에 archive/remove가 개입할 수 있다는 기존 제어흐름 위험이다. 이번 delta와 AC006 동시 Add에서 semantic failure는 관측되지 않았다. Required fix: 없음(선택적 hardening으로만 유지; 필요하면 결정적 barrier 재현을 먼저 추가).
- **F3 [Low] [optional] [not reproduced] [confidence: 1.00]** `internal/kanban/todo_identity_red_test.go:247-310` — AC002 직접 fixture는 매체별 `LoadPure` 1회다. combined 및 기존 idempotence 보조 범위에서 실패는 관측되지 않았다. Required fix: 없음(원하면 각 동일 fixture 2회 read를 보강).

### Modified-file scope

Delta 판단에서 직접 읽은 파일:

- `internal/kanban/todo_identity.go` — F1 production 조건
- `internal/kanban/todo_identity_red_test.go` — 신규 partial UNIQUE reader/writer 회귀 및 AC 전체
- `.moai/reports/SPEC-TODO-IDENTITY-001/implementation.md` — F1 RED/GREEN 구현자 근거(주장으로만 취급하고 독립 재실행으로 검증)
- `.moai/specs/SPEC-TODO-IDENTITY-001/progress.md` — AC006 및 run-phase 상태 기록
- 이전 전체 감사 대상 production/test 파일은 combined/full/build/vet/lint 영향 범위로 재실행했으며, 이번 delta에서 새 production 변경은 `todo_identity.go` 조건뿐이었다.

최종 측정 hash:

```text
edb45684b7d17d2c18f0224cf6dd5a883bc3ef87434cf62ad091ab295137d007  internal/kanban/todo_identity.go
7f161190b64d36e89bd32d7d67b1e4d87f632d1d3404170593fb9c4ec3274ae7  internal/kanban/todo_identity_red_test.go
dcbfe6a0292027f9d3958763ed8332423ac6c8d00ad83afef7eae47295d9bc08  .moai/reports/SPEC-TODO-IDENTITY-001/implementation.md
78e00c48df03b1867a1c62b7e4ab45162182781d7e84bfdca20364a9952d771d  .moai/specs/SPEC-TODO-IDENTITY-001/progress.md
c54fc0c28eb58c9fc38d300e3b2442049e22824248bda2112ac56e85c6692f43  .moai/reports/SPEC-TODO-IDENTITY-001/sync-audit.md
```

## Evidence

모든 Go 명령은 다음 환경 scrub을 같은 compound invocation 안에서 적용했다.

```bash
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR <command>
```

### F1 source 및 신규 회귀

```text
internal/kanban/todo_identity.go:121: var seq, unique, partial int
internal/kanban/todo_identity.go:127: if unique == 1 && partial == 0 {
internal/kanban/todo_identity_red_test.go:857: t.Run("partial-uuid-unique-"+op, func(t *testing.T) {
internal/kanban/todo_identity_red_test.go:862: CREATE UNIQUE INDEX uuid_project_only ON todo_identities(uuid) WHERE entity_kind='project';
internal/kanban/todo_identity_red_test.go:864: INSERT ... ('card','t1',?),('card','t2',?);
internal/kanban/todo_identity_red_test.go:871-872: duplicateCount != 2 => partial UNIQUE counterexample not reached
internal/kanban/todo_identity_red_test.go:874-883: before/after identityPersistentState; err != nil 및 mutation_zero 검사
```

신규 회귀는 assertion을 약화하지 않았다. 오류만 확인하는 테스트가 아니라 partial index가 동일 UUID card 2행을 실제 허용했는지 먼저 `duplicateCount == 2`로 고정하고, reader와 writer 모두에서 `err != nil` 및 schema/meta/items/identity 행 snapshot equality를 요구한다.

독립 partial selector:

```bash
go test ./internal/kanban -run '^TestTodoIdentityAC006FutureIdentityVersionFailClosed$/partial-uuid-unique-' -count=1 -v -timeout=120s
```

```text
=== RUN   TestTodoIdentityAC006FutureIdentityVersionFailClosed/partial-uuid-unique-LoadPure
=== RUN   TestTodoIdentityAC006FutureIdentityVersionFailClosed/partial-uuid-unique-writer
    todo_identity_red_test.go:979: AC-TID-006 executed future_version operations=2 schema_guards=6 partial_unique_operations=2 row_guards=4 preferred_guards=3
--- PASS: TestTodoIdentityAC006FutureIdentityVersionFailClosed (0.33s)
    --- PASS: TestTodoIdentityAC006FutureIdentityVersionFailClosed/partial-uuid-unique-LoadPure (0.14s)
    --- PASS: TestTodoIdentityAC006FutureIdentityVersionFailClosed/partial-uuid-unique-writer (0.19s)
PASS
ok  github.com/modu-ai/moai-adk/internal/kanban  2.066s
```

### AC006, combined identity, fixed regressions, race

```text
$ go test ./internal/kanban -run '^TestTodoIdentityAC006' -count=1 -v -timeout=120s
todo_identity_red_test.go:789: AC-TID-006 executed handles=2 barrier=1 successes=2 cards=2 UUIDs=2
todo_identity_red_test.go:979: AC-TID-006 executed future_version operations=2 schema_guards=6 partial_unique_operations=2 row_guards=4 preferred_guards=3
todo_identity_red_test.go:1004: AC-TID-006 retired regression executed writers=2 mutation_zero=1
PASS
ok  github.com/modu-ai/moai-adk/internal/kanban  4.099s

$ go test ./internal/kanban -run '^TestTodoIdentityAC' -count=1 -v -timeout=120s
todo_identity_red_test.go:244: AC-TID-001 executed fixtures=2 git=1 non_git=1 cards=4
todo_identity_red_test.go:266: AC-TID-002 executed JSON=1 live=1 archive=1
todo_identity_red_test.go:286: AC-TID-002 executed SQLite=1 live=1 archive=1
todo_identity_red_test.go:310: AC-TID-002 executed active_WAL=1 committed_snapshot=1 live=1 archive=1
todo_identity_red_test.go:334: AC-TID-003 entropy fault reached=1 writer_error=issue project UUIDv7: unexpected EOF
todo_identity_red_test.go:394: AC-TID-003 backfill fault reached=1 writer_error=insert identity: constraint failed: identity-fault-reached (1811) identity_rows=0
todo_identity_red_test.go:467: AC-TID-003 schema fault reached=1 writer_error=stamp identity schema: constraint failed: identity-stamp-fault-reached (1811) table=0 stamp=0 rows=0
todo_identity_red_test.go:577: AC-TID-004 executed mutate=1 archive=1 restore=1 reopen=1
todo_identity_red_test.go:649: AC-TID-005 executed runs=1 assignments=2 live=1 archive=1 state_updates=1
todo_identity_red_test.go:670: AC-TID-005 regression guard executed missing=1 mutation_zero=1
todo_identity_red_test.go:730: AC-TID-005 ambiguous candidates=2 writer_error=runtime card "t1" is ambiguous (2 candidates) mutation_zero=true
todo_identity_red_test.go:789: AC-TID-006 executed handles=2 barrier=1 successes=2 cards=2 UUIDs=2
todo_identity_red_test.go:979: AC-TID-006 executed future_version operations=2 schema_guards=6 partial_unique_operations=2 row_guards=4 preferred_guards=3
todo_identity_red_test.go:1004: AC-TID-006 retired regression executed writers=2 mutation_zero=1
todo_identity_red_test.go:1096: AC-TID-007 executed UUIDv7_timestamps=2 authority_fields=added_at,state,reported_state,owner_label,provenance_json
PASS
ok  github.com/modu-ai/moai-adk/internal/kanban  5.761s

$ go test ./internal/kanban -run '^(TestBacklogArchive_PerItemContractFrozen|TestTodoHistoryAddsNoSchemaChange|TestTodoRuntimeSafetyRelocationPreservesRuntime|TestBacklogAdd_CreatesVersion1File|TestMigrationPartialFailureRemovesArtifacts|TestDuplicateIDRejectedByStorage)$' -count=1 -v -timeout=180s
--- PASS: TestBacklogArchive_PerItemContractFrozen (0.00s)
--- PASS: TestTodoHistoryAddsNoSchemaChange (0.02s)
--- PASS: TestTodoRuntimeSafetyRelocationPreservesRuntime (0.41s)
--- PASS: TestMigrationPartialFailureRemovesArtifacts (0.03s)
--- PASS: TestBacklogAdd_CreatesVersion1File (0.04s)
--- PASS: TestDuplicateIDRejectedByStorage (0.05s)
PASS
ok  github.com/modu-ai/moai-adk/internal/kanban  2.019s

$ go test -race ./internal/kanban -run '^TestTodoIdentityAC006ConcurrentDistinctHandles$' -count=1 -v -timeout=120s
todo_identity_red_test.go:789: AC-TID-006 executed handles=2 barrier=1 successes=2 cards=2 UUIDs=2
--- PASS: TestTodoIdentityAC006ConcurrentDistinctHandles (0.17s)
PASS
ok  github.com/modu-ai/moai-adk/internal/kanban  3.310s
```

### Full package와 coverage

```text
$ go test ./internal/kanban -count=1 -timeout=300s
ok  github.com/modu-ai/moai-adk/internal/kanban  183.230s
exit=0

$ go test ./internal/kanban -run '^TestTodoIdentityAC' -count=1 -coverprofile=<temp>/identity.cover -covermode=atomic -timeout=120s
ok  github.com/modu-ai/moai-adk/internal/kanban  6.072s  coverage: 30.8% of statements
todo_identity.go aggregate: 180/211 = 85.3%
```

30.8%는 identity selector가 전체 `internal/kanban` package를 분모로 삼은 선택 테스트 수치다. 변경 핵심 파일의 per-file aggregate는 85.3%이며 85% 문턱을 넘는다.

### Static, build, consistency, security

```text
$ go vet ./internal/kanban ./internal/cli
[no stdout/stderr]
exit=0

$ go build ./...
[no stdout/stderr]
exit=0

$ CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...
[no stdout/stderr]
exit=0

$ golangci-lint run ./internal/kanban/... ./internal/cli/... --output.text.colors=false --output.text.print-issued-lines=false
0 issues.
exit=0

$ gofmt -d internal/kanban/todo_identity.go internal/kanban/todo_identity_red_test.go
[no stdout/stderr]
exit=0

$ moai spec lint .moai/specs/SPEC-TODO-IDENTITY-001 --strict --json
[]
exit=0

$ git diff --check
[no stdout/stderr]
exit=0

$ rg -o '\[TID:(LINK|PURE|TX|RETURN)\]' internal/kanban/todo_runtime.go internal/kanban/backlog_store.go | wc -l
5
exit=0

$ physical runtime UUID declaration scan
PHYSICAL_RUNTIME_UUID_DECLARATIONS=0
exit=0

$ go mod verify
all modules verified
exit=0

$ govulncheck ./internal/kanban ./internal/cli
=== Symbol Results ===

No vulnerabilities found.

Your code is affected by 0 vulnerabilities.
This scan also found 0 vulnerabilities in packages you import and 3
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
exit=0
```

`internal/cli/todo.go:547-551`은 `store := newTodoReadStore()` 뒤 `rec, err := store.LoadPure()`를 호출한다. 동적 SQL/secret grep은 `internal/cli/todo.go`의 설명문에 있는 일반 단어 `token`만 출력했고 credential-shaped literal이나 새 interpolation sink는 관측하지 않았다.

## Baseline-attribution

최종 보고서 작성 직전 다음을 다시 측정했다.

```text
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-unified

$ git rev-parse HEAD
a315dad9af0d3a0e04862e6106b3993d9a3812f7

$ git branch --show-current
WT-todo-unified

$ git rev-list --count --left-right origin/main...HEAD
0  3005
```

`git fetch origin main`을 바로 앞에서 실행했고 `origin/main...HEAD`는 `0 3005`였다. 검증 대상은 위 HEAD 위의 현재 dirty/untracked implementation이며, production/test hash는 Modified-file scope에 기록했다. 첫 sync 감사의 기준선 및 장기 결과를 이번 PASS의 대체 근거로 사용하지 않았다. 이번 보고서의 모든 필수 명령은 iteration 2에서 다시 실행했다.

Iteration history:

| Iteration | Verdict | Score | 설명 |
|---|---|---:|---|
| 1 | FAIL | 84/100 | F1 partial UNIQUE를 전역 UNIQUE로 오인; Functionality must-pass 실패. F2/F3 optional. |
| 2 | **PASS** | **98/100** | F1 최소 조건 및 counterexample reader/writer 회귀 확인; 모든 필수 delta 검증 PASS. F2/F3 optional 유지. |

Plan audit iteration 1/2의 최종 상태는 **FAIL 0.75 후 사용자-informed BYPASSED**다. 이번 구현 sync PASS는 그 plan audit를 PASS로 바꾸지 않는다.

## Gaps

- 추가 보조 CLI selector 묶음은 124.454초 동안 출력 없이 실행되어 감사자가 `SIGINT`로 중단했다: `signal: interrupt`, `FAIL github.com/modu-ai/moai-adk/internal/cli 124.454s`. 이는 test assertion 실패나 timeout이 아니라 감사자가 중단한 **NOT COMPLETED** gap이다. 이번 delta 필수 목록이 아니며, 현재 F1 변경은 kanban schema validator/test에 국한된다. CLI 경로는 source에서 `newTodoReadStore`→`LoadPure`로 확인했고 host build/vet/lint는 PASS했지만 이 중단 실행 자체를 PASS로 세지 않았다.
- repository-wide `go test ./...`, 원격 CI, commit, push, PR, merge, 운영 DB migration/card mutation은 실행하지 않았다.
- `-race`는 AC006 process-local selector에 한정했다. SQLite의 모든 다중 process 스케줄을 증명하지 않는다.
- coverage는 identity selector의 change-focused per-file aggregate다. package 전체 30.8%를 변경 파일 coverage로 오해하지 않는다.
- cross-model optional audit는 iteration 1에서 장기 무응답으로 inconclusive였고 iteration 2에서는 재호출하지 않았다. PASS 근거에 포함하지 않았다.

## Residual-risk

- F2는 commit 뒤 `LoadPure` interleaving의 구조적 가능성일 뿐 이번 실행에서 재현되지 않았다. 차단 결함으로 승격하지 않으며 선택적 결정론적 재현 전에는 자동 수정을 요구하지 않는다.
- F3은 AC002 각 직접 fixture가 같은 객체에서 두 번 읽지 않는 진단력 위험이다. 기존 idempotence 범위와 현재 combined 결과에서 실패는 없었고 요구 동작의 현재 위반으로 재현되지 않았다.
- 로컬 PASS는 원격 통합 branch의 CI/리뷰/배포 승인을 대신하지 않는다. 운영 DB는 열거나 변경하지 않았으므로 운영 migration 상태는 알 수 없다.
- SQLite file/WAL byte equality는 관련 legacy/active-WAL 테스트가 다루지만, F1 writer rollback 회귀는 논리 영속 상태(schema/meta/items/identity rows)를 비교한다. WAL의 내부 바이트 배열 동일성을 일반화하지 않는다.

## Final verdict

**PASS — 98/100.** F1은 `unique == 1 && partial == 0` production 조건과 실제 duplicate-card counterexample 기반 reader/writer mutation-zero 회귀로 닫혔다. Functionality와 Security must-pass 모두 충족했다. F2/F3은 재현되지 않은 optional residual risk이며 merge-blocking finding은 없다.
