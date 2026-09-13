# 구독 SSE 형식 보강 검증

## Claim

core SPEC 0.9.0 design §4.3의 구독 SSE 형식 보강을 구현했다. fixed AuthPKCE subscription의 streaming 응답만 Content-Type 부재를 허용한다. 명시적으로 잘못되거나 중복·대소문자 중복·malformed인 media와 지원하지 않는 Content-Encoding은 거부한다. 부재를 허용한 본문도 UTF-8, strict JSON, event 구조, item 상태, byte limit을 모두 검증하며 HTML/임의 본문은 성공 terminal을 만들지 못한다.

`ResponseContext.SubscriptionStream`은 일반 `Stream`과 같은 검증 상태기계를 사용한다. 차이는 terminal output이 없거나 빈 배열일 때 이미 완료 검증한 `output_item.done`의 final item을 output index 순서로 사용하는 것뿐이다. nonempty terminal output은 기존과 같이 순서·ID·type·전체 내용이 같아야 한다. null·잘못된 타입은 sparse output으로 취급하지 않는다. item만 완료되고 전체 terminal이 없거나, added/delta/done이 모순되면 성공을 합성하지 않는다.

일반 API-key Stream, native adapter, outgoing max_tokens, opaque receipt 및 root 활성화는 변경하지 않았다.

## Evidence

RED 명령:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test ./internal/gateway/translate ./internal/gateway -run 'TestSubscriptionStream|TestOpenAISubscriptionMedia|TestOpenAISubscriptionAbsent' -count=1 -timeout 30s
```

관측 출력:

```text
internal/gateway/translate/subscription_stream_test.go:44:16: c.SubscriptionStream undefined (type *ResponseContext has no field or method SubscriptionStream)
internal/gateway/translate/subscription_stream_test.go:78:16: c.SubscriptionStream undefined (type *ResponseContext has no field or method SubscriptionStream)
FAIL github.com/modu-ai/moai-adk/internal/gateway/translate [build failed]
internal/gateway/openai_subscription_test.go:26:13: undefined: openAIResponseMedia
FAIL github.com/modu-ai/moai-adk/internal/gateway [build failed]
FAIL
```

최종 GREEN 명령과 출력:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -race ./internal/gateway/translate ./internal/gateway -run 'TestSubscriptionStream|TestOpenAI|TestStream' -count=1 -timeout 45s -coverprofile=/tmp/gateway-subscription-coverage.out
ok  github.com/modu-ai/moai-adk/internal/gateway/translate 1.469s coverage: 77.0% of statements
ok  github.com/modu-ai/moai-adk/internal/gateway 1.710s coverage: 28.9% of statements
```

시험의 실제 판정:

- text 및 interleaved function call fixture의 빈/없는 terminal output을 처리한 출력이 원래 full terminal Stream 출력과 byte 단위로 같다.
- 같은 sparse fixture를 일반 Stream에 넣으면 실패한다.
- missing done, duplicate done, wrong identity/index, delta와 final text 불일치, no public output, null output, nonempty snapshot 불일치, reversed terminal order는 성공 message_stop을 만들지 못한다.
- 모든 text terminal 이전 절단점은 실패한다.
- UTF-8가 아닌 SSE comment, HTML, event/전체 byte limit 초과는 실패한다.
- blocked reader를 context 취소로 닫고 error를 반환한다.
- TLS mock subscription endpoint의 Content-Type 없는 ordinary SSE는 같은 Store credential 경계를 통과한 뒤 정상 변환된다. 같은 조건의 HTML은 stream read error이며 성공 terminal이 아니다. 각 경우 upstream 호출은 한 번이다.
- 기존 TestStream ordering/buffering/상태·종료 및 TestOpenAI API-key·Store ownership·취소 시험도 같은 명령에서 통과했다.

마지막 범위 커버리지 측정:

```text
openAIResponseMedia 100.0%
SubscriptionStream 100.0%
stream 92.3%
event 94.0%
doneItem 86.7%
layout 100.0%
```

`GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/gateway/translate ./internal/gateway`: exit 0, 출력 없음.
`gofmt -l` 수정 네 파일: 출력 없음.

## Baseline-attribution

WT moai-proxy-unified, HEAD `81c1d58f9`. 변경 파일:

- `internal/gateway/openai.go`
- `internal/gateway/openai_subscription_test.go` (신규)
- `internal/gateway/translate/stream.go`
- `internal/gateway/translate/subscription_stream_test.go` (신규)

작업 전 `auth-responses-runtime-observation.md`, `protocol-functional-review-iter2.md`, core design §4.3을 읽었다. 이번 시험은 기존 synthetic ordinary text/function fixture와 모의 TLS/Store만 사용했다. 실제 token·opaque 응답을 fixture로 복제하지 않았고 provider/client를 실행하지 않았다.

## Gaps

- raw reasoning 포함 실제 구독 응답의 전체 변환·opaque receipt 구현을 완료한 것은 아니다.
- outgoing max_output_tokens 거절과 출력 상한 계약은 별도 결정 대상이며 이번에 매핑을 삭제하지 않았다.
- nonstream 구독 지원을 새로 선언하지 않는다. 기존 명시 application/json 경로의 동작을 보존했다.
- 새로운 변경분의 독립 기능 감사는 아직 받지 않았다. PFR iter2 PASS는 이전 ordering 수리 범위의 판정이다.
- integration branch CI의 repository-wide 판정 및 native Windows 실행은 PENDING이다.

## Residual-risk

streaming은 전체 응답을 메모리에 모아 완료 후 반환하는 방식이 아니다. 앞부분이 유효하면 일부 delta가 보인 뒤 잘못된 terminal·EOF 때문에 오류로 끝날 수 있다. 성공 message_stop을 내보내지 않는 것이 오류 계약이며, HTTP header의 200만으로 전체 스트림 성공을 판정해서는 안 된다. 이 변경은 기존 PFR의 terminal 이전 정상 delta 가시성을 유지한다.

Content-Type 부재 예외는 AuthPKCE 행의 fixed subscription endpoint streaming에만 적용된다. 새 일반 URL·다른 provider 우회·runtime 기능 스위치를 만들지 않았다. 새 factory와 root의 실제 연결은 여전히 별도 게이트다.
