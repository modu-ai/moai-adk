auditor-model: claude-opus-5-5

# t2081 sync 재감사 보고서 (2회차, 수리 델타 한정, sync-auditor)

verdict: PASS-WITH-DEBT
audited_sha: c1827b7c4dc58b9313c4476ae4de4bd0e91cc082

- 카드: t2081 (`git worktree move` 병렬화, 동시 이동 상한 4)
- 감사 대상 커밋: `c1827b7c4dc58b9313c4476ae4de4bd0e91cc082` (`git rev-parse HEAD`로 감사 시작과 Codex 호출 뒤에 각각 확인)
- 범위: 수리 델타 `git diff 3be3e51f36 c1827b7c4d -- internal` (수리 커밋 3bc707a906, 이후 커밋은 기록 파일만)
- 종합 판정: PASS-WITH-DEBT. BLOCKING 0건. 1회차 BLOCKING F1·F2는 닫혔다. 1회차 SHOULD-FIX F4와 MINOR F5~F9는 부채로 남고, 이번에 MINOR 2건(N1, N2)을 새로 기록한다.
- 점수(가중): Functionality 90 (PASS) · Security 92 (PASS) · Craft 82 · Consistency 88 → 가중 합계 약 89
- Codex 영수증: `rcpt-a9b987efb30157fd8062d7ba` (이번 감사 단계의 두 번째이자 마지막 호출, 판정 pass, 소견 0건)

## 1. Claim (주장)

1. F1(시임 대체 함수의 공유 슬라이스 경합)은 닫혔다. `migrate_hardening_d_t2053_test.go`의 append를 mutex로 감쌌고, 해당 테스트는 `-race`에서 3회 연속 통과한다.
2. F2(멈춘 엔진이 반환한 뒤에도 미리 읽기 캡처가 계속 돌아 재실행이 잠금에 막힘)는 닫혔다. `manifestPrefetcher.stop()`이 이제 공급 고루틴(`feeder`)과 실행 중인 캡처 전부(`running`)를 기다린다. 새 테스트는 수리를 되돌린 변이체에서 실패하고(2, 2, 1개가 반환 시점까지 실행 중), 수리본에서는 통과한다. 레인 기록의 RED 수치(1, 3, 2)와 같은 성격의 결과다.
3. 수리는 교착을 만들지 않는다. `stop()`이 `quit`을 먼저 닫으므로 슬롯이 가득 찬 채 막혀 있던 공급 고루틴도 `select`의 `<-pf.quit` 분기로 빠져나온다. `running.Wait()`는 `<-pf.feeder` 뒤에 오므로 모든 `Add(1)`이 `Wait`보다 먼저 일어난다(WaitGroup 오용 없음). 캡처마다 걸리는 시간은 git 호출별 시간 제한이 묶는다(`captureTreeManifest`는 시간 초과 시 즉시 중단).
4. 수리는 저널·이동 목록·재키·재바인딩·비교의 순서를 바꾸지 않는다. 델타가 건드린 생산 코드는 `manifestPrefetcher` 구조체, `newManifestPrefetcher`, `stop()` 세 곳뿐이다.
5. 이 카드는 범위 밖인 `TestApply_PendingIntentButNeverMoved_IsSettledThenMoved`(t2106)를 더 나쁘게 만들지 않는다. 그 테스트가 기대는 `revLines`/`appendMoveListLine`(movelist.go)을 수리 델타가 건드리지 않았다.
6. 소스 스캔 가드 테스트 중 이 카드 변경으로 실패하는 것은 없다. `TestHomeGuard_RepositoryHasNoNewHomeAccess`는 미리 알려진 두 위반(이 카드 소유가 아닌 파일)만 보고한다.

## 2. Evidence (명령 + 출력 원문)

### 2.1 HEAD와 수리 델타

