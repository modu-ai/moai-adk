---
id: SPEC-20261123-statusline-order-bars
title: "The status line puts the versions on the flow line and brings the limit bars back"
status: draft
revision: 1
created: 2026-11-23
parent: SPEC-20261122-statusline-limits-left
---

# 상태줄: 버전을 흐름 줄 앞으로, 한도 줄에 막대 그래프 다시

## Intent

사용자 지시예요(2026-11-23, 리더 세션, 원문): `🤖 2.1.330 | 🗿 4.2648.20 | 🗂️ 176 → 🟡 6 → 📨 13 → 🛬 8 → ✅ 6 | 🚧 7 🙋 9 | 🎯 C-287 -> 이 배열로 다시 변경`, `🔋 CW 73% | 🔋 5H 49% (43m) | 🔋 7D 87% (3d 22h) -> 그래프 다시 추가`.

바꾸는 것은 둘이에요.

1. **둘째 줄(흐름 줄)의 순서**: `🤖 호스트 버전 | 🗿 moai 버전 | 🗂️ 흐름 | 🚧 🙋 | 🎯 카드`. 지금은 `🎯 <카드> <단계> | 🗂️ 흐름 | 🚧 🙋` 이고 두 버전은 셋째 줄 끝에 있어요. 두 버전이 셋째 줄에서 빠지고 둘째 줄 앞으로 와요.
2. **셋째 줄(한도 줄)**: 🔋 `CW`·`5H`·`7D` 만 남기고(`💰` 는 한도 정보가 없을 때만 나오던 그대로예요), 각 창의 값 옆에 막대 그래프를 다시 그려요. 막대 모양은 지워지기 전(`b6e8e9c977` 의 부모 커밋) 것을 따라요: 10칸, 채운 칸 `█` 와 빈 칸 `░`, 채운 칸 수는 `퍼센트 × 10 / 100` 내림, 위치는 라벨과 숫자 사이(`🔋 CW ███████░░░ 73%`), 콜론 없음. 숫자(남은 비율), 되돌림 시간, 배터리 아이콘 경계는 그대로예요.

**이 지시가 바꾸는 이전 결정**: `SPEC-20261121-statusline-detached` 의 REQ-5("숫자만, 막대 없음", 2026-11-21 사용자의 마지막 결정)는 이 SPEC 의 REQ-3 로 **대체**돼요. 그 SPEC 은 고치지 않고 이 SPEC 이 대체 관계를 적어요(REQ-7, 7 은 그 SPEC 의 REQ-7 과 무관). `SPEC-20261122-statusline-limits` 의 줄 3 REQ-1(항목 순서에 두 버전 포함)과 `SPEC-20261122-statusline-limits-left` 의 REQ-4(폭 사다리에서 버전이 가장 먼저 빠짐)도 이 SPEC 의 REQ-1·REQ-2·REQ-5 로 바뀐 부분만 대체돼요. 그 밖의 규칙(숫자는 남은 비율, `n/a`, 숨김 옵션, 한 칸 간격, 카운트다운 모양, 폭 예산)은 그대로예요.

## Non-goals

- 첫 줄(누구이고 어디인지)과 카드 단계·흐름의 계산. 흐름 줄의 항목과 아이콘은 그대로예요.
- 막대의 그라데이션 색. 이전 막대는 채운 칸마다 초록→빨강 색을 칠했는데, 지금 막대는 **남은** 양을 그리므로 가득 찬 막대가 빨강으로 끝나면 뜻이 거꾸로예요(D-3). 색은 이 SPEC 이 새로 정한 두 가지뿐이에요.
- 제품 문서의 상태줄 설명 전체를 다시 쓰는 일. 이 변경이 거짓으로 만드는 문장만 고쳐요.

## Requirements

