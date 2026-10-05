# moai gpt 메인 모델 연결 — 허용 경로와 품질·속도 재검토

2026-09-14 조사 · Claude Code 메인 추론은 GPT · 제품 변경 없음

## 1. 결론 — 메인 GPT는 가능성이 있지만, 현재 무문제 판정은 불가

이번 요구는 Claude가 GPT에 전문 작업을 맡기는 방식이 아니다. moai gpt가 실행한 Claude Code에서 첫 응답부터 도구 선택·후속 추론까지 GPT가 담당해야 한다. MCP 위임형, OpenCode로 UI 교체, Codex CLI만 실행하는 형태는 요구 충족 대안에서 제외한다.

권고 후보는 Claude Code → MoAI Messages Bridge → 공식 Codex App Server → GPT다. 본인 ChatGPT 구독의 로그인·토큰 갱신·모델 통신은 공식 Codex가 담당하고, Claude Code가 실제 파일·셸 도구 실행을 맡도록 역중개한다. 이는 GPT 메인 세션이다. Claude 모델이 감독자로 추가 호출되는 구조가 아니다.

그러나 다음 세 조건을 동시에 만족하는 검증된 완제품은 이번 조사에서 확인하지 못했다: ChatGPT 구독, Claude Code GPT main, 기존 기능·품질·속도 문제 없음. 공식 구성요소가 있다는 사실만으로 전체 조합에 대한 양사 승인이나 성능 동등성을 주장할 수 없다.

판정: 요구 적합성은 조건부 가능. 공개 근거 기반 구독 후보는 공식 App Server. 출시 준비는 NOT READY. 공식 API 키 경로는 별도 대안이며 구독 요구를 충족한 것처럼 바꿔 말하지 않는다.

## 2. 합법성·약관·공식 지원은 서로 다른 판단

오픈소스 라이선스, 인증 성공, 서비스 약관, 제조사 기술 지원은 다른 문제다. 공식 비지원은 곧 불법이라는 뜻이 아니며, 공식 인증 API 존재도 모든 용도의 허가를 뜻하지 않는다. 이 보고서는 법률 의견서가 아닌 공개 문서 기반 기술·정책 검토다.

| 경로 | GPT main | ChatGPT 구독 | 정책·지원 판단 |
|---|---|---|---|
| 공식 App Server + Messages Bridge | 설계상 충족 | managed login 경로 있음 | 우선 검증 후보. 전체 조합의 명시적 승인·운영 보증 미확인 |
| 공식 Responses API + Messages Bridge | 설계상 충족 | 아니오, API 별도 과금 | OpenAI 접근 계약이 명확한 대안. CC 비-Claude 지원 한계는 동일 |
| 구독 토큰 직접 추출 + 내부 backend 요청 | 커뮤니티 구현 존재 | 사용 사례 주장 있음 | 공식 허용 근거·안정성 부족. 기본 제품안에서 제외 |
| Claude main + Codex MCP/플러그인 | 불충족 | 가능 경로 있음 | 이번 요구의 답이 아님 |
| OpenCode·OMP·Codex 자체 메인 | GPT 가능 | 각 경로별 상이 | Claude Code 필수 조건 불충족 |

