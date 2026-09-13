# Native 정책 구현 검증

## Claim

승인된 core 0.9.0 native 정책을 명시적 프로필 안에 구현했다. 기본 subset과 기존 reasoning receipt gate는 유지한다. GPT 네 exact 모델의 high/adaptive 및 exact title schema 변환, title 결과 검사, Anthropic Opus 5·Sonnet 5 프로필의 native thinking/signature/redacted block 보존, 출력 schema 입력 계량을 범위 시험으로 확인했다. 실제 provider 또는 전체 moai gpt 통합 성공을 주장하지 않는다.

변경한 파일:

- internal/gateway/anthropic.go
- internal/gateway/input_estimate.go
- internal/gateway/native_policy_test.go
- internal/gateway/translate/native_policy.go
- internal/gateway/translate/native_policy_test.go
- internal/gateway/translate/request.go
- internal/gateway/translate/request_test.go
- internal/gateway/translate/response.go

stream.go는 수정하지 않았다. 기존 terminal 경로가 ResponseContext.response를 호출하므로 title 실패는 streaming 성공 terminal 이전에 거절된다. AUTH·receipt 패키지·launcher·factory·catalog·SPEC 본문은 수정하지 않았다. 출력 상한 subscription/API-key 분기와 SSE 보강은 유지했다.

## Evidence

TDD 첫 API RED:

```text
go test ./internal/gateway/translate -run 'TestNativePolicy|TestNativeTitle' -count=1
internal/gateway/translate/native_policy_test.go:15:150: unknown field PolicyProfile in struct literal of type Limits
internal/gateway/translate/native_policy_test.go:15:164: undefined: PolicyGPTNative
internal/gateway/translate/native_policy_test.go:18:65: c.title undefined (type *ResponseContext has no field or method title)
FAIL github.com/modu-ai/moai-adk/internal/gateway/translate [build failed]
```

이후 실제 동작 RED 두 건:

```text
go test ./internal/gateway -run TestNativePolicyWire -count=1
--- FAIL: TestNativePolicyWireAndProviderIsolation (0.00s)
    native_policy_test.go:72: same-model native wire rewritten
FAIL
```

```text
go test ./internal/gateway/translate -run TestNativeIdentity -count=1
--- FAIL: TestNativeIdentityNeverFallsThroughOrdinaryMetadata (0.00s)
    native_policy_test.go:113: native identity escaped without receipt
FAIL
```

native streaming 첫 fixture는 limits를 비워 기존 byte gate에서 실패했다. 실제 adapter가 공급하는 양수 event/output 한도를 fixture에 넣었으며 제품 한도를 완화하지 않았다.

새 metadata gate 이후 기존 golden 시험은 다음과 같이 실패했다:

```text
--- FAIL: TestCapturedPublicConversationGolden (0.05s)
    request_test.go:108: native receipt authorization required
```

기존 시험은 public-only projection이라고 하면서 native policy 세 필드만 제거하고 account/device/session JSON metadata를 foreign wire에 남겼다. 승인된 metadata 로컬 귀속 계약에 따라 해당 projection에서 metadata도 제외했다. 원래 raw fixture는 변경하지 않았으며 별도 TestNativeIdentityNeverFallsThroughOrdinaryMetadata는 기본 및 새 GPT profile 모두에서 송신 bytes가 반환되지 않음을 판정한다.

최종 실행:

```text
go test -race ./internal/gateway ./internal/gateway/translate -count=1 -coverprofile=/tmp/gateway-native-policy-final-coverage.out
ok  github.com/modu-ai/moai-adk/internal/gateway 9.061s coverage: 92.0% of statements
ok  github.com/modu-ai/moai-adk/internal/gateway/translate 1.872s coverage: 93.6% of statements
```

```text
go vet ./internal/gateway ./internal/gateway/translate
(exit 0, output empty)
```

coverage profile의 statement 가중합 직접 집계:

```text
translate/native_policy.go 73 76 96.05
translate/request.go 353 376 93.88
translate/response.go 117 120 97.5
anthropic.go 359 392 91.58
input_estimate.go 22 24 91.67
```

시험 범위: 네 GPT exact 모델 매핑, profile unknown/wrong profile, missing/null/type/enum/unknown/duplicate/invalid schema, invalid UTF-8·고립 surrogate, title의 누락/추가/중복/비문자열/깨진 JSON, title streaming success 및 실패 뒤 성공 terminal 없음/byte bound, native empty thinking·signature delta·redacted history·nonstream 문법, 잘못된 delta/block 결합과 EOF, 원본 same-model wire bytes, native invalid policy의 TLS 송신 0, 기존 text/tool·취소·바이트·SSE 회귀군이다.

## Baseline-attribution

작업 트리 /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified. 실행 readback:

```text
git rev-parse --short HEAD
81c1d58f9
git branch --show-current
WT-unified-gateway
git fetch -q origin main
git rev-list --count --left-right origin/main...HEAD
0 2879
```

HEAD는 미커밋 공유 작업의 기준이며 현재 dirty source를 함께 시험했다. Windows AUTH와 receipt 패키지의 다른 작성자는 별도 파일에서 작업했다. 위 커버리지는 이 실행의 dirty tree와 연결되며 clean-commit CI 판정이 아니다.

