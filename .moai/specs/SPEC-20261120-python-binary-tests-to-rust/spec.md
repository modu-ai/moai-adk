---
id: SPEC-20261120-python-binary-tests-to-rust
title: "Product behaviour tests that run the built binary live in Rust; the Python modules that did are ported, deleted as duplicates, or kept with a recorded reason"
status: draft
revision: 1
created: 2026-11-20
---

# 바이너리를 실행하는 파이썬 테스트를 Rust 로

## Intent

사용자 결정(2026-11-20): 제품 동작 테스트는 Rust 로만 둬요. 지금 `tests/*_test.py` 중 18개가 빌드한 `moai` 바이너리를 실행하고(`rg -l -e rust_bin -e MOAI_CLI_BIN -e target/debug tests -g '*_test.py'` 로 센 값), 그중 여럿은 Rust 통합 테스트와 같은 흐름을 따로 한 번 더 단정해요. 예: `tests/update_tux_tty_test.py` 와 `tests/update_command_cli.rs` 는 둘 다 `update` → 저장 흐름을 단정하고, 오늘 둘 중 파이썬만 실패했어요(기본값이 바뀐 뒤 한쪽만 고쳐짐). 같은 사실을 두 곳에서 쓰면 한쪽이 낡아요.

이 SPEC 은 18개를 영역별 조각으로 나눠 하나씩 처리해요. 모듈마다 셋 중 하나예요: (a) Rust 테스트가 이미 같은 사실을 단정하면 파이썬을 지움, (b) 없으면 Rust 통합 테스트로 옮김, (c) 바이너리를 곁다리로만 부르는 개발 도구 테스트(도구 자체가 파이썬·셸)이면 이유를 적고 남김. 기대값은 글자를 베끼지 않고 소스에서 끌어와요(`model_effort_settings::default_choice()`, `supported()`, JSON 출력).

이 조각(조각 1)은 update 영역이에요: `tests/update_tux_tty_test.py`(테스트 10개)를 `tests/update_command_cli.rs` 로 옮겨요. 터미널 흐름은 지금 Rust 에서 BSD `script` 로 돌아 macOS 에서만 도는데, 이 조각이 `libc::openpty` 로 만든 공용 도구 `tests/common/pty.rs` 를 두어 리눅스와 macOS 에서 같이 돌게 해요. 이미 `tests/help_layout.rs` 가 같은 `openpty` 호출을 써서 새 크레이트는 필요 없어요.

## Non-goals

- 도구 자체를 시험하는 파이썬 모듈(`ci_scope_test`, `local_check_test`, `check_guard` 계열 등 26개 도구)을 Rust 로 옮기기. 개발 도구는 `moai` 하위 명령으로 옮겨질 때 따로 다뤄요.
- 조각 1 밖의 17개 모듈(아래 Plan). 각 조각에서 이 SPEC 의 Result 에 처분을 적어요.
- 제품 코드(`src/**`) 변경. 이 조각은 테스트 파일만 바꿔요.
- 새 크레이트. `libc` 는 이미 의존성이에요.

## Requirements

