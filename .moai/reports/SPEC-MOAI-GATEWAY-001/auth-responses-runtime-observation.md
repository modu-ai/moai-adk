# 구독 Responses 실측 — 네 모델 함수 왕복과 Luna opaque 후속 요청

## Claim

19시 이후 부모가 새 로그인으로 만든 전용 MoAI Store를 사용하여 고정 구독 endpoint를 직접 관측했다. 로그인·refresh·logout·원시 자격 파일 읽기/출력은 수행하지 않았다. `Store.SendAuthorized`로만 전송했고 공개 유료 API나 환경 자격으로 fallback하지 않았다. **이는 직접 Responses 프로토콜 관측이며 Claude 통합·제품 활성화 PASS가 아니다.**

| 항목 | 직접 관측 | 판정 경계 |
|---|---|---|
| GPT 네 ID 텍스트 | 각 HTTP200, bounded SSE parser에서 response.completed 확인 | Content-Type 헤더가 모두 없어 기존 제품 media acceptance는 false 유지 |
| 출력 상한 | Astra의 max_output_tokens:64 요청은 HTTP400, 알려진 unsupported_parameter 분류 | Claude max_tokens를 제품에서 조용히 제거할 수 있다는 뜻 아님 |
| 네 모델 합성 함수 | 각 high effort로 function_call 수신, 같은 call_id에 고정 함수 결과를 넣은 다음 요청도 HTTP200/SSE 완료 | 외부 도구 실행 없음; 첫 응답에 opaque는 0개 |
| terminal output | tool/text item은 output_item.done으로 도착하지만 response.completed.output은 [] | done 이벤트로 ordered item을 재구성해야 하며 제품 계약 변경은 별도 감사 필요 |
| Luna의 도구 없는 opaque 응답 | 함수 결과 뒤 응답이 reasoning→message 순서, encrypted_content 1400바이트 | 실제 no-tool-call assistant 응답에도 opaque가 존재한다는 관측 |
| Luna 다음 턴 | 그 reasoning와 message를 순서대로 재전송하고 새 합성 user 메시지를 추가하니 HTTP200/SSE 완료 | opaque 제거 mutant는 수행하지 않아 서버가 그 항목을 필수로 요구하는지는 미판정 |

### 정책 지원 상태

- `reasoning.effort: low`: 네 모델 직접 텍스트 요청 수용을 관측했다.
- `reasoning.effort: high`: 네 모델 직접 함수 요청·결과 후속 요청 수용을 관측했다.
- `include:["reasoning.encrypted_content"]`: 요청했다. 실제 opaque 반환과 다음 턴 보존 수용은 Luna에서만 양성 관측했다.
- Claude native `thinking:adaptive` → OpenAI effort 매핑, `clear_thinking keep all`의 정책 해석은 이번 직접 요청에서 시험하지 않았다.
- title용 JSON schema output format은 이번 직접 endpoint 요청에서 시험하지 않았다. 앞선 picker raw에는 실제 Claude가 해당 필드를 넣었지만 그것은 로컬 mock 관측이다.
- `max_output_tokens`의 대체 협상 상한은 근거를 확보하지 못했다. byte cap·deadline·추정 token은 제공자 출력 상한의 대체가 아니다.

## Evidence

작업 디렉터리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`.

도구: `/tmp/gateway-auth-live-preflight`. 원시 응답·후속 요청은 ignored `auth-live/responses/`에 0600 파일로만 저장했다. 인증 헤더·쿠키·자격 파일은 캡처하지 않았다. 아래 파일명은 모두 다음 디렉터리 기준이다.

`.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-live/`

### 호출 순서와 실제 결과

```text
$ /tmp/gateway-auth-live-preflight -mode probe -shape codex -model gpt-6-astra
[exit 1: HTTP200, Content-Type 값 0개, body7017B, SSE parse completed=true, production completed=false]

$ /tmp/gateway-auth-live-preflight -mode probe -shape codex -diagnostic
[exit 0: 네 모델을 각 1회 관측; 네 건 모두 HTTP200, Content-Type 값 0개, SSE parse completed=true]

$ /tmp/gateway-auth-live-preflight -mode probe -shape bounded -model gpt-6-astra
[exit 1: HTTP400, application/json, body53B, known_error_class=unsupported_parameter, error_parameter=max_output_tokens]

$ /tmp/gateway-auth-live-preflight -mode tool-probe
[exit 1: Astra 첫 응답은 HTTP200/SSE 완료였지만 completed.output=[]를 읽는 초기 하네스가 중단]

