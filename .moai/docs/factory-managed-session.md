# Factory 관리 세션 (managed session) 사용 안내

`moai cc -f` / `moai glm -f` / `moai codex` 의 Factory 런치는, 레인(`lane-<n>`)과
리더(`-l/--leader`) 세션을 MoAI가 직접 소유하는 "관리 세션"으로 띄운다. 평범한 대화형
세션과 달리 런처가 세션 프로세스의 부모로 남아, 브로커(`internal/factorymsg`)에 쌓인
수신 메시지를 세션이 한가할 때 알아서 넘겨준다. 이 문서는 운영자가 눈으로 보게 되는
동작과 지켜야 할 경계를 정리한다.

일상 사용에는 읽을 필요가 없다. 관리 세션이 왜 그렇게 동작하는지 궁금하거나, 메시지가
전달되지 않을 때 원인을 좁히는 용도다.

## 무엇이 달라지나

| 구분 | 동작 |
|---|---|
| 런치 분기 | Factory 환경(`MOAI_KANBAN_ID` 등)이 있는 레인/리더 런치만 관리 소유자로 분기한다. 환경이 없는 런치는 예전 exec 경로 그대로다. |
| Claude / GLM | 런처가 `--print` + stream-JSON 입출력 플래그를 직접 붙이고, 연산자가 같은 플래그를 덮어쓰려 하면 거부한다. |
| Codex | 런처가 로컬 App Server를 직접 띄워 루프백 WebSocket으로 붙는다. |
| 전달 방식 | 세션이 한가할 때만 브로커에서 한 묶음을 claim하고, 다음 턴 프롬프트에 **메타데이터만**(메시지 id, claim 토큰, 종류, 보낸 슬롯, task 참조) 넣는다. 본문은 프롬프트에 들어가지 않고, 모델이 claim 토큰으로 `factory_msg_body` 를 호출해 읽는다. |
| 직렬 처리 | 연산자 입력과 브로커 수신 메시지는 한 줄짜리 턴 큐로 순서대로 처리된다. 작업 중인 세션은 끼어들기를 당하지 않는다. |
| 영수증 | 모델이 처리 뒤 `factory_msg_receipt` 로 영수증을 남기면 브로커가 "전달됨"으로 기록한다. 영수증은 **전달 증거**일 뿐 카드 완료 판정이 아니다. 판정은 수신 세션이 본문을 보고 내린다. |
| 런치 실패 | 시작에 실패한 런치는 launch-pending 등록을 되돌려서, 영구적인 대기 행이 남지 않는다. |

## 운영자가 알아야 할 규칙

- **Claude 전용 run 에서는 `factory_msg_send` 가 거부된다.** Claude 세션끼리는 기존
  SendMessage 정책을 그대로 쓴다. 브로커 송신이 필요하면 Codex가 낀 혼합 run 이어야 한다.
- **전송은 루프백 전용이다.** Codex App Server 주소는 `ws://` + 루프백 IP 리터럴
  (`127.0.0.0/8` 또는 `::1`)만 허용하고, 호스트 이름 해석은 하지 않는다. 가드가 네트워크
  접근보다 먼저 걸리므로 루프백이 아닌 대상은 프로세스 밖으로 나가지 않는다. 연결 토큰은
  런치마다 무작위로 만든다.
- **MCP 승인 인수는 소유한 프로세스에만 붙는다.** MoAI MCP 도구 호출을 승인 창 없이
  받기 위한 `-c mcp_servers.moai...approval_mode="approve"` 인수는 런처가 띄운 App Server
  명령행에만 들어간다. 프로젝트 설정의 `default_tools_approval_mode = "writes"` 는 건드리지
  않는다. 옛 Factory 세대가 프로젝트 전역에 승인 오버라이드를 남겨 두었다면 `moai doctor`
  가 경고만 하고 사용자 소유 설정 파일은 고쳐 쓰지 않는다.
- **어휘는 정규형만 쓴다.** 레인은 `-f lane` / `-f lane-<n>`, 리더 탐색은 `-l/--leader`.
  `agent-<n>` / `worker-<n>` 이름은 되살리지 않는다.
- **범위는 전달까지다.** 카드 완료 판정, 병합 자동화, 컨트롤러 재알림은 관리 세션의 일이
  아니다(SPEC-FACTORY-CONTROLLER-001 소관).
- **Windows 도 자식 프로세스 소유 모델로 동작한다.** 관리 계층은 `syscall` 을 쓰지 않는다.

## 확인 방법

- 루프백 왕복(claim, 메타데이터 주입, 본문 조회, 영수증)은 두 번째 호스트 없이
  `go test ./internal/cli -run '^TestManagedSessionLoopbackRoundTrip$'` 로 검증된다.
- 실제 Codex 세션 왕복은 `MOAI_FACTORY_LIVE_ROOT` / `MOAI_FACTORY_LIVE_RUN` 을 둔 환경에서만
  `TestManagedCodexFactoryBrokerLive` 가 실행하고, 두 값이 없으면 skip 으로 통과한다.
- 메시지가 오지 않을 때: `factory_msg_status` 도구로 브로커 상태(대기/확인 수)를 먼저
  보고, 레인이 launch-pending 에 머물러 있는지 확인한다.

SPEC: `.moai/specs/SPEC-FACTORY-MANAGED-SESSION-001/` (card t1375)
