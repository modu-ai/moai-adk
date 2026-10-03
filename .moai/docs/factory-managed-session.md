# Factory 관리 세션 (managed session) 사용 안내

관리 세션을 켜면 리더(`moai cc -f` / `moai glm -f`)와 레인(`moai cc -l` / `moai glm -l` / `moai codex -l`,
`lane-<n>`)의 Factory 런치가, 그 세션을 MoAI가 직접 소유하는 "관리 세션"으로 뜬다. 평범한 대화형
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

- Claude / GLM 은 `moai cc -f|-l` / `moai glm -f|-l` 런치가 스위치가 있을 때만 관리 세션으로 들어간다.
- Codex 는 두 문으로 관리 소유자에 닿는다. 하나는 프로세스 환경에 이미 스탬프와 스위치가 함께 있는
  평범한 `moai codex` 실행이다. 다른 하나는 `moai codex -l` 레인 루프로, 스위치와 레인 스탬프가
  함께 있으면 카드마다 띄우는 자식 세션을 관리 소유자로 보낸다(SPEC-FACTORY-MANAGED-CARD-CHILD-001,
  카드 t1440). 스위치가 꺼져 있으면 레인 루프의 카드 자식은 예전과 같은 직접 exec 경로로 간다.
- 레인 루프의 관리 카드 자식은 카드 워크트리를 소유자의 런치 디렉터리로 받고(`-C` 인수는 넘기지 않는다),
  레인 claim 과 브로커 endpoint 소유자는 런처 프로세스에 그대로 남는다. 앵커 락은 걸지 않는다.
- 관리 카드 자식의 환경은 직접 exec 카드 자식과 두 곳이 다르다. 첫째, 관리 소유자가 요구하는 run id 를
  `MOAI_KANBAN_ID` 로 싣는다(직접 exec 카드 자식은 이 키를 싣지 않는다). 둘째, 소유자가 App Server
  환경에 `MOAI_SESSION_PID`(런처 PID)를 더한다(직접 exec 카드 자식 환경에서는 이 키를 지운다).
- 관리되는 Codex 런치는 런처의 이후 디버그 추적 단계와 `RUST_LOG` 주입을 건너뛴다.
- 스위치를 켠 상태에서 Claude / GLM 의 `--continue` / `-c` 는 오류로 거부된다. 관리 세션이 런치 형태를
  직접 정하기 때문이고, `--continue` 는 일반 런치의 이어 하기 기능이다.
- Codex 의 `--spawn`(tmux 창) 런치는 스위치를 켜도 관리 소유자로 분기하지 않고 일반 런치로 남는다.
  디버그 추적에 관리 계층을 연결하는 일은 아직 카드가 없는 후속 과제다.

## 무엇이 달라지나

