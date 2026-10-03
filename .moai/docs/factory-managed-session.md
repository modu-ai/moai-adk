# Factory 관리 세션 (managed session) 사용 안내

관리 세션을 켜면 `moai cc -f` / `moai glm -f` / `moai codex` 의 Factory 런치가, 레인(`lane-<n>`)과
리더(`-l/--leader`) 세션을 MoAI가 직접 소유하는 "관리 세션"으로 뜬다. 평범한 대화형
세션과 달리 런처가 세션 프로세스의 부모로 남아, 브로커(`internal/factorymsg`)에 쌓인
수신 메시지를 세션이 한가할 때 알아서 넘겨준다. **관리 세션은 기본적으로 꺼져 있고,
명시적으로 켜야만 동작한다.** 이 문서는 켜는 방법, 운영자가 눈으로 보게 되는 동작, 지켜야 할
경계를 정리한다.

일상 사용에는 읽을 필요가 없다. 관리 세션이 왜 그렇게 동작하는지 궁금하거나, 메시지가
전달되지 않을 때 원인을 좁히는 용도다.

## 켜는 방법

환경 변수 `MOAI_FACTORY_MANAGED` 를 `1` 또는 `true` 로 두고(대소문자는 구분하지 않는다), Factory
스탬프(`MOAI_KANBAN_ID` 등)가 있는 레인/리더 런치를 띄우면 관리 세션이 된다.

- 스위치가 없으면 Factory 스탬프가 있어도 모든 런치는 예전 exec 경로 그대로다.
- 스위치만 있고 Factory 스탬프가 없어도 마찬가지로 예전 경로 그대로다.
- 둘 다 있을 때만 관리 소유자로 분기한다.

## 켜도 닿지 않는 범위

- Claude / GLM 은 `moai cc -f` / `moai glm -f` 런치가 스위치가 있을 때만 관리 세션으로 들어간다.
- Codex 는 프로세스 환경에 이미 스탬프와 스위치가 함께 있는 평범한 `moai codex` 실행에서만 분기한다.
  `moai codex -f lane` 이 띄우는 카드 자식 세션은 스위치 값과 관계없이 관리되지 않고 직접 exec 경로로 간다.
- 관리되는 Codex 런치는 런처의 이후 디버그 추적 단계와 `RUST_LOG` 주입을 건너뛴다.
- 스위치를 켠 상태에서 Claude / GLM 의 `--continue` / `-c` 는 오류로 거부된다. 관리 세션이 런치 형태를
  직접 정하기 때문이고, `--continue` 는 일반 런치의 이어 하기 기능이다.
- Codex 의 `--spawn`(tmux 창) 런치는 스위치를 켜도 관리 소유자로 분기하지 않고 일반 런치로 남는다.
  Codex 레인 카드 자식과 디버그 추적에 관리 계층을 연결하는 일은 아직 카드가 없는 후속 과제다.

## 무엇이 달라지나

