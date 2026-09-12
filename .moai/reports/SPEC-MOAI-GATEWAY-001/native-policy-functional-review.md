# Native 정책 및 출력 상한 독립 감사

## Claim

SPEC: SPEC-MOAI-GATEWAY-001 0.9.0, native 정책과 승인 출력 상한 변경분.
Overall Verdict: **FAIL — F1, CRLF 이벤트 원문 크기 상한 위반**
Overall Score: **83.75/100, 변경 범위 한정**.
Iteration: 1. 이전 전체 응답 상한 수리 기록과 이번 이벤트 한 개의 상한 결함은 구별한다.

네 GPT exact 모델 high/title 매핑, title 결과 검증, Anthropic profile의 thinking/signature 보존, metadata/keep-all receipt gate, 고정 AuthPKCE만 출력 상한 생략 및 API 키 값 보존은 로컬 범위 시험을 통과했다. LF/CRLF 전체 응답 상한도 통과했다. 다만 이벤트별 MaxEventBytes는 CRLF를 제거한 뒤 크기를 계산하여 실제 원문 상한을 넘는 이벤트에서 성공 terminal을 내보낸다.

## Dimension Scores

활성 profile은 default.md이며 hierarchical 모드는 아니다. Functionality와 Security의 must-pass 조건을 적용한다. F1로 Functionality가 실패하여 가중합과 관계없이 전체 FAIL이다.

| Dimension | Score | Verdict | Evidence — 이번 실행 원문 발췌 |
|---|---:|---|---|
| Functionality (40%) | 75/100 | FAIL | `CRLF=true fragmented=false largest_raw_event=148 event_limit=147 success=true err=<nil>` |
| Security (25%) | 75/100 | PASS, 범위 한정 | `--- PASS: TestNativeIdentityNeverFallsThroughOrdinaryMetadata (0.00s)` |
| Craft (20%) | 100/100 | PASS, macOS 두 패키지 | `coverage: 91.9% of statements` / `coverage: 93.6% of statements` |
| Consistency (15%) | 100/100 | PASS, 범위 한정 | 대상 gofmt 및 두 패키지 go vet 출력 없음, exit 0 |

## Findings

- **F1 [Medium] [blocking] [confidence: High] internal/gateway/anthropic.go:486** — Scanner.Text가 CRLF를 제거한 뒤 `frame.Len()+len(line)+1`을 MaxEventBytes와 비교한다. 전체 응답 계수는 split 함수의 원문 advance를 사용하지만 이벤트 계수에는 같은 보정이 없다. 독립 overlay에서 최대 원문 이벤트 148 byte, 한도 147일 때 전체 reader와 one-byte reader 모두 성공했다. LF의 145/144 음성 및 LF/CRLF exact/+1 양성 대조군을 함께 실행했다. **Required fix:** 이벤트별 크기도 원문 소비 byte로 계산하고 이벤트 종료에서 초기화한다. 전체 응답 계수와 bounded streaming을 보존한다. LF/CRLF/mixed와 분할 reader의 limit−1/exact/+1 및 EOF·취소 회귀를 추가한다. 전체 응답 한도가 남아 있으므로 이번 resource bound 결함을 Medium으로 평가했다.

추가로 검증된 결함은 없다. receipt/launcher 미구현 접점은 아래 Gaps이며, 이번 부분 구현의 완료 범위에 포함하지 않았다.

## Evidence

CWD는 아래 Baseline WT다. 작성자 보고서의 PASS를 가져오지 않고 직접 실행했다.

```text
$ unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL Z_AI_API_KEY && go test -race ./internal/gateway ./internal/gateway/translate -count=1 -timeout=90s -coverprofile=/tmp/gateway-native-independent.cover
ok  	github.com/modu-ai/moai-adk/internal/gateway	9.192s	coverage: 91.9% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.982s	coverage: 93.6% of statements
$ go vet ./internal/gateway ./internal/gateway/translate
(no output; exit 0)
$ gofmt -l internal/gateway/anthropic.go internal/gateway/openai.go internal/gateway/input_estimate.go internal/gateway/translate/native_policy.go internal/gateway/translate/request.go internal/gateway/translate/response.go
(no output; exit 0)
```

집중 시험 명령:

```sh
unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL Z_AI_API_KEY && go test ./internal/gateway ./internal/gateway/translate -run 'TestNativePolicy|TestNativeTitle|TestNativeIdentity|TestNativeThinking|TestInputEstimate|TestOpenAISubscriptionOutputPolicyWireAndValidation|TestAnthropicNativeRawWireBoundaries' -count=1 -v -timeout=45s
```

