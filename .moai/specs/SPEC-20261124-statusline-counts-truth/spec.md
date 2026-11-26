---
id: SPEC-20261124-statusline-counts-truth
title: "The status line flow counts match the ledger and its limit windows show the share used"
status: draft
revision: 3
created: 2026-11-24
---

# 상태줄 흐름 줄의 숫자를 원장과 맞추고, 한도 항목은 쓴 양으로 보여요

## Intent

**개정 3 (2026-11-24, 카드 C-466).** C-452 감사(노트 8519)의 N1 을 닫아요. 요구사항, 기준, 범위는 그대로고 Supersedes 표만 고쳐요: 빠졌던 줄(flow REQ-1·AC-1·AC-3, order-bars AC-4, limits-left AC-7, bar-color REQ-1)을 더하고, ready-receipt-followup 줄이 `REPAIRED` 를 "세지 않는다"고 잘못 쓴 것을 바로잡아요(그 SPEC 의 REQ-4 는 `REPAIRED` 를 넘김 표식으로 이미 정해요).

**개정 2 (2026-11-24, 사용자 결정).** 사용자가 리더 세션에서 직접 정했어요(리더 노트 8394): 한도 줄 `CW ███████░░░ 75% | 🔋 5H ██████░░░░ 68% (16m) | 🪫 7D ██░░░░░░░░ 22% (2d 15h)` 의 "수치와 그래프를 반대로, 남아 있는 게 아닌 사용한 것으로 변경". 이 SPEC 이 카드 C-452 의 기준 5 로 받아서 같은 파일(`src/statusline/render.rs`)을 쥔 개수 수리와 한 개정에 묶어요. 개정 1 의 내용(결함 아홉과 REQ-1~11, D-1~D-7)은 그대로예요. 세부는 리더 판단(되돌릴 수 있음)으로 정해졌고 D-8~D-11 에 있어요: 숫자는 쓴 %, 막대는 쓴 만큼 채움, 배터리 아이콘은 지금과 같은 상태(쓴 양이 70% 를 넘으면 `🪫`), 막대 색은 같은 경계를 쓴 양으로 읽은 초록 → 노랑 → 주황 → 빨강, 리셋 시간 `(16m)` 은 그대로(남은 시간). 사용자의 예는 쓴 양으로 `🔋 CW ██░░░░░░░░ 25% | 🔋 5H ███░░░░░░░ 32% (16m) | 🪫 7D ███████░░░ 78% (2d 15h)` 로 나와요. 이 개정이 뒤집는 것은 `SPEC-20261122-statusline-limits-left` 의 REQ-1·REQ-2·REQ-3 의 "남은 몫", `SPEC-20261123-statusline-order-bars` 의 REQ-3·REQ-7 의 "숫자와 막대는 남은 비율", `SPEC-20261124-statusline-bar-color` 의 REQ-2·D-1 의 "색 단계는 남은 비율" 이에요(아래 Supersedes).

사용자 보고예요(2026-11-24, 카드 C-452). 상태줄이 `🗂️ 81 → 🟡 3 → 📨 9 → 🛬 6 → ✅ 50 | 🚧 8 🙋 96` 을 그렸고 사용자는 "숫자가 틀렸는지 확인하고 고치라"고 했어요. 리더가 read-only 조사로 같은 숫자를 코드에서 다시 만들고(NeXT `02746488f9`, 원장 2026-11-24T07:32Z) 결함 아홉을 찾았어요. 이 레인은 같은 모양의 원장을 임시 HOME 에 픽스처로 다시 만들어 일곱 숫자를 그대로 재현했어요(`tests/statusline_counts_truth.rs` 의 `the_0732z_ledger_shape_draws_the_seven_numbers_of_the_report`, 아래 표). 원장 자체는 읽지 않았고 부분 구성(각 묶음 안의 개수)은 숫자를 맞추려고 고른 모양이에요.

