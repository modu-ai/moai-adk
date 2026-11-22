---
id: SPEC-20261122-statusline-limits-left
title: "The limit row shows what is left of the context window and the week"
status: draft
revision: 3
created: 2026-11-22
---

# 상태줄 한도 줄: 쓴 비율이 아니라 남은 비율

## Intent

사용자 지시예요. 상태줄 3번째 줄(한도)의 `CW`·`5H`·`7D` 숫자는 쓴 비율이 아니라 **남은 비율**이에요. 배터리 아이콘은 남은 양으로 골라요: 넉넉하면 `🔋`, 얼마 안 남으면 `🪫`. 줄의 순서는 `CW | 5H(있을 때) | 7D | 버전` 이에요. `SPEC-20261122-statusline-limits` 의 REQ-2·REQ-3 이 적은 숫자를 "남은 비율"로 바꿔 읽고, REQ-4 와 D-2 의 "`5H` 는 70% 를 넘을 때만"을 이 SPEC 이 바꿔요. 그 밖의 것(카운트다운, `💰`, 버전, 폭에 따른 모양, 숨김 옵션, 사용량 기록)은 그대로예요.

개정 2(사용자 말: "5H 값이 있으면 쓰고, 없을 때도 있다"): `5H` 는 입력에 5시간 창의 비율이 있으면 **어느 수준이든** 보여요. 개정 1 은 오늘의 조건(70% 넘게 썼을 때만)을 남겨 두었어요. 호스트가 5시간 창을 항상 보내지는 않아서, 창이 없거나 창에 비율이 없으면 `5H` 항목은 없어요(`5H n/a` 는 쓰지 않아요).

개정 3(사용자 결정: 진행판·상태줄·레인 보고가 **하나의 아이콘 세트**를 써요). 흐름은 `🗂️` 카드가 추가돼 차례를 기다림 → `🟡` 진행 중 → `📨` 넘겨서 확인을 기다림 → `🔍` 확인(감사) 중 → `🔁` 확인 실패 뒤 수리 → `🛬` 확인 통과, 랜딩을 기다림 → `✅` 완료(랜딩 또는 닫힘)이고, 흐름 옆에 `🚧` 막힘, `🙋` 사용자의 선택 필요, `🔴` 위험, `❌` 실패·취소가 있어요. `🔵 🟢 ⬜ 👀 ⛔ ⏸️` 와 글자 코드는 빠져요. 아이콘과 글 사이는 어디서나 **한 칸**이에요(`🗂️` 뒤도). 이 SPEC 의 브랜치에서 같이 고치라는 지시라 새 SPEC 을 만들지 않고 이 SPEC 의 개정으로 담았어요(브랜치의 SPEC 을 훅이 이 id 로 읽어요). 상태줄 2번째 줄(카드 흐름)은 REQ-5~REQ-7 이에요. 원장이 구분하지 않는 단계(`🔍` 확인 중, `🔁` 수리)는 상태줄에 나오지 않아요.

## Requirements

- REQ-1: The status line shall show each known window as `<battery> <label> <n>%` where `<n>` is 100 minus the whole used percentage · failure: `CW n/a` or `7D n/a` with no battery icon when the percentage is unknown
- REQ-2: The status line shall pick `🪫` when less than 30 percent is left and `🔋` otherwise · failure: none — 30 is 100 minus the high-use percentage 70 so the icon marks the same state as before
- REQ-3: When the stdin carries `rate_limits.five_hour` with a percentage, the status line shall show `<battery> 5H <n>%` followed by its reset countdown between the context window and the seven-day window · failure: none — an absent window or one without a percentage or one whose reset time has passed is left out with its separator
- REQ-4: The status line shall keep the hide options `context` and `usage_5h` and `usage_7d` and the countdowns and the version items and the width forms as they were · failure: none — the narrowest form drops the five-hour window while the context window and the seven-day window stay
- REQ-5: The status line shall write exactly one space between every icon and its text on all three lines · failure: none — an icon written with a variation selector takes one space too and the width of a form is counted from the written text
- REQ-6: The status line shall write its flow line as `🎯 C-<n> <stage icon> | 🗂️ <ready> → 🟡 <working> → 📨 <submitted> → 🛬 <audited plus landed> → ✅ <done today> | 🚧 <blocked> 🙋 <hold>` without letter codes · failure: none — an item whose count is zero is left out with its arrow or separator and the drafts and the ended leases are not on the line
- REQ-7: When the width budget is smaller than the flow line, the status line shall drop the leftmost item of the flow first and keep the own card and `🚧 🙋` to the end · failure: none — when the notes cannot be read the line is `🗂️ <ready> 🟡 <working>` without an arrow and never a guess

