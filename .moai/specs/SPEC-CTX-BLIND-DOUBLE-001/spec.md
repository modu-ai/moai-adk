---
id: SPEC-CTX-BLIND-DOUBLE-001
title: "컨텍스트를 무시하는 테스트 대역이 컨텍스트 결함을 가린다 — 판별식을 세우고 뮤턴트로 판정한다"
version: "0.2.0"
status: draft
created: 2026-09-08
updated: 2026-09-08
author: GOOS
priority: P1
phase: "v3.1.4 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "test-double, context-cancellation, mutation-testing, glm-audit, faithful-double, ctxsweep"
tier: M
era: V3R6
---

# SPEC: 컨텍스트를 무시하는 테스트 대역이 컨텍스트 결함을 가린다

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-08 | GOOS | 최초 작성 — 카드 t539(t532 파생). plan 이전에 수행한 기계적 스윕(`.moai/reports/t539/sweep.md`, HEAD `52f863f36`)을 근거로 4축 판별식을 확정하고, M2 뮤턴트 생존(GLM audit 경로)·M1 뮤턴트 검출(codex 백그라운드 경로)을 각각 수리 대상·제외 대상의 실측 사례로 고정 |
| 0.2.0 | 2026-09-08 | GOOS | plan-audit 1회차(FAIL 0.84) 지적 반영 — D0 해소(운영자 결정: 측정만 이 카드, 수리는 후속 카드), M3 후보군을 다섯으로 닫고 구성원 84건을 `plan.md` §F 에 `파일:줄` 로 전수 열거(D3), 범위 밖 42건 내역 재계수(D1), LSP transport (a)축 근거 교체 + 프로덕션 한정 재측정(D2), M1 인수 기준 3건을 regression-guard 로 재분류(D4), GLM 테스트 파일 분모 7→22 정정(D5), `sweep-raw.tsv` 폐기 표기(D6), modality 라벨 정정(D7), `-list` 기준선 보존 절차 추가(D8), `deployer.go` (a) 계수 4→2 정정(D9), 추적성 요구 `REQ-CBD-016` 신설(D10), `REQ-CBD-009` 절 참조 §4→§6 정정(D12) |

## 1. 문제 — 측정된 형태

테스트 대역이 자기가 받은 컨텍스트를 보지 않으면, 그 대역은 자기가 흉내 내는 실물보다 **관대해진다**.
실물이 죽은 컨텍스트의 요청을 거절하는데 대역은 받아준다면, 그 자리를 지나는 테스트는
"컨텍스트를 잘못 다루는 코드"를 통과시킨다. 결함이 없어서 초록인지, 대역이 눈멀어서 초록인지
출력만으로는 구분되지 않는다.

### 1.1 기원 — t532 의 1차 재현이 초록이었던 이유

`stubGLMDoer.Do`(`internal/cli/mcp_glm_test.go:34`)는 `req.URL` 과 `req.Body` 만 읽고
`req.Context()` 를 한 번도 부르지 않는다. 실물인 `*http.Client` 는 죽은 컨텍스트의 요청에
`context.Canceled` 를 돌려주므로, 이 대역은 실물보다 관대하다.
t532 의 첫 재현이 초록이었던 것은 결함이 없어서가 아니라 이 대역이 눈멀어서였고,
컨텍스트를 실제로 보는 대역 `ctxAwareGLMDoer`(`internal/cli/glm_task_bg_context_test.go:48`,
`req.Context().Err()` 를 먼저 확인)를 넣고 나서야 RED 가 됐다.

두 겹의 관대함이 겹쳤다 — **눈먼 대역**, 그리고 **항상 `context.Background()` 를 쓰는 픽스처**.
어느 한쪽만 있어도 결함은 드러나지 않는다.

### 1.2 기계적 스윕 — 얼마나 넓은가

`.moai/reports/t539/ctxsweep/main.go`(Go AST 스캐너, 재실행 가능)로 `internal/` 전체를 훑었다.
컨텍스트를 **받는** 테스트 대역 메서드 182건 중 153건(84%)이 컨텍스트를 **보지 않는다**
(무명 매개변수 127 + 이름은 있으나 미참조 26). 원본 표는 `sweep-test.tsv` / `sweep-prod.tsv` 에 있다.

