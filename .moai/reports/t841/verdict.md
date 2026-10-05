# t841 판정서 — GPT reasoning-effort 허용표를 OpenAI 공식 모델 페이지와 일치

- 카드: t841 (Class B, Tier S)
- 브랜치: `WT-effort-allowlist` (base develop `4da5d1c4e`)
- 코드 커밋: `2b7d9b2ad`

## Claim

1. `internal/gateway/translate/native_policy.go`의 effort 허용표가 모델별로 나뉘어 OpenAI 공식 모델 페이지의 집합과 일치한다. gpt-5.6-sol · gpt-5.6-terra · gpt-5.6-luna는 `none/low/medium/high/xhigh/max`를 받고, gpt-6-astra는 `low..max`만 받으며 `none`은 거부한다.
2. gpt-6-astra `max`→`xhigh` 클램프(`maxEffortRejected`)는 제거됐고, 허용된 값은 Responses API `reasoning.effort`에 그대로 투사된다. `none`도 문자열 그대로 실린다.
3. 전략 보고서(`.moai/reports/model-matrix-strategy-20260913.md` 2항)의 주장은 공식 페이지와 모순되지 않는다. 보고서가 언급하지 않은 terra는 sol·luna와 같은 집합이다.
4. 공식 페이지는 t695 D2 실측(astra `max` 4/4 non-200)과 어긋난다. 리드 지시대로 공식 문서를 기준으로 구현했고, 이 모순은 Residual-risk에 남긴다.

## Evidence

### 공식 모델 페이지 (WebFetch, 2026-09-14)

`platform.openai.com/docs/models/<model>`은 301로 `developers.openai.com`으로 넘어가며, 최종 URL에서 읽은 문장은 다음과 같다.

| 모델 | 최종 URL | 인용 |
|---|---|---|
| gpt-6-astra | https://developers.openai.com/api/docs/models/gpt-6-astra | "`reasoning.effort` supports `low`, `medium`, `high`, `xhigh`, and `max`." |
| gpt-5.6-sol | https://developers.openai.com/api/docs/models/gpt-5.6-sol | "Reasoning.effort supports: none, low, medium (default), high, xhigh, and max." |
| gpt-5.6-luna | https://developers.openai.com/api/docs/models/gpt-5.6-luna | "Reasoning.effort supports: none, low, medium (default), high, xhigh, and max." |
| gpt-5.6-terra | https://developers.openai.com/api/docs/models/gpt-5.6-terra | "Reasoning.effort supports: none, low, medium (default), high, xhigh, and max." |

### RED (수리 전, base `4da5d1c4e` + 테스트 교체)

```
$ go test ./internal/gateway/translate/ -run TestNativePolicyGPTEffortAllowlistAndMapping
--- FAIL: TestNativePolicyGPTEffortAllowlistAndMapping (0.00s)
    native_policy_test.go:200: gpt-6-astra max: wire effort is not max
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/translate	0.432s
```

맵 순회 순서상 astra `max` 클램프에서 먼저 멈췄다. sol/luna/terra `none` 거부는 같은 테스트의 다음 반복에서 잡히는 항목이며, 수리 후 GREEN으로 함께 검증됐다.

### GREEN (수리 후, `2b7d9b2ad`)

```
$ go test ./internal/gateway/...
FAIL	github.com/modu-ai/moai-adk/internal/gateway	7.366s
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	5.171s
ok  	github.com/modu-ai/moai-adk/internal/gateway/conversation	3.061s
ok  	github.com/modu-ai/moai-adk/internal/gateway/opaque	0.499s
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	4.288s
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	4.629s

$ go test ./internal/gateway/ -skip TestAppServerSubprocessHTTPToolContinuation
ok  	github.com/modu-ai/moai-adk/internal/gateway	8.444s

$ go vet ./internal/gateway/...        → 출력 없음, exit 0
$ golangci-lint run ./internal/gateway/...
0 issues.
```

`internal/gateway` 패키지의 유일한 실패 `TestAppServerSubprocessHTTPToolContinuation`은 python3 가짜 앱 서버 서브프로세스를 띄우는 통합 테스트로, `git archive HEAD`로 내보낸 base `4da5d1c4e` 사본에서도 동일하게 실패한다.

```
$ (base 4da5d1c4e 사본) go test ./internal/gateway/ -run TestAppServerSubprocessHTTPToolContinuation -v
    appserver_integration_test.go:71: app server start failed
--- FAIL: TestAppServerSubprocessHTTPToolContinuation/chatgpt (0.00s)
--- FAIL: TestAppServerSubprocessHTTPToolContinuation/apiKey (0.00s)
```

## Baseline-attribution

- base: develop `4da5d1c4e` (이 워크트리 HEAD, 작업 시작 시 `git rev-parse --short HEAD`로 확인)
- 모든 명령은 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t841`에서 이 실행 중에 관측했다. 기존-실패 판별만 HEAD 내보내기 사본(scratchpad)에서 쟀다.

## 변경 파일

- `internal/gateway/translate/native_policy.go` — `gptEffortAllowlist`에 `none` 추가(GPT 프로필 합집합), `gptModelEffortAllowed` 신설(astra `none` 거부), `maxEffortRejected`와 `applyGPT`의 클램프 제거, 주석에 t695 D2 이력 유지 + 공식 문서가 기준임을 명시
- `internal/gateway/translate/request.go` — `nativePolicy` 뒤에 모델별 effort 검사 추가(`unsupported effort`), `applyGPT` 시그니처에서 model 제거
- `internal/gateway/translate/native_policy_test.go` — 허용표 테스트를 공식 모델별 집합으로 교체, astra `none` 거부 단언 추가

## Gaps

- 실제 업스트림에 요청을 보내 `none`·astra `max`가 200을 돌려주는지는 재지 않았다. 공식 문서 문장만 근거다.
- `TestAppServerSubprocessHTTPToolContinuation`의 실패 원인(로컬 python3 픽스처)은 진단하지 않았다. 이 카드 범위 밖이다.
- Anthropic 프로필 허용표(`{high}`)는 손대지 않았고, 재검증도 하지 않았다.
- CI(darwin/windows 매트릭스)는 push 금지 지시에 따라 돌지 않았다.

## Residual-risk

- 공식 문서와 t695 D2 실측이 astra `max`에서 서로 어긋난다. 구독 업스트림이 문서와 달리 여전히 `max`를 거부하면, 이 수리 뒤 astra `max` 요청은 400을 그대로 받는다(이전엔 조용히 xhigh로 낮춰 200). 실측 재확인 카드가 필요하다.
- `none`은 `thinking.type: adaptive`와 함께 오면 통과한다(adaptive는 "검증된 effort 존재"만 요구). 업스트림이 이 조합을 거부하는지 미확인.
