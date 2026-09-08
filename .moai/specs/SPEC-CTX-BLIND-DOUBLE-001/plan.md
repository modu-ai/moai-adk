# 구현 계획 — SPEC-CTX-BLIND-DOUBLE-001

> Tier M · Class C · 카드 t539 · 개발 모드 **tdd**(`quality.yaml` `constitution.development_mode`)
> 워크트리 `.claude/worktrees/t539` · 브랜치 `WT-ctx-blind-double` · 기준 `52f863f36`

## §A 맥락

이 계획은 plan 이전에 이미 끝난 측정 위에 선다. 스윕은 수행됐고(`.moai/reports/t539/sweep.md`),
뮤턴트 두 건이 서로 반대 방향의 실측 사례를 만들었다 — M1 검출(제외 근거), M2 생존(수리 근거).
따라서 run-phase 가 처음부터 새로 재는 것은 **`research.md` §6 순위 2~6 의 다섯 후보군**(구성원 84건,
§F M3 에 전수 열거)뿐이다.

가장 되돌리기 어려운 결정부터 놓는다. 판별식(M1)이 틀리면 M2·M3 의 판정이 전부 흔들리므로 먼저 온다.
M2 는 코드가 바뀌는 자리이자 이 카드의 논지를 증명하는 자리이므로 그다음이다.
M3 는 판별식과 절차가 확정된 뒤 반복 적용하는 기계적 작업이므로 마지막이다.

## §B 알려진 문제

1. `stubGLMDoer.Do`(`internal/cli/mcp_glm_test.go:34`)는 `req.Context()` 를 부르지 않는다.
   `req.URL` / `req.Body` 만 읽으므로 1차 기계 스캔이 "본다"로 오탐했고, 판정 기준을
   "`req.Context()` 호출 여부"로 좁혀 바로잡았다. 같은 오탐이 M3 후보에서도 나올 수 있다.
2. GLM audit 경로에는 취소·타임아웃 컨텍스트를 만드는 픽스처가 없다. 대역의 눈멂과 픽스처의
   `context.Background()` 편중이 겹쳐 M2 가 생존했다.
3. M1 이 보여준 대로 **소스 판독은 (d) 축에서 틀린다.** 계획의 어느 단계도 판독만으로 판정을 닫지 않는다.

## §C 사전 점검 (run-phase 진입 시 1회)

- `git rev-parse --show-toplevel` 이 `.claude/worktrees/t539` 인지 확인한다.
- `git diff --stat` 무출력을 확인한다 — 스윕 단계에서 주입했던 뮤턴트가 남아 있지 않아야 한다.
- `go run .moai/reports/t539/ctxsweep/main.go internal` 을 재실행해 sweep.md §E1 의 집계
  (method 127/26/29, funclit 196/52/12)가 재현되는지 확인한다. 어긋나면 트리가 다른 것이므로
  기준을 다시 잡고 시작한다.

## §D 제약

### 검증 범위

- **로컬에서 `go test ./...` 를 돌리지 않는다.** 건드린 패키지만 돌리고 전 패키지 판정은 CI 에 맡긴다.
- `internal/cli` 는 `-timeout 1800s` 를 준다(다른 카드에서 702.5s 실측). 나머지 패키지는 기본값으로 충분하다.
- 판정 명령은 `go test ./internal/<pkg>/... -run <셀렉터> -count=1` 형태이며,
  같은 셀렉터의 `-list` 건수를 항상 함께 기록한다(REQ-CBD-013).
- **`-list` 출력은 파일로 보존한다.** 건수를 본문에 적는 것만으로는 재유도가 안 된다 —
  plan 단계의 기준선 223 / 99 / 32 는 본문 숫자로만 존재했고 증거 파일이 없었다(plan-audit D8).
  run-phase 는 셀렉터마다
  `go test ./internal/cli/ -list <셀렉터> -count=1 > .moai/reports/t539/list-<셀렉터>.log` 로 남긴다.