| # | 결함 | 위치 (NeXT `f5a5c43784`) | 이 SPEC 이 닫는 곳 |
|---|---|---|---|
| 1 | `🟡` 이 working 만 그려 stale 5장이 줄에서 빠져요(26 대 31) | `render.rs:670` | REQ-7 |
| 2 | 리더의 결정 노트가 `LANE-BLOCKED` 에 답해도 표식이 아니라서 `🚧` 로 남아요 | `state.rs:1980-2004` `marker_of` | REQ-5 |
| 3 | `REPAIRED` 를 넘김으로 읽지 않아요(`ready_receipt.rs:20-24` 는 읽어요) | `state.rs:1987-1990` | REQ-4 |
| 4 | 레인이 넘기고 반납하면 카드가 `active` 에서 `ready` 로 돌아가 `🗂️` 로 세요 | `state.rs:1661-1663`, `1716-1720`, `reservation.rs:1217-1222` | REQ-1, REQ-2, REQ-3 |
| 5 | `🙋` 이 hold 카드를 전부 세요(96장 모두 철회된 카드) | `state.rs:1720`, `render.rs:680` | REQ-8 |
| 6 | "오늘 닫은 수" 가 UTC 날이에요(50, 사용자의 서울 날로는 103) | `state.rs:1700-1701` (flow D-4) | REQ-9, REQ-10 |
| 7 | 넘김·감사·막힘 카드가 임대가 끝나도 그 단계로 남아요 | `state.rs:1884-1894` (flow D-2) | 바꾸지 않아요(D-4), 숫자가 사실이에요 |
| 8 | `card.landed` 이벤트가 언제 있었든 이겨요: 한 번 랜딩한 뒤 다시 연 카드가 `🛬` 로 보여요 | `state.rs:1664` | REQ-6 |
| 9 | 문서: `state.rs:206` 은 `🗂️` 에 초안이 든다고 하는데 코드와 limits-left REQ-6 은 빼요 | `state.rs:206` | Plan 5 |

### 재현한 원장 모양

한 ledger, 렌더 시계 2026-11-24T07:32:00Z(`--now-unix 1795505520`). 숫자는 지금 코드가 그린 값(재현 시험이 줄 전체로 고정)과, 이 SPEC 의 규칙이 그릴 값이에요.

| 묶음 (카드 수) | 원장 모양 | 지금 그리는 곳 | 규칙 뒤 |
|---|---|---|---|
| 평범한 ready 76 + 풀린 뒤 다시 take 한 1 | 마지막 take 뒤에 표식 없음 | 🗂️ 77 | 🗂️ 77 |
| 풀린 ready, 마지막 take 뒤 `LANE-READY-INTEGRATION` / `REPAIRED` | 2 | 🗂️ | 📨 2 |
| 풀린 ready, 마지막 take 뒤 `AUDIT … verdict: PASS` | 2 | 🗂️ | 🛬 2 |
| active, 임대 살아 있음, 표식 없음 | 2 | 🟡 | 🟡 |
| active, 임대 살아 있음, `REPAIRED` | 1 | 🟡 | 📨 |
| active, 임대 끝, 표식 없음 (stale) | 5 | 안 그려요 | 🟡 |
| review 5 + active 에 `LANE-READY-INTEGRATION` 4 | 9 (임대 끝 6) | 📨 9 | 📨 9 |
| active `AUDIT PASS` 3 + landed 2 | 5 (임대 끝) | 🛬 5 | 🛬 5 |
| 랜딩한 뒤 다시 take 한 active (새 `ACK`) | 1 | 🛬 | 🟡 |
| `LANE-BLOCKED` 가 가장 새 표식, 답 없음 | 4 (임대 끝) | 🚧 | 🚧 |
| `LANE-BLOCKED` 뒤에 리더의 결정 노트, 임대 살아 있음 | 4 | 🚧 | 🟡 |
| hold + `card.withdraw` | 96 | 🙋 96 | 안 그려요 |
| done: UTC 오늘 50 + 그 앞 53 | 103 | ✅ 50 | ✅ 50 (UTC), ✅ 103 (서울) |

그려지는 줄: 지금 `🗂️ 81 → 🟡 3 → 📨 9 → 🛬 6 → ✅ 50 | 🚧 8 🙋 96`, 규칙 뒤 서울에서 `🗂️ 77 → 🟡 12 → 📨 12 → 🛬 7 → ✅ 103 | 🚧 4`. `🟡 + 📨 + 🛬 + 🚧` 는 35 = active·review 31 + 풀린 뒤 넘김으로 읽히는 ready 4 예요(REQ-7).

## Non-goals

- 원장 쓰기, 새 상태, 새 명령, 새 표. `release` 가 카드를 `ready` 로 돌리는 것(`src/ledger/reservation.rs`)은 그대로 두고 상태줄이 표식으로 읽어요.
- 임대가 끝난 넘김·감사·막힘 카드를 줄에서 빼거나 다른 단계로 옮기는 것(결함 7): 그 카드는 실제로 그 단계에 있고 정리는 리더의 `card reclaim` 몫이에요.
- 유형 있는 표식 행(typed rows)으로 첫 줄 읽기를 바꾸는 것. 첫 줄 읽기는 계속 임시 규칙이고 `marker_of` 한 곳에만 단어를 더해요.
- 표식 단어와 노트의 일반 규칙을 다른 명령(`card inbox` 등)에 맞추는 것.

