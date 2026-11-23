---
id: SPEC-20261123-test-home-isolation
title: "테스트 프로세스는 빈 임시 HOME 아래에서 돌고, 실제 HOME 을 읽거나 바꾸면 그 테스트가 실패해요"
status: approved
revision: 2
created: 2026-11-23
approval: {by: "leader (사용자 위임 2026-11-23T13:57Z)", at: "2026-11-23", basis: "리더 결정 노트 7600 (2026-11-23T15:03:16Z, 사용자 위임, 리더 판단): 노트 7598 의 선택지 A(기준선 파일)에 따라 판 2(기준선 요구, 이름이 실제 테스트인지 확인하는 테스트, 인수 기준 하나)를 승인으로 봐요. 조건(기준선은 줄기만 하고 늘지 않음, 기본 enforce, 389줄의 출처를 Result 에)은 이 판에 반영했어요. 판 1 은 노트 7562 가 승인했어요. 사용자가 직접 고른 승인이 아니에요", digest: "sha256:84b9bf41a063b339bce27fd32c78d07cc162d361dfd6edd6709d65fd2cd8f00c"}
---

# 테스트 프로세스는 빈 임시 HOME 아래에서 돌고, 실제 HOME 을 읽거나 바꾸면 그 테스트가 실패해요

## Interview

- none: 리더 배차가 요구사항을 전달했어요 — 카드의 본문과 감사 노트 둘(실제 HOME 을 읽은 길)이 요구를 정했고, 방식 선택은 리더 결정 노트에 따랐으며, 별도 사용자 인터뷰는 없었어요.

## Intent

같은 원인이 두 번 나왔어요. 한 작업의 감사에서 lib 의 `session_check` 테스트가 실제 `$HOME/.moai` 를 읽었어요. 1차는 `language.yaml` 이었고, 고친 뒤에도 `identity::identity`(`src/identity.rs`, env `HOME`)에서 `home_state::home_go_markers` 로 `user-assets.json` 을 읽었어요. 감사자는 가짜 HOME 의 `atime`(마지막으로 읽은 시각)을 재서 네 테스트 모두에서 재현했어요. nextest 는 테스트 프로세스에 호출자의 HOME, 곧 실제 HOME 을 그대로 줘요. 그래서 "테스트가 실제 HOME 을 읽음"은 리뷰 규칙으로 사람이 눈으로만 잡았어요. `src` 에서 env `HOME` 을 읽는 파일이 41개라 어느 테스트든 닿을 수 있어요(`rg -c 'var_os\("HOME"\)|var\("HOME"\)|home_dir\(' src`).

이 SPEC 은 nextest 의 wrapper script(테스트 프로세스마다 앞뒤로 도는 감싸기 스크립트)로 모든 테스트 프로세스에 빈 임시 HOME 을 주고, 그 안의 미끼 파일을 읽거나 폴더가 바뀌면 그 테스트를 실패시키는 장치(`enforce` 모드)를 만들어요. 쓰는 사람은 `moai check affected` 와 리더의 전체 `cargo nextest run` 을 돌리는 모든 세션이에요. 장치가 막는 것은 읽는 일이 아니라, 실제 사용자 폴더가 테스트 결과에 섞이는 일이에요. 새로 막는 검사라서 먼저 `observe`(알리기만, 실패시키지 않음)로 전체를 한 번 돌려 막힐 개수를 쟀고(D-5: 389개 테스트, 모두 `identity::identity` 를 거쳐 미끼 `user-assets.json` 을 읽음), 리더가 그 개수를 보고 기본을 `enforce` 로 정했어요(D-2). 389개는 뿌리 수리(identity 가 HOME 을 인자로 받음)가 따로 올 때까지 기준선 파일에 두고, 그 파일은 줄기만 하고 늘지 않아요. 기준선에 없는 테스트가 HOME 을 건드리면 실패해요.

## Non-goals

- Windows 호스트. wrapper 가 `sh` 를 쓰므로 unix 호스트에만 걸고, Windows 결과는 말하지 않아요.
- 없는 경로를 `stat` 하거나 `exists()` 로 묻는 일. 파일의 `atime` 이 바뀌지 않아서 잡히지 않아요.
- 운영체제 수준의 파일 접근 추적(`fs_usage`, `inotify`).
- nextest 가 없을 때의 `cargo test` 대체 실행. 그 실행에는 장치가 없어요.
- 테스트가 자식 프로세스에 스스로 임시 HOME 을 주는 방식(`TempHome`)을 바꾸는 일. 그대로 두고, 이 장치는 그 아래에서 더 바깥을 막아요.

