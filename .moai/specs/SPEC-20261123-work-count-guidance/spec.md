---
id: SPEC-20261123-work-count-guidance
title: "test-design says how to bound work by a count instead of a wall clock"
status: draft
revision: 1
created: 2026-11-23
---

# test-design says how to bound work by a count instead of a wall clock

## Intent

테스트가 "이 코드는 입력 길이에 비례하는 시간 안에 끝난다"를 지켜야 할 때 벽시계 초(秒)로 판정하면 부하에서 플레이키 테스트가 돼요(카드 비밀 글 검사 테스트가 부하 79~90 에서 두 번 떨어졌어요). 이미 `card::input::tests::secret_screen_work_stays_linear_on_one_long_word` 와 `tests/ac_command_lint.rs` 가 시간 대신 작업량을 세는 방식으로 고쳐졌고, 그 방법을 `skills/moai/references/internal/test-design.md` 에 적는 일만 남았어요(이전 SPEC 의 Non-goals 가 그 파일을 다른 카드가 쥐고 있다는 이유로 미뤘고, 그 카드는 끝났어요). 에이전트가 다음에 시간 한도 테스트를 쓸 때 같은 방법을 처음부터 쓰게 해요.

이 SPEC 은 안내 문장 하나와 그 문장을 고정하는 테스트만 바꿔요. 검사 코드와 이미 고친 테스트는 건드리지 않고, 현재 팁에서 그 테스트가 여전히 안 흔들리는지만 다시 재요.

## Non-goals

- 남은 벽시계 판정(경과 시간 한도 단언, 고정 상한 기다림)을 더 고치는 일: 목록은 이전 SPEC 의 Result AC-5 에 있고, 부하로 떨어진 것만 그때그때 고쳐요.
- 제품 코드나 이미 고친 테스트를 바꾸는 일.
- 카운터를 제품 빌드에 넣는 일.

## Requirements

- REQ-1: The test-design reference shall tell a test author to decide a time bound by a count of work and not by a wall clock.
- REQ-2: The test-design reference shall name the four parts of the count: a counter on the thread that runs the code, a limit set from the input length, a stop of the code at the limit, and a check that the counter counted something.
- REQ-3: The test-design reference shall tell a test author to prove the count with a mutant that makes the code quadratic and to expect the test to fail on it.
- REQ-4: The test-design reference shall say that the count sees only the work that calls the counter.
- REQ-5: The source manifest shall carry the hash of the test-design reference as it is after this change.
- REQ-6: The new text shall carry no card number, Issue number or host-specific word.
- REQ-7: The secret screen work-count test shall pass on the current integration tip in every run of 30 runs under a machine load of 100 or more.
- REQ-8: The secret screen work-count test shall fail on a mutant that rebuilds the camel-hump name at every hump.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): `tests/sdd_guidance.rs` holds a test that reads the reference and finds one numbered rule that names the count, the thread, the limit from the input length, the stop at the limit, the counter check, the quadratic mutant and the uncounted-work caveat; the test fails before the text is added and passes after · verify: `cargo test --locked --offline --test sdd_guidance work_count`
- AC-2 (REQ-5): the manifest check finds the hash of `references/internal/test-design.md` equal to the file's hash · verify: `cargo test --locked --offline --test agent_defs_package` and the manifest step of the repository check
- AC-3 (REQ-6): the guidance scans that cover this file pass · verify: `cargo test --locked --offline --test sdd_guidance`
- AC-4 (REQ-7): 30 of 30 runs of `secret_screen_work` pass with the one-minute load average at 100 or more at the start of each run, the load printed for each run · verify: the built test binary run 30 times under burner processes
- AC-5 (REQ-8): the camel-hump rebuild mutant makes the test fail quickly, the source is restored and the tree is clean afterwards · verify: a script that edits `src/card/input.rs`, runs `card::input::tests::secret_screen_work`, restores the file and checks `git status --short`

## Scope

- write: `.moai/specs/SPEC-20261123-work-count-guidance/**`, `skills/moai/references/internal/test-design.md`, `skills/moai/references/source-manifest.json` (hash only), `tests/sdd_guidance.rs`
- read-only: `src/card/input.rs`, `src/redact.rs`, `tests/ac_command_lint.rs`, `.moai/specs/SPEC-20261120-card-screen-work-count/spec.md`
- forbidden: any other path, the real HOME

## Plan

1. SPEC (this file), then `moai spec check`.
2. RED: the `work_count` test in `tests/sdd_guidance.rs`; it fails on the unchanged reference.
3. GREEN: one numbered rule in `test-design.md` and the manifest hash written by the repository's manifest tool.
4. Evidence on the current tip: 30 runs under load and the camel-hump mutant (AC-4, AC-5); each run's load and result written in the Result.
5. Neighbours: the guidance tests, the package tests, fmt, clippy, `git diff --check`, `moai trailer-check`, one bare `check affected`.

## Risks

- A rule that is too long is skipped by readers: it is one numbered item like the others, with the method in four parts.
- A test that pins words can break on a harmless rewording: the test looks for the named parts (count, thread, limit, stop, counter check, mutant, uncounted work) in one rule and not for a whole sentence.
- Load of 100 or more may not be reachable on the day: the run prints the load of each run, and a run below 100 is reported as it is and not counted for AC-4.