## Requirements

- REQ-1: When a ready card carries a submit marker written at or after its last reservation, the status line shall count it as submitted · failure: none — a note older than the last reservation changes nothing and the card stays ready
- REQ-2: When a ready card carries an audit PASS written at or after its last reservation, the status line shall count it as audited · failure: none — a card with no audit PASS after its last reservation stays ready
- REQ-3: When a card is active or review, the status line shall take its stage from the newest marker written at or after its last reservation · failure: none — a card with no such marker is working with a live lease, stale without one, and submitted when it is in review
- REQ-4: The status line shall read a note whose first line starts with `REPAIRED` as a submit marker.
- REQ-5: When a decision note signed by the leader is newer than a `LANE-BLOCKED` note of the same card, the status line shall treat that block as ended · failure: none — a decision note signed by anyone else, or older than the block, leaves the block standing
- REQ-6: When the last `card.landed` event of a card is older than its last reservation, the status line shall not place the card in the stage landed · failure: none — a card with no reservation row counts any landing
- REQ-7: The status line shall draw the working and the stale cards as one `🟡` item so that the drawn items add up to the open cards that are neither draft nor hold.
- REQ-8: When a hold card has a `card.withdraw` or `card.fold` event, the status line shall not count it as a card that waits for the user · failure: none — a hold card with no such event is counted
- REQ-9: The status line shall count as done today the cards closed at or after the start of the local day of the machine's time zone.
- REQ-10 (REQ-9): IF the time zone of the machine cannot be read, THEN the status line shall count the UTC day.
- REQ-11: The status line shall read the stages read-only in no more than four statements within the render budget.
- REQ-12: The status line shall write the number of each known limit window as its whole used percentage and fill the first used times ten divided by one hundred (rounded down) of the ten cells of its bar · failure: none — the percentage is clamped to 0 to 100 so the bar never has fewer than zero or more than ten filled cells
- REQ-13: When the used percentage of a limit window is above 70, the status line shall show `🪫` in place of `🔋` · failure: none — 70 itself is `🔋`, the state the icon already marked when less than 30 was left
- REQ-14: The status line shall paint the filled cells of a bar green up to 30 percent used, yellow up to 50, peach up to 70 and red above 70 · failure: none — without a color profile the line is the same bytes without an escape sequence

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): the ledger fixture of the 07:32Z shape draws `🗂️ 77 → 🟡 12 → 📨 12 → 🛬 7 → ✅ 50 | 🚧 4` in UTC, the drawn items add up to the open cards that are neither draft nor hold, and each of the seven rows of the table above lands in the item the table names · verify: `cargo nextest run --locked --offline --test statusline_counts_truth the_0732z_ledger_shape_draws_what_the_ledger_holds`
- AC-2 (REQ-9): the same fixture with `TZ=Asia/Seoul` for the child only draws `✅ 103` and with `TZ=UTC` draws `✅ 50`, and the count changes at the local midnight and not a second before · verify: `cargo nextest run --locked --offline --test statusline_counts_truth the_done_count_follows_the_local_day` and `cargo nextest run --locked --offline --lib statusline::state::done_today_tests`
- AC-3 (REQ-8): a hold card with a withdraw event adds nothing to `🙋`, a hold card without one is counted, and the item is hidden at 0; a mutant that counts every hold card fails the named test · verify: `cargo nextest run --locked --offline --test statusline_counts_truth a_withdrawn_hold_card_is_not_a_pick_the_user_owes`
- AC-4 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): each marker word at its edge — a submit note one second before the last take (stays ready), at the take (counts), after it; `REPAIRED` with and without its space; a leader decision note before, after and equal in order to the block; a note of kind `decision` signed by a lane is no answer · verify: `cargo nextest run --locked --offline --lib statusline::state::stage_tests_after_reserve`
- AC-5 (REQ-6): a card with one landing and a later take is working, a card with a landing after its last take is landed · verify: `cargo nextest run --locked --offline --lib statusline::state::stage_tests_after_reserve`
- AC-6 (REQ-10): without a readable zone the day is the UTC day and the count equals the one the earlier rule gave · verify: `cargo nextest run --locked --offline --lib statusline::state::done_today_tests`
- AC-7 (REQ-7, REQ-11): the unsplit form still never guesses, the four statements are counted by the existing test and the ledger bytes stay the same after a render · verify: `cargo nextest run --locked --offline --lib statusline::state` and `cargo nextest run --locked --offline --test statusline_flow`
- AC-9 (REQ-12): the user's example, 25%, 32% and 78% used, is `🔋 CW ██░░░░░░░░ 25% | 🔋 5H ███░░░░░░░ 32% (16m) | 🪫 7D ███████░░░ 78% (2d 15h)`; the filled cells are 0, 0, 1, 1, 1, 2, 9 and 10 at 0, 9, 10, 11, 19, 20, 99 and 100 used, 0 to 10 for an input below 0 or above 100, and the number is the used percentage at each of them · verify: `cargo nextest run --locked --offline --lib statusline::render::window_tests::the_row_shows_what_is_used_of_each_window` and `cargo nextest run --locked --offline --lib statusline::render::window_tests::the_bar_has_ten_cells_and_changes_at_each_tenth`
- AC-10 (REQ-13): at 70 used the icon is `🔋` and at 71 it is `🪫`, for `CW`, `5H` and `7D` alike, with the expected values read from `HIGH_USE_PERCENT` · verify: `cargo nextest run --locked --offline --lib statusline::render::window_tests::the_battery_icon_turns_above_the_high_use_boundary`
- AC-11 (REQ-14): the level color changes at each step in the two themes and the two profiles — 30 green, 31 yellow, 50 yellow, 51 peach, 70 peach, 71 red used — and never goes back from red toward green as the used share rises, and the empty cells and the head and the number keep their colors · verify: `cargo nextest run --locked --offline --lib statusline::render::window_tests::the_level_color_changes_at_each_step_point` and `cargo nextest run --locked --offline --lib statusline::render::window_tests::the_filled_cells_carry_the_level_color`
- AC-12 (REQ-12, REQ-13, REQ-14): the real binary writes the used numbers in the line ladder and in every golden case, the colored goldens carry the new bar colors, the plain goldens differ from the earlier ones in the limit numbers and bars only, and the product document says "used" in place of "left" · verify: `cargo nextest run --locked --offline --test statusline_cards --test statusline_golden --test statusline_numbers_only`
- AC-8: the goldens and the statusline targets pass with the changed lines rewritten and read, and `moai check affected` has a PASS receipt for the head · verify: `moai check affected <base>..<head>`