원문 PASS 행 발췌:

```text
--- PASS: TestAnthropicNativeRawWireBoundaries (0.00s)
--- PASS: TestNativeThinkingPolicyAndPreservation (0.00s)
--- PASS: TestNativeThinkingStream (0.00s)
--- PASS: TestInputEstimateIncludesOutputSchemaOnly (0.00s)
--- PASS: TestNativePolicyWireAndProviderIsolation (0.00s)
--- PASS: TestInputEstimateRejectsMalformedOutputProjection (0.00s)
--- PASS: TestNativePolicyRequestRejectsMalformedBeforeWire (0.00s)
--- PASS: TestOpenAISubscriptionOutputPolicyWireAndValidation (0.03s)
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.838s
--- PASS: TestNativePolicyGPTMapping (0.00s)
--- PASS: TestNativePolicyInvalidAndReceiptGate (0.00s)
--- PASS: TestNativeTitleResult (0.00s)
--- PASS: TestNativePolicySyntaxMatrix (0.00s)
--- PASS: TestNativeTitleStreamPositiveAndBounds (0.00s)
--- PASS: TestNativeIdentityNeverFallsThroughOrdinaryMetadata (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.400s
```

직접 읽고 실행한 시험은 고정 strict schema/high, null·중복·미지원 정책, 미검증 metadata/keep-all 거절, title의 잘못된 JSON·추가 필드·잘림·성공 terminal 없음, 동일 native model body, signature delta·빈 thinking 보존을 검증한다. 출력 상한 시험은 합성 AUTH Store와 로컬 TLS receiver에서 실제 body·Host/path·dial 횟수를 검사한다. 실제 외부 provider 요청은 없었다.

F1 독립 overlay 명령:

```sh
unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL Z_AI_API_KEY && go test -overlay /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/native-event-audit-e78tbex_/overlay.json ./internal/gateway -run '^TestNativeEventRawLimitIndependent$' -count=1 -v -timeout=30s
```

원문 출력:

```text
=== RUN   TestNativeEventRawLimitIndependent
    native_event_independent_test.go:11: CRLF=false fragmented=false largest_raw_event=145 event_limit=144 success=false err=native response stream failed
    native_event_independent_test.go:11: CRLF=false fragmented=false largest_raw_event=145 event_limit=145 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=false fragmented=false largest_raw_event=145 event_limit=146 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=false fragmented=true largest_raw_event=145 event_limit=144 success=false err=native response stream failed
    native_event_independent_test.go:11: CRLF=false fragmented=true largest_raw_event=145 event_limit=145 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=false fragmented=true largest_raw_event=145 event_limit=146 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=true fragmented=false largest_raw_event=148 event_limit=147 success=true err=<nil>
    native_event_independent_test.go:12: raw event over limit accepted
    native_event_independent_test.go:11: CRLF=true fragmented=false largest_raw_event=148 event_limit=148 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=true fragmented=false largest_raw_event=148 event_limit=149 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=true fragmented=true largest_raw_event=148 event_limit=147 success=true err=<nil>
    native_event_independent_test.go:12: raw event over limit accepted
    native_event_independent_test.go:11: CRLF=true fragmented=true largest_raw_event=148 event_limit=148 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=true fragmented=true largest_raw_event=148 event_limit=149 success=true err=<nil>
--- FAIL: TestNativeEventRawLimitIndependent (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway	0.401s
FAIL
```

추가 독립 overlay는 네 GPT 모델 title/ordinary context를 병렬 실행하여 기본 effort 유지, 요청별 title 검증 분리, 순서대로 나뉜 두 output_text의 title JSON 합성을 확인했다.

```text
$ go test -race -overlay /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/native-event-audit-e78tbex_/overlay.json ./internal/gateway/translate -run '^TestNativePolicyIndependentIsolation$' -count=1 -v -timeout=30s
```

원문 종료 부분:

```text
--- PASS: TestNativePolicyIndependentIsolation (0.00s)
    --- PASS: TestNativePolicyIndependentIsolation/gpt-5.6-luna (0.00s)
    --- PASS: TestNativePolicyIndependentIsolation/gpt-6-astra (0.00s)
    --- PASS: TestNativePolicyIndependentIsolation/gpt-5.6-terra (0.00s)
    --- PASS: TestNativePolicyIndependentIsolation/gpt-5.6-sol (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.355s
```

