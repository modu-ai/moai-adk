# 일반 Messages·Responses 프로토콜 기능 검증

Functional Verdict: **FAIL — 공개 텍스트 순서 불일치 1건**.

검증 범위는 `translate.Request`, `Response`, `Stream`, OpenAI의 순수 stream body 변환, native Messages의 순수 body reader에서 공개 text 순서·tool call ID·정상 종료·잘린 EOF의 의미다. 인증·credential·보안 평가·외부 시스템 호출·전체 adapter HTTP 통합은 이번 작업에서 제외했다. 제품 파일을 수정하지 않았다.

## Claim

기존 변환기 시험과 race·vet는 통과했다. 별도로 만든 합성 SSE에서는 Stream이 성공하면서 같은 최종 응답의 Response 변환과 공개 텍스트 순서가 달라지는 현상을 재현했다. 서로 다른 output item 0·1을 먼저 열고 item 1의 text part를 먼저 완료하면, terminal output 배열은 `[first, second]`인데 클라이언트의 content block index로 재구성한 결과는 `[second, first]`였다.

단순 delta 도착 순서를 비교한 것이 아니다. 시험은 출력 SSE의 `content_block_delta.index`별로 text를 누적하고 index 순서로 조립한다. 비교 대상은 **동일한 terminal response JSON**을 `Response()`에 넣어 얻은 content 배열이다.

### PFR-F1 — Medium/P2 · blocking · 신뢰도 높음

- 위치: `internal/gateway/translate/stream.go:235-238`의 `s.nextBlock` 배정. item별 output_index보다 text part가 열린 순서에 따라 Anthropic block index가 정해진다.
- 입력: message output item 0과 1을 이 순서로 연다. 두 item은 공개 text만 가지며 각 id, content_index, 완료 text가 서로 일치한다. 이후 item 1의 content part 시작·delta·완료를 먼저 보내고 item 0을 보낸다. 최종 completed response의 output 배열은 item 0, item 1 순이다.
- 관측: `Stream()`은 오류 없이 종료하고 두 text의 content block index를 반대로 배정한다. 같은 final JSON의 `Response()`는 원래 output 배열 순서를 보존한다.
- 영향: 수락한 동일 공개 응답이 stream/nonstream 여부에 따라 다른 순서의 대화 기록이 된다. 이 합성 interleaving이 실제 공급자에서 발생했다는 주장은 하지 않는다. 현재 코드가 해당 입력을 성공으로 수락한다는 사실을 검증했다.
- 관련 계약: core design §4.3의 일반 공개 변환, REQ-MG-012 및 AC-MG-007·008의 스트림 index/완료 의미, 이번 검증 요청의 공개 text 순서 보존. 입력 history에 관한 design §4.2·AC-MG-009도 읽었으나 이를 출력 순서 검증의 대체 근거로 사용하지 않았다.
- 수리 요구: 유효한 일반 interleaving을 지원하면서 output item/content index에 맞는 최종 순서를 보존한다. 새 미지원으로 거절하거나 전체 stream을 terminal까지 묶어 증상을 숨기는 방식은 부모의 수리 방향에서 제외되어 있다. 먼저 열 수 있는 정상 prefix의 streaming 특성과 완료 결과의 순서를 함께 검증해야 한다.
- 재검증: 아래 원본 SSE와 동일 terminal JSON을 그대로 사용하고, 기존 tool 교차 흐름 및 같은 item의 여러 content part 대조군을 함께 실행한다.

### 확인한 정상 대조군

- Request의 최상위 system text 연결과 메시지 내 system 위치, 공개 history·완료 tool pair 순서와 ID, 요청별 이름 역매핑: 기존 단위시험을 읽고 변환 패키지 전체 시험을 실행했다.
- Responses의 item ID와 call ID가 다른 두 교차 tool stream: 기존 `TestStreamInterleavedToolsDistinctItemAndCallIDs`가 포함된 패키지 시험 통과. 각 index의 인수와 원래 이름 복원을 단언한다.
- 같은 message item의 content part 0·1을 순서대로 먼저 연 뒤 delta·완료를 1·0 순서로 보내는 독립 대조군: `[first, second]` 유지.
- Responses 공개 text stream을 terminal 이전의 모든 이벤트 경계에서 자르는 독립 시험: 모두 오류이며 message_stop 없음.
- OpenAI 순수 변환 body의 완전한 공개 text stream: text 한 번·정상 terminal 한 번. terminal 이전 모든 이벤트 경계에서 잘린 경우 오류.
- Native 순수 body reader의 완전한 공개 text stream: 원문과 byte-for-byte 동일. terminal 이전 모든 이벤트 경계에서 잘린 경우 오류·message_stop 없음.
- Native 두 tool의 교차 인수: 원문 SSE 보존, 깨진 JSON 인수 대조군은 정상 terminal 없이 오류.

