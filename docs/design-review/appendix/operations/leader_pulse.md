# `leader-pulse`

operation `leader.pulse` · 코드 `src/ops/leader_pulse.rs`, `src/leader_pulse.rs` · Issue #3015 C5 (PLAN 댓글 6905006699, 리더 결정 #1975: 이름 `leader-pulse`, CLI 는 글과 명령 줄만 만들고 `orca` 호출은 스킬이 함, 턴이 안 시작되면 보고만)

`leader-pulse --terminal HANDLE --text LINE [--interval-secs 300] [--rounds N]` 은 리더가 자기 터미널에 N초마다 넣을 라운드 지시와 `orca terminal send` 명령 줄을 검사해서 낸다. 아무것도 실행하지 않고(`writes: false`) 터미널도 읽지 않는다.

- 입력 검사: 핸들은 1–200자 `A-Z a-z 0-9 . _ : -`, `-` 로 시작하지 않음. 글은 한 줄, 1–1000자, 제어 문자 없음, `-` 로 시작하지 않음(옵션으로 읽히지 않게). 간격 60–3600초(기본 300). 횟수 1–10000(없으면 멈출 때까지).
- 출력: `terminal`, `text`, `interval_secs`, `rounds`, `program: "orca"`, `args`(인자마다 따로), `command_line`(POSIX 셸용 따옴표), `rules`, `writes: false`, `verified: not_run`.
- 명령: `orca terminal send --terminal <handle> --text <글> --enter --wait-submit 20 --json`. `orca terminal create` 와 `--interrupt` 는 쓰지 않는다.
- `rules`: `create_terminal: false`, `resend_when_turn_not_started: false`, `on_turn_not_started: "report"`. 스킬(`mode-leader.md` Orca 절 9번)이 간격마다 명령을 한 번 돌리고, 턴이 안 시작되면 보고만 한다.
- 실측: 리더가 사람 입력 없이 2회 이상 라운드를 도는 실측(#3015 AC-5)은 아직 하지 않았다(`verified: not_run`). `orca terminal send --json` 의 영수증은 해석하지 않는다.
