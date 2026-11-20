---
id: SPEC-20261120-card-screen-work-count
title: "the card secret screen's time bound is a count of work, not a wall clock"
status: draft
revision: 2
created: 2026-11-20
---

# the card secret screen's time bound is a count of work, not a wall clock

## Intent

카드 글(제목·본문·노트·`--ac`)의 비밀 검사는 낱말 하나가 아주 길어도 길이에 비례하는 시간 안에 끝나야 해요(길이의 제곱이면 40 KB 가 30초 걸렸어요). 지금 그 약속을 지키는 시험은 벽시계 15초로 판정해요: `card::input::tests::secret_screen_time_on_one_long_word_has_a_bound` 가 부하 79~90 에서 한 번, 병합 전 검사(부하 90~130)에서 한 번, 오늘만 두 번 떨어졌어요. 혼자 돌리면 3~4초에 끝나는 시험이에요. `verify.md` 는 시험이 벽시계에 기대면 안 된다고 하고 감사자는 그런 한도를 발견으로 다뤄요. 같은 종류의 문제가 두 번 나왔으니 시험을 고쳐요.

고치는 방식은 이미 한 번 쓴 방식이에요(`agents_local_merge` 의 반복 접두 검사): 검사가 처리한 글자 수를 시험에서만 세고, 입력 길이의 몇 배 이하인지로 판정해요. 센 값은 기계 부하와 상관없이 같아서 어느 기계에서나 같은 답이 나와요. 세는 곳은 스레드마다 따로예요(시험 스레드들이 한 카운터를 두고 다투지 않게). 시험은 한도를 넘는 순간 검사를 멈추게 해서, 제곱 시간으로 되돌아간 변경이 몇 분을 먹지 않고 바로 실패하게 해요.

벽시계 15초를 쓰는 시험은 셋이에요: 위 단위 시험, 그리고 `tests/card_secret_followups.rs` 의 `separators_long_one_word_stays_within_the_time_bound` 와 `camel_long_name_stays_within_the_time_bound`. 통합 시험은 라이브러리의 시험 전용 카운터에 닿지 못하므로 같은 입력과 같은 답을 단위 시험으로 옮기고 통합 쪽의 시간 시험은 지워요(답을 확인하는 부분은 단위 시험이 그대로 해요). 낙타 혹 문자열을 계속 만들던 옛 제곱 시간 결함(감사 B1)은 이 시험에서 계속 실패해야 해요.

같은 종류의 문제가 하나 더 있어요(리더 노트로 범위에 더해졌어요): `tests/home_retire_cli.rs` 의 시험 여럿이 "이 폴더가 사용 중"을 흉내 내려고 60초 뒤에 끝나는 `sleep` 을 띄워요. 부하로 시험이 60초를 넘기면(135초 걸린 적이 있어요) 흉내 프로세스가 먼저 끝나서 시험이 떨어져요. 흉내를 `cat` 으로 바꿔요: 아무도 쓰지 않는 파이프를 기다리는 `cat` 은 시험이 지우거나(`Kill`) 시험 프로세스가 끝날 때(파이프가 닫혀 파일 끝을 읽을 때)까지 살아 있어서, 얼마나 사는지를 시계가 정하지 않아요.

## Non-goals

- 비밀 검사의 판정 규칙을 바꾸는 일: 어떤 낱말을 막는지는 그대로예요.
- 시험 설계 가이드(`skills/moai/references/internal/test-design.md`)에 "시간 대신 작업량 카운터" 단락을 더하는 일: 그 파일과 매니페스트를 다른 카드가 쥐고 있어서, 파일이 풀린 뒤 따로 해요.
- 저장소의 다른 벽시계 판정을 모두 고치는 일: 목록만 남기고, 떨어진 적이 있는 것만 고쳐요(이 SPEC 의 시험 셋).
- 제품 코드에 카운터를 넣는 일: 센 값은 시험 빌드에만 있어요.

## Requirements

