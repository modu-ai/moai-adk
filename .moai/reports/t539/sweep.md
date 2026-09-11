# t539 — 컨텍스트를 무시하는 테스트 대역 스윕 (plan-phase 조사 자료)

- **카드**: t539 (t532 파생, Tier M, Class C)
- **브랜치**: `WT-ctx-blind-double` · **트리**: `.claude/worktrees/t539` · **기준 커밋**: `52f863f36` (develop tip, ff)
- **도구**: `.moai/reports/t539/ctxsweep/main.go` (Go AST 스캐너, 재실행 가능)
- **원본 표**: `sweep-test.tsv` (테스트 파일, **정본**), `sweep-prod.tsv` (프로덕션 파일)
- **폐기**: `sweep-raw.tsv` 는 `*http.Request` 판정 기준을 `req.Context()` 호출 여부로 정정하기 **이전**의 1차 스캔이다. `sweep-test.tsv` 와 6행에서 판정이 다르며(전부 `yes` → `no`: `cli/audit_pin_live_test.go:108`, `cli/mcp_glm_test.go:34`, `cli/update_version_test.go:251`, `statusline/usage_test.go:435`, `statusline/usage_test.go:993`, `update/checker_test.go:724`), **근거로 쓰지 않는다.** 정본은 `sweep-test.tsv` 다.
- **뮤턴트 로그**: `m1-codex-mutant.log`, `m2-run-{GLM,Audit,Converg}.log`

이 문서는 SPEC 작성 전에 카드가 [HARD] 로 요구한 "첫 일은 스윕" 을 수행한 기록이다. 수리는 하지 않았다. 프로덕션 파일은 뮤턴트 주입 후 전부 되돌렸다(`git diff --stat` 무출력으로 확인).

---

## 1. Claim

1. 이 저장소의 `*_test.go` 에서 컨텍스트를 **받는** 대역 메서드는 182건이고, 그중 153건(84%)이 컨텍스트를 **보지 않는다**(무명 매개변수 127 + 이름은 있으나 미참조 26).
2. 그러나 "안 본다"는 것만으로는 수리 대상이 아니다. 카드가 준 세 질문 (a)(b)(c) 에 **네 번째 질문 (d)** 를 더해야 판별식이 선다 — M1 실측이 그 근거다.
3. GLM **감사(audit) 경로**에서 눈먼 대역이 컨텍스트 결함을 실제로 가린다는 것을 **뮤턴트로 실측**했다(M2 생존). 이것이 카드 본문의 "구조적으로 가린다" 주장의 첫 번째 실측 사례이며, t532 가 고친 task 경로와는 다른 경로다.
4. codex 백그라운드 경로에서는 대역이 눈멀었음에도 결함이 가려지지 **않았다**(M1 검출). 이유는 대역 위층의 프로덕션 코드가 컨텍스트를 직접 보기 때문이다.

## 2. Evidence

### E1 — 기계적 스윕 (b축: 대역이 보는가)

```
$ go run .moai/reports/t539/ctxsweep/main.go internal > .moai/reports/t539/sweep-test.tsv
method/blank 127   method/no 26   method/yes 29        (합 182)
funclit/blank 196  funclit/no 52  funclit/yes 12       (합 260)
```

판정 기준(도구 헤더 주석과 동일): `context.Context` 는 식별자 사용 여부, `*http.Request` 는 `req.Context()` 호출 여부. `req.URL`/`req.Body` 만 읽는 것은 보는 것이 아니다 — 이 기준으로 `stubGLMDoer.Do` 가 "안 본다"로 바로잡혔다(1차 스캔에서는 `req` 사용 탓에 오탐).

`*http.Request` 를 받는 대역: 메서드 4건 — `stubGLMDoer.Do`(안 봄), `teeGLMDoer.Do`(안 봄 — 단, 아래 §3 위임 예외), `blockingGLMDoer.Do`(봄), `ctxAwareGLMDoer.Do`(봄). funclit 43건은 전부 `httptest` **서버 핸들러**(`func(w, r)`)로, 클라이언트 대역이 아니다 — 그 테스트들의 클라이언트는 실제 `http.Client` 라 충실하다. 범위 밖.

### E2 — 프로덕션 쪽 (a축: 실물이 보는가)

