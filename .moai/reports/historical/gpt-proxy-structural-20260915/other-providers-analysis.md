# claude-code-proxy 기타 제공자 전수 구조 분석

## 결론

MoAI가 흡수할 대상은 다른 제공자의 인증 우회나 모델 목록이 아니라 **증분 스트리밍, 종료 무결성, 도구 식별자, 오류 분류, 재현 테스트의 계약**이다. 조사한 네 제공자는 구현 수준과 의미 보존 방식이 상당히 다르다. 특히 Cursor의 XML 기반 도구 재개는 Codex AppServer의 실제 도구 응답 왕복을 대체하지 못한다. 현재 GPT 네 모델 allowlist와 AppServer 구조를 유지해야 한다.

이 보고서는 소스 읽기 분석이다. Rust 테스트 실행, 공급자 실계정 호출, 구독 약관 승인, 실제 성능 벤치마크를 수행한 보고서가 아니다. Rust 소스를 참고한다는 사실은 MoAI를 Rust로 재개발한다는 뜻이 아니다.

## 측정 기준과 증거

분석 체크아웃: `/tmp/moai-proxy-structural.IArJmM/source`.

실행 명령과 원문 결과:

```text
git rev-parse HEAD
b642a6ee1a5fff8ebc9bedfc4b438ddfa3ae8f8e

git status --short
[출력 없음, exit 0]

find src/providers/kimi src/providers/grok src/providers/opencode src/providers/cursor -type f -exec wc -l {} +
   20571 total
```

위 네 디렉터리의 52파일 20,571줄과 `translate_shared.rs` 232줄, `mod.rs` 6줄을 합해 **54파일 20,809줄을 첫 줄부터 EOF까지 읽었다.** 파일별 읽기 장부는 마지막 절에 있다. 큰 파일은 연속 `sed -n` 구간으로 나누었고, 출력이 잘린 Kimi reducer 테스트 구간은 별도 재독했다. 후속 `rg -n`은 인용 줄 번호 확인에만 사용했다.

## 제공자 구조 비교

| 구분 | 인증·연결 | 스트리밍 | 대화·도구 계약 | MoAI 적용 판단 |
|---|---|---|---|---|
| Kimi | 자체 OAuth/device 저장소, Kimi CLI 형태 헤더, blocking HTTP | upstream 전체 bytes 수집 후 BufferedSse | 전체 대화를 Chat Completions로 재작성 | 변환 테스트는 참고. 버퍼링·묵시적 기본 모델은 제외 |
| Grok | 자체 OAuth PKCE/device, CLI 전용 Responses HTTP | 공유 async client와 실제 증분 SSE | 함수 call/result 검증, reducer/renderer 분리 | 스트림 무결성·진단·refresh 중복 억제 계약 우선 참고 |
| OpenCode Go | 별도 API key, 모델별 Messages/Chat/Responses endpoint | 세 프로토콜 모두 LiveSse 경로 | 프로토콜별 변환, 응답별 도구 ID namespace | 프로토콜 capability 분리와 테스트 방식 참고. Codex 구독 인증 대체 아님 |
| Cursor | 자체 로그인·토큰 저장, Connect HTTP/2/protobuf | 전체 응답 수집 후 BufferedSse | 전체 대화를 XML 텍스트로 평탄화, XML tool_use 복구 및 저장 이벤트 재출력 | 일반 GPT 세션·모든 Claude 기능 보장 설계로 채택 불가 |

### Kimi: 빠른 응답에 그대로 쓰면 안 되는 버퍼 구조

`kimi/client.rs:188`은 성공 응답에서 `resp.bytes()`로 전부 수집한다. `kimi/mod.rs:279`는 이 결과를 `GenerationBody::BufferedSse`로 반환한다. SSE 문법을 만들더라도 upstream 생성 도중 사용자에게 텍스트를 전달하는 구조는 아니다. 이는 소스의 데이터 흐름 관찰이며, 몇 초 느리다는 실측은 하지 않았다.

