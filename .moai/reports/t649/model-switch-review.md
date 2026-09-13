# t649 모델 변경·오류 후 재개 독립 검토

SPEC: SPEC-MOAI-GATEWAY-001

Overall Verdict: **PASS — 지정된 수정 범위에 한정**

검토 회차: 모델 변경 및 API 오류 후 재개 수정의 첫 독립 검토. 전체 SPEC, 전체 릴리스, 실제 구독 서버의 모델 호환성 승인을 뜻하지 않는다.

## Claim

검토 대상은 `conversation/native.go`, `translate/receipt_history.go`, `openai.go`, `cli/gateway_factory.go`와 해당 회귀 테스트다. 다음 조건은 이번 트리의 실행 결과로 확인했다.

- 이전에 완료된 답변이 있고 그 뒤 실패한 요청에 합성 API 오류 및 로컬 명령 기록이 남은 경우, 마지막 성공 모델로 재개할 수 있다.
- 완료된 답변이 없는 기록, 미완료 도구/assistant 출력, 다른 세션의 오류 기록, 사용자가 입력한 명령 마크업은 이 예외로 승인되지 않는다.
- GPT-5.6/Sol 계열과 GPT-6/Astra 사이의 혼합 기록은 원래 영수증 도메인에서 검증한다. owner, session, conversation family 및 공개 접두부와 opaque digest 검증을 유지한다.
- 반대 도메인에 빈 영수증 후보가 있어도 필수 reasoning 영수증을 제거한 요청은 거부한다.
- 사용자에게 표시하는 복구 안내는 고정 문구이며 내부 오류 문자열을 노출하지 않는다.

## Dimension Scores

기본 evaluator profile의 40/25/20/15 가중치와 Functionality/Security 필수 통과 기준을 적용했다. 점수는 위 변경분의 지정 조건에 대한 충족도다. 전체 SPEC 점수는 산정하지 않았다.

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 100/100 | PASS (delta) | `ok  \tgithub.com/modu-ai/moai-adk/internal/cli\t1.111s`; 아래 실제 기록 검사 원문과 패키지 결과 참조 |
| Security (25%) | 100/100 | PASS (delta) | `--- PASS: TestSubscriptionReceiptModelSwitchPreservesSourceBindings (0.07s)` 및 거부 조건별 결과 원문 |
| Craft (20%) | 100/100 | PASS (delta) | `pooled statements=2929/3264 coverage=89.7%`; `go vet` exit 0, 출력 없음 |
| Consistency (15%) | 100/100 | PASS (delta) | 지정 파일 `gofmt -l` exit 0, 출력 없음; 아래 소스 바인딩 확인 |

Weighted delta score: **100/100**. 이 값은 알려진 전체 제품 미검증 항목을 통과로 변경하지 않는다.

## Evidence

모든 명령의 작업 디렉터리:

`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`

### 패키지 테스트 및 커버리지

```text
$ go test -count=1 -coverprofile=/tmp/t649-switch-review.cover ./internal/gateway/conversation ./internal/gateway/translate ./internal/gateway/receipt ./internal/gateway
ok  	github.com/modu-ai/moai-adk/internal/gateway/conversation	0.982s	coverage: 80.0% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.900s	coverage: 91.2% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	1.543s	coverage: 88.9% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway	7.352s	coverage: 92.1% of statements

$ go test -count=1 -run '^TestGatewayFactory.*(Switch|Receipt)' ./internal/cli
ok  	github.com/modu-ai/moai-adk/internal/cli	1.111s
```

`/tmp/t649-switch-review.cover`의 statement count를 합산한 결과:

```text
pooled statements=2929/3264 coverage=89.7%
github.com/modu-ai/moai-adk/internal/gateway/conversation/native.go: statements=141/158 coverage=89.2%
github.com/modu-ai/moai-adk/internal/gateway/translate/receipt_history.go: statements=100/104 coverage=96.2%
github.com/modu-ai/moai-adk/internal/gateway/openai.go: statements=199/218 coverage=91.3%
```

`go tool cover -func=/tmp/t649-switch-review.cover`에서 직접 확인한 핵심 함수:

```text
github.com/modu-ai/moai-adk/internal/gateway/openai.go:280:			openAITranslationError			100.0%
github.com/modu-ai/moai-adk/internal/gateway/openai.go:287:			openAIError				100.0%
github.com/modu-ai/moai-adk/internal/gateway/translate/receipt_history.go:170:	checkObserved				94.4%
```

