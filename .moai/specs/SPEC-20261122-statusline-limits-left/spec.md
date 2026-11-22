---
id: SPEC-20261122-statusline-limits-left
title: "The limit row shows what is left of the context window and the week"
status: draft
revision: 2
created: 2026-11-22
---

# 상태줄 한도 줄: 쓴 비율이 아니라 남은 비율

## Intent

사용자 지시예요. 상태줄 3번째 줄(한도)의 `CW`·`5H`·`7D` 숫자는 쓴 비율이 아니라 **남은 비율**이에요. 배터리 아이콘은 남은 양으로 골라요: 넉넉하면 `🔋`, 얼마 안 남으면 `🪫`. 줄의 순서는 `CW | 5H(있을 때) | 7D | 버전` 이에요. `SPEC-20261122-statusline-limits` 의 REQ-2·REQ-3 이 적은 숫자를 "남은 비율"로 바꿔 읽고, REQ-4 와 D-2 의 "`5H` 는 70% 를 넘을 때만"을 이 SPEC 이 바꿔요. 그 밖의 것(카운트다운, `💰`, 버전, 폭에 따른 모양, 숨김 옵션, 사용량 기록)은 그대로예요.

개정 2(사용자 말: "5H 값이 있으면 쓰고, 없을 때도 있다"): `5H` 는 입력에 5시간 창의 비율이 있으면 **어느 수준이든** 보여요. 개정 1 은 오늘의 조건(70% 넘게 썼을 때만)을 남겨 두었어요. 호스트가 5시간 창을 항상 보내지는 않아서, 창이 없거나 창에 비율이 없으면 `5H` 항목은 없어요(`5H n/a` 는 쓰지 않아요).

## Requirements

- REQ-1: The status line shall show each known window as `<battery> <label> <n>%` where `<n>` is 100 minus the whole used percentage · failure: `CW n/a` or `7D n/a` with no battery icon when the percentage is unknown
- REQ-2: The status line shall pick `🪫` when less than 30 percent is left and `🔋` otherwise · failure: none — 30 is 100 minus the high-use percentage 70 so the icon marks the same state as before
- REQ-3: When the stdin carries `rate_limits.five_hour` with a percentage, the status line shall show `<battery> 5H <n>%` followed by its reset countdown between the context window and the seven-day window · failure: none — an absent window or one without a percentage or one whose reset time has passed is left out with its separator
- REQ-4: The status line shall keep the hide options `context` and `usage_5h` and `usage_7d` and the countdowns and the version items and the width forms as they were · failure: none — the narrowest form drops the five-hour window while the context window and the seven-day window stay

## Acceptance

- AC-1 (REQ-1, REQ-3): 18% 를 쓴 컨텍스트 창은 `🔋 CW 82%`, 75% 를 쓴 5시간 창은 `🪫 5H 25%`, 95% 를 쓴 주간 창은 `🪫 7D 5%` 로 나오고, 5시간 창이 없는 입력에는 `5H` 항목이 없어요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_row_shows_what_is_left_of_each_window`
- AC-2 (REQ-1, REQ-4): 알 수 없는 창은 아이콘 없이 `CW n/a`·`7D n/a` 이고, `context`·`usage_5h`·`usage_7d` 가 각자의 항목을 줄에서 숨겨요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::unknown_windows_and_the_hide_list_stay_as_they_were`
- AC-3 (REQ-2): 경계가 `CW`·`5H`·`7D` 셋 다 고정돼요 — 70% 사용(30% 남음)은 `🔋`, 71% 사용(29% 남음)은 `🪫`. 기대값은 코드의 상수에서 나와요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_battery_icon_turns_at_the_low_left_boundary`
- AC-4 (REQ-3): 25% 를 쓴 5시간 창은 `🔋 5H 75%` 로 카운트다운과 함께 보이고, 리셋 시각이 없으면 카운트다운 없이 보이고, 5시간 창이 없거나 창에 비율이 없으면 항목과 구분자가 없어요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_five_hour_window_shows_whenever_the_input_carries_it`
- AC-5 (REQ-4): 폭이 좁아지면 Claude Code 버전, MoAI 버전, 카운트다운, 그다음 `5H`(와 `💰`) 순서로 빠지고 `CW`·`7D` 는 끝까지 남아요 · verify: `cargo test --offline --locked --test statusline_cards line_three_ladder_`

## Scope

- write: `.moai/specs/SPEC-20261122-statusline-limits-left/**`, `src/statusline/render.rs`, `src/statusline/input.rs`, `tests/statusline_*.rs`, `tests/fixtures/statusline-claude/**`, `docs/rust-product-path.md`

## Plan

1. `render.rs` 의 테스트를 남은 비율로 고치고 새 테스트를 더해 red 를 확인해요.
2. `battery()` 가 남은 비율을 쓰고 남은 양으로 아이콘을 고르게 해요.
3. (개정 2) `5H` 의 "70% 넘게 썼을 때만" 조건을 빼요. 테스트 먼저, red 확인 뒤 코드.
4. 통합 테스트, 골든 파일, 제품 문서를 맞춰요.