그러나 **84% 는 수리 대상 수가 아니다.** 어떤 대역은 눈먼 것이 옳다 — 그 대역이 흉내 내는
실물 자체가 컨텍스트를 보지 않기 때문이다(예: `internal/goal/*.go` 의 러너는 컨텍스트 참조 0건).
눈멂 자체는 결함이 아니고, **실물과 어긋나는 눈멂**만이 결함이다.

### 1.3 어긋남조차 항상 가리지는 않는다 — M1 이 기각한 가설

뮤턴트 M1: `internal/cli/codex_task.go:316` 의 `go runCodexBackgroundJob(context.WithoutCancel(ctx), …)` 를
`go runCodexBackgroundJob(ctx, …)` 로 되돌렸다(t514 수리의 역주입).
가설은 "이 경로의 대역 `fakeCodexSession.start` 가 컨텍스트를 안 받으니 t514 가드는 공허하다" 였다.

```
$ go test ./internal/cli/ -run TestCodexTask_BackgroundJobSurvivesRequestContextEnd -count=1 -v
--- FAIL: TestCodexTask_BackgroundJobSurvivesRequestContextEnd (0.01s)
```

**가설 기각.** 잡은 층은 대역이 아니라 `internal/cli/codex_task.go:124` 의 `case <-ctx.Done():` —
턴 루프가 컨텍스트를 직접 본다. 테스트 진입점과 결함 지점 사이에 컨텍스트를 보는
**프로덕션 층이 하나라도 있으면**, 그 아래 대역이 눈멀어도 결함은 가려지지 않는다.

이 측정이 판별식에 네 번째 축을 강제한다. 소스 판독으로 세운 가설이 실측으로 뒤집혔다는 사실
자체가, 판정을 판독이 아니라 뮤턴트에 맡겨야 하는 이유다.

### 1.4 실제로 가려지는 자리 — M2 생존

뮤턴트 M2: `internal/cli/mcp_glm.go:305` 의 `http.NewRequestWithContext(ctx, …)` 를
`http.NewRequest(…)` 로 바꿨다. 의미는 "MCP 호스트가 요청을 취소해도 z.ai 감사 호출이 끊기지 않는다".

```
$ go test ./internal/cli/ -run GLM     -count=1   → ok  (6.996s)   [-list GLM     = 223 tests]
$ go test ./internal/cli/ -run Audit   -count=1   → ok  (16.604s)  [-list Audit   = 99 tests]
$ go test ./internal/cli/ -run Converg -count=1   → ok  (0.814s)   [-list Converg = 32 tests]
```

**생존.** 변이된 경로(`callGLMAudit` / `handleGLMAudit`)를 지나는 테스트 파일 4개
(`mcp_glm_test.go`, `mcp_glm_audit_pin_test.go`, `mcp_build_identity_test.go`, `audit_pin_live_test.go`)가
쓰는 대역은 `stubGLMDoer`(눈멂)와 `teeGLMDoer`(라이브 전용, 키 없으면 skip)뿐이고,
GLM 테스트 파일 **22개**(`ls internal/cli/*glm*_test.go | wc -l` → 22) 중 취소·타임아웃 컨텍스트를
만드는 파일은 `glm_task_bg_context_test.go` 하나 — 그것도 **task 경로**다. audit 경로에는 그런 픽스처가 없다.
근거 명령: `/usr/bin/grep -ln 'context.WithCancel\|context.WithTimeout\|context.WithDeadline' internal/cli/*glm*_test.go`
→ 출력 1행(`internal/cli/glm_task_bg_context_test.go`). 분모를 22 로 넓혀도 결론은 바뀌지 않는다.

t532 가 task 경로에서 고친 것과 **같은 모양이 audit 경로에 그대로 남아 있다.**
프로덕션은 지금 옳지만(`NewRequestWithContext` 가 있다), 그것이 옳다는 사실을 지키는 가드가 없다.

## 2. 요구사항 (GEARS)

### M1 — 판별식과 스윕

- **REQ-CBD-001** (Ubiquitous): run-phase 의 첫 산출물은 수리가 아니라 **스윕**이다.
  이 저장소에서 컨텍스트를 받는 테스트 대역을 열거하고, 각각에 대해 (a) 실물이 컨텍스트를 보는가,
  (b) 대역이 보는가, (c) 둘이 어긋나는가를 판정해 `research.md` 에 표로 기록한다.