| 구분 | 동작 |
|---|---|
| 런치 분기 | 스위치(`MOAI_FACTORY_MANAGED`)와 Factory 스탬프(`MOAI_KANBAN_ID` 등)가 모두 있는 레인/리더 런치만 관리 소유자로 분기한다. 둘 중 하나라도 없는 런치는 예전 exec 경로 그대로다. |
| Claude / GLM | 런처가 `--print` + stream-JSON 입출력 플래그를 직접 붙이고, 연산자가 같은 플래그를 덮어쓰려 하면 거부한다. |
| Codex | 런처가 로컬 App Server를 직접 띄워 루프백 WebSocket으로 붙는다. 터미널과 `codex` 기능 조건이 맞으면 같은 App Server·같은 스레드에 Codex TUI 를 두 번째 클라이언트로 붙이고, 조건이 맞지 않으면 예전처럼 헤드리스로 남는다(아래 "알려진 한계"의 TUI 항목들). App Server는 런치 디렉터리(프로젝트 루트, `-w`를 쓰면 그 워크트리)에서 시작한다. |
| 전달 방식 | 세션이 한가할 때만 브로커에서 한 묶음을 claim하고, 다음 턴 프롬프트에 **메타데이터만**(메시지 id, claim 토큰, 종류, 보낸 슬롯, task 참조) 넣는다. 본문은 프롬프트에 들어가지 않고, 모델이 claim 토큰으로 `factory_msg_body` 를 호출해 읽는다. |
| 직렬 처리 | 연산자 입력과 브로커 수신 메시지는 한 줄짜리 턴 큐에서 도착한 순서대로(FIFO) 처리된다. 작업 중인 세션은 끼어들기를 당하지 않으며, 연산자 입력이라고 해서 앞서 가지도 않는다. 이미 claim한 inbox 묶음이 있으면 그 뒤에 도착한 연산자 입력은 묶음 다음 차례가 된다. Codex TUI 가 붙은 동안에는 연산자 입력이 TUI 로 가므로 이 큐를 거치지 않는다(아래 TUI 항목). |
| 영수증 | 모델이 처리 뒤 `factory_msg_receipt` 로 영수증을 남기면 브로커가 "전달됨"으로 기록한다. 영수증은 **전달 증거**일 뿐 카드 완료 판정이 아니다. 판정은 수신 세션이 본문을 보고 내린다. |
| 런치 실패 | 시작에 실패한 런치는 launch-pending 등록을 되돌려서, 영구적인 대기 행이 남지 않는다. |
| 서버 요청 응답 (Codex) | 소유자는 App Server가 먼저 보내는 요청에 전부 답한다. 명령 실행·파일 변경 승인과 MCP elicitation 은 `decline` 으로 거부한다. 레거시 승인 두 종(`applyPatchApproval`, `execCommandApproval`)은 스키마가 정한 객체 `{"decision":{"denied":{"rejection":"<고정 영어 문구>"}}}` 로 거부한다. 문자열 `"denied"` 는 codex 0.160.0 스키마에 없는 값이라 쓰지 않고, 문자열 `abort` 는 턴을 끊으므로 쓰지 않는다. 권한 요청에는 빈 권한을, 동적 도구 호출에는 `success:false` 결과를 돌려준다. 사용자 입력·인증 토큰 갱신·attestation 요청은 JSON-RPC 오류(`-32000`)로, 표에 없는 method 는 `-32601` 오류로 답한다. 어느 답도 무언가를 허용하지 않는다. 결과를 돌려주는 답 일곱 종의 결과 본문은 vendoring 한 codex 0.160.0 응답 스키마(`internal/cli/testdata/codex-0.160.0/`)에 대해 시험으로 검증한다. 오류로 답하는 세 종과 미지 method 폴백은 시험이 오류 객체를 다시 조립해 `JSONRPCError.json` 에 대조하므로, 와이어에서 필수 필드가 빠져도 이 가드는 잡지 못한다(오류 코드와 메시지는 정책 시험이 정확한 값으로 따로 단언한다). 답한 요청마다 stderr 에 `Factory server request answered: …` 한 줄이 남는다. 서버 요청의 id가 대기 중인 클라이언트 id와 같거나 문자열이어도 응답으로 오인하지 않고 읽기 루프도 끊기지 않으며, 연결 쓰기는 뮤텍스 하나를 거친다. |
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
- **어휘는 정규형만 쓴다.** 레인 합류는 인자 없는 `-l`/`--lane`, 리더 지정은 `-l`/`--lane`과 함께 쓰는 긴 형태 `--leader <이름>`.
  `agent-<n>` / `worker-<n>` 이름은 되살리지 않는다.
- **범위는 전달까지다.** 카드 완료 판정, 병합 자동화, 컨트롤러 재알림은 관리 세션의 일이
  아니다(SPEC-FACTORY-CONTROLLER-001 소관).
- **Windows 도 자식 프로세스 소유 모델로 동작한다.** 관리 계층은 `syscall` 을 쓰지 않는다.

## 알려진 한계

