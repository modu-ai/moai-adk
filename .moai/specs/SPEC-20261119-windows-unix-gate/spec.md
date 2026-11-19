---
id: SPEC-20261119-windows-unix-gate
title: "A scan test fails when a file under tests/ names std::os::unix where Windows still compiles it"
status: draft
revision: 1
created: 2026-11-19
---

# A scan test fails when a file under tests/ names std::os::unix where Windows still compiles it

## Intent

카드 C-110(p1)이에요. tip의 push run `windows-build`(Windows에서 `cargo build --all-targets`)가 PR #3216 병합 뒤부터 빨개요(run 44461021913, 44462889815). 원인은 `tests/hooks_cli.rs:1391`의 `std::os::unix::fs::symlink`에 `cfg`가 없다는 거예요. PR의 CI는 Linux(와 필요하면 macOS)만 돌려서, 병합 전에는 아무 검사도 이걸 못 봤어요.

이 SPEC은 같은 사실을 Linux·macOS에서 읽는 스캔 시험을 더해요: `tests/**/*.rs`에서 `os::unix`를 쓰는 줄은 모두 Windows에서 `cfg`가 꺼 버리는 항목·블록·문장·파일 안에 있어야 해요. 컴파일러가 아니라 글자 읽기라서 주석·문자열을 먼저 지우고, 한 번 훑어요.

`tests/hooks_cli.rs`의 실제 게이트(`#[cfg(unix)]`)는 C-86이 그 파일을 예약하고 있어서 이 SPEC의 쓰기 범위 밖이에요. 리더가 카드 노트로 정했어요(결정 (a)): 이 SPEC과 스캔 시험을 push하면 agent-3이 그 위에 `cfg(unix)` 커밋 하나를 올려요. 그래서 이 가지의 스캔 시험은 `tests/hooks_cli.rs`가 고쳐지기 전까지 그 한 줄(`tests/hooks_cli.rs:1391`)에서 실패하는 것이 맞아요(RED).

## Non-goals

- `tests/hooks_cli.rs`와 다른 시험 파일 고치기(C-86, agent-3).
- `src/**`와 CI 워크플로(`.github/workflows/**`) 바꾸기.
- 컴파일러 수준의 `cfg` 해석: 이 스캔은 글자 읽기예요. `cfg` 값은 참·거짓·모름 셋으로 읽고 모름(`test`, 기능, `target_vendor` 등)은 게이트로 보지 않아요. 그래서 모르는 이름의 게이트 아래 `os::unix`는 보고돼요(작성자가 `unix`·`target_os`·`target_family`로 게이트를 써요). 알려진 한계는 Risks에 적었어요.

## Requirements

- REQ-1: The scan test shall fail when a `.rs` file under `tests/` names `os::unix` outside an item, block or statement that a `cfg` attribute removes on Windows.
- REQ-2: The scan test shall treat `#[cfg(unix)]`, `#[cfg(not(windows))]`, `#[cfg(target_os = "macos")]`, `#[cfg(target_family = "unix")]` and an `all(...)` that contains one of them as gates.
- REQ-3: The scan test shall treat `#[cfg(any(unix, windows))]`, `#[cfg(not(unix))]`, `#[cfg(windows)]`, `#[cfg(test)]` and `#[test]` as no gate.
- REQ-4: The scan test shall treat an inner `#![cfg(...)]` that Windows turns off as a gate for the whole file.
- REQ-5: The scan test shall skip comments, string literals, raw string literals and character literals when it looks for `os::unix`.
- REQ-6: The scan test shall judge a helper file under `tests/common/` (not `mod.rs`) as gated only when at least one file includes it and every including file is gated as a whole file.
- REQ-7 (REQ-1): IF the scan finds 100 `.rs` files or fewer under `tests/` (the folder moved or the walk failed), THEN the scan test shall fail with the file count it saw instead of passing on a short list.
- REQ-8: The scan test shall end a gate at the end of the item it stands on: a `;` outside brackets or the close of the next block.
- REQ-9: The scan test shall name each place it reports as `path:line`.
- REQ-10: The scan test shall read a `cfg` name it cannot decide (`test`, a feature, `target_vendor`) as no gate, also inside `not(...)`, and shall let one false part settle an `all(...)` and one true part settle an `any(...)`.
- REQ-11: The scan test shall read every top-level name of a `use std::os::{...}` group and report the group when one of the names is `unix`.

## Acceptance

- AC-1 (REQ-1, REQ-9): 이 가지에서 스캔 시험이 `tests/hooks_cli.rs:1391`을 `path:line`으로 대며 실패하고(RED), 그 줄이 `#[cfg(unix)]` 아래로 가면 통과해요.
  verify: `cargo test --locked --offline --test windows_unix_gate_scan no_file_under_tests`