## Acceptance

- AC-1 (REQ-1, REQ-3): 18% 를 쓴 컨텍스트 창은 `🔋 CW 82%`, 75% 를 쓴 5시간 창은 `🪫 5H 25%`, 95% 를 쓴 주간 창은 `🪫 7D 5%` 로 나오고, 5시간 창이 없는 입력에는 `5H` 항목이 없어요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_row_shows_what_is_left_of_each_window`
- AC-2 (REQ-1, REQ-4): 알 수 없는 창은 아이콘 없이 `CW n/a`·`7D n/a` 이고, `context`·`usage_5h`·`usage_7d` 가 각자의 항목을 줄에서 숨겨요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::unknown_windows_and_the_hide_list_stay_as_they_were`
- AC-3 (REQ-2): 경계가 `CW`·`5H`·`7D` 셋 다 고정돼요 — 70% 사용(30% 남음)은 `🔋`, 71% 사용(29% 남음)은 `🪫`. 기대값은 코드의 상수에서 나와요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_battery_icon_turns_at_the_low_left_boundary`
- AC-4 (REQ-3): 25% 를 쓴 5시간 창은 `🔋 5H 75%` 로 카운트다운과 함께 보이고, 리셋 시각이 없으면 카운트다운 없이 보이고, 5시간 창이 없거나 창에 비율이 없으면 항목과 구분자가 없어요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_five_hour_window_shows_whenever_the_input_carries_it`
- AC-5 (REQ-4): 폭이 좁아지면 Claude Code 버전, MoAI 버전, 카운트다운, 그다음 `5H`(와 `💰`) 순서로 빠지고 `CW`·`7D` 는 끝까지 남아요 · verify: `cargo test --offline --locked --test statusline_cards line_three_ladder_`
- AC-6 (REQ-5): 표의 모든 아이콘 뒤가 한 칸이고(`🗂️`·`🅱️`·`⬇️` 포함), 흐름 줄이 그릴 수 있는 모든 모양(여러 흐름을 폭 0~70 과 폭 없음으로 렌더)에서 아이콘 뒤가 정확히 한 칸이에요 · verify: `cargo test --offline --locked --lib statusline::render::flow_line_tests::every_rendered_form_has_one_space_after_every_icon`
- AC-7 (REQ-6): 흐름 줄이 `🎯 C-<n> 🟡 | 🗂️ 150 → 🟡 14 → 📨 11 → 🛬 3 → ✅ 19 | 🚧 3 🙋 7` 모양이고, 0 인 항목은 화살표·구분자와 함께 숨고, 초안과 리스가 끝난 카드는 줄에 없고, 내 카드는 아홉 단계마다 그 단계의 아이콘을 보여요 · verify: `cargo test --offline --locked --lib statusline::render::flow_line_tests::a_zero_count_hides_its_item`
- AC-8 (REQ-7): 폭이 한 칸 모자랄 때마다 `🗂️`, `🟡`, `📨`, `🛬`, `✅` 순서로 빠지고 내 카드와 `🚧 🙋` 는 끝까지 남고, 노트를 못 읽으면 `🗂️ <ready> 🟡 <working>` 한 모양이에요 · verify: `cargo test --offline --locked --lib statusline::render::flow_line_tests::every_form_is_exactly_its_string_and_changes_at_its_boundary`

## Scope

- write: `.moai/specs/SPEC-20261122-statusline-limits-left/**`, `src/statusline/render.rs`, `src/statusline/input.rs`, `src/statusline/state.rs`, `src/statusline/width.rs`, `tests/statusline_*.rs`, `tests/fixtures/statusline-claude/**`, `docs/rust-product-path.md`

## Plan

