# Ordered opaque envelope와 tool ID codec

## Claim

독립 패키지 internal/gateway/opaque의 Encode/Decode 및 BindToolID/RestoreToolID를 구현했다. JSON 원문 byte·encrypted string·item 순서와 원래 tool call ID를 보존하고 version/provider/ID/hash/크기/중복/unknown/UTF-8/고립 surrogate를 검사한다. session·receipt 권한이나 provider 서명 진위를 검증하지 않는다.

형식은 `moai_opaque_v1_` 뒤 unpadded canonical base64url JSON `{version:1,provider:"openai",items:[{output_index,raw}]}`다. raw는 원래 reasoning item JSON의 문자열이므로 JSON 공백·문자 escape까지 왕복 보존한다. output_index 단위는 **원래 Responses output 배열의 index**다. 공개 item 사이에 reasoning이 있던 위치를 보존하되 공개 item 재구성은 이 패키지가 하지 않는다.

Tool marker는 승인된 `toolu_moai_v1_` 뒤 compact base64url JSON `{call_id,opaque_sha256}`이며 exact envelope bytes SHA-256을 비교한 뒤 원래 ID를 복원한다. tool_use/tool_result 호출자는 같은 API를 사용한다. 반환된 Items/Bytes를 수정해도 Envelope 원본은 바뀌지 않는다. native redacted.data에는 ErrNotEnvelope를 반환하고, `moai_opaque` 예약 접두어가 잘못된 경우 일반 native 자료로 통과시키지 않는다. `toolu_moai_` 예약 ID 재binding도 거절한다.

경계는 envelope JSON 8MiB, 원래 단일 item 4MiB, 1024 items, output index 0..100000 strictly increasing, ID 1..1024 ASCII 영숫자·underscore·hyphen, tool marker 최대 4096 bytes다. 입력/output body의 별도 실행 한도를 대체하지 않는다. 이 수치는 codec 로컬 방어 한도이며 native/server 최대치 관측이 아니다.

## Evidence

첫 API RED:

```text
go test ./internal/gateway/opaque -count=1
internal/gateway/opaque/codec_test.go:12:9: undefined: Encode
internal/gateway/opaque/codec_test.go:12:18: undefined: Item
internal/gateway/opaque/codec_test.go:13:9: undefined: Decode
FAIL github.com/modu-ai/moai-adk/internal/gateway/opaque [build failed]
```

예약 접두어와 공식 content 변형의 동작 RED:

```text
go test ./internal/gateway/opaque -run TestReserved -count=1
--- FAIL: TestReservedPrefixesAndDuplicateItemIDs (0.00s)
    codec_test.go:83: reserved prefix treated as native
FAIL
```

```text
go test ./internal/gateway/opaque -run TestReasoningShapes -count=1
--- FAIL: TestReasoningShapesAndOrder (0.00s)
    codec_test.go:100: invalid MoAI opaque encoding
FAIL
```

최종 GREEN:

```text
go test -race ./internal/gateway/opaque -coverprofile=/tmp/opaque-final-cover.out -count=1
ok  github.com/modu-ai/moai-adk/internal/gateway/opaque 1.418s coverage: 95.2% of statements

go vet ./internal/gateway/opaque
(exit 0, output empty)
```

coverage: Encode92.0%, Decode97.0%, validateItem91.9%, BindToolID/RestoreToolID100%, strict JSON walk89.7%, object100%, Unicode escape95.8%. 시험은 합성 reasoning의 정확 raw 왕복, index 사이 공개 item 간격, hash 변조, 중복ID/index, 뒤집힌 순서, unknown/null/type/필수 필드/미완료 status, 중복 JSON key, canonical base64/JSON, 잘못된 version/provider, Unicode/깊이/크기, alias mutation을 판정한다.

보존 private Luna fixture를 Python json으로 읽어 값 없이 구조만 확인했다:

```text
output_index 3 shape {'id': 'str', 'type': 'str', 'content': {'type': 'array', 'length': 0, 'members': []}, 'encrypted_content': 'str', 'summary': {'type': 'array', 'length': 0, 'members': []}}
```

출처는 `.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-live/responses/gpt-5.6-luna-opaque-followup-request-790052897.json`이다. 이 index는 관측 fixture input 위치이며 제품 envelope의 output index를 추정한 값이 아니다. 위 shape만 synthetic tests에 반영했다. 실제 opaque bytes의 제품 roundtrip 또는 provider 재송신을 실행했다고 주장하지 않는다.

공식 로컬 Codex snapshot `/tmp/openai-codex-audit.zbvNXB/codex-rs/protocol/src/models.rs:1971–1981`의 enum을 직접 읽어 summary_text와 content의 reasoning_text/text 문법을 확인했다. status absent 및 completed를 허용하고 in_progress는 거절한다. summary/content의 text는 해석하거나 공개 응답으로 변환하지 않고 raw 안에 그대로 보존한다.

## Baseline-attribution

WT /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified, 실행 HEAD81c1d58f9, branchWT-unified-gateway. 구현 소유 파일은 internal/gateway/opaque/codec.go, json.go, codec_test.go와 이 보고서뿐이다. 다른 패키지·SPEC·원격은 변경하지 않았다. source_session_id는 부모가 지정한 01a08e7b-6aa0-7361-ab7e-ea8da1f02228이다. child moai session은 앞서 fallback/not-available였으며 그 UUID를 직접 조회했다고 주장하지 않는다.

## Gaps

gateway/translate 연동, 전체 public/opaque interleaving 재구성, tool pair 의미 검사, no-tool receipt 누락 검출, 세션 UUID·private store 귀속, foreign strip, terminal 이전 receipt publish, 실제 native tool/no-tool carrier, 실제 GPT 후속·재개는 미실행이다. 기능을 활성화하지 않았다. codec은 pure bounded CPU 변환이며 네트워크·goroutine·저장소를 만들지 않는다. 새 LSP baseline은 없고 compiler/race/vet를 실행했다. integration branch GitHub CI가 repository-wide 판정 소유자이며 run ID 없는 현재 판정은 PENDING이다.

## Residual-risk

두 base64 계층과 JSON string escaping이 carrier 크기를 늘릴 수 있다. integration은 변환 후 실제 body/output 한도를 확인해야 한다. canonical envelope의 hash는 바이트 일관성만 확인하며 공격자의 새 carrier 생성이나 다른 세션 replay를 막지 않는다. 정확 receipt 검사가 선행해야 한다. 공개 item splitting·모델 전환 정책을 codec의 index 보존 성공으로 대체할 수 없다.
