# 조사 — SPEC-CTX-BLIND-DOUBLE-001

> 정본: `.moai/reports/t539/sweep.md`. 이 문서는 그 측정을 SPEC 산출물로 고정한 것이며,
> 새 사실을 주장하지 않는다. 원본 표는 `sweep-test.tsv` / `sweep-prod.tsv`,
> 뮤턴트 로그는 `m1-codex-mutant.log` / `m2-run-{GLM,Audit,Converg}.log`.
>
> **귀속**: 모든 측정은 `.claude/worktrees/t539`, HEAD `52f863f36`(develop tip), 2026-09-08 세션.
> 뮤턴트는 미커밋 상태로 주입·측정·되돌림했고, 되돌림은 각각
> `git status --short internal/` / `git diff --stat` 무출력으로 확인했다.

## 1. 도구 — `ctxsweep`

`.moai/reports/t539/ctxsweep/main.go`. Go AST 스캐너, 재실행 가능. `-prod` 플래그로 프로덕션 파일 스캔.

판정 기준(도구 헤더 주석과 이 문서가 일치해야 한다):

- `context.Context` 매개변수: **식별자 사용 여부**. 무명(`_` 또는 이름 없음)이면 "안 봄",
  이름은 있으나 본문에서 참조되지 않으면 "안 봄", 참조되면 "봄".
- `*http.Request` 매개변수: **`req.Context()` 호출 여부**. `req.URL` / `req.Body` 만 읽는 것은
  보는 것이 **아니다**. 이 기준 정정으로 `stubGLMDoer.Do` 가 1차 스캔의 오탐("봄")에서
  "안 봄" 으로 바로잡혔다.

### 1.1 `sweep-raw.tsv` 는 폐기됐다 — 정본은 `sweep-test.tsv`

`.moai/reports/t539/sweep-raw.tsv` 는 위 `*http.Request` 기준을 정정하기 **이전**의 1차 스캔이다.
`sweep-test.tsv` 와 **6행**에서 판정이 다르며, 차이는 전부 `yes` → `no` 방향이다:

```
$ diff .moai/reports/t539/sweep-raw.tsv .moai/reports/t539/sweep-test.tsv
cli/audit_pin_live_test.go:108  teeGLMDoer.Do          yes → no
cli/mcp_glm_test.go:34          stubGLMDoer.Do         yes → no
cli/update_version_test.go:251  (funclit)              yes → no
statusline/usage_test.go:435    (funclit)              yes → no
statusline/usage_test.go:993    (funclit)              yes → no
update/checker_test.go:724      (funclit)              yes → no
```

**run-phase 는 `sweep-raw.tsv` 를 근거로 쓰지 않는다.** 이 문서와 `plan.md` 의 모든 열거는
`sweep-test.tsv` 에서 나왔다. (plan-audit D6 은 차이를 5행으로 적었으나 재측정 결과 6행이다 —
`update/checker_test.go:724` 이 빠져 있었다.)

## 2. 스윕 결과 (b 축 — 대역이 보는가)

```
$ go run .moai/reports/t539/ctxsweep/main.go internal > .moai/reports/t539/sweep-test.tsv
method/blank 127   method/no 26   method/yes 29        (합 182)
funclit/blank 196  funclit/no 52  funclit/yes 12       (합 260)
```

컨텍스트를 받는 테스트 대역 메서드 **182건 중 153건(84%)이 컨텍스트를 보지 않는다.**

`*http.Request` 를 받는 대역 메서드는 4건뿐이다:

| 대역 | 위치 | 본다? | 비고 |
|---|---|---|---|
| `stubGLMDoer.Do` | `internal/cli/mcp_glm_test.go:34` | **아니오** | `req.URL` / `req.Body` 만 읽음 |
| `teeGLMDoer.Do` | `internal/cli` | 스캔은 아니오 | **위임 예외** — 요청을 안쪽 실물 doer 에 그대로 넘김 → 충실 |
| `blockingGLMDoer.Do` | `internal/cli` | 예 | |
| `ctxAwareGLMDoer.Do` | `internal/cli/glm_task_bg_context_test.go:48` | 예 | `req.Context().Err()` 를 먼저 확인 |

funclit 260건 중 `*http.Request` 를 받는 43건은 전부 `httptest` **서버 핸들러**(`func(w, r)`)로
클라이언트 대역이 아니다 — 그 테스트들의 클라이언트는 실제 `*http.Client` 라 이미 충실하다. **범위 밖.**

