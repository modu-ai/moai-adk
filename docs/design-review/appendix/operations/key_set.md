# `key-set`

operation `key.set` · 코드 `src/ops/key_set.rs`, `src/key_prompt.rs` · #2393, 카드 C-54 S1

`key-set PROVIDER [--shared] [--prompt]` 는 공급자 키를 암호화해 금고에 넣는다(지금 프로젝트, 또는 `--shared` 로 모든 프로젝트). 값은 stdin 으로 받고(argv 아님), 대화형 터미널이면 거절한다. 답에는 값이 없다. `src/main.rs` 에서 옮김(#2834 R3).

`--prompt`(카드 C-54 S1, 사용자 결정 2026-11-16 "Key entry"): 값을 터미널의 숨긴 입력란에서 받는다.

- stdin 이 터미널일 때만 묻는다. 터미널이 아니면 거절하고, stdin 을 읽지 않으며 아무것도 저장하지 않는다. `--prompt` 가 없으면 동작은 전과 같다.
- 묻기 전에 공급자 이름과 범위를 먼저 확인한다. 질문은 stderr 에 한 줄로 나온다.
- 입력하는 동안 터미널 에코를 끄고(Unix `termios`, 이미 쓰는 `libc`; 새 크레이트 없음), 끝나면 원래 모드로 되돌린다. 값을 저장했을 때, 거절했을 때, 오류가 났을 때, 그리고 `SIGINT`(Ctrl-C)·`SIGTERM`·`SIGHUP`·`SIGQUIT` 로 끝날 때 모두 되돌린다. 묻는 동안 Ctrl-Z(일시 중지)는 꺼 둔다.
- 한 줄을 받는다: Enter 또는 Ctrl-D 로 끝나고, Backspace 는 한 글자, Ctrl-U 는 줄 전체를 지운다. 그 밖의 제어 문자(화살표 키 포함)가 들어오면 값 전체를 거절한다.
- 값 규칙과 크기 한도는 stdin 과 같다: 빈 값·공백뿐인 값·제어 문자는 거절, 65,535 바이트까지 저장하고 65,536 바이트부터 거절한다. 한도를 넘은 줄은 잘라 저장하지 않는다.
- 값, 길이, 지문은 터미널에도 stdout 에도 나오지 않는다. 답은 stdin 으로 넣었을 때와 같은 바이트다.
- Unix 가 아니면(`termios` 없음) 거절한다: 값을 파이프로 넣으라고 알려 준다.
- 에이전트 세션의 셸 도구에는 터미널이 없으므로 `--prompt` 는 사용자가 자기 터미널 창에서 실행한다. 세션은 그 뒤 `key-status` 로 확인한다.

시험: `tests/key_set_prompt_tty_test.py`(pty), `tests/key_vault_cli.rs`, `src/key_prompt.rs` 의 단위 시험.