- AC-5 (REQ-7, REQ-10, REQ-11): 파일 수 하한은 100/101 경계 시험이 있고, 모르는 `cfg` 이름과 `not(모름)`은 게이트가 아니고, brace 묶음은 이름마다 읽혀요.
  verify: `cargo test --locked --offline --test windows_unix_gate_scan the_file_count_floor`; `cargo test --locked --offline --test windows_unix_gate_scan a_name_the_scan_cannot_decide`; `cargo test --locked --offline --test windows_unix_gate_scan the_names_of_a_brace_group`
- AC-2 (REQ-2, REQ-3, REQ-4, REQ-5, REQ-8): 상자(fixture) 시험이 게이트가 되는 꼴과 안 되는 꼴, 안쪽 `#![cfg]`, 주석·문자열·수명 표기, 게이트가 끝나는 자리를 하나씩 못 박아요.
  verify: `cargo test --locked --offline --test windows_unix_gate_scan -- --skip no_file_under_tests`
- AC-3 (REQ-6): `tests/common/` 도우미는 포함하는 파일이 모두 파일 단위로 게이트일 때만 안전하다고 읽혀요(`mod.rs`는 예외).
  verify: `cargo test --locked --offline --test windows_unix_gate_scan a_helper_under_common`
- AC-4: 이웃 검사.
  verify: `cargo test --locked --offline --test home_isolation_scan`; `python3 -m unittest tests.ci_scope_test`

## Scope

- write: `.moai/specs/SPEC-20261119-windows-unix-gate/spec.md`, `tests/windows_unix_gate_scan.rs`
- forbidden: 실제 HOME 쓰기, `tests/hooks_cli.rs`(C-86 예약; agent-3이 게이트 커밋을 올려요), `src/**`, `.github/workflows/**`

## Plan

1. SPEC을 쓰고 시험을 먼저 더해요(RED): tip에서 스캔 시험이 `tests/hooks_cli.rs:1391`로 실패해야 해요.
2. 상자 시험을 더하고, 스캔의 한 곳씩을 바꾸는 변이가 시험에 잡히는지 읽어요.
3. 로컬 Windows 확인(`cargo check --tests --target x86_64-pc-windows-msvc`)을 해요. 이 맥에서는 sqlite 의 C 컴파일(CLT 헤더에 `stdlib.h` 없음)이 막으므로, 저장소 밖의 가짜 C 컴파일러와 `ar` 스크립트(`CC_x86_64_pc_windows_msvc`·`AR_x86_64_pc_windows_msvc`)로 빌드 스크립트만 통과시켜 타입 검사를 읽어요. 링크는 안 해요.
4. lib·fmt·clippy `--all-targets`를 돌리고 push해요. 게이트 커밋은 agent-3이 위에 올려요.

## Risks

- 글자 읽기라서 스캔이 놓치는 꼴이 있어요(알려진 한계; 이 저장소의 실제 트리에는 없어요): 게이트가 붙은 enum 변형·구조체 필드·match 팔(게이트가 다음 블록까지 이어져요), `mod { … }` 안에 쓴 `#![cfg(unix)]`(파일 전체를 게이트로 읽어요), `cfg_attr`로 쓴 게이트, `os::fd`·`os::linux` 같은 다른 OS 전용 경로(`os::unix` 만 찾아요). 반대로 모르는 이름의 게이트(`target_vendor` 등)와 `#[path]`로 끌려온 게이트된 `mod`는 게이트로 읽지 않아 오탐으로 보고돼요.
- 폴더가 옮겨지면 시험이 빈 목록으로 통과하지 않게 파일 수 하한(100 초과)을 두고 100/101 경계를 시험해요.
- 이 가지만 합치면 `tests/hooks_cli.rs`가 아직 고쳐지지 않아 새 시험이 실패해요 → 묶음에는 agent-3의 게이트 커밋이 함께 들어가야 해요(리더 결정 (a)).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 컴파일러가 아니라 글자 읽기로 스캔한다 | PR 의 CI 는 Linux·macOS 라 Windows 컴파일 결과를 거기서 읽을 수 없고, 이 사실(`os::unix` 를 가진 줄이 Windows 에서 꺼졌는가)은 글자로 충분히 읽혀요. 로컬 Windows 타입 검사는 가짜 C 도구가 있어야 해서(Plan 3) 매 PR 의 검사로 둘 수 없어요 | 시험 파일을 지우면 돼요 |
| D-2 | `cfg` 값을 참·거짓·모름 셋으로 읽고 모름은 게이트로 보지 않는다(`not` 안에서도) | 모르는 이름의 게이트를 믿으면 Windows 에서 컴파일되는 줄을 지나쳐요(감사 지적: `not(test)` 가 게이트로 읽히던 누락). 모르는 이름은 보고되는 쪽이 안전하고, 작성자는 `unix`·`target_os`·`target_family` 로 게이트를 쓰면 돼요 | `cfg_on_windows` 의 `None` 갈래를 바꾸면 돼요 |
| D-3 | `tests/hooks_cli.rs` 게이트는 이 SPEC에서 쓰지 않는다 | C-86 예약, 리더 결정 (a) 2026-11-19 | agent-3의 커밋을 되돌리면 돼요 |