$ /tmp/gateway-auth-live-preflight -mode tool-probe -resume-first /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-live/responses/gpt-6-astra-tool-first-1208924732.raw
[Astra 첫 응답은 파일에서 이어감, 새 호출 없음. Astra 후속1 + Sol2 + Terra2 + Luna2 요청 모두 HTTP200/SSE 완료]
[exit 1: 하네스 분기 실수로 이어진 추가 Astra bounded1 요청이 HTTP400에서 중단]

$ /tmp/gateway-auth-live-preflight -mode opaque-followup
[exit 0: Luna 한 번만 추가 요청, HTTP200/SSE 완료, 이전 opaque 항목 보존]
```

하네스 실수로 **예정 밖 Astra bounded 요청 1건**이 추가되었다. tool-probe case에 일반 probe 루프가 잘못 포함된 원인이며 이를 부모에게 즉시 보고했다. 분기를 고친 뒤 `TestProbeModesHaveNoCrossModeLiveCall`로 두 모드의 호출 경로가 섞이지 않음을 확인했다. 오류를 숨기거나 전체 tool 시험을 다시 돌리지 않았다. 이 위임 중 실제 응답은 총 **16건**이며 manifest에 해시·크기·권한을 기록했다. 그 중 위 예정 밖 400도 포함된다.

Content-Type 원인 진단의 실제 출력 필드:

```json
{"content_type_get_empty":true,"content_type_value_count":0,"content_type_parse_valid":false,"content_encoding":"","go_uncompressed":false,"body_class":"sse_prefix","utf8":true,"sse_parse_completed":true}
```

Header.Get과 모든 header key의 case-insensitive 검색이 모두 부재를 확인했다. 대소문자나 압축 때문에 놓친 것이 아니다. 명시 diagnostic 모드는 header acceptance를 true로 바꾸지 않고 body framing 관측만 계속한다. status가 200이 아니거나 SSE 완료를 검증하지 못하면 중단한다.

상한 오류는 원문을 출력하지 않고 구조와 알려진 parameter만 분류했다.

```text
response_object_keys: ['detail']
detail_is_string: True
error_type_present: False
error_code_present: False
HTTP status: 400
known_error_class: unsupported_parameter
error_parameter: max_output_tokens
body bytes: 53
body SHA256: 8cdeda4a96a2d259baa0f7d984f190bea4bcf1baf2b461cb0969355aa9d6eb55
```

따라서 OpenAI 표준 error.type/code가 반환됐다고 주장하지 않는다. 반환된 detail의 알려진 unsupported parameter 표현만 검사했다. 근거 파일은 `responses/gpt-6-astra-bounded-158048658.raw`다.

### 네 모델 함수 왕복

함수는 `moai_preflight_echo`, strict object schema(value string), parallel_tool_calls false, high effort, store false, stream true를 사용했다. 첫 요청은 지정 함수 호출을 요구하고, 후속은 동일 call_id에 고정 합성 JSON 결과를 넣으며 tool_choice none을 사용했다. 실제 외부 도구는 실행하지 않았다.

| 모델 | 첫 응답 bytes | 후속 응답 bytes | 첫 output item | 후속 output item | 후속 SSE 완료 |
|---|---:|---:|---|---|---|
| gpt-6-astra | 7643 | 8287 | function_call | message | true |
| gpt-5.6-sol | 7643 | 8287 | function_call | message | true |
| gpt-5.6-terra | 7649 | 8293 | function_call | message | true |
| gpt-5.6-luna | 7646 | 11629 | function_call | reasoning, message | true |

각 tool-first에는 output_index0의 function_call added/done 및 arguments delta/done이 있었고 completed.output은 빈 배열이었다. `terminalOutput` 판독은 strict JSON, 중복/빠진 index, start/done identity·call_id 불일치, 중복 done, 실패/누락 terminal, terminal 후 이벤트를 거절한다. 출력은 done index 순서로 재구성한다. terminal에 output이 실제로 있다면 done item과도 일치해야 한다. 이 변경은 ignored 관측 도구에만 적용했다.

### Luna opaque 보존 후속

실제 세 번째 요청의 최종 구조화 결과:

```json
{
  "result":"OPAQUE_FOLLOWUP_OBSERVATION",
  "model":"gpt-5.6-luna",
  "opaque_items":1,
  "ordered_copied_items":[
    {"type":"reasoning","compact_item_bytes":1530,"compact_item_sha256":"dbad8c689626ad50670adf129d1ffa9cf65faafa7fd906a6f0007fbf576c08f9","opaque_bytes":1400,"opaque_sha256":"61556e343f349221dc5a6edfc23d6d15ef03b1b08422178e36b9dd0442875207"},
    {"type":"message","compact_item_bytes":231,"compact_item_sha256":"a7d86f56671296d45cd99c3553669de4f734c4b17c5754b29c4c379b83914678"}
  ],
  "wire_compact_item_bytes_equal":true,
  "followup_parsed_completion":true,
  "production_media_accepted":false,
  "request_sha256":"3d6ae8372d5a671bb2e3d62bce1638229dcaf9afa25a48732c02591edd56be9c"
}
```

기존 output을 compact JSON으로 만든 bytes와 실제 송신 input item bytes를 비교했다. opaque 문자열 자체는 바꾸지 않았으며 reasoning→message 순서를 유지했다.

수신 응답은 HTTP200, 12152B, SSE 완료였다. SHA256은 `4cf0a7fcd8dca8eb2ca848e9c53d5f83437a96ec12c7b0cfd16279c95795a737`이다. raw 파일은 `responses/gpt-5.6-luna-opaque-followup-1462378748.raw`, request는 `responses/gpt-5.6-luna-opaque-followup-request-790052897.json`이다. 보고서에는 암호화 내용 자체를 싣지 않는다.

### 공식 Codex 소스 대조

공개 checkout `/tmp/openai-codex-audit.zbvNXB`, HEAD `5a9eb14`에서 직접 읽었다.

- `codex-api/src/endpoint/responses.rs:165–187`: Accept text/event-stream을 설정한 stream 응답을 spawn_response_stream으로 넘긴다.
- `codex-api/src/sse/responses.rs:575`: body byte stream을 `.eventsource()`로 읽는다. 이 두 구간에는 Content-Type 필수 판정이 없다. 전체 transport 어디에도 없다는 전수 주장은 하지 않는다.
- 같은 파일 357–361행: output_item.done을 별도의 OutputItemDone 이벤트로 처리한다.
- 같은 파일 483–499행: response.completed는 ID·usage 등을 가진 Completed 이벤트로 처리한다. 이 구간은 completed.output으로 이전 done item 전체를 재구성하지 않는다.
- `codex-api/src/common.rs:259`의 ResponsesApiRequest에는 max_output_tokens 필드가 없다. 대체 출력 상한을 보장하는 소스는 확보하지 못했다.

## Baseline-attribution

```text
$ moai session current
hint: session.id not available from the runtime; emitted canonical fallback. Run 'moai session doctor' to diagnose.
source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>
$ git rev-parse --short HEAD
81c1d58f9
```

실행자는 `/root/gateway_translation`이다. 부모 source session `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`을 증거 경로에 사용했다. 이 child의 별도 UUID는 runtime에서 얻지 못했으므로 부모 UUID를 child UUID라고 주장하지 않는다.

마지막 코드 검사:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-live -count=1
ok  	github.com/modu-ai/moai-adk/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-live	0.513s
$ GOCACHE=/tmp/gateway-translation-cache go vet ./.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-live
[exit 0, 출력 없음]
$ [response manifest 실측]
raw_response_count 16 all_mode0600 True
```