### 실제 실패 기록의 읽기 전용 재개 검사

원래 JSONL을 수정하지 않고 Go overlay로 `transcriptModel`을 호출했다. 테스트는 지정 UUID, cwd, private config root와 마지막 성공 모델을 검사한다.

```text
$ go test -count=1 -overlay=/tmp/t649-native-resume-readonly-overlay.json -run '^TestReadOnlyActualFailedTranscript$' -v ./internal/gateway/conversation
=== RUN   TestReadOnlyActualFailedTranscript
    actual_readonly_test.go:6: model=gpt-5.6-sol error=<nil>
--- PASS: TestReadOnlyActualFailedTranscript (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/conversation	0.452s
```

### Race 검사 및 거부 조건

```text
$ go test -race -count=1 -run 'TestSubscriptionReceiptModelSwitchPreservesSourceBindings|TestNativeResumeAfterTerminalAPIError|TestOpenAIHistoryErrorGuidesRecoveryWithoutExposingDetails' -v ./internal/gateway/translate ./internal/gateway/conversation ./internal/gateway
=== RUN   TestSubscriptionReceiptModelSwitchPreservesSourceBindings
--- PASS: TestSubscriptionReceiptModelSwitchPreservesSourceBindings (0.07s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.430s
=== RUN   TestNativeResumeAfterTerminalAPIError
=== RUN   TestNativeResumeAfterTerminalAPIError/completed-then-failed-request
=== RUN   TestNativeResumeAfterTerminalAPIError/local-command-after-api-error
=== RUN   TestNativeResumeAfterTerminalAPIError/typed-local-command-spoof
=== RUN   TestNativeResumeAfterTerminalAPIError/repeated-api-error
=== RUN   TestNativeResumeAfterTerminalAPIError/error-only
=== RUN   TestNativeResumeAfterTerminalAPIError/unfinished-tool
=== RUN   TestNativeResumeAfterTerminalAPIError/unfinished-assistant
=== RUN   TestNativeResumeAfterTerminalAPIError/arbitrary-synthetic
=== RUN   TestNativeResumeAfterTerminalAPIError/foreign-error
=== RUN   TestNativeResumeAfterTerminalAPIError/wrong-role
=== RUN   TestNativeResumeAfterTerminalAPIError/error-with-tool-content
=== RUN   TestNativeResumeAfterTerminalAPIError/later-incomplete-request
--- PASS: TestNativeResumeAfterTerminalAPIError (0.22s)
    --- PASS: TestNativeResumeAfterTerminalAPIError/completed-then-failed-request (0.02s)
    --- PASS: TestNativeResumeAfterTerminalAPIError/local-command-after-api-error (0.02s)
    --- PASS: TestNativeResumeAfterTerminalAPIError/typed-local-command-spoof (0.02s)
    --- PASS: TestNativeResumeAfterTerminalAPIError/repeated-api-error (0.02s)
    --- PASS: TestNativeResumeAfterTerminalAPIError/error-only (0.02s)
    --- PASS: TestNativeResumeAfterTerminalAPIError/unfinished-tool (0.02s)
    --- PASS: TestNativeResumeAfterTerminalAPIError/unfinished-assistant (0.02s)
    --- PASS: TestNativeResumeAfterTerminalAPIError/arbitrary-synthetic (0.02s)
    --- PASS: TestNativeResumeAfterTerminalAPIError/foreign-error (0.02s)
    --- PASS: TestNativeResumeAfterTerminalAPIError/wrong-role (0.02s)
    --- PASS: TestNativeResumeAfterTerminalAPIError/error-with-tool-content (0.02s)
    --- PASS: TestNativeResumeAfterTerminalAPIError/later-incomplete-request (0.02s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/conversation	1.718s
=== RUN   TestOpenAIHistoryErrorGuidesRecoveryWithoutExposingDetails
--- PASS: TestOpenAIHistoryErrorGuidesRecoveryWithoutExposingDetails (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.770s
```

### 정적 검사 및 의존성 변경 확인