### (a) 축 측정 방법

- (a) 축(실물이 컨텍스트를 보는가) grep 은 **프로덕션 파일로 한정**한다. `_test.go` 를 배제하지 않는
  글롭(`request*.go` 류)은 쓰지 않는다 — plan 단계에서 이 글롭이 `request_test.go` 의 히트를
  `request.go` 의 근거로 올렸다(plan-audit D2). 파일을 개별로 지정하거나
  `--include='*.go' --exclude='*_test.go'` 를 붙인다.
- 계수 규칙은 "**3토큰(`exec.CommandContext` / `ctx.Done()` / `NewRequestWithContext`) 출현 수**"로
  고정한다. 뒤따르는 `return ctx.Err()` 는 세지 않는다 — 이 규칙으로 `template/deployer.go` 는 4 가
  아니라 **2** 다(plan-audit D9).

### 뮤턴트 위생

- 뮤턴트는 커밋하지 않는다. 주입 → 측정 → 되돌림을 한 창 안에서 끝내고,
  되돌림은 `git diff --stat` 무출력으로 확인한다(REQ-CBD-014).
- 뮤턴트가 트리에 있는 동안에는 다른 검증을 병행하지 않는다 — 그 창에서 나온 초록은 귀속 불가다.

### 부재 주장

- 셸의 `grep` 은 ugrep 래퍼라 바이너리 추정·gitignore 로 조용히 건너뛴다.
  "이 픽스처가 없다" 류 주장은 전부 `/usr/bin/grep` 으로 측정한다(REQ-CBD-015).

### 추적성

- 모든 커밋 메시지에 `t539` 가 들어간다.
- 증거는 `.moai/reports/t539/verdict.md` 에 모은다. 뮤턴트 로그는 같은 디렉터리에 후보별로 남긴다.

## §E 자기 검증

각 마일스톤 종료 시 다음을 재측정하고 출력을 인용한다.

1. 해당 마일스톤이 건드린 패키지의 `go test` 결과 + 셀렉터별 `-list` 건수
2. 뮤턴트 판정: 주입 전 초록 → 주입 후 상태 → 되돌림 후 초록
3. `git diff --stat` 무출력(뮤턴트 잔재 없음)
4. `go vet ./internal/<pkg>/...`

## §F 마일스톤

### M1 — 판별식 확정 + 스윕 산출물 고정

**결정 밀도 최상.** (a)(b)(c)(d) 정의가 이후 모든 판정의 기준이 되므로 여기서 확정한다.

1. `research.md` 를 SPEC 산출물로 확정한다 — 스윕 표, (a)(b)(c)(d) 판정, 위임 예외,
   "옳게 눈멂" 목록과 그 (a) 근거.
2. `ctxsweep` 스캐너를 `.moai/reports/t539/ctxsweep/` 에 그대로 둔다. `internal/` 승격 없음(REQ-CBD-004).
   판정 기준 주석(컨텍스트 = 식별자 사용 여부, `*http.Request` = `req.Context()` 호출 여부)이
   도구 헤더와 `research.md` 에서 일치하는지 확인한다.
3. 재실행 재현성을 확인한다 — 같은 트리·같은 집계.

산출물: `research.md`(확정), `ctxsweep` 재실행 로그.
코드 변경 없음.

### M2 — GLM audit 경로에 취소 가드를 세운다

**이 카드의 논지를 증명하는 자리.** RED 는 논증이 아니라 뮤턴트로 세운다.

1. **RED-now 수립**: `internal/cli/mcp_glm.go:305` 의 `http.NewRequestWithContext(ctx, …)` 를
   `http.NewRequest(…)` 로 변이하고, 새 가드 테스트를 **아직 넣지 않은 상태**에서
   `-run GLM|Audit|Converg` 가 초록임을 재확인한다(생존 = 현재 가드가 없다는 증거).
   되돌린다.