## Requirements

<!-- This SPEC adds a check that blocks, so it starts in observe: the Result records the would-block count and the false positives; the switch to enforce is recorded as D-2. -->
- REQ-1: Where the host is unix, the nextest configuration shall run every test process through `tools/test_home_guard.sh`, except the tests named in its allowlist, with the `wrapper-scripts` feature enabled and a `nextest-version` that has it.
- REQ-2: When the wrapper starts a test, it shall create a fresh folder under `TMPDIR` named `moai-test-home.*` that holds decoy files at `.moai/user-assets.json`, `.moai/language.yaml` and `.claude/settings.json` whose access time is older than their modification time, and run the test with `HOME` and `USERPROFILE` set to that folder.
- REQ-3 (REQ-2): IF the wrapper cannot create the folder or its decoys, THEN the wrapper shall exit 70 without running the test.
- REQ-4: When the test process ends and any entry under its HOME was read, added, changed or removed, the wrapper shall print a line that starts with `HOME-GUARD:` and names the test's arguments.
- REQ-5 (REQ-4): IF `MOAI_HOME_GUARD` is not `observe` and the test exited 0, THEN the wrapper shall exit 1, unless REQ-12 tolerates the test.
- REQ-6: In every other case the wrapper shall keep the test's own exit code, and read an unset `MOAI_HOME_GUARD` as `enforce`.
- REQ-7: The wrapper shall remove only the folder it made, and only when its name starts with `moai-test-home.` under `TMPDIR`.
- REQ-8: The wrapper shall keep `CARGO_HOME` and `RUSTUP_HOME` at the values they had, or at their default under the caller's HOME, so cargo and rustup children of a test behave as before.
- REQ-9: When `MOAI_HOME_GUARD_LOG` names a file, the wrapper shall append each `HOME-GUARD:` line to it as well. · failure: none — a log write that fails is ignored and the test result is unchanged.
- REQ-10: A repository test shall fail on an allowlist entry of `.config/nextest.toml` that has no `# reason:` line or names no existing test or binary, and on a baseline that is longer than its recorded cap, is not sorted without repeats, or names a test that does not exist.
- REQ-11 (REQ-1): IF `tools/test_home_guard.sh` is missing or cannot run the test binary, THEN a nextest run shall fail that test instead of passing it.
- REQ-12: Where a test is listed in `tests/fixtures/test_home_guard/baseline.tsv` and touched nothing under its HOME but the decoy `user-assets.json` and the two folders above it, the wrapper shall print the line marked `tolerated` and keep the test's own exit code. · failure: none — a test that touches more, or one that is not listed, is REQ-5's case.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4, REQ-5, REQ-6): in a scratch crate that uses the repository's `.config/nextest.toml` and wrapper in its default mode (`enforce`), a test that reads `$HOME/.moai/user-assets.json` fails with a `HOME-GUARD:` line, a test that writes a file in its HOME fails the same way, a failing test that also reads keeps its own exit code, and a test that touches nothing passes; with `MOAI_HOME_GUARD=observe` the reading test passes and the line is printed.
  verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-2 (REQ-3, REQ-7, REQ-8, REQ-9): with `TMPDIR` pointing at a folder that cannot be written the wrapper exits 70 and runs no test; a link or a folder not named `moai-test-home.*` is never removed; a child `cargo --version` of a test still finds the toolchain; `MOAI_HOME_GUARD_LOG` receives the line.
  verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-3 (REQ-10, REQ-1): an allowlist entry without a `# reason:` line fails, an entry that names no existing test or binary fails, the allowlist in the repository passes, and the configuration enables `wrapper-scripts` and names a `nextest-version`.
  verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-6 (REQ-10, REQ-12): a listed test that reads only the decoy `user-assets.json` passes with a `tolerated` line, a listed test that also writes, and a test that is not listed, fail; the baseline in the repository is sorted without repeats, holds no more than 389 lines, and every line names a test that exists.
  verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-4 (REQ-11): in the scratch crate with `tools/test_home_guard.sh` removed, and again with a copy of it that does not run its arguments, the planted test does not pass.
  verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-5 (REQ-1): the whole `cargo nextest run --locked --offline --no-fail-fast` under the device passes with the device on by default, every test that touched its HOME is on the baseline, and `moai check affected` runs under the same device.
  verify: `cargo nextest run --locked --offline --no-fail-fast`, and the count quoted in the Result