`translate/request.rs`는 전체 이력을 Chat Completions messages로 변환한다. K3는 system을 user 앞에 붙이고, 다른 모델은 system role을 사용한다. 도구 결과 주변 일반 텍스트·이미지와 error 표기를 변환하는 테스트가 있다. 하지만 MoAI의 native thread, 승인된 입력 경계, 취소 후 이미 전달한 결과 중복 방지와는 상태 모델이 다르다. 이 변환기를 복사해 AppServer scope 검사를 없애는 해결은 성립하지 않는다.

주의할 소스 계약:

- 알 수 없는 모델은 기본 모델로 해석하는 테스트가 있다. MoAI의 정확한 네 모델 allowlist에 적용하지 않는다.
- reducer는 malformed JSON을 건너뛰고 종료 이벤트를 합성한다. MoAI는 실패·중단·정상 완료를 구분해야 하므로 복사하지 않는다.
- thinking signature는 자체 base64 표식이다. 인증·소유권 증거가 아니다.
- count_tokens는 단어·구두점 기반 추정과 이미지 고정값이다. 공급자 실측 TPS나 정확한 컨텍스트 사용량으로 취급하지 않는다.
- `client.rs:63`은 `Ok(response)`의 401에서 refresh하지만 `client.rs:163`은 401/403을 Err로 반환한다. 동일 함수 흐름에서 분기 형태가 맞지 않는 정적 관찰이다. mock 실행으로 refresh 호출 수를 검증하지 않았으므로 실행 확인된 결함으로 단정하지 않는다.

### Grok: 재사용 가치가 높은 엄격한 스트림 경계

`grok/client.rs`는 공유 async client, redirect 금지, connect timeout, TCP_NODELAY/HTTP2 keepalive를 구성한다. `into_stream`은 chunk를 즉시 전달하고 non-stream 수집에는 8MiB 상한을 둔다. 인증 401은 한 번 refresh 후 재시도하며, `auth/manager.rs`는 mutex 안에서 최신 저장 토큰을 재확인하여 동시 stale-401이 같은 refresh를 반복하지 않도록 한다. 이미 제공자가 관리하는 Codex 인증에는 이 구현을 이식하지 말고, **동시 요청의 중복 재시도를 막는 일반 계약만** 참고한다.

`translate/stream.rs`의 SseDecoder는 CR/LF 경계, 분할 UTF-8/프레임, 주석, data 없는 프레임, 1MiB 프레임 상한을 처리한다. decoder 완료와 reducer terminal 완료를 각각 확인한다. `translate/reducer.rs`는 다음을 명시적으로 거부한다.

- terminal 이후 추가 이벤트
- 순서를 벗어난 함수 인자 delta
- 중복 함수 ID, 완료되지 않은 함수, 인자 delta와 done 본문 불일치
- 알 수 없는 이벤트 또는 malformed 이벤트

함수 인자 delta 1MiB와 미완료 함수 128개 상한도 있다. 단, done-only 인자 전달 등 모든 분기에서 동일 상한을 보장하는지는 별도 실행 검증하지 않았다.

`grok/mod.rs`의 stream state는 transport/decoder/json/reducer/render를 구분하고, 비민감 stage/kind 및 bytes/chunks를 기록한다. downstream drop도 abandonment로 분리한다. 이는 MoAI의 공개 `appserver_scope_mismatch`를 유지하면서 내부 거부 지점만 안전하게 남기는 방향과 잘 맞는다.

usage에서 **누락과 실제 0을 구분**하는 점도 참고할 만하다. `input_tokens: Option<u64>`를 유지하며 provider가 미제공하면 terminal에 0을 덮어쓰지 않는다. 단 초기 추정값이 실측값이 되는 것은 아니다.

이식하면 안 되는 Grok 특화 정책:

- 일부 hosted-search 모드는 자연어 의도를 보고 도구 집합을 바꾸거나 검색만 강제한다. Claude Code의 사용자 도구·권한을 보존해야 하는 MoAI 공통 경로에 넣지 않는다.
- 이미지 omit/reattach/inline 정책은 URL 이미지 제외, 최근 4개 제한, 크기 제한에 따라 내용을 생략할 수 있다. 시각 작업을 온전히 처리했다는 보장과 다르다.
- `max_uses`, 일부 cache/context/eager-streaming 필드는 수용 후 upstream에 전달하지 않는 테스트가 있다. 요청 수용은 해당 기능 구현의 증거가 아니다.
- 함수 result 참조 검사는 stateless 이력 내 검증이다. durable AppServer 세션 소유권 검사를 대체하지 않는다.

### OpenCode Go: 인증과 wire protocol을 구분해야 한다

`opencode/client.rs`는 API key를 요구한다. Messages는 `x-api-key`와 anthropic-version, Chat/Responses는 Bearer를 사용한다. `x-opencode-session`은 유효한 요청 session 값 또는 새 UUID다. 이 소스는 GPT 구독으로 Claude Code에 로그인하는 계약이 아니다.

`model.rs`는 모델별 wire endpoint를 분리한다. protocol-native 경로를 사용할 수 있으면 불필요한 변환을 줄이는 설계는 참고할 수 있다. 다만 현재 36개 모델 catalog를 MoAI에 수입하거나 네 GPT 모델 제한을 풀 이유는 없다.

좋은 참고점:

- `chat.rs:960`의 도구 ID는 응답 ID와 upstream slot을 조합한다. upstream이 매번 `Agent_0` 같은 ID를 재사용해도 부모·자식 대화가 충돌하지 않도록 하는 테스트가 있다. MoAI에서는 durable receipt와 연결되는 기존 ID 계약을 먼저 확인해야 한다.
- Chat 변환은 병렬 함수의 tool result를 연속 배치하고 이미지·일반 사용자 내용을 뒤에 붙이는 테스트가 있다. 순서 변경의 의미까지 승인된 프로토콜 범위에서만 적용한다.
- `responses.rs:115` 이후는 completion을 보류하고 EOF까지 잘못된 tail을 검사한다. 정상 텍스트는 먼저 내보내되 최종 성공을 성급히 만들지 않는 구조다. 모든 chunk split 위치를 반복하는 테스트가 있다.
- max_output_tokens 종료와 content_filter 종료를 분리한다. 제한 종료를 일반 전송 실패로 취급하거나 필터 실패를 정상 max_tokens로 표시하면 안 된다.
- 401/403/402/429를 구분하고 Retry-After를 보존한다. 무조건 502로 감싸는 것보다 원인과 재시도 여부가 명확하다.

주의점:

- Chat와 Messages의 live 경로는 terminal 발견 시 해당 chunk 이후 upstream을 더 읽지 않는 반면 Responses는 EOF까지 읽는다. terminal-tail 정책이 프로토콜마다 다르다. 모든 경로가 동일한 종료 검증을 한다고 말할 수 없다.
- 공통 normalize_content의 알 수 없는 블록 생략, 문자열 연결, 기본값 적용은 완전한 의미 보존 보장이 아니다.
- Chat usage는 누락을 0으로 만드는 부분이 있다. Grok의 누락/0 계약과 동일하지 않다.
- upstream 오류 message를 공개 오류로 전달하는 경로가 있다. MoAI는 자격 정보나 원문이 섞일 수 있는 detail 대신 공개 고정 코드와 비민감 내부 진단을 유지한다.

### Cursor: 실제 native 도구 왕복과 다른 구조

