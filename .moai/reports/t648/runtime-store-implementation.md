# t648 내부 Todo 실행 저장 구현 — 검증 기록

## Claim

운영자가 승인한 검증 순서 예외(`staged-verification-approval.md`)에 따라 내부 저장 토대를 구현했다. 기존 세 Factory 기록 API는 호출 signature를 유지하면서 프로젝트 Todo DB의 `todo_runtime_runs`, `todo_runtime_assignments`에 기록한다. 일반 폴더·SPEC 없는 입력을 지원하고, reported_state는 카드 완료 권위로 해석하지 않는다.

TodoRuntime DTO는 빈 배열 계약을 유지한다. 공개 LoadPure는 기존 DEFERRED transaction의 같은 snapshot으로 카드와 runtime을 읽는다. writer는 기존 BoardLock과 SQLite transaction을 공유하고, 암묵 run과 assignment는 함께 성공하거나 rollback된다. 기존 카드 Mutate는 runtime을 덮어쓰지 않는다. migration publication에서만 runtime을 별도 복사하고 parity에 포함한다. 신규 CLI/hooks 자동 완료, 운영 이관, push는 하지 않았다.

## Evidence

모든 Go 명령은 아래 환경을 한 호출에서 제거하고 WT에서 실행했다.

```sh
env -u MOAI_HOME -u MOAI_CONFIG_DIR -u MOAI_PROJECT_ROOT -u CLAUDE_PROJECT_DIR -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR
```

기존 계약 RED는 `runtime-store-red.md`의 원문을 따른다. 내부 기반 구현 뒤 추가 재현한 semantic RED는 다음과 같다.

```sh
go test ./internal/kanban -run '^TestTodoRuntimeStorePartialSchema' -count=1 -v -timeout=60s
```

```text
=== RUN   TestTodoRuntimeStorePartialSchemaRefusedWithoutMutation
    todo_runtime_store_test.go:369: partial extension pure read must refuse corruption, got <nil>
    todo_runtime_store_test.go:372: partial extension writer must refuse corruption, got SQL logic error: table todo_runtime_runs has no column named backend (1)
--- FAIL: TestTodoRuntimeStorePartialSchemaRefusedWithoutMutation (0.36s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.927s
FAIL
```

```sh
go test ./internal/kanban -run '^TestTodoRuntimeSafetyStampedPartial' -v -count=1 -timeout=60s
```

```text
=== RUN   TestTodoRuntimeSafetyStampedPartialSchemaRefused
    todo_runtime_safety_test.go:103: stamped partial read: SQL logic error: no such table: todo_runtime_assignments (1)
    todo_runtime_safety_test.go:106: stamped partial writer: <nil>
    todo_runtime_safety_test.go:113: stamped partial rejection changed bytes
--- FAIL: TestTodoRuntimeSafetyStampedPartialSchemaRefused (0.33s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.815s
FAIL
```

두 경우 모두 runtime metadata/table inventory를 읽기 전용으로 확인하고 부분 설치를 거절하는 가드로 교정했다.

```sh
go test ./internal/kanban -run '^TestTodoRuntimeSafety(Relocation|Parity)' -count=1 -v -timeout=60s
```

```text
=== RUN   TestTodoRuntimeSafetyRelocationPreservesRuntime
    todo_runtime_safety_test.go:301: relocation dropped runtime: before={"runs":[{"run_id":"relocate","backend":"","manifest_json":"{}"}],"assignments":[{"run_id":"relocate","card_id":"t1","owner_label":"owner","reported_state":"picked","event_kind":"card.assigned","provenance_json":"{\"spec_id\":\"\",\"spec_path\":\"\",\"spec_sha256\":\"\",\"git_commit\":\"\",\"captured_at\":\"2026-09-12T05:47:13.288177Z\"}"}]} after={"runs":[],"assignments":[]}
--- FAIL: TestTodoRuntimeSafetyRelocationPreservesRuntime (0.53s)
=== RUN   TestTodoRuntimeSafetyParityRejectsRuntimeLoss
    todo_runtime_safety_test.go:309: migration parity accepted lost runtime
--- FAIL: TestTodoRuntimeSafetyParityRejectsRuntimeLoss (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	1.389s
FAIL
```