| 구분 | 동작 |
|---|---|
| 런치 분기 | 스위치(`MOAI_FACTORY_MANAGED`)와 Factory 스탬프(`MOAI_KANBAN_ID` 등)가 모두 있는 레인/리더 런치만 관리 소유자로 분기한다. 둘 중 하나라도 없는 런치는 예전 exec 경로 그대로다. |
| Claude / GLM | 런처가 `--print` + stream-JSON 입출력 플래그를 직접 붙이고, 연산자가 같은 플래그를 덮어쓰려 하면 거부한다. |
| Codex | 런처가 로컬 App Server를 헤드리스로 직접 띄워 루프백 WebSocket으로 붙는다. TUI는 붙지 않으므로 화면에는 아무것도 나타나지 않는다(아래 "알려진 한계"). App Server는 런치 디렉터리(프로젝트 루트, `-w`를 쓰면 그 워크트리)에서 시작한다. |
| 전달 방식 | 세션이 한가할 때만 브로커에서 한 묶음을 claim하고, 다음 턴 프롬프트에 **메타데이터만**(메시지 id, claim 토큰, 종류, 보낸 슬롯, task 참조) 넣는다. 본문은 프롬프트에 들어가지 않고, 모델이 claim 토큰으로 `factory_msg_body` 를 호출해 읽는다. |
| 직렬 처리 | 연산자 입력과 브로커 수신 메시지는 한 줄짜리 턴 큐에서 도착한 순서대로(FIFO) 처리된다. 작업 중인 세션은 끼어들기를 당하지 않으며, 연산자 입력이라고 해서 앞서 가지도 않는다. 이미 claim한 inbox 묶음이 있으면 그 뒤에 도착한 연산자 입력은 묶음 다음 차례가 된다. |
| 영수증 | 모델이 처리 뒤 `factory_msg_receipt` 로 영수증을 남기면 브로커가 "전달됨"으로 기록한다. 영수증은 **전달 증거**일 뿐 카드 완료 판정이 아니다. 판정은 수신 세션이 본문을 보고 내린다. |
| 런치 실패 | 시작에 실패한 런치는 launch-pending 등록을 되돌려서, 영구적인 대기 행이 남지 않는다. |
| 서버 요청 응답 (Codex) | 소유자는 App Server가 먼저 보내는 요청에 전부 답한다. 명령 실행·파일 변경·`applyPatchApproval`·`execCommandApproval` 승인과 MCP elicitation 은 거부하고, 권한 요청에는 빈 권한을, 동적 도구 호출에는 `success:false` 결과를 돌려준다. 사용자 입력·인증 토큰 갱신·attestation 요청은 JSON-RPC 오류(`-32000`)로, 표에 없는 method 는 `-32601` 오류로 답한다. 어느 답도 무언가를 허용하지 않는다. 답한 요청마다 stderr 에 `Factory server request answered: …` 한 줄이 남는다. 서버 요청의 id가 대기 중인 클라이언트 id와 같거나 문자열이어도 응답으로 오인하지 않고 읽기 루프도 끊기지 않으며, 연결 쓰기는 뮤텍스 하나를 거친다. |
| 턴 단위 실패 격리 | 우선 턴 이후의 턴 하나가 실패해도 세션은 이어진다. 격리 대상은 세 가지다: 스트림의 `result.is_error`, `completed` 가 아닌 상태로 끝난 Codex 턴, 거부된 MoAI 브로커 elicitation 을 겪은 Codex 턴(`serverName` 이 `moai` 인 요청이 그 턴에 귀속될 때). 드라이버는 `Factory turn failed (k/N consecutive)` 를 stderr 에 남기고 다음 턴으로 간다. 연속 N(=3)번 실패하면 마지막 오류를 반환해 세션을 끝내고, 성공한 턴은 횟수를 0으로 되돌린다. claim된 메시지는 건드리지 않는다. |

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

## 알려진 한계

