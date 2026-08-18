# t52 — TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey 실패 조사·수정

- 카드: t52 (칸반 배치 tjv7iy round 3, cluster C)
- 워크트리: `.claude/worktrees/agent-add5f3f520fd7123d` (branch `WT-t52`, base = `release/v3.1.1` fast-forward, 조사 시점 HEAD `051a2fa94`)
- 대상: `internal/cli/codex_review_gate_live_test.go` + 게이트 구현부 — 라이브 실패 `decision="" err=<nil>`

## 1. 주장 (Claim)

실패의 원인은 **게이트 결함**(카드 조사 단계 ③)이다. 이 머신의 codex 계정이 사용량 한도(usage limit)에 걸려 리뷰 턴이 diff를 평가하기 **전에** 실패하는데, codex는 그 실패를 `turn/completed`의 `status:"failed"` + `turn.error`(그리고 `method:"error"` 노티피케이션)로 신호한다. 게이트의 `awaitCodexTurnReview`는 `turn/completed`를 종료 신호로만 쓰고 이 상태를 전혀 소비하지 않았고, codex가 실패 시 내놓는 **자리표 문자열** "Reviewer failed to output a response."를 리뷰 본문인 양 수집한다. `synthesizeReviewOutput`은 이 문자열에서 finding bullet(`- [P1] ...`)을 찾지 못해 verdict "pass"를 합성하고, `err=<nil>`과 함께 ALLOW를 반환한다 — **리뷰가 일어나지 않았는데 "리뷰했고 문제없음"으로 세탁되는 경로**. 이는 `HandleCodexReviewGate` 자신의 주석("a gate that was turned on but structurally unable to reach a verdict [must not] look exactly like a gate that had reviewed the change and found nothing wrong")이 이름 붙인 정확히 그 위험이다.

수정: `turn/completed`의 `status`/`error`를 권위적 종료 판정으로 소비해서, `completed`가 아닌 종료 상태(failed/interrupted)면 **inconclusive + error 표면화**(기존 fail-open 패턴)를 반환하고 합성기를 절대 거치지 않게 한다. fail-open(REQ-MCP-012)은 유지 — 실패한 리뷰어가 세션을 가두지 않는다는 설계는 그대로이되, 실패가 보이지 않게 사라지지 않는다.

## 2. 증거 (Evidence) — 카드에 명시된 3단계 조사 순서대로

### 사전 확인 — 구현부 위치

release/v3.1.1에서 `handleCodexReviewGate`(소문자)로는 테스트 파일만 검색되나, 실제 심볼은 `internal/cli/codex_review_gate.go:66`의 `HandleCodexReviewGate`(대문자)이며 판정 경로는 `runCodexReviewRPC` -> `openCodexSession`(initialize -> thread/start handshake) -> `runTurn` -> `awaitCodexTurnReview` -> `synthesizeReviewOutput` (`internal/cli/mcp_codex.go`). 설치된 codex: `codex-cli 0.147.0` at `/Users/goos/.local/bin/codex` — 라이브 테스트는 skip되지 않고 실제로 실행된다.

### RED — 라이브 실패 재현 (ambient 칸반 env 그대로; `MOAI_KANBAN_LABEL=sync-tjv7iy`, `MOAI_AUTONOMY_TIER=fully-autonomous` 하에서 실측)

```
$ go test ./internal/cli/ -run 'TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey' -count=1 -v -timeout 300s
=== RUN   TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey
    codex_review_gate_live_test.go:89: expected BLOCK on injection+AWS-key fixture; got decision="" err=<nil>
        NOTE: a non-BLOCK here means either the protocol fix did not land OR codex returned an inconclusive/pass verdict on this fixture (a real result — report it).
--- FAIL: TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey (7.25s)
```

`err=<nil>` + decision 빈 값은 게이트의 유일한 비-BLOCK 무오류 경로(판정 합성 후 `return allow, nil`) 도달을 뜻한다 — 세션이 죽거나 RPC가 거부된 것이 아니라, 텍스트가 수집되고 "pass"로 합성되었다는 뜻.

### 단계 ① — codex 버전/모델 drift 여부: **프로토콜 형식 drift 아님(호환 유지), 실제 트리거는 계정 사용량 한도. 단, 게이트가 소비하지 않던 실패 신호 계약이 존재**