```
$ go run .moai/reports/t539/ctxsweep/main.go -prod internal > .moai/reports/t539/sweep-prod.tsv
method/blank 17   method/no 42   method/yes 184
```

눈먼 대역 뒤의 실물이 컨텍스트를 보는지 (`exec.CommandContext` / `ctx.Done()` / `NewRequestWithContext` 출현 수):

| 실물 | 보는가 | 근거 |
|---|---|---|
| `*http.Client` (glmHTTPDoer 실물) | 예 | 표준 라이브러리 — 죽은 ctx 의 요청에 `context.Canceled` |
| `realCodexSessionRunner.start` | 예 | `mcp_codex.go:433` `exec.CommandContext(ctx…)`, `:448` `readLoop(ctx)` |
| LSP transport `request.go` | 예 | `request.go:53` `t.Call(ctx, …)`, `:70` `ctx.Err()` — [정정] 최초 근거 "`ctx.Done` 1" 은 `request_test.go` 의 히트였다(프로덕션 `request.go` 는 0); (a)축 grep 은 `_test.go` 를 제외해야 한다 (plan-audit D2) |
| statusline `forge/github/landed/usage` | 예 | 1/3/1/3 |
| gh 클라이언트 (`github/gh.go`, `branch_protection.go`, `guardstate/produce.go`) | 예 | 2/2/1 |
| update `checker.go`/`updater.go` | 예 | 2/1 (`NewRequestWithContext`) |
| template `deployer.go` | 예 | 2 — [정정] 최초 "4" 는 `return ctx.Err()` 등을 포함해 센 값; 3토큰 규칙(`exec.CommandContext`/`ctx.Done()`/`NewRequestWithContext`)으로는 2 (plan-audit D9) |
| hook `registry.go`, loop `controller.go`/`go_feedback.go`, lsp `core/manager.go` | 예 | 2/2/2/3 |
| goal runner (`internal/goal/*.go`) | **아니오** | 0 — 눈먼 대역이 옳은 사례 |
| hook 핸들러 18종 (`hook/*.go` `Handle`) + `hook/quality` formatter/linter 3 | **아니오** | 프로덕션 자체가 ctx 를 버림 (§5 관찰) — [정정] 최초 "21종" 은 `hook/quality` 3건을 `Handle` 에 합산한 값 (plan-audit D1) |

### E3 — 뮤턴트 M1: codex 백그라운드 (t514 수리 되돌림) → **검출됨**

주입: `codex_task.go:316` `go runCodexBackgroundJob(context.WithoutCancel(ctx), …)` → `go runCodexBackgroundJob(ctx, …)`.
가설: 이 경로의 대역 `fakeCodexSession.start` 가 ctx 를 안 받으므로(무명 매개변수) t514 가드가 공허할 것이다.

```
$ go test ./internal/cli/ -run TestCodexTask_BackgroundJobSurvivesRequestContextEnd -count=1 -v
    codex_task_failcause_test.go:144: background job ended "failed" with error "codex_task turn was cancelled by the caller, …"
--- FAIL: TestCodexTask_BackgroundJobSurvivesRequestContextEnd (0.01s)
```

**가설 기각.** 잡은 층은 대역이 아니라 `codex_task.go:124` `case <-ctx.Done():` — 턴 루프가 ctx 를 직접 본다. 대역이 눈멀어도, 테스트 진입점과 결함 사이에 ctx 를 보는 **프로덕션 층이 하나라도 있으면** 가려지지 않는다. 되돌림 후 `git status --short internal/` 무출력.

### E4 — 뮤턴트 M2: GLM 감사 요청에서 ctx 제거 → **생존 (가려짐)**

주입: `mcp_glm.go:305` `http.NewRequestWithContext(ctx, …)` → `http.NewRequest(…)`. 컴파일 OK. 의미: MCP 호스트가 요청을 취소해도 z.ai 감사 호출이 끊기지 않는다(클라이언트 자체 Timeout 까지 매달림).

```
$ go test ./internal/cli/ -run GLM     -count=1   → ok  (6.996s)   [-list GLM     = 223 tests]
$ go test ./internal/cli/ -run Audit   -count=1   → ok  (16.604s)  [-list Audit   = 99 tests]
$ go test ./internal/cli/ -run Converg -count=1   → ok  (0.814s)   [-list Converg = 32 tests]
```

