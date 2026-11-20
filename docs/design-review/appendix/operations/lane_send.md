# `lane-send`

operation `lane.send`(SPEC-20261119-lane-send-guard, 카드 C-120 조각 B). 코드: `src/ops/lane_send.rs`(명령)와 `moai_cli::lane_status::lane_send`(`src/lane_send.rs`, 판정과 `orca` 호출). 화면 읽기는 `lane-screen` 과 같은 분류기(`moai_cli::lane_status::lane_screen`)를 써요.

`lane-send --terminal HANDLE --text TEXT [--enter] [--now YYYY-MM-DDTHH:MM:SSZ]` 는 Orca 터미널 `HANDLE` 에 글 하나를 보내되, **보내기 전에 그 터미널의 그려진 화면을 읽고** 레인이 `working` 이거나 `idle` 일 때만 보내요. 2026-11-19 에 Yes/No 확인 창이 떠 있는 레인에 글과 Enter 가 들어가 `git reset --hard` 의 `Yes` 가 눌린 일을 막으려는 도구예요. `orca terminal send` 는 대상 화면을 보지 않아요.

- **읽기**: `orca terminal read --terminal HANDLE --screen --json` 을 부르고, 답이 Orca 의 JSON 이 아니면(글자만 오면) 화면으로 믿지 않고 `unreachable` 로 거절해요. `source` 가 `screen` 이 아닌 읽기는 `unknown` 이에요.
- **거절**: `working`·`idle` 이 아닌 모든 상태(`permission_prompt`·`input_pending`·`usage_limit`·`unknown`·`unreachable`)는 아무것도 보내지 않고 종료 코드 1 과 오류 글을 줘요. 오류 글에 상태, 이유, 읽어 낸 화면 줄 하나가 있어요. 번호가 붙은 `Yes`·`No` 줄 한 쌍은 번호 앞에 무엇이 있든(목록에 없는 커서 `◆` 도) 확인 창이라 거절해요(B1). 아래쪽 8줄 안에 확인 질문(`Do you want …` 류)이 있고 그 질문 줄이나 아래 줄에 선택 낱말(`yes`·`no`·`allow`·`deny`·`y`·`n`)이 있거나 아래에 번호 줄이 둘 이상인데 선택 줄로는 읽지 못한 화면도 `unknown` 이라 거절해요(R-N1, R-B1, `lane-screen` 과 같은 규칙: `Do you want to proceed? (yes/no)` 처럼 선택이 질문과 같은 줄에 있어도 거절). 질문으로 대화를 끝내고 입력 줄에서 기다리는 레인은 질문 줄과 그 아래에 선택 낱말이 없고 번호 줄이 둘 미만이면 거절하지 않고 보내요(S-N2). 대가는 입력 줄 아래 표시줄에 선택 낱말이 있는 레인도 거절된다는 것이에요. `usage_limit` 은 해제 시각도 UTC 로 적고, 시간대가 없거나 모르는 시간대면 "reset time is unknown" 이라고 해요(`lane-screen` 과 같은 규칙).
- **두 번 읽기**: 보내기 직전에 한 번 더 읽어서 상태가 첫 읽기와 다르면 보내지 않아요(`working` 에서 `idle` 로만 바뀌어도 거절해요). 두 읽기 사이의 틈은 0 이 되지 않아요.
- **글 검사**(`orca` 를 부르기 전에): 비어 있지 않고, `-` 로 시작하지 않고(`orca` 의 옵션으로 읽힐 수 있어요), 2000자 이하이고, `lane-duty` 의 입력 안전 규칙(제어 문자·줄 구분자·양방향 문자 거절)을 통과해야 해요. 거절한 글은 오류 글에 되풀이하지 않아요. 터미널 핸들은 영문자·숫자·`.`·`_`·`-` 128자 이하이고 영문자나 숫자로 시작해야 해요.
- **Enter**: `--enter` 를 줄 때만 보내요. 기본은 글만 넣어요.
- **보내기 실패**: `orca terminal send` 가 실패하면 종료 코드 1 과 "lane send failed" 오류를 줘요. 보냈다고 말하지 않고 다시 보내지도 않아요. 프로그램이 종료 코드 0 으로 끝나도 답을 읽어서, 답이 JSON 객체이고 `ok` 가 `true` 일 때만 보냈다고 해요(읽기 쪽과 같게). `ok: false` 면 "answered ok:false" 와 Orca 의 오류 글 한 줄(제어 문자 없이 160자까지)로, 답이 글자뿐이거나 비어 있거나 `ok` 가 없거나 불리언이 아니면 "not an Orca answer" 와 "글이 레인에 닿았을 수 있어요" 로 실패해요(B2).
- **`orca` 호출**: 프로그램은 `MOAI_ORCA_BIN`(없으면 PATH 의 `orca`)이고, 호출 하나는 15초(`MOAI_ORCA_TIMEOUT_SECS` 로 1에서 300초) 안에 끝나야 해요. 없거나 늦으면 `unreachable` 로 거절해요. 프로그램이 끝난 뒤 그 출력을 읽는 데도 같은 제한이 걸려요: 자식 프로세스가 출력을 계속 쥐고 있으면 제한(적어도 0.25초의 말미)이 지난 뒤 읽기를 버리고 "did not finish its output" 으로 거절해요(`lane-screen --scan` 도 같은 호출을 써요).

성공하면 결과에 `sent`(참), `terminal`, `state`(보낸 때의 상태), `evidence`(읽은 화면 줄), `enter`, `orca`(`orca terminal send` 의 JSON 답)가 있어요. 이 도구는 확인 창에 대신 답하지 않고, 기다렸다 다시 보내지도 않아요. `orca terminal send --help` 는 다른 에이전트에게는 `orca orchestration send` 를 쓰라고 해요: 이 도구는 그 길을 바꾸지 않고 `terminal send` 앞에서 화면만 지켜요.
