---
id: SPEC-20261126-statusline-stale-submit
title: "The status line counts a submission nobody audited for a day as ready again"
status: draft
revision: 1
created: 2026-11-26
---

# The status line counts a submission nobody audited for a day as ready again

## Intent

사용자 질문(2026-11-25T19:3xZ, 상태줄 `📋 77 → 🔨 4 → 👀 25 → 🚀 17 → ✅ 23 | ⛔ 2 ⏸️ 98` 이 정확한가)에 리더가 원장을 `sqlite3` 로 다시 센 결과예요. 숫자는 `src/statusline/state.rs` 의 규칙대로 맞아요(독립 계산 76/4/26/17/2, 한 장 차이는 두 읽기 사이의 노트). 그러나 `👀` 의 뜻이 틀려요: 25장 가운데 24장이 하루 넘게 지난 READY 노트를 가진 `ready` 카드예요. 레인이 제출하고 예약을 돌려준 뒤(`release` 가 카드를 `ready` 로 돌려요) 아무도 감사하지 않는 카드인데 "검사 중"으로 보여요. 지금 실제로 감사 중인 카드는 두세 장이에요.

`SPEC-20261124-statusline-counts-truth` 의 REQ-1 은 "마지막 예약 때나 그 뒤에 쓴 제출 표식이 있는 `ready` 카드는 제출됨" 이라고만 말해서 표식의 나이를 안 봐요. 이 SPEC 은 그 REQ-1 한 곳에 나이 한계를 더해요: 한계(하루)를 넘긴 제출 표식만 있는 `ready` 카드는 다시 `ready`(`📋`)로 세요. 나머지 계산(감사 PASS, `active`·`review` 카드, 막힘, 착지)과 리더가 쓴 AUDIT PASS 노트를 판정으로 세지 않는 규칙(`answers_a_block`)은 그대로예요.

## Non-goals

- 새 항목, 새 아이콘, `src/statusline/render.rs` 변경: 오래된 제출은 이미 그려지는 `📋` 항목에 들어가요(D-2).
- 감사 PASS 표식의 나이 한계: 오래돼도 착지를 기다리는 카드라서 `🚀` 그대로예요(D-3).
- `active`·`review` 카드의 단계: 예약이 살아 있는 동안은 한계를 두지 않아요(D-4).
- 원장 쓰기, 새 상태, 새 명령, 오래된 제출을 정리하는 일(리더의 `card reclaim` 과 재배차 몫).

## Requirements

- REQ-1: When the newest submit marker note of a ready card is older than the stale-submit limit at the render's clock, the status line shall count the card as ready · failure: none — a marker exactly at the limit or younger keeps the card submitted
- REQ-2: The status line shall keep the audited stage of a ready card whose audit PASS marker is older than the stale-submit limit.
- REQ-3: The status line shall keep the stage of an active or review card whatever the age of its marker.
- REQ-4: The status line shall read the flow in the same four read-only statements within the render budget.

## Acceptance

- AC-1 (REQ-1): 임시 HOME 의 원장 픽스처에서 제출 표식이 한계보다 한 초 오래된 `ready` 카드는 `📋` 에, 한계와 같은 나이와 더 젊은 카드는 `👀` 에 세고, 그린 항목의 합은 draft 와 hold 가 아닌 열린 카드 수 그대로임 · verify: `cargo nextest run --locked --offline --test statusline_counts_truth an_old_submission_on_a_ready_card_counts_as_ready`
- AC-2 (REQ-1, REQ-2, REQ-3): 단계를 정하는 순수 함수의 표가 나이 경계(한계, 한계+1초)에서 `ready`+제출은 `Ready`, `ready`+감사 PASS 는 `Audited`, `active`+제출과 `review` 는 `Submitted` 로 고정함 · verify: `cargo nextest run --locked --offline --lib statusline::state::stale_submit_tests`
- AC-3 (REQ-1, REQ-2, REQ-3): 이 SPEC 이 안 바꾸는 계산은 기존 시험이 그대로 지킴(앞선 SPEC 의 07:32Z 원장 픽스처 줄, 표식 단어의 경계, 막힘·착지, 리더 결정 노트) · verify: `cargo nextest run --locked --offline --lib statusline::state` and `cargo nextest run --locked --offline --test statusline_counts_truth --test statusline_flow --test statusline_cards`
- AC-4 (REQ-4): 흐름 읽기가 여전히 문장 네 개이고 SELECT 뿐이며 렌더 뒤 원장 바이트가 같음 · verify: `cargo nextest run --locked --offline --lib statusline::state` and `cargo nextest run --locked --offline --test statusline_flow`
- AC-5: 가벼운 검사 · verify: `git diff --check`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`; `moai spec check`; `moai trailer-check`

## Scope

- write: `.moai/specs/SPEC-20261126-statusline-stale-submit/**`, `src/statusline/state.rs`, `tests/statusline_counts_truth.rs`, `tests/statusline_flow.rs`, `tests/statusline_cards.rs`
- read-only: `src/statusline/render.rs`, `src/ledger/**`, `src/card/ready_receipt.rs`
- forbidden: the real HOME and ledger, the network in tests, any write to a ledger from the status line, a new crate

## Plan

1. Red first: `an_old_submission_on_a_ready_card_counts_as_ready` in `tests/statusline_counts_truth.rs` (a ready card with a submit note two days before the render's clock) and the table of `stale_submit_tests` fail on the current head because the card is counted submitted.
2. `state.rs`: the notes statement also reads each note's `created_at`; the first marker of a card keeps its time; a pure function takes the card's state, marker and marker time with the render's clock and answers the marker `stage_of` should see (a submit marker of a ready card older than `STALE_SUBMIT_SECS` is none). `stage_of` itself keeps its signature and its table.
3. Run the existing statusline tests unchanged (AC-3, AC-4); fix a fixture only where its note times are older than the limit by accident, and say so in the Result.
4. Guard mutant: the pure function returns the marker unchanged; the new tests fail.
5. Result with the commands and exits; `check affected --card` at READY.

## Risks

- **A fixture that stamps old notes.** An existing fixture whose submit notes are more than a day older than its render clock would move from `👀` to `📋`. The run of the existing targets shows it; a changed fixture is named in the Result and its expectation is not weakened.
- **One day is a convention.** The limit is the wording of the report ("하루 넘게"), one constant. A lane that submits and waits longer than a day for its audit shows as ready again until the audit starts; that is the point, the leader sees the card is waiting, not being checked.
- **Clock skew.** The render's clock and the note time come from the same machine's ledger; a note newer than the clock counts as fresh.

## Decisions

Reversible, category `none`; the leader's delegation covers decisions of this kind. The leader's dispatch asked for a recommended default in the SPEC draft and said a tier 1 change needs no approval.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the limit is one day (86400 seconds): a submit marker strictly older than it is stale, at the limit it is not | the report wording "하루 넘게 지난" and the leader's dispatch "하루(또는 정한 기준)" | change `STALE_SUBMIT_SECS` |
| D-2 | a stale submission is counted as ready (`📋`), the recommended default; the alternative is a separate item with its own icon | no new icon or render change, the drawn items still add up to the open cards, and a card nobody checks is a card waiting for a lane; a separate item needs an icon choice that is the user's | draw `Stage::Stale`-like item in `render.rs` |
| D-3 | an audit PASS marker keeps its stage whatever its age | a PASS card waits to land, not to be checked, so `🚀` stays true | add the age test to the audited branch |
| D-4 | `active` and `review` cards are not aged | the reservation is held; the existing live and stale lease rules already say what they are | add the age test to the held branches |
