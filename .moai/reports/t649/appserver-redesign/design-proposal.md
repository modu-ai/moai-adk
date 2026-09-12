# Codex App Server 전환 설계 제안

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