OpenAI App Server의 chatgpt managed 모드는 공식 Codex가 OAuth·보관·갱신을 소유한다. 외부 토큰 모드는 별개의 실험적 기능이므로 MoAI가 구독 토큰을 직접 다룰 이유로 삼지 않는다. [공식 인증 프로토콜](https://learn.chatgpt.com/docs/app-server#authentication-modes).

Anthropic 문서는 어떤 Gateway를 쓰든 비-Claude 모델로 라우팅하는 것을 지원하지 않는다고 명시한다. API 키로 바꿔도 이 지원 경계는 사라지지 않는다. [공식 Gateway 안내](https://code.claude.com/docs/en/llm-gateway).

OpenAI ROW 약관의 계정 공유·보호 조치 우회 금지 등은 그대로 적용된다. 무위험이 필수 승인 기준이라면 제품의 정확한 연결·개인 사용 범위·인증 소유권을 양사에 제시하고 명시적 확인을 받아야 한다. 이 외부 확인은 이번에 수행하지 않았다. [OpenAI 약관](https://openai.com/policies/row-terms-of-use/).

또한 App Server 제거 안내 페이지는 app-server를 experimental·production 비지원으로 설명한다. 상세 App Server 문서는 stable/experimental 기능 구분을 제공하지만, 필요한 dynamicTools는 명시적으로 experimental이다. 따라서 상세 문서가 있다는 이유로 production 지원이 확정됐다고 해석하지 않는다. [MCP 제거와 App Server 안내](https://learn.chatgpt.com/docs/mcp-server), [dynamic tools](https://learn.chatgpt.com/docs/app-server#dynamic-tool-calls-experimental).

## 3. GLM처럼 단순 연결하면 되는가

사용자 경험은 GLM처럼 만들 수 있어야 하지만 내부 기술은 같지 않다. GLM의 공급자 제공 Anthropic 호환 endpoint와 Codex의 JSON-RPC 에이전트 실행기는 서로 다른 인터페이스다. GPT 구독 연결은 URL·모델명 변경만으로 완성되지 않는다.

사용자 경험의 목표는 moai gpt 실행 → 필요한 경우 공식 로그인 → 모델·effort 선택 → GPT main 시작이다. 내부에서는 세션 식별, 도구 중개, streaming, 중단·재개, fork·compact를 처리한다. 자동으로 Claude main을 대신 실행하거나 API 과금으로 전환하지 않는다.

프록시 프로젝트도 Claude Code의 실용적인 API 사용 범위가 목표이지 완전한 프로토콜 동등성은 아니라고 설명한다. 따라서 GLM의 기존 체감 안정성은 이번 GPT 구조가 동등함을 증명하는 baseline이 아니다. [claude-code-proxy 호환 범위](https://claude-code-proxy.raine.dev/reference/compatibility-and-limitations/).

## 4. 권고 구조 — 추론은 GPT, 도구 실행은 Claude Code

1. Claude Code가 사용자 메시지·도구 schema를 MoAI의 로컬 Messages endpoint로 보낸다.
2. MoAI가 인증된 session/fork/lane을 공식 Codex thread와 연결한다. 텍스트의 가짜 thread ID를 신뢰하지 않는다.
3. Codex가 GPT 추론을 수행한다. Claude 도구는 검증된 dynamicTools로 노출한다.
4. GPT의 도구 요청을 Claude tool_use로 반환하고, 해당 Codex RPC는 결과 대기 상태를 유지한다.
5. Claude Code가 기존 승인·hook·도구를 실행한다. tool_result를 동일 RPC 요청에 정확히 한 번 전달한다.
6. GPT가 같은 turn에서 계속 추론한다. 실제 turn 완료 또는 도구 경계에서만 적절한 Claude 종료 상태를 반환한다.

단순히 매 요청마다 codex exec에 전체 대화를 전달하면 공식 실행기를 쓰더라도 문맥 재처리·중복 도구 루프 문제가 생길 수 있다. 제안 구조에서는 프로세스와 thread를 유지하고 필요한 delta만 전달한다. 단, 이 설계의 절감 효과는 아직 측정하지 않았다.

Codex 내장 shell·patch가 Claude 도구를 우회하지 못하도록 실행 경로를 제한해야 한다. read-only sandbox만으로 모든 내장 읽기·검색·외부 호출이 사라진다고 가정하지 않는다. 지원 설정과 실제 이벤트를 검증하고, 미승인 도구 경로는 차단한다. 이 조건을 충족할 수 없으면 모든 도구가 Claude 승인·hook을 통과한다는 주장을 철회해야 한다.

모델·effort는 독립적으로 전달하며, 실제 계정 catalog와 지원 범위로 검증한다. Claude 지침·MoAI 규칙·도구 설명의 의미는 유지하되 Codex 기본 지침과 중복·충돌하는 부분은 역할별로 비교한다. GPT 모델 사용이 곧 Codex native와 같은 품질이라는 등식은 성립하지 않는다.

## 5. 현재 코드에서 확인한 차이

| 관찰 | 실제 근거 | 의미 |
|---|---|---|
| App Server 어댑터 존재 | internal/gateway/appserver.go | 구독 연결 후보의 구성요소가 있음 |
| 생산 진입점은 기존 factory에 연결 | internal/cli/root.go:164, gateway_product_binding.go:198–255 | 현재 읽은 함수는 CodexBroker·PolicyGPTNative 경로. 어댑터 존재만으로 실제 moai gpt 경로 전환 완료 아님 |
| 응답 구간을 모은 뒤 SSE 조립 | appserver.go:60–79, codexbridge/engine.go:574–579 | 점진적인 token 표시 속도를 현재 코드로 주장할 수 없음 |
| usage 값 0으로 응답 조립 | appserver.go의 message/start/delta | 실측 사용량으로 해석 불가. 0과 미확인을 구별하는 계측 필요 |
| 도구 경계에서 turn 유지 | codexbridge/engine.go:347 이후 | 방향은 요구와 맞음. fixture 계약과 실제 Claude 동작은 별도 |
| 기존 history 오류 문자열 남음 | translate/receipt_history.go:40 | 과거 오류 분류를 재검증할 지점. 이번에 같은 live 400을 재현한 것은 아님 |

rg로 NewAppServerAdapter의 사용 위치와 productionGatewayHandlerFactory의 연결을 찾고 함수 본문을 읽었다. 이것은 현재 inspected tree의 정적 경로 관찰이며 설치된 바이너리의 live 경로 검증이 아니다. 신규 untracked 계약 테스트도 있어 다른 세션의 작업으로 보존했다.

## 6. 기존 문제를 다시 만들지 않는 필수 계약

| 영역 | 피해야 하는 처리 | 필수 검증 |
|---|---|---|
| reasoning·history | 숨겨진 GPT continuation을 Claude thinking 텍스트로 복구 | 공식 runtime 상태 유지, 순서·도구 ID 보존 |
| fork·agent_summary | 다른 세션을 같은 thread에 붙이거나 중복 전체 replay | 명시적 자식 소유권·prefix 계약, 부모/자식 분리 |
| compact | Claude와 Codex가 동시에 무관한 압축 수행 | 하나의 논리적 압축 경계, 요약·이력 적용 순서 검증 |
| resume·재시작 | 대기 중 RPC를 완료로 처리하거나 도구 자동 재실행 | idle/waiting/unknown 구분, 결과 대조 후 복구 |
| model switch | 도구 결과 대기 중 모델·계정 교체 | 안전한 경계에서만 변경, 실제 모델 표시 |
| tool_use·tool_result | 오류를 성공 텍스트로 바꾸거나 call ID 재사용 | 거부·오류·중복·교차 thread 오염 테스트 |
| 권한·trust | 프록시 편의를 위해 bypass를 기본값으로 강제 | Claude 승인·hook 보존, 경로별 권한 검증 |
| 환경·Factory | tmux 전역 auth/endpoint 덮어쓰기 | session/lane별 환경·수명·동시성 분리 |
| 이미지·MCP | 텍스트 변환 성공을 모든 multimodal 지원으로 확대 | 이미지 파일·도구 결과·취소 경로 별도 검증 |

App Server는 인증·대화 상태의 일부를 공식 실행기에 맡길 수 있게 하지만, Claude transcript와 Codex thread의 대응 관계까지 자동 해결하지 않는다. 기존 history 오류가 이름만 바뀌어 재현되지 않도록 경계 테스트가 필요하다.

API 대안 역시 reasoning continuation을 버리면 안 된다. 공식 Responses 문서는 암호화된 reasoning 등 재생 가능한 output item의 보존을 안내한다. 따라서 API 키로 바꾸기만 하면 품질·이력 문제가 사라진다는 결론도 틀리다. [공식 reasoning 상태 안내](https://developers.openai.com/api/docs/guides/reasoning).

## 7. 속도 — SSE라는 이름보다 실제 첫 글자 도착

현재 어댑터는 Engine.Step이 구간을 반환한 다음 SSE 이벤트를 메모리 버퍼에 조립한다. 이는 코드로 관찰한 동작이다. 실제 지연 수치는 측정하지 않았다.

권고는 인증·모델·schema 검증 후 안전한 텍스트 delta를 즉시 flush하는 것이다. 도구 호출 확정과 완료 상태는 검증·기록 뒤 전송한다. 스트리밍 일부를 보낸 뒤 오류가 나면 성공 종료를 위조하지 않고 중단 상태로 보고한다. 이미 보낸 요청을 다른 전송 방식으로 재실행해 도구 효과를 중복시키지 않는다.

최적화는 고정 프로세스 재사용, 연결 재사용, 모델 catalog 캐시, delta 입력, lane별 backpressure 순으로 검토한다. 매 응답에 전체 history를 다시 읽거나 여러 번 직렬화하는 경로는 trace로 비용을 분리한다. 사고 복구 기록을 없애는 식의 속도 개선은 금지한다.

측정값은 첫 upstream delta까지의 시간, bridge 처리부터 downstream flush까지의 시간, 첫 화면 표시, 전체 task 완료, 성공률로 나눈다. 같은 모델·effort·작업·계정·cache 조건의 official Codex와 비교하고, bridge 자체는 기록된 이벤트 fixture로 독립 측정한다. native Codex와 Claude harness의 품질 차이는 adapter 지연과 분리한다.

제안 출시 기준: 기능 회귀·미승인 도구 실행·중복 side effect는 테스트 corpus에서 0건. p50/p95 지연·성공 작업당 비용·수락률을 보고하고 허용 오차는 실측 baseline 이후 사용자와 확정한다. 아직 숫자를 정하지 않은 항목은 PASS가 아니라 기준 미확정이다.

## 8. 커뮤니티 구현에서 배울 점과 채택하지 않을 점

| 프로젝트 | 이번에 확인한 내용 | 판단 |
|---|---|---|
| raine/claude-code-proxy | 모델 라우팅·진단·streaming, 상세 제한 문서 제공 | 구조·관찰성 참조. 완전 호환·계정 무위험 근거는 아님 |
| karem505/claudecodex | GPT main launcher, process-local env, 기본 권한 bypass와 effort 제한 안내 | 격리 UX는 참고, bypass 기본값과 한계 은폐는 채택하지 않음 |
| hanjun-lin/claude-gpt | 내부 Codex backend와 client 사칭의 약관 위험을 README가 직접 경고, API 모드 제시 | 구독 기본안에서 제외. 저자의 경고를 법적 확정 판결로 확대하지 않음 |

근거: [raine 저장소](https://github.com/raine/claude-code-proxy), [claudecodex](https://github.com/karem505/claudecodex), [claude-gpt](https://github.com/hanjun-lin/claude-gpt/blob/main/README.md).

raine의 문서도 provider별 continuation, local token estimate, 일부 필드 생략, 모델 변경 시 상태 차이를 인정한다. public endorsement 언급은 있지만 이번 조회에서 해당 발언의 원문·적용 범위를 독립 검증하지 않았다. 이를 공급자의 계약상 승인으로 인용하지 않는다. 설치·계정 로그인·실행 속도 비교는 하지 않았다.

## 9. 구현을 진행한다면 이 순서

P0 정책 확인: 공식 managed login만 사용한다는 범위, 개인 개발과 배포 제품의 범위를 문서화한다. 엄격한 무위험 조건에는 양사의 구체적 확인이 필요하다. 비지원 경계를 수용할 수 없다면 현재 조건 전체를 충족하는 출시안은 없다.

P1 단일 main 수직 검증: moai gpt 진입부터 실제 GPT 응답, Claude Read/Bash 도구 승인, 결과 후속 추론까지 하나의 실제 session에서 검증한다. mock production counter나 adapter 단위 테스트만으로 통과시키지 않는다.

P2 이력·권한: 긴 대화, fork·summary, compact, resume, 취소, 재시작, 모델 변경, 이미지 포함 tool_result를 실행한다. known error fixture와 실계정 재현을 분리한다.

P3 속도·품질: 실제 incremental flush를 확인하고 동일 과제로 비교한다. UI 응답성뿐 아니라 요구 충족·회귀·반복 수정량을 평가한다.

P4 Factory: 단일 세션을 통과한 뒤 lane별 thread·account·권한·취소 격리를 검증한다. 한 lane의 실패를 다른 lane으로 전파하지 않는다.

P5 배포: 검증된 Claude Code·Codex 버전 조합을 기록하고 업데이트 시 회귀를 실행한다. 문제가 생기면 GPT 경로를 안전하게 중단한다. 사용자 몰래 Claude main 또는 유료 API로 교체하지 않는다.

GPT 이미지 생성 요구는 유지한다. 다만 main GPT 성공과 이미지 생성은 다른 합격 항목이다. 공식 생성 경로의 파일 회수·과금·권한을 별도 검증하며 프록시의 이미지 passthrough 광고만으로 구독 허용성을 확정하지 않는다.

## 10. Claim · Evidence · Baseline-attribution

Claim: 요구에 맞는 구독 후보 구조를 특정했고, 공개 문서의 지원 경계와 현행 코드의 구간 buffering을 확인했다. 선택한 gateway/bridge 계약 테스트 7개가 통과했다. 품질·속도 동등성·live 연결·법률상 무위험은 주장하지 않는다.

Baseline: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop, HEAD 4056f69e1c20d942d4f9fc7363d3d79bffde899a. 실행 전후 HEAD가 같았다. 수정된 SPEC와 신규 untracked 파일이 있는 작업 트리 기준이며 깨끗한 commit 검증이 아니다. 제품 파일을 수정하지 않았다.

명령 A:

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN OPENAI_API_KEY Z_AI_API_KEY && GOCACHE=/tmp/moai-gpt-main-research-cache go test ./internal/gateway -run '^(TestAppServerTextAndToolSSEBoundaries|TestAppServerAdapterRejectsUntrustedBindingsBeforeTurn|TestManagedAuthorityRejectsModeFallbackAndGenerationRace)$' -count=1 -v -timeout 60s
```

실제 출력 중 상위 테스트·종료 행 발췌. 하위 권한 거부 8개 subtest도 PASS였으며, 전체 로그를 재구성한 것으로 표현하지 않는다.

```text
=== RUN   TestAppServerAdapterRejectsUntrustedBindingsBeforeTurn
--- PASS: TestAppServerAdapterRejectsUntrustedBindingsBeforeTurn (0.00s)
=== RUN   TestAppServerTextAndToolSSEBoundaries
--- PASS: TestAppServerTextAndToolSSEBoundaries (0.00s)
=== RUN   TestManagedAuthorityRejectsModeFallbackAndGenerationRace
--- PASS: TestManagedAuthorityRejectsModeFallbackAndGenerationRace (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.476s
```

명령 B:

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN OPENAI_API_KEY Z_AI_API_KEY && GOCACHE=/tmp/moai-gpt-main-research-cache go test ./internal/codexbridge -run '^(TestToolSegmentRetainsRPCAndCompletesExactlyOnce|TestResumeAttachesOwnedThreadWithoutRebuild|TestForkChildStartsNewThreadAtInheritedPrefix|TestCompactionIsANormalSummaryTurnWithZeroCompactRPCs)$' -count=1 -v -timeout 60s
```

```text
=== RUN   TestCompactionIsANormalSummaryTurnWithZeroCompactRPCs
--- PASS: TestCompactionIsANormalSummaryTurnWithZeroCompactRPCs (0.16s)
=== RUN   TestToolSegmentRetainsRPCAndCompletesExactlyOnce
--- PASS: TestToolSegmentRetainsRPCAndCompletesExactlyOnce (0.20s)
=== RUN   TestForkChildStartsNewThreadAtInheritedPrefix
--- PASS: TestForkChildStartsNewThreadAtInheritedPrefix (0.28s)
=== RUN   TestResumeAttachesOwnedThreadWithoutRebuild
--- PASS: TestResumeAttachesOwnedThreadWithoutRebuild (0.24s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/codexbridge	1.241s
```

이 테스트는 stub/fixture 계약 검증이다. SSE 경계 테스트는 완성된 segment를 조립한 응답을 읽으며 실제 첫 token 지연을 측정하지 않는다. 기존 메모리의 agent_summary 400은 조사 지점을 정하는 데만 사용했고 이번 장애 재현으로 인용하지 않았다.

## 11. Gaps · Residual-risk · 최종 판단

Gaps: 양사 서면 확인, 실계정 로그인·GPT main 호출, installed binary 경로, 속도·품질 benchmark, 이미지 생성, Factory 운영은 수행하지 않았다. 커뮤니티 프로젝트 설치·테스트도 하지 않았다. HTML은 기존 렌더링 자산을 재사용하며 누락된 design-tokens.md 대신 스킬 본문 토큰을 따른다.

Residual-risk: 공식 구성요소의 실험적 API와 Claude Code 비지원 경계가 남는다. 두 runtime 사이의 history·도구 의미 차이는 프로토콜 번역만으로 제거되지 않는다. 현재 통과한 국소 테스트가 production 연결 또는 실제 성능을 보증하지 않는다.

최종 판단: moai gpt를 GPT 메인으로 유지하는 요구는 유지한다. 구독 우선 후보는 공식 Codex App Server 기반 Messages Bridge, API 대안은 공식 Responses API 기반 Bridge다. 토큰 직접 재사용 프록시로 무위험을 주장하지 않는다. 기존 문제 없는 완제품이라는 결론은 현재 근거로 낼 수 없으며, 위 검증을 통과하기 전에는 출시 준비 미완료로 판정한다.
