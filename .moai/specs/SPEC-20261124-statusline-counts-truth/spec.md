---
id: SPEC-20261124-statusline-counts-truth
title: "The status line flow counts match the ledger: hand-backs after release, answered blocks, REPAIRED, withdrawn holds, local day"
status: draft
revision: 1
created: 2026-11-24
---

# 상태줄 흐름 줄의 숫자를 원장과 맞춰요: 풀린 뒤의 넘김, 답한 막힘, REPAIRED, 철회된 보류, 로컬 날

## Intent

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

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): the ledger fixture of the 07:32Z shape draws `🗂️ 77 → 🟡 12 → 📨 12 → 🛬 7 → ✅ 50 | 🚧 4` in UTC, the drawn items add up to the open cards that are neither draft nor hold, and each of the seven rows of the table above lands in the item the table names · verify: `cargo nextest run --locked --offline --test statusline_counts_truth the_0732z_ledger_shape_draws_what_the_ledger_holds`
- AC-2 (REQ-9): the same fixture with `TZ=Asia/Seoul` for the child only draws `✅ 103` and with `TZ=UTC` draws `✅ 50`, and the count changes at the local midnight and not a second before · verify: `cargo nextest run --locked --offline --test statusline_counts_truth the_done_count_follows_the_local_day` and `cargo nextest run --locked --offline --lib statusline::state::done_today_tests`
- AC-3 (REQ-8): a hold card with a withdraw event adds nothing to `🙋`, a hold card without one is counted, and the item is hidden at 0; a mutant that counts every hold card fails the named test · verify: `cargo nextest run --locked --offline --test statusline_counts_truth a_withdrawn_hold_card_is_not_a_pick_the_user_owes`
- AC-4 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): each marker word at its edge — a submit note one second before the last take (stays ready), at the take (counts), after it; `REPAIRED` with and without its space; a leader decision note before, after and equal in order to the block; a note of kind `decision` signed by a lane is no answer · verify: `cargo nextest run --locked --offline --lib statusline::state::stage_tests_after_reserve`
- AC-5 (REQ-6): a card with one landing and a later take is working, a card with a landing after its last take is landed · verify: `cargo nextest run --locked --offline --lib statusline::state::stage_tests_after_reserve`
- AC-6 (REQ-10): without a readable zone the day is the UTC day and the count equals the one the earlier rule gave · verify: `cargo nextest run --locked --offline --lib statusline::state::done_today_tests`
- AC-7 (REQ-7, REQ-11): the unsplit form still never guesses, the four statements are counted by the existing test and the ledger bytes stay the same after a render · verify: `cargo nextest run --locked --offline --lib statusline::state` and `cargo nextest run --locked --offline --test statusline_flow`
- AC-8: the goldens and the statusline targets pass with the changed lines rewritten and read, and `moai check affected` has a PASS receipt for the head · verify: `moai check affected <base>..<head>`

## Scope

- write: `.moai/specs/SPEC-20261124-statusline-counts-truth/**`, `src/statusline/state.rs`, `src/statusline/render.rs`, `tests/statusline_counts_truth.rs`
- write, as the changed lines need (the leader widens the card's scope; not held by this card today): `tests/statusline_flow.rs`, `tests/statusline_cards.rs`, `tests/fixtures/statusline-claude/golden/**`, `tests/fixtures/statusline-claude/cases.json` (the checksum line the golden tool writes), `docs/rust-product-path.md` (the flow-line item only), the superseded lines of `.moai/specs/SPEC-20261122-statusline-flow/**` and `.moai/specs/SPEC-20261122-statusline-limits-left/**` (see Supersedes)
- read-only: `src/ledger/**`, `src/card/ready_receipt.rs`, `tests/statusline_counts_refresh.rs` (card C-175 holds it)
- forbidden: the real HOME and ledger, the network in tests, any write to a ledger from the status line, a new crate, `TZ` set on the test process (the child only)

## Plan

1. Red first. Tests of AC-1 to AC-6 in the order of the table; each fails on the current head for the reason the table gives. The reproduction tests already in the file become the AC-1 and AC-2 tests by changing their expected lines.
2. `state.rs`: statement 1 tells a withdrawn hold card from another; statement 2 reads the notes of `ready`, `active` and `review` cards written at or after the last reservation, with kind and author; statement 3 keeps only the landings at or after the last reservation; `marker_of` gets `REPAIRED`; the note walk skips an answered block; `Flow.ready` counts only the cards of the stage ready; the local day start.
3. `render.rs`: `🟡` is `working + stale`, `🙋` counts the non-withdrawn hold cards.
4. Goldens for the changed lines (`b13`, `b14` and any other the targets show), read by hand; the seven numbers of the report as the first golden-like check.
5. The comment at `state.rs:206` and the `docs/rust-product-path.md` flow-line item say what the code does (drafts are not on the line).

## Risks

1. **First lines are still a convention.** `REPAIRED` and the leader decision are two more rules in the temporary parser (flow Risk 1). A lane that writes its hand-back in another form stays invisible; the fixture pins the words, not the lanes.
2. **"At or after" at the same second.** A hand-back and the next take written in the same second would read as after the take. Seconds are the ledger's finest time; the hand-back of one lane and the take of another in one second needs two different sessions on one card, which the reservation refuses.
3. **The decision note is read by author and kind**, not by words. A leader who answers with a `report` note or another author name leaves the block standing (the card shows 🚧 until a lane note follows). Chosen so that no text of a note can end a block by accident.
4. **The local day on a day with a zone change.** `mktime` takes the zone's rules at the date; a day of 23 or 25 hours counts from the true local midnight. The status line has no DST fixture on every OS; the tests use fixed-offset zones and the child `TZ`.
5. **Not a unix machine.** `libc` is a unix-only dependency; elsewhere the day is the UTC day (REQ-10). Windows is not claimed.

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

### Supersedes

This SPEC changes what earlier SPECs pinned; they stay as the record of what was true. The leader's check names which of the two ways below the old SPECs take (D-1 of the digest).

| earlier SPEC | requirement | what changes |
|---|---|---|
| `SPEC-20261122-statusline-flow` | D-1 (the marker rule) | `REPAIRED` is a submit marker; a decision note of the leader ends a block; the newest marker counts from the card's last reservation |
| `SPEC-20261122-statusline-flow` | D-2 (stale is its own stage and not on the line) | stale stays its own stage and is drawn in `🟡` |
| `SPEC-20261122-statusline-flow` | D-4 and REQ-5 (done today is the UTC day) | the local day of the machine's zone |
| `SPEC-20261122-statusline-flow` | REQ-2 and REQ-3 (stage by notes of the whole history; a landing at any time wins) | stage by the notes at or after the last reservation; a landing counts at or after it |
| `SPEC-20261122-statusline-limits-left` | REQ-6 (`🟡 <working>`, `🙋 <hold>`, "the ended leases are not on the line") | `🟡 <working plus stale>`, `🙋 <hold cards without a withdraw event>`; the drafts are still not on the line |
