# Codex provider 전체 독해 및 MoAI 구조 개선 제안

기준: `/tmp/moai-proxy-structural.IArJmM/source`, SHA `b642a6ee1a5fff8ebc9bedfc4b438ddfa3ae8f8e`. **담당 범위 35개 Rust 파일 30,465행 전부 독해 완료.** 제품 코드 변경 없이 분석했다. 구독 인증 경로를 복제하지 않고 MoAI의 Codex App Server 경계를 유지한다. 저장소 전체 감사와 실제 모델 접속 검증은 별개의 범위다.

## 실제 독해 범위

`src/providers/codex/`의 35개 Rust 파일, 30,465행을 기준으로 한다. 아래 범위는 출력이 잘리지 않은 본문을 실제 읽은 범위다.

|파일|총행|읽은 범위|
|---|---:|---|
|events.rs|572|1–572|
|client.rs|5336|1–5336|
|websocket.rs|4303|1–4303|
|continuation.rs|1165|1–1165|
|compaction.rs|977|1–977|
|translate/mod.rs|15|1–15|
|translate/accumulate.rs|622|1–622|
|translate/read_rewrite.rs|139|1–139|
|translate/reasoning_signature.rs|145|1–145|
|translate/reducer.rs|1874|1–1874|
|translate/model_allowlist.rs|222|1–222|
|translate/web_search_compat.rs|205|1–205|
|translate/live_stream.rs|1703|1–1703|
|translate/stream.rs|758|1–758|
|translate/request.rs|2607|1–2607|
|mod.rs|2550|1–2550|
|count_tokens.rs|260|1–260|
|images.rs|976|1–976|
|native.rs|792|1–792|
|request_summary.rs|322|1–322|
|search.rs|683|1–683|
|transcription.rs|502|1–502|
|chat_completions/mod.rs|471|1–471|
|chat_completions/request.rs|549|1–549|
|chat_completions/response.rs|240|1–240|
|chat_completions/stream.rs|325|1–325|
|auth/browser_login.rs|491|1–491|
|auth/constants.rs|9|1–9|
|auth/device.rs|624|1–624|
|auth/jwt.rs|203|1–203|
|auth/manager.rs|372|1–372|
|auth/mod.rs|9|1–9|
|auth/pkce.rs|198|1–198|
|auth/test_http.rs|132|1–132|
|auth/token_store.rs|114|1–114|

완독 **35/35파일, 30,465/30,465행(100%)**. 담당 범위 내 미독 파일은 없다. 대체로 600–800행 단위의 `sed -n` 청크 또는 작은 파일 `cat`으로 끝까지 읽었다. 중간 출력이 잘린 `translate/stream.rs` 앞부분은 1–280행을 별도 재독해했다. `find src/providers/codex -name '*.rs' -exec wc -l {} +`의 최종 출력은 `30465 total`이다.

## 실제 검증

동일 상류 체크아웃에서 실행:

```text
cargo test --offline --locked providers::codex::translate -- --test-threads=2
Finished `test` profile [unoptimized + debuginfo] target(s) in 32.62s
test result: ok. 126 passed; 0 failed; 0 ignored; 0 measured; 758 filtered out; finished in 0.02s

cargo test --offline --locked --lib providers::codex --quiet -- --test-threads=2
running 394 tests
test result: ok. 394 passed; 0 failed; 0 ignored; 0 measured; 490 filtered out; finished in 55.16s

git status --short
(출력 없음)
```

기준 귀속: 상류 SHA의 로컬 Rust 테스트다. 실제 GPT 접속, 전체 테스트, MoAI 통합 성공, 성능 수치를 의미하지 않는다. 테스트 명세 자체의 타당성도 별도 검토 대상이다.