임시 진단 프로브(패키지 내 `codexSession` seam을 래핑해 실제 codex 0.147.0과의 NDJSON 트랜스크립트를 양방향으로 녹취 — 조사 후 삭제, 커밋 안 됨)로 verbatim 캡처. 핵심 라인:

```
PROBE RX: {"method":"error","params":{"error":{"message":"You've hit your usage limit. Upgrade to Pro (https://chatgpt.com/explore/pro), visit https://chatgpt.com/codex/settings/usage to purchase more credits or try again at Aug 20th, 2026 4:32 PM.","codexErrorInfo":"usageLimitExceeded","additionalDetails":null},"willRetry":false,"threadId":"01a00ba2-4a23-...","turnId":"01a00ba2-4d0e-..."}}
PROBE RX: {"method":"item/completed","params":{"item":{"type":"exitedReviewMode","id":"...","review":"Reviewer failed to output a response."},"threadId":"01a00ba2-4a23-...","turnId":"01a00ba2-4d0e-..."}}
PROBE RX: {"method":"turn/completed","params":{"threadId":"01a00ba2-4a23-...","turn":{"id":"01a00ba2-4d0e-...","items":[],"itemsView":"notLoaded","status":"failed","error":{"message":"You've hit your usage limit. ... try again at Aug 20th, 2026 4:32 PM.","codexErrorInfo":"usageLimitExceeded",...},"startedAt":1786901581,"completedAt":1786901584,"durationMs":3760}}}
PROBE RESULT: verdict="pass" err=<nil>
PROBE SUMMARY (verbatim, 37 bytes): "Reviewer failed to output a response."
PROBE BULLET MATCH: false
```

- handshake(initialize -> thread/start -> review/start ack)는 0.147.0에서도 동일 형식으로 정상 동작 — 버전 drift로 인한 프로토콜 파손은 아님.
- 턴은 개시 3.7초 만에 `usageLimitExceeded`로 실패 — **fixture는 모델에게 도달조차 못함**(그래서 단계 ②의 내용이 이번 실패와 무관함이 함께 확정).
- 권위 스키마로 종료 상태 계약 확인: `codex app-server generate-json-schema --out /tmp/t52_schema` (0.147.0 자체 생성) -> `TurnStatus: {"enum": ["completed", "interrupted", "failed", "inProgress"]}`, `Turn.error`는 "Only populated when the Turn's status is failed" — 실패 신호는 처음부터 계약에 존재했으나 게이트가 읽지 않았다.

### 단계 ② — fixture가 여전히 injection+AWS-key를 담고 있는지: **충족 (원인 아님)**

`internal/cli/codex_review_gate_live_test.go`의 `fixtureVulnGo()`는 현재도 (a) `exec.Command("sh", "-c", q).Run()` 커맨드 인젝션 싱크와 (b) 접두사 splice로 조립된 `AKIA...EXAMPLE` 액세스 키 + `wJalrXUtnFEMI/...EXAMPLEKEY` 시크릿을 모두 생성한다. 다만 위 트랜스크립트가 증명하듯 이번 실패에서 fixture 내용은 평가되지 못했다 — 턴이 쿼터 소진으로 죽었기 때문. fixture 결함 아님.

### 단계 ③ — 게이트가 decision을 잃는 경로: **여기가 결함 (gate defect)**

`internal/cli/mcp_codex.go`의 수정 전 코드:

```go
case "turn/completed":
    return bestCodexReviewText(reviewText, agentText)   // 상태/에러 미검사 — 종료 신호로만 사용
```

- `awaitCodexTurnReview`는 `turn/completed`의 `turn.status`/`turn.error`를 전혀 파싱하지 않았고 `method:"error"` 노티피케이션도 무조건 폐기(switch에 case 없음).
- 그 결과 codex의 자리표 문자열 "Reviewer failed to output a response."가 진짜 리뷰 본문처럼 흘러들어가 bullet 매칭 실패 -> `synthesizeReviewOutput`이 "pass" 합성 -> `HandleCodexReviewGate`이 `err=<nil>` ALLOW 반환.
- 단위 수준 RED(캔 세션, 쿼터 불필요 — 라이브 실패 형상을 그대로 재현):