## Evidence

모든 명령은 아래 WT에서 실행했다. 생성한 probe는 JSON 문자열과 `io.Reader`·`io.Pipe`만 사용하며 endpoint를 호출하거나 credential 메서드를 실행하지 않는다. 기존 root 시험은 인증 없이 순수 body를 검사하는 이름만 명시적으로 선택했다.

### 기존 일반 변환기 시험

```sh
GOCACHE=/tmp/gateway-protocol-review-cache go test ./internal/gateway/translate -count=1 -coverprofile=/tmp/gateway-protocol-review.cover -timeout 20s
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.410s	coverage: 92.6% of statements
```

```sh
GOCACHE=/tmp/gateway-protocol-review-cache go test -race ./internal/gateway/translate -count=1 -timeout 20s && GOCACHE=/tmp/gateway-protocol-review-cache go vet ./internal/gateway/translate
```

exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.711s
```

vet stdout/stderr는 빈 출력이었다. 이 기존 시험 통과를 독립 순서 반례의 통과로 해석하지 않았다.

### 독립 순서·EOF·같은 item 대조

```sh
GOCACHE=/tmp/gateway-protocol-review-cache go test -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/functional-review/overlay.json ./internal/gateway/translate -run '^TestFunctional' -count=1 -v -timeout 15s
```

exit 1:

```text
=== RUN   TestFunctionalStreamTextOrderMatchesCompleteResponse
    protocol_functional_test.go:26: successful stream text order=[second first]; complete response order=[first second]
--- FAIL: TestFunctionalStreamTextOrderMatchesCompleteResponse (0.00s)
=== RUN   TestFunctionalEveryTruncatedTextPrefixHasNoSuccessfulTerminal
    protocol_functional_test.go:31: all pre-terminal text prefixes failed without message_stop
--- PASS: TestFunctionalEveryTruncatedTextPrefixHasNoSuccessfulTerminal (0.00s)
=== RUN   TestFunctionalSameItemReverseDeltaArrivalPreservesContentIndices
    protocol_functional_test.go:43: same-item reverse delta/completion arrival preserves content block indices
