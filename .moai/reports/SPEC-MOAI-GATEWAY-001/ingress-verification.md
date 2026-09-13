# HTTP ingress 구현 검증

## Claim

새 server.go/router.go/policy.go와 server_test.go로 인증·본문 정책·정확 일치 라우팅·로컬 응답·adapter 호출 경계를 구현했다. 제품 supervisor 또는 실제 provider adapter의 완료 판정은 아니다.

- 세션 헤더와 토큰, 본문 상한을 NewServer에 명시적으로 주입한다. 기본 인증 헤더나 생산 본문 상한은 정하지 않았다. 토큰 SHA-256 두 값을 constant-time으로 비교한 뒤에만 path나 본문을 읽는다.
- POST /v1/messages, POST /v1/messages/count_tokens, GET /v1/models만 허용한다. 앞의 둘은 query 없음 또는 beta=true 하나, models는 query 없음 또는 limit=1000 하나를 받는다. 다른 method/path/query는 명시 오류다.
- 압축 전후 크기를 같은 양수 상한으로 검사하며 gzip 이외 Content-Encoding, 잘못된 gzip/JSON, JSON 중복 키를 외부 adapter 호출 전에 거절한다.
- 검증 형태는 먼저 판정하고 catalog 부재 404, credential 부재 401, 존재 시 Anthropic Messages 최소 응답 200을 로컬에서 반환한다.
- 일반 요청은 선택된 entry의 credential Provider와 Generation을 확인하고 호출 직전 다시 Generation을 확인한다. 실패하면 다른 provider로 재시도하지 않는다.
- /v1/models는 data 배열 안에 세션 catalog의 id만 반환한다. 스키마와 limit=1000 근거는 부모가 조회한 공식 문서 https://code.claude.com/docs/en/llm-gateway-protocol 의 discovery 계약이다.
- count_tokens는 모든 provider에 messages/system/tools JSON의 rune 수를 4로 나눈 올림값을 준다. 응답에 accuracy=estimate, algorithm=json-runes-div4, provider를 명시하며 tokenizer 정확값을 주장하지 않는다. 로컬 추정이므로 credential을 요구하지 않으며 adapter를 호출하지 않는다.

## API 계약

`NewServer(ServerConfig) (*Server,error)`의 Server는 http.Handler다. Catalog, ResolveCredential, Adapters를 주입하며 adapter map은 복사한다. RoutedRequest에는 Entry, Body, Headers, Credential, Generation이 있다. Headers에는 Anthropic-Version과 Anthropic-Beta만 배열까지 복사한다. 그 키가 세션 인증 헤더와 같으면 제외한다. Authorization, x-api-key, 사용자 지정 세션 인증 헤더, 임의 헤더·URL·endpoint는 넘기지 않는다. Anthropic adapter가 protocol 헤더를 사용하고 Z.AI는 beta를 제거하며 OpenAI는 무시해야 한다.

`Adapter.Send(context.Context, RoutedRequest) (*http.Response,error)`는 고정 provider endpoint에만 송신해야 한다. adapter는 history 정규화 후 Credential.Apply를 호출하고 RoundTrip 직전 Generation을 다시 검사해야 한다. 서버에서 확인한 후 adapter 안에서 logout이 일어날 수 있으므로 이 의무는 생략할 수 없다. 이번 범위에는 구체 adapter가 없다.

응답은 상태와 Content-Type/Retry-After/Request-Id만 전달하며 body를 복사하고 flush한다. body 중단 뒤 성공 terminal event를 만들거나 재시도하지 않는다. 실제 SSE 변환·history 규칙은 다음 단계다.

## Evidence

RED·GREEN·coverage·race·vet 명령은 한 호출에서 아래 접두부를 썼다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
```

구현 전 `test ./internal/gateway/...` RED, exit 1:

```text
# github.com/modu-ai/moai-adk/internal/gateway [github.com/modu-ai/moai-adk/internal/gateway.test]
internal/gateway/server_test.go:23:92: undefined: RoutedRequest
internal/gateway/server_test.go:24:47: undefined: RoutedRequest
internal/gateway/server_test.go:25:79: undefined: Server
internal/gateway/server_test.go:25:183: undefined: NewServer
internal/gateway/server_test.go:25:193: undefined: ServerConfig
internal/gateway/server_test.go:25: undefined: Adapter
internal/gateway/server_test.go:26:17: undefined: Server
internal/gateway/server_test.go:43:72: undefined: ServerConfig
internal/gateway/server_test.go:43: undefined: NewServer
FAIL	github.com/modu-ai/moai-adk/internal/gateway [build failed]
?   	github.com/modu-ai/moai-adk/internal/gateway/auth	[no test files]
FAIL
```

첫 `test ./internal/gateway/...` GREEN, exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.423s
?   	github.com/modu-ai/moai-adk/internal/gateway/auth	[no test files]
```

protocol 헤더 보정 전 `test -coverpkg=./internal/gateway/... -coverprofile=/tmp/gateway-ingress-cover.out ./internal/gateway/...`, exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.369s	coverage: 96.5% of statements in ./internal/gateway/...
	github.com/modu-ai/moai-adk/internal/gateway/auth		coverage: 0.0% of statements