## Result

- AC-1: met — RED→GREEN. 고치기 전(tip 7d744a28f3 위 7bf56ed59f): `cargo test --locked --offline --test windows_unix_gate_scan no_file_under_tests` → `FAILED`, 잡힌 줄은 `tests/hooks_cli.rs:1391` 하나(다른 파일은 안 걸려요: `tests/common/go_project_real_run.rs:72,366` 은 포함하는 파일이 모두 `#![cfg(unix)]` 라 안전으로 읽혀요). 고친 뒤(agent-3 의 커밋 fe40fedfa6, 결정 (a): `green_lock_bash_edges_of_folders_globs_and_moves` 의 symlink 문과 그 아래 for 루프를 `#[cfg(unix)] { … }` 블록으로 감쌈, 13 추가·9 삭제, `tests/hooks_cli.rs` 하나): 내 작업 트리에서 감사 수리까지 합친 head 04eb226cc9: `--test windows_unix_gate_scan` `11 passed; 0 failed`(실제 저장소를 읽는 `no_file_under_tests…` 포함), `--test hooks_cli` `37 passed`, `--test home_isolation_scan` `18 passed`, `cargo clippy --locked --offline --all-targets -- -D warnings` `exit=0`, `cargo fmt --check` `exit=0`. **Windows 타입 검사**: 같은 head 에서 `cargo check --locked --offline --tests --target x86_64-pc-windows-msvc`(`CC_x86_64_pc_windows_msvc`·`AR_x86_64_pc_windows_msvc` 를 저장소 밖의 가짜 C 컴파일러·ar 스크립트로 줘서 sqlite 의 C 컴파일을 통과시킴; 링크·실행은 안 함) → `Finished … in 46.78s`, `check-exit=0`, 모든 시험 타깃에 오류 0. 대조: 문제의 문장만 담은 파일을 같은 타깃으로 `rustc` 에 주면 게이트 없을 때 `error[E0433]: failed to resolve: could not find unix in os`(CI 와 같은 오류), `#[cfg(unix)] { … }` 안에 두면 `rustc-exit=0`; 감사자가 전체 `cargo check --test hooks_cli` 로 한 대조도 고치기 전 파일에서 같은 오류(`tests/hooks_cli.rs:1391`)를 냈어요. 가짜 도구가 필요해서 매 PR 의 검사로 둘 수는 없고(D-1), 푸시 run 의 실제 `windows-build`(링크 포함)는 병합 뒤 그 job 이 확인해요(`not_run`).
- AC-2: met — `cargo test --locked --offline --test windows_unix_gate_scan -- --skip no_file_under_tests` → 감사 수리 뒤 `test result: ok. 10 passed`(수리 전 7): 게이트 8꼴(`unix`·`not(windows)`·`target_os` macos/linux·`target_family = "unix"`·`all(unix, …)` 둘·공백 낀 `cfg( unix )`)과 비게이트 8꼴(`any(unix, windows)`·`not(unix)`·`windows`·`test`·`target_os`/`target_family = "windows"`·`any(unix, test)`·`#[test]`), 안쪽 `#![cfg]`, 주석·문자열·원시 문자열·수명 표기, 게이트가 끝나는 자리(블록 안 게이트, 문장 `;`, 괄호 속 `;`, 다음 모듈, 겹친 블록)를 못 박아요. 처음 돌렸을 때 `cfg(target_os = "windows")` 상자가 실패해서 스캔의 진짜 버그를 잡았어요: 문자열을 지우는 단계가 속성 값까지 지웠어요 → 속성은 원문에서 읽도록 고쳤어요.
- AC-3: met — 같은 시험 `a_helper_under_common_is_read_through_its_includers` 가 통과하고, 변이 M4(`all` → `any`)를 잡아요.
- AC-4: met — `cargo test --locked --offline --test home_isolation_scan` 18 passed; `python3 -m unittest tests.ci_scope_test` `Ran 40 tests … OK`.
- AC-5: met — 감사(57ebd2d533, PASS)의 비차단을 수리한 04eb226cc9: `a_name_the_scan_cannot_decide_never_gates`(`not(test)`·`not(feature = "x")`·`target_vendor`·`any(feature, unix)`·`not(debug_assertions)`·`all(feature, not(test))` 는 게이트 아님; `all(unix, feature)`·`not(any(windows, feature))` 는 게이트), `the_names_of_a_brace_group_are_read_one_by_one`(`os::{raw, unix}`·`os::{self, unix::fs::…}`·줄바꿈 낀 묶음은 보고; `unixish`·`other::{a, unix}`·`os::{raw, fd}`·`a::os::{b, unix_x}` 는 보고 안 함), `the_file_count_floor_has_a_boundary`(0·100 은 거부, 101 은 통과) 가 모두 통과해요. 이 셋은 위 `10 passed` 에 들어 있어요.

