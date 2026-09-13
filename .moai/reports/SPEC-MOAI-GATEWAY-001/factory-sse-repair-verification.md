# Factory·SSE F1/F2 수리 검증

## Claim
독립 감사 F1/F2를 한정 수리했다. SSE는 Scanner가 정규화하기 전 소비한 원문 advance를 계수한다. Scanner의 기존 MaxEventBytes 버퍼 제한과 취소·Close 흐름은 유지한다. read-ahead를 처리 완료한 메시지로 잘못 계산하지 않는다. 최상위 private JSON 키는 version/session_token/model_ids의 정확한 철자만 허용한다. 기존 ValidateJSONObject의 동일 키 중복 거절은 유지한다.

TDD: 수정 전에 새 두 회귀 시험을 실행하여 실패를 관측했다. GREEN 후 재실행 및 기존 독립 overlay를 실행했다. 추가로 원문 경계 시험을 1 byte 단위 reader로 바꿔 LF/CRLF/혼합/주석 내부 CR을 양 모드에서 상한 -1/정확/+1로 확인했다. JSON 시험은 세 키의 대문자/첫 글자 대문자, 단독 alias/앞뒤 중복/동일 중복을 실제 GPT Store 개방 전 거절하며 canonical payload의 개방 양성 대조군을 둔다. 토큰 값은 출력하지 않는다.

## Evidence
공통 작업 디렉터리: 아래 Baseline-attribution의 WT.
공통 시험 env: unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache

### RED
명령: go test ./internal/cli ./internal/gateway/translate -run 'TestGatewayFactoryExactPayloadKeysBeforeStore|TestStreamRawWireByteBoundaries' -count=1 -timeout 30s
go test 결과 FAIL. 원문:

```text
--- FAIL: TestGatewayFactoryExactPayloadKeysBeforeStore (0.00s)
    gateway_factory_test.go:308: field 0 alias or duplicate reached resource
    gateway_factory_test.go:308: field 0 alias or duplicate reached resource
    gateway_factory_test.go:308: field 0 alias or duplicate reached resource
    gateway_factory_test.go:308: field 0 alias or duplicate reached resource
    gateway_factory_test.go:308: field 0 alias or duplicate reached resource
    gateway_factory_test.go:308: field 0 alias or duplicate reached resource
    gateway_factory_test.go:308: field 1 alias or duplicate reached resource
    gateway_factory_test.go:308: field 1 alias or duplicate reached resource
    gateway_factory_test.go:308: field 1 alias or duplicate reached resource
    gateway_factory_test.go:308: field 1 alias or duplicate reached resource
    gateway_factory_test.go:308: field 1 alias or duplicate reached resource
    gateway_factory_test.go:308: field 1 alias or duplicate reached resource
    gateway_factory_test.go:308: field 2 alias or duplicate reached resource
    gateway_factory_test.go:308: field 2 alias or duplicate reached resource
    gateway_factory_test.go:308: field 2 alias or duplicate reached resource
    gateway_factory_test.go:308: field 2 alias or duplicate reached resource
    gateway_factory_test.go:308: field 2 alias or duplicate reached resource
    gateway_factory_test.go:308: field 2 alias or duplicate reached resource
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.277s
--- FAIL: TestStreamRawWireByteBoundaries (0.00s)
    subscription_stream_test.go:145: CRLF subscription=false limit offset=-1 accepted excess wire bytes
    subscription_stream_test.go:145: mixed subscription=false limit offset=-1 accepted excess wire bytes
    subscription_stream_test.go:145: embedded CR subscription=false limit offset=-1 accepted excess wire bytes
    subscription_stream_test.go:145: CRLF subscription=true limit offset=-1 accepted excess wire bytes
    subscription_stream_test.go:145: mixed subscription=true limit offset=-1 accepted excess wire bytes
    subscription_stream_test.go:145: embedded CR subscription=true limit offset=-1 accepted excess wire bytes
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/translate	1.397s
FAIL
```

### GREEN
명령: go test -race ./internal/cli ./internal/gateway ./internal/gateway/translate -run "^TestGatewayFactory|^TestGatewayChild|^TestSubscriptionStream|^TestOpenAISubscription|^TestStream" -count=1 -timeout 45s -coverprofile=/tmp/factory-sse-repair.cover
exit=0
```text
ok  	github.com/modu-ai/moai-adk/internal/cli	4.207s	coverage: 6.1% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway	3.568s	coverage: 10.8% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.680s	coverage: 77.1% of statements
```

