---
id: SPEC-20261122-statusline-limits-left
title: "The limit row shows what is left of the context window and the week"
status: draft
revision: 1
created: 2026-11-22
---

# 상태줄 한도 줄: 쓴 비율이 아니라 남은 비율

## Intent

사용자 지시예요. 상태줄 3번째 줄(한도)의 `CW`·`5H`·`7D` 숫자는 쓴 비율이 아니라 **남은 비율**이에요. 배터리 아이콘은 남은 양으로 골라요: 넉넉하면 `🔋`, 얼마 안 남으면 `🪫`. 줄의 순서는 `CW | 5H(있을 때) | 7D | 버전` 그대로예요. `SPEC-20261122-statusline-limits` 의 REQ-2·REQ-3·REQ-4 가 적은 숫자와 D-2 의 기준을 이 SPEC 이 "남은 비율"로 바꿔 읽어요. 그 밖의 것(카운트다운, `💰`, 버전, 폭에 따른 모양, 숨김 옵션, 사용량 기록)은 그대로예요.

`5H` 가 보이는 조건은 오늘과 같은 실제 상태예요: 입력에 5시간 창이 있고 70% 를 넘게 썼을 때, 곧 남은 것이 30% 미만일 때만 보여요(사용자 정정: 5H 는 빼지 않고 오늘처럼 조건부로 둔다).

## Requirements

- REQ-1: The status line shall show each known window as `<battery> <label> <n>%` where `<n>` is 100 minus the whole used percentage · failure: `CW n/a` or `7D n/a` with no battery icon when the percentage is unknown
- REQ-2: The status line shall pick `🪫` when less than 30 percent is left and `🔋` otherwise · failure: none — 30 is 100 minus the high-use percentage 70 so the icon marks the same state as before
- REQ-3: When the stdin carries `rate_limits.five_hour` with less than 30 percent left, the status line shall show `🪫 5H <n>%` followed by its reset countdown between the context window and the seven-day window · failure: none — an absent window or one with 30 percent or more left is left out with its separator
- REQ-4: The status line shall keep the hide options `context` and `usage_5h` and `usage_7d` and the countdowns and the version items as they were · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-3): 18% 를 쓴 컨텍스트 창은 `🔋 CW 82%`, 75% 를 쓴 5시간 창은 `🪫 5H 25%`, 95% 를 쓴 주간 창은 `🪫 7D 5%` 로 나오고, 5시간 창이 없는 입력에는 `5H` 항목이 없어요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_row_shows_what_is_left_of_each_window`
- AC-2 (REQ-1, REQ-4): 알 수 없는 창은 아이콘 없이 `CW n/a`·`7D n/a` 이고, `context`·`usage_5h`·`usage_7d` 가 각자의 항목을 줄에서 숨겨요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::unknown_windows_and_the_hide_list_stay_as_they_were`
- AC-3 (REQ-2, REQ-3): 경계가 고정돼요 — 70% 사용(30% 남음)은 `🔋` 이고 `5H` 는 숨고, 71% 사용(29% 남음)은 `🪫` 이고 `5H` 가 보여요. 기대값은 코드의 상수에서 나와요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_battery_icon_turns_at_the_low_left_boundary`

## Scope

- write: `.moai/specs/SPEC-20261122-statusline-limits-left/**`, `src/statusline/render.rs`, `src/statusline/input.rs`, `tests/statusline_*.rs`, `tests/fixtures/statusline-claude/**`, `docs/rust-product-path.md`

## Plan

1. `render.rs` 의 테스트를 남은 비율로 고치고 새 테스트 셋을 더해 red 를 확인해요.
2. `battery()` 가 남은 비율을 쓰고 남은 양으로 아이콘을 고르게 하고, `5H` 조건을 남은 양으로 적어요.
3. 통합 테스트, 골든 파일, 제품 문서를 "남은 비율"로 맞춰요.

## Result

구현했어요(2026-11-23, 브랜치 `spec/20261122-statusline-limits-left`). 감사는 아직이에요. red: 동작을 바꾸기 전 `cargo nextest run --lib -E 'test(statusline::render)'` 가 exit 100, 67개 중 25개가 "남은 비율" 단언에서 실패했어요. green: `cargo nextest run --locked --offline --no-fail-fast --lib` 와 `--test` 21개(statusline 10개, `status_line_cli`, help·scan 이름의 타깃)가 exit 0, 1940 passed 0 failed 1 skipped.

- AC-1: met — `statusline::render::window_tests::the_row_shows_what_is_left_of_each_window` 가 `🔋 CW 82% | 🪫 5H 25% (1h 10m) | 🪫 7D 5% (4d 2h)` 와, 5시간 창이 없는 입력의 `🔋 CW 82% | 🪫 7D 5% (4d 2h)` 를 고정해요. 손으로 한 번 렌더한 줄도 같아요.
- AC-2: met — `statusline::render::window_tests::unknown_windows_and_the_hide_list_stay_as_they_were` 가 `CW n/a | 7D n/a` 와 `context`·`usage_5h`·`usage_7d` 숨김을 고정해요.
- AC-3: met — `statusline::render::window_tests::the_battery_icon_turns_at_the_low_left_boundary` 가 `HIGH_USE_PERCENT`(70 사용, 30 남음)에서 `🔋`·`5H` 숨김, 그 다음 값(71 사용, 29 남음)에서 `🪫`·`5H` 표시를 `CW`·`5H`·`7D` 셋 다 고정해요. 기대값은 `render.rs` 의 두 상수에서 나와요.

골든 파일: `python3 -I tools/statusline_golden.py --bin target/debug/moai --write` 뒤 `--check` 가 114 outputs ok. 바뀐 plain 줄은 모두 숫자가 100 에서 뺀 값으로 바뀐 것뿐이고 아이콘은 그대로예요(예: `🔋 CW 8% | 🔋 7D 41% (4d 23h)` → `🔋 CW 92% | 🔋 7D 59% (4d 23h)`, `🪫 CW 93%` → `🪫 CW 7%`). `cases.json` 은 도구가 적는 바이너리 sha256 만 바뀌었어요.

남은 것: 제품 문서(`docs/rust-product-path.md`)의 3번 항목은 숫자·아이콘·`5H` 문장만 고쳤어요. 그 항목의 `(resets in …)` 표기와 줄 배치 설명은 이 변경 전부터 코드와 달랐고(`SPEC-20261122-statusline-limits` 계획 6단계), `tests/statusline_numbers_only.rs` 가 `resets in` 을 요구해서 여기서는 건드리지 않았어요.
