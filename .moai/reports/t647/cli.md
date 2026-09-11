# t647 CLI 구현 및 검증 증거

## Claim

todo CLI의 저장소 오귀속, dropped 직접 pick, 불완전 PR 조회의 확정적 결과, 여러 줄 입력에 의한 출력 파손, 없는 ID 조회, 음수 limit, landed 도움말 조기 계산을 수정하였다. analyze는 카드별 정규화와 토큰 집합을 한 번 계산하며, 기존 판정·이력을 유지한다. PR/history의 읽기 전용 경로는 legacy JSON을 마이그레이션하지 않는다.

## Baseline-attribution

- 작업 경로: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-audit`
- HEAD: `ee99507fbe3b4a22c6a0a74815723d222dfdc04d`
- 브랜치: `WT-todo-audit`
- 기준 시점: 2026-09-11
- 테스트 대상: 위 HEAD에 적용한 현재 CLI 변경 및 동시에 적용된 t647 저장소 변경
- 최초 RED 및 기준 성능은 CLI 구현 전 관측한 값이다.
- 실제 사용자 큐·홈을 테스트 데이터로 사용하지 않았다. 테스트의 `todoFixture` 및 `t.TempDir`만 사용하였다.

## Evidence — 최초 RED

명령:

```sh
go test ./internal/cli -run '^TestTodoAudit' -count=1 -v
```

출력:

```text
=== RUN   TestTodoAuditPRIncompleteIsUnknown
=== RUN   TestTodoAuditPRIncompleteIsUnknown/false
    todo_audit_regression_test.go:26: incomplete result=[{"card_id":"t1","outcome":"no-link"}]
         err=<nil>
=== RUN   TestTodoAuditPRIncompleteIsUnknown/true
    todo_audit_regression_test.go:26: incomplete result=[{"card_id":"t1","outcome":"no-link"}]
         err=<nil>
--- FAIL: TestTodoAuditPRIncompleteIsUnknown (0.94s)
    --- FAIL: TestTodoAuditPRIncompleteIsUnknown/false (0.47s)
    --- FAIL: TestTodoAuditPRIncompleteIsUnknown/true (0.48s)
=== RUN   TestTodoAuditPRMissingSkipsNetwork
    todo_audit_regression_test.go:35: stdout="queue is empty\n" err=<nil> calls=[{gh [pr list --state open --limit 100 --json number,title,body,state]}]
--- FAIL: TestTodoAuditPRMissingSkipsNetwork (0.39s)
=== RUN   TestTodoAuditMultilineRows
=== RUN   TestTodoAuditMultilineRows/pr
    todo_audit_regression_test.go:46: stdout="t1\tno-link\t\t\tqueued\t\tcard text\nforged\trow\rreturn\n" err=<nil>
=== RUN   TestTodoAuditMultilineRows/list
    todo_audit_regression_test.go:46: stdout="t1\tqueued\tcard text\nforged\trow\rreturn\n" err=<nil>
=== RUN   TestTodoAuditMultilineRows/next
    todo_audit_regression_test.go:46: stdout="t1\tcard text\nforged\trow\rreturn\n" err=<nil>
=== RUN   TestTodoAuditMultilineRows/history
    todo_audit_regression_test.go:46: stdout="t1\tlive\tqueued\tcard text\nforged\trow\rreturn\n" err=<nil>
--- FAIL: TestTodoAuditMultilineRows (1.41s)
    --- FAIL: TestTodoAuditMultilineRows/pr (0.42s)
    --- FAIL: TestTodoAuditMultilineRows/list (0.38s)
    --- FAIL: TestTodoAuditMultilineRows/next (0.28s)
    --- FAIL: TestTodoAuditMultilineRows/history (0.32s)
=== RUN   TestTodoAuditNegativeLimitEmpty
    todo_audit_regression_test.go:56: negative limit accepted: "queue is empty\n"
