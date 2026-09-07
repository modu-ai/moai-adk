# M3 귀속표 — 84건 전수 대조 (AC-CBD-008) · t539

> 종결 조건은 **닫힌 집합** 위에서만 성립한다. 먼저 열거를 고정한다.
>
> ```
> $ /usr/bin/grep -E '^- `internal/[^`]+_test\.go:[0-9]+` `' \
>     .moai/specs/SPEC-CTX-BLIND-DOUBLE-001/plan.md | wc -l
> 84
> ```
>
> (느슨한 패턴 `^- \`internal/` 은 85를 낸다. 초과 1건은 §D 산문 줄
> "- `internal/cli` 는 `-timeout 1800s` 를 준다…" 로, `file:line` 형태가 아니다.
> 위 엄격 패턴이 그 줄을 배제한다.)
>
> 미측정으로 남은 항목: **0건**. 아래 표의 합이 84이고 모든 행이 로그를 갖는다.

## 후보군별 집계

| # | 후보군 | 구성원 | 씨앗 수 | 판정 | 로그 |
|---|---|---|---|---|---|
| 1 | update 체커/업데이터 | 4 | 2 (A 뮤턴트, B seam 부재) | **생존** (3) + 옳게 눈멂 (1) | `m3-update.log` |
| 2 | statusline 프로바이더 | 4 | 1 | **생존** | `m3-statusline.log` |
| 3 | LSP 클라이언트/라우터/트랜스포트 | 46 | 3 | 검출 31 + **생존 12** + 검출 3 | `m3-lsp.log` |
| 4 | template 배포기 | 20 | 2 | **생존** | `m3-template-deployer.log` |
| 5 | gh 클라이언트 | 10 | 3 | **생존** | `m3-gh-client.log` |
| | **합** | **84** | **11** | 생존 49 · 검출 34 · seam 부재 1 | |

검증 (두 방향 모두 84로 닫힌다):

- 후보군별: 4 + 4 + 46 + 20 + 10 = **84**
- 판정별: 생존 **49** + 검출 **34** + seam 부재 **1** = **84**
  - 생존 49 = update 3 · statusline 4 · LSP core 12 · template 20 · gh 10
  - 검출 34 = LSP aggregator 31 · LSP transport 3
  - seam 부재 1 = update 씨앗 B (`fakeUpdateChecker.CheckLatest`)

## 전수 귀속 (84행)

### 후보군 1 — update (4)

| # | 파일:줄 | 대역.메서드 | 씨앗 | 판정 |
|---|---|---|---|---|
| 1 | `internal/cli/init_update_notice_test.go:417` | `fakeUpdateChecker.CheckLatest` | 씨앗 B (seam 부재) | 옳게 눈멂 |
| 2 | `internal/update/orchestrator_test.go:19` | `mockChecker.CheckLatest` | `orchestrator.go:55` | 생존 |
| 3 | `internal/update/orchestrator_test.go:33` | `mockUpdater.Download` | `orchestrator.go:55` | 생존 |
| 4 | `internal/update/orchestrator_test.go:37` | `mockUpdater.Replace` | `orchestrator.go:55` | 생존 |

### 후보군 2 — statusline (4)

| # | 파일:줄 | 대역.메서드 | 씨앗 | 판정 |
|---|---|---|---|---|
| 5 | `internal/statusline/builder_test.go:21` | `mockGitProvider.CollectGitStatus` | `builder.go:362` | 생존 |
| 6 | `internal/statusline/builder_test.go:31` | `mockUpdateProvider.CheckUpdate` | `builder.go:362` | 생존 |
| 7 | `internal/statusline/builder_test.go:543` | `mockUsageProvider.CollectUsage` | `builder.go:362` | 생존 |
| 8 | `internal/statusline/forge_spawn_gate_test.go:151` | `fakeGitProvider.CollectGitStatus` | `builder.go:362` | 생존 |

### 후보군 3 — LSP (46)

씨앗 3a `aggregator.go:174` → **검출**. 구성원 31건 (`internal/lsp/aggregator/aggregator_test.go`):

| # | 줄 | 대역.메서드 | # | 줄 | 대역.메서드 |
|---|---|---|---|---|---|
| 9 | :29 | `fakeClient.Start` | 25 | :201 | `errorClient.Start` |
| 10 | :30 | `fakeClient.Shutdown` | 26 | :202 | `errorClient.Shutdown` |
| 11 | :31 | `fakeClient.OpenFile` | 27 | :203 | `errorClient.OpenFile` |
| 12 | :32 | `fakeClient.DidSave` | 28 | :204 | `errorClient.DidSave` |
| 13 | :33 | `fakeClient.FindReferences` | 29 | :205 | `errorClient.FindReferences` |
| 14 | :36 | `fakeClient.GotoDefinition` | 30 | :208 | `errorClient.GotoDefinition` |
| 15 | :44 | `fakeClient.GetDiagnostics` | 31 | :215 | `errorClient.GetDiagnostics` |
| 16 | :56 | `fakeRouter.RouteFor` | 32 | :224 | `errorRouter.RouteFor` |
| 17 | :157 | `blockingClient.Start` | 33 | :417 | `dispatchRouter.RouteFor` |
| 18 | :158 | `blockingClient.Shutdown` | 34 | :464 | `slowClient.Start` |
| 19 | :159 | `blockingClient.OpenFile` | 35 | :465 | `slowClient.Shutdown` |
| 20 | :160 | `blockingClient.DidSave` | 36 | :466 | `slowClient.OpenFile` |
| 21 | :161 | `blockingClient.FindReferences` | 37 | :467 | `slowClient.DidSave` |
| 22 | :164 | `blockingClient.GotoDefinition` | 38 | :468 | `slowClient.FindReferences` |
| 23 | :171 | `blockingClient.GetDiagnostics` | 39 | :471 | `slowClient.GotoDefinition` |
| 24 | :184 | `blockingRouter.RouteFor` | | | |

씨앗 3b `lsp/core/manager.go:366` → **생존**. 구성원 12건:

| # | 파일:줄 | 대역.메서드 | # | 파일:줄 | 대역.메서드 |
|---|---|---|---|---|---|
| 40 | `core/manager_test.go:32` | `fakeClient.Start` | 46 | `core/manager_test.go:59` | `fakeClient.GotoDefinition` |
| 41 | `core/manager_test.go:43` | `fakeClient.Shutdown` | 47 | `core/document_test.go:37` | `fakeNotifyTransport.Call` |
| 42 | `core/manager_test.go:51` | `fakeClient.OpenFile` | 48 | `core/document_test.go:49` | `fakeNotifyTransport.Notify` |
| 43 | `core/manager_test.go:52` | `fakeClient.DidSave` | 49 | `core/queries_test.go:70` | `fakeQueryTransport.Notify` |
| 44 | `core/manager_test.go:53` | `fakeClient.GetDiagnostics` | 50 | `core/client_stderr_drain_test.go:45` | `nilStderrLauncher.Launch` |
| 45 | `core/manager_test.go:56` | `fakeClient.FindReferences` | 51 | `core/client_stderr_drain_test.go:66` | `largeStderrLauncher.Launch` |

씨앗 3c `lsp/transport/request.go:53` → **검출**. 구성원 3건:

| # | 파일:줄 | 대역.메서드 |
|---|---|---|
| 52 | `transport/error_transport_test.go:28` | `mockRPCConn.Call` |
| 53 | `transport/error_transport_test.go:32` | `mockRPCConn.Notify` |
| 54 | `transport/request_test.go:154` | `blockingTransport.Notify` |

### 후보군 4 — template 배포기 (20)

씨앗 4a `core/project/initializer.go:434` → **생존**. 구성원 8건:

| # | 파일:줄 | 대역.메서드 | # | 파일:줄 | 대역.메서드 |
|---|---|---|---|---|---|
| 55 | `initializer_test.go:26` | `mockDeployer.Deploy` | 59 | `initializer_test.go:780` | `trackingMockDeployer.Deploy` |
| 56 | `initializer_test.go:39` | `mockDeployer.ValidateAll` | 60 | `initializer_test.go:812` | `trackingMockDeployer.ValidateAll` |
| 57 | `initializer_test.go:49` | `capturingDeployer.Deploy` | 61 | `initializer_mirror_notice_test.go:38` | `mirrorResultDeployer.DeployWithResult` |
| 58 | `initializer_test.go:56` | `capturingDeployer.ValidateAll` | 62 | `initializer_mirror_notice_test.go:45` | `mirrorResultDeployer.ValidateAll` |

씨앗 4b `cli/mirror_notice.go:37,:42` → **생존**. 구성원 12건:

| # | 파일:줄 | 대역.메서드 | # | 파일:줄 | 대역.메서드 |
|---|---|---|---|---|---|
| 63 | `cli/update_clean_install_test.go:47` | `stubDeployer.Deploy` | 69 | `cli/update_mirror_notice_test.go:51` | `mirrorReportingDeployer.DeployWithResult` |
| 64 | `cli/update_clean_install_test.go:58` | `stubDeployer.ValidateAll` | 70 | `cli/update_mirror_notice_test.go:58` | `mirrorReportingDeployer.ValidateAll` |
| 65 | `cli/update_clean_install_test.go:1069` | `gitignoreClobberingDeployer.Deploy` | 71 | `cli/update_mirror_notice_test.go:69` | `plainDeployer.Deploy` |
| 66 | `cli/update_clean_install_test.go:1076` | `gitignoreClobberingDeployer.ValidateAll` | 72 | `cli/update_mirror_notice_test.go:75` | `plainDeployer.ValidateAll` |
| 67 | `cli/update_clean_install_config_preserve_test.go:39` | `overwritingDeployer.Deploy` | 73 | `cli/update/merge/merge_test.go:621` | `mockDeployer.Deploy` |
| 68 | `cli/update_clean_install_config_preserve_test.go:73` | `overwritingDeployer.ValidateAll` | 74 | `cli/update/merge/merge_test.go:626` | `mockDeployer.ValidateAll` |

### 후보군 5 — gh 클라이언트 (10)

| # | 파일:줄 | 대역.메서드 | 씨앗 | 판정 |
|---|---|---|---|---|
| 75 | `cli/branch_protection_test.go:26` | `stubGhClient.Run` | `branch_protection.go:65` | 생존 |
| 76 | `cli/branch_protection_test.go:34` | `stubGhClient.RunWithStdin` | `branch_protection.go:65` | 생존 |
| 77 | `github/gh_test.go:31` | `mockGHClient.PRCreate` | `pr_reviewer.go:110` | 생존 |
| 78 | `github/gh_test.go:35` | `mockGHClient.PRView` | `pr_reviewer.go:110` | 생존 |
| 79 | `github/gh_test.go:40` | `mockGHClient.PRMerge` | `pr_reviewer.go:110` | 생존 |
| 80 | `github/gh_test.go:48` | `mockGHClient.PRChecks` | `pr_reviewer.go:110` | 생존 |
| 81 | `github/gh_test.go:53` | `mockGHClient.Push` | `pr_reviewer.go:110` | 생존 |
| 82 | `github/gh_test.go:59` | `mockGHClient.IsAuthenticated` | `pr_reviewer.go:110` | 생존 |
| 83 | `guardstate/axis_test.go:28` | `fakeQuerier.RunsForSubject` | `evaluate.go:180` | 생존 |
| 84 | `guardstate/axis_test.go:36` | `fakeQuerier.AllRuns` | `evaluate.go:180` | 생존 |

## 귀속의 한계 (Gaps — 과장하지 않는다)

**한 씨앗의 판정은 그 씨앗이 실제로 지나간 자리에 대한 실측이고, 같은 후보군의 형제 구성원에
대해서는 귀속(attribution)이지 개별 실측이 아니다.** 84행 전부가 로그를 갖는다는 것과,
84행 전부가 자기 자신의 뮤턴트를 가졌다는 것은 다른 주장이며 이 카드가 세운 것은 전자다.

구체적으로:

- 후보군 2 는 gitProvider seam 1건만 실측했다. updateProvider(`builder.go:373`) 와
  usageProvider(`:388`) 는 같은 함수·같은 `ctx`·같은 모양이지만 **개별 측정하지 않았다.**
- 후보군 5 의 `mockGHClient` 6개 메서드 중 뮤턴트가 실제로 지나간 것은 `PRView` 하나다.
  나머지 5개(`PRCreate` / `PRMerge` / `PRChecks` / `Push` / `IsAuthenticated`)는 같은
  인터페이스·같은 패키지라는 이유로 귀속했을 뿐이다.
- 후보군 3a 의 31건 중 실제로 실행 경로를 탄 것은 `GetDiagnostics` 계열이며,
  `Shutdown` / `DidSave` / `GotoDefinition` 등은 그 판정에 귀속됐다.

이 한계는 plan §F 가 정한 절차("후보마다 실물에서 컨텍스트를 소비하는 지점을 **하나** 고른다")를
그대로 따른 결과이며, 절차 위반이 아니다. 구성원 단위 전수 뮤턴트는 84개의 씨앗을 뜻하고
그것은 Tier M 의 폭이 아니다. 남은 정밀도는 후속 카드 요청에 명시한다.