1. `render.rs` 의 테스트를 남은 비율로 고치고 새 테스트를 더해 red 를 확인해요.
2. `battery()` 가 남은 비율을 쓰고 남은 양으로 아이콘을 고르게 해요.
3. (개정 2) `5H` 의 "70% 넘게 썼을 때만" 조건을 빼요. 테스트 먼저, red 확인 뒤 코드.
4. 통합 테스트, 골든 파일, 제품 문서를 맞춰요.
5. (개정 3) 흐름 줄의 테스트를 새 아이콘 세트와 한 칸 간격으로 고쳐 red 를 확인한 뒤, `width::icon_text` 의 두 칸을 없애고 `flow_forms` 를 다시 써요. `Stage::letter()` 는 `Stage::icon()` 이 돼요.

## Result

구현했어요(2026-11-23, 브랜치 `spec/20261122-statusline-limits-left`). 감사는 아직이에요.

- 첫 커밋(남은 비율과 아이콘): red 는 `cargo nextest run --lib -E 'test(statusline::render)'` exit 100, 67개 중 25개가 "남은 비율" 단언에서 실패. 경계 뮤턴트 하나(`left < LOW_LEFT_PERCENT` 를 `<=` 로)는 `window_tests` 8개 중 2개를 실패시켰어요.
- 둘째 커밋(`5H` 를 어느 수준이든): red 는 `cargo nextest run --lib -E 'test(statusline::)'` exit 100, 210개 중 7개가 새 `5H` 기대값에서 실패.
- green(둘째 커밋의 트리): `cargo nextest run --locked --offline --no-fail-fast --lib` 와 `--test` 21개(statusline 10개, `status_line_cli`, help·scan 이름의 타깃)가 exit 0, 1941 passed 0 failed 1 skipped.

- AC-1: met — `statusline::render::window_tests::the_row_shows_what_is_left_of_each_window` 가 `🔋 CW 82% | 🪫 5H 25% (1h 10m) | 🪫 7D 5% (4d 2h)` 와, 5시간 창이 없는 입력의 `🔋 CW 82% | 🪫 7D 5% (4d 2h)` 를 고정해요. 손으로 렌더한 줄도 같아요.
- AC-2: met — `statusline::render::window_tests::unknown_windows_and_the_hide_list_stay_as_they_were` 가 `CW n/a | 7D n/a` 와 `context`·`usage_5h`·`usage_7d` 숨김을 고정해요.
- AC-3: met — `statusline::render::window_tests::the_battery_icon_turns_at_the_low_left_boundary` 가 `HIGH_USE_PERCENT`(70 사용, 30 남음)에서 `🔋`, 그 다음 값(71 사용, 29 남음)에서 `🪫` 를 `CW`·`5H`·`7D` 셋 다 고정해요. 기대값은 `render.rs` 의 두 상수에서 나와요.
- AC-4: met — `statusline::render::window_tests::the_five_hour_window_shows_whenever_the_input_carries_it` 가 25% 사용의 `🔋 5H 75% (2h 10m)`, 리셋 시각 없는 `🔋 5H 75%`, 그리고 창 없음·빈 창·비율 없는 창·`null` 창의 "항목 없음"을 고정해요. `tests/statusline_cards.rs` 의 `line_three_items_subscriber_with_a_five_hour_window_at_any_level` 과 `line_three_items_subscriber_without_the_five_hour_window` 가 바이너리로 같은 것을 확인해요.
- AC-5: met — `tests/statusline_cards.rs` 의 `line_three_ladder_every_form_is_the_string_of_the_layouts_section` 이 `🪫 CW 26% | 🪫 5H 15% | 🔋 7D 79%`(Z3) 다음에 `🪫 CW 26% | 🔋 7D 79%`(Z4)로 `5H` 가 먼저 빠지는 것을 경계 폭에서 고정해요. 폭 사다리 코드는 바꾸지 않았어요.
- AC-6: met — `statusline::width::tests::one_space_follows_every_icon` 이 표의 아이콘 22개 뒤 한 칸을, `statusline::render::flow_line_tests::every_rendered_form_has_one_space_after_every_icon` 이 여러 흐름을 폭 0~70 과 폭 없음으로 렌더한 모든 모양의 아이콘 뒤 한 칸(두 칸 없음, 글자 코드 없음)을 고정해요. 첫 줄은 `every_icon_of_the_first_line_is_followed_by_one_space`, 셋째 줄은 `tests/statusline_cards.rs` 의 line 3 아이콘 테스트가 확인해요.
- AC-7: met — `flow_line_tests` 의 `every_form_is_exactly_its_string_and_changes_at_its_boundary`(가장 넓은 모양 `🎯 C-<n> 🟡 | 🗂️ 150 → 🟡 14 → 📨 11 → 🛬 3 → ✅ 19 | 🚧 3 🙋 7`), `a_zero_count_hides_its_item`, `drafts_and_ended_leases_are_not_on_the_line`, `my_card_shows_the_icon_of_each_stage` 가 고정해요. 단계 아이콘은 `statusline::state::stage_tests_edges::every_stage_has_its_icon_from_the_one_icon_set` 이 고정해요: 초안·ready `🗂️`, working·리스 끝남 `🟡`, submitted `📨`, audited·landed `🛬`, blocked `🚧`, hold `🙋`.
- AC-8: met — 같은 `every_form_is_exactly_its_string_and_changes_at_its_boundary` 가 각 모양의 폭 N 에서 그 모양, N-1 에서 다음 모양(`🗂️`, `🟡`, `📨`, `🛬`, `✅` 순서로 빠짐)을 고정하고, `the_unsplit_flow_has_one_form_and_never_a_guess` 가 `🗂️ 150 🟡 15` 한 모양을 고정해요. 바이너리로는 `tests/statusline_cards.rs` 의 `width_ladder_changes_the_flow_line_exactly_at_each_boundary` 가 확인해요.