- REQ-1: The status line shall write its second line as the Claude Code version, the MoAI version, the card flow, the blocked and hold counts and the held card in this order · failure: none — an item without its input is left out with its separator, the line holds the versions alone when there is no flow to show and is left out only when nothing is left
- REQ-2: While the width budget is smaller than the second line, the status line shall drop the Claude Code version first, then the MoAI version with its update arrow, then the leftmost item of the flow, and keep the held card and the blocked and hold counts in every form · failure: none — the narrowest form is written when none fits
- REQ-3: The status line shall write each known limit window as `<battery> <label> <bar> <n>%` where `<n>` is the share left, the bar has ten cells, the first `<n>` times ten divided by one hundred (rounded down) cells are `█` and the others `░` · failure: none — the percentage is clamped to 0 to 100 so the bar never has fewer than zero or more than ten filled cells
- REQ-4 (REQ-3): IF the percentage of a window is unknown, THEN the status line shall write `<label> n/a` with no battery icon and no bar · failure: none — the screen never shows an empty placeholder bar or a guessed number
- REQ-5: While the width budget is smaller than the third line, the status line shall drop the reset countdowns first, then the bars, then the five-hour window and the cost, and keep `CW` and `7D` in every form · failure: none — the narrowest form is written when none fits
- REQ-6: The status line shall paint the filled cells of a bar in the base color of the line and the empty cells in the muted color so that the colored line minus its escapes equals the plain line · failure: none — without color the bar is the same text
- REQ-7: The status line shall keep the numbers as the share left, the battery icon boundary, the reset countdown forms, the `context`, `usage_5h`, `usage_7d`, `claude_version`, `moai_version` and `cards` options and one space after every icon as they were · failure: none

## Acceptance

- AC-1 (REQ-1): 흐름 줄이 `🤖 2.1.330 | 🗿 4.2648.20 | 🗂️ 176 → 🟡 6 → 📨 13 → 🛬 8 → ✅ 6 | 🚧 7 🙋 9 | 🎯 C-12 🟡` 모양이고, 흐름이 없으면 두 버전만, 버전이 없으면 흐름만 나오고, `claude_version`·`moai_version`·`cards` 옵션이 각자의 항목을 숨겨요 · verify: `cargo test --offline --locked --lib statusline::render::flow_line_tests::the_versions_open_the_flow_line_and_my_card_closes_it`
- AC-2 (REQ-2): 폭이 한 칸 모자랄 때마다 `🤖`, `🗿`(업데이트 화살표와 함께), `🗂️`, `🟡`, `📨`, `🛬`, `✅` 순서로 빠지고 내 카드와 `🚧 🙋` 는 끝까지 남아요. 모든 모양이 아래 Layouts 표의 문자열 그대로예요 · verify: `cargo test --offline --locked --lib statusline::render::flow_line_tests::every_form_is_exactly_its_string_and_changes_at_its_boundary`
- AC-3 (REQ-3): 막대 칸 수가 남은 비율 0·9·10·11·19·20·99·100 에서 0·0·1·1·1·2·9·10 이고(9→10 경계에서 한 칸이 생기고 99→100 에서 열 칸이 차요), 100 을 넘거나 0 보다 작은 입력도 0..10 칸이고, `🔋` 와 `🪫` 의 70/71 경계가 그대로예요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_bar_has_ten_cells_and_changes_at_each_tenth`
- AC-4 (REQ-3): 사용자의 예 — 27%·51%·13% 사용 — 가 `🔋 CW ███████░░░ 73% | 🔋 5H ████░░░░░░ 49% (43m) | 🔋 7D ████████░░ 87% (3d 22h)` 로 나와요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_row_shows_what_is_left_of_each_window`
- AC-5 (REQ-4): 알 수 없는 `CW`·`7D` 는 `CW n/a`·`7D n/a`, 아이콘과 막대 글자(`█`·`░`) 가 없고, 퍼센트가 없는 주간 창도 같아요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::unknown_windows_and_the_hide_list_stay_as_they_were`
- AC-6 (REQ-5): 한도 줄의 모든 모양이 Layouts 표의 문자열 그대로이고 경계 폭 N 에서 그 모양, N-1 에서 다음 모양이에요 · verify: `cargo test --offline --locked --test statusline_cards line_three_ladder_`
- AC-7 (REQ-6): 색 있는 줄에서 이스케이프를 뺀 글이 색 없는 줄과 같고, 채운 칸과 빈 칸이 서로 다른 색 구간이에요 · verify: `cargo test --offline --locked --lib statusline::render::window_tests::the_colored_bar_has_the_plain_text`
- AC-8 (REQ-1, REQ-3, REQ-5): 실제 바이너리의 출력이 골든 사례 전부(`COLUMNS` 넓음·좁음 포함)에서 고정된 문자열이에요 · verify: `cargo test --offline --locked --test statusline_golden`
- AC-9 (REQ-7): 숫자만 보여 주던 옛 단언(`statusline_numbers_only` 의 "막대 낱말 없음" 세 테스트)은 이 SPEC 의 막대 단언으로 바뀌고, 제품 문서가 새 줄 순서와 막대를 말해요 · verify: `cargo test --offline --locked --test statusline_numbers_only`

### Layouts

네 줄 모두 빈 값은 항목이 구분자와 함께 빠져요. 폭은 `unicode-width` 로 세요(테스트가 `display_width` 로 경계를 정해요). 값은 고정 입력이에요.

흐름 줄(보유 카드 `C-12`, 단계 `🟡`, 사용자의 예):

| 모양 | 글 |
|---|---|
| F0 가장 넓음 | `🤖 2.1.330 \| 🗿 4.2648.20 \| 🗂️ 176 → 🟡 6 → 📨 13 → 🛬 8 → ✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F1 호스트 버전 없음 | `🗿 4.2648.20 \| 🗂️ 176 → 🟡 6 → 📨 13 → 🛬 8 → ✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F2 버전 없음 | `🗂️ 176 → 🟡 6 → 📨 13 → 🛬 8 → ✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F3 | `🟡 6 → 📨 13 → 🛬 8 → ✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F4 | `📨 13 → 🛬 8 → ✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F5 | `🛬 8 → ✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F6 | `✅ 6 \| 🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| F7 가장 좁음 | `🚧 7 🙋 9 \| 🎯 C-12 🟡` |
| 흐름이 없음 | `🤖 2.1.330 \| 🗿 4.2648.20` (좁으면 `🗿 4.2648.20`, 더 좁아도 마지막 모양) |
| 노트를 못 읽음(나누지 못한 흐름) | `🤖 2.1.330 \| 🗿 4.2648.20 \| 🗂️ 150 🟡 15` (모양은 버전만 줄어요) |