## Scope

- write: `.moai/specs/SPEC-20261124-statusline-counts-truth/**`, `src/statusline/state.rs`, `src/statusline/render.rs`, `tests/statusline_counts_truth.rs`
- write, as the changed lines need (the leader widened the card's scope, note 8366): `tests/statusline_flow.rs`, `tests/statusline_cards.rs`, `tests/fixtures/statusline-claude/golden/**`, `tests/fixtures/statusline-claude/cases.json` (the checksum line the golden tool writes), the superseded lines of `.moai/specs/SPEC-20261122-statusline-flow/**` and `.moai/specs/SPEC-20261122-statusline-limits-left/**` (see Supersedes)
- write, widened by the leader for the used numbers (notes 8420, 8425, 8427): `tests/statusline_numbers_only.rs`, `src/statusline/input.rs` (its test module only) and `tests/statusline_counts_refresh.rs` (existing expectations only, no reorder: card C-175 adds a test there and the leader merges the two at landing)
- write by line lease (card C-449 holds the file): `docs/rust-product-path.md`, the flow-line item and the limit-row item (3) only
- read-only: `src/ledger/**`, `src/card/ready_receipt.rs`
- forbidden: the real HOME and ledger, the network in tests, any write to a ledger from the status line, a new crate, `TZ` set on the test process (the child only)

## Plan

1. Red first. Tests of AC-1 to AC-6 in the order of the table; each fails on the current head for the reason the table gives. The reproduction tests already in the file become the AC-1 and AC-2 tests by changing their expected lines.
2. `state.rs`: statement 1 tells a withdrawn hold card from another; statement 2 reads the notes of `ready`, `active` and `review` cards written at or after the last reservation, with kind and author; statement 3 keeps only the landings at or after the last reservation; `marker_of` gets `REPAIRED`; the note walk skips an answered block; `Flow.ready` counts only the cards of the stage ready; the local day start.
3. `render.rs`: `🟡` is `working + stale`, `🙋` counts the non-withdrawn hold cards.
4. Goldens for the changed lines (`b13`, `b14` and any other the targets show), read by hand; the seven numbers of the report as the first golden-like check.
5. The comment at `state.rs:206` and the `docs/rust-product-path.md` flow-line item say what the code does (drafts are not on the line).
6. (revision 2) Red first for the limit windows: the tests of `window_tests` that pin the share left change to the share used (`the_row_shows_what_is_used_of_each_window`, `the_battery_icon_turns_above_the_high_use_boundary`, the bar and the level tests) and fail on the current head; then `Window::known` writes the used percentage, the filled cells and the icon from it, and `level_color` reads the used share with the constants `LEVEL_GREEN_UP_TO` (30), `LEVEL_YELLOW_UP_TO` (50) and `HIGH_USE_PERCENT` (70); `LOW_LEFT_PERCENT`, `LEVEL_GREEN_FROM` and `LEVEL_YELLOW_FROM` go. The same step rewrites the limit lines of `tests/statusline_cards.rs`, the goldens (`python3 -I tools/statusline_golden.py --write`, then `--check`, every changed golden read) and the limit-row item of `docs/rust-product-path.md`.

## Risks

1. **First lines are still a convention.** `REPAIRED` and the leader decision are two more rules in the temporary parser (flow Risk 1). A lane that writes its hand-back in another form stays invisible; the fixture pins the words, not the lanes.
2. **"At or after" at the same second.** A hand-back and the next take written in the same second would read as after the take. Seconds are the ledger's finest time; the hand-back of one lane and the take of another in one second needs two different sessions on one card, which the reservation refuses.
3. **The decision note is read by author and kind**, not by words. A leader who answers with a `report` note or another author name leaves the block standing (the card shows 🚧 until a lane note follows). Chosen so that no text of a note can end a block by accident.
4. **The local day on a day with a zone change.** `mktime` takes the zone's rules at the date; a day of 23 or 25 hours counts from the true local midnight. The status line has no DST fixture on every OS; the tests use fixed-offset zones and the child `TZ`.
5. **Not a unix machine.** `libc` is a unix-only dependency; elsewhere the day is the UTC day (REQ-10). Windows is not claimed.
6. **A full bar now means a nearly spent window.** The earlier bar was full when much was left; a user who learned that reading sees it flip, and the bar color goes the other way (green when little is used). The user chose it (note 8394), the icon and the countdown keep their meaning, and the golden churn is every limit line: the plain goldens change in numbers and cells, the colored goldens in the filled cells' color.
7. **One file, two jobs.** `render.rs` carries the flow line and the limit row. The leader's order is one revision and one approval; the flow part is built and tested on its own and the limit part follows the approval, so a failure of one is not hidden by the other.

## Decisions

Reversible, category `none`; the leader's delegation of 2026-11-23T13:57Z covers decisions of this kind. The first three rows are the card's leader decisions (card C-452), the rest are this SPEC's reading of them and wait for the leader's check.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the stage of a card comes from the newest marker note written at or after the card's last reservation (the largest `acquired_at` of any status), also for a `ready` card; a card never taken reads no marker | card C-452 body; the last reservation is the one fact the ledger keeps for "this round of work" | one comparison in statement 2 |
| D-2 | a `ready` card honours only the submit and the audit PASS markers; a blocked or working marker on a card a lane released leaves it ready | card C-452 ("a hand-back or PASS after the last reserve"); a released block waits for a lane like any ready card | add the two markers to the ready branch |
| D-3 | a block ends when a note of kind `decision` signed `leader` is newer than it; the next older marker (or none) decides | card C-452 ("a LEADER decision note newer than a LANE-BLOCKED"); the typed `kind` and `author` columns, no text | read a word instead of the author |
| D-4 | submitted, audited and blocked cards keep their stage whatever their lease says; a card with no marker whose lease ended is `stale` and is drawn in `🟡` with the working cards | card C-452 (defect 7 stays: the numbers are true); flow D-2 changed only in what is drawn | draw stale as its own item |
| D-5 | a landing counts only when it is at or after the last reservation | card C-452 (defect 8) | read any landing |
| D-6 | `🙋` counts the hold cards that have no `card.withdraw` or `card.fold` event; a withdrawn hold card is in no drawn stage; the item is hidden at 0. Today every hold card is withdrawn, so the item is 0 until a hold with a user's pick exists | card C-452 ("🙋 counts only a typed user-pick signal, 0 until one exists"); the same rule `moai card list` uses to hide withdrawn cards; the existing `icon_spacing_in_the_real_output` fixture (a hold card with no event shows `🙋 1`) stays true | count nothing in `🙋` until a typed signal exists |
| D-7 | done today = closed at or after the local midnight of the machine's zone (`libc::mktime` on unix, UTC elsewhere) | card C-452 ("today is the local day of the machine's time zone"); the user reads the line in Seoul | UTC again, one function |

Revision 2 (user decision 2026-11-24, leader note 8394; the details are the leader's, reversible, category `none`):

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-8 | the number of `CW`, `5H` and `7D` is the whole used percentage (the fraction cut off, 0 to 100) | the user: "수치와 그래프를 반대로, 남아 있는 게 아닌 사용한 것으로" | `100 - used` in one function (`left_percent`) |
| D-9 | the bar fills `used * 10 / 100` (rounded down) of its ten cells and the other cells stay `░` | the same decision: the bar and the number show one quantity, as `SPEC-20261123-statusline-order-bars` D-1 chose | the filled-cells line |
| D-10 | the icon is `🪫` above `HIGH_USE_PERCENT` (70) used and `🔋` at 70 or less: the same state as before, a boundary `less than 30 left` equals `more than 70 used` | the leader: "아이콘 = 지금과 같은 상태" | a `<` in one place |
| D-11 | the level color of the filled cells goes by used: green up to 30, yellow up to 50, peach up to 70 and red above 70 (the boundaries of `SPEC-20261124-statusline-bar-color` D-1 read as used); themes keep their four colors | the leader: "쓴 양이 적으면 초록 → 많으면 빨강(같은 경계를 쓴 양으로 바꿔 읽음)" | the three constants |
| D-12 | the reset countdown `(16m)` is the time left and does not change | the leader: "시간은 남은 시간" | none needed |

### Supersedes

This SPEC changes what earlier SPECs pinned; they stay as the record of what was true. The leader's check names which of the two ways below the old SPECs take (D-1 of the digest).

| earlier SPEC | requirement | what changes |
|---|---|---|
| `SPEC-20261122-statusline-flow` | D-1 (the marker rule) | `REPAIRED` is a submit marker; a decision note of the leader ends a block; the newest marker counts from the card's last reservation |
| `SPEC-20261122-statusline-flow` | D-2 (stale is its own stage and not on the line) | stale stays its own stage and is drawn in `🟡` |
| `SPEC-20261122-statusline-flow` | D-4 and REQ-5 (done today is the UTC day) | the local day of the machine's zone |
| `SPEC-20261122-statusline-flow` | REQ-2 and REQ-3 (stage by notes of the whole history; a landing at any time wins) | stage by the notes at or after the last reservation; a landing counts at or after it |
| `SPEC-20261122-statusline-limits-left` | REQ-6 (`🟡 <working>`, `🙋 <hold>`, "the ended leases are not on the line") | `🟡 <working plus stale>`, `🙋 <hold cards without a withdraw event>`; the drafts are still not on the line |
| `SPEC-20261122-statusline-flow` | REQ-1 and AC-1 (the nine counts add up to the cards that are not done) | the drawn items add up to the open cards that are neither draft nor hold (REQ-7 here), and a hold card with a withdraw or fold event is in no drawn stage (REQ-8, D-6); the drafts are still not on the line |
| `SPEC-20261122-statusline-flow` | AC-3 (the done-today count changes at the UTC midnight) | it changes at the local midnight of the machine's zone (REQ-9, REQ-10 here, D-7) |
| `SPEC-20261124-ready-receipt-followup` | REQ-1 and REQ-4 (the status line judges a hand-back with the one function `marker_of_line`, whose markers are `LANE-READY-INTEGRATION`, `REPAIRED` and `READY `) | nothing in that SPEC changes: the stage rule of the status line had a marker list of its own that left `REPAIRED` out, and REQ-4 here makes it a submit marker so the stage rule agrees with that function |
| `SPEC-20261122-statusline-limits-left` | AC-7 (the flow line example and the rule that the cards with an ended lease are not on the line) | the example keeps its shape; the work item counts the working and the stale cards, and the question item counts the hold cards without a withdraw or fold event |
| `SPEC-20261122-statusline-limits-left` | REQ-1, REQ-2 and REQ-3 (the number of each window is the share left, `🪫` below 30 left) | the number is the share used (REQ-12), `🪫` above 70 used (REQ-13), the same state as before; `5H` shown at any level and the countdowns stay |
| `SPEC-20261122-statusline-limits-left` | AC-1, AC-3 and AC-4 (their examples and boundaries are shares left) | the examples and boundaries are shares used (AC-9, AC-10) |
| `SPEC-20261123-statusline-order-bars` | REQ-3 and REQ-7 (the bar draws the share left, the numbers are the share left), D-1 | the bar draws the share used with the same ten cells and the same rounding (REQ-12, D-9) |
| `SPEC-20261123-statusline-order-bars` | AC-4 (the example 27%, 51% and 13% used reads `73%`, `49%` and `87%`) | the same input reads `27%`, `51%` and `13%` (AC-9, AC-10 here) |
| `SPEC-20261124-statusline-bar-color` | REQ-1 (the filled cells are painted in one level color chosen from the share left) | the one level color is chosen from the share used (REQ-14, D-11 here) |
| `SPEC-20261124-statusline-bar-color` | REQ-2, D-1 (the level colors go by the share left: green from 70, yellow from 50, peach from 30, red below) | the level colors go by the share used: green up to 30, yellow up to 50, peach up to 70, red above (REQ-14, D-11); the four theme colors and the empty-cell color stay |

## Result

구현했어요(2026-11-24, 브랜치 `spec/statusline-counts-truth-20261124`, 코드·시험·골든 커밋 `3136e9e6c3`). 개정 1(개수)은 SPEC 개정 1 승인(리더 노트 8366) 뒤, 개정 2(한도 쓴 양)는 승인(리더 노트 8403) 뒤에 구현했어요. 개정 2 의 red 시험은 코드를 바꾸기 전에 먼저 돌렸어요(`window_tests` 14개 중 13개 실패, 옛 코드, 남은 양 문자열). 개정 1 의 시험은 재현 커밋 `ef55e2b537` 이 옛 줄을 고정한 뒤 최종 기대값으로 고쳐 옛 코드에서 실패하는 것을 확인했어요. 독립 감사는 이 레인이 하지 않아요(리더가 다른 세션에 맡겨요).

- 시험 결과(헤드 `3136e9e6c3` 의 작업 트리, 마지막 실행 전체는 READY 노트의 `moai check affected` 영수증에 있어요): `statusline::` lib 와 `statusline_cards`, `statusline_counts_truth`, `statusline_counts_refresh`, `statusline_flow`, `statusline_numbers_only`, `statusline_tag_order`, `statusline_golden` 을 한 번에 돌려 299개 중 298 통과·1 실패(`statusline::input::...a_mistyped_field_changes_only_itself`, 내가 `input.rs` 기대값을 남은 양 79 의 짝(21)이 아니라 41 로 잘못 고친 시험 쪽 실수), 고친 뒤 `statusline::input::` 13개 모두 통과했어요.
- 골든: `tools/statusline_golden.py --write` 로 114개를 다시 쓰고 바뀐 plain 24개의 모든 줄을 읽었어요. 한도 줄은 모두 `100 − 옛 숫자` 쌍이에요(92→8, 59→41, 46→54, 77→23, 20→80, 7→93, 70→30, 88→12, 33→67, 75→25, 68→32), 막대 칸 수는 쓴 양의 `× 10 / 100` 내림이고 🔋/🪫 는 70 초과에서만 🪫 이에요. 장부가 있는 골든(`b13`, `b14`)만 `🟡 3`, `🟡 2` 가 늘었어요. `cases.json` 은 바이너리 SHA-256 줄만 바뀌었어요.
- 뮤턴트(레인이 직접, 커밋 뒤 하나씩 적용하고 되돌림, 되돌린 뒤 `git status` 변경 없음): (1) 철회된 hold 카드를 세게 하는 SQL 조건(`AND 0`)에서 `--test statusline_counts_truth` 가 exit 100 으로 `a_withdrawn_hold_card_is_not_a_pick_the_user_owes` 와 `the_0732z_ledger_shape_draws_what_the_ledger_holds` 실패, (2) `DayZone::Machine` 이 UTC 자정을 쓰게 한 것에서 `the_done_count_follows_the_local_day` 실패(exit 100), (3) 🪫 경계를 `>` 에서 `>=` 로 바꾼 것에서 `window_tests` 4개(`the_bar_has_ten_cells_and_changes_at_each_tenth`, `the_battery_icon_switches_at_70_and_71`, `the_battery_icon_turns_above_the_high_use_boundary`, `the_five_hour_window_shows_whenever_the_input_carries_it`) 실패(exit 100), (4) 초록 경계를 `<=` 에서 `<` 로 바꾼 것에서 `the_filled_cells_carry_the_level_color`·`the_level_color_changes_at_each_step_point` 실패(exit 100).
- 해석 (나)(리더 노트 8403 승인): 예약 행이 없는 active/review 카드는 모든 노트를 읽어요(옛 동작). 바이너리 골든 장부 픽스처가 이 모양이라서예요.
- 시험 파일 정리(승인된 범위 확대, 노트 8420·8425·8427): `statusline_numbers_only.rs`, `input.rs` 시험 모듈, `statusline_counts_refresh.rs` 는 기존 기대값 줄만 바꿨고 순서와 다른 줄은 그대로예요(C-175 가 같은 파일에 새 시험을 더해요).
- 문서: `docs/rust-product-path.md` 는 카드 C-449 가 파일 전체로 예약 중이라 처음 line-lease 가 거절됐고(리더 노트 8454), 리더가 give-back 을 요청한 뒤(노트 8456) line-lease 로 세 곳(한도 항목, 예시 두 줄, 흐름 줄의 끝난 lease 문장)만 고쳐 커밋 `91e561b298` 로 넣었어요.

- AC-1: met — `the_0732z_ledger_shape_draws_what_the_ledger_holds` 가 UTC 에서 `🗂️ 77 → 🟡 12 → 📨 12 → 🛬 7 → ✅ 50 | 🚧 4`, 그린 항목의 합이 draft 와 hold 가 아닌 열린 카드 수인 것을 고정해요.
- AC-2: met — `the_done_count_follows_the_local_day` 가 자식 프로세스에만 `TZ` 를 주고 Asia/Seoul·Etc/GMT-9 에서 `✅ 103`, UTC 에서 `✅ 50`, Etc/GMT+8 에서 `✅ 103` 을, `done_today_tests` 가 자정 경계 초 단위를 고정해요.
- AC-3: met — `a_withdrawn_hold_card_is_not_a_pick_the_user_owes` 가 `🗂️ 1`(철회된 hold 만)과 plain hold 를 더한 `🗂️ 1 | 🙋 1`, 0 일 때 숨김을 고정하고, 모든 hold 를 세는 뮤턴트에서 실패해요(뮤턴트 항목 1).
- AC-4: met — `stage_tests_after_reserve` 가 표식 단어의 경계(예약 한 초 전은 ready, 예약 시각부터 제출됨), `REPAIRED`, 리더 결정 노트의 앞·뒤·같은 순서, lane 이 쓴 decision 은 답이 아닌 것을 고정해요.
- AC-5: met — 같은 모듈이 착지 뒤 다시 take 한 카드는 working, 마지막 take 뒤 착지는 landed 인 것을 고정해요.
- AC-6: met — `done_today_tests` 가 읽을 수 없는 시간대에서 UTC 날로 세는 것을 고정해요.
- AC-7: met — `the_four_statements_read_and_never_write` 가 네 문장이 SELECT 뿐인 것을(낱말 단위 검사), `--test statusline_flow` 와 `statusline_counts_truth` 가 렌더 뒤 HOME 스냅숏이 같은 것을 고정해요.
- AC-8: not_run — `moai check affected` 영수증은 READY 노트가 헤드와 함께 담아요.
- AC-9: met — `the_row_shows_what_is_used_of_each_window` 가 사용자 예 `🔋 CW ██░░░░░░░░ 25% | 🔋 5H ███░░░░░░░ 32% (16m) | 🪫 7D ███████░░░ 78% (2d 15h)` 를, `the_bar_has_ten_cells_and_changes_at_each_tenth` 가 0·9·10·11·19·20·99·100 의 칸 수와 0~100 모든 값에서 숫자가 쓴 양인 것을 고정해요.
- AC-10: met — `the_battery_icon_turns_above_the_high_use_boundary` 가 70 은 🔋, 71 은 🪫 를 `CW`·`5H`·`7D` 에서 `HIGH_USE_PERCENT` 로 읽은 값으로 고정해요.
- AC-11: met — `the_level_color_changes_at_each_step_point` 가 30 초록·31 노랑·50 노랑·51 복숭아·70 복숭아·71 빨강과 단조성, `the_filled_cells_carry_the_level_color` 가 두 테마·두 프로필의 실제 바이트를, `only_the_filled_cells_take_the_level_color` 가 채운 칸 밖에 레벨 색이 없는 것을 고정해요.
- AC-12: met — 실제 바이너리가 `statusline_cards` 의 줄 사다리와 골든 114개에서 쓴 양 숫자를 쓰고, 색 골든은 새 막대 색을 담고, plain 골든 24개는 한도 숫자와 막대만 바뀌었어요(위 골든 항목). 제품 문서는 커밋 `91e561b298` 가 "남은 → 쓴"으로 고쳤어요. 문서를 읽는 `statusline_numbers_only` 시험은 READY 노트의 영수증 실행이 보여요.