--- FAIL: TestTodoAuditNegativeLimitEmpty (0.44s)
=== RUN   TestTodoAuditPickDropped
    todo_audit_regression_test.go:66: stdout="picked t1 [DROPPED — obsolete] original card\n" err=<nil> rec=&{Version:1 LastSeq:1 Items:[{ID:t1 Text:[DROPPED — obsolete] original card AddedAt:2026-09-11T09:25:40Z SpecID:<nil> State:picked Landing:<nil>}] Findings:[] Archived:[]} load=<nil>
--- FAIL: TestTodoAuditPickDropped (0.47s)
=== RUN   TestTodoAuditLandedHelpLazy
    todo_audit_regression_test.go:72: landed help materialized before rendering
--- FAIL: TestTodoAuditLandedHelpLazy (0.18s)
=== RUN   TestTodoAuditPRUsesQueueRoot
=== RUN   TestTodoAuditPRUsesQueueRoot/pr
    todo_audit_regression_test.go:94: stdout=[{"card_id":"t1","outcome":"landed"}]
         stderr="" err=<nil>
=== RUN   TestTodoAuditPRUsesQueueRoot/done
    todo_audit_regression_test.go:97: archived using unrelated project: done t1 landing=landed ref=origin/main
--- FAIL: TestTodoAuditPRUsesQueueRoot (1.55s)
    --- FAIL: TestTodoAuditPRUsesQueueRoot/pr (0.78s)
    --- FAIL: TestTodoAuditPRUsesQueueRoot/done (0.76s)
=== RUN   TestTodoAuditAnalyzeAllocationBound
    todo_audit_regression_test.go:106: 100-card analysis allocates 79201 objects; want <=5000