개정 3 의 red 와 green: red 는 `cargo nextest run --lib -E 'test(statusline::)'` exit 100, 211개 중 33개가 새 아이콘·한 칸 기대값에서 실패. green 은 `--lib` 와 `--test` 25개(앞의 21개에 `home_style_cli`, `home_style_ownership_cli`, `update_user_layer_cli`, `product_cli`)가 exit 0, 2072 passed 0 failed 1 skipped. 첫 줄의 가장 넓은 모양은 두 칸이 한 칸이 되면서 112 칸에서 111 칸이 됐어요(메인 체크아웃 86→84, 일반 프로젝트 77→76). 골든 파일은 도구로 다시 썼고(114 outputs ok), 바뀐 plain 줄은 `🅱️`·`⬇️` 뒤의 한 칸과 `b13`·`b14` 의 흐름 줄(`🗂️ 25 → 📨 2 → ✅ 12 | 🙋 5`, `🗂️ 25 → 📨 3 → ✅ 12 | 🙋 5`)뿐이에요.

카운트다운: `5H` 는 `7D` 와 같은 규칙이에요 — 입력에 리셋 시각이 있으면 언제나 보이고(남은 양과 무관), 31일보다 먼 시각은 카운트다운 없이, 폭 Z3 부터는 두 창 모두 카운트다운이 빠져요.

골든 파일: `python3 -I tools/statusline_golden.py --bin target/debug/moai --write` 뒤 `--check` 가 114 outputs ok. 첫 커밋에서 바뀐 plain 줄은 숫자가 100 에서 뺀 값으로 바뀐 것뿐이에요. 둘째 커밋에서는 `d01`–`d04` 네 케이스에 `🔋 5H 77% (3h 3m)` 가 더해졌어요(23.5% 사용, 리셋이 렌더 시각 뒤). 같은 5시간 창을 가진 `a02` 등 세 케이스는 리셋 시각이 렌더 시각보다 앞이라 예전 규칙대로 항목이 없고, `a16`(비율 없는 5시간 창)은 `🔋 CW 68% | 🗿 4.2647.0` 로 `5H` 항목이 없어요. `cases.json` 은 도구가 적는 바이너리 sha256 과 `a16` 의 설명 한 줄(`-> 0%` 를 `-> no 5H item` 으로)만 바뀌었어요.

남은 것: 제품 문서(`docs/rust-product-path.md`)의 3번 항목은 숫자·아이콘·`5H` 문장만 고쳤어요. 그 항목의 `(resets in …)` 표기와 줄 배치 설명은 이 변경 전부터 코드와 달랐고(`SPEC-20261122-statusline-limits` 계획 6단계), `tests/statusline_numbers_only.rs` 가 `resets in` 을 요구해서 여기서는 건드리지 않았어요.