## 3. 프로덕션 스캔 (a 축 — 실물이 보는가)

```
$ go run .moai/reports/t539/ctxsweep/main.go -prod internal > .moai/reports/t539/sweep-prod.tsv
method/blank 17   method/no 42   method/yes 184
```

눈먼 대역 뒤의 실물이 컨텍스트를 소비하는지 — `exec.CommandContext` / `ctx.Done()` /
`NewRequestWithContext` **3토큰의 출현 수**(뒤따르는 `return ctx.Err()` 는 세지 않는다):

**측정 방법 [정정]**: (a) 축 grep 은 **프로덕션 파일로 한정**한다. `_test.go` 를 배제하지 않는
글롭(`request*.go` 류)을 쓰면 테스트 파일의 히트가 프로덕션 근거로 새어 나온다 — 아래 LSP transport
행이 실제로 그렇게 오염됐다(plan-audit D2). 파일을 개별로 지정하거나
`--include='*.go' --exclude='*_test.go'` 를 붙인다.

| 실물 | (a) | 근거 |
|---|---|---|
| `*http.Client` (`glmHTTPDoer` 실물) | **예** | 표준 라이브러리 — 죽은 컨텍스트의 요청에 `context.Canceled` |
| `realCodexSessionRunner.start` | **예** | `internal/cli/mcp_codex.go:433` `exec.CommandContext(ctx…)`, `:448` `readLoop(ctx)` |
| LSP transport `request.go` | **예** | `internal/lsp/transport/request.go:70` `ctx.Err()`, `:53` `t.Call(ctx, method, params, result)` — 컨텍스트를 아래로 넘기고 직접 읽는다(전이적 (a)). **`ctx.Done` 은 이 파일에 0건**이며, 종전에 근거로 적었던 1건은 `request_test.go:154` 의 히트였다 |
| statusline `forge` / `github` / `landed` / `usage` | **예** | 1 / 3 / 1 / 3 |
| gh 클라이언트 (`github/gh.go`, `cli/branch_protection.go`, `guardstate/produce.go`) | **예** | 2 / 2 / 1 |
| update `checker.go` / `updater.go` | **예** | 2 / 1 (`NewRequestWithContext`) |
| template `deployer.go` | **예** | **2** (`:167`, `:354` 의 `case <-ctx.Done():`). 종전의 4 는 뒤따르는 `return ctx.Err()` 2줄까지 센 값이다 — 위 3토큰 규칙으로는 2 |
| hook `registry.go`, loop `controller.go` / `go_feedback.go`, lsp `core/manager.go` | **예** | 2 / 2 / 2 / 3 |
| goal 러너 (`internal/goal/*.go`) | **아니오** | 참조 0건 — 눈먼 대역이 **옳은** 사례 |
| hook `Handle` 18건 (`internal/hook/**`) | **아니오** | 프로덕션 자체가 컨텍스트를 버림 (§7 관찰) |

## 4. 판별식 — 네 축 (SPEC 확정본)

| 축 | 질문 | 판정 방법 | 예 |
|---|---|---|---|
| **(a)** | 실물이 컨텍스트를 보는가 | 프로덕션 스캔 + 표준 라이브러리 지식. **전이적**으로 본다: 실물이 컨텍스트를 `exec.CommandContext` / `http.NewRequestWithContext` / transport 에 넘기면 "본다" | `*http.Client` 예, goal 러너 아니오 |
| **(b)** | 대역이 보는가 | `ctxsweep` 기계 판정. **위임 예외**: 요청을 안쪽 실물에 그대로 넘기는 래퍼는 안쪽이 보면 보는 것 | `stubGLMDoer` 아니오, `blockingGLMDoer` 예, `teeGLMDoer` 예(위임) |
| **(c)** | `(a) ∧ ¬(b)` — 어긋나는가 | (a)(b) 표 결합 | audit 경로 `stubGLMDoer` 참 |
| **(d)** | 대역이 컨텍스트를 보는 **유일한 층**인가 | 테스트 진입점과 결함 지점 사이에 컨텍스트를 보는 프로덕션 층이 있는가. 있으면 눈멂이 아무것도 가리지 않는다 | codex 백그라운드: 있음(`codex_task.go:124`) → 대상 아님 |

**수리 대상 = `(c) ∧ (d)`.**
**채택 증거 = 뮤턴트** — 충실한 대역을 넣은 뒤, 그 충실함이 없었다면 통과했을 결함이 RED 로 바뀌는 것을 보인다.

