# M6 native Messages adapter 로컬 검증

## Claim

GLM과 Anthropic API 키 전용 native Messages 중계를 구현했다. 공개 이력과 도구 pair, 고정 endpoint·credential·헤더,
SSE 순서·오류·EOF·취소를 합성 입력과 로컬 TLS 서버로 검사했다. 실제 공급자 수용·Claude Code 호환성·전체 M6/코어 완료 판정은 아니다.
OAuth passthrough 구체 타입·전달 코드·catalog 항목을 만들지 않았다.

- `NewAnthropicAdapter(MessagesConfig)`, `NewGLMAdapter(MessagesConfig)`는 shared `MessagesAdapter`를 만든다.
  provider와 AuthMethod를 정확히 판정하며 각각 `https://api.anthropic.com/v1/messages`,
  `https://api.z.ai/api/anthropic/v1/messages`에만 POST한다. GLM 주소는 기존 `internal/config/defaults.go:136`의 기본 base와
  `internal/cli/glm_task_test.go:206`의 Messages 경로를 읽어 확인했다. API 키/구독 fallback과 redirect는 없다.
- 설정에 명시한 Anthropic-Version과 beta 허용 목록만 사용한다. GLM은 beta를 모두 제거한다. 입력 인증/임의 헤더를 복사하지 않는다.
- `auth.NewAnthropicAPIKey(explicitKey)`는 X-Api-Key만 적용하며 Authorization을 제거한다. 환경/OAuth를 읽지 않는다.
- `auth.NewGLMCredential()`은 기존 `glmcred.Load`를 사용한다. 임시 MOAI_HOME의 `glmcred.Save`와 실제 저장소 reader로
  달러·역슬래시 포함 키 왕복과 키 변경 거절을 시험했다. 상속 Z_AI_API_KEY는 대체 인증원으로 사용하지 않는다.
- Native Messages를 OpenAI 형태로 바꾸지 않는다. system 메시지 위치, stop_sequences, `tool_result.is_error: true`를 보존한다.
  선택된 canonical model만 바꾼다. 공개 text/tool pair를 검사하며 미완료 pair·미지원 이미지/PDF·opaque content를 거절한다.
- Native SSE Reader는 converter goroutine을 만들지 않는다. 제한된 이벤트 하나를 검증한 뒤 원래 이벤트를 반환한다.
  시작/블록/최종 delta/종료 순서, 교차 tool index·ID·JSON 인수를 확인하며 오류·중간 EOF에 성공 terminal을 합성하지 않는다.
  Body.Close와 context 취소는 upstream을 닫는다. upstream 오류 본문·민감 문자열은 반영하지 않고 상태와 검증된 Retry-After만 보존한다.

## Evidence

모든 실행의 작업 디렉터리는 아래 WT다. 새 credential 시험을 구현 전에 실행한 RED:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway/auth -run 'TestAnthropicExplicit|TestGLMStored' -count=1
# github.com/modu-ai/moai-adk/internal/gateway/auth [github.com/modu-ai/moai-adk/internal/gateway/auth.test]
internal/gateway/auth/credential_anthropic_test.go:12:14: undefined: NewAnthropicAPIKey
internal/gateway/auth/credential_anthropic_test.go:16:10: undefined: NewAnthropicAPIKey
internal/gateway/auth/credential_anthropic_test.go:26:36: undefined: AnthropicEndpoint
internal/gateway/auth/credential_glm_test.go:15:13: undefined: NewGLMCredential
internal/gateway/auth/credential_glm_test.go:21:12: undefined: NewGLMCredential
internal/gateway/auth/credential_glm_test.go:28:36: undefined: GLMEndpoint
internal/gateway/auth/credential_glm_test.go:35:38: undefined: AnthropicEndpoint
internal/gateway/auth/credential_glm_test.go:45:35: undefined: GLMEndpoint
internal/gateway/auth/credential_glm_test.go:49:13: undefined: NewGLMCredential
FAIL	github.com/modu-ai/moai-adk/internal/gateway/auth [build failed]
FAIL
```

Exit 1. 구현 뒤 같은 시험의 첫 GREEN:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	0.373s
```