2. **가드 작성**: audit 경로용 충실한 대역을 넣는다. `ctxAwareGLMDoer`
   (`internal/cli/glm_task_bg_context_test.go:48`)는 같은 `package cli` 에 있으므로
   audit 테스트에서 그대로 참조 가능하다 — 새 타입을 만들 필요가 있는지 여부는 구현 시 판단하되,
   중복 타입을 만들지 않는 쪽을 우선한다.
   가드 테스트는 이미 끝난 컨텍스트로 audit 호출을 구동하고, 호출이 z.ai 요청을 실제로 시도하지
   않거나 컨텍스트 오류로 끝나는 것을 단언한다.
3. **RED 시연**: 가드가 들어간 상태에서 같은 뮤턴트를 다시 주입해 **죽는지** 확인한다.
   되돌린 뒤 초록을 확인한다.
4. **비회귀**: `-run GLM` / `-run Audit` / `-run Converg` 셀렉터가 여전히 초록이고
   `-list` 건수(223 / 99 / 32)가 줄지 않았는지 확인한다. task 경로 가드도 초록이어야 한다.
   세 셀렉터의 `-list` 출력을 `.moai/reports/t539/list-{GLM,Audit,Converg}.log` 로 보존한다
   — plan 단계에는 이 파일이 없었으므로 기준선을 증거 집합에서 재유도할 수 없었다(§D 참조).

산출물: `internal/cli` 테스트 파일 변경. 프로덕션 코드 변경 **없음**
(프로덕션은 이미 옳다 — 없는 것은 가드다).

### M3 — 남은 다섯 후보를 각각 뮤턴트로 판정

**기계적 반복.** 후보마다 같은 절차를 돌린다.

측정 대상(REQ-CBD-009 의 확정 목록). **집합은 닫혀 있다** — 아래 다섯 후보군의 구성원 **84건**이
전부이고, 여기 없는 대역은 이 카드의 측정 대상이 아니다. 각 `파일:줄` 은
`.moai/reports/t539/sweep-test.tsv` 의 `method` 행 중 `consults ≠ yes` 인 것을 그대로 옮긴 값이며,
줄 번호는 **메서드 선언 줄**(타입 선언 줄이 아니다)이다.

| # | 후보군 | 구성원 | 실물 (a) 근거 |
|---|---|---|---|
| 1 | update 체커/업데이터 | 4 | `update/checker.go` `NewRequestWithContext` 2, `update/updater.go` 1 |
| 2 | statusline 프로바이더 | 4 | statusline `forge` / `github` / `landed` / `usage` 1 / 3 / 1 / 3 |
| 3 | LSP 클라이언트/라우터/트랜스포트 | 46 | `lsp/transport/request.go:70` `ctx.Err()` + `:53` `t.Call(ctx, …)`, `lsp/core/manager.go` 3 |
| 4 | template 배포기 | 20 | `template/deployer.go` 2 (`:167`, `:354` 의 `case <-ctx.Done():`) |
| 5 | gh 클라이언트 | 10 | `github/gh.go` 2, `cli/branch_protection.go` 2, `guardstate/produce.go` 1 |
| | **합** | **84** | |

#### 후보군 1 — update 체커/업데이터 (4)

- `internal/cli/init_update_notice_test.go:417` `fakeUpdateChecker.CheckLatest`
- `internal/update/orchestrator_test.go:19` `mockChecker.CheckLatest`
- `internal/update/orchestrator_test.go:33` `mockUpdater.Download`
- `internal/update/orchestrator_test.go:37` `mockUpdater.Replace`

#### 후보군 2 — statusline 프로바이더 (4)

- `internal/statusline/builder_test.go:21` `mockGitProvider.CollectGitStatus`
- `internal/statusline/builder_test.go:31` `mockUpdateProvider.CheckUpdate`
- `internal/statusline/builder_test.go:543` `mockUsageProvider.CollectUsage`
- `internal/statusline/forge_spawn_gate_test.go:151` `fakeGitProvider.CollectGitStatus`

#### 후보군 3 — LSP 클라이언트/라우터/트랜스포트 (46)

