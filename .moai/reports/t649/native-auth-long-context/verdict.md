# t649 기존 인증 복원과 GPT 문맥 확장

## Claim

사용자 결정에 따라 gateway는 moai gpt에만 적용한다. moai cc와 moai glm은 기존 런처, 프로필, 인증 경로를 사용한다. GPT의 카탈로그 문맥 창은 872,000, 요청 본문 제한은 16 MiB로 조정했다. 기존 프로필에서 현재 작업 경로에 이미 기록된 신뢰 결정만 GPT의 새 대화 폴더에 계승한다.

## Evidence

```text
go test -p 1 ./internal/cli ./internal/gateway -run 'TestGateway|TestDefaultLaunch|TestUnifiedLaunch|TestLaunchClaude|TestApplyCC|TestApplyGLM|TestNativeGateway|TestCatalog|TestAnthropic' -count=1 -timeout=120s
ok  github.com/modu-ai/moai-adk/internal/cli  1.677s
ok  github.com/modu-ai/moai-adk/internal/gateway  0.514s

go test -p 1 ./internal/cli -run TestGatewayProductionLongContextBudget -count=1 -timeout=90s
ok  github.com/modu-ai/moai-adk/internal/cli  1.062s
```

- native-launch-fixture.json: 실제 빌드의 cc·glm 명령 모두 exit 0. cc base_url=null, glm base_url=https://api.z.ai/api/anthropic. 같은 기존 프로필 경로 사용, 인증·MCP fixture 파일 불변.
- gpt-launch-fixture.json: 실제 빌드의 gpt 명령 exit 0, 로컬 gateway 경로 유지, max_context=872000. 이 검증은 대역 Claude를 사용하며 추론 요청을 보내지 않았다.
- ../long-context-runtime/gpt-5.6-sol.json: 공식 App Server, 마지막 요청 inputTokens=833112, 출력 MAPLE71,OTTER83,CEDAR29, completed.
- ../long-context-runtime/gpt-6-astra.json: 공식 App Server, 마지막 요청 inputTokens=833282, 출력 MAPLE71,OTTER83,CEDAR29, completed.
- 두 실증은 합성 입력을 두 차례의 정상 대화로 누적했다. 단일 1,600,100자 입력은 App Server의 1,048,576자 제한으로 거절되었다. 여러 턴 누적은 서버 입력 규칙을 유지한다.

## Baseline-attribution

WT-unified-gateway, HEAD 81c1d58f9cf7045594ee61d5e4ff380948ce9eba와 기존 미커밋 소스. 정확한 수정 소스 SHA와 설치 바이너리 SHA는 build.json. 설치 경로 /Users/goos/go/bin/moai. 빌드 v3.2.0-rc.8-t649-native-auth-97684da3e8ce. 모델 캐시 fetched_at 2026-09-12T07:52:20.950154Z에서 GPT 네 모델 모두 기본 272000, 최대 872000을 읽었다.

## Gaps

실제 사용자의 Claude Max 로그인 화면 복구와 MCP 접속 개수는 아직 재관찰하지 않았다. 실제 모델 실증은 공식 App Server에서 수행했으며 설치된 기존 gateway의 Claude UI 전체 장문 경로 실증과 다르다. 전체 App Server 전환 완료를 주장하지 않는다. 1M 입력 성공, API 키 모드 1,050,000 창, Windows 실행은 검증하지 않았다.

## Residual-risk

구독의 선언 가능한 최대 창과 실효 문맥 창은 다르다. App Server는 872000 설정에서 modelContextWindow=828400을 보고했다. 95% 예약 정책이 반영된 값이다. 게이트웨이 입력 계산은 정확한 tokenizer가 아닌 추정이다. 서비스가 실제 최종 입력 허용 여부를 판정한다. 기존 실행 중 세션에는 새 바이너리와 환경 변경이 적용되지 않는다.