```
$ go test ./internal/cli/ -run 'TestReviewGate_FailedTurnSurfacesErrorNotPass|TestRunCodexReviewRPC_FailedTurnIsInconclusiveNotPass|TestReviewGate_ErrorNotificationWithRetryingTurnStillReviews' -count=1 -v
=== RUN   TestReviewGate_FailedTurnSurfacesErrorNotPass/failed_with_usage-limit_error
    codex_review_gate_test.go:222: a failed turn MUST surface an error; got err=<nil> (the gate fabricated a pass)
=== RUN   TestReviewGate_FailedTurnSurfacesErrorNotPass/interrupted_without_error_object
    codex_review_gate_test.go:222: a interrupted turn MUST surface an error; got err=<nil> (the gate fabricated a pass)
--- FAIL: TestRunCodexReviewRPC_FailedTurnIsInconclusiveNotPass
    codex_review_gate_test.go:251: failed turn must return an error; got nil with verdict="pass"
(TestReviewGate_ErrorNotificationWithRetryingTurnStillReviews — PASS: willRetry 후 completed되는 턴은 여전히 정상 리뷰여야 한다는 대조군, 수정 전후 동일)
```

### 수정 (fix)

`internal/cli/mcp_codex.go`:

1. `awaitCodexTurnReview`가 `(string, error)` 반환 — `turn/completed`에서 새 파서 `codexTurnFailure`가 `turn.status`/`turn.error`를 읽어 `completed`가 아닌 종료 상태면 codex 자체 메시지를 실은 error 반환. `error` 노티피케이션 단독(willRetry 시맨틱스)은 판정 근거로 삼지 않음 — 권위는 오직 `turn/completed`의 종료 상태(일시 오류 재시도 후 성공한 리뷰를 오염시키지 않기 위해).
2. `runTurn`은 turn 실패 시 수집 텍스트(자리표)를 버리고 `inconclusiveReview(오류) + error` 반환 — 기존 fail-open-표면화 패턴(`HandleCodexReviewGate`의 `return allow, rpcErr`)과 동일 궤적.

`internal/cli/codex_review_gate_live_test.go`: 라이브 테스트를 2분기 계약으로 교체 — (1) 턴이 `completed`면(err==nil) BLOCK **필수**(보안 단정, 그대로), (2) 턴 자체가 실패하면(err!=nil) ALLOW + 오류 표면화 **필수**. 치명 형상은 `err==nil && 비-BLOCK`뿐 — 즉 "리뷰가 일어났는데 문제없었다"는 주장을 아무 실제 리뷰도 뒷받침하지 않는 경우. 기대를 낮춘 것이 아니라, 쿼터 유무와 무관하게 항상 단정 가능한 불변식으로 올린 것이다(쿼터가 있는 환경에선 BLOCK 경로가 여전히 완전 단정됨).

`internal/cli/codex_review_gate_test.go`: 신규 3건 — `TestReviewGate_FailedTurnSurfacesErrorNotPass`(failed/interrupted 표), `TestRunCodexReviewRPC_FailedTurnIsInconclusiveNotPass`(verdict가 pass가 아니라 inconclusive여야 함), `TestReviewGate_ErrorNotificationWithRetryingTurnStillReviews`(재시도 후 완료 턴은 실제 리뷰로 BLOCK).

### GREEN — 라이브 (실제 codex 0.147.0 대상, ambient 동일)

```
$ go test ./internal/cli/ -run 'TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey' -count=1 -v -timeout 300s
=== RUN   TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey
    codex_review_gate_live_test.go:118: codex review turn did not complete — gate failed open with the error surfaced (correct behavior): codex review turn failed: You've hit your usage limit. Upgrade to Pro (https://chatgpt.com/explore/pro), visit https://chatgpt.com/codex/settings/usage to purchase more credits or try again at Aug 20th, 2026 4:32 PM.
--- SKIP: TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey (4.47s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	7.949s
```

게이트가 이제 codex의 실제 실패 사유를 verbatim으로 표면화한다(더 이상 `decision="" err=<nil>` 세탁 없음). 턴이 완료되는 환경(쿼터 복귀 후)에는 BLOCK 분기가 그대로 단정된다.

### GREEN — 단위·배터리