(d) 는 소스 판독으로 **추정**하되 판정은 뮤턴트로 확정한다. 근거는 §5.1 — 판독으로 세운 가설이 실측에 기각됐다.

## 5. 뮤턴트 실측 두 건

### 5.1 M1 — codex 백그라운드 (t514 수리 역주입) → **검출됨**

주입: `internal/cli/codex_task.go:316`
`go runCodexBackgroundJob(context.WithoutCancel(ctx), …)` → `go runCodexBackgroundJob(ctx, …)`.

가설: 이 경로의 대역 `fakeCodexSession.start` 가 컨텍스트를 무명으로 받으므로 t514 가드는 공허할 것이다.

```
$ go test ./internal/cli/ -run TestCodexTask_BackgroundJobSurvivesRequestContextEnd -count=1 -v
    codex_task_failcause_test.go:144: background job ended "failed" with error
        "codex_task turn was cancelled by the caller, …"
--- FAIL: TestCodexTask_BackgroundJobSurvivesRequestContextEnd (0.01s)
```

**가설 기각.** 잡은 층은 대역이 아니라 `internal/cli/codex_task.go:124` 의 `case <-ctx.Done():` —
턴 루프가 컨텍스트를 직접 본다. 대역이 눈멀어도 **위층이 보면 가려지지 않는다.**
이 측정이 (d) 축을 강제했고, "판독은 틀린다" 를 세웠다.
로그: `.moai/reports/t539/m1-codex-mutant.log`. 되돌림 후 `git status --short internal/` 무출력.

### 5.2 M2 — GLM 감사 요청에서 컨텍스트 제거 → **생존 (가려짐)**

주입: `internal/cli/mcp_glm.go:305`
`http.NewRequestWithContext(ctx, http.MethodPost, url, …)` → `http.NewRequest(http.MethodPost, url, …)`.
컴파일 OK. 의미: MCP 호스트가 요청을 취소해도 z.ai 감사 호출이 끊기지 않는다.

```
$ go test ./internal/cli/ -run GLM     -count=1   → ok  (6.996s)   [-list GLM     = 223 tests]
$ go test ./internal/cli/ -run Audit   -count=1   → ok  (16.604s)  [-list Audit   = 99 tests]
$ go test ./internal/cli/ -run Converg -count=1   → ok  (0.814s)   [-list Converg = 32 tests]
```

`-list` 건수가 셋 다 0보다 크므로 이 초록은 공허하지 않다 — 셀렉터는 실제로 테스트를 골랐고,
그 테스트들이 변이된 경로를 지나고도 통과했다.

> **증거 보존 간극 (plan 단계)**: 223 / 99 / 32 는 본문 숫자로만 존재하고, 그 수를 낸 `-list` 출력을
> 담은 파일은 `.moai/reports/t539/` 에 없다(`m2-run-*.log` 3개는 `ok …` 한 줄뿐이다). 즉 이 기준선은
> plan 증거 집합에서 재유도되지 않는다. run-phase 는 셀렉터마다
> `go test ./internal/cli/ -list <셀렉터> -count=1 > .moai/reports/t539/list-<셀렉터>.log` 로 보존한다
> (REQ-CBD-013 · `AC-CBD-007` · `plan.md` §D).

변이된 경로(`callGLMAudit` / `handleGLMAudit`)를 실제로 지나는 테스트 파일 4개:
`mcp_glm_test.go`, `mcp_glm_audit_pin_test.go`, `mcp_build_identity_test.go`, `audit_pin_live_test.go`.
이들이 쓰는 대역은 `stubGLMDoer`(눈멂)와 `teeGLMDoer`(라이브 전용, 키 없으면 skip)뿐이다.

취소·타임아웃 컨텍스트를 만드는 GLM 테스트 파일은 하나뿐이다 — 분모는 **22** 다:

```
$ ls internal/cli/*glm*_test.go | wc -l
22
$ /usr/bin/grep -ln 'context.WithCancel\|context.WithTimeout\|context.WithDeadline' internal/cli/*glm*_test.go
internal/cli/glm_task_bg_context_test.go
```

그것도 **task 경로**다. (종전에 적었던 분모 "7" 은 어떤 집합을 가리키는지 정의되지 않았고 명령도
기록돼 있지 않았다 — plan-audit D5. 분모를 22 로 넓히면 주장은 오히려 강해진다.)