한도 줄(사용자의 예, 사용 27%·51%·13%):

| 모양 | 글 |
|---|---|
| Z0 가장 넓음 | `🔋 CW ███████░░░ 73% \| 🔋 5H ████░░░░░░ 49% (43m) \| 🔋 7D ████████░░ 87% (3d 22h)` |
| Z1 카운트다운 없음 | `🔋 CW ███████░░░ 73% \| 🔋 5H ████░░░░░░ 49% \| 🔋 7D ████████░░ 87%` |
| Z2 막대 없음 | `🔋 CW 73% \| 🔋 5H 49% \| 🔋 7D 87%` |
| Z3 가장 좁음 | `🔋 CW 73% \| 🔋 7D 87%` |
| 한도 정보가 없는 사용자(CW 사용 41%) | Z0·Z1 `🔋 CW █████░░░░░ 59% \| 💰 3.87`, Z2 `🔋 CW 59% \| 💰 3.87`, Z3 `🔋 CW 59%` |
| 알 수 없는 창 | `CW n/a \| 7D n/a` (모든 모양에서 막대 없음) |

## Scope

- write: `.moai/specs/SPEC-20261123-statusline-order-bars/**`, `src/statusline/**`, `tests/statusline*.rs`, `tests/fixtures/statusline/**`, `tests/fixtures/statusline-claude/**`, `docs/rust-product-path.md`
- read-only: `tools/statusline_golden.py` (골든 파일을 `--write` 로 다시 쓰고 `--check` 로 확인하는 데만 실행해요), 이 SPEC 이 말하는 이전 SPEC 들
- forbidden: `.claude/worktrees/**`(다른 레인의 작업 폴더), `.moai-cli/**`

## Plan

1. 테스트를 먼저 고쳐 red 를 봐요: 흐름 줄 순서(`flow_line_tests`), 막대 칸 수와 경계(`window_tests`), 알 수 없는 창, 한도 줄 사다리(`statusline_cards`), `statusline_numbers_only` 의 옛 단언.
2. `render.rs`: `Piece` 에 "앞 조각에 붙임" 표지를 더해 막대의 채운 칸과 빈 칸을 공백 없이 다른 색으로 이어요. `flow_forms` 는 두 버전을 앞에 두고 내 카드를 뒤로 보내며 버전부터 빠지는 사다리를 가져요. `line_three_forms` 는 버전을 빼고 막대 단계를 카운트다운과 5H 사이에 둬요.
3. 골든 파일을 도구로 다시 쓰고(`tools/statusline_golden.py --write`) 바뀐 plain 줄을 하나씩 읽어요. 검증은 Rust 테스트 `statusline_golden` 이 해요.
4. 제품 문서의 줄 순서와 막대 문장을 고치고 `statusline_numbers_only` 의 문서 테스트를 새 문장에 맞춰요.
5. 실제 바이너리 출력(넓은 폭·좁은 폭)을 Result 에 붙여요.

