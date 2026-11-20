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

(조각 1 을 마치면 채워요.)