`internal/lsp/aggregator/aggregator_test.go` (31) — 아래 줄 번호는 전부 이 파일의 것:

- `internal/lsp/aggregator/aggregator_test.go:29` `fakeClient.Start`
- `internal/lsp/aggregator/aggregator_test.go:30` `fakeClient.Shutdown`
- `internal/lsp/aggregator/aggregator_test.go:31` `fakeClient.OpenFile`
- `internal/lsp/aggregator/aggregator_test.go:32` `fakeClient.DidSave`
- `internal/lsp/aggregator/aggregator_test.go:33` `fakeClient.FindReferences`
- `internal/lsp/aggregator/aggregator_test.go:36` `fakeClient.GotoDefinition`
- `internal/lsp/aggregator/aggregator_test.go:44` `fakeClient.GetDiagnostics`
- `internal/lsp/aggregator/aggregator_test.go:56` `fakeRouter.RouteFor`
- `internal/lsp/aggregator/aggregator_test.go:157` `blockingClient.Start`
- `internal/lsp/aggregator/aggregator_test.go:158` `blockingClient.Shutdown`
- `internal/lsp/aggregator/aggregator_test.go:159` `blockingClient.OpenFile`
- `internal/lsp/aggregator/aggregator_test.go:160` `blockingClient.DidSave`
- `internal/lsp/aggregator/aggregator_test.go:161` `blockingClient.FindReferences`
- `internal/lsp/aggregator/aggregator_test.go:164` `blockingClient.GotoDefinition`
- `internal/lsp/aggregator/aggregator_test.go:171` `blockingClient.GetDiagnostics`
- `internal/lsp/aggregator/aggregator_test.go:184` `blockingRouter.RouteFor`
- `internal/lsp/aggregator/aggregator_test.go:201` `errorClient.Start`
- `internal/lsp/aggregator/aggregator_test.go:202` `errorClient.Shutdown`
- `internal/lsp/aggregator/aggregator_test.go:203` `errorClient.OpenFile`
- `internal/lsp/aggregator/aggregator_test.go:204` `errorClient.DidSave`
- `internal/lsp/aggregator/aggregator_test.go:205` `errorClient.FindReferences`
- `internal/lsp/aggregator/aggregator_test.go:208` `errorClient.GotoDefinition`
- `internal/lsp/aggregator/aggregator_test.go:215` `errorClient.GetDiagnostics`
- `internal/lsp/aggregator/aggregator_test.go:224` `errorRouter.RouteFor`
- `internal/lsp/aggregator/aggregator_test.go:417` `dispatchRouter.RouteFor`
- `internal/lsp/aggregator/aggregator_test.go:464` `slowClient.Start`
- `internal/lsp/aggregator/aggregator_test.go:465` `slowClient.Shutdown`
- `internal/lsp/aggregator/aggregator_test.go:466` `slowClient.OpenFile`
- `internal/lsp/aggregator/aggregator_test.go:467` `slowClient.DidSave`
- `internal/lsp/aggregator/aggregator_test.go:468` `slowClient.FindReferences`
- `internal/lsp/aggregator/aggregator_test.go:471` `slowClient.GotoDefinition`

`internal/lsp/core` (12):

- `internal/lsp/core/manager_test.go:32` `fakeClient.Start`
- `internal/lsp/core/manager_test.go:43` `fakeClient.Shutdown`
- `internal/lsp/core/manager_test.go:51` `fakeClient.OpenFile`
- `internal/lsp/core/manager_test.go:52` `fakeClient.DidSave`
- `internal/lsp/core/manager_test.go:53` `fakeClient.GetDiagnostics`
- `internal/lsp/core/manager_test.go:56` `fakeClient.FindReferences`
- `internal/lsp/core/manager_test.go:59` `fakeClient.GotoDefinition`
- `internal/lsp/core/document_test.go:37` `fakeNotifyTransport.Call`
- `internal/lsp/core/document_test.go:49` `fakeNotifyTransport.Notify`
- `internal/lsp/core/queries_test.go:70` `fakeQueryTransport.Notify`
- `internal/lsp/core/client_stderr_drain_test.go:45` `nilStderrLauncher.Launch`
- `internal/lsp/core/client_stderr_drain_test.go:66` `largeStderrLauncher.Launch`