- REQ-1: When a Rust integration test asks `tests/common/pty.rs` to run a command at a terminal, the helper shall give the command one pseudo terminal made with `libc::openpty` as stdin, stdout and stderr (or `/dev/null` as stdin when the case asks for a non-terminal stdin), type the case's answer only after `[y/N] ` ends the screen, and return the screen text with `\r\n` read as `\n` and the exit code.
- REQ-2: When `tests/update_command_cli.rs` runs a terminal case, the case shall go through `common::pty` without `/usr/bin/script` and without `cfg(target_os = "macos")`, so the same cases run on Linux and macOS · failure: none — the file is a test; a case that behaves differently on one system names the system in its message instead of hiding behind a cfg.
- REQ-3: When `tests/update_tux_tty_test.py` is removed, `tests/update_command_cli.rs` shall hold a test for every one of its ten cases that no Rust test already covers: the pipe envelope with `--json` giving the same bytes, `--json` on a terminal printing the envelope, end of input at the question writing nothing, end of input at the model stage keeping the applied update and writing no settings, a package changed after the preview being refused with the new preview, a terminal whose stdin is not a terminal being shown the confirming command and never asked, an error on a terminal showing the frame while a pipe still gets the failure envelope, and a Korean setting keeping the English screens.
- REQ-4: When the terminal cases check the saved model settings, the expected model and effort shall be the values of `model_effort_settings::default_choice("claude", "leader")` read at run time and never a copied literal · failure: none — a missing default fails the test with its tool and role in the message.
- REQ-5: When the whole `tests/update_command_cli.rs` runs, each ported case shall fail when the behaviour it names is broken, shown by at most five mutants of the source that turn a test red and are restored · failure: none — a mutant that survives names a missing assertion, and the test gets it before this requirement counts as met.
- REQ-6 (REQ-1): IF the command has not ended 60 seconds after it started or after the answer was typed, THEN `tests/common/pty.rs` shall kill the command and fail with the screen seen so far.
- REQ-7 (REQ-3): IF a case of the Python module cannot be written as a Rust test, THEN the Result shall name that case with its reason and the Python module stays in place.

## Acceptance

- AC-1 (REQ-1): 공용 도구가 질문 뒤에만 답을 입력하고, 60초 안에 안 끝나는 자식은 죽이고 화면을 담아 실패해요.
  verify: `cargo test --offline --locked --test pty_helper`
- AC-2 (REQ-2): update 의 터미널 테스트 6개가 `script` 없이 `common::pty` 로 돌고 macOS cfg 가 없어요.
  verify: `cargo test --offline --locked --test update_command_cli`
- AC-3 (REQ-3): 파이썬 10개 사례가 모두 Rust 테스트로 있거나 그대로 남은 이유가 Result 에 있고, 다 옮겼으면 `tests/update_tux_tty_test.py` 가 없어요.
  verify: `cargo test --offline --locked --test update_command_cli`; `git ls-files tests/update_tux_tty_test.py` 가 아무것도 내지 않아요.
- AC-4 (REQ-4): 저장된 모델과 effort 의 기대값이 `default_choice` 에서 와요.
  verify: `cargo test --offline --locked --test update_command_cli -- model_stage`
- AC-5 (REQ-5): 변형 5개 이하가 모두 잡혀요.
  verify: Result 절의 변형 표.

## Scope

- write: `.moai/specs/SPEC-20261120-python-binary-tests-to-rust/**`, `tests/update_tux_tty_test.py`(삭제), `tests/update_command_cli.rs`, `tests/common/pty.rs`, `tests/common/mod.rs`(`pub mod pty;` 한 줄), 시험 `tests/pty_helper.rs`
- read-only: `src/update_tux.rs`, `src/update.rs`, `src/model_effort_settings.rs`, `tests/help_layout.rs`, `tests/self_audit_run.py`
- forbidden: `src/**`(제품 코드), 실제 HOME, 네트워크 쓰기, 다른 조각의 모듈

## Plan

조각(영역) 순서. 조각 1 만 이번 head 에서 해요. 나머지 처분은 각 조각의 Result 에 적어요.

1. update: `update_tux_tty_test.py` ↔ `update_command_cli.rs` (이 조각).
2. 터미널 화면: `ui_tty_test.py`, `ui_korean_tty_test.py`, `ui_summary_tty_test.py`, `ui_progress_confirm_test.py`, `error_frame_tty_test.py`, `help_english_tty_test.py`, `install_inspect_tty_test.py`, `key_set_prompt_tty_test.py`.
3. 개발 도구 테스트가 바이너리를 곁다리로 부르는 것: `build_folder_test.py`, `goal_block_test.py`, `hook_latency_test.py`, `ci_smoke_step_test.py`, `self_check_test.py`, `smoke_harness_test.py`, `github_assets_test.py`, `heavy_check_slot_test.py`, `repo_scans_test.py` — 모듈마다 바이너리를 부르는 사례만 가려 (a)(b)(c) 처분.