- REQ-1: The unit test of the card secret screen shall decide a long word's time bound by a count of the work the screen did and not by a wall clock.
- REQ-2: The counter shall add every pass of the screen's own scans, every byte lower-cased and every byte handed to the shared credential screen, on the thread that runs the screen.
- REQ-3: IF the work of one screen passes the limit the test set for it THEN the counter shall stop that screen with a panic that names the count and the limit.
- REQ-4: IF a screen counted less work than the length of its text THEN the test shall fail.
- REQ-5: The unit test shall run the jobs of the three time tests with the answers they asserted: the four unit shapes and the shape at the end, the three separator shapes with the value after the last separator, and the camel names joined, spaced and with a secret word behind the last hump.
- REQ-6: The product build shall carry the counter only as test-build statements and answer every text as it did.
- REQ-7: The result of this work shall list every other wall-clock assertion of `src/` and `tests/` by file and line.
- REQ-8: The in-use stand-ins of `tests/home_retire_cli.rs` shall be programs that live until the test drops them or the test process ends and not programs that end after a fixed time.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-5): `secret_screen_work_stays_linear_on_one_long_word` passes and no card secret test holds a 15 second clock bound any more · verify: `cargo test --locked --offline --lib card::input::tests::secret_screen_work`, and `grep` of `from_secs(15)` over `src/card` and `tests/card_secret_followups.rs` finds none
- AC-2 (REQ-1): the test passes 30 of 30 runs while the machine's load average is 100 or more, the load printed for each run · verify: the built test binary run 30 times
- AC-3 (REQ-2, REQ-3, REQ-4): a camel-hump rebuild at every hump, a shared-screen call on every separator's rest of the word, a name check on a growing prefix at every equals sign and a counter that counts nothing each make the test fail, the three quadratic ones fast · verify: a script that edits the sources, runs the test and restores them
- AC-4 (REQ-6): the module's other tests and the card integration tests answer as before and the product binaries build · verify: `cargo test --locked --offline --lib card::`, `cargo test --locked --offline --test card_secret_followups`, `cargo build --locked --offline --bins`
- AC-5 (REQ-7): the Result lists the remaining wall-clock assertions · verify: `grep` over `src/` and `tests/`
- AC-6 (REQ-8): no `sleep` program is started in `tests/home_retire_cli.rs`, its tests pass, and a stand-in that ends early fails them (so they do depend on the stand-in living and the old stand-in's end was the cause) · verify: `grep` of `"sleep"` in the file, `cargo test --locked --offline --test home_retire_cli`, and a mutant that makes the stand-in end at once

## Scope

- write: `src/card/input.rs`, `src/redact.rs`, `tests/card_secret_followups.rs`, `tests/home_retire_cli.rs`, `.moai/specs/SPEC-20261120-card-screen-work-count/**`
- read-only: `tests/common/**`, `src/ledger/workflow.rs`
- forbidden: `skills/**`, `docs/**`, `schemas/**`, the real HOME

## Plan

Tests first, in the order below, one commit each.

1. RED: the counter module in `src/redact.rs` (test build only) and the new unit test `secret_screen_work_stays_linear_on_one_long_word` with all jobs and an aliveness check; no call sites yet, so the screen reports no work and the test fails on REQ-4.
2. GREEN: the call sites in `src/card/input.rs` and `src/redact.rs`, the lower-casing routed through one counted helper, the bound measured and written down.
3. REFACTOR: remove the three wall-clock tests and the `timed` helper.
4. Mutants (at most 5): camel rebuild per hump, the shared screen on every separator's rest, a growing-prefix name check, a dead counter, a limit that never stops the screen.
5. Stand-ins: one helper in `tests/home_retire_cli.rs` starts the in-use program (`cat` on a pipe) and the five `sleep 60` spawns call it; a mutant that ends the stand-in at once shows the tests depend on it.
6. Neighbours: `--lib` of the card module, the card integration tests, the home retire tests, fmt, clippy, `git diff --check`.

## Risks

- A bound set too loose lets a middle-ground slowdown pass: the bound is the measured linear factor with a small margin, and a mutant between the linear and the square factor is part of the evidence.
- A counter that no code path feeds passes nothing by accident: REQ-4's aliveness check fails then.
- Work done in a function nobody counts is invisible: the three hot spots of the earlier audits (the camel names, the shared screen and the name check) are counted, and the SPEC's Result says what is counted. The count is linear in the counted work only: the audit's mutant that copied the rest of the name at every 64th hump in `camel_hump_secret`, which calls no `work::add`, passed the test (the test took 7.4 s against 2.6 s unmutated). The bodies of `camel_hump_secret` per hump and of `word_has_secret_shape` (about twelve `contains` passes charged as one `word.len()`) are counted as one unit each, so a new copy or pass inside them must call `work::add` too; a repository scan that bans `to_string`, `to_owned`, `format!` and `collect` inside the hump loop would close this class and is left to a follow-up.

## Result

All six criteria are met. Commits on `spec/card-screen-work-count-20261120` (base `341f6c7abb`): SPEC `5582cdc5a6`, RED `1e9d2e02ad` (the new test failed with `0 units counted for 65543 bytes — the counter counted nothing`, and `moai tdd red` recorded it), GREEN `b162e01b49` (`moai tdd green` recorded the check as passed), rustfmt `76827215bb`, stand-ins `b9728fadbb`, SPEC revision 2 `33bded6887`, and the commit of this Result.

- AC-1: met — `card::input::tests::secret_screen_work_stays_linear_on_one_long_word` passes (1 passed, 1.9 to 5.1 s) and runs 17 jobs on their own threads: the four unit shapes and the shape at the end at 1 MiB, the three separator shapes with the value at the end at 200000 and 1048575 bytes, and the camel name joined, spaced and with a secret word behind the last hump at 65536 and 1048575 bytes; each asserts the answer the clock tests asserted. `grep -rn 'from_secs(15)' src/card tests/card_secret_followups.rs` finds 0 lines; the old unit test, the `timed` helper and the two integration time tests are gone. The only clock left in the test is a 900 second hang guard on the channel that never decides it (every thread ends by an answer or by the limit).
- AC-2: met — the count is the same on every machine and at every size: 10.50 units per byte for the separator shapes and the shape at the end, 13.00 for `a,` and `a:`, 13.44 for the JSON shape, 15.00 for `a=`, 12.00, 17.00 and 18.00 for the three camel jobs, at 64 KiB and at 1 MiB alike; the limit is 20. Run three times in a row, the printed work lines hash alike (`287e7a535bde` each time, once the `test <name> ... ` prefix that libtest writes in front of one of the lines is cut off). 30 of 30 runs passed with 55 `nice 10` CPU burners started by the script, the one-minute load average 111 to 195 over the runs (33 before the burners); an earlier loop without burners (load 35 to 74) also ended 30 of 30, but its output was not kept and its binary was not checked, so it is not relied on. The old test failed at load 79 to 90 and again in a merge check at 93 to 130.
- AC-3: met — five mutants (python3 of a script that edits the sources, runs `card::input::tests::secret_screen_work` and restores them; 5 of 5 killed, `restored: True`): a lower-casing of the rest of the name at every camel hump (the audit B1 bug) stopped at 1,376,264 units for 65538 bytes, the shared screen run on the rest of the word at every separator (audit B2) stopped at 4,200,335 units for 200020 bytes, the name check on the whole prefix at every equals sign stopped at 20,971,615 units for 1,048,576 bytes, a counter that counts nothing failed the aliveness check, and two extra byte passes of the shared screen per call (a small slowdown) went past 20 units per byte (the spaced camel job stopped at 1,376,268 units for 65540 bytes). The three square-time mutants were stopped by the limit, not by waiting; the run time of each was not measured.
- AC-4: met — on head `33bded6887`: `check-affected base=341f6c7abb90 head=33bded688781 targets=61 py=8 scans=listed:41 diff_check=0 fmt=0 clippy=0 test_lib=0 test_bins=0 test_targets=0 build_cli=0 python=0 result=PASS` (`--lib` 1638 passed; the 61 test targets it ran include 20 `card_*` targets, `card_secret_followups` among them, and `home_retire_cli`; the binaries built). An earlier run of the same command on `b162e01b49` had `fmt=1` (three statements of the new test), fixed in `76827215bb`; every other step of it was 0.
- AC-5: met — the remaining assertions of an elapsed bound in `src/` and `tests/`, 31 lines found by two rounds of `grep -rnE` (the first for `elapsed() <`, `took <`, `waited <`, `as_secs() <` and the like; the second, after the audit, for a duration variable such as `median`, `secs` or `waited` compared with a `Duration`, a `BOUND` or a `BUSY_TIMEOUT`), listed and not changed. A list found by grep is the lines those patterns match, not proof that no other bound is written some other way. The lines: `src/agents_local_merge.rs:1048`, `src/decide/transport.rs:539`, `tests/forge_cli.rs:510`, `tests/status_line_cli.rs:663`, `tests/launcher_auto_init.rs:84,477,596`, `tests/statusline_counts_refresh.rs:176`, `tests/run_verify.rs:559,571`, `tests/lane_send_cli.rs:440,528`, `tests/lane_screen_cli.rs:1520`, `tests/edit_diagnostics_cli.rs:436,518,598`, `tests/heavy_slot_cli.rs:161`, `tests/ledger_recover.rs:1886`, `tests/launcher_agent_numbering.rs:241`, `tests/hook_dispatch_cli.rs:579`, `tests/card_timeline_cli.rs:435`, `tests/home_state_plan.rs:202`, `tests/product_cli.rs:3344,3392,3436`, `tests/common/prop.rs:118`, and the five the audit added: `src/statusline/state.rs:2455` and `tests/statusline_cards.rs:288` (a wait under half of `BUSY_TIMEOUT`), `tests/edit_diagnostics_cli.rs:793` (a median under 50 ms), `tests/hook_secret_tokens_cli.rs:547,560` (30 seconds, `BOUND_SECS`, on the secret-shape near-miss and long-run timings: the same kind as the tests this work fixed, and the likeliest next to fail under load). Waits with a fixed ceiling (`recv_timeout` or a deadline loop): `src/profile.rs:674`, `src/hook_guards.rs:3336`, `src/hook_config.rs:689`, `src/lane_loop.rs:535`, `src/versioning.rs:1604`, `src/decide/transport.rs:168,384,438`, `tests/forge_cli.rs:507`, `tests/statusline_counts_refresh.rs:125`, `tests/ledger_recover.rs:1509`, `tests/index_cli.rs:129`, `tests/profile_plan_cli.rs:488`, `tests/boundary_output_scope.rs:31`, `tests/ledger_folders.rs:546`, `tests/ledger_open.rs:613`, `tests/ledger_reservation.rs:1144`. None of these is known to have failed under load in this lane's runs; each is a candidate for the same treatment when one does.
- AC-6: met — `grep -n '"sleep"' tests/home_retire_cli.rs` finds 0 lines (five `sleep 60` spawns replaced by one helper, `in_use`, that starts `cat` on a pipe); `cargo test --test home_retire_cli` 24 passed unmutated (141 s under load); the mutant that makes the stand-in end at once (`true` instead of `cat`) fails exactly the five tests that use a stand-in (19 passed, 5 failed) and the file is restored (`restored: True`), so the tests do depend on the stand-in living and a stand-in that ends first fails them, as the old one did on a loaded machine.
- Not done, by design: the paragraph on work counters in `skills/moai/references/internal/test-design.md` that the leader asked for later; that file and the manifest are held by other cards (named in Non-goals), so it follows as its own small change.
- Not run: a Linux or Windows build, the CI of a real pull request, and an independent read-only audit of this branch (the lane starts it). The burner load is made by the script, not by other lanes alone, and `nice 10` burners still count in the load average.
- Slips: the first 30-run loop ran at load 35 to 74, below the 100 the criterion asks, so it was repeated with burners. Both loop scripts picked their test binary as the newest file of the build folder, and a later hand run picked a wrong file that way (its output had no work lines); the lib test binary was then named by cargo (`--message-format=json`, `moai_cli-692052709cdd41e5`) and used for the three-run comparison. Only the burner loop is relied on: it printed a hash of the work lines for each run and every hash was of a non-empty set (the 17 lines only this test prints). The hashes differed from run to run only by the libtest name prefix, which was checked.
- Audit round 1 (independent read-only auditor, head `2f49f81bee`) and the repair: verdict FAIL on one blocking finding, B1, the AC-5 list lacked five lines (`src/statusline/state.rs:2455`, `tests/statusline_cards.rs:288`, `tests/edit_diagnostics_cli.rs:793`, `tests/hook_secret_tokens_cli.rs:547,560`); they are in the list above (26 plus 5, 31 lines), and the wording no longer claims the list is exhaustive. Non-blocking findings: N1 and N4 (work in code that calls no `work::add` stays invisible; the audit's mutant that copied the rest of the name at every 64th hump passed) are written into Risks as "linear in the counted work", with the scan that would close the class left to a follow-up; N2 (the test did not compare the count with the limit itself, so a limit check that never fires left it green) is repaired by one check in the test loop (`taken <= text.len() * WORK_PER_BYTE`), and the audit's mutant for it (the limit check never fires together with four extra counted bytes per byte) now fails the test (`1376304 units for 65540 bytes`, 0.23 s) while the unmutated test still passes (1 passed, worst 18.00 per byte, limit 20); N3 (the aliveness check catches only a counter that counts nothing) is accepted, because a floor near the measured minimum of 10.5 per byte would fail a screen that gets faster.
