---
id: SPEC-20261123-test-home-isolation
title: "테스트 프로세스는 빈 임시 HOME 아래에서 돌고, 실제 HOME 을 읽거나 바꾸면 그 테스트가 실패해요"
status: approved
revision: 1
created: 2026-11-23
approval: {by: "leader (사용자 위임 2026-11-23T13:57Z)", at: "2026-11-23", basis: "리더 결정 노트 7562 (2026-11-23T14:15Z, 사용자 위임, 리더 판단): 결정 1 SPEC 판 1 을 (a) nextest wrapper 로 승인. 조건 4개(원문 읽기, 실험 위험을 R 에, fail-closed 시험, enforce 는 개수를 본 뒤 노트)는 이 판 본문에 반영했어요. 사용자가 직접 고른 승인이 아니에요", digest: "sha256:6b180fe5c23ae9de94987e5cbc35afbc06862ab76986b8313e938965b360ac8a"}
---

# 테스트 프로세스는 빈 임시 HOME 아래에서 돌고, 실제 HOME 을 읽거나 바꾸면 그 테스트가 실패해요

## Interview

- none: 리더 배차가 요구사항을 전달했어요 — 카드의 본문과 감사 노트 둘(실제 HOME 을 읽은 길)이 요구를 정했고, 방식 선택은 리더 결정 노트에 따랐으며, 별도 사용자 인터뷰는 없었어요.

## Intent

같은 원인이 두 번 나왔어요. 한 작업의 감사에서 lib 의 `session_check` 테스트가 실제 `$HOME/.moai` 를 읽었어요. 1차는 `language.yaml` 이었고, 고친 뒤에도 `identity::identity`(`src/identity.rs`, env `HOME`)에서 `home_state::home_go_markers` 로 `user-assets.json` 을 읽었어요. 감사자는 가짜 HOME 의 `atime`(마지막으로 읽은 시각)을 재서 네 테스트 모두에서 재현했어요. nextest 는 테스트 프로세스에 호출자의 HOME, 곧 실제 HOME 을 그대로 줘요. 그래서 "테스트가 실제 HOME 을 읽음"은 리뷰 규칙으로 사람이 눈으로만 잡았어요. `src` 에서 env `HOME` 을 읽는 파일이 41개라 어느 테스트든 닿을 수 있어요(`rg -c 'var_os\("HOME"\)|var\("HOME"\)|home_dir\(' src`).

이 SPEC 은 nextest 의 wrapper script(테스트 프로세스마다 앞뒤로 도는 감싸기 스크립트)로 모든 테스트 프로세스에 빈 임시 HOME 을 주고, 그 안의 미끼 파일을 읽거나 폴더가 바뀌면 그 테스트를 실패시키는 장치(`enforce` 모드)를 만들어요. 쓰는 사람은 `moai check affected` 와 리더의 전체 `cargo nextest run` 을 돌리는 모든 세션이에요. 장치가 막는 것은 읽는 일이 아니라, 실제 사용자 폴더가 테스트 결과에 섞이는 일이에요. 새로 막는 검사이므로 처음 기본값은 `observe`(알리기만, 실패시키지 않음)이고, `enforce` 로 바꾸는 일은 개수를 본 뒤 리더의 기록(D-2)으로 해요.

## Non-goals

- Windows 호스트. wrapper 가 `sh` 를 쓰므로 unix 호스트에만 걸고, Windows 결과는 말하지 않아요.
- 없는 경로를 `stat` 하거나 `exists()` 로 묻는 일. 파일의 `atime` 이 바뀌지 않아서 잡히지 않아요.
- 운영체제 수준의 파일 접근 추적(`fs_usage`, `inotify`).
- nextest 가 없을 때의 `cargo test` 대체 실행. 그 실행에는 장치가 없어요.
- 테스트가 자식 프로세스에 스스로 임시 HOME 을 주는 방식(`TempHome`)을 바꾸는 일. 그대로 두고, 이 장치는 그 아래에서 더 바깥을 막아요.

## Requirements