- **REQ-CBD-002** (Ubiquitous): 판별식은 **네 축**이다 — (a) 실물이 컨텍스트를 보는가(전이적:
  실물이 컨텍스트를 `exec.CommandContext` / `http.NewRequestWithContext` / transport 에 넘기면 본다),
  (b) 대역이 보는가, (c) `(a) ∧ ¬(b)`, (d) 테스트 진입점과 결함 지점 사이에서 **대역이 컨텍스트를 보는 유일한 층인가**.
  수리 대상은 `(c) ∧ (d)` 이며, 이 정의는 `spec.md` 와 `research.md` 양쪽에 기록된다.
- **REQ-CBD-003** (unwanted): 눈먼 대역을 일괄로 교체하지 **않는다**. (a) 가 아니오인 대역
  — 흉내 내는 실물 자체가 컨텍스트를 보지 않는 대역 — 은 눈먼 것이 옳으므로 손대지 않는다.
- **REQ-CBD-004** (Ubiquitous): `ctxsweep` 스캐너는 `.moai/reports/t539/ctxsweep/` 에 증거 도구로 남으며
  `internal/` 이하 제품 코드로 승격하지 **않는다**. 스캐너는 재실행 가능해야 하고, 같은 트리에서
  같은 집계를 낸다.
- **REQ-CBD-005** (Where — 위임 예외): Where 대역이 받은 요청을 안쪽 실물에 그대로 넘기는 래퍼인 경우,
  기계 스캔이 "안 본다"로 세더라도 안쪽 실물이 보면 그 대역은 **충실**한 것으로 판정한다
  (`teeGLMDoer.Do` 가 이 예외에 해당한다).

### M2 — GLM audit 경로

- **REQ-CBD-006** (event-driven): When 컨텍스트가 이미 끝난 요청이 GLM audit 경로의 테스트 대역에 도달하면,
  그 대역은 `req.Context().Err()` 를 돌려주고 성공 응답을 내지 **않는다**.
- **REQ-CBD-007** (Ubiquitous): GLM audit 경로에는 취소 가드 테스트가 존재하며, 그 테스트의 RED 는
  뮤턴트 M2(`internal/cli/mcp_glm.go:305` 의 `http.NewRequestWithContext` → `http.NewRequest`)로 시연된다
  — 수리 전에는 생존하고 수리 후에는 죽는다.
- **REQ-CBD-008** (Ubiquitous): task 경로의 기존 가드
  (`internal/cli/glm_task_bg_context_test.go`)는 M2 수리 후에도 초록을 유지하며,
  `stubGLMDoer` 를 쓰는 나머지 GLM 테스트의 의미는 바뀌지 않는다.

### M3 — 남은 후보 측정

- **REQ-CBD-009** (Ubiquitous): M3 의 측정 대상은 `research.md` **§6** 순위 2~6 의
  **다섯 후보군으로 한정**된다 — (1) update 체커/업데이터 대역, (2) statusline 프로바이더 대역,
  (3) LSP 클라이언트/라우터/트랜스포트 대역, (4) template 배포기 대역, (5) gh 클라이언트 대역.
  다섯 후보군의 구성원은 `plan.md` §F M3 에 `sweep-test.tsv` 의 `파일:줄` + 타입.메서드 로
  **전수 열거되며 총 84건**이다. 열거에 없는 대역은 이 카드의 측정 대상이 아니다 — 집합이
  닫혀 있어야 "미측정 0건"(`AC-CBD-008`)이 반증 가능해진다.
- **REQ-CBD-010** (event-driven): When 후보의 뮤턴트가 **검출되면**(생존하지 않으면), 그 후보는
  수리하지 않고 "옳게 눈멂 / 위층이 컨텍스트를 본다" 로 CLOSE 하며, 검출 로그를 증거로 남긴다.
- **REQ-CBD-011** (event-driven): When 후보의 뮤턴트가 **생존하면**, 이 카드는 그 사실을 로그로 남기고
  해당 패키지의 **후속 카드 요청**을 리드에게 보고하며, 대역 수리는 하지 **않는다**.
  이 카드가 수리하는 자리는 M2 의 GLM audit 경로 하나뿐이다
  (운영자 결정 2026-09-08 — "측정만 이 카드, 수리는 후속 카드"; `plan.md` §F 참조).

### 전역 검증 규율

- **REQ-CBD-012** (Ubiquitous): 수리의 채택 증거는 **뮤턴트**다. 커버리지 상승은 대역이 무언가를
  잡았다는 사실과 다른 사실이므로 채택 증거로 쓰지 않는다.
