# MoAI Gateway 통합 설계 — claude-code-proxy 전수 문서 조사

## 1. 결정 요약

권고는 Claude Code를 그대로 실행하면서 MoAI Gateway를 공급자별 프로토콜 어댑터로 확장하는 것이다. moai gpt의 메인 추론 모델은 GPT다. Claude가 GPT에 일을 위임하는 MCP 구조로 요구를 바꾸지 않는다. moai cc·glm·gpt, Factory lead·lane, 에이전트별 모델·effort 설정은 같은 선택 정책을 사용한다.

claude-code-proxy는 이 요구에 맞는 공개 참고 구현이다. Codex App Server를 감싼 프로그램이 아니라, Anthropic Messages 요청을 각 공급자의 서비스 계약에 맞게 변환하는 Rust 프록시다. 다만 소스의 존재와 문서의 성공 사례는 MoAI의 품질·속도·약관 적합성 검증을 대신하지 않는다. [연결 구조](https://claude-code-proxy.raine.dev/how-it-works/)

최선안은 기존 Gateway 기반 + 검증된 프로토콜 변환 + 공급자별 인증·정책 분리다. 모든 공급자를 등록할 수 있게 하되, 미확인 구독 연결까지 모두 기본 활성화하지 않는다. 공개 API·공식 Coding Plan은 우선 통합 후보, Codex 직접 구독은 별도 승인·호환성 검증 후보, Cursor 내부 프로토콜은 보류 후보로 둔다.

이 문서는 설계 제안이다. 제품 코드, 실행 설정, 자격증명, 기존 SPEC은 변경하지 않았다. 계정 연결·실제 추론·이미지 생성·성능 측정은 수행하지 않았다.

## 2. 조사 범위와 기준

조사일은 2026-09-14다. 사이트맵의 문서 21개를 조립 본문으로 읽고, 저장소의 대응 문서 20개와 CHANGELOG.md를 매핑했다. 모델·인증·라우팅·HTTP 입구·일부 변환 소스를 교차 확인했다. 전체 Rust 저장소를 전수 보안 감사한 것은 아니다.

| 기준 | 이번 조사에서 확인한 값 |
|---|---|
| 원본 프로젝트 | raine/claude-code-proxy |
| 고정 소스 | bcfa0a401a6bfe0eea278c65edecfcedaad919c4 |
| 커밋 시각 | 2026-09-10T15:19:54Z |
| 공개 문서 릴리스 | v0.1.39 |
| 사이트맵 / 소스 문서 대응 | 21 / 21 |
| Codex 기본 모델 | 10개, -fast는 별칭 |
| Grok | 3개 |
| Kimi 공개 식별자 | 5개, 실제 upstream은 2종 |
| OpenCode Go | 36개: Chat 23 / Messages 9 / Responses 4 |
| Cursor | 동적 카탈로그 + 3개 공개 접두사, 계정 카탈로그 조회 안 함 |
| MoAI 비교 기준 | develop worktree, HEAD 4056f69e1, 기존 미커밋 변경 존재 |

[문서·소스 해시 목록](corpus-manifest.json), [모델 카탈로그](model-catalog.json), [기계 집계](inventory-output.json)를 함께 제공한다. 모델 등록 수는 사용 가능 계정 수나 실서비스 성공 수가 아니다.

## 3. 원본은 어떻게 연결하는가

Claude Code가 로컬 /v1/messages로 요청한다. registry가 모델 접미사·별칭을 정리하고 공급자를 선택한다. 공급자별 인증 저장소가 토큰을 공급하고, 어댑터가 system·messages·tools·tool results·thinking·images를 변환한다. 상류 스트림을 Anthropic SSE로 돌려주며 도구 실행은 Claude Code가 담당한다. [연결 구조](https://claude-code-proxy.raine.dev/how-it-works/)

| 공급자 | 실제 연결 | 인증 | 핵심 차이 |
|---|---|---|---|
| Codex | chatgpt.com/backend-api/codex/responses | 프록시 소유 OAuth, PKCE·device | 구독용 내부 Responses, WS 기본·HTTP SSE 선택 |
| Kimi | api.kimi.com/coding/v1/chat/completions | device OAuth·기기 ID | Coding 서비스의 Chat 변환 |
| Grok | cli-chat-proxy.grok.com/v1/responses | 브라우저 PKCE·device | CLI용 Responses 서비스 |
| OpenCode Go | opencode.ai/zen/go/v1 | 구독 API 키 | 모델마다 Chat·Messages·Responses 구분 |
| Cursor | api2.cursor.sh/agent.v1.AgentService/Run | 프록시 로그인·토큰 | HTTP/2 Connect와 로컬 Agent 번들의 protobuf |

원본은 native Codex·Grok·Cursor의 기존 로그인 파일을 읽는 대신 자기 저장소를 사용한다. 하지만 자기 저장소를 쓴다는 사실만으로 서비스 제공자의 제3자 OAuth 허용까지 입증되지는 않는다. 클라이언트의 ANTHROPIC_AUTH_TOKEN은 상류 인증에 쓰지 않으며, 로컬 요청 인증도 하지 않는다. [인증·HTTP 계약](https://claude-code-proxy.raine.dev/reference/http-api/)

주목할 장점은 공급자별 책임 분리와 동일 스트림 계약이다. MoAI가 가져올 것은 이 동작 계약과 회귀 사례다. 로그인·모델·세션·보안 처리를 또 하나의 독립 제품으로 중복 구현할 필요는 없다.

## 4. 모든 모델을 흡수하는 범위

흡수는 모델 가중치를 가져오는 것이 아니라, 아래 모든 라우팅 계열을 MoAI 카탈로그와 어댑터로 등록한다는 뜻이다. 공급자가 목록을 바꾸면 새 catalog revision으로 반영하고, 활성 세션의 선택은 고정한다.

| 계열 | 고정 소스의 전체 기본 모델 또는 식별자 |
|---|---|
| Codex | gpt-5.2, gpt-5.3-codex, gpt-5.3-codex-spark, gpt-5.4, gpt-5.4-mini, gpt-5.5, gpt-5.6-luna, gpt-5.6-sol, gpt-5.6-terra, gpt-6-astra |
| Codex 옵션 | 각 기본 모델의 -fast → priority 요청; 별칭은 독립 모델로 세지 않음 |
| Kimi 공개 ID | kimi-for-coding, kimi-k2.6, k2.6, kimi-k3, k3 |
| Kimi upstream | kimi-for-coding, k3; K2.6 별칭은 기본 coding 모델로 해석 |
| Grok | grok-composer-2.5-fast, grok-4.5, grok-4.6 |
| Cursor 공개 접두사 | cursor:, cursor-plan:, cursor-ask: 뒤에 동적 모델 ID |
| Cursor legacy | cursor, cursor-agent, cursor-composer, cursor-composer-fast, cursor-plan, cursor-ask, composer-2.5, composer-2.5-fast |
| 이미지 | Codex 내부 이미지 어댑터의 gpt-image-2; 채팅 모델 목록과 분리 |

OpenCode Go 전체 36개는 다음과 같다. 모든 이름은 opencode-go/ 접두사를 붙여 저장한다. 동일 GPT·Grok·Kimi 이름이라도 직접 구독과 OpenCode Go는 별도 인증·요금·데이터 처리 경로다.

| 프로토콜 | 모델 전체 목록 |
|---|---|
| Responses · 4 | grok-4.6, gpt-5.6-luna, muse-spark-1.3-contributor, muse-spark-1.2-contributor |
| Messages · 9 | minimax-m3, minimax-m2.7, minimax-m2.5, qwen3.8-max, qwen3.8-flash, qwen3.7-max, qwen3.7-plus, qwen3.6-plus, qwen3.5-plus |
| Chat · 23 | grok-4.5, glm-5.2, glm-5.3-flash, glm-5.3, glm-5.1, glm-5, kimi-k3, kimi-k2.7-code, kimi-k2.6, kimi-k2.5, longcat-2.0, deepseek-v4-pro, deepseek-v4-flash, deepseek-flash, deepseek-v4-flash-vision-exp, mimo-v2-pro, mimo-v2-omni, mimo-v2.5, mimo-v2.5-pro, hy3, hy4-preview, hy3-preview, omen-alpha |

[고정 모델 소스](https://github.com/raine/claude-code-proxy/blob/bcfa0a401a6bfe0eea278c65edecfcedaad919c4/src/providers/opencode/model.rs), [상위 registry](https://github.com/raine/claude-code-proxy/blob/bcfa0a401a6bfe0eea278c65edecfcedaad919c4/src/registry.rs).

기존 Z.AI 직접 GLM과 Claude native도 유지한다. 범용 확장은 OpenAI API, Anthropic API, 다른 공식 Coding Plan의 커스텀 endpoint 등록으로 처리한다. 커스텀 endpoint에는 HTTPS·자격증명 host 고정·리다이렉트 차단·명시적 사설망 허용을 적용한다. 이름만 등록된 모델은 아직 실행 후보가 아니다.

## 5. 문서와 구현의 차이

| 관찰 | 근거 | 설계 반영 |
|---|---|---|
| Kimi 설명 페이지에는 K2.6 중심의 목록, 소스에는 K3도 존재 | providers/kimi 문서와 registry·model_allowlist 대조 | 문서 복사 대신 revision별 카탈로그 생성 |
| HTTP API 문서에는 음성 경로가 없지만 소스는 선택적으로 등록 | server.rs:282의 /v1/audio/transcriptions | 대화 오디오와 별도 전사 API를 혼동하지 않음; 추가 옵션으로만 설계 |
| 원본 소스의 OpenCode 목록과 공급자 최신 목록이 다름 | 공급자 공식 목록에 deepseek-v4.1-flash, 고정 소스 목록에는 없음 | 소스 목록과 upstream 목록의 차이를 표시; 자동 활성화 금지 |
| GPT 최신 계열은 Responses Lite 분기를 사용 | codex/mod.rs:195, :606 및 model_allowlist.rs | 공개 API와 동일한 인증·헤더 계약으로 취급하지 않음 |
| 자동 hosted search에서는 Luna가 Sol로 바뀔 수 있음 | full_lane_web_search_model, apply_model_lane_for_request | 요청 모델·실제 모델·변경 이유 노출, 변경 금지 모드 지원 |

이 표는 정적 소스·문서의 관찰이다. 해당 경로를 실계정에서 재현한 결함 보고가 아니다. 문서상의 전체 모델 지원을 실계정 성공으로 확대 해석하지 않는다. [공급자 최신 목록](https://opencode.ai/docs/go/), [분기 소스](https://github.com/raine/claude-code-proxy/blob/bcfa0a401a6bfe0eea278c65edecfcedaad919c4/src/providers/codex/translate/model_allowlist.rs).

| 호환성 경계 | 원본의 동작 | MoAI 기본 정책 제안 |
|---|---|---|
| Cursor 도구 | 인식하는 Read·Write·Bash 중심, streaming과 안정적 session 필요 | Agent·MCP·Artifact까지 동일 지원한다고 표시하지 않음 |
| Grok 이미지 | 기본 omit, 선택적으로 vision 전달 | 이미지가 필요한 작업에서 pixels를 생략한 성공 금지 |
| Codex Chat Completions 입구 | 함수 도구·이미지 등 미지원; Responses 입구와 다름 | 기능은 공급자뿐 아니라 ingress·protocol 조합으로 검사 |
| search | 공급자마다 필터·max_uses·hosted tool 의미가 다름 | 보안·도메인 제한을 손실시키는 매핑 거절 |
| token count | 로컬 추정, 실제 과금 token과 다름 | estimate·reported·unknown을 분리 |
| native compaction | 추가 요청과 메모리 상태, 재시작 시 portable 요약 fallback | 지연·비용·회상 손실까지 별도 표시 |

[공급자별 호환성](https://claude-code-proxy.raine.dev/reference/compatibility-and-limitations/), [Cursor 제한](https://claude-code-proxy.raine.dev/providers/cursor-agent/), [Grok 이미지 정책](https://claude-code-proxy.raine.dev/providers/grok/).

## 6. ToS 판정 — 소프트웨어와 서비스 권리는 별개

법률 자문이나 무위험 보증이 아니라 공개 약관·문서에 근거한 통합 위험 평가다. 적용 국가, 개인·조직 계약, 이용 목적, 공급자 승인 범위에 따라 결론이 달라진다. OAuth 성공, 유료 구독, MIT 라이선스 중 어느 하나도 나머지 권한을 자동으로 부여하지 않는다.

| 연결 | 확인한 근거 | 배포 정책 제안 |
|---|---|---|
| 원본 코드 이용·수정 | MIT, Raine Virta 저작권 표시와 허가문 보존 조건 | 코드·테스트 이식 가능. THIRD_PARTY_NOTICES와 출처 추적 필요 |
| Claude 구독 | 수정하지 않은 Claude Code의 사용자 직접 로그인은 공식 문서에서 허용 범위를 설명 | moai cc는 native 로그인 유지. Gateway가 Claude 구독 토큰을 수집·중계하는 기능은 기본 경로에서 제외 |
| ChatGPT Codex 직접 구독 | 공식 Codex 구독 제공은 확인. 이 프로젝트는 타 하네스를 환영한다는 OpenAI 관계자 게시물을 인용 | 직접 연결을 유력 후보로 유지하되, 정확한 OAuth client·Lite·이미지·Factory 범위 승인 확인 전 정식 지원 판정 보류 |
| OpenAI 공개 API | 공식 API 인증·문서에 따른 통합 | 별도 API 과금임을 표시하고 우선 구현 후보 |
| OpenCode Go | 공식 문서가 다른 코딩 에이전트 사용과 Claude Code 검증을 명시 | 공식 key 방식 우선. 자기 User-Agent와 안정적 세션 ID, 요금·지역 제한 준수 |
| Kimi Code | 최신 공식 문서가 CC·Roo·OpenCode·OpenClaw·Hermes 사용을 안내 | 공식 Coding API key 우선. 원본의 별도 device OAuth는 별도 확인 |
| GLM Coding Plan | Z.AI가 Claude Code 설정을 직접 문서화 | 기존 공식 key·endpoint 경로 유지; 플랜 범위·한도 준수 |
| Grok CLI OAuth | 기술 구현 존재. AUP는 비인가 자동 접근·보호조치 우회 등을 제한 | 정확한 CLI 구독의 제3자 사용 승인 확인 전 제한. 공식 API·OpenCode 경로와 분리 |
| Cursor Agent 내부 Connect | 설치된 JS 번들의 schema 이용. 약관은 역공학·내부 구조 접근 등에 제한 | 기본 비활성화. 공급자 승인 또는 문서화된 대체 인터페이스 전 운영 핵심에서 제외 |

근거: [MIT 원문](https://github.com/raine/claude-code-proxy/blob/bcfa0a401a6bfe0eea278c65edecfcedaad919c4/LICENSE), [Anthropic 인증·배포 조건](https://code.claude.com/docs/en/legal-and-compliance), [OpenAI 약관](https://openai.com/policies/row-terms-of-use/), [Codex 구독](https://help.openai.com/en/articles/11369540-using-codex-with-your-chatgpt-plan), [ChatGPT Pro 이용 제한](https://help.openai.com/en/articles/9793128-what-is-chatgpt-pro), [OpenCode Go](https://opencode.ai/docs/go/), [Kimi 공식 안내](https://www.kimi.com/en/help/kimi-code/third-party-agents), [GLM 공식 안내](https://docs.z.ai/devpack/tool/claude), [Grok AUP](https://x.ai/legal/acceptable-use-policy), [Cursor 약관](https://cursor.com/terms-of-service).

## 7. GPT 구독은 무엇을 더 확인해야 하는가

직접 구독 프록시를 일괄적으로 불법이라고 단정하지 않는다. 프로젝트 문서는 OpenAI 관계자의 타 하네스 이용 환영 게시물을 연결한다. 다만 이번에는 X 원문을 독립적으로 읽지 못했다. 웹 열기 오류, oEmbed 빈 응답, syndication의 빈 객체를 관찰했으므로 이 인용을 독립 검증된 승인으로 격상하지 않는다. [프로젝트의 인용](https://claude-code-proxy.raine.dev/providers/codex/), [연결된 X 원문](https://x.com/thsottiaux/status/2075830097488249060).

고정 소스에서 gpt-5.6-luna·sol·terra와 gpt-6-astra는 Lite 분기로 들어간다. 해당 헤더 빌더는 originator를 codex_cli_rs로 두고 x-openai-internal-codex-responses-lite를 보낸다. 전사 헤더에는 Codex Desktop 문자열도 있다. 이것은 관찰한 구현이지 법률상 위반 확정은 아니다. 그러나 단순한 자체 이름의 공개 API 호출과 동일하다고 설명하면 안 된다. [헤더 소스](https://github.com/raine/claude-code-proxy/blob/bcfa0a401a6bfe0eea278c65edecfcedaad919c4/src/providers/codex/client.rs#L117)

정식 배포 전에 확인할 질문은 구체적이어야 한다.

- MoAI 자체 OAuth client와 개인 구독 사용을 허용하는가? 기존 공개 client ID 재사용 조건은 무엇인가?
- Claude Code를 프런트로 둔 로컬 개인 개발과 Factory 병렬 세션이 허용 범위에 들어가는가?
- Responses Lite·native compaction·내부 이미지 서비스·전사 endpoint 사용 권한은 각각 무엇인가?
- 허용되는 client 식별값·User-Agent·서비스 한도·팀 좌석 경계는 무엇인가?

개인 로컬 개발을 공유 계정 API 서버·구독 재판매·다른 사용자를 위한 SaaS로 확대하지 않는다. 체크박스에 동의했다고 공급자 권한이 생기는 것도 아니다. 정책 확인은 계약 또는 공식 문서 근거로 기록한다.

## 8. 이전 App Server 중심안에서 바꿀 점

기존 최종 검토안은 MoAI Messages와 Codex App Server 연결을 중심에 둔다. 이번 조사로 직접 프로토콜 프록시도 중요한 비교 후보임이 확인됐다. App Server만이 GPT 메인 세션을 만들 수 있는 유일한 기술 경로라는 전제는 채택하지 않는다. [이전 문서](../gpt-final-design-20260914-b40ff9f4/report.html)

| 경로 | 장점 | 부담 | 권고 |
|---|---|---|---|
| Messages → 직접 Codex 구독 프로토콜 | CC 도구 실행을 유지; 원본의 실시간 변환·세션 설계 참고 가능 | 내부 계약·OAuth 허용·모델별 Lite 차이 | 정책·인수 충족 시 moai gpt 기본 후보 |
| Messages → Codex App Server | 공식 실행 서비스의 인증·thread·turn 계약 활용 | CC와 Codex의 도구·이력 소유권 조정 필요 | 기존 bridge 유지, 비교 시험과 대체 adapter 후보 |
| Messages → 공개 OpenAI Responses | 공식 API 계약, 독립 인증 | 별도 과금·구독과 사용 가능 모델 차이 | 명시적으로 선택하는 공식 API 후보 |
| Messages → OpenCode Go GPT | 공식 코딩 플랜 API, 여러 공급자 통합 | ChatGPT 구독 아님; 제공 GPT 목록 제한 | 경제형 GPT 메인 선택지 |
| Claude main → GPT MCP | 분석·감사·이미지의 선택적 위임에 유리 | 메인 모델이 GPT가 아님 | 보조 기능으로만 제공 |

MoAI develop에서 gateway_product_binding.go는 app-server 표시 계약을 가지며, openai.go에는 직접 API·구독 변환 경로가 존재한다. 어느 경로가 실제 실행되는지는 표시 문자열만으로 판정할 수 없다. 새 설계는 배포 binding·receipt의 실제 adapter ID를 확인하고, 기존 구현을 버리기 전에 같은 작업으로 비교한다.

## 9. 목표 아키텍처

MoAI Launcher는 사용자 선택과 프로세스 환경만 조립한다. Factory는 작업·lane을 배치한다. Gateway는 인증된 요청의 모델을 해석하고 프로토콜을 변환한다. Claude Code가 도구 권한·승인·실행을 맡는다. MCP는 같은 Gateway 정책을 호출하는 보조 입구다. 토큰을 각 계층이 중복 소유하지 않는다.

| 계층 | 책임 | 두지 않을 책임 |
|---|---|---|
| moai cc / glm / gpt | 기존 UX, 세션 프로필, 환경 격리 | 사용자 전역 설정 덮어쓰기 |
| Factory lead·lane | 카드·worktree·역할·작업 소유권 | 다른 계정 한도로 우회 |
| Route Resolver | 역할·모델·effort 결정, revision 고정 | 스트림 도중 몰래 모델 교체 |
| Gateway core | 입구 인증, 모델·기능 검사, 취소·관측 | shell·Edit 재실행 |
| Provider adapter | wire 변환·오류·상류 인증 주입 | 임의 URL로 credential 전달 |
| Credential broker | 계정별 저장·refresh·권한 범위 | Claude native 구독 토큰 수집 |
| moai-mcp | llm_task·review·image 도구 입구 | 별도 무제한 실행 엔진 |

서버는 사용자별 로컬 daemon 하나를 기본으로 하되, 각 CC 프로세스에는 범위가 좁은 세션 capability token을 발급한다. 별도 계정·조직·정책은 별도 credential namespace로 격리한다. 장애 격리가 필요한 공급자는 프로세스를 분리하되 무조건 공급자 수만큼 daemon을 늘리지 않는다.

## 10. 공통 데이터 계약

아래 필드는 새 설계다. 이미 존재하는 CLI·구조체로 오해하지 않는다. 모델명·공급자·인증·프로토콜을 하나의 문자열에 섞지 않고 각각 기록한다.

| 객체 | 필수 항목 |
|---|---|
| Provider | provider_id, endpoint_allowlist, auth_modes, policy_status, source_url, reviewed_at |
| ModelEntry | qualified_id, upstream_id, protocol, catalog_revision, supported_efforts, context_limit, tool·image·schema capability |
| RouteBinding | project, session, lane, agent, role, account_ref, model_entry, requested_effort, effective_effort, fallback_policy |
| RequestEnvelope | request_id, turn_id, tool_call_ids, normalized_messages, tool_schema_hash, data_policy, deadline |
| Continuation | owner_key, model·account scope, socket_generation, opaque_payload, portable_history_hash, compaction_epoch |
| Receipt | requested·effective route, actual model if known, degradation_reason, usage_kind, finish_reason, output_artifacts |

ModelEntry의 상태는 discovered → policy-eligible → contract-tested → live-verified → enabled로 분리한다. 정책·기능·실서비스 검증을 하나의 supported 불리언으로 뭉치지 않는다. 공급자가 실제 모델명을 반환하지 않으면 actual_model=unknown으로 기록하고 확인됐다고 표시하지 않는다.

기존 MoAI의 ModelEntry·auth·translate·receipt·conversation 기능을 먼저 재사용한다. 새 추상화는 기존 타입의 확장으로 시작하고, 서로 다른 wire 계약을 억지로 OpenAI 하나에 맞추지 않는다.

## 11. 일반 모드·Factory·에이전트 모델 선택

기존 명령은 유지한다. moai gpt는 GPT를 메인으로, moai glm은 GLM을 메인으로, moai cc는 기본 Claude native로 시작한다. Factory에서도 각 lane은 독립 Claude Code 세션이며 그 세션의 실제 주 모델을 선택한다. -f와 --factory의 정확한 기존 명령 계약은 구현 시 현행 CLI에 맞춰 유지한다.

선택 순서는 명시적 작업 pin → 에이전트 역할 pin → lane 프로필 → 세션 프로필 → 프로젝트 기본값이다. 조직·데이터·권한 제한은 항상 상위 제약으로 적용한다. 전역 모델 override로 자식의 명시적 선택을 덮어쓰지 않는다.

사용자 설정 흐름은 공급자 등록 → 공식 인증 방식 선택 → 사용 가능 모델 확인 → 경제형·균형형·품질형 배치 → 적용 범위 미리보기 → 새 세션 실행 순서다. 시작 화면에서 실제 main·small·review·image 모델과 과금 출처를 먼저 보여준다. 실행 중 base URL을 바꾸려면 새 프로세스가 필요함을 안내하고, 기존 세션을 강제로 종료하지 않는다.

추가 CLI의 제안 이름은 moai gateway providers, moai gateway login, moai gateway models, moai gateway doctor, moai gateway profile이다. 아직 구현된 명령으로 제시하는 것이 아니다. 새 공급자를 추가할 때마다 새로운 최상위 명령을 늘리지 않고 기존 cc·glm·gpt는 익숙한 프로필 진입점으로 유지한다.

| 사용 상황 | 제안 동작 |
|---|---|
| GPT 메인 + GLM 구현 에이전트 | GPT 메인은 유지; 해당 Agent 요청의 model ID만 GLM adapter로 라우팅 |
| Claude 구독 메인 + GPT 분석 | native Claude 로그인 유지; GPT는 MCP 도구 또는 별도 GPT lane에 위임 |
| Claude API 메인 + GPT subagent | 같은 Gateway에서 API 모델별 라우팅; 각 공급자 자격증명 분리 |
| Factory lead GPT / lane GLM·GPT·OpenCode | lane 시작 전에 모델·effort·예산·프로필 고정; 결과만 lead에 반환 |
| 공급자 변경 | 이전 opaque state를 전달하지 않음; portable context로 새 binding 구성·전환 표시 |

Claude native 구독 프로세스의 subagent만 다른 endpoint로 보내는 것을 기본 환경변수 하나로 된다고 주장하지 않는다. 같은 프로세스의 base URL은 공통이다. Claude 구독 토큰 중계를 피하면서 혼합하려면 MCP 또는 별도 lane을 써야 한다. 모든 모델을 같은 Gateway로 보내려면 Claude 쪽도 공식 API 인증으로 통일하는 선택이 더 명확하다.

최신 Claude Code 공식 문서는 기본적으로 메인 아래 3단계 중첩을 설명한다. 사용자 아이디어는 현재 문서와 맞는다. 다만 MoAI 자체의 역할·깊이 제한은 별도이며, 3단계를 기본 강제하지 않는다. 관련 프로젝트 규칙 변경은 별도 승인 후 다룬다. [중첩·모델·MCP 공식 문서](https://code.claude.com/docs/en/subagents#let-subagents-spawn-their-own-subagents)

## 12. effort와 경제성 정책

모델의 잘하는 영역은 초기 가설이지 이 조사에서 측정한 순위가 아니다. 설계·코드·감사·이미지별 후보를 두고 MoAI 작업셋으로 검증한다. 모든 단계에 고가 모델과 max effort를 쓰거나, 모든 작업을 저가 모델로 처리하는 단일 전략은 채택하지 않는다.

| 작업 | 초기 후보 정책 | 승격 조건 |
|---|---|---|
| 요구 분석·복잡한 설계 | 검증된 GPT 또는 Claude 상위 모델, high | 불확실성·교차 모듈 영향·상충 조건 |
| 명확한 구현·테스트 작성 | GLM 또는 검증된 OpenCode 코딩 모델, medium | 동일 실패 반복 또는 요구 누락 |
| 짧은 탐색·문서 정리 | 검증된 저비용 모델, low | 맥락 회상 실패·중요 사실 누락 |
| 독립 리뷰 | 작성자와 다른 모델 계열, 필요할 때만 | 보안·권한·데이터 변경에는 상위 검토 |
| 이미지 | 이미지 전용 모델·예산 | 사용자가 추가 생성·수정 선택 |

effort의 같은 이름이 같은 추론량이라는 전제를 두지 않는다. 지원되지 않는 값은 거절하거나 사용자가 승인한 매핑을 표시한다. priority 서비스와 reasoning effort는 서로 다른 필드다. 원본 Grok의 xhigh/max 매핑처럼 다운그레이드되는 경우 effective_effort를 반드시 보여준다.

비용은 월 구독액뿐 아니라 완료 작업당 사용량·실패 재시도·캐시 손실·검토 비용·이미지 비용으로 본다. OpenCode Go는 조회 시 월 $10이고 모델별 한도가 있으며, Zen 잔액 사용을 켜면 한도 이후 추가 사용이 가능하다. 따라서 무료·무제한으로 표시하지 않는다. 기본 초과 과금은 OFF로 제안한다. [현재 요금·한도](https://opencode.ai/docs/go/)

선택 UI는 경제형·균형형·품질형 세 프로필을 제공하되 실제 배치는 측정 결과로 갱신한다. 비공개 코드의 외부 공급자 전송 허용은 비용 최적화보다 우선한다. 캐시를 보존하기 위해 역할을 불필요하게 쪼개거나 turn마다 공급자를 바꾸지 않는다.

## 13. 400·이력 손실·중복 도구 방지

원본은 continuation을 선택적으로 켜며, 같은 살아 있는 WS와 append-only 이력일 때만 previous_response_id를 붙인다. Main과 direct Agent별 상태를 분리한다. 식별자가 불명확하면 요청 전체를 거절하지 않고 continuation 없이 진행한다. 이 원칙은 참고할 가치가 있다. [원본 Codex 세션 정책](https://claude-code-proxy.raine.dev/providers/codex/)

| 사건 | MoAI 정책 제안 |
|---|---|
| 같은 세션·같은 계정·같은 모델의 후속 turn | 검증된 continuation 사용 |
| agent fork·역할 변경·대화 분기 | 새 owner key; 부모 opaque 이력 재사용 금지 |
| 연결 재생성·daemon 재시작 | 이전 WS response ID 폐기; 검증된 portable history로 복구 |
| compact | 원본 요약과 native artifact의 경계·해시·epoch를 함께 기록 |
| 이력 바뀜 | 무조건 400 또는 무조건 재생하지 않음. 안전한 새 context가 있으면 명시적 전환, 없으면 복구 요청 |
| tool-call 후 전송 오류 | 도구 실행 여부 확인 전 재전송 금지. ledger로 중복 적용 방지 |
| 이미 출력된 stream 오류 | 성공처럼 끝내지 않음. partial 상태·재개 지점 반환 |
| 이미지·필드 변환 불가 | 조용히 텍스트 placeholder로 바꾸지 않고 오류 또는 사용자 승인된 degradation |

opaque reasoning·compaction은 해독하거나 타 모델에 텍스트로 주입하지 않는다. owner key는 사용자·프로젝트·세션·lane·direct agent·account scope·model·epoch를 포함한다. parent agent ID는 lineage 검증용이며 현재 자식의 소유권을 대체하지 않는다.

도구 ledger의 식별자는 owner + turn + tool_call_id + payload hash다. pending/running/completed/unknown 상태와 결과 hash를 기록한다. HTTP 재시도와 shell·파일 변경 재실행을 분리한다. 분산 환경에서 exactly-once를 보증한다고 표현하지 않는다.

전체 이력 재전송으로 400이 사라져도 회상 품질이 보존됐다는 뜻은 아니다. 요약 전환은 quality_degraded 여부를 기록하고 오래된 요구·파일명·결정 회상 평가로 검증한다.

## 14. 스트리밍·속도·한도

Gateway가 provider 응답을 끝까지 모은 다음 SSE로 포장하는 것을 정상 실시간 스트리밍으로 부르지 않는다. 텍스트·reasoning summary·tool argument delta를 검증 가능한 순서대로 전달한다. 도구 실행은 유효한 완성 JSON과 권한 확인 뒤에만 허용한다.

원본에서 배울 항목은 bounded event size, keepalive, cancellation, 오류 상태 보존, Retry-After, WS와 HTTP의 제한적 전환이다. HTTP fallback은 요청이 아직 전달되지 않은 경우와 결과가 불확실한 경우를 구분해야 한다. [문제 해결 계약](https://claude-code-proxy.raine.dev/using/troubleshooting/)

| 지표 | 실제로 측정할 것 | 인수 기준 제안 |
|---|---|---|
| TTFT | 직접 연결 대비 Gateway 첫 유효 delta 시간 | 같은 조건의 분포 비교; 허용 증가폭을 사용자와 확정 |
| 최종 완료 시간 | 도구 실행·재시도·compact 포함 | TTFT만 줄이고 완료 시간이 늘지 않는지 평가 |
| 도구 정확성 | Read/Edit/Bash/Agent/MCP/Artifact 인수·결과 | 필수 fixture 100% 일치; 중복 부작용 0 |
| 이력 | fork·resume·compact 뒤 요구 회상 | 중요 제약 누락 0; 품질 대조 평가 |
| 비용 | 실제 usage·추정치·구독 차감·재시도 | 완료 작업 단위 비교, 0을 unknown 대용으로 쓰지 않음 |
| 병렬성 | 1·2·4 lane 및 허용 범위 내 확대 | 429·403·교차 세션 혼입·메모리 추적 |

이 문서는 속도 향상 수치나 장애 제로를 보장하지 않는다. 현재 baseline의 지연·품질 수치는 NOT_RUN이다. [1m]은 클라이언트 힌트일 뿐이므로 공급자·플랜·모델별 실제 한도와 안전 여유로 compact 시점을 계산한다. 원본의 272K와 기존 MoAI 상수 872K를 모든 GPT에 공통 적용하지 않는다.

## 15. moai-mcp와 이미지 생성

MCP는 보조 입구로 통합한다. 기존 mcp_glm.go에는 glm_audit가 Z.AI Messages를 호출하는 코드가 있다. 이를 참고해 공급자별 별도 인증·정책 구현을 늘리기보다 공통 Gateway service를 호출하도록 설계한다. 기존 함수가 이미 GPT를 지원한다는 뜻은 아니다.

| 제안 도구 | 역할 | 제한 |
|---|---|---|
| llm_task | 정해진 입력의 분석·생성 | 기본 read-only, 예산·모델 pin |
| llm_review | 독립 검토·근거 반환 | 검토 결과와 PASS 판정 근거 분리 |
| image_generate | 이미지 생성 artifact | 명시적 이미지 예산·출력 경로 |
| image_edit | 선택된 파일 수정 | 파일 범위·MIME·크기·개수 검증 |
| job_status / cancel | 진행·취소 | 소유 세션만 조회·취소 |

MCP 함수 호출만으로 외부 모델이 Claude Code의 모든 도구를 자동 획득하지 않는다. 코드 실행이 필요하면 제한된 worker 세션 또는 명시적 도구 루프를 써야 하며, 권한을 이중으로 관리하지 않는다. gpt-agent-plan/run 이름은 역할 프로필로 만들 수 있지만 이름이 실제 모델을 보증하지 않으므로 receipt의 effective route로 확인한다.

이미지는 두 adapter로 나눈다. openai-images-api는 공식 공개 API·별도 과금의 우선 후보다. codex-images-subscription은 원본의 gpt-image-2 내부 구독 endpoint를 참고한 선택 기능이며 해당 서비스 권한·계약 확인 전 기본 OFF다. 채팅 GPT 구독의 성공을 이미지 권한으로 확대하지 않는다. [공개 이미지 API](https://developers.openai.com/api/docs/guides/image-generation), [원본 이미지 계약](https://claude-code-proxy.raine.dev/reference/http-api/)

원본 이미지 기능은 생성·편집을 지원하고, 편집은 JSON data URL 또는 multipart 입력을 받는다. mask·remote URL·variation 등의 제한이 있다. MoAI는 capability별로 UI를 숨기거나 명확히 거절한다. 출력은 artifact ID·파일 hash·MIME·크기·생성 route·usage로 반환하고 base64 대용량 본문을 채팅 이력에 반복 삽입하지 않는다. 파일은 새 경로에 저장하며 기존 자산 덮어쓰기는 별도 승인한다.

## 16. 보안·운영 계약

원본의 loopback 기본값은 유익하지만 무인증 리스너는 그대로 가져오지 않는다. 로컬의 다른 프로세스도 접근할 수 있기 때문이다. 이 조사에서 악용을 재현한 것은 아니며 원본 문서가 인증 부재를 명시한다. [HTTP 보안 경계](https://claude-code-proxy.raine.dev/reference/http-api/)

- 세션별 토큰을 실제 검증하고 프로젝트·공급자·모델·예산·만료를 제한한다. 웹 Origin 검사는 방어 수단의 일부이지 인증 대체가 아니다.
- 자격증명은 OS 보안 저장소 우선, fallback 파일은 소유권·0600·원자적 갱신·로그 마스킹을 적용한다. refresh는 계정별 single-flight로 묶는다.
- 계정별 동시성·요청 대기열·429 cooldown을 공유한다. 여러 lane을 독립 무제한 한도로 취급하지 않는다.
- 요청 중 공급자·계정 자동 전환 금지. 이미 송신한 작업의 결과가 불확실하면 unknown 상태로 남긴다.
- 원문 트래픽 로그는 기본 OFF. 프롬프트·코드·도구 결과·이미지까지 민감정보로 다룬다. 헤더 마스킹만으로 안전한 로그가 되지 않는다.
- 상태 화면은 requested/effective model, effort, account 별칭, 추정·실사용량, 정책 상태, context mode를 보여주되 토큰은 표시하지 않는다.
- 자동 보안 분류기는 가격 때문에 임의 교체하지 않는다. 권한 판단의 기존 의미와 평가를 유지한다.
- /healthz는 생존 검사, readiness는 adapter 구성 검사, 실제 계정 접근 probe는 사용자 요청에 따른 별도 유료 가능 작업으로 구분한다.

## 17. 통합 방식과 구현 순서

| 대안 | 판단 |
|---|---|
| Rust 프록시를 이름만 바꿔 배포 | 기존 MoAI auth·receipt·프로필과 중복. 원본 한계도 함께 들어오므로 비권고 |
| 선택적 sidecar로 활용 | 비교 실험·빠른 호환성 확인용. 공급자 정책 확인 후 제한적으로 사용 |
| Go Gateway에 계약·테스트 이식 | 최종 권고. 기존 구조를 재사용하고 원본의 기능을 adapter 단위로 이식 |
| 모든 기능 일괄 재작성 | 인증·세션·이미지·도구 변경이 한 번에 겹치므로 비권고 |

| 순서·우선순위 | 산출물 | 통과 조건 |
|---|---|---|
| A · High | 정책·모델 카탈로그, 기존 CC/GLM/GPT binding 명세 | 인증 방식·endpoint·모델 충돌·적용 범위 검토 |
| B · High | 공통 라우터·프로토콜 적합성 시험 | 본문·도구·오류·취소 fixture, 인증 격리 |
| C · High | OpenCode Go·Kimi 공식 key adapter, 기존 GLM 유지 | 공식 경로의 단일 CC 세션 인수 |
| D · High | GPT 직접 구독 vs App Server vs API 비교 | 정확한 정책 근거, 실제 모델·tool loop·장기 이력 비교 |
| E · High | Factory lane·Agent pin·예산·ledger | 1 lane → 병렬 lane → fork/resume 순 인수 |
| F · Medium | MCP 공통 입구·공개 이미지 API | 권한·artifact·비용·편집 검증 |
| G · Conditional | Grok OAuth·Cursor·내부 이미지·음성 | 공급자 허용·계약·기능 충족 시에만 승격 |

원본 코드의 실질적 부분·테스트를 가져오면 MIT 고지와 upstream SHA를 남긴다. 모든 Rust 의존성을 Go로 옮기지 않고, 테스트 입력·기대 이벤트·실패 경계를 우선 이식한다. 카탈로그·호환성 변경은 버전 고정과 명시적 갱신으로 운영한다. 기존 SPEC의 App Server 선택을 직접 바꾸는 작업은 이번 범위 밖이다.

## 18. 성공 사례에서 가져올 실무 원칙

성공 사례는 공급자의 공식 검증과 프로젝트의 수정 이력을 구분한다. 벤치마크 없는 커뮤니티 사용담을 성능 보증으로 사용하지 않았다.

| 근거 | 확인된 내용 | MoAI에 반영할 원칙 |
|---|---|---|
| OpenCode Go 공식 문서 | Claude Code 검증, native session header 인식 | 세션 ID를 모든 주·보조 요청에서 보존 |
| Kimi·Z.AI 공식 안내 | Claude Code Coding Plan 연결 방법 제공 | 불필요한 OAuth 역공학보다 공식 key·Messages 경로 우선 |
| 원본 v0.1.31 및 v0.1.36 | Agent 상태 분리·compaction 경합·후속 tool continuation 수정 공지 | 각 회귀 사례를 독립 fixture로 확보 |
| 원본 v0.1.37 | OpenCode session header 누락 수정 공지 | 텍스트 한 번 응답이 아니라 후속 요청까지 검사 |
| 원본 v0.1.38 | Claude Code Artifact 도구 schema 관련 수정 공지 | CC 버전별 실제 도구 정의를 회귀 계약으로 관리 |
| 원본 v0.1.39 | OpenCode 카탈로그·stream 오류 처리 갱신 공지 | 모델 목록·late error를 지속 검증 |

위 릴리스 항목은 프로젝트 작성자의 수정 발표이며 MoAI에서 재현·통과한 결과가 아니다. [전체 변경 이력](https://claude-code-proxy.raine.dev/reference/changelog/), [공식 OpenCode 검증](https://opencode.ai/docs/go/).

## 19. 문서 21개별 검토 결과

| 문서 | 핵심 검토·반영 사항 |
|---|---|
| [개요](https://claude-code-proxy.raine.dev/) | 다중 공급자용 Anthropic 호환 프록시; 공식 서비스 권한과 구분 |
| [시작](https://claude-code-proxy.raine.dev/getting-started/) | 설치·로그인·serve·CC 환경 계층; 이번에 설치 안 함 |
| [작동 원리](https://claude-code-proxy.raine.dev/how-it-works/) | registry→auth→translate→stream, 상태·token count 경계 |
| [공급자 선택](https://claude-code-proxy.raine.dev/providers/choosing-a-provider/) | 계정·프로토콜·도구 차이; 동일 supported 판정 금지 |
| [Codex](https://claude-code-proxy.raine.dev/providers/codex/) | 구독 OAuth·WS·reasoning·continuation·compaction·이미지 |
| [Kimi](https://claude-code-proxy.raine.dev/providers/kimi/) | device 인증·기기 ID·Chat 변환; 소스 K3 추가 확인 |
| [Grok](https://claude-code-proxy.raine.dev/providers/grok/) | PKCE·Responses·search·effort·vision 기본 생략 |
| [OpenCode Go](https://claude-code-proxy.raine.dev/providers/opencode-go/) | key·모델별 3종 프로토콜·접두사 충돌 처리 |
| [Cursor](https://claude-code-proxy.raine.dev/providers/cursor-agent/) | Connect·번들 schema·동적 모델·Read/Write/Bash 제한 |
| [CC 설정](https://claude-code-proxy.raine.dev/using/configure-claude-code/) | main/small 모델·재시도·compact·환경 분리 |
| [모델 라우팅](https://claude-code-proxy.raine.dev/using/models-and-routing/) | alias·fast·[1m]·discovery 필터 |
| [전환](https://claude-code-proxy.raine.dev/using/switching-models-and-backends/) | base URL은 프로세스 재시작, 모델은 요청별 변경 |
| [모니터](https://claude-code-proxy.raine.dev/using/monitor-tui/) | 세션·usage·throughput·진행·로그·demo |
| [에이전트 안내](https://claude-code-proxy.raine.dev/using/for-coding-agents/) | 현재 catalog 조회·민감한 traffic·권한 경계 |
| [문제 해결](https://claude-code-proxy.raine.dev/using/troubleshooting/) | 400·429·인증·WS proxy·중복 tool·context |
| [명령 참조](https://claude-code-proxy.raine.dev/reference/command-reference/) | serve·models·login/device/status/logout·demo |
| [설정](https://claude-code-proxy.raine.dev/reference/configuration/) | CLI/env/config 우선순위·선택 기능·기본값 |
| [파일 저장](https://claude-code-proxy.raine.dev/reference/files-and-storage/) | OS별 auth·state·error·traffic 저장 영역 |
| [HTTP API](https://claude-code-proxy.raine.dev/reference/http-api/) | Messages·count·models·Responses·Chat·Images, 무인증 |
| [호환성](https://claude-code-proxy.raine.dev/reference/compatibility-and-limitations/) | 이미지·도구·search·unsupported fields·추정 usage |
| [변경 이력](https://claude-code-proxy.raine.dev/reference/changelog/) | 릴리스별 인증·이력·도구·모델·스트림 수정 사례 |

사이트 전체라는 범위는 사이트맵 21개 문서와 조립 본문이다. 링크된 모든 GitHub issue·PR·외부 웹사이트·이미지까지 전수 조사했다는 뜻은 아니다. 소스의 핵심 경로와 관련 공식 정책을 추가로 검토했다.

## 20. Claim · Evidence · Baseline-attribution

Claim: 문서 21개 대응과 정적 카탈로그 집계를 수행했다. OpenCode Go 모델 파일의 독립 단위 테스트 4개를 이번에 컴파일·실행해 통과를 관찰했다. 이는 네트워크·로그인·생성 테스트가 아니다.

Evidence — 아래 명령과 출력은 이번 조사에서 직접 관찰했다.

```text
moai session current
01a09dfb-5908-7e50-ac3a-bde14d9af79f

git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop rev-parse --short HEAD
4056f69e1

git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop branch --show-current
develop

node /Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop/reports/raine-proxy-gateway-design-20260914-01a09dfb/inventory.cjs
{"sitemap_pages":21,"mapped_pages":21,"codex_base_models":10,"kimi_public_ids":5,"grok_models":3,"opencode_models":36,"opencode_protocols":{"ChatCompletions":23,"Messages":9,"Responses":4}}
```

테스트 명령은 rustc --edition=2024 --test /tmp/raine-proxy-audit-iNNtqV/raine-claude-code-proxy-bcfa0a4/src/providers/opencode/model.rs -o /tmp/raine-proxy-audit-iNNtqV/opencode-model-tests 다음 해당 바이너리 실행이다.

```text
running 4 tests
test tests::canonical_prefix_resolves_and_unknown_models_do_not ... ok
test tests::supported_catalog_is_partitioned_by_wire_protocol ... ok
test tests::conflicting_provider_ids_are_only_advertised_with_prefix ... ok
test tests::refreshed_models_resolve_and_are_advertised ... ok

test result: ok. 4 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s
```

Baseline-attribution: Rust 테스트는 다운로드한 고정 upstream 소스의 opencode/model.rs만 대상이다. MoAI의 전체 테스트 결과가 아니다. develop에는 기존 SPEC·테스트·orchestration 미커밋 변경이 있었고 그대로 보존했다. 이 보고서의 제안 기능은 아직 구현되지 않았다.

HTML은 브라우저에서 21개 본문 섹션과 내부 앵커 누락 0을 확인했다. 1280px 화면은 문서 폭 1280px, 390px 화면은 문서 폭 390px였다. 두 화면의 스크린샷을 확인했으며 표는 좁은 화면에서 독립적으로 가로 스크롤한다. 전체 접근성·인쇄·모든 외부 링크의 성공을 검증한 것은 아니다. [화면 검사 근거](browser-qa.md)

## 21. Gaps · Residual-risk · 최종 권고

Gaps: 실계정 OAuth·refresh·추론·도구·3단계 Agent·Factory·resume·이미지·음성·Windows 런타임 검증은 NOT_RUN이다. Cursor 계정별 모델 목록, 각 모델 entitlement, GPT Lite·내부 이미지의 MoAI 통합 승인, 공급자별 계약 예외는 확인하지 못했다. X 게시물 원문 독립 확인에 실패했다. 원본 전체 소스의 보안 감사나 전체 Rust 테스트를 실행하지 않았다.

Residual-risk: 공급자 내부 endpoint·OAuth·모델 catalog·약관은 바뀔 수 있다. 모형별 tool semantics·reasoning·context 한도가 다르고 요약 복구는 정보를 잃을 수 있다. 정책에 맞는 연결이라도 품질·속도·연속성은 별도로 검증해야 한다. 개인 구독은 공유 Factory 서버용 무제한 자원으로 취급할 수 없다.

최종 권고: MoAI Gateway를 모든 공급자의 등록·선택·검증을 통합하는 계층으로 만들고, 사용자는 기존 CC UI와 명령을 유지한다. GPT 메인은 직접 구독 adapter·App Server adapter·공식 API·OpenCode Go를 구분해 선택한다. GLM과 Kimi 공식 Coding Plan, OpenCode Go를 경제형 실행에 연결하고, 이미지와 리뷰는 같은 정책을 공유하는 MCP 입구를 둔다. 승인되지 않은 인증 방식이나 검증되지 않은 기능은 목록에 표시하되 실행 기본값으로 승격하지 않는다.

따라서 “모두 흡수”는 가능하도록 설계하되 “전부 ToS 무위험·동일 품질·동일 속도”라는 약속은 하지 않는다. 사용자 편의성은 프로필과 일관된 진단으로, 경제성은 완료 작업당 비용으로, 품질은 실제 도구·이력 인수로 확보한다.