조각 1 단계:

1. RED: `tests/pty_helper.rs`(도구 시험)와 옮길 사례를 `tests/update_command_cli.rs` 에 먼저 쓰고 실패를 확인해요.
2. GREEN: `tests/common/pty.rs` 를 두고 기존 `script` 호출 6곳을 바꿔요.
3. 파이썬 사례마다 Rust 사례와 대조표를 만들고, 다 옮겼으면 `tests/update_tux_tty_test.py` 를 지워요.
4. 변형 5개, 이웃 검사(`home_isolation_scan`, `repo_scans` 목록 전부), macOS 에서 실행, 리눅스는 로컬에서 못 돌리므로 CI 결과로만 확인한다고 Result 에 적어요.

## Decisions

- D1 (되돌릴 수 있음, 범주 none): 공용 pty 도구는 `libc::openpty` 를 써요(`tests/help_layout.rs` 와 같은 호출). 대안: `script` 를 OS 별 인자로 감싸기(BSD 와 util-linux 의 인자가 달라 두 갈래를 유지해야 함), `portable-pty` 크레이트(새 의존성). 되돌리기: 도구 파일 하나를 지우고 호출처를 되돌려요.
- D2 (되돌릴 수 있음, 범주 none): 파이썬이 실제 `skills/moai` 와 `agents/` 를 복사해 돌리던 사례도 Rust 쪽은 기존 최소 픽스처(`v1()`)로 옮겨요. 제품 패키지 내용에 기대는 단정이 아니라 `update` 흐름이 목적이라서요. 패키지가 실제로 설치되는지는 `product_cli skill_package::` 가 이미 봐요.
- D3 (되돌릴 수 있음, 범주 none): 화면 글자("Nothing was written." 등)는 `src/update_tux.rs` 안 함수의 비공개 값이라 소스에서 끌어올 길이 없어서 글자로 단정해요. 모델·effort 만 `default_choice` 에서 끌어와요(카드의 `default_choice()` 예).
- D4 (되돌릴 수 있음, 범주 none): 리눅스 실행은 이 Mac 에서 못 해요. 컴파일(`cargo check --tests`)과 macOS 실행만 로컬에서 하고, 리눅스는 CI 의 결과를 Result 에 적기 전에는 "리눅스 통과"라고 쓰지 않아요.

## Risks

- 리눅스 의사 터미널은 자식이 끝나고 슬레이브가 닫히면 마스터 읽기가 `EIO` 로 끝나요(macOS 는 0 을 돌려줘요). 도구는 둘 다 끝으로 읽어요.
- 자식이 컨트롤 터미널 없이 슬레이브를 stdin 으로 받아요(파이썬 도구와 같은 방식). `setsid` 를 쓰지 않아요.
- 입력을 질문 앞에 미리 쓰면 에코로 화면에 섞여 "질문 뒤에만 입력" 단정이 깨져요. 도구는 화면이 `[y/N] ` 로 끝날 때까지 기다린 뒤에만 써요.

## Result

### 조각 1 (update) — head c03e375005996701b2c4dd53db3315c2b586b2bb