Native adapter 구현 전 RED:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway -run 'TestAnthropicNative|TestGLMNative' -count=1
# github.com/modu-ai/moai-adk/internal/gateway [github.com/modu-ai/moai-adk/internal/gateway.test]
internal/gateway/anthropic_test.go:22:39: undefined: MessagesConfig
internal/gateway/anthropic_test.go:23:9: undefined: MessagesConfig
internal/gateway/anthropic_test.go:64:10: undefined: NewAnthropicAdapter
internal/gateway/anthropic_test.go:108:11: undefined: NewAnthropicAdapter
internal/gateway/anthropic_test.go:150:11: undefined: NewAnthropicAdapter
internal/gateway/glm_test.go:28:10: undefined: NewGLMAdapter
FAIL	github.com/modu-ai/moai-adk/internal/gateway [build failed]
FAIL
```

Exit 1. 구현 뒤 첫 native GREEN:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway -run 'TestAnthropicNative|TestGLMNative' -count=1 -timeout 30s
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.489s
```

공유 JSON validator wrapper도 구현 전 undefined RED를 관측했다.

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway/translate -run TestValidateJSONObjectPublicBoundary -count=1
# github.com/modu-ai/moai-adk/internal/gateway/translate [github.com/modu-ai/moai-adk/internal/gateway/translate.test]
internal/gateway/translate/request_test.go:203:11: undefined: ValidateJSONObject
internal/gateway/translate/request_test.go:208:11: undefined: ValidateJSONObject
FAIL	github.com/modu-ai/moai-adk/internal/gateway/translate [build failed]
FAIL
```

Exit 1. wrapper 추가 뒤 같은 시험: `ok github.com/modu-ai/moai-adk/internal/gateway/translate 0.397s`, exit 0.

최종 독립 검사들을 한 묶음으로 실행했고, 이후 교차 도구 SSE·읽기 중 취소 시험을 추가하여 해당 root 시험과 vet를 다시 실행했다.

```text
$ GOCACHE=/tmp/gateway-translation-cache go test -race ./internal/gateway -run 'TestAnthropicNative|TestGLMNative' -count=1 -timeout 30s -coverprofile=/tmp/gateway-m6-native-cover.out
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.540s	coverage: 35.1% of statements
$ GOCACHE=/tmp/gateway-translation-cache go test -race ./internal/gateway/auth -run 'TestAnthropicExplicit|TestGLMStored' -count=1 -coverprofile=/tmp/gateway-m6-credential-cover.out
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	1.652s	coverage: 5.7% of statements
$ GOCACHE=/tmp/gateway-translation-cache go test -race ./internal/gateway/translate -count=1
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.503s
$ GOCACHE=/tmp/gateway-translation-cache go vet ./internal/gateway ./internal/gateway/auth ./internal/gateway/translate
$ GOCACHE=/tmp/gateway-translation-cache gopls check internal/gateway/anthropic.go internal/gateway/glm.go internal/gateway/auth/credential_anthropic.go internal/gateway/auth/credential_glm.go internal/gateway/translate/request.go
```

각 exit 0. vet/gopls stdout·stderr는 비었다. root/auth 수치는 선택한 M6 시험만 실행했을 때의 패키지 전체 수치다.
coverage 파일의 해당 제품 파일 statement/count를 합산한 실제 출력은 다음과 같다.

```text
anthropic.go: 318/349 (91.1%)
glm.go: 1/1 (100.0%)
credential_anthropic.go: 13/14 (92.9%)
credential_glm.go: 21/22 (95.5%)
```

네 제품 파일 각각 85% 이상이다. loopback TLS listener 실행은 승인 후 수행했다. 실제 Claude/provider/account는 호출하지 않았다.

## Baseline-attribution

- WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`; 종료 전 `git rev-parse --short HEAD` → `81c1d58f9`, exit 0.
- 읽은 계약: m6-local-adapters-brief, 현재 plan M6, design ingress·adapter·passthrough gate, REQ-MG-011~018, 대응 acceptance,
  기존 glmcred writer/reader와 테스트 seam. 다른 작업자들의 미커밋 변경과 함께 있는 트리다.