MoAI 비교 기준은 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/gpt-session-repair`, HEAD `b45c813493751106cd6619dce60fba1c61e70583`와 현재 공유 작업 변경분이다. 상류 SHA와 혼동하지 않는다. 해당 트리에서 추가 실행:

```text
go test -race ./internal/codexbridge ./internal/codextools -count=1
ok  github.com/modu-ai/moai-adk/internal/codexbridge 16.052s
ok  github.com/modu-ai/moai-adk/internal/codextools 1.856s
```

## 확인한 구조와 개선 제안

- 요청 변환은 Anthropic messages/tools를 Responses input/function 도구로 바꾼다. lite 경로는 도구와 instructions를 developer prefix로 옮기는 별도 변형이다. 이 비공개 형태를 MoAI에 복사하지 않고 App Server의 공개 계약 안에서 처리해야 한다.
- `events.rs`는 control/structural/semantic/terminal-success/terminal-failure를 분리한다. live HTTP는 semantic 출력 전만 재시도하고 구조 이벤트를 보류한다. 정보성 ping은 사용자 의미 출력과 구별한다. MoAI에서도 재시도 허용 지점을 HTTP 헤더 전송 여부만으로 결정하지 않고 도구 실행/의미 출력 경계로 정의하는 것이 적절하다.
- buffered stream과 non-stream은 공통 reducer를 사용하지만 live translator는 별도 구현이다. buffered reducer는 terminal 이후 이벤트와 열린 블록을 거부하지만 live는 완료 이후 무시하고 열린 블록을 닫는다. 통일된 정규 이벤트 reducer와 출력 어댑터로 상태 규칙의 중복을 줄이는 방향을 제안한다. 이는 소스 차이 관찰이며 사용자 장애 재현 주장이 아니다.
- continuation은 typed owner(main/agent), turn generation, 실제 origin socket ID를 함께 검증한다. 오래된 비동기 완료/취소가 새 상태를 지우거나 게시하지 않도록 보호한다. 함수 인자는 raw 문자열이 아니라 JSON 의미 비교를 사용한다. owner별 2MB/전체20MB/TTL 예산이 있다.
- compaction은 Preparing→Unconfirmed→Anchored 단계, attempt ID, 모델 일치, portable summary anchor, TTL/용량 제한을 사용한다. MoAI에는 세대·앵커 검증 원칙을 흡수하고, opaque encrypted reasoning/compaction replay는 App Server에 맡긴다.
- 상류에는 Read 인자 1024자 공백 정체 시 JSON을 닫아서 tool-use 완료로 간주하는 호환성 처리, 큰 offset 제거, schema pattern 제거, URL 이미지 placeholder, luna의 hosted-search 모델 치환이 있다. 이를 일반적인 품질 보장으로 간주하거나 그대로 복제하지 않는다. 사용자 지정 모델 보존과 명시적 capability 오류가 우선이다.
- HTTP/WS live pending 구조 이벤트 Vec와 buffered HTTP 응답 Vec에서 이 단계의 코드상 총 byte cap은 보이지 않는다. 채널64의 역압력과 SSE 단일 프레임8MiB 제한은 있으나 합계 보류량과는 다르다. 실제 메모리 폭증 재현은 아직 하지 않았다. MoAI의 owner별 byte mailbox처럼 각 보류 저장소에 합계 예산을 적용할 필요를 검증 항목으로 남긴다.

## 흐름별 상세 비교

|영역|상류 실제 구현|MoAI 확인 지점|개선 방향|
|---|---|---|---|
|전송 소유권|`client.rs`가 HTTP/WS 요청, 재인증, fallback을 직접 담당|`gateway/appserver.go:34`는 ManagedGrant 확인 후 Engine으로 전달하며 credential을 받지 않음|App Server가 계정·네트워크·원격 thread를 소유하고 gateway는 Claude 계약만 소유한다. 상류 직접 endpoint나 내부 header 복제 금지|
|상태 소유권|`continuation.rs:149,327`의 typed owner/turn 예약, `websocket.rs`의 정확한 socket ID|`codexbridge/engine.go:92` conversation에 phase/turn/pending/deferred/usage/failure 저장|현재 owner 검증을 유지하되 phase를 enum으로 명시하고 전이 함수를 단일화한다. 비동기 결과에 owner+turn+attempt 식별을 요구한다|
|도구 실행|완성된 function args를 Claude tool_use로 재포장. Read 인자 일부 수정|`codextools/registry.go:98,282`가 전체 schema 검증. `engine.go:702` 부근은 검증 실패를 최대3회 명시적 실패 결과로 전달|MoAI 방식을 보존한다. JSON 자동 수선·pattern 제거로 성공처럼 보이게 하지 않는다. 새 수정도 schema 위반 시 실행0회를 입증해야 한다|
|도구 결과|`translate/request.rs`에서 순서대로 text/image function output 변환|`engine.go:480` 이후 pending ID 전체 일치, 중복 금지, 크기 제한; `responding` 저장 후 RPC 응답|callID/RPCID/publicID의 의미를 타입으로 분리하고, 부분 응답 후 재전송 금지를 불변식으로 고정한다|
|실시간 출력|live translator와 buffered reducer가 분리됨. provider가 전체 SSE 재보관 후 종료 시 buffered reducer로 continuation 검사|`engine.go:426,433` Step/StepStream이 같은 루프 사용. `appserver.go:84`는 text 즉시 출력, tool/terminal은 durable barrier 이후|MoAI의 공통 실행 루프는 장점이다. stream/nonstream에 별도 상태 머신을 추가하지 않고 Segment/이벤트 렌더러만 분리한다|
|이벤트 보관|WS live `stream_ws_events`가 sse_body를 계속 축적하고 provider도 upstream_sse_body를 축적|`event_queue.go:13` byte mailbox, 64KiB 최소, 초과 owner만 실패|현재 예산을 유지한다. mailbox 외 pending tools·deferred·segment text·모든 owner 합계에 대한 예산 명세를 별도로 둔다. 상류 무제한 Vec 패턴은 흡수하지 않는다|
|재시도|HTTP 자체3회, provider WS10회, 빈완료10회, origin403 handshake1회. HTTP 소진을 provider가 다시 곱하지 않도록 분기|`engine.go:523` 이후 불확실한 RPC/도구 응답 자동 재시도 금지|App Server 위에 이 재시도 수치를 복사하지 않는다. `not_sent / accepted / semantic_published / side_effect_unknown` 구분과 단일 예산을 먼저 정의한다|
|취소|Drop guard와 receiver closed로 owner 예약 취소. retry handoff는 예약을 유지하고 이전 socket만 정리|`appserver.go:77,84` body close→cancel, `engine.go` canceled recovery는 정확한 turn terminal 확인 후만 허용|취소·완료·overflow 경쟁을 한 전이 API로 처리한다. late-start의 turn ID를 버리지 않는지도 회귀 사례로 고정한다|
|압축|`compaction.rs:37` 단계와 attempt/model/summary anchor 검증. 내부 opaque history를 별도 replay|MoAI bridge는 App Server thread resume을 사용하고 native transcript를 별도로 검증|native 압축 책임은 App Server에 둔다. gateway는 receipt/summary/원래 completed prefix의 연결만 검증한다. 두 개의 원격 이력 엔진을 만들지 않는다|
|측정|텍스트는 o200k tokenizer, 이미지2000 등 추정. `request_summary.rs`는 내용 없이 크기·개수 요약|`appserver.go:265` nil usage는 unknown, protocol placeholder0과 구분|추정/관측/placeholder 구분을 보고서와 telemetry에서도 유지한다. 첫 의미 출력·도구 대기·gateway 처리·원격 대기를 분리 측정한다|

### 단순히 상류를 복제하면 안 되는 이유

1. 상류의 강점은 형식 변환과 실패 분류, owner/turn/socket 세대 보호다. App Server가 이미 담당하는 인증·네트워크 재시도·암호화된 추론 재전송까지 이식하면 책임이 겹친다.
2. 상류 `mod.rs:651` 이후는 준비/재시도/첫 출력/나머지 출력/정리로 경계를 나누지만, live와 buffered의 성공 판정은 다르다. 테스트394개 통과는 각 구현의 명세 통과이지 두 구현의 완전한 의미 동등성 증명이 아니다.
3. 상류 WS 신규 연결은 전역1초 시작 간격, origin403 재연결 대기3초를 사용한다. 이를 App Server 앞에 추가하면 원격 클라이언트 정책 위에 대기가 중첩된다. MoAI에는 해당 정책을 복제할 이유가 확인되지 않았다.
4. `images.rs:6`의 모델은 **gpt-image-2**다. 사용자 요구 **gpt-image-2.5** 지원을 이 코드로 입증할 수 없다. 이미지 byte cap, 동시성 제한, 허용 header, redirect 거부 원칙만 참고하고 실제 지정 모델은 별도 capability/실접속 gate로 검증한다.
5. 상류 Chat Completions 경로는 text-only이며 tools/max_tokens 등 미지원 파라미터를 명시적으로 거부한다. 이것은 Claude Code의 도구 세션을 대체하는 경로가 아니다. MoAI 주경로에 채택하지 않는다.
6. standalone search는 실제 structured results가 없으면 URL을 지어내지 않지만 buffered hosted web_search 호환 경로는 답변 URL에서 결과를 추출한다. 모든 검색 결과가 동일한 증거 수준이라는 주장은 피해야 한다.

## 구조적 개발 지시 제안 — 기존 장점을 보존하는 최소 분리

우선순위 High는 `Engine.StepStream`의 구현을 갈아엎는 것이 아니라, 현재 동작을 characterization 테스트로 고정한 다음 **전이·부수효과·렌더링** 세 책임을 나누는 것이다.

1. `TurnState`와 `Transition(event)`를 추출한다. new/starting/active/waiting/responding/idle/failed/canceled/retired를 enum으로 표현한다. 출력은 text/tool-ready/terminal/failure와 RPC 응답·저장·interrupt 명령이다. RPC read goroutine은 상태를 임의 변경하지 않고 해당 owner의 실패 전이만 요청한다. 단, 공유 reader에서 blocking 효과를 실행하면 안 된다.
2. `EffectRunner`는 기존 Store/RPC 인터페이스를 그대로 사용한다. durable waiting/idle 저장이 성공해야만 tool/성공 terminal을 게시한다. 저장 실패, owner overflow, 취소는 성공 전이를 이길 수 있어야 한다. 별도 framework나 새로운 transport를 만들지 않는다.
3. `RetryDecision`은 transport request가 시작되기 전의 재시도만 우선 허용한다. RPC 쓰기 후 결과 불확실, tool 응답 일부 전송, 이미 공개 text가 있는 경우에는 replay하지 않는다. 원격 capacity를 새 모델로 조용히 바꾸지 않는다.
4. `CapabilityContract`는 네 모델과 effort, 동적 도구, 이미지 입력, 이미지 생성 경로를 시작 시 관측한 App Server 버전에 귀속한다. 미지원은 명시 오류로 끝내고 숨은 대체를 하지 않는다. gpt-image-2.5는 text 모델 allowlist와 별도로 관리한다.
5. event classification은 control/progress/semantic/tool-request/usage/terminal로 정리한다. 알 수 없는 RPC 요청은 명시 거부, 알 수 없는 notification은 bounded 처리한다. 정보성 이벤트를 버릴 때도 turn/thread 불일치는 정상으로 덮지 않는다.
6. telemetry에는 phase, cause code, queue bytes, pending RPC 수, first-semantic latency, tool wait, request cancellation을 기록한다. prompt·도구 출력·token 원문은 기본 기록하지 않는다. 원격 지연과 gateway CPU/대기를 나눠야 속도 문제를 증명할 수 있다.

### 구현 시 필수 회귀 gate

- 같은 native fixture를 stream/nonstream으로 실행하여 최종 content/tool IDs/usage/stop reason이 동일한지 비교한다.
- semantic 출력 전 오류만 제한적으로 재시도하고, 공개 출력 후·tool 실행 후·responding 중 실패는 재실행0회인지 확인한다.
- 느린 owner의 tool 대기 중 다른 owner RPC 응답이 막히지 않는지 확인한다.
- selected quiet timer, queued terminal, overflow, cancellation, late-start 응답을 순서별로 재배열한 결정론 테스트를 둔다.
- generation이 다른 completion/abort/compact 응답이 새 owner state를 바꾸지 못하는지 확인한다.
- schema required/type/pattern/enum/$ref, malformed JSON, 도구별 authority 변경, 결과 순서 변경·누락·중복을 검사한다.
- 실제 Claude Code에서 main GPT 대화, 병렬 Agent, ToolSearch, tool 결과 이미지, resume, compact 후 resume, Ctrl-C, 잘못된 model/effort, capacity 제한을 별도로 검증한다.
- 속도 gate는 단일 총시간이 아니라 cold/warm startup, first-semantic, 도구 왕복 gateway 추가지연, multi-owner 간섭으로 기록한다. 현재 보고서는 해당 live 수치를 주장하지 않는다.

## 미검증 및 잔여 위험

담당 Codex 파일은 모두 읽었지만 저장소 밖/다른 provider/전체 integration fixture는 이 담당 범위가 아니다. 상류394개 lib 테스트와 translate126개는 중복 포함이며 합산하여520개 독립 사례라고 보고하면 안 된다. MoAI 전체 suite 및 integration branch CI verdict는 **PENDING**이다. 상류 실서버 호환성·인증 약관·실제 네 모델·gpt-image-2.5 가용성을 판정하지 않았다. 메모리 보관 차이는 소스 관찰이며 OOM 실험으로 확인된 결함 주장이 아니다. 이 보고서의 개발 지시는 제안이고 제품 코드는 변경하지 않았다. 실제 테스트된 범위 밖의 Claude Code 모든 기능 완전 지원·속도 무회귀·합법성 보장을 선언하지 않는다.