## Scope

- write: `.moai/specs/SPEC-20261123-test-home-isolation/**`, `.config/nextest.toml`, `tools/test_home_guard.sh`, `tests/test_home_guard.rs`, `tests/fixtures/test_home_guard/baseline.tsv`
- read-only: `src/check_affected.rs` (held by another card; the device needs no change there because `check affected` runs `cargo nextest run` when nextest exists), every other `src/` and `tests/` file
- forbidden: the real HOME folder and the `.claude` and `.moai` folders in it; any shell `HOME=` assignment in a session

## Plan

1. RED: write `tests/test_home_guard.rs` first. It builds a scratch crate in a temporary folder, copies the repository's `.config/nextest.toml` and `tools/test_home_guard.sh` into it, plants the reading and the writing test, and expects `HOME-GUARD:`. It fails because the script and the nextest rule do not exist.
2. GREEN: add `tools/test_home_guard.sh` and the rule and allowlist to `.config/nextest.toml` (`experimental = ["wrapper-scripts"]`, `nextest-version`, a `cfg(unix)` platform filter). Keep the number of child processes per test small and measure the added time per test.
3. Observe: before the run put one line on the screen so it does not overlap the leader's full run, then run the whole suite once with `MOAI_HOME_GUARD=observe` and `MOAI_HOME_GUARD_LOG` in one heavy slot (`--max 4`), record the would-block count and each name in the Result, and sort each as a real read, a write, or a false positive.
4. Baseline (D-5): write the 389 names of the observation into `tests/fixtures/test_home_guard/baseline.tsv`, let the wrapper tolerate a listed test that reads only the decoy `user-assets.json` (REQ-12), and add the tests that cap the file and check its names (REQ-10).
5. Enforce (D-2): the default is `enforce`; run the whole suite again with no failure, then one `moai check affected` at READY.

## Risks