```
$ git -C <wt> rev-parse HEAD --show-toplevel
c1827b7c4dc58b9313c4476ae4de4bd0e91cc082
/Users/goos/.moai/worktrees/projects/moai-adk-go/t2081

$ git -C <wt> diff --stat 3be3e51f36 c1827b7c4d
 .moai/reports/t2081/progress.md                    |  21 +-
 .moai/reports/t2081/sync-audit.md                  | 212 +++++++++++++++++++++
 internal/cli/worktree/migrate.go                   |  20 +-
 .../cli/worktree/migrate_hardening_d_t2053_test.go |   6 +
 .../cli/worktree/migrate_parallel_t2081_test.go    |  36 ++++
 5 files changed, 291 insertions(+), 4 deletions(-)

$ git -C <wt> diff --stat 3bc707a906 c1827b7c4d
 .moai/reports/t2081/progress.md   |  21 +++-
 .moai/reports/t2081/sync-audit.md | 212 ++++++++++++++++++++++++++++++++++++++
 2 files changed, 232 insertions(+), 1 deletion(-)
```

수리 코드(`migrate.go:1129-1190`): `feeder chan struct{}`(공급 고루틴이 `defer close(pf.feeder)`), `running sync.WaitGroup`(각 캡처 고루틴 앞 `Add(1)`, 안에서 `defer Done()`), `stop()` = `once.Do(close(quit))` → `<-pf.feeder` → `pf.running.Wait()`. 호출 지점은 `migrate.go:955`의 `defer prefetch.stop()` 하나뿐이고, `migrate.go:944`의 `defer progress.end()`보다 늦게 등록되어 먼저 실행된다. 엔진은 호출자의 `withMigrationLock` 안에서 돌기 때문에 잠금이 풀리기 전에 캡처가 모두 끝난다.

### 2.2 F2 수리 확인 (`-race`, 3회)

```
$ GOCACHE=<scratch>/gocache-t1702 go test -C <wt> ./internal/cli/worktree -race -run 'TestMigrationEngine_ReturningEarlyJoinsTheReadAheadCaptures|TestMigrationEngine_StoppedAfter' -count=3 -v
--- PASS: TestMigrationEngine_StoppedAfterAWholeBatchIsPreparedThenRerunMovesEverything (8.37s)
--- PASS: TestMigrationEngine_StoppedAfterABatchMovedThenRollbackRestoresEveryTree (7.69s)
--- PASS: TestMigrationEngine_ReturningEarlyJoinsTheReadAheadCaptures (7.16s)
--- PASS: TestMigrationEngine_StoppedAfterAWholeBatchIsPreparedThenRerunMovesEverything (9.80s)
--- PASS: TestMigrationEngine_StoppedAfterABatchMovedThenRollbackRestoresEveryTree (5.73s)
--- PASS: TestMigrationEngine_ReturningEarlyJoinsTheReadAheadCaptures (3.72s)
--- PASS: TestMigrationEngine_StoppedAfterAWholeBatchIsPreparedThenRerunMovesEverything (6.23s)
--- PASS: TestMigrationEngine_StoppedAfterABatchMovedThenRollbackRestoresEveryTree (4.57s)
--- PASS: TestMigrationEngine_ReturningEarlyJoinsTheReadAheadCaptures (6.18s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	62.494s
```

### 2.3 새 테스트가 실패할 수 있는지 (변이체, 워크트리 밖 overlay)

변이: `stop()`에서 `<-pf.feeder`와 `pf.running.Wait()`를 지워 수리 이전 본문(`once.Do(close(quit))`만)으로 되돌린 사본을 scratch에 두고 `-overlay`로 바꿔 끼웠다. 워크트리 파일은 건드리지 않았다.

```
$ GOCACHE=<scratch>/gocache-t1702 go test -C <wt> -overlay <scratch>/overlay-nojoin.json ./internal/cli/worktree -run 'TestMigrationEngine_ReturningEarlyJoinsTheReadAheadCaptures' -count=3 -v
=== RUN   TestMigrationEngine_ReturningEarlyJoinsTheReadAheadCaptures
    migrate_parallel_t2081_test.go:518: 2 manifest captures were still running when the engine returned
--- FAIL: TestMigrationEngine_ReturningEarlyJoinsTheReadAheadCaptures (13.76s)
=== RUN   TestMigrationEngine_ReturningEarlyJoinsTheReadAheadCaptures
    migrate_parallel_t2081_test.go:518: 2 manifest captures were still running when the engine returned
--- FAIL: TestMigrationEngine_ReturningEarlyJoinsTheReadAheadCaptures (6.72s)
=== RUN   TestMigrationEngine_ReturningEarlyJoinsTheReadAheadCaptures
    migrate_parallel_t2081_test.go:518: 1 manifest captures were still running when the engine returned
--- FAIL: TestMigrationEngine_ReturningEarlyJoinsTheReadAheadCaptures (5.67s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	26.941s
FAIL
```

