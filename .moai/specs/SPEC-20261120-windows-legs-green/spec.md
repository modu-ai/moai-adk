---
id: SPEC-20261120-windows-legs-green
title: "The Windows legs of the release pull request run: the smoke harness skips where it cannot run, and the crate compiles for Windows with clippy warnings denied"
status: draft
revision: 1
created: 2026-11-20
---

# The Windows legs of the release pull request run: the smoke harness skips where it cannot run, and the crate compiles for Windows with clippy warnings denied

## Intent

정식 릴리즈 PR 은 Linux·macOS·Windows 세 곳에서 CI 를 돌려요(사용자 결정). 지금 코드에는 Windows 다리를 빨갛게 만드는 결함이 둘 있어요(독립 감사가 찾았어요).

1. 연기 시험 도구(`tools/smoke/smoke.py`)가 맨 위에서 `pwd` 와 `pty` 를 가져와요. 둘 다 POSIX 에만 있고(`pty` 는 `termios` 를 끌어와요), CI 의 "Smoke harness" 단계는 전체 범위 실행이면 모든 다리에서 돌아요. Windows 에서는 가져오기에서 죽어요.
2. `cargo clippy --target x86_64-pc-windows-msvc --all-targets -- -D warnings` 가 실패해요: unix 전용 항목이 cfg 문 없이 쓰이거나(`user_assets.rs` 의 `flock`), unix 에서만 쓰는 변수·가져오기가 Windows 에서는 안 쓰여서 `unused_mut`·`unused variable`·`unused import` 가 돼요.

수리는 cfg 문으로만 해요(`allow` 로 경고를 끄지 않아요). 이 Mac 에서는 진짜 Windows 빌드가 안 되므로, 가짜 C 컴파일러·보관 도구를 쓴 **타입 검사만** 해요(링크도 실행도 아니에요); 진짜 Windows 빌드는 푸시 때의 `windows-build` 작업이 마지막 말을 해요.

## Non-goals

- 연기 시험 도구를 Windows 에서 돌게 만드는 일. 이 도구는 `/usr/bin` 과 `/bin` 의 PATH, 터미널 의사 장치, 계정 항목에 기대어요; Windows 에서는 이유 한 줄을 남기고 건너뛰어요.
- 경고를 `allow` 로 끄는 일.
- 진짜 Windows 환경에서의 실행 시험(이 Mac 에서는 못 해요).

## Requirements

- REQ-1: The smoke harness shall import `pwd` and `pty` without failing on a platform that has neither.
- REQ-2: When the smoke harness runs on a platform without `pwd` or `pty`, it shall print one line that starts `smoke: <area> skip` and gives the reason, run no step and exit with code 0. · failure: none — the line is printed before any step or file is touched, so there is nothing to fail.
- REQ-3: When the smoke harness runs on a platform that has `pwd` and `pty`, it shall print the same lines and exit with the same codes as before this change. · failure: none — the change only adds a check that is false on such a platform, so no run on it can fail because of it.
- REQ-4: The crate shall pass `cargo clippy --target x86_64-pc-windows-msvc --all-targets -- -D warnings` as a type check, with unix-only items behind `cfg` gates and no `allow` attribute added.
- REQ-5: The user-asset acquire guard shall take its exclusive lock through the standard library's `File::try_lock` on a platform that is not unix.
- REQ-6: The user-asset acquire guard shall keep the `flock` call it has on unix.

## Acceptance

- AC-1 (REQ-1, REQ-2): with `pwd` and `pty` made unimportable, the harness module loads and `main` prints one `smoke: install skip` line with a reason, runs no step and returns 0.
  verify: `python3 -m unittest tests.smoke_harness_test.PlatformTest`
- AC-2 (REQ-3): every other test of the harness still passes unchanged.
  verify: `python3 -m unittest tests.smoke_harness_test tests.ci_smoke_step_test`
- AC-3 (REQ-4, REQ-5, REQ-6): with a stand-in C compiler and archiver for the Windows target, `cargo clippy --target x86_64-pc-windows-msvc --all-targets -- -D warnings` exits 0 (a type check: nothing is linked or run).
  verify: `cargo clippy --offline --locked --target x86_64-pc-windows-msvc --all-targets -- -D warnings`
- AC-4 (REQ-4): the Unix build keeps passing the same check.
  verify: `cargo clippy --offline --locked --all-targets -- -D warnings`
- AC-5 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6): the SPEC passes the checker.
  verify: `moai spec check SPEC-20261120-windows-legs-green`

## Scope

- write: `.moai/specs/SPEC-20261120-windows-legs-green/**`, `tools/smoke/smoke.py`, `tests/smoke_harness_test.py`, and the source files the Windows clippy run names (the folder builders that are made `mut` only for unix, the unix-only test variables, the user-asset lock).
- read-only: every other file, the real HOME.
- tier 1: an independent audit of the exact head.
- not here: `src/update.rs` (a test module imports `super::*` that no test uses off unix) is held by another card; the line goes to that card or to a later short take.

## Plan

1. SPEC and the red Python test (the harness must load without `pwd`/`pty`).
2. `tools/smoke/smoke.py`: guarded imports and the skip line at the top of `main`.
3. Windows clippy baseline with the stand-in compiler tools, then one cfg-gated fix per site until the Windows run exits 0; the Unix clippy run stays 0.
4. The unnarrowed check set, then an independent audit.

## Risks

- A cfg gate that hides a Unix-only path from a Windows build must not change the Unix build: every gated edit is read against the Unix clippy and the Unix tests.
- The type check proves types only; the real Windows job on the pull request is the final word.

## Result

(written at SYNC)
