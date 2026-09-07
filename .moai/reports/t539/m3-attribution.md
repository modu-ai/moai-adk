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
> **정정 (sync-audit F1·F2, HEAD `73a588054`)**: 최초 작성본은 "미측정 0건" 이라고 적었다.
> 그 문장은 **거짓이었다.** 뮤턴트가 도달하지 못한 자리는 생존도 사망도 아닌 **무판정**이고,
> 무판정에 귀속된 행은 측정된 행이 아니다. 실제 분해:
>
> - **81행** — 도달이 확인된 뮤턴트의 판정을 갖는다 (생존 47 + 검출 34)
> - **3행** — **프로덕션 seam 부재**로 잴 대상 자체가 없다 (1행 + 73·74행)
>
> 81 + 3 = 84. "미측정 0건" 이 아니라 **"뮤턴트 판정 81건 + seam 부재 3건, 판정 공백 0건"** 이
> 정확한 진술이다. 어느 행도 근거 없이 남아 있지 않되, 3행의 근거는 뮤턴트가 아니라
> 인터페이스 시그니처라는 기계적 사실이다.

## 후보군별 집계

| # | 후보군 | 구성원 | 씨앗 수 | 판정 | 로그 |
|---|---|---|---|---|---|
| 1 | update 체커/업데이터 | 4 | 2 (A 뮤턴트, B seam 부재) | **생존** (3) + 옳게 눈멂 (1) | `m3-update.log` |
| 2 | statusline 프로바이더 | 4 | 1 | **생존** | `m3-statusline.log` |
| 3 | LSP 클라이언트/라우터/트랜스포트 | 46 | 3 | 검출 31 + **생존 12** + 검출 3 | `m3-lsp.log` |
| 4 | template 배포기 | 20 | 2 | **생존 18** + seam 부재 2 (F2) | `m3-template-deployer.log` · `m3-merge.log` |
| 5 | gh 클라이언트 | 10 | 3 (5a 무효 → 5a′ 2 seam) | **생존** | `m3-gh-client.log` + `sync-audit-5a-*.log` |
| | **합** | **84** | **12** | 생존 47 · 검출 34 · seam 부재 3 | |

검증 (두 방향 모두 84로 닫힌다):

- 후보군별: 4 + 4 + 46 + 20 + 10 = **84**
- 판정별: 생존 **47** + 검출 **34** + seam 부재 **3** = **84**
  - 생존 47 = update 3 · statusline 4 · LSP core 12 · template 18 · gh 10
  - 검출 34 = LSP aggregator 31 · LSP transport 3
  - seam 부재 3 = update 씨앗 B (`fakeUpdateChecker.CheckLatest`) ·
    merge 73·74행 (`mockDeployer.Deploy` / `.ValidateAll`, F2)

씨앗 수가 11 → 12 로 바뀐 이유: 씨앗 5a(`branch_protection.go:65`)가 **무도달로 무효 처리**되고,
그 자리를 살아 있는 두 seam `:122`·`:158`(5a′)이 대체했다. 무효 씨앗은 계수에서 뺐다.

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

씨앗 3a `aggregator.go:175` → **검출**. 구성원 31건 (`internal/lsp/aggregator/aggregator_test.go`):

> 줄 번호 정정 (sync-audit F3): 최초 작성본은 `:174` 로 인용했으나 174행은 `var getErr error`
> 이고 seam 은 **175행**이다. 인용한 before/after 텍스트가 유일해 재현은 막히지 않았고,
> 감사자가 175행에서 재주입해 두 실패 메시지의 축자 일치를 확인했다.

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
| 67 | `cli/update_clean_install_config_preserve_test.go:39` | `overwritingDeployer.Deploy` | — | (73·74 는 아래로 분리) | — |
| 68 | `cli/update_clean_install_config_preserve_test.go:73` | `overwritingDeployer.ValidateAll` | — | | — |

씨앗 4b 의 도달성은 두 분기 모두 확인됐다(감사자, `sync-audit-reach-4b-{mirror,update}.log`):
`:42` → `REACHPROBE42` (`-run Mirror`), `:37` → `REACHPROBE37` (`-run Update`). 위 12행 중
**10행(63~72)은 건전**하다.

**73·74행 — seam 부재 / 측정 불가** (sync-audit F2, 로그 `m3-merge.log`):

| # | 파일:줄 | 대역.메서드 | 판정 |
|---|---|---|---|
| 73 | `cli/update/merge/merge_test.go:621` | `mockDeployer.Deploy` | **seam 부재 / 측정 불가** |
| 74 | `cli/update/merge/merge_test.go:626` | `mockDeployer.ValidateAll` | **seam 부재 / 측정 불가** |

이 두 행은 씨앗 4b 에 귀속돼 있었으나 그 귀속은 **패키지 경계를 넘지 못한다**:
`deployWithMirrorNotice` 는 package `cli` 의 비공개 함수이고 `merge_test.go` 는 package `merge`
이므로, merge 테스트 바이너리는 그 seam 을 컴파일조차 하지 않는다. 원래 근거로 나열했던
`ok 0.321s (-list 51)` 는 그 뮤턴트에 대해 아무것도 재지 않았다.