- <!-- anchor:tui-debt-status --> **Codex TUI 부착(`SPEC-FACTORY-MANAGED-TUI-001`, 카드 t1408): 알려진 부채 1은 `addressed`, 실제 TUI 동작은 `unverified`.** 관리 세션이 화면에 아무것도 보여 주지 않던 문제를 다루는 변경이 들어왔다. 통과한 것은 가짜 codex 를 상대로 한 시험뿐이고 실제 TUI 를 돌려 본 관측은 없다. 운영자가 수동 점검 AC-MT-015 를 실행해 1·2·3·6단계를 관측으로 기록하기 전에는 이 부채를 "해결됨"이라고 쓰지 않는다.
- <!-- anchor:tui-opt-out --> **TUI 부착은 `MOAI_FACTORY_MANAGED_TUI` 로 끈다.** 관리 게이트(`MOAI_FACTORY_MANAGED` + Factory 스탬프) 안에서는 기본으로 TUI 를 붙인다. `MOAI_FACTORY_MANAGED_TUI` 를 `0`, `false`, `off`(대소문자와 앞뒤 공백은 무시) 중 하나로 두면 예전처럼 헤드리스로 남고 안내 한 줄만 찍힌다. 두 번째 옵트인이 아니라 게이트 안에서 새 동작만 끄는 장치다.
- <!-- anchor:tui-log-file --> **TUI 가 떠 있는 동안 런처 출력은 로그 파일로 간다.** 터미널은 TUI 가 쓰므로 런처의 운영자용 로그 줄과 App Server 의 stderr 는 프로젝트의 `.moai/logs/factory-managed-<run-id>-<레인 라벨>.log` 에 쌓이고, 경로는 TUI 가 뜨기 전에 stderr 에 한 번 찍힌다. `Factory server request answered: …`, `Factory turn failed …` 같은 HARDEN-001 의 흔적도 이 동안에는 화면이 아니라 이 파일에 있다. TUI 가 끝나면 이후 런처 로그 줄은 다시 터미널로 나오고 App Server 의 stderr 는 끝까지 파일에 남는다.
- <!-- anchor:tui-quit --> **TUI 가 붙은 동안 `/exit`·`/quit` 는 런처 명령이 아니다.** 터미널 입력은 TUI 가 받는다. 런처는 stdin 을 읽지 않고 `/exit`·`/quit` 도 해석하지 않으므로 세션은 TUI 의 `/quit` 같은 자체 명령으로 끝낸다. TUI 종료 상태가 런처 종료 코드가 되고(0 이면 0, 시그널로 끝나면 1), 연산자 입력은 런처의 턴 큐를 거치지 않는다. 헤드리스로 남은 경우(옵트아웃·조건 불충족)는 기존 그대로다.
- <!-- anchor:tui-probe --> **부착 조건은 버전이 아니라 기능 탐지로 정한다.** stdin 과 stdout 이 모두 터미널이고 `codex resume --help` 출력에 `--remote` 와 `--remote-auth-token-env` 가 둘 다 있을 때만 붙인다. 하나라도 어긋나거나 TUI 를 띄우지 못하면 헤드리스로 이어 가며 이유를 밝힌 안내 한 줄을 stderr 에 한 번 찍는다. 이 경우 종료 상태는 달라지지 않는다.
- <!-- anchor:tui-signals --> **시그널 공백은 그대로 열려 있다.** TUI 부착은 이 공백을 닫지 않는다. 런처가 속한 프로세스 그룹에 SIGHUP·SIGINT 가 가면 정리 코드 없이 끝나 App Server 가 남을 수 있다. 시그널 처리와 `Start`/`Close` 재설계는 계속 카드 t1459 가 맡는다.
- <!-- anchor:tui-unobserved --> **미관측 1 — 런처 연결이 TUI 가 시작한 턴의 수명주기 프레임을 받는지.** 받지 못하면 런처는 스레드가 바쁜 것을 알 수 없어 브로커 묶음을 미루지 못하고 진행 중인 턴에 합류시킬 수 있다(교착은 아니다). 가짜 서버로만 확인했다.
- <!-- anchor:tui-unobserved --> **미관측 2 — 두 연결이 한 스레드를 구독할 때 서버 요청이 어디로 가는지.** turnId 로 범위를 가르는 규칙은 어느 경로에서도 성립하게 만들었지만, turnId 가 없는 요청 네 종(과 turnId 가 null 인 elicitation)은 예전의 자동 응답을 유지하므로 서버가 두 연결에 모두 보내면 운영자의 답과 경합할 수 있다.
- <!-- anchor:tui-unobserved --> **미관측 3 — 실제 TUI 의 동작.** `codex resume --remote` 가 스레드 id 로 실행 중인 스레드를 찾는지, 런처가 시작한 턴을 화면에 그리는지, `/quit` 가 0 으로 끝나는지, 인터럽트 뒤 kill 로 끝낸 TUI 가 터미널을 복구하는지, 터미널 감지가 실제 tty 에서 맞는지는 시험이 보지 못한다.
- <!-- anchor:tui-manual-check --> **수동 점검 AC-MT-015 는 운영자 몫이고 CI 에서 돌지 않는다.** 터미널이 지정되지 않아 현재 기록은 "operator confirmation pending, no terminal designated" 이다. 절차는 `.moai/specs/SPEC-FACTORY-MANAGED-TUI-001/acceptance.md` §4, 결과는 `.moai/reports/t1408/manual-attach-check.md` 에 남긴다.
- <!-- anchor:tui-limits --> **TUI 부착의 나머지 한계.** 활성 턴이 길면 브로커 묶음은 그 턴이 끝날 때까지 미뤄지고 시간으로 풀리지 않는다(`turn/completed` 프레임을 잃으면 중계가 굶고, 단서는 로그 파일의 반복 줄뿐이다). 한가함 확인과 `turn/start` 사이의 경주 구간에 운영자가 시작한 턴에는 합류한다. 런처 자신의 `turn/start` 가 끝나기 전에 도착해 turn id 를 모르는 요청은 런처의 턴으로 보고 거부한다. 토큰은 TUI 자식 프로세스의 환경에 있어 같은 사용자의 프로세스 조회로 보일 수 있다. Windows 실행과 codex 버전 범위는 인증하지 않았다.
- <!-- anchor:tui-armed-window --> **알려진 한계(수리하지 않음) — 운영자 턴이 소유자의 `turn/start` 응답보다 먼저 시작해 끝나는 경우.** 소유자가 자기 턴을 시작하려고 창을 연 사이에 운영자 턴이 시작·종료하면 그 턴이 소유자의 것으로 귀속된다. 이후 소유자 자신의 수명주기 프레임은 걸러지고, 소유자는 턴 제한 시간이 찰 때까지 자기 턴을 기다린다. 고치려면 응답 id 로 귀속하고 프레임을 붙들어 두어야 해서 REQ-MT-008 의 예외를 건드린다. 현재 동작은 `TestManagedOperatorTurnInsideArmedWindowKnownDebt` 가 고정한다(독립 sync 감사 F4, SPEC 부채 14).
- <!-- anchor:tui-log-race --> **알려진 한계 — 세션 로그 열기의 열린 우회.** `openManagedTUILog` 는 로그 경로의 말단을 심볼릭 링크로 바꿔치기하는 것과 파일 모드가 너무 열린 것을 막는다(`TestManagedTUILogFileIsSafe`). 그러나 로그 경로에 놓인 하드 링크(같은 파일 검사를 통과해서 추가 쓰기와 모드 조임이 원본 파일에 닿는다)와 심볼릭 링크로 된 상위 디렉터리(`.moai/logs` 가 프로젝트 밖을 가리키는 경우)는 열린 알려진 한계이고, 마지막 검사와 파일 사용 사이의 경합도 남는다. 모두 같은 사용자의 로그 위치 쓰기 권한이 있어야 악용된다(SPEC 부채 15 — SPEC 본문은 좁은 경합만 적고 있어 이 문서가 더 정확하다). 연결이 TUI 부착 전에 끊기는 경우(F5)는 고쳤다(`TestManagedCodexConnectionLostBeforeAttachStopsTUI`).
- **`moai codex -l` 의 관리 카드 자식은 TUI 를 붙이지 않고 헤드리스로 남는다.** 이전 문서의 "Codex 관리 세션은 화면에 아무것도 보여 주지 않는다"라는 문장은 일반 `moai codex` 관리 세션에는 더 이상 맞지 않고, 이 카드 자식에만 그대로 맞는다. 레인 루프가 소유자에게 터미널 파일이 아니라 입력 어댑터를 건네므로, 부착 조건(stdin 이 터미널 파일일 것)이 맞지 않는다. 그래서 위 TUI 부착은 카드 자식이 아니라 일반 `moai codex` 관리 세션에만 닿고, 카드 자식의 모델 출력은 여전히 레인 터미널에 나오지 않는다. 코드(`internal/cli/managed_codex_tui.go` 의 `operatorTUIPreconditions`, `internal/cli/codex_launcher.go` 의 두 시임)를 읽어 확인한 내용이며 실제 레인 터미널로 관측하지는 않았다.
- **런처는 시그널을 처리하지 않는다.** 런처 프로세스가 SIGTERM이나 SIGHUP으로 끝나면 정리 코드가 돌지 않아 자식 프로세스와 임시 파일이 남을 수 있다. 이 공백은 해결되지 않았다. 시그널 처리와 `Start`/`Close` 수명주기는 카드 t1459 가 맡는다.
- **실제 codex 세션에서는 관측하지 않았다.** 서버 요청 응답과 턴 단위 실패 격리는 가짜 App Server를 상대로 한 시험으로만 확인했다. 실제 세션에서 거부 응답 뒤 모델이 어떻게 움직이는지, elicitation 요청의 `serverName` 에 `moai` 가 실리는지(귀속 판정이 기대는 값이다)는 확인하지 못했다. 레거시 승인 요청에 돌려주는 거부 객체 `{"decision":{"denied":{"rejection":…}}}` 는 vendoring 한 스키마에 대한 유효성만 확인했고, 어떤 실제 codex 세션도 그 객체를 받아 본 적이 없다. 서버가 그 객체를 받은 뒤 실제로 어떻게 처리하는지는 알지 못한다.
- **독 메시지는 TTL까지 재배달될 수 있다.** 턴 단위 실패가 연속으로 쌓이면 세션이 끝나지만, 그 사이에 성공한 턴(운영자 입력이나 다른 메시지)이 끼면 횟수가 0으로 돌아간다. 매번 턴을 실패시키는 메시지는 lease(2분)가 끝날 때마다 다시 배달되어 TTL(최대 7일)에 닿을 때까지 이어질 수 있다. 관리 계층은 claim을 풀거나 되돌리지 않고 브로커에도 시도 횟수 상한이 없다.
- **거부된 elicitation 이 영수증을 막으면 재배달이 TTL까지 이어질 수 있다.** 소유자는 elicitation 을 거부하는데, 모델이 영수증을 쓰려던 브로커 도구 호출이 이 때문에 막히면 턴은 `completed` 로 끝나도 메시지는 확인되지 않은 채 남는다. 그 턴은 실패로 세어지지만 위와 같은 이유로 연속 상한은 사이에 성공한 턴이 없을 때만 반복을 멈춘다. 이런 elicitation 이 실제로 오는지는 위 관측 한계 때문에 알지 못한다.
- **턴 타임아웃은 세션을 끝낸다.** 격리 대상은 위 "턴 단위 실패 격리" 세 가지뿐이다. 10분 턴 타임아웃, 우선 턴 실패, 닫힌 스트림이나 연결, 쓰기 실패, 분류되지 않은 오류는 그대로 세션 전체를 끝낸다.
- **연속 실패 상한 3은 관례이지 측정값이 아니다.** 실패율 자료 없이 저장소의 "재시도 최대 3회" 관례를 가져왔다(`internal/config/defaults.go` 의 `DefaultManagedSessionMaxConsecutiveTurnFailures`).
- **쓰기 데드라인이 없다.** App Server 연결의 쓰기가 막히면 연결이 죽거나 세션이 닫힐 때까지 풀리지 않는다. 턴 타임아웃도 막힌 쓰기는 풀지 못하고, 읽기 쪽 답장 쓰기도 같은 뮤텍스에 걸려 함께 멈춘다. 소스를 읽어 확인한 내용이며 실행으로 재현하지는 않았다.
- **`id: null` 프레임에는 답하지 않는다.** `id` 가 없거나 `null` 인 프레임은 서버 요청으로 분류되지 않는다. `method` 가 있으면 알림으로 다루는데 알림은 대부분 버려지고, `method` 가 없으면 그대로 버려진다.
- **스키마 가드는 오류 프레임의 필수 필드 누락을 잡지 못한다.** 오류로 답하는 세 종과 미지 method 폴백은 가드가 와이어 원문이 아니라 디코딩한 구조체로 `{"id","error":{"code","message"}}` 를 다시 조립해 `JSONRPCError.json` 에 대조하므로, 와이어에서 `message` 가 빠져도 가드는 통과한다. 독립 델타 감사가 JSON 태그를 바꿔 재현했다(감사 부채 D1). 현재 프로덕션 오류 프레임은 `message` 를 항상 쓰고 오류 코드와 메시지는 정책 시험이 정확히 단언하지만, 이 방향의 미래 회귀를 가드가 지켜 주지는 않는다.
- **서버가 보낸 `turnId` 는 로그 줄에 따옴표 없이 실린다.** elicitation 응답 로그 줄의 `turn=<id>` 는 서버가 보낸 값을 그대로 찍는다(`serverName` 과 method 는 따옴표로 감싼다). 개행이 든 `turnId` 가 로그 줄 하나를 위조할 수 있다는 소스 판독이며 실행으로 재현하지는 않았다. 서버는 토큰으로 인증된 루프백 자식이라 위험은 낮다. SPEC 이 로그 문법을 `turn=<id>` 로 고정해 두어 이 카드에서는 고치지 않았다. `Factory turn failed (…): <오류>` 줄도 오류 문구를 따옴표 없이 그대로 싣는다.
- **무인 레인은 첫 카드 이후로 진행하지 않는다.** 관리 카드 자식은 우선 턴 뒤 유휴 상태이고, 레인 루프가 카드 시작 프롬프트를 주입하지 않기 때문이다. 운영자가 입력하거나 브로커 메시지가 오기 전에는 세션이 스스로 카드 작업을 시작하지 않는다. 그래서 `moai codex -l` 의 관리 경로는 운영자가 붙어 있는 레인에서만 이득이 닿는다. 카드 시작 프롬프트를 주입하는 일은 아직 카드가 없는 후속 과제다. 가짜 App Server 시험으로 확인한 범위이며 무인 레인을 실제로 돌려 보지는 않았다.
- **관리 카드 자식의 운영자 입력은 줄 단위로 나뉜다.** 레인 루프는 stdin 을 한 줄씩 세션에 건네고, 세션을 끝낸 `/exit` 또는 `/quit` 줄 이후의 입력은 끝난 세션에 주지 않고 다음 카드 세션이 받는다. 보장은 그 두 토큰으로 끝난 세션에 한정된다. 치명 오류처럼 다른 이유로 끝난 세션이 이미 받은 줄은 되찾지 않고, `/exit` 를 연달아 두 번 치면 두 번째 줄이 다음 카드 세션을 곧바로 끝낸다. 입력 어댑터는 드라이버의 종료 토큰 두 값을 복제하므로 드라이버에 토큰이 더해지면 어댑터가 따라가지 못한다. 이 어긋남은 루프 수준 시험도 `/exit`·`/quit` 만 돌려 잡지 못한다. 한 줄은 개행을 포함해 `bufio.MaxScanTokenSize`(65,536바이트)까지만 버퍼에 담는다. 이 한도를 넘는 줄은 세션에 건네지 않고 입력이 거기서 끝나며, 드라이버의 스캐너가 너무 긴 줄에서 입력을 끝내는 동작을 그대로 따른다(`TestManagedOperatorInputPumpBoundsLineLength`). 어댑터는 소유자 세션이 스스로 닫힌 뒤에야 닫히므로, 치명 오류로 끝나는 세션의 닫힘 구간에 친 줄은 이미 죽은 세션으로 갈 수 있다. 위 "이미 받은 줄" 범위가 문서보다 넓을 수 있다는 뜻이다. 코드를 읽어 확인한 내용이며 실행으로 관측하지 않았다.
- **치명 오류로 끝난 세션은 드라이버의 입력 읽기 고루틴을 프로세스가 끝날 때까지 남길 수 있다.** 드라이버의 8칸 입력 채널이 가득 찬 상태에서(운영자 줄이 9개 이상 쌓여 있을 때) 세션이 치명 오류로 끝나면, 소유자의 입력 고루틴(`internal/cli/managed_factory_session.go` 의 `readManagedOperatorInput`, 약 294행)이 채널 송신에서 풀리지 않는다. `Close()` 도 `stop()` 도 이 고루틴을 풀지 못한다. 독립 감사가 재현했다(`Close()` 와 `stop()` 뒤에도 해당 줄에서 `chan send` 로 대기). 비용은 그런 세션 하나당 작은 고루틴 하나와 붙들린 줄 최대 8개다. 온전한 수리는 REQ-CC-008 이 고정한 소유자·드라이버 파일을 바꿔야 해서 이 카드에서는 부채로 받아들였다.
- **stdin 이 EOF 인 레인은 `/exit`·`/quit`·치명 오류 없이는 세션이 끝나지 않는다.** 드라이버의 끝 조건을 이 카드에서 바꾸지 않았다.
- **연속 시작 실패에 상한이 없다.** 관리 시작이 계속 실패하면 레인 루프가 큐의 카드를 연달아 lease 하며 소진할 수 있다. 카드는 lease 가 끝날 때까지 묶인다. 예전 직접 exec 루프도 같은 상한 없음이다.
- **개발자 지침 쌍이 스레드에 적용되는지는 관측하지 못했다.** 카드 워크트리의 지역 지침을 `-c developer_instructions=…` 인수로 소유자에 넘기지만, 가짜 App Server 로는 그 값이 실제 codex 스레드에 닿는지 볼 수 없다.
- **옛 직접 exec 문이 POSIX 에서 런처 프로세스를 교체해 스위치를 끈 레인 루프가 첫 카드 뒤에 이어지지 않던 결함은 카드 t1488 이 고쳤다. 이 카드는 그 문을 바꾸지 않았다.** 이 카드의 측정(P-1)이 먼저 그 결함을 보였다. 기본 직접 문 `defaultCodexDirectLaunch`(`internal/cli/codex_direct_posix.go`)가 `syscall.Exec` 로 프로세스를 바꿔치기했고, darwin/arm64 에서 가짜 codex 를 상대로 잰 결과 호출은 한 번뿐이었으며(`card=t1`) 가짜 codex 의 pid 가 런처 시험 프로세스의 pid 와 같았고 루프는 돌아오지 않았다. 교체하지 않는 대조 문에서는 두 카드가 순서대로 실행됐다. t1488 병합 뒤의 코드는 카드 환경(`MOAI_KANBAN_CARD`)이 있으면 `codexStartAndWait`(`internal/cli/codex_direct_wait.go`)로 자식을 띄우고 기다리므로, 레인 카드 자식은 직접 문에서도 런처가 부모로 남는다. 카드 환경이 없는 직접 실행은 여전히 `syscall.Exec` 를 쓴다. 이 문단은 병합 트리의 코드를 읽어 확인한 내용이고, 병합 트리에서 다시 돌린 시험은 `.moai/specs/SPEC-FACTORY-MANAGED-CARD-CHILD-001/progress.md` §E.2 에 있다. Windows 직접 문은 재지 않았다. 최초 측정 기록은 같은 파일 §E.2 의 M0 P-1 항목에 있다.