공식 relocateQueueArtifacts를 실제 임시 DB에 실행한 결과다. publishBacklogRecord의 비공개 staging DB에만 runtime 복사를 추가하고 parity를 보완했다. 기존 Mutate의 runtime 비갱신 계약은 유지한다.

최종 기능 범위 실행:

```sh
go test ./internal/kanban -run '^(TestTodoRuntime|TestRecordFactoryRunStart)' -count=1 -v -coverprofile=/tmp/todo-runtime-final-coverage.out -timeout=180s
```

종료 코드 0, 최상위 20개 PASS. 출력 마지막 원문:

```text
PASS
coverage: 26.7% of statements
ok  	github.com/modu-ai/moai-adk/internal/kanban	24.974s	coverage: 26.7% of statements
```

실제 abort trigger 두 종류(`fixture_assignment_abort`, `fixture_install_abort`)가 호출자 오류 문자열로 관측됐고, run seed/assignment 및 확장 DDL의 rollback을 검증했다. 초기 기반부터 정상 동작한 이 안전 검사는 보존/실패 주입 PASS이며, 가짜 RED를 만들지 않았다. 활성 WAL bytes 불변과 동시 writer 30회/reader 30회 카드·runtime 동일 snapshot 검사도 PASS다. archive·findings·last_seq=100·정렬·upsert·stale DTO·비Git·무SPEC·누락 카드·퇴역 writer를 검사했다.

이전 범위 race 실행(이전 복사 보완 전, 최종 race 아님):

```sh
go test -race ./internal/kanban -run '^(TestTodoRuntime|TestRecordFactoryRunStart)' -count=1 -coverprofile=/tmp/todo-runtime-safety-coverage.out -timeout=180s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/kanban	23.622s	coverage: 22.9% of statements
```

`gopls check`로 생산 파일 5개 검사: 종료 코드 0, 출력 없음. `git diff --check`: 종료 코드 0, 출력 없음.

이전 보완까지 포함한 영향 범위 회귀 명령:

```sh
go test ./internal/kanban -run '(Backlog|Migrate|Quarantine|Parity|TodoRuntime|RecordFactoryRunStart|Pure)' -count=1 -coverprofile=/tmp/todo-runtime-regression-final.out -timeout=180s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/kanban	34.255s	coverage: 47.0% of statements
```

종료 코드 0. 분모는 kanban 패키지 전체 statement이고 실행 범위만 위 정규식이다. 47.0%는 85% 미만이므로 품질 게이트 PASS가 아니다. 이와 별도로 기능 범위 `/tmp/todo-runtime-final-coverage.out`에서 새 `todo_runtime.go`는 93/108=86.11%, 변경 `factory_runtime.go`는 23/25=92.00%다. 이 파일별 수치로 패키지 기준을 대체하지 않는다.

## Baseline-attribution

최종 동결 소스의 race 재검사:

```sh
go test -race ./internal/kanban -run '^(TestTodoRuntime|TestRecordFactoryRunStart)' -count=1 -timeout=180s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/kanban	35.885s
```

종료 코드 0. 같은 동결 소스의 gopls 5개 파일 검사와 git diff --check도 각각 출력 없이 종료 코드 0이다. 최종 영향범위 회귀의 파일별 statement 측정은 아래와 같다. 패키지 분모 47.0%와 구별한다.

| 파일 | 실행/전체 statement | 파일별 비율 |
|---|---:|---:|
| todo_runtime.go | 94/108 | 87.04% |
| factory_runtime.go | 23/25 | 92.00% |
| backlog_store.go | 244/263 | 92.78% |
| backlog_sqlite.go | 117/141 | 82.98% |
| backlog_migrate.go | 277/354 | 78.25% |

- WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-unified`, branch `WT-todo-unified`.
- HEAD `a315dad9af0d3a0e04862e6106b3993d9a3812f7` 위 미커밋 작업.
- 수정 직전 fetch 성공, `origin/main...HEAD`: `0 3005`.
- ModeSelection: 단일 도메인 SQLite 저장, 단일 manager-develop writer. 읽기 전용 조사·감사만 병렬. 기존 child Tier 유지.
- 생산 소유: 새 `todo_runtime.go`와 기존 `factory_runtime.go`, `backlog_store.go`, `backlog_sqlite.go`, `backlog_migrate.go`. 새 helper는 기존 카드 저장 경계와 runtime 전용 쓰기를 분리하기 위한 것이며 다른 도메인 추상화는 추가하지 않았다.
- 테스트 소유: 기존 `factory_runtime_test.go`의 저장 위치 기대만 Todo readback으로 변경하고 provenance assertions는 유지. `todo_runtime_store_test.go`, 새 `todo_runtime_safety_test.go`.
- 다른 세션의 Factory schema guard와 SPEC 본문은 수정하지 않았다.
- 동결 시 SHA-256: todo_runtime.go `e53929950fd5013262d47235b971ab68d96ca363ff61929d7fd7d132f5a08292`, todo_runtime_safety_test.go `4d68575aa5fc2b3a6c11a762b730725ad18c52e219ad61055bf2ee7f36956b37`, todo_runtime_store_test.go `091feb3e2568abdda5266c5936fab02d26a51c0415bcde76872bb4f107f3dd82`.

최종 생산 파일 해시 명령: `shasum -a 256 internal/kanban/factory_runtime.go internal/kanban/backlog_store.go internal/kanban/backlog_sqlite.go internal/kanban/backlog_migrate.go internal/kanban/todo_runtime.go`.

```text
2d9e915f70fa170a4d7c97c67dd5d02783ff6ae40befe5ab4f0fe0275d901f85  internal/kanban/factory_runtime.go
2ca44fcdfc5336dbacfdc25dc25bb35eeea4c8ca1d844a73122842f0344b6398  internal/kanban/backlog_store.go
22ca752be7482daa01b06db22da150b16ae6503e963a2423a1c8cd3dcb30e70a  internal/kanban/backlog_sqlite.go
b988013611a398c5b8d075ed2f3b6802528db22e216ac6e202df97d373482b31  internal/kanban/backlog_migrate.go
e53929950fd5013262d47235b971ab68d96ca363ff61929d7fd7d132f5a08292  internal/kanban/todo_runtime.go
```

## Gaps

- 전체 패키지 실행은 `go test ./internal/kanban -count=1 -coverprofile=/tmp/todo-runtime-all-coverage.out -timeout=240s`에서 전역 시간 제한에 걸렸다. 원문 발췌 `panic: test timed out after 4m0s`, `FAIL github.com/modu-ai/moai-adk/internal/kanban 240.587s`. 긴 stack은 tool 출력에서 일부 잘렸으며 전체 원문 보존을 주장하지 않는다. 특정 코드 결함·deadlock이나 인프라 원인으로 단정하지 않는다. 전체 패키지 PASS는 없다.
- 구현 전 LSP baseline을 별도로 캡처하지 못했다. 구현 후 gopls clean을 baseline 무회귀 증거로 바꾸지 않는다.
- 전체 저장소 CI는 통합 브랜치 CI 담당이며 PENDING이다. commit/push/설치/운영 DB/실제 카드 변경은 없다.
- 운영자가 승인한 예외는 검증 순서만 한정한다. 전체23AC와 85% 품질 기준을 면제하지 않는다.

## Residual-risk

### 독립 감사 후 최종 lint 범위 교정

root가 나머지 16건을 Todo storage 품질 범위로 명시 승인했다. backlog_integrity_audit_test.go의 10건, backlog_migrate.go의 3건, backlog_pure_reader_test.go의 3건만 수정했다. 정리 호출은 반환값 무시를 명시했고, 두 QueryRow.Scan은 실패하면 t.Fatal로 검사를 중단하도록 바꿨다. 다른 리팩터링은 하지 않았다.

`golangci-lint run ./internal/kanban --timeout=2m`: 종료 코드 0, 원문:

```text
0 issues.
```

위 환경 scrub 후 다음 정확한 영향 범위 명령을 실행했다.

```sh
go test ./internal/kanban -run '^(TestAuditPureReadMutatesSchema|TestAuditInterruptedMigrationShadowsLegacy|TestAuditAddRewritesWholeArchive|TestAuditFutureSchemaMutatedBeforeRefusal|TestAuditReadSnapshot|TestPureBacklogReaderRejectsSQLWrites|TestPureBacklogReaderNeverCreatesMissingDB|TestPureBacklogReaderSeesCommittedWAL|TestTodoRuntimeSafetyRelocationPreservesRuntime)$' -count=1 -v -timeout=180s
```

9개 PASS, 종료 코드 0, 원문 마지막:

```text
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	3.382s
```

독립 감사가 이 16건 교정 **이전 소스**에서 패키지 전체를 `-parallel=4`, `-timeout=600s`로 실행한 로그 `/tmp/todo-runtime-audit.PNUqKd/package-tests.log`를 읽어 전체 PASS와 86.2%를 확인했다. 앞서 기록한 47.0%는 범위 선택 검사의 패키지분모 수치였으며 실제 전체검사 86.2%로 구별해야 한다. 최종 정리 호출 수정 이후의 전체 패키지 값으로 소급하지 않는다. 최종 품질 판정은 독립 감사 담당자에게 넘겼고 중복 전체검사는 수행하지 않았다.

최종 수정 세 파일 해시:

```text
b7d866eb036ffe459b55556c0cca7359f34f70aed5e35753108b54b3c94bbe82  internal/kanban/backlog_integrity_audit_test.go
f24692a0309fb13a0fc49320d2802afdb74faaef2e75f30a949e97ce22bb6fdb  internal/kanban/backlog_migrate.go
744b01b1662722704683046d53fbf891bcc8a1d57de87f279ce1f5e976ca0be0  internal/kanban/backlog_pure_reader_test.go
```

git diff --check 출력 없음, 종료 코드 0. HEAD와 branch는 a315dad9a / WT-todo-unified이며 fetch 뒤 divergence는 0 3005다.

### 독립 감사 2회차 — 새 파일 lint 교정

감사에서 신규 두 파일의 cleanup 반환값 미표시 12건을 확인했다. `todo_runtime.go`의 rows.Close/tx.Rollback/eng.close와 신규 safety test의 동일 cleanup에만 기존 파일에서 사용하는 `_ =` 패턴을 적용했다. Scan/rows.Err, write Commit, BoardLock Release의 오류 전파는 유지한다. 다른 파일의 lint를 일괄 수정하지 않았다.

명령 `golangci-lint run ./internal/kanban --timeout=2m`, 종료 코드 1. `/tmp/todo-runtime-lint-r2.log` 원문 마지막:

```text
16 issues:
* errcheck: 16
```

같은 로그에서 `rg 'todo_runtime'` 출력 없음. 신규 두 파일의 12건은 사라졌지만 패키지 lint PASS가 아니다. 나머지 16건은 이번 수정 위치 밖이며, 별도 사전 baseline을 측정하지 않았으므로 기왕 결함으로 단정하지 않는다.

2회차 재동결 해시(앞의 최초 동결 해시를 대체):

```text
3e4d7711c29dbcf4189f2991a94e4132d035c925d80a7fbbe4b77317b35fd8c8  internal/kanban/todo_runtime.go
2ac89c95f80c45a1c5cdb73b878b8518a598e57e88b2466df8d38a5861fca93e  internal/kanban/todo_runtime_safety_test.go
```

나머지 생산 파일과 테스트는 그대로다. diff --check 출력 없음, 종료 코드 0. 앞의 coverage/race는 최초 동결 소스의 측정이며 cleanup 수정 후 동일 측정으로 간주하지 않는다.

2회차 기능 범위 명령: 위 환경 scrub + `go test ./internal/kanban -run '^(TestTodoRuntime|TestRecordFactoryRunStart)' -count=1 -timeout=180s`. 종료 코드 0, 원문:

```text
ok  	github.com/modu-ai/moai-adk/internal/kanban	19.623s
```

이 저장 기반은 소유권 CAS, 실행 세대, 완료 receipt, Git/nonGit 완료 정책, hooks 자동 완료, old/new binary 혼재 보호를 구현하지 않는다. 구 Factory 과거 이력은 보존하지만 신규 기록 경로만 Todo로 옮겼다. 운영 전환 전 후속 child 및 독립 감사가 필요하다.
