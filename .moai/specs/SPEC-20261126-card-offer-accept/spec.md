---
id: SPEC-20261126-card-offer-accept
title: "A card holder offers a card to one named session and that session accepts it, with every refusal writing nothing"
status: approved
revision: 1
created: 2026-11-26
parent: SPEC-20261120-launcher-liveness
approval:
  by: leader
  at: 2026-11-26
  basis: "리더 DELEGATED-APPROVAL 노트 9599 (사용자 위임 2026-11-25 ~14:50Z, 리더 판단, 사용자 직접 선택 아님, 노트 9598 에 답함): 개정 1 과 D-1~D-4 기본값을 승인으로 기록했어요; 레인은 그 노트를 인용만 해요"
  digest: "sha256:78c90124401f94ee7a67d9211fc77e2af80bb5823ef4176046e3231e4cca4e52"
---

# A card holder offers a card to one named session and that session accepts it, with every refusal writing nothing

## Interview

- Q1: Why two new verbs when `card handoff --to` exists? -> A: A handoff moves the card at once and the receiver has no say; the design report (section 5.2, slice d) asks for an offer the receiver accepts, and it keeps `card handoff` as it is. The design report also records no `handoff` row in the real ledger when it was written, so the new verbs change no habit.
- Q2: How long is an offer valid, and how long is the new lease? -> A: 1800 seconds by default and a new 4-hour lease at the accept (decisions D-1; the leader confirmed both defaults with the approval).
- Q3: Do slices (b) and (c) of the card need code here? -> A: No. Own-session resume is `card adopt` (SPEC-20261123-lane-continues-card, test `ac12_adoptions_do_not_use_up_the_reclaim_allowance`); the `lease_expired` recount is a report in the Result.
- Q4: How does the target learn of an open offer? -> A: In this slice by a cross-session message from the offering session; showing an open offer in `card show` or `card inbox` is a later slice (Non-goals).

## Intent