`internal/lsp/transport` (3):

- `internal/lsp/transport/error_transport_test.go:28` `mockRPCConn.Call`
- `internal/lsp/transport/error_transport_test.go:32` `mockRPCConn.Notify`
- `internal/lsp/transport/request_test.go:154` `blockingTransport.Notify`

#### 후보군 4 — template 배포기 (20)

- `internal/core/project/initializer_test.go:26` `mockDeployer.Deploy`
- `internal/core/project/initializer_test.go:39` `mockDeployer.ValidateAll`
- `internal/core/project/initializer_test.go:49` `capturingDeployer.Deploy`
- `internal/core/project/initializer_test.go:56` `capturingDeployer.ValidateAll`
- `internal/core/project/initializer_test.go:780` `trackingMockDeployer.Deploy`
- `internal/core/project/initializer_test.go:812` `trackingMockDeployer.ValidateAll`
- `internal/core/project/initializer_mirror_notice_test.go:38` `mirrorResultDeployer.DeployWithResult`
- `internal/core/project/initializer_mirror_notice_test.go:45` `mirrorResultDeployer.ValidateAll`
- `internal/cli/update/merge/merge_test.go:621` `mockDeployer.Deploy`
- `internal/cli/update/merge/merge_test.go:626` `mockDeployer.ValidateAll`
- `internal/cli/update_clean_install_test.go:47` `stubDeployer.Deploy`
- `internal/cli/update_clean_install_test.go:58` `stubDeployer.ValidateAll`
- `internal/cli/update_clean_install_test.go:1069` `gitignoreClobberingDeployer.Deploy`
- `internal/cli/update_clean_install_test.go:1076` `gitignoreClobberingDeployer.ValidateAll`
- `internal/cli/update_clean_install_config_preserve_test.go:39` `overwritingDeployer.Deploy`
- `internal/cli/update_clean_install_config_preserve_test.go:73` `overwritingDeployer.ValidateAll`
- `internal/cli/update_mirror_notice_test.go:51` `mirrorReportingDeployer.DeployWithResult`
- `internal/cli/update_mirror_notice_test.go:58` `mirrorReportingDeployer.ValidateAll`
- `internal/cli/update_mirror_notice_test.go:69` `plainDeployer.Deploy`
- `internal/cli/update_mirror_notice_test.go:75` `plainDeployer.ValidateAll`

#### 후보군 5 — gh 클라이언트 (10)

- `internal/cli/branch_protection_test.go:26` `stubGhClient.Run`
- `internal/cli/branch_protection_test.go:34` `stubGhClient.RunWithStdin`
- `internal/github/gh_test.go:31` `mockGHClient.PRCreate`
- `internal/github/gh_test.go:35` `mockGHClient.PRView`
- `internal/github/gh_test.go:40` `mockGHClient.PRMerge`
- `internal/github/gh_test.go:48` `mockGHClient.PRChecks`
- `internal/github/gh_test.go:53` `mockGHClient.Push`
- `internal/github/gh_test.go:59` `mockGHClient.IsAuthenticated`
- `internal/guardstate/axis_test.go:28` `fakeQuerier.RunsForSubject`
- `internal/guardstate/axis_test.go:36` `fakeQuerier.AllRuns`

후보별 절차:

