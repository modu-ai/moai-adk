# Private child handler factory 검증

## Claim

strict private payload에서 승인된 정확한 catalog 부분집합을 고르고 기존 고정 endpoint adapter·credential resolver를 연결하는 factory를 구현했다. root 등록은 여전히 `newGatewayChildCommand(nil)`이며 제품 활성화는 하지 않았다.

payload는 `version:1`, `session_token`, `model_ids`만 허용한다. duplicate JSON key·malformed·unknown field·잘못된 version·unknown model·중복 model·빈 model 집합은 resource를 열기 전에 거부한다. caller가 검증한 Models/Capabilities, Limits, MeasureInput, AnthropicVersion/AllowedBetas, transport를 주입해야 한다. payload에서 URL·credential·capability를 덮어쓰는 통로는 없다.

GPT 행에는 Broker와 RefreshVerifier가 필수이며 Store.ResolveFresh에 request context가 전달된다. 없는 로그인은 자동 로그인하지 않는다. OAuth는 기존 request-scoped router가 처리하고 GLM은 기존 credential reader를 사용한다. API-key catalog 행은 factory 구성에서 명시적으로 거부하며 기존 별도 API-key adapter 자체는 보존했다.

## Evidence

RED 명령:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test ./internal/cli -run '^TestGatewayFactory' -count=1 -timeout 30s
```

관측 출력:

```text
internal/cli/gateway_factory_test.go:22:40: undefined: gatewayFactoryDependencies
internal/cli/gateway_factory_test.go:24:9: undefined: gatewayFactoryDependencies
internal/cli/gateway_factory_test.go:30:15: undefined: newGatewayHandlerFactory
internal/cli/gateway_factory_test.go:35:31: undefined: newGatewayHandlerFactory
internal/cli/gateway_factory_test.go:47:15: undefined: newGatewayHandlerFactory
internal/cli/gateway_factory_test.go:60:15: undefined: newGatewayHandlerFactory
FAIL github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

첫 구현 후 시험에서 macOS temp 경로의 symlink를 AUTH store가 의도대로 거부했다. 시험의 temp 부모를 기존 AUTH 시험과 같은 EvalSymlinks 처리로 고쳐 실제 소유 Store를 열도록 했다.

최종 GREEN 명령과 출력:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -race ./internal/cli -run '^TestGatewayFactory|^TestGatewayChild|^TestGPT' -count=1 -timeout 45s -coverprofile=/tmp/gateway-factory-coverage.out
ok  github.com/modu-ai/moai-adk/internal/cli 2.967s coverage: 7.0% of statements
```

시험에서 확인한 동작:

- payload 부적합 시 OpenStore 호출 0.
- 승인 metadata·MeasureInput·GPT refresh verifier가 없거나 API-key/다른 upstream 행이면 구성 거부.
- TLS mock에서 OAuth는 api.anthropic.com/v1/messages, GLM은 api.z.ai/api/anthropic/v1/messages로 송신. OAuth inbound Bearer는 GLM에 전달되지 않고 저장된 모의 GLM key를 사용. GLM beta와 별도 session header 제거.
- native 401 그대로 반환. 무인증/취소 요청은 추가 송신 0.
- GPT 로그인 부재는 401, broker 호출 0.
- adapter 초기화 실패 후 실제 Store.Close, 정상 handler의 idempotent Close, child의 잘못된 parent로 시작 실패할 때 handler.Close 관측.
- Close가 진행 중 request context를 취소하고 요청 종료를 기다림. 닫힌 handler는 503.

`GOCACHE=/tmp/gateway-foundation-cache go tool cover -func=/tmp/gateway-factory-coverage.out` 관련 출력:

```text
newGatewayHandlerFactory 87.2%
ServeHTTP 100.0%
Close 100.0%
openGPTAuthStore 77.8%
```

마지막 함수는 기존 GPT CLI의 경로 로직을 그대로 추출한 것이며 새 factory 로직 커버리지가 아니다. 전체 CLI의 7.0% 역시 범위 밖 기능을 시험하지 않은 분모다.

`GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/cli`: exit 0, 출력 없음.
`gofmt -l` 네 파일: 출력 없음.
`rg -n 'newGatewayChildCommand' internal/cli/root.go`:

```text
156: rootCmd.AddCommand(newGatewayChildCommand(nil))
```

## Baseline-attribution

WT moai-proxy-unified, HEAD `81c1d58f9`. 실제 provider 호출 없는 로컬 TLS/mock/Store 시험이다. 변경 파일은 다음 네 개다.

- `internal/cli/gateway_factory.go` (신규)
- `internal/cli/gateway_factory_test.go` (신규)
- `internal/cli/gateway_child.go` (io.Closer 정리)
- `internal/cli/gpt_auth.go` (기존 withStore 경로를 openGPTAuthStore로 최소 추출)

openGPTAuthStore는 기존 MoaiHome→MkdirAll0700→EvalSymlinks→gateway-auth/OpenStore 순서를 유지한다. factory와 login/logout/status가 같은 helper를 쓰며 CODEX_HOME을 조회하지 않는다.

## Gaps

- production root/launcher, 실제 capability·MeasureInput·beta 정책, private picker state·opaque receipt는 별도 게이트다.
- GPT의 실제 송신은 이 시험에서 관측하지 않았다. Store와 OpenAI adapter의 기존 소유권 계약을 연결하고 credential 부재의 무송신 경계를 검증했다.
- API-key 행은 이 최소 factory에서 지원하지 않는다. 행을 조용히 삭제하거나 ambient key로 대신하지 않는다.
- Windows Store native 실행과 전체 repository 판정은 integration branch GitHub CI 소관이며 PENDING이다.
- 프로덕션 Store의 close/refresh 동시성 전체 감사는 별도 AUTH 계약이다. 이 wrapper는 요청을 취소·join한 뒤 Store.Close한다.

## Residual-risk

승인 metadata 의존성을 주입할 수 있다는 것은 실제 지원을 입증했다는 뜻이 아니다. 기능이 관측되기 전에 caller가 값만 채워 factory를 활성화해서는 안 된다. Transport는 복제되어 기존 adapter의 검증·고정 endpoint 정책을 사용하며 새로운 일반 transport나 credential 복사 체계를 만들지 않았다. Store/init 오류는 private payload나 credential을 출력하지 않는 일반 오류로 반환한다.
