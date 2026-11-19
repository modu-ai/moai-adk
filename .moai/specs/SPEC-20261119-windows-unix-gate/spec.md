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
- 컴파일러 수준의 `cfg` 해석: 이 스캔은 글자 읽기이고, 놓치는 쪽(누락)은 있어도 잘못 잡는 쪽(오탐)은 없게 둬요.

## Requirements

- REQ-1: The scan test shall fail when a `.rs` file under `tests/` names `os::unix` outside an item, block or statement that a `cfg` attribute removes on Windows, and it shall name each such place as `path:line`.
- REQ-2: The scan test shall treat `#[cfg(unix)]`, `#[cfg(not(windows))]`, `#[cfg(target_os = "macos")]`, `#[cfg(target_family = "unix")]` and an `all(...)` that contains one of them as gates.
- REQ-3: The scan test shall treat `#[cfg(any(unix, windows))]`, `#[cfg(not(unix))]`, `#[cfg(windows)]`, `#[cfg(test)]` and `#[test]` as no gate.
- REQ-4: The scan test shall treat an inner `#![cfg(...)]` that Windows turns off as a gate for the whole file.
- REQ-5: The scan test shall skip comments, string literals, raw string literals and character literals when it looks for `os::unix`.
- REQ-6: The scan test shall judge a helper file under `tests/common/` (not `mod.rs`) as gated only when at least one file includes it and every including file is gated as a whole file.
- REQ-7 (REQ-1): IF the scan finds no `.rs` file under `tests/` (the folder moved or the walk failed), THEN the scan test shall fail with the file count it saw instead of passing on an empty list.
- REQ-8: The scan test shall end a gate at the end of the item it stands on: a `;` outside brackets or the close of the next block.

## Acceptance

- AC-1 (REQ-1, REQ-7): 이 가지에서 스캔 시험이 `tests/hooks_cli.rs:1391`을 `path:line`으로 대며 실패하고(RED), 그 줄이 `#[cfg(unix)]` 아래로 가면 통과해요. 파일 수가 100 이하면 개수를 대며 실패해요.
  verify: `cargo test --locked --offline --test windows_unix_gate_scan no_file_under_tests`
- AC-2 (REQ-2, REQ-3, REQ-4, REQ-5, REQ-8): 상자(fixture) 시험이 게이트가 되는 꼴과 안 되는 꼴, 안쪽 `#![cfg]`, 주석·문자열·수명 표기, 게이트가 끝나는 자리를 하나씩 못 박아요.
  verify: `cargo test --locked --offline --test windows_unix_gate_scan -- --skip no_file_under_tests`
- AC-3 (REQ-6): `tests/common/` 도우미는 포함하는 파일이 모두 파일 단위로 게이트일 때만 안전하다고 읽혀요(`mod.rs`는 예외).
  verify: `cargo test --locked --offline --test windows_unix_gate_scan a_helper_under_common`
- AC-4: 이웃 검사.
  verify: `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test ci_scope_test`

## Scope

- write: `.moai/specs/SPEC-20261119-windows-unix-gate/spec.md`, `tests/windows_unix_gate_scan.rs`
- forbidden: 실제 HOME 쓰기, `tests/hooks_cli.rs`(C-86 예약; agent-3이 게이트 커밋을 올려요), `src/**`, `.github/workflows/**`

## Plan

1. SPEC을 쓰고 시험을 먼저 더해요(RED): tip에서 스캔 시험이 `tests/hooks_cli.rs:1391`로 실패해야 해요.
2. 상자 시험을 더하고, 스캔의 한 곳씩을 바꾸는 변이가 시험에 잡히는지 읽어요.
3. 로컬 Windows 확인(`cargo check --tests --target x86_64-pc-windows-msvc`)은 시도해요; 안 되면 `not_run`으로 적고 CI로 보내요.
4. lib·fmt·clippy `--all-targets`를 돌리고 push해요. 게이트 커밋은 agent-3이 위에 올려요.

## Risks

- 글자 읽기라서 `cfg`가 붙은 속성을 `cfg_attr`로 우회하면 못 봐요 → 이 저장소 시험은 `cfg_attr`로 `os::unix`를 게이트하지 않아요(스캔이 `cfg_attr`을 게이트로 읽지 않으니 오탐 없음, 누락만 가능).
- 새 시험이 `tests/` 파일 수를 100 아래로 만들 일은 없지만, 폴더가 옮겨지면 시험이 빈 목록으로 통과하지 않게 개수 하한을 둬요.
- 이 가지만 합치면 `tests/hooks_cli.rs`가 아직 고쳐지지 않아 새 시험이 실패해요 → 묶음에는 agent-3의 게이트 커밋이 함께 들어가야 해요(리더 결정 (a)).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 컴파일러가 아니라 글자 읽기로 스캔한다 | Linux·macOS에서 Windows 컴파일 결과를 알 방법이 없고(로컬 Windows `cargo check`는 C 도구 때문에 막혀요) 이 사실은 글자로 충분히 읽혀요 | 시험 파일을 지우면 돼요 |
| D-2 | 알 수 없는 `cfg` 이름은 Windows에서 참으로 쳐서 게이트로 보지 않는다 | 놓치는 쪽이 잘못 잡는 쪽보다 낫고, 새 시험이 이유 없이 빨개지지 않아요 | 해당 분기를 바꾸면 돼요 |
| D-3 | `tests/hooks_cli.rs` 게이트는 이 SPEC에서 쓰지 않는다 | C-86 예약, 리더 결정 (a) 2026-11-19 | agent-3의 커밋을 되돌리면 돼요 |

## Result

(구현 뒤에 AC마다 한 줄씩 적어요.)
