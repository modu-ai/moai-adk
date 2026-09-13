# t649 GPT 인증 확인창과 문맥 경고 수정

- Claim: GPT 런처가 생성한 로컬 세션 토큰을 ANTHROPIC_API_KEY 대신 ANTHROPIC_AUTH_TOKEN으로 전달한다. Gateway는 Authorization: Bearer 값을 검증하고 upstream에는 전달하지 않는다. GPT 자식에서 CLAUDE_CODE_DISABLE_1M_CONTEXT를 제거하며 872000 선언은 유지한다. cc·glm은 앞서 복원한 기존 경로를 유지한다.
- Evidence: `go test -p 1 ./internal/cli ./internal/gateway -run 'TestGateway|TestGPTBinding|TestProductionGateway|TestNativeGateway|TestDefaultLaunch' -count=1 -timeout=120s` 출력:

```text
ok  github.com/modu-ai/moai-adk/internal/cli  9.455s
ok  github.com/modu-ai/moai-adk/internal/gateway  0.543s [no tests to run]
```

- Evidence: `python3 .moai/reports/t649/gpt-bearer-launch/interactive_probe.py`의 실제 Claude 대화형 화면 결과: api_key_prompt=false, context_warning=false, login_prompt=false, reply_seen=true, success=true. 입력한 합성 문구 요청에 GPT가 MOAI_BEARER_READY를 답변했다. 원문 JSON과 정확한 명령은 interactive-result.json 참조. 이 실행은 사용자 프로젝트의 mo.ai.kr 프로필을 사용하되 도구·훅·MCP는 테스트 인수로 비활성화했다.
- Baseline-attribution: WT-unified-gateway, HEAD 81c1d58f9와 기존 미커밋 변경. build.json에 수정 소스 해시, 바이너리 해시, 설치 후 version 출력과 백업 경로를 기록했다.
- Gaps: 이 실행에서 전체 MCP/도구 흐름, Windows, 1M 실제 입력 용량을 검증하지 않았다. gateway 패키지는 위 정규식에 일치하는 테스트가 없어 해당 행을 패키지 기능 검증 성공으로 세지 않는다.
- Residual-risk: 872k는 현재 구독 모델 정보의 최대 설정값이며 1M 지원을 의미하지 않는다. 기존 실행 중 세션에는 변경된 환경이 반영되지 않으므로 종료 후 새로 실행해야 한다. 개인 Claude/Anthropic 인증 정보나 승인 캐시는 수정하지 않았다.