`cursor/client.rs`는 직접 Connect/protobuf 프레임과 CLI 형태 헤더를 만들고 새 conversation UUID를 생성한다. 프레임 사이 1500/800/400ms 지연과 5초 heartbeat를 넣는다. 응답은 전체 수집한 뒤 반환한다. 첫 byte 전 60초, 수신 후 5초 inactivity 분기가 있으나 이를 모델 속도의 정상 기준으로 사용할 수 없다. request sender의 취소·초기 send 실패 cleanup도 별도 검증이 필요하다.

`request.rs`는 전체 대화를 XML 태그가 들어 있는 단일 prompt 문자열로 만든다. base64 이미지만 별도 선택 이미지로 넣고 URL 이미지는 텍스트 표기다. 이는 원래 role/tool/input 구조를 그대로 유지하는 AppServer adapter와 다르다.

핵심 소스 흐름은 다음과 같다.

1. 응답 텍스트의 `<tool_use>`를 `tool_use_xml.rs`로 복구한다.
2. `tool_bridge.rs`는 Read/Write/Bash만 pending 형태로 저장한다. registry key는 session 문자열이며 durable receipt/계정/프로젝트/모델 복합 소유권 검사가 아니다.
3. `resume_cursor_tool_bridge`는 도구 결과에서 `result_messages`를 만들지만 재개 SSE는 **이미 저장된 remaining_events**로 생성한다.
4. `cursor/mod.rs:91`은 `_result_messages`를 버리고, `mod.rs:263`도 첫 tuple 값을 버린다. 재개 함수에는 네트워크 클라이언트나 살아 있는 upstream 채널이 없다.

따라서 이 resume 경로는 새 Read/Bash 결과를 upstream 모델에 전달하여 다음 추론을 받는 구현과 같지 않다. 이 소스 흐름을 읽은 결과이며, 실제 서비스에 호출하여 결과 변화 여부를 실증한 것은 아니다. MoAI의 사용자 요구인 모든 Claude 도구 기능·정확한 결과 반영을 만족하는 대체 설계로 추천하지 않는다.

추가 정적 주의점: 재개 parser는 allowed_tool_names=None으로 재생성하고, 두 번째 pause 저장 상태는 pending/remaining을 채우지 않는다. 첫 recovery 루프에서는 같은 회수 묶음의 두 번째 ToolUse가 보관되지 않는 분기가 있다. 이들은 mock 재현이 필요한 확인 후보이며 이번 보고서는 실제 실행 결함 판정으로 포장하지 않는다.

`response.rs`는 decode 실패 프레임을 건너뛰는 분기가 있고, `sse.rs`는 upstream이 비어 있어도 정상 start/stop을 합성한다. 관련 테스트도 그 계약을 기대한다. input 사용량을 응답 텍스트 길이로 보정하고 cache를 0으로 만드는 부분도 있으므로 MoAI의 성능 계측과 성공 판정에는 가져오지 않는다.

## MoAI 적용 우선순위

1. **High — 공개 스트림과 내부 상태의 완료 경계.** 텍스트는 생성 즉시 내보내되 native turn 완료·도구 결과 전달·취소 확인을 구분한다. transport 끊김을 end_turn으로 합성하지 않는다.
2. **High — 실제 lifecycle 회귀 테스트.** tool result 전달 → cancel ack → 다음 사용자 입력에서 결과가 한 번만 전달되는지 검사한다. summary는 실제 실패 body/진단 지점을 확보한 후 별도 ownership으로 고친다.
3. **High — 비민감 stage 진단.** 허용된 stage/kind, phase, 도구/결과 수, 첫 이벤트 시간, 최종 완료 시간만 기록한다. 전체 프롬프트·키·토큰은 기록하지 않는다.
4. **Medium — 분할 경계 테스트.** byte split 전수, interleaved 도구, 완료 뒤 잘못된 이벤트, downstream 취소, 실제 0과 누락 usage, 병렬 ID 충돌을 합성 fixture로 검증한다.
5. **Medium — retry 단일 책임.** provider capacity/429와 입력 오류400, 승인 실패, native transport 종료를 구분한다. 이미 도구를 실행한 요청을 자동 replay하지 않는다. 네트워크 retry가 가능하다는 사실만으로 부작용이 있는 turn을 재실행하지 않는다.