<!-- This SPEC adds a check that blocks, so it starts in observe: the Result records the would-block count and the false positives; the switch to enforce is recorded as D-2. -->
- REQ-1: Where the host is unix, the nextest configuration shall run every test process through `tools/test_home_guard.sh`, except the tests named in its allowlist.
- REQ-2: When the wrapper starts a test, it shall create a fresh folder under `TMPDIR` named `moai-test-home.*` that holds decoy files at `.moai/user-assets.json`, `.moai/language.yaml` and `.claude/settings.json` whose access time is older than their modification time, and run the test with `HOME` and `USERPROFILE` set to that folder.
- REQ-3 (REQ-2): IF the wrapper cannot create the folder or its decoys, THEN the wrapper shall exit 70 without running the test.
- REQ-4: When the test process ends and any entry under its HOME was read, added, changed or removed, the wrapper shall print a line that starts with `HOME-GUARD:` and names the test's arguments.
- REQ-5 (REQ-4): IF `MOAI_HOME_GUARD` is `enforce` and the test exited 0, THEN the wrapper shall exit 1.
- REQ-6: In every other case the wrapper shall keep the test's own exit code, and read an unset `MOAI_HOME_GUARD` as `observe`.
- REQ-7: The wrapper shall remove only the folder it made, and only when its name starts with `moai-test-home.` under `TMPDIR`.
- REQ-8: The wrapper shall keep `CARGO_HOME` and `RUSTUP_HOME` at the values they had, or at their default under the caller's HOME, so cargo and rustup children of a test behave as before.
- REQ-9: When `MOAI_HOME_GUARD_LOG` names a file, the wrapper shall append each `HOME-GUARD:` line to it as well. · failure: none — a log write that fails is ignored and the test result is unchanged.
- REQ-10: A repository test shall fail on an allowlist entry of `.config/nextest.toml` that has no `# reason:` line or names no existing test or binary.
- REQ-11 (REQ-1): IF `tools/test_home_guard.sh` is missing or cannot run the test binary, THEN a nextest run shall fail that test instead of passing it.
- REQ-12: The nextest configuration shall name a `nextest-version` that has the wrapper-scripts feature and enable it with `experimental = ["wrapper-scripts"]`.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4, REQ-5, REQ-6): in a scratch crate that uses the repository's `.config/nextest.toml` and wrapper with `MOAI_HOME_GUARD=enforce`, a test that reads `$HOME/.moai/user-assets.json` fails with a `HOME-GUARD:` line, a test that writes a file in its HOME fails the same way, a failing test that also reads keeps its own exit code, and a test that touches nothing passes; with the default mode the reading test passes and the line is printed.
  verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-2 (REQ-3, REQ-7, REQ-8, REQ-9): with `TMPDIR` pointing at a folder that cannot be written the wrapper exits 70 and runs no test; a link or a folder not named `moai-test-home.*` is never removed; a child `cargo --version` of a test still finds the toolchain; `MOAI_HOME_GUARD_LOG` receives the line.
  verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-3 (REQ-10, REQ-12): an allowlist entry without a `# reason:` line fails, an entry that names no existing test or binary fails, the allowlist in the repository passes, and the configuration enables `wrapper-scripts` and names a `nextest-version`.
  verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-4 (REQ-11): in the scratch crate with `tools/test_home_guard.sh` removed, and again with a copy of it that does not run its arguments, the planted test does not pass.
  verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-5 (REQ-1): the whole `cargo nextest run --locked --offline --no-fail-fast` under the device passes, every test that touched its HOME is fixed or on the allowlist, and `moai check affected` runs under the same device.
  verify: `cargo nextest run --locked --offline --no-fail-fast`, and the count quoted in the Result

## Scope

- write: `.moai/specs/SPEC-20261123-test-home-isolation/**`, `.config/nextest.toml`, `tools/test_home_guard.sh`, `tests/test_home_guard.rs`, and the test files the count names (listed in the Result before they change; the leader widens the card for them)
- read-only: `src/check_affected.rs` (held by another card; the device needs no change there because `check affected` runs `cargo nextest run` when nextest exists), every other `src/` and `tests/` file
- forbidden: the real HOME folder and the `.claude` and `.moai` folders in it; any shell `HOME=` assignment in a session

## Plan