레인 기록의 RED(1, 3, 2)와 이번 측정(2, 2, 1)은 같은 방향이다. 수치가 다른 것은 캡처가 150ms 잠드는 동안의 스케줄링 차이로 보이며, 판정에 영향은 없다.

테스트가 타이밍으로 통과하는지: GREEN 쪽은 결정적이다. `running.Wait()`가 반환 전에 모든 캡처를 합류시키므로 `active`는 0일 수밖에 없다. 타이밍에 기대는 것은 RED 쪽(늦게 시작된 캡처가 반환 시점에 아직 자고 있음)뿐이고, 그 쪽이 흔들리면 검출력이 떨어질 뿐 거짓 실패는 나지 않는다. `started > moveConcurrency` 전제는 lookahead 3 + 배치 4 구성에서 `get(0..3)`이 슬롯을 차례로 풀어 7개 중 최소 6개 캡처가 시작되므로 안정적으로 성립한다(이번 6회 실행 모두 이 전제 실패 없음).

### 2.4 F1 수리 확인 (`-race`, 3회)

```
$ GOCACHE=<scratch>/gocache-t1702 go test -C <wt> ./internal/cli/worktree -race -run 'TestMigrateApply_ForwardFailure_UnreadableDestinationIdentity_StaysPending' -count=3 -v
--- PASS: TestMigrateApply_ForwardFailure_UnreadableDestinationIdentity_StaysPending (3.14s)
--- PASS: TestMigrateApply_ForwardFailure_UnreadableDestinationIdentity_StaysPending (3.85s)
--- PASS: TestMigrateApply_ForwardFailure_UnreadableDestinationIdentity_StaysPending (3.54s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	12.667s
```

1회차에서는 같은 명령이 `WARNING: DATA RACE` … `FAIL`이었다(`sync-audit.md` 2.4).

### 2.5 서식·정적 점검 (Craft / Consistency)

```
$ gofmt -l <wt>/internal/cli/worktree <wt>/internal/config
(출력 없음)
$ GOCACHE=<scratch>/gocache-t1702 go vet -C <wt> ./internal/cli/worktree ./internal/config
(출력 없음)
$ GOOS=windows GOCACHE=<scratch>/gocache-t1702 go vet -C <wt> ./internal/cli/worktree ./internal/config
(출력 없음)
```

### 2.6 가드 테스트 수색 (필수)

카드가 바꾼 파일:

```
$ git -C <wt> diff main...HEAD --name-only
.moai/reports/t2070/probe-end-to-end.go.txt
.moai/reports/t2070/probe-stage-timing.go.txt
.moai/reports/t2070/progress.md
.moai/reports/t2070/sync-audit.md
.moai/reports/t2081/probe-end-to-end-width.go.txt
.moai/reports/t2081/progress.md
.moai/reports/t2081/sync-audit.md
internal/cli/worktree/migrate.go
internal/cli/worktree/migrate_hardening_d_t2053_test.go
internal/cli/worktree/migrate_parallel_t2081_test.go
internal/cli/worktree/migrate_progress.go
internal/cli/worktree/migrate_speed_t2070_test.go
internal/config/defaults.go
```

찾은 방법 (이름이 아니라 무엇을 훑는지로 골랐다):

1. `grep -rlE "go/parser|parser\.ParseFile|parser\.ParseDir|filepath\.WalkDir|filepath\.Walk\(" internal --include='*_test.go'` — 소스를 파싱하거나 디렉터리를 걷는 테스트 전체(약 190개 파일).
2. 그중 걷는 루트를 확인(`grep -nE "WalkDir\(|...|root *:?= |range \[\]string|..."`)해서 `internal/` 전체, `internal/cli`, `internal/config`를 걷는 것만 남겼다. 템플릿·고정 디렉터리·`t.TempDir()`만 걷는 것은 제외했다.
3. `grep -rlE "defaults\.go|DefaultWorktreeMigrate|migrate\.go|..." internal --include='*_test.go'` — 바뀐 파일·상수 이름을 직접 언급하는 테스트. `state_backup_past_keep_ship_test.go`(defaults.go 단일 선언 확인)와 카드 자신의 테스트, `merge_test.go`(문자열 픽스처)뿐이었다.
4. `grep -rnE "ReadFile\(\"[a-z_]+\.go\"|..."` — 소스 파일을 이름으로 읽는 테스트. `migrate.go`/`defaults.go`를 읽는 것은 없었다.
5. `internal/cli/worktree`, `internal/config`, `internal/graph`, `internal/worktree`에서 `go/ast|go/parser|ReadDir(".")|Glob("*.go")|packages.Load` — `config`의 `audit_plan_test.go`(ImportsOnly, 별개 파일), `shipped_key_reader_test.go`(types.go만), `anthropic_env_ssot_test.go`만 걸렸다. `cli/worktree` 안의 WalkDir 세 곳(`migrate_review_fix`, `migrate_legacy_manifest`, `hoist_tracked_t2013`)은 픽스처 디렉터리를 걷는다.