서버가 먼저 보내는 요청에 대한 응답(F3)과 턴 단위 실패 격리(F4)는 `SPEC-FACTORY-MANAGED-HARDEN-001`(카드 t1409)로 해결됐다. 독립 sync 감사가 레거시 승인 응답 값이 스키마 위반이라는 점(F1)을 찾아냈고, 수리와 SPEC 개정으로 바로잡았다. 위 한계는 그 뒤에도 남는 부분이다. `moai codex -l` 카드 자식을 관리 경로에 연결하는 일은 `SPEC-FACTORY-MANAGED-CARD-CHILD-001`(카드 t1440)로 해결됐고, 위 "켜도 닿지 않는 범위"와 한계 목록이 그 뒤의 상태다. 그 밖에 아래 세 가지는 후속 카드 t1410 이 맡는다.

- 시작이 일찍 실패하면 토큰 임시 디렉터리가 지워지지 않는다.
- App Server 준비 핸드셰이크의 시간 예산(10초)이 빠듯하다. 독립 감사의 실제 codex 프로브에서 패키지 하위 디렉터리를 cwd로 했을 때 한 번 시간 초과가 관측됐다.
- `/readyz` 확인이 리디렉션을 따라가면서 루프백 검사를 다시 적용하지 않는다.

## 확인 방법