- **REQ-CBD-013** (Ubiquitous): 모든 `go test -run <셀렉터>` 판정은 같은 셀렉터의 `-list` 건수를 함께 기록한다.
  0건 셀렉터의 초록은 공허하므로 판정으로 쓰지 않는다.
- **REQ-CBD-014** (state-driven): While 뮤턴트가 작업 트리에 주입돼 있으면, 그 창 안에서
  커밋·푸시하지 않으며, 되돌림은 `git diff --stat` 무출력으로 확인한 뒤 다음 단계로 넘어간다.
- **REQ-CBD-015** (Ubiquitous): 부재 주장(어떤 토큰·픽스처가 없다는 주장)은 `/usr/bin/grep` 으로 측정한다.
  셸의 `grep` 은 ugrep 래퍼라 조용히 건너뛰므로 부재 근거로 쓰지 않는다.
- **REQ-CBD-016** (Ubiquitous): 이 카드의 산출물은 추적 가능하다 — 모든 커밋 제목에 카드 id `t539` 가
  들어가고, 증거는 `.moai/reports/t539/verdict.md` 에 모이며, 그 안에서 인용된 로그 경로는 전부 실재한다.
  브랜치 이름(`WT-ctx-blind-double`)이 카드를 식별하지 않으므로 이 세 운반체가 유일한 추적 경로다.

## 3. 범위 밖 (Out of Scope)

### Out of Scope — 프로덕션의 컨텍스트 버림

- 컨텍스트를 받고 버리는 **프로덕션 메서드 42건**은 이 카드에서 손대지 않는다. 내역(합 42):

  | 항목 | 건수 | 근거 |
  |---|---|---|
  | `internal/hook/**` 의 `Handle` | 18 | `hook/*` prod-`no` 21건 중 `Handle` 만 |
  | `internal/hook/quality` formatter/linter | 3 | `formatter.go:43 FormatFile`, `linter.go:37 LintFile`, `linter.go:85 AutoFix` |
  | `internal/web/**` 핸들러 | 14 | `screens.go` 7 + `profile_crud.go` 4 + `handlers.go` 3 |
  | `internal/web/**` 기타 | 2 | `app.go:256`, `glmkey.go:115` |
  | `internal/statusline/git.go` `gitCollector.CollectGitStatus` | 1 | |
  | `internal/update/local.go` | 3 | `:63 CheckLatest`, `:180 Download`, `:190 Replace` |
  | `internal/update/updater.go` `updaterImpl.Replace` | 1 | `:126` |

  **`hook 핸들러` 항목은 `hook/quality` 3건을 포함하지 않는다** — 0.1.0 의 내역이 그 3건을 두 번 세어
  합계가 43 이 됐고, `web` 을 14 로 적어 `app.go` / `glmkey.go` 2건을 누락했다. 위 표는
  `sweep-prod.tsv` 의 `method` × `no` 42행을 재계수한 값이다.
  이들에 대해서는 눈먼 대역이 옳다. 프로덕션이 컨텍스트를 버리는 것 자체가 별도 결함 부류일 수 있으므로
  리드에게 **별도 카드 후보로만 보고**한다.

### Out of Scope — 대역이 아닌 것

- `httptest` **서버 핸들러** func literal 43건(`func(w, r)`)은 클라이언트 대역이 아니다.
  그 테스트들의 클라이언트는 실제 `*http.Client` 라 이미 충실하다.
- 프로덕션 코드의 컨텍스트 전파 방식 자체를 바꾸는 리팩터링. 이 카드는 **가드의 부재**를 고치지,
  현재 옳게 동작하는 전파를 다시 쓰지 않는다.

### Out of Scope — 이미 판정된 경로

- codex **백그라운드** 경로(`fakeCodexSession` / `hangingCodexSession` / `fixedConnSessionRunner`).
  M1 뮤턴트가 검출됐으므로 (d) 가 거짓이고, 수리 대상이 아니다. 단 백그라운드 이외의 codex 경로
  (`codexReviewRPC` 등)는 미측정이며 이 카드의 M3 후보 목록에도 들어 있지 않다.
- goal 러너 대역 5종. (a) 가 아니오이므로 옳게 눈멀었다.

### Out of Scope — M3 생존 뮤턴트의 수리