초기 context fixture는 전체 JSON에서 text 문자열을 찾아 정상 input_text의 text 필드까지 잘못 잡았다. decode한 top-level 키를 확인하도록 시험을 고쳤다. 이 자체 fixture 오류는 제품 결함 RED가 아니다.

보안 경계 및 manifest probe:

```text
$ rg -n 'native receipt authorization required|Subscription.Owns|delete\(body, "max_output_tokens"\)|MaxOutputBytes - b.total' internal/gateway/{openai.go,anthropic.go} internal/gateway/translate/request.go
internal/gateway/anthropic.go:443:		remaining := b.limits.MaxOutputBytes - b.total
internal/gateway/translate/request.go:84:			return nil, nil, errors.New("native receipt authorization required")
internal/gateway/translate/request.go:112:			return nil, nil, errors.New("native receipt authorization required")
internal/gateway/openai.go:114:		if a.config.Subscription == nil || !a.config.Subscription.Owns(q.Credential) {
internal/gateway/openai.go:124:		delete(body, "max_output_tokens")
$ rg -n '^go |golang.org/x/sys ' go.mod
3:go 1.26.8
30:	golang.org/x/sys v0.47.0
```

검색은 경계 식별용이며 텍스트 존재만으로 보안 PASS를 내리지 않는다. 송신 0과 metadata 거절은 위 시험으로 판정했다.

## Baseline-attribution

```text
$ git rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
$ git branch --show-current
WT-unified-gateway
```

WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`의 macOS dirty tree다. 시작과 보고서 작성 직전 HEAD/branch는 같다. source_session_id `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`은 부모가 전달했으며 새 runtime UUID 관측을 주장하지 않는다. 제품·SPEC·workflow·원격 상태는 수정하지 않았다. 본 보고서와 verdict 사본만 소유하며 다른 작성자의 Windows AUTH·receipt 파일은 보존한다.

감사한 소스 SHA-256:

```text
anthropic.go c8e3723f0631135f005cb9184471561da7dcb3e75b90343658faf926399f7d45
openai.go 3dd9c4005132c4af8abef0a3f8663ffa7588a958a043b361dbd364aa89d3837f
input_estimate.go 35000d38eb83bce9cadd33a0f6d25b8738111bfb3f2bcb990350521b55be6280
translate/native_policy.go 2b61bc9334989ed252d3e8bf0e75f216d62714bc473271990561fba950c651ce
translate/request.go b2f988c596da5c1a04bced149e6afd6702518e03e50485ce30e089cd08eebfcd
translate/response.go 34861010f5c28d2771c9c6ee0f373e2c50e6079dcd478f10f8049fd898dca792
translate/stream.go a9b1699eb6f624f56e9a9158c38f01ad6f3a29c899e41129cc5653536aa636c5
```

## Gaps

- receipt codec·UUID bootstrap·metadata 귀속 승인·foreign strip·launcher profile binding은 완료로 판정하지 않는다. KeepAll/UserID의 명시 거절은 통합 완료가 아니다.
- 실제 네 모델 title/main·도구/회상/continue/resume·provider OAuth·Claude native 실행은 미관측이다. mock wire나 profile 이름을 live capability로 확대하지 않는다.
- 구독 출력 정책의 사용자 표시는 launcher 인계 사항이다. 생성 token 상한 강제를 주장하지 않는다.
- 272000/872000/95 cap 활성화, Claude/GLM 실제 cap, 큰 입력·이미지 수용은 미관측이다. 계량은 schema를 포함한 JSON projection 추정이며 tokenizer 상한이 아니다.
- Windows native, repository-wide CI run ID/artifact, LSP 전용 baseline, 취약점 DB 대조는 실행하지 않았다. 일반 JSON Schema/manual reasoning budget은 이번 지원 범위가 아니다.

## Residual-risk

F1 수정은 전체 byte 계수·이벤트별 검증·취소·EOF·성공 terminal 조건을 보존해야 한다. title 검증의 통과는 후속 receipt codec의 opaque/public projection 통합을 대신하지 않는다. native signature 보존은 문법과 byte 보존이며 진위 인증이 아니다.

## Recommendations

F1을 좁게 수정하고 독립 LF/CRLF 이벤트 probe 및 기존 전체 응답·취소·EOF 회귀를 재실행한다. 정책/cap의 통과 범위는 유지하되 전체 목표는 receipt·launcher·실제 통합 증거까지 남긴다. push/PR/merge/dispatch는 수행하지 않았다.