### 독립 probe
명령: go test -overlay=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-audit-znc5n4it/overlay.json -race ./internal/cli ./internal/gateway/translate -run "TestGatewayFactoryAudit|TestSubscriptionStreamAudit" -count=1 -timeout 30s -v
exit=0
```text
=== RUN   TestGatewayFactoryAuditCaseAlias
--- PASS: TestGatewayFactoryAuditCaseAlias (0.00s)
=== RUN   TestGatewayFactoryAuditFrozenDependenciesAndSubset
--- PASS: TestGatewayFactoryAuditFrozenDependenciesAndSubset (0.00s)
=== RUN   TestGatewayFactoryAuditConcurrentClose
--- PASS: TestGatewayFactoryAuditConcurrentClose (0.00s)
=== RUN   TestGatewayFactoryAuditRejectBeforeGPTStore
--- PASS: TestGatewayFactoryAuditRejectBeforeGPTStore (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.241s
=== RUN   TestSubscriptionStreamAuditCRLFByteLimit
    sse_independent_audit_test.go:5: wire_bytes=1447 limit=1446 terminal=false err=upstream SSE read: stream exceeds output byte limit
--- PASS: TestSubscriptionStreamAuditCRLFByteLimit (0.00s)
=== RUN   TestSubscriptionStreamAuditStrictStateAndJSON
=== RUN   TestSubscriptionStreamAuditStrictStateAndJSON/duplicate_terminal_field
=== RUN   TestSubscriptionStreamAuditStrictStateAndJSON/wrong_response_id
=== RUN   TestSubscriptionStreamAuditStrictStateAndJSON/output_object
--- PASS: TestSubscriptionStreamAuditStrictStateAndJSON (0.00s)
    --- PASS: TestSubscriptionStreamAuditStrictStateAndJSON/duplicate_terminal_field (0.00s)
    --- PASS: TestSubscriptionStreamAuditStrictStateAndJSON/wrong_response_id (0.00s)
    --- PASS: TestSubscriptionStreamAuditStrictStateAndJSON/output_object (0.00s)
=== RUN   TestSubscriptionStreamAuditWriterErrorClosesUpstream
--- PASS: TestSubscriptionStreamAuditWriterErrorClosesUpstream (0.00s)
=== RUN   TestSubscriptionStreamAuditCRLFBaselineControl
    sse_independent_audit_test.go:19: subscription=false wire_bytes=1598 limit=1597 terminal=false err=upstream SSE read: stream exceeds output byte limit
    sse_independent_audit_test.go:20: subscription=false LF negative control err=upstream SSE read: stream exceeds output byte limit
    sse_independent_audit_test.go:19: subscription=true wire_bytes=1447 limit=1446 terminal=false err=upstream SSE read: stream exceeds output byte limit
    sse_independent_audit_test.go:20: subscription=true LF negative control err=upstream SSE read: stream exceeds output byte limit
--- PASS: TestSubscriptionStreamAuditCRLFBaselineControl (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.434s
```

### 1 byte 분할 수신
명령: go test -race ./internal/gateway/translate -run "^TestStreamRawWireByteBoundaries$" -count=1 -timeout 30s
exit=0
```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.492s
```

go vet ./internal/cli ./internal/gateway ./internal/gateway/translate: exit=0, stdout/stderr 없음.
gofmt -l (수정 네 파일): exit=0, stdout 없음.
go tool cover -func=/tmp/factory-sse-repair.cover:
```text
newGatewayHandlerFactory 89.1%
ServeHTTP 100.0%
Close 100.0%
stream 92.8%
```

## Baseline-attribution
- source_session_id: 01a08e7b-6aa0-7361-ab7e-ea8da1f02228 (moai session current 실제 출력)
- WT: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
- git branch --show-current: WT-unified-gateway
- git rev-parse --short HEAD: 81c1d58f9
- git fetch -q origin main && git rev-list --count --left-right origin/main...HEAD: 0 2879
- 미커밋 작업을 측정했다. 수정한 파일은 아래 네 파일뿐이다. 별도 report를 추가했다.
- internal/cli/gateway_factory.go: 15917f81924580527f468cb2e36509c0b786996b59a0677f8f563876b3ddcc8e
- internal/cli/gateway_factory_test.go: f72c526235b1aa1d02e6ad181b6dffc957b3de4ace52b34e6277ea94aabbbf2b
- internal/gateway/translate/stream.go: a9b1699eb6f624f56e9a9158c38f01ad6f3a29c899e41129cc5653536aa636c5
- internal/gateway/translate/subscription_stream_test.go: cf695894c19046b2d6ff72c266064e3dbb2be50270b3d4d31a9989f1c9b2dbd4

## Gaps
독립 감사관의 새 판정은 아직 없다. 이 보고는 수리자의 재현 시험 결과다. LSP 전용 baseline은 수집하지 않았고 Go 컴파일/race/vet를 사용했다. 표시된 함수 커버리지를 저장소 전체 커버리지로 해석하지 않는다. repository-wide 시험 판정 소유자는 integration branch의 GitHub CI run이며 현재 PENDING (run ID 없음)이다. 실제 upstream/Windows/전체 launcher는 실행하지 않았다. 커밋·push·PR은 하지 않았다.

## Residual-risk
수신 byte 계수는 SSE parser가 소비한 메시지까지 검사한다. terminal 뒤 추가 바이트를 읽어 서비스 종료를 기다리는 동작은 추가하지 않았다. 기존 Scanner의 LF/CRLF 구문 범위를 넓히지 않았다. CR-only 이벤트 분리 지원을 주장하지 않는다. root의 newGatewayChildCommand(nil)은 유지했다. 출력 token 정책·Windows·receipt·SPEC는 수정하지 않았다.