```

`go tool cover -func=/tmp/gateway-ingress-cover.out` 합산 profile에서 auth 메서드 4개, catalog 함수 4개, router credential 함수는 모두 100.0%다. ingress 함수의 실제 출력:

```text
github.com/modu-ai/moai-adk/internal/gateway/policy.go:14:		readBounded		100.0%
github.com/modu-ai/moai-adk/internal/gateway/policy.go:24:		readRequestBody		100.0%
github.com/modu-ai/moai-adk/internal/gateway/policy.go:44:		pathPolicy		94.7%
github.com/modu-ai/moai-adk/internal/gateway/router.go:29:		credential		100.0%
github.com/modu-ai/moai-adk/internal/gateway/server.go:35:		NewServer		90.9%
github.com/modu-ai/moai-adk/internal/gateway/server.go:56:		ServeHTTP		95.8%
github.com/modu-ai/moai-adk/internal/gateway/server.go:154:		Write			100.0%
github.com/modu-ai/moai-adk/internal/gateway/server.go:162:		models			100.0%
github.com/modu-ai/moai-adk/internal/gateway/server.go:169:		writeCount		100.0%
github.com/modu-ai/moai-adk/internal/gateway/server.go:180:		writeError		100.0%
github.com/modu-ai/moai-adk/internal/gateway/server.go:183:		writeJSON		100.0%
total:									(statements)		96.5%
```

`test -race ./internal/gateway/...`, exit 0 (최종 resolver nil/error 사례 두 개 추가 전; 제품 코드는 동일):

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.417s
?   	github.com/modu-ai/moai-adk/internal/gateway/auth	[no test files]
```

`vet ./internal/gateway/...`: exit 0, 출력 없음. 구현 후 `gopls check`로 server.go/router.go/policy.go를 검사했다: exit 0, 출력 없음. 구현 전 LSP 기준선은 없어 전후 차이의 주장은 하지 않는다.

테스트는 잘못된 인증 시 본문 Read 0회, 로컬 응답·잘못된 요청의 mock adapter 호출 0회, 선택된 provider의 상태 200/401/429/503 전달 및 호출 1회, concurrent provider 3종 요청별 선택, 요청 취소/credential 교체/없는 adapter 거절을 단언했다. 모든 검증은 프로세스 안에서 httptest.ResponseRecorder와 mock adapter로 수행했다. 외부 API 또는 Claude를 실행하지 않았다.

모델 집합의 추가·누락·중복 0건을 별도 단언하는 `test ./internal/gateway/... -run TestModelsContainsExactlySessionCatalog -count=1`도 위 환경 접두부로 실행했다. exit 0, 출력:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.402s
?   	github.com/modu-ai/moai-adk/internal/gateway/auth	[no test files]
```

### protocol 헤더 계약 보정과 최종 재검증

부모 요청에 따라 Anthropic-Version·Anthropic-Beta만 복사하는 allowlist를 추가했다. 헤더 이름이 세션 인증 헤더와 겹치는 경우는 제외한다. 원본 slice와의 alias가 없음을 검사했다.

위 환경 접두부로 `test ./internal/gateway/... -run 'TestAdapterReceivesOnlyClonedProtocolHeaders|TestSessionHeaderIsNeverForwardedEvenWhenProtocolNamed'`를 먼저 실행했다. RED exit 1:

```text
# github.com/modu-ai/moai-adk/internal/gateway [github.com/modu-ai/moai-adk/internal/gateway.test]
internal/gateway/server_test.go:368:13: in.Headers undefined (type RoutedRequest has no field or method Headers)
internal/gateway/server_test.go:369:49: in.Headers undefined (type RoutedRequest has no field or method Headers)
internal/gateway/server_test.go:372:10: in.Headers undefined (type RoutedRequest has no field or method Headers)
internal/gateway/server_test.go:376:6: in.Headers undefined (type RoutedRequest has no field or method Headers)
internal/gateway/server_test.go:387:8: r.Headers undefined (type RoutedRequest has no field or method Headers)
FAIL	github.com/modu-ai/moai-adk/internal/gateway [build failed]
?   	github.com/modu-ai/moai-adk/internal/gateway/auth	[no test files]
FAIL
```

같은 명령 GREEN exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.351s
?   	github.com/modu-ai/moai-adk/internal/gateway/auth	[no test files]
```

최종 `test -coverpkg=./internal/gateway/... -coverprofile=/tmp/gateway-ingress-cover.out ./internal/gateway/...`, exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.844s	coverage: 96.5% of statements in ./internal/gateway/...
	github.com/modu-ai/moai-adk/internal/gateway/auth		coverage: 0.0% of statements
```

최종 `test -race ./internal/gateway/...`, exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.525s
?   	github.com/modu-ai/moai-adk/internal/gateway/auth	[no test files]
```

최종 `vet ./internal/gateway/...`: exit 0, 출력 없음. 앞선 cover -func 행 번호는 보정 전 측정의 원문을 보존한 것이다.

## Baseline-attribution

WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, 브랜치 WT-unified-gateway, HEAD 81c1d58f9, SPEC 0.8.0. 앞서 작성한 foundation 위에서 실행했다. 새 ingress 파일 4개와 Run 기록만 변경했다. 커밋·push·상태 전이는 하지 않았다.

## Gaps

실제 socket listener·upstream 서버, Claude TUI, 모델 전환, provider credential 구체 구현, history/Responses 변환, SSE event 정합성, M0 OAuth, CLI/supervisor, 전체 AC, Windows는 이번 판정 밖이다. 실제 outbound 송신 직전 Generation 검사는 향후 adapter에서 시험해야 한다. 저장소 전체 시험 판정은 통합 브랜치 CI 소관이며 PENDING이다.

## Residual-risk

count_tokens는 JSON 표현에 따른 거친 추정이며 모델의 실제 토큰 수와 차이가 클 수 있다. client가 새로운 query·path를 요구하면 현행 정책은 거절하므로 그 실제 요청을 확인한 뒤 범위를 넓혀야 한다. data id만 제공하는 /v1/models가 bare GPT ID를 Claude picker에 보이게 한다는 주장은 하지 않는다. 부모가 조사한 공식 discovery 필터 한계는 PICKER 구현에서 따로 해결해야 한다.