```
$ go test ./internal/cli/ -run 'CodexReviewGate|ReviewGate_|RunCodexReviewRPC' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	7.289s
(상세 -v: --- PASS 29건, --- SKIP 1건 = 라이브 테스트의 의도된 skip)
$ go test ./internal/cli/ -run 'CodexSession|CodexJob|CodexTask|CodexRPC|TestSynthesize|TestAwait|TestBestCodex|McpCodex|Convergence' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.978s   (--- PASS 52건)
$ go test ./internal/cli/ -run 'TestCodexTask_|TestCodexReviewGate_Denies' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.581s   (--- PASS 15건)
```

정적 검증: `gofmt -l` — 변경 3개 파일 출력 없음(패키지 내 다른 미정렬 파일들은 HEAD 시점부터 존재, 미건드림), `go vet ./internal/cli/` 통과.

## 3. Baseline 귀속

- 전 실측은 위 워크트리(HEAD `051a2fa944ce5f8294ae44a0d9964f27cacaa232` = release/v3.1.1 선두, merge 후 `git rev-list --count` 양방향 0 확인)에서 이루어졌고 출력은 해당 실행의 verbatim이다.
- RED는 ambient 칸반 환경(`MOAI_KANBAN*` 3종 세팅, `MOAI_AUTONOMY_TIER=fully-autonomous`)에서 재현 — 카드가 지시한 "관찰된 실패 조건과 동일 환경에서 먼저 재현" 충족. 이 결함은 환경변수가 아니라 codex 계정 쿼터에 구동되므로 스크럽 실행은 불필요(변수 조작으로 판정이 바뀌지 않음을 트랜스크립트가 직접 증명).
- `git status --short` 실행 전후 동일 — 테스트는 트리를 오염시키지 않음(카드 기록과 부합).

## 4. 미검증 (Gaps)

- **쿼터 복귀 후(2026-08-20 이후)의 라이브 BLOCK 경로** — 이번 실행에서는 codex 계정이 사용량 한도에 걸려 `turn/completed status:"completed"`인 실제 리뷰를 관측할 수 없었다. BLOCK 합성 경로 자체는 캔 세션 단위 테스트(`TestReviewGate_CodexFailBlocks` — bullet 리뷰 -> BLOCK, `TestReviewGate_ErrorNotificationWithRetryingTurnStillReviews` 포함)로 결정론적 검증됨.
- `-race` 미실시(CI `test-race` 잡 소관 — 로컬 전체 스위트/레이스 회피 규율).
- `internal/cli` 외 패키지 미실시 — 변경이 전부 `internal/cli` 내부(핸들 시그니처는 모듈 외 미노출)이므로 파급 없음. 전체 판정은 배치 PR의 CI 몫.
- `codex_live_protocol_probe_test.go`(opt-in, `MOAI_CODEX_LIVE_PROBE=1`)는 실행하지 않음 — 매 턴이 실쿼터를 소모하며 현재 쿼터 소진 상태.

## 5. 잔여 위험 (Residual-risk)

- **사용량 한도는 2026-08-20 4:32 PM(원문 표기 그대로)까지 지속** — 그 전까지 이 머신의 라이브 테스트는 의도된 skip으로 통과하며, 쿼터가 회복되면 BLOCK 분기가 자동으로 재단정된다.
- `awaitCodexTurnReview`의 EOF/deadline 경로(턴 완료 노티 없이 스트림 종료)는 기존 동작 유지 — 수집 텍스트 합성. 이 경로가 자리표를 실을 가능성은 낮으나(EOF 시 보통 텍스트도 없음) `turn/completed` 수신 전 사망은 여전히 "무평가"로 처리된다. 별도 카드 후보.
- `codex_task`/백그라운드 잡의 실패한 턴도 이제 error를 반환한다(이전에는 자리표에서 "pass" 합성 가능) — 잡 레코드 소비자가 error를 정상 처리하는지는 `CodexJob`/`CodexTask` 배터리 52+15건 통과로 간접 검증됨.
- 게이트가 `decision:"block"`을 내는 유일한 경로는 bullet 합성뿐 — codex가 bullet 없이 통과하는 리뷰는 여전히 ALLOW다(설계된 fail-open). fixture가 모델에게 도달했는데도 BLOCK이 안 나오는 형상은 라이브 테스트의 `err==nil && 비-BLOCK` 치명 분기가 계속 잡는다.
