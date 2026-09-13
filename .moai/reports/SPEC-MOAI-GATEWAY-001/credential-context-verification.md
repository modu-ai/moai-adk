# Credential resolver 요청 context 전달 검증

## Claim

ServerConfig.ResolveCredential과 Server 내부 resolver의 기존 단일 서명을 `func(context.Context, ModelEntry) (CredentialRef,error)`로 바꿨다. ServeHTTP의 `r.Context()`를 router.credential을 거쳐 resolver에 그대로 전달한다. 별도 legacy field나 우회 API는 없다. factory가 ResolveFresh에 요청 취소·기한을 전달할 수 있는 연결이며 실제 factory 활성화는 하지 않았다.

수정 파일은 server.go, router.go, server_test.go, openai_test.go다. 후자의 기존 resolver lambda와 server_test의 resolver 실패 함수 배열도 새 서명으로 맞췄다. session 인증·body/path/model 판정 순서와 로컬 validation 정책은 바꾸지 않았다.

## Evidence

신규 요청 취소 시험을 먼저 작성한 RED:

`go test ./internal/gateway -run TestServerCancellation -count=1 -timeout 10s`

```text
# github.com/modu-ai/moai-adk/internal/gateway [github.com/modu-ai/moai-adk/internal/gateway.test]
internal/gateway/server_test.go:411:214: cannot use func(ctx context.Context, _ ModelEntry) (CredentialRef, error) {…} (value of type func(ctx context.Context, _ ModelEntry) (CredentialRef, error)) as func(ModelEntry) (CredentialRef, error) value in struct literal
FAIL	github.com/modu-ai/moai-adk/internal/gateway [build failed]
FAIL
```

서명 변경 후 기존 실패 resolver 배열의 옛 타입이 한 곳 더 드러났고 같은 시험 파일에서 보정했다. 초기 runtime 명령은 로컬 httptest bind가 sandbox에서 거절되어 실패했다. 이는 제품 반례 RED로 세지 않았다.

```text
panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted
```

승인된 sandbox 밖 재실행(로컬 TLS fixture만 사용):

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -race ./internal/gateway -run 'TestServer|TestOpenAI|TestCredentialResolver|TestSession|TestIngress|TestModels|TestConcurrentIngress|TestAdapterReceives' -count=1 -timeout 30s && GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/gateway
```

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.652s
```

명령 exit 0. vet stdout/stderr는 빈 출력이다. selector는 server_test.go의 기존 15개와 신규 취소 시험, OpenAI local adapter 시험을 포함한다. 새로운 취소 시험은 resolver 진입 channel을 관측한 뒤 request context를 취소하고 resolver가 context.Canceled를 받는지, handler가 종료하는지, adapter 호출이 0이며 기존 401 오류 정책인지 확인한다. deadline을 가진 context와 완료 channel로 fixture 수명을 제한한다.

## Baseline-attribution

2026-09-11 WT moai-proxy-unified, 부모가 고정한 HEAD 81c1d58f9. 변경 전에 네 파일의 resolver 선언·호출·lambda와 router 본문을 읽었다. 다른 작성자의 input_estimate 등 현재 변경을 보존했다. 이번 작업에 auth·CLI·SPEC·생산 활성화·실제 client/provider 호출·git mutation은 없다.

## Gaps

실제 factory에서 ctx를 ResolveFresh/broker/verifier에 넘기는 연결과 실계정 refresh는 아직 별도 작업이다. CredentialRef.Generation은 기존 무context 인터페이스 그대로다. 요청 context 전달을 모든 credential 작업의 취소 가능성으로 확대하지 않는다. 취소를 무시하는 임의 resolver를 서버가 강제 중단하는 API는 만들지 않았다. 실제 provider·Windows·전체 repository CI 판정은 없으며 통합 브랜치 CI verdict는 PENDING이다.

## Residual-risk

resolver가 context를 무시하면 호출은 그 구현의 종료까지 기다린다. 현재 취소 오류는 다른 credential 오류와 같은 401로 축약하며 HTTP 취소 전용 상태를 새로 정의하지 않았다. 실제 연결 종료 후 클라이언트가 그 응답을 읽는다는 뜻은 아니다. 향후 factory는 supplied ctx를 그대로 사용하고 SendAuthorized의 최종 세대 장벽을 유지해야 한다.