--- PASS: TestFunctionalSameItemReverseDeltaArrivalPreservesContentIndices (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/translate	0.394s
FAIL
```

초기 probe는 delta 도착 순서만 비교했으므로 판정 방법을 수정했다. 위 최종 probe는 content block index로 재조립하며 같은 실패를 관측했다. 단순 arrival-order 비교의 초기 실행을 최종 판정 근거로 쓰지 않는다.

### Native 및 OpenAI 순수 body 대조

```sh
GOCACHE=/tmp/gateway-protocol-review-cache go test -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/functional-review/overlay.json ./internal/gateway -run '^TestFunctional|^TestAnthropicNativeInterleavedToolStreams$' -count=1 -v -timeout 15s
```

exit 0:

```text
=== RUN   TestAnthropicNativeInterleavedToolStreams
--- PASS: TestAnthropicNativeInterleavedToolStreams (0.00s)
=== RUN   TestFunctionalNativeBodyTextAndAllTruncatedPrefixes
    protocol_functional_body_test.go:11: native public text preserved byte-for-byte; every pre-terminal prefix failed without message_stop
--- PASS: TestFunctionalNativeBodyTextAndAllTruncatedPrefixes (0.00s)
=== RUN   TestFunctionalOpenAIStreamBodyTextAndAllTruncatedPrefixes
    protocol_functional_body_test.go:22: OpenAI body emitted text once and one normal terminal; every pre-terminal prefix returned an error
--- PASS: TestFunctionalOpenAIStreamBodyTextAndAllTruncatedPrefixes (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.399s
```

### 재현 자료

경로 기준:

`.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/functional-review/`

- `functional_test.go`: Responses 순서 반례, 모든 truncated text prefix, 같은 item 역순 delta 대조.
- `body_test.go`: OpenAI/native 순수 body 대조.
- `overlay.json`: 위 두 시험 파일만 해당 패키지에 주입한다. 제품 코드를 덮지 않는다.
- `interleaved-text.sse`: 실제 반례 시험이 생성한 원본 SSE, **2,833 bytes**, SHA256 `237cd4961fbd3d3ad73f22f692a7f5f8567976f907ca6ab6b41e13314e394941`.
- `interleaved-text-terminal.json`: 같은 시험의 정확한 terminal response, **395 bytes**, SHA256 `1567613753d4016cc31484de54c1f88857efa10ee9dcdd34616fda4e3defb8a1`.

terminal JSON을 디스크에서 읽은 원문:

```json
{"id":"resp-1","model":"gpt-5.6-sol","output":[{"content":[{"annotations":[],"text":"first","type":"output_text"}],"id":"item-0","role":"assistant","status":"completed","type":"message"},{"content":[{"annotations":[],"text":"second","type":"output_text"}],"id":"item-1","role":"assistant","status":"completed","type":"message"}],"status":"completed","usage":{"input_tokens":7,"output_tokens":3}}
```

## Baseline-attribution

```text
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
WT-unified-gateway
FUNCTIONAL_SOURCE_HASH_CHANGES 0 []
```

source_session_id: `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`.

일반 변환기 제품·기존 시험과 body를 포함한 9개 기준 파일의 SHA256을 위 증거 디렉터리 `baseline-sha256.txt`에 기록했다. probe 실행 뒤 변경은 0건이었다. 핵심 제품 SHA256:

```text
075cc45aea7051c310f08b5b2709271ae078f743aac2bb14fef212d136a7b59d internal/gateway/translate/request.go
c3b63250191182ab5bce876d869249b349904f60250729a60afac469e4fe84e0 internal/gateway/translate/response.go
fa09aa2991e032b48b85fddf969f208fcf87426d6a944574e9898f49dbb6a78f internal/gateway/translate/stream.go
```

`translation-verification.md`, `openai-adapter-verification.md`, `m6-local-adapters-verification.md`에서 일반 기능의 기존 주장과 명시적 미지원 경계를 읽었다. 이 보고서는 그 과거 측정값을 새 실행 결과로 쓰지 않는다. 실제 검증 범위는 위 새 명령과 출력이다.

## Gaps

- 인증·credential·보안 평가는 요청된 범위 밖이며 별도 판정을 하지 않았다.
- 외부 endpoint·Claude·GPT·GLM 서비스·실제 계정과 모델을 호출하지 않았다. 실제 공급자의 이 interleaving 발생 가능성은 측정하지 않았다.
- root adapter의 전체 HTTP 송수신은 실행하지 않았다. ordinary Responses 변환 함수와 순수 stream body reader를 검사했다.
- thinking·context_management·output_config·opaque reasoning은 현재 명시적 미지원 경계이며 이 기능 검증으로 완료 처리하지 않았다.
- 같은 item에서 content part 시작 자체가 1→0인 경우는 실행하지 않았다. 추가 대조는 **시작 0→1, delta·완료 1→0**이라는 조건이다.
- Windows, 전체 저장소 suite, 전체 provider 통합, 전체 core AC, 커밋·push·PR·병합은 미실행이다.

## Residual-risk

합성 JSON/SSE는 관측한 일반 형태만 확인한다. 공급자별 추가 이벤트·실제 네트워크 단절·대화 복구는 다른 증거가 필요하다. 순서 수리 뒤에도 조각이 늦게 도착하는 항목 때문에 앞부분 streaming이 불필요하게 멈추지 않는지 함께 확인해야 한다. 다른 기능의 기존 PASS가 이 순서 불일치를 가려서는 안 되며, 반대로 이 한 반례를 근거로 현재 검증하지 않은 인증·다른 provider 동작의 결함을 주장해서도 안 된다.