## Result

구현했어요(2026-11-23, 브랜치 `spec/20261122-statusline-limits-left`, 커밋 둘). 감사는 아직이에요.

- 첫 커밋(남은 비율과 아이콘): red 는 `cargo nextest run --lib -E 'test(statusline::render)'` exit 100, 67개 중 25개가 "남은 비율" 단언에서 실패. 경계 뮤턴트 하나(`left < LOW_LEFT_PERCENT` 를 `<=` 로)는 `window_tests` 8개 중 2개를 실패시켰어요.
- 둘째 커밋(`5H` 를 어느 수준이든): red 는 `cargo nextest run --lib -E 'test(statusline::)'` exit 100, 210개 중 7개가 새 `5H` 기대값에서 실패.
- green(둘째 커밋의 트리): `cargo nextest run --locked --offline --no-fail-fast --lib` 와 `--test` 21개(statusline 10개, `status_line_cli`, help·scan 이름의 타깃)가 exit 0, 1941 passed 0 failed 1 skipped.

- AC-1: met — `statusline::render::window_tests::the_row_shows_what_is_left_of_each_window` 가 `🔋 CW 82% | 🪫 5H 25% (1h 10m) | 🪫 7D 5% (4d 2h)` 와, 5시간 창이 없는 입력의 `🔋 CW 82% | 🪫 7D 5% (4d 2h)` 를 고정해요. 손으로 렌더한 줄도 같아요.
- AC-2: met — `statusline::render::window_tests::unknown_windows_and_the_hide_list_stay_as_they_were` 가 `CW n/a | 7D n/a` 와 `context`·`usage_5h`·`usage_7d` 숨김을 고정해요.
- AC-3: met — `statusline::render::window_tests::the_battery_icon_turns_at_the_low_left_boundary` 가 `HIGH_USE_PERCENT`(70 사용, 30 남음)에서 `🔋`, 그 다음 값(71 사용, 29 남음)에서 `🪫` 를 `CW`·`5H`·`7D` 셋 다 고정해요. 기대값은 `render.rs` 의 두 상수에서 나와요.
- AC-4: met — `statusline::render::window_tests::the_five_hour_window_shows_whenever_the_input_carries_it` 가 25% 사용의 `🔋 5H 75% (2h 10m)`, 리셋 시각 없는 `🔋 5H 75%`, 그리고 창 없음·빈 창·비율 없는 창·`null` 창의 "항목 없음"을 고정해요. `tests/statusline_cards.rs` 의 `line_three_items_subscriber_with_a_five_hour_window_at_any_level` 과 `line_three_items_subscriber_without_the_five_hour_window` 가 바이너리로 같은 것을 확인해요.
- AC-5: met — `tests/statusline_cards.rs` 의 `line_three_ladder_every_form_is_the_string_of_the_layouts_section` 이 `🪫 CW 26% | 🪫 5H 15% | 🔋 7D 79%`(Z3) 다음에 `🪫 CW 26% | 🔋 7D 79%`(Z4)로 `5H` 가 먼저 빠지는 것을 경계 폭에서 고정해요. 폭 사다리 코드는 바꾸지 않았어요.

카운트다운: `5H` 는 `7D` 와 같은 규칙이에요 — 입력에 리셋 시각이 있으면 언제나 보이고(남은 양과 무관), 31일보다 먼 시각은 카운트다운 없이, 폭 Z3 부터는 두 창 모두 카운트다운이 빠져요.

골든 파일: `python3 -I tools/statusline_golden.py --bin target/debug/moai --write` 뒤 `--check` 가 114 outputs ok. 첫 커밋에서 바뀐 plain 줄은 숫자가 100 에서 뺀 값으로 바뀐 것뿐이에요. 둘째 커밋에서는 `d01`–`d04` 네 케이스에 `🔋 5H 77% (3h 3m)` 가 더해졌어요(23.5% 사용, 리셋이 렌더 시각 뒤). 같은 5시간 창을 가진 `a02` 등 세 케이스는 리셋 시각이 렌더 시각보다 앞이라 예전 규칙대로 항목이 없고, `a16`(비율 없는 5시간 창)은 `🔋 CW 68% | 🗿 4.2647.0` 로 `5H` 항목이 없어요. `cases.json` 은 도구가 적는 바이너리 sha256 과 `a16` 의 설명 한 줄(`-> 0%` 를 `-> no 5H item` 으로)만 바뀌었어요.

남은 것: 제품 문서(`docs/rust-product-path.md`)의 3번 항목은 숫자·아이콘·`5H` 문장만 고쳤어요. 그 항목의 `(resets in …)` 표기와 줄 배치 설명은 이 변경 전부터 코드와 달랐고(`SPEC-20261122-statusline-limits` 계획 6단계), `tests/statusline_numbers_only.rs` 가 `resets in` 을 요구해서 여기서는 건드리지 않았어요.