- 다섯 후보군에서 생존 뮤턴트가 나오더라도 이 카드는 그 대역을 **수리하지 않는다**
  (운영자 결정 2026-09-08). 이 카드가 수리하는 자리는 M2 의 GLM audit 경로 하나뿐이며,
  나머지 생존은 로그와 **패키지별 후속 카드 요청**으로만 남긴다. 측정은 다섯 후보군 전부에 대해
  이 카드에서 끝낸다 — 갈리는 것은 수리의 소속이다.
- `plan.md` §F M3 의 84건 열거에 **들어 있지 않은** 대역은 이 카드의 측정 대상이 아니다.
  경계에 가까워 오해되기 쉬운 것들을 명시한다 — `internal/lsp/core/client_test.go`
  `fakeTransport` / `fakeLauncher` 3건, `internal/lsp/hook/diagnostics_test.go`
  `mockLSPClient` / `mockFallbackDiagnostics` 2건, `internal/lsp/transport/transport_test.go`
  `fakeTransport` 2건.

### Out of Scope — 도구의 제품화

- `ctxsweep` 스캐너를 `internal/` 이하 제품 코드로 옮기거나, CI 게이트로 승격하는 일.
  이 카드에서 그것은 증거 도구이며 `.moai/reports/t539/` 에 머문다.

## 4. 잔여 위험

1. M2 가 가리던 결함은 실운영에서 클라이언트 자체 Timeout(120s)이 상한을 준다 — 무한 매달림이 아니라
   **취소 불응**이다. 이 SPEC 이 고치는 것은 그 결함의 현재 상태가 아니라, 그 결함이 재도입돼도
   아무도 모르는 **가드의 부재**다.
2. `ctxsweep` 는 문법 트리만 본다. 대역이 컨텍스트를 구조체 필드에 저장해 두고 별도 고루틴에서 소비하면
   "본다"로 세지만 그것이 충실하다는 뜻은 아니다. 반대로 위임형 래퍼는 "안 본다"로 세지만 충실하다
   (REQ-CBD-005 의 예외가 이 축을 덮는다).
3. (a) 축 판정 일부는 grep 기반이다. 실물이 컨텍스트를 다른 방식(`cmd.Cancel`, `time.AfterFunc`)으로
   소비하면 "옳게 눈멂"으로 오분류될 수 있다. 이 위험은 뮤턴트 판정이 흡수한다 — 오분류된 후보의
   뮤턴트는 생존할 것이기 때문이다.
   **이 위험은 가설이 아니라 최소 1건 실현됐다**: LSP transport 행의 근거 `ctx.Done 1` 은
   `request.go` 가 아니라 `request_test.go` 의 히트였다(`request*.go` 글롭이 테스트를 삼킨 형태).
   판정 자체는 `request.go:70 ctx.Err()` / `:53 t.Call(ctx, …)` 로 여전히 "예"지만, 근거는 틀렸었다.
   그래서 (a) 축 grep 은 **프로덕션 파일로 한정**한다 — `_test.go` 를 배제하지 않는 글롭을 쓰지 않는다
   (`research.md` §3 방법 주석).
4. M3 후보 다섯은 서로 다른 패키지에 흩어져 있다(`internal/update`, `internal/statusline`,
   `internal/lsp/core`, `internal/lsp/aggregator`, `internal/core/project`, `internal/cli/update/merge`,
   `internal/guardstate`). 수리가 발생하면 검증 범위가 패키지마다 늘어난다 — `plan.md` §D 참조.

## 5. 참고

- 조사 자료: `.moai/reports/t539/sweep.md`(정본), `sweep-test.tsv`, `sweep-prod.tsv`,
  `m1-codex-mutant.log`, `m2-run-{GLM,Audit,Converg}.log`
- **`sweep-raw.tsv` 는 폐기됐다.** `*http.Request` 판정 기준을 `req.Context()` 호출 여부로 정정하기
  **이전**의 1차 스캔이며, `sweep-test.tsv` 와 **6행**에서 판정이 다르다(전부 `yes` → `no`):
  `cli/audit_pin_live_test.go:108`, `cli/mcp_glm_test.go:34`, `cli/update_version_test.go:251`,
  `statusline/usage_test.go:435`, `statusline/usage_test.go:993`, `update/checker_test.go:724`.
  정본은 `sweep-test.tsv` 이며, run-phase 는 `sweep-raw.tsv` 를 근거로 쓰지 않는다.
- 파생 원본 카드: t532(GLM task 경로의 백그라운드 컨텍스트 수리)
- 관련 수리: t514(codex 백그라운드 `context.WithoutCancel`), M1 뮤턴트의 역주입 대상