`card handoff --to` gives a card away in one step: the receiver has no say, and a mistaken session id moves the card to nobody. The collab-board design (3rd edition, section 5.2, BOARD 03 slice d) asks for two verbs instead: the holder offers the card to one named session, nothing moves until that session accepts, and every situation where the offer is no longer good (it ran out, the holder's lease ended, the holder cancelled, the generation moved, the wrong session asks) is refused and leaves the holder in charge. This SPEC adds `card offer` and `card accept` on the ledger as it is: the ledger schema stays version 1 (no table, index or journal mode changes), `card handoff` stays as it is, and the two verbs write only `event` rows plus, on an accept, the same reservation rows a handoff writes.

It also records, for the card that owns this slice, that slices (b) and (c) of the card are settled by other work: own-session resume is `card adopt` (SPEC-20261123-lane-continues-card, test `ac12_adoptions_do_not_use_up_the_reclaim_allowance`), and the `lease_expired` recount is a report (see the Result).

## Non-goals

- Changing `card handoff`, `card reclaim`, `card adopt` or the reclaim allowance of three, or the 4-hour lease default.
- An automatic lease renew (the design's `ma-2`): the recount in the Result only reports a number; the decision is the leader's with the user.
- A ledger v2, BOARD 02, BOARD 06, a new table, a new index, or a host hook.
- Showing an open offer in `card show`, `card inbox` or the statusline: a later slice. Until then the offering session tells the target by a cross-session message, and the message is a notice, the ledger is the truth.
- Sending the message itself, or any wake-up of the target session.

## Requirements

- REQ-1: When the session that holds a card at its current generation runs `card offer --to <session id>`, the `card offer` command shall write one `card.offer` event that names the holder, the generation, the target, the end of the offer's validity and the handover text, and leave the reservation, the generation and the card's state as they are.
- REQ-2 (REQ-1): IF the caller is not the holder at that generation, or the lease has ended, or the generation is not the current one, or `--to` is the holder itself or is not a session id, THEN the `card offer` command shall refuse and change no reservation; a refused holder call is recorded as one `fence_reject` event exactly as for `card handoff`.
- REQ-3 (REQ-1): When a holder offers a card that already has an offer at the same generation, the `card offer` command shall make the newest offer the only valid one · failure: none — the earlier offer stays in the event table as history and `card accept` reads only the newest event of the offer family.
- REQ-4: When the holder runs `card offer --cancel` for a card whose newest event of the offer family is a `card.offer` at the current generation, the `card offer` command shall write one `card.offer_cancel` event that voids that offer and leave the reservation as it is.
- REQ-9 (REQ-4): IF the card has no open offer at its current generation, THEN the `card offer` command shall refuse a `--cancel` with the code `no_offer` and write nothing.
- REQ-5: When the session named in the newest valid offer runs `card accept` at the offered generation, the `card accept` command shall, in one ledger transaction, end the old reservation as `released`, give the accepting session the next generation with the same write scope, worktree and branch and a new lease from the ledger clock, and write one `card.accept` event.
- REQ-6 (REQ-5): IF the card has no offer at that generation, or the newest offer was cancelled, or the caller is not the target, or the offer's validity ended, or the holder's lease ended, or the offered generation is not the current one or the reservation is not `held`, THEN the `card accept` command shall refuse with a code that names the cause and write nothing.
- REQ-7: When the `card offer` command or the `card accept` command writes, the ledger shall hold no new `card.reclaim` event · failure: none — the count of `card.reclaim` events is compared before and after, and the allowance of three reclaims is never read.
- REQ-8: While an offer waits on a card, the ledger shall answer a `card take`, a `card reclaim` and a `card adopt` of that card as it does for any held card · failure: none — the offer is an event only and changes no reservation row.

## Acceptance

- AC-1 (REQ-1, REQ-5, REQ-7): 유효 시간 안에 대상이 수락하면 이전 예약은 `released`, 대상이 다음 세대의 `held` 예약(같은 범위·워크트리·가지, 새 임대)을 갖고, `card.offer` 와 `card.accept` 이벤트가 하나씩이며 `card.reclaim` 은 없어요 · verify: `cargo nextest run --locked --offline --test card -E 'test(an_offer_in_time_moves_the_card_to_the_target_on_accept)'`
- AC-2 (REQ-6): 제안의 유효 시간이 지나면 수락이 `offer_expired` 로 거절되고 원장에 새 행이 없고 기존 담당이 그대로예요 · verify: `cargo nextest run --locked --offline --test card -E 'test(an_expired_offer_is_refused_and_the_holder_keeps_the_card)'`
- AC-3 (REQ-6): 제안 뒤 기존 담당의 임대가 끝나면 수락이 `lease_expired` 로 거절되고 회수 길이 그대로 열려 있어요(`card reclaim` 이 같은 카드를 받아요) · verify: `cargo nextest run --locked --offline --test card -E 'test(an_accept_after_the_holders_lease_ended_is_refused_and_reclaim_still_works)'`
- AC-4 (REQ-4, REQ-6): 기존 담당이 제안을 취소하면 수락이 `offer_cancelled` 로 거절되고, 취소할 열린 제안이 없으면 `card offer --cancel` 이 `no_offer` 로 거절돼요 · verify: `cargo nextest run --locked --offline --test card -E 'test(a_cancelled_offer_is_refused_on_accept)'`
- AC-5 (REQ-6): 제안 뒤 세대가 바뀌면(담당이 `handoff` 로 넘김) 옛 세대로 수락은 `stale_generation`, 새 세대로 수락은 제안이 없어 `no_offer` 로 거절되고 새 담당이 그대로예요 · verify: `cargo nextest run --locked --offline --test card -E 'test(an_offer_does_not_survive_a_generation_change)'`
- AC-6 (REQ-5, REQ-6): 같은 대상이 두 번 수락하면 둘째는 `stale_generation` 으로 거절되고 첫 수락의 결과가 그대로예요 · verify: `cargo nextest run --locked --offline --test card -E 'test(a_second_accept_by_the_same_target_is_refused_as_stale)'`
- AC-7 (REQ-6): 대상이 아닌 세션의 수락은 `not_target` 으로 거절되고 기존 담당이 그대로예요 · verify: `cargo nextest run --locked --offline --test card -E 'test(a_session_that_is_not_the_target_is_refused)'`
- AC-8 (REQ-3, REQ-6): 제안이 있는데 다시 제안하면 새 제안만 유효해요: 첫 대상의 수락은 `not_target`, 새 대상의 수락은 받아들여져요 · verify: `cargo nextest run --locked --offline --test card -E 'test(a_new_offer_replaces_the_earlier_one)'`
- AC-9 (REQ-8): 수락을 기다리는 동안 다른 세션의 `take`·`reclaim`·`adopt` 가 held 카드의 기존 답대로 거절되고 제안은 그대로 유효해요 · verify: `cargo nextest run --locked --offline --test card -E 'test(an_open_offer_does_not_change_take_reclaim_or_adopt)'`
- AC-10 (REQ-2): 담당이 아닌 세션·끝난 임대·낡은 세대·자기 자신에게·세션 id 꼴이 아닌 `--to` 의 제안은 거절되고, 예약과 세대와 `card.offer` 이벤트 수가 그대로이며 담당 거절은 `fence_reject` 한 줄만 남겨요 · verify: `cargo nextest run --locked --offline --test card -E 'test(an_offer_that_the_fence_refuses_changes_no_reservation)'`
- AC-11 (REQ-7): 제안·수락·취소를 거친 카드에서 `card.reclaim` 이벤트가 0이고, 그 뒤 죽은 담당의 회수는 한도 셋까지 받아들여지고 넷째는 `reclaim_limit` 으로 거절돼요 · verify: `cargo nextest run --locked --offline --test card -E 'test(offer_and_accept_leave_the_reclaim_allowance_whole)'`
- AC-12 (REQ-5, REQ-8): 위 시험 전부가 끝난 뒤 각 카드에 `held` 예약이 많아야 하나이고 그 세대가 카드의 현재 세대와 같아요(한 담당 불변식 위반 0), 기존 넘김·회수·이어받기 시험은 그대로 통과해요 · verify: `cargo nextest run --locked --offline --test card -E 'test(every_offer_scenario_ends_with_one_owner)'`, `cargo nextest run --locked --offline --test card --test prop_reservation -E 'binary(prop_reservation) | (binary(card) & (test(handoff_cli) | test(card_reclaim_cli) | test(card_reclaim_early_cli) | test(card_adopt_cli) | test(ledger_reservation)))'`

## Scope

- write: `.moai/specs/SPEC-20261126-card-offer-accept/`
- write: `src/ledger/offer.rs`, `src/ledger/mod.rs` (one `pub mod` line)
- write: `src/card/offer.rs`, `src/card/mod.rs` (one `pub mod` line)
- write: `tests/card/card_offer_cli.rs`, `tests/card/main.rs` (one `mod` line)
- write (shared, by `card line-lease`): `src/ops/card.rs` (the two verbs), `src/help.rs` and the help golden of `card` (the two verb lines), `docs/design-review/appendix/operations/card.md` (the two rows)
- read-only: `src/ledger/reservation.rs`, `src/ledger/adopt.rs`, `src/card/adopt.rs`, `tests/prop_reservation.rs`, the design report `collab-board-final-design-2026-11-20.md` (section 5.2)
- forbidden: `src/ledger/schema.rs` (schema version 1 stays), `src/ledger/reservation.rs`, every other card's write scope

## Plan

1. RED first: `tests/card/card_offer_cli.rs` with the twelve tests of the Acceptance section, each driving the real binary against a temporary HOME and a temporary ledger; the clock is never slept on: a lease is aged by the same `UPDATE reservation SET lease_until` the reclaim tests use, and an offer whose validity ended is a second `card.offer` row copied from the real offer with `valid_until` in the past.
2. `src/ledger/offer.rs`: `offer` (fenced, live lease, writes `card.offer` or `card.offer_cancel`) and `accept` (one transaction: state check as `reclaim` makes it, then the newest offer-family event of the card, target, validity, holder lease, then the handoff rows and one `card.accept` event). Refusals of `accept` write nothing.
3. `src/card/offer.rs`: the command layer (card reference, session id from the environment, error words as `card conflict in <code>`), `src/ops/card.rs`, help and the operations appendix.
4. Run the tests until green, then the neighbouring ledger tests named in AC-12, then `check affected --card` once.
5. Guard mutants for the READY note: the target check, the validity check, the lease check, the newest-offer rule and the cancel rule each removed once, every one failing a test (exit 100).

## Risks

- Another card holds `src/ledger/**` for the `card next` read-speed work (SPEC-20261126-card-next-speed). This change adds one new file and one `pub mod` line there; a conflict on the `mod.rs` line is a one-line merge.
- A holder could offer a card and then hand it off by `card handoff` to somebody else: the offer then names a generation that is gone (AC-5). The target learns this from the refusal, not before; showing an open offer is a later slice.
- An offer says nothing about the work in the worktree: the handover text is free text the holder writes, and the accepting session still reads the card's notes before it writes.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The offer is valid for 1800 seconds by default (`--valid-secs`, 60 to 86400) and a new lease of the usual 4 hours starts at the accept (options: keep the old lease end as `handoff` does / start a new lease). Default: a new lease from the ledger clock, because the accepting session has not used the old one. Reversible, classification none | The design report names a validity time but no number; the card body keeps the 4-hour default | Change the two constants |
| D-2 | `card offer --cancel` is part of `card offer`, not a third verb (options: a flag / a verb `offer-cancel`). Default: a flag. Reversible, classification none | The design table has one row for a cancel and the two named verbs stay two | Add the verb later |
| D-3 | The refusals of `accept` write nothing, not even a `fence_reject` event (options: write one / write none). Default: none, as `reclaim` and `adopt` do, because the caller is not the holder the fence checks and a stranger could otherwise fill the event table. Reversible, classification none | `card reclaim` and `card adopt` write nothing on a refusal | Record a refusal event later |
| D-4 | The two events are named `card.offer`, `card.offer_cancel` and `card.accept`, free strings of the existing `event.kind` column (no schema change). Reversible, classification none | The design report: the kind is a free string | Rename in one place |
