# 승인된 구독 출력 정책 구현

## Claim
AuthPKCE의 고정 SubscriptionEndpoint 분기에서만 outgoing max_output_tokens를 생략했다. translate.Request는 수정하지 않았고, 공통 max_tokens 검증과 API 키 매핑을 그대로 사용한다. 생략은 Subscription Store.Owns 확인 뒤 이루어지며 고정 목적지와 SendAuthorized 경계를 유지한다.
TLS mock으로 실제 HTTP body와 Host/path를 읽었다. 합성 인증을 가진 임시 실제 AUTH Store의 구독 요청에는 cap가 없고 API 키 요청에는 정확히 10이 있었다. 양 경로의 누락·0·음수·소수·문자열·null·양방향 중복·int64 범위 초과·거대 지수·body 초과 입력은 400이며 dial 횟수가 유효 요청 1회에서 늘지 않았다. 실제 계정/네트워크 provider는 사용하지 않았다.

## Baseline-attribution
source_session_id: 01a08e7b-6aa0-7361-ab7e-ea8da1f02228.
WT: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified.
git branch --show-current: WT-unified-gateway.
git rev-parse --short HEAD: 81c1d58f9.
git fetch -q origin main && git rev-list --count --left-right origin/main...HEAD: 0 2879.
이 HEAD의 미커밋 작업을 측정했다. 다른 작성자의 파일은 수정하지 않았다.

## Evidence
공통 작업 디렉터리는 위 WT다. 아래 go test/vet는 한 invocation에서 unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache로 실행했다.

RED 명령: go test ./internal/gateway -run 'TestOpenAISubscriptionOutputPolicyWireAndValidation|TestAnthropicNativeRawWireBoundaries' -count=1 -timeout 30s
구현 전 두 시험을 추가해 FAIL을 관측했다:

```text
--- FAIL: TestAnthropicNativeRawWireBoundaries (0.00s)
    anthropic_test.go:347: CRLF fragmented=false exceeded raw limit
    anthropic_test.go:347: mixed fragmented=false exceeded raw limit
    anthropic_test.go:347: CRLF fragmented=true exceeded raw limit
    anthropic_test.go:347: mixed fragmented=true exceeded raw limit
--- FAIL: TestOpenAISubscriptionOutputPolicyWireAndValidation (0.05s)
    openai_subscription_test.go:84: subscription policy or endpoint mismatch
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway	0.587s
FAIL
```
GREEN 명령: go test -race ./internal/gateway -run '^TestOpenAI|^TestAnthropic' -count=1 -timeout 45s -coverprofile=/tmp/cap-native.cover
exit=0:
```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	2.527s	coverage: 61.2% of statements
```
go vet ./internal/gateway ./internal/gateway/translate: exit=0, stdout/stderr 없음.
gofmt -l (이 위임의 네 변경 파일): exit=0, stdout 없음.

go tool cover -func=/tmp/cap-native.cover 관측: openai.go Send 86.4%.

변경 파일 SHA-256:
- internal/gateway/openai.go: 3dd9c4005132c4af8abef0a3f8663ffa7588a958a043b361dbd364aa89d3837f
- internal/gateway/openai_subscription_test.go: 48efa6cb14131cf626d2b960604adc2ceb4c3d74c7f8f0772db416d88ae1e7b8

## Gaps
AC의 사용자 대상 구독 출력 정책 표시는 이 adapter에 UI가 없어 launcher 통합 소유자에게 남는다. 요청별 생성 token 상한 강제나 입력 truncation 부재를 이 수정으로 입증했다고 주장하지 않는다. 독립 재감사 및 실제 moai gpt 통합은 아직 수행하지 않았다.
LSP 전용 baseline은 수집하지 않고 Go compile/race/vet로 확인했다. 함수 커버리지를 전체 저장소 수치로 확대하지 않는다. integration branch GitHub CI run이 repository-wide 판정을 소유하며 현재 PENDING (run ID 없음)이다. Windows 실행은 이 위임에 포함하지 않았다. 커밋·push·PR 없음.

## Residual-risk
실제 provider·root/launcher 활성화는 별도 검증이다. stream.go/gateway_factory.go/receipt/Windows/SPEC/root binding을 수정하지 않았다. 수리자 검증이며 새 독립 감사 판정은 아직 없다.