해당 목록과 실행:

| 테스트 | 패키지 | 훑는 범위 | 결과 |
|---|---|---|---|
| TestHomeGuard_RepositoryHasNoNewHomeAccess | internal/testguard | 저장소 전체 테스트 파일 파싱 | FAIL — 알려진 두 위반만(아래) |
| TestFixturePackagesScrubProcess | internal/gitenv | internal·pkg·cmd 테스트 디렉터리 | ok |
| TestPinLiteralSweep_RealTreeClean | internal/config | internal/ | ok |
| TestWorkflowWorktreeKeyDocsMatchReaders | internal/config | internal/ | ok |
| TestAC_UBR_10_BudgetKeyShipsOnce | internal/config | internal/ (defaults.go 단일 선언) | ok |
| TestNoBareAnthropicEnvVarLiteralsInProduction | internal/config | 생산 소스 파싱 | ok |
| TestNoBareOSWriteFileToConfigPaths | internal/config/atomicfile | internal/config, internal/cli | ok |
| TestBacklogJSONLiteralStaysSeamScoped | internal/factory | internal·pkg·cmd | ok |
| TestHomeJoinSiteCountIsPinned | internal/hook | 저장소 루트 | PASS |

```
$ GOCACHE=<scratch>/gocache-t1702 go test -C <wt> ./internal/testguard ./internal/gitenv ./internal/config ./internal/config/atomicfile ./internal/factory -run 'TestHomeGuard_RepositoryHasNoNewHomeAccess|TestFixturePackagesScrubProcess|TestPinLiteralSweep_RealTreeClean|TestWorkflowWorktreeKeyDocsMatchReaders|TestAC_UBR_10_BudgetKeyShipsOnce|TestNoBareAnthropicEnvVarLiteralsInProduction|TestNoBareOSWriteFileToConfigPaths|TestBacklogJSONLiteralStaysSeamScoped' -count=1
--- FAIL: TestHomeGuard_RepositoryHasNoNewHomeAccess (13.07s)
    repo_guard_test.go:87: NEW home access internal/agenticfactory/core/sanitize_test.go:95: ~/ literal: "~/work/notes.md", // a tilde path names no user
          Use testguard.RedirectHome(t) or testguard.RealHome(), or add `// home-guard:allow <reason>` on the line or the line above.
    repo_guard_test.go:87: NEW home access internal/cli/factory_fixture_home_test.go:36: ~/ literal: t.Errorf("MoaiHome() = %q resolves outside the test temp dir %q: factory tests would write project folders into the real ~/.moai", home, tmp)
          Use testguard.RedirectHome(t) or testguard.RealHome(), or add `// home-guard:allow <reason>` on the line or the line above.
    repo_guard_test.go:96: scanned 3925 test files: 162 violations, seed 184 entries, 2 new, 24 stale seed entries (stale entries do not fail; regenerate with TESTGUARD_WRITE_SEED=1)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/testguard	13.360s
ok  	github.com/modu-ai/moai-adk/internal/gitenv	8.098s
ok  	github.com/modu-ai/moai-adk/internal/config	17.067s
ok  	github.com/modu-ai/moai-adk/internal/config/atomicfile	1.990s
ok  	github.com/modu-ai/moai-adk/internal/factory	3.031s
FAIL

