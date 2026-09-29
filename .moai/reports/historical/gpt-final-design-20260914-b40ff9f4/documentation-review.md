# 공식 문서 조사 장부 — 2026-09-14

## Claim·범위

지정 5개 페이지와 페이지 내 모든 하위 목차를 읽었다. LLM gateway 계열의 연결·배포·프로토콜 문서는 모두 포함한다. 5개 페이지에서 추출한 Markdown/HTML 직접 링크는 앵커별 중복 제거 기준 102개다. 사이트 전체를 재귀적으로 읽은 것은 아니며, NOT_READ는 조사 완료로 간주하지 않는다. 이미지·동일 페이지 앵커·문서 프레임의 전체 사이트 nav는 별도 하위 문서로 세지 않았다.

## Evidence·Baseline

공식 URL 끝에 `.md`를 붙여 `curl -L --fail --silent --show-error --max-time 40 <URL>`로 읽었다. 아래 5개 원문은 exit 0이고 마지막까지 분할 판독했다. 원문을 보고서에 통째로 복제하지 않고 목차와 참조 목록만 기록한다. chars는 JavaScript string length, lines는 LF 분리 개수다. 아래 상태는 내용 판독 범위이며 웹 서비스가 변하지 않는다는 보증이 아니다.

| 문서 | chars | lines | 목차 수 | 상태 |
|---|---:|---:|---:|---|
| [app-server](https://learn.chatgpt.com/ko-KR/docs/app-server) | 80211 | 2342 | 80 | FULL |
| [llm-gateway](https://code.claude.com/docs/ko/llm-gateway) | 4164 | 65 | 4 | FULL |
| [llm-gateway-connect](https://code.claude.com/docs/ko/llm-gateway-connect) | 42312 | 610 | 29 | FULL |
| [llm-gateway-rollout](https://code.claude.com/docs/ko/llm-gateway-rollout) | 23426 | 301 | 13 | FULL |
| [llm-gateway-protocol](https://code.claude.com/docs/ko/llm-gateway-protocol) | 24737 | 269 | 16 | FULL |

## 직접 링크의 판독 상태

- NOT_READ: 71개 URL
- FULL: 21개 URL
- SECTION: 9개 URL
- GAP: 1개 URL

숫자는 읽은 문서 수가 아니라 직접 링크(앵커 포함) 수다. 같은 페이지의 서로 다른 앵커가 각각 포함된다.

## app-server — 전체 하위 목차

- CLI 터미널 UI 연결
- 원격 Code Mode 호스트 연결
- 프로토콜
- 메시지 스키마
- 시작하기
- 핵심 구성 요소
- 수명 주기 개요
- 초기화
- 실험적 API 사용 설정
- API 개요
- 모델
- 모델 목록 조회(`model/list`)
- 실험적 기능 목록 조회(`experimentalFeature/list`)
- 실행 환경 확인(실험적)
- 스레드
- 스레드 시작 또는 재개
- 스레드 목표 관리
- 저장된 스레드 읽기(재개 없이)
- 스레드 턴 목록 조회
- 스레드 목록 조회(페이지네이션 및 필터)
- 저장된 스레드 메타데이터 업데이트
- 스레드 상태 변경 추적
- 로드된 스레드 목록 조회
- 로드된 스레드 구독 해제
- 스레드 보관
- 스레드 삭제
- 스레드 보관 해제
- 스레드 컨텍스트 압축 실행
- 스레드 셸 명령어 실행
- 백그라운드 터미널 정리
- 최근 턴 롤백
- 턴
- 샌드박스 읽기 권한(`ReadOnlyAccess`)
- 턴 시작
- 스레드에 항목 삽입
- 활성 턴 조정
- 턴 시작(스킬 호출)
- 턴 중단
- 검토
- 프로세스 실행
- 명령어 실행
- 관리자 요구 사항 조회(`configRequirements/read`)
- Windows 샌드박스 설정(`windowsSandbox/setupStart`)
- 파일 시스템
- 이벤트
- 알림 수신 거부
- 퍼지 파일 검색 이벤트(실험적)
- 경고 이벤트
- Windows 샌드박스 설정 이벤트
- 턴 이벤트
- 항목
- 항목 델타
- 오류
- 승인
- 명령어 실행 승인
- 파일 변경 승인
- `tool/requestUserInput`
- 권한 요청
- MCP 서버의 유도 요청
- 동적 도구 호출(실험적)
- MCP 도구 호출 승인(앱)
- 스킬
- 앱(커넥터)
- 앱 설정을 위한 구성 RPC 예제
- 외부 에이전트 설정 감지 및 가져오기
- 인증 엔드포인트
- 인증 모드
- API 개요
- 1) 인증 상태 확인
- 2) API 키로 로그인
- 3) ChatGPT로 로그인(브라우저 플로우)
- 3b) ChatGPT로 로그인(기기 코드 플로우)
- 3c) 외부에서 관리하는 ChatGPT 토큰으로 로그인(`chatgptAuthTokens`)
- 4) ChatGPT 로그인 취소
- 5) 로그아웃
- 6) 요청 한도(ChatGPT)
- 7) 토큰 사용량(ChatGPT)
- 8) 획득한 요청 한도 재설정 기회(ChatGPT)
- 9) 워크스페이스 소유자에게 한도 알림 보내기
- 10) 워크스페이스 메시지(ChatGPT)

