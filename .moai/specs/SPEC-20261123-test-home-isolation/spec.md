---
id: SPEC-20261123-test-home-isolation
title: "테스트 프로세스는 빈 임시 HOME 아래에서 돌고, 실제 HOME 을 읽거나 바꾸면 그 테스트가 실패해요"
status: draft
revision: 1
created: 2026-11-23
---

# 테스트 프로세스는 빈 임시 HOME 아래에서 돌고, 실제 HOME 을 읽거나 바꾸면 그 테스트가 실패해요

## Intent

같은 원인이 두 번 나왔어요. 한 작업의 감사에서 lib 의 `session_check` 테스트가 실제 `$HOME/.moai` 를 읽었어요. 1차는 `language.yaml` 이었고, 고친 뒤에도 `identity::identity`(`src/identity.rs`, env `HOME`)에서 `home_state::home_go_markers` 로 `user-assets.json` 을 읽었어요. 감사자는 가짜 HOME 의 `atime`(마지막으로 읽은 시각)을 재서 네 테스트 모두에서 재현했어요. nextest 는 테스트 프로세스에 호출자의 HOME, 곧 실제 HOME 을 그대로 줘요. 그래서 "테스트가 실제 HOME 을 읽음"은 리뷰 규칙으로 사람이 눈으로만 잡았어요. `src` 에서 env `HOME` 을 읽는 파일이 41개라 어느 테스트든 닿을 수 있어요(`rg -c 'var_os\("HOME"\)|var\("HOME"\)|home_dir\(' src`).

이 SPEC 은 nextest 의 wrapper script(테스트 프로세스마다 앞뒤로 도는 감싸기 스크립트)로 모든 테스트 프로세스에 빈 임시 HOME 을 주고, 그 안의 미끼 파일을 읽거나 폴더를 바꾸면 그 테스트를 실패시켜요. 쓰는 사람은 `moai check affected` 와 리더의 전체 `cargo nextest run` 을 돌리는 모든 세션이에요. 장치가 막는 것은 읽는 일이 아니라, 실제 사용자 폴더가 테스트 결과에 섞이는 일이에요.

## Non-goals

- Windows 호스트. wrapper 가 `sh` 를 쓰므로 unix 호스트에만 걸고, Windows 결과는 말하지 않아요.
- 없는 경로를 `stat` 하거나 `exists()` 로 묻는 일. 파일의 `atime` 이 바뀌지 않아서 잡히지 않아요.
- 운영체제 수준의 파일 접근 추적(`fs_usage`, `inotify`).
- nextest 가 없을 때의 `cargo test` 대체 실행. 그 실행에는 장치가 없어요.
- 테스트가 자식 프로세스에 스스로 임시 HOME 을 주는 방식(`TempHome`)을 바꾸는 일. 그대로 두고, 이 장치는 그 아래에서 더 바깥을 막아요.

## Requirements

<!-- This SPEC adds a check that blocks, so it starts in observe: the Result records the would-block count and the false positives; the switch to enforce is recorded as D-2. -->
- REQ-1: Where the host is unix, the nextest configuration shall run every test process through `tools/test_home_guard.sh`, except the tests named in its allowlist.
- REQ-2: When the wrapper starts a test, it shall create a fresh folder under `TMPDIR` named `moai-test-home.*` that holds decoy files at `.moai/user-assets.json`, `.moai/language.yaml` and `.claude/settings.json` whose access time is older than their modification time.
- REQ-3 (REQ-2): IF the wrapper cannot create the folder or its decoys, THEN the wrapper shall exit 70 without running the test.
- REQ-4: The wrapper shall run the test with `HOME` and `USERPROFILE` set to the folder it made.
- REQ-5: When the test process ends and any entry under its HOME was read, added, changed or removed, the wrapper shall print a line that starts with `HOME-GUARD:` and exit 1 when the test itself exited 0.
- REQ-6 (REQ-5): IF the test process exits non-zero, THEN the wrapper shall keep that exit code and still print the `HOME-GUARD:` line.
- REQ-7: The wrapper shall remove only the folder it made, and only when its name starts with `moai-test-home.` under `TMPDIR`.
- REQ-8: The wrapper shall keep `CARGO_HOME` and `RUSTUP_HOME` at the values they had, or at their default under the caller's HOME, so cargo and rustup children of a test behave as before.
- REQ-9: When `MOAI_HOME_GUARD` is `observe`, the wrapper shall print the `HOME-GUARD:` line and keep the test's own exit code. · failure: none — the mode only switches the exit code off, it starts nothing that can fail.
- REQ-10: Every test that needs the real HOME shall be named in the allowlist of `.config/nextest.toml` with one `# reason:` line.
- REQ-11: A repository test shall fail on an allowlist entry that has no reason line or names nothing.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4, REQ-5, REQ-6): in a scratch crate that uses the repository's `.config/nextest.toml` and wrapper, a test that reads `$HOME/.moai/user-assets.json` fails with a `HOME-GUARD:` line, a test that writes a file in its HOME fails the same way, a failing test that also reads keeps its own exit code, and a test that touches nothing passes.
  verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-2 (REQ-3, REQ-7, REQ-8, REQ-9): with `TMPDIR` pointing at a folder that cannot be written the wrapper exits 70 and runs no test; a link or a folder not named `moai-test-home.*` is never removed; a child `cargo --version` of a test still finds the toolchain; `MOAI_HOME_GUARD=observe` prints the line and exits 0.
  verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-3 (REQ-10, REQ-11): an allowlist entry without a `# reason:` line fails, an entry that names no existing test or binary fails, and the allowlist in the repository passes.
  verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-4 (REQ-1): the whole `cargo nextest run --locked --offline --no-fail-fast` under the device passes, every test that failed under it is fixed or on the allowlist, and `moai check affected` runs under the same device.
  verify: `cargo nextest run --locked --offline --no-fail-fast`, and the count quoted in the Result

