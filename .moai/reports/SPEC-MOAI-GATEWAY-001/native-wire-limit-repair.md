# Native SSE 원문 상한 수리

## Claim
nativeBody Scanner에 raw advance 계수를 추가했다. CRLF 정규화 전 소비 byte를 확인하여 전체 응답 상한 초과면 성공 terminal을 내보내지 않는다. Scanner의 bounded buffer, 기존 frame 크기 검사와 상태/취소/Close 구조는 유지했다.
새 시험에서 LF·CRLF·혼합 줄바꿈의 원문 상한 -1/정확/+1을 일반 reader와 1 byte reader로 확인했다. terminal delimiter 누락/미완료 comment/terminal 누락 EOF는 오류이며 terminal이 없다. 기존 blocked-read 취소·상태 adversary·interleaved tool 시험도 같은 race 묶음에 포함했다.

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


독립 원래 실패 probe 재실행:
go test -overlay /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/gateway-native-wire-audit-ybwwqxec/overlay.json ./internal/gateway -run '^TestNativeWireRawLimitAudit$' -count=1 -v -timeout 30s
exit=0:
```text
=== RUN   TestNativeWireRawLimitAudit
    native_wire_audit_test.go:5: CRLF=false bytes=622 limit=621 success=false err=native response stream failed
    native_wire_audit_test.go:5: CRLF=false bytes=622 limit=622 success=true err=<nil>
    native_wire_audit_test.go:5: CRLF=false bytes=622 limit=623 success=true err=<nil>
    native_wire_audit_test.go:5: CRLF=true bytes=640 limit=639 success=false err=native response stream failed
    native_wire_audit_test.go:5: CRLF=true bytes=640 limit=640 success=true err=<nil>
    native_wire_audit_test.go:5: CRLF=true bytes=640 limit=641 success=true err=<nil>
--- PASS: TestNativeWireRawLimitAudit (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.473s
```
go tool cover -func=/tmp/cap-native.cover 관측: newNativeBody 100.0%, Close 100.0%, Read 93.8%, next 97.1%.

변경 파일 SHA-256:
- internal/gateway/anthropic.go: d1abed719f6ff5e80a6cf8bce0a7fe8fd8be18786fc4461d1040bc86ef817828
- internal/gateway/anthropic_test.go: df51371e1c84497da7937dbb9c13b64629044ddf19e9880e0f60437492b32097

## Gaps
기존 LF/CRLF 파싱 범위를 넓히지 않았다. CR-only delimiter나 terminal 뒤 미소비 원문 검증은 추가하지 않았다. 이번 수정은 전체 MaxOutputBytes 계수이며 MaxEventBytes의 정규화된 frame 의미를 개편하지 않았다. native adaptive/title/receipt 기능을 추가하거나 입증하지 않았다.
LSP 전용 baseline은 수집하지 않고 Go compile/race/vet로 확인했다. 함수 커버리지를 전체 저장소 수치로 확대하지 않는다. integration branch GitHub CI run이 repository-wide 판정을 소유하며 현재 PENDING (run ID 없음)이다. Windows 실행은 이 위임에 포함하지 않았다. 커밋·push·PR 없음.

## Residual-risk
실제 provider·root/launcher 활성화는 별도 검증이다. stream.go/gateway_factory.go/receipt/Windows/SPEC/root binding을 수정하지 않았다. 수리자 검증이며 새 독립 감사 판정은 아직 없다.