**두 겹의 관대함이 겹친 자리이며, t532 가 task 경로에서 본 것과 같은 모양이 audit 경로에 남아 있다.**
로그: `.moai/reports/t539/m2-run-{GLM,Audit,Converg}.log`. 되돌림 후 `git diff --stat` 무출력.

## 6. run-phase 후보 (측정 순서 제안 — 판정은 전부 뮤턴트)

순위 2~6 이 M3 의 다섯 후보군이다. **집합은 닫혀 있다** — 각 후보군의 구성원은 `plan.md` §F M3 에
`sweep-test.tsv` 의 `파일:줄` + 타입.메서드 로 전수 열거돼 있고 **합 84건**이다. 아래 표는 후보군의
경계와 구성원 수만 고정하며, 열거 자체는 `plan.md` 가 소유한다. 열거에 없는 대역은 측정 대상이 아니다.

| 순위 | 후보군 / 경로 | 구성원 | (a) | (b) | (d) 추정 | 상태 |
|---|---|---|---|---|---|---|
| 1 | `stubGLMDoer` × GLM audit (`mcp_glm.go:305`) | 1 | 예 | 아니오 | 유일 | **M2 로 가려짐 확정** → 이 카드의 유일한 수리 대상 |
| 2 | update 체커/업데이터 — `mockChecker` / `mockUpdater` / `fakeUpdateChecker` | 4 | 예 | 부분 | 미측정 | 뮤턴트 필요 (M3) |
| 3 | statusline 프로바이더 — `mockGitProvider` / `mockUpdateProvider` / `mockUsageProvider` / `fakeGitProvider` | 4 | 예 | 아니오 | 미측정 | 뮤턴트 필요 (M3) |
| 4 | LSP 클라이언트/라우터/트랜스포트 — `fakeClient` / `blockingClient` / `errorClient` / `slowClient` / `fakeRouter` / `blockingRouter` / `errorRouter` / `dispatchRouter` / `fakeNotifyTransport` / `fakeQueryTransport` / `nilStderrLauncher` / `largeStderrLauncher` / `mockRPCConn` / `blockingTransport` | 46 | 예(transport) | 아니오 | `core/manager.go` 가 봄(3) → 부분 | 뮤턴트 필요 (M3) |
| 5 | template 배포기 — `mockDeployer` / `capturingDeployer` / `trackingMockDeployer` / `mirrorResultDeployer` / `stubDeployer` / `gitignoreClobberingDeployer` / `overwritingDeployer` / `mirrorReportingDeployer` / `plainDeployer` | 20 | 예(2) | 아니오 | 미측정 | 뮤턴트 필요 (M3) |
| 6 | gh 클라이언트 — `stubGhClient` / `mockGHClient` / `fakeQuerier` | 10 | 예 | 아니오 | 미측정 | 뮤턴트 필요 (M3) |
| – | codex `fakeCodexSession` / `hangingCodexSession` / `fixedConnSessionRunner` | – | 예 | 아니오 | **아님**(`codex_task.go:124`) | **M1 로 제외.** 단 백그라운드 이외 codex 경로(`codexReviewRPC` 등)는 미측정이며 이 카드 범위 밖 |
| – | goal `fakeRunner` 등 5종 | – | **아니오** | 아니오 | – | **옳게 눈멂** — 손대지 않음 |
| – | hook 핸들러 대역 (`mockHandler`, `slowSuccessHandler` …) | – | **아니오**(프로덕션이 버림) | 아니오 | – | **옳게 눈멂** — §7 로 이관 |
| – | `httptest` 서버 핸들러 funclit 43건 | – | – | – | – | **대역 아님** — 범위 밖 |
| – | LSP 잔여 대역 — `lsp/core/client_test.go` `fakeTransport` / `fakeLauncher` 3, `lsp/hook/diagnostics_test.go` 2, `lsp/transport/transport_test.go` `fakeTransport` 2 | 7 | 예 | 아니오 | 미측정 | **후보군 4 의 열거에 없음** — 이 카드 범위 밖(`spec.md` §3) |

## 7. 범위 밖 관찰 (보고만)

프로덕션 메서드 **42건**이 컨텍스트를 받고 버린다. 내역은 `sweep-prod.tsv` 의 `method` × `no` 42행을
재계수한 값이며, 합이 정확히 42 다:

| 항목 | 건수 | 근거 |
|---|---|---|
| `internal/hook/**` 의 `Handle` | 18 | `hook/*` prod-`no` 21건 중 `Handle` 만 세는 값 |
| `internal/hook/quality` formatter/linter | 3 | `formatter.go:43 FormatFile`, `linter.go:37 LintFile`, `linter.go:85 AutoFix` |
| `internal/web/**` 핸들러 | 14 | `screens.go` 7 + `profile_crud.go` 4 + `handlers.go` 3 |
| `internal/web/**` 기타 | 2 | `app.go:256`, `glmkey.go:115` |
| `internal/statusline/git.go` `gitCollector.CollectGitStatus` | 1 | `:26` |
| `internal/update/local.go` | 3 | `:63 CheckLatest`, `:180 Download`, `:190 Replace` |
| `internal/update/updater.go` `updaterImpl.Replace` | 1 | `:126` |

**[정정]** 종전 내역("hook 핸들러 21 + web 핸들러 14 + … + `hook/quality` 3")은 `hook/quality` 3건을
`hook 핸들러 21` 안에서 한 번, 별도 항목으로 또 한 번 세어 합이 43 이 됐고, `web` 을 14 로 적어
`app.go` / `glmkey.go` 2건을 누락했다(plan-audit D1). **`hook 핸들러` 항목은 `hook/quality` 를
포함하지 않는다.**

이들에 대해서는 눈먼 대역이 **옳다**. 다만 프로덕션이 컨텍스트를 버리는 것 자체가 별도 결함 부류일 수
있다 — 예컨대 `gitCollector.CollectGitStatus` 가 컨텍스트 없이 git 을 실행한다면 statusline 타임아웃이
그 수집기에는 닿지 않는다. **리드에게 별도 카드 후보로 보고**하며 이 카드에서는 손대지 않는다.

## 8. Gaps

1. (d) 축은 M1·M2 두 경로만 실측했다. §6 의 순위 2~6 은 소스 판독 기반 **추정**이며
   **결함 주장이 아니다** — run-phase 뮤턴트가 판정한다.
2. `ctxsweep` 는 문법 트리만 본다. 대역이 컨텍스트를 구조체 필드에 저장해 두고 별도 고루틴에서
   소비하는 경우는 "본다"로 세지만 충실하다는 뜻은 아니다. 반대로 위임형 래퍼는 "안 본다"로 세지만
   충실하다(§4 위임 예외).
3. `internal/cli` 밖 패키지의 테스트는 M2 셀렉터에 들어 있지 않다. 변이가 `internal/cli` 안에
   갇혀 있으므로 M2 판정에는 충분하다.
4. `teeGLMDoer` 경로(`audit_pin_live_test.go`)는 라이브 키가 없어 skip 됐을 가능성이 크다.
   M2 결과가 이 파일에 의존하지 않는다는 것은 나머지 3개 파일이 같은 경로를 지나는 것으로 충분하다.
5. **`-list` 기준선 223 / 99 / 32 의 출력이 plan 단계 증거 집합에 없다.** 수는 본문에만 있고
   `m2-run-*.log` 3개는 `ok …` 한 줄뿐이라, 이 기준선은 증거 파일에서 재유도되지 않는다.
   run-phase 가 셀렉터별 `-list` 로그를 `.moai/reports/t539/` 에 남겨 이 간극을 닫는다(§5.2 상자).
6. **(a) 축 grep 오염이 이 문서에서 1건 실현됐고, 나머지 행은 전수 재측정되지 않았다.**
   LSP transport 행의 근거가 테스트 파일에서 새어 나왔다(위 §3 정정). 같은 방식으로 만들어진
   다른 행이 우연히 맞았을 수는 있으나, 표에 오르지 못한 실물이 같은 오염을 겪었다면 이 문서는
   그것을 보지 못한다. 방어선은 뮤턴트 판정 하나뿐이다(§9 잔여 위험 2).

## 9. Residual-risk

1. M2 가 가리던 결함은 실운영에서 클라이언트 자체 Timeout(120s)이 상한을 준다 — 무한 매달림이 아니라
   **취소 불응**이다. 심각도는 이 SPEC §1.4 / §4 에서 "가드의 부재" 로 재정의했다.
2. §6 에서 "옳게 눈멂" 으로 분류한 항목은 (a) 판정이 grep 기반이다. 실물이 컨텍스트를 다른 방식
   (`cmd.Cancel`, `time.AfterFunc`)으로 소비하면 오분류다. 오분류된 후보의 뮤턴트는 생존할 것이므로
   M3 절차가 이 위험을 흡수한다.
