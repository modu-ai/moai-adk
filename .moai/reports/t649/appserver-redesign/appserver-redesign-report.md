# MoAI GPT · Codex App Server 전환 설계 보고

2026-09-12 · 공식 문서: https://learn.chatgpt.com/docs/app-server · 상태: 설계 및 초기 실증, 제품 전환 미완료

## 실행 결정과 검증 상태

구독 경로는 공식 Codex App Server를 사용하고, Claude Code 화면과 제공자별 /model 제한을 유지한다. GPT API 키 경로도 함께 지원한다. 사용자가 승인한 순서는 HTML 보고 → 카드 발행 → 구현·검증이다.

설치 Codex 0.154.0에서 실제 구독 계정으로 dynamicTools → item/tool/call → 도구 결과 → 최종 표식 답변을 두 차례 확인했다. 두 번째 시험은 environments:[]와 hook·도구 비활성 설정을 사용했다. hook 이벤트는 없었고 native 도구 실행도 관측되지 않았으나 MCP 시작 상태 이벤트 1건의 의미와 전체 native 도구 차단은 미확인이다. 성공한 단일 왕복을 제품 전체 성공으로 표시하지 않는다.

- [첫 실제 도구 왕복 증거](probe/dynamic-result.json)
- [격리 설정 왕복 증거](probe/isolation-result.json)
- [설치 바이너리 스키마 조사](probe/schema-findings.json)
- [400 오류 원인 및 보존 패치](../subagent400/verdict.md)

직접 backend 경로의 tool_reference 결함은 최소 패치와 단위 테스트 증거로 보존했다. 새 App Server 제품 전환이나 로컬 배포를 완료한 상태는 아니다.

## 도구 검색 실증에 따른 설계 보정

SPEC-MOAI-GATEWAY-001 0.11.0에 공식 App Server 전환 계약을 반영했다. 아래 상세 설계는 최초 제안의 기록이며, 현재 계약은 SPEC 0.11.0과 이 보정 결과를 따른다.

Claude Code 2.1.269와 로컬 모의 Anthropic 서버로 실제 요청을 포착했다. ToolSearch 활성 상태에서는 첫 요청의 12개 도구에 시험용 MCP echo의 상세 정의가 없었으며, ToolSearch 뒤 두 번째 요청에서 정의와 tool_reference가 추가되었다. 따라서 첫 요청만으로 전체 도구를 선등록한다는 전제는 해당 시험에서 성립하지 않았다.

ToolSearch 비활성 대조군은 첫 요청에 22개 도구와 echo의 상세 정의를 전달했다. 구조 필드의 JSON 크기는 10,998바이트에서 17,295바이트로 늘었다. 이 시험은 실제 토큰 비용이나 일반 프로젝트의 도구 수를 측정하지 않았다. 비활성화는 도구 검색 흐름을 변경하므로 자동 채택하지 않았다. 검색을 보존하는 공통 호출 통로는 실제 GPT-5.6 Sol 구독의 단일 턴에서 검색→새 도구 호출→최종 표식 답변을 완료했다. 두 호출의 이름과 인수를 시험용 정의로 검증했다. 이는 단순 픽스처의 실행 가능성 증거이며, 실제 Claude Code 통합이나 일반 도구 선택 품질의 증거는 아니다. 초기 도구의 원래 정의를 유지하고 새로 발견한 도구만 공통 통로로 연결하는 방식은 2026-09-12 사용자가 승인했다.

증거: [활성·비활성 비교](probe/claude-tool-surface-comparison.json), [활성 요청](probe/claude-tool-surface-result.json), [비활성 요청](probe/claude-tool-surface-disabled-result.json).

기반 t650 구현은 진행 중이며, 도구 연결 t651 이후의 제품 통합과 로컬 배포는 미완료다.

## 상세 설계

작성: 2026-09-12. 카드: t649. 이 문서는 설계 제안이며 canonical SPEC 변경이나 구현 완료 기록이 아니다.
대상: moai-proxy-unified / WT-unified-gateway / 81c1d58f9 및 기존 미커밋 작업.