- 새 제품 SHA-256:
  - anthropic.go: `bbbb64f1c94e8a1aef0eba76aeb43e451bae4af29dad0ca11e43b3e6e7d472e4`
  - glm.go: `21cf027f874217154bd59967e073f84e9c08a94eb9bcef54d5a6b50c99ef3d31`
  - credential_anthropic.go: `6ff19f8ab3f866633d8da246c5083d6f251cead5651f1fc4271b06d0853cdb14`
  - credential_glm.go: `1b44b0e210ec8c39f7625904816cd60a0c742d831d5fbf63259acae511cf2841`
- 추가로 승인받은 translate.ValidateJSONObject wrapper는 기존 objectJSON에 위임하며 Responses 정책 검사를 하지 않는다.
  기존 root gateway.ValidateJSONObject와 ingress 구조는 변경하지 않았다. 현 translate/request.go SHA-256은
  `075cc45aea7051c310f08b5b2709271ae078f743aac2bb14fef212d136a7b59d`다.
- 기존 auth broker/store/send/ownership, supervisor, CLI/factory, SPEC, Git 이력은 변경하지 않았다.

## Gaps

- 일반 native public content 범위다. Thinking enabled/adaptive, context_management/output_config, 이미지/PDF, opaque 이력은
  명시 거절한다. thinking disabled만 native 정책으로 보존한다. 같은/다른 provider reasoning 보존·전환은 별도 게이트다.
- 실제 Claude의 system-in-messages 수용, 실제 GLM/Anthropic 도구·SSE·과금·모델 접근을 측정하지 않았다.
  현재 public fixture 성공을 기존 GLM UX 전체 보존이나 전체 M6 완료로 확대하지 않는다.
- ContextTokens는 catalog 값과 caller의 합성 로컬 측정값으로 시험했다. 실제 tokenizer/측정 factory 연결은 없다.
- GLM Generation은 immutable ref의 키와 현재 저장소 키를 비교해 변화 시 ErrCredentialChanged를 내며, 유효할 때 1이다.
  마지막 재검사 뒤 concurrent Save와 실제 socket write 사이의 원자 barrier는 없다. 키를 바꿨다 되돌리는 ABA 방지도 없다.
  저장소 파일의 세대 journal이나 writer lock을 도입하지 않았다. 키 회전 음성군은 순차적 변경의 거절을 검증했으며 원자 송신 증명이 아니다.
- 기존 glmcred.Load의 documented MOAI_TEST_GLM_KEY seam은 그대로 존재한다. 실제 자식 실행 환경 scrub는 이 adapter 작업 범위 밖이다.
- OAuth passthrough는 M0 양성 전 금지 상태다. 구체 OAuth 타입·코드·catalog 항목은 없으며 API 키 검사만 수행했다.
- 변경 전 별도 LSP snapshot은 없고 현재 제품 파일 진단만 관측했다. 저장소 전체 suite·Windows·통합 CI는 수행하지 않았다.
  저장소 전체 verdict 담당은 통합 브랜치 CI이며 PENDING이다. Commit/push/PR/merge/factory 연결은 하지 않았다.

## Residual-risk

합성 TLS와 Messages 이벤트의 통과는 실제 upstream의 추가 필드·정책·반환 순서를 보장하지 않는다. 지원 범위 밖 출력은 오류가 된다.
호출자는 HTTP body를 끝까지 읽거나 Close해야 하며 caller context 전파가 필요하다. GLM 저장소 읽기와 전송의 race/ABA 공백은
별도 송신 장벽 계약 없이는 보안 완료로 표시할 수 없다. 전체 사용자 목표는 실제 정책·reasoning·인증·CLI 통합 시험이 끝나야 판정할 수 있다.