- **Codex 관리 세션은 화면에 아무것도 보여 주지 않는다.** 소유자가 TUI 없는 헤드리스 App Server라서, 관리 세션으로 뜬 `moai codex`는 모델 출력이 터미널에 나오지 않는다. 소스를 읽어 확인한 내용이며 실제 실행으로 관측한 것은 아니다. TUI 부착은 후속 카드 t1408 에서 다룬다.
- **런처는 시그널을 처리하지 않는다.** 런처 프로세스가 SIGTERM이나 SIGHUP으로 끝나면 정리 코드가 돌지 않아 자식 프로세스와 임시 파일이 남을 수 있다. 이 공백은 해결되지 않았다. 시그널 처리와 `Start`/`Close` 수명주기는 카드 t1459 가 맡는다.
- **실제 codex 세션에서는 관측하지 않았다.** 서버 요청 응답과 턴 단위 실패 격리는 가짜 App Server를 상대로 한 시험으로만 확인했다. 실제 세션에서 거부 응답 뒤 모델이 어떻게 움직이는지, elicitation 요청의 `serverName` 에 `moai` 가 실리는지(귀속 판정이 기대는 값이다)는 확인하지 못했다.
- **독 메시지는 TTL까지 재배달될 수 있다.** 턴 단위 실패가 연속으로 쌓이면 세션이 끝나지만, 그 사이에 성공한 턴(운영자 입력이나 다른 메시지)이 끼면 횟수가 0으로 돌아간다. 매번 턴을 실패시키는 메시지는 lease(2분)가 끝날 때마다 다시 배달되어 TTL(최대 7일)에 닿을 때까지 이어질 수 있다. 관리 계층은 claim을 풀거나 되돌리지 않고 브로커에도 시도 횟수 상한이 없다.
- **거부된 elicitation 이 영수증을 막으면 재배달이 TTL까지 이어질 수 있다.** 소유자는 elicitation 을 거부하는데, 모델이 영수증을 쓰려던 브로커 도구 호출이 이 때문에 막히면 턴은 `completed` 로 끝나도 메시지는 확인되지 않은 채 남는다. 그 턴은 실패로 세어지지만 위와 같은 이유로 연속 상한은 사이에 성공한 턴이 없을 때만 반복을 멈춘다. 이런 elicitation 이 실제로 오는지는 위 관측 한계 때문에 알지 못한다.
- **턴 타임아웃은 세션을 끝낸다.** 격리 대상은 위 "턴 단위 실패 격리" 세 가지뿐이다. 10분 턴 타임아웃, 우선 턴 실패, 닫힌 스트림이나 연결, 쓰기 실패, 분류되지 않은 오류는 그대로 세션 전체를 끝낸다.
- **연속 실패 상한 3은 관례이지 측정값이 아니다.** 실패율 자료 없이 저장소의 "재시도 최대 3회" 관례를 가져왔다(`internal/config/defaults.go` 의 `DefaultManagedSessionMaxConsecutiveTurnFailures`).
- **쓰기 데드라인이 없다.** App Server 연결의 쓰기가 막히면 연결이 죽거나 세션이 닫힐 때까지 풀리지 않는다. 턴 타임아웃도 막힌 쓰기는 풀지 못하고, 읽기 쪽 답장 쓰기도 같은 뮤텍스에 걸려 함께 멈춘다. 소스를 읽어 확인한 내용이며 실행으로 재현하지는 않았다.
- **`id: null` 프레임에는 답하지 않는다.** `id` 가 없거나 `null` 인 프레임은 서버 요청으로 분류되지 않는다. `method` 가 있으면 알림으로 다루는데 알림은 대부분 버려지고, `method` 가 없으면 그대로 버려진다.

서버가 먼저 보내는 요청에 대한 응답(F3)과 턴 단위 실패 격리(F4)는 `SPEC-FACTORY-MANAGED-HARDEN-001`(카드 t1409)로 해결됐다. 위 한계는 그 뒤에도 남는 부분이다. 그 밖에 아래 세 가지는 후속 카드 t1410 이 맡는다.

- 시작이 일찍 실패하면 토큰 임시 디렉터리가 지워지지 않는다.
- App Server 준비 핸드셰이크의 시간 예산(10초)이 빠듯하다. 독립 감사의 실제 codex 프로브에서 패키지 하위 디렉터리를 cwd로 했을 때 한 번 시간 초과가 관측됐다.
- `/readyz` 확인이 리디렉션을 따라가면서 루프백 검사를 다시 적용하지 않는다.

## 확인 방법

- 루프백 왕복(claim, 메타데이터 주입, 본문 조회, 영수증)은 두 번째 호스트 없이
  `go test ./internal/cli -run '^TestManagedSessionLoopbackRoundTrip$'` 로 검증된다.
- 실제 Codex 세션 왕복은 `MOAI_FACTORY_LIVE_ROOT` / `MOAI_FACTORY_LIVE_RUN` 을 둔 환경에서만
  `TestManagedCodexFactoryBrokerLive` 가 실행하고, 두 값이 없으면 skip 으로 통과한다.
- 서버 요청 응답은 `go test -race ./internal/cli -run '^TestManagedCodexServerRequestPolicy$'` 로, 턴 단위 실패 격리와 연속 실패 상한은 `go test -race ./internal/cli -run '^TestManagedDriverIsolatesTurnFailure$'` 와 `go test ./internal/cli -run '^TestManagedDriverConsecutiveFailureCeiling$'` 로 가짜 서버를 상대로 검증된다. 이 시험은 실제 codex 세션을 쓰지 않는다.
- 메시지가 오지 않을 때: `factory_msg_status` 도구로 브로커 상태(대기/확인 수)를 먼저
  보고, 레인이 launch-pending 에 머물러 있는지 확인한다.

SPEC: `.moai/specs/SPEC-FACTORY-MANAGED-SESSION-001/` (card t1375)
후속 SPEC: `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/` (card t1409, 서버 요청 응답과 턴 단위 실패 격리)