1. 실물에서 컨텍스트를 소비하는 지점을 하나 고른다.
2. 그 지점의 컨텍스트를 떨어뜨리는 뮤턴트를 주입한다(컴파일되는 최소 변이).
3. 해당 패키지 테스트를 셀렉터와 `-list` 건수를 기록하며 돌린다.
4. **검출됨** → 그 후보는 수리하지 않고 CLOSE 한다. 로그를 남기고 "옳게 눈멂 / 위층이 본다" 로 기록(REQ-CBD-010).
5. **생존함** → 로그를 남기고 해당 패키지의 **후속 카드 요청**을 리드에게 보고한다. **수리하지 않는다**(REQ-CBD-011).
6. 되돌림을 `git diff --stat` 무출력으로 확인한다.

#### 수리의 소속 — 해소됨 (운영자 결정, 2026-09-08)

plan-audit 1회차가 must-pass 로 잡은 미해소 질문("생존 뮤턴트가 여러 후보군에서 나오면 수리를
전부 이 카드에서 마감할 것인가")에 대한 운영자의 답: **측정만 이 카드, 수리는 후속 카드.**

- 이 카드는 다섯 후보군 **84건 전부**를 뮤턴트로 측정한다. 측정은 어느 쪽으로도 갈리지 않는다.
- 이 카드가 **수리하는 자리는 M2 의 GLM audit 경로 하나뿐**이다 — 이미 생존이 실측된 자리다.
- 그 밖의 생존 뮤턴트는 로그를 `.moai/reports/t539/` 에 남기고 **패키지별 후속 카드 요청**으로
  리드에게 보고한다. 이 카드에서 대역을 고치지 않는다.
- 검출된 뮤턴트는 "옳게 눈멂 / 위층이 컨텍스트를 본다" 로 로그와 함께 CLOSE 한다.

근거: 후보가 `internal/update` / `internal/statusline` / `internal/lsp/*` / `internal/core/project` /
`internal/cli/update/merge` / `internal/guardstate` 7개 패키지에 흩어져 있어, 전부 수리하면
검증 범위와 diff 가 Tier M 의 폭(5~15 파일)을 넘긴다. 측정과 수리를 갈라 두면 이 카드는
Tier M 안에 머물고, 후속 카드는 패키지 단위로 자기 검증 범위를 갖는다.

## §G 안티패턴

- **커버리지를 채택 증거로 쓰기.** 대역을 충실하게 만들면 커버리지가 오를 수 있으나, 그것은
  "대역이 무언가를 잡았다"와 다른 사실이다. 채택은 뮤턴트만이 판정한다.
- **일괄 교체.** "눈먼 대역 153건" 을 수리 대상 수로 읽는 것. 84% 중 대다수는 눈먼 것이 옳다.
- **판독으로 (d) 를 닫기.** M1 에서 이미 한 번 틀렸다. 추정은 순서를 정하는 데만 쓰고 판정에는 쓰지 않는다.
- **`-run` 셀렉터의 초록을 세어 보지 않고 인용하기.** 0건 셀렉터의 `ok` 는 아무것도 주장하지 않는다.
- **경계가 열린 목록 위에서 "0건 남았다" 고 주장하기.** `외 …` / `(+ …)` 같은 꼬리를 단 후보 목록은
  원소가 확정되지 않아 여집합이 비었다는 주장이 어떤 관측으로도 거짓이 될 수 없다 —
  그것도 공허한 초록이다. §F M3 의 84건 열거가 이 카드에서 그 꼬리를 잘라낸 조치다.
- **뮤턴트가 주입된 창에서 다른 검증 돌리기.** 그 창의 초록은 귀속 불가다.
- **부재 주장을 셸 `grep` 으로 세우기.** ugrep 래퍼가 조용히 건너뛴다.

## §H 상호 참조

- 조사 정본: `.moai/reports/t539/sweep.md`
- 원본 카드 t532 의 수리: `internal/cli/glm_task_bg_context_test.go`
- M1 역주입 대상 t514: `internal/cli/codex_task.go:316`
- 형제 교훈: `feedback_confident_verdicts_over_empty_sets.md`,
  `feedback_an_assertion_narrower_than_the_mechanism_passes_while_it_fires.md`,
  `feedback_a_guard_anchored_below_the_failing_layer_does_not_guard_it.md`
