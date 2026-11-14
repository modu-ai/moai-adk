# `lane-duty-line`

operation `lane.duty_line` · 코드 `src/ops/lane_duty_line.rs`, `src/lane_duty.rs` · Issue #2902 O3a (계획 6874299470, 리더 결정 6874308433)

`lane-duty-line --lane NAME [--first TASK] [--host claude|codex|omp]` 는 리더가 레인 터미널에 입력할 한 줄을 `result.line` 으로 돌려준다. 읽기만 하고 아무것도 쓰지 않는다.

- 형식: Claude·OMP 는 `/moai --mode-agent 레인 <NAME> — <BODY>`, Codex 는 `$moai --mode-agent 레인 <NAME> — <BODY>`. `--host` 기본값은 `claude` 이고 다른 값은 명령줄 오류(`usage`)다.
- BODY 는 사용자가 정한 고정 문구다(`moai_cli::lane_duty::BODY`, golden `tests/lane_duty_line.rs`). 정본은 O1 #2900 문서이고, 그 문구가 바뀌면 상수와 golden 을 함께 고친다.
- `--first TASK` 는 끝에 ` 먼저: TASK` 를 붙인다. 이 줄은 다른 터미널에 입력되므로(화면 주입), 제어 문자(탭·줄바꿈·CR·ESC·DEL 등 C0/C1), 줄·문단 구분자 U+2028·U+2029, 양방향·보이지 않는 서식 문자(U+200B–200F, U+202A–202E, U+2060–2064, U+2066–2069, U+061C, U+FEFF)가 든 TASK, 빈 TASK, 300자를 넘는 TASK 는 `invalid_input` 으로 거부한다. 거부 메시지에 TASK 를 되풀이하지 않는다.
- 레인 이름은 `session-register` 와 같은 규칙(`lane_registry::valid_name`: `A-Z a-z 0-9 . _ -` 1~64자, 첫 글자는 영문자나 숫자)이다.

`session-register` 의 리더 해제(같은 Issue): `--show --name NAME` 이 현재 holder 의 `instance`·`lease_until`(또는 `free`)을 쓰지 않고 보여 주고, 리더는 그 instance 로 `--release --by leader` 를 부른다(`session_register.md`).