변이(스캔 코드를 한 곳씩 틀고 `--skip no_file_under_tests` 로 상자 시험만 돌림, 변이마다 `git checkout --` 뒤 작업 트리 clean): M1 `"unix" => false` → `true` KILLED(5개 실패); M2 `;` 가 괄호 안에서도 게이트를 끝냄 KILLED(`a_gate_covers_exactly_the_item_it_stands_on`); M3 `not(...)` 뒤집지 않음 KILLED(2개); M4 includers `all` → `any` KILLED(`a_helper_under_common_…`); M5 중첩 블록이 바깥 게이트를 잃음 → 처음에는 SURVIVED(시험이 중첩 블록을 안 봤어요) → 중첩 블록 상자를 더한 뒤 KILLED. 5/5 KILLED. 변이 스크립트의 판정 글자는 래퍼(`run.sh`)가 마지막에 `echo exit=$?` 를 해서 늘 0 이라 믿지 않고, 실패한 시험 이름과 `test result` 줄로 판정했어요. 감사 수리(04eb226cc9) 뒤 다시: 기준선(변이 없음) `11 passed`; N1 `not(...)` 의 모름을 다시 참으로 읽음(옛 동작) KILLED(`a_name_the_scan_cannot_decide_never_gates`); N2 모름 부분이 `all`·`any` 를 모름으로 만드는 갈래 삭제 KILLED(2개: `a_name_the_scan_cannot_decide_never_gates`, `a_cfg_that_keeps_the_item_on_windows_does_not_gate_it`); N3 brace 묶음의 첫 이름만 읽음 KILLED(`the_names_of_a_brace_group_are_read_one_by_one`); N4 중괄호 속 쉼표를 세지 않음 KILLED(같은 시험, 중첩 `other::{a, unix}` 상자); N5 하한 `>` → `>=` KILLED(`the_file_count_floor_has_a_boundary`). 5/5 KILLED, 변이마다 `git checkout --` 뒤 작업 트리 clean. N4 는 중첩 상자 `other::{a, unix}`(쉼표 분리가 결과를 바꾸는 꼴)로 돌렸어요.

검사: `cargo fmt --check` 출력 없음; `cargo clippy --locked --offline --all-targets -- -D warnings` `exit=0`(처음에는 `slice::from_ref` 지적 하나 → 고침); `cargo test --locked --offline --lib` 1452 passed, 1 failed(`decide::outbound::tests::a_64_kib_field_of_a_repeated_key_prefix_is_screened_quickly`, 2초 벽시계 시험, 그때 기계 부하 77): 같은 시험을 따로 다시 돌리면 `1 passed … 2.87s` 였어요(`src/` 는 건드리지 않았어요; 시간 제한은 올리지 않았어요). 수리 head 04eb226cc9 에서는 lib 전체를 다시 돌리지 않았어요(시험 파일 하나만 바뀌어 `not_run`). 로컬 Windows 확인 `cargo check --tests --target x86_64-pc-windows-msvc` 는 처음에는 `libsqlite3-sys` 의 C 컴파일(CLT 헤더에 `stdlib.h` 없음)에서 멈췄고, 가짜 C 도구로 통과시켜 AC-1 에 적은 대로 `exit 0` 을 읽었어요.

S4 가지: `origin/spec/20261119-enforce-merge-confirm-r2` 의 `tests/hooks_cli.rs:1917` 에 같은 꼴(cfg 없는 `std::os::unix::fs::symlink`)이 있어요. 이 SPEC 에서 고치지 않고 카드 노트로 보고했어요(note 2761).
