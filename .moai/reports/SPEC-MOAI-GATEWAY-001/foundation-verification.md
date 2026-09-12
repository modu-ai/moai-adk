# M1 인식기·M2 기반 구현 검증

## Claim

0.8.0의 캡처 기반 인식기와 모델 registry·CredentialRef seam을 구현했다. 실제 캡처 5건은 익명화 후 internal/gateway/testdata/validation-capture/claude-code-2.1.268에 저장했다. stream 부재 또는 JSON false, 정수 1, user 메시지 한 개, tools 부재 또는 빈 배열을 동시에 요구한다. 중복 키는 모든 깊이에서 오류로 거절한다.

CatalogSnapshot은 입력과 반환 목록을 복사한다. 네 GPT full ID, claude-opus-5·claude-sonnet-5, 설정에서 받은 GLM tier ID를 정확 일치로 찾는다. 같은 GLM ID는 한 항목으로 묶고 충돌하는 등록은 거절한다. Capabilities는 아직 검증된 값이 없어 0값이며 지원을 선언하지 않는다. OAuth passthrough 구체 타입은 없다.

## Evidence

Go test·vet 검증 명령은 아래 한 호출의 환경 정리를 앞에 두고 실행했다. GOCACHE는 쓰기 가능한 임시 경로다. 최초 기본 cache 접근 오류는 RED로 세지 않았다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified test ./internal/gateway/...
```

구현 전 RED (exit 1):

```text
FAIL	github.com/modu-ai/moai-adk/internal/gateway [setup failed]
# github.com/modu-ai/moai-adk/internal/gateway
internal/gateway/catalog_test.go:6:2: no required module provides package github.com/modu-ai/moai-adk/internal/gateway/auth; to add it:
	go get github.com/modu-ai/moai-adk/internal/gateway/auth
FAIL
```

GREEN (exit 0):

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.403s
?   	github.com/modu-ai/moai-adk/internal/gateway/auth	[no test files]
```

같은 환경 정리로 `go -C <WT> test -coverpkg=./internal/gateway/... -coverprofile=/tmp/gateway-foundation-cover.out ./internal/gateway/...` (exit 0):

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.405s	coverage: 96.2% of statements in ./internal/gateway/...
	github.com/modu-ai/moai-adk/internal/gateway/auth		coverage: 0.0% of statements
```

`go -C <WT> tool cover -func=/tmp/gateway-foundation-cover.out`의 합산 결과:

```text
github.com/modu-ai/moai-adk/internal/gateway/auth/credential.go:35:	Provider		100.0%
github.com/modu-ai/moai-adk/internal/gateway/auth/credential.go:36:	Generation		100.0%
github.com/modu-ai/moai-adk/internal/gateway/auth/credential.go:37:	Apply			100.0%
github.com/modu-ai/moai-adk/internal/gateway/auth/credential.go:38:	Redacted		100.0%
github.com/modu-ai/moai-adk/internal/gateway/catalog.go:67:		NewCatalog		100.0%
github.com/modu-ai/moai-adk/internal/gateway/catalog.go:92:		NewSessionCatalog	100.0%
github.com/modu-ai/moai-adk/internal/gateway/catalog.go:106:		Resolve			100.0%
github.com/modu-ai/moai-adk/internal/gateway/catalog.go:113:		Entries			100.0%
github.com/modu-ai/moai-adk/internal/gateway/validation.go:16:		IsValidationRequest	100.0%
github.com/modu-ai/moai-adk/internal/gateway/validation.go:54:		uniqueJSONValue		87.5%
total:									(statements)		96.2%
```

auth 자체에는 별도 테스트 파일이 없으며 gateway 테스트가 auth 공개 계약을 실행한다. 따라서 합산 cover profile을 기준으로 판정한다.

같은 환경 정리로 `go -C <WT> vet ./internal/gateway/...`: exit 0, 출력 없음.

같은 환경 정리로 `go -C <WT> test -race ./internal/gateway/...` (exit 0):

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.427s
?   	github.com/modu-ai/moai-adk/internal/gateway/auth	[no test files]
```

Python 재귀 비교: 원본 SHA-256, 선언된 변경 경로와 실제 diff 일치, 객체 키·배열 길이·JSON 타입 보존 단언. 출력:

```text
PASS: all 5 original SHA-256 hashes unchanged; recursive diff equals declared changed_fields; keys, JSON types and array lengths preserved
```

구현 후 `gopls check`로 catalog.go, validation.go, auth/credential.go를 검사했다: exit 0, 출력 없음. 구현 전 LSP baseline은 없으므로 전후 차이 주장은 하지 않는다. import 그룹 정리 전 coverage 행 번호를 위 원문에 보존했다.

## Baseline-attribution

이번 실행의 WT는 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, HEAD는 `81c1d58f9`다. 기존 미추적 SPEC·보고서를 보존했다. 코드·픽스처·테스트는 새 `internal/gateway/`에만 추가했다. 커밋하지 않았으므로 SPEC status를 변경하지 않았다.

## Gaps

M1 네 번 모델 전환, HTTP local response와 송신 계수, 실제 gateway 실행, M0 OAuth 정상 응답·refresh, 모든 AC, Windows 검증은 미완료다. LSP 구현 전 기준선은 캡처하지 않았다. API-key·GLM·GPT 실제 credential 구현과 요청 직전 Generation 재검사 실행부도 다음 단계 소관이다. 저장소 전체 시험 판정은 통합 브랜치 CI 소관이며 PENDING이다.

## Residual-risk

인식기에는 HTTP method/path/본문 크기 제한이 포함되지 않는다. 호출하는 HTTP 계층이 먼저 적용해야 한다. CredentialRef는 provider 목적지 검증 의무를 정의하지만 인터페이스 자체로 강제하지 않으며 향후 구체 구현에서 시험해야 한다. 모델의 실제 upstream 가용성과 capability를 이번 mock 단위 시험으로 확인한 것은 아니다. 역사적 캡처의 Sonnet 4.5 모델명은 원문 관측값이며 후속 실제 시험에는 Opus 5·Sonnet 5를 사용해야 한다.