## llm-gateway — 전체 하위 목차

- [gateway가 제공하는 것](https://code.claude.com/docs/ko/llm-gateway#what-a-gateway-provides)
- [gateway 배포](https://code.claude.com/docs/ko/llm-gateway#roll-out-a-gateway)
- [구독 및 gateway](https://code.claude.com/docs/ko/llm-gateway#subscriptions-and-gateways)
- [관련 페이지](https://code.claude.com/docs/ko/llm-gateway#related-pages)

## llm-gateway-connect — 전체 하위 목차

- [기존 구성 확인](https://code.claude.com/docs/ko/llm-gateway-connect#check-for-an-existing-configuration)
- [Claude Code 직접 구성](https://code.claude.com/docs/ko/llm-gateway-connect#configure-claude-code-yourself)
- [자격 증명 변수 설정](https://code.claude.com/docs/ko/llm-gateway-connect#set-the-credential-variable)
- [기본 URL과 자격 증명 설정](https://code.claude.com/docs/ko/llm-gateway-connect#set-the-base-url-and-credential)
- [셸 환경 변수로 설정](https://code.claude.com/docs/ko/llm-gateway-connect#set-as-shell-environment-variables)
- [설정 파일에서 설정](https://code.claude.com/docs/ko/llm-gateway-connect#set-in-a-settings-file)
- [연결 확인](https://code.claude.com/docs/ko/llm-gateway-connect#verify-the-connection)
- [Claude Code에서 확인](https://code.claude.com/docs/ko/llm-gateway-connect#confirm-in-claude-code)
- [자격 증명 변수가 헤더에 매핑되는 방식](https://code.claude.com/docs/ko/llm-gateway-connect#how-the-credential-variable-maps-to-a-header)
- [기존 로그인과의 충돌](https://code.claude.com/docs/ko/llm-gateway-connect#conflicts-with-an-existing-login)
- [각 표면 구성](https://code.claude.com/docs/ko/llm-gateway-connect#configure-each-surface)
- [VS Code 확장](https://code.claude.com/docs/ko/llm-gateway-connect#vs-code-extension)
- [데스크톱 앱](https://code.claude.com/docs/ko/llm-gateway-connect#desktop-app)
- [GitHub Actions](https://code.claude.com/docs/ko/llm-gateway-connect#github-actions)
- [Agent SDK](https://code.claude.com/docs/ko/llm-gateway-connect#agent-sdk)
- [Slack, 웹 및 Remote Control](https://code.claude.com/docs/ko/llm-gateway-connect#slack-web-and-remote-control)
- [추가 구성](https://code.claude.com/docs/ko/llm-gateway-connect#additional-configuration)
- [추가 헤더 전송](https://code.claude.com/docs/ko/llm-gateway-connect#send-additional-headers)
- [게이트웨이 모델을 모델 선택기에 추가](https://code.claude.com/docs/ko/llm-gateway-connect#add-gateway-models-to-the-model-picker)
- [apiKeyHelper로 자격 증명 회전](https://code.claude.com/docs/ko/llm-gateway-connect#rotate-credentials-with-apikeyhelper)
- [게이트웨이 경로 외부의 트래픽 끄기](https://code.claude.com/docs/ko/llm-gateway-connect#turn-off-traffic-outside-the-gateway-path)
- [게이트웨이를 통해 클라우드 제공자로 라우팅](https://code.claude.com/docs/ko/llm-gateway-connect#route-to-a-cloud-provider-through-a-gateway)
- [Amazon Bedrock](https://code.claude.com/docs/ko/llm-gateway-connect#amazon-bedrock)
- [Google Cloud의 Agent Platform](https://code.claude.com/docs/ko/llm-gateway-connect#google-cloud’s-agent-platform)
- [Microsoft Foundry](https://code.claude.com/docs/ko/llm-gateway-connect#microsoft-foundry)
- [AWS의 Claude Platform](https://code.claude.com/docs/ko/llm-gateway-connect#claude-platform-on-aws)
- [제공자 경로 확인](https://code.claude.com/docs/ko/llm-gateway-connect#confirm-the-provider-route)
- [게이트웨이 오류 문제 해결](https://code.claude.com/docs/ko/llm-gateway-connect#troubleshoot-gateway-errors)
- [관련 리소스](https://code.claude.com/docs/ko/llm-gateway-connect#related-resources)

## llm-gateway-rollout — 전체 하위 목차

- [필수 조건](https://code.claude.com/docs/ko/llm-gateway-rollout#prerequisites)
- [게이트웨이 요구사항](https://code.claude.com/docs/ko/llm-gateway-rollout#gateway-requirements)
- [롤아웃 단계](https://code.claude.com/docs/ko/llm-gateway-rollout#rollout-steps)
- [게이트웨이가 모델을 라우팅하는지 확인](https://code.claude.com/docs/ko/llm-gateway-rollout#confirm-the-gateway-routes-your-models)
- [개발자 자격증명 발급](https://code.claude.com/docs/ko/llm-gateway-rollout#issue-developer-credentials)
- [게이트웨이에 대해 Claude Code 테스트](https://code.claude.com/docs/ko/llm-gateway-rollout#test-claude-code-against-the-gateway)
- [구성 배포](https://code.claude.com/docs/ko/llm-gateway-rollout#distribute-the-configuration)
- [배포할 항목](https://code.claude.com/docs/ko/llm-gateway-rollout#what-to-distribute)
- [관리되는 설정을 통해 배포](https://code.claude.com/docs/ko/llm-gateway-rollout#distribute-through-managed-settings)
- [개발자에게 값을 직접 설정하도록 합니다.](https://code.claude.com/docs/ko/llm-gateway-rollout#hand-developers-the-values-to-set-themselves)
- [롤아웃 확인](https://code.claude.com/docs/ko/llm-gateway-rollout#verify-the-rollout)
- [게이트웨이 유지 관리](https://code.claude.com/docs/ko/llm-gateway-rollout#maintain-the-gateway)
- [관련 리소스](https://code.claude.com/docs/ko/llm-gateway-rollout#related-resources)

## llm-gateway-protocol — 전체 하위 목차

- [API 형식](https://code.claude.com/docs/ko/llm-gateway-protocol#api-formats)
- [Foundry 및 AWS의 Claude Platform](https://code.claude.com/docs/ko/llm-gateway-protocol#foundry-and-claude-platform-on-aws)
- [선택 사항 엔드포인트 및 시작 트래픽](https://code.claude.com/docs/ko/llm-gateway-protocol#optional-endpoints-and-startup-traffic)
- [스트리밍](https://code.claude.com/docs/ko/llm-gateway-protocol#streaming)
- [업스트림과의 형식 불일치](https://code.claude.com/docs/ko/llm-gateway-protocol#format-mismatch-with-the-upstream)
- [요청 헤더](https://code.claude.com/docs/ko/llm-gateway-protocol#request-headers)
- [개방형 목록으로 전달](https://code.claude.com/docs/ko/llm-gateway-protocol#forward-as-open-lists)
- [시스템 프롬프트 속성 블록](https://code.claude.com/docs/ko/llm-gateway-protocol#system-prompt-attribution-block)
- [기능 통과](https://code.claude.com/docs/ko/llm-gateway-protocol#feature-pass-through)
- [자동 재시도 및 오류 전달](https://code.claude.com/docs/ko/llm-gateway-protocol#automatic-retry-and-error-forwarding)
- [사전 릴리스 기능 비활성화](https://code.claude.com/docs/ko/llm-gateway-protocol#disable-pre-release-capabilities)
- [모델 검색](https://code.claude.com/docs/ko/llm-gateway-protocol#model-discovery)
- [검색이 실행되는 경우](https://code.claude.com/docs/ko/llm-gateway-protocol#when-discovery-runs)
- [요청 및 응답](https://code.claude.com/docs/ko/llm-gateway-protocol#request-and-response)
- [선택기 항목 및 캐싱](https://code.claude.com/docs/ko/llm-gateway-protocol#picker-entries-and-caching)
- [관련 리소스](https://code.claude.com/docs/ko/llm-gateway-protocol#related-resources)

## 연결 참조 102개 — 생략 없는 목록

| URL | 참조한 지정 문서 | 판독 상태 |
|---|---|---|
| [github.com/openai/codex/tree/main/codex-rs/app-server](https://github.com/openai/codex/tree/main/codex-rs/app-server) | app-server | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [learn.chatgpt.com/ko-KR/codex/open-source](https://learn.chatgpt.com/ko-KR/codex/open-source) | app-server | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [learn.chatgpt.com/codex/codex-sdk](https://learn.chatgpt.com/codex/codex-sdk) | app-server | FULL — SDK 전체 |
| [modelcontextprotocol.io/](https://modelcontextprotocol.io/) | app-server | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [chatgpt.com/public/admin/api-reference#tag/Codex](https://chatgpt.com/public/admin/api-reference#tag/Codex) | app-server | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [learn.chatgpt.com/ko-KR/codex/config-file/config-reference#requirementstoml](https://learn.chatgpt.com/ko-KR/codex/config-file/config-reference#requirementstoml) | app-server | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/claude-apps-gateway](https://code.claude.com/docs/ko/claude-apps-gateway) | llm-gateway, llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/gateways](https://code.claude.com/docs/ko/gateways) | llm-gateway, llm-gateway-protocol | SECTION — gateway 종류·구독·별도 설정 |
| [code.claude.com/docs/ko/llm-gateway-connect](https://code.claude.com/docs/ko/llm-gateway-connect) | llm-gateway, llm-gateway-rollout, llm-gateway-protocol | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/llm-gateway-rollout](https://code.claude.com/docs/ko/llm-gateway-rollout) | llm-gateway, llm-gateway-connect, llm-gateway-protocol | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/llm-gateway-protocol](https://code.claude.com/docs/ko/llm-gateway-protocol) | llm-gateway, llm-gateway-connect, llm-gateway-rollout | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/llm-gateway-protocol#api-formats](https://code.claude.com/docs/ko/llm-gateway-protocol#api-formats) | llm-gateway, llm-gateway-rollout | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/third-party-integrations](https://code.claude.com/docs/ko/third-party-integrations) | llm-gateway | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/managed-settings#delivery-mechanisms](https://code.claude.com/docs/ko/managed-settings#delivery-mechanisms) | llm-gateway, llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/llm-gateway-connect#check-for-an-existing-configuration](https://code.claude.com/docs/ko/llm-gateway-connect#check-for-an-existing-configuration) | llm-gateway | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/admin-setup](https://code.claude.com/docs/ko/admin-setup) | llm-gateway, llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/llm-gateway-connect#set-the-credential-variable](https://code.claude.com/docs/ko/llm-gateway-connect#set-the-credential-variable) | llm-gateway, llm-gateway-rollout, llm-gateway-protocol | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/llm-gateway-connect#set-the-base-url-and-credential](https://code.claude.com/docs/ko/llm-gateway-connect#set-the-base-url-and-credential) | llm-gateway | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/llm-gateway-protocol#request-headers](https://code.claude.com/docs/ko/llm-gateway-protocol#request-headers) | llm-gateway, llm-gateway-rollout | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/llm-gateway](https://code.claude.com/docs/ko/llm-gateway) | llm-gateway-connect, llm-gateway-protocol | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/managed-settings](https://code.claude.com/docs/ko/managed-settings) | llm-gateway-connect, llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/agent-view#how-background-sessions-are-hosted](https://code.claude.com/docs/ko/agent-view#how-background-sessions-are-hosted) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/agent-view#llm-gateway](https://code.claude.com/docs/ko/agent-view#llm-gateway) | llm-gateway-connect | SECTION — 해당 절 전체 |
| [code.claude.com/docs/ko/settings](https://code.claude.com/docs/ko/settings) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/vs-code](https://code.claude.com/docs/ko/vs-code) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [claude.com/docs/third-party/claude-desktop/gateway](https://claude.com/docs/third-party/claude-desktop/gateway) | llm-gateway-connect, llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/llm-gateway-rollout#distribute-through-managed-settings](https://code.claude.com/docs/ko/llm-gateway-rollout#distribute-through-managed-settings) | llm-gateway-connect | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/remote-control](https://code.claude.com/docs/ko/remote-control) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/github-actions](https://code.claude.com/docs/ko/github-actions) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [github.com/anthropics/claude-code-action#readme](https://github.com/anthropics/claude-code-action#readme) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/agent-sdk/overview](https://code.claude.com/docs/ko/agent-sdk/overview) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/slack](https://code.claude.com/docs/ko/slack) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/claude-code-on-the-web](https://code.claude.com/docs/ko/claude-code-on-the-web) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/voice-dictation](https://code.claude.com/docs/ko/voice-dictation) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/env-vars](https://code.claude.com/docs/ko/env-vars) | llm-gateway-connect, llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/server-managed-settings#environment-variables-and-the-approval-dialog](https://code.claude.com/docs/ko/server-managed-settings#environment-variables-and-the-approval-dialog) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/settings-reference#when-claude-code-applies-env-values](https://code.claude.com/docs/ko/settings-reference#when-claude-code-applies-env-values) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/settings-reference#modelpicker](https://code.claude.com/docs/ko/settings-reference#modelpicker) | llm-gateway-connect, llm-gateway-protocol | SECTION — 영어 공식 modelPicker 절로 보완 |
| [code.claude.com/docs/ko/llm-gateway-protocol#model-discovery](https://code.claude.com/docs/ko/llm-gateway-protocol#model-discovery) | llm-gateway-connect, llm-gateway-rollout | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/errors#your-apikeyhelper-script-is-failing](https://code.claude.com/docs/ko/errors#your-apikeyhelper-script-is-failing) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/settings-reference#apikeyhelper](https://code.claude.com/docs/ko/settings-reference#apikeyhelper) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/analytics#access-analytics-for-api-customers](https://code.claude.com/docs/ko/analytics#access-analytics-for-api-customers) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/fast-mode](https://code.claude.com/docs/ko/fast-mode) | llm-gateway-connect, llm-gateway-rollout, llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/data-usage#webfetch-domain-safety-check](https://code.claude.com/docs/ko/data-usage#webfetch-domain-safety-check) | llm-gateway-connect, llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/data-usage#telemetry-services](https://code.claude.com/docs/ko/data-usage#telemetry-services) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/amazon-bedrock#2-configure-aws-credentials](https://code.claude.com/docs/ko/amazon-bedrock#2-configure-aws-credentials) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/google-vertex-ai#4-configure-claude-code](https://code.claude.com/docs/ko/google-vertex-ai#4-configure-claude-code) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/google-vertex-ai#5-pin-model-versions](https://code.claude.com/docs/ko/google-vertex-ai#5-pin-model-versions) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/model-config#customize-pinned-model-display-and-capabilities](https://code.claude.com/docs/ko/model-config#customize-pinned-model-display-and-capabilities) | llm-gateway-connect, llm-gateway-protocol | SECTION — 매핑·custom picker·gateway context·기능 변수 관련 절; 페이지 전체 미독 |
| [code.claude.com/docs/ko/claude-platform-on-aws](https://code.claude.com/docs/ko/claude-platform-on-aws) | llm-gateway-connect, llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/errors#automatic-retries](https://code.claude.com/docs/ko/errors#automatic-retries) | llm-gateway-connect, llm-gateway-rollout, llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/errors#unable-to-connect-to-api](https://code.claude.com/docs/ko/errors#unable-to-connect-to-api) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/errors#api-returned-an-empty-or-malformed-response](https://code.claude.com/docs/ko/errors#api-returned-an-empty-or-malformed-response) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/llm-gateway-protocol#feature-pass-through](https://code.claude.com/docs/ko/llm-gateway-protocol#feature-pass-through) | llm-gateway-connect, llm-gateway-rollout | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/model-config](https://code.claude.com/docs/ko/model-config) | llm-gateway-connect, llm-gateway-rollout, llm-gateway-protocol | SECTION — 매핑·custom picker·gateway context·기능 변수 관련 절; 페이지 전체 미독 |
| [code.claude.com/docs/ko/errors#prompt-is-too-long](https://code.claude.com/docs/ko/errors#prompt-is-too-long) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/fast-mode#use-fast-mode-behind-proxies-and-llm-gateways](https://code.claude.com/docs/ko/fast-mode#use-fast-mode-behind-proxies-and-llm-gateways) | llm-gateway-connect, llm-gateway-rollout, llm-gateway-protocol | SECTION — gateway 종류·구독·별도 설정 |
| [code.claude.com/docs/ko/permissions#what-runs-before-you-trust-a-folder](https://code.claude.com/docs/ko/permissions#what-runs-before-you-trust-a-folder) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/network-config#ca-certificate-store](https://code.claude.com/docs/ko/network-config#ca-certificate-store) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/errors#authentication-errors](https://code.claude.com/docs/ko/errors#authentication-errors) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/authentication](https://code.claude.com/docs/ko/authentication) | llm-gateway-connect | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [platform.claude.com/settings/keys](https://platform.claude.com/settings/keys) | llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/amazon-bedrock#prerequisites](https://code.claude.com/docs/ko/amazon-bedrock#prerequisites) | llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/google-vertex-ai#prerequisites](https://code.claude.com/docs/ko/google-vertex-ai#prerequisites) | llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/microsoft-foundry#prerequisites](https://code.claude.com/docs/ko/microsoft-foundry#prerequisites) | llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/admin-setup#decide-how-settings-reach-devices](https://code.claude.com/docs/ko/admin-setup#decide-how-settings-reach-devices) | llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/llm-gateway-protocol#streaming](https://code.claude.com/docs/ko/llm-gateway-protocol#streaming) | llm-gateway-rollout | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/claude-apps-gateway-config#upstream-error-messages](https://code.claude.com/docs/ko/claude-apps-gateway-config#upstream-error-messages) | llm-gateway-rollout, llm-gateway-protocol | GAP — 대상 원문에 앵커와 token 상세 목록 없음 |
| [code.claude.com/docs/ko/llm-gateway-protocol#disable-pre-release-capabilities](https://code.claude.com/docs/ko/llm-gateway-protocol#disable-pre-release-capabilities) | llm-gateway-rollout | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/costs#background-token-usage](https://code.claude.com/docs/ko/costs#background-token-usage) | llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/llm-gateway-connect#route-to-a-cloud-provider-through-a-gateway](https://code.claude.com/docs/ko/llm-gateway-connect#route-to-a-cloud-provider-through-a-gateway) | llm-gateway-rollout | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/errors#administrator-policy-requires-a-cloud-gateway-sign-in](https://code.claude.com/docs/ko/errors#administrator-policy-requires-a-cloud-gateway-sign-in) | llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/server-managed-settings#platform-availability](https://code.claude.com/docs/ko/server-managed-settings#platform-availability) | llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/llm-gateway-connect#rotate-credentials-with-apikeyhelper](https://code.claude.com/docs/ko/llm-gateway-connect#rotate-credentials-with-apikeyhelper) | llm-gateway-rollout, llm-gateway-protocol | FULL — 지정 문서 전체 |
| [claude.com/docs/third-party/claude-desktop/configuration](https://claude.com/docs/third-party/claude-desktop/configuration) | llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/llm-gateway-connect#configure-each-surface](https://code.claude.com/docs/ko/llm-gateway-connect#configure-each-surface) | llm-gateway-rollout | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/settings-reference#wslinheritswindowssettings](https://code.claude.com/docs/ko/settings-reference#wslinheritswindowssettings) | llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/llm-gateway-connect#configure-claude-code-yourself](https://code.claude.com/docs/ko/llm-gateway-connect#configure-claude-code-yourself) | llm-gateway-rollout | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/llm-gateway-connect#troubleshoot-gateway-errors](https://code.claude.com/docs/ko/llm-gateway-connect#troubleshoot-gateway-errors) | llm-gateway-rollout | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/settings#which-value-claude-code-uses](https://code.claude.com/docs/ko/settings#which-value-claude-code-uses) | llm-gateway-rollout | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/amazon-bedrock#streaming-errors-behind-a-gateway-or-proxy](https://code.claude.com/docs/ko/amazon-bedrock#streaming-errors-behind-a-gateway-or-proxy) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/sub-agents](https://code.claude.com/docs/ko/sub-agents) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/agent-teams](https://code.claude.com/docs/ko/agent-teams) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/permission-modes#eliminate-prompts-with-auto-mode](https://code.claude.com/docs/ko/permission-modes#eliminate-prompts-with-auto-mode) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/authentication#anthropic-profiles-and-federation-credentials](https://code.claude.com/docs/ko/authentication#anthropic-profiles-and-federation-credentials) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/model-config#adjust-effort-level](https://code.claude.com/docs/ko/model-config#adjust-effort-level) | llm-gateway-protocol | SECTION — 매핑·custom picker·gateway context·기능 변수 관련 절; 페이지 전체 미독 |
| [platform.claude.com/docs/en/build-with-claude/context-editing](https://platform.claude.com/docs/en/build-with-claude/context-editing) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [platform.claude.com/docs/en/build-with-claude/context-windows#context-window-sizes-by-model](https://platform.claude.com/docs/en/build-with-claude/context-windows#context-window-sizes-by-model) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [platform.claude.com/docs/en/build-with-claude/extended-thinking#interleaved-thinking](https://platform.claude.com/docs/en/build-with-claude/extended-thinking#interleaved-thinking) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [platform.claude.com/docs/en/agents-and-tools/tool-use/overview](https://platform.claude.com/docs/en/agents-and-tools/tool-use/overview) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [platform.claude.com/docs/en/build-with-claude/effort](https://platform.claude.com/docs/en/build-with-claude/effort) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [platform.claude.com/docs/en/build-with-claude/structured-outputs](https://platform.claude.com/docs/en/build-with-claude/structured-outputs) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/prompt-caching](https://code.claude.com/docs/ko/prompt-caching) | llm-gateway-protocol | SECTION — 이력·도구·모델·compact·resume 관련 절 |
| [platform.claude.com/docs/en/build-with-claude/token-counting](https://platform.claude.com/docs/en/build-with-claude/token-counting) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/amazon-bedrock#use-the-mantle-endpoint](https://code.claude.com/docs/ko/amazon-bedrock#use-the-mantle-endpoint) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [platform.claude.com/docs/en/build-with-claude/extended-thinking](https://platform.claude.com/docs/en/build-with-claude/extended-thinking) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/mcp#scale-with-mcp-tool-search](https://code.claude.com/docs/ko/mcp#scale-with-mcp-tool-search) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [platform.claude.com/docs/en/api/beta-headers](https://platform.claude.com/docs/en/api/beta-headers) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/llm-gateway-connect#turn-off-traffic-outside-the-gateway-path](https://code.claude.com/docs/ko/llm-gateway-connect#turn-off-traffic-outside-the-gateway-path) | llm-gateway-protocol | FULL — 지정 문서 전체 |
| [code.claude.com/docs/ko/settings-reference#availablemodels](https://code.claude.com/docs/ko/settings-reference#availablemodels) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |
| [code.claude.com/docs/ko/model-config#work-with-fable](https://code.claude.com/docs/ko/model-config#work-with-fable) | llm-gateway-protocol | SECTION — 매핑·custom picker·gateway context·기능 변수 관련 절; 페이지 전체 미독 |
| [platform.claude.com/docs/en/api/messages](https://platform.claude.com/docs/en/api/messages) | llm-gateway-protocol | NOT_READ — 직접 교차 참조 목록화; 해당 문서 전체 미독 |

## 추가로 읽은 범위

- Codex SDK: 전체. 기존 SDK 선택을 재평가하되 새 런타임 의존성 추가는 제안하지 않음.
- model-config: gateway context window, custom model, alias env, 표시·기능 변수, modelOverrides, caching 관련 절. 해당 페이지 전체 판독으로 세지 않음.
- agent-view: 백그라운드 전환에서 전달되는 것, 모델 설정, 설정·공급자, LLM gateway 절.
- settings-reference: 한국어 curl 2회는 exit 56, 웹 열기 실패. 영어 공식 .md fetch exit 0 후 modelPicker 절(다음 modelPricing 절 전까지) 전체 판독. modelPicker는 project/local에서 무시된다는 범위를 설계에 반영.
- prompt-caching: 일반 prefix 구성, 모델·effort·tool 변경, compact·이미지 제거·upgrade, file edit·system reminder·resume·child 관련 절. 일부 묶음 출력 잘림 때문에 페이지 전체 FULL로 집계하지 않음.
- claude-apps-gateway-config: 내려받기 성공; upstream-error-messages, capability_rejected, prompt_too_long 검색은 0건. 대상 목차를 조사했으나 페이지 전체를 읽었다고 표시하지 않음.
- OpenAI Terms of Use, Anthropic Commercial Terms: 공식 페이지를 열어 적용 범위·제한 조건을 참조. 개별 계정 계약·조직 정책에 대한 법률 검토는 아님.

## Gaps

직접 교차 링크의 NOT_READ 항목과 그 문서들의 추가 하위 링크는 미독이다. 특히 전체 cloud-provider 배포 문서, Desktop/Slack/웹 사용 설명, 관리·권한·MCP 전체 reference 및 API 전체 사양을 전부 읽었다고 주장하지 않는다. 구현 단계에서 해당 기능을 지원 범위에 넣으면 각 계약을 먼저 읽고 버전별 인수 검사를 추가해야 한다.

## Residual-risk

문서와 설치 바이너리는 다를 수 있다. 문서의 capability 표식만으로 런타임 지원을 판정하지 않는다. 한국어 오류 토큰 상세 링크의 누락을 임의 토큰 생성으로 보완하지 않는다. 이 장부는 원문 판독 범위를 남기는 자료이며 모든 연결 페이지의 감사를 완료했다는 선언이 아니다.