`response-manifest.json`에 16개 raw 응답의 실제 SHA256·길이·권한이 있다. 제품 AUTH/translator/CLI/SPEC 파일은 수정하지 않았다. ignored 도구와 관측 파일, 본 보고서만 작성했다. Store를 삭제하거나 로그아웃하지 않았다.

## Gaps

출력 상한의 호환 계약은 해결되지 않았다. title JSON schema·adaptive 변환·keep-all 정책·opaque 제거 음성 시험·Claude와 실제 upstream을 잇는 통합은 미검증이다. 실제 encrypted item 반환과 다음 턴 수용의 양성 근거는 Luna 한 모델뿐이다. no-tool reasoning의 필수 재전송 여부를 삭제 시험 없이 단정하지 않는다.

대소문자·압축 진단과 raw fixture 재생으로 media header 부재와 빈 terminal output을 관측했지만, 제품이 이를 어떻게 허용할지는 별도 계약 결정과 감사가 필요하다. 실행 환경의 네트워크 중계 전체가 헤더를 변경하지 않았다는 packet-level 증거는 없다. 관측은 Store.SendAuthorized가 받은 HTTP 응답 경계에 귀속한다.

## Residual-risk

상한을 생략한 codex 형태는 이 명시적인 직접 프로토콜 시험에만 사용했다. 사용자에게 약속한 Claude max_tokens 보존을 조용히 제거하는 제품 변경의 근거가 아니다. 앞으로도 parsed SSE completion과 production media acceptance를 분리해야 한다. 하네스 분기 실수로 발생한 추가 400 요청을 포함한 호출 수를 숨기지 않았으며, 수정 후 다른 모델을 다시 호출하지 않았다.