변이된 경로(`callGLMAudit`/`handleGLMAudit`)를 실제로 지나는 테스트 파일 4개: `mcp_glm_test.go`, `mcp_glm_audit_pin_test.go`, `mcp_build_identity_test.go`, `audit_pin_live_test.go`. 이들이 쓰는 대역은 `stubGLMDoer`(눈멂)와 `teeGLMDoer`(라이브 전용, 키 없으면 skip)뿐이고, GLM 테스트 파일 7개 중 취소·타임아웃 ctx 를 만드는 파일은 `glm_task_bg_context_test.go` 하나(task 경로)다. **두 겹의 관대함이 겹친 자리이며, 카드가 t532 에서 본 것과 같은 모양이 audit 경로에 남아 있다.** 되돌림 후 `git diff --stat` 무출력.

## 3. 판별식 초안 (SPEC 이 확정할 것)

| 질문 | 판정 방법 | 예 |
|---|---|---|
| (a) 실물이 ctx 를 보는가 | 프로덕션 스캔 + 표준 라이브러리 지식. **전이적**으로 본다: 실물이 ctx 를 `CommandContext`/`NewRequestWithContext`/transport 에 넘기면 "본다" | `*http.Client` 예, goal runner 아니오 |
| (b) 대역이 보는가 | `ctxsweep` 기계 판정. **위임 예외**: `teeGLMDoer` 처럼 요청을 안쪽 실물에 그대로 넘기는 래퍼는 안쪽이 보면 보는 것 | `stubGLMDoer` 아니오, `blockingGLMDoer` 예 |
| (c) (a)∧¬(b) — 어긋나는가 | 표 결합 | audit 경로 `stubGLMDoer` 참 |
| **(d) 대역이 유일한 층인가** | 테스트 진입점 → 결함 지점 사이에 ctx 를 보는 프로덕션 층이 있는가. 있으면 가려지지 않는다 (M1) | codex 백그라운드: 있음(`:124`) → 수리 대상 아님 |

**수리 대상 = (c) ∧ (d)**, 그리고 **채택 증거 = 뮤턴트** — 충실한 대역을 넣은 뒤, 그 충실함이 없었다면 통과했을 결함(M2 류)이 RED 로 바뀌는 것을 보인다. (d) 는 소스 판독으로 추정하되, 판정은 M1/M2 처럼 뮤턴트로 확정한다 — M1 이 보여주듯 판독은 틀린다.

## 4. run-phase 후보 (측정 순서 제안 — 판정은 전부 뮤턴트)

| 순위 | 대역 / 경로 | (a) | (b) | (d) 추정 | 상태 |
|---|---|---|---|---|---|
| 1 | `stubGLMDoer` × GLM audit (`mcp_glm.go:305`) | 예 | 아니오 | 유일 | **M2 로 가려짐 확정** → 수리 대상 |
| 2 | update `mockChecker`/`mockUpdater` (`update/orchestrator_test.go`, ctx 참조는 하나 미소비) + `httptest` 기반 checker 테스트 | 예 | 부분 | 미측정 | 뮤턴트 필요 |
| 3 | statusline `mockGitProvider`/`mockUpdateProvider`/`mockUsageProvider` (`builder_test.go`) — builder 가 collector 별 타임아웃을 거는지에 따라 | 예 | 아니오 | 미측정 | 뮤턴트 필요 |
| 4 | LSP `fakeClient`/`blockingClient`/`errorClient`/`slowClient`/`fakeRouter` (`lsp/aggregator`, `lsp/core/manager_test.go`) | 예(transport) | 아니오 | manager.go 가 ctx 를 봄(3) → 부분 | 뮤턴트 필요 |
| 5 | template `mockDeployer`/`stubDeployer`/`capturingDeployer` 등 8종 | 예(4) | 아니오 | 미측정 | 뮤턴트 필요 |
| 6 | gh `stubGhClient`/`mockGHClient`/`fakeQuerier` | 예 | 아니오 | 미측정 | 뮤턴트 필요 |
| – | codex `fakeCodexSession`/`hangingCodexSession`/`fixedConnSessionRunner` | 예 | 아니오 | **아님**(`:124`) | M1 로 제외 — 단 백그라운드 이외 codex 경로(`codexReviewRPC` 등)는 미측정 |
| – | goal `fakeRunner` 등 5종 | 아니오 | 아니오 | – | 옳은 눈멂 — 손대지 않음 |
| – | hook 핸들러 대역 (`mockHandler`, `slowSuccessHandler` …) | 아니오(프로덕션이 버림) | 아니오 | – | 옳은 눈멂 — §5 로 이관 |
| – | `httptest` 서버 핸들러 funclit 43건 | – | – | – | 대역 아님 — 범위 밖 |