- 만든 것: `tests/common/pty.rs`(`libc::openpty` 한 쌍, 질문 `[y/N] ` 뒤에만 답 입력, 질문에 답이 없으면 end of input, 60초 한도에 죽이고 화면을 담아 실패), `tests/pty_helper.rs`(도구 시험 8개), `tests/update_command_cli.rs` 의 터미널 시험 6개를 `script` 와 macOS cfg 없이 도구로 옮김, 파이썬 사례에서 Rust 에 없던 8개 시험 추가, `tests/update_tux_tty_test.py` 삭제.
- 대조표(파이썬 10개 사례 → Rust):
  - 파이프 봉투와 `--json` 같은 바이트 → `a_pipe_keeps_the_envelope_and_json_is_the_same_bytes`
  - 터미널 화면(MoAI update, Checking, 봉투 없음, N 은 쓰지 않음) → `in_a_terminal_update_asks_and_applies_only_on_yes` 에 단정 추가
  - 터미널의 `--json` → `json_on_a_terminal_prints_the_envelope`
  - y 로 적용, 끝 화면, 두 번째 실행 → `in_a_terminal_an_applied_update_asks_the_model_stage_and_saves_after_a_yes` 에 단정 추가
  - 모델 단계의 end of input → `end_of_input_at_the_model_stage_keeps_the_applied_update`
  - 질문의 end of input → `end_of_input_at_the_question_writes_nothing`
  - 미리보기 뒤 패키지 변경 → `a_plan_that_changes_before_the_write_is_refused`
  - stdin 이 터미널이 아님 → `stdin_that_is_not_a_terminal_is_never_asked`
  - 터미널의 오류는 프레임 → `an_error_on_a_terminal_is_the_frame_not_json`
  - 한국어 설정도 영어 화면 → `a_korean_setting_keeps_the_english_screens`
- 소스에서 끌어온 값: 저장된 모델·effort 는 `model_effort_settings::default_choice("claude", "leader")`, 도구 질문 문장은 `model_effort_questions::tool_question(&[]).prompt`. 화면 글자는 비공개 함수의 값이라 글자로 단정해요(D3).
- 이 head 에서 돌린 것(`python3 -I target/run-ready.py` → `target/ready.log`, heavy-slot): `cargo fmt --check` 0, `cargo clippy --all-targets -D warnings` 0, `update_*` 대상 전부와 `pty_helper` 213 passed(그중 `update_command_cli` 27, `pty_helper` 8), `--lib` 전체 1604 passed, `--bins` 56 passed, macOS cfg 대상 6개 112 passed, 파이썬 `tests/update_*_test.py` 는 이제 없음, `tools/repo_scans.json` 39개 검사 모두 exit 0, 마지막 줄 `DONE failed steps: []`.
- 변형 5개(상한 5, 모두 잡힘, 변형 뒤 트리 변경 0): 확인 사이 계획이 바뀌어도 적용(`without_a_matching_confirmation_nothing_is_written`), stdin 이 터미널이 아니어도 질문(`stdin_that_is_not_a_terminal_is_never_asked`), `--json` 무시(`json_on_a_terminal_prints_the_envelope`), 빈 답·end of input 을 yes 로(`end_of_input_at_the_question_writes_nothing`), 낡은 계획 거절 문장 교체(`a_plan_that_changes_before_the_write_is_refused`). 첫째 변형은 옛 시험(`without_a_matching_confirmation_nothing_is_written`)만 잡았고, 새 `a_plan_that_changes_before_the_write_is_refused` 는 다섯째 변형(거절 문장)만 잡았어요. 첫째 변형이 고친 확인 줄(`src/update.rs` 의 `fresh.plan_digest != digest`)은 터미널 yes 흐름이 거치지 않는 줄이라, 터미널 흐름의 낡은 계획 거절을 코드로 깨뜨리는 변형은 이 조각에 없어요(한계로 남겨요).
- 리눅스: 이 Mac 에서 돌릴 수 없어요(D4). 같은 `openpty` 호출이 `tests/help_layout.rs` 에서 이미 리눅스 CI 를 통과하고 있다는 점이 근거이고, 새 시험의 리눅스 결과는 CI 가 처음 보여 줘요. 그 전에는 "리눅스 통과"라고 쓰지 않아요.

### 조각 2 (터미널 화면) — 코드 head 8d61808e27e6502ee10e5e2302d1bfd535039341