package merge 안에서 대체 seam 을 찾았으나 없다 — 유일한 `template.Deployer` 소비 함수
`merge.go:306 AnalyzeMergeChanges(deployer, projectRoot)` 는 **컨텍스트 파라미터가 없고**
`ListTemplates()` 만 부른다:
```
$ /usr/bin/grep -n '\.Deploy(\|\.ValidateAll(' internal/cli/update/merge/merge.go internal/cli/update/merge/base.go
grep_exit=1     ← 프로덕션 호출 0건
$ /usr/bin/grep -n '\.Deploy(\|\.ValidateAll(' internal/cli/update/merge/*_test.go
grep_exit=1     ← 이 패키지 자신의 테스트에서도 0건
```
두 메서드는 `mockDeployer` 가 인터페이스를 만족하도록 시그니처만 채운 스텁이며 어떤 코드도
부르지 않는다. 떨어뜨릴 컨텍스트가 seam 을 건너지 않으므로 **주입할 뮤턴트가 없다** —
"재지 않았다" 가 아니라 **잴 대상이 존재하지 않는다**. 씨앗 1B 와 같은 부류다.

### 후보군 5 — gh 클라이언트 (10)

| # | 파일:줄 | 대역.메서드 | 씨앗 | 판정 |
|---|---|---|---|---|
| 75 | `cli/branch_protection_test.go:26` | `stubGhClient.Run` | `branch_protection.go:122` (`PreflightGh`) | 생존 (도달 증명 후) |
| 76 | `cli/branch_protection_test.go:34` | `stubGhClient.RunWithStdin` | `branch_protection.go:158` (`ApplyBranchProtection`) | 생존 (도달 증명 후) |

> **75·76 의 씨앗 정정 (sync-audit F1 [High])**: 최초 작성본은 두 행을
> `branch_protection.go:65` 에 귀속했다. `:65` 는 `DiscoverOwnerRepo` 내부이고 이 함수는
> **저장소 전체에 호출자가 0건**이다(`/usr/bin/grep -rn 'DiscoverOwnerRepo' --include='*.go' .`
> → 정의 2줄뿐). 도달성 프로브가 그것을 확인했다 — `:65` 에 `panic` 을 넣고
> `-run BranchProtection` 을 돌려도 `ok 0.846s`, panic 미발생.
> 따라서 원래의 `ok 0.862s (-list 5)` 는 **생존이 아니라 무판정**이었고 그 기록은 **무효**다.
>
> 감사자가 `stubGhClient` 가 실제로 쓰이는 두 seam 에서 도달성을 먼저 세운 뒤 재측정했다:
> `:122` → `panic: REACH122` (`TestPreflightGh_Authed`, `sync-audit-5a-reach-preflight.log`),
> `:158` → `panic: REACH158` (`TestApplyBranchProtection_Success`, `sync-audit-5a-reach158.log`).
> 도달 확인 후 뮤턴트: `-run PreflightGh` → `ok 0.887s` (`-list 2`),
> `-run BranchProtection` → `ok 0.688s` (`-list 5`)
> (`sync-audit-5a-live-preflight.log` · `sync-audit-5a-live-apply.log`).
>
> **결론(생존)은 같지만 그것을 세운 증거가 달라졌다** — 원래 증거는 공허했다.
> `DiscoverOwnerRepo` 가 죽은 공개 코드라는 사실은 별개의 관찰로 후속 카드 요청에 부기한다.
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

## 절차 정정 — 도달성이 판정의 전제다 (sync-audit F1 이 강제)

plan §F 의 M3 절차에는 **도달성 확인 단계가 없었다**. 씨앗 5a 가 그 구멍으로 통과해,
실행되지도 않은 뮤턴트의 초록이 "생존" 으로 기록됐다.

[HARD] **생존 판정은 도달성을 먼저 세운 뒤에만 채택한다.**

```
0. 주입할 그 줄에 panic 프로브를 넣고 같은 셀렉터를 돌린다 → panic 발생해야 도달.
   미발생이면 그 씨앗은 아무것도 재지 못하므로 폐기한다.
1. 프로브를 되돌린다.
2. 도달이 선 뒤에야 뮤턴트를 주입한다.
3. 생존/사망을 기록한다.
```

**도달하지 못한 뮤턴트가 내는 `ok` 는 생존이 아니라 무판정이다.** 둘은 출력이 완전히 같아
구별되지 않으며, 구별하는 유일한 관측이 도달성 프로브다. `-list > 0`(AC-CBD-011 공허성 가드)은
"테스트가 존재한다" 만 보장하고 "그 테스트가 이 줄을 지나간다" 는 보장하지 않는다 —
씨앗 5a 는 `-list 5` 로 그 AC 를 문언대로 충족하면서 아무것도 재지 못했다.

감사자는 생존 씨앗 8자리 전수에 이 프로브를 돌렸고, 5a 를 제외한 전부가 도달로 확인됐다
(`sync-audit-reach-*.log`). 후속 SPEC 의 AC 는 이 전제를 문언에 넣어야 한다(후속 카드 요청 F8).