- The `wrapper-scripts` feature of nextest is experimental in its own documentation ("This is an advanced feature, and it can cause your tests to silently stop working if used incorrectly") and can change in a later nextest. The configuration pins `nextest-version` and enables the feature by name, so an older nextest stops with an error instead of running unguarded. A newer nextest that changes the feature is not covered and shows up as failed tests first. The Result names the version run (0.9.152 here).
- A wrapper that does not run its test makes nextest report a pass: the wrapper returns the test's own exit code and AC-4 plants a missing and a non-running wrapper to show the test does not pass.
- Cost per test: the wrapper adds child processes to each of about 7300 test processes (a median of 50.8 ms added wall time per wrapped process, 41 to 68 ms over five loops of 100 runs of `true` with and without the wrapper at a load of about 32; the audit saw 50 to 90 ms). The Result quotes the loop; the whole-suite cost is not measured apart from that.
- A decoy changes a test that expects an empty HOME: the observation run in step 3 found none (7317 of 7320 passed, the 3 failures were repository scans that the new files tripped, fixed since).
- The baseline hides a new reader that arrives under the name of a listed test, or a listed test that stops reading: the first is bounded by REQ-12 (it tolerates only the one decoy and its folders), the second shows only when the baseline is rewritten by the follow-up that makes `identity` take HOME as an argument.
- `atime` is not kept up on a file system mounted without it, and Linux and Windows were not measured here: AC-1 reads a decoy the same way on every host that runs the suite and fails there when the read is not seen, so a host that cannot see reads is not reported as guarded.
- The wrapper replaces `HOME` for tools a test starts: git, cargo and rustup paths are the first ones to break, which the observation run shows by name.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A wrapper script per test process, not a setup script and not cargo `[env]`. A setup script runs once and exports one HOME to every test through `$NEXTEST_ENV` (measured in a scratch crate: one run of the script, the same HOME in the lib and the integration binary, the caller's HOME overridden), so a read cannot be told apart per test, and nextest has no cleanup step after it. cargo `[env]` is static text and cannot make a temporary folder. | Documents read in full from the nextest repository at tag `cargo-nextest-0.9.152` on 2026-11-23T14:15:49Z: `site/src/docs/configuration/wrapper-scripts.md` and `setup-scripts.md` (raw.githubusercontent.com/nextest-rs/nextest/cargo-nextest-0.9.152/…): wrapper scripts since 0.9.98, `relative-to` since 0.9.99, `platform = { host = "cfg(unix)" }` is the platform form, a wrapper gets the test binary as its first argument; cargo `[env]` text read through a WebFetch digest on 2026-11-23 (https://doc.rust-lang.org/cargo/reference/config.html#env, digest only, so unverified as raw text); plus the scratch-crate runs of this lane (nextest 0.9.152, macOS arm64): the wrapper failed only the planted reading test of four. Unverified: Linux and Windows behavior, whether `$NEXTEST_ENV` of a later nextest still overrides the caller's HOME | Delete the rule and the script |
| D-2 | The default is `enforce`. The device started in observe, the whole suite ran once under it (D-5), and the leader read the count and chose to enforce with a baseline. | The observe-first rule of `internal/sdd`; the leader's notes 7562 ("enforce 를 기본으로 바꾸는 것은 개수를 본 뒤 다시 노트로") and 7600 ("기본은 enforce") | Set the default to observe in the script |
| D-3 | A read is seen through the `atime` of decoy files and folders, a change through the modification time of the entries against a reference file made after the decoys. | Measured on APFS in a temporary folder: a read moves a decoy's `atime` from 2000-01-01 to now | Replace the decoys with a file-access tracer |
| D-4 | unix only; a Windows host runs its tests as before. | The wrapper needs `sh`; no Windows run exists here | Add a Windows wrapper |
| D-5 | The 389 tests that read the decoy through `identity::identity` go on a baseline file that only shrinks, until a separate change makes `identity` take HOME as an argument (the leader opened a card for it). Their source: one observe run of the whole suite, head `57bcc17de2`, nextest 0.9.152, 2026-11-23, 7320 tests run, log `scratchpad/guard-observe.log` of the lane session (389 lines, each a distinct test, one distinct set of touched paths, 0 writes), reduced to `binary<TAB>test` lines. | The leader's note 7600 (option A of note 7598) | Delete the file and the REQ-12 branch of the wrapper; the 389 tests then fail under enforce |

## Result

Measured 2026-11-23/24 on macOS arm64 in this branch's worktree, nextest 0.9.152, in-tree `target/`, heavy slots `--max 4`. Linux and Windows were not run.

- AC-1 met: in a scratch crate with the repository's `.config/nextest.toml` and wrapper, the planted reader and writer fail with a `HOME-GUARD:` line in the default mode, a failing reader keeps its own failure, and `MOAI_HOME_GUARD=observe` passes it with the line printed (`tests/test_home_guard.rs`, 37 tests, 37 passed after the repair below; 35 of 35 at the first READY).
- AC-2 met: an unwritable `TMPDIR` exits 70 before the test runs; a link or lookalike folders, among them a `moai-test-home.1` folder and `.ref` file of another process, lose nothing, and neither does a folder or a link that already carries the name the old wrapper would have built from its own process id; a child `cargo --version` finds its toolchain; `MOAI_HOME_GUARD_LOG` receives the line.
- AC-3 met: allowlist entries without a reason, with an empty reason, naming nothing, or outside the filter fail; the repository allowlist (0 entries, filter `all()`) passes; the configuration holds `experimental = ["wrapper-scripts"]`, `nextest-version`, `cfg(unix)` and the script path.
- AC-4 met: with the script missing, and with a script that exits 3 without running its arguments, the planted test is attempted and does not pass.
- AC-5 met for the suite: the whole `cargo nextest run --offline --locked --no-fail-fast` under the default `enforce` ran 7333 tests, 7333 passed, 4 skipped, in 1249 s at a load of about 30; its log (`MOAI_HOME_GUARD_LOG`) holds 389 `tolerated (baseline)` lines and no `touched` line. The clause that `moai check affected` runs under the same device is not_run here and is recorded in the READY note.
- AC-6 met: a listed test that reads only the decoy passes with a `tolerated` line (also with the baseline found next to the script and no path setting); a listed test that also writes, one that reads another decoy, one that is not listed, and one listed under another binary fail; a tolerated test that fails keeps its own exit code. The repository baseline has 389 lines, sorted without repeats, under a cap of 389, and every line names an existing test.

Baseline source (D-5): one observe run of the whole suite at head `57bcc17de2`, nextest 0.9.152, 2026-11-23, 7320 tests run, 7317 passed and 3 failed (three repository scan tests tripped by the new files, fixed in `048e9cb4d8`), log `guard-observe.log` in the lane session's scratchpad (389 `HOME-GUARD:` lines). All 389 are reads of the decoy `user-assets.json` through `identity::identity` (env HOME → `home_state::home_go_markers`); 0 writes; no false positive. The log is not in the repository: its 389 distinct `binary<TAB>test` pairs are the baseline file.

Cost (measured 2026-11-24, macOS arm64, load about 32, in a temporary `TMPDIR` under the scratchpad): five loops of 100 runs of `/usr/bin/true`, plain against `sh tools/test_home_guard.sh /usr/bin/true`, added 67.5, 50.8, 46.0, 40.8 and 59.1 ms per run, a median of 50.8 ms, and the loops left no folder behind. This replaces the earlier figures of about 14 ms and 56 ms, which came from a load of about 70 and from an earlier version of the script, and which the audit did not reproduce (50 to 90 ms). The cost of the whole suite was not measured apart.

Repair after the audit (head of the first READY `13ccc8de68`, B1): the wrapper named its folder `moai-test-home.$$` and its reference file `moai-test-home.$$.ref` from the process id and set the cleanup trap before it made them, so a leftover under the same name (a killed process and a reused id) was removed along with `exit 70` (D1), and a link in the place of the reference file had its target emptied (D2). The folder and the reference file now come from `mktemp -d "$TMPDIR/moai-test-home.XXXXXX"` and `mktemp "$TMPDIR/moai-test-home-ref.XXXXXX"`, and each trap is set only after its `mktemp` returned. The tests `a_folder_left_under_the_name_of_this_process_is_not_removed` and `a_link_left_in_the_place_of_the_reference_file_keeps_its_target` came first and failed as the audit saw (exit 70; a target of 0 bytes) in commit `dbab204592`; they pass with the fix in `210f6dd9fd`. `tests/test_home_guard.rs` now has 37 tests, 37 passed. The dangerous path (`rm -rf` of the wrapper) ran in a temporary `TMPDIR` only, as those two tests.

Guard mutants (each applied to a committed state, the filtered test binary run without the wrapper rule, then restored; the always-failing `this_process_runs_under_a_guarded_home_when_nextest_runs_it` is excluded because it fails in an unwrapped run by design):
- remove the decoy aging `touch -a` → killed by `every_decoy_starts_with_an_access_time_before_2001`.
- remove every `|| exit 70` of the setup → killed by `an_unwritable_tmpdir_stops_the_wrapper_with_70_before_the_test_runs` (removing only the `mkdir` one survives: the next lines exit 70 too).
- remove the cleanup trap → killed by `the_wrapper_removes_only_the_folder_it_made`, `a_test_that_touches_nothing_passes_under_a_guarded_home`.
- remove the `-anewer` clause (reads unseen) → killed by 11 tests.
- invert the enforce check → killed by 13 tests.
- every test counts as listed → killed by `a_test_that_is_not_listed_fails_and_so_does_one_under_another_binary`.
- ignore the extra touched paths → killed by `a_listed_test_that_also_writes_is_not_tolerated`, `a_listed_test_that_reads_another_decoy_is_not_tolerated`.
- stop setting `USERPROFILE` → killed by `the_userprofile_follows_the_home`.
- cleanup that also removes `moai-test-home.*` of others → survived the first lookalike case (a dash name); the test now holds a dotted folder and a `.ref` file of another process and kills it.
- repair of B1, each applied on the committed fix and restored (the tree was clean after): remove the same-name folder before `mktemp -d` (the audit's M7) → killed by `a_folder_left_under_the_name_of_this_process_is_not_removed`; remove the same-name reference file before its `mktemp` → killed by `a_link_left_in_the_place_of_the_reference_file_keeps_its_target`; make the reference file again with `: >` under the predicted name → killed by the same test; make the folder again with `mkdir` under the predicted name → killed by `a_folder_left_under_the_name_of_this_process_is_not_removed`.
- test side: no sorted check, no cap check, no name check → killed by `an_unsorted_or_repeating_baseline_is_refused`, `a_baseline_that_grew_is_refused`, `a_baseline_line_that_names_nothing_is_refused`.

Gaps: Linux and Windows are not measured; a host that keeps no access time (`noatime`, coarse timestamps) does not see reads, and the decoy test fails there instead of reporting the host as guarded; `wrapper-scripts` is experimental in nextest; the plain `cargo test` runner is not guarded. The root fix that lets `identity` take HOME as an argument, and so empties the baseline, is the leader's separate card.
