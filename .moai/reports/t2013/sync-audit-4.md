auditor-model: claude-opus-5-5
verdict: FAIL
audited_sha: 6cfedbcf5e1af1326eebb985779f54a7cae95a68
score: 0.75

# Sync 감사 4회차(운영자 승인 확인 감사) — 카드 t2013 (SPEC-FACTORY-COMPLETION-RECOVERY-001 개정)

- 트리: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2013`, 브랜치 `WT-agent-trunk-factory`
- 감사 대상: `git rev-parse 6cfedbcf5e` → `6cfedbcf5e1af1326eebb985779f54a7cae95a68` (= HEAD). 작업 트리에는 `.moai/reports/t2013/progress.md` 수정 1건만 있음(보고서 파일, Go 코드 아님).
- 1차 범위(델타): `6cb0c11f32..6cfedbcf5e` 5커밋 — 79218493e9(감사 보고서 2·3 기록), 1f0a22acc8(RED)·96722fecde(GREEN) C1, 568269325b(RED)·6cfedbcf5e(GREEN) C3.
- 2차 범위: 카드 범위 `cb30e05b3e..6cfedbcf5e` 가운데 두 수리가 만든 회귀만.
- 승인: 결정 보드 d-20261103T002644Z-9259의 확인 감사 1회. 이후 감사 없음.
- 감사계획(`moai verify audit-plan --project-root <tree>`): `config_status: ok`, `model: codex`, codex `required`, claude `off`, glm `off`, `cross_model_required: true`, `enforced_required: ["codex"]`.

## 종합 판정

**FAIL (0.75).**

1. **이번 확인 대상인 C1·C3는 해소됐고, 두 수리가 만든 회귀는 찾지 못했다.** F1~F4도 회귀하지 않았다. 지정된 시험을 직접 돌려 확인했다(cli 지정 선택자 PASS·FAIL 0, hook 218건 PASS·FAIL 0, 착지 간선 깨우기 시험 PASS, golangci-lint `0 issues.`).
2. 그러나 **required codex 다리가 다시 `fail`을 냈다**(receipt `rcpt-6ea9449ffc3f0730594f9a22`). 지적 3건은 모두 이번 두 수리가 건드리지 않은 카드 범위의 기존 코드에 있고, 1~3회차에서 나오지 않은 새 지적이다. required 다리의 FAIL은 PASS를 막는다(3회차와 같은 처리).
3. 3건을 소스로 직접 대조했다. 2건(K1: BranchGuard `eval` 우회, K3: 운영자 대기를 닫는 경로 없음)은 정확성 결함으로 확인되어 blocking으로 분류했다. K1은 기준점 cb30e05b3e가 거절하던 명령을 카드가 새로 허용하게 만든 회귀다. K2(변경 경로 인용 처리)도 확인됐지만 영향 범위가 좁아 optional로 분류했다.

Must-pass firewall: Functionality(blocking K1·K3)와 Security(High K1)가 FAIL이므로 다른 점수와 상관없이 전체 FAIL이다.

## 해소 확인표

| 항목 | 수리 커밋 | 판정 | 근거 |
|---|---|---|---|
| C1 착지 뒤 단계 선행 카드에 대한 대기 | 1f0a22acc8 / 96722fecde | **해소** | `factoryLandedState`(factory_wait.go:38-45)가 `merged-local`·`merged-pr`·`pushed`·`ci-green`·`done`을 참으로 본다. 이 집합은 T2 `after` 가드 `predecessorMerged`(homestate/card_picked.go:301 `case CardMergedLocal, CardMergedPR, CardPushed, CardCIGreen, CardDone:`)와 정확히 같다. 시험 `TestFactoryWaitPredecessorPastLandingSettlesAtRecord` 10개 하위 사례 PASS: 착지 이후 5개 상태는 기록 시점에 깨우기 1줄 + `woken_at` 설정, `run`·`merge-ready`·`blocked`·`abandoned`·`failed`는 열린 채 남음. |
| C1 부수효과: 착지 간선 깨우기 | 96722fecde | **정상(회귀 없음)** | 같은 술어가 `factoryPrintWakesSuffixed`(factory_wait.go:51)를 가르지만, 착지 간선 호출부는 두 곳뿐이고 각각 방금 `merged-local`(factory_card.go:3483)·`merged-pr`(factory_card_pr.go:626)로 전이한 카드를 넘긴다. `pushed`/`ci-green`/`done` 전이 경로에는 호출이 없으므로 런타임 동작은 바뀌지 않는다. `woken_at` 가드(card_wait.go:166 `WHERE seq=? AND woken_at=''`)가 한 번만 깨우게 막는다. `factoryWaitAlreadyDelivered`(:291-304)도 같은 술어를 쓰게 되어, 선행 카드가 `pushed`로 넘어간 뒤 같은 대기를 다시 기록하면 새 대기를 열지 않는다(수리 전에는 이 경우 영원히 열린 대기를 새로 만들었다 — 개선). 기존 시험 `TestFactoryWakeOnLanding` 6사례·`TestFactoryWakeTwoWaitsOnePredecessor` PASS. |
| C3 go.mod/go.sum만 바뀐 경우 lint·test 생략 | 568269325b / 6cfedbcf5e | **해소** | `factoryFastCheckPackages`(factory_release_route.go:299-312)가 바뀐 경로 중 `path.Base`가 `go.mod`/`go.sum`/`go.work`/`go.work.sum`이면 `./...`로 넓힌다. 그러면 단계는 build + lint(`go vet ./...` 또는 `golangci-lint run … ./...`) + `go test ./...`가 된다. 시험 `TestT2013FastCheckStepsModuleInputOnlyChangePlansLintAndTest`(go.mod only·go.sum only) PASS, `TestT2013FastCheckStepsDocOnlyChangeKeepsBuildOnly` PASS, 기존 패키지 매핑 시험 13개 하위 사례 PASS. 넓히는 방향은 검사를 늘리기만 한다. |
| F1 BranchGuard 명시 대상 | — (델타가 건드리지 않음) | **회귀 없음** | `git diff --stat 6cb0c11f32..6cfedbcf5e -- internal/hook internal/cli/factory_checkpoint.go internal/homestate` 출력 없음. hook 시험 218 PASS, `TestBranchGuardExplicitTarget*` 4건 PASS. (단, 새 codex 지적 K1은 F1과 같은 계열의 남은 구멍이다 — 아래.) |
| F2·F3 checkpoint | — | **회귀 없음** | 같은 diff 출력 없음. `TestFactoryCheckpoint*` 8건(하위 3건 포함) PASS. |
| F4 / D16 | — | **회귀 없음** | 장부·인용 변경 없음(델타의 문서 변경은 sync-audit-2·3 보고서 추가뿐). |
| C2 카드 임대 갱신(optional) | — | **변화 없음** | factory_card.go 미변경. 3회차 판정(안전한 거절, optional) 그대로. |
| C4 lint 도구 버전(optional) | — | **변화 없음** | factory_release_route.go:228 `toolVersionCmd: []string{"go", "version"}` 그대로. C3 수리로 lint가 `./...` 전체에 도는 경우가 늘었지만 도구 식별 결함 자체는 같은 크기. |
| O1 `cd /wt && git -C /primary commit`(optional) | — | **변화 없음** | branch_guard.go 미변경. |

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 70/100 | FAIL | C1·C3 해소와 F1~F4 무회귀를 시험으로 확인(cli `ok … 72.495s`, 착지 간선 `ok … 2.952s`, hook `ok … 40.996s`). 감점: required codex 다리 fail, blocking으로 확인된 K1·K3(카드 범위 기존 코드). |
| Security (25%) | 70/100 | FAIL | 델타는 보안 표면을 건드리지 않음(새 SQL 없음, 경로는 `path.Base` 비교뿐). 감점: K1 [High] — 주 체크아웃 보호 커밋 거절을 `eval '…cd…'`로 우회. 2회차 F1(High)과 같은 계열이므로 같은 심각도로 분류. |
| Craft (20%) | 85/100 | PASS | RED→GREEN 커밋 쌍 유지(RED 커밋은 시험 파일만 추가), `gofmt -l` 출력 없음, `go vet ./internal/cli/` 출력 없음, `golangci-lint run ./internal/cli/` `0 issues.`. 감점: `go.work`·"소유 패키지 없는 .go" 분기는 시험이 없음, `factoryChangedGoPackages`가 이제 시험에서만 쓰이는 래퍼로 남음. |
| Consistency (15%) | 90/100 | PASS | C1 술어가 homestate의 T2 가드 집합과 일치, 주석이 새 의미를 정확히 서술. 새 함수 이름·주석 밀도가 파일 관례를 따름. |

가중 조화평균: 0.40/70 + 0.25/70 + 0.20/85 + 0.15/90 = 0.005714 + 0.003571 + 0.002353 + 0.001667 = 0.013305 → 75.2 → **0.75**.

## Codex 다리 (mcp__moai__codex_audit, 1회)

- 호출: `codex_audit(project_root=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2013, target=baseBranch, mode=native)`
- 결과: `verdict: fail`, `audit_receipt: rcpt-6ea9449ffc3f0730594f9a22`, `build_commit: b1806498fe`, `review_base: main (merge base cb30e05b3e5f64774b8dabd87e80c24332f4551d)`. `gate_unmet` 필드 없음(verdict 자체가 `fail`).
- 3회차 지적 C1·C3는 이번 결과에 다시 나오지 않았다.
- `audit_multi`는 호출하지 않았다(지시가 codex_audit 단일 호출을 명시). 오케스트레이터가 `moai verify audit-plan --result-file`로 확인할 값: per-backend `{backend: codex, gate: required, verdict: fail}`. 그 확인 전까지 이 판정은 최종이 아니다.

| codex | 내용 | 이 감사의 소스 대조 | 분류 |
|---|---|---|---|
| K1 P1 `internal/hook/branch_guard.go:479` | 주 체크아웃에서 `cd /worktree && eval 'cd /primary' && git commit -m x`를 실행하면 따옴표 안의 두 번째 `cd`가 검사에서 사라져 대상이 `/worktree`로 판정되고 커밋이 허용된다. | **확인.** `explicitGitTarget`(:472-498)은 선행 cd 뒤 나머지를 `substituteQuotedArguments`(:206-208, 따옴표 구간을 ` X `로 바꿈)로 처리한 뒤 `dirChangePattern`(`\b(?:cd\|pushd\|popd)\b\|\(`)을 찾는다. `eval  X  && git commit -m x`에는 일치하는 것이 없으므로 `/worktree`를 대상으로 채택하고 그 경로에서 primary 판정을 한다 → worktree로 판정 → 허용. 기준점 cb30e05b3e는 훅 cwd(primary)만 봤으므로 같은 명령을 거절했다. 즉 카드(fa276e7b1b/f722ee92c5)가 새로 만든 허용이다. 같은 형태의 `source`/`.` 스크립트도 판독상 같은 결과다. 이 세션이 primary 체크아웃이라 worktree 격리 가드는 이 형태를 막지 않는다. | blocking (High, 보안·정확성) |
| K2 P2 `internal/cli/factory_release_route.go:287` | `git diff --name-only`의 기본 `core.quotePath` 출력은 비ASCII 경로를 따옴표·8진 이스케이프로 감싼다. `strings.Fields`가 이를 풀지 않아 소유 패키지를 못 찾고 build만 돈다. | **확인(판독).** `cliGitIn`(factory_card.go:3533-3539)은 `-c core.quotePath=off`도 `-z`도 쓰지 않는다. 인용된 토큰은 `.go"`로 끝나므로 C3 수리의 `.go` 분기에도 걸리지 않는다. 공백 포함 경로도 `strings.Fields`가 쪼갠다. 카드 기존 코드(8d318e63f6/0ea81fbd91)이고, 이번 수리는 같은 줄을 `factoryChangedPaths`로 옮기기만 했다. 영향은 비ASCII·공백 이름 파일만 바뀐 카드로 좁다. | optional (Low-Medium, 정확성) |
| K3 P2 `internal/cli/factory_wait.go:171` | `wait --operator --next …`로 기록한 대기는 운영자가 답한 뒤에도 닫을 길이 없다. | **확인.** `woken_at`을 쓰는 곳은 `SettleLandedWaits`(card_wait.go:148·166)뿐이고 `cond_kind=predecessor-landed`만 고른다. `card_waits`를 갱신하는 다른 코드는 없다(`grep -rln card_waits internal`: factory.go는 DDL, card_wake_events.go는 SELECT만). `--ack`는 delivered 이벤트가 먼저 있어야 해서 거절된다(factory_wait.go:186-192). 그런데 `stopGoalValidWait`(hook_stop_goal_wait.go:75-80)은 `WokenAt == ""`이고 `next`가 있는 대기가 하나라도 있으면 유효한 대기로 보고 stop-goal 차단을 끈다. 그래서 그 카드 트리의 goal 강제는 영구히 꺼지고 checkpoint도 계속 대기로 표시한다. `--operator` 도움말("released by an operator ruling")이 약속한 해제 경로가 구현되지 않았다. | blocking (Medium, 정확성) |

## Findings (structured defect-list)

- **K1 [High] [blocking] `internal/hook/branch_guard.go:472-498`** (신뢰도 높음): 선행 `cd <abs> &&` 대상 채택 판정이 따옴표를 지운 뒤에 디렉터리 변경을 찾기 때문에, `eval '<cd …>'`(및 `source`/`.` 스크립트)가 대상을 다른 곳으로 옮겨도 대상을 채택한다. 기준점 대비 새 허용이다. Required fix: `explicitGitTarget`에서 나머지 명령에 `eval`, `source`, `.`(명령 위치) 같이 문자열이나 파일을 현재 셸에서 실행하는 형태가 있으면 `redirected = true`로 처리해 훅 cwd 판정으로 넘긴다(넘긴 뒤에는 거절만 늘어난다). 주 체크아웃에서 `cd /wt && eval 'cd /primary' && git commit -m x`가 거절되는지 고정하는 시험을 `TestBranchGuardExplicitTargetNotAdoptedWhenLaterSegmentRedirects`에 추가한다.
- **K3 [Medium] [blocking] `internal/cli/factory_wait.go:204-254`, `internal/homestate/card_wait.go`** (신뢰도 중간 — 운영 빈도는 측정하지 않음): `operator-ruling` 대기에 해제 경로가 없다. Required fix: 운영자 판정으로 대기를 닫는 명시적 경로(예: `moai factory wait <card> --resolve` 또는 `--ack`의 운영자 판정 형태)를 추가해 `woken_at`과 이벤트를 남기고, 닫힌 뒤 stop-goal이 다시 차단하는지 시험으로 고정한다. 범위 결정이 필요하면 리더 판단 사안이다.
- **K2 [Low-Medium] [optional] `internal/cli/factory_release_route.go:283-292`** (신뢰도 높음, 재현 안 함): 변경 경로를 `git -c core.quotePath=off diff --name-only -z …`로 읽고 NUL로 나눈다. 비ASCII 이름 파일만 바꾼 fixture 시험을 추가한다.
- **N1 [Low] [optional] `internal/cli/factory_release_route.go:307-309`**: "소유 패키지가 없는 .go 변경 → `./...`" 분기와 `go.work`/`go.work.sum` 분기에 시험이 없다. 넓히는 방향이라 위험은 낮다.
- **N2 [Low] [optional] `internal/cli/factory_release_route.go:277-279`**: `factoryChangedGoPackages`가 이제 시험에서만 호출된다. 시험을 `factoryFastCheckPackages`나 `factoryOwningGoPackages` 직접 호출로 옮기면 래퍼를 지울 수 있다.
- **N3 [Low] [optional] `internal/cli/factory_wait.go:38-45`**: 착지 뒤 `needs-decision`(T21: merged-local/pushed/ci-green → needs-decision)에 있는 선행 카드에 대기를 기록하면 열린 채 남고, T22 재개 전이에는 깨우기 호출이 없다. `needs-decision`은 착지 전 상태에서도 들어오므로 술어에 그냥 넣을 수는 없다. 드문 경로라 optional.
- **N4 [Low] [optional] `internal/cli/factory_wait.go:258-269`**: `factoryLandedPredecessor`는 모든 run의 카드를 본다. 술어가 `done`까지 넓어져, 같은 카드 id가 이전 run에서 `done`이고 현재 run에서 진행 중이면 대기가 기록 시점에 깨어난다. T2 가드도 같은 "모든 run" 의미를 쓰므로 관례와는 일치한다. 같은 id의 재발행이 실제로 일어나는지는 측정하지 않았다.

## Binding run conditions (SPEC progress.md 재확인)

| 부채 | 상태 | 증거 |
|---|---|---|
| D16 | 처분 | progress.md:22 "처분됨 — 카드 t2013 커밋", 3회차 확인 뒤 변경 없음 |
| D22 / D23 / D24 | 처분 | progress.md:24·27·28 |
| D15 / D17 / D18 / D20 | 기록된 이월(리더 승인 D10) | progress.md:21·23·25·26, :30 "D10 이월 승인 기록" — 1~3회차와 같은 처리. 기록과 어긋나는 사실은 찾지 못함 |

## Evidence

| Claim | 명령 | 관측 출력 (verbatim) |
|---|---|---|
| 감사 SHA | `git -C <tree> rev-parse --show-toplevel HEAD 6cfedbcf5e cb30e05b3e` | `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2013` / `6cfedbcf5e1af1326eebb985779f54a7cae95a68` / `6cfedbcf5e1af1326eebb985779f54a7cae95a68` / `cb30e05b3e5f64774b8dabd87e80c24332f4551d` |
| 델타 커밋 | `git -C <tree> log --oneline 6cb0c11f32..6cfedbcf5e` | `6cfedbcf5e [t2013] fix(release): decide the fast check scope from the changed file list so go.mod or go.sum widens it to the module (card t2013)` / `568269325b [t2013] test(release): RED - …` / `96722fecde [t2013] fix(wait): a predecessor past landing (pushed, ci-green, done) settles a wait at record time (card t2013)` / `1f0a22acc8 [t2013] test(wait): RED - …` / `79218493e9 [t2013] docs(reports): sync audit verdicts 2 and 3 …` |
| 델타 파일 | `git -C <tree> diff --stat 6cb0c11f32..6cfedbcf5e` | `6 files changed, 420 insertions(+), 6 deletions(-)` (Go: factory_fastcheck_modinputs_t2013_test.go 64, factory_release_route.go 31, factory_wait.go 16, factory_wait_past_landing_t2013_test.go 51) |
| RED 커밋 = 시험만 | `git -C <tree> show --stat --format=%h 1f0a22acc8 568269325b` | `1 file changed, 51 insertions(+)` (factory_wait_past_landing_t2013_test.go) / `1 file changed, 64 insertions(+)` (factory_fastcheck_modinputs_t2013_test.go) |
| F1~F3 파일 미변경 | `git -C <tree> diff --stat 6cb0c11f32..6cfedbcf5e -- internal/hook internal/cli/factory_checkpoint.go internal/homestate` | 출력 없음 |
| 감사계획 | `moai verify audit-plan --project-root <tree>` | `"config_status": "ok"`, `"model": "codex"`, codex `"gate": "required"`, claude/glm `"gate": "off"`, `"cross_model_required": true`, `"enforced_required": ["codex"]` |
| cli 지정 시험 | `go test -count=1 -v -run '^(TestFactoryWait.*\|TestFactoryRelease.*\|TestT2013FastCheck.*\|TestStopGoal.*\|TestFactoryCheckpoint.*)$' ./internal/cli/` | 최상위 PASS 38건(checkpoint 8, T2013FastCheck 4, FactoryWait 16, StopGoal 10), FAIL 0(출력을 직접 세어 확인). 주요 줄: `--- PASS: TestFactoryWaitPredecessorPastLandingSettlesAtRecord (11.57s)`(하위 merged-local·merged-pr·pushed·ci-green·done·run·merge-ready·blocked·abandoned·failed 모두 PASS) / `--- PASS: TestT2013FastCheckStepsModuleInputOnlyChangePlansLintAndTest (0.56s)`(go.mod_only·go.sum_only PASS) / `--- PASS: TestT2013FastCheckStepsDocOnlyChangeKeepsBuildOnly (0.28s)` / `--- PASS: TestT2013FastCheckPackagesMapNonGoChangesToOwningPackage (4.15s)`(하위 12건 PASS) / `--- PASS: TestFactoryWaitOnAlreadyLandedPredecessorWakesOnceAtRecord (2.40s)` / `--- PASS: TestFactoryWaitIdenticalReRecordAfterRecordTimeWakeIsSilent (1.98s)` / `--- PASS: TestStopGoalWait_WokenOrAbsentWaitBlocks (4.05s)` / `--- PASS: TestFactoryCheckpointPrefersLiveRowOverNewerAbandoned (1.79s)` / `--- PASS: TestFactoryCheckpointAbsentWorktreeReportsNoForeignGitState (2.14s)` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/cli	72.495s` |
| 착지 간선 깨우기 시험 | `go test -count=1 -v -run '^(TestFactoryWake.*\|TestFactoryFastCheck.*\|TestFactoryComplete.*Wake.*)$' ./internal/cli/` | `--- PASS: TestFactoryWakeOnLanding (0.56s)`(하위 6건 PASS, "predecessor only has an expired lease, not landed: nothing" 포함) / `--- PASS: TestFactoryWakeTwoWaitsOnePredecessor (0.04s)` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/cli	2.952s` |
| hook 시험 | `go test -count=1 -v -run '^(TestBranchGuard.*\|TestProtectedCommit.*\|TestMatchProtectedCommitCommand.*)$' ./internal/hook/` | 출력이 커서 도구 결과 파일에 저장됨. 그 파일에 `grep -c -e "--- PASS"` → `218`; `grep -E -e "--- FAIL" -e "^(ok\|FAIL\|PASS)" -e "TestBranchGuardExplicit"` → `--- PASS: TestBranchGuardExplicitTargetDeniesPrimary (0.51s)` / `--- PASS: TestBranchGuardExplicitTargetAllowsWorktree (0.65s)` / `--- PASS: TestBranchGuardExplicitTargetUnchangedForeignForms (0.73s)` / `--- PASS: TestBranchGuardExplicitTargetNotAdoptedWhenLaterSegmentRedirects (0.80s)` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/hook	40.996s` (FAIL 줄 없음) |
| 서식 | `gofmt -l internal/cli/factory_wait.go internal/cli/factory_release_route.go internal/cli/factory_wait_past_landing_t2013_test.go internal/cli/factory_fastcheck_modinputs_t2013_test.go` | 출력 없음 |
| vet | `go vet ./internal/cli/` | 출력 없음 |
| lint | `golangci-lint run --allow-parallel-runners ./internal/cli/` | `0 issues.` |
| C1 술어 일치 | `sed -n 290,310p <tree>/internal/homestate/card_picked.go` | `case CardMergedLocal, CardMergedPR, CardPushed, CardCIGreen, CardDone:` / `return nil` |
| 착지 간선 호출부 | `grep -rn -e factoryLandedState -e factoryPrintWakes -e factoryLandedPredecessor <tree>/internal` | 비시험 호출부: `factory_card.go:3483: factoryPrintWakes(ctx, out, db, done)`, `factory_card_pr.go:626: factoryPrintWakes(ctx, out, db, merged)`, factory_wait.go 내부 285·299 |
| 전이 표 | `sed -n 130,175p <tree>/internal/homestate/card_transition.go` | `{"T17", CardMergedLocal, CardPushed, guardPush},` / `{"T18", CardMergedLocal, CardDone, guardNoRemote},` / `{"T20", CardCIGreen, CardDone, guardApprovalDone},` / T21 `… CardMergedLocal, CardPushed, CardCIGreen` → needs-decision |
| K1 근거 | Read branch_guard.go:173-208, :440-498 | `var quotedArgumentPattern = regexp.MustCompile(\`'[^']*'\|"[^"]*"\`)` / `const quotedArgumentPlaceholder = " X "` / `dirChangePattern = regexp.MustCompile(\`(?i)(?:\b(?:cd\|pushd\|popd)\b\|\()\`)` / `after := substituteQuotedArguments(rest[m[1]:])` / `if dirChangePattern.MatchString(after) \|\| cOptionPattern.MatchString(after) {` |
| K1 카드 출처 | `git -C <tree> log --oneline cb30e05b3e..6cfedbcf5e -- internal/hook/branch_guard.go` | `f722ee92c5 …`, `fa276e7b1b [t2013] fix(hook): BranchGuard evaluates primary-checkout and HEAD at the path named by git -C <abs> or a leading cd <abs> && (card t2013)`, `db661c5c0a …` |
| K2 근거 | `grep -rn -A 12 "^func cliGitIn(" <tree>/internal/cli` | `cmd := exec.Command("git", args...)` / `cmd.Dir = dir` / `cmd.Env = gitenv.Env()` / `out, err := cmd.Output()` (quotePath·-z 설정 없음) |
| K3 근거 | `grep -rn "woken_at *=" <tree>/internal/homestate` + `grep -rln "card_waits" <tree>/internal` + `grep -n … hook_stop_goal_wait.go` | 쓰기는 `card_wait.go:166: UPDATE card_waits SET woken_at=? WHERE seq=? AND woken_at=''` 하나(대상 조회 :148 `WHERE cond_kind=? AND cond_ref=? AND woken_at=''`, 인자 `WaitCondPredecessorLanded`) / card_waits 언급 파일: factory.go(:173 DDL), card_wait.go, card_wake_events.go(:86 SELECT), 시험 3개 / `hook_stop_goal_wait.go:80: if w.WokenAt == "" && factoryWaitNextEvent(w.Reason) != "" {` |
| codex 다리 | `mcp__moai__codex_audit(target=baseBranch, mode=native)` | `"verdict":"fail"`, `"audit_receipt":"rcpt-6ea9449ffc3f0730594f9a22"`, `"build_commit":"b1806498fe"`, `"review_base":"main (merge base cb30e05b3e5f64774b8dabd87e80c24332f4551d)"`, findings P1 branch_guard.go:479 / P2 factory_release_route.go:287 / P2 factory_wait.go:171 |

## Baseline-attribution

Go 시험·vet·gofmt는 트리 `6cfedbcf5e1af1326eebb985779f54a7cae95a68`의 소스를 그 자리에서 빌드해 판정했으므로, 판정한 빌드가 곧 트리다. golangci-lint는 설치본 `/Users/goos/go/bin/golangci-lint`가 이 트리 소스를 읽었다(도구 버전은 기록하지 않음). `moai verify audit-plan`과 codex 다리는 설치 빌드 `b1806498fe`(v3.2.0-rc.38)가 실행했고, 이 빌드는 트리 HEAD보다 오래됐다(§2.2). 감사계획은 설정 파일을 읽는 값이고 codex 다리는 정적 리뷰라 결과에 영향은 없다고 판단하지만, 같은 트리로 만든 빌드로 재측정하지는 않았다. K1의 "기준점 대비 새 허용" 판단은 카드 커밋 이력과 3회차 O1 기록(기준점은 훅 cwd만 봄)을 소스 판독으로 대조한 것이다.

## Gaps

- **G1 명령 거절 1건.** `grep -n -e "func substituteQuotedArguments" -e "eval" … branch_guard.go`가 worktree 격리 가드에 거절됐다("this command runs a string through eval, which can't be verified to stay inside the worktree"). 같은 형태로 다시 돌리지 않았다. 대신 함수 이름만 찾는 별도 grep과 Read로 소스를 읽었다. 따라서 branch_guard.go 안에 `eval`을 따로 다루는 코드가 없다는 판단은 grep이 아니라 `explicitGitTarget`·`substituteQuotedArguments` 본문 판독에 근거한다. 이 거절은 `.moai/lessons-inbox.jsonl`에 `tool_failure` 행으로 남았을 수 있다.
- **G2 K1·K2·K3·N3·N4는 실행으로 재현하지 않았다.** 코드 수정 금지 지시 때문에 재현 시험을 쓰지 않았다. 결정적 정규식·SQL 분기라 판독으로 결론이 나지만 재현 기록은 없다.
- **G3 RED 실패 출력 미관측.** RED 커밋(1f0a22acc8, 568269325b)이 시험 파일만 추가했다는 것은 확인했지만, 그 커밋에서 시험이 실제로 실패했는지는 체크아웃 없이 재현하지 않았다.
- **G4 `TestFactoryRelease.*` 선택자는 0건.** 지정 선택자의 이 부분과 일치하는 시험이 없다(출력에 해당 이름 없음). fast check 시험은 `TestT2013FastCheck*` 이름으로 돌았다.
- **G5 커버리지·Windows 빌드 미측정.**
- **G6 audit_multi 미호출.** 지시대로 codex_audit 단일 호출만 했다. audit_multi digest와 card second-review 기록이 없다.
- **G7 2차 범위 한정.** 카드 범위는 두 수리가 만든 회귀만 봤다. codex 새 지적 3건 말고 카드 기존 코드를 처음부터 다시 감사하지 않았다.
- **G8 작업 트리 미커밋 변경.** `.moai/reports/t2013/progress.md` 수정은 내용을 읽지 않았다(보고서 파일).

## Residual-risk

- K1과 같은 계열의 형태(`bash -c`로 감싼 git, 변수 확장 경로, `git -c … -C`)는 기존부터 가드 밖이다. 선행 cd 채택 자체가 정규식 판정이라, 셸이 현재 디렉터리를 바꾸는 다른 형태가 더 있을 수 있다.
- 4회 연속 required codex 다리가 카드 기존 코드에서 새 지적을 냈다. 같은 기존 코드에 아직 나오지 않은 지적이 더 있을 가능성이 높다.
- 판정 도구 빌드(b1806498fe)가 트리보다 오래됐다.

## Iteration history

| 회차 | audited_sha | 판정 | codex receipt |
|---|---|---|---|
| 1 | 2513910f8423dc907b69f6faf3606418af1b29e2 | FAIL 0.68 | rcpt-1adafac80e1966d1648a5c24 (fail) |
| 2 | ff2158a312658cd56329e0c51f1e42272a735d4d | FAIL 0.64 | rcpt-809e05e68eaf48b16036000a (fail) |
| 3 | 6cb0c11f325fdabbb9b80e1bfabf0df92c251ed0 | FAIL 0.76 | rcpt-7ca077fcb8808436b4e45880 (fail) |
| 4 | 6cfedbcf5e1af1326eebb985779f54a7cae95a68 | FAIL 0.75 | rcpt-6ea9449ffc3f0730594f9a22 (fail) |

4회차는 3회차 지적 C1·C3가 해소됐고 F1~F4가 회귀하지 않았음을 확인했다. FAIL이 된 이유는 required codex 다리의 새 지적 가운데 확인된 blocking 2건(K1·K3) 때문이다. 이 회차가 승인된 마지막 감사이므로, K1·K3를 이 카드 안에서 고칠지 후속 카드로 넘길지는 리더가 결정할 일이다.