moai session current는 child runtime에서 canonical environment-fallback과 not-available을 출력했다. 부모가 지정한 source_session_id는 01a08e7b-6aa0-7361-ab7e-ea8da1f02228이며, child에서 UUID를 직접 재확인했다고 표시하지 않는다.

부모가 직접 읽은 native raw 12개의 metadata 구조는 account_uuid string empty, device_id/session_id string nonempty였다. 이 구조만 synthetic fixture에 반영했다. 원문 credential·계정·장치·대화 값을 tracked tests에 복사하지 않았다. native disabled+high 허용은 부모가 확인한 공식 Opus 5 문서 근거를 따른다. 이는 live OAuth 계정 cap 또는 서명 진위 검증이 아니다.

## 통합 접점

`translate.Limits.PolicyProfile`의 기본값은 기존 subset이다. `PolicyGPTNative`는 네 GPT exact model에서만 요청 정책을 활성화한다. `PolicyAnthropicNative`는 MessagesAdapter의 Anthropic provider와 Opus 5/Sonnet 5 exact upstream 모델에만 허용된다. GLM이 이 프로필을 자동 상속하지 않는다. 기존 adapter의 auth method·fixed endpoint 검사도 유지한다. catalog/root의 profile binding은 변경하지 않았다.

`translate.ValidateNativePolicy(body, profile)`은 엄격한 policy 문법만 검증하고 `NativePolicy{High,Title,KeepAll,UserID}`를 반환한다. 이것은 receipt 검증 결과가 아니다. GPT Request는 KeepAll 또는 UserID가 남아 있으면 receipt authorization 오류로 거절한다. 순차 codec 작성자는 원래 policy 문법을 검증하고 승인된 UUID/receipt 귀속을 먼저 확인한 뒤 metadata/context_management를 foreign wire에서 제거하여 변환기로 인계해야 한다. 임의 verified boolean이나 별도 receipt 저장소를 추가하지 않았다.

ResponseContext.response는 title 결과의 공개 text를 순서대로 합쳐 exact schema를 판정한다. receipt codec이 나중에 검증된 reasoning carrier를 content에 넣는 경우, title 검사는 공개 text 투영으로 연결하고 carrier를 title 문자열로 읽거나 무조건 삭제해서는 안 된다. 현재 Responses reasoning은 기존 gate에서 거절된다.

계량은 messages/system/tools에 output_config.format.schema를 추가한다. effort·metadata를 임의 reasoning token 값으로 만들지 않는다. json-runes-div4는 source projection 추정이며 tokenizer나 보장 상한이 아니다. 272000/872000/95 수치 또는 새 capability는 활성화하지 않았다.

## Gaps

실제 provider 송신, 네 모델의 전체 Claude Code 경로, receipt/UUID authorization·opaque codec·foreign strip·resume, live keep-all, production root/profile/cap binding, Windows runtime, 독립 감사는 이 작업에서 실행하지 않았다. LSP 진단 전용 도구 baseline은 수집하지 않았고 Go compiler/race/vet 결과만 있다. 서명 검사는 문법 보존이며 native signature 진위를 검증하지 않는다. core progress/status를 완료로 바꾸지 않았다.

프로젝트 integration branch의 GitHub CI run이 repository-wide test verdict의 소유자다. 아직 push/CI run 식별자가 없으므로 해당 전체 판정은 PENDING이다. 이 보고는 로컬 범위 시험만 기록한다.

## Residual-risk

명시된 프로필 외 native 입력은 여전히 실패할 수 있다. 추가 실제 필수 변형은 증거와 시험으로 보강해야 하며 완료 범위에서 삭제하지 않는다. title strict schema 양성은 일반 JSON Schema 지원을 뜻하지 않는다. high 의도 번역은 provider 간 계산량·토큰량 동등성을 뜻하지 않는다. 공유 dirty tree의 병행 변경은 이후 재측정이 필요할 수 있다.

## 인계 시 소스 SHA-256

```text
c8e3723f0631135f005cb9184471561da7dcb3e75b90343658faf926399f7d45  internal/gateway/anthropic.go
35000d38eb83bce9cadd33a0f6d25b8738111bfb3f2bcb990350521b55be6280  internal/gateway/input_estimate.go
47d4d493e85d40e9cb1211b0e27f7f4b9b913b2dddf731484bbc754e72b6777c  internal/gateway/native_policy_test.go
2b61bc9334989ed252d3e8bf0e75f216d62714bc473271990561fba950c651ce  internal/gateway/translate/native_policy.go
6c16c269d5885decbc12b900084a6570f428515a33f5a67796bc914d698f4fa3  internal/gateway/translate/native_policy_test.go
b2f988c596da5c1a04bced149e6afd6702518e03e50485ce30e089cd08eebfcd  internal/gateway/translate/request.go
ea3398182c8aa3fa05fef7a238e4800b8305843a0f06f59c5a32c68798541ca0  internal/gateway/translate/request_test.go
34861010f5c28d2771c9c6ee0f373e2c50e6079dcd478f10f8049fd898dca792  internal/gateway/translate/response.go
```