```text
$ go vet ./internal/gateway/conversation ./internal/gateway/translate ./internal/gateway/receipt ./internal/gateway
[출력 없음, exit 0]
$ gofmt -l internal/gateway/conversation/native.go internal/gateway/conversation/native_test.go internal/gateway/translate/receipt_history.go internal/gateway/translate/receipt_switch_test.go internal/gateway/openai.go internal/gateway/openai_test.go internal/cli/gateway_factory.go internal/cli/gateway_factory_test.go
[출력 없음, exit 0]
$ rg -n '(exec.Command|http.Get|os.Getenv|Authorization|InsecureSkipVerify|ReadFile)' internal/gateway/translate/receipt_history.go internal/gateway/conversation/native.go
[일치 없음]
$ git diff --name-only -- go.mod go.sum
[출력 없음]
$ git status --short -- go.mod go.sum
[출력 없음]
```

패턴 일치 없음은 보안 전체 통과의 근거로 사용하지 않았다. owner/session/family 격리 및 누락 reasoning 거부는 위 실행 테스트로 별도 확인했다. 이번 수정은 의존성 manifest 변경을 포함하지 않는다.

## Baseline-attribution

```text
$ moai session current
01a08e7b-6aa0-7361-ab7e-ea8da1f02228
$ git rev-parse --short HEAD
81c1d58f9
$ git branch --show-current
WT-unified-gateway
$ shasum -a 256 internal/gateway/conversation/native.go internal/gateway/translate/receipt_history.go internal/gateway/openai.go internal/cli/gateway_factory.go
393f0c7562cf3c8e6dfcfabd49378d6760558f1c90147ad847a850721984387b  internal/gateway/conversation/native.go
56bfca634eeba6d955e2d68f62e5b87dde969f0c577f0c738c8ceaec5e7a3df2  internal/gateway/translate/receipt_history.go
a4f00535fcc2dab80dc41ec21a3306614a0a6ccef068d444b552bf8914f29464  internal/gateway/openai.go
a12230425c501e271711ff21adcd22c60d94e9087237340dca124344e616ef3b  internal/cli/gateway_factory.go
```

네 production 파일 hash는 테스트 전후 동일했다. 기존에 변경/미추적 파일이 많은 공동 작업 트리임을 확인했다. 구현 파일 변경, commit, push 및 설치는 수행하지 않았다. 이 보고서만 작성했다.

## Findings

지정된 수정 범위에서 blocking 결함은 관찰하지 못했다.

- F1 [Low, optional, confidence: High] `internal/gateway/conversation` 패키지 전체 커버리지는 80.0%로 기본 profile의 85%보다 낮다. 이번에 검토한 `native.go`는 89.2%, 지정 네 패키지 합산은 89.7%다. 전체 패키지 품질 승인으로 확대하려면 별도 범위에서 부족한 분기를 검토해야 한다. 기존 패키지 커버리지 부족의 도입 시점은 측정하지 않았으므로 이번 수정이 만든 회귀라고 판단하지 않는다. 이 제한된 수정 검토에서 추가 구현을 자동 요구하지 않는다.

## Gaps

- 실제 구독 서버 요청, CLI에서 `/model` 변경 후 화면 답변, 빌드 및 설치는 이 감사자가 실행하지 않았다. 상위 담당자의 별도 실행 증거가 필요하다.
- 전체 SPEC 인수 조건, 전체 저장소 coverage, Windows CI, 1M context 설정은 이번 감사 범위가 아니다.
- 공개 의존성 취약점 DB 검사와 전체 OWASP 감사를 실행하지 않았다. 이번 검토는 변경된 입력 검증·계정/세션 격리·오류 노출에 한정한다.
- CLI 패키지 전체 커버리지는 측정하지 않았다. factory 결합 테스트만 실행했다.

## Residual-risk

실제 공급자의 reasoning ciphertext 모델 간 호환성은 고정된 로컬 fixture만으로 보장할 수 없다. 이번 PASS는 호환성을 켰을 때 기존 영수증 권한 검증을 약화하지 않는다는 범위다. 공급자 측 정책 변화와 긴 실제 대화의 동작은 별도 운영 증거로 관리해야 한다. 이후 네 production 파일이 변경되면 이 판정을 그대로 재사용할 수 없다.

## Recommendations

상위 담당자는 이 hash와 같은 소스로 빌드한 바이너리에서 실제 모델 변경 및 오류 후 재개를 검증한 뒤 로컬 설치를 진행한다. 전체 SPEC/card 상태는 이 delta PASS만으로 완료 처리하지 않는다.