## 결정할 핵심

Claude Code 화면과 `/model`을 유지하고 GPT 구독의 실행·인증·reasoning 이력 소유권을 Codex App Server로 옮긴다.
MoAI는 Claude Messages와 App Server thread/turn/item을 잇는 어댑터를 담당한다. Codex UI로 대체하지 않는다.
cc는 Claude, glm은 GLM, gpt는 GPT라는 세션 허용 목록은 유지한다. GPT 구독과 API 키는 명시적으로 구분한다.

가장 먼저 검증할 병목은 Claude의 도구 실행권을 유지하는 동적 도구 왕복이다. 공식 App Server에는 이를 위한
`dynamicTools`와 `item/tool/call`이 있지만 실험 API다. 프로토콜 존재는 MoAI 전체 왕복 성공의 증거가 아니다.
[공식 App Server 문서](https://learn.chatgpt.com/docs/app-server#dynamic-tool-calls-experimental)

## 관측 근거

로컬 Codex 소스는 `/tmp/openai-codex-audit.zbvNXB`, 읽은 HEAD는 다음과 같다.

```text
git -C /tmp/openai-codex-audit.zbvNXB rev-parse HEAD
5a9eb145c4c05fcfc7158d7c25b80e1322eccae1
```

다음은 이 커밋의 소스 판독이다. 해당 Rust 시험을 이번 작업에서 실행하지 않았다.

| 근거 | 확인한 구조 | 설계 영향 |
|---|---|---|
| `app-server-protocol/src/protocol/v2/thread.rs:138-144` | `thread/start.dynamicTools`가 실험 필드 | 시작 때 도구 목록 확정·기능 협상 필요 |
| 같은 파일의 dynamic_tools 검색 | 선언은 start에만 존재 | turn/resume에서 임의 목록 교체를 가정하지 않음 |
| `protocol/src/dynamic_tools.rs:20-43` | 이름·설명·input_schema·defer_loading 및 namespace | 최초 전체 schema를 받으면 원형 등록이 우선 |
| `app-server-protocol/src/protocol/v2/item.rs:1642-1672` | 요청에 thread/turn/call ID, tool·arguments; 응답 text/image/audio와 success | Claude tool_use/tool_result 상관관계 필요 |
| `app-server/src/dynamic_tools.rs:21-59` | 클라이언트 응답을 기다렸다가 DynamicToolResponse 제출 | HTTP 요청이 끝나도 App Server turn/RPC를 유지해야 함 |
| `app-server-protocol/src/protocol/v2/turn.rs:156-215` | turn 시작 입력·model override | `/model`은 다음 turn의 model에 대응 가능 |
| `app-server-protocol/src/protocol/common.rs:571,708` | thread/fork, thread/compact/start 등록 | fork/compact는 별도 비동기 수명 |
| `app-server/tests/suite/v2/dynamic_tools.rs:341` | dynamic tool text 왕복 시험 존재 | 제품 재현 시험의 형태 참고, 실행 PASS 주장 아님 |
| `core/src/tools/spec_plan.rs:1079` 및 `spec_plan_tests.rs:1016` | ShellTool 비활성 시 shell registry 제외 분기와 시험 | approval never만으로 native 도구가 꺼지지 않음 |

고정 소스 링크: [thread 계약](https://github.com/openai/codex/blob/5a9eb145c4c05fcfc7158d7c25b80e1322eccae1/codex-rs/app-server-protocol/src/protocol/v2/thread.rs),
[동적 도구 응답](https://github.com/openai/codex/blob/5a9eb145c4c05fcfc7158d7c25b80e1322eccae1/codex-rs/app-server/src/dynamic_tools.rs),
[동적 도구 시험](https://github.com/openai/codex/blob/5a9eb145c4c05fcfc7158d7c25b80e1322eccae1/codex-rs/app-server/tests/suite/v2/dynamic_tools.rs).

## 최소 구성과 재사용

기존 loopback 인증·제공자별 picker·MoAI launcher·대화 소유권 경계를 유지하고 GPT 구독 adapter 뒤만 App Server로 바꾼다.
직접 구독 backend 요청과 MoAI의 구독 토큰 읽기·refresh는 이 경로에서 제거한다. Codex managed auth가 소유한다.
로그인 화면의 세 선택은 제안이다: ChatGPT 브라우저 managed / ChatGPT device-code managed / OpenAI API 키.
앞의 둘은 같은 구독 과금 경로의 로그인 방식 차이이며 세 제공자를 뜻하지 않는다. 기존 제품의 관측된 메뉴라고 표시하지 않는다.
API 키 모드는 별도 명시 선택으로 지원하고 Codex의 apiKey 모드를 우선 검토한다. 구독 실패 시 API 과금으로 자동 전환하지 않는다.
[공식 인증 모드](https://learn.chatgpt.com/docs/app-server#authentication-modes)

`internal/cli/mcp_codex.go:425` 이후에 NDJSON stdio 연결, spawn/close, handshake와 재사용 session handle이 이미 있다.
`codexConn`의 send/recv/close, 직렬 송신 lock, bounded scanner, PID 소유권 및 cancel ID 관리가 재사용 후보다.
`codex_session_test.go`, `codex_live_protocol_probe_test.go`도 존재한다. 보고된 `codex_session.go`는 이 트리에서 없었다.
review 전용 완료 판독을 GPT adapter에 가져오지 않는다. 중립 RPC 계층을 추출할 때 기존 review 소비자의 회귀를 실행한다.
현재 단일 recv 흐름에 서버 요청·알림·응답의 ID별 분배와 backpressure·취소를 더해야 하며 별도 중복 프로세스 프레임워크는 만들지 않는다.

## 도구 목록과 ToolSearch

우선안 A는 최초 Claude 요청에 deferred 도구의 전체 input_schema가 실제로 들어오는지 캡처로 확인한 뒤,
전체 정의를 이름 역매핑과 함께 thread/start에 등록하는 것이다. deferLoading/namespace 지원 여부는 설치 버전으로 검사한다.
모델이 도구를 호출하면 MoAI가 원래 Claude 도구 이름으로 tool_use를 돌려주고 Claude가 승인·실행한다.
ToolSearch 결과의 tool_reference는 타입이 있는 발견 메타데이터로 처리하며 일반 text 변환기에 밀어 넣지 않는다.
발견된 이름은 현재 대화의 실제 등록 schema와 대조한다. 새 도구·schema가 뒤늦게 추가되는 경우의 정책을 별도 시험한다.

A가 불가능할 때의 후보 B는 고정 dispatcher 하나에 `{name, arguments}`를 전달하고 최신 Claude 도구 registry로 검증하는 방식이다.
동적 등록 교체가 필요 없어지지만 모델에 제공되는 도구별 JSON schema의 수가 N개에서 dispatcher schema 1개로 줄어든다.
도구별 schema를 설명 텍스트로 제공해도 native schema-guided 호출과 동등하다고 주장할 수 없다. 따라서 B는 자동 채택하지 않는다.
동일 도구 corpus로 잘못된 인수·이름·선택·추가 왕복을 대조하고 판단 근거를 보고해야 한다. 아직 실행 비교 수치는 없다.
루트가 별도 실행 담당에게 받은 dynamic tool 왕복 PASS는 이 문서 작성자가 직접 실행한 증거가 아니며,
그 실행에 hooks/MCP가 상속되었다는 보고도 있다. 따라서 단일 도구 왕복과 실행권 격리 통과를 분리해야 한다.
`environments: []`, hooks 비활성, MCP 비활성은 다음 격리 프로브의 후보이며 지원 확인 전 확정 설정으로 쓰지 않는다.
어느 안이든 Codex shell/file/MCP/agent 자체 실행을 차단하고 Claude 실행만 허용하는 native 도구 inventory·실행 0 증거가 필요하다.
readonly sandbox나 approvalPolicy never는 그 증거를 대신하지 않는다. thread/shellCommand도 호출하지 않는다.

## thread/turn/item 대응

| Claude/MoAI 사건 | 대응 제안 | 실패 경계 |
|---|---|---|
| 첫 사용자 turn | 소유권을 확인한 대화에 thread/start 후 turn/start | 모델·auth·tool 목록 검증 전 생성 요청 금지 |
| 텍스트 응답 | App Server agentMessage 이벤트를 Messages SSE로 변환 | turn ack를 응답 완료로 표시 금지 |
| tool 호출 | item/tool/call의 ID를 보관하고 Claude tool_use로 응답 | HTTP 종료와 App Server turn 종료를 혼동하지 않음 |
| tool_result 다음 HTTP 요청 | 소유 대화·call·결과를 검증해 보관된 RPC에 응답 | 중복·foreign·변조 결과를 다른 turn에 주입 금지 |
| tool 이후 최종 응답 | 이어지는 item/turn 완료를 같은 사용자 turn에 연결 | tool 실행 재시도·중복 비용을 조용히 발생시키지 않음 |
| 정상 새 사용자 메시지 | 이미 반영한 prefix 이후의 새 입력만 turn/start | 전체 Claude history를 매번 다시 넣지 않음 |
| 프로세스 재시작 뒤 재개 | 저장된 소유 thread ID로 thread/resume | 불안정 history 주입이나 raw reasoning 재생 금지 |
| 미완료 tool 대기 중 crash | pending 실행 여부를 판정해 복구 불가 시 명시 실패 | 사라진 RPC ID를 새 프로세스에 답하거나 도구 재실행 금지 |
| `/model` GPT 변경 | idle 경계에서 turn model override, 지원 목록 확인 | active tool 도중 변경 및 미검증 family 호환을 숨기지 않음 |
| compaction | 검증된 compact 사건을 thread/compact/start와 대응 | `{}` ack를 완료로 세지 않음; Claude와 Codex의 이중 압축 금지 |
| Claude Agent/fork | 실제 agent 식별·공통 prefix 확인 뒤 thread/fork 또는 새 thread | 이름/텍스트만으로 부모 thread를 추측하지 않음 |

thread/turn/call ID, 반영 prefix digest, 결과 수신·소비 상태를 대화별로 원자 저장한다. 이는 RPC 재생 권한을 주는 토큰이 아니다.
reasoning·암호문은 Codex가 관리하고 MoAI는 raw 내부 history 조회·복원을 정상 경로로 요구하지 않는다.
Claude가 `/compact` 후 history를 바꾼 경우를 자동으로 일반 새 입력으로 취급하지 않는다. 실제 요청의 압축 식별 표면과
prefix 변경 판정을 먼저 캡처해야 한다. 표면이 없으면 명시적인 MoAI compact 명령 연결이 필요하며 지원으로 가장하지 않는다.

## context와 기능 표시

이전 direct endpoint에서 관측한 최대 921k 입력은 App Server 수용 한도로 승계하지 않는다.
오케스트레이터가 보고한 모델 metadata 872k는 출처가 다른 값이며, 설치 App Server model/list와 실제 turn 결과로 다시 판정한다.
UI에는 모델 명목 창, 현재 경로의 유효 한도, 누적 사용량을 구분한다. 1M 표시를 근거 없이 붙이지 않는다.
구독 출력은 서버 정책을 따르며 Claude max_tokens와 같은 생성량 보장을 주장하지 않는다.

## 의존 카드 제안

아래 항목은 아직 발급된 카드가 아니다. 각 카드의 근거와 선행 관계를 HTML 검토 후 실제 큐에 등록한다.
현재 미커밋 gateway 코드 의존 때문에 새 원격 기준 worktree에 이미 구현이 있다고 가정해서는 안 된다.
우선 기존 승인 t649 worktree의 추적 하위 카드로 관계를 기록하고 작업 소유 파일을 분리하는 방식을 제안한다.
독립 카드의 통합 완료로 가장하지 않는다. 새 worktree가 필요하면 재현 가능한 의존 snapshot과 출처를 먼저 확보한다.

1. AS-1 High: 설치 App Server 프로토콜·managed/API 인증·tool inventory 실증 및 기존 RPC 재사용 경계. 후속 모두의 선행.
2. AS-2 High: Claude 도구 왕복·ToolSearch·타입 있는 결과와 단일 실행권. AS-1 뒤, A/B 방식은 여기서 확정.
3. AS-3 High: HTTP 간 turn 유지·pending call·crash 복구·취소·중복 방지. AS-2 뒤.
4. AS-4 High: thread resume/fork/compact/model 전환 및 prefix 정합성. AS-3 뒤.
5. AS-5 Medium: 세 launcher picker·auth 표시·context·실제 PTY 회귀와 API/구독 대조, Windows CI. AS-4 뒤.

## Gaps와 잔여 위험

이번 산출물은 소스·공식 문서 조사와 설계 제안이다. 설치 바이너리의 dynamicTools 왕복·native 도구 실행 차단,
전체 deferred schema의 최초 수신, 실제 ToolSearch·Claude Agent·compaction 표면 및 crash 복구를 직접 실행하지 않았다.
실험 API의 버전 차이와 동적 도구 고정 목록은 배포 전 차단 게이트다. canonical SPEC·기존 partial fix는 수정하지 않았다.

## 판단과 사용자 결정의 경계

현재 자료에서 Claude UI 유지와 App Server 방향이 논리적으로 모순된다는 증거는 없다. 새 허락을 먼저 요구할 사안보다
설치 버전·전체 schema·native 실행 차단·HTTP 간 RPC 유지라는 구현 전 실증 게이트가 우선이다.
A 방식이 실패할 때 B의 schema 품질 변화나 지원 도구 축소가 필요해지면, 그때 실제 비교 결과와 함께 제품 범위 결정을 보고한다.
Codex UI 대체·Claude 도구 실행권 이전·구독→API 자동 전환은 이미 승인한 목표를 바꾸므로 묵시 채택할 수 없다.


## App Server 전환 수용 기준 제안

t649 하위 설계 검토용 초안. canonical AC 추가나 완료 판정이 아니다. 각각 Given/When/Then으로 독립 판정한다.
검증은 설치 Codex 버전·schema hash·MoAI HEAD와 작업 diff·auth 모드·실행 명령·비밀값 없는 원출력을 함께 기록한다.

## AS-1 기반과 인증

**AC-AS-001** Given 설치 Codex의 schema와 기능 협상이 있을 때, When App Server를 시작하고 initialize·thread/start를
수행하면, Then 동적 도구 capability를 확인하고 unsupported 버전은 명시 오류로 끝난다. 자동 설치나 direct backend 우회는 없다.

**AC-AS-002** Given 구독 managed 계정과 API 키 전용 프로필을 각각 준비했을 때, When 로그인·상태·생성·갱신·로그아웃을
수행하면, Then 선택한 auth 모드로만 처리된다. 구독에서 MoAI의 토큰 파일 읽기·refresh·direct backend 요청은 0이고,
구독 만료/한도 오류가 API 과금 경로로 넘어가지 않는다. 같은 프로필을 경쟁 로그인으로 덮어쓰지 않는다.

**AC-AS-003** Given Codex native 도구가 가능한 기본 환경과 MoAI 제한 환경이 있을 때, When shell·file·MCP·agent 실행을
각각 유도하면, Then 제한 환경의 native 실행 계수는 0이고 실제 Claude 도구만 호출된다. 모델 도구 inventory도 대조한다.
approval never·readonly 또는 단순 설정 파일 존재로 이 AC를 통과시키지 않는다.

## AS-2 도구 실행권과 ToolSearch

**AC-AS-004** Given 실제 Claude 초기 요청의 일반·deferred 도구 N개와 각각의 schema가 있을 때, When A 방식으로
thread/start에 등록하면, Then 누락·이름 충돌·schema 변형 없이 N개가 대응된다. 최초 요청에 전체 schema가 없으면 A는
Gap/FAIL로 기록하고 B와의 비교를 보고한다. B는 native schema 수 N→1의 변경을 명시한다.

**AC-AS-005** Given App Server dynamic call이 발생했을 때, When Claude가 tool_use를 받아 승인·실행하고 다음 HTTP로
tool_result를 보내면, Then App Server의 같은 thread/turn/call에 결과가 한 번만 전달되고 최종 답변이 Claude 화면에 보인다.
텍스트·로컬 이미지·도구 실패를 각각 판정하며 unsupported 결과는 명시 오류다.

**AC-AS-006** Given ToolSearch가 `tool_reference`와 설명을 반환하는 실제 Claude 요청이 있을 때, When 검색 결과를
처리하고 발견된 도구를 호출하면, Then 400 없이 타입·이름·등록 schema가 검증되고 도구가 Claude에서 정확히 한 번 실행된다.
잘못된 tool_reference·미등록 이름·중복 발견·schema 변경·다른 대화의 참조는 송신 전 거절한다.
기존 textContent/receipt projection 거절 보고는 회귀 픽스처의 유래이며, 최신 partial fix 상태는 실행으로 재확인한다.

## AS-3 수명·권한·복구

**AC-AS-007** Given 도구 대기 때문에 첫 HTTP 응답이 끝난 turn이 있을 때, When 다음 HTTP 결과 요청이 오면,
Then App Server process와 pending RPC가 살아 있고 원 turn을 잇는다. 동시 대화 A/B의 결과를 바꾸면 둘 다 서로의 call을
소비하지 않는다. 같은 결과 재전송은 도구 재실행 없이 일관된 처리 또는 명시 중복 오류다.

**AC-AS-008** Given pending call의 실행 전·실행 후·결과 저장 후 각 지점에서 프로세스를 강제 종료했을 때,
When 재개하면, Then 실행 여부가 불명확한 도구를 자동 재실행하지 않는다. 복구 불가능한 active RPC ID는 명시 실패와
복구 안내를 제공한다. 성공 상태를 합성하거나 새 process에 옛 RPC 응답을 주입하지 않는다.

**AC-AS-009** Given 활동 중 turn 또는 병렬 tool call이 있을 때, When 사용자가 취소하거나 연결이 끊기면,
Then turn/interrupt와 대기 호출 정리가 해당 turn에만 적용되고 뒤늦은 결과가 다른 turn을 진행시키지 않는다.
EOF·RPC 오류·출력 상한 초과 때 성공 Messages terminal이 없어야 한다.

## AS-4 이력·모델·압축·fork

**AC-AS-010** Given GPT 대화를 종료하고 새 MoAI process로 시작했을 때, When 소유 thread ID로 resume하면,
Then 합성 사실·도구 결과가 유지된 정상 답변이 나온다. 다른 계정·다른 대화·변조 mapping은 거절한다.
thread/resume.history 또는 raw reasoning을 직접 가져와 재구축한 결과로 통과할 수 없다.

**AC-AS-011** Given 같은 GPT thread가 idle이고 허용 모델이 있을 때, When `/model`로 GPT-6 Astra와 5.6을 각각
선택하면, Then 실제 turn model이 선택 ID와 일치하며 App Server 경고·실패가 명시된다. 다른 제공자와 bare gpt-6는
거절된다. active tool 대기 중 전환은 잘못된 turn에 적용되지 않는다. family 호환 성공은 실제 계정별 별도 증거가 필요하다.

**AC-AS-012** Given 실제 Claude `/compact` 또는 자동 압축의 캡처된 표면이 있을 때, When 압축을 수행하면,
Then Codex contextCompaction 완료와 Claude 표시 이력이 정합하고 중복 입력·이중 압축·reasoning 유실이 없다.
단순 compact ack를 완료로 세지 않는다. 압축 표면을 식별할 수 없으면 자동 지원은 Gap으로 남긴다.

**AC-AS-013** Given 실제 Claude Agent가 병렬 자식 둘을 만들었을 때, When 각 자식 요청과 fork/resume를 처리하면,
Then 부모 및 자식 thread 소유권·결과·취소가 섞이지 않는다. agent 식별자를 얻지 못하면 텍스트나 모델 ID로 추정하여
같은 thread에 넣지 않는다. 각 자식의 tool 실행도 Claude가 담당한다.

## AS-5 제품 검증

**AC-AS-014** Given 세 MoAI launcher와 오염된 전역 모델 기본값이 있을 때, When 실제 PTY `/model`의 Default·현재 행·
목록·직접 입력·Enter·s·재개를 수행하면, Then cc는 Claude, glm은 GLM, gpt는 GPT만 선택·요청하며 공유 설정이 변하지 않는다.
구독과 API 인증 표시는 실제 선택과 일치해야 한다. Codex TUI를 열어 성공한 것은 이 AC의 증거가 아니다.

**AC-AS-015** Given 설치 App Server model metadata 및 토큰 사용량이 있을 때, When 한도 경계 안팎의 실제 입력을
보내면, Then 현재 경로 한도와 오류를 정확히 표시하고 direct endpoint 921k나 미검증 1M을 수용 보장으로 쓰지 않는다.
872k metadata도 실제 수용 시험과 별개로 표기한다. 구독 출력은 서버 정책이며 local 취소·출력 바이트 한도는 동작한다.

**AC-AS-016** Given 기존 review RPC 소비자와 Windows GitHub CI가 있을 때, When 공통 RPC 추출의 회귀 및
Windows process 종료·재개·파일 flush/원자교체·권한 시험을 실행하면, Then 기존 review 의미와 승인된 Windows API 계약이 유지된다.
Windows cross compile만으로 runtime AC를 통과시키지 않는다.

## 완료 판정

AS-1→AS-2→AS-3→AS-4→AS-5 순서의 증거를 모으고 독립 검토를 통과해야 한다. 하나의 실제 도구 왕복이나 로그인
성공으로 전체를 완료하지 않는다. mock·단위·프로토콜·제품 PTY·실제 구독/API·Windows 결과는 각각 명시한다.
계정·CI·실험 API 제약으로 수행하지 못한 항목은 Gap으로 남기며 기능을 완료로 광고하지 않는다.

## 발행 카드와 현재 상태

| 카드 | 범위 | 상태 |
|---|---|---|
| t650 | 기반·인증·격리 | picked, 구현 중 |
| t651 | 도구·ToolSearch | picked, 등록 검증 완료·제품 연결 잔여 |
| t652 | HTTP 대기·취소·복구 | picked, 구현·감사 중 |
| t653 | 재개·fork·압축·모델 | queued |
| t654 | 통합·한도·CI·배포 판정 | queued |

카드는 t649의 contains 관계로 등록했다. t650은 최신 origin/develop 9935e4e3e에서 launcher로 생성한 WT-gateway-appserver-core, `.claude/worktrees/t650`에서 구현한다. 기존 gateway 미커밋 트리는 보존했다.

## 승인 뒤 혼합 경로 실증

초기 native ToolSearch → 후발 dispatcher echo의 첫 두 호출과 최종 표식은 관측했다. 다만 성공 뒤 같은 의미의 도구 요청 세 건이 추가되어 정확히 두 번 호출하는 시험 기준은 FAIL이다. RPC ID 재생 여부는 이 기록으로 판정하지 않는다. 제품은 정상적인 새 호출과 동일 RPC 결과 재생을 구별해야 한다. 이 결과를 실제 Claude 제품 통합 성공으로 표시하지 않는다.

[혼합 경로 원시 결과](probe/hybrid-result.json)

## API 모드 출력 정책 승인

사용자 결정: API도 App Server 출력 정책 사용. 구독과 API 키 모두 공식 App Server로 연결한다. MoAI의 응답 바이트 제한과 취소는 유지하되 Claude max_tokens와 동일한 생성 토큰 상한을 보장하지 않는다. 인증과 과금 모드는 사용자 선택대로 유지하며 자동 과금 경로 전환은 없다.

## 압축과 병렬 자식의 추가 관측

[압축 실행 증거](probe/compact-hook-verdict.md)와 [병렬 자식 실행 증거](probe/agent-identity-parallel-verdict.md)를 확인했다. Claude 2.1.269와 모의 서버 시험이며 유료 호출은 없었다. 자동 압축·중첩 자식·정확한 부모 호출 매핑·제품 통합은 미검증이다.