## Scope

- write: `.moai/specs/SPEC-20261123-test-home-isolation/**`, `.config/nextest.toml`, `tools/test_home_guard.sh`, `tests/test_home_guard.rs`, and the test files the count names (listed in the Result before they change)
- read-only: `src/check_affected.rs` (held by another card; the device needs no change there because `check affected` runs `cargo nextest run` when nextest exists), every other `src/` and `tests/` file
- forbidden: the real HOME folder and the `.claude` and `.moai` folders in it; any shell `HOME=` assignment in a session

## Plan

1. RED: write `tests/test_home_guard.rs` first. It builds a scratch crate in a temporary folder, copies the repository's `.config/nextest.toml` and `tools/test_home_guard.sh` into it, plants the reading and the writing test, and expects `HOME-GUARD:`. It fails because the script and the nextest rule do not exist.
2. GREEN: add `tools/test_home_guard.sh` and the rule and allowlist to `.config/nextest.toml` (`experimental = ["wrapper-scripts"]`, `nextest-version`, a `cfg(unix)` platform filter). Keep the number of child processes per test small.
3. Observe: run the whole suite once with `MOAI_HOME_GUARD=observe` in one heavy slot, record the would-block count and each name in the Result, and sort each as a real read, a write, or a false positive.
4. Fix each real read by giving the test its own temporary HOME, or add it to the allowlist with its reason; add the file to the scope request first.
5. Enforce (D-2), run the whole suite again with no failure, then one `moai check affected` at READY.

## Risks

- Cost per test: the wrapper adds child processes to each of about 7300 test processes. Plan step 2 measures the added time against the run without the wrapper and the Result quotes both numbers.
- A decoy changes a test that expects an empty HOME: the observation run in step 3 finds these; the decoy list stays this short on purpose.
- `atime` is not kept up on a file system mounted without it, and Linux and Windows were not measured here: AC-1 reads a decoy the same way on every host that runs the suite and fails there when the read is not seen, so a host that cannot see reads is not reported as guarded.
- The `wrapper-scripts` feature of nextest is marked experimental in its documentation: the configuration pins `nextest-version` and the Result names the version run (0.9.152 here).
- The wrapper replaces `HOME` for tools a test starts: git, cargo and rustup paths are the first ones to break, which the observation run shows by name.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A wrapper script per test process, not a setup script and not cargo `[env]`. A setup script runs once and exports one HOME to every test through `$NEXTEST_ENV` (measured in a scratch crate: one run of the script, the same HOME in the lib and the integration binary, the caller's HOME overridden), so a read cannot be told apart per test, and nextest has no cleanup step after it. cargo `[env]` is static text and cannot make a temporary folder. | Documents read 2026-11-23 through WebFetch digests, so the quotes are unverified as raw text: https://nexte.st/docs/configuration/setup-scripts/, https://nexte.st/docs/configuration/wrapper-scripts/, https://nexte.st/docs/configuration/env-vars/, https://doc.rust-lang.org/cargo/reference/config.html#env; plus the scratch-crate runs in this lane (nextest 0.9.152, macOS arm64): the wrapper failed only the planted reading test of four | Delete the rule and the script |
| D-2 | The device starts in observe, then enforces by default once the observation run of Plan step 3 is recorded and every would-block is fixed or listed. `MOAI_HOME_GUARD=observe` remains the way to triage. | The observe-first rule of `internal/sdd`; the switch needs the leader's recorded note | Set the default to observe in the script |
| D-3 | A read is seen through the `atime` of decoy files and folders, a change through a before-and-after listing. | Measured on APFS in a temporary folder: a read moves a decoy's `atime` from 2000-01-01 to now | Replace the decoys with a file-access tracer |
| D-4 | unix only; a Windows host runs its tests as before. | The wrapper needs `sh`; no Windows run exists here | Add a Windows wrapper |