- 만든 것: 터미널 화면 파이썬 8개(67 사례)를 Rust 시험 8개 파일(67 시험)로 옮기고 파이썬 8개를 모두 지웠어요. 공용 도구(`tests/common/pty.rs`)는 stdout 이나 stderr 를 파이프로 받아 돌려주기(`stdout_not_a_terminal`, `stderr_not_a_terminal`), stdin 에 글 넣기(`stdin_text`), 한 줄로 쓰는 실행 함수(`piped`, `piped_with`, `stderr_on_terminal`, `stdout_on_terminal`, `env_with`)를 얻었고, 화면 읽기는 `tests/common/screens.rs`(색 시퀀스 제거, 표 행 찾기)에 둬요. `tests/pty_helper.rs` 는 8개에서 13개로 늘었어요.
- 대조표(파이썬 모듈 사례 수 → Rust 파일 시험 수, 모두 같아요):
  - `install_inspect_tty_test.py` 3 → `tests/install_inspect_terminal_cli.rs` 3
  - `help_english_tty_test.py` 3 → `tests/help_english_terminal_cli.rs` 3
  - `ui_tty_test.py` 9(배너 7, 봉투 2) → `tests/ui_terminal_cli.rs` 9
  - `ui_summary_tty_test.py` 7 → `tests/ui_summary_terminal_cli.rs` 7
  - `error_frame_tty_test.py` 10 → `tests/error_frame_terminal_cli.rs` 10
  - `ui_progress_confirm_test.py` 8(key-remove 5, index-build 2, forge-read 1) → `tests/ui_progress_confirm_terminal_cli.rs` 8
  - `ui_korean_tty_test.py` 12(화면 8, 오류 프레임 4) → `tests/ui_korean_terminal_cli.rs` 12
  - `key_set_prompt_tty_test.py` 15 → `tests/key_set_prompt_terminal_cli.rs` 15(그리고 자기 재실행용 `prompt_shell` 1개, 아래)
