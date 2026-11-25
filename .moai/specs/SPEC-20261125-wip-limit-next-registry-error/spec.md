---
id: SPEC-20261125-wip-limit-next-registry-error
title: "card next names no card, an audit-waiting card included, when the session registry cannot be read"
status: draft
revision: 1
created: 2026-11-25
parent: SPEC-20261125-wip-limit-live-lock-lanes
---

# card next names no card, an audit-waiting card included, when the session registry cannot be read

## Intent

SPEC-20261125-wip-limit-live-lock-lanes REQ-7 says that `card next` names no card when `wip.registry_error` is set, but the older rule of SPEC-20261125-landing-moves-checkout (`audit_needed` at the limit) still names the card that waits for an audit, so `card next` and `card take` disagree about an unreadable registry. This SPEC makes the code say what REQ-7 says (the gap recorded in that SPEC's Result).

<!-- This SPEC only makes `card next` stricter in the case where the limit cannot be checked; it removes no refusal and loosens no limit, so the gate is not weakened. -->

## Requirements

- REQ-1: IF `wip.registry_error` is not null, THEN `card next` shall answer `next` null, `next_is` `limit` and an empty `audit_needed`. · failure: none — `card next` only reads, and `card take` refuses in the same case, so the two answers agree.
- REQ-2: IF the asking session already holds a card, THEN `card next` shall keep naming that card as `mine`, whatever `wip.registry_error` says. · failure: none — resuming a held card takes no new card and needs no registry.
- REQ-3: IF `wip.registry_error` is null, THEN `card next` shall answer `audit_needed` and `next_is` `audit` at the limit as before. · failure: none — the audit path of SPEC-20261125-landing-moves-checkout is unchanged.

## Acceptance

- AC-1 (REQ-1): with the sessions folder replaced by a link and one card in review that the asking session did not write, `card next` answers `wip.registry_error` as a string, `next_is` `limit`, `next` null and `audit_needed` `[]` · verify: `cargo test --locked --offline --test card card_wip_limit::registry_error_next_names_no_card_not_even_one_that_waits_for_an_audit`
- AC-2 (REQ-2, REQ-3): the existing `next_at_the_limit_names_the_cards_that_wait_for_an_audit` still passes (audit named at the count limit, not for the writer of the card) · verify: `cargo test --locked --offline --test card card_wip_limit::next_at_the_limit_names_the_cards_that_wait_for_an_audit`

## Scope

- write: `.moai/specs/SPEC-20261125-wip-limit-next-registry-error/**`, `src/card/work.rs`, `tests/card/card_wip_limit.rs`
- forbidden: `src/card/wip_limit.rs` and `src/lane_registry.rs` (SPEC-20261125-session-registry-session-key changes them), any loosening of the limit, Python

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | (A, 기본·권장) 레지스트리를 못 읽을 때 `card next` 는 감사 대기 카드도 내놓지 않아요: 승인된 SPEC-20261125-wip-limit-live-lock-lanes 의 REQ-7("name no card")과 D-4 문장 그대로예요. 상한을 확인할 수 없으니 먼저 폴더를 고치거나 `wip_limit` 0 으로 가요. (B) 감사는 상한과 무관하니 계속 내놓아요: 이 경우 REQ-7 을 "take 할 카드만"으로 개정해야 하고 승인된 SPEC 의 digest 가 바뀌어요. | 카드 본문 "SPEC 문장과 다른 쪽을 고르면 SPEC 개정이 필요해요", 승인본 문장 | `registry_error` 일 때 `audit` 를 비우는 한 줄을 지워요 |

## Plan

1. RED: `tests/card/card_wip_limit.rs` 에 AC-1 의 시험을 더해요. 지금 코드에서는 `next_is` 가 `audit` 라서 실패해요.
2. `src/card/work.rs` 에서 `wip.registry_error` 가 있으면 `audit_needed` 를 읽지 않고 빈 목록으로 해요. GREEN.
3. 가드 뮤턴트(그 한 줄을 빼면 AC-1 이 죽어요), `cargo fmt --all`, `moai check affected --card` 영수증, READY.