1. RED: write `tests/test_home_guard.rs` first. It builds a scratch crate in a temporary folder, copies the repository's `.config/nextest.toml` and `tools/test_home_guard.sh` into it, plants the reading and the writing test, and expects `HOME-GUARD:`. It fails because the script and the nextest rule do not exist.
2. GREEN: add `tools/test_home_guard.sh` and the rule and allowlist to `.config/nextest.toml` (`experimental = ["wrapper-scripts"]`, `nextest-version`, a `cfg(unix)` platform filter). Keep the number of child processes per test small and measure the added time per test.
3. Observe: before the run put one line on the screen so it does not overlap the leader's full run, then run the whole suite once with `MOAI_HOME_GUARD=observe` and `MOAI_HOME_GUARD_LOG` in one heavy slot (`--max 4`), record the would-block count and each name in the Result, and sort each as a real read, a write, or a false positive.
4. Fix each real read by giving the test its own temporary HOME, or add it to the allowlist with its reason; ask the leader to widen the card for each test file first.
5. After the leader records D-2, set the default to `enforce`, run the whole suite again with no failure, then one `moai check affected` at READY.

## Risks

- The `wrapper-scripts` feature of nextest is experimental in its own documentation ("This is an advanced feature, and it can cause your tests to silently stop working if used incorrectly") and can change in a later nextest. The configuration pins `nextest-version` and enables the feature by name, so an older nextest stops with an error instead of running unguarded. A newer nextest that changes the feature is not covered and shows up as failed tests first. The Result names the version run (0.9.152 here).
- A wrapper that does not run its test makes nextest report a pass: the wrapper returns the test's own exit code and AC-4 plants a missing and a non-running wrapper to show the test does not pass.
- Cost per test: the wrapper adds child processes to each of about 7300 test processes (about 56 ms added wall time in a scratch loop of 40 runs, mostly waiting on process start). Plan step 2 measures it on the repository and the Result quotes the run with and without the wrapper.
- A decoy changes a test that expects an empty HOME: the observation run in step 3 finds these; the decoy list stays this short on purpose.
- `atime` is not kept up on a file system mounted without it, and Linux and Windows were not measured here: AC-1 reads a decoy the same way on every host that runs the suite and fails there when the read is not seen, so a host that cannot see reads is not reported as guarded.
- The wrapper replaces `HOME` for tools a test starts: git, cargo and rustup paths are the first ones to break, which the observation run shows by name.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A wrapper script per test process, not a setup script and not cargo `[env]`. A setup script runs once and exports one HOME to every test through `$NEXTEST_ENV` (measured in a scratch crate: one run of the script, the same HOME in the lib and the integration binary, the caller's HOME overridden), so a read cannot be told apart per test, and nextest has no cleanup step after it. cargo `[env]` is static text and cannot make a temporary folder. | Documents read in full from the nextest repository at tag `cargo-nextest-0.9.152` on 2026-11-23T14:15:49Z: `site/src/docs/configuration/wrapper-scripts.md` and `setup-scripts.md` (raw.githubusercontent.com/nextest-rs/nextest/cargo-nextest-0.9.152/…): wrapper scripts since 0.9.98, `relative-to` since 0.9.99, `platform = { host = "cfg(unix)" }` is the platform form, a wrapper gets the test binary as its first argument; cargo `[env]` text read through a WebFetch digest on 2026-11-23 (https://doc.rust-lang.org/cargo/reference/config.html#env, digest only, so unverified as raw text); plus the scratch-crate runs of this lane (nextest 0.9.152, macOS arm64): the wrapper failed only the planted reading test of four. Unverified: Linux and Windows behavior, whether `$NEXTEST_ENV` of a later nextest still overrides the caller's HOME | Delete the rule and the script |
| D-2 | The device starts in observe (the script default), and the leader records the switch to `enforce` after reading the would-block count and the false positives in the Result. | The observe-first rule of `internal/sdd`; the leader's note on the card: "enforce 를 기본으로 바꾸는 것은 개수를 본 뒤 다시 노트로" | Set the default to observe in the script |
| D-3 | A read is seen through the `atime` of decoy files and folders, a change through the modification time of the entries against a reference file made after the decoys. | Measured on APFS in a temporary folder: a read moves a decoy's `atime` from 2000-01-01 to now | Replace the decoys with a file-access tracer |
| D-4 | unix only; a Windows host runs its tests as before. | The wrapper needs `sh`; no Windows run exists here | Add a Windows wrapper |