인증은 공식 Codex AppServer/로그인 경계를 유지한다. OAuth 헤더·토큰 저장소·비공개 서비스 endpoint 구현을 가져오는 것은 이 보고서의 권고가 아니다. 소스가 열려 있고 연결 코드가 존재한다는 사실은 ToS 허용이나 공식 지원을 증명하지 않는다. 약관 판정은 별도 공식 자료 분석이 필요하다.

## 검증 공백과 잔여 위험

- 이 소스의 Rust 단위 테스트는 **읽었지만 실행하지 않았다**. 실제 공급자 호출과 구독 로그인도 실행하지 않았다.
- 소스 주석의 과거 live success는 이 세션의 측정값이 아니다.
- top-level router/auth storage/traffic redaction/codex translator 내부는 다른 조사 범위다. 이 보고서는 외부 호출을 따라가지 않았으므로 전체 저장소 보안 판정이 아니다.
- MoAI 기존 테스트 PASS를 이 proxy 코드의 PASS로 전용하지 않는다. CI의 저장소 전체 판정은 별도이며 여기서는 PENDING/미측정이다.
- 앞선 두 프로젝트 로그와 proxy의 구조적 차이를 확인했지만 이 네 제공자 코드가 MoAI의 400/502 원인이라는 증거는 없다. 이들은 회귀 방지 설계 참고자료다.

## 파일별 EOF 읽기 장부

아래 표의 각 값은 `상대경로: 마지막 줄`이며 모든 항목에서 **1–마지막 줄 전부 읽음**, 미독 구간 0이다. 기준 디렉터리는 `src/providers/`다.

| 범위 | 파일과 마지막 줄 |
|---|---|
| 공통 | `mod.rs:6`, `translate_shared.rs:232` |
| Kimi 기본 | `kimi/client.rs:203`, `kimi/count_tokens.rs:231`, `kimi/mod.rs:383` |
| Kimi auth | `kimi/auth/constants.rs:13`, `device_id.rs:54`, `headers.rs:53`, `jwt.rs:53`, `login.rs:125`, `manager.rs:230`, `mod.rs:7`, `token_store.rs:61` |
| Kimi translate | `kimi/translate/accumulate.rs:211`, `mod.rs:6`, `model_allowlist.rs:129`, `reducer.rs:511`, `request.rs:1173`, `signature.rs:18`, `stream.rs:383` |
| Grok 기본 | `grok/client.rs:319`, `grok/count_tokens.rs:138`, `grok/mod.rs:1038` |
| Grok auth | `grok/auth/device.rs:408`, `login.rs:353`, `manager.rs:180`, `mod.rs:5`, `pkce.rs:52`, `token_store.rs:45` |
| Grok translate | `grok/translate/accumulate.rs:305`, `mod.rs:6`, `model_allowlist.rs:11`, `reducer.rs:562`, `request.rs:2877`, `search_text.rs:46`, `stream.rs:699` |
| OpenCode | `opencode/chat.rs:1701`, `client.rs:368`, `messages.rs:314`, `mod.rs:950`, `model.rs:278`, `responses.rs:494` |
| Cursor 전반 | `cursor/auth.rs:476`, `client.rs:452`, `connect.rs:373`, `mod.rs:509`, `model.rs:176`, `proto.rs:186` |
| Cursor 변환·도구 | `cursor/request.rs:518`, `response.rs:462`, `sse.rs:537`, `test_frames.rs:58`, `tool_bridge.rs:1313`, `tool_use_xml.rs:518` |

읽기 범위 밖: 위 54파일 이외의 모든 파일. 제품 코드 변경 없음. 보고서만 작성했다.