## Risks

1. **`█`·`░` 의 폭.** 이 문자들은 동아시아 모호 폭(East Asian Ambiguous)이에요. `unicode-width` 는 1칸으로 세는데, 한국어 로케일의 일부 터미널은 2칸으로 그려요. 그러면 막대 줄이 센 것보다 넓어져 호스트가 줄을 자르거나 감쌀 수 있어요. 지워지기 전에도 같은 글자였고 사용자가 다시 요청했으니 그대로 쓰되, 문제가 보이면 폭이 좁은 모양(막대 없음)으로 내려가는 사다리가 안전판이에요.
2. **막대의 방향.** 숫자가 남은 비율이라서 막대도 남은 비율을 그려요(가득 = 넉넉). 지워지기 전 막대는 쓴 비율이었어요. 숫자와 막대가 같은 양을 가리키게 한 것이 이 SPEC 의 선택이고 D-1 에서 되돌릴 수 있어요.
3. **골든 파일 다수가 바뀌어요.** 두 버전의 이동과 막대로 plain 줄이 거의 모든 사례에서 바뀌어요. 바뀐 줄을 하나씩 읽고, 의도하지 않은 변화가 없는지(첫 줄과 카운트다운 숫자) 확인해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 막대는 숫자와 같은 양, 곧 남은 비율을 그려요(채운 칸 = 남은 비율 × 10 / 100 내림) | 사용자: "숫자와 되돌림 시간은 그대로". 숫자가 남은 비율이라 쓴 비율 막대는 같은 줄에서 서로 어긋나요. 리더 결정 범위 안의 되돌릴 수 있는 선택(레인이 정하고 READY 에 적어요) | 채운 칸 계산에 `100 - left` 를 쓰는 한 줄 |
| D-2 | 내 카드는 단계 아이콘을 그대로 달고 `🎯 C-<n> <단계>` 로 맨 뒤에 가요 | 사용자의 예 `🎯 C-287` 에는 단계 아이콘이 없지만 카드는 순서만 바꾸라고 했고 단계 아이콘을 빼라는 말은 없었어요 | `MyCard` 조각에서 `stage.icon()` 을 빼는 한 줄 |
| D-3 | 막대 색은 채운 칸 기본 색, 빈 칸 흐린 색 두 가지뿐이고 그라데이션은 되살리지 않아요 | 남은 양을 그리는 막대가 초록에서 빨강으로 끝나면 뜻이 거꾸로예요. 색만이 신호가 되지 않게 `█`/`░` 모양이 이미 구분해요 | 테마에 `stops` 를 되돌리고 칸마다 색을 칠하기 |
| D-4 | 한도 줄의 사다리는 카운트다운 → 막대 → `5H`·`💰` 순서로 빠져요 | 지금 규칙(카운트다운이 먼저, 창은 나중)을 지키고 막대는 숫자를 꾸미는 장식이라 숫자가 있는 창보다 먼저 빠져요. 사용자 말: "좁은 폭에서 줄이는 순서는 지금 규칙을 따라요" | 막대 단계를 사다리의 다른 칸으로 옮기기 |
| D-5 | 흐름 줄의 사다리는 호스트 버전 → moai 버전 → 흐름의 왼쪽 항목 순서로 빠져요 | 두 버전은 지금도 사다리에서 가장 먼저 빠지는 항목이에요(`limits` SPEC 의 Z1·Z2) | 버전 조각을 사다리 끝으로 옮기기 |
| D-6 | 흐름이 없어도 버전은 둘째 줄에 남아요 | 버전은 원장과 무관한 항목이고 지금도 원장 없이 나와요 | 버전을 흐름이 있을 때만 쓰게 하기 |
| D-7 | `💰` 는 지금 규칙(한도 정보가 없을 때만, 셋째 줄)대로 그 줄에 남아요 | 카드가 한도 줄에서 빼라고 한 것은 두 버전뿐이에요 | `💰` 를 둘째 줄로 옮기기 |

## Result

(구현 뒤 SYNC 에서 적어요.)
