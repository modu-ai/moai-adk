# 요청 단위 Anthropic OAuth passthrough 구현 검증

## Claim

M0의 실제 refresh POST 200→반환 Bearer hash→동일 본문 200 관측을 근거로, 요청 단위 Anthropic OAuth credential과 router·adapter 연결을 구현했다. 해당 관측은 `m0-refresh-transport-observation.md`와 ignored `gateway-entry/m0-refresh-transport-20260911T101909Z.json`에 있다. 세션 인증은 측정된 별도 `X-MoAI-Session-Token`이다.

`auth.NewAnthropicOAuth(ctx, headers)`는 요청 하나의 Bearer를 검증해 immutable 참조로 보유한다. 비어 있거나 중복된 Authorization(대소문자 키 변형 포함), 잘못된 Bearer 형식·padding·비ASCII·과도한 길이, x-api-key 동시 입력을 거부한다. 저장소·환경변수를 읽지 않고 refresh를 자체 수행하지 않는다. 원 요청 또는 outbound 요청이 취소되면 Apply를 거부한다. 정확한 Anthropic Messages endpoint/POST/Host만 허용하며 출력 표현은 Redacted로 고정한다.

router는 OAuth catalog 행만 이 참조로 해석한다. GPT/GLM/API-key resolver에는 inbound Bearer를 넘기지 않는다. 기존 `ResolveCredential(context.Context, ModelEntry)` 시그니처는 유지한다. `NewAnthropicOAuthAdapter`가 명시적 API-key adapter와 분리되어 있고, upstream 401은 같은 status와 authentication_error로 반환되어 Claude가 native refresh를 수행할 수 있다. adapter 자체 fallback/retry는 없다.

## Evidence

RED 명령:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test ./internal/gateway/auth ./internal/gateway -run 'TestAnthropicOAuth|TestOAuth' -count=1 -timeout 30s
```

관측 출력:

```text
internal/gateway/auth/credential_anthropic_oauth_test.go:14:16: undefined: NewAnthropicOAuth
internal/gateway/auth/credential_anthropic_oauth_test.go:18:11: undefined: NewAnthropicOAuth
internal/gateway/oauth_test.go:15:129: undefined: AuthOAuthPassthrough
internal/gateway/oauth_test.go:40:9: undefined: NewAnthropicOAuthAdapter
internal/gateway/oauth_test.go:41:35: undefined: AuthOAuthPassthrough
internal/gateway/oauth_test.go:42:24: undefined: auth.NewAnthropicOAuth
FAIL github.com/modu-ai/moai-adk/internal/gateway/auth [build failed]
FAIL github.com/modu-ai/moai-adk/internal/gateway [build failed]
FAIL
```

첫 TLS 시험은 sandbox의 listener bind 거부로 실행하지 못했다. 승인된 재실행에서는 기본 catalog가 OAuth로 바뀌면서 기존 범용 ingress fixture의 X-Test-Session 생성이 거부되었다. 해당 fixture는 명시적 API-key catalog로 바꾸어 저장소 resolver 계약을 계속 검증하도록 했다. OAuth 경로는 별도 실제 헤더 fixture로 검증한다.

GREEN 명령:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -race ./internal/gateway/auth ./internal/gateway -run 'TestAnthropic|TestOAuth|TestServer|TestSession|TestIngress|TestModels|TestConcurrentIngress|TestAdapterReceives|TestCatalog' -count=1 -timeout 45s -coverprofile=/tmp/gateway-oauth-coverage.out
ok  github.com/modu-ai/moai-adk/internal/gateway/auth 1.862s coverage: 6.9% of statements
ok  github.com/modu-ai/moai-adk/internal/gateway 2.103s coverage: 57.4% of statements
```

추가 credential 경계 시험 후:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -race ./internal/gateway/auth -run 'TestAnthropic' -count=1 -timeout 30s -coverprofile=/tmp/gateway-oauth-auth-coverage.out
ok  github.com/modu-ai/moai-adk/internal/gateway/auth 1.510s coverage: 8.1% of statements
```

`GOCACHE=/tmp/gateway-foundation-cache go tool cover -func=/tmp/gateway-oauth-auth-coverage.out`에서 새 `credential_anthropic_oauth.go`의 NewAnthropicOAuth, Provider, Generation, Redacted, String, GoString, Apply는 각각 100.0%였다. 패키지 전체의 낮은 비율은 의도적으로 다른 AUTH 기능 시험을 실행하지 않은 분모이며 전체 구현 커버리지 주장이 아니다.

마지막 401 응답 본문의 authentication_error 판정 추가 후:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -race ./internal/gateway -run 'TestOAuth' -count=1 -timeout 30s
ok  github.com/modu-ai/moai-adk/internal/gateway 2.213s
```

`GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/gateway/auth ./internal/gateway`: exit 0, 출력 없음.

## Baseline-attribution

이번 측정의 WT는 moai-proxy-unified, HEAD는 `81c1d58f9`, branch는 `WT-unified-gateway`다. 모든 시험은 실제 provider 대신 로컬 mock/TLS 서버를 사용했다. 이 구현 작업에서는 실제 Claude를 추가 호출하지 않았다. 코드가 다른 작업자의 미커밋 구현 위에 있으므로 git diff 통계가 각 신규 파일의 이번 변경량을 나타내지는 않는다.

변경 파일은 다음 9개다.

- `internal/gateway/auth/credential_anthropic_oauth.go` (신규)
- `internal/gateway/auth/credential_anthropic_oauth_test.go` (신규)
- `internal/gateway/oauth_test.go` (신규)
- `internal/gateway/auth/credential_anthropic.go` (기존 API 설명만 정합화)
- `internal/gateway/router.go`
- `internal/gateway/server.go`
- `internal/gateway/catalog.go`
- `internal/gateway/server_test.go` (명시적 API-key fixture)
- `internal/gateway/anthropic.go` (OAuth constructor와 설명만 추가)

## Gaps

- root/production factory는 활성화하지 않았다.
- nativeRequest validator는 변경하지 않았다. adaptive thinking, output_config, context_management 등 실제 client 본문 정책은 별도 후속 구현·검증 대상이다.
- AllowedBetas는 factory가 실제 측정된 필요한 beta를 명시해야 한다. 이 범위 시험은 configured oauth beta의 전달을 입증하며 실제 전체 beta 집합을 정한 것은 아니다.
- 기본 catalog Claude 행은 OAuth로 바뀌었지만 capability 0은 유지한다. 실제 측정한 capability/MeasureInput 없이 native adapter 송신이 열리지 않는다.
- API-key 행은 NewCatalog로 명시적으로 구성할 수 있다. 같은 Anthropic provider에서 인증 방법이 다른 행을 동시에 쓰는 factory의 adapter 선택은 아직 연결하지 않았다.
- generation=1은 immutable 요청 참조의 수명 안에서만 의미가 있다. 다음 native refresh 요청은 새 참조이며 전역 로그인 generation이나 logout barrier를 대신하지 않는다.
- 저장소 전체 시험 및 integration branch CI의 repository-wide 판정은 PENDING이다.

## Residual-risk

M0 양성은 이 구현의 사전 조건을 충족한 관측이고, 제품 전체 PASS나 배포 승인이 아니다. 새로운 참조는 요청 수명만큼 Bearer를 메모리에 보유한다. Redacted/String/GoString이 로그 안전 표현을 제공하지만 모든 호출자의 별도 raw-header logging 부재까지 감사한 것은 아니다. 다른 AUTH worker의 store/Windows 변경은 손대지 않았다.