--- FAIL: TestTodoAuditAnalyzeAllocationBound (0.01s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 6.331s
FAIL
```

추가 RED 명령:

```sh
go test ./internal/cli -run '^TestTodoAudit(DropReason|FindingNote)' -count=1 -v
```

```text
=== RUN   TestTodoAuditDropReasonSingleLine
    todo_audit_regression_test.go:203: stdout="dropped t1 original (reason: reason\nsecond\tcolumn)\n" err=<nil>
--- FAIL: TestTodoAuditDropReasonSingleLine (0.50s)
=== RUN   TestTodoAuditFindingNoteSingleLine
    todo_audit_regression_test.go:210: finding="\t↳  t2 () — reason\nsecond\tcolumn — moai todo drop t1 | moai todo edit t1 \"<text>\""
--- FAIL: TestTodoAuditFindingNoteSingleLine (0.00s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 1.337s
FAIL
```

```sh
go test ./internal/cli -run '^TestTodoAuditPRDoesNotMigrateLegacyJSON$' -count=1 -v
```

```text
=== RUN   TestTodoAuditPRDoesNotMigrateLegacyJSON
    todo_audit_regression_test.go:221: read migrated legacy JSON: <nil>
--- FAIL: TestTodoAuditPRDoesNotMigrateLegacyJSON (0.87s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 1.853s
FAIL
```

```sh
go test ./internal/cli -run '^TestTodoAuditSubprocessScopeIgnoresInheritedRepo$' -count=1 -v
```

```text
=== RUN   TestTodoAuditSubprocessScopeIgnoresInheritedRepo
    todo_audit_regression_test.go:259: gh scope="/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoAuditSubprocessScopeIgnoresInheritedRepo1156174201/001|unrelated/repository" want="/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoAuditSubprocessScopeIgnoresInheritedRepo1156174201/001"|
--- FAIL: TestTodoAuditSubprocessScopeIgnoresInheritedRepo (0.56s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 1.419s
FAIL
```

## Evidence — 성능

수정 전 명령(기준 develop checkout, 동일 커밋):

```sh
go test -overlay /tmp/todo-cli-audit.4G91GB/overlay.json ./internal/cli -run '^$' -bench '^BenchmarkTodoAuditAnalyze$' -benchtime=1x -benchmem -count=1
```

```text
goos: darwin
goarch: arm64
pkg: github.com/modu-ai/moai-adk/internal/cli
cpu: Apple M4 Max
BenchmarkTodoAuditAnalyze/100-16                1       5168417 ns/op      5945600 B/op      79208 allocs/op
BenchmarkTodoAuditAnalyze/500-16                1     128476166 ns/op    149705664 B/op    1996012 allocs/op
PASS
ok github.com/modu-ai/moai-adk/internal/cli 1.173s
```

수정 후 명령(todo-audit checkout):

```sh
go test ./internal/cli -run '^$' -bench '^BenchmarkTodoAuditAnalyze$' -benchtime=1x -benchmem -count=1
```

```text
goos: darwin
goarch: arm64
pkg: github.com/modu-ai/moai-adk/internal/cli
cpu: Apple M4 Max
BenchmarkTodoAuditAnalyze/100-16                1        410000 ns/op        45912 B/op        503 allocs/op
BenchmarkTodoAuditAnalyze/500-16                1       8709042 ns/op       228312 B/op       2503 allocs/op
PASS
ok github.com/modu-ai/moai-adk/internal/cli 1.559s
```

## Evidence — CLI 회귀 GREEN

최종 CLI 변경과 history fixture 수정 후 명령:

```sh
go test ./internal/cli -run '^TestTodoHistoryDegradesWithoutArchiveTables$|^TestTodoAudit' -count=1
```

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	11.986s
```

13개의 Audit 테스트와 history의 archive-table 부재 반복 읽기 검사를 통과하였다. 여기에는 GH_REPO·GIT_DIR·GIT_WORK_TREE 환경 격리, 다른 저장소의 커밋으로 done을 허용하지 않는 실제 git 검사, 원본 카드 텍스트 보존, legacy JSON 비마이그레이션 검사가 포함된다.

이전 넓은 범위 실행(최종 pure-reader 변경 전 바이너리):

```sh
go test ./internal/cli -run 'TestTodo|TestLandedRef|TestHelpNamesTheResolvedLandedRef|TestUsageNamesTheResolvedLandedRef' -count=1
```

```text
--- FAIL: TestTodoHistoryDegradesWithoutArchiveTables (1.26s)
    --- FAIL: TestTodoHistoryDegradesWithoutArchiveTables/dropped_archive_tables (0.83s)
        todo_history_test.go:318: DROP TABLE archived_items: SQL logic error: no such table: archived_items (1)
--- FAIL: TestTodoHistoryLeavesStorageByteIdentical (2.79s)
    todo_history_test.go:586: new state artifact .moai/state/todo/backlog.db-shm appeared during history reads
    todo_history_test.go:586: new state artifact .moai/state/todo/backlog.db-wal appeared during history reads
moai-cli-test: userHomeDirFn redirected real home to sandbox /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-cli-home-796369021
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 365.545s
FAIL
```

이 실패는 timeout이 아니다. 첫 번째는 기존 fixture가 읽기의 schema 재생성을 전제한 것이므로, 표를 한 번 제거하고 반복 조회 후 여전히 없는지 검사하도록 고쳤다. 두 번째는 저장소의 SQLite 읽기 연결이 WAL/SHM 파일을 생성하는 문제로 저장소 담당 구현에 전달하였다. 후속 검사 결과가 아래에 별도로 기록되기 전에는 전체 읽기 비변경 검사가 통과했다고 주장하지 않는다.

커버리지 수집 명령:

```sh
go test ./internal/cli -run '^TestTodoAudit|^TestTodoPR|^TestTodoDone|^TestTodoHistory|^TestTodoAnalyze|^TestTodoAnalysis|^TestTodoLanded|^TestLandedRef|^TestHelpNamesTheResolvedLandedRef|^TestUsageNamesTheResolvedLandedRef' -count=1 -timeout 5m -coverprofile /tmp/todo-cli-audit.4G91GB/coverage.out
```

이 실행도 timeout 없이 177.676s에 종료하였으나 위 저장소 WAL/SHM 부작용으로 실패하였다. 저장소 수리 전 수집된 커버리지이므로 검증 성공의 근거로 사용하지 않는다. 관측된 함수 커버리지는 `newTodoLandedCmd 100.0%`, `todoRequireLanded 100.0%`, `todoTokenSetScore 100.0%`, `analyzeQueue 88.5%`, `runTodoPR 90.0%`다. 전체 CLI 패키지의 8.2%는 이 명령의 선택 범위 밖 코드를 포함한 분모다.

## Evidence — 저장소 후속 수정 후 최종 GREEN

저장소 담당자의 query-only 연결·WAL 가시성 보존 수정 후, 앞선 실패 전부와 새 CLI 회귀를 함께 실행하였다.

```sh
go test ./internal/cli -run '^TestTodoHistoryLeavesStorageByteIdentical$|^TestTodoHistoryDegradesWithoutArchiveTables$|^TestTodoPR_QueueDirUnchanged$|^TestTodoPR_ProjectRootUnchangedWithEvidence$|^TestTodoAudit' -count=1 -timeout 3m
```

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	19.805s
```

위 명령으로 history/PR의 영속 파일 비변경, 기존 evidence 보존, 반복 degraded schema 조회, 모든 새 Audit 회귀를 통과하였다. SQLite의 읽기는 transient coordination 파일을 사용할 수 있으므로, 도움말은 카드·schema·cache를 수정하지 않고 migration 및 queue mutation lock을 수행하지 않는다고 구체화하였다. 운영 중의 모든 순간에 임시 파일조차 생기지 않는다는 주장은 하지 않는다.

## 변경 파일 목록

- `internal/cli/mcp_build_identity_test.go`: 허가된 ancestry source-sweep 기대 좌표 3개와 대응 주석만 실제 위치로 동기화

- `internal/cli/todo.go`: pure reader 생성자, limit 검사, dropped pick 거부, 행 출력
- `internal/cli/todo_pr.go`: subprocess 디렉터리·환경, PR 불완전 상태, ID 검사, pure read
- `internal/cli/todo_landed.go`: 도움말 지연 초기화
- `internal/cli/todo_history.go`: pure reader 및 행 출력
- `internal/cli/todo_disclosure.go`: pure reader로 경로 조회
- `internal/cli/todo_drop.go`: 사유 출력의 행 구분자 처리
- `internal/cli/todo_analysis.go`: 정규화 캐시 및 finding note 행 처리
- `internal/cli/todo_pr_test.go`: 불완전 조회의 변경된 unknown 계약
- `internal/cli/todo_history_test.go`: pure reader가 schema를 재생성하지 않는 반복 읽기 계약
- `internal/cli/todo_audit_regression_test.go`: 회귀 테스트, 비교 검사, 벤치마크

## Gaps

- 실서비스 GitHub 네트워크를 호출하는 E2E는 수행하지 않았다. GH 오류·포화는 응답 fixture, GH 디렉터리·환경은 임시 실행 파일로 검증한다.
- 성능 값은 각각 1회 실행한 합성 입력의 관측이다. 운영 지연 보장이나 반복 측정 신뢰구간이 아니다.
- 전체 프로젝트 테스트와 Windows 실행 검증은 이 CLI 작업의 범위 밖이다.

## Residual-risk

- 분석은 여전히 전체 쌍을 비교한다. 토큰화 중복을 제거했지만 비교 횟수 자체는 N(N-1)/2이다.
- GH 페이지가 포화하면 불확실성을 표시한다. 추가 페이지를 자동 조회하지 않는다.
- 정상 JSON 모양은 유지되지만 불완전 PR 조회에서 `unknown` 및 `pr_lookup` 필드를 읽도록 소비자가 적응해야 한다.
- 과거 findings는 삭제·덮어쓰지 않으며, 편집 이후의 현재 유사도를 뜻한다고 해석해서는 안 된다.
- 읽기 전용 SQLite 연결에서도 transient coordination 파일이 사용될 수 있다. 영속 데이터와 schema를 변경하지 않는다는 보장과 구별한다.

## 통합 검사 메타데이터 동기화

통합 검사에서 `TestAuditLagUsesBinlagSeam`의 수동 좌표 fixture 불일치를 확인하였다. `home_state_coverage.go`의 두 위치는 기준 커밋부터 `245/253`인데 fixture는 `243/251`이었다. 이 production 파일은 기준과 동일하다. `todo_landed.go`만 이번 도움말 지연 초기화 변경 때문에 `216 → 217`로 이동하였다.

기준 귀속 검사는 source sweep의 모든 non-test CLI 소스를 `ee99507fbe3b4a22c6a0a74815723d222dfdc04d` git 객체에서 임시 디렉터리로 구성하고, 변경 전 원래 테스트를 호출하는 안전 overlay로 수행하였다. 전체 패키지 baseline 재컴파일이 아니라 source-sweep baseline 재현이다.

```sh
go test -overlay /tmp/todo-baseline-audit.A3904o/overlay.json ./internal/cli -run '^TestTodoBaselineAuditLag$' -count=1 -v -timeout 2m
```

```text
=== RUN   TestTodoBaselineAuditLag
    todo_baseline_external_test.go:26: source sweep baseline=ee99507fbe3b4a22c6a0a74815723d222dfdc04d; all top-level non-test CLI Go source materialized from git objects; test source unchanged from baseline
    mcp_build_identity_test.go:643: sweep: baseline hit home_state_coverage.go:251 MISSING — the ancestry comparison moved or was reimplemented outside binlag
    mcp_build_identity_test.go:643: sweep: baseline hit home_state_coverage.go:243 MISSING — the ancestry comparison moved or was reimplemented outside binlag
    mcp_build_identity_test.go:648: sweep: NEW ancestry hit home_state_coverage.go:245 — a second comparison outside binlag.Evaluate (REQ-ABI-006 violation)
    mcp_build_identity_test.go:648: sweep: NEW ancestry hit home_state_coverage.go:253 — a second comparison outside binlag.Evaluate (REQ-ABI-006 violation)
--- FAIL: TestTodoBaselineAuditLag (11.66s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 12.586s
FAIL
```

추가 허가에 따라 수동 map의 좌표 세 값과 대응 주석만 동기화하였다. `home_state_coverage.go` production 코드, ancestry predicate, 예외 파일 집합 및 exact-set 비교는 수정하지 않았다.

좌표 전용 변경 검증 명령:

```sh
ruby -e 'path="internal/cli/mcp_build_identity_test.go"; old=IO.popen(["git","show","ee99507fbe:"+path], &:read); expected=old.gsub("todo_landed.go:216","todo_landed.go:217").gsub("home_state_coverage.go:243","home_state_coverage.go:245").gsub("home_state_coverage.go:251","home_state_coverage.go:253").gsub("home_state_coverage.go:245/251","home_state_coverage.go:245/253").gsub("integrity: 243","integrity: 245").gsub("and 251 checks","and 253 checks"); abort("UNEXPECTED_DIFF") unless File.binread(path).bytes==expected.bytes; puts "COORDINATE_ONLY_PASS: three map coordinates and corresponding comments; all other bytes unchanged"'
```

```text
COORDINATE_ONLY_PASS: three map coordinates and corresponding comments; all other bytes unchanged
```

```sh
go test ./internal/cli -run '^TestAuditLagUsesBinlagSeam$' -count=1 -v -timeout 2m
```

```text
=== RUN   TestAuditLagUsesBinlagSeam
--- PASS: TestAuditLagUsesBinlagSeam (0.73s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.597s
```