## 5. 범위 밖 관찰 (보고만, 이 카드에서 손대지 않음)

프로덕션 메서드 42건이 `ctx` 를 받고 버린다. 내역(`sweep-prod.tsv` 의 `method` × `no` 42행 재계수, 합 42): hook `Handle` **18**, `hook/quality` formatter/linter 3, web 핸들러 14(`screens.go` 7 + `profile_crud.go` 4 + `handlers.go` 3), web 기타 2(`app.go:256`, `glmkey.go:115`), `statusline/git.go` `gitCollector.CollectGitStatus` 1, `update/local.go` 3, `update/updater.go` `Replace` 1.

> **[정정 2026-09-08]** 이 절의 최초 내역("hook 핸들러 21 + web 핸들러 14 + … + `hook/quality` 3")은 합이 **43** 이었다. 원인 둘: (1) `hook/*` prod-`no` 21건 중 `hook/quality` 3건을 `hook 핸들러 21` 안에서 한 번, 별도 항목으로 또 한 번 세었다 — `Handle` 만 세면 18 이다. (2) `web` prod-`no` 는 실제 **16**건이고, "14" 는 `app.go:256` / `glmkey.go:115` 2건을 누락했다. 위 내역이 재계수 값이다(plan-audit D1).

이들에 대해서는 눈먼 대역이 **옳지만**, 프로덕션이 ctx 를 버리는 것 자체가 별도 결함 부류일 수 있다(예: `gitCollector.CollectGitStatus` 가 ctx 없이 git 을 실행한다면 statusline 타임아웃이 그 수집기에는 닿지 않는다). 리드에게 별도 카드 후보로 올린다.

## 6. Baseline-attribution

모든 측정은 `.claude/worktrees/t539`, HEAD `52f863f36`, 이 세션(2026-09-08)에서 나왔다. M1·M2 는 미커밋 상태로 주입·측정·되돌림했고, 되돌림은 각각 `git status --short internal/` / `git diff --stat` 무출력으로 확인했다.

## 7. Gaps

1. (d) 축은 M1·M2 두 경로만 실측했다. §4 의 2~6 순위는 소스 판독 기반 추정이며 **결함 주장이 아니다** — run-phase 뮤턴트가 판정한다.
2. `ctxsweep` 는 문법 트리만 본다. 대역이 ctx 를 다른 이름으로 받아 넘기는 경우(예: 구조체 필드에 저장 후 별도 고루틴에서 소비)는 "본다"로 세지만 그것이 충실하다는 뜻은 아니다. 반대로 위임형 래퍼는 "안 본다"로 세지만 충실하다(§3 위임 예외).
3. `internal/cli` 밖 패키지의 테스트는 M2 셀렉터에 들어 있지 않다. 변이가 `internal/cli` 안에 갇혀 있으므로 판정에는 충분하다.
4. `teeGLMDoer` 경로(`audit_pin_live_test.go`)는 라이브 키가 없어 skip 됐을 가능성이 크다 — M2 결과가 이 파일에 의존하지 않음은 다른 3개 파일이 같은 경로를 지나는 것으로 충분하다.

## 8. Residual-risk

1. M2 의 "가려진 결함" 은 실운영에서는 클라이언트 Timeout(120s) 이 상한을 준다 — 무한 매달림이 아니라 취소 불응이다. 심각도는 SPEC 에서 정한다.
2. §4 에서 "옳은 눈멂"으로 분류한 항목은 (a) 판정이 grep 기반이다. 실물이 ctx 를 다른 방식(예: `cmd.Cancel`, `time.AfterFunc`)으로 소비하면 오분류다.