$ GOCACHE=<scratch>/gocache-t1702 go test -C <wt> ./internal/hook -run 'TestHomeJoinSiteCountIsPinned' -count=1 -v
=== RUN   TestHomeJoinSiteCountIsPinned
--- PASS: TestHomeJoinSiteCountIsPinned (1.74s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	2.910s
```

`TestHomeGuard_RepositoryHasNoNewHomeAccess`의 새 위반은 정확히 두 건이고, 둘 다 이 카드가 바꾸지 않은 파일(`internal/agenticfactory/core/sanitize_test.go:95`, `internal/cli/factory_fixture_home_test.go:36`)이다. 지시에 적힌 main의 기존 실패와 같으므로 이 카드의 결함이 아니라 기존 실패로 기록한다. 패키지마다 `[no tests to run]` 표시가 없으니 지정한 테스트가 실제로 돌았다.

### 2.7 Codex (이번 감사 단계 두 번째 호출, 1회)

근거: 보드 결정 `d-20261106T080039Z-305c`(standing, `moai decision read --scope card:t2081`로 직접 읽음) — 감사 단계마다 최대 2회, 두 번째는 수리 델타 한정.

호출: `mcp__moai__codex_audit`, `project_root=/Users/goos/.moai/worktrees/projects/moai-adk-go/t2081`, `target=baseBranch`, `delta_base=3be3e51f36402c6ab8e01f7958b21f6f9cd2ab77`, `mode=native`. 재시도 없음.

반환(요약 원문 일부): `"verdict":"pass"`, `"findings":[]`, summary "지정한 기준 커밋 이후 변경에서 지적할 결함은 찾지 못했노라. 정적 검토만 했으며 테스트·빌드·재현 명령은 실행하지 않았어.", `"audit_receipt":"rcpt-a9b987efb30157fd8062d7ba"`, `"build_commit":"04d8c27a7e"`, `"review_base":"delta_base (merge base 3be3e51f36402c6ab8e01f7958b21f6f9cd2ab77)"`, resolved_model `gpt-6.1-sol`, `error_code` 빈 값. `gate_unmet` 없음.

### 2.8 감사 계획

```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2081
"config_status": "ok", "model": "codex", ... claude gate off, codex gate required, glm gate off,
"cross_model_active": true, "cross_model_required": true, "enforced_required": ["codex"]
```

계획은 `audit_multi`를 가리키지만, 운영자 결정(`d-20261106T080039Z-305c`: "app-server 직접·audit_multi 금지")에 따라 `codex_audit` 단일 호출로 codex 게이트를 채웠다. 필수 게이트는 codex 하나이고 그 판정은 pass다.

## 3. Baseline-attribution (무엇을 기준으로 쟀나)

- 모든 테스트·vet·gofmt는 이번 실행에서 이 워크트리(HEAD `c1827b7c4d`, `git status --short` 비어 있음)에 대해 Go 툴체인(`go env GOVERSION` = go1.27.3)으로 직접 돌렸다. 빌드 캐시는 트리 밖 scratch(`gocache-t1702`)다. 변이체는 overlay로만 끼웠고 워크트리 파일은 바뀌지 않았다.
- 1회차 증거(F1 DATA RACE, F2 잠금 실패)는 `sync-audit.md`(감사 SHA `3be3e51f36`)를 기준으로 비교했다.
- Codex 리뷰는 MCP 서버 빌드 `04d8c27a7e`가 트리의 git diff(`3be3e51f36...HEAD`)를 읽은 결과다. 이 감사는 moai CLI 결과를 판정 근거로 쓰지 않았고(감사 계획 출력은 경로 선택에만 사용), Codex는 git diff를 읽으므로 빌드 지연은 판정에 영향이 없다고 본다.

## 4. 소견 (structured defect list)

### 1회차 소견의 처리

| ID | 1회차 심각도 | 현재 상태 | 판단 |
|---|---|---|---|
| F1 | BLOCKING | 닫힘 | mutex로 수리, `-race` 3/3 통과(2.4). 같은 시임을 바꾸는 다른 테스트는 Rename/Chmod만 한다는 레인 판단은 1회차 감사와 일치. Codex 원 지적도 이번 리뷰에서 재지적 없음. |
| F2 | BLOCKING | 닫힘 | `stop()` 합류로 수리, 변이체 RED → 수리본 GREEN(2.2, 2.3). 원인(fork~exec 사이 잠금 fd 공유)은 여전히 가설이지만, 반환 뒤에 엔진이 새 git 자식을 띄우지 않는다는 성질 자체는 테스트로 고정됐다. |
| F3 | SHOULD-FIX | 정정됨 | progress.md:55가 "카드 PR CI는 ubuntu만, 3-OS는 릴리스 PR"로 고쳐졌고 Linux·Windows 미관측을 Gap으로 명시. 수용. |
| F4 | SHOULD-FIX(가설) | 남음 → 부채 | `worktree.useRelativePaths=true`에서의 config.lock 경합 가설. 관측된 적 없고, 일어나도 엔진은 해당 트리를 pending/reversed로 다룬다(fail-safe). merge-ready를 막을 근거는 아니지만 부채로 남긴다. |
| F5 | MINOR | 남음 | `journal.settle` 오류 시 형제 최대 3개가 "이동됐지만 마무리 안 됨"으로 남는 경로 미시험. `--rollback`으로 수렴하는 상태라 수용 가능. |
| F6 | MINOR | 남음 | rollback 테스트가 move-list·레지스트리·저널을 확인하지 않음. 시험 보강 사항, 수용 가능. |
| F7 | MINOR | 남음 | 같은 basename/중첩 트리가 한 배치에 들 때 순서 비결정. 결과는 안전 방향이고 이 저장소에 해당 등록 없음. 수용 가능. |
| F8 | MINOR | 남음 | `readerFailures` 미단언. 사소함, 수용 가능. |
| F9 | MINOR | 남음 | 진행 표시가 배치당 한 트리만 보임. 표시 문제, 수용 가능. |

### 이번에 새로 기록한 소견

| ID | 심각도 | 분류 | 위치 태그 | 요약 | 위치 |
|---|---|---|---|---|---|
| N1 | MINOR | optional | in-diff | `stop()` 주석의 "per-call git timeout of the captures in flight"는 실제 상한보다 좁게 쓰였다. 캡처 하나가 git을 여러 번 부르므로 기다림의 상한은 (캡처당 git 호출 수 × 호출별 시간 제한)이다. 주석 정정 정도의 사안. 신뢰도 중. | `internal/cli/worktree/migrate.go:1180-1185` |
| N2 | MINOR | optional | in-diff | `quit`이 닫힌 뒤에도 슬롯이 비어 있으면 Go `select`가 두 분기 중 무작위로 골라 공급 고루틴이 캡처를 더 시작할 수 있다(최대 lookahead 폭, 슬롯은 반환 후 풀리지 않으므로 유한). `stop()`이 그것까지 기다리므로 정확성 문제는 없고, 조기 반환이 캡처 하나 몫만큼 늦어질 수 있다. 루프 앞에서 `quit`을 먼저 확인하면 없어진다. 신뢰도 높음(언어 규칙), 영향 낮음. | `internal/cli/worktree/migrate.go:1154-1160` |

교착·누수·순서 점검 결과(지시 항목):

- 공급 고루틴이 가득 찬 슬롯 채널에서 막혀 있을 때 `stop()`: `quit`을 먼저 닫으므로 `select`가 빠져나온다. 교착 없음.
- 취소되지 않는 컨텍스트에서 막힘: 공급 고루틴은 컨텍스트를 쓰지 않는다. 캡처 쪽은 git 호출별 시간 제한으로 끝난다.
- 오류·조기 반환 경로의 고루틴/git 자식 누수: 이동 고루틴은 `wg.Wait()`(1회차 확인), 캡처 고루틴은 이제 `stop()`이 합류. 엔진의 모든 반환 경로(`berr`, `jerr`, 정상 종료, 패닉)에서 `defer`가 실행된다. 계획 단계 오류(`planErr`)는 prefetcher 생성 전에 반환한다.
- 저널/이동 목록 효과의 순서: 수리 델타가 건드리지 않았다(2.1).
- 타이밍으로 통과하는 테스트: GREEN 쪽은 결정적(2.3).

## 5. Gaps (관측하지 않은 것)

- 거부된 도구 호출: 임시 `HOME`을 지정한 두 `go test` 명령이 워크트리 가드에 거부됐다(원문: "This session is isolated in the worktree /Users/goos/.moai/worktrees/projects/moai-adk-go/t2081, but this command sets HOME, injecting git configuration whose effect on where git writes can't be verified. Refusing to run it ..."). 대신 `HOME`을 지정하지 않고 같은 테스트를 돌렸다. 이 테스트들은 자체적으로 `t.TempDir()`/`testguard.RedirectHome`로 홈을 격리하며, 이 감사는 실제 `~/.claude`·`~/.moai`를 직접 읽거나 쓰지 않았다. 다만 테스트 프로세스가 실제 HOME을 환경으로 받은 것은 사실이다.
- 레인이 기록한 `-race -count=15`(30회, 186.9 s)와 `-count=20`(105.4 s)은 재현하지 않았다. 이 감사의 `-race` 반복은 3회다. 1회차 실패율(1/6~6/10)을 생각하면 3회 무실패는 수리를 지지하지만 증명은 아니다.
- `<-pf.feeder` 합류만 지운 두 번째 변이체(Add와 Wait의 순서 보장 검사)는 돌리지 않았다. 그 줄이 하중을 받는다는 판단은 코드 읽기에 근거한다.
- F2 원인 가설(fork~exec 사이 fd 상속)은 검증하지 않았다.
- 패키지 전체, `go test ./...`, golangci-lint는 돌리지 않았다(지시). 시임을 바꾸는 다른 테스트 전부를 `-race`로 다시 돌리지도 않았다.
- `TestApply_PendingIntentButNeverMoved_IsSettledThenMoved`(t2106)는 돌리지 않았다. "악화 없음" 판단은 수리 델타가 movelist.go를 건드리지 않았다는 diff 근거뿐이다.
- Linux·Windows git에서는 아무것도 돌리지 않았다(이 감사도 macOS만). Windows는 `go vet`만 했다.
- 성능 수치(23%/37%)는 재측정하지 않았다.
- `audit_multi` 다이제스트(`overall_verdict`/`gate_unmet`/`plan_source`)는 없다. 운영자 결정이 `audit_multi`를 금지했고, 필수 게이트(codex)는 `codex_audit` 결과(pass, 영수증 위)로 채웠다. 오케스트레이터의 `--result-file` 검사는 이 경로에 해당 입력이 없다.

## 6. Residual-risk (관측했는데도 남는 위험)

- 동시 이동의 안전성 근거는 여전히 macOS·Apple git 2.54.0·기본 config에서의 측정이다. Linux·Windows와 `useRelativePaths`(F4)는 관측되지 않았다. 실패는 fail-safe 경로로 가지만 운영자에게 pending/rollback 작업을 남긴다.
- kill 한 번이 남길 수 있는 미정리 트리 수가 1에서 최대 4로 늘었다(문서에 명시).
- F2의 실제 원인이 가설과 다르다면(엔진 반환 뒤 남는 다른 백그라운드 작업), 같은 프로세스에서 엔진을 다시 부르는 경로에서 잠금 충돌이 드물게 남을 수 있다. 이번 3회와 레인 기록의 반복 실행에서는 재발하지 않았다.
- 조기 반환 시 `stop()`이 캡처 합류를 기다리므로, git이 느린 환경에서 중단 후 반환이 최대 (lookahead 폭 × 캡처 시간)만큼 늦어진다. 의도된 교환이다.
- Codex 두 번째 리뷰는 정적 검토였다(스스로 밝힘). 독립 의견은 이 1건이다.

## 7. Iteration history

- 1회차(`3be3e51f36`): FAIL. BLOCKING F1(Codex 동일 지적)·F2, SHOULD-FIX F3·F4, MINOR F5~F10. Codex 영수증 `rcpt-8b29f47630e6e68e12831a06`.
- 2회차(이번, `c1827b7c4d`): PASS-WITH-DEBT. F1·F2 닫힘(재현 증거), F3 정정, F4~F9 부채로 남김, N1·N2 새 MINOR(optional). 수리에서 생긴 BLOCKING 없음(수리 중단 규칙 해당 없음). Codex 영수증 `rcpt-a9b987efb30157fd8062d7ba`(pass, 소견 0).

## 8. 부채 목록 (merge-ready 뒤 후속 카드 후보)

- F4: `git -c worktree.useRelativePaths=true` 변형 스트레스 라운드 또는 git 소스 확인 기록.
- F5·F6: settle 오류 경로 테스트, rollback 뒤 move-list·레지스트리·저널 확인.
- Linux·Windows 관측(릴리스 PR의 3-OS 매트릭스에서 증명 테스트 결과 확인).
- N1·N2: 주석 정정, 루프 앞 `quit` 확인(선택).
