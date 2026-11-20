---
id: SPEC-20261120-card-screen-work-count
title: "the card secret screen's time bound is a count of work, not a wall clock"
status: draft
revision: 1
created: 2026-11-20
---

# the card secret screen's time bound is a count of work, not a wall clock

## Intent

카드 글(제목·본문·노트·`--ac`)의 비밀 검사는 낱말 하나가 아주 길어도 길이에 비례하는 시간 안에 끝나야 해요(길이의 제곱이면 40 KB 가 30초 걸렸어요). 지금 그 약속을 지키는 시험은 벽시계 15초로 판정해요: `card::input::tests::secret_screen_time_on_one_long_word_has_a_bound` 가 부하 79~90 에서 한 번, 병합 전 검사(부하 90~130)에서 한 번, 오늘만 두 번 떨어졌어요. 혼자 돌리면 3~4초에 끝나는 시험이에요. `verify.md` 는 시험이 벽시계에 기대면 안 된다고 하고 감사자는 그런 한도를 발견으로 다뤄요. 같은 종류의 문제가 두 번 나왔으니 시험을 고쳐요.

고치는 방식은 이미 한 번 쓴 방식이에요(`agents_local_merge` 의 반복 접두 검사): 검사가 처리한 글자 수를 시험에서만 세고, 입력 길이의 몇 배 이하인지로 판정해요. 센 값은 기계 부하와 상관없이 같아서 어느 기계에서나 같은 답이 나와요. 세는 곳은 스레드마다 따로예요(시험 스레드들이 한 카운터를 두고 다투지 않게). 시험은 한도를 넘는 순간 검사를 멈추게 해서, 제곱 시간으로 되돌아간 변경이 몇 분을 먹지 않고 바로 실패하게 해요.

벽시계 15초를 쓰는 시험은 셋이에요: 위 단위 시험, 그리고 `tests/card_secret_followups.rs` 의 `separators_long_one_word_stays_within_the_time_bound` 와 `camel_long_name_stays_within_the_time_bound`. 통합 시험은 라이브러리의 시험 전용 카운터에 닿지 못하므로 같은 입력과 같은 답을 단위 시험으로 옮기고 통합 쪽의 시간 시험은 지워요(답을 확인하는 부분은 단위 시험이 그대로 해요). 낙타 혹 문자열을 계속 만들던 옛 제곱 시간 결함(감사 B1)은 이 시험에서 계속 실패해야 해요.

## Non-goals

- 비밀 검사의 판정 규칙을 바꾸는 일: 어떤 낱말을 막는지는 그대로예요.
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

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-5): `secret_screen_work_stays_linear_on_one_long_word` passes and no card secret test holds a 15 second clock bound any more · verify: `cargo test --locked --offline --lib card::input::tests::secret_screen_work`, and `grep` of `from_secs(15)` over `src/card` and `tests/card_secret_followups.rs` finds none
- AC-2 (REQ-1): the test passes 30 of 30 runs while the machine's load average is 100 or more, the load printed for each run · verify: the built test binary run 30 times
- AC-3 (REQ-2, REQ-3, REQ-4): a camel-hump rebuild at every hump, a shared-screen call on every separator's rest of the word, a name check on a growing prefix at every equals sign and a counter that counts nothing each make the test fail, the three quadratic ones fast · verify: a script that edits the sources, runs the test and restores them
- AC-4 (REQ-6): the module's other tests and the card integration tests answer as before and the product binaries build · verify: `cargo test --locked --offline --lib card::`, `cargo test --locked --offline --test card_secret_followups`, `cargo build --locked --offline --bins`
- AC-5 (REQ-7): the Result lists the remaining wall-clock assertions · verify: `grep` over `src/` and `tests/`

## Scope

- write: `src/card/input.rs`, `src/redact.rs`, `tests/card_secret_followups.rs`, `.moai/specs/SPEC-20261120-card-screen-work-count/**`
- read-only: `tests/common/**`, `src/ledger/workflow.rs`
- forbidden: `skills/**`, `docs/**`, `schemas/**`, the real HOME

## Plan

Tests first, in the order below, one commit each.

1. RED: the counter module in `src/redact.rs` (test build only) and the new unit test `secret_screen_work_stays_linear_on_one_long_word` with all jobs and an aliveness check; no call sites yet, so the screen reports no work and the test fails on REQ-4.
2. GREEN: the call sites in `src/card/input.rs` and `src/redact.rs`, the lower-casing routed through one counted helper, the bound measured and written down.
3. REFACTOR: remove the three wall-clock tests and the `timed` helper.
4. Mutants (at most 5): camel rebuild per hump, the shared screen on every separator's rest, a growing-prefix name check, a dead counter, a limit that never stops the screen.
5. Neighbours: `--lib` of the card module, the card integration tests, fmt, clippy, `git diff --check`.

## Risks

- A bound set too loose lets a middle-ground slowdown pass: the bound is the measured linear factor with a small margin, and a mutant between the linear and the square factor is part of the evidence.
- A counter that no code path feeds passes nothing by accident: REQ-4's aliveness check fails then.
- Work done in a function nobody counts is invisible: the three hot spots of the earlier audits (the camel names, the shared screen and the name check) are counted, and the SPEC's Result says what is counted.

## Result

(pending)