- 루프백 왕복(claim, 메타데이터 주입, 본문 조회, 영수증)은 두 번째 호스트 없이
  `go test ./internal/cli -run '^TestManagedSessionLoopbackRoundTrip$'` 로 검증된다.
- 실제 Codex 세션 왕복은 `MOAI_FACTORY_LIVE_ROOT` / `MOAI_FACTORY_LIVE_RUN` 을 둔 환경에서만
  `TestManagedCodexFactoryBrokerLive` 가 실행하고, 두 값이 없으면 skip 으로 통과한다.
- 서버 요청 응답은 `go test -race ./internal/cli -run '^TestManagedCodexServerRequestPolicy$'` 로, 턴 단위 실패 격리와 연속 실패 상한은 `go test -race ./internal/cli -run '^TestManagedDriverIsolatesTurnFailure$'` 와 `go test ./internal/cli -run '^TestManagedDriverConsecutiveFailureCeiling$'` 로 가짜 서버를 상대로 검증된다. 이 시험은 실제 codex 세션을 쓰지 않는다.
- 서버 요청에 돌려주는 결과 본문(일곱 종)의 모양이 codex 0.160.0 스키마를 지키는지는 `go test ./internal/cli -run '^TestManagedServerRequestPolicyMatchesCodexSchema$'` 로 확인한다. `codex` 바이너리도 네트워크도 필요 없고, `internal/cli/testdata/codex-0.160.0/` 에 vendoring 한 스키마 사본을 쓴다. 최소 지원 codex 버전이 올라가면 사본을 다시 만들고(생성 명령은 그 폴더의 `README.md`) 시험을 다시 돌려야 한다. 이 시험은 모양만 보며, 스키마상 유효한 오답(`abort`, accept 계열)은 `TestManagedCodexServerRequestPolicy` 의 정확한 응답 단언이 잡는다.
- TUI 부착은 `go test -race ./internal/cli -run '^TestManagedCodexTUIAttachCommand$'` 같은 시험들이 가짜 codex(테스트 바이너리를 다시 실행하는 App Server·TUI 겸용 헬퍼)를 상대로 확인한다. 실제 TUI 는 쓰지 않는다. 실제 터미널 확인 절차는 `.moai/specs/SPEC-FACTORY-MANAGED-TUI-001/manual-check.md`.
- 레인 루프의 관리 카드 자식은 가짜 App Server 를 상대로 `go test ./internal/cli -run '^TestManagedCardChild'` 계열 시험(분기, 런치 형태, 환경, 앵커 락 부재, 연속 카드 바인딩, 세션 종료 뒤 루프 연속, 입력 전달)과 `TestManagedOperatorInputPumpDetachesEndedSession` 로 확인한다. 스위치를 끈 직접 exec 경로의 불변은 `TestManagedCardChildSwitchOffKeepsDirectDoor` 가 `*exec.Cmd` 전체를 리터럴 기대값과 비교한다. 실행 전에 레인 환경 변수를 한 호출 안에서 지워야 한다(`unset MOAI_KANBAN … MOAI_KANBAN_BACKEND && go test …`, 전체 목록은 SPEC 의 `acceptance.md` §1.1). 이 시험도 실제 codex 세션을 쓰지 않는다.
- 메시지가 오지 않을 때: `factory_msg_status` 도구로 브로커 상태(대기/확인 수)를 먼저
  보고, 레인이 launch-pending 에 머물러 있는지 확인한다.

SPEC: `.moai/specs/SPEC-FACTORY-MANAGED-SESSION-001/` (card t1375)
후속 SPEC: `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/` (card t1409, 서버 요청 응답과 턴 단위 실패 격리)
후속 SPEC: `.moai/specs/SPEC-FACTORY-MANAGED-CARD-CHILD-001/` (card t1440, `moai codex -l` 카드 자식의 관리 경로 연결)