- 달라진 점: (1) 숨김 입력 시험은 파이썬 보조 셸 대신 시험 프로그램을 같은 이름으로 다시 실행해 세션 대장(`setsid`, `TIOCSCTTY`)으로 쓰고(`prompt_shell`, 환경 변수가 없으면 아무것도 안 해요), 명령의 stdout 은 그 보조 셸이 다섯 번째 파일 기술자로 받아요. 종료 신호는 `ExitStatus::signal()` 로 `-신호` 로 보고해 파이썬과 같은 값(`-SIGINT`, `128 + SIGQUIT` 등)을 단정해요. (2) 정규식 크레이트가 없어 digest 치환과 `ESC [ … m` 제거를 손으로 짰어요. (3) 파이썬 `capture_error` 의 `binary` 인자는 이 모듈의 사례가 쓰지 않아서 옮기지 않았어요. (4) 이 모듈이 부르던 `self_audit_run.rust_bin` 은 `common::MOAI_CLI_BIN` 이 대신해요.
- 소스에서 끌어온 값: 화면 기대값은 `tests/fixtures/ui/en/*.golden`(`home-routing-plan`, `key-remove-refused`, `index-build-stages`, `error-*`)에서 읽고, 안내 문장(`install-inspect` 의 PATH 문장, 빈 stdin 오류 등)은 파이썬 시험이 쥐고 있던 글자를 그대로 가져왔어요. 이 글자들은 제품 소스의 비공개 값이라 끌어올 길이 없어요(D3 와 같은 이유).
- CI 범위 규칙: `tools/ci_scope.py` 는 `_tty_test.py` 로 끝나는 파일을 바꾸는 PR 에만 macOS 검사를 붙였어요. 모듈을 Rust 로 옮기면 새 시험을 바꾸는 PR 이 macOS 를 건너뛰게 되므로 `_terminal_cli.rs` 이름과 `tests/common/pty.rs`, `tests/pty_helper.rs` 를 같은 `pty/terminal` 분류에 넣고 `tests/ci_scope_test.py` 에 세 사례를 더했어요. 옛 `_tty_test.py` 규칙은 남겨 두었어요(`ci_scope_test` 가 옛 PR 파일 목록으로 확인해요). 그 밖의 참조 둘(`tests/key_vault_cli.rs` 주석, `docs/design-review/appendix/operations/key_set.md`)은 Rust 시험을 가리키게 바꿨어요. `.moai/specs/SPEC-20261118-project-decide-model/spec.md` 의 "read-only" 줄은 다른 카드의 기록이라 건드리지 않았어요.
- 이 head 에서 돌린 것(`cargo test --offline --locked`, heavy-slot): `error_frame_terminal_cli` 10, `key_set_prompt_terminal_cli` 16, `pty_helper` 13, `ui_korean_terminal_cli` 12, `ui_progress_confirm_terminal_cli` 8, `ui_summary_terminal_cli` 7, `ui_terminal_cli` 9, `home_isolation_scan` 18 모두 passed(exit 0); 앞서 `install_inspect_terminal_cli` 3, `help_english_terminal_cli` 3 passed. 이웃 검사 전체(`python3 -I target/run-ready.py` → `target/ready.log`: fmt, clippy 전 대상, 터미널·update 대상, `--lib`, `--bins`, macOS cfg 대상, `tests/ci_scope_test.py`, `repo_scans` 39개)는 이 항목을 쓰는 시점에 진행 중이에요. 결과는 끝나는 대로 이 항목에 적어요.
- 변형 5개(상한 5, 모두 잡힘, 변형 뒤 트리 변경 0, 시험 대상은 새 Rust 파일만): ① `src/banner.rs` 의 `w <= LOGO_WIDTH` 를 `w < LOGO_WIDTH` → `ui_terminal_cli` 의 `a_narrow_terminal_gets_the_line_not_the_logo`; ② `src/ui.rs` 오류 색을 빨강에서 초록으로 → `an_operation_error_is_framed_on_a_terminal_and_unchanged_on_a_pipe`, `the_color_flag_comes_before_the_environment`; ③ `src/key_prompt.rs` 신호 처리기의 `TCSAFLUSH` 를 `TCSANOW` 로 → `a_signal_drops_what_was_typed_and_not_read_yet` 외 2개; ④ 같은 파일 종료 신호 목록에서 `SIGHUP` 을 `SIGINT` 로 바꿔 빼기 → `ending_signals_leave_the_terminal_as_it_was_and_store_nothing`; ⑤ `src/ui_text.rs` 의 둘째 단계 줄을 `(2/4)` 로 → `stages_show_on_a_terminal_and_never_on_a_pipe` 외 2개.
- 리눅스: 이 Mac 에서 돌릴 수 없어요(D4). 새 시험의 리눅스 결과는 CI 가 처음 보여 줘요. 신호·`setsid`·`TIOCSCTTY` 를 쓰는 `key_set_prompt_terminal_cli` 가 가장 눈여겨볼 곳이에요. 그 전에는 "리눅스 통과"라고 쓰지 않아요.

### 수용 기준 결과

- AC-1: met — `cargo test --test pty_helper` 8 passed(질문 뒤 입력, 터미널 stdin 없음, 정리, 한도에 죽임, 깨끗한 환경 포함).
- AC-2: met — `cargo test --test update_command_cli` 27 passed, 터미널 시험 6개에 `script` 와 `cfg(target_os = "macos")` 가 없어요.
- AC-3: met — 대조표의 10개 사례가 모두 Rust 시험으로 있고 `git ls-files tests/update_tux_tty_test.py` 는 아무것도 내지 않아요.
- AC-4: met — 저장된 모델과 effort 의 기대값이 `default_choice` 에서 와요.
- AC-5: met — 변형 5개 모두 잡힘(위).

### 남은 조각(카드는 열려 있어요)

- 조각 3: 개발 도구 테스트가 바이너리를 곁다리로 부르는 9개 — 처분과 이유.
